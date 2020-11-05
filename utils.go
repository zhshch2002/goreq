package goreq

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net/url"
	"sort"
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

// GetRequestHash return a hash of url,header,cookie and body data from a request
func GetRequestHash(r *Request) [md5.Size]byte {
	u := r.URL
	UrtStr := u.Scheme + "://"
	if u.User != nil {
		UrtStr += u.User.String() + "@"
	}
	UrtStr += strings.ToLower(u.Host)
	path := u.EscapedPath()
	if path != "" && path[0] != '/' {
		UrtStr += "/"
	}
	UrtStr += path
	if u.RawQuery != "" {
		QueryParam := u.Query()
		var QueryK []string
		for k := range QueryParam {
			QueryK = append(QueryK, k)
		}
		sort.Strings(QueryK)
		var QueryStrList []string
		for _, k := range QueryK {
			val := QueryParam[k]
			sort.Strings(val)
			for _, v := range val {
				QueryStrList = append(QueryStrList, url.QueryEscape(k)+"="+url.QueryEscape(v))
			}
		}
		UrtStr += "?" + strings.Join(QueryStrList, "&")
	}

	Header := r.Header
	var HeaderK []string
	for k := range Header {
		HeaderK = append(HeaderK, k)
	}
	sort.Strings(HeaderK)
	var HeaderStrList []string
	for _, k := range HeaderK {
		val := Header[k]
		sort.Strings(val)
		for _, v := range val {
			HeaderStrList = append(HeaderStrList, url.QueryEscape(k)+"="+url.QueryEscape(v))
		}
	}
	HeaderStr := strings.Join(HeaderStrList, "&")

	var Cookie []string
	for _, i := range r.Cookies() {
		Cookie = append(Cookie, i.Name+"="+i.Value)
	}
	CookieStr := strings.Join(Cookie, "&")

	data := []byte(strings.Join([]string{UrtStr, HeaderStr, CookieStr}, "@#@"))
	if br, err := r.GetBody(); err == nil {
		if b, err := ioutil.ReadAll(br); err == nil {
			data = append(data, b...)
		}
	}
	has := md5.Sum(data)
	return has
}
