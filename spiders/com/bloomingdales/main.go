package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	//"github.com/gosimple/slug"
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

	logger glog.Log
}

func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient:          client,
		categoryPathMatcher: regexp.MustCompile(`^((\?!product).)*`),
		productPathMatcher:  regexp.MustCompile(`^(/[a-z0-9_-]+)?/shop((\/product\/))([/a-z0-9_-]+)$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "fe2669c60fa94d8595a59f10027a7877"
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
	options.MustCookies = append(options.MustCookies) //&http.Cookie{Name: "geocountry", Value: `US`, Path: "/"},
	// &http.Cookie{Name: "browseCountry", Value: "US", Path: "/"},
	// &http.Cookie{Name: "browseCurrency", Value: "USD", Path: "/"},
	// &http.Cookie{Name: "browseLanguage", Value: "en-US", Path: "/"},
	// &http.Cookie{Name: "browseSizeSchema", Value: "US", Path: "/"},
	// &http.Cookie{Name: "browseSizeSchema", Value: "US", Path: "/"},
	// &http.Cookie{Name: "storeCode", Value: "US", Path: "/"},

	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"www.bloomingdales.com"}
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

	if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yield)
	} else if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	}

	return fmt.Errorf("unsupported url %s", resp.Request.URL.String())
}

// nextIndex used to get sharingData from context
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")) + 1)
}

var (
	prodDataExtraReg      = regexp.MustCompile(`(data-bootstrap="page/discovery-pages"\s*type="application/json">)([^<]+)</script>`)
	prodDataPaginationReg = regexp.MustCompile(`(data-bootstrap="feature/canvas"\s*type="application/json">)([^<]+)</script>`)
)

type parseCategoryPagination struct {
	Row   int `json:"row"`
	Model struct {
		Pagination struct {
			NextURL       string `json:"nextURL"`
			BaseURL       string `json:"baseURL"`
			NumberOfPages int    `json:"numberOfPages"`
			CurrentPage   int    `json:"currentPage"`
		} `json:"pagination"`
	} `json:"model"`
}

type parseCategoryData struct {
	Meta struct {
		Analytics struct {
			Coremetrics struct {
				ClientID      string `json:"clientID"`
				CmHostURL     string `json:"cmHostUrl"`
				PageID        string `json:"pageID"`
				CategoryID    string `json:"categoryID"`
				SearchResults string `json:"searchResults"`
				Attributes    []struct {
					Name  string `json:"name"`
					Value string `json:"value"`
					Seq   string `json:"seq"`
				} `json:"attributes"`
				TrackingBreadcrumb   string `json:"trackingBreadcrumb"`
				BtCategoryID         string `json:"btCategoryID"`
				CategoryName         string `json:"categoryName"`
				ParentCategoryID     string `json:"parentCategoryID"`
				ParentCategoryName   string `json:"parentCategoryName"`
				FobCategoryID        string `json:"fobCategoryID"`
				FobCategoryName      string `json:"fobCategoryName"`
				TopLevelCategoryID   string `json:"topLevelCategoryID"`
				TopLevelCategoryName string `json:"topLevelCategoryName"`
			} `json:"coremetrics"`
			Data struct {
				CategoryID             string   `json:"categoryID"`
				SearchResults          string   `json:"searchResults"`
				TrackBreadcrumb        string   `json:"trackBreadcrumb"`
				BtCategory             string   `json:"btCategory"`
				CategoryName           string   `json:"categoryName"`
				ParentCategoryID       string   `json:"parentCategoryID"`
				ParentCategoryName     string   `json:"parentCategoryName"`
				FobCategoryID          string   `json:"fobCategoryID"`
				FobCategoryName        string   `json:"fobCategoryName"`
				TopLevelCategoryID     string   `json:"topLevelCategoryID"`
				TopLevelCategoryName   string   `json:"topLevelCategoryName"`
				ProductPlacementReason string   `json:"productPlacementReason"`
				ProductRating          []string `json:"productRating"`
				ProductReviews         []string `json:"productReviews"`
				ProductPricingState    []string `json:"productPricingState"`
				ProductID              []string `json:"productID"`
				ResultsCurrentPage     string   `json:"resultsCurrentPage"`
				ResultsPerPage         string   `json:"resultsPerPage"`
				SortType               string   `json:"sortType"`
				TotalResults           string   `json:"totalResults"`
				SearchPass             string   `json:"searchPass"`
				NewMarkDownProducts    []string `json:"newMarkDownProducts"`
				NewArrivalProducts     []string `json:"newArrivalProducts"`
			} `json:"data"`
		} `json:"analytics"`
	} `json:"meta"`
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

	var (
		r parseCategoryData
		p parseCategoryPagination
	)

	// -------------------------------------------------------------------- //
	// product list
	matched := prodDataExtraReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		return fmt.Errorf("extract json from product list page %s failed", resp.Request.URL)
	}

	matched[2] = bytes.ReplaceAll(bytes.ReplaceAll(matched[2], []byte("\\'"), []byte("'")), []byte(`\\"`), []byte(`\"`))
	if err = json.Unmarshal(matched[2], &r); err != nil {
		c.logger.Debugf("parse %s failed, error=%s", matched[1], err)
		return err
	}
	// -------------------------------------------------------------------- //
	// Product Pagination
	matched = prodDataPaginationReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		return fmt.Errorf("extract json from product list page %s failed", resp.Request.URL)
	}

	matched[2] = bytes.ReplaceAll(bytes.ReplaceAll(matched[2], []byte("\\'"), []byte("'")), []byte(`\\"`), []byte(`\"`))
	if err = json.Unmarshal(matched[2], &p); err != nil {
		c.logger.Debugf("parse %s failed, error=%s", matched[1], err)
		return err
	}
	// -------------------------------------------------------------------- //

	categoryId := r.Meta.Analytics.Data.CategoryID
	lastIndex := nextIndex(ctx)
	for _, idv := range r.Meta.Analytics.Data.ProductID {

		rawurl := fmt.Sprintf("%s://%s/shop/product/a?ID=%s&CategoryID=%s", resp.Request.URL.Scheme, resp.Request.URL.Host, idv, categoryId)
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
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("Pageindex"))
	if page == 0 {
		page = 1
	}

	// check if this is the last page
	totalResults, _ := strconv.ParseInt(r.Meta.Analytics.Data.TotalResults)
	// p.Model.Pagination.CurrentPage >= p.Model.Pagination.NumberOfPages ||
	if lastIndex >= int(totalResults) {
		return nil
	}

	u := fmt.Sprintf("%s://%s%s", resp.Request.URL.Scheme, resp.Request.URL.Host, p.Model.Pagination.NextURL)

	req, _ := http.NewRequest(http.MethodGet, u, nil)
	// update the index of last page
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
}

