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
	 "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/media"
	 "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/regulation"
	 pbItem "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
	"google.golang.org/protobuf/types/known/anypb"
)

type _Crawler struct {
	httpClient http.Client

	categoryPathMatcher     *regexp.Regexp
	productPathMatcher      *regexp.Regexp
	logger                  glog.Log
}


func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient:              client,
		categoryPathMatcher:     regexp.MustCompile(`^(/[a-z0-9_-]+)?/shopping/(women|men)([/a-z0-9_-]+)items.aspx$`),
		productPathMatcher:      regexp.MustCompile(`^(/[a-z0-9_-]+)+(/[a-z0-9_-]+)-item-[0-9]+.aspx$`),
		logger:                  logger.New("_Crawler"),
	}
	return &c, nil
}
// ID
func (c *_Crawler) ID() string {
	return "350d1122d8d2ae45b9e0dc3255f7102f"
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
		//&http.Cookie{Name: "geocountry", Value: `US`, Path: "/"},
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
	return []string{"www.farfetch.com"}
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
		//return c.parseProduct(ctx, resp, yield)
		
	}
	
	return fmt.Errorf("unsupported url %s", resp.Request.URL.String())
}

// nextIndex used to get sharingData from context
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")) + 1)
}

var prodDataExtraReg1 = regexp.MustCompile(`(window\['__initialState__']) = "([^;)]+)";`)
var prodDataExtraReg = regexp.MustCompile(`(window\['__initialState_portal-slices-listing__'\])\s*=\s*({.*})?</script>`)

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

	 // write the whole body at once
	//  err = ioutil.WriteFile("C:\\Rinal\\ServiceBasedPRojects\\VoilaWork_new\\VoilaCrawl\\output.txt", respBody, 0644)
	//  if err != nil {
	// 	 panic(err)
	//  }

	// next page
	matched := prodDataExtraReg.FindSubmatch(respBody)
	 if matched == nil {
		matched = prodDataExtraReg1.FindSubmatch(respBody) //__initialState__
	 }
	if len(matched) <= 1 {
		return fmt.Errorf("extract json from product list page %s failed", resp.Request.URL)
	}
	var r struct {
		ListingHeader struct {
			Type        string `json:"type"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Breadcrumb  []struct {
				Text string `json:"text"`
				Href string `json:"href"`
			} `json:"breadcrumb"`
			ShouldRequestTitle bool        `json:"shouldRequestTitle"`
			HomologousBrands   interface{} `json:"homologousBrands"`
		} `json:"listingHeader"`
		Listing struct {
			Sort               string      `json:"sort"`
			SearchTerm         interface{} `json:"searchTerm"`
			PreferredDesigners struct {
				Type      string      `json:"type"`
				Designers interface{} `json:"designers"`
			} `json:"preferredDesigners"`
			Designer interface{} `json:"designer"`
			Category struct {
				ID       string      `json:"id"`
				URLToken string      `json:"urlToken"`
				Name     interface{} `json:"name"`
			} `json:"category"`
			Store      interface{} `json:"store"`
			Categories []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"categories"`
			MoreProducts     string      `json:"moreProducts"`
			SetID            interface{} `json:"setId"`
			PromotionID      interface{} `json:"promotionId"`
			LabelID          interface{} `json:"labelId"`
			SimilarProductID interface{} `json:"similarProductId"`
			SeoProductID     interface{} `json:"seoProductId"`
			FfMedium         string      `json:"ffMedium"`
			UtmMedium        string      `json:"utmMedium"`
			IsMixMode        bool        `json:"isMixMode"`
			IsExclusive      bool        `json:"isExclusive"`
			Strategy         interface{} `json:"strategy"`
			RecommendedPage  struct {
				ProductID       interface{} `json:"productId"`
				Referrer        interface{} `json:"referrer"`
				EncodedStrategy interface{} `json:"encodedStrategy"`
			} `json:"recommendedPage"`
		} `json:"listing"`
		Labels struct {
			Piece                               string      `json:"piece"`
			Pieces                              string      `json:"pieces"`
			ShowPiece                           string      `json:"showPiece"`
			ShowPieces                          string      `json:"showPieces"`
			View                                string      `json:"view"`
			PaginationOf                        string      `json:"paginationOf"`
			BackToTop                           string      `json:"backToTop"`
			LooksCloseTo                        string      `json:"looksCloseTo"`
			Customize                           string      `json:"customize"`
			PaginationPage                      string      `json:"paginationPage"`
			PaginationPrevious                  string      `json:"paginationPrevious"`
			PaginationNext                      string      `json:"paginationNext"`
			NoResultsFoundFor                   string      `json:"noResultsFoundFor"`
			SearchDesigners                     string      `json:"searchDesigners"`
			AvailableIn                         string      `json:"availableIn"`
			SeeSimilar                          string      `json:"seeSimilar"`
			Sort                                string      `json:"sort"`
			Refine                              string      `json:"refine"`
			More                                string      `json:"more"`
			Less                                string      `json:"less"`
			Apply                               string      `json:"apply"`
			FilterValueAll                      string      `json:"filterValueAll"`
			SameDayTo                           string      `json:"sameDayTo"`
			NinetyMinutesTo                     string      `json:"ninetyMinutesTo"`
			Filters                             string      `json:"filters"`
			Done                                string      `json:"done"`
			SortBy                              string      `json:"sortBy"`
			Discover                            string      `json:"discover"`
			PiecesInSale                        string      `json:"piecesInSale"`
			PieceInSale                         string      `json:"pieceInSale"`
			FullPriceSingular                   string      `json:"fullPriceSingular"`
			FullPricePlural                     string      `json:"fullPricePlural"`
			FilterNoResultsTitle                string      `json:"filterNoResultsTitle"`
			FilterNoResultsDescription          string      `json:"filterNoResultsDescription"`
			ClearAll                            string      `json:"clearAll"`
			UnavailableDesigners                string      `json:"unavailableDesigners"`
			Women                               string      `json:"women"`
			Men                                 string      `json:"men"`
			Kids                                string      `json:"kids"`
			ShopWomen                           string      `json:"shopWomen"`
			ShopMen                             string      `json:"shopMen"`
			ShopKids                            interface{} `json:"shopKids"`
			Popular                             string      `json:"popular"`
			FavouriteTabTitle                   string      `json:"favouriteTabTitle"`
			DesignerAZ                          string      `json:"designerAZ"`
			Featured                            string      `json:"featured"`
			ShopNow                             string      `json:"shopNow"`
			FavouriteDesigners                  string      `json:"favouriteDesigners"`
			AllFavouriteDesigners               string      `json:"allFavouriteDesigners"`
			FavouriteDesignerUpdateButton       string      `json:"favouriteDesignerUpdateButton"`
			FavouriteDesignerUpdateTitle        string      `json:"favouriteDesignerUpdateTitle"`
			LoadNextButton                      string      `json:"loadNextButton"`
			LoadPreviousButton                  string      `json:"loadPreviousButton"`
			Progress                            string      `json:"progress"`
			PrivateClientExclusives             string      `json:"privateClientExclusives"`
			SizeProfileYourSizeFitProfile       string      `json:"sizeProfileYourSizeFitProfile"`
			SizeProfileDescription              string      `json:"sizeProfileDescription"`
			SizeProfileCreateProfile            string      `json:"sizeProfileCreateProfile"`
			SizeProfileYourFFitProfile          string      `json:"sizeProfileYourFFitProfile"`
			SizeProfileCreateProfileTooltip     string      `json:"sizeProfileCreateProfileTooltip"`
			SizeProfileTitleModal               string      `json:"sizeProfileTitleModal"`
			SizeProfileSubTitleModal            string      `json:"sizeProfileSubTitleModal"`
			SizeProfileSelectYourSizes          string      `json:"sizeProfileSelectYourSizes"`
			SizeProfileSelectYourPreferredSizes string      `json:"sizeProfileSelectYourPreferredSizes"`
			SizeProfileSaveAndApplyFilters      string      `json:"sizeProfileSaveAndApplyFilters"`
			SizeProfileFFitProfile              string      `json:"sizeProfileFFitProfile"`
			TryAgain                            string      `json:"tryAgain"`
			ErrorLoadingSizeProfile             string      `json:"errorLoadingSizeProfile"`
			PleaseRefreshOrTryAgainLater        string      `json:"pleaseRefreshOrTryAgainLater"`
			GotIt                               string      `json:"gotIt"`
			Available                           string      `json:"available"`
			SeeAllSizes                         string      `json:"seeAllSizes"`
			SimilarDesignersForYou              string      `json:"similarDesignersForYou"`
			DutiesAndTaxesIncluded              string      `json:"dutiesAndTaxesIncluded"`
			PriceInputMin                       string      `json:"priceInputMin"`
			PriceInputMax                       string      `json:"priceInputMax"`
			PriceInputMinErrorMessage           string      `json:"priceInputMinErrorMessage"`
			PriceInputMaxErrorMessage           string      `json:"priceInputMaxErrorMessage"`
			PriceInputClear                     string      `json:"priceInputClear"`
			QuickView                           string      `json:"quickView"`
			ErrorLoadingProductTitle            string      `json:"errorLoadingProductTitle"`
			SizeProfileOnBoardTitle             string      `json:"sizeProfileOnBoardTitle"`
			SizeProfileOnBoardDescription       string      `json:"sizeProfileOnBoardDescription"`
			Cancel                              string      `json:"cancel"`
			DeliveryOptions                     string      `json:"deliveryOptions"`
			DeliveryInNinetyMinutes             string      `json:"deliveryInNinetyMinutes"`
			DeliverySameDay                     string      `json:"deliverySameDay"`
			Filter                              string      `json:"filter"`
		} `json:"labels"`
		Subfolder   string `json:"subfolder"`
		SortOptions []struct {
			Value string `json:"value"`
			Label string `json:"label"`
		} `json:"sortOptions"`
		Defaults struct {
			Page                   int    `json:"page"`
			Sort                   string `json:"sort"`
			FavouriteDesignersSort string `json:"favouriteDesignersSort"`
		} `json:"defaults"`
		PageType    string `json:"pageType"`
		Path        string `json:"path"`
		FlatContent struct {
		} `json:"flatContent"`
		Gender      string `json:"gender"`
		GenderID    int    `json:"genderId"`
		CultureInfo struct {
			CountryCultureCode      string `json:"countryCultureCode"`
			CurrencyCode            string `json:"currencyCode"`
			CurrencyPositivePattern int    `json:"currencyPositivePattern"`
			CurrencySymbol          string `json:"currencySymbol"`
		} `json:"cultureInfo"`
		PriceInfo struct {
			Max                   int         `json:"max"`
			MaxBoundary           int         `json:"maxBoundary"`
			MinBoundary           int         `json:"minBoundary"`
			Description           string      `json:"description"`
			Filter                interface{} `json:"filter"`
			IsServerSideRendering bool        `json:"isServerSideRendering"`
		} `json:"priceInfo"`
		SeoLinks        []interface{} `json:"seoLinks"`
		PriceType       string        `json:"priceType"`
		RenderDirection string        `json:"renderDirection"`
		Toggles         struct {
			ShowProductsCount          bool `json:"showProductsCount"`
			IntegratedContentEnabled   bool `json:"integratedContentEnabled"`
			IsSizeProfileToggleEnabled bool `json:"isSizeProfileToggleEnabled"`
			SimilarDesignersToggle     bool `json:"similarDesignersToggle"`
			IsNewPaginationMechanism   bool `json:"isNewPaginationMechanism"`
			FiltersHide                bool `json:"filtersHide"`
		} `json:"toggles"`
		AbTests struct {
		} `json:"abTests"`
		HomologousBrands interface{} `json:"homologousBrands"`
		AdsInfo          struct {
			CriteoAdsAPIURL       string   `json:"criteoAdsApiUrl"`
			CriteoTrackingDomains []string `json:"criteoTrackingDomains"`
			StaticImagesBaseURI   string   `json:"staticImagesBaseUri"`
			Installments          int      `json:"installments"`
		} `json:"adsInfo"`
		SimilarDesigners struct {
			Threshold         int         `json:"threshold"`
			ShouldBeDisplayed interface{} `json:"shouldBeDisplayed"`
		} `json:"similarDesigners"`
		SizeProfile struct {
			CombinedCategoriesMapping struct {
				Num135979 []int `json:"135979"`
				Num135981 []int `json:"135981"`
				Num135983 []int `json:"135983"`
				Num135985 []int `json:"135985"`
				Num136021 []int `json:"136021"`
				Num136045 []int `json:"136045"`
				Num136071 []int `json:"136071"`
				Num136089 []int `json:"136089"`
				Num136091 []int `json:"136091"`
				Num136093 []int `json:"136093"`
				Num136099 []int `json:"136099"`
				Num136101 []int `json:"136101"`
				Num136103 []int `json:"136103"`
				Num136105 []int `json:"136105"`
				Num136107 []int `json:"136107"`
				Num136137 []int `json:"136137"`
				Num136147 []int `json:"136147"`
				Num136149 []int `json:"136149"`
				Num136157 []int `json:"136157"`
				Num136175 []int `json:"136175"`
				Num136177 []int `json:"136177"`
				Num136179 []int `json:"136179"`
				Num136181 []int `json:"136181"`
				Num136183 []int `json:"136183"`
				Num136185 []int `json:"136185"`
				Num136187 []int `json:"136187"`
				Num136189 []int `json:"136189"`
				Num136191 []int `json:"136191"`
				Num136193 []int `json:"136193"`
				Num136195 []int `json:"136195"`
				Num136216 []int `json:"136216"`
				Num136217 []int `json:"136217"`
				Num136218 []int `json:"136218"`
				Num136220 []int `json:"136220"`
				Num136221 []int `json:"136221"`
				Num136222 []int `json:"136222"`
				Num136223 []int `json:"136223"`
				Num136224 []int `json:"136224"`
				Num136225 []int `json:"136225"`
				Num136226 []int `json:"136226"`
				Num136227 []int `json:"136227"`
				Num136228 []int `json:"136228"`
				Num136229 []int `json:"136229"`
				Num136230 []int `json:"136230"`
				Num136231 []int `json:"136231"`
				Num136232 []int `json:"136232"`
				Num136233 []int `json:"136233"`
				Num136234 []int `json:"136234"`
				Num136235 []int `json:"136235"`
				Num136236 []int `json:"136236"`
				Num136237 []int `json:"136237"`
				Num136238 []int `json:"136238"`
				Num136239 []int `json:"136239"`
				Num136240 []int `json:"136240"`
				Num136241 []int `json:"136241"`
				Num136242 []int `json:"136242"`
				Num136243 []int `json:"136243"`
				Num136244 []int `json:"136244"`
				Num136245 []int `json:"136245"`
				Num136246 []int `json:"136246"`
				Num136247 []int `json:"136247"`
				Num136248 []int `json:"136248"`
				Num136249 []int `json:"136249"`
				Num136250 []int `json:"136250"`
				Num136251 []int `json:"136251"`
				Num136252 []int `json:"136252"`
				Num136253 []int `json:"136253"`
				Num136254 []int `json:"136254"`
				Num136255 []int `json:"136255"`
				Num136257 []int `json:"136257"`
				Num136258 []int `json:"136258"`
				Num136259 []int `json:"136259"`
				Num136260 []int `json:"136260"`
				Num136261 []int `json:"136261"`
				Num136262 []int `json:"136262"`
				Num136263 []int `json:"136263"`
				Num136264 []int `json:"136264"`
				Num136265 []int `json:"136265"`
				Num136266 []int `json:"136266"`
				Num136267 []int `json:"136267"`
				Num136268 []int `json:"136268"`
				Num136269 []int `json:"136269"`
				Num136270 []int `json:"136270"`
				Num136271 []int `json:"136271"`
				Num136272 []int `json:"136272"`
				Num136273 []int `json:"136273"`
				Num136274 []int `json:"136274"`
				Num136275 []int `json:"136275"`
				Num136276 []int `json:"136276"`
				Num136277 []int `json:"136277"`
				Num136278 []int `json:"136278"`
				Num136279 []int `json:"136279"`
				Num136280 []int `json:"136280"`
				Num136281 []int `json:"136281"`
				Num136282 []int `json:"136282"`
				Num136283 []int `json:"136283"`
				Num136284 []int `json:"136284"`
				Num136285 []int `json:"136285"`
				Num136286 []int `json:"136286"`
				Num136287 []int `json:"136287"`
				Num136288 []int `json:"136288"`
				Num136289 []int `json:"136289"`
				Num136290 []int `json:"136290"`
				Num136291 []int `json:"136291"`
				Num136292 []int `json:"136292"`
				Num136294 []int `json:"136294"`
				Num136295 []int `json:"136295"`
				Num136298 []int `json:"136298"`
				Num136299 []int `json:"136299"`
				Num136300 []int `json:"136300"`
				Num136301 []int `json:"136301"`
				Num136302 []int `json:"136302"`
				Num136303 []int `json:"136303"`
				Num136304 []int `json:"136304"`
				Num136305 []int `json:"136305"`
				Num136306 []int `json:"136306"`
				Num136307 []int `json:"136307"`
				Num136308 []int `json:"136308"`
				Num136309 []int `json:"136309"`
				Num136310 []int `json:"136310"`
				Num136466 []int `json:"136466"`
				Num136467 []int `json:"136467"`
				Num136481 []int `json:"136481"`
				Num136482 []int `json:"136482"`
				Num136484 []int `json:"136484"`
				Num136485 []int `json:"136485"`
				Num136488 []int `json:"136488"`
				Num136490 []int `json:"136490"`
				Num136491 []int `json:"136491"`
				Num136495 []int `json:"136495"`
				Num136497 []int `json:"136497"`
				Num137118 []int `json:"137118"`
				Num137119 []int `json:"137119"`
				Num137120 []int `json:"137120"`
				Num137121 []int `json:"137121"`
				Num137122 []int `json:"137122"`
				Num137123 []int `json:"137123"`
				Num137124 []int `json:"137124"`
				Num137125 []int `json:"137125"`
				Num137126 []int `json:"137126"`
				Num137127 []int `json:"137127"`
				Num137128 []int `json:"137128"`
				Num137129 []int `json:"137129"`
				Num137130 []int `json:"137130"`
				Num137131 []int `json:"137131"`
				Num137132 []int `json:"137132"`
				Num137133 []int `json:"137133"`
				Num137135 []int `json:"137135"`
				Num137136 []int `json:"137136"`
				Num137166 []int `json:"137166"`
				Num137191 []int `json:"137191"`
				Num137192 []int `json:"137192"`
				Num137193 []int `json:"137193"`
				Num137400 []int `json:"137400"`
				Num137402 []int `json:"137402"`
				Num137410 []int `json:"137410"`
				Num137411 []int `json:"137411"`
				Num137427 []int `json:"137427"`
				Num137428 []int `json:"137428"`
				Num137429 []int `json:"137429"`
				Num138229 []int `json:"138229"`
			} `json:"combinedCategoriesMapping"`
		} `json:"sizeProfile"`
		IsMobile     bool `json:"isMobile"`
		ListingItems struct {
			Items []struct {
				ID               int    `json:"id"`
				ShortDescription string `json:"shortDescription"`
				MerchantID       int    `json:"merchantId"`
				Brand            struct {
					ID   int    `json:"id"`
					Name string `json:"name"`
				} `json:"brand"`
				Gender string `json:"gender"`
				Images struct {
					CutOut string      `json:"cutOut"`
					Model  string      `json:"model"`
					All    interface{} `json:"all"`
				} `json:"images"`
				PriceInfo struct {
					FormattedFinalPrice   string      `json:"formattedFinalPrice"`
					FormattedInitialPrice string      `json:"formattedInitialPrice"`
					FinalPrice            int         `json:"finalPrice"`
					InitialPrice          int         `json:"initialPrice"`
					CurrencyCode          string      `json:"currencyCode"`
					IsOnSale              bool        `json:"isOnSale"`
					DiscountLabel         interface{} `json:"discountLabel"`
					InstallmentsLabel     interface{} `json:"installmentsLabel"`
				} `json:"priceInfo"`
				MerchandiseLabel      interface{} `json:"merchandiseLabel"`
				MerchandiseLabelField string      `json:"merchandiseLabelField"`
				IsCustomizable        bool        `json:"isCustomizable"`
				AvailableSizes        interface{} `json:"availableSizes"`
				StockTotal            int         `json:"stockTotal"`
				HasSimilarProducts    bool        `json:"hasSimilarProducts"`
				URL                   string      `json:"url"`
				HeroProductType       interface{} `json:"heroProductType"`
				Type                  string      `json:"type"`
				Properties            struct {
				} `json:"properties"`
			} `json:"items"`
		} `json:"listingItems"`
		ListingPagination struct {
			Index                int    `json:"index"`
			View                 int    `json:"view"`
			TotalItems           int    `json:"totalItems"`
			TotalPages           int    `json:"totalPages"`
			NormalizedTotalItems string `json:"normalizedTotalItems"`
		} `json:"listingPagination"`
		ListingFilters struct {
			Facets struct {
				Category struct {
					Values []struct {
						URLToken    string `json:"urlToken"`
						URL         string `json:"url"`
						Value       string `json:"value"`
						Description string `json:"description"`
						Count       int    `json:"count"`
						Deep        int    `json:"deep"`
					} `json:"values"`
					Description           string `json:"description"`
					Filter                string `json:"filter"`
					IsServerSideRendering bool   `json:"isServerSideRendering"`
				} `json:"category"`
				Designer struct {
					Values []struct {
						URLToken    string `json:"urlToken"`
						URL         string `json:"url"`
						Value       string `json:"value"`
						Description string `json:"description"`
						Count       int    `json:"count"`
						Deep        int    `json:"deep"`
					} `json:"values"`
					Description           string `json:"description"`
					Filter                string `json:"filter"`
					IsServerSideRendering bool   `json:"isServerSideRendering"`
				} `json:"designer"`
				Size struct {
					Values []struct {
						Category     string      `json:"category"`
						Dependencies interface{} `json:"dependencies"`
						Value        string      `json:"value"`
						Description  string      `json:"description"`
						Count        int         `json:"count"`
						Deep         int         `json:"deep"`
					} `json:"values"`
					Description           string      `json:"description"`
					Filter                interface{} `json:"filter"`
					IsServerSideRendering bool        `json:"isServerSideRendering"`
				} `json:"size"`
				Colour struct {
					Values []struct {
						Value       string `json:"value"`
						Description string `json:"description"`
						Count       int    `json:"count"`
						Deep        int    `json:"deep"`
					} `json:"values"`
					Description           string      `json:"description"`
					Filter                interface{} `json:"filter"`
					IsServerSideRendering bool        `json:"isServerSideRendering"`
				} `json:"colour"`
				Discount struct {
					Values                []interface{} `json:"values"`
					Description           string        `json:"description"`
					Filter                interface{}   `json:"filter"`
					IsServerSideRendering bool          `json:"isServerSideRendering"`
				} `json:"discount"`
				Price struct {
					Max                   int         `json:"max"`
					MaxBoundary           int         `json:"maxBoundary"`
					MinBoundary           int         `json:"minBoundary"`
					Description           string      `json:"description"`
					Filter                interface{} `json:"filter"`
					IsServerSideRendering bool        `json:"isServerSideRendering"`
				} `json:"price"`
				Labels struct {
					Values []struct {
						Value       string `json:"value"`
						Description string `json:"description"`
						Count       int    `json:"count"`
						Deep        int    `json:"deep"`
					} `json:"values"`
					Description           string `json:"description"`
					Filter                string `json:"filter"`
					IsServerSideRendering bool   `json:"isServerSideRendering"`
				} `json:"labels"`
			} `json:"facets"`
			Filters struct {
			} `json:"filters"`
			Scale struct {
				Values        []interface{} `json:"values"`
				DefaultScale  int           `json:"defaultScale"`
				SelectedScale int           `json:"selectedScale"`
			} `json:"scale"`
		} `json:"listingFilters"`
	}

	// err = ioutil.WriteFile("C:\\Rinal\\ServiceBasedPRojects\\VoilaWork\\VoilaCrawl\\output_0.txt", matched[0], 0644)
	// err = ioutil.WriteFile("C:\\Rinal\\ServiceBasedPRojects\\VoilaWork\\VoilaCrawl\\output_1.txt", matched[1], 0644)
	// err = ioutil.WriteFile("C:\\Rinal\\ServiceBasedPRojects\\VoilaWork\\VoilaCrawl\\output_2.txt", matched[2], 0644)
	

	matched[2] = bytes.ReplaceAll(bytes.ReplaceAll(matched[2], []byte("\\'"), []byte("'")), []byte(`\\"`), []byte(`\"`))
	// rawData, err := strconv.Unquote(string(matched[1]))
	//if err != nil {
	//	c.logger.Errorf("unquote raw string failed, error=%s", err)
	//	return err
	//}
	if err = json.Unmarshal(matched[2], &r); err != nil {
		c.logger.Debugf("parse %s failed, error=%s", matched[1], err)
		return err
	}

	cid := r.ListingPagination.Index
	nctx := context.WithValue(ctx, "page", cid)
	lastIndex := nextIndex(ctx)
	for _, prod := range r.ListingItems.Items {
		rawurl := fmt.Sprintf("%s://%s/us/%s", resp.Request.URL.Scheme, resp.Request.URL.Host, prod.URL)
		if strings.HasPrefix(prod.URL, "http:") || strings.HasPrefix(prod.URL, "https:") {
			rawurl = prod.URL
		}

		if req, err := http.NewRequest(http.MethodGet, rawurl, nil); err != nil {
			c.logger.Debug(err)
			return err
		} else {
			nnctx := context.WithValue(nctx, "item.index", lastIndex+1)
			lastIndex += 1
			if err = yield(nnctx, req); err != nil {
				return err
			}
		}
	}

	u := *resp.Request.URL
	// u.Path = fmt.Sprintf("/api/product/search/v2/categories/%v", cid)
	 vals := url.Values{}
	// for key, val := range r.SliceListing.ListingPagination {
	// 	if key == "cid" || key == "page" {
	// 		continue
	// 	}
	// 	vals.Set(key, fmt.Sprintf("%v", val))
	// }
	// vals.Set("offset", strconv.Format(len(r.Search.Products)))
	 vals.Set("page", strconv.Format((r.ListingPagination.Index + 1)))
	 u.RawQuery = vals.Encode()

	 fmt.Println(u.String())

	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	return yield(context.WithValue(nctx, "item.index", lastIndex), req)
}

