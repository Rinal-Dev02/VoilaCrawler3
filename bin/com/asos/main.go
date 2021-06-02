package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gosimple/slug"
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
	"google.golang.org/protobuf/types/known/anypb"
)

type _Crawler struct {
	httpClient http.Client

	categoryPathMatcher     *regexp.Regexp
	categoryJsonPathMatcher *regexp.Regexp
	productGroupPathMatcher *regexp.Regexp
	productPathMatcher      *regexp.Regexp
	logger                  glog.Log
}

func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient:              client,
		categoryPathMatcher:     regexp.MustCompile(`^(/[a-z0-9_-]+)?/(women|men)(/[a-z0-9_-]+){1,6}/cat/?$`),
		categoryJsonPathMatcher: regexp.MustCompile(`^/api/product/search/v2/categories/([a-z0-9]+)`),
		productGroupPathMatcher: regexp.MustCompile(`^(/[a-z0-9_-]+)?(/[a-z0-9_-]+){2}/grp/[0-9]+/?$`),
		productPathMatcher:      regexp.MustCompile(`^(/[a-z0-9_-]+)?(/[a-z0-9_-]+){2}/prd/[0-9]+/?$`),
		logger:                  logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "701fdaa85a5a18866ccbb357ad2ccff9"
}

// Version
func (c *_Crawler) Version() int32 {
	return 1
}

// CrawlOptions
func (c *_Crawler) CrawlOptions(u *url.URL) *crawler.CrawlOptions {
	options := crawler.NewCrawlOptions()
	options.EnableHeadless = true
	options.EnableSessionInit = true
	options.DisableCookieJar = true
	options.Reliability = pbProxy.ProxyReliability_ReliabilityMedium
	options.MustCookies = append(options.MustCookies,
		&http.Cookie{Name: "geocountry", Value: `US`, Path: "/"},
		// &http.Cookie{Name: "browseCountry", Value: "US", Path: "/"},
		// &http.Cookie{Name: "browseCurrency", Value: "USD", Path: "/"},
		// &http.Cookie{Name: "browseLanguage", Value: "en-US", Path: "/"},
		// &http.Cookie{Name: "browseSizeSchema", Value: "US", Path: "/"},
		// &http.Cookie{Name: "browseSizeSchema", Value: "US", Path: "/"},
		// &http.Cookie{Name: "storeCode", Value: "US", Path: "/"},
		// &http.Cookie{Name: "currency", Value: "2", Path: "/"},
	)

	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.asos.com"}
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
		u.Host = "www.asos.com"
	}
	if c.productPathMatcher.MatchString(u.Path) || c.productGroupPathMatcher.MatchString(u.Path) {
		u.RawQuery = ""
		return u.String(), nil
	}
	return rawurl, nil
}

func (c *_Crawler) Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}
	p := strings.TrimSuffix(resp.Request.URL.Path, "/")
	if p == "" {
		return crawler.ErrUnsupportedPath
	}
	if p == "/us/women" || p == "/us/men" {
		return c.parseCategories(ctx, resp, yield)
	}

	if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	} else if c.categoryJsonPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProductsJson(ctx, resp, yield)
	} else if c.productGroupPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProductGroup(ctx, resp, yield)
	} else if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yield)
	}
	return crawler.ErrUnsupportedPath
}

func (c *_Crawler) parseCategories(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	dom, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		c.logger.Error(err)
		return err
	}

	sel := dom.Find(`#chrome-sticky-header nav li>a[href]`)
	for i := range sel.Nodes {
		node := sel.Eq(i)
		href := node.AttrOr("href", "")
		if href == "" {
			continue
		}
		u, err := url.Parse(href)
		if err != nil {
			c.logger.Errorf("parse url %s failed", href)
			continue
		}
		if strings.Contains(u.Path, "/us/gift-vouchers") {
			continue
		}

		if c.categoryPathMatcher.MatchString(u.Path) {
			// here reset tracing id to distiguish different category crawl
			// This may exists duplicate requests
			nctx := context.WithValue(ctx, crawler.TracingIdKey, randutil.MustNewRandomID())
			req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
			if err := yield(nctx, req); err != nil {
				return err
			}
		}
	}
	return nil
}

// nextIndex used to get sharingData from context
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")) + 1)
}

var prodDataExtraReg = regexp.MustCompile(`(?U)window\.asos\.plp\._data\s*=\s*JSON\.parse\('(.*)'\);`)

