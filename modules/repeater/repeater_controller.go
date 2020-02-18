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

// Controller represents the controller of the repeater module
type Controller struct {
	core.ControllerModule
	Module *Repeater
	Gui    *Gui
	Sess   *core.Session
}

// NewController creates a new controller for the repeater module
func NewController(module *Repeater, gui *Gui, s *core.Session) *Controller {
	c := &Controller{
		Module: module,
		Gui:    gui,
		Sess:   s,
	}
	c.Gui.GoClick = c.GoClick
	return c
}

// GetGui returns the Gui of the current controller
func (c *Controller) GetGui() core.GuiModule {
	return c.Gui
}

// GetModule returns the module of the current controller
func (c *Controller) GetModule() core.Module {
	return c.Module
}

// ExecCommand execs commands submitted by other modules
func (c *Controller) ExecCommand(m string, args ...interface{}) {
	if m == "send-to" {
		r := args[0].(*model.Request)
		print(r.Host)
		c.Gui.AddNewTab(fmt.Sprintf("%s://%s", r.URL.Scheme, r.Host), fmt.Sprintf("%s\n", r.ToString()))
	}
}

// GoClick is the event fired when clicking the Go button in a repeater tab
func (c *Controller) GoClick(t *Tab) {
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
