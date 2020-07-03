package goreq

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMethods(t *testing.T) {
	cb := func(resp *Response) *Response {
		fmt.Println(resp.Text)
		return resp
	}
	assert.NoError(t, Get("https://httpbin.org/get").SetCallback(cb).Do().Error())
	assert.NoError(t, Post("https://httpbin.org/post").SetCallback(cb).Do().Error())
	assert.NoError(t, Head("https://httpbin.org/head").SetCallback(cb).Do().Error())
	assert.NoError(t, Put("https://httpbin.org/put").SetCallback(cb).Do().Error())
	assert.NoError(t, Delete("https://httpbin.org/delete").SetCallback(cb).Do().Error())
	assert.NoError(t, Connect("https://httpbin.org/connect").SetCallback(cb).Do().Error())
	assert.NoError(t, Options("https://httpbin.org/options").SetCallback(cb).Do().Error())
	assert.NoError(t, Trace("https://httpbin.org/trace").SetCallback(cb).Do().Error())
	assert.NoError(t, Patch("https://httpbin.org/patch").SetCallback(cb).Do().Error())
}

func TestGet(t *testing.T) {
	resp := Get("https://httpbin.org/get").Do()
	t.Log(resp.Text)
	assert.NoError(t, resp.Err)
}

func TestPost(t *testing.T) {
	resp := Post("https://httpbin.org/post").Do()
	t.Log(resp.Text)
	assert.NoError(t, resp.Err)
}

func TestRequest_DoCallback(t *testing.T) {
	s := make(chan struct{})
	go Get("https://httpbin.org/get").SetCallback(func(resp *Response) *Response {
		t.Log(resp.Text)
		assert.NoError(t, resp.Err)
		s <- struct{}{}
		return resp
	}).Do()
	_ = <-s
}

func TestRequest_SetMultipartBody(t *testing.T) {
	f, err := os.Open("./req.go")
	assert.NoError(t, err)
	resp := Post("https://httpbin.org/post").SetMultipartBody(
		FormField{
			Name:  "AAA",
			Value: "BBB",
		},
		FormFile{
			FieldName:   "CCC",
			FileName:    "req.go",
			ContentType: "",
			File:        f,
		},
	).Do()
	t.Log(resp.Text)
	assert.NoError(t, resp.Err)
}

func TestRequest_SetFormBody(t *testing.T) {
	resp := Post("https://httpbin.org/post").SetFormBody(map[string]string{
		"a": "1",
	}).Do()
	t.Log(resp.Text)
	assert.NoError(t, resp.Err)
	j, _ := resp.JSON()
	assert.Equal(t, "1", j.Get("form.a").String())
}

func TestRequest_SetJsonBody(t *testing.T) {
	resp := Post("https://httpbin.org/post").SetJsonBody(map[string]string{
		"a": "1",
	}).Do()
	t.Log(resp.Text)
	assert.NoError(t, resp.Err)
	j, _ := resp.JSON()
	assert.Equal(t, "1", j.Get("json.a").String())
}

func TestRequest_SetRawBody(t *testing.T) {
	resp := Post("https://httpbin.org/post").SetRawBody([]byte("1")).Do()
	t.Log(resp.Text)
	assert.NoError(t, resp.Err)
	j, _ := resp.JSON()
	assert.Equal(t, "1", j.Get("data").String())
}

func TestRequest_Do(t *testing.T) {
	resp := Post("https://httpbin.org/post?a=1").
		AddParams(map[string]string{
			"b": "2",
		}).
		AddHeaders(map[string]string{
			"req": "golang",
		}).
		AddCookie(&http.Cookie{
			Name:  "c",
			Value: "3",
		}).
		SetUA("goreq").
		SetBasicAuth("goreq", "golang").
		//SetProxy("http://127.0.0.1:1080/").
		SetMultipartBody(
			FormField{
				Name:  "d",
				Value: "4",
			},
			FormFile{
				FieldName:   "e",
				FileName:    "e.txt",
				ContentType: "",
				File:        bytes.NewReader([]byte("5")),
			},
		).
		Do()
	fmt.Println(resp.Text)
	j, err := resp.JSON()
	assert.NoError(t, err)
	assert.Equal(t, "1", j.Get("args.a").String())
	assert.Equal(t, "2", j.Get("args.b").String())
	assert.Equal(t, "c=3", j.Get("headers.Cookie").String())
	assert.Equal(t, "4", j.Get("form.d").String())
	assert.Equal(t, "5", j.Get("files.e").String())
	assert.Equal(t, "Basic Z29yZXE6Z29sYW5n", j.Get("headers.Authorization").String())
	assert.Equal(t, "golang", j.Get("headers.Req").String())
	assert.Equal(t, "goreq", j.Get("headers.User-Agent").String())
}
func setupProxy(t *testing.T) *gin.Engine {
	r := gin.New()
	r.GET("/:a", func(c *gin.Context) {
		all, err := ioutil.ReadAll(c.Request.Body)
		assert.NoError(t, err)
		c.String(200, string(all))
	})
	return r
}

func TestProxy(t *testing.T) {
	router := setupProxy(t)
	ts := httptest.NewServer(http.HandlerFunc(router.ServeHTTP))
	defer ts.Close()
	proxyTs := httptest.NewServer(http.HandlerFunc(router.ServeHTTP))
	defer proxyTs.Close()
	txt, err := Get(ts.URL + "/login").SetRawBody([]byte(proxyTs.URL)).SetProxy(proxyTs.URL).Do().Txt()
	assert.NoError(t, err)
	assert.Equal(t, txt, proxyTs.URL)
}