type parseProductData struct {
	Properties struct {
		KlarnaLibraryID        string `json:"klarnaLibraryId"`
		KlarnaLibrary          string `json:"klarnaLibrary"`
		ProxyServiceHost       string `json:"proxyServiceHost"`
		FooterServiceEndpoint  string `json:"footerServiceEndpoint"`
		HeaderServiceEndpoint  string `json:"headerServiceEndpoint"`
		KillswitchXapiURI      string `json:"killswitchXapiUri"`
		FooterXapiURI          string `json:"footerXapiUri"`
		HeaderXapiURI          string `json:"headerXapiUri"`
		IsBot                  bool   `json:"isBot"`
		FacebookAppID          string `json:"facebookAppId"`
		WebConcurrency         string `json:"webConcurrency"`
		StoreURLHost           string `json:"storeUrlHost"`
		ImageBaseURL           string `json:"imageBaseUrl"`
		AssetHost              string `json:"assetHost"`
		Host                   string `json:"host"`
		SecurityCspEnabled     bool   `json:"securityCspEnabled"`
		IsTealiumEnabled       bool   `json:"isTealiumEnabled"`
		WebcollageAPI          string `json:"webcollageApi"`
		BrightcoveAPI          string `json:"brightcoveApi"`
		BazaarvoicePrrAPI      string `json:"bazaarvoicePrrApi"`
		BazaarvoiceIshipAPI    string `json:"bazaarvoiceIshipApi"`
		BazaarvoiceAPI         string `json:"bazaarvoiceApi"`
		CustomerServiceHost    string `json:"customerServiceHost"`
		IsProduction           bool   `json:"isProduction"`
		TagEnv                 string `json:"tagEnv"`
		MasheryLocalHostHeader string `json:"masheryLocalHostHeader"`
		ProductXapiHost        string `json:"productXapiHost"`
		EntryPoint             string `json:"entryPoint"`
		NodeEnv                string `json:"nodeEnv"`
		XapiHost               string `json:"xapiHost"`
		Key1                   string `json:"key1"`
		RtoHost                string `json:"rtoHost"`
		Brand                  string `json:"brand"`
		DataMode               string `json:"dataMode"`
		FooterSideCaching      string `json:"footerSideCaching"`
		DtCollectorName        string `json:"dtCollectorName"`
		DtNodeAgentPath        string `json:"dtNodeAgentPath"`
		DtAgentName            string `json:"dtAgentName"`
	} `json:"properties"`
	Product struct {
		ID     int `json:"id"`
		Detail struct {
			Name                 string `json:"name"`
			Description          string `json:"description"`
			SecondaryDescription string `json:"secondaryDescription"`
			SeoKeywords          string `json:"seoKeywords"`
			Flags                struct {
				Chanel                                     bool `json:"chanel"`
				Hermes                                     bool `json:"hermes"`
				Coach                                      bool `json:"coach"`
				HasWarranty                                bool `json:"hasWarranty"`
				BigTicketItem                              bool `json:"bigTicketItem"`
				PhoneOnly                                  bool `json:"phoneOnly"`
				Registrable                                bool `json:"registrable"`
				MasterProduct                              bool `json:"masterProduct"`
				MemberProduct                              bool `json:"memberProduct"`
				GwpIndicator                               bool `json:"gwpIndicator"`
				TruefitEligible                            bool `json:"truefitEligible"`
				FitPredictorEligible                       bool `json:"fitPredictorEligible"`
				IsStoreOnlyProductOnline                   bool `json:"isStoreOnlyProductOnline"`
				EligibleForPreOrder                        bool `json:"eligibleForPreOrder"`
				CountryEligible                            bool `json:"countryEligible"`
				HasColors                                  bool `json:"hasColors"`
				Rebates                                    bool `json:"rebates"`
				GiftCard                                   bool `json:"giftCard"`
				SuppressColorSwatches                      bool `json:"suppressColorSwatches"`
				HasColorSwatches                           bool `json:"hasColorSwatches"`
				Beauty                                     bool `json:"beauty"`
				EligibleForShopRunner                      bool `json:"eligibleForShopRunner"`
				HasAdditionalImages                        bool `json:"hasAdditionalImages"`
				BigTicketV2CItem                           bool `json:"bigTicketV2CItem"`
				OnlineExclusive                            bool `json:"onlineExclusive"`
				StoreOnlySpecial                           bool `json:"storeOnlySpecial"`
				FinishLine                                 bool `json:"finishLine"`
				Sitewidesale                               bool `json:"sitewidesale"`
				ProtectionPlanEligible                     bool `json:"protectionPlanEligible"`
				BannerForKidsChokeHazard                   bool `json:"bannerForKidsChokeHazard"`
				SizePersistForMen                          bool `json:"sizePersistForMen"`
				SizesDropdownForShoesEnabled               bool `json:"sizesDropdownForShoesEnabled"`
				BigTicketDeliveryFeeRestructureEligible    bool `json:"bigTicketDeliveryFeeRestructureEligible"`
				DimensionsCopyGroupEnabled                 bool `json:"dimensionsCopyGroupEnabled"`
				MaterialCareSectionEnabled                 bool `json:"materialCareSectionEnabled"`
				SizeAndFitEnabled                          bool `json:"sizeAndFitEnabled"`
				ArBeauty                                   bool `json:"arBeauty"`
				ArFurniture                                bool `json:"arFurniture"`
				VirtualTryOn                               bool `json:"virtualTryOn"`
				Experience3D                               bool `json:"experience3D"`
				Experience360                              bool `json:"experience360"`
				BackInStockOptOut                          bool `json:"backInStockOptOut"`
				ConsolidatedProductComplex                 bool `json:"consolidatedProductComplex"`
				WriteAReviewRedesignExpEnabled             bool `json:"writeAReviewRedesignExpEnabled"`
				SiteMonetizationProduct                    bool `json:"siteMonetizationProduct"`
				ConsolidatedMaster                         bool `json:"consolidatedMaster"`
				ProcessedProdDesc                          bool `json:"processedProdDesc"`
				EsecRemoveSecureUserTokenQueryParamEnabled bool `json:"esecRemoveSecureUserTokenQueryParamEnabled"`
				SeeMoreExperienceEnabled                   bool `json:"seeMoreExperienceEnabled"`
				SeeMoreAndSizeChartExperienceEnabled       bool `json:"seeMoreAndSizeChartExperienceEnabled"`
				SizeChartExperienceEnabled                 bool `json:"sizeChartExperienceEnabled"`
				GwpExperienceEnabled                       bool `json:"gwpExperienceEnabled"`
				BcomsyndigoEnabled                         bool `json:"bcomsyndigoEnabled"`
				StyleMeEnabled                             bool `json:"styleMeEnabled"`
				ZeekitEnabled                              bool `json:"zeekitEnabled"`
				AltModelSizesExperimentEnabled             bool `json:"altModelSizesExperimentEnabled"`
				PDPColorized                               bool `json:"PDPColorized"`
				IsShoeSizeSelectorsEnabled                 bool `json:"isShoeSizeSelectorsEnabled"`
				IsEligibleForColorwayPromoBadging          bool `json:"isEligibleForColorwayPromoBadging"`
				IsFewLeftMessageRedesignEnabled            bool `json:"isFewLeftMessageRedesignEnabled"`
				IsTrueFitSizeAutoSelectEnabled             bool `json:"isTrueFitSizeAutoSelectEnabled"`
				PdpProductEngagementPromptTrtTwoEnabled    bool `json:"pdpProductEngagementPromptTrtTwoEnabled"`
				KlarnaEligible                             bool `json:"klarnaEligible"`
				IsPdpBVReviewFormUpdatesEnabled            bool `json:"isPdpBVReviewFormUpdatesEnabled"`
				Phase2DesktopBVAPITrt1Enabled              bool `json:"phase2DesktopBVApiTrt1Enabled"`
				Phase2DesktopBVAPITrt2Enabled              bool `json:"phase2DesktopBVApiTrt2Enabled"`
				Phase2MobileBVAPITrt1Enabled               bool `json:"phase2MobileBVApiTrt1Enabled"`
				IsNew                                      bool `json:"isNew"`
				IsFindationEnabled                         bool `json:"isFindationEnabled"`
				PdpExpansionEnabled                        bool `json:"pdpExpansionEnabled"`
				IsReviewPhotosUploadEnabled                bool `json:"isReviewPhotosUploadEnabled"`
				IsWebCollageOutFromTabs                    bool `json:"isWebCollageOutFromTabs"`
				IsDressSizeSelectorsEnabled                bool `json:"isDressSizeSelectorsEnabled"`
				IsVideoImageRailEnabled                    bool `json:"isVideoImageRailEnabled"`
				IsTrueCollectionsEnabled                   bool `json:"isTrueCollectionsEnabled"`
				IsReviewsPage                              bool `json:"isReviewsPage"`
				IsEGiftCard                                bool `json:"isEGiftCard"`
			} `json:"flags"`
			ReviewStatistics struct {
				Aggregate struct {
					Rating           float64 `json:"rating"`
					RatingPercentage int     `json:"ratingPercentage"`
					Count            int     `json:"count"`
				} `json:"aggregate"`
			} `json:"reviewStatistics"`
			OrderedMasterGroupList []interface{} `json:"orderedMasterGroupList"`
			MemberDisplayGroupsMap struct {
			} `json:"memberDisplayGroupsMap"`
			BulletText            []string `json:"bulletText"`
			MaterialsAndCare      []string `json:"materialsAndCare"`
			MaxQuantity           int      `json:"maxQuantity"`
			TypeName              string   `json:"typeName"`
			AdditionalImagesCount int      `json:"additionalImagesCount"`
			NumberOfColors        int      `json:"numberOfColors"`
			Brand                 struct {
				Name          string `json:"name"`
				ID            int    `json:"id"`
				URL           string `json:"url"`
				SubBrand      string `json:"subBrand"`
				BrandBreakout bool   `json:"brandBreakout"`
			} `json:"brand"`
			BulletLinks            []interface{} `json:"bulletLinks"`
			PdfEmailDescription    string        `json:"pdfEmailDescription"`
			MemberProductCount     int           `json:"memberProductCount"`
			CompleteName           string        `json:"completeName"`
			DimensionsHeaderText   string        `json:"dimensionsHeaderText"`
			DressOccasion          string        `json:"dressOccasion"`
			SizeLabel              string        `json:"sizeLabel"`
			KidsChokingHazardLabel string        `json:"kidsChokingHazardLabel"`
			Klarna                 struct {
				KlarnaDataClientID    string `json:"klarnaDataClientId"`
				KlarnaOnsiteJsSdkPath string `json:"klarnaOnsiteJsSdkPath"`
			} `json:"klarna"`
			ReviewTitle string `json:"reviewTitle"`
		} `json:"detail"`
		Traits struct {
			Colors struct {
				SelectedColor int `json:"selectedColor"`
				ColorMap      map[string]struct {
					//Num1817529 struct {
					ID          int    `json:"id"`
					Name        string `json:"name"`
					NormalName  string `json:"normalName"`
					SwatchImage struct {
						FilePath             string `json:"filePath"`
						Name                 string `json:"name"`
						ShowJumboSwatch      bool   `json:"showJumboSwatch"`
						SwatchSpriteOffset   int    `json:"swatchSpriteOffset"`
						SwatchSpriteURLIndex int    `json:"swatchSpriteUrlIndex"`
					} `json:"swatchImage"`
					Imagery struct {
						Images []struct {
							FilePath             string `json:"filePath"`
							Name                 string `json:"name"`
							ShowJumboSwatch      bool   `json:"showJumboSwatch"`
							SwatchSpriteOffset   int    `json:"swatchSpriteOffset"`
							SwatchSpriteURLIndex int    `json:"swatchSpriteUrlIndex"`
						} `json:"images"`
						SmallImagesSprites struct {
							SpriteUrls      []string `json:"spriteUrls"`
							ImagesWidth     int      `json:"imagesWidth"`
							ImagesHeight    int      `json:"imagesHeight"`
							ImagesPerSprite int      `json:"imagesPerSprite"`
						} `json:"smallImagesSprites"`
						LargeImagesSprites struct {
							SpriteUrls      []string `json:"spriteUrls"`
							ImagesWidth     int      `json:"imagesWidth"`
							ImagesHeight    int      `json:"imagesHeight"`
							ImagesPerSprite int      `json:"imagesPerSprite"`
						} `json:"largeImagesSprites"`
						PrimaryImage struct {
							FilePath             string `json:"filePath"`
							Name                 string `json:"name"`
							ShowJumboSwatch      bool   `json:"showJumboSwatch"`
							SwatchSpriteOffset   int    `json:"swatchSpriteOffset"`
							SwatchSpriteURLIndex int    `json:"swatchSpriteUrlIndex"`
						} `json:"primaryImage"`
					} `json:"imagery"`
					Sizes   []int `json:"sizes"`
					Pricing struct {
						Price struct {
							PriceType struct {
								OnEdv               bool `json:"onEdv"`
								OnSale              bool `json:"onSale"`
								UpcOnSale           bool `json:"upcOnSale"`
								UpcOnEdv            bool `json:"upcOnEdv"`
								MemberProductOnSale bool `json:"memberProductOnSale"`
								WillBe              bool `json:"willBe"`
								ApplicableToAllUpcs bool `json:"applicableToAllUpcs"`
								SelectItemsOnSale   bool `json:"selectItemsOnSale"`
								IsMasterNonRanged   bool `json:"isMasterNonRanged"`
								IsFinalOfferType    bool `json:"isFinalOfferType"`
								ShowFinalOfferText  bool `json:"showFinalOfferText"`
							} `json:"priceType"`
							Policy struct {
								Text string `json:"text"`
								URL  string `json:"url"`
							} `json:"policy"`
							TieredPrice []struct {
								Label  string `json:"label"`
								Values []struct {
									Value          float64 `json:"value"`
									FormattedValue string  `json:"formattedValue"`
									Type           string  `json:"type"`
								} `json:"values"`
							} `json:"tieredPrice"`
							PriceTypeID int `json:"priceTypeId"`
						} `json:"price"`
						BadgeIds []string `json:"badgeIds"`
					} `json:"pricing"`
					//} `json:"1817529"`
				} `json:"colorMap"`
				SwatchSprite struct {
					SwatchSpriteUrls  []string `json:"swatchSpriteUrls"`
					SpriteSwatchSize  int      `json:"spriteSwatchSize"`
					SwatchesPerSprite int      `json:"swatchesPerSprite"`
				} `json:"swatchSprite"`
				LargeSwatchSprite struct {
					SwatchSpriteUrls  []string `json:"swatchSpriteUrls"`
					SpriteSwatchSize  int      `json:"spriteSwatchSize"`
					SwatchesPerSprite int      `json:"swatchesPerSprite"`
				} `json:"largeSwatchSprite"`
				OrderedColorsByID   []int `json:"orderedColorsById"`
				OrderedColorsByName []int `json:"orderedColorsByName"`
			} `json:"colors"`
			Sizes struct {
				OrderedSizesBySeqNumber []int  `json:"orderedSizesBySeqNumber"`
				SizeChartID             string `json:"sizeChartId"`
				SizeMap                 map[string]struct {
					//Num0 struct {
					ID          int    `json:"id"`
					Name        string `json:"name"`
					DisplayName string `json:"displayName"`
					Colors      []int  `json:"colors"`
					//} `json:"0"`
				} `json:"sizeMap"`
			} `json:"sizes"`
			TraitsMaps struct {
				UpcMap struct {
					One8175291 int `json:"1817529_1"`
					One8175290 int `json:"1817529_0"`
					One8175293 int `json:"1817529_3"`
					One8175292 int `json:"1817529_2"`
					One8175294 int `json:"1817529_4"`
				} `json:"upcMap"`
				PriceToColors []struct {
					Price    string `json:"price"`
					ColorIds []int  `json:"colorIds"`
					OnSale   bool   `json:"onSale"`
				} `json:"priceToColors"`
			} `json:"traitsMaps"`
		} `json:"traits"`
	} `json:"product"`
	UtagData struct {
		ProductRating  []string `json:"product_rating"`
		ProductReviews []string `json:"product_reviews"`
	} `json:"utagData"`
}

