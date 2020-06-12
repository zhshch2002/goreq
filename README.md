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
	// make a request
	res := req.Do(req.Get("https://httpbin.org/get"))
	if res.Err != nil {
		fmt.Println(res.Err)
	} else {
		fmt.Println(res.Text)
	}

	// set params
	// AddHeader,AddHeaders,AddCookie work as the same way
	if res, err := req.Do(
		req.Get("https://httpbin.org/get?hi=myself").
			AddParam("aaa", "111"). // set a single param
			AddParams(map[string]string{
				"bbb": "222",
			}), // set a group param
	).Resp(); err == nil {
		fmt.Println(res.Text)
	} else {
		fmt.Println(err)
	}
	/*
		Output:
			{
			  "args": {
			    "aaa": "111",
			    "bbb": "222",
			    "hi": "myself"
			  },
			  "headers": {
			    "Accept-Encoding": "gzip",
			    "Host": "httpbin.org",
			    "User-Agent": "Go-http-client/2.0",
			    "X-Amzn-Trace-Id": "Root=1-5ed1a57f-f84b50dcdd2d1a4680190bcf"
			  },
			  "url": "https://httpbin.org/get?hi=myself&aaa=111&bbb=222"
			}
	*/

	// parse response
	if res, err := req.Do(
		req.Get("https://httpbin.org/get?hi=myself").
			AddParam("aaa", "111"). // set a single param
			AddParams(map[string]string{
				"bbb": "222",
			}), // set a group param
	).Resp(); err == nil {
		if j, err := res.JSON(); err == nil {
			fmt.Println("as json", j.Get("args"))
		}
		if h, err := res.HTML(); err == nil {
			fmt.Println("as html")
			fmt.Println(h.Html())
		}
		if h, err := res.XML(); err == nil {
			fmt.Println("as XML")
			fmt.Println(h.String())
		}
	} else {
		fmt.Println(err)
	}

	// active middleware
	c:=req.NewClient()
	c.Use(req.WithRandomUA())
	// cache the duplicate request
	// c.Use(req.WithCache(cache.New(1*time.Hour, 10*time.Minute)))
	// use proxy or follow http_proxy env var
	// c.Use(req.WithProxy())
	if res, err := req.Do(
		req.Get("https://httpbin.org/get"),
	).Resp(); err == nil {
		fmt.Println(res.Text)
	} else {
		fmt.Println(err)
	}
}
```