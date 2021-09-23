package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
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
	httpClient             http.Client
	categoryPathApiMatcher *regexp.Regexp
	categoryPathMatcher    *regexp.Regexp
	productPathMatcher     *regexp.Regexp
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
		categoryPathApiMatcher: regexp.MustCompile(`^(/us/en_us/api/catalog/products/[/A-Za-z0-9_-]+)$`),
		categoryPathMatcher:    regexp.MustCompile(`^/us/en_us/([/A-Za-z0-9_-]+)$`),
		//productPathMatcher:  regexp.MustCompile(`^(/[/A-Za-z0-9_-]+.html)$`),
		productPathMatcher: regexp.MustCompile(`^/us/en_us/(.*)+\d+$`),

		logger: logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "6033d0038b295db3ca84f5a834eb50ec"
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

	return opts
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.hunterboots.com"}
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
		u.Host = "www.hunterboots.com"
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

	if c.productPathMatcher.MatchString(p) {
		return c.parseProduct(ctx, resp, yield)
	} else if c.categoryPathMatcher.MatchString(p) || c.categoryPathApiMatcher.MatchString(p) {
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

	rootUrl := "https://www.hunterboots.com/us/en_us"
	req, _ := http.NewRequest(http.MethodGet, rootUrl, nil)
	opts := c.CrawlOptions(req.URL)
	req.Header.Set("accept-language", "en-GB,en-US;q=0.9,en")
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")

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

		sel := dom.Find(`.navigation__grid-container>.navigation__sections>li`)

		for i := range sel.Nodes {

			node := sel.Eq(i)
			catname := strings.TrimSpace(node.Find(`a h2`).First().Text())
			if catname == "" || catname == "Help & Info" {
				continue
			}

			subSel := node.Find(`.navigation-column>.navigation-linkblock`)
			for k := range subSel.Nodes {
				subNode2 := subSel.Eq(k)
				subcat1 := strings.TrimSpace(subNode2.Find(`.navigation-category__link`).First().Text())

				subNode2list := subNode2.Find(`ul>li`)
				for j := range subNode2list.Nodes {
					subNode := subNode2list.Eq(j)
					subcat2 := strings.TrimSpace(subNode.Find(`.image-with-description__title`).First().Text())

					if subcat2 == "" {
						subcat2 = strings.TrimSpace(subNode.Find(`a`).First().Text())
					}

					href := subNode.Find(`a`).First().AttrOr("href", "")
					if href == "" || strings.Contains(href, `/discover/`) || strings.ToLower(subcat1) == "offers & discounts" {
						continue
					}

					canonicalHref, err := c.CanonicalUrl(href)
					if err != nil {
						c.logger.Errorf("got invalid url %s", href)
						continue
					}

					u, _ := url.Parse(canonicalHref)

					if c.categoryPathMatcher.MatchString(u.Path) {
						if err := yield([]string{catname, subcat1, subcat2}, canonicalHref); err != nil {
							return err
						}
					}
				}
				if len(subNode2list.Nodes) == 0 {
					href := subNode2.Find(`.navigation-category__link`).First().AttrOr("href", "")
					if href == "" || strings.Contains(href, `/discover/`) {
						continue
					}

					canonicalHref, err := c.CanonicalUrl(href)
					if err != nil {
						c.logger.Errorf("got invalid url %s", href)
						continue
					}

					u, _ := url.Parse(canonicalHref)
					if c.categoryPathMatcher.MatchString(u.Path) {
						if err := yield([]string{catname, subcat1}, canonicalHref); err != nil {
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

type categoryProductsResponse struct {
	Meta struct {
		Page struct {
			MaxPerPage int `json:"max_per_page"`
			Pages      int `json:"pages"`
			Results    int `json:"results"`
			Page       int `json:"page"`
		} `json:"page"`
	} `json:"meta"`
	Links struct {
		Next string `json:"next"`
	} `json:"links"`
	Data []struct {
		Attributes struct {
			URL string `json:"url"`
		} `json:"attributes"`
	} `json:"data"`
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

	var viewData categoryProductsResponse

	if !c.categoryPathApiMatcher.MatchString(resp.Request.URL.Path) {

		s := strings.Split(strings.TrimSuffix(resp.Request.URL.Path, `/`), `/`)
		pid := s[len(s)-1]

		rootURL := "https://www.hunterboots.com/us/en_us/api/catalog/products/" + pid + "/us/EUPG01/en_US/?page[number]=1&page[size]=24"

		respBodyC, err := c.variationRequest(ctx, rootURL, resp.Request.URL.String())
		if err != nil {
			c.logger.Errorf("request %s error=%s", rootURL, err)
			return err
		}

		if err := json.Unmarshal(respBodyC, &viewData); err != nil {
			c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
			return err
		}
	} else {
		if err := json.Unmarshal(respBody, &viewData); err != nil {
			c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
			return err
		}
	}

	if len(viewData.Data) == 0 {
		fmt.Println(`not proper response `)
		return nil
	}

	for _, idv := range viewData.Data {

		rawurl, err := c.CanonicalUrl(idv.Attributes.URL)
		if err != nil {
			continue
		}

		req, err := http.NewRequest(http.MethodGet, rawurl, nil)
		if err != nil {
			c.logger.Errorf("load http request of url %s failed, error=%s", rawurl, err)
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

	// check if this is the last page
	if lastIndex >= viewData.Meta.Page.Results || viewData.Meta.Page.Page >= viewData.Meta.Page.Pages {
		return nil
	}

	u, _ := url.QueryUnescape(viewData.Links.Next)

	req, _ := http.NewRequest(http.MethodGet, u, nil)
	// update the index of last page
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
}

func TrimSpaceNewlineInString(s string) string {
	re := regexp.MustCompile(`\n`)
	resp := re.ReplaceAllString(s, " ")
	resp = strings.ReplaceAll(resp, "\\n", " ")
	resp = strings.ReplaceAll(resp, "\r", " ")
	resp = strings.ReplaceAll(resp, "\t", " ")
	resp = strings.ReplaceAll(resp, "  ", " ")
	return resp
}

type parseProductResponse struct {
	ID                  int    `json:"id"`
	Reference           string `json:"reference"`
	Name                string `json:"name"`
	URL                 string `json:"url"`
	StyleCode           string `json:"styleCode"`
	ColorName           string `json:"colorName"`
	ColorCode           string `json:"colorCode"`
	CategoryPath        string `json:"categoryPath"`
	AvailableToPurchase bool   `json:"availableToPurchase"`

	Images []struct {
		Ratio   string `json:"ratio"`
		RiasURL string `json:"riasUrl"`
	} `json:"images"`
	Pricing struct {
		AsString   string `json:"asString"`
		Min        string `json:"min"`
		MinAsMoney struct {
			Value    int    `json:"value"`
			Currency string `json:"currency"`
		} `json:"minAsMoney"`
		Max        string `json:"max"`
		MaxAsMoney struct {
			Value    int    `json:"value"`
			Currency string `json:"currency"`
		} `json:"maxAsMoney"`
		MaxRetail        string `json:"maxRetail"`
		MaxRetailAsMoney struct {
			Value    int    `json:"value"`
			Currency string `json:"currency"`
		} `json:"maxRetailAsMoney"`
		Currency       string `json:"currency"`
		TaxIsInclusive bool   `json:"taxIsInclusive"`
	} `json:"pricing"`
	Siblings []struct {
		ID        int      `json:"id"`
		Reference string   `json:"reference"`
		URL       string   `json:"url"`
		ColorName string   `json:"colorName"`
		ColorHex  []string `json:"colorHex"`
		ColorCode string   `json:"colorCode"`
		Images    []struct {
			Ratio   string `json:"ratio"`
			RiasURL string `json:"riasUrl"`
		} `json:"images"`
		Pricing struct {
			AsString   string `json:"asString"`
			Min        string `json:"min"`
			MinAsMoney struct {
				Value    int    `json:"value"`
				Currency string `json:"currency"`
			} `json:"minAsMoney"`
			Max        string `json:"max"`
			MaxAsMoney struct {
				Value    int    `json:"value"`
				Currency string `json:"currency"`
			} `json:"maxAsMoney"`
			MaxRetail        string `json:"maxRetail"`
			MaxRetailAsMoney struct {
				Value    int    `json:"value"`
				Currency string `json:"currency"`
			} `json:"maxRetailAsMoney"`
			Currency       string `json:"currency"`
			TaxIsInclusive bool   `json:"taxIsInclusive"`
		} `json:"pricing"`
	} `json:"siblings"`

	ProductDetails struct {
		Specification []struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"specification"`
		HasWarranty bool   `json:"hasWarranty"`
		Description string `json:"description"`
		Features    string `json:"features"`
	} `json:"productDetails"`
}
type parseProductVariationResponse []struct {
	Ean13       string `json:"ean13"`
	Size        string `json:"size"`
	HasStock    bool   `json:"hasStock"`
	LowStock    bool   `json:"lowStock"`
	CanPreOrder bool   `json:"canPreOrder"`
	CanNotify   bool   `json:"canNotify"`
}

var productsExtractReg = regexp.MustCompile(`(?Ums)<script type="application/ld\+json">\s*({.*})\s*</script>`)

var productsNewExtractReg = regexp.MustCompile(`(?Ums)window\['REACT_QUERY_INITIAL_DATA'\]\.concat\(\[({.*})\]\);`)

type ProductPageData struct {
	Context     string `json:"@context"`
	Type        string `json:"@type"`
	Sku         string `json:"sku"`
	Name        string `json:"name"`
	Image       string `json:"image"`
	Description string `json:"description"`
	Brand       struct {
		Type string `json:"@type"`
		Name string `json:"name"`
	} `json:"brand"`
	Offers []struct {
		Type          string `json:"@type"`
		PriceCurrency string `json:"priceCurrency"`
		Price         int    `json:"price"`
		Sku           string `json:"sku"`
		Gtin          string `json:"gtin"`
		URL           string `json:"url"`
		ItemCondition string `json:"itemCondition"`
		Availability  string `json:"availability"`
		Seller        struct {
			Type string `json:"@type"`
			Name string `json:"name"`
		} `json:"seller"`
	} `json:"offers"`
	AggregateRating struct {
		Type        string `json:"@type"`
		BestRating  string `json:"bestRating"`
		RatingValue string `json:"ratingValue"`
		ReviewCount string `json:"reviewCount"`
	} `json:"aggregateRating"`
	Review []struct {
		Type          string `json:"@type"`
		Author        string `json:"author"`
		DatePublished string `json:"datePublished"`
		ReviewBody    string `json:"reviewBody"`
		Name          string `json:"name"`
		ReviewRating  struct {
			Type        string `json:"@type"`
			BestRating  string `json:"bestRating"`
			RatingValue string `json:"ratingValue"`
			WorstRating string `json:"worstRating"`
		} `json:"reviewRating"`
	} `json:"review"`
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

	var viewDataProduct ProductPageData
	matched := productsExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}

	if err := json.Unmarshal(matched[1], &viewDataProduct); err != nil {
		c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		return err
	}

	rootURL := "https://www.hunterboots.com/us/en_us/api/product/skus/EUPG01/" + viewDataProduct.Sku

	respBodyV, err := c.variationRequest(ctx, rootURL, resp.Request.URL.String())
	if err != nil {
		c.logger.Errorf("request %s error=%s", rootURL, err)
		return err
	}

	var viewDataVariation parseProductVariationResponse

	if err := json.Unmarshal(respBodyV, &viewDataVariation); err != nil {
		c.logger.Errorf("unmarshal parseProductVariation data fetched from %s failed, error=%s", resp.Request.URL, err)
		return err
	}

	var viewDataDetail parseProductResponse
	matched = productsNewExtractReg.FindSubmatch(respBody)
	if len(matched) > 1 {

		ret := map[string]json.RawMessage{}
		if err := json.Unmarshal((matched[1]), &ret); err != nil {
			fmt.Println(err)
		}

		ret["data"] = bytes.TrimSuffix(bytes.TrimPrefix(bytes.ReplaceAll(bytes.ReplaceAll(ret["data"], []byte(`\\u`), []byte(`\u`)), []byte(`\"`), []byte(`"`)), []byte(`"`)), []byte(`"`))

		if err := json.Unmarshal([]byte(html.UnescapeString(string(ret["data"]))), &viewDataDetail); err != nil {
			c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
			return err
		}
	}

	canUrl, _ := c.CanonicalUrl(doc.Find(`link[rel="canonical"]`).AttrOr("href", ""))
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}

	brand := viewDataProduct.Brand.Name
	if brand == "" {
		brand = "Hunter Boot Ltd"
	}

	reviews, _ := strconv.ParseInt(viewDataProduct.AggregateRating.ReviewCount)
	rating, _ := strconv.ParseFloat(viewDataProduct.AggregateRating.RatingValue)

	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           strconv.Format(viewDataProduct.Sku),
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
		},
		BrandName: brand,
		Title:     viewDataDetail.Name,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		Stats: &pbItem.Stats{
			ReviewCount: int32(reviews),
			Rating:      float32(rating),
		},
	}

	// desc
	description := viewDataDetail.ProductDetails.Description + viewDataDetail.ProductDetails.Features
	item.Description = TrimSpaceNewlineInString(htmlTrimRegp.ReplaceAllString(description, ``))

	//images
	var medias []*pbMedia.Media
	for j, mid := range viewDataDetail.Images {
		imgurl := strings.ReplaceAll(mid.RiasURL, `\/`, `/`)
		if !strings.HasPrefix(mid.RiasURL, `http`) {
			imgurl = "https:" + imgurl
		}
		imgurl = strings.ReplaceAll(strings.ReplaceAll(imgurl, `{quality}`, `85`), `{extension}`, `webp`)

		medias = append(medias, pbMedia.NewImageMedia(
			strconv.Format(j),
			strings.ReplaceAll(imgurl, `{width}`, `600`),
			strings.ReplaceAll(imgurl, `{width}`, `1000`),
			strings.ReplaceAll(imgurl, `{width}`, `800`),
			strings.ReplaceAll(imgurl, `{width}`, `500`),
			"", j == 0))
	}

	for i, breadcrumb := range strings.Split(viewDataDetail.CategoryPath, `/`) {
		breadcrumb := strings.TrimSuffix(breadcrumb, `\`)

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

	currentPrice, _ := strconv.ParsePrice(viewDataDetail.Pricing.MinAsMoney.Value)
	msrp, _ := strconv.ParsePrice(viewDataDetail.Pricing.MaxRetailAsMoney.Value)

	if msrp == 0 {
		msrp = currentPrice
	}
	discount := 0
	if msrp > currentPrice {
		discount = (int)(((msrp - currentPrice) / msrp) * 100)
	}

	for _, rawSku := range viewDataVariation {

		sku := pbItem.Sku{
			SourceId: fmt.Sprintf(rawSku.Ean13),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(currentPrice * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},
			Medias: medias,
			Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}

		if rawSku.HasStock {
			sku.Stock.StockStatus = pbItem.Stock_InStock
			item.Stock.StockStatus = pbItem.Stock_InStock
		}

		// color
		if viewDataDetail.ColorName != "" {
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    viewDataDetail.ColorName,
				Name:  viewDataDetail.ColorName,
				Value: viewDataDetail.ColorName,
			})
		}

		// size
		if rawSku.Size != "" {
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    rawSku.Size,
				Name:  rawSku.Size,
				Value: rawSku.Size,
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

		item.SkuItems = append(item.SkuItems, &sku)
	}

	if len(viewDataVariation) == 0 {

		sku := pbItem.Sku{
			SourceId: fmt.Sprintf(viewDataProduct.Sku),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(currentPrice * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},
			Medias: medias,
			Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}

		if viewDataDetail.AvailableToPurchase {
			sku.Stock.StockStatus = pbItem.Stock_InStock
			item.Stock.StockStatus = pbItem.Stock_InStock
		}

		// color
		if viewDataDetail.ColorName != "" {
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    viewDataDetail.ColorName,
				Name:  viewDataDetail.ColorName,
				Value: viewDataDetail.ColorName,
			})
		} else {
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    "-",
				Name:  "-",
				Value: "-",
			})
		}

	}

	// yield item result
	if err = yield(ctx, &item); err != nil {
		c.logger.Errorf("yield sub request failed, error=%s", err)
		return err
	}

	// other products
	if ctx.Value("groupId") == nil {
		nctx := context.WithValue(ctx, "groupId", item.GetSource().GetId())
		for _, colorSizeOption := range viewDataDetail.Siblings {
			if colorSizeOption.Reference == viewDataDetail.Reference {
				continue
			}
			nextProductUrl, _ := c.CanonicalUrl(strings.ReplaceAll(colorSizeOption.URL, `\/`, `/`))

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

	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("referer", referer)

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
		//"https://www.hunterboots.com/us/en_us",
		//"https://www.hunterboots.com/us/en_us/womens-footwear-rainboots",
		// "https://www.hunterboots.com/us/en_us/womens-ankle-boots",
		//"https://www.hunterboots.com/us/en_us/womens-ankle-boots/womens-original-chelsea-boots/yellow/6503",
		//"https://www.hunterboots.com/us/en_us/sale-womens-sale-footwear",
		//"https://www.hunterboots.com/us/en_us/mens-winter-footwear/mens-insulated-roll-top-sherpa-boots/black/7226",
		"https://www.hunterboots.com/us/en_us/kids-rain-boots/kids-first-classic-nebula-rain-boots/blue/7249",
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
