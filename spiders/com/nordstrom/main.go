package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
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
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
)

type _Crawler struct {
	httpClient http.Client

	categoryPathMatcher *regexp.Regexp
	productPathMatcher  *regexp.Regexp
	logger              glog.Log
}

func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient:          client,
		categoryPathMatcher: regexp.MustCompile(`^/(browse|brands)(/[a-z0-9-]+){2,6}$`),
		productPathMatcher:  regexp.MustCompile(`^/s(/[a-z0-9-]+){1,3}/[0-9]+/?$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "4b95dd02f3f535e5f2cc6254d64f56fe"
}

// Version
func (c *_Crawler) Version() int32 {
	return 1
}

// CrawlOptions
func (c *_Crawler) CrawlOptions() *crawler.CrawlOptions {
	options := crawler.NewCrawlOptions()
	options.EnableHeadless = true
	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.nordstrom.com"}
}

func (c *_Crawler) IsUrlMatch(u *url.URL) bool {
	if c == nil || u == nil {
		return false
	}

	for _, reg := range []*regexp.Regexp{
		c.categoryPathMatcher,
		c.productPathMatcher,
	} {
		if reg.MatchString(u.Path) {
			return true
		}
	}
	return false
}

func (c *_Crawler) Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		// https://us.topshop.com/en/tsus/category/bags-accessories-7594012
		return c.parseCategoryProducts(ctx, resp, yield)
	} else if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yield)
	}
	return fmt.Errorf("unsupported url %s", resp.Request.URL.String())
}

// nextIndex used to get sharingData from context
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")) + 1)
}

var (
	productsExtractReg = regexp.MustCompile(`(?U)<script\s*>window.__INITIAL_CONFIG__\s*=\s*({.*});?\s*</script>`)
	defaultPageSize    = 24
)

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
			AllSkusOnSale          bool   `json:"allSkusOnSale"`
			CurrentMinPrice        int    `json:"currentMinPrice"`
			CurrentMaxPrice        int    `json:"currentMaxPrice"`
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
		} `json:"style"`
	} `json:"price"`
	Promotion struct {
		PromoType               string `json:"promoType"`
		StartDateTime           string `json:"startDateTime"`
		EndDateTime             string `json:"endDateTime"`
		MaximumPromotionalPrice int    `json:"maximumPromotionalPrice"`
		MinimumPromotionalPrice int    `json:"minimumPromotionalPrice"`
		MaximumPercentOff       int    `json:"maximumPercentOff"`
		MinimumPercentOff       int    `json:"minimumPercentOff"`
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
			LtsPrice                     int         `json:"ltsPrice"`
			Price                        int         `json:"price"`
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

// parseCategoryProducts parse api url from web page url
func (c *_Crawler) parseCategoryProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	matched := productsExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}
	// c.logger.Debugf("data: %s", matched[1])

	var viewData CategoryView
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		return err
	}

	lastIndex := nextIndex(ctx)
	for _, idv := range viewData.ProductResults.Query.ResultProductIds {
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
		nctx := context.WithValue(ctx, "item.index", lastIndex)
		if err := yield(nctx, req); err != nil {
			return err
		}
	}

	if len(viewData.ProductResults.Query.ResultProductIds) < viewData.ViewData.Query.PageProductCount {
		return nil
	}
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}
	if page >= int64(viewData.ViewData.Query.PageCount) {
		return nil
	}

	u := *resp.Request.URL
	vals := u.Query()
	vals.Set("page", strconv.Format(page+1))
	u.RawQuery = vals.Encode()

	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
}

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
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}
	c.logger.Debugf("data: %s", matched[1])

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
			Stats: &pbItem.Stats{
				ReviewCount: int32(p.NumberOfReviews),
				Rating:      float32(p.ReviewAverageRating / 5.0),
			},
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
							m.IsDefault,
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
		if err = yield(ctx, &item); err != nil {
			return err
		}
	}
	return nil
}

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		"https://www.nordstrom.com/browse/activewear/women-clothing?breadcrumb=Home%2FWomen%2FClothing%2FActivewear&origin=topnav",
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

func (c *_Crawler) CheckTestResponse(ctx context.Context, resp *http.Response) error {
	if err := c.Parse(ctx, resp, func(c context.Context, i interface{}) error {
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// local test
func main() {
	var (
		apiToken = os.Getenv("PC_API_TOKEN")
		jsToken  = os.Getenv("PC_JS_TOKEN")
	)
	if apiToken == "" || jsToken == "" {
		panic("env PC_API_TOKEN or PC_JS_TOKEN is not set")
	}

	logger := glog.New(glog.LogLevelDebug)
	client, err := proxy.NewProxyClient(
		cookiejar.New(), logger,
		proxy.Options{
			APIToken: apiToken,
			JSToken:  jsToken,
		},
	)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest(http.MethodGet, "https://www.nordstrom.com/browse/men/clothing/jeans?breadcrumb=Home%2FMen%2FClothing%2FJeans&origin=topnav", nil)
	if err != nil {
		panic(err)
	}
	resp, err := client.DoWithOptions(context.Background(), req, http.Options{EnableProxy: true, EnableHeadless: true})
	if err != nil {
		panic(err)
	}
	resp.Body.Close()

	spider, err := New(client, logger)
	if err != nil {
		panic(err)
	}
	opts := spider.CrawlOptions()

	var callback func(ctx context.Context, val interface{}) error
	callback = func(ctx context.Context, val interface{}) error {
		switch i := val.(type) {
		case *http.Request:
			logger.Debugf("Access %s", i.URL)

			for k := range opts.MustHeader {
				i.Header.Set(k, opts.MustHeader.Get(k))
			}
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
			if i.URL.Scheme == "" {
				i.URL.Scheme = "https"
			}
			if i.URL.Host == "" {
				i.URL.Host = "www.nordstrom.com"
			}

			resp, err := client.DoWithOptions(ctx, i, http.Options{EnableProxy: true, EnableHeadless: false, ProxyLevel: http.ProxyLevelReliable})
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			return spider.Parse(ctx, resp, callback)
		default:
			data, err := json.Marshal(i)
			if err != nil {
				return err
			}
			logger.Infof("data: %s", data)
		}
		return nil
	}

	for _, req := range spider.NewTestRequest(context.Background()) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
		defer cancel()

		if err := callback(ctx, req); err != nil {
			logger.Fatal(err)
		}
	}
}
