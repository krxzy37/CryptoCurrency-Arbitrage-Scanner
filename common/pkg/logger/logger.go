package shared

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

type Logger struct {
	*zap.Logger
	file os.File
}

func newLogger(logLevel string) (*Logger, error) {
	zapLvl := zap.NewAtomicLevel()
	if err := zapLvl.UnmarshalText([]byte(logLevel)); err != nil {
		return nil, fmt.Errorf("error parsing log level: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(logLevel), 0755); err != nil {
	}
}