// parseCategoryProducts parse api url from web page url
func (c *_Crawler) parseCategoryProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Debug(err)
		return err
	}

	matched := prodDataExtraReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract json from product list page %s failed", resp.Request.URL)
	}
	var r struct {
		Search struct {
			Products []struct {
				Id  int    `json:"id"`
				Url string `json:"url"`
			} `json:"products"`
			Query map[string]interface{} `json:"query"`
		} `json:"search"`
	}

	matched[1] = bytes.ReplaceAll(bytes.ReplaceAll(matched[1], []byte("\\'"), []byte("'")), []byte(`\\"`), []byte(`\"`))
	// rawData, err := strconv.Unquote(string(matched[1]))
	//if err != nil {
	//	c.logger.Errorf("unquote raw string failed, error=%s", err)
	//	return err
	//}
	if err = json.Unmarshal(matched[1], &r); err != nil {
		c.logger.Debugf("parse %s failed, error=%s", matched[1], err)
		return err
	}

	cid := r.Search.Query["cid"].(string)
	nctx := context.WithValue(ctx, "cid", cid)
	lastIndex := nextIndex(ctx)
	for _, prod := range r.Search.Products {
		rawurl := fmt.Sprintf("%s://%s/us/%s&cid=%s", resp.Request.URL.Scheme, resp.Request.URL.Host, prod.Url, cid)
		if strings.HasPrefix(prod.Url, "http:") || strings.HasPrefix(prod.Url, "https:") {
			rawurl = prod.Url
		}

		req, err := http.NewRequest(http.MethodGet, rawurl, nil)
		if err != nil {
			c.logger.Debug(err)
			return err
		}
		nnctx := context.WithValue(nctx, "item.index", lastIndex+1)
		lastIndex += 1
		if err = yield(nnctx, req); err != nil {
			return err
		}
	}

	u := *resp.Request.URL
	u.Path = fmt.Sprintf("/api/product/search/v2/categories/%s", cid)
	vals := url.Values{}
	for key, val := range r.Search.Query {
		if key == "cid" || key == "page" {
			continue
		}
		vals.Set(key, fmt.Sprintf("%v", val))
	}
	vals.Set("offset", strconv.Format(len(r.Search.Products)))
	vals.Set("limit", strconv.Format(len(r.Search.Products)))
	u.RawQuery = vals.Encode()

	nctx = context.WithValue(nctx, "referer", resp.Request.URL.String())
	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	return yield(context.WithValue(nctx, "item.index", lastIndex), req)
}

