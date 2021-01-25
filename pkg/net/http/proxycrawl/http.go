package proxycrawl

import (
	"context"
	"errors"
	"io/ioutil"
	rhttp "net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/voiladev/VoilaCrawl/pkg/net/http"
)

const (
	gatewayAddr = "https://api.proxycrawl.com/"
)

type clientOptions struct {
	APIToken string
	JSToken  string
}

type ClientOptionFunc func(opts *clientOptions) error

func WithAPITokenOption(token string) ClientOptionFunc {
	return func(opts *clientOptions) error {
		if opts == nil {
			return nil
		}
		opts.APIToken = token

		return nil
	}
}

func WithJSTokenOption(token string) ClientOptionFunc {
	return func(opts *clientOptions) error {
		if opts == nil {
			return nil
		}
		opts.JSToken = token
		return nil
	}
}

// proxyCrawlClient
type proxyCrawlClient struct {
	httpClient *rhttp.Client
	options    clientOptions
}

func NewProxyCrawlClient(opts ...ClientOptionFunc) (http.Client, error) {
	client := proxyCrawlClient{
		httpClient: &rhttp.Client{},
	}

	for _, optFunc := range opts {
		if optFunc == nil {
			continue
		}
		optFunc(&client.options)
	}
	if client.options.APIToken == "" && client.options.JSToken == "" {
		return nil, errors.New("missing proxy api access token")
	}

	return &client, nil
}

func (c *proxyCrawlClient) Do(ctx context.Context, r *http.Request) (*http.Response, error) {
	if c == nil {
		return nil, nil
	}
	return c.DoWithOptions(ctx, r, http.Options{})
}

func (c *proxyCrawlClient) DoWithOptions(ctx context.Context, r *http.Request, opts http.Options) (*http.Response, error) {
	if c == nil || r == nil {
		return nil, nil
	}

	var (
		err error
		req *http.Request = r
	)
	if opts.EnableProxy {
		u, _ := url.Parse(gatewayAddr)

		// Set params
		vals := u.Query()
		if opts.EnableHeadless {
			vals.Set("token", c.options.JSToken)
		} else {
			vals.Set("token", c.options.APIToken)
		}
		vals.Set("url", r.URL.String())

		if len(r.Header) > 0 {
			var header string
			for k := range r.Header {
				if k == "Cookie" {
					vals.Set("cookies", r.Header.Get(k))
				} else {
					if header == "" {
						header = k + ":" + r.Header.Get(k)
					} else {
						header = header + "|" + k + ":" + r.Header.Get(k)
					}
				}
			}
			vals.Set("request_headers", header)
		}
		vals.Set("get_headers", "true")
		vals.Set("get_cookies", "true")
		// Set this param for we develop almost all crawler on desktop
		vals.Set("device", "desktop")

		u.RawQuery = vals.Encode()

		if req, err = http.NewRequest(http.MethodGet, u.String(), r.Body); err != nil {
			return nil, err
		}
	}
	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body = http.NewReader(data)

	if opts.EnableProxy {
		res := http.Response{Header: rhttp.Header{}}
		for key := range resp.Header {
			key := strings.ToLower(key)
			if key == "original_status" {
				code, _ := strconv.ParseInt(resp.Header.Get("original_status"), 10, 32)
				res.StatusCode = int(code)
			} else if strings.HasPrefix(key, "original_") {
				// TODO: there may exists header with underline
				realKey := strings.Replace(strings.TrimPrefix(key, "original_"), "_", "-", -1)
				res.Header.Set(realKey, resp.Header.Get(key))
			}
		}
		res.Body = resp.Body
		res.Request = r
		resp = &res
	}
	return resp, nil
}
