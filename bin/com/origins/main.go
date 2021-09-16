package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/voiladev/VoilaCrawler/pkg/cli"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	pbMedia "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
)

// _Crawler defined the crawler struct/class for which is not necessory to be exportable
type _Crawler struct {
	crawler.MustImplementCrawler

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
//func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
func (_ *_Crawler) New(_ *cli.Context, client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient: client,
		// this regular used to match category page url path
		categoryPathMatcher: regexp.MustCompile(`^(/([/A-Za-z0-9_-]+))$`),
		productPathMatcher:  regexp.MustCompile(`^/product/[/A-Za-z0-9_-]+$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "60d0fbbcc81649a893936447abb855b6"
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
	opts := &crawler.CrawlOptions{
		EnableHeadless: false,
		// use js api to init session for the first request of the crawl
		EnableSessionInit: false,
		Reliability:       proxy.ProxyReliability_ReliabilityDefault,
	}

	return opts
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.origins.com"}
}

// CanonicalUrl
func (c *_Crawler) CanonicalUrl(rawurl string) (string, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	if u.Host == "" {
		u.Host = "www.origins.com"
	}
	if c.productPathMatcher.MatchString(u.Path) {
		u.RawQuery = ""
	}
	return u.String(), nil
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

	p := strings.TrimSuffix(resp.Request.URL.Path, "/")
	if p == "" {
		return crawler.ErrUnsupportedPath
	}

	if c.productPathMatcher.MatchString(p) {
		return c.parseProduct(ctx, resp, yield)
	} else if c.categoryPathMatcher.MatchString(p) {
		return c.parseCategoryProducts(ctx, resp, yield)
	}

	return crawler.ErrUnsupportedPath
}

// nextIndex used to get the index from the shared data.
// item.index is a const key for item index.
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")))
}
func (c *_Crawler) GetCategories(ctx context.Context) ([]*pbItem.Category, error) {
	req, _ := http.NewRequest(http.MethodGet, "https://www.origins.com/", nil)
	opts := c.CrawlOptions(req.URL)
	resp, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
		EnableProxy:       true,
		EnableHeadless:    opts.EnableHeadless,
		EnableSessionInit: opts.EnableSessionInit,
		Reliability:       opts.Reliability,
	})

	if err != nil {
		c.logger.Error(err)
		return nil, err
	}

	dom, err := resp.Selector()
	if err != nil {
		c.logger.Error(err)
		return nil, err
	}

	var (
		cates   []*pbItem.Category
		cateMap = map[string]*pbItem.Category{}
	)
	if err := func(yield func(names []string, url string) error) error {

		sel := dom.Find(`.page-header__nav__inner .gnav-menu-item.gnav-menu-item--has-children`)

		for i := range sel.Nodes {
			node := sel.Eq(i)
			cateName := strings.TrimSpace(node.Find(`.gnav-menu-item__title-wrap`).First().Text())

			if cateName == "" || strings.ToLower(cateName) == "skin services" || strings.ToLower(cateName) == "about us" {
				continue
			}

			subSel := node.Find(`.js-gnav-formatter--v1.js-gnav-loyalty__formatter.gnav-formatter.gnav-formatter-v1.gnav-formatter--lvl-3`)
			if len(subSel.Nodes) == 0 {
				subSel = node.Find(`.gnav-formatter__list.gnav-formatter__list--lvl-3.js-gnav-formatter-lvl.js-gnav-formatter--lvl-3`)
			}
			for k := range subSel.Nodes {
				subNode2 := subSel.Eq(k)

				subcat2 := strings.TrimSpace(subNode2.Find(`.gnav-menu-label`).First().Text())
				if subcat2 == "" {
					subcat2 = strings.TrimSpace(subNode2.Find(`span`).Last().Text())
				}

				subNode2list := subNode2.Find(`.gnav-menu-link`)
				for j := range subNode2list.Nodes {
					subNode3 := subNode2list.Eq(j)
					subcat3 := strings.TrimSpace(subNode3.Find(`.gnav-menu-link__item`).First().Text())

					href, err := c.CanonicalUrl(subNode3.Find(`a`).AttrOr("href", ""))
					if href == "" || subcat3 == "" || err != nil || strings.ToLower(subcat3) == "gift cards" || strings.ToLower(subcat3) == "build a custom set" {
						continue
					}

					u, err := url.Parse(href)
					if err != nil {
						c.logger.Error("parse url %s failed", href)
						continue
					}

					if c.categoryPathMatcher.MatchString(u.Path) {
						if err := yield([]string{cateName, subcat2, subcat3}, href); err != nil {
							return err
						}
					}
				}

				if len(subNode2list.Nodes) == 0 {
					href, err := c.CanonicalUrl(subNode2.Find(`a`).First().AttrOr("href", ""))
					if href == "" || err != nil {
						continue
					}

					u, err := url.Parse(href)
					if err != nil {
						c.logger.Error("parse url %s failed", href)
						continue
					}

					if c.categoryPathMatcher.MatchString(u.Path) {
						if err := yield([]string{cateName, subcat2}, href); err != nil {
							return err
						}
					}
				}
			}
		}

		return nil
	}(func(names []string, url string) error {
		if len(names) == 0 {
			return errors.New("no valid category name found")
		}

		var (
			lastCate *pbItem.Category
			path     string
		)

		for i, name := range names {
			path = strings.Join([]string{path, name}, "-")

			name = strings.Title(strings.ToLower(name))
			if cate, _ := cateMap[path]; cate != nil {
				lastCate = cate
				continue
			} else {
				cate = &pbItem.Category{
					Name: name,
				}
				cateMap[path] = cate
				if lastCate != nil {
					lastCate.Children = append(lastCate.Children, cate)
				}
				lastCate = cate

				if i == 0 {
					cates = append(cates, cate)
				}
			}
		}
		lastCate.Url = url
		return nil
	}); err != nil {
		c.logger.Error(err)
		return nil, err
	}
	return cates, nil
}

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

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	lastIndex := nextIndex(ctx)

	sel := doc.Find(`.mpp__content`).Find(`.product-brief__title`)
	if len(sel.Nodes) == 0 {
		sel = doc.Find(`.basic-noderef`)
	}

	for i := range sel.Nodes {
		node := sel.Eq(i)

		if href, err := c.CanonicalUrl(node.Find(`a`).AttrOr("href", ``)); err == nil && href != "" {
			if c.productPathMatcher.MatchString(href) {
				req, err := http.NewRequest(http.MethodGet, href, nil)
				if err != nil {
					c.logger.Error(err)
					continue
				}
				lastIndex += 1
				nctx := context.WithValue(ctx, "item.index", lastIndex)
				if err := yield(nctx, req); err != nil {
					return err
				}
			}
		}
	}

	//Note: ProductCount > display on page
	//Next page link not found.
	//Tillnow all records areavailble in 1 link.

	// req, _ := http.NewRequest(http.MethodGet, nextUrl, nil)
	// nctx := context.WithValue(ctx, "item.index", lastIndex)
	// return yield(nctx, req)
	return nil
}

func TrimSpaceNewlineInString(s []byte) []byte {
	re := regexp.MustCompile(`\n`)
	resp := re.ReplaceAll(s, []byte(" "))
	resp = bytes.ReplaceAll(resp, []byte("\\n"), []byte(""))
	resp = bytes.ReplaceAll(resp, []byte("\r"), []byte(""))
	resp = bytes.ReplaceAll(resp, []byte("\t"), []byte(""))
	resp = bytes.ReplaceAll(resp, []byte("&lt;"), []byte("<"))
	resp = bytes.ReplaceAll(resp, []byte("&gt;"), []byte(">"))
	resp = bytes.ReplaceAll(resp, []byte("  "), []byte(""))
	return resp
}

var productsDetailsReg = regexp.MustCompile(`(?Ums)<script type="application/json"\s*id="page_data">\s*({.*})\s*</script>`)

var productsReviewExtractReg = regexp.MustCompile(`(?Ums)<script type="application/ld\+json">\s*({.*})\s*</script>`)

// used to trim html labels in description
var htmlTrimRegp = regexp.MustCompile(`</?[^>]+>`)

type parseProductResponse struct {
	AnalyticsDatalayer struct {
		ProductCategoryName []string `json:"product_category_name"`
		ProductName         []string `json:"product_name"`
		ProductID           []string `json:"product_id"`
	} `json:"analytics-datalayer"`
	CatalogSpp struct {
		Products []struct {
			Path                   string      `json:"path"`
			Description            string      `json:"DESCRIPTION"`
			Feature                interface{} `json:"FEATURE"`
			MiscFlag               int         `json:"MISC_FLAG"`
			ContentVideoTitle      interface{} `json:"CONTENT_VIDEO_TITLE"`
			FamilyCode             string      `json:"FAMILY_CODE"`
			VideoSource1           string      `json:"VIDEO_SOURCE_1"`
			ImageMV2               []string    `json:"IMAGE_M_V2"`
			RatingRange            interface{} `json:"RATING_RANGE"`
			MetaDescription        string      `json:"META_DESCRIPTION"`
			ImageSV2               []string    `json:"IMAGE_S_V2"`
			ShopAllFilter          string      `json:"SHOP_ALL_FILTER"`
			Video1File             interface{} `json:"VIDEO_1_FILE"`
			VideoPoster            []string    `json:"VIDEO_POSTER"`
			ProdSkinTypeText       string      `json:"PROD_SKIN_TYPE_TEXT"`
			SkinConcernAttr        interface{} `json:"SKIN_CONCERN_ATTR"`
			ContentVideoPoster     string      `json:"CONTENT_VIDEO_POSTER"`
			Video2Name             interface{} `json:"VIDEO_2_NAME"`
			ImageLV2               []string    `json:"IMAGE_L_V2"`
			WorksWith              []string    `json:"worksWith"`
			Shaded                 int         `json:"shaded"`
			DefaultPath            string      `json:"defaultPath"`
			VideoSource2           string      `json:"VIDEO_SOURCE_2"`
			VideoSource3           string      `json:"VIDEO_SOURCE_3"`
			FilterFinish           interface{} `json:"FILTER_FINISH"`
			ProdCatDisplayOrder    int         `json:"PROD_CAT_DISPLAY_ORDER"`
			ImageM                 []string    `json:"IMAGE_M"`
			ImageL                 []string    `json:"IMAGE_L"`
			Formula                interface{} `json:"FORMULA"`
			TestimonialsProduct    interface{} `json:"TESTIMONIALS_PRODUCT"`
			VideoPoster1           string      `json:"VIDEO_POSTER_1"`
			VideoPoster2           string      `json:"VIDEO_POSTER_2"`
			VideoPoster3           string      `json:"VIDEO_POSTER_3"`
			ImageXl                []string    `json:"IMAGE_XL"`
			SkinConcernLabel       string      `json:"SKIN_CONCERN_LABEL"`
			TestimonialsProdSource interface{} `json:"TESTIMONIALS_PROD_SOURCE"`
			AttributeSkintype      interface{} `json:"ATTRIBUTE_SKINTYPE"`
			RatingImage            string      `json:"RATING_IMAGE"`
			KeyIngredient          interface{} `json:"KEY_INGREDIENT"`
			WhyItIsDifferent       interface{} `json:"WHY_IT_IS_DIFFERENT"`
			Sized                  int         `json:"sized"`
			Finish                 interface{} `json:"FINISH"`
			ProdCatImageName       string      `json:"PROD_CAT_IMAGE_NAME"`
			Video1Title            interface{} `json:"VIDEO_1_TITLE"`
			ProductID              string      `json:"PRODUCT_ID"`
			RecommendedCount       interface{} `json:"RECOMMENDED_COUNT"`

			Skus []struct {
				PRODUCTSIZE       string   `json:"PRODUCT_SIZE"`
				PRICE2            float64  `json:"PRICE2"`
				SKUID             string   `json:"SKU_ID"`
				IsOutOfStock      int      `json:"isOutOfStock"`
				LARGEIMAGEV2      []string `json:"LARGE_IMAGE_V2"`
				PRICE             float64  `json:"PRICE"`
				SHADENAME         string   `json:"SHADENAME"`
				PRODUCTID         string   `json:"PRODUCT_ID"`
				UPCCODE           string   `json:"UPC_CODE"`
				IsPreOrder        int      `json:"isPreOrder"`
				RsSkuAvailability int      `json:"rs_sku_availability"`
			} `json:"skus"`
			PRODRGNSUBHEADING string `json:"PROD_RGN_SUBHEADING"`
		} `json:"products"`
	} `json:"catalog-spp"`
}

type parseProductReview struct {
	AggregateRating struct {
		Type        string  `json:"@type"`
		RatingValue float64 `json:"ratingValue"`
		ReviewCount int     `json:"reviewCount"`
	} `json:"aggregateRating"`
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

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	var viewData parseProductResponse
	matched := productsDetailsReg.FindSubmatch([]byte(respBody))
	if len(matched) > 1 {
		if err := json.Unmarshal(matched[1], &viewData); err != nil {
			c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
			return err
		}
	}

	var viewDatavariant parseProductReview
	matched1 := productsReviewExtractReg.FindSubmatch([]byte(respBody))
	if len(matched) > 1 {
		if err := json.Unmarshal(matched1[1], &viewDatavariant); err != nil {
			c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
			return err
		}
	}

	reviewCount := int64(0)
	rating := float64(0)

	reviewCount, _ = strconv.ParseInt(viewDatavariant.AggregateRating.ReviewCount)
	rating, _ = strconv.ParseFloat(viewDatavariant.AggregateRating.RatingValue)

	canUrl, _ := c.CanonicalUrl(doc.Find(`link[rel="canonical"]`).AttrOr("href", ""))
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}

	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           viewData.AnalyticsDatalayer.ProductID[0],
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
			//GroupId:      doc.Find(`meta[property="product:age_group"]`).AttrOr(`content`, ``),
		},
		BrandName:   "Origins",
		Title:       viewData.AnalyticsDatalayer.ProductName[0],
		Description: htmlTrimRegp.ReplaceAllString(viewData.CatalogSpp.Products[0].Description, ``),
		Category:    viewData.AnalyticsDatalayer.ProductCategoryName[0],
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stats: &pbItem.Stats{
			ReviewCount: int32(reviewCount),
			Rating:      float32(rating),
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
	}

	for _, itema := range viewData.CatalogSpp.Products {
		var videos []*pbMedia.Media
		mediaCounter := -1
		if itema.VideoSource1 != "" {
			mediaCounter++
			videos = append(videos, pbMedia.NewVideoMedia(
				strconv.Format(mediaCounter),
				"",
				"https://www.youtube.com/embed/"+itema.VideoSource1+"?autoplay=1",
				0, 0, 0, "https://www.origins.com"+itema.VideoPoster1, "",
				false))
		}
		if itema.VideoSource2 != "" {
			mediaCounter++
			videos = append(videos, pbMedia.NewVideoMedia(
				strconv.Format(mediaCounter),
				"",
				"https://www.youtube.com/embed/"+itema.VideoSource2+"?autoplay=1",
				0, 0, 0, "https://www.origins.com"+itema.VideoPoster2, "",
				false))
		}
		if itema.VideoSource3 != "" {
			mediaCounter++
			videos = append(videos, pbMedia.NewVideoMedia(
				strconv.Format(mediaCounter),
				"",
				"https://www.youtube.com/embed/"+itema.VideoSource3+"?autoplay=1",
				0, 0, 0, "https://www.origins.com"+itema.VideoPoster3, "",
				false))
		}

		for _, rawSku := range itema.Skus {

			currentPrice, _ := strconv.ParsePrice(rawSku.PRICE)
			msrp, _ := strconv.ParsePrice(rawSku.PRICE2)

			if msrp == 0 {
				msrp = currentPrice
			}
			discount := int32(0)
			if msrp > currentPrice {
				discount = int32(((msrp - currentPrice) / msrp) * 100)
			}

			sku := pbItem.Sku{
				//	SourceId: fmt.Sprintf(rawSku.SKUBASEID, i),
				SourceId: rawSku.SKUID,
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(currentPrice * 100),
					Msrp:     int32(msrp * 100),
					Discount: int32(discount),
				},

				Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}

			//images
			for j, img := range rawSku.LARGEIMAGEV2 {
				mediaCounter++
				imgurl, err := c.CanonicalUrl(img)
				if err != nil {
					continue
				}

				sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
					strconv.Format(mediaCounter),
					imgurl,
					imgurl, imgurl+"?v=1", imgurl+"?v=1",
					// strings.ReplaceAll(imgurl, `1000x1000`, "600x600_gray"),
					// strings.ReplaceAll(imgurl, `1000x1000`, "100x100"),
					"", j == 0))
			}
			sku.Medias = append(sku.Medias, videos...)

			if rawSku.RsSkuAvailability > 0 {
				item.Stock.StockStatus = pbItem.Stock_InStock
				sku.Stock.StockStatus = pbItem.Stock_InStock
			}

			// color
			if rawSku.SHADENAME != "" {
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecColor,
					Id:    rawSku.SHADENAME,
					Name:  rawSku.SHADENAME,
					Value: rawSku.SHADENAME,
				})
			}

			// size
			if rawSku.PRODUCTSIZE != "" {
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecSize,
					Id:    rawSku.PRODUCTSIZE,
					Name:  rawSku.PRODUCTSIZE,
					Value: rawSku.PRODUCTSIZE,
				})
			}

			subTitle := itema.PRODRGNSUBHEADING
			if subTitle == "" {
				subTitle = "-"
			}
			if rawSku.PRODUCTSIZE == "" && rawSku.SHADENAME == "" {
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecColor,
					Id:    subTitle,
					Name:  subTitle,
					Value: subTitle,
				})
			}

			item.SkuItems = append(item.SkuItems, &sku)
		}
	}

	// yield item result
	return yield(ctx, &item)
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://www.origins.com/products/15352/skincare/moisturize/moisturizers",
		//"https://www.origins.com/products/15332/bath-body",
		//"https://www.origins.com/whats-new",
		//"https://www.origins.com/product/15370/68585/makeup/face-makeup/foundation/pretty-in-bloom/flower-infused-long-wear-foundation-spf20",
		//"https://www.origins.com/product/15348/90674/skincare/treat/eye-care/ginzing/vitamin-c-niacinamide-eye-cream-to-brighten-and-depuff",
		//	"https://www.origins.com/product/15346/66858/skincare/treat/mask/glow-co-nuts/hydrating-coconut-mask",
		//"https://www.origins.com/product/15347/84612/skincare/treat/serums/plantscription/multi-powered-youth-serum",
		//"https://www.origins.com/product/15352/39259/skincare/moisturize/moisturizers/dr-andrew-weil-for-origins/mega-defense-barrier-boosting-essence-oil#/sku/68083",
		//"https://www.origins.com/product/15352/23831/skincare/moisturize/moisturizers/make-a-difference-plus/rejuvenating-moisturizer",
		//"https://www.origins.com/product/15346/62427/skincare/treat/mask/clear-improvement/active-charcoal-mask-to-clear-pores#/sku/98640",
		"https://www.origins.com/best-beauty-gifts",
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
	cli.NewApp(&_Crawler{}).Run(os.Args)
}
