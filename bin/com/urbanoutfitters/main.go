package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/voiladev/VoilaCrawler/pkg/cli"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
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
	logger                  glog.Log
}

func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient:              client,
		categoryPathMatcher:     regexp.MustCompile(`^/[a-z0-9\-]+$`),
		categoryJsonPathMatcher: regexp.MustCompile(`^/api/catalog/v[0-9]+/[a-z0-9\-]+/pools/US_DIRECT/navigation-items/[a-z0-9\-]+/products$`),
		productPathMatcher:      regexp.MustCompile(`^/shop/[a-z0-9\-]+$`),
		logger:                  logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "34b3c9715b013ad08c502aeac92849f3"
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
	options.MustHeader.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_2_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.90 Safari/537.36")
	options.MustCookies = append(options.MustCookies,
		// urbn_auth_payload
		&http.Cookie{Name: "localredirected", Value: "False", Path: "/"},
		&http.Cookie{Name: "siteId", Value: "uo-us", Path: "/"},
		&http.Cookie{Name: "urbn_channel", Value: "web", Path: "/"},
		&http.Cookie{Name: "urbn_country", Value: "US", Path: "/"},
		&http.Cookie{Name: "urbn_currency", Value: "USD", Path: "/"},
		&http.Cookie{Name: "urbn_data_center_id", Value: "US-NV", Path: "/"},
		&http.Cookie{Name: "urbn_device_type", Value: "LARGE", Path: "/"},
		&http.Cookie{Name: "urbn_edgescape_site_id", Value: "uo-us", Path: "/"},
		&http.Cookie{Name: "urbn_geo_region", Value: "US-NV", Path: "/"},
		&http.Cookie{Name: "urbn_inventory_pool", Value: "US_DIRECT", Path: "/"},
		&http.Cookie{Name: "urbn_language", Value: "en-US", Path: "/"},
		&http.Cookie{Name: "urbn_privacy_restriction_region", Value: "CA", Path: "/"},
		&http.Cookie{Name: "urbn_site_id", Value: "uo-us", Path: "/"},
	)
	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.urbanoutfitters.com"}
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
		u.Host = "www.urbanoutfitters.com"
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

	if resp.Request.URL.Path == "/unsupported" {
		return fmt.Errorf("invalud request url")
	}
	p := strings.TrimSuffix(resp.Request.URL.Path, "/")

	if p == "" || p == "/womens-clothing" || p == "/mens-clothing" || p == "/home" || p == "/lifestyle" || p == "/beauty-products" || p == "/sale" {
		return c.parseCategories(ctx, resp, yield)
	}

	if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yield)
	} else if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	} else if c.categoryJsonPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryJsonProducts(ctx, resp, yield)
	}
	return crawler.ErrUnsupportedPath
}

type categoryStructure struct {
	Navigationitems []struct {
		Displayname     string `json:"displayName"`
		Navigationitems []struct {
			Displayname string `json:"displayName"`
			Slug        string `json:"slug"`
		} `json:"navigationItems"`
	} `json:"navigationItems"`
}

var authExtractReg = regexp.MustCompile(`(?U)authToken%22%3A%22(.*)%22`)

