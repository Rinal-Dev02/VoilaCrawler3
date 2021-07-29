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
		categoryPathMatcher: regexp.MustCompile(`^/us/(mens|womens)(/[a-zA-Z0-9\-]+){1,6}$`),
		// this regular used to match product page url path
		productPathMatcher: regexp.MustCompile(`^/us/products(/[a-zA-Z0-9\-]+){1,4}\-\d+?$`),
		logger:             logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "25eccd2dfaa90e141b35f01ad165af43"
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
	opts.MustCookies = append(opts.MustCookies,
		&http.Cookie{Name: "language", Value: "en_US"},
		&http.Cookie{Name: "country", Value: "USA"},
		&http.Cookie{Name: "billingCurrency", Value: "USD"},
		&http.Cookie{Name: "saleRegion", Value: "US"},
		&http.Cookie{Name: "MR", Value: "0"},
		&http.Cookie{Name: "loggedIn", Value: "false"},
		&http.Cookie{Name: "_dy_geo", Value: "US.NA.US_CA.US_CA_Los%20Angeles"},
	)
	return opts
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.matchesfashion.com"}
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
		u.Host = "www.matchesfashion.com"
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

	if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	} else if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yield)
	}
	return crawler.ErrUnsupportedPath
}

// nextIndex used to get the index from the shared data.
// item.index is a const key for item index.
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")))
}

// below are the golang json data struct of raw website.
// if you get the raw json data of the website,
// then you can use https://mholt.github.io/json-to-go/ to convert it to a golang struct

type CategoryData []struct {
	Images   []string `json:"images"`
	Name     string   `json:"name"`
	Designer string   `json:"designer"`
	URL      string   `json:"url"`
	Price    string   `json:"price"`
	Index    int      `json:"index"`
	Code     string   `json:"code"`
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

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Debug(err)
		return err
	}

	lastIndex := nextIndex(ctx)
	sel := dom.Find(`.lister .lister__wrapper .lister__item`)
	if len(sel.Nodes) == 0 {
		return nil
	}
	for i := range sel.Nodes {
		node := sel.Eq(i)
		href := node.Find(`.productMainLink`).AttrOr("href", "")
		if href == "" {
			continue
		}
		if !strings.HasPrefix(href, "/us") {
			href = "/us" + href
		}
		req, err := http.NewRequest(http.MethodGet, href, nil)
		if err != nil {
			c.logger.Errorf("load http request of url %s failed, error=%s", href, err)
			return err
		}

		lastIndex += 1
		nctx := context.WithValue(ctx, "item.index", lastIndex)
		// yield sub request
		if err := yield(nctx, req); err != nil {
			return err
		}
	}

	href := dom.Find(".redefine__right__pager .next a").AttrOr("href", "")
	if href == "" {
		return nil
	}
	if !strings.HasPrefix(href, "/us") {
		href = "/us" + href
	}

	req, _ := http.NewRequest(http.MethodGet, href, nil)
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
}

