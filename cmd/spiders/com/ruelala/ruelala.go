package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	// "github.com/PuerkitoBio/goquery"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	urlutil "github.com/voiladev/VoilaCrawler/pkg/net/url"
	"github.com/voiladev/VoilaCrawler/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawler/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawler/protoc-gen-go/chameleon/smelter/v1/crawler/item"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
	"google.golang.org/protobuf/types/known/anypb"
)

var (
	_ crawler.Crawler       = (*_Crawler)(nil)
	_ crawler.HealthChecker = (*_Crawler)(nil)
)

type _Crawler struct {
	categoryPathMatcher    *regexp.Regexp
	productPathMatcher     *regexp.Regexp
	productJsonPathMatcher *regexp.Regexp
	logger                 glog.Log
}

func New(logger glog.Log) (*_Crawler, error) {
	c := _Crawler{
		categoryPathMatcher:    regexp.MustCompile(`^/nav/([^/]+){2,5}$`),
		productPathMatcher:     regexp.MustCompile(`^/boutique/product/[0-9]+/[0-9]+/?$`),
		productJsonPathMatcher: regexp.MustCompile(`^/api/v[0-9]/products/[0-9]+/?$`),
		logger:                 logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "70cf93e2c360816ba186c294fecbba06"
}

// Version
func (c *_Crawler) Version() int32 {
	return 1
}

// CrawlOptions
func (c *_Crawler) CrawlOptions() *crawler.CrawlOptions {
	options := crawler.NewCrawlOptions()
	options.DisableProxy = false
	options.EnableHeadless = false
	options.LoginRequired = false
	// NOTE: not need to set useragent here for user agent is dynamic
	options.MustHeader.Set("Accept-Language", "en-US,en;q=0.8")
	options.MustHeader.Set("X-Requested-With", "XMLHttpRequest")
	options.MustCookies = append(options.MustCookies,
		&http.Cookie{Name: "geolocation_data", Value: `{"continent":"NA","timezone":"PST","country":"US","city":"CA","lat":"37.5741","long":"-122.3193"}`, Path: "/"},
		&http.Cookie{Name: "bfx.country", Value: "US", Path: "/"},
		&http.Cookie{Name: "bfx.currency", Value: "USD", Path: "/"},
		&http.Cookie{Name: "bfx.apiKey", Value: "c9f2ab70-8028-11e6-bf37-d180220906db", Path: "/" /* TODO: check is this value changeable */},
		&http.Cookie{Name: "bfx.env", Value: "PROD", Path: "/"},
		&http.Cookie{Name: "bfx.logLevel", Value: "ERROR", Path: "/"},
		&http.Cookie{Name: "bfx.language", Value: "en", Path: "/"},
		&http.Cookie{Name: "bfx.sessionId", Value: "0076DDB9-BFE7-4882-BC40-F80853BA3B77", Path: "/"},
	)
	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"www.ruelala.com"}
}

func (c *_Crawler) IsUrlMatch(u *url.URL) bool {
	if c == nil || u == nil {
		return false
	}

	for _, reg := range []*regexp.Regexp{
		c.categoryPathMatcher,
		c.productPathMatcher,
		c.productJsonPathMatcher,
	} {
		if reg.MatchString(u.Path) {
			return true
		}
	}
	return false
}

func (c *_Crawler) parseCategoryProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {

	return nil
}

