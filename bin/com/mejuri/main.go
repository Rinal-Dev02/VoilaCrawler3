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
	"github.com/voiladev/VoilaCrawler/pkg/cli"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/media"
	pbMedia "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	pbProxy "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
	"google.golang.org/protobuf/types/known/anypb"
)

type _Crawler struct {
	httpClient                   http.Client
	categorySearchAPIPathMatcher *regexp.Regexp
	categoryPathMatcher          *regexp.Regexp
	productPathMatcher           *regexp.Regexp
	productApiPathMatcher        *regexp.Regexp
	logger                       glog.Log
}

func (*_Crawler) New(_ *cli.Context, client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient:                   client,
		categorySearchAPIPathMatcher: regexp.MustCompile(`^/shop/t/([/A-Za-z0-9_-]+)$`),
		categoryPathMatcher:          regexp.MustCompile(`^/([/A-Za-z_-]+)$`),
		productPathMatcher:           regexp.MustCompile(`^/shop/products([/A-Za-z0-9_-]+)$`),
		productApiPathMatcher:        regexp.MustCompile(`^/api/v2/products([/A-Za-z0-9_-]+)$`),
		logger:                       logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "cdd3d26d97e1fae1e1f1d358f83b3114"
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

	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"mejuri.com", "*.mejuri.com"}
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
		u.Host = "www.mejuri.com"
	}
	if c.productPathMatcher.MatchString(u.Path) {
		u.RawQuery = ""
	}
	return u.String(), nil
}

func (c *_Crawler) Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	p := strings.TrimSuffix(resp.RawUrl().Path, "/")
	if p == "" {
		return crawler.ErrUnsupportedPath
	}
	if c.productPathMatcher.MatchString(p) || c.productApiPathMatcher.MatchString(p) {
		return c.parseProduct(ctx, resp, yield)
	} else if c.categoryPathMatcher.MatchString(p) {
		return c.parseCategoryProducts(ctx, resp, yield)
	}
	return crawler.ErrUnsupportedPath
}

type categoryStructure struct {
	Props struct {
		PageProps struct {
			PageData struct {
				Header struct {
					ID          string `json:"_id"`
					ContentType struct {
						ID string `json:"_id"`
					} `json:"_contentType"`
					Locale   string `json:"_locale"`
					Text     string `json:"text"`
					Type     string `json:"type"`
					Children []struct {
						ID          string `json:"_id"`
						ContentType struct {
							ID string `json:"_id"`
						} `json:"_contentType"`
						Locale   string `json:"_locale"`
						Text     string `json:"text"`
						URL      string `json:"url"`
						Children []struct {
							ID          string `json:"_id"`
							ContentType struct {
								ID string `json:"_id"`
							} `json:"_contentType"`
							Locale   string `json:"_locale"`
							Text     string `json:"text"`
							Type     string `json:"type,omitempty"`
							Children []struct {
								ID          string `json:"_id"`
								ContentType struct {
									ID string `json:"_id"`
								} `json:"_contentType"`
								Locale string `json:"_locale"`
								Text   string `json:"text"`
								URL    string `json:"url"`
								Slug   string `json:"slug,omitempty"`
							} `json:"children"`
							Slug string `json:"slug,omitempty"`
							Pos  bool   `json:"pos,omitempty"`
						} `json:"children,omitempty"`
						Type string `json:"type,omitempty"`
					} `json:"children"`
					Slug string `json:"slug"`
				} `json:"header"`
			} `json:"pageData"`
		} `json:"pageProps"`
	} `json:"props"`
}

var productsExtractReg = regexp.MustCompile(`(?U)id="__NEXT_DATA__"\s*type="application/json">\s*({.*})\s*</script>`)