func (c *_Crawler) parseCategories(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	fmt.Println(`parseCategories enter`)
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}
	s := []string{}

	p := strings.TrimSuffix(resp.Request.URL.Path, "/")
	if p == "" {
		sel := dom.Find(`.c-pwa-header-navigation__item`)
		for i := range sel.Nodes {
			node := sel.Eq(i)
			href := "https://www.urbanoutfitters.com/api/catalog/v0/uo-us/navigation?slug=" + node.AttrOr("data-nav-slug", "") + "&projection-slug=navdepth2"
			s = append(s, href)
		}
	}

	if p == "/womens-clothing" || p == "/mens-clothing" || p == "/home" || p == "/lifestyle" || p == "/beauty-products" || p == "/sale" {
		href := "https://www.urbanoutfitters.com/api/catalog/v0/uo-us/navigation?slug=" + strings.TrimPrefix(p, "/") + "&projection-slug=navdepth2"
		s = append(s, href)
	}

	cookie := resp.Response.Header

	urbn_auth_payload := authExtractReg.FindSubmatch([]byte(strings.Join(cookie.Values(`Set-Cookie`), ";")))

	for _, catUrl := range s {
		fmt.Println(`catUrl `, catUrl)
		//catUrl := "https://www.urbanoutfitters.com/api/catalog/v0/uo-us/navigation?slug=mens-clothing&projection-slug=navdepth2"

		req, err := http.NewRequest(http.MethodGet, catUrl, nil)
		req.Header.Add("accept", "application/json, text/plain, */*")

		req.Header.Add("authorization", "Bearer "+string(urbn_auth_payload[1]))
		req.Header.Add("referer", "https://www.urbanoutfitters.com/")
		req.Header.Add("accept-language", "en-GB,en-US;q=0.9,en;q=0.8")
		req.Header.Add("x-urbn-channel", "web")
		req.Header.Add("x-urbn-language", "en-US")

		catreq, err := c.httpClient.Do(ctx, req)
		if err != nil {
			panic(err)
		}
		defer catreq.Body.Close()

		catBody, err := ioutil.ReadAll(catreq.Body)
		if err != nil {
			c.logger.Error(err)
			return err
		}

		var viewData categoryStructure
		if err := json.Unmarshal(catBody, &viewData); err != nil {
			c.logger.Errorf("unmarshal cat detail data fialed, error=%s", err)
			return err
		}

		for _, rawCat := range viewData.Navigationitems {

			cateName := rawCat.Displayname
			if cateName == "" {
				continue
			}
			nnctx := context.WithValue(ctx, "Category", cateName)
			fmt.Println(`cateName `, cateName)
			for _, rawsubCat := range rawCat.Navigationitems {

				href := "https://www.urbanoutfitters.com/" + rawsubCat.Slug
				if rawsubCat.Slug == "" {
					continue
				}

				_, err := url.Parse(href)
				if err != nil {
					c.logger.Error("parse url %s failed", href)
					continue
				}

				subCateName := rawsubCat.Displayname
				fmt.Println(subCateName)
				nnnctx := context.WithValue(nnctx, "SubCategory", subCateName)
				req, _ := http.NewRequest(http.MethodGet, href, nil)
				if err := yield(nnnctx, req); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// nextIndex used to get sharingData from context
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")) + 1)
}

var (
	tilesPerPageReg = regexp.MustCompile(`\\"tilesPerPage\\":\s*([0-9]+),`)
	defaultPageSize = 96
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
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}
	sel := doc.Find(`.c-pwa-tile-tiles>.s-pwa-tile-grid>.c-pwa-tile-grid-inner`)

	c.logger.Debugf("found %d", len(sel.Nodes))
	for i := range sel.Nodes {
		node := sel.Eq(i)
		href := node.Find(`.c-pwa-product-tile>a`).AttrOr("href", "")
		if href == "" {
			href = node.Find(`.o-pwa-product-tile>a`).AttrOr("href", "")
		}
		if href != "" {
			nctx := context.WithValue(ctx, "item.index", lastIndex+1)
			lastIndex += 1
			if req, err := http.NewRequest(http.MethodGet, href, nil); err != nil {
				return err
			} else if err = yield(nctx, req); err != nil {
				return err
			}
		}
	}

	matched := tilesPerPageReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		return fmt.Errorf("find tilesPerPage for %s failed, error=%s", resp.Request.URL, err)
	}
	perPageCount := strconv.MustParseInt(string(matched[1]))
	if len(sel.Nodes) < int(perPageCount) {
		return nil
	}

	// category
	fields := strings.Split(resp.Request.URL.Path, "/")
	category := fields[len(fields)-1]

	currentPage, _ := strconv.ParseInt(resp.Request.URL.Query().Get("page"))
	if currentPage == 0 {
		currentPage = 1
	}
	currentPage += 1

	nctx := context.WithValue(ctx, "item.index", lastIndex)
	u := fmt.Sprintf("/api/catalog/v2/uo-us/pools/US_DIRECT/navigation-items/%s/products?page-size=%v&skip=%v&projection-slug=categorytiles", category, perPageCount, currentPage*perPageCount)
	if req, err := http.NewRequest(http.MethodGet, u, nil); err != nil {
		return err
	} else {
		req.Header.Set("x-urbn-channel", "web")
		req.Header.Set("x-urbn-country", "US")
		req.Header.Set("x-urbn-currency", "USD")
		req.Header.Set("x-urbn-experience", "ss")
		req.Header.Set("x-urbn-geo-region", "US-NV")
		req.Header.Set("x-urbn-language", "en-US")
		req.Header.Set("x-urbn-primary-data-center-id", "US-NV")
		req.Header.Set("x-urbn-site-id", "uo-us")
		if err = yield(nctx, req); err != nil {
			return err
		}
	}
	return nil
}

