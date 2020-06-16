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
