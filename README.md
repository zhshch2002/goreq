# Req
An elegant and concise Go HTTP request library.
一个优雅并简洁的Go HTTP请求库。

```shell script
go get -u github.com/zhshch2002/goreq
```

```go
package main

import (
	"fmt"
	"github.com/zhshch2002/goreq"
)

func main() {
	res := req.Do(req.Post("https://httpbin.org/post?hello=world").
		SetFormBody(map[string]string{
			"aaa": "123",
		}).AddParams(map[string]string{
		"bbb": "312",
	}).AddHeader("Req-Client", "GoReq"))
	fmt.Println(res.Text)
	j, err := res.JSON()
	fmt.Println(err)
	fmt.Println(j.Get("form"))
}
```

## Feature
* 自动解码
* 便捷代理设置
* HTML、JSON、XML解析
* 链式配置请求
* 缓存
* 中间件
    * 失败重试
    * 日志
    * 随机UA
    * 填充Referer
### TODO
* 上传下载文件
* 格式化打印