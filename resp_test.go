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
	assert.NotEqual(t, 0, h.Find("title").Length())
}

func TestResponse_JSON(t *testing.T) {
	assert.True(t, jsonResp.IsJSON())
	t.Log(jsonResp.Text)
	j, _ := jsonResp.JSON()
	assert.Equal(t, "https://httpbin.org/get", j.Get("url").String())

	var data struct {
		Url string `json:"url"`
	}
	assert.Nil(t, jsonResp.BindJSON(&data))
	assert.Equal(t, "https://httpbin.org/get", data.Url)
}
