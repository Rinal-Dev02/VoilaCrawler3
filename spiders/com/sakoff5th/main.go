package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"math"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/voiladev/VoilaCrawl/pkg/crawler"
	"github.com/voiladev/VoilaCrawl/pkg/net/http"
	"github.com/voiladev/VoilaCrawl/pkg/net/http/cookiejar"
	"github.com/voiladev/VoilaCrawl/pkg/proxy"
	pbMedia "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	pbProxy "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// _Crawler defined the crawler struct/class for which is not necessory to be exportable
type _Crawler struct {
	// httpClient is the object of an http client
	httpClient                 http.Client
	categoryPathMatcher        *regexp.Regexp
	categoryDynamicLoadMatcher *regexp.Regexp
	productPathMatcher         *regexp.Regexp
	// logger is the log tool
	logger glog.Log
}

// New returns an object of interface crawler.Crawler.
// this is the entry of the spider plugin. the plugin manager will call this func to init the plugin.
// view pkg/crawler/spec.go to know more about the interface `Crawler`
func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient:                 client,
		categoryPathMatcher:        regexp.MustCompile(`^/c(/[a-z0-9\pL\pS\-.]+){1,6}$`),
		categoryDynamicLoadMatcher: regexp.MustCompile(`^/on/demandware.store/Sites-SaksOff5th-Site/en_US/Search-UpdateGrid$`),
		productPathMatcher:         regexp.MustCompile(`^/product(/[a-z0-9\pL\pS\-.]+){1,4}\-\d+.html$`),
		logger:                     logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "a233105f4d384fa2bcf56131653bac56"
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
		Reliability:       pbProxy.ProxyReliability_ReliabilityDefault,
	}

	opts.MustCookies = append(opts.MustCookies,
		&http.Cookie{Name: "E4X_CURRENCY", Value: "USD", Path: "/"},
		&http.Cookie{Name: "bfx.isWelcomed", Value: "true", Path: "/"},
		&http.Cookie{Name: "bfx.language", Value: "en", Path: "/"},
		&http.Cookie{Name: "bfx.country", Value: "US", Path: "/"},
		&http.Cookie{Name: "bfx.currency", Value: "USD", Path: "/"},
		&http.Cookie{Name: "bfx.isInternational", Value: "false", Path: "/"},
		&http.Cookie{Name: "s_cc", Value: "false", Path: "/"},
	)
	return opts
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.saksoff5th.com"}
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

	if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) || c.categoryDynamicLoadMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	} else if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct2(ctx, resp, yield)
	}
	return fmt.Errorf("unsupported url %s", resp.Request.URL.String())
}

// nextIndex used to get the index from the shared data.
// item.index is a const key for item index.
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")))
}

// used to extract embaded json data in website page.
// more about golang regulation see here https://golang.org/pkg/regexp/syntax/
var productsExtractOtherDetailReg = regexp.MustCompile(`({.*})`)

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

	if !bytes.Contains(respBody, []byte("product bfx-disable-product standard")) {
		return errors.New("products not found")
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	lastIndex := nextIndex(ctx)

	sel := doc.Find(`.image-container`)
	c.logger.Debugf("nodes %d", len(sel.Nodes))
	for i := range sel.Nodes {
		node := sel.Eq(i)
		href := node.Find(`a.thumb-link`).AttrOr("href", "")
		if href == "" {
			href = node.Find(`.tile-body .pdp-link>.link`).AttrOr("href", "")
		}
		if href == "" {
			continue
		}

		req, err := http.NewRequest(http.MethodGet, href, nil)
		if err != nil {
			c.logger.Error(err)
			continue
		}

		nctx := context.WithValue(ctx, "item.index", lastIndex)
		lastIndex += 1

		if err := yield(nctx, req); err != nil {
			return err
		}
	}

	nextUrl := html.UnescapeString(doc.Find(`div.show-more>div>button`).AttrOr("data-url", ""))
	if nextUrl == "" {
		nextUrl = html.UnescapeString(doc.Find(`.pagination-container .page-wrapper .page-item.next>a`).AttrOr("href", ""))
		if nextUrl == "" || strings.ToLower(nextUrl) == "null" {
			return nil
		}
	}

	if nextUrl != "" {
		nctx := context.WithValue(ctx, "item.index", lastIndex)
		req, _ := http.NewRequest(http.MethodGet, nextUrl, nil)
		vals := req.URL.Query()
		req.URL.RawQuery = vals.Encode()
		return yield(nctx, req)
	}
	return nil
}

