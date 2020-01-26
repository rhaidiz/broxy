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
		c.Gui.AddNewTab(fmt.Sprintf("%s://%s", r.Url.Scheme, r.Host), fmt.Sprintf("%s\n", r.ToString()))
	}
}

func (c *RepeaterController) GoClick(t *RepeaterTab) {
	c.Sess.Debug(c.Module.Name(), "Go pressed")
	r_raw := util.NormalizeRequest(t.RequestEditor.ToPlainText())
	t.RequestEditor.SetPlainText(r_raw)

	r := strings.NewReader(r_raw)
	buf := bufio.NewReader(r)

	req, err := http.ReadRequest(buf)

	if err != nil {
		c.Sess.Err(c.Module.Name(), fmt.Sprintf("ReadRequest %v", err))
		return
	} else {
		c.Sess.Debug(c.Module.Name(), req.Host)
	}

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
