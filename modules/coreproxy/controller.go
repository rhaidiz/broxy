package coreproxy

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	_ "net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/elazarl/goproxy"
	"github.com/rhaidiz/broxy/core"
	"github.com/rhaidiz/broxy/modules/coreproxy/model"
	"github.com/rhaidiz/broxy/util"
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
	c.Gui.HideAllTabs()
	actual_row := c.model.Index(r, 0, qtcore.NewQModelIndex()).Data(model.ID).ToInt(nil)
	req, edited_req, resp, edited_resp := c.model.Custom.GetReqResp(actual_row - 1)
	if req != nil {
		c.Gui.ShowReqTab(req.ToString())
	}
	if edited_req != nil {
		c.Gui.ShowEditedReqTab(edited_req.ToString())
	}
	if resp != nil {
		c.Gui.ShowRespTab(resp.ToString())
	}
	if edited_resp != nil {
		c.Gui.ShowEditedRespTab(edited_resp.ToString())
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

	http_item := model.NewHttpItem(nil)

	var bodyBytes []byte
	if r != nil {
		bodyBytes, _ = ioutil.ReadAll(r.Body)
		// Restore the io.ReadCloser to its original state
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	}
	http_item.Resp = &model.Response{Status: r.Status,
		Body:          bodyBytes,
		Proto:         r.Proto,
		ContentLength: int64(len(bodyBytes)),
		Headers:       r.Header,
	}
	// activate interceptor
	if c.interceptor_status && c.intercept_responses {
		// if the response is nil, it means the interceptor did not change the response

		r.ContentLength = int64(len(bodyBytes))
		edited_resp := c.interceptorResponseActions(nil, r)
		// the response was edited
		if edited_resp != nil {
			var edited_bodyBytes []byte
			edited_bodyBytes, _ = ioutil.ReadAll(edited_resp.Body)
			edited_resp.Body = ioutil.NopCloser(bytes.NewBuffer(edited_bodyBytes))
			http_item.EditedResp = &model.Response{
				Status:        edited_resp.Status,
				Proto:         edited_resp.Proto,
				Body:          edited_bodyBytes,
				ContentLength: int64(len(edited_bodyBytes)),
				Headers:       edited_resp.Header,
			}
			r = edited_resp
		}
	}

	// TODO: [BUG] For whatever reason, I have to use a full HttpItem insteam of a Resp
	c.model.Custom.EditItem(http_item, ctx.Session)

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
	http_item := model.NewHttpItem(nil)
	c.id = c.id + 1
	http_item.ID = c.id

	// this is the original request, save it for the history
	http_item.Req = &model.Request{
		Path:          r.URL.Path,
		Schema:        r.URL.Scheme,
		Method:        r.Method,
		Body:          bodyBytes,
		Host:          r.Host,
		ContentLength: r.ContentLength,
		Headers:       r.Header,
		Proto:         r.Proto,
	}

	// activate interceptor
	if c.interceptor_status && c.intercept_requests {

		edited_req, edited_resp := c.interceptorRequestActions(r, nil)

		if edited_req != nil {
			var edited_bodyBytes []byte
			edited_bodyBytes, _ = ioutil.ReadAll(edited_req.Body)
			edited_req.Body = ioutil.NopCloser(bytes.NewBuffer(edited_bodyBytes))

			http_item.EditedReq = &model.Request{
				Path:          edited_req.URL.Path,
				Schema:        edited_req.URL.Scheme,
				Method:        edited_req.Method,
				Body:          edited_bodyBytes,
				Host:          edited_req.Host,
				ContentLength: edited_req.ContentLength,
				Headers:       edited_req.Header,
				Proto:         edited_req.Proto,
			}
			r = edited_req
			resp = edited_resp

		}

	}

	// add the request to the history only at the end
	c.model.Custom.AddItem(http_item, ctx.Session)

	return r, resp
}

