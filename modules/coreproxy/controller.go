package coreproxy

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	_ "net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"

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

	interceptor       bool
	interceptRequests bool
	interceptResponse bool
	requestC          chan bool
	nextRequestC      chan bool
	requestC1         chan *http.Request
	queue             int
	forwardC          chan bool
	interceptorC      chan bool
	dropC             chan bool
}

var mutex = &sync.Mutex{}

func NewCoreproxyController(proxy *Coreproxy, proxygui *CoreproxyGui, s *core.Session) *CoreproxyController {
	c := &CoreproxyController{
		Proxy:             proxy,
		Gui:               proxygui,
		Sess:              s,
		isRunning:         false,
		id:                0,
		interceptor:       false,
		interceptRequests: false,
		interceptResponse: false,
		requestC:          make(chan bool),
		nextRequestC:      make(chan bool),
		requestC1:         make(chan *http.Request),
		forwardC:          make(chan bool),
		interceptorC:      make(chan bool),
		dropC:             make(chan bool),
		queue:             0,
	}

	c.model = model.NewSortFilterModel(nil)

	c.Proxy.OnReq = c.OnReq
	c.Proxy.OnResp = c.OnResp

	c.Gui.SetTableModel(c.model)
	c.Gui.StartProxy = c.StartProxy
	c.Gui.RowClicked = c.RowClicked
	c.Gui.Toggle = c.interceptorToggle
	c.Gui.Forward = c.forward
	c.Gui.Drop = c.drop
	return c
}

func (c *CoreproxyController) RowClicked(r int) {
	actual_row := c.model.Index(r, 0, qtcore.NewQModelIndex()).Data(model.ID).ToInt(nil)
	// load the request in the request\response tab
	req, resp := c.model.Custom.GetReqResp(actual_row - 1)
	c.Gui.RequestText.SetPlainText(req.ToString())
	c.Gui.ResponseText.SetPlainText(resp.ToString())
}

func (c *CoreproxyController) interceptorToggle(b bool) {
	if !c.interceptor {
		c.interceptor = true
		//go func() { c.nextRequestC <- true }()
	} else {
		c.interceptor = false
		if c.queue > 0 {
			c.interceptorC <- true
		}
	}
	c.Sess.Debug(c.Proxy.Name(), fmt.Sprintf("Interceptor is: %v", c.interceptor))
}

func (c *CoreproxyController) interceptorActions(req *http.Request, resp *http.Response) (*http.Request, *http.Response) {

	select {
	case <-c.forwardC:
		// pressed forward
		r := strings.NewReader(c.Gui.InterceptorEditor.ToPlainText())
		buf := bufio.NewReader(r)

		req, err := http.ReadRequest(buf)
		if err != nil {
			c.Sess.Err(c.Proxy.Name(), fmt.Sprintf("Error: %p", err))
			return nil, nil
		} else {
			return req, nil
		}
	case <-c.interceptorC:
		// pressed intercetor to turn it off
		r := strings.NewReader(c.Gui.InterceptorEditor.ToPlainText())
		buf := bufio.NewReader(r)

		req, err := http.ReadRequest(buf)
		if err != nil {
			c.Sess.Err(c.Proxy.Name(), fmt.Sprintf("Error: %p", err))
			return nil, nil
		} else {
			return req, nil
		}
	case <-c.dropC:
		// pressed drop
		return req, goproxy.NewResponse(req,
			goproxy.ContentTypeText, http.StatusForbidden,
			"Request droppped")
	}
}

func (c *CoreproxyController) forward(b bool) {
	go func() {
		// activate only if there's something waiting
		if c.queue > 0 {
			c.Sess.Debug(c.Proxy.Name(), "pressing forward")
			c.forwardC <- true
		}
	}()
}

func (c *CoreproxyController) drop(b bool) {
	go func() {
		// activate only if there's something waiting
		if c.queue > 0 {
			c.Sess.Debug(c.Proxy.Name(), "pressing drop")
			c.dropC <- true
		}
	}()
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
	// activate the interceptor

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

func (c *CoreproxyController) OnReq(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	c.queue = c.queue + 1
	var resp *http.Response
	var bodyBytes []byte
	if r != nil {
		bodyBytes, _ = ioutil.ReadAll(r.Body)
		// Restore the io.ReadCloser to its original state
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	}
	req := model.NewHItem(nil)
	c.id = c.id + 1
	req.ID = c.id

	// this is the original request, I save it before tampering with it
	req.Req = &model.Request{Path: r.URL.Path, Schema: r.URL.Scheme, Method: r.Method, Body: bodyBytes, Host: r.Host, ContentLength: r.ContentLength, Headers: r.Header, Proto: r.Proto}
	//c.model.Add()

	// activate interceptor
	if c.interceptor {
		fmt.Printf("(1) waiting in queue %p\n", c.queue)
		mutex.Lock()
		c.Gui.InterceptorEditor.SetPlainText(req.Req.ToString() + "\n")
		// now wait until a decision is made
		tmp_scheme := r.URL.Scheme
		tmp_host := r.URL.Host
		fmt.Printf("(2) waiting in queue %p\n", c.queue)
		for r, resp = c.interceptorActions(r, nil); r == nil; r, resp = c.interceptorActions(r, nil) {
			// doesn't look good but makes sence ... I guess
			// continue to perform intercetorAction until I don't get nil as response
		}
		// reset scheme and host
		fmt.Printf("host %p\n", tmp_host)
		fmt.Printf("scheme %p\n", tmp_scheme)
		fmt.Printf("r.URL %p\n", r.URL)
		r.URL.Scheme = tmp_scheme
		r.URL.Host = tmp_host
		r.RequestURI = ""
		c.queue = c.queue - 1
		c.Gui.InterceptorEditor.SetPlainText("")
		mutex.Unlock()
	}

	// add the request to the history only at the end
	c.model.Custom.AddItem(req, ctx.Session)
	return r, resp
}
