package main

// This site https://www.ruelala.com/ host brands,categories,items.
// there only exists one sku spec for all items.
// You should signin before signin or mock the access cookie as below.
//
// NOTE: the mock cookie may not stable for all the time.

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	// "github.com/PuerkitoBio/goquery"
	"github.com/voiladev/VoilaCrawler/pkg/cli"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	"github.com/voiladev/VoilaCrawler/pkg/util"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	pbProxy "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
	"google.golang.org/protobuf/types/known/anypb"
)

type _Crawler struct {
	httpClient http.Client

	categoryPathMatcher     *regexp.Regexp
	categoryJsonPathMatcher *regexp.Regexp
	productPathMatcher      *regexp.Regexp
	productJsonPathMatcher  *regexp.Regexp
	logger                  glog.Log
}

func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient:              client,
		categoryPathMatcher:     regexp.MustCompile(`^/nav(/[^/]+){2,5}$`),
		productPathMatcher:      regexp.MustCompile(`^/boutique/product/[0-9]+/[0-9]+/?$`),
		categoryJsonPathMatcher: regexp.MustCompile(`^/api/v[0-9](.[0-9]+)?/catalog/products/?$`),
		productJsonPathMatcher:  regexp.MustCompile(`^/api/v[0-9]/products/[0-9]+/?$`),
		logger:                  logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "40908fc41a3a2092282b64734c76c285"
}

// Version
func (c *_Crawler) Version() int32 {
	return 1
}

// CrawlOptions
func (c *_Crawler) CrawlOptions(u *url.URL) *crawler.CrawlOptions {
	options := crawler.NewCrawlOptions()
	options.EnableHeadless = false
	options.EnableSessionInit = true
	options.Reliability = pbProxy.ProxyReliability_ReliabilityMedium

	// options.MustHeader.Set("X-Requested-With", "XMLHttpRequest")
	options.MustCookies = append(options.MustCookies,
		&http.Cookie{
			Name:  "geolocation_data",
			Value: `{"continent":"NA","timezone":"PST","country":"US","city":"CA","lat":"37.5741","long":"-122.3193"}`,
			Path:  "/",
		},
		&http.Cookie{Name: "bfx.country", Value: "US", Path: "/"},
		&http.Cookie{Name: "bfx.currency", Value: "USD", Path: "/"},
		// &http.Cookie{
		// 	Name:  "bfx.apiKey",
		// 	Value: "c9f2ab70-8028-11e6-bf37-d180220906db",
		// 	Path:  "/", /* TODO: check is this value changeable */
		// },
		&http.Cookie{Name: "bfx.env", Value: "PROD", Path: "/"},
		&http.Cookie{Name: "bfx.logLevel", Value: "ERROR", Path: "/"},
		&http.Cookie{Name: "bfx.language", Value: "en", Path: "/"},
		// &http.Cookie{Name: "bfx.sessionId", Value: "0076DDB9-BFE7-4882-BC40-F80853BA3B77", Path: "/"},
	)
	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.ruelala.com"}
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
		u.Host = "www.ruelala.com"
	}
	if c.productPathMatcher.MatchString(u.Path) {
		u.RawQuery = ""
		return u.String(), nil
	}
	return rawurl, nil
}

func (c *_Crawler) Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	} else if c.categoryJsonPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProductsJson(ctx, resp, yield)
	} else if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		// omit
	} else if c.productJsonPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProductJson(ctx, resp, yield)
	}
	return crawler.ErrUnsupportedPath
}

const defaultCategoryProductsPageSize = "54"

// parseCategoryProducts parse api url from web page url
func (c *_Crawler) parseCategoryProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	var division, department, class string
	fields := strings.FieldsFunc(resp.Request.URL.Path, func(r rune) bool {
		return r == '/'
	})
	for i, field := range fields {
		fields[i], _ = url.QueryUnescape(field)
		switch i {
		case 1:
			division = fields[i]
		case 2:
			department = fields[i]
		case 3:
			class = fields[i]
		}
	}

	var page = "0"
	if p := resp.Request.URL.Query().Get("page"); p != "" {
		page = p
	}

	u := resp.Request.URL
	u.Fragment = ""
	u.RawFragment = ""
	u.Path = "/api/v3.5/catalog/products"
	vals := url.Values{}
	vals.Set("page", page)
	vals.Set("pageSize", defaultCategoryProductsPageSize)
	vals.Set("division", division)
	if department != "" {
		vals.Set("department", department)
	}
	if class != "" {
		vals.Set("class", class)
	}
	vals.Set("bindingFilters", "division|department|class|boutiqueContextId")
	u.RawQuery = vals.Encode()

	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	return yield(ctx, req)
}