type productDataStructVersion1 struct {
	Props struct {
		PageProps struct {
			Host         string `json:"host"`
			AsPath       string `json:"asPath"`
			CanonicalURL string `json:"canonicalURL"`
			QueryParams  struct {
				Locale       string `json:"locale"`
				CanonicalURL string `json:"canonicalURL"`
			} `json:"queryParams"`
			CurrentLanguage   string            `json:"currentLanguage"`
			CurrentCountry    string            `json:"currentCountry"`
			Cookies           map[string]string `json:"cookies"`
			ProductDataParams struct {
				Code     string `json:"code"`
				Country  string `json:"country"`
				Currency string `json:"currency"`
				Language string `json:"language"`
			} `json:"productDataParams"`
			Product struct {
				BasicInfo struct {
					Code           string `json:"code"`
					Name           string `json:"name"`
					DesignerCode   string `json:"designerCode"`
					DesignerName   string `json:"designerName"`
					DesignerNameEn string `json:"designerNameEn"`
					DesignerURL    string `json:"designerUrl"`
					Gender         string `json:"gender"`
					Buyable        string `json:"buyable"`
					BuyableStatus  string `json:"buyableStatus"`
					Slug           string `json:"slug"`
					Colour         string `json:"colour"`
					ColourEn       string `json:"colourEn"`
					ColourCode     string `json:"colourCode"`
				} `json:"basicInfo"`
				Gallery struct {
					Images []struct {
						Template string   `json:"template"`
						Sequence []string `json:"sequence"`
						Alt      string   `json:"alt"`
					} `json:"images"`
					Videos []struct {
						URL      string `json:"url"`
						Alt      string `json:"alt"`
						Is360    bool   `json:"is360"`
						HasModel bool   `json:"hasModel"`
					} `json:"videos"`
				} `json:"gallery"`
				Pricing struct {
					Billing struct {
						Divisor      int    `json:"divisor"`
						Amount       int    `json:"amount"`
						Currency     string `json:"currency"`
						DisplayValue string `json:"displayValue"`
					} `json:"billing"`
					Rrp struct {
						Divisor      int    `json:"divisor"`
						Amount       int    `json:"amount"`
						Currency     string `json:"currency"`
						DisplayValue string `json:"displayValue"`
					} `json:"rrp"`
				} `json:"pricing"`
				Sizes []struct {
					Code        string `json:"code"`
					DisplayName string `json:"displayName"`
					Stock       string `json:"stock"`
				} `json:"sizes"`
				Editorial struct {
					Description       string   `json:"description"`
					DetailBullets     []string `json:"detailBullets"`
					SizeAndFitBullets []string `json:"sizeAndFitBullets"`
				} `json:"editorial"`
				RelatedProducts struct {
					ViewAllCategories []struct {
						Code string `json:"code"`
						Name string `json:"name"`
						URL  string `json:"url"`
					} `json:"viewAllCategories"`
				} `json:"relatedProducts"`
				Outfits struct {
					Gallery struct {
						Images []struct {
							Template string   `json:"template"`
							Sequence []string `json:"sequence"`
							Alt      string   `json:"alt"`
						} `json:"images"`
					} `json:"gallery"`
					Products []struct {
						Code         string `json:"code"`
						Name         string `json:"name"`
						DesignerName string `json:"designerName"`
						ProductURL   string `json:"productUrl"`
						Buyable      string `json:"buyable"`
					} `json:"products"`
				} `json:"outfits"`
				Categories []struct {
					Name   string `json:"name"`
					NameEn string `json:"nameEn"`
					URL    string `json:"url"`
				} `json:"categories"`
				Shipping struct {
					TaxAndDuty string `json:"taxAndDuty"`
					Cites      bool   `json:"cites"`
				} `json:"shipping"`
				Seo struct {
					Canonical string `json:"canonical"`
					Alternate []struct {
						Href     string `json:"href"`
						Hreflang string `json:"hreflang"`
					} `json:"alternate"`
				} `json:"seo"`
			} `json:"product"`
		} `json:"pageProps"`
	} `json:"props"`
	Page  string `json:"page"`
	Query struct {
		Locale       string `json:"locale"`
		CanonicalURL string `json:"canonicalURL"`
	} `json:"query"`
	BuildID       string `json:"buildId"`
	AssetPrefix   string `json:"assetPrefix"`
	RuntimeConfig struct {
		NEXTPUBLICPRODUCTSAPI             string `json:"NEXT_PUBLIC_PRODUCTS_API"`
		NEXTPUBLICSITEFURNITUREAPI        string `json:"NEXT_PUBLIC_SITE_FURNITURE_API"`
		ASSETSPREFIX                      string `json:"ASSETS_PREFIX"`
		NEXTPUBLICDATADOGCLIENTTOKEN      string `json:"NEXT_PUBLIC_DATADOG_CLIENT_TOKEN"`
		NEXTPUBLICDATADOGRUMAPPLICATIONID string `json:"NEXT_PUBLIC_DATADOG_RUM_APPLICATION_ID"`
		DDENV                             string `json:"DD_ENV"`
		NEXTPUBLICDDENV                   string `json:"NEXT_PUBLIC_DD_ENV"`
		NEXTPUBLICVERSION                 string `json:"NEXT_PUBLIC_VERSION"`
		NEXTPUBLICADOBEID                 string `json:"NEXT_PUBLIC_ADOBE_ID"`
		NEXTPUBLICDYID                    string `json:"NEXT_PUBLIC_DY_ID"`
		NEXTPUBLICAPPURL                  string `json:"NEXT_PUBLIC_APP_URL"`
	} `json:"runtimeConfig"`
	IsFallback   bool     `json:"isFallback"`
	DynamicIds   []string `json:"dynamicIds"`
	CustomServer bool     `json:"customServer"`
	Gip          bool     `json:"gip"`
}