// parseCategoryProductsJson
func (c *_Crawler) parseCategoryProductsJson(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil {
		return nil
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var r struct {
		ItemCount int `json:"itemCount"`
		Products  []struct {
			Id  int32  `json:"id"`
			Url string `json:"url"`
		} `json:"products"`
	}
	if err := json.Unmarshal(respBody, &r); err != nil {
		c.logger.Debugf("decode %s failed, error=%s", respBody, err)
		return err
	}
	if len(r.Products) == 0 {
		return fmt.Errorf("no response from %s", resp.Request.URL)
	}
	pathes := strings.Split(resp.Request.URL.Path, "/")

	var (
		lastIndex = nextIndex(ctx)
		cid       = pathes[len(pathes)-1]
	)
	for _, prod := range r.Products {
		rawurl := fmt.Sprintf("%s://%s/us/%s&cid=%s", resp.Request.URL.Scheme, resp.Request.URL.Host, prod.Url, cid)
		if strings.HasPrefix(prod.Url, "http:") || strings.HasPrefix(prod.Url, "https:") {
			rawurl = prod.Url
		}

		if req, err := http.NewRequest(http.MethodGet, rawurl, nil); err != nil {
			return err
		} else {
			req.Header.Set("Referer", resp.Request.Header.Get("Referer"))
			nctx := context.WithValue(ctx, "item.index", lastIndex+1)
			lastIndex += 1
			if err = yield(nctx, req); err != nil {
				return err
			}
		}
	}
	offset := strconv.MustParseInt(resp.Request.URL.Query().Get("offset"))
	limit := strconv.MustParseInt(resp.Request.URL.Query().Get("limit"))
	if offset <= 0 || offset+int64(len(r.Products)) > int64(r.ItemCount) || int64(len(r.Products)) < limit {
		return nil
	}

	u := *resp.Request.URL
	vals := u.Query()
	vals.Set("offset", strconv.Format(offset+int64(len(r.Products))))
	u.RawQuery = vals.Encode()

	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	if val := ctx.Value("referer"); val != nil {
		if referer, _ := val.(string); referer != "" {
			req.Header.Set("Referer", referer)
		}
	}
	if req.Header.Get("Referer") == "" && resp.Request.Header.Get("Referer") != "" {
		req.Header.Set("Referer", resp.Request.Header.Get("Referer"))
	}
	return yield(context.WithValue(ctx, "item.index", lastIndex), req)
}

type parseProductGroupResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	BrandName string `json:"brandName"`
	Products  []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"products"`
}

// parseProductGroup parse every item
func (c *_Crawler) parseProductGroup(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error(err)
		return err
	}

	matched := detailReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		return fmt.Errorf("products not found for url %s", resp.Request.URL)
	}

	var i parseProductGroupResponse
	if err := json.Unmarshal(matched[1], &i); err != nil {
		return fmt.Errorf("extract products for response of url %s failed, error=%s", resp.Request.URL, err)
	}

	nctx := context.WithValue(ctx, "groupId", strconv.Format(i.ID))
	for _, prod := range i.Products {
		path := fmt.Sprintf("/%s/%s/prd/%v", slug.Make(i.BrandName), slug.Make(prod.Name), prod.ID)
		u := *resp.Request.URL
		u.Path = path
		if strings.HasPrefix(u.Path, "/us") {
			u.Path = "/us" + path
		}
		u.Fragment = ""

		if req, err := http.NewRequest(http.MethodGet, u.String(), nil); err != nil {
			return err
		} else if err = yield(nctx, req); err != nil {
			return err
		}
	}
	return nil
}

type parseProductResponse struct {
	ProductCode               string `json:"productCode"`
	Name                      string `json:"name"`
	Gender                    string `json:"gender"`
	ID                        int    `json:"id"`
	IsNoSize                  bool   `json:"isNoSize"`
	IsOneSize                 bool   `json:"isOneSize"`
	IsInStock                 bool   `json:"isInStock"`
	IsDeadProduct             bool   `json:"isDeadProduct"`
	PdpLayout                 string `json:"pdpLayout"`
	HasVariantsWithProp65Risk bool   `json:"hasVariantsWithProp65Risk"`
	BrandName                 string `json:"brandName"`
	Variants                  []struct {
		VariantID   int    `json:"variantId"`
		Size        string `json:"size"`
		SizeID      int    `json:"sizeId"`
		Colour      string `json:"colour"`
		ColourWayID int    `json:"colourWayId"`
		IsPrimary   bool   `json:"isPrimary"`
		SizeOrder   int    `json:"sizeOrder"`
	} `json:"variants"`
	Images []struct {
		IsPrimary     bool   `json:"isPrimary"`
		Colour        string `json:"colour"`
		ColourWayID   int    `json:"colourWayId"`
		ImageType     string `json:"imageType"`
		URL           string `json:"url"`
		ProductID     int    `json:"productId"`
		AlternateText string `json:"alternateText"`
		IsVisible     bool   `json:"isVisible"`
	} `json:"images"`
	TotalNumberOfColours int `json:"totalNumberOfColours"`
	Media                struct {
		CatwalkURL    string `json:"catwalkUrl"`
		ThreeSixtyURL string `json:"threeSixtyUrl"`
	} `json:"media"`
	BuyTheLookID         int    `json:"buyTheLookId"`
	SizeGuideVisible     bool   `json:"sizeGuideVisible"`
	SizeGuide            string `json:"sizeGuide"`
	ShippingRestrictions struct {
		ShippingRestrictionsLabel   interface{} `json:"shippingRestrictionsLabel"`
		ShippingRestrictionsVisible bool        `json:"shippingRestrictionsVisible"`
	} `json:"shippingRestrictions"`
	SellingFast       bool `json:"sellingFast"`
	PaymentPromotions struct {
		KlarnaPI4 struct {
			Us struct {
				Usd struct {
					MinimumTransactionAmount int `json:"minimumTransactionAmount"`
					MaximumTransactionAmount int `json:"maximumTransactionAmount"`
				} `json:"usd"`
			} `json:"us"`
		} `json:"klarnaPI4"`
	} `json:"paymentPromotions"`
	HasPaymentPromotionAvailable bool `json:"hasPaymentPromotionAvailable"`
}

type productVariant struct {
	ID                   int         `json:"id"`
	VariantID            int         `json:"variantId"`
	Sku                  string      `json:"sku"`
	IsInStock            bool        `json:"isInStock"`
	IsLowInStock         bool        `json:"isLowInStock"`
	StockLastUpdatedDate time.Time   `json:"stockLastUpdatedDate"`
	Warehouse            interface{} `json:"warehouse"`
	Source               interface{} `json:"source"`
	Price                struct {
		Current struct {
			Value        float64 `json:"value"`
			Text         string  `json:"text"`
			VersionID    string  `json:"versionId"`
			ConversionID string  `json:"conversionId"`
		} `json:"current"`
		Previous struct {
			Value        float64 `json:"value"`
			Text         string  `json:"text"`
			VersionID    string  `json:"versionId"`
			ConversionID string  `json:"conversionId"`
		} `json:"previous"`
		Rrp struct {
			Value        interface{} `json:"value"`
			Text         interface{} `json:"text"`
			VersionID    string      `json:"versionId"`
			ConversionID string      `json:"conversionId"`
		} `json:"rrp"`
		Xrp struct {
			Value        float64 `json:"value"`
			Text         string  `json:"text"`
			VersionID    string  `json:"versionId"`
			ConversionID string  `json:"conversionId"`
		} `json:"xrp"`
		Currency      string    `json:"currency"`
		IsMarkedDown  bool      `json:"isMarkedDown"`
		IsOutletPrice bool      `json:"isOutletPrice"`
		StartDateTime time.Time `json:"startDateTime"`
	} `json:"price"`
}

type parseProductStockPrice struct {
	ProductID    int    `json:"productId"`
	ProductCode  string `json:"productCode"`
	ProductPrice struct {
		Current struct {
			Value        float64 `json:"value"`
			Text         string  `json:"text"`
			VersionID    string  `json:"versionId"`
			ConversionID string  `json:"conversionId"`
		} `json:"current"`
		Previous struct {
			Value        float64 `json:"value"`
			Text         string  `json:"text"`
			VersionID    string  `json:"versionId"`
			ConversionID string  `json:"conversionId"`
		} `json:"previous"`
		Rrp struct {
			Value        float64 `json:"value"`
			Text         string  `json:"text"`
			VersionID    string  `json:"versionId"`
			ConversionID string  `json:"conversionId"`
		} `json:"rrp"`
		Xrp struct {
			Value        float64 `json:"value"`
			Text         string  `json:"text"`
			VersionID    string  `json:"versionId"`
			ConversionID string  `json:"conversionId"`
		} `json:"xrp"`
		Currency      string    `json:"currency"`
		IsMarkedDown  bool      `json:"isMarkedDown"`
		IsOutletPrice bool      `json:"isOutletPrice"`
		StartDateTime time.Time `json:"startDateTime"`
	} `json:"productPrice"`
	Variants []*productVariant `json:"variants"`
}

type parseProductRatingResponse struct {
	TotalReviewCount         int     `json:"totalReviewCount"`
	AverageOverallRating     float64 `json:"averageOverallRating"`
	AverageOverallStarRating float64 `json:"averageOverallStarRating"`
	DisplayRatingsSection    bool    `json:"displayRatingsSection"`
	PercentageRecommended    float64 `json:"percentageRecommended"`
	RatingDistribution       []struct {
		RatingsValue float64 `json:"ratingsValue"`
		Count        float64 `json:"count"`
	} `json:"ratingDistribution"`
	MostRecent struct {
		Rating            float64     `json:"rating"`
		Title             string      `json:"title"`
		ReviewText        string      `json:"reviewText"`
		SubmissionRecency string      `json:"submissionRecency"`
		SyndicationSource interface{} `json:"syndicationSource"`
		BadgesOrder       []string    `json:"badgesOrder"`
		Photos            []struct {
			ThumbnailURL string `json:"thumbnailUrl"`
			FullSizeURL  string `json:"fullSizeUrl"`
		} `json:"photos"`
	} `json:"mostRecent"`
}

var (
	detailReg     = regexp.MustCompile(`(?U)window\.asos\.pdp\.config\.product\s*=\s*({[^;]+});`)
	stockPriceReg = regexp.MustCompile(`(?U)window\.asos\.pdp\.config\.stockPriceApiUrl\s*=\s*'(/api/product/catalogue/[^;]+)'\s*;`)
	appVersionReg = regexp.MustCompile(`(?U)window\.asos\.pdp\.config\.appVersion\s*=\s*'([a-z0-9-.]+)';`)
	ratingReg     = regexp.MustCompile(`(?U)window\.asos\.pdp\.config\.ratings\s*=\s*({.*});`)
	descReg       = regexp.MustCompile(`(?U)<script\s+id="split\-structured\-data"\s+type="application/ld\+json">(.*)</script>`)
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
	matched := detailReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		c.logger.Debugf("data %s", respBody)
		return fmt.Errorf("extract produt json from page %s content failed", resp.Request.URL)
	}
	matchedStock := stockPriceReg.FindSubmatch(respBody)
	if len(matchedStock) <= 1 {
		return fmt.Errorf("extract stock url from page %s content failed", resp.Request.URL)
	}
	matchedRating := ratingReg.FindSubmatch(respBody)
	matchedDesc := descReg.FindSubmatch(respBody)

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	var (
		i      parseProductResponse
		sp     *parseProductStockPrice
		rating parseProductRatingResponse
		desc   struct {
			Desc string `json:"description"`
		}
		variants = map[int]*productVariant{}
	)
	if err = json.Unmarshal(matched[1], &i); err != nil {
		c.logger.Error(err)
		return err
	}
	if len(matchedRating) > 1 {
		if err = json.Unmarshal(matchedRating[1], &rating); err != nil {
			c.logger.Errorf("%s, error=%s", matchedRating[1], err)
			return err
		}
	}
	if len(matchedDesc) > 1 {
		if err = json.Unmarshal(matchedDesc[1], &desc); err != nil {
			c.logger.Error(err)
			return err
		}
	}

	matched = appVersionReg.FindSubmatch(respBody)
	if len(matched) > 1 {
		// fetch stock
		stockUrl := fmt.Sprintf("%s://%s%s", resp.Request.URL.Scheme, resp.Request.URL.Host, matchedStock[1])
		req, err := http.NewRequest(http.MethodGet, stockUrl, nil)
		if err != nil {
			c.logger.Error(err)
			return err
		}
		vals := req.URL.Query()
		vals.Set("store", "US")
		vals.Set("currency", "USD")
		req.URL.RawQuery = vals.Encode()

		opts := c.CrawlOptions(resp.Request.URL)
		for k := range opts.MustHeader {
			req.Header.Set(k, opts.MustHeader.Get(k))
		}
		req.Header.Set("accept-encoding", "gzip, deflate, br")
		req.Header.Set("accept", "*/*")
		req.Header.Set("Referer", resp.Request.URL.String())
		req.Header.Set("User-Agent", resp.Request.Header.Get("User-Agent"))
		req.Header.Set("asos-c-name", "asos-web-productpage")
		req.Header.Set("asos-c-version", string(matched[1]))

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

		resp, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
			EnableProxy:    true,
			EnableHeadless: c.CrawlOptions(resp.Request.URL).EnableHeadless,
			Reliability:    c.CrawlOptions(resp.Request.URL).Reliability,
		})
		if err != nil {
			c.logger.Error(err)
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			c.logger.Errorf("status is %v", resp.StatusCode)
			return fmt.Errorf(resp.Status)
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			c.logger.Error(err)
			return err
		}

		var stockPrices []*parseProductStockPrice
		if err := json.Unmarshal(data, &stockPrices); err != nil {
			c.logger.Errorf("%s, error=%s", data, err)
			return err
		}
		if len(stockPrices) > 0 {
			sp = stockPrices[0]
			for _, v := range sp.Variants {
				variants[v.VariantID] = v
			}
		} else {
			return fmt.Errorf("get no stock price from url %s", stockUrl)
		}
	}

	canUrl := dom.Find(`link[rel="canonical"]`).AttrOr("href", "")
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           strconv.Format(i.ID),
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
		},
		Title:       i.Name,
		Description: desc.Desc,
		BrandName:   i.BrandName,
		CrowdType:   i.Gender,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
			Current:  int32(sp.ProductPrice.Current.Value * 100),
		},
		Stats: &pbItem.Stats{
			Rating:      float32(rating.AverageOverallRating),
			ReviewCount: int32(rating.TotalReviewCount),
		},
	}
	if i.IsInStock {
		item.Stock = &pbItem.Stock{
			StockStatus: pbItem.Stock_InStock,
		}
	}
	breadSel := dom.Find(`nav[aria-label="breadcrumbs"]>ol>li`)
	for i := range breadSel.Nodes {
		if i == len(breadSel.Nodes)-1 {
			break
		}
		node := breadSel.Eq(i)
		switch i {
		case 1:
			item.Category = strings.TrimSpace(node.Find("a").Text())
		case 2:
			item.SubCategory = strings.TrimSpace(node.Find("a").Text())
		case 3:
			item.SubCategory2 = strings.TrimSpace(node.Find("a").Text())
		case 4:
			item.SubCategory3 = strings.TrimSpace(node.Find("a").Text())
		case 5:
			item.SubCategory4 = strings.TrimSpace(node.Find("a").Text())
		}
	}

	for _, img := range i.Images {
		itemImg, _ := anypb.New(&media.Media_Image{
			OriginalUrl: img.URL,
			LargeUrl:    img.URL + "?wid=1000&fit=constrain", // $S$, $XXL$
			MediumUrl:   img.URL + "?wid=650&fit=constrain",
			SmallUrl:    img.URL + "?wid=500&fit=constrain",
		})
		item.Medias = append(item.Medias, &media.Media{
			Detail:    itemImg,
			IsDefault: img.IsPrimary,
		})
	}

	for _, variant := range i.Variants {
		vv, ok := variants[variant.VariantID]
		if !ok {
			continue
		}
		sku := pbItem.Sku{
			SourceId:    strconv.Format(variant.VariantID),
			Title:       i.Name,
			Description: "",
			Price: &pbItem.Price{
				// 接口里返回的都是美元价格，请求的页面path有个 /us 前缀
				Currency: regulation.Currency_USD,
				Current:  int32(vv.Price.Current.Value * 100),
				Msrp:     int32(vv.Price.Previous.Value * 100),
			},
			Stock: &pbItem.Stock{
				StockStatus: pbItem.Stock_OutOfStock,
			},
			Specs: []*pbItem.SkuSpecOption{
				{
					Type:  pbItem.SkuSpecType_SkuSpecColor,
					Name:  variant.Colour,
					Value: strconv.Format(variant.ColourWayID),
				},
				{
					Type:  pbItem.SkuSpecType_SkuSpecSize,
					Name:  variant.Size,
					Value: strconv.Format(variant.SizeID),
				},
			},
		}
		if vv.IsInStock {
			sku.Stock.StockStatus = pbItem.Stock_InStock
		}
		if item.Price.Msrp == 0 {
			item.Price.Msrp = sku.Price.Msrp
		}
		item.SkuItems = append(item.SkuItems, &sku)
	}
	return yield(ctx, &item)
}

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		// "https://www.asos.com/us/women/outlet/ctas/timed-sales/timed-sale-1/cat/?cid=28030",
		// "https://www.asos.com/us/women/accessories/scarves/cat/?cid=6452&nlid=ww|accessories|shop+by+product",
		"https://www.asos.com/us/women/tops/cat/?cid=4169&currentpricerange=0-365&nlid=ww|clothing|shop%20by%20product&refine=attribute_1047:8392|attribute_10156:50477",
		// "https://www.asos.com/us/women/face-body/makeup/cat/?cid=5020&ctaref=cat_header&currentpricerange=5-75&refine=attribute_1047:9778,8461",
		// "https://www.asos.com/us/catch/catch-exclusive-ribbed-tie-cardigan-set-in-beige/grp/34104?colourwayid=60431494&cid=2623",
		// "https://www.asos.com/us/olivia-burton/olivia-burton-white-dial-midi-mesh-watch-in-rose-gold/prd/22313628?colourwayid=60389207&cid=5088",
		// "https://www.asos.com/us/women/new-in/new-in-clothing/cat/?cid=2623&nlid=ww%7Cclothing%7Cshop%20by%20product&page=1",
		// "https://www.asos.com/api/product/search/v2/categories/2623?channel=desktop-web&country=US&currency=USD&keyStoreDataversion=3pmn72e-27&lang=en-US&limit=72&nlid=ww%7Cclothing%7Cshop+by+product&offset=72&rowlength=4&store=US",
		// "https://www.asos.com/us/missguided-plus/missguided-plus-oversized-long-sleeve-t-shirt-in-gray-snake-tie-dye/prd/23385813?colourwayid=60477943&SearchQuery=&cid=4169",
		// "https://www.asos.com/us/asos-design/asos-design-tie-front-maxi-beach-set-in-black/grp/33060?colourwayid=60343707#22019820&SearchQuery=&cid=2623",
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
