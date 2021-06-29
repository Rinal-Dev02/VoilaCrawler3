package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	pbMedia "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"

	// pbProxy "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
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
		categoryPathMatcher: regexp.MustCompile(`^/shop(/[a-z0-9\pL\pS._\-]+){2,6}$`),
		productPathMatcher:  regexp.MustCompile(`^/shop/product/[a-z0-9\pL\pS._\-]+$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "f330e29cf7fb7dc313fd101fde1d5aa5"
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
	//options.DisableCookieJar = true
	options.Reliability = proxy.ProxyReliability_ReliabilityDefault
	//options.MustHeader.Set("accept-encoding", "gzip, deflate, br")
	options.MustHeader.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	options.MustHeader.Set("accept-language", "en-GB,en-US;q=0.9,en;q=0.8")

	options.MustCookies = append(options.MustCookies,
		&http.Cookie{Name: "currency", Value: `USD`, Path: "/"},
		&http.Cookie{Name: "shippingCountry", Value: `US`, Path: "/"},
		&http.Cookie{Name: "mercury", Value: `true`, Path: "/"},
	)

	if u != nil {
		// options.MustCookies = append(options.MustCookies, &http.Cookie{
		// 	Name:  "FORWARDPAGE_KEY",
		// 	Value: url.QueryEscape(u.String()),
		// })
		// options.MustHeader.Set("Referer", u.String())
	}
	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.bloomingdales.com"}
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
		u.Host = "www.bloomingdales.com"
	}

	if c.productPathMatcher.MatchString(u.Path) {
		u.RawQuery = fmt.Sprintf("ID=%s", u.Query().Get("ID"))
		return u.String(), nil
	}
	return u.String(), nil
}

func (c *_Crawler) Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}
	p := strings.TrimSuffix(resp.Request.URL.Path, "/")

	if p == "/index" || p == "" {
		return c.parseCategories(ctx, resp, yield)
	}
	if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yield)
	} else if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	}
	return crawler.ErrUnsupportedPath
}

func (c *_Crawler) parseCategories(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	catUrl := "https://www.bloomingdales.com/xapi/navigate/v1/header?bypass_redirect=yes&viewType=Responsive&currencyCode=USD&_regionCode=US&_navigationType=BROWSE&_shoppingMode=SITE"
	req, err := http.NewRequest(http.MethodGet, catUrl, nil)
	req.Header.Add("accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Add("referer", "https://www.bloomingdales.com/")
	req.Header.Add("accept-language", "en-GB,en-US;q=0.9,en;q=0.8")
	req.Header.Add("x-requested-with", "XMLHttpRequest")

	catreq, err := c.httpClient.Do(ctx, req)
	if err != nil {
		panic(err)
	}
	defer catreq.Body.Close()

	catBody, err := ioutil.ReadAll(catreq.Body)
	if err != nil {
		c.logger.Error(err)
		return err
	}

	var viewData categoryStructure
	if err := json.Unmarshal(catBody, &viewData); err != nil {
		c.logger.Errorf("unmarshal cat detail data fialed, error=%s", err)
		return err
	}

	for _, rawcat := range viewData.Menu {

		nnctx := context.WithValue(ctx, "Category", rawcat.Text)

		for _, rawsubcat := range rawcat.Children[0].Group {

			if len(rawsubcat.Children) > 0 {

				for _, rawsub2cat := range rawsubcat.Children[0].Group {

					href := rawsub2cat.URL
					if href == "" {
						continue
					}

					fmt.Println(rawsubcat.Text + " > " + rawsub2cat.Text)
					u, err := url.Parse(href)
					if err != nil {
						c.logger.Errorf("parse url %s failed", href)
						continue
					}

					if c.categoryPathMatcher.MatchString(u.Path) {
						nctx := context.WithValue(nnctx, "SubCategory", rawsubcat.Text+" > "+rawsub2cat.Text)
						req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
						if err := yield(nctx, req); err != nil {
							return err
						}
					}
				}
			} else {

				href := rawsubcat.URL
				if href == "" {
					continue
				}

				fmt.Println(rawsubcat.Text)
				u, err := url.Parse(href)
				if err != nil {
					c.logger.Errorf("parse url %s failed", href)
					continue
				}

				if c.categoryPathMatcher.MatchString(u.Path) {
					nctx := context.WithValue(nnctx, "SubCategory", rawsubcat.Text)
					req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
					if err := yield(nctx, req); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

type categoryStructure struct {
	Menu []struct {
		ID       string `json:"id"`
		Text     string `json:"text"`
		URL      string `json:"url"`
		Children []struct {
			Group []struct {
				ID       string `json:"id"`
				Text     string `json:"text"`
				URL      string `json:"url"`
				Children []struct {
					Group []struct {
						ID   string `json:"id"`
						Text string `json:"text"`
						URL  string `json:"url"`
					} `json:"group"`
				} `json:"children"`
			} `json:"group"`
		} `json:"children"`
	} `json:"menu"`
}

// nextIndex used to get sharingData from context
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")) + 1)
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

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	lastIndex := nextIndex(ctx)

	sel := dom.Find(".items>.cell")
	if len(sel.Nodes) == 0 {
		return errors.New("no product found")
	}
	for i := range sel.Nodes {
		node := sel.Eq(i)
		href := node.Find(`.productThumbnail .productDescLink`).AttrOr("href", "")

		req, err := http.NewRequest(http.MethodGet, href, nil)
		if err != nil {
			c.logger.Errorf("load http request of url %s failed, error=%s", href, err)
			return err
		}
		req.Header.Set("Referer", resp.Request.URL.String())
		req.AddCookie(&http.Cookie{Name: "FORWARDPAGE_KEY", Value: url.QueryEscape(resp.Request.URL.String())})
		nctx := context.WithValue(ctx, "item.index", lastIndex)
		lastIndex++

		// yield sub request
		if err := yield(nctx, req); err != nil {
			return err
		}
	}

	pageIndex := dom.Find(`#sort-pagination-select-bottom > option[selected="selected"] + option`).AttrOr("value", "")
	if pageIndex == "" {
		return nil
	}

	u := *resp.Request.URL
	fields := strings.Split(u.Path, "/")
	if len(fields) > 3 && fields[len(fields)-2] == "Pageindex" {
		fields[len(fields)-1] = pageIndex
	} else {
		fields = append(fields, "Pageindex", pageIndex)
	}
	u.Path = strings.Join(fields, "/")

	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	req.Header.Set("Referer", resp.Request.URL.String())
	req.AddCookie(&http.Cookie{Name: "FORWARDPAGE_KEY", Value: url.QueryEscape(resp.Request.URL.String())})

	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
}

