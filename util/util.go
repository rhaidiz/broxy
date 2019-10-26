package util

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strings"
)

func RequestsEquals(r1 *http.Request, r2 *http.Request) bool {
	if r1 == nil || r2 == nil {
		return false
	}
	if r1.Method != r2.Method || r1.URL.Path != r2.URL.Path || r1.Host != r2.Host {
		return false
	}
	if !reflect.DeepEqual(r1.Header, r2.Header) {
		return false
	}

	return bodyEquals(&r1.Body, &r2.Body)

}

func ResponsesEquals(r1 *http.Response, r2 *http.Response) bool {
	if r1 == nil || r2 == nil {
		return false
	}
	if r1.Status != r2.Status || r1.StatusCode != r2.StatusCode || r1.Proto != r2.Proto {
		return false
	}
	if !reflect.DeepEqual(r1.Header, r2.Header) {
		return false
	}

	return bodyEquals(&r1.Body, &r2.Body)
}

func bodyEquals(r1 *io.ReadCloser, r2 *io.ReadCloser) bool {

	var bodyBytes1 []byte
	var bodyBytes2 []byte

	bodyBytes1, _ = ioutil.ReadAll(*r1)
	*r1 = ioutil.NopCloser(bytes.NewBuffer(bodyBytes1))
	body1 := string(bodyBytes1)

	bodyBytes2, _ = ioutil.ReadAll(*r2)
	*r2 = ioutil.NopCloser(bytes.NewBuffer(bodyBytes2))
	body2 := string(bodyBytes2)

	return body1 == body2
}

func NormalizeRequest(raw_req string) string {
	a := regexp.MustCompile(`\n\n`)
	s := a.Split(raw_req, 2)
	if len(s) == 2 {
		c_l := len(s[1])
		if c_l == 0 {
			return raw_req
		}
		h := strings.Split(s[0], "\n")
		new_header := ""
		for _, v := range h {
			if !strings.HasPrefix(v, "Content-Length") {
				new_header = new_header + v + "\n"
			}
		}
		new_req_raw := fmt.Sprintf("%s%s%d\n\n%s", new_header, "Content-Length: ", c_l, s[1])
		return new_req_raw
	}
	return raw_req
}

func RequestToString(r *http.Request) string {
	if r == nil {
		return ""
	}
	ret := fmt.Sprintf("%s %s %s\nHost: %s\n", r.Method, r.URL.Path, r.Proto, r.Host)
	for k, v := range r.Header {
		values := ""
		for _, s := range v {
			values = values + s
		}
		ret = ret + fmt.Sprintf("%s: %s\n", k, values)
	}
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	if len(bodyBytes) > 0 {
		ret = ret + fmt.Sprintf("\n%s", string(bodyBytes))
	}
	return ret
}

func ResponseToString(r *http.Response, body_bytes bool) string {
	if r == nil {
		return ""
	}
	ret := fmt.Sprintf("%s %s\n", r.Proto, r.Status)
	for k, v := range r.Header {
		values := ""
		for _, s := range v {
			values = values + s
		}
		ret = ret + fmt.Sprintf("%s: %s\n", k, values)
	}
	//ret = ret + fmt.Sprintf("Content-Length: %v\n", r.ContentLength)
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	if len(bodyBytes) > 0 && !body_bytes {
		ret = ret + fmt.Sprintf("\n%s", string(bodyBytes))
	} else {
		ret = ret + fmt.Sprintf("\n%x", bodyBytes)
	}
	return ret
}
