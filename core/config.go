package core

// Config represents the global configuration
type Config struct {
	Address       string
	Port          int
	Interceptor   bool
	ReqIntercept  bool
	RespIntercept bool
	CACertificate []byte
	CAPrivateKey  []byte
}
