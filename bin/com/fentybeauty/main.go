package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
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
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
	"github.com/voiladev/go-framework/text"
)

// _Crawler defined the crawler struct/class for which is not necessory to be exportable
type _Crawler struct {
	// httpClient is the object of an http client
	httpClient           http.Client
	categoryPathMatcher  *regexp.Regexp
	productPathMatcher   *regexp.Regexp
	productPathMatcherV1 *regexp.Regexp
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
		categoryPathMatcher: regexp.MustCompile(`^(/[a-z0-9-]+){1,6}$`),
		// this regular used to match product page url path
		productPathMatcher:   regexp.MustCompile(`^(/[a-zA-Z0-9\-]+){1,4}.html$`),
		productPathMatcherV1: regexp.MustCompile(`^/products(/[a-zA-Z0-9\-]+){1,4}$`),
		logger:               logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "297d5024b6a8058c6f1d76f571c7bf4c"
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
	}
	return opts
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.fentybeauty.com"}
}

func (c *_Crawler) CanonicalUrl(rawurl string) (string, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	if u.Host == "" {
		u.Host = "www.fentybeauty.com"
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

	if p == "" || p == "/pages/fentyskin" {
		return c.parseCategories(ctx, resp, yield)
	}

	if c.productPathMatcher.MatchString(resp.RawUrl().Path) {
		return c.parseProduct(ctx, resp, yield)
	} else if c.productPathMatcherV1.MatchString(resp.RawUrl().Path) {
		return c.parseProductV1(ctx, resp, yield)
	} else if c.categoryPathMatcher.MatchString(resp.RawUrl().Path) {
		respBody, err := resp.RawBody()
		if err != nil {
			return err
		}
		if bytes.Contains(respBody, []byte(`grid-tile`)) {
			return c.parseCategoryProducts(ctx, resp, yield)
		} else if bytes.Contains(respBody, []byte(`is="collection"`)) {
			return c.parseCategoryProductsV1(ctx, resp, yield)
		}
	}
	return crawler.ErrUnsupportedPath
}

func (c *_Crawler) parseCategories(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	respBody, err := resp.RawBody()
	if err != nil {
		return err
	}

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	sel := dom.Find(`nav.header-nav--desktop ul.header-nav__items > li`)
	for i := range sel.Nodes {
		node := sel.Eq(i)
		cateName := text.Clean(node.Find(`a`).First().AttrOr("title", node.Find(`a`).First().Find("span>span").Text()))
		if cateName == "" {
			continue
		}
		nnctx := context.WithValue(ctx, "Category", cateName)

		subSel := node.Find(`.header-nav__subnav .shop-meganav__items>div`)
		for k := range subSel.Nodes {
			subNode := subSel.Eq(k)
			if strings.Contains(subNode.AttrOr("class", ""), "shop-meganav__item-wrapper") {
				aSel := subNode.Find(`a.js-sub-link`)
				for ai := range aSel.Nodes {
					aNode := aSel.Eq(ai)
					href := aNode.AttrOr("href", "")
					u, err := url.Parse(href)
					if err != nil {
						c.logger.Error("parse url %s failed", href)
						continue
					}
					subCate := text.Clean(aNode.AttrOr("title", aNode.Text()))
					if c.categoryPathMatcher.MatchString(u.Path) {
						nnnctx := context.WithValue(nnctx, "SubCategory", subCate)
						req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
						if err := yield(nnnctx, req); err != nil {
							return err
						}
					}
				}
			} else if strings.Contains(subNode.AttrOr("class", ""), "shop-meganav__submenu") {
				aNode := subNode.Find(`a`).First()
				href := aNode.AttrOr("href", "")
				u, err := url.Parse(href)
				if err != nil {
					c.logger.Error("parse url %s failed", href)
					continue
				}
				subCate := text.Clean(aNode.AttrOr("title", aNode.Text()))
				nnnctx := context.WithValue(nnctx, "SubCategory", subCate)
				if c.categoryPathMatcher.MatchString(u.Path) {
					req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
					if err := yield(nnnctx, req); err != nil {
						return err
					}
				}

				subSel2 := subNode.Find(`.shop-meganav__submenu__links a.js-sub-link`)
				for l := range subSel2.Nodes {
					aNode := subSel2.Eq(l)
					href := aNode.AttrOr("href", "")
					u, err := url.Parse(href)
					if err != nil {
						c.logger.Error("parse url %s failed", href)
						continue
					}
					subCate := text.Clean(aNode.AttrOr("title", aNode.Text()))
					if c.categoryPathMatcher.MatchString(u.Path) {
						nnnnctx := context.WithValue(nnnctx, "SubCategory2", subCate)
						req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
						if err := yield(nnnnctx, req); err != nil {
							return err
						}
					}
				}
			}

		}
	}
	return nil
}

// nextIndex used to get the index from the shared data.
// item.index is a const key for item index.
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")))
}

