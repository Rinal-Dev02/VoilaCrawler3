package main

import (
	"bytes"
	"context"
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
		categoryPathMatcher:    regexp.MustCompile(`^/([/A-Za-z0-9_-]+)$`),
		categoryAPIPathMatcher: regexp.MustCompile(`^/category-search-ajax$`),
		productPathMatcher:     regexp.MustCompile(`^/([/A-Za-z0-9_-]+)\d+$`),
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
	return []string{"*.aldoshoes.in"}
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

	rootUrl := "https://www.aldoshoes.in/index.php?route=extension/module/megamenu/getMenu&_=1629952316290"
	req, _ := http.NewRequest(http.MethodGet, rootUrl, nil)
	opts := c.CrawlOptions(req.URL)
	req.Header.Set("accept-language", "en-GB,en-US;q=0.9,en")
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("referer", "https://www.aldoshoes.in/")

	for _, c := range opts.MustCookies {
		req.AddCookie(c)
	}
	for k := range opts.MustHeader {
		req.Header.Set(k, opts.MustHeader.Get(k))
	}
	resp, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
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

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	sel := dom.Find(`.megamenu-pattern > div > ul > li`)
	fmt.Println(len(sel.Nodes))

	for a := range sel.Nodes {
		node := sel.Eq(a)

		catname := strings.TrimSpace(node.Find(`a`).First().Find(`span > strong`).Text())
		if catname == "" {
			continue
		}
		fmt.Println(`CategoryName >>`, catname)
		//nctx := context.WithValue(ctx, "Category", cateName)

		sublvl1div := node.Find(`.sub-menu > div > div > div > div > div > .withchild`)
		for b := range sublvl1div.Nodes {
			sublvl1 := sublvl1div.Eq(b)
			sublvl1name := strings.TrimSpace(sublvl1.Find(`h4`).First().Text())
			if sublvl1name == "" {
				continue
			}
			fmt.Println(`SubLevel1 >>`, catname+`>>`+sublvl1name)

			sublvl2 := sublvl1.Find(`ul > li`)
			for c := range sublvl2.Nodes {
				selsublvl3 := sublvl2.Eq(c)
				sublvl2name := strings.TrimSpace(selsublvl3.Find(`a`).First().Text())
				if sublvl2name == "" {
					continue
				}
				fmt.Println(`SubLevel2 `, sublvl1name+`>>`+sublvl2name)

				// nnnctx := context.WithValue(nnctx, "SubCategory", sublvl2name)
				// req, _ := http.NewRequest(http.MethodGet, href, nil)
				// if err := yield(nnnctx, req); err != nil {
				// return err
			}
		}
	}
	return nil
}

var productsTotalCountReg = regexp.MustCompile(`var\s*total_pages\s*=\s*'\d+';`)

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

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	lastIndex := nextIndex(ctx)

	ioutil.WriteFile("C:\\NewGIT_SVN\\Project_VoilaCrawler\\VoilaCrawler\\Output"+strconv.Format(lastIndex)+".html", respBody, 0644)

	sel := dom.Find(`.product-name`)
	for i := range sel.Nodes {
		fmt.Println(lastIndex)
		node := sel.Eq(i)
		href := node.Find(`a`).AttrOr("href", "")
		if href == "" {
			continue
		}

		// req, err := http.NewRequest(http.MethodGet, href, nil)
		// if err != nil {
		// 	c.logger.Errorf("load http request of url %s failed, error=%s", href, err)
		// 	return err
		// }

		// nctx := context.WithValue(ctx, "item.index", lastIndex)
		lastIndex += 1
		// if err := yield(nctx, req); err != nil {
		// 	return err
		// }
	}

	matched := productsTotalCountReg.FindSubmatch(respBody)
	if len(matched) > 1 {
		fmt.Println(`matched `, string(matched[1]))
	}

	// get current page number
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}

	if len(sel.Nodes) < 27 {
		return nil
	}

	// set pagination
	u := *resp.Request.URL
	vals := u.Query()
	vals.Set("page", strconv.Format(page+1))
	vals.Set("is_ajax", "0")
	u.RawQuery = vals.Encode()

	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
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

	brand := doc.Find(`#logo`).Find(`img`).AttrOr("title", "")
	if brand == "" {
		brand = "ALDO"
	}

	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           "",
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
		},
		BrandName: brand,
		Title:     doc.Find(`.product-title`).Text(),
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
	}

	// desc
	description := htmlTrimRegp.ReplaceAllString(doc.Find(`.custom-description`).Text(), " ")
	item.Description = string(TrimSpaceNewlineInString([]byte(description)))

	// if strings.Contains(doc.Find(`.availability product-availability`).AttrOr(`data-available`, ``), "true") {
	// 	item.Stock.StockStatus = pbItem.Stock_InStock
	// }

	currentPrice, _ := strconv.ParsePrice(doc.Find(`li[class="price-new"]`).Find(`h2`).Text())
	msrp, _ := strconv.ParsePrice(doc.Find(`li[class="price-old"]`).Find(`span`).Text())

	if msrp == 0 {
		msrp = currentPrice
	}
	discount := 0
	if msrp > currentPrice {
		discount = (int)(((msrp - currentPrice) / msrp) * 100)
	}

	//images
	sel := doc.Find(`.zoom-thumbnails > div`)
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
	sel = doc.Find(`.grop-product-product > a`)
	for i := range sel.Nodes {
		node := sel.Eq(i)

		if strings.Contains(node.AttrOr(`active`, ``), ``) {
			// cid = node.AttrOr(`data-variationgroupid`, "")
			icon := strings.ReplaceAll(strings.ReplaceAll(node.Find(`img`).AttrOr(`src`, ""), "background-image: url(", ""), ")", "")
			// colorName = node.AttrOr(`colour-name`, "")
			colorSelected = &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    cid,
				Name:  colorName,
				Value: colorName,
				Icon:  icon,
			}
		}
	}

	sel = doc.Find(`#product`).Find(`.radio-type-button2`)
	for i := range sel.Nodes {
		node := sel.Eq(i)

		sid := (node.Find(`label`).First().Text())
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

		// if strings.Contains(node.Find(`input`).First().AttrOr("class", ""), "js-show-notify-modal") {
		// 	sku.Stock.StockStatus = pbItem.Stock_OutOfStock
		// }

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
		//"https://www.aldoshoes.in/",
		//"https://www.aldoshoes.in/aldo-accessories-women/aldo-accessories-women",
		//"https://www.aldoshoes.in/aldo-women-footware-pumps-61671/1184031",
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
