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
		categoryPathMatcher: regexp.MustCompile(`^((\?/p/).)*`),
		// this regular used to match product page url path
		productPathMatcher: regexp.MustCompile(`^(.*)(&lvrid=_p)(.*)$`),
		logger:             logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "85f6d7dab5b34606ac89bd11a2016900"
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
	return []string{"*.luisaviaroma.com"}
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

type CategoryData struct {
	Items []struct {
		ItemCode                 string      `json:"ItemCode"`
		SeasonID                 string      `json:"SeasonId"`
		CollectionID             string      `json:"CollectionId"`
		ItemID                   string      `json:"ItemId"`
		VendorColorID            string      `json:"VendorColorId"`
		Description              string      `json:"Description"`
		Designer                 string      `json:"Designer"`
		DesignerID               string      `json:"DesignerId"`
		URL                      string      `json:"Url"`
		Image                    string      `json:"Image"`
		ImageOver                string      `json:"ImageOver"`
		ImageAlternate           string      `json:"ImageAlternate"`
		Sizes                    string      `json:"Sizes"`
		MultiPrice               bool        `json:"MultiPrice"`
		ListPrice                string      `json:"ListPrice"`
		ListPriceDiscounted      string      `json:"ListPriceDiscounted"`
		FinalPrice               interface{} `json:"FinalPrice"`
		Discount                 int         `json:"Discount"`
		PromoReduction           int         `json:"PromoReduction"`
		PromoReductionType       string      `json:"PromoReductionType"`
		PromoReductionPriceLabel interface{} `json:"PromoReductionPriceLabel"`
		Tags                     []struct {
			ID                string      `json:"Id"`
			Description       string      `json:"Description"`
			Class             interface{} `json:"Class"`
			ShowInSite        bool        `json:"ShowInSite"`
			ShowInApp         bool        `json:"ShowInApp"`
			ShowInStorage     bool        `json:"ShowInStorage"`
			ShowInUserCluster bool        `json:"ShowInUserCluster"`
			UserTags          interface{} `json:"UserTags"`
		} `json:"Tags"`
		Badges    []interface{} `json:"Badges"`
		Note      interface{}   `json:"Note"`
		ExtraInfo struct {
			Available    string      `json:"Available"`
			Size         string      `json:"Size"`
			Colors       interface{} `json:"Colors"`
			PriceTooltip string      `json:"PriceTooltip"`
		} `json:"ExtraInfo"`
		Variants       []interface{} `json:"Variants"`
		Section        string        `json:"Section"`
		UniqueID       string        `json:"UniqueId"`
		ItemParameters struct {
			ItemCode      string `json:"ItemCode"`
			SeasonID      string `json:"SeasonId"`
			CollectionID  string `json:"CollectionId"`
			ItemID        string `json:"ItemId"`
			VendorColorID string `json:"VendorColorId"`
			SeasonMemo    string `json:"SeasonMemo"`
			GenderMemo    string `json:"GenderMemo"`
		} `json:"ItemParameters"`
		IsMadeToMeasure bool `json:"IsMadeToMeasure"`
		OfferMetaInfo   struct {
			Availability string `json:"Availability"`
			Currency     string `json:"Currency"`
			FinalPrice   string `json:"FinalPrice"`
		} `json:"OfferMetaInfo"`
		PrivateSaleInfoList interface{} `json:"PrivateSaleInfoList"`
	} `json:"Items"`
	Pagination struct {
		TotalePages   int         `json:"TotalePages"`
		TotaleRecords int         `json:"TotaleRecords"`
		CurrentPage   int         `json:"CurrentPage"`
		URLTemplate   string      `json:"UrlTemplate"`
		SeoNext       string      `json:"SeoNext"`
		SeoPrev       interface{} `json:"SeoPrev"`
		FirstPageURL  string      `json:"FirstPageUrl"`
	} `json:"Pagination"`
}