// used to extract embaded json data in website page.
// more about golang regulation see here https://golang.org/pkg/regexp/syntax/
var productsExtractReg = regexp.MustCompile(`(?U)"impressions":\s*(\[.*\]),`)

// parseCategoryProducts parse api url from web page url
func (c *_Crawler) parseCategoryProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	respBody, err := resp.RawBody()
	if err != nil {
		return err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	lastIndex := nextIndex(ctx)

	sel := doc.Find(`li.grid-tile`)
	c.logger.Debugf("nodes %d", len(sel.Nodes))
	for i := range sel.Nodes {
		node := sel.Eq(i)
		if href, _ := node.Find(`.product-tile .product-link[itemprop="url"]`).Attr("href"); href != "" {
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

	if !bytes.Contains(respBody, []byte("<div class=\"infinite-scroll-placeholder\"")) {
		// nextpage not found
		return nil
	}
	nextUrl := doc.Find(".infinite-scroll-placeholder").AttrOr("data-grid-url", "")
	if nextUrl == "" {
		return nil
	}
	nextUrl = html.UnescapeString(nextUrl)

	req, _ := http.NewRequest(http.MethodGet, nextUrl, nil)
	vals := req.URL.Query()
	vals.Set("sz", "48")
	req.URL.RawQuery = vals.Encode()

	// update the index of last page
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
}

func (c *_Crawler) parseCategoryProductsV1(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	respBody, err := resp.RawBody()
	if err != nil {
		return err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	matched := doc.Find(`div[is="collection"]`).AttrOr("first-products", "")
	if matched == "" {
		return nil
	}

	prodInfoStr, err := url.QueryUnescape(matched)
	if err != nil {
		c.logger.Error(err)
		return err
	}
	var prodInfo productVariants

	if err := json.Unmarshal([]byte(prodInfoStr), &prodInfo); err != nil {
		c.logger.Error(err)
		return err
	}

	lastIndex := nextIndex(ctx)

	for _, item := range prodInfo {

		if href := item.URL; href != "" {

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

	if !bytes.Contains(respBody, []byte("<div class=\"infinite-scroll-placeholder\"")) {
		// nextpage not found
		return nil
	}
	nextUrl := doc.Find(".infinite-scroll-placeholder").AttrOr("data-grid-url", "")
	if nextUrl == "" {
		return nil
	}
	nextUrl = html.UnescapeString(nextUrl)

	req, _ := http.NewRequest(http.MethodGet, nextUrl, nil)
	vals := req.URL.Query()
	vals.Set("sz", "48")
	req.URL.RawQuery = vals.Encode()

	// update the index of last page
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
}

type ImageDetail struct {
	Large []struct {
		Zoom  string `json:"zoom"`
		URL   string `json:"url"`
		Alt   string `json:"alt"`
		Title string `json:"title"`
	} `json:"large"`
}

var productInfoReg = regexp.MustCompile(`(?U)_bluecoreTrack\.push\(\["trackProductView",\s*"(.*)",\s*\d+\s*\]\);`)

// parseProduct
// TODO: for product set, not yield sub request for every prod.
func (c *_Crawler) parseProduct(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {

	if c == nil || yield == nil {
		return nil
	}

	respbody, err := resp.RawBody()
	if err != nil {
		return err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respbody))
	if err != nil {
		return err
	}

	matched := productInfoReg.FindStringSubmatch(string(respbody))
	if len(matched) < 2 {
		return fmt.Errorf("website format changed")
	}
	prodInfoStr, err := url.QueryUnescape(matched[1])
	if err != nil {
		c.logger.Error(err)
		return err
	}
	var prodInfo struct {
		ID           string   `json:"id"`
		Name         string   `json:"name"`
		URL          string   `json:"url"`
		Brand        string   `json:"brand"`
		Price        float64  `json:"price"`
		Image        string   `json:"image"`
		OutOfStock   bool     `json:"outOfStock"`
		Category     string   `json:"category"`
		Breadcrumbs  []string `json:"breadcrumbs"`
		IsMaster     bool     `json:"isMaster"`
		IsProduct    bool     `json:"isProduct"`
		IsProductSet bool     `json:"isProductSet"`
		IsVariant    bool     `json:"isVariant"`
	}
	if err := json.Unmarshal([]byte(prodInfoStr), &prodInfo); err != nil {
		c.logger.Error(err)
		return err
	}

	canUrl := doc.Find(`link[rel="canonical"]`).AttrOr("href", "")
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           prodInfo.ID,
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
		},
		Title:       prodInfo.Name,
		Description: strings.TrimSpace(doc.Find(`div[itemprop="description"]`).Text()),
		BrandName:   prodInfo.Brand,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
	}
	if item.Source.Id == "" {
		item.Source.Id = strings.TrimSpace(doc.Find(`.product-number span[itemprop="productID"]`).Text())
		if item.Source.Id == "" {
			item.Source.Id = doc.Find(`div[data-bv-show="rating_summary"]`).AttrOr(`data-bv-productId`, "")
		}
	}
	if item.Source.CanonicalUrl == "" {
		item.Source.CanonicalUrl = prodInfo.URL
	}
	if item.Title == "" {
		item.Title = strings.TrimSpace(doc.Find(`#product-content .justdetails div.product-name:first`).Text())
	}
	if item.BrandName == "" {
		item.BrandName = "Fenty Beauty"
	}
	if prodInfo.OutOfStock {
		item.Stock.StockStatus = pbItem.Stock_OutOfStock
	}
	if len(prodInfo.Breadcrumbs) > 0 {
		for i, cate := range prodInfo.Breadcrumbs {
			switch i {
			case 0:
				item.Category = cate
			case 1:
				item.SubCategory = cate
			case 2:
				item.SubCategory2 = cate
			case 3:
				item.SubCategory3 = cate
			case 4:
				item.SubCategory4 = cate
			}
		}
	} else {
		sel := doc.Find(`.breadcrumb>li`)
		for i := range sel.Nodes {
			if i == 0 {
				continue
			}
			node := sel.Eq(i)
			breadcrumb := strings.TrimSpace(node.Text())

			if i == 1 {
				item.Category = breadcrumb
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
	}

	currentPrice := prodInfo.Price
	msrp := prodInfo.Price
	discount := float64(0)

	if prodInfo.IsProductSet {
		if currentPrice == 0 {
			fields := strings.SplitN(strings.TrimSpace(doc.Find(`.justdetails .product-price .product-price .price-sales`).Text()), " ", 2)
			currentPrice, _ = strconv.ParsePrice(fields[0])
			if len(fields) > 1 {
				msrp, _ = strconv.ParsePrice(fields[1])
			}
		}
	} else {
		if val := doc.Find(`.justdetails .product-price .price-sales meta[itemprop="price"]`).AttrOr("content", ""); val != "" {
			currentPrice, _ = strconv.ParsePrice(val)
		}
		if val := doc.Find(`.justdetails .product-price .price-standard meta[itemprop="price"]`).AttrOr("content", ""); val != "" {
			msrp, _ = strconv.ParsePrice(val)
		}
	}
	if msrp == 0 {
		msrp = currentPrice
	}
	if msrp > currentPrice {
		discount = math.Ceil((msrp - currentPrice) / msrp * 100)
	}

	colorGroupSel := doc.Find(`.product-variations .attribute.Color .swatches-wrap ul.swatches.Color`)
	for i := range colorGroupSel.Nodes {
		colorSel := colorGroupSel.Eq(i)
		if strings.Contains(colorSel.AttrOr("class", ""), "filtered") {
			continue
		}
		sel := colorSel.Find(`li`)
		for j := range sel.Nodes {
			node := sel.Eq(j)
			val := strings.TrimSpace(node.Find(`a .swatch-displayvalue`).Text())
			name := strings.TrimSpace(node.Find(`a .swatch-subname`).Text())

			colorSpec := pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    node.Find(`a`).AttrOr("data-swatchcolor", node.Find(`a[role="radio"]`).AttrOr("data-vv-id", "")),
				Name:  name,
				Value: val,
			}

			sku := pbItem.Sku{
				SourceId: val,
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(currentPrice * 100),
					Msrp:     int32(msrp * 100),
					Discount: int32(discount),
				},
				Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}

			if !strings.Contains(node.AttrOr("class", ""), `unselectable`) {
				sku.Stock.StockStatus = pbItem.Stock_InStock
				item.Stock.StockStatus = pbItem.Stock_InStock
			}

			sku.Specs = append(sku.Specs, &colorSpec)

			imgval, _ := node.Find(`a`).Attr(`data-allimages`)
			var imgData ImageDetail
			if err := json.Unmarshal([]byte(imgval), &imgData); err != nil {
				c.logger.Debugf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
				return err
			}
			for j, mediumUrl := range imgData.Large {
				sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
					strconv.Format(j),
					mediumUrl.Zoom,
					mediumUrl.Zoom+"?sw=1000",
					mediumUrl.Zoom+"?sw=600",
					mediumUrl.Zoom+"?sw=500",
					"",
					j == 0,
				))
			}
			item.SkuItems = append(item.SkuItems, &sku)
		}
		item.Medias = append(item.Medias, item.GetMedias()...)
		break
	}
	if len(item.SkuItems) == 0 {
		sizeval := "One Size"
		Sourceid, _ := doc.Find(`#pid`).Attr(`value`)
		sku := pbItem.Sku{
			SourceId: Sourceid,
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(currentPrice * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},
			Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}
		if bytes.Contains(respbody, []byte("<p class=\"in-stock-msg\">In Stock</p>")) {
			sku.Stock.StockStatus = pbItem.Stock_InStock
			item.Stock.StockStatus = pbItem.Stock_InStock
		}

		var sel *goquery.Selection
		if bytes.Contains(respbody, []byte("class=\"product-thumbnails-carousel")) {
			sel = doc.Find(`ul[class^="product-thumbnails-carousel"]`).Find(`li`).Find(`img`)
		} else {
			sel = doc.Find(`#product-set-list>img`)
		}

		for j := range sel.Nodes {
			node := sel.Eq(j)
			mediumUrl, _ := node.Attr("data-src")

			if mediumUrl == "" {
				mediumUrl, _ = node.Attr("srcset")
				if mediumUrl == "" {
					continue
				}
			}
			s := strings.Split(mediumUrl, ".jpg")

			sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
				strconv.Format(j),
				s[0]+".jpg",
				s[0]+".jpg?sw=1000",
				s[0]+".jpg?sw=600",
				s[0]+".jpg?sw=500",
				"",
				j == 0))
		}
		item.Medias = sku.Medias
		// size
		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Id:    Sourceid,
			Name:  sizeval,
			Value: sizeval,
		})
		item.SkuItems = append(item.SkuItems, &sku)
	}

	// yield item result
	if err = yield(ctx, &item); err != nil {
		return err
	}
	return nil
}

