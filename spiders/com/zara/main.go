package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/voiladev/VoilaCrawler/pkg/cli"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	pbMedia "github.com/voiladev/VoilaCrawler/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawler/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawler/protoc-gen-go/chameleon/smelter/v1/crawl/item"
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
		categoryPathMatcher: regexp.MustCompile(`^(/[a-zA-Z0-9\-]+){0,6}/[a-zA-Z0-9\pL\pS%\-]+-l[a-z0-9\pL\-]+\.html$`),
		// this regular used to match product page url path
		productPathMatcher: regexp.MustCompile(`^(/[a-zA-Z0-9\-]+){0,3}/[a-zA-Z0-9\pL\pS%\-]+-p[a-z0-9\pL\-]+\.html$`),
		logger:             logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "646b5037ae64494f8550d84cf323660f"
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
	options := crawler.CrawlOptions{
		EnableHeadless: false,
		// use js api to init session for the first request of the crawl
		EnableSessionInit: false,
	}
	options.MustCookies = append(options.MustCookies,
		&http.Cookie{Name: "storepath", Value: "us%2Fen", Path: "/"},
		&http.Cookie{Name: "web_version", Value: "STANDARD", Path: "/"},
	)
	return &options
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.zara.com"}
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
		u.Host = "www.zara.com"
	}
	if c.productPathMatcher.MatchString(u.Path) {
		// NOTE: this site may got some bug, canonical url not work
		// so here keep the raw url
		// u.RawQuery = ""
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

	if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	} else if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yield)
	}
	return crawler.ErrUnsupportedPath
}

// nextIndex used to get the index from the shared data.
// item.index is a const key for item index.
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")))
}

// below are the golang json data struct of raw website.
// if you get the raw json data of the website,
// then you can use https://mholt.github.io/json-to-go/ to convert it to a golang struct

