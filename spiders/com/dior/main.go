package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/voiladev/VoilaCrawl/pkg/crawler"
	"github.com/voiladev/VoilaCrawl/pkg/net/http"
	"github.com/voiladev/VoilaCrawl/pkg/net/http/cookiejar"
	"github.com/voiladev/VoilaCrawl/pkg/proxy"

	media "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/media"
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
		categoryPathMatcher: regexp.MustCompile(`^((\?!products).)*`),
		// this regular used to match product page url path
		productPathMatcher: regexp.MustCompile(`^(.*)(products)(.*)$`),
		logger:             logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "ea84f5d3688142cb9400227d2672ceb7"
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
	return []string{"*.dior.com"}
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

type CategoryStructure struct {
	Props struct {
		InitialReduxState struct {
			ALGOLIA struct {
				Searches map[string]struct {
					//MenRtwSuitsJackets struct {
					Result struct {
						NbHits      int `json:"nbHits"`
						Page        int `json:"page"`
						NbPages     int `json:"nbPages"`
						HitsPerPage int `json:"hitsPerPage"`
						Hits        []struct {
							Type       string `json:"type"`
							Columns    int    `json:"columns,omitempty"`
							ObjectID   string `json:"objectID"`
							Attributes struct {
								ProductLink struct {
									URI string `json:"uri"`
								} `json:"productLink"`
							} `json:"attributes,omitempty"`
						} `json:"hits"`
					} `json:"result"`
					//} `json:"men-rtw-suits-jackets"`
				} `json:"searches"`
			} `json:"ALGOLIA"`
		} `json:"initialReduxState"`
	} `json:"props"`
}

