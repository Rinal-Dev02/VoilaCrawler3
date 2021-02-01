package main

// This crawler is a little complex because every colour sku is an isolate product
// there not exists a unique key to point out that all the 'products' related to the same product in concept.

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/voiladev/VoilaCrawl/pkg/crawler"
	"github.com/voiladev/VoilaCrawl/pkg/net/http"
	"github.com/voiladev/VoilaCrawl/pkg/net/http/proxycrawl"
	pbMedia "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
	"google.golang.org/protobuf/types/known/anypb"
)

type _Crawler struct {
	httpClient http.Client

	categoryPathMatcher     *regexp.Regexp
	categoryJsonPathMatcher *regexp.Regexp
	productPathMatcher      *regexp.Regexp
	logger                  glog.Log
}

func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient:              client,
		categoryPathMatcher:     regexp.MustCompile(`^(/en)?/ts[a-z0-9]+/category(/[a-z0-9-]+-[0-9]+){1,3}$`),
		categoryJsonPathMatcher: regexp.MustCompile(`^/api/products$`),
		productPathMatcher:      regexp.MustCompile(`^(/en)?/ts[a-z0-9]+/product(/[a-z0-9-]+-[0-9]+){2,}$`),
		logger:                  logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "75b7e41bae2fde3a0706acaf58575ba6"
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
	// options.MustHeader.Set("X-Requested-With", "XMLHttpRequest")
	options.MustCookies = append(options.MustCookies,
		// source,akavpau_VP_TS,bm_mi,bm_sv,_arc_be_jwt
		&http.Cookie{Name: "deviceType", Value: "laptop", Path: "/"},
		&http.Cookie{Name: "viewport", Value: "laptp", Path: "/"},
		&http.Cookie{Name: "GEOIP", Value: "US", Path: "/"},
		&http.Cookie{Name: "ENSEARCH", Value: "BENVER=1", Path: "/"},
	)
	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"us.topshop.com", "topshop.com"}
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
		// https://us.topshop.com/en/tsus/category/bags-accessories-7594012
		return c.parseCategoryProducts(ctx, resp, yield)
	} else if c.categoryJsonPathMatcher.MatchString(resp.Request.URL.Path) {
		// https://us.topshop.com/api/products?currentPage=3&pageSize=24&category=208582
		return c.parseCategoryProductsJson(ctx, resp, yield)
	} else if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		// https://us.topshop.com/en/tsus/product/bags-accessories-7594012/hats-70518/borg-bucket-hat-10114468
		return c.parseProduct(ctx, resp, yield)
	}
	return fmt.Errorf("unsupported url %s", resp.Request.URL.String())
}

// nextIndex used to get sharingData from context
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")) + 1)
}

