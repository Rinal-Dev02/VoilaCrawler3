package crawler

import "errors"

var (
	// ErrUnsupportedPath
	ErrUnsupportedPath = errors.New("unsupporped url path")
	// ErrUnsupportedTarget
	ErrUnsupportedTarget = errors.New("unsupporped target type")
	// ErrAbort abort this request for by reasons to reduce useless retry. reasons may be 404 and so on.
	ErrAbort = errors.New("abort this request")
)
