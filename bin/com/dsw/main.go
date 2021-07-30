package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/gosimple/slug"
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

type _Crawler struct {
	httpClient http.Client

	categoryPathMatcher *regexp.Regexp
	categoryAPIMatcher  *regexp.Regexp
	productPathMatcher  *regexp.Regexp
	productAPIMatcher   *regexp.Regexp

	logger glog.Log
}

func (_ *_Crawler) New(_ *cli.Context, client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient: client,

		categoryPathMatcher: regexp.MustCompile(`^(/[a-z]+){0,2}/category(/[a-zA-Z0-9\-]+){1,6}/N-[a-zA-Z0-9]+/?$`),
		categoryAPIMatcher:  regexp.MustCompile(`^/api/v\d+/content/pages/_/N\-[a-zA-Z0-9]+$`),
		productPathMatcher:  regexp.MustCompile(`^(/[a-z]+){0,2}/product/[a-zA-Z0-9\-]+/\d+/?$`),
		productAPIMatcher:   regexp.MustCompile(`^/api/v\d+/products/\d+$`),

		logger: logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "0ba843594ce70359942295bb15691d85"
}

// Version
func (c *_Crawler) Version() int32 {
	return 1
}

// CrawlOptions
func (c *_Crawler) CrawlOptions(u *url.URL) *crawler.CrawlOptions {
	options := crawler.NewCrawlOptions()
	options.EnableHeadless = false
	options.Reliability = pbProxy.ProxyReliability_ReliabilityMedium
	options.MustCookies = append(options.MustCookies) // &http.Cookie{Name: "geocountry", Value: `US`, Path: "/"},
	// &http.Cookie{Name: "browseCountry", Value: "US", Path: "/"},
	// &http.Cookie{Name: "browseCurrency", Value: "USD", Path: "/"},
	// &http.Cookie{Name: "browseLanguage", Value: "en-US", Path: "/"},
	// &http.Cookie{Name: "browseSizeSchema", Value: "US", Path: "/"},
	// &http.Cookie{Name: "browseSizeSchema", Value: "US", Path: "/"},
	// &http.Cookie{Name: "storeCode", Value: "US", Path: "/"},

	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.dsw.com"}
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
		u.Host = "www.dsw.com"
	}

	if c.productPathMatcher.MatchString(u.Path) {
		u.RawQuery = ""
		return u.String(), nil
	}
	return u.String(), nil
}

type categoryStructure struct {
	ContentContentItem struct {
		Contents []struct {
			TopNavList []struct {
				Name         string `json:"name"`
				CategoryName string `json:"categoryName,omitempty"`
				Otws         []struct {
					Name             string `json:"name"`
					HeaderNavigation []struct {
						Name     string `json:"name"`
						LinkText struct {
							Path string `json:"path"`
						} `json:"linkText,omitempty"`
					} `json:"HeaderNavigation"`
				} `json:"otws,omitempty"`
				Shoes []struct {
					Name             string `json:"name"`
					HeaderNavigation []struct {
						Name     string `json:"name"`
						LinkText struct {
							Path string `json:"path"`
						} `json:"linkText"`
					} `json:"HeaderNavigation"`
				} `json:"shoes,omitempty"`
				Spotlight []struct {
					Name             string `json:"name"`
					HeaderNavigation []struct {
						Name     string `json:"name"`
						LinkText struct {
							Path        string `json:"path"`
							LinkType    string `json:"linkType"`
							QueryString string `json:"queryString"`
						} `json:"linkText,omitempty"`
					} `json:"HeaderNavigation"`
				} `json:"spotlight,omitempty"`
			} `json:"TopNavList"`
		} `json:"contents"`
	} `json:"contentContentItem"`
}

