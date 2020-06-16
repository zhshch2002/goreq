package main

import (
	"bytes"
	"fmt"
	"github.com/zhshch2002/goreq"
	"net/http"
)

func main() {
	resp := greq.Post("https://httpbin.org/post?a=1").
		AddParam("b", "2").
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
			greq.FormField{
				Name:  "d",
				Value: "4",
			},
			greq.FormFile{
				FieldName:   "e",
				FileName:    "e.txt",
				ContentType: "",
				File:        bytes.NewReader([]byte("55555")),
			},
		).
		SetCallback(func(resp *req.Response) *req.Response {
			fmt.Println("here is the call back func")
			return resp
		}).
		Do()
	fmt.Println(resp.Text)
}
