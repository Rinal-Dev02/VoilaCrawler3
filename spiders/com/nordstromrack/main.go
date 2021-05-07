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

	"github.com/PuerkitoBio/goquery"
	"github.com/voiladev/go-crawler/pkg/cli"
	"github.com/voiladev/go-crawler/pkg/crawler"
	"github.com/voiladev/go-crawler/pkg/net/http"
	pbMedia "github.com/voiladev/go-crawler/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/go-crawler/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/go-crawler/protoc-gen-go/chameleon/smelter/v1/crawl/item"
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
		///shop/Women/Clothing/Tops
		categoryPathMatcher: regexp.MustCompile(`^/(category|shop|c|events)(/[,\s%&a-zA-Z0-9_\-]+){1,6}$`),
		// this regular used to match product page url path
		productPathMatcher: regexp.MustCompile(`^/s(/.*){0,4}/n?\d+$`),
		logger:             logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "402d469cae679c43f0ab9ef99a6b59dd"
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
		EnableSessionInit: false,
	}
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.nordstromrack.com"}
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
		u.Host = "www.nordstromrack.com"
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
	c.logger.Debugf("%s", resp.Request.URL.Path)

	if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
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
	Catalog struct {
		HasFilters bool `json:"hasFilters"`
		Filters    struct {
			Brands            []interface{} `json:"brands"`
			Categories        []interface{} `json:"categories"`
			Class             string        `json:"class"`
			Colors            []interface{} `json:"colors"`
			Context           interface{}   `json:"context"`
			Department        string        `json:"department"`
			Division          string        `json:"division"`
			IncludeFlash      bool          `json:"includeFlash"`
			IncludePersistent bool          `json:"includePersistent"`
			Limit             int           `json:"limit"`
			Page              int           `json:"page"`
			PriceRanges       []interface{} `json:"priceRanges"`
			Query             interface{}   `json:"query"`
			Shops             []interface{} `json:"shops"`
			Sizes             []interface{} `json:"sizes"`
			Sort              string        `json:"sort"`
			Subclass          interface{}   `json:"subclass"`
			NestedColors      bool          `json:"nestedColors"`
		} `json:"filters"`
		CatalogURLBase         string `json:"catalogUrlBase"`
		CurrentLoadedRowIndex  int    `json:"currentLoadedRowIndex"`
		IsBrandSearch          bool   `json:"isBrandSearch"`
		IsCustomCategorySearch bool   `json:"isCustomCategorySearch"`
		IsClearanceSearch      bool   `json:"isClearanceSearch"`
		IsLandingPage          bool   `json:"isLandingPage"`
		IsQuerySearch          bool   `json:"isQuerySearch"`
		IsQuickLookInProgress  bool   `json:"isQuickLookInProgress"`
		IsQuickLookVisible     bool   `json:"isQuickLookVisible"`
		IsShopsSearch          bool   `json:"isShopsSearch"`

		PageBase        string `json:"pageBase"`
		PageTitle       string `json:"pageTitle"`
		PageDescription string `json:"pageDescription"`
		Pages           []struct {
			Href       string `json:"href,omitempty"`
			IsCurrent  bool   `json:"isCurrent"`
			Label      string `json:"label"`
			PageNumber int    `json:"pageNumber,omitempty"`
		} `json:"pages"`
		Products []struct {
			AltImageSrc         string      `json:"altImageSrc,omitempty"`
			Brand               string      `json:"brand"`
			Color               string      `json:"color"`
			CustomerChoiceID    string      `json:"customerChoiceId"`
			EventID             interface{} `json:"eventId"`
			InitialImageSrc     string      `json:"initialImageSrc"`
			InventoryLevelLabel interface{} `json:"inventoryLevelLabel"`
			IsClearance         bool        `json:"isClearance"`
			IsInventoryLow      bool        `json:"isInventoryLow"`
			IsOnHold            bool        `json:"isOnHold"`
			IsSoldOut           bool        `json:"isSoldOut"`
			IsOnSale            bool        `json:"isOnSale"`
			IsClearTheRack      bool        `json:"isClearTheRack"`
			IsPriceVisible      bool        `json:"isPriceVisible"`
			ProductHref         string      `json:"productHref"`
			Source              string      `json:"source"`
			StyleID             int         `json:"styleId"`
			Title               string      `json:"title"`
			WebStyleID          interface{} `json:"webStyleId"`
			Prices              struct {
				Retail struct {
					Min float64 `json:"min"`
					Max float64 `json:"max"`
				} `json:"retail"`
				Regular struct {
					Min float64 `json:"min"`
					Max float64 `json:"max"`
				} `json:"regular"`
				Sale struct {
					Min float64 `json:"min"`
					Max float64 `json:"max"`
				} `json:"sale"`
			} `json:"prices"`
			Discounts struct {
				RetailToRegular struct {
					Min float64 `json:"min"`
					Max float64 `json:"max"`
				} `json:"retailToRegular"`
				RegularToSale struct {
					Min float64 `json:"min"`
					Max float64 `json:"max"`
				} `json:"regularToSale"`
				RetailToSale struct {
					Min float64 `json:"min"`
					Max float64 `json:"max"`
				} `json:"retailToSale"`
			} `json:"discounts"`
		} `json:"products"`
		QuickLookIndex int `json:"quickLookIndex"`
		Total          int `json:"total"`
	} `json:"catalog"`
}

