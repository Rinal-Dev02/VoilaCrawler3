package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/voiladev/VoilaCrawl/pkg/crawler"
	"github.com/voiladev/VoilaCrawl/pkg/net/http"
	"github.com/voiladev/VoilaCrawl/pkg/net/http/cookiejar"
	"github.com/voiladev/VoilaCrawl/pkg/proxy"
	pbProxy "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type _Crawler struct {
	httpClient http.Client

	categoryPathMatcher *regexp.Regexp
	categoryAPIMatcher  *regexp.Regexp
	productPathMatcher  *regexp.Regexp
	logger              glog.Log
}

func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient:          client,
		categoryPathMatcher: regexp.MustCompile(`^(/[a-z0-9_\-]+){1,4}/cat(\d)+(/[a-z0-9_\-]+){0,4}$`),
		categoryAPIMatcher:  regexp.MustCompile(`^/cic/browse/v\d+$`),
		productPathMatcher:  regexp.MustCompile(`^/(u|t)(/[a-zA-Z0-9_\-]+){2,4}$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "e37a370f2a444978a1f8fb4dd0d8008c"
}

// Version
func (c *_Crawler) Version() int32 {
	return 1
}

// CrawlOptions
func (c *_Crawler) CrawlOptions(u *url.URL) *crawler.CrawlOptions {
	options := crawler.NewCrawlOptions()
	options.EnableHeadless = false
	options.LoginRequired = false
	options.Reliability = pbProxy.ProxyReliability_ReliabilityMedium
	options.MustCookies = append(options.MustCookies,
		&http.Cookie{Name: "geoloc", Value: "cc=US,rc=CA,tp=vhigh,tz=PST,la=33.9733,lo=-118.2487"},
	)
	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.express.com"}
}

func (c *_Crawler) Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	} else if c.categoryAPIMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryAPIProducts(ctx, resp, yield)
	} else if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yield)
	}
	return fmt.Errorf("unsupported url %s", resp.Request.URL.String())
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
			Products []struct {
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
			} `json:"products"`
			Source   string `json:"source"`
			Typename string `json:"__typename"`
		} `json:"getUnbxdCategory"`
	} `json:"data"`
	Extensions struct {
		Platform string `json:"platform"`
	} `json:"extensions"`
}

var productsExtractReg = regexp.MustCompile(`window\.INITIAL_REDUX_STATE=({.*});?\s*</script>`)

// parseCategoryProducts parse api url from web page url
func (c *_Crawler) parseCategoryProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	opts := c.CrawlOptions(resp.Request.URL)
	// read the response data from http response
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	ioutil.WriteFile("C:\\Rinal\\ServiceBasedPRojects\\VoilaWork_new\\VoilaCrawl\\Output.html", respBody, 0644)
	matched := productsExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respBody)
		//return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}

	rawurl := "https://www.express.com/graphql"
	postSend := "{\"operationName\":\"CategoryQuery\",\"variables\":{\"categoryId\":\"cat550007\",\"filter\":\"\",\"overrideCatApi\":\"\",\"rows\":56,\"sort\":\"\",\"start\":0},\"query\":\"query CategoryQuery($categoryId: String, $start: Int!, $rows: Int, $filter: String, $sort: String, $overrideCatApi: String, $uc_param: String) {\n  getUnbxdCategory(categoryId: $categoryId, start: $start, rows: $rows, filter: $filter, sort: $sort, overrideCatApi: $overrideCatApi, uc_param: $uc_param) {\n    categoryId\n    categoryName\n    facets {\n      facetId\n      name\n      position\n      values\n      __typename\n    }\n    pagination {\n      totalProductCount\n      pageNumber\n      pageSize\n      start\n      end\n      __typename\n    }\n    products {\n      averageOverallRating\n      colors {\n        color\n        skuUpc\n        defaultSku\n        __typename\n      }\n      EFOProduct\n      ensembleListPrice\n      ensembleSalePrice\n      key\n      isEnsemble\n      listPrice\n      marketplaceProduct\n      name\n      newProduct\n      onlineExclusive\n      onlineExclusivePromoMsg\n      paginationEnd\n      paginationStart\n      productDescription\n      productId\n      productImage\n      productURL\n      promoMessage\n      salePrice\n      totalReviewCount\n      __typename\n    }\n    source\n    __typename\n  }\n}\n\"}"

	req, _ := http.NewRequest(http.MethodPost, rawurl, bytes.NewReader([]byte(postSend)))
	req.Header.Set("Referer", resp.Request.URL.String())
	if resp.Request.Header.Get("User-Agent") != "" {
		req.Header.Set("User-Agent", resp.Request.Header.Get("User-Agent"))
		req.Header.Set("accept", "*/*")
		req.Header.Set("content-type", "application/json")
	}
	c.logger.Debugf("Access images %s", rawurl)
	respNew, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
		EnableProxy:       true,
		EnableHeadless:    opts.EnableHeadless,
		EnableSessionInit: opts.EnableSessionInit,
		Reliability:       opts.Reliability,
	})

	respBodyNew, err := ioutil.ReadAll(respNew.Body)
	if err != nil {
		return err
	}

	ioutil.WriteFile("C:\\Rinal\\ServiceBasedPRojects\\VoilaWork_new\\VoilaCrawl\\Output_1.html", respBodyNew, 0644)

	// c.logger.Debugf("data: %s", matched[1])

	var viewData CategoryView
	if err := json.Unmarshal(respBodyNew, &viewData); err != nil {
		c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		return err
	}

	lastIndex := nextIndex(ctx)
	for _, idv := range viewData.Data.GetUnbxdCategory.Products {

		rawurl := idv.ProductURL
		fmt.Println(rawurl)
		// req, err := http.NewRequest(http.MethodGet, rawurl, nil)
		// if err != nil {
		// 	c.logger.Errorf("load http request of url %s failed, error=%s", rawurl, err)
		// 	return err
		// }

		lastIndex += 1
		// // set the index of the product crawled in the sub response
		// nctx := context.WithValue(ctx, "item.index", lastIndex)
		// // yield sub request
		// if err := yield(nctx, req); err != nil {
		// 	return err
		// }
	}

	if viewData.Data.GetUnbxdCategory.Pagination.TotalProductCount <= lastIndex {

		postSend = "{\"operationName\":\"CategoryQuery\",\"variables\":{\"categoryId\":\"" + strconv.Format(viewData.Data.GetUnbxdCategory.CategoryID) + "\",\"filter\":\"\",\"overrideCatApi\":\"\",\"rows\":56,\"sort\":\"\",\"start\":" + strconv.Format(viewData.Data.GetUnbxdCategory.Pagination.End+1) + "},\"query\":\"query CategoryQuery($categoryId: String, $start: Int!, $rows: Int, $filter: String, $sort: String, $overrideCatApi: String, $uc_param: String) {\n  getUnbxdCategory(categoryId: $categoryId, start: $start, rows: $rows, filter: $filter, sort: $sort, overrideCatApi: $overrideCatApi, uc_param: $uc_param) {\n    categoryId\n    categoryName\n    facets {\n      facetId\n      name\n      position\n      values\n      __typename\n    }\n    pagination {\n      totalProductCount\n      pageNumber\n      pageSize\n      start\n      end\n      __typename\n    }\n    products {\n      averageOverallRating\n      colors {\n        color\n        skuUpc\n        defaultSku\n        __typename\n      }\n      EFOProduct\n      ensembleListPrice\n      ensembleSalePrice\n      key\n      isEnsemble\n      listPrice\n      marketplaceProduct\n      name\n      newProduct\n      onlineExclusive\n      onlineExclusivePromoMsg\n      paginationEnd\n      paginationStart\n      productDescription\n      productId\n      productImage\n      productURL\n      promoMessage\n      salePrice\n      totalReviewCount\n      __typename\n    }\n    source\n    __typename\n  }\n}\n\"}"
		req, _ := http.NewRequest(http.MethodPost, rawurl, bytes.NewReader([]byte(postSend)))
		// update the index of last page
		nctx := context.WithValue(ctx, "item.index", lastIndex)
		return yield(nctx, req)
	}
	return nil
}

func (c *_Crawler) parseCategoryAPIProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil {
		return nil
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var viewData struct {
		Data struct {
			Products struct {
				Products []struct {
					URL string `json:"url"`
				} `json:"products"`
				Pages struct {
					Prev           string      `json:"prev"`
					Next           string      `json:"next"`
					TotalPages     int         `json:"totalPages"`
					TotalResources int         `json:"totalResources"`
					SearchSummary  interface{} `json:"searchSummary"`
				} `json:"pages"`
				ExternalResponses interface{} `json:"externalResponses"`
				TraceID           string      `json:"traceId"`
			} `json:"products"`
		} `json:"data"`
	}

	if err := json.Unmarshal([]byte(respBody), &viewData); err != nil {
		c.logger.Error(err)
		return err
	}

	lastIndex := nextIndex(ctx)
	for _, prod := range viewData.Data.Products.Products {
		u := "https://www.nike.com" + strings.ReplaceAll(prod.URL, "{countryLang}", "")

		req, err := http.NewRequest(http.MethodGet, u, nil)
		if err != nil {
			c.logger.Error(err)
			return err
		}
		nctx := context.WithValue(ctx, "item.index", lastIndex)
		lastIndex += 1

		if err := yield(nctx, req); err != nil {
			return err
		}
	}

	if viewData.Data.Products.Pages.Next != "" {
		u := *resp.Request.URL
		vals := u.Query()
		vals.Set("endpoint", viewData.Data.Products.Pages.Next)

		u.RawQuery = vals.Encode()

		req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
		// update the index of last page
		nctx := context.WithValue(ctx, "item.index", lastIndex)
		return yield(nctx, req)
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

func (c *_Crawler) parseProduct(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}
	if resp.StatusCode == http.StatusForbidden {
		return errors.New("access denied")
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

	// var item *pbItem.Product
	// for _, p := range viewData.Threads.Products {
	// 	if item == nil {
	// 		item = &pbItem.Product{
	// 			Source: &pbItem.Source{
	// 				Id:       p.ProductGroupID,
	// 				CrawlUrl: resp.Request.URL.String(),
	// 			},
	// 			BrandName:   p.Brand,
	// 			Title:       p.FullTitle,
	// 			Description: p.Description,
	// 			CrowdType:   strings.ToLower(strings.Join(p.Genders, ",")),
	// 			Price: &pbItem.Price{
	// 				Currency: regulation.Currency_USD,
	// 			},
	// 			Stats: &pbItem.Stats{
	// 				ReviewCount: int32(viewData.Reviews.Total),
	// 				Rating:      float32(viewData.Reviews.AverageRating),
	// 			},
	// 		}
	// 	}

	// 	colorSpec := pbItem.SkuSpecOption{
	// 		Type:  pbItem.SkuSpecType_SkuSpecColor,
	// 		Id:    p.StyleColor,
	// 		Name:  p.ColorDescription,
	// 		Value: p.StyleColor,
	// 	}

	// 	var medias []*pbMedia.Media
	// 	for ki, m := range p.Nodes[0].Nodes {
	// 		template := strings.ReplaceAll(m.Properties.Squarish.URL, "t_default", "t_PDP_864_v1")
	// 		medias = append(medias, pbMedia.NewImageMedia(
	// 			strconv.Format(m.ID),
	// 			strings.ReplaceAll(m.Properties.Squarish.URL, "t_default", "t_PDP_1280_v1"),
	// 			strings.ReplaceAll(m.Properties.Squarish.URL, "t_default", "t_PDP_1280_v1"),
	// 			template,
	// 			template,
	// 			"",
	// 			ki == 0,
	// 		))
	// 	}

	// 	for _, rawSku := range p.Skus {
	// 		originalPrice, _ := strconv.ParseFloat(p.FullPrice)
	// 		msrp, _ := strconv.ParseFloat(p.FullPrice)
	// 		discount := math.Ceil((msrp - originalPrice) / msrp * 100)

	// 		sku := pbItem.Sku{
	// 			SourceId:    rawSku.ID,
	// 			Title:       p.FullTitle,
	// 			Description: p.Description,
	// 			Price: &pbItem.Price{
	// 				Currency: regulation.Currency_USD,
	// 				Current:  int32(originalPrice * 100),
	// 				Msrp:     int32(msrp * 100),
	// 				Discount: int32(discount),
	// 			},
	// 			Medias: medias,
	// 			Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
	// 		}
	// 		if p.State == "IN_STOCK" {
	// 			sku.Stock.StockStatus = pbItem.Stock_InStock
	// 		}
	// 		sku.Specs = append(sku.Specs, &colorSpec)

	// 		// size
	// 		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
	// 			Type:  pbItem.SkuSpecType_SkuSpecSize,
	// 			Id:    rawSku.NikeSize,
	// 			Name:  rawSku.LocalizedSize,
	// 			Value: rawSku.LocalizedSize,
	// 		})
	// 		item.SkuItems = append(item.SkuItems, &sku)
	// 	}
	// }
	// if item != nil {
	// 	return yield(ctx, item)
	// }
	return errors.New("no product found")
}

func (c *_Crawler) NewTestRequest(ctx context.Context) []*http.Request {
	var reqs []*http.Request
	for _, u := range []string{
		"https://www.express.com/womens-clothing/dresses/cat550007",
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
	logger := glog.New(glog.LogLevelDebug)
	// build a http client
	// get proxy's microservice address from env
	os.Setenv("VOILA_PROXY_URL", "http://52.207.171.114:30216")
	client, err := proxy.NewProxyClient(os.Getenv("VOILA_PROXY_URL"), cookiejar.New(), logger)
	if err != nil {
		panic(err)
	}

	// instance the spider locally
	spider, err := New(client, logger)
	if err != nil {
		panic(err)
	}

	// this callback func is used to do recursion call of sub requests.
	var callback func(ctx context.Context, val interface{}) error
	callback = func(ctx context.Context, val interface{}) error {
		switch i := val.(type) {
		case *http.Request:
			logger.Debugf("Access %s", i.URL)
			// crawler := spider.(*_Crawler)
			// if crawler.productPathMatcher.MatchString(i.URL.Path) {
			// 	return nil
			// }

			opts := spider.CrawlOptions(i.URL)

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
				i.URL.Host = "www.express.com"
			}

			// do http requests here.
			nctx, cancel := context.WithTimeout(ctx, time.Minute*5)
			defer cancel()
			resp, err := client.DoWithOptions(nctx, i, http.Options{
				EnableProxy:       true,
				EnableHeadless:    false,
				EnableSessionInit: opts.EnableSessionInit,
				KeepSession:       opts.KeepSession,
				Reliability:       opts.Reliability,
			})
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			return spider.Parse(ctx, resp, callback)
		default:
			// output the result
			data, err := protojson.Marshal(i.(proto.Message))
			if err != nil {
				return err
			}
			logger.Infof("data: %s", data)
		}
		return nil
	}

	ctx := context.WithValue(context.Background(), "tracing_id", fmt.Sprintf("express_%d", time.Now().UnixNano()))
	// start the crawl request
	for _, req := range spider.NewTestRequest(context.Background()) {
		if err := callback(ctx, req); err != nil {
			logger.Fatal(err)
		}
	}
}
