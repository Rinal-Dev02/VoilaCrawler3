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
		categoryPathMatcher: regexp.MustCompile(`^/(c)(/[a-z0-9-]+){2,6}$`),
		// this regular used to match product page url path
		productPathMatcher: regexp.MustCompile(`^(.*)/(Product-Variation)(.*)$`),
		logger:             logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "a233105f4d384fa2bcf56131653bac56"
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
	return []string{"www.saksoff5th.com"}
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
		bytes.Contains(respBody, []byte("Access Denied")) ||
		bytes.Contains(respBody, []byte("Your browser does not support cookies"))
}
func TrimSpaceNewlineInString(s []byte) []byte {
	re := regexp.MustCompile(`\n`)
	return re.ReplaceAll(s, []byte(" "))
}

const defaultCategoryProductsPageSize = 24

// nextIndex used to get the index from the shared data.
// item.index is a const key for item index.
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")))
}

// used to extract embaded json data in website page.
// more about golang regulation see here https://golang.org/pkg/regexp/syntax/
var productsExtractReg = regexp.MustCompile(`(?U)pageDataObj\s*=\s*({.*});?\s*</script>`)
var productsExtractOtherDetailReg = regexp.MustCompile(`({.*})`)

// parseCategoryProducts parse api url from web page url
func (c *_Crawler) parseCategoryProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if resp.StatusCode == http.StatusForbidden {
		return errors.New("access denied")
	}
	c.logger.Debugf("parse %s", resp.Request.URL)

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if isRobotCheckPage(respBody) {
		return errors.New("robot check page")
	}
	// c.logger.Debugf("%s", respBody)

	if !bytes.Contains(respBody, []byte("product bfx-disable-product standard")) {
		return errors.New("products not found")
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	lastIndex := nextIndex(ctx)
	sel := doc.Find(`.product.bfx-disable-product.standard`)
	c.logger.Debugf("nodes %d", len(sel.Nodes))
	for i := range sel.Nodes {
		node := sel.Eq(i)
		//colorSwatches := node.Find(`.d-none.d-lg-block>a`)
		colorSwatches := node.Find(`.hover-content>.color-swatches>.swatches>.d-none.d-lg-block>a`)

		if len(colorSwatches.Nodes) > 0 { // color variations
			for j := range colorSwatches.Nodes {
				nodeV := colorSwatches.Eq(j)

				if href, _ := nodeV.Attr("data-valueurl"); href != "" {
					c.logger.Debugf("yield %s", href)
					// req, err := http.NewRequest(http.MethodGet, href, nil)
					// if err != nil {
					// 	c.logger.Error(err)
					// 	continue
					// }
					// lastIndex += 1
					// nctx := context.WithValue(ctx, "item.index", lastIndex)
					// if err := yield(nctx, req); err != nil {
					// 	return err
					// }
				}
			}
		} else {
			colorSwatches, _ := node.Attr("data-pid")
			href := "https://www.saksoff5th.com/on/demandware.store/Sites-SaksOff5th-Site/en_US/Product-Variation?pid=" + colorSwatches + "&quantity=1"

			if colorSwatches != "" {
				c.logger.Debugf("yield %s", href)
				// req, err := http.NewRequest(http.MethodGet, href, nil)
				// if err != nil {
				// 	c.logger.Error(err)
				// 	continue
				// }
				// lastIndex += 1
				// nctx := context.WithValue(ctx, "item.index", lastIndex)
				// if err := yield(nctx, req); err != nil {
				// 	return err
				// }
			}
		}
	}
	if len(sel.Nodes) < defaultCategoryProductsPageSize {
		return nil
	}

	if bytes.Contains(respBody, []byte("aria-label=\"Next\"")) {
		// More Results
	} else if bytes.Contains(respBody, []byte("\"show-more invisible\">")) {
		// Next
	} else {
		return nil // no next page
	}

	var start int64 = 0
	if p := resp.Request.URL.Query().Get("start"); p != "" {
		start, _ = strconv.ParseInt(p)
	}

	u := resp.Request.URL
	vals := u.Query()
	vals.Set("start", strconv.Format(start+defaultCategoryProductsPageSize))
	vals.Set("sz", strconv.Format(defaultCategoryProductsPageSize))
	u.RawQuery = vals.Encode()

	nctx := context.WithValue(ctx, "item.index", lastIndex)
	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	return yield(nctx, req)
}

// used to trim html labels in description
var htmlTrimRegp = regexp.MustCompile(`</?[^>]+>`)

// below are the golang json data struct of raw website.
// if you get the raw json data of the website,
// then you can use https://mholt.github.io/json-to-go/ to convert it to a golang struct

