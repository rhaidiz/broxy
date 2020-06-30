package repeater

import (
	"crypto/tls"
	"net/http"

	"github.com/rhaidiz/broxy/core"
)

// Repeater represents the repeater module
type Repeater struct {
	core.Module

	Gui  *core.GuiModule
	Sess *core.Session

	client *http.Client
}

// NewRepeater returns a new repeater module
func NewRepeater(s *core.Session) *Repeater {
	// disable x509 certificate check
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return &Repeater{Sess: s, client: &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	},
	}
}

// Name returns the name of the current module
func (r *Repeater) Name() string {
	return "Repeater"
}

// Description returns the description of the current module
func (r *Repeater) Description() string {
	return "This is the magical repeater"
}

// RunRequest performs the given HTTP request and return the HTTP response
func (r *Repeater) RunRequest(req *http.Request) (*http.Response, error) {
	//TODO: I might need to use the configuration stored in session
	return r.client.Do(req)

}
