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

	"github.com/voiladev/VoilaCrawl/pkg/crawler"
	"github.com/voiladev/VoilaCrawl/pkg/net/http"
	"github.com/voiladev/VoilaCrawl/pkg/net/http/cookiejar"
	"github.com/voiladev/VoilaCrawl/pkg/proxy"
	pbMedia "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	pbProxy "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// _Crawler defined the crawler struct/class for which is not necessory to be exportable
type _Crawler struct {
	// httpClient is the object of an http client
	httpClient             http.Client
	categoryPathMatcher    *regexp.Regexp
	categoryApiPathMatcher *regexp.Regexp
	productPathMatcher     *regexp.Regexp
	productApiPathMatcher  *regexp.Regexp
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
		categoryPathMatcher:    regexp.MustCompile(`^(/collections(/[a-z0-9\-]+){1,6})$`),
		categoryApiPathMatcher: regexp.MustCompile(`^(/api/v3/collections(/[a-z0-9\-]+){1,6})$`),

		// this regular used to match product page url path
		productPathMatcher:    regexp.MustCompile(`^(/products(/[a-z0-9\-]+){1,6})$`),
		productApiPathMatcher: regexp.MustCompile(`^(/api/v2/product_groups(/[a-z0-9\-]+){0,6})$`),
		logger:                logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "fbd0d81c0b6340618187f0b0b417a11f"
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
	opts := &crawler.CrawlOptions{
		EnableHeadless:    false,
		EnableSessionInit: true,
		Reliability:       pbProxy.ProxyReliability_ReliabilityDefault,
	}
	opts.MustCookies = append(opts.MustCookies,
		&http.Cookie{Name: "country_data", Value: "US~en", Path: "/"},
		&http.Cookie{Name: "backoptinpopin2", Value: "0", Path: "/"},
	)
	return opts
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.everlane.com"}
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

	if c.productPathMatcher.MatchString(resp.Request.URL.Path) || c.productApiPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yield)
	} else if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) || c.categoryApiPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	}
	return fmt.Errorf("unsupported url %s", resp.Request.URL.String())
}

// nextIndex used to get the index from the shared data.
// item.index is a const key for item index.
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")))
}

