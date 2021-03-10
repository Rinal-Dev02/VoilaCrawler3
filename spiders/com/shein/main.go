package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/voiladev/VoilaCrawl/pkg/crawler"
	"github.com/voiladev/VoilaCrawl/pkg/net/http"
	"github.com/voiladev/VoilaCrawl/pkg/net/http/cookiejar"
	"github.com/voiladev/VoilaCrawl/pkg/proxy"

	pbMedia "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/item"

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
		categoryPathMatcher: regexp.MustCompile(`-c-[0-9]+.html`),
		// this regular used to match product page url path
		productPathMatcher: regexp.MustCompile(`-p-[0-9]+-cat-[0-9]+.html`),
		logger:             logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "359c18d126814836b290b0bfb96549f9"
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
func (c *_Crawler) CrawlOptions() *crawler.CrawlOptions {
	return &crawler.CrawlOptions{
		EnableHeadless: false,
		// use js api to init session for the first request of the crawl
		EnableSessionInit: true,
	}
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.shein.com"}
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
	return fmt.Errorf("unsupported url %s", resp.Request.URL.String())
}

// nextIndex used to get the index from the shared data.
// item.index is a const key for item index.
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")))
}

// below are the golang json data struct of raw website.
// if you get the raw json data of the website,
// then you can use https://mholt.github.io/json-to-go/ to convert it to a golang struct

type ProductCategoryStructure struct {
	Results struct {
		Goods []struct {
			Brand             interface{} `json:"brand"`
			ProductRelationID string      `json:"productRelationID"`
			RelatedColor      []struct {
				GoodsID string `json:"goods_id"`
				CatID   string `json:"cat_id"`
			} `json:"relatedColor"`
			PretreatInfo struct {
				GoodsName             string `json:"goodsName"`
				SeriesOrBrandAnalysis string `json:"seriesOrBrandAnalysis"`
				GoodsDetailURL        string `json:"goodsDetailUrl"`
			} `json:"pretreatInfo"`
		} `json:"goods"`
		TemplateType string `json:"templateType"`
		Sum          int    `json:"sum"`
		SumForPage   int    `json:"sumForPage"`
		TraceID      string `json:"trace_id"`
		GoodsCrowID  struct {
		} `json:"goodsCrowId"`
		Nlkt            int  `json:"nlkt"`
		PdeCacheControl bool `json:"pdeCacheControl"`
		MinPrice        int  `json:"min_price"`
		MaxPrice        int  `json:"max_price"`
		IsPlusSize      bool `json:"isPlusSize"`
	} `json:"results"`
}

