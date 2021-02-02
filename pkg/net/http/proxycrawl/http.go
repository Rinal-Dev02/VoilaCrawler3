package proxycrawl

import (
	"context"
	"crypto/tls"
	"errors"
	"io/ioutil"
	rhttp "net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	// "golang.org/x/time/rate"
	"github.com/voiladev/VoilaCrawl/pkg/net/http"
	"github.com/voiladev/go-framework/glog"
	"golang.org/x/net/publicsuffix"
)

const (
	gatewayAddr = "https://api.proxycrawl.com/"
)

type Options struct {
	APIToken string
	JSToken  string
}

// proxyCrawlClient
type proxyCrawlClient struct {
	httpClient      *rhttp.Client
	httpProxyClient *rhttp.Client
	options         Options
	logger          glog.Log
}

// NewProxyCrawlClient returns a http client which support native http.Client and
// supports ProxyCrawl Backconnect proxy, Crawling API.
func NewProxyCrawlClient(logger glog.Log, opts Options) (http.Client, error) {
	if opts.APIToken == "" && opts.JSToken == "" {
		return nil, errors.New("missing proxy api access token")
	}

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, err
	}
	client := proxyCrawlClient{
		httpClient: &rhttp.Client{
			Transport: &rhttp.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:     time.Second * 30,
			},
			Jar: jar,
		},
		httpProxyClient: &rhttp.Client{
			Transport: &rhttp.Transport{
				// MaxIdleConns:      100,
				// DisableKeepAlives: true,
				IdleConnTimeout: time.Second,
				TLSNextProto:    map[string]func(string, *tls.Conn) rhttp.RoundTripper{}, // disable http2
				Proxy:           rhttp.ProxyURL(&url.URL{Host: "proxy.proxycrawl.com:9000"}),
			},
			Jar: jar,
		},
		options: opts,
		logger:  logger,
	}
	return &client, nil
}

func (c *proxyCrawlClient) Do(ctx context.Context, r *http.Request) (*http.Response, error) {
	if c == nil {
		return nil, nil
	}
	return c.DoWithOptions(ctx, r, http.Options{})
}

// DoWithOptions
// If proxy is enabled, the client will try backconnect proxy for at most twice.
// If all the two try failed, then this proxy will try crawling api.
// To remember that timeout is controlled by ctx
func (c *proxyCrawlClient) DoWithOptions(ctx context.Context, r *http.Request, opts http.Options) (*http.Response, error) {
	if c == nil || r == nil {
		return nil, nil
	}
	if opts.EnableHeadless {
		opts.EnableProxy = true
	}

	var (
		err  error
		req  *http.Request = r
		resp *http.Response
	)

	if opts.EnableProxy {
		if req.Header.Get("User-Agent") == "" {
			req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.96 Safari/537.36")
		}

		var (
			retryCount = 5
		)
		for i := 0; i < retryCount; i++ {
			creq := req.Clone(ctx)
			if resp, err = c.httpProxyClient.Do(creq); err != nil {
				c.logger.Debugf("do http request failed, error=%s", err)
				time.Sleep(time.Millisecond * 200)
				continue
			} else if resp.StatusCode == http.StatusRequestTimeout ||
				resp.StatusCode == http.StatusInternalServerError ||
				resp.StatusCode == http.StatusServiceUnavailable {
				c.logger.Debugf("do http request with status %v, retry...", resp.StatusCode)
				time.Sleep(time.Millisecond * 200)
				continue
			} else if resp.StatusCode == http.StatusForbidden {
				c.logger.Debugf("do http request with status %v, retry...", resp.StatusCode)
				retryCount = 10
				time.Sleep(time.Millisecond * 5000)
				continue
			}
			break
		}

		if resp == nil {
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
					} else if k == "User-Agent" {
						continue
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
			req = req.WithContext(ctx)

			for i := 0; i < 3; i++ {
				if resp, err = c.httpClient.Do(req); err != nil {
					c.logger.Debugf("access %s failed, error=%s", err)
					time.Sleep(time.Millisecond * 100)
					continue
				} else if resp.StatusCode == http.StatusRequestTimeout ||
					resp.StatusCode == http.StatusInternalServerError ||
					resp.StatusCode == http.StatusServiceUnavailable {
					time.Sleep(time.Millisecond * 100)
					continue
				}
				break
			}
			if err != nil {
				return nil, err
			}

			for key := range resp.Header {
				key := strings.ToLower(key)
				if key == "original_status" {
					code, _ := strconv.ParseInt(resp.Header.Get("original_status"), 10, 32)
					resp.StatusCode = int(code)
					resp.Header.Del(key)
				} else if strings.HasPrefix(key, "original_") {
					// TODO: there may exists header with underline
					realKey := strings.Replace(strings.TrimPrefix(key, "original_"), "_", "-", -1)
					resp.Header.Set(realKey, resp.Header.Get(key))
					resp.Header.Del(key)
				} else if key == "pc_status" {
					resp.Header.Del(key)
				}
			}
			resp.Request = r

			cookies := resp.Cookies()
			if len(cookies) > 0 {
				c.httpClient.Jar.SetCookies(resp.Request.URL, cookies)
			}
		}
	} else {
		retryCount := 2
		for i := 0; i < retryCount; i++ {
			creq := req.Clone(ctx)
			if resp, err = c.httpClient.Do(creq); err != nil {
				c.logger.Debugf("access %s failed, error=%s", err)
				time.Sleep(time.Millisecond * 100)
				continue
			} else if resp.StatusCode == http.StatusRequestTimeout ||
				resp.StatusCode == http.StatusInternalServerError ||
				resp.StatusCode == http.StatusServiceUnavailable {
				time.Sleep(time.Millisecond * 100)
				continue
			} else if resp.StatusCode == http.StatusForbidden {
				c.logger.Debugf("do http request with status %v, retry...", resp.StatusCode)
				time.Sleep(time.Millisecond * 5000)
				retryCount = 5
				continue
			}
			break
		}
		if err != nil {
			return nil, err
		}
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body = http.NewReader(data)

	return resp, nil
}
