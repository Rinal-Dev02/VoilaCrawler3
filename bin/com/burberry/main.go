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
		EnableHeadless: false,
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
		u.Host = "us.burberry.com"
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

	if c.productPathMatcher.MatchString(p) {
		return c.parseProduct(ctx, resp, yield)
	} else if c.categoryPathMatcher.MatchString(p) {
		return c.parseCategoryProducts(ctx, resp, yield)
	}
	return crawler.ErrUnsupportedPath
}

var categoryExtractReg = regexp.MustCompile(`(?U)<script id="__NEXT_DATA__" type="application/json">\s*({.*})\s*</script>`)

func (c *_Crawler) GetCategories(ctx context.Context) ([]*pbItem.Category, error) {
	req, _ := http.NewRequest(http.MethodGet, "https://us.burberry.com/", nil)
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
	respBody, _ := resp.RawBody()
	//	dom, err := resp.Selector()
	if err != nil {
		c.logger.Error(err)
		return nil, err
	}

	matched := categoryExtractReg.FindSubmatch([]byte(respBody))
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respBody)
		//	return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}

	var viewData categoryStructure
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		// return err
	}

	var (
		cates   []*pbItem.Category
		cateMap = map[string]*pbItem.Category{}
	)
	if err := func(yield func(names []string, url string) error) error {

		for _, nodes := range viewData.Props.PageProps.HeaderNavigation {

			cateName := nodes.Link.Title

			if cateName == "" {
				continue
			}

			for _, rawcat := range nodes.Items {

				subcat2 := rawcat.Link.Title

				if subcat2 == "" {
					continue
				}

				for _, rawcatdata := range rawcat.Items {

					subcat3 := rawcatdata.Link.Title

					href, err := c.CanonicalUrl(rawcatdata.Link.Href)
					if rawcatdata.Link.Href == "" || subcat3 == "" || err != nil {
						continue
					}

					u, err := url.Parse(href)
					if err != nil {
						c.logger.Errorf("parse url %s failed", href)
						continue
					}

					if c.categoryPathMatcher.MatchString(u.Path) {
						if err := yield([]string{cateName, subcat2, subcat3}, href); err != nil {
							return err
						}

					}
				}

				if len(rawcat.Items) == 0 {

					href, err := c.CanonicalUrl(rawcat.Link.Href)
					if rawcat.Link.Href == "" || err != nil || strings.ToLower(subcat2) == "for you" {
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

type categoryStructure struct {
	Props struct {
		PageProps struct {
			InternalRoutes     []string      `json:"internalRoutes"`
			ExperimentalRoutes []interface{} `json:"experimentalRoutes"`
			HeaderNavigation   []struct {
				Link struct {
					Title    string `json:"title"`
					TitleEn  string `json:"titleEn"`
					Href     string `json:"href"`
					External bool   `json:"external"`
				} `json:"link"`
				Items []struct {
					Link struct {
						Title    string `json:"title"`
						TitleEn  string `json:"titleEn"`
						Href     string `json:"href"`
						External bool   `json:"external"`
					} `json:"link"`
					Items []struct {
						Link struct {
							Title    string `json:"title"`
							TitleEn  string `json:"titleEn"`
							Href     string `json:"href"`
							External bool   `json:"external"`
						} `json:"link"`
					} `json:"items"`
				} `json:"items,omitempty"`
				Media string `json:"media,omitempty"`
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
			//Num80025371 struct {
			ID  string `json:"id"`
			URL string `json:"url"`
			//} `json:"80025371"`
		} `json:"productCards"`

		Pages map[string]struct {
			//WomensCoats struct {
			ID             string `json:"id"`
			NavigationID   string `json:"navigationId"`
			CuratedProduct string `json:"curatedProduct"`
			ProductCount   int    `json:"productCount"`
			ProductsURL    string `json:"productsUrl"`
			URL            string `json:"url"`
			QueryParams    struct {
				Country  string `json:"country"`
				Language string `json:"language"`
			} `json:"queryParams"`
			PresaleRedirectURL string `json:"presaleRedirectUrl"`
			//} `json:"/womens-coats/"`
		} `json:"pages"`
	} `json:"db"`

	Data struct {
		Entities struct {
			ProductCards map[string]struct {
				//Num80025371 struct {
				ID  string `json:"id"`
				URL string `json:"url"`
				//} `json:"80025371"`
			} `json:"productCards"`
		} `json:"entities"`
		Result     []string    `json:"result"`
		ResponseID interface{} `json:"responseId"`
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

	lastIndex := nextIndex(ctx)

	var viewData categoryProductStructure
	matched := categoryProductReg.FindSubmatch([]byte(respBody))
	if len(matched) > 1 {
		if err := json.Unmarshal(matched[1], &viewData); err != nil {
			c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
			return err
		}
	} else {
		if err := json.Unmarshal(respBody, &viewData); err != nil {
			c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
			return err
		}
	}

	r := viewData.Db.ProductCards
	if len(viewData.Db.ProductCards) == 0 {
		r = viewData.Data.Entities.ProductCards
	}

	for _, rawcat := range r {

		href, err := c.CanonicalUrl(rawcat.URL)
		if href == "" || rawcat.URL == "" || err != nil {
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

	nextUrl := ""

	for _, raw := range viewData.Db.Pages {
		if raw.ProductCount <= lastIndex {
			return nil
		}
		nextUrl = `https://us.burberry.com/web-api/pages/products?location=/` + raw.NavigationID + `&offset=` + strconv.Format(lastIndex) + `&limit=1000&order_by=&pagePath=` + raw.ProductsURL + `&country=US&language=en`
	}
	if len(viewData.Db.Pages) == 0 && len(viewData.Data.Entities.ProductCards) == 100 {

		u := *resp.Request.URL
		vals := u.Query()
		vals.Set("offset", strconv.Format(lastIndex))
		u.RawQuery = vals.Encode()
		nextUrl = u.String()
	}

	req, _ := http.NewRequest(http.MethodGet, nextUrl, nil)
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)

}

// used to trim html labels in description
var htmlTrimRegp = regexp.MustCompile(`</?[^>]+>`)

type productStructure struct {
	Db struct {
		Pages map[string]struct {
			//TheMidLengthChelseaHeritageTrenchCoatP40733751 struct {
			ID   string `json:"id"`
			Data struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
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
				Content struct {
					Title       string `json:"title"`
					Description string `json:"description"`
					Features    []struct {
						Text string `json:"text"`
					} `json:"features"`
				} `json:"content"`
				GalleryItems []struct {
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
					Label         string `json:"label"`
					StockQuantity int    `json:"stockQuantity"`
					IsInStock     bool   `json:"isInStock"`
					Sku           string `json:"sku"`
				} `json:"sizes"`

				SwatchItems []struct {
					ID         string `json:"id"`
					URL        string `json:"url"`
					Image      string `json:"image"`
					Label      string `json:"label"`
					IsSelected bool   `json:"isSelected"`
				} `json:"swatchItems"`
				ProductFamilyItems []struct {
					ID  string `json:"id"`
					URL string `json:"url"`
					Key string `json:"key"`
				} `json:"productFamilyItems"`
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

		desc := htmlTrimRegp.ReplaceAllString(prod.Data.Content.Description, ``) + " "
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

			if rawsku.IsInStock {
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

		// other products
		if ctx.Value("groupId") == nil {
			nctx := context.WithValue(ctx, "groupId", item.GetSource().GetId())
			for _, colorSizeOption := range prod.Data.ProductFamilyItems {

				if colorSizeOption.ID == prod.ID {
					continue
				}

				nextProductUrl, _ := c.CanonicalUrl(colorSizeOption.URL)
				if req, err := http.NewRequest(http.MethodGet, nextProductUrl, nil); err != nil {
					return err
				} else if err = yield(nctx, req); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://us.burberry.com/",
		//"https://us.burberry.com/womens-coats/",
		//"https://us.burberry.com/womens-trench-coats/",
		//"https://us.burberry.com/womens-coats-jackets/",
		//"https://us.burberry.com/girl/",
		//"https://us.burberry.com/tb-summer-monogram-collection-women/",
		//"https://us.burberry.com/small-leather-tb-bag-p80345521",
		//"https://us.burberry.com/monogram-print-nylon-bucket-hat-online-exclusive-p80502851",
		//"https://us.burberry.com/canvas-leathersuede-sneakers-p80421971",
		//"https://us.burberry.com/bold-lash-mascara-chestnut-brown-no02-p39544251",
		//"https://us.burberry.com/my-burberry-blush-eau-de-parfum-90ml-p40493291",
		//"https://us.burberry.com/bold-lash-mascara-chestnut-brown-no02-p39544251",
		//"https://us.burberry.com/the-long-chelsea-heritage-trench-coat-p40733791",
		"https://us.burberry.com/the-long-chelsea-heritage-trench-coat-p80279981",
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
