package http

import "net/http"

type Request struct {
	http.Request
}

type Response struct {
	http.Response
}
