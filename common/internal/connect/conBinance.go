package internal

import (
	"fmt"
	"strconv"

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

func getBidByPrice(price float64) btree.CompareAgainst[*OrderBookEntry] {
	return func(e *OrderBookEntry) int {
		switch {
		case e.Price > price:

			return -1
		case e.Price < price:
			return 1
		default:
			return 0
		}
	}
}

func getAskByPrice(price float64) btree.CompareAgainst[*OrderBookEntry] {
	return func(e *OrderBookEntry) int {
		switch {
		case e.Price < price:
			return -1
		case e.Price > price:
			return 1
		default:
			return 0
		}
	}
}

func (ob *OrderBook) handleDepthResponse(res BinanceDepthResult) {
	for _, ask := range res.Asks {

		price, _ := strconv.ParseFloat(ask[0], 64)
		volume, _ := strconv.ParseFloat(ask[1], 64)
		if volume == 0 {
			if entry, ok := ob.Asks.Get(getAskByPrice(price)); ok {
				fmt.Printf("-- deleting level %.2f", price)
				ob.Asks.Delete(entry)
			}
			continue
		}
		entry := OrderBookEntry{
			Price:  price,
			Volume: volume,
		}
		ob.Asks.Insert(&entry)
	}
	for _, bid := range res.Bids {

		price, _ := strconv.ParseFloat(bid[0], 64)
		volume, _ := strconv.ParseFloat(bid[1], 64)
		if volume == 0 {
			if thing, ok := ob.Bids.Get(getBidByPrice(price)); ok {
				fmt.Printf("-- deleting level %.2f", price)
				ob.Bids.Delete(thing)
			}
			continue
		}
		entry := OrderBookEntry{
			Price:  price,
			Volume: volume,
		}
		ob.Bids.Insert(&entry)
	}
}

type BinanceDepthResult struct {

	//price | size (volume)
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
		it := ob.Asks.Iterator(nil, nil)
		for it.Next() {

			fmt.Printf("%+v\n", it.Item())
		}
	}
}
