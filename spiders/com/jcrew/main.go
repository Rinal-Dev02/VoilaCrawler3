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
	return "33d5a6c960e344b081a34fbf554b0a48"
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
	}
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.jcrew.com"}
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

type CategoryView struct {
	Props struct {
		IsServer     bool `json:"isServer"`
		InitialState struct {
			Array struct {
				Data struct {
					HasSplitResults bool `json:"hasSplitResults"`
					IsFetching      bool `json:"isFetching"`
					IsFiltering     bool `json:"isFiltering"`
					ProductArray    struct {
						CategoryHeader struct {
							CatLink                     interface{} `json:"catLink"`
							SkuOnlyFolder               string      `json:"skuOnlyFolder"`
							ExtLink                     string      `json:"extLink"`
							CreativeJsPath              string      `json:"creativeJsPath"`
							CreativeCSSPath             string      `json:"creativeCssPath"`
							Name                        string      `json:"name"`
							DimID                       int         `json:"dimId"`
							HasSubFolders               string      `json:"hasSubFolders"`
							SeoH1                       string      `json:"seoH1"`
							ATRExtendedSizeFolder       string      `json:"aTRExtendedSizeFolder"`
							DisplayCategorySeo          bool        `json:"displayCategorySeo"`
							CatStyle                    string      `json:"catStyle"`
							DisplayRectangularImageCrop bool        `json:"displayRectangularImageCrop"`
							CatLocation                 string      `json:"catLocation"`
							Priority                    string      `json:"priority"`
							CatDescription              interface{} `json:"catDescription"`
							Label                       string      `json:"label"`
							InvalidSubcategory          bool        `json:"invalidSubcategory"`
							SeoHTMLContent              string      `json:"seoHTMLContent"`
							CatName                     string      `json:"catName"`
							FolderID                    string      `json:"folderId"`
							Gender                      string      `json:"gender"`
							PersistCreativeObjects      bool        `json:"persistCreativeObjects"`
							ATROmniProp25               string      `json:"aTROmniProp25"`
							MixedProducts               string      `json:"mixedProducts"`
						} `json:"categoryHeader"`
						SortByOrderType []struct {
							Label string `json:"label"`
							Value string `json:"value"`
						} `json:"sortByOrderType"`
						ProductList []struct {
							Header   string `json:"header"`
							Products []struct {
								ProductID          string `json:"productId"`
								ProductDescription string `json:"productDescription"`
								URL                string `json:"url"`
								ProductCode        string `json:"productCode"`
							} `json:"products"`
						} `json:"productList"`
						Navigation struct {
							Refinements []struct {
								Name   string `json:"name"`
								Values []struct {
									ID        int    `json:"id"`
									Count     string `json:"count"`
									Label     string `json:"label"`
									SortOrder int    `json:"sortOrder"`
									QueryName string `json:"queryName"`
									Gender    string `json:"gender"`
									FolderID  string `json:"folderId"`
									Seo       string `json:"seo"`
								} `json:"values"`
								Label         string `json:"label,omitempty"`
								Priority      int    `json:"priority"`
								SelectedCount int    `json:"selectedCount,omitempty"`
							} `json:"refinements"`
							Breadcrumbs []interface{} `json:"breadcrumbs"`
						} `json:"navigation"`
						ResultCount int `json:"resultCount"`
						Pagination  struct {
							PageIndex int `json:"pageIndex"`
							TotalPage int `json:"totalPage"`
						} `json:"pagination"`
					} `json:"productArray"`
					SearchTerm       string `json:"searchTerm"`
					LastFilterAction struct {
					} `json:"lastFilterAction"`
					DefaultRefinements []string `json:"defaultRefinements"`
				} `json:"data"`
			} `json:"array"`
		} `json:"initialState"`
	} `json:"props"`
	Page  string `json:"page"`
	Query struct {
		CountryCode string `json:"countryCode"`
		Gender      string `json:"gender"`
		Category    string `json:"category"`
	} `json:"query"`
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
	for _, idv := range viewData.Props.InitialState.Array.Data.ProductArray.ProductList[0].Products {

		req, err := http.NewRequest(http.MethodGet, idv.URL, nil)
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
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("Npge"))
	if page == 0 {
		page = 1
	}
	// check if this is the last page
	if len(viewData.Props.InitialState.Array.Data.ProductArray.ProductList[0].Products) >= viewData.Props.InitialState.Array.Data.ProductArray.ResultCount ||
		page >= int64(viewData.Props.InitialState.Array.Data.ProductArray.Pagination.TotalPage) {
		return nil
	}

	// set pagination
	u := *resp.Request.URL
	vals := u.Query()
	vals.Set("Npge", strconv.Format(page+1))
	u.RawQuery = vals.Encode()

	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	// update the index of last page
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
}