type parseCategoryJsonProductsResponse struct {
	Records []struct {
		AllMeta struct {
			Tile struct {
				Product struct {
					SupplierSku string `json:"supplierSku"`
					DisplayName string `json:"displayName"`
					Facets      struct {
						Colors []struct {
							ColorID      string `json:"colorId"`
							FaceOutImage string `json:"faceOutImage"`
							HoverImage   string `json:"hoverImage"`
						} `json:"colors"`
					} `json:"facets"`
					DefaultImage         string `json:"defaultImage"`
					StyleNumber          string `json:"styleNumber"`
					IsGiftCard           bool   `json:"isGiftCard"`
					ConfigurationEnabled bool   `json:"configurationEnabled"`
					IsEgiftCard          bool   `json:"isEgiftCard"`
					DefaultColorCode     string `json:"defaultColorCode"`
					ProductSlug          string `json:"productSlug"`
					DisplaySoldOut       bool   `json:"displaySoldOut"`
					ProductID            string `json:"productId"`
				} `json:"product"`
				HoverImage   string `json:"hoverImage"`
				FaceOutImage string `json:"faceOutImage"`
				SkuInfo      struct {
					ListPriceHigh   float64 `json:"listPriceHigh"`
					MarkdownState   string  `json:"markdownState"`
					ListPriceLow    float64 `json:"listPriceLow"`
					HasMarkdown     bool    `json:"hasMarkdown"`
					HasAvailableSku bool    `json:"hasAvailableSku"`
					SalePriceLow    float64 `json:"salePriceLow"`
					SalePriceHigh   float64 `json:"salePriceHigh"`
					PrimarySlice    struct {
						SliceItems []struct {
							Code        string   `json:"code"`
							DisplayName string   `json:"displayName"`
							HexColor    string   `json:"hexColor"`
							SwatchURL   string   `json:"swatchUrl"`
							Images      []string `json:"images"`
							ID          string   `json:"id"`
						} `json:"sliceItems"`
					} `json:"primarySlice"`
				} `json:"skuInfo"`
				FaceOutColorCode string `json:"faceOutColorCode"`
				Reviews          struct {
					Count         int     `json:"count"`
					AverageRating float32 `json:"averageRating"`
				} `json:"reviews"`
			} `json:"tile"`
		} `json:"allMeta"`
	} `json:"records"`
	TotalRecordCount int `json:"totalRecordCount"`
}

func (c *_Crawler) parseCategoryJsonProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.logger.Debugf("resp: %s, error=%s", respBody, err)
		return err
	}

	lastIndex := nextIndex(ctx)
	fields := strings.Split(resp.Request.URL.Path, "/")
	category := fields[len(fields)-2]

	var r parseCategoryJsonProductsResponse
	if err := json.Unmarshal(respBody, &r); err != nil {
		c.logger.Debug(err)
		return err
	}
	for _, record := range r.Records {
		prod := record.AllMeta.Tile.Product
		if prod.ProductSlug == "" {
			return fmt.Errorf("get empty slug of product from %s", resp.Request.URL)
		}
		lastIndex += 1

		u := url.URL{Path: fmt.Sprintf(`/shop/%s`, prod.ProductSlug)}
		vals := url.Values{}
		vals.Set("category", category)
		if prod.DefaultColorCode != "" {
			vals.Set("color", prod.DefaultColorCode)
		} else {
			for _, color := range prod.Facets.Colors {
				vals.Set("color", color.ColorID)
				break
			}
		}

		// query params type, quantity, size is auto set by js, ignore them
		if req, err := http.NewRequest(http.MethodGet, u.String(), nil); err != nil {
			return err
		} else if err = yield(context.WithValue(ctx, "item.index", lastIndex), req); err != nil {
			return err
		}
	}

	// pagination
	perPageCount := strconv.MustParseInt(resp.Request.URL.Query().Get("page-size"))
	skipCount := strconv.MustParseInt(resp.Request.URL.Query().Get("skip"))
	if len(r.Records) < int(perPageCount) || skipCount >= int64(r.TotalRecordCount) {
		return nil
	}

	nctx := context.WithValue(ctx, "item.index", lastIndex)
	u := fmt.Sprintf("/api/catalog/v2/uo-us/pools/US_DIRECT/navigation-items/%s/products?page-size=%v&skip=%v&projection-slug=categorytiles", category, perPageCount, skipCount+perPageCount)
	if req, err := http.NewRequest(http.MethodGet, u, nil); err != nil {
		return err
	} else {
		req.Header.Set("x-urbn-channel", "web")
		req.Header.Set("x-urbn-country", "US")
		req.Header.Set("x-urbn-currency", "USD")
		req.Header.Set("x-urbn-experience", "ss")
		req.Header.Set("x-urbn-geo-region", "US-NV")
		req.Header.Set("x-urbn-language", "en-US")
		req.Header.Set("x-urbn-primary-data-center-id", "US-NV")
		req.Header.Set("x-urbn-site-id", "uo-us")
		if err = yield(nctx, req); err != nil {
			return err
		}
	}
	return nil
}

