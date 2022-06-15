package logging


import (
	"go.uber.org/zap"
	"os"
	"fmt"
)


var Sugar *zap.SugaredLogger;

func InitLogging() *zap.Logger {
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Errorf("can't initialize zap logger: %v", err)
		os.Exit(-1)

	}
	Sugar = logger.Sugar();
	return logger

}
