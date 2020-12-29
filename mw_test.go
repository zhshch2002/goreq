package goreq

import (
	"fmt"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWithRetry(t *testing.T) {
	i := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, i)
		i += 1
	}))
	defer ts.Close()
	c := NewClient()
	c.Use(WithRetry(10, func(resp *Response) bool {
		if i < 3 {
			return false
		}
		return true
	}))
	err := Get(ts.URL).SetDebug(true).SetClient(c).Do().Error()
	assert.NoError(t, err)
	assert.Equal(t, 3, i)
}

func TestWithCache(t *testing.T) {
	i := 1
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, i)
		i += 1
	}))
	defer ts.Close()
	cli := NewClient()
	cli.Use(WithCache(cache.New(10*time.Second, 10*time.Second)))
	a, err := Post(ts.URL).SetRawBody([]byte("test")).SetClient(cli).Do().Resp()
	assert.NoError(t, err)
	b, err := Post(ts.URL).SetRawBody([]byte("test")).SetClient(cli).Do().Resp()
	assert.NoError(t, err)
	fmt.Println(a.Text, b.Text)
	assert.Equal(t, a.Text, b.Text)
	c, err := Post(ts.URL).NoCache().SetRawBody([]byte("test")).SetClient(cli).Do().Resp()
	assert.NoError(t, err)
	fmt.Println(a.Text, c.Text)
	assert.NotEqual(t, a.Text, c.Text)
}

func TestWithRandomUA(t *testing.T) {
	c := NewClient()
	c.Use(WithRandomUA())
	resp, err := Get("https://httpbin.org/get").SetClient(c).Do().Resp()
	assert.NoError(t, err)
	t.Log(resp.Text)
	j, _ := resp.JSON()
	assert.NotEqual(t, "Go-http-client/2.0", j.Get("headers.User-Agent").String())
}

func TestWithRefererFiller(t *testing.T) {
	c := NewClient()
	c.Use(WithRefererFiller())
	resp, err := Get("https://httpbin.org/get").SetClient(c).Do().Resp()
	assert.NoError(t, err)
	t.Log(resp.Text)
	j, _ := resp.JSON()
	assert.True(t, j.Get("headers.Referer").Exists())
}

func TestWithDebug(t *testing.T) {
	c := NewClient()
	c.Use(WithDebug())
	err := Get("https://httpbin.org/get").SetClient(c).Do().Error()
	assert.NoError(t, err)
}
