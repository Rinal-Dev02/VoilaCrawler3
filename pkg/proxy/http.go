package proxy

import (
	"context"
	"crypto/tls"
	"errors"
	"io/ioutil"
	rhttp "net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	// "golang.org/x/time/rate"
	"github.com/voiladev/VoilaCrawl/pkg/net/http"
	"github.com/voiladev/go-framework/glog"
)

type __disableHttpRedirect__ struct{}

var disableHttpRedirect __disableHttpRedirect__

const (
	gatewayAddr = "https://api.proxycrawl.com/"
)

type Options struct {
	APIToken string
	JSToken  string
	Proxy    *url.URL
}

// proxyClient
type proxyClient struct {
	jar             http.CookieJar
	httpClient      *rhttp.Client
	httpProxyClient *rhttp.Client
	options         Options
	logger          glog.Log
}

// NewProxyClient returns a http client which support native http.Client and
// supports Proxy Backconnect proxy, Crawling API.
func NewProxyClient(cookieJar http.CookieJar, logger glog.Log, opts Options) (http.Client, error) {
	if opts.APIToken == "" && opts.JSToken == "" {
		return nil, errors.New("missing proxy api access token")
	}
	if opts.Proxy == nil {
		opts.Proxy = &url.URL{Host: "proxy.proxycrawl.com:9000"}
	}

	client := proxyClient{
		jar:     cookieJar,
		options: opts,
		logger:  logger,
	}
	client.httpClient = &rhttp.Client{
		Transport: &rhttp.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     time.Second * 30,
		},
		CheckRedirect: client.checkRedirect,
	}
	client.httpProxyClient = &rhttp.Client{
		Transport: &rhttp.Transport{
			// MaxIdleConns:      100,
			// DisableKeepAlives: true,
			IdleConnTimeout: time.Second,
			TLSNextProto:    map[string]func(string, *tls.Conn) rhttp.RoundTripper{}, // disable http2
			Proxy:           rhttp.ProxyURL(opts.Proxy),
		},
		CheckRedirect: client.checkRedirect,
	}
	return &client, nil
}

func (c *proxyClient) Jar() http.CookieJar {
	if c == nil {
		return nil
	}
	return c.jar
}

func (c *proxyClient) Do(ctx context.Context, r *http.Request) (*http.Response, error) {
	if c == nil {
		return nil, nil
	}
	return c.DoWithOptions(ctx, r, http.Options{})
}

// DoWithOptions
// If proxy is enabled, the client will try backconnect proxy for at most twice.
// If all the two try failed, then this proxy will try crawling api.
// To remember that timeout is controlled by ctx
func (c *proxyClient) DoWithOptions(ctx context.Context, r *http.Request, opts http.Options) (*http.Response, error) {
	if c == nil || r == nil {
		return nil, nil
	}

	if opts.DisableRedirect {
		ctx = context.WithValue(ctx, disableHttpRedirect, true)
	}

	if c.jar != nil {
		if cookies, err := c.jar.Cookies(ctx, r.URL); err != nil {
			c.logger.Warnf("get cookies failed, error=%s", err)
		} else if len(cookies) > 0 {
			for _, c := range cookies {
				r.AddCookie(c)
			}
		} else if opts.EnableSessionInit {
			opts.EnableHeadless = true
		}
	} else if opts.EnableSessionInit || opts.KeepSession {
		return nil, errors.New("not cookie jar set")
	}

	if opts.EnableHeadless {
		opts.EnableProxy = true
		opts.ProxyLevel = http.ProxyLevelReliable
	}

	var tracingId string
	if opts.KeepSession {
		v := ctx.Value("tracing_id")
		if v != nil {
			tracingId = v.(string)
		}
		if tracingId == "" {
			c.logger.Warnf("no tracing_id found in context")
		}
	}

	var (
		err   error
		req   *http.Request = r
		resp  *http.Response
		level = opts.ProxyLevel
	)

	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.96 Safari/537.36")
	}

	if opts.EnableProxy {
		if level <= http.ProxyLevelSharing {
			level += 1

			retryCount := 5
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
		}

		// TODO: add more proxy level support
		if resp == nil {
			u, _ := url.Parse(gatewayAddr)
			// Set params
			vals := u.Query()
			if opts.EnableHeadless {
				vals.Set("token", c.options.JSToken)
				vals.Set("page_wait", "1000")
				vals.Set("ajax_wait", "false")
			} else {
				vals.Set("token", c.options.APIToken)
			}
			if opts.KeepSession {
				vals.Set("proxy_session", tracingId)
			}
			vals.Set("url", r.URL.String())

			if len(r.Header) > 0 {
				var header string
				for k := range r.Header {
					if k == "Cookie" {
						vals.Set("cookies", r.Header.Get(k))
					} else if k == "User-Agent" {
						// ignore user-agent
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
			// vals.Set("proxy_session", "1234567890")
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
					for _, v := range resp.Header.Values(key) {
						resp.Header.Add(realKey, v)
					}
					resp.Header.Del(key)
				} else if key == "pc_status" {
					resp.Header.Del(key)
				}
			}
			resp.Request = r
		}
	} else {
		retryCount := 2
		for i := 0; i < retryCount; i++ {
			creq := req.Clone(ctx)
			c.logger.Debugf("%+v", creq.Header)
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

	if c.jar != nil {
		cookies := resp.Cookies()
		if len(cookies) > 0 {
			c.jar.SetCookies(ctx, resp.Request.URL, cookies)
		}
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body = http.NewReader(data)

	return resp, nil
}

// checkRedirect
func (c *proxyClient) checkRedirect(req *rhttp.Request, via []*rhttp.Request) error {
	if c == nil {
		return nil
	}
	c.logger.Debugf("redirect to %s, header: %+v", req.URL.String(), req.Response.Cookies())

	var (
		ctx             = req.Context()
		disableRedirect bool
	)
	disableRedirectVal := ctx.Value(disableHttpRedirect)
	if disableRedirectVal != nil {
		disableRedirect = disableRedirectVal.(bool)
	}
	if disableRedirect {
		return rhttp.ErrUseLastResponse
	}

	if req.Response != nil && c.jar != nil {
		cookies := req.Response.Cookies()
		if len(cookies) > 0 {
			c.jar.SetCookies(ctx, req.Response.Request.URL, cookies)
		}
	}

	// init cookies
	if cookies, err := c.jar.Cookies(ctx, req.URL); err != nil {
		return err
	} else {
		for _, c := range cookies {
			req.AddCookie(c)
		}
	}
	return nil
}
