package log

import (
	"fmt"
	"github.com/rhaidiz/broxy/core"
	"github.com/rhaidiz/broxy/modules/log/model"
	"github.com/rhaidiz/broxy/core/project/decoder"
)

// Controller represents the controller of the log module
type Controller struct {
	core.ControllerModule
	Module 	*Log
	Gui    	*Gui
	Sess   	*core.Session
	encoder	*decoder.Encoder

	//model     *model.CustomTableModel
	modelSort *model.SortFilterModel
}

// NewController returns a controller of the log module
func NewController(m *Log, g *Gui, s *core.Session) *Controller {
	c := &Controller{
		Module: m,
		Gui:    g,
		Sess:   s,
	}

	//c.model = model.NewCustomTableModel(nil)
	c.modelSort = model.NewSortFilterModel(nil)

	c.Gui.SetTableModel(c.modelSort)

	decoder, err := c.Sess.PersistentProject.FileDecoder2("logs")
	if err != nil {
		panic(fmt.Sprintf("Error while loading log file\n%s",err))
	}
	// load stuff
	for {
		l := &core.Log{}
		if err := decoder.Decode(&l); err != nil {
			break
		}
		c.modelSort.Custom.AddItem(*l)
	}

	encoder, err := c.Sess.PersistentProject.FileEncoder2("logs")
	if err != nil{
		panic(fmt.Sprintf("Error while loading log file\n%s",err))
	}
	c.encoder = &encoder
	go c.logEvent()
	return c
}

// GetGui returns the Gui of the log module
func (c *Controller) GetGui() core.GuiModule {
	return c.Gui
}

// GetModule returns the module of the log module
func (c *Controller) GetModule() core.Module {
	return c.Module
}

// ExecCommand execs commands submitted by other modules
func (c *Controller) ExecCommand(m string, args ...interface{}) {

}

func (c *Controller) logEvent() {
	for l := range c.Sess.LogEvent {
		(*c.encoder).Encode(l)
		c.modelSort.Custom.AddItem(l)
	}
}
