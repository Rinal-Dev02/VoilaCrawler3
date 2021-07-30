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
	"github.com/voiladev/VoilaCrawler/pkg/cli"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	pbMedia "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
)

// _Crawler defined the crawler struct/class for which is not necessory to be exportable
type _Crawler struct {
	// httpClient is the object of an http client
	httpClient            http.Client
	categoryPathMatcher   *regexp.Regexp
	productPathMatcher    *regexp.Regexp
	productPathMatcherNew *regexp.Regexp
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
		categoryPathMatcher: regexp.MustCompile(`^/collections(/[a-z0-9\-]+){1,6}$`),
		// this regular used to match product page url path
		productPathMatcher:    regexp.MustCompile(`^(/[a-z0-9\-]+){1,4}/\d+\.html$`),
		productPathMatcherNew: regexp.MustCompile(`^(/products(/[a-z0-9\-]+){1,5})|(/collections(/[a-z0-9\-]+){1,6}/products(/[a-z0-9\-]+){1,5})$`),
		logger:                logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "9618d9d4d07d8eb150b801367fb9af6f"
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
		EnableHeadless: true,
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

func (c *_Crawler) CanonicalUrl(rawurl string) (string, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	if u.Host == "" {
		u.Host = "www.modcloth.com"
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
	p := strings.TrimSuffix(resp.RawUrl().Path, "/")

	if p == "/en_us" || p == "" {
		return c.parseCategories(ctx, resp, yield)
	}
	if c.productPathMatcherNew.MatchString(resp.RawUrl().Path) {
		return c.parseProductNew(ctx, resp, yield)
	} else if c.categoryPathMatcher.MatchString(resp.RawUrl().Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	} else if c.productPathMatcher.MatchString(resp.RawUrl().Path) {
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

	matched := dom.Find(`.component-mobile-menu-basic`).AttrOr(`data-props`, "")

	var viewData categoryStructure
	if err := json.Unmarshal([]byte(matched), &viewData); err != nil {
		c.logger.Errorf("unmarshal category detail data fialed, error=%s", err)
		return err
	}

	for _, rawcat := range viewData.Links {
		nnctx := context.WithValue(ctx, "Category", rawcat.Title)

		for _, rawsubcat := range rawcat.Links {

			for _, rawsubcatlvl2 := range rawsubcat.Links {
				href := rawsubcatlvl2.URL
				if href == "" {
					continue
				}
				u, err := url.Parse(href)
				if err != nil {
					c.logger.Errorf("parse url %s failed", href)
					continue
				}
				subCatName := rawsubcat.Title + " > " + rawsubcatlvl2.Title

				if c.categoryPathMatcher.MatchString(u.Path) {
					nctx := context.WithValue(nnctx, "SubCategory", subCatName)
					req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
					if err := yield(nctx, req); err != nil {
						return err
					}
				}
			}

			if len(rawsubcat.Links) == 0 { // No sub categories
				href := rawsubcat.URL
				if href == "" {
					continue
				}

				u, err := url.Parse(href)
				if err != nil {
					c.logger.Errorf("parse url %s failed", href)
					continue
				}

				if c.categoryPathMatcher.MatchString(u.Path) {
					nctx := context.WithValue(nnctx, "SubCategory", rawsubcat.Title)
					req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
					if err := yield(nctx, req); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

type categoryStructure struct {
	Title string `json:"title"`
	Links []struct {
		URL   string `json:"url"`
		Title string `json:"title"`
		Links []struct {
			URL   string `json:"url"`
			Title string `json:"title"`
			Links []struct {
				URL   string `json:"url"`
				Title string `json:"title"`
			} `json:"links,omitempty"`
		} `json:"links"`
	} `json:"links"`
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
		Type        string      `json:"@type"`
		RatingValue interface{} `json:"ratingValue"`
		ReviewCount interface{} `json:"reviewCount"`
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

type RawProductDetailsNew struct {
	ID                   int64    `json:"id"`
	Title                string   `json:"title"`
	Handle               string   `json:"handle"`
	Description          string   `json:"description"`
	PublishedAt          string   `json:"published_at"`
	CreatedAt            string   `json:"created_at"`
	Vendor               string   `json:"vendor"`
	Type                 string   `json:"type"`
	Tags                 []string `json:"tags"`
	Price                int      `json:"price"`
	PriceMin             int      `json:"price_min"`
	PriceMax             int      `json:"price_max"`
	Available            bool     `json:"available"`
	PriceVaries          bool     `json:"price_varies"`
	CompareAtPrice       int      `json:"compare_at_price"`
	CompareAtPriceMin    int      `json:"compare_at_price_min"`
	CompareAtPriceMax    int      `json:"compare_at_price_max"`
	CompareAtPriceVaries bool     `json:"compare_at_price_varies"`
	Variants             []struct {
		ID                     int64         `json:"id"`
		Title                  string        `json:"title"`
		Option1                string        `json:"option1"`
		Option2                string        `json:"option2"`
		Option3                interface{}   `json:"option3"`
		Sku                    string        `json:"sku"`
		RequiresShipping       bool          `json:"requires_shipping"`
		Taxable                bool          `json:"taxable"`
		FeaturedImage          interface{}   `json:"featured_image"`
		Available              bool          `json:"available"`
		Name                   string        `json:"name"`
		PublicTitle            string        `json:"public_title"`
		Options                []string      `json:"options"`
		Price                  int           `json:"price"`
		Weight                 int           `json:"weight"`
		CompareAtPrice         int           `json:"compare_at_price"`
		InventoryManagement    string        `json:"inventory_management"`
		Barcode                string        `json:"barcode"`
		RequiresSellingPlan    bool          `json:"requires_selling_plan"`
		SellingPlanAllocations []interface{} `json:"selling_plan_allocations"`
	} `json:"variants"`
	Images        []string `json:"images"`
	FeaturedImage string   `json:"featured_image"`
	Options       []string `json:"options"`
	Media         []struct {
		Alt          string `json:"alt"`
		ID           int64  `json:"id"`
		Position     int    `json:"position"`
		PreviewImage struct {
			AspectRatio float64 `json:"aspect_ratio"`
			Height      int     `json:"height"`
			Width       int     `json:"width"`
			Src         string  `json:"src"`
		} `json:"preview_image"`
		AspectRatio float64 `json:"aspect_ratio"`
		Height      int     `json:"height"`
		MediaType   string  `json:"media_type"`
		Src         string  `json:"src"`
		Width       int     `json:"width"`
	} `json:"media"`
	RequiresSellingPlan bool          `json:"requires_selling_plan"`
	SellingPlanGroups   []interface{} `json:"selling_plan_groups"`
	Content             string        `json:"content"`
}

func TrimSpaceNewlineInString(s string) string {
	re := regexp.MustCompile(`\n`)
	resp := re.ReplaceAllString(s, " ")
	resp = strings.ReplaceAll(resp, "\\n", " ")
	resp = strings.ReplaceAll(resp, "\r", " ")
	resp = strings.ReplaceAll(resp, "\t", " ")
	resp = strings.ReplaceAll(resp, "  ", "")
	return resp
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

var productsExtractRegNew = regexp.MustCompile(`(?Us)var\s*afterpay_product\s*=\s*({.*});`)

// parseProduct
func (c *_Crawler) parseProduct(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil {
		return nil
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	//looks like this page doesn’t fit what you were looking for
	if bytes.Contains(respBody, []byte(`looks like this page doesn’t fit what you were looking for`)) {
		fmt.Println(`Not found`)
		return nil
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

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Debug(err)
		return err
	}
	canUrl := dom.Find(`link[rel="canonical"]`).AttrOr("href", "")
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}

	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           strconv.Format(viewData.ProductID),
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
		},
		BrandName:   viewData.Brand.Name,
		Title:       viewData.Name,
		Description: TrimSpaceNewlineInString(viewData.Description),
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		// Stats: &pbItem.Stats{
		// 	ReviewCount: int32(viewData.AggregateRating.ReviewCount),
		// 	Rating:      float32(viewData.AggregateRating.RatingValue),
		// },
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
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
	item.Medias = medias

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

			for k, rawVariation := range rawSku.ProductVariants {
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
					item.Stock.StockStatus = pbItem.Stock_InStock
				}

				if colorSpec != nil {
					sku.Specs = append(sku.Specs, colorSpec)
				}
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecSize,
					Id:    fmt.Sprintf("%s-%v", rawVariation.Upc, k),
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

var htmlTrimRegp = regexp.MustCompile(`</?[^>]+>`)

// parseProduct
func (c *_Crawler) parseProductNew(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil {
		return nil
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	//looks like this page doesn’t fit what you were looking for
	if bytes.Contains(respBody, []byte(`looks like this page doesn’t fit what you were looking for`)) {
		fmt.Println(`Not found`)
		return nil
	}

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Debug(err)
		return err
	}

	matched := productsExtractRegNew.FindSubmatch(respBody)
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract product info from %s failed, error=%s", resp.Request.URL, err)
	}
	// c.logger.Debugf("data: %s", matched[1])

	var (
		viewData       RawProductDetailsNew
		viewDataReview RawProductDetails
	)

	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}

	// review
	if len(dom.Find(`.y-rich-snippet-script`).Nodes) > 0 {
		if err := json.Unmarshal([]byte(dom.Find(`.y-rich-snippet-script`).Text()), &viewDataReview); err != nil {
			c.logger.Errorf("unmarshal product review detail data fialed, error=%s", err)
			return err
		}
	}

	canUrl := dom.Find(`link[rel="canonical"]`).AttrOr("href", "")
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}

	colorIndex := -1
	sizeIndex := -1
	for k, key := range viewData.Options {
		if key == "Color" {
			colorIndex = k
		} else if key == "Size" {
			sizeIndex = k
		}
	}

	rating, _ := strconv.ParseFloat(viewDataReview.AggregateRating.RatingValue)
	reviewcount, _ := strconv.ParseInt(viewDataReview.AggregateRating.ReviewCount)

	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           strconv.Format(viewData.ID),
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
		},
		BrandName:   viewData.Vendor,
		Title:       viewData.Title,
		Description: TrimSpaceNewlineInString(strings.TrimSpace((htmlTrimRegp.ReplaceAllString(viewData.Description, " ")))),
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stats: &pbItem.Stats{
			ReviewCount: int32(reviewcount),
			Rating:      float32(rating),
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
	}

	breadSel := dom.Find(`.breadcrumb`).Find(`a`)
	for i := range breadSel.Nodes {

		node := breadSel.Eq(i)
		switch i {
		case 1:
			item.Category = strings.TrimSpace(node.Text())
		case 2:
			item.SubCategory = strings.TrimSpace(node.Text())
		case 3:
			item.SubCategory2 = strings.TrimSpace(node.Text())
		case 4:
			item.SubCategory3 = strings.TrimSpace(node.Text())
		case 5:
			item.SubCategory4 = strings.TrimSpace(node.Text())
		}
	}

	var medias []*pbMedia.Media
	for m, mid := range viewData.Images {
		s := strings.Split(mid, "?")
		medias = append(medias, pbMedia.NewImageMedia(
			strconv.Format(m),
			s[0],
			strings.ReplaceAll(s[0], ".jpg", "_900x.jpg"),
			strings.ReplaceAll(s[0], ".jpg", "_600x.jpg"),
			strings.ReplaceAll(s[0], ".jpg", "_500x.jpg"),
			"",
			m == 0,
		))
	}
	item.Medias = medias

	for _, rawVariation := range viewData.Variants {

		curretnPrice := int32(rawVariation.Price)
		msrp := int32(rawVariation.CompareAtPrice)
		if msrp == 0 {
			msrp = curretnPrice
		}
		discount := (int32)(0)
		if msrp > curretnPrice {
			discount, _ = strconv.ParseInt32(((rawVariation.CompareAtPrice - rawVariation.Price) / rawVariation.CompareAtPrice) * 100)
		}
		sku := pbItem.Sku{
			SourceId: strconv.Format(rawVariation.ID),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(curretnPrice),
				Msrp:     int32(msrp),
				Discount: discount,
			},
			Medias: medias,
			Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}
		if rawVariation.Available {
			sku.Stock.StockStatus = pbItem.Stock_InStock
			//sku.Stock.StockCount = int32(rawVariation.UnitsAvailable)
			item.Stock.StockStatus = pbItem.Stock_InStock
		}

		if colorIndex > -1 {
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    rawVariation.Barcode,
				Name:  rawVariation.Options[colorIndex],
				Value: rawVariation.Options[colorIndex],
			})
		}

		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Id:    rawVariation.Sku,
			Name:  rawVariation.Options[sizeIndex],
			Value: rawVariation.Options[sizeIndex],
		})
		item.SkuItems = append(item.SkuItems, &sku)
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
		//"https://www.modcloth.com",
		//"https://www.modcloth.com/shop/best-selling-shoes",
		//"https://www.modcloth.com/shop/shoes/t-u-k-the-zest-is-history-heel-in-burgundy/128047.html",
		//"https://modcloth.com/products/spritely-as-spring-midi-dress-yellow?nosto=notfound-nosto-1",
		//"https://modcloth.com/collections/cocktail-dresses/products/lace-lady-lace-fit-and-flare-dress-pink",
		"https://modcloth.com/collections/flats/products/step-out-of-the-crowd-ballet-flat-mustard",
		//"https://modcloth.com/products/step-out-of-the-crowd-ballet-flat-mustard",
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
