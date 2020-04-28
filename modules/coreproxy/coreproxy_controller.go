package coreproxy

import (
	"bytes"
	"log"
	//TODO: I don't like to have the encoding/json here, create a generic 
	// encofing interface instead
	// "encoding/json"
	"github.com/rhaidiz/broxy/core/project/decoder"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/elazarl/goproxy"
	"github.com/rhaidiz/broxy/core"
	_ "github.com/rhaidiz/broxy/core/project"
	"github.com/rhaidiz/broxy/modules/coreproxy/model"
	"github.com/atotto/clipboard"
)

// Controller represents the controller for the main intercetp proxy
type Controller struct {
	core.ControllerModule
	Module *Coreproxy
	Gui    *Gui
	Sess   *core.Session
	filter *model.Filter

	isRunning   bool
	model       *model.SortFilterModel
	id          int
	ignoreHTTPS bool

	forwardChan chan bool
	dropChan    chan bool
	// will maintain the number of requests in queue
	requestsQueue  int
	responsesQueue int

	dropped map[int64]bool

	// encoders
	requestEnc 				decoder.Encoder
	requestEditedEnc		decoder.Encoder
	responseEnc				decoder.Encoder
	responseEditedEnc		decoder.Encoder

	// decoders
	requestDec				decoder.Decoder
	requestEditedDec		decoder.Decoder
	responseDec				decoder.Decoder
	responseEditedDec		decoder.Decoder
}

var mutex = &sync.Mutex{}
var count int64

// NewController creates a new controller for the core intercetp proxy
func NewController(proxy *Coreproxy, proxygui *Gui, s *core.Session) *Controller {
	c := &Controller{
		Module:         proxy,
		Gui:            proxygui,
		Sess:           s,
		isRunning:      false,
		id:             0,
		ignoreHTTPS:    false,
		forwardChan:    make(chan bool),
		dropChan:       make(chan bool),
		requestsQueue:  0,
		responsesQueue: 0,
		dropped:        make(map[int64]bool),
		filter:         &model.Filter{},		
	}

	// get the encoders
	c.requestEnc, _ 		= s.PersistentProject.FileEncoder2("requests")
	c.requestEditedEnc, _ 	= s.PersistentProject.FileEncoder2("requests_edited")
	c.responseEnc, _ 		= s.PersistentProject.FileEncoder2("response")
	c.responseEditedEnc, _ 	= s.PersistentProject.FileEncoder2("response_edited")

	// get the decoders, which I only need to read the history the first time
	c.requestDec, _ 		= s.PersistentProject.FileDecoder2("requests")
	c.requestEditedDec, _ 	= s.PersistentProject.FileDecoder2("requests_edited")
	c.responseDec, _ 		= s.PersistentProject.FileDecoder2("response")
	c.responseEditedDec, _ 	= s.PersistentProject.FileDecoder2("response_edited")

	// load the history
	count = 1
	for {
		// load the requests
		var req model.Request
		var row model.Row
		if err := c.requestDec.Decode(&req); err != nil {
			break
		}
		row.ID = req.ID
		if req.ID > count{
			count = req.ID
		}
		row.Req = &req
		model.History = append(model.History, &row)
		model.HashMapHistory[row.ID] = len(model.History) - 1
	}
	for{
		// load edited requests
		var req model.Request
		if err := c.requestEditedDec.Decode(&req); err != nil {
			break
		}
		row := model.History[model.HashMapHistory[req.ID]]
		(*row).EditedReq = &req
	}
	for{
		// load the responses
		var resp model.Response
		if err := c.responseDec.Decode(&resp); err != nil {
			break
		}
		row := model.History[model.HashMapHistory[resp.ID]]
		(*row).Resp = &resp
	}
	for{
		// load edited responses
		var resp model.Response
		if err := c.responseEditedDec.Decode(&resp); err != nil {
			break
		}
		row := model.History[model.HashMapHistory[resp.ID]]
		(*row).EditedResp = &resp
	}

	//count = int64(len(model.History))
	log.Printf("count: %d\n", count)
	// load settings and save settings
	c.Sess.PersistentProject.LoadSettings("coreproxy", Stg)
	c.Sess.PersistentProject.SaveSettings("coreproxy", Stg)



	c.model = model.NewSortFilterModel(nil)
	//c.model.Custom.Refresh()

	c.Module.OnReq = c.onReq
	c.Module.OnResp = c.onResp
	c.Module.Proxyh.OnRequest().HandleConnect(goproxy.FuncHttpsHandler(c.broxyConnectHandle))
	c.Gui.SetTableModel(c.model)
	
	// UI events
	c.Gui.RowClicked = c.selectRow
	c.Gui.Forward = c.forward
	c.Gui.Drop = c.drop
	c.Gui.ApplyFilters = c.applyFilter
	c.Gui.ResetFilters = c.resetFilter
	c.Gui.SaveCAClicked = c.downloadCAClicked
	c.Gui.RightItemClicked = c.rightItemClicked

	// UI settings events
	c.Gui.StartProxy = c.startProxy
	c.Gui.Toggle = c.interceptorToggle
	c.Gui.CheckReqInterception = c.checkReqInterception
	c.Gui.CheckRespInterception = c.checkRespInterception
	c.Gui.CheckIgnoreHTTPS = c.ignoreHTTPSToggle

	// UI init
	c.Gui.ControllerInit = c.initUIContent
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

}

