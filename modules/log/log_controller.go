package log

import (
	"github.com/rhaidiz/broxy/core"
	"github.com/rhaidiz/broxy/modules/log/model"
)

// LogController represents the controller of the log module
type LogController struct {
	core.ControllerModule
	Module *Log
	Gui    *LogGui
	Sess   *core.Session

	//model     *model.CustomTableModel
	modelSort *model.SortFilterModel
}

// NewLogController returns a controller of the log module
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

// GetGui returns the Gui of the log module
func (c *LogController) GetGui() core.GuiModule {
	return c.Gui
}

// GetModule returns the module of the log module
func (c *LogController) GetModule() core.Module {
	return c.Module
}

// ExecCommand execs commands submitted by other modules
func (c *LogController) ExecCommand(m string, args ...interface{}) {

}

func (c *LogController) logEvent() {
	for l := range c.Sess.LogC {
		c.modelSort.Custom.AddItem(l)
	}
}
