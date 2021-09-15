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
	httpClient             http.Client
	categoryPathMatcher    *regexp.Regexp
	categoryAPIPathMatcher *regexp.Regexp
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
		categoryPathMatcher:    regexp.MustCompile(`^/([/A-Za-z0-9_-]+)$`),
		categoryAPIPathMatcher: regexp.MustCompile(`^/category-search-ajax$`),
		productPathMatcher:     regexp.MustCompile(`^/([/A-Za-z0-9_-]+)\d+$`),
		logger:                 logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "740a70f485afb9d035f01c0e86be529c"
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
	return []string{"*.aldoshoes.com"}
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
		u.Host = "www.aldoshoes.com"
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
	} else if c.categoryPathMatcher.MatchString(p) || c.categoryAPIPathMatcher.MatchString(p) {
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

	rootUrl := "https://www.aldoshoes.com/us/en_US"
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

		sel := dom.Find(`.c-navigation.c-navigation--primary>ul>li`)

		c.logger.Val("sel1.Nodes", len(sel.Nodes))
		for a := range sel.Nodes {
			node := sel.Eq(a)

			catname := strings.TrimSpace(node.Find(`span`).First().Text())
			if catname == "" {
				continue
			}

			sublvl1div := node.Find(`li[class="u-hide@md-mid"]`)

			for b := range sublvl1div.Nodes {
				sublvl1 := sublvl1div.Eq(b)
				sublvl1name := strings.TrimSpace(sublvl1.Find(`span`).First().Text())

				sublvl2 := sublvl1.Find(`li`)
				for k := range sublvl2.Nodes {
					selsublvl3 := sublvl2.Eq(k)
					sublvl2name := strings.TrimSpace(selsublvl3.Find(`a`).First().Text())

					href := selsublvl3.Find(`a`).AttrOr("href", "")
					if href == "" || sublvl2name == "" {
						continue
					}

					canonicalHref, err := c.CanonicalUrl(href)
					if err != nil {
						c.logger.Errorf("got invalid url %s", href)
						continue
					}

					u, _ := url.Parse(canonicalHref)

					if c.categoryPathMatcher.MatchString(u.Path) {
						if err := yield([]string{catname, sublvl1name, sublvl2name}, canonicalHref); err != nil {
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

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	lastIndex := nextIndex(ctx)

	sel := dom.Find(`.c-product-tile__link-product`)
	for i := range sel.Nodes {

		node := sel.Eq(i)
		href := node.AttrOr("href", "")
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

	nextUrl := dom.Find(`link[rel="next"]`).AttrOr(`href`, ``)
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
	resp = bytes.ReplaceAll(resp, []byte("  "), []byte(" "))
	return resp
}

var imgReg = regexp.MustCompile(`_\d+x\d+.jpg`)

// used to trim html labels in description
var htmlTrimRegp = regexp.MustCompile(`</?[^>]+>`)
var prodDataExtraReg = regexp.MustCompile(`(?Ums)window.__INITIAL_STATE__\s*=\s*({.*});`)

type parseProductDetail struct {
	ProductDetails struct {
		Product struct {
			AvailableForPickup bool `json:"availableForPickup"`
			BaseOptions        []struct {
				Options []struct {
					ApprovalStatus      string   `json:"approvalStatus"`
					Code                string   `json:"code"`
					ColourGroup         string   `json:"colourGroup"`
					ComingSoon          bool     `json:"comingSoon"`
					DirectCategoryCodes []string `json:"directCategoryCodes"`
					PriceData           struct {
						CurrencyIso        string  `json:"currencyIso"`
						FormattedValue     string  `json:"formattedValue"`
						ListPrice          float64 `json:"listPrice"`
						ListPriceFormatted string  `json:"listPriceFormatted"`
						PriceType          string  `json:"priceType"`
						Value              float64 `json:"value"`
					} `json:"priceData"`
					Repurchasable bool `json:"repurchasable"`
					Stock         struct {
						StockLevel       int    `json:"stockLevel"`
						StockLevelStatus string `json:"stockLevelStatus"`
					} `json:"stock"`
					URL                     string `json:"url"`
					VariantOptionQualifiers []struct {
						Img struct {
							AltText string `json:"altText"`
							Types   []struct {
								Backgrounds []struct {
									Background  string `json:"background"`
									Proportions []struct {
										Proportion string `json:"proportion"`
										Dimensions []struct {
											H string `json:"h"`
											W string `json:"w"`
										} `json:"dimensions"`
									} `json:"proportions"`
								} `json:"backgrounds"`
								Type        string `json:"type"`
								Code        string `json:"code"`
								SrcTemplate string `json:"srcTemplate,omitempty"`
							} `json:"types"`
							SrcTemplate string `json:"srcTemplate"`
						} `json:"img"`
						Name      string `json:"name"`
						Qualifier string `json:"qualifier"`
						Value     string `json:"value"`
					} `json:"variantOptionQualifiers"`
				} `json:"options"`
				Selected struct {
					ApprovalStatus      string   `json:"approvalStatus"`
					Code                string   `json:"code"`
					ColourGroup         string   `json:"colourGroup"`
					ComingSoon          bool     `json:"comingSoon"`
					DirectCategoryCodes []string `json:"directCategoryCodes"`
					PriceData           struct {
						CurrencyIso        string  `json:"currencyIso"`
						FormattedValue     string  `json:"formattedValue"`
						ListPrice          float64 `json:"listPrice"`
						ListPriceFormatted string  `json:"listPriceFormatted"`
						PriceType          string  `json:"priceType"`
						Value              float64 `json:"value"`
					} `json:"priceData"`
					Repurchasable bool `json:"repurchasable"`
					Stock         struct {
						StockLevel       int    `json:"stockLevel"`
						StockLevelStatus string `json:"stockLevelStatus"`
					} `json:"stock"`
					URL                     string `json:"url"`
					VariantOptionQualifiers []struct {
						Img struct {
							AltText string `json:"altText"`
							Types   []struct {
								Backgrounds []struct {
									Background  string `json:"background"`
									Proportions []struct {
										Proportion string `json:"proportion"`
										Dimensions []struct {
											H string `json:"h"`
											W string `json:"w"`
										} `json:"dimensions"`
									} `json:"proportions"`
								} `json:"backgrounds"`
								Type        string `json:"type"`
								Code        string `json:"code"`
								SrcTemplate string `json:"srcTemplate,omitempty"`
							} `json:"types"`
							SrcTemplate string `json:"srcTemplate"`
						} `json:"img"`
						Name      string `json:"name"`
						Qualifier string `json:"qualifier"`
						Value     string `json:"value"`
					} `json:"variantOptionQualifiers"`
				} `json:"selected"`
				VariantType string `json:"variantType"`
			} `json:"baseOptions"`
			BaseProduct string `json:"baseProduct"`
			Categories  []struct {
				ClearanceCategory           bool   `json:"clearanceCategory"`
				Code                        string `json:"code"`
				Gallery                     bool   `json:"gallery"`
				NewArrivalCategory          bool   `json:"newArrivalCategory"`
				OutletCategory              bool   `json:"outletCategory"`
				SaleCategory                bool   `json:"saleCategory"`
				URL                         string `json:"url"`
				VisualSequencerTemplateView struct {
					TemplateView string `json:"templateView"`
				} `json:"visualSequencerTemplateView"`
			} `json:"categories"`
			Clearance       bool   `json:"clearance"`
			Code            string `json:"code"`
			ColourGroupName string `json:"colourGroupName"`
			ComingSoon      bool   `json:"comingSoon"`
			Description     string `json:"description"`
			Gender          string `json:"gender"`
			GiftCard        bool   `json:"giftCard"`
			Image           struct {
				AltText string `json:"altText"`
				Types   []struct {
					Backgrounds []struct {
						Background  string `json:"background"`
						Proportions []struct {
							Proportion string `json:"proportion"`
							Dimensions []struct {
								H string `json:"h"`
								W string `json:"w"`
							} `json:"dimensions"`
						} `json:"proportions"`
					} `json:"backgrounds"`
					Type        string `json:"type"`
					Code        string `json:"code"`
					SrcTemplate string `json:"srcTemplate,omitempty"`
				} `json:"types"`
				SrcTemplate string `json:"srcTemplate"`
			} `json:"image"`
			KidsGroup string `json:"kidsGroup"`
			MetaData  []struct {
				Content string `json:"content"`
				Name    string `json:"name"`
			} `json:"metaData"`
			Name            string `json:"name"`
			NewArrival      bool   `json:"newArrival"`
			NonReturnable   bool   `json:"nonReturnable"`
			OnlineExclusive bool   `json:"onlineExclusive"`
			Outlet          bool   `json:"outlet"`
			PreOrder        bool   `json:"preOrder"`
			Price           struct {
				CurrencyIso    string  `json:"currencyIso"`
				FormattedValue string  `json:"formattedValue"`
				PriceType      string  `json:"priceType"`
				Value          float64 `json:"value"`
			} `json:"price"`
			PriceRange struct {
				MaxPrice struct {
					CurrencyIso    string  `json:"currencyIso"`
					FormattedValue string  `json:"formattedValue"`
					PriceType      string  `json:"priceType"`
					Value          float64 `json:"value"`
				} `json:"maxPrice"`
			} `json:"priceRange"`
			ProductMessage struct {
			} `json:"productMessage"`
			Purchasable   bool   `json:"purchasable"`
			Repurchasable bool   `json:"repurchasable"`
			Sale          bool   `json:"sale"`
			Scarce        bool   `json:"scarce"`
			SizeChartID   string `json:"sizeChartID"`
			Stock         struct {
				StockLevel       int    `json:"stockLevel"`
				StockLevelStatus string `json:"stockLevelStatus"`
			} `json:"stock"`
			URL            string `json:"url"`
			VariantOptions []struct {
				ApprovalStatus string `json:"approvalStatus"`
				Code           string `json:"code"`
				ColourGroup    string `json:"colourGroup"`
				ComingSoon     bool   `json:"comingSoon"`
				PriceData      struct {
					CurrencyIso        string  `json:"currencyIso"`
					FormattedValue     string  `json:"formattedValue"`
					ListPrice          float64 `json:"listPrice"`
					ListPriceFormatted string  `json:"listPriceFormatted"`
					PriceType          string  `json:"priceType"`
					Value              float64 `json:"value"`
				} `json:"priceData"`
				Repurchasable bool `json:"repurchasable"`
				Stock         struct {
					StockLevel       int    `json:"stockLevel"`
					StockLevelStatus string `json:"stockLevelStatus"`
				} `json:"stock"`
				URL                     string `json:"url"`
				VariantOptionQualifiers []struct {
					Img struct {
						AltText string `json:"altText"`
						Options []struct {
							Backgrounds []struct {
								Background string `json:"background"`
								Dimension  []struct {
									H string `json:"h"`
									W string `json:"w"`
								} `json:"dimension"`
							} `json:"backgrounds,omitempty"`
							Code      string `json:"code"`
							Type      string `json:"type"`
							Dimension []struct {
								H string `json:"h"`
								W string `json:"w"`
							} `json:"dimension,omitempty"`
							SrcTemplate string `json:"srcTemplate,omitempty"`
						} `json:"options"`
						SrcTemplate string `json:"srcTemplate"`
					} `json:"img"`
					Name      string `json:"name"`
					Qualifier string `json:"qualifier"`
					Value     string `json:"value"`
				} `json:"variantOptionQualifiers"`
			} `json:"variantOptions"`
			SocialImages         []interface{} `json:"socialImages"`
			SizeChart            interface{}   `json:"sizeChart"`
			IsBrandDetailEnabled bool          `json:"isBrandDetailEnabled"`
		} `json:"product"`
	} `json:"productDetails"`
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

	// Parse json
	matched := prodDataExtraReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		return fmt.Errorf("extract json from product page %s failed", resp.Request.URL)
	}
	matched[1] = bytes.ReplaceAll(matched[1], []byte(`:undefined`), []byte(`:"undefined"`))

	var viewData parseProductDetail
	if err = json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Debugf("parse %s failed, error=%s", matched[1], err)
		return err
	}

	canUrl, _ := c.CanonicalUrl(doc.Find(`link[rel="canonical"]`).AttrOr("href", ""))
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}

	brand := "ALDO"

	s := strings.Split(resp.Request.URL.Path, `/`)
	pid := s[len(s)-1]

	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           pid,
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
		},
		BrandName: brand,
		Title:     doc.Find(`.c-buy-module__product-title`).Text(),
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_InStock},
	}

	// desc
	description := htmlTrimRegp.ReplaceAllString(doc.Find(`.c-product-description`).First().Text(), " ")
	item.Description = string(TrimSpaceNewlineInString([]byte(description)))

	msrp, _ := strconv.ParsePrice(doc.Find(`.c-product-price__formatted-price--original`).Text())
	currentPrice, _ := strconv.ParsePrice(doc.Find(`.c-product-price__formatted-price--is-reduced`).Text())

	if len(doc.Find(`.c-product-price__formatted-price--is-reduced`).Nodes) == 0 {
		currentPrice, _ = strconv.ParsePrice(doc.Find(`.c-product-price__formatted-price`).Text())
	}

	if msrp == 0 {
		msrp = currentPrice
	}
	discount := 0
	if msrp > currentPrice {
		discount = (int)(((msrp - currentPrice) / msrp) * 100)
	}

	//images
	// sel := doc.Find(`.c-carousel__indicator`)
	counter := -1
	sel := doc.Find(`.c-product-detail__carousel-wrapper`).Find(`.c-picture`)
	imgMap := make(map[string]bool)
	for j := range sel.Nodes {
		node := sel.Eq(j)

		imgurl := imgReg.ReplaceAllString(strings.Split(strings.TrimSpace(node.Find(`img`).AttrOr(`data-srcset`, ``)), " ")[0], "")
		if imgurl == "" {
			imgurl = imgReg.ReplaceAllString(strings.Split(strings.TrimSpace(node.Find(`img`).AttrOr(`srcset`, ``)), " ")[0], "")
		}

		if imgurl == "" {
			continue
		} else if strings.HasPrefix(imgurl, `//`) {
			imgurl = "https:" + imgurl
		}
		counter++
		if imgMap[imgurl] {
			continue
		} else {
			imgMap[imgurl] = true
		}
		item.Medias = append(item.Medias, pbMedia.NewImageMedia(
			strconv.Format(j),
			imgurl+"_600x600.jpg",
			imgurl+"_1000x1000.jpg",
			imgurl+"_800x800.jpg",
			imgurl+"_600x600.jpg",
			"", counter == 0))
	}

	// itemListElement
	sel = doc.Find(`.c-breadcrumb__list`).First().Find(`li`)
	c.logger.Debugf("nodes %d", len(sel.Nodes))
	for i := range sel.Nodes {
		node := sel.Eq(i)
		breadcrumb := strings.TrimSpace(node.Find(`a`).First().Text())

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

	// Color
	colorName := ""
	var colorSelected *pbItem.SkuSpecOption
	sel = doc.Find(`.o-style-option__list-item`)

	for i := range sel.Nodes {
		node := sel.Eq(i)

		if strings.Contains(node.Find(`a`).AttrOr(`class`, ``), `o-style-option__option--is-checked`) {

			icon := strings.Split(node.Find(`img`).AttrOr(`data-srcset`, ""), " ")[0]
			if icon != "" && strings.HasPrefix(icon, `//`) {
				icon = "https:" + icon
			}

			colorName = node.Find(`a`).AttrOr(`title`, ``)
			colorSelected = &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    colorName,
				Name:  colorName,
				Value: colorName,
				Icon:  icon,
			}
		}
	}

	sizeIndex := -1
	//colorIndex := -1

	for _, key := range viewData.ProductDetails.Product.VariantOptions {
		for i, keyV := range key.VariantOptionQualifiers {
			if keyV.Name == "Colour" || keyV.Name == "Color" {
				//colorIndex = i
			} else if keyV.Name == "Size" {
				sizeIndex = i
				break
			}
		}

	}

	//sel = doc.Find(`.c-product-option__list--size`).Find(`li`)
	for _, rawSku := range viewData.ProductDetails.Product.VariantOptions {

		sku := pbItem.Sku{
			SourceId: rawSku.Code,
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(currentPrice * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},
			Medias: item.Medias,
			Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}

		if rawSku.Stock.StockLevel > 0 {
			sku.Stock.StockStatus = pbItem.Stock_InStock
		}

		if colorSelected != nil {
			sku.Specs = append(sku.Specs, colorSelected)
		}

		// size
		if sizeIndex > -1 {
			if rawSku.VariantOptionQualifiers[sizeIndex].Value != "" && rawSku.VariantOptionQualifiers[sizeIndex].Name == "Size" {
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecSize,
					Id:    rawSku.VariantOptionQualifiers[sizeIndex].Value,
					Name:  rawSku.VariantOptionQualifiers[sizeIndex].Value,
					Value: rawSku.VariantOptionQualifiers[sizeIndex].Value,
				})
			}
		}

		// In this site, the skuid is unique
		//for _, spec := range sku.Specs {
		//	sku.SourceId += fmt.Sprintf("-%s", spec.Id)
		//}
		item.SkuItems = append(item.SkuItems, &sku)
	}

	if len(viewData.ProductDetails.Product.VariantOptions) == 0 {

		sku := pbItem.Sku{
			SourceId: pid,
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(currentPrice * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},
			Medias: item.Medias,
			Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_InStock},
		}

		if strings.Contains(doc.Find(`.c-purchase-buttons__button--add-to-bag`).AttrOr("aria-disabled", ""), "true") {
			sku.Stock.StockStatus = pbItem.Stock_OutOfStock
		}

		if colorSelected != nil {
			sku.Specs = append(sku.Specs, colorSelected)
		} else {
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    "-",
				Name:  "-",
				Value: "-",
			})
		}
		// In this site, the skuid is unique
		//for _, spec := range sku.Specs {
		//	sku.SourceId += fmt.Sprintf("-%s", spec.Id)
		//}
		item.SkuItems = append(item.SkuItems, &sku)
	}

	for _, spec := range item.SkuItems {
		if spec.Stock.StockStatus == pbItem.Stock_InStock {
			item.Stock.StockStatus = pbItem.Stock_InStock
			break
		}
	}

	// yield item result
	if err = yield(ctx, &item); err != nil {
		c.logger.Errorf("yield sub request failed, error=%s", err)
		return err
	}

	///found other color
	if ctx.Value("groupId") == nil {
		sel = doc.Find(`#PdpProductColorSelectorOpts>li`)
		for i := range sel.Nodes {
			nctx := context.WithValue(ctx, "groupId", item.GetSource().GetId())
			node := sel.Eq(i)
			//color := strings.TrimSpace(node.Find(`a`).AttrOr("title", ""))
			if !strings.Contains(node.Find(`a`).AttrOr(`class`, ``), `o-style-option__option--is-checked`) {

				nextProductUrl, err := c.CanonicalUrl(node.Find(`a`).AttrOr("href", ""))

				if nextProductUrl == "" || err != nil {
					continue
				}

				if req, err := http.NewRequest(http.MethodGet, nextProductUrl, nil); err != nil {
					c.logger.Error(err)
					continue
				} else if err := yield(nctx, req); err != nil {
					c.logger.Error(err)
					return err
				}
			}
		}
	}

	return nil
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://www.aldoshoes.com/us/en_US",
		//"https://www.aldoshoes.com/us/en_US/women/new-arrivals/footwear",
		//"https://www.aldoshoes.com/aldo-accessories-women/aldo-accessories-women",
		//"https://www.aldoshoes.com/us/en_US/etealia-light-purple/p/13189117",
		//"https://www.aldoshoes.com/us/en_US/gleawia-white/p/13189060",
		//"https://www.aldoshoes.com/us/en_US/women/footwear/boots",
		//"https://www.aldoshoes.com/us/en_US/women/lilya-bone/p/13265499",
		//"https://www.aldoshoes.com/us/en_US/women/unilax-white/p/13265491",
		//"https://www.aldoshoes.com/us/en_US/women/grevillea-light-pink/p/13087793",
		"https://www.aldoshoes.com/us/en_US/women-s-thermal-insoles-no-color/p/12652382",
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
