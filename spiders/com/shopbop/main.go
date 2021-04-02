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
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/voiladev/VoilaCrawl/pkg/crawler"
	"github.com/voiladev/VoilaCrawl/pkg/net/http"
	"github.com/voiladev/VoilaCrawl/pkg/net/http/cookiejar"
	"github.com/voiladev/VoilaCrawl/pkg/proxy"
	pbMedia "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	pbProxy "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
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
		categoryPathMatcher: regexp.MustCompile(`^(.*)/br/(.*)$`),
		// this regular used to match product page url path
		productPathMatcher: regexp.MustCompile(`^(.*)/vp/(.*)$`),
		logger:             logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "9e95fb1603d84f0387a2f75f9d4e85cf"
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
		Reliability:       pbProxy.ProxyReliability_ReliabilityMedium,
	}

	return opts
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.shopbop.com"}
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

// nextIndex used to get the index from the shared data.
// item.index is a const key for item index.
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")))
}

// used to extract embaded json data in website page.
// more about golang regulation see here https://golang.org/pkg/regexp/syntax/
var productsExtractReg = regexp.MustCompile(`(?U)type="sui-state"\s*data-key="pdp\.state">\s*({.*})\s*</script>`)

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

	if !bytes.Contains(respBody, []byte("data-productid")) {
		return errors.New("products not found")
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	lastIndex := nextIndex(ctx)
	sel := doc.Find(`.info.clearfix>a`)
	c.logger.Debugf("nodes %d", len(sel.Nodes))
	for i := range sel.Nodes {
		node := sel.Eq(i)
		if href, _ := node.Attr("href"); href != "" {
			//fmt.Println(string(href), " -- ", lastIndex)
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
	baseIndex, _ := strconv.ParseInt(resp.Request.URL.Query().Get("baseIndex"))

	// check if this is the last page
	if !bytes.Contains(respBody, []byte("class=\"next \"")) {
		return nil
	}

	// set pagination
	u := *resp.Request.URL
	vals := u.Query()
	vals.Set("baseIndex", strconv.Format(baseIndex+100))
	u.RawQuery = vals.Encode()

	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	// update the index of last page
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
}

type parseProductResponse struct {
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
			ModelSizes            []interface{} `json:"modelSizes"`
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
		ShortDescription      string      `json:"shortDescription"`
		LongDescription       string      `json:"longDescription"`
		SizeAndFitDetail      struct {
			SizeAndFitDescription    string        `json:"sizeAndFitDescription"`
			SizeAndFitNote           string        `json:"sizeAndFitNote"`
			MeasurementsFromSize     interface{}   `json:"measurementsFromSize"`
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
			FloweryDescription string        `json:"floweryDescription"`
		} `json:"longDescriptionDetail"`
		Proposition65                                   bool        `json:"proposition65"`
		Proposition65ChemicalsCancer                    interface{} `json:"proposition65ChemicalsCancer"`
		Proposition65ChemicalsReproductiveHarm          interface{} `json:"proposition65ChemicalsReproductiveHarm"`
		Proposition65ChemicalsCancerAndReproductiveHarm interface{} `json:"proposition65ChemicalsCancerAndReproductiveHarm"`
		GiftWithPurchase                                bool        `json:"giftWithPurchase"`
		HasAnyMeasurements                              bool        `json:"hasAnyMeasurements"`
		InStock                                         bool        `json:"inStock"`
		AltText                                         string      `json:"altText"`
	} `json:"product"`
}

// used to trim html labels in description
var htmlTrimRegp = regexp.MustCompile(`</?[^>]+>`)
var imageReg = regexp.MustCompile(`_UX[0-9]+_`)

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
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}

	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:       strconv.Format(viewData.Product.Sin),
			CrawlUrl: resp.Request.URL.String(),
		},
		BrandName:   viewData.Product.BrandLabel,
		Title:       viewData.Product.ShortDescription,
		Description: htmlTrimRegp.ReplaceAllString(viewData.Product.LongDescription, ", "),
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
	}
	for _, rawColor := range viewData.Product.StyleColors {

		originalPrice, _ := strconv.ParseFloat(rawColor.Prices[0].SaleAmount)
		msrp, _ := strconv.ParseFloat(rawColor.Prices[0].RetailAmount)
		discount, _ := strconv.ParseInt(rawColor.Prices[0].SalePercentage)

		for ks, rawSku := range rawColor.StyleColorSizes {
			sku := pbItem.Sku{
				SourceId: strconv.Format(viewData.Product.StyleNumber),
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(originalPrice * 100),
					Msrp:     int32(msrp * 100),
					Discount: int32(discount),
				},
				Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}
			if rawSku.InStock {
				sku.Stock.StockStatus = pbItem.Stock_InStock
				//sku.Stock.StockCount = int32(rawSku.TotalQuantityAvailable)
			}

			// color
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    strconv.Format(rawColor.Color.Code),
				Name:  rawColor.Color.Label,
				Value: rawColor.Color.Label,
				Icon:  rawColor.SwatchImage.URL,
			})

			if ks == 0 {
				isDefault := true
				for ki, m := range rawColor.Images {
					if ki > 0 {
						isDefault = false
					}
					template := m.URL
					sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
						strconv.Format(rawColor.StyleColorCode),
						template,
						imageReg.ReplaceAllString(template, "_UX700_"),
						imageReg.ReplaceAllString(template, "_UX500_"),
						imageReg.ReplaceAllString(template, "_UX400_"),
						"",
						isDefault,
					))
				}
			}

			// size
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    rawSku.Sin,
				Name:  rawSku.Size.Label,
				Value: rawSku.Size.Label,
			})

			item.SkuItems = append(item.SkuItems, &sku)
		}
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
		// "https://www.shopbop.com/active-clothing-shorts/br/v=1/65919.htm",
		"https://www.shopbop.com/hilary-bootie-sam-edelman/vp/v=1/1504954305.htm?folderID=15539&fm=other-shopbysize-viewall&os=false&colorId=1071C&ref_=SB_PLP_NB_12&breadcrumb=Sale%3EShoes",
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

	ctx := context.WithValue(context.Background(), "tracing_id", fmt.Sprintf("asos_%d", time.Now().UnixNano()))
	// start the crawl request
	for _, req := range spider.NewTestRequest(context.Background()) {
		if err := callback(ctx, req); err != nil {
			logger.Fatal(err)
		}
	}
}
