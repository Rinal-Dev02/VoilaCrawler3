package main

import (
	"bytes"
	"context"
	"encoding/json"
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
	pbProxy "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
)

// _Crawler defined the crawler struct/class for which is not necessory to be exportable
type _Crawler struct {
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
		categoryPathMatcher: regexp.MustCompile(`^/us/(mens|womens)(/[a-zA-Z0-9\-]+){1,6}$`),
		// this regular used to match product page url path
		productPathMatcher: regexp.MustCompile(`^/us/products(/[a-zA-Z0-9\-]+){1,4}\-\d+?$`),
		logger:             logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "bf8f094546d142a2aab6f6cea8c034b4"
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
		EnableHeadless:    false,
		EnableSessionInit: false,
		Reliability:       pbProxy.ProxyReliability_ReliabilityLow,
		MustHeader:        make(http.Header),
	}
	// opts.MustCookies = append(opts.MustCookies,
	// 	&http.Cookie{Name: "language", Value: "en_US"},
	// 	&http.Cookie{Name: "country", Value: "USA"},
	// 	&http.Cookie{Name: "billingCurrency", Value: "USD"},
	// 	&http.Cookie{Name: "saleRegion", Value: "US"},
	// 	&http.Cookie{Name: "MR", Value: "0"},
	// 	&http.Cookie{Name: "loggedIn", Value: "false"},
	// 	&http.Cookie{Name: "_dy_geo", Value: "US.NA.US_CA.US_CA_Los%20Angeles"},
	// )
	return opts
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.laroche-posay.us"}
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
		u.Host = "www.laroche-posay.us"
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

	if p == "" {
		return c.parseCategories(ctx, resp, yield)
	}

	if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	} else if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yield)
	}
	return crawler.ErrUnsupportedPath
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

	sel := dom.Find(`.c-navigation > ul > li`)
	for i := range sel.Nodes {
		node := sel.Eq(i)
		cateName := strings.TrimSpace(node.Find(`.c-navigation__item-title`).Find(`a`).First().Text())
		if cateName == "" {
			continue
		}
		nnctx := context.WithValue(ctx, "Category", cateName)

		subSel := node.Find(`.c-navigation__flyout-element > .m-level-2 > .c-navigation__item`)

		if len(subSel.Nodes) > 0 {

			for j := range subSel.Nodes {

				subNode := subSel.Eq(j)

				subCateName := strings.TrimSpace(subNode.Find(`.c-navigation__item-title`).Find(`a`).First().Text())

				if subCateName == "View all" {
					continue
				}

				subSel1 := subNode.Find(`.m-level-3 > .c-navigation__item`)

				if len(subSel1.Nodes) > 0 {

					for k := range subSel1.Nodes {

						subNode1 := subSel1.Eq(k)

						lastcat := strings.TrimSpace(subNode1.Find(`.c-navigation__item-title`).Find(`a`).First().Text())

						if lastcat == "View all" {
							continue
						}

						href := subNode.Find(`.c-navigation__item-title`).Find(`a`).First().AttrOr("href", "")
						if href == "" {
							continue
						}

						_, err := url.Parse(href)
						if err != nil {
							c.logger.Error("parse url %s failed", href)
							continue
						}

						if subCateName == "" {
							subCateName = lastcat
						} else {
							subCateName = subCateName + " >> " + lastcat
						}

						nnnctx := context.WithValue(nnctx, "SubCategory", subCateName)
						req, _ := http.NewRequest(http.MethodGet, href, nil)
						if err := yield(nnnctx, req); err != nil {
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

// parseCategoryProducts parse api url from web page url
func (c *_Crawler) parseCategoryProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	// read the response data from http response
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Debug(err)
		return err
	}

	lastIndex := nextIndex(ctx)

	sel := doc.Find(`.c-product-tile__name`)

	c.logger.Debugf("nodes %d", len(sel.Nodes))

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

	nextUrl := doc.Find(`.l-plp__pagination > .c-load-more`).Find(`a`).First().AttrOr("href", "")
	if nextUrl == "" {
		return nil
	}

	req, _ := http.NewRequest(http.MethodGet, nextUrl, nil)
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
}

type parseProductResponse struct {
	Context     string `json:"@context"`
	Type        string `json:"@type"`
	Sku         string `json:"sku"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Brand       struct {
		Type string `json:"@type"`
		Name string `json:"name"`
	} `json:"brand"`
	Offers struct {
		Type          string  `json:"@type"`
		Price         float64 `json:"price"`
		PriceCurrency string  `json:"priceCurrency"`
		Availability  string  `json:"availability"`
		URL           string  `json:"url"`
	} `json:"offers"`
	ID string `json:"@id"`
}

var productsReviewExtractReg = regexp.MustCompile(`(?Ums)<script type="application/ld\+json">\s*({.*})\s*</script>`)

// used to trim html labels in description
var htmlTrimRegp = regexp.MustCompile(`</?[^>]+>`)

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

	var viewData parseProductResponse
	matched := productsReviewExtractReg.FindSubmatch([]byte(respBody))
	if len(matched) > 1 {
		if err := json.Unmarshal(matched[1], &viewData); err != nil {
			c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
			return err
		}
	}

	canUrl, _ := c.CanonicalUrl(doc.Find(`link[rel="canonical"]`).AttrOr("href", ""))
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}

	brand := viewData.Brand.Name
	if brand == "" {
		brand = "2422"
	}

	rating, _ := strconv.ParseFloat(doc.Find(`.bvseo-aggregateRatingSection`).Find(`.bvseo-ratingValue`).Text())
	reviewcount, _ := strconv.ParseInt(doc.Find(`.bvseo-aggregateRatingSection`).Find(`.bvseo-reviewCount`).Text())

	currentPrice, _ := strconv.ParsePrice(viewData.Offers.Price)
	msrp, _ := strconv.ParsePrice(viewData.Offers.Price)

	if msrp == 0 {
		msrp = currentPrice
	}
	discount := 0
	if msrp > currentPrice {
		discount = (int)(((msrp - currentPrice) / msrp) * 100)
	}

	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           viewData.Sku,
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
			//GroupId:      doc.Find(`meta[property="product:age_group"]`).AttrOr(`content`, ``),
		},
		BrandName: brand,
		Title:     viewData.Name,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
			Current:  int32(currentPrice * 100),
			Msrp:     int32(msrp) * 100,
			Discount: int32(discount),
		},
		Stats: &pbItem.Stats{
			ReviewCount: int32(reviewcount),
			Rating:      float32(rating),
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
	}

	item.Description = viewData.Description

	if strings.Contains(viewData.Offers.Availability, "https://schema.org/InStock") {
		item.Stock.StockStatus = pbItem.Stock_InStock
	}

	//images
	sel := doc.Find(`.c-product-detail-image__image-link`)
	for j := range sel.Nodes {
		node := sel.Eq(j)
		imgurl := strings.Split(node.Find(`img`).AttrOr(`src`, ``), "?")[0]

		item.Medias = append(item.Medias, pbMedia.NewImageMedia(
			strconv.Format(j),
			strings.Replace(imgurl, "?sw=68&sfrm=jpg&q=70", "?sw=1024&sfrm=jpg&q=80", 1),
			strings.Replace(imgurl, "?sw=68&sfrm=jpg&q=70", "?sw=800&sfrm=jpg&q=80", 1),
			strings.Replace(imgurl, "?sw=68&sfrm=jpg&q=70", "?sw=500&sfrm=jpg&q=80", 1),
			strings.Replace(imgurl, "?sw=68&sfrm=jpg&q=70", "?sw=300&sfrm=jpg&q=80", 1),
			"", j == 0))
	}

	// itemListElement
	sel = doc.Find(`.c-breadcrumbs > ol > li`)
	c.logger.Debugf("nodes %d", len(sel.Nodes))
	for i := range sel.Nodes {
		node := sel.Eq(i).Find(`span`)
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

	// Color
	cid := ""
	// colorName := ""
	// var colorSelected *pbItem.SkuSpecOption
	// sel = doc.Find(`.attribute > .select-color`).Find(`button`)
	// for i := range sel.Nodes {
	// 	node := sel.Eq(i)

	// 	if strings.Contains(node.AttrOr(`class`, ``), `selected`) {
	// 		//cid = node.AttrOr(`data-variationgroupid`, "")
	// 		icon := node.Find(`img`).AttrOr(`src`, ``)
	// 		colorName = node.AttrOr(`data-color-text`, "")
	// 		colorSelected = &pbItem.SkuSpecOption{
	// 			Type:  pbItem.SkuSpecType_SkuSpecColor,
	// 			Id:    cid,
	// 			Name:  colorName,
	// 			Value: colorName,
	// 			Icon:  icon,
	// 		}
	// 	}
	// }

	sel = doc.Find(`.c-carousel__inner`).Find(`.c-carousel__content`).Find(`.c-carousel__item`)
	for i := range sel.Nodes {
		node := sel.Eq(i)

		if strings.Contains(node.Find(`a`).AttrOr(`class`, ``), `m-selected`) {

			sid := (node.Find(`a`).AttrOr("data-js-pid", ""))
			if sid == "" {
				continue
			}

			sName := (node.Find(`a`).AttrOr("data-js-value", ""))

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

			sku.Stock.StockStatus = pbItem.Stock_InStock

			// if colorSelected != nil {
			// 	sku.Specs = append(sku.Specs, colorSelected)
			// }

			// size
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    sid,
				Name:  sName,
				Value: sName,
			})

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

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		"https://www.laroche-posay.us/",
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
