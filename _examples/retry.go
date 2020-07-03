package main

import (
	"fmt"
	"github.com/zhshch2002/goreq"
)

func main() {
	i := 0
	// 配置失败重试中间件，第二个参数函数用来检查是否为可接受的响应，传入 nil 使用默认函数。
	c := goreq.NewClient(goreq.WithRetry(10, func(resp *goreq.Response) bool {
		if i < 3 { // 为了演示模拟几次失败
			i += 1
			return false
		}
		return true
	}))
	fmt.Println(goreq.Get("https://httpbin.org/get").SetDebug(true).SetClient(c).Do().Text)
}
