package main

import (
	"bytes"
	"context"
	"errors"
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
func (_ *_Crawler) New(_ *cli.Context, client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient: client,
		// this regular used to match category page url path
		categoryPathMatcher: regexp.MustCompile(`^/en([/A-Za-z0-9_-]+)$`),
		productPathMatcher:  regexp.MustCompile(`^/en([/A-Za-z0-9_-]+).html$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "a3de59805d1e4a20abade7b58c6cf66b"
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
		EnableSessionInit: false,
		Reliability:       proxy.ProxyReliability_ReliabilityDefault,
	}

	return opts
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.sandro-paris.com"}
}

// CanonicalUrl
func (c *_Crawler) CanonicalUrl(rawurl string) (string, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	if u.Host == "" {
		u.Host = "www.sandro-paris.com"
	}
	if c.productPathMatcher.MatchString(u.Path) {
		u.RawQuery = ""
	}
	return u.String(), nil
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
	if p == "/en/womens" || p == "/en/mens" {
		return crawler.ErrUnsupportedPath
	}

	if c.productPathMatcher.MatchString(p) {
		return c.parseProduct(ctx, resp, yield)
	} else if c.categoryPathMatcher.MatchString(p) {
		return c.parseCategoryProducts(ctx, resp, yield)
	}

	return crawler.ErrUnsupportedPath
}

// nextIndex used to get the index from the shared data.
// item.index is a const key for item index.
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")))
}

type categorylists struct {
	URL   string
	Label string
}

func (c *_Crawler) GetCategories(ctx context.Context) ([]*pbItem.Category, error) {

	var viewData []categorylists

	viewData = append(viewData, categorylists{URL: "https://us.sandro-paris.com/en/womens/", Label: "Women"})

	viewData = append(viewData, categorylists{URL: "https://us.sandro-paris.com/en/mens/", Label: "Men"})

	var (
		cates   []*pbItem.Category
		cateMap = map[string]*pbItem.Category{}
	)
	if err := func(yield func(names []string, url string) error) error {

		for _, catUrl := range viewData {

			req, err := http.NewRequest(http.MethodGet, catUrl.URL, nil)
			req.Header.Add("accept", "*/*")
			req.Header.Add("referer", "https://us.sandro-paris.com/")
			req.Header.Add("accept-language", "en-US,en;q=0.9")
			req.Header.Add("x-requested-with", "XMLHttpRequest")

			catreq, err := c.httpClient.Do(ctx, req)
			if err != nil {
				panic(err)
			}
			defer catreq.Body.Close()

			catBody, err := ioutil.ReadAll(catreq.Body)
			if err != nil {
				c.logger.Error(err)
				return err
			}

			dom, err := goquery.NewDocumentFromReader(bytes.NewReader(catBody))
			if err != nil {
				c.logger.Error(err)
				return err
			}

			sel := dom.Find(`.menu-category>li`)
			for i := range sel.Nodes {
				node := sel.Eq(i)

				cateName := strings.TrimSpace(node.Find(`a`).First().Text())

				if cateName == "" {
					continue
				}

				subSel := node.Find(`.menu-wrapper.level-2>div`)

				subcat2 := ""

				for k := range subSel.Nodes {
					subNode := subSel.Eq(k)

					if strings.TrimSpace(subNode.Find(`p`).First().Text()) != "other categories" {
						subcat2 = strings.TrimSpace(subNode.Find(`p`).First().Text())
					}

					subNode3 := subNode.Find(`li`)
					for j := range subNode3.Nodes {

						subNode1 := subNode3.Eq(j)
						subcat3 := strings.TrimSpace(subNode1.Find(`a`).Text())

						href := subNode1.Find(`a`).First().AttrOr("href", "")
						if href == "" || subcat3 == "" {
							continue
						}

						canonicalhref, err := c.CanonicalUrl(href)
						if err != nil {
							c.logger.Error("parse url %s failed", href)
							continue
						}

						u, err := url.Parse(canonicalhref)
						if err != nil {
							c.logger.Error("parse url %s failed", href)
							continue
						}

						if c.categoryPathMatcher.MatchString(u.Path) {
							if err := yield([]string{catUrl.Label, cateName, subcat2, subcat3}, canonicalhref); err != nil {
								return err
							}
						}
					}
				}
			}
		}
		return nil
	}(func(names []string, url string) error {
		if len(names) == 0 {
			return errors.New("no valid category name found")
		}

		var (
			lastCate *pbItem.Category
			path     string
		)
		for i, name := range names {
			path = strings.Join([]string{path, name}, "-")

			name = strings.Title(strings.ToLower(name))
			if cate, _ := cateMap[path]; cate != nil {
				lastCate = cate
				continue
			} else {
				cate = &pbItem.Category{
					Name: name,
				}
				cateMap[path] = cate
				if lastCate != nil {
					lastCate.Children = append(lastCate.Children, cate)
				}
				lastCate = cate

				if i == 0 {
					cates = append(cates, cate)
				}
			}
		}
		lastCate.Url = url
		return nil
	}); err != nil {
		c.logger.Error(err)
		return nil, err
	}
	return cates, nil
}

// parseCategoryProducts parse api url from web page url
func (c *_Crawler) parseCategoryProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error(err)
		return err
	}

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	lastIndex := nextIndex(ctx)
	sel := dom.Find(`.product-name`)
	for i := range sel.Nodes {
		node := sel.Eq(i)

		href := node.Find("a").AttrOr("href", "")
		if href == "" {
			c.logger.Warnf("no href found")
			continue
		}

		req, err := http.NewRequest(http.MethodGet, href, nil)
		if err != nil {
			c.logger.Errorf("create request with url %s failed", href)
			continue
		}
		nctx := context.WithValue(ctx, "item.index", lastIndex)
		lastIndex += 1
		if err := yield(nctx, req); err != nil {
			return err
		}
	}

	totalProducts, _ := strconv.ParsePrice(dom.Find(`.nbProducts`).Text())

	if lastIndex >= (int)(totalProducts) {
		return nil
	}

	nextUrl := dom.Find(`link[rel="next"]`).AttrOr("href", "")
	if nextUrl == "" {
		return nil
	}
	nextUrl = strings.ReplaceAll(nextUrl, `&sz=24`, `&sz=96`)
	req, err := http.NewRequest(http.MethodGet, nextUrl, nil)
	if err != nil {
		return c.logger.Errorf("create request with url %s failed", nextUrl).ToError()
	}
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
}

func TrimSpaceNewlineInString(s string) string {
	re := regexp.MustCompile(`\n`)
	resp := re.ReplaceAllString(s, " ")
	resp = strings.ReplaceAll(resp, "\\n", " ")
	resp = strings.ReplaceAll(resp, "\r", " ")
	resp = strings.ReplaceAll(resp, "\t", " ")
	re = regexp.MustCompile(`\s+`)
	resp = strings.TrimSpace(re.ReplaceAllString(resp, " "))
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
		brand = "Sandro"
	}

	pid := doc.Find(`#pid`).AttrOr("value", "")
	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           pid,
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
			GroupId:      doc.Find(`meta[property="product:age_group"]`).AttrOr(`content`, ``),
		},
		BrandName: brand,
		Title:     TrimSpaceNewlineInString(doc.Find(`.prod-title`).Text()),
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
	}

	if strings.Contains(doc.Find(`meta[property="product:availability"]`).AttrOr(`content`, ``), "instock") {
		item.Stock.StockStatus = pbItem.Stock_InStock
	}

	description := doc.Find(`.shortDescription`).Text()
	item.Description = TrimSpaceNewlineInString(description)

	sel := doc.Find(`.breadcrumb>ol>li`)
	for i := range sel.Nodes {
		if i == len(sel.Nodes)-1 {
			continue
		}
		node := sel.Eq(i)
		breadcrumb := strings.TrimSpace(node.Text())

		if i == 0 {
			item.Category = breadcrumb
		} else if i == 1 {
			item.SubCategory = breadcrumb
		} else if i == 2 {
			item.SubCategory2 = breadcrumb
		} else if i == 3 {
			item.SubCategory3 = breadcrumb
		} else if i == 4 {
			item.SubCategory4 = breadcrumb
		}
	}

	msrp, _ := strconv.ParsePrice(doc.Find(`.product-price`).Find(`.price-standard`).Text())
	originalPrice, _ := strconv.ParsePrice(doc.Find(`.product-price`).Find(`.price-sales`).Text())
	discount := 0.0
	if msrp == 0 {
		msrp = originalPrice
	}
	if msrp > originalPrice {
		discount = math.Ceil((msrp - originalPrice) / msrp * 100)
	}

	dom := doc
	selColor := doc.Find(`.swatches.Color>li`)
	for ic := range selColor.Nodes {
		nodeColor := selColor.Eq(ic)

		cid := strings.Split(nodeColor.Find(`a`).AttrOr("data-variationparameter", ""), "=")[1]
		variationName := nodeColor.Find(`a`).AttrOr("title", "")
		variationurl := nodeColor.Find(`a`).AttrOr("href", "")

		for _, spec := range item.SkuItems {
			for _, specColor := range spec.Specs {
				if specColor.Type == pbItem.SkuSpecType_SkuSpecColor && specColor.Name == variationName {
					variationName = ""
					break
				}
			}
		}

		if variationName == "" {
			continue
		}
		if subClass := nodeColor.AttrOr(`class`, ``); strings.Contains(subClass, `selected`) {
			dom = doc
		} else {
			// new request
			respBodyJs, err := c.variationRequest(ctx, variationurl, resp.Request.URL.String())
			if err != nil {
				c.logger.Error(err)
				return err
			}

			dom, err = goquery.NewDocumentFromReader(bytes.NewReader(respBodyJs))
			if err != nil {
				c.logger.Error(err)
				return err
			}
		}

		msrp, _ := strconv.ParsePrice(doc.Find(`.product-price`).Find(`.price-standard`).Text())
		originalPrice, _ := strconv.ParsePrice(doc.Find(`.product-price`).Find(`.price-sales`).Text())
		discount = 0.0
		if msrp == 0 {
			msrp = originalPrice
		}
		if msrp > originalPrice {
			discount = math.Ceil((msrp - originalPrice) / msrp * 100)
		}

		var medias []*pbMedia.Media
		//images
		sel := dom.Find(`.image-container.image-zoom`).Find(`picture`)
		for j := range sel.Nodes {
			node := sel.Eq(j)
			imgurl := strings.Split(node.Find(`img`).AttrOr(`data-src`, ``), "?")[0]
			medias = append(medias, pbMedia.NewImageMedia(
				strconv.Format(j),
				imgurl,
				imgurl+"?sw=1000&q=80",
				imgurl+"?sw=800&q=80",
				imgurl+"?sw=500&q=80",
				"", j == 0))
		}

		sel1 := dom.Find(`.swatches.size`).Find(`li`)
		for i := range sel1.Nodes {
			node1 := sel1.Eq(i)

			sid := strings.Split(node1.Find(`a`).AttrOr("data-variationparameter", ""), "=")[1]

			sku := pbItem.Sku{
				SourceId: pid,

				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(originalPrice * 100),
					Msrp:     int32(msrp * 100),
					Discount: int32(discount),
				},
				Medias: medias,
				Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_InStock},
			}

			if strings.Contains(node1.Find(`a`).AttrOr("title", ""), "unavailable") {
				sku.Stock.StockStatus = pbItem.Stock_OutOfStock
			}

			if variationName != "" {
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecColor,
					Id:    cid,
					Name:  variationName,
					Value: variationName,
				})
			}

			sizeName := TrimSpaceNewlineInString(strings.ReplaceAll(node1.Find(`.defaultSize`).Text(), `- Sold Out`, ``))
			if sizeName != "" {
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecSize,
					Id:    sid,
					Name:  sizeName,
					Value: sizeName,
				})
			}

			if len(sku.Specs) == 0 {
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecColor,
					Id:    "-",
					Name:  "-",
					Value: "-",
				})
			}

			for _, spec := range sku.Specs {
				sku.SourceId += fmt.Sprintf("-%s", spec.Id)
			}

			item.SkuItems = append(item.SkuItems, &sku)
		}

		// only color
		if len(sel1.Nodes) == 0 {
			sku := pbItem.Sku{
				SourceId: pid,
				//SourceId: sid,
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(originalPrice * 100),
					Msrp:     int32(msrp * 100),
					Discount: int32(discount),
				},
				Medias: medias,
				Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_InStock},
			}

			if strings.Contains(nodeColor.AttrOr("class", ""), "unselectable") {
				sku.Stock.StockStatus = pbItem.Stock_OutOfStock
			}

			if variationName != "" {
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecColor,
					Id:    cid,
					Name:  variationName,
					Value: variationName,
				})
			}

			if len(sku.Specs) == 0 {
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecColor,
					Id:    "-",
					Name:  "-",
					Value: "-",
				})
			}

			for _, spec := range sku.Specs {
				sku.SourceId += fmt.Sprintf("-%s", spec.Id)
			}

			item.SkuItems = append(item.SkuItems, &sku)
		}
	}

	// yield item result
	if err = yield(ctx, &item); err != nil {
		c.logger.Errorf("yield sub request failed, error=%s", err)
		return err
	}

	return nil
}

