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
	pbProxy "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
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
		categoryPathMatcher: regexp.MustCompile(`^/(browse|brands)(/[a-z0-9-]+){2,6}$`),
		// this regular used to match product page url path
		productPathMatcher: regexp.MustCompile(`^/s(/[a-z0-9-]+){1,3}/[0-9]+$`),
		logger:             logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "8651bc392ba4680d2424e0f382b7b940"
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
		EnableSessionInit: true,
		// Reliability high match to crawl api, and log match to backconnect
		Reliability: pbProxy.ProxyReliability_ReliabilityMedium,
	}
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.nordstrom.com"}
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
		u.Host = "www.nordstrom.com"
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

	if resp.Request.URL.Path == "/invitation.html" {
		return fmt.Errorf("robot forbidden")
	}

	p := strings.TrimSuffix(resp.RawUrl().Path, "/")

	if p == "" {
		return c.parseCategories(ctx, resp, yield)
	}
	if c.categoryPathMatcher.MatchString(resp.RawUrl().Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	} else if c.productPathMatcher.MatchString(resp.RawUrl().Path) {
		return c.parseProduct(ctx, resp, yield)
	}

	return crawler.ErrUnsupportedPath
}

func (c *_Crawler) parseCategories(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}
	respBody, err := resp.RawBody()
	if err != nil {
		return err
	}

	matched := categoryExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
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

type RawColor struct {
	ID             string   `json:"id"`
	Label          string   `json:"label"`
	SpriteIndex    int      `json:"spriteIndex"`
	MediaIds       []string `json:"mediaIds"`
	StandardColors []string `json:"standardColors"`
	SwatchMediaID  string   `json:"swatchMediaId"`
}

type RawMedia struct {
	ID    string `json:"id"`
	Group string `json:"group"`
	Type  string `json:"type"`
	Src   string `json:"src"`
}

type RawProduct struct {
	ID          int    `json:"id"`
	BrandID     int    `json:"brandId"`
	BrandName   string `json:"brandName"`
	StyleNumber string `json:"styleNumber"`
	Name        string `json:"name"`
	PricesByID  struct {
		Original struct {
			ID                  string      `json:"id"`
			MinItemPercentOff   int         `json:"minItemPercentOff"`
			MaxItemPercentOff   int         `json:"maxItemPercentOff"`
			MinItemPrice        string      `json:"minItemPrice"`
			MaxItemPrice        string      `json:"maxItemPrice"`
			PriceValidUntilDate interface{} `json:"priceValidUntilDate"`
		} `json:"original"`
	} `json:"pricesById"`
	ProductPageURL   string  `json:"productPageUrl"`
	ReviewCount      int     `json:"reviewCount"`
	ReviewStarRating float64 `json:"reviewStarRating"`
}

type RawViewData struct {
	ProductsByID map[string]RawProduct `json:"productsById"`
	Query        struct {
		PageCount        int      `json:"pageCount"`
		ProductOffset    int      `json:"productOffset"`
		PageProductCount int      `json:"pageProductCount"`
		PageSelected     int      `json:"pageSelected"`
		ResultCount      int      `json:"resultCount"`
		ResultProductIds []int    `json:"resultProductIds"`
		SortSelectedID   string   `json:"sortSelectedId"`
		SortOptionIds    []string `json:"sortOptionIds"`
	} `json:"query"`
}

