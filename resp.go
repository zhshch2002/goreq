package goreq

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"github.com/PuerkitoBio/goquery"
	"github.com/saintfish/chardet"
	"github.com/tidwall/gjson"
	"golang.org/x/net/html/charset"
	"gopkg.in/xmlpath.v2"
	"io/ioutil"
	"net/http"
	"strings"
)

// Response is a object of HTTP response
type Response struct {
	*http.Response
	Body           []byte
	NotDecodedBody []byte
	Text           string
	Req            *Request
	Err            error
}

func (s *Response) Resp() (*Response, error) {
	return s, s.Err
}

func (s *Response) Txt() (string, error) {
	return s.Text, s.Err
}

func (s *Response) HTML() (*goquery.Document, error) {
	if s.Err != nil {
		return nil, s.Err
	}
	return goquery.NewDocumentFromReader(bytes.NewReader(s.Body))
}

func (s *Response) XML() (*xmlpath.Node, error) {
	if s.Err != nil {
		return nil, s.Err
	}
	return xmlpath.Parse(bytes.NewReader(s.Body))
}

func (s *Response) JSON() (gjson.Result, error) {
	return gjson.Parse(s.Text), s.Err
}

func (s *Response) Error() error {
	return s.Err
}

// DecodeAndParas decodes the body to text and try to parse it to html or json.
func (s *Response) DecodeAndParse() error {
	if s.Err != nil {
		return s.Err
	}
	if len(s.Body) == 0 {
		return nil
	}
	s.NotDecodedBody = s.Body
	contentType := strings.ToLower(s.Header.Get("Content-Type"))
	if strings.Contains(contentType, "text/") ||
		strings.Contains(contentType, "/json") {
		if !strings.Contains(contentType, "charset") {
			if s.Req.RespEncode != "" {
				contentType += "; charset=" + s.Req.RespEncode
			} else {
				r, err := chardet.NewTextDetector().DetectBest(s.Body)
				if err != nil {
					return err
				}
				contentType += "; charset=" + r.Charset
			}
		}
		if strings.Contains(contentType, "utf-8") || strings.Contains(contentType, "utf8") {
			s.Text = string(s.Body)
		} else {
			tmpBody, err := encodeBytes(s.Body, contentType)
			if err != nil {
				return err
			}
			s.Body = tmpBody
			s.Text = string(s.Body)
		}
	}
	return nil
}

func (s *Response) IsHTML() bool {
	contentType := strings.ToLower(s.Header.Get("Content-Type"))
	return strings.Contains(contentType, "/html")
}

func (s *Response) IsJSON() bool {
	contentType := strings.ToLower(s.Header.Get("Content-Type"))
	return strings.Contains(contentType, "/json")
}

func (s *Response) BindJSON(i interface{}) error {
	if s.Err != nil {
		return s.Err
	}
	return json.Unmarshal(s.Body, i)
}

func (s *Response) BindXML(i interface{}) error {
	if s.Err != nil {
		return s.Err
	}
	return xml.Unmarshal(s.Body, i)
}

func encodeBytes(b []byte, contentType string) ([]byte, error) {
	r, err := charset.NewReader(bytes.NewReader(b), contentType)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(r)
}
