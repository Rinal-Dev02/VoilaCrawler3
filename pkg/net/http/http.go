package http

import (
	"context"
	"net/http"
)

const (
	MethodGet     = "GET"
	MethodHead    = "HEAD"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodPatch   = "PATCH" // RFC 5789
	MethodDelete  = "DELETE"
	MethodConnect = "CONNECT"
	MethodOptions = "OPTIONS"
	MethodTrace   = "TRACE"
)

type (
	Request  = http.Request
	Response = http.Response
)

var (
	NewRequest = http.NewRequest
)

type Options struct {
	EnableProxy    bool
	EnableHeadless bool
}

type Client interface {
	Do(context.Context, *http.Request) (*http.Response, error)
	DoWithOptions(ctx context.Context, req *http.Request, opts Options) (*http.Response, error)
}
