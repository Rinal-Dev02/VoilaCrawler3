package http

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

func init() {
	http.Request
}

type Request struct {
	ctx    context.Context
	Method string
	URL    *url.URL
	Header http.Header
	Body   io.Reader
}

func NewRequest(method string, url string, body io.Reader) (*Request, error) {
}
