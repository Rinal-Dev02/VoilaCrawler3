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

	"github.com/PuerkitoBio/goquery"
	"github.com/voiladev/VoilaCrawler/pkg/cli"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	pbMedia "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
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
func (_ *_Crawler) New(_ *cli.Context, client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient: client,
		// this regular used to match category page url path
		///shop/Women/Clothing/Tops
		categoryPathMatcher: regexp.MustCompile(`^/(category|shop|c|events)(/[,\s%&a-zA-Z0-9_\-]+){1,6}$`),
		// this regular used to match product page url path
		productPathMatcher: regexp.MustCompile(`^/s(/.*){0,4}/n?\d+$`),
		logger:             logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "402d469cae679c43f0ab9ef99a6b59dd"
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
	return &crawler.CrawlOptions{
		EnableHeadless: false,
		// use js api to init session for the first request of the crawl
		EnableSessionInit: false,
	}
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.nordstromrack.com"}
}

// CanonicalUrl
func (c *_Crawler) CanonicalUrl(rawurl string) (string, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	if u.Host == "" {
		u.Host = "www.nordstromrack.com"
	}
	if c.productPathMatcher.MatchString(u.Path) {
		u.RawQuery = ""
		return u.String(), nil
	}
	return u.String(), nil
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
	c.logger.Debugf("%s", resp.Request.URL.Path)

	p := strings.TrimSuffix(resp.Request.URL.Path, "/")

	if p == "" {
		return c.parseCategories(ctx, resp, yield)
	}

	if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yield)
	} else if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	}
	return crawler.ErrUnsupportedPath
}

func (c *_Crawler) parseCategories(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	matched := categoryExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		//c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract category info from %s failed, error=%s", resp.Request.URL, err)
	}

	var viewData categoryStructure

	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal category detail data fialed, error=%s", err)
		return err
	}

	for _, rawcat := range viewData.Headerdesktop.Navigation {

		cateName := rawcat.Name
		if cateName == "" {
			continue
		}

		nnctx := context.WithValue(ctx, "Category", cateName)

		fmt.Println(`cateName `, cateName)

		for _, rawsubcat := range rawcat.Columns {

			for _, rawGroup := range rawsubcat.Groups {

				for _, rawsub2Node := range rawGroup.Nodes {

					for _, rawGroup2 := range rawsub2Node.Groups {

						for _, rawsubNode2 := range rawGroup2.Nodes {

							href := rawsub2Node.URI
							if href == "" {
								continue
							}

							u, err := url.Parse(href)
							if err != nil {
								c.logger.Error("parse url %s failed", href)
								continue
							}

							subCateName := rawsub2Node.Name + " > " + rawsubNode2.Name

							fmt.Println(subCateName)
							if c.categoryPathMatcher.MatchString(u.Path) {
								nnnctx := context.WithValue(nnctx, "SubCategory", subCateName)
								req, _ := http.NewRequest(http.MethodGet, href, nil)
								if err := yield(nnnctx, req); err != nil {
									return err
								}
							}
						}
					}

					if len(rawsub2Node.Groups) == 0 {

						href := rawsub2Node.URI
						if href == "" {
							continue
						}

						u, err := url.Parse(href)
						if err != nil {
							c.logger.Error("parse url %s failed", href)
							continue
						}

						subCateName := rawsub2Node.Name

						fmt.Println(subCateName)
						if c.categoryPathMatcher.MatchString(u.Path) {
							nnnctx := context.WithValue(nnctx, "SubCategory", subCateName)
							req, _ := http.NewRequest(http.MethodGet, href, nil)
							if err := yield(nnnctx, req); err != nil {
								return err
							}
						}

					}
				}
			}
		}
	}
	return nil
}

