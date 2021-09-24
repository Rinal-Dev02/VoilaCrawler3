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
//func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
func (_ *_Crawler) New(_ *cli.Context, client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient: client,
		// this regular used to match category page url path
		categoryPathMatcher: regexp.MustCompile(`^/en-us/(women|men|search)(/[A-Za-z0-9_-]+){0,3}$`),
		productPathMatcher:  regexp.MustCompile(`^/en-us/([A-Za-z0-9_.-]+).html$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "bfd2ff77736f4520a2044c29b9b2d0d8"
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
		MustHeader:        make(http.Header),
	}

	return opts
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.balenciaga.com"}
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
		u.Host = "www.balenciaga.com"
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
	if p == "" {
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
func (c *_Crawler) GetCategories(ctx context.Context) ([]*pbItem.Category, error) {
	req, _ := http.NewRequest(http.MethodGet, "https://www.balenciaga.com/en-us", nil)
	opts := c.CrawlOptions(req.URL)
	opts.MustHeader.Add(`cookie`, `AKA_A2=A; wrongMarketAlwaysDisplay=true; _cs_mk=0.4863887933323847_1632199040148; _ga=GA1.2.138317686.1632199041; _gid=GA1.2.637548053.1632199041; _gat_UA-32773223-11=1; dwanonymous_e915f8af445b9fa3794b1beb5d8bc1e2=acYYFvuxvRBPjSb1ezSpErJeaH; OptanonAlertBoxClosed=2021-09-21T04:37:25.519Z; _gcl_au=1.1.1214406998.1632199046; _scid=f6412e98-9484-4041-9618-26022df3e0b1; _cs_c=1; dwanonymous_365619827c6097cc7e6c2fce6acfa12a=bcgk8c1mKG5ZiMHlXXyy5ai4oD; _fbp=fb.1.1632199046263.1489727116; _pin_unauth=dWlkPVptWTJaRFkyTURNdFpqWTJaQzAwTmpjNUxUa3hNVE10TWpNMllqTXpZalkxTkdNMQ; _sctr=1|1632162600000; yOrbRD=guid=c4f4fae2-e40c-476b-9556-a12a389333a6; dwanonymous_464a9602712a6d0724dcc178c1feeb0c=adMknN60Labl6pN1imxKJTi6px; sid=N-_F7uSXIoBjR2uWQCMi9tMEQLqBWCYQpPA; dwsid=305KLYuZXhVNgY9gahb3gzulz3nuVlqFijaOVWppRQWJ_Xy0UNlESdJXtl05KqU-YImwDaoFSIMXvAH2Wuyu7w==; RT="z=1&dm=balenciaga.com&si=quchloa1rdd&ss=kttl5zem&sl=0&tt=0"; _uetsid=9ee3c9f01a9511ecabf0f9febc8f26da; _uetvid=9ee45c901a9511ecb79a47b051f1c996; OptanonConsent=isIABGlobal=false&datestamp=Tue+Sep+21+2021+10:07:43+GMT+0530+(India+Standard+Time)&version=6.15.0&hosts=&consentId=3aadf240-55b6-4892-bcc5-1c14e5854c3e&interactionCount=1&landingPath=NotLandingPage&groups=C0001:1,C0002:1,C0003:1,C0004:1&geolocation=IN;GJ&AwaitingReconsent=false; _cs_id=0c7efdce-5a50-a7b9-ea0d-3dd150a62240.1632199046.1.1632199063.1632199046.1.1666363046065; _cs_s=2.0.0.1632200863754; stc117820=tsa:1632199046348.597560813.4049835.9845535958363314.9:20210921050743|env:1|20211022043726|20210921050743|2|1073207:20220921043743|uid:1632199046347.1090284037.3425717.117820.724052943.:20220921043743|srchist:1073207:1:20211022043726:20220921043743; __cq_dnt=0; dw_dnt=0; dwac_683a9c37b2b2b6b18cc209f1b8=N-_F7uSXIoBjR2uWQCMi9tMEQLqBWCYQpPA=|dw-only|||USD|false|Etc/GMT-4|true; cquid=||; inside-us5=362630670-662680f03b8a281637281810845b69afb98f5f71c4b93679db5453ee8b9e8a49-0-0; dwac_e3bc7858d4f1f8dc2c808afbfe=N-_F7uSXIoBjR2uWQCMi9tMEQLqBWCYQpPA=|dw-only|||EUR|false|Etc/GMT+2|true; cqcid=acYYFvuxvRBPjSb1ezSpErJeaH`)
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

		sel := dom.Find(`.c-nav__list.c-nav__level1>li`)

		for i := range sel.Nodes {
			node := sel.Eq(i)
			cateName := strings.TrimSpace(node.Find(`a`).First().Text())

			if cateName == "" {
				continue
			}

			subSel := node.Find(`.c-nav__list.c-nav__level2>li`)
			if len(subSel.Nodes) == 0 {
				subSel = node.Find(`li[data-ref="group"]`)
			}
			for k := range subSel.Nodes {
				subNode2 := subSel.Eq(k)

				subcat2 := strings.TrimSpace(subNode2.Find(`.c-nav__link`).First().Text())
				if subcat2 == "" {
					subcat2 = strings.TrimSpace(subNode2.Find(`a`).Last().Text())
				}

				subNode2list := subNode2.Find(`li`)
				for j := range subNode2list.Nodes {
					subNode3 := subNode2list.Eq(j)
					subcat3 := strings.TrimSpace(subNode3.Find(`a`).First().Text())

					href := subNode3.Find(`a`).AttrOr("href", "")
					if href == "" || subcat3 == "" {
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

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	lastIndex := nextIndex(ctx)

	sel := doc.Find(`.c-product__inner.c-product__focus`)

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

	totalProducts, _ := strconv.ParsePrice(doc.Find(`.c-filters__count`).Text())

	if lastIndex >= (int)(totalProducts) {
		return nil
	}

	nextUrl := doc.Find(`link[rel="next"]`).AttrOr("href", "")
	if nextUrl == "" {
		return nil
	}
	nextUrl = strings.ReplaceAll(nextUrl, `&sz=12`, `&sz=60`)
	req, err := http.NewRequest(http.MethodGet, nextUrl, nil)
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

type parseVariationStructure struct {
	Product struct {
		AkeneoImages struct {
			Packshot []struct {
				MissingImages  bool   `json:"missingImages"`
				Ecom           string `json:"ecom"`
				Large          string `json:"large"`
				Small          string `json:"small"`
				Medium         string `json:"medium"`
				Thumbnail      string `json:"thumbnail"`
				SmallThumbnail string `json:"smallThumbnail"`
				Swatch         string `json:"swatch"`
				IsCentered     bool   `json:"isCentered"`
			} `json:"packshot"`
		} `json:"akeneoImages"`

		VariationAttributes []struct {
			AttributeID   string `json:"attributeId"`
			SelectedLabel string `json:"selectedLabel"`
			SelectedValue string `json:"selectedValue"`
			DisplayName   string `json:"displayName"`
			ID            string `json:"id"`
			Swatchable    bool   `json:"swatchable"`
			Values        []struct {
				ID           string      `json:"id"`
				Description  interface{} `json:"description"`
				DisplayValue string      `json:"displayValue"`
				Value        string      `json:"value"`
				Selectable   bool        `json:"selectable"`
				URL          string      `json:"url"`
				Images       struct {
					Swatch []interface{} `json:"swatch"`
				} `json:"images"`
			} `json:"values"`
			ResetURL struct {
			} `json:"resetUrl"`
		} `json:"variationAttributes"`
	} `json:"product"`
}

// used to trim html labels in description
// used to trim html labels in description
var htmlTrimRegp = regexp.MustCompile(`</?[^>]+>`)
var imageRegp = regexp.MustCompile(`/[A-Z-a-z_]+-`)
var productID = regexp.MustCompile(`-[A-Z-0-9]+.html`)

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
	brand := doc.Find(`type[Brand="name"]`).Text()
	if brand == "" {
		brand = "balenciaga"
	}
	pid := strings.TrimSpace(doc.Find(`span[data-bind="styleMaterialColor"]`).Text())
	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           pid,
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
			// GroupId:      doc.Find(`meta[property="product:age_group"]`).AttrOr(`content`, ``),
		},
		BrandName: brand,
		Title:     strings.TrimSpace(doc.Find(`.c-product__name`).Text()),
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
	}

	// desc
	description := (doc.Find(`.c-product__shortdesc`).Text())
	item.Description = string(TrimSpaceNewlineInString([]byte(description)))

	if strings.Contains(doc.Find(`.c-product__availabilitymsg`).AttrOr(`data-available`, ``), "true") {
		item.Stock.StockStatus = pbItem.Stock_InStock
	}

	sel := doc.Find(`.c-breadcrumbs.c-breadcrumbs--null>li`)
	for i := range sel.Nodes {
		if i >= len(sel.Nodes)-1 {
			continue
		}

		node := sel.Eq(i)
		breadcrumb := strings.TrimSpace(node.Text())

		if i == 1 {
			item.Category = breadcrumb
			item.CrowdType = breadcrumb
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

	currentPrice, _ := strconv.ParseFloat(doc.Find(`.c-price__value--current`).AttrOr("content", ""))
	msrp, _ := strconv.ParseFloat(doc.Find(`.c-price__value--old`).AttrOr("content", ""))

	if msrp == 0 {
		msrp = currentPrice
	}
	discount := 0
	if msrp > currentPrice {
		discount = (int)(((msrp - currentPrice) / msrp) * 100)
	}

	// Color
	cid := ""
	colorName := ""

	sel = doc.Find(`.c-swatches`).Find(`.c-swatches__item`)
	for ic := range sel.Nodes {
		node := sel.Eq(ic)
		var colorSelected *pbItem.SkuSpecOption

		cid = node.Find(`input`).AttrOr(`data-attr-value`, "")
		icon := node.Find(`.c-swatches__itemimage`).AttrOr(`style`, "")
		if strings.Contains(icon, "background-color") {
			icon = ""
		} else {
			icon = strings.ReplaceAll(strings.ReplaceAll(icon, "background-image: url(", ""), ")", "")
		}
		colorName = node.Find(`.c-swatches__itemimage`).AttrOr(`data-display-value`, "")
		if colorName != "" {
			colorSelected = &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    cid,
				Name:  colorName,
				Value: cid,
				Icon:  icon,
			}
		}

		if strings.Contains(node.Find(`.c-swatches__itemimage`).AttrOr(`class`, ``), `selected`) {
			//images
			var imgList []*pbMedia.Media
			sel1 := doc.Find(`.c-productcarousel__wrapper`).Find(`li`)
			for j := range sel1.Nodes {
				nodeImg := sel1.Eq(j)
				imgurl := strings.Split(nodeImg.Find(`img`).AttrOr(`data-src`, ``), "?")[0]

				imgList = append(imgList, pbMedia.NewImageMedia(
					strconv.Format(j),
					imgurl,
					imageRegp.ReplaceAllString(imgurl, "/Large-"),
					imageRegp.ReplaceAllString(imgurl, "/Medium-"),
					imageRegp.ReplaceAllString(imgurl, "/Small-"),
					"", j == 0))
			}

			//size
			selsize := doc.Find(`.c-product__sizebutton`).Find(`button`)
			for i := range selsize.Nodes {
				nodeSize := selsize.Eq(i)

				sku := pbItem.Sku{
					SourceId: pid,
					Price: &pbItem.Price{
						Currency: regulation.Currency_USD,
						Current:  int32(currentPrice * 100),
						Msrp:     int32(msrp * 100),
						Discount: int32(discount),
					},
					Medias: imgList,
					Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_InStock},
				}

				sid := (nodeSize.AttrOr("data-attr-value", ""))

				if strings.Contains(nodeSize.AttrOr("class", ""), "unselectable") {
					sku.Stock.StockStatus = pbItem.Stock_OutOfStock
				}

				if colorSelected != nil {
					sku.Specs = append(sku.Specs, colorSelected)
				}

				// size
				if sid != "" {
					sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
						Type:  pbItem.SkuSpecType_SkuSpecSize,
						Id:    sid,
						Name:  sid,
						Value: sid,
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

				for _, spec := range sku.Specs {
					sku.SourceId += fmt.Sprintf("-%s", spec.Id)
				}
				item.SkuItems = append(item.SkuItems, &sku)
			}
		} else {
			// non selected color

			var viewData parseVariationStructure

			variationURL := node.Find(`input`).AttrOr(`data-attr-href`, "")
			varResponse, err := c.VariationRequest(ctx, variationURL)
			if err != nil {
				c.logger.Errorf("request %s failed, err=%s", variationURL, err)
				return err
			}
			if err := json.Unmarshal(varResponse, &viewData); err != nil {
				c.logger.Errorf("extract product list %s failed", variationURL)
			}

			var imgList []*pbMedia.Media
			for j, rawItem := range viewData.Product.AkeneoImages.Packshot {
				imgurl := strings.Split(rawItem.Swatch, "?")[0]

				imgList = append(imgList, pbMedia.NewImageMedia(
					strconv.Format(j),
					imgurl,
					imageRegp.ReplaceAllString(imgurl, "/Large-"),
					imageRegp.ReplaceAllString(imgurl, "/Medium-"),
					imageRegp.ReplaceAllString(imgurl, "/Small-"),
					"", j == 0))
			}

			for _, rawItemColor := range viewData.Product.VariationAttributes {
				if rawItemColor.AttributeID == "size" {
					for _, rawItemSize := range rawItemColor.Values {

						sku := pbItem.Sku{
							SourceId: pid,
							Price: &pbItem.Price{
								Currency: regulation.Currency_USD,
								Current:  int32(currentPrice * 100),
								Msrp:     int32(msrp * 100),
								Discount: int32(discount),
							},
							Medias: imgList,
							Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
						}

						sid := (rawItemSize.DisplayValue)

						if rawItemSize.Selectable {
							sku.Stock.StockStatus = pbItem.Stock_InStock
						}

						if colorSelected != nil {
							sku.Specs = append(sku.Specs, colorSelected)
						}

						// size
						if sid != "" {
							sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
								Type:  pbItem.SkuSpecType_SkuSpecSize,
								Id:    sid,
								Name:  sid,
								Value: sid,
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

						for _, spec := range sku.Specs {
							sku.SourceId += fmt.Sprintf("-%s", spec.Id)
						}
						item.SkuItems = append(item.SkuItems, &sku)

					}
				}
			}
		}
	}

	// yield item result
	if err = yield(ctx, &item); err != nil {
		c.logger.Errorf("yield sub request failed, error=%s", err)
		return err
	}

	return nil
}

func (c *_Crawler) VariationRequest(ctx context.Context, rootUrl string) ([]byte, error) {

	req, _ := http.NewRequest(http.MethodGet, rootUrl, nil)
	opts := c.CrawlOptions(req.URL)
	req.Header.Set("accept-language", "en-GB,en-US;q=0.9,en;q=0.8")
	req.Header.Set("accept", "application/json")
	req.Header.Set("x-requested-with", "XMLHttpRequest")

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
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://www.balenciaga.com/en-us/women/discover/newness",
		//"https://www.balenciaga.com/en-us/women/previous-collection/view-all",
		//"https://www.balenciaga.com/en-us/track-sandal-black-pink-617543W3AJ11050.html",
		//"https://www.balenciaga.com/en-us/cash-earpods-pro-holder-dark-red-6556791LRRM6515.html",
		//"https://www.balenciaga.com/en-us/love-earth-flatground-large-fit-t-shirt-white-657059TKV949000.html",
		"https://www.balenciaga.com/en-us/caps-destroyed-flatground-t-shirt-white-651795TKVB89040.html",
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
