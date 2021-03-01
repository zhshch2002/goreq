package goreq

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

var DefaultClient = NewClient()

var Debug = false

func Do(req *Request) *Response {
	return DefaultClient.Do(req)
}

type Handler func(*Request) *Response
type Middleware func(*Client, Handler) Handler

type RequestError struct {
	error
}

var ReqRejectedErr = errors.New("request is rejected")

type Client struct {
	Client  *http.Client
	handler Handler
}

func NewClient(m ...Middleware) *Client {
	j, _ := cookiejar.New(nil)
	c := &Client{
		Client: &http.Client{
			Jar: j,
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					if addr, ok := req.Context().Value(ctxProxy).(string); ok && addr != "" {
						return url.Parse(addr)
					}
					return nil, nil
				},
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if fn, ok := req.Context().Value(ctxCheckRedirect).(func(*http.Request, []*http.Request) error); ok && fn != nil {
					return fn(req, via)
				}
				if len(via) >= 10 {
					return errors.New("stopped after 10 redirects")
				}
				return nil
			},
		},
	}
	c.handler = basicHttpDo(c, nil)
	c.Use(m...)
	return c
}

func (s *Client) Use(mid ...Middleware) *Client {
	for _, m := range mid {
		s.handler = m(s, s.handler)
	}
	return s
}

func (s *Client) Do(req *Request) *Response {
	if req.Err != nil {
		return &Response{
			Req: req,
			Err: RequestError{req.Err},
		}
	}
	res := s.handler(req)
	if res == nil {
		return &Response{
			Req: req,
			Err: ReqRejectedErr,
		}
	}
	if len(res.NotDecodedBody) == 0 && res.Err == nil {
		res.Err = res.DecodeAndParse()
	}
	return res
}

func basicHttpDo(c *Client, next Handler) Handler {
	return func(req *Request) *Response {
		resp := &Response{
			Req:         req,
			Text:        "",
			Body:        []byte{},
			IsFromCache: false,
		}

		resp.Response, resp.Err = c.Client.Do(req.Request)
		if resp.Err != nil {
			return resp
		}
		defer resp.Response.Body.Close()

		resp.Body, resp.Err = ioutil.ReadAll(resp.Response.Body)
		if resp.Err != nil {
			return resp
		}
		if resp.Err == nil {
			resp.Err = resp.DecodeAndParse()
		}
		return resp
	}
}
