package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"regexp"
	"strings"
	"time"

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
		categoryPathMatcher: regexp.MustCompile(`^(/[a-z0-9-]+)+(-l)([a-z0-9-]+).html(.*)$`),
		// this regular used to match product page url path
		productPathMatcher: regexp.MustCompile(`^(/[a-z0-9-]+)+(-p)([a-z0-9-]+).html(.*)$`),
		logger:             logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "18fda7c867f64308a22ebdf81bc17aba"
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
	return []string{"*.zara.com"}
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

// below are the golang json data struct of raw website.
// if you get the raw json data of the website,
// then you can use https://mholt.github.io/json-to-go/ to convert it to a golang struct

type CategoryStructure struct {
	ProductGroups []struct {
		Type     string `json:"type"`
		Elements []struct {
			ID                   string `json:"id"`
			Layout               string `json:"layout,omitempty"`
			CommercialComponents []struct {
				ID        int    `json:"id"`
				Reference string `json:"reference"`
				Type      string `json:"type"`
				Kind      string `json:"kind"`
				Brand     struct {
					BrandID        int    `json:"brandId"`
					BrandGroupID   int    `json:"brandGroupId"`
					BrandGroupCode string `json:"brandGroupCode"`
				} `json:"brand"`
				Xmedia []struct {
					Datatype       string   `json:"datatype"`
					Set            int      `json:"set"`
					Type           string   `json:"type"`
					Kind           string   `json:"kind"`
					Path           string   `json:"path"`
					Name           string   `json:"name"`
					Width          int      `json:"width"`
					Height         int      `json:"height"`
					Timestamp      string   `json:"timestamp"`
					AllowedScreens []string `json:"allowedScreens"`
					ExtraInfo      struct {
						Style struct {
							Top      int  `json:"top"`
							Left     int  `json:"left"`
							Width    int  `json:"width"`
							Margined bool `json:"margined"`
						} `json:"style"`
					} `json:"extraInfo"`
				} `json:"xmedia"`
				Name          string  `json:"name"`
				Description   string  `json:"description"`
				Price         float64 `json:"price"`
				Section       int     `json:"section"`
				SectionName   string  `json:"sectionName"`
				FamilyName    string  `json:"familyName"`
				SubfamilyName string  `json:"subfamilyName"`
				Seo           struct {
					Keyword          string `json:"keyword"`
					SeoProductID     string `json:"seoProductId"`
					DiscernProductID int    `json:"discernProductId"`
				} `json:"seo"`
				Availability string        `json:"availability"`
				TagTypes     []interface{} `json:"tagTypes"`
				ExtraInfo    struct {
					IsDivider      bool `json:"isDivider"`
					HighlightPrice bool `json:"highlightPrice"`
				} `json:"extraInfo"`
				Detail struct {
					Reference        string `json:"reference"`
					DisplayReference string `json:"displayReference"`
					Colors           []struct {
						ID        string `json:"id"`
						ProductID int    `json:"productId"`
						Name      string `json:"name"`
						StylingID string `json:"stylingId"`
						Xmedia    []struct {
							Datatype       string   `json:"datatype"`
							Set            int      `json:"set"`
							Type           string   `json:"type"`
							Kind           string   `json:"kind"`
							Path           string   `json:"path"`
							Name           string   `json:"name"`
							Width          int      `json:"width"`
							Height         int      `json:"height"`
							Timestamp      string   `json:"timestamp"`
							AllowedScreens []string `json:"allowedScreens"`
							ExtraInfo      struct {
								Style struct {
									Top      int  `json:"top"`
									Left     int  `json:"left"`
									Width    int  `json:"width"`
									Margined bool `json:"margined"`
								} `json:"style"`
							} `json:"extraInfo"`
						} `json:"xmedia"`
						Price        float64 `json:"price"`
						Availability string  `json:"availability"`
						Reference    string  `json:"reference"`
					} `json:"colors"`
				} `json:"detail"`
				ServerPage               int      `json:"serverPage"`
				GridPosition             int      `json:"gridPosition"`
				ZoomedGridPosition       int      `json:"zoomedGridPosition"`
				HasMoreColors            bool     `json:"hasMoreColors"`
				ProductTag               []string `json:"productTag"`
				ProductTagDynamicClasses string   `json:"productTagDynamicClasses"`
				ColorList                string   `json:"colorList"`
				IsDivider                bool     `json:"isDivider"`
				HasXmediaDouble          bool     `json:"hasXmediaDouble"`
				SimpleXmedia             []struct {
					Datatype       string   `json:"datatype"`
					Set            int      `json:"set"`
					Type           string   `json:"type"`
					Kind           string   `json:"kind"`
					Path           string   `json:"path"`
					Name           string   `json:"name"`
					Width          int      `json:"width"`
					Height         int      `json:"height"`
					Timestamp      string   `json:"timestamp"`
					AllowedScreens []string `json:"allowedScreens"`
					ExtraInfo      struct {
						Style struct {
							Top      int  `json:"top"`
							Left     int  `json:"left"`
							Width    int  `json:"width"`
							Margined bool `json:"margined"`
						} `json:"style"`
					} `json:"extraInfo"`
				} `json:"simpleXmedia"`
				ShowAvailability bool `json:"showAvailability"`
				PriceUnavailable bool `json:"priceUnavailable"`
			} `json:"commercialComponents,omitempty"`
			HasStickyBanner bool   `json:"hasStickyBanner,omitempty"`
			NeedsSeparator  bool   `json:"needsSeparator,omitempty"`
			Header          string `json:"header,omitempty"`
			Description     string `json:"description,omitempty"`
		} `json:"elements"`
		HasStickyBanner bool `json:"hasStickyBanner"`
	} `json:"productGroups"`
	ProductsCount    int `json:"productsCount"`
	ProductsPage     int `json:"productsPage"`
	ProductsPageSize int `json:"productsPageSize"`
}

