package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"math"
	"net/url"
	"os"
	"regexp"
	"strings"

	uuid "github.com/satori/go.uuid"
	"github.com/voiladev/VoilaCrawler/pkg/cli"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	pbMedia "github.com/voiladev/VoilaCrawler/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawler/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawler/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	pbProxy "github.com/voiladev/VoilaCrawler/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/randutil"
	"github.com/voiladev/go-framework/strconv"
)

type _Crawler struct {
	httpClient http.Client

	searchPathMatcher   *regexp.Regexp
	categoryPathMatcher *regexp.Regexp
	productPathMatcher  *regexp.Regexp
	logger              glog.Log
}

func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient:          client,
		searchPathMatcher:   regexp.MustCompile(`^/exp/search$`),
		categoryPathMatcher: regexp.MustCompile(`^(?:/[a-z0-9_\-]+){1,4}/(cat\d+)(?:/[a-z0-9_\-]+){0,4}$`),
		productPathMatcher:  regexp.MustCompile(`^(/clothing(?:/[.a-zA-Z0-9\pL\pS\-]+){1,4}/pro/([0-9]+))(?:/cat\d+)?(?:/color/[a-zA-Z0-9\s%]+(?:/[a-z0-9]+){0,2})?/?$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "78ba5f22841c93a71ede0c4e7a6c99a8"
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
	options.MustCookies = append(options.MustCookies,
		&http.Cookie{Name: "AKA_A2", Value: "A", Path: "/"},
		&http.Cookie{Name: "siteType", Value: "A", Path: "/"},
		&http.Cookie{Name: "isMobile", Value: "false", Path: "/"},
		&http.Cookie{Name: "isTablet", Value: "false", Path: "/"},
		&http.Cookie{Name: "AWS_Exp_100", Value: "TRUE", Path: "/"},
		&http.Cookie{Name: "awsexp", Value: "true", Path: "/"},
		&http.Cookie{Name: "at_check", Value: "true", Path: "/"},
		&http.Cookie{Name: "s_sess", Value: "%20s_cc%3Dtrue%3B", Path: "/"},
		// &http.Cookie{Name: "geoloc", Value: "cc=US,rc=CA,tp=vhigh,tz=PST,la=33.9733,lo=-118.2487"},
	)
	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.express.com"}
}

func (c *_Crawler) CanonicalUrl(rawurl string) (string, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	matched := c.productPathMatcher.FindStringSubmatch(u.Path)
	if len(matched) == 3 {
		u.Path = matched[1]
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	if u.Host == "" {
		u.Host = "www.express.com"
	}
	return u.String(), nil
}

func (c *_Crawler) Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	if c.searchPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseSearch(ctx, resp, yield)
	} else if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) || resp.Request.URL.String() == "https://www.express.com/graphql" {
		return c.parseCategoryProducts(ctx, resp, yield)
	} else if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yield)
	}
	return crawler.ErrUnsupportedPath
}

type product struct {
	AverageOverallRating float64 `json:"averageOverallRating"`
	Colors               []struct {
		Color      string `json:"color"`
		SkuUpc     string `json:"skuUpc"`
		DefaultSku bool   `json:"defaultSku"`
		Typename   string `json:"__typename"`
	} `json:"colors"`
	EFOProduct              bool        `json:"EFOProduct"`
	EnsembleListPrice       interface{} `json:"ensembleListPrice"`
	EnsembleSalePrice       interface{} `json:"ensembleSalePrice"`
	Key                     string      `json:"key"`
	IsEnsemble              bool        `json:"isEnsemble"`
	ListPrice               string      `json:"listPrice"`
	MarketplaceProduct      interface{} `json:"marketplaceProduct"`
	Name                    string      `json:"name"`
	NewProduct              bool        `json:"newProduct"`
	OnlineExclusive         bool        `json:"onlineExclusive"`
	OnlineExclusivePromoMsg interface{} `json:"onlineExclusivePromoMsg"`
	PaginationEnd           interface{} `json:"paginationEnd"`
	PaginationStart         int         `json:"paginationStart"`
	ProductDescription      string      `json:"productDescription"`
	ProductID               string      `json:"productId"`
	ProductImage            string      `json:"productImage"`
	ProductURL              string      `json:"productURL"`
	PromoMessage            string      `json:"promoMessage"`
	SalePrice               string      `json:"salePrice"`
	TotalReviewCount        int         `json:"totalReviewCount"`
	Typename                string      `json:"__typename"`
}

var productsReviewExtractReg = regexp.MustCompile(`(?Ums)<script type="application/ld\+json"([ a-z\-\="])+>({.*})</script>`)

type SearchView struct {
	Data struct {
		GetUnbxdSearch struct {
			Facets []struct {
				FacetID  string   `json:"facetId"`
				Name     string   `json:"name"`
				Position int      `json:"position"`
				Values   []string `json:"values"`
				Typename string   `json:"__typename"`
			} `json:"facets"`
			DidYouMean interface{} `json:"didYouMean"`
			Pagination struct {
				TotalProductCount int    `json:"totalProductCount"`
				PageNumber        int    `json:"pageNumber"`
				PageSize          int    `json:"pageSize"`
				Start             int    `json:"start"`
				End               int    `json:"end"`
				Typename          string `json:"__typename"`
			} `json:"pagination"`
			Products []*product  `json:"products"`
			Redirect interface{} `json:"redirect"`
			Source   string      `json:"source"`
			Typename string      `json:"__typename"`
		} `json:"getUnbxdSearch"`
	} `json:"data"`
	Extensions struct {
		Platform string `json:"platform"`
	} `json:"extensions"`
}

var searchAPIReqBody = `{"operationName":"SearchQuery","variables":{"filter":"","rows":56,"searchTerm":"%s","sort":"","start":%d},"query":"query SearchQuery($searchTerm: String!, $start: Int!, $rows: Int, $filter: String, $sort: String) {\n  getUnbxdSearch(searchTerm: $searchTerm, start: $start, rows: $rows, filter: $filter, sort: $sort) {\n    facets {\n      facetId\n      name\n      position\n      values\n      __typename\n    }\n    didYouMean {\n      suggestion\n      frequency\n      __typename\n    }\n    pagination {\n      totalProductCount\n      pageNumber\n      pageSize\n      start\n      end\n      __typename\n    }\n    products {\n      colors {\n        color\n        skuUpc\n        defaultSku\n        __typename\n      }\n      isEnsemble\n      ensembleListPrice\n      ensembleSalePrice\n      key\n      listPrice\n      name\n      newProduct\n      onlineExclusive\n      onlineExclusivePromoMsg\n      paginationEnd\n      paginationStart\n      productDescription\n      productId\n      productImage\n      productURL\n      promoMessage\n      salePrice\n      __typename\n    }\n    redirect {\n      type\n      value\n      __typename\n    }\n    source\n    __typename\n  }\n}\n"}`

// parseSearch parse api url from web page url
func (c *_Crawler) parseSearch(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	opts := c.CrawlOptions(resp.Request.URL)

	vals := resp.Request.URL.Query()
	q := strings.TrimSpace(vals.Get("q"))
	if q == "" {
		return nil
	}

	var (
		start  = 0
		subCtx = ctx
	)
	for {
		reqBody := fmt.Sprintf(searchAPIReqBody, q, start)
		req, _ := http.NewRequest(http.MethodPost, "https://www.express.com/graphql", strings.NewReader(reqBody))
		req.Header.Add("accept", "*/*")
		req.Header.Add("referer", resp.Request.URL.String())
		req.Header.Add("origin", "https://www.express.com")
		req.Header.Add("content-type", "application/json")
		req.Header.Add("cache-control", "no-cache")
		if ua := resp.Request.Header.Get("User-Agent"); ua != "" {
			req.Header.Set("User-Agent", ua)
		}
		req.Header.Add("x-exp-request-id", uuid.NewV4().String())
		req.Header.Add("x-exp-rvn-cacheable", "false")
		req.Header.Add("x-exp-rvn-query-classification", "getUnbxdCategory")
		req.Header.Add("x-exp-rvn-source", "app_express.com")

		c.logger.Debugf("Access %s CategoryQuery %d", req.URL, start)
		resp, err := c.httpClient.DoWithOptions(subCtx, req, http.Options{
			EnableProxy:       true,
			EnableHeadless:    false,
			EnableSessionInit: false,
			Reliability:       opts.Reliability,
			RequestFilterKeys: []string{"CategoryQuery"},
		})
		if err != nil {
			c.logger.Error(err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusForbidden ||
			resp.StatusCode == -1 {
			return fmt.Errorf("net work request failed, %s", resp.Status)
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			c.logger.Error(err)
			return err
		}
		var viewData SearchView
		if err := json.Unmarshal(respBody, &viewData); err != nil {
			c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
			return err
		}

		lastIndex := nextIndex(ctx)
		for _, idv := range viewData.Data.GetUnbxdSearch.Products {
			rawurl := idv.ProductURL
			req, err := http.NewRequest(http.MethodGet, rawurl, nil)
			if err != nil {
				c.logger.Errorf("load http request of url %s failed, error=%s", rawurl, err)
				continue
			}
			nctx := context.WithValue(ctx, "item.index", lastIndex)
			lastIndex += 1
			if err := yield(nctx, req); err != nil {
				return err
			}
		}

		if viewData.Data.GetUnbxdSearch.Pagination.End >= viewData.Data.GetUnbxdSearch.Pagination.TotalProductCount ||
			lastIndex >= viewData.Data.GetUnbxdSearch.Pagination.TotalProductCount {
			break
		}
		start = viewData.Data.GetUnbxdSearch.Pagination.End + 1
		subCtx = context.WithValue(ctx, crawler.ReqIdKey, randutil.MustNewRandomID())
	}
	return nil
}

type CategoryView struct {
	Data struct {
		GetUnbxdCategory struct {
			CategoryID   string `json:"categoryId"`
			CategoryName string `json:"categoryName"`
			Facets       []struct {
				FacetID  string   `json:"facetId"`
				Name     string   `json:"name"`
				Position int      `json:"position"`
				Values   []string `json:"values"`
				Typename string   `json:"__typename"`
			} `json:"facets"`
			Pagination struct {
				TotalProductCount int    `json:"totalProductCount"`
				PageNumber        int    `json:"pageNumber"`
				PageSize          int    `json:"pageSize"`
				Start             int    `json:"start"`
				End               int    `json:"end"`
				Typename          string `json:"__typename"`
			} `json:"pagination"`
			Products []*product `json:"products"`
			Source   string     `json:"source"`
			Typename string     `json:"__typename"`
		} `json:"getUnbxdCategory"`
	} `json:"data"`
	Extensions struct {
		Platform string `json:"platform"`
	} `json:"extensions"`
}

var cateAPIReqBody = `{"operationName":"CategoryQuery","variables":{"categoryId":"%s","filter":"","overrideCatApi":"","rows":56,"sort":"","start":%d},"query":"query CategoryQuery($categoryId: String, $start: Int!, $rows: Int, $filter: String, $sort: String, $overrideCatApi: String, $uc_param: String) {\n  getUnbxdCategory(categoryId: $categoryId, start: $start, rows: $rows, filter: $filter, sort: $sort, overrideCatApi: $overrideCatApi, uc_param: $uc_param) {\n    categoryId\n    categoryName\n    facets {\n      facetId\n      name\n      position\n      values\n      __typename\n    }\n    pagination {\n      totalProductCount\n      pageNumber\n      pageSize\n      start\n      end\n      __typename\n    }\n    products {\n      averageOverallRating\n      colors {\n        color\n        skuUpc\n        defaultSku\n        __typename\n      }\n      EFOProduct\n      ensembleListPrice\n      ensembleSalePrice\n      key\n      isEnsemble\n      listPrice\n      marketplaceProduct\n      name\n      newProduct\n      onlineExclusive\n      onlineExclusivePromoMsg\n      paginationEnd\n      paginationStart\n      productDescription\n      productId\n      productImage\n      productURL\n      promoMessage\n      salePrice\n      totalReviewCount\n      __typename\n    }\n    source\n    __typename\n  }\n}\n"}`

// parseCategoryProducts parse api url from web page url
func (c *_Crawler) parseCategoryProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	opts := c.CrawlOptions(resp.Request.URL)

	matched := c.categoryPathMatcher.FindStringSubmatch(resp.Request.URL.Path)
	cateId := matched[1]

	var (
		start  = 0
		subCtx = ctx
	)
	for {
		reqBody := fmt.Sprintf(cateAPIReqBody, cateId, start)
		req, _ := http.NewRequest(http.MethodPost, "https://www.express.com/graphql", strings.NewReader(reqBody))
		req.Header.Add("accept", "*/*")
		req.Header.Add("referer", resp.Request.URL.String())
		req.Header.Add("origin", "https://www.express.com")
		req.Header.Add("content-type", "application/json")
		req.Header.Add("cache-control", "no-cache")
		if ua := resp.Request.Header.Get("User-Agent"); ua != "" {
			req.Header.Set("User-Agent", ua)
		}
		req.Header.Add("x-exp-request-id", uuid.NewV4().String())
		req.Header.Add("x-exp-rvn-cacheable", "false")
		req.Header.Add("x-exp-rvn-query-classification", "getUnbxdCategory")
		req.Header.Add("x-exp-rvn-source", "app_express.com")

		c.logger.Debugf("Access %s CategoryQuery %d", req.URL, start)
		resp, err := c.httpClient.DoWithOptions(subCtx, req, http.Options{
			EnableProxy:       true,
			EnableHeadless:    false,
			EnableSessionInit: false,
			Reliability:       opts.Reliability,
			RequestFilterKeys: []string{"CategoryQuery"},
		})
		if err != nil {
			c.logger.Error(err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusForbidden ||
			resp.StatusCode == -1 {
			return fmt.Errorf("net work request failed, %s", resp.Status)
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			c.logger.Error(err)
			return err
		}
		var viewData CategoryView
		if err := json.Unmarshal(respBody, &viewData); err != nil {
			c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
			return err
		}

		lastIndex := nextIndex(ctx)
		for _, idv := range viewData.Data.GetUnbxdCategory.Products {
			rawurl := idv.ProductURL
			req, err := http.NewRequest(http.MethodGet, rawurl, nil)
			if err != nil {
				c.logger.Errorf("load http request of url %s failed, error=%s", rawurl, err)
				continue
			}
			nctx := context.WithValue(ctx, "item.index", lastIndex)
			lastIndex += 1
			if err := yield(nctx, req); err != nil {
				return err
			}
		}

		if viewData.Data.GetUnbxdCategory.Pagination.End >= viewData.Data.GetUnbxdCategory.Pagination.TotalProductCount ||
			lastIndex >= viewData.Data.GetUnbxdCategory.Pagination.TotalProductCount {
			break
		}
		start = viewData.Data.GetUnbxdCategory.Pagination.End + 1
		subCtx = context.WithValue(ctx, crawler.ReqIdKey, randutil.MustNewRandomID())
	}
	return nil
}

// nextIndex used to get sharingData from context
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")) + 1)
}

type parseProductResponse struct {
	Data struct {
		Product struct {
			BopisEligible         bool        `json:"bopisEligible"`
			ClearancePromoMessage interface{} `json:"clearancePromoMessage"`
			Collection            interface{} `json:"collection"`
			CrossRelDetailMessage interface{} `json:"crossRelDetailMessage"`
			CrossRelProductURL    interface{} `json:"crossRelProductURL"`
			EFOProduct            bool        `json:"EFOProduct"`
			ExpressProductType    interface{} `json:"expressProductType"`
			FabricCare            string      `json:"fabricCare"`
			FabricDetailImages    []struct {
				Caption  string `json:"caption"`
				Image    string `json:"image"`
				Typename string `json:"__typename"`
			} `json:"fabricDetailImages"`
			Gender                         string      `json:"gender"`
			InternationalShippingAvailable bool        `json:"internationalShippingAvailable"`
			ListPrice                      string      `json:"listPrice"`
			MarketPlaceProduct             bool        `json:"marketPlaceProduct"`
			Name                           string      `json:"name"`
			NewProduct                     interface{} `json:"newProduct"`
			OnlineExclusive                bool        `json:"onlineExclusive"`
			OnlineExclusivePromoMsg        interface{} `json:"onlineExclusivePromoMsg"`
			ProductDescription             []struct {
				Type     string   `json:"type"`
				Content  []string `json:"content"`
				Typename string   `json:"__typename"`
			} `json:"productDescription"`
			ProductFeatures     []string    `json:"productFeatures"`
			ProductID           string      `json:"productId"`
			ProductImage        string      `json:"productImage"`
			ProductInventory    int         `json:"productInventory"`
			ProductURL          string      `json:"productURL"`
			PromoMessage        string      `json:"promoMessage"`
			RecsAlgorithm       string      `json:"recsAlgorithm"`
			OriginRecsAlgorithm interface{} `json:"originRecsAlgorithm"`
			SalePrice           string      `json:"salePrice"`
			Type                string      `json:"type"`
			BreadCrumbCategory  struct {
				CategoryName   string      `json:"categoryName"`
				H1CategoryName interface{} `json:"h1CategoryName"`
				Links          []struct {
					Rel      string `json:"rel"`
					Href     string `json:"href"`
					Typename string `json:"__typename"`
				} `json:"links"`
				BreadCrumbCategory interface{} `json:"breadCrumbCategory"`
				Typename           string      `json:"__typename"`
			} `json:"breadCrumbCategory"`
			ColorSlices []struct {
				Color             string `json:"color"`
				DefaultSlice      bool   `json:"defaultSlice"`
				IPColorCode       string `json:"ipColorCode"`
				HasWaistAndInseam bool   `json:"hasWaistAndInseam"`
				SwatchURL         string `json:"swatchURL"`
				ImageMap          struct {
					All struct {
						LARGE    []string `json:"LARGE"`
						MAIN     []string `json:"MAIN"`
						Typename string   `json:"__typename"`
					} `json:"All"`
					Default struct {
						LARGE    []string `json:"LARGE"`
						MAIN     []string `json:"MAIN"`
						Typename string   `json:"__typename"`
					} `json:"Default"`
					Model1 struct {
						LARGE    []string `json:"LARGE"`
						MAIN     []string `json:"MAIN"`
						Typename string   `json:"__typename"`
					} `json:"Model1"`
					Model2 struct {
						LARGE    []interface{} `json:"LARGE"`
						MAIN     []interface{} `json:"MAIN"`
						Typename string        `json:"__typename"`
					} `json:"Model2"`
					Model3 struct {
						LARGE    []interface{} `json:"LARGE"`
						MAIN     []interface{} `json:"MAIN"`
						Typename string        `json:"__typename"`
					} `json:"Model3"`
					Typename string `json:"__typename"`
				} `json:"imageMap"`
				OnlineSkus []string `json:"onlineSkus"`
				Skus       []struct {
					BackOrderable         bool        `json:"backOrderable"`
					BackOrderDate         interface{} `json:"backOrderDate"`
					DisplayMSRP           string      `json:"displayMSRP"`
					DisplayPrice          string      `json:"displayPrice"`
					Ext                   string      `json:"ext"`
					Inseam                interface{} `json:"inseam"`
					InStoreInventoryCount int         `json:"inStoreInventoryCount"`
					InventoryMessage      interface{} `json:"inventoryMessage"`
					IsFinalSale           bool        `json:"isFinalSale"`
					IsInStockOnline       bool        `json:"isInStockOnline"`
					MiraklOffer           interface{} `json:"miraklOffer"`
					MarketPlaceSku        bool        `json:"marketPlaceSku"`
					OnClearance           bool        `json:"onClearance"`
					OnSale                bool        `json:"onSale"`
					OnlineExclusive       bool        `json:"onlineExclusive"`
					OnlineInventoryCount  int         `json:"onlineInventoryCount"`
					Size                  string      `json:"size"`
					SizeName              string      `json:"sizeName"`
					SkuID                 string      `json:"skuId"`
					Typename              string      `json:"__typename"`
				} `json:"skus"`
				Typename string `json:"__typename"`
			} `json:"colorSlices"`
			OriginRecs      interface{} `json:"originRecs"`
			RelatedProducts interface{} `json:"relatedProducts"`
			Icons           []struct {
				Icon     string `json:"icon"`
				Category string `json:"category"`
				Typename string `json:"__typename"`
			} `json:"icons"`
			Typename string `json:"__typename"`
		} `json:"product"`
	} `json:"data"`
	Extensions struct {
		Platform string `json:"platform"`
	} `json:"extensions"`
}

type parseProductReviewResponse struct {
	AggregateRating struct {
		ReviewCount string `json:"reviewCount"`
		RatingValue string `json:"ratingValue"`
	} `json:"aggregateRating"`
}

var productGraphQLQuery string = `{"operationName":"ProductQuery","variables":{"productId":"%v"},"query":"query ProductQuery($productId: String!) {\n  product(id: $productId) {\n    bopisEligible\n    clearancePromoMessage\n    collection\n    crossRelDetailMessage\n    crossRelProductURL\n    EFOProduct\n    expressProductType\n    fabricCare\n    fabricDetailImages {\n      caption\n      image\n      __typename\n    }\n    gender\n    internationalShippingAvailable\n    listPrice\n    marketPlaceProduct\n    name\n    newProduct\n    onlineExclusive\n    onlineExclusivePromoMsg\n    productDescription {\n      type\n      content\n      __typename\n    }\n    productFeatures\n    productId\n    productImage\n    productInventory\n    productURL\n    promoMessage\n    recsAlgorithm\n    originRecsAlgorithm\n    salePrice\n    type\n    breadCrumbCategory {\n      categoryName\n      h1CategoryName\n      links {\n        rel\n        href\n        __typename\n      }\n      breadCrumbCategory {\n        categoryName\n        h1CategoryName\n        links {\n          rel\n          href\n          __typename\n        }\n        __typename\n      }\n      __typename\n    }\n    colorSlices {\n      color\n      defaultSlice\n      ipColorCode\n      hasWaistAndInseam\n      swatchURL\n      imageMap {\n        All {\n          LARGE\n          MAIN\n          __typename\n        }\n        Default {\n          LARGE\n          MAIN\n          __typename\n        }\n        Model1 {\n          LARGE\n          MAIN\n          __typename\n        }\n        Model2 {\n          LARGE\n          MAIN\n          __typename\n        }\n        Model3 {\n          LARGE\n          MAIN\n          __typename\n        }\n        __typename\n      }\n      onlineSkus\n      skus {\n        backOrderable\n        backOrderDate\n        displayMSRP\n        displayPrice\n        ext\n        inseam\n        inStoreInventoryCount\n        inventoryMessage\n        isFinalSale\n        isInStockOnline\n        miraklOffer {\n          minimumShippingPrice\n          sellerId\n          sellerName\n          __typename\n        }\n        marketPlaceSku\n        onClearance\n        onSale\n        onlineExclusive\n        onlineInventoryCount\n        size\n        sizeName\n        skuId\n        __typename\n      }\n      __typename\n    }\n    originRecs {\n      listPrice\n      marketPlaceProduct\n      name\n      productId\n      productImage\n      productURL\n      salePrice\n      __typename\n    }\n    relatedProducts {\n      listPrice\n      marketPlaceProduct\n      name\n      productId\n      productImage\n      productURL\n      salePrice\n      colorSlices {\n        color\n        defaultSlice\n        __typename\n      }\n      __typename\n    }\n    icons {\n      icon\n      category\n      __typename\n    }\n    __typename\n  }\n}\n"}`

var imgWidthReg = regexp.MustCompile(`wid=\d+`)

func (c *_Crawler) parseProduct(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	var viewReviewData parseProductReviewResponse
	{
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		matched := productsReviewExtractReg.FindSubmatch([]byte(html.UnescapeString(string(respBody))))
		if len(matched) > 2 {
			if err := json.Unmarshal(matched[2], &viewReviewData); err != nil {
				c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
			}
		} else {
			c.logger.Error("review data not found")
		}
	}

	opts := c.CrawlOptions(resp.Request.URL)
	matched := c.productPathMatcher.FindStringSubmatch(resp.Request.URL.Path)
	req, _ := http.NewRequest(http.MethodPost, "https://www.express.com/graphql",
		bytes.NewReader([]byte(fmt.Sprintf(productGraphQLQuery, matched[2]))))

	c.logger.Infof("Access %s ProductQuery", req.URL)
	req.Header.Add("accept", "*/*")
	req.Header.Add("referer", resp.Request.URL.String())
	req.Header.Add("origin", "https://www.express.com")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("cache-control", "no-cache")
	if ua := resp.Request.Header.Get("User-Agent"); ua != "" {
		req.Header.Set("User-Agent", resp.Request.Header.Get("User-Agent"))
	}
	req.Header.Add("x-exp-request-id", uuid.NewV4().String())
	req.Header.Add("x-exp-rvn-cacheable", "false")
	req.Header.Add("x-exp-rvn-query-classification", "product")
	req.Header.Add("x-exp-rvn-source", "app_express.com")

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
	apiResp, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
		EnableProxy:       true,
		EnableHeadless:    false,
		EnableSessionInit: false,
		Reliability:       opts.Reliability,
		RequestFilterKeys: []string{"ProductQuery"},
	})
	if err != nil {
		c.logger.Errorf("do http request failed, error=%s", err)
		return err
	}
	defer apiResp.Body.Close()

	var viewData parseProductResponse
	{
		data, err := io.ReadAll(apiResp.Body)
		if err != nil {
			c.logger.Error(err)
			return err
		}
		if err := json.Unmarshal(data, &viewData); err != nil {
			c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
			return err
		}
	}

	reviewCount, _ := strconv.ParseInt(viewReviewData.AggregateRating.ReviewCount)
	rating, _ := strconv.ParseInt(viewReviewData.AggregateRating.RatingValue)

	canUrl, _ := c.CanonicalUrl(resp.Request.URL.String())
	prod := viewData.Data.Product
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           prod.ProductID,
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
		},
		BrandName:   "Express",
		Title:       viewData.Data.Product.Name,
		CrowdType:   viewData.Data.Product.Gender,
		Category:    prod.Gender,
		SubCategory: prod.BreadCrumbCategory.CategoryName,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stock: &pbItem.Stock{
			StockStatus: pbItem.Stock_OutOfStock,
		},
		Stats: &pbItem.Stats{
			ReviewCount: int32(reviewCount),
			Rating:      float32(rating),
		},
	}
	for _, desc := range prod.ProductDescription {
		if item.Description == "" {
			item.Description = strings.Join(desc.Content, "<br/>")
		} else {
			item.Description = "<br/>" + strings.Join(desc.Content, "<br/>")
		}
	}
	item.Description += "<ul><li>" + strings.Join(prod.ProductFeatures, "</li><li>") + "</li></ul>"
	if prod.ProductInventory > 0 {
		item.Stock.StockStatus = pbItem.Stock_InStock
		item.Stock.StockCount = int32(prod.ProductInventory)
	}

	for _, p := range viewData.Data.Product.ColorSlices {
		colorSpec := pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecColor,
			Id:    p.IPColorCode,
			Name:  p.Color,
			Value: p.Color,
			Icon:  p.SwatchURL,
		}

		var medias []*pbMedia.Media
		for ki, m := range p.ImageMap.All.LARGE {
			medias = append(medias, pbMedia.NewImageMedia(
				strconv.Format(ki),
				m,
				m,
				m,
				imgWidthReg.ReplaceAllString(m, "wid=480"),
				"",
				ki == 0,
			))
		}

		for _, rawSku := range p.Skus {
			originalPrice, _ := strconv.ParsePrice(rawSku.DisplayPrice)
			msrp, _ := strconv.ParsePrice(rawSku.DisplayMSRP)
			discount := 0.0
			if msrp > originalPrice {
				discount = math.Ceil((msrp - originalPrice) / msrp * 100)
			}

			sku := pbItem.Sku{
				SourceId: rawSku.SkuID,
				Title:    viewData.Data.Product.Name,
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(originalPrice * 100),
					Msrp:     int32(msrp * 100),
					Discount: int32(discount),
				},
				Medias: medias,
				Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}
			if rawSku.OnlineInventoryCount > 0 {
				sku.Stock.StockStatus = pbItem.Stock_InStock
				sku.Stock.StockCount = int32(rawSku.OnlineInventoryCount)
			}
			sku.Specs = append(sku.Specs, &colorSpec)
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    rawSku.SkuID,
				Name:  rawSku.SizeName,
				Value: rawSku.SizeName,
			})
			item.SkuItems = append(item.SkuItems, &sku)
		}
	}
	if err = yield(ctx, &item); err != nil {
		return err
	}
	return nil
}

func (c *_Crawler) NewTestRequest(ctx context.Context) []*http.Request {
	var reqs []*http.Request
	for _, u := range []string{
		"https://www.express.com/exp/search?q=Totes",
		// "https://www.express.com/womens-clothing/dresses/cat550007",
		// "https://www.express.com/clothing/women/body-contour-cropped-square-neck-cami/pro/06418402/color/Light%20Pink/",
		// "https://www.express.com/clothing/men/solid-performance-polo/pro/05047075/color/Pitch%20Black",
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