type productDataStructVersion2 struct {
	Products []struct {
		Product struct {
			Code     string `json:"code"`
			URL      string `json:"url"`
			Name     string `json:"name"`
			Slug     string `json:"slug"`
			Designer struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"designer"`
			Image struct {
				Thumbnail string `json:"thumbnail"`
				Medium    string `json:"medium"`
				Large     string `json:"large"`
			} `json:"image"`
			ImageAltText string `json:"image.altText"`
			PriceData    struct {
				FormattedValue string `json:"formattedValue"`
			} `json:"priceData"`
			IsOneSize       bool `json:"isOneSize"`
			IsMyStylistOnly bool `json:"isMyStylistOnly"`
			VariantOptions  []struct {
				Code             string `json:"code"`
				SizeDataCode     string `json:"sizeDataCode"`
				SizeDataBaseCode string `json:"sizeDataBaseCode"`
				StockLevelCode   string `json:"stockLevelCode"`
				ComingSoon       string `json:"comingSoon"`
				Qty              string `json:"qty"`
				SizeData         string `json:"sizeData"`
			} `json:"variantOptions"`
		} `json:"product"`
	} `json:"products"`
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

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	if rawData := strings.TrimSpace(dom.Find(`script#__NEXT_DATA__`).Text()); rawData != "" {
		var viewData productDataStructVersion1

		if err := json.Unmarshal([]byte(rawData), &viewData); err != nil {
			c.logger.Debugf("%s %s", rawData, respBody)
			c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
			return err
		}

		prod := viewData.Props.PageProps.Product

		canUrl := dom.Find(`link[rel="canonical"]`).AttrOr("href", "")
		if canUrl == "" {
			canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
		}
		// build product data
		item := pbItem.Product{
			Source: &pbItem.Source{
				Id:           prod.BasicInfo.Code,
				CrawlUrl:     resp.Request.URL.String(),
				CanonicalUrl: canUrl,
			},
			BrandName: prod.BasicInfo.DesignerNameEn,
			Title:     prod.BasicInfo.Name,
			CrowdType: prod.BasicInfo.Gender,
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
			},
			// Stats: &pbItem.Stats{
			// 	ReviewCount: int32(p.NumberOfReviews),
			// 	Rating:      float32(p.ReviewAverageRating / 5.0),
			// },
		}
		for i, cate := range prod.Categories {
			switch i {
			case 0:
				item.Category = cate.NameEn
			case 1:
				item.SubCategory = cate.NameEn
			case 2:
				item.SubCategory2 = cate.NameEn
			case 3:
				item.SubCategory3 = cate.NameEn
			case 4:
				item.SubCategory4 = cate.NameEn
			}
		}

		colorSpec := pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecColor,
			Id:    prod.BasicInfo.ColourCode,
			Name:  prod.BasicInfo.ColourEn,
			Value: prod.BasicInfo.ColourCode,
		}
		price := pbItem.Price{
			Currency: regulation.Currency_USD,
			Current:  int32(prod.Pricing.Billing.Amount),
			Msrp:     int32(prod.Pricing.Rrp.Amount),
		}
		if price.Msrp == 0 {
			price.Msrp = int32(prod.Pricing.Billing.Amount)
		}
		price.Discount = int32(math.Ceil(100 * float64(price.Msrp-price.Current) / float64(price.Msrp)))

		medias := []*pbMedia.Media{}
		for _, img := range prod.Gallery.Images {
			tpl := img.Template
			for i, seq := range img.Sequence {
				tpl2 := strings.ReplaceAll(tpl, "{SEQUENCE}", seq)
				medias = append(medias, pbMedia.NewImageMedia(
					"",
					strings.ReplaceAll(tpl2, "{WIDTH}", "1000"),
					strings.ReplaceAll(tpl2, "{WIDTH}", "1000"),
					strings.ReplaceAll(tpl2, "{WIDTH}", "600"),
					strings.ReplaceAll(tpl2, "{WIDTH}", "500"),
					img.Alt,
					i == 0,
				))
			}
		}
		for _, video := range prod.Gallery.Videos {
			medias = append(medias, pbMedia.NewVideoMedia("", "", video.URL, 0, 0, 0, "", video.Alt, false))
		}

		for _, size := range prod.Sizes {
			sku := pbItem.Sku{
				SourceId: size.Code,
				Medias:   medias,
				Price:    &price,
				Stock:    &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}
			if size.Stock == "inStock" {
				sku.Stock.StockStatus = pbItem.Stock_InStock
			}
			sku.Specs = append(sku.Specs, &colorSpec, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    size.Code,
				Name:  size.DisplayName,
				Value: size.DisplayName,
			})
			item.SkuItems = append(item.SkuItems, &sku)
		}

		// yield item result
		if err = yield(ctx, &item); err != nil {
			return err
		}
	} else if rawData := dom.Find(`#pdpShopTheLookJSONData`).AttrOr("data-stl-json", ""); rawData != "" {
		c.logger.Debugf("dynamic %s", respBody)
		viewData := map[string]productDataStructVersion2{}
		if err := json.Unmarshal([]byte(rawData), &viewData); err != nil {
			c.logger.Error(err)
			return err
		}
		for key, group := range viewData {
			key = strings.TrimPrefix(key, "outfit_")
			for _, _prod := range group.Products {
				if _prod.Product.Code != key {
					continue
				}
				prod := _prod.Product

				item := pbItem.Product{
					Source: &pbItem.Source{
						Id:       prod.Code,
						CrawlUrl: resp.Request.URL.String(),
					},
					BrandName: prod.Designer.Name,
					Title:     prod.Name,
					Price: &pbItem.Price{
						Currency: regulation.Currency_USD,
					},
				}

				breadSel := dom.Find(`#breadcrumb>ul>li`)
				for i := range breadSel.Nodes {
					node := breadSel.Eq(i)
					text := strings.TrimSpace(node.Find("a").Text())
					switch i {
					case 0:
						item.Category = text
					case 1:
						item.SubCategory = text
					case 2:
						item.SubCategory2 = text
					}
				}

				currentPrice, _ := strconv.ParsePrice(prod.PriceData.FormattedValue)
				price := pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(currentPrice * 100),
					Msrp:     int32(currentPrice * 100),
				}
				orgPriceVal := strings.TrimSpace(dom.Find(`.product__header .pdp-price>:first-child`).Text())
				orgPrice, _ := strconv.ParsePrice(orgPriceVal)
				if orgPrice*100 > float64(price.Msrp) {
					price.Msrp = int32(orgPrice * 100)
					price.Discount = int32(math.Ceil(100 * float64(price.Msrp-price.Current) / float64(price.Msrp)))
				}

				var medias []*pbMedia.Media
				mediaSel := dom.Find(`.fs-slider .fs-slider__thumbnails>ul>li`)
				imgReg := regexp.MustCompile(`img/product/(?:[a-z]+_)?(\d+_\d+)(?:_[a-z]*)?(\.[a-z]*)`)
				for i := range mediaSel.Nodes {
					node := mediaSel.Eq(i)
					href := node.AttrOr("data-full-image", "")
					if href == "" {
						continue
					}
					matched := imgReg.FindStringSubmatch(href)
					if len(matched) != 3 {
						return fmt.Errorf("extract image from %s failed", href)
					}
					id := matched[1] + matched[2]
					tpl := strings.ReplaceAll("https://assetsprx.matchesfashion.com/img/product/{WIDTH}/{ID}", "{ID}", id)
					medias = append(medias, pbMedia.NewImageMedia(
						"",
						strings.ReplaceAll(tpl, "{WIDTH}", "1000"),
						strings.ReplaceAll(tpl, "{WIDTH}", "1000"),
						strings.ReplaceAll(tpl, "{WIDTH}", "600"),
						strings.ReplaceAll(tpl, "{WIDTH}", "500"),
						"",
						i == 0,
					))
				}

				for _, size := range prod.VariantOptions {
					sku := pbItem.Sku{
						SourceId: size.Code,
						Medias:   medias,
						Price:    &price,
						Stock:    &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
					}
					if size.StockLevelCode == "inStock" {
						sku.Stock.StockStatus = pbItem.Stock_InStock
					}
					sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
						Type:  pbItem.SkuSpecType_SkuSpecSize,
						Id:    size.Code,
						Name:  size.SizeData,
						Value: size.SizeData,
					})
					item.SkuItems = append(item.SkuItems, &sku)
				}
				if err := yield(ctx, &item); err != nil {
					return err
				}
			}
		}
	} else {
		c.logger.Debugf("%s", respBody)
		return errors.New("no product found")
	}
	return nil
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		// "https://www.matchesfashion.com/us/mens/shop/shoes",
		// "https://www.matchesfashion.com/us/products/Raey-Chest-pocket-cotton-blend-jacket--1317200",
		"https://www.matchesfashion.com/us/womens/shop/clothing/lingerie/briefs",
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
