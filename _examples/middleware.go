package main

import (
	"fmt"
	"github.com/zhshch2002/goreq"
)

func main() {
	// you can config `req.DefaultClient.Use()` to set global middleware
	c := goreq.NewClient()                                                  // create a new client
	c.Use(goreq.WithRandomUA())                                             // Add a builtin middleware
	c.Use(func(client *goreq.Client, handler goreq.Handler) goreq.Handler { // Add another middleware
		return func(r *goreq.Request) *goreq.Response {
			fmt.Println("this is a middleware")
			r.Header.Set("req", "goreq")
			return handler(r)
		}
	})

	txt, err := goreq.Get("https://httpbin.org/get").SetClient(c).Do().Txt()
	fmt.Println(txt, err)
}
