package domain

import (
	"fmt"
	"sync"

	"github.com/VictorLowther/btree"
)

const (
	BuySide  OrderSide = "BUY"
	SellSide OrderSide = "SELL"
)

type OrderSide string
type OrderBookManager struct {
	mu    sync.RWMutex
	books map[string]*OrderBook
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

func byBestBid(a, b *OrderBookEntry) bool {
	return a.Price > b.Price
}
func byBestAsk(a, b *OrderBookEntry) bool {
	return a.Price < b.Price
}

func NewOrderBookManager() *OrderBookManager {
	return &OrderBookManager{
		books: make(map[string]*OrderBook),
	}
}

func (m *OrderBookManager) GetOrCreateOrderBookManager(symbol string) *OrderBook {

	m.mu.RLock()
	book, ok := m.books[symbol]
	m.mu.RUnlock()
	if ok {
		return book
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if book, ok = m.books[symbol]; ok {
		return book
	}

	book = NewOrderBook()
	m.books[symbol] = book
	return book
}

func (o *OrderBook) GetVWAP(side OrderSide, targetVolume float64) (float64, error) {
	if targetVolume <= 0 {
		return 0, fmt.Errorf("target volume must be positive")
	}

	switch side {
	case BuySide:
		return calculateVWAPTree(o.Asks, targetVolume)
	case SellSide:
		return calculateVWAPTree(o.Bids, targetVolume)
	default:
		return 0, fmt.Errorf("invalid order-side: %s", side)
	}

}

func calculateVWAPTree(tree *btree.Tree[*OrderBookEntry], targetVolume float64) (float64, error) {
	var totalCost float64
	originalVolume := targetVolume

	iter := tree.Iterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		entry := iter.Item()

		if targetVolume > entry.Volume {
			targetVolume -= entry.Volume
			totalCost += entry.Volume * entry.Price
		} else {
			totalCost += targetVolume * entry.Price
			targetVolume = 0
			break
		}
	}

	if targetVolume > 0 {
		return 0, fmt.Errorf("not enough liquidity: missing %f volume", targetVolume)
	}

	return totalCost / originalVolume, nil
}
