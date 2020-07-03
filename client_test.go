package goreq

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClient_Use(t *testing.T) {
	m := 0
	c := NewClient(func(c *Client, h Handler) Handler {
		fmt.Println("Middleware 1")
		m += 1
		return func(req *Request) *Response {
			return h(req)
		}
	}, func(c *Client, h Handler) Handler {
		fmt.Println("Middleware 2")
		m += 1
		return func(req *Request) *Response {
			return h(req)
		}
	})
	fmt.Println(c.Do(Get("https://github.com/")).StatusCode)
	assert.Equal(t, 2, m)
}
