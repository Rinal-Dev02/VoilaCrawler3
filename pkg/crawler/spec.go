package crawler

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/voiladev/go-crawler/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
)

var (
	// ErrNotSupportedPath
	ErrNotSupportedPath = errors.New("not supporped url path")
)

// CrawlOptions
type CrawlOptions struct {
	// EnableHeadless
	EnableHeadless bool `json:"enableHeadless"`

	// EnableSessionInit init the session with current request url if the session is not inited
	// which will get the full cookies. This mostly simplified the work todo with decrypt one website.
	EnableSessionInit bool `json:"enableSessionInit"`
	// KeepSession keep the session for all the sub requests
	KeepSession bool `json:"keepSession"`
	// SessionTTL if not set, will set session ttl according to last cookie expires
	SessionTtl int32 `json:"sessionTtl"`
	// DisableCookieJar disable cookie save
	DisableCookieJar bool `json:"disableCookieJar"`

	// DisableRedirect
	DisableRedirect bool `json:"disableRedirect"`

	// (TODO) LoginRequired indicates that this website needs login before crawl
	// there must be an login subsystem with manages all the robot accounts
	// and cache the cookies after signin.
	// LoginRequired bool `json:"loginRequired"`

	// MustHeader specify the musted http headers
	MustHeader http.Header `json:"mustHeader"`

	// MustCookies specify the musted cookies
	MustCookies []*http.Cookie `json:"mustCookies"`

	// ProxyReliability
	Reliability proxy.ProxyReliability
}

func NewCrawlOptions() *CrawlOptions {
	return &CrawlOptions{MustHeader: http.Header{}}
}

// HealthChecker used to test if website struct changed
type HealthChecker interface {
	// NewTestRequest generate a test request
	NewTestRequest(ctx context.Context) []*http.Request

	// CheckTestResponse used to check whether website struct changed
	CheckTestResponse(ctx context.Context, resp *http.Response) error
}

// Crawler
type Crawler interface {
	HealthChecker

	// ID returns crawler unique id, which must be the same for all the version.
	ID() string

	// Version returns the version of current this crawler, which should be an active number.
	Version() int32

	// // StoreID returns the store ID this crawler binded to. This id is used to verify the store.
	// // if the store is offline, then all crawleres bined to this store will be offline.
	// StoreID() string

	// CrawlOptions return crawler action requirement
	CrawlOptions(u *url.URL) *CrawlOptions

	// AllowedDomains returns the domains this crawler supportes
	AllowedDomains() []string

	// Deprecated. IsUrlMatch check whether the supplied url matched the crawler's url set.
	// if matched, the can use crawler to extract info from the response of this url.
	// IsUrlMatch(*url.URL) bool

	// CanonicalUrl returns canonical url the proviced url
	CanonicalUrl(rawurl string) (string, error)

	// Parser used to parse http request parse.
	//   param ctx used to share info between parent and child. and it can set the max ttl for parse job.
	//   param resp represents the http response, with act as a real http response.
	//   param yield use to yield data with can be final data or an other http request
	Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error
}

// MustImplementCrawler
type MustImplementCrawler struct{}

// CanonicalUrl
func (c *MustImplementCrawler) CanonicalUrl(rawurl string) (string, error) {
	return "", nil
}

type New func(client http.Client, logger glog.Log) (Crawler, error)

var ErrNotImplementNewType = errors.New("not implements func New(logger glog.Log) (*crawler.Crawler, error)")
