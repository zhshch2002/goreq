package req

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func NewRequest(method, urladdr string) *Request {
	req, err := http.NewRequest(method, urladdr, nil)
	return &Request{
		Request:    req,
		RespEncode: "",
		ProxyURL:   "",
		Err:        err,
	}
}

func Get(urladdr string) *Request {
	return NewRequest("GET", urladdr)
}
func Post(urladdr string) *Request {
	return NewRequest("POST", urladdr)
}
func Head(urladdr string) *Request {
	return NewRequest("HEAD", urladdr)
}
func Put(urladdr string) *Request {
	return NewRequest("PUT", urladdr)
}
func Delete(urladdr string) *Request {
	return NewRequest("DELETE", urladdr)
}
func Connect(urladdr string) *Request {
	return NewRequest("CONNECT", urladdr)
}
func Options(urladdr string) *Request {
	return NewRequest("OPTIONS", urladdr)
}
func Trace(urladdr string) *Request {
	return NewRequest("TRACE", urladdr)
}
func Patch(urladdr string) *Request {
	return NewRequest("PATCH", urladdr)
}

// Request is a object of HTTP request
type Request struct {
	*http.Request
	// ProxyURL is the proxy address that handles the request
	ProxyURL string

	RespEncode string

	Writer io.Writer

	Err error
}

func (s *Request) SetProxy(urladdr string) *Request {
	s.ProxyURL = urladdr
	return s
}

// AddCookie adds a cookie to the request.
func (s *Request) AddCookie(c *http.Cookie) *Request {
	if s.Err == nil {
		s.Request.AddCookie(c)
	}
	return s
}

// SetHeader sets the header entries associated with key
// to the single element value.
func (s *Request) AddHeader(key, value string) *Request {
	if s.Err == nil {
		s.Request.Header.Add(key, value)
	}
	return s
}
func (s *Request) AddHeaders(v map[string]string) *Request {
	if s.Err == nil {
		for k, v := range v {
			s.AddHeader(k, v)
		}
	}
	return s
}

// SetProxy sets user-agent url of request header.
func (s *Request) SetUA(ua string) *Request {
	if s.Err == nil {
		s.AddHeader("User-Agent", ua)
	}
	return s
}

// AddParam adds a query param of request url.
func (s *Request) AddParam(k, v string) *Request {
	if s.Err == nil {
		if len(s.Request.URL.RawQuery) > 0 {
			s.Request.URL.RawQuery += "&"
		}
		s.Request.URL.RawQuery += url.QueryEscape(k) + "=" + url.QueryEscape(v)
	}
	return s
}
func (s *Request) AddParams(v map[string]string) *Request {
	if s.Err == nil {
		for k, v := range v {
			s.AddParam(k, v)
		}
	}
	return s
}

func (s *Request) SetBody(b io.Reader) *Request {
	if s.Err == nil {
		rc, ok := b.(io.ReadCloser)
		if !ok && b != nil {
			rc = ioutil.NopCloser(b)
		}
		s.Request.Body = rc

		switch v := b.(type) {
		case *bytes.Buffer:
			s.ContentLength = int64(v.Len())
			buf := v.Bytes()
			s.GetBody = func() (io.ReadCloser, error) {
				r := bytes.NewReader(buf)
				return ioutil.NopCloser(r), nil
			}
		case *bytes.Reader:
			s.ContentLength = int64(v.Len())
			snapshot := *v
			s.GetBody = func() (io.ReadCloser, error) {
				r := snapshot
				return ioutil.NopCloser(&r), nil
			}
		case *strings.Reader:
			s.ContentLength = int64(v.Len())
			snapshot := *v
			s.GetBody = func() (io.ReadCloser, error) {
				r := snapshot
				return ioutil.NopCloser(&r), nil
			}
		default:
		}
	}
	return s
}
func (s *Request) SetRawBody(b []byte) *Request {
	if s.Err == nil {
		s.SetBody(bytes.NewReader(b))
	}
	return s
}
func (s *Request) SetFormBody(v map[string]string) *Request {
	if s.Err == nil {
		var u url.URL
		q := u.Query()
		for k, v := range v {
			q.Add(k, v)
		}
		s.SetRawBody([]byte(q.Encode()))
		s.AddHeader("Content-Type", "application/x-www-form-urlencoded")
	}
	return s
}
func (s *Request) SetJsonBody(v interface{}) *Request {
	if s.Err == nil {
		body, err := json.Marshal(v)
		s.SetRawBody(body)
		s.Err = err
		s.AddHeader("Content-Type", "application/json")
	}
	return s
}

//func (s *Request) Format(f fmt.State, c rune) {
//	if s == nil {
//		fmt.Print(nil)
//		return
//	}
//	if s.Err != nil {
//		fmt.Println("request error", s.Err)
//		return
//	}
//
//	if f.Flag('+') {
//		fmt.Println(s.Method, s.URL.Path, s.Proto)
//		for k, v := range s.Header {
//			for _, a := range v {
//				fmt.Println(k+":", a)
//			}
//		}
//		if r, err := s.GetBody(); err == nil {
//			if b, err := ioutil.ReadAll(r); err == nil {
//				fmt.Print("\n", b, "\n")
//			}
//
//		}
//	} else {
//		fmt.Println(s.Method, s.URL)
//	}
//}
