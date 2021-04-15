package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
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
		categoryPathMatcher: regexp.MustCompile(`^((\?!product).)*`),
		// this regular used to match product page url path
		productPathMatcher: regexp.MustCompile(`(.*)(product)(.*)`),
		logger:             logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "9d1a40adb8994961b0ff5d9e777119c4"
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
	opts.MustCookies = append(opts.MustCookies,
		&http.Cookie{Name: "country", Value: "US", Path: "/"},
		&http.Cookie{Name: "preferredLanguage", Value: "en", Path: "/"},
		&http.Cookie{Name: "lang", Value: "en_US", Path: "/"},
		&http.Cookie{Name: "gdprCountry", Value: "false", Path: "/"},
	)
	return opts
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.ssense.com"}
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
		u.Host = "www.ssense.com"
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

// below are the golang json data struct of raw website.
// if you get the raw json data of the website,
// then you can use https://mholt.github.io/json-to-go/ to convert it to a golang struct

// parseCategoryProducts parse api url from web page url
func (c *_Crawler) parseCategoryProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	// read the response data from http response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error(err)
		return err
	}

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	lastIndex := nextIndex(ctx)
	sel := dom.Find(`.plp-products__row .plp-products__column`)
	for i := range sel.Nodes {
		node := sel.Eq(i)
		href := node.Find(`a`).AttrOr("href", "")
		if href == "" {
			html, _ := node.Html()
			c.logger.Debugf("%s", html)
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

	pageSel := dom.Find(`.pagination__wrapper .nav>.pagination__item`)
	if len(pageSel.Nodes) == 0 {
		c.logger.Debugf("no nodes found")
		return nil
	}
	pageNode := pageSel.Eq(len(pageSel.Nodes) - 1)
	if strings.ToLower(strings.TrimSpace(pageNode.Find("a>span").Text())) != "next" {
		return nil
	}
	nextUrl := pageNode.Find(`a`).AttrOr("href", "")
	if nextUrl == "" {
		c.logger.Debug("no href found")
		return nil
	}
	req, _ := http.NewRequest(http.MethodGet, nextUrl, nil)
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
}

type ProductPageData struct {
	Products struct {
		Current struct {
			ID              int      `json:"id"`
			Name            string   `json:"name"`
			Images          []string `json:"images"`
			Gender          string   `json:"gender"`
			Sku             string   `json:"sku"`
			Composition     string   `json:"composition"`
			Description     string   `json:"description"`
			CreationDate    string   `json:"creationDate"`
			CountryOfOrigin string   `json:"countryOfOrigin"`
			InStock         bool     `json:"inStock"`
			Brand           struct {
				ID         int    `json:"id"`
				Name       string `json:"name"`
				SeoKeyword struct {
					Ko string `json:"ko"`
					Ja string `json:"ja"`
					En string `json:"en"`
					Fr string `json:"fr"`
					Zh string `json:"zh"`
				} `json:"seoKeyword"`
			} `json:"brand"`
			Category struct {
				SeoKeyword struct {
					Ko string `json:"ko"`
					Ja string `json:"ja"`
					En string `json:"en"`
					Fr string `json:"fr"`
					Zh string `json:"zh"`
				} `json:"seoKeyword"`
				ParentID       int    `json:"parentId"`
				ID             int    `json:"id"`
				Name           string `json:"name"`
				AllCategoryIds string `json:"allCategoryIds"`
			} `json:"category"`
			Price struct {
				Regular            int    `json:"regular"`
				Sale               int    `json:"sale"`
				Currency           string `json:"currency"`
				FormattedPrice     string `json:"formattedPrice"`
				FormattedSale      string `json:"formattedSale"`
				FullFormattedPrice string `json:"fullFormattedPrice"`
				FullFormattedSale  string `json:"fullFormattedSale"`
				Discount           int    `json:"discount"`
				FullFormat         string `json:"fullFormat"`
			} `json:"price"`
			ShowFinalSaleMessage bool `json:"showFinalSaleMessage"`
			Promotions           []struct {
				Translation string `json:"translation"`
				Limit       struct {
					Outbound int `json:"outbound"`
					Returns  int `json:"returns"`
				} `json:"limit"`
			} `json:"promotions"`
			Sizes []struct {
				ID               int         `json:"id"`
				InStock          bool        `json:"inStock"`
				Number           string      `json:"number"`
				Sku              string      `json:"sku"`
				Name             string      `json:"name"`
				Sequence         int         `json:"sequence"`
				NameSystemCode   interface{} `json:"nameSystemCode"`
				NumberSystemCode interface{} `json:"numberSystemCode"`
				LowStock         int         `json:"lowStock,omitempty"`
			} `json:"sizes"`
			IsUniSize bool `json:"isUniSize"`
			Display   bool `json:"display"`
		} `json:"current"`
	} `json:"products"`
}

// used to extract embaded json data in website page.
// more about golang regulation see here https://golang.org/pkg/regexp/syntax/
var productsExtractReg = regexp.MustCompile(`(?U)window\.INITIAL_STATE\s*=\s*({.*})\s*</script>`)

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

	var viewData ProductPageData
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}

	canUrl, _ := c.CanonicalUrl(resp.Request.URL.String())
	prodid := viewData.Products.Current.ID
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           strconv.Format(prodid),
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
		},
		BrandName:   viewData.Products.Current.Brand.Name,
		CrowdType:   viewData.Products.Current.Gender,
		Title:       viewData.Products.Current.Name,
		Description: viewData.Products.Current.Description,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
	}

	var (
		cates map[int]string
	)
	if item.CrowdType != "" {
		cates, err = func() (map[int]string, error) {
			u := fmt.Sprintf("https://www.ssense.com/en-us/data/mobilerefine.json?gender=%s", item.CrowdType)
			req, _ := http.NewRequest(http.MethodGet, u, nil)
			req.Header.Set("Accept", "*/*")
			req.Header.Set("Referer", fmt.Sprintf("https://www.ssense.com/en-us/%s", item.CrowdType))
			resp, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
				EnableProxy: true,
				Reliability: c.CrawlOptions(req.URL).Reliability,
			})
			if err != nil {
				c.logger.Error(err)
				return nil, err
			}
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				c.logger.Error(err)
				return nil, err
			}
			var facts struct {
				Nav map[string]struct {
					Categories []struct {
						ID         int    `json:"id"`
						Name       string `json:"name"`
						SeoKeyword string `json:"seoKeyword"`
						Children   []struct {
							ID         int           `json:"id"`
							Name       string        `json:"name"`
							SeoKeyword string        `json:"seoKeyword"`
							Children   []interface{} `json:"children"`
						} `json:"children"`
					} `json:"categories"`
				} `json:"nav"`
			}
			if err := json.Unmarshal([]byte(respBody), &facts); err != nil {
				return nil, err
			}
			cates := map[int]string{}
			for _, typ := range facts.Nav {
				for _, cate := range typ.Categories {
					cates[cate.ID] = cate.Name
					for _, child := range cate.Children {
						cates[child.ID] = child.Name
					}
				}
			}
			return cates, nil
		}()
		if err != nil {
			c.logger.Error(err)
		}
	}
	if cates == nil {
		cates = map[int]string{}
	}

	prod := viewData.Products.Current
	if prod.Category.ParentID > 0 {
		item.Category = cates[prod.Category.ParentID]
		item.SubCategory = prod.Category.Name
	} else {
		item.Category = prod.Category.Name
	}

	colorname := ""
	matched1 := strings.Split(viewData.Products.Current.Description, "Supplier color:")
	if len(matched1) > 1 {
		matched1 = strings.Split(matched1[1], ".")
		colorname = matched1[0]
	}

	var colorSku *pbItem.SkuSpecOption
	if colorname != "" {
		colorSku = &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecColor,
			Id:    colorname,
			Name:  colorname,
			Value: colorname,
			//Icon:  colorname,
		}
	}

	var medias []*pbMedia.Media
	for ki, mid := range viewData.Products.Current.Images {
		medias = append(medias, pbMedia.NewImageMedia(
			strconv.Format(ki),
			strings.ReplaceAll(mid, "__IMAGE_PARAMS__", "h_1000"),
			strings.ReplaceAll(mid, "__IMAGE_PARAMS__", "h_1000"),
			strings.ReplaceAll(mid, "__IMAGE_PARAMS__", "h_750"),
			strings.ReplaceAll(mid, "__IMAGE_PARAMS__", "h_600"),
			"",
			ki == 0,
		))
	}

	current, _ := strconv.ParseInt(viewData.Products.Current.Price.Sale)
	msrp, _ := strconv.ParseInt(viewData.Products.Current.Price.Regular)
	discount := viewData.Products.Current.Price.Discount
	if current == 0 {
		current = msrp
	}

	for _, rawSku := range viewData.Products.Current.Sizes {
		sku := pbItem.Sku{
			SourceId: strconv.Format(prodid),
			Medias:   medias,
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(current * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},
			Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}
		if rawSku.InStock {
			sku.Stock.StockStatus = pbItem.Stock_InStock
			//sku.Stock.StockCount = int32(rawSku.Number)
		}
		if colorSku != nil {
			sku.Specs = append(sku.Specs, colorSku)
		}
		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Id:    strconv.Format(rawSku.ID),
			Name:  rawSku.Name,
			Value: rawSku.Name,
		})
		item.SkuItems = append(item.SkuItems, &sku)
	}
	if err = yield(ctx, &item); err != nil {
		return err
	}
	return nil
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		// "https://www.ssense.com/en-in/women/bags",
		"https://www.ssense.com/en-us/men/shoes",
		// "https://www.ssense.com/en-in/women/product/burberry/black-econylr-logo-drawcord-pouch/6045701",
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
				i.URL.Host = "www.ssense.com"
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

	ctx := context.WithValue(context.Background(), "tracing_id", fmt.Sprintf("tracing_%d", time.Now().UnixNano()))
	// start the crawl request
	for _, req := range spider.NewTestRequest(context.Background()) {
		if err := callback(ctx, req); err != nil {
			logger.Fatal(err)
		}
	}
}