// used to extract embaded json data in website page.
// more about golang regulation see here https://golang.org/pkg/regexp/syntax/
var productsExtractReg = regexp.MustCompile(`(?U)window\.zara\.viewPayload\s*=\s*({.*});</script>`)

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

	var viewData CategoryStructure
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		return err
	}

	lastIndex := nextIndex(ctx)
	for _, pg := range viewData.ProductGroups {
		if pg.Type != "main" {
			continue
		}

		for _, idv := range pg.Elements {
			if idv.ID == "seo-info" {
				continue
			}
			for _, pc := range idv.CommercialComponents {
				if pc.Type != "Product" {
					continue
				}

				rawurl := fmt.Sprintf("%s://%s/%s-p%s.html?v1=%v", resp.Request.URL.Scheme, resp.Request.URL.Host, pc.Seo.Keyword, pc.Seo.SeoProductID, pc.Seo.DiscernProductID)
				fmt.Println(rawurl)

				req, err := http.NewRequest(http.MethodGet, rawurl, nil)
				if err != nil {
					c.logger.Errorf("load http request of url %s failed, error=%s", rawurl, err)
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
		}
	}

	pageSize := float64(viewData.ProductsPageSize)
	productsCount := float64(viewData.ProductsCount)
	totalPages := int64(math.Ceil(productsCount / pageSize))
	// get current page number
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}

	// check if this is the last page
	if page >= totalPages {
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

// Product json
type ProductStructure struct {
	Product struct {
		ID    int `json:"id"`
		Brand struct {
			BrandID        int    `json:"brandId"`
			BrandGroupID   int    `json:"brandGroupId"`
			BrandGroupCode string `json:"brandGroupCode"`
		} `json:"brand"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Detail      struct {
			Description      string        `json:"description"`
			RawDescription   string        `json:"rawDescription"`
			Reference        string        `json:"reference"`
			DisplayReference string        `json:"displayReference"`
			Composition      []interface{} `json:"composition"`
			Care             []interface{} `json:"care"`
			Colors           []struct {
				ID               string        `json:"id"`
				HexCode          string        `json:"hexCode"`
				ProductID        int           `json:"productId"`
				Name             string        `json:"name"`
				Reference        string        `json:"reference"`
				StylingID        string        `json:"stylingId"`
				DetailImages     []interface{} `json:"detailImages"`
				DetailFlatImages []interface{} `json:"detailFlatImages"`
				SizeGuideImages  []interface{} `json:"sizeGuideImages"`
				Videos           []interface{} `json:"videos"`
				Xmedia           []struct {
					Datatype       string   `json:"datatype"`
					Set            int      `json:"set"`
					Type           string   `json:"type"`
					Kind           string   `json:"kind"`
					Path           string   `json:"path"`
					Name           string   `json:"name"`
					Width          int      `json:"width"`
					Height         int      `json:"height"`
					Timestamp      string   `json:"timestamp"`
					AllowedScreens []string `json:"allowedScreens"`
					Gravity        string   `json:"gravity"`
					Order          int      `json:"order,omitempty"`
				} `json:"xmedia"`
				Price    float64 `json:"price"`
				OldPrice float64 `json:"oldPrice"`
				Sizes    []struct {
					Availability     string  `json:"availability"`
					EquivalentSizeID int     `json:"equivalentSizeId"`
					ID               int     `json:"id"`
					Name             string  `json:"name"`
					Price            float64 `json:"price"`
					OldPrice         float64 `json:"oldPrice"`
					Reference        string  `json:"reference"`
					Sku              int     `json:"sku"`
				} `json:"sizes"`
				Description    string `json:"description"`
				RawDescription string `json:"rawDescription"`
				ExtraInfo      struct {
					Preorder struct {
						Message    string `json:"message"`
						IsPreorder bool   `json:"isPreorder"`
					} `json:"preorder"`
					IsStockInStoresAvailable bool `json:"isStockInStoresAvailable"`
				} `json:"extraInfo"`
				DetailedComposition struct {
					Parts      []interface{} `json:"parts"`
					Exceptions []interface{} `json:"exceptions"`
				} `json:"detailedComposition"`
				ColorCutImg struct {
					Datatype       string   `json:"datatype"`
					Set            int      `json:"set"`
					Type           string   `json:"type"`
					Kind           string   `json:"kind"`
					Path           string   `json:"path"`
					Name           string   `json:"name"`
					Width          int      `json:"width"`
					Height         int      `json:"height"`
					Timestamp      string   `json:"timestamp"`
					AllowedScreens []string `json:"allowedScreens"`
					Gravity        string   `json:"gravity"`
				} `json:"colorCutImg"`
				MainImgs []struct {
					Datatype       string   `json:"datatype"`
					Set            int      `json:"set"`
					Type           string   `json:"type"`
					Kind           string   `json:"kind"`
					Path           string   `json:"path"`
					Name           string   `json:"name"`
					Width          int      `json:"width"`
					Height         int      `json:"height"`
					Timestamp      string   `json:"timestamp"`
					AllowedScreens []string `json:"allowedScreens"`
					Gravity        string   `json:"gravity"`
					Order          int      `json:"order"`
				} `json:"mainImgs"`
			} `json:"colors"`
			DetailedComposition struct {
				Parts      []interface{} `json:"parts"`
				Exceptions []interface{} `json:"exceptions"`
			} `json:"detailedComposition"`
			Categories []interface{} `json:"categories"`
			IsBuyable  bool          `json:"isBuyable"`
		} `json:"detail"`
	} `json:"product"`
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

	var viewData ProductStructure

	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}

	colorcode := resp.Request.URL.Query().Get("v1")

	//Prepare Product Data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:       strconv.Format(viewData.Product.ID),
			CrawlUrl: resp.Request.URL.String(),
		},
		BrandName:   viewData.Product.Brand.BrandGroupCode,
		Title:       viewData.Product.Name,
		Description: viewData.Product.Description,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		// Stats: &pbItem.Stats{
		// 	//ReviewCount: int32(p.NumberOfReviews),
		// 	//Rating:      float32(p.ReviewAverageRating / 5.0),
		// },
	}

	for _, rawColor := range viewData.Product.Detail.Colors {

		if colorcode != "" {
			if strconv.Format(rawColor.ProductID) != colorcode {
				continue
			}
		}

		for ks, rawSku := range rawColor.Sizes {

			originalPrice, _ := strconv.ParseFloat(rawSku.Price)
			msrp, _ := strconv.ParseFloat(rawSku.OldPrice)
			discount := 0.0
			if msrp > 0 {
				discount = (msrp - originalPrice) / msrp * 100
			}
			sku := pbItem.Sku{
				SourceId: strconv.Format(rawSku.ID),
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(originalPrice),
					Msrp:     int32(msrp),
					Discount: int32(discount),
				},
				Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}
			if viewData.Product.Detail.IsBuyable {
				sku.Stock.StockStatus = pbItem.Stock_InStock
				//sku.Stock.StockCount = int32(rawSku.TotalQuantityAvailable)
			}

			// color

			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    strconv.Format(rawColor.ProductID),
				Name:  rawColor.Name,
				Value: rawColor.Name,
				//Icon:  color.SwatchMedia.Mobile,
			})

			if ks == 0 {
				isDefault := true
				for _, mid := range rawColor.MainImgs {
					template := "https://static.zara.net/photos" + mid.Path + mid.Name + ".jpg?ts=" + mid.Timestamp
					if mid.Order > 1 {
						isDefault = false
					}
					if mid.Type == "image" {
						sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
							strconv.Format(mid.Name),
							template,
							template+"&sw=1000&sh=1200",
							template+"&sw=600&sh=800",
							template+"&sw=500&sh=600",
							"",
							isDefault,
						))
					}
				}
			}

			// size
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    strconv.Format(rawSku.Sku),
				Name:  rawSku.Name,
				Value: rawSku.Name,
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
		"https://www.zara.com/us/en/woman-bags-l1024.html?v1=1719123",
		//"https://www.zara.com/us/en/fabric-bucket-bag-with-pocket-p16619710.html?v1=100626185&v2=1719123",
		"https://www.zara.com/us/en/quilted-velvet-maxi-crossbody-bag-p16311710.html?v1=95124768&v2=1719102",
		//"https://www.zara.com/us/en/text-detail-belt-bag-p16363710.html?v1=79728740&v2=1719123",
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
				i.URL.Host = "www.zara.com"
			}

			// do http requests here.
			nctx, cancel := context.WithTimeout(ctx, time.Minute*5)
			defer cancel()
			resp, err := client.DoWithOptions(nctx, i, http.Options{
				EnableProxy:       true,
				EnableHeadless:    false,
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