func (c *_Crawler) GetCategories(ctx context.Context) ([]*pbItem.Category, error) {
	req, _ := http.NewRequest(http.MethodGet, "https://www.mejuri.com/", nil)
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
	defer resp.Body.Close()

	catBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error(err)
		return nil, err
	}

	matched := productsExtractReg.FindSubmatch(catBody)
	if len(matched) <= 1 {
		c.logger.Debugf("%s", catBody)
		return nil, err
	}

	var viewData categoryStructure
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal cat detail data fialed, error=%s", err)
		return nil, err
	}

	var (
		cates   []*pbItem.Category
		cateMap = map[string]*pbItem.Category{}
	)
	if err := func(yield func(names []string, url string) error) error {

		for _, rawCat := range viewData.Props.PageProps.PageData.Header.Children {

			catname := rawCat.Text
			if catname == "" {
				continue
			}
			if catname == "About" || catname == "Style Edit" || strings.ToLower(catname) == "our commitment" {
				continue
			}

			for _, rawsub1Cat := range rawCat.Children {

				subcat2 := rawsub1Cat.Text

				for _, rawsub2Cat := range rawsub1Cat.Children {
					subcat3 := rawsub2Cat.Text
					if subcat3 == "" {
						continue
					}

					if subcat3 == "Mejuri Icons" || subcat3 == "Shop Insta" || subcat3 == "Discover the Icons" || subcat3 == "Gift Cards" || subcat3 == "Goop X Mejuri" {
						continue
					}

					href := rawsub2Cat.URL
					if href == "" {
						continue
					}

					canonicalurl, err := c.CanonicalUrl(href)
					if err != nil {
						continue
					}

					u, err := url.Parse(canonicalurl)
					if err != nil {
						c.logger.Error("parse url %s failed", canonicalurl)
						continue
					}

					if !c.productPathMatcher.MatchString(u.Path) && c.categoryPathMatcher.MatchString(u.Path) {
						if err := yield([]string{rawCat.Text, subcat2, subcat3}, canonicalurl); err != nil {
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

// nextIndex used to get sharingData from context
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")) + 1)
}

type parseProductListStructure []struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Products []struct {
		ID                   int    `json:"id"`
		Slug                 string `json:"slug"`
		Name                 string `json:"name"`
		AccurateName         string `json:"accurate_name"`
		MaterialName         string `json:"material_name"`
		MaterialCategoryIcon string `json:"material_category_icon"`
		MaterialStones       []struct {
			ID        int    `json:"id"`
			Name      string `json:"name"`
			Permalink string `json:"permalink"`
		} `json:"material_stones"`

		CreatedAt         string        `json:"created_at"`
		MaterialGroupName string        `json:"material_group_name"`
		Variants          []interface{} `json:"variants"`
		ProductLabel      string        `json:"product_label"`
	} `json:"products"`
}

type parseSessionResp struct {
	Current struct {
		Country     string      `json:"country"`
		Csrf        string      `json:"csrf"`
		GaCookie    string      `json:"ga_cookie"`
		SessionID   string      `json:"session_id"`
		TestsConfig []string    `json:"tests_config"`
		Pos         interface{} `json:"pos"`
		Order       struct {
			Number string `json:"number"`
			State  string `json:"state"`
			Token  string `json:"token"`
		} `json:"order"`
		Region struct {
			Name                string   `json:"name"`
			AvailableCurrencies []string `json:"available_currencies"`
		} `json:"region"`
		ExternalCheckout bool `json:"external_checkout"`
	} `json:"current"`
}

func setSessionCtx(ctx context.Context, session *parseSessionResp) context.Context {
	orderNumber := "M114315860ZX"
	spreeOrderToken := "9pIuUXRb0NATvSKC5BkmPg1632395340843"
	if session != nil {
		if session.Current.Order.Number != "" {
			orderNumber = session.Current.Order.Number
		}
		if session.Current.Order.Token != "" {
			spreeOrderToken = session.Current.Order.Token
		}
	}
	ctx = context.WithValue(ctx, "session.orderNumber", orderNumber)
	ctx = context.WithValue(ctx, "session.spreeOrderToken", spreeOrderToken)
	return ctx
}

func getSessionFromCtx(ctx context.Context) *parseSessionResp {
	var resp parseSessionResp
	if orderNumber := ctx.Value("session.orderNumber"); orderNumber != nil {
		resp.Current.Order.Number = orderNumber.(string)
	}
	if resp.Current.Order.Number == "" {
		resp.Current.Order.Number = "M114315860ZX"
	}
	if spreeOrderToken := ctx.Value("session.spreeOrderToken"); spreeOrderToken != nil {
		resp.Current.Order.Token = spreeOrderToken.(string)
	}
	if resp.Current.Order.Token == "" {
		resp.Current.Order.Token = "9pIuUXRb0NATvSKC5BkmPg1632395340843"
	}
	return &resp
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

	var viewData parseProductListStructure
	var materialFilter []string
	if p := resp.RawUrl().Path; c.categorySearchAPIPathMatcher.MatchString(p) {
		categoryURL := "https://mejuri.com/api/v1/taxon/collections-by-categories/USD" + strings.TrimPrefix(p, "/shop/t")
		varResponse, err := c.variationRequest(ctx, categoryURL, resp.Request.URL.String(), nil)
		if err != nil {
			c.logger.Errorf("extract product list %s failed", categoryURL)
			return err
		}
		if err := json.Unmarshal(varResponse, &viewData); err != nil {
			c.logger.Errorf("extract product list %s failed", categoryURL)
			return err
		}

		u := *resp.Request.URL
		vals := u.Query()
		materialFilter = strings.Split(vals.Get("fbm"), `,`)
	}

	// init session
	var sessionData parseSessionResp
	sessionResponse, err := c.variationRequest(ctx, "https://mejuri.com/session_current", resp.Request.URL.String(), nil)
	if err != nil {
		c.logger.Errorf("init session failed")
		return err
	}
	if err := json.Unmarshal(sessionResponse, &sessionData); err != nil {
		c.logger.Errorf("json.Unmarshal session failed, err=%s", err)
		return err
	}
	ctx = setSessionCtx(ctx, &sessionData)

	lastIndex := nextIndex(ctx)
	sel := doc.Find(`.product-name`)
	if len(sel.Nodes) == 0 {
		sel = doc.Find(`a[data-h="product-card-link"]`)
	}
	for i := range sel.Nodes {
		node := sel.Eq(i)
		href, _ := node.Find("a").Attr("href")
		if href == "" {
			href, _ = node.Attr("href")
		}

		if href != "" {
			parseurl, err := c.CanonicalUrl(href)
			if parseurl == "" || err != nil {
				continue
			}

			req, err := http.NewRequest(http.MethodGet, parseurl, nil)
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

	for _, list := range viewData {

		for _, prod := range list.Products {

			if contains(materialFilter, prod.MaterialName) {

				if href, _ := c.CanonicalUrl("/shop/products/" + prod.Slug); href != "" {
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
	}
	// check if this is the last page
	totalResults, _ := strconv.ParseInt(strings.Split(doc.Find(`.collections-products__products-amount`).Text(), " ")[0])
	if lastIndex >= int(totalResults) {
		return nil
	}

	// no next page
	return nil
}

func contains(s []string, searchterm string) bool {
	for _, a := range s {
		if strings.Contains(searchterm, a) {
			return true
		}
	}
	return false
}

type parseProductResponse struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Slug            string `json:"slug"`
	MetaDescription string `json:"meta_description"`
	MetaKeywords    string `json:"meta_keywords"`
	DisplayName     string `json:"display_name"`
	PriceRange      struct {
		Usd struct {
			Min string `json:"min"`
			Max string `json:"max"`
		} `json:"USD"`
	} `json:"price_range"`
	ProductCare           string      `json:"product_care"`
	Details               string      `json:"details"`
	EngravingType         interface{} `json:"engraving_type"`
	MaterialName          string      `json:"material_name"`
	Jit                   bool        `json:"jit"`
	PreOrder              bool        `json:"pre_order"`
	FakeReceivingPreorder bool        `json:"fake_receiving_preorder"`
	MaterialDescriptions  []struct {
		IconName    string `json:"icon_name"`
		Name        string `json:"name"`
		Description string `json:"description"`
		IconURL     string `json:"icon_url"`
	} `json:"material_descriptions"`
	Sample         bool `json:"sample"`
	TravelCase     bool `json:"travel_case"`
	EngagementRing bool `json:"engagement_ring"`
	Available      bool `json:"available"`
	Images         []struct {
		Position   int         `json:"position"`
		Alt        interface{} `json:"alt"`
		Attachment struct {
			URLOriginal string `json:"url_original"`
			URLMini     string `json:"url_mini"`
			URLSmall    string `json:"url_small"`
			URLMedium   string `json:"url_medium"`
			URLLarge    string `json:"url_large"`
		} `json:"attachment"`
	} `json:"images"`
	MaterialGroupProducts []struct {
		ID               int    `json:"id"`
		Slug             string `json:"slug"`
		MaterialCategory struct {
			Name        string `json:"name"`
			IconFullURL string `json:"icon_full_url"`
		} `json:"material_category"`
	} `json:"material_group_products"`
	Material string `json:"material"`
	Master   struct {
		ID  int    `json:"id"`
		Sku string `json:"sku"`
	} `json:"master"`
	Videos []struct {
		ID            int    `json:"id"`
		ThumbnailPath string `json:"thumbnail_path"`
		URL           string `json:"url"`
	} `json:"videos"`
	Variants []struct {
		ID           int    `json:"id"`
		Sku          string `json:"sku"`
		OptionValues []struct {
			Name         string `json:"name"`
			Presentation string `json:"presentation"`
			OptionTypeID int    `json:"option_type_id"`
		} `json:"option_values"`
		Prices []struct {
			Currency string `json:"currency"`
			Amount   string `json:"amount"`
		} `json:"prices"`
	} `json:"variants"`
	OptionTypes []struct {
		ID           int    `json:"id"`
		Name         string `json:"name"`
		Presentation string `json:"presentation"`
	} `json:"option_types"`
}

type parseRatingResponse struct {
	Status struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"status"`
	Response struct {
		Bottomline struct {
			TotalReview         int     `json:"total_review"`
			AverageScore        float64 `json:"average_score"`
			TotalOrganicReviews int     `json:"total_organic_reviews"`
			OrganicAverageScore int     `json:"organic_average_score"`
		} `json:"bottomline"`
	} `json:"response"`
}

type parseStockResponse struct {
	Online struct {
		Backorderable  bool   `json:"backorderable"`
		StockCount     int    `json:"stock_count"`
		ShipDate       string `json:"ship_date"`
		ReserveWording string `json:"reserve_wording"`
	} `json:"online"`
}

var (
	detailReg      = regexp.MustCompile(`({.*})`)
	imgWidthTplReg = regexp.MustCompile(`,+w_\d+`)
	stockApi       = func(variantId int, orderNumber string) string {
		return fmt.Sprintf("https://mejuri.com/api/v2/variants/%d/stock?order_number=%s", variantId, orderNumber)
	}
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

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	opts := c.CrawlOptions(resp.RawUrl())
	if !c.productApiPathMatcher.MatchString(resp.RawUrl().Path) {

		produrl := strings.ReplaceAll(resp.RawUrl().String(), "/shop/", "/api/v2/")

		respBody, err = c.variationRequest(ctx, produrl, resp.RawUrl().String(), nil)
		if err != nil {
			c.logger.Errorf("extract product %s failed", produrl)
			return err
		}
	}

	matched := detailReg.FindSubmatch(respBody)
	if len(matched) == 0 {
		c.httpClient.Jar().Clear(ctx, resp.Request.URL)
		return fmt.Errorf("extract produt json from page %s content failed", resp.Request.URL)
	}

	var viewData parseProductResponse

	if err := json.Unmarshal(respBody, &viewData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}

	materialGrpProducts := make([]string, 0, len(viewData.MaterialGroupProducts))
	materialGrpProducts = append(materialGrpProducts, viewData.Slug)
	for _, prodItem := range viewData.MaterialGroupProducts {
		if prodItem.ID == viewData.ID {
			continue
		}
		materialGrpProducts = append(materialGrpProducts, prodItem.Slug)
	}

	crawlUrl := resp.RawUrl().String()
	canUrl := doc.Find(`link[rel="canonical"]`).AttrOr("href", "")
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.RawUrl().String())
	}

	review, _ := strconv.ParseInt(strings.Split(strings.TrimSpace(doc.Find(`.yotpo-sum-reviews`).Text()), " ")[0])
	rating, _ := strconv.ParseFloat(strconv.Format(strings.ReplaceAll(doc.Find(`.yotpo-stars`).Text(), " star rating", "")))

	if review == 0 && rating == 0 {
		var viewReviewData parseRatingResponse

		produrl := `https://api.yotpo.com/v1/widget/EolV1WOLJ2UcFKuPJlrtxAIQCCoiDU7c8YqoW2pm/products/` + strconv.Format(viewData.ID) + `/reviews.json?widget=bottomline`

		respReviewBody, err := c.variationRequest(ctx, produrl, resp.Request.URL.String(), nil)
		if err != nil {
			c.logger.Errorf("extract product review %s failed", produrl)
			return err
		}
		if err := json.Unmarshal(respReviewBody, &viewReviewData); err != nil {
			c.logger.Errorf("unmarshal product review data fialed, error=%s", err)
			return err
		}

		review, _ = strconv.ParseInt(viewReviewData.Response.Bottomline.TotalReview)
		rating, _ = strconv.ParseFloat(viewReviewData.Response.Bottomline.AverageScore)
	}

	for pi, prodslug := range materialGrpProducts {
		if pi > 0 && prodslug != "" {
			produrl, _ := c.CanonicalUrl("/api/v2/products/" + prodslug)
			req, err := http.NewRequest(http.MethodGet, produrl, nil)
			if err != nil {
				c.logger.Error(err)
				return err
			}
			req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
			req.Header.Set("content-type", "application/json;charset=utf-8")
			req.Header.Set("referer", crawlUrl)

			respNew, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
				EnableProxy: true,
				Reliability: opts.Reliability,
			})

			if err != nil {
				c.logger.Error(err)
				return err
			}

			respBody, err = ioutil.ReadAll(respNew.Body)
			respNew.Body.Close()
			if err != nil {
				return err
			}
			crawlUrl = strings.ReplaceAll(crawlUrl, viewData.Slug, prodslug)
			canUrl = strings.ReplaceAll(canUrl, viewData.Slug, prodslug)
			if err := json.Unmarshal(respBody, &viewData); err != nil {
				c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
				return err
			}
		}

		color := viewData.MaterialName
		desc := viewData.Description + " " + viewData.Details
		for _, proditem := range viewData.MaterialDescriptions {
			desc = strings.Join(([]string{desc, proditem.Name, ": ", proditem.Description}), " ")
		}

		item := pbItem.Product{
			Source: &pbItem.Source{
				Id:           strconv.Format(viewData.ID),
				CrawlUrl:     crawlUrl,
				CanonicalUrl: canUrl,
			},
			Title:       viewData.Name,
			Description: desc,
			BrandName:   "Mejuri",
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
			},
			Stats: &pbItem.Stats{
				ReviewCount: int32(review),
				Rating:      float32(rating),
			},
			Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}

		var medias []*media.Media

		for _, img := range viewData.Images {
			itemImg, _ := anypb.New(&media.Media_Image{
				OriginalUrl: imgWidthTplReg.ReplaceAllString(img.Attachment.URLOriginal, ""),
				LargeUrl:    imgWidthTplReg.ReplaceAllString(img.Attachment.URLLarge, ",w_1000"),
				MediumUrl:   imgWidthTplReg.ReplaceAllString(img.Attachment.URLMedium, ",w_700"),
				SmallUrl:    imgWidthTplReg.ReplaceAllString(img.Attachment.URLSmall, ",w_500"),
			})
			medias = append(medias, &media.Media{
				Detail:    itemImg,
				IsDefault: len(medias) == 0,
			})
		}

		// Video
		for _, video := range viewData.Videos {
			medias = append(medias, pbMedia.NewVideoMedia(
				strconv.Format(video.ID),
				"",
				video.URL,
				0, 0, 0, video.ThumbnailPath, "",
				len(medias) == 0))
		}

		item.Medias = medias

		current, _ := strconv.ParseFloat(viewData.PriceRange.Usd.Min)
		msrp, _ := strconv.ParseFloat(viewData.PriceRange.Usd.Max)
		discount := 0.0

		sizetypeid := 0
		for _, opttype := range viewData.OptionTypes {
			if strings.Contains(strings.ToLower(opttype.Name), "size") || strings.Contains(strings.ToLower(opttype.Presentation), "size") {
				sizetypeid = opttype.ID
				break
			}
		}

		if len(viewData.Variants) > 0 {

			for _, variation := range viewData.Variants {

				if variation.OptionValues[0].OptionTypeID != sizetypeid {
					continue
				}

				sku := pbItem.Sku{
					SourceId: strconv.Format(variation.ID),
					Price: &pbItem.Price{
						Currency: regulation.Currency_USD,
						Current:  int32(current * 100),
						Msrp:     int32(msrp * 100),
						Discount: int32(discount),
					},

					Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
				}
				for _, price := range variation.Prices {
					if strings.ToUpper(strings.TrimSpace(price.Currency)) == "USD" && price.Amount != "" {
						amount, _ := strconv.ParseFloat(price.Amount)
						amount = amount * 100
						sku.Price.Msrp = int32(amount)
						sku.Price.Current = int32(amount)
						sku.Price.Discount = 0
						break
					}
				}

				// stock
				var stockData parseStockResponse
				sessionData := getSessionFromCtx(ctx)
				stockRespBody, err := c.variationRequest(ctx, stockApi(variation.ID, sessionData.Current.Order.Number), crawlUrl, map[string]string{
					"x-spree-order-token": sessionData.Current.Order.Token,
				})
				if err != nil {
					c.logger.Errorf("get sku %s stock error=%s", variation.ID, err)
					return err
				}
				if err := json.Unmarshal(stockRespBody, &stockData); err != nil {
					c.logger.Errorf("unmarshal sku %s stock data failed, error=%s", variation.ID, err)
					return err
				}
				if stockData.Online.StockCount > 0 {
					sku.Stock.StockStatus = pbItem.Stock_InStock
					item.Stock.StockStatus = pbItem.Stock_InStock
					sku.Stock.StockCount = int32(stockData.Online.StockCount)
				}

				if color != "" {
					//sku.SourceId = fmt.Sprintf("%s-%v", color, viewData.ID)
					sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
						Type:  pbItem.SkuSpecType_SkuSpecColor,
						Id:    color,
						Name:  color,
						Value: color,
					})
				}

				if variation.OptionValues[0].Name != "" {
					sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
						Type:  pbItem.SkuSpecType_SkuSpecSize,
						Id:    variation.OptionValues[0].Name,
						Name:  variation.OptionValues[0].Name,
						Value: variation.OptionValues[0].Name,
					})
				}

				if len(sku.Specs) == 0 {
					sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
						Type:  pbItem.SkuSpecType_SkuSpecColor,
						Id:    "-",
						Name:  "-",
						Value: "-",
					})
				}
				item.SkuItems = append(item.SkuItems, &sku)
			}
		} else {

			sku := pbItem.Sku{
				SourceId: strconv.Format(viewData.Master.ID),
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(current * 100),
					Msrp:     int32(msrp * 100),
					Discount: int32(discount),
				},

				Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}
			// stock
			var stockData parseStockResponse
			sessionData := getSessionFromCtx(ctx)
			stockRespBody, err := c.variationRequest(ctx, stockApi(viewData.Master.ID, sessionData.Current.Order.Number), crawlUrl, map[string]string{
				"x-spree-order-token": sessionData.Current.Order.Token,
			})
			if err != nil {
				c.logger.Errorf("get sku %s stock error=%s", viewData.Master.ID, err)
				return err
			}
			if err := json.Unmarshal(stockRespBody, &stockData); err != nil {
				c.logger.Errorf("unmarshal sku %s stock data failed, error=%s", viewData.Master.ID, err)
				return err
			}
			if stockData.Online.StockCount > 0 {
				sku.Stock.StockStatus = pbItem.Stock_InStock
				item.Stock.StockStatus = pbItem.Stock_InStock
				sku.Stock.StockCount = int32(stockData.Online.StockCount)
			}

			if color != "" {
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecColor,
					Id:    color,
					Name:  color,
					Value: color,
				})
			}

			if len(sku.Specs) == 0 {
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecColor,
					Id:    "-",
					Name:  "-",
					Value: "-",
				})
			}
			item.SkuItems = append(item.SkuItems, &sku)
		}

		if err = yield(ctx, &item); err != nil {
			return err
		}
	}
	return nil
}

