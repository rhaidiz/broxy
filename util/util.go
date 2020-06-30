package util

import (
	"os"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os/user"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
)

// RequestsEquals returns if two requests are equal
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

// ResponsesEquals returns if two responses are equal
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

// NormalizeRequest adds missing headers to a request
func NormalizeRequest(rawReq string) string {
	a := regexp.MustCompile(`\n\n`)
	s := a.Split(rawReq, 2)
	if len(s) == 2 {
		cL := len(s[1])
		if cL == 0 {
			return rawReq
		}
		h := strings.Split(s[0], "\n")
		newHeader := ""
		for _, v := range h {
			if !strings.HasPrefix(v, "Content-Length") {
				newHeader = newHeader + v + "\n"
			}
		}
		newReqRaw := fmt.Sprintf("%s%s%d\n\n%s", newHeader, "Content-Length: ", cL, s[1])
		return newReqRaw
	}
	return rawReq
}

// RequestToString returns a string representation of a given HTTP request
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

// ResponseToString returns a string representation of a given HTTP response
func ResponseToString(r *http.Response, responseBodyBytes bool) string {
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

	if len(bodyBytes) > 0 && !responseBodyBytes {
		ret = ret + fmt.Sprintf("\n%s", string(bodyBytes))
	} else {
		ret = ret + fmt.Sprintf("\n%x", bodyBytes)
	}
	return ret
}

// GetSettingsDir returns the path to the settings folder
func GetSettingsDir() string {

	usr, err := user.Current()
	if err != nil {
		return "./"
	}

	if runtime.GOOS == "linux" {
		return filepath.Join(usr.HomeDir, ".config/broxy/")
	}

	if runtime.GOOS == "darwin" {
		return filepath.Join(usr.HomeDir, ".config/broxy/")
	}

	if runtime.GOOS == "windows" {
		return filepath.Join(usr.HomeDir, ".\\broxy\\")
	}

	return "./"

}

// GetTmpDir returns the path to a temporary folder
func GetTmpDir() string {
	if runtime.GOOS == "linux" {
		return filepath.Join("/tmp")
	}

	if runtime.GOOS == "darwin" {
		return filepath.Join("/tmp")
	}

	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("TEMP"))
	}
	return "./"
}

func IsNil(i interface{}) bool {
   if i == nil {
      return true
   }
   switch reflect.TypeOf(i).Kind() {
   case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
      return reflect.ValueOf(i).IsNil()
   }
   return false
}
