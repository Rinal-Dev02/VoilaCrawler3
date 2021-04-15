package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
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
	"github.com/voiladev/go-framework/randutil"
	"github.com/voiladev/go-framework/strconv"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// _Crawler defined the crawler struct/class for which is not necessory to be exportable
type _Crawler struct {
	// httpClient is the object of an http client
	httpClient           http.Client
	categoryPathMatcher  *regexp.Regexp
	categoryPathMatcher1 *regexp.Regexp
	productPathMatcher   *regexp.Regexp
	productPathMatcher1  *regexp.Regexp
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
		categoryPathMatcher:  regexp.MustCompile(`^/us/en(/[a-z0-9\-]+){1,6}$`),
		categoryPathMatcher1: regexp.MustCompile(`(.*)/Search-UpdateGrid`),
		// this regular used to match product page url path
		productPathMatcher:  regexp.MustCompile(`^/us/en(/[a-z0-9\-]+){1,3}/[^/]+\-[A-Z0-9]+\.html$`),
		productPathMatcher1: regexp.MustCompile(`(.*)/Product-Variation`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "baf02094073b457d9434f929b119ee30"
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
		EnableHeadless:    false,
		EnableSessionInit: true,
		Reliability:       pbProxy.ProxyReliability_ReliabilityDefault,
	}
	opts.MustCookies = append(opts.MustCookies,
		&http.Cookie{Name: "country_data", Value: "US~en", Path: "/"},
		&http.Cookie{Name: "backoptinpopin2", Value: "0", Path: "/"},
	)
	return opts
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.makeupforever.com"}
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
		u.Host = "www.makeupforever.com"
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

	if c.productPathMatcher.MatchString(resp.Request.URL.Path) || c.productPathMatcher1.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yield)
	} else if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) || c.categoryPathMatcher1.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	}
	return fmt.Errorf("unsupported url %s", resp.Request.URL.String())
}

// nextIndex used to get the index from the shared data.
// item.index is a const key for item index.
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")))
}

