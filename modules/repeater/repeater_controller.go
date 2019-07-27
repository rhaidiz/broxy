package repeater

import (
	"bufio"
	"fmt"
	"github.com/rhaidiz/broxy/core"
	"github.com/rhaidiz/broxy/util"
	"net/http"
	"net/url"
	"strings"
)

type RepeaterController struct {
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

func (c *RepeaterController) GoClick(b bool) {
	c.Sess.Debug(c.Module.Name(), "Go pressed")
	r := strings.NewReader(c.Gui.RequestEditor.ToPlainText())
	buf := bufio.NewReader(r)

	req, err := http.ReadRequest(buf)

	if err != nil {
		c.Sess.Debug(c.Module.Name(), fmt.Sprintf("errore %v", err))
		return
	} else {
		c.Sess.Debug(c.Module.Name(), req.Host)
	}

	url, err := url.Parse(c.Gui.HostLine.Text())
	req.URL = url
	req.RequestURI = ""

	go func() {
		resp, err := c.Module.RunRequest(req)

		if err != nil {
			c.Sess.Debug(c.Module.Name(), fmt.Sprintf("errore %v", err))
		} else {
			c.Gui.ResponseEditor.SetPlainText(util.ResponseToString(resp))
		}
	}()
}
