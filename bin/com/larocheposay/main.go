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
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/voiladev/VoilaCrawler/pkg/cli"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	pbMedia "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	pbProxy "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
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
func (_ *_Crawler) New(_ *cli.Context, client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient: client,
		// this regular used to match category page url path
		categoryPathMatcher: regexp.MustCompile(`^/([/a-zA-Z0-9\-]+)$`),
		// this regular used to match product page url path
		productPathMatcher: regexp.MustCompile(`^(.*).html$`),
		logger:             logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "bf8f094546d142a2aab6f6cea8c034b4"
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
		EnableHeadless:    false,
		EnableSessionInit: false,
		Reliability:       pbProxy.ProxyReliability_ReliabilityLow,
		MustHeader:        make(http.Header),
	}

	return opts
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.laroche-posay.us"}
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
		u.Host = "www.laroche-posay.us"
	}
	if c.productPathMatcher.MatchString(u.Path) {
		u.RawQuery = ""
		return u.String(), nil
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
		return c.parseCategories(ctx, resp, yield)
	}

	if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yield)
	} else if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	}
	return crawler.ErrUnsupportedPath
}

func (c *_Crawler) GetCategories(ctx context.Context) ([]*pbItem.Category, error) {
	req, _ := http.NewRequest(http.MethodGet, "https://www.laroche-posay.us/", nil)
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
		sel := dom.Find(`.c-navigation > ul > li`)
		for i := range sel.Nodes {
			node := sel.Eq(i)
			cateName := strings.TrimSpace(node.Find(`.c-navigation__item-title`).Find(`a`).First().Text())
			if cateName == "" {
				continue
			}

			subSel := node.Find(`.c-navigation__flyout-element > .m-level-2 > .c-navigation__item`)

			if len(subSel.Nodes) > 0 {

				for j := range subSel.Nodes {

					subNode := subSel.Eq(j)

					subCateName := strings.TrimSpace(subNode.Find(`.c-navigation__item-title`).Find(`a`).First().Text())

					if subCateName == "View all" {
						continue
					}

					subSel1 := subNode.Find(`.m-level-3 > .c-navigation__item`)

					if len(subSel1.Nodes) > 0 {

						for k := range subSel1.Nodes {

							subNode1 := subSel1.Eq(k)

							lastcat := strings.TrimSpace(subNode1.Find(`.c-navigation__item-title`).Find(`a`).First().Text())

							href := subNode1.Find(`.c-navigation__item-title`).Find(`a`).First().AttrOr("href", "")

							if href == "" {
								continue
							}

							u, err := url.Parse(href)
							if err != nil {
								c.logger.Error("parse url %s failed", href)
								continue
							}

							if !c.productPathMatcher.MatchString(u.Path) && c.categoryPathMatcher.MatchString(u.Path) {
								if err := yield([]string{cateName, subCateName, lastcat}, href); err != nil {
									return err
								}
							}
						}
					} else {

						subCateName := strings.TrimSpace(subNode.Find(`.c-navigation__item-title`).Find(`a`).First().Text())

						href := subNode.Find(`.c-navigation__item-title`).Find(`a`).First().AttrOr("href", "")

						if href == "" {
							continue
						}

						u, err := url.Parse(href)
						if err != nil {
							c.logger.Error("parse url %s failed", href)
							continue
						}

						if !c.productPathMatcher.MatchString(u.Path) && c.categoryPathMatcher.MatchString(u.Path) {
							if err := yield([]string{cateName, subCateName}, href); err != nil {
								return err
							}
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

// @deprecated
func (c *_Crawler) parseCategories(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	sel := dom.Find(`.c-navigation > ul > li`)
	for i := range sel.Nodes {
		node := sel.Eq(i)
		cateName := strings.TrimSpace(node.Find(`.c-navigation__item-title`).Find(`a`).First().Text())
		if cateName == "" {
			continue
		}
		//nnctx := context.WithValue(ctx, "Category", cateName)

		subSel := node.Find(`.c-navigation__flyout-element > .m-level-2 > .c-navigation__item`)

		if len(subSel.Nodes) > 0 {

			for j := range subSel.Nodes {

				subNode := subSel.Eq(j)

				subCateName := strings.TrimSpace(subNode.Find(`.c-navigation__item-title`).Find(`a`).First().Text())

				// if subCateName == "View all" {
				// 	continue
				// }

				subSel1 := subNode.Find(`.m-level-3 > .c-navigation__item`)

				if len(subSel1.Nodes) > 0 {

					for k := range subSel1.Nodes {

						subNode1 := subSel1.Eq(k)

						lastcat := strings.TrimSpace(subNode1.Find(`.c-navigation__item-title`).Find(`a`).First().Text())

						if lastcat == "View all" {
							continue
						}

						href := subNode1.Find(`.c-navigation__item-title`).Find(`a`).First().AttrOr("href", "")
						if href == "" {
							continue
						}

						_, err := url.Parse(href)
						if err != nil {
							c.logger.Error("parse url %s failed", href)
							continue
						}

						fmt.Println(cateName, " > ", subCateName, " > ", lastcat)

						// nnnctx := context.WithValue(nnctx, "SubCategory", subCateName)
						// req, _ := http.NewRequest(http.MethodGet, href, nil)
						// if err := yield(nnnctx, req); err != nil {
						// 	return err
						// }
					}
				}
			}
		}
	}
	return nil
}

// nextIndex used to get the index from the shared data.
// item.index is a const key for item index.
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")))
}

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

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Debug(err)
		return err
	}

	lastIndex := nextIndex(ctx)

	sel := doc.Find(`.l-plp__product-results`).Find(`.c-product-tile__name`)

	for i := range sel.Nodes {

		node := sel.Eq(i)

		if href, _ := node.Find(`a`).Attr("href"); href != "" {

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

	nextUrl := doc.Find(`.c-load-more__button`).AttrOr("href", "")
	if nextUrl == "" {
		return nil
	}

	nextUrl = strings.ReplaceAll(nextUrl, `&sz=6`, `&sz=60`)

	req, _ := http.NewRequest(http.MethodGet, nextUrl, nil)
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
}

type parseProductResponse struct {
	Context     string `json:"@context"`
	Type        string `json:"@type"`
	Sku         string `json:"sku"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Brand       struct {
		Type string `json:"@type"`
		Name string `json:"name"`
	} `json:"brand"`
	Offers struct {
		Type          string  `json:"@type"`
		Price         float64 `json:"price"`
		PriceCurrency string  `json:"priceCurrency"`
		Availability  string  `json:"availability"`
		URL           string  `json:"url"`
	} `json:"offers"`
	ID string `json:"@id"`
}

type parseProductVariantsResponse struct {
	Productmedia struct {
		Items []struct {
			Image         string `json:"image,omitempty"`
			ImageIndex    int    `json:"imageIndex,omitempty"`
			ImageTypeUsed string `json:"imageTypeUsed,omitempty"`
			VideoID       struct {
				DataComponent    string `json:"dataComponent"`
				ComponentOptions struct {
					FrontendComponent string `json:"frontendComponent"`
					Loop              int    `json:"loop"`
					Mute              int    `json:"mute"`
					ImageOptions      struct {
						ViewType string `json:"viewType"`
					} `json:"imageOptions"`
					VideoID      string `json:"videoId"`
					VideoSources []struct {
						Src  string `json:"src"`
						Type string `json:"type"`
					} `json:"videoSources"`
					Breakpoints struct {
					} `json:"breakpoints"`
					A11Y struct {
						ControlAriaLabelStop string `json:"controlAriaLabelStop"`
						ControlAriaLabelPlay string `json:"controlAriaLabelPlay"`
					} `json:"a11y"`
				} `json:"componentOptions"`
				Cover                     string    `json:"cover"`
				URL                       string    `json:"url"`
				PosterURL                 string    `json:"posterUrl"`
				ThumbnailURL              string    `json:"thumbnailUrl"`
				HideDescription           bool      `json:"hideDescription"`
				Title                     string    `json:"title"`
				TitleTagName              string    `json:"titleTagName"`
				Description               string    `json:"description"`
				UploadDate                time.Time `json:"uploadDate"`
				AriaLabel                 string    `json:"ariaLabel"`
				ShowAccessibilityControl  bool      `json:"showAccessibilityControl"`
				AccessibilityControlClass string    `json:"accessibilityControlClass"`
				CSSClasses                struct {
					AssetLink string `json:"assetLink"`
					InfoName  string `json:"infoName"`
				} `json:"cssClasses"`
				Text struct {
					AutoplayAriaLabelStop string `json:"autoplayAriaLabelStop"`
				} `json:"text"`
			} `json:"videoID,omitempty"`
			IsVideo bool `json:"isVideo,omitempty"`
		} `json:"items"`
	} `json:"productmedia"`
	Analytics struct {
		Products []struct {
			Pid         string `json:"pid"`
			Title       string `json:"title"`
			Description struct {
			} `json:"description"`
			URL                       string      `json:"url"`
			ImgURL                    string      `json:"imgUrl"`
			Currency                  string      `json:"currency"`
			Price                     float64     `json:"price"`
			Name                      string      `json:"name"`
			ID                        string      `json:"id"`
			SalePrice                 float64     `json:"salePrice"`
			Brand                     string      `json:"brand"`
			Category                  string      `json:"category"`
			ProductTopCategory        string      `json:"productTopCategory"`
			Variant                   string      `json:"variant"`
			Size                      string      `json:"size"`
			Color                     string      `json:"color"`
			Fragrance                 string      `json:"fragrance"`
			Stock                     string      `json:"stock"`
			AutoReplenishmentInterval string      `json:"autoReplenishmentInterval"`
			Upc                       string      `json:"upc"`
			RegularPrice              interface{} `json:"regularPrice"`
			IsProductSet              bool        `json:"isProductSet"`
			IsProductGroup            bool        `json:"isProductGroup"`
			IsBundle                  bool        `json:"isBundle"`
			BundleID                  string      `json:"bundleID"`
			Rating                    string      `json:"rating"`
			NumberReviews             int         `json:"numberReviews"`
			VtoState                  string      `json:"vtoState"`
			Collection                []string    `json:"collection"`
		} `json:"products"`
	} `json:"analytics"`
}

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
	if err := json.Unmarshal([]byte(doc.Find(`script[type="application/ld+json"]`).Last().Text()), &viewData); err != nil {
		c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		return err
	}

	canUrl, _ := c.CanonicalUrl(doc.Find(`link[rel="canonical"]`).AttrOr("href", ""))
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}

	rating, _ := strconv.ParseFloat(doc.Find(`.bvseo-ratingValue`).Text())
	reviewcount, _ := strconv.ParseInt(doc.Find(`.bvseo-reviewCount`).Text())

	currentPrice, _ := strconv.ParsePrice(viewData.Offers.Price)
	msrp, _ := strconv.ParsePrice(viewData.Offers.Price)

	if msrp == 0 {
		msrp = currentPrice
	}

	pid := strings.Split(resp.Request.URL.Path, "-")

	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           strings.TrimSuffix(pid[len(pid)-1], `.html`),
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
			//GroupId:      doc.Find(`meta[property="product:age_group"]`).AttrOr(`content`, ``),
		},
		BrandName: "La Roche-Posay",
		Title:     viewData.Name,
		Stats: &pbItem.Stats{
			ReviewCount: int32(reviewcount),
			Rating:      float32(rating),
		},
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
	}

	item.Description = viewData.Description

	if strings.Contains(viewData.Offers.Availability, "InStock") {
		item.Stock.StockStatus = pbItem.Stock_InStock
	}

	// itemListElement
	sel := doc.Find(`.c-breadcrumbs > ol > li`)
	c.logger.Debugf("nodes %d", len(sel.Nodes))
	for i := range sel.Nodes {
		if i >= len(sel.Nodes)-1 {
			continue
		}
		node := sel.Eq(i).Find(`span`)
		breadcrumb := strings.TrimSpace(node.Text())

		if i == 1 {
			item.Category = breadcrumb
		} else if i == 2 {
			item.SubCategory = breadcrumb
		} else if i == 3 {
			item.SubCategory2 = breadcrumb
		} else if i == 4 {
			item.SubCategory3 = breadcrumb
		} else if i == 5 {
			item.SubCategory4 = breadcrumb
		}
	}

	sel = doc.Find(`.m-variations`).Find(`.c-carousel__content>div`)
	for i := range sel.Nodes {
		node := sel.Eq(i)

		currentPrice, _ = strconv.ParsePrice(node.Find(`.c-product-price__value`).Text())
		msrp, _ = strconv.ParsePrice(node.Find(`.c-product-price__value`).Text())

		if msrp == 0 {
			msrp = currentPrice
		}
		discount := 0
		if msrp > currentPrice {
			discount = (int)(((msrp - currentPrice) / msrp) * 100)
		}

		sid := (node.Find(`a`).AttrOr("data-js-pid", ""))
		if sid == "" {
			continue
		}

		sName := (node.Find(`a`).AttrOr("data-js-value", ""))

		sku := pbItem.Sku{
			SourceId: sid,
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(currentPrice * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},
			Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}

		colorName := ""
		if !strings.Contains(node.Find(`a`).AttrOr(`class`, ``), `m-selected`) {

			// request new for image
			variantURL := node.Find(`a`).AttrOr(`href`, ``) + "&ajax=true"
			respBodyV := c.VariationRequest(ctx, variantURL, resp.Request.URL.String())

			var viewDataImage parseProductVariantsResponse
			if err := json.Unmarshal([]byte(respBodyV), &viewDataImage); err != nil {
				//c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
				fmt.Println(err)
			}

			for m, mid := range viewDataImage.Productmedia.Items {

				domI, _ := goquery.NewDocumentFromReader(bytes.NewReader([]byte(mid.Image)))

				if mid.IsVideo {
					cover := mid.VideoID.ThumbnailURL
					if cover != "" && strings.HasPrefix(cover, "//") {
						cover = "https:" + mid.VideoID.ThumbnailURL
					}

					sku.Medias = append(sku.Medias, pbMedia.NewVideoMedia(
						strconv.Format(m),
						"", mid.VideoID.URL,
						0, 0, 0, cover, "",
						m == 0))
				}

				imgurl := strings.Split(domI.Find(`img`).AttrOr(`src`, ``), "?")[0]
				if imgurl == "" {
					continue
				}
				sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
					strconv.Format(m),
					imgurl,
					imgurl+"?sw=1000&sfrm=jpg&q=70",
					imgurl+"?sw=800&sfrm=jpg&q=70",
					imgurl+"?sw=500&sfrm=jpg&q=70",
					"",
					m == 0,
				))
			}

			if viewDataImage.Analytics.Products[0].Stock == "in stock" {
				sku.Stock.StockStatus = pbItem.Stock_InStock
			}

			if viewDataImage.Analytics.Products[0].Color != "" {
				colorName = viewDataImage.Analytics.Products[0].Color
			}

			if viewDataImage.Analytics.Products[0].Size != "" {
				sName = viewDataImage.Analytics.Products[0].Size
			}

			currentPrice, _ = strconv.ParsePrice(viewDataImage.Analytics.Products[0].SalePrice)
			msrp, _ = strconv.ParsePrice(viewDataImage.Analytics.Products[0].RegularPrice)

			if msrp == 0 {
				msrp = currentPrice
			}
			discount = 0
			if msrp > currentPrice {
				discount = (int)(((msrp - currentPrice) / msrp) * 100)
			}
			sku.Price.Current = int32(currentPrice * 100)
			sku.Price.Msrp = int32(msrp * 100)
			sku.Price.Discount = int32(discount)

		} else {

			//images
			sel := doc.Find(`.c-product-detail-image__image-link`)
			for m := range sel.Nodes {
				node := sel.Eq(m)
				imgurl := strings.Split(node.Find(`img`).AttrOr(`src`, ``), "?")[0]
				if imgurl == "" {
					continue
				}
				sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
					strconv.Format(len(sku.Medias)),
					imgurl,
					imgurl+"?sw=1000&sfrm=jpg&q=70",
					imgurl+"?sw=800&sfrm=jpg&q=70",
					imgurl+"?sw=500&sfrm=jpg&q=70",
					"", len(sku.Medias) == 0))
			}

			// video
			sel = doc.Find(`.c-product-detail-image__alternatives`).Find(`.c-video-asset__link`)
			for m := range sel.Nodes {
				node := sel.Eq(m)
				cover := node.Find(`img`).First().AttrOr("data-src", "")
				if cover != "" && strings.HasPrefix(cover, "//") {
					cover = "https:" + strings.ReplaceAll(node.Find(`img`).AttrOr("data-src", ""), `&sh=0`, ``)
				}

				videourl := node.AttrOr(`data-url`, ``)
				if videourl == "" {
					continue
				}
				sku.Medias = append(sku.Medias, pbMedia.NewVideoMedia(
					strconv.Format(len(sku.Medias)),
					"", videourl,
					0, 0, 0, cover, "",
					len(sku.Medias) == 0))
			}

			if strings.Contains(viewData.Offers.Availability, "InStock") {
				sku.Stock.StockStatus = pbItem.Stock_InStock
			}
		}

		// Color
		if colorName != "" {
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    colorName,
				Name:  colorName,
				Value: colorName,
			})
		}

		// size
		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Id:    sName,
			Name:  sName,
			Value: sName,
		})

		item.SkuItems = append(item.SkuItems, &sku)

	}
	if len(sel.Nodes) == 0 {
		// no variation
		currentPrice, _ = strconv.ParsePrice(doc.Find(`.c-product-main__price`).Find(`.c-product-price__value`).Text())
		msrp, _ = strconv.ParsePrice(doc.Find(`.c-product-main__price`).Find(`.c-product-price__value`).Text())

		if msrp == 0 {
			msrp = currentPrice
		}
		discount := 0
		if msrp > currentPrice {
			discount = (int)(((msrp - currentPrice) / msrp) * 100)
		}

		sku := pbItem.Sku{
			SourceId: fmt.Sprintf("%s", viewData.Sku),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(currentPrice * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},
			Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}
		if strings.Contains(viewData.Offers.Availability, "https://schema.org/InStock") {
			sku.Stock.StockStatus = pbItem.Stock_InStock
		}

		//images
		sel := doc.Find(`.c-product-detail-image__image-link`)
		for m := range sel.Nodes {
			node := sel.Eq(m)
			imgurl := strings.Split(node.Find(`img`).AttrOr(`src`, ``), "?")[0]
			if imgurl == "" {
				continue
			}
			sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
				strconv.Format(m),
				imgurl,
				imgurl+"?sw=1000&sfrm=jpg&q=70",
				imgurl+"?sw=800&sfrm=jpg&q=70",
				imgurl+"?sw=500&sfrm=jpg&q=70",
				"", m == 0))
		}
		// video
		sel = doc.Find(`.c-product-detail-image__alternatives`).Find(`.c-video-asset__link `)
		for m := range sel.Nodes {
			node := sel.Eq(m)
			cover := node.Find(`img`).AttrOr("src", "")
			if cover != "" && strings.HasPrefix(cover, "//") {
				cover = "https:" + cover
			}

			videourl := node.AttrOr(`data-url`, ``)
			if videourl == "" {
				continue
			}
			sku.Medias = append(sku.Medias, pbMedia.NewVideoMedia(
				strconv.Format(m),
				"", videourl,
				0, 0, 0, cover, "",
				m == 0))
		}

		subTitle := strings.TrimSpace(doc.Find(`.c-product-main__subtitle`).Text())
		if subTitle == "" {
			subTitle = "-"
		}

		// Variants
		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecColor,
			Id:    subTitle,
			Name:  subTitle,
			Value: subTitle,
		})

		item.SkuItems = append(item.SkuItems, &sku)
	}

	// yield item result
	if err = yield(ctx, &item); err != nil {
		c.logger.Errorf("yield sub request failed, error=%s", err)
		return err
	}

	return nil
}
func (c *_Crawler) VariationRequest(ctx context.Context, rootUrl string, referer string) (reqs []byte) {

	req, _ := http.NewRequest(http.MethodGet, rootUrl, nil)
	opts := c.CrawlOptions(req.URL)
	req.Header.Set("accept-language", "en-GB,en-US;q=0.9,en;q=0.8")
	req.Header.Set("accept", "application/json")
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
		//return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)

	return respBody
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://www.laroche-posay.us/",
		//"https://www.laroche-posay.us/shop-by-concern/skin-concern",
		//"https://www.laroche-posay.us/our-products/face",
		//"https://www.laroche-posay.us/our-products/acne-oily-skin/face-wash/effaclar-gel-facial-wash-for-oily-skin-effaclargelcleanser.html",
		//"https://www.laroche-posay.us/our-products/anti-aging/anti-aging-moisturizer/active-vitamin-c-10%25-wrinkle-cream-3337872414053.html",
		//"https://www.laroche-posay.us/our-products/acne-oily-skin/face-wash/effaclar-micellar-water-for-oily-skin-effaclarmicellarwaterultra.html",
		//"https://www.laroche-posay.us/our-products/sun/body-sunscreen/anthelios-cooling-water-sunscreen-lotion-spf-60-3606000403826.html",
		"https://www.laroche-posay.us/our-products/anti-aging/anti-aging-serum/hyalu-b5-pure-hyaluronic-acid-serum-hyaluB5serum.html",
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