// used to extract embaded json data in website page.
// more about golang regulation see here https://golang.org/pkg/regexp/syntax/
var productsExtractReg = regexp.MustCompile(`var gbProductListSsrData\s*=\s*({.*})`)
var productDataExtractReg = regexp.MustCompile(`productIntroData:\s*({.*}),`)

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
	// c.logger.Debugf("data: %s", matched[1])

	var viewData ProductCategoryStructure
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		return err
	}

	lastIndex := nextIndex(ctx)
	for _, idv := range viewData.Results.Goods {

		rawurl := fmt.Sprintf("%s://%s/%s", resp.Request.URL.Scheme, resp.Request.URL.Host, idv.PretreatInfo.GoodsDetailURL)
		//c.logger.Debugf(rawurl)
		req, err := http.NewRequest(http.MethodGet, rawurl, nil)
		if err != nil {
			c.logger.Errorf("load http request of url %s failed, error=%s", idv.PretreatInfo.GoodsDetailURL, err)
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

	// get current page number
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}

	totalPageCount := viewData.Results.Sum

	// check if this is the last page
	if lastIndex >= int(totalPageCount) {
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

type parseProductData struct {
	GoodsImgs struct {
		MainImage struct {
			OriginImage     string `json:"origin_image"`
			Thumbnail       string `json:"thumbnail"`
			MediumImage     string `json:"medium_image"`
			ThumbnailWebp   string `json:"thumbnail_webp"`
			OriginImageWebp string `json:"origin_image_webp"`
			MediumImageWebp string `json:"medium_image_webp"`
		} `json:"main_image"`
		DetailImage []struct {
			OriginImage     string `json:"origin_image"`
			MediumImage     string `json:"medium_image"`
			Thumbnail       string `json:"thumbnail"`
			ThumbnailWebp   string `json:"thumbnail_webp"`
			OriginImageWebp string `json:"origin_image_webp"`
			MediumImageWebp string `json:"medium_image_webp"`
		} `json:"detail_image"`
		VideoURL string `json:"video_url"`
	} `json:"goods_imgs"`
	Detail struct {
		GoodsID       string `json:"goods_id"`
		CatID         string `json:"cat_id"`
		GoodsSn       string `json:"goods_sn"`
		GoodsURLName  string `json:"goods_url_name"`
		SupplierID    string `json:"supplier_id"`
		GoodsName     string `json:"goods_name"`
		OriginalImg   string `json:"original_img"`
		GoodsThumb    string `json:"goods_thumb"`
		GoodsImg      string `json:"goods_img"`
		IsStockEnough string `json:"is_stock_enough"`
		Brand         string `json:"brand"`
		SizeTemplate  struct {
			ImageURL         string `json:"image_url"`
			DescriptionMulti []struct {
				Sort        int    `json:"sort"`
				Name        string `json:"name"`
				Description string `json:"description"`
			} `json:"description_multi"`
		} `json:"sizeTemplate"`
		GoodsDesc             string `json:"goods_desc"`
		SupplierTopCategoryID string `json:"supplier_top_category_id"`
		ParentID              string `json:"parent_id"`
		IsOnSale              string `json:"is_on_sale"`
		IsVirtualStock        string `json:"is_virtual_stock"`
		Stock                 string `json:"stock"`
		IsInit                string `json:"is_init"`
		IsPreSale             string `json:"is_pre_sale"`
		IsPreSaleEnd          string `json:"is_pre_sale_end"`
		ProductDetails        []struct {
			AttrID      int    `json:"attr_id"`
			AttrValueID string `json:"attr_value_id"`
			AttrName    string `json:"attr_name"`
			AttrNameEn  string `json:"attr_name_en"`
			ValueSort   int    `json:"value_sort"`
			AttrSelect  int    `json:"attr_select"`
			AttrSort    int    `json:"attr_sort"`
			LeftShow    int    `json:"left_show"`
			AttrValue   string `json:"attr_value"`
			AttrValueEn string `json:"attr_value_en"`
		} `json:"productDetails"`
		Comment struct {
			CommentNum  string `json:"comment_num"`
			CommentRank string `json:"comment_rank"`
		} `json:"comment"`
		IsSubscription string        `json:"is_subscription"`
		PromotionInfo  []interface{} `json:"promotionInfo"`
		Promotion      interface{}   `json:"promotion"`
		RetailPrice    struct {
			Amount              string `json:"amount"`
			AmountWithSymbol    string `json:"amountWithSymbol"`
			UsdAmount           string `json:"usdAmount"`
			UsdAmountWithSymbol string `json:"usdAmountWithSymbol"`
		} `json:"retailPrice"`
		ProductRelationID string `json:"productRelationID"`
		ColorImage        string `json:"color_image"`
		GoodsImgWebp      string `json:"goods_img_webp"`
		OriginalImgWebp   string `json:"original_img_webp"`
		SalePrice         struct {
			Amount              string `json:"amount"`
			AmountWithSymbol    string `json:"amountWithSymbol"`
			UsdAmount           string `json:"usdAmount"`
			UsdAmountWithSymbol string `json:"usdAmountWithSymbol"`
		} `json:"salePrice"`
		UnitDiscount    string `json:"unit_discount"`
		SpecialPriceOld struct {
			Amount              string `json:"amount"`
			AmountWithSymbol    string `json:"amountWithSymbol"`
			UsdAmount           string `json:"usdAmount"`
			UsdAmountWithSymbol string `json:"usdAmountWithSymbol"`
		} `json:"special_price_old"`
		IsClearance int    `json:"is_clearance"`
		LimitCount  string `json:"limit_count"`
		FlashGoods  struct {
			IsFlashGoods int `json:"is_flash_goods"`
		} `json:"flash_goods"`
		IsPriceConfigured int           `json:"isPriceConfigured"`
		AppPromotion      []interface{} `json:"appPromotion"`
		RewardPoints      int           `json:"rewardPoints"`
		DoublePoints      int           `json:"doublePoints"`
		ColorType         string        `json:"color_type"`
		BeautyCategory    bool          `json:"beautyCategory"`
		NeedAttrRelation  bool          `json:"needAttrRelation"`
		BrandInfo         interface{}   `json:"brandInfo"`
		Series            interface{}   `json:"series"`
	} `json:"detail"`
	AttrSizeList []struct {
		AttrID       string `json:"attr_id"`
		AttrValueID  string `json:"attr_value_id"`
		AttrName     string `json:"attr_name"`
		AttrValue    string `json:"attr_value"`
		AttrValueEn  string `json:"attr_value_en"`
		Stock        int    `json:"stock"`
		AttrStdValue string `json:"attr_std_value"`
		Price        struct {
			RetailPrice struct {
				Amount              string `json:"amount"`
				AmountWithSymbol    string `json:"amountWithSymbol"`
				UsdAmount           string `json:"usdAmount"`
				UsdAmountWithSymbol string `json:"usdAmountWithSymbol"`
			} `json:"retailPrice"`
			UnitDiscount int `json:"unit_discount"`
			SalePrice    struct {
				Amount              string `json:"amount"`
				AmountWithSymbol    string `json:"amountWithSymbol"`
				UsdAmount           string `json:"usdAmount"`
				UsdAmountWithSymbol string `json:"usdAmountWithSymbol"`
			} `json:"salePrice"`
		} `json:"price"`
		RewardPoints int `json:"rewardPoints"`
		DoublePoints int `json:"doublePoints"`
	} `json:"attrSizeList"`
	CommentInfo struct {
		CommentNum         int    `json:"comment_num"`
		CommentRankAverage string `json:"comment_rank_average"`
	} `json:"commentInfo"`
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

	matched := productDataExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}

	var viewData parseProductData

	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}

	rating, _ := strconv.ParseFloat(viewData.CommentInfo.CommentRankAverage)
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:       strconv.Format(viewData.Detail.GoodsID),
			CrawlUrl: resp.Request.URL.String(),
		},
		BrandName:   viewData.Detail.Brand,
		Title:       viewData.Detail.GoodsName,
		Description: viewData.Detail.GoodsDesc,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stats: &pbItem.Stats{
			ReviewCount: int32(viewData.CommentInfo.CommentNum),
			Rating:      float32(rating),
		},
	}

	colorName := ""
	for _, rowdes := range viewData.Detail.ProductDetails {
		if rowdes.AttrName == "Color" {
			colorName = rowdes.AttrValue
			break
		}
	}

	for ks, rawSku := range viewData.AttrSizeList {
		originalPrice, _ := strconv.ParseFloat(rawSku.Price.SalePrice.UsdAmount)
		msrp, _ := strconv.ParseFloat(rawSku.Price.RetailPrice.UsdAmount)
		discount := ((originalPrice - msrp) / msrp) * 100

		sku := pbItem.Sku{
			SourceId: strconv.Format(ks),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(originalPrice * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},
			Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}
		if rawSku.Stock > 0 {
			sku.Stock.StockStatus = pbItem.Stock_InStock
			sku.Stock.StockCount = int32(rawSku.Stock)
		}

		// color

		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecColor,
			Id:    strconv.Format(viewData.Detail.GoodsID),
			Name:  colorName,
			Value: colorName,
			//Icon:  color.SwatchMedia.Mobile,
		})

		if ks == 0 {
			m := viewData.GoodsImgs.MainImage
			sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
				strconv.Format(m),
				m.OriginImage,
				strings.ReplaceAll(m.OriginImage, ".jpg", "_thumbnail_900x.jpg"),
				strings.ReplaceAll(m.OriginImage, ".jpg", "_thumbnail_600x.jpg"),
				strings.ReplaceAll(m.OriginImage, ".jpg", "_thumbnail_500x.jpg"),
				"",
				true,
			))

			for _, m := range viewData.GoodsImgs.DetailImage {
				sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
					strconv.Format(m),
					m.OriginImage,
					strings.ReplaceAll(m.OriginImage, ".jpg", "_thumbnail_900x.jpg"),
					strings.ReplaceAll(m.OriginImage, ".jpg", "_thumbnail_600x.jpg"),
					strings.ReplaceAll(m.OriginImage, ".jpg", "_thumbnail_500x.jpg"),
					"",
					false,
				))
			}
		}

		// size
		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Id:    rawSku.AttrID,
			Name:  rawSku.AttrValue,
			Value: rawSku.AttrValue,
		})

		item.SkuItems = append(item.SkuItems, &sku)
	}

	// yield item result
	if err = yield(ctx, &item); err != nil {
		return err
	}

	return nil
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		// "https://www.nordstrom.com/browse/activewear/women-clothing?breadcrumb=Home%2FWomen%2FClothing%2FActivewear&origin=topnav",
		"https://us.shein.com/T-Shirts-c-1738.html?ici=us_tab01navbar04menu01dir02&scici=navbar_WomenHomePage~~tab01navbar04menu01dir02~~4_1_2~~real_1738~~~~0~~50001&srctype=category&userpath=category%3ECLOTHING%3ETOPS%3ET-Shirts",
		"https://us.shein.com//Slogan-Graphic-Crop-Tee-p-1854713-cat-1738.html?scici=navbar_WomenHomePage~~tab01navbar04menu01dir02~~4_1_2~~real_1738~~~~0~~50001",
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

