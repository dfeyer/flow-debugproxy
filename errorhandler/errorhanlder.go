package errorhandler

import (
	"github.com/dfeyer/flow-debugproxy/logger"
	"os"
)

func Handling(err error) {
	if err != nil {
		logger.Warn(err.Error())
		os.Exit(1)
	}
}
