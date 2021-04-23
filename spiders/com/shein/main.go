package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/voiladev/go-crawler/pkg/cli"
	"github.com/voiladev/go-crawler/pkg/crawler"
	"github.com/voiladev/go-crawler/pkg/net/http"
	pbMedia "github.com/voiladev/go-crawler/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/go-crawler/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/go-crawler/protoc-gen-go/chameleon/smelter/v1/crawl/item"
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
		categoryPathMatcher: regexp.MustCompile(`-(c|sc)-[0-9]+.html`),
		// this regular used to match product page url path
		productPathMatcher: regexp.MustCompile(`-p-[0-9]+-cat-[0-9]+.html`),
		logger:             logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "6919a61af19b618b11b3d176a12a39b9"
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
func (c *_Crawler) CrawlOptions(u *url.URL) *crawler.CrawlOptions {
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

func (c *_Crawler) CanonicalUrl(rawurl string) (string, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	if u.Host == "" {
		u.Host = "us.shein.com"
	}
	if c.productPathMatcher.MatchString(u.Path) {
		u.RawQuery = ""
		return u.String(), nil
	}
	return rawurl, nil
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
		Nlkt int `json:"nlkt"`
		// PdeCacheControl bool `json:"pdeCacheControl"`
		MinPrice   int  `json:"min_price"`
		MaxPrice   int  `json:"max_price"`
		IsPlusSize bool `json:"isPlusSize"`
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
		rawurl := idv.PretreatInfo.GoodsDetailURL
		//c.logger.Debugf(rawurl)
		req, err := http.NewRequest(http.MethodGet, rawurl, nil)
		if err != nil {
			c.logger.Errorf("load http request of url %s failed, error=%s", idv.PretreatInfo.GoodsDetailURL, err)
			return err
		}
		if strings.HasSuffix(req.URL.Path, ".html") {
			req.URL.RawQuery = ""
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

type parentCategory struct {
	CatID       string `json:"cat_id"`
	URLCatID    string `json:"url_cat_id"`
	GoodsTypeID string `json:"goods_type_id"`
	CatURLName  string `json:"cat_url_name"`
	CatName     string `json:"cat_name"`
	ParentID    string `json:"parent_id"`
	SortOrder   string `json:"sort_order"`
	IsLeaf      string `json:"is_leaf"`
	GoodsNum    string `json:"goods_num"`
	Multi       struct {
		CatID           string `json:"cat_id"`
		CatName         string `json:"cat_name"`
		MetaTitle       string `json:"meta_title"`
		MetaKeywords    string `json:"meta_keywords"`
		MetaDescription string `json:"meta_description"`
		CatDesc         string `json:"cat_desc"`
		LanguageFlag    string `json:"language_flag"`
	} `json:"multi"`
	Children []*parentCategory `json:"children"`
}

type parseProductData struct {
	ParentCats parentCategory `json:"parentCats"`
	GoodsImgs  struct {
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
		GoodsID               string `json:"goods_id"`
		CatID                 string `json:"cat_id"`
		GoodsSn               string `json:"goods_sn"`
		GoodsURLName          string `json:"goods_url_name"`
		SupplierID            string `json:"supplier_id"`
		GoodsName             string `json:"goods_name"`
		OriginalImg           string `json:"original_img"`
		GoodsThumb            string `json:"goods_thumb"`
		GoodsImg              string `json:"goods_img"`
		IsStockEnough         string `json:"is_stock_enough"`
		Brand                 string `json:"brand"`
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
			AttrID      interface{} `json:"attr_id"`
			AttrValueID string      `json:"attr_value_id"`
			AttrName    string      `json:"attr_name"`
			AttrNameEn  string      `json:"attr_name_en"`
			ValueSort   int         `json:"value_sort"`
			AttrSelect  int         `json:"attr_select"`
			AttrSort    int         `json:"attr_sort"`
			LeftShow    int         `json:"left_show"`
			AttrValue   string      `json:"attr_value"`
			AttrValueEn string      `json:"attr_value_en"`
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
		ColorType         interface{}   `json:"color_type"`
		BeautyCategory    bool          `json:"beautyCategory"`
		NeedAttrRelation  bool          `json:"needAttrRelation"`
		BrandInfo         interface{}   `json:"brandInfo"`
		Series            interface{}   `json:"series"`
	} `json:"detail"`
	AttrSizeList []struct {
		AttrID       string      `json:"attr_id"`
		AttrValueID  string      `json:"attr_value_id"`
		AttrName     string      `json:"attr_name"`
		AttrValue    string      `json:"attr_value"`
		AttrValueEn  string      `json:"attr_value_en"`
		Stock        interface{} `json:"stock"`
		AttrStdValue string      `json:"attr_std_value"`
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
	RelationColor []struct {
		GoodsID      string `json:"goods_id"`
		CatID        string `json:"cat_id"`
		GoodsSn      string `json:"goods_sn"`
		GoodsURLName string `json:"goods_url_name"`
		SupplierID   string `json:"supplier_id"`
		GoodsName    string `json:"goods_name"`
	} `json:"relation_color"`
	SoldoutColor []struct {
		GoodsID      string `json:"goods_id"`
		CatID        string `json:"cat_id"`
		GoodsSn      string `json:"goods_sn"`
		GoodsURLName string `json:"goods_url_name"`
		SupplierID   string `json:"supplier_id"`
		GoodsName    string `json:"goods_name"`
	} `json:"soldoutColor"`
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
	canUrl, _ := c.CanonicalUrl(resp.Request.URL.String())
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           strconv.Format(viewData.Detail.GoodsID),
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
			GroupId:      viewData.Detail.ProductRelationID,
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
	item.CrowdType = viewData.ParentCats.CatName
	item.Category = viewData.ParentCats.CatName
	if len(viewData.ParentCats.Children) > 0 {
		item.SubCategory = viewData.ParentCats.Children[0].CatName
		if len(viewData.ParentCats.Children[0].Children) > 0 {
			item.SubCategory2 = viewData.ParentCats.Children[0].Children[0].CatName
			if len(viewData.ParentCats.Children[0].Children[0].Children) > 0 {
				item.SubCategory3 = viewData.ParentCats.Children[0].Children[0].Children[0].CatName
				if len(viewData.ParentCats.Children[0].Children[0].Children[0].Children) > 0 {
					item.SubCategory4 = viewData.ParentCats.Children[0].Children[0].Children[0].Children[0].CatName
				}
			}
		}
	}

	colorName := ""
	colorId := ""
	for _, rowdes := range viewData.Detail.ProductDetails {
		if rowdes.AttrName == "Color" {
			colorName = rowdes.AttrValue
			colorId = rowdes.AttrValueID
			break
		}
	}

	colorSpec := pbItem.SkuSpecOption{
		Type:  pbItem.SkuSpecType_SkuSpecColor,
		Id:    colorId,
		Name:  colorName,
		Value: colorName,
		Icon:  viewData.Detail.ColorImage,
	}

	var medias []*pbMedia.Media
	m := viewData.GoodsImgs.MainImage
	medias = append(medias, pbMedia.NewImageMedia(
		"",
		m.OriginImage,
		strings.ReplaceAll(m.OriginImage, ".jpg", "_thumbnail_900x.jpg"),
		strings.ReplaceAll(m.OriginImage, ".jpg", "_thumbnail_600x.jpg"),
		strings.ReplaceAll(m.OriginImage, ".jpg", "_thumbnail_500x.jpg"),
		"",
		true,
	))

	for _, m := range viewData.GoodsImgs.DetailImage {
		medias = append(medias, pbMedia.NewImageMedia(
			"",
			m.OriginImage,
			strings.ReplaceAll(m.OriginImage, ".jpg", "_thumbnail_900x.jpg"),
			strings.ReplaceAll(m.OriginImage, ".jpg", "_thumbnail_600x.jpg"),
			strings.ReplaceAll(m.OriginImage, ".jpg", "_thumbnail_500x.jpg"),
			"",
			false,
		))
	}

	if len(viewData.AttrSizeList) > 0 {
		for _, rawSku := range viewData.AttrSizeList {
			originalPrice, _ := strconv.ParseFloat(rawSku.Price.SalePrice.UsdAmount)
			msrp, _ := strconv.ParseFloat(rawSku.Price.RetailPrice.UsdAmount)
			discount := ((originalPrice - msrp) / msrp) * 100

			sku := pbItem.Sku{
				SourceId: fmt.Sprintf("%s-%s", viewData.Detail.GoodsSn, rawSku.AttrValueID),
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(originalPrice * 100),
					Msrp:     int32(msrp * 100),
					Discount: int32(discount),
				},
				Medias: medias,
				Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}
			stock, _ := strconv.ParseInt(rawSku.Stock)
			c.logger.Debugf("%s %d", rawSku.Stock, stock)
			if stock > 0 {
				sku.Stock.StockStatus = pbItem.Stock_InStock
				sku.Stock.StockCount = int32(stock)
			}
			sku.Specs = append(sku.Specs, &colorSpec, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    rawSku.AttrValueID,
				Name:  rawSku.AttrValueEn,
				Value: rawSku.AttrValueEn,
			})
			item.SkuItems = append(item.SkuItems, &sku)
		}
	} else {
		prod := viewData.Detail
		originalPrice, _ := strconv.ParseFloat(prod.SalePrice.UsdAmount)
		msrp, _ := strconv.ParseFloat(prod.RetailPrice.UsdAmount)
		discount := ((originalPrice - msrp) / msrp) * 100

		sku := pbItem.Sku{
			SourceId: viewData.Detail.GoodsSn,
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(originalPrice * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},
			Specs:  []*pbItem.SkuSpecOption{&colorSpec},
			Medias: medias,
			Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}
		stock, _ := strconv.ParseInt(prod.Stock)
		if stock > 0 {
			sku.Stock.StockStatus = pbItem.Stock_InStock
			sku.Stock.StockCount = int32(stock)
		}
		item.SkuItems = append(item.SkuItems, &sku)
	}

	if len(item.SkuItems) > 0 {
		// yield item result
		if err = yield(ctx, &item); err != nil {
			return err
		}
	} else {
		return errors.New("no sku spec found")
	}

	for _, color := range viewData.RelationColor {
		u := fmt.Sprintf("https://us.shein.com/%s-p-%s-cat-%s.html", strings.ReplaceAll(color.GoodsURLName, " ", "-"), color.GoodsID, color.CatID)
		req, err := http.NewRequest(http.MethodGet, u, nil)
		if err != nil {
			c.logger.Error(err)
			continue
		}
		if err := yield(ctx, req); err != nil {
			return err
		}
	}
	for _, color := range viewData.SoldoutColor {
		u := fmt.Sprintf("https://us.shein.com/%s-p-%s-cat-%s.html", strings.ReplaceAll(color.GoodsURLName, " ", "-"), color.GoodsID, color.CatID)
		req, err := http.NewRequest(http.MethodGet, u, nil)
		if err != nil {
			c.logger.Error(err)
			continue
		}
		if err := yield(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		// "https://www.nordstrom.com/browse/activewear/women-clothing?breadcrumb=Home%2FWomen%2FClothing%2FActivewear&origin=topnav",
		// "https://us.shein.com/T-Shirts-c-1738.html?ici=us_tab01navbar04menu01dir02&scici=navbar_WomenHomePage~~tab01navbar04menu01dir02~~4_1_2~~real_1738~~~~0~~50001&srctype=category&userpath=category%3ECLOTHING%3ETOPS%3ET-Shirts",
		// "https://us.shein.com//Slogan-Graphic-Crop-Tee-p-1854713-cat-1738.html?scici=navbar_WomenHomePage~~tab01navbar04menu01dir02~~4_1_2~~real_1738~~~~0~~50001",
		// "https://us.shein.com/Rhinestone-Decor-Hair-Hoop-p-1315510-cat-1778.html?scici=navbar_2~~tab01navbar08menu11~~8_11~~real_1765~~SPcCccWomenCategory_default~~0~~0",
		// "https://us.shein.com/Allover-Print-Surplice-Front-Layered-Hem-Dress-p-1575863-cat-1727.html?scici=navbar_WomenHomePage~~tab01navbar06menu05dir01~~6_5_1~~itemPicking_00109364~~~~0~~50001",
		// "https://us.shein.com//Planet-Embroidered-Bucket-Hat-p-2127810-cat-1772.html?scici=navbar_2~~tab01navbar08menu11dir06~~8_11_6~~real_1772~~SPcCccWomenCategory_default~~0~~0",
		"https://us.shein.com/category/Sleepwear-and-Nightwear-sc-00821505.html?icn=category&ici=us_tab01navbar02menu14dir02&srctype=category&userpath=category%3EWOMEN%3ECLOTHING%3ELoungewear-Sleepwear%3ESleepwear&scici=navbar_2~~tab01navbar02menu14dir02~~2_14_2~~itemPicking_00821505~~SPcCccWomenCategory_default~~0~~0",
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

// main func is the entry of golang program. this will not be used by plugin, just for local spider test.
func main() {
	cli.NewApp(New).Run(os.Args)
}