// used to trim html labels in description
var htmlTrimRegp = regexp.MustCompile(`</?[^>]+>`)

type productVariationAttributes struct {
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
		// Images       struct {
		// 	Swatch []struct {
		// 		Alt      string `json:"alt"`
		// 		URL      string `json:"url"`
		// 		Title    string `json:"title"`
		// 		HiresURL string `json:"hiresURL"`
		// 	} `json:"swatch"`
		// } `json:"images"`
	} `json:"values"`
	SelectedAttribute struct {
	} `json:"selectedAttribute"`
	AttrDisplay            string `json:"attrDisplay"`
	AttrEditDisplay        string `json:"attrEditDisplay"`
	SelectedSizeClass      string `json:"selectedSizeClass"`
	AttributeSelectedValue string `json:"attributeSelectedValue"`
	ResetURL               string `json:"resetUrl,omitempty"`
}

type ProductDataJson struct {
	Product struct {
		MasterProductID string `json:"masterProductID"`
		Brand           struct {
			Name string `json:"name"`
			URL  struct {
			} `json:"url"`
		} `json:"brand"`
		UUID                 string      `json:"uuid"`
		ID                   string      `json:"id"`
		ProductName          string      `json:"productName"`
		ProductType          string      `json:"productType"`
		Purchaselimit        interface{} `json:"purchaselimit"`
		LongDescriptionStyle string      `json:"longDescriptionStyle"`
		DropShipShipping     struct {
			Name        interface{} `json:"name"`
			DisplayName string      `json:"displayName"`
			URL         struct {
			} `json:"url"`
		} `json:"DropShipShipping"`
		UspsShipOK               interface{}   `json:"uspsShipOK"`
		PdRestrictedShipTypeText interface{}   `json:"pdRestrictedShipTypeText"`
		MRecommendations         []interface{} `json:"mRecommendations"`
		FeaturedType             struct {
			Value interface{} `json:"value"`
			Color string      `json:"color"`
		} `json:"featuredType"`
		IsNotReturnable struct {
			Value bool   `json:"value"`
			Color string `json:"color"`
		} `json:"isNotReturnable"`
		Badge struct {
			IsNew struct {
				Value bool   `json:"value"`
				Color string `json:"color"`
			} `json:"isNew"`
			IsSale struct {
				Value bool   `json:"value"`
				Color string `json:"color"`
			} `json:"isSale"`
			IsClearance          bool   `json:"isClearance"`
			IsFinalSale          bool   `json:"isFinalSale"`
			LimitedInvBadgeColor string `json:"limitedInvBadgeColor"`
		} `json:"badge"`
		DisplayQuicklook  string `json:"displayQuicklook"`
		Wishlist          string `json:"wishlist"`
		SizeChartTemplate string `json:"sizeChartTemplate"`
		PlpPromos         struct {
		} `json:"plpPromos"`
		PdRestrictedWarningText bool   `json:"pdRestrictedWarningText"`
		PdpURL                  string `json:"pdpURL"`
		Price                   struct {
			Sales struct {
				Value              float64 `json:"value"`
				Currency           string  `json:"currency"`
				Formatted          string  `json:"formatted"`
				DecimalPrice       string  `json:"decimalPrice"`
				FormatAmount       string  `json:"formatAmount"`
				PriceBandFormatted string  `json:"priceBandFormatted"`
			} `json:"sales"`
			List struct {
				Value              float64 `json:"value"`
				Currency           string  `json:"currency"`
				Formatted          string  `json:"formatted"`
				DecimalPrice       string  `json:"decimalPrice"`
				FormatAmount       string  `json:"formatAmount"`
				PriceBandFormatted string  `json:"priceBandFormatted"`
			} `json:"list"`
			Savings        float64 `json:"savings"`
			SavePercentage string  `json:"savePercentage"`
			HTML           string  `json:"html"`
		} `json:"price"`
		Images struct {
			// Large []struct {
			// 	Alt      string `json:"alt"`
			// 	URL      string `json:"url"`
			// 	Title    string `json:"title"`
			// 	HiresURL string `json:"hiresURL"`
			// } `json:"large"`
			// Small []struct {
			// 	Alt      string `json:"alt"`
			// 	URL      string `json:"url"`
			// 	Title    string `json:"title"`
			// 	HiresURL string `json:"hiresURL"`
			// } `json:"small"`
			HiRes []struct {
				Alt   string `json:"alt"`
				URL   string `json:"url"`
				Title string `json:"title"`
				// HiresURL string `json:"hiresURL"`
			} `json:"hi-res"`
			// Swatch []struct {
			// 	Alt      string `json:"alt"`
			// 	URL      string `json:"url"`
			// 	Title    string `json:"title"`
			// 	HiresURL string `json:"hiresURL"`
			// } `json:"swatch"`
			// Video []struct {
			// 	Alt      string `json:"alt"`
			// 	URL      string `json:"url"`
			// 	Title    string `json:"title"`
			// 	HiresURL struct {
			// 	} `json:"hiresURL"`
			// } `json:"video"`
		} `json:"images"`
		SelectedQuantity    int                           `json:"selectedQuantity"`
		MinOrderQuantity    int                           `json:"minOrderQuantity"`
		MaxOrderQuantity    int                           `json:"maxOrderQuantity"`
		VariationAttributes []*productVariationAttributes `json:"variationAttributes"`
		LongDescription     string                        `json:"longDescription"`
		ShortDescription    interface{}                   `json:"shortDescription"`
		Rating              float64                       `json:"rating"`
		Promotions          interface{}                   `json:"promotions"`
		Availability        struct {
			Messages                 []interface{} `json:"messages"`
			ButtonName               string        `json:"buttonName"`
			IsInPurchaselimit        bool          `json:"isInPurchaselimit"`
			IsInPurchaselimitMessage string        `json:"isInPurchaselimitMessage"`
			IsAboveThresholdLevel    bool          `json:"isAboveThresholdLevel"`
			Outofstockmessage        string        `json:"outofstockmessage"`
			InStockDate              interface{}   `json:"inStockDate"`
		} `json:"availability"`
		TurntoReviewCount  int `json:"turntoReviewCount"`
		PromotionalPricing struct {
			IsPromotionalPrice bool   `json:"isPromotionalPrice"`
			PromoMessage       string `json:"promoMessage"`
			PriceHTML          string `json:"priceHtml"`
		} `json:"promotionalPricing"`

		AllAvailableProductsSoldOut bool `json:"allAvailableProductsSoldOut"`
		AllAvailableProducts        []struct {
			AvailableDc string `json:"available_dc"`
			Sku         string `json:"sku"`
		} `json:"allAvailableProducts"`
		StarRating float64 `json:"starRating"`
		// AttributesHTML string `json:"attributesHtml"`
		// PromotionsHTML string `json:"promotionsHtml"`
		// FinalSaleHTML  string `json:"finalSaleHtml"`
	} `json:"product"`
}