type RawProductDetail struct {
	ID                   int      `json:"id"`
	AgeGroups            []string `json:"ageGroups"`
	ReviewAverageRating  float64  `json:"reviewAverageRating"`
	EnticementPlacements []struct {
		Name        string `json:"name"`
		Enticements []struct {
			Type  string `json:"type"`
			Title string `json:"title"`
		} `json:"enticements"`
	} `json:"enticementPlacements"`
	Brand struct {
		BrandName  string `json:"brandName"`
		BrandURL   string `json:"brandUrl"`
		ImsBrandID int    `json:"imsBrandId"`
	} `json:"brand"`
	Consumers           []string `json:"consumers"`
	Description         string   `json:"description"`
	CustomizationCode   string   `json:"customizationCode"`
	DefaultGalleryMedia struct {
		StyleMediaID  int    `json:"styleMediaId"`
		ColorID       string `json:"colorId"`
		IsTrimmed     bool   `json:"isTrimmed"`
		StyleMediaIds []int  `json:"styleMediaIds"`
	} `json:"defaultGalleryMedia"`
	Features []string `json:"features"`
	Filters  struct {
		Color struct {
			ByID map[string]struct {
				//Num001 struct {
				ID                   string      `json:"id"`
				Code                 string      `json:"code"`
				IsSelected           bool        `json:"isSelected"`
				IsDefault            bool        `json:"isDefault"`
				Value                string      `json:"value"`
				DisplayValue         string      `json:"displayValue"`
				FilterType           string      `json:"filterType"`
				IsAvailableWith      string      `json:"isAvailableWith"`
				RelatedSkuIds        []int       `json:"relatedSkuIds"`
				SoldOutRelatedSkuIds interface{} `json:"soldOutRelatedSkuIds"`
				StyleMediaIds        []int       `json:"styleMediaIds"`
				SwatchMedia          struct {
					Desktop string `json:"desktop"`
					Mobile  string `json:"mobile"`
					Preview string `json:"preview"`
				} `json:"swatchMedia"`
				//} `json:"001"`
			} `json:"byId"`
			AllIds []string `json:"allIds"`
		} `json:"color"`
		Size struct {
			ByID map[string]struct {
				//BigKid35M struct {
				ID                   string      `json:"id"`
				Value                string      `json:"value"`
				DisplayValue         string      `json:"displayValue"`
				GroupValue           string      `json:"groupValue"`
				FilterType           string      `json:"filterType"`
				RelatedSkuIds        []int       `json:"relatedSkuIds"`
				IsAvailableWith      string      `json:"isAvailableWith"`
				SoldOutRelatedSkuIds interface{} `json:"soldOutRelatedSkuIds"`
				//} `json:"big kid-3.5 m"`
			} `json:"byId"`
			AllIds []interface{} `json:"allIds"`
		} `json:"size"`
		Width struct {
			ByID struct {
			} `json:"byId"`
			AllIds []interface{} `json:"allIds"`
		} `json:"width"`
	} `json:"filters"`
	FilterOptions      []string `json:"filterOptions"`
	FitCategory        string   `json:"fitCategory"`
	Gender             string   `json:"gender"`
	ProductTypeCode    string   `json:"productTypeCode"`
	Ingredients        string   `json:"ingredients"`
	IsAnniversaryStyle bool     `json:"isAnniversaryStyle"`
	IsAvailable        bool     `json:"isAvailable"`
	NumberOfReviews    int      `json:"numberOfReviews"`
	PathAlias          string   `json:"pathAlias"`
	ProductEngagement  struct {
		PageViews struct {
			Count string `json:"count"`
			Copy  string `json:"copy"`
		} `json:"pageViews"`
	} `json:"productEngagement"`
	ProductName            string        `json:"productName"`
	ProductTitle           string        `json:"productTitle"`
	ProductTypeName        string        `json:"productTypeName"`
	ProductTypeParentName  string        `json:"productTypeParentName"`
	SalesVideoShot         interface{}   `json:"salesVideoShot"`
	SellingStatement       string        `json:"sellingStatement"`
	ShopperSizePreferences []interface{} `json:"shopperSizePreferences"`
	Skus                   struct {
		ByID map[int]struct {
			//Num18432420 struct {
			ID                     int         `json:"id"`
			BackOrderDate          interface{} `json:"backOrderDate"`
			ColorID                string      `json:"colorId"`
			DisplayPercentOff      string      `json:"displayPercentOff"`
			DisplayPrice           string      `json:"displayPrice"`
			IsAvailable            bool        `json:"isAvailable"`
			IsBackOrder            bool        `json:"isBackOrder"`
			Price                  float64     `json:"price"`
			SizeID                 string      `json:"sizeId"`
			WidthID                string      `json:"widthId"`
			RmsSkuID               int         `json:"rmsSkuId"`
			TotalQuantityAvailable int         `json:"totalQuantityAvailable"`
			IsFinalSale            bool        `json:"isFinalSale"`
			IsClearTheRack         bool        `json:"isClearTheRack"`
			//} `json:"18432420"`

		} `json:"byId"`
		AllIds []int `json:"allIds"`
	} `json:"skus"`

	StyleMedia struct {
		ByID map[int]struct {
			//Num2827730 struct {
			ID            int    `json:"id"`
			ColorID       string `json:"colorId"`
			ColorName     string `json:"colorName"`
			ImageMediaURI struct {
				SmallDesktop string `json:"smallDesktop"`
				LargeDesktop string `json:"largeDesktop"`
				Zoom         string `json:"zoom"`
				MobileSmall  string `json:"mobileSmall"`
				MobileZoom   string `json:"mobileZoom"`
				Mini         string `json:"mini"`
			} `json:"imageMediaUri"`
			IsDefault      bool   `json:"isDefault"`
			IsSelected     bool   `json:"isSelected"`
			IsTrimmed      bool   `json:"isTrimmed"`
			MediaGroupType string `json:"mediaGroupType"`
			MediaType      string `json:"mediaType"`
			SortID         int    `json:"sortId"`
			//} `json:"2827730"`
		} `json:"byId"`
		AllIds []int `json:"allIds"`
	} `json:"styleMedia"`
}