var htmlTrimRegp = regexp.MustCompile(`</?[^>]+>`)

func (c *_Crawler) parseProductV1(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {

	if c == nil || yield == nil {
		return nil
	}

	respbody, err := resp.RawBody()
	if err != nil {
		return err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respbody))
	if err != nil {
		return err
	}

	matched := doc.Find(`section[is="product-selector"]`).AttrOr("variants", "")
	if matched == "" {
		matched = doc.Find(`section[is="product-hero"]`).AttrOr("product-json", "")
		if matched != "" {
			matched = "[" + matched + "]"
		} else {
			return nil
		}
	}

	prodInfoStr, err := url.QueryUnescape(matched)
	if err != nil {
		c.logger.Error(err)
		return err
	}
	var prodInfo productVariants

	if err := json.Unmarshal([]byte(prodInfoStr), &prodInfo); err != nil {
		c.logger.Error(err)
		return err
	}

	canUrl := doc.Find(`link[rel="canonical"]`).AttrOr("href", "")
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}

	for _, prod := range prodInfo {

		item := pbItem.Product{
			Source: &pbItem.Source{
				Id:           prod.ID,
				CrawlUrl:     resp.Request.URL.String(),
				CanonicalUrl: prod.URL,
			},
			Title:       prod.Title,
			Description: strings.TrimSpace(htmlTrimRegp.ReplaceAllString(doc.Find(`.product-detail__contents`).Text(), "")),
			BrandName:   "Fenty Beauty",
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
			},
			Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}

		if prod.Available {
			item.Stock.StockStatus = pbItem.Stock_InStock
		}

		//Note: Category Not available
		//item.Category = "Home"

		sizeIndex := -1
		colorIndex := -1
		for i, key := range prod.OptionNames {
			if key == "Color" {
				colorIndex = i
			} else if key == "Size" {
				sizeIndex = i
			}
		}

		var medias []*pbMedia.Media
		for j, mediumUrl := range prod.Media {
			template := strings.Split(mediumUrl.Src, ".jpg")[0]
			medias = append(medias, pbMedia.NewImageMedia(
				strconv.Format(j),
				template+".jpg",
				mediumUrl.Src+"_1000x.jpg",
				mediumUrl.Src+"_600x.jpg",
				mediumUrl.Src+"_500x.jpg",
				"",
				j == 0,
			))
		}
		item.Medias = append(item.Medias, medias...)

		for _, rawsku := range prod.Variants {

			currentPrice, _ := strconv.ParseInt(rawsku.Price)
			msrp, _ := strconv.ParseInt(rawsku.CompareAtPrice)
			discount := float32(0)

			if msrp == 0 {
				msrp = currentPrice
			}
			if msrp > currentPrice {
				discount, _ = strconv.ParseFloat32((msrp - currentPrice) / msrp * 100)
			}

			sku := pbItem.Sku{
				SourceId: rawsku.Sku,
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(currentPrice),
					Msrp:     int32(msrp),
					Discount: int32(discount),
				},
				Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
				Medias: medias,
			}

			if prod.Available {
				sku.Stock.StockStatus = pbItem.Stock_InStock
				item.Stock.StockStatus = pbItem.Stock_InStock
			}

			if colorIndex > -1 {
				colorVal := ""
				if colorIndex == 0 {
					colorVal = rawsku.Option1
				} else if colorIndex == 1 {
					colorVal = rawsku.Option2
				} else {
					colorVal = rawsku.Option3
				}

				colorSpec := pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecColor,
					Id:    rawsku.ID,
					Name:  colorVal,
					Value: colorVal,
				}

				sku.Specs = append(sku.Specs, &colorSpec)
			}

			if sizeIndex > -1 {
				sizeVal := ""
				if sizeIndex == 0 {
					sizeVal = rawsku.Option1
				} else if sizeIndex == 1 {
					sizeVal = rawsku.Option2
				} else {
					sizeVal = rawsku.Option3
				}

				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecSize,
					Id:    rawsku.ID,
					Name:  sizeVal,
					Value: sizeVal,
				})
			}

			item.SkuItems = append(item.SkuItems, &sku)
		}

		// yield item result
		if err = yield(ctx, &item); err != nil {
			return err
		}
	}
	return nil
}

