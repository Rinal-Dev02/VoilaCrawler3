package main

import (
	"bytes"
	"context"
	"encoding/json"
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
	pbMedia "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
)

type _Crawler struct {
	httpClient http.Client

	categoryPathMatcher *regexp.Regexp
	productPathMatcher  *regexp.Regexp
	logger              glog.Log
}

func (_ *_Crawler) New(_ *cli.Context, client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient:          client,
		categoryPathMatcher: regexp.MustCompile(`^/([/A-Za-z0-9_-]+)$`),
		productPathMatcher:  regexp.MustCompile(`^/([/A-Za-z0-9_-]+)-p\d+$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "0954de0e642940cba0210f2e3ab74b18"
}

// Version
func (c *_Crawler) Version() int32 {
	return 1
}

// CrawlOptions
func (c *_Crawler) CrawlOptions(u *url.URL) *crawler.CrawlOptions {
	opts := &crawler.CrawlOptions{
		EnableHeadless: true,
		// use js api to init session for the first request of the crawl
		EnableSessionInit: false,
		Reliability:       proxy.ProxyReliability_ReliabilityDefault,
	}
	return opts
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.burberry.com"}
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
		u.Host = "www.burberry.com"
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
	p := strings.TrimSuffix(resp.RawUrl().Path, "/")
	if p == "" {
		return c.parseCategories(ctx, resp, yield)
	}

	if c.productPathMatcher.MatchString(p) {
		fmt.Println(`productPathMatcher`)
		return c.parseProduct(ctx, resp, yield)
	} else if c.categoryPathMatcher.MatchString(p) {
		fmt.Println(`categoryPathMatcher`)
		return c.parseCategoryProducts(ctx, resp, yield)
	}
	return crawler.ErrUnsupportedPath
}

// --------------------------------------------------

// var categoriesExtractReg = regexp.MustCompile(`<script id="__NEXT_DATA__" type="application/json">\s*({.*});`)

var categoryExtractReg = regexp.MustCompile(`(?U)<script id="__NEXT_DATA__" type="application/json">\s*({.*})\s*</script>`)

func (c *_Crawler) parseCategories(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	respBody, err := resp.RawBody()
	if err != nil {
		return err
	}

	matched := categoryExtractReg.FindSubmatch([]byte(respBody))
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}

	var viewData categoryStructure
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		return err
	}

	for _, nodes := range viewData.Props.PageProps.HeaderNavigation {

		nctx := context.WithValue(ctx, "Category", nodes.Link.Href)

		for _, rawcat := range nodes.Items {

			nnctx := context.WithValue(nctx, "SubCategory", rawcat.Link.Href)

			for _, rawcatdata := range rawcat.Items {

				href := rawcatdata.Link.Href

				if href == "" {
					continue
				}

				u, err := url.Parse(href)
				if err != nil {
					c.logger.Errorf("parse url %s failed", href)
					continue
				}

				if c.categoryPathMatcher.MatchString(u.Path) {
					nnnctx := context.WithValue(nnctx, "SubCategory2", rawcatdata.Link.Href)
					req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
					if err := yield(nnnctx, req); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

type categoryStructure struct {
	Props struct {
		PageProps struct {
			InternalRoutes     []string      `json:"internalRoutes"`
			ExperimentalRoutes []interface{} `json:"experimentalRoutes"`
			HeaderNavigation   []struct {
				ID               string      `json:"id"`
				Key              string      `json:"key"`
				NavigationPath   string      `json:"navigationPath"`
				Type             string      `json:"type"`
				Group            string      `json:"group"`
				Selected         bool        `json:"selected"`
				PageType         string      `json:"pageType,omitempty"`
				Secondary        bool        `json:"secondary"`
				HeroCategory     bool        `json:"heroCategory"`
				PromotedCategory bool        `json:"promotedCategory"`
				ShowOnSitemap    bool        `json:"showOnSitemap"`
				BadgeLabel       interface{} `json:"badgeLabel"`
				Link             struct {
					Title    string `json:"title"`
					TitleEn  string `json:"titleEn"`
					Href     string `json:"href"`
					External bool   `json:"external"`
				} `json:"link"`
				Items []struct {
					ID               string `json:"id"`
					Key              string `json:"key"`
					ParentID         string `json:"parentId"`
					NavigationPath   string `json:"navigationPath"`
					Type             string `json:"type"`
					Selected         bool   `json:"selected"`
					PageType         string `json:"pageType"`
					Secondary        bool   `json:"secondary"`
					HeroCategory     bool   `json:"heroCategory"`
					PromotedCategory bool   `json:"promotedCategory"`
					ShowOnSitemap    bool   `json:"showOnSitemap"`
					NavBanner        struct {
						VariationsMap struct {
							Native struct {
								CellID  string `json:"cell_id"`
								Options struct {
									Image                string      `json:"image"`
									ServiceGateChildType string      `json:"serviceGateChildType"`
									ImageAlt             interface{} `json:"image_alt"`
									ImageAttributes      struct {
									} `json:"image_attributes"`
									Hotspots      interface{} `json:"hotspots"`
									FallbackImage string      `json:"fallbackImage"`
								} `json:"options"`
								TemplateGroup     string `json:"template_group"`
								TemplateItem      string `json:"template_item"`
								CountdownEnabled  bool   `json:"countdown_enabled"`
								HideTopPadding    bool   `json:"hide_top_padding"`
								HideBottomPadding bool   `json:"hide_bottom_padding"`
							} `json:"native"`
						} `json:"variations_map"`
					} `json:"navBanner"`
					BadgeLabel interface{} `json:"badgeLabel"`
					Link       struct {
						Title    string `json:"title"`
						TitleEn  string `json:"titleEn"`
						Href     string `json:"href"`
						External bool   `json:"external"`
					} `json:"link"`
					Items []struct {
						ID               string `json:"id"`
						Key              string `json:"key"`
						ParentID         string `json:"parentId"`
						NavigationPath   string `json:"navigationPath"`
						Type             string `json:"type"`
						Selected         bool   `json:"selected"`
						PageType         string `json:"pageType,omitempty"`
						Secondary        bool   `json:"secondary"`
						HeroCategory     bool   `json:"heroCategory"`
						PromotedCategory bool   `json:"promotedCategory"`
						ShowOnSitemap    bool   `json:"showOnSitemap"`
						NavBanner        struct {
							CellID string `json:"cell_id"`
							Copy   struct {
								EnTitle string `json:"en_title"`
								Title   string `json:"title"`
							} `json:"copy"`
							Link struct {
								URL              string `json:"url"`
								OpensInNewWindow bool   `json:"opens_in_new_window"`
							} `json:"link"`
							Options struct {
								Image                string      `json:"image"`
								ServiceGateChildType string      `json:"serviceGateChildType"`
								ImageAlt             interface{} `json:"image_alt"`
								ImageAttributes      struct {
								} `json:"image_attributes"`
								Hotspots      interface{} `json:"hotspots"`
								FallbackImage string      `json:"fallbackImage"`
							} `json:"options"`
							TemplateGroup     string `json:"template_group"`
							TemplateItem      string `json:"template_item"`
							CountdownEnabled  bool   `json:"countdown_enabled"`
							HideTopPadding    bool   `json:"hide_top_padding"`
							HideBottomPadding bool   `json:"hide_bottom_padding"`
						} `json:"navBanner,omitempty"`
						BadgeLabel interface{} `json:"badgeLabel"`
						Link       struct {
							Title    string `json:"title"`
							TitleEn  string `json:"titleEn"`
							Href     string `json:"href"`
							External bool   `json:"external"`
						} `json:"link"`
					} `json:"items"`
				} `json:"items,omitempty"`
				Media      string `json:"media,omitempty"`
				MainBanner struct {
					CellID  string `json:"cell_id"`
					Options struct {
						Image                string      `json:"image"`
						ServiceGateChildType string      `json:"serviceGateChildType"`
						ImageAlt             interface{} `json:"image_alt"`
						ImageAttributes      struct {
						} `json:"image_attributes"`
						Hotspots      interface{} `json:"hotspots"`
						FallbackImage string      `json:"fallbackImage"`
					} `json:"options"`
					TemplateGroup     string `json:"template_group"`
					TemplateItem      string `json:"template_item"`
					CountdownEnabled  bool   `json:"countdown_enabled"`
					HideTopPadding    bool   `json:"hide_top_padding"`
					HideBottomPadding bool   `json:"hide_bottom_padding"`
				} `json:"mainBanner,omitempty"`
				NavBanner struct {
					CellID string `json:"cell_id"`
					Copy   struct {
						EnTitle string `json:"en_title"`
						Title   string `json:"title"`
					} `json:"copy"`
					Link struct {
						URL              string `json:"url"`
						OpensInNewWindow bool   `json:"opens_in_new_window"`
					} `json:"link"`
					Options struct {
						Image                string      `json:"image"`
						ServiceGateChildType string      `json:"serviceGateChildType"`
						ImageAlt             interface{} `json:"image_alt"`
						ImageAttributes      struct {
						} `json:"image_attributes"`
						Hotspots      interface{} `json:"hotspots"`
						FallbackImage string      `json:"fallbackImage"`
					} `json:"options"`
					TemplateGroup     string `json:"template_group"`
					TemplateItem      string `json:"template_item"`
					CountdownEnabled  bool   `json:"countdown_enabled"`
					HideTopPadding    bool   `json:"hide_top_padding"`
					HideBottomPadding bool   `json:"hide_bottom_padding"`
				} `json:"navBanner,omitempty"`
			} `json:"headerNavigation"`
		} `json:"pageProps"`
		NSSP bool `json:"__N_SSP"`
	} `json:"props"`
}

// --------------------------------------------------

// nextIndex used to get sharingData from context
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")) + 1)
}

func TrimSpaceNewlineInString(s string) string {
	re := regexp.MustCompile(`\n`)
	resp := re.ReplaceAllString(s, " ")
	resp = strings.ReplaceAll(resp, "\\n", " ")
	resp = strings.ReplaceAll(resp, "\r", " ")
	resp = strings.ReplaceAll(resp, "\t", " ")
	resp = strings.ReplaceAll(resp, "  ", "")
	return resp
}

type categoryProductStructure struct {
	Db struct {
		ProductCards map[string]struct {
			//Num80097771 struct {
			ID  string `json:"id"`
			URL string `json:"url"`
			//Key string `json:"key"`
			//} `json:"80097771"`
		} `json:"productCards"`
	} `json:"db"`
	Data struct {
		Entities struct {
			ProductCards map[string]struct {
				//Num40733711 struct {
				ID  string `json:"id"`
				URL string `json:"url"`
				//} `json:"40733711"`
			} `json:"productCards"`
		} `json:"entities"`
		Result []string `json:"result"`
	} `json:"data"`
}

var categoryProductReg = regexp.MustCompile(`__PRELOADED_STATE__\s*=\s*({.*});`)

// parseCategoryProducts parse api url from web page url
func (c *_Crawler) parseCategoryProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}
	respBody, err := resp.RawBody()
	if err != nil {
		return err
	}

	matched := categoryProductReg.FindSubmatch([]byte(respBody))
	if len(matched) <= 1 {
		matched = append(matched, []byte(``))
		matched = append(matched, []byte(respBody))
		if len(matched) <= 1 {
			c.logger.Debugf("%s", respBody)
			return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
		}
	}

	var viewData categoryProductStructure
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		return err
	}

	lastIndex := nextIndex(ctx)

	nextURL := ""

	if len(viewData.Db.ProductCards) == 0 && len(viewData.Data.Entities.ProductCards) == 0 {
		return nil
	}

	for _, rawcat := range viewData.Db.ProductCards {

		href := rawcat.URL
		if href == "" {
			continue
		}

		req, err := http.NewRequest(http.MethodGet, href, nil)
		if err != nil {
			c.logger.Error(err)
			continue
		}

		nctx := context.WithValue(ctx, "item.index", lastIndex)
		lastIndex += 1

		if err := yield(nctx, req); err != nil {
			return err
		}
	}

	if viewData.Data.Entities.ProductCards != nil {
		fmt.Println(`type 2`)
		for _, rawcat := range viewData.Data.Entities.ProductCards {

			href := rawcat.URL
			if href == "" {
				continue
			}

			req, err := http.NewRequest(http.MethodGet, href, nil)
			if err != nil {
				c.logger.Error(err)
				continue
			}

			nctx := context.WithValue(ctx, "item.index", lastIndex)
			lastIndex += 1

			if err := yield(nctx, req); err != nil {
				return err
			}
		}

		if len(viewData.Data.Entities.ProductCards) < 100 {
			return nil
		}
	}

	if strings.Contains(resp.Request.URL.Path, `web-api/pages/products`) {
		nextURL = resp.Request.URL.String()
	} else {
		nextURL = "https://us.burberry.com/web-api/pages/products?offset=20&limit=100&country=US&language=en&pagePath=" + resp.Request.URL.Path + "products" + resp.Request.URL.Query().Encode()
	}

	// set pagination
	u, _ := url.Parse(nextURL)
	vals := u.Query()
	vals.Set("offset", strconv.Format(lastIndex-1))
	vals.Set("limit", strconv.Format(100))
	u.RawQuery = vals.Encode()

	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)

	nctx := context.WithValue(ctx, "item.index", lastIndex-1)
	return yield(nctx, req)
}

type productStructure struct {
	Db struct {
		Pages map[string]struct {
			//TheMidLengthChelseaHeritageTrenchCoatP40733751 struct {
			ID             string      `json:"id"`
			NavigationID   string      `json:"navigationId"`
			CuratedProduct bool        `json:"curatedProduct"`
			ProductCount   int         `json:"productCount"`
			Placements     interface{} `json:"placements"`
			PageType       string      `json:"pageType"`
			Data           struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				DefaultName string `json:"defaultName"`
				Color       string `json:"color"`
				Breadcrumbs []struct {
					ID      string `json:"id"`
					Label   string `json:"label"`
					Title   string `json:"title"`
					URLLink string `json:"url_link"`
				} `json:"breadcrumbs"`
				Features interface{} `json:"features"`
				Product  struct {
					Sku string `json:"sku"`
				} `json:"product"`
				MonogrammingPromotionalText interface{} `json:"monogrammingPromotionalText"`
				OnSale                      bool        `json:"onSale"`
				SaleRestricted              bool        `json:"saleRestricted"`
				Label                       string      `json:"label"`
				Content                     struct {
					Title       string `json:"title"`
					Description string `json:"description"`
					Features    []struct {
						Text           string `json:"text"`
						TranslationKey string `json:"translationKey,omitempty"`
					} `json:"features"`
					ShowInfoPanelButton bool `json:"showInfoPanelButton"`
				} `json:"content"`
				ShouldShowKCMark bool   `json:"shouldShowKCMark"`
				CountryCode      string `json:"countryCode"`
				GalleryItems     []struct {
					Image struct {
						Key          string `json:"key"`
						ImageDefault string `json:"imageDefault"`
						ImageAlt     string `json:"imageAlt"`
						Sources      []struct {
							Media  string `json:"media"`
							SrcSet string `json:"srcSet"`
						} `json:"sources"`
						ImageFallback string `json:"imageFallback"`
					} `json:"image"`
				} `json:"galleryItems"`
				Price struct {
					Current struct {
						Value     int    `json:"value"`
						Currency  string `json:"currency"`
						Formatted string `json:"formatted"`
					} `json:"current"`
					Old struct {
						Value     int    `json:"value"`
						Currency  string `json:"currency"`
						Formatted string `json:"formatted"`
					} `json:"old"`
				} `json:"price"`
				Sizes []struct {
					Label              string `json:"label"`
					PreOrderStockLevel int    `json:"preOrderStockLevel"`
					SleeveLength       string `json:"sleeveLength"`
					StockQuantity      int    `json:"stockQuantity"`
					IsPreOrderStock    bool   `json:"isPreOrderStock"`
					IsInStock          bool   `json:"isInStock"`
					Sku                string `json:"sku"`
				} `json:"sizes"`
				Dimensions []struct {
					Label    string `json:"label"`
					Products []struct {
						ID       string `json:"id"`
						Link     string `json:"link"`
						Label    string `json:"label"`
						Selected bool   `json:"selected"`
					} `json:"products"`
					ActiveDimensionLabel string `json:"activeDimensionLabel"`
				} `json:"dimensions"`
				SwatchItems []struct {
					ID         string `json:"id"`
					URL        string `json:"url"`
					Image      string `json:"image"`
					Label      string `json:"label"`
					IsSelected bool   `json:"isSelected"`
				} `json:"swatchItems"`
			} `json:"data"`

			//} `json:"/the-mid-length-chelsea-heritage-trench-coat-p40733751"`
		} `json:"pages"`
	} `json:"db"`
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

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	matched := categoryProductReg.FindSubmatch([]byte(respBody))
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}

	var viewData productStructure
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		return err
	}

	canUrl, _ := c.CanonicalUrl(doc.Find(`link[rel="canonical"]`).AttrOr("href", ""))
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}

	for _, prod := range viewData.Db.Pages {

		desc := TrimSpaceNewlineInString(prod.Data.Content.Description) + " "
		for _, item := range prod.Data.Content.Features {
			desc = desc + item.Text + ", "
		}

		item := pbItem.Product{
			Source: &pbItem.Source{
				Id:           prod.Data.ID,
				CrawlUrl:     resp.Request.URL.String(),
				CanonicalUrl: canUrl,
			},
			Title:       prod.Data.Name,
			Description: desc,
			BrandName:   "Burberry",
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
			},
			Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}

		for _, node := range prod.Data.Breadcrumbs {
			item.Category = node.Title + ", "

		}
		var medias []*pbMedia.Media
		for m, mid := range prod.Data.GalleryItems {

			template := strings.Split(mid.Image.ImageDefault, "?")[0]
			medias = append(medias, pbMedia.NewImageMedia(
				strconv.Format(m),
				template,
				template+"?wid=1000&hei=1000",
				template+"?wid=800&hei=800",
				template+"?wid=600&hei=600",
				"",
				m == 0,
			))
		}

		item.Medias = append(item.Medias, medias...)

		var colorSelected *pbItem.SkuSpecOption

		for _, rawcolor := range prod.Data.SwatchItems {
			if rawcolor.IsSelected {

				colorSelected = &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecColor,
					Id:    rawcolor.ID,
					Name:  rawcolor.Label,
					Value: rawcolor.Label,
					Icon:  rawcolor.Image,
				}
			}
		}

		for i, rawsku := range prod.Data.Sizes {

			current, _ := strconv.ParsePrice(prod.Data.Price.Current.Value)
			msrp, _ := strconv.ParsePrice(prod.Data.Price.Old.Value)
			discount := 0.0
			if msrp == 0.0 {
				msrp = current
			}
			if msrp > current {
				discount = ((msrp - current) / msrp) * 100
			}

			sku := pbItem.Sku{
				SourceId: fmt.Sprintf("%s-%d", rawsku.Sku, i),
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(current),
					Msrp:     int32(msrp),
					Discount: int32(discount),
				},

				Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}

			if rawsku.IsPreOrderStock {
				sku.Stock.StockStatus = pbItem.Stock_InStock
				item.Stock.StockStatus = pbItem.Stock_InStock
			}

			if colorSelected != nil {
				sku.Specs = append(sku.Specs, colorSelected)
			}
			sizeName := rawsku.Label
			if sizeName == "" {
				sizeName = "One Size"
			}
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    fmt.Sprintf("%s-s-%d", rawsku.Sku, i),
				Name:  sizeName,
				Value: sizeName,
			})

			item.SkuItems = append(item.SkuItems, &sku)
		}

		if err = yield(ctx, &item); err != nil {
			return err
		}
	}
	return nil
}

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://us.burberry.com/",
		//"https://us.burberry.com/womens-coats/",
		//	"https://us.burberry.com/tb-summer-monogram-collection-women/",
		//"https://us.burberry.com/small-leather-tb-bag-p80345521",
		"https://us.burberry.com/canvas-leathersuede-sneakers-p80421971",
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
