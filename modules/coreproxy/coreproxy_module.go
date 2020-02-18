package coreproxy

import (
	"github.com/rhaidiz/broxy/core"
)

// LoadCoreProxyModule loads the core proxy module in the given session
func LoadCoreProxyModule(s *core.Session) *Controller {
	m := NewCoreProxy(s)
	g := NewGui(s)
	c := NewController(m, g, s)
	return c
}
