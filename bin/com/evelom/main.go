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
		categoryPathMatcher: regexp.MustCompile(`^/([/A-Za-z0-9_-]+)$`),
		productPathMatcher:  regexp.MustCompile(`^/([/A-Za-z0-9_-]+)/products([/A-Za-z0-9_-]+)$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "1a802ce5da394208b6feeac90dacd332"
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
	return []string{"*.evelom.com"}
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
		u.Host = "www.evelom.com"
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

	if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yield)
	} else if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
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
	req, _ := http.NewRequest(http.MethodGet, "https://www.evelom.com/", nil)
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
		sel := dom.Find(`#shopify-section-header`).Find(`nav[class="nav-bar"] > ul > li`)

		for a := range sel.Nodes {
			node := sel.Eq(a)

			catname := strings.TrimSpace(node.Find(`a`).First().Text())
			if catname == "" {
				continue
			}

			sublvl1div := node.Find(`ul > li`)
			for b := range sublvl1div.Nodes {
				sublvl1 := sublvl1div.Eq(b)
				sublvl1name := strings.TrimSpace(sublvl1.Find(`a`).First().Text())
				if sublvl1name == "" {
					continue
				}

				href := sublvl1.Find(`a`).First().AttrOr("href", "")
				if href == "" || sublvl1name == "" {
					continue
				}
				u, err := url.Parse(href)
				if err != nil {
					c.logger.Error("parse url %s failed", href)
					continue
				}

				if c.categoryPathMatcher.MatchString(u.Path) {
					if err := yield([]string{catname, sublvl1name}, href); err != nil {
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

	sel := dom.Find(`#shopify-section-header`).Find(`nav[class="nav-bar"] > ul > li`)
	fmt.Println(len(sel.Nodes))

	for a := range sel.Nodes {
		node := sel.Eq(a)

		catname := strings.TrimSpace(node.Find(`a`).First().Text())
		if catname == "" {
			continue
		}
		fmt.Println()
		fmt.Println(`CategoryName >>`, catname)

		sublvl1div := node.Find(`ul > li`)
		for b := range sublvl1div.Nodes {
			sublvl1 := sublvl1div.Eq(b)
			sublvl1name := strings.TrimSpace(sublvl1.Find(`a`).First().Text())
			if sublvl1name == "" {
				continue
			}
			fmt.Println(sublvl1name)

			href := sublvl1.Find(`a`).First().AttrOr("href", "")
			if href == "" {
				continue
			}
			_, err := url.Parse(href)
			if err != nil {
				//c.logger.Error("parse url %s failed", href)
				continue
			}

			// nnnctx := context.WithValue(nnctx, "SubCategory", sublvl2name)
			// req, _ := http.NewRequest(http.MethodGet, href, nil)
			// if err := yield(nnnctx, req); err != nil {
			// return err
		}
	}
	return nil
}

type categoryStructure struct {
	Context         string `json:"@context"`
	Type            string `json:"@type"`
	ItemListElement []struct {
		Type     string `json:"@type"`
		Position int    `json:"position"`
		URL      string `json:"url"`
		Name     string `json:"name"`
	} `json:"itemListElement"`
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

	lastIndex := nextIndex(ctx)

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	jsonstr := dom.Find(`script[type="application/ld+json"]`).Last().Text()
	var viewData categoryStructure
	if err := json.Unmarshal([]byte(jsonstr), &viewData); err != nil {
		c.logger.Errorf("unmarshal category detail data fialed, error=%s", err)
		return err
	}

	for _, procat := range viewData.ItemListElement {

		href := procat.URL
		if href == "" {
			continue
		}

		req, err := http.NewRequest(http.MethodGet, href, nil)
		if err != nil {
			c.logger.Errorf("load http request of url %s failed, error=%s", href, err)
			return err
		}

		nctx := context.WithValue(ctx, "item.index", lastIndex)
		lastIndex += 1
		if err := yield(nctx, req); err != nil {
			return err
		}
	}

	nextUrl := dom.Find(`link[rel="next"]`).AttrOr("href", "")
	if nextUrl == "" {
		return nil
	}

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
	resp = bytes.ReplaceAll(resp, []byte("  "), []byte(""))
	return resp
}

// used to trim html labels in description
var prodRegp = regexp.MustCompile(`(?Ums)<script type="application/json" id="ProductJson-product-template">\s*({.*})\s*</script>`)

type parseProductResponse struct {
	ID                   int64    `json:"id"`
	Title                string   `json:"title"`
	Handle               string   `json:"handle"`
	Description          string   `json:"description"`
	PublishedAt          string   `json:"published_at"`
	CreatedAt            string   `json:"created_at"`
	Vendor               string   `json:"vendor"`
	Type                 string   `json:"type"`
	Tags                 []string `json:"tags"`
	Price                int      `json:"price"`
	PriceMin             int      `json:"price_min"`
	PriceMax             int      `json:"price_max"`
	Available            bool     `json:"available"`
	PriceVaries          bool     `json:"price_varies"`
	CompareAtPrice       int      `json:"compare_at_price"`
	CompareAtPriceMin    int      `json:"compare_at_price_min"`
	CompareAtPriceMax    int      `json:"compare_at_price_max"`
	CompareAtPriceVaries bool     `json:"compare_at_price_varies"`
	Variants             []struct {
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
		CompareAtPrice         int           `json:"compare_at_price"`
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
		Alt          interface{} `json:"alt"`
		ID           int64       `json:"id"`
		Position     int         `json:"position"`
		PreviewImage struct {
			AspectRatio float64 `json:"aspect_ratio"`
			Height      int     `json:"height"`
			Width       int     `json:"width"`
			Src         string  `json:"src"`
		} `json:"preview_image"`
		AspectRatio float64 `json:"aspect_ratio"`
		Height      int     `json:"height"`
		MediaType   string  `json:"media_type"`
		Src         string  `json:"src"`
		Width       int     `json:"width"`
	} `json:"media"`
	RequiresSellingPlan bool          `json:"requires_selling_plan"`
	SellingPlanGroups   []interface{} `json:"selling_plan_groups"`
	Content             string        `json:"content"`
}

type parseProductReviewResponse struct {
	Context      string `json:"@context"`
	Type         string `json:"@type"`
	ReviewCount  string `json:"reviewCount"`
	RatingValue  string `json:"ratingValue"`
	ItemReviewed struct {
		Type   string `json:"@type"`
		Name   string `json:"name"`
		Offers struct {
			Type          string `json:"@type"`
			LowPrice      string `json:"lowPrice"`
			HighPrice     string `json:"highPrice"`
			PriceCurrency string `json:"priceCurrency"`
		} `json:"offers"`
	} `json:"itemReviewed"`
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

	matched := doc.Find(`script[type="application/ld+json"]`).Last().Text()
	var viewDataReview parseProductReviewResponse
	if err := json.Unmarshal([]byte(matched), &viewDataReview); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
	}

	var viewData parseProductResponse
	{
		matched := prodRegp.FindSubmatch(respBody)
		if len(matched) > 1 {
			if err := json.Unmarshal(matched[1], &viewData); err != nil {
				c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
			}
		}
	}

	reviewCount, _ := strconv.ParseInt(viewDataReview.ReviewCount)
	rating, _ := strconv.ParseFloat(viewDataReview.RatingValue)

	canUrl, _ := c.CanonicalUrl(doc.Find(`link[rel="canonical"]`).AttrOr("href", ""))
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}
	brand := viewData.Vendor
	if brand == "" {
		brand = "EVE LOM US"
	}

	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           strconv.Format(viewData.ID),
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
			//	GroupId:      doc.Find(`meta[property="product:age_group"]`).AttrOr(`content`, ``),
		},
		BrandName: brand,
		Title:     viewData.Title,
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
	description := string(TrimSpaceNewlineInString([]byte(viewData.Description)))
	item.Description = htmlTrimRegp.ReplaceAllString(description, " ")

	if viewData.Available {
		item.Stock.StockStatus = pbItem.Stock_InStock
	}

	//IMAGES  shopify
	var medias []*pbMedia.Media
	for m, mid := range viewData.Images {

		template := strings.Split(mid, "?")[0]
		medias = append(medias, pbMedia.NewImageMedia(
			strconv.Format(m),
			"https:"+template,
			"https:"+strings.ReplaceAll(template, `.jpg`, `_1000x.jpg`),
			"https:"+strings.ReplaceAll(template, `.jpg`, `_800x.jpg`),
			"https:"+strings.ReplaceAll(template, `.jpg`, `_600x.jpg`),
			"",
			m == 0,
		))
	}

	item.Medias = append(item.Medias, medias...)

	colorIndex := -1
	sizeIndex := -1
	for k, key := range viewData.Options {
		if key == "Color" {
			colorIndex = k
		} else if key == "Size" {
			sizeIndex = k
		}
	}

	for _, rawVariation := range viewData.Variants {
		current, _ := strconv.ParsePrice(viewData.Price)
		msrp, _ := strconv.ParsePrice(viewData.CompareAtPrice)
		discount := 0.0
		if msrp == 0.0 {
			msrp = current
		}
		if msrp > current {
			discount = ((msrp - current) / msrp) * 100
		}

		sku := pbItem.Sku{
			SourceId: fmt.Sprintf("%s-%s", rawVariation.Barcode, rawVariation.Sku),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(current),
				Msrp:     int32(msrp),
				Discount: int32(discount),
			},
			Medias: medias,
			Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}

		if rawVariation.Available {
			sku.Stock.StockStatus = pbItem.Stock_InStock
		}

		if colorIndex > -1 {
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    rawVariation.Barcode,
				Name:  rawVariation.Options[colorIndex],
				Value: rawVariation.Options[colorIndex],
			})
		}

		if sizeIndex > -1 {
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    rawVariation.Sku,
				Name:  rawVariation.Options[sizeIndex],
				Value: rawVariation.Options[sizeIndex],
			})
		}

		if sizeIndex == -1 && colorIndex == -1 {
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    rawVariation.Sku,
				Name:  rawVariation.Options[0],
				Value: rawVariation.Options[0],
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
		//"https://www.evelom.com/",
		//"https://www.evelom.com/collections/masks",
		//"https://www.evelom.com/collections/anti-aging",
		//"https://www.evelom.com/collections/masks/products/rescue-mask-100-ml",
		"https://www.evelom.com/collections/shop-all/products/copy-of-begin-end-ornament",
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
