package repeater

import (
	"crypto/tls"
	"github.com/rhaidiz/broxy/core"
	"net/http"
)

type Repeater struct {
	core.Module

	Gui  *core.GuiModule
	Sess *core.Session

	client *http.Client
}

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

func (r *Repeater) Name() string {
	return "Repeater"
}

func (r *Repeater) Description() string {
	return "This is the magical repeater"
}

func (r *Repeater) RunRequest(req *http.Request) (*http.Response, error) {
	//TODO: I might need to use the configuration stored in session
	return r.client.Do(req)

}

// TODO: remove the following methods

func (r *Repeater) Status() bool {
	return true
}

func (r *Repeater) Start() error {
	return nil
}
func (r *Repeater) Stop() error {
	return nil
}
