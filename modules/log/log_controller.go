package log

import (
	"github.com/rhaidiz/broxy/core"
	"github.com/rhaidiz/broxy/modules/log/model"
)

type LogController struct {
	Module *Log
	Gui    *LogGui
	Sess   *core.Session

	model *model.CustomTableModel
}

func NewLogController(m *Log, g *LogGui, s *core.Session) *LogController {
	c := &LogController{
		Module: m,
		Gui:    g,
		Sess:   s,
	}

	c.model = model.NewCustomTableModel(nil)

	c.Gui.SetTableModel(c.model)
	go c.logEvent()
	return c
}

func (g *LogController) logEvent() {
	for l := range g.Sess.LogC {
		g.model.AddItem(l)
	}
}
