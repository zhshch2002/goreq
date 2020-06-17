package goreq

import (
	"fmt"
	"strings"
)

func ModifyLink(url string) string {
	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		return url
	}
	if strings.HasPrefix(url, ":") {
		return fmt.Sprintf("http://127.0.0.1%s", url)
	}
	if strings.HasPrefix(url, "/") {
		return fmt.Sprintf("http://127.0.0.1%s", url)
	}
	return fmt.Sprintf("http://%s", url)
}
