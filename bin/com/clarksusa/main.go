package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/voiladev/VoilaCrawler/pkg/cli"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	pbMedia "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
)

// _Crawler defined the crawler struct/class for which is not necessory to be exportable
type _Crawler struct {
	crawler.MustImplementCrawler

	// httpClient is the object of an http client
	httpClient             http.Client
	categoryPathMatcher    *regexp.Regexp
	categoryAPIPathMatcher *regexp.Regexp
	productPathMatcher     *regexp.Regexp
	// logger is the log tool
	logger glog.Log
}

// New returns an object of interface crawler.Crawler.
// this is the entry of the spider plugin. the plugin manager will call this func to init the plugin.
// view pkg/crawler/spec.go to know more about the interface `Crawler`
func (_ *_Crawler) New(_ *cli.Context, client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient: client,
		// this regular used to match category page url path
		categoryPathMatcher:    regexp.MustCompile(`^/([/A-Za-z0-9_-]+)/c([/A-Za-z0-9_-]+)$`),
		categoryAPIPathMatcher: regexp.MustCompile(`^/category-search-ajax$`),
		productPathMatcher:     regexp.MustCompile(`^/c([/A-Za-z0-9_-]+)/p/([/A-Za-z0-9_-]+)`),
		logger:                 logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "1a802ce5da394208b6feeac90dacd332"
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
		EnableSessionInit: false,
		Reliability:       proxy.ProxyReliability_ReliabilityDefault,
	}

	return opts
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.clarksusa.com"}
}

// CanonicalUrl
func (c *_Crawler) CanonicalUrl(rawurl string) (string, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return "", err
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

	p := strings.TrimSuffix(resp.Request.URL.Path, "/")
	if p == "" {
		return c.parseCategories(ctx, resp, yield)
	}

	if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		fmt.Println(`productPathMatcher`)
		return c.parseProduct(ctx, resp, yield)
	} else if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) || c.categoryAPIPathMatcher.MatchString(resp.Request.URL.Path) {
		fmt.Println(`categoryPathMatcher`)
		return c.parseCategoryProducts(ctx, resp, yield)
	}

	return fmt.Errorf("unsupported url %s", resp.Request.URL.String())
}

// nextIndex used to get the index from the shared data.
// item.index is a const key for item index.
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")))
}

func (c *_Crawler) parseCategories(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	sel := dom.Find(`.new-header__main-navigation-list > li`)
	fmt.Println(len(sel.Nodes))

	for a := range sel.Nodes {
		node := sel.Eq(a)
		cateName := strings.TrimSpace(node.Find(`button`).First().Text())
		attrName := node.Find(`button`).AttrOr("data-flyout", "")
		if cateName == "" {
			continue
		}
		fmt.Println(`CateName >>`, cateName)
		//nctx := context.WithValue(ctx, "Category", cateName)

		attrM := "#" + attrName
		subdiv := dom.Find(attrM)
		test := subdiv.Find(`.new-header__flyout-menu-list > li`)
		for b := range test.Nodes {
			sublvl2 := test.Eq(b)
			sublvl2name := strings.TrimSpace(sublvl2.Find(`h2`).First().Text())
			fmt.Println(`sublvl2 >>`, cateName+`>>`+sublvl2name)

			selsublvl3 := sublvl2.Find(`ul > li`)
			for c := range selsublvl3.Nodes {
				sublvl3 := selsublvl3.Eq(c)
				sublvl3name := strings.TrimSpace(sublvl3.Find(`a`).First().Text())
				fmt.Println(`sublvl3 `, sublvl2name+`>>`+sublvl3name)

				// nnnctx := context.WithValue(nnctx, "SubCategory", sublvl3name)
				// req, _ := http.NewRequest(http.MethodGet, href, nil)
				// if err := yield(nnnctx, req); err != nil {
				// return err
			}
		}
	}
	return nil
}

