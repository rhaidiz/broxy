package coreproxy

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"regexp"

	"github.com/elazarl/goproxy"
	"github.com/elazarl/goproxy/transport"
	"github.com/rhaidiz/broxy/core"
)

// Coreproxy represents the intercept proxy
type Coreproxy struct {
	core.Module

	Sess *core.Session

	Address string
	Port    int
	Proxyh  *goproxy.ProxyHttpServer
	Req     int
	Resp    int
	Srv     *http.Server
	OnReq   func(*http.Request, *goproxy.ProxyCtx) (*http.Request, *http.Response)
	OnResp  func(*http.Response, *goproxy.ProxyCtx) *http.Response
	status  bool
	tr      *transport.Transport
}

// NewCoreProxy creates a new intercept proxy
func NewCoreProxy(s *core.Session) *Coreproxy {
	setCa(s.Config.CACertificate, s.Config.CAPrivateKey)
	p := &Coreproxy{
		Address: s.Config.Address,
		Port:    s.Config.Port,
		Proxyh:  goproxy.NewProxyHttpServer(),
		Req:     0,
		Resp:    0,
		Sess:    s,
		status:  false,
		tr:      &transport.Transport{Proxy: transport.ProxyFromEnvironment, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}

	// this is the golang HTTP server with its handler
	p.Srv = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", p.Address, p.Port),
		Handler: p.Proxyh,
	}

	// enable always HTTPS mitm
	//p.Proxyh.OnRequest().HandleConnect(goproxy.AlwaysMitm)

	// set the default behavior onReq\Resp
	p.Proxyh.OnRequest().DoFunc(p.onReqDef)
	p.Proxyh.OnResponse().DoFunc(p.onRespDef)

	return p
}

// ChangeIPPort is used to change the ip and port of the current intercept proxy
func (p *Coreproxy) ChangeIPPort(ip string, port int) error {

	ipReg := "^((?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?).){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?))"
	portReg := "(6553[0-5]|655[0-2][0-9]|65[0-4][0-9][0-9]|6[0-4][0-9][0-9][0-9]|[1-5]?[0-9]?[0-9]?[0-9]?[0-9])?$"

	rIP := regexp.MustCompile(ipReg)
	rPort := regexp.MustCompile(portReg)

	if s := rIP.FindStringSubmatch(ip); s == nil {
		return fmt.Errorf("Not a valid ip %s", ip)
	}

	if s := rPort.FindStringSubmatch(string(port)); s == nil {
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

// Name returns the name of the current module
func (p *Coreproxy) Name() string {
	return "Proxy"
}

// Description returns the description of the current module
func (p *Coreproxy) Description() string {
	return "The main core proxy module, the one that logs and sees everything"
}

// Status returns the status of the current module if any
func (p *Coreproxy) Status() bool {
	return p.status
}

// Start bind the proxy for listening
func (p *Coreproxy) Start() error {
	return p.Srv.ListenAndServe()
}

// Stop stops the proxy
func (p *Coreproxy) Stop() error {

	return p.Srv.Shutdown(context.Background())
}

func (p *Coreproxy) onReqDef(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	r1, rsp := p.OnReq(r, ctx)
	// ctx.RoundTripper = goproxy.RoundTripperFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (resp *http.Response, err error) {
	// 	ctx.UserData, resp, err = p.tr.DetailedRoundTrip(req)
	// 	return
	// })
	return r1, rsp
}

// Run when a response is received
func (p *Coreproxy) onRespDef(r *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	if r != nil {
		r = p.OnResp(r, ctx)
	}

	return r
}

func setCa(caCert, caKey []byte) error {
	goproxyCa, err := tls.X509KeyPair(caCert, caKey)
	if err != nil {
		return err
	}
	if goproxyCa.Leaf, err = x509.ParseCertificate(goproxyCa.Certificate[0]); err != nil {
		return err
	}
	goproxy.GoproxyCa = goproxyCa
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: core.TLSConfigFromCA(&goproxyCa)}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: core.TLSConfigFromCA(&goproxyCa)}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: core.TLSConfigFromCA(&goproxyCa)}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: core.TLSConfigFromCA(&goproxyCa)}
	return nil
}
