package goreq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
	"time"
)

func NewRequest(method, urladdr string) *Request {
	req, err := http.NewRequest(method, ModifyLink(urladdr), nil)
	return &Request{
		Request:    req,
		RespEncode: "",
		client:     DefaultClient,
		Err:        err,
		Debug:      Debug,
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

	RespEncode string

	Writer io.Writer

	Debug bool

	callback func(resp *Response) *Response
	client   *Client

	Err error
}

func (s *Request) addContextValue(k, v interface{}) *Request {
	s.Request = s.WithContext(context.WithValue(s.Request.Context(), k, v))
	return s
}

func (s *Request) SetDebug(d bool) *Request {
	s.Debug = d
	return s
}

func (s *Request) SetTimeout(t time.Duration) *Request {
	ctx, _ := context.WithTimeout(s.Context(), t)
	s.Request = s.WithContext(ctx)
	return s
}

type ctxProxyType struct{}

var ctxProxy = &ctxProxyType{}

func (s *Request) SetProxy(urladdr string) *Request {
	u, err := url.Parse(urladdr)
	if err != nil {
		if s.Debug {
			log.Println("set proxy error:", err)
		}
		s.Err = err
		return s
	}
	return s.addContextValue(ctxProxy, u)
}

type ctxNoCacheType struct{}

var ctxNoCache = &ctxNoCacheType{}

func (s *Request) NoCache() *Request {
	return s.addContextValue(ctxNoCache, struct{}{})
}

type ctxCacheExpirationType struct{}

var ctxCacheExpiration = &ctxCacheExpirationType{}

func (s *Request) SetCacheExpiration(e time.Duration) *Request {
	return s.addContextValue(ctxCacheExpiration, e)
}

type ctxCheckRedirectType struct{}

var ctxCheckRedirect = &ctxCheckRedirectType{}

func (s *Request) SetCheckRedirect(fn func(req *http.Request, via []*http.Request) error) *Request {
	return s.addContextValue(ctxCheckRedirect, fn)
}

func (s *Request) DisableRedirect() *Request {
	s.SetCheckRedirect(func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	})
	return s
}

// AddCookie adds a cookie to the request.
func (s *Request) AddCookie(c *http.Cookie) *Request {
	s.Request.AddCookie(c)
	return s
}

// AddCookies adds some cookie to the request at once.
func (s *Request) AddCookies(cs ...*http.Cookie) *Request {
	for _, c := range cs {
		s.Request.AddCookie(c)
	}
	return s
}

// AddHeader sets the header entries associated with key
// to the single element value.
func (s *Request) AddHeader(key, value string) *Request {
	s.Request.Header.Add(key, value)
	return s
}
func (s *Request) AddHeaders(v map[string]string) *Request {
	for k, v := range v {
		s.AddHeader(k, v)
	}
	return s
}

// SetUA sets user-agent url of request header.
func (s *Request) SetUA(ua string) *Request {
	s.AddHeader("User-Agent", ua)
	return s
}

// AddParam adds a query param of request url.
func (s *Request) AddParam(k, v string) *Request {
	if len(s.Request.URL.RawQuery) > 0 {
		s.Request.URL.RawQuery += "&"
	}
	s.Request.URL.RawQuery += url.QueryEscape(k) + "=" + url.QueryEscape(v)
	return s
}
func (s *Request) AddParams(v map[string]string) *Request {
	for k, v := range v {
		s.AddParam(k, v)
	}
	return s
}

func (s *Request) SetBasicAuth(username, password string) *Request {
	s.Request.SetBasicAuth(username, password)
	return s
}

func (s *Request) SetBody(b io.Reader) *Request {
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
	return s
}
func (s *Request) SetRawBody(b []byte) *Request {
	s.SetBody(bytes.NewReader(b))
	return s
}
func (s *Request) SetFormBody(v map[string]string) *Request {
	var u url.URL
	q := u.Query()
	for k, v := range v {
		q.Add(k, v)
	}
	s.SetRawBody([]byte(q.Encode()))
	s.AddHeader("Content-Type", "application/x-www-form-urlencoded")
	return s
}
func (s *Request) SetJsonBody(v interface{}) *Request {
	body, err := json.Marshal(v)
	s.SetRawBody(body)
	s.Err = err
	s.AddHeader("Content-Type", "application/json")
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
	if s.Err != nil {
		if s.Debug {
			log.Println("request has error", s.Err)
		}
		return s
	}
	buff := bytes.NewBuffer([]byte{})
	wr := multipart.NewWriter(buff)
	for _, v := range data {
		switch v.(type) {
		case FormField:
			s.Err = wr.WriteField(v.(FormField).Name, v.(FormField).Value)
			if s.Err != nil {
				if s.Debug {
					log.Println("set multipart FormField body error", s.Err)
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
				if s.Debug {
					log.Println("set multipart body FormFile error", s.Err)
				}
				return s
			}
			_, s.Err = io.Copy(w, v.(FormFile).File)
			if s.Err != nil {
				if s.Debug {
					log.Println("set multipart body FormFile error", s.Err)
				}
				return s
			}
		}
	}
	s.Err = wr.Close()
	if s.Err != nil {
		if s.Debug {
			log.Println("set multipart body FormFile error", s.Err)
		}
		return s
	}
	s.SetBody(buff)
	s.Header.Set("Content-Type", wr.FormDataContentType())
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
