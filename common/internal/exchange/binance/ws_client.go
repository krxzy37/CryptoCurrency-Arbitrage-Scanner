package binance

import (
	"fmt"
	"strconv"

	"github.com/VictorLowther/btree"
	"github.com/gorilla/websocket"
	"github.com/krzy37/arbitrage-scanner/common/internal/domain"
)

const wsendpoint = "wss://stream.binance.com:9443/stream?streams=btcusdt@depth"

func getBidByPrice(price float64) btree.CompareAgainst[*domain.OrderBookEntry] {
	return func(e *domain.OrderBookEntry) int {
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

func getAskByPrice(price float64) btree.CompareAgainst[*domain.OrderBookEntry] {
	return func(e *domain.OrderBookEntry) int {
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

func handleDepthResponse(ob *domain.OrderBook, res BinanceDepthResult) {
	for _, ask := range res.Asks {
		price, _ := strconv.ParseFloat(ask[0], 64)
		volume, _ := strconv.ParseFloat(ask[1], 64)
		if volume == 0 {
			if entry, ok := ob.Asks.Get(getAskByPrice(price)); ok {
				fmt.Printf("-- deleting level %.2f\n", price)
				ob.Asks.Delete(entry)
			}
			continue
		}
		entry := domain.OrderBookEntry{ // Используем пакет domain
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
				fmt.Printf("-- deleting level %.2f\n", price)
				ob.Bids.Delete(thing)
			}
			continue
		}
		entry := domain.OrderBookEntry{
			Price:  price,
			Volume: volume,
		}
		ob.Bids.Insert(&entry)
	}
}

// ИСПРАВЛЕНИЕ: Также превращаем в обычную функцию
func updateSide(
	tree btree.Tree[*domain.OrderBookEntry],
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

func Connect() {
	conn, _, err := websocket.DefaultDialer.Dial(wsendpoint, nil)
	if err != nil {
		err = fmt.Errorf("websocket default dial err: %w", err)
		panic(err)
	}

	var (
		ob     = domain.NewOrderBook()
		result BinanceDepthResponse
	)
	for {
		if err := conn.ReadJSON(&result); err != nil {
			panic(err)
		}

		handleDepthResponse(ob, result.Data)
		it := ob.Asks.Iterator(nil, nil)
		for it.Next() {

			fmt.Printf("%+v\n", it.Item())
		}
	}
}
