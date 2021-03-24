package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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
		categoryPathMatcher: regexp.MustCompile(`^(/[a-z0-9_-]+)?/en-us/(women|men)(/[a-z0-9_-]+)+$`),
		// this regular used to match product page url path
		productPathMatcher: regexp.MustCompile(`^(.*)(\d)$`),
		logger:             logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "eac744e6d5bc4515bca921d2e4723119"
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
func (c *_Crawler) CrawlOptions() *crawler.CrawlOptions {
	return &crawler.CrawlOptions{
		EnableHeadless: false,
		// use js api to init session for the first request of the crawl
		EnableSessionInit: true,
	}
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.24s.com"}
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

	if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	} else if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yield)
	}
	return fmt.Errorf("unsupported url %s", resp.Request.URL.String())
}

func isRobotCheckPage(respBody []byte) bool {
	return bytes.Contains(respBody, []byte("we believe you are using automation tools to browse the website")) ||
		bytes.Contains(respBody, []byte("Javascript is disabled or blocked by an extension")) ||
		bytes.Contains(respBody, []byte("Your browser does not support cookies"))
}

// nextIndex used to get the index from the shared data.
// item.index is a const key for item index.
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")))
}

// below are the golang json data struct of raw website.
// if you get the raw json data of the website,
// then you can use https://mholt.github.io/json-to-go/ to convert it to a golang struct

type CategoryView struct {
	OperatingCountryCode string `json:"operatingCountryCode"`
}

