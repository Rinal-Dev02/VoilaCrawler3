package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
		categoryPathMatcher: regexp.MustCompile(`^(/us/en([/A-Za-z0-9_-]+))|(/on/demandware.store([/A-Za-z0-9_-]+))$`),
		productPathMatcher:  regexp.MustCompile(`^/us/en([/A-Za-z0-9_-]+).html$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "d777aaa29d7e4d5183c6b3ac38d66cde"
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
	return []string{"*.acnestudios.com"}
}

// CanonicalUrl
func (c *_Crawler) CanonicalUrl(rawurl string) (string, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	if u.Host == "" {
		u.Host = "www.acnestudios.com"
	}
	if c.productPathMatcher.MatchString(u.Path) {
		u.RawQuery = ""
	}
	return u.String(), nil
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
	if p == "/us/en/home" {
		return crawler.ErrUnsupportedPath
	}

	if c.productPathMatcher.MatchString(p) {
		return c.parseProduct(ctx, resp, yield)
	} else if c.categoryPathMatcher.MatchString(p) {
		return c.parseCategoryProducts(ctx, resp, yield)
	}

	return crawler.ErrUnsupportedPath
}

// nextIndex used to get the index from the shared data.
// item.index is a const key for item index.
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")))
}

var productsExtractReg = regexp.MustCompile(`(?U)window.searchSuggestions\s*=\s*JSON.parse\('({.*})'\);`)

func (c *_Crawler) GetCategories(ctx context.Context) ([]*pbItem.Category, error) {
	req, _ := http.NewRequest(http.MethodGet, "https://www.acnestudios.com/us/en/home", nil)
	opts := c.CrawlOptions(req.URL)
	resp, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
		EnableProxy:       true,
		EnableHeadless:    opts.EnableHeadless,
		EnableSessionInit: opts.EnableSessionInit,
		Reliability:       opts.Reliability,
	})
	if err != nil {
		c.logger.Error(err)
		return nil, err
	}
	respBody, _ := resp.RawBody()
	//	dom, err := resp.Selector()
	if err != nil {
		c.logger.Error(err)
		return nil, err
	}

	matched := productsExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respBody)
		c.logger.Error("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}

	var viewData categoryStructure
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal category detail data fialed, error=%s", err)
	}

	var (
		cates   []*pbItem.Category
		cateMap = map[string]*pbItem.Category{}
	)
	if err := func(yield func(names []string, url string) error) error {

		for _, nodes := range viewData.SearchSuggestions {

			cateName := strings.ReplaceAll(nodes.Category, "shop-", "")

			if cateName == "" {
				continue
			}

			for _, rawcat := range nodes.SubCategories {

				subcat2 := rawcat.Section

				subcat3 := rawcat.Name

				href := rawcat.URL
				if href == "" || subcat3 == "" || strings.ToLower(subcat2) == "runway" {
					continue
				}

				canonicalhref, err := c.CanonicalUrl(href)
				if err != nil {
					c.logger.Error("parse url %s failed", href)
					continue
				}

				u, err := url.Parse(canonicalhref)
				if err != nil {
					c.logger.Error("parse url %s failed", href)
					continue
				}

				if c.categoryPathMatcher.MatchString(u.Path) {
					if err := yield([]string{cateName, subcat2, subcat3}, canonicalhref); err != nil {
						return err
					}
				}
			}
		}

		return nil
	}(func(names []string, url string) error {
		if len(names) == 0 {
			return errors.New("no valid category name found")
		}

		var (
			lastCate *pbItem.Category
			path     string
		)

		for i, name := range names {
			path = strings.Join([]string{path, name}, "-")

			name = strings.Title(strings.ToLower(name))
			if cate, _ := cateMap[path]; cate != nil {
				lastCate = cate
				continue
			} else {
				cate = &pbItem.Category{
					Name: name,
				}
				cateMap[path] = cate
				if lastCate != nil {
					lastCate.Children = append(lastCate.Children, cate)
				}
				lastCate = cate

				if i == 0 {
					cates = append(cates, cate)
				}
			}
		}
		lastCate.Url = url
		return nil
	}); err != nil {
		c.logger.Error(err)
		return nil, err
	}
	return cates, nil
}

type categoryStructure struct {
	SearchSuggestions []struct {
		Category      string `json:"category"`
		SubCategories []struct {
			Parent  []string `json:"parent"`
			Name    string   `json:"name"`
			Section string   `json:"section"`
			URL     string   `json:"url"`
		} `json:"subCategories"`
	} `json:"searchSuggestions"`
}

type parseCategoryPagination struct {
	InfiniteScrollObserver struct {
		URL string `json:"url"`
	} `json:"infiniteScrollObserver"`
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

	sel := dom.Find(`.tile__link`)
	if len(sel.Nodes) == 0 {
		return nil
	}

	for i := range sel.Nodes {
		node := sel.Eq(i)

		href := node.AttrOr("href", "")

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

	// get current page number
	pageStart, _ := strconv.ParseInt(resp.Request.URL.Query().Get("start"))
	size, _ := strconv.ParseInt(resp.Request.URL.Query().Get("sz"))
	if size == 0 {
		size = 12
	}
	pageStart = pageStart + size

	if len(sel.Nodes) < int(size) {
		return nil
	}

	cgid := dom.Find(`.page.page-search-show`).AttrOr("data-querystring", "")
	if cgid == "" {
		cgid = resp.RawUrl().Query().Get("cgid")
		if cgid == "" {
			return nil
		} else {
			cgid = "cgid=" + resp.RawUrl().Query().Get("cgid")
		}
	}

	nextUrl := "https://www.acnestudios.com/on/demandware.store/Sites-acne_us-Site/en_US/Search-UpdateGrid?" + cgid

	// set pagination
	u, _ := url.Parse(nextUrl)
	vals := u.Query()
	//vals.Set("start", "0")
	vals.Set("start", strconv.Format(pageStart))
	vals.Set("sz", "96")
	u.RawQuery = vals.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
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

var productID = regexp.MustCompile(`-[A-Z-0-9]+.html`)
var imgReg = regexp.MustCompile(`/\d+x/`)
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
	pid := doc.Find(`.product-detail`).AttrOr("data-pid", "")

	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           pid,
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
		},
		BrandName: "Acnestudios",
		Title:     strings.TrimSpace(doc.Find(`h1.cell--span-6---small-down `).Text()),
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
	}

	if strings.Contains(doc.Find(`meta[property="product:availability"]`).AttrOr(`content`, ``), "instock") {
		item.Stock.StockStatus = pbItem.Stock_InStock
	}

	description := doc.Find(`div[data-content-toggle-id="pdp-description"]`).Text()
	item.Description = string(TrimSpaceNewlineInString([]byte(description)))

	sel := doc.Find(`nav[class="breadcrumbs hide-for-small-down"]`).Find(`li`)
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

	originalPrice, _ := strconv.ParsePrice(doc.Find(`.pdp__price`).Find(`.sales`).Find(`.value`).AttrOr("content", ""))

	msrp := float64(0)

	if msrp == 0 {
		msrp = originalPrice
	}
	discount := int32(0)
	if msrp > originalPrice {
		discount = int32(((msrp - originalPrice) / msrp) * 100)
	}

	//data-attr="color"
	var colorSelected *pbItem.SkuSpecOption
	cid := ""
	colorName := ""

	sel = doc.Find(`div[data-attr="color"]`).Find(`a`)
	for i := range sel.Nodes {
		node := sel.Eq(i)

		if strings.Contains(node.AttrOr(`class`, ``), `selected`) || len(sel.Nodes) == 1 {
			colorName = strings.TrimSpace(node.AttrOr(`data-attr-label`, ""))
			cid = strings.TrimSpace(node.AttrOr(`data-attr-value`, ""))

			colorSelected = &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    cid,
				Name:  colorName,
				Value: cid,
				//Icon:
			}
		}
	}

	//more_images
	var media []*pbMedia.Media
	sel = doc.Find(`.pdp-gallery__image`)
	for j := range sel.Nodes {
		node := sel.Eq(j)
		imgurl := node.Find(`img`).AttrOr(`data-src`, ``)
		if imgurl == "" {
			continue
		} else if !strings.HasPrefix(imgurl, "http") {
			imgurl = "https://www.acnestudios.com" + imgurl
		}
		media = append(media, pbMedia.NewImageMedia(
			strconv.Format(j),
			imgurl,
			imgReg.ReplaceAllString(imgurl, `/1120x/`),
			imgReg.ReplaceAllString(imgurl, `/1120x/`),
			imgReg.ReplaceAllString(imgurl, `/560x/`),
			"", j == 0))
	}

	//Size swatches size
	sel1 := doc.Find(`.size-variations`).First().Find(`a`)

	for i := range sel1.Nodes {

		node := sel1.Eq(i)
		Size := strings.TrimSpace(node.AttrOr("data-value", ""))

		s := strings.Split(node.AttrOr("data-variation-url", ""), `_size=`)
		s = strings.Split(s[len(s)-1], `&`)
		skuID := pid + s[0]

		if Size == "" {
			continue
		}

		sku := pbItem.Sku{
			SourceId: skuID,
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(originalPrice * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},
			Medias: media,
			Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_InStock},
		}

		if strings.Contains(node.AttrOr("class", ""), "disabled") {
			sku.Stock.StockStatus = pbItem.Stock_OutOfStock
		}

		if colorSelected != nil {
			sku.Specs = append(sku.Specs, colorSelected)
		}

		// size
		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Id:    Size,
			Name:  Size,
			Value: Size,
		})

		// for _, spec := range sku.Specs {
		// 	sku.SourceId += fmt.Sprintf("-%s", spec.Id)
		// }

		item.SkuItems = append(item.SkuItems, &sku)
	}

	if len(sel1.Nodes) == 0 {
		sku := pbItem.Sku{
			SourceId: pid,
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(originalPrice * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},
			Medias: media,
			Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_InStock},
		}

		if strings.Contains(doc.AttrOr("class", ""), "disabled") {
			sku.Stock.StockStatus = pbItem.Stock_OutOfStock
		} else {
			item.Stock.StockStatus = pbItem.Stock_InStock
		}

		if colorSelected != nil {
			sku.Specs = append(sku.Specs, colorSelected)
		}

		if bytes.Contains(respBody, []byte(`>One size`)) {
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    "One size",
				Name:  "One size",
				Value: "One size",
			})
		}

		if len(sku.Specs) == 0 {
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    "-",
				Name:  "-",
				Value: "-",
			})
		}

		// for _, spec := range sku.Specs {
		// 	sku.SourceId += fmt.Sprintf("-%s", spec.Id)
		// }

		item.SkuItems = append(item.SkuItems, &sku)
	}

	// yield item result
	if err = yield(ctx, &item); err != nil {
		c.logger.Errorf("yield sub request failed, error=%s", err)
		return err
	}

	// found other color

	if ctx.Value("groupId") == nil {
		sel = doc.Find(`div[data-attr="color"]`).Find(`a`)

		for i := range sel.Nodes {
			nctx := context.WithValue(ctx, "groupId", item.GetSource().GetId())
			node := sel.Eq(i)

			if subClass := node.AttrOr(`class`, ``); !strings.Contains(subClass, `selected`) {

				u := node.AttrOr(`href`, "")
				if u == "" {
					continue
				}

				href, err := c.CanonicalUrl(u)
				if err != nil {
					continue
				}

				if req, err := http.NewRequest(http.MethodGet, href, nil); err != nil {
					c.logger.Error(err)
					continue
				} else if err := yield(nctx, req); err != nil {
					c.logger.Error(err)
					return err
				}
			}
		}
	}
	return nil
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://www.acnestudios.com/us/en/home",
		//"https://www.acnestudios.com/us/en/woman/suit-jackets/",
		//"https://www.acnestudios.com/us/en/woman/outerwear/",
		//"https://www.acnestudios.com/us/en/woman/new-arrivals/",
		//"https://www.acnestudios.com/us/en/man/new-arrivals/",
		//"https://www.acnestudios.com/us/en/man/outerwear/",
		//"https://www.acnestudios.com/us/en/man/hats/",
		//"https://www.acnestudios.com/us/en/checked-suit-jacket-brown-orange/AH0159-BOD.html",
		//"https://www.acnestudios.com/us/en/belted-trench-coat-walnut-brown/A90364-BZH.html",
		//"https://www.acnestudios.com/us/en/acne-paper-book-17/EN0050-000.html",
		//"https://www.acnestudios.com/us/en/woman/fw21-show-collection/",
		"https://www.acnestudios.com/us/en/musubi-mini/A10093-640000.html",
		//"https://www.acnestudios.com/us/en/v-neck-cardigan-navy-grey/B60144-AQ5.html",
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
