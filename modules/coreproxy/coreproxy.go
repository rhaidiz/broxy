package coreproxy

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/elazarl/goproxy"
	"github.com/elazarl/goproxy/transport"
	"github.com/rhaidiz/broxy/core"
)

type Coreproxy struct {
	core.Module

	Sess *core.Session

	Address string
	Port    int
	Proxyh  *goproxy.ProxyHttpServer
	Req     int
	Resp    int
	Srv     *http.Server
	OnReq   func(*http.Request, *goproxy.ProxyCtx)
	OnResp  func(*http.Response, *goproxy.ProxyCtx)
	status  bool
	tr      *transport.Transport
	//History			map[int64]*model.HItem
	//History2 []model.HItem
}

//var mutex = &sync.Mutex{}

// Create a new proxy
func NewCoreProxy(s *core.Session) *Coreproxy {
	// this is my struct that I use to represent the proxy
	p := &Coreproxy{
		Address: s.Config.Address,
		Port:    s.Config.Port,
		Proxyh:  goproxy.NewProxyHttpServer(),
		Req:     0,
		Resp:    0,
		OnReq:   func(*http.Request, *goproxy.ProxyCtx) {},
		OnResp:  func(*http.Response, *goproxy.ProxyCtx) {},
		Sess:    s,
		status:  false,
		tr:      &transport.Transport{Proxy: transport.ProxyFromEnvironment},

		//History: make(map[int64]*model.HItem),
		//History2: make([]model.HItem, 0),
	}

	// this is the golang HTTP server with its handler
	p.Srv = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", p.Address, p.Port),
		Handler: p.Proxyh,
	}

	// set the default behavior onReq\Resp
	p.Proxyh.OnRequest().DoFunc(p.onReqDef)
	p.Proxyh.OnResponse().DoFunc(p.onRespDef)

	return p
}

func (p *Coreproxy) ChangeIpPort(ip string, port int) error {

	ip_regexp := "^((?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?).){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?))"
	port_regexp := "(6553[0-5]|655[0-2][0-9]|65[0-4][0-9][0-9]|6[0-4][0-9][0-9][0-9]|[1-5]?[0-9]?[0-9]?[0-9]?[0-9])?$"

	r_ip := regexp.MustCompile(ip_regexp)
	r_port := regexp.MustCompile(port_regexp)

	if s := r_ip.FindStringSubmatch(ip); s == nil {
		return fmt.Errorf("Not a valid ip %s", ip)
	}

	if s := r_port.FindStringSubmatch(string(port)); s == nil {
		return fmt.Errorf("Not a valid port %s", port)
	}

	p.Address = ip
	p.Port = port

	p.Srv = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", p.Address, p.Port),
		Handler: p.Proxyh,
	}

	return nil
}

func (p *Coreproxy) Name() string {
	return "Proxy"
}

func (p *Coreproxy) Description() string {
	return "The main core proxy module, the one that logs and sees everything"
}

func (p *Coreproxy) Status() bool {
	return p.status
}

// Start the proxy
func (p *Coreproxy) Start() error {
	return p.Srv.ListenAndServe()
}

// Stop the proxy
func (p *Coreproxy) Stop() error {

	return p.Srv.Shutdown(context.Background())
	//if e := p.Srv.Shutdown(context.Background()); e != nil {
	//		// Error from closing listeners, or context timeout:
	//		fmt.Printf("HTTP server Shutdown: %v", e)
	//}
	//fmt.Printf("Stopping %s:%d\n", p.Address, p.Port)
}

func (p *Coreproxy) onReqDef(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	// count the requests
	ctx.RoundTripper = goproxy.RoundTripperFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (resp *http.Response, err error) {
		ctx.UserData, resp, err = p.tr.DetailedRoundTrip(req)
		return
	})
	p.Req = p.Req + 1
	// save the request in the history
	//mutex.Lock()
	//defer mutex.Unlock()
	//p.History[ctx.Session] = &HItem{Req: r}
	//p.History2 = append(p.History2, model.HItem{Method: r.Method})
	//fmt.Println("Resp: ", p.Req)
	p.OnReq(r, ctx)

	return r, nil
}

// Run when a response is received
func (p *Coreproxy) onRespDef(r *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	// count the responses
	p.Resp = p.Resp + 1
	// save the response in the history
	//mutex.Lock()
	//defer mutex.Unlock()
	//p.History[ctx.Session].Resp = r
	//fmt.Println("Req: ", p.Resp)
	p.OnResp(r, ctx)

	return r
}
