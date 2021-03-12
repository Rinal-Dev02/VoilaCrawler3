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
		categoryPathMatcher: regexp.MustCompile(`^((\?!/dp/).)*`),
		productPathMatcher:  regexp.MustCompile(`^(.*)(/dp/)(.*)$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "afaf1423616e408e9daede874a2c0a12"
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
	return []string{"*.revolve.com"}
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

	if !bytes.Contains(respBody, []byte("<li class=\"gc u-center item js-plp-container plp-container")) {
		return errors.New("products not found")
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	lastIndex := nextIndex(ctx)
	sel := doc.Find(`.js-plp-pdp-link.plp__image-link.plp__image-link--lg`)
	c.logger.Debugf("nodes %d", len(sel.Nodes))
	for i := range sel.Nodes {
		node := sel.Eq(i)
		if href, _ := node.Attr("href"); href != "" {
			//c.logger.Debugf("yield %w%s", lastIndex, href)
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
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("pageNum"))
	if page == 0 {
		page = 1
	}

	var trregx = regexp.MustCompile(`updateItemCount\((.*)\);`)
	totalResultCount, _ := strconv.ParseInt(string(trregx.FindSubmatch(respBody)[1]))

	if bytes.Contains(respBody, []byte("pagination__controls link link--noline u-inline-block--top\" href=\"javascript:void(0)\"")) {
		// nextpage not found
		return nil
	}

	if lastIndex >= int(totalResultCount) {
		return nil
	}

	// set pagination
	u := *resp.Request.URL
	vals := u.Query()
	vals.Set("pageNum", strconv.Format(page+1))
	u.RawQuery = vals.Encode()

	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	// update the index of last page
	nctx := context.WithValue(ctx, "item.index", lastIndex)

	return yield(nctx, req)
}

// nextIndex used to get sharingData from context
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")) + 1)
}

// used to trim html labels in description
var htmlTrimRegp = regexp.MustCompile(`</?[^>]+>`)

// Generate data struct from json https://mholt.github.io/json-to-go/
type productDetailPage struct {
	Context string `json:"@context"`
	Type    string `json:"@type"`
	URL     string `json:"url"`
	Name    string `json:"name"`
	Sku     string `json:"sku"`
	Brand   struct {
		Type string `json:"@type"`
		Name string `json:"name"`
	} `json:"brand"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Offers      struct {
		Type            string    `json:"@type"`
		Availability    string    `json:"availability"`
		Price           float64   `json:"price"`
		PriceCurrency   string    `json:"priceCurrency"`
		PriceValidUntil time.Time `json:"priceValidUntil"`
		URL             string    `json:"url"`
	} `json:"offers"`
	AggregateRating struct {
		Type        string  `json:"@type"`
		RatingValue float64 `json:"ratingValue"`
		ReviewCount int     `json:"reviewCount"`
		BestRating  int     `json:"bestRating"`
		WorstRating int     `json:"worstRating"`
	} `json:"aggregateRating"`
	Review []struct {
		Type          string `json:"@type"`
		Author        string `json:"author"`
		DatePublished string `json:"datePublished"`
		Description   string `json:"description"`
		ReviewRating  struct {
			Type        string `json:"@type"`
			RatingValue int    `json:"ratingValue"`
			BestRating  int    `json:"bestRating"`
			WorstRating int    `json:"worstRating"`
		} `json:"reviewRating"`
	} `json:"review"`
}

var productsDataExtractReg = regexp.MustCompile(`(?U)<script type="application/ld\+json">\s*({.*})\s*</script>`)

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

	matched := productsDataExtractReg.FindSubmatch(respbody)
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respbody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respbody))
	if err != nil {
		return err
	}

	var p productDetailPage
	if err := json.Unmarshal(matched[1], &p); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		//return err //check
	}

	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:       p.Sku,
			CrawlUrl: resp.Request.URL.String(),
		},
		Title:       p.Name,
		Description: htmlTrimRegp.ReplaceAllString(doc.Find(`.product-details__description-content`).Text(), ""),
		BrandName:   p.Brand.Name,
		CrowdType:   "",
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
	}

	//itemListElement
	sel := doc.Find(`.crumbs>li`)
	c.logger.Debugf("nodes %d", len(sel.Nodes))
	for i := range sel.Nodes {
		node := sel.Eq(i)
		breadcrumb := strings.TrimSpace(node.Text())

		if i == 0 {
			item.Category = breadcrumb
		} else if i == 1 {
			item.SubCategory = breadcrumb
		} else if i == 2 {
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

	sel = doc.Find(".size-options__item>input")
	for i := range sel.Nodes {
		snode := sel.Eq(i)

		sizeval, _ := snode.Attr(`data-size`)
		qty, _ := snode.Attr(`data-qty`)
		quantity, _ := strconv.ParseInt(qty)

		currentPrice, _ := snode.Attr(`data-regular-price`)
		cp, _ := strconv.ParseFloat(currentPrice)

		msrp, _ := snode.Attr(`data-price`)
		mp, _ := strconv.ParseFloat(msrp)

		discount := 0.0
		if mp > 0.0 {
			discount = ((cp - mp) / mp) * 100
		} else {
			msrp = currentPrice
		}

		sku := pbItem.Sku{
			SourceId: p.Sku,
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(cp * 100),
				Msrp:     int32(mp * 100),
				Discount: int32(discount),
			},
			Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}

		if quantity > 0 {
			sku.Stock.StockStatus = pbItem.Stock_InStock
			sku.Stock.StockCount = int32(quantity)
		}

		// color
		colorName := doc.Find(`.u-font-primary.u-margin-l--md.selectedColor.u-text--xl`).Text()

		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecColor,
			Id:    strconv.Format(i),
			Name:  colorName,
			Value: colorName,
			//Icon:  color.SwatchMedia.Mobile,
		})

		if i == 0 {
			isDefault := true
			sel = doc.Find(`.js-primary-slideshow__pager-thumb.slideshow__pager-thumb`)
			for j := range sel.Nodes {
				if j > 1 {
					isDefault = false
				}

				node := sel.Eq(j)

				if href, _ := node.Attr("style"); href == "display:none" {
					continue
				}

				mediumUrl, _ := node.Attr("data-image")

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
			Id:    strconv.Format(i),
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
		"https://www.revolve.com/skirts/br/8b6a66/?navsrc=subclothing",
		//"https://www.revolve.com/nphilanthropy-scarlett-leather-jogger-in-camel/dp/PHIR-WP63/?d=Womens&sectionURL=%2Fpants%2Fbr%2F44d522%2F%3F%26s%3Dc%26c%3DPants%26navsrc%3Dsubclothing&code=PHIR-WP63",
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
				i.URL.Host = "www.revolve.com"
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
