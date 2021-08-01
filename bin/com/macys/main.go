package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/voiladev/VoilaCrawler/pkg/cli"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	pbProxy "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
	"google.golang.org/protobuf/types/known/anypb"
)

type _Crawler struct {
	httpClient http.Client

	categoryPathMatcher *regexp.Regexp
	productPathMatcher  *regexp.Regexp

	logger glog.Log
}

func (_ *_Crawler) New(_ *cli.Context, client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient:          client,
		categoryPathMatcher: regexp.MustCompile(`^(/[a-z0-9_\-]+)?/shop(/[a-zA-Z0-9\-]+){1,4}(/Pageindex/\d+)?$`),
		productPathMatcher:  regexp.MustCompile(`^(/[a-z0-9_\-]+)?/shop/product/([/a-z0-9_\-]+)$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "d0ad08367a3a33b6c11f78eace13b421"
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
	options.DisableCookieJar = true
	options.Reliability = pbProxy.ProxyReliability_ReliabilityMedium
	options.MustHeader.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.16; rv:85.0) Gecko/20100101 Firefox/85.0")

	// options.MustCookies = append(options.MustCookies,
	// 	&http.Cookie{Name: "shippingCountry", Value: "US", Path: "/"},
	// 	&http.Cookie{Name: "currency", Value: "USD", Path: "/"},
	// )
	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.macys.com"}
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
		u.Host = "www.macys.com"
	}
	if c.productPathMatcher.MatchString(u.Path) {
		vals := u.Query()
		id := vals.Get("ID")
		if id == "" {
			id = vals.Get("id")
		}
		u.RawQuery = "ID=" + id

		return u.String(), nil
	}
	return u.String(), nil
}
func (c *_Crawler) GetCategories(ctx context.Context) ([]*pbItem.Category, error) {
	req, _ := http.NewRequest(http.MethodGet, "https://www.macys.com", nil)
	opts := c.CrawlOptions(req.URL)
	for k := range opts.MustHeader {
		req.Header.Set(k, opts.MustHeader.Get(k))
	}
	resp, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
		EnableProxy:       true,
		EnableHeadless:    opts.EnableHeadless,
		EnableSessionInit: opts.EnableSessionInit,
		Reliability:       opts.Reliability,
		DisableCookieJar:  opts.DisableCookieJar,
	})
	if err != nil {
		c.logger.Error(err)
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error(err)
		return nil, err
	}

	matched := categoryExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		return nil, fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}

	var viewData categoryStructure
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal category detail data fialed, error=%s", err)
		return nil, err
	}

	var cates []*pbItem.Category
	for _, rawCat := range viewData {
		cateName := rawCat.Text
		if cateName == "" {
			continue
		}
		cate := pbItem.Category{Name: cateName}
		cates = append(cates, &cate)

		for _, rawsubCat := range rawCat.Children {
			for _, rawsubCatGrp := range rawsubCat.Group {
				subCatName := rawsubCatGrp.Text
				subCate := pbItem.Category{Name: subCatName}
				cate.Children = append(cate.Children, &subCate)

				for _, rawsubcatlvl2 := range rawsubCatGrp.Children {
					for _, rawsubcatlvl2Grp := range rawsubcatlvl2.Group {
						subCate2Name := rawsubcatlvl2Grp.Text
						href, _ := c.CanonicalUrl(rawsubcatlvl2Grp.URL)
						if href == "" {
							continue
						}
						subCate2 := pbItem.Category{Name: subCate2Name, Url: href}
						subCate.Children = append(subCate.Children, &subCate2)
					}
				}
			}
		}
	}
	return cates, nil
}

type categoryStructure []struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	URL      string `json:"url"`
	Children []struct {
		Group []struct {
			ID       string `json:"id"`
			Text     string `json:"text"`
			Children []struct {
				Group []struct {
					Text string `json:"text"`
					URL  string `json:"url"`
				} `json:"group"`
			} `json:"children"`
		} `json:"group"`
	} `json:"children"`
}

func (c *_Crawler) Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	if c.productPathMatcher.MatchString(resp.RawUrl().Path) {
		return c.parseProduct(ctx, resp, yield)
	} else if c.categoryPathMatcher.MatchString(resp.RawUrl().Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	}
	return crawler.ErrUnsupportedPath
}

// nextIndex used to get sharingData from context
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")) + 1)
}

var (
	prodDataExtraReg      = regexp.MustCompile(`(data-bootstrap="page/discovery-pages" type="application/json">)([^<]+)</script>`)
	prodDataPaginationReg = regexp.MustCompile(`(data-bootstrap="feature/canvas"  type="application/json">)([^<]+)</script>`)
)

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

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}
	sel := dom.Find(`.items > .productThumbnailItem`)

	lastIndex := nextIndex(ctx)
	for i := range sel.Nodes {
		node := sel.Eq(i)

		detailUrl := node.Find(".productDescription>a").AttrOr("href", "")
		if detailUrl == "" {
			continue
		}
		req, err := http.NewRequest(http.MethodGet, detailUrl, nil)
		if err != nil {
			c.logger.Errorf("invalud product detail url %s", detailUrl)
		}
		req.Header.Set("Referer", resp.Request.URL.String())

		nctx := context.WithValue(ctx, "item.index", lastIndex)
		lastIndex += 1

		if err := yield(nctx, req); err != nil {
			return err
		}
	}

	var pagination struct {
		Row   int `json:"row"`
		Model struct {
			Pagination struct {
				NextURL       string `json:"nextURL"`
				BaseURL       string `json:"baseURL"`
				NumberOfPages int    `json:"numberOfPages"`
				CurrentPage   int    `json:"currentPage"`
			} `json:"pagination"`
		} `json:"model"`
	}
	pRawData := strings.TrimSpace(dom.Find(`script[data-bootstrap="feature/canvas"]`).Text())
	if err := json.Unmarshal([]byte(pRawData), &pagination); err != nil {
		c.logger.Errorf("unmarshal pagination info %s failed, error=%s", respBody, err)
		return err
	}
	if pagination.Model.Pagination.NextURL != "" {
		req, _ := http.NewRequest(http.MethodGet, pagination.Model.Pagination.NextURL, nil)
		req.Header.Set("Referer", resp.Request.URL.String())
		nctx := context.WithValue(ctx, "item.index", lastIndex)
		return yield(nctx, req)
	}
	return nil
}

type parseProductResponse struct {
	Properties struct {
		ASSETHOST string `json:"ASSET_HOST"`
		Recaptcha struct {
			ScriptURL string `json:"scriptUrl"`
			SiteKey   string `json:"siteKey"`
		} `json:"recaptcha"`
	} `json:"properties"`
	ISCLIENTLOGSENABLED bool   `json:"_IS_CLIENT_LOGS_ENABLED"`
	PDPBOOTSTRAPDATA    string `json:"_PDP_BOOTSTRAP_DATA"`
}

type parseProductData struct {
	UtagData struct {
		ProductRating  []string `json:"product_rating"`
		ProductReviews []string `json:"product_reviews"`
	} `json:"utagData"`
	Product struct {
		ID     int `json:"id"`
		Detail struct {
			Name                 string   `json:"name"`
			Description          string   `json:"description"`
			SecondaryDescription string   `json:"secondaryDescription"`
			BulletText           []string `json:"bulletText"`
			Brand                struct {
				Name string `json:"name"`
			} `json:"brand"`
		} `json:"detail"`
		Relationships struct {
			Taxonomy struct {
				Categories []struct {
					Name string `json:"name"`
				} `json:"categories"`
			} `json:"taxonomy"`
			Upcs map[string]struct {
				//Num44742859 struct {
				ID           int `json:"id"`
				Availability struct {
					Available bool `json:"available"`
				} `json:"availability"`
				Traits struct {
					Colors struct {
						SelectedColor int `json:"selectedColor"`
					} `json:"colors"`
					Sizes struct {
						SelectedSize int `json:"selectedSize"`
					} `json:"sizes"`
				} `json:"traits"`
			} `json:"upcs"`
		} `json:"relationships"`
		Imagery struct {
			Images []struct {
				FilePath             string `json:"filePath"`
				Name                 string `json:"name"`
				ShowJumboSwatch      bool   `json:"showJumboSwatch"`
				SwatchSpriteOffset   int    `json:"swatchSpriteOffset"`
				SwatchSpriteURLIndex int    `json:"swatchSpriteUrlIndex"`
			} `json:"images"`
		} `json:"imagery"`
		Traits struct {
			Colors struct {
				SelectedColor int `json:"selectedColor"`
				ColorMap      map[string]struct {
					ID         int    `json:"id"`
					Name       string `json:"name"`
					NormalName string `json:"normalName"`
					Imagery    struct {
						Images []struct {
							FilePath string `json:"filePath"`
							Name     string `json:"name"`
						} `json:"images"`
					} `json:"imagery"`
					Sizes   []int `json:"sizes"`
					Pricing struct {
						Price struct {
							TieredPrice []struct {
								Label  string `json:"label"`
								Values []struct {
									Value          float64 `json:"value"`
									FormattedValue string  `json:"formattedValue"`
									Type           string  `json:"type"`
								} `json:"values"`
							} `json:"tieredPrice"`
							PriceTypeID int `json:"priceTypeId"`
						} `json:"price"`
					} `json:"pricing"`
				} `json:"colorMap"`
			} `json:"colors"`
			Sizes struct {
				SizeMap map[string]struct {
					ID          int    `json:"id"`
					Name        string `json:"name"`
					DisplayName string `json:"displayName"`
					Colors      []int  `json:"colors"`
				} `json:"sizeMap"`
			} `json:"sizes"`
		} `json:"traits"`
	} `json:"product"`
}

var (
	detailReg          = regexp.MustCompile(`(?U)<script[^>]*>\s*window.__INITIAL_STATE__\s*=\s*({.*});?\s*</script>`)
	categoryExtractReg = regexp.MustCompile(`(?U)<script\s*type='application/json'\s*data-mcom-header-menu-desktop='context\.header\.menu'>(\[.*\])\s*</script>`)
)

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
	if len(matched) <= 1 {
		c.logger.Debugf("data %s", respBody)
		return fmt.Errorf("extract produt json from page %s content failed", resp.Request.URL)
	}

	var (
		i  parseProductResponse
		pd parseProductData
	)

	if err = json.Unmarshal(matched[1], &i); err != nil {
		c.logger.Error(err)
		return err
	}

	if err = json.Unmarshal([]byte(i.PDPBOOTSTRAPDATA), &pd); err != nil {
		c.logger.Error(err)
		return err
	}

	var (
		reviewCount int64
		rating      float64
	)

	if len(pd.UtagData.ProductReviews) > 0 {
		reviewCount, _ = strconv.ParseInt(pd.UtagData.ProductReviews[0])
	}
	if len(pd.UtagData.ProductRating) > 0 {
		rating, _ = strconv.ParseFloat(pd.UtagData.ProductRating[0])
	}

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	canUrl := dom.Find(`link[rel="canonical"]`).AttrOr("href", "")
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}

	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           strconv.Format(pd.Product.ID),
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
		},
		Title:       pd.Product.Detail.Name,
		Description: pd.Product.Detail.Description,
		BrandName:   pd.Product.Detail.Brand.Name,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stats: &pbItem.Stats{
			ReviewCount: int32(reviewCount),
			Rating:      float32(rating),
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
	}
	for i, cate := range pd.Product.Relationships.Taxonomy.Categories {
		switch i {
		case 0:
			item.Category = cate.Name
		case 1:
			item.SubCategory = cate.Name
		case 2:
			item.SubCategory2 = cate.Name
		case 3:
			item.SubCategory3 = cate.Name
		case 4:
			item.SubCategory4 = cate.Name
		}
	}

	colors := pd.Product.Traits.Colors
	sizes := pd.Product.Traits.Sizes

	for _, color := range pd.Product.Traits.Colors.ColorMap {
		for i, img := range color.Imagery.Images {
			itemImg, _ := anypb.New(&media.Media_Image{
				OriginalUrl: "https://slimages.macysassets.com/is/image/MCY/products/" + img.FilePath,
				LargeUrl:    "https://slimages.macysassets.com/is/image/MCY/products/" + img.FilePath + "?wid=1230&hei=1500&op_sharpen=1", // $S$, $XXL$
				MediumUrl:   "https://slimages.macysassets.com/is/image/MCY/products/" + img.FilePath + "?wid=640&hei=780&op_sharpen=1",
				SmallUrl:    "https://slimages.macysassets.com/is/image/MCY/products/" + img.FilePath + "?wid=500&hei=609&op_sharpen=1",
			})
			item.Medias = append(item.Medias, &media.Media{
				Detail:    itemImg,
				IsDefault: i == 0,
			})
		}
	}

	for id, p := range pd.Product.Relationships.Upcs {

		sku := pbItem.Sku{
			SourceId: id,
			Price:    &pbItem.Price{Currency: regulation.Currency_USD},
			Stock:    &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}
		if p.Availability.Available {
			sku.Stock.StockStatus = pbItem.Stock_InStock
			item.Stock.StockStatus = pbItem.Stock_InStock
		}

		if color, ok := colors.ColorMap[strconv.Format(p.Traits.Colors.SelectedColor)]; ok {
			// build color sku
			spec := pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    strconv.Format(color.ID),
				Name:  color.NormalName,
				Value: color.Name,
			}
			sku.Specs = append(sku.Specs, &spec)

			current := 0.0
			msrp := color.Pricing.Price.TieredPrice[0].Values[0].Value
			if len(color.Pricing.Price.TieredPrice) > 1 {
				current = color.Pricing.Price.TieredPrice[1].Values[0].Value
			}
			discount := math.Ceil((msrp - current) * 100 / msrp)
			sku.Price.Current = int32(current * 100)
			sku.Price.Msrp = int32(msrp * 100)
			sku.Price.Discount = int32(discount)

			for i, img := range color.Imagery.Images {
				itemImg, _ := anypb.New(&media.Media_Image{
					OriginalUrl: "https://slimages.macysassets.com/is/image/MCY/products/" + img.FilePath,
					LargeUrl:    "https://slimages.macysassets.com/is/image/MCY/products/" + img.FilePath + "?wid=1230&hei=1500&op_sharpen=1", // $S$, $XXL$
					MediumUrl:   "https://slimages.macysassets.com/is/image/MCY/products/" + img.FilePath + "?wid=640&hei=780&op_sharpen=1",
					SmallUrl:    "https://slimages.macysassets.com/is/image/MCY/products/" + img.FilePath + "?wid=500&hei=609&op_sharpen=1",
				})
				sku.Medias = append(sku.Medias, &media.Media{
					Detail:    itemImg,
					IsDefault: i == 0,
				})
			}
		}

		if size, ok := sizes.SizeMap[strconv.Format(p.Traits.Sizes.SelectedSize)]; ok {
			spec := pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    strconv.Format(size.ID),
				Name:  size.DisplayName,
				Value: size.Name,
			}
			sku.Specs = append(sku.Specs, &spec)
		}
		if len(sku.Specs) == 0 {
			return fmt.Errorf("got invalid sku, got no sku spec")
		}
		item.SkuItems = append(item.SkuItems, &sku)
	}

	if err = yield(ctx, &item); err != nil {
		return err
	}
	return nil
}

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://www.macys.com/?lid=glbtopnav_macys_icon-us",
		//"https://www.macys.com/shop/womens-clothing/womens-sale-clearance?id=10066",
		//"https://www.macys.com/shop/product/style-co-ribbed-hoodie-sweater-created-for-macys?ID=11393511&CategoryID=10066",
		//"https://www.macys.com/shop/product/levis-womens-501-cotton-high-rise-denim-shorts?ID=11473203&tdp=cm_app~zMCOM-NAVAPP~xcm_zone~zHP_ZONE_D~xcm_choiceId~zcidM66MOD-42138865-b423-4b86-a737-f411c4941424%40H75%40get%2Binspired%24168342%2411473203~xcm_pos~zPos1~xcm_srcCatID~z28589~xcm_contentId~zContent_12931~xcm_prosSource~zcol~",
		//"https://www.macys.com/shop/product/style-co-mixed-stitch-pointelle-sweater-created-for-macys?ID=11484711&tdp=cm_app~zMCOM-NAVAPP~xcm_zone~zPDP_ZONE_A~xcm_choiceId~zcidM05MSN-59ff27b5-314e-43a3-bf9b-56669b72d87e%40HB2%40Customers%2Balso%2Bshopped%24260%2411484711~xcm_pos~zPos2~xcm_srcCatID~z260",
		"https://www.macys.com/shop/product/levis-high-rise-distressed-denim-shorts?ID=10438767&tdp=cm_app~zMCOM-NAVAPP~xcm_zone~zPDP_ZONE_B~xcm_choiceId~zcidM06MNK-3aa02007-8001-4092-aaf8-0dbd86768014%40HB1%40Customers%2Balso%2Bloved%2428589%2410438767~xcm_pos~zPos1~xcm_srcCatID~z28589",
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