type categoryStructure struct {
	Headerdesktop struct {
		Navigation []struct {
			Name       string `json:"name"`
			URI        string `json:"uri"`
			Breadcrumb string `json:"breadcrumb"`
			Linkstyle  string `json:"linkStyle,omitempty"`
			Columns    []struct {
				Groups []struct {
					Name  string `json:"name"`
					Type  string `json:"type"`
					Nodes []struct {
						Name       string `json:"name"`
						URI        string `json:"uri"`
						Breadcrumb string `json:"breadcrumb"`
						Linkstyle  string `json:"linkStyle"`
						Groups     []struct {
							Name  string `json:"name"`
							Type  string `json:"type"`
							Nodes []struct {
								Name       string        `json:"name"`
								URI        string        `json:"uri"`
								Breadcrumb string        `json:"breadcrumb"`
								Linkstyle  string        `json:"linkStyle"`
								Groups     []interface{} `json:"groups"`
							} `json:"nodes"`
						} `json:"groups"`
					} `json:"nodes"`
				} `json:"groups"`
			} `json:"columns"`
		} `json:"navigation"`
	} `json:"headerDesktop"`
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
	Catalog struct {
		HasFilters bool `json:"hasFilters"`
		Filters    struct {
			Brands            []interface{} `json:"brands"`
			Categories        []interface{} `json:"categories"`
			Class             string        `json:"class"`
			Colors            []interface{} `json:"colors"`
			Context           interface{}   `json:"context"`
			Department        string        `json:"department"`
			Division          string        `json:"division"`
			IncludeFlash      bool          `json:"includeFlash"`
			IncludePersistent bool          `json:"includePersistent"`
			Limit             int           `json:"limit"`
			Page              int           `json:"page"`
			PriceRanges       []interface{} `json:"priceRanges"`
			Query             interface{}   `json:"query"`
			Shops             []interface{} `json:"shops"`
			Sizes             []interface{} `json:"sizes"`
			Sort              string        `json:"sort"`
			Subclass          interface{}   `json:"subclass"`
			NestedColors      bool          `json:"nestedColors"`
		} `json:"filters"`
		CatalogURLBase         string `json:"catalogUrlBase"`
		CurrentLoadedRowIndex  int    `json:"currentLoadedRowIndex"`
		IsBrandSearch          bool   `json:"isBrandSearch"`
		IsCustomCategorySearch bool   `json:"isCustomCategorySearch"`
		IsClearanceSearch      bool   `json:"isClearanceSearch"`
		IsLandingPage          bool   `json:"isLandingPage"`
		IsQuerySearch          bool   `json:"isQuerySearch"`
		IsQuickLookInProgress  bool   `json:"isQuickLookInProgress"`
		IsQuickLookVisible     bool   `json:"isQuickLookVisible"`
		IsShopsSearch          bool   `json:"isShopsSearch"`

		PageBase        string `json:"pageBase"`
		PageTitle       string `json:"pageTitle"`
		PageDescription string `json:"pageDescription"`
		Pages           []struct {
			Href       string `json:"href,omitempty"`
			IsCurrent  bool   `json:"isCurrent"`
			Label      string `json:"label"`
			PageNumber int    `json:"pageNumber,omitempty"`
		} `json:"pages"`
		Products []struct {
			AltImageSrc         string      `json:"altImageSrc,omitempty"`
			Brand               string      `json:"brand"`
			Color               string      `json:"color"`
			CustomerChoiceID    string      `json:"customerChoiceId"`
			EventID             interface{} `json:"eventId"`
			InitialImageSrc     string      `json:"initialImageSrc"`
			InventoryLevelLabel interface{} `json:"inventoryLevelLabel"`
			IsClearance         bool        `json:"isClearance"`
			IsInventoryLow      bool        `json:"isInventoryLow"`
			IsOnHold            bool        `json:"isOnHold"`
			IsSoldOut           bool        `json:"isSoldOut"`
			IsOnSale            bool        `json:"isOnSale"`
			IsClearTheRack      bool        `json:"isClearTheRack"`
			IsPriceVisible      bool        `json:"isPriceVisible"`
			ProductHref         string      `json:"productHref"`
			Source              string      `json:"source"`
			StyleID             int         `json:"styleId"`
			Title               string      `json:"title"`
			WebStyleID          interface{} `json:"webStyleId"`
			Prices              struct {
				Retail struct {
					Min float64 `json:"min"`
					Max float64 `json:"max"`
				} `json:"retail"`
				Regular struct {
					Min float64 `json:"min"`
					Max float64 `json:"max"`
				} `json:"regular"`
				Sale struct {
					Min float64 `json:"min"`
					Max float64 `json:"max"`
				} `json:"sale"`
			} `json:"prices"`
			Discounts struct {
				RetailToRegular struct {
					Min float64 `json:"min"`
					Max float64 `json:"max"`
				} `json:"retailToRegular"`
				RegularToSale struct {
					Min float64 `json:"min"`
					Max float64 `json:"max"`
				} `json:"regularToSale"`
				RetailToSale struct {
					Min float64 `json:"min"`
					Max float64 `json:"max"`
				} `json:"retailToSale"`
			} `json:"discounts"`
		} `json:"products"`
		QuickLookIndex int `json:"quickLookIndex"`
		Total          int `json:"total"`
	} `json:"catalog"`
}

// used to extract embaded json data in website page.
// more about golang regulation see here https://golang.org/pkg/regexp/syntax/
var categoryExtractReg = regexp.MustCompile(`(?U)<script\s*>window.__INITIAL_CONFIG__\s*=\s*(.*)</script>`)

var productsExtractReg = regexp.MustCompile(`(?U)<script\s*>window.__INITIAL_STATE__\s*=\s*(.*)</script>`)

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
	s := [][]byte{matched[1], []byte("}}")}
	bytesResult := bytes.Join(s, []byte(""))

	if err := json.Unmarshal(bytesResult, &viewData); err != nil {
		c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		return err
	}

	lastIndex := nextIndex(ctx)
	for _, idv := range viewData.Catalog.Products {

		req, err := http.NewRequest(http.MethodGet, idv.ProductHref, nil)
		if err != nil {
			c.logger.Errorf("load http request of url %s failed, error=%s", idv.ProductHref, err)
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
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}

	lastPageNo := len(viewData.Catalog.Pages)
	if lastPageNo < 2 {
		return nil
	}
	lastPageNo = viewData.Catalog.Pages[lastPageNo-2].PageNumber
	// check if this is the last page
	if page >= int64(lastPageNo) {
		return nil
	}

	// set pagination
	u := *resp.Request.URL
	vals := u.Query()
	vals.Set("page", strconv.Format(page+1))
	u.RawQuery = vals.Encode()

	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	// update the index of last page
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
}

