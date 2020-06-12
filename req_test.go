package req

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
)

func init() {
	Debug = true
}

func TestMethods(t *testing.T) {
	cb := func(resp *Response) *Response {
		fmt.Println(resp.Text)
		return resp
	}
	assert.Nil(t, Get("https://httpbin.org/get").Callback(cb).Do().Error())
	assert.Nil(t, Post("https://httpbin.org/post").Callback(cb).Do().Error())
	assert.Nil(t, Head("https://httpbin.org/head").Callback(cb).Do().Error())
	assert.Nil(t, Put("https://httpbin.org/put").Callback(cb).Do().Error())
	assert.Nil(t, Delete("https://httpbin.org/delete").Callback(cb).Do().Error())
	assert.Nil(t, Connect("https://httpbin.org/connect").Callback(cb).Do().Error())
	assert.Nil(t, Options("https://httpbin.org/options").Callback(cb).Do().Error())
	assert.Nil(t, Trace("https://httpbin.org/trace").Callback(cb).Do().Error())
	assert.Nil(t, Patch("https://httpbin.org/patch").Callback(cb).Do().Error())
}

func TestGet(t *testing.T) {
	resp := Get("https://httpbin.org/get").Do()
	t.Log(resp.Text)
	if resp.Err != nil {
		t.Error(resp.Err)
	}
}

func TestPost(t *testing.T) {
	resp := Post("https://httpbin.org/post").Do()
	t.Log(resp.Text)
	assert.Nil(t, resp.Err)
}

func TestRequest_DoCallback(t *testing.T) {
	s := make(chan struct{})
	go Get("https://httpbin.org/get").Callback(func(resp *Response) *Response {
		t.Log(resp.Text)
		assert.Nil(t, resp.Err)
		s <- struct{}{}
		return resp
	}).Do()
	_ = <-s
}

func TestRequest_SetMultipartBody(t *testing.T) {
	f, err := os.Open("./req.go")
	assert.Nil(t, err)
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
	assert.Nil(t, resp.Err)
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
	assert.Nil(t, err)
	assert.Equal(t, "1", j.Get("args.a").String())
	assert.Equal(t, "2", j.Get("args.b").String())
	assert.Equal(t, "c=3", j.Get("headers.Cookie").String())
	assert.Equal(t, "4", j.Get("form.d").String())
	assert.Equal(t, "5", j.Get("files.e").String())
	assert.Equal(t, "Basic Z29yZXE6Z29sYW5n", j.Get("headers.Authorization").String())
	assert.Equal(t, "golang", j.Get("headers.Req").String())
	assert.Equal(t, "goreq", j.Get("headers.User-Agent").String())
}
