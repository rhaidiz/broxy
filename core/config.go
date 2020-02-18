package core

// Config represents the configuration of the main intercept proxy
type Config struct {
	Address       string
	Port          int
	Interceptor   bool
	ReqIntercept  bool
	RespIntercept bool
}