type parseProductResponse struct {
	Links     struct{} `json:"_links"`
	AbTesting struct {
		PdpReviewcontactDesktop string `json:"pdp_reviewcontact_desktop"`
		PdpReviewcontactMobile  string `json:"pdp_reviewcontact_mobile"`
	} `json:"abTesting"`
	ApplePayInfo struct {
		IsVisible bool `json:"isVisible"`
	} `json:"applePayInfo"`
	ComplementaryCategories struct {
		Abtests     interface{}   `json:"abtests"`
		CategoryIds []interface{} `json:"categoryIds"`
	} `json:"complementaryCategories"`
	Config struct {
		ContactFormURI              string      `json:"contactFormUri"`
		FitPredictorEnv             string      `json:"fitPredictorEnv"`
		NoRedirectOnAddToBagEnabled interface{} `json:"noRedirectOnAddToBagEnabled"`
		ShoppingBagURL              string      `json:"shoppingBagUrl"`
		SizeGuideSliceID            string      `json:"sizeGuideSliceId"`
		StaticContentBaseURI        string      `json:"staticContentBaseUri"`
	} `json:"config"`
	ContactUs struct {
		CustomerServiceEmail string `json:"customerServiceEmail"`
	} `json:"contactUs"`
	CrossSelling struct {
		ActiveTypes             []int64 `json:"activeTypes"`
		ComplementaryCategories struct {
			Abtests     interface{}   `json:"abtests"`
			CategoryIds []interface{} `json:"categoryIds"`
		} `json:"complementaryCategories"`
		ShopTheLook struct {
			Abtests struct {
				HasSkeleton                   bool `json:"hasSkeleton"`
				IsCompleteYourLookEnabled     bool `json:"isCompleteYourLookEnabled"`
				IsInComplementaryProductsMode bool `json:"isInComplementaryProductsMode"`
				IsInMainLookMode              bool `json:"isInMainLookMode"`
				IsInMainLookModeMobile        bool `json:"isInMainLookModeMobile"`
				IsInModalModeMobile           bool `json:"isInModalModeMobile"`
				IsStlRenamed                  bool `json:"isStlRenamed"`
			} `json:"abtests"`
			OutfitID   int64   `json:"outfitId"`
			ProductIds []int64 `json:"productIds"`
			Products   []int64 `json:"products"`
			Settings   struct {
				HasProductsOnline        bool  `json:"hasProductsOnline"`
				IsBlacklistedBrand       bool  `json:"isBlacklistedBrand"`
				IsMainProductShoesOrBags bool  `json:"isMainProductShoesOrBags"`
				IsSkeletonEnabled        bool  `json:"isSkeletonEnabled"`
				ModelImageStyle          int64 `json:"modelImageStyle"`
			} `json:"settings"`
		} `json:"shopTheLook"`
	} `json:"crossSelling"`
	Culture struct {
		ContextGenderID        int64  `json:"contextGenderId"`
		CountryCode            string `json:"countryCode"`
		CountryCultureCode     string `json:"countryCultureCode"`
		CountryID              string `json:"countryId"`
		CurrentSubfolder       string `json:"currentSubfolder"`
		Domain                 string `json:"domain"`
		IsGDPRCompliantCountry bool   `json:"isGDPRCompliantCountry"`
		Language               string `json:"language"`
		LanguageCultureCode    string `json:"languageCultureCode"`
		RenderDirection        string `json:"renderDirection"`
	} `json:"culture"`
	CustomizationSettings struct {
		CustomizationMessage string `json:"customizationMessage"`
		IframeURL            string `json:"iframeUrl"`
	} `json:"customizationSettings"`
	Flats       interface{} `json:"flats"`
	IsSSRMobile bool        `json:"isSSRMobile"`
	Labels      struct {
		AddOriginalStyle                    string `json:"addOriginalStyle"`
		AddToBag                            string `json:"addToBag"`
		AddToBagError                       string `json:"addToBagError"`
		AddToBagLoading                     string `json:"addToBagLoading"`
		AddToBagShortDescription            string `json:"addToBagShortDescription"`
		AddToWishlist                       string `json:"addToWishlist"`
		AddedToYourBag                      string `json:"addedToYourBag"`
		All                                 string `json:"all"`
		AllMeasurementsFarfetch             string `json:"allMeasurementsFarfetch"`
		AlsoAvailableIn                     string `json:"alsoAvailableIn"`
		AproximateTimeframe                 string `json:"aproximateTimeframe"`
		BrandColour                         string `json:"brandColour"`
		BrandStyle                          string `json:"brandStyle"`
		BuyNow                              string `json:"buyNow"`
		ByEmail                             string `json:"byEmail"`
		CentimetersLiteral                  string `json:"centimetersLiteral"`
		ChatOnline                          string `json:"chatOnline"`
		CheckPostcodeAndDetails             string `json:"checkPostcodeAndDetails"`
		CheckSizeGuide                      string `json:"checkSizeGuide"`
		ClimateConsciousDelivery            string `json:"climateConsciousDelivery"`
		CloseDialog                         string `json:"closeDialog"`
		ComplementaryCategoriesTitle        string `json:"complementaryCategoriesTitle"`
		CompleteTheLook                     string `json:"completeTheLook"`
		CompleteYourLook                    string `json:"completeYourLook"`
		Composition                         string `json:"composition"`
		CompositionAndCare                  string `json:"compositionAndCare"`
		Contact                             string `json:"contact"`
		ContactFormSuccessBody              string `json:"contactFormSuccessBody"`
		ContactFormSuccessTitle             string `json:"contactFormSuccessTitle"`
		ContactUs                           string `json:"contactUs"`
		ContinueShopping                    string `json:"continueShopping"`
		CountryRegion                       string `json:"countryRegion"`
		CustomisableFlatTitle               string `json:"customisableFlatTitle"`
		CustomisedPiece                     string `json:"customisedPiece"`
		CustomizeThisStyle                  string `json:"customizeThisStyle"`
		CustomsIncluded                     string `json:"customsIncluded"`
		DedicatedStylist                    string `json:"dedicatedStylist"`
		DeliveryAvailable                   string `json:"deliveryAvailable"`
		DeliveryInfo                        string `json:"deliveryInfo"`
		DesignerAndMore                     string `json:"designerAndMore"`
		DesignerBackstory                   string `json:"designerBackstory"`
		DesignerButtonPrefix                string `json:"designerButtonPrefix"`
		DesignerColour                      string `json:"designerColour"`
		DesignerStyleID                     string `json:"designerStyleID"`
		DetailsStlDescription               string `json:"detailsStlDescription"`
		DetailsStlProductsTitle             string `json:"detailsStlProductsTitle"`
		DifferentPrice                      string `json:"differentPrice"`
		DifferentTo                         string `json:"differentTo"`
		DifferentToPrice                    string `json:"differentToPrice"`
		Disclaimer                          string `json:"disclaimer"`
		DiscoverMore                        string `json:"discoverMore"`
		DisplayingMeasurementsForSizeScale  string `json:"displayingMeasurementsForSizeScale"`
		DontForgetFreeReturns               string `json:"dontForgetFreeReturns"`
		DutiesAndTaxes                      string `json:"dutiesAndTaxes"`
		Email                               string `json:"email"`
		EmailMeWhenItsBack                  string `json:"emailMeWhenItsBack"`
		EmailUs                             string `json:"emailUs"`
		ErrorLoadingSizeGuide               string `json:"errorLoadingSizeGuide"`
		ErrorTermsAndConditions             string `json:"errorTermsAndConditions"`
		EstimatedDelivery                   string `json:"estimatedDelivery"`
		ExclusiveDiscount                   string `json:"exclusiveDiscount"`
		Express                             string `json:"express"`
		F90                                 string `json:"f90"`
		F90service                          string `json:"f90service"`
		FarfetchID                          string `json:"farfetchID"`
		FarfetchItemID                      string `json:"farfetchItemID"`
		FfAccessBronze                      string `json:"ffAccessBronze"`
		FfAccessGold                        string `json:"ffAccessGold"`
		FfAccessPlatinum                    string `json:"ffAccessPlatinum"`
		FfAccessPrivate                     string `json:"ffAccessPrivate"`
		FfAccessProgramName                 string `json:"ffAccessProgramName"`
		FfAccessSilver                      string `json:"ffAccessSilver"`
		FfAccessUpgradeMessage              string `json:"ffAccessUpgradeMessage"`
		FindOutWhy                          string `json:"findOutWhy"`
		FittingTitle                        string `json:"fittingTitle"`
		ForOrders                           string `json:"forOrders"`
		ForSomeSizes                        string `json:"forSomeSizes"`
		FreeGlobalReturns                   string `json:"freeGlobalReturns"`
		FreeReturnAndPickupService          string `json:"freeReturnAndPickupService"`
		FreeShipping                        string `json:"freeShipping"`
		FtaAndImportDuties                  string `json:"ftaAndImportDuties"`
		GenericErrorMessage                 string `json:"genericErrorMessage"`
		Get90                               string `json:"get90"`
		GetToday                            string `json:"getToday"`
		GoToBag                             string `json:"goToBag"`
		GotIt                               string `json:"gotIt"`
		GreatChoice                         string `json:"greatChoice"`
		HassleDelivery                      string `json:"hassleDelivery"`
		Help                                string `json:"help"`
		HelpAndAdvice                       string `json:"helpAndAdvice"`
		HelpAndContactUs                    string `json:"helpAndContactUs"`
		HeresHow                            string `json:"heresHow"`
		Highlights                          string `json:"highlights"`
		HowItWorksQuestion                  string `json:"howItWorksQuestion"`
		ImportDutiesInformation             string `json:"importDutiesInformation"`
		InAHurry                            string `json:"inAHurry"`
		InBag                               string `json:"inBag"`
		InchesLiteral                       string `json:"inchesLiteral"`
		ItemAdded                           string `json:"itemAdded"`
		LabelComma                          string `json:"labelComma"`
		LabelDot                            string `json:"labelDot"`
		LastOneLeft                         string `json:"lastOneLeft"`
		LikeThisPiece                       string `json:"likeThisPiece"`
		MakeItYours                         string `json:"makeItYours"`
		Measurements                        string `json:"measurements"`
		ModelIsWearing                      string `json:"modelIsWearing"`
		ModelIsWearingV2                    string `json:"modelIsWearingV2"`
		ModelMeasurements                   string `json:"modelMeasurements"`
		MoreFromDesigner                    string `json:"moreFromDesigner"`
		MoreInformation                     string `json:"moreInformation"`
		MySwearError                        string `json:"mySwearError"`
		MySwearLoadingError                 string `json:"mySwearLoadingError"`
		NPieces                             string `json:"nPieces"`
		NeedMoreInformation                 string `json:"needMoreInformation"`
		NeedThisIn90Minutes                 string `json:"needThisIn90Minutes"`
		NeedThisToday                       string `json:"needThisToday"`
		NeedToConvertSizes                  string `json:"needToConvertSizes"`
		NewPrice                            string `json:"newPrice"`
		NinetyMinutesToDoor                 string `json:"ninetyMinutesToDoor"`
		NotesAndCare                        string `json:"notesAndCare"`
		NotesAndSizing                      string `json:"notesAndSizing"`
		NotifyMeBack                        string `json:"notifyMeBack"`
		NotifyMeLowStock                    string `json:"notifyMeLowStock"`
		OneSizeAvailable                    string `json:"oneSizeAvailable"`
		OnlyOneLeft                         string `json:"onlyOneLeft"`
		Or                                  string `json:"or"`
		OrderBy                             string `json:"orderBy"`
		OrderByPhone                        string `json:"orderByPhone"`
		OrderReady                          string `json:"orderReady"`
		OrderWithUs                         string `json:"orderWithUs"`
		OrdersAndShipping                   string `json:"ordersAndShipping"`
		OurModelIs                          string `json:"ourModelIs"`
		OutOfStockMultipleVariantTitle      string `json:"outOfStockMultipleVariantTitle"`
		OutOfStockProductDetailsTitle       string `json:"outOfStockProductDetailsTitle"`
		OutOfStockRecentlyViewedModuleAlt   string `json:"outOfStockRecentlyViewedModuleAlt"`
		OutOfStockRecentlyViewedTitle       string `json:"outOfStockRecentlyViewedTitle"`
		OutOfStockSameBrandModuleAlt        string `json:"outOfStockSameBrandModuleAlt"`
		OutOfStockSameBrandSubtitle         string `json:"outOfStockSameBrandSubtitle"`
		OutOfStockSameBrandTitle            string `json:"outOfStockSameBrandTitle"`
		OutOfStockSeeAll                    string `json:"outOfStockSeeAll"`
		OutOfStockSimilarModuleAlt          string `json:"outOfStockSimilarModuleAlt"`
		OutOfStockSimilarProductsTitle      string `json:"outOfStockSimilarProductsTitle"`
		OutOfStockSingleVariantTitle        string `json:"outOfStockSingleVariantTitle"`
		OutOfStockSubCategoryLinksTitle     string `json:"outOfStockSubCategoryLinksTitle"`
		OutOfStockTitle                     string `json:"outOfStockTitle"`
		OutOfStockUsefulLinksTitle          string `json:"outOfStockUsefulLinksTitle"`
		OutOfStockVariantsModuleAlt         string `json:"outOfStockVariantsModuleAlt"`
		PersonalStylist                     string `json:"personalStylist"`
		Phone                               string `json:"phone"`
		PhotoOfThisStyle                    string `json:"photoOfThisStyle"`
		PleaseEnterValidEmail               string `json:"pleaseEnterValidEmail"`
		PleaseRefreshOrTryAgainLater        string `json:"pleaseRefreshOrTryAgainLater"`
		PleaseSelectASize                   string `json:"pleaseSelectASize"`
		PleaseSignUp                        string `json:"pleaseSignUp"`
		Price                               string `json:"price"`
		PriceBeforeDiscount                 string `json:"priceBeforeDiscount"`
		PriceChangeMessage                  string `json:"priceChangeMessage"`
		PriorityPhoneLineDisclaimerPlatinum string `json:"priorityPhoneLineDisclaimerPlatinum"`
		PriorityPhoneLineDisclaimerPrivate  string `json:"priorityPhoneLineDisclaimerPrivate"`
		ProductMeasurementInfo              string `json:"productMeasurementInfo"`
		ProductMeasurementInfoOneSize       string `json:"productMeasurementInfoOneSize"`
		ProductMeasurements                 string `json:"productMeasurements"`
		ProductMeasurementsForSize          string `json:"productMeasurementsForSize"`
		ReadMore                            string `json:"readMore"`
		RemoveFromWishlist                  string `json:"removeFromWishlist"`
		ReturnsAndRefunds                   string `json:"returnsAndRefunds"`
		SameDay                             string `json:"sameDay"`
		SameDayDeliveryAvailable            string `json:"sameDayDeliveryAvailable"`
		SeeAllImages                        string `json:"seeAllImages"`
		SeeFullDetails                      string `json:"seeFullDetails"`
		SeeLess                             string `json:"seeLess"`
		SeeMore                             string `json:"seeMore"`
		SeeMoreMeasurements                 string `json:"seeMoreMeasurements"`
		SeeMoreOf                           string `json:"seeMoreOf"`
		SeeSimilar                          string `json:"seeSimilar"`
		SeeSomethingSimilar                 string `json:"seeSomethingSimilar"`
		SeeingDifferentPrice                string `json:"seeingDifferentPrice"`
		Select                              string `json:"select"`
		Selected                            string `json:"selected"`
		Send                                string `json:"send"`
		Service                             string `json:"service"`
		ServiceAvailable                    string `json:"serviceAvailable"`
		ShareThis                           string `json:"shareThis"`
		Shipping                            string `json:"shipping"`
		ShippingElsewhere                   string `json:"shippingElsewhere"`
		ShippingElsewhereMessage            string `json:"shippingElsewhereMessage"`
		ShippingFreeReturns                 string `json:"shippingFreeReturns"`
		ShippingReturns                     string `json:"shippingReturns"`
		ShippingToAnotherCountry            string `json:"shippingToAnotherCountry"`
		ShopTheLook                         string `json:"shopTheLook"`
		Similar                             string `json:"similar"`
		SimilarNotAvailable                 string `json:"similarNotAvailable"`
		Size                                string `json:"size"`
		SizeFit                             string `json:"sizeFit"`
		SizeGuide                           string `json:"sizeGuide"`
		SizeGuideMeasurements               string `json:"sizeGuideMeasurements"`
		SizeMissing                         string `json:"sizeMissing"`
		SizeUnavailabeGetNotified           string `json:"sizeUnavailabeGetNotified"`
		SizesWithDutiesIncluded             string `json:"sizesWithDutiesIncluded"`
		SlideshowImagesAlt                  string `json:"slideshowImagesAlt"`
		SoldOut                             string `json:"soldOut"`
		SomethingWentWrong                  string `json:"somethingWentWrong"`
		SpecialDeliveryAvailable            string `json:"specialDeliveryAvailable"`
		SpeedierService                     string `json:"speedierService"`
		Standard                            string `json:"standard"`
		StillNeedHelp                       string `json:"stillNeedHelp"`
		StoreToDoor                         string `json:"storeToDoor"`
		StyleItWith                         string `json:"styleItWith"`
		TabDesignerID                       string `json:"tabDesignerID"`
		TermsAndConditionChinaDomain        string `json:"termsAndConditionChinaDomain"`
		TermsAndConditions                  string `json:"termsAndConditions"`
		TheDetails                          string `json:"theDetails"`
		ThisPieceCanBeYoursIn               string `json:"thisPieceCanBeYoursIn"`
		ThisPieceHas                        string `json:"thisPieceHas"`
		TryOurSizeGuide                     string `json:"tryOurSizeGuide"`
		ViewAll                             string `json:"viewAll"`
		ViewMore                            string `json:"viewMore"`
		ViewProduct                         string `json:"viewProduct"`
		ViewSizeGuide                       string `json:"viewSizeGuide"`
		ViewTheLook                         string `json:"viewTheLook"`
		WantThisIn90minutes                 string `json:"wantThisIn90minutes"`
		WantThisToday                       string `json:"wantThisToday"`
		WeDoFreeReturns                     string `json:"weDoFreeReturns"`
		WeGotYourBack                       string `json:"weGotYourBack"`
		WeSpeak                             string `json:"weSpeak"`
		WearItWith                          string `json:"wearItWith"`
		Wearf90                             string `json:"wearf90"`
		WearingDescription                  string `json:"wearingDescription"`
		WhyNotTryWith                       string `json:"whyNotTryWith"`
		Wishlist                            string `json:"wishlist"`
	} `json:"labels"`
	OutOfStockPage struct {
		OutOfStockLinks     interface{} `json:"outOfStockLinks"`
		ShowNewOutOfStock   bool        `json:"showNewOutOfStock"`
		ShowOutOfStockSlice bool        `json:"showOutOfStockSlice"`
	} `json:"outOfStockPage"`
	ProductViewModel struct {
		Breadcrumb []struct {
			Data_ffref string `json:"data-ffref"`
			Data_type  string `json:"data-type"`
			Href       string `json:"href"`
			Text       string `json:"text"`
		} `json:"breadcrumb"`
		Care struct {
			Disclaimer   interface{} `json:"disclaimer"`
			Instructions struct {
				Washing_instructions []string `json:"washing instructions"`
			} `json:"instructions"`
		} `json:"care"`
		Categories struct {
			All struct {
				One35967 string `json:"135967"`
				One35983 string `json:"135983"`
				One36099 string `json:"136099"`
			} `json:"all"`
			Category struct {
				ID   int64  `json:"id"`
				Name string `json:"name"`
			} `json:"category"`
			SubCategory struct {
				ID   int64  `json:"id"`
				Name string `json:"name"`
			} `json:"subCategory"`
		} `json:"categories"`
		CategoriesTree []struct {
			ID            int64  `json:"id"`
			Name          string `json:"name"`
			SubCategories []struct {
				ID            int64  `json:"id"`
				Name          string `json:"name"`
				SubCategories []struct {
					ID            int64         `json:"id"`
					Name          string        `json:"name"`
					SubCategories []interface{} `json:"subCategories"`
				} `json:"subCategories"`
			} `json:"subCategories"`
		} `json:"categoriesTree"`
		Composition struct {
			Materials struct {
				_ []string `json:""`
			} `json:"materials"`
		} `json:"composition"`
		DesignerDetails struct {
			Description     string `json:"description"`
			DesignerColour  string `json:"designerColour"`
			DesignerStyleID string `json:"designerStyleId"`
			ID              int64  `json:"id"`
			Link            struct {
				Data_ffref string `json:"data-ffref"`
				Href       string `json:"href"`
				Text       string `json:"text"`
			} `json:"link"`
			Name string `json:"name"`
		} `json:"designerDetails"`
		Details struct {
			AgeGroup    string `json:"ageGroup"`
			Colors      string `json:"colors"`
			Department  string `json:"department"`
			Description string `json:"description"`
			Gender      int64  `json:"gender"`
			GenderName  string `json:"genderName"`
			Link        struct {
				Data_ffref interface{} `json:"data-ffref"`
				Href       string      `json:"href"`
				Text       interface{} `json:"text"`
			} `json:"link"`
			MadeIn struct {
				Label string `json:"label"`
			} `json:"madeIn"`
			MadeInLabel           string      `json:"madeInLabel"`
			MerchandiseTag        string      `json:"merchandiseTag"`
			MerchandiseTagField   string      `json:"merchandiseTagField"`
			MerchandiseTagID      interface{} `json:"merchandiseTagId"`
			MerchandisingLabelIds []int64     `json:"merchandisingLabelIds"`
			MerchantID            int64       `json:"merchantId"`
			ProductID             int64       `json:"productId"`
			RichText              struct {
				Description []struct {
					Blocks []struct {
						DisplayOptions struct{} `json:"displayOptions"`
						Items          []struct {
							DisplayOptions struct{}    `json:"displayOptions"`
							Name           interface{} `json:"name"`
							Type           string      `json:"type"`
							Value          string      `json:"value"`
						} `json:"items"`
						Name interface{} `json:"name"`
						Type string      `json:"type"`
					} `json:"blocks"`
					DisplayOptions struct{}    `json:"displayOptions"`
					Name           interface{} `json:"name"`
					Type           string      `json:"type"`
				} `json:"description"`
				Highlights []struct {
					Blocks []struct {
						DisplayOptions struct{} `json:"displayOptions"`
						Items          []struct {
							DisplayOptions struct {
								Display string `json:"display"`
							} `json:"displayOptions"`
							Name  interface{} `json:"name"`
							Type  string      `json:"type"`
							Value string      `json:"value"`
						} `json:"items"`
						Name string `json:"name"`
						Type string `json:"type"`
					} `json:"blocks"`
					DisplayOptions struct{}    `json:"displayOptions"`
					Name           interface{} `json:"name"`
					Type           string      `json:"type"`
				} `json:"highlights"`
			} `json:"richText"`
			ShortDescription string `json:"shortDescription"`
			StyleID          int64  `json:"styleId"`
		} `json:"details"`
		FitPredictor struct {
			Alternative    string      `json:"alternative"`
			IsContextValid bool        `json:"isContextValid"`
			SizeSystemID   interface{} `json:"sizeSystemId"`
		} `json:"fitPredictor"`
		Images struct {
			Details struct {
				Six00     string `json:"600"`
				Alt       string `json:"alt"`
				Index     int64  `json:"index"`
				Large     string `json:"large"`
				Medium    string `json:"medium"`
				Size200   string `json:"size200"`
				Size240   string `json:"size240"`
				Size300   string `json:"size300"`
				Small     string `json:"small"`
				Thumbnail string `json:"thumbnail"`
				Zoom      string `json:"zoom"`
			} `json:"details"`
			HasRunway               bool `json:"hasRunway"`
			IsEligibleForLimoncello bool `json:"isEligibleForLimoncello"`
			Main                    []struct {
				Six00     string `json:"600"`
				Alt       string `json:"alt"`
				Index     int64  `json:"index"`
				Large     string `json:"large"`
				Medium    string `json:"medium"`
				Size200   string `json:"size200"`
				Size240   string `json:"size240"`
				Size300   string `json:"size300"`
				Small     string `json:"small"`
				Thumbnail string `json:"thumbnail"`
				Zoom      string `json:"zoom"`
			} `json:"main"`
		} `json:"images"`
		Measurements struct {
			Available           []string      `json:"available"`
			Category            interface{}   `json:"category"`
			DefaultMeasurement  int64         `json:"defaultMeasurement"`
			DefaultSize         int64         `json:"defaultSize"`
			ExtraMeasurements   []interface{} `json:"extraMeasurements"`
			FittingInformation  []string      `json:"fittingInformation"`
			FriendlyScaleName   string        `json:"friendlyScaleName"`
			IsOneSize           bool          `json:"isOneSize"`
			IsSingleMeasurement bool          `json:"isSingleMeasurement"`
			ModelHeight         []string      `json:"modelHeight"`
			ModelIsWearing      string        `json:"modelIsWearing"`
			ModelMeasurements   struct {
				Bust_Chest []string `json:"bust/Chest"`
				Height     []string `json:"height"`
				Hips       []string `json:"hips"`
				Waist      []string `json:"waist"`
			} `json:"modelMeasurements"`
			SizeDescription struct{} `json:"sizeDescription"`
			Sizes           struct{} `json:"sizes"`
		} `json:"measurements"`
		PriceInfo struct {
			Default struct {
				CurrencyCode                  string `json:"currencyCode"`
				FinalPrice                    int64  `json:"finalPrice"`
				FormattedFinalPrice           string `json:"formattedFinalPrice"`
				FormattedFinalPriceInternal   string `json:"formattedFinalPriceInternal"`
				FormattedInitialPrice         string `json:"formattedInitialPrice"`
				FormattedInitialPriceInternal string `json:"formattedInitialPriceInternal"`
				InitialPrice                  int64  `json:"initialPrice"`
				IsOnSale                      bool   `json:"isOnSale"`
				Labels                        struct {
					Duties string `json:"duties"`
				} `json:"labels"`
				PriceTags []string `json:"priceTags"`
			} `json:"default"`
		} `json:"priceInfo"`
		ProductHeroPage struct {
			RulesApplyToProduct bool `json:"rulesApplyToProduct"`
		} `json:"productHeroPage"`
		ProductOfferVariant string      `json:"productOfferVariant"`
		SeeMoreLinks        interface{} `json:"seeMoreLinks"`
		Share               struct {
			SocialIcons []struct {
				Type string `json:"type"`
				URL  string `json:"url"`
			} `json:"socialIcons"`
		} `json:"share"`
		ShippingInformations struct {
			Details struct {
				Nine445 struct {
					CityID                      int64       `json:"cityId"`
					CountryCode                 string      `json:"countryCode"`
					DeliveryBefore              interface{} `json:"deliveryBefore"`
					DeliveryBy                  interface{} `json:"deliveryBy"`
					DeliveryCityMessage         interface{} `json:"deliveryCityMessage"`
					DeliveryGreetingsMessage    interface{} `json:"deliveryGreetingsMessage"`
					DeliveryIn                  interface{} `json:"deliveryIn"`
					DeliveryType                interface{} `json:"deliveryType"`
					EndTime                     interface{} `json:"endTime"`
					FarfetchOwned               interface{} `json:"farfetchOwned"`
					IsFromEurasianCustomsUnion  bool        `json:"isFromEurasianCustomsUnion"`
					IsLocalStock                bool        `json:"isLocalStock"`
					LocalStockCountry           interface{} `json:"localStockCountry"`
					MerchandiseLabel            string      `json:"merchandiseLabel"`
					OrderTimeFrame              interface{} `json:"orderTimeFrame"`
					PostCodesMessage            interface{} `json:"postCodesMessage"`
					ShippingAndFreeReturns      interface{} `json:"shippingAndFreeReturns"`
					ShippingAndFreeReturnsTitle interface{} `json:"shippingAndFreeReturnsTitle"`
					ShippingContactUs           interface{} `json:"shippingContactUs"`
					ShippingFromMessage         string      `json:"shippingFromMessage"`
					ShippingMarketplaceSeller   string      `json:"shippingMarketplaceSeller"`
					ShippingTitle               interface{} `json:"shippingTitle"`
					StartTime                   interface{} `json:"startTime"`
				} `json:"9445"`
				Default struct {
					CityID                      int64       `json:"cityId"`
					CountryCode                 string      `json:"countryCode"`
					DeliveryBefore              interface{} `json:"deliveryBefore"`
					DeliveryBy                  interface{} `json:"deliveryBy"`
					DeliveryCityMessage         interface{} `json:"deliveryCityMessage"`
					DeliveryGreetingsMessage    interface{} `json:"deliveryGreetingsMessage"`
					DeliveryIn                  interface{} `json:"deliveryIn"`
					DeliveryType                interface{} `json:"deliveryType"`
					EndTime                     interface{} `json:"endTime"`
					FarfetchOwned               interface{} `json:"farfetchOwned"`
					IsFromEurasianCustomsUnion  bool        `json:"isFromEurasianCustomsUnion"`
					IsLocalStock                bool        `json:"isLocalStock"`
					LocalStockCountry           interface{} `json:"localStockCountry"`
					MerchandiseLabel            string      `json:"merchandiseLabel"`
					OrderTimeFrame              interface{} `json:"orderTimeFrame"`
					PostCodesMessage            interface{} `json:"postCodesMessage"`
					ShippingAndFreeReturns      interface{} `json:"shippingAndFreeReturns"`
					ShippingAndFreeReturnsTitle interface{} `json:"shippingAndFreeReturnsTitle"`
					ShippingContactUs           interface{} `json:"shippingContactUs"`
					ShippingFromMessage         string      `json:"shippingFromMessage"`
					ShippingMarketplaceSeller   string      `json:"shippingMarketplaceSeller"`
					ShippingTitle               interface{} `json:"shippingTitle"`
					StartTime                   interface{} `json:"startTime"`
				} `json:"default"`
			} `json:"details"`
			VisibleOnDetails bool `json:"visibleOnDetails"`
		} `json:"shippingInformations"`
		SimilarProducts interface{} `json:"similarProducts"`
		Sizes           struct {
			Available struct {
				Two0 struct {
					Description string `json:"description"`
					LastInStock bool   `json:"lastInStock"`
					Quantity    int64  `json:"quantity"`
					SizeID      int64  `json:"sizeId"`
					StoreID     int64  `json:"storeId"`
					VariantID   string `json:"variantId"`
				} `json:"20"`
				Two1 struct {
					Description string `json:"description"`
					LastInStock bool   `json:"lastInStock"`
					Quantity    int64  `json:"quantity"`
					SizeID      int64  `json:"sizeId"`
					StoreID     int64  `json:"storeId"`
					VariantID   string `json:"variantId"`
				} `json:"21"`
				Two2 struct {
					Description string `json:"description"`
					LastInStock bool   `json:"lastInStock"`
					Quantity    int64  `json:"quantity"`
					SizeID      int64  `json:"sizeId"`
					StoreID     int64  `json:"storeId"`
					VariantID   string `json:"variantId"`
				} `json:"22"`
			} `json:"available"`
			CleanScaleDescription string      `json:"cleanScaleDescription"`
			ConvertedScaleID      interface{} `json:"convertedScaleId"`
			FriendlyScaleName     string      `json:"friendlyScaleName"`
			IsOneSize             bool        `json:"isOneSize"`
			IsOnlyOneLeft         bool        `json:"isOnlyOneLeft"`
			ScaleDescription      string      `json:"scaleDescription"`
			ScaleID               int64       `json:"scaleId"`
			SelectedSize          interface{} `json:"selectedSize"`
			ShowUnisex            bool        `json:"showUnisex"`
		} `json:"sizes"`
		ViewMoreLinks interface{} `json:"viewMoreLinks"`
	} `json:"productViewModel"`
	Promotions interface{} `json:"promotions"`
	Requests   struct {
		GetContacts          string `json:"getContacts"`
		GetSameStyleProducts string `json:"getSameStyleProducts"`
	} `json:"requests"`
	ShopTheLookInfo struct {
		Abtests struct {
			HasSkeleton                   bool `json:"hasSkeleton"`
			IsCompleteYourLookEnabled     bool `json:"isCompleteYourLookEnabled"`
			IsInComplementaryProductsMode bool `json:"isInComplementaryProductsMode"`
			IsInMainLookMode              bool `json:"isInMainLookMode"`
			IsInMainLookModeMobile        bool `json:"isInMainLookModeMobile"`
			IsInModalModeMobile           bool `json:"isInModalModeMobile"`
			IsStlRenamed                  bool `json:"isStlRenamed"`
		} `json:"abtests"`
		OutfitID   int64   `json:"outfitId"`
		ProductIds []int64 `json:"productIds"`
		Products   []int64 `json:"products"`
		Settings   struct {
			HasProductsOnline        bool  `json:"hasProductsOnline"`
			IsBlacklistedBrand       bool  `json:"isBlacklistedBrand"`
			IsMainProductShoesOrBags bool  `json:"isMainProductShoesOrBags"`
			IsSkeletonEnabled        bool  `json:"isSkeletonEnabled"`
			ModelImageStyle          int64 `json:"modelImageStyle"`
		} `json:"settings"`
	} `json:"shopTheLookInfo"`
	SizePredictor struct {
		FitAnalytics struct {
			AllProductIds    []string `json:"allProductIds"`
			CurrentProductID string   `json:"currentProductId"`
			MainThumbnail    string   `json:"mainThumbnail"`
			ShopCountry      string   `json:"shopCountry"`
			ShopLanguage     string   `json:"shopLanguage"`
			Sizes            []struct {
				Available        bool        `json:"available"`
				SizeAbbreviation interface{} `json:"sizeAbbreviation"`
				SizeDescription  string      `json:"sizeDescription"`
			} `json:"sizes"`
			UserID string `json:"userId"`
		} `json:"fitAnalytics"`
		FitPredictor struct {
			Alternative    string      `json:"alternative"`
			IsContextValid bool        `json:"isContextValid"`
			SizeSystemID   interface{} `json:"sizeSystemId"`
		} `json:"fitPredictor"`
		Zeekit struct {
			Alternative  string      `json:"alternative"`
			IsEnabled    bool        `json:"isEnabled"`
			IsMetric     bool        `json:"isMetric"`
			Labels       interface{} `json:"labels"`
			Language     interface{} `json:"language"`
			PdpScriptURL interface{} `json:"pdpScriptUrl"`
			ProductID    interface{} `json:"productId"`
			ProjectID    interface{} `json:"projectId"`
		} `json:"zeekit"`
	} `json:"sizePredictor"`
	StylingAdviceNavigation interface{} `json:"stylingAdviceNavigation"`
	Toggles                 struct {
		IsBackInStockEnabled          bool `json:"isBackInStockEnabled"`
		IsPreferredMerchantRedirected bool `json:"isPreferredMerchantRedirected"`
		IsRichTextEnabled             bool `json:"isRichTextEnabled"`
		IsSTLSkeletonEnabled          bool `json:"isSTLSkeletonEnabled"`
	} `json:"toggles"`
}