// GetCategories
func (c *_Crawler) GetCategories(ctx context.Context) ([]*pbItem.Category, error) {
	catUrl := "https://www.dsw.com/api/v1/content/zones?contentCollection=%2Fcontent%2FDSW%2FContents%2FSharedContents%2FHeaderContent&locale=en_US&pushSite=DSW&tier=GUEST"

	req, err := http.NewRequest(http.MethodGet, catUrl, nil)
	req.Header.Add("accept", "application/json, text/plain, */*")
	req.Header.Add("referer", "https://www.dsw.com/en/us/")
	opts := c.CrawlOptions(req.URL)
	resp, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
		EnableProxy:       true,
		EnableHeadless:    opts.EnableHeadless,
		EnableSessionInit: opts.EnableSessionInit,
		DisableCookieJar:  opts.DisableCookieJar,
	})
	if err != nil {
		c.logger.Error(err)
		return nil, err
	}

	var viewData categoryStructure
	if err := json.NewDecoder(resp.Body).Decode(&viewData); err != nil {
		c.logger.Errorf("unmarshal cat detail data fialed, error=%s", err)
		return nil, err
	}
	var cates []*pbItem.Category
	for _, rawcat := range viewData.ContentContentItem.Contents[0].TopNavList {
		cate := pbItem.Category{Name: rawcat.CategoryName}
		cates = append(cates, &cate)
		for _, rawsubcat := range rawcat.Shoes {
			subCate := pbItem.Category{Name: rawsubcat.Name}
			cate.Children = append(cate.Children, &subCate)
			for _, rawlastcat := range rawsubcat.HeaderNavigation {
				href, _ := c.CanonicalUrl(rawlastcat.LinkText.Path)
				if href == "" {
					continue
				}
				subCate2 := pbItem.Category{
					Name: rawlastcat.Name,
					Url:  href,
				}
				subCate.Children = append(subCate2.Children, &subCate2)
			}
		}

		for _, rawsubcat := range rawcat.Spotlight {
			subCate := pbItem.Category{Name: rawsubcat.Name}
			cate.Children = append(cate.Children, &subCate)
			for _, rawlastcat := range rawsubcat.HeaderNavigation {
				href, _ := c.CanonicalUrl(rawlastcat.LinkText.Path)
				if href == "" {
					continue
				}
				subCate2 := pbItem.Category{
					Name: rawlastcat.Name,
					Url:  href,
				}
				subCate.Children = append(subCate.Children, &subCate2)
			}
		}

		for _, rawsubcat := range rawcat.Otws {
			subCate := pbItem.Category{Name: rawsubcat.Name}
			cate.Children = append(cate.Children, &subCate)
			for _, rawlastcat := range rawsubcat.HeaderNavigation {
				href, _ := c.CanonicalUrl(rawlastcat.LinkText.Path)
				if href == "" {
					continue
				}
				subCate2 := pbItem.Category{
					Name: rawlastcat.Name,
					Url:  href,
				}
				subCate.Children = append(subCate.Children, &subCate2)
			}
		}
	}
	return cates, nil
}

func (c *_Crawler) Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}
	p := strings.TrimSuffix(resp.RawUrl().Path, "/")
	if p == "/en/us" {
		return crawler.ErrUnsupportedPath
	}
	if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) || c.categoryAPIMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	} else if c.productPathMatcher.MatchString(resp.Request.URL.Path) || c.productAPIMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yield)
	}
	return crawler.ErrUnsupportedPath
}

var productsExtractReg = regexp.MustCompile(`(?U)window\.INITIAL_STATE\s*=\s*({.*})\s*</script>`)

// nextIndex used to get sharingData from context
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")) + 1)
}