type parseProductResponse struct {
	Viewdata struct {
		Accountcreation struct {
			Hasvalidpassword       bool `json:"hasValidPassword"`
			Passwordlength         int  `json:"passwordLength"`
			Triggeraccountcreation bool `json:"triggerAccountCreation"`
			Ispaymentdefault       bool `json:"isPaymentDefault"`
			Issuccess              bool `json:"isSuccess"`
			Error                  struct {
			} `json:"error"`
		} `json:"accountCreation"`
		Addresses     []interface{} `json:"addresses"`
		Apistatuscode interface{}   `json:"apiStatusCode"`
		Borderfreeurl string        `json:"borderfreeUrl"`
		Buyxgety      []interface{} `json:"buyxgety"`
		Canvas        struct {
		} `json:"canvas"`
		Creditcards []interface{} `json:"creditCards"`
		Customer    struct {
			Isregistered    bool   `json:"isRegistered"`
			Isloyaltymember bool   `json:"isLoyaltyMember"`
			Firstname       string `json:"firstName"`
			Lastname        string `json:"lastName"`
		} `json:"customer"`
		Donationdetails struct {
		} `json:"donationDetails"`
		Eligiblegwps    []interface{} `json:"eligibleGwps"`
		Eligiblesamples struct {
		} `json:"eligibleSamples"`
		Employee struct {
			Discountpercentage int `json:"discountPercentage"`
		} `json:"employee"`
		Enticements struct {
		} `json:"enticements"`
		Errors      []interface{} `json:"errors"`
		Exceptions  []interface{} `json:"exceptions"`
		Fulfillment struct {
			Preferredfulfillmentmethod interface{} `json:"preferredFulfillmentMethod"`
			Deliverypostalcode         string      `json:"deliveryPostalCode"`
			Pickuplocation             struct {
				Postalcode  interface{} `json:"postalCode"`
				Storenumber interface{} `json:"storeNumber"`
			} `json:"pickupLocation"`
		} `json:"fulfillment"`
		Giftcards   []interface{} `json:"giftCards"`
		Giftoptions struct {
		} `json:"giftOptions"`
		Inputvalidationerrors struct {
		} `json:"inputValidationErrors"`
		Ischeckingoutasguest bool   `json:"isCheckingOutAsGuest"`
		Isfasttrackflow      bool   `json:"isFastTrackFlow"`
		Isloading            bool   `json:"isLoading"`
		Isregistereduser     string `json:"isRegisteredUser"`
		Rewards              struct {
		} `json:"rewards"`
		Manualnotes []interface{} `json:"manualNotes"`
		Marketing   struct {
		} `json:"marketing"`
		Order struct {
			Contact struct {
				Email     string `json:"email"`
				Phone     string `json:"phone"`
				Marketing bool   `json:"marketing"`
			} `json:"contact"`
			Donations []interface{} `json:"donations"`
			Employee  struct {
				Number string `json:"number"`
			} `json:"employee"`
			Gwp      []interface{} `json:"gwp"`
			ID       interface{}   `json:"id"`
			Items    []interface{} `json:"items"`
			Metadata []interface{} `json:"metadata"`
			Payment  struct {
				Creditcard      interface{}   `json:"creditCard"`
				Paypal          interface{}   `json:"payPal"`
				Giftcards       []interface{} `json:"giftCards"`
				Manualnotes     []interface{} `json:"manualNotes"`
				Systematicnotes []interface{} `json:"systematicNotes"`
			} `json:"payment"`
			Pickupperson struct {
				Firstname string `json:"firstName"`
				Lastname  string `json:"lastName"`
			} `json:"pickupPerson"`
			Promocodes []interface{} `json:"promoCodes"`
			Rewards    struct {
			} `json:"rewards"`
			Samples         []interface{} `json:"samples"`
			Shippingaddress struct {
				ID           interface{} `json:"id"`
				Firstname    string      `json:"firstName"`
				Lastname     string      `json:"lastName"`
				Addressline1 string      `json:"addressLine1"`
				Addressline2 string      `json:"addressLine2"`
				City         string      `json:"city"`
				State        string      `json:"state"`
				Postalcode   string      `json:"postalCode"`
				Countrycode  string      `json:"countryCode"`
				Storenumber  string      `json:"storeNumber"`
				Default      bool        `json:"default"`
			} `json:"shippingAddress"`
			Shippingcostoverride interface{} `json:"shippingCostOverride"`
		} `json:"order"`
		Ordernumber       string        `json:"orderNumber"`
		Ordersubmiterrors []interface{} `json:"orderSubmitErrors"`
		Pricing           struct {
		} `json:"pricing"`
		Pricingsummary struct {
			Discount          int `json:"discount"`
			Estimatedshipping int `json:"estimatedShipping"`
			Estimatedsubtotal int `json:"estimatedSubtotal"`
			Estimatedtax      int `json:"estimatedTax"`
			Estimatedtotal    int `json:"estimatedTotal"`
			Exchangecredit    int `json:"exchangeCredit"`
			Expectedrefund    int `json:"expectedRefund"`
		} `json:"pricingSummary"`
		Promises struct {
		} `json:"promises"`
		Promisesbypostalcode struct {
		} `json:"promisesByPostalCode"`
		Promisesbystorenumber struct {
		} `json:"promisesByStoreNumber"`
		Promotions struct {
		} `json:"promotions"`
		Savedpaypal struct {
		} `json:"savedPayPal"`
		Shiptostore struct {
			Modalpostalcode             interface{}   `json:"modalPostalCode"`
			Storeslist                  []interface{} `json:"storesList"`
			Selectedstore               interface{}   `json:"selectedStore"`
			Isstoreselectionunavailable bool          `json:"isStoreSelectionUnavailable"`
		} `json:"shipToStore"`
		Shiptypepricing struct {
		} `json:"shipTypePricing"`
		Skudetails struct {
		} `json:"skuDetails"`
		Stores struct {
		} `json:"stores"`
		Storesbypostalcode struct {
		} `json:"storesByPostalCode"`
		Storesbystorenumber interface{}   `json:"storesByStoreNumber"`
		Systematicnotes     []interface{} `json:"systematicNotes"`
		Wishlistinfo        struct {
		} `json:"wishListInfo"`
		Viewdata struct {
			Checkout struct {
				View string `json:"view"`
			} `json:"checkout"`
			Checkoutsignin struct {
				Forcedsignouttype                   string `json:"forcedSignOutType"`
				Issettinginternationalshipping      bool   `json:"isSettingInternationalShipping"`
				Shouldinternationalcheckoutredirect bool   `json:"shouldInternationalCheckoutRedirect"`
			} `json:"checkoutSignIn"`
			Contactinfo struct {
				Mode         string `json:"mode"`
				Emailwarning bool   `json:"emailWarning"`
				Phonewarning bool   `json:"phoneWarning"`
			} `json:"contactInfo"`
			Fulfillment struct {
				Mode                     string      `json:"mode"`
				Pickupmodalpostalcode    interface{} `json:"pickupModalPostalCode"`
				Pickuppostalcode         interface{} `json:"pickupPostalCode"`
				Postalcode               interface{} `json:"postalCode"`
				Selecteddeliveryshiptype string      `json:"selectedDeliveryShipType"`
			} `json:"fulfillment"`
			Gifting struct {
				Initwishlistgiftingmodal       bool   `json:"initWishListGiftingModal"`
				Isgiftingmodalvisible          bool   `json:"isGiftingModalVisible"`
				Isusingwishlistaddress         bool   `json:"isUsingWishListAddress"`
				Iswishlistflow                 bool   `json:"isWishListFlow"`
				Shouldopenwishlistgiftingmodal bool   `json:"shouldOpenWishListGiftingModal"`
				Usesharedmessage               bool   `json:"useSharedMessage"`
				Wishlistid                     string `json:"wishListId"`
				Wishlistsharekey               string `json:"wishListShareKey"`
			} `json:"gifting"`
			Isloadordercomplete bool `json:"isLoadOrderComplete"`
			Itemssummary        struct {
				Isexpanded           bool   `json:"isExpanded"`
				Mode                 string `json:"mode"`
				Pricemodifyactivesku string `json:"priceModifyActiveSku"`
			} `json:"itemsSummary"`
			Modal struct {
				Ismodalvisible bool   `json:"isModalVisible"`
				Modaltype      string `json:"modalType"`
			} `json:"modal"`
			Order struct {
				View string `json:"view"`
			} `json:"order"`
			Ordersummary struct {
				Mode string `json:"mode"`
			} `json:"orderSummary"`
			Paylink struct {
				Carddata struct {
				} `json:"cardData"`
				Customerheaders struct {
				} `json:"customerHeaders"`
				Exception                     string `json:"exception"`
				Haspaylinkexception           bool   `json:"hasPaylinkException"`
				Ispaylinkpending              bool   `json:"isPaylinkPending"`
				Ispaylinkretrieved            bool   `json:"isPaylinkRetrieved"`
				Ispaylinksent                 bool   `json:"isPaylinkSent"`
				Paylinkoption                 string `json:"paylinkOption"`
				Submitpaymentauthdeclinecount int    `json:"submitPaymentAuthDeclineCount"`
			} `json:"paylink"`
			Payment struct {
				Autofill                   bool   `json:"autofill"`
				Billingsameasship          bool   `json:"billingSameAsShip"`
				Cardnumberdisplay          string `json:"cardNumberDisplay"`
				Isaffirmpayment            bool   `json:"isAffirmPayment"`
				Ispaypalfasttrackflow      bool   `json:"isPayPalFastTrackFlow"`
				Isretrievingaddressdetails bool   `json:"isRetrievingAddressDetails"`
				Mode                       string `json:"mode"`
				Nocvverror                 bool   `json:"noCvvError"`
				Securitycode               string `json:"securityCode"`
				Securitycodedisplay        string `json:"securityCodeDisplay"`
				Shouldsendpaylink          bool   `json:"shouldSendPaylink"`
			} `json:"payment"`
			Paymentcreditoptions struct {
				Availablepaymentcreditoptions []string `json:"availablePaymentCreditOptions"`
				Selectedpaymentcreditoption   string   `json:"selectedPaymentCreditOption"`
			} `json:"paymentCreditOptions"`
			Paymentoptions struct {
				Ispaymentoptionsrendered bool   `json:"isPaymentOptionsRendered"`
				Mode                     string `json:"mode"`
				Selectedpaymentoption    string `json:"selectedPaymentOption"`
			} `json:"paymentOptions"`
			Pickupperson struct {
				Mode string `json:"mode"`
			} `json:"pickupPerson"`
			Promotioncode struct {
				Isemployeediscountformvisible bool          `json:"isEmployeeDiscountFormVisible"`
				Ispromotioncodeformvisible    bool          `json:"isPromotionCodeFormVisible"`
				Validationexceptions          []interface{} `json:"validationExceptions"`
			} `json:"promotionCode"`
			Rewards struct {
				Isactivatingpersonalbonuspoints bool `json:"isActivatingPersonalBonusPoints"`
			} `json:"rewards"`
			Scrolltarget    string `json:"scrollTarget"`
			Shippingaddress struct {
				Autofill                   bool   `json:"autofill"`
				Isretrievingaddressdetails bool   `json:"isRetrievingAddressDetails"`
				Mode                       string `json:"mode"`
			} `json:"shippingAddress"`
			Suppressexceptions bool `json:"suppressExceptions"`
			Validationerror    bool `json:"validationError"`
			Viewportwidth      int  `json:"viewportWidth"`
		} `json:"viewData"`
		Isfetching           bool     `json:"isFetching"`
		Error                bool     `json:"error"`
		ID                   int      `json:"id"`
		Agegroups            []string `json:"ageGroups"`
		Reviewaveragerating  float64  `json:"reviewAverageRating"`
		Enticementplacements []struct {
			Name        string `json:"name"`
			Enticements []struct {
				Type  string `json:"type"`
				Title string `json:"title"`
			} `json:"enticements"`
		} `json:"enticementPlacements"`
		Brand struct {
			Brandname  string `json:"brandName"`
			Brandurl   string `json:"brandUrl"`
			Imsbrandid int    `json:"imsBrandId"`
		} `json:"brand"`
		Consumers           []string `json:"consumers"`
		Description         string   `json:"description"`
		Customizationcode   string   `json:"customizationCode"`
		Defaultgallerymedia struct {
			Stylemediaid  int    `json:"styleMediaId"`
			Colorid       string `json:"colorId"`
			Istrimmed     bool   `json:"isTrimmed"`
			Stylemediaids []int  `json:"styleMediaIds"`
		} `json:"defaultGalleryMedia"`
		Eventflags struct {
			Anniversaryphase string `json:"anniversaryPhase"`
		} `json:"eventFlags"`
		Features []string `json:"features"`
		Filters  struct {
			Color struct {
				Byid map[int]struct {
					//Num300 struct {
					ID              string `json:"id"`
					Code            string `json:"code"`
					Isselected      bool   `json:"isSelected"`
					Isdefault       bool   `json:"isDefault"`
					Value           string `json:"value"`
					Displayvalue    string `json:"displayValue"`
					Filtertype      string `json:"filterType"`
					Isavailablewith string `json:"isAvailableWith"`
					Relatedskuids   []int  `json:"relatedSkuIds"`
					Stylemediaids   []int  `json:"styleMediaIds"`
					Swatchmedia     struct {
						Desktop string `json:"desktop"`
						Mobile  string `json:"mobile"`
						Preview string `json:"preview"`
					} `json:"swatchMedia"`
					//	} `json:"300"`

				} `json:"byId"`
				Allids []string `json:"allIds"`
			} `json:"color"`
			Size struct {
				Byid map[string]struct {
					//Small struct {
					ID              string `json:"id"`
					Value           string `json:"value"`
					Displayvalue    string `json:"displayValue"`
					Groupvalue      string `json:"groupValue"`
					Filtertype      string `json:"filterType"`
					Relatedskuids   []int  `json:"relatedSkuIds"`
					Isavailablewith string `json:"isAvailableWith"`
					//} `json:"small"`
				} `json:"byId"`
				Allids []string `json:"allIds"`
			} `json:"size"`
			Width struct {
				Byid struct {
				} `json:"byId"`
				Allids []interface{} `json:"allIds"`
			} `json:"width"`
			Group struct {
				Byid map[string]struct {
					//Regular struct {
					Value               string   `json:"value"`
					Displayvalue        string   `json:"displayValue"`
					Filtertype          string   `json:"filterType"`
					Originalstylenumber string   `json:"originalStyleNumber"`
					Shoulddisplay       bool     `json:"shouldDisplay"`
					Relatedsizeids      []string `json:"relatedSizeIds"`
					//} `json:"regular"`
				} `json:"byId"`
				Allids []string `json:"allIds"`
			} `json:"group"`
		} `json:"filters"`
		Filteroptions []string `json:"filterOptions"`
		Fitandsize    struct {
			Contextualsizedetail string      `json:"contextualSizeDetail"`
			Fitguidetitle        string      `json:"fitGuideTitle"`
			Fitguideurl          string      `json:"fitGuideUrl"`
			Fitvideotitle        string      `json:"fitVideoTitle"`
			Fitvideourl          string      `json:"fitVideoUrl"`
			Sizecharttitle       interface{} `json:"sizeChartTitle"`
			Sizecharturl         string      `json:"sizeChartUrl"`
			Sizedetail           interface{} `json:"sizeDetail"`
		} `json:"fitAndSize"`
		Imtfitandsize struct {
		} `json:"imtFitAndSize"`
		Fitcategory          string `json:"fitCategory"`
		Gender               string `json:"gender"`
		Healthhazardcategory string `json:"healthHazardCategory"`
		Holidaydeliverycopy  struct {
		} `json:"holidayDeliveryCopy"`
		Poducttypecode                    string `json:"poductTypeCode"`
		Ingredients                       string `json:"ingredients"`
		Isanniversarystyle                bool   `json:"isAnniversaryStyle"`
		Isavailable                       bool   `json:"isAvailable"`
		Isbackordered                     bool   `json:"isBackOrdered"`
		Isbeauty                          bool   `json:"isBeauty"`
		Iscallcenteractive                bool   `json:"isCallCenterActive"`
		Ischanel                          bool   `json:"isChanel"`
		Iscustomerfirstaccesseligible     bool   `json:"isCustomerFirstAccessEligible"`
		Ispickupstoreeligible             bool   `json:"isPickUpStoreEligible"`
		Isinstoreonlybridal               bool   `json:"isInStoreOnlyBridal"`
		Isinternational                   bool   `json:"isInternational"`
		Isloadedfromserver                bool   `json:"isLoadedFromServer"`
		Ismwpv6                           bool   `json:"isMwpV6"`
		Isdynamicrenderon                 bool   `json:"isDynamicRenderOn"`
		Isplatformecustomizable           bool   `json:"isPlatformECustomizable"`
		Ispreorder                        bool   `json:"isPreOrder"`
		Isproductbuynoweligible           bool   `json:"isProductBuyNowEligible"`
		Isstylerestrictedfromintlshipping bool   `json:"isStyleRestrictedFromIntlShipping"`
		Isumap                            bool   `json:"isUmap"`
		Maxorderquantity                  int    `json:"maxOrderQuantity"`
		Number                            string `json:"number"`
		Numberofreviews                   int    `json:"numberOfReviews"`
		Pathalias                         string `json:"pathAlias"`
		Price                             struct {
			Byskuid map[int]struct {
				//Num30109227 struct {
				Currentpercentoff      string `json:"currentPercentOff"`
				Isinternationalpricing bool   `json:"isInternationalPricing"`
				Isoriginalpricerange   bool   `json:"isOriginalPriceRange"`
				Ispercentoffcompareat  bool   `json:"isPercentOffCompareAt"`
				Isrange                bool   `json:"isRange"`
				Maxpercentageoff       string `json:"maxPercentageOff"`
				Originalpricestring    string `json:"originalPriceString"`
				Previousepercentoff    string `json:"previousePercentOff"`
				Previouspricestring    string `json:"previousPriceString"`
				Pricestring            string `json:"priceString"`
				Saletype               string `json:"saleType"`
				Showsoldoutmessage     bool   `json:"showSoldOutMessage"`
				Showumapmessage        bool   `json:"showUMapMessage"`
				Showumapprice          bool   `json:"showUMapPrice"`
				Styleid                int    `json:"styleId"`
				Valuestatement         string `json:"valueStatement"`
				//} `json:"30109227"`
			} `json:"bySkuId"`
			Allskuids []int `json:"allSkuIds"`
			Style     struct {
				Allskusonsale          bool    `json:"allSkusOnSale"`
				Currentminprice        float64 `json:"currentMinPrice"`
				Currentmaxprice        float64 `json:"currentMaxPrice"`
				Currentpercentoff      string  `json:"currentPercentOff"`
				Iscleartherack         bool    `json:"isClearTheRack"`
				Isinternationalpricing bool    `json:"isInternationalPricing"`
				Isoriginalpricerange   bool    `json:"isOriginalPriceRange"`
				Ispercentoffcompareat  bool    `json:"isPercentOffCompareAt"`
				Isrange                bool    `json:"isRange"`
				Maxpercentageoff       string  `json:"maxPercentageOff"`
				Originalpricestring    string  `json:"originalPriceString"`
				Previousepercentoff    string  `json:"previousePercentOff"`
				Previouspricestring    string  `json:"previousPriceString"`
				Pricestring            string  `json:"priceString"`
				Saleenddate            string  `json:"saleEndDate"`
				Saletype               string  `json:"saleType"`
				Showsoldoutmessage     bool    `json:"showSoldOutMessage"`
				Showumapmessage        bool    `json:"showUMapMessage"`
				Showumapprice          bool    `json:"showUMapPrice"`
				Styleid                int     `json:"styleId"`
				Valuestatement         string  `json:"valueStatement"`
			} `json:"style"`
		} `json:"price"`
		Productengagement struct {
			Pageviews struct {
				Count string `json:"count"`
				Copy  string `json:"copy"`
			} `json:"pageViews"`
		} `json:"productEngagement"`
		Productname            string        `json:"productName"`
		Producttitle           string        `json:"productTitle"`
		Producttypename        string        `json:"productTypeName"`
		Producttypeparentname  string        `json:"productTypeParentName"`
		Salesvideoshot         interface{}   `json:"salesVideoShot"`
		Sellingstatement       string        `json:"sellingStatement"`
		Shoppersizepreferences []interface{} `json:"shopperSizePreferences"`
		Skus                   struct {
			Byid map[int]struct {
				//Num30109227 struct {
				ID                     int         `json:"id"`
				Backorderdate          interface{} `json:"backOrderDate"`
				Colorid                string      `json:"colorId"`
				Displaypercentoff      string      `json:"displayPercentOff"`
				Displayprice           string      `json:"displayPrice"`
				Isavailable            bool        `json:"isAvailable"`
				Isbackorder            bool        `json:"isBackOrder"`
				Price                  float64     `json:"price"`
				Sizeid                 string      `json:"sizeId"`
				Widthid                string      `json:"widthId"`
				Rmsskuid               int         `json:"rmsSkuId"`
				Totalquantityavailable int         `json:"totalQuantityAvailable"`
				Isfinalsale            bool        `json:"isFinalSale"`
				Iscleartherack         bool        `json:"isClearTheRack"`
				//} `json:"30109227"`
			} `json:"byId"`
			Allids []int `json:"allIds"`
		} `json:"skus"`
		Stylemedia struct {
			Byid map[int]struct {
				//Num5622024 struct {
				ID            int    `json:"id"`
				Colorid       string `json:"colorId"`
				Colorname     string `json:"colorName"`
				Imagemediauri struct {
					Smalldesktop string `json:"smallDesktop"`
					Largedesktop string `json:"largeDesktop"`
					Zoom         string `json:"zoom"`
					Mobilesmall  string `json:"mobileSmall"`
					Mobilezoom   string `json:"mobileZoom"`
					Mini         string `json:"mini"`
				} `json:"imageMediaUri"`
				Isdefault      bool   `json:"isDefault"`
				Isselected     bool   `json:"isSelected"`
				Istrimmed      bool   `json:"isTrimmed"`
				Mediagrouptype string `json:"mediaGroupType"`
				Mediatype      string `json:"mediaType"`
				Sortid         int    `json:"sortId"`
				//} `json:"5622024"`
			} `json:"byId"`
			Allids []int `json:"allIds"`
		} `json:"styleMedia"`
		Stylenumber string `json:"styleNumber"`
		Stylevideos struct {
			Byid struct {
			} `json:"byId"`
			Allids []interface{} `json:"allIds"`
		} `json:"styleVideos"`
		Treatmenttype   string      `json:"treatmentType"`
		Vendorvideoshot interface{} `json:"vendorVideoShot"`
		Page            struct {
			Pagename   string `json:"PageName"`
			Components []struct {
				ID            string `json:"Id"`
				Componentname string `json:"ComponentName"`
				Schemaname    string `json:"SchemaName"`
				Modules       []struct {
					Heading  string `json:"heading"`
					Text     string `json:"text"`
					Livetext string `json:"liveText,omitempty"`
				} `json:"Modules,omitempty"`
				Sections struct {
					Boflex struct {
						Location    string `json:"Location"`
						Setlocation string `json:"SetLocation"`
					} `json:"Boflex"`
					Freepickupradiobutton struct {
						Choosestore  string `json:"ChooseStore"`
						Freepickup   string `json:"FreePickup"`
						Freepickupat string `json:"FreePickupAt"`
						Notavailable string `json:"NotAvailable"`
					} `json:"FreePickupRadioButton"`
					Freeshippingradiobutton struct {
						Freeshipping string `json:"FreeShipping"`
						Shippingeta  string `json:"ShippingEta"`
					} `json:"FreeShippingRadioButton"`
					Pickuppromises struct {
						Notinstock         string `json:"NotInStock"`
						Tryanotherlocation string `json:"TryAnotherLocation"`
					} `json:"PickUpPromises"`
					Pickupanddeliverymodal struct {
						Freeshipping struct {
							Arrival string `json:"Arrival"`
							Title   string `json:"Title"`
						} `json:"FreeShipping"`
						Pickupsection struct {
							Available               string `json:"Available"`
							Locations               string `json:"Locations"`
							Notavailabledescription string `json:"NotAvailableDescription"`
							Notavailabletitle       string `json:"NotAvailableTitle"`
						} `json:"PickupSection"`
						Submitbutton string `json:"SubmitButton"`
						Title        string `json:"Title"`
						Zipcodeform  struct {
							Description  string `json:"Description"`
							Errormessage string `json:"ErrorMessage"`
							Submitbutton string `json:"SubmitButton"`
							Tryagain     string `json:"TryAgain"`
						} `json:"ZipCodeForm"`
					} `json:"PickupAndDeliveryModal"`
					Promotions struct {
					} `json:"Promotions"`
					Shiptopromises struct {
						Arrival                string `json:"Arrival"`
						Shipping               string `json:"Shipping"`
						Shippingspecialmessage string `json:"ShippingSpecialMessage"`
					} `json:"ShipToPromises"`
					Zipcodeform struct {
						Enteryourzipcode string `json:"EnterYourZipcode"`
						Errormessage     string `json:"ErrorMessage"`
						Tryagainmessage  string `json:"TryAgainMessage"`
					} `json:"ZipCodeForm"`
				} `json:"Sections,omitempty"`
			} `json:"Components"`
		} `json:"page"`
		Isfinalsale                  bool   `json:"isFinalSale"`
		Rmsstylegroupid              string `json:"rmsStyleGroupId"`
		Ispostatboriginalpricehidden bool   `json:"isPostAtbOriginalPriceHidden"`
	} `json:"viewData"`
}

