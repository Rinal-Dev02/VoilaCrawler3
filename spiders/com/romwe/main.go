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

	"github.com/PuerkitoBio/goquery"
	"github.com/voiladev/go-crawler/pkg/cli"
	"github.com/voiladev/go-crawler/pkg/crawler"
	"github.com/voiladev/go-crawler/pkg/net/http"
	"github.com/voiladev/go-crawler/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/go-crawler/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/go-crawler/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	pbProxy "github.com/voiladev/go-crawler/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
)

type _Crawler struct {
	httpClient http.Client

	categoryPathMatcher *regexp.Regexp
	productPathMatcher  *regexp.Regexp
	logger              glog.Log
}

func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient:          client,
		categoryPathMatcher: regexp.MustCompile(`^/[A-Za-z0-9_-]+(c|sc)\-\d+.html$`),
		productPathMatcher:  regexp.MustCompile(`^/[A-Za-z0-9_-]+(p-\d+-cat-\d+).html$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "8aae1ca41208ffb9f39ddaf389e1c951"
}

// Version
func (c *_Crawler) Version() int32 {
	return 1
}

// CrawlOptions
func (c *_Crawler) CrawlOptions(u *url.URL) *crawler.CrawlOptions {
	options := crawler.NewCrawlOptions()
	options.EnableHeadless = false
	options.EnableSessionInit = false
	options.Reliability = pbProxy.ProxyReliability_ReliabilityMedium
	options.MustCookies = append(options.MustCookies,
		&http.Cookie{Name: "ckm-ctx-sf", Value: `%2F`, Path: "/"},
		&http.Cookie{Name: "country", Value: `US`, Path: "/"},
		&http.Cookie{Name: "countryId", Value: `226`, Path: "/"},
		&http.Cookie{Name: "currency", Value: `USD`, Path: "/"},
		&http.Cookie{Name: "default_currency", Value: `USD`, Path: "/"},
		&http.Cookie{Name: "default_currency_expire", Value: `1`, Path: "/"},
		&http.Cookie{Name: "app_country", Value: "US", Path: "/"},
		&http.Cookie{Name: "default_currency_expire", Value: `1`, Path: "/"},
	)
	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.romwe.com"}
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
		u.Host = "us.romwe.com"
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

	if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	} else if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yield)
	}
	return crawler.ErrUnsupportedPath
}

// nextIndex used to get sharingData from context
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")) + 1)
}

type productListType struct {
	Results struct {
		Goods []struct {
			Index        int `json:"index"`
			PretreatInfo struct {
				GoodsName             string `json:"goodsName"`
				SeriesOrBrandAnalysis string `json:"seriesOrBrandAnalysis"`
				GoodsDetailURL        string `json:"goodsDetailUrl"`
			} `json:"pretreatInfo"`
		} `json:"goods"`
		CatInfo struct {
			//Page        int    `json:"page"`
			Limit       int    `json:"limit"`
			OriginalURL string `json:"originalUrl"`
		} `json:"cat_info"`
		Sum int `json:"sum"`
	} `json:"results"`
}

var prodDataExtraReg1 = regexp.MustCompile(`(?U)var\s*gbProductListSsrData\s*=\s*({.*})\s*</script>`)

// parseCategoryProducts parse api url from web page url
func (c *_Crawler) parseCategoryProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.logger.Debug(err)
		return err
	}
	// next page
	matched := prodDataExtraReg1.FindSubmatch(respBody)
	if len(matched) == 0 {
		c.httpClient.Jar().Clear(ctx, resp.Request.URL)
		return fmt.Errorf("extract json from product list page %s failed", resp.Request.URL)
	}

	var r productListType
	if err = json.Unmarshal([]byte(matched[1]), &r); err != nil {
		c.logger.Debugf("parse %s failed, error=%s", matched[1], err)
		return err
	}

	lastIndex := nextIndex(ctx)
	for _, prod := range r.Results.Goods {
		//fmt.Println(prod.PretreatInfo.GoodsDetailURL)

		if req, err := http.NewRequest(http.MethodGet, prod.PretreatInfo.GoodsDetailURL, nil); err != nil {
			c.logger.Debug(err)
			return err
		} else {
			nctx := context.WithValue(ctx, "item.index", lastIndex+1)
			lastIndex += 1
			if err = yield(nctx, req); err != nil {
				return err
			}
		}
	}

	// get current page number
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}
	// check if this is the last page
	totalResults, _ := strconv.ParseInt(r.Results.Sum)
	if lastIndex >= int(totalResults) {
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

type parseProductResponse struct {
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
	Country string `json:"country"`
	Detail  struct {
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
	RelationColor []struct {
		GoodsID      string `json:"goods_id"`
		CatID        string `json:"cat_id"`
		GoodsSn      string `json:"goods_sn"`
		GoodsURLName string `json:"goods_url_name"`
		SupplierID   string `json:"supplier_id"`
		GoodsName    string `json:"goods_name"`
		OriginalImg  string `json:"original_img"`
		Brand        string `json:"brand"`
		SizeTemplate struct {
			ImageURL         string `json:"image_url"`
			DescriptionMulti []struct {
				Sort        int    `json:"sort"`
				Name        string `json:"name"`
				Description string `json:"description"`
			} `json:"description_multi"`
		} `json:"sizeTemplate"`
		IsInit         string `json:"is_init"`
		Stock          string `json:"stock"`
		IsOnSale       string `json:"is_on_sale"`
		IsVirtualStock string `json:"is_virtual_stock"`
		ProductDetails []struct {
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
		PromotionInfo []interface{} `json:"promotionInfo"`
		Promotion     interface{}   `json:"promotion"`
		RelatedColor  []struct {
			GoodsID         string `json:"goods_id"`
			GoodsRelationID string `json:"goods_relation_id"`
			CatID           string `json:"cat_id"`
			GoodsSn         string `json:"goods_sn"`
			GoodsURLName    string `json:"goods_url_name"`
			GoodsName       string `json:"goods_name"`
			GoodsColorImage string `json:"goods_color_image"`
			OriginalImg     string `json:"original_img"`
			GoodsThumb      string `json:"goods_thumb"`
			GoodsImg        string `json:"goods_img"`
		} `json:"relatedColor"`
		ParentID              string `json:"parent_id"`
		SupplierTopCategoryID string `json:"supplier_top_category_id"`
		RetailPrice           struct {
			Amount              string `json:"amount"`
			AmountWithSymbol    string `json:"amountWithSymbol"`
			UsdAmount           string `json:"usdAmount"`
			UsdAmountWithSymbol string `json:"usdAmountWithSymbol"`
		} `json:"retailPrice"`
		ProductRelationID string `json:"productRelationID"`
		ColorImage        string `json:"color_image"`
		SalePrice         struct {
			Amount              string `json:"amount"`
			AmountWithSymbol    string `json:"amountWithSymbol"`
			UsdAmount           string `json:"usdAmount"`
			UsdAmountWithSymbol string `json:"usdAmountWithSymbol"`
		} `json:"salePrice"`
		UnitDiscount string `json:"unit_discount"`
		IsClearance  int    `json:"is_clearance"`
		LimitCount   string `json:"limit_count"`
		FlashGoods   struct {
			IsFlashGoods int `json:"is_flash_goods"`
		} `json:"flash_goods"`
		IsPriceConfigured int    `json:"isPriceConfigured"`
		RewardPoints      int    `json:"rewardPoints"`
		DoublePoints      int    `json:"doublePoints"`
		ColorType         string `json:"color_type"`
		BeautyCategory    bool   `json:"beautyCategory"`
		NeedAttrRelation  bool   `json:"needAttrRelation"`
		Comment           struct {
			CommentNum  string `json:"comment_num"`
			CommentRank string `json:"comment_rank"`
		} `json:"comment"`
		IsSubscription  string `json:"is_subscription"`
		IsPreSale       string `json:"is_pre_sale"`
		IsPreSaleEnd    string `json:"is_pre_sale_end"`
		IsStockEnough   string `json:"is_stock_enough"`
		GoodsThumb      string `json:"goods_thumb"`
		GoodsImg        string `json:"goods_img"`
		GoodsImgWebp    string `json:"goods_img_webp"`
		OriginalImgWebp string `json:"original_img_webp"`
	} `json:"relation_color"`
	CurrentCat struct {
		CatID                  string `json:"cat_id"`
		SiteID                 string `json:"site_id"`
		URLCatID               string `json:"url_cat_id"`
		CatURLName             string `json:"cat_url_name"`
		GoodsTypeID            string `json:"goods_type_id"`
		ShowInNav              string `json:"show_in_nav"`
		Image                  string `json:"image"`
		ImageApp               string `json:"image_app"`
		ParentID               string `json:"parent_id"`
		SortOrder              string `json:"sort_order"`
		IsLeaf                 string `json:"is_leaf"`
		IsReturn               string `json:"is_return"`
		CatName                string `json:"cat_name"`
		MetaTitle              string `json:"meta_title"`
		MetaKeywords           string `json:"meta_keywords"`
		MetaDescription        string `json:"meta_description"`
		CatDesc                string `json:"cat_desc"`
		CategoryDescriptionSeo string `json:"category_description_seo"`
		LeftDescription        string `json:"left_description"`
		LanguageFlag           string `json:"language_flag"`
	} `json:"currentCat"`
	ParentCats struct {
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
		Children []struct {
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
			Children []struct {
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
				Children []struct {
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
					Children []interface{} `json:"children"`
				} `json:"children"`
			} `json:"children"`
		} `json:"children"`
	} `json:"parentCats"`
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
}

var (
	detailReg = regexp.MustCompile(`productIntroData\s*:\s*({.*})`)
	imgUrlReg = regexp.MustCompile(`(?:_thumbnail_\d+x\d*)?\.(jpg|jpeg|png|webp)$`)
)

func newMedia(id, rawurl string, isDefault bool) (*media.Media, error) {
	if strings.HasPrefix(rawurl, "//") {
		rawurl = "https:" + rawurl
	}
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	matched := imgUrlReg.FindStringSubmatch(u.Path)
	if len(matched) == 0 {
		return nil, errors.New("url path not match the regulation expression")
	}
	ext := matched[1]

	return media.NewImageMedia(id,
		rawurl,
		strings.ReplaceAll(rawurl, matched[0], fmt.Sprintf("_thumbnail_%dx.%s", 1000, ext)),
		strings.ReplaceAll(rawurl, matched[0], fmt.Sprintf("_thumbnail_%dx.%s", 600, ext)),
		strings.ReplaceAll(rawurl, matched[0], fmt.Sprintf("_thumbnail_%dx.%s", 500, ext)),
		"",
		isDefault,
	), nil
}

func (c *_Crawler) parseProduct(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error(err)
		return err
	}

	matched := detailReg.FindSubmatch(respBody)
	if len(matched) == 0 {
		c.httpClient.Jar().Clear(ctx, resp.Request.URL)
		return fmt.Errorf("extract produt json from page %s content failed", resp.Request.URL)
	}

	c.logger.Debugf("%s", matched[1])

	var viewData parseProductResponse
	if err = json.Unmarshal(bytes.TrimRight(matched[1], ","), &viewData); err != nil {
		c.logger.Error(err)
		return err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	canUrl, _ := c.CanonicalUrl(doc.Find(`link[rel="canonical"]`).AttrOr("href", ""))
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}
	review, _ := strconv.ParseInt(viewData.Detail.Comment.CommentNum)
	rating, _ := strconv.ParseFloat(viewData.Detail.Comment.CommentRank)

	var (
		colorSpec *pbItem.SkuSpecOption
		desc      = ""
	)
	for _, items := range viewData.Detail.ProductDetails {
		desc = strings.Join([]string{desc, items.AttrNameEn, ":", items.AttrValueEn, ","}, "")
		if items.AttrNameEn == "Color" {
			colorSpec = &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    strconv.Format(items.AttrValueID),
				Name:  items.AttrValueEn,
				Value: items.AttrValueEn,
			}
		}
	}
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           strconv.Format(viewData.Detail.GoodsID),
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
			GroupId:      viewData.Detail.ProductRelationID, // TODO: check if this relation is an group id
		},
		Title:       viewData.Detail.GoodsName,
		Description: desc,
		BrandName:   viewData.Detail.Brand,
		CrowdType:   viewData.ParentCats.CatName,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stats: &pbItem.Stats{
			ReviewCount: int32(review),
			Rating:      float32(rating),
		},
	}

	item.Category = viewData.ParentCats.CatName
	if len(viewData.ParentCats.Children) > 0 {
		item.SubCategory = viewData.ParentCats.Children[0].CatName
	}
	if len(viewData.ParentCats.Children[0].Children) > 0 {
		item.SubCategory2 = viewData.ParentCats.Children[0].Children[0].CatName
	}
	if len(viewData.ParentCats.Children[0].Children[0].Children) > 0 {
		item.SubCategory3 = viewData.ParentCats.Children[0].Children[0].Children[0].CatName
	}

	var medias []*media.Media
	if m, err := newMedia("", viewData.GoodsImgs.MainImage.OriginImage, true); err != nil {
		c.logger.Error(err)
		return err
	} else {
		medias = append(medias, m)
	}
	for _, img := range viewData.GoodsImgs.DetailImage {
		if m, err := newMedia("", img.OriginImage, false); err != nil {
			c.logger.Error(err)
			return err
		} else {
			medias = append(medias, m)
		}
	}
	item.Medias = medias

	for _, rawSize := range viewData.AttrSizeList {
		discount, _ := strconv.ParseInt(rawSize.Price.UnitDiscount)
		current, _ := strconv.ParseFloat(rawSize.Price.SalePrice.UsdAmount)
		msrp, _ := strconv.ParseFloat(rawSize.Price.RetailPrice.UsdAmount)
		if current == 0 {
			current = msrp
		}

		sku := pbItem.Sku{
			SourceId: strconv.Format(rawSize.AttrValueID),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(current * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},
			Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}
		sku.Medias = medias

		if rawSize.Stock > 0 {
			sku.Stock.StockStatus = pbItem.Stock_InStock
			sku.Stock.StockCount = int32(rawSize.Stock)
		}
		if colorSpec != nil {
			sku.Specs = append(sku.Specs, colorSpec)
		}
		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Id:    strconv.Format(rawSize.AttrValueID),
			Name:  rawSize.AttrValueEn,
			Value: rawSize.AttrValueEn,
		})
		item.SkuItems = append(item.SkuItems, &sku)
	}

	if err = yield(ctx, &item); err != nil {
		return err
	}
	return nil
}

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		// "https://us.romwe.com/style/Tee-Dresses-sc-00102523.html?ici=us_tab01navbar02menu04dir02&scici=navbar_GirlsHomePage~~tab01navbar02menu04dir02~~2_4_2~~itemPicking_00102523~~~~0~~50001&srctype=category&userpath=category%3EClothing%3EDresses%3ETee%20Dresses",
		"https://us.romwe.com/Ditsy-Floral-Cami-Dress-p-590009-cat-767.html?scici=navbar_GirlsHomePage~~tab01navbar02menu04dir03~~2_4_3~~itemPicking_00116595~~~~0~~50001",
		//"https://us.romwe.com/Drop-Shoulder-Heather-Gray-Tee-Dress-Without-Belt-p-789904-cat-767.html?scici=navbar_GirlsHomePage~~tab01navbar02menu04dir02~~2_4_2~~itemPicking_00102523~~~~0~~50001",
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
