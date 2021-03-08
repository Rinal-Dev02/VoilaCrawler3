package main

// this website exists api robot check. should controller frequence

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
	"github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/regulation"

	pbMedia "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/media"
	pbItem "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
)

type _Crawler struct {
	httpClient http.Client

	categoryPathMatcher *regexp.Regexp
	productPathMatcher  *regexp.Regexp
	logger              glog.Log
}

func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient:          client,
		categoryPathMatcher: regexp.MustCompile(`^((\?!product).)*`),
		productPathMatcher:  regexp.MustCompile(`^(.*)(/product/)(.*)$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "d69b784496654212a96587c2ebcc152b"
}

// Version
func (c *_Crawler) Version() int32 {
	return 1
}

// CrawlOptions
func (c *_Crawler) CrawlOptions() *crawler.CrawlOptions {
	options := crawler.NewCrawlOptions()
	options.EnableHeadless = false
	options.LoginRequired = false
	options.Reliability = 1
	// NOTE: no need to set useragent here for user agent is dynamic
	// options.MustHeader.Set("Accept-Language", "en-US,en;q=0.8")
	// options.MustHeader.Set("X-Requested-With", "XMLHttpRequest")

	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.finishline.com"}
}

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

func isRobotCheckPage(respBody []byte) bool {
	return bytes.Contains(respBody, []byte("we believe you are using automation tools to browse the website")) ||
		bytes.Contains(respBody, []byte("Javascript is disabled or blocked by an extension")) ||
		bytes.Contains(respBody, []byte("Your browser does not support cookies"))
}

const defaultCategoryProductsPageSize = 40

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

	if !bytes.Contains(respBody, []byte("product-card__details")) {
		return errors.New("products not found")
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	lastIndex := nextIndex(ctx)
	sel := doc.Find(`.product-card__details>a`)
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
	// if len(sel.Nodes) < defaultCategoryProductsPageSize {
	// 	return nil
	// }

	if bytes.Contains(respBody, []byte("button pag-button next light-gray ml-1 disabled")) {
		// nextpage not found
		return nil
	}

	sel1, _ := doc.Find(`.button.pag-button.next.light-gray.ml-1`).Attr("href")

	//fmt.Println(sel1)
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	req, _ := http.NewRequest(http.MethodGet, sel1, nil)
	return yield(nctx, req)
}

// nextIndex used to get sharingData from context
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")) + 1)
}

// Generate data struct from json https://mholt.github.io/json-to-go/
type productDetailPage struct {
}

var productsExtractReg = regexp.MustCompile(`(?U)var\s*utag_data\s*=\s*({.*});\s*`)

func (c *_Crawler) parseProduct(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
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

	productStyleSourceID, _ := doc.Find(`#productStyleId`).Attr(`value`)
	var brandregx = regexp.MustCompile(`"product_brand"\s*:\s*\["(.*)"\],`)
	brandName := string(brandregx.FindSubmatch(respbody)[1])

	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:       productStyleSourceID,
			CrawlUrl: resp.Request.URL.String(),
		},
		Title:       doc.Find(`.hmb-2.titleDesk`).Text(),
		Description: doc.Find(`#productDescription`).Text(),
		BrandName:   brandName,
		CrowdType:   "",
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
	}

	//itemListElement
	sel := doc.Find(`.breadcrumbs>li`)
	c.logger.Debugf("nodes %d", len(sel.Nodes))
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

	for _, v := range []string{"man", "men", "male"} {
		if strings.Contains(strings.ToLower(item.Category), v) {
			item.CrowdType = "men"
			break
		}
	}

	for _, v := range []string{"woman", "women", "female"} {
		if strings.Contains(strings.ToLower(item.Category), v) {
			item.CrowdType = "women"
			break
		}
	}

	for _, v := range []string{"kid", "child", "girl", "boy"} {
		if strings.Contains(strings.ToLower(item.Category), v) {
			item.CrowdType = "kids"
			break
		}
	}

	var sizeID = regexp.MustCompile(`"product_id"\s*:\s*\["(.*)"\],`)
	sizelist := "#sizes_" + string(sizeID.FindSubmatch(respbody)[1]) + ">div>button"

	var priceregx = regexp.MustCompile(`"product_unit_price"\s*:\s*\["(.*)"\],`)
	currentPrice, _ := strconv.ParseFloat(string(priceregx.FindSubmatch(respbody)[1]))

	priceregx = regexp.MustCompile(`"product_list_price"\s*:\s*\["(.*)"\],`)
	msrp, _ := strconv.ParseFloat(string(priceregx.FindSubmatch(respbody)[1]))

	discount := ((currentPrice - msrp) / msrp) * 100

	sel = doc.Find(sizelist)
	for i := range sel.Nodes {
		snode := sel.Eq(i)
		sizeSourceid, _ := snode.Attr(`data-val`)
		sizeval, _ := snode.Attr(`data-size`)

		sku := pbItem.Sku{
			SourceId: sizeSourceid,
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(currentPrice * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},
			Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}

		if !strings.Contains(snode.Text(), `disabled`) {
			sku.Stock.StockStatus = pbItem.Stock_InStock
			//sku.Stock.StockCount = int32(rawSku.AvailableDc)
		}

		// color
		colorSourceID, _ := doc.Find(`#productColorId`).Attr(`value`)
		colorName, _ := doc.Find(`#tfc_colorid`).Attr(`value`)

		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecColor,
			Id:    colorSourceID,
			Name:  colorName,
			Value: colorName,
			//Icon:  color.SwatchMedia.Mobile,
		})

		if i == 0 {
			isDefault := true
			sel = doc.Find(`.thumbSlide`)
			for j := range sel.Nodes {
				if j > 1 {
					isDefault = false
				}
				node := sel.Eq(j)
				mediumUrl, _ := node.Find(`.over5.pdp-image.isShoe`).Attr("data-thumb")

				sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
					strconv.Format(j),
					mediumUrl,
					mediumUrl+"&w=1000&&h=1000",
					mediumUrl+"&w=700&&h=700",
					mediumUrl+"&w=600&&h=600",
					"", isDefault))
			}
		}

		// size
		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Id:    sizeSourceid,
			Name:  sizeval,
			Value: sizeval,
		})

		item.SkuItems = append(item.SkuItems, &sku)
	}

	if err := yield(ctx, &item); err != nil {
		return err
	}

	return nil
}

func (c *_Crawler) NewTestRequest(ctx context.Context) []*http.Request {
	var reqs []*http.Request
	for _, u := range []string{
		"https://www.finishline.com/store/women/shoes/running/_/N-nat3jh?mnid=women_shoes_running",
		"https://www.finishline.com/store/product/womens-nike-air-max-270-casual-shoes/prod2770847?styleId=AH6789&colorId=001",
		//"https://www.finishline.com/store/product/womens-puma-future-rider-play-on-casual-shoes/prod2795926?styleId=38182501&colorId=100",
		// "https://www.neimanmarcus.com/p/moncler-moka-shiny-fitted-puffer-coat-with-hood-and-matching-items-prod213210002?childItemId=NMTA8BE_&focusProductId=prod180340224&navpath=cat000000_cat000001_cat58290731_cat77190754&page=0&position=27",
	} {
		req, _ := http.NewRequest(http.MethodGet, u, nil)
		reqs = append(reqs, req)
	}
	return reqs
}

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
				i.URL.Host = "www.finishline.com"
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