// parseCategoryProducts parse api url from web page url
func (c *_Crawler) parseCategoryProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.logger.Debug(err)
		return err
	}
	ioutil.WriteFile("C:\\NewGIT_SVN\\Project_VoilaCrawler\\VoilaCrawler\\Output.html", respBody, 0644)

	lastIndex := nextIndex(ctx)

	var viewData categoryStructure

	if !c.categoryAPIPathMatcher.MatchString(resp.Request.URL.Path) {

		s := strings.Split(resp.Request.URL.Path, "/c/")

		rootUrl := "https://www.clarksusa.com/category-search-ajax?categoryCode=" + s[len(s)-1]
		req, _ := http.NewRequest(http.MethodGet, rootUrl, nil)
		opts := c.CrawlOptions(req.URL)
		req.Header.Set("accept-language", "en-GB,en-US;q=0.9,en;q=0.8")
		req.Header.Set("accept", "application/json, text/plain, */*")
		req.Header.Set("referer", resp.Request.URL.String())

		for _, c := range opts.MustCookies {
			req.AddCookie(c)
		}
		for k := range opts.MustHeader {
			req.Header.Set(k, opts.MustHeader.Get(k))
		}
		resp, err = c.httpClient.DoWithOptions(ctx, req, http.Options{
			EnableProxy:       true,
			EnableHeadless:    false,
			EnableSessionInit: false,
			Reliability:       opts.Reliability,
		})
		if err != nil {
			c.logger.Error(err)
			//return nil, err
		}
		defer resp.Body.Close()

		respBody, err = ioutil.ReadAll(resp.Body)
		ioutil.WriteFile("C:\\NewGIT_SVN\\Project_VoilaCrawler\\VoilaCrawler\\Output_json.html", respBody, 0644)

		if err := json.NewDecoder(bytes.NewReader(respBody)).Decode(&viewData); err != nil {
			c.logger.Errorf("unmarshal cat detail data fialed, error=%s", err)
			//return nil, err
		}
	} else {

		if err := json.NewDecoder(bytes.NewReader(respBody)).Decode(&viewData); err != nil {
			c.logger.Errorf("unmarshal cat detail data fialed, error=%s", err)
			//return nil, err
		}
	}

	for _, items := range viewData.Products {
		if href := items.URL; href != "" {
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

	// Next page url not found
	if viewData.Pagination.NumberOfPages > 1 {
		fmt.Println(`implement pagination`)
		c.logger.Errorf("implement pagination=%s")
		return nil
	}
	return nil
}

type categoryStructure struct {
	Pagination struct {
		PageSize             int    `json:"pageSize"`
		Sort                 string `json:"sort"`
		CurrentPage          int    `json:"currentPage"`
		TotalNumberOfResults int    `json:"totalNumberOfResults"`
		NumberOfPages        int    `json:"numberOfPages"`
	} `json:"pagination"`
	Products []struct {
		URL string `json:"url"`
	} `json:"products"`
}

func TrimSpaceNewlineInString(s []byte) []byte {
	re := regexp.MustCompile(`\n`)
	resp := re.ReplaceAll(s, []byte(" "))
	resp = bytes.ReplaceAll(resp, []byte("\\n"), []byte(""))
	resp = bytes.ReplaceAll(resp, []byte("\r"), []byte(""))
	resp = bytes.ReplaceAll(resp, []byte("\t"), []byte(""))
	resp = bytes.ReplaceAll(resp, []byte("&lt;"), []byte("<"))
	resp = bytes.ReplaceAll(resp, []byte("&gt;"), []byte(">"))
	resp = bytes.ReplaceAll(resp, []byte("  "), []byte(""))
	return resp
}

var productsReviewExtractReg = regexp.MustCompile(`(?Ums)<script type="application/ld\+json">\s*({.*})\s*</script>`)

// used to trim html labels in description
var htmlTrimRegp = regexp.MustCompile(`</?[^>]+>`)

type parseProductResponse struct {
	Context     string `json:"@context"`
	Type        string `json:"@type"`
	Name        string `json:"name"`
	Image       string `json:"image"`
	Description string `json:"description"`
	Sku         string `json:"sku"`
	Offers      struct {
		Type          string `json:"@type"`
		Price         string `json:"price"`
		URL           string `json:"url"`
		PriceCurrency string `json:"priceCurrency"`
		Availability  string `json:"availability"`
	} `json:"offers"`
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

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	canUrl, _ := c.CanonicalUrl(doc.Find(`link[rel="canonical"]`).AttrOr("href", ""))
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}

	brand := doc.Find(`#ao-logo-img`).Text()
	if brand == "" {
		brand = "Alice and Olivia"
	}

	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           "",
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
			//GroupId:      doc.Find(`meta[property="product:age_group"]`).AttrOr(`content`, ``),
		},
		BrandName: brand,
		Title:     doc.Find(`#product-name-panel__h1`).Text(),
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
	}

	// desc
	description := htmlTrimRegp.ReplaceAllString(doc.Find(`.product-description`).Text(), " ")
	item.Description = string(TrimSpaceNewlineInString([]byte(description)))

	if strings.Contains(doc.Find(`.availability product-availability`).AttrOr(`data-available`, ``), "true") {
		item.Stock.StockStatus = pbItem.Stock_InStock
	}

	currentPrice, _ := strconv.ParsePrice(doc.Find(`.js-prev-price`).AttrOr("content", ""))
	msrp, _ := strconv.ParsePrice(doc.Find(`.js-current-price`).AttrOr("content", ""))

	if msrp == 0 {
		msrp = currentPrice
	}
	discount := 0
	if msrp > currentPrice {
		discount = (int)(((msrp - currentPrice) / msrp) * 100)
	}

	//images
	sel := doc.Find(`#carousel`).Find(`#indicatorList > li`)
	for j := range sel.Nodes {
		node := sel.Eq(j)
		imgurl := strings.Split(node.Find(`img`).AttrOr(`src`, ``), "?")[0]

		item.Medias = append(item.Medias, pbMedia.NewImageMedia(
			strconv.Format(j),
			imgurl,
			imgurl+"?sw=800&sh=800&q=80",
			imgurl+"?sw=500&sh=500&q=80",
			imgurl+"?sw=300&sh=300&q=80",
			"", j == 0))
	}

	// itemListElement
	sel = doc.Find(`.breadcrumb>li`)
	c.logger.Debugf("nodes %d", len(sel.Nodes))
	for i := range sel.Nodes {
		node := sel.Eq(i)
		breadcrumb := strings.TrimSpace(node.Find(`a`).First().Text())

		if i == 0 {
			item.Category = breadcrumb
		} else if i == 1 {
			item.SubCategory = breadcrumb
		} else if i == 2 {
			item.SubCategory2 = breadcrumb
		} else if i == 3 {
			item.SubCategory3 = breadcrumb
		} else if i == 4 {
			item.SubCategory4 = breadcrumb
		}
	}

	// Color
	cid := ""
	colorName := ""
	var colorSelected *pbItem.SkuSpecOption
	sel = doc.Find(`.color-swatch-selector > ul > li`)
	for i := range sel.Nodes {
		node := sel.Eq(i)

		if strings.Contains(node.Find(`.color-swatch-selector__color-options`).AttrOr(`class`, ``), `selected`) {
			cid = node.AttrOr(`data-variationgroupid`, "")
			icon := strings.ReplaceAll(strings.ReplaceAll(node.Find(`img`).AttrOr(`src`, ""), "background-image: url(", ""), ")", "")
			colorName = node.AttrOr(`colour-name`, "")
			colorSelected = &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    cid,
				Name:  colorName,
				Value: colorName,
				Icon:  icon,
			}
		}
	}

	sel = doc.Find(`.choose-size`).Find(`.box-selectors__wrapper > label`)
	for i := range sel.Nodes {
		node := sel.Eq(i)

		sid := (node.Find(`input`).First().AttrOr("data-prompt-value", ""))
		if sid == "" {
			continue
		}

		sku := pbItem.Sku{
			SourceId: fmt.Sprintf("%s-%s", cid, sid),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(currentPrice),
				Msrp:     int32(msrp),
				Discount: int32(discount),
			},
			Stock: &pbItem.Stock{StockStatus: pbItem.Stock_InStock},
		}

		if strings.Contains(node.Find(`input`).First().AttrOr("class", ""), "js-show-notify-modal") {
			sku.Stock.StockStatus = pbItem.Stock_OutOfStock
		}

		if colorSelected != nil {
			sku.Specs = append(sku.Specs, colorSelected)
		}

		// size
		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Id:    sid,
			Name:  sid,
			Value: sid,
		})

		item.SkuItems = append(item.SkuItems, &sku)
	}

	// yield item result
	if err = yield(ctx, &item); err != nil {
		c.logger.Errorf("yield sub request failed, error=%s", err)
		return err
	}

	return nil
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://www.clarksusa.com/",
		"https://www.clarksusa.com/Womens-Best-Sellers/c/us182",
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
	os.Setenv("VOILA_PROXY_URL", "http://52.207.171.114:30216")
	cli.NewApp(&_Crawler{}).Run(os.Args)
}