// used to extract embaded json data in website page.
// more about golang regulation see here https://golang.org/pkg/regexp/syntax/
var productsExtractReg = regexp.MustCompile(`(?U)<script\s*>window.__INITIAL_CONFIG__\s*=\s*({.*});?\s*</script>`)
var productDetailExtractReg = regexp.MustCompile(`(?U)type="application/json" crossorigin="anonymous">\s*({.*})\s*</script>`)

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

	if isRobotCheckPage(respBody) {
		return errors.New("robot check page")
	}

	if !bytes.Contains(respBody, []byte("<a class=\"item\"")) {
		return errors.New("products not found")
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	lastIndex := nextIndex(ctx)
	sel := doc.Find(`a.item`)
	c.logger.Debugf("nodes %d", len(sel.Nodes))
	for i := range sel.Nodes {
		node := sel.Eq(i)
		if href, _ := node.Attr("href"); href != "" {
			//c.logger.Debugf("yield %v-->%s", lastIndex, href)
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

	if !bytes.Contains(respBody, []byte("<link rel=\"next\" href=")) {
		// nextpage not found
		return nil
	}

	// get current page number
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}

	// set pagination
	u := *resp.Request.URL
	vals := u.Query()
	vals.Set("page", strconv.Format(page+1))
	u.RawQuery = vals.Encode()

	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	// update the index of last page
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
}

type parseProductResponse struct {
	Props struct {
		InitialState struct {
			Pdp struct {
				Product string `json:"product"`
				Sizes   struct {
					References []string `json:"references"`
					Mapping    map[string][]struct {
						//WI038 []struct {
						Reference string `json:"reference"`
						Label     string `json:"label"`
						Stock     bool   `json:"stock"`
						//} `json:"WI038"`
					} `json:"mapping"`
				} `json:"sizes"`
				SelectedSize    interface{} `json:"selectedSize"`
				SelectedColor   string      `json:"selectedColor"`
				ProductFormated struct {
					Upsell []struct {
						PriceInclVat int         `json:"priceInclVat"`
						PriceExclVat interface{} `json:"priceExclVat"`
						Pictures     []struct {
							Priority int    `json:"priority"`
							Path     string `json:"path"`
						} `json:"pictures"`
						OfferID        string      `json:"offerId"`
						New            bool        `json:"new"`
						Name           string      `json:"name"`
						Model          string      `json:"model"`
						LongSKU        string      `json:"longSKU"`
						Exclusive      bool        `json:"exclusive"`
						DiscountPrice  interface{} `json:"discountPrice"`
						DiscountAmount int         `json:"discountAmount"`
						Brand          struct {
							Slug string `json:"slug"`
							Name string `json:"name"`
						} `json:"brand"`
						DiscountPriceInclVat int         `json:"discountPriceInclVat"`
						DiscountPriceExclVat interface{} `json:"discountPriceExclVat"`
					} `json:"upsell"`
					SizeAvailable []struct {
						SizeCode      string `json:"sizeCode"`
						LongSKU       string `json:"longSKU"`
						HasOffer      bool   `json:"hasOffer"`
						SellerSKU     string `json:"sellerSKU,omitempty"`
						Stock         int    `json:"stock,omitempty"`
						Replenishment bool   `json:"replenishment,omitempty"`
					} `json:"sizeAvailable"`
					ShippingExpress []struct {
						PriceInclVat         int `json:"priceInclVat"`
						DiscountAmount       int `json:"discountAmount"`
						DiscountPriceInclVat int `json:"discountPriceInclVat"`
					} `json:"shippingExpress"`
					ProductInformation struct {
						Year                string      `json:"year"`
						SkinType            interface{} `json:"skinType"`
						Season              string      `json:"season"`
						ProductDetails      interface{} `json:"productDetails"`
						ProductCode         string      `json:"productCode"`
						Preview             bool        `json:"preview"`
						PreferentialOrigin  interface{} `json:"preferentialOrigin"`
						Packaging           interface{} `json:"packaging"`
						ManufacturerID      string      `json:"manufacturerId"`
						ManufacturerAddress interface{} `json:"manufacturerAddress"`
						MadeInLabel         string      `json:"madeInLabel"`
						MadeIn              string      `json:"madeIn"`
						HeelSize            string      `json:"heelSize"`
						FacetColorCode      string      `json:"facetColorCode"`
						FacetColor          string      `json:"facetColor"`
						DimensionsWidth     interface{} `json:"dimensionsWidth"`
						DimensionsHeight    interface{} `json:"dimensionsHeight"`
						DimensionsDepth     interface{} `json:"dimensionsDepth"`
						Dimensions          string      `json:"dimensions"`
						CompositionEn       string      `json:"compositionEn"`
						Composition         interface{} `json:"composition"`
						Collection          string      `json:"collection"`
						CareInstructions    interface{} `json:"careInstructions"`
						CapacityWeight      interface{} `json:"capacityWeight"`
						CapacityLiter       interface{} `json:"capacityLiter"`
						BrandInformation    string      `json:"brandInformation"`
						BrandColorFront     interface{} `json:"brandColorFront"`
						BrandColor          string      `json:"brandColor"`
						Avacode             string      `json:"avacode"`
						DisplayMadeIn       bool        `json:"displayMadeIn"`
						Size                string      `json:"size"`
						SizingChart         string      `json:"sizingChart"`
					} `json:"productInformation"`
					ProductFamily string `json:"productFamily"`
					Pictures      []struct {
						Priority int    `json:"priority"`
						Path     string `json:"path"`
					} `json:"pictures"`
					OfferID             string      `json:"offerId"`
					Name                string      `json:"name"`
					MainCategoryBrand   string      `json:"mainCategoryBrand"`
					LongSKU             string      `json:"longSKU"`
					FacetColorAvailable interface{} `json:"facetColorAvailable"`
					Destination         string      `json:"destination"`
					Description         string      `json:"description"`
					Cites               string      `json:"cites"`
					CategoriesMaster    []string    `json:"categoriesMaster"`
					CategoriesBrands    []string    `json:"categoriesBrands"`
					Breadcrumbs         []struct {
						Slug  string `json:"slug"`
						Label string `json:"label"`
					} `json:"breadcrumbs"`
					Brand struct {
						Slug string `json:"slug"`
						Name string `json:"name"`
					} `json:"brand"`
					AvailableShippingMethod []string `json:"availableShippingMethod"`
					HSCode                  string   `json:"HSCode"`
					BulletPoints            []string `json:"bulletPoints"`
					Timestamp               int64    `json:"timestamp"`
				} `json:"productFormated"`
			} `json:"pdp"`
		} `json:"initialState"`
	} `json:"props"`
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
	matched := productDetailExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}

	var viewData parseProductResponse
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	p := viewData.Props.InitialState.Pdp.ProductFormated

	description := ""

	for _, rawDesc := range p.BulletPoints {
		description = description + " " + rawDesc
	}
	if p.ProductInformation.CompositionEn != "" {
		description = description + ", INGREDIENTS : " + p.ProductInformation.CompositionEn
	}
	if p.ProductInformation.HeelSize != "" {
		description = description + ", HEEL HEIGHT : " + p.ProductInformation.HeelSize
	}
	if p.ProductInformation.BrandColor != "" {
		description = description + ", COLOR : " + p.ProductInformation.BrandColor
	}
	if p.ProductInformation.Dimensions != "" {
		description = description + ", DIMENSIONS : " + p.ProductInformation.Dimensions
	}
	if p.ProductInformation.MadeIn != "" {
		description = description + ", MADE IN : " + p.ProductInformation.MadeIn
	}

	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:       strconv.Format(viewData.Props.InitialState.Pdp.Product),
			CrawlUrl: resp.Request.URL.String(),
		},
		BrandName:   p.Brand.Name,
		Title:       p.Name,
		Description: htmlTrimRegp.ReplaceAllString(p.Description, ""),
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
	}

	currentPrice := 0
	if p.ShippingExpress[0].DiscountPriceInclVat > 0 {
		currentPrice = p.ShippingExpress[0].DiscountPriceInclVat
	} else {
		currentPrice = p.ShippingExpress[0].PriceInclVat
	}

	originalPrice := (p.ShippingExpress[0].PriceInclVat)
	discount := p.ShippingExpress[0].DiscountAmount

	for i, rawSku := range p.SizeAvailable {

		sku := pbItem.Sku{
			SourceId: strconv.Format(rawSku.LongSKU),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(currentPrice),
				Msrp:     int32(originalPrice),
				Discount: int32(discount),
			},
			Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}
		if rawSku.Stock > 0 {
			sku.Stock.StockStatus = pbItem.Stock_InStock
			sku.Stock.StockCount = int32(rawSku.Stock)
		}

		// color
		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecColor,
			Id:    strconv.Format(p.ProductInformation.ProductCode),
			Name:  p.ProductInformation.BrandColor,
			Value: p.ProductInformation.BrandColor,
			//Icon:  color.SwatchMedia.Mobile,
		})

		if i == 0 {

			isDefault := true
			for j := range p.Pictures {
				if j > 0 {
					isDefault = false
				}

				elementSelector := ".slick-slide[data-index=\"" + strconv.Format(j) + "\"]>div>picture>img"
				mediumUrl, _ := doc.Find(elementSelector).Attr("src")

				sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
					strconv.Format(j),
					mediumUrl,
					mediumUrl,
					mediumUrl,
					mediumUrl,
					"", isDefault))
			}
		}

		// size
		sizeStruct := viewData.Props.InitialState.Pdp.Sizes.Mapping[rawSku.SizeCode]
		sizeVal := ""
		for _, mid := range sizeStruct {
			sizeVal = strings.Join([]string{sizeVal, mid.Label}, " / ")
		}

		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Id:    rawSku.SizeCode,
			Name:  (strings.TrimPrefix(sizeVal, " / ")),
			Value: (strings.TrimPrefix(sizeVal, " / ")),
		})

		item.SkuItems = append(item.SkuItems, &sku)

	}

	// yield item result
	if err = yield(ctx, &item); err != nil {
		return err
	}

	return nil
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		"https://www.24s.com/en-us/women/ready-to-wear/coats",
		//"https://www.24s.com/en-us/long-coat-acne-studios_ACNEWD32BEIWD03800?color=camel-melange",
		//"https://www.24s.com/en-us/jinn-85-pumps-jimmy-choo_JCHZK4R3GEESI39500?color=dark-moss",
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

// local test
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
	opts := spider.CrawlOptions()

	// this callback func is used to do recursion call of sub requests.
	var callback func(ctx context.Context, val interface{}) error
	callback = func(ctx context.Context, val interface{}) error {
		switch i := val.(type) {
		case *http.Request:
			logger.Debugf("Access %s", i.URL)

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
				i.URL.Host = "www.24s.com"
			}

			// do http requests here.
			nctx, cancel := context.WithTimeout(ctx, time.Minute*5)
			defer cancel()
			resp, err := client.DoWithOptions(nctx, i, http.Options{
				EnableProxy:       true,
				EnableHeadless:    true,
				EnableSessionInit: spider.CrawlOptions().EnableSessionInit,
				KeepSession:       spider.CrawlOptions().KeepSession,
				Reliability:       spider.CrawlOptions().Reliability,
			})
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			return spider.Parse(ctx, resp, callback)
		default:
			// output the result
			data, err := json.Marshal(i)
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
