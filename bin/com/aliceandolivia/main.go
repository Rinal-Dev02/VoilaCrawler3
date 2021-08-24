package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
		categoryPathMatcher: regexp.MustCompile(`^(/([/A-Za-z0-9_-]+))|(/on/demandware\.store([/A-Za-z0-9_-]+))$`),
		productPathMatcher:  regexp.MustCompile(`^(/[/A-Za-z0-9_-]+.html)`),
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
		MustCookies: []*http.Cookie{
			{Name: "bfx.country", Value: "US", Path: "/"},
			{Name: "bfx.currency", Value: "USD", Path: "/"},
			{Name: "bfx.env", Value: "PROD", Path: "/"},
		},
	}

	return opts
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.aliceandolivia.com"}
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
		u.Host = "www.aliceandolivia.com"
	}
	if c.productPathMatcher.MatchString(u.Path) {
		u.RawQuery = ""
		return u.String(), nil
	} else if c.categoryPathMatcher.MatchString(u.Path) {
		// 这里没找到方法直接获取c=US的链接，手动修改
		vals := u.Query()
		vals.Set("c", "US")
		u.RawQuery = vals.Encode()
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
		//return c.parseCategories(ctx, resp, yield)
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
	req, _ := http.NewRequest(http.MethodGet, "https://www.aliceandolivia.com/", nil)
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
		sel := dom.Find(`.menu-group>ul>li`)
		for i := range sel.Nodes {
			node := sel.Eq(i)
			cateName := strings.TrimSpace(node.Find(`a>h2`).First().Text())

			if cateName == "" {
				continue
			}

			//nctx := context.WithValue(ctx, "Category", cateName)

			subSel := node.Find(`.dropdown-item.dropdown`)

			for k := range subSel.Nodes {
				subNode2 := subSel.Eq(k)

				subcat2 := strings.TrimSpace(subNode2.Find(`.nav-item-l2.gtm-nav-title`).First().Text())

				//nnctx := context.WithValue(nctx, "SubCategory", subcat2)

				subNode2list := subNode2.Find(`.dropdown-item`)
				for j := range subNode2list.Nodes {
					subNode3 := subNode2list.Eq(j)
					subcat3 := strings.TrimSpace(subNode3.Find(`.nav-item-l3.gtm-nav-title`).First().Text())
					if subcat3 == "" {
						continue
					}

					href := subNode3.Find(`a`).AttrOr("href", "")
					if href == "" {
						continue
					}
					href, err := c.CanonicalUrl(href)
					if err != nil {
						c.logger.Errorf("got invalid url %s", href)
						continue
					}
					u, _ := url.Parse(href)

					subCateName := strings.TrimSpace(subNode3.Text())

					if c.categoryPathMatcher.MatchString(u.Path) {
						if err := yield([]string{cateName, subcat2, subCateName}, href); err != nil {
							return err
						}
					}
				}

				if len(subNode2list.Nodes) == 0 {
					href := subNode2.Find(`a`).AttrOr("href", "")
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
						if err := yield([]string{cateName, subcat2}, href); err != nil {
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

	sel := dom.Find(`.menu-group>ul>li`)

	for i := range sel.Nodes {
		node := sel.Eq(i)
		cateName := strings.TrimSpace(node.Find(`a>h2`).First().Text())

		if cateName == "" {
			continue
		}

		nctx := context.WithValue(ctx, "Category", cateName)

		subSel := node.Find(`.dropdown-item.dropdown`)

		for k := range subSel.Nodes {
			subNode2 := subSel.Eq(k)

			subcat2 := strings.TrimSpace(subNode2.Find(`.nav-item-l2.gtm-nav-title`).First().Text())

			nnctx := context.WithValue(nctx, "SubCategory", subcat2)

			subNode2list := subNode2.Find(`.dropdown-item`)
			for j := range subNode2list.Nodes {
				subNode3 := subNode2list.Eq(j)
				subcat3 := strings.TrimSpace(subNode3.Find(`.nav-item-l3.gtm-nav-title`).First().Text())
				if subcat3 == "" {
					continue
				}

				href := subNode3.Find(`a`).AttrOr("href", "")
				if href == "" {
					continue
				}

				u, err := url.Parse(href)
				if err != nil {
					c.logger.Error("parse url %s failed", href)
					continue
				}

				subCateName := strings.TrimSpace(subNode3.Text())

				if c.categoryPathMatcher.MatchString(u.Path) {
					nnnctx := context.WithValue(nnctx, "SubCategory2", subCateName)
					req, _ := http.NewRequest(http.MethodGet, href, nil)
					if err := yield(nnnctx, req); err != nil {
						return err
					}
				}
			}

			if len(subNode2list.Nodes) == 0 {
				href := subNode2.Find(`a`).AttrOr("href", "")
				if href == "" {
					continue
				}

				u, err := url.Parse(href)
				if err != nil {
					c.logger.Error("parse url %s failed", href)
					continue
				}

				if c.categoryPathMatcher.MatchString(u.Path) {
					req, _ := http.NewRequest(http.MethodGet, href, nil)
					if err := yield(nnctx, req); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
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

	sel := doc.Find(`.product-tile`)

	for i := range sel.Nodes {
		node := sel.Eq(i)
		if href, _ := node.Find(`a`).Attr("href"); href != "" {

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

	nextUrl := doc.Find(`.btn.btn-outline-primary.col-12.col-sm-4.button-secondary`).AttrOr("data-url", "")
	if nextUrl == "" {
		return nil
	}
	nextUrl = strings.ReplaceAll(nextUrl, "&sz=24", "&sz=96")

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

var productsReviewExtractReg = regexp.MustCompile(`(?Ums)<script type="application/ld\+json">\s*({.*})\s*</script>`)

// used to trim html labels in description
var htmlTrimRegp = regexp.MustCompile(`</?[^>]+>`)

type parseProductResponse struct {
	Context     string `json:"@context"`
	Type        string `json:"@type"`
	Name        string `json:"name"`
	Image       string `json:"image"`
	Description string `json:"description"`
	Sku         string `json:"sku"`
	Offers      struct {
		Type          string `json:"@type"`
		Price         string `json:"price"`
		URL           string `json:"url"`
		PriceCurrency string `json:"priceCurrency"`
		Availability  string `json:"availability"`
	} `json:"offers"`
}

// 发送接口请求商品信息返回数据
type productVariationResp struct {
	Action      string `json:"action"`
	QueryString string `json:"queryString"`
	Locale      string `json:"locale"`
	Product     struct {
		UUID        string      `json:"uuid"`
		ID          string      `json:"id"`
		ProductName string      `json:"productName"`
		ProductType string      `json:"productType"`
		Brand       interface{} `json:"brand"`
		Price       struct {
			Sales struct {
				Value        int    `json:"value"`
				Currency     string `json:"currency"`
				Formatted    string `json:"formatted"`
				DecimalPrice string `json:"decimalPrice"`
			} `json:"sales"`
			List struct {
				Value        int    `json:"value"`
				Currency     string `json:"currency"`
				Formatted    string `json:"formatted"`
				DecimalPrice string `json:"decimalPrice"`
			} `json:"list"`
			PercentageOff string `json:"percentageOff"`
			HTML          string `json:"html"`
		} `json:"price"`
		Images struct {
			HiRes []struct {
				Alt   string `json:"alt"`
				URL   string `json:"url"`
				Title string `json:"title"`
			} `json:"hi-res"`
			Large []struct {
				Alt   string `json:"alt"`
				URL   string `json:"url"`
				Title string `json:"title"`
			} `json:"large"`
			Medium []struct {
				Alt   string `json:"alt"`
				URL   string `json:"url"`
				Title string `json:"title"`
			} `json:"medium"`
			Small []struct {
				Alt   string `json:"alt"`
				URL   string `json:"url"`
				Title string `json:"title"`
			} `json:"small"`
			Shopthelook []struct {
				Alt   string `json:"alt"`
				URL   string `json:"url"`
				Title string `json:"title"`
			} `json:"shopthelook"`
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
				ID               string `json:"id"`
				Description      string `json:"description"`
				DisplayValue     string `json:"displayValue"`
				Value            string `json:"value"`
				Selected         bool   `json:"selected"`
				Selectable       bool   `json:"selectable"`
				VariationGroupID string `json:"variationGroupID"`
				URL              string `json:"url"`
				Images           struct {
					Swatch []struct {
						Alt   string `json:"alt"`
						URL   string `json:"url"`
						Title string `json:"title"`
					} `json:"swatch"`
				} `json:"images"`
			} `json:"values"`
			ResetURL string `json:"resetUrl,omitempty"`
		} `json:"variationAttributes"`
		LongDescription  string      `json:"longDescription"`
		ShortDescription string      `json:"shortDescription"`
		Rating           float64     `json:"rating"`
		Promotions       interface{} `json:"promotions"`
		Attributes       interface{} `json:"attributes"`
		Availability     struct {
			Messages    []string    `json:"messages"`
			InStockDate interface{} `json:"inStockDate"`
			ClassType   string      `json:"classType"`
		} `json:"availability"`
		Available  bool          `json:"available"`
		Options    []interface{} `json:"options"`
		Quantities []struct {
			Value    string `json:"value"`
			Selected bool   `json:"selected"`
			URL      string `json:"url"`
		} `json:"quantities"`
		SelectedProductURL string      `json:"selectedProductUrl"`
		ReadyToOrder       bool        `json:"readyToOrder"`
		Online             bool        `json:"online"`
		PageTitle          interface{} `json:"pageTitle"`
		PageDescription    interface{} `json:"pageDescription"`
		PageKeywords       interface{} `json:"pageKeywords"`
		PageMetaTags       []struct {
		} `json:"pageMetaTags"`
		Template            interface{} `json:"template"`
		RecommendationItems []struct {
			ProductID string `json:"productID"`
		} `json:"recommendationItems"`
		PrimaryVendor  string      `json:"primaryVendor"`
		IsFinalSale    interface{} `json:"isFinalSale"`
		ThresholdStock interface{} `json:"thresholdStock"`
		IsNew          bool        `json:"isNew"`
		IsBestSeller   interface{} `json:"isBestSeller"`
		ExclusiveBadge struct {
		} `json:"exclusiveBadge"`
		WearWithItAssetID    interface{} `json:"wearWithItAssetId"`
		SizeGuideContent     interface{} `json:"sizeGuideContent"`
		PrintsContentAssetID interface{} `json:"printsContentAssetID"`
		Sreligible           struct {
		} `json:"sreligible"`
		SalesBadge                interface{} `json:"salesBadge"`
		IsGiftCertificate         interface{} `json:"isGiftCertificate"`
		DefaultVariantID          string      `json:"defaultVariantId"`
		HasSalePricebook          bool        `json:"hasSalePricebook"`
		PdpVideoID                interface{} `json:"pdpVideoID"`
		NarvarCategory            interface{} `json:"narvarCategory"`
		HasPromotions             bool        `json:"hasPromotions"`
		VariationsInclude         string      `json:"variationsInclude"`
		AttributesHTML            string      `json:"attributesHtml"`
		PromotionsHTML            string      `json:"promotionsHtml"`
		OptionsHTML               string      `json:"optionsHtml"`
		RecommendationsHTML       string      `json:"recommendationsHtml"`
		SaleBadgingHTML           string      `json:"saleBadgingHtml"`
		ExclusiveBadgingHTML      string      `json:"exclusiveBadgingHtml"`
		DescriptionAndDetailsHTML string      `json:"descriptionAndDetailsHtml"`
	} `json:"product"`
	Resources struct {
		InfoSelectforstock    string `json:"info_selectforstock"`
		AssistiveSelectedText string `json:"assistiveSelectedText"`
	} `json:"resources"`
	AjaxRequest      bool   `json:"ajaxRequest"`
	HasSalePricebook bool   `json:"hasSalePricebook"`
	PageVisited      string `json:"pageVisited"`
	PageType         string `json:"pageType"`
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

	var viewData parseProductResponse
	matched := productsReviewExtractReg.FindSubmatch([]byte(respBody))
	if len(matched) > 1 {
		if err := json.Unmarshal(matched[1], &viewData); err != nil {
			c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
			return err
		}
	}

	canUrl, _ := c.CanonicalUrl(doc.Find(`link[rel="canonical"]`).AttrOr("href", ""))
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}
	brand := doc.Find(`#ao-logo-img`).Text()
	if brand == "" {
		brand = "Alice and Olivia"
	}
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
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
	}

	// desc
	description := viewData.Description + htmlTrimRegp.ReplaceAllString(doc.Find(`#collapsible-details-1`).Text(), " ")
	item.Description = string(TrimSpaceNewlineInString([]byte(description)))

	// 最后计算库存状态
	//if strings.Contains(doc.Find(`.availability product-availability`).AttrOr(`data-available`, ``), "true") {
	//	item.Stock.StockStatus = pbItem.Stock_InStock
	//}

	currentPrice, _ := strconv.ParsePrice(viewData.Offers.Price)
	currentPrice *= 100
	msrp := float64(0)
	// 这里直接获取msrp无法获取，动态Json
	//msrp, _ := strconv.ParsePrice(doc.Find(`.value bfx-price bfx-list-price`).Find(`.price-sales`).AttrOr("content", ""))
	productVariantSel := doc.Find(`.color-attribute`)
	for i := range productVariantSel.Nodes {
		node := productVariantSel.Eq(i)
		productId := node.AttrOr("data-variationgroupid", "")
		if productId != item.GetSource().GetId() {
			continue
		}
		// 发送请求获取商品msrp信息
		productVariantUrl := node.AttrOr("data-url", "")
		if productVariantUrl == "" {
			return fmt.Errorf("get productVariantUrl error, pid=%s", item.GetSource().GetId())
		}
		productVariantUrl = strings.ReplaceAll(productVariantUrl, "&amp;", "&")
		req, err := http.NewRequest(http.MethodGet, productVariantUrl, nil)
		if err != nil {
			c.logger.Error(err)
			return err
		}

		opts := c.CrawlOptions(resp.Request.URL)
		for k := range opts.MustHeader {
			req.Header.Set(k, opts.MustHeader.Get(k))
		}
		req.Header.Set("accept-encoding", "gzip, deflate, br")
		req.Header.Set("accept", "*/*")
		req.Header.Set("Referer", resp.Request.URL.String())
		req.Header.Set("User-Agent", resp.Request.Header.Get("User-Agent"))

		resp, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
			EnableProxy:    true,
			EnableHeadless: c.CrawlOptions(resp.Request.URL).EnableHeadless,
			Reliability:    c.CrawlOptions(resp.Request.URL).Reliability,
		})
		if err != nil {
			c.logger.Error(err)
			return err
		}

		if resp.StatusCode != http.StatusOK {
			c.logger.Errorf("status is %v", resp.StatusCode)
			return fmt.Errorf(resp.Status)
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			c.logger.Error(err)
			return err
		}
		resp.Body.Close()

		var productVariation productVariationResp
		if err := json.Unmarshal(data, &productVariation); err != nil {
			c.logger.Errorf("%s, error=%s", data, err)
			return err
		}
		msrp = float64(productVariation.Product.Price.List.Value * 100)
		break
	}

	if msrp == 0 {
		msrp = currentPrice
	}
	discount := int32(0)
	if msrp > currentPrice {
		discount = int32(((msrp - currentPrice) / msrp) * 100)
	}

	//images
	sel := doc.Find(`.carousel-inner`).Find(`div`)
	for j := range sel.Nodes {
		node := sel.Eq(j)
		imgurl := strings.Split(node.Find(`img`).AttrOr(`data-zoom`, ``), "?")[0]

		item.Medias = append(item.Medias, pbMedia.NewImageMedia(
			strconv.Format(j),
			imgurl,
			imgurl+"?sw=800&sh=800&q=80",
			imgurl+"?sw=500&sh=500&q=80",
			imgurl+"?sw=300&sh=300&q=80",
			"", j == 0))
	}

	// itemListElement
	sel = doc.Find(`.breadcrumb>li`)
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

	// Color
	cid := ""
	colorName := ""
	var colorSelected *pbItem.SkuSpecOption
	sel = doc.Find(`.attribute.color-swatch-wrapper`).Find(`button`)
	for i := range sel.Nodes {
		node := sel.Eq(i)

		if strings.Contains(node.Find(`.swatch-value`).AttrOr(`class`, ``), `selected`) {
			cid = node.AttrOr(`data-variationgroupid`, "")
			icon := strings.ReplaceAll(strings.ReplaceAll(node.Find(`.swatch-value`).AttrOr(`style`, ""), "background-image: url(", ""), ")", "")
			colorName = node.Find(`.swatch-value`).AttrOr(`data-attr-displayvalue`, "")
			colorSelected = &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    colorName,
				Name:  colorName,
				Value: colorName,
				Icon:  icon,
			}
		}
	}

	sel = doc.Find(`.size-selections`).Find(`button`)
	for i := range sel.Nodes {
		node := sel.Eq(i)

		sid := node.AttrOr("data-attr-value", "")
		if sid == "" {
			continue
		}

		sku := pbItem.Sku{
			SourceId: fmt.Sprintf("%s-%s", cid, sid),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(currentPrice),
				Msrp:     int32(msrp),
				Discount: discount,
			},
			Stock: &pbItem.Stock{StockStatus: pbItem.Stock_InStock},
		}

		if spanClass := node.Find(`span`).AttrOr("class", ""); strings.Contains(spanClass, "unselectable") {
			sku.Stock.StockStatus = pbItem.Stock_OutOfStock
		} else if strings.Contains(spanClass, "selectable") {
			sku.Stock.StockStatus = pbItem.Stock_InStock
			if item.Stock.StockStatus == pbItem.Stock_OutOfStock {
				item.Stock.StockStatus = pbItem.Stock_InStock
			}
		}

		if colorSelected != nil {
			sku.Specs = append(sku.Specs, colorSelected)
		}

		// size
		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Id:    sid,
			Name:  sid,
			Value: sid,
		})

		item.SkuItems = append(item.SkuItems, &sku)
	}
	// 计算商品库存状态

	// 不同的颜色属于不同的商品，发送新的请求
	productSel := doc.Find(`.color-attribute`)
	for i, _ := range productSel.Nodes {
		productNode := productSel.Eq(i)
		productId := productNode.AttrOr(`data-variationgroupid`, "")
		if productId == item.GetSource().GetId() {
			continue
		}
		otherProductUrl := strings.ReplaceAll(item.GetSource().GetCanonicalUrl(), item.GetSource().GetId(), productId)

		req, _ := http.NewRequest(http.MethodGet, otherProductUrl, nil)
		if err := yield(ctx, req); err != nil {
			c.logger.Errorf("yield sub request failed, error=%s", err)
			return err
		}
	}

	// yield item result
	return yield(ctx, &item)
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://www.aliceandolivia.com/",
		//"https://www.aliceandolivia.com/womens-clothing/dresses/",
		//	"https://www.aliceandolivia.com/womens-clothing/dresses/mini-dresses/?c=US",
		//"https://www.aliceandolivia.com/ciara-crewneck-cropped-pullover/192772440624.html",
		"https://www.aliceandolivia.com/rianna-puff-sleeve-crop-top/CC103T20022G618.html",
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
