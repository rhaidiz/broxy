package core

// Config represents the global configuration
type Config struct {
	CACertificate []byte `xml:"CACert"`
	CAPrivateKey  []byte `xml:"CAPvt"`
}