type categoryParse struct {
	ID                            int         `json:"id"`
	Gender                        string      `json:"gender"`
	Disabled                      bool        `json:"disabled"`
	Notification                  string      `json:"notification"`
	Permalink                     string      `json:"permalink"`
	ProductsPerRow                int         `json:"products_per_row"`
	Title                         string      `json:"title"`
	UpdatedAt                     time.Time   `json:"updated_at"`
	Description                   string      `json:"description"`
	ShowSubnav                    bool        `json:"show_subnav"`
	SitemapVisible                bool        `json:"sitemap_visible"`
	BreadcrumbTitle               string      `json:"breadcrumb_title"`
	TitleTag                      string      `json:"title_tag"`
	DisplayGroupAspectRatio       string      `json:"display_group_aspect_ratio"`
	DesktopContentPageID          interface{} `json:"desktop_content_page_id"`
	MobileContentPageID           interface{} `json:"mobile_content_page_id"`
	DesktopMarketingContentPageID interface{} `json:"desktop_marketing_content_page_id"`
	MobileMarketingContentPageID  interface{} `json:"mobile_marketing_content_page_id"`
	DesktopFooterContentPageID    interface{} `json:"desktop_footer_content_page_id"`
	MobileFooterContentPageID     interface{} `json:"mobile_footer_content_page_id"`
	DisabledDesktopContentPageID  interface{} `json:"disabled_desktop_content_page_id"`
	DisabledMobileContentPageID   interface{} `json:"disabled_mobile_content_page_id"`
	Groupings                     struct {
		DisplayGroup []struct {
			ID                       int           `json:"id"`
			Name                     string        `json:"name"`
			Description              string        `json:"description"`
			DesktopContentPageID     interface{}   `json:"desktop_content_page_id"`
			MobileContentPageID      interface{}   `json:"mobile_content_page_id"`
			ProductGridContentPageID interface{}   `json:"product_grid_content_page_id"`
			BuilderEditorialTileKey  interface{}   `json:"builder_editorial_tile_key"`
			Products                 []int         `json:"products"`
			ProductPermalinks        []string      `json:"product_permalinks"`
			DesktopProducts          []int         `json:"desktop_products"`
			DesktopProductPermalinks []string      `json:"desktop_product_permalinks"`
			MobileProducts           []int         `json:"mobile_products"`
			MobileProductPermalinks  []string      `json:"mobile_product_permalinks"`
			Platforms                []string      `json:"platforms"`
			BuilderBlocks            []interface{} `json:"builder_blocks"`
		} `json:"display_group"`
		ProductGroup []struct {
			ID                int           `json:"id"`
			Name              string        `json:"name"`
			Label             string        `json:"label"`
			Products          []int         `json:"products"`
			ProductPermalinks []string      `json:"product_permalinks"`
			ProductColorOrder []interface{} `json:"product_color_order"`
		} `json:"product_group"`
	} `json:"groupings"`
	Products []struct {
		ID                  int           `json:"id"`
		Permalink           string        `json:"permalink"`
		ProductGroupID      int           `json:"product_group_id"`
		OrderableState      string        `json:"orderable_state"`
		DisplayName         string        `json:"display_name"`
		GenderedDisplayName string        `json:"gendered_display_name"`
		Price               string        `json:"price"`
		ChooseWhatYouPay    bool          `json:"choose_what_you_pay"`
		ReviewedAt          time.Time     `json:"reviewed_at"`
		ProductFits         []interface{} `json:"product_fits"`
		Retired             bool          `json:"retired"`
		MainImage           string        `json:"main_image"`
		OriginalPrice       float64       `json:"original_price"`
		IsPromo             bool          `json:"is_promo"`
		FinalSale           bool          `json:"final_sale"`
		Color               struct {
			Name      string `json:"name"`
			HexValue  string `json:"hex_value"`
			HexValue2 string `json:"hex_value_2"`
		} `json:"color"`
		BaseColor struct {
			Name      string      `json:"name"`
			HexValue  string      `json:"hex_value"`
			HexValue2 interface{} `json:"hex_value_2"`
		} `json:"base_color"`
		Albums struct {
			Square []struct {
				Src string      `json:"src"`
				Tag interface{} `json:"tag"`
			} `json:"square"`
		} `json:"albums,omitempty"`
		PrimaryCollection struct {
			BreadcrumbTitle string `json:"breadcrumb_title"`
			Permalink       string `json:"permalink"`
		} `json:"primary_collection"`
		SizeChart struct {
			Content   string      `json:"content"`
			MainImage interface{} `json:"main_image"`
		} `json:"size_chart"`
		Variants []struct {
			ID                  int    `json:"id"`
			Sku                 string `json:"sku"`
			Upc                 string `json:"upc"`
			Available           int    `json:"available"`
			OrderableState      string `json:"orderable_state"`
			InventoryCount      int    `json:"inventory_count"`
			Size                string `json:"size"`
			AbbreviatedSize     string `json:"abbreviated_size"`
			Annotation          string `json:"annotation"`
			Name                string `json:"name"`
			RestockDate         string `json:"restock_date"`
			LaunchDate          string `json:"launch_date"`
			FulfillmentCenterID int    `json:"fulfillment_center_id"`
			IsDigitalGiftcard   bool   `json:"is_digital_giftcard"`
			SingleReturnVariant bool   `json:"single_return_variant"`
			MetaGarment         string `json:"meta_garment"`
		} `json:"variants"`
		CwypDiscounts []float64 `json:"cwyp_discounts"`
	} `json:"products"`
	Metadata      []interface{} `json:"metadata"`
	Subcategories []struct {
		ID            int      `json:"id"`
		Name          string   `json:"name"`
		DisplayGroups []string `json:"display_groups"`
	} `json:"subcategories"`
	CollectionNavigationItems []struct {
		ID               int    `json:"id"`
		CollectionID     int    `json:"collection_id"`
		Label            string `json:"label"`
		ImageURL         string `json:"image_url"`
		PermalinkType    string `json:"permalink_type"`
		PermalinkID      int    `json:"permalink_id"`
		CollectionFilter string `json:"collection_filter"`
		DisplayFormat    string `json:"display_format"`
		Position         int    `json:"position"`
		Permalink        string `json:"permalink"`
	} `json:"collection_navigation_items"`
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if strings.ToLower(v) == strings.ToLower(str) {
			return true
		} else if strings.Contains(strings.ToLower(str), strings.ToLower(v)) {
			return true
		}
	}
	return false
}
func containsInt(s []int, str int) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
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

	// doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	// if err != nil {
	// 	return err
	// }

	if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		//matched := productsExtractReg.FindSubmatch([]byte(resp.Request.URL.Path))
		produrl := strings.ReplaceAll(resp.Request.URL.String(), "/collections", "/api/v3/collections")

		req, err := http.NewRequest(http.MethodGet, produrl, nil)
		req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
		req.Header.Set("Referer", fmt.Sprintf("%s://%s/%s", resp.Request.URL.Scheme, resp.Request.URL.Host, resp.Request.URL.Path))
		req.Header.Set("x-requested-with", "XMLHttpRequest")
		req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.114 Safari/537.36")

		respNew, err := c.httpClient.Do(ctx, req)
		if err != nil {
			panic(err)
		}

		respBody, err = ioutil.ReadAll(respNew.Body)
		if err != nil {
			return err
		}
	}

	if !bytes.Contains(respBody, []byte(`products`)) {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract category info from %s failed, error=%s", resp.Request.URL, err)
	}

	var displaygroup []string
	var listproducts []int
	if strings.Contains(resp.Request.URL.String(), "style=") {
		styleList := strings.Split(strings.ReplaceAll(resp.Request.URL.String(), "&style=", "style="), "style=")
		for i, value := range styleList {
			if i > 0 {
				if !contains(displaygroup, value) {
					strdecode, _ := url.QueryUnescape(value)
					displaygroup = append(displaygroup, strdecode)
				}
			}
		}
	}

	var viewData categoryParse
	if err := json.Unmarshal(respBody, &viewData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}

	if len(displaygroup) == 0 {
		for _, value := range viewData.Groupings.ProductGroup {
			listproducts = append(listproducts, value.Products...)
		}
	} else {
		for _, value := range viewData.Groupings.DisplayGroup {
			if contains(displaygroup, value.Name) {
				listproducts = append(listproducts, value.Products...)
			}
		}
	}

	lastIndex := nextIndex(ctx)

	for _, value := range viewData.Products {

		if !containsInt(listproducts, value.ID) {
			continue
		}

		rawurl := "https://www.everlane.com/api/v2/product_groups?product_permalink=" + value.Permalink
		fmt.Println(rawurl)
		req, err := http.NewRequest(http.MethodGet, rawurl, nil)
		if err != nil {
			c.logger.Error(err)
			continue
		}

		lastIndex += 1
		nctx := context.WithValue(ctx, "item.index", lastIndex)
		if err := yield(nctx, req); err != nil {
			return err
		}

	}

	return nil
	// matched := nextPageReg.FindStringSubmatch(string(respBody))
	// if len(matched) < 2 {
	// 	return nil
	// }
	// nexturl := matched[1]
	// req, _ := http.NewRequest(http.MethodGet, nexturl, nil)
	// // update the index of last page
	// nctx := context.WithValue(ctx, "item.index", lastIndex)
	// return yield(nctx, req)
}