type rawProduct struct {
	CatalogData struct {
		Product struct {
			ParentCategoryID      string `json:"parentCategoryId"`
			StyleNumber           string `json:"styleNumber"`
			AdditionalDescription string `json:"additionalDescription"`
			Vintage               struct {
			} `json:"vintage"`
			MerchandiseClass string `json:"merchandiseClass"`
			DefaultImage     string `json:"defaultImage"`
			ProductID        string `json:"productId"`
			AvailableForISPU bool   `json:"availableForISPU"`
			WebExclusive     bool   `json:"webExclusive"`
			ParentCategory   struct {
				DisplayName string `json:"displayName"`
				ID          string `json:"id"`
			} `json:"parentCategory"`
			Brand                 string        `json:"brand"`
			DefaultSizeType       string        `json:"defaultSizeType"`
			IsGiftCard            bool          `json:"isGiftCard"`
			LongDescription       string        `json:"longDescription"`
			FamilyProducts        []interface{} `json:"familyProducts"`
			DefaultColorCode      string        `json:"defaultColorCode"`
			SupplierSku           string        `json:"supplierSku"`
			RequestSwatch         bool          `json:"requestSwatch"`
			RemoveForLegalReasons bool          `json:"removeForLegalReasons"`
			PreorderFlag          bool          `json:"preorderFlag"`
			DisplayName           string        `json:"displayName"`
			DisplaySoldOut        bool          `json:"displaySoldOut"`
			BrandDescription      string        `json:"brandDescription"`
			SizeGuide             string        `json:"sizeGuide"`
			IsVintage             bool          `json:"isVintage"`
			IsEgiftCard           bool          `json:"isEgiftCard"`
			UrbnExclusive         bool          `json:"urbnExclusive"`
			ProductSlug           string        `json:"productSlug"`
			IsMarketPlace         bool          `json:"isMarketPlace"`
		} `json:"product"`
		LastModified int `json:"lastModified"`
		SkuInfo      struct {
			ListPriceHigh float64 `json:"listPriceHigh"`
			ListPriceLow  float64 `json:"listPriceLow"`
			SalePriceLow  float64 `json:"salePriceLow"`
			SalePriceHigh float64 `json:"salePriceHigh"`
			PrimarySlice  struct {
				DisplayLabel string `json:"displayLabel"`
				SliceItems   []struct {
					Code         string `json:"code"`
					DisplayName  string `json:"displayName"`
					IncludedSkus []struct {
						SkuID                   string      `json:"skuId"`
						ShipRestriction         interface{} `json:"shipRestriction"`
						ColorID                 string      `json:"colorId"`
						CollectionPointEligible bool        `json:"collectionPointEligible"`
						MarkdownState           string      `json:"markdownState"`
						StockLevel              int         `json:"stockLevel"`
						IsDropShip              bool        `json:"isDropShip"`
						AvailableStatus         int         `json:"availableStatus"`
						SizeID                  string      `json:"sizeId"`
						IsFlatRate              bool        `json:"isFlatRate"`
						Backorder               int         `json:"backorder"`
						Size                    string      `json:"size"`
						Afterpay                struct {
							Status        string  `json:"status"`
							NumOfPayments float64 `json:"numOfPayments"`
							Payment       float64 `json:"payment"`
						} `json:"afterpay"`
						AvailabilityDate  int     `json:"availabilityDate"`
						ListPrice         float64 `json:"listPrice"`
						SalePrice         float64 `json:"salePrice"`
						ReturnRestockInfo struct {
						} `json:"returnRestockInfo"`
					} `json:"includedSkus"`
					HexColor               string        `json:"hexColor"`
					ProductsRelatedToColor []interface{} `json:"productsRelatedToColor"`
					StockLevel             int           `json:"stockLevel"`
					SwatchURL              string        `json:"swatchUrl"`
					Images                 []string      `json:"images"`
					ID                     string        `json:"id"`
				} `json:"sliceItems"`
			} `json:"primarySlice"`
		} `json:"skuInfo"`
		Reviews struct {
			Count         int     `json:"count"`
			AverageRating float32 `json:"averageRating"`
		} `json:"reviews"`
		ProductSlug string `json:"productSlug"`
		ProductID   string `json:"productId"`
	} `json:"catalogData"`
	Badges struct {
		Visual  interface{}   `json:"visual"`
		Textual []interface{} `json:"textual"`
	} `json:"badges"`
	Breadcrumbs []struct {
		CategoryID  string `json:"categoryId"`
		DisplayName string `json:"displayName"`
		Slug        string `json:"slug"`
		TypeCode    string `json:"typeCode"`
	} `json:"breadcrumbs"`
	RecommendationQuery  interface{} `json:"recommendationQuery"`
	CategoryQuery        string      `json:"categoryQuery"`
	IsFamilyProduct      bool        `json:"isFamilyProduct"`
	IsQuickshopProduct   bool        `json:"isQuickshopProduct"`
	QuickshopEdit        interface{} `json:"quickshopEdit"`
	NonSlugCategoryQuery interface{} `json:"nonSlugCategoryQuery"`
	SalesAttributes      interface{} `json:"salesAttributes"`
	ShowOosColors        bool        `json:"showOosColors"`
	SkuSelection         struct {
		Bopis struct {
			CurrentStore interface{} `json:"currentStore"`
			SkuInventory struct {
			} `json:"skuInventory"`
		} `json:"bopis"`
		SelectedColor     string      `json:"selectedColor"`
		SelectedFit       string      `json:"selectedFit"`
		SelectedSize      interface{} `json:"selectedSize"`
		SelectedQuantity  interface{} `json:"selectedQuantity"`
		ProductModuleName string      `json:"productModuleName"`
		SizeGuide         interface{} `json:"sizeGuide"`
	} `json:"skuSelection"`
}

