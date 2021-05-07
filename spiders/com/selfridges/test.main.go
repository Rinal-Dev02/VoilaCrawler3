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
	"github.com/voiladev/go-crawler/pkg/cli"
	"github.com/voiladev/go-crawler/pkg/crawler"
	"github.com/voiladev/go-crawler/pkg/net/http"
	pbMedia "github.com/voiladev/go-crawler/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/go-crawler/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/go-crawler/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	pbProxy "github.com/voiladev/go-crawler/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
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
		categoryPathMatcher: regexp.MustCompile(`^/US/en/cat(/[A-Za-z0-9_-]+){1,5}`),
		// this regular used to match product page url path
		productPathMatcher: regexp.MustCompile(`((.*)previewAttribute=(.*))|((.*)_[/A-Za-z0-9]+)`),
		logger:             logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "49343fe1da77b7d234384ea594d59769"
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
	return &crawler.CrawlOptions{
		EnableHeadless: true,
		// use js api to init session for the first request of the crawl
		EnableSessionInit: true,
		Reliability:       pbProxy.ProxyReliability_ReliabilityDefault,
	}
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.selfridges.com"}
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
		u.Host = "www.selfridges.com"
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
	if c.productPathMatcher.MatchString(resp.Request.URL.String()) || c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yield)
	} else if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	}
	return crawler.ErrUnsupportedPath
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
	Context     string `json:"@context"`
	Type        string `json:"@type"`
	URL         string `json:"url"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MainEntity  struct {
		Type            string `json:"@type"`
		ItemListElement []struct {
			Type           string `json:"@type"`
			Brand          string `json:"brand"`
			Description    string `json:"description"`
			AdditionalType string `json:"additionalType"`
			Sku            string `json:"sku"`
			Depth          int    `json:"depth"`
			Image          string `json:"image"`
			URL            string `json:"url"`
			Name           string `json:"name"`
			Offers         struct {
				Type          string `json:"@type"`
				Price         string `json:"price"`
				PriceCurrency string `json:"priceCurrency"`
				Category      string `json:"category"`
				Sku           string `json:"sku"`
				Description   string `json:"description"`
				Name          string `json:"name"`
				URL           string `json:"url"`
			} `json:"offers"`
			Category string `json:"category"`
		} `json:"itemListElement"`
	} `json:"mainEntity"`
	Breadcrumb struct {
		Type            string `json:"@type"`
		ItemListElement []struct {
			Type     string `json:"@type"`
			Position int    `json:"position"`
			Item     struct {
				ID   string `json:"@id"`
				Name string `json:"name"`
			} `json:"item"`
		} `json:"itemListElement"`
	} `json:"breadcrumb"`
}

// used to extract embaded json data in website page.
// more about golang regulation see here https://golang.org/pkg/regexp/syntax/
var productsExtractReg = regexp.MustCompile(`(?U)<script type="application/ld\+json">\s*({.*})</script>`)
var productsDataExtractReg = regexp.MustCompile(`(?U)window\.dataLayerExtension\s*=\s*({.*});`)

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

	matched := productsExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}
	// c.logger.Debugf("data: %s", matched[1])

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	var viewData CategoryView
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		return err
	}

	lastIndex := nextIndex(ctx)
	sel := doc.Find(`.c-prod-card__cta-box>a`)
	c.logger.Debugf("nodes %d", len(sel.Nodes))
	for i := range sel.Nodes {
		node := sel.Eq(i)
		if href, _ := node.Attr("href"); href != "" {

			//fmt.Println(href)
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

	// get current page number
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("pn"))
	if page == 0 {
		page = 1
	}

	tp, _ := doc.Find(`.plp-listing-load-status.c-list-footer__counter`).Attr(`data-total-pages-count`)
	totalPageCount, _ := strconv.ParseInt(tp)

	tp, _ = doc.Find(`.plp-listing-load-status.c-list-footer__counter`).Attr(`data-total-products-count`)
	totalResultCount, _ := strconv.ParseInt(tp)

	// check if this is the last page
	if len(viewData.MainEntity.ItemListElement) >= int(totalResultCount) ||
		page >= int64(totalPageCount) {
		return nil
	}

	// set pagination
	u := *resp.Request.URL
	vals := u.Query()
	vals.Set("pn", strconv.Format(page+1))
	u.RawQuery = vals.Encode()

	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	// update the index of last page
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
}

type parseProductResponse struct {
	PageBreadcrumb                 string   `json:"page_breadcrumb"`
	PageCategoryID                 string   `json:"page_category_id"`
	PageCategoryName               string   `json:"page_category_name"`
	PageName                       string   `json:"page_name"`
	PageType                       string   `json:"page_type"`
	ProductBrand                   []string `json:"product_brand"`
	ProductCategory                []string `json:"product_category"`
	ProductDepartmentID            []string `json:"product_department_id"`
	ProductDepartmentName          []string `json:"product_department_name"`
	ProductDivisionID              []string `json:"product_division_id"`
	ProductDivisionName            []string `json:"product_division_name"`
	ProductID                      []string `json:"product_id"`
	ProductImage                   []string `json:"product_image"`
	ProductName                    []string `json:"product_name"`
	ProductObcn                    []string `json:"product_obcn"`
	ProductPersonalised            []string `json:"product_personalised"`
	ProductPrice                   []string `json:"product_price"`
	ProductPriceGbp                []string `json:"product_price_gbp"`
	ProductPriceType               []string `json:"product_price_type"`
	ProductSeason                  []string `json:"product_season"`
	ProductStock                   []string `json:"product_stock"`
	ProductSubscriptionEligibility []string `json:"product_subscription_eligibility"`
	ProductSubtype                 []string `json:"product_subtype"`
	ProductType                    []string `json:"product_type"`
	ProductWcid                    []string `json:"product_wcid"`
	WasPrice                       []string `json:"was_price"`
	WasWasPrice                    []string `json:"was_was_price"`
	ProductGroupID                 []string `json:"product_group_id"`
	ProductGroupName               []string `json:"product_group_name"`
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

	matched := productsDataExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	var viewData parseProductResponse

	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}

	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:       strconv.Format(viewData.ProductID),
			CrawlUrl: resp.Request.URL.String(),
		},
		BrandName:   viewData.ProductBrand[0],
		Title:       viewData.ProductName[0],
		Description: htmlTrimRegp.ReplaceAllString(doc.Find(`#content1`).Text(), ""),
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
	}

	originalPrice, _ := strconv.ParseFloat(viewData.ProductPrice[0])
	msrp := 0.0
	discount := 0.0
	if len(viewData.WasPrice) > 0 {
		msrp, _ = strconv.ParseFloat(viewData.WasPrice[0])
	}
	if msrp == 0.0 {
		msrp = originalPrice
	}
	if msrp > originalPrice {
		discount = ((originalPrice - msrp) / msrp) * 100
	}

	// Note: Color variation is available on product list page therefor not considering multiple color of a product
	colorName, _ := doc.Find(`[class^="c-select c-filter__select --colour"]`).Find(`[class^="c-select__dropdown-item --selected"]`).Attr(`data-js-action`)

	sel := doc.Find(`[class^="c-select c-filter__select --size"]`).Find(`.c-select__dropdown-item`)
	for i := range sel.Nodes {

		sku := pbItem.Sku{
			SourceId: strconv.Format(viewData.ProductID[0]),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(originalPrice * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},
			Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}
		if sizes, _ := sel.Eq(i).Html(); !strings.Contains(sizes, "c-select__dropdown-item --oos") {
			sku.Stock.StockStatus = pbItem.Stock_InStock
			//sku.Stock.StockCount = int32(rawSku.TotalQuantityAvailable)
		}

		// color
		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecColor,
			Id:    strconv.Format(viewData.ProductID[0]),
			Name:  colorName,
			Value: colorName,
			//Icon:  color.SwatchMedia.Mobile,
		})

		if i == 0 {

			isDefault := true
			imgs := doc.Find(`#big-image-overlay>#nextArrow>span`)
			for j := range imgs.Nodes {
				if j > 0 {
					isDefault = false
				}
				m, _ := imgs.Eq(j).Attr(`zoomsrc`)

				sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
					strconv.Format(j),
					m,
					strings.ReplaceAll(m, "$PDP_M_ZOOM$", "scl=1.6"),
					strings.ReplaceAll(m, "$PDP_M_ZOOM$", "scl=3"),
					strings.ReplaceAll(m, "$PDP_M_ZOOM$", "scl=2.5"),
					"",
					isDefault,
				))
			}
		}

		// size
		sizeValue, _ := sel.Eq(i).Attr(`data-js-action`)

		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Id:    strconv.Format(i),
			Name:  sizeValue,
			Value: sizeValue,
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
		"https://www.selfridges.com/US/en/cat/womens/clothing/tops/?cm_sp=MegaMenu-_-Women-_-Clothing-Tops&pn=1&scrollToProductId=R03735136_GARNET_ALT10",
		//"https://www.selfridges.com/US/en/cat/stella-mccartney-branded-relaxed-fit-cotton-jersey-vest_R03733067/?previewAttribute=PURE%20WHITE",
		//"https://www.selfridges.com/US/en/cat/a_R03647747/?previewAttribute=Sky Blue Cream",
		//"https://www.selfridges.com/US/en/cat/a_R03673363/?previewAttribute=CLOUD%20DANCER",
		"https://www.selfridges.com/US/en/cat/tom-ford-floral-print-leather-card-holder_R03714356/?previewAttribute=BLK+WHT",
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
