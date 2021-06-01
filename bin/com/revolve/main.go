package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/voiladev/VoilaCrawler/pkg/cli"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	pbMedia "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	pbProxy "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
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
		productPathMatcher:  regexp.MustCompile(`^(((.*)(/dp/)(.*))|(/r/DisplayProduct.jsp))$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "49fef71d0fb271666a3095da56533b5e"
}

// Version
func (c *_Crawler) Version() int32 {
	return 1
}

// CrawlOptions
func (c *_Crawler) CrawlOptions(u *url.URL) *crawler.CrawlOptions {
	options := crawler.NewCrawlOptions()
	options.EnableHeadless = false
	options.Reliability = pbProxy.ProxyReliability_ReliabilityMedium
	options.MustCookies = append(options.MustCookies,
		&http.Cookie{Name: "currencyOverride", Value: "USD"},
		&http.Cookie{Name: "currency", Value: "USD"},
		&http.Cookie{Name: "userLanguagePref", Value: "en"},
		&http.Cookie{Name: "requestBrowserIdMapping", Value: "0"},
		&http.Cookie{Name: "originalsource", Value: "0"},
	)
	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.revolve.com"}
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
		u.Host = "www.revolve.com"
	}
	if c.productPathMatcher.MatchString(u.Path) {
		u.RawQuery = ""
		return u.String(), nil
	}
	return rawurl, nil
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
	return crawler.ErrUnsupportedPath
}

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

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	lastIndex := nextIndex(ctx)

	sel := doc.Find(`#plp-prod-list .js-plp-container`)
	for i := range sel.Nodes {
		node := sel.Eq(i)
		if href, _ := node.Find(".js-plp-pdp-link").Attr("href"); href != "" {
			//c.logger.Debugf("yield %w%s", lastIndex, href)
			//fmt.Println(href)
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

	lazyLoadUrl := doc.Find(`#plp-prod-list`).AttrOr("data-lazy-load-url", "")
	if lazyLoadUrl != "" {
		ru := resp.Request.URL
		lazyLoadUrl = ru.Scheme + "://" + ru.Host + lazyLoadUrl + "&_=" + strconv.Format(time.Now().UnixNano()/1000000)

		req, err := http.NewRequest(http.MethodGet, lazyLoadUrl, nil)
		req.Header.Add(`Referer`, resp.Request.URL.String())

		// opts := c.CrawlOptions(req.URL)

		resp, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
			EnableProxy:       true,
			EnableHeadless:    true,
			EnableSessionInit: true,
			DisableCookieJar:  false,
			Reliability:       1,
		})

		if err != nil {
			c.logger.Error(err)
			return err
		}
		respBody, err = ioutil.ReadAll(resp.Body)

		defer resp.Body.Close()

		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
		if err != nil {
			c.logger.Error(err)
			return err
		}

		sel := doc.Find(`.js-plp-container`)

		for i := range sel.Nodes {
			node := sel.Eq(i)
			if href, _ := node.Find(".js-plp-pdp-link").Attr("href"); href != "" {
				//c.logger.Debugf("yield %w%s", lastIndex, href)
				//fmt.Println(href)
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
		//}
	}

	// get current page number
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("pageNum"))
	if page == 0 {
		page = 1
	}

	nextPageNum := doc.Find(`#tr-pagination__controls--next`).AttrOr("href", "")
	if strings.HasPrefix(nextPageNum, "javascript:setPageNumber(") {
		pageNum, _ := strconv.ParseInt(strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(nextPageNum, "javascript:setPageNumber("), ")")))
		if pageNum > page {
			// set pagination
			u := *resp.Request.URL
			vals := u.Query()
			vals.Set("pageNum", strconv.Format(pageNum))
			u.RawQuery = vals.Encode()

			req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
			// update the index of last page
			nctx := context.WithValue(ctx, "item.index", lastIndex)

			return yield(nctx, req)
		}
	}
	return nil
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
		// return err //check
	}

	canUrl := doc.Find(`link[rel="canonical"]`).AttrOr("href", "")
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           p.Sku,
			CrawlUrl:     resp.Request.URL.String(),
			GroupId:      strings.SplitN(p.Sku, "-", 2)[0],
			CanonicalUrl: canUrl,
		},
		Title:       p.Name,
		Description: strings.TrimSpace(doc.Find(`.product-details__description-content`).Text()),
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

	var medias []*pbMedia.Media
	sel = doc.Find(`#js-primary-slideshow__pager .js-primary-slideshow__pager-thumb`)
	for j := range sel.Nodes {
		node := sel.Eq(j)

		mediumUrl := node.AttrOr("data-image", "")
		label := "?"
		if strings.Contains(mediumUrl, "?") {
			label = "&"
		}

		medias = append(medias, pbMedia.NewImageMedia(
			strconv.Format(j),
			mediumUrl,
			mediumUrl+label+"w=1000&&h=1000",
			mediumUrl+label+"w=700&&h=700",
			mediumUrl+label+"w=600&&h=600",
			"", j == 0))
	}

	colorName := strings.TrimSpace(doc.Find(`.product-sections .selectedColor`).Text())

	sel = doc.Find(".product-sizes .js-size-option")
	if len(sel.Nodes) > 0 {
		for i := range sel.Nodes {
			snode := sel.Eq(i)

			sizeval, _ := snode.Attr(`data-size`)
			qty, _ := snode.Attr(`data-qty`)
			quantity, _ := strconv.ParseInt(qty)

			currentPrice, _ := snode.Attr(`data-price`)
			cp, _ := strconv.ParseFloat(currentPrice)

			msrp, _ := snode.Attr(`data-regular-price`)
			mp, _ := strconv.ParseFloat(msrp)

			if cp == 0 {
				cp = mp
			}

			discount := 0.0
			if mp > 0.0 {
				discount = math.Ceil(((mp - cp) / mp) * 100)
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
				Medias: medias,
				Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}

			if quantity > 0 {
				sku.Stock.StockStatus = pbItem.Stock_InStock
				sku.Stock.StockCount = int32(quantity)
			}

			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    colorName,
				Name:  colorName,
				Value: colorName,
			})

			// size
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    sizeval,
				Name:  sizeval,
				Value: sizeval,
			})
			item.SkuItems = append(item.SkuItems, &sku)
		}
	} else if colorName != "" {
		currentPrice, _ := strconv.ParseFloat(doc.Find(`meta[property="wanelo:product:price"]`).AttrOr(`content`, ``))
		msrp, _ := strconv.ParseFloat(doc.Find(`meta[property="wanelo:product:price"]`).AttrOr(`content`, ``))

		discount := 0.0
		if msrp > 0.0 {
			discount = math.Ceil(((msrp - currentPrice) / msrp) * 100)
		} else {
			msrp = currentPrice
		}

		sku := pbItem.Sku{
			SourceId: p.Sku,
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(currentPrice * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},
			Medias: medias,
			Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}

		availability := doc.Find(`meta[property="wanelo:product:availability"]`).AttrOr(`content`, ``)
		if availability != "OutOfStock" {
			sku.Stock.StockStatus = pbItem.Stock_InStock
		}
		//sku.Stock.StockCount = int32(quantity)

		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecColor,
			Id:    colorName,
			Name:  colorName,
			Value: colorName,
		})

		item.SkuItems = append(item.SkuItems, &sku)
	}

	if len(item.SkuItems) > 0 {
		if err := yield(ctx, &item); err != nil {
			return err
		}
	} else {
		return errors.New("no invalud sku spec found")
	}

	// found other color
	sel = doc.Find(`#product-swatches>li`)
	for i := range sel.Nodes {
		node := sel.Eq(i)
		color := strings.TrimSpace(node.AttrOr("data-color-name", ""))
		c.logger.Debugf("found color %s %t", color, color == colorName)
		if color == "" || color == colorName {
			continue
		}
		u := node.AttrOr("data-swatch-url", "")
		if u == "" {
			continue
		}
		req, err := http.NewRequest(http.MethodGet, u, nil)
		if err != nil {
			c.logger.Error(err)
			continue
		}
		if err := yield(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

func (c *_Crawler) NewTestRequest(ctx context.Context) []*http.Request {
	var reqs []*http.Request
	for _, u := range []string{
		"https://www.revolve.com/lingerie-sleepwear-underwear/br/aadd43/?navsrc=left",
		//"https://www.revolve.com/its-now-cool-contour-crop-bikini-top/dp/ITSR-WX19/?d=Womens&page=31&lc=67&itrownum=119&itcurrpage=31&itview=05",
		//"https://www.revolve.com/michael-costello-x-revolve-electra-dress/dp/MELR-WD347/?d=Womens&page=28&lc=75&itrownum=34&itcurrpage=28&itview=05",
		//"https://www.revolve.com/denim/br/2664ce/?navsrc=left",
		// "https://www.revolve.com/skirts/br/8b6a66/?navsrc=subclothing",
		// "https://www.revolve.com/haight-panneaux-skirt/dp/HGHT-WQ2/?d=Womens&page=1&lc=9&itrownum=3&itcurrpage=1&itview=05",
		// "https://www.revolve.com/nphilanthropy-scarlett-leather-jogger-in-camel/dp/PHIR-WP63/?d=Womens&sectionURL=%2Fpants%2Fbr%2F44d522%2F%3F%26s%3Dc%26c%3DPants%26navsrc%3Dsubclothing&code=PHIR-WP63",
		// "https://www.revolve.com/daydreamer-rolling-stones-classic-tongue-tee/dp/DDRE-WS437/?d=Womens&page=1&lc=1&itrownum=1&itcurrpage=1&itview=05",
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
	//os.Setenv("VOILA_PROXY_URL", "http://52.207.171.114:30216")
	cli.NewApp(New).Run(os.Args)
}