type CategoryView struct {
	OperatingCountryCode string      `json:"operatingCountryCode"`
	ProductResults       RawViewData `json:"productResults"`
	ViewData             RawViewData `json:"viewData"`
}

// used to extract embaded json data in website page.
// more about golang regulation see here https://golang.org/pkg/regexp/syntax/
var productsExtractReg = regexp.MustCompile(`(?U)<script\s*>window.__INITIAL_CONFIG__\s*=\s*({.*});?\s*</script>`)
var categoryExtractReg = regexp.MustCompile(`(?U)<script\s*>window.__INITIAL_CONFIG__\s*=\s*(.*)</script>`)

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
		if err := c.httpClient.Jar().Clear(ctx, resp.Request.URL); err != nil {
			c.logger.Errorf("clear cookie for %s failed, error=%s", resp.Request.URL, err)
		}
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}
	// c.logger.Debugf("data: %s", matched[1])

	var viewData CategoryView
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		return err
	}

	lastIndex := nextIndex(ctx)
	for _, idv := range viewData.ViewData.Query.ResultProductIds {
		p, ok := viewData.ViewData.ProductsByID[fmt.Sprintf("%d", idv)]
		if !ok {
			c.logger.Warnf("product %v not found", idv)
			continue
		}

		req, err := http.NewRequest(http.MethodGet, p.ProductPageURL, nil)
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
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}
	// check if this is the last page
	if len(viewData.ViewData.Query.ResultProductIds) < viewData.ViewData.Query.PageProductCount ||
		page >= int64(viewData.ViewData.Query.PageCount) {
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
		// clean cookie
		if err := c.httpClient.Jar().Clear(ctx, resp.Request.URL); err != nil {
			c.logger.Errorf("clear cookie for %s failed, error=%s", resp.Request.URL, err)
		}
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}
	// c.logger.Debugf("data: %s", matched[1])

	var viewData struct {
		StylesById struct {
			Data map[string]RawProductDetail `json:"data"`
		} `json:"stylesById"`
	}
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

	for _, p := range viewData.StylesById.Data {
		item := pbItem.Product{
			Source: &pbItem.Source{
				Id:           strconv.Format(p.ID),
				CrawlUrl:     resp.Request.URL.String(),
				CanonicalUrl: canUrl,
			},
			CrowdType:   p.Gender,
			BrandName:   p.Brand.BrandName,
			Title:       p.ProductName,
			Description: htmlTrimRegp.ReplaceAllString(p.Description, "") + " " + strings.Join(p.Features, " "),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
			},
			Stock: &pbItem.Stock{
				StockStatus: pbItem.Stock_OutOfStock,
			},
			Stats: &pbItem.Stats{
				ReviewCount: int32(p.NumberOfReviews),
				Rating:      float32(p.ReviewAverageRating / 5.0),
			},
		}
		// TODO: no category found

		for _, mid := range p.DefaultGalleryMedia.StyleMediaIds {
			m := p.StyleMedia.ByID[(mid)]
			if m.MediaType == "Image" {
				item.Medias = append(item.Medias, pbMedia.NewImageMedia(
					strconv.Format(m.ID),
					m.ImageMediaURI.LargeDesktop,
					m.ImageMediaURI.Zoom,
					m.ImageMediaURI.Mini,
					m.ImageMediaURI.SmallDesktop,
					"",
					m.IsDefault && m.IsSelected,
				))
			}
		}
		if p.IsAvailable {
			item.Stock.StockStatus = pbItem.Stock_InStock
		}
		for _, rawSku := range p.Skus.ByID {
			originalPrice, _ := strconv.ParseFloat(rawSku.DisplayPrice)
			discount, _ := strconv.ParseInt(strings.TrimSuffix(rawSku.DisplayPercentOff, "%"))
			sku := pbItem.Sku{
				SourceId: strconv.Format(rawSku.ID),
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(rawSku.Price * 100),
					Msrp:     int32(originalPrice * 100),
					Discount: int32(discount),
				},
				Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}
			if rawSku.TotalQuantityAvailable > 0 {
				sku.Stock.StockStatus = pbItem.Stock_InStock
				sku.Stock.StockCount = int32(rawSku.TotalQuantityAvailable)
			}

			// color

			color := p.Filters.Color.ByID[strconv.Format(rawSku.ColorID)]
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    strconv.Format(color.ID),
				Name:  color.DisplayValue,
				Value: color.Value,
				Icon:  color.SwatchMedia.Mobile,
			})
			for k, mid := range color.StyleMediaIds {
				m := p.StyleMedia.ByID[mid]
				if m.MediaType == "Image" {
					sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
						strconv.Format(m.ID),
						m.ImageMediaURI.LargeDesktop,
						m.ImageMediaURI.Zoom,
						m.ImageMediaURI.Mini,
						m.ImageMediaURI.SmallDesktop,
						"",
						k == 0,
					))
				} else if m.MediaType == "Video" {
					// TODO
				}
			}

			// size
			if rawSku.SizeID != "" {
				size := p.Filters.Size.ByID[rawSku.SizeID]
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecSize,
					Id:    size.ID,
					Name:  size.DisplayValue,
					Value: size.Value,
				})
			}

			item.SkuItems = append(item.SkuItems, &sku)
		}

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
		//"https://www.nordstrom.com",
		// "https://www.nordstrom.com/browse/activewear/women-clothing?breadcrumb=Home%2FWomen%2FClothing%2FActivewear&origin=topnav",
		// "https://www.nordstrom.com/s/the-north-face-mountain-water-repellent-hooded-jacket/5500919",
		// "https://www.nordstrom.com/s/anastasia-beverly-hills-liquid-liner/5369732",
		//"https://www.nordstrom.com/s/chanel-le-crayon-khol-intense-eye-pencil/2826730",
		"https://www.nordstrom.com/s/nike-court-borough-low-2-sneaker-baby-walker-toddler-little-kid-big-kid/5756069?origin=category-personalizedsort&breadcrumb=Home%2FKids%2FAll%20Boys%2FTween%20Boys&color=100",
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
