package http

import (
	"io"
	"net/http"
)

func init() {
	http.Response
}

type Response struct {
	StatusCode int32
	Header     http.Header
	Body       io.Reader
	Request    *Request
}
