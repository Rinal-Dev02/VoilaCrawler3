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
	logger              glog.Log
}

func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient:          client,
		categoryPathMatcher: regexp.MustCompile(`^(/[a-z0-9_-]+)?/shopping/(women|men)([/a-z0-9_-]+)items.aspx$`),
		productPathMatcher:  regexp.MustCompile(`^(/[a-z0-9_-]+)+(/[a-z0-9_-]+)-item-[0-9]+.aspx$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "c0458359a95c408b9cb70d11c92f9ec7"
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
	options.EnableSessionInit = true
	options.Reliability = pbProxy.ProxyReliability_ReliabilityMedium
	options.MustCookies = append(options.MustCookies,
		&http.Cookie{Name: "ckm-ctx-sf", Value: `%2F`, Path: "/"},
	)
	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.farfetch.com"}
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

var prodDataExtraReg1 = regexp.MustCompile(`(window\['__initialState__'\])\s*=\s*"(.*)";</script>`)
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

	// next page
	matched := prodDataExtraReg.FindSubmatch(respBody)
	if matched == nil {
		// matched[2] = bytes.ReplaceAll(bytes.ReplaceAll(matched[2], []byte(`\"`), []byte(`"`)), []byte(`\\"`), []byte(`\\\"`))
		matched = prodDataExtraReg1.FindSubmatch(respBody) //__initialState__
	}
	if len(matched) <= 1 {
		c.httpClient.Jar().Clear(ctx, resp.Request.URL)
		return fmt.Errorf("extract json from product list page %s failed", resp.Request.URL)
	}
	var r struct {
		ListingItems struct {
			Items []struct {
				ID  int    `json:"id"`
				URL string `json:"url"`
			} `json:"items"`
		} `json:"listingItems"`
		ListingPagination struct {
			Index                int    `json:"index"`
			View                 int    `json:"view"`
			TotalItems           int    `json:"totalItems"`
			TotalPages           int    `json:"totalPages"`
			NormalizedTotalItems string `json:"normalizedTotalItems"`
		} `json:"listingPagination"`
	}

	// matched[2] = bytes.ReplaceAll(bytes.ReplaceAll(matched[2], []byte("\\'"), []byte("'")), []byte(`\\"`), []byte(`\"`))
	// rawData, err := strconv.Unquote(string(matched[1]))
	//if err != nil {
	//	c.logger.Errorf("unquote raw string failed, error=%s", err)
	//	return err
	//}
	if err = json.Unmarshal(matched[2], &r); err != nil {
		c.logger.Debugf("parse %s failed, error=%s", matched[1], err)
		return err
	}

	lastIndex := nextIndex(ctx)
	for _, prod := range r.ListingItems.Items {
		if prod.URL == "" {
			continue
		}

		if req, err := http.NewRequest(http.MethodGet, prod.URL, nil); err != nil {
			c.logger.Debug(err)
			return err
		} else {
			nctx := context.WithValue(ctx, "item.index", lastIndex+1)
			lastIndex += 1
			if err = yield(nctx, req); err != nil {
				return err
			}
		}
	}

	// get current page number
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}
	// check if this is the last page
	if page >= int64(r.ListingPagination.TotalPages) {
		return nil
	}

	// set pagination
	u := *resp.Request.URL
	vals := u.Query()
	vals.Set("page", strconv.Format(page+1))
	vals.Set("view", strconv.Format(r.ListingPagination.View))
	u.RawQuery = vals.Encode()

	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	// update the index of last page
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
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
				CurrencyCode                  string  `json:"currencyCode"`
				FinalPrice                    float32 `json:"finalPrice"`
				FormattedFinalPrice           string  `json:"formattedFinalPrice"`
				FormattedFinalPriceInternal   string  `json:"formattedFinalPriceInternal"`
				FormattedInitialPrice         string  `json:"formattedInitialPrice"`
				FormattedInitialPriceInternal string  `json:"formattedInitialPriceInternal"`
				InitialPrice                  float32 `json:"initialPrice"`
				IsOnSale                      bool    `json:"isOnSale"`
				Labels                        struct {
					Duties   string `json:"duties"`
					Discount string `json:"discount"`
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
			Details map[string]struct {
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
			} `json:"details"`
			VisibleOnDetails bool `json:"visibleOnDetails"`
		} `json:"shippingInformations"`
		SimilarProducts interface{} `json:"similarProducts"`
		Sizes           struct {
			Available map[string]struct {
				Description string `json:"description"`
				LastInStock bool   `json:"lastInStock"`
				Quantity    int64  `json:"quantity"`
				SizeID      int64  `json:"sizeId"`
				StoreID     int64  `json:"storeId"`
				VariantID   string `json:"variantId"`
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
	detailReg  = regexp.MustCompile(`(window\['__initialState_slice-pdp__'\])\s*=\s*([^;]+)<\/script>`)
	detailReg1 = regexp.MustCompile(`(window\['__initialState__']) = "([^;)]+)";`)
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
	if matched == nil {
		matched = detailReg1.FindSubmatch(respBody)
	}
	if len(matched) <= 1 {
		c.httpClient.Jar().Clear(ctx, resp.Request.URL)

		return fmt.Errorf("extract produt json from page %s content failed", resp.Request.URL)
	}

	var (
		i parseProductResponse
	)

	if err = json.Unmarshal(matched[2], &i); err != nil {
		c.logger.Error(err)
		return err
	}

	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:       strconv.Format(i.ProductViewModel.Details.ProductID),
			CrawlUrl: resp.Request.URL.String(),
		},
		Title:       i.ProductViewModel.Details.ShortDescription,
		Description: i.ProductViewModel.Details.Description,
		BrandName:   i.ProductViewModel.DesignerDetails.Name,
		CrowdType:   i.ProductViewModel.Details.GenderName,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
	}

	discount, _ := strconv.ParseInt(strings.TrimSuffix(i.ProductViewModel.PriceInfo.Default.Labels.Discount, "% Off"))
	current, _ := strconv.ParseFloat(i.ProductViewModel.PriceInfo.Default.FinalPrice)
	msrp, _ := strconv.ParseFloat(i.ProductViewModel.PriceInfo.Default.InitialPrice)

	for _, rawSize := range i.ProductViewModel.Sizes.Available {
		sku := pbItem.Sku{
			SourceId: strconv.Format(rawSize.SizeID),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(current * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},
			Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}
		if rawSize.Quantity > 0 {
			sku.Stock.StockStatus = pbItem.Stock_InStock
			sku.Stock.StockCount = int32(rawSize.Quantity)
		}

		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Id:    strconv.Format(rawSize.SizeID),
			Name:  rawSize.Description,
			Value: strconv.Format(rawSize.SizeID),
		})

		item.SkuItems = append(item.SkuItems, &sku)
	}

	isDefault := true
	for _, img := range i.ProductViewModel.Images.Main {
		if img.Index > 1 {
			isDefault = false
		}
		itemImg, _ := anypb.New(&media.Media_Image{
			OriginalUrl: img.Zoom,
			LargeUrl:    img.Zoom, // $S$, $XXL$
			MediumUrl:   strings.ReplaceAll(img.Zoom, "_1000.jpg", "_600.jpg"),
			SmallUrl:    strings.ReplaceAll(img.Zoom, "_1000.jpg", "_400.jpg"),
		})
		item.Medias = append(item.Medias, &media.Media{
			Detail:    itemImg,
			IsDefault: isDefault,
		})
	}

	// yield item result
	if err = yield(ctx, &item); err != nil {
		return err
	}
	return nil
}

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		"https://www.farfetch.com/shopping/women/denim-1/items.aspx",
		//"https://www.farfetch.com/shopping/women/gucci-x-ken-scott-floral-print-shirt-item-16359693.aspx?storeid=9445",
		//"https://www.farfetch.com/shopping/women/escada-floral-print-shirt-item-13761571.aspx?rtype=portal_pdp_outofstock_b&rpos=3&rid=027c2611-6135-4842-abdd-59895d30e924",
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
				i.URL.Host = "www.farfetch.com"
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
			// data, err := json.Marshal(i)
			// if err != nil {
			// 	return err
			// }
			// logger.Infof("data: %s", data)
			logger.Debugf("got item")
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
