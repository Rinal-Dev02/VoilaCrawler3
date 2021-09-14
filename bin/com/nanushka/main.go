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
		EnableHeadless: false,
		// use js api to init session for the first request of the crawl
		EnableSessionInit: false,
		Reliability:       proxy.ProxyReliability_ReliabilityDefault,
		MustHeader:        crawler.NewCrawlOptions().MustHeader,
	}

	opts.MustHeader.Add("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	opts.MustHeader.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36")

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
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	if u.Host == "" {
		u.Host = "www.nanushka.com"
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
	req, _ := http.NewRequest(http.MethodGet, "https://www.nanushka.com/", nil)
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

		sel := dom.Find(`.style_HeaderContent__13PY_>div>section`)

		for i := range sel.Nodes {
			node := sel.Eq(i)
			cateName := strings.TrimSpace(node.Find(`p`).First().Text())

			if cateName == "" || strings.ToLower(cateName) == "our world" {
				continue
			}
			// } else if strings.ToLower(cateName) != "women" || strings.ToLower(cateName) != "men" {
			// 	continue
			// }

			subSel := node.Find(`.style_SubmenuHeaderItem__content__2pANX>div`)

			for k := range subSel.Nodes {
				subNode2 := subSel.Eq(k)

				subcat2 := strings.TrimSpace(subNode2.Find(`div`).First().Text())

				subNode2list := subNode2.Find(`li`)

				for j := range subNode2list.Nodes {
					subNode := subNode2list.Eq(j)
					subcat3 := strings.TrimSpace(subNode.Find(`a`).First().Text())
					if subcat3 == "" {
						continue
					}

					href, err := c.CanonicalUrl(subNode.Find(`a`).AttrOr("href", ""))
					if href == "" || err != nil {
						continue
					}

					u, err := url.Parse(href)
					if err != nil {
						c.logger.Error("parse url %s failed", href)
						continue
					}

					if c.categoryPathMatcher.MatchString(u.Path) {
						if err := yield([]string{cateName, subcat2, subcat3}, href); err != nil {
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
var reduxidReg = regexp.MustCompile(`/app-data-redux\.(\d+)\.js`)

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

	var viewDataNew categorySlugStructure
	matched := categorySlugExtractReg.FindSubmatch(respBody)
	if len(matched) > 1 {
		if err := json.Unmarshal(matched[1], &viewDataNew); err != nil {
			c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
			return err
		}
	}

	matched = reduxidReg.FindSubmatch(respBody)
	jsUrl := "https://www.nanushka.com/__skala" + string(matched[0])

	respBodyJs, err := c.variationRequest(ctx, jsUrl, resp.Request.URL.String())
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

		href, err := c.CanonicalUrl(viewData.Variant.Variants[itemIds].Slug)
		if href == "" || err != nil {
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

	// Note : next page not found
	return nil

}

type VariationDetail struct {
	Props struct {
		PageProps struct {
			VariantPageDetails struct {
				PagePhotoSrcSets []string      `json:"pagePhotoSrcSets"`
				Description      string        `json:"description"`
				ShortDescription string        `json:"shortDescription"`
				MetaTitle        string        `json:"metaTitle"`
				MetaKeywords     string        `json:"metaKeywords"`
				MetaDescription  string        `json:"metaDescription"`
				Materials        []interface{} `json:"materials"`
				CareOptions      []string      `json:"careOptions"`
			} `json:"variantPageDetails"`
		} `json:"pageProps"`
	} `json:"props"`
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

	var viewDataVariation VariationDetail
	if err := json.Unmarshal([]byte(doc.Find(`#__NEXT_DATA__`).Text()), &viewDataVariation); err != nil {
		c.logger.Errorf("unmarshal product variation detail data fialed, error=%s", err)
		return err
	}

	matched := reduxidReg.FindSubmatch(respBody)
	jsUrl := "https://www.nanushka.com/__skala" + string(matched[0])

	respBodyJs, err := c.variationRequest(ctx, jsUrl, resp.Request.URL.String())
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
	item.Description = viewDataVariation.Props.PageProps.VariantPageDetails.Description + " " + strings.Join(viewDataVariation.Props.PageProps.VariantPageDetails.CareOptions, `, `)
	//item.Description = string(TrimSpaceNewlineInString([]byte(description)))

	if viewData.Variant.Variants[prodId].Availability {
		item.Stock.StockStatus = pbItem.Stock_InStock
	}

	item.Category = viewData.Variant.Variants[prodId].RootCategory
	item.SubCategory = viewData.Variant.Variants[prodId].MainTaxon

	msrp, _ := strconv.ParsePrice(viewData.Variant.Variants[prodId].OldPrice)
	originalPrice, _ := strconv.ParsePrice(viewData.Variant.Variants[prodId].Price)
	discount := 0.0
	if msrp == 0 {
		msrp = originalPrice
	}
	if msrp > originalPrice {
		discount = math.Ceil((msrp - originalPrice) / msrp * 100)
	}

	//images

	for j, img := range viewDataVariation.Props.PageProps.VariantPageDetails.PagePhotoSrcSets {

		imgurl := "https://monotikcdn.com/1920x1920/in/1/webp/" + img
		if img == "" {
			continue
		}
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
	var colorSelected *pbItem.SkuSpecOption
	if viewData.Variant.Variants[prodId].RealColor.Name != "" {
		colorSelected = &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecColor,
			Id:    viewData.Variant.Variants[prodId].RealColor.Name,
			Name:  viewData.Variant.Variants[prodId].RealColor.Name,
			Value: viewData.Variant.Variants[prodId].RealColor.Name,
			Icon:  viewData.Variant.Variants[prodId].RealColor.SwatchImageSrc,
		}
	}

	for _, rawSku := range viewData.Variant.Variants[prodId].Sizes {

		sku := pbItem.Sku{
			SourceId: strconv.Format(prodId),
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
			item.Stock.StockStatus = pbItem.Stock_InStock
		}

		if colorSelected != nil {
			sku.Specs = append(sku.Specs, colorSelected)
		}

		// size
		if rawSku.Name != "" {
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    rawSku.Name,
				Name:  rawSku.Name,
				Value: rawSku.Name,
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

	if len(viewData.Variant.Variants[prodId].Sizes) == 0 {

		sku := pbItem.Sku{
			SourceId: strconv.Format(prodId),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(originalPrice),
				Msrp:     int32(msrp),
				Discount: int32(discount),
			},
			Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}

		if viewData.Variant.Variants[prodId].Availability {
			sku.Stock.StockStatus = pbItem.Stock_InStock
			item.Stock.StockStatus = pbItem.Stock_InStock
		}

		if colorSelected != nil {
			sku.Specs = append(sku.Specs, colorSelected)
		}

		if colorSelected == nil {
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

	// yield item result
	if err = yield(ctx, &item); err != nil {
		c.logger.Errorf("yield sub request failed, error=%s", err)
		return err
	}

	// other products
	if ctx.Value("groupId") == nil {
		nctx := context.WithValue(ctx, "groupId", item.GetSource().GetId())
		for _, colorSizeOption := range viewData.Variant.Variants[prodId].SiblingVariants {
			if colorSizeOption == prodId {
				continue
			}
			nextProductUrl := fmt.Sprintf("https://www.nanushka.com/products/%s", viewData.Variant.Variants[colorSizeOption].Slug)
			fmt.Println(nextProductUrl)
			if req, err := http.NewRequest(http.MethodGet, nextProductUrl, nil); err != nil {
				return err
			} else if err = yield(nctx, req); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *_Crawler) variationRequest(ctx context.Context, url string, referer string) ([]byte, error) {

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
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://www.nanushka.com",
		//"https://www.nanushka.com/shop/women-dresses",

		//"https://www.nanushka.com/products/10846-wisemoon-vegan-leather-bag-mole",
		//"https://www.nanushka.com/products/9429-jasper-straight-leg-jeans-apricot",
		"https://www.nanushka.com/products/12613-raisa-hooded-wool-and-silk-blend-sweater-rust-gray",
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
