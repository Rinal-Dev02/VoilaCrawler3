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
	"github.com/robertkrimen/otto"
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
		categoryPathMatcher: regexp.MustCompile(`^/en_usd(/[a-zA-Z0-9-._]+){1,6}.html$`),
		// this regular used to match product page url path
		productPathMatcher: regexp.MustCompile(`^/en_usd(/[a-zA-Z0-9-._]+){1,6}[0-9]+.html$`),
		logger:             logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "c6fa53e4dd3c4cf0b6f410bb8f5015ba"
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
	return []string{"*.stories.com"}
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

func isRobotCheckPage(respBody []byte) bool {
	return bytes.Contains(respBody, []byte("we believe you are using automation tools to browse the website")) ||
		bytes.Contains(respBody, []byte("Javascript is disabled or blocked by an extension")) ||
		bytes.Contains(respBody, []byte("Your browser does not support cookies"))
}

// used to extract embaded json data in website page.
// more about golang regulation see here https://golang.org/pkg/regexp/syntax/
var productsExtractReg = regexp.MustCompile(`(?U)var\s*productArticleDetails\s*=\s*({.*});\s*`)
var articleCodeReg = regexp.MustCompile(`(\d)+`)

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

	if !bytes.Contains(respBody, []byte("class=\"a-link no-styling")) {
		return errors.New("products not found")
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	lastIndex := nextIndex(ctx)
	sel := doc.Find(`.a-link.no-styling`)

	c.logger.Debugf("nodes %d", len(sel.Nodes))
	for i := range sel.Nodes {
		node := sel.Eq(i)
		if href, _ := node.Attr("href"); href != "" {

			//c.logger.Debugf("yield %v --> %s", lastIndex, href)
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

	ps, _ := doc.Find(`#productStart`).Attr(`class`)
	productStart, _ := strconv.ParseInt(ps)

	ppp, _ := doc.Find(`#productPerPage`).Attr(`class`)
	productPerPage, _ := strconv.ParseInt(ppp)

	pc, _ := doc.Find(`#productCount`).Attr(`class`)
	productCount, _ := strconv.ParseInt(pc)

	if productCount <= productPerPage {
		return nil
	} else if (productStart + productPerPage) >= productCount {
		return nil
	}

	nexturl, _ := doc.Find(`#productPath`).Attr(`class`)

	nexturl = "https://www.stories.com" + nexturl + "?start=" + strconv.Format((productStart + productPerPage))

	req, _ := http.NewRequest(http.MethodGet, nexturl, nil)
	// update the index of last page
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
}

type Article struct {
	Title             string `json:"title"`
	Name              string `json:"name"`
	ColorCode         string `json:"colorCode"`
	Description       string `json:"description"`
	ArticleWeight     string `json:"articleWeight"`
	ArticleWeightUnit string `json:"articleWeightUnit"`
	Volume            string `json:"volume"`
	VolumeUnit        string `json:"volumeUnit"`
	ComparativePrice  struct {
		Price            string `json:"price"`
		ComparativeValue string `json:"comparativeValue"`
		ComparativeUnit  string `json:"comparativeUnit"`
		FormattedValue   string `json:"formattedValue"`
	} `json:"comparativePrice"`
	AtelierName        string `json:"atelierName"`
	PercentageDiscount string `json:"percentageDiscount"`
	BrandName          string `json:"brandName"`
	ColorLoc           string `json:"colorLoc"`
	PdpLink            string `json:"pdpLink"`
	OriginCountry      string `json:"originCountry"`
	StyleWithArticles  []struct {
		Code      string `json:"code"`
		Name      string `json:"name"`
		BrandName string `json:"brandName"`
		URL       string `json:"url"`
		ImageURL  string `json:"imageUrl"`
		ImageAlt  string `json:"imageAlt"`
		Price     string `json:"price"`
		//PriceOriginal  bool        `json:"priceOriginal"`
		PriceValue     string      `json:"priceValue"`
		PriceSaleValue interface{} `json:"priceSaleValue"`
		ColorName      string      `json:"colorName"`
		Color          []struct {
			ColorAlt string `json:"colorAlt"`
			ColorSrc string `json:"colorSrc"`
		} `json:"color"`
		Marker []interface{} `json:"marker"`
	} `json:"styleWithArticles"`
	Variants []struct {
		VariantCode string `json:"variantCode"`
		SizeCode    string `json:"sizeCode"`
		SizeName    string `json:"sizeName"`
	} `json:"variants"`
	ProductFrontImages []interface{} `json:"productFrontImages"`
	LogoImages         []interface{} `json:"logoImages"`
	DataSheetImages    []interface{} `json:"dataSheetImages"`
	ThumbnailImages    []interface{} `json:"thumbnailImages"`
	OtherImages        []interface{} `json:"otherImages"`
	NormalImages       []struct {
		Thumbnail  string `json:"thumbnail"`
		Image      string `json:"image"`
		Fullscreen string `json:"fullscreen"`
		Zoom       string `json:"zoom"`
	} `json:"normalImages"`
	DetailImages []interface{} `json:"detailImages"`
	Images       []interface{} `json:"images"`
	VAssets      []struct {
		Thumbnail  string `json:"thumbnail"`
		Image      string `json:"image"`
		Fullscreen string `json:"fullscreen"`
		Zoom       string `json:"zoom"`
	} `json:"vAssets"`
	Price string `json:"price"`
	//PriceOriginal    string `json:"priceOriginal"`
	PriceValue       string `json:"priceValue"`
	PriceSaleValue   string `json:"priceSaleValue"`
	MarketingMarkers []struct {
		URL         string `json:"url"`
		Alt         string `json:"alt"`
		MarkerTxt   string `json:"markerTxt"`
		MarkerColor string `json:"markerColor"`
		Style       string `json:"style"`
	} `json:"marketingMarkers"`
	PromoMarkerURL       string   `json:"promoMarkerUrl"`
	PromoMarkerAlt       string   `json:"promoMarkerAlt"`
	PromoMarkerText      string   `json:"promoMarkerText"`
	PromoMarkerLegalText string   `json:"promoMarkerLegalText"`
	PromoMarkerLabelText string   `json:"promoMarkerLabelText"`
	PromoMarkerStyle     string   `json:"promoMarkerStyle"`
	Compositions         []string `json:"compositions"`
	CareInstructions     []string `json:"careInstructions"`
	URL                  string   `json:"url"`
}

type parseProductResponse struct {
	ArticleCode         string   `json:"articleCode"`
	BaseProductCode     string   `json:"baseProductCode"`
	AncestorProductCode string   `json:"ancestorProductCode"`
	MainCategorySummary string   `json:"mainCategorySummary"`
	Name                string   `json:"name"`
	StyleWithArticles   []string `json:"styleWithArticles"`
}

type parseProductsAvailability struct {
	Availability []string `json:"availability"`
}

func DecodeResponse(respBody string) ([]Article, error) {
	//var art []Article
	art := make([]Article, 0)
	ret := map[string]json.RawMessage{}
	if err := json.Unmarshal([]byte(respBody), &ret); err != nil {
		fmt.Println(err)
		//return art, err
	}

	for key, msg := range ret {
		if regexp.MustCompile(`[0-9]+`).MatchString(key) {
			var (
				rawData, _ = msg.MarshalJSON()
				article    Article
			)
			if err := json.Unmarshal(rawData, &article); err != nil {
				continue
			}
			art = append(art, article)
		}
	}
	return art, nil
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

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	respBody = bytes.ReplaceAll(respBody, []byte("\n"), []byte(""))
	matched := productsExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}

	var (
		viewData parseProductResponse
		q        parseProductsAvailability
	)

	vm := otto.New()
	jsonStr := "var productArticleDetails = " + string(matched[1])
	_, err = vm.Run(jsonStr)

	vm.Run(`
		obj = JSON.stringify(productArticleDetails);		
	`)

	responseJS := ""
	value, err := vm.Get("obj")
	{
		responseJS, _ = value.ToString()

		if err := json.Unmarshal([]byte(responseJS), &viewData); err != nil {
			c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
			return err
		}
	}

	article, err := DecodeResponse(responseJS)

	availabilityURL := "https://www.stories.com/webservices_stories/service/product/stories-us/availability/" + viewData.AncestorProductCode + ".json"

	req, err := http.NewRequest(http.MethodGet, availabilityURL, nil)

	availreq, err := c.httpClient.Do(ctx, req)
	if err != nil {
		panic(err)
	}
	defer availreq.Body.Close()

	respBodyImg, err := ioutil.ReadAll(availreq.Body)
	if err != nil {
		c.logger.Error(err)
		return err
	}

	if err = json.Unmarshal(respBodyImg, &q); err != nil {
		c.logger.Debugf("parse %s failed, error=%s", respBodyImg, err)
		return err
	}

	for i := range article {

		// build product data
		item := pbItem.Product{
			Source: &pbItem.Source{
				Id:       strconv.Format(i),
				CrawlUrl: resp.Request.URL.String(),
			},
			BrandName:   article[i].BrandName,
			Title:       article[i].Name,
			Description: htmlTrimRegp.ReplaceAllString(article[i].Description, ""),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
			},
		}

		originalPrice, _ := strconv.ParseFloat(article[i].PriceSaleValue)
		MSRP, _ := strconv.ParseFloat(article[i].PriceValue)
		discount, _ := strconv.ParseInt(strings.TrimPrefix(strings.TrimSuffix(article[i].PercentageDiscount, "%"), "-"))
		if originalPrice == 0.0 {
			originalPrice = MSRP
		}

		sel := doc.Find(`.m-breadcrumb.u-align-to-logo.pdp-breadcrumb.new-breadcrumb>ol>li`)
		for b := range sel.Nodes {
			cate := strings.TrimSpace(sel.Eq(b).Text())
			switch b {
			case 1:
				item.Category = cate
			case 2:
				item.SubCategory = cate
			case 3:
				item.SubCategory2 = cate
			}
		}

		for _, v := range []string{"woman", "women", "female"} {
			if strings.Contains(strings.ToLower(item.Category), v) {
				item.CrowdType = "women"
				break
			}
		}
		for _, v := range []string{"man", "men", "male"} {
			if strings.Contains(strings.ToLower(item.Category), v) {
				item.CrowdType = "men"
				break
			}
		}
		for _, v := range []string{"kid", "child", "girl", "boy"} {
			if strings.Contains(strings.ToLower(item.Category), v) {
				item.CrowdType = "kids"
				break
			}
		}

		for kv, rawSku := range article[i].Variants {

			sku := pbItem.Sku{
				SourceId: strconv.Format(rawSku.VariantCode),
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(originalPrice * 100),
					Msrp:     int32(MSRP * 100),
					Discount: int32(discount),
				},
				Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}
			for j := range q.Availability {
				if q.Availability[j] == rawSku.VariantCode {
					sku.Stock.StockStatus = pbItem.Stock_InStock
					break
				}
			}

			imgIcon := ""
			sel = doc.Find(`.a-image.is-hidden.Resolve`)
			currarticlecode := articleCodeReg.FindSubmatch([]byte(article[i].URL))
			for x := range sel.Nodes {
				if articlecode, _ := sel.Eq(x).Attr("data-articlecode"); articlecode == string(currarticlecode[0]) {
					resolvechain, _ := sel.Eq(x).Attr("data-resolvechain")
					imgIcon = "https://lp.stories.com/app005prod?set=key[resolve.pixelRatio],value[1]&set=key[resolve.width],value[50]&set=key[resolve.height],value[10000]&set=key[resolve.imageFit],value[containerwidth]&set=key[resolve.allowImageUpscaling],value[0]&set=key[resolve.format],value[webp]&set=key[resolve.quality],value[90]&" + resolvechain
					break
				}
			}
			// color
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    article[i].ColorCode,
				Name:  article[i].ColorLoc,
				Value: article[i].ColorLoc,
				Icon:  imgIcon,
			})

			if kv == 0 {
				isDefault := true
				for m, mid := range article[i].VAssets {
					if m > 0 {
						isDefault = false
					}
					sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
						strconv.Format(m),
						fmt.Sprintf("https:%s", mid.Thumbnail),
						fmt.Sprintf("https:%s&set=key[resolve.width],value[900]", mid.Thumbnail),
						fmt.Sprintf("https:%s&set=key[resolve.width],value[600]", mid.Thumbnail),
						fmt.Sprintf("https:%s&set=key[resolve.width],value[500]", mid.Thumbnail),
						"",
						isDefault,
					))
				}
			}

			// size
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    rawSku.VariantCode,
				Name:  rawSku.SizeName,
				Value: rawSku.SizeName,
			})

			item.SkuItems = append(item.SkuItems, &sku)
		}

		// yield item result
		if err = yield(ctx, &item); err != nil {
			return err
		}
	}
	return nil
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://www.stories.com/en_usd/sale/all-sale.html",
		"https://www.stories.com/en_usd/clothing/blouses-shirts/shirts/product.oversized-wool-blend-workwear-shirt-brown.0764033007.html",
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
				i.URL.Host = "www.stories.com"
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
