package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/voiladev/go-crawler/pkg/cli"
	"github.com/voiladev/go-crawler/pkg/crawler"
	"github.com/voiladev/go-crawler/pkg/net/http"
	pbMedia "github.com/voiladev/go-crawler/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/go-crawler/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/go-crawler/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	pbProxy "github.com/voiladev/go-crawler/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
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
		categoryPathMatcher: regexp.MustCompile(`^(/(c|r))?(/[a-z0-9\-_]+){1,5}$`),
		// this regular used to match product page url path
		productPathMatcher: regexp.MustCompile(`^/(p|m)(/[a-z0-9\-_]+){1,6}/[0-9A-Z]+/?$`),

		logger: logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "30a9ec50922dfa71c781820c6c0a7f94"
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
func (c *_Crawler) CrawlOptions(u *url.URL) *crawler.CrawlOptions {
	options := &crawler.CrawlOptions{
		EnableHeadless: false,
		// use js api to init session for the first request of the crawl
		EnableSessionInit: false,
		// Medium
		Reliability: pbProxy.ProxyReliability_ReliabilityMedium,
	}
	options.MustCookies = append(options.MustCookies,
		&http.Cookie{Name: "jcrew_country", Value: "US", Path: "/"},
		&http.Cookie{Name: "bluecoreNV", Value: "false", Path: "/"},
		&http.Cookie{Name: "us_site", Value: "true", Path: "/"},
		&http.Cookie{Name: "MR", Value: "0"},
		&http.Cookie{Name: "jcrew_wc", Value: "yes"},
		&http.Cookie{Name: "AKA_A2", Value: "B"},
		&http.Cookie{Name: "s_slt", Value: "%5B%5BB%5D%5D"},
		&http.Cookie{Name: "s_sq", Value: "%5B%5BB%5D%5D"},
		&http.Cookie{Name: "s_cc", Value: "true"},
	)
	return options
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.jcrew.com"}
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
		u.Host = "www.jcrew.com"
	}
	if c.productPathMatcher.MatchString(u.Path) {
		u.RawQuery = ""
		return u.String(), nil
	}
	return rawurl, nil
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

	if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yield)
	} else if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
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

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Debug(err)
		return err
	}
	rawData := strings.TrimSpace(dom.Find("script#__NEXT_DATA__").Text())

	var viewData CategoryView
	if err := json.Unmarshal([]byte(rawData), &viewData); err != nil {
		c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		return err
	}

	lastIndex := nextIndex(ctx)
	productData := viewData.Props.InitialState.Array.Data.ProductArray
	for _, prodList := range productData.ProductList {
		for _, idv := range prodList.Products {
			u, err := url.Parse(idv.URL)
			if err != nil {
				c.logger.Error(err)
				continue
			}
			if strings.HasPrefix(u.Path, "/nl/") {
				u.Path = strings.TrimPrefix(u.Path, "/nl")
			}
			req, err := http.NewRequest(http.MethodGet, u.String(), nil)
			if err != nil {
				c.logger.Errorf("load http request of url %s failed, error=%s", resp.Request.URL, err)
				return err
			}

			// set the index of the product crawled in the sub response
			nctx := context.WithValue(ctx, "item.index", lastIndex)
			lastIndex += 1

			// yield sub request
			if err := yield(nctx, req); err != nil {
				return err
			}
		}
	}

	if productData.Pagination.PageIndex >= productData.Pagination.TotalPage {
		return nil
	}

	// set pagination
	u := *resp.Request.URL
	vals := u.Query()
	vals.Set("Npge", strconv.Format(productData.Pagination.PageIndex+1))
	u.RawQuery = vals.Encode()

	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	nctx := context.WithValue(ctx, "item.index", lastIndex)

	return yield(nctx, req)
}

