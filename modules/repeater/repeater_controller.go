package repeater

import (
	"bufio"
	"fmt"
	"github.com/rhaidiz/broxy/core"
	"github.com/rhaidiz/broxy/modules/coreproxy/model"
	"github.com/rhaidiz/broxy/util"
	"net/http"
	"net/url"
	"strings"
)

type RepeaterController struct {
	core.ControllerModule
	Module *Repeater
	Gui    *RepeaterGui
	Sess   *core.Session
}

func NewRepeaterController(module *Repeater, gui *RepeaterGui, s *core.Session) *RepeaterController {
	c := &RepeaterController{
		Module: module,
		Gui:    gui,
		Sess:   s,
	}
	c.Gui.GoClick = c.GoClick
	return c
}

func (c *RepeaterController) GetGui() core.GuiModule {
	return c.Gui
}

func (c *RepeaterController) GetModule() core.Module {
	return c.Module
}

func (c *RepeaterController) Name() string {
	return "repeater"
}

func (c *RepeaterController) ExecCommand(m string, args ...interface{}) {
	if m == "send-to" {
		r := args[0].(*model.Request)
		print(r.Host)
		c.Gui.AddNewTab(fmt.Sprintf("%s://%s", r.Schema, r.Host), fmt.Sprintf("%s\n", r.ToString()))
	}
}

func (c *RepeaterController) GoClick(b bool) {
	c.Sess.Debug(c.Module.Name(), "Go pressed")

	r_raw := util.NormalizeRequest(c.Gui.RequestEditor.ToPlainText())
	c.Gui.RequestEditor.SetPlainText(r_raw)

	r := strings.NewReader(r_raw)
	buf := bufio.NewReader(r)

	req, err := http.ReadRequest(buf)

	if err != nil {
		c.Sess.Err(c.Module.Name(), fmt.Sprintf("ReadRequest %v", err))
		return
	} else {
		c.Sess.Debug(c.Module.Name(), req.Host)
	}

	url, err := url.Parse(c.Gui.HostLine.Text())
	req.URL.Scheme = url.Scheme
	req.URL.Host = url.Host
	req.RequestURI = ""

	go func() {
		resp, err := c.Module.RunRequest(req)

		if err != nil {
			c.Sess.Err(c.Module.Name(), fmt.Sprintf("RunRequest %v", err))
		} else {
			c.Gui.ResponseEditor.SetPlainText(util.ResponseToString(resp, false))
		}
	}()
}
