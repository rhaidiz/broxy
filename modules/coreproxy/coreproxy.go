package coreproxy

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"regexp"

	"github.com/elazarl/goproxy"
	_ "github.com/elazarl/goproxy/transport"
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
	//tr      *transport.Transport
	//History			map[int64]*model.HItem
	//History2 []model.HItem
}

//var mutex = &sync.Mutex{}

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
		//OnReq:   func(*http.Request, *goproxy.ProxyCtx) *http.Request {},
		//OnResp: func(*http.Response, *goproxy.ProxyCtx) {},
		Sess:   s,
		status: false,
		//tr:      &transport.Transport{Proxy: transport.ProxyFromEnvironment},

		//History: make(map[int64]*model.HItem),
		//History2: make([]model.HItem, 0),
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
	//if e := p.Srv.Shutdown(context.Background()); e != nil {
	//		// Error from closing listeners, or context timeout:
	//		fmt.Printf("HTTP server Shutdown: %v", e)
	//}
	//fmt.Printf("Stopping %s:%d\n", p.Address, p.Port)
}

func (p *Coreproxy) onReqDef(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	// count the requests
	//ctx.RoundTripper = goproxy.RoundTripperFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (resp *http.Response, err error) {
	//	ctx.UserData, resp, err = p.tr.DetailedRoundTrip(req)
	//	return
	//})
	p.Req = p.Req + 1
	// save the request in the history
	//mutex.Lock()
	//defer mutex.Unlock()
	//p.History[ctx.Session] = &HItem{Req: r}
	//p.History2 = append(p.History2, model.HItem{Method: r.Method})
	//fmt.Println("Resp: ", p.Req)
	r1, rsp := p.OnReq(r, ctx)

	return r1, rsp
}