func (c *Controller) initUIContent() {
	c.setDefaultFilter()
	c.Gui.ListenerLineEdit.SetText(fmt.Sprintf("%s:%d", Stg.IP, Stg.Port))
	if Stg.Interceptor {
		c.Gui.InterceptorToggleButton.SetChecked(true)
	}
	if Stg.ReqIntercept {
		c.Gui.ReqInterceptCheckBox.SetChecked(true)
	}
	if Stg.RespIntercept {
		c.Gui.RespInterceptCheckBox.SetChecked(true)
	}
}

func (c *Controller) rightItemClicked(s string, r int) {
	req, _, _, _ := c.model.Custom.GetReqResp(c.model.GetRowId(r))
	if s == CopyURLLabel {
		clipboard.WriteAll(fmt.Sprintf("%s://%s%s", req.URL.Scheme, req.Host, req.URL.Path))
	} else if s == CopyBaseURLLabel {
		clipboard.WriteAll(fmt.Sprintf("%s://%s", req.URL.Scheme, req.Host))
	} else if s == RepeatLabel {
		// FIXME: I **really** don't like this
		c.Sess.Exec("Repeater", "send-to", req)
	} else if s == ClearHistoryLabel {
		c.model.Custom.ClearHistory()
		c.id = 0
	}
}

func (c *Controller) downloadCAClicked(b bool) {
	c.Gui.FileSaveAs(string(c.Sess.Settings.CACertificate))
}

func (c *Controller) checkReqInterception(b bool) {
	Stg.ReqIntercept = c.Gui.ReqInterceptCheckBox.IsChecked()
	c.saveSettings()
}

func (c *Controller) checkRespInterception(b bool) {
	Stg.RespIntercept = c.Gui.RespInterceptCheckBox.IsChecked()
	c.saveSettings()
}

