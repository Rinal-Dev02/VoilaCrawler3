package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/voiladev/VoilaCrawler/pkg/cli"
	"github.com/voiladev/VoilaCrawler/pkg/context"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	pbProxy "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/randutil"
	"github.com/voiladev/go-framework/strconv"
)

type _Crawler struct {
	httpClient http.Client

	categoryPathMatcher *regexp.Regexp
	productPathMatcher  *regexp.Regexp
	logger              glog.Log
}

func (_ *_Crawler) New(_ *cli.Context, client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient: client,

		categoryPathMatcher: regexp.MustCompile(`^/collections(/[a-zA-Z0-9\-]+){1,3}(/?((\+?category|color|price|size)_[a-z0-9_\-]+)*)?$`),
		productPathMatcher:  regexp.MustCompile(`^(/collections(/[a-zA-Z0-9\-]+){1,3})?/products/[a-z0-9\-]+(/?((\+?category|color|price|size)_[a-z0-9_\-]+)*)?$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "6216a41f3c706ae711a24c7d7a389953"
}

// Version
func (c *_Crawler) Version() int32 {
	return 1
}

// CrawlOptions
func (c *_Crawler) CrawlOptions(u *url.URL) *crawler.CrawlOptions {
	options := crawler.NewCrawlOptions()
	options.EnableSessionInit = true
	options.Reliability = pbProxy.ProxyReliability_ReliabilityMedium
	options.MustCookies = append(options.MustCookies,
		&http.Cookie{Name: "cart_currency", Value: "USD", Path: "/"},
		&http.Cookie{Name: "_landing_page", Value: "%2F", Path: "/"},
		&http.Cookie{Name: "zCountry", Value: "US", Path: "/"},
		&http.Cookie{Name: "zHello", Value: "1", Path: "/"},
	)
	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"jingus.com", "*.jingus.com"}
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
		u.Host = "jingus.com"
	}
	if c.productPathMatcher.MatchString(u.Path) {
		if !strings.HasPrefix(u.Path, "/products/") {
			fields := strings.SplitN(u.Path, "/products/", 2)
			u.Path = "/products/" + fields[1]
		}
		u.RawQuery = ""
		return u.String(), nil
	}
	return rawurl, nil
}

func (c *_Crawler) Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	if !crawler.IsTargetTypeSupported(ctx, &pbItem.Product{}) {
		return crawler.ErrUnsupportedTarget
	}

	if resp.Request.URL.Path == "" || resp.Request.URL.Path == "/" {
		return c.parseCategories(ctx, resp, yield)
	}

	if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yield)
	} else if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	}
	return crawler.ErrUnsupportedPath
}

// parseCategories
func (c *_Crawler) parseCategories(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	dom, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		c.logger.Error(err)
		return err
	}

	urlFilters := map[string]struct{}{}
	sel := dom.Find(`.Header__MainNav a[href]`)

	for i := range sel.Nodes {
		node := sel.Eq(i)

		href := node.AttrOr("href", "")
		if href == "" {
			continue
		}
		u, err := url.Parse(href)
		if err != nil {
			c.logger.Error("got invalid url %s", href)
			continue
		}

		if matched, _ := filepath.Match("/pages/*", u.Path); matched || u.Path == "" || u.Path == "/" {
			continue
		}
		if _, ok := urlFilters[u.Path]; ok {
			continue
		}
		urlFilters[u.Path] = struct{}{}

		nctx := ctx
		if !c.productPathMatcher.MatchString(u.Path) && c.categoryPathMatcher.MatchString(u.Path) {
			nctx = context.WithValue(ctx, crawler.TracingIdKey, randutil.MustNewRandomID())
		}
		req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
		if err = yield(nctx, req); err != nil {
			return err
		}
	}
	return nil
}

// nextIndex used to get sharingData from context
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")) + 1)
}

// parseCategoryProducts parse api url from web page url
func (c *_Crawler) parseCategoryProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error(err)
		return err
	}
	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	lastIndex := nextIndex(ctx)
	sel := dom.Find(`.ProductList .ProductItem .ProductItem__Wrapper`)
	for i := range sel.Nodes {
		node := sel.Eq(i)

		href := node.Find("a.ProductItem__ImageWrapper").AttrOr("href", "")
		if href == "" {
			c.logger.Warnf("no href found")
			continue
		}

		req, err := http.NewRequest(http.MethodGet, href, nil)
		if err != nil {
			c.logger.Errorf("create request with url %s failed", href)
			continue
		}
		nctx := context.WithValue(ctx, "item.index", lastIndex)
		lastIndex += 1
		if err := yield(nctx, req); err != nil {
			return err
		}
	}

	nextUrl := dom.Find(`.Pagination__Nav>a[rel="next"]`).AttrOr("href", "")
	if nextUrl == "" {
		return nil
	}

	req, err := http.NewRequest(http.MethodGet, nextUrl, nil)
	if err != nil {
		return c.logger.Errorf("create request with url %s failed", nextUrl).ToError()
	}
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
}

type product struct {
	Product struct {
		ID                   uint64   `json:"id"`
		Title                string   `json:"title"`
		Handle               string   `json:"handle"`
		Description          string   `json:"description"`
		PublishedAt          string   `json:"published_at"`
		CreatedAt            string   `json:"created_at"`
		Vendor               string   `json:"vendor"`
		Type                 string   `json:"type"`
		Tags                 []string `json:"tags"`
		Price                int32    `json:"price"`
		PriceMin             int32    `json:"price_min"`
		PriceMax             int32    `json:"price_max"`
		Available            bool     `json:"available"`
		PriceVaries          bool     `json:"price_varies"`
		CompareAtPrice       int32    `json:"compare_at_price"`
		CompareAtPriceMin    int32    `json:"compare_at_price_min"`
		CompareAtPriceMax    int32    `json:"compare_at_price_max"`
		CompareAtPriceVaries bool     `json:"compare_at_price_varies"`
		Variants             []struct {
			ID                     uint64        `json:"id"`
			Title                  string        `json:"title"`
			Option1                string        `json:"option1"`
			Option2                string        `json:"option2"`
			Option3                string        `json:"option3"`
			Sku                    string        `json:"sku"`
			RequiresShipping       bool          `json:"requires_shipping"`
			Taxable                bool          `json:"taxable"`
			FeaturedImage          interface{}   `json:"featured_image"`
			Available              bool          `json:"available"`
			Name                   string        `json:"name"`
			PublicTitle            string        `json:"public_title"`
			Options                []string      `json:"options"`
			Price                  int32         `json:"price"`
			Weight                 int32         `json:"weight"`
			CompareAtPrice         int32         `json:"compare_at_price"`
			InventoryManagement    string        `json:"inventory_management"`
			Barcode                string        `json:"barcode"`
			RequiresSellingPlan    bool          `json:"requires_selling_plan"`
			SellingPlanAllocations []interface{} `json:"selling_plan_allocations"`
		} `json:"variants"`
		Images        []string `json:"images"`
		FeaturedImage string   `json:"featured_image"`
		Options       []string `json:"options"`
		Media         []struct {
			Alt          string `json:"alt"`
			ID           uint64 `json:"id"`
			Position     int    `json:"position"`
			PreviewImage struct {
				AspectRatio float64 `json:"aspect_ratio"`
				Height      int     `json:"height"`
				Width       int     `json:"width"`
				Src         string  `json:"src"`
			} `json:"preview_image"`
			AspectRatio float64 `json:"aspect_ratio"`
			Height      int     `json:"height"`
			MediaType   string  `json:"media_type"`
			Src         string  `json:"src"`
			Width       int     `json:"width"`
		} `json:"media"`
		RequiresSellingPlan bool          `json:"requires_selling_plan"`
		SellingPlanGroups   []interface{} `json:"selling_plan_groups"`
		Content             string        `json:"content"`
	} `json:"product"`
	SelectedVariantID int64 `json:"selected_variant_id"`
	Inventories       map[string]struct {
		InventoryManagement string `json:"inventory_management"`
		InventoryPolicy     string `json:"inventory_policy"`
		InventoryQuantity   int32  `json:"inventory_quantity"`
		InventoryMessage    string `json:"inventory_message"`
	} `json:"inventories"`
}

// CB-A610PK+A617PKS

var (
	groupIdReg = regexp.MustCompile(`((?:\d+)?[A-Z]+\d+)[A-Z/]+`)
)

func (c *_Crawler) parseProduct(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error(err)
		return err
	}
	if bytes.Contains(respBody, []byte("The page you are looking for cannot be found.")) {
		return crawler.ErrAbort
	}

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	rawProductJson := strings.TrimSpace(dom.Find(`script[data-product-json]`).Text())
	if rawProductJson == "" {
		err := errors.New("get no product json data from html")
		c.logger.Debug(err)
		return err
	}

	var viewData product
	if err := json.Unmarshal([]byte(rawProductJson), &viewData); err != nil {
		c.logger.Error(err)
		return err
	}
	prod := viewData.Product

	canurl := dom.Find(`link[rel="canonical"]`).AttrOr("href", "")
	if canurl == "" {
		canurl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           strconv.Format(prod.ID),
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canurl,
		},
		BrandName:   prod.Vendor,
		Title:       prod.Title,
		Description: prod.Description,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		ExtraInfo: map[string]string{
			"product_type": prod.Type,
			"tags":         strings.Join(prod.Tags, ","),
		},
	}

	// for category
	breadSel := dom.Find(`.CollectionToolbar .CollectionToolbar__Item nav.CurrentPath>a`)
	for i := range breadSel.Nodes {
		node := breadSel.Eq(i)
		switch i {
		case 1:
			item.Category = strings.TrimSpace(node.Text())
		case 2:
			item.SubCategory = strings.TrimSpace(node.Text())
		case 3:
			item.SubCategory2 = strings.TrimSpace(node.Text())
		}
	}

	optType := map[int]pbItem.SkuSpecType{}
	for i, opt := range prod.Options {
		switch strings.ToLower(opt) {
		case "color", "colour":
			optType[i+1] = pbItem.SkuSpecType_SkuSpecColor
		case "size":
			optType[i+1] = pbItem.SkuSpecType_SkuSpecSize
		}
	}

	var medias []*media.Media
	for _, img := range prod.Media {
		u, err := url.Parse(img.Src)
		if err != nil {
			c.logger.Errorf("got invalid img url %s", img.Src)
		}
		fields := strings.SplitN(u.Path, ".", 2)
		tpl := strings.Replace(img.Src, u.Path, fields[0]+"_%s."+fields[1], -1)
		medias = append(medias, media.NewImageMedia(
			strconv.Format(img.ID),
			img.Src,
			fmt.Sprintf(tpl, "1000x"),
			fmt.Sprintf(tpl, "600x"),
			fmt.Sprintf(tpl, "500x"),
			img.Alt,
			img.Position == 1))
	}

	// variants
	for _, variant := range prod.Variants {
		current := variant.Price
		msrp := variant.CompareAtPrice
		discount := 0.0
		if msrp > current {
			discount = math.Round(1000 * float64(msrp-current) / float64(msrp))
		}

		if item.Source.GroupId == "" {
			matched := groupIdReg.FindAllStringSubmatch(variant.Sku, -1)
			groupId := ""
			for _, field := range matched {
				if len(field) != 2 {
					continue
				}
				if groupId == "" {
					groupId = field[1]
				} else {
					groupId = groupId + "-" + field[1]
				}
			}
			item.Source.GroupId = groupId
		}

		sku := pbItem.Sku{
			SourceId: variant.Sku,
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(current),
				Msrp:     int32(msrp),
				Discount: int32(discount),
			},
			Stock: &pbItem.Stock{
				StockStatus: pbItem.Stock_OutOfStock,
			},
		}
		if quan, ok := viewData.Inventories[strconv.Format(variant.ID)]; ok && quan.InventoryQuantity > 0 {
			sku.Stock.StockStatus = pbItem.Stock_InStock
			sku.Stock.StockCount = quan.InventoryQuantity
		}

		for i, opt := range []string{variant.Option1, variant.Option2, variant.Option3} {
			if opt == "" {
				continue
			}
			if typ, ok := optType[i+1]; ok {
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  typ,
					Id:    opt,
					Name:  opt,
					Value: opt,
				})
			}
		}
		sku.Medias = medias
		item.SkuItems = append(item.SkuItems, &sku)
	}
	if err := yield(ctx, &item); err != nil {
		return err
	}
	colorSel := dom.Find(`.ColorSwatchList .HorizontalList__Item>a`)
	for i := range colorSel.Nodes {
		node := colorSel.Eq(i)

		href := node.AttrOr("href", "")
		req, err := http.NewRequest(http.MethodGet, href, nil)
		if err != nil {
			c.logger.Errorf("new request failed, error=%s", err)
			return err
		}
		if err := yield(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		"https://jingus.com/collections/sale/products/canyon-white-strappy-ladder-back-bra-top",
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

func main() {
	cli.NewApp(&_Crawler{}).Run(os.Args)
}
