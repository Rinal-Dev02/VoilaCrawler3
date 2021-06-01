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

	// pbProxy "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
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
		categoryPathMatcher: regexp.MustCompile(`^/shop(/[a-z0-9\pL\pS._\-]+){2,6}$`),
		productPathMatcher:  regexp.MustCompile(`^/shop/product/[a-z0-9\pL\pS._\-]+$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "f330e29cf7fb7dc313fd101fde1d5aa5"
}

// Version
func (c *_Crawler) Version() int32 {
	return 1
}

// CrawlOptions
func (c *_Crawler) CrawlOptions(u *url.URL) *crawler.CrawlOptions {
	options := crawler.NewCrawlOptions()
	options.EnableHeadless = false
	options.EnableSessionInit = false
	options.DisableCookieJar = true
	// options.Reliability = pbProxy.ProxyReliability_ReliabilityMedium
	options.MustHeader.Set("accept-encoding", "gzip, deflate, br")
	options.MustCookies = append(options.MustCookies,
		&http.Cookie{Name: "currency", Value: `USD`, Path: "/"},
		&http.Cookie{Name: "shippingCountry", Value: `US`, Path: "/"},
		&http.Cookie{Name: "mercury", Value: `false`, Path: "/"},
	)

	if u != nil {
		// options.MustCookies = append(options.MustCookies, &http.Cookie{
		// 	Name:  "FORWARDPAGE_KEY",
		// 	Value: url.QueryEscape(u.String()),
		// })
		// options.MustHeader.Set("Referer", u.String())
	}
	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.bloomingdales.com"}
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
		u.Host = "www.bloomingdales.com"
	}

	if c.productPathMatcher.MatchString(u.Path) {
		u.RawQuery = fmt.Sprintf("ID=%s", u.Query().Get("ID"))
		return u.String(), nil
	}
	return rawurl, nil
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
	return crawler.ErrUnsupportedPath
}

// nextIndex used to get sharingData from context
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")) + 1)
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

	sel := dom.Find(".items>.cell")
	if len(sel.Nodes) == 0 {
		return errors.New("no product found")
	}
	for i := range sel.Nodes {
		node := sel.Eq(i)
		href := node.Find(`.productThumbnail .productDescLink`).AttrOr("href", "")

		req, err := http.NewRequest(http.MethodGet, href, nil)
		if err != nil {
			c.logger.Errorf("load http request of url %s failed, error=%s", href, err)
			return err
		}
		req.Header.Set("Referer", resp.Request.URL.String())
		req.AddCookie(&http.Cookie{Name: "FORWARDPAGE_KEY", Value: url.QueryEscape(resp.Request.URL.String())})
		nctx := context.WithValue(ctx, "item.index", lastIndex)
		lastIndex++

		// yield sub request
		if err := yield(nctx, req); err != nil {
			return err
		}
	}

	pageIndex := dom.Find(`#sort-pagination-select-bottom > option[selected="selected"] + option`).AttrOr("value", "")
	if pageIndex == "" {
		return nil
	}

	u := *resp.Request.URL
	fields := strings.Split(u.Path, "/")
	if len(fields) > 3 && fields[len(fields)-2] == "Pageindex" {
		fields[len(fields)-1] = pageIndex
	} else {
		fields = append(fields, "Pageindex", pageIndex)
	}
	u.Path = strings.Join(fields, "/")

	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	req.Header.Set("Referer", resp.Request.URL.String())
	req.AddCookie(&http.Cookie{Name: "FORWARDPAGE_KEY", Value: url.QueryEscape(resp.Request.URL.String())})

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
		Relationships struct {
			Taxonomy struct {
				Categories []struct {
					Name string `json:"name"`
					URL  string `json:"url"`
					ID   int    `json:"id"`
				} `json:"categories"`
				DefaultCategoryID int `json:"defaultCategoryId"`
			} `json:"taxonomy"`
			Upcs map[string]struct {
				ID         int `json:"id"`
				Identifier struct {
					UpcNumber string `json:"upcNumber"`
				} `json:"identifier"`
				Availability struct {
					CheckInStoreEligibility bool   `json:"checkInStoreEligibility"`
					Available               bool   `json:"available"`
					ShipDays                int    `json:"shipDays"`
					Message                 string `json:"message"`
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
			} `json:"upcs"`
		} `json:"relationships"`

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
					ID          int    `json:"id"`
					Name        string `json:"name"`
					DisplayName string `json:"displayName"`
					Colors      []int  `json:"colors"`
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
	} `json:"product"`
	UtagData struct {
		ProductRating  []string `json:"product_rating"`
		ProductReviews []string `json:"product_reviews"`
	} `json:"utagData"`
}

