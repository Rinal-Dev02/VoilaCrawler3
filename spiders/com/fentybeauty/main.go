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

	"github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/item"

	pbMedia "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/media"
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
		categoryPathMatcher: regexp.MustCompile(`^(/[a-z0-9-]+){1,6}$`),
		// this regular used to match product page url path
		productPathMatcher: regexp.MustCompile(`^(/[A-Za-z0-9-]+){1,4}.html$`),
		logger:             logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "64ab074281c34284a357dccf74c2625d"
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
	return []string{"*.fentybeauty.com"}
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
var productsExtractReg = regexp.MustCompile(`(?U)"impressions":\s*(\[.*\]),`)

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

	if !bytes.Contains(respBody, []byte("class=\"grid-tile")) {
		return errors.New("products not found")
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	lastIndex := nextIndex(ctx)
	sel := doc.Find(`.thumb-link`)
	c.logger.Debugf("nodes %d", len(sel.Nodes))
	for i := range sel.Nodes {
		node := sel.Eq(i)
		if href, _ := node.Attr("href"); href != "" {
			//c.logger.Debugf("yield %s", href)
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

	if !bytes.Contains(respBody, []byte("<div class=\"infinite-scroll-placeholder\"")) {
		// nextpage not found
		return nil
	}

	nexturl, _ := doc.Find(`.infinite-scroll-placeholder`).Attr(`data-grid-url`)
	nexturl = strings.ReplaceAll(nexturl, "&sz=12", "&sz=48")

	req, _ := http.NewRequest(http.MethodGet, nexturl, nil)
	// update the index of last page
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
}

// used to trim html labels in description
var htmlTrimRegp = regexp.MustCompile(`</?[^>]+>`)

type ImageDetail struct {
	Large []struct {
		Zoom  string `json:"zoom"`
		URL   string `json:"url"`
		Alt   string `json:"alt"`
		Title string `json:"title"`
	} `json:"large"`
}

// parseProduct
func (c *_Crawler) parseProduct(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil {
		return nil
	}

	if resp.StatusCode == http.StatusForbidden {
		return errors.New("access denied")
	}

	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if isRobotCheckPage(respbody) {
		return errors.New("robot check page")
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respbody))
	if err != nil {
		return err
	}

	productStyleSourceID, _ := doc.Find(`div[data-bv-show="rating_summary"]`).Attr(`data-bv-productId`)

	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:       productStyleSourceID,
			CrawlUrl: resp.Request.URL.String(),
		},
		Title:       doc.Find(`.product-name`).Text(),
		Description: htmlTrimRegp.ReplaceAllString(doc.Find(`div[itemprop="description"]`).Text(), ""),
		BrandName:   "Fenty Beauty",
		CrowdType:   "",
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
	}

	//itemListElement
	sel := doc.Find(`.breadcrumb>li`)
	//c.logger.Debugf("nodes %d", len(sel.Nodes))
	for i := range sel.Nodes {
		node := sel.Eq(i)
		breadcrumb := strings.TrimSpace(node.Text())

		if i == 1 {
			item.Category = breadcrumb
		} else if i == 2 {
			item.SubCategory = breadcrumb
		} else if i == 3 {
			item.SubCategory2 = breadcrumb
		}
	}

	currentPrice := int64(0)
	msrp := int64(0)

	replacer := strings.NewReplacer("Original price:", "", "$", "", "Sale price:", "", "</?[^>]+>", "")

	if bytes.Contains(respbody, []byte("Original price:</span>")) {
		msrp, _ = strconv.ParseInt(replacer.Replace(doc.Find(`.price-standard`).First().Text()))
		currentPrice, _ = strconv.ParseInt(replacer.Replace(doc.Find(`.price-sales`).First().Text()))
	} else {
		currentPrice, _ = strconv.ParseInt(replacer.Replace(doc.Find(`.price-sales`).First().Text()))
		msrp, _ = strconv.ParseInt(replacer.Replace(doc.Find(`.price-sales`).First().Text()))
	}

	discount := int64(0)
	if msrp > 0 {
		discount = ((currentPrice - msrp) / msrp) * 100
	}

	if strings.Contains(resp.Request.URL.String(), "_color") || bytes.Contains(respbody, []byte("class=\"swatches Color")) {

		if bytes.Contains(respbody, []byte("class=\"swatches Color filtered\"")) {
			sel = doc.Find(`.swatches.Color.filtered`).First().Find(`li`)
		} else if bytes.Contains(respbody, []byte("class=\"swatches Color")) {
			sel = doc.Find(`.swatches.Color`).First().Find(`li`)
		} else {
			sel = doc.Find(`.swatches.Color`).First().Find(`li`)
		}

		for i := range sel.Nodes {
			snode := sel.Eq(i)
			colorSourceid, _ := snode.Find(`a`).Attr(`data-vv-id`)
			colorval, _ := snode.Find(`a`).Attr(`title`)
			class, _ := snode.Attr(`class`)

			sku := pbItem.Sku{
				SourceId: colorSourceid,
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(currentPrice * 100),
					Msrp:     int32(msrp * 100),
					Discount: int32(discount),
				},
				Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}

			if !strings.Contains(class, `unselectable`) {
				sku.Stock.StockStatus = pbItem.Stock_InStock
				//sku.Stock.StockCount = int32(rawSku.AvailableDc)
			}

			// color
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    colorSourceid,
				Name:  colorval,
				Value: colorval,
				//Icon:  color.SwatchMedia.Mobile,
			})

			imgval, _ := snode.Find(`a`).Attr(`data-allimages`)

			var imgData ImageDetail
			if err := json.Unmarshal([]byte(imgval), &imgData); err != nil {
				c.logger.Debugf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
				//return err
			}

			isDefault := true

			for j, mediumUrl := range imgData.Large {
				if j > 1 {
					isDefault = false
				}

				sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
					strconv.Format(j),
					mediumUrl.Zoom,
					mediumUrl.Zoom+"?sw=1000",
					mediumUrl.Zoom+"?sw=600",
					mediumUrl.Zoom+"?sw=500",
					"", isDefault))
			}

			// // size - no size detail available
			// sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			// 	Type:  pbItem.SkuSpecType_SkuSpecSize,
			// 	Id:    sizeSourceid,
			// 	Name:  sizeval,
			// 	Value: sizeval,
			// })

			item.SkuItems = append(item.SkuItems, &sku)
		}

	} else {
		sizeval := "One Size"
		Sourceid, _ := doc.Find(`#pid`).Attr(`value`)
		sku := pbItem.Sku{
			SourceId: Sourceid,
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(currentPrice * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},
			Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}

		if bytes.Contains(respbody, []byte("<p class=\"in-stock-msg\">In Stock</p>")) {
			sku.Stock.StockStatus = pbItem.Stock_InStock
		}

		if bytes.Contains(respbody, []byte("class=\"product-thumbnails-carousel")) {
			sel = doc.Find(`ul[class^="product-thumbnails-carousel"]`).Find(`li`).Find(`img`)
		} else {
			sel = doc.Find(`#product-set-list>img`)
		}

		isDefault := true

		for j := range sel.Nodes {
			if j > 1 {
				isDefault = false
			}
			node := sel.Eq(j)
			mediumUrl, _ := node.Attr("data-src")

			if mediumUrl == "" {
				mediumUrl, _ = node.Attr("srcset")
				if mediumUrl == "" {
					continue
				}
			}
			s := strings.Split(mediumUrl, ".jpg")

			sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
				strconv.Format(j),
				s[0]+".jpg",
				s[0]+".jpg?sw=1000",
				s[0]+".jpg?sw=600",
				s[0]+".jpg?sw=500",
				"", isDefault))
		}

		// // size
		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Id:    Sourceid,
			Name:  sizeval,
			Value: sizeval,
		})

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
		//"https://www.fentybeauty.com/makeup-face",
		"https://www.fentybeauty.com/powder-puff-setting-brush-170/27464.html?cgid=makeup-face-powder",
		//"https://www.fentybeauty.com/pro-filtr-instant-retouch-setting-powder/FB30011.html?dwvar_FB30011_color=FB9005&cgid=makeup-face-powder",
		//"https://www.fentybeauty.com/soft-matte-complexion-essentials-with-brush/pro-filter-foundation-essentials-brush.html?cgid=makeup-face",
		//"https://www.fentybeauty.com/pro-filtr-soft-matte-longwear-foundation/FB30006.html?dwvar_FB30006_color=FB0340&cgid=makeup-face-foundation",
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

			// set scheme,host for sub requests. for the product url in category page is just the path without hosts info.
			// here is just the test logic. when run the spider online, the controller will process automatically
			if i.URL.Scheme == "" {
				i.URL.Scheme = "https"
			}
			if i.URL.Host == "" {
				i.URL.Host = "www.fentybeauty.com"
			}

			// do http requests here.
			nctx, cancel := context.WithTimeout(ctx, time.Minute*5)
			defer cancel()
			resp, err := client.DoWithOptions(nctx, i, http.Options{
				EnableProxy:       false,
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