type parseCategoryProductsJsonResponse struct {
	Data []struct {
		ID                int     `json:"id"`
		Name              string  `json:"name"`
		BusinessID        string  `json:"businessId"`
		Brand             string  `json:"brand"`
		BackOrderEnabled  bool    `json:"backOrderEnabled"`
		HasMultipleColors bool    `json:"hasMultipleColors"`
		ShowMsrp          bool    `json:"showMsrp"`
		MsrpMin           int     `json:"msrpMin"`
		MsrpMax           int     `json:"msrpMax"`
		ListPriceMin      float64 `json:"listPriceMin"`
		ListPriceMax      float64 `json:"listPriceMax"`
		ImageUrls         struct {
			PRODUCTLIST                 string `json:"PRODUCT_LIST"`
			PRODUCTLISTALT              string `json:"PRODUCT_LIST_ALT"`
			PRODUCTLISTMOBILE           string `json:"PRODUCT_LIST_MOBILE"`
			PRODUCTLISTMOBILEHIGHRES    string `json:"PRODUCT_LIST_MOBILE_HIGH_RES"`
			PRODUCTLISTMOBILEHIGHRESALT string `json:"PRODUCT_LIST_MOBILE_HIGH_RES_ALT"`
		} `json:"imageUrls,omitempty"`
		PercentOff                     int    `json:"percentOff"`
		ShowPercentOff                 bool   `json:"showPercentOff"`
		ProductPage                    string `json:"productPage"`
		Inventory                      int    `json:"inventory"`
		ShowLowInventoryWarning        bool   `json:"showLowInventoryWarning"`
		AvailableForInternationalUsers bool   `json:"availableForInternationalUsers"`
		ShortDescription               string `json:"shortDescription"`
		Skus                           []struct {
			SkuContextID   string      `json:"skuContextId"`
			SkuNumber      string      `json:"skuNumber"`
			ColorDisplay   interface{} `json:"colorDisplay"`
			ColorSortOrder interface{} `json:"colorSortOrder"`
			FilterColor    interface{} `json:"filterColor"`
			Color          interface{} `json:"color"`
			Inventory      int         `json:"inventory"`
			Size           []string    `json:"size"`
			SizeDisplay    string      `json:"sizeDisplay"`
			SizeSortOrder  int         `json:"sizeSortOrder"`
		} `json:"skus"`
		GfhInfo struct {
			ShowInfo         bool        `json:"showInfo"`
			Message          interface{} `json:"message"`
			MessageTreatment string      `json:"messageTreatment"`
		} `json:"gfhInfo"`
	} `json:"data"`
	Meta struct {
		TotalObjects int    `json:"totalObjects"`
		TotalPages   int    `json:"totalPages"`
		CurrentPage  int    `json:"currentPage"`
		PageSize     int    `json:"pageSize"`
		Next         string `json:"next"`
	} `json:"meta"`
	Messages []interface{} `json:"messages"`
}

// nextIndex used to get sharingData from context
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")) + 1)
}

// parseCategoryProductsJson parse api url from web page url
// because of the list not return the detail price,
// here fetch all the detail url and use parseProductJson to get product detail
func (c *_Crawler) parseCategoryProductsJson(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil {
		return nil
	}

	var i parseCategoryProductsJsonResponse
	if err := json.NewDecoder(resp.Body).Decode(&i); err != nil {
		return err
	}
	if len(i.Data) == 0 {
		return fmt.Errorf("no response from %s", resp.Request.URL)
	}

	var (
		lastIndex = nextIndex(ctx)
		vals      = url.Values{}
	)
	for index, item := range i.Data {
		u := *resp.Request.URL
		u.Path = path.Join("/api/v3/products", strconv.Format(item.ID))
		vals.Set("pos", strconv.Format(index+1))
		u.RawQuery = ""

		req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
		req.Header.Set("Referer", fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, item.ProductPage))
		req.Header.Set("X-Requested-With", "XMLHttpRequest")

		lastIndex += 1
		nctx := context.WithValue(ctx, "item.index", lastIndex)
		yield(nctx, req)
	}

	if i.Meta.Next != "" {
		req, err := http.NewRequest(http.MethodGet, i.Meta.Next, nil)
		if err != nil {
			return err
		}
		nctx := context.WithValue(ctx, "item.index", lastIndex)
		return yield(nctx, req)
	}
	return nil
}