// Defaut history filters
func (c *Controller) setDefaultFilter() {
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

func (c *Controller) applyFilter(b bool) {
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
	c.filter.ShowExt = make(map[string]bool)
	if c.Gui.ShowOnlyCheckBox.IsChecked() {
		for _, e := range strings.Split(strings.Replace(c.Gui.ShowExtensionLineEdit.DisplayText(), " ", "", -1), ",") {
			c.filter.ShowExt[e] = true
		}
	}
	c.filter.HideExt = make(map[string]bool)
	if c.Gui.HideOnlyCheckBox.IsChecked() {
		for _, e := range strings.Split(strings.Replace(c.Gui.HideExtensionLineEdit.DisplayText(), " ", "", -1), ",") {
			c.filter.HideExt[e] = true
		}
	}
	c.model.SetFilter(c.filter)
}

func (c *Controller) resetFilter(b bool) {
	c.setDefaultFilter()
}

func (c *Controller) selectRow(r int) {
	c.Gui.HideAllTabs()
	req, editedReq, resp, editedResp := c.model.Custom.GetReqResp(c.model.GetRowId(r))
	if req != nil {
		c.Gui.ShowReqTab(req.ToString())
	}
	if editedReq != nil {
		c.Gui.ShowEditedReqTab(editedReq.ToString())
	}
	if resp != nil {
		c.Gui.ShowRespTab(resp.ToString())
	}
	if editedResp != nil {
		c.Gui.ShowEditedRespTab(editedResp.ToString())
	}
}

func (c *Controller) startProxy(b bool) {

	if !c.isRunning {
		// Start and stop the proxy
		IPPortReg := "^((?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?).){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)):(6553[0-5]|655[0-2][0-9]|65[0-4][0-9][0-9]|6[0-4][0-9][0-9][0-9]|[1-5]?[0-9]?[0-9]?[0-9]?[0-9])?$"

		r := regexp.MustCompile(IPPortReg)

		if s := r.FindStringSubmatch(c.Gui.ListenerLineEdit.DisplayText()); s != nil {
			p, _ := strconv.Atoi(s[2])
			if e := c.Module.ChangeIPPort(s[1], p); e == nil {
				// if I can change ip and port, change it also in the config struct
				Stg.IP = s[1]
				Stg.Port = p
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
		if Stg.Interceptor {
			c.interceptorToggle(false)
		}
		c.Module.Stop()
		c.isRunning = false
		c.Sess.Info(c.Module.Name(), "Stopping proxy.")
		c.Gui.StartStopButton.SetText("Start")
	}
	c.saveSettings()
}

// Executed when a response arrives
func (c *Controller) onResp(r *http.Response, ctx *goproxy.ProxyCtx) *http.Response {

	httpItem := model.NewHTTPItem(nil)

	var bodyBytes []byte
	if r != nil {
		bodyBytes, _ = ioutil.ReadAll(r.Body)
		// Restore the io.ReadCloser to its original state
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	}
	httpItem.Resp = &model.Response{
		ID:			   count + ctx.Session,
		Status:        r.Status,
		StatusCode:    r.StatusCode,
		Body:          bodyBytes,
		Proto:         r.Proto,
		ContentLength: int64(len(bodyBytes)),
		Headers:       cloneHeaders(r.Header),
	}

	go func(){ c.responseEnc.Encode(httpItem.Resp) }()
	c.model.Custom.AddResponse(httpItem.Resp, count+ctx.Session)
	// activate interceptor
	_, dropped := c.dropped[ctx.Session]
	if Stg.Interceptor && Stg.RespIntercept && !dropped {
		// if the response is nil, it means the interceptor did not change the response

		r.ContentLength = int64(len(bodyBytes))
		editedResp := c.interceptorResponseActions(ctx.Req, r)
		// the response was edited
		if editedResp != nil {
			var editedBodyBytes []byte
			editedBodyBytes, _ = ioutil.ReadAll(editedResp.Body)
			editedResp.Body = ioutil.NopCloser(bytes.NewBuffer(editedBodyBytes))
			httpItem.EditedResp = &model.Response{
				ID:			   count + ctx.Session,
				Status:        editedResp.Status,
				StatusCode:    editedResp.StatusCode,
				Proto:         editedResp.Proto,
				Body:          editedBodyBytes,
				ContentLength: int64(len(editedBodyBytes)),
				Headers:       cloneHeaders(editedResp.Header),
			}
			r = editedResp
			go func(){ c.responseEditedEnc.Encode(httpItem.EditedResp) }()
			c.model.Custom.AddEditedResponse(httpItem.EditedResp, count+ctx.Session)
		}
		
	}
	return r
}

// Executed when a request arrives
func (c *Controller) onReq(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	var resp *http.Response
	var bodyBytes []byte
	if r != nil {
		bodyBytes, _ = ioutil.ReadAll(r.Body)
		// Restore the io.ReadCloser to its original state
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	}
	httpItem := model.NewHTTPItem(nil)
	c.id = c.id + 1
	httpItem.ID = c.id

	re := regexp.MustCompile(`\.(\w*)($|\?|\#)`)
	matches := re.FindStringSubmatch(r.URL.Path)
	ext := ""
	if len(matches) >= 1 {
		ext = matches[1]
	}
	params := false
	if len(r.URL.RawQuery) > 0 || len(bodyBytes) > 0 {
		params = true
	}
	// this is the original request, save it for the history
	httpItem.Req = &model.Request{
		ID:			   count + ctx.Session,
		URL:           r.URL,
		Method:        r.Method,
		Body:          bodyBytes,
		Host:          r.Host,
		ContentLength: r.ContentLength,
		Headers:       cloneHeaders(r.Header),
		Proto:         r.Proto,
		Extension:     ext,
		Params:        params,
	}

	go func(){ c.requestEnc.Encode(httpItem.Req) }()
	c.model.Custom.AddRequest(httpItem.Req, count+ctx.Session)

	// activate interceptor
	if Stg.Interceptor && Stg.ReqIntercept {

		editedReq, editedResp := c.interceptorRequestActions(r, nil, ctx)

		if editedReq != nil {
			var editedBodyBytes []byte
			editedBodyBytes, _ = ioutil.ReadAll(editedReq.Body)
			editedReq.Body = ioutil.NopCloser(bytes.NewBuffer(editedBodyBytes))

			re := regexp.MustCompile(`\.(\w*)($|\?|\#)`)
			matches := re.FindStringSubmatch(r.URL.Path)
			ext := ""
			if len(matches) >= 1 {
				ext = matches[1]
			}

			httpItem.EditedReq = &model.Request{
				ID:			   count + ctx.Session,
				URL:           editedReq.URL,
				Method:        editedReq.Method,
				Body:          editedBodyBytes,
				Host:          editedReq.Host,
				ContentLength: editedReq.ContentLength,
				Headers:       cloneHeaders(editedReq.Header),
				Proto:         editedReq.Proto,
				Extension:     ext,
			}
			r = editedReq
			resp = editedResp
			go func(){ c.requestEditedEnc.Encode(httpItem.EditedReq) }()
			c.model.Custom.AddEditedRequest(httpItem.EditedReq, count+ctx.Session)
		}

	}
	return r, resp
}

func (c *Controller) ignoreHTTPSToggle(b bool) {
	c.ignoreHTTPS = !c.ignoreHTTPS
}

func (c *Controller) broxyConnectHandle(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
	if c.ignoreHTTPS {
		return goproxy.OkConnect, host
	}
	return goproxy.MitmConnect, host
}

func (c *Controller) saveSettings(){
	c.Sess.PersistentProject.SaveSettings("coreproxy", Stg)
}