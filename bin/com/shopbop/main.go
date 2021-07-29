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
	"github.com/voiladev/VoilaCrawler/pkg/cli"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	pbMedia "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	pbProxy "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
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
func (_ *_Crawler) New(_ *cli.Context, client http.Client, logger glog.Log) (crawler.Crawler, error) {
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
	return "bfa928d05c46dd3f4ca707cfd39803a0"
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
		EnableSessionInit: true,
		Reliability:       pbProxy.ProxyReliability_ReliabilityMedium,
	}
	opts.MustCookies = append(opts.MustCookies,
		&http.Cookie{Name: "llc", Value: "US-EN-USD", Path: "/"},
		&http.Cookie{Name: "s_cc", Value: "false", Path: "/"},
		&http.Cookie{Name: "showEmailPopUp", Value: fmt.Sprintf("false_%d", time.Now().UnixNano()/1000000), Path: "/"},
	)
	return opts
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.shopbop.com"}
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
		u.Host = "www.shopbop.com"
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

	if resp.RawUrl().Path == "" || resp.RawUrl().Path == "/" {
		return c.parseCategories(ctx, resp, yield)
	}
	if c.categoryPathMatcher.MatchString(resp.RawUrl().Path) {

		return c.parseCategoryProducts(ctx, resp, yield)
	} else if c.productPathMatcher.MatchString(resp.RawUrl().Path) {
		return c.parseProduct(ctx, resp, yield)
	}
	return crawler.ErrUnsupportedPath
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

func (c *_Crawler) parseCategories(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	dom, err := resp.Selector()

	if err != nil {
		c.logger.Error(err)
		return err
	}

	sel := dom.Find(`.top-nav-list-item`)
	for j := range sel.Nodes {
		subnode := sel.Eq(j)

		nnctx := context.WithValue(ctx, "Category", TrimSpaceNewlineInString(subnode.Find(`.top-nav-list-item-link span`).Text()))

		subnodes1 := subnode.Find(`.sub-navigation-list>.sub-navigation-list-item`)
		for k := range subnodes1.Nodes {
			sub2node := subnodes1.Eq(k)

			subnodes2 := sub2node.Find(`.sub-navigation-list-item-link`)
			SubCategory := TrimSpaceNewlineInString(subnodes2.Text())
			//fmt.Println(SubCategory)
			href := subnodes2.AttrOr("href", "")
			if href == "" {
				continue
			}
			u, err := url.Parse(href)
			if err != nil {
				c.logger.Errorf("parse url %s failed", href)
				continue
			}

			if c.categoryPathMatcher.MatchString(u.Path) {
				// here reset tracing id to distiguish different category crawl
				// This may exists duplicate requests
				nctx := context.WithValue(nnctx, "SubCategory", SubCategory)
				req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
				if err := yield(nctx, req); err != nil {
					return err

				}
			}
		}
	}
	return nil
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
	sel := doc.Find(`#product-container .product`)
	c.logger.Debugf("nodes %d", len(sel.Nodes))
	for i := range sel.Nodes {
		node := sel.Eq(i)
		if href, _ := node.Find(`.url`).Attr("href"); href != "" &&
			(strings.HasPrefix(href, "/") || strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://")) {
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

	nextNode := doc.Find(`.pages .next`)
	if strings.Contains(nextNode.AttrOr("class", ""), "disabled") {
		return nil
	}
	nextUrl := nextNode.AttrOr("data-next-link", "")
	if nextUrl == "" {
		return nil
	}

	req, _ := http.NewRequest(http.MethodGet, nextUrl, nil)
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
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}
	// c.logger.Debugf("data: %s", matched[1])

	var viewData parseProductResponse
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
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
	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           strconv.Format(viewData.Product.Sin),
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
		},
		BrandName:   viewData.Product.BrandLabel,
		Title:       viewData.Product.ShortDescription,
		Description: viewData.Product.LongDescription,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
	}
	sel := dom.Find(`#bread-crumbs .bread-crumb-list>li[itemprop="itemListElement"]`)
	for i := range sel.Nodes {
		node := sel.Eq(i)
		switch i {
		case 0:
			item.Category = node.Find(`span[itemprop="name"]`).Text()
		case 1:
			item.SubCategory = node.Find(`span[itemprop="name"]`).Text()
		case 2:
			item.SubCategory2 = node.Find(`span[itemprop="name"]`).Text()
		case 3:
			item.SubCategory3 = node.Find(`span[itemprop="name"]`).Text()
		case 4:
			item.SubCategory4 = node.Find(`span[itemprop="name"]`).Text()
		}
	}

	for _, rawColor := range viewData.Product.StyleColors {
		originalPrice, _ := strconv.ParseFloat(rawColor.Prices[0].SaleAmount)
		msrp, _ := strconv.ParseFloat(rawColor.Prices[0].RetailAmount)
		discount, _ := strconv.ParseInt(rawColor.Prices[0].SalePercentage)

		var medias []*pbMedia.Media
		for ki, m := range rawColor.Images {
			template := m.URL
			medias = append(medias, pbMedia.NewImageMedia(
				"",
				template,
				imageReg.ReplaceAllString(template, "_UX1000_"),
				imageReg.ReplaceAllString(template, "_UX600_"),
				imageReg.ReplaceAllString(template, "_UX500_"),
				"",
				ki == 0,
			))
		}

		colorSpec := pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecColor,
			Id:    strconv.Format(rawColor.Color.Code),
			Name:  rawColor.Color.Label,
			Value: rawColor.Color.Label,
			Icon:  rawColor.SwatchImage.URL,
		}

		for _, rawSku := range rawColor.StyleColorSizes {
			sku := pbItem.Sku{
				SourceId: strconv.Format(rawSku.SkuCode),
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(originalPrice * 100),
					Msrp:     int32(msrp * 100),
					Discount: int32(discount),
				},
				Medias: medias,
				Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}
			if rawSku.InStock {
				sku.Stock.StockStatus = pbItem.Stock_InStock
			}
			sku.Specs = append(sku.Specs, &colorSpec)
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    rawSku.Size.Code,
				Name:  rawSku.Size.Label,
				Value: rawSku.Size.Code,
			})
			item.SkuItems = append(item.SkuItems, &sku)
		}
	}
	for _, rawSku := range item.SkuItems {
		if rawSku.Stock.StockStatus == pbItem.Stock_InStock {
			item.Stock = &pbItem.Stock{StockStatus: pbItem.Stock_InStock}
			break
		}
	}
	if item.Stock == nil {
		item.Stock = &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock}
	}
	if err = yield(ctx, &item); err != nil {
		return err
	}
	return nil
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://www.shopbop.com/",
		// "https://www.shopbop.com/active-clothing-shorts/br/v=1/65919.htm",
		// "https://www.shopbop.com/active-bags/br/v=1/65741.htm",
		// "https://www.shopbop.com/hilary-bootie-sam-edelman/vp/v=1/1504954305.htm?folderID=15539&fm=other-shopbysize-viewall&os=false&colorId=1071C&ref_=SB_PLP_NB_12&breadcrumb=Sale%3EShoes",
		// "https://www.shopbop.com/recycled-ripstop-quilt-coat-ganni/vp/v=1/1569753880.htm?folderID=64802&fm=other-shopbysize-viewall&os=false&colorId=1A608&ref_=SB_PLP_NB_1&breadcrumb=Clothing%3EJackets",
		"https://www.shopbop.com/medium-weekender-marc-jacobs/vp/v=1/1552457368.htm?folderID=65741&fm=other-viewall&os=false&colorId=1A039&ref_=SB_PLP_DB_10&breadcrumb=Active%3EBags",
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
