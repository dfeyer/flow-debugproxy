package pathmapperfactory

import (
	"github.com/dfeyer/flow-debugproxy/config"
	"github.com/dfeyer/flow-debugproxy/dummypathmapper"
	"github.com/dfeyer/flow-debugproxy/flowpathmapper"
	"github.com/dfeyer/flow-debugproxy/logger"
	"github.com/dfeyer/flow-debugproxy/pathmapping"
	"github.com/dfeyer/flow-debugproxy/xdebugproxy"

	"errors"
)

const (
	flow  = "flow"
	goaop = "goaop"
	dummy = "dummy"
)

// Create return a pathmapper for the given framework
func Create(c *config.Config, p *pathmapping.PathMapping, l *logger.Logger) (xdebugproxy.XDebugProcessorPlugin, error) {
	var pathmapper xdebugproxy.XDebugProcessorPlugin
	switch {
	case c.Framework == flow:
		pathmapper = &flowpathmapper.PathMapper{
			Config:      c,
			Logger:      l,
			PathMapping: p,
		}
		return pathmapper, nil
	case c.Framework == dummy:
		pathmapper = &dummypathmapper.PathMapper{
			Config:      c,
			Logger:      l,
			PathMapping: p,
		}
		return pathmapper, nil
	case c.Framework == goaop:
		return nil, errors.New("Go! AOP Framework support is under development ...")
	}

	return nil, errors.New("Unsupported framework")
}
