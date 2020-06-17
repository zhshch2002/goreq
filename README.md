# Req
[![goproxy.cn](https://goproxy.cn/stats/github.com/zhshch2002/goreq/badges/download-count.svg)](https://goproxy.cn)
![Go Test](https://github.com/zhshch2002/goreq/workflows/Go%20Test/badge.svg)
[![codecov](https://codecov.io/gh/zhshch2002/goreq/branch/master/graph/badge.svg)](https://codecov.io/gh/zhshch2002/goreq)

An clean and simple Go HTTP request client library.
一个优雅并简洁的Go HTTP请求库。

```shell script
go get -u github.com/zhshch2002/goreq
```

> Warning
>
> goreq还处于v0的状态，缺少完整的文档和大量的测试。当前goreq已经可以稳定使用并用于我的一些个人项目。
> goreq可以保证正常使用并处理issues反馈，但意外情况下不保证API和部分功能一定不变或向下兼容。
>
> Goreq is still in v0 state, lacking complete documentation and extensive testing. At present, goreq can be used steadily and used in some of my personal projects.
> Goreq can ensure normal use and process issues feedback, but in unexpected cases, it does not guarantee that the API and some functions will be unchanged or backward compatible.


## Feature
* Thread-safe | 线程安全
* Auto Charset Decode | 自动解码
* [Easy to set proxy for each req | 便捷代理设置](#Request)
* [Chain config request | 链式配置请求](#Request)
* [Multipart post support](#Request)
* [Parse HTML,JSON,XML | HTML、JSON、XML解析](#Response)
* [Middleware | 中间件](#Middleware)
    * Cache | 缓存
    * Retry | 失败重试
    * Log | 日志
    * Random UserAgent | 随机UA
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
	h, err := greq.Get("https://httpbin.org/").Do().HTML()
	if err != nil {
		panic(err)
	}
	fmt.Println(h.Find("title").Text())
}
```

### Request

```go
// Create a request
greq.Get("https://httpbin.org/")
greq.Post("https://httpbin.org/")
greq.Put("https://httpbin.org/")
// ...
```

#### Config chain
```go
greq.Get("https://httpbin.org/").
    AddHeader("Req-Client", "GoReq").
    AddParams(map[string]string{ // https://httpbin.org/?bbb=312 
        "bbb": "312",
    })
```

* `AddParam(k, v string)`
* `AddParams(v map[string]string)`
* `AddHeader(key, value string)`
* `AddHeaders(v map[string]string)`
* `AddCookie(c *http.Cookie)`
* `SetUA(ua string)`
* `SetBasicAuth(username, password string)`
* `SetProxy(urladdr string)`
* Set request body data
    * `SetBody(b io.Reader)` basic setting
    * `SetRawBody(b []byte)`
    * `SetFormBody(v map[string]string)`
    * `SetJsonBody(v interface{})`
    * `SetMultipartBody(data ...interface{})` Set a slice of `FormField` and `FormFile` struct as body data
* `Callback(fn func(resp *Response)` Set a callback func run after req `Do()`

```go
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
		SetCallback(func(resp *greq.Response) *greq.Response {
			fmt.Println("here is the call back func")
			return resp
		}).
		Do()
	fmt.Println(resp.Text)
}
```

### Response
* type of(`greq.Post("https://httpbin.org/post")`) is `*Request`
* type of(`greq.Post("https://httpbin.org/post").SetUA("goreq")`) is `*Request`
* type of(`greq.Post("https://httpbin.org/post").Do()`) is `*Response`

After calling the request's `Do()`, it will return a `*Response` and execute Callback
```go
func (s *Request) Do() *Response {
	return s.callback(s.client.Do(s))
}
```

The `*Response` contains the request and the decoded body.

```go
package main

import (
	"fmt"
	"github.com/zhshch2002/goreq"
)

func main() {
	resp := greq.Get("https://example.com/").Do()
	if resp.Err != nil {
		panic(resp.Err)
	}
	fmt.Println(resp.Text) // Get the decode text,same as `text,err:=resp.Txt()`

	j, err := resp.JSON() // Parse as json with gjson
	fmt.Println(resp.IsJSON(), j, err)

	h, err := resp.HTML() // Parse as html with goquery
	fmt.Println(resp.IsHTML(), h, err)

	x, err := resp.XML() // Parse as xml with xmlpath
	fmt.Println(x, err)

	var data struct {
		Url string `json:"url" xml:"url"`
	}
	err = resp.BindJSON(&data) // Parse as json
	fmt.Println(data, err)
	err = resp.BindXML(&data) // Parse as xml
	fmt.Println(data, err)
}
```

### Middleware
```go
package main

import (
	"fmt"
	"github.com/zhshch2002/goreq"
)

func main() {
	// you can config `greq.DefaultClient.Use()` to set global middleware
	c := greq.NewClient() // create a new client
	c.Use(req.WithRandomUA()) // Add a builtin middleware
	c.Use(func(client *greq.Client, handler greq.Handler) greq.Handler { // Add another middleware
		return func(r *greq.Request) *greq.Response {
			fmt.Println("this is a middleware")
			r.Header.Set("req", "goreq")
			return handler(r)
		}
	})

	txt, err := greq.Get("https://httpbin.org/get").SetClient(c).Do().Txt()
	fmt.Println(txt, err)
}
```
#### Builtin middleware
* `WithDebug()`
* `WithCache(ca *cache.Cache)` Cache of `*Response` by go-cache
* `WithRetry(maxTimes int, isRespOk func(*Response)` set `nil` for `isRespOk` means no check func
* `WithProxy(p ...string)` set a list of proxy or it will follow `all_proxy` `https_proxy` and `http_proxy` env
* `WithRefererFiller()`
* `WithRandomUA()`