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
//func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
func (_ *_Crawler) New(_ *cli.Context, client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient: client,
		// this regular used to match category page url path
		categoryPathMatcher: regexp.MustCompile(`^((/us([/A-Za-z0-9_-]+))|(/on/demandware.store/Sites-jurlique-us-Site/en_US/Search-UpdateGrid))$`),
		productPathMatcher:  regexp.MustCompile(`^(/us[/A-Za-z0-9_-]+.html)`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "456e2e193aef4aa08ef2bff592341f9b"
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
		KeepSession:       false,
		Reliability:       proxy.ProxyReliability_ReliabilityDefault,
		MustHeader:        crawler.NewCrawlOptions().MustHeader,
	}

	return opts
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.jurlique.com"}
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
		u.Host = "www.jurlique.com"
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
	req, _ := http.NewRequest(http.MethodGet, "https://www.jurlique.com/", nil)
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

		sel := dom.Find(`.megamenu__list>li`)

		for i := range sel.Nodes {
			node := sel.Eq(i)

			cateName := strings.TrimSpace(node.Find(`a`).First().Text())
			if cateName == "" {
				continue
			}

			subSel := node.Find(`.dropdown-menu.dropdown-menu__level-second.js-navbar-sublist>li`)
			if len(subSel.Nodes) == 0 {
				subSel = node.Find(`.dropdown-menu__container.container>li`)
			}
			for k := range subSel.Nodes {
				subNode2 := subSel.Eq(k)

				subcat2 := strings.TrimSpace(subNode2.Find(`.dropdown-link.dropdown-menu__level-second__category-link`).First().Text())
				if subcat2 == "" {
					subcat2 = strings.TrimSpace(subNode2.Find(`a`).Last().Text())
				}

				subNode2list := subNode2.Find(`.dropdown-menu__level-thirds>li`)
				for j := range subNode2list.Nodes {
					subNode3 := subNode2list.Eq(j)
					subcat3 := strings.TrimSpace(subNode3.Find(`.dropdown-link`).First().Text())

					href := subNode3.Find(`a`).AttrOr("href", "")
					if href == "" || subcat3 == "" {
						continue
					}

					canonicalHref, err := c.CanonicalUrl(href)
					if err != nil {
						c.logger.Errorf("got invalid url %s", href)
						continue
					}

					u, _ := url.Parse(canonicalHref)

					if c.categoryPathMatcher.MatchString(u.Path) {
						if err := yield([]string{cateName, subcat2, subcat3}, canonicalHref); err != nil {
							return err
						}
					}
				}

				if len(subNode2list.Nodes) == 0 {
					href := subNode2.Find(`a`).First().AttrOr("href", "")
					if href == "" {
						continue
					}

					canonicalHref, err := c.CanonicalUrl(href)
					if err != nil {
						c.logger.Errorf("got invalid url %s", href)
						continue
					}

					u, _ := url.Parse(canonicalHref)

					if c.categoryPathMatcher.MatchString(u.Path) {
						if err := yield([]string{cateName, subcat2}, canonicalHref); err != nil {
							return err
						}
					}
				}
			}

			if len(subSel.Nodes) == 0 {
				if strings.ToLower(cateName) != "bestsellers" {
					continue
				}

				href := node.Find(`a`).First().AttrOr("href", "")
				if href == "" {
					continue
				}
				canonicalHref, err := c.CanonicalUrl(href)
				if err != nil {
					c.logger.Errorf("got invalid url %s", href)
					continue
				}

				u, _ := url.Parse(canonicalHref)

				if c.categoryPathMatcher.MatchString(u.Path) {
					if err := yield([]string{cateName}, canonicalHref); err != nil {
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

	sel := doc.Find(`.image-container`)

	for i := range sel.Nodes {
		node := sel.Eq(i)
		if href, _ := node.Find(`a`).Attr("href"); href != "" {
			canonicalHref, err := c.CanonicalUrl(href)
			if err != nil {
				c.logger.Errorf("got invalid url %s", href)
				continue
			}

			req, err := http.NewRequest(http.MethodGet, canonicalHref, nil)
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
	nextUrl := doc.Find(`.show-more`).Find(`button`).AttrOr(`data-url`, ``)
	if nextUrl == "" {
		return nil
	}
	nextUrl = strings.ReplaceAll(nextUrl, "&sz=12", "&sz=96")

	req, _ := http.NewRequest(http.MethodGet, nextUrl, nil)
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

type parseProductResponse struct {
	Product struct {
		ID          string `json:"id"`
		ProductName string `json:"productName"`
		ProductType string `json:"productType"`
		Price       struct {
			Sales struct {
				Value        int    `json:"value"`
				Currency     string `json:"currency"`
				Formatted    string `json:"formatted"`
				DecimalPrice string `json:"decimalPrice"`
			} `json:"sales"`
			List interface{} `json:"list"`
			HTML string      `json:"html"`
		} `json:"price"`
		Images struct {
			Large []struct {
				Alt   string `json:"alt"`
				URL   string `json:"url"`
				Title string `json:"title"`
			} `json:"large"`
			ZoomImage []struct {
				Alt   string `json:"alt"`
				URL   string `json:"url"`
				Title string `json:"title"`
			} `json:"zoomImage"`
			AltImage []struct {
				Alt   string `json:"alt"`
				URL   string `json:"url"`
				Title string `json:"title"`
			} `json:"altImage"`
			Small []struct {
				Alt   string `json:"alt"`
				URL   string `json:"url"`
				Title string `json:"title"`
			} `json:"small"`
		} `json:"images"`
		SelectedQuantity    int `json:"selectedQuantity"`
		MinOrderQuantity    int `json:"minOrderQuantity"`
		MaxOrderQuantity    int `json:"maxOrderQuantity"`
		VariationAttributes []struct {
			AttributeID string `json:"attributeId"`
			DisplayName string `json:"displayName"`
			ID          string `json:"id"`
			Swatchable  bool   `json:"swatchable"`
			Values      []struct {
				ID           string      `json:"id"`
				Description  interface{} `json:"description"`
				DisplayValue string      `json:"displayValue"`
				Value        string      `json:"value"`
				Selected     bool        `json:"selected"`
				Selectable   bool        `json:"selectable"`
				URL          string      `json:"url"`
			} `json:"values"`
			ResetURL string `json:"resetUrl"`
		} `json:"variationAttributes"`
		LongDescription  string  `json:"longDescription"`
		ShortDescription string  `json:"shortDescription"`
		Rating           float64 `json:"rating"`

		Attributes   interface{} `json:"attributes"`
		Availability struct {
			Messages        []string    `json:"messages"`
			StockStateClass string      `json:"stockStateClass"`
			InStockDate     interface{} `json:"inStockDate"`
		} `json:"availability"`
		Available bool          `json:"available"`
		Options   []interface{} `json:"options"`
	} `json:"product"`
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

	reviewCount, _ := strconv.ParseInt(doc.Find(`.bvseo-reviewCount`).Text())
	rating, _ := strconv.ParseFloat(doc.Find(`.bvseo-ratingValue`).Text())

	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           doc.Find(`.product`).AttrOr("data-pid", ""),
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
			//GroupId:      doc.Find(`meta[property="product:age_group"]`).AttrOr(`content`, ``),
		},
		BrandName: "Jurlique",
		Title:     doc.Find(`meta[property="og:title"]`).AttrOr(`content`, ``),
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stats: &pbItem.Stats{
			ReviewCount: int32(reviewCount),
			Rating:      float32(rating),
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
	}

	// desc
	description := doc.Find(`.details`).Text()
	item.Description = string(TrimSpaceNewlineInString([]byte(description)))

	// itemListElement
	sel := doc.Find(`.breadcrumb>li`)
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
		} else if i == 3 {
			item.SubCategory3 = breadcrumb
		} else if i == 4 {
			item.SubCategory4 = breadcrumb
		}
	}

	// Note : Products with discount do not found
	currentPrice, _ := strconv.ParsePrice(doc.Find(`.product-detail-container`).Find(`.sales`).Text())
	msrp, _ := strconv.ParsePrice(doc.Find(`.product-detail-container`).Find(`.sales`).Text())

	if msrp == 0 {
		msrp = currentPrice
	}
	discount := int32(0)
	if msrp > currentPrice {
		discount = int32(((msrp - currentPrice) / msrp) * 100)
	}

	sel = doc.Find(`.select-size`).First().Find(`option`)
	for i := range sel.Nodes {
		node := sel.Eq(i)

		var media []*pbMedia.Media
		var viewData parseProductResponse

		sid := node.AttrOr(`data-attr-value`, ``)
		if sid == "" {
			continue
		}

		pid := ""
		isSelected := node.AttrOr(`selected`, ``)
		if isSelected != "" {

			pid = doc.Find(`.product-wrapper`).AttrOr(`data-pid`, ``)
			currentPrice, _ = strconv.ParsePrice(doc.Find(`.product-detail-container`).Find(`.sales`).Text())
			msrp, _ = strconv.ParsePrice(doc.Find(`.product-detail-container`).Find(`.sales`).Text())

			if msrp == 0 {
				msrp = currentPrice
			}
			discount = int32(0)
			if msrp > currentPrice {
				discount = int32(((msrp - currentPrice) / msrp) * 100)
			}

			//images
			sel := doc.Find(`.product-thumbnail.js-carousel`).Find(`div`)
			for j := range sel.Nodes {
				node := sel.Eq(j)
				imgurl := strings.Split(node.Find(`img`).AttrOr(`src`, ``), "?")[0]

				media = append(media, pbMedia.NewImageMedia(
					strconv.Format(j),
					imgurl,
					imgurl+"?sw=1000&sh=1000&q=80",
					imgurl+"?sw=600&sh=600&q=80",
					imgurl+"?sw=500&sh=500&q=80",
					"", j == 0))
			}

		} else {
			// new variation request
			variantURL := node.AttrOr(`value`, ``)

			respBodyV, err := c.variationRequest(ctx, variantURL, resp.Request.URL.String())
			if err != nil {
				return err
			}

			if err := json.Unmarshal(respBodyV, &viewData); err != nil {
				c.logger.Warnf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
				//return err
			}

			pid = viewData.Product.ID
			currentPrice, _ = strconv.ParsePrice(viewData.Product.Price.Sales.Formatted)
			msrp, _ = strconv.ParsePrice(viewData.Product.Price.Sales.Formatted)

			if msrp == 0 {
				msrp = currentPrice
			}
			discount = int32(0)
			if msrp > currentPrice {
				discount = int32(((msrp - currentPrice) / msrp) * 100)
			}

			//images

			for j, img := range viewData.Product.Images.Large {
				imgurl := strings.Split(img.URL, "?")[0]

				media = append(media, pbMedia.NewImageMedia(
					strconv.Format(j),
					imgurl,
					imgurl+"?sw=1000&sh=1000&q=80",
					imgurl+"?sw=600&sh=600&q=80",
					imgurl+"?sw=500&sh=500&q=80",
					"", j == 0))
			}
		}

		sku := pbItem.Sku{
			SourceId: pid,
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(currentPrice * 100),
				Msrp:     int32(msrp * 100),
				Discount: discount,
			},
			Medias: media,
			Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}

		if isSelected != "" {
			if strings.Contains(doc.Find(`.availability.col-12.product-availability`).AttrOr(`data-available`, ``), "true") {
				item.Stock.StockStatus = pbItem.Stock_InStock
				sku.Stock.StockStatus = pbItem.Stock_InStock
			}
		} else {
			if viewData.Product.Available {
				item.Stock.StockStatus = pbItem.Stock_InStock
				sku.Stock.StockStatus = pbItem.Stock_InStock
			}
		}

		// size
		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Id:    sid,
			Name:  sid,
			Value: sid,
		})

		for _, spec := range sku.Specs {
			sku.SourceId += fmt.Sprintf("-%s", spec.Id)
		}
		item.SkuItems = append(item.SkuItems, &sku)
	}

	if len(sel.Nodes) == 0 {

		var media []*pbMedia.Media

		sid := strings.TrimSpace(doc.Find(`.one-size-variation`).First().Text())
		if sid == "" {
			sid = "-"
		}

		pid := doc.Find(`.product-wrapper`).AttrOr(`data-pid`, ``)
		currentPrice, _ = strconv.ParsePrice(doc.Find(`.product-detail-container`).Find(`.sales`).Text())
		msrp, _ = strconv.ParsePrice(doc.Find(`.product-detail-container`).Find(`.sales`).Text())

		if msrp == 0 {
			msrp = currentPrice
		}
		discount = int32(0)
		if msrp > currentPrice {
			discount = int32(((msrp - currentPrice) / msrp) * 100)
		}

		//images
		sel := doc.Find(`.product-thumbnail.js-carousel`).Find(`div`)
		for j := range sel.Nodes {
			node := sel.Eq(j)
			imgurl := strings.Split(node.Find(`img`).AttrOr(`src`, ``), "?")[0]

			media = append(media, pbMedia.NewImageMedia(
				strconv.Format(j),
				imgurl,
				imgurl+"?sw=1000&sh=1000&q=80",
				imgurl+"?sw=600&sh=600&q=80",
				imgurl+"?sw=500&sh=500&q=80",
				"", j == 0))
		}

		sku := pbItem.Sku{
			SourceId: pid,
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(currentPrice * 100),
				Msrp:     int32(msrp * 100),
				Discount: discount,
			},
			Medias: media,
			Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}

		if strings.Contains(doc.Find(`.availability.col-12.product-availability`).AttrOr(`data-available`, ``), "true") {
			item.Stock.StockStatus = pbItem.Stock_InStock
			sku.Stock.StockStatus = pbItem.Stock_InStock
		}

		// size
		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Id:    sid,
			Name:  sid,
			Value: sid,
		})

		for _, spec := range sku.Specs {
			sku.SourceId += fmt.Sprintf("-%s", spec.Id)
		}
		item.SkuItems = append(item.SkuItems, &sku)
	}

	// yield item result
	if err = yield(ctx, &item); err != nil {
		c.logger.Errorf("yield sub request failed, error=%s", err)
		return err
	}

	return nil
}