type parseProductData struct {
	Products []struct {
		ID                   int           `json:"id"`
		Permalink            string        `json:"permalink"`
		ProductGroupID       int           `json:"product_group_id"`
		OrderableState       string        `json:"orderable_state"`
		DesktopContentPageID int           `json:"desktop_content_page_id"`
		MobileContentPageID  int           `json:"mobile_content_page_id"`
		InfographicID        int           `json:"infographic_id"`
		ProductFits          []interface{} `json:"product_fits"`
		Retired              bool          `json:"retired"`
		MainImage            string        `json:"main_image"`
		FlatImage            string        `json:"flat_image"`
		DisplayName          string        `json:"display_name"`
		GenderedDisplayName  string        `json:"gendered_display_name"`
		Price                string        `json:"price"`
		OriginalPrice        float64       `json:"original_price"`
		Gender               string        `json:"gender"`
		DisclaimerTitle      interface{}   `json:"disclaimer_title"`
		DisclaimerBody       interface{}   `json:"disclaimer_body"`
		ChooseWhatYouPay     bool          `json:"choose_what_you_pay"`
		IsPromo              bool          `json:"is_promo"`
		PromotionMessages    []interface{} `json:"promotion_messages"`
		Color                struct {
			Name      string      `json:"name"`
			HexValue  string      `json:"hex_value"`
			HexValue2 interface{} `json:"hex_value_2"`
		} `json:"color"`
		BaseColor struct {
			Name      string      `json:"name"`
			HexValue  string      `json:"hex_value"`
			HexValue2 interface{} `json:"hex_value_2"`
		} `json:"base_color"`
		ProductVideo interface{} `json:"product_video"`
		Albums       struct {
			Portrait []interface{} `json:"portrait"`
			Square   []struct {
				Src string `json:"src"`
				Tag string `json:"tag"`
			} `json:"square"`
		} `json:"albums"`
		PrimaryCollection struct {
			ID              int    `json:"id"`
			Permalink       string `json:"permalink"`
			Title           string `json:"title"`
			BreadcrumbTitle string `json:"breadcrumb_title"`
			Gender          string `json:"gender"`
			Subcategories   []struct {
				ID           int       `json:"id"`
				CollectionID int       `json:"collection_id"`
				Name         string    `json:"name"`
				Active       bool      `json:"active"`
				Position     int       `json:"position"`
				CreatedAt    time.Time `json:"created_at"`
				UpdatedAt    time.Time `json:"updated_at"`
			} `json:"subcategories"`
		} `json:"primary_collection"`
		RelatedProductLink string      `json:"related_product_link"`
		FitScale           interface{} `json:"fit_scale"`
		PreLaunchPolicy    bool        `json:"pre_launch_policy"`
		CanShowReviews     bool        `json:"can_show_reviews"`
		TraditionalPrice   int         `json:"traditional_price"`
		SizeChart          struct {
			Content   string      `json:"content"`
			Caption   string      `json:"caption"`
			MainImage interface{} `json:"main_image"`
		} `json:"size_chart"`
		BodySizeChart struct {
			Content [][]string `json:"content"`
			Name    string     `json:"name"`
		} `json:"body_size_chart"`
		InternationalSizeChart struct {
			Content [][]string `json:"content"`
			Name    string     `json:"name"`
		} `json:"international_size_chart"`
		Details struct {
			Model struct {
				Height int    `json:"height"`
				Size   string `json:"size"`
			} `json:"model"`
			Fabric struct {
				Type string `json:"type"`
				Care string `json:"care"`
			} `json:"fabric"`
			Fit                   []string    `json:"fit"`
			AdditionalDetails     []string    `json:"additional_details"`
			Description           string      `json:"description"`
			Sustainability        interface{} `json:"sustainability"`
			SustainabilityDetails []string    `json:"sustainability_details"`
			Factory               struct {
				Location  string `json:"location"`
				Country   string `json:"country"`
				Permalink string `json:"permalink"`
			} `json:"factory"`
		} `json:"details"`
		Variants []struct {
			ID                  int         `json:"id"`
			Sku                 string      `json:"sku"`
			Upc                 string      `json:"upc"`
			Available           int         `json:"available"`
			OrderableState      string      `json:"orderable_state"`
			InventoryCount      int         `json:"inventory_count"`
			Size                string      `json:"size"`
			AbbreviatedSize     string      `json:"abbreviated_size"`
			Annotation          interface{} `json:"annotation"`
			Name                string      `json:"name"`
			RestockDate         interface{} `json:"restock_date"`
			LaunchDate          string      `json:"launch_date"`
			FulfillmentCenterID int         `json:"fulfillment_center_id"`
			IsDigitalGiftcard   bool        `json:"is_digital_giftcard"`
			Bundle              struct {
				ID                     int     `json:"id"`
				Name                   string  `json:"name"`
				ThresholdCount         int     `json:"threshold_count"`
				DiscountedPricePerUnit float64 `json:"discounted_price_per_unit"`
			} `json:"bundle"`
			SingleReturnVariant bool   `json:"single_return_variant"`
			MetaGarment         string `json:"meta_garment"`
		} `json:"variants"`
		CwypDiscounts       []float64     `json:"cwyp_discounts"`
		FinalSale           bool          `json:"final_sale"`
		ProductContentPages []interface{} `json:"product_content_pages"`
		ProductSku          string        `json:"product_sku"`
	} `json:"products"`
	ProductFits []interface{} `json:"product_fits"`
}