var (
	detailReg = regexp.MustCompile(`(?U)<script\s+data-bootstrap="page/product"\s*type="application/json">({.*})</script>`)
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

	if err = json.Unmarshal(matched[1], &pd); err != nil {
		c.logger.Error(err)
		return err
	}
	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	reviewCount, _ := strconv.ParseFloat(pd.UtagData.ProductReviews[0])
	rating, _ := strconv.ParseFloat(pd.UtagData.ProductRating[0])

	canUrl := dom.Find(`link[rel="canonical"]`).AttrOr("href", "")
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}

	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           strconv.Format(pd.Product.ID),
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
		},
		Title:       pd.Product.Detail.Name,
		Description: pd.Product.Detail.Description,
		BrandName:   pd.Product.Detail.Brand.Name,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stats: &pbItem.Stats{
			ReviewCount: int32(reviewCount),
			Rating:      float32(rating),
		},
	}
	for i, cate := range pd.Product.Relationships.Taxonomy.Categories {
		switch i {
		case 0:
			item.Category = cate.Name
		case 1:
			item.SubCategory = cate.Name
		case 2:
			item.SubCategory2 = cate.Name
		case 3:
			item.SubCategory3 = cate.Name
		case 4:
			item.SubCategory4 = cate.Name
		}
	}
	for skuId, rawSku := range pd.Product.Relationships.Upcs {
		colorId := strconv.Format(rawSku.Traits.Colors.SelectedColor)
		color := pd.Product.Traits.Colors.ColorMap[colorId]
		sizeId := strconv.Format(rawSku.Traits.Sizes.SelectedSize)
		size := pd.Product.Traits.Sizes.SizeMap[sizeId]

		var (
			current, msrp, discount float64
		)
		for _, p := range color.Pricing.Price.TieredPrice {
			if len(p.Values) == 0 {
				continue
			}
			if current == 0 && msrp == 0 {
				current, msrp = p.Values[0].Value, p.Values[0].Value
			} else if p.Values[0].Value > msrp {
				msrp = p.Values[0].Value
			} else if p.Values[0].Value < current {
				current = p.Values[0].Value
			}
		}
		if msrp == 0 {
			return fmt.Errorf("no msrp price found for %s", resp.Request.URL)
		}
		discount = math.Ceil((msrp - current) * 100 / msrp)

		var medias []*pbMedia.Media
		for key, img := range color.Imagery.Images {
			medias = append(medias, pbMedia.NewImageMedia(
				"",
				"https://images.bloomingdalesassets.com/is/image/BLM/products/"+img.FilePath,
				"https://images.bloomingdalesassets.com/is/image/BLM/products/"+img.FilePath+"?op_sharpen=1&wid=1000",
				"https://images.bloomingdalesassets.com/is/image/BLM/products/"+img.FilePath+"?op_sharpen=1&wid=700",
				"https://images.bloomingdalesassets.com/is/image/BLM/products/"+img.FilePath+"?op_sharpen=1&wid=450",
				"",
				key == 0,
			))
		}

		sku := pbItem.Sku{
			SourceId: skuId,
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(current * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},
			Medias: medias,
			Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_InStock},
		}
		// if rawSize.StockLevelStatus == "inStock" {  // ASk ?
		// 	sku.Stock.StockStatus = pbItem.Stock_InStock
		// 	//sku.Stock.StockCount = int32(rawSize.Quantity)
		// }

		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecColor,
			Id:    colorId,
			Name:  color.NormalName,
			Value: color.NormalName,
		})
		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Id:    sizeId,
			Name:  size.DisplayName,
			Value: size.Name,
		})
		item.SkuItems = append(item.SkuItems, &sku)
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
		// "https://www.bloomingdales.com/shop/product/aqua-passion-sleeveless-maxi-dress-100-exclusive?ID=3996369&CategoryID=21683",
		// "https://www.bloomingdales.com/shop/product/a.l.c.-kati-puff-sleeve-tee?ID=3202505&CategoryID=5619",
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
	cli.NewApp(New).Run(os.Args)
}
