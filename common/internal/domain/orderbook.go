package domain

import (
	"sync"

	"github.com/VictorLowther/btree"
)

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
