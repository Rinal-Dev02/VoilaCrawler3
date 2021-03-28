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
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type _Crawler struct {
	httpClient http.Client

	categoryPathMatcher     *regexp.Regexp
	categoryJsonPathMatcher *regexp.Regexp
	productGroupPathMatcher *regexp.Regexp
	productPathMatcher      *regexp.Regexp
	logger                  glog.Log
}

func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient: client,
		// /men-bags/COjWAcABAuICAgEY.zso
		categoryPathMatcher: regexp.MustCompile(`^(/[a-z0-9\-]+){1,5}/[a-zA-Z0-9]+\.zso$`),
		productPathMatcher:  regexp.MustCompile(`^/p(/[a-z0-9_-]+)/product/\d+(/[a-z0-9]+/\d+)?$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "e656205bf04d4436886b680d7b20a93b"
}

// Version
func (c *_Crawler) Version() int32 {
	return 1
}

// CrawlOptions
func (c *_Crawler) CrawlOptions(u *url.URL) *crawler.CrawlOptions {
	options := crawler.NewCrawlOptions()
	options.EnableHeadless = false
	options.LoginRequired = false
	options.MustCookies = append(options.MustCookies,
		&http.Cookie{Name: "geo", Value: "US/CA/803/LOSANGELES"},
		&http.Cookie{Name: "clouddc", Value: "west1"},
	)
	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.zappos.com"}
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

var productsExtractReg = regexp.MustCompile(`(?U)<script\s*>window.__INITIAL_STATE__\s*=\s*({.*});?\s*</script>`)

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

	matched := productsExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		return fmt.Errorf("extract json from product list page %s failed", resp.Request.URL)
	}

	var r struct {
		Filters struct {
			Page      int64 `json:"page"`
			PageCount int64 `json:"pageCount"`
		} `json:"filters"`
		Meta struct {
			DocumentMeta struct {
				Link struct {
					Rel struct {
						Next string `json:"next"`
					} `json:"rel"`
				} `json:"link"`
			} `json:"documentMeta"`
		} `json:"meta"`
		Products struct {
			TotalProductCount int           `json:"totalProductCount"`
			IsLoading         bool          `json:"isLoading"`
			RequestedURL      string        `json:"requestedUrl"`
			ExecutedSearchURL string        `json:"executedSearchUrl"`
			NoResultsRecos    []interface{} `json:"noResultsRecos"`
			InlineRecos       interface{}   `json:"inlineRecos"`
			OosMessaging      interface{}   `json:"oosMessaging"`
			HeartsList        struct {
			} `json:"heartsList"`
			ProductLimit int `json:"productLimit"`
			List         []struct {
				Sizing struct {
				} `json:"sizing"`
				ProductID   string `json:"productId"`
				ProductName string `json:"productName"`
				ProductURL  string `json:"productUrl"`
			} `json:"list"`
			AllProductsCount int `json:"allProductsCount"`
		} `json:"products"`
	}

	if err = json.Unmarshal(matched[1], &r); err != nil {
		c.logger.Debugf("parse %s failed, error=%s", matched[1], err)
		return err
	}

	lastIndex := nextIndex(ctx)
	for _, prod := range r.Products.List {
		if prod.ProductURL == "" {
			continue
		}
		req, err := http.NewRequest(http.MethodGet, prod.ProductURL, nil)
		if err != nil {
			c.logger.Error(err)
			continue
		}

		nctx := context.WithValue(ctx, "item.index", lastIndex)
		lastIndex += 1
		if err = yield(nctx, req); err != nil {
			return err
		}
	}

	// get current page number
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("p"))

	// check if this is the last page
	totalpages := r.Filters.PageCount
	if page >= totalpages || r.Meta.DocumentMeta.Link.Rel.Next == "" {
		return nil
	}

	req, _ := http.NewRequest(http.MethodGet, r.Meta.DocumentMeta.Link.Rel.Next, nil)
	nctx := context.WithValue(ctx, "item.index", lastIndex)

	return yield(nctx, req)

}

type productPageStructure struct {
	PixelServer struct {
		Data struct {
			CustomerCountryCode string `json:"customerCountryCode"`
			PageID              string `json:"pageId"`
			PageLang            string `json:"pageLang"`
			PageTitle           bool   `json:"pageTitle"`
			Product             struct {
				Sku         string `json:"sku"`
				StyleID     string `json:"styleId"`
				Price       string `json:"price"`
				Name        string `json:"name"`
				Brand       string `json:"brand"`
				Category    string `json:"category"`
				SubCategory string `json:"subCategory"`
				Gender      string `json:"gender"`
			} `json:"product"`
		} `json:"data"`
		PageType    string `json:"pageType"`
		QueryString string `json:"queryString"`
	} `json:"pixelServer"`
	Product struct {
		SelectedSizing struct {
			D7 string `json:"d7"`
		} `json:"selectedSizing"`
		Validation struct {
			Dimensions struct {
			} `json:"dimensions"`
		} `json:"validation"`
		ReviewData struct {
			SubmittedReviews []interface{} `json:"submittedReviews"`
			LoadingReviews   []interface{} `json:"loadingReviews"`
		} `json:"reviewData"`
		SearchReviewData struct {
			SearchTerm string `json:"searchTerm"`
		} `json:"searchReviewData"`
		IsDescriptionExpanded bool        `json:"isDescriptionExpanded"`
		CarouselIndex         int         `json:"carouselIndex"`
		SizingPredictionID    interface{} `json:"sizingPredictionId"`
		IsOnDemandEligible    interface{} `json:"isOnDemandEligible"`
		BrandPromo            struct {
		} `json:"brandPromo"`
		AvailableDimensionsForColor struct {
			Available struct {
				D7 struct {
					Num80325 bool `json:"80325"`
				} `json:"d7"`
			} `json:"available"`
		} `json:"availableDimensionsForColor"`
		Symphony struct {
			LoadingSymphonyComponents bool `json:"loadingSymphonyComponents"`
		} `json:"symphony"`
		SymphonyStory struct {
			LoadingSymphonyStoryComponents bool          `json:"loadingSymphonyStoryComponents"`
			Stories                        []interface{} `json:"stories"`
		} `json:"symphonyStory"`
		GenericSizeBiases struct {
		} `json:"genericSizeBiases"`
		SizingPredictionValue          interface{} `json:"sizingPredictionValue"`
		IsSimilarStylesLoading         bool        `json:"isSimilarStylesLoading"`
		OosButtonClicked               bool        `json:"oosButtonClicked"`
		IsSelectSizeTooltipVisible     bool        `json:"isSelectSizeTooltipVisible"`
		IsSelectSizeTooltipHighlighted bool        `json:"isSelectSizeTooltipHighlighted"`
		IsLoading                      bool        `json:"isLoading"`
		Detail                         struct {
			ReviewSummary struct {
				ReviewWithMostVotes   interface{} `json:"reviewWithMostVotes"`
				ReviewWithLeastVotes  interface{} `json:"reviewWithLeastVotes"`
				TotalCriticalReviews  string      `json:"totalCriticalReviews"`
				TotalFavorableReviews string      `json:"totalFavorableReviews"`
				TotalReviews          string      `json:"totalReviews"`
				TotalReviewScore      interface{} `json:"totalReviewScore"`
				AverageOverallRating  string      `json:"averageOverallRating"`
				ComfortRating         struct {
					Num4 string `json:"4"`
					Num5 string `json:"5"`
				} `json:"comfortRating"`
				OverallRating struct {
					Num5 string `json:"5"`
				} `json:"overallRating"`
				LookRating struct {
					Num5 string `json:"5"`
				} `json:"lookRating"`
				ArchRatingCounts struct {
				} `json:"archRatingCounts"`
				OverallRatingCounts struct {
					Num5 string `json:"5"`
				} `json:"overallRatingCounts"`
				SizeRatingCounts struct {
				} `json:"sizeRatingCounts"`
				WidthRatingCounts struct {
				} `json:"widthRatingCounts"`
				ArchRatingPercentages    interface{} `json:"archRatingPercentages"`
				OverallRatingPercentages struct {
					Num5 string `json:"5"`
				} `json:"overallRatingPercentages"`
				SizeRatingPercentages      interface{} `json:"sizeRatingPercentages"`
				WidthRatingPercentages     interface{} `json:"widthRatingPercentages"`
				MaxArchRatingPercentage    interface{} `json:"maxArchRatingPercentage"`
				MaxOverallRatingPercentage struct {
					Percentage string `json:"percentage"`
					Text       string `json:"text"`
				} `json:"maxOverallRatingPercentage"`
				MaxSizeRatingPercentage  interface{} `json:"maxSizeRatingPercentage"`
				MaxWidthRatingPercentage interface{} `json:"maxWidthRatingPercentage"`
				ReviewingAShoe           string      `json:"reviewingAShoe"`
				HasFitRatings            string      `json:"hasFitRatings"`
				AggregateRating          float64     `json:"aggregateRating"`
			} `json:"reviewSummary"`
			Sizing struct {
				AllUnits []struct {
					ID     string `json:"id"`
					Name   string `json:"name"`
					Rank   string `json:"rank"`
					Values []struct {
						ID    string `json:"id"`
						Rank  string `json:"rank"`
						Value string `json:"value"`
					} `json:"values"`
				} `json:"allUnits"`
				AllValues []struct {
					ID    string `json:"id"`
					Rank  string `json:"rank"`
					Value string `json:"value"`
				} `json:"allValues"`
				DimensionsSet []string `json:"dimensionsSet"`
				Dimensions    []struct {
					ID    string `json:"id"`
					Rank  string `json:"rank"`
					Name  string `json:"name"`
					Units []struct {
						ID     string `json:"id"`
						Name   string `json:"name"`
						Rank   string `json:"rank"`
						Values []struct {
							ID    string `json:"id"`
							Rank  string `json:"rank"`
							Value string `json:"value"`
						} `json:"values"`
					} `json:"units"`
				} `json:"dimensions"`
				StockData []struct {
					ID     string `json:"id"`
					Color  string `json:"color"`
					OnHand string `json:"onHand"`
					D7     string `json:"d7"`
				} `json:"stockData"`
				ValuesSet struct {
					D7 struct {
						U58309 []string `json:"u58309"`
					} `json:"d7"`
				} `json:"valuesSet"`
				ConvertedValueIDToValueID struct {
					Num80325 string `json:"80325"`
				} `json:"convertedValueIdToValueId"`
				DimensionIDToName struct {
					D7 string `json:"d7"`
				} `json:"dimensionIdToName"`
				DimensionIDToTagToUnitAndValues struct {
				} `json:"dimensionIdToTagToUnitAndValues"`
				DimensionIDToUnitID struct {
					D7 string `json:"d7"`
				} `json:"dimensionIdToUnitId"`
				Toggle struct {
				} `json:"toggle"`
				UnitIDToName struct {
					U58309 string `json:"u58309"`
				} `json:"unitIdToName"`
				ValueIDToName struct {
					Num80325 struct {
						Value     string `json:"value"`
						AbbrValue string `json:"abbrValue"`
					} `json:"80325"`
				} `json:"valueIdToName"`
				HypercubeSizingData struct {
					Num80325 struct {
						Min int   `json:"min"`
						Max int64 `json:"max"`
					} `json:"80325"`
				} `json:"hypercubeSizingData"`
			} `json:"sizing"`
			Videos             []interface{} `json:"videos"`
			Genders            []string      `json:"genders"`
			DefaultProductURL  string        `json:"defaultProductUrl"`
			DefaultImageURL    string        `json:"defaultImageUrl"`
			DefaultSubCategory interface{}   `json:"defaultSubCategory"`
			PreferredSubsite   interface{}   `json:"preferredSubsite"`
			OverallRating      []string      `json:"overallRating"`
			ProductName        string        `json:"productName"`
			ProductRating      string        `json:"productRating"`
			SizeFit            struct {
				Text       string `json:"text"`
				Percentage string `json:"percentage"`
			} `json:"sizeFit"`
			Description struct {
				BulletPoints []string      `json:"bulletPoints"`
				SizeCharts   []interface{} `json:"sizeCharts"`
			} `json:"description"`
			Zombie             bool   `json:"zombie"`
			BrandID            string `json:"brandId"`
			DefaultProductType string `json:"defaultProductType"`
			DefaultCategory    string `json:"defaultCategory"`
			ReviewCount        string `json:"reviewCount"`
			Styles             []struct {
				StyleDescription interface{} `json:"styleDescription"`
				Color            string      `json:"color"`
				OriginalPrice    string      `json:"originalPrice"`
				PercentOff       string      `json:"percentOff"`
				Price            string      `json:"price"`
				ProductURL       string      `json:"productUrl"`
				Stocks           []struct {
					OnHand  string `json:"onHand"`
					StockID string `json:"stockId"`
					SizeID  string `json:"sizeId"`
					Upc     string `json:"upc"`
					Size    string `json:"size"`
					Width   string `json:"width"`
					Asin    string `json:"asin"`
				} `json:"stocks"`
				HardLaunchDate     interface{} `json:"hardLaunchDate"`
				OnSale             string      `json:"onSale"`
				TaxonomyAttributes []struct {
					AttributeID int    `json:"attributeId"`
					Name        string `json:"name"`
					ValueID     int    `json:"valueId"`
					Value       string `json:"value"`
				} `json:"taxonomyAttributes"`
				StyleID   string `json:"styleId"`
				ImageURL  string `json:"imageUrl"`
				ColorID   string `json:"colorId"`
				ProductID string `json:"productId"`
				IsNew     string `json:"isNew"`
				Images    []struct {
					ImageID string `json:"imageId"`
					Type    string `json:"type"`
				} `json:"images"`
				SwatchImageID string `json:"swatchImageId"`
				Drop          bool   `json:"drop"`
				FinalSale     bool   `json:"finalSale"`
				TsdImages     struct {
				} `json:"tsdImages"`
				ImageID string `json:"imageId"`
			} `json:"styles"`
			VideoURL  interface{} `json:"videoUrl"`
			ProductID string      `json:"productId"`
			ArchFit   struct {
				Text       string `json:"text"`
				Percentage string `json:"percentage"`
			} `json:"archFit"`
			BrandName string `json:"brandName"`
			WidthFit  struct {
				Percentage string `json:"percentage"`
				Text       string `json:"text"`
			} `json:"widthFit"`
			Oos   bool `json:"oos"`
			Brand struct {
				ID             string `json:"id"`
				Name           string `json:"name"`
				CleanName      string `json:"cleanName"`
				BrandURL       string `json:"brandUrl"`
				ImageURL       string `json:"imageUrl"`
				HeaderImageURL string `json:"headerImageUrl"`
				RealBrandID    string `json:"realBrandId"`
			} `json:"brand"`
			BrandProductName      string      `json:"brandProductName"`
			IsReviewableWithMedia bool        `json:"isReviewableWithMedia"`
			HasHalfSizes          interface{} `json:"hasHalfSizes"`
			ReceivedDescription   string      `json:"receivedDescription"`
			YoutubeData           struct {
			} `json:"youtubeData"`
		} `json:"detail"`
		StyleThumbnails []struct {
			Color     string `json:"color"`
			ColorID   string `json:"colorId"`
			Src       string `json:"src"`
			TsdSrc    string `json:"tsdSrc"`
			StyleID   string `json:"styleId"`
			SwatchSrc string `json:"swatchSrc"`
		} `json:"styleThumbnails"`
		ReviewsTotalPages         int    `json:"reviewsTotalPages"`
		SeoProductURL             string `json:"seoProductUrl"`
		DimensionValueLengthTypes struct {
			D7 string `json:"d7"`
		} `json:"dimensionValueLengthTypes"`
		CalledClientSide bool `json:"calledClientSide"`
	} `json:"product"`
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
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}

	var viewData productPageStructure
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}

	reviews, _ := strconv.ParseInt(viewData.Product.Detail.ReviewSummary.TotalReviews)
	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:       strconv.Format(viewData.Product.Detail.ProductID),
			CrawlUrl: resp.Request.URL.String(),
		},
		BrandName:   viewData.Product.Detail.BrandName,
		Title:       viewData.Product.Detail.ProductName,
		Description: viewData.Product.Detail.ReceivedDescription,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		CrowdType:   viewData.PixelServer.Data.Product.Gender,
		Category:    viewData.PixelServer.Data.Product.Category,
		SubCategory: viewData.PixelServer.Data.Product.SubCategory,
		Stats: &pbItem.Stats{
			ReviewCount: int32(reviews),
			Rating:      float32(viewData.Product.Detail.ReviewSummary.AggregateRating),
		},
	}

	for _, rawSku := range viewData.Product.Detail.Styles {
		originalPrice, _ := strconv.ParsePrice(rawSku.OriginalPrice)
		price, _ := strconv.ParsePrice(rawSku.Price)
		discount, _ := strconv.ParseInt(strings.TrimSuffix(rawSku.PercentOff, "%"))

		var medias []*pbMedia.Media
		for _, m := range rawSku.Images {
			medias = append(medias, pbMedia.NewImageMedia(
				m.ImageID,
				fmt.Sprintf("https://m.media-amazon.com/images/I/%s.jpg", m.ImageID),
				fmt.Sprintf("https://m.media-amazon.com/images/I/%s._AC_SR1000,750_.jpg", m.ImageID),
				fmt.Sprintf("https://m.media-amazon.com/images/I/%s._AC_SR700,525_.jpg", m.ImageID),
				fmt.Sprintf("https://m.media-amazon.com/images/I/%s._AC_SR500,375_.jpg", m.ImageID),
				"",
				m.Type == "Main",
			))
		}

		for _, stock := range rawSku.Stocks {
			sku := pbItem.Sku{
				SourceId: rawSku.StyleID,
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(price * 100),
					Msrp:     int32(originalPrice * 100),
					Discount: int32(discount),
				},
				Medias: medias,
				Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}
			onhandCount, _ := strconv.ParseInt(stock.OnHand)
			if onhandCount > 0 {
				sku.Stock.StockStatus = pbItem.Stock_InStock
				sku.Stock.StockCount = int32(onhandCount)
			}

			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    strconv.Format(rawSku.ColorID),
				Name:  rawSku.Color,
				Value: rawSku.Color,
			})

			if len(rawSku.Stocks) > 0 {
				stock := rawSku.Stocks[0]
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecSize,
					Id:    stock.SizeID,
					Name:  stock.Size,
					Value: stock.Size,
				})
			}
			item.SkuItems = append(item.SkuItems, &sku)
		}
	}

	// yield item result
	if err = yield(ctx, &item); err != nil {
		return err
	}

	return nil
}

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		"https://www.zappos.com/men-bags/COjWAcABAuICAgEY.zso",
		// "https://www.zappos.com/p/nike-tanjun-black-white/product/8619473/color/151",
	} {
		req, _ := http.NewRequest(http.MethodGet, u, nil)
		reqs = append(reqs, req)
	}
	return reqs
}

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

	// this callback func is used to do recursion call of sub requests.
	var callback func(ctx context.Context, val interface{}) error
	callback = func(ctx context.Context, val interface{}) error {
		switch i := val.(type) {
		case *http.Request:
			logger.Debugf("Access %s", i.URL)
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
				i.URL.Host = "www.zappos.com"
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
