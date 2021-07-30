package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
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
		categoryPathMatcher: regexp.MustCompile(`^(/[a-z0-9_-]+)?/en\-us/(women|men)(/[a-z0-9_\-]+)+$`),
		// this regular used to match product page url path
		// /en-us/long-waterproof-cotton-coat-bottega-veneta_BOT84658YELWI04600
		productPathMatcher: regexp.MustCompile(`^/en\-us/[a-z0-9\-]+_[A-Z0-9]+$`),
		logger:             logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "7e5bdeec087b826be4f4f2a28d1818e1"
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
		EnableHeadless: false,
		// use js api to init session for the first request of the crawl
		EnableSessionInit: true,
		Reliability:       proxy.ProxyReliability_ReliabilityDefault,
	}
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.24s.com"}
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
		u.Host = "www.24s.com"
	}

	if c.productPathMatcher.MatchString(u.Path) {
		u.RawQuery = ""
		return u.String(), nil
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

	if p == "/en-us/women" || p == "/en-us/men" {
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

	sel := dom.Find(`#categoryTabContent > ul`)

	for i := range sel.Nodes {
		node := sel.Eq(i)
		cateName := "Men"
		if strings.Contains(node.AttrOr("id", ""), "women-nav") {
			cateName = "Women"
		}

		nnctx := context.WithValue(ctx, "Category", cateName)
		subSel := node.Find(`.nav-item .nav-item-parent`)

		for k := range subSel.Nodes {
			subNode2 := subSel.Eq(k)
			subcat2 := subNode2.Find(`span`).First().Text()

			subNode2list := subNode2.Find(`li`)

			for j := range subNode2list.Nodes {
				subNode := subNode2list.Eq(j)
				href := subNode.Find(`a`).AttrOr("href", "")
				if href == "" {
					continue
				}

				u, err := url.Parse(href)
				if err != nil {
					c.logger.Error("parse url %s failed", href)
					continue
				}

				subCateName := subcat2 + " > " + strings.TrimSpace(subNode.Text())

				if c.categoryPathMatcher.MatchString(u.Path) {
					nnnctx := context.WithValue(nnctx, "SubCategory", subCateName)
					req, _ := http.NewRequest(http.MethodGet, href, nil)
					if err := yield(nnnctx, req); err != nil {
						return err
					}
				}
			}

			if len(subNode2list.Nodes) == 0 {

				href := subNode2.Find(`a`).AttrOr("href", "")
				if href == "" {
					continue
				}

				u, err := url.Parse(href)
				if err != nil {
					c.logger.Error("parse url %s failed", href)
					continue
				}

				subCateName := subcat2
				if c.categoryPathMatcher.MatchString(u.Path) {
					nnnctx := context.WithValue(nnctx, "SubCategory", subCateName)
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

type productDetail struct {
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
	SizeAvailable []*struct {
		SizeCode      string `json:"sizeCode"`
		LongSKU       string `json:"longSKU"`
		HasOffer      bool   `json:"hasOffer"`
		SizeLabel     string `json:"sizeLabel"`
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
	OfferID             string `json:"offerId"`
	Name                string `json:"name"`
	MainCategoryBrand   string `json:"mainCategoryBrand"`
	LongSKU             string `json:"longSKU"`
	FacetColorAvailable []*struct {
		Path            string      `json:"path"`
		LongSKU         string      `json:"longSKU"`
		HasOffer        bool        `json:"hasOffer"`
		BrandColorFront interface{} `json:"brandColorFront"`
		BrandColor      string      `json:"brandColor"`
		Stock           int         `json:"stock"`
		Replenishment   bool        `json:"replenishment"`
	} `json:"facetColorAvailable"`
	Destination      string   `json:"destination"`
	Description      string   `json:"description"`
	Cites            string   `json:"cites"`
	CategoriesMaster []string `json:"categoriesMaster"`
	CategoriesBrands []string `json:"categoriesBrands"`
	Breadcrumbs      []struct {
		Slug  string `json:"slug"`
		Label string `json:"label"`
	} `json:"breadcrumbs"`
	Brand struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
	} `json:"brand"`
	AvailableShippingMethod []string      `json:"availableShippingMethod"`
	HSCode                  string        `json:"HSCode"`
	BulletPoints            []interface{} `json:"bulletPoints"`
	HierarchicalCategories  []struct {
		Label string `json:"label"`
	} `json:"hierarchicalCategories"`
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
				SelectedSize    interface{}    `json:"selectedSize"`
				SelectedColor   string         `json:"selectedColor"`
				ProductFormated *productDetail `json:"productFormated"`
			} `json:"pdp"`
		} `json:"initialState"`
	} `json:"props"`
	RuntimeConfig struct {
		NEXTPUBLICADYENORIGINKEY       string `json:"NEXT_PUBLIC_ADYEN_ORIGIN_KEY"`
		NEXTPUBLICTHUMBORHOST          string `json:"NEXT_PUBLIC_THUMBOR_HOST"`
		NEXTPUBLICBASEURL              string `json:"NEXT_PUBLIC_BASE_URL"`
		NEXTPUBLICOFFERBASEURL         string `json:"NEXT_PUBLIC_OFFER_BASE_URL"`
		NEXTPUBLICRECAPTCHA            string `json:"NEXT_PUBLIC_RECAPTCHA"`
		NEXTPUBLICRISKIFIEDDOMAIN      string `json:"NEXT_PUBLIC_RISKIFIED_DOMAIN"`
		NEXTPUBLICTHUMBORKEY           string `json:"NEXT_PUBLIC_THUMBOR_KEY"`
		NEXTPUBLICOFFERAUTH            string `json:"NEXT_PUBLIC_OFFER_AUTH"`
		NEXTPUBLICFRONTAUTH            string `json:"NEXT_PUBLIC_FRONT_AUTH"`
		NEXTPUBLICSTRIPEKEYPUBLIC      string `json:"NEXT_PUBLIC_STRIPE_KEY_PUBLIC"`
		NEXTPUBLICERRORBOUNDARIESDEBUG string `json:"NEXT_PUBLIC_ERROR_BOUNDARIES_DEBUG"`
		NEXTPUBLICBABYLONEBASEURL      string `json:"NEXT_PUBLIC_BABYLONE_BASE_URL"`
		NEXTPUBLICAPIGWBASEURL         string `json:"NEXT_PUBLIC_APIGW_BASE_URL"`
	} `json:"runtimeConfig"`
}

// used to trim html labels in description
var htmlTrimRegp = regexp.MustCompile(`</?[^>]+>`)

func generateImageUrl(host, key, size, path string) string {
	if host == "" {
		host = "https://img.zolaprod.babsta.net"
	}
	secret := "unsafe"
	subpath := "fit-in/" + size + "/" + path

	if key != "" {
		mac := hmac.New(sha1.New, []byte(key))
		mac.Write([]byte(subpath))
		secret = base64.URLEncoding.WithPadding(base64.StdPadding).EncodeToString(mac.Sum(nil))
	}
	return host + "/" + secret + "/" + subpath
}

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
		//c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		//return err
	}

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	var (
		p        = viewData.Props.InitialState.Pdp.ProductFormated
		colorSku *pbItem.SkuSpecOption
	)
	if p.ProductInformation.BrandColor != "" {
		colorSku = &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecColor,
			Id:    p.ProductInformation.BrandColor,
			Name:  p.ProductInformation.BrandColor,
			Value: p.ProductInformation.BrandColor,
		}
	}

	canUrl := dom.Find(`link[rel="canonical"]`).AttrOr("href", "")
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}
	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           p.LongSKU,
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
			GroupId:      p.HSCode,
		},
		BrandName: p.Brand.Name,
		Title:     p.Name,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
	}

	for i, bread := range p.HierarchicalCategories {
		switch i {
		case 0:
			item.CrowdType = bread.Label
			item.Category = bread.Label
		case 1:
			item.SubCategory = bread.Label
		case 2:
			item.SubCategory2 = bread.Label
		case 3:
			item.SubCategory3 = bread.Label
		case 4:
			item.SubCategory4 = bread.Label
		}
	}

	description := p.Description
	for _, rawDesc := range p.BulletPoints {
		description = description + rawDesc.(string) + ". "
	}
	description = description + ", Material: " + p.ProductInformation.CompositionEn + " Color: " + p.ProductInformation.BrandColor + ","
	if p.ProductInformation.Dimensions != "" {
		description = description + " SIZE & MEASUREMENTS: " + p.ProductInformation.Dimensions + " " + p.ProductInformation.BrandInformation
	}

	item.Description = description

	currentPrice := 0
	if len(p.ShippingExpress) == 0 {
		return fmt.Errorf("no longer available")
	}
	if p.ShippingExpress[0].DiscountPriceInclVat > 0 {
		currentPrice = p.ShippingExpress[0].DiscountPriceInclVat
	} else {
		currentPrice = p.ShippingExpress[0].PriceInclVat
	}

	var (
		originalPrice = (p.ShippingExpress[0].PriceInclVat)
		discount      = p.ShippingExpress[0].DiscountAmount

		thumborHost = viewData.RuntimeConfig.NEXTPUBLICTHUMBORHOST
		thumborKey  = viewData.RuntimeConfig.NEXTPUBLICTHUMBORKEY
	)

	for j, pic := range p.Pictures {
		item.Medias = append(item.Medias, pbMedia.NewImageMedia(
			pic.Path,
			generateImageUrl(thumborHost, thumborKey, "", pic.Path),
			generateImageUrl(thumborHost, thumborKey, "800x", pic.Path),
			generateImageUrl(thumborHost, thumborKey, "600x", pic.Path),
			generateImageUrl(thumborHost, thumborKey, "500x", pic.Path),
			"", j == 0))
	}

	for _, rawSku := range p.SizeAvailable {
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
			item.Stock.StockStatus = pbItem.Stock_InStock
		}

		// color
		if colorSku != nil {
			sku.Specs = append(sku.Specs, colorSku)
		}

		// size
		sizeStruct := viewData.Props.InitialState.Pdp.Sizes.Mapping[rawSku.SizeCode]
		sizeVal := ""
		for _, mid := range sizeStruct {
			if sizeVal == "" {
				sizeVal = mid.Label
			} else {
				sizeVal = sizeVal + " / " + mid.Label
			}
			break
		}
		if sizeVal == "" {
			sizeVal = rawSku.SizeCode
		}
		if sizeVal == "" {
			sizeVal = rawSku.SizeLabel
		}

		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Id:    rawSku.SizeCode,
			Name:  sizeVal,
			Value: sizeVal,
		})
		item.SkuItems = append(item.SkuItems, &sku)
	}

	// yield item result
	if err = yield(ctx, &item); err != nil {
		c.logger.Errorf("yield sub request failed, error=%s", err)
		return err
	}

	// for other colors
	for _, color := range p.FacetColorAvailable {
		if strings.Replace(color.BrandColor, "_", " ", -1) == strings.Replace(p.ProductInformation.BrandColor, "_", " ", -1) {
			continue
		}
		u := strings.Replace(resp.Request.URL.String(), "_"+p.LongSKU, "_"+color.LongSKU, -1)
		req, err := http.NewRequest(http.MethodGet, u, nil)
		if err != nil {
			c.logger.Errorf("yield new color crawl failed, error=%s", err)
			continue
		}
		req.Header.Set("Referer", resp.Request.Header.Get("Referer"))

		if err := yield(ctx, req); err != nil {
			c.logger.Errorf("yield sub request failed, error=%s", err)
			return err
		}
	}
	return nil
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://www.24s.com/en-us/women",
		//"https://www.24s.com/en-us/nash-face-t-shirt-acne-studios_ACNDDJX8GRYMZZXS00",
		//"https://www.24s.com/en-us/women/ready-to-wear/coats",
		//"https://www.24s.com/en-us/long-coat-acne-studios_ACNEWD32BEIWD03800?color=camel-melange",
		//"https://www.24s.com/en-us/jinn-85-pumps-jimmy-choo_JCHZK4R3GEESI39500?color=dark-moss",
		"https://www.24s.com/en-us/printed-t-shirt-undercover_UNDXQQ5GWHTMNZ0200?color=white",
		//"https://www.24s.com/en-us/fluo-pink-enamel-v-ring-djula_DJU282F7GOLLU1A100",
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
