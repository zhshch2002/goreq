# Goreq
[![goproxy.cn](https://goproxy.cn/stats/github.com/zhshch2002/goreq/badges/download-count.svg)](https://goproxy.cn)
![Go Test](https://github.com/zhshch2002/goreq/workflows/Go%20Test/badge.svg)
[![codecov](https://codecov.io/gh/zhshch2002/goreq/branch/master/graph/badge.svg)](https://codecov.io/gh/zhshch2002/goreq)

易用于网页、API 环境下的 Golang HTTP Client 封装库。

> 让`net/http`为人类服务。

```shell script
go get -u github.com/zhshch2002/goreq
```

## Feature
* Thread-safe | 线程安全
* Auto Charset Decode | 自动解码
* Easy to set proxy for each req | 便捷代理设置
* Chain config request | 链式配置请求
* Multipart post support
* Parse HTML,JSON,XML | HTML、JSON、XML 解析
* Middleware | 中间件
    * Cache | 缓存
    * Retry | 失败重试
    * Log | 日志
    * Random UserAgent | 随机 UA
    * Referer | 填充 Referer
    * Rate Delay and Parallelism limiter | 设置速率、延时、并发限制

**Goreq 是线程安全的**，意味着您无论在多线程还是单线程下开发，都无需改动代码。

**Goreq 会自动处理网页编码**，对于下载下来的网页，Goreq 会根据 HTTP 报头、内容推断编码并加以解码。而且您任可以访问原始的未解码内容。

## Request
Goreq 是设计为一般情况下访问网页和 API 而设计的“便携”工具。
```go
// `req`是一个请求`*goreq.Request`，是一个包装过的`net/http`下的 `*http.Request`
req := goreq.Get("https://httpbin.org/")

// 请求可以被链式配置，如果配置过程中出现错误，`req.Err`将不再为`nil`
req = goreq.Get("https://httpbin.org/get").AddParam("A","a")
if req.Err!=nil {
    panic(req.Err)
}
```

* `AddParam(k, v string)`
* `AddParams(v map[string]string)`
* `AddHeader(key, value string)`
* `AddHeaders(v map[string]string)`
* `AddCookie(c *http.Cookie)`
* `SetUA(ua string)`
* `SetBasicAuth(username, password string)`
* `SetProxy(urladdr string)`
* `SetTimeout(t time.Duration)`
* `DisableRedirect()`
* `SetCheckRedirect(fn func(req *http.Request, via []*http.Request) error)`
* Set request body data
    * `SetBody(b io.Reader)` basic setting
    * `SetRawBody(b []byte)`
    * `SetFormBody(v map[string]string)`
    * `SetJsonBody(v interface{})`
    * `SetMultipartBody(data ...interface{})` Set a slice of `FormField` and `FormFile` struct as body data
* `Callback(fn func(resp *Response)` Set a callback func run after req `Do()`

## Response
一个“请求”需要被“执行”来获得响应。

```go
resp:=goreq.Get("https://httpbin.org/get").AddParam("A","a").Do()

// 等效于执行了
resp:=goreq.DefaultClient.Do(goreq.Get("https://httpbin.org/get").AddParam("A","a"))
```
执行“请求”的是`*goreq.Client`，可以对其配置来添加中间件以实现扩展功能。

```go
c:=goreq.NewClient()
c.Use(goreq.WithRandomUA()) // 添加一个自动随机浏览器 UA 的内置中间件
resp:=goreq.Get("https://httpbin.org/get").AddParam("A","a").SetClient(c).Do()
fmt.Println(resp.Txt())
```

`*goreq.Response`可以通过下述函数来获取响应数据。
* `Resp() (*Response, error)` 获取响应本身以及网络请求错误。
* `Txt() (string, error)` 自动处理完编码并解析为文本后的内容以及网络请求错误。
* `HTML() (*goquery.Document, error)`
* `XML() (*xmlpath.Node, error)`
* `BindXML(i interface{}) error`
* `JSON() (gjson.Result, error)`
* `BindJSON(i interface{}) error`
* `Error() error` 网络请求错误。（正常情况下为`nil`）

## Middleware
```go
package main

import (
	"fmt"
	"github.com/zhshch2002/goreq"
)

func main() {
	// you can config `goreq.DefaultClient.Use()` to set global middleware
	c := goreq.NewClient() // create a new client
	c.Use(req.WithRandomUA()) // Add a builtin middleware
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
```
### Builtin middleware
* `WithDebug()`
* `WithCache(ca *cache.Cache)` Cache of `*Response` by go-cache
* `WithRetry(maxTimes int, isRespOk func(*Response)` set `nil` for `isRespOk` means no check func
* `WithProxy(p ...string)` set a list of proxy or it will follow `all_proxy` `https_proxy` and `http_proxy` env
* `WithRefererFiller()`
* `WithRandomUA()`
* Limiter | control Rate Delay and Parallelism
  * `WithFilterLimiter(noneMatchAllow bool, opts ...*FilterLimiterOpinion)` 
  * `WithDelayLimiter(eachSite bool, opts ...*DelayLimiterOpinion)` 
  * `WithRateLimiter(eachSite bool, opts ...*RateLimiterOpinion)` 
  * `WithParallelismLimiter(eachSite bool, opts ...*ParallelismLimiterOpinion)` 

### 失败重试 | WithRetry
```go
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
```

Output:
```
[Retry 1 times] got error on request https://httpbin.org/get <nil>
[Retry 2 times] got error on request https://httpbin.org/get <nil>
[Retry 3 times] got error on request https://httpbin.org/get <nil>
{
  "args": {},
  "headers": {
    "Accept-Encoding": "gzip",
    "Host": "httpbin.org",
    "User-Agent": "Go-http-client/2.0",
    "X-Amzn-Trace-Id": "Root=1-5efe9c40-bbf2d5a095e0f6d0c3aaf4c0"
  },
  "url": "https://httpbin.org/get"
}
```