package main

import (
	//"bytes"
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
	//"github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
	//"google.golang.org/protobuf/types/known/anypb"
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
		categoryPathMatcher:     regexp.MustCompile(`^(/[^.]+)?/(category)(.*)$`),
		productPathMatcher:      regexp.MustCompile(`^(/[^.]+)?/(product)(.*)$`),
		logger:                  logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "aa52a9912b124c248f308833b0315793"
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
		&http.Cookie{Name: "geocountry", Value: `US`, Path: "/"},
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
	return []string{"www.dsw.com"}
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

var prodDataExtraReg = regexp.MustCompile(`(.*)`)

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

	// extract html content
	// doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	// if err != nil {
	// 	return err
	// }
	// doc.Find(`div[data-auto-id="productList"]>section>article[data-auto-id="productTile"]>a`).Each(func(i int, s *goquery.Selection) {
	// 	if u, exists := s.Attr("href"); exists {
	// 		req, _ := http.NewRequest(http.MethodGet, u, nil)
	// 		yield(ctx, req)
	// 	}
	// })

	// next page
	matched := prodDataExtraReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		return fmt.Errorf("extract json from product list page %s failed", resp.Request.URL)
	}

	var r struct {
		PageContentItem struct {
			RuleLimit          string        `json:"ruleLimit"`
			Name               string        `json:"name"`
			TemplateTypes      []string      `json:"templateTypes"`
			TemplateIds        []interface{} `json:"templateIds"`
			Type               string        `json:"@type"`
			ContentPaths       []string      `json:"contentPaths"`
			EndecaSiteRootPath string        `json:"endeca:siteRootPath"`
			EndecaContentPath  string        `json:"endeca:contentPath"`
			Contents           []struct {
				Name        string `json:"name"`
				Type        string `json:"@type"`
				MainContent []struct {
					RuleLimit     string        `json:"ruleLimit"`
					Name          string        `json:"name"`
					TemplateTypes []string      `json:"templateTypes"`
					TemplateIds   []interface{} `json:"templateIds"`
					Type          string        `json:"@type"`
					ContentPaths  []string      `json:"contentPaths"`
					Contents      []struct {
						LastRecNum int `json:"lastRecNum"`
						Records    []struct {
							Attributes struct {
								ProductMinPrice              []string `json:"product.min_price"`
								ProductProductTypeWeb        []string `json:"product.productTypeWeb"`
								Gender                       []string `json:"gender"`
								BrandLogoTileAvailable       []string `json:"brand.logoTileAvailable"`
								NonMemberPrice               []string `json:"nonMemberPrice"`
								ProductOriginalStyleID       []string `json:"product.originalStyleId"`
								ProductDisplayName           []string `json:"product.displayName"`
								Rating                       []string `json:"rating"`
								AllAncestorsRepositoryID     []string `json:"allAncestors.repositoryId"`
								ProductHasAnimatedImage      []string `json:"product.hasAnimatedImage"`
								ProductCategory              []string `json:"product.category"`
								ProductNonMemberMaxPrice     []string `json:"product.nonMemberMaxPrice"`
								IsClearance                  []string `json:"isClearance"`
								Msrp                         []string `json:"msrp"`
								ProductDswBrandRepositoryID  []string `json:"product.dswBrand.repositoryId"`
								Brand                        []string `json:"brand"`
								ProductOriginalPrice         []string `json:"product.originalPrice"`
								ProductReviewCount           []string `json:"product.reviewCount"`
								ProductShowPriceInCart       []string `json:"product.showPriceInCart"`
								ProductIsMinPriceinClearance []string `json:"product.isMinPriceinClearance"`
								RecordID                     []string `json:"record.id"`
								ProductSelectedColorCode     []string `json:"product.selectedColorCode"`
								ProductRepositoryID          []string `json:"product.repositoryId"`
								ProductOnClearance           []string `json:"product.on_clearance"`
								ProductDefaultColorCode      []string `json:"product.defaultColorCode"`
								ProductColorNames            []string `json:"product.colorNames"`
								ProductIsMaxPriceinClearance []string `json:"product.isMaxPriceinClearance"`
								ProductColorCodes            []string `json:"product.colorCodes"`
								ProductNonMemberMinPrice     []string `json:"product.nonMemberMinPrice"`
								ProductOnSale                []string `json:"product.on_sale"`
								ProductMaxPrice              []string `json:"product.max_price"`
							} `json:"attributes,omitempty"`
							DetailsAction struct {
								SiteRootPath string `json:"siteRootPath"`
								ContentPath  string `json:"contentPath"`
								SiteState    struct {
									SiteID         string `json:"siteId"`
									SiteDefinition struct {
										ID          string        `json:"id"`
										Patterns    []interface{} `json:"patterns"`
										DisplayName string        `json:"displayName"`
									} `json:"siteDefinition"`
									SiteDisplayName string `json:"siteDisplayName"`
									ContentPath     string `json:"contentPath"`
									Properties      struct {
									} `json:"properties"`
									ValidSite bool `json:"validSite"`
								} `json:"siteState"`
								RecordState string `json:"recordState"`
							} `json:"detailsAction,omitempty"`
							Records []struct {
								Attributes struct {
									ToeShape           []string `json:"toeShape"`
									HeelHeight         []string `json:"heelHeight"`
									Color              []string `json:"color"`
									NonMemberPrice     []string `json:"nonMemberPrice"`
									SkuIsClearanceItem []string `json:"sku.isClearanceItem"`
									SkuInventory       []string `json:"sku.inventory"`
									Materials          []string `json:"materials"`
									Width              []string `json:"width"`
									ColorCode          []string `json:"colorCode"`
									ListPrice          []string `json:"listPrice"`
								} `json:"attributes"`
								DetailsAction struct {
									SiteRootPath string `json:"siteRootPath"`
									ContentPath  string `json:"contentPath"`
									SiteState    struct {
										SiteID         string `json:"siteId"`
										SiteDefinition struct {
											ID          string        `json:"id"`
											Patterns    []interface{} `json:"patterns"`
											DisplayName string        `json:"displayName"`
										} `json:"siteDefinition"`
										SiteDisplayName string `json:"siteDisplayName"`
										ContentPath     string `json:"contentPath"`
										Properties      struct {
										} `json:"properties"`
										ValidSite bool `json:"validSite"`
									} `json:"siteState"`
									RecordState string `json:"recordState"`
								} `json:"detailsAction"`
								NumRecords int `json:"numRecords"`
							} `json:"records,omitempty"`
						} `json:"records"`
						PagingActionTemplate struct {
							SiteRootPath string `json:"siteRootPath"`
							ContentPath  string `json:"contentPath"`
							SiteState    struct {
								SiteID         string `json:"siteId"`
								SiteDefinition struct {
									ID          string        `json:"id"`
									Patterns    []interface{} `json:"patterns"`
									DisplayName string        `json:"displayName"`
								} `json:"siteDefinition"`
								SiteDisplayName string `json:"siteDisplayName"`
								ContentPath     string `json:"contentPath"`
								Properties      struct {
								} `json:"properties"`
								ValidSite bool `json:"validSite"`
							} `json:"siteState"`
							NavigationState string `json:"navigationState"`
						} `json:"pagingActionTemplate"`
						TotalNumRecs int `json:"totalNumRecs"`
					} `json:"contents"`
				} `json:"mainContent"`
				MappedURL string `json:"mappedUrl"`
			} `json:"contents"`
		} `json:"pageContentItem"`
	}

	//matched[1] = bytes.ReplaceAll(bytes.ReplaceAll(matched[1], []byte("\\'"), []byte("'")), []byte(`\\"`), []byte(`\"`))
	
	// rawData, err := strconv.Unquote(string(matched[1]))
	//if err != nil {
	//	c.logger.Errorf("unquote raw string failed, error=%s", err)
	//	return err
	//}
	
	if err = json.Unmarshal(matched[0], &r); err != nil {
		c.logger.Debugf("parse %s failed, error=%s", matched[1], err)
		return err
	}

	lastrecordno := 0
	totalrecords := 0
	nexturl := ""
	lastIndex := nextIndex(ctx)
	for _, prod := range r.PageContentItem.Contents[0].MainContent {
		if prod.Name != "ResultList Zone" {
			continue
		}

		lastrecordno = int(prod.Contents[0].LastRecNum)
		totalrecords = int(prod.Contents[0].TotalNumRecs)
		nexturl = prod.Contents[0].PagingActionTemplate.SiteState.ContentPath
		for _, result := range prod.Contents[0].Records {

		
		rawurl := fmt.Sprintf("%s://%s/us/%s", resp.Request.URL.Scheme, resp.Request.URL.Host, result.Attributes.ProductOriginalStyleID)
		
		fmt.Println(result.Attributes.ProductDisplayName, rawurl)
		
		
			// if req, err := http.NewRequest(http.MethodGet, rawurl, nil); err != nil {
			// 	c.logger.Debug(err)
			// 	return err
			// } else {
			// 	nnctx := context.WithValue(nctx, "item.index", lastIndex+1)
			// 	lastIndex += 1
			// 	if err = yield(nnctx, req); err != nil {
			// 		return err
			// 	}
			// }
		}
	}

	// get current page number
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("Pageindex"))
	if page == 0 {
		page = 1
	}

	// check if this is the last page

	if lastIndex >= int(totalrecords) {
		return nil
	}

	// set pagination
	// u := *resp.Request.URL
	// vals := u.Query()
	// vals.Set("Pageindex", strconv.Format(page+1))
	// u.RawQuery = vals.Encode()

	u := fmt.Sprintf("%s://%s%s%s%v", resp.Request.URL.Scheme, resp.Request.URL.Host, nexturl, "?pagePath=/pages/DSW/category&skipHeaderFooterContent=true&locale=en_US&pushSite=DSW&tier=GUEST&No=", lastrecordno)
		
	fmt.Println("new url", u)
	req, _ := http.NewRequest(http.MethodGet, u, nil)
	// update the index of last page
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)

}