// Generate data struct from json https://mholt.github.io/json-to-go/
type parseProductJsonResponse struct {
	Data struct {
		ID                 int32         `json:"id"`
		Brand              string        `json:"brand"`
		Name               string        `json:"name"`
		BusinessID         string        `json:"businessId"`
		BackOrderEnabled   bool          `json:"backOrderEnabled"`
		ShortDescription   string        `json:"shortDescription"`
		IsFinalSale        bool          `json:"isFinalSale"`
		BoutiqueID         string        `json:"boutiqueId"`
		SizeChart          string        `json:"sizeChart"`
		MaxPerCart         int32         `json:"maxPerCart"`
		Type               string        `json:"type"`
		ShowMsrp           bool          `json:"showMsrp"`
		Division           []string      `json:"division"`
		Department         string        `json:"department"`
		Class              string        `json:"class"`
		SubClass           string        `json:"subClass"`
		BoutiqueBusinessID string        `json:"boutiqueBusinessId"`
		ReturnMessage      string        `json:"returnMessage"`
		Locations          []interface{} `json:"locations"`
		Afterpay           struct {
			Available bool   `json:"available"`
			Message   string `json:"message"`
			HelpURL   string `json:"helpURL"`
		} `json:"afterpay"`
		ParentPage              string  `json:"parentPage"`
		TotalInventory          int32   `json:"totalInventory"`
		ShowLowInventoryWarning bool    `json:"showLowInventoryWarning"`
		ShipsMessage            string  `json:"shipsMessage"`
		MsrpMin                 float64 `json:"msrpMin"`
		MsrpMax                 float64 `json:"msrpMax"`
		ListPriceMin            float64 `json:"listPriceMin"`
		ListPriceMax            float64 `json:"listPriceMax"`
		PercentOff              float64 `json:"percentOff"`
		ShowPercentOff          bool    `json:"showPercentOff"`
		Skus                    []struct {
			ID       string `json:"id"`
			Afterpay struct {
				Available bool   `json:"available"`
				Message   string `json:"message"`
				HelpURL   string `json:"helpURL"`
			} `json:"afterpay"`
			SkuContextID            string  `json:"skuContextId"`
			SkuNumber               string  `json:"skuNumber"`
			Size                    string  `json:"size"`
			Color                   string  `json:"color"`
			Price                   float64 `json:"price"`
			Msrp                    float64 `json:"msrp"`
			PercentOff              float64 `json:"percentOff"`
			ShowPercentOff          bool    `json:"showPercentOff"`
			ShippingUpcharge        float64 `json:"shippingUpcharge"`
			Inventory               int     `json:"inventory"`
			ShowLowInventoryWarning bool    `json:"showLowInventoryWarning"`
			Features                string  `json:"features"`
			Highlights              string  `json:"highlights"`
			Terms                   string  `json:"terms"`
		} `json:"skus"`
		Attributes struct {
			Colors []struct {
				InternalValue     string   `json:"internal_value"`
				DisplayValue      string   `json:"display_value"`
				Swatch            string   `json:"swatch"`
				ImagesDetail      []string `json:"images_detail"`
				ImagesZoom        []string `json:"images_zoom"`
				ImagesAlt         []string `json:"images_alt"`
				ImagesTablet      []string `json:"images_tablet"`
				ImagesTabletHires []string `json:"images_tablet_hires"`
			} `json:"colors"`
			Sizes []struct {
				InternalValue string `json:"internal_value"`
				DisplayValue  string `json:"display_value"`
			} `json:"sizes"`
		} `json:"attributes"`
		AvailableForInternationalUsers bool   `json:"availableForInternationalUsers"`
		ShipsInternationalMessage      string `json:"shipsInternationalMessage"`
		ReturnsInternationalMessage    string `json:"returnsInternationalMessage"`
		ShippingProgram                struct {
			Type    string `json:"type"`
			MinDays int32  `json:"minDays"`
			MaxDays int32  `json:"maxDays"`
		} `json:"shippingProgram"`
		InternationalShippingProgram struct {
			Type    string `json:"type"`
			MinDays int32  `json:"minDays"`
			MaxDays int32  `json:"maxDays"`
		} `json:"internationalShippingProgram"`
		GfhInfo struct {
			ShowInfo         bool   `json:"showInfo"`
			Message          string `json:"message"`
			MessageTreatment string `json:"messageTreatment"`
		} `json:"gfhInfo"`
	} `json:"data"`
	Meta struct {
	} `json:"meta"`
	Messages []interface{} `json:"messages"`
}

