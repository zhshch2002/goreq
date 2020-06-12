package req

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func init() {
	Debug = true
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