// Generate data struct from json https://mholt.github.io/json-to-go/
type _GetProductResponse struct {
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
		ParentPage              string `json:"parentPage"`
		TotalInventory          int32  `json:"totalInventory"`
		ShowLowInventoryWarning bool   `json:"showLowInventoryWarning"`
		ShipsMessage            string `json:"shipsMessage"`
		MsrpMin                 int32  `json:"msrpMin"`
		MsrpMax                 int32  `json:"msrpMax"`
		ListPriceMin            int32  `json:"listPriceMin"`
		ListPriceMax            int32  `json:"listPriceMax"`
		PercentOff              int32  `json:"percentOff"`
		ShowPercentOff          bool   `json:"showPercentOff"`
		Skus                    []struct {
			ID       string `json:"id"`
			Afterpay struct {
				Available bool   `json:"available"`
				Message   string `json:"message"`
				HelpURL   string `json:"helpURL"`
			} `json:"afterpay"`
			SkuContextID            string `json:"skuContextId"`
			SkuNumber               string `json:"skuNumber"`
			Size                    string `json:"size"`
			Color                   string `json:"color"`
			Price                   int32  `json:"price"`
			Msrp                    int32  `json:"msrp"`
			PercentOff              int32  `json:"percentOff"`
			ShowPercentOff          bool   `json:"showPercentOff"`
			ShippingUpcharge        int32  `json:"shippingUpcharge"`
			Inventory               int32  `json:"inventory"`
			ShowLowInventoryWarning bool   `json:"showLowInventoryWarning"`
			Features                string `json:"features"`
			Highlights              string `json:"highlights"`
			Terms                   string `json:"terms"`
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

	var i _GetProductResponse
	if err := json.NewDecoder(resp.Body).Decode(&i); err != nil {
		return err
	}

	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:       strconv.Format(i.Data.ID),
			CrawlUrl: fmt.Sprintf("https://www.ruelala.com/boutique/product/%v/%v/", i.Data.BoutiqueID, i.Data.ID),
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
	for i, color := range i.Data.Attributes.Colors {
		img := media.Media_Image{}
		if len(color.ImagesAlt) > 0 { // 65x85
			img.ThumbnailSmallUrl = urlutil.Format(color.ImagesAlt[0])
		}
		if len(color.ImagesDetail) > 0 { // 400x500
			img.ThumbnailUrl = urlutil.Format(color.ImagesDetail[0])
		}
		if len(color.ImagesTablet) > 0 { //528x660
			img.ThumbnailLargeUrl = urlutil.Format(color.ImagesTablet[0])
		}
		if len(color.ImagesZoom) > 0 { // 864x1080
			img.OriginalUrl = urlutil.Format(color.ImagesZoom[0])
		}
		if len(color.ImagesTabletHires) > 0 && img.OriginalUrl == "" { // 1056x1320
			img.OriginalUrl = urlutil.Format(color.ImagesTabletHires[0])
		}
		imgData, _ := anypb.New(&img)
		m := media.Media{Detail: imgData, IsDefault: i == 0}
		medias[fmt.Sprintf("%s-%v", pbItem.SkuSpecType_SkuSpecColor, strings.ToLower(color.DisplayValue))] = []*media.Media{&m}

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
				Current:   u.Price,
				Msrp:      u.Msrp,
				Discount:  u.PercentOff,
				Discount1: 0,
			},
			Stock: &pbItem.Stock{
				StockStatus: pbItem.Stock_OutOfStock,
				StockCount:  u.Inventory,
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

func (c *_Crawler) Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	} else if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
	} else if c.productJsonPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProductJson(ctx, resp, yield)
	}
	return errors.New("unsupported")
}

func (c *_Crawler) NewTestRequest(ctx context.Context) []*http.Request {
	// req, _ := http.NewRequest(http.MethodGet, "https://www.ruelala.com/boutique/product/174603/119603536/?dsi=CAT-1267617049--4dee2f9b-246e-4f10-a0bc-dedfbf503be5&lsi=b09cfcbf-cdc7-41d1-81fe-0d48f800ada5&pos=17", nil)
	req, _ := http.NewRequest(http.MethodGet, "https://www.ruelala.com/api/v3/products/119603585", nil)

	return []*http.Request{req}
}

func (c *_Crawler) CheckTestResponse(ctx context.Context, resp *http.Response) error {
	return nil
}

// local test
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	spider, err := New(glog.New(glog.LogLevelDebug))
	if err != nil {
		panic(err)
	}
	opts := spider.CrawlOptions()

	for _, req := range spider.NewTestRequest(ctx) {
		for k := range opts.MustHeader {
			req.Header.Set(k, opts.MustHeader.Get(k))
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.141 Safari/537.36")
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
		req = req.WithContext(ctx)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		if err := spider.Parse(ctx, resp, func(ctx context.Context, val interface{}) error {
			data, err := json.Marshal(val)
			if err != nil {
				return err
			}
			glog.Infof("product: %s", data)

			return nil
		}); err != nil {
			panic(err)
		}
	}
}
