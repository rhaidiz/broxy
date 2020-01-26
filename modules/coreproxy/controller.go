package coreproxy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/elazarl/goproxy"
	"github.com/rhaidiz/broxy/core"
	"github.com/rhaidiz/broxy/modules/coreproxy/model"
	qtcore "github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
)

type CoreproxyController struct {
	core.ControllerModule
	Module *Coreproxy
	Gui    *CoreproxyGui
	Sess   *core.Session
	filter *model.Filter

	isRunning bool
	model     *model.SortFilterModel
	id        int

	forward_chan chan bool
	drop_chan    chan bool
	// will maintain the number of requests in queue
	requests_queue  int
	responses_queue int

	dropped map[int64]bool
}

var mutex = &sync.Mutex{}

func NewCoreproxyController(proxy *Coreproxy, proxygui *CoreproxyGui, s *core.Session) *CoreproxyController {
	c := &CoreproxyController{
		Module:          proxy,
		Gui:             proxygui,
		Sess:            s,
		isRunning:       false,
		id:              0,
		forward_chan:    make(chan bool),
		drop_chan:       make(chan bool),
		requests_queue:  0,
		responses_queue: 0,
		dropped:         make(map[int64]bool),
		filter:          &model.Filter{},
	}

	c.model = model.NewSortFilterModel(nil)
	c.Module.OnReq = c.onReq
	c.Module.OnResp = c.onResp
	c.Gui.SetTableModel(c.model)
	c.Gui.StartProxy = c.startProxy
	c.Gui.RowClicked = c.selectRow
	c.Gui.Toggle = c.interceptorToggle
	c.Gui.Forward = c.forward
	c.Gui.Drop = c.drop
	c.Gui.ApplyFilters = c.applyFilter
	c.Gui.ResetFilters = c.resetFilter
	c.Gui.ControllerInit = c.initUIContent
	c.Gui.CheckReqInterception = c.checkReqInterception
	c.Gui.CheckRespInterception = c.checkRespInterception
	c.Gui.DownloadCAClicked = c.downloadCAClicked
	c.Gui.RightItemClicked = c.rightItemClicked
	return c
}

func (c *CoreproxyController) GetGui() core.GuiModule {
	return c.Gui
}

func (c *CoreproxyController) Name() string {
	return "coreproxy"
}

func (c *CoreproxyController) GetModule() core.Module {
	return c.Module
}

func (c *CoreproxyController) ExecCommand(m string, args ...interface{}) {

}

// init UI content
func (c *CoreproxyController) initUIContent() {
	c.setDefaultFilter()
	c.Gui.ListenerLineEdit.SetText(fmt.Sprintf("%s:%d", c.Sess.Config.Address, c.Sess.Config.Port))
	if c.Sess.Config.Interceptor {
		c.Gui.InterceptorToggleButton.SetChecked(true)
	}
	if c.Sess.Config.ReqIntercept {
		c.Gui.ReqInterceptCheckBox.SetChecked(true)
	}
	if c.Sess.Config.RespIntercept {
		c.Gui.RespInterceptCheckBox.SetChecked(true)
	}
}

func (c *CoreproxyController) rightItemClicked(s string, r int) {
	clipboard := c.Sess.QApp.Clipboard()
	actual_row := c.model.Index(r, 0, qtcore.NewQModelIndex()).Data(model.ID).ToInt(nil)
	req, _, _, _ := c.model.Custom.GetReqResp(actual_row - 1)
	if s == CopyURLLabel {
		clipboard.SetText(fmt.Sprintf("%s://%s%s", req.Url.Scheme, req.Host, req.Url.Path), gui.QClipboard__Clipboard)
	} else if s == CopyBaseURLLabel {
		clipboard.SetText(fmt.Sprintf("%s://%s", req.Url.Scheme, req.Host), gui.QClipboard__Clipboard)
	} else if s == RepeatLabel {
		// FIXME: I **really** don't like this
		c.Sess.Exec("repeater", "send-to", req)
	} else if s == ClearHistoryLabel {
		c.model.Custom.ClearHistory()
		c.id = 0
	}
}

