package domain

import "github.com/VictorLowther/btree"

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

func byBestBid(a, b *OrderBookEntry) bool {
	return a.Price >= b.Price
}
func byBestAsk(a, b *OrderBookEntry) bool {
	return a.Price < b.Price
}
