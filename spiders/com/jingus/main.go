package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/voiladev/go-crawler/pkg/cli"
	"github.com/voiladev/go-crawler/pkg/crawler"
	"github.com/voiladev/go-crawler/pkg/net/http"
	"github.com/voiladev/go-crawler/pkg/sdk/shopify"
	"github.com/voiladev/go-crawler/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/go-crawler/protoc-gen-go/chameleon/api/regulation"
	pbCrawl "github.com/voiladev/go-crawler/protoc-gen-go/chameleon/smelter/v1/crawl"
	pbItem "github.com/voiladev/go-crawler/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	pbProxy "github.com/voiladev/go-crawler/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
)

type _Crawler struct {
	httpClient    http.Client
	shopifyClient *shopify.ShopifyClient
	collections   sync.Map

	categoryPathMatcher *regexp.Regexp
	productPathMatcher  *regexp.Regexp
	logger              glog.Log
}

const (
	__shopName__   = "jinglimited"
	__apiVersion__ = "2021-04"
)

func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	shopifyClient, err := shopify.New(__shopName__, __apiVersion__,
		os.Getenv("JING_API_KEY"), os.Getenv("JING_API_SECRET"), os.Getenv("JING_API_ACCESSTOKEN"))
	if err != nil {
		return nil, err
	}

	c := _Crawler{
		httpClient:    client,
		shopifyClient: shopifyClient,

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
	options.Reliability = pbProxy.ProxyReliability_ReliabilityMedium
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

type PageState struct {
	A       int64  `json:"a"`
	Offset  int64  `json:"offset"`
	Reqid   string `json:"reqid"`
	Pageurl string `json:"pageurl"`
	U       string `json:"u"`
	P       string `json:"p"`
	Rtyp    string `json:"rtyp"`
	Rid     uint64 `json:"rid"`
}

func (c *_Crawler) Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) || c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		// extract product id or collection id
		dom, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return err
		}

		st := strings.TrimSpace(dom.Find("#__st").Text())
		st = strings.TrimSuffix(strings.TrimPrefix(st, "var __st="), ";")

		var pageState PageState
		if err := json.Unmarshal([]byte(st), &pageState); err != nil {
			return fmt.Errorf("parse page state %s failed, error=%s", st, err)
		}
		switch pageState.Rtyp {
		case "collection":
			return c.parseCategoryProducts(ctx, strconv.Format(pageState.Rid), yield)
		case "product":
			return c.parseProduct(ctx, strconv.Format(pageState.Rid), nil, yield)
		default:
			return fmt.Errorf("unsupported resource type %s", pageState.Rtyp)
		}
	}
	return fmt.Errorf("unsupported url %s", resp.Request.URL.String())
}

// nextIndex used to get sharingData from context
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")) + 1)
}

// parseCategoryProducts parse api url from web page url
func (c *_Crawler) parseCategoryProducts(ctx context.Context, cid string, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	nextLink := ""
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		resp, err := c.shopifyClient.ListCollectionProducts(ctx, &shopify.ListCollectionProductsRequest{
			CollectionID: cid,
			Link:         nextLink,
		})
		if err != nil {
			c.logger.Error(err)
			return err
		}

		lastIndex := nextIndex(ctx)
		for _, prod := range resp.Products {
			nctx := context.WithValue(ctx, "item.index", lastIndex)
			lastIndex += 1
			if err := c.parseProduct(nctx, strconv.Format(prod.ID), prod, yield); err != nil {
				return err
			}
		}
		nextLink = resp.NextLink
		if nextLink == "" {
			break
		}
	}
	return nil
}

var groupIdReg = regexp.MustCompile(`^(\d+[A-Z]+\d+)[A-Z/]+$`)

func (c *_Crawler) parseProduct(ctx context.Context, pid string, prod *shopify.Product, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}
	if pid == "" {
		return errors.New("invalid product id")
	}

	if prod == nil {
		// get product detail
		resp, err := c.shopifyClient.GetProduct(ctx, &shopify.GetProductRequest{ProductID: pid})
		if err != nil {
			return err
		}
		prod = resp.Product
	}

	canurl := fmt.Sprintf("https://jingus.com/products/%s", prod.Handle)
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           pid,
			CrawlUrl:     canurl,
			CanonicalUrl: canurl,
		},
		BrandName:   prod.Vendor,
		Title:       prod.Title,
		Description: prod.BodyHTML,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		ExtraInfo: map[string]string{
			"product_type": prod.ProductType,
			"tags":         prod.Tags,
		},
	}

	// get product collections
	if resp, err := c.shopifyClient.ListCollects(ctx, &shopify.ListCollectsRequest{
		Limit:     250,
		ProductID: pid,
	}); err != nil {
		c.logger.Error(err)
		return err
	} else {
		for _, collect := range resp.Collects {
			var (
				id   = strconv.Format(collect.CollectionID)
				coll *shopify.Collection
			)
			if val, ok := c.collections.Load(id); ok {
				coll = val.(*shopify.Collection)
			} else if resp, err := c.shopifyClient.GetCollection(ctx, &shopify.GetCollectionRequest{ID: id}); err != nil {
				c.logger.Error(err)
				continue
			} else {
				coll = resp.Collection
				c.collections.Store(id, coll)
			}
			item.Categories = append(item.Categories, &pbItem.Category{
				Id:   id,
				Name: coll.Title,
			})
		}
	}

	optType := map[int]pbItem.SkuSpecType{}
	for _, opt := range prod.Options {
		switch strings.ToLower(opt.Name) {
		case "color", "colour":
			optType[opt.Position] = pbItem.SkuSpecType_SkuSpecColor
		case "size":
			optType[opt.Position] = pbItem.SkuSpecType_SkuSpecSize
		}
	}

	// variants
	for _, variant := range prod.Variants {
		current, _ := strconv.ParsePrice(variant.Price)
		msrp, _ := strconv.ParsePrice(variant.CompareAtPrice)
		discount := 0.0
		if msrp > current {
			discount = math.Round(100 * float64(msrp-current) / float64(msrp))
		}

		if item.Source.GroupId == "" {
			matched := groupIdReg.FindStringSubmatch(variant.Sku)
			if len(matched) != 2 {
				yield(ctx, &pbCrawl.Error{ErrMsg: "no group id found"})
			}
			item.Source.GroupId = matched[1]
		}

		sku := pbItem.Sku{
			SourceId: variant.Sku,
			Title:    variant.Title,
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(current * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},
			Stock: &pbItem.Stock{
				StockStatus: pbItem.Stock_OutOfStock,
			},
		}
		if variant.InventoryQuantity > 0 {
			sku.Stock.StockStatus = pbItem.Stock_InStock
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

		var medias []*media.Media
		for _, img := range prod.Images {
			if img.ID != variant.ImageID && variant.ImageID != 0 {
				continue
			}
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
		sku.Medias = medias

		item.SkuItems = append(item.SkuItems, &sku)
	}
	return yield(ctx, &item)
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
	cli.NewApp(New).Run(os.Args)
}