type parseProductResponse struct {
	Response struct {
		Product struct {
			Occasion []struct {
				DisplayName string `json:"displayName"`
			} `json:"occasion"`
			LongDescription string `json:"longDescription"`
			ToeShape        struct {
				DisplayName string `json:"displayName"`
			} `json:"toeShape"`
			HeelHeight struct {
				DisplayName string `json:"displayName"`
			} `json:"heelHeight"`
			DisplayCompareAtPrice   bool          `json:"displayCompareAtPrice"`
			DisplayName             string        `json:"displayName"`
			StyleAssociatedProducts []interface{} `json:"styleAssociatedProducts"`
			NonMemberMaxPrice       float64       `json:"nonMemberMaxPrice"`
			IsActive                bool          `json:"isActive"`
			DefaultColorCode        string        `json:"defaultColorCode"`
			IsClearance             bool          `json:"isClearance"`
			ShowSize                bool          `json:"showSize"`
			DswBrand                struct {
				DisplayNameDefault string `json:"displayNameDefault"`
				NavStringURL       string `json:"navStringURL"`
			} `json:"dswBrand"`
			AfterPayInstallmentPrice float64 `json:"afterPayInstallmentPrice"`
			ID                       string  `json:"id"`
			IsPreOrder               bool    `json:"isPreOrder"`
			BvReviewCount            int     `json:"bvReviewCount"`
			BvRating                 float64 `json:"bvRating"`
			ShowWidth                bool    `json:"showWidth"`
			ParentCategories         []struct {
				DisplayName           string      `json:"displayName"`
				Description           interface{} `json:"description"`
				DefaultParentCategory interface{} `json:"defaultParentCategory"`
				ID                    string      `json:"id"`
				Type                  string      `json:"type"`
			} `json:"parentCategories"`
			Bullets         []string `json:"bullets"`
			Productitemtype []struct {
				DisplayName string `json:"displayName"`
			} `json:"productitemtype"`
			ChildSKUs []struct {
				IsClearanceItem bool `json:"isClearanceItem"`
				Color           struct {
					DisplayName string `json:"displayName"`
					ColorCode   string `json:"colorCode"`
				} `json:"color"`
				OriginalPrice  float64 `json:"originalPrice"`
				NonMemberPrice float64 `json:"nonMemberPrice"`
				Upc            string  `json:"upc"`
				IsDropShipItem bool    `json:"isDropShipItem"`
				Size           struct {
					DisplayName string  `json:"displayName"`
					SizeCode    float64 `json:"sizeCode"`
				} `json:"size"`
				Materials []struct {
					DisplayName string `json:"displayName"`
				} `json:"materials"`
				IsPreOrderItem bool   `json:"isPreOrderItem"`
				ID             string `json:"id"`
				Dimension      struct {
					DisplayName      string  `json:"displayName"`
					DimensionCode    string  `json:"dimensionCode"`
					DimensionSeqCode float64 `json:"dimensionSeqCode"`
				} `json:"dimension"`
				SkuStockLevel      int  `json:"skuStockLevel"`
				IsExclusiveLicense bool `json:"isExclusiveLicense"`
			} `json:"childSKUs"`
			SpinColorCode       string `json:"spinColorCode"`
			HasAnimatedImage    bool   `json:"hasAnimatedImage"`
			ProductTypeWeb      string `json:"productTypeWeb"`
			RecommendationsURL  string `json:"recommendationsUrl"`
			ProductStockLevel   int    `json:"productStockLevel"`
			ExpeditedRestricted bool   `json:"expeditedRestricted"`
			DefaultSKU          struct {
				IsClearanceItem bool `json:"isClearanceItem"`
				Color           struct {
					DisplayName string `json:"displayName"`
					ColorCode   string `json:"colorCode"`
				} `json:"color"`
				OriginalPrice  float64 `json:"originalPrice"`
				NonMemberPrice float64 `json:"nonMemberPrice"`
				Upc            string  `json:"upc"`
				IsDropShipItem bool    `json:"isDropShipItem"`
				Size           struct {
					DisplayName string  `json:"displayName"`
					SizeCode    float64 `json:"sizeCode"`
				} `json:"size"`
				Materials []struct {
					DisplayName string `json:"displayName"`
				} `json:"materials"`
				IsPreOrderItem bool   `json:"isPreOrderItem"`
				ID             string `json:"id"`
				Dimension      struct {
					DisplayName      string  `json:"displayName"`
					DimensionCode    string  `json:"dimensionCode"`
					DimensionSeqCode float64 `json:"dimensionSeqCode"`
				} `json:"dimension"`
				SkuStockLevel      int  `json:"skuStockLevel"`
				IsExclusiveLicense bool `json:"isExclusiveLicense"`
			} `json:"defaultSKU"`
			ProductGender      string  `json:"productGender"`
			PriceInCart        bool    `json:"priceInCart"`
			NonMemberMinPrice  float64 `json:"nonMemberMinPrice"`
			AncestorCategories []struct {
				DisplayName string `json:"displayName"`
				ID          string `json:"id"`
			} `json:"ancestorCategories"`
			IsGWPItem bool `json:"isGWPItem"`
			Style     []struct {
				DisplayName string `json:"displayName"`
			} `json:"style"`
			CurrencyCode string `json:"currencyCode"`
			Breadcrumbs  []struct {
				Text string `json:"text"`
				URL  string `json:"url"`
			} `json:"breadcrumbs"`
		} `json:"product"`
	} `json:"Response"`
}

