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
		categoryPathMatcher: regexp.MustCompile(`^(.*)(.zso)(.*)$`),
		productPathMatcher:  regexp.MustCompile(`(/[a-z0-9_-]+)/(product)`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "74b57a389ee549588cc6017f231ac53b"
}

// Version
func (c *_Crawler) Version() int32 {
	return 1
}

// CrawlOptions
func (c *_Crawler) CrawlOptions() *crawler.CrawlOptions {
	options := crawler.NewCrawlOptions()
	options.EnableHeadless = false
	options.LoginRequired = false
	options.MustHeader.Set("Accept-Language", "en-US,en;q=0.8")
	options.MustCookies = append(options.MustCookies,
		&http.Cookie{Name: "geocountry", Value: `US`, Path: "/"},
		// &http.Cookie{Name: "browseCountry", Value: "US", Path: "/"},
		// &http.Cookie{Name: "browseCurrency", Value: "USD", Path: "/"},
		// &http.Cookie{Name: "browseLanguage", Value: "en-US", Path: "/"},
		// &http.Cookie{Name: "browseSizeSchema", Value: "US", Path: "/"},
		// &http.Cookie{Name: "browseSizeSchema", Value: "US", Path: "/"},
		// &http.Cookie{Name: "storeCode", Value: "US", Path: "/"},
	)
	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"www.6pm.com"}
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

var prodDataExtraReg = regexp.MustCompile(`window.__INITIAL_STATE__\s*=\s*({.*});?\s*</script>`)

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

	// next page
	matched := prodDataExtraReg.FindSubmatch(respBody)
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
		rawurl := fmt.Sprintf("%s://%s%s", resp.Request.URL.Scheme, resp.Request.URL.Host, prod.ProductURL)
		if strings.HasPrefix(prod.ProductURL, "http:") || strings.HasPrefix(prod.ProductURL, "https:") {
			rawurl = prod.ProductURL
		}
		//fmt.Println(rawurl)

		// // prod page
		req, err := http.NewRequest(http.MethodGet, rawurl, nil)
		if err != nil {
			c.logger.Errorf("load http request of url %s failed, error=%s", rawurl, err)
			return err
		}

		lastIndex++
		// set the index of the product crawled in the sub response
		nctx := context.WithValue(ctx, "item.index", lastIndex)

		// yield sub request
		if err := yield(nctx, req); err != nil {
			return err
		}
	}

	// get current page number
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("p"))

	// check if this is the last page
	totalpages, _ := strconv.ParseInt(r.Filters.PageCount)
	if page >= totalpages || lastIndex >= int(r.Products.TotalProductCount) {
		return nil
	}

	u := fmt.Sprintf("%s://%s%s", resp.Request.URL.Scheme, resp.Request.URL.Host, r.Meta.DocumentMeta.Link.Rel.Next)

	req, _ := http.NewRequest(http.MethodGet, u, nil)
	// update the index of last page
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)

}

var productsExtractReg = regexp.MustCompile(`window.__INITIAL_STATE__\s*=\s*({.*});?\s*</script>`)

type productPageStructure struct {
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

	reviewCount, _ := strconv.ParseInt(viewData.Product.Detail.ReviewCount)
	rating, _ := strconv.ParseFloat(viewData.Product.Detail.ProductRating)

	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:       strconv.Format(viewData.Product.Detail.ProductID),
			CrawlUrl: resp.Request.URL.String(),
		},
		BrandName:   viewData.Product.Detail.BrandName,
		Title:       viewData.Product.Detail.ProductName,
		Description: htmlTrimRegp.ReplaceAllString(viewData.Product.Detail.ReceivedDescription, ""),
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stats: &pbItem.Stats{
			ReviewCount: int32(reviewCount),
			Rating:      float32(rating),
		},
	}

	for _, rawSku := range viewData.Product.Detail.Styles {

		for ks, rawSkusize := range rawSku.Stocks {

			originalPrice, _ := strconv.ParseFloat(strings.ReplaceAll(rawSku.Price, "$", ""))
			msrp, _ := strconv.ParseFloat(strings.ReplaceAll(rawSku.OriginalPrice, "$", ""))
			discount, _ := strconv.ParseInt(strings.TrimSuffix(rawSku.PercentOff, "%"))
			sku := pbItem.Sku{
				SourceId: strconv.Format(rawSku.ProductID),
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(originalPrice * 100),
					Msrp:     int32(msrp * 100),
					Discount: int32(discount),
				},
				Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}
			stock, _ := strconv.ParseInt(rawSkusize.OnHand)
			if stock > 0 {
				sku.Stock.StockStatus = pbItem.Stock_InStock
				sku.Stock.StockCount = int32(stock)
			}

			// color
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    strconv.Format(rawSku.ColorID),
				Name:  rawSku.Color,
				Value: rawSku.Color,
				//Icon:  color.SwatchMedia.Mobile,
			})

			if ks == 0 {
				isdefalut := true
				for ki, m := range rawSku.Images {
					if m.Type == "main" {
						isdefalut = false
					}

					sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
						strconv.Format(ki),
						"https://m.media-amazon.com/images/I/"+m.ImageID+"._SX1200_.jpg",
						"https://m.media-amazon.com/images/I/"+m.ImageID+"._SX1200_.jpg",
						"https://m.media-amazon.com/images/I/"+m.ImageID+"._SX700_.jpg",
						"https://m.media-amazon.com/images/I/"+m.ImageID+"._SX400_.jpg",
						"",
						isdefalut,
					))
				}
			}
			// size
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    rawSkusize.Upc,
				Name:  rawSkusize.Size,
				Value: rawSkusize.Size,
			})

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
		"https://www.6pm.com/women-dresses/CKvXARDE1wHAAQHiAgMBAhg.zso?s=isNew%2Fdesc%2FgoLiveDate%2Fdesc%2FrecentSalesStyle%2Fdesc%2F&p=1",
		"https://www.6pm.com/p/tommy-hilfiger-short-sleeve-polo-dress-sky-captain-bright-white/product/9496908/color/858095",
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

// local test
func main() {
	logger := glog.New(glog.LogLevelDebug)
	// build a http client.
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
				i.URL.Host = "www.6pm.com"
			}

			// do http requests here.
			nctx, cancel := context.WithTimeout(ctx, time.Minute*5)
			defer cancel()
			resp, err := client.DoWithOptions(nctx, i, http.Options{
				EnableProxy:       true,
				EnableHeadless:    false,
				EnableSessionInit: false,
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

	ctx := context.WithValue(context.Background(), "tracing_id", fmt.Sprintf("tracing_%d", time.Now().UnixNano()))
	// start the crawl request
	for _, req := range spider.NewTestRequest(context.Background()) {
		if err := callback(ctx, req); err != nil {
			logger.Fatal(err)
		}
	}
}
