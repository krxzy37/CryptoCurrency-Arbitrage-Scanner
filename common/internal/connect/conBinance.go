package internal

import (
	"fmt"

	"github.com/VictorLowther/btree"
	"github.com/gorilla/websocket"
)

const wsendpoint = "wss://stream.binance.com:9443/stream?streams=btcusdt@depth"

func byBestBid(a, b *OrderBookEntry) bool {
	return a.Price >= b.Price
}
func byBestAsk(a, b *OrderBookEntry) bool {
	return a.Price < b.Price
}

type OrderBookEntry struct {
	Price  float64
	Volume float64
}
type OrderBook struct {
	Asks *btree.Tree[*OrderBookEntry]
	Bids *btree.Tree[*OrderBookEntry]
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		Asks: btree.New(byBestAsk),
		Bids: btree.New(byBestBid),
	}
}

func (ob *OrderBook) handleDepthResponse(res BinanceDepthResult) {
	for _, ask := range res.Asks {
		fmt.Println(ask)
	}
}

type BinanceDepthResult struct {
	Asks [][]string `json:"a"`
	Bids [][]string `json:"b"`
}
type BinanceDepthResponse struct {
	Stream string             `json:"stream"`
	Data   BinanceDepthResult `json:"data"`
}

func BinanceConnect() {
	conn, _, err := websocket.DefaultDialer.Dial(wsendpoint, nil)
	if err != nil {
		err = fmt.Errorf("websocket default dial err: %w", err)
		panic(err)
	}

	var (
		ob     = NewOrderBook()
		result BinanceDepthResponse
	)
	for {
		if err := conn.ReadJSON(&result); err != nil {
			panic(err)
		}

		ob.handleDepthResponse(result.Data)
	}
}
