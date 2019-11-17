package log

import (
	"github.com/rhaidiz/broxy/core"
)

func LoadLogModule(s *core.Session) *LogController {
	m := NewLog(s)
	g := NewLogGui(s)
	c := NewLogController(m, g, s)
	return c
}