func (c *_Crawler) variationRequest(ctx context.Context, url string, referer string, exHeader map[string]string) ([]byte, error) {

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	opts := c.CrawlOptions(req.URL)
	req.Header.Set("accept-language", "en-GB,en-US;q=0.9,en;q=0.8")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("content-type", "application/json;charset=utf-8")
	req.Header.Set("x-requested-with", "XMLHttpRequest")
	req.Header.Set("referer", referer)
	if exHeader != nil {
		for s, s2 := range exHeader {
			req.Header.Set(s, s2)
		}
	}

	for _, c := range opts.MustCookies {
		req.AddCookie(c)
	}
	for k := range opts.MustHeader {
		req.Header.Set(k, opts.MustHeader.Get(k))
	}
	resp, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
		EnableProxy:       true,
		EnableHeadless:    false,
		EnableSessionInit: false,
		Reliability:       opts.Reliability,
	})
	if err != nil {
		c.logger.Error(err)
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)

}

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://mejuri.com/",
		//"https://mejuri.com/shop/t/new-arrivals",
		//	"https://mejuri.com/shop/products/large-diamond-necklace",
		//"https://mejuri.com/shop/products/heirloom-ring-garnet",
		//"https://mejuri.com/shop/t/type/pendants",
		//"https://mejuri.com/shop/products/heirloom-ring-garnet",
		//"https://mejuri.com/shop/t/type/earrings",

		//"https://mejuri.com/shop/t/type/rings",
		//"https://mejuri.com/shop/t/type/single-earrings",
		//"https://mejuri.com/shop/products/single-opal-u-hoop",
		//"https://mejuri.com/shop/products/tiny-diamond-stud",
		//"https://mejuri.com/shop/t/type?fbm=14k%20White%20Gold",
		//"https://mejuri.com/shop/products/diamond-necklace-white-gold",
		//"https://mejuri.com/shop/products/golden-crew-sweatshirt",
		//"https://mejuri.com/shop/t/type",
		//"https://www.mejuri.com/collections/charlotte-family",
		"https://mejuri.com/shop/products/cable-chain-tag-necklace",
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
