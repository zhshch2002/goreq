# Req
[![goproxy.cn](https://goproxy.cn/stats/github.com/zhshch2002/goreq/badges/download-count.svg)](https://goproxy.cn)
![Go Test](https://github.com/zhshch2002/goreq/workflows/Go%20Test/badge.svg)
[![codecov](https://codecov.io/gh/zhshch2002/goreq/branch/master/graph/badge.svg)](https://codecov.io/gh/zhshch2002/goreq)

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
	h, err := req.Get("https://httpbin.org/").Do().HTML()
	if err != nil {
		panic(err)
	}
	fmt.Println(h.Find("title").Text())
}
```

## Request

```go
// Create a request
req.Get("https://httpbin.org/")
req.Post("https://httpbin.org/")
req.Put("https://httpbin.org/")
// ...
```

### Config chain
```go
req.Get("https://httpbin.org/").
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
	resp := req.Post("https://httpbin.org/post?a=1").
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
			req.FormField{
				Name:  "d",
				Value: "4",
			},
			req.FormFile{
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
```

## Response
* type of(`req.Post("https://httpbin.org/post")`) is `*Request`
* type of(`req.Post("https://httpbin.org/post").SetUA("goreq")`) is `*Request`
* type of(`req.Post("https://httpbin.org/post").Do()`) is `*Response`

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
	resp := req.Get("https://example.com/").Do()
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