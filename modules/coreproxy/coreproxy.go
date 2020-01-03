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

// Create a new proxy
func NewCoreProxy(s *core.Session) *Coreproxy {
	// this is my struct that I use to represent the proxy
	setCa(caCert, caKey)
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
	p.Proxyh.OnRequest().HandleConnect(goproxy.AlwaysMitm)

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
}

func (p *Coreproxy) onReqDef(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	r1, rsp := p.OnReq(r, ctx)
	ctx.RoundTripper = goproxy.RoundTripperFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (resp *http.Response, err error) {
		ctx.UserData, resp, err = p.tr.DetailedRoundTrip(req)
		return
	})
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
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	return nil
}
