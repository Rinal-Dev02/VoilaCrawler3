package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math"
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
	httpClient          http.Client
	categoryPathMatcher *regexp.Regexp
	productPathMatcher  *regexp.Regexp
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
		categoryPathMatcher: regexp.MustCompile(`^/en([/A-Za-z0-9_-]+)$`),
		productPathMatcher:  regexp.MustCompile(`^/en([/A-Za-z0-9_-]+).html$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "622cff28a7e340d3afb1abef9274b33e"
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
		Reliability:       proxy.ProxyReliability_ReliabilityDefault,
	}
	//opts.MustHeader.Add("cookie", "__cf_bm=47aac5aba38023fe43aaac32ad36c3e2a2138de9-1626322522-1800-Ac1IkZ+PEazbK0+Swk4mn6u/eDEzMzjmTsG1WZi5fxpyX0OT5zVSdrTeXJAduFjHrafwdDkQjooNQRAlrmflCss=; _gid=GA1.2.851363828.1626322555; dwac_d64768f7aa8fa81249fd00e9b3=aVw5bzgxTK4ibpANnVhJBpwJTz6BP68plUU=|dw-only|||USD|false|America/Los_Angeles|true; cqcid=abOdJN4Ol8btSCur7fobkoG4Ob; cquid=||; sid=aVw5bzgxTK4ibpANnVhJBpwJTz6BP68plUU; GlobalE_Data={\"countryISO\":\"US\",\"cultureCode\":\"en-US\",\"currencyCode\":\"USD\",\"apiVersion\":\"2.1.4\"}; dwanonymous_63e67731e3cfae956fe285e9880bda4e=abOdJN4Ol8btSCur7fobkoG4Ob; __cq_dnt=0; dw_dnt=0; dwsid=ZDUHetxLX8-LHDWvWv-RjJFgcStlRbxIPDeauqsnGk0mlgpOGCclNuHpcY8mm-KD6uJ0FCymioe03Mvs8Y_HUA==; GlobalE_CT_Data={\"CUID\":\"186140072.678876305.705\",\"CHKCUID\":null}; GlobalE_Full_Redirect=false; __cq_uuid=abOdJN4Ol8btSCur7fobkoG4Ob; __cq_seg=0~0.00!1~0.00!2~0.00!3~0.00!4~0.00!5~0.00!6~0.00!7~0.00!8~0.00!9~0.00; _mibhv=anon-1626322564759-260434991_7217; _gcl_au=1.1.369013815.1626322569; _scid=8cf0e1a4-8f66-400c-89b3-7d21b03600af; _br_uid_2=uid=2085050983736:v=11.8:ts=1626322564244:hc=2; __cq_bc={\"bfkh-sandro-paris\":[{\"id\":\"2000441987\"}]}; _ga=GA1.2.214182795.1626318891; tfc-s={\"v\":\"tfc-fitrec-catalog=1&tfc-fitrec-product=1\"}; tfc-l={\"a\":{\"v\":\"b750fc43-69a0-4fa6-b7fc-7ec018611fc7\",\"e\":1626408973},\"u\":{\"v\":\"7b1fm0munnpoeanoq7g0tsj1ch\",\"e\":1689221769},\"s\":{\"v\":\"\",\"e\":1689221768},\"k\":{\"v\":\"7b1fm0munnpoeanoq7g0tsj1ch\",\"e\":1689221769},\"c\":{\"v\":\"adult\",\"e\":1689221768}}; QuantumMetricSessionID=e10df02f08114fa936ae623c6040fa9e; QuantumMetricUserID=8ba94cac168c8faa7726ba9c0666f5eb; _fbp=fb.1.1626322574847.1262081683; _dy_ses_load_seq=29795:1626322576173; _dy_csc_ses=t; _dy_c_exps=; _pin_unauth=dWlkPU1ESmxZalUxWmpndFl6WTBZeTAwTVRSa0xUa3hZV010TnpGbU5XRmxaVFZqTmpRMQ; TTSVID=9b18d566-24bc-45f1-9086-98e02bc31a57; _dycnst=dg; __attentive_id=c598a93546494fca8d90906a452d8e2f; __attentive_cco=1626322577861; __attentive_pv=1; __attentive_ss_referrer=\"https://www.sandro-paris.com/us/shop/catalog/category/plus/plus-size-new-arrivals-tops\"; _dyid=1242993661989336721; _dyfs=1626322577950; _dyjsession=233d79bcdc89e12bec55528b1b40e867; dy_fs_page=www.sandro-paris.com/us/2000441987.html?dwvar_2000441987_color=01; _dy_lu_ses=233d79bcdc89e12bec55528b1b40e867:1626322577952; _dycst=dk.w.c.ms.; _dy_geo=US.NA.US_DC.US_DC_Washington; _dy_df_geo=United States.District Of Columbia.Washington; _dy_toffset=0; _attn_=eyJ1Ijoie1widmFsXCI6XCJjNTk4YTkzNTQ2NDk0ZmNhOGQ5MDkwNmE0NTJkOGUyZlwiLFwiY29cIjoxNjI2MzIyNTc4MjAwLFwidW9cIjoxNjI2MzIyNTc4MjAwLFwibWFcIjoyMTkwMH0ifQ==; _sctr=1|1626287400000; __attentive_dv=1; _dy_soct=517924.977215.1626322576*482065.880923.1626322576*539966.1076297.1626322578*539967.1076350.1626322578*539978.1082806.1626322578*505255.938114.1626322583; __idcontext=eyJjb29raWVJRCI6IlQ0VDJQVFlTNkxIQkQ1VVFXWlpYUks2SDNUNU5QNldNN0JRSDJETEFIREFRPT09PSIsImRldmljZUlEIjoiVDRUMlBUWVNaMzNEUDdNRlpVWlczRVBNMkNaNVBZRkk0UkFTS05CUEhXSlE9PT09IiwiaXYiOiIzSlgyRFg2T0QyQlVZSU5XM05OVVRGRFJSVT09PT09PSIsInYiOjF9; OptanonAlertBoxClosed=2021-07-15T04:18:40.212Z; OptanonConsent=isGpcEnabled=0&datestamp=Thu+Jul+15+2021+09:48:40+GMT+0530+(India+Standard+Time)&version=6.20.0&isIABGlobal=false&hosts=&landingPath=NotLandingPage&groups=C0001:1,C0002:1,C0003:1,C0004:1&AwaitingReconsent=false; _uetsid=619ce810e52311eba21fa359e8a46f2e; _uetvid=619d52d0e52311eba53d1f8c2b8ae6b6; _ga_L1T9QR4S4K=GS1.1.1626322557.2.1.1626322725.55; _dc_gtm_UA-233210-1=1")
	// opts.MustCookies = append(opts.MustCookies,
	// 	&http.Cookie{Name: "GlobalE_Data", Value: `{"countryISO":"US","cultureCode":"en-US","currencyCode":"USD","apiVersion":"2.1.4"}`, Path: "/"},
	// 	//&http.Cookie{Name: "_dy_geo", Value: "US.NA.US_DC.US_DC_Washington", Path: "/"},
	// )
	return opts
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.sandro-paris.com"}
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
	if p == "/en/womens" || p == "/en/mens" {
		return c.parseCategories(ctx, resp, yield)
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

	sel := dom.Find(`.menu-category>li`)

	for i := range sel.Nodes {
		node := sel.Eq(i)

		cateName := strings.TrimSpace(node.Find(`a`).First().Text())
		if cateName == "" {
			continue
		}

		nctx := context.WithValue(ctx, "Category", cateName)

		subSel := node.Find(`.menu-wrapper.level-2>div`)

		subcat2 := ""

		for k := range subSel.Nodes {
			subNode := subSel.Eq(k)

			if strings.TrimSpace(subNode.Find(`p`).First().Text()) != "other categories" {
				subcat2 = strings.TrimSpace(subNode.Find(`p`).First().Text())
			}

			nnctx := context.WithValue(nctx, "SubCategory", subcat2)

			subNode3 := subNode.Find(`li`)
			for j := range subNode3.Nodes {

				subNode1 := subNode3.Eq(j)
				subcat3 := strings.TrimSpace(subNode1.Find(`a`).Text())

				href := subNode1.Find(`a`).First().AttrOr("href", "")
				if href == "" || subcat3 == "" {
					continue
				}

				u, err := url.Parse(href)
				if err != nil {
					c.logger.Error("parse url %s failed", href)
					continue
				}

				if c.categoryPathMatcher.MatchString(u.Path) {
					nnnctx := context.WithValue(nnctx, "SubCategory2", subcat3)
					req, _ := http.NewRequest(http.MethodGet, href, nil)
					if err := yield(nnnctx, req); err != nil {
						return err
					}
				}
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
		c.logger.Error(err)
		return err
	}

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	lastIndex := nextIndex(ctx)
	sel := dom.Find(`.product-name`)
	for i := range sel.Nodes {
		node := sel.Eq(i)

		href := node.Find("a").AttrOr("href", "")
		if href == "" {
			c.logger.Warnf("no href found")
			continue
		}

		req, err := http.NewRequest(http.MethodGet, href, nil)
		if err != nil {
			c.logger.Errorf("create request with url %s failed", href)
			continue
		}
		nctx := context.WithValue(ctx, "item.index", lastIndex)
		lastIndex += 1
		if err := yield(nctx, req); err != nil {
			return err
		}
	}

	totalProducts, _ := strconv.ParsePrice(dom.Find(`.nbProducts`).Text())

	if lastIndex >= (int)(totalProducts) {
		return nil
	}

	nextUrl := dom.Find(`link[rel="next"]`).AttrOr("href", "")
	if nextUrl == "" {
		return nil
	}

	req, err := http.NewRequest(http.MethodGet, nextUrl, nil)
	if err != nil {
		return c.logger.Errorf("create request with url %s failed", nextUrl).ToError()
	}
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
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

// used to trim html labels in description
var htmlTrimRegp = regexp.MustCompile(`</?[^>]+>`)

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
	brand := doc.Find(`meta[itemprop="brand"]`).Text()
	if brand == "" {
		brand = "Sandro"
	}
	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           doc.Find(`#pid`).AttrOr("value", ""),
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
			GroupId:      doc.Find(`meta[property="product:age_group"]`).AttrOr(`content`, ``),
		},
		BrandName: brand,
		Title:     doc.Find(`.prod-title`).Text(),
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
	}

	if strings.Contains(doc.Find(`meta[property="product:availability"]`).AttrOr(`content`, ``), "instock") {
		item.Stock.StockStatus = pbItem.Stock_InStock
	}

	description := doc.Find(`.shortDescription`).Text()
	item.Description = string(TrimSpaceNewlineInString([]byte(description)))

	sel := doc.Find(`.breadcrumb>ol>li`)
	for i := range sel.Nodes {
		if i == len(sel.Nodes)-1 {
			continue
		}
		node := sel.Eq(i)
		breadcrumb := strings.TrimSpace(node.Text())

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

	msrp, _ := strconv.ParsePrice(doc.Find(`.product-price`).Find(`.price-standard`).Text())
	originalPrice, _ := strconv.ParsePrice(doc.Find(`.product-price`).Find(`.price-sales`).Text())
	discount := 0.0
	if msrp == 0 {
		msrp = originalPrice
	}
	if msrp > originalPrice {
		discount = math.Ceil((msrp - originalPrice) / msrp * 100)
	}

	//more_images
	sel = doc.Find(`.image-container.image-zoom`)
	for j := range sel.Nodes {
		node := sel.Eq(j)
		imgurl := strings.Split(node.Find(`img`).AttrOr(`data-src`, ``), "?")[0]

		item.Medias = append(item.Medias, pbMedia.NewImageMedia(
			strconv.Format(j),
			imgurl,
			imgurl+"?sw=1000",
			imgurl+"?sw=800",
			imgurl+"?sw=500",
			"", j == 0))
	}

	var colorSelected *pbItem.SkuSpecOption
	cid := strings.Split(doc.Find(`meta[property="product:product_link"]`).AttrOr(`content`, ``), "=")[1]
	selectedColor := string(TrimSpaceNewlineInString([]byte(doc.Find(`.selectedColor`).First().Text())))

	colorSelected = &pbItem.SkuSpecOption{
		Type:  pbItem.SkuSpecType_SkuSpecColor,
		Id:    cid,
		Name:  selectedColor,
		Value: selectedColor,
		Icon:  "https://us.sandro-paris.com" + strings.TrimSpace(doc.Find(`.selectedColor`).Find(`img`).AttrOr(`src`, ``)),
	}

	//Size swatches size
	sel1 := doc.Find(`.swatches.size>li`)
	for i := range sel1.Nodes {

		node := sel1.Eq(i)
		Size := strings.TrimSpace(node.Find(`.defaultSize`).First().Text())

		if Size == "" {
			continue
		}
		sid := strings.Split(node.Find(`a`).AttrOr("data-variationparameter", ""), "=")[1]

		sku := pbItem.Sku{
			SourceId: fmt.Sprintf("%s-%s", cid, sid),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(originalPrice),
				Msrp:     int32(msrp),
				Discount: int32(discount),
			},
			Stock: &pbItem.Stock{StockStatus: pbItem.Stock_InStock},
		}

		if strings.Contains(node.Find(`a`).AttrOr("class", ""), "unavailable") {
			sku.Stock.StockStatus = pbItem.Stock_OutOfStock
		}

		if colorSelected != nil {
			sku.Specs = append(sku.Specs, colorSelected)
		}

		// size
		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Id:    sid,
			Name:  strings.TrimSpace(node.Find(`.defaultSize`).Text()),
			Value: strings.TrimSpace(node.Find(`.defaultSize`).Text()),
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
		//"https://us.sandro-paris.com/en/womens/",
		//"https://us.sandro-paris.com/en/mens/clothing/sweatshirts/",
		//"https://us.sandro-paris.com/en/mens/clothing/t-shirts-and-polos/linen-t-shirt/SHPTS00222.html",
		//"https://us.sandro-paris.com/en/womens/clothing/dresses/",
		//"https://us.sandro-paris.com/en/womens/clothing/dresses/dress-with-tailored-collar/SFPRO01694.html?dwvar_SFPRO01694_color=55",
		"https://us.sandro-paris.com/en/mens/clothing/shirts/flowing-shirt-with-stripe-print/SHPCM00393.html?dwvar_SHPCM00393_color=C102#start=1",
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
	cli.NewApp(&_Crawler{}).Run(os.Args)
}
