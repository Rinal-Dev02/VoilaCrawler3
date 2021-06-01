package http

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Response struct {
	*http.Response

	body   []byte
	isRead bool
}

func NewResponse(resp *http.Response) (*Response, error) {
	if resp == nil {
		return nil, errors.New("invalid response")
	}

	r := Response{Response: resp}
	if resp.Body != nil {
		var err error
		if r.body, err = io.ReadAll(resp.Body); err != nil {
			return nil, err
		}
		resp.Body = NewReader(r.body)
		r.isRead = true
	}
	return &r, nil
}

func (res *Response) CurrentUrl() *url.URL {
	if res == nil {
		return nil
	}
	return res.Request.URL
}

func (res *Response) RawUrl() *url.URL {
	if res == nil {
		return nil
	}
	for p := res.Request; ; p = p.Response.Request {
		if p.Response == nil {
			return p.URL
		}
	}
}

// RawBody return the raw response data
func (res *Response) RawBody() ([]byte, error) {
	if res == nil {
		return nil, nil
	}

	// TODO: add support for raw data decoding
	if !res.isRead && res.Response.Body != nil {
		var err error
		if res.body, err = io.ReadAll(res.Response.Body); err != nil {
			return nil, err
		}
		res.isRead = true
		res.Response.Body.Close()
	}
	return res.body, nil
}

// Selector returns a goquery based css selector for html node select
func (res *Response) Selector() (*goquery.Document, error) {
	if res == nil {
		return nil, nil
	}

	ct := res.Header.Get("Content-Type")
	if strings.Contains(ct, "text/html") ||
		strings.Contains(ct, "application/xhtml+xml") ||
		strings.Contains(ct, "application/xml") {

		data, err := res.RawBody()
		if err != nil {
			return nil, err
		}
		if len(data) > 0 {
			return goquery.NewDocumentFromReader(bytes.NewReader(data))
		}
	}
	return &goquery.Document{}, nil
}
