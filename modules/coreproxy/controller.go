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
	Module *Coreproxy
	Gui    *CoreproxyGui
	Sess   *core.Session

	isRunning bool
	model     *model.SortFilterModel
	id        int

	// interceptor
	interceptor_status  bool
	intercept_requests  bool
	intercept_responses bool
	forward_chan        chan bool
	drop_chan           chan bool
	// will maintain the number of requests in queue
	requests_queue  int
	responses_queue int

	// Qt signals
	_ func() `signal:"mySignal"`
}

var mutex = &sync.Mutex{}

func NewCoreproxyController(proxy *Coreproxy, proxygui *CoreproxyGui, s *core.Session) *CoreproxyController {
	c := &CoreproxyController{
		Module:              proxy,
		Gui:                 proxygui,
		Sess:                s,
		isRunning:           false,
		id:                  0,
		interceptor_status:  false,
		intercept_requests:  true,
		intercept_responses: true,
		forward_chan:        make(chan bool),
		drop_chan:           make(chan bool),
		requests_queue:      0,
		responses_queue:     0,
	}

	c.model = model.NewSortFilterModel(nil)

	c.Module.OnReq = c.OnReq
	c.Module.OnResp = c.OnResp

	c.Gui.SetTableModel(c.model)
	c.Gui.StartProxy = c.startProxy
	c.Gui.RowClicked = c.selectRow
	c.Gui.Toggle = c.interceptorToggle
	c.Gui.Forward = c.forward
	c.Gui.Drop = c.drop
	return c
}

// buttons logic

func (c *CoreproxyController) selectRow(r int) {
	actual_row := c.model.Index(r, 0, qtcore.NewQModelIndex()).Data(model.ID).ToInt(nil)
	// load the request in the request\response tab
	req, resp := c.model.Custom.GetReqResp(actual_row - 1)
	if req != nil {
		c.Gui.RequestText.SetPlainText(req.ToString())
	}
	if resp != nil && resp.ContentLength >= 1e+8 {
		c.Gui.ResponseText.SetPlainText("Response too big")
	} else if resp != nil {
		c.Gui.ResponseText.SetPlainText(resp.ToString())
	}
}

func (c *CoreproxyController) interceptorToggle(b bool) {
	if !c.interceptor_status {
		c.interceptor_status = true
	} else {
		c.interceptor_status = false
		if c.requests_queue > 0 || c.responses_queue > 0 {
			tmp := c.requests_queue + c.responses_queue
			for i := 0; i < tmp; i++ {
				//fmt.Printf("interceptor waiting: %d\n", tmp)
				c.forward_chan <- true
			}
		}
	}
	c.Sess.Debug(c.Module.Name(), fmt.Sprintf("Interceptor is: %v", c.interceptor_status))
}

func (c *CoreproxyController) forward(b bool) {
	go func() {
		// activate only if there's something waiting
		if c.requests_queue > 0 || c.responses_queue > 0 {
			c.forward_chan <- true
		}
	}()
}

func (c *CoreproxyController) drop(b bool) {
	go func() {
		// activate only if there's something waiting
		if c.requests_queue > 0 || c.responses_queue > 0 {
			c.drop_chan <- true
		}
	}()
}

func (c *CoreproxyController) startProxy(b bool) {
	if !c.isRunning {
		// Start and stop the proxy
		ip_port_regxp := "^((?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?).){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)):(6553[0-5]|655[0-2][0-9]|65[0-4][0-9][0-9]|6[0-4][0-9][0-9][0-9]|[1-5]?[0-9]?[0-9]?[0-9]?[0-9])?$"

		r := regexp.MustCompile(ip_port_regxp)

		if s := r.FindStringSubmatch(c.Gui.ListenerLineEdit.DisplayText()); s != nil {
			p, _ := strconv.Atoi(s[2])
			if e := c.Module.ChangeIpPort(s[1], p); e == nil {
				// if I can change ip and port, change it also in the config struct
				c.Sess.Config.Address = s[1]
				c.Sess.Config.Port = p
				go func() {
					c.Gui.StartStopBtn.SetText("Stop")
					c.isRunning = true
					c.Sess.Info(c.Module.Name(), "Starting proxy ...")
					if e := c.Module.Start(); e != nil && e != http.ErrServerClosed {
						c.Sess.Err(c.Module.Name(), fmt.Sprintf("Error starting the proxy %s\n", e))
						c.isRunning = false
						c.Gui.StartStopBtn.SetText("Start")
					}
				}()
			} else {
				c.Sess.Err(c.Module.Name(), fmt.Sprintf("Error starting the proxy %s\n", e))
			}
		} else {
			c.Sess.Err(c.Module.Name(), "Wrong input")
		}
	} else {
		if c.interceptor_status {
			c.interceptorToggle(false)
		}
		c.Module.Stop()
		c.isRunning = false
		c.Sess.Info(c.Module.Name(), "Stopping proxy.")
		c.Gui.StartStopBtn.SetText("Start")
	}
}

