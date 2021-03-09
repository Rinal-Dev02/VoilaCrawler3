package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

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
func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient: client,
		// this regular used to match category page url path
		categoryPathMatcher: regexp.MustCompile(`^/(browse|brands)(/[a-z0-9-]+){2,6}$`),
		// this regular used to match product page url path
		productPathMatcher: regexp.MustCompile(`^/s(/[a-z0-9-]+){1,3}/[0-9]+/?$`),
		logger:             logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "4b95dd02f3f535e5f2cc6254d64f56fe"
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
func (c *_Crawler) CrawlOptions() *crawler.CrawlOptions {
	return &crawler.CrawlOptions{
		EnableHeadless: false,
		// use js api to init session for the first request of the crawl
		EnableSessionInit: true,
		// Reliability high match to crawl api, and log match to backconnect
		Reliability: pbProxy.ProxyReliability_ReliabilityMedium,
	}
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.nordstrom.com"}
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
	return fmt.Errorf("unsupported url %s", resp.Request.URL.String())
}

// nextIndex used to get the index from the shared data.
// item.index is a const key for item index.
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")))
}

// below are the golang json data struct of raw website.
// if you get the raw json data of the website,
// then you can use https://mholt.github.io/json-to-go/ to convert it to a golang struct

type RawColor struct {
	ID             string   `json:"id"`
	Label          string   `json:"label"`
	SpriteIndex    int      `json:"spriteIndex"`
	MediaIds       []string `json:"mediaIds"`
	StandardColors []string `json:"standardColors"`
	SwatchMediaID  string   `json:"swatchMediaId"`
}

type RawMedia struct {
	ID    string `json:"id"`
	Group string `json:"group"`
	Type  string `json:"type"`
	Src   string `json:"src"`
}

type RawProduct struct {
	ID                int                 `json:"id"`
	BrandID           int                 `json:"brandId"`
	BrandName         string              `json:"brandName"`
	StyleNumber       string              `json:"styleNumber"`
	ColorCount        int                 `json:"colorCount"`
	ColorDefaultID    string              `json:"colorDefaultId"`
	ColorSpriteURL    string              `json:"colorSpriteUrl"`
	ColorsByID        map[string]RawColor `json:"colorsById"`
	ColorIds          []string            `json:"colorIds"`
	EnticementIds     []interface{}       `json:"enticementIds"`
	Enticements       interface{}         `json:"enticements"`
	IsBeauty          bool                `json:"isBeauty"`
	IsNordstromMade   bool                `json:"isNordstromMade"`
	IsOutfit          bool                `json:"isOutfit"`
	IsFeatured        bool                `json:"isFeatured"`
	IsUmap            bool                `json:"isUmap"`
	MediaByID         map[string]RawMedia `json:"mediaById"`
	Name              string              `json:"name"`
	PriceCurrencyCode string              `json:"priceCurrencyCode"`
	PriceCountryCode  string              `json:"priceCountryCode"`
	PricesByID        struct {
		Original struct {
			ID                  string      `json:"id"`
			MinItemPercentOff   int         `json:"minItemPercentOff"`
			MaxItemPercentOff   int         `json:"maxItemPercentOff"`
			MinItemPrice        string      `json:"minItemPrice"`
			MaxItemPrice        string      `json:"maxItemPrice"`
			PriceValidUntilDate interface{} `json:"priceValidUntilDate"`
		} `json:"original"`
	} `json:"pricesById"`
	ProductPageURL   string  `json:"productPageUrl"`
	ReviewCount      int     `json:"reviewCount"`
	ReviewStarRating float64 `json:"reviewStarRating"`
}

type RawViewData struct {
	ProductsByID map[string]RawProduct `json:"productsById"`
	Query        struct {
		PageCount        int      `json:"pageCount"`
		ProductOffset    int      `json:"productOffset"`
		PageProductCount int      `json:"pageProductCount"`
		PageSelected     int      `json:"pageSelected"`
		ResultCount      int      `json:"resultCount"`
		ResultProductIds []int    `json:"resultProductIds"`
		SortSelectedID   string   `json:"sortSelectedId"`
		SortOptionIds    []string `json:"sortOptionIds"`
	} `json:"query"`
}