type CategoryStructure struct {
	ProductGroups []struct {
		Type     string `json:"type"`
		Elements []struct {
			ID                   string `json:"id"`
			Layout               string `json:"layout,omitempty"`
			CommercialComponents []struct {
				ID        int    `json:"id"`
				Reference string `json:"reference"`
				Type      string `json:"type"`
				Kind      string `json:"kind"`
				Brand     struct {
					BrandID        int    `json:"brandId"`
					BrandGroupID   int    `json:"brandGroupId"`
					BrandGroupCode string `json:"brandGroupCode"`
				} `json:"brand"`
				Xmedia []struct {
					Datatype       string   `json:"datatype"`
					Set            int      `json:"set"`
					Type           string   `json:"type"`
					Kind           string   `json:"kind"`
					Path           string   `json:"path"`
					Name           string   `json:"name"`
					Width          int      `json:"width"`
					Height         int      `json:"height"`
					Timestamp      string   `json:"timestamp"`
					AllowedScreens []string `json:"allowedScreens"`
					ExtraInfo      struct {
						Style struct {
							Top      int  `json:"top"`
							Left     int  `json:"left"`
							Width    int  `json:"width"`
							Margined bool `json:"margined"`
						} `json:"style"`
					} `json:"extraInfo"`
				} `json:"xmedia"`
				Name          string  `json:"name"`
				Description   string  `json:"description"`
				Price         float64 `json:"price"`
				Section       int     `json:"section"`
				SectionName   string  `json:"sectionName"`
				FamilyName    string  `json:"familyName"`
				SubfamilyName string  `json:"subfamilyName"`
				Seo           struct {
					Keyword          string `json:"keyword"`
					SeoProductID     string `json:"seoProductId"`
					DiscernProductID int    `json:"discernProductId"`
				} `json:"seo"`
				Availability string        `json:"availability"`
				TagTypes     []interface{} `json:"tagTypes"`
				ExtraInfo    struct {
					IsDivider      bool `json:"isDivider"`
					HighlightPrice bool `json:"highlightPrice"`
				} `json:"extraInfo"`
				Detail struct {
					Reference        string `json:"reference"`
					DisplayReference string `json:"displayReference"`
					Colors           []struct {
						ID        string `json:"id"`
						ProductID int    `json:"productId"`
						Name      string `json:"name"`
						StylingID string `json:"stylingId"`
						Xmedia    []struct {
							Datatype       string   `json:"datatype"`
							Set            int      `json:"set"`
							Type           string   `json:"type"`
							Kind           string   `json:"kind"`
							Path           string   `json:"path"`
							Name           string   `json:"name"`
							Width          int      `json:"width"`
							Height         int      `json:"height"`
							Timestamp      string   `json:"timestamp"`
							AllowedScreens []string `json:"allowedScreens"`
							ExtraInfo      struct {
								Style struct {
									Top      int  `json:"top"`
									Left     int  `json:"left"`
									Width    int  `json:"width"`
									Margined bool `json:"margined"`
								} `json:"style"`
							} `json:"extraInfo"`
						} `json:"xmedia"`
						Price        float64 `json:"price"`
						Availability string  `json:"availability"`
						Reference    string  `json:"reference"`
					} `json:"colors"`
				} `json:"detail"`
				ServerPage               int      `json:"serverPage"`
				GridPosition             int      `json:"gridPosition"`
				ZoomedGridPosition       int      `json:"zoomedGridPosition"`
				HasMoreColors            bool     `json:"hasMoreColors"`
				ProductTag               []string `json:"productTag"`
				ProductTagDynamicClasses string   `json:"productTagDynamicClasses"`
				ColorList                string   `json:"colorList"`
				IsDivider                bool     `json:"isDivider"`
				HasXmediaDouble          bool     `json:"hasXmediaDouble"`
				SimpleXmedia             []struct {
					Datatype       string   `json:"datatype"`
					Set            int      `json:"set"`
					Type           string   `json:"type"`
					Kind           string   `json:"kind"`
					Path           string   `json:"path"`
					Name           string   `json:"name"`
					Width          int      `json:"width"`
					Height         int      `json:"height"`
					Timestamp      string   `json:"timestamp"`
					AllowedScreens []string `json:"allowedScreens"`
					ExtraInfo      struct {
						Style struct {
							Top      int  `json:"top"`
							Left     int  `json:"left"`
							Width    int  `json:"width"`
							Margined bool `json:"margined"`
						} `json:"style"`
					} `json:"extraInfo"`
				} `json:"simpleXmedia"`
				ShowAvailability bool `json:"showAvailability"`
				PriceUnavailable bool `json:"priceUnavailable"`
			} `json:"commercialComponents,omitempty"`
			HasStickyBanner bool   `json:"hasStickyBanner,omitempty"`
			NeedsSeparator  bool   `json:"needsSeparator,omitempty"`
			Header          string `json:"header,omitempty"`
			Description     string `json:"description,omitempty"`
		} `json:"elements"`
		HasStickyBanner bool `json:"hasStickyBanner"`
	} `json:"productGroups"`
	ProductsCount    int `json:"productsCount"`
	ProductsPage     int `json:"productsPage"`
	ProductsPageSize int `json:"productsPageSize"`
}

