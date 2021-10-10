package http

import (
	"context"
	"net/http"
	"net/url"

	pbProxy "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
)

const (
	MethodGet     = "GET"
	MethodHead    = "HEAD"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodOptions = "OPTIONS"
)

var SupportedHttpMethods = []string{MethodGet, MethodPost, MethodPut, MethodHead, MethodOptions}

type (
	Request  = http.Request
	Cookie   = http.Cookie
	Header   = http.Header
	SameSite = http.SameSite
)

var (
	NewRequest            = http.NewRequest
	NewRequestWithContext = http.NewRequestWithContext
)

type Options struct {
	// EnableProxy
	EnableProxy bool
	// EnableHeadless
	EnableHeadless bool
	// JSWaitDuration default 6 seconds
	JsWaitDuration int64

	// EnableSessionInit
	EnableSessionInit bool
	// KeepSession
	KeepSession bool
	// DisableCookieJar
	DisableCookieJar bool

	// DisableRedirect disable http redirect when do http request
	DisableRedirect bool

	// Reliability proxies will try from low level to high level
	Reliability pbProxy.ProxyReliability

	// RequestFilterKeys use to filter the response from multi request of the same url(for headless cached request)
	RequestFilterKeys []string

	// Tags proxy filter by tags
	Tags map[string]string
}

type Client interface {
	Jar() CookieJar
	Do(ctx context.Context, req *Request) (*Response, error)
	DoWithOptions(ctx context.Context, req *Request, opts Options) (*Response, error)
}

// CookieJar
type CookieJar interface {
	// Jar returns the standard jar
	// Jar() http.CookieJar

	// Clear cookies
	Clear(ctx context.Context, u *url.URL) error

	// SetCookies handles the receipt of the cookies in a reply for the
	// given URL.  It may or may not choose to save the cookies, depending
	// on the jar's policy and implementation.
	SetCookies(ctx context.Context, u *url.URL, cookies []*Cookie) error

	// Cookies returns the cookies to send in a request for the given URL.
	// It is up to the implementation to honor the standard cookie use
	// restrictions such as in RFC 6265.
	Cookies(ctx context.Context, u *url.URL) ([]*Cookie, error)
}
