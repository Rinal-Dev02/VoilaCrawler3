package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
		categoryPathMatcher: regexp.MustCompile(`^(/([/A-Za-z0-9_-]+)/br/v=1/\d+.htm)|(/products)$`),
		productPathMatcher:  regexp.MustCompile(`^/([/A-Za-z0-9_-]+)/vp/v=1/\d+.htm$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "5d472ca532ab4db798d0b6cfabc753c0"
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
		MustCookies: []*http.Cookie{
			{Name: "llc", Value: "US-EN-USD", Path: "/"},
		},
		MustHeader: crawler.NewCrawlOptions().MustHeader,
	}

	return opts
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.eastdane.com"}
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
		u.Host = "www.eastdane.com"
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
	if p == "" {
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

func (c *_Crawler) GetCategories(ctx context.Context) ([]*pbItem.Category, error) {
	req, _ := http.NewRequest(http.MethodGet, "http://www.eastdane.com/", nil)
	opts := c.CrawlOptions(req.URL)
	resp, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
		EnableProxy:       true,
		EnableHeadless:    opts.EnableHeadless,
		EnableSessionInit: opts.EnableSessionInit,
		Reliability:       opts.Reliability,
	})
	if err != nil {
		c.logger.Error(err)
		return nil, err
	}

	dom, err := resp.Selector()
	if err != nil {
		c.logger.Error(err)
		return nil, err
	}

	var (
		cates   []*pbItem.Category
		cateMap = map[string]*pbItem.Category{}
	)
	if err := func(yield func(names []string, url string) error) error {

		sel := dom.Find(`#categories>li`)

		for i := range sel.Nodes {
			node := sel.Eq(i)
			cateName := strings.TrimSpace(node.Find(`a>span`).First().Text())

			if cateName == "" {
				continue
			}

			subSel := node.Find(`.nested-navigation-section`)
			for k := range subSel.Nodes {
				subNode2 := subSel.Eq(k)

				subcat2 := strings.TrimSpace(subNode2.Find(`.sub-navigation-header`).First().Text())
				if subcat2 == "" {
					subcat2 = strings.TrimSpace(subNode2.Find(`.sub-navigation-list-item-image-text`).First().Text())
				}

				subNode2list := subNode2.Find(`.sub-navigation-list>li`)
				for j := range subNode2list.Nodes {
					subNode4 := subNode2list.Eq(j)
					subcat3 := strings.TrimSpace(subNode4.Find(`.sub-navigation-list-item-link-text`).First().Text())

					if subcat3 == "" {
						subcat3 = strings.TrimSpace(subNode4.Find(`.sub-navigation-list-item-cta`).First().Text())
					}
					href := subNode4.Find(`a`).AttrOr("href", "")
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
						if err := yield([]string{cateName, subcat2, subcat3}, canonicalhref); err != nil {
							return err
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

type categoryProductStructure struct {
	Metadata struct {
		TotalProductCount   int    `json:"totalProductCount"`
		InStockProductCount int    `json:"inStockProductCount"`
		ProcessingTime      int    `json:"processingTime"`
		FilterContext       string `json:"filterContext"`
		ShowRatings         bool   `json:"showRatings"`
		WeddingBoutique     bool   `json:"weddingBoutique"`
	} `json:"metadata"`
	Products []struct {
		ProductID         int    `json:"productId"`
		ProductDetailLink string `json:"productDetailLink"`
		QuickShopLink     string `json:"quickShopLink"`
	} `json:"products"`
}

var categoryProductReg = regexp.MustCompile(`filters\s*=\s*({.*});`)

// parseCategoryProducts parse api url from web page url

func (c *_Crawler) parseCategoryProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}
	respBody, err := resp.RawBody()
	if err != nil {
		return err
	}

	matched := categoryProductReg.FindSubmatch([]byte(respBody))
	if len(matched) <= 1 {
		matched = append(matched, []byte(``))
		matched = append(matched, []byte(respBody))
		if len(matched) <= 1 {
			c.logger.Debugf("%s", respBody)
			return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
		}
	}

	var viewData categoryProductStructure
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		return err
	}

	lastIndex := nextIndex(ctx)

	for _, rawcat := range viewData.Products {

		href := rawcat.ProductDetailLink
		if href == "" {
			continue
		}

		req, err := http.NewRequest(http.MethodGet, href, nil)
		if err != nil {
			c.logger.Error(err)
			continue
		}
		nctx := context.WithValue(ctx, "item.index", lastIndex)
		lastIndex += 1
		if err := yield(nctx, req); err != nil {
			return err
		}

	}

	nextURL := "https://www.eastdane.com/products?filter&department=19210&filterContext=19210&baseIndex=40"

	totalCount := viewData.Metadata.TotalProductCount

	if lastIndex >= (int)(totalCount) {
		return nil
	}

	// set pagination
	u, _ := url.Parse(nextURL)
	vals := u.Query()
	vals.Set("department", viewData.Metadata.FilterContext)
	vals.Set("filterContext", viewData.Metadata.FilterContext)
	vals.Set("baseIndex", strconv.Format(lastIndex-1))
	vals.Set("limit", strconv.Format(100))
	u.RawQuery = vals.Encode()

	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)

	nctx := context.WithValue(ctx, "item.index", lastIndex-1)
	return yield(nctx, req)

}

func TrimSpaceNewlineInString(s string) string {
	re := regexp.MustCompile(`\n`)
	resp := re.ReplaceAllString(s, " ")
	resp = strings.ReplaceAll(resp, "\\n", " ")
	resp = strings.ReplaceAll(resp, "\r", " ")
	resp = strings.ReplaceAll(resp, "\t", " ")
	re = regexp.MustCompile(`\s+`)
	resp = re.ReplaceAllString(resp, " ")
	return resp
}

// used to trim html labels in description
var htmlTrimRegp = regexp.MustCompile(`</?[^>]+>`)

var productProductReg = regexp.MustCompile(`<script\s*type="sui-state"\s*data-key="pdp\.state">\s*({.*})</script>`)

type productStructure struct {
	Product struct {
		Sin           int           `json:"sin"`
		StyleNumber   string        `json:"styleNumber"`
		BrandCode     string        `json:"brandCode"`
		BrandLabel    string        `json:"brandLabel"`
		VendorName    string        `json:"vendorName"`
		PromotionTags []interface{} `json:"promotionTags"`
		Hearted       bool          `json:"hearted"`
		MyDesigner    bool          `json:"myDesigner"`
		StyleColors   []struct {
			Prices []struct {
				RetailAmount        float64 `json:"retailAmount"`
				SaleAmount          float64 `json:"saleAmount"`
				RetailDisplayAmount string  `json:"retailDisplayAmount"`
				SaleDisplayAmount   string  `json:"saleDisplayAmount"`
				OnSale              bool    `json:"onSale"`
				OnFinalSale         bool    `json:"onFinalSale"`
				SalePercentage      float64 `json:"salePercentage"`
				CurrencyCode        string  `json:"currencyCode"`
				ForeignCurrency     bool    `json:"foreignCurrency"`
				DisplayAmount       string  `json:"displayAmount"`
			} `json:"prices"`
			StyleColorSizes []struct {
				Sin     string `json:"sin"`
				SkuCode string `json:"skuCode"`
				Size    struct {
					Code     string `json:"code"`
					Label    string `json:"label"`
					Priority int    `json:"priority"`
				} `json:"size"`
				InStock bool `json:"inStock"`
			} `json:"styleColorSizes"`
			Color struct {
				Code  string `json:"code"`
				Label string `json:"label"`
			} `json:"color"`
			Images []struct {
				URL string `json:"url"`
			} `json:"images"`
			SwatchImage struct {
				URL string `json:"url"`
			} `json:"swatchImage"`
			StyleColorCode        string        `json:"styleColorCode"`
			ModelSize             interface{}   `json:"modelSize"`
			SupportedVideos       []interface{} `json:"supportedVideos"`
			InStock               bool          `json:"inStock"`
			DefaultStyleColorSize interface{}   `json:"defaultStyleColorSize"`
		} `json:"styleColors"`
		CustomerReviewSummary interface{} `json:"customerReviewSummary"`
		Orderable             bool        `json:"orderable"`
		OneBy                 bool        `json:"oneBy"`
		Division              string      `json:"division"`
		Classification        string      `json:"classification"`
		DetailPageURL         string      `json:"detailPageUrl"`
		SiteEligibilityList   []string    `json:"siteEligibilityList"`
		Department            string      `json:"department"`
		Gender                string      `json:"gender"`
		ShortDescription      string      `json:"shortDescription"`
		LongDescription       string      `json:"longDescription"`
		SizeAndFitDetail      struct {
			SizeAndFitDescription string `json:"sizeAndFitDescription"`
		} `json:"sizeAndFitDetail"`
	} `json:"product"`
}

var imgregx = regexp.MustCompile(`_UX\d+_`)

// parseProduct
func (c *_Crawler) parseProduct(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil {
		return nil
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error(err)
		return err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	matched := productProductReg.FindSubmatch([]byte(respBody))
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}

	var viewData productStructure
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		return err
	}

	canUrl, _ := c.CanonicalUrl(doc.Find(`link[rel="canonical"]`).AttrOr("href", ""))
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}

	desc := htmlTrimRegp.ReplaceAllString(viewData.Product.LongDescription+" "+viewData.Product.SizeAndFitDetail.SizeAndFitDescription, " ")

	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           strconv.Format(viewData.Product.Sin),
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
		},
		Title:       viewData.Product.ShortDescription,
		Description: desc,
		BrandName:   viewData.Product.BrandLabel,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
	}

	sel := doc.Find(`li[itemprop="itemListElement"]`)
	for i := range sel.Nodes {
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

	for _, rawsku := range viewData.Product.StyleColors {

		current := 0.0
		msrp := 0.0
		discount := 0.0
		for _, rawprice := range rawsku.Prices {
			current, _ = strconv.ParsePrice(rawprice.SaleAmount)
			msrp, _ = strconv.ParsePrice(rawprice.RetailAmount)
			break
		}

		if msrp == 0.0 {
			msrp = current
		}
		if msrp > current {
			discount = ((msrp - current) / msrp) * 100
		}

		var medias []*pbMedia.Media
		for m, mid := range rawsku.Images {
			template := strings.Split(mid.URL, "?")[0]

			medias = append(medias, pbMedia.NewImageMedia(
				strconv.Format(m),
				template,
				imgregx.ReplaceAllString(template, "_UX1000_"),
				imgregx.ReplaceAllString(template, "_UX800_"),
				imgregx.ReplaceAllString(template, "_UX500_"),
				"",
				m == 0,
			))
		}

		var colorSelected = &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecColor,
			Id:    rawsku.Color.Code,
			Name:  rawsku.Color.Label,
			Value: rawsku.Color.Label,
			Icon:  rawsku.SwatchImage.URL,
		}

		for _, rawsize := range rawsku.StyleColorSizes {

			sku := pbItem.Sku{
				SourceId: strconv.Format(rawsize.SkuCode),
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(current * 100),
					Msrp:     int32(msrp * 100),
					Discount: int32(discount),
				},
				Medias: medias,
				Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}

			if rawsize.InStock {
				sku.Stock.StockStatus = pbItem.Stock_InStock
				item.Stock.StockStatus = pbItem.Stock_InStock
			}

			if colorSelected != nil {
				sku.Specs = append(sku.Specs, colorSelected)
			}

			if rawsize.Size.Label != "" {
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecSize,
					Id:    rawsize.Size.Code,
					Name:  rawsize.Size.Label,
					Value: rawsize.Size.Label,
				})
			}

			// for _, spec := range sku.Specs {
			// 	sku.SourceId += fmt.Sprintf("-%s", spec.Id)
			// }

			item.SkuItems = append(item.SkuItems, &sku)
		}

		if len(rawsku.StyleColorSizes) == 0 {

			sku := pbItem.Sku{
				SourceId: strconv.Format(rawsku.StyleColorCode),
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(current * 100),
					Msrp:     int32(msrp * 100),
					Discount: int32(discount),
				},
				Medias: medias,
				Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}

			if rawsku.InStock {
				sku.Stock.StockStatus = pbItem.Stock_InStock
				item.Stock.StockStatus = pbItem.Stock_InStock
			}

			if colorSelected != nil {
				sku.Specs = append(sku.Specs, colorSelected)
			}

			if len(sku.Specs) == 0 {
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecColor,
					Id:    "-",
					Name:  "-",
					Value: "-",
				})
			}
			item.SkuItems = append(item.SkuItems, &sku)
		}
	}

	if err = yield(ctx, &item); err != nil {
		return err
	}

	return nil
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://www.eastdane.com/",
		//"https://www.eastdane.com/clothing-pants/br/v=1/19210.htm",
		//"https://www.eastdane.com/men-accessories-home-gifts/br/v=1/19226.htm",
		//"https://www.eastdane.com/ami-coeur-sweater/vp/v=1/1527786695.htm",
		//"https://www.eastdane.com/midweight-terry-relaxed-sweatpant-reigning/vp/v=1/1501978273.htm",
		//"https://www.eastdane.com/bandwagon-sunglasses-le-specs/vp/v=1/1540691449.htm?",
		//"https://www.eastdane.com/blouson-alex-apc/vp/v=1/1553421893.htm",

		//"https://www.eastdane.com/men-shops-activewear/br/v=1/47430.htm",
		//"https://www.eastdane.com/midweight-terry-relaxed-sweatpant-reigning/vp/v=1/1501978273.htm",
		"https://www.eastdane.com/hitch-backpack-coach-new-york/vp/v=1/1534138121.htm",
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
