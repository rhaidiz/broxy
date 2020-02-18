package repeater

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/rhaidiz/broxy/core"
	"github.com/rhaidiz/broxy/modules/coreproxy/model"
	"github.com/rhaidiz/broxy/util"
)

// RepeaterController represents the controller of the repeater module
type RepeaterController struct {
	core.ControllerModule
	Module *Repeater
	Gui    *RepeaterGui
	Sess   *core.Session
}

// NewRepeaterController creates a new controller for the repeater module
func NewRepeaterController(module *Repeater, gui *RepeaterGui, s *core.Session) *RepeaterController {
	c := &RepeaterController{
		Module: module,
		Gui:    gui,
		Sess:   s,
	}
	c.Gui.GoClick = c.GoClick
	return c
}

// GetGui returns the Gui of the current controller
func (c *RepeaterController) GetGui() core.GuiModule {
	return c.Gui
}

// GetModule returns the module of the current controller
func (c *RepeaterController) GetModule() core.Module {
	return c.Module
}

// ExecCommand execs commands submitted by other modules
func (c *RepeaterController) ExecCommand(m string, args ...interface{}) {
	if m == "send-to" {
		r := args[0].(*model.Request)
		print(r.Host)
		c.Gui.AddNewTab(fmt.Sprintf("%s://%s", r.URL.Scheme, r.Host), fmt.Sprintf("%s\n", r.ToString()))
	}
}

// GoClick is the event fired when clicking the Go button in a repeater tab
func (c *RepeaterController) GoClick(t *RepeaterTab) {
	c.Sess.Debug(c.Module.Name(), "Go pressed")
	rRaw := util.NormalizeRequest(t.RequestEditor.ToPlainText())
	t.RequestEditor.SetPlainText(rRaw)

	r := strings.NewReader(rRaw)
	buf := bufio.NewReader(r)

	req, err := http.ReadRequest(buf)

	if err != nil {
		c.Sess.Err(c.Module.Name(), fmt.Sprintf("ReadRequest %v", err))
		return
	}
	c.Sess.Debug(c.Module.Name(), req.Host)

	url, err := url.Parse(t.HostLine.Text())
	req.URL.Scheme = url.Scheme
	req.URL.Host = url.Host
	req.RequestURI = ""

	go func() {
		resp, err := c.Module.RunRequest(req)

		if err != nil {
			c.Sess.Err(c.Module.Name(), fmt.Sprintf("RunRequest %v", err))
		} else {
			t.ResponseEditor.SetPlainText(util.ResponseToString(resp, false))
		}
	}()
}
