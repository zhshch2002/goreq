package main

import (
	"fmt"
	"github.com/zhshch2002/goreq"
)

func main() {
	//req.DefaultClient.Use(req.WithCache(cache.New(1*time.Hour, 10*time.Minute)))
	//req.DefaultClient.Use(req.WithDebug())
	res := greq.Do(greq.Post("https://httpbin.org/post?hello=world").
		SetFormBody(map[string]string{
			"aaa": "123",
		}).AddParams(map[string]string{
		"bbb": "312",
	}).AddHeader("Req-Client", "GoReq").SetBasicAuth("me", "123456"))
	fmt.Println(res.Text)
	j, err := res.JSON()
	fmt.Println(err)
	fmt.Println(j.Get("form"))
}