type parseCategoryProductsResponse struct {
	Products struct {
		Breadcrumbs []struct {
			Label    string `json:"label"`
			URL      string `json:"url,omitempty"`
			Category string `json:"category,omitempty"`
		} `json:"breadcrumbs"`
		Refinements []struct {
			Label             string `json:"label"`
			RefinementOptions []struct {
				Type   string `json:"type"`
				Label  string `json:"label"`
				Value  string `json:"value"`
				Count  int    `json:"count"`
				SeoURL string `json:"seoUrl"`
			} `json:"refinementOptions"`
		} `json:"refinements"`
		CanonicalURL string `json:"canonicalUrl"`
		Products     []struct {
			ProductID  int    `json:"productId"`
			LineNumber string `json:"lineNumber"`
			Name       string `json:"name"`
			UnitPrice  string `json:"unitPrice"`
			ProductURL string `json:"productUrl"`
			SeoURL     string `json:"seoUrl"`
			Assets     []struct {
				AssetType string `json:"assetType"`
				Index     int    `json:"index"`
				URL       string `json:"url"`
			} `json:"assets"`
			AdditionalAssets []struct {
				AssetType string `json:"assetType"`
				Index     int    `json:"index"`
				URL       string `json:"url"`
			} `json:"additionalAssets"`
			Items                   []interface{} `json:"items"`
			BundleProducts          []interface{} `json:"bundleProducts"`
			ColourSwatches          []interface{} `json:"colourSwatches"`
			TpmLinks                []interface{} `json:"tpmLinks"`
			BundleSlots             []interface{} `json:"bundleSlots"`
			AgeVerificationRequired bool          `json:"ageVerificationRequired"`
			IsBundleOrOutfit        bool          `json:"isBundleOrOutfit"`
			ProductBaseImageURL     string        `json:"productBaseImageUrl"`
			OutfitBaseImageURL      string        `json:"outfitBaseImageUrl"`
			WasPrice                string        `json:"wasPrice,omitempty"`
		} `json:"products"`
		CategoryTitle       string `json:"categoryTitle"`
		CategoryDescription string `json:"categoryDescription"`
		TotalProducts       int    `json:"totalProducts"`
		SortOptions         []struct {
			Label           string `json:"label"`
			Value           string `json:"value"`
			NavigationState string `json:"navigationState"`
		} `json:"sortOptions"`
		DefaultImgType string `json:"defaultImgType"`
		Title          string `json:"title"`
		ShouldIndex    bool   `json:"shouldIndex"`
		Config         struct {
			BrandDisplayName string `json:"brandDisplayName"`
		} `json:"config"`
	} `json:"products"`
	CurrentProduct struct {
		PageTitle    string `json:"pageTitle"`
		IscmCategory string `json:"iscmCategory"`
		Breadcrumbs  []struct {
			Label string `json:"label"`
			URL   string `json:"url,omitempty"`
		} `json:"breadcrumbs"`
		ProductID          int    `json:"productId"`
		Grouping           string `json:"grouping"`
		LineNumber         string `json:"lineNumber"`
		Colour             string `json:"colour"`
		Name               string `json:"name"`
		Description        string `json:"description"`
		NotifyMe           bool   `json:"notifyMe"`
		UnitPrice          string `json:"unitPrice"`
		StockEmail         bool   `json:"stockEmail"`
		StoreDelivery      bool   `json:"storeDelivery"`
		StockThreshold     int    `json:"stockThreshold"`
		WcsColourKey       string `json:"wcsColourKey"`
		WcsColourADValueID string `json:"wcsColourADValueId"`
		WcsSizeKey         string `json:"wcsSizeKey"`
		Assets             []struct {
			Index     int    `json:"index"`
			AssetType string `json:"assetType"`
			URL       string `json:"url"`
		} `json:"assets"`
		Items []struct {
			AttrName    string `json:"attrName"`
			Quantity    int    `json:"quantity"`
			CatEntryID  int    `json:"catEntryId"`
			AttrValueID int    `json:"attrValueId"`
			Selected    bool   `json:"selected"`
			StockText   string `json:"stockText"`
			AttrValue   string `json:"attrValue"`
			Size        string `json:"size"`
			Sku         string `json:"sku"`
			SizeMapping string `json:"sizeMapping"`
		} `json:"items"`
		BundleProducts []interface{} `json:"bundleProducts"`
		Attributes     struct {
			AverageOverallRating      string `json:"AverageOverallRating"`
			ECMCPRODCE3JEANSTYLE2     string `json:"ECMC_PROD_CE3_JEAN_STYLE_2"`
			ThresholdMessage          string `json:"ThresholdMessage"`
			SizeFit                   string `json:"SizeFit"`
			ProductDefaultCopy        string `json:"ProductDefaultCopy"`
			EcmcUpdatedTimestamp      string `json:"ecmcUpdatedTimestamp"`
			COLOURCODE                string `json:"COLOUR_CODE"`
			IOThumbnailSuffixes       string `json:"IOThumbnailSuffixes"`
			Has360                    string `json:"has360"`
			REALCOLOURS               string `json:"REAL_COLOURS"`
			RRP                       string `json:"RRP"`
			Department                string `json:"Department"`
			VersionGroup              string `json:"Version_Group"`
			Version                   string `json:"Version"`
			EmailBackInStock          string `json:"EmailBackInStock"`
			ECMCPRODSIZEGUIDE2        string `json:"ECMC_PROD_SIZE_GUIDE_2"`
			ECMCPRODCE3PRODUCTTYPE2   string `json:"ECMC_PROD_CE3_PRODUCT_TYPE_2"`
			EcmcCreatedTimestamp      string `json:"ecmcCreatedTimestamp"`
			ECMCPRODPRODUCTTYPE2      string `json:"ECMC_PROD_PRODUCT_TYPE_2"`
			StyleCode                 string `json:"StyleCode"`
			ThumbnailImageSuffixes    string `json:"thumbnailImageSuffixes"`
			BHasVideo                 string `json:"b_hasVideo"`
			ECMCPRODCE3BRANDS2        string `json:"ECMC_PROD_CE3_BRANDS_2"`
			HasVideo                  string `json:"hasVideo"`
			CountryExclusion          string `json:"countryExclusion"`
			STYLECODE                 string `json:"STYLE_CODE"`
			ECMCPRODCE3JEANFIT2       string `json:"ECMC_PROD_CE3_JEAN_FIT_2"`
			ShopTheOutfitBundleCode   string `json:"shopTheOutfitBundleCode"`
			IFSeason                  string `json:"IFSeason"`
			SearchKeywords            string `json:"SearchKeywords"`
			BHasImage                 string `json:"b_hasImage"`
			CE3ThumbnailSuffixes      string `json:"CE3ThumbnailSuffixes"`
			ECMCPRODFIT2              string `json:"ECMC_PROD_FIT_2"`
			BHas360                   string `json:"b_has360"`
			NotifyMe                  string `json:"NotifyMe"`
			ECMCPRODCE3JEANWAISTTYPE2 string `json:"ECMC_PROD_CE3_JEAN_WAIST_TYPE_2"`
			ECMCPRODCOLOUR2           string `json:"ECMC_PROD_COLOUR_2"`
		} `json:"attributes"`
		ColourSwatches []struct {
			ColourName string `json:"colourName"`
			ImageUrl   string `json:"imageUrl"`
			ProductId  string `json:"productId"`
			ProductUrl string `json:"productUrl"`
		} `json:"colourSwatches"`
		TpmLinks [][]struct {
			CatentryID  string `json:"catentryId"`
			TPMURL      string `json:"TPMUrl"`
			IsTPMActive bool   `json:"isTPMActive"`
			TPMName     string `json:"TPMName"`
		} `json:"tpmLinks"`
		BundleSlots             []interface{} `json:"bundleSlots"`
		SourceURL               string        `json:"sourceUrl"`
		CanonicalURL            string        `json:"canonicalUrl"`
		CanonicalURLSet         bool          `json:"canonicalUrlSet"`
		AgeVerificationRequired bool          `json:"ageVerificationRequired"`
		IsBundleOrOutfit        bool          `json:"isBundleOrOutfit"`
		ProductDataQuantity     struct {
			ColourAttributes struct {
				AttrValue string `json:"attrValue"`
				AttrName  string `json:"attrName"`
			} `json:"colourAttributes"`
			Quantities         []int `json:"quantities"`
			InventoryPositions []struct {
				CatentryID string `json:"catentryId"`
				Inventorys []struct {
					Cutofftime   interface{} `json:"cutofftime"`
					Quantity     int         `json:"quantity"`
					FfmcenterID  int         `json:"ffmcenterId"`
					Expressdates interface{} `json:"expressdates"`
				} `json:"inventorys,omitempty"`
			} `json:"inventoryPositions"`
			SKUs []struct {
				Skuid              string `json:"skuid"`
				Value              string `json:"value"`
				Availableinventory string `json:"availableinventory"`
				Partnumber         string `json:"partnumber"`
				AttrName           string `json:"attrName"`
			} `json:"SKUs"`
		} `json:"productDataQuantity"`
		Version string `json:"version"`
		Espots  struct {
			CEProductEspotCol1Pos1 struct {
				EspotContents struct {
					CmsMobileContent struct {
						PageID           int    `json:"pageId"`
						PageName         string `json:"pageName"`
						Breadcrumb       string `json:"breadcrumb"`
						Baseline         string `json:"baseline"`
						Revision         string `json:"revision"`
						LastPublished    string `json:"lastPublished"`
						ContentPath      string `json:"contentPath"`
						SeoURL           string `json:"seoUrl"`
						MobileCMSURL     string `json:"mobileCMSUrl"`
						ResponsiveCMSURL string `json:"responsiveCMSUrl"`
					} `json:"cmsMobileContent"`
					EncodedcmsMobileContent string `json:"encodedcmsMobileContent"`
				} `json:"EspotContents"`
			} `json:"CEProductEspotCol1Pos1"`
			CEProductEspotCol2Pos1 struct {
				EspotContents struct {
					CmsMobileContent struct {
						PageID        int    `json:"pageId"`
						PageName      string `json:"pageName"`
						Baseline      string `json:"baseline"`
						Revision      string `json:"revision"`
						LastPublished string `json:"lastPublished"`
						ContentPath   string `json:"contentPath"`
						SeoURL        string `json:"seoUrl"`
						MobileCMSURL  string `json:"mobileCMSUrl"`
					} `json:"cmsMobileContent"`
					EncodedcmsMobileContent string `json:"encodedcmsMobileContent"`
				} `json:"EspotContents"`
			} `json:"CEProductEspotCol2Pos1"`
			CEProductEspotCol2Pos2 struct {
				EspotContents struct {
					CmsMobileContent struct {
						PageID           int    `json:"pageId"`
						PageName         string `json:"pageName"`
						Breadcrumb       string `json:"breadcrumb"`
						Baseline         string `json:"baseline"`
						Revision         string `json:"revision"`
						LastPublished    string `json:"lastPublished"`
						ContentPath      string `json:"contentPath"`
						SeoURL           string `json:"seoUrl"`
						MobileCMSURL     string `json:"mobileCMSUrl"`
						ResponsiveCMSURL string `json:"responsiveCMSUrl"`
					} `json:"cmsMobileContent"`
					EncodedcmsMobileContent string `json:"encodedcmsMobileContent"`
				} `json:"EspotContents"`
			} `json:"CEProductEspotCol2Pos2"`
			CEProductEspotCol2Pos4 struct {
				EspotContents struct {
					CmsMobileContent struct {
						PageID           int    `json:"pageId"`
						PageName         string `json:"pageName"`
						Breadcrumb       string `json:"breadcrumb"`
						Baseline         string `json:"baseline"`
						Revision         string `json:"revision"`
						LastPublished    string `json:"lastPublished"`
						ContentPath      string `json:"contentPath"`
						SeoURL           string `json:"seoUrl"`
						MobileCMSURL     string `json:"mobileCMSUrl"`
						ResponsiveCMSURL string `json:"responsiveCMSUrl"`
					} `json:"cmsMobileContent"`
					EncodedcmsMobileContent string `json:"encodedcmsMobileContent"`
				} `json:"EspotContents"`
			} `json:"CEProductEspotCol2Pos4"`
			CE3ContentEspot1 struct {
				EspotContents struct {
					CmsMobileContent struct {
						PageID           int    `json:"pageId"`
						PageName         string `json:"pageName"`
						Breadcrumb       string `json:"breadcrumb"`
						Baseline         string `json:"baseline"`
						Revision         string `json:"revision"`
						LastPublished    string `json:"lastPublished"`
						ContentPath      string `json:"contentPath"`
						SeoURL           string `json:"seoUrl"`
						MobileCMSURL     string `json:"mobileCMSUrl"`
						ResponsiveCMSURL string `json:"responsiveCMSUrl"`
					} `json:"cmsMobileContent"`
					EncodedcmsMobileContent string `json:"encodedcmsMobileContent"`
				} `json:"EspotContents"`
			} `json:"CE3ContentEspot1"`
		} `json:"espots"`
		ShopTheLookProducts bool   `json:"shopTheLookProducts"`
		BundleDisplayURL    string `json:"bundleDisplayURL"`
		AdditionalAssets    []struct {
			AssetType string `json:"assetType"`
			Index     int    `json:"index"`
			URL       string `json:"url"`
		} `json:"additionalAssets"`
		AmplienceAssets struct {
			Images []string `json:"images"`
		} `json:"amplienceAssets"`
		IsDDPProduct       bool `json:"isDDPProduct"`
		BnplPaymentOptions struct {
			Klarna struct {
				Instalments int    `json:"instalments"`
				Amount      string `json:"amount"`
			} `json:"klarna"`
			Clearpay struct {
				Instalments int    `json:"instalments"`
				Amount      string `json:"amount"`
			} `json:"clearpay"`
		} `json:"bnplPaymentOptions"`
		WasPrice           string `json:"wasPrice"`
		TotalMarkdownValue string `json:"totalMarkdownValue"`
	} `json:"currentProduct"`
}

