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
		categorySearchAPIPathMatcher: regexp.MustCompile(`^/shop/t/type([/A-Za-z0-9_-]+)$`),
		categoryPathMatcher:          regexp.MustCompile(`^/([/A-Za-z_-]+)$`),
		productPathMatcher:           regexp.MustCompile(`^/shop/products([/A-Za-z0-9_-]+)$`),
		productApiPathMatcher:        regexp.MustCompile(`^/api/v2/products([/A-Za-z0-9_-]+)$`),
		logger:                       logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "3c608f04da5f4bc6927b473ebcebd17d"
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
	return []string{"*.mejuri.com"}
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

var countriesPrefix = map[string]struct{}{"/ad": {}, "/ae": {}, "/ar-ae": {}, "/af": {}, "/ag": {}, "/ai": {}, "/al": {}, "/am": {}, "/an": {}, "/ao": {}, "/aq": {}, "/ar": {}, "/at": {}, "/au": {}, "/aw": {}, "/az": {}, "/ba": {}, "/bb": {}, "/bd": {}, "/be": {}, "/bf": {}, "/bg": {}, "/bh": {}, "/ar-bh": {}, "/bi": {}, "/bj": {}, "/bm": {}, "/bn": {}, "/bo": {}, "/br": {}, "/bs": {}, "/bt": {}, "/bv": {}, "/bw": {}, "/by": {}, "/bz": {}, "/ca": {}, "/cc": {}, "/cf": {}, "/cg": {}, "/ch": {}, "/ci": {}, "/ck": {}, "/cl": {}, "/cm": {}, "/cn": {}, "/co": {}, "/cr": {}, "/cv": {}, "/cx": {}, "/cy": {}, "/cz": {}, "/de": {}, "/dj": {}, "/dk": {}, "/dm": {}, "/do": {}, "/dz": {}, "/ec": {}, "/ee": {}, "/eg": {}, "/ar-eg": {}, "/eh": {}, "/es": {}, "/et": {}, "/fi": {}, "/fj": {}, "/fk": {}, "/fm": {}, "/fo": {}, "/fr": {}, "/ga": {}, "/uk": {}, "/gd": {}, "/ge": {}, "/gf": {}, "/gg": {}, "/gh": {}, "/gi": {}, "/gl": {}, "/gm": {}, "/gn": {}, "/gp": {}, "/gq": {}, "/gr": {}, "/gt": {}, "/gu": {}, "/gw": {}, "/gy": {}, "/hk": {}, "/hn": {}, "/hr": {}, "/ht": {}, "/hu": {}, "/ic": {}, "/id": {}, "/ie": {}, "/il": {}, "/in": {}, "/io": {}, "/iq": {}, "/ar-iq": {}, "/is": {}, "/it": {}, "/je": {}, "/jm": {}, "/jo": {}, "/ar-jo": {}, "/jp": {}, "/ke": {}, "/kg": {}, "/kh": {}, "/ki": {}, "/km": {}, "/kn": {}, "/kr": {}, "/kv": {}, "/kw": {}, "/ar-kw": {}, "/ky": {}, "/kz": {}, "/la": {}, "/lb": {}, "/ar-lb": {}, "/lc": {}, "/li": {}, "/lk": {}, "/ls": {}, "/lt": {}, "/lu": {}, "/lv": {}, "/ma": {}, "/mc": {}, "/md": {}, "/me": {}, "/mg": {}, "/mh": {}, "/mk": {}, "/ml": {}, "/mn": {}, "/mo": {}, "/mp": {}, "/mq": {}, "/mr": {}, "/ms": {}, "/mt": {}, "/mu": {}, "/mv": {}, "/mw": {}, "/mx": {}, "/my": {}, "/mz": {}, "/na": {}, "/nc": {}, "/ne": {}, "/nf": {}, "/ng": {}, "/ni": {}, "/nl": {}, "/no": {}, "/np": {}, "/nr": {}, "/nu": {}, "/nz": {}, "/om": {}, "/ar-om": {}, "/pa": {}, "/pe": {}, "/pf": {}, "/pg": {}, "/ph": {}, "/pk": {}, "/pl": {}, "/pm": {}, "/pn": {}, "/pr": {}, "/pt": {}, "/pw": {}, "/py": {}, "/qa": {}, "/ar-qa": {}, "/re": {}, "/ro": {}, "/rs": {}, "/ru": {}, "/rw": {}, "/sa": {}, "/ar-sa": {}, "/sb": {}, "/sc": {}, "/se": {}, "/sg": {}, "/sh": {}, "/si": {}, "/sk": {}, "/sl": {}, "/sm": {}, "/sn": {}, "/sr": {}, "/st": {}, "/sv": {}, "/sz": {}, "/tc": {}, "/td": {}, "/tg": {}, "/th": {}, "/tj": {}, "/tk": {}, "/tl": {}, "/tn": {}, "/to": {}, "/tr": {}, "/tt": {}, "/tv": {}, "/tw": {}, "/tz": {}, "/ua": {}, "/ug": {}, "/uy": {}, "/uz": {}, "/va": {}, "/vc": {}, "/ve": {}, "/vg": {}, "/vi": {}, "/vn": {}, "/vu": {}, "/wf": {}, "/xc": {}, "/ye": {}, "/za": {}, "/zm": {}, "/zw": {}}

func getPathFirstSection(p string) string {
	return "/" + strings.SplitN(strings.TrimPrefix(p, "/"), "/", 2)[0]
}