func (c *_Crawler) variationRequest(ctx context.Context, url string, referer string) ([]byte, error) {

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	opts := c.CrawlOptions(req.URL)
	req.Header.Set("accept-language", "en-GB,en-US;q=0.9,en;q=0.8")
	req.Header.Set("accept", "text/html, */*; q=0.01")
	req.Header.Set("x-requested-with", "XMLHttpRequest")
	req.Header.Set("referer", referer)

	for _, c := range opts.MustCookies {
		req.AddCookie(c)
	}
	for k := range opts.MustHeader {
		req.Header.Set(k, opts.MustHeader.Get(k))
	}
	resp, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
		EnableProxy:       true,
		EnableHeadless:    false,
		EnableSessionInit: false,
		Reliability:       opts.Reliability,
	})
	if err != nil {
		c.logger.Error(err)
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://us.sandro-paris.com/en/womens/",
		//"https://us.sandro-paris.com/en/mens/clothing/sweatshirts/",
		//"https://us.sandro-paris.com/en/mens/clothing/t-shirts-and-polos/linen-t-shirt/SHPTS00222.html",
		//"https://us.sandro-paris.com/en/womens/clothing/dresses/",
		//"https://us.sandro-paris.com/en/mens/clothing/pants-and-shorts/jersey-pants/SHPPA00609.html",
		//"https://us.sandro-paris.com/en/womens/clothing/dresses/dress-with-tailored-collar/SFPRO01694.html",
		//"https://us.sandro-paris.com/en/mens/clothing/shirts/flowing-shirt-with-stripe-print/SHPCM00393.html?dwvar_SHPCM00393_color=C102#start=1",
		"https://us.sandro-paris.com/en/mens/clothing/coats/long-coat-fastened-with-three-buttons/SHPMA00127.html",
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