var (
	detailReg = regexp.MustCompile(`(window\['__initialState_slice-pdp__'\])\s*=\s*([^;]+)<\/script>`)
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
	
	 // write the whole body at once
	//  err = ioutil.WriteFile("C:\\Rinal\\ServiceBasedPRojects\\VoilaWork\\VoilaCrawl\\output_p.txt", respBody, 0644)
	//  if err != nil {
	// 	 panic(err)
	//  }

	matched := detailReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		c.logger.Debugf("data %s", respBody)
		return fmt.Errorf("extract produt json from page %s content failed", resp.Request.URL)
	}
	
	var (
		i      parseProductResponse
	)
	
	// err = ioutil.WriteFile("C:\\Rinal\\ServiceBasedPRojects\\VoilaWork\\VoilaCrawl\\output_p_0.txt", matched[0], 0644)
	// err = ioutil.WriteFile("C:\\Rinal\\ServiceBasedPRojects\\VoilaWork\\VoilaCrawl\\output_p_1.txt", matched[1], 0644)
	// err = ioutil.WriteFile("C:\\Rinal\\ServiceBasedPRojects\\VoilaWork\\VoilaCrawl\\output_p_2.txt", matched[2], 0644)

	if err = json.Unmarshal(matched[2], &i); err != nil {
		c.logger.Error(err)
		return err
	}
	//var IDs = i.ProductViewModel.Details.ProductID
	// if IDs.IsSSRMobile {

	// }

	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:       strconv.Format(i.ProductViewModel.Details.ProductID),
			CrawlUrl: resp.Request.URL.String(),
			//GroupId:  groupId,
		},
		Title:        i.ProductViewModel.Details.ShortDescription,
		Description:  i.ProductViewModel.Details.Description,
		BrandName:    i.ProductViewModel.DesignerDetails.Name,
		CrowdType:    i.ProductViewModel.Details.GenderName,
		Category:     "", // auto set by crawl job info
		SubCategory:  "",
		SubCategory2: "",
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
			Current:  int32(i.ProductViewModel.PriceInfo.Default.FinalPrice),
		},
		Stats: &pbItem.Stats{
			Rating:      0, //float32(rating.AverageOverallRating),
			ReviewCount: 0,  //int32(rating.TotalReviewCount),
		},
	}
	// if i.IsInStock { //ASK ?
	// 	item.Stock = &pbItem.Stock{
	// 		StockStatus: pbItem.Stock_InStock,
	// 	}
	// }
	for _, img := range i.ProductViewModel.Images.Main {
		itemImg, _ := anypb.New(&media.Media_Image{  // ask?
			OriginalUrl: strings.ReplaceAll(img.Zoom, "_1000.jpg", ""),
			LargeUrl:    img.Zoom, // $S$, $XXL$
			MediumUrl:   strings.ReplaceAll(img.Zoom, "_1000.jpg", "_600.jpg"),
			SmallUrl:    strings.ReplaceAll(img.Zoom, "_1000.jpg", "_400.jpg"),
		})
		item.Medias = append(item.Medias, &media.Media{
			Detail:    itemImg,
			// if img.Index == 1 {
			// 	IsDefault: img.IsPrimary, 
			// }
		})
	}

	// for _, variant := range i.ProductViewModel.Sizes.Available {
	// 	// vv, ok := variants[variant.VariantID] // ASK Why ?
	// 	// if !ok {
	// 	// 	continue
	// 	// }
	// 	sku := pbItem.Sku{
	// 		SourceId:    strconv.Format(variant.VariantID),
	// 		Title:       i.ProductViewModel.Details.ShortDescription,
	// 		Description: "",
	// 		Price: &pbItem.Price{				
	// 			Currency: regulation.Currency_USD,
	// 			//Current:  int32(vv.Price.Current.Value * 100), //ask ??
	// 		},
	// 		Stock: &pbItem.Stock{
	// 			StockStatus: pbItem.Stock_OutOfStock,
	// 		},
	// 		Specs: []*pbItem.SkuSpecOption{
	// 			{
	// 				Type:  pbItem.SkuSpecType_SkuSpecColor,
	// 				Name:  variant.Colour,
	// 				Value: strconv.Format(variant.ColourWayID),
	// 			},
	// 			{
	// 				Type:  pbItem.SkuSpecType_SkuSpecSize,
	// 				Name:  variant.Size,
	// 				Value: strconv.Format(variant.SizeID),
	// 			},
	// 		},
	// 	}
	// 	// if vv.IsInStock {
	// 	// 	sku.Stock.StockStatus = pbItem.Stock_InStock
	// 	// }
	// 	item.SkuItems = append(item.SkuItems, &sku)
	//  }
	return yield(ctx, &item)
}


func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		"https://www.farfetch.com/ae/shopping/women/gucci/items.aspx",		
		//"https://www.farfetch.com/shopping/women/gucci-x-ken-scott-floral-print-shirt-item-16359693.aspx?storeid=9445",
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
	var (
		apiToken = os.Getenv("PC_API_TOKEN")
		jsToken  = os.Getenv("PC_JS_TOKEN")
	)
	apiToken = "123"
	jsToken = "123"
	// if apiToken == "" || jsToken == "" {
	// 	panic("env PC_API_TOKEN or PC_JS_TOKEN is not set")
	// }

	logger := glog.New(glog.LogLevelDebug)
	client, err := proxy.NewProxyClient(
		cookiejar.New(), logger,
		proxy.Options{APIToken: apiToken, JSToken: jsToken},
	)
	if err != nil {
		panic(err)
	}

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

			resp, err := client.DoWithOptions(ctx, i, http.Options{EnableProxy: false})
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
