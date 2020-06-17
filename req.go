package goreq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
)

func NewRequest(method, urladdr string) *Request {
	req, err := http.NewRequest(method, ModifyLink(urladdr), nil)
	return &Request{
		Request:    req,
		RespEncode: "",
		ProxyURL:   "",
		client:     DefaultClient,
		Err:        err,
		callback: func(resp *Response) *Response {
			return resp
		},
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

	callback func(resp *Response) *Response
	client   *Client

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

func (s *Request) SetBasicAuth(username, password string) *Request {
	if s.Err == nil {
		s.Request.SetBasicAuth(username, password)
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

type FormField struct {
	Name, Value string
}

type FormFile struct {
	FieldName, FileName, ContentType string
	File                             io.Reader
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func (s *Request) SetMultipartBody(data ...interface{}) *Request {
	if s.Err == nil {
		buff := bytes.NewBuffer([]byte{})
		wr := multipart.NewWriter(buff)
		for _, v := range data {
			switch v.(type) {
			case FormField:
				s.Err = wr.WriteField(v.(FormField).Name, v.(FormField).Value)
				if s.Err != nil {
					if Debug {
						fmt.Println(s.Err)
					}
					return s
				}
			case FormFile:
				var w io.Writer
				h := make(textproto.MIMEHeader)
				h.Set("Content-Disposition",
					fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
						escapeQuotes(v.(FormFile).FieldName), escapeQuotes(v.(FormFile).FieldName)))
				if v.(FormFile).ContentType != "" {
					h.Set("Content-Type", v.(FormFile).ContentType)
				} else {
					h.Set("Content-Type", "application/octet-stream")
				}
				w, s.Err = wr.CreatePart(h)
				if s.Err != nil {
					if Debug {
						fmt.Println(s.Err)
					}
					return s
				}
				_, s.Err = io.Copy(w, v.(FormFile).File)
				if s.Err != nil {
					if Debug {
						fmt.Println(s.Err)
					}
					return s
				}
			}
		}
		s.Err = wr.Close()
		if s.Err != nil {
			if Debug {
				fmt.Println(s.Err)
			}
			return s
		}
		s.SetBody(buff)
		s.Header.Set("Content-Type", wr.FormDataContentType())
	}
	return s
}

func (s *Request) SetCallback(fn func(resp *Response) *Response) *Request {
	s.callback = fn
	return s
}

func (s *Request) SetClient(c *Client) *Request {
	s.client = c
	return s
}

func (s *Request) Do() *Response {
	return s.callback(s.client.Do(s))
}

func (s *Request) String() string {
	return s.URL.String()
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