type parseProductData struct {
	Product struct {
		ID     int `json:"id"`
		Detail struct {
			Name                  string   `json:"name"`
			Description           string   `json:"description"`
			SecondaryDescription  string   `json:"secondaryDescription"`
			BulletText            []string `json:"bulletText"`
			MaterialsAndCare      []string `json:"materialsAndCare"`
			MaxQuantity           int      `json:"maxQuantity"`
			TypeName              string   `json:"typeName"`
			AdditionalImagesCount int      `json:"additionalImagesCount"`
			NumberOfColors        int      `json:"numberOfColors"`
			Brand                 struct {
				Name          string `json:"name"`
				ID            int    `json:"id"`
				URL           string `json:"url"`
				SubBrand      string `json:"subBrand"`
				BrandBreakout bool   `json:"brandBreakout"`
			} `json:"brand"`
			ReviewTitle string `json:"reviewTitle"`
		} `json:"detail"`
		Relationships struct {
			Taxonomy struct {
				Categories []struct {
					Name string `json:"name"`
					URL  string `json:"url"`
					ID   int    `json:"id"`
				} `json:"categories"`
				DefaultCategoryID int `json:"defaultCategoryId"`
			} `json:"taxonomy"`
			Upcs map[string]struct {
				ID         int `json:"id"`
				Identifier struct {
					UpcNumber string `json:"upcNumber"`
				} `json:"identifier"`
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
				ProtectionPlans        []interface{} `json:"protectionPlans"`
				HolidayMessageEligible bool          `json:"holidayMessageEligible"`
			} `json:"upcs"`
		} `json:"relationships"`
		Traits struct {
			Colors struct {
				SelectedColor int `json:"selectedColor"`
				ColorMap      map[string]struct {
					//Num1817529 struct {
					ID          int    `json:"id"`
					Name        string `json:"name"`
					NormalName  string `json:"normalName"`
					SwatchImage struct {
						FilePath             string `json:"filePath"`
						Name                 string `json:"name"`
						ShowJumboSwatch      bool   `json:"showJumboSwatch"`
						SwatchSpriteOffset   int    `json:"swatchSpriteOffset"`
						SwatchSpriteURLIndex int    `json:"swatchSpriteUrlIndex"`
					} `json:"swatchImage"`
					Imagery struct {
						Images []struct {
							FilePath             string `json:"filePath"`
							Name                 string `json:"name"`
							ShowJumboSwatch      bool   `json:"showJumboSwatch"`
							SwatchSpriteOffset   int    `json:"swatchSpriteOffset"`
							SwatchSpriteURLIndex int    `json:"swatchSpriteUrlIndex"`
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
						BadgeIds []string `json:"badgeIds"`
					} `json:"pricing"`
					//} `json:"1817529"`
				} `json:"colorMap"`
			} `json:"colors"`
			Sizes struct {
				OrderedSizesBySeqNumber []int  `json:"orderedSizesBySeqNumber"`
				SizeChartID             string `json:"sizeChartId"`
				SizeMap                 map[string]struct {
					ID          int    `json:"id"`
					Name        string `json:"name"`
					DisplayName string `json:"displayName"`
					Colors      []int  `json:"colors"`
				} `json:"sizeMap"`
			} `json:"sizes"`
		} `json:"traits"`
	} `json:"product"`
	UtagData struct {
		ProductRating  []string `json:"product_rating"`
		ProductReviews []string `json:"product_reviews"`
	} `json:"utagData"`
}

var (
	detailReg = regexp.MustCompile(`(?U)<script\s+data-bootstrap="page/product"\s*type="application/json">({.*})</script>`)
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
		pd parseProductData
	)

	if err = json.Unmarshal(matched[1], &pd); err != nil {
		c.logger.Error(err)
		return err
	}
	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	reviewCount, _ := strconv.ParseFloat(pd.UtagData.ProductReviews[0])
	rating, _ := strconv.ParseFloat(pd.UtagData.ProductRating[0])

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
		Description: pd.Product.Detail.Description + strings.Join(pd.Product.Detail.BulletText, ", "),
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
	for skuId, rawSku := range pd.Product.Relationships.Upcs {
		colorId := strconv.Format(rawSku.Traits.Colors.SelectedColor)
		color := pd.Product.Traits.Colors.ColorMap[colorId]
		sizeId := strconv.Format(rawSku.Traits.Sizes.SelectedSize)
		size := pd.Product.Traits.Sizes.SizeMap[sizeId]

		var (
			current, msrp, discount float64
		)
		for _, p := range color.Pricing.Price.TieredPrice {
			if len(p.Values) == 0 {
				continue
			}
			if current == 0 && msrp == 0 {
				current, msrp = p.Values[0].Value, p.Values[0].Value
			} else if p.Values[0].Value > msrp {
				msrp = p.Values[0].Value
			} else if p.Values[0].Value < current {
				current = p.Values[0].Value
			}
		}
		if msrp == 0 {
			return fmt.Errorf("no msrp price found for %s", resp.Request.URL)
		}
		discount = math.Ceil((msrp - current) * 100 / msrp)

		var medias []*pbMedia.Media
		for key, img := range color.Imagery.Images {
			medias = append(medias, pbMedia.NewImageMedia(
				"",
				"https://images.bloomingdalesassets.com/is/image/BLM/products/"+img.FilePath,
				"https://images.bloomingdalesassets.com/is/image/BLM/products/"+img.FilePath+"?op_sharpen=1&wid=1000",
				"https://images.bloomingdalesassets.com/is/image/BLM/products/"+img.FilePath+"?op_sharpen=1&wid=700",
				"https://images.bloomingdalesassets.com/is/image/BLM/products/"+img.FilePath+"?op_sharpen=1&wid=450",
				"",
				key == 0,
			))
		}

		sku := pbItem.Sku{
			SourceId: skuId,
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(current * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},
			Medias: medias,
			Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}
		item.Medias = append(item.Medias, medias...)
		if rawSku.Availability.Available {
			sku.Stock.StockStatus = pbItem.Stock_InStock
			item.Stock.StockStatus = pbItem.Stock_InStock
			//sku.Stock.StockCount = int32(rawSize.Quantity)
		}

		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecColor,
			Id:    colorId,
			Name:  color.NormalName,
			Value: color.NormalName,
		})
		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Id:    sizeId,
			Name:  size.DisplayName,
			Value: size.Name,
		})
		item.SkuItems = append(item.SkuItems, &sku)
	}

	// yield item result
	if err = yield(ctx, &item); err != nil {
		return err
	}
	return nil
}

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://www.bloomingdales.com/?cm_sp=NAVIGATION-_-TOP_NAV-_-TOP_BLOOMIES_ICON",
		//"https://www.bloomingdales.com/shop/womens-apparel/tops-tees?id=5619",
		// "https://www.bloomingdales.com/shop/product/aqua-passion-sleeveless-maxi-dress-100-exclusive?ID=3996369&CategoryID=21683",
		//"https://www.bloomingdales.com/shop/product/a.l.c.-kati-puff-sleeve-tee?ID=3202505&CategoryID=5619",
		//"https://www.bloomingdales.com/shop/product/simon-miller-auto-club-graphic-oversized-tee?ID=4047681&CategoryID=2910",
		"https://www.bloomingdales.com/shop/product/sunset-spring-embellished-denim-jacket-100-exclusive?ID=4077611&CategoryID=1001940",
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
