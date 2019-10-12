package util

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

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
	if len(bodyBytes) > 0 {
		ret = ret + fmt.Sprintf("\n%s", string(bodyBytes))
	}
	return ""
}

func ResponseToString(r *http.Response) string {
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
	ret = ret + fmt.Sprintf("Content-Length: %v\n", r.ContentLength)
	print(r.ContentLength)
	var bodyBytes []byte
	bodyBytes, _ = ioutil.ReadAll(r.Body)

	if len(bodyBytes) > 0 {
		ret = ret + fmt.Sprintf("\n%s", string(bodyBytes))
	}
	return ret
}
