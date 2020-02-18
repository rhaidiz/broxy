package repeater

import (
	"github.com/rhaidiz/broxy/core"
)

// LoadRepeaterModule loads the repeater module in the given session
func LoadRepeaterModule(s *core.Session) *RepeaterController {
	m := NewRepeater(s)
	g := NewRepeaterGui(s)
	c := NewRepeaterController(m, g, s)
	return c
}