var (
	prodDataExtraReg = regexp.MustCompile(`window.__INITIAL_STATE__=({.*,\s*"contentId":.*}});`)
	defaultPageSize  = 24
)

// parseCategoryProducts parse api url from web page url
func (c *_Crawler) parseCategoryProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	lastIndex := nextIndex(ctx)

	// extract html content
	// doc, err := goquery.NewDocumentFromReader(resp.Body)
	// if err != nil {
	// 	return err
	// }
	// sel := doc.Find(`.ProductList>.ProductList-products>.Product-images-container>.Product-link`)
	// for i := range sel.Nodes {
	// 	node := sel.Eq(i)
	// 	if u, exists := node.Attr("href"); exists {
	// 		req, _ := http.NewRequest(http.MethodGet, u, nil)
	// 		nctx := context.WithValue(ctx, "item.index", lastIndex+1)
	// 		if err := yield(nctx, req); err != nil {
	// 			return err
	// 		}
	// 	}
	// }

	matched := prodDataExtraReg.FindSubmatch(respBody)
	if len(matched) < 1 {
		return fmt.Errorf("product info not found for url %s", resp.Request.URL)
	}
	var initData parseCategoryProductsResponse
	if err := json.Unmarshal(matched[1], &initData); err != nil {
		return fmt.Errorf("parse json data from %s failed, error=%s", resp.Request.URL, err)
	}

	for _, prod := range initData.Products.Products {
		req, _ := http.NewRequest(http.MethodGet, prod.ProductURL, nil)
		nctx := context.WithValue(ctx, "item.index", lastIndex+1)
		if err := yield(nctx, req); err != nil {
			return err
		}
	}

	currentPage := strconv.MustParseInt(resp.Request.URL.Query().Get("currentPage"))
	if currentPage == 0 {
		currentPage = 1
	}
	if int(currentPage)*defaultPageSize >= initData.Products.TotalProducts {
		return nil
	}
	cate := initData.Products.Breadcrumbs[len(initData.Products.Breadcrumbs)-1].Category
	pageUrl := fmt.Sprintf("/api/products?currentPage=%v&pageSize=%d&category=%s", currentPage+1, defaultPageSize, cate)
	if req, err := http.NewRequest(http.MethodGet, pageUrl, nil); err != nil {
		return err
	} else {
		req.Header.Set("brand-code", "tsus")
		return yield(context.WithValue(ctx, "item.index", lastIndex), req)
	}
}

