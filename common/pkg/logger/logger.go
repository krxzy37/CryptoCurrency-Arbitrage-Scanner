package shared

import "go.uber.org/zap"

type Logger struct {
	*zap.Logger
}
