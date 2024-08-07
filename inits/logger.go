package inits

import (
	"fmt"

	"github.com/open-binance/logger"
	"go.uber.org/zap"
)

var (
	BestPriceLogger  *zap.SugaredLogger
	ReturnRateLogger *zap.SugaredLogger
)

// 初始化日志文件
func InitFileLogger(fileCfg FileConfigInfo) error {
	bestPrice := fileCfg.BestPrice
	priceCfg := logger.FileConfig{
		Enable:     true,
		LocalTime:  true,
		Compress:   bestPrice.Compress,
		MaxSize:    bestPrice.MaxSize,
		MaxAge:     bestPrice.MaxAge,
		MaxBackups: bestPrice.MaxBackups,
		Level:      bestPrice.Level,
		Filename:   bestPrice.Filename, // the file to write logs to
		TimeKey:    "time",
		CallerKey:  "",
		LevelKey:   "",
	}
	bestPriceLogger, err := logger.NewFileLogger(priceCfg)
	if err != nil {
		return fmt.Errorf("failed to new best price file logger, err: %s", err.Error())
	}
	BestPriceLogger = bestPriceLogger

	returnRate := fileCfg.ReturnRate
	returnRateCfg := logger.FileConfig{
		Enable:     true,
		LocalTime:  true,
		Compress:   returnRate.Compress,
		MaxSize:    returnRate.MaxSize,
		MaxAge:     returnRate.MaxAge,
		MaxBackups: returnRate.MaxBackups,
		Level:      returnRate.Level,
		Filename:   returnRate.Filename, // the file to write logs to
		TimeKey:    "time",
		CallerKey:  "",
		LevelKey:   "",
	}
	returnRateLogger, err := logger.NewFileLogger(returnRateCfg)
	if err != nil {
		return fmt.Errorf("failed to new return ratefile logger, err: %s", err.Error())
	}
	ReturnRateLogger = returnRateLogger

	return nil
}
