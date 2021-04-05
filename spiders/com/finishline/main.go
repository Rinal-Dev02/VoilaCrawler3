package main

// this website exists api robot check. should controller frequence

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
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
	"github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/regulation"

	pbMedia "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/media"
	pbItem "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	pbProxy "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type _Crawler struct {
	httpClient http.Client

	categoryPathMatcher *regexp.Regexp
	productPathMatcher  *regexp.Regexp
	logger              glog.Log
}

func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient: client,
		// /store/women/shoes/running/_/N-nat3jh
		categoryPathMatcher: regexp.MustCompile(`^/store(/[a-zA-Z0-9\-_]+){1,5}$`),
		productPathMatcher:  regexp.MustCompile(`^((.*)(/product/)(.*))|(/store/product/(.*)(/prod\d+))$`),
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
func (c *_Crawler) CrawlOptions(u *url.URL) *crawler.CrawlOptions {
	opts := crawler.NewCrawlOptions()
	opts.EnableHeadless = false
	opts.LoginRequired = false
	opts.EnableSessionInit = false
	opts.Reliability = pbProxy.ProxyReliability_ReliabilityDefault
	return opts
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
	if c == nil {
		return nil
	}

	if resp.StatusCode == http.StatusForbidden {
		return errors.New("access denied")
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	lastIndex := nextIndex(ctx)

	parseItem := func(sel *goquery.Selection) error {
		c.logger.Debugf("nodes %d", len(sel.Nodes))
		for i := range sel.Nodes {
			node := sel.Eq(i)
			if href, _ := node.Attr("href"); href != "" {
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
		return nil
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	if err := parseItem(doc.Find(`.product-card>a`)); err != nil {
		c.logger.Error(err)
		return err
	}
	subDom, err := goquery.NewDocumentFromReader(strings.NewReader(doc.Find("#additionalProducts").Text()))
	if err != nil {
		c.logger.Error(err)
		return err
	}
	if err := parseItem(subDom.Find(`.product-card>a`)); err != nil {
		c.logger.Error(err)
		return err
	}

	nextSel := doc.Find(`.downPagination .pag-button.next`)
	if strings.Contains(nextSel.AttrOr("class", ""), "disabled") {
		return nil
	}
	nextUrl := nextSel.AttrOr("href", "")
	if nextUrl == "" {
		return nil
	}

	nctx := context.WithValue(ctx, "item.index", lastIndex)
	req, _ := http.NewRequest(http.MethodGet, nextUrl, nil)
	return yield(nctx, req)
}

// nextIndex used to get sharingData from context
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")) + 1)
}

var (
	imgWidthTplReg  = regexp.MustCompile(`&+w=\d+`)
	imgHeightTplReg = regexp.MustCompile(`&+h=\d+`)
)

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

	opts := c.CrawlOptions(resp.Request.URL)
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respbody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	var brandregx = regexp.MustCompile(`"(?U)product_brand"\s*:\s*\["(.*)"\],`)
	matched := brandregx.FindSubmatch(respbody)
	if len(matched) < 2 {
		return fmt.Errorf("not brand found")
	}
	brandName := strings.TrimSpace(string(matched[1]))

	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           doc.Find("#productItemId").AttrOr("value", doc.Find("#tfc_productid").AttrOr("value", "")),
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: doc.Find(`link[rel="canonical"]`).AttrOr("href", ""),
		},
		Title:       doc.Find(`.hmb-2.titleDesk`).Text(),
		Description: doc.Find(`#productDescription`).Text(),
		BrandName:   brandName,
		CrowdType:   "",
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
	}

	{
		sel := doc.Find(`.breadcrumbs>li`)
		for i := range sel.Nodes {
			node := sel.Eq(i)
			breadcrumb := strings.TrimSpace(node.Text())

			if i == 1 {
				item.Category = breadcrumb
			} else if i == 2 {
				item.SubCategory = breadcrumb
			} else if i == 3 {
				item.SubCategory2 = breadcrumb
			} else if i == 4 {
				item.SubCategory3 = breadcrumb
			} else if i == 5 {
				item.SubCategory4 = breadcrumb
			}
		}
	}

	colorSel := doc.Find(`#alternateColors .colorway`)
	for i := range colorSel.Nodes {
		node := colorSel.Eq(i)
		style := strings.TrimSpace(node.Find(`a`).AttrOr("data-styleid", node.Find(`a`).AttrOr("data-productid", "")))

		var medias []*pbMedia.Media
		if !strings.Contains(node.Find(`a>.color-image`).AttrOr(`class`, ""), "selected") {
			// load images
			rawurl := fmt.Sprintf("https://www.finishline.com/store/browse/gadgets/alternateImage.jsp?colorID=%s&styleID=%s&productName=&productItemId=%s&productIsShoe=true&productIsAccessory=false&productIsGiftCard=false&renderType=desktop&pageName=pdp", style, style, item.Source.Id)
			req, _ := http.NewRequest(http.MethodGet, rawurl, nil)
			req.Header.Set("Referer", resp.Request.URL.String())
			if resp.Request.Header.Get("User-Agent") != "" {
				req.Header.Set("User-Agent", resp.Request.Header.Get("User-Agent"))
			}
			c.logger.Debugf("Access images %s", rawurl)
			resp, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
				EnableProxy:       true,
				EnableHeadless:    opts.EnableHeadless,
				EnableSessionInit: opts.EnableSessionInit,
				Reliability:       opts.Reliability,
			})
			if err != nil {
				c.logger.Error(err)
				return err
			}
			if resp.StatusCode != 200 {
				data, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("%d %s", resp.StatusCode, data)
			}
			dom, err := goquery.NewDocumentFromReader(resp.Body)
			resp.Body.Close()
			if err != nil {
				c.logger.Error(err)
				return err
			}

			seli := dom.Find(`#thumbSlides .thumbSlide .pdp-image`)
			for j := range seli.Nodes {
				node := seli.Eq(j)
				murl := node.AttrOr("data-large", node.AttrOr("data-thumb", ""))
				if murl == "" {
					continue
				}
				medias = append(medias, pbMedia.NewImageMedia(
					strconv.Format(j),
					imgHeightTplReg.ReplaceAllString(imgWidthTplReg.ReplaceAllString(murl, ""), ""),
					imgHeightTplReg.ReplaceAllString(imgWidthTplReg.ReplaceAllString(murl, "&w=1000"), "&h=1000"),
					imgHeightTplReg.ReplaceAllString(imgWidthTplReg.ReplaceAllString(murl, "&w=700"), "&h=700"),
					imgHeightTplReg.ReplaceAllString(imgWidthTplReg.ReplaceAllString(murl, "&w=600"), "&h=600"),
					"",
					j == 0))
			}
		} else {
			seli := doc.Find(`#thumbSlides .thumbSlide .pdp-image`)
			for j := range seli.Nodes {
				node := seli.Eq(j)
				murl := node.AttrOr("data-large", node.AttrOr("data-thumb", ""))
				if murl == "" {
					continue
				}
				medias = append(medias, pbMedia.NewImageMedia(
					strconv.Format(j),
					imgHeightTplReg.ReplaceAllString(imgWidthTplReg.ReplaceAllString(murl, ""), ""),
					imgHeightTplReg.ReplaceAllString(imgWidthTplReg.ReplaceAllString(murl, "&w=1000"), "&h=1000"),
					imgHeightTplReg.ReplaceAllString(imgWidthTplReg.ReplaceAllString(murl, "&w=700"), "&h=700"),
					imgHeightTplReg.ReplaceAllString(imgWidthTplReg.ReplaceAllString(murl, "&w=600"), "&h=600"),
					"",
					j == 0))
			}
		}

		colorSpec := pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecColor,
			Id:    style,
			Name:  node.Find("a .color-image>img").AttrOr("alt", style),
			Value: style,
			Icon:  node.Find("a .color-image>img").AttrOr("data-src", ""),
		}

		priceSelNode := doc.Find(fmt.Sprintf(`#prices_%s .productPrice`, style))
		orgPrice, _ := strconv.ParsePrice(priceSelNode.Find(`.wasPrice`).Text())
		currentPrice, _ := strconv.ParsePrice(priceSelNode.Find(`.nowPrice`).Text())
		discount := float64(0)
		if currentPrice == 0 {
			currentPrice, _ = strconv.ParsePrice(priceSelNode.Find(`.fullPrice`).Text())
		}
		if orgPrice == 0 {
			orgPrice = currentPrice
		}
		if orgPrice != currentPrice {
			discount = math.Round((orgPrice - currentPrice) / orgPrice * 100)
		}

		// Note: Color variation is available on product list page therefor not considering multiple color of a product
		sizeSel := doc.Find(fmt.Sprintf(`#sizes_%s .sizeOptions`, style))
		for i := range sizeSel.Nodes {
			snode := sizeSel.Eq(i)

			skuId, _ := base64.RawStdEncoding.DecodeString(snode.AttrOr(`data-sku`, ""))
			if len(skuId) == 0 {
				c.logger.Errorf("invalid sku id %s", style)
				continue
			}
			sizeName := strings.TrimSpace(snode.Text())

			sku := pbItem.Sku{
				SourceId: string(skuId),
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(currentPrice * 100),
					Msrp:     int32(orgPrice * 100),
					Discount: int32(discount),
				},
				Medias: medias,
				Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}
			if !strings.Contains(snode.AttrOr("class", ""), "disabled") {
				sku.Stock.StockStatus = pbItem.Stock_InStock
			}

			sku.Specs = append(sku.Specs, &colorSpec)
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    sizeName,
				Name:  sizeName,
				Value: sizeName,
			})
			item.SkuItems = append(item.SkuItems, &sku)
		}
	}
	if err := yield(ctx, &item); err != nil {
		return err
	}
	return nil
}

func (c *_Crawler) NewTestRequest(ctx context.Context) []*http.Request {
	var reqs []*http.Request
	for _, u := range []string{
		// "https://www.finishline.com/store/only-sale-items/shoes/_/N-1g2z5sjZ1nc1fo0?icid=LP_sale_shoes50_PDCT",
		"https://www.finishline.com/store/women/shoes/running/_/N-nat3jh?mnid=women_shoes_running",
		// "https://www.finishline.com/store/product/womens-nike-air-max-270-casual-shoes/prod2770847?styleId=AH6789&colorId=001",
		// "https://www.finishline.com/store/product/mens-nike-challenger-og-casual-shoes/prod2820864?styleId=CW7645&colorId=003",
		//"https://www.finishline.com/store/product/womens-puma-future-rider-play-on-casual-shoes/prod2795926?styleId=38182501&colorId=100",
		// "https://www.finishline.com/store/product/big-kids-nike-air-force-1-low-casual-shoes/prod796065?styleId=314192&colorId=117",
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
				i.URL.Host = "www.finishline.com"
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
