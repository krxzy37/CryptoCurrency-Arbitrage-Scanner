package binance

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/VictorLowther/btree"
	"github.com/gorilla/websocket"
	"github.com/krzy37/arbitrage-scanner/common/internal/domain"
	shared "github.com/krzy37/arbitrage-scanner/common/pkg/logger"
	"go.uber.org/zap"
)

const (
	pongWait = 20 * time.Second
)

var testChan = make(chan []byte, 100)

type RingBuffer struct {
	data       []*Data
	size       int
	lastInsert int
	nextRead   int
	emitTime   time.Time
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		data:       make([]*Data, size),
		size:       size,
		lastInsert: -1,
	}
}

func (r *RingBuffer) Insert(input Data) {

	r.lastInsert = (r.lastInsert + 1) % r.size
	r.data[r.lastInsert] = &input

	if r.nextRead == r.lastInsert {
		r.nextRead = (r.nextRead + 1) % r.size
	}
}

func (r *RingBuffer) Emit() []*Data {
	output := make([]*Data, r.size)
	for {
		if r.data[r.nextRead] != nil {
			output = append(output, r.data[r.nextRead])
			r.data[r.nextRead] = nil
		}
		if r.nextRead == r.lastInsert || r.lastInsert == -1 {
			break
		}
		r.nextRead = (r.nextRead + 1) % r.size
	}
	return output
}

type Data struct {
	stamp time.Time
	data  []*DepthResponse
}

type Client struct {
	logger  *shared.Logger
	manager *domain.OrderBookManager
}

func NewClient(logger *shared.Logger) *Client {
	return &Client{
		logger:  logger,
		manager: domain.NewOrderBookManager(),
	}
}

func getBidByPrice(price float64) btree.CompareAgainst[*domain.OrderBookEntry] {
	return func(e *domain.OrderBookEntry) int {
		switch {
		case e.Price > price:

			return 1
		case e.Price < price:
			return -1
		default:
			return 0
		}
	}
}

func getAskByPrice(price float64) btree.CompareAgainst[*domain.OrderBookEntry] {
	return func(e *domain.OrderBookEntry) int {
		switch {
		case e.Price < price:
			return 1
		case e.Price > price:
			return -1
		default:
			return 0
		}
	}
}

func (c *Client) handleDepthResponse(manager *domain.OrderBook, res DepthResult) {
	c.updateSide(manager.Asks, res.Asks, getAskByPrice)
	c.updateSide(manager.Bids, res.Bids, getBidByPrice)
}

func (c *Client) updateSide(
	tree *btree.Tree[*domain.OrderBookEntry],
	updates [][]string,
	getCompareFunc func(float64) btree.CompareAgainst[*domain.OrderBookEntry]) {
	for _, item := range updates {
		price, err := strconv.ParseFloat(item[0], 64)
		if err != nil {
			c.logger.Warn("Failed to parse price", zap.Error(err))
			return
		}
		volume, err := strconv.ParseFloat(item[1], 64)
		if err != nil {
			c.logger.Warn("Failed to parse volume", zap.Error(err))
			return
		}

		if volume == 0 {
			if entry, ok := tree.Get(getCompareFunc(price)); ok {
				fmt.Printf("-- deleting level %.2f\n", price)
				tree.Delete(entry)
			}
			continue
		}

		entry := &domain.OrderBookEntry{
			Price:  price,
			Volume: volume,
		}
		tree.Insert(entry)
	}
}

func (c *Client) Connect(endpoint string) error {
	conn, _, err := websocket.DefaultDialer.Dial(endpoint, nil)
	if err != nil {
		c.logger.Fatal("websocket default dial err", zap.Error(err))
		panic(err)
	}

	for {

		go c.ReadPump(conn, testChan)

	}
}

func CreateURLStream(stream ...string) string {
	baseEndPoint := "wss://stream.binance.com:9443/stream?streams="

	var builder strings.Builder

	builder.WriteString(baseEndPoint)

	for i := 0; i < len(stream); i++ {
		builder.WriteString(stream[i])
		builder.WriteString("@depth20")

		if i != len(stream)-1 {
			builder.WriteString("/")
		}

	}

	return builder.String()
}

func (c *Client) ReadPump(conn *websocket.Conn, messageChan chan<- []byte) {
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Printf("Failed to close connection: %v", err)
		}
	}()

	conn.SetReadLimit(4096)
	if err := conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		fmt.Printf("Failed to set read deadline: %v", err)
		return
	}
	conn.SetPongHandler(func(string) error {
		if err := conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			fmt.Printf("Failed to set read deadline: %v", err)
			return err
		}
		return nil
	})

	var result DepthResponse

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Unexpected websocket close error: %v", err)
			}
			break
		}
		if err = conn.ReadJSON(&result); err != nil {
			c.logger.Error("failed to read message from websocket", zap.Error(err))
			return
		}

		symbol := strings.Split(result.Stream, "@")[0]
		targetBook := c.manager.GetOrCreateOrderBookManager(symbol)
		fmt.Printf("Торговая пара: %v,\n %+v\n", symbol, targetBook)

		messageChan <- message

	}
}