type productPageResponse struct {
	Props struct {
		IsServer     bool `json:"isServer"`
		InitialState struct {
			Signin struct {
				Login struct {
					DidInvalidate         bool        `json:"didInvalidate"`
					IsFetching            bool        `json:"isFetching"`
					FetchLoginComplete    bool        `json:"fetchLoginComplete"`
					SuccessComplete       bool        `json:"successComplete"`
					DidEmailInvalidate    bool        `json:"didEmailInvalidate"`
					DidPasswordInvalidate bool        `json:"didPasswordInvalidate"`
					ShowSigninModal       bool        `json:"showSigninModal"`
					IsForgotPassword      bool        `json:"isForgotPassword"`
					IsResetPassword       bool        `json:"isResetPassword"`
					IsResetLinkExpired    bool        `json:"isResetLinkExpired"`
					ToLink                string      `json:"toLink"`
					Response              interface{} `json:"response"`
					UserResetEmail        string      `json:"userResetEmail"`
				} `json:"login"`
				Register struct {
					DidInvalidate               bool   `json:"didInvalidate"`
					IsFetching                  bool   `json:"isFetching"`
					FetchRegisterComplete       bool   `json:"fetchRegisterComplete"`
					SuccessComplete             bool   `json:"successComplete"`
					CountryCode                 string `json:"countryCode"`
					CountryName                 string `json:"countryName"`
					DidEmailInvalidate          bool   `json:"didEmailInvalidate"`
					DidPasswordInvalidate       bool   `json:"didPasswordInvalidate"`
					DidFirstNameInvalidate      bool   `json:"didFirstNameInvalidate"`
					DidLastNameInvalidate       bool   `json:"didLastNameInvalidate"`
					DidBirthDateMonthInvalidate bool   `json:"didBirthDateMonthInvalidate"`
					DidBirthDateDayInvalidate   bool   `json:"didBirthDateDayInvalidate"`
					DidDateInvalidate           bool   `json:"didDateInvalidate"`
					ShowRegisterModal           bool   `json:"showRegisterModal"`
				} `json:"register"`
			} `json:"signin"`
			Content struct {
				FooterContent struct {
					SlTranslate string `json:"sl_translate"`
					Help        struct {
						Content struct {
							Label         string `json:"label"`
							Twitter       string `json:"twitter"`
							TelephoneInfo []struct {
								TelephoneText string `json:"telephoneText"`
								Country       string `json:"country,omitempty"`
							} `json:"telephoneInfo"`
							ContactEmailIntl string `json:"contactEmailIntl"`
							ContactEmail     string `json:"contactEmail"`
							LiveChat         string `json:"liveChat"`
						} `json:"content"`
						Urls struct {
							Twitter          string `json:"twitter"`
							ContactEmail     string `json:"contactEmail"`
							ContactEmailIntl string `json:"contactEmailIntl"`
							LiveChat         string `json:"liveChat"`
						} `json:"urls"`
					} `json:"help"`
					Modal struct {
						Loyalty struct {
							CallToActionText string `json:"callToActionText"`
							CallToActionURL  string `json:"callToActionUrl"`
							Header           string `json:"header"`
							Intro            string `json:"intro"`
						} `json:"loyalty"`
					} `json:"modal"`
					Signup struct {
						Label              string `json:"label"`
						PlaceHolder        string `json:"placeHolder"`
						Button             string `json:"button"`
						SignupText         string `json:"signupText"`
						InvalidEmailText   string `json:"invalidEmailText"`
						SubmitSuccessLine1 string `json:"submitSuccessLine1"`
						SubmitSuccessLine2 string `json:"submitSuccessLine2"`
					} `json:"signup"`
					CountryContext struct {
						Content struct {
							ShipTo  string `json:"shipTo"`
							Flag    string `json:"flag"`
							Country string `json:"country"`
							Change  string `json:"change"`
						} `json:"content"`
						Urls struct {
							Change string `json:"change"`
						} `json:"urls"`
					} `json:"countryContext"`
					FullSite struct {
						Label string `json:"label"`
						URL   string `json:"url"`
					} `json:"fullSite"`
					SizeCharts struct {
						Label string `json:"label"`
						URL   string `json:"url"`
					} `json:"sizeCharts"`
					SeoPromo struct {
						Label string `json:"label"`
						URL   string `json:"url"`
					} `json:"seoPromo"`
					SafetyRecall struct {
						Message1  string `json:"message1"`
						Message2  string `json:"message2"`
						AriaLabel string `json:"ariaLabel"`
						Link      string `json:"link"`
					} `json:"safetyRecall"`
					AdditionalLinks []struct {
						Label       string `json:"label"`
						ColumnIndex int    `json:"columnIndex"`
						RowIndex    int    `json:"rowIndex"`
						CountryCode string `json:"countryCode"`
						List        []struct {
							Label string `json:"label"`
							Link  string `json:"link"`
						} `json:"list"`
					} `json:"additionalLinks"`
				} `json:"footerContent"`
				HeaderContent struct {
					SlTranslate string `json:"sl_translate"`
					HeaderInfo  struct {
						Content struct {
							Seo struct {
								H1              string `json:"h1"`
								MetaDescription string `json:"metaDescription"`
								Title           string `json:"title"`
								SiteName        string `json:"siteName"`
							} `json:"seo"`
							Header struct {
								Menu                    string `json:"menu"`
								Search                  string `json:"search"`
								Stores                  string `json:"stores"`
								Signin                  string `json:"signin"`
								Or                      string `json:"or"`
								Register                string `json:"register"`
								Wishlist                string `json:"wishlist"`
								Shoppingbag             string `json:"shoppingbag"`
								MyRewards               string `json:"myRewards"`
								Account                 string `json:"account"`
								MyAccount               string `json:"myAccount"`
								MyDetails               string `json:"myDetails"`
								Welcome                 string `json:"welcome"`
								OrderHistory            string `json:"orderHistory"`
								SignOut                 string `json:"signOut"`
								RewardsStatus           string `json:"rewardsStatus"`
								Country                 string `json:"country"`
								ViewAll                 string `json:"viewAll"`
								BackToShoppingBag       string `json:"backToShoppingBag"`
								AssociateSignin         string `json:"associateSignin"`
								AssociateSignout        string `json:"associateSignout"`
								BrowsingAsCustomer      string `json:"browsingAsCustomer"`
								TimeLeft                string `json:"timeLeft"`
								ExtendSession           string `json:"extendSession"`
								AssociateSessionExpired string `json:"associateSessionExpired"`
							} `json:"header"`
							Search struct {
								PlaceHolder string `json:"placeHolder"`
								Text        string `json:"text"`
								Button      string `json:"button"`
							} `json:"search"`
							FactoryLink struct {
								Text string `json:"text"`
								URL  string `json:"url"`
							} `json:"factoryLink"`
							Nojs struct {
								Message string `json:"message"`
								Help    string `json:"help"`
							} `json:"nojs"`
						} `json:"content"`
						HeaderURL struct {
							RegisterURL            string `json:"registerUrl"`
							SigninURL              string `json:"signinUrl"`
							SignoutURL             string `json:"signoutUrl"`
							OrderhistoryURL        string `json:"orderhistoryUrl"`
							WishlistURL            string `json:"wishlistUrl"`
							ShoppingbagURL         string `json:"shoppingbagUrl"`
							CountryURL             string `json:"countryUrl"`
							StoresURL              string `json:"storesUrl"`
							LoyaltyDashboardURL    string `json:"loyaltyDashboardURL"`
							AccountHomeURL         string `json:"accountHomeURL"`
							AccountDetailsURL      string `json:"accountDetailsURL"`
							SidecarOrderHistoryURL string `json:"sidecarOrderHistoryURL"`
							AssociateSigninURL     string `json:"associateSigninUrl"`
						} `json:"headerUrl"`
						Months  []string `json:"months"`
						Rewards struct {
							Jcrew           string `json:"jcrew"`
							CreditCardName  string `json:"creditCardName"`
							CreditCard      string `json:"creditCard"`
							Balance         string `json:"balance"`
							LastUpdated     string `json:"lastUpdated"`
							LastUpdatedTime string `json:"lastUpdatedTime"`
							Points          string `json:"points"`
							CurrentPoints   string `json:"currentPoints"`
							PointsToNext    string `json:"pointsToNext"`
							AccountManage   string `json:"accountManage"`
							PromoMessage1   string `json:"promoMessage1"`
							PromoMessage2   string `json:"promoMessage2"`
							Urls            struct {
								AccountManage  string `json:"accountManage"`
								RewardsBalance string `json:"rewardsBalance"`
								Help           string `json:"help"`
							} `json:"urls"`
						} `json:"rewards"`
						WelcomeMatReact struct {
							JcrewClothing string `json:"jcrewClothing"`
							HelloCanada   string `json:"helloCanada"`
							CanadaByline  struct {
								Text1 string `json:"text1"`
								Text2 string `json:"text2"`
								Text3 string `json:"text3"`
							} `json:"canadaByline"`
							ShipsAll           string `json:"shipsAll"`
							AroundWorld        string `json:"aroundWorld"`
							FlatRateRestOfIntl string `json:"flatRateRestOfIntl"`
							FlatRateCanadaMsg1 string `json:"flatRateCanadaMsg1"`
							FlatRateCanadaMsg2 string `json:"flatRateCanadaMsg2"`
							DutyFree           string `json:"dutyFree"`
							DutyFreeCanadaMsg1 string `json:"dutyFreeCanadaMsg1"`
							DutyFreeCanadaMsg2 string `json:"dutyFreeCanadaMsg2"`
							NeedHelp           string `json:"needHelp"`
							NeedHelpCanadaMsg1 string `json:"needHelpCanadaMsg1"`
							NeedHelpCanadaMsg2 struct {
								PartOne string `json:"partOne"`
								Email   string `json:"email"`
								Phone   string `json:"phone"`
							} `json:"needHelpCanadaMsg2"`
							ContactText struct {
								EmailText string `json:"emailText"`
								OrText    string `json:"orText"`
								CallText  string `json:"callText"`
								AndText   string `json:"andText"`
							} `json:"contactText"`
							Contact []struct {
								Country string `json:"country"`
								Info    struct {
									Email string `json:"email"`
									Phone string `json:"phone"`
								} `json:"info"`
							} `json:"contact"`
							StartShopping string `json:"startShopping"`
							TakeMeTo      string `json:"takeMeTo"`
							Terms         struct {
								Text1         string `json:"text1"`
								TermsOfUse    string `json:"termsOfUse"`
								PrivacyPolicy string `json:"privacyPolicy"`
								Text2         string `json:"text2"`
								Cookies       string `json:"cookies"`
								Text3         string `json:"text3"`
							} `json:"terms"`
						} `json:"welcomeMatReact"`
						WelcomeMat struct {
							HelloCanada        string `json:"helloCanada"`
							CanadaByline       string `json:"canadaByline"`
							ShipsAll           string `json:"shipsAll"`
							AroundWorld        string `json:"aroundWorld"`
							FlatRate           string `json:"flatRate"`
							FlatRateCanadaMsg1 string `json:"flatRateCanadaMsg1"`
							FlatRateCanadaMsg2 string `json:"flatRateCanadaMsg2"`
							DutyFree           string `json:"dutyFree"`
							DutyFreeCanadaMsg1 string `json:"dutyFreeCanadaMsg1"`
							DutyFreeCanadaMsg2 string `json:"dutyFreeCanadaMsg2"`
							NeedHelp           string `json:"needHelp"`
							NeedHelpCanadaMsg1 string `json:"needHelpCanadaMsg1"`
							NeedHelpCanadaMsg2 string `json:"needHelpCanadaMsg2"`
							Contact            []struct {
								Country string `json:"country"`
								Info    string `json:"info"`
							} `json:"contact"`
							StartShopping string `json:"startShopping"`
							TakeMeTo      string `json:"takeMeTo"`
							Terms         string `json:"terms"`
						} `json:"welcomeMat"`
					} `json:"headerInfo"`
					ProgressBar struct {
						InitialValue string `json:"initialValue"`
						FreeShipping struct {
							IncompleteTitle       string `json:"incompleteTitle"`
							CompleteTitle         string `json:"completeTitle"`
							CompleteBottomMessage string `json:"completeBottomMessage"`
						} `json:"freeShipping"`
					} `json:"progressBar"`
					GlobalPromo struct {
						Details string `json:"details"`
						Close   string `json:"close"`
					} `json:"globalPromo"`
					InlineError struct {
						OopsText   string `json:"oopsText"`
						PleaseText string `json:"pleaseText"`
						RetryText  string `json:"retryText"`
					} `json:"inlineError"`
					EmailCapture struct {
						Default struct {
							PromoHeaderText       string `json:"promoHeaderText"`
							PromoHeaderTextMobile string `json:"promoHeaderTextMobile"`
							PromoIntroText        string `json:"promoIntroText"`
							ButtonText            string `json:"buttonText"`
							ButtonTextDesktop     string `json:"buttonTextDesktop"`
							HeaderText            string `json:"headerText"`
							HeaderTextMobile      string `json:"headerTextMobile"`
							IntroText             string `json:"introText"`
							IntroTextDesktop      string `json:"introTextDesktop"`
							InvalidEmailText      string `json:"invalidEmailText"`
							PlaceholderText       string `json:"placeholderText"`
							PrivacyPolicyText     string `json:"privacyPolicyText"`
							PrivacyPolicyURL      string `json:"privacyPolicyUrl"`
							SeeText               string `json:"seeText"`
							SeeTextDesktop1       string `json:"seeTextDesktop1"`
							SeeTextDesktop2       string `json:"seeTextDesktop2"`
							SuccessHeaderText     string `json:"successHeaderText"`
							SuccessText           string `json:"successText"`
							Terms                 []struct {
								CountryCode string `json:"countryCode"`
								Copy        string `json:"copy"`
							} `json:"terms"`
							TermsReact struct {
								Hk struct {
									Copy1 string `json:"copy1"`
									Copy2 string `json:"copy2"`
								} `json:"hk"`
								Fr struct {
									Copy1       string `json:"copy1"`
									Copy2       string `json:"copy2"`
									Email       string `json:"email"`
									PrivacyText string `json:"privacyText"`
								} `json:"fr"`
							} `json:"termsReact"`
							Details  string `json:"details"`
							Close    string `json:"close"`
							Nothanks string `json:"nothanks"`
						} `json:"default"`
						Affiliates struct {
							ButtonText        string `json:"buttonText"`
							ButtonTextDesktop string `json:"buttonTextDesktop"`
							HeaderText        string `json:"headerText"`
							IntroText         string `json:"introText"`
							IntroTextDesktop  string `json:"introTextDesktop"`
							InvalidEmailText  string `json:"invalidEmailText"`
							PlaceholderText   string `json:"placeholderText"`
							PrivacyPolicyText string `json:"privacyPolicyText"`
							PrivacyPolicyURL  string `json:"privacyPolicyUrl"`
							SeeText           string `json:"seeText"`
							SeeTextDesktop1   string `json:"seeTextDesktop1"`
							SeeTextDesktop2   string `json:"seeTextDesktop2"`
							SuccessHeaderText string `json:"successHeaderText"`
							SuccessText       string `json:"successText"`
							Terms             []struct {
								CountryCode string `json:"countryCode"`
								Copy        string `json:"copy"`
							} `json:"terms"`
							TermsReact struct {
								Hk struct {
									Copy1 string `json:"copy1"`
									Copy2 string `json:"copy2"`
								} `json:"hk"`
								Fr struct {
									Copy1       string `json:"copy1"`
									Copy2       string `json:"copy2"`
									Email       string `json:"email"`
									PrivacyText string `json:"privacyText"`
								} `json:"fr"`
							} `json:"termsReact"`
							Details  string `json:"details"`
							Close    string `json:"close"`
							Nothanks string `json:"nothanks"`
						} `json:"affiliates"`
						Social struct {
							ButtonText        string `json:"buttonText"`
							ButtonTextDesktop string `json:"buttonTextDesktop"`
							HeaderText        string `json:"headerText"`
							IntroText         string `json:"introText"`
							IntroTextDesktop  string `json:"introTextDesktop"`
							InvalidEmailText  string `json:"invalidEmailText"`
							PlaceholderText   string `json:"placeholderText"`
							PrivacyPolicyText string `json:"privacyPolicyText"`
							PrivacyPolicyURL  string `json:"privacyPolicyUrl"`
							SeeText           string `json:"seeText"`
							SeeTextDesktop1   string `json:"seeTextDesktop1"`
							SeeTextDesktop2   string `json:"seeTextDesktop2"`
							SuccessHeaderText string `json:"successHeaderText"`
							SuccessText       string `json:"successText"`
							Terms             []struct {
								CountryCode string `json:"countryCode"`
								Copy        string `json:"copy"`
							} `json:"terms"`
							TermsReact struct {
								Hk struct {
									Copy1 string `json:"copy1"`
									Copy2 string `json:"copy2"`
								} `json:"hk"`
								Fr struct {
									Copy1       string `json:"copy1"`
									Copy2       string `json:"copy2"`
									Email       string `json:"email"`
									PrivacyText string `json:"privacyText"`
								} `json:"fr"`
							} `json:"termsReact"`
							Details  string `json:"details"`
							Close    string `json:"close"`
							Nothanks string `json:"nothanks"`
						} `json:"social"`
					} `json:"emailCapture"`
					PasswordCapture struct {
						ButtonTextDesktop    string `json:"buttonTextDesktop"`
						HeaderText           string `json:"headerText"`
						PreIntroTextDesktop  string `json:"preIntroTextDesktop"`
						IntroLinkTextDesktop string `json:"introLinkTextDesktop"`
						IntroLinkURL         string `json:"introLinkUrl"`
						PostIntroTextDesktop string `json:"postIntroTextDesktop"`
						PostIntroTextMobile  string `json:"postIntroTextMobile"`
						InvalidPasswordText  string `json:"invalidPasswordText"`
						PlaceholderText      string `json:"placeholderText"`
						TermsDesktop         string `json:"termsDesktop"`
						TermsMobile          string `json:"termsMobile"`
						TermsText            string `json:"termsText"`
						TermsURL             string `json:"termsUrl"`
						PrivacyPolicyText    string `json:"privacyPolicyText"`
						PrivacyPolicyURL     string `json:"privacyPolicyUrl"`
						SeeTextDesktop1      string `json:"seeTextDesktop1"`
						SeeTextDesktop2      string `json:"seeTextDesktop2"`
						SuccessHeaderText    string `json:"successHeaderText"`
						SuccessText          string `json:"successText"`
						Nothanks             string `json:"nothanks"`
					} `json:"passwordCapture"`
					MiniBag struct {
						MobileText        string `json:"mobileText"`
						Backordered       string `json:"backordered"`
						Checkout          string `json:"checkout"`
						CheckoutTitle     string `json:"checkoutTitle"`
						MonogramMessage   string `json:"monogramMessage"`
						Edit              string `json:"edit"`
						Remove            string `json:"remove"`
						ShoppingBagLink   string `json:"shoppingBagLink"`
						Preordered        string `json:"preordered"`
						ThereIs           string `json:"thereIs"`
						ThereAre          string `json:"thereAre"`
						SingleItem        string `json:"singleItem"`
						MultipleItems     string `json:"multipleItems"`
						ShoppingBag       string `json:"shoppingBag"`
						ShoppingBagTitle  string `json:"shoppingBagTitle"`
						Subtotal          string `json:"subtotal"`
						PartialQuantity1  string `json:"partialQuantity1"`
						PartialQuantity2  string `json:"partialQuantity2"`
						Size              string `json:"size"`
						Color             string `json:"color"`
						SaleAlert         string `json:"saleAlert"`
						LowInventoryAlert string `json:"lowInventoryAlert"`
						Quantity          string `json:"quantity"`
						ModalTitle        string `json:"modalTitle"`
						ContinueBtn       string `json:"continueBtn"`
						CheckoutNowBtn    string `json:"checkoutNowBtn"`
					} `json:"miniBag"`
					CreateAccount struct {
						HeaderText1        string `json:"headerText1"`
						HeaderText2        string `json:"headerText2"`
						HeaderText3        string `json:"headerText3"`
						Blurb1             string `json:"blurb1"`
						Blurb2             string `json:"blurb2"`
						Blurb3             string `json:"blurb3"`
						FirstName          string `json:"firstName"`
						LastName           string `json:"lastName"`
						Password           string `json:"password"`
						SignupLarge        string `json:"signupLarge"`
						SignupSmall        string `json:"signupSmall"`
						InvalidFirstName   string `json:"invalidFirstName"`
						InvalidLastName    string `json:"invalidLastName"`
						InvalidPassword    string `json:"invalidPassword"`
						EmailAlreadyExists string `json:"emailAlreadyExists"`
						SuccessHeader      string `json:"successHeader"`
						SuccessBlurb       string `json:"successBlurb"`
						StartShopping      string `json:"startShopping"`
					} `json:"createAccount"`
					BagAlerts struct {
						GreetingPrimary   string `json:"greetingPrimary"`
						GreetingSecondary string `json:"greetingSecondary"`
						Abandoned         string `json:"abandoned"`
						An                string `json:"an"`
						AnUpper           string `json:"anUpper"`
						SingleItem        string `json:"singleItem"`
						MultiItems        string `json:"multiItems"`
						InYour            string `json:"inYour"`
						ShoppingBag       string `json:"shoppingBag"`
						SingleVerb        string `json:"singleVerb"`
						PluralVerb        string `json:"pluralVerb"`
						NowOnSale         string `json:"nowOnSale"`
						LowInventory      string `json:"lowInventory"`
					} `json:"bagAlerts"`
					LoyaltyMessage struct {
						Title     string `json:"title"`
						MsgLine1  string `json:"msgLine1"`
						MsgLine2  string `json:"msgLine2"`
						MsgLine3  string `json:"msgLine3"`
						Signup    string `json:"signup"`
						SignupURL string `json:"signupUrl"`
						MsgLine4  string `json:"msgLine4"`
					} `json:"loyaltyMessage"`
				} `json:"headerContent"`
				Navigation struct {
					SlTranslate string `json:"sl_translate"`
					Labels      struct {
						HomeButtonText         string `json:"homeButtonText"`
						BackButtonText         string `json:"backButtonText"`
						SignInButtonText       string `json:"signInButtonText"`
						MyAccountButtonText    string `json:"myAccountButtonText"`
						StoreLocatorButtonText string `json:"storeLocatorButtonText"`
						CtaText                string `json:"ctaText"`
						FeaturedStoryHeader    string `json:"featuredStoryHeader"`
						FeaturedStoryLink      string `json:"featuredStoryLink"`
						SaleLink               string `json:"saleLink"`
						SaleGroupLabel         string `json:"saleGroupLabel"`
						SaleText               string `json:"saleText"`
						ShopText               string `json:"shopText"`
						Items                  string `json:"items"`
					} `json:"labels"`
					Urls struct {
						HomeURL         string `json:"homeUrl"`
						SignInURL       string `json:"signInUrl"`
						MyAccountURL    string `json:"myAccountUrl"`
						StoreLocatorURL string `json:"storeLocatorUrl"`
					} `json:"urls"`
				} `json:"navigation"`
				P struct {
					SlTranslate string `json:"sl_translate"`
					Product     struct {
						AddToBagSingleText       string `json:"addToBagSingleText"`
						AddToBagText             string `json:"addToBagText"`
						AddToBagOfflineText      string `json:"addToBagOfflineText"`
						AddToWishlistOfflineText string `json:"addToWishlistOfflineText"`
						BackorderedNoSizeMessage string `json:"backorderedNoSizeMessage"`
						BackorderedSizeMessage   string `json:"backorderedSizeMessage"`
						BackorderedText          string `json:"backorderedText"`
						BackToProduct            string `json:"backToProduct"`
						BasketQuotaExceededText1 string `json:"basketQuotaExceededText1"`
						BasketQuotaExceededText2 string `json:"basketQuotaExceededText2"`
						BasketQuotaExceededText3 string `json:"basketQuotaExceededText3"`
						CallText                 string `json:"callText"`
						ColorText                string `json:"colorText"`
						ConfirmAddItemMessage    string `json:"confirmAddItemMessage"`
						FinalSaleMessage         string `json:"finalSaleMessage"`
						FinalSaleText            string `json:"finalSaleText"`
						FitText                  string `json:"fitText"`
						GenericAddToBagText      string `json:"genericAddToBagText"`
						InStockText              string `json:"inStockText"`
						IntlShippingHeadline     string `json:"intlShippingHeadline"`
						IntlShippingBody         string `json:"intlShippingBody"`
						IntlShippingContact      string `json:"intlShippingContact"`
						IntlTaxMessage           string `json:"intlTaxMessage"`
						ItemText                 string `json:"itemText"`
						FreeShippingText         string `json:"freeShippingText"`
						JustReduced              string `json:"justReduced"`
						LowInventoryText         string `json:"lowInventoryText"`
						MarketplaceHelp          struct {
							HelpDescription1 string `json:"helpDescription1"`
							HelpDescription2 string `json:"helpDescription2"`
							HelpTitle        string `json:"helpTitle"`
						} `json:"marketplaceHelp"`
						MarketplaceReturns struct {
							FirstSection            string `json:"firstSection"`
							SecondSection           string `json:"secondSection"`
							ShippingAndReturnsLabel string `json:"shippingAndReturnsLabel"`
							ShippingPolicy          string `json:"shippingPolicy"`
							And                     string `json:"and"`
							ReturnsPolicy           string `json:"returnsPolicy"`
						} `json:"marketplaceReturns"`
						MonogramMessage1          string `json:"monogramMessage1"`
						MonogramMessage2          string `json:"monogramMessage2"`
						NoLongerAvailableMessage  string `json:"noLongerAvailableMessage"`
						NowText                   string `json:"nowText"`
						OrText                    string `json:"orText"`
						OutOfStockText            string `json:"outOfStockText"`
						PartialQuantityMessage1   string `json:"partialQuantityMessage1"`
						PartialQuantityMessage2   string `json:"partialQuantityMessage2"`
						PartialQuantityMessage3   string `json:"partialQuantityMessage3"`
						PercentOffText            string `json:"percentOffText"`
						PleaseText                string `json:"pleaseText"`
						PreorderNoSizeMessage     string `json:"preorderNoSizeMessage"`
						PreorderText              string `json:"preorderText"`
						ProductDescriptionText    string `json:"productDescriptionText"`
						QuantityText              string `json:"quantityText"`
						RestrictedStateText       string `json:"restrictedStateText"`
						Share                     string `json:"share"`
						ShipMessage               string `json:"shipMessage"`
						SelectColorsText          string `json:"selectColorsText"`
						SelectSizeMessage         string `json:"selectSizeMessage"`
						SoldOutMessage1           string `json:"soldOutMessage1"`
						SiteName                  string `json:"siteName"`
						SizeChartText             string `json:"sizeChartText"`
						SizeFitLink               string `json:"sizeFitLink"`
						SizeFitText               string `json:"sizeFitText"`
						SizeText                  string `json:"sizeText"`
						SelectSizeText            string `json:"selectSizeText"`
						UsSizeText                string `json:"usSizeText"`
						ViewFullProductText       string `json:"viewFullProductText"`
						VpsEmail                  string `json:"vpsEmail"`
						VpsTelephone              string `json:"vpsTelephone"`
						VpsMessage1               string `json:"vpsMessage1"`
						VpsMessage2               string `json:"vpsMessage2"`
						VpsMessage3               string `json:"vpsMessage3"`
						WasText                   string `json:"wasText"`
						ThereIsNow                string `json:"thereIsNow"`
						ThereAreNow               string `json:"thereAreNow"`
						AddToWishlistSingleText   string `json:"addToWishlistSingleText"`
						AddToWishlistMultipleText string `json:"addToWishlistMultipleText"`
						AddToWishlistFullMsg1     string `json:"addToWishlistFullMsg1"`
						AddToWishlistFullMsg2     string `json:"addToWishlistFullMsg2"`
						AddToWishlistFullMsg3     string `json:"addToWishlistFullMsg3"`
						ShopTheLook               string `json:"shopTheLook"`
						GoodNewsThisItemShipsFree string `json:"goodNewsThisItemShipsFree"`
						WantFreeShipping          string `json:"wantFreeShipping"`
						JustAddThisItemToYourBag  string `json:"justAddThisItemToYourBag"`
						CrewcutsShipsFree         string `json:"crewcutsShipsFree"`
						ShippingPolicyLink        string `json:"shippingPolicyLink"`
						ReturnsPageLink           string `json:"returnsPageLink"`
						ShopTheLookControls       struct {
							NextItemLabel     string `json:"nextItemLabel"`
							PreviousItemLabel string `json:"previousItemLabel"`
						} `json:"shopTheLookControls"`
						AsSeenIn      string `json:"asSeenIn"`
						SeeTheLooks   string `json:"seeTheLooks"`
						MiniArrayShow string `json:"miniArrayShow"`
						ArrayLinkText string `json:"arrayLinkText"`
						ShipToStore   struct {
							STStext                             string `json:"STStext"`
							FindInStore                         string `json:"findInStore"`
							InStorePickup                       string `json:"inStorePickup"`
							InStorePickupTemporarilyUnavailable string `json:"inStorePickupTemporarilyUnavailable"`
							Near                                string `json:"near"`
							YourLocation                        string `json:"yourLocation"`
							Search                              string `json:"search"`
							Miles                               string `json:"miles"`
							Change                              string `json:"change"`
							Close                               string `json:"close"`
							Hide                                string `json:"hide"`
							Today                               string `json:"today"`
							Tomorrow                            string `json:"tomorrow"`
							ForText                             string `json:"forText"`
							Monogramming                        string `json:"monogramming"`
							FindItem                            string `json:"findItem"`
							InStock                             string `json:"inStock"`
							LimitedQuantity                     string `json:"limitedQuantity"`
							EstimatedPickup                     string `json:"estimatedPickup"`
							SelectSizeMessage                   string `json:"selectSizeMessage"`
							NoNearbyStockMessage                string `json:"noNearbyStockMessage"`
							NoNearbyStoresMessage               string `json:"noNearbyStoresMessage"`
							BadSearchZipcodeMessage             string `json:"badSearchZipcodeMessage"`
							BackorderedItemsMessage1            string `json:"backorderedItemsMessage1"`
							BackorderedItemsMessage2            string `json:"backorderedItemsMessage2"`
							MonogrammingMessage                 string `json:"monogrammingMessage"`
							StoreInfo                           string `json:"storeInfo"`
							YourStore                           string `json:"yourStore"`
							ShowMoreStores                      string `json:"showMoreStores"`
							ShowLessStores                      string `json:"showLessStores"`
							StsFindInStoreDisclaimer            string `json:"stsFindInStoreDisclaimer"`
							StsInStorePickupDisclaimer          string `json:"stsInStorePickupDisclaimer"`
							PlaceholderText                     string `json:"placeholderText"`
							UseCurrentLocation                  string `json:"useCurrentLocation"`
						} `json:"shipToStore"`
						ItemHotnessContent struct {
							View struct {
								Hour string `json:"hour"`
								Day  string `json:"day"`
							} `json:"view"`
							Addtobag struct {
								Hour string `json:"hour"`
								Day  string `json:"day"`
							} `json:"addtobag"`
							Purchase struct {
								Hour string `json:"hour"`
								Day  string `json:"day"`
							} `json:"purchase"`
						} `json:"itemHotnessContent"`
						ModelSelector struct {
							Is                 string `json:"is"`
							SeeOn              string `json:"seeOn"`
							OtherSizes         string `json:"otherSizes"`
							SeeOnOtherSizes    string `json:"seeOnOtherSizes"`
							SeeOnMoreBodyTypes string `json:"seeOnMoreBodyTypes"`
						} `json:"modelSelector"`
						TrueFitMessageHeader  string `json:"trueFitMessageHeader"`
						TrueFitMessageContent string `json:"trueFitMessageContent"`
					} `json:"product"`
					Button struct {
						AddTo            string `json:"addTo"`
						AddedTo          string `json:"addedTo"`
						AddToBag         string `json:"addToBag"`
						ItemAdded        string `json:"itemAdded"`
						Checkout         string `json:"checkout"`
						ClearFilters     string `json:"clearFilters"`
						ClearAllFilters  string `json:"clearAllFilters"`
						HideFilters      string `json:"hideFilters"`
						ShowFilters      string `json:"showFilters"`
						ToggleFilters    string `json:"toggleFilters"`
						CreateNewAccount string `json:"createNewAccount"`
						Done             string `json:"done"`
						NoText           string `json:"noText"`
						NoText2          string `json:"noText2"`
						Preorder         string `json:"preorder"`
						Refine           string `json:"refine"`
						Signin           string `json:"signin"`
						UpdateBag        string `json:"updateBag"`
						Update           string `json:"update"`
						Wishlist         string `json:"wishlist"`
						WishlistFull     string `json:"wishlistFull"`
						WishlistWide     string `json:"wishlistWide"`
						FindInStore      string `json:"findInStore"`
						YesText          string `json:"yesText"`
						YesText2         string `json:"yesText2"`
					} `json:"button"`
					Quickshop struct {
						LinkToPDPtext            string `json:"LinkToPDPtext"`
						SizeFitText              string `json:"sizeFitText"`
						AddToBagConfirmationMsg1 string `json:"addToBagConfirmationMsg1"`
						AddToBagConfirmationMsg2 string `json:"addToBagConfirmationMsg2"`
						ShoppingbagURL           string `json:"shoppingbagUrl"`
						LoadingText              string `json:"loadingText"`
					} `json:"quickshop"`
					Monogram struct {
						Title                        string `json:"title"`
						LettersLabel                 string `json:"lettersLabel"`
						StampLabel                   string `json:"stampLabel"`
						Placement                    string `json:"placement"`
						ThreadColor                  string `json:"threadColor"`
						PriceLabel                   string `json:"priceLabel"`
						PriceValue                   string `json:"priceValue"`
						EditButton                   string `json:"editButton"`
						DeleteButton                 string `json:"deleteButton"`
						ClassicBlockText             string `json:"classicBlockText"`
						DiamondInsigniaText          string `json:"diamondInsigniaText"`
						DisclaimerTitle              string `json:"disclaimerTitle"`
						Disclaimer                   string `json:"disclaimer"`
						AddMonogramButton            string `json:"addMonogramButton"`
						AddEmbossButton              string `json:"addEmbossButton"`
						LinkToPDPtext                string `json:"linkToPDPtext"`
						CallUsText                   string `json:"callUsText"`
						NoteTitle                    string `json:"noteTitle"`
						NoteText                     string `json:"noteText"`
						NoteText2                    string `json:"noteText2"`
						CancelLabel                  string `json:"cancelLabel"`
						AddMonogram                  string `json:"addMonogram"`
						AddEmboss                    string `json:"addEmboss"`
						SizeLabel                    string `json:"sizeLabel"`
						ColorLabel                   string `json:"colorLabel"`
						ItemText                     string `json:"itemText"`
						PlacementLabel               string `json:"placementLabel"`
						SelectStampLabel             string `json:"selectStampLabel"`
						Height                       string `json:"height"`
						SelectLettersLabel           string `json:"selectLettersLabel"`
						SelectLettersText1           string `json:"selectLettersText1"`
						SelectLettersText2           string `json:"selectLettersText2"`
						SelectLettersText3           string `json:"selectLettersText3"`
						SelectLettersText4           string `json:"selectLettersText4"`
						SelectLettersText5           string `json:"selectLettersText5"`
						ThreadColorLabel             string `json:"threadColorLabel"`
						ConfirmSelections            string `json:"confirmSelections"`
						SaveButton                   string `json:"saveButton"`
						MonogramLocationError        string `json:"monogramLocationError"`
						MonogramStyleError           string `json:"monogramStyleError"`
						MonogramClassicBlockError    string `json:"monogramClassicBlockError"`
						MonogramDiamondInsigniaError string `json:"monogramDiamondInsigniaError"`
						MonogramThreadColorError     string `json:"monogramThreadColorError"`
						Questions                    string `json:"questions"`
						CallUs                       string `json:"callUs"`
					} `json:"monogram"`
					Badging struct {
						IB struct {
							Label     string `json:"label"`
							LinkLabel string `json:"linkLabel"`
						} `json:"IB"`
						PB struct {
							Label string `json:"label"`
						} `json:"PB"`
					} `json:"badging"`
					ProductReviews struct {
						FitHeaderText   string `json:"fitHeaderText"`
						WidthHeaderText string `json:"widthHeaderText"`
						RunsSmallText   string `json:"runsSmallText"`
						RunsLargeText   string `json:"runsLargeText"`
						TrueToSizeText  string `json:"trueToSizeText"`
						BasedOnText     string `json:"basedOnText"`
						UserReviewsText string `json:"userReviewsText"`
						ReviewText      string `json:"reviewText"`
						ReviewsText     string `json:"reviewsText"`
					} `json:"productReviews"`
					ProductReviewComments struct {
						ADA struct {
							Expanded        string `json:"expanded"`
							Collapsed       string `json:"collapsed"`
							Review          string `json:"review"`
							Reviews         string `json:"reviews"`
							Star            string `json:"star"`
							Stars           string `json:"stars"`
							AverageRatingOf string `json:"averageRatingOf"`
							View            string `json:"view"`
						} `json:"ADA"`
					} `json:"productReviewComments"`
					ProductRecommendations struct {
						HeaderText       string `json:"headerText"`
						MoreText         string `json:"moreText"`
						AlsoText         string `json:"alsoText"`
						SelectColorsText string `json:"selectColorsText"`
						YourPriceText    string `json:"yourPriceText"`
						QuickShop        string `json:"quickShop"`
					} `json:"productRecommendations"`
				} `json:"p"`
			} `json:"content"`
			GlobalState struct {
				IsBreakpointMediumPlus       bool `json:"isBreakpointMediumPlus"`
				IsBreakpointLarge            bool `json:"isBreakpointLarge"`
				IsBreakpointLargePlus        bool `json:"isBreakpointLargePlus"`
				IsBreakpointXLarge           bool `json:"isBreakpointXLarge"`
				IsBreakpointXXLarge          bool `json:"isBreakpointXXLarge"`
				ShowGlobalOverlay            bool `json:"showGlobalOverlay"`
				IsPastNav                    bool `json:"isPastNav"`
				IsGlobalOverlayTransitioning bool `json:"isGlobalOverlayTransitioning"`
				ScrollHeight                 int  `json:"scrollHeight"`
				WindowWidth                  int  `json:"windowWidth"`
				SitewidePromos               struct {
					CouponCode string        `json:"couponCode"`
					PromoIDs   string        `json:"promoIDs"`
					PromoList  []interface{} `json:"promoList"`
				} `json:"sitewidePromos"`
			} `json:"globalState"`
			WelcomeMat struct {
				Display         bool   `json:"display"`
				ShouldRedirect  bool   `json:"shouldRedirect"`
				RedirectCountry string `json:"redirectCountry"`
			} `json:"welcomeMat"`
			InfoModal struct {
				ShowInfoModal             bool `json:"showInfoModal"`
				UseTermsAndConditionsCopy bool `json:"useTermsAndConditionsCopy"`
				UsePrivacyPolicyCopy      bool `json:"usePrivacyPolicyCopy"`
			} `json:"infoModal"`
			Header struct {
				Promotion struct {
				} `json:"promotion"`
				ShippingPromotion struct {
				} `json:"shippingPromotion"`
				HasSecondaryHeader   bool `json:"hasSecondaryHeader"`
				NavContainerHeight   int  `json:"navContainerHeight"`
				NavigationIsExpanded bool `json:"navigationIsExpanded"`
			} `json:"header"`
			GlobalPromo struct {
				Animation struct {
					ShowGlobalPromo  bool `json:"showGlobalPromo"`
					MaxHeight        int  `json:"maxHeight"`
					SectionHeight    int  `json:"sectionHeight"`
					IsPromoHeightSet bool `json:"isPromoHeightSet"`
				} `json:"animation"`
				Interaction struct {
					IsDetailsOpen     bool `json:"isDetailsOpen"`
					IsShipDetailsOpen bool `json:"isShipDetailsOpen"`
					DetailsOffsetLeft int  `json:"detailsOffsetLeft"`
				} `json:"interaction"`
			} `json:"globalPromo"`
			BagAlert struct {
				IsOpen        bool `json:"isOpen"`
				HasTransition bool `json:"hasTransition"`
				WasClosed     bool `json:"wasClosed"`
			} `json:"bagAlert"`
			Navigation struct {
				Data struct {
					Nav             []interface{} `json:"nav"`
					IsFetching      bool          `json:"isFetching"`
					IsFetchComplete bool          `json:"isFetchComplete"`
					CmsNav          struct {
						PricesAsMarked struct {
							Kids  []interface{} `json:"kids"`
							Men   []interface{} `json:"men"`
							Women []interface{} `json:"women"`
						} `json:"pricesAsMarked"`
						FeatureStories struct {
							Boys []struct {
								URL            string `json:"url"`
								URLTargetBlank bool   `json:"urlTargetBlank"`
								Title          string `json:"title"`
								Image          struct {
									Alt string `json:"alt"`
									Xl  string `json:"xl"`
									Lg  string `json:"lg"`
									Md  string `json:"md"`
									Sm  string `json:"sm"`
								} `json:"image"`
								Ctas []interface{} `json:"ctas"`
							} `json:"boys"`
							Girls []struct {
								URL            string `json:"url"`
								URLTargetBlank bool   `json:"urlTargetBlank"`
								Title          string `json:"title"`
								Image          struct {
									Alt string `json:"alt"`
									Xl  string `json:"xl"`
									Lg  string `json:"lg"`
									Md  string `json:"md"`
									Sm  string `json:"sm"`
								} `json:"image"`
								Ctas []interface{} `json:"ctas"`
							} `json:"girls"`
							Kids []struct {
								URL            string `json:"url"`
								URLTargetBlank bool   `json:"urlTargetBlank"`
								Title          string `json:"title"`
								Image          struct {
									Alt string `json:"alt"`
									Xl  string `json:"xl"`
									Lg  string `json:"lg"`
									Md  string `json:"md"`
									Sm  string `json:"sm"`
								} `json:"image"`
								Ctas []interface{} `json:"ctas"`
							} `json:"kids"`
							Labels []struct {
								URL            string `json:"url"`
								URLTargetBlank bool   `json:"urlTargetBlank"`
								Title          string `json:"title"`
								Image          struct {
									Alt string `json:"alt"`
									Xl  string `json:"xl"`
									Lg  string `json:"lg"`
									Md  string `json:"md"`
									Sm  string `json:"sm"`
								} `json:"image"`
								Ctas []interface{} `json:"ctas"`
							} `json:"labels"`
							Men []struct {
								URL            string `json:"url"`
								URLTargetBlank bool   `json:"urlTargetBlank"`
								Title          string `json:"title"`
								Image          struct {
									Alt string `json:"alt"`
									Xl  string `json:"xl"`
									Lg  string `json:"lg"`
									Md  string `json:"md"`
									Sm  string `json:"sm"`
								} `json:"image"`
								Ctas []interface{} `json:"ctas"`
							} `json:"men"`
							Women []struct {
								URL            string `json:"url"`
								URLTargetBlank bool   `json:"urlTargetBlank"`
								Title          string `json:"title"`
								Image          struct {
									Alt string `json:"alt"`
									Xl  string `json:"xl"`
									Lg  string `json:"lg"`
									Md  string `json:"md"`
									Sm  string `json:"sm"`
								} `json:"image"`
								Ctas []interface{} `json:"ctas"`
							} `json:"women"`
						} `json:"featureStories"`
						GlobalPromo []struct {
							Text           string `json:"text"`
							URL            string `json:"url"`
							URLTargetBlank bool   `json:"urlTargetBlank"`
							Details        string `json:"details"`
							HTMLText       string `json:"htmlText"`
						} `json:"globalPromo"`
						Labels []struct {
							Header string `json:"header"`
							Links  []struct {
								ID             string `json:"_id"`
								URL            string `json:"url"`
								URLTargetBlank bool   `json:"urlTargetBlank"`
								Text           string `json:"text"`
								Navigation     struct {
									Badge string `json:"badge"`
								} `json:"navigation"`
								Label string `json:"label"`
								Badge string `json:"badge"`
							} `json:"links"`
						} `json:"labels"`
						Sale []struct {
							Header string `json:"header"`
							Links  []struct {
								ID             string `json:"_id"`
								Text           string `json:"text"`
								URL            string `json:"url"`
								URLTargetBlank bool   `json:"urlTargetBlank"`
								Label          string `json:"label"`
								Badge          string `json:"badge"`
							} `json:"links"`
						} `json:"sale"`
						Baby []interface{} `json:"baby"`
						Gift []struct {
							Header string `json:"header"`
							Links  []struct {
								ID             string `json:"_id"`
								Text           string `json:"text"`
								URL            string `json:"url"`
								URLTargetBlank bool   `json:"urlTargetBlank"`
								Label          string `json:"label"`
								Badge          string `json:"badge"`
							} `json:"links"`
						} `json:"gift"`
						Cashmere []struct {
							Header string `json:"header"`
							Links  []struct {
								ID             string `json:"_id"`
								Text           string `json:"text"`
								URL            string `json:"url"`
								URLTargetBlank bool   `json:"urlTargetBlank"`
								Label          string `json:"label"`
								Badge          string `json:"badge"`
							} `json:"links"`
						} `json:"cashmere"`
						Swim []struct {
							Header string `json:"header"`
							Links  []struct {
								ID             string `json:"_id"`
								Text           string `json:"text"`
								URL            string `json:"url"`
								URLTargetBlank bool   `json:"urlTargetBlank"`
								Label          string `json:"label"`
								Badge          string `json:"badge"`
							} `json:"links"`
						} `json:"swim"`
						New []struct {
							Header string `json:"header"`
							Links  []struct {
								ID             string `json:"_id"`
								Text           string `json:"text"`
								URL            string `json:"url"`
								URLTargetBlank bool   `json:"urlTargetBlank"`
								Label          string `json:"label"`
								Badge          string `json:"badge"`
							} `json:"links"`
						} `json:"new"`
					} `json:"cmsNav"`
					DeptData struct {
						Women struct {
							FilteredArrays struct {
								New struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"new"`
								RecentlyReduced struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"recentlyReduced"`
								All struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"all"`
								Size struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"size"`
								BrandsWeLove struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"brandsWeLove"`
								BestSeller struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"bestSeller"`
								Sale struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"sale"`
								TopRated struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"topRated"`
								GiftGuide struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"giftGuide"`
								SaleBySize struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"saleBySize"`
								SaleNew struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"saleNew"`
								EasyDoesIt struct {
									URL   string `json:"url"`
									Label string `json:"label"`
								} `json:"easyDoesIt"`
							} `json:"filteredArrays"`
							Shoes []struct {
								URL     string `json:"url"`
								APILink string `json:"apiLink"`
								Label   string `json:"label"`
								ID      string `json:"id,omitempty"`
							} `json:"shoes"`
							Accessories []struct {
								URL     string `json:"url"`
								APILink string `json:"apiLink"`
								Label   string `json:"label"`
								ID      string `json:"id,omitempty"`
							} `json:"accessories"`
							Clothing []struct {
								URL     string `json:"url"`
								APILink string `json:"apiLink"`
								Label   string `json:"label"`
								ID      string `json:"id,omitempty"`
							} `json:"clothing"`
							Link struct {
								IsNonSidecarLink bool   `json:"isNonSidecarLink"`
								Label            string `json:"label"`
								ID               string `json:"id"`
								URL              string `json:"url"`
							} `json:"link"`
						} `json:"women"`
						Boys struct {
							FilteredArrays struct {
								New struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"new"`
								RecentlyReduced struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"recentlyReduced"`
								All struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"all"`
								Size struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"size"`
								BrandsWeLove struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"brandsWeLove"`
								BestSeller struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"bestSeller"`
								Sale struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"sale"`
								TopRated struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"topRated"`
								GiftGuide struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"giftGuide"`
								SaleBySize struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"saleBySize"`
								SaleNew struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"saleNew"`
							} `json:"filteredArrays"`
							Shoes []struct {
								URL     string `json:"url"`
								APILink string `json:"apiLink"`
								Label   string `json:"label"`
								ID      string `json:"id,omitempty"`
							} `json:"shoes"`
							Accessories []struct {
								URL     string `json:"url"`
								APILink string `json:"apiLink"`
								Label   string `json:"label"`
								ID      string `json:"id,omitempty"`
							} `json:"accessories"`
							Clothing []struct {
								URL     string `json:"url"`
								APILink string `json:"apiLink"`
								Label   string `json:"label"`
								ID      string `json:"id,omitempty"`
							} `json:"clothing"`
							Link struct {
								IsNonSidecarLink bool   `json:"isNonSidecarLink"`
								Label            string `json:"label"`
								ID               string `json:"id"`
								URL              string `json:"url"`
							} `json:"link"`
						} `json:"boys"`
						Girls struct {
							FilteredArrays struct {
								New struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"new"`
								RecentlyReduced struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"recentlyReduced"`
								All struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"all"`
								Size struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"size"`
								BrandsWeLove struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"brandsWeLove"`
								BestSeller struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"bestSeller"`
								Sale struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"sale"`
								TopRated struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"topRated"`
								GiftGuide struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"giftGuide"`
								SaleBySize struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"saleBySize"`
								SaleNew struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"saleNew"`
							} `json:"filteredArrays"`
							Shoes []struct {
								URL     string `json:"url"`
								APILink string `json:"apiLink"`
								Label   string `json:"label"`
								ID      string `json:"id,omitempty"`
							} `json:"shoes"`
							Accessories []struct {
								URL     string `json:"url"`
								APILink string `json:"apiLink"`
								Label   string `json:"label"`
								ID      string `json:"id,omitempty"`
							} `json:"accessories"`
							Clothing []struct {
								URL     string `json:"url"`
								APILink string `json:"apiLink"`
								Label   string `json:"label"`
								ID      string `json:"id,omitempty"`
							} `json:"clothing"`
							Link struct {
								IsNonSidecarLink bool   `json:"isNonSidecarLink"`
								Label            string `json:"label"`
								ID               string `json:"id"`
								URL              string `json:"url"`
							} `json:"link"`
						} `json:"girls"`
						Men struct {
							FilteredArrays struct {
								New struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"new"`
								RecentlyReduced struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"recentlyReduced"`
								All struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"all"`
								Size struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"size"`
								BrandsWeLove struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"brandsWeLove"`
								BestSeller struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"bestSeller"`
								Sale struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"sale"`
								TopRated struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"topRated"`
								GiftGuide struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"giftGuide"`
								SaleBySize struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"saleBySize"`
								SaleNew struct {
									URL     string `json:"url"`
									APILink string `json:"apiLink"`
									Label   string `json:"label"`
								} `json:"saleNew"`
								EasyDoesIt struct {
									URL   string `json:"url"`
									Label string `json:"label"`
								} `json:"easyDoesIt"`
							} `json:"filteredArrays"`
							Shoes []struct {
								URL     string `json:"url"`
								APILink string `json:"apiLink"`
								Label   string `json:"label"`
								ID      string `json:"id,omitempty"`
							} `json:"shoes"`
							Accessories []struct {
								URL     string `json:"url"`
								APILink string `json:"apiLink"`
								Label   string `json:"label"`
								ID      string `json:"id,omitempty"`
							} `json:"accessories"`
							Clothing []struct {
								URL     string `json:"url"`
								APILink string `json:"apiLink"`
								Label   string `json:"label"`
								ID      string `json:"id,omitempty"`
							} `json:"clothing"`
							Link struct {
								IsNonSidecarLink bool   `json:"isNonSidecarLink"`
								Label            string `json:"label"`
								ID               string `json:"id"`
								URL              string `json:"url"`
							} `json:"link"`
						} `json:"men"`
						Shoes struct {
							Women []struct {
								URL     string `json:"url"`
								APILink string `json:"apiLink"`
								Label   string `json:"label"`
								ID      string `json:"id,omitempty"`
							} `json:"women"`
							Boys []struct {
								URL     string `json:"url"`
								APILink string `json:"apiLink"`
								Label   string `json:"label"`
								ID      string `json:"id,omitempty"`
							} `json:"boys"`
							Girls []struct {
								URL     string `json:"url"`
								APILink string `json:"apiLink"`
								Label   string `json:"label"`
								ID      string `json:"id,omitempty"`
							} `json:"girls"`
							Men []struct {
								URL     string `json:"url"`
								APILink string `json:"apiLink"`
								Label   string `json:"label"`
								ID      string `json:"id,omitempty"`
							} `json:"men"`
							FilteredArrays struct {
								TheGiftGuide bool `json:"theGiftGuide"`
							} `json:"filteredArrays"`
						} `json:"shoes"`
						Accessories struct {
							Women []struct {
								URL     string `json:"url"`
								APILink string `json:"apiLink"`
								Label   string `json:"label"`
								ID      string `json:"id,omitempty"`
							} `json:"women"`
							Boys []struct {
								URL     string `json:"url"`
								APILink string `json:"apiLink"`
								Label   string `json:"label"`
								ID      string `json:"id,omitempty"`
							} `json:"boys"`
							Girls []struct {
								URL     string `json:"url"`
								APILink string `json:"apiLink"`
								Label   string `json:"label"`
								ID      string `json:"id,omitempty"`
							} `json:"girls"`
							Men []struct {
								URL     string `json:"url"`
								APILink string `json:"apiLink"`
								Label   string `json:"label"`
								ID      string `json:"id,omitempty"`
							} `json:"men"`
							FilteredArrays struct {
								TheGiftGuide bool `json:"theGiftGuide"`
							} `json:"filteredArrays"`
						} `json:"accessories"`
					} `json:"deptData"`
				} `json:"data"`
				HamburgerNav struct {
					IsOpen bool `json:"isOpen"`
					Level  int  `json:"level"`
				} `json:"hamburgerNav"`
			} `json:"navigation"`
			Country struct {
				CountryObj struct {
					Currencies struct {
						CurrencyItem string `json:"currencyItem"`
					} `json:"currencies"`
					CountryCode      string `json:"countryCode"`
					Display          string `json:"display"`
					ChangedCountries bool   `json:"changedCountries"`
				} `json:"countryObj"`
				Regions []struct {
					Code      string `json:"code"`
					Display   string `json:"display"`
					Priority  string `json:"priority"`
					Sort      string `json:"sort"`
					Countries []struct {
						Currencies struct {
							CurrencyItem string `json:"currencyItem"`
						} `json:"currencies"`
						CountryCode      string `json:"countryCode"`
						Display          string `json:"display"`
						LiveChat         string `json:"live_chat,omitempty"`
						ShowFinalSale    string `json:"showFinalSale"`
						ShowFreeShipping string `json:"showFreeShipping"`
						EndecaSaleID     string `json:"endecaSaleId,omitempty"`
					} `json:"countries"`
				} `json:"regions"`
			} `json:"country"`
			Products struct {
				ProductsByProductCode map[string]struct {
					//AB613 struct {
					ProductCode               string        `json:"productCode"`
					ProductDataFetched        bool          `json:"productDataFetched"`
					LastUpdated               int           `json:"lastUpdated"`
					PdpIntlMessage            string        `json:"pdpIntlMessage"`
					IsPreorder                bool          `json:"isPreorder"`
					ShipRestricted            bool          `json:"shipRestricted"`
					IsFindInStore             bool          `json:"isFindInStore"`
					LimitQuantity             interface{}   `json:"limit-quantity"`
					JspURL                    string        `json:"jspUrl"`
					ProductDescriptionRomance string        `json:"productDescriptionRomance"`
					ProductDescriptionFit     []string      `json:"productDescriptionFit"`
					IsVPS                     bool          `json:"isVPS"`
					StyledWithSkus            string        `json:"styledWithSkus"`
					PriceCallArgs             []interface{} `json:"price-call-args"`
					OlapicCopy                string        `json:"olapicCopy"`
					SwatchOrderAlphabetical   bool          `json:"swatchOrderAlphabetical"`
					ColorsMap                 map[string]struct {
						//RD5923 struct {
						Three034 string `json:"30/34"`
						Three532 string `json:"35/32"`
						Three632 string `json:"36/32"`
						Three334 string `json:"33/34"`
						Three234 string `json:"32/34"`
						Three432 string `json:"34/32"`
						Three330 string `json:"33/30"`
						Two930   string `json:"29/30"`
						Three630 string `json:"36/30"`
						Three230 string `json:"32/30"`
						Two932   string `json:"29/32"`
						Three232 string `json:"32/32"`
						Three134 string `json:"31/34"`
						Three332 string `json:"33/32"`
						Three032 string `json:"30/32"`
						Two832   string `json:"28/32"`
						Three430 string `json:"34/30"`
						Three030 string `json:"30/30"`
						Three130 string `json:"31/30"`
						Three132 string `json:"31/32"`
						Three434 string `json:"34/34"`
						//} `json:"RD5923"`
					} `json:"colorsMap"`
					IsFreeShipping  bool   `json:"isFreeShipping"`
					ProductName     string `json:"productName"`
					ProductsByComma string `json:"products-by-comma"`
					Brand           string `json:"brand"`
					PriceEndpoint   string `json:"priceEndpoint"`
					SizeChart       string `json:"sizeChart"`
					SizesMap        struct {
						Three034 struct {
							GR6062 string `json:"GR6062"`
							RD5923 string `json:"RD5923"`
							BR6565 string `json:"BR6565"`
							WZ0451 string `json:"WZ0451"`
							GR8250 string `json:"GR8250"`
						} `json:"30/34"`
						Three532 struct {
							GR6062 string `json:"GR6062"`
							GR8250 string `json:"GR8250"`
							BR6565 string `json:"BR6565"`
							RD5923 string `json:"RD5923"`
							RD6067 string `json:"RD6067"`
							BL7505 string `json:"BL7505"`
							WZ0451 string `json:"WZ0451"`
						} `json:"35/32"`
						Three632 struct {
							BR6565 string `json:"BR6565"`
							GR6062 string `json:"GR6062"`
							RD5923 string `json:"RD5923"`
							BL7505 string `json:"BL7505"`
							GR8250 string `json:"GR8250"`
							WZ0451 string `json:"WZ0451"`
						} `json:"36/32"`
						Three334 struct {
							WZ0451 string `json:"WZ0451"`
							GR8250 string `json:"GR8250"`
							RD5923 string `json:"RD5923"`
							BR6565 string `json:"BR6565"`
							BL7505 string `json:"BL7505"`
							GR6062 string `json:"GR6062"`
						} `json:"33/34"`
						Three234 struct {
							WZ0451 string `json:"WZ0451"`
							BR6565 string `json:"BR6565"`
							GR8250 string `json:"GR8250"`
							RD6067 string `json:"RD6067"`
							GR6062 string `json:"GR6062"`
							RD5923 string `json:"RD5923"`
							BL7505 string `json:"BL7505"`
						} `json:"32/34"`
						Four432 struct {
							GR6062 string `json:"GR6062"`
						} `json:"44/32"`
						Three432 struct {
							BR6565 string `json:"BR6565"`
							WZ0451 string `json:"WZ0451"`
							GR8250 string `json:"GR8250"`
							BL7505 string `json:"BL7505"`
							GR6062 string `json:"GR6062"`
							RD5923 string `json:"RD5923"`
							RD6067 string `json:"RD6067"`
						} `json:"34/32"`
						Three330 struct {
							WZ0451 string `json:"WZ0451"`
							BR6565 string `json:"BR6565"`
							GR6062 string `json:"GR6062"`
							GR8250 string `json:"GR8250"`
							RD5923 string `json:"RD5923"`
							BL7505 string `json:"BL7505"`
						} `json:"33/30"`
						Two930 struct {
							RD5923 string `json:"RD5923"`
							GR8250 string `json:"GR8250"`
						} `json:"29/30"`
						Three630 struct {
							BR6565 string `json:"BR6565"`
							RD5923 string `json:"RD5923"`
							RD6067 string `json:"RD6067"`
						} `json:"36/30"`
						Three230 struct {
							BR6565 string `json:"BR6565"`
							GR6062 string `json:"GR6062"`
							WZ0451 string `json:"WZ0451"`
							RD6067 string `json:"RD6067"`
							RD5923 string `json:"RD5923"`
						} `json:"32/30"`
						Two932 struct {
							RD5923 string `json:"RD5923"`
							GR8250 string `json:"GR8250"`
							GR6062 string `json:"GR6062"`
							WZ0451 string `json:"WZ0451"`
							RD6067 string `json:"RD6067"`
							BL7505 string `json:"BL7505"`
						} `json:"29/32"`
						Three232 struct {
							WZ0451 string `json:"WZ0451"`
							BR6565 string `json:"BR6565"`
							GR8250 string `json:"GR8250"`
							GR6062 string `json:"GR6062"`
							RD6067 string `json:"RD6067"`
							RD5923 string `json:"RD5923"`
						} `json:"32/32"`
						Three134 struct {
							RD5923 string `json:"RD5923"`
							BL7505 string `json:"BL7505"`
							BR6565 string `json:"BR6565"`
							GR8250 string `json:"GR8250"`
							GR6062 string `json:"GR6062"`
							WZ0451 string `json:"WZ0451"`
						} `json:"31/34"`
						Three332 struct {
							WZ0451 string `json:"WZ0451"`
							GR8250 string `json:"GR8250"`
							BR6565 string `json:"BR6565"`
							GR6062 string `json:"GR6062"`
							RD5923 string `json:"RD5923"`
							BL7505 string `json:"BL7505"`
							RD6067 string `json:"RD6067"`
						} `json:"33/32"`
						Three032 struct {
							GR6062 string `json:"GR6062"`
							WZ0451 string `json:"WZ0451"`
							RD5923 string `json:"RD5923"`
							BL7505 string `json:"BL7505"`
							GR8250 string `json:"GR8250"`
						} `json:"30/32"`
						Two832 struct {
							RD5923 string `json:"RD5923"`
							BL7505 string `json:"BL7505"`
							GR8250 string `json:"GR8250"`
							GR6062 string `json:"GR6062"`
							WZ0451 string `json:"WZ0451"`
							RD6067 string `json:"RD6067"`
						} `json:"28/32"`
						Three430 struct {
							BR6565 string `json:"BR6565"`
							WZ0451 string `json:"WZ0451"`
							GR8250 string `json:"GR8250"`
							BL7505 string `json:"BL7505"`
							GR6062 string `json:"GR6062"`
							RD5923 string `json:"RD5923"`
							RD6067 string `json:"RD6067"`
						} `json:"34/30"`
						Three030 struct {
							GR6062 string `json:"GR6062"`
							RD5923 string `json:"RD5923"`
							WZ0451 string `json:"WZ0451"`
						} `json:"30/30"`
						Three634 struct {
							BL7505 string `json:"BL7505"`
							BR6565 string `json:"BR6565"`
							GR8250 string `json:"GR8250"`
							RD6067 string `json:"RD6067"`
						} `json:"36/34"`
						Three130 struct {
							GR6062 string `json:"GR6062"`
							RD5923 string `json:"RD5923"`
							BR6565 string `json:"BR6565"`
							WZ0451 string `json:"WZ0451"`
							RD6067 string `json:"RD6067"`
						} `json:"31/30"`
						Three132 struct {
							BR6565 string `json:"BR6565"`
							GR8250 string `json:"GR8250"`
							GR6062 string `json:"GR6062"`
							WZ0451 string `json:"WZ0451"`
							RD6067 string `json:"RD6067"`
							RD5923 string `json:"RD5923"`
						} `json:"31/32"`
						Four234 struct {
							GR6062 string `json:"GR6062"`
						} `json:"42/34"`
						Three434 struct {
							GR8250 string `json:"GR8250"`
							BR6565 string `json:"BR6565"`
							BL7505 string `json:"BL7505"`
							GR6062 string `json:"GR6062"`
							RD5923 string `json:"RD5923"`
							RD6067 string `json:"RD6067"`
							WZ0451 string `json:"WZ0451"`
						} `json:"34/34"`
					} `json:"sizesMap"`
					ListPrice struct {
						Amount    string `json:"amount"`
						Formatted string `json:"formatted"`
					} `json:"listPrice"`
					CrossSellProduct interface{} `json:"crossSellProduct"`
					URL              string      `json:"url"`
					SizesList        []string    `json:"sizesList"`
					ColorsList       []struct {
						ListPrice struct {
							Amount    string `json:"amount"`
							Formatted string `json:"formatted"`
						} `json:"listPrice"`
						Price struct {
							Amount    string `json:"amount"`
							Formatted string `json:"formatted"`
						} `json:"price"`
						Colors []struct {
							Code string `json:"code"`
							Name string `json:"name"`
						} `json:"colors"`
						DiscountPercentage int `json:"discountPercentage"`
					} `json:"colorsList"`
					ExcludePromo           bool        `json:"excludePromo"`
					Gender                 string      `json:"gender"`
					ProductDescriptionTech []string    `json:"productDescriptionTech"`
					ProductStory           interface{} `json:"productStory"`
					SeoProperties          struct {
						Title               string `json:"title"`
						Description         string `json:"description"`
						CanonicalURL        string `json:"canonicalUrl"`
						SidecarCanonicalURL string `json:"sidecarCanonicalUrl"`
						ParentCategory      string `json:"parentCategory"`
					} `json:"seoProperties"`
					ShotTypes []string `json:"shotTypes"`
					Skus      map[string]struct {
						//Num99105744175 struct {
						ShowOnSale    bool   `json:"show-on-sale"`
						ColorName     string `json:"colorName"`
						Variant       string `json:"variant"`
						IsFinalSale   bool   `json:"isFinalSale"`
						ColorCode     string `json:"colorCode"`
						Size          string `json:"size"`
						Backorderable bool   `json:"backorderable"`
						ListPrice     struct {
							Amount    string `json:"amount"`
							Formatted string `json:"formatted"`
						} `json:"listPrice"`
						SkuID      string `json:"skuId"`
						ShowOnFull bool   `json:"show-on-full"`
						Price      struct {
							Amount    string `json:"amount"`
							Formatted string `json:"formatted"`
						} `json:"price"`
						//} `json:"99105744175"`

					} `json:"skus"`
					OlapicEnabled         bool          `json:"olapicEnabled"`
					BaseProductCode       string        `json:"baseProductCode"`
					RelatedProducts       interface{}   `json:"relatedProducts"`
					ProductDescriptionSEO []string      `json:"productDescriptionSEO"`
					DefaultColorCode      string        `json:"defaultColorCode"`
					BaseProductColorCode  string        `json:"baseProductColorCode"`
					FeaturedInProducts    interface{}   `json:"featuredInProducts"`
					EnabledBadges         string        `json:"enabledBadges"`
					ContactPhone          string        `json:"contactPhone"`
					CountryCode           string        `json:"countryCode"`
					PromoText             interface{}   `json:"promoText"`
					Sale                  string        `json:"sale"`
					IsFromSale            bool          `json:"isFromSale"`
					ColorName             string        `json:"color_name"`
					Merged                string        `json:"merged"`
					Category              string        `json:"category"`
					Subcategory           string        `json:"subcategory"`
					ProductSlug           string        `json:"productSlug"`
					IsSaleProduct         bool          `json:"isSaleProduct"`
					ColorName             string        `json:"colorName"`
					SelectedProductCode   string        `json:"selectedProductCode"`
					SelectedColorName     string        `json:"selectedColorName"`
					SelectedColorCode     string        `json:"selectedColorCode"`
					SelectedQuantity      int           `json:"selectedQuantity"`
					SelectedSize          string        `json:"selectedSize"`
					FitModelDetails       []interface{} `json:"fitModelDetails"`
					UseProductAPI         bool          `json:"useProductApi"`
					IsStoresShowAll       bool          `json:"isStoresShowAll"`
					PriceModel            map[string]struct {
						//AB613 struct {
						ListPrice struct {
							Amount    int    `json:"amount"`
							Formatted string `json:"formatted"`
						} `json:"listPrice"`
						Colors []struct {
							SalePrice struct {
								Amount    int    `json:"amount"`
								Formatted string `json:"formatted"`
							} `json:"salePrice"`
							ShowOnSale       bool   `json:"show-on-sale"`
							ColorName        string `json:"colorName"`
							DefaultColorCode string `json:"default-color-code"`
							ProductDesc      string `json:"productDesc"`
							ShotType         string `json:"shot-type"`
							ColorCode        string `json:"colorCode"`
							ProductID        string `json:"productId"`
							SaleFlag         string `json:"sale-flag"`
							ExtendedSize     string `json:"extended-size"`
							HasOnfig         bool   `json:"hasOnfig"`
							ShowOnFull       bool   `json:"show-on-full"`
							SkuShotType      string `json:"skuShotType"`
							Variations       string `json:"variations"`
						} `json:"colors"`
						Was struct {
							Amount    int    `json:"amount"`
							Formatted string `json:"formatted"`
						} `json:"was"`
						Now struct {
							Amount    int    `json:"amount"`
							Formatted string `json:"formatted"`
						} `json:"now"`
						DiscountPercentage int    `json:"discountPercentage"`
						ExtendedSize       string `json:"extendedSize"`
						Badge              struct {
							Label    string `json:"label"`
							Type     string `json:"type"`
							Priority int    `json:"priority"`
							IconPath string `json:"iconPath"`
						} `json:"badge"`
						//} `json:"AB613"`
					} `json:"priceModel"`
					HasVariations bool `json:"hasVariations"`
					//} `json:"AB613"`
				} `json:"productsByProductCode"`
				IsFetching     bool `json:"isFetching"`
				PdpDynamicData struct {
					IsFetching      bool          `json:"isFetching"`
					IsFetchComplete bool          `json:"isFetchComplete"`
					JSON            []interface{} `json:"json"`
				} `json:"pdpDynamicData"`
				LastUpdated time.Time `json:"lastUpdated"`
				Helpers     struct {
					Tooltip struct {
						Tooltip struct {
							ClassName string `json:"className"`
							Message   string `json:"message"`
							Top       int    `json:"top"`
							Left      int    `json:"left"`
						} `json:"tooltip"`
						ShouldShowTooltip bool `json:"shouldShowTooltip"`
					} `json:"tooltip"`
					Message struct {
						AddToBagText               string `json:"addToBagText"`
						WishlistText               string `json:"wishlistText"`
						ShowSkuMessage             bool   `json:"showSkuMessage"`
						ShowOtherMessage           bool   `json:"showOtherMessage"`
						IsBagDisabled              bool   `json:"isBagDisabled"`
						IsWishlistDisabled         bool   `json:"isWishlistDisabled"`
						AddToBagClasses            string `json:"addToBagClasses"`
						WishlistClasses            string `json:"wishlistClasses"`
						ProductDetailsHeight       string `json:"productDetailsHeight"`
						ShowMonogram               bool   `json:"showMonogram"`
						HiddenMessage              int    `json:"hiddenMessage"`
						ShowMessage                int    `json:"showMessage"`
						ShowLowInventory           bool   `json:"showLowInventory"`
						ShowFinalSale              bool   `json:"showFinalSale"`
						ShowBackordered            bool   `json:"showBackordered"`
						ShowBasketQuotaExceeded    bool   `json:"showBasketQuotaExceeded"`
						ShowGenericAddToBag        bool   `json:"showGenericAddToBag"`
						BackorderedDate            string `json:"backorderedDate"`
						ShowNoMonogramFromRetail   bool   `json:"showNoMonogramFromRetail"`
						ShowForgotSize             bool   `json:"showForgotSize"`
						ShowAddedToBag             bool   `json:"showAddedToBag"`
						ShowAddedToBagOffline      bool   `json:"showAddedToBagOffline"`
						CheckoutFeatureParams      string `json:"checkoutFeatureParams"`
						ShowAddedToWishlist        bool   `json:"showAddedToWishlist"`
						ShowAddedToWishlistOffline bool   `json:"showAddedToWishlistOffline"`
						AddToWishlistMessage1      string `json:"addToWishlistMessage1"`
						AddToWishlistMessage2      string `json:"addToWishlistMessage2"`
						AddToWishlistMessage3      string `json:"addToWishlistMessage3"`
						WishlistSize               string `json:"wishlistSize"`
						Response                   struct {
						} `json:"response"`
						ShowPostPartialQuantity bool   `json:"showPostPartialQuantity"`
						ShowPostSoldOut         bool   `json:"showPostSoldOut"`
						ShowPostBackordered     bool   `json:"showPostBackordered"`
						UpdateWishlistMessage   bool   `json:"updateWishlistMessage"`
						ToastMessage            string `json:"toastMessage"`
					} `json:"message"`
					Carousel struct {
						CarouselSliderWrap struct {
							Width     string `json:"width"`
							Height    string `json:"height"`
							Transform string `json:"transform"`
						} `json:"carouselSliderWrap"`
						ImageWidth         string `json:"imageWidth"`
						SelectedPage       int    `json:"selectedPage"`
						CurrentPage        int    `json:"currentPage"`
						InitialCarouselSet bool   `json:"initialCarouselSet"`
						HasTransition      bool   `json:"hasTransition"`
					} `json:"carousel"`
					FlyoutZoom struct {
						FlyoutZoomWrapper struct {
							Left   int `json:"left"`
							Width  int `json:"width"`
							Height int `json:"height"`
						} `json:"flyoutZoomWrapper"`
						ImageZoomWrapper struct {
							Height    int    `json:"height"`
							Width     int    `json:"width"`
							Transform string `json:"transform"`
						} `json:"imageZoomWrapper"`
						FlyoutZoomLens struct {
							Top       int    `json:"top"`
							Left      int    `json:"left"`
							Width     int    `json:"width"`
							Height    int    `json:"height"`
							Transform string `json:"transform"`
						} `json:"flyoutZoomLens"`
						ReferencePoints struct {
						} `json:"referencePoints"`
						ReferencePointsSet bool `json:"referencePointsSet"`
						HasZoomStarted     bool `json:"hasZoomStarted"`
						FlyoutZoomSet      bool `json:"flyoutZoomSet"`
					} `json:"flyoutZoom"`
				} `json:"helpers"`
			} `json:"products"`
			ProductDetail struct {
				EVar78Pdp string `json:"eVar78pdp"`
				Flyzoom   struct {
					ImageIndex   int           `json:"imageIndex"`
					ImageURL     string        `json:"imageURL"`
					ImagePresets []interface{} `json:"imagePresets"`
					IsReady      bool          `json:"isReady"`
					Image        struct {
						Width  int `json:"width"`
						Height int `json:"height"`
						X      int `json:"x"`
						Y      int `json:"y"`
					} `json:"image"`
					Wrapper struct {
						Width  int `json:"width"`
						Height int `json:"height"`
						X      int `json:"x"`
						Y      int `json:"y"`
					} `json:"wrapper"`
				} `json:"flyzoom"`
				Hover struct {
					HoverColorCode   interface{} `json:"hoverColorCode"`
					HoverSize        interface{} `json:"hoverSize"`
					HoverVariant     interface{} `json:"hoverVariant"`
					HoverProductCode interface{} `json:"hoverProductCode"`
				} `json:"hover"`
				Personalize struct {
					Customization struct {
						IsSaved   bool        `json:"isSaved"`
						Type      string      `json:"type"`
						Placement interface{} `json:"placement"`
						Style     interface{} `json:"style"`
						Initials  []string    `json:"initials"`
						Color     interface{} `json:"color"`
					} `json:"customization"`
					Tooltip struct {
						IsOpen bool   `json:"isOpen"`
						Type   string `json:"type"`
					} `json:"tooltip"`
					Modal struct {
						IsOpen     bool          `json:"isOpen"`
						Type       string        `json:"type"`
						Errors     []interface{} `json:"errors"`
						HoverColor interface{}   `json:"hoverColor"`
					} `json:"modal"`
					TempCopy interface{} `json:"tempCopy"`
				} `json:"personalize"`
				ViewingProduct struct {
					ColorsVisited struct {
					} `json:"colorsVisited"`
					ThumbnailTray struct {
					} `json:"thumbnailTray"`
					IsTrueFitVisible    bool   `json:"isTrueFitVisible"`
					ShowFamilyPrice     bool   `json:"showFamilyPrice"`
					PdpLoaded           bool   `json:"pdpLoaded"`
					PdpPartiallyLoaded  bool   `json:"pdpPartiallyLoaded"`
					SelectedColorCode   string `json:"selectedColorCode"`
					SelectedColorName   string `json:"selectedColorName"`
					ProductCode         string `json:"productCode"`
					SelectedProductCode string `json:"selectedProductCode"`
					SelectedSize        string `json:"selectedSize"`
				} `json:"viewingProduct"`
				ShipToStore struct {
					Modal struct {
						Show bool `json:"show"`
					} `json:"modal"`
				} `json:"shipToStore"`
				Hotness struct {
					Display struct {
						Show bool `json:"show"`
					} `json:"display"`
				} `json:"hotness"`
				MergedProduct struct {
					VisitedMergedProducts struct {
					} `json:"visitedMergedProducts"`
				} `json:"mergedProduct"`
			} `json:"productDetail"`
			Quickshop struct {
				ProductObj struct {
					ColorsVisited struct {
					} `json:"colorsVisited"`
					ShowSizeFit              bool        `json:"showSizeFit"`
					ShowDetails              bool        `json:"showDetails"`
					IsSizeSet                bool        `json:"isSizeSet"`
					LoadingText              string      `json:"loadingText"`
					SelectedColorCode        string      `json:"selectedColorCode"`
					SelectedColorName        string      `json:"selectedColorName"`
					CustomStlRecommendations interface{} `json:"customStlRecommendations"`
					ShowFamilyPrice          bool        `json:"showFamilyPrice"`
					SelectedShot             int         `json:"selectedShot"`
				} `json:"productObj"`
				Hover struct {
					HoverColorCode   interface{} `json:"hoverColorCode"`
					HoverSize        interface{} `json:"hoverSize"`
					HoverVariant     interface{} `json:"hoverVariant"`
					HoverProductCode interface{} `json:"hoverProductCode"`
				} `json:"hover"`
				SetQuickShop struct {
					Show bool   `json:"show"`
					Top  string `json:"top"`
				} `json:"setQuickShop"`
				MergedProduct struct {
					VisitedMergedProducts struct {
					} `json:"visitedMergedProducts"`
				} `json:"mergedProduct"`
			} `json:"quickshop"`
			Configuration struct {
				IsCheckoutStarted                 bool   `json:"isCheckoutStarted"`
				IsExperimental                    bool   `json:"isExperimental"`
				Brand                             string `json:"brand"`
				OlapicAPI                         string `json:"olapicAPI"`
				BvKey                             string `json:"bvKey"`
				BvProductAPIKey                   string `json:"bvProductApiKey"`
				OlapicKey                         string `json:"olapicKey"`
				TrueFitEnvironment                string `json:"trueFitEnvironment"`
				TrueFitClientID                   string `json:"trueFitClientID"`
				BvSubmissionURL                   string `json:"bvSubmissionURL"`
				OlapicInstance                    string `json:"olapicInstance"`
				BvPathAPI                         string `json:"bvPathAPI"`
				BvReviewsAPI                      string `json:"bvReviewsAPI"`
				BvImageURL                        string `json:"bvImageURL"`
				AmazonS3URL                       string `json:"amazonS3URL"`
				DynamicYieldID                    string `json:"dynamicYieldId"`
				NewCheckoutURL                    string `json:"newCheckoutUrl"`
				SearchResultsPerPage              int    `json:"searchResultsPerPage"`
				ArrayResultsPerPage               int    `json:"arrayResultsPerPage"`
				ArrayViewAllResultsPerPage        int    `json:"arrayViewAllResultsPerPage"`
				ArrayViewAllPageIndex             int    `json:"arrayViewAllPageIndex"`
				CountryCookie                     string `json:"countryCookie"`
				GoogleTagManager                  string `json:"googleTagManager"`
				ImageServerURL                    string `json:"imageServerURL"`
				VideoServerURL                    string `json:"videoServerURL"`
				TypeKitID                         string `json:"typeKitId"`
				TypeKitID2018                     string `json:"typeKitId2018"`
				SiteVerificationContent           string `json:"siteVerificationContent"`
				AjaxTimeout                       string `json:"ajaxTimeout"`
				TypeKitURL                        string `json:"typeKitURL"`
				SaleLabel                         string `json:"saleLabel"`
				GoogleMapsKey                     string `json:"googleMapsKey"`
				GoogleRecaptchaV2Key              string `json:"googleRecaptchaV2Key"`
				GoogleRecaptchaV2SigninKey        string `json:"googleRecaptchaV2SigninKey"`
				GoogleRecaptchaV2EmailContactKey  string `json:"googleRecaptchaV2EmailContactKey"`
				AssociateSessionExpiryWarningMins int    `json:"associateSessionExpiryWarningMins"`
				LowStockBadgeSkuThresholdPercent  int    `json:"lowStockBadgeSkuThresholdPercent"`
				LowStockBadgeSkuInvThreshold      int    `json:"lowStockBadgeSkuInvThreshold"`
				MaxVisitsAnalyzed                 int    `json:"maxVisitsAnalyzed"`
				MaxCategoriesStored               int    `json:"maxCategoriesStored"`
				EnvHostName                       string `json:"envHostName"`
				SidecarCMSHost                    string `json:"sidecarCMSHost"`
				ResponsysEmailPreferencesURL      string `json:"responsysEmailPreferencesURL"`
				AfterpayThresholds                struct {
					Minimum int `json:"minimum"`
					Maximum int `json:"maximum"`
				} `json:"afterpayThresholds"`
				LiveChatAvailabilityThreshold     int `json:"liveChatAvailabilityThreshold"`
				HomepageRecommendationsThresholds struct {
					Women int `json:"women"`
					Men   int `json:"men"`
				} `json:"homepageRecommendationsThresholds"`
				HomepageGenderThresholds struct {
					Womens   int `json:"womens"`
					Mens     int `json:"mens"`
					CrewCuts int `json:"crewCuts"`
				} `json:"homepageGenderThresholds"`
				Image struct {
					PRESETCAT           string `json:"PRESET_CAT"`
					PRESETPDPXSMALL     string `json:"PRESET_PDP_XSMALL"`
					PRESETPDPSMALL      string `json:"PRESET_PDP_SMALL"`
					PRESETPDPSMALLCROP  string `json:"PRESET_PDP_SMALL_CROP"`
					PRESETPDPMEDIUM     string `json:"PRESET_PDP_MEDIUM"`
					PRESETPDPMEDIUMCROP string `json:"PRESET_PDP_MEDIUM_CROP"`
					PRESETPDPLARGE      string `json:"PRESET_PDP_LARGE"`
					PRESETPDPXLARGE     string `json:"PRESET_PDP_XLARGE"`
					PRESETPDPENLARGE    string `json:"PRESET_PDP_ENLARGE"`
					PRESETSWATCH        string `json:"PRESET_SWATCH"`
				} `json:"image"`
				Hotness struct {
					Ranking    []string `json:"ranking"`
					Thresholds struct {
						View struct {
							Hour struct {
								Women int `json:"women"`
								Men   int `json:"men"`
								Kids  int `json:"kids"`
							} `json:"hour"`
							Day struct {
								Women int `json:"women"`
								Men   int `json:"men"`
								Kids  int `json:"kids"`
							} `json:"day"`
						} `json:"view"`
						Addtobag struct {
							Hour struct {
								Women int `json:"women"`
								Men   int `json:"men"`
								Kids  int `json:"kids"`
							} `json:"hour"`
							Day struct {
								Women int `json:"women"`
								Men   int `json:"men"`
								Kids  int `json:"kids"`
							} `json:"day"`
						} `json:"addtobag"`
						Purchase struct {
							Hour struct {
								Women int `json:"women"`
								Men   int `json:"men"`
								Kids  int `json:"kids"`
							} `json:"hour"`
							Day struct {
								Women int `json:"women"`
								Men   int `json:"men"`
								Kids  int `json:"kids"`
							} `json:"day"`
						} `json:"purchase"`
					} `json:"thresholds"`
				} `json:"hotness"`
				ControlTitleSEO                        bool   `json:"controlTitleSEO"`
				DisableCrawlURLWithHyphen              bool   `json:"disableCrawlURLWithHyphen"`
				EnableAMP                              bool   `json:"enableAMP"`
				EnableAsyncOmniture                    bool   `json:"enableAsyncOmniture"`
				EnableHomepageAMP                      bool   `json:"enableHomepageAMP"`
				EnableBaynote                          bool   `json:"enableBaynote"`
				EnableBaynoteTracking                  bool   `json:"enableBaynoteTracking"`
				EnableBaynoteRecentlyViewed            bool   `json:"enableBaynoteRecentlyViewed"`
				EnableBaynoteProductRecs               bool   `json:"enableBaynoteProductRecs"`
				EnableBaynoteProductRecsPDP            bool   `json:"enableBaynoteProductRecsPDP"`
				EnableCertona                          bool   `json:"enableCertona"`
				EnableCertonaProductRecs               bool   `json:"enableCertonaProductRecs"`
				EnableCertonaRecentlyViewed            bool   `json:"enableCertonaRecentlyViewed"`
				EnableCertonaProductRecsPDP            bool   `json:"enableCertonaProductRecsPDP"`
				EnableCertonaProductRecsOrderHistory   bool   `json:"enableCertonaProductRecsOrderHistory"`
				EnableCertonaPDPTopPosition            bool   `json:"enableCertonaPDPTopPosition"`
				ShowDefaultRecsOffBrand                bool   `json:"showDefaultRecsOffBrand"`
				EnableFindInStore                      bool   `json:"enableFindInStore"`
				EnableFlyoutSaleStyleLinks             bool   `json:"enableFlyoutSaleStyleLinks"`
				EnableGoogleTagManager                 bool   `json:"enableGoogleTagManager"`
				EnableLoyaltyMessage                   bool   `json:"enableLoyaltyMessage"`
				EnableMiniArray                        bool   `json:"enableMiniArray"`
				EnableMonetate                         bool   `json:"enableMonetate"`
				EnableOlapic                           bool   `json:"enableOlapic"`
				EnablePreconnectResources              bool   `json:"enablePreconnectResources"`
				EnableReCaptcha                        bool   `json:"enableReCaptcha"`
				EnableSEO                              bool   `json:"enableSEO"`
				EnableSeoWordClouds                    bool   `json:"enableSeoWordClouds"`
				EnableUPSContextChooser                bool   `json:"enableUPSContextChooser"`
				EnableUSLightboxPromo                  bool   `json:"enableUSLightboxPromo"`
				EnableTeacherStudent                   bool   `json:"enableTeacherStudent"`
				EnableTrueFit                          bool   `json:"enableTrueFit"`
				ShowTrueFitAtTop                       bool   `json:"showTrueFitAtTop"`
				EnableVps                              bool   `json:"enableVps"`
				IsCyclePromo                           bool   `json:"isCyclePromo"`
				IsProductCustomerPhotosEnabled         bool   `json:"isProductCustomerPhotosEnabled"`
				NewSaleSection                         bool   `json:"newSaleSection"`
				UseSigninRegisterModal                 bool   `json:"useSigninRegisterModal"`
				UseLegacyLogin                         bool   `json:"useLegacyLogin"`
				ShowForcePasswordUpdateModal           bool   `json:"showForcePasswordUpdateModal"`
				LockUserOnForcePasswordUpdate          bool   `json:"lockUserOnForcePasswordUpdate"`
				ShowOrderHistoryRecs                   bool   `json:"showOrderHistoryRecs"`
				EnableAssociateSignin                  bool   `json:"enableAssociateSignin"`
				CanAssociateViewOrderHistory           bool   `json:"canAssociateViewOrderHistory"`
				CanAssociateViewWishlist               bool   `json:"canAssociateViewWishlist"`
				CanAssociateViewRewardsStatus          bool   `json:"canAssociateViewRewardsStatus"`
				ShowAccountDropdown                    bool   `json:"showAccountDropdown"`
				ShowAfterpay                           bool   `json:"showAfterpay"`
				ShowAlternativeFeatureStory            bool   `json:"showAlternativeFeatureStory"`
				ShowBadging                            bool   `json:"showBadging"`
				ShowBagAlerts                          bool   `json:"showBagAlerts"`
				ShowBazaarVoiceSpotlight               bool   `json:"showBazaarVoiceSpotlight"`
				ShowCanadaWelcomemat                   bool   `json:"showCanadaWelcomemat"`
				ShowCategoryFiltersViewText            bool   `json:"showCategoryFiltersViewText"`
				ShowCategoryHeader                     bool   `json:"showCategoryHeader"`
				ShowCategoryItemCount                  bool   `json:"showCategoryItemCount"`
				ShowContextChooser                     bool   `json:"showContextChooser"`
				ShowDiscountFilterOnCategory           bool   `json:"showDiscountFilterOnCategory"`
				ShowDiscountFilterOnNewSearchSale      bool   `json:"showDiscountFilterOnNewSearchSale"`
				ShowDiscountPercentage                 bool   `json:"showDiscountPercentage"`
				ShowModelSelector                      bool   `json:"showModelSelector"`
				ShowModelSelectorEnhanced              bool   `json:"showModelSelectorEnhanced"`
				ShowEmailCapture                       bool   `json:"showEmailCapture"`
				ShowFactoryLink                        bool   `json:"showFactoryLink"`
				ShowFooterSafetyRecall                 bool   `json:"showFooterSafetyRecall"`
				ShowFooterSeoPromoLink                 bool   `json:"showFooterSeoPromoLink"`
				ShowFooterSizecharts                   bool   `json:"showFooterSizecharts"`
				ShowFooterHolidayPromoLinks            bool   `json:"showFooterHolidayPromoLinks"`
				ShowHeaderNavBadging                   bool   `json:"showHeaderNavBadging"`
				ShowIntlFaq                            bool   `json:"showIntlFaq"`
				ShowLoyaltyCreateAccountLightbox       bool   `json:"showLoyaltyCreateAccountLightbox"`
				ShowJccc2XPointsMsg                    bool   `json:"showJccc2xPointsMsg"`
				ShowMiraklMarketplace                  bool   `json:"showMiraklMarketplace"`
				ShowMiniBagPriceStrikethrough          bool   `json:"showMiniBagPriceStrikethrough"`
				ShowMonogram                           bool   `json:"showMonogram"`
				ShowNewCategoryFilters                 bool   `json:"showNewCategoryFilters"`
				ShowOlapicCopy                         bool   `json:"showOlapicCopy"`
				ShowGamificationProgressBar            bool   `json:"showGamificationProgressBar"`
				ShowRecentlyViewed                     bool   `json:"showRecentlyViewed"`
				ShowRegistration                       bool   `json:"showRegistration"`
				ShowReviewsMobileAccordion             bool   `json:"showReviewsMobileAccordion"`
				ShowReviewsSliderFit                   bool   `json:"showReviewsSliderFit"`
				ShowSearchSaleBadging                  bool   `json:"showSearchSaleBadging"`
				ShowSocialButtons                      bool   `json:"showSocialButtons"`
				ShowStyledWith                         bool   `json:"showStyledWith"`
				ShowStyledWithOnArrayPages             bool   `json:"showStyledWithOnArrayPages"`
				ShowWelcomeMat                         bool   `json:"showWelcomeMat"`
				UseRedirectJSP                         bool   `json:"useRedirectJSP"`
				UseSidecarCanonicalURL                 bool   `json:"useSidecarCanonicalURL"`
				UseStyledEmailCaptureHeader            bool   `json:"useStyledEmailCaptureHeader"`
				UseReactHeader                         bool   `json:"useReactHeader"`
				UseTermsAndConditionsModal             bool   `json:"useTermsAndConditionsModal"`
				UsePrivacyPolicyModal                  bool   `json:"usePrivacyPolicyModal"`
				EnableSimplifiedHeaderAndFooter        bool   `json:"enableSimplifiedHeaderAndFooter"`
				ShowExpandedFiltersByDefault           bool   `json:"showExpandedFiltersByDefault"`
				EnableMPulseTest                       bool   `json:"enableMPulseTest"`
				DisplayReactAsSeenIn                   bool   `json:"displayReactAsSeenIn"`
				ShowPDPThumbnailCarousel               bool   `json:"showPDPThumbnailCarousel"`
				ShowProductStory                       bool   `json:"showProductStory"`
				UseMultiSelectCategory                 bool   `json:"useMultiSelectCategory"`
				DeferImageLoad                         bool   `json:"deferImageLoad"`
				ShowFilterColorSwatches                bool   `json:"showFilterColorSwatches"`
				ShowMultipleProductRecommendations     bool   `json:"showMultipleProductRecommendations"`
				UseTestProductRecommendationsEndPoint  bool   `json:"useTestProductRecommendationsEndPoint"`
				IsLoyaltyEnabled                       bool   `json:"isLoyaltyEnabled"`
				UseNewTypekit                          bool   `json:"useNewTypekit"`
				InjectContentRowIntoArray              bool   `json:"injectContentRowIntoArray"`
				ShowFindInStoreFilter                  bool   `json:"showFindInStoreFilter"`
				ShowFindInStoreToggle                  bool   `json:"showFindInStoreToggle"`
				ShowNewForYouFilter                    bool   `json:"showNewForYouFilter"`
				ShowItemHotness                        bool   `json:"showItemHotness"`
				ItemHotnessURL                         string `json:"itemHotnessURL"`
				TrackItemHotness                       bool   `json:"trackItemHotness"`
				ShowFooterRewardsLinks                 bool   `json:"showFooterRewardsLinks"`
				UseNewFooter                           bool   `json:"useNewFooter"`
				UseNewCrewNav                          bool   `json:"useNewCrewNav"`
				UseNewCheckout                         bool   `json:"useNewCheckout"`
				UseGraphQLaddToBag                     bool   `json:"useGraphQLaddToBag"`
				EnableArrayVideos                      bool   `json:"enableArrayVideos"`
				EnableEnhancedEiec                     bool   `json:"enableEnhancedEiec"`
				EnableFitGuide                         bool   `json:"enableFitGuide"`
				ShowArrayLink                          bool   `json:"showArrayLink"`
				EnableAkamaiRequestHeaderOverride      bool   `json:"enableAkamaiRequestHeaderOverride"`
				EnableNavShowAllPromos                 bool   `json:"enableNavShowAllPromos"`
				EnableTrueFitTracking                  bool   `json:"enableTrueFitTracking"`
				ShowBabyInNav                          bool   `json:"showBabyInNav"`
				ShowCashmereInNav                      bool   `json:"showCashmereInNav"`
				ShowGiftGuideInNav                     bool   `json:"showGiftGuideInNav"`
				ShowGiftsInGenderFlyout                bool   `json:"showGiftsInGenderFlyout"`
				ShowSwimInNav                          bool   `json:"showSwimInNav"`
				ShowBrandsWeLoveInNav                  bool   `json:"showBrandsWeLoveInNav"`
				ShowShoesInNav                         bool   `json:"showShoesInNav"`
				ShowAccessoriesInNav                   bool   `json:"showAccessoriesInNav"`
				GetSwimCashmereGiftNavLinksFromCms     bool   `json:"getSwimCashmereGiftNavLinksFromCms"`
				GetNewArrivalsFromCms                  bool   `json:"getNewArrivalsFromCms"`
				EnableArraySignposts                   bool   `json:"enableArraySignposts"`
				EnableDenimWashFilter                  bool   `json:"enableDenimWashFilter"`
				ShowQuickShopStlBar                    bool   `json:"showQuickShopStlBar"`
				ShowBrandsWeLoveContentTiles           bool   `json:"showBrandsWeLoveContentTiles"`
				EnableArrayImageLazyLoad               bool   `json:"enableArrayImageLazyLoad"`
				EnableSPAForNext                       bool   `json:"enableSPAForNext"`
				UseDataKiboReviews                     bool   `json:"useDataKiboReviews"`
				ShowBadgeRecommendations               bool   `json:"showBadgeRecommendations"`
				ShowItemCountAboveArray                bool   `json:"showItemCountAboveArray"`
				ShowNewFactoryPriceStyling             bool   `json:"showNewFactoryPriceStyling"`
				UseNewPagination                       bool   `json:"useNewPagination"`
				IsResetPasswordLinkEnabled             bool   `json:"isResetPasswordLinkEnabled"`
				ShowFindInStoreOnPDP                   bool   `json:"showFindInStoreOnPDP"`
				Use2019MobileFilters                   bool   `json:"use2019MobileFilters"`
				UseDisabledRefinementValues            bool   `json:"useDisabledRefinementValues"`
				ShowUpdatedRewardsModule               bool   `json:"showUpdatedRewardsModule"`
				ShowLiveChatInFooter                   bool   `json:"showLiveChatInFooter"`
				UseTtecLiveChat                        bool   `json:"useTtecLiveChat"`
				DisableFeatureQuickshop                bool   `json:"disableFeatureQuickshop"`
				GetBVProductReviewViaBrowser           bool   `json:"getBVProductReviewViaBrowser"`
				BurySwimFlag                           bool   `json:"burySwimFlag"`
				Show199PointsInfo                      bool   `json:"show199PointsInfo"`
				EnableShopAllArray                     bool   `json:"enableShopAllArray"`
				UseReactPDP                            bool   `json:"useReactPDP"`
				EnableReactErrorHandler                bool   `json:"enableReactErrorHandler"`
				UseReactRouterLink                     bool   `json:"useReactRouterLink"`
				SetAccountGiftProperties               bool   `json:"setAccountGiftProperties"`
				UseNewCategoryService                  bool   `json:"useNewCategoryService"`
				HideTopNav                             bool   `json:"hideTopNav"`
				EnableDynamicYield                     bool   `json:"enableDynamicYield"`
				EnableAccountDirectLink                bool   `json:"enableAccountDirectLink"`
				ShowCategoryHeaderImage                bool   `json:"showCategoryHeaderImage"`
				SaleLandingImage                       bool   `json:"saleLandingImage"`
				ShowAllColorsOnFigure                  bool   `json:"showAllColorsOnFigure"`
				UseNewSearchSaleService                bool   `json:"useNewSearchSaleService"`
				ShowMobileLoyaltyCreateAccountLightbox bool   `json:"showMobileLoyaltyCreateAccountLightbox"`
				UseDataGeoEndpoint                     bool   `json:"useDataGeoEndpoint"`
				ShowLaydownImagesOnArray               bool   `json:"showLaydownImagesOnArray"`
				ToggleArrayImageOnHover                bool   `json:"toggleArrayImageOnHover"`
				EnableMobileVideo                      bool   `json:"enableMobileVideo"`
				DisablePDPRecLazyLoad                  bool   `json:"disablePDPRecLazyLoad"`
				ShowStoreOrderHistory                  bool   `json:"showStoreOrderHistory"`
				UseSearchBoostRule                     bool   `json:"useSearchBoostRule"`
				UseDefaultSearchInterface              bool   `json:"useDefaultSearchInterface"`
				EnableFamilyProducts                   bool   `json:"enableFamilyProducts"`
				EnableTrueFitMessage                   bool   `json:"enableTrueFitMessage"`
				ShowFamilyArrayTiles                   bool   `json:"showFamilyArrayTiles"`
				EnableGlobalTopPromo                   bool   `json:"enableGlobalTopPromo"`
				EnableGlobalTopPromoShowcase           bool   `json:"enableGlobalTopPromoShowcase"`
				EnableFamilyArrayTiles                 bool   `json:"enableFamilyArrayTiles"`
				UseSingleStoreCall                     bool   `json:"useSingleStoreCall"`
				ShowEmailModal                         bool   `json:"showEmailModal"`
				AddNoIndexToWexPdp                     bool   `json:"addNoIndexToWexPdp"`
				AddNoIndexToWexArray                   bool   `json:"addNoIndexToWexArray"`
				UseServerDataURL                       bool   `json:"useServerDataURL"`
				EnableClaripDNS                        bool   `json:"enableClaripDNS"`
				EnableResponsiveQuickshop              bool   `json:"enableResponsiveQuickshop"`
				ShowRQSButton                          bool   `json:"showRQSButton"`
				ForceArrayToRQS                        bool   `json:"forceArrayToRQS"`
				EnableYottaaScript                     bool   `json:"enableYottaaScript"`
				AddNoFollowFilters                     bool   `json:"addNoFollowFilters"`
				OmnitureAccount                        string `json:"omnitureAccount"`
				SiteID                                 string `json:"siteID"`
				SourceCode                             string `json:"sourceCode"`
				MobileEmailSourceID                    string `json:"mobileEmailSourceID"`
				DesktopEmailSourceID                   string `json:"desktopEmailSourceID"`
				FooterEmailSourceID                    string `json:"footerEmailSourceID"`
				VendorEmailSourceID                    string `json:"vendorEmailSourceID"`
				LoyaltyEmailSourceID                   string `json:"loyaltyEmailSourceID"`
				GenericEmailSourceID                   string `json:"genericEmailSourceID"`
				MonetateTag                            string `json:"monetateTag"`
				FbAdmins                               string `json:"fbAdmins"`
				FbAppID                                string `json:"fbAppId"`
				UseProductAPI                          bool   `json:"useProductApi"`
				IsThirdParty                           bool   `json:"isThirdParty"`
				IsEmbed                                bool   `json:"isEmbed"`
				IsFromAMP                              bool   `json:"isFromAMP"`
				InitialGlobalTopPromoPage              bool   `json:"initialGlobalTopPromoPage"`
				AkamaiRefererCookie                    string `json:"akamaiRefererCookie"`
				UseOcapi                               bool   `json:"useOcapi"`
				Host                                   string `json:"host"`
				AppInfo                                struct {
					PubVersion string `json:"pubVersion"`
				} `json:"appInfo"`
				IsMobile bool   `json:"isMobile"`
				IsTablet bool   `json:"isTablet"`
				Referrer string `json:"referrer"`
				Page     string `json:"page"`
			} `json:"configuration"`
			Array struct {
				Supplements struct {
					InStockStyledWithList []interface{} `json:"inStockStyledWithList"`
				} `json:"supplements"`
				Filters struct {
					RefinementsByID struct {
					} `json:"refinementsById"`
					RefinementGroups struct {
					} `json:"refinementGroups"`
					SelectedByID struct {
					} `json:"selectedById"`
					ShouldPersistFilters bool          `json:"shouldPersistFilters"`
					SelectedIdsOrdered   []interface{} `json:"selectedIdsOrdered"`
				} `json:"filters"`
				FindInStoreSfcc struct {
					IsFetchingStores bool          `json:"isFetchingStores"`
					IsToggledOn      bool          `json:"isToggledOn"`
					Stores           []interface{} `json:"stores"`
					NoResults        bool          `json:"noResults"`
					ShowModal        bool          `json:"showModal"`
				} `json:"findInStoreSfcc"`
				Pagination struct {
					PageIndex      int `json:"pageIndex"`
					ResultsPerPage int `json:"resultsPerPage"`
				} `json:"pagination"`
				SelectedSort struct {
					Value string `json:"value"`
					Label string `json:"label"`
				} `json:"selectedSort"`
				SelectedSubcategory struct {
					Name     string `json:"name"`
					SafeName string `json:"safeName"`
					Value    string `json:"value"`
				} `json:"selectedSubcategory"`
				Data struct {
					HasSplitResults bool `json:"hasSplitResults"`
					IsFetching      bool `json:"isFetching"`
					IsFiltering     bool `json:"isFiltering"`
					ProductArray    struct {
					} `json:"productArray"`
					SearchTerm       string `json:"searchTerm"`
					LastFilterAction struct {
					} `json:"lastFilterAction"`
				} `json:"data"`
				ArrayDynamicData struct {
					IsFetching      bool `json:"isFetching"`
					IsFetchComplete bool `json:"isFetchComplete"`
				} `json:"arrayDynamicData"`
				CreativeContent struct {
				} `json:"creativeContent"`
				ProductTiles struct {
					ToggleArrayImageOnHover  bool          `json:"toggleArrayImageOnHover"`
					ShowLaydownImagesOnArray bool          `json:"showLaydownImagesOnArray"`
					ShowFabricSwatchOnArray  bool          `json:"showFabricSwatchOnArray"`
					ExcludedFabricSwatch     []interface{} `json:"excludedFabricSwatch"`
				} `json:"productTiles"`
			} `json:"array"`
			Cart struct {
				CartSize                  int  `json:"cartSize"`
				IsMiniCartFetching        bool `json:"isMiniCartFetching"`
				IsMiniCartFetchComplete   bool `json:"isMiniCartFetchComplete"`
				HasMiniCartErrors         bool `json:"hasMiniCartErrors"`
				ShowMiniCart              bool `json:"showMiniCart"`
				ShowFirstItem             bool `json:"showFirstItem"`
				IsPartialQuantity         bool `json:"isPartialQuantity"`
				IsRemoveItemFetchComplete bool `json:"isRemoveItemFetchComplete"`
				IsRemoveItemSuccess       bool `json:"isRemoveItemSuccess"`
				IsFixedAddToBag           bool `json:"isFixedAddToBag"`
				ShouldRedirectToBag       bool `json:"shouldRedirectToBag"`
				IsAddingToBag             bool `json:"isAddingToBag"`
				MiniCartInfo              struct {
				} `json:"miniCartInfo"`
			} `json:"cart"`
			User struct {
				IsFetchComplete bool `json:"isFetchComplete"`
				HasErrors       bool `json:"hasErrors"`
				Location        struct {
				} `json:"location"`
			} `json:"user"`
			Account struct {
				ShowHeaderMenu    bool `json:"showHeaderMenu"`
				ShowRewardsWidget bool `json:"showRewardsWidget"`
			} `json:"account"`
			OrderHistory struct {
				CurrentPage            int    `json:"currentPage"`
				FetchingProductRecs    bool   `json:"fetchingProductRecs"`
				FetchedProductRecs     bool   `json:"fetchedProductRecs"`
				OrderHistoryFetched    bool   `json:"orderHistoryFetched"`
				OrderHistoryFetchError string `json:"orderHistoryFetchError"`
				OrdersByID             struct {
				} `json:"ordersById"`
				Pages                 []interface{} `json:"pages"`
				ShowMoreRecs          bool          `json:"showMoreRecs"`
				SubmittingOrderUpdate bool          `json:"submittingOrderUpdate"`
				TotalOrders           int           `json:"totalOrders"`
				OrderHistorySource    string        `json:"orderHistorySource"`
				StoreOrderHistory     struct {
					OrderHistoryFetched       bool          `json:"orderHistoryFetched"`
					FetchingStoreOrderHistory bool          `json:"fetchingStoreOrderHistory"`
					TotalOrders               int           `json:"totalOrders"`
					OrderHistoryFetchError    string        `json:"orderHistoryFetchError"`
					Pages                     []interface{} `json:"pages"`
					OrdersByID                struct {
					} `json:"ordersById"`
				} `json:"storeOrderHistory"`
			} `json:"orderHistory"`
			EmailCapture struct {
				Display          bool `json:"display"`
				DisplayError     bool `json:"displayError"`
				DisplaySuccess   bool `json:"displaySuccess"`
				ShouldTransition bool `json:"shouldTransition"`
				Bottom           int  `json:"bottom"`
			} `json:"emailCapture"`
			Breadcrumbs []struct {
				Label string `json:"label"`
				Path  string `json:"path"`
			} `json:"breadcrumbs"`
			Clienteling struct {
				AssociateSignin struct {
					SuccessComplete            bool        `json:"successComplete"`
					DidCustomerEmailInvalidate bool        `json:"didCustomerEmailInvalidate"`
					DidAssociateIDInvalidate   bool        `json:"didAssociateIdInvalidate"`
					DidInvalidate              bool        `json:"didInvalidate"`
					Response                   interface{} `json:"response"`
				} `json:"associateSignin"`
				AssociateSignout struct {
				} `json:"associateSignout"`
				ExtendAssociateSession struct {
					SuccessComplete bool        `json:"successComplete"`
					Response        interface{} `json:"response"`
				} `json:"extendAssociateSession"`
			} `json:"clienteling"`
			Sale struct {
				Data struct {
					SaleData struct {
					} `json:"saleData"`
					FetchedSalePromo bool `json:"fetchedSalePromo"`
				} `json:"data"`
			} `json:"sale"`
			CategoryByPath struct {
			} `json:"categoryByPath"`
			CreativePage struct {
			} `json:"creativePage"`
			ModuleManager struct {
			} `json:"moduleManager"`
			SeoProperties struct {
				Title               string `json:"title"`
				Description         string `json:"description"`
				CanonicalURL        string `json:"canonicalUrl"`
				SidecarCanonicalURL string `json:"sidecarCanonicalUrl"`
				ParentCategory      string `json:"parentCategory"`
				Category            struct {
					CategoryCanonicalURL string `json:"categoryCanonicalUrl"`
					CategoryPrevURL      string `json:"categoryPrevUrl"`
					CategoryNextURL      string `json:"categoryNextUrl"`
				} `json:"category"`
			} `json:"seoProperties"`
			Baynote struct {
				BadgeRecs struct {
				} `json:"badgeRecs"`
				BaynoteRecentlyViewed struct {
				} `json:"baynoteRecentlyViewed"`
				ProductRecs struct {
				} `json:"productRecs"`
			} `json:"baynote"`
			Certona struct {
			} `json:"certona"`
			Feature struct {
			} `json:"feature"`
			FindInStore struct {
				IsFetching bool `json:"isFetching"`
				Error      bool `json:"error"`
				StoresByID struct {
				} `json:"storesById"`
				ValidStoresByPath struct {
				} `json:"validStoresByPath"`
			} `json:"findInStore"`
			Omniture struct {
				PageName   string `json:"pageName"`
				Server     string `json:"server"`
				Channel    string `json:"channel"`
				PageType   string `json:"pageType"`
				Prop1      string `json:"prop1"`
				Prop2      string `json:"prop2"`
				Prop3      string `json:"prop3"`
				Prop4      string `json:"prop4"`
				Prop5      string `json:"prop5"`
				Prop6      string `json:"prop6"`
				Prop7      string `json:"prop7"`
				Prop8      string `json:"prop8"`
				Prop9      string `json:"prop9"`
				Prop10     string `json:"prop10"`
				Prop11     string `json:"prop11"`
				Prop12     string `json:"prop12"`
				Prop13     string `json:"prop13"`
				Prop14     string `json:"prop14"`
				Prop15     string `json:"prop15"`
				Prop16     string `json:"prop16"`
				Prop17     string `json:"prop17"`
				Prop18     string `json:"prop18"`
				Prop19     string `json:"prop19"`
				Prop30     string `json:"prop30"`
				Prop20     string `json:"prop20"`
				Prop21     string `json:"prop21"`
				Prop23     string `json:"prop23"`
				Prop31     string `json:"prop31"`
				Prop33     string `json:"prop33"`
				Prop39     string `json:"prop39"`
				Prop44     string `json:"prop44"`
				Prop50     string `json:"prop50"`
				Prop64     string `json:"prop64"`
				Prop69     string `json:"prop69"`
				Campaign   string `json:"campaign"`
				State      string `json:"state"`
				Zip        string `json:"zip"`
				Events     string `json:"events"`
				Products   string `json:"products"`
				PurchaseID string `json:"purchaseID"`
				EVar1      string `json:"eVar1"`
				EVar2      string `json:"eVar2"`
				EVar3      string `json:"eVar3"`
				EVar4      string `json:"eVar4"`
				EVar5      string `json:"eVar5"`
				EVar6      string `json:"eVar6"`
				EVar7      string `json:"eVar7"`
				EVar8      string `json:"eVar8"`
				EVar9      string `json:"eVar9"`
				EVar10     string `json:"eVar10"`
				EVar11     string `json:"eVar11"`
				EVar12     string `json:"eVar12"`
				EVar13     string `json:"eVar13"`
				EVar14     string `json:"eVar14"`
				EVar15     string `json:"eVar15"`
				EVar16     string `json:"eVar16"`
				EVar17     string `json:"eVar17"`
				EVar18     string `json:"eVar18"`
				EVar19     string `json:"eVar19"`
				EVar20     string `json:"eVar20"`
				EVar21     string `json:"eVar21"`
				EVar22     string `json:"eVar22"`
				EVar23     string `json:"eVar23"`
				EVar24     string `json:"eVar24"`
				EVar25     string `json:"eVar25"`
				EVar26     string `json:"eVar26"`
				EVar27     string `json:"eVar27"`
				EVar28     string `json:"eVar28"`
				EVar29     string `json:"eVar29"`
				EVar30     string `json:"eVar30"`
				EVar31     string `json:"eVar31"`
				EVar32     string `json:"eVar32"`
				EVar33     string `json:"eVar33"`
				EVar34     string `json:"eVar34"`
				EVar35     string `json:"eVar35"`
				EVar36     string `json:"eVar36"`
				EVar37     string `json:"eVar37"`
				EVar38     string `json:"eVar38"`
				EVar39     string `json:"eVar39"`
				EVar40     string `json:"eVar40"`
				EVar44     string `json:"eVar44"`
				EVar50     string `json:"eVar50"`
				EVar57     string `json:"eVar57"`
				EVar62     string `json:"eVar62"`
				EVar69     string `json:"eVar69"`
				EVar73     string `json:"eVar73"`
				EVar74     string `json:"eVar74"`
				EVar95     string `json:"eVar95"`
				List1      string `json:"list1"`
			} `json:"omniture"`
			Routing struct {
				Query struct {
					Sale        string `json:"sale"`
					IsFromSale  string `json:"isFromSale"`
					ColorName   string `json:"color_name"`
					CountryCode string `json:"countryCode"`
					Merged      string `json:"merged"`
					Gender      string `json:"gender"`
					Category    string `json:"category"`
					Subcategory string `json:"subcategory"`
					ProductSlug string `json:"productSlug"`
					ProductCode string `json:"productCode"`
				} `json:"query"`
				PrevQuery struct {
				} `json:"prevQuery"`
				Pathname              string `json:"pathname"`
				Search                string `json:"search"`
				IsRouteChangeComplete bool   `json:"isRouteChangeComplete"`
			} `json:"routing"`
			RequestHeaders struct {
				XRequestSessionID string `json:"x-request-session-id"`
				XRequestID        string `json:"x-request-id"`
				XJcgDomain        string `json:"x-jcg-domain"`
			} `json:"requestHeaders"`
			Debug struct {
				ProductAPI struct {
				} `json:"productApi"`
			} `json:"debug"`
			LiveChat struct {
				IsOpen bool   `json:"isOpen"`
				Origin string `json:"origin"`
			} `json:"liveChat"`
		} `json:"initialState"`
		InitialProps struct {
			PageProps struct {
				IsInitialLoad bool `json:"isInitialLoad"`
				Location      struct {
					Query struct {
						Sale        string `json:"sale"`
						IsFromSale  string `json:"isFromSale"`
						ColorName   string `json:"color_name"`
						CountryCode string `json:"countryCode"`
						Merged      string `json:"merged"`
						Gender      string `json:"gender"`
						Category    string `json:"category"`
						Subcategory string `json:"subcategory"`
						ProductSlug string `json:"productSlug"`
						ProductCode string `json:"productCode"`
					} `json:"query"`
					Pathname string `json:"pathname"`
				} `json:"location"`
				ViewType string `json:"viewType"`
			} `json:"pageProps"`
			IsInitialLoad bool `json:"isInitialLoad"`
			Query         struct {
				Sale        string `json:"sale"`
				IsFromSale  string `json:"isFromSale"`
				ColorName   string `json:"color_name"`
				CountryCode string `json:"countryCode"`
				Merged      string `json:"merged"`
				Gender      string `json:"gender"`
				Category    string `json:"category"`
				Subcategory string `json:"subcategory"`
				ProductSlug string `json:"productSlug"`
				ProductCode string `json:"productCode"`
			} `json:"query"`
		} `json:"initialProps"`
	} `json:"props"`
	Page  string `json:"page"`
	Query struct {
		Sale        string `json:"sale"`
		IsFromSale  string `json:"isFromSale"`
		ColorName   string `json:"color_name"`
		CountryCode string `json:"countryCode"`
		Merged      string `json:"merged"`
		Gender      string `json:"gender"`
		Category    string `json:"category"`
		Subcategory string `json:"subcategory"`
		ProductSlug string `json:"productSlug"`
		ProductCode string `json:"productCode"`
	} `json:"query"`
	BuildID       string `json:"buildId"`
	RuntimeConfig struct {
		BAYNOTEHASH          string `json:"BAYNOTE_HASH"`
		BRAND                string `json:"BRAND"`
		CMSURL               string `json:"CMS_URL"`
		DATAURL              string `json:"DATA_URL"`
		DOCUMENTCSSHASH      string `json:"DOCUMENT_CSS_HASH"`
		SWHASH               string `json:"SW_HASH"`
		USEAKAMAIORIGINCHECK string `json:"USE_AKAMAI_ORIGIN_CHECK"`
	} `json:"runtimeConfig"`
	IsFallback   bool            `json:"isFallback"`
	CustomServer bool            `json:"customServer"`
	Gip          bool            `json:"gip"`
	AppGip       bool            `json:"appGip"`
	Head         [][]interface{} `json:"head"`
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
	// c.logger.Debugf("data: %s", matched[1])

	var viewData productPageResponse
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}

	prodid := viewData.Props.InitialState.Products.ProductsByProductCode.ProductCode
	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:       strconv.Format(viewData.Props.InitialState.Products.ProductsByProductCode.ProductCode),
			CrawlUrl: resp.Request.URL.String(),
		},
		BrandName:   viewData.Props.InitialState.Products.ProductsByProductCode.Brand,
		Title:       viewData.Props.InitialState.Products.ProductsByProductCode.ProductName,
		Description: viewData.Props.InitialState.Products.ProductsByProductCode.ProductDescriptionRomance + " " + strings.Join(viewData.Props.InitialState.Products.ProductsByProductCode.ProductDescriptionTech, " ") + " " + strings.Join(viewData.Props.InitialState.Products.ProductsByProductCode.ProductDescriptionFit, " "),
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		// Stats: &pbItem.Stats{
		// 	ReviewCount: int32(p.NumberOfReviews),
		// 	Rating:      float32(p.ReviewAverageRating / 5.0),
		// },
	}

	var colorImg [len(viewData.Props.InitialState.Products.ProductsByProductCode.ColorsList)]string
	var colorIndex int = 0

	for _, rawSku := range viewData.Props.InitialState.Products.ProductsByProductCode.Skus {

		originalPrice, _ := strconv.ParseFloat(rawSku.Price.Amount)
		msrp, _ := strconv.ParseFloat(rawSku.ListPrice.Amount)
		discount, _ := (msrp - originalPrice) * 100 / msrp
		sku := pbItem.Sku{
			SourceId: strconv.Format(rawSku.ID),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(originalPrice * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},
			Stock: &pbItem.Stock{StockStatus: pbItem.Stock_InStock},
		}
		// if rawSku.TotalQuantityAvailable > 0 {
		// 	sku.Stock.StockStatus = pbItem.Stock_InStock
		// 	sku.Stock.StockCount = int32(rawSku.TotalQuantityAvailable)
		// }

		// color

		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecColor,
			Id:    strconv.Format(rawSku.ColorCode),
			Name:  rawSku.ColorName,
			Value: rawSku.ColorName,
			//Icon:  color.SwatchMedia.Mobile,
		})

		if Contains(colorImg, rawSku.ColorName) == false {
			colorImg[colorIndex] = rawSku.ColorName
			colorIndex += 1
			skuShotCode := ""
			for _, mid := range viewData.Props.InitialState.Products.ProductsByProductCode.PriceModel.Colors {
				if mid.ColorCode == rawSku.ColorCode
				{
					skuShotCode = mid.SkuShotType
				}
			}
			
			//img
			isDefault := true
			for ki, mid := range strings.Split(skuShotCode, ",") {
				str_im := ""
				if ki > 0 {
					isDefault = false
				}

				if mid == "" {
					str_im = "https://www.jcrew.com/s7-img-facade/" + prodid + "_" + rawSku.ColorCode + "?fmt=jpeg&qlt=90,0&resMode=sharp&op_usm=.1,0,0,0&crop=0,0,0,0"

				} else {
					str_im = "https://www.jcrew.com/s7-img-facade/" + prodid + "_" + rawSku.ColorCode + mid + "?fmt=jpeg&qlt=90,0&resMode=sharp&op_usm=.1,0,0,0&crop=0,0,0,0"
				}
				    
				sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
					strconv.Format(prodid + "_" + rawSku.ColorCode),
					str_im & "&wid=1200&hei=1200",
					str_im & "&wid=1200&hei=1200",
					str_im & "&wid=500&hei=500",
					str_im & "&wid=700&hei=700",
					"",
					isDefault,
				))

			}
		}

		// size

		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Id:    rawSku.SkuID,
			Name:  rawSku.Size,
			Value: rawSku.Size,
		})

		item.SkuItems = append(item.SkuItems, &sku)
	}

	// yield item result
	if err = yield(ctx, &item); err != nil {
		return err
	}

	return nil
}

func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
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
	var (
		// get ProxyCrawl's API Token from you run environment
		apiToken = os.Getenv("PC_API_TOKEN")
		// get ProxyCrawl's Javascript Token from you run environment
		jsToken = os.Getenv("PC_JS_TOKEN")
	)
	if apiToken == "" || jsToken == "" {
		panic("env PC_API_TOKEN or PC_JS_TOKEN is not set")
	}

	// build a logger object.
	logger := glog.New(glog.LogLevelDebug)
	// build a http client
	client, err := proxy.NewProxyClient(
		// cookie jar used for auto cookie management.
		cookiejar.New(),
		logger,
		proxy.Options{APIToken: apiToken, JSToken: jsToken},
	)
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
				i.URL.Host = "www.jcrew.com"
			}

			// do http requests here.
			nctx, cancel := context.WithTimeout(ctx, time.Minute*5)
			defer cancel()
			resp, err := client.DoWithOptions(nctx, i, http.Options{
				EnableProxy:       true,
				EnableHeadless:    false,
				EnableSessionInit: spider.CrawlOptions().EnableSessionInit,
				KeepSession:       spider.CrawlOptions().KeepSession,
				ProxyLevel:        http.ProxyLevelReliable,
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

	ctx := context.WithValue(context.Background(), "tracing_id", "jcrew_123456")
	// start the crawl request
	for _, req := range spider.NewTestRequest(context.Background()) {
		if err := callback(ctx, req); err != nil {
			logger.Fatal(err)
		}
	}
}
