package fetch

import (
	"go.uber.org/zap"
)

var logger *zap.Logger

// func NewFileLogger(fNames ...string) *zap.Logger {
// 	//logger, _ := zap.NewProduction()
// 	cfg := zap.NewProductionConfig()
// 	//cfg.OutputPaths = []string{"info.log"}
// 	cfg.OutputPaths = fNames
// 	//cfg.DisableCaller = true
// 	cfg.DisableStacktrace = true
// 	encoderCfg := zapcore.EncoderConfig{
// 		TimeKey:        "ts",
// 		MessageKey:     "msg",
// 		LevelKey:       "level",
// 		NameKey:        "fetcher",
// 		EncodeLevel:    zapcore.CapitalLevelEncoder,
// 		EncodeTime:     zapcore.ISO8601TimeEncoder,
// 		EncodeDuration: zapcore.StringDurationEncoder,
// 		EncodeName:     zapcore.FullNameEncoder,
// 	}
// 	cfg.EncoderConfig = encoderCfg
// 	//cfg.EncoderConfig.TimeKey = ""
// 	//cfg.EncoderConfig.LevelKey = ""

// 	logger, _ := cfg.Build()
// 	return logger
// }

func init() {
	//logger, _ = zap.NewProduction()
	cfg := zap.NewProductionConfig()
	//cfg.DisableCaller = true
	cfg.DisableStacktrace = true
	// encoderCfg := zapcore.EncoderConfig{
	// 	TimeKey:        "ts",
	// 	MessageKey:     "msg",
	// 	LevelKey:       "level",
	// 	NameKey:        "fetcher",
	// 	EncodeLevel:    zapcore.CapitalLevelEncoder,
	// 	EncodeTime:     zapcore.ISO8601TimeEncoder,
	// 	EncodeDuration: zapcore.StringDurationEncoder,
	// 	EncodeName:     zapcore.FullNameEncoder,
	// }
	// core := zapcore.NewCore(zapcore.NewConsoleEncoder(encoderCfg), os.Stdout, zapcore.DebugLevel)
	// logger = zap.New(core)
	// defer logger.Sync()
	//cfg.EncoderConfig = encoderCfg

	logger, _ = cfg.Build()
	//logger, _ = zap.NewProduction()
	//logger = NewFileLogger("info.log")
}