func (c *CoreproxyController) downloadCAClicked(b bool) {
	c.Gui.FileSaveAs(string(caCert))
}

func (c *CoreproxyController) checkReqInterception(b bool) {
	c.Sess.Config.ReqIntercept = c.Gui.ReqInterceptCheckBox.IsChecked()
}

func (c *CoreproxyController) checkRespInterception(b bool) {
	c.Sess.Config.RespIntercept = c.Gui.RespInterceptCheckBox.IsChecked()
}

// Defaut history filters
func (c *CoreproxyController) setDefaultFilter() {
	c.Gui.TextSearchLineEdit.SetText("")
	c.Gui.S100CheckBox.SetChecked(true)
	c.Gui.S200CheckBox.SetChecked(true)
	c.Gui.S300CheckBox.SetChecked(true)
	c.Gui.S400CheckBox.SetChecked(true)
	c.Gui.S500CheckBox.SetChecked(true)
	c.Gui.ShowOnlyCheckBox.SetChecked(false)
	c.Gui.HideOnlyCheckBox.SetChecked(true)
	c.Gui.ShowExtensionLineEdit.SetText("asp, aspx, jsp, php, html, htm")
	c.Gui.HideExtensionLineEdit.SetText("png, jpg, css, woff2, ico")
	c.applyFilter(true)
}

func (c *CoreproxyController) applyFilter(b bool) {
	c.filter.Search = c.Gui.TextSearchLineEdit.DisplayText()
	var status []int
	if c.Gui.S100CheckBox.IsChecked() {
		status = append(status, 100)
	}
	if c.Gui.S200CheckBox.IsChecked() {
		status = append(status, 200)
	}
	if c.Gui.S300CheckBox.IsChecked() {
		status = append(status, 300)
	}
	if c.Gui.S400CheckBox.IsChecked() {
		status = append(status, 400)
	}
	if c.Gui.S500CheckBox.IsChecked() {
		status = append(status, 500)
	}
	// this also looks bad, creating a new status each time and replacing it ... bleah ...
	//IMP: make me pretier
	c.filter.StatusCode = status
	c.filter.Show_ext = make(map[string]bool)
	if c.Gui.ShowOnlyCheckBox.IsChecked() {
		for _, e := range strings.Split(strings.Replace(c.Gui.ShowExtensionLineEdit.DisplayText(), " ", "", -1), ",") {
			c.filter.Show_ext[e] = true
		}
	}
	c.filter.Hide_ext = make(map[string]bool)
	if c.Gui.HideOnlyCheckBox.IsChecked() {
		for _, e := range strings.Split(strings.Replace(c.Gui.HideExtensionLineEdit.DisplayText(), " ", "", -1), ",") {
			c.filter.Hide_ext[e] = true
		}
	}
	c.model.SetFilter(c.filter)
}

func (c *CoreproxyController) resetFilter(b bool) {
	c.setDefaultFilter()
}

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
					c.Gui.StartStopButton.SetText("Stop")
					c.isRunning = true
					c.Sess.Info(c.Module.Name(), "Starting proxy ...")
					if e := c.Module.Start(); e != nil && e != http.ErrServerClosed {
						c.Sess.Err(c.Module.Name(), fmt.Sprintf("Error starting the proxy %s\n", e))
						c.isRunning = false
						c.Gui.StartStopButton.SetText("Start")
					}
				}()
			} else {
				c.Sess.Err(c.Module.Name(), fmt.Sprintf("Error starting the proxy %s\n", e))
			}
		} else {
			c.Sess.Err(c.Module.Name(), "Wrong input")
		}
	} else {
		if c.Sess.Config.Interceptor {
			c.interceptorToggle(false)
		}
		c.Module.Stop()
		c.isRunning = false
		c.Sess.Info(c.Module.Name(), "Stopping proxy.")
		c.Gui.StartStopButton.SetText("Start")
	}
}