// Executed when a response arrives
func (c *CoreproxyController) OnResp(r *http.Response, ctx *goproxy.ProxyCtx) *http.Response {

	item := model.NewHItem(nil)
	var bodyBytes []byte
	if r != nil {
		bodyBytes, _ = ioutil.ReadAll(r.Body)
		// Restore the io.ReadCloser to its original state
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	}
	item.Resp = &model.Response{Status: r.Status, Body: bodyBytes, Proto: r.Proto, ContentLength: int64(len(bodyBytes)), Headers: r.Header}
	// activate interceptor
	if c.interceptor_status && c.intercept_responses {
		// increase the requests in queue
		//fmt.Printf("Responses waiting: %d\n", c.responses_queue)
		c.responses_queue = c.responses_queue + 1
		mutex.Lock()
		// if response is bigger than 100mb, show message that is not supported
		if r.ContentLength >= 1e+8 {
			c.Gui.InterceptorEditor.SetPlainText("Response too big")
		} else {
			c.Gui.InterceptorEditor.SetPlainText(item.Resp.ToString())
		}
		// now wait until a decision is made
		for r = c.interceptorResponseActions(nil, r); r == nil; r = c.interceptorResponseActions(nil, r) {
			// doesn't look good but makes sence ... I guess
			// continue to perform intercetorAction until I don't get nil as response
		}
		// decrease the requests in queue
		c.responses_queue = c.responses_queue - 1
		// rest the editor
		c.Gui.InterceptorEditor.SetPlainText("")
		mutex.Unlock()
	}
	// For whatever reason, I have to send a full HItem insteam of a Resp
	c.model.Custom.EditItem(item, ctx.Session)

	return r
}

// Executed when a request arrives
func (c *CoreproxyController) OnReq(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
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

	// activate interceptor
	if c.interceptor_status && c.intercept_requests {
		// increase the requests in queue
		c.requests_queue = c.requests_queue + 1
		mutex.Lock()
		c.Gui.InterceptorEditor.SetPlainText(req.Req.ToString() + "\n")
		// now wait until a decision is made
		for r, resp = c.interceptorRequestActions(r, nil); r == nil; r, resp = c.interceptorRequestActions(r, nil) {
			// doesn't look good but makes sence ... I guess
			// continue to perform intercetorAction until I don't get nil as response
		}
		// decrease the requests in queue
		c.requests_queue = c.requests_queue - 1
		// rest the editor
		c.Gui.InterceptorEditor.SetPlainText("")
		mutex.Unlock()
	}

	// add the request to the history only at the end
	c.model.Custom.AddItem(req, ctx.Session)
	return r, resp
}

func (c *CoreproxyController) interceptorRequestActions(req *http.Request, resp *http.Response) (*http.Request, *http.Response) {

	select {
	case <-c.forward_chan:
		if !c.interceptor_status {
			return req, nil
		}
		// pressed forward
		reader := strings.NewReader(c.Gui.InterceptorEditor.ToPlainText())
		buf := bufio.NewReader(reader)

		r, err := http.ReadRequest(buf)
		if err != nil {
			c.Sess.Err(c.Module.Name(), fmt.Sprintf("Forward Req: %s", err.Error()))
			return nil, nil
		}
		r.URL.Scheme = req.URL.Scheme
		r.URL.Host = req.URL.Host
		r.RequestURI = ""
		return r, nil
	case <-c.drop_chan:
		// pressed drop
		return req, goproxy.NewResponse(req,
			goproxy.ContentTypeText, http.StatusForbidden, "Request droppped")
	}
}

func (c *CoreproxyController) interceptorResponseActions(req *http.Request, resp *http.Response) *http.Response {

	select {
	case <-c.forward_chan:
		if !c.interceptor_status {
			return resp
		}
		// if response is bigger than 100mb, just don't process the text editor
		if resp.ContentLength >= 1e+8 {
			return resp
		}
		// pressed forward
		reader := strings.NewReader(c.Gui.InterceptorEditor.ToPlainText())
		buf := bufio.NewReader(reader)

		resp, err := http.ReadResponse(buf, nil)
		if err != nil {
			//c.Sess.Err(c.Module.Name(), fmt.Sprintf("Forward Resp: %s", err.Error()))
			print(err)
			return nil
		} else {
			return resp
		}
	case <-c.drop_chan:
		// pressed drop
		resp.Body = ioutil.NopCloser(bytes.NewReader([]byte("Response dropped by user")))
		return resp
	}
}
