package main

import (
	"fmt"
	"req"
)

func main() {
	//req.DefaultClient.Use(req.WithCache(cache.New(1*time.Hour, 10*time.Minute)))
	//req.DefaultClient.Use(req.WithDebug())
	res := req.Do(req.Post("https://httpbin.org/post").SetRawBody([]byte("aaaaa")))
	fmt.Println(res)
}
