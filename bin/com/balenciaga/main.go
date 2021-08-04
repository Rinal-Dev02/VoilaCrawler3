package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
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
func (_ *_Crawler) New(_ *cli.Context, client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient: client,
		// this regular used to match category page url path
		categoryPathMatcher: regexp.MustCompile(`^(/en([/A-Za-z0-9_-]+))|(/on/demandware\.store([/A-Za-z0-9_-]+))$`),
		productPathMatcher:  regexp.MustCompile(`^/en([/A-Za-z0-9_-]+).html$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "157036442a1e4eb782c76ef4115186e6"
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
	return []string{"*.balenciaga.com"}
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
	if p == "/en-us" {
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

	sel := dom.Find(`.c-nav__list.c-nav__level1>li`)

	for i := range sel.Nodes {
		node := sel.Eq(i)
		cateName := strings.TrimSpace(node.Find(`a`).First().Text())

		if cateName == "" {
			continue
		}

		nctx := context.WithValue(ctx, "Category", cateName)

		subSel := node.Find(`li[data-ref="group"]`)

		for k := range subSel.Nodes {
			subNode2 := subSel.Eq(k)

			subcat2 := strings.TrimSpace(subNode2.Find(`button`).First().Text())
			if subcat2 == "" {
				subcat2 = strings.TrimSpace(subNode2.Find(`a`).First().Text())
			}

			subNode3 := subNode2.Find(`li`)
			for j := range subNode3.Nodes {
				subNode := subNode3.Eq(j)
				subcategory3 := strings.TrimSpace(subNode.Find(`a`).First().Text())
				if subcategory3 == "" {
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

				if c.categoryPathMatcher.MatchString(u.Path) {
					nnctx := context.WithValue(nctx, "SubCategory", subcat2)
					nnnctx := context.WithValue(nnctx, "SubCategory2", subcategory3)
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

	sel := doc.Find(`.c-product__inner.c-product__focus`)

	for i := range sel.Nodes {
		node := sel.Eq(i)
		if href, _ := node.Attr("href"); href != "" {

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

	nextUrl := doc.Find(`.c-loadmore__btn.c-button--animation`).AttrOr("data-url", "")
	if nextUrl == "" {
		return nil
	}
	nextUrl = strings.ReplaceAll(nextUrl, "&sz=12", "&sz=96")

	totalCount, _ := strconv.ParsePrice(strings.TrimSpace(doc.Find(`.c-filters__count`).Text()))

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
var imageRegp = regexp.MustCompile(`/[A-Z-a-z_]+-`)
var productID = regexp.MustCompile(`-[A-Z-0-9]+.html`)

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
	brand := doc.Find(`type[Brand="name"]`).Text()
	if brand == "" {
		brand = "Balenciaga"
	}
	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           strings.TrimSpace(doc.Find(`span[data-bind="styleMaterialColor"]`).Text()),
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
			// GroupId:      doc.Find(`meta[property="product:age_group"]`).AttrOr(`content`, ``),
		},
		BrandName: brand,
		Title:     strings.TrimSpace(doc.Find(`.c-product__name`).Text()),
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
	}

	// desc
	description := (doc.Find(`.c-product__shortdesc`).Text())
	item.Description = string(TrimSpaceNewlineInString([]byte(description)))

	if strings.Contains(doc.Find(`.c-product__availabilitymsg`).AttrOr(`data-available`, ``), "true") {
		item.Stock.StockStatus = pbItem.Stock_InStock
	}

	sel := doc.Find(`.c-breadcrumbs.c-breadcrumbs--null>li`)
	for i := range sel.Nodes {
		if i >= len(sel.Nodes)-1 {
			continue
		}

		node := sel.Eq(i)
		breadcrumb := strings.TrimSpace(node.Text())

		if i == 1 {
			item.Category = breadcrumb
			item.CrowdType = breadcrumb
		} else if i == 2 {
			item.SubCategory = breadcrumb
		} else if i == 3 {
			item.SubCategory2 = breadcrumb
		} else if i == 4 {
			item.SubCategory3 = breadcrumb
		} else if i == 5 {
			item.SubCategory4 = breadcrumb
		}
	}

	currentPrice, _ := strconv.ParseFloat(doc.Find(`.c-price__value--current`).AttrOr("content", ""))
	msrp, _ := strconv.ParseFloat(doc.Find(`.c-price__value--old`).AttrOr("content", ""))

	if msrp == 0 {
		msrp = currentPrice
	}
	discount := 0
	if msrp > currentPrice {
		discount = (int)(((msrp - currentPrice) / msrp) * 100)
	}

	//images
	sel1 := doc.Find(`.c-productcarousel__wrapper`).Find(`li`)
	for j := range sel1.Nodes {
		node := sel1.Eq(j)
		imgurl := strings.Split(node.Find(`img`).AttrOr(`src`, ``), "?")[0]

		item.Medias = append(item.Medias, pbMedia.NewImageMedia(
			strconv.Format(j),
			imgurl,
			imageRegp.ReplaceAllString(imgurl, "/Large-"),
			imageRegp.ReplaceAllString(imgurl, "/Medium-"),
			imageRegp.ReplaceAllString(imgurl, "/Small-"),
			"", j == 0))
	}
	// Color
	cid := ""
	colorName := ""
	var colorSelected *pbItem.SkuSpecOption
	sel = doc.Find(`.c-swatches`).Find(`.c-swatches__item`)
	for i := range sel.Nodes {
		node := sel.Eq(i)

		if strings.Contains(node.Find(`.c-swatches__itemimage`).AttrOr(`class`, ``), `selected`) {
			cid = node.Find(`input`).AttrOr(`data-attr-value`, "")
			icon := strings.ReplaceAll(strings.ReplaceAll(node.Find(`.c-swatches__itemimage`).AttrOr(`style`, ""), "background-image: url(", ""), ")", "")
			colorName = node.Find(`.c-swatches__itemimage`).AttrOr(`data-display-value`, "")
			colorSelected = &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    cid,
				Name:  colorName,
				Value: colorName,
				Icon:  icon,
			}
		}
	}

	sel = doc.Find(`.c-product__sizebutton`).Find(`button`)
	for i := range sel.Nodes {
		node := sel.Eq(i)

		sid := (node.AttrOr("data-attr-value", ""))

		sku := pbItem.Sku{
			SourceId: fmt.Sprintf("%s-%s", cid, sid),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(currentPrice),
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
			Name:  sid,
			Value: sid,
		})

		item.SkuItems = append(item.SkuItems, &sku)
	}

	// yield item result
	if err = yield(ctx, &item); err != nil {
		c.logger.Errorf("yield sub request failed, error=%s", err)
		return err
	}

	// found other color
	sel = doc.Find(`.c-swatches`).Find(`.c-swatches__item`)
	for i := range sel.Nodes {
		node := sel.Eq(i)
		color := node.Find(`.c-swatches__itemimage`).AttrOr(`data-display-value`, "")
		c.logger.Debugf("found color %s %t", color, color == colorName)
		if color == "" || color == colorName {
			continue
		}
		u := node.Find(`input`).AttrOr("data-attr-href", "")
		if u == "" {
			continue
		}

		pid := "-" + strings.Split(strings.Split(u, "pid=")[1], "&")[0] + ".html"
		u = productID.ReplaceAllString(resp.Request.URL.Path, pid)

		req, err := http.NewRequest(http.MethodGet, u, nil)
		if err != nil {
			c.logger.Error(err)
			continue
		}
		if err := yield(ctx, req); err != nil {
			return err
		}
	}

	return nil
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://www.balenciaga.com/en-us",
		//"https://www.balenciaga.com/en-us/women/sales/view-all",
		//"https://www.balenciaga.com/en-us/track-sneaker-black-542023W1GB61002.html",
		"https://www.balenciaga.com/en-us/triple-s-clear-sole-sneaker-red-544351W2CE16500.html",
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
	cli.NewApp(&_Crawler{}).Run(os.Args)
}
