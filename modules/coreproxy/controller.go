package coreproxy

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/elazarl/goproxy"
	"github.com/rhaidiz/broxy/core"
	"github.com/rhaidiz/broxy/modules/coreproxy/model"
	qtcore "github.com/therecipe/qt/core"
	"io/ioutil"
)

type CoreproxyController struct {
	Proxy *Coreproxy
	Gui   *CoreproxyGui
	Sess  *core.Session

	isRunning bool
	model     *model.SortFilterModel
	id        int

	_ func() `signal:"mySignal"`
}

func NewCoreproxyController(proxy *Coreproxy, proxygui *CoreproxyGui, s *core.Session) *CoreproxyController {
	c := &CoreproxyController{
		Proxy:     proxy,
		Gui:       proxygui,
		Sess:      s,
		isRunning: false,
		id:        0,
	}

	c.model = model.NewSortFilterModel(nil)

	c.Proxy.OnReq = c.OnReq
	c.Proxy.OnResp = c.OnResp

	c.Gui.SetTableModel(c.model)
	c.Gui.StartProxy = c.StartProxy
	c.Gui.RowClicked = c.RowClicked
	return c
}

func (c *CoreproxyController) RowClicked(r int) {
	actual_row := c.model.Index(r, 0, qtcore.NewQModelIndex()).Data(model.ID).ToInt(nil)
	// load the request in the request\response tab
	req, resp := c.model.Custom.GetReqResp(actual_row - 1)
	c.Gui.RequestText.SetPlainText(req.ToString())
	c.Gui.ResponseText.SetPlainText(resp.ToString())
}

func (c *CoreproxyController) StartProxy(b bool) {
	if !c.isRunning {
		// Start and stop the proxy
		ip_port_regxp := "^((?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?).){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)):(6553[0-5]|655[0-2][0-9]|65[0-4][0-9][0-9]|6[0-4][0-9][0-9][0-9]|[1-5]?[0-9]?[0-9]?[0-9]?[0-9])?$"

		r := regexp.MustCompile(ip_port_regxp)

		if s := r.FindStringSubmatch(c.Gui.ListenerLineEdit.DisplayText()); s != nil {
			p, _ := strconv.Atoi(s[2])
			if e := c.Proxy.ChangeIpPort(s[1], p); e == nil {
				// if I can change ip and port, change it also in the config struct
				c.Sess.Config.Address = s[1]
				c.Sess.Config.Port = p
				go func() {
					c.Gui.StartStopBtn.SetText("Stop")
					c.isRunning = true
					c.Sess.Info(c.Proxy.Name(), "Starting proxy ...")
					if e := c.Proxy.Start(); e != nil && e != http.ErrServerClosed {
						c.Sess.Err(c.Proxy.Name(), fmt.Sprintf("Error starting the proxy %s\n", e))
						c.isRunning = false
						c.Gui.StartStopBtn.SetText("Start")
					}
				}()
			} else {
				c.Sess.Err(c.Proxy.Name(), fmt.Sprintf("Error starting the proxy %s\n", e))
			}
		} else {
			c.Sess.Err(c.Proxy.Name(), "Wrong input")
		}
	} else {
		c.Proxy.Stop()
		c.isRunning = false
		c.Sess.Info(c.Proxy.Name(), "Stopping proxy.")
		c.Gui.StartStopBtn.SetText("Start")
	}
}

func (c *CoreproxyController) OnResp(r *http.Response, ctx *goproxy.ProxyCtx) {
	item := model.NewHItem(nil)
	var bodyBytes []byte
	if r != nil {
		bodyBytes, _ = ioutil.ReadAll(r.Body)
		// Restore the io.ReadCloser to its original state
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	}
	item.Resp = &model.Response{Status: r.Status, Body: bodyBytes, Proto: r.Proto, ContentLength: r.ContentLength, Headers: r.Header}
	// For whatever reason, I have to send a full HItem insteam of a Resp
	c.model.Custom.EditItem(item, ctx.Session)
}

func (c *CoreproxyController) OnReq(r *http.Request, ctx *goproxy.ProxyCtx) {
	var bodyBytes []byte
	if r != nil {
		bodyBytes, _ = ioutil.ReadAll(r.Body)
		// Restore the io.ReadCloser to its original state
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	}
	req := model.NewHItem(nil)
	c.id = c.id + 1
	req.ID = c.id
	req.Req = &model.Request{Path: r.URL.Path, Schema: r.URL.Scheme, Method: r.Method, Body: bodyBytes, Host: r.Host, ContentLength: r.ContentLength, Headers: r.Header, Proto: r.Proto}
	//c.model.Add()
	c.model.Custom.AddItem(req, ctx.Session)
}
