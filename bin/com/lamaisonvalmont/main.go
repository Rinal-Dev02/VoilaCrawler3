package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
		categoryPathMatcher: regexp.MustCompile(`^/us/en/[A-Za-z0-9_-]+(/[A-Za-z0-9_-]+){1,3}.html$`),
		productPathMatcher:  regexp.MustCompile(`^/us/en/[A-Za-z0-9_-]+.html$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "74fd61ff6a46b8e7b4b39797d671e72e"
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
	return []string{"*.lamaisonvalmont.com"}
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
		u.Host = "www.lamaisonvalmont.com"
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
		return crawler.ErrUnsupportedPath
	}

	respBody, _ := resp.RawBody()
	if bytes.Contains(respBody, []byte(`content="product"`)) {
		return c.parseProduct(ctx, resp, yield)
	} else if c.categoryPathMatcher.MatchString(p) {
		return c.parseCategoryProducts(ctx, resp, yield)
	}

	// if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
	// 	return c.parseProduct(ctx, resp, yield)
	// } else if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
	// 	return c.parseCategoryProducts(ctx, resp, yield)
	// }

	return crawler.ErrUnsupportedPath
}

// nextIndex used to get the index from the shared data.
// item.index is a const key for item index.
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")))
}

func (c *_Crawler) GetCategories(ctx context.Context) ([]*pbItem.Category, error) {
	req, _ := http.NewRequest(http.MethodGet, "https://www.lamaisonvalmont.com/us/en/", nil)
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
		sel := dom.Find(`.pagebuilder-column.valmont-navigation__column`)

		for i := range sel.Nodes {
			node := sel.Eq(i)
			cateName := strings.TrimSpace(node.Find(`.valmont-category-cms__menu-link`).First().Text())

			if cateName == "" {
				continue
			}

			//nctx := context.WithValue(ctx, "Category", cateName)

			subSel := node.Find(`.pagebuilder-column-group`).Find(`p`)
			subcat2 := ""

			for k := range subSel.Nodes {
				subNode2 := subSel.Eq(k)

				if strings.Contains(subNode2.AttrOr("class", ""), "valmont-category-cms__main-link") {
					subcat2 = strings.TrimSpace(subNode2.Find(`a`).First().Text())
					if subcat2 == "" {
						subcat2 = strings.TrimSpace(subNode2.Find(`span`).First().Text())
					}
				} else if strings.Contains(subNode2.AttrOr("class", ""), "valmont-category-cms__secondary-link") {
					subcat2 = strings.TrimSpace(subNode2.Find(`span`).First().Text())
				}

				subcat3 := strings.TrimSpace(subNode2.Find(`a`).First().Text())

				href := strings.TrimSpace(subNode2.Find(`a`).AttrOr("href", ""))
				if href == "" || subcat3 == "" {
					continue
				}
				href, err := c.CanonicalUrl(href)
				if err != nil {
					c.logger.Errorf("got invalid url %s", href)
					continue
				}

				u, err := url.Parse(href)
				if err != nil {
					c.logger.Error("parse url %s failed", href)
					continue
				}
				//if u.Host != "www.lamaisonvalmont.com" {
				//	c.logger.Warnf("get other website url %s in homepage", href)
				//	continue
				//}

				if p := strings.TrimSuffix(u.Path, "/"); c.categoryPathMatcher.MatchString(p) {
					//nnctx := context.WithValue(nctx, "SubCategory", subcat2)
					//nnnctx := context.WithValue(nnctx, "SubCategory2", subcat3)
					if err := yield([]string{cateName, subcat2, subcat3}, href); err != nil {
						return err
					}
				}
				// 从主页直接找到点进去 https://www.lamaisonvalmont.com/us/en/private-mind.html
				// 从列表中找到点进去 https://www.lamaisonvalmont.com/us/en/private-mind.html
				// 是同一个商品，可忽略主页中的商品链接
				//else if c.productPathMatcher.MatchString(p) {
				//	//c.logger.Warnf("get product url %s in homepage", href)
				//	if err := yield([]string{cateName, subcat2, subcat3}, href); err != nil {
				//		return err
				//	}
				//}
				//else {
				//	c.logger.Warnf("unsupport path (%s)", href)
				//	continue
				//}
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

	respBody, err := resp.RawBody()
	if err != nil {
		return err
	}

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	sel := dom.Find(`.pagebuilder-column.valmont-navigation__column`)

	for i := range sel.Nodes {
		node := sel.Eq(i)
		cateName := strings.TrimSpace(node.Find(`.valmont-category-cms__menu-link`).First().Text())

		if cateName == "" {
			continue
		}

		nctx := context.WithValue(ctx, "Category", cateName)

		subSel := node.Find(`.pagebuilder-column-group`).Find(`p`)
		subcat2 := ""

		for k := range subSel.Nodes {
			subNode2 := subSel.Eq(k)

			if strings.Contains(subNode2.AttrOr("class", ""), "valmont-category-cms__main-link") {
				subcat2 = strings.TrimSpace(subNode2.Find(`a`).First().Text())
				if subcat2 == "" {
					subcat2 = strings.TrimSpace(subNode2.Find(`span`).First().Text())
				}
			} else if strings.Contains(subNode2.AttrOr("class", ""), "valmont-category-cms__secondary-link") {
				subcat2 = strings.TrimSpace(subNode2.Find(`span`).First().Text())
			}

			subcat3 := strings.TrimSpace(subNode2.Find(`a`).First().Text())

			href := subNode2.Find(`a`).AttrOr("href", "")
			if href == "" || subcat3 == "" {
				continue
			}

			u, err := url.Parse(href)
			if err != nil {
				c.logger.Error("parse url %s failed", href)
				continue
			}

			if c.categoryPathMatcher.MatchString(u.Path) {
				nnctx := context.WithValue(nctx, "SubCategory", subcat2)
				nnnctx := context.WithValue(nnctx, "SubCategory2", subcat3)
				req, _ := http.NewRequest(http.MethodGet, href, nil)
				if err := yield(nnnctx, req); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

var productListExtractReg = regexp.MustCompile(`(?Ums)var\s*tc_vars\s*=\s*({.*})\s*var `)

type CategoryData struct {
	ProductArray []struct {
		ProductBrand              string      `json:"product_brand"`
		ProductID                 string      `json:"product_id"`
		ProductSKU                string      `json:"product_SKU"`
		ProductName               string      `json:"product_name"`
		ProductUnitprice          int         `json:"product_unitprice"`
		ProductUnitpriceTf        int         `json:"product_unitprice_tf"`
		ProductDiscount           int         `json:"product_discount"`
		ProductDiscountTf         int         `json:"product_discount_tf"`
		ProductURL                string      `json:"product_url"`
		ProductURLImg             string      `json:"product_url_img"`
		ProductCategoryExternalID string      `json:"product_category_external_id"`
		ProductCategory           string      `json:"product_category"`
		ProductVolume             interface{} `json:"product_volume"`
		ProductQty                string      `json:"product_qty"`
	} `json:"product_array"`
	ProductOptions []struct {
		ProductBrand              string      `json:"product_brand"`
		ProductID                 string      `json:"product_id"`
		ProductSKU                string      `json:"product_SKU"`
		ProductName               string      `json:"product_name"`
		ProductUnitprice          int         `json:"product_unitprice"`
		ProductUnitpriceTf        int         `json:"product_unitprice_tf"`
		ProductDiscount           int         `json:"product_discount"`
		ProductDiscountTf         int         `json:"product_discount_tf"`
		ProductURL                string      `json:"product_url"`
		ProductURLImg             string      `json:"product_url_img"`
		ProductCategoryExternalID string      `json:"product_category_external_id"`
		ProductCategory           string      `json:"product_category"`
		ProductVolume             interface{} `json:"product_volume"`
		ProductQty                string      `json:"product_qty"`
	} `json:"product_options"`
}

// parseCategoryProducts parse api url from web page url
func (c *_Crawler) parseCategoryProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	// read the response data from http response
	respBody, err := resp.RawBody()
	if err != nil {
		return err
	}

	matched := productListExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)

	}

	var viewData CategoryData
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		return err
	}

	lastIndex := nextIndex(ctx)
	for _, idv := range viewData.ProductArray {

		if idv.ProductURL == "" {
			continue
		}
		req, err := http.NewRequest(http.MethodGet, idv.ProductURL, nil)
		if err != nil {
			c.logger.Errorf("load http request of url %s failed, error=%s", idv.ProductURL, err)
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

	// nextpage not found
	return nil
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
	Attributes map[int]struct {
		//Num277 struct {
		ID      string `json:"id"`
		Code    string `json:"code"`
		Label   string `json:"label"`
		Options []struct {
			ID                 string        `json:"id"`
			Label              string        `json:"label"`
			Products           []string      `json:"products"`
			ProductsOutOfStock []interface{} `json:"products_out_of_stock"`
			RetailerURL        interface{}   `json:"retailer_url"`
		} `json:"options"`
		Position string `json:"position"`
		//} `json:"277"`
	} `json:"attributes"`
	Template       string `json:"template"`
	CurrencyFormat string `json:"currencyFormat"`
	OptionPrices   map[int]struct {
		//Num1906 struct {
		OldPrice struct {
			Amount int `json:"amount"`
		} `json:"oldPrice"`
		BasePrice struct {
			Amount int `json:"amount"`
		} `json:"basePrice"`
		FinalPrice struct {
			Amount int `json:"amount"`
		} `json:"finalPrice"`
		TierPrices []interface{} `json:"tierPrices"`
		MsrpPrice  struct {
			Amount interface{} `json:"amount"`
		} `json:"msrpPrice"`
		//} `json:"1906"`
	} `json:"optionPrices"`
	PriceFormat struct {
		Pattern           string `json:"pattern"`
		Precision         string `json:"precision"`
		RequiredPrecision string `json:"requiredPrecision"`
		DecimalSymbol     string `json:"decimalSymbol"`
		GroupSymbol       string `json:"groupSymbol"`
		GroupLength       int    `json:"groupLength"`
		IntegerRequired   bool   `json:"integerRequired"`
	} `json:"priceFormat"`
	Prices struct {
		OldPrice struct {
			Amount int `json:"amount"`
		} `json:"oldPrice"`
		BasePrice struct {
			Amount int `json:"amount"`
		} `json:"basePrice"`
		FinalPrice struct {
			Amount int `json:"amount"`
		} `json:"finalPrice"`
	} `json:"prices"`
	ProductID  string `json:"productId"`
	ChooseText string `json:"chooseText"`
	Images     map[int][]struct {
		//Num1906 []struct {
		Thumb    string      `json:"thumb"`
		Img      string      `json:"img"`
		Full     string      `json:"full"`
		Caption  interface{} `json:"caption"`
		Position string      `json:"position"`
		IsMain   bool        `json:"isMain"`
		Type     string      `json:"type"`
		VideoURL interface{} `json:"videoUrl"`
		//} `json:"1906"`
	} `json:"images"`
	Index map[int]string `json:"index"` // struct {
	//Num1906 struct {
	//Num277 string `json:"277"`
	//} `json:"1906"`
	//} `json:"index"`
}

type parseProductBreadCrumbData struct {
	ItemListElement []struct {
		Item struct {
			Name string `json:"name"`
			ID   string `json:"@id"`
		} `json:"item"`
	} `json:"itemListElement"`
}

var (
	productsPageExtractReg = regexp.MustCompile(`(?Ums)"jsonConfig":\s*({.*})\s*,\s*"jsonSwatch`)
	productsDataExtractReg = regexp.MustCompile(`(?U)<script type="application/ld\+json">\s*({.*})\s*</script>`)
	// 获取真正的商品ID and name
	productIdReg   = regexp.MustCompile(`(?U)\('product.id',\s*'(.*)'\s*\)`)
	productNameReg = regexp.MustCompile(`(?U)\('product.label',\s*'(.*)'\s*\)`)
	// category
	categoryNameReg = regexp.MustCompile(`(?U)\('category.label',\s*'(.*)'\s*\)`)
)

