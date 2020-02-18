package log

import (
	"github.com/rhaidiz/broxy/core"
)

// LoadLogModule loads the log module in the given session
func LoadLogModule(s *core.Session) *LogController {
	m := NewLog(s)
	g := NewLogGui(s)
	c := NewLogController(m, g, s)
	return c
}
