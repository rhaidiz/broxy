package log

import (
	"github.com/rhaidiz/broxy/core"
)

func LoadLogModule(s *core.Session) (*Log, *LogGui) {
	m := NewLog(s)
	g := NewLogGui(s)
	NewLogController(m, g, s)
	return m, g
}