func (c *_Crawler) variationRequest(ctx context.Context, url string, referer string) ([]byte, error) {

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	opts := c.CrawlOptions(req.URL)
	req.Header.Set("accept-language", "en-GB,en-US;q=0.9,en;q=0.8")
	req.Header.Set("accept", "*/*")
	req.Header.Set("x-requested-with", "XMLHttpRequest")
	req.Header.Set("referer", referer)

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
		//"https://www.jurlique.com/us/en/homepage",
		//"https://www.jurlique.com/us/face/by-category/shop-all-face-care",
		//"https://www.jurlique.com/us/calendula-redness-rescue-soothing-moisturising-cream-CRRC.html",
		//"https://www.jurlique.com/us/moisture-plus-rare-rose-gel-cream-RMPRRG.html",
		//"https://www.jurlique.com/us/rose-love-balm-R09.html",
		//"https://www.jurlique.com/us/jojoba-carrier-oil-J01.html",
		//"https://www.jurlique.com/us/rose-silk-finishing-powder-R03.html",
		//"https://www.jurlique.com/us/herbal-recovery-signature-serum-HRS03.html",
		//"https://www.jurlique.com/us/face/by-category/shop-all-face-care",
		//"https://www.jurlique.com/us/rosewater-balancing-mist-RBM01.html",
		"https://www.jurlique.com/us/rose-silk-finishing-powder-R03.html",
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
