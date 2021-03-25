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

	"github.com/PuerkitoBio/goquery"
	"github.com/voiladev/VoilaCrawl/pkg/crawler"
	"github.com/voiladev/VoilaCrawl/pkg/net/http"
	"github.com/voiladev/VoilaCrawl/pkg/net/http/cookiejar"
	"github.com/voiladev/VoilaCrawl/pkg/proxy"
	"github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	pbProxy "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
	"google.golang.org/protobuf/types/known/anypb"
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
		categoryPathMatcher: regexp.MustCompile(`^(/[a-z0-9_\-]+)?/shop(/[a-zA-Z0-9\-]+){1,4}(/Pageindex/\d+)?$`),
		productPathMatcher:  regexp.MustCompile(`^(/[a-z0-9_\-]+)?/shop/product/([/a-z0-9_\-]+)$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "b790c41a7de54f33b698170352afe0e1"
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
	options.EnableSessionInit = false
	options.Reliability = pbProxy.ProxyReliability_ReliabilityMedium

	// options.MustCookies = append(options.MustCookies,
	// 	&http.Cookie{Name: "shippingCountry", Value: "US", Path: "/"},
	// 	&http.Cookie{Name: "currency", Value: "USD", Path: "/"},
	// )
	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.macys.com"}
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
	prodDataExtraReg      = regexp.MustCompile(`(data-bootstrap="page/discovery-pages" type="application/json">)([^<]+)</script>`)
	prodDataPaginationReg = regexp.MustCompile(`(data-bootstrap="feature/canvas"  type="application/json">)([^<]+)</script>`)
)

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
	sel := dom.Find(`.items > .productThumbnailItem`)

	lastIndex := nextIndex(ctx)
	for i := range sel.Nodes {
		node := sel.Eq(i)

		detailUrl := node.Find(".productDescription>a").AttrOr("href", "")
		if detailUrl == "" {
			continue
		}
		req, err := http.NewRequest(http.MethodGet, detailUrl, nil)
		if err != nil {
			c.logger.Errorf("invalud product detail url %s", detailUrl)
		}
		req.Header.Set("Referer", resp.Request.URL.String())

		nctx := context.WithValue(ctx, "item.index", lastIndex)
		lastIndex += 1

		if err := yield(nctx, req); err != nil {
			return err
		}
	}

	var pagination struct {
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
	pRawData := strings.TrimSpace(dom.Find(`script[data-bootstrap="feature/canvas"]`).Text())
	if err := json.Unmarshal([]byte(pRawData), &pagination); err != nil {
		c.logger.Errorf("unmarshal pagination info %s failed, error=%s", respBody, err)
		return err
	}
	if pagination.Model.Pagination.NextURL != "" {
		req, _ := http.NewRequest(http.MethodGet, pagination.Model.Pagination.NextURL, nil)
		req.Header.Set("Referer", resp.Request.URL.String())
		nctx := context.WithValue(ctx, "item.index", lastIndex)
		return yield(nctx, req)
	}
	return nil
}

type parseProductResponse struct {
	Properties struct {
		ASSETHOST string `json:"ASSET_HOST"`
		Recaptcha struct {
			ScriptURL string `json:"scriptUrl"`
			SiteKey   string `json:"siteKey"`
		} `json:"recaptcha"`
	} `json:"properties"`
	ISCLIENTLOGSENABLED bool   `json:"_IS_CLIENT_LOGS_ENABLED"`
	PDPBOOTSTRAPDATA    string `json:"_PDP_BOOTSTRAP_DATA"`
}

type parseProductData struct {
	UtagData struct {
		ProductRating  []string `json:"product_rating"`
		ProductReviews []string `json:"product_reviews"`
	} `json:"utagData"`
	Product struct {
		ID         int `json:"id"`
		Identifier struct {
			ProductURL           string `json:"productUrl"`
			ProductID            int    `json:"productId"`
			TopLevelCategoryID   string `json:"topLevelCategoryID"`
			TopLevelCategoryName string `json:"topLevelCategoryName"`
		} `json:"identifier"`
		Messages struct {
			Info []struct {
			} `json:"info"`
		} `json:"messages"`
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
				StyleMeEnabled                             bool `json:"styleMeEnabled"`
				ZeekitEnabled                              bool `json:"zeekitEnabled"`
				AltModelSizesExperimentEnabled             bool `json:"altModelSizesExperimentEnabled"`
				SeeMoreAndSizeChartExperienceEnabled       bool `json:"seeMoreAndSizeChartExperienceEnabled"`
				SizeChartExperienceEnabled                 bool `json:"sizeChartExperienceEnabled"`
				GwpExperienceEnabled                       bool `json:"gwpExperienceEnabled"`
				BcomsyndigoEnabled                         bool `json:"bcomsyndigoEnabled"`
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
				IsBVAPIExpArmOneEnabled                    bool `json:"isBVApiExpArmOneEnabled"`
				PdpExpansionEnabled                        bool `json:"pdpExpansionEnabled"`
				IsWebCollageOutFromTabs                    bool `json:"isWebCollageOutFromTabs"`
				IsDressSizeSelectorsEnabled                bool `json:"isDressSizeSelectorsEnabled"`
				IsTrueCollectionsEnabled                   bool `json:"isTrueCollectionsEnabled"`
				IsReviewPhotosUploadEnabled                bool `json:"isReviewPhotosUploadEnabled"`
				IsVideoImageRailEnabled                    bool `json:"isVideoImageRailEnabled"`
				IsReviewsPage                              bool `json:"isReviewsPage"`
				IsBVAPIExpArmTwoEnabled                    bool `json:"isBVApiExpArmTwoEnabled"`
				IsNew                                      bool `json:"isNew"`
				IsFindationEnabled                         bool `json:"isFindationEnabled"`
			} `json:"flags"`
			ReviewStatistics struct {
				Aggregate struct {
					Rating           float64 `json:"rating"`
					RatingPercentage int     `json:"ratingPercentage"`
					Count            int     `json:"count"`
				} `json:"aggregate"`
			} `json:"reviewStatistics"`
			QuestionAnswer struct {
				QuestionCount int `json:"questionCount"`
				AnswerCount   int `json:"answerCount"`
			} `json:"questionAnswer"`
			OrderedMasterGroupList []interface{} `json:"orderedMasterGroupList"`
			MemberDisplayGroupsMap struct {
			} `json:"memberDisplayGroupsMap"`
			BulletText            []string `json:"bulletText"`
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
			BulletLinks         []interface{} `json:"bulletLinks"`
			PdfEmailDescription string        `json:"pdfEmailDescription"`
			MemberProductCount  int           `json:"memberProductCount"`
			CompleteName        string        `json:"completeName"`
			ProcessedProdDesc   struct {
				ProductDetails []string `json:"productDetails"`
				SizeAndFit     []string `json:"sizeAndFit"`
				FabricAndCare  []string `json:"fabricAndCare"`
			} `json:"processedProdDesc"`
			Metric struct {
				ProductUnitSalesCount        string `json:"productUnitSalesCount"`
				ProductUnitSalesCountMessage string `json:"productUnitSalesCountMessage"`
			} `json:"metric"`
			Klarna struct {
				KlarnaDataClientID    string `json:"klarnaDataClientId"`
				KlarnaOnsiteJsSdkPath string `json:"klarnaOnsiteJsSdkPath"`
			} `json:"klarna"`
		} `json:"detail"`
		Shipping struct {
			ReturnConstraintMessage string   `json:"returnConstraintMessage"`
			Notes                   []string `json:"notes"`
			FreeShippingMessages    []string `json:"freeShippingMessages"`
		} `json:"shipping"`
		Relationships struct {
			Taxonomy struct {
				Categories []struct {
					Name string `json:"name"`
					URL  string `json:"url"`
					Type string `json:"type"`
					ID   int    `json:"id"`
				} `json:"categories"`
				DefaultCategoryID int `json:"defaultCategoryId"`
			} `json:"taxonomy"`
			Upcs map[string]struct {
				//Num44742859 struct {
				ID         int `json:"id"`
				Identifier struct {
					UpcNumber string `json:"upcNumber"`
				} `json:"identifier"`
				Department struct {
					DepartmentID   int    `json:"departmentId"`
					DepartmentName string `json:"departmentName"`
				} `json:"department"`
				ClassCode     int    `json:"classCode"`
				SubClassCode  int    `json:"subClassCode"`
				VendorCode    int    `json:"vendorCode"`
				MarkStyleCode string `json:"markStyleCode"`
				Messages      struct {
					Info []struct {
					} `json:"info"`
				} `json:"messages"`
				Relationships struct {
				} `json:"relationships"`
				Availability struct {
					CheckInStoreEligibility bool   `json:"checkInStoreEligibility"`
					Available               bool   `json:"available"`
					ShipDays                int    `json:"shipDays"`
					Message                 string `json:"message"`
					AvailabilityMessage     string `json:"availabilityMessage"`
					OrderType               string `json:"orderType"`
					BopsAvailability        bool   `json:"bopsAvailability"`
					BossAvailability        bool   `json:"bossAvailability"`
					StoreAvailability       bool   `json:"storeAvailability"`
				} `json:"availability"`
				Traits struct {
					Colors struct {
						SelectedColor int `json:"selectedColor"`
					} `json:"colors"`
					Sizes struct {
						SelectedSize int `json:"selectedSize"`
					} `json:"sizes"`
				} `json:"traits"`
				ProtectionPlans        []interface{} `json:"protectionPlans"`
				HolidayMessageEligible bool          `json:"holidayMessageEligible"`
				//} `json:"44742859"`

			} `json:"upcs"`
		} `json:"relationships"`
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
			ItemQty                        int  `json:"itemQty"`
			HasImagesRail                  bool `json:"hasImagesRail"`
			ApplyFixForManyAltImagesMobile bool `json:"applyFixForManyAltImagesMobile"`
		} `json:"imagery"`
		Availability struct {
			CheckInStoreEligibility bool `json:"checkInStoreEligibility"`
			Available               bool `json:"available"`
			BopsAvailability        bool `json:"bopsAvailability"`
			BossAvailability        bool `json:"bossAvailability"`
			StoreAvailability       bool `json:"storeAvailability"`
		} `json:"availability"`
		Traits struct {
			Colors struct {
				SelectedColor int `json:"selectedColor"`
				ColorMap      map[string]struct {
					//Num6278243 struct {
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
							FinalPrice  struct {
								Label  string `json:"label"`
								Values []struct {
									Value          float64 `json:"value"`
									FormattedValue string  `json:"formattedValue"`
									Type           string  `json:"type"`
								} `json:"values"`
								MaskPromotion        bool     `json:"maskPromotion"`
								ApplicablePromotions []string `json:"applicablePromotions"`
								PromoCode            string   `json:"promoCode"`
							} `json:"finalPrice"`
						} `json:"price"`
						BadgeIds []string `json:"badgeIds"`
					} `json:"pricing"`
					//} `json:"6278243"`
				} `json:"colorMap"`
				SwatchSprite struct {
					SwatchSpriteUrls  []string `json:"swatchSpriteUrls"`
					SpriteSwatchSize  int      `json:"spriteSwatchSize"`
					SwatchesPerSprite int      `json:"swatchesPerSprite"`
				} `json:"swatchSprite"`
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
				UpcMap        map[string]int `json:"upcMap"`
				PriceToColors []struct {
					Price    string `json:"price"`
					ColorIds []int  `json:"colorIds"`
					OnSale   bool   `json:"onSale"`
				} `json:"priceToColors"`
			} `json:"traitsMaps"`
		} `json:"traits"`
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
				FinalPrice  struct {
					Label  string `json:"label"`
					Values []struct {
						Value          float64 `json:"value"`
						FormattedValue string  `json:"formattedValue"`
						Type           string  `json:"type"`
					} `json:"values"`
					MaskPromotion        bool     `json:"maskPromotion"`
					ApplicablePromotions []string `json:"applicablePromotions"`
					PromoCode            string   `json:"promoCode"`
				} `json:"finalPrice"`
			} `json:"price"`
			BadgesMap map[string]struct {
				//Num19909501 struct {
				WalletEligible          bool   `json:"walletEligible"`
				CheckoutDescription     string `json:"checkoutDescription"`
				Description             string `json:"description"`
				PromoID                 string `json:"promoId"`
				Header                  string `json:"header"`
				ApplicableToAllUpcs     bool   `json:"applicableToAllUpcs"`
				Offer                   string `json:"offer"`
				PromotionType           string `json:"promotionType"`
				HasMorePromotionDetails bool   `json:"hasMorePromotionDetails"`
				//} `json:"19909501"`
			} `json:"badgesMap"`
			BadgeIds []string `json:"badgeIds"`
		} `json:"pricing"`
		Review struct {
			HasErrors bool `json:"hasErrors"`
			Reviews   []struct {
				ReviewID           int     `json:"reviewId"`
				Rating             float32 `json:"rating"`
				Title              string  `json:"title"`
				ReviewText         string  `json:"reviewText"`
				TopContributor     bool    `json:"topContributor"`
				Anonymous          bool    `json:"anonymous"`
				DisplayName        string  `json:"displayName"`
				IncentivizedReview bool    `json:"incentivizedReview"`
			} `json:"reviews"`
		} `json:"review"`
		ProtectionPlans []interface{} `json:"protectionPlans"`
		URLTemplate     struct {
			Swatch       string `json:"swatch"`
			SwatchSprite string `json:"swatchSprite"`
			Product      string `json:"product"`
		} `json:"urlTemplate"`
		HolidayMessageEligible bool `json:"holidayMessageEligible"`
		Seotags                struct {
			Seotags string `json:"seotags"`
		} `json:"seotags"`
	} `json:"product"`
}

var (
	detailReg = regexp.MustCompile(`(?U)<script[^>]*>\s*window.__INITIAL_STATE__\s*=\s*({.*});?\s*</script>`)
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

	c.logger.Debugf("json %s", matched[1])

	var (
		i  parseProductResponse
		pd parseProductData
	)

	if err = json.Unmarshal(matched[1], &i); err != nil {
		c.logger.Error(err)
		return err
	}

	if err = json.Unmarshal([]byte(i.PDPBOOTSTRAPDATA), &pd); err != nil {
		c.logger.Error(err)
		return err
	}

	var (
		reviewCount int64
		rating      float64
	)

	if len(pd.UtagData.ProductReviews) > 0 {
		reviewCount, _ = strconv.ParseInt(pd.UtagData.ProductReviews[0])
	}
	if len(pd.UtagData.ProductRating) > 0 {
		rating, _ = strconv.ParseFloat(pd.UtagData.ProductRating[0])
	}

	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:       strconv.Format(pd.Product.ID),
			CrawlUrl: resp.Request.URL.String(),
		},
		Title:       pd.Product.Detail.Name,
		Description: pd.Product.Detail.Description,
		BrandName:   pd.Product.Detail.Brand.Name,
		// CrowdType:    i.Details.GenderName,  // ASK ?
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stats: &pbItem.Stats{
			ReviewCount: int32(reviewCount),
			Rating:      float32(rating),
		},
	}

	for _, p := range pd.Product.Traits.Colors.ColorMap {
		current, _ := strconv.ParseFloat(p.Pricing.Price.FinalPrice.Values[0].Value)
		msrp, _ := strconv.ParseFloat(p.Pricing.Price.TieredPrice[1].Values[0].Value)
		// discount := strconv.ParseFloat(p.Pricing.Price.TieredPrice[1].Values[0])
		discount := (msrp - current) * 100 / msrp

		for _, rawSize := range p.Sizes {

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
			// if rawSize.StockLevelStatus == "inStock" {
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

			item.SkuItems = append(item.SkuItems, &sku)
		}

		isDefault := true
		for key, img := range p.Imagery.Images {
			if key > 1 {
				isDefault = false
			}
			itemImg, _ := anypb.New(&media.Media_Image{ // ask?
				OriginalUrl: "https://slimages.macysassets.com/is/image/MCY/products/" + img.FilePath,
				LargeUrl:    "https://slimages.macysassets.com/is/image/MCY/products/" + img.FilePath + "?wid=1230&hei=1500&op_sharpen=1", // $S$, $XXL$
				MediumUrl:   "https://slimages.macysassets.com/is/image/MCY/products/" + img.FilePath + "?wid=640&hei=780&op_sharpen=1",
				SmallUrl:    "https://slimages.macysassets.com/is/image/MCY/products/" + img.FilePath + "?wid=500&hei=609&op_sharpen=1",
			})
			item.Medias = append(item.Medias, &media.Media{
				Detail:    itemImg,
				IsDefault: isDefault,
			})
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
		// "https://www.macys.com/shop/womens-clothing/womens-sale-clearance?id=10066",
		"https://www.macys.com/shop/product/style-co-ribbed-hoodie-sweater-created-for-macys?ID=11393511&CategoryID=10066",
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
	opts := spider.CrawlOptions()

	// this callback func is used to do recursion call of sub requests.
	var callback func(ctx context.Context, val interface{}) error
	callback = func(ctx context.Context, val interface{}) error {
		switch i := val.(type) {
		case *http.Request:
			logger.Debugf("Access %s", i.URL)

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
				i.URL.Host = "www.macys.com"
			}

			// do http requests here.
			nctx, cancel := context.WithTimeout(ctx, time.Minute*5)
			defer cancel()

			// nctx = context.WithValue(nctx, "tracing_id", randutil.MustNewRandomID())
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

	ctx := context.WithValue(context.Background(), "tracing_id", fmt.Sprintf("macy_%d", time.Now().UnixNano()))
	// start the crawl request
	for _, req := range spider.NewTestRequest(context.Background()) {
		if err := callback(ctx, req); err != nil {
			logger.Fatal(err)
		}
	}
}
