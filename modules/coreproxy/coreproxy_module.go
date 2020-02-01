package coreproxy

import (
	"github.com/rhaidiz/broxy/core"
)

func LoadCoreProxyModule(s *core.Session) *CoreproxyController {
	m := NewCoreProxy(s)
	g := NewCoreproxyGui(s)
	c := NewCoreproxyController(m, g, s)
	return c
}