var initStateReg = regexp.MustCompile(`window\.urbn\.initialState\s*=\s*JSON\.parse\((".*"),\s*freezeReviver\);`)

func (c *_Crawler) parseProduct(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	matched := initStateReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		c.httpClient.Jar().Clear(ctx, resp.Request.URL)

		c.logger.Debugf("product data not found")
		return fmt.Errorf("extract product data from %s failed, may dynamic page", resp.Request.URL)
	}

	initState := map[string]json.RawMessage{}
	initStateData, err := strconv.Unquote(string(matched[1]))
	if err != nil {
		return fmt.Errorf("unquote init state data failed, error=%s", err)
	}

	if err := json.Unmarshal([]byte(initStateData), &initState); err != nil {
		c.logger.Debugf("data: %s", matched[1])
		return err
	}

	var p *rawProduct
	for key, val := range initState {
		if val != nil && strings.HasPrefix(key, "product--") {
			var prod rawProduct
			data, _ := val.MarshalJSON()
			if err := json.Unmarshal(data, &prod); err != nil {
				return err
			}
			p = &prod
			break
		}
	}
	if p == nil {
		return fmt.Errorf("extract product info from %s failed, not product info found", resp.Request.URL)
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
			Id:           p.CatalogData.ProductID,
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
		},
		BrandName:   p.CatalogData.Product.Brand,
		Title:       p.CatalogData.Product.DisplayName,
		Description: html.UnescapeString(p.CatalogData.Product.LongDescription),
		CrowdType:   p.CatalogData.Product.ParentCategoryID,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
			Current:  int32(p.CatalogData.SkuInfo.SalePriceLow * 100),
			Msrp:     int32(p.CatalogData.SkuInfo.ListPriceHigh * 100),
		},
		//Stock: &pbItem.Stock{},
		Stats: &pbItem.Stats{
			ReviewCount: int32(p.CatalogData.Reviews.Count),
			Rating:      float32(p.CatalogData.Reviews.AverageRating),
		},
		ExtraInfo: map[string]string{},
	}
	for i, cate := range p.Breadcrumbs {
		switch i {
		case 0:
			item.Category = cate.DisplayName
		case 1:
			item.SubCategory = cate.DisplayName
		case 2:
			item.SubCategory2 = cate.DisplayName
		case 3:
			item.SubCategory3 = cate.DisplayName
		case 4:
			item.SubCategory4 = cate.DisplayName
		}
	}

	// skus
	for _, sliceItem := range p.CatalogData.SkuInfo.PrimarySlice.SliceItems {
		// NOTE: if this image domain failed, changed to images.urbanoutfitters.com
		medias := []*media.Media{}
		imgRawUrl := fmt.Sprintf("https://s7d5.scene7.com/is/image/UrbanOutfitters/%s", sliceItem.ID)
		for i, size := range sliceItem.Images {
			imgdata, _ := anypb.New(&media.Media_Image{
				OriginalUrl: fmt.Sprintf("%s_%s", imgRawUrl, size),
				SmallUrl:    fmt.Sprintf("%s_%s?$xlarge$&fit=constrain&qlt=80&wid=500", imgRawUrl, size),
				MediumUrl:   fmt.Sprintf("%s_%s?$xlarge$&fit=constrain&qlt=80&wid=600", imgRawUrl, size),
				LargeUrl:    fmt.Sprintf("%s_%s?$xlarge$&fit=constrain&qlt=80&wid=1000", imgRawUrl, size),
			})
			medias = append(medias, &media.Media{
				Detail:    imgdata,
				IsDefault: i == 0,
			})
		}
		for _, i := range sliceItem.IncludedSkus {
			sku := pbItem.Sku{
				SourceId: i.SkuID,
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(i.SalePrice * 100),
					Msrp:     int32(i.ListPrice * 100),
				},
				Stock: &pbItem.Stock{
					StockStatus: pbItem.Stock_InStock,
					StockCount:  int32(i.StockLevel),
				},
				Medias: medias,
			}
			if i.StockLevel == 0 {
				sku.Stock.StockStatus = pbItem.Stock_OutOfStock
			}
			sku.Specs = append(sku.Specs,
				&pbItem.SkuSpecOption{
					Id:    sliceItem.Code,
					Type:  pbItem.SkuSpecType_SkuSpecColor,
					Value: sliceItem.DisplayName,
					Name:  sliceItem.DisplayName,
					Icon:  sliceItem.SwatchURL,
				},
				&pbItem.SkuSpecOption{
					Id:    i.SizeID,
					Type:  pbItem.SkuSpecType_SkuSpecSize,
					Value: i.Size,
					Name:  i.Size,
				},
			)
			item.SkuItems = append(item.SkuItems, &sku)
		}
	}
	fmt.Println(`item.Stock `, item.Stock)
	for _, rawSku := range item.SkuItems {
		if rawSku.Stock.StockStatus == pbItem.Stock_InStock {
			item.Stock = &pbItem.Stock{StockStatus: pbItem.Stock_InStock}
		}
	}
	fmt.Println(`item.Stock `, item.Stock)
	if item.Stock == nil {
		item.Stock = &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock}
	}
	fmt.Println(`item.Stock `, item.Stock)
	return yield(ctx, &item)
}

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://www.urbanoutfitters.com/",
		// "https://www.urbanoutfitters.com/shop/uo-rugby-cotton-shirt?category=mens-clothing-sale&color=001",
		"https://www.urbanoutfitters.com/shop/uo-retro-sport-colorblock-crew-neck-sweatshirt?category=mens-clothing-sale&color=004&type=REGULAR&quantity=1&size=L",
		//"https://www.urbanoutfitters.com/mens-clothing-sale",
		// "https://www.urbanoutfitters.com/api/catalog/v2/uo-us/pools/US_DIRECT/navigation-items/mens-clothing-sale/products?page-size=96&skip=192&projection-slug=categorytiles",
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

// main func is the entry of golang program. this will not be used by plugin, just for local spider test.
func main() {
	cli.NewApp(New).Run(os.Args)
}
