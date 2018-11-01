package parse

import (
	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	cfg := zap.NewProductionConfig()
	cfg.DisableStacktrace = true
	logger, _ = cfg.Build()
}