// parseCategoryProducts parse api url from web page url
func (c *_Crawler) parseCategoryProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	var (
		subresp *http.Response = resp
		err     error
	)
	if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		// do api request
		parts := strings.Split(strings.TrimSuffix(resp.Request.URL.Path, "/"), "/")
		rawUrl := fmt.Sprintf("https://www.dsw.com/api/v1/content/pages/_/%s?pagePath=%%2Fpages%%2FDSW%%2Fcategory&skipHeaderFooterContent=true&No=0&locale=en_US&pushSite=DSW&tier=GUEST", parts[len(parts)-1])
		u, _ := url.Parse(rawUrl)
		vals := u.Query()
		rawVals := resp.Request.URL.Query()
		if rawVals.Get("No") != "" {
			vals.Set("No", rawVals.Get("No"))
		}
		u.RawQuery = vals.Encode()
		req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
		req.Header.Set("Referer", resp.Request.URL.String())
		subresp, err = c.httpClient.DoWithOptions(ctx, req, http.Options{
			EnableProxy: true,
			Reliability: c.CrawlOptions(resp.Request.URL).Reliability,
		})
		if err != nil {
			c.logger.Error(err)
			return err
		}
		defer subresp.Body.Close()
	}

	respBody, err := io.ReadAll(subresp.Body)
	if err != nil {
		c.logger.Debug(err)
		return err
	}

	var r struct {
		PageContentItem struct {
			Contents []struct {
				MainContent []struct {
					Name     string `json:"name"`
					Contents []struct {
						LastRecNum int `json:"lastRecNum"`
						Records    []struct {
							Attributes struct {
								ProductOriginalStyleID []string `json:"product.originalStyleId"`
								ProductDisplayName     []string `json:"product.displayName"`
								Brand                  []string `json:"brand"`
							} `json:"attributes,omitempty"`
							Records []struct {
							} `json:"records,omitempty"`
						} `json:"records"`
						TotalNumRecs int `json:"totalNumRecs"`
						RecsPerPage  int `json:"recsPerPage"`
					} `json:"contents"`
				} `json:"mainContent"`
				MappedURL string `json:"mappedUrl"`
			} `json:"contents"`
		} `json:"pageContentItem"`
	}

	if err = json.Unmarshal(respBody, &r); err != nil {
		c.logger.Debugf("parse %s failed, error=%s", respBody, err)
		return err
	}

	totalrecords := 0
	nrpp := 0

	lastIndex := nextIndex(ctx)
	for _, prod := range r.PageContentItem.Contents[0].MainContent {
		if prod.Name != "ResultList Zone" {
			continue
		}

		totalrecords = int(prod.Contents[0].TotalNumRecs)
		nrpp = int(prod.Contents[0].RecsPerPage)
		//nexturl = prod.Contents[0].PagingActionTemplate.SiteState.ContentPath
		for _, result := range prod.Contents[0].Records {

			if result.Records == nil {
				continue
			} else if len(result.Attributes.ProductOriginalStyleID) == 0 {
				continue
			}

			nameSlug := slug.Make(fmt.Sprintf("%s %s",
				result.Attributes.Brand[0], result.Attributes.ProductDisplayName[0]))
			rawurl := fmt.Sprintf("https://www.dsw.com/en/us/product/%s/%v",
				nameSlug, result.Attributes.ProductOriginalStyleID[0])

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

	q := subresp.Request.URL.Query()
	lastNo, _ := strconv.ParseInt(q.Get("No"))

	if nrpp == 0 {
		nrpp = 60
	}
	// check if this is the last page
	if int(lastNo)+nrpp >= totalrecords {
		return nil
	}

	// set pagination
	u := *resp.Request.URL
	vals := u.Query()
	vals.Set("No", strconv.Format(int(lastNo)+nrpp))
	vals.Set("Nrpp", strconv.Format(nrpp))
	u.RawQuery = vals.Encode()

	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	// update the index of last page
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
}

type parseProductResponse struct {
	Response struct {
		Product struct {
			LongDescription       string `json:"longDescription"`
			DisplayCompareAtPrice bool   `json:"displayCompareAtPrice"`
			DisplayName           string `json:"displayName"`
			DswBrand              struct {
				DisplayNameDefault string `json:"displayNameDefault"`
				NavStringURL       string `json:"navStringURL"`
			} `json:"dswBrand"`
			AfterPayInstallmentPrice float64 `json:"afterPayInstallmentPrice"`
			ID                       string  `json:"id"`
			BvReviewCount            int     `json:"bvReviewCount"`
			BvRating                 float64 `json:"bvRating"`
			ShowWidth                bool    `json:"showWidth"`
			ChildSKUs                []struct {
				Color struct {
					DisplayName string `json:"displayName"`
					ColorCode   string `json:"colorCode"`
				} `json:"color"`
				OriginalPrice  float64 `json:"originalPrice"`
				NonMemberPrice float64 `json:"nonMemberPrice"`
				Upc            string  `json:"upc"`
				IsDropShipItem bool    `json:"isDropShipItem"`
				Size           struct {
					DisplayName string  `json:"displayName"`
					SizeCode    float64 `json:"sizeCode"`
				} `json:"size"`
				ID            string `json:"id"`
				SkuStockLevel int    `json:"skuStockLevel"`
			} `json:"childSKUs"`
			ProductGender     string  `json:"productGender"`
			PriceInCart       bool    `json:"priceInCart"`
			NonMemberMinPrice float64 `json:"nonMemberMinPrice"`
			Breadcrumbs       []struct {
				Text string `json:"text"`
				URL  string `json:"url"`
			} `json:"breadcrumbs"`
		} `json:"product"`
	} `json:"Response"`
}

type parseProductImageResponse struct {
	Set struct {
		Item []struct {
			I struct {
				IsDefault string `json:"isDefault"`
				N         string `json:"n"`
			} `json:"i,omitempty"`
		} `json:"item"`
	} `json:"set"`
}

var (
	imageRegStart = regexp.MustCompile(`({.*})`)
	viewData      parseProductResponse
	q             parseProductImageResponse
)

var htmlTrimRegp = regexp.MustCompile(`</?[^>]+>`)

// parseProduct
func (c *_Crawler) parseProduct(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil {
		return nil
	}

	var (
		subresp  *http.Response = resp
		err      error
		crawlUrl string
		canUrl   string
	)
	if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		fields := strings.Split(strings.TrimSuffix(resp.Request.URL.Path, "/"), "/")
		rawurl := fmt.Sprintf("https://www.dsw.com/api/v1/products/%s?locale=en_US&pushSite=DSW", fields[len(fields)-1])
		req, _ := http.NewRequest(http.MethodGet, rawurl, nil)
		req.Header.Set("accept", "application/json, text/plain, */*")
		req.Header.Set("Referer", resp.Request.URL.String())
		cookie := resp.Response.Header

		urbn_auth_payload := (strings.Join(cookie.Values(`Set-Cookie`), ";"))
		req.Header.Set("cookie", urbn_auth_payload)
		subresp, err = c.httpClient.DoWithOptions(ctx, req, http.Options{
			EnableProxy: true,
			Reliability: c.CrawlOptions(resp.Request.URL).Reliability,
		})
		if err != nil {
			c.logger.Error(err)
			return err
		}
		defer subresp.Body.Close()

		crawlUrl = resp.Request.URL.String()
	}

	respBody, err := io.ReadAll(subresp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(respBody, &viewData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}
	req_product_id := strconv.Format(viewData.Response.Product.ID)

	if crawlUrl == "" {
		// the slug may not the same as the org slug
		canUrl = fmt.Sprintf("https://www.dsw.com/en/us/product/%s/%v",
			slug.Make(fmt.Sprintf("%s %s", viewData.Response.Product.DswBrand.DisplayNameDefault, viewData.Response.Product.DisplayName)),
			viewData.Response.Product.ID,
		)
		crawlUrl = canUrl
	} else {
		canUrl, _ = c.CanonicalUrl(crawlUrl)
	}

	desc, _ := url.QueryUnescape(viewData.Response.Product.LongDescription)
	//Prepare product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           strconv.Format(viewData.Response.Product.ID),
			CrawlUrl:     crawlUrl,
			CanonicalUrl: canUrl,
		},
		BrandName:   viewData.Response.Product.DswBrand.DisplayNameDefault,
		Title:       viewData.Response.Product.DisplayName,
		Description: htmlTrimRegp.ReplaceAllString(desc, " "),
		CrowdType:   strings.ToLower(viewData.Response.Product.ProductGender),
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stats: &pbItem.Stats{
			ReviewCount: int32(viewData.Response.Product.BvReviewCount),
			Rating:      float32(viewData.Response.Product.BvRating),
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
	}
	for i, cate := range viewData.Response.Product.Breadcrumbs {
		switch i {
		case 0:
			item.Category = strings.TrimSpace(strings.TrimPrefix(cate.Text, viewData.Response.Product.ProductGender))
		case 1:
			item.SubCategory = cate.Text
		case 2:
			item.SubCategory2 = cate.Text
		}
	}

	mediasDict := map[string][]*pbMedia.Media{}
	for _, rawSku := range viewData.Response.Product.ChildSKUs {
		originalPrice, _ := strconv.ParseFloat(rawSku.OriginalPrice)
		currentPrice, _ := strconv.ParseFloat(rawSku.NonMemberPrice)
		discount := math.Ceil((originalPrice - currentPrice) * 100.0 / originalPrice)

		sku := pbItem.Sku{
			SourceId: strconv.Format(rawSku.Upc),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(currentPrice * 100),
				Msrp:     int32(originalPrice * 100),
				Discount: int32(discount),
			},
			Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}
		if rawSku.SkuStockLevel > 0 {
			sku.Stock.StockStatus = pbItem.Stock_InStock
			sku.Stock.StockCount = int32(rawSku.SkuStockLevel)
		}

		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecColor,
			Id:    strconv.Format(rawSku.Color.ColorCode),
			Name:  rawSku.Color.DisplayName,
			Value: rawSku.Color.ColorCode,
			Icon: fmt.Sprintf("https://images.dsw.com/is/image/DSWShoes/%v_%s_sw?$slswatches$",
				viewData.Response.Product.ID, rawSku.Color.ColorCode),
		})

		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Id:    strconv.Format(rawSku.Size.SizeCode),
			Name:  rawSku.Size.DisplayName,
			Value: strconv.Format(rawSku.Size.SizeCode),
		})

		web_Product_ID := rawSku.Color.ColorCode
		if medias, ok := mediasDict[web_Product_ID]; ok {
			sku.Medias = medias
		} else {
			imgrequest := "https://images.dsw.com/is/image/DSWShoes/?imageset={id}_{code}_ss_01,{id}_{code}_ss_02,{id}_{code}_ss_03,{id}_{code}_ss_04,{id}_{code}_ss_05,{id}_{code}_ss_06,{id}_{code}_ss_07,{id}_{code}_ss_08,{id}_{code}_ss_09,{id}_{code}_ss_010&req=set,json&handler=ng_jsonp_callback_0"
			imgUrl := strings.ReplaceAll((strings.ReplaceAll(imgrequest, "{id}", req_product_id)), "{code}", web_Product_ID)
			req, err := http.NewRequest(http.MethodGet, imgUrl, nil)
			req.Header.Set("Referer", resp.Request.URL.String())

			imgreq, err := c.httpClient.Do(ctx, req)
			if err != nil {
				panic(err)
			}
			defer imgreq.Body.Close()

			respBodyImg, err := io.ReadAll(imgreq.Body)
			if err != nil {
				c.logger.Error(err)
				return err
			}

			matched := imageRegStart.FindSubmatch(respBodyImg)
			if len(matched) <= 1 {
				c.logger.Debugf("data %s", respBodyImg)
				return fmt.Errorf("extract produt json from page %s content failed", resp.Request.URL)
			}

			if err = json.Unmarshal(matched[0], &q); err != nil {
				c.logger.Debugf("parse %s failed, error=%s", matched[2], err)
				return err
			}

			isDefault := true
			for key, img := range q.Set.Item {
				if key > 0 {
					isDefault = false
				}
				if strings.Contains(img.I.N, "Image_Not") || strings.Contains(img.I.N, "_video") {
					continue
				}
				imgURLDefault := "https://images.dsw.com/is/image/" + img.I.N

				sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
					strconv.Format(req_product_id+web_Product_ID),
					imgURLDefault,
					imgURLDefault+"?scl=1.4&qlt=70&fmt=jpeg&wid=1000&hei=1200&op_sharpen=1",
					imgURLDefault+"?scl=2.1&qlt=70&fmt=jpeg&wid=690&hei=810&op_sharpen=1",
					imgURLDefault+"?scl=2.45&qlt=70&fmt=jpeg&wid=590&hei=700&op_sharpen=1",
					"",
					isDefault,
				))
			}
		}
		mediasDict[web_Product_ID] = sku.Medias

		item.SkuItems = append(item.SkuItems, &sku)
		item.Medias = append(item.Medias, sku.Medias...)
	}

	// yield item result
	if err = yield(ctx, &item); err != nil {
		return err
	}

	return nil
}

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://www.dsw.com/en/us",
		//"https://www.dsw.com/en/us/product/aston-grey-leu-oxford/386780?activeColor=240",
		"https://www.dsw.com/en/us/product/birkenstock-cotton-slub-womens-crew-socks/497778?activeColor=050",
		//"https://www.dsw.com/en/us/category/womens-socks/N-1z141jrZ1z128ueZ1z141dn?No=0",
		//"https://www.dsw.com/api/v1/content/pages/_/N-1z141hwZ1z128ujZ1z141ju?pagePath=/pages/DSW/category&skipHeaderFooterContent=true&No=0&locale=en_US&pushSite=DSW&tier=GUEST",
		//"https://www.dsw.com/api/v1/products/499002?locale=en_US&pushSite=DSW",
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
	cli.NewApp(&_Crawler{}).Run(os.Args)
}
