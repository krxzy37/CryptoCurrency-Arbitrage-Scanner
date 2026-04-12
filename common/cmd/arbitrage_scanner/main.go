package main

import (
	"fmt"

	internal "github.com/krzy37/arbitrage-scanner/common/internal/connect"
	shared "github.com/krzy37/arbitrage-scanner/common/pkg/logger"
)

func main() {

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

	internal.BinanceConnect()
}
