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
		categoryPathMatcher: regexp.MustCompile(`^(/([/A-Za-z0-9_-]+))$`),
		productPathMatcher:  regexp.MustCompile(`^(/[/A-Za-z0-9&_-]+.html)`),
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
		Reliability:       proxy.ProxyReliability_ReliabilityDefault,
	}

	return opts
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.bareminerals.com"}
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
		u.Host = "www.bareminerals.com"
	}
	if c.productPathMatcher.MatchString(u.Path) {
		u.RawQuery = ""
		return u.String(), nil
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
		return c.parseCategories(ctx, resp, yield)
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
	req, _ := http.NewRequest(http.MethodGet, "https://www.bareminerals.com/", nil)
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

		sel := dom.Find(`.nav.navbar-nav.nav-menu.level-01>li`)

		for i := range sel.Nodes {
			node := sel.Eq(i)
			cateName := strings.TrimSpace(node.Find(`a`).First().Text())

			// skip OUR PURPOSE & DISCOVER, as product list not found
			if cateName == "" || strings.ToLower(cateName) == "discover" || strings.ToLower(cateName) == "our purpose" {
				continue
			}

			subSel := node.Find(`.contains-sub-sub-category.third-level-navigator`)
			if len(subSel.Nodes) == 0 {
				subSel = node.Find(`.second-level`)
			}
			for k := range subSel.Nodes {
				subNode2 := subSel.Eq(k)

				subcat2 := strings.TrimSpace(subNode2.Find(`a`).First().Text())
				if subcat2 == "" {
					subcat2 = strings.TrimSpace(subNode2.Find(`a`).Last().Text())
				}

				subNode2list := subNode2.Find(`.nav.navbar-nav>li`)
				for j := range subNode2list.Nodes {
					subNode3 := subNode2list.Eq(j)
					subcat3 := strings.TrimSpace(subNode3.Find(`a`).First().Text())

					href := subNode3.Find(`a`).AttrOr("href", "")
					if href == "" || subcat3 == "" {
						continue
					}

					u, err := url.Parse(href)
					if err != nil {
						c.logger.Error("parse url %s failed", href)
						continue
					}

					if c.categoryPathMatcher.MatchString(u.Path) {
						if !strings.Contains(href, ".bareminerals.com") {
							href = "https://www.bareminerals.com" + href
						}
						if err := yield([]string{cateName, subcat2, subcat3}, href); err != nil {
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
						if strings.ToLower(subcat2) == "offers" || strings.ToLower(subcat2) == "virtual beauty services" {
							continue
						}
						if !strings.Contains(href, ".bareminerals.com") {
							href = "https://www.bareminerals.com" + href
						}
						if err := yield([]string{cateName, subcat2}, href); err != nil {
							return err
						}
					}
				}
			}

			if len(subSel.Nodes) == 0 {

				href := node.Find(`a`).First().AttrOr("href", "")
				if href == "" {
					continue
				}

				href, err := c.CanonicalUrl(href)
				if err != nil {
					c.logger.Errorf("got invalid url %s", href)
					continue
				}
				u, _ := url.Parse(href)

				if c.categoryPathMatcher.MatchString(u.Path) {
					if !strings.Contains(href, ".bareminerals.com") {
						href = "https://www.bareminerals.com" + href
					}
					if err := yield([]string{cateName}, href); err != nil {
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

	sel := dom.Find(`.nav.navbar-nav.nav-menu.level-01>li`)

	for i := range sel.Nodes {
		node := sel.Eq(i)
		cateName := strings.TrimSpace(node.Find(`a`).First().Text())

		// skip OUR PURPOSE & DISCOVER, as product list not found
		if cateName == "" || strings.ToLower(cateName) == "discover" || strings.ToLower(cateName) == "our purpose" {
			continue
		}
		fmt.Println()
		fmt.Println("Category", cateName)

		// nctx := context.WithValue(ctx, "Category", cateName)

		subSel := node.Find(`.contains-sub-sub-category.third-level-navigator`)

		for k := range subSel.Nodes {
			subNode2 := subSel.Eq(k)

			subcat2 := strings.TrimSpace(subNode2.Find(`a`).First().Text())
			if subcat2 == "" {
				continue
			}

			//	nnctx := context.WithValue(nctx, "SubCategory", subcat2)
			//fmt.Println("SubCategory", subcat2)

			subNode2list := subNode2.Find(`.nav.navbar-nav>li`)
			for j := range subNode2list.Nodes {
				subNode3 := subNode2list.Eq(j)
				subcat3 := strings.TrimSpace(subNode3.Find(`a`).First().Text())
				if subcat3 == "" {
					continue
				}

				href := subNode3.Find(`a`).AttrOr("href", "")
				if href == "" {
					continue
				}

				_, err := url.Parse(href)
				if err != nil {
					//	c.logger.Error("parse url %s failed", href)
					continue
				}

				subCateName := strings.TrimSpace(subNode3.Text())
				fmt.Println(subcat2, " > ", subCateName)

				// if c.categoryPathMatcher.MatchString(u.Path) {
				// 	nnnctx := context.WithValue(nnctx, "SubCategory2", subCateName)
				// 	req, _ := http.NewRequest(http.MethodGet, href, nil)
				// 	if err := yield(nnnctx, req); err != nil {
				// 		return err
				// 	}
				// }
			}

			if len(subNode2list.Nodes) == 0 {
				href := subNode2.Find(`a`).AttrOr("href", "")
				if href == "" {
					continue
				}

				_, err := url.Parse(href)
				if err != nil {
					c.logger.Error("parse url %s failed", href)
					continue
				}

				// if c.categoryPathMatcher.MatchString(u.Path) {
				// 	req, _ := http.NewRequest(http.MethodGet, href, nil)
				// 	if err := yield(nnctx, req); err != nil {
				// 		return err
				// 	}
				// }
			}
		}
	}
	return nil
}

var categoryProductReg = regexp.MustCompile(`(?Ums)certonaRecommendations\(({.*})\);`)

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

	sel := doc.Find(`.name-link`)

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

	//special category link
	if len(sel.Nodes) == 0 {
		if len(doc.Find(`.pt_categorylanding`).Nodes) > 0 {
			categoryName := strings.Split(strings.TrimSuffix(resp.Request.URL.Path, "/"), `/`)
			requestURL := "https://www.res-x.com/ws/r2/Resonance.aspx?appid=bareminerals01&tk=958677643273121&pg=525672038920870&sg=1&ev=content&ei=&bx=true&sc=campaign1_rr&sc=campaign2_rr&no=20&AllCategories=" + categoryName[len(categoryName)-1] + "&ccb=certonaRecommendations&vr=5.11x&ref=&url=" + resp.Request.URL.String()
			respBodyV := c.variationRequest(ctx, requestURL, resp.Request.URL.String())

			var viewData categoryStructure
			{
				matched := categoryProductReg.FindSubmatch(respBodyV)
				if len(matched) > 1 {
					if err := json.Unmarshal(matched[1], &viewData); err != nil {
						c.logger.Errorf("unmarshal category detail data fialed, error=%s", err)
					}
				}
			}

			for _, item := range viewData.Resonance.Schemes {
				for _, itemDetail := range item.Items {

					if href := itemDetail.ProductDetailURL; href != "" {
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
			}
		}
	}

	nextUrl := doc.Find(`.infinite-scroll-placeholder`).AttrOr("data-grid-url", "")
	if nextUrl == "" {
		return nil
	}
	nextUrl = strings.ReplaceAll(nextUrl, "&sz=12", "&sz=24")

	req, _ := http.NewRequest(http.MethodGet, nextUrl, nil)
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
}

type categoryStructure struct {
	Resonance struct {
		Schemes []struct {
			Items []struct {
				ID               string `json:"ID"`
				ProductName      string `json:"ProductName"`
				ProductImageURL  string `json:"ProductImageURL"`
				ProductDetailURL string `json:"ProductDetailURL"`
				Rating           string `json:"rating"`
				ListPrice        string `json:"listPrice"`
				SalePrice        string `json:"salePrice"`
				Reviews          string `json:"reviews"`
			} `json:"items"`
		} `json:"schemes"`
	} `json:"resonance"`
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

type parseProductResponse struct {
	Context     string `json:"@context"`
	Type        string `json:"@type"`
	Name        string `json:"name"`
	Image       string `json:"image"`
	Description string `json:"description"`
	ID          string `json:"@id"`
	Sku         string `json:"sku"`
	ProductID   string `json:"productID"`
	Brand       []struct {
		Type string `json:"@type"`
		Name string `json:"name"`
	} `json:"brand"`
	AggregateRating []struct {
		Type        string `json:"@type"`
		RatingValue string `json:"ratingValue"`
		ReviewCount string `json:"reviewCount"`
	} `json:"aggregateRating"`
	Offers []struct {
		Type          string `json:"@type"`
		PriceCurrency string `json:"priceCurrency"`
		Price         int    `json:"price"`
		ItemCondition string `json:"itemCondition"`
		Availability  string `json:"availability"`
		Seller        []struct {
			Type string `json:"@type"`
			Name string `json:"name"`
		} `json:"seller"`
	} `json:"offers"`
}
type parseProductVariant struct {
	ID                string `json:"id"`
	Sku               string `json:"sku"`
	Name              string `json:"name"`
	VariantProductID  string `json:"variantProductID"`
	ProductOutOfStock string `json:"productOutOfStock"`
	ProductColor      string `json:"productColor"`
	ProductType       string `json:"productType"`
	Price             int    `json:"price"`
	Brand             string `json:"brand"`
	Variant           string `json:"variant"`
	Category          string `json:"category"`
}
type parseProductKitVariant struct {
	ProductSets []struct {
		ID       string `json:"id"`
		HashID   string `json:"hashId"`
		Name     string `json:"name"`
		ImgURL   string `json:"imgURL"`
		ItemImg  string `json:"itemImg,omitempty"`
		IsBucket bool   `json:"isBucket"`
		Products []struct {
			ID              string `json:"id"`
			Name            string `json:"name"`
			IsMasterProduct bool   `json:"isMasterProduct"`
			ImgURL          string `json:"imgURL"`
			ImgURLMedium    string `json:"imgURLMedium"`
			InStock         bool   `json:"inStock"`
		} `json:"products"`
		AccordianName string `json:"accordianName,omitempty"`
	} `json:"productSets"`
	RecommendedFrequency interface{} `json:"recommendedFrequency"`
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

	// For kit product
	// example https://www.bareminerals.com/makeup/face/foundation-kits/i-am-an-original-get-started-makeup-kit/US102713.html
	if len(doc.Find(`#bundleJson`).Nodes) > 0 {
		jsonContent := doc.Find(`#bundleJson`).AttrOr(`value`, ``)

		var viewData parseProductKitVariant
		{
			if err := json.Unmarshal([]byte(jsonContent), &viewData); err != nil {
				fmt.Println(err)
				//c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
			}
		}

		for _, items := range viewData.ProductSets {
			for _, itemsproducts := range items.Products {

				u := "https://www.bareminerals.com/" + itemsproducts.ID + ".html"
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
		}
		return nil
	}

	var viewData parseProductResponse
	var viewVariationData parseProductVariant

	sel := doc.Find(`script[type="application/ld+json"]`)

	for i := range sel.Nodes {
		node := sel.Eq(i)

		if strings.Contains(node.Text(), `"@type":"Product"`) {
			if err := json.Unmarshal([]byte(node.Text()), &viewData); err != nil {
				fmt.Println(err)
				//c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
				//return err
			}
			break
		}
	}

	if err := json.Unmarshal([]byte(doc.Find(`.tealiumProductDetails>div`).AttrOr(`data-tealium-product-variant`, ``)), &viewVariationData); err != nil {
		fmt.Println(err)
		//c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		//return err
	}

	reviewCount := int64(0)
	rating := float64(0)
	if len(viewData.AggregateRating) > 0 {
		reviewCount, _ = strconv.ParseInt(viewData.AggregateRating[0].ReviewCount)
		rating, _ = strconv.ParseFloat(viewData.AggregateRating[0].RatingValue)
	}
	canUrl, _ := c.CanonicalUrl(doc.Find(`link[rel="canonical"]`).AttrOr("href", ""))
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}

	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           viewData.ProductID,
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
			//GroupId:      doc.Find(`meta[property="product:age_group"]`).AttrOr(`content`, ``),
		},
		BrandName:   "Bare Minerals",
		Title:       viewData.Name,
		Description: strings.TrimSpace(htmlTrimRegp.ReplaceAllString(viewData.Description, "")),
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stats: &pbItem.Stats{
			ReviewCount: int32(reviewCount),
			Rating:      float32(rating),
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
	}

	sel = doc.Find(`.breadcrumb .breadcrumb-element`)
	for i := range sel.Nodes {

		node := sel.Eq(i)
		breadcrumb := strings.TrimSpace(node.Text())

		if i == 1 {
			item.Category = breadcrumb
			//item.CrowdType = breadcrumb
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

	if strings.Contains(doc.Find(`meta[property="og:availability"]`).AttrOr(`content`, ``), "IN_STOCK") {
		item.Stock.StockStatus = pbItem.Stock_InStock
	}

	currentPrice, _ := strconv.ParsePrice(doc.Find(`.price-text`).Text())
	msrp := float64(0)
	if len(doc.Find(`.price-value.hide`).Nodes) > 0 {
		msrp, _ = strconv.ParsePrice(doc.Find(`.price-value.hide`).Text())
	} else {
		msrp, _ = strconv.ParsePrice(doc.Find(`.price-standard`).Text())
	}

	if msrp == 0 {
		msrp = currentPrice
	}
	discount := int32(0)
	if msrp > currentPrice {
		discount = int32(((msrp - currentPrice) / msrp) * 100)
	}

	//images
	sel1 := doc.Find(`.tealium-pdp-image-click`)
	for j := range sel1.Nodes {
		node := sel1.Eq(j)
		imgurl := strings.Split(node.Find(`img`).AttrOr(`data-src`, ``), "?")[0]
		if imgurl == "" {
			continue
		} else if !strings.HasPrefix(imgurl, "http") {
			imgurl = "https:" + strings.Split(node.Find(`img`).AttrOr(`data-src`, ``), "?")[0]
		}
		item.Medias = append(item.Medias, pbMedia.NewImageMedia(
			strconv.Format(j),
			imgurl,
			imgurl+"?fmt=pjpeg&wid=1000&hei=1000",
			imgurl+"?fmt=pjpeg&wid=600&hei=600",
			imgurl+"?fmt=pjpeg&wid=500&hei=500",
			"", j == 0))
	}

	// Video
	sel1 = doc.Find(`.tealium-pdp-video-click`)
	for j := range sel1.Nodes {
		node := sel1.Eq(j).Parent()
		videoId := node.AttrOr(`data-video-id`, ``)
		if videoId == "" {
			continue
		}

		requestURL := "https://www.bareminerals.com/on/demandware.store/Sites-BareMinerals_US_CA-Site/en_US/Product-GetVideo?videoname=" + videoId
		respBodyV := c.variationRequest(ctx, requestURL, resp.Request.URL.String())

		docV, err := goquery.NewDocumentFromReader(bytes.NewReader(respBodyV))
		if err != nil {
			return err
		}

		videoURL := docV.Find(`iframe`).AttrOr(`src`, ``)
		if videoURL == "" {
			continue
		} else if !strings.HasPrefix(videoURL, "http") {
			videoURL = "https:" + docV.Find(`iframe`).AttrOr(`src`, ``)
		}
		item.Medias = append(item.Medias, pbMedia.NewVideoMedia(
			strconv.Format(j),
			"",
			videoURL,
			300, 300, 0, "", "",
			j == 0))
	}

	//variation-select amount-select
	sel = doc.Find(`.variation-select.amount-select>option`)

	if len(sel.Nodes) > 0 {
		for i := range sel.Nodes {
			node := sel.Eq(i)

			idVal := `#tealiumProductDetails-` + node.AttrOr(`data-product_variant_id`, ``)

			variantURL := node.AttrOr(`value`, ``)
			respBodyV := c.variationRequest(ctx, variantURL, resp.Request.URL.String())

			docV, err := goquery.NewDocumentFromReader(bytes.NewReader(respBodyV))
			if err != nil {
				return err
			}

			currentPrice, _ = strconv.ParsePrice(docV.Find(`.price-text`).Text())
			msrp = float64(0)
			if len(docV.Find(`.price-value.hide`).Nodes) > 0 {
				msrp, _ = strconv.ParsePrice(docV.Find(`.price-value.hide`).Text())
			} else {
				msrp, _ = strconv.ParsePrice(docV.Find(`.price-standard`).Text())
			}
			if msrp == 0 {
				msrp = currentPrice
			}
			discount = int32(0)
			if msrp > currentPrice {
				discount = int32(((msrp - currentPrice) / msrp) * 100)
			}

			if err := json.Unmarshal([]byte(doc.Find(idVal).First().AttrOr(`data-tealium-product-variant`, ``)), &viewVariationData); err != nil {
				fmt.Println(err)
				//c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
				//return err
			}

			sku := pbItem.Sku{
				SourceId: viewVariationData.Sku,
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(currentPrice * 100),
					Msrp:     int32(msrp * 100),
					Discount: int32(discount),
				},
				Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}

			if strings.Contains(viewVariationData.ProductOutOfStock, "In Stock") {
				sku.Stock.StockStatus = pbItem.Stock_InStock
			}

			//images
			sel1 := docV.Find(`.tealium-pdp-image-click`)
			for j := range sel1.Nodes {
				node := sel1.Eq(j)
				imgurl := strings.Split(node.Find(`img`).AttrOr(`data-src`, ``), "?")[0]
				if imgurl == "" {
					continue
				} else if !strings.HasPrefix(imgurl, "http") {
					imgurl = "https:" + strings.Split(node.Find(`img`).AttrOr(`data-src`, ``), "?")[0]
				}

				sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
					strconv.Format(j),
					imgurl,
					imgurl+"?fmt=pjpeg&wid=1000&hei=1000",
					imgurl+"?fmt=pjpeg&wid=600&hei=600",
					imgurl+"?fmt=pjpeg&wid=500&hei=500",
					"", j == 0))
			}

			// Video
			sel1 = docV.Find(`.tealium-pdp-video-click`)
			for j := range sel1.Nodes {
				node := sel1.Eq(j).Parent()
				videoId := node.AttrOr(`data-video-id`, ``)
				if videoId == "" {
					continue
				}

				requestURL := "https://www.bareminerals.com/on/demandware.store/Sites-BareMinerals_US_CA-Site/en_US/Product-GetVideo?videoname=" + videoId
				respBodyV := c.variationRequest(ctx, requestURL, resp.Request.URL.String())

				docV, err := goquery.NewDocumentFromReader(bytes.NewReader(respBodyV))
				if err != nil {
					return err
				}

				videoURL := docV.Find(`iframe`).AttrOr(`src`, ``)
				if videoURL == "" {
					continue
				} else if !strings.HasPrefix(videoURL, "http") {
					videoURL = "https:" + docV.Find(`iframe`).AttrOr(`src`, ``)
				}

				item.Medias = append(item.Medias, pbMedia.NewVideoMedia(
					strconv.Format(j),
					"",
					videoURL,
					300, 300, 0, "", "",
					j == 0))
			}

			//color
			if viewVariationData.ProductColor != "None" {
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecColor,
					Id:    viewVariationData.ProductColor,
					Name:  viewVariationData.ProductColor,
					Value: viewVariationData.ProductColor,
				})
			}

			// size
			if viewVariationData.Variant != "" {
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecSize,
					Id:    viewVariationData.Variant,
					Name:  viewVariationData.Variant,
					Value: viewVariationData.Variant,
				})
			}

			item.SkuItems = append(item.SkuItems, &sku)
		}
	} else {
		// single variation
		sku := pbItem.Sku{
			SourceId: viewVariationData.Sku,
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(currentPrice * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},
			Medias: item.Medias,
			Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}

		if strings.Contains(viewVariationData.ProductOutOfStock, "In Stock") {
			sku.Stock.StockStatus = pbItem.Stock_InStock
		}

		//color
		if viewVariationData.ProductColor != "None" {
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    viewVariationData.ProductColor,
				Name:  viewVariationData.ProductColor,
				Value: viewVariationData.ProductColor,
			})
		}

		// size
		if viewVariationData.Variant != "" {
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    viewVariationData.Variant,
				Name:  viewVariationData.Variant,
				Value: viewVariationData.Variant,
			})
		}

		if viewVariationData.Variant == "" && viewVariationData.ProductColor == "None" {
			subTitle := strings.TrimSpace(doc.Find(`.sub-header`).First().Text())
			if subTitle == "" {
				subTitle = strings.TrimSpace(doc.Find(`.product-title-block`).First().Find(`.product-name`).Text())
			}
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    subTitle,
				Name:  subTitle,
				Value: subTitle,
			})
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

