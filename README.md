# Req
An elegant and concise Go HTTP request library.
一个优雅并简洁的Go HTTP请求库。

```shell script
go get -u github.com/zhshch2002/goreq
```

## Feature
* Auto Charset Decode | 自动解码
* Easy to set proxy for each req | 便捷代理设置
* Parse HTML,JSON,XML | HTML、JSON、XML解析
* Config req as chain | 链式配置请求
* Multipart post support
* Middleware | 中间件
    * Cache | 缓存
    * Retry | 失败重试
    * Log | 日志
    * Random UserAgent| 随机UA
    * Referer | 填充Referer
### TODO
* Download & Upload
* Format Print

## Usage

```go
package main

import (
	"fmt"
	"github.com/zhshch2002/goreq"
)


func main() {
	h,err:=req.Get("https://httpbin.org/").Do().HTML()
	if err != nil {
		panic(err)
	}
    fmt.Println(h.Find("title").Text())
}
```