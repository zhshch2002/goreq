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
	Debug = true
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, i)
		i += 1
	}))
	c := NewClient()
	c.Use(WithRetry(10, func(resp *Response) bool {
		if i < 3 {
			return false
		}
		return true
	}))
	err := Get(ts.URL).SetClient(c).Do().Error()
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
	c := NewClient()
	c.Use(WithCache(cache.New(10*time.Second, 10*time.Second)))
	a, err := Get(ts.URL).SetClient(c).Do().Txt()
	assert.NoError(t, err)
	b, err := Get(ts.URL).SetClient(c).Do().Txt()
	assert.NoError(t, err)
	fmt.Println(a, b)
	assert.Equal(t, a, b)
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
