package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/voiladev/VoilaCrawl/pkg/crawler"
	"github.com/voiladev/VoilaCrawl/pkg/net/http"
	"github.com/voiladev/VoilaCrawl/pkg/net/http/cookiejar"
	"github.com/voiladev/VoilaCrawl/pkg/proxy"
	pbMedia "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
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
func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient: client,
		// this regular used to match category page url path
		categoryPathMatcher: regexp.MustCompile(`^(/[a-z0-9\-]+){1,6}$`),
		// this regular used to match product page url path
		productPathMatcher: regexp.MustCompile(`^(/[a-z0-9\-]+){1,4}/\d+\.html$`),
		logger:             logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "f125b86f832f492db2676b739aeefa84"
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
	return []string{"*.modcloth.com"}
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

// below are the golang json data struct of raw website.
// if you get the raw json data of the website,
// then you can use https://mholt.github.io/json-to-go/ to convert it to a golang struct

type RawProductDetails struct {
	Context     string `json:"@context"`
	Type        string `json:"@type"`
	ProductID   string `json:"productID"`
	Description string `json:"description"`
	Name        string `json:"name"`
	Image       string `json:"image"`
	URL         string `json:"url"`
	Brand       struct {
		Type string `json:"@type"`
		Name string `json:"name"`
	} `json:"brand"`
	Offers struct {
		Type            string    `json:"@type"`
		Availability    string    `json:"availability"`
		Price           float64   `json:"price"`
		PriceCurrency   string    `json:"priceCurrency"`
		URL             string    `json:"url"`
		PriceValidUntil time.Time `json:"priceValidUntil"`
	} `json:"offers"`
	AggregateRating struct {
		Type        string  `json:"@type"`
		RatingValue float64 `json:"ratingValue"`
		ReviewCount int     `json:"reviewCount"`
	} `json:"aggregateRating"`
}

type RawProductVariationDetails []struct {
	VariationGroupID string `json:"variationGroupID"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	URL              string `json:"url"`
	ProductVariants  []struct {
		Upc            string `json:"upc"`
		Size           string `json:"size"`
		UnitsAvailable int    `json:"units_available"`
		Archived       bool   `json:"archived"`
		Online         bool   `json:"online"`
	} `json:"product_variants"`
	IsSelected     bool    `json:"isSelected"`
	Archived       bool    `json:"archived"`
	Online         bool    `json:"online"`
	ReviewsCount   int     `json:"reviewsCount"`
	ReviewsRanking float64 `json:"reviewsRanking"`
	UnitsAvailable int     `json:"units_available"`
}

type RawProductOtherDetails struct {
	PageName               string        `json:"page_name"`
	PageType               string        `json:"page_type"`
	PageSubtype            string        `json:"page_subtype"`
	PageContextType        string        `json:"page_context_type"`
	PageContextTitle       string        `json:"page_context_title"`
	ABTestVariant          string        `json:"a_b_test_variant"`
	PageURL                string        `json:"page_url"`
	NumItemsInCart         string        `json:"num_items_in_cart"`
	UserAnonymous          string        `json:"user_anonymous"`
	UserAuthenticated      string        `json:"user_authenticated"`
	CustomerLoggedInStatus string        `json:"customer_logged_in_status"`
	CustomerGroup          string        `json:"customer_group"`
	CustomerLovedItems     []interface{} `json:"customer_loved_items"`
	UserRegistered         string        `json:"user_registered"`
	AccountID              string        `json:"account_id"`
	CustomerType           string        `json:"customer_type"`
	VisitNumber            int           `json:"visit_number"`
	OrderCount             interface{}   `json:"order_count"`
	CountryCode            string        `json:"country_code"`
	LanguageCode           string        `json:"language_code"`
	ProductCategory        []string      `json:"product_category"`
	ProductSubcategory     []string      `json:"product_subcategory"`
	ProductOriginalPrice   []string      `json:"product_original_price"`
	ProductUnitPrice       []string      `json:"product_unit_price"`
	ProductID              []string      `json:"product_id"`
	MasterGroupID          []string      `json:"master_group_id"`
	ProductName            []string      `json:"product_name"`
	ProductBrand           []string      `json:"product_brand"`
	ProductColor           []string      `json:"product_color"`
	ProductSku             []string      `json:"product_sku"`
	ProductImgURL          []string      `json:"product_img_url"`
	ProductRating          string        `json:"product_rating"`
	SiteFormat             string        `json:"site_format"`
	SiteSection            string        `json:"site_section"`
	NewCustomer            interface{}   `json:"new_customer"`
	HasOrders              interface{}   `json:"has_orders"`
	SessionCurrency        string        `json:"session_currency"`
}

type CategoriesView struct {
	PageName               string        `json:"page_name"`
	PageType               string        `json:"page_type"`
	PageSubtype            string        `json:"page_subtype"`
	PageContextType        string        `json:"page_context_type"`
	PageContextTitle       string        `json:"page_context_title"`
	ABTestVariant          string        `json:"a_b_test_variant"`
	PageURL                string        `json:"page_url"`
	NumItemsInCart         string        `json:"num_items_in_cart"`
	UserAnonymous          string        `json:"user_anonymous"`
	UserAuthenticated      string        `json:"user_authenticated"`
	CustomerLoggedInStatus string        `json:"customer_logged_in_status"`
	CustomerGroup          string        `json:"customer_group"`
	CustomerLovedItems     []interface{} `json:"customer_loved_items"`
	UserRegistered         string        `json:"user_registered"`
	AccountID              string        `json:"account_id"`
	CustomerType           string        `json:"customer_type"`
	VisitNumber            int           `json:"visit_number"`
	OrderCount             interface{}   `json:"order_count"`
	CountryCode            string        `json:"country_code"`
	LanguageCode           string        `json:"language_code"`
	ProductID              []string      `json:"product_id"`
	CategoryID             string        `json:"category_id"`
	PageTemplate           string        `json:"page_template"`
	CategorySort           string        `json:"category_sort"`
	PageNumber             string        `json:"page_number"`
	ProductCategory        []string      `json:"product_category"`
	ProductSubcategory     []string      `json:"product_subcategory"`
	PageCategory           string        `json:"page_category"`
	SiteFormat             string        `json:"site_format"`
	SiteSection            string        `json:"site_section"`
	NewCustomer            interface{}   `json:"new_customer"`
	HasOrders              interface{}   `json:"has_orders"`
	SessionCurrency        string        `json:"session_currency"`
}

func TrimSpaceNewlineInByte(s []byte) []byte {
	re := regexp.MustCompile(`\n`)
	return re.ReplaceAll(s, []byte(" "))
}

// parseCategoryProducts parse api url from web page url
func (c *_Crawler) parseCategoryProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	// read the response data from http response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Debug(err)
		return err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Debug(err)
		return err
	}

	lastIndex := nextIndex(ctx)
	sel := doc.Find(`#search-result-items .grid-tile`)
	for i := range sel.Nodes {
		node := sel.Eq(i)
		href := node.Find(`.product-tile>.product-name>.name-link`).AttrOr("href", "")
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
			c.logger.Error(err)
			return err
		}
	}

	// set pagination
	pageSel := doc.Find(`.search-result-options.bottom .pagination li.first-last.right-arrow`)
	if strings.Contains(pageSel.AttrOr("class", ""), "disabled") {
		return nil
	}
	href := pageSel.Find(".page-next").AttrOr("href", "")
	if href == "" {
		return nil
	}

	req, _ := http.NewRequest(http.MethodGet, href, nil)
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
}

