package internal

import (
	"fmt"

	"github.com/gorilla/websocket"
)

const (
	wsendpoint = "wss://stream.binance.com:9443/stream?streams=btcusdt@depth"
)

func BinanceConnect() {
	conn, _, err := websocket.DefaultDialer.Dial(wsendpoint, nil)
	if err != nil {
		err = fmt.Errorf("websocket default dial err: %w", err)
		panic(err)
	}

	fmt.Println(conn)
}
