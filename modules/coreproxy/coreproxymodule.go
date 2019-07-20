package coreproxy

import (
	"github.com/rhaidiz/broxy/core"
)

func LoadCoreProxyModule(s *core.Session) (*Coreproxy, *CoreproxyGui) {
	m := NewCoreProxy(s)
	g := NewCoreproxyGui(s)
	NewCoreproxyController(m, g, s)
	return m, g
}