func (c *_Crawler) parseProductJson(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var i parseProductJsonResponse
	if err := json.Unmarshal(respbody, &i); err != nil {
		c.logger.Debugf("%s", respbody)
		return err
	}

	canUrl := fmt.Sprintf("https://www.ruelala.com/boutique/product/%v/%v/", i.Data.BoutiqueID, i.Data.ID)
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           strconv.Format(i.Data.ID),
			CrawlUrl:     canUrl,
			CanonicalUrl: canUrl,
		},
		Title:        i.Data.Name,
		Description:  i.Data.ShortDescription,
		BrandName:    i.Data.Brand,
		Category:     i.Data.Department,
		SubCategory:  i.Data.Class,
		SubCategory2: i.Data.SubClass,
		Stock: &pbItem.Stock{
			StockCount: i.Data.TotalInventory,
		},
	}

	skuSpec := map[string]*pbItem.SkuSpecOption{}
	medias := map[string][]*media.Media{}
	for _, color := range i.Data.Attributes.Colors {
		for j := 0; ; j++ {
			if len(color.ImagesZoom) <= j {
				break
			}

			img := media.Media_Image{}
			if len(color.ImagesZoom) > j { // 864x1080
				img.OriginalUrl = util.UrlCompletion(color.ImagesZoom[j])
			}
			if len(color.ImagesDetail) > j { // 400x500
				img.SmallUrl = util.UrlCompletion(color.ImagesDetail[j])
			}
			if len(color.ImagesTablet) > j { //528x660
				img.MediumUrl = util.UrlCompletion(color.ImagesTablet[j])
			}
			if len(color.ImagesTabletHires) > j && img.OriginalUrl == "" { // 1056x1320
				img.LargeUrl = util.UrlCompletion(color.ImagesTabletHires[j])
			}

			imgData, _ := anypb.New(&img)
			m := media.Media{Detail: imgData, IsDefault: j == 0}

			key := fmt.Sprintf("%s-%v", pbItem.SkuSpecType_SkuSpecColor, strings.ToLower(color.DisplayValue))
			medias[key] = append(medias[key], &m)
		}

		if color.DisplayValue != "" && color.InternalValue != "" {
			spec := pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Name:  color.DisplayValue,
				Value: color.DisplayValue,
			}
			skuSpec[fmt.Sprintf("%s-%v", pbItem.SkuSpecType_SkuSpecColor, strings.ToLower(spec.Name))] = &spec
		}
	}
	for _, size := range i.Data.Attributes.Sizes {
		spec := pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Name:  size.DisplayValue,
			Value: size.DisplayValue,
		}
		skuSpec[fmt.Sprintf("%s-%v", pbItem.SkuSpecType_SkuSpecSize, strings.ToLower(spec.Name))] = &spec
	}

	for _, u := range i.Data.Skus {
		sku := pbItem.Sku{
			SourceId:    strconv.Format(u.ID),
			Title:       i.Data.Name,
			Description: u.Features,
			Price: &pbItem.Price{
				// 接口里返回的都是美元价格，页面上的结算价格是根据当前的IP来判断的
				Currency:  regulation.Currency_USD,
				Current:   int32(u.Price),
				Msrp:      int32(u.Msrp),
				Discount:  int32(u.PercentOff),
				Discount1: 0,
			},
			Stock: &pbItem.Stock{
				StockStatus: pbItem.Stock_OutOfStock,
				StockCount:  int32(u.Inventory),
			},
		}
		if u.Color != "" {
			if spec, ok := skuSpec[fmt.Sprintf("%s-%v", pbItem.SkuSpecType_SkuSpecColor, strings.ToLower(u.Color))]; ok {
				sku.Specs = append(sku.Specs, spec)
			}
			if ms, ok := medias[fmt.Sprintf("%s-%v", pbItem.SkuSpecType_SkuSpecColor, strings.ToLower(u.Color))]; ok {
				sku.Medias = ms
			}
		}
		if u.Size != "" {
			if spec, ok := skuSpec[fmt.Sprintf("%s-%v", pbItem.SkuSpecType_SkuSpecSize, strings.ToLower(u.Size))]; ok {
				sku.Specs = append(sku.Specs, spec)
			}
		}
		if u.Inventory > 0 {
			// 这里不再更细化的区分是有很多库存，还是有几个库存
			sku.Stock.StockStatus = pbItem.Stock_InStock
		}
		item.SkuItems = append(item.SkuItems, &sku)
	}
	if len(medias) == 1 {
		for _, m := range medias {
			item.Medias = m
		}
	}
	return yield(ctx, &item)
}

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		"https://www.ruelala.com/nav/women/clothing/dresses%20&%20skirts?dsi=CAT-1267617049--bdfc116b-6bba-4b4b-8a91-25c170e607ef&lsi=d8bf02ed-e287-4873-aab9-7aeb8f43ccd3",
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

// main func is the entry of golang program. this will not be used by plugin, just for local spider test.
func main() {
	cli.NewApp(New).Run(os.Args)
}
