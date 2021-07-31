package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
//func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
func (_ *_Crawler) New(_ *cli.Context, client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient: client,
		// this regular used to match category page url path
		categoryPathMatcher: regexp.MustCompile(`^/([/A-Za-z0-9_-]+)$`),
		productPathMatcher:  regexp.MustCompile(`^/products([/A-Za-z0-9_-]+)$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "2d710c1e01e640878d69a808d7e4348c"
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
		EnableHeadless: true,
		// use js api to init session for the first request of the crawl
		EnableSessionInit: true,
		Reliability:       proxy.ProxyReliability_ReliabilityDefault,
		MustHeader:        crawler.NewCrawlOptions().MustHeader,
	}

	opts.MustHeader.Add("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	opts.MustHeader.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36")

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
	return []string{"*.nanushka.com"}
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

func productsList(ctx context.Context) string {
	return (ctx.Value("productsList").(string))
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

	sel := dom.Find(`.style_HeaderContent__13PY_>div>section`)

	for i := range sel.Nodes {
		node := sel.Eq(i)
		cateName := strings.TrimSpace(node.Find(`p`).First().Text())

		if cateName == "" {
			continue
		}

		nctx := context.WithValue(ctx, "Category", cateName)

		subSel := node.Find(`.style_SubmenuHeaderItem__content__2pANX>div`)

		for k := range subSel.Nodes {
			subNode2 := subSel.Eq(k)

			subcat2 := strings.TrimSpace(subNode2.Find(`div`).First().Text())

			nnctx := context.WithValue(nctx, "SubCategory", subcat2)

			subNode2list := subNode2.Find(`li`)

			for j := range subNode2list.Nodes {
				subNode := subNode2list.Eq(j)
				subcat3 := strings.TrimSpace(subNode.Find(`a`).First().Text())
				if subcat3 == "" {
					continue
				}

				href := subNode.Find(`a`).AttrOr("href", "")
				if href == "" {
					continue
				}

				u, err := url.Parse(href)
				if err != nil {
					c.logger.Error("parse url %s failed", href)
					continue
				}

				subCateName := strings.TrimSpace(subNode.Text())

				if c.categoryPathMatcher.MatchString(u.Path) {
					nnnctx := context.WithValue(nnctx, "SubCategory2", subCateName)
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

var categorySlugExtractReg = regexp.MustCompile(`(?U)id="__NEXT_DATA__"\s*type="application/json">\s*({.*})\s*</script>`)
var categoryProductsExtractReg = regexp.MustCompile(`(?U)window.__NEXT_REDUX_DATA__\s*=\s*({.*})\s*;`)

///\d+x\d+/
var imgExtractReg = regexp.MustCompile(`/\d+x\d+/`)

type categorySlugStructure struct {
	Query struct {
		ProductGroupSlug string `json:"product-group-slug"`
	} `json:"query"`
}

type categoryProductStructure struct {
	Variant struct {
		Variants map[int]struct {
			//Num1429 struct {
			ID          int    `json:"id"`
			Code        string `json:"code"`
			Slug        string `json:"slug"`
			ProductName string `json:"productName"`
			VariantName string `json:"variantName"`
			FullName    string `json:"fullName"`
			BaseColor   struct {
				ID               int    `json:"id"`
				Name             string `json:"name"`
				SwatchPreference string `json:"swatchPreference"`
				Color            string `json:"color"`
				SwatchImageSrc   string `json:"swatchImageSrc"`
			} `json:"baseColor"`
			RealColor struct {
				ID               int    `json:"id"`
				Name             string `json:"name"`
				SwatchPreference string `json:"swatchPreference"`
				Color            string `json:"color"`
				SwatchImageSrc   string `json:"swatchImageSrc"`
			} `json:"realColor"`
			ModelViewPhotoSrcSet   string `json:"modelViewPhotoSrcSet"`
			ProductViewPhotoSrcSet string `json:"productViewPhotoSrcSet"`
			Sizes                  map[int]struct {
				//Num5888 struct {
				ID        int    `json:"id"`
				Name      string `json:"name"`
				Available bool   `json:"available"`
				SizeID    int    `json:"sizeId"`
				Position  int    `json:"position"`
				Quantity  int    `json:"quantity"`
				Barcode   string `json:"barcode"`
				//} `json:"5888"`
			} `json:"sizes"`
			SiblingVariants      []int    `json:"siblingVariants"`
			IsOutOfStock         bool     `json:"isOutOfStock"`
			OldPrice             int      `json:"oldPrice"`
			Price                int      `json:"price"`
			Availability         bool     `json:"availability"`
			ComingSoon           bool     `json:"comingSoon"`
			BaseMaterial         string   `json:"baseMaterial"`
			Collection           string   `json:"collection"`
			RootCategory         string   `json:"rootCategory"`
			SustainabilityLabels []string `json:"sustainabilityLabels"`
			MainTaxon            string   `json:"mainTaxon"`
			//} `json:"1429"`
		} `json:"variants"`
		VariantIds []int `json:"variantIds"`
	} `json:"variant"`
	Menu struct {
		Menus []struct {
			ID       int    `json:"id"`
			Name     string `json:"name"`
			Position int    `json:"position"`
			Children []struct {
				ID       int    `json:"id"`
				Name     string `json:"name"`
				Position int    `json:"position"`
				Children []struct {
					ID       int           `json:"id"`
					Name     string        `json:"name"`
					Position int           `json:"position"`
					URL      string        `json:"url"`
					Children []interface{} `json:"children"`
				} `json:"children"`
			} `json:"children"`
		} `json:"menus"`
	} `json:"menu"`
	ProductGroup struct {
		ProductGroups map[string]struct {
			//WomenAllProducts struct {
			Slug        string `json:"slug"`
			Title       string `json:"title"`
			Description string `json:"description"`
			VariantIds  []int  `json:"variantIds"`
			//} `json:"women-all-products"`
		} `json:"productGroups"`
		ProductGroupSlugs           []string    `json:"productGroupSlugs"`
		LastVisitedProductGroupSlug interface{} `json:"lastVisitedProductGroupSlug"`
	} `json:"productGroup"`
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

	var viewDataNew categorySlugStructure
	matched := categorySlugExtractReg.FindSubmatch(respBody)
	if len(matched) > 1 {
		if err := json.Unmarshal(matched[1], &viewDataNew); err != nil {
			c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
			return err
		}
	}

	jsUrl := "https://www.nanushka.com/__skala/app-data-redux.1627662666101.js"
	reqn, err := http.NewRequest(http.MethodGet, jsUrl, nil)
	reqn.Header.Set("Referer", resp.Request.URL.String())

	jsreq, err := c.httpClient.Do(ctx, reqn)
	if err != nil {
		panic(err)
	}
	defer jsreq.Body.Close()

	respBodyJs, err := io.ReadAll(jsreq.Body)
	if err != nil {
		c.logger.Error(err)
		return err
	}

	var viewData categoryProductStructure
	matched = categoryProductsExtractReg.FindSubmatch(respBodyJs)
	if len(matched) > 1 {
		if err := json.Unmarshal(matched[1], &viewData); err != nil {
			c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
			return err
		}
	}

	lastIndex := nextIndex(ctx)
	for _, itemIds := range viewData.ProductGroup.ProductGroups[viewDataNew.Query.ProductGroupSlug].VariantIds {

		href := viewData.Variant.Variants[itemIds].Slug
		if href == "" {
			continue
		}

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

	totalCount, _ := strconv.ParsePrice(doc.Find(`.style_title__RXe71`).Text())

	if lastIndex >= (int)(totalCount) {
		return nil
	}

	return nil

	req, _ := http.NewRequest(http.MethodGet, "", nil)
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

var prodIdREgx = regexp.MustCompile(`/(\d+)-`)

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

	jsUrl := "https://www.nanushka.com/__skala/app-data-redux.1627662666101.js"
	reqn, err := http.NewRequest(http.MethodGet, jsUrl, nil)
	reqn.Header.Set("Referer", resp.Request.URL.String())

	jsreq, err := c.httpClient.Do(ctx, reqn)
	if err != nil {
		panic(err)
	}
	defer jsreq.Body.Close()

	respBodyJs, err := io.ReadAll(jsreq.Body)
	if err != nil {
		c.logger.Error(err)
		return err
	}

	var viewData categoryProductStructure
	matched := categoryProductsExtractReg.FindSubmatch(respBodyJs)
	if len(matched) > 1 {
		if err := json.Unmarshal(matched[1], &viewData); err != nil {
			c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
			return err
		}
	}

	canUrl, _ := c.CanonicalUrl(doc.Find(`link[rel="canonical"]`).AttrOr("href", ""))
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}

	matched = prodIdREgx.FindSubmatch([]byte(resp.Request.URL.Path))
	prodId := (int)(strconv.MustParseInt(string(matched[1])))

	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           strconv.Format(prodId),
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
			//GroupId:      doc.Find(`meta[property="product:age_group"]`).AttrOr(`content`, ``),
		},
		BrandName: "Nanushka",
		Title:     viewData.Variant.Variants[prodId].FullName,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
	}

	// desc
	item.Description = (doc.Find(`.style_ProductDetailContent__informations__2WglJ`).First().Text())
	//item.Description = string(TrimSpaceNewlineInString([]byte(description)))

	if viewData.Variant.Variants[prodId].Availability {
		item.Stock.StockStatus = pbItem.Stock_InStock
	}

	item.Category = viewData.Variant.Variants[prodId].RootCategory
	item.SubCategory = viewData.Variant.Variants[prodId].MainTaxon

	msrp, _ := strconv.ParsePrice(doc.Find(`.style_ProductDetailContent__price--old__3obif`).Text())
	originalPrice, _ := strconv.ParsePrice(doc.Find(`.style_ProductDetailContent__price--new__1IvjR`).Text())
	discount := 0.0
	if msrp == 0 {
		msrp = originalPrice
	}
	if msrp > originalPrice {
		discount = math.Ceil((msrp - originalPrice) / msrp * 100)
	}

	//images
	sel := doc.Find(`.style_Thumbnails__image__2GenJ`)
	for j := range sel.Nodes {
		node := sel.Eq(j)

		imgurl := node.Find(`img`).AttrOr("src", "")
		matched = imgExtractReg.FindSubmatch([]byte(imgurl))

		item.Medias = append(item.Medias, pbMedia.NewImageMedia(
			strconv.Format(j),
			imgurl,
			strings.ReplaceAll(imgurl, string(matched[0]), "/1920x1920/"),
			imgurl,
			imgurl,
			"", j == 0))
	}

	// Color
	colorSelected := &pbItem.SkuSpecOption{
		Type:  pbItem.SkuSpecType_SkuSpecColor,
		Id:    strconv.Format(viewData.Variant.Variants[prodId].RealColor.ID),
		Name:  viewData.Variant.Variants[prodId].RealColor.Name,
		Value: viewData.Variant.Variants[prodId].RealColor.Name,
		//Icon:  rawSku.Hoverimage,
	}

	for _, rawSku := range viewData.Variant.Variants[prodId].Sizes {

		sku := pbItem.Sku{
			SourceId: fmt.Sprintf("%d-%d", viewData.Variant.Variants[prodId].RealColor.ID, rawSku.ID),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(originalPrice),
				Msrp:     int32(msrp),
				Discount: int32(discount),
			},
			Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}

		if rawSku.Available {
			sku.Stock.StockStatus = pbItem.Stock_InStock
		}

		if colorSelected != nil {
			sku.Specs = append(sku.Specs, colorSelected)
		}

		// size
		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Id:    strconv.Format(rawSku.ID),
			Name:  rawSku.Name,
			Value: rawSku.Name,
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
		//"https://www.nanushka.com",
		//"https://www.nanushka.com/shop/women-dresses",
		//"https://us.maje.com/en/homepage",
		//"https://www.nanushka.com/products/10846-wisemoon-vegan-leather-bag-mole",
		"https://www.nanushka.com/products/9429-jasper-straight-leg-jeans-apricot",
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