// used to extract embaded json data in website page.
// more about golang regulation see here https://golang.org/pkg/regexp/syntax/
var productsExtractReg = regexp.MustCompile(`(?U)window\.zara\.viewPayload\s*=\s*({.*});</script>`)

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

	var viewData CategoryStructure
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		return err
	}

	lastIndex := nextIndex(ctx)
	for _, pg := range viewData.ProductGroups {
		if pg.Type != "main" {
			continue
		}

		for _, idv := range pg.Elements {
			if idv.ID == "seo-info" {
				continue
			}
			for _, pc := range idv.CommercialComponents {
				if pc.Type != "Product" {
					continue
				}

				rawurl := fmt.Sprintf("%s://%s/us/en/%s-p%s.html?v1=%v", resp.Request.URL.Scheme, resp.Request.URL.Host, pc.Seo.Keyword, pc.Seo.SeoProductID, pc.Seo.DiscernProductID)
				req, err := http.NewRequest(http.MethodGet, rawurl, nil)
				if err != nil {
					c.logger.Errorf("load http request of url %s failed, error=%s", rawurl, err)
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
		}
	}

	pageSize := float64(viewData.ProductsPageSize)
	productsCount := float64(viewData.ProductsCount)
	totalPages := int64(math.Ceil(productsCount / pageSize))
	// get current page number
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}

	// check if this is the last page
	if page >= totalPages {
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

// Product json
type ProductStructure struct {
	Product struct {
		ID    int `json:"id"`
		Brand struct {
			BrandID        int    `json:"brandId"`
			BrandGroupID   int    `json:"brandGroupId"`
			BrandGroupCode string `json:"brandGroupCode"`
		} `json:"brand"`
		Name                     string `json:"name"`
		Description              string `json:"description"`
		Price                    int64  `json:"price"`
		OldPrice                 int64  `json:"oldPrice"`
		DisplayDiscountPercentag int64  `json:"displayDiscountPercentag"`
		Detail                   struct {
			Description      string        `json:"description"`
			RawDescription   string        `json:"rawDescription"`
			Reference        string        `json:"reference"`
			DisplayReference string        `json:"displayReference"`
			Composition      []interface{} `json:"composition"`
			Care             []interface{} `json:"care"`
			Colors           []struct {
				ID               string        `json:"id"`
				HexCode          string        `json:"hexCode"`
				ProductID        int           `json:"productId"`
				Name             string        `json:"name"`
				Reference        string        `json:"reference"`
				StylingID        string        `json:"stylingId"`
				DetailImages     []interface{} `json:"detailImages"`
				DetailFlatImages []interface{} `json:"detailFlatImages"`
				SizeGuideImages  []interface{} `json:"sizeGuideImages"`
				Videos           []interface{} `json:"videos"`
				Xmedia           []struct {
					Datatype       string   `json:"datatype"`
					Set            int      `json:"set"`
					Type           string   `json:"type"`
					Kind           string   `json:"kind"`
					Path           string   `json:"path"`
					Name           string   `json:"name"`
					Width          int      `json:"width"`
					Height         int      `json:"height"`
					Timestamp      string   `json:"timestamp"`
					AllowedScreens []string `json:"allowedScreens"`
					Gravity        string   `json:"gravity"`
					Order          int      `json:"order,omitempty"`
				} `json:"xmedia"`
				Price    float64 `json:"price"`
				OldPrice float64 `json:"oldPrice"`
				Sizes    []struct {
					Availability     string  `json:"availability"`
					EquivalentSizeID int     `json:"equivalentSizeId"`
					ID               int     `json:"id"`
					Name             string  `json:"name"`
					Price            float64 `json:"price"`
					OldPrice         float64 `json:"oldPrice"`
					Reference        string  `json:"reference"`
					Sku              int     `json:"sku"`
				} `json:"sizes"`
				Description    string `json:"description"`
				RawDescription string `json:"rawDescription"`
				ExtraInfo      struct {
					Preorder struct {
						Message    string `json:"message"`
						IsPreorder bool   `json:"isPreorder"`
					} `json:"preorder"`
					IsStockInStoresAvailable bool `json:"isStockInStoresAvailable"`
				} `json:"extraInfo"`
				DetailedComposition struct {
					Parts      []interface{} `json:"parts"`
					Exceptions []interface{} `json:"exceptions"`
				} `json:"detailedComposition"`
				ColorCutImg struct {
					Datatype       string   `json:"datatype"`
					Set            int      `json:"set"`
					Type           string   `json:"type"`
					Kind           string   `json:"kind"`
					Path           string   `json:"path"`
					Name           string   `json:"name"`
					Width          int      `json:"width"`
					Height         int      `json:"height"`
					Timestamp      string   `json:"timestamp"`
					AllowedScreens []string `json:"allowedScreens"`
					Gravity        string   `json:"gravity"`
				} `json:"colorCutImg"`
				MainImgs []struct {
					Datatype       string   `json:"datatype"`
					Set            int      `json:"set"`
					Type           string   `json:"type"`
					Kind           string   `json:"kind"`
					Path           string   `json:"path"`
					Name           string   `json:"name"`
					Width          int      `json:"width"`
					Height         int      `json:"height"`
					Timestamp      string   `json:"timestamp"`
					AllowedScreens []string `json:"allowedScreens"`
					Gravity        string   `json:"gravity"`
					Order          int      `json:"order"`
				} `json:"mainImgs"`
			} `json:"colors"`
			DetailedComposition struct {
				Parts      []interface{} `json:"parts"`
				Exceptions []interface{} `json:"exceptions"`
			} `json:"detailedComposition"`
			Categories []interface{} `json:"categories"`
			IsBuyable  bool          `json:"isBuyable"`
		} `json:"detail"`
	} `json:"product"`
	Category struct {
		Id          int64  `json:"id"`
		SectionName string `json:"sectionName"`
		Seo         struct {
			Keyword string `json:"keyword"`
		} `json:"seo"`
		BreadCrumb []struct {
			Id      int64  `json:"id"`
			Text    string `json:"text"`
			Keyword string `json:"keyword"`
		} `json:"breadCrumb"`
	} `json:"category"`
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

	var viewData ProductStructure
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}
	canUrl := dom.Find(`link[rel="canonical"]`).AttrOr("href", "")
	//Prepare Product Data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           strconv.Format(viewData.Product.ID),
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
		},
		BrandName:   viewData.Product.Brand.BrandGroupCode,
		CrowdType:   viewData.Category.SectionName,
		Title:       viewData.Product.Name,
		Description: viewData.Product.Description,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
			Current:  int32(viewData.Product.Price),
			Msrp:     int32(viewData.Product.OldPrice),
			Discount: int32(viewData.Product.DisplayDiscountPercentag),
		},
	}
	if item.Price.Msrp == 0 {
		item.Price.Msrp = item.Price.Current
	}

	for i, cate := range viewData.Category.BreadCrumb {
		switch i {
		case 0:
			item.Category = cate.Text
		case 1:
			item.SubCategory = cate.Text
		case 2:
			item.SubCategory2 = cate.Text
		}
	}

	for _, rawColor := range viewData.Product.Detail.Colors {
		colorSpec := pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecColor,
			Id:    strconv.Format(rawColor.ID),
			Name:  rawColor.Name,
			Value: rawColor.HexCode,
		}

		var medias []*pbMedia.Media
		for _, mid := range rawColor.MainImgs {
			template := "https://static.zara.net/photos" + mid.Path + "/{width}" + mid.Name + ".jpg?ts=" + mid.Timestamp
			if mid.Type == "image" {
				medias = append(medias, pbMedia.NewImageMedia(
					strconv.Format(mid.Name),
					strings.ReplaceAll(template, "{width}", ""),
					strings.ReplaceAll(template, "{width}", "w/1280/"),
					strings.ReplaceAll(template, "{width}", "w/600/"),
					strings.ReplaceAll(template, "{width}", "w/500/"),
					"",
					mid.Order == 1,
				))
			}
		}

		for _, rawSku := range rawColor.Sizes {
			originalPrice, _ := strconv.ParseFloat(rawSku.Price)
			msrp, _ := strconv.ParseFloat(rawSku.OldPrice)
			if msrp == 0 {
				msrp = originalPrice
			}
			discount := 0.0
			if msrp != originalPrice {
				discount = math.Ceil((msrp - originalPrice) / msrp * 100)
			}
			sku := pbItem.Sku{
				SourceId:    strconv.Format(rawSku.Sku),
				Medias:      medias,
				Description: rawColor.Description,
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(originalPrice),
					Msrp:     int32(msrp),
					Discount: int32(discount),
				},
				Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}
			if viewData.Product.Detail.IsBuyable {
				sku.Stock.StockStatus = pbItem.Stock_InStock
			}

			sku.Specs = append(sku.Specs, &colorSpec, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    strconv.Format(rawSku.ID),
				Name:  rawSku.Name,
				Value: rawSku.Name,
			})
			item.SkuItems = append(item.SkuItems, &sku)
		}
	}
	if len(item.SkuItems) > 0 {
		// yield item result
		if err = yield(ctx, &item); err != nil {
			return err
		}
	} else {
		return errors.New("no product sku spec found")
	}

	return nil
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		"https://www.zara.com/us/en/woman-bags-l1024.html?v1=1719123",
		// "https://www.zara.com/us/en/fabric-bucket-bag-with-pocket-p16619710.html?v1=100626185&v2=1719123",
		// "https://www.zara.com/us/en/quilted-velvet-maxi-crossbody-bag-p16311710.html?v1=95124768&v2=1719102",
		// "https://www.zara.com/us/en/text-detail-belt-bag-p16363710.html?v1=79728740&v2=1719123",
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