func (c *_Crawler) Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	prefix := getPathFirstSection(resp.Request.URL.Path)
	if _, ok := countriesPrefix[prefix]; ok {
		req := resp.Request
		req.URL.Path = strings.TrimPrefix(req.URL.Path, prefix)

		opts := c.CrawlOptions(req.URL)
		for k := range opts.MustHeader {
			req.Header.Set(k, opts.MustHeader.Get(k))
		}
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
		c.logger.Infof("Access %s", req.URL.String())
		if res, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
			EnableProxy:       true,
			EnableHeadless:    opts.EnableHeadless,
			EnableSessionInit: opts.EnableSessionInit,
			DisableCookieJar:  opts.DisableCookieJar,
			Reliability:       opts.Reliability,
		}); err != nil {
			return err
		} else {
			resp = res
		}
	}

	yieldWrap := func(ctx context.Context, val interface{}) error {
		switch v := val.(type) {
		case *http.Request:
			prefix := getPathFirstSection(v.URL.Path)
			if _, ok := countriesPrefix[prefix]; ok {
				v.URL.Path = strings.TrimPrefix(v.URL.Path, prefix)
			}
			return yield(ctx, v)
		default:
			return yield(ctx, val)
		}
	}

	p := strings.TrimSuffix(resp.RawUrl().Path, "/")
	if p == "" {
		return crawler.ErrUnsupportedPath
	}
	if c.productPathMatcher.MatchString(p) || c.productApiPathMatcher.MatchString(p) {
		return c.parseProduct(ctx, resp, yieldWrap)
	} else if c.categoryPathMatcher.MatchString(p) {
		return c.parseCategoryProducts(ctx, resp, yieldWrap)
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
	if c.categorySearchAPIPathMatcher.MatchString(resp.Request.URL.Path) {
		categoryURL := "https://mejuri.com/api/v1/taxon/collections-by-categories/USD/type"
		varResponse, err := c.variationRequest(ctx, categoryURL, resp.Request.URL.String())
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

var (
	detailReg      = regexp.MustCompile(`({.*})`)
	imgWidthTplReg = regexp.MustCompile(`,+w_\d+`)
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

	opts := c.CrawlOptions(resp.Request.URL)
	if !c.productApiPathMatcher.MatchString(resp.Request.URL.Path) {

		produrl := strings.ReplaceAll(resp.Request.URL.String(), "/shop/", "/api/v2/")

		respBody, err = c.variationRequest(ctx, produrl, resp.Request.URL.String())
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

	var materialgrpproducts = make([]string, len(viewData.MaterialGroupProducts)+1)
	for i, proditem := range viewData.MaterialGroupProducts {
		if proditem.ID == viewData.ID {
			materialgrpproducts[0] = proditem.Slug
		} else {
			materialgrpproducts[i+1] = proditem.Slug
		}
	}

	canUrl, _ := c.CanonicalUrl(doc.Find(`link[rel="canonical"]`).AttrOr("href", ""))
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}

	review, _ := strconv.ParseInt(strings.Split(strings.TrimSpace(doc.Find(`.yotpo-sum-reviews`).Text()), " ")[0])
	rating, _ := strconv.ParseFloat(strconv.Format(strings.ReplaceAll(doc.Find(`.yotpo-stars`).Text(), " star rating", "")))

	if review == 0 && rating == 0 {
		var viewReviewData parseRatingResponse

		produrl := `https://api.yotpo.com/v1/widget/EolV1WOLJ2UcFKuPJlrtxAIQCCoiDU7c8YqoW2pm/products/` + strconv.Format(viewData.ID) + `/reviews.json?widget=bottomline`

		respReviewBody, err := c.variationRequest(ctx, produrl, resp.Request.URL.String())
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

	for pi, prodslug := range materialgrpproducts {

		if pi > 0 && prodslug != "" {

			produrl, _ := c.CanonicalUrl("/api/v2/products/" + prodslug)

			req, err := http.NewRequest(http.MethodGet, produrl, nil)
			if err != nil {
				c.logger.Error(err)
				return err
			}
			req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
			req.Header.Set("content-type", "application/json;charset=utf-8")
			req.Header.Set("referer", canUrl)

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

			if err := json.Unmarshal(respBody, &viewData); err != nil {
				c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
				return err
			}
		}
		// else if prodslug == "" {
		// 	//continue
		// }

		color := viewData.MaterialName
		desc := viewData.Description + " " + viewData.Details
		for _, proditem := range viewData.MaterialDescriptions {
			desc = strings.Join(([]string{desc, proditem.Name, ": ", proditem.Description}), " ")
		}

		item := pbItem.Product{
			Source: &pbItem.Source{
				Id:           strconv.Format(viewData.ID),
				CrawlUrl:     resp.Request.URL.String(),
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
				if viewData.Available {
					sku.Stock.StockStatus = pbItem.Stock_InStock
					item.Stock.StockStatus = pbItem.Stock_InStock
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
				SourceId: strconv.Format(viewData.ID),
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(current * 100),
					Msrp:     int32(msrp * 100),
					Discount: int32(discount),
				},

				Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}
			if viewData.Available {
				sku.Stock.StockStatus = pbItem.Stock_InStock
				item.Stock.StockStatus = pbItem.Stock_InStock
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

func (c *_Crawler) variationRequest(ctx context.Context, url string, referer string) ([]byte, error) {

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	opts := c.CrawlOptions(req.URL)
	req.Header.Set("accept-language", "en-GB,en-US;q=0.9,en;q=0.8")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("content-type", "application/json;charset=utf-8")
	req.Header.Set("x-requested-with", "XMLHttpRequest")
	req.Header.Set("referer", referer)

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
