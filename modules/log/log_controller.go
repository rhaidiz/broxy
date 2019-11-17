package log

import (
	"github.com/rhaidiz/broxy/core"
	"github.com/rhaidiz/broxy/modules/log/model"
	_ "github.com/therecipe/qt/core"
)

type LogController struct {
	core.ControllerModule
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

func (c *LogController) GetGui() core.GuiModule {
	return c.Gui
}

func (c *LogController) GetModule() core.Module {
	return c.Module
}

func (c *LogController) Name() string {
	return "log"
}

func (c *LogController) ExecCommand(m string, args ...interface{}) {

}

func (c *LogController) logEvent() {
	for l := range c.Sess.LogC {
		c.modelSort.Custom.AddItem(l)
	}
}