type productVariants []struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Vendor      string `json:"vendor"`
	Type        string `json:"type"`
	Media       []struct {
		Src               string `json:"src"`
		Alt               string `json:"alt"`
		AttachedToVariant string `json:"attached_to_variant"`
		Width             string `json:"width"`
		Height            string `json:"height"`
	} `json:"media"`
	Videos   []interface{} `json:"videos"`
	Variants []struct {
		Available string `json:"available"`
		Title     string `json:"title"`
		ID        string `json:"id"`
		Image     struct {
			Src string `json:"src"`
			Alt string `json:"alt"`
		} `json:"image"`
		Price          string `json:"price"`
		Sku            string `json:"sku"`
		CompareAtPrice string `json:"compare_at_price"`
		Option1        string `json:"option1"`
		Option2        string `json:"option2"`
		Option3        string `json:"option3"`
	} `json:"variants"`
	Categories        []interface{} `json:"categories"`
	Available         bool          `json:"available"`
	Handle            string        `json:"handle"`
	Title             string        `json:"title"`
	FeaturedImage     string        `json:"featuredImage"`
	Tags              []string      `json:"tags"`
	Price             string        `json:"price"`
	OptionNames       []string      `json:"optionNames"`
	OptionsWithValues struct {
		Color []string `json:"Color"`
	} `json:"optionsWithValues"`
	MicroCollectionHandle string        `json:"micro_collection_handle"`
	MicroCollectionTags   []string      `json:"micro_collection_tags"`
	Shade                 string        `json:"shade"`
	ShadeHandle           string        `json:"shade_handle"`
	ShadeCount            string        `json:"shadeCount"`
	HideShadeRangeFilter  string        `json:"hide_shade_range_filter"`
	IsBundle              bool          `json:"isBundle"`
	BundledProducts       []interface{} `json:"bundled_products"`
	FullPriceValue        string        `json:"full_price_value"`
	FullPriceLabel        string        `json:"full_price_label"`
	SwatchColorID         string        `json:"swatch_color_id"`
	URL                   string        `json:"url"`
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://www.fentybeauty.com/",
		//"https://www.fentybeauty.com/makeup-face",
		// "https://www.fentybeauty.com/powder-puff-setting-brush-170/27464.html?cgid=makeup-face-powder",
		// "https://www.fentybeauty.com/pro-filtr-instant-retouch-setting-powder/FB30011.html?dwvar_FB30011_color=FB9005&cgid=makeup-face-powder",
		// "https://www.fentybeauty.com/soft-matte-complexion-essentials-with-brush/pro-filter-foundation-essentials-brush.html?cgid=makeup-face",
		// "https://www.fentybeauty.com/pro-filtr-soft-matte-longwear-foundation/FB30006.html?dwvar_FB30006_color=FB0340&cgid=makeup-face-foundation",
		//"https://www.fentybeauty.com/two-lil-stunnas-mini-longwear-fluid-lip-color-duo/47670.html?cgid=sale",
		//"https://www.fentybeauty.com/mattifying-complexion-essentials-with-sponge/mattifying-foundation-essentials-sponge.html",

		//"https://fentybeauty.com/products/mattemoiselle-plush-matte-lipstick-freckle-fiesta",
		//"https://fentybeauty.com/products/mattemoiselle-plush-matte-lipstick-shawty",
		//"https://fentybeauty.com/products/fenty-skin-startr-set-us",
		"https://fentybeauty.com/collections/best-sellers",
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
