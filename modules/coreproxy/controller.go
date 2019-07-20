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
	"io/ioutil"
)

//type qController struct {
//	core.QObject
//
//	// signals
//	_ func(message string) 			`signal:"ErrorMsg"`
//}

type CoreproxyController struct {
	Proxy *Coreproxy
	Gui   *CoreproxyGui
	Sess  *core.Session
	//QController	*qController

	isRunning bool
	model     *model.CustomTableModel

	_ func() `signal:"mySignal"`
}

func NewCoreproxyController(proxy *Coreproxy, proxygui *CoreproxyGui, s *core.Session) *CoreproxyController {
	c := &CoreproxyController{
		Proxy: proxy,
		Gui:   proxygui,
		Sess:  s,
		//QController	: NewQController(nil),
		isRunning: false,
	}

	c.model = model.NewCustomTableModel(nil)

	c.Proxy.OnReq = c.OnReq
	c.Proxy.OnResp = c.OnResp

	c.Gui.SetTableModel(c.model)
	c.Gui.StartProxy = c.StartProxy
	c.Gui.RowClicked = c.RowClicked
	return c
}

func (c *CoreproxyController) RowClicked(r int) {
	// load the request in the request\response tab
	req, resp := c.model.GetReqResp(r)
	c.Gui.RequestText.SetPlainText(req.ToString())
	c.Gui.ResponseText.SetPlainText(resp.ToString())
}

func (c *CoreproxyController) StartProxy(b bool) {
	if !c.isRunning {
		// Start and stop the proxy
		// fmt.Printf("Line: %s\n", c.Gui.ListenerLineEdit.DisplayText())
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
					fmt.Println("Proxy running")
					if e := c.Proxy.Start(); e != nil && e != http.ErrServerClosed {
						fmt.Printf("Error %s\n", e)
						//c.Maingui.Guicon.ShowErrorMsg(e.Error())
						//c.QController.ErrorMsg(e.Error())
						c.isRunning = false
						c.Gui.StartStopBtn.SetText("Start")
					}
				}()
			} else {
				fmt.Printf("Error starting the proxy %s\n", e)
				//c.Maingui.Guicon.ShowErrorMsg(e.Error())
				//c.QController.ErrorMsg(e.Error())
			}
		} else {
			//c.Maingui.Guicon.ShowErrorMsg("Wrong input")
			//c.QController.ErrorMsg("Wrong input")
		}
	} else {
		c.Proxy.Stop()
		c.isRunning = false
		fmt.Println("Proxy stopped")
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
	c.model.EditItem(item, ctx.Session)
}

func (c *CoreproxyController) OnReq(r *http.Request, ctx *goproxy.ProxyCtx) {
	var bodyBytes []byte
	if r != nil {
		bodyBytes, _ = ioutil.ReadAll(r.Body)
		// Restore the io.ReadCloser to its original state
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	}
	req := model.NewHItem(nil)
	req.Req = &model.Request{Path: r.URL.Path, Schema: r.URL.Scheme, Method: r.Method, Body: bodyBytes, Host: r.Host, ContentLength: r.ContentLength, Headers: r.Header, Proto: r.Proto}
	//c.model.Add()
	c.model.AddItem(req, ctx.Session)
}
