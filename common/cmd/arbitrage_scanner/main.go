package main

import (
	"fmt"
	"time"

	"github.com/krzy37/arbitrage-scanner/common/internal/exchange/binance"
	shared "github.com/krzy37/arbitrage-scanner/common/pkg/logger"
	"go.uber.org/zap"
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

	client := binance.NewClient(logger)

	if err := client.Connect(); err != nil {
		logger.Error("binance connect err", zap.Error(err))
	}

	time.Sleep(time.Duration(5) * time.Second)
}
