package pathmapperfactory

import (
	"github.com/dfeyer/flow-debugproxy/config"
	"github.com/dfeyer/flow-debugproxy/flowpathmapper"
	"github.com/dfeyer/flow-debugproxy/xdebugproxy"

	"errors"
)

const (
	flow  = "flow"
	goaop = "goaop"
)

// Create return a pathmapper for the given framework
func Create(c *config.Config) (xdebugproxy.XDebugProcessorPlugin, error) {
	var pathmapper xdebugproxy.XDebugProcessorPlugin
	switch {
	case c.Framework == flow:
		pathmapper = &flowpathmapper.PathMapper{
			Config: c,
		}
		return pathmapper, nil
	case c.Framework == goaop:
		return nil, errors.New("Go! AOP Framework support is under development ...")
	}

	return nil, errors.New("Unsupported framework")
}
