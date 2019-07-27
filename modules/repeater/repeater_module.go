package repeater

import (
	"github.com/rhaidiz/broxy/core"
)

func LoadRepeaterModule(s *core.Session) (*Repeater, *RepeaterGui) {
	m := NewRepeater(s)
	g := NewRepeaterGui(s)
	NewRepeaterController(m, g, s)
	return m, g
}