var productInfoReg = regexp.MustCompile(`(?Ums)<script\s+type="text/javascript">\s*pageDataObj\s*=\s*({.*});\s*</script>`)

// parseProduct do http request for each sku
func (c *_Crawler) parseProduct(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil {
		return nil
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var productSkus struct {
		Products []struct {
			AverageRating string `json:"average_rating"`
			Brand         string `json:"brand"`
			Code          string `json:"code"`
			Name          string `json:"name"`
			OriginalPrice string `json:"original_price"`
			Price         string `json:"price"`
			Skus          []struct {
				AvailableDc string `json:"available_dc"`
				Sku         string `json:"sku"`
			} `json:"skus"`
			Tags struct {
				FeatureType    string `json:"feature_type"`
				InventoryLabel string `json:"inventory_label"`
				PipText        string `json:"pip_text"`
				PriceType      string `json:"price_type"`
				PublishDate    string `json:"publish_date"`
				Returnable     string `json:"returnable"`
			} `json:"tags"`
			TotalReviews string `json:"total_reviews"`
		} `json:"products"`
	}

	matched := productInfoReg.FindSubmatch(respBody)
	if len(matched) < 2 {
		c.logger.Errorf("extract product skus failed %s", respBody)
		return fmt.Errorf("extract product skus failed")
	}
	if err := json.Unmarshal(matched[1], &productSkus); err != nil {
		c.logger.Errorf("decode product skus failed, error=%s", err)
		return err
	}

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	for _, prodInfo := range productSkus.Products {
		rating, _ := strconv.ParseFloat(prodInfo.AverageRating)

		var (
			orgPrice float64
			price    float64
			discount float64
			medias   []*pbMedia.Media
		)
		if orgPrice != price {
			discount = math.Round((orgPrice - price) / orgPrice * 100)
		}

		item := pbItem.Product{
			Source: &pbItem.Source{
				Id:           prodInfo.Code,
				CrawlUrl:     resp.Request.URL.String(),
				CanonicalUrl: dom.Find(`link[rel="canonical"]`).AttrOr("href", ""),
			},
			BrandName:   prodInfo.Brand,
			Title:       prodInfo.Name,
			Description: strings.TrimSpace(dom.Find("#collapsible-details-1").Text()),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(price * 100),
				Msrp:     int32(orgPrice) * 100,
				Discount: int32(discount),
			},
			Stats: &pbItem.Stats{
				ReviewCount: 0,
				Rating:      float32(rating),
			},
		}

		for _, skuInfo := range prodInfo.Skus {
			avaUrl := fmt.Sprintf("https://www.saksfifthavenue.com/on/demandware.store/Sites-SaksFifthAvenue-Site/en_US/Product-AvailabilityAjax?pid=%s&quantity=1&readyToOrder=true", skuInfo.Sku)
			req, _ := http.NewRequest(http.MethodGet, avaUrl, nil)
			req.Header.Set("Referer", resp.Request.URL.String())
			req.Header.Set("Accept", "*/*")
			req.Header.Set("x-requested-with", "XMLHttpRequest")

			var (
				skuResp *http.Response
				e       error
			)
			for i := 0; i < 3; i++ {
				c.logger.Debugf("access sku %s", skuInfo.Sku)
				if skuResp, e = c.httpClient.DoWithOptions(ctx, req, http.Options{
					EnableProxy: true,
					KeepSession: true,
					Reliability: c.CrawlOptions(resp.Request.URL).Reliability,
				}); e != nil {
					continue
				} else if skuResp.StatusCode == http.StatusNotFound {
					skuResp.Body.Close()

					e = errors.New("not found")
					break
				} else if skuResp.StatusCode == http.StatusForbidden ||
					skuResp.StatusCode == -1 {
					skuResp.Body.Close()

					e = fmt.Errorf("status %d %s", skuResp.StatusCode, skuResp.Status)
					continue
				}
				break
			}
			if e != nil {
				c.logger.Error(e)
				return e
			}
			defer skuResp.Body.Close()

			var viewData ProductDataJson
			if err := json.NewDecoder(skuResp.Body).Decode(&viewData); err != nil {
				c.logger.Error(err)
				return err
			}

			price, _ = strconv.ParsePrice(viewData.Product.Price.Sales.Value)
			orgPrice, _ = strconv.ParsePrice(viewData.Product.Price.List.Value)
			if orgPrice == 0 {
				orgPrice = price
			}
			if orgPrice != price {
				discount = math.Ceil((orgPrice - price) / orgPrice * 100)
			}

			medias = medias[0:0]
			for ki, mid := range viewData.Product.Images.HiRes {
				template := mid.URL
				medias = append(medias, pbMedia.NewImageMedia(
					strconv.Format(ki),
					template,
					strings.ReplaceAll(template, "wid=undefined&hei=undefined&", "wid=1000&hei=1333&"),
					strings.ReplaceAll(template, "wid=undefined&hei=undefined&", "wid=600&hei=800&"),
					strings.ReplaceAll(template, "wid=undefined&hei=undefined&", "wid=495&hei=660&"),
					"",
					ki == 0,
				))
			}

			sku := pbItem.Sku{
				SourceId: skuInfo.Sku,
				Title:    viewData.Product.ProductName,
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(price * 100),
					Msrp:     int32(orgPrice * 100),
					Discount: int32(discount),
				},
				Medias: medias,
				Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
				Stats: &pbItem.Stats{
					Rating:      float32(viewData.Product.Rating),
					ReviewCount: int32(viewData.Product.TurntoReviewCount),
				},
			}

			selectable := true
			for _, attr := range viewData.Product.VariationAttributes {
				switch attr.AttributeID {
				case "color":
					for _, val := range attr.Values {
						if !val.Selected {
							continue
						}
						sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
							Type:  pbItem.SkuSpecType_SkuSpecColor,
							Id:    val.ID,
							Name:  val.DisplayValue,
							Value: val.Value,
						})
						selectable = selectable && val.Selectable
						break
					}
				case "size":
					for _, val := range attr.Values {
						if !val.Selected {
							continue
						}
						sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
							Type:  pbItem.SkuSpecType_SkuSpecSize,
							Id:    val.ID,
							Name:  val.DisplayValue,
							Value: val.Value,
						})
						selectable = selectable && val.Selectable
						break
					}
				}
			}
			if selectable {
				sku.Stock.StockStatus = pbItem.Stock_InStock
			}

			item.SkuItems = append(item.SkuItems, &sku)
		}
		if err = yield(ctx, &item); err != nil {
			return err
		}
	}
	return nil
}

