package flowstacktraceprocessor

import (
	"github.com/clbanning/mxj"
	"github.com/dfeyer/flow-debugproxy/config"
	"github.com/dfeyer/flow-debugproxy/errorhandler"
	"github.com/dfeyer/flow-debugproxy/logger"
	"github.com/dfeyer/flow-debugproxy/pathmapping"

	"strings"
)

// Processor manipulate the stack trace returned by xDebug
type Processor struct {
	Config      *config.Config
	Logger      *logger.Logger
	PathMapping *pathmapping.PathMapping
}

// ApplyMappingToTextProtocol change file path in xDebug text protocol
func (p *Processor) ApplyMappingToTextProtocol(message []byte) []byte {
	p.Logger.Debug("%s", "Call ApplyMappingToTextProtocol in package flowstacktraceprocessor")
	return message
}

// ApplyMappingToXML change file path in xDebug XML protocol
func (p *Processor) ApplyMappingToXML(message []byte) []byte {
	p.Logger.Debug("%s", "Call ApplyMappingToXML in package flowstacktraceprocessor")

	s := strings.Split(string(message), "\x00")

	m, err := mxj.NewMapXml([]byte(s[1]))
	errorhandler.PanicHandling(err, p.Logger)

	message, err = m.Xml()
	errorhandler.PanicHandling(err, p.Logger)

	return message
}

// CanProcess check if this processor support the current configuration
func (p *Processor) CanProcess(c *config.Config) bool {
	return true
}
