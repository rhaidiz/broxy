package repeater

import (
	"github.com/rhaidiz/broxy/core"
)

// LoadRepeaterModule loads the repeater module in the given session
func LoadRepeaterModule(s *core.Session) *Controller {
	m := NewRepeater(s)
	g := NewGui(s)
	c := NewController(m, g, s)
	return c
}
