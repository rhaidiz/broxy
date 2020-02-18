package log

import (
	"github.com/rhaidiz/broxy/core"
)

// LoadLogModule loads the log module in the given session
func LoadLogModule(s *core.Session) *Controller {
	m := NewLog(s)
	g := NewGui(s)
	c := NewController(m, g, s)
	return c
}
