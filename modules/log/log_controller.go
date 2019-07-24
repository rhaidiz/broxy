package log

import (
	"github.com/rhaidiz/broxy/core"
	"github.com/rhaidiz/broxy/modules/log/model"
	_ "github.com/therecipe/qt/core"
)

type LogController struct {
	Module *Log
	Gui    *LogGui
	Sess   *core.Session

	//model     *model.CustomTableModel
	modelSort *model.SortFilterModel
}

func NewLogController(m *Log, g *LogGui, s *core.Session) *LogController {
	c := &LogController{
		Module: m,
		Gui:    g,
		Sess:   s,
	}

	//c.model = model.NewCustomTableModel(nil)
	c.modelSort = model.NewSortFilterModel(nil)

	c.Gui.SetTableModel(c.modelSort)
	go c.logEvent()
	return c
}

func (c *LogController) logEvent() {
	for l := range c.Sess.LogC {
		c.modelSort.Custom.AddItem(l)
	}
}