// used to extract embaded json data in website page.
// more about golang regulation see here https://golang.org/pkg/regexp/syntax/
var productsExtractReg = regexp.MustCompile(`(?Us)var\s*utag_data\s*=\s*({.*});`)
var productDataVariationExtractReg = regexp.MustCompile(`(?Us)mc_global.product\s*=\s*(\[.*\]);`)
var productDateMainExtractReg = regexp.MustCompile(`(?Us)<script\s*type="application/ld\+json">\s*({.*})\s*</script>`)

// parseProduct
func (c *_Crawler) parseProduct(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil {
		return nil
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	matched := productsExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract product info from %s failed, error=%s", resp.Request.URL, err)
	}
	// c.logger.Debugf("data: %s", matched[1])

	var (
		viewData      RawProductDetails
		imgData       RawProductOtherDetails
		variationData RawProductVariationDetails
	)
	if err := json.Unmarshal(matched[1], &imgData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}

	matched1 := productDateMainExtractReg.FindAllSubmatch(respBody, -1)
	if len(matched1) <= 1 {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}
	if err := json.Unmarshal(matched1[1][1], &viewData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}

	matched = productDataVariationExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}
	if err := json.Unmarshal(matched[1], &variationData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}

	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:       strconv.Format(viewData.ProductID),
			CrawlUrl: resp.Request.URL.String(),
		},
		BrandName:   viewData.Brand.Name,
		Title:       viewData.Name,
		Description: viewData.Description,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stats: &pbItem.Stats{
			ReviewCount: int32(viewData.AggregateRating.ReviewCount),
			Rating:      float32(viewData.AggregateRating.RatingValue),
		},
	}
	if len(imgData.MasterGroupID) > 0 {
		item.Source.GroupId = imgData.MasterGroupID[0]
	}
	if len(imgData.ProductCategory) > 0 {
		item.Category = imgData.ProductCategory[0]
	}
	if len(imgData.ProductSubcategory) > 0 {
		item.SubCategory = imgData.ProductSubcategory[0]
	}

	var colorSpec *pbItem.SkuSpecOption
	if len(imgData.ProductColor) > 0 {
		colorSpec = &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecColor,
			Id:    imgData.ProductColor[0],
			Name:  imgData.ProductColor[0],
			Value: imgData.ProductColor[0],
		}
	}

	var medias []*pbMedia.Media
	for m, mid := range imgData.ProductImgURL {
		s := strings.Split(mid, "?")
		medias = append(medias, pbMedia.NewImageMedia(
			strconv.Format(m),
			s[0],
			s[0]+"?sw=913&sm=fit",
			s[0]+"?sw=600&sm=fit",
			s[0]+"?sw=450&sm=fit",
			"",
			m == 0,
		))
	}

	for _, rawSku := range variationData {
		if rawSku.VariationGroupID == viewData.ProductID {
			var (
				current, msrp, discount float64
			)
			if len(imgData.ProductOriginalPrice) > 0 {
				msrp, _ = strconv.ParseFloat(imgData.ProductOriginalPrice[0])
			}
			if len(imgData.ProductUnitPrice) > 0 {
				current, _ = strconv.ParseFloat(imgData.ProductUnitPrice[0])
			}
			discount = math.Ceil((msrp - current) / msrp * 100)

			for _, rawVariation := range rawSku.ProductVariants {
				sku := pbItem.Sku{
					SourceId: rawVariation.Upc,
					Price: &pbItem.Price{
						Currency: regulation.Currency_USD,
						Current:  int32(viewData.Offers.Price * 100),
						Msrp:     int32(viewData.Offers.Price * 100),
						Discount: int32(discount),
					},
					Medias: medias,
					Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
				}
				if rawVariation.UnitsAvailable > 0 {
					sku.Stock.StockStatus = pbItem.Stock_InStock
					sku.Stock.StockCount = int32(rawVariation.UnitsAvailable)
				}

				if colorSpec != nil {
					sku.Specs = append(sku.Specs, colorSpec)
				}
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecSize,
					Id:    rawVariation.Upc,
					Name:  rawVariation.Size,
					Value: rawVariation.Size,
				})
				item.SkuItems = append(item.SkuItems, &sku)
			}
		} else {
			req, err := http.NewRequest(http.MethodGet, rawSku.URL, nil)
			if err != nil {
				c.logger.Error(err)
				continue
			}
			if err := yield(ctx, req); err != nil {
				c.logger.Error(err)
				return err
			}
		}
	}
	if len(item.SkuItems) > 0 {
		if err = yield(ctx, &item); err != nil {
			return err
		}
	} else {
		return errors.New("not sku found")
	}
	return nil
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		// "https://www.modcloth.com/shop/best-selling-shoes",
		"https://www.modcloth.com/shop/shoes/t-u-k-the-zest-is-history-heel-in-burgundy/128047.html",
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
	logger := glog.New(glog.LogLevelDebug)
	// build a http client
	// get proxy's microservice address from env
	client, err := proxy.NewProxyClient(os.Getenv("VOILA_PROXY_URL"), cookiejar.New(), logger)
	if err != nil {
		panic(err)
	}

	// instance the spider locally
	spider, err := New(client, logger)
	if err != nil {
		panic(err)
	}

	reqFilter := map[string]struct{}{}

	// this callback func is used to do recursion call of sub requests.
	var callback func(ctx context.Context, val interface{}) error
	callback = func(ctx context.Context, val interface{}) error {
		switch i := val.(type) {
		case *http.Request:
			if _, ok := reqFilter[i.URL.String()]; ok {
				return nil
			}
			reqFilter[i.URL.String()] = struct{}{}

			logger.Debugf("Access %s", i.URL)

			// crawler := spider.(*_Crawler)
			// if crawler.productPathMatcher.MatchString(i.URL.Path) {
			// 	return nil
			// }

			opts := spider.CrawlOptions(i.URL)

			// process logic of sub request

			// init custom headers
			for k := range opts.MustHeader {
				i.Header.Set(k, opts.MustHeader.Get(k))
			}

			// init custom cookies
			for _, c := range opts.MustCookies {
				if strings.HasPrefix(i.URL.Path, c.Path) || c.Path == "" {
					val := fmt.Sprintf("%s=%s", c.Name, c.Value)
					if c := i.Header.Get("Cookie"); c != "" {
						i.Header.Set("Cookie", c+"; "+val)
					} else {
						i.Header.Set("Cookie", val)
					}
				}
			}

			// set scheme,host for sub requests. for the product url in category page is just the path without hosts info.
			// here is just the test logic. when run the spider online, the controller will process automatically
			if i.URL.Scheme == "" {
				i.URL.Scheme = "https"
			}
			if i.URL.Host == "" {
				i.URL.Host = "www.modcloth.com"
			}

			// do http requests here.
			nctx, cancel := context.WithTimeout(ctx, time.Minute*5)
			defer cancel()
			resp, err := client.DoWithOptions(nctx, i, http.Options{
				EnableProxy:       true,
				EnableHeadless:    false,
				EnableSessionInit: opts.EnableSessionInit,
				KeepSession:       opts.KeepSession,
				Reliability:       opts.Reliability,
			})
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			return spider.Parse(ctx, resp, callback)
		default:
			// output the result
			data, err := protojson.Marshal(i.(proto.Message))
			if err != nil {
				return err
			}
			logger.Infof("data: %s", data)
		}
		return nil
	}

	ctx := context.WithValue(context.Background(), "tracing_id", fmt.Sprintf("tracing_%d", time.Now().UnixNano()))
	// start the crawl request
	for _, req := range spider.NewTestRequest(context.Background()) {
		if err := callback(ctx, req); err != nil {
			logger.Fatal(err)
		}
	}
}
