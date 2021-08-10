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
	return "25b88a3a3fa14e0fa4e18833d11bcf4e"
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

	sel := dom.Find(`#categories>li`)

	for i := range sel.Nodes {
		node := sel.Eq(i)
		cateName := strings.TrimSpace(node.Find(`a>span`).First().Text())

		if cateName == "" {
			continue
		}

		nctx := context.WithValue(ctx, "Category", cateName)

		subSel := node.Find(`.nested-navigation-section`)

		for k := range subSel.Nodes {
			subNode2 := subSel.Eq(k)

			subcat2 := strings.TrimSpace(subNode2.Find(`.sub-navigation-header`).First().Text())

			nnctx := context.WithValue(nctx, "SubCategory", subcat2)

			subNode3 := subNode2.Find(`.sub-navigation-list>li`)

			for j := range subNode3.Nodes {
				subNode := subNode3.Eq(j)
				subcat3 := strings.TrimSpace(subNode.Find(`.sub-navigation-list-item-link-text`).First().Text())
				if subcat3 == "" {
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

				subCateName := strings.TrimSpace(subNode.Text())

				if c.categoryPathMatcher.MatchString(u.Path) {
					nnnctx := context.WithValue(nnctx, "SubCategory2", subCateName)
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
			StyleColorCode string      `json:"styleColorCode"`
			ModelSize      interface{} `json:"modelSize"`
			ModelSizes     []struct {
				ModelName    string `json:"modelName"`
				ShotWithSize string `json:"shotWithSize"`
				DisplaySize  string `json:"displaySize"`
				Bust         struct {
					Inches          string `json:"inches"`
					Feet            string `json:"feet"`
					InchesRemainder string `json:"inchesRemainder"`
					Centimeters     string `json:"centimeters"`
				} `json:"bust"`
				Height struct {
					Inches          string `json:"inches"`
					Feet            string `json:"feet"`
					InchesRemainder string `json:"inchesRemainder"`
					Centimeters     string `json:"centimeters"`
				} `json:"height"`
				Hips struct {
					Inches          string `json:"inches"`
					Feet            string `json:"feet"`
					InchesRemainder string `json:"inchesRemainder"`
					Centimeters     string `json:"centimeters"`
				} `json:"hips"`
				Waist struct {
					Inches          string `json:"inches"`
					Feet            string `json:"feet"`
					InchesRemainder string `json:"inchesRemainder"`
					Centimeters     string `json:"centimeters"`
				} `json:"waist"`
			} `json:"modelSizes"`
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
			SizeAndFitDescription    string        `json:"sizeAndFitDescription"`
			SizeAndFitNote           string        `json:"sizeAndFitNote"`
			MeasurementsFromSize     string        `json:"measurementsFromSize"`
			Measurements             []string      `json:"measurements"`
			SizeScale                string        `json:"sizeScale"`
			SizeScaleLocale          string        `json:"sizeScaleLocale"`
			GarmentFit               []interface{} `json:"garmentFit"`
			SizeGuidance             []interface{} `json:"sizeGuidance"`
			NewMeasurementAttributes bool          `json:"newMeasurementAttributes"`
			LegacyMeasurements       bool          `json:"legacyMeasurements"`
		} `json:"sizeAndFitDetail"`
		LongDescriptionDetail struct {
			PreNotes           []interface{} `json:"preNotes"`
			Attributes         []string      `json:"attributes"`
			FloweryDescription interface{}   `json:"floweryDescription"`
		} `json:"longDescriptionDetail"`
		Proposition65                                   bool        `json:"proposition65"`
		Proposition65ChemicalsCancer                    interface{} `json:"proposition65ChemicalsCancer"`
		Proposition65ChemicalsReproductiveHarm          interface{} `json:"proposition65ChemicalsReproductiveHarm"`
		Proposition65ChemicalsCancerAndReproductiveHarm interface{} `json:"proposition65ChemicalsCancerAndReproductiveHarm"`
		GiftWithPurchase                                bool        `json:"giftWithPurchase"`
		DefaultStyleColor                               struct {
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
			StyleColorCode string      `json:"styleColorCode"`
			ModelSize      interface{} `json:"modelSize"`
			ModelSizes     []struct {
				ModelName    string `json:"modelName"`
				ShotWithSize string `json:"shotWithSize"`
				DisplaySize  string `json:"displaySize"`
				Bust         struct {
					Inches          string `json:"inches"`
					Feet            string `json:"feet"`
					InchesRemainder string `json:"inchesRemainder"`
					Centimeters     string `json:"centimeters"`
				} `json:"bust"`
				Height struct {
					Inches          string `json:"inches"`
					Feet            string `json:"feet"`
					InchesRemainder string `json:"inchesRemainder"`
					Centimeters     string `json:"centimeters"`
				} `json:"height"`
				Hips struct {
					Inches          string `json:"inches"`
					Feet            string `json:"feet"`
					InchesRemainder string `json:"inchesRemainder"`
					Centimeters     string `json:"centimeters"`
				} `json:"hips"`
				Waist struct {
					Inches          string `json:"inches"`
					Feet            string `json:"feet"`
					InchesRemainder string `json:"inchesRemainder"`
					Centimeters     string `json:"centimeters"`
				} `json:"waist"`
			} `json:"modelSizes"`
			SupportedVideos       []interface{} `json:"supportedVideos"`
			InStock               bool          `json:"inStock"`
			DefaultStyleColorSize interface{}   `json:"defaultStyleColorSize"`
		} `json:"defaultStyleColor"`
		OneSize            bool   `json:"oneSize"`
		AltText            string `json:"altText"`
		HasAnyMeasurements bool   `json:"hasAnyMeasurements"`
		InStock            bool   `json:"inStock"`
	} `json:"product"`
	PageInfo struct {
		CountDownDateMillis interface{} `json:"countDownDateMillis"`
		APIRoot             string      `json:"apiRoot"`
		CdnURL              string      `json:"cdnUrl"`
		SiteID              int         `json:"siteId"`
		Language            string      `json:"language"`
		IsMobile            bool        `json:"isMobile"`
		SbmWeblab           bool        `json:"sbmWeblab"`
		LocalizedStrings    struct {
			Hours                  string `json:"hours"`
			Months                 string `json:"months"`
			Weeks                  string `json:"weeks"`
			Week                   string `json:"week"`
			Year                   string `json:"year"`
			Minutes                string `json:"minutes"`
			ModelMeasurementHeight string `json:"modelMeasurementHeight"`
			ModelMeasurementInches string `json:"modelMeasurementInches"`
			Years                  string `json:"years"`
			Minute                 string `json:"minute"`
			Second                 string `json:"second"`
			Seconds                string `json:"seconds"`
			Month                  string `json:"month"`
			Hour                   string `json:"hour"`
			ModelMeasurementCm     string `json:"modelMeasurementCm"`
			Days                   string `json:"days"`
			Day                    string `json:"day"`
		} `json:"localizedStrings"`
		FolderID int `json:"folderId"`
	} `json:"pageInfo"`
	Brand struct {
		BrandCode   string `json:"brandCode"`
		DesignerBio string `json:"designerBio"`
		Folder      struct {
			ID             int    `json:"id"`
			URLDescription string `json:"urlDescription"`
			Name           string `json:"name"`
			Label          string `json:"label"`
		} `json:"folder"`
		MyDesigner bool `json:"myDesigner"`
	} `json:"brand"`
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
			Id:           strconv.Format(viewData.Product.StyleNumber),
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

			fmt.Println(imgregx.ReplaceAllString(template, "_UX1000_"))
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

		item.Medias = append(item.Medias, medias...)

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
					Current:  int32(current),
					Msrp:     int32(msrp),
					Discount: int32(discount),
				},

				Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}

			if rawsku.InStock {
				sku.Stock.StockStatus = pbItem.Stock_InStock
				item.Stock.StockStatus = pbItem.Stock_InStock
			}

			if colorSelected != nil {
				sku.Specs = append(sku.Specs, colorSelected)
			}

			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    fmt.Sprintf("%s-%s", rawsize.Sin, rawsize.Size.Code),
				Name:  rawsize.Size.Label,
				Value: rawsize.Size.Label,
			})

			item.SkuItems = append(item.SkuItems, &sku)
		}
	}

	jsonData, err := json.Marshal(item)

	fmt.Println(string(jsonData))

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
		"https://www.eastdane.com/ami-coeur-sweater/vp/v=1/1527786695.htm",
		//"https://www.eastdane.com/clothing-pants/br/v=1/19210.htm",
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
