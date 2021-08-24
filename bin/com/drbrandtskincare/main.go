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
	pbProxy "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
)

// _Crawler defined the crawler struct/class for which is not necessory to be exportable
type _Crawler struct {
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
		categoryPathMatcher: regexp.MustCompile(`^(/[/a-zA-Z0-9\-]+){1,6}$`),
		// this regular used to match product page url path
		productPathMatcher: regexp.MustCompile(`^(/[/a-zA-Z0-9\-]+)/products(/[/a-zA-Z0-9\-]+)$`),
		logger:             logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "116e38308fbd44caafaef7e949674104"
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
		EnableHeadless:    false,
		EnableSessionInit: false,
		Reliability:       pbProxy.ProxyReliability_ReliabilityLow,
		MustHeader:        make(http.Header),
	}

	return opts
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.drbrandtskincare.com"}
}

func (c *_Crawler) CanonicalUrl(rawurl string) (string, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	if u.Host == "" {
		u.Host = "www.drbrandtskincare.com"
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
	return crawler.ErrUnsupportedPath
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

	sel := dom.Find(`.multi-level-nav > .tier-1 > ul > li`)
	for i := range sel.Nodes {
		node := sel.Eq(i)
		cateName := strings.TrimSpace(node.Find(`a`).First().Text())
		if cateName == "" {
			continue
		}
		//nnctx := context.WithValue(ctx, "Category", cateName)
		fmt.Println()
		fmt.Println(cateName)

		subSel := node.Find(`.nav-columns > .contains-children`)

		if len(subSel.Nodes) > 0 {
			for j := range subSel.Nodes {
				subNode := subSel.Eq(j)

				subSel1 := subNode.Find(`ul>li`)

				if len(subSel1.Nodes) > 0 {

					for k := range subSel1.Nodes {

						subCateName := strings.TrimSpace(subNode.Find(`a`).First().Text())

						subNode1 := subSel1.Eq(k)

						lastCateName := strings.TrimSpace(subNode1.Find(`a`).First().Text())

						href := subNode.Find(`a`).First().AttrOr("href", "")
						if href == "" {
							continue
						}

						_, err := url.Parse(href)
						if err != nil {
							c.logger.Error("parse url %s failed", href)
							continue
						}

						if subCateName == "" {
							subCateName = lastCateName
						} else {
							subCateName = subCateName + ">>" + lastCateName
						}

						fmt.Println(subCateName)
						// nnnctx := context.WithValue(nnctx, "SubCategory", subCateName)
						// req, _ := http.NewRequest(http.MethodGet, href, nil)
						// if err := yield(nnnctx, req); err != nil {
						// 	return err
						// }
					}
				} else {
					// category
				}

			}
		}
	}
	return nil
}

func (c *_Crawler) GetCategories(ctx context.Context) ([]*pbItem.Category, error) {

	var (
		cates   []*pbItem.Category
		cateMap = map[string]*pbItem.Category{}
	)
	if err := func(yield func(names []string, url string) error) error {

		rootUrl := "https://www.drbrandtskincare.com/"
		req, _ := http.NewRequest(http.MethodGet, rootUrl, nil)
		opts := c.CrawlOptions(req.URL)
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

		var cates []*pbItem.Category
		sel := dom.Find(`.multi-level-nav > .tier-1 > ul > li`)
		for i := range sel.Nodes {
			node := sel.Eq(i)
			cateName := strings.TrimSpace(node.Find(`a`).First().Text())
			if cateName == "" {
				continue
			}
			//nnctx := context.WithValue(ctx, "Category", cateName)
			fmt.Println()
			fmt.Println(cateName)

			cate := pbItem.Category{
				Name:  cateName,
				Url:   node.Find(`a`).First().AttrOr("href", ""),
				Depth: 1,
			}
			cates = append(cates, &cate)

			subSel := node.Find(`.nav-columns > .contains-children`)

			if len(subSel.Nodes) > 0 {
				for j := range subSel.Nodes {
					subNode := subSel.Eq(j)

					subCateName := strings.TrimSpace(subNode.Find(`a`).First().Text())

					subCate := pbItem.Category{
						Name: subCateName,
						Url:  subNode.Find(`a`).First().AttrOr("href", ""),
					}

					cate.Children = append(cate.Children, &subCate)

					subSel1 := subNode.Find(`ul>li`)

					if len(subSel1.Nodes) > 0 {

						for k := range subSel1.Nodes {

							subNode1 := subSel1.Eq(k)

							lastCateName := strings.TrimSpace(subNode1.Find(`a`).First().Text())

							href := subNode.Find(`a`).First().AttrOr("href", "")
							if href == "" {
								continue
							}

							_, err := url.Parse(href)
							if err != nil {
								c.logger.Error("parse url %s failed", href)
								continue
							}

							subCate2 := pbItem.Category{
								Name: lastCateName,
								Url:  href,
							}

							subCate.Children = append(subCate.Children, &subCate2)

							if err := yield([]string{cate.Name, subCate.Name, subCate2.Name}, subCate2.Url); err != nil {
								return err
							}

						}
					}
				}
			} else {
				// category
				if err := yield([]string{cate.Name}, cate.Url); err != nil {
					return err
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

// nextIndex used to get the index from the shared data.
// item.index is a const key for item index.
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")))
}

type parseCategoryResponse []struct {
	Type string `json:"@type"`
	URL  string `json:"url"`
}

// parseCategoryProducts parse api url from web page url
func (c *_Crawler) parseCategoryProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	// read the response data from http response
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	matched := doc.Find(`script[type="application/ld+json"]`).Last().Text()

	var r parseCategoryResponse
	if err = json.Unmarshal([]byte(matched), &r); err != nil {
		c.logger.Debugf("parse %s failed, error=%s", matched, err)
		return err
	}

	lastIndex := nextIndex(ctx)

	for _, prod := range r {

		if prod.Type != "Product" {
			continue
		}

		if req, err := http.NewRequest(http.MethodGet, prod.URL, nil); err != nil {
			c.logger.Debug(err)
			return err
		} else {
			nctx := context.WithValue(ctx, "item.index", lastIndex+1)
			lastIndex += 1
			if err = yield(nctx, req); err != nil {
				return err
			}
		}
	}

	// next page not found
	nextUrl := ""
	if nextUrl == "" {
		return nil
	}

	req, _ := http.NewRequest(http.MethodGet, nextUrl, nil)
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
}

type parseProductResponse struct {
	Context     string `json:"@context"`
	Type        string `json:"@type"`
	URL         string `json:"url"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Brand       struct {
		Name string `json:"name"`
	} `json:"brand"`
	Sku             string `json:"sku"`
	Weight          string `json:"weight"`
	AggregateRating struct {
		Type        string `json:"@type"`
		Description string `json:"description"`
		RatingValue string `json:"ratingValue"`
		ReviewCount string `json:"reviewCount"`
	} `json:"aggregateRating"`
}

type parseProductVariantsResponse struct {
	ID          int64    `json:"id"`
	Title       string   `json:"title"`
	Handle      string   `json:"handle"`
	Description string   `json:"description"`
	Vendor      string   `json:"vendor"`
	Type        string   `json:"type"`
	Tags        []string `json:"tags"`
	Available   bool     `json:"available"`
	Variants    []struct {
		ID                     int64         `json:"id"`
		Title                  string        `json:"title"`
		Option1                string        `json:"option1"`
		Option2                string        `json:"option2"`
		Option3                string        `json:"option3"`
		Sku                    string        `json:"sku"`
		RequiresShipping       bool          `json:"requires_shipping"`
		Taxable                bool          `json:"taxable"`
		FeaturedImage          interface{}   `json:"featured_image"`
		Available              bool          `json:"available"`
		Name                   string        `json:"name"`
		PublicTitle            string        `json:"public_title"`
		Options                []string      `json:"options"`
		Price                  int           `json:"price"`
		Weight                 int           `json:"weight"`
		CompareAtPrice         interface{}   `json:"compare_at_price"`
		InventoryQuantity      int           `json:"inventory_quantity"`
		InventoryManagement    string        `json:"inventory_management"`
		InventoryPolicy        string        `json:"inventory_policy"`
		Barcode                string        `json:"barcode"`
		RequiresSellingPlan    bool          `json:"requires_selling_plan"`
		SellingPlanAllocations []interface{} `json:"selling_plan_allocations"`
	} `json:"variants"`
	Images        []string `json:"images"`
	FeaturedImage string   `json:"featured_image"`
	Options       []string `json:"options"`
	Media         []struct {
		Alt      interface{} `json:"alt"`
		ID       int64       `json:"id"`
		Position int         `json:"position"`
		Src      string      `json:"src"`
		Width    int         `json:"width"`
	} `json:"media"`
}

// used to trim html labels in description
var htmlTrimRegp = regexp.MustCompile(`</?[^>]+>`)

var productVariantsReviewExtractReg = regexp.MustCompile(`(?Ums)<script type="application/json" id="cc-product-json-\d+">\s*({.*})\s*</script>`)

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

	var viewData parseProductResponse
	matched := doc.Find(`script[type="application/ld+json"]`).Last().Text()
	if err = json.Unmarshal([]byte(matched), &viewData); err != nil {
		c.logger.Debugf("parse %s failed, error=%s", matched, err)
		return err
	}

	var viewProductVariantsData parseProductVariantsResponse
	matchedProductVariants := productVariantsReviewExtractReg.FindSubmatch([]byte(respBody))
	if len(matchedProductVariants) > 1 {
		if err := json.Unmarshal(matchedProductVariants[1], &viewProductVariantsData); err != nil {
			c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
			return err
		}
	}

	canUrl, _ := c.CanonicalUrl(doc.Find(`link[rel="canonical"]`).AttrOr("href", ""))
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}

	brand := viewData.Brand.Name
	if brand == "" {
		brand = "Dr. Brandt Skincare"
	}

	rating, _ := strconv.ParseFloat(viewData.AggregateRating.RatingValue)
	reviewcount, _ := strconv.ParseInt(viewData.AggregateRating.ReviewCount)

	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           viewData.Sku,
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
			//GroupId:      doc.Find(`meta[property="product:age_group"]`).AttrOr(`content`, ``),
		},
		BrandName: brand,
		Title:     viewData.Name,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stats: &pbItem.Stats{
			ReviewCount: int32(reviewcount),
			Rating:      float32(rating),
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
	}

	item.Description = htmlTrimRegp.ReplaceAllString(viewProductVariantsData.Description, " ")

	if viewProductVariantsData.Available {
		item.Stock.StockStatus = pbItem.Stock_InStock
	}

	// itemListElement
	sel := doc.Find(`.breadcrumbs > ol > li`)
	c.logger.Debugf("nodes %d", len(sel.Nodes))
	for i := range sel.Nodes {
		if i == len(sel.Nodes)-1 {
			continue
		}
		breadcrumb := strings.TrimSpace(sel.Eq(i).Find(`a`).Text())

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

	sizeIndex := -1
	colorIndex := -1
	for i, key := range viewProductVariantsData.Options {
		if key == "Color" {
			colorIndex = i
		} else if key == "Size" {
			sizeIndex = i
		}
	}

	//shopify
	var medias []*pbMedia.Media
	for j, mediumUrl := range viewProductVariantsData.Media {
		template := strings.Split(mediumUrl.Src, ".jpg")[0]
		medias = append(medias, pbMedia.NewImageMedia(
			strconv.Format(j),
			template+".jpg",
			mediumUrl.Src+"_1000x.jpg",
			mediumUrl.Src+"_600x.jpg",
			mediumUrl.Src+"_500x.jpg",
			"",
			j == 0,
		))
	}
	item.Medias = append(item.Medias, medias...)

	for _, rawsku := range viewProductVariantsData.Variants {

		currentPrice, _ := strconv.ParseInt(rawsku.Price)
		msrp, _ := strconv.ParseInt(rawsku.CompareAtPrice)
		discount := float32(0)

		if msrp == 0 {
			msrp = currentPrice
		}
		if msrp > currentPrice {
			discount, _ = strconv.ParseFloat32((msrp - currentPrice) / msrp * 100)
		}

		sku := pbItem.Sku{
			SourceId: fmt.Sprintf("%d-%s", rawsku.ID, rawsku.Sku),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(currentPrice),
				Msrp:     int32(msrp),
				Discount: int32(discount),
			},
			Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			Medias: medias,
		}

		if viewProductVariantsData.Available {
			sku.Stock.StockStatus = pbItem.Stock_InStock
			item.Stock.StockStatus = pbItem.Stock_InStock
		}

		if colorIndex > -1 {
			colorVal := ""
			if colorIndex == 0 {
				colorVal = rawsku.Option1
			} else if colorIndex == 1 {
				colorVal = rawsku.Option2
			} else {
				colorVal = rawsku.Option3
			}

			colorSpec := pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    strconv.Format(rawsku.ID),
				Name:  colorVal,
				Value: colorVal,
			}
			sku.Specs = append(sku.Specs, &colorSpec)
		}

		if sizeIndex > -1 {
			sizeVal := ""
			if sizeIndex == 0 {
				sizeVal = rawsku.Option1
			} else if sizeIndex == 1 {
				sizeVal = rawsku.Option2
			} else {
				sizeVal = rawsku.Option3
			}

			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    rawsku.Sku,
				Name:  sizeVal,
				Value: sizeVal,
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

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://www.drbrandtskincare.com/",
		//"https://www.drbrandtskincare.com/collections/all",
		//"https://www.drbrandtskincare.com/collections/all/products/do-not-age-with-dr-brandt-triple-peptide-eye-creamdo-not-age-with-dr-brandt-triple-peptide-eye-cream",
		"https://www.drbrandtskincare.com/collections/all/products/cool-biotic",
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
