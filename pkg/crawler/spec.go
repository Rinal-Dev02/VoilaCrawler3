package crawler

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/voiladev/go-framework/glog"
)

// CrawlOptions
type CrawlOptions struct {
	// EnableHeadless
	EnableHeadless bool `json:"enableHeadless"`

	// LoginRequired indicates that this website needs login before crawl
	// there must be an login subsystem with manages all the robot accounts
	// and cache the cookies after signin.
	LoginRequired bool `json:"loginRequired"`

	// MustHeader specify the musted http headers
	MustHeader http.Header `json:"mustHeader"`

	// MustCookies specify the musted cookies
	MustCookies []*http.Cookie `json:"mustCookies"`
}

func NewCrawlOptions() *CrawlOptions {
	return &CrawlOptions{MustHeader: http.Header{}}
}

// HealthChecker used to test if website struct changed
type HealthChecker interface {
	// NewRequest generate a test request
	NewTestRequest(ctx context.Context) []*http.Request

	// Check used to check whether website struct changed
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

	// CrawlOptions return crawler action action requirement
	CrawlOptions() *CrawlOptions

	// AllowedDomains returns the domains this crawler supportes
	AllowedDomains() []string

	// IsUrlMatch check whether the supplied url matched the crawler's url set.
	// if matched, the can use crawler to extract info from the response of this url.
	IsUrlMatch(*url.URL) bool

	// Parser used to parse http request parse.
	//   param ctx used to share info between parent and child. and it can set the max ttl for parse job.
	//   param resp represents the http response, with act as a real http response.
	//   param yield use to yield data with can be final data or an other http request
	Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error
}

type New func(logger glog.Log) (Crawler, error)

var ErrNotImplementNewType = errors.New("not implements func New(logger glog.Log) (*crawler.Crawler, error)")