// local test
func main() {
	logger := glog.New(glog.LogLevelDebug)
	// build a http client
	// get proxy's microservice address from env
	client, err := proxy.NewProxyClient(os.Getenv("VOILA_PROXY_URL"), cookiejar.New(), logger)
	if err != nil {
		panic(err)
	}

	// instance the spider locally
	spider, err := New(client, logger)
	if err != nil {
		panic(err)
	}
	opts := spider.CrawlOptions()

	// this callback func is used to do recursion call of sub requests.
	var callback func(ctx context.Context, val interface{}) error
	callback = func(ctx context.Context, val interface{}) error {
		switch i := val.(type) {
		case *http.Request:
			logger.Debugf("Access %s", i.URL)

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
				i.URL.Host = "www.shein.com"
			}

			// do http requests here.
			nctx, cancel := context.WithTimeout(ctx, time.Minute*5)
			defer cancel()
			resp, err := client.DoWithOptions(nctx, i, http.Options{
				EnableProxy:       false,
				EnableHeadless:    false,
				EnableSessionInit: false,
				KeepSession:       spider.CrawlOptions().KeepSession,
				Reliability:       spider.CrawlOptions().Reliability,
			})
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			return spider.Parse(ctx, resp, callback)
		default:
			// output the result
			data, err := json.Marshal(i)
			if err != nil {
				return err
			}
			logger.Infof("data: %s", data)
		}
		return nil
	}

	ctx := context.WithValue(context.Background(), "tracing_id", fmt.Sprintf("tracing_%d", time.Now().UnixNano()))
	// start the crawl request
	for _, req := range spider.NewTestRequest(context.Background()) {
		if err := callback(ctx, req); err != nil {
			logger.Fatal(err)
		}
	}
}