func (c *CoreproxyController) interceptorRequestActions(req *http.Request, resp *http.Response) (*http.Request, *http.Response) {

	// the request to return
	var _req *http.Request
	var _resp *http.Response

	c.requests_queue = c.requests_queue + 1
	mutex.Lock()
	delete(req.Header, "Connection")
	c.Gui.InterceptorEditor.SetPlainText(util.RequestToString(req) + "\n")

	for {
		parse_error := false
		select {
		// pressed forward
		case <-c.forward_chan:
			if !c.interceptor_status {
				_req = req
				_resp = nil
				break
			}
			var r *http.Request
			var err error

			reader := strings.NewReader(util.NormalizeRequest(c.Gui.InterceptorEditor.ToPlainText()))
			buf := bufio.NewReader(reader)

			r, err = http.ReadRequest(buf)
			if err != nil && err == io.ErrUnexpectedEOF {
				reader := strings.NewReader(util.NormalizeRequest(c.Gui.InterceptorEditor.ToPlainText()) + "\n\n")
				buf := bufio.NewReader(reader)
				// this is so ugly
				r, err = http.ReadRequest(buf)
				if err != nil {
					c.Sess.Err(c.Module.Name(), fmt.Sprintf("Forward Req: %s", err.Error()))
					parse_error = true
				}
			}
			if err == nil {
				r.URL.Scheme = req.URL.Scheme
				r.URL.Host = req.URL.Host
				r.RequestURI = ""
				if util.RequestsEquals(req, r) {
					_req = nil
					_resp = nil
				} else {
					_req = r
					_resp = nil
				}
			}
		// pressed drop
		case <-c.drop_chan:
			_req = req
			_resp = goproxy.NewResponse(req,
				goproxy.ContentTypeText, http.StatusForbidden, "Request droppped")
		}
		if !parse_error {
			break
		}
	}
	// decrease the requests in queue
	c.requests_queue = c.requests_queue - 1
	// rest the editor
	c.Gui.InterceptorEditor.SetPlainText("")
	mutex.Unlock()

	return _req, _resp
}

func (c *CoreproxyController) interceptorResponseActions(req *http.Request, resp *http.Response) *http.Response {

	var _resp *http.Response

	// increase the requests in queue
	c.responses_queue = c.responses_queue + 1
	mutex.Lock()
	// if response is bigger than 100mb, show message that is too big
	if resp.ContentLength >= 1e+8 {
		c.Gui.InterceptorEditor.SetPlainText("Response too big")
	} else {
		c.Gui.InterceptorEditor.SetPlainText(util.ResponseToString(resp))
	}
	for {
		parse_error := false
		select {
		case <-c.forward_chan:
			if !c.interceptor_status {
				_resp = resp
				break
			}
			// if response is bigger than 100mb, ignore the content of the QPlainTextEditor
			if resp.ContentLength >= 1e+8 {
				_resp = resp
			}
			// FIXME: images cannot be tampered. I show them in case the user wants to drop, but cannot be tampered
			// because the QPlaintTextEditor only supports utf-8 character encoding
			if _, ok := resp.Header["Content-Type"]; ok && strings.HasPrefix(resp.Header["Content-Type"][0], "image") {
				_resp = resp
			}

			var tmp *http.Response
			var err error
			// pressed forward
			// remove "Content-Length" so that the ReadResponse will compute the right ContentLength

			var re = regexp.MustCompile(`(Content-Length: *\d+)\n?`)
			s := re.ReplaceAllString(c.Gui.InterceptorEditor.ToPlainText(), "")

			reader := strings.NewReader(s)
			buf := bufio.NewReader(reader)

			tmp, err = http.ReadResponse(buf, nil)
			// so bad, fix me
			_resp = tmp

			if err != nil && err == io.ErrUnexpectedEOF {
				reader := strings.NewReader(s + "\n\n")
				buf := bufio.NewReader(reader)
				// this is so ugly
				tmp, err = http.ReadResponse(buf, nil)
				_resp = tmp
				if err != nil {
					c.Sess.Err(c.Module.Name(), fmt.Sprintf("Forward Req: %s", err.Error()))
					parse_error = true
				}
			}

			if err == nil {
				if util.ResponsesEquals(resp, _resp) {
					// response not edited
					_resp = nil
				}
			}
		case <-c.drop_chan:
			// pressed drop
			resp.Body = ioutil.NopCloser(bytes.NewReader([]byte("Response dropped by user")))
			_resp = resp
		}
		if !parse_error {
			break
		}
	}

	// decrease the requests in queue
	c.responses_queue = c.responses_queue - 1
	// rest the editor
	c.Gui.InterceptorEditor.SetPlainText("")
	mutex.Unlock()

	return _resp
}
