package greq

import (
	"context"
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

type Client struct {
	cli        *http.Client
	middleware []Middleware
	handler    Handler
}

func NewClient() *Client {
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
		},
		middleware: []Middleware{},
	}
	c.handler = basicHttpDo(c, nil)
	return c
}

func (s *Client) Use(m ...Middleware) {
	s.middleware = append(s.middleware, m...)
	//s.handler = basicHttpDo(s, nil)
	for i := len(s.middleware) - 1; i >= 0; i-- {
		s.handler = s.middleware[i](s, s.handler)
	}
}

func (s *Client) Do(req *Request) *Response {
	res := s.handler(req)
	res.Err = res.DecodeAndParse()
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
		if req.Err != nil {
			return resp
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
