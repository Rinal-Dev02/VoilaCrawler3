// SEO this spider is used to do seo info fetch

package main

import (
	"html"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/voiladev/VoilaCrawler/pkg/cli"
	"github.com/voiladev/VoilaCrawler/pkg/context"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	pbProxy "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
)

type _Crawler struct {
	httpClient http.Client

	categoryPathMatcher     *regexp.Regexp
	categoryJsonPathMatcher *regexp.Regexp
	productGroupPathMatcher *regexp.Regexp
	productPathMatcher      *regexp.Regexp
	logger                  glog.Log
}

func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient:              client,
		categoryPathMatcher:     regexp.MustCompile(`^(/[a-z0-9_-]+)?/(women|men)(/[a-z0-9_-]+){1,6}/cat/?$`),
		categoryJsonPathMatcher: regexp.MustCompile(`^/api/product/search/v2/categories/([a-z0-9]+)`),
		productGroupPathMatcher: regexp.MustCompile(`^(/[a-z0-9_-]+)?(/[a-z0-9_-]+){2}/grp/[0-9]+/?$`),
		productPathMatcher:      regexp.MustCompile(`^(/[a-z0-9_-]+)?(/[a-z0-9_-]+){2}/prd/[0-9]+/?$`),
		logger:                  logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "0e7484236e51f67f520ce1ae0a11a6a1"
}

// Version
func (c *_Crawler) Version() int32 {
	return 1
}

// CrawlOptions
func (c *_Crawler) CrawlOptions(u *url.URL) *crawler.CrawlOptions {
	options := crawler.NewCrawlOptions()
	options.EnableHeadless = false
	options.EnableSessionInit = false
	options.Reliability = pbProxy.ProxyReliability_ReliabilityMedium
	options.MustCookies = append(options.MustCookies)

	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*"}
}

func (c *_Crawler) CanonicalUrl(rawurl string) (string, error) {
	return rawurl, nil
}

func textClean(s string) string {
	return strings.TrimSpace(html.UnescapeString(s))
}

func (c *_Crawler) Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	dom, err := resp.Selector()
	if err != nil {
		c.logger.Error(err)
		return err
	}

	// ogType := textClean(dom.Find(`meta[property="og:type"]`).AttrOr("content", ""))
	// if ogType != "" && ogType != "product" {
	// 	// unsupported path means this page may not a product detail page
	// 	return crawler.ErrUnsupportedPath
	// }

	item := pbItem.OpenGraph_Product{
		Site:  &pbItem.OpenGraph_Site{},
		Price: &pbItem.OpenGraph_Price{},
	}
	item.Site.Name = dom.Find(`meta[property="og:site_name"]`).AttrOr("content", "")

	for _, key := range []string{
		`branch:deeplink:product`,
	} {
		v := textClean(dom.Find(key).AttrOr("content", ""))
		if v != "" {
			item.Id = v
			break
		}
	}

	for _, key := range []string{
		`meta[property="og:title"]`,
		`meta[name="twitter:title"]`,
		`meta[name="title"]`,
	} {
		v := textClean(dom.Find(key).AttrOr("content", ""))
		if v != "" {
			item.Title = v
			break
		}
	}
	if item.Title == "" {
		item.Title = textClean(dom.Find(`title`).Text())
	}

	for _, key := range []string{
		`meta[property="og:description"]`,
		`meta[property="description"]`,
		`meta[name="twitter:description"]`,
		`meta[name="description"]`,
	} {
		v := textClean(dom.Find(key).AttrOr("content", ""))
		if v != "" {
			item.Description = v
			break
		}
	}

	for _, key := range []string{
		`meta[property="og:url"]`,
		`meta[property="url"]`,
		`link[rel="canonical"]`,
	} {
		v := dom.Find(key).AttrOr("content", dom.Find(key).AttrOr("href", ""))
		if v != "" {
			item.Url = v
			break
		}
	}

	for _, key := range []string{
		`meta[property="og:image:secure_url"]`,
		`meta[property="og:image"]`,
		`meta[property="image:secure_url"]`,
		`meta[property="image"]`,
	} {
		sel := dom.Find(key)
		for i := range sel.Nodes {
			node := sel.Eq(i)
			if v := node.AttrOr("content", ""); v != "" {
				if strings.HasPrefix(v, "//") {
					v = "https:" + v
				}
				item.Medias = append(item.Medias, media.NewImageMedia("", v, "", "", "", "", false))
			}
		}
		if len(item.Medias) > 0 {
			break
		}
	}

	for _, key := range []string{
		`meta[property="og:price:currency"]`,
		`meta[property="og:product:price:currency"]`,
	} {
		v := strings.ToUpper(dom.Find(key).AttrOr("content", ""))
		if v == "USD" {
			item.Price.Currency = regulation.Currency_USD
		}
	}
	if item.Price.Currency == 0 {
		item.Price.Currency = regulation.Currency_USD
	}

	for _, key := range []string{
		`meta[property="og:price:amount"]`,
		`meta[property="og:product:price:amount"]`,
	} {
		v := dom.Find(key).AttrOr("content", "0")
		if v != "" {
			if vv, _ := strconv.ParsePrice(v); vv > 0 {
				item.Price.Value = int32(vv * 100)
				break
			}
		}
	}
	return yield(ctx, &item)
}

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		"https://us.princesspolly.com/collections/basics/products/madelyn-top-green",
	} {
		req, _ := http.NewRequest(http.MethodGet, u, nil)
		reqs = append(reqs, req)
	}
	return reqs
}

func (c *_Crawler) CheckTestResponse(ctx context.Context, resp *http.Response) error {
	if err := c.Parse(ctx, resp, func(c context.Context, i interface{}) error {
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// main func is the entry of golang program. this will not be used by plugin, just for local spider test.
func main() {
	cli.NewApp(New).Run(os.Args)
}