// Run when a response is received
func (p *Coreproxy) onRespDef(r *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	if r != nil {
		// count the responses
		p.Resp = p.Resp + 1
		// save the response in the history
		//mutex.Lock()
		//defer mutex.Unlock()
		//p.History[ctx.Session].Resp = r
		//fmt.Println("Req: ", p.Resp)
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

var caCert = []byte(`-----BEGIN CERTIFICATE-----
MIIDbDCCAlQCCQD0lxkKLVXsWzANBgkqhkiG9w0BAQsFADB4MQswCQYDVQQGEwJJ
VDEOMAwGA1UECAwFSXRhbHkxDjAMBgNVBAcMBU1pbGFuMRAwDgYDVQQKDAdyaGFp
ZGl6MRAwDgYDVQQDDAdyaGFpZGl6MSUwIwYJKoZIhvcNAQkBFhZyaGFpZGl6QHBy
b3Rvbm1haWwuY29tMB4XDTE5MTAwOTEyMjgzNVoXDTMwMDMzMDEyMjgzNVoweDEL
MAkGA1UEBhMCSVQxDjAMBgNVBAgMBUl0YWx5MQ4wDAYDVQQHDAVNaWxhbjEQMA4G
A1UECgwHcmhhaWRpejEQMA4GA1UEAwwHcmhhaWRpejElMCMGCSqGSIb3DQEJARYW
cmhhaWRpekBwcm90b25tYWlsLmNvbTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCC
AQoCggEBANvluPGGaHIe9s49GvmZchVWNqtjFpqrBGPX/7ud3wHk8bzxOtfw56rM
nZwNxZxr8cPQ89KGTusRvj2b6fpYTdOzWW7lytZ4nvCKcQmt+NmsabuAtOvZWcMY
pKRsiTuCC+csKf6n2+Wtg8T0wMkDkPETgGsnaBJpLUaduconJ0NsnmVv/9UiEMbe
4EcUCBprjkPxqq/1zf89nGXWxiVUVL0F5OzejYUIgl9c5ZqdYc1FvY8AviRCihdV
1aUG3PeAaObLqcYZCie9AkRD8vEx/ZgFHEFGUEaIZz777f5Sp3mABoYxyr4aWo8J
WzDEqHFVxb2O6PyKiBP/mVHkE7NUlMECAwEAATANBgkqhkiG9w0BAQsFAAOCAQEA
XotpJNnhYp4wYwwXGk1wWKlS9wxBUCtlC43bkg8AKxd+NONXVHyJ1nnoZQK4HMKn
U/SDynSD8eIOFkxuQhwEgrFjoPM8OlRAo1fLj0lcHUE1sNQ9svzhUeoBcM2RkbVy
DC4l1W5OlTS21PLqRDWEs0m/AIUTg5L8v48+ghqaxmoDnSILh4eS6FB4jG6ECf2o
U2nFncaF+K50mvl7Hh83yE+hZ/Ny2tn7qq7Q+FxTw9zgnQyVjXlvfzWLTY0szy11
Snqsn021kY8IXH6ZmmVtG81WYPuVKx9j9WNAepDjkRfVMPqxb6HIDC1sPaGTMa65
8YugZLH75NhWbfJWDIwadg==
-----END CERTIFICATE-----`)

var caKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA2+W48YZoch72zj0a+ZlyFVY2q2MWmqsEY9f/u53fAeTxvPE6
1/DnqsydnA3FnGvxw9Dz0oZO6xG+PZvp+lhN07NZbuXK1nie8IpxCa342axpu4C0
69lZwxikpGyJO4IL5ywp/qfb5a2DxPTAyQOQ8ROAaydoEmktRp25yicnQ2yeZW//
1SIQxt7gRxQIGmuOQ/Gqr/XN/z2cZdbGJVRUvQXk7N6NhQiCX1zlmp1hzUW9jwC+
JEKKF1XVpQbc94Bo5supxhkKJ70CREPy8TH9mAUcQUZQRohnPvvt/lKneYAGhjHK
vhpajwlbMMSocVXFvY7o/IqIE/+ZUeQTs1SUwQIDAQABAoIBAHK94ww8W0G5QIWL
Qwkc9XeGvg4eLUxVknva2Ll4fkZJxY4WveKx9OCd1lv4n7WoacYIwUGIDaQBZShW
s/eKnkmqGy+PvpC87gqL4sHvQpuqqJ1LYpxylLEFqduWOuGPUVC2Lc+QnWCycsCS
CgqZzsbMq0S+kkKRGSvw32JJneZCzqLgLNssQNVk+Gm6SI3s4jJsGPesjhnvoPaa
xZK14uFpltaA05GSTDaQeZJFEdnnb3f/eNPc2xMEfi0S2ZlJ6Q92WJEOepAetDlR
cRFi004bNyTb4Bphg8s4+9Cti5is199aFkGCRDWxeqEnc6aMY3Ezu9Qg3uttLVUd
uy830GUCgYEA7qS0X+9UH1R02L3aoANyADVbFt2ZpUwQGauw9WM92pH52xeHAw1S
ohus6FI3OC8xQq2CN525tGLUbFDZnNZ3YQHqFsfgevfnTs1//gbKXomitev0oFKh
VT+WYS4lkgYtPlXzhdGuk32q99T/wIocAguvCUY3PiA7yBz93ReyausCgYEA6+P8
bugMqT8qjoiz1q/YCfxsw9bAGWjlVqme2xmp256AKtxvCf1BPsToAaJU3nFi3vkw
ICLxUWAYoMBODJ3YnbOsIZOavdXZwYHv54JqwqFealC3DG0Du6fZYZdiY8pK+E6m
3fiYzP1WoVK5tU4bH8ibuIQvpcI8j7Gy0cV6/AMCgYBHl7fZNAZro72uLD7DVGVF
9LvP/0kR0uDdoqli5JPw12w6szM40i1hHqZfyBJy042WsFDpeHL2z9Nkb1jpeVm1
C4r7rJkGqwqElJf6UHUzqVzb8N6hnkhyN7JYkyyIQzwdgFGfaslRzBiXYxoa3BQM
9Q5c3OjDxY3JuhDa3DoVYwKBgDNqrWJLSD832oHZAEIicBe1IswJKjQfriWWsV6W
mHSbdtpg0/88aZVR/DQm+xLFakSp0jifBTS0momngRu06Dtvp2xmLQuF6oIIXY97
2ON1owvPbibSOEcWDgb8pWCU/oRjOHIXts6vxctCKeKAFN93raGphm0+Ck9T72NU
BTubAoGBAMEhI/Wy9wAETuXwN84AhmPdQsyCyp37YKt2ZKaqu37x9v2iL8JTbPEz
pdBzkA2Gc0Wdb6ekIzRrTsJQl+c/0m9byFHsRsxXW2HnezfOFX1H4qAmF6KWP0ub
M8aIn6Rab4sNPSrvKGrU6rFpv/6M33eegzldVnV9ku6uPJI1fFTC
-----END RSA PRIVATE KEY-----`)
