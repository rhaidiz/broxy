package core

// Config represents the global configuration
type Config struct {
	CACertificate   []byte    `xml:"CACert"`
	CAPrivateKey    []byte    `xml:"CAPvt"`
	Project					*Project
}

type History struct {
	H []*Project `xml:"ProjectsHistory"`
}

type Project struct {
	Title string
	Path  string
}