type RawProductDetail struct {
	Products []struct {
		AverageRating string `json:"average_rating"`
		Brand         string `json:"brand"`
		Code          string `json:"code"`
		Name          string `json:"name"`
		OriginalPrice string `json:"original_price"`
		Price         string `json:"price"`
		Skus          []struct {
			AvailableDc string `json:"available_dc"`
			Sku         string `json:"sku"`
		} `json:"skus"`
		Tags struct {
			FeatureType    string `json:"feature_type"`
			InventoryLabel string `json:"inventory_label"`
			PipText        string `json:"pip_text"`
			PriceType      string `json:"price_type"`
			PublishDate    string `json:"publish_date"`
			Returnable     string `json:"returnable"`
		} `json:"tags"`
		TotalReviews string `json:"total_reviews"`
	} `json:"products"`
}

type OtherProductDetail struct {
	Context     string `json:"@context"`
	Type        string `json:"@type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Mpn         string `json:"mpn"`
	Sku         string `json:"sku"`
	Gtin13      string `json:"gtin13"`
	Brand       struct {
		Type string `json:"@type"`
		Name string `json:"name"`
	} `json:"brand"`
	Image  []string `json:"image"`
	Offers struct {
		URL           string `json:"url"`
		Type          string `json:"@type"`
		PriceCurrency string `json:"priceCurrency"`
		Price         string `json:"price"`
		Availability  string `json:"availability"`
	} `json:"offers"`
}

type ProductDataJson struct {
	Action      string `json:"action"`
	QueryString string `json:"queryString"`
	Locale      string `json:"locale"`
	Product     struct {
		MasterProductID string `json:"masterProductID"`
		Brand           struct {
			Name string `json:"name"`
			URL  struct {
			} `json:"url"`
		} `json:"brand"`
		UUID                 string      `json:"uuid"`
		ID                   string      `json:"id"`
		ProductName          string      `json:"productName"`
		ProductType          string      `json:"productType"`
		Purchaselimit        interface{} `json:"purchaselimit"`
		LongDescriptionStyle string      `json:"longDescriptionStyle"`
		DropShipShipping     struct {
			Name        interface{} `json:"name"`
			DisplayName string      `json:"displayName"`
			URL         struct {
			} `json:"url"`
		} `json:"DropShipShipping"`
		UspsShipOK               interface{}   `json:"uspsShipOK"`
		PdRestrictedShipTypeText interface{}   `json:"pdRestrictedShipTypeText"`
		MRecommendations         []interface{} `json:"mRecommendations"`
		FeaturedType             struct {
			Value interface{} `json:"value"`
			Color string      `json:"color"`
		} `json:"featuredType"`
		IsNotReturnable struct {
			Value bool   `json:"value"`
			Color string `json:"color"`
		} `json:"isNotReturnable"`
		Badge struct {
			IsNew struct {
				Value bool   `json:"value"`
				Color string `json:"color"`
			} `json:"isNew"`
			IsSale struct {
				Value bool   `json:"value"`
				Color string `json:"color"`
			} `json:"isSale"`
			IsClearance          bool   `json:"isClearance"`
			IsFinalSale          bool   `json:"isFinalSale"`
			LimitedInvBadgeColor string `json:"limitedInvBadgeColor"`
		} `json:"badge"`
		DisplayQuicklook  string `json:"displayQuicklook"`
		Wishlist          string `json:"wishlist"`
		SizeChartTemplate string `json:"sizeChartTemplate"`
		PlpPromos         struct {
		} `json:"plpPromos"`
		PdRestrictedWarningText bool   `json:"pdRestrictedWarningText"`
		PdpURL                  string `json:"pdpURL"`
		Price                   struct {
			Sales struct {
				Value              float64 `json:"value"`
				Currency           string  `json:"currency"`
				Formatted          string  `json:"formatted"`
				DecimalPrice       string  `json:"decimalPrice"`
				FormatAmount       string  `json:"formatAmount"`
				PriceBandFormatted string  `json:"priceBandFormatted"`
			} `json:"sales"`
			List struct {
				Value              float64 `json:"value"`
				Currency           string  `json:"currency"`
				Formatted          string  `json:"formatted"`
				DecimalPrice       string  `json:"decimalPrice"`
				FormatAmount       string  `json:"formatAmount"`
				PriceBandFormatted string  `json:"priceBandFormatted"`
			} `json:"list"`
			Savings        float64 `json:"savings"`
			SavePercentage string  `json:"savePercentage"`
			HTML           string  `json:"html"`
		} `json:"price"`
		Images struct {
			Large []struct {
				Alt      string `json:"alt"`
				URL      string `json:"url"`
				Title    string `json:"title"`
				HiresURL string `json:"hiresURL"`
			} `json:"large"`
			Small []struct {
				Alt      string `json:"alt"`
				URL      string `json:"url"`
				Title    string `json:"title"`
				HiresURL string `json:"hiresURL"`
			} `json:"small"`
			HiRes []struct {
				Alt      string `json:"alt"`
				URL      string `json:"url"`
				Title    string `json:"title"`
				HiresURL string `json:"hiresURL"`
			} `json:"hi-res"`
			Swatch []struct {
				Alt      string `json:"alt"`
				URL      string `json:"url"`
				Title    string `json:"title"`
				HiresURL string `json:"hiresURL"`
			} `json:"swatch"`
			Video []struct {
				Alt      string `json:"alt"`
				URL      string `json:"url"`
				Title    string `json:"title"`
				HiresURL struct {
				} `json:"hiresURL"`
			} `json:"video"`
		} `json:"images"`
		SelectedQuantity    int `json:"selectedQuantity"`
		MinOrderQuantity    int `json:"minOrderQuantity"`
		MaxOrderQuantity    int `json:"maxOrderQuantity"`
		VariationAttributes []struct {
			AttributeID string `json:"attributeId"`
			DisplayName string `json:"displayName"`
			ID          string `json:"id"`
			Swatchable  bool   `json:"swatchable"`
			Values      []struct {
				ID           string      `json:"id"`
				Description  interface{} `json:"description"`
				DisplayValue string      `json:"displayValue"`
				Value        string      `json:"value"`
				Selected     bool        `json:"selected"`
				Selectable   bool        `json:"selectable"`
				URL          string      `json:"url"`
				Images       struct {
					Swatch []struct {
						Alt      string `json:"alt"`
						URL      string `json:"url"`
						Title    string `json:"title"`
						HiresURL string `json:"hiresURL"`
					} `json:"swatch"`
				} `json:"images"`
			} `json:"values"`
			SelectedAttribute struct {
			} `json:"selectedAttribute"`
			AttrDisplay            string `json:"attrDisplay"`
			AttrEditDisplay        string `json:"attrEditDisplay"`
			SelectedSizeClass      string `json:"selectedSizeClass"`
			AttributeSelectedValue string `json:"attributeSelectedValue"`
			ResetURL               string `json:"resetUrl,omitempty"`
		} `json:"variationAttributes"`
		LongDescription  string      `json:"longDescription"`
		ShortDescription interface{} `json:"shortDescription"`
		Rating           int         `json:"rating"`
		Promotions       interface{} `json:"promotions"`
		Attributes       []struct {
			ID         string `json:"ID"`
			Name       string `json:"name"`
			Attributes []struct {
				Label string   `json:"label"`
				Value []string `json:"value"`
			} `json:"attributes"`
		} `json:"attributes"`
		Availability struct {
			Messages                 []interface{} `json:"messages"`
			ButtonName               string        `json:"buttonName"`
			IsInPurchaselimit        bool          `json:"isInPurchaselimit"`
			IsInPurchaselimitMessage string        `json:"isInPurchaselimitMessage"`
			IsAboveThresholdLevel    bool          `json:"isAboveThresholdLevel"`
			HexColorCode             struct {
			} `json:"hexColorCode"`
			Outofstockmessage string      `json:"outofstockmessage"`
			InStockDate       interface{} `json:"inStockDate"`
		} `json:"availability"`
		Available                   bool          `json:"available"`
		OrderableNotInPurchaselimit bool          `json:"orderableNotInPurchaselimit"`
		Options                     []interface{} `json:"options"`
		Quantities                  []struct {
			Value    string `json:"value"`
			Selected bool   `json:"selected"`
			URL      string `json:"url"`
		} `json:"quantities"`
		SizeChartID        interface{} `json:"sizeChartId"`
		SelectedProductURL string      `json:"selectedProductUrl"`
		ReadyToOrder       bool        `json:"readyToOrder"`
		ReadyToOrderMsg    string      `json:"readyToOrderMsg"`
		Online             bool        `json:"online"`
		PageTitle          interface{} `json:"pageTitle"`
		PageDescription    interface{} `json:"pageDescription"`
		PageKeywords       interface{} `json:"pageKeywords"`
		PageMetaTags       []struct {
		} `json:"pageMetaTags"`
		Template                  interface{} `json:"template"`
		SearchableIfUnavailable   bool        `json:"searchableIfUnavailable"`
		HbcProductType            string      `json:"hbcProductType"`
		Waitlistable              bool        `json:"waitlistable"`
		DropShipInd               bool        `json:"dropShipInd"`
		DiscountAppliedInCheckout bool        `json:"discountAppliedInCheckout"`
		HudsonPoint               int         `json:"hudsonPoint"`
		SpdCollectionName         string      `json:"spdCollectionName"`
		IsAvailableForInstore     bool        `json:"isAvailableForInstore"`
		IsReveiwable              bool        `json:"isReveiwable"`
		TurntoReviewCount         int         `json:"turntoReviewCount"`
		PromotionalPricing        struct {
			IsPromotionalPrice bool   `json:"isPromotionalPrice"`
			PromoMessage       string `json:"promoMessage"`
			PriceHTML          string `json:"priceHtml"`
		} `json:"promotionalPricing"`
		GwpButtonCopy               interface{} `json:"gwpButtonCopy"`
		GwpLink                     interface{} `json:"gwpLink"`
		AllAvailableProductsSoldOut bool        `json:"allAvailableProductsSoldOut"`
		AllAvailableProducts        []struct {
			AvailableDc string `json:"available_dc"`
			Sku         string `json:"sku"`
		} `json:"allAvailableProducts"`
		StarRating     string    `json:"starRating"`
		AttributesHTML time.Time `json:"attributesHtml"`
		PromotionsHTML string    `json:"promotionsHtml"`
		FinalSaleHTML  string    `json:"finalSaleHtml"`
	} `json:"product"`
	Resources struct {
		InfoSelectforstock    string `json:"info_selectforstock"`
		AssistiveSelectedText string `json:"assistiveSelectedText"`
		Soldout               string `json:"soldout"`
		Addtocart             string `json:"addtocart"`
		LimitedInventory      string `json:"limitedInventory"`
		Movetobag             string `json:"movetobag"`
		Addtobag              string `json:"addtobag"`
	} `json:"resources"`
	AvailabilityURL        string `json:"availabilityUrl"`
	AvailabilityPromptText string `json:"availabilityPromptText"`
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
	respBody = TrimSpaceNewlineInString(respBody)

	matched := productsExtractOtherDetailReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}
	// c.logger.Debugf("data: %s", matched[1])

	var viewData ProductDataJson

	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		//return err //check
	}

	rating, _ := strconv.ParseFloat(viewData.Product.StarRating)
	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:       strconv.Format(viewData.Product.MasterProductID),
			CrawlUrl: resp.Request.URL.String(),
		},
		BrandName:   viewData.Product.Brand.Name,
		Title:       viewData.Product.ProductName,
		Description: htmlTrimRegp.ReplaceAllString(viewData.Product.LongDescription, ""),
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stats: &pbItem.Stats{
			ReviewCount: int32(viewData.Product.TurntoReviewCount),
			Rating:      float32(rating),
		},
	}

	originalPrice, _ := strconv.ParseFloat(viewData.Product.Price.Sales.Value)
	msrp, _ := strconv.ParseFloat(viewData.Product.Price.List.Value)
	discount, _ := strconv.ParseInt(viewData.Product.Price.SavePercentage)
	colorIndex := -1
	sizeIndex := -1

	for kv, rawVariation := range viewData.Product.VariationAttributes {
		if rawVariation.ID == "size" {
			sizeIndex = kv
		}
		if rawVariation.ID == "color" {
			colorIndex = kv
		}
	}

	for kv, rawSku := range viewData.Product.VariationAttributes[sizeIndex].Values {

		sku := pbItem.Sku{
			SourceId: strconv.Format(rawSku.ID),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(originalPrice * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},
			Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}

		if rawSku.Selectable {
			sku.Stock.StockStatus = pbItem.Stock_InStock
			//sku.Stock.StockCount = int32(rawSku.AvailableDc)
		}

		// color
		if colorIndex > -1 {
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    strconv.Format(colorIndex),
				Name:  viewData.Product.VariationAttributes[colorIndex].AttributeSelectedValue,
				Value: viewData.Product.VariationAttributes[colorIndex].AttributeSelectedValue,
				//Icon:  color.SwatchMedia.Mobile,
			})
		}

		if kv == 0 {

			isDefault := true
			for ki, mid := range viewData.Product.Images.HiRes {
				template := mid.URL
				if ki > 0 {
					isDefault = false
				}
				sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
					strconv.Format(ki),
					template,
					strings.ReplaceAll(template, "", ""),
					strings.ReplaceAll(template, "", ""),
					strings.ReplaceAll(template, "", ""),
					"",
					isDefault,
				))
			}
		}

		// size
		if sizeIndex > -1 {
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    rawSku.ID,
				Name:  rawSku.DisplayValue,
				Value: rawSku.Value,
			})
		}

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
		// "https://www.saksoff5th.com/c/men/shoes/drivers",
		"https://www.saksoff5th.com/on/demandware.store/Sites-SaksOff5th-Site/en_US/Product-Variation?dwvar_0400012201537_color=BLACK&pid=0400012201537&quantity=1",
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
	os.Setenv("VOILA_PROXY_URL", "http://3.239.93.53:30216")
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
				i.URL.Host = "www.saksoff5th.com"
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
