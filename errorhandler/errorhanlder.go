package errorhandler

import (
	"github.com/dfeyer/flow-debugproxy/logger"
	"os"
)

// PanicHandling handle error and output log message
func PanicHandling(err error, logger *logger.Logger) {
	if err != nil {
		logger.Warn(err.Error())
		os.Exit(1)
	}
}