// parseProduct
func (c *_Crawler) parseProduct(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil {
		return nil
	}
	respBody, err := resp.RawBody()
	if err != nil {
		return err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	matched := productListExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}

	var viewData CategoryData
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		return err
	}

	var viewVariationData parseProductResponse
	{
		matched := productsPageExtractReg.FindSubmatch(respBody)
		if len(matched) > 1 {
			if err := json.Unmarshal(matched[1], &viewVariationData); err != nil {
				c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
			}
		} else {
			c.logger.Error("review data not found")
		}
	}

	var productBreadCrumb parseProductBreadCrumbData
	{
		matched := productsDataExtractReg.FindSubmatch(respBody)
		if len(matched) > 1 {
			if err := json.Unmarshal(matched[1], &productBreadCrumb); err != nil {
				c.logger.Errorf("unmarshal product breadcrumb data fialed, error=%s", err)
			}
		}
	}

	canUrl, _ := c.CanonicalUrl(doc.Find(`link[rel="canonical"]`).AttrOr("href", ""))
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}

	// get product id and name
	productIdMatcher := productIdReg.FindSubmatch(respBody)
	if len(productIdMatcher) <= 1 {
		return fmt.Errorf("product not found id in url=%s", resp.Request.URL.String())
	}
	productNameMatcher := productNameReg.FindSubmatch(respBody)
	if len(productNameMatcher) <= 1 {
		return fmt.Errorf("product not found name in url=%s", resp.Request.URL.String())
	}

	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           string(productIdMatcher[1]),
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
		},
		BrandName: viewData.ProductArray[0].ProductBrand,
		Title:     string(productNameMatcher[1]),
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_InStock},
	}

	// desc
	item.Description = string(TrimSpaceNewlineInString([]byte(strings.TrimSpace(doc.Find(`.product-benefits_info`).Text()))))
	if item.Description == "" {
		item.Description = string(TrimSpaceNewlineInString([]byte(strings.TrimSpace(doc.Find(`.product.attribute.description`).Find(`.value`).Text()))))
	}

	for i, prodBreadcrumb := range productBreadCrumb.ItemListElement {
		switch i {
		case 1:
			item.Category = prodBreadcrumb.Item.Name
		case 2:
			item.SubCategory = prodBreadcrumb.Item.Name
		case 3:
			item.SubCategory2 = prodBreadcrumb.Item.Name
		case 4:
			item.SubCategory3 = prodBreadcrumb.Item.Name
		case 5:
			item.SubCategory4 = prodBreadcrumb.Item.Name
		}
	}
	if len(productBreadCrumb.ItemListElement) == 0 {
		categoryNameMatcher := categoryNameReg.FindSubmatch(respBody)
		if len(categoryNameMatcher) > 1 {
			item.Category = string(categoryNameMatcher[1])
		}
	}

	//images
	for i, iItem := range viewVariationData.Images {
		for j, imgItem := range iItem {
			item.Medias = append(item.Medias, pbMedia.NewImageMedia(
				fmt.Sprintf("%d-%d", i, j),
				imgItem.Img,
				imgItem.Full+"?sw=800&sfrm=jpg&q=70&width=800",
				imgItem.Full+"?sw=600&sfrm=jpg&q=70&width=600",
				imgItem.Thumb+"?sw=500&sfrm=jpg&q=70&width=500",
				"", i == 0))
		}
	}

	if len(viewVariationData.Images) == 0 {

		sel := doc.Find(`.gallery-placeholder._block-content-loading`).Find(`img`)
		for j := range sel.Nodes {
			node := sel.Eq(j)
			imgurl := strings.Split(node.AttrOr(`data-amsrc`, ``), "?")[0]

			item.Medias = append(item.Medias, pbMedia.NewImageMedia(
				strconv.Format(j),
				imgurl,
				imgurl+"?width=800",
				imgurl+"?width=600",
				imgurl+"?width=500",
				"", j == 0))
		}

	}

	for _, itema := range viewVariationData.Attributes {

		for _, rawSku := range itema.Options {
			// Out of stock products and SKUs are not shown on this site
			if len(rawSku.Products) == 0 {
				continue
			}
			sku := pbItem.Sku{
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
				},
				Stock: &pbItem.Stock{StockStatus: pbItem.Stock_InStock},
			}
			// 这个网站的SKU还有个ProductId，用这个值去找别的信息
			skuProductId := 0
			// get sku id
		findSkuInfoLoop:
			for _, productOption := range viewData.ProductOptions {
				for _, product := range rawSku.Products {
					if product == productOption.ProductID {
						sku.SourceId = productOption.ProductSKU
						skuPId, err := strconv.ParseInt(productOption.ProductID)
						if err != nil {
							return fmt.Errorf("sku.ProductId ParseInt errorin url=%s, err=%s", resp.Request.URL.String(), err)
						}
						skuProductId = int(skuPId)
						msrp, _ := strconv.ParsePrice(productOption.ProductUnitprice)
						currentPrice, _ := strconv.ParsePrice(productOption.ProductUnitpriceTf)
						sku.Price.Current = int32(currentPrice * 100)
						sku.Price.Msrp = int32(msrp * 100)
						if msrp > currentPrice {
							sku.Price.Discount = int32(math.Ceil((msrp - currentPrice) / msrp * 100))
						}
						break findSkuInfoLoop
					}
				}
			}
			// not found
			if sku.SourceId == "" {
				return fmt.Errorf("sku.SourceId not found in url=%s", resp.Request.URL.String())
			}
			// get sku price
			optionPrice, ok := viewVariationData.OptionPrices[skuProductId]
			if !ok {
				return fmt.Errorf("sku.optionPrice not found in url=%s", resp.Request.URL.String())
			}
			currentPrice, _ := strconv.ParsePrice(optionPrice.FinalPrice.Amount)
			msrp, _ := strconv.ParsePrice(optionPrice.OldPrice.Amount)
			if msrp == 0 {
				msrp = currentPrice
			}
			// 这里暂时覆盖数据，没有找到折扣商品，不太确定是用哪个字段
			if msrp != 0 && currentPrice != 0 {
				sku.Price.Current = int32(currentPrice * 100)
				sku.Price.Msrp = int32(msrp * 100)
				if msrp > currentPrice {
					sku.Price.Discount = int32(math.Ceil((msrp - currentPrice) / msrp * 100))
				}
			}
			// get sku image
			skuImage, ok := viewVariationData.Images[skuProductId]
			if ok {
				for i, img := range skuImage {
					sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
						strconv.Format(i),
						img.Img,
						img.Img+"&width=800",
						img.Img+"&width=600",
						img.Img+"&width=500",
						"",
						i == 0,
					))
				}
			}

			//if len(rawSku.Products) == 0 {
			//	sku.Stock.StockStatus = pbItem.Stock_OutOfStock
			//}

			// 这里统一SkuSpecOption.Id都用Name，因为原网站不统一，多选时才能获取到，单选时没有
			if strings.ToLower(itema.Code) == "color" || strings.ToLower(itema.Label) == "color" {
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type: pbItem.SkuSpecType_SkuSpecColor,
					//Id:    rawSku.ID,
					Id:    rawSku.Label,
					Name:  rawSku.Label,
					Value: rawSku.Label,
				})
			} else if strings.ToLower(itema.Code) == "size" || strings.ToLower(itema.Label) == "size" {
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecSize,
					Id:    rawSku.Label,
					Name:  rawSku.Label,
					Value: rawSku.Label,
				})
			} else {
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecUnknown,
					Id:    rawSku.Label,
					Name:  rawSku.Label,
					Value: rawSku.Label,
				})
			}

			// 发现有一个商品有颜色和型号两个属性，但是这个型号没有对应的skuSpecOptionId
			if productVolume, ok := viewData.ProductArray[0].ProductVolume.(string); ok {
				if productVolume != "" && (itema.Code == "color" || strings.ToLower(itema.Label) == "color") {
					sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
						Type:  pbItem.SkuSpecType_SkuSpecSize,
						Id:    productVolume,
						Name:  productVolume,
						Value: productVolume,
					})
				}
			}

			item.SkuItems = append(item.SkuItems, &sku)
		}
	}

	if len(viewVariationData.Attributes) == 0 {

		for _, rawSku := range viewData.ProductOptions {

			sku := pbItem.Sku{
				SourceId: rawSku.ProductSKU,
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
				},
				Stock: &pbItem.Stock{StockStatus: pbItem.Stock_InStock},
			}

			if len(rawSku.ProductQty) == 0 {
				sku.Stock.StockStatus = pbItem.Stock_OutOfStock
			}

			// price
			msrp, _ := strconv.ParsePrice(rawSku.ProductUnitprice)
			currentPrice, _ := strconv.ParsePrice(rawSku.ProductUnitpriceTf)
			sku.Price.Current = int32(currentPrice * 100)
			sku.Price.Msrp = int32(msrp * 100)
			if msrp > currentPrice {
				sku.Price.Discount = int32(math.Ceil((msrp - currentPrice) / msrp * 100))
			}
			// media
			skuPId, err := strconv.ParseInt(rawSku.ProductID)
			if err != nil {
				return fmt.Errorf("sku.ProductId ParseInt errorin url=%s, err=%s", resp.Request.URL.String(), err)
			}
			skuImage, ok := viewVariationData.Images[int(skuPId)]
			if ok {
				for i, img := range skuImage {
					sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
						strconv.Format(i),
						img.Img,
						img.Img+"&width=800",
						img.Img+"&width=600",
						img.Img+"&width=500",
						"",
						i == 0,
					))
				}
			}
			if productVolume, ok := rawSku.ProductVolume.(string); ok {
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecSize,
					Id:    productVolume,
					Name:  productVolume,
					Value: productVolume,
				})
			} else {
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecSize,
					Id:    "-",
					Name:  "-",
					Value: "-",
				})
			}

			item.SkuItems = append(item.SkuItems, &sku)
		}
	}

	// calc product stock status
	for _, skuItem := range item.SkuItems {
		if skuItem.Stock.StockStatus == pbItem.Stock_InStock {
			item.Stock.StockStatus = pbItem.Stock_InStock
		}
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
		//"https://www.lamaisonvalmont.com/us/en/",
		//"https://www.lamaisonvalmont.com/us/en/brands/valmont.html",
		//"https://www.lamaisonvalmont.com/us/en/brands/l-elixir-des-glaciers/precious-collection.html",
		//"https://www.lamaisonvalmont.com/us/en/teint-precieux.html",
		//"https://www.lamaisonvalmont.com/us/en/sea-bliss.html",
		"https://www.lamaisonvalmont.com/us/en/deto2x-pack.html",
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
