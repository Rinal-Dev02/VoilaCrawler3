package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/voiladev/VoilaCrawler/pkg/cli"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	pbMedia "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
)

// _Crawler defined the crawler struct/class for which is not necessory to be exportable
type _Crawler struct {
	crawler.MustImplementCrawler

	// httpClient is the object of an http client
	httpClient          http.Client
	categoryPathMatcher *regexp.Regexp
	productPathMatcher  *regexp.Regexp
	// logger is the log tool
	logger glog.Log
}

// New returns an object of interface crawler.Crawler.
// this is the entry of the spider plugin. the plugin manager will call this func to init the plugin.
// view pkg/crawler/spec.go to know more about the interface `Crawler`
func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient: client,
		// this regular used to match category page url path
		categoryPathMatcher: regexp.MustCompile(`^/en/categories([/A-Za-z0-9_-]+)$`),
		productPathMatcher:  regexp.MustCompile(`^/en/categories([/A-Za-z0-9_-]+).html$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "2d710c1e01e640878d69a808d7e4348c"
}

// Version
func (c *_Crawler) Version() int32 {
	// every update of this spider should update this version number
	return 1
}

// CrawlOptions returns the options of this crawler.
// These options tells the spider controller how to do http requests.
// And defined the public headers/cookies.
// for the means of every options please see the definition.
func (c *_Crawler) CrawlOptions(u *url.URL) *crawler.CrawlOptions {
	opts := &crawler.CrawlOptions{
		EnableHeadless: false,
		// use js api to init session for the first request of the crawl
		EnableSessionInit: true,
		Reliability:       proxy.ProxyReliability_ReliabilityDefault,
	}

	// opts.MustCookies = append(opts.MustCookies,
	// 	&http.Cookie{Name: "GlobalE_Data", Value: `{"countryISO":"US","cultureCode":"en-US","currencyCode":"USD","apiVersion":"2.1.4"}`, Path: "/"},
	// 	//&http.Cookie{Name: "_dy_geo", Value: "US.NA.US_DC.US_DC_Washington", Path: "/"},
	// )

	return opts
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.maje.com"}
}

// CanonicalUrl
func (c *_Crawler) CanonicalUrl(rawurl string) (string, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	if c.productPathMatcher.MatchString(u.Path) {
		u.RawQuery = ""
		return u.String(), nil
	}
	return rawurl, nil
}

// Parse is the entry to run the spider.
// ctx is the context of this run. if may contains the shared values in it.
//   you can alse set some value by context.WithValue().
//   but, to be sure that, the key must be string type, and the value must stringable,
//   as string,int,int32 and so on.
// resp is the http response, with contains the response data from target url.
// yield is a callback to emit sub request, or the crawled target object.
//   if you got an sub url, then you can use http.NewRequest to build a new request
//   and emit it to spider controller for schedule. the ctx can be used to share the
//   values between current response and next response.
//   if you got an product item, then you can just emit it.
// returns error when there are any errors happened.
func (c *_Crawler) Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	p := strings.TrimSuffix(resp.Request.URL.Path, "/")
	if p == "/en/homepage" {
		return c.parseCategories(ctx, resp, yield)
	}

	if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yield)
	} else if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	}

	return fmt.Errorf("unsupported url %s", resp.Request.URL.String())
}

// nextIndex used to get the index from the shared data.
// item.index is a const key for item index.
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")))
}

func (c *_Crawler) parseCategories(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	sel := dom.Find(`.menuMainMaje>ul>li`)

	for i := range sel.Nodes {
		node := sel.Eq(i)
		cateName := strings.TrimSpace(node.Find(`a>span`).First().Text())

		if cateName == "" {
			continue
		}

		nctx := context.WithValue(ctx, "Category", cateName)

		subSel := node.Find(`.subMain .titleMain.clearfix>li`)

		for k := range subSel.Nodes {
			subNode2 := subSel.Eq(k)

			subcat2 := strings.TrimSpace(subNode2.Find(`.subMain-title`).First().Text())

			nnctx := context.WithValue(nctx, "SubCategory", subcat2)
			subNode2list := subNode2.Find(`.subMenu-level2 .column`)
			for j := range subNode2list.Nodes {
				subNode := subNode2list.Eq(j)
				subcat3 := strings.TrimSpace(subNode.Find(`a>span`).First().Text())
				if subcat3 == "" {
					continue
				}

				href := subNode.Find(`a`).AttrOr("href", "")
				if href == "" {
					continue
				}

				u, err := url.Parse(href)
				if err != nil {
					c.logger.Error("parse url %s failed", href)
					continue
				}

				subCateName := strings.TrimSpace(subNode.Text())

				if c.categoryPathMatcher.MatchString(u.Path) {
					nnnctx := context.WithValue(nnctx, "SubCategory2", subCateName)
					req, _ := http.NewRequest(http.MethodGet, href, nil)
					if err := yield(nnnctx, req); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// parseCategoryProducts parse api url from web page url
func (c *_Crawler) parseCategoryProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.logger.Debug(err)
		return err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}
	lastIndex := nextIndex(ctx)

	sel := doc.Find(`.titleProduct`)

	for i := range sel.Nodes {
		node := sel.Eq(i)
		if href, _ := node.Find(`a`).Attr("href"); href != "" {
			req, err := http.NewRequest(http.MethodGet, href, nil)
			if err != nil {
				c.logger.Error(err)
				continue
			}
			lastIndex += 1
			nctx := context.WithValue(ctx, "item.index", lastIndex)
			if err := yield(nctx, req); err != nil {
				return err
			}
		}

	}

	nextUrl := doc.Find(`link[rel="next"]`).AttrOr("href", "")
	if nextUrl == "" {
		return nil
	}

	totalCount, _ := strconv.ParsePrice(doc.Find(`.count`).Text())

	if lastIndex >= (int)(totalCount) {
		return nil
	}

	req, _ := http.NewRequest(http.MethodGet, nextUrl, nil)
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
}

func TrimSpaceNewlineInString(s []byte) []byte {
	re := regexp.MustCompile(`\n`)
	resp := re.ReplaceAll(s, []byte(" "))
	resp = bytes.ReplaceAll(resp, []byte("\\n"), []byte(""))
	resp = bytes.ReplaceAll(resp, []byte("\r"), []byte(""))
	resp = bytes.ReplaceAll(resp, []byte("\t"), []byte(""))
	resp = bytes.ReplaceAll(resp, []byte("&lt;"), []byte("<"))
	resp = bytes.ReplaceAll(resp, []byte("&gt;"), []byte(">"))
	resp = bytes.ReplaceAll(resp, []byte("  "), []byte(""))
	return resp
}

// used to trim html labels in description
var htmlTrimRegp = regexp.MustCompile(`</?[^>]+>`)

// parseProduct
func (c *_Crawler) parseProduct(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil {
		return nil
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	canUrl, _ := c.CanonicalUrl(doc.Find(`link[rel="canonical"]`).AttrOr("href", ""))
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}
	brand := doc.Find(`meta[itemprop="brand"]`).Text()
	if brand == "" {
		brand = "Maje"
	}
	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           doc.Find(`#pid`).AttrOr("value", ""),
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
			GroupId:      doc.Find(`meta[property="product:age_group"]`).AttrOr(`content`, ``),
		},
		BrandName: brand,
		Title:     doc.Find(`.productSubname`).Text(),
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
	}

	// desc
	description := (doc.Find(`.wrapper-tabs`).Text())
	item.Description = string(TrimSpaceNewlineInString([]byte(description)))

	if strings.Contains(doc.Find(`meta[property="product:availability"]`).AttrOr(`content`, ``), "instock") {
		item.Stock.StockStatus = pbItem.Stock_InStock
	}

	item.Category = doc.Find(`meta[property="product:category"]`).AttrOr(`content`, ``)

	msrp, _ := strconv.ParsePrice(doc.Find(`.productPrices`).Find(`.price-standard`).Text())
	originalPrice, _ := strconv.ParsePrice(doc.Find(`.productPrices`).Find(`.price-sales`).Text())
	discount := 0.0
	if msrp == 0 {
		msrp = originalPrice
	}
	if msrp > originalPrice {
		discount = math.Ceil((msrp - originalPrice) / msrp * 100)
	}

	//images
	sel := doc.Find(`.product-primary-image`).Find(`picture`)
	for j := range sel.Nodes {
		node := sel.Eq(j)
		imgurl := strings.Split(node.Find(`img`).AttrOr(`data-src`, ``), "?")[0]

		item.Medias = append(item.Medias, pbMedia.NewImageMedia(
			strconv.Format(j),
			imgurl,
			imgurl+"?sw=1000&q=80",
			imgurl+"?sw=800&q=80",
			imgurl+"?sw=500&q=80",
			"", j == 0))
	}

	// Color
	cid := ""
	var colorSelected *pbItem.SkuSpecOption
	sel = doc.Find(`.swatches.Color`).Find(`li`)
	for i := range sel.Nodes {
		node := sel.Eq(i)
		if strings.Contains(node.AttrOr(`class`, ``), `selected`) {
			cid = strings.Split(node.Find(`a`).AttrOr("data-variationparameter", ""), "=")[1]
			colorSelected = &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    cid,
				Name:  node.Find(`a`).AttrOr("title", ""),
				Value: node.Find(`a`).AttrOr("title", ""),
				//Icon:  rawSku.Hoverimage,
			}
		}
	}

	sel = doc.Find(`.swatches.size`).Find(`li`)
	for i := range sel.Nodes {
		node := sel.Eq(i)

		sid := strings.Split(node.Find(`a`).AttrOr("data-variationparameter", ""), "=")[1]

		sku := pbItem.Sku{
			SourceId: fmt.Sprintf("%s-%s", cid, sid),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(originalPrice),
				Msrp:     int32(msrp),
				Discount: int32(discount),
			},
			Stock: &pbItem.Stock{StockStatus: pbItem.Stock_InStock},
		}

		if strings.Contains(node.AttrOr("class", ""), "unselectable") {
			sku.Stock.StockStatus = pbItem.Stock_OutOfStock
		}

		if colorSelected != nil {
			sku.Specs = append(sku.Specs, colorSelected)
		}

		// size
		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Id:    sid,
			Name:  strings.TrimSpace(node.Find(`.defaultSize`).Text()),
			Value: strings.TrimSpace(node.Find(`.defaultSize`).Text()),
		})

		item.SkuItems = append(item.SkuItems, &sku)
	}

	// yield item result
	if err = yield(ctx, &item); err != nil {
		c.logger.Errorf("yield sub request failed, error=%s", err)
		return err
	}

	return nil
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://us.maje.com/en/homepage",
		//"https://www.maje.com/us/2000441987.html?dwvar_2000441987_color=01",
		//"https://us.maje.com/en/categories/view-all-bags/",
		//"https://us.maje.com/en/categories/dresses/221rythonela/MFPRO01861.html?dwvar_MFPRO01861_color=L012",
		"https://us.maje.com/en/categories/medium-bags/220abag/MFASA00329.html?dwvar_MFASA00329_color=2517",
	} {
		req, err := http.NewRequest(http.MethodGet, u, nil)
		if err != nil {
			c.logger.Fatal(err)
		} else {
			reqs = append(reqs, req)
		}
	}
	return
}

// CheckTestResponse used to validate the response by test request.
// is error returns, there must be some error of the spider.
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