// used to extract embaded json data in website page.
// more about golang regulation see here https://golang.org/pkg/regexp/syntax/
var productsExtractReg = regexp.MustCompile(`<script\s*>window.__INITIAL_STATE__\s*=\s*(.*)}}</script>`)

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

	var viewData CategoryView
	s := [][]byte{matched[1], []byte("}}")}
	bytesResult := bytes.Join(s, []byte(""))

	if err := json.Unmarshal(bytesResult, &viewData); err != nil {
		c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		return err
	}

	lastIndex := nextIndex(ctx)
	for _, idv := range viewData.Catalog.Products {

		req, err := http.NewRequest(http.MethodGet, idv.ProductHref, nil)
		if err != nil {
			c.logger.Errorf("load http request of url %s failed, error=%s", idv.ProductHref, err)
			return err
		}

		lastIndex += 1
		// set the index of the product crawled in the sub response
		nctx := context.WithValue(ctx, "item.index", lastIndex)
		// yield sub request
		if err := yield(nctx, req); err != nil {
			return err
		}
	}

	// get current page number
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}

	lastPageNo := len(viewData.Catalog.Pages)
	if lastPageNo < 2 {
		return nil
	}
	lastPageNo = viewData.Catalog.Pages[lastPageNo-2].PageNumber
	// check if this is the last page
	if page >= int64(lastPageNo) {
		return nil
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
	ProductPage struct {
		Product struct {
			AdditionalDetails     string `json:"additionalDetails"`
			AdditionalInformation string `json:"additionalInformation"`
			Availability          string `json:"availability"`
			BrandName             string `json:"brandName"`
			CanonicalPath         string `json:"canonicalPath"`
			Care                  string `json:"care"`
			Classification        string `json:"classification"`
			ClassificationUrls    struct {
				Division          string `json:"division"`
				Department        string `json:"department"`
				Classification    string `json:"classification"`
				SubClassification string `json:"subClassification"`
			} `json:"classificationUrls"`
			// ColorAvailabilityBySize map[string][]struct {

			// } `json:"colorAvailabilityBySize"`
			Colors []struct {
				IsClearTheRack bool `json:"isClearTheRack"`
				IsFinalSale    bool `json:"isFinalSale"`
				IsOnSale       bool `json:"isOnSale"`
				IsPriceVisible bool `json:"isPriceVisible"`
				Sizes          []struct {
					Sku            int    `json:"sku"`
					IsRmsSku       bool   `json:"isRmsSku"`
					RmsSku         int    `json:"rmsSku"`
					IsOnSale       bool   `json:"isOnSale"`
					IsClearTheRack bool   `json:"isClearTheRack"`
					IsPriceVisible bool   `json:"isPriceVisible"`
					Value          string `json:"value"`
					IsAvailable    bool   `json:"isAvailable"`
					IsLowQuantity  bool   `json:"isLowQuantity"`
					IsSoldOut      bool   `json:"isSoldOut"`
					IsClearance    bool   `json:"isClearance"`
					LowQuantity    int    `json:"lowQuantity"`
					Color          string `json:"color"`
					Prices         struct {
						Retail  float64 `json:"retail"`
						Regular float64 `json:"regular"`
						Sale    float64 `json:"sale"`
					} `json:"prices"`
					Discounts struct {
						RetailToRegular float64 `json:"retailToRegular"`
						RegularToSale   float64 `json:"regularToSale"`
						RetailToSale    float64 `json:"retailToSale"`
					} `json:"discounts"`
					IsReturnable bool   `json:"isReturnable"`
					StandardSize string `json:"standardSize"`
				} `json:"sizes"`
				Value                  string   `json:"value"`
				IsAvailable            bool     `json:"isAvailable"`
				IsLowQuantity          bool     `json:"isLowQuantity"`
				IsSoldOut              bool     `json:"isSoldOut"`
				IsClearance            bool     `json:"isClearance"`
				LowQuantity            int      `json:"lowQuantity"`
				ImageTemplates         []string `json:"imageTemplates"`
				OriginalImageTemplates []string `json:"originalImageTemplates"`
				Prices                 struct {
					Retail struct {
						Min float64 `json:"min"`
						Max float64 `json:"max"`
					} `json:"retail"`
					Regular struct {
						Min float64 `json:"min"`
						Max float64 `json:"max"`
					} `json:"regular"`
					Sale struct {
						Min float64 `json:"min"`
						Max float64 `json:"max"`
					} `json:"sale"`
				} `json:"prices"`
				Discounts struct {
					RetailToRegular struct {
						Min float64 `json:"min"`
						Max float64 `json:"max"`
					} `json:"retailToRegular"`
					RegularToSale struct {
						Min float64 `json:"min"`
						Max float64 `json:"max"`
					} `json:"regularToSale"`
					RetailToSale struct {
						Min float64 `json:"min"`
						Max float64 `json:"max"`
					} `json:"retailToSale"`
				} `json:"discounts"`
				Swatch string `json:"swatch"`
			} `json:"colors"`
			Department                      string        `json:"department"`
			Description                     string        `json:"description"`
			Division                        string        `json:"division"`
			ExtraNameCopy                   string        `json:"extraNameCopy"`
			FiberContent                    string        `json:"fiberContent"`
			IsChokeHazard                   bool          `json:"isChokeHazard"`
			IsFlash                         bool          `json:"isFlash"`
			IsPersistent                    bool          `json:"isPersistent"`
			IsQualifiedForExpeditedShipping bool          `json:"isQualifiedForExpeditedShipping"`
			IsQualifiedForFreeShipping      bool          `json:"isQualifiedForFreeShipping"`
			IsWebStyle                      bool          `json:"isWebStyle"`
			Material                        string        `json:"material"`
			Name                            string        `json:"name"`
			ShippingExclusions              []interface{} `json:"shippingExclusions"`
			SizeChartIframeSrc              string        `json:"sizeChartIframeSrc"`
			SizeChartSrc                    string        `json:"sizeChartSrc"`
			SizeInfo                        string        `json:"sizeInfo"`
			Skus                            []interface{} `json:"skus"`
			StyleNumber                     string        `json:"styleNumber"`
			StyleID                         int           `json:"styleId"`
			SubClassification               string        `json:"subClassification"`
			URITemplate                     string        `json:"uriTemplate"`
			WebStyleID                      int           `json:"webStyleId"`
			IsClearance                     bool          `json:"isClearance"`
			Prices                          struct {
				Retail struct {
					Min float64 `json:"min"`
					Max float64 `json:"max"`
				} `json:"retail"`
				Regular struct {
					Min float64 `json:"min"`
					Max float64 `json:"max"`
				} `json:"regular"`
				Sale struct {
					Min float64 `json:"min"`
					Max float64 `json:"max"`
				} `json:"sale"`
			} `json:"prices"`
			Discounts struct {
				RetailToRegular struct {
					Min float64 `json:"min"`
					Max float64 `json:"max"`
				} `json:"retailToRegular"`
				RegularToSale struct {
					Min float64 `json:"min"`
					Max float64 `json:"max"`
				} `json:"regularToSale"`
				RetailToSale struct {
					Min float64 `json:"min"`
					Max float64 `json:"max"`
				} `json:"retailToSale"`
			} `json:"discounts"`
			ShipTime string `json:"shipTime"`
		} `json:"product"`
		RecommendedProducts   []interface{} `json:"recommendedProducts"`
		SelectedAltImageIndex int           `json:"selectedAltImageIndex"`
		SelectedColor         string        `json:"selectedColor"`
		SelectedQuantity      interface{}   `json:"selectedQuantity"`
		BreadcrumbLinks       []struct {
			Href                  string   `json:"href"`
			NavigationBreadcrumbs []string `json:"navigationBreadcrumbs"`
			Text                  string   `json:"text"`
		} `json:"breadcrumbLinks"`
	} `json:"productPage"`
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
	matched := productsExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}
	// c.logger.Debugf("data: %s", matched[1])

	var viewData parseProductResponse
	s := [][]byte{matched[1], []byte("}}")}
	bytesResult := bytes.Join(s, []byte(""))

	if err := json.Unmarshal(bytesResult, &viewData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}
	canUrl := dom.Find(`link[rel="canonical"]`).AttrOr("href", "")
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           strconv.Format(viewData.ProductPage.Product.StyleID),
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
		},
		BrandName:   viewData.ProductPage.Product.BrandName,
		Title:       viewData.ProductPage.Product.Name,
		Description: viewData.ProductPage.Product.Description,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		// Stats: &pbItem.Stats{
		// 	ReviewCount: int32(p.NumberOfReviews),
		// 	Rating:      float32(p.ReviewAverageRating / 5.0),
		// },
	}
	links := viewData.ProductPage.BreadcrumbLinks
	for i, l := range links {
		switch i {
		case 0:
			item.Category = l.Text
		case 1:
			item.SubCategory = l.Text
		case 2:
			item.SubCategory2 = l.Text
		case 3:
			item.SubCategory3 = l.Text
		case 4:
			item.SubCategory4 = l.Text
		}
	}

	for _, rawSkuColor := range viewData.ProductPage.Product.Colors {
		for k, rawSku := range rawSkuColor.Sizes {
			currentPrice, _ := strconv.ParseFloat(rawSku.Prices.Sale)
			originalPrice, _ := strconv.ParseFloat(rawSku.Prices.Regular)
			discount, _ := strconv.ParseFloat(rawSku.Discounts.RetailToSale)
			sku := pbItem.Sku{
				SourceId: strconv.Format(rawSku.Sku),
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(currentPrice * 100),
					Msrp:     int32(originalPrice * 100),
					Discount: int32(discount * 100),
				},
				Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}
			if rawSku.LowQuantity > 0 {
				sku.Stock.StockStatus = pbItem.Stock_InStock
				sku.Stock.StockCount = int32(rawSku.LowQuantity)
			}

			// color
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    strconv.Format(rawSku.RmsSku),
				Name:  rawSku.Color,
				Value: rawSku.Color,
				Icon:  rawSkuColor.Swatch,
			})

			if k == 0 {
				// img based on color
				isDefault := true
				for ki, mid := range rawSkuColor.OriginalImageTemplates {
					if ki > 0 {
						isDefault = false
					}

					sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
						strconv.Format(rawSku.RmsSku+ki),
						mid,
						mid+"?height=1300&width=868",
						mid+"?height=750&width=500",
						mid+"?height=600&width=400",
						"",
						isDefault,
					))
				}
			}

			// size
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    strconv.Format(rawSku.Sku),
				Name:  rawSku.Value,
				Value: rawSku.Value,
			})

			item.SkuItems = append(item.SkuItems, &sku)
		}
	}

	// yield item result
	return yield(ctx, &item)
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		"https://www.nordstromrack.com/s/levis-512-%E2%84%A2-slim-taper-marshmallow-burn-out-jeans/n3298652?color=MARSHMALLOW%20BURNOUT%20DX&eid=482253",
		// "https://www.nordstromrack.com/s/free-people-riptide-tie-dye-print-t-shirt/n3327050?color=SEAFOAM%20COMBO",
		// "https://www.nordstromrack.com/shop/Women/Clothing/Tops",
		// "https://www.nordstromrack.com/events/472159",
		// "https://www.nordstromrack.com/shop/Women/Accessories/Hats,%20Gloves%20&%20Scarves/Gloves",
		// "https://www.nordstromrack.com/c/women/clothing/skirts/pencil",
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
