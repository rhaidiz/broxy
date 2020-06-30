package modules

import (
	"github.com/rhaidiz/broxy/core"
	"github.com/rhaidiz/broxy/modules/coreproxy"
	"github.com/rhaidiz/broxy/modules/log"
	"github.com/rhaidiz/broxy/modules/repeater"
)

// LoadModules loads all the main modules
func LoadModules(s *core.Session) {
	s.LoadModule(log.LoadLogModule(s))
	s.LoadModule(repeater.LoadRepeaterModule(s))
	s.LoadModule(coreproxy.LoadCoreProxyModule(s))
}
