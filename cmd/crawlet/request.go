package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
)

// NewRequest
func NewRequest(r *pbCrawl.Command_Request) (*http.Request, error) {
	if r == nil {
		return nil, errors.New("invalid request command")
	}

	var body io.Reader
	if r.Method != http.MethodGet && r.GetBody() != "" {
		body = bytes.NewReader([]byte(r.GetBody()))
	}

	req, err := http.NewRequest(r.Method, r.Url, body)
	if err != nil {
		return nil, err
	}
	for key, val := range r.CustomHeaders {
		req.Header.Set(key, val)
	}
	for _, cookie := range r.CustomCookies {
		if cookie.Path != "" && !strings.HasPrefix(req.URL.Path, cookie.Path) {
			continue
		}
		v := fmt.Sprintf("%s=%s", cookie.Name, cookie.Value)
		if c := req.Header.Get("Cookie"); c != "" {
			req.Header.Set("Cookie", c+"; "+v)
		} else {
			req.Header.Set("Cookie", v)
		}
	}
	if req.Header.Get("Referer") == "" && r.GetParent().GetUrl() != "" {
		req.Header.Set("Referer", r.GetParent().GetUrl())
	}

	return req, nil
}
