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
		categoryPathMatcher: regexp.MustCompile(`^((\?!product).)*`),
		// this regular used to match product page url path
		productPathMatcher: regexp.MustCompile(`(/[a-z0-9_-]+)?/product([/a-z0-9_-]+)`),
		logger:             logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "190667e8cdcf4002add19be8190bfd38"
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
	return []string{"*.theoutnet.com"}
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
	}	else if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
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
		Plp struct {
			Listing struct {
				VisibleProducts []struct {
					Products []struct {						
						Seo struct {
							SeoURLKeyword string `json:"seoURLKeyword"`
						} `json:"seo"`
					} `json:"products"`					
				} `json:"visibleProducts"`
				Response struct {
					Body struct {
						RecordSetTotal int `json:"recordSetTotal"`						
						TotalPages   int         `json:"totalPages"`
					} `json:"body"`					
				} `json:"response"`				
			} `json:"listing"`			
		} `json:"plp"`		
}

// used to extract embaded json data in website page.
// more about golang regulation see here https://golang.org/pkg/regexp/syntax/
var productsExtractReg = regexp.MustCompile(`window.state=({.*})</script>`)

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
	for _, idv := range viewData.Plp.Listing.VisibleProducts[0].Products {
		
		rawurl := fmt.Sprintf("%s://%s%s", resp.Request.URL.Scheme, resp.Request.URL.Host, idv.Seo.SeoURLKeyword)
		
		fmt.Println(rawurl)
		// req, err := http.NewRequest(http.MethodGet, rawurl, nil)
		// if err != nil {
		// 	c.logger.Errorf("load http request of url %s failed, error=%s", rawurl, err)
		// 	return err
		// }

		// lastIndex += 1
		// // set the index of the product crawled in the sub response
		// nctx := context.WithValue(ctx, "item.index", lastIndex)
		// // yield sub request
		// if err := yield(nctx, req); err != nil {
		// 	return err
		// }
	}

	// get current page number
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("pageNumber"))
	if page == 0 {
		page = 1
	}
	// check if this is the last page
	if len(viewData.Plp.Listing.VisibleProducts[0].Products) > (viewData.Plp.Listing.Response.Body.RecordSetTotal) ||
		page >= int64(viewData.Plp.Listing.Response.Body.TotalPages) {
		return nil
	}

	// set pagination
	u := *resp.Request.URL
	vals := u.Query()
	vals.Set("pageNumber", strconv.Format(page+1))
	u.RawQuery = vals.Encode()

	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	// update the index of last page
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
}