// Executed when a response arrives
func (c *CoreproxyController) onResp(r *http.Response, ctx *goproxy.ProxyCtx) *http.Response {

	http_item := model.NewHttpItem(nil)

	var bodyBytes []byte
	if r != nil {
		bodyBytes, _ = ioutil.ReadAll(r.Body)
		// Restore the io.ReadCloser to its original state
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	}
	http_item.Resp = &model.Response{
		Status:        r.Status,
		StatusCode:    r.StatusCode,
		Body:          bodyBytes,
		Proto:         r.Proto,
		ContentLength: int64(len(bodyBytes)),
		Headers:       cloneHeaders(r.Header),
	}
	// activate interceptor
	_, dropped := c.dropped[ctx.Session]
	if c.Sess.Config.Interceptor && c.Sess.Config.RespIntercept && !dropped {
		// if the response is nil, it means the interceptor did not change the response

		r.ContentLength = int64(len(bodyBytes))
		edited_resp := c.interceptorResponseActions(ctx.Req, r)
		// the response was edited
		if edited_resp != nil {
			var edited_bodyBytes []byte
			edited_bodyBytes, _ = ioutil.ReadAll(edited_resp.Body)
			edited_resp.Body = ioutil.NopCloser(bytes.NewBuffer(edited_bodyBytes))
			http_item.EditedResp = &model.Response{
				Status:        edited_resp.Status,
				StatusCode:    edited_resp.StatusCode,
				Proto:         edited_resp.Proto,
				Body:          edited_bodyBytes,
				ContentLength: int64(len(edited_bodyBytes)),
				Headers:       cloneHeaders(edited_resp.Header),
			}
			r = edited_resp
		}
	}

	// add the response to the history
	// TODO: For whatever reason, I have to use a full HttpItem insteam of a Resp
	c.model.Custom.EditItem(http_item, ctx.Session)

	return r
}

// Executed when a request arrives
func (c *CoreproxyController) onReq(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
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

	re := regexp.MustCompile(`\.(\w*)($|\?|\#)`)
	matches := re.FindStringSubmatch(r.URL.Path)
	ext := ""
	if len(matches) >= 1 {
		ext = matches[1]
	}
	// this is the original request, save it for the history
	http_item.Req = &model.Request{
		Url:           r.URL,
		QueryString:   r.URL.RawQuery,
		Method:        r.Method,
		Body:          bodyBytes,
		Host:          r.Host,
		ContentLength: r.ContentLength,
		Headers:       cloneHeaders(r.Header),
		Proto:         r.Proto,
		Extension:     ext,
	}

	if len(http_item.Req.QueryString) > 0 {
		fmt.Println("Query string: %s", http_item.Req.QueryString)
	}

	// activate interceptor
	if c.Sess.Config.Interceptor && c.Sess.Config.ReqIntercept {

		edited_req, edited_resp := c.interceptorRequestActions(r, nil, ctx)

		if edited_req != nil {
			var edited_bodyBytes []byte
			edited_bodyBytes, _ = ioutil.ReadAll(edited_req.Body)
			edited_req.Body = ioutil.NopCloser(bytes.NewBuffer(edited_bodyBytes))

			re := regexp.MustCompile(`\.(\w*)($|\?|\#)`)
			matches := re.FindStringSubmatch(r.URL.Path)
			ext := ""
			if len(matches) >= 1 {
				ext = matches[1]
			}

			http_item.EditedReq = &model.Request{
				Url:           edited_req.URL,
				Method:        edited_req.Method,
				Body:          edited_bodyBytes,
				Host:          edited_req.Host,
				ContentLength: edited_req.ContentLength,
				Headers:       cloneHeaders(edited_req.Header),
				Proto:         edited_req.Proto,
				Extension:     ext,
			}
			r = edited_req
			resp = edited_resp

		}

	}

	// add the request to the history
	c.model.Custom.AddItem(http_item, ctx.Session)

	return r, resp
}