var (
	productsExtractReg  = regexp.MustCompile(`(.*)`)
	viewData parseProductResponse
)

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

	if err := json.Unmarshal(matched[0], &viewData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}

	//Prepare product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:       strconv.Format(viewData.Response.Product.ID),
			CrawlUrl: resp.Request.URL.String(),
		},
		BrandName:   viewData.Response.Product.DswBrand.DisplayNameDefault,
		Title:       viewData.Response.Product.DisplayName,
		Description: viewData.Response.Product.LongDescription,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stats: &pbItem.Stats{
			ReviewCount: int32(viewData.Response.Product.BvReviewCount),
			Rating:      float32(viewData.Response.Product.BvRating),
		},
	}

	for _, rawSku := range viewData.Response.Product.ChildSKUs {
		currentPrice, _ := strconv.ParseFloat(rawSku.OriginalPrice)
		originalPrice, _ := strconv.ParseFloat(rawSku.NonMemberPrice)
		//discount, _ := strconv.ParseInt(strings.TrimSuffix(rawSku.DisplayPercentOff, "%"))
		sku := pbItem.Sku{
			SourceId: strconv.Format(rawSku.Upc),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(currentPrice * 100),
				Msrp:     int32(originalPrice * 100),
				//Discount: int32(discount),
			},
			Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}
		if rawSku.SkuStockLevel > 0 {
			sku.Stock.StockStatus = pbItem.Stock_InStock
			sku.Stock.StockCount = int32(rawSku.SkuStockLevel)
		}

		// color
			
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    strconv.Format(rawSku.Color.ColorCode),
				Name:  rawSku.Color.DisplayName,
				Value: rawSku.Color.DisplayName,
				//Icon:  color.SwatchMedia.Mobile,
			})

		// size		
			
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    strconv.Format(rawSku.Size.SizeCode),
				Name:  rawSku.Size.DisplayName,
				Value: rawSku.Size.DisplayName,
			})

		
				// for _, mid := range color.StyleMediaIds {
				// 	m := p.StyleMedia.ByID[strconv.Format(mid)]
				// 	if m.MediaType == "Image" {
				// 		sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
				// 			strconv.Format(m.ID),
				// 			m.ImageMediaURI.MaxLargeDesktop,
				// 			m.ImageMediaURI.SmallZoom,
				// 			m.ImageMediaURI.MobileLarge,
				// 			m.ImageMediaURI.MobileMedium,
				// 			"",
				// 			m.IsDefault,
				// 		))
				// 	} else if m.MediaType == "Video" {
				// 		// TODO
				// 	}
				// }

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
		"https://www.dsw.com/api/v1/content/pages/_/N-1z141hwZ1z128ujZ1z141ju?pagePath=/pages/DSW/category&skipHeaderFooterContent=true&No=0&locale=en_US&pushSite=DSW&tier=GUEST",
		
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

	if apiToken == "" || jsToken == "" {
		panic("env PC_API_TOKEN or PC_JS_TOKEN is not set")
	}

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

	for _, req := range spider.NewTestRequest(context.Background()) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
		defer cancel()

		logger.Debugf("Access %s", req.URL)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.96 Safari/537.36")
		for k := range opts.MustHeader {
			req.Header.Set(k, opts.MustHeader.Get(k))
		}
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
	}
}
