package s3

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
)

type S3Client struct {
	host   string
	bucket string

	httpClient *http.Client
}

func New(host, bucket string) (*S3Client, error) {
	if host == "" {
		return nil, errors.New("invalid host")
	}
	if bucket == "" {
		return nil, errors.New("invalid bucket name")
	}
	c := S3Client{
		host:       host,
		bucket:     bucket,
		httpClient: &http.Client{},
	}
	return &c, nil
}

type object struct {
	Name   string `json:"name"`
	Scheme string `json:"scheme"`
	Domain string `json:"domain"`
	Path   string `json:"path"`
}

func (c *S3Client) Put(ctx context.Context, name string, reader io.Reader) (*object, error) {
	if c == nil {
		return nil, errors.New("nil client")
	}
	if reader == nil {
		return nil, errors.New("nil data")
	}

	body := bytes.NewBuffer(nil)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("", filepath.Base(name))
	if err != nil {
		return nil, err
	}
	io.Copy(part, reader)
	writer.Close()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("http://%s/paas/s3/object/%s", c.host, c.bucket),
		body,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upload failed, error=%s", data)
	}
	var r struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    object `json:"data"`
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	if r.Code != 0 {
		return nil, fmt.Errorf("upload failed with error %s", r.Message)
	}
	return &r.Data, nil
}
