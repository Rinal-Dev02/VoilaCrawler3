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

	"github.com/voiladev/go-framework/glog"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/voiladev/go-framework/strconv"
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
		categoryPathMatcher:  regexp.MustCompile(`^/us/en(/[a-z0-9-]+){1,6}$`),
		categoryPathMatcher1: regexp.MustCompile(`(.*)/Search-UpdateGrid`),
		// this regular used to match product page url path
		productPathMatcher:  regexp.MustCompile(`^/us/en(/[a-z0-9-]+){1,3}/[A-Za-z0-9-]+.html$`),
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
	return []string{"*.makeupforever.com"}
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

	if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) || c.categoryPathMatcher1.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	} else if c.productPathMatcher.MatchString(resp.Request.URL.Path) || c.productPathMatcher1.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yield)
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

	if !bytes.Contains(respBody, []byte("class=\"productLinkto")) {
		return errors.New("products not found")
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	lastIndex := nextIndex(ctx)
	sel := doc.Find(`.row.product-grid`).Find(`li`).Find(`.product`)
	if len(sel.Nodes) == 0 {
		sel = doc.Find(`li`).Find(`.product`)
	}

	c.logger.Debugf("nodes %d", len(sel.Nodes))
	for i := range sel.Nodes {
		node := sel.Eq(i)
		if href, _ := node.Attr("data-pid"); href != "" {

			rawurl := "https://www.makeupforever.com/on/demandware.store/Sites-MakeUpForEver-US-Site/en_US/Product-Variation?quantity=1&pid=" + href

			fmt.Println(rawurl)
			req, err := http.NewRequest(http.MethodGet, rawurl, nil)
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

	totalRecords := 0
	if bytes.Contains(respBody, []byte("class=\"numberproduct-bold\"")) {
		resultTR, _ := strconv.ParseInt(doc.Find(`.numberproduct-bold`).Text())
		totalRecords = int(resultTR)
	}

	if totalRecords > 0 && totalRecords <= lastIndex {
		// nextpage not found
		return nil
	}

	if !bytes.Contains(respBody, []byte("<div class=\"show-more\" id=\"show-more\"")) {
		// nextpage not found
		return nil
	}

	nexturl, _ := doc.Find(`.show-more`).Find(`button`).Attr(`data-url`)
	re := regexp.MustCompile(`&start=[a-z0-9&=]+`)

	nexturl = string(re.ReplaceAll([]byte(nexturl), []byte("&start="+strconv.Format(lastIndex+1)+"&sz=12")))

	req, _ := http.NewRequest(http.MethodGet, nexturl, nil)
	// update the index of last page
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
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
				Value        int    `json:"value"`
				Currency     string `json:"currency"`
				Formatted    string `json:"formatted"`
				DecimalPrice string `json:"decimalPrice"`
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
			AttributeID string `json:"attributeId"`
			DisplayName string `json:"displayName"`
			ID          string `json:"id"`
			Swatchable  bool   `json:"swatchable"`
			Values      []struct {
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
			} `json:"values"`
			Sections struct {
				All []struct {
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
				} `json:"all"`
			} `json:"sections"`
			FullViewSections struct {
				All []struct {
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
				} `json:"all"`
			} `json:"fullViewSections"`
			ResetURL string `json:"resetUrl,omitempty"`
		} `json:"variationAttributes"`
		LongDescription  string `json:"longDescription"`
		ShortDescription string `json:"shortDescription"`
		Ingredients      string `json:"ingredients"`
		MainIngredients  string `json:"mainIngredients"`
		HowToUse         struct {
		} `json:"howToUse"`
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
	} `json:"product"`
	Resources struct {
		ProductsLeftNbr    string `json:"productsLeftNbr"`
		InfoSelectforstock string `json:"info_selectforstock"`
		OutOfStock         string `json:"outOfStock"`
	} `json:"resources"`
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

	if !bytes.Contains(respBody, []byte(`"Product-Variation",`)) {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}

	var viewData parseProductData

	if err := json.Unmarshal(respBody, &viewData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}

	colorIndex := -1
	sizeIndex := -1

	for kv, rawVariation := range viewData.Product.VariationAttributes {
		if rawVariation.AttributeID == "size" {
			sizeIndex = kv
		}
		if rawVariation.AttributeID == "color" {
			colorIndex = kv
		}
	}

	counters := 0
	loopIndex := 0
	if sizeIndex > -1 {
		loopIndex = sizeIndex
	}

	rating, _ := strconv.ParseFloat(viewData.Product.Rating)
	for _, rawColor := range viewData.Product.VariationAttributes[colorIndex].Values {
		counters++
		// build product data
		item := pbItem.Product{
			Source: &pbItem.Source{
				Id: viewData.Product.MasterID,
				//CrawlUrl: resp.Request.URL.String(),  // not found
				//https://www.makeupforever.com/us/en/a-MI000044405.html?dwvar_MI000044405_color=368&pid=MI000044405&quantity=1
				CrawlUrl: "https://www.makeupforever.com/us/en/a-" + viewData.Product.MasterID + ".html?dwvar_" + viewData.Product.MasterID + "_color=" + rawColor.ID + "&pid=" + viewData.Product.MasterID + "&quantity=1",
			},
			BrandName:   viewData.Product.Brand,
			Title:       viewData.Product.ProductName,
			Description: htmlTrimRegp.ReplaceAllString(viewData.Product.LongDescription, ""),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
			},
			Stats: &pbItem.Stats{
				//ReviewCount: int32(viewData.Product.TurntoReviewCount),
				Rating: float32(rating),
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
			}
		}

		originalPrice, _ := strconv.ParseFloat(viewData.Product.Price.Sales.Value)
		msrp, _ := strconv.ParseFloat(viewData.Product.Price.List.Value)
		discount := 0.0
		if msrp > 0 {
			discount = ((originalPrice - msrp) / msrp) * 100
		}

		for kv, rawSku := range viewData.Product.VariationAttributes[loopIndex].Values {

			if sizeIndex == -1 { // size not available
				if kv > 0 {
					break
				}
			}

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
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    strconv.Format(rawColor.ID),
				Name:  rawColor.COREShadeNumber + " - " + rawColor.DisplayValue,
				Value: rawColor.COREShadeNumber + " - " + rawColor.DisplayValue,
				//Icon:  color.SwatchMedia.Mobile,
			})

			if kv == 0 {

				isDefault := true
				for ki, mid := range rawColor.Images.Swatch {
					template := mid.URL
					if ki > 0 {
						isDefault = false
					}
					s := strings.Split(template, "?sw=")

					sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
						strconv.Format(ki),
						s[0],
						s[0]+"?sw=1000&sh=1200",
						s[0]+"?sw=600&sh=800",
						s[0]+"?sw=500&sh=600",
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
	}

	return nil
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		"https://www.makeupforever.com/us/en/face/foundation",
		//"https://www.makeupforever.com/on/demandware.store/Sites-MakeUpForEver-US-Site/en_US/Product-Variation?pid=MI000044405&quantity=0",
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
	opts := spider.CrawlOptions(nil)

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
			//i.Header.Add("x-requested-with", "XMLHttpRequest")
			i.Header.Set("Accept", "%")
			i.Header.Add("referer", "https://www.makeupforever.com/us/en/eyes/eyeshadow/star-lit-diamond-powder-MI000090111.html")

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
			resp, err := client.DoWithOptions(nctx, i, http.Options{
				EnableProxy:       true,
				EnableHeadless:    true,
				EnableSessionInit: spider.CrawlOptions(nil).EnableSessionInit,
				KeepSession:       spider.CrawlOptions(nil).KeepSession,
				Reliability:       spider.CrawlOptions(nil).Reliability,
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

	ctx := context.WithValue(context.Background(), "tracing_id", fmt.Sprintf("tracing_%d", time.Now().UnixNano()))
	// start the crawl request
	for _, req := range spider.NewTestRequest(context.Background()) {
		if err := callback(ctx, req); err != nil {
			logger.Fatal(err)
		}
	}
}
