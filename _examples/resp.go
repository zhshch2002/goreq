package main

import (
	"fmt"
	"github.com/zhshch2002/goreq"
)

func main() {
	resp := req.Get("https://example.com/").Do()
	fmt.Println(resp.Text, resp.Err) // Get the decode text,same as `text,err:=resp.Txt()`
	j, err := resp.JSON()
	fmt.Println(resp.IsJSON(), j, err)
	h, err := resp.HTML()
	fmt.Println(resp.IsHTML(), h, err)
	x, err := resp.XML()
	fmt.Println(x, err)
	var data struct {
		Url string `json:"url" xml:"url"`
	}
	err = resp.BindJSON(&data)
	fmt.Println(data, err)
	err = resp.BindXML(&data)
	fmt.Println(data, err)
}
