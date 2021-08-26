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
		categoryPathMatcher: regexp.MustCompile(`^/us/en/([/A-Za-z0-9_-]+)/c/\d+$`),
		productPathMatcher:  regexp.MustCompile(`/([/A-Za-z0-9_-]+)/p/\d+`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "51ef793e40a04c8692766485b987dd8c"
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
	return []string{"*.drmartens.com"}
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
		u.Host = "www.drmartens.com"
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
	if p == "" || p == "/us/en" {
		return c.parseCategories(ctx, resp, yield)
	}
	if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yield) // product deatils page
	} else if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yield) // category >> productlist page
	}

	return crawler.ErrUnsupportedPath
}

// nextIndex used to get the index from the shared data.
// item.index is a const key for item index.
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")))
}

func (c *_Crawler) GetCategories(ctx context.Context) ([]*pbItem.Category, error) {
	req, _ := http.NewRequest(http.MethodGet, "https://www.drmartens.com/us/en/", nil)
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

	dom, err := resp.Selector()
	if err != nil {
		c.logger.Error(err)
		return nil, err
	}

	var (
		cates   []*pbItem.Category
		cateMap = map[string]*pbItem.Category{}
	)
	if err := func(yield func(names []string, url string) error) error {
		sel := dom.Find(`.dm-primary-nav>li`)

		for i := range sel.Nodes {
			node := sel.Eq(i)
			cateName := strings.TrimSpace(node.Find(`a`).First().Text())
			if cateName == "" {
				continue
			}

			subSel := node.Find(`.sub-navigation-section`)

			for k := range subSel.Nodes {
				subNode2 := subSel.Eq(k)
				subcat := strings.TrimSpace(subNode2.Find(`a`).First().Text())

				subNode2list := subNode2.Find(`.yCmsComponent`)
				for j := range subNode2list.Nodes {
					subNode := subNode2list.Eq(j)

					subcatname := strings.TrimSpace(subNode.Find(`a`).First().Text())

					if subcatname == "" {
						continue
					}

					href := subNode.Find(`a`).First().AttrOr("href", "")
					if href == "" {
						continue
					}

					u, err := url.Parse(href)
					if err != nil {
						c.logger.Error("parse url %s failed", href)
						continue
					}

					if c.categoryPathMatcher.MatchString(u.Path) {
						if err := yield([]string{cateName, subcat, subcatname}, href); err != nil {
							return err
						}
					}

				}

				if len(subNode2list.Nodes) == 0 {

					href := subNode2.Find(`a`).First().AttrOr("href", "")
					if href == "" {
						continue
					}

					u, err := url.Parse(href)
					if err != nil {
						c.logger.Error("parse url %s failed", href)
						continue
					}

					if c.categoryPathMatcher.MatchString(u.Path) {
						if err := yield([]string{cateName, subcat}, href); err != nil {
							return err
						}
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

// @deprecated
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

	sel := dom.Find(`.dm-primary-nav>li`)

	for i := range sel.Nodes {
		node := sel.Eq(i)
		cateName := strings.TrimSpace(node.Find(`a`).First().Text())
		if cateName == "" {
			continue
		}

		//nctx := context.WithValue(ctx, "Category", cateName)

		subSel := node.Find(`.sub-navigation-section`)

		for k := range subSel.Nodes {
			subNode2 := subSel.Eq(k)
			subcat := strings.TrimSpace(subNode2.Find(`a`).First().Text())

			subNode2list := subNode2.Find(`.yCmsComponent`)
			for j := range subNode2list.Nodes {
				subNode := subNode2list.Eq(j)

				subcatname := strings.TrimSpace(subNode.Find(`a`).First().Text())

				if subcatname == "" {
					continue
				}

				href := subNode.Find(`a`).First().AttrOr("href", "")
				if href == "" {
					continue
				}

				finalsubCatName := ""
				if subcat != "" {
					finalsubCatName = subcat + " > " + subcatname
				} else {
					finalsubCatName = subcatname
				}

				fmt.Println(finalsubCatName)

				// u, err := url.Parse(href)
				// if err != nil {
				// 	c.logger.Error("parse url %s failed", href)
				// 	continue
				// }

				// if c.categoryPathMatcher.MatchString(u.Path) {
				// 	nnctx := context.WithValue(nctx, "SubCategory", finalsubCatName)
				// 	req, _ := http.NewRequest(http.MethodGet, href, nil)
				// 	if err := yield(nnctx, req); err != nil {
				// 		return err
				// 	}
				// }

			}

			if len(subNode2list.Nodes) == 0 {
				fmt.Println(subcat)
			}
		}
	}
	return nil
}

type CategoryData struct {
	Results []struct {
		Current struct {
			Code                     string  `json:"code"`
			BaseProductCode          string  `json:"baseProductCode"`
			Summary                  string  `json:"summary"`
			Name                     string  `json:"name"`
			URL                      string  `json:"url"`
			SwatchHexCode            string  `json:"swatchHexCode"`
			ThumbnailImgURL          string  `json:"thumbnailImgUrl"`
			AlternateThumbnailImgURL string  `json:"alternateThumbnailImgUrl"`
			FormattedPrice           string  `json:"formattedPrice"`
			LabelHex                 string  `json:"labelHex"`
			InSale                   bool    `json:"inSale"`
			DisplayPriority          float64 `json:"displayPriority"`
		} `json:"current,omitempty"`
		Siblings []struct {
			Code                     string  `json:"code"`
			BaseProductCode          string  `json:"baseProductCode"`
			Summary                  string  `json:"summary"`
			Name                     string  `json:"name"`
			URL                      string  `json:"url"`
			SwatchHexCode            string  `json:"swatchHexCode"`
			ThumbnailImgURL          string  `json:"thumbnailImgUrl"`
			AlternateThumbnailImgURL string  `json:"alternateThumbnailImgUrl"`
			FormattedPrice           string  `json:"formattedPrice"`
			LabelHex                 string  `json:"labelHex,omitempty"`
			InSale                   bool    `json:"inSale"`
			DisplayPriority          float64 `json:"displayPriority,omitempty"`
			NoOfReviews              int     `json:"noOfReviews,omitempty"`
		} `json:"siblings"`
		Sid     string  `json:"sid"`
		Ratings float64 `json:"ratings"`
		Reviews int     `json:"reviews"`
	} `json:"results"`
	Pagination struct {
		NumberOfPages        int `json:"numberOfPages"`
		TotalNumberOfResults int `json:"totalNumberOfResults"`
		PageSize             int `json:"pageSize"`
		CurrentPage          int `json:"currentPage"`
	} `json:"pagination"`
	Facets []struct {
		Code        string `json:"code"`
		Name        string `json:"name"`
		MultiSelect bool   `json:"multiSelect"`
		Visible     bool   `json:"visible"`
		Values      []struct {
			Code  string `json:"code"`
			Name  string `json:"name"`
			Count int    `json:"count"`
			Query struct {
				Value string `json:"value"`
			} `json:"query"`
			Selected bool `json:"selected"`
		} `json:"values"`
		FacetID string `json:"facetId"`
	} `json:"facets"`
	SortFields []struct {
		Code     string `json:"code"`
		Selected bool   `json:"selected,omitempty"`
		Name     string `json:"name"`
		Desc     bool   `json:"desc,omitempty"`
	} `json:"sortFields"`
	Rid string `json:"rid"`
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

	var viewData CategoryData
	matched := productsListExtractReg.FindSubmatch([]byte(respBody))
	if len(matched) > 1 {

		if err := json.Unmarshal(matched[1], &viewData); err != nil {
			c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
			return err
		}
	}
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		return err
	}

	lastIndex := nextIndex(ctx)
	for _, idv := range viewData.Results {

		req, err := http.NewRequest(http.MethodGet, idv.Current.URL, nil)
		if err != nil {
			c.logger.Errorf("load http request of url %s failed, error=%s", idv.Current.URL, err)
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

	// get current page number
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}

	// check if this is the last page
	if lastIndex >= viewData.Pagination.TotalNumberOfResults || page >= int64(viewData.Pagination.NumberOfPages) {
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

func TrimSpaceNewlineInString(s []byte) []byte {
	re := regexp.MustCompile(`\n`)
	resp := re.ReplaceAll(s, []byte(" "))
	resp = bytes.ReplaceAll(resp, []byte("\\n"), []byte(""))
	resp = bytes.ReplaceAll(resp, []byte("\r"), []byte(""))
	resp = bytes.ReplaceAll(resp, []byte("\t"), []byte(""))
	resp = bytes.ReplaceAll(resp, []byte("&lt;"), []byte("<"))
	resp = bytes.ReplaceAll(resp, []byte("&gt;"), []byte(">"))
	resp = bytes.ReplaceAll(resp, []byte("  "), []byte(" "))
	return resp
}

func TrimSpaceNewlineInByte(s []byte) []byte {
	re := regexp.MustCompile(`\n`)
	resp := re.ReplaceAll(s, []byte(" "))
	resp = bytes.ReplaceAll(resp, []byte("\\n"), []byte(""))
	resp = bytes.ReplaceAll(resp, []byte("\r"), []byte(""))
	resp = bytes.ReplaceAll(resp, []byte("\t"), []byte(""))
	resp = bytes.ReplaceAll(resp, []byte("&lt;"), []byte("<"))
	resp = bytes.ReplaceAll(resp, []byte("&gt;"), []byte(">"))
	resp = bytes.ReplaceAll(resp, []byte("  "), []byte(" "))

	resp = bytes.ReplaceAll(resp, []byte("} , }"), []byte("} }"))

	return resp
}

var productDetailExtractReg = regexp.MustCompile(`(?Ums)ACC.productTabs.tabsData\s*=\s*({.*});`)
var productsExtractReg = regexp.MustCompile(`(?Ums)<script type="application/ld\+json">\s*({.*})\s*</script>`)
var productsListExtractReg = regexp.MustCompile(`(?Ums)ACC.productList.initPageLoad\(({.*})\);`)

type ProductPageData struct {
	Context     string `json:"@context"`
	Type        string `json:"@type"`
	ID          string `json:"@id"`
	Name        string `json:"name"`
	Image       string `json:"image"`
	Description string `json:"description"`
	Mpn         string `json:"mpn"`
	URL         string `json:"url"`
	Brand       struct {
		Type string `json:"@type"`
		Name string `json:"name"`
	} `json:"brand"`
	Offers struct {
		Type          string `json:"@type"`
		Price         string `json:"price"`
		PriceCurrency string `json:"priceCurrency"`
		ItemCondition string `json:"itemCondition"`
		Seller        struct {
			Type string `json:"@type"`
			Name string `json:"name"`
		} `json:"seller"`
	} `json:"offers"`
}

type ProductDetailData struct {
	ProdDetail struct {
		CloseTitle string `json:"closeTitle"`
		ViewTitle  string `json:"viewTitle"`
		Title      string `json:"title"`
		Content    string `json:"content"`
	} `json:"prodDetail"`
	HowMade struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	} `json:"howMade"`
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

	var viewData ProductPageData
	var viewDetail ProductDetailData
	matched := productsExtractReg.FindSubmatch([]byte(respBody))
	if len(matched) > 1 {
		if err := json.Unmarshal(matched[1], &viewData); err != nil {
			c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
			return err
		}
	}

	matched = productDetailExtractReg.FindSubmatch([]byte(respBody))
	if len(matched) > 1 {
		if err := json.Unmarshal(TrimSpaceNewlineInByte(matched[1]), &viewDetail); err != nil {
			c.logger.Errorf("unmarshal Detail data fetched from %s failed, error=%s", resp.Request.URL, err)
			//return err
		}
	}

	canUrl, _ := c.CanonicalUrl(doc.Find(`link[rel="canonical"]`).AttrOr("href", ""))
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}

	brand := viewData.Brand.Name
	if brand == "" {
		brand = "Dr. Martens"
	}

	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           viewData.Mpn,
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
		},
		BrandName: brand,
		Title:     viewData.Name,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
	}

	// desc
	description := htmlTrimRegp.ReplaceAllString(viewDetail.ProdDetail.Content+viewDetail.HowMade.Content, " ")
	item.Description = string(TrimSpaceNewlineInString([]byte(description)))

	currentPrice, _ := strconv.ParsePrice(doc.Find(`span[class="current-price special-price"]`).Text())
	msrp, _ := strconv.ParsePrice(doc.Find(`span[class="current-price"]`).Text())

	if msrp == 0 {
		msrp = currentPrice
	}
	discount := 0
	if msrp > currentPrice {
		discount = (int)(((msrp - currentPrice) / msrp) * 100)
	}

	//images
	sel := doc.Find(`.slider-pdp-nav-thumbnails`).Find(`picture`)
	for j := range sel.Nodes {
		node := sel.Eq(j)
		imgurl := strings.Split(node.Find(`img`).AttrOr(`data-src`, ``), "?")[0]

		item.Medias = append(item.Medias, pbMedia.NewImageMedia(
			strconv.Format(j),
			imgurl,
			imgurl+"?sw=1000&sh=1000&q=80",
			imgurl+"?sw=600&sh=600&q=80",
			imgurl+"?sw=500&sh=500&q=80",
			"", j == 0))
	}

	// itemListElement
	sel = doc.Find(`.breadcrumb>li`)
	c.logger.Debugf("nodes %d", len(sel.Nodes))
	for i := range sel.Nodes {
		if i == len(sel.Nodes)-1 {
			continue
		}
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

	details := map[string]json.RawMessage{}

	// Color
	cid := ""
	colorName := ""
	var colorSelected *pbItem.SkuSpecOption
	sel = doc.Find(`.variant-list.js-variant-list`).Find(`li`)
	for i := range sel.Nodes {
		node := sel.Eq(i)

		if spanClass := node.AttrOr(`class`, ``); strings.Contains(spanClass, `active`) {
			if err := json.Unmarshal([]byte(node.Find(`a`).AttrOr("data-json", "")), &details); err != nil {
				fmt.Println(err)
				continue
			}
			cid = strings.Trim(string(details["code"]), `"`)
			colorName = strings.Trim(string(details["name"]), `"`)
			colorSelected = &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    cid,
				Name:  colorName,
				Value: colorName,
				Icon:  strings.Trim(string(details["img"]), `"`),
			}
		}
	}

	sel = doc.Find(`.facet__list__type-productsize`).Find(`li`)
	for i := range sel.Nodes {
		node := sel.Eq(i)

		sid := node.Find(`a`).AttrOr(`data-sku-code`, ``) + "_" + node.Find(`a`).AttrOr(`data-sku-size`, ``)
		sku := pbItem.Sku{
			SourceId: fmt.Sprintf("%s-%s", cid, sid),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(currentPrice * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},
			Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}

		if spanClass := node.Find(`a`).AttrOr("class", ""); strings.Contains(spanClass, "stock-inStock") {
			sku.Stock.StockStatus = pbItem.Stock_InStock
			item.Stock.StockStatus = pbItem.Stock_InStock
		}

		if colorSelected != nil {
			sku.Specs = append(sku.Specs, colorSelected)
		}

		// size
		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Id:    sid,
			Name:  node.Find(`a`).AttrOr("data-label", ""),
			Value: node.Find(`a`).AttrOr("data-label", ""),
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
		//"https://www.drmartens.com/us/en/",
		//"https://www.drmartens.com/us/en/womens/boots/c/01010000",
		"https://www.drmartens.com/us/en/p/26228100",
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
	cli.NewApp(&_Crawler{}).Run(os.Args)
}