// used to extract embaded json data in website page.
// more about golang regulation see here https://golang.org/pkg/regexp/syntax/
var productListExtractReg = regexp.MustCompile(`({.*})`)
var productsExtractReg = regexp.MustCompile(`window\.__BODY_MODEL__\s*=\s*({.*});`)

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

	matched := productListExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}
	// c.logger.Debugf("data: %s", matched[1])

	var viewData CategoryData
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		return err
	}

	lastIndex := nextIndex(ctx)
	for _, idv := range viewData.Items {

		fmt.Println(idv.URL)
		req, err := http.NewRequest(http.MethodGet, idv.URL, nil)
		if err != nil {
			c.logger.Errorf("load http request of url %s failed, error=%s", idv.URL, err)
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
	// check if this is the last page
	if lastIndex > viewData.Pagination.TotaleRecords || page >= int64(viewData.Pagination.TotalePages) {
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

type ProductPageData struct {
	ItemKey struct {
		ItemCode string `json:"ItemCode"`
	} `json:"ItemKey"`
	AllItemKeys []struct {
		ItemCode string `json:"ItemCode"`
	} `json:"AllItemKeys"`
	ItemCode                 string      `json:"ItemCode"`
	IsPrivateSale            bool        `json:"IsPrivateSale"`
	IsExclusiveSale          bool        `json:"IsExclusiveSale"`
	PrivateSaleNeedEmpty     bool        `json:"PrivateSaleNeedEmpty"`
	URLToItemDetailsUserLang interface{} `json:"UrlToItemDetailsUserLang"`
	PrivateSaleLink          interface{} `json:"PrivateSaleLink"`
	MultiPrice               bool        `json:"MultiPrice"`
	SizeUnique               bool        `json:"SizeUnique"`
	SizeShow                 bool        `json:"SizeShow"`
	IsPrimaryContext         bool        `json:"IsPrimaryContext"`
	Availability             []struct {
		SizeKey           string `json:"SizeKey"`
		SizeTypeID        string `json:"SizeTypeId"`
		SizeValue         string `json:"SizeValue"`
		SizeCorrValue     string `json:"SizeCorrValue"`
		SizeOrd           int    `json:"SizeOrd"`
		ColorAvailability []struct {
			ComColorID           int    `json:"ComColorId"`
			ComColorDescription  string `json:"ComColorDescription"`
			VendorColorID        string `json:"VendorColorId"`
			EncodedVendorColorID string `json:"EncodedVendorColorId"`
			SampleColorPhoto     string `json:"SampleColorPhoto"`
			ID                   string `json:"Id"`
			Description          string `json:"Description"`
			QuantitiesTotal      struct {
				Available    int    `json:"Available"`
				PreOrder     int    `json:"PreOrder"`
				PreOrderDate string `json:"PreOrderDate"`
				Max          int    `json:"Max"`
				Bag          int    `json:"Bag"`
			} `json:"QuantitiesTotal"`
			QuantitiesByItemCode struct {
				Seven3IGH4011 struct {
					Available    int    `json:"Available"`
					PreOrder     int    `json:"PreOrder"`
					PreOrderDate string `json:"PreOrderDate"`
					Max          int    `json:"Max"`
					Bag          int    `json:"Bag"`
				} `json:"73I-GH4011"`
			} `json:"QuantitiesByItemCode"`
			HasColorThrottling bool `json:"HasColorThrottling"`
		} `json:"ColorAvailability"`
		Pricing struct {
			ItemKey struct {
				ItemCode string `json:"ItemCode"`
			} `json:"ItemKey"`
			Prices []struct {
				CurrencyID      string `json:"CurrencyId"`
				ListPrice       int    `json:"ListPrice"`
				DiscountedPrice int    `json:"DiscountedPrice"`
				FinalPrice      int    `json:"FinalPrice"`
				FinalPriceNoVat int    `json:"FinalPriceNoVat"`
			} `json:"Prices"`
			Discount      int    `json:"Discount"`
			DiscountPromo int    `json:"DiscountPromo"`
			PromoID       string `json:"PromoId"`
			NotePromo     string `json:"NotePromo"`
		} `json:"Pricing"`
		ID                  string `json:"Id"`
		Description         string `json:"Description"`
		PlainDescription    string `json:"PlainDescription"`
		SelectedDescription string `json:"SelectedDescription"`
	} `json:"Availability"`
	AvailabilityByColor []struct {
		ID                   string `json:"Id"`
		Description          string `json:"Description"`
		ComColorID           int    `json:"ComColorId"`
		ComColorDescription  string `json:"ComColorDescription"`
		VendorColorID        string `json:"VendorColorId"`
		EncodedVendorColorID string `json:"EncodedVendorColorId"`
		SampleColorPhoto     string `json:"SampleColorPhoto"`
		HasColorThrottling   bool   `json:"HasColorThrottling"`
		SizeAvailability     []struct {
			SizeKey              string `json:"SizeKey"`
			SizeTypeID           string `json:"SizeTypeId"`
			SizeValue            string `json:"SizeValue"`
			SizeCorrValue        string `json:"SizeCorrValue"`
			SizeOrd              int    `json:"SizeOrd"`
			ComColorID           int    `json:"ComColorId"`
			ComColorDescription  string `json:"ComColorDescription"`
			VendorColorID        string `json:"VendorColorId"`
			EncodedVendorColorID string `json:"EncodedVendorColorId"`
			SampleColorPhoto     string `json:"SampleColorPhoto"`
			Pricing              struct {
				ItemKey struct {
					ItemCode string `json:"ItemCode"`
				} `json:"ItemKey"`
				Prices []struct {
					CurrencyID      string `json:"CurrencyId"`
					ListPrice       int    `json:"ListPrice"`
					DiscountedPrice int    `json:"DiscountedPrice"`
					FinalPrice      int    `json:"FinalPrice"`
					FinalPriceNoVat int    `json:"FinalPriceNoVat"`
				} `json:"Prices"`
				Discount      int    `json:"Discount"`
				DiscountPromo int    `json:"DiscountPromo"`
				PromoID       string `json:"PromoId"`
				NotePromo     string `json:"NotePromo"`
			} `json:"Pricing"`
			ID                  string `json:"Id"`
			Description         string `json:"Description"`
			PlainDescription    string `json:"PlainDescription"`
			SelectedDescription string `json:"SelectedDescription"`
			QuantitiesTotal     struct {
				Available    int    `json:"Available"`
				PreOrder     int    `json:"PreOrder"`
				PreOrderDate string `json:"PreOrderDate"`
				Max          int    `json:"Max"`
				Bag          int    `json:"Bag"`
			} `json:"QuantitiesTotal"`
			QuantitiesByItemCode struct {
				Seven3IGH4011 struct {
					Available    int    `json:"Available"`
					PreOrder     int    `json:"PreOrder"`
					PreOrderDate string `json:"PreOrderDate"`
					Max          int    `json:"Max"`
					Bag          int    `json:"Bag"`
				} `json:"73I-GH4011"`
			} `json:"QuantitiesByItemCode"`
			HasColorThrottling bool `json:"HasColorThrottling"`
		} `json:"SizeAvailability"`
	} `json:"AvailabilityByColor"`
	Details []struct {
		//Type    string      `json:"Type"`
		Text string `json:"Text"`
		//Link    interface{} `json:"Link"`
		SubList []string `json:"SubList"`
	} `json:"Details"`
	SustainableDetails             []interface{}       `json:"SustainableDetails"`
	Ingredients                    interface{}         `json:"Ingredients"`
	IngredientsAndNutritionalInfos interface{}         `json:"IngredientsAndNutritionalInfos"`
	PhotosAll                      []string            `json:"PhotosAll"`
	PhotosContext                  []string            `json:"PhotosContext"`
	PhotosByColor                  map[string][]string `json:"PhotosByColor"`
	//One122124 []string `json:"112|2124"`
	PhotoPath            string `json:"PhotoPath"`
	PhotoPathBig         string `json:"PhotoPathBig"`
	PhotoPathBigRetina   string `json:"PhotoPathBigRetina"`
	PhotoPathZoom        string `json:"PhotoPathZoom"`
	PhotoRetinaAvailable bool   `json:"PhotoRetinaAvailable"`
	PhotoFirst           string `json:"PhotoFirst"`
	PhotoFirstAlt        string `json:"PhotoFirstAlt"`
	ImageOver            string `json:"ImageOver"`
	ItemCodeDetails      struct {
		Seven3IGH4011 struct {
			Tag                      string `json:"Tag"`
			TagID                    string `json:"TagId"`
			Discounted               bool   `json:"Discounted"`
			Discount                 int    `json:"Discount"`
			PriceList                string `json:"PriceList"`
			PriceDiscounted          string `json:"PriceDiscounted"`
			PriceListLabel           string `json:"PriceListLabel"`
			PriceDiscountedLabel     string `json:"PriceDiscountedLabel"`
			FinalPrice               string `json:"FinalPrice"`
			InvoiceFinalPriceValue   int    `json:"InvoiceFinalPriceValue"`
			InvoiceFinalPrice        string `json:"InvoiceFinalPrice"`
			PromoReduction           string `json:"PromoReduction"`
			PromoReductionPriceLabel string `json:"PromoReductionPriceLabel"`
			PromoReductionType       string `json:"PromoReductionType"`
		} `json:"73I-GH4011"`
	} `json:"ItemCodeDetails"`
	DesignerID              string `json:"DesignerId"`
	DesignerText            string `json:"DesignerText"`
	DesignerLink            string `json:"DesignerLink"`
	DesignerCorrectCaseText string `json:"DesignerCorrectCaseText"`
	DesignerParameters      struct {
		Gender   string `json:"Gender"`
		Season   string `json:"Season"`
		Designer string `json:"Designer"`
	} `json:"DesignerParameters"`
	DescriptionText       string `json:"DescriptionText"`
	DescriptionLink       string `json:"DescriptionLink"`
	DescriptionParameters struct {
		Gender   string `json:"Gender"`
		Season   string `json:"Season"`
		Subline  string `json:"Subline"`
		Category string `json:"Category"`
	} `json:"DescriptionParameters"`
	ShowSizeChart             bool     `json:"ShowSizeChart"`
	SizeChartID               string   `json:"SizeChartId"`
	SizeDescription           string   `json:"SizeDescription"`
	SizeCountryDescr          string   `json:"SizeCountryDescr"`
	SublineMemoCode           string   `json:"SublineMemoCode"`
	SublineEnglishDescription string   `json:"SublineEnglishDescription"`
	ItemTags                  []string `json:"ItemTags"`
	Tags                      []struct {
		ID                string      `json:"Id"`
		Description       string      `json:"Description"`
		Class             interface{} `json:"Class"`
		ShowInSite        bool        `json:"ShowInSite"`
		ShowInApp         bool        `json:"ShowInApp"`
		ShowInStorage     bool        `json:"ShowInStorage"`
		ShowInUserCluster bool        `json:"ShowInUserCluster"`
		UserTags          interface{} `json:"UserTags"`
	} `json:"Tags"`
	Badges                      []interface{} `json:"Badges"`
	Discount                    int           `json:"Discount"`
	CategoryEnglishDescription  string        `json:"CategoryEnglishDescription"`
	CategoryLangDescription     string        `json:"CategoryLangDescription"`
	ShortDescription            string        `json:"ShortDescription"`
	EditorNote                  string        `json:"EditorNote"`
	SizeCorrDescr               string        `json:"SizeCorrDescr"`
	SizeCorrCountryDescr        string        `json:"SizeCorrCountryDescr"`
	SizeTypeDescrCorrID         int           `json:"SizeTypeDescrCorrID"`
	SizeTypeDescrSrcID          int           `json:"SizeTypeDescrSrcID"`
	ShowDoubleColumnSizeSelect  bool          `json:"ShowDoubleColumnSizeSelect"`
	ShowDoubleColumnColorSelect bool          `json:"ShowDoubleColumnColorSelect"`
	ShareEnabled                bool          `json:"ShareEnabled"`
	ShareFacebook               string        `json:"ShareFacebook"`
	ShareTwitter                string        `json:"ShareTwitter"`
	AntavoShareFacebook         string        `json:"AntavoShareFacebook"`
	AntavoShareTwitter          string        `json:"AntavoShareTwitter"`
	AntavoShareTwitterMessage   string        `json:"AntavoShareTwitterMessage"`
	ShareGooglePlus             string        `json:"ShareGooglePlus"`
	ShareWeibo                  string        `json:"ShareWeibo"`
	SharePinterest              string        `json:"SharePinterest"`
	BreadcrumbEnabled           bool          `json:"BreadcrumbEnabled"`
	BreadcrumbGender            string        `json:"BreadcrumbGender"`
	BreadcrumbGenderURL         string        `json:"BreadcrumbGenderUrl"`
	BreadcrumbGenderParameters  struct {
		Gender string `json:"Gender"`
		Season string `json:"Season"`
	} `json:"BreadcrumbGenderParameters"`
	BreadcrumbSubline           string `json:"BreadcrumbSubline"`
	BreadcrumbSublineURL        string `json:"BreadcrumbSublineUrl"`
	BreadcrumbSublineParameters struct {
		Gender  string `json:"Gender"`
		Season  string `json:"Season"`
		Subline string `json:"Subline"`
	} `json:"BreadcrumbSublineParameters"`
	BreadcrumbCategory           string `json:"BreadcrumbCategory"`
	BreadcrumbCategoryURL        string `json:"BreadcrumbCategoryUrl"`
	BreadcrumbCategoryParameters struct {
		Gender   string `json:"Gender"`
		Season   string `json:"Season"`
		Subline  string `json:"Subline"`
		Category string `json:"Category"`
	} `json:"BreadcrumbCategoryParameters"`
	BreadcrumbDescription string `json:"BreadcrumbDescription"`
	BreadcrumbOwnLink     string `json:"BreadcrumbOwnLink"`
	SizeTypeDescrID       int    `json:"SizeTypeDescrId"`
	SizeLabel             string `json:"SizeLabel"`
	IsBeautyByDefinition  bool   `json:"IsBeautyByDefinition"`
	URLByColor            struct {
		One122124 string `json:"112|2124"`
	} `json:"UrlByColor"`
	TitlesByColor struct {
		One122124 string `json:"112|2124"`
	} `json:"TitlesByColor"`
	ModelSampleSize interface{} `json:"ModelSampleSize"`
	ModelGeneralFit interface{} `json:"ModelGeneralFit"`
	Detail          struct {
		Tag                      string `json:"Tag"`
		TagID                    string `json:"TagId"`
		Discounted               bool   `json:"Discounted"`
		Discount                 int    `json:"Discount"`
		PriceList                string `json:"PriceList"`
		PriceDiscounted          string `json:"PriceDiscounted"`
		PriceListLabel           string `json:"PriceListLabel"`
		PriceDiscountedLabel     string `json:"PriceDiscountedLabel"`
		FinalPrice               string `json:"FinalPrice"`
		InvoiceFinalPriceValue   int    `json:"InvoiceFinalPriceValue"`
		InvoiceFinalPrice        string `json:"InvoiceFinalPrice"`
		PromoReduction           string `json:"PromoReduction"`
		PromoReductionPriceLabel string `json:"PromoReductionPriceLabel"`
		PromoReductionType       string `json:"PromoReductionType"`
	} `json:"Detail"`
	PageTitle   string `json:"PageTitle"`
	MetaSharing []struct {
		Type            string      `json:"Type,omitempty"`
		SiteName        string      `json:"SiteName,omitempty"`
		FbPageID        string      `json:"FbPageId,omitempty"`
		Title           string      `json:"Title"`
		Description     string      `json:"Description"`
		URL             string      `json:"Url"`
		Image           string      `json:"Image"`
		ImageWidth      interface{} `json:"ImageWidth"`
		ImageHeight     interface{} `json:"ImageHeight"`
		Card            string      `json:"Card,omitempty"`
		AccountID       string      `json:"AccountId,omitempty"`
		Site            string      `json:"Site,omitempty"`
		ID              interface{} `json:"Id,omitempty"`
		Brand           string      `json:"Brand,omitempty"`
		SKU             string      `json:"SKU,omitempty"`
		PotentialAction interface{} `json:"PotentialAction,omitempty"`
		Offers          struct {
			Type           string      `json:"Type"`
			URL            string      `json:"Url"`
			PriceCurrency  string      `json:"PriceCurrency"`
			Price          string      `json:"Price"`
			Seller         interface{} `json:"Seller"`
			Availability   string      `json:"Availability"`
			EligibleRegion string      `json:"EligibleRegion"`
		} `json:"Offers,omitempty"`
		SameAs        interface{} `json:"SameAs,omitempty"`
		AlternateName interface{} `json:"AlternateName,omitempty"`
		ContactPoint  interface{} `json:"ContactPoint,omitempty"`
		Logo          interface{} `json:"Logo,omitempty"`
	} `json:"MetaSharing"`
	TrackingInfo struct {
		PageTitle                    string   `json:"PageTitle"`
		ProductURL                   string   `json:"ProductUrl"`
		URLEN                        string   `json:"UrlEN"`
		PageSubline                  string   `json:"PageSubline"`
		PageCategory                 string   `json:"PageCategory"`
		PageDesigner                 string   `json:"PageDesigner"`
		ProductLine                  string   `json:"ProductLine"`
		ProductSubline               string   `json:"ProductSubline"`
		ProductCategory              string   `json:"ProductCategory"`
		ProductDesigner              string   `json:"ProductDesigner"`
		ProductID                    string   `json:"ProductId"`
		ProductName                  string   `json:"ProductName"`
		ProductTags                  []string `json:"ProductTags"`
		ProductGenderMemoCode        string   `json:"ProductGenderMemoCode"`
		ProductPrimaryGenderMemoCode string   `json:"ProductPrimaryGenderMemoCode"`
		ProductCurrencyID            string   `json:"ProductCurrencyId"`
		ProductFinalPrice            int      `json:"ProductFinalPrice"`
		ProductDiscount              int      `json:"ProductDiscount"`
		ProductExtraDiscount         int      `json:"ProductExtraDiscount"`
		ProductCurrencyIDBill        string   `json:"ProductCurrencyIdBill"`
		ProductFinalPriceBill        int      `json:"ProductFinalPriceBill"`
		ProductIsInStock             bool     `json:"ProductIsInStock"`
	} `json:"TrackingInfo"`
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

	var viewData ProductPageData

	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}

	prodid := viewData.ItemCode
	indexMetaSharing := 0

	for i, raw := range viewData.MetaSharing {
		if raw.Type == "Product" && raw.Brand != "" {
			indexMetaSharing = i
		}
	}
	descriptions := ""
	for _, desc := range viewData.Details {
		descriptions = descriptions + " " + desc.Text + " " + strings.Join(desc.SubList, " ")
	}

	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:       strconv.Format(prodid),
			CrawlUrl: resp.Request.URL.String(),
		},
		BrandName:   viewData.MetaSharing[indexMetaSharing].Brand,
		Title:       viewData.ShortDescription,
		Description: descriptions,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
	}

	for _, rawcolor := range viewData.AvailabilityByColor {

		for ks, rawSku := range rawcolor.SizeAvailability {

			originalPrice := (rawSku.Pricing.Prices[0].FinalPrice)
			msrp := (rawSku.Pricing.Prices[0].ListPrice)
			discount := (rawSku.Pricing.Discount)
			sku := pbItem.Sku{
				SourceId: strconv.Format(rawSku.ID),
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(originalPrice * 100),
					Msrp:     int32(msrp * 100),
					Discount: int32(discount),
				},
				Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}
			if rawSku.QuantitiesTotal.Available > 0 {
				sku.Stock.StockStatus = pbItem.Stock_InStock
				sku.Stock.StockCount = int32(rawSku.QuantitiesTotal.Available)
			}

			// color
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    strconv.Format(rawcolor.ComColorID),
				Name:  rawcolor.Description,
				Value: rawcolor.Description,
				//Icon:  color.SwatchMedia.Mobile,
			})

			if ks == 0 {
				colorcode := string(rawcolor.ComColorID) + "|" + rawcolor.VendorColorID
				isDefault := true
				for i, mid := range viewData.PhotosByColor[colorcode] {
					if i > 0 {
						isDefault = false
					}
					sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
						strconv.Format(i),
						"https://images.lvrcdn.com/zoom"+mid,
						"https://images.lvrcdn.com/zoom"+mid,
						"https://images.lvrcdn.com/zoom"+mid,
						"https://images.lvrcdn.com/Big"+mid,
						"",
						isDefault,
					))
				}
			}

			// size
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    strconv.Format(rawSku.ComColorID) + "_" + strconv.Format(ks),
				Name:  rawSku.Description + "_" + rawSku.SizeCorrValue,
				Value: rawSku.Description + "_" + rawSku.SizeCorrValue,
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

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		"https://www.luisaviaroma.com/en-us/shop/women/bags?lvrid=_gw_i22_s&Page=2&ajax=true",
		"https://www.luisaviaroma.com/en-us/p/the-attico/women/top-handle-bags/72I-RSI001?ColorId=MTAw0&SubLine=bags&CategoryId=81&lvrid=_p_d4UN_gw_c81",
		"https://www.luisaviaroma.com/72I-0LT009?ColorId=MDBC0&lvrid=_p",
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

// local test
func main() {
	logger := glog.New(glog.LogLevelDebug)
	// build a http client
	// get proxy's microservice address from env
	//os.Setenv("VOILA_PROXY_URL", "http://3.239.93.53:30216")
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
				i.URL.Host = "www.luisaviaroma.com"
			}

			// do http requests here.
			nctx, cancel := context.WithTimeout(ctx, time.Minute*5)
			defer cancel()
			resp, err := client.DoWithOptions(nctx, i, http.Options{
				EnableProxy:       true,
				EnableHeadless:    true,
				EnableSessionInit: true,
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
