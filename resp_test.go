package req

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"unicode/utf8"
)

var jsonResp *Response
var htmlResp *Response

func init() {
	jsonResp = Get("https://httpbin.org/get").Do()
	if jsonResp.Err != nil {
		panic(jsonResp.Err)
	}

	htmlResp = Get("https://httpbin.org/").Do()
	if htmlResp.Err != nil {
		panic(htmlResp.Err)
	}
}

func TestResponse_DecodeAndParse(t *testing.T) {
	resp := Get("http://stock.10jqka.com.cn/zhuanti/hlw_list/").Do()
	assert.False(t, utf8.Valid(resp.NoDecodeBody))
	assert.True(t, utf8.Valid(resp.Body))
}

func TestResponse_HTML(t *testing.T) {
	assert.True(t, htmlResp.IsHTML())
	h, _ := htmlResp.HTML()
	t.Log(h.Find("title").Text())
	assert.NotEqual(t, h.Find("title").Length(), 0)
}

func TestResponse_JSON(t *testing.T) {
	assert.True(t, jsonResp.IsJSON())
	t.Log(jsonResp.Text)
	j, _ := jsonResp.JSON()
	assert.Equal(t, j.Get("url").String(), "https://httpbin.org/get")
}
