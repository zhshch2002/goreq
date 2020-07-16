package goreq

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

var DefaultClient = NewClient()

var Debug = false //TODO

func Do(req *Request) *Response {
	return DefaultClient.Do(req)
}

type Handler func(*Request) *Response
type Middleware func(*Client, Handler) Handler

type ReqError struct {
	error
}

var ReqRejectedErr = errors.New("request is rejected")

type Client struct {
	cli        *http.Client
	middleware []Middleware
	handler    Handler
}

func NewClient(m ...Middleware) *Client {
	j, _ := cookiejar.New(nil)
	c := &Client{
		cli: &http.Client{
			Jar: j,
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					if addr, ok := req.Context().Value("proxy").(string); ok && addr != "" {
						return url.Parse(addr)
					}
					return nil, nil
				},
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if fn, ok := req.Context().Value("CheckRedirect").(func(*http.Request, []*http.Request) error); ok && fn != nil {
					return fn(req, via)
				}
				if len(via) >= 10 {
					return errors.New("stopped after 10 redirects")
				}
				return nil
			},
		},
		middleware: []Middleware{},
	}
	c.handler = basicHttpDo(c, nil)
	c.Use(m...)
	return c
}

func (s *Client) Use(m ...Middleware) *Client {
	s.middleware = append(s.middleware, m...)
	//s.handler = basicHttpDo(s, nil)
	for i := 0; i < len(s.middleware); i++ {
		s.handler = s.middleware[i](s, s.handler)
	}
	return s
}

func (s *Client) Do(req *Request) *Response {
	if req.Err != nil {
		return &Response{
			Req: req,
			Err: ReqError{req.Err},
		}
	}
	res := s.handler(req)
	if res == nil {
		return &Response{
			Req: req,
			Err: ReqRejectedErr,
		}
	}
	if res.Err == nil {
		res.Err = res.DecodeAndParse()
	}
	return res
}

func basicHttpDo(c *Client, next Handler) Handler {
	return func(req *Request) *Response {
		resp := &Response{
			Req:  req,
			Text: "",
			Body: []byte{},
			Err:  req.Err,
		}

		if req.ProxyURL != "" {
			req.Request = req.Request.WithContext(context.WithValue(req.Request.Context(), "proxy", req.ProxyURL))
		}

		resp.Response, resp.Err = c.cli.Do(req.Request)
		if resp.Err != nil {
			return resp
		}
		defer resp.Response.Body.Close()

		resp.Body, resp.Err = ioutil.ReadAll(resp.Response.Body)
		if resp.Err != nil {
			return resp
		}
		return resp
	}
}
