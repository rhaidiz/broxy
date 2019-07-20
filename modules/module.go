package modules

import (
	"github.com/rhaidiz/broxy/core"
	"github.com/rhaidiz/broxy/modules/coreproxy"
	"github.com/rhaidiz/broxy/modules/repeater"
)

func LoadModules(s *core.Session) {
	s.LoadModule(coreproxy.LoadCoreProxyModule(s))
	s.LoadModule(repeater.NewRepeater(s), repeater.NewRepeaterGui(s))
}