func (c *_Crawler) variationRequest(ctx context.Context, url string, referer string) []byte {

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	opts := c.CrawlOptions(req.URL)
	req.Header.Set("accept-language", "en-GB,en-US;q=0.9,en;q=0.8")
	req.Header.Set("accept", "text/html, */*; q=0.01")
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
		//return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)

	return respBody
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://bareminerals.com/",
		//"https://www.bareminerals.com/skincare/skincare-explore/bareblends-customizable-skin-care/",
		//"https://www.bareminerals.com/makeup/face/all-face/",
		//"https://www.bareminerals.com/skincare/moisturizers/ageless-phyto-retinol-neck-cream/US41700210101.html",
		//"https://www.bareminerals.com/makeup/eyes/brow/strength-%26-length-serum-infused-brow-gel/USmasterslbrowgel.html",
		//"https://www.bareminerals.com/makeup/lips/all-lips/maximal-color%2C-minimal-ingredients/US41700127101.html?rrec=true",
		//"https://www.bareminerals.com/skincare/category/lips/ageless-phyto-retinol-lip-balm/US41700893101.html",
		//"https://www.bareminerals.com/offers/sale/barepro-performance-wear-powder-foundation/USmasterbareprosale.html",
		//"https://www.bareminerals.com/offers/sale/barepro-performance-wear-liquid-foundation-spf-20/USmasterbareproliquidsale.html",
		//"https://www.bareminerals.com/skincare/all-skincare/skinlongevity-green-tea-herbal-eye-mask/US41700067101.html",
		//"https://www.bareminerals.com/makeup/face/foundation-kits/i-am-an-original-get-started-makeup-kit/US102713.html",
		//"https://www.bareminerals.com/new/new-explore/poreless-skincare-2/poreless-3-step-regimen/US92860.html",
		//"https://www.bareminerals.com/makeup/face/blush/gen-nude-powder-blush/USmastergnblush.html",
		"https://www.bareminerals.com/makeup/makeup-brushes/face-brushes/beautiful-finish-foundation-brush/US77069.html",
		//"https://www.bareminerals.com/US41700196101.html",
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