type productPageResponse struct {
	Props struct {
		IsServer     bool `json:"isServer"`
		InitialState struct {
			Content struct {
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
			Navigation struct {
				Data struct {
					Nav             []interface{} `json:"nav"`
					IsFetching      bool          `json:"isFetching"`
					IsFetchComplete bool          `json:"isFetchComplete"`
				} `json:"data"`
			} `json:"navigation"`
			Products struct {
				ProductsByProductCode map[string]struct {
					ProductCode               string                       `json:"productCode"`
					ProductDataFetched        bool                         `json:"productDataFetched"`
					LastUpdated               int                          `json:"lastUpdated"`
					PdpIntlMessage            string                       `json:"pdpIntlMessage"`
					IsPreorder                bool                         `json:"isPreorder"`
					ShipRestricted            bool                         `json:"shipRestricted"`
					IsFindInStore             bool                         `json:"isFindInStore"`
					LimitQuantity             interface{}                  `json:"limit-quantity"`
					JspURL                    string                       `json:"jspUrl"`
					ProductDescriptionRomance string                       `json:"productDescriptionRomance"`
					ProductDescriptionFit     []string                     `json:"productDescriptionFit"`
					IsVPS                     bool                         `json:"isVPS"`
					StyledWithSkus            string                       `json:"styledWithSkus"`
					PriceCallArgs             []interface{}                `json:"price-call-args"`
					OlapicCopy                string                       `json:"olapicCopy"`
					SwatchOrderAlphabetical   bool                         `json:"swatchOrderAlphabetical"`
					ColorsMap                 map[string]map[string]string `json:"colorsMap"`
					IsFreeShipping            bool                         `json:"isFreeShipping"`
					ProductName               string                       `json:"productName"`
					ProductsByComma           string                       `json:"products-by-comma"`
					Brand                     string                       `json:"brand"`
					PriceEndpoint             string                       `json:"priceEndpoint"`
					SizeChart                 string                       `json:"sizeChart"`
					SizesMap                  map[string]map[string]string `json:"sizesMap"`
					ListPrice                 struct {
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
					ShotTypes  []string `json:"shotTypes"`
					Variations []struct {
						APILink         string `json:"apiLink"`
						Name            string `json:"name"`
						ProductName     string `json:"productName"`
						ATRFreeShipping string `json:"ATR_free_shipping"`
						Label           string `json:"label"`
						URL             string `json:"url"`
						ProductCode     string `json:"productCode"`
						PrdID           int64  `json:"prd_id"`
						CanonicalURL    string `json:"canonicalUrl"`
					} `json:"variations"`
					Skus map[string]struct {
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
					SelectedProductCode   string        `json:"selectedProductCode"`
					SelectedColorName     string        `json:"selectedColorName"`
					SelectedColorCode     string        `json:"selectedColorCode"`
					SelectedQuantity      int           `json:"selectedQuantity"`
					SelectedSize          string        `json:"selectedSize"`
					FitModelDetails       []interface{} `json:"fitModelDetails"`
					UseProductAPI         bool          `json:"useProductApi"`
					IsStoresShowAll       bool          `json:"isStoresShowAll"`
					PriceModel            map[string]struct {
						ListPrice struct {
							Amount    float64 `json:"amount"`
							Formatted string  `json:"formatted"`
						} `json:"listPrice"`
						Colors []struct {
							SalePrice struct {
								Amount    float64 `json:"amount"`
								Formatted string  `json:"formatted"`
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
							Amount    float64 `json:"amount"`
							Formatted string  `json:"formatted"`
						} `json:"was"`
						Now struct {
							Amount    float64 `json:"amount"`
							Formatted string  `json:"formatted"`
						} `json:"now"`
						DiscountPercentage int    `json:"discountPercentage"`
						ExtendedSize       string `json:"extendedSize"`
						Badge              struct {
							Label    string `json:"label"`
							Type     string `json:"type"`
							Priority int    `json:"priority"`
							IconPath string `json:"iconPath"`
						} `json:"badge"`
					} `json:"priceModel"`
					HasVariations bool `json:"hasVariations"`
				} `json:"productsByProductCode"`
				IsFetching     bool `json:"isFetching"`
				PdpDynamicData struct {
					IsFetching      bool          `json:"isFetching"`
					IsFetchComplete bool          `json:"isFetchComplete"`
					JSON            []interface{} `json:"json"`
				} `json:"pdpDynamicData"`
				LastUpdated time.Time `json:"lastUpdated"`
			} `json:"products"`
			ProductDetail struct {
				// EVar78Pdp string `json:"eVar78pdp"`
				Flyzoom struct {
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
			Breadcrumbs []struct {
				Label string `json:"label"`
				Path  string `json:"path"`
			} `json:"breadcrumbs"`
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
	IsFallback   bool `json:"isFallback"`
	CustomServer bool `json:"customServer"`
	Gip          bool `json:"gip"`
	AppGip       bool `json:"appGip"`
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

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Debug(err)
		return err
	}
	rawData := strings.TrimSpace(dom.Find("script#__NEXT_DATA__").Text())

	var viewData productPageResponse
	if err := json.Unmarshal([]byte(rawData), &viewData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}

	colorImgs := map[string][]*pbMedia.Media{}
	for code, prod := range viewData.Props.InitialState.Products.ProductsByProductCode {
		canUrl, _ := c.CanonicalUrl("https://www.jcrew.com" + prod.URL)
		item := pbItem.Product{
			Source: &pbItem.Source{
				Id:           code,
				CrawlUrl:     "https://www.jcrew.com" + prod.URL,
				CanonicalUrl: canUrl,
			},
			BrandName:   prod.Brand,
			Title:       prod.ProductName,
			Description: prod.ProductDescriptionRomance + " " + strings.Join(prod.ProductDescriptionTech, " "),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
			},
			CrowdType:   strings.TrimRight(viewData.Query.Gender, "_category"),
			Category:    strings.Replace(viewData.Query.Category, "_", " ", -1),
			SubCategory: strings.Replace(viewData.Query.Subcategory, "_", " ", -1),
			// Stats: &pbItem.Stats{
			// 	ReviewCount: int32(p.NumberOfReviews),
			// 	Rating:      float32(p.ReviewAverageRating / 5.0),
			// },
		}

		for _, model := range prod.PriceModel {
			for _, color := range model.Colors {
				var medias []*pbMedia.Media
				for ki, mid := range strings.Split(color.SkuShotType, ",") {
					str_im := ""
					if mid == "" {
						str_im = "https://www.jcrew.com/s7-img-facade/" + code +
							"_" + color.ColorCode + "?fmt=jpeg&qlt=90,0&resMode=sharp&op_usm=.1,0,0,0&crop=0,0,0,0"
					} else {
						str_im = "https://www.jcrew.com/s7-img-facade/" + code +
							"_" + color.ColorCode + mid + "?fmt=jpeg&qlt=90,0&resMode=sharp&op_usm=.1,0,0,0&crop=0,0,0,0"
					}

					medias = append(medias, pbMedia.NewImageMedia(
						"",
						str_im+"&wid=1200&hei=1200",
						str_im+"&wid=1200&hei=1200",
						str_im+"&wid=500&hei=500",
						str_im+"&wid=700&hei=700",
						"",
						ki == 0,
					))
					colorImgs[color.ColorCode] = medias
				}
			}
		}

		for _, rawSku := range prod.Skus {
			originalPrice, _ := strconv.ParsePrice(rawSku.Price.Amount)
			msrp, _ := strconv.ParsePrice(rawSku.ListPrice.Amount)
			discount := math.Ceil((msrp - originalPrice) * 100 / msrp)

			sku := pbItem.Sku{
				SourceId: rawSku.SkuID,
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(originalPrice * 100),
					Msrp:     int32(msrp * 100),
					Discount: int32(discount),
				},
				Medias: colorImgs[rawSku.ColorCode],
				Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}
			if rawSku.ShowOnSale {
				sku.Stock.StockStatus = pbItem.Stock_InStock
			}

			// color
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    strconv.Format(rawSku.ColorCode),
				Name:  rawSku.ColorName,
				Value: rawSku.ColorCode,
			})

			// size
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    rawSku.Size,
				Name:  rawSku.Size,
				Value: rawSku.Size,
			})

			item.SkuItems = append(item.SkuItems, &sku)
		}
		if err = yield(ctx, &item); err != nil {
			return err
		}

		if prod.HasVariations {
			for _, variation := range prod.Variations {
				if variation.ProductCode == code {
					continue
				}

				req, err := http.NewRequest(http.MethodGet, variation.URL, nil)
				if err != nil {
					c.logger.Error(err)
					return err
				}
				if err := yield(ctx, req); err != nil {
					return err
				}
			}
		}
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
		// "https://www.jcrew.com/c/womens_category/sweatshirts_sweatpants",
		// "https://www.jcrew.com/r/sale/men/discount-60-70-off/discount-70-and-above",
		"https://www.jcrew.com/p/mens_category/shirts/secret_wash/slim-stretch-secret-wash-shirt-in-organic-cotton-gingham/AA429",
		// "https://www.jcrew.com/p/womens_category/sweatshirts_sweatpants/pullovers/mariner-cloth-buttonup-hoodie/AW153?color_name=white-navy-mira-stripe",
		// "https://www.jcrew.com/p/girls_category/pajamas/nightgowns/girls-flannel-nightgown-in-tartan/AE512?color_name=white-out-plaid-red-navy",
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
	cli.NewApp(New).Run(os.Args)
}