// used to extract embaded json data in website page.
// more about golang regulation see here https://golang.org/pkg/regexp/syntax/
var productsExtractReg = regexp.MustCompile(`(?U)id="__NEXT_DATA__"\s*type="application/json">\s*({.*})\s*</script>`)

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

	searchString := ""
	lastIndex := nextIndex(ctx)
	for key, is := range viewData.Props.InitialReduxState.ALGOLIA.Searches {

		searchString = key
		for _, idv := range is.Result.Hits {

			if idv.Type == "EDITOIMAGE" || idv.Type == "EDITOCONTENT" {
				continue
			}

			rawurl := ""
			if idv.Attributes.ProductLink.URI != "" {
				rawurl = fmt.Sprintf("%s://%s%s", resp.Request.URL.Scheme, resp.Request.URL.Host, idv.Attributes.ProductLink.URI)

			} else {
				rawurl = fmt.Sprintf("%s://%s/en_us/products/couture-%s", resp.Request.URL.Scheme, resp.Request.URL.Host, idv.ObjectID)
			}
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
	// get current page number
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}
	// check if this is the last page
	if len(viewData.Props.InitialReduxState.ALGOLIA.Searches[searchString].Result.Hits) >= viewData.Props.InitialReduxState.ALGOLIA.Searches[searchString].Result.NbHits ||
		page >= int64(viewData.Props.InitialReduxState.ALGOLIA.Searches[searchString].Result.NbPages) {
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

// used to trim html labels in description
var htmlTrimRegp = regexp.MustCompile(`</?[^>]+>`)

type parseProductData struct {
	Props struct {
		InitialReduxState struct {
			CONTENT struct {
				CmsContent struct {
					Type     string `json:"type"`
					Elements []struct {
						Type            string `json:"type"`
						Text            string `json:"text,omitempty"`
						BackgroundColor string `json:"backgroundColor,omitempty"`
						TextColor       string `json:"textColor,omitempty"`
						TimeToLive      int    `json:"timeToLive,omitempty"`
						Capping         int    `json:"capping,omitempty"`
						NodeID          string `json:"nodeId,omitempty"`
						CallToAction    struct {
							Type  string `json:"type"`
							Title string `json:"title"`
							URI   string `json:"uri"`
						} `json:"callToAction,omitempty"`
						Items []struct {
							Type     string `json:"type"`
							Title    string `json:"title"`
							TitleInt string `json:"titleInt,omitempty"`
							URL      string `json:"url"`
							Images   []struct {
								Target   string `json:"target"`
								URI      string `json:"uri"`
								Width    int    `json:"width"`
								Height   int    `json:"height"`
								ViewCode string `json:"viewCode"`
								Alt      string `json:"alt"`
							} `json:"images"`
						} `json:"items,omitempty"`
						Title             string        `json:"title,omitempty"`
						Subtitle          string        `json:"subtitle,omitempty"`
						Reference         string        `json:"reference,omitempty"`
						Tags              []interface{} `json:"tags,omitempty"`
						Universe          string        `json:"universe,omitempty"`
						PrimaryCategoryID string        `json:"primaryCategoryId,omitempty"`
						//ProductType          string        `json:"productType,omitempty"`
						PreorderShippingDate interface{} `json:"preorderShippingDate,omitempty"`
						VariationsType       string      `json:"variationsType,omitempty"`
						Code                 string      `json:"code,omitempty"`
						Color                string      `json:"color,omitempty"`
						Variations           []struct {
							Title  string `json:"title"`
							Code   string `json:"code"`
							Sku    string `json:"sku"`
							Detail string `json:"detail"`
							Status string `json:"status"`
							Price  struct {
								Value    int    `json:"value"`
								Currency string `json:"currency"`
							} `json:"price"`
							Tracking []struct {
								Events        []string `json:"events"`
								AddToCartType string   `json:"addToCartType"`
								PageType      string   `json:"pageType"`
								Ecommerce     struct {
									CurrencyCode string `json:"currencyCode"`
									Add          struct {
										Products []struct {
											ID          string `json:"id"`
											Name        string `json:"name"`
											Price       int    `json:"price"`
											Brand       string `json:"brand"`
											Category    string `json:"category"`
											Variant     string `json:"variant"`
											Quantity    int    `json:"quantity"`
											Dimension16 string `json:"dimension16"`
											Dimension32 string `json:"dimension32"`
										} `json:"products"`
									} `json:"add"`
								} `json:"ecommerce"`
							} `json:"tracking"`
						} `json:"variations,omitempty"`
						Sections []struct {
							TitleKey string `json:"titleKey"`
							Content  string `json:"content"`
							Type     string `json:"type"`
						} `json:"sections,omitempty"`
						Declinations []struct {
							Title     string `json:"title"`
							Color     string `json:"color"`
							ColorCode string `json:"colorCode"`
							URI       string `json:"uri"`
							Image     struct {
								Target   string `json:"target"`
								URI      string `json:"uri"`
								Width    int    `json:"width"`
								Height   int    `json:"height"`
								ViewCode string `json:"viewCode"`
								Alt      string `json:"alt"`
							} `json:"image"`
						} `json:"declinations,omitempty"`
						Images []struct {
							Width  int    `json:"width"`
							Height int    `json:"height"`
							Alt    string `json:"alt"`
							URI    string `json:"uri"`
							Target string `json:"target"`
						} `json:"images,omitempty"`
						Price struct {
							Value    int    `json:"value"`
							Currency string `json:"currency"`
						} `json:"price"`
						Sku         string      `json:"sku"`
						Status      string      `json:"status"`
						SizeLabel   string      `json:"sizeLabel"`
						Description interface{} `json:"description,omitempty"`
					} `json:"elements"`
				} `json:"cmsContent"`
			} `json:"CONTENT"`
		} `json:"initialReduxState"`
	} `json:"props"`
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

	var viewData parseProductData

	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}

	contentIndex := getIndex(viewData, "PRODUCTTITLE")
	contentData := viewData.Props.InitialReduxState.CONTENT.CmsContent.Elements[contentIndex]
	colorName := contentData.Subtitle

	// color
	var itemColor pbItem.SkuSpecOption
	itemColor = pbItem.SkuSpecOption{
		Type:  pbItem.SkuSpecType_SkuSpecColor,
		Id:    strconv.Format(contentData.Reference),
		Name:  colorName,
		Value: colorName,
	}

	contentIndex = getIndex(viewData, "PRODUCTSECTIONDESCRIPTION")
	contentDescriptionData := viewData.Props.InitialReduxState.CONTENT.CmsContent.Elements[contentIndex]

	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:       strconv.Format(contentData.Title),
			CrawlUrl: resp.Request.URL.String(),
		},
		BrandName:   "DIOR",
		Title:       contentData.Reference,
		Description: htmlTrimRegp.ReplaceAllString(contentDescriptionData.Sections[0].Content, ""),
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
	}

	// image
	var itemImg []*media.Media
	contentIndex = getIndex(viewData, "PRODUCTMEDIAS")
	contentImgData := viewData.Props.InitialReduxState.CONTENT.CmsContent.Elements[contentIndex]
	isDefault := true
	for ki, mid := range contentImgData.Items {
		if ki > 0 {
			isDefault = false
		}

		imgsize := "/" + strconv.Format(mid.Images[0].Width) + "x" + strconv.Format(mid.Images[0].Height) + "/"
		iZoom := strings.NewReplacer(imgsize, "/3000x2000/", "_ZHC.", "_ZH.", "_GH.", "_ZH.", "cover_image_", "zoom_image_", "grid_image_", "zoom_image_")
		iMid := strings.NewReplacer(imgsize, "/870x580/", "_ZH.", "_ZHC.", "_GH.", "_ZHC.", "grid_image_", "cover_image_")
		iSmall := strings.NewReplacer(imgsize, "/460x497/", "_ZHC.", "_GH.", "_ZH.", "_GH.", "cover_image_", "grid_image_")
		itemImg = append(itemImg, pbMedia.NewImageMedia(
			strconv.Format(ki),
			mid.Images[0].URI,
			iZoom.Replace(mid.Images[0].URI),
			iMid.Replace(mid.Images[0].URI),
			iSmall.Replace(mid.Images[0].URI),
			"",
			isDefault,
		))
	}

	contentIndex = getIndex(viewData, "PRODUCTVARIATIONS")
	if contentIndex > -1 {
		contentData = viewData.Props.InitialReduxState.CONTENT.CmsContent.Elements[contentIndex]

		if contentData.VariationsType != "SIZE" {
			fmt.Println("Variation is not type of SIZE")
		}

		for ks, rawSku := range contentData.Variations {
			originalPrice := (rawSku.Price.Value)
			//discount, _ := strconv.ParseInt(strings.TrimSuffix(rawSku.DisplayPercentOff, "%"))
			sku := pbItem.Sku{
				SourceId: strconv.Format(rawSku.Sku),
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(originalPrice * 100),
					Msrp:     int32(originalPrice * 100),
				},
				Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}
			if rawSku.Status == "AVAILABLE" {
				sku.Stock.StockStatus = pbItem.Stock_InStock
			}

			// color
			sku.Specs = append(sku.Specs, &itemColor)

			//image
			if ks == 0 {
				sku.Medias = itemImg
			}

			// size
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    rawSku.Sku,
				Name:  rawSku.Title,
				Value: rawSku.Title,
			})

			item.SkuItems = append(item.SkuItems, &sku)
		}
	} else {
		// no size variation

		contentIndex = getIndex(viewData, "PRODUCTUNIQUE")
		if contentIndex > -1 {

			contentData = viewData.Props.InitialReduxState.CONTENT.CmsContent.Elements[contentIndex]

			originalPrice := (contentData.Price.Value)
			//discount, _ := strconv.ParseInt(strings.TrimSuffix(rawSku.DisplayPercentOff, "%"))
			sku := pbItem.Sku{
				SourceId: strconv.Format(contentData.Sku),
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(originalPrice * 100),
					Msrp:     int32(originalPrice * 100),
				},
				Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}
			if contentData.Status == "AVAILABLE" {
				sku.Stock.StockStatus = pbItem.Stock_InStock
			}

			// color
			sku.Specs = append(sku.Specs, &itemColor)

			//image
			sku.Medias = itemImg

			// size
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    contentData.Sku,
				Name:  contentData.SizeLabel,
				Value: contentData.SizeLabel,
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

func getIndex(viewData parseProductData, types string) int {
	cindex := -1

	for i, raw := range viewData.Props.InitialReduxState.CONTENT.CmsContent.Elements {
		if types == raw.Type {
			cindex = i
		}
	}
	return cindex
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://www.dior.com/en_us/womens-fashion/ready-to-wear/all-ready-to-wear",
		// "https://www.dior.com/en_us/products/couture-124V03BM211_X5685-ribbed-knit-bar-jacket-navy-blue-double-breasted-virgin-wool",
		"https://www.dior.com/en_us/products/couture-93C1046A0121_C975-dior-oblique-tie-blue-and-black-silk",
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
				i.URL.Host = "www.dior.com"
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
