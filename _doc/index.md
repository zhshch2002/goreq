# Goreq

[![goproxycn](https://goproxy.cn/stats/github.com/zhshch2002/goreq/badges/download-count.svg)](https://goproxy.cn) ![Go Test](https://github.com/zhshch2002/goreq/workflows/Go%20Test/badge.svg) [![codecov](https://codecov.io/gh/zhshch2002/goreq/branch/master/graph/badge.svg)](https://codecov.io/gh/zhshch2002/goreq)

`Goreq`是对标准库`net/http`的包装。目的在于简化HTTP请求和接受数据的初步处理。`Goreq`主要是为了HTML网页和API请求设计的。

> 让`net/http`为人类服务。

```sh
go get -u github.com/zhshch2002/goreq
```

## Feature

- 线程安全
- 自动解码
- 便捷代理设置
- 链式配置请求
- 支持 Multipart post
- HTML、JSON、XML解析
- 中间件
  - 缓存
  - 失败重试
  - 随机UA
  - 填充Referer
  - 设置速率、延时、并发限制

**Goreq 是线程安全的**，意味着您无论在多线程还是单线程下开发，都无需改动代码。

**Goreq 会自动处理网页编码**，对于下载下来的网页，Goreq 会根据 HTTP 报头、内容推断编码并加以解码。而且您任可以访问原始的未解码内容。

在Goreq中主要有三个概念。

* **Request** 对`*http.Request`的封装。描述一个HTTP请求的地址、头部、代理、是否缓存等信息。
* **Client** 由 `*http.Client`组成。用于将`Request`转化为`Response`。
* **Response** 对``*http.Response``的封装。并经由`Client`自动处理编码。提供快速解析HTML、JSON、XML的接口。

## 构造请求

```go
type Request struct {
   *http.Request
   RespEncode string
   Writer io.Writer
   Debug bool
   callback func(resp *Response) *Response
   client   *Client
   Err error
}
```

```go
req := goreq.Get("https://httpbin.org/get?a=1").  // <- Notice here we got is req (as Request)
		AddParam("b", "2").
		AddHeaders(map[string]string{
			"req": "golang",
		}).
		AddCookie(&http.Cookie{
			Name:  "c",
			Value: "3",
		}).
		SetUA("goreq")
```

- SetDebug(d bool)
- AddParam(k, v string)
- AddParams(v map[string]string)
- AddHeader(key, value string)
- AddHeaders(v map[string]string)
- AddCookie(c *http.Cookie)
- AddCookies(cs ...*http.Cookie)
- SetUA(ua string)
- SetBasicAuth(username, password string)
- SetProxy(urladdr string)
- SetTimeout(t time.Duration)
- NoCache()
- SetCacheExpiration(e time.Duration)
- DisableRedirect()
- SetCheckRedirect(fn func(req \*http.Request, via []\*http.Request) error)
- 设置请求Body数据
  - SetBody(b io.Reader) basic setting
  - SetRawBody(b []byte)
  - SetFormBody(v map[string]string)
  - SetJsonBody(v interface{})
  - SetMultipartBody(data ...interface{})
- Callback(fn func(resp *Response)
- SetClient(c *Client) 这是一个很重要函数。Goreq有很多功能通过`Client`的中间件实现，为此需要使用自定义的`Client`执行请求。使用此函数可以改变调用`Do()`的目标`Client`。

## 发送请求

Request需要使用Client“执行”来得到Response。

```go
resp := goreq.Get("https://httpbin.org/get?a=1").
		AddParam("b", "2").
		AddHeaders(map[string]string{
			"req": "golang",
		}).
		AddCookie(&http.Cookie{
			Name:  "c",
			Value: "3",
		}).
		SetUA("goreq").Do()
```

这里使用了Goreq的全局默认`Client`执行。

```go
c := goreq.NewClient(goreq.WithRandomUA())
resp := goreq.Get("https://httpbin.org/get").SetClient(c).Do()
```

这里使用了自定义的`Client`，并使用了随机UA中间件。

## 获取数据

```go
type Response struct {
	*http.Response
	Body           []byte
	NotDecodedBody []byte
	Text           string
	Req            *Request
	CacheHash      string
	Err            error
}
```

- Resp() (*Response, error) 获取响应本身以及网络请求错误。
- Txt() (string, error) 自动处理完编码并解析为文本后的内容以及网络请求错误。
- RespAndTxt() (*Response, string, error)
- HTML() (*goquery.Document, error)
- RespAndHTML() (*Response, *goquery.Document, error)
- IsHTML() bool
- XML() (*xmlpath.Node, error)
- RespAndXML() (*Response, *xmlpath.Node, error)
- BindXML(i interface{}) error
- JSON() (gjson.Result, error)
- RespAndJSON() (*Response, gjson.Result, error)
- BindJSON(i interface{}) error
- IsJSON() bool
- Error() error 网络请求错误。（正常情况下为`nil`）