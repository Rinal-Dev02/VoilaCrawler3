package http

import (
	"bytes"
	"io"
)

type buffer struct {
	*bytes.Reader
}

func NewReader(data []byte) io.ReadCloser {
	return &buffer{Reader: bytes.NewReader(data)}
}

func (buf *buffer) Close() error {
	return nil
}
