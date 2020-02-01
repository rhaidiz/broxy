package coreproxy

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"github.com/elazarl/goproxy"
	"github.com/rhaidiz/broxy/util"
	"io"
	"net/http"
	"regexp"
	"strings"
)

func (c *CoreproxyController) interceptorToggle(b bool) {
	if !c.Sess.Config.Interceptor {
		c.Sess.Config.Interceptor = true
	} else {
		c.Sess.Config.Interceptor = false
		if c.requests_queue > 0 || c.responses_queue > 0 {
			tmp := c.requests_queue + c.responses_queue
			for i := 0; i < tmp; i++ {
				c.forward_chan <- true
			}
		}
	}
	c.Sess.Debug(c.Module.Name(), fmt.Sprintf("Interceptor is: %v", c.Sess.Config.Interceptor))
}

func (c *CoreproxyController) forward(b bool) {
	go func() {
		// activate only if there's something waiting
		if c.requests_queue > 0 || c.responses_queue > 0 {
			c.forward_chan <- true
		}
	}()
}

func (c *CoreproxyController) drop(b bool) {
	go func() {
		// activate only if there's something waiting
		if c.requests_queue > 0 || c.responses_queue > 0 {
			c.drop_chan <- true
		}
	}()
}

func (c *CoreproxyController) interceptorResponseActions(req *http.Request, resp *http.Response) *http.Response {

	var _resp *http.Response
	body_hex := false
	// increase the requests in queue
	c.responses_queue = c.responses_queue + 1
	mutex.Lock()
	// if response is bigger than 100mb, show message that is too big
	// if the response has come sort of encoding, show the body as hex
	// and confert it back to string after the editing
	if resp.ContentLength >= 1e+8 {
		c.Gui.InterceptorTextEdit.SetPlainText("Response too big")
	} else {
		_, content_type_ok := resp.Header["Content-Type"]
		_, content_encoding_ok := resp.Header["Content-Encoding"]
		if (content_type_ok && strings.HasPrefix(resp.Header["Content-Type"][0], "image")) || content_encoding_ok {
			c.Gui.InterceptorTextEdit.SetPlainText(util.ResponseToString(resp, true))
			body_hex = true
		} else {
			c.Gui.InterceptorTextEdit.SetPlainText(util.ResponseToString(resp, false))
		}
	}
	for {
		parse_error := false
		select {
		case <-c.forward_chan:
			if !c.Sess.Config.Interceptor {
				_resp = resp
				break
			}
			// if response is bigger than 100mb, ignore the content of the QPlainTextEditor
			if resp.ContentLength >= 1e+8 {
				_resp = resp
			}

			var tmp *http.Response
			var err error
			// pressed forward
			// remove "Content-Length" so that the ReadResponse will compute the right ContentLength
			var re = regexp.MustCompile(`(Content-Length: *\d+)\n?`)
			s := re.ReplaceAllString(c.Gui.InterceptorTextEdit.ToPlainText(), "")

			if body_hex {
				a := regexp.MustCompile(`\n\n`)
				s1 := a.Split(s, 2)
				if len(s1) == 2 {
					br, err := hex.DecodeString(s1[1])
					if err != nil {
						c.Sess.Err(c.Module.Name(), fmt.Sprintf("Forward Resp: %s", err.Error()))
						parse_error = true
					} else {
						body_hex = false
						s = fmt.Sprintf("%s\n%s", s1[0], string(br))
					}
				}
			} else {

				reader := strings.NewReader(s)
				buf := bufio.NewReader(reader)

				tmp, err = http.ReadResponse(buf, nil)
				// so bad, fix me
				_resp = tmp

				if err != nil && err == io.ErrUnexpectedEOF {
					reader := strings.NewReader(s + "\n\n")
					buf := bufio.NewReader(reader)
					// this is so ugly
					tmp, err = http.ReadResponse(buf, nil)
					_resp = tmp
					if err != nil {
						c.Sess.Err(c.Module.Name(), fmt.Sprintf("Forward Resp: %s", err.Error()))
						parse_error = true
					}
				}

				if err == nil {
					if util.ResponsesEquals(resp, _resp) {
						// response not edited
						_resp = nil
					}
				}
			}
		case <-c.drop_chan:
			// pressed drop
			_resp = goproxy.NewResponse(req,
				goproxy.ContentTypeText, http.StatusForbidden, "Request droppped")
		}
		if !parse_error {
			break
		}
	}

	// decrease the requests in queue
	c.responses_queue = c.responses_queue - 1
	// rest the editor
	c.Gui.InterceptorTextEdit.SetPlainText("")
	mutex.Unlock()

	return _resp
}

func cloneHeaders(src http.Header) http.Header {
	dst := http.Header{}
	for k, vs := range src {
		for _, v := range vs {
			dst.Add(k, v)
		}
	}
	return dst
}

func (c *CoreproxyController) interceptorRequestActions(req *http.Request, resp *http.Response, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {

	// the request to return
	var _req *http.Request
	var _resp *http.Response

	c.requests_queue = c.requests_queue + 1
	mutex.Lock()
	delete(req.Header, "Connection")
	c.Gui.InterceptorTextEdit.SetPlainText(util.RequestToString(req) + "\n")

	for {
		parse_error := false
		select {
		// pressed forward
		case <-c.forward_chan:
			if !c.Sess.Config.Interceptor {
				_req = req
				_resp = nil
				break
			}
			var r *http.Request
			var err error

			reader := strings.NewReader(util.NormalizeRequest(c.Gui.InterceptorTextEdit.ToPlainText()))
			buf := bufio.NewReader(reader)

			r, err = http.ReadRequest(buf)
			if err != nil && err == io.ErrUnexpectedEOF {
				reader := strings.NewReader(util.NormalizeRequest(c.Gui.InterceptorTextEdit.ToPlainText()) + "\n\n")
				buf := bufio.NewReader(reader)
				// this is so ugly
				r, err = http.ReadRequest(buf)
				if err != nil {
					c.Sess.Err(c.Module.Name(), fmt.Sprintf("Forward Req: %s", err.Error()))
					parse_error = true
				}
			}
			if err == nil {
				r.URL.Scheme = req.URL.Scheme
				r.URL.Host = req.URL.Host
				r.RequestURI = ""
				if util.RequestsEquals(req, r) {
					_req = nil
					_resp = nil
				} else {
					_req = r
					_resp = nil
				}
			}
		// pressed drop
		case <-c.drop_chan:
			c.dropped[ctx.Session] = true
			_req = req
			_resp = goproxy.NewResponse(req,
				goproxy.ContentTypeText, http.StatusForbidden, "Request droppped")
		}
		if !parse_error {
			break
		}
	}
	// decrease the requests in queue
	c.requests_queue = c.requests_queue - 1
	// rest the editor
	c.Gui.InterceptorTextEdit.SetPlainText("")
	mutex.Unlock()

	return _req, _resp
}