// parseProduct
func (c *_Crawler) parseProduct(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil {
		return nil
	}
	fmt.Println(`parseProduct`)
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	ioutil.WriteFile("C:\\Rinal\\ServiceBasedPRojects\\VoilaWork_new\\VoilaCrawl\\Output.html", respBody, 0644)

	matched := categoryExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		//c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}
	// c.logger.Debugf("data: %s", matched[1])

	var viewData parseProductResponse

	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}
	canUrl := dom.Find(`link[rel="canonical"]`).AttrOr("href", "")
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           strconv.Format(viewData.Viewdata.ID),
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
		},
		CrowdType:   viewData.Viewdata.Gender,
		BrandName:   viewData.Viewdata.Brand.Brandname,
		Title:       viewData.Viewdata.Productname,
		Description: viewData.Viewdata.Description,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stats: &pbItem.Stats{
			ReviewCount: int32(viewData.Viewdata.Numberofreviews),
			Rating:      float32(viewData.Viewdata.Reviewaveragerating),
		},
	}
	// links := viewData.ProductPage.BreadcrumbLinks
	// for i, l := range links {
	// 	switch i {
	// 	case 0:
	// 		item.Category = l.Text
	// 	case 1:
	// 		item.SubCategory = l.Text
	// 	case 2:
	// 		item.SubCategory2 = l.Text
	// 	case 3:
	// 		item.SubCategory3 = l.Text
	// 	case 4:
	// 		item.SubCategory4 = l.Text
	// 	}
	// }

	for _, rawSkuColor := range viewData.Viewdata.Filters.Color.Byid {

		for k, rawNumber := range rawSkuColor.Relatedskuids {
			sizeSkuID := rawNumber
			rawSku := viewData.Viewdata.Skus.Byid[sizeSkuID]

			currentPrice, _ := strconv.ParseFloat(viewData.Viewdata.Skus.Byid[sizeSkuID].Price)
			originalPrice, _ := strconv.ParseFloat(viewData.Viewdata.Skus.Byid[sizeSkuID].Price)
			discount, _ := strconv.ParseFloat(strings.TrimSuffix(viewData.Viewdata.Skus.Byid[sizeSkuID].Displaypercentoff, "%"))
			sku := pbItem.Sku{
				SourceId: strconv.Format(rawSku.ID),
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(currentPrice * 100),
					Msrp:     int32(originalPrice * 100),
					Discount: int32(discount * 100),
				},
				Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}
			if rawSku.Isavailable {
				sku.Stock.StockStatus = pbItem.Stock_InStock
				sku.Stock.StockCount = int32(rawSku.Totalquantityavailable)
			}

			// color
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    strconv.Format(rawSkuColor.ID),
				Name:  rawSkuColor.Displayvalue,
				Value: rawSkuColor.Value,
				Icon:  rawSkuColor.Swatchmedia.Desktop,
			})

			if k == 0 {

				// img based on color
				isDefault := true
				for ki, mid := range rawSkuColor.Stylemediaids {
					rawMedia := viewData.Viewdata.Stylemedia.Byid[mid]

					if ki > 0 {
						isDefault = false
					}

					sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
						strconv.Format(rawMedia.ID),
						rawMedia.Imagemediauri.Mini,
						rawMedia.Imagemediauri.Largedesktop,
						rawMedia.Imagemediauri.Zoom,
						rawMedia.Imagemediauri.Smalldesktop,
						"",
						isDefault,
					))
				}
			}

			// size
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    fmt.Sprintf("%s-%s", rawSku.ID, rawSku.Colorid),
				Name:  viewData.Viewdata.Filters.Size.Byid[rawSku.Sizeid].Value,
				Value: viewData.Viewdata.Filters.Size.Byid[rawSku.Sizeid].Value,
			})

			item.SkuItems = append(item.SkuItems, &sku)
		}
	}
	for _, rawSku := range item.SkuItems {
		if rawSku.Stock.StockStatus == pbItem.Stock_InStock {
			item.Stock = &pbItem.Stock{StockStatus: pbItem.Stock_InStock}
		}
	}
	if item.Stock == nil {
		item.Stock = &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock}
	}
	// yield item result
	return yield(ctx, &item)
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://www.nordstromrack.com",
		//"https://www.nordstromrack.com/s/levis-512-%E2%84%A2-slim-taper-marshmallow-burn-out-jeans/n3298652?color=MARSHMALLOW%20BURNOUT%20DX&eid=482253",
		"https://www.nordstromrack.com/s/free-people-riptide-tie-dye-print-t-shirt/n3327050?color=SEAFOAM%20COMBO",
		// "https://www.nordstromrack.com/shop/Women/Clothing/Tops",
		// "https://www.nordstromrack.com/events/472159",
		// "https://www.nordstromrack.com/shop/Women/Accessories/Hats,%20Gloves%20&%20Scarves/Gloves",
		// "https://www.nordstromrack.com/c/women/clothing/skirts/pencil",
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
	cli.NewApp(&_Crawler{}).Run(os.Args)
}