// parseCategoryProductsJson
func (c *_Crawler) parseCategoryProductsJson(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil {
		return nil
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	c.logger.Debugf("%s", respBody)

	var r parseCategoryProductsResponse
	if err := json.Unmarshal(respBody, &r); err != nil {
		return err
	}

	lastIndex := nextIndex(ctx)
	for _, prod := range r.Products.Products {
		req, _ := http.NewRequest(http.MethodGet, prod.ProductURL, nil)
		nctx := context.WithValue(ctx, "item.index", lastIndex+1)
		if err := yield(nctx, req); err != nil {
			return err
		}
	}

	currentPage := strconv.MustParseInt(resp.Request.URL.Query().Get("currentPage"))
	if currentPage == 0 {
		currentPage = 1
	}
	pageSize := strconv.MustParseInt(resp.Request.URL.Query().Get("pageSize"))
	if int(currentPage)*defaultPageSize >= r.Products.TotalProducts || len(r.Products.Products) < int(pageSize) {
		return nil
	}

	u := *resp.Request.URL
	vals := u.Query()
	vals.Set("currentPage", strconv.Format(currentPage+1))
	u.RawQuery = vals.Encode()

	if req, err := http.NewRequest(http.MethodGet, u.String(), nil); err != nil {
		return err
	} else {
		req.Header.Set("brand-code", "tsus")
		return yield(context.WithValue(ctx, "item.index", lastIndex), req)
	}
}

func (c *_Crawler) parseProduct(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error(err)
		return err
	}

	matched := prodDataExtraReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		return fmt.Errorf("parse product detail from url %s failed, error=%s", resp.Request.URL, err)
	}

	var r parseCategoryProductsResponse
	if err := json.Unmarshal(matched[1], &r); err != nil {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("parse fetch product detail failed, error=%s", err)
	}

	var (
		p        = r.CurrentProduct
		uniqName = p.Name
		skuSpec  *pbItem.SkuSpecOption
	)

	if len(p.ColourSwatches) > 1 {
		skuSpec = &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecColor,
			Name:  strings.Title(p.Colour),
			Value: strings.Title(p.Colour),
		}
		for _, colour := range p.ColourSwatches {
			if colour.ProductId == strconv.Format(p.ProductID) {
				skuSpec.Icon = colour.ImageUrl
			} else if req, err := http.NewRequest(http.MethodGet, colour.ProductUrl, nil); err != nil {
				c.logger.Errorf("new request of url %s failed, error=%s", colour.ProductUrl, err)
			} else if err = yield(ctx, req); err != nil {
				return err
			}
		}
	}

	for _, tpmLinks := range p.TpmLinks {
		for _, link := range tpmLinks {
			if link.CatentryID == strconv.Format(p.ProductID) {
				continue
			}
			if link.TPMURL != "" {
				if req, err := http.NewRequest(http.MethodGet, link.TPMURL, nil); err != nil {
					c.logger.Errorf("new request for url %s failed, error=%s", link.TPMURL, err)
				} else if err = yield(ctx, req); err != nil {
					return err
				}
			}
		}
	}

	if index := strings.Index(strings.ToLower(p.Name), strings.ToLower(p.Colour)); index >= 0 {
		uniqName = strings.TrimSpace(p.Name[index+len(p.Colour):])
	}
	uniqName = fmt.Sprintf("%s-%s", p.Attributes.STYLECODE, uniqName)

	item := pbItem.Product{
		Source: &pbItem.Source{
			Id: strconv.Format(p.ProductID),
			// NOTE: this field is very important to point. use this field to distiguish the same product
			GlobalUniqId: uniqName,
			GroupId:      p.Grouping,
		},
		BrandName:   "Topshop",
		Title:       p.Name,
		Description: html.UnescapeString(p.Description),
		CrowdType:   "women", // as if products on this site is for women
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
			Current:  int32(strconv.MustParseFloat(p.UnitPrice) * 100),
			Msrp:     int32(strconv.MustParseFloat(p.WasPrice) * 100),
		},
		Stock: &pbItem.Stock{},
		Stats: &pbItem.Stats{
			Rating: float32(strconv.MustParseFloat(p.Attributes.AverageOverallRating)),
		},
		ExtraInfo: map[string]string{
			"fit": p.Attributes.ECMCPRODFIT2,
		},
	}
	if len(r.CurrentProduct.Breadcrumbs) > 2 {
		item.Category = r.CurrentProduct.Breadcrumbs[1].Label
	}
	if len(r.CurrentProduct.Breadcrumbs) > 3 {
		item.SubCategory = r.CurrentProduct.Breadcrumbs[2].Label
	}
	if len(r.CurrentProduct.Breadcrumbs) > 4 {
		item.SubCategory2 = r.CurrentProduct.Breadcrumbs[3].Label
	}
	imgMap := map[int]struct{}{}
	for _, assert := range p.Assets {
		if _, ok := imgMap[assert.Index]; ok {
			continue
		}

		splited := strings.Split(assert.URL, "?")
		imgData, _ := anypb.New(&pbMedia.Media_Image{
			OriginalUrl: splited[0],
			LargeUrl:    fmt.Sprintf("%s?$w1000$&fmt=jpeg", splited[0]),
			MediumUrl:   fmt.Sprintf("%s?$w600$&fmt=jpeg", splited[0]),
			SmallUrl:    fmt.Sprintf("%s?$w500$&fmt=jpeg", splited[0]),
		})
		item.Medias = append(item.Medias, &pbMedia.Media{
			Detail:    imgData,
			IsDefault: len(item.Medias) == 0,
		})
		imgMap[assert.Index] = struct{}{}
	}

	// skus
	for _, i := range p.Items {
		sku := pbItem.Sku{
			SourceId: i.Sku,
			Stock: &pbItem.Stock{
				StockStatus: pbItem.Stock_InStock,
				StockCount:  int32(i.Quantity),
			},
		}
		if skuSpec != nil {
			sku.Specs = append(sku.Specs, skuSpec)
		}
		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Id:    strconv.Format(i.AttrValueID),
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Value: i.AttrValue,
			Name:  i.AttrValue,
		})
		if i.Quantity == 0 {
			sku.Stock.StockStatus = pbItem.Stock_OutOfStock
		}
		item.SkuItems = append(item.SkuItems, &sku)
		if sku.GetStock().GetStockCount() > 0 {
			item.Stock.StockStatus = pbItem.Stock_InStock
		}
	}
	return yield(ctx, &item)
}

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		// "https://us.topshop.com/en/tsus/product/clothing-70483/dresses-70497/idol-ruffle-patchwork-print-midi-dress-10411905",
		// "https://us.topshop.com/en/tsus/product/clothing-70483/jeans-4593087/ecru-jamie-skinny-jeans-9611474",
		// "https://us.topshop.com/en/tsus/product/clothing-70483/jeans-4593087/blue-black-jamie-skinny-jeans-9713412",
		// "https://us.topshop.com/en/tsus/product/clothing-70483/petite-70510/petite-black-jamie-skinny-stretch-jeans-10407726",
		// "https://us.topshop.com/en/tsus/product/clothing-70483/jeans-4593087/black-jamie-skinny-jeans-9611011",
		// "https://us.topshop.com/en/tsus/category/sale-6923951/shop-all-sale-7108379",
		"https://us.topshop.com/api/products?currentPage=2&pageSize=24&category=386999,397549",
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
	if apiToken == "" || jsToken == "" {
		panic("env PC_API_TOKEN or PC_JS_TOKEN is not set")
	}

	logger := glog.New(glog.LogLevelDebug)
	client, err := proxycrawl.NewProxyCrawlClient(logger,
		proxycrawl.Options{APIToken: apiToken, JSToken: jsToken},
	)
	if err != nil {
		panic(err)
	}

	spider, err := New(client, logger)
	if err != nil {
		panic(err)
	}
	opts := spider.CrawlOptions()

	for _, req := range spider.NewTestRequest(context.Background()) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
		defer cancel()

		logger.Debugf("Access %s", req.URL)
		for k := range opts.MustHeader {
			req.Header.Set(k, opts.MustHeader.Get(k))
		}
		req.Header.Set("brand-code", "tsus")
		for _, c := range opts.MustCookies {
			if strings.HasPrefix(req.URL.Path, c.Path) || c.Path == "" {
				val := fmt.Sprintf("%s=%s", c.Name, c.Value)
				if c := req.Header.Get("Cookie"); c != "" {
					req.Header.Set("Cookie", c+"; "+val)
				} else {
					req.Header.Set("Cookie", val)
				}
			}
		}

		resp, err := client.DoWithOptions(ctx, req, http.Options{EnableProxy: true})
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		if err := spider.Parse(ctx, resp, func(ctx context.Context, val interface{}) error {
			switch i := val.(type) {
			case *http.Request:
				logger.Infof("new request %s", i.URL)
			default:
				data, err := json.Marshal(i)
				if err != nil {
					return err
				}
				logger.Infof("data: %s", data)
			}
			return nil
		}); err != nil {
			panic(err)
		}
		break
	}
}
