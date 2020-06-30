package coreproxy

// Settings represents the settings for the core intercept proxy
// TODO: fix it because you also have core.Config
type Settings struct {
	IP            string
	Port          int
	Interceptor   bool
	ReqIntercept  bool
	RespIntercept bool
}

var Stg = &Settings{
	IP:            "127.0.0.1",
	Port:          8080,
	Interceptor:   false,
	ReqIntercept:  true,
	RespIntercept: false,
}