var (
	detailReg = regexp.MustCompile(`(data-bootstrap="page/product"\s*type="application/json">)({.*})</script>`)
)

func (c *_Crawler) parseProduct(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error(err)
		return err
	}

	matched := detailReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		c.logger.Debugf("data %s", respBody)
		return fmt.Errorf("extract produt json from page %s content failed", resp.Request.URL)
	}

	var (
		pd parseProductData
	)

	matched[2] = bytes.ReplaceAll(matched[2], []byte("\\\\\\\""), []byte("\\\\\\\\\""))
	matched[2] = bytes.ReplaceAll(bytes.ReplaceAll(matched[2], []byte("\\'"), []byte("'")), []byte(`\\"`), []byte(`\"`))

	if err = json.Unmarshal(matched[2], &pd); err != nil {
		c.logger.Error(err)
		return err
	}

	reviewCount, _ := strconv.ParseFloat(pd.UtagData.ProductReviews[0])
	rating, _ := strconv.ParseFloat(pd.UtagData.ProductRating[0])

	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:       strconv.Format(pd.Product.ID),
			CrawlUrl: resp.Request.URL.String(),
		},
		Title:       pd.Product.Detail.Name,
		Description: pd.Product.Detail.Description,
		BrandName:   pd.Product.Detail.Brand.Name,
		//CrowdType:    i.Details.GenderName,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},

		Stats: &pbItem.Stats{
			ReviewCount: int32(reviewCount),
			Rating:      float32(rating),
		},
	}

	for _, p := range pd.Product.Traits.Colors.ColorMap {

		current, _ := strconv.ParseFloat(p.Pricing.Price.TieredPrice[1].Values[0].Value)
		msrp, _ := strconv.ParseFloat(p.Pricing.Price.TieredPrice[0].Values[0].Value)
		discount := (msrp - current) * 100 / msrp

		for ks, rawSize := range p.Sizes {

			sizeID := strconv.Format(rawSize)
			sku := pbItem.Sku{
				SourceId: strconv.Format(p.ID) + "_" + strconv.Format(rawSize),
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(current * 100),
					Msrp:     int32(msrp * 100),
					Discount: int32(discount),
				},
				Stock: &pbItem.Stock{StockStatus: pbItem.Stock_InStock},
			}
			// if rawSize.StockLevelStatus == "inStock" {  // ASk ?
			// 	sku.Stock.StockStatus = pbItem.Stock_InStock
			// 	//sku.Stock.StockCount = int32(rawSize.Quantity)
			// }

			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    sizeID,
				Name:  pd.Product.Traits.Sizes.SizeMap[sizeID].Name,
				Value: pd.Product.Traits.Sizes.SizeMap[sizeID].DisplayName,
			})

			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{ // color ASK
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    strconv.Format(p.ID),
				Name:  p.NormalName,
				Value: p.NormalName,
			})

			if ks == 0 {

				isDefault := true
				for key, img := range p.Imagery.Images {
					if key > 1 {
						isDefault = false
					}
					sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
						strconv.Format(p.ID),
						"https://images.bloomingdalesassets.com/is/image/BLM/products/"+img.FilePath,
						"https://images.bloomingdalesassets.com/is/image/BLM/products/"+img.FilePath+"?op_sharpen=1&wid=1000",
						"https://images.bloomingdalesassets.com/is/image/BLM/products/"+img.FilePath+"?op_sharpen=1&wid=700",
						"https://images.bloomingdalesassets.com/is/image/BLM/products/"+img.FilePath+"?op_sharpen=1&wid=450",
						"",
						isDefault,
					))
				}

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
		"https://www.bloomingdales.com/shop/womens-apparel/tops-tees?id=5619",
		"https://www.bloomingdales.com/shop/product/aqua-puff-sleeve-top-100-exclusive?ID=3845413&CategoryID=5619",
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
				i.URL.Host = "www.bloomingdales.com"
			}

			// do http requests here.
			nctx, cancel := context.WithTimeout(ctx, time.Minute*5)
			defer cancel()
			resp, err := client.DoWithOptions(nctx, i, http.Options{
				EnableProxy:       false,
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
