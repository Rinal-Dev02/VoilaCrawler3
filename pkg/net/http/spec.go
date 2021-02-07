package http

import (
	"context"
	"net/http"
	"net/url"
)

const (
	MethodGet   = "GET"
	MethodHead  = "HEAD"
	MethodPost  = "POST"
	MethodPut   = "PUT"
	MethodPatch = "PATCH" // RFC 5789
	//MethodDelete  = "DELETE"
	//MethodConnect = "CONNECT"
	MethodOptions = "OPTIONS"
	//MethodTrace   = "TRACE"
)

type (
	Request  = http.Request
	Response = http.Response
	Cookie   = http.Cookie
	Header   = http.Header
	SameSite = http.SameSite
)

var (
	NewRequest = http.NewRequest
)

type ProxyLevel int

const (
	// using sharing proxys, as if ProxyCrawl's Backconnect
	ProxyLevelSharing ProxyLevel = iota
	// currently using sharing proxys, as if ProxyCrawl's Backconnect
	ProxyLevelSharingButReliable
	// as if ProxyCrawl's Crawl API
	ProxyLevelReliable
	// as if ProxyCrawl's Crawl API
	ProxyLevelHA
)

type Options struct {
	// EnableProxy
	EnableProxy bool
	// EnableHeadless
	EnableHeadless bool

	// EnableSessionInit
	EnableSessionInit bool
	// KeepSession
	KeepSession bool

	// ProxyLevel proxies will try from low level to high level
	ProxyLevel ProxyLevel
}

type Client interface {
	Jar() CookieJar
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	DoWithOptions(ctx context.Context, req *http.Request, opts Options) (*http.Response, error)
}

// CookieJar
type CookieJar interface {
	// Jar returns the standard jar
	Jar() http.CookieJar

	// SetCookies handles the receipt of the cookies in a reply for the
	// given URL.  It may or may not choose to save the cookies, depending
	// on the jar's policy and implementation.
	SetCookies(ctx context.Context, u *url.URL, cookies []*Cookie) error

	// Cookies returns the cookies to send in a request for the given URL.
	// It is up to the implementation to honor the standard cookie use
	// restrictions such as in RFC 6265.
	Cookies(ctx context.Context, u *url.URL) ([]*Cookie, error)
}
