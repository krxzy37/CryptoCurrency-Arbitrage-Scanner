package binance

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/VictorLowther/btree"
	"github.com/gorilla/websocket"
	"github.com/krzy37/arbitrage-scanner/common/internal/domain"
	shared "github.com/krzy37/arbitrage-scanner/common/pkg/logger"
	"go.uber.org/zap"
)

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
		price, _ := strconv.ParseFloat(item[0], 64)
		volume, _ := strconv.ParseFloat(item[1], 64)

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

	var result DepthResponse
	for {
		if err := conn.ReadJSON(&result); err != nil {
			c.logger.Error("failed to read message from websocket", zap.Error(err))
			return err
		}
		symbol := strings.Split(result.Stream, "@")[0]
		targetBook := c.manager.GetOrCreateOrderBookManager(symbol)

		fmt.Printf("Торговая пара: %v,\n %+v\n", symbol, targetBook)

		c.handleDepthResponse(targetBook, result.Data)

	}
}

func CreateURLStream(stream ...string) string {
	baseEndPoint := "wss://stream.binance.com:9443/stream?streams="

	var builder strings.Builder

	builder.WriteString(baseEndPoint)

	for i := 0; i < len(stream); i++ {
		builder.WriteString(stream[i])
		builder.WriteString("@depth")

		if i != len(stream)-1 {
			builder.WriteString("/")
		}

	}

	return builder.String()
}
