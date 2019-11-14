package core

type Config struct {
	Address       string
	Port          int
	Interceptor   bool
	ReqIntercept  bool
	RespIntercept bool
}