type parseReviewData struct {
	Includes struct {
		Products map[string]struct {
			ReviewStatistics struct {
				HelpfulVoteCount     int     `json:"HelpfulVoteCount"`
				TotalReviewCount     int     `json:"TotalReviewCount"`
				AverageOverallRating float64 `json:"AverageOverallRating"`
			} `json:"ReviewStatistics"`
		} `json:"Products"`
		ProductsOrder []string `json:"ProductsOrder"`
	} `json:"Includes"`
}

// used to trim html labels in description
var htmlTrimRegp = regexp.MustCompile(`</?[^>]+>`)
var productsExtractReg = regexp.MustCompile(`([A-Z0-9])+.html`)

// parseProduct
func (c *_Crawler) parseProduct(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil {
		return nil
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		matched := strings.Split(resp.Request.URL.Path, "/")
		produrl := "https://www.everlane.com/api/v2/product_groups?product_permalink=" + matched[len(matched)-1]

		fmt.Println(produrl)
		req, err := http.NewRequest(http.MethodGet, produrl, nil)
		req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
		req.Header.Set("Referer", fmt.Sprintf("%s://%s/%s", resp.Request.URL.Scheme, resp.Request.URL.Host, resp.Request.URL.Path))
		req.Header.Set("x-requested-with", "XMLHttpRequest")
		req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.114 Safari/537.36")

		respNew, err := c.httpClient.Do(ctx, req)
		if err != nil {
			panic(err)
		}

		respBody, err = ioutil.ReadAll(respNew.Body)
		if err != nil {
			return err
		}
	}

	if !bytes.Contains(respBody, []byte(`permalink`)) {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}

	var viewData parseProductData
	if err := json.Unmarshal(respBody, &viewData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}

	for _, proditem := range viewData.Products {

		//priceURL := "https://www.everlane.com/api/v2/product_groups/" + strconv.Format(proditem.ProductGroupID) + "/prices?country=US&currency=USD"
		reviewURL := "https://www.everlane.com/api/v2/reviews/filter?reviews[data][Include]=Products&reviews[data][Stats]=Reviews&reviews[data][Limit]=1&reviews[data][Offset]=0&reviews[filters][Filter][]=ProductId:" + strconv.Format(proditem.ID)
		fmt.Println(reviewURL)

		req, err := http.NewRequest(http.MethodGet, reviewURL, nil)
		req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
		req.Header.Set("Referer", fmt.Sprintf("%s://%s/%s", resp.Request.URL.Scheme, resp.Request.URL.Host, resp.Request.URL.Path))
		req.Header.Set("x-requested-with", "XMLHttpRequest")

		respNew, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
			EnableProxy:       true,
			EnableHeadless:    true,
			EnableSessionInit: false,
			KeepSession:       false,
			Reliability:       0,
		})
		if err != nil {
			panic(err)
		}

		respBody, err = ioutil.ReadAll(respNew.Body)
		if err != nil {
			return err
		}

		var viewReviewData parseReviewData
		if err := json.Unmarshal(respBody, &viewReviewData); err != nil {
			c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
			//return err
		}

		review := 0
		rating := 0.0

		if viewReviewData.Includes.Products != nil {
			review = viewReviewData.Includes.Products[strconv.Format(proditem.ID)].ReviewStatistics.TotalReviewCount
			rating = (viewReviewData.Includes.Products[strconv.Format(proditem.ID)].ReviewStatistics.AverageOverallRating)
		}

		description := ""
		if proditem.Details.Model.Height > 0 && proditem.Details.Model.Size != "" {
			description = "Model is " + strconv.Format(math.Round(float64(proditem.Details.Model.Height)/(12))) +
				", Wearing size " + strconv.Format(proditem.Details.Model.Size)
		} else if proditem.Details.Model.Size != "" {
			description = "Wearing size " + strconv.Format(proditem.Details.Model.Size)
		}

		for _, itemdesc := range proditem.Details.Fit {
			description = description + ", " + itemdesc
		}

		for _, itemdesc := range proditem.Details.AdditionalDetails {
			description = description + ", " + itemdesc
		}

		if proditem.Details.Fabric.Type != "" {
			description = description + ", Type: " + proditem.Details.Fabric.Type
		}
		if proditem.Details.Fabric.Care != "" {
			description = description + ", Care:" + proditem.Details.Fabric.Care
		}

		description = description + ", Made in " + proditem.Details.Factory.Location + ", " + proditem.Details.Factory.Country
		description = description + ", " + proditem.Details.Description

		// build product data
		item := pbItem.Product{
			Source: &pbItem.Source{
				Id:       strconv.Format(proditem.ID),
				CrawlUrl: strconv.Format("https://www.everlane.com/products/" + proditem.Permalink),
			},

			CrowdType:   proditem.PrimaryCollection.Gender,
			BrandName:   "Everlane",
			Title:       proditem.DisplayName,
			Description: description,
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
			},
			Stats: &pbItem.Stats{
				ReviewCount: int32(review),
				Rating:      float32(rating),
			},
		}

		item.SubCategory = proditem.PrimaryCollection.BreadcrumbTitle

		if proditem.PrimaryCollection.Gender == "male" {
			item.Category = "Men"
		} else if proditem.PrimaryCollection.Gender == "female" {
			item.Category = "Women"
		}

		originalPrice, _ := strconv.ParseFloat(proditem.Price)
		msrp := proditem.OriginalPrice
		discount := 0.0
		if msrp > 0 && msrp > originalPrice {
			discount = math.Ceil((msrp - originalPrice) / msrp * 100)
		}

		for kv, rawSku := range proditem.Variants {

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

			if rawSku.Available > 0 {
				sku.Stock.StockStatus = pbItem.Stock_InStock
				sku.Stock.StockCount = int32(rawSku.Available)
			}

			// color
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    proditem.Color.HexValue,
				Name:  proditem.Color.Name,
				Value: proditem.Color.Name,
			})

			if kv == 0 {
				for ki, mid := range proditem.Albums.Square {

					sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
						strconv.Format(ki),
						mid.Src,
						strings.ReplaceAll(mid.Src, ",q_auto,w_auto", ",h_1333,q_40,w_1000"),
						strings.ReplaceAll(mid.Src, ",q_auto,w_auto", ",h_1333,q_40,w_500"),
						strings.ReplaceAll(mid.Src, ",q_auto,w_auto", ",h_1333,q_40,w_600"),
						"",
						mid.Tag == "primary",
					))
				}
			}

			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    rawSku.Sku,
				Name:  rawSku.Name,
				Value: rawSku.Name,
			})

			item.SkuItems = append(item.SkuItems, &sku)

		}

		if err = yield(ctx, &item); err != nil {
			return err
		}
	}
	return nil
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		"https://www.everlane.com/collections/womens-outerwear",
		//"https://www.everlane.com/api/v3/collections/womens-bottoms?style=Slim%2FSkinny+Leg&style=Perform+%26+Sweatpants",
		//"https://www.everlane.com/api/v2/product_groups?product_permalink=womens-fixed-waist-work-pant-militaryolive",
		//"https://www.everlane.com/api/v2/product_groups?product_permalink=womens-human-box-cut-tee-white-black",
		//"https://www.everlane.com/products/womens-human-box-cut-tee-white-black?collection=womens-sale",
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

	client, err := proxy.NewProxyClient(os.Getenv("VOILA_PROXY_URL"), cookiejar.New(), logger)
	if err != nil {
		panic(err)
	}

	// instance the spider locally
	spider, err := New(client, logger)
	if err != nil {
		panic(err)
	}
	opts := spider.CrawlOptions(nil)

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
				i.URL.Host = "www.everlane.com"
			}

			// do http requests here.
			nctx, cancel := context.WithTimeout(ctx, time.Minute*5)
			defer cancel()
			resp, err := client.DoWithOptions(nctx, i, http.Options{
				EnableProxy:       true,
				EnableHeadless:    true,
				EnableSessionInit: spider.CrawlOptions(nil).EnableSessionInit,
				KeepSession:       spider.CrawlOptions(nil).KeepSession,
				Reliability:       spider.CrawlOptions(nil).Reliability,
			})
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			return spider.Parse(ctx, resp, callback)
		default:
			// output the result
			data, err := protojson.Marshal(i.(proto.Message))
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