var nextPageReg = regexp.MustCompile(`\[<a\s+href="(https://www.makeupforever.com/[^"]+)"\s*>\s*Next\s*</a>\]`)

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

	if !bytes.Contains(respBody, []byte("class=\"productLinkto")) {
		return errors.New("products not found")
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	lastIndex := nextIndex(ctx)
	sel := doc.Find(`ul[itemid="#product"]>li`)
	c.logger.Debugf("nodes %d", len(sel.Nodes))
	for i := range sel.Nodes {
		node := sel.Eq(i)

		if href, _ := node.Find(`.product .product-tile .productLinkto`).Attr("href"); href != "" {
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

	matched := nextPageReg.FindStringSubmatch(string(respBody))
	if len(matched) < 2 {
		return nil
	}
	nexturl := matched[1]
	req, _ := http.NewRequest(http.MethodGet, nexturl, nil)
	// update the index of last page
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
}

type AttributeValue struct {
	ID              string      `json:"id"`
	Description     interface{} `json:"description"`
	DisplayValue    string      `json:"displayValue"`
	Value           string      `json:"value"`
	Selected        bool        `json:"selected"`
	Selectable      bool        `json:"selectable"`
	IsFormatVoyage  bool        `json:"isFormatVoyage"`
	FormatVoyageMsg interface{} `json:"formatVoyageMsg"`
	IsBigFormat     bool        `json:"isBigFormat"`
	BigFormatMsg    interface{} `json:"bigFormatMsg"`
	URL             string      `json:"url"`
	Images          struct {
		Swatch []struct {
			Alt   string `json:"alt"`
			URL   string `json:"url"`
			Title string `json:"title"`
		} `json:"swatch"`
	} `json:"images"`
	IsMonoColor             bool          `json:"isMonoColor"`
	CORESingleShadeHexaCode string        `json:"CORE_single_shade_hexa_code"`
	COREShadeNumber         string        `json:"CORE_shade_number"`
	MufeShadeType           []interface{} `json:"mufeShadeType"`
	DefaultShadePosition    int           `json:"defaultShadePosition"`
	FullShadePosition       int           `json:"fullShadePosition"`
}

type parseProductData struct {
	Action      string `json:"action"`
	QueryString string `json:"queryString"`
	Locale      string `json:"locale"`
	Product     struct {
		UUID                  string      `json:"uuid"`
		ID                    string      `json:"id"`
		ProductName           string      `json:"productName"`
		ProductSubName        string      `json:"productSubName"`
		ProductType           string      `json:"productType"`
		Brand                 string      `json:"brand"`
		CoreProductType       interface{} `json:"coreProductType"`
		COREShadeNumber       interface{} `json:"CORE_shade_number"`
		MufePromotionalBanner interface{} `json:"mufePromotionalBanner"`
		Price                 struct {
			Sales struct {
				Value        float64 `json:"value"`
				Currency     string  `json:"currency"`
				Formatted    string  `json:"formatted"`
				DecimalPrice string  `json:"decimalPrice"`
			} `json:"sales"`
			List struct {
				Value        float64 `json:"value"`
				Currency     string  `json:"currency"`
				Formatted    string  `json:"formatted"`
				DecimalPrice string  `json:"decimalPrice"`
			} `json:"list"`
			HTML string `json:"html"`
		} `json:"price"`
		Images struct {
			Large []struct {
				Alt   string `json:"alt"`
				URL   string `json:"url"`
				Title string `json:"title"`
			} `json:"large"`
			Small []struct {
				Alt   string `json:"alt"`
				URL   string `json:"url"`
				Title string `json:"title"`
			} `json:"small"`
			Tracemedium []interface{} `json:"tracemedium"`
		} `json:"images"`
		ImageStamp          interface{} `json:"imageStamp"`
		ImageStampVariant   interface{} `json:"imageStampVariant"`
		MinOrderQuantity    int         `json:"minOrderQuantity"`
		MaxOrderQuantity    int         `json:"maxOrderQuantity"`
		SelectedQuantity    int         `json:"selectedQuantity"`
		VariationAttributes []struct {
			AttributeID string            `json:"attributeId"`
			DisplayName string            `json:"displayName"`
			ID          string            `json:"id"`
			Swatchable  bool              `json:"swatchable"`
			Values      []*AttributeValue `json:"values"`
			Sections    struct {
				All []*AttributeValue `json:"all"`
			} `json:"sections"`
			FullViewSections struct {
				All []*AttributeValue `json:"all"`
			} `json:"fullViewSections"`
			ResetURL string `json:"resetUrl,omitempty"`
		} `json:"variationAttributes"`
		LongDescription  string `json:"longDescription"`
		ShortDescription string `json:"shortDescription"`
		Ingredients      string `json:"ingredients"`
		MainIngredients  string `json:"mainIngredients"`
		// HowToUse         struct {
		// } `json:"howToUse"`
		SelectedProductURL string  `json:"selectedProductUrl"`
		Rating             float64 `json:"rating"`
		Promotions         []struct {
			CalloutMsg     string `json:"calloutMsg"`
			Details        string `json:"details"`
			Enabled        bool   `json:"enabled"`
			ID             string `json:"id"`
			Name           string `json:"name"`
			PromotionClass string `json:"promotionClass"`
			Rank           int    `json:"rank"`
		} `json:"promotions"`
		Available               bool          `json:"available"`
		InStock                 bool          `json:"inStock"`
		Template                interface{}   `json:"template"`
		Badge                   string        `json:"badge"`
		Retailers               []interface{} `json:"Retailers"`
		Recommendations         []interface{} `json:"recommendations"`
		MasterID                string        `json:"masterID"`
		DefaultVariant          string        `json:"defaultVariant"`
		FirstAvailableVariant   string        `json:"firstAvailableVariant"`
		DefaultVariantAvailable bool          `json:"defaultVariantAvailable"`
		VariantList             []struct {
			Index             int     `json:"index"`
			ProdID            string  `json:"prodID"`
			Available         float64 `json:"available"`
			Online            bool    `json:"online"`
			CustomName        string  `json:"customName"`
			CustomDescription string  `json:"customDescription"`
			CustomShadeColor  string  `json:"customShadeColor"`
		} `json:"variantList"`
	} `json:"product"`
	Resources struct {
		ProductsLeftNbr    string `json:"productsLeftNbr"`
		InfoSelectforstock string `json:"info_selectforstock"`
		OutOfStock         string `json:"outOfStock"`
	} `json:"resources"`
}

// used to trim html labels in description
var htmlTrimRegp = regexp.MustCompile(`</?[^>]+>`)
var productsExtractReg = regexp.MustCompile(`([A-Z0-9])+.html`)

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

	re := regexp.MustCompile(`[^\d]`)
	review, _ := strconv.ParseInt(re.ReplaceAllString(doc.Find(`.bv_numReviews_text`).Text(), ""))

	if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		matched := productsExtractReg.FindSubmatch([]byte(resp.Request.URL.Path))
		produrl := "https://www.makeupforever.com/on/demandware.store/Sites-MakeUpForEver-US-Site/en_US/Product-Variation?quantity=1&pid=" + strings.ReplaceAll(string(matched[0]), ".html", "")

		req, err := http.NewRequest(http.MethodGet, produrl, nil)
		req.Header.Set("Accept", "%")
		req.Header.Set("Referer", fmt.Sprintf("%s://%s/%s", resp.Request.URL.Scheme, resp.Request.URL.Host, resp.Request.URL.Path))

		respNew, err := c.httpClient.Do(ctx, req)
		if err != nil {
			panic(err)
		}

		respBody, err = ioutil.ReadAll(respNew.Body)
		if err != nil {
			return err
		}
	}

	if !bytes.Contains(respBody, []byte(`"Product-Variation",`)) {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}

	var viewData parseProductData
	if err := json.Unmarshal(respBody, &viewData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}

	canUrl, _ := c.CanonicalUrl(resp.Request.URL.String())
	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id: viewData.Product.MasterID,
			//CrawlUrl: resp.Request.URL.String(),  // not found
			//https://www.makeupforever.com/us/en/a-MI000044405.html?dwvar_MI000044405_color=368&pid=MI000044405&quantity=1
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
		},
		BrandName:   viewData.Product.Brand,
		Title:       viewData.Product.ProductName,
		Description: htmlTrimRegp.ReplaceAllString(viewData.Product.LongDescription, ""),
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stats: &pbItem.Stats{
			ReviewCount: int32(review),
			Rating:      float32(viewData.Product.Rating),
		},
	}

	list := strings.Split(viewData.Product.SelectedProductURL, "/")
	for j, l := range list {
		if j < 3 || j >= len(list)-1 {
			continue
		}
		switch j {
		case 3:
			item.Category = l
		case 4:
			item.SubCategory = l
		case 5:
			item.SubCategory2 = l
		case 6:
			item.SubCategory3 = l
		case 7:
			item.SubCategory4 = l
		}
	}

	var (
		colorMap = map[string]*AttributeValue{}
		sizes    []*AttributeValue
	)
	for _, attr := range viewData.Product.VariationAttributes {
		switch strings.ToLower(attr.AttributeID) {
		case "color":
			for _, val := range attr.Values {
				colorMap[val.COREShadeNumber] = val
				colorMap[val.CORESingleShadeHexaCode] = val
			}
		case "size":
			sizes = append(sizes, attr.Values...)
		default:
			c.logger.Errorf("unsupported attribute %s", attr.AttributeID)
		}
	}

	originalPrice := viewData.Product.Price.Sales.Value
	msrp := viewData.Product.Price.List.Value
	discount := 0.0
	if msrp > 0 {
		discount = math.Ceil((msrp - originalPrice) / msrp * 100)
	}
	var (
		medias    []*pbMedia.Media
		colorSpec *pbItem.SkuSpecOption
		color     *AttributeValue
	)
	for _, rawSku := range viewData.Product.VariantList {
		if rawSku.CustomName != "" {
			color = colorMap[rawSku.CustomName]
			if color == nil {
				continue
			}

			// color
			colorSpec = &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    strconv.Format(color.ID),
				Name:  color.COREShadeNumber + " - " + color.DisplayValue,
				Value: color.CORESingleShadeHexaCode,
			}

			medias = medias[0:0]
			for ki, mid := range color.Images.Swatch {
				s := strings.Split(mid.URL, "?")
				medias = append(medias, pbMedia.NewImageMedia(
					strconv.Format(ki),
					s[0],
					s[0]+"?sw=1000&sh=1200",
					s[0]+"?sw=600&sh=800",
					s[0]+"?sw=500&sh=600",
					"",
					ki == 0,
				))
			}
		} else {
			medias = medias[0:0]
			for ki, mid := range viewData.Product.Images.Large {
				s := strings.Split(mid.URL, "?")
				medias = append(medias, pbMedia.NewImageMedia(
					strconv.Format(ki),
					s[0],
					s[0]+"?sw=1000&sh=1200",
					s[0]+"?sw=600&sh=800",
					s[0]+"?sw=500&sh=600",
					"",
					ki == 0,
				))
			}
		}

		for _, size := range sizes {
			skuId := rawSku.ProdID
			if len(sizes) > 1 {
				skuId = fmt.Sprintf("%s-%s", rawSku.ProdID, size.ID)
			}
			sku := pbItem.Sku{
				SourceId: skuId,
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(originalPrice * 100),
					Msrp:     int32(msrp * 100),
					Discount: int32(discount),
				},
				Medias: medias,
				Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}
			if (color == nil || color.Selectable) && size.Selectable {
				sku.Stock.StockStatus = pbItem.Stock_InStock
			}

			if colorSpec != nil {
				sku.Specs = append(sku.Specs, colorSpec)
			}
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    size.ID,
				Name:  size.DisplayValue,
				Value: size.Value,
			})
			item.SkuItems = append(item.SkuItems, &sku)
		}
	}

	if err = yield(ctx, &item); err != nil {
		return err
	}
	return nil
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		"https://www.makeupforever.com/us/en/tools",
		// "https://www.makeupforever.com/us/en/face/bronzer/pro-sculpting-palette-MI000014320.html",
		// "https://www.makeupforever.com/us/en/eyes/eyeshadow/artist-color-shadow-refill-MI000079830.html",
		// "https://www.makeupforever.com/us/en/face/foundation/make-up-for-ever-%E2%80%93-reboot-MI000028230.html",
		// "https://www.makeupforever.com/us/en/face/foundation",
		// "https://www.makeupforever.com/on/demandware.store/Sites-MakeUpForEver-US-Site/en_US/Product-Variation?pid=MI000044405&quantity=0",
		// "https://www.makeupforever.com/us/en/face/foundation/ultra-hd-foundation-palette-MI000041000.html",
		// "https://www.makeupforever.com/us/en/tools/sponge/matte-velvet-skin-sponge-MI000015023.html",
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
	var disableParseDetail bool
	flag.BoolVar(&disableParseDetail, "disable-detail", false, "disable parse detail")
	flag.Parse()

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
			if disableParseDetail {
				crawler := spider.(*_Crawler)
				if crawler.productPathMatcher.MatchString(i.URL.Path) {
					return nil
				}
			}
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
				i.URL.Host = "www.makeupforever.com"
			}

			// do http requests here.
			nctx, cancel := context.WithTimeout(ctx, time.Minute*5)
			defer cancel()
			nctx = context.WithValue(nctx, "req_id", randutil.MustNewRandomID())
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

			return spider.Parse(nctx, resp, callback)
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

	ctx := context.WithValue(context.Background(), "tracing_id", fmt.Sprintf("tracing_%d", time.Now().UnixNano()))
	// start the crawl request
	for _, req := range spider.NewTestRequest(context.Background()) {
		if err := callback(ctx, req); err != nil {
			logger.Fatal(err)
		}
	}
}