// parseProduct2
func (c *_Crawler) parseProduct2(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil {
		return nil
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	var productSkus struct {
		Products []struct {
			AverageRating string `json:"average_rating"`
			Brand         string `json:"brand"`
			Code          string `json:"code"`
			Name          string `json:"name"`
			OriginalPrice string `json:"original_price"`
			Price         string `json:"price"`
			Skus          []struct {
				AvailableDc string `json:"available_dc"`
				Sku         string `json:"sku"`
			} `json:"skus"`
			Tags struct {
				FeatureType    string `json:"feature_type"`
				InventoryLabel string `json:"inventory_label"`
				PipText        string `json:"pip_text"`
				PriceType      string `json:"price_type"`
				PublishDate    string `json:"publish_date"`
				Returnable     string `json:"returnable"`
			} `json:"tags"`
			TotalReviews string `json:"total_reviews"`
		} `json:"products"`
	}

	matched := productInfoReg.FindSubmatch(respBody)
	if len(matched) < 2 {
		c.logger.Errorf("extract product skus failed %s", respBody)
		return fmt.Errorf("extract product skus failed")
	}
	if err := json.Unmarshal(matched[1], &productSkus); err != nil {
		c.logger.Errorf("decode product skus failed, error=%s", err)
		return err
	}

	var (
		orgPrice float64
		price    float64
		discount float64
		medias   []*pbMedia.Media
		opts     = c.CrawlOptions(resp.Request.URL)
	)
	for _, prodInfo := range productSkus.Products {
		rating, _ := strconv.ParseFloat(prodInfo.AverageRating)

		item := pbItem.Product{
			Source: &pbItem.Source{
				Id:           prodInfo.Code,
				CrawlUrl:     resp.Request.URL.String(),
				CanonicalUrl: dom.Find(`link[rel="canonical"]`).AttrOr("href", ""),
			},
			BrandName:   prodInfo.Brand,
			Title:       prodInfo.Name,
			Description: strings.TrimSpace(dom.Find("#collapsible-details-1").Text()),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
			},
			Stats: &pbItem.Stats{
				ReviewCount: 0,
				Rating:      float32(rating),
			},
		}
		breadSel := dom.Find(`.product-detail .product-breadcrumb`)
		for i := range breadSel.Nodes {
			node := breadSel.Eq(i)
			cateSel := node.Find(`div[role="navigation"] ol.breadcrumb .breadcrumb-item`)
			for j := range cateSel.Nodes {
				if j == 0 {
					continue
				}
				cateNode := cateSel.Eq(j)
				switch j {
				case 1:
					item.Category = strings.TrimSpace(cateNode.Find("a").Text())
				case 2:
					item.SubCategory = strings.TrimSpace(cateNode.Find("a").Text())
				case 3:
					item.SubCategory2 = strings.TrimSpace(cateNode.Find("a").Text())
				case 4:
					item.SubCategory3 = strings.TrimSpace(cateNode.Find("a").Text())
				case 5:
					item.SubCategory4 = strings.TrimSpace(cateNode.Find("a").Text())
				}
			}
			break
		}

		colorSel := dom.Find(`.color-wrapper>li[role="radio"] .color-attribute`)
		if len(colorSel.Nodes) == 0 {
			colorSel = dom.Find(`.attribute .color .attr-name`)
		}
		for i := range colorSel.Nodes {
			node := colorSel.Eq(i)
			color := node.AttrOr("data-adobelaunchproductcolor", node.Find(`.color-attribute`).AttrOr("title", ""))

			u, _ := url.Parse("/on/demandware.store/Sites-SaksOff5th-Site/en_US/Product-Variation")
			u.Scheme = resp.Request.URL.Scheme
			u.Host = resp.Request.URL.Host
			vals := u.Query()
			vals.Set(fmt.Sprintf("dwvar_%s_color", prodInfo.Code), color)
			vals.Set("pid", prodInfo.Code)
			vals.Set("quantity", "1")
			u.RawQuery = vals.Encode()

			req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
			req.Header.Set("Referer", resp.Request.URL.String())
			req.Header.Set("Accept", "*/*")
			req.Header.Set("x-requested-with", "XMLHttpRequest")
			for k := range opts.MustHeader {
				req.Header.Set(k, opts.MustHeader.Get(k))
			}
			for _, c := range opts.MustCookies {
				if strings.HasPrefix(req.URL.Path, c.Path) || c.Path == "" {
					val := fmt.Sprintf("%s=%s", c.Name, c.Value)
					if c := req.Header.Get("Cookie"); c != "" {
						req.Header.Set("Cookie", c+"; "+val)
					} else {
						req.Header.Set("Cookie", val)
					}
				}
			}

			var (
				colorResp *http.Response
				e         error
			)
			for i := 0; i < 3; i++ {
				c.logger.Debugf("access sku %s", req.URL)

				if colorResp, e = c.httpClient.DoWithOptions(ctx, req, http.Options{
					EnableProxy: true,
					KeepSession: true,
					Reliability: c.CrawlOptions(resp.Request.URL).Reliability,
				}); e != nil {
					continue
				} else if colorResp.StatusCode == http.StatusNotFound {
					colorResp.Body.Close()

					e = errors.New("not found")
					break
				} else if colorResp.StatusCode == http.StatusForbidden ||
					colorResp.StatusCode == -1 {
					colorResp.Body.Close()

					e = fmt.Errorf("status %d %s", colorResp.StatusCode, colorResp.Status)
					continue
				}
				break
			}
			if e != nil {
				c.logger.Error(e)
				return e
			}
			defer colorResp.Body.Close()

			var viewData ProductDataJson
			if err := json.NewDecoder(colorResp.Body).Decode(&viewData); err != nil {
				c.logger.Error(err)
				return err
			}

			price, _ = strconv.ParsePrice(viewData.Product.Price.Sales.Value)
			orgPrice, _ = strconv.ParsePrice(viewData.Product.Price.List.Value)
			if orgPrice == 0 {
				orgPrice = price
			}
			if orgPrice != price {
				discount = math.Ceil((orgPrice - price) / orgPrice * 100)
			}

			medias = medias[0:0]
			for ki, mid := range viewData.Product.Images.HiRes {
				template := mid.URL
				medias = append(medias, pbMedia.NewImageMedia(
					strconv.Format(ki),
					template,
					strings.ReplaceAll(template, "wid=undefined&hei=undefined&", "wid=1000&hei=1333&"),
					strings.ReplaceAll(template, "wid=undefined&hei=undefined&", "wid=600&hei=800&"),
					strings.ReplaceAll(template, "wid=undefined&hei=undefined&", "wid=495&hei=660&"),
					"",
					ki == 0,
				))
			}

			var (
				colorAttr *productVariationAttributes
				sizeAttr  *productVariationAttributes
			)
			for _, attr := range viewData.Product.VariationAttributes {
				if attr.AttributeID == "color" && colorAttr == nil {
					colorAttr = attr
				} else if attr.AttributeID == "size" && sizeAttr == nil {
					sizeAttr = attr
				}
				if colorAttr != nil && sizeAttr != nil {
					break
				}
			}
			for _, colorVal := range colorAttr.Values {
				if !colorVal.Selected {
					continue
				}
				colorSpec := pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecColor,
					Id:    colorVal.ID,
					Name:  colorVal.DisplayValue,
					Value: colorVal.Value,
				}

				for _, sizeVal := range sizeAttr.Values {
					sku := pbItem.Sku{
						SourceId: fmt.Sprintf("%s-%s", colorVal.ID, sizeVal.ID),
						Title:    viewData.Product.ProductName,
						Price: &pbItem.Price{
							Currency: regulation.Currency_USD,
							Current:  int32(price * 100),
							Msrp:     int32(orgPrice * 100),
							Discount: int32(discount),
						},
						Medias: medias,
						Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
						Stats: &pbItem.Stats{
							Rating:      float32(viewData.Product.Rating),
							ReviewCount: int32(viewData.Product.TurntoReviewCount),
						},
					}
					if colorVal.Selectable && sizeVal.Selectable {
						sku.Stock.StockStatus = pbItem.Stock_InStock
					}
					sku.Specs = append(sku.Specs, &colorSpec)
					sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
						Type:  pbItem.SkuSpecType_SkuSpecSize,
						Id:    sizeVal.ID,
						Name:  sizeVal.DisplayValue,
						Value: sizeVal.Value,
					})
					item.SkuItems = append(item.SkuItems, &sku)
				}
			}
		}

		if err := yield(ctx, &item); err != nil {
			return err
		}
	}
	return nil
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		"https://www.saksoff5th.com/c/women/apparel/activewear",
		// "https://www.saksoff5th.com/product/swims-ms-lace-driving-shoes-0400013974163.html?dwvar_0400013974163_color=ORANGE",
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
	var disableParseDetail bool
	flag.BoolVar(&disableParseDetail, "disable-detail", false, "disable parse detail")
	flag.Parse()

	logger := glog.New(glog.LogLevelDebug)
	// build a http client
	// get proxy's microservice address from env
	client, err := proxy.NewProxyClient(os.Getenv("VOILA_PROXY_URL"), cookiejar.New(), logger)
	if err != nil {
		panic(err)
	}

	// instance the spider locally
	spider, err := New(client, logger)
	if err != nil {
		panic(err)
	}

	reqFilter := map[string]struct{}{}

	// this callback func is used to do recursion call of sub requests.
	var callback func(ctx context.Context, val interface{}) error
	callback = func(ctx context.Context, val interface{}) error {
		switch i := val.(type) {
		case *http.Request:
			if _, ok := reqFilter[i.URL.String()]; ok {
				return nil
			}
			reqFilter[i.URL.String()] = struct{}{}

			logger.Debugf("Access %s", i.URL)

			if disableParseDetail {
				crawler := spider.(*_Crawler)
				if crawler.productPathMatcher.MatchString(i.URL.Path) {
					return nil
				}
			}
			opts := spider.CrawlOptions(i.URL)

			// process logic of sub request

			// init custom headers
			for k := range opts.MustHeader {
				i.Header.Set(k, opts.MustHeader.Get(k))
			}

			// init custom cookies
			for _, c := range opts.MustCookies {
				if strings.HasPrefix(i.URL.Path, c.Path) || c.Path == "" {
					val := fmt.Sprintf("%s=%s", c.Name, c.Value)
					if c := i.Header.Get("Cookie"); c != "" {
						i.Header.Set("Cookie", c+"; "+val)
					} else {
						i.Header.Set("Cookie", val)
					}
				}
			}

			// set scheme,host for sub requests. for the product url in category page is just the path without hosts info.
			// here is just the test logic. when run the spider online, the controller will process automatically
			if i.URL.Scheme == "" {
				i.URL.Scheme = "https"
			}
			if i.URL.Host == "" {
				i.URL.Host = "www.saksoff5th.com"
			}

			// do http requests here.
			nctx, cancel := context.WithTimeout(ctx, time.Minute*5)
			defer cancel()
			resp, err := client.DoWithOptions(nctx, i, http.Options{
				EnableProxy:       true,
				EnableHeadless:    false,
				EnableSessionInit: opts.EnableSessionInit,
				KeepSession:       opts.KeepSession,
				Reliability:       opts.Reliability,
			})
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			return spider.Parse(ctx, resp, callback)
		default:
			// output the result
			data, err := protojson.Marshal(i.(proto.Message))
			if err != nil {
				return err
			}
			logger.Infof("data: %s", data)
		}
		return nil
	}

	ctx := context.WithValue(context.Background(), "tracing_id", fmt.Sprintf("tracing_%d", time.Now().UnixNano()))
	// start the crawl request
	for _, req := range spider.NewTestRequest(context.Background()) {
		if err := callback(ctx, req); err != nil {
			logger.Fatal(err)
		}
	}
}
