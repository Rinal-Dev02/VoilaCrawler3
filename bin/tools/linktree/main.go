// Linktree

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
	pbItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	pbProxy "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
)

type _Crawler struct {
	httpClient http.Client

	categoryPathMatcher     *regexp.Regexp
	categoryJsonPathMatcher *regexp.Regexp
	productGroupPathMatcher *regexp.Regexp
	productPathMatcher      *regexp.Regexp
	logger                  glog.Log
}

func (_ *_Crawler) New(_ *cli.Context, client http.Client, logger glog.Log) (crawler.Crawler, error) {
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
	return "__linktree__"
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
	options.Reliability = pbProxy.ProxyReliability_ReliabilityRealtime
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

	item := pbItem.Linktree_Item{
		Profile: &pbItem.Linktree_Item_Profile{
			Name:   strings.TrimSpace(dom.Find(`div>h1[id]`).Text()),
			Avatar: dom.Find(`img[data-testid="ProfileImage"]`).AttrOr("src", ""),
		},
	}
	sel := dom.Find(`div>div[data-testid="StyledContainer"]>a`)
	for i := range sel.Nodes {
		node := sel.Eq(i)
		title := strings.TrimSpace(node.Text())
		href := node.AttrOr("href", "")
		u, _ := url.Parse(href)
		if u == nil || href == "" {
			c.logger.Errorf(`loaded invalid link "%s"`, href)
		}
		link := pbItem.Linktree_Item_Link{
			Title: title,
			Url:   u.String(),
		}
		item.Links = append(item.Links, &link)
	}
	return yield(ctx, &item)
}

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		"https://linktr.ee/bytrendypep/",
		"https://linktr.ee/kellyydoan",
		"https://linktr.ee/Clairebridges",
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
	cli.NewApp(&_Crawler{}).Run(os.Args)
}
