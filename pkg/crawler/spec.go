package crawler

import (
	"context"
	"errors"
	rhttp "net/http"
	"net/url"

	"github.com/voiladev/go-crawler/pkg/net/http"
	pbHttp "github.com/voiladev/go-crawler/protoc-gen-go/chameleon/api/http"
	pbCrawl "github.com/voiladev/go-crawler/protoc-gen-go/chameleon/smelter/v1/crawl"
	"github.com/voiladev/go-crawler/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
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
	MustHeader rhttp.Header `json:"mustHeader"`

	// MustCookies specify the musted cookies
	MustCookies []*rhttp.Cookie `json:"mustCookies"`

	// ProxyReliability
	Reliability proxy.ProxyReliability
}

func (opts *CrawlOptions) Unmarshal(ret interface{}) error {
	if opts == nil || ret == nil {
		return nil
	}

	switch v := ret.(type) {
	case *pbCrawl.CrawlerOptions:
		v.EnableHeadless = opts.EnableHeadless
		v.EnableSessionInit = opts.EnableSessionInit
		v.KeepSession = opts.KeepSession
		v.SessoinTtl = int64(opts.SessionTtl)
		v.DisableCookieJar = opts.DisableCookieJar
		v.DisableRedirect = opts.DisableRedirect
		v.Reliability = opts.Reliability
		v.Headers = map[string]string{}
		for k := range opts.MustHeader {
			v.Headers[k] = opts.MustHeader.Get(k)
		}
		for _, c := range opts.MustCookies {
			v.Cookies = append(v.Cookies, &pbHttp.Cookie{
				Name:   c.Name,
				Value:  c.Value,
				Domain: c.Domain,
				Path:   c.Path,
			})
		}
	default:
		return errors.New("unsupported type")
	}
	return nil
}

func NewCrawlOptions() *CrawlOptions {
	return &CrawlOptions{MustHeader: rhttp.Header{}}
}

// HealthChecker used to test if website struct changed
type HealthChecker interface {
	// NewTestRequest generate a test request
	NewTestRequest(ctx context.Context) []*rhttp.Request

	// CheckTestResponse used to check whether website struct changed
	CheckTestResponse(ctx context.Context, resp *rhttp.Response) error
}

// Crawler
type Crawler interface {
	HealthChecker

	// ID returns crawler unique id, this commonly should be the hosted id of this site called store Id.
	ID() string

	// Version returns the version of current this crawler, which should be an active number.
	Version() int32

	// SupportedTypes returns the yield item type defined in package chameleon.smelter.v1.crawl.item
	// default type is Product
	// SupportedTypes() []proto.Message

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
	Parse(ctx context.Context, resp *rhttp.Response, yield func(context.Context, interface{}) error) error
}

// MustImplementCrawler
type MustImplementCrawler struct{}

// CanonicalUrl
func (c *MustImplementCrawler) CanonicalUrl(rawurl string) (string, error) {
	return "", nil
}

type New func(client http.Client, logger glog.Log) (Crawler, error)
