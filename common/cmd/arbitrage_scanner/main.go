package main

import (
	"time"

	internal "github.com/krzy37/arbitrage-scanner/common/internal/exchange/binance"
)

func main() {
	/*
		cfg := shared.NewConfigMust()
		logger, err := shared.NewLogger(cfg)
		if err != nil {
			panic(fmt.Sprintf("error creating logger: %v", err))
		}
		defer logger.Close()
		defer func() {
			_ = logger.Sync()
		}()

		logger.Info("logger init success")
	*/
	internal.BinanceConnect()

	time.Sleep(time.Duration(5) * time.Second)
}