type parseProductResponse struct {
	
	Pdp struct {
		DetailsState struct {
			Response struct {
				Body struct {					
					Products	[]struct {
						Dynamic     bool `json:"dynamic"`
						Visible     bool `json:"visible"`
						DesignerSeo struct {
							SeoURLKeyword string `json:"seoURLKeyword"`
						} `json:"designerSeo"`
						Displayable              bool     `json:"displayable"`
						DesignerNameEN           string   `json:"designerNameEN"`
						Type                     string   `json:"type"`
						ExternalReccomendationID []string `json:"externalReccomendationId"`
						Name                     string   `json:"name"`
						DesignerIdentifier       string   `json:"designerIdentifier"`
						SHOES1                   struct {
							Values []struct {
								Values []struct {
									Label      string `json:"label"`
									Identifier string `json:"identifier"`
								} `json:"values"`
								Usage      string `json:"usage"`
								Label      string `json:"label"`
								Identifier string `json:"identifier"`
							} `json:"values"`
						} `json:"SHOES_1"`
						ForceLogIn   bool   `json:"forceLogIn"`
						MfPartNumber string `json:"mfPartNumber"`
						SHOES5       struct {
							Values []struct {
								Values []struct {
									Label      string `json:"label"`
									Identifier string `json:"identifier"`
								} `json:"values"`
								Usage      string `json:"usage"`
								Label      string `json:"label"`
								Identifier string `json:"identifier"`
							} `json:"values"`
						} `json:"SHOES_5"`
						SHOES3 struct {
							Values []struct {
								Values []struct {
									Label      string `json:"label"`
									Identifier string `json:"identifier"`
								} `json:"values"`
								Usage      string `json:"usage"`
								Label      string `json:"label"`
								Identifier string `json:"identifier"`
							} `json:"values"`
						} `json:"SHOES_3"`
						MF1 struct {
							Values []struct {
								Values []struct {
									Label      string `json:"label"`
									Identifier string `json:"identifier"`
								} `json:"values"`
								Usage      string `json:"usage"`
								Label      string `json:"label"`
								Identifier string `json:"identifier"`
							} `json:"values"`
						} `json:"M&F_1"`
						PartNumber string `json:"partNumber"`
						SHOES4     struct {
							Values []struct {
								Values []struct {
									Label      string `json:"label"`
									Identifier string `json:"identifier"`
								} `json:"values"`
								Usage      string `json:"usage"`
								Label      string `json:"label"`
								Identifier string `json:"identifier"`
							} `json:"values"`
						} `json:"SHOES_4"`
						WCSGRPFITDETAILS struct {
							Values []struct {
								Values []struct {
									Label      string `json:"label"`
									Identifier string `json:"identifier"`
								} `json:"values"`
								Usage      string `json:"usage"`
								Label      string `json:"label"`
								Identifier string `json:"identifier"`
							} `json:"values"`
						} `json:"WCS_GRP_FIT_DETAILS"`
						ProductColours []struct {
							Visible              bool   `json:"visible"`
							EditorialDescription string `json:"editorialDescription"`
							DetailsAndCare       string `json:"detailsAndCare"`
							Displayable          bool   `json:"displayable"`
							Type                 string `json:"type"`
							Swatch               struct {
								HEX string `json:"HEX"`
							} `json:"swatch"`
							ExternalReccomendationID []string  `json:"externalReccomendationId"`
							TechnicalDescription     string    `json:"technicalDescription"`
							Selected                 bool      `json:"selected"`
							IsDefault                bool      `json:"isDefault"`
							ForceLogIn               bool      `json:"forceLogIn"`
							MfPartNumber             string    `json:"mfPartNumber"`
							PartNumber               string    `json:"partNumber"`
							ImageViews               []string  `json:"imageViews"`
							FirstVisibleDate         time.Time `json:"firstVisibleDate"`
							LowStockOnline           bool      `json:"lowStockOnline"`
							Label                    string    `json:"label"`
							SKUs                     []struct {
								SkuUniqueID        int64 `json:"skuUniqueID"`
								WCSGRPMEASUREMENTS struct {
									Values []struct {
										Values []struct {
											Label      string `json:"label"`
											Identifier string `json:"identifier"`
										} `json:"values"`
										Usage      string `json:"usage"`
										Label      string `json:"label"`
										Identifier string `json:"identifier"`
									} `json:"values"`
								} `json:"WCS_GRP_MEASUREMENTS"`
								Displayable bool   `json:"displayable"`
								Type        string `json:"type"`
								COO         struct {
									Values []struct {
										Values []struct {
											Label      string `json:"label"`
											Identifier string `json:"identifier"`
										} `json:"values"`
										Usage      string `json:"usage"`
										Label      string `json:"label"`
										Identifier string `json:"identifier"`
									} `json:"values"`
								} `json:"COO_!"`
								Banned bool `json:"banned"`
								Size   struct {
									CentralSizeLabel string `json:"centralSizeLabel"`
									Schemas          []struct {
										Name   string   `json:"name"`
										Labels []string `json:"labels"`
									} `json:"schemas"`
									ScaleLabel string `json:"scaleLabel"`
									LabelSize  string `json:"labelSize"`
								} `json:"size"`
								SoldOutOnline bool `json:"soldOutOnline,omitempty"`
								Badges        []struct {
									Label string `json:"label"`
									Type  string `json:"type"`
									Key   string `json:"key"`
								} `json:"badges,omitempty"`
								Selected bool `json:"selected"`
								Price    struct {
									SellingPrice struct {
										Amount  int `json:"amount"`
										Divisor int `json:"divisor"`
									} `json:"sellingPrice"`
									RdSellingPrice struct {
										Amount  int `json:"amount"`
										Divisor int `json:"divisor"`
									} `json:"rdSellingPrice"`
									RdWasPrice struct {
										Amount  int `json:"amount"`
										Divisor int `json:"divisor"`
									} `json:"rdWasPrice"`
									WasPrice struct {
										Amount  int `json:"amount"`
										Divisor int `json:"divisor"`
									} `json:"wasPrice"`
									RdDiscount struct {
										Amount  int `json:"amount"`
										Divisor int `json:"divisor"`
									} `json:"rdDiscount"`
									Currency struct {
										Symbol string `json:"symbol"`
										Label  string `json:"label"`
									} `json:"currency"`
									Discount struct {
										Amount  int `json:"amount"`
										Divisor int `json:"divisor"`
									} `json:"discount"`
								} `json:"price"`
								Buyable     bool   `json:"buyable"`
								Composition string `json:"composition"`
								ForceLogIn  bool   `json:"forceLogIn"`
								Attributes  []struct {
									Values []struct {
										Label      string `json:"label"`
										Identifier string `json:"identifier"`
									} `json:"values"`
									Usage      string `json:"usage"`
									Label      string `json:"label"`
									Identifier string `json:"identifier"`
								} `json:"attributes"`
								PartNumber       string `json:"partNumber"`
								NotStockedOnline bool   `json:"notStockedOnline,omitempty"`
								OneLeftOnline    bool   `json:"oneLeftOnline,omitempty"`
								LowStockOnline   bool   `json:"lowStockOnline,omitempty"`
							} `json:"sKUs"`
							Banned    bool  `json:"banned"`
							ProductID int64 `json:"productId"`
							Price     struct {
								SellingPrice struct {
									Amount  int `json:"amount"`
									Divisor int `json:"divisor"`
								} `json:"sellingPrice"`
								RdSellingPrice struct {
									Amount  int `json:"amount"`
									Divisor int `json:"divisor"`
								} `json:"rdSellingPrice"`
								RdWasPrice struct {
									Amount  int `json:"amount"`
									Divisor int `json:"divisor"`
								} `json:"rdWasPrice"`
								WasPrice struct {
									Amount  int `json:"amount"`
									Divisor int `json:"divisor"`
								} `json:"wasPrice"`
								RdDiscount struct {
									Amount  int `json:"amount"`
									Divisor int `json:"divisor"`
								} `json:"rdDiscount"`
								Currency struct {
									Symbol string `json:"symbol"`
									Label  string `json:"label"`
								} `json:"currency"`
								Discount struct {
									Amount  int `json:"amount"`
									Divisor int `json:"divisor"`
								} `json:"discount"`
							} `json:"price"`
							ImageTemplate    string `json:"imageTemplate"`
							ShortDescription string `json:"shortDescription"`
							Buyable          bool   `json:"buyable"`
							Seo              struct {
								SeoURLKeyword string `json:"seoURLKeyword"`
							} `json:"seo"`
							Attributes []struct {
								Values []struct {
									Label      string `json:"label"`
									Identifier string `json:"identifier"`
								} `json:"values"`
								Usage      string `json:"usage"`
								Label      string `json:"label"`
								Identifier string `json:"identifier"`
							} `json:"attributes"`
							Identifier string   `json:"identifier"`
							ImageIds   []string `json:"imageIds"`
						} `json:"productColours"`
						SizeAndFit        string    `json:"sizeAndFit"`
						CentralSizeScheme string    `json:"centralSizeScheme"`
						FirstVisibleDate  time.Time `json:"firstVisibleDate"`
						LowStockOnline    bool      `json:"lowStockOnline"`
						MasterCategory    struct {
							Child struct {
								LabelEN    string `json:"labelEN"`
								CategoryID string `json:"categoryId"`
								Label      string `json:"label"`
								Identifier string `json:"identifier"`
							} `json:"child"`
							LabelEN    string `json:"labelEN"`
							CategoryID string `json:"categoryId"`
							Label      string `json:"label"`
							Identifier string `json:"identifier"`
						} `json:"masterCategory"`
						ProductID string `json:"productId"`
						Price     struct {
							SellingPrice struct {
								Amount  int `json:"amount"`
								Divisor int `json:"divisor"`
							} `json:"sellingPrice"`
							RdSellingPrice struct {
								Amount  int `json:"amount"`
								Divisor int `json:"divisor"`
							} `json:"rdSellingPrice"`
							RdWasPrice struct {
								Amount  int `json:"amount"`
								Divisor int `json:"divisor"`
							} `json:"rdWasPrice"`
							WasPrice struct {
								Amount  int `json:"amount"`
								Divisor int `json:"divisor"`
							} `json:"wasPrice"`
							RdDiscount struct {
								Amount  int `json:"amount"`
								Divisor int `json:"divisor"`
							} `json:"rdDiscount"`
							Currency struct {
								Symbol string `json:"symbol"`
								Label  string `json:"label"`
							} `json:"currency"`
							Discount struct {
								Amount  int `json:"amount"`
								Divisor int `json:"divisor"`
							} `json:"discount"`
						} `json:"price"`
						Thumbnail string `json:"thumbnail"`
						Tracking  struct {
							PrimaryCategory struct {
								Child struct {
									Child struct {
										CategoryID string `json:"categoryId"`
										Label      string `json:"label"`
										Identifier string `json:"identifier"`
									} `json:"child"`
									CategoryID string `json:"categoryId"`
									Label      string `json:"label"`
									Identifier string `json:"identifier"`
								} `json:"child"`
								CategoryID string `json:"categoryId"`
								Label      string `json:"label"`
								Identifier string `json:"identifier"`
							} `json:"primaryCategory"`
							DesignerName string `json:"designerName"`
							Name         string `json:"name"`
						} `json:"tracking"`
						DesignerName string `json:"designerName"`
						Buyable      bool   `json:"buyable"`
						Images       []struct {
							ID   string `json:"id"`
							View string `json:"view"`
							URL  string `json:"url"`
							Size struct {
								Height int `json:"height"`
								Width  int `json:"width"`
							} `json:"size"`
						} `json:"images"`
						Seo struct {
							Title           string `json:"title"`
							AlternateText   string `json:"alternateText"`
							MetaDescription string `json:"metaDescription"`
							MetaKeyword     string `json:"metaKeyword"`
							SeoURLKeyword   string `json:"seoURLKeyword"`
						} `json:"seo"`
						Attributes []struct {
							Values []struct {
								Label      string `json:"label"`
								Identifier string `json:"identifier"`
							} `json:"values"`
							Usage      string `json:"usage"`
							Label      string `json:"label"`
							Identifier string `json:"identifier"`
						} `json:"attributes"`
						SalesCategories []struct {
							Child struct {
								Child struct {
									CategoryID string `json:"categoryId"`
									Label      string `json:"label"`
									Seo        struct {
										SeoURLKeyword string `json:"seoURLKeyword"`
									} `json:"seo"`
									Identifier string `json:"identifier"`
								} `json:"child"`
								CategoryID string `json:"categoryId"`
								Label      string `json:"label"`
								Seo        struct {
									SeoURLKeyword string `json:"seoURLKeyword"`
								} `json:"seo"`
								Identifier string `json:"identifier"`
							} `json:"child"`
							Primary    bool   `json:"primary"`
							CategoryID string `json:"categoryId"`
							Label      string `json:"label"`
							Seo        struct {
								SeoURLKeyword string `json:"seoURLKeyword"`
							} `json:"seo"`
							Identifier string `json:"identifier"`
						} `json:"salesCategories"`
					} `json:"products"`
					
				} `json:"body"`				
			} `json:"response"`			
		} `json:"detailsState"`		
	} `json:"pdp"`

}

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

	var viewData parseProductResponse

	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}

	for _, p := range viewData.Pdp.DetailsState.Response.Body.Products[0].ProductColours {
		// build product data

		if !p.Visible {
			continue
		}

		item := pbItem.Product{
			Source: &pbItem.Source{
				Id:       strconv.Format(p.PartNumber),
				CrawlUrl: resp.Request.URL.String(),
			},
			BrandName:   viewData.Pdp.DetailsState.Response.Body.Products[0].DesignerName,
			Title:       p.ShortDescription,
			Description: p.TechnicalDescription,
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
			},			
		}

		for key, rawSku := range p.SKUs {
			originalPrice, _ := strconv.ParseFloat(rawSku.Price.WasPrice.Amount)  //  / 100
			currentPrice, _ := strconv.ParseFloat(rawSku.Price.SellingPrice.Amount)
			discount, _ := strconv.ParseFloat(rawSku.Price.Discount.Amount)
			sku := pbItem.Sku{
				SourceId: strconv.Format(rawSku.PartNumber),
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(currentPrice),
					Msrp:     int32(originalPrice),
					Discount: int32(discount),
				},
				Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}
			if rawSku.Selected {
				sku.Stock.StockStatus = pbItem.Stock_InStock
				//sku.Stock.StockCount = int32(rawSku.TotalQuantityAvailable)
			}

			// color			
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecColor,
					Id:    strconv.Format(p.PartNumber),
					Name:  p.Label,
					Value: p.Label,
					//Icon:  color.SwatchMedia.Mobile,
				})				
			

			// size			
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecSize,
					Id:    rawSku.PartNumber,
					Name:  rawSku.Size.LabelSize,
					Value: rawSku.Size.LabelSize,
				})
			

			if key == 0	{
				// image  // please check
				isDefault := true
				for ki, mid := range p.ImageViews {
					if ki > 0{
						isDefault = false
					}
					imgTemplate := strings.ReplaceAll(p.ImageTemplate,"{view}",mid)
					sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
						strconv.Format(mid),
						strings.ReplaceAll(imgTemplate, "{width}","920"),
						strings.ReplaceAll(imgTemplate, "{width}","920"),
						strings.ReplaceAll(imgTemplate, "{width}","600"),
						strings.ReplaceAll(imgTemplate, "{width}","400"),
						"",
						isDefault,
					))
				}
			}		

			item.SkuItems = append(item.SkuItems, &sku)			
		}

		fmt.Println(&item)
		// yield item result
		if err = yield(ctx, &item); err != nil {
			return err
		}
	}
	return nil
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://www.theoutnet.com/en-in/shop/clothing/jeans",
		"https://www.theoutnet.com/en-us/shop/product/acne-studios/jeans/straight-leg-jeans/log-high-rise-straight-leg-jeans/17476499598965898",
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
				i.URL.Host = "www.nordstrom.com"
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

	ctx := context.WithValue(context.Background(), "tracing_id", "nordstrom_123456")
	// start the crawl request
	for _, req := range spider.NewTestRequest(context.Background()) {
		if err := callback(ctx, req); err != nil {
			logger.Fatal(err)
		}
	}
}
