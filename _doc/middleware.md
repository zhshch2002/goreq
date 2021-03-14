# 中间件

## 内建中间件

### 限制器 Limiter

限制器是一组Goreq提供的中间件。可以轻松的过滤可请求、设置延时、限制速率、并发。

如下，这个Client将不会接受发送向*.taobao.com的请求。

```go go l
c := goreq.NewClient(goreq.WithFilterLimiter(true, &goreq.FilterLimiterOpinion{
  LimiterMatcher: goreq.LimiterMatcher{
    Glob: "*.taobao.com",
  },
  Allow: false,
}))
fmt.Println(c.Do(goreq.Get("https://www.taobao.com/")).Err) // ReqRejectedErr
```

#### LimiterMatcher

在上面的例子中，出现了`LimiterMatcher`。Goreq的每个限制器都会要求设置这个参数。`LimiterMatcher`是用来表示，当前这个限制器将会对哪些`Host`起作用。

```
LimiterMatcher: goreq.LimiterMatcher{
	Glob: "*.taobao.com",
	Regexp: "(.*?).taobao.com,
},
```

`Glob`和`Regexp`只能选择其一。若同时设置，将会使用`Glob`。

#### WithFilterLimiter

此限制器用于过滤请求。

```go
func WithFilterLimiter(noneMatchAllow bool, opts ...*FilterLimiterOpinion) Middleware
```

`noneMatchAllow`表示经过此过滤器，未名中Matcher的请求是否允许放行。

```go
c := goreq.NewClient(goreq.WithFilterLimiter(true, &goreq.FilterLimiterOpinion{
   LimiterMatcher: goreq.LimiterMatcher{
      Glob: "*.taobao.com",
   },
   Allow: false,
}))
```

#### DelayLimiterOpinion

此限制器用于控制请求间的延时。

```go
func WithDelayLimiter(eachSite bool, opts ...*DelayLimiterOpinion) Middleware
```

`eachSite`表示是否为不同的Matcher分别计时。

```go
c = goreq.NewClient(goreq.WithDelayLimiter(false, &goreq.DelayLimiterOpinion{
   LimiterMatcher: goreq.LimiterMatcher{
      Glob: "*",
   },
   Delay: 5 * time.Second,
   // RandomDelay: 5 * time.Second,
}))
```

使用`RandomDelay`将从0到`RandomDelay`之间随机取一个时间延时。

#### RateLimiterOpinion

此限制器用于控制发送请求的速率。

```go
func WithRateLimiter(eachSite bool, opts ...*RateLimiterOpinion) Middleware
```

`eachSite`表示是否为不同的Matcher分别控制。

```go
c = goreq.NewClient(goreq.WithRateLimiter(false, &goreq.RateLimiterOpinion{
   LimiterMatcher: goreq.LimiterMatcher{
      Glob: "*",
   },
   Rate: 2,
}))
```

#### ParallelismLimiterOpinion

此限制器用于控制并发数量。

```go
func WithParallelismLimiter(eachSite bool, opts ...*ParallelismLimiterOpinion) Middleware
```

`eachSite`表示是否为不同的Matcher分别控制。

```go
c = goreq.NewClient(goreq.WithParallelismLimiter(false, &goreq.ParallelismLimiterOpinion{
   LimiterMatcher: goreq.LimiterMatcher{
      Glob: "*",
   },
   Parallelism: 2,
}))
```

### WithCookie

```go
func WithCookie(urlAddr string, cookies ...*http.Cookie) Middleware
```

向`cookie jar`添加`cookie`。

```go
c := goreq.NewClient(goreq.WithCookie("https://example.com", &http.Cookie{
   Name:  "token",
   Value: "admin",
}))
```

### WithCache

缓存响应。

```go
func WithCache(ca *cache.Cache) Middleware
```

```go
c := goreq.NewClient(goreq.WithCache(cache.New(1*time.Hour, 10*time.Minute)))
```

WithCache的缓存是基于对Request的Hash。具体可以查看`utils.go`的`GetRequestHash`。

### WithRetry

自动重试。

```go
func WithRetry(maxTimes int, isRespOk func(*Response) bool) Middleware
```

* **maxTimes** 最大重试次数。
* isRespOk 用于判断请求是否成功的函数。若为`nil`则只检查`Response.Err`是否为`nil`。

### WithProxy

自动配置代理。

```go
func WithProxy(p ...string) Middleware
```

可以传入一或多个代理URL（如：http://localhost:1080）。若传入多个时将在每次请求随机选取一个。

若不传入参数，会自动使用`all_proxy`，`http_proxy`，`https_proxy`的环境变量。

### WithRefererFiller

自动把Referer头部填写为当前请求地址的根地址。用于处理防盗链。

```go
func WithRefererFiller() Middleware
```

### WithRandomUA

当请求头部UA为空时，随机添加一个UA。

```go
func WithRandomUA() Middleware
```

备选UA参考`mw.go`末尾。

## 开发中间件

### 中间件是什么

```go
type Middleware func(*Client, Handler) Handler
type Handler func(*Request) *Response
```

中间件，即`Middleware`是给Client添加的一个处理函数（`Handler`）。`Handler`则是负责将`Request`转化为`Response`的函数。

例如使用`WithRefererFiller`中间件为例。

```go
func WithRefererFiller() Middleware {
   return func(x *Client, h Handler) Handler {
      return func(req *Request) *Response {
         if req.Header.Get("Referer") == "" {
            req.AddHeader("Referer", req.URL.Scheme+"://"+req.URL.Host)
         }
         res := h(req)
         return res
      }
   }
}
```

最外层函数一般留作传入配置参数，比如`WithProxy`中间件就需要传入代理的配置信息。此函数返回的就是`Middleware`，即为中间件。

中间件`Middleware`将被交给`Client`调用，`Client`将传入的中间件组合嵌套，并在执行`Do()`时调用组合好的中间件队列。

### 执行顺序

```go
func main() {
   c := goreq.NewClient(
      func(x *goreq.Client, h goreq.Handler) goreq.Handler {
         return func(req *goreq.Request) *goreq.Response {
            fmt.Println("middleware 1")
            return h(req)
         }
      },
      func(x *goreq.Client, h goreq.Handler) goreq.Handler {
         return func(req *goreq.Request) *goreq.Response {
            fmt.Println("middleware 2")
            return h(req)
         }
      },
   )
   goreq.Get("https://httpbin.org/get").SetClient(c).Do()
}
```

输出：

```
middleware 2
middleware 1
```

后添加的中间件会被先执行。