type RawProductDetail struct {
	ID                  int     `json:"id"`
	ReviewAverageRating float32 `json:"reviewAverageRating"`
	Brand               struct {
		BrandName  string `json:"brandName"`
		BrandURL   string `json:"brandUrl"`
		ImsBrandID int    `json:"imsBrandId"`
	} `json:"brand"`
	CategoryID          string `json:"categoryId"`
	Description         string `json:"description"`
	DefaultGalleryMedia struct {
		StyleMediaID  int   `json:"styleMediaId"`
		ColorID       int   `json:"colorId"`
		IsTrimmed     bool  `json:"isTrimmed"`
		StyleMediaIds []int `json:"styleMediaIds"`
	} `json:"defaultGalleryMedia"`
	Features []string `json:"features"`
	Filters  struct {
		Color struct {
			ByID map[string]struct {
				ID              int    `json:"id"`
				Code            string `json:"code"`
				IsSelected      bool   `json:"isSelected"`
				IsDefault       bool   `json:"isDefault"`
				Value           string `json:"value"`
				DisplayValue    string `json:"displayValue"`
				ShouldDisplay   bool   `json:"shouldDisplay"`
				FilterType      string `json:"filterType"`
				IsAvailableWith string `json:"isAvailableWith"`
				RelatedSkuIds   []int  `json:"relatedSkuIds"`
				StyleMediaIds   []int  `json:"styleMediaIds"`
				SwatchMedia     struct {
					Desktop string `json:"desktop"`
					Mobile  string `json:"mobile"`
					Preview string `json:"preview"`
				} `json:"swatchMedia"`
			} `json:"byId"`
			AllIds []int `json:"allIds"`
		} `json:"color"`
		Size struct {
			ByID map[string]struct {
				ID              string `json:"id"`
				Value           string `json:"value"`
				DisplayValue    string `json:"displayValue"`
				GroupValue      string `json:"groupValue"`
				FilterType      string `json:"filterType"`
				RelatedSkuIds   []int  `json:"relatedSkuIds"`
				IsAvailableWith string `json:"isAvailableWith"`
			} `json:"byId"`
			AllIds []string `json:"allIds"`
		}
		Width struct {
			ByID struct { // TODO:
			} `json:"byId"`
			AllIds []interface{} `json:"allIds"`
		} `json:"width"`
		Group struct {
			ByID struct {
				Regular struct {
					Value               string      `json:"value"`
					DisplayValue        string      `json:"displayValue"`
					FilterType          string      `json:"filterType"`
					OriginalStyleNumber string      `json:"originalStyleNumber"`
					ShouldDisplay       bool        `json:"shouldDisplay"`
					RelatedSizeIds      interface{} `json:"relatedSizeIds"`
				} `json:"regular"`
			} `json:"byId"`
			AllIds []string `json:"allIds"`
		} `json:"group"`
	} `json:"filters"`
	FilterOptions []string `json:"filterOptions"`
	FitAndSize    struct {
		ContextualSizeDetail string      `json:"contextualSizeDetail"`
		FitGuideTitle        string      `json:"fitGuideTitle"`
		FitGuideURL          string      `json:"fitGuideUrl"`
		FitVideoTitle        string      `json:"fitVideoTitle"`
		FitVideoURL          string      `json:"fitVideoUrl"`
		HasSizeChart         bool        `json:"hasSizeChart"`
		SizeChartTitle       interface{} `json:"sizeChartTitle"`
		SizeChartURL         string      `json:"sizeChartUrl"`
		SizeDetail           interface{} `json:"sizeDetail"`
	} `json:"fitAndSize"`
	ImtFitAndSize struct {
	} `json:"imtFitAndSize"`
	FitCategory            string `json:"fitCategory"`
	Gender                 string `json:"gender"`
	GiftWithPurchase       string `json:"giftWithPurchase"`
	HealthHazardCategory   string `json:"healthHazardCategory"`
	ImsProductTypeID       int    `json:"imsProductTypeId"`
	ImsProductTypeParentID int    `json:"imsProductTypeParentId"`
	Ingredients            string `json:"ingredients"`
	MaxOrderQuantity       int    `json:"maxOrderQuantity"`
	Number                 string `json:"number"`
	NumberOfReviews        int    `json:"numberOfReviews"`
	PathAlias              string `json:"pathAlias"`
	Price                  struct {
		BySkuID map[string]struct {
			CurrentPercentOff      string `json:"currentPercentOff"`
			IsInternationalPricing bool   `json:"isInternationalPricing"`
			IsOriginalPriceRange   bool   `json:"isOriginalPriceRange"`
			IsRange                bool   `json:"isRange"`
			OriginalPriceString    string `json:"originalPriceString"`
			MaxPercentageOff       string `json:"maxPercentageOff"`
			PreviousPriceString    string `json:"previousPriceString"`
			PriceString            string `json:"priceString"`
			SaleEndDate            string `json:"saleEndDate"`
			SaleType               string `json:"saleType"`
			ShowSoldOutMessage     bool   `json:"showSoldOutMessage"`
			ShowUMapMessage        bool   `json:"showUMapMessage"`
			ShowUMapPrice          bool   `json:"showUMapPrice"`
			StyleID                int    `json:"styleId"`
			ValueStatement         string `json:"valueStatement"`
		} `json:"bySkuId"`
		AllSkuIds []int `json:"allSkuIds"`
		Style     struct {
			AllSkusOnSale          bool    `json:"allSkusOnSale"`
			CurrentMinPrice        float64 `json:"currentMinPrice"`
			CurrentMaxPrice        float64 `json:"currentMaxPrice"`
			CurrentPercentOff      string  `json:"currentPercentOff"`
			IsInternationalPricing bool    `json:"isInternationalPricing"`
			IsOriginalPriceRange   bool    `json:"isOriginalPriceRange"`
			IsRange                bool    `json:"isRange"`
			OriginalPriceString    string  `json:"originalPriceString"`
			MaxPercentageOff       string  `json:"maxPercentageOff"`
			PreviousPriceString    string  `json:"previousPriceString"`
			PriceString            string  `json:"priceString"`
			SaleEndDate            string  `json:"saleEndDate"`
			SaleType               string  `json:"saleType"`
			ShowSoldOutMessage     bool    `json:"showSoldOutMessage"`
			ShowUMapMessage        bool    `json:"showUMapMessage"`
			ShowUMapPrice          bool    `json:"showUMapPrice"`
			StyleID                int     `json:"styleId"`
			ValueStatement         string  `json:"valueStatement"`
		} `json:"style"`
	} `json:"price"`
	Promotion struct {
		PromoType               string  `json:"promoType"`
		StartDateTime           string  `json:"startDateTime"`
		EndDateTime             string  `json:"endDateTime"`
		MaximumPromotionalPrice float64 `json:"maximumPromotionalPrice"`
		MinimumPromotionalPrice float64 `json:"minimumPromotionalPrice"`
		MaximumPercentOff       float64 `json:"maximumPercentOff"`
		MinimumPercentOff       float64 `json:"minimumPercentOff"`
	} `json:"promotion"`
	PrimaryCategoryPathString string        `json:"primaryCategoryPathString"`
	ProductName               string        `json:"productName"`
	ProductTitle              string        `json:"productTitle"`
	ProductTypeName           string        `json:"productTypeName"`
	ProductTypeParentName     string        `json:"productTypeParentName"`
	SaleType                  string        `json:"saleType"`
	SalesVideoShot            interface{}   `json:"salesVideoShot"`
	SellingStatement          string        `json:"sellingStatement"`
	ShopperSizePreferences    []interface{} `json:"shopperSizePreferences"`
	Skus                      struct {
		ByID map[string]struct {
			ID                           int         `json:"id"`
			BackOrderDate                interface{} `json:"backOrderDate"`
			ColorID                      int         `json:"colorId"`
			DisplayOriginalPrice         string      `json:"displayOriginalPrice"`
			DisplayPercentOff            string      `json:"displayPercentOff"`
			DisplayPrice                 string      `json:"displayPrice"`
			IsAvailable                  bool        `json:"isAvailable"`
			IsBackOrder                  bool        `json:"isBackOrder"`
			IsDefault                    bool        `json:"isDefault"`
			LtsPrice                     float64     `json:"ltsPrice"`
			Price                        float64     `json:"price"`
			SizeID                       string      `json:"sizeId"`
			WidthID                      string      `json:"widthId"`
			RmsSkuID                     int         `json:"rmsSkuId"`
			TotalQuantityAvailable       int         `json:"totalQuantityAvailable"`
			IsAvailableFulfillmentCenter bool        `json:"isAvailableFulfillmentCenter"`
			FulfillmentChannelID         int         `json:"fulfillmentChannelId"`
		} `json:"byId"`
		AllIds []int `json:"allIds"`
	} `json:"skus"`
	StyleMedia struct {
		ByID map[string]struct {
			ID            int    `json:"id"`
			ColorID       int    `json:"colorId"`
			ColorName     string `json:"colorName"`
			ImageMediaURI struct {
				SmallDesktop              string `json:"smallDesktop"`
				MediumDesktop             string `json:"mediumDesktop"`
				LargeDesktop              string `json:"largeDesktop"`
				MaxLargeDesktop           string `json:"maxLargeDesktop"`
				SmallZoom                 string `json:"smallZoom"`
				Zoom                      string `json:"zoom"`
				MobileSmall               string `json:"mobileSmall"`
				MobileSmallTrimWhiteSpace string `json:"mobileSmallTrimWhiteSpace"`
				MobileMedium              string `json:"mobileMedium"`
				MobileLarge               string `json:"mobileLarge"`
				MobileZoom                string `json:"mobileZoom"`
				Mini                      string `json:"mini"`
			} `json:"imageMediaUri"`
			IsDefault      bool   `json:"isDefault"`
			IsSelected     bool   `json:"isSelected"`
			IsTrimmed      bool   `json:"isTrimmed"`
			MediaGroupType string `json:"mediaGroupType"`
			MediaType      string `json:"mediaType"`
			SortID         int    `json:"sortId"`
		} `json:"byId"`
		AllIds []int `json:"allIds"`
	} `json:"styleMedia"`
	StyleNumber string `json:"styleNumber"`
	StyleVideos struct {
		ByID struct {
		} `json:"byId"`
		AllIds []interface{} `json:"allIds"`
	} `json:"styleVideos"`
	TreatmentType string `json:"treatmentType"`
}

type CategoryView struct {
	OperatingCountryCode string      `json:"operatingCountryCode"`
	ProductResults       RawViewData `json:"productResults"`
	ViewData             RawViewData `json:"viewData"`
}

// used to extract embaded json data in website page.
// more about golang regulation see here https://golang.org/pkg/regexp/syntax/
var productsExtractReg = regexp.MustCompile(`(?U)<script\s*>window.__INITIAL_CONFIG__\s*=\s*({.*});?\s*</script>`)

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

	matched := productsExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		if err := c.httpClient.Jar().Clear(ctx, resp.Request.URL); err != nil {
			c.logger.Errorf("clear cookie for %s failed, error=%s", resp.Request.URL, err)
		}
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}
	// c.logger.Debugf("data: %s", matched[1])

	var viewData CategoryView
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		return err
	}

	lastIndex := nextIndex(ctx)
	for _, idv := range viewData.ViewData.Query.ResultProductIds {
		p, ok := viewData.ViewData.ProductsByID[fmt.Sprintf("%d", idv)]
		if !ok {
			c.logger.Warnf("product %v not found", idv)
			continue
		}

		req, err := http.NewRequest(http.MethodGet, p.ProductPageURL, nil)
		if err != nil {
			c.logger.Errorf("load http request of url %s failed, error=%s", p.ProductPageURL, err)
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

	// get current page number
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}
	// check if this is the last page
	if len(viewData.ViewData.Query.ResultProductIds) < viewData.ViewData.Query.PageProductCount ||
		page >= int64(viewData.ViewData.Query.PageCount) {
		return nil
	}

	// set pagination
	u := *resp.Request.URL
	vals := u.Query()
	vals.Set("page", strconv.Format(page+1))
	u.RawQuery = vals.Encode()

	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	// update the index of last page
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
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
	matched := productsExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		// clean cookie
		if err := c.httpClient.Jar().Clear(ctx, resp.Request.URL); err != nil {
			c.logger.Errorf("clear cookie for %s failed, error=%s", resp.Request.URL, err)
		}
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}
	// c.logger.Debugf("data: %s", matched[1])

	var viewData struct {
		StylesById struct {
			Data map[string]RawProductDetail `json:"data"`
		} `json:"stylesById"`
	}
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}
	for _, p := range viewData.StylesById.Data {
		// build product data
		item := pbItem.Product{
			Source: &pbItem.Source{
				Id:       strconv.Format(p.ID),
				CrawlUrl: resp.Request.URL.String(),
			},
			BrandName:   p.Brand.BrandName,
			Title:       p.ProductName,
			Description: htmlTrimRegp.ReplaceAllString(p.Description, "") + " " + strings.Join(p.Features, " "),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
			},
			Stock: &pbItem.Stock{
				StockStatus: pbItem.Stock_OutOfStock,
			},
			Stats: &pbItem.Stats{
				ReviewCount: int32(p.NumberOfReviews),
				Rating:      float32(p.ReviewAverageRating / 5.0),
			},
		}
		for _, mid := range p.DefaultGalleryMedia.StyleMediaIds {
			m := p.StyleMedia.ByID[strconv.Format(mid)]
			if m.MediaType == "Image" {
				item.Medias = append(item.Medias, pbMedia.NewImageMedia(
					strconv.Format(m.ID),
					m.ImageMediaURI.MaxLargeDesktop,
					m.ImageMediaURI.SmallZoom,
					m.ImageMediaURI.MobileLarge,
					m.ImageMediaURI.MobileMedium,
					"",
					m.IsDefault && m.IsSelected,
				))
			}
		}

		for _, rawSku := range p.Skus.ByID {
			originalPrice, _ := strconv.ParseFloat(rawSku.DisplayOriginalPrice)
			discount, _ := strconv.ParseInt(strings.TrimSuffix(rawSku.DisplayPercentOff, "%"))
			sku := pbItem.Sku{
				SourceId: strconv.Format(rawSku.ID),
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(rawSku.Price * 100),
					Msrp:     int32(originalPrice * 100),
					Discount: int32(discount),
				},
				Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}
			if rawSku.TotalQuantityAvailable > 0 {
				sku.Stock.StockStatus = pbItem.Stock_InStock
				sku.Stock.StockCount = int32(rawSku.TotalQuantityAvailable)
			}

			// color
			if rawSku.ColorID > 0 {
				color := p.Filters.Color.ByID[strconv.Format(rawSku.ColorID)]
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecColor,
					Id:    strconv.Format(color.ID),
					Name:  color.DisplayValue,
					Value: color.Value,
					Icon:  color.SwatchMedia.Mobile,
				})
				for _, mid := range color.StyleMediaIds {
					m := p.StyleMedia.ByID[strconv.Format(mid)]
					if m.MediaType == "Image" {
						sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
							strconv.Format(m.ID),
							m.ImageMediaURI.MaxLargeDesktop,
							m.ImageMediaURI.SmallZoom,
							m.ImageMediaURI.MobileLarge,
							m.ImageMediaURI.MobileMedium,
							"",
							m.IsDefault && m.IsSelected,
						))
					} else if m.MediaType == "Video" {
						// TODO
					}
				}
			}

			// size
			if rawSku.SizeID != "" {
				size := p.Filters.Size.ByID[rawSku.SizeID]
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecSize,
					Id:    size.ID,
					Name:  size.DisplayValue,
					Value: size.Value,
				})
			}

			item.SkuItems = append(item.SkuItems, &sku)
		}

		// yield item result
		if err = yield(ctx, &item); err != nil {
			return err
		}
	}
	return nil
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		// "https://www.nordstrom.com/browse/activewear/women-clothing?breadcrumb=Home%2FWomen%2FClothing%2FActivewear&origin=topnav",
		// "https://www.nordstrom.com/s/the-north-face-mountain-water-repellent-hooded-jacket/5500919",
		// "https://www.nordstrom.com/s/anastasia-beverly-hills-liquid-liner/5369732",
		"https://www.nordstrom.com/s/chanel-le-crayon-khol-intense-eye-pencil/2826730",
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
	opts := spider.CrawlOptions()

	// this callback func is used to do recursion call of sub requests.
	var callback func(ctx context.Context, val interface{}) error
	callback = func(ctx context.Context, val interface{}) error {
		switch i := val.(type) {
		case *http.Request:
			logger.Debugf("Access %s", i.URL)

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
				i.URL.Host = "www.nordstrom.com"
			}

			// do http requests here.
			nctx, cancel := context.WithTimeout(ctx, time.Minute*5)
			defer cancel()
			resp, err := client.DoWithOptions(nctx, i, http.Options{
				EnableProxy:       true,
				EnableHeadless:    false,
				EnableSessionInit: spider.CrawlOptions().EnableSessionInit,
				KeepSession:       spider.CrawlOptions().KeepSession,
				Reliability:       spider.CrawlOptions().Reliability,
			})
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			return spider.Parse(ctx, resp, callback)
		default:
			// output the result
			data, err := json.Marshal(i)
			if err != nil {
				return err
			}
			logger.Infof("data: %s", data)
		}
		return nil
	}

	ctx := context.WithValue(context.Background(), "tracing_id", "nordstrom_123456")
	// start the crawl request
	for _, req := range spider.NewTestRequest(context.Background()) {
		if err := callback(ctx, req); err != nil {
			logger.Fatal(err)
		}
	}
}
