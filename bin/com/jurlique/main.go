package main

import (
	"bytes"
	"context"
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
//func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
func (_ *_Crawler) New(_ *cli.Context, client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient: client,
		// this regular used to match category page url path
		categoryPathMatcher: regexp.MustCompile(`^(/us([/A-Za-z0-9_-]+))$`),
		productPathMatcher:  regexp.MustCompile(`^(/us[/A-Za-z0-9_-]+.html)`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "456e2e193aef4aa08ef2bff592341f9b"
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
		KeepSession:       false,
		Reliability:       proxy.ProxyReliability_ReliabilityIntelligent,
		MustHeader:        crawler.NewCrawlOptions().MustHeader,
	}
	opts.MustHeader.Add(`Cookie`, `dwanonymous_a3e8e72baadb8c7db8d6e0e51d5d8b06=ada1g6AJGWWRyBC73NbY6rvDxg; _ga=GA1.2.271018505.1630392405; _fbp=fb.1.1630392407976.1542890411; recordID=4dba3b5a-598f-408c-8f00-e9e3f5a80d6f; BVBRANDID=616e4ac8-dcb8-4793-b823-e08f814b7c8e; _hjid=e1588610-98df-46b8-b0b3-c696c3bece82; newsletter_welcome_popup=true; isTrackingConsentedByCustomer=false; dwanonymous_bfdd72b037a890b648fa70a6f737ee44=abv76awzMfR8k6J1qM5Bf2Dk33; isRedirected_jurlique-hk=Yes; _hjDonePolls=653255; dwanonymous_83dc38f9c0e8771277ca4c4977e0234c=bccNK7F8s3MapUjHVxlcF79Yso; isRedirected_jurlique-uk=Yes; accessedSiteCountryCode=""; stc120301=env:1|20211002053936|20210901060936|1|1098173:20220901053936|uid:!anon!:20220901053936|srchist:1098173:1:20211002053936:20220901053936|a-ldt:1630474776846:20210901060936|tsa:1630474776847.893419687.3292747.39277173291785905.:20210901060936; __cq_dnt=1; dw_dnt=1; _gid=GA1.2.2094442184.1631100048; tfa_tra_src=Other; sid=yQtywFlh3ly63eMgTUrW85-BzUIYdWgXuug; dwsid=CvDqBLBybaFL9FLSmQ2dA1x6VYbxZtvqS8FUBOGs1TXWA6z-WSxF3nOi62yfeLR4WjZsqDHGjMjifg6EPUP3Rw==; dmSessionID=f0dafaf0-975c-4546-aa07-3d85f0d5247a; _hjIncludedInPageviewSample=1; _hjAbsoluteSessionInProgress=0; _hjIncludedInSessionSample=0; BVBRANDSID=833440d7-7d4d-4f08-ae4d-48c082e52219; ts_uid=a1770e7cb0b31686a4a3490501; isCountrySelector=No; stc111613=env:1|20211010055955|20210909062955|1|1014204:20220909055955|uid:!anon!:20220909055955|srchist:1014204:1:20211001121019|1014203:1:20211002035701|1014204:1:20211010055955:20220909055955|a-rfd:www.jurlique.com:20220909060555|a-ldt:1631167195492:20210909062955|tsa:1631167195493.2028449852.2317781.14054785900666333:20210909062955; cto_bundle=9Z119l9oYU9yJTJGelg5NmVSWllROFA5UFlSc1olMkYzTnlOd2JwVENKS1VxeDdRbWp6JTJCb0lzOE5VcnVPZUdIb01NTGZLZ2VsWTNJemRMSXg0TDdoVVVRSzZRbW50a1h4R3ZlQkpvNE8lMkYlMkZFR2wlMkZtdWtVN0RkZFhDc24lMkZTaXkwZGZJSUp6R0dmUEUxTzJIVW5EeTVQSVR5MU43d1ZRdyUzRCUzRA; _uetsid=dd3ab4f0109611ec853b71eee981ac98; _uetvid=35f9a0a00a2711ec9ab19db092971826`)
	//	opts.MustCookies = append(opts.MustCookies,
	//	&http.Cookie{Name: "GlobalE_Data", Value: `{"countryISO":"US","cultureCode":"en-US","currencyCode":"USD","apiVersion":"2.1.4"}`, Path: "/"},
	//	&http.Cookie{Name: "_dy_geo", Value: "US.NA.US_DC.US_DC_Washington", Path: "/"},
	//	)
	return opts
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.jurlique.com"}
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
		u.Host = "www.jurlique.com"
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
	if p == "us/" {
		//return c.parseCategories(ctx, resp, yield)
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
	req, _ := http.NewRequest(http.MethodGet, "https://www.jurlique.com/", nil)
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

		sel := dom.Find(`.nav.navbar-nav.megamenu__list .nav-item.megamenu__item.dropdown.js-navbar-item`)

		for i := range sel.Nodes {
			node := sel.Eq(i)

			cateName := strings.TrimSpace(node.Find(`a`).First().Text())
			if cateName == "" {
				continue

			}

			subSel := node.Find(`.dropdown-menu.dropdown-menu__level-second.js-navbar-sublist>li`)
			if len(subSel.Nodes) == 0 {
				subSel = node.Find(`.dropdown-menu__container.container>li`)
			}
			for k := range subSel.Nodes {
				subNode2 := subSel.Eq(k)

				subcat2 := strings.TrimSpace(subNode2.Find(`.dropdown-link.dropdown-menu__level-second__category-link`).First().Text())
				if subcat2 == "" {
					subcat2 = strings.TrimSpace(subNode2.Find(`a`).Last().Text())
				}

				subNode2list := subNode2.Find(`.dropdown-menu__level-thirds>li`)
				for j := range subNode2list.Nodes {
					subNode3 := subNode2list.Eq(j)
					subcat3 := strings.TrimSpace(subNode3.Find(`.dropdown-link`).First().Text())

					href := subNode3.Find(`a`).AttrOr("href", "")
					if href == "" || subcat3 == "" {
						continue
					}

					u, err := url.Parse(href)
					if err != nil {
						c.logger.Error("parse url %s failed", href)
						continue
					}

					if c.categoryPathMatcher.MatchString(u.Path) {
						if !strings.Contains(href, ".jurlique.com") {
							href = "https://www.jurlique.com" + href
						}
						if err := yield([]string{cateName, subcat2, subcat3}, href); err != nil {
							return err
						}
					}
				}

				if len(subNode2list.Nodes) == 0 {
					href := subNode2.Find(`a`).First().AttrOr("href", "")
					if href == "" {
						continue
					}

					u, err := url.Parse(href)
					if err != nil {
						c.logger.Error("parse url %s failed", href)
						continue
					}

					if c.categoryPathMatcher.MatchString(u.Path) {

						if !strings.Contains(href, ".jurlique.com") {
							href = "https://www.jurlique.com" + href
						}
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

	sel := doc.Find(`.image-container`)

	for i := range sel.Nodes {
		node := sel.Eq(i)
		if href, _ := node.Find(`a`).Attr("href"); href != "" {
			fmt.Println(lastIndex)
			// req, err := http.NewRequest(http.MethodGet, href, nil)
			// if err != nil {
			// 	c.logger.Error(err)
			// 	continue
			// }
			lastIndex += 1
			// nctx := context.WithValue(ctx, "item.index", lastIndex)
			// if err := yield(nctx, req); err != nil {
			// 	return err
			// }
		}

	}
	nextUrl := doc.Find(`.btn.btn-block.btn-secondary-main`).AttrOr(`data-url`, ``)

	if nextUrl ==
		"" {
		return nil
	}

	//nextUrl = strings.ReplaceAll(nextUrl, "&start=12", "&sz=96")

	req, _ := http.NewRequest(http.MethodGet, nextUrl, nil)
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
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

var productsReviewExtractReg = regexp.MustCompile(`(?Ums)<script type="application/ld\+json">\s*({.*})\s*</script>`)

// used to trim html labels in description
var htmlTrimRegp = regexp.MustCompile(`</?[^>]+>`)

// parseProduct
func (c *_Crawler) parseProduct(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil {
		return nil
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	ioutil.WriteFile("C:\\Maulika\\Output.html", respBody, 0644)
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	canUrl, _ := c.CanonicalUrl(doc.Find(`link[rel="canonical"]`).AttrOr("href", ""))
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}

	reviewCount, _ := strconv.ParseInt(doc.Find(`.bvseo-reviewCount`).Text())
	rating, _ := strconv.ParseFloat(doc.Find(`.bvseo-ratingValue`).Text())

	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           doc.Find(`.product`).AttrOr("data-pid", ""),
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
			//GroupId:      doc.Find(`meta[property="product:age_group"]`).AttrOr(`content`, ``),
		},
		BrandName: "Jurlique",
		Title:     doc.Find(`meta[property="og:title"]`).AttrOr(`content`, ``),
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stats: &pbItem.Stats{
			ReviewCount: int32(reviewCount),
			Rating:      float32(rating),
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
	}

	// desc
	description := doc.Find(`.details`).Text()
	item.Description = string(TrimSpaceNewlineInString([]byte(description)))

	// 最后计算库存状态

	currentPrice, _ := strconv.ParsePrice(doc.Find(`.container.product-detail-container.js-sticky-column`).Find(`.sales`).Text())
	msrp := float64(0)

	if msrp == 0 {
		msrp = currentPrice
	}
	discount := int32(0)
	if msrp > currentPrice {
		discount = int32(((msrp - currentPrice) / msrp) * 100)
	}

	//images
	sel := doc.Find(`.product-thumbnail.js-carousel`).Find(`div`)
	for j := range sel.Nodes {
		node := sel.Eq(j)
		imgurl := strings.Split(node.Find(`img`).AttrOr(`src`, ``), "?")[0]

		item.Medias = append(item.Medias, pbMedia.NewImageMedia(
			strconv.Format(j),
			imgurl,
			imgurl+"?sw=1000&sh=1000&q=80",
			imgurl+"?sw=600&sh=600&q=80",
			imgurl+"?sw=500&sh=500&q=80",
			"", j == 0))
	}

	// itemListElement
	sel = doc.Find(`.breadcrumb>li`)
	c.logger.Debugf("nodes %d", len(sel.Nodes))
	for i := range sel.Nodes {
		node := sel.Eq(i)
		breadcrumb := strings.TrimSpace(node.Text())

		if i == 0 {
			item.Category = breadcrumb
		} else if i == 1 {
			item.SubCategory = breadcrumb
		} else if i == 2 {
			item.SubCategory2 = breadcrumb
		} else if i == 3 {
			item.SubCategory3 = breadcrumb
		} else if i == 4 {
			item.SubCategory4 = breadcrumb
		}
	}

	cid := ""
	cid = doc.Find(`.container.product-detail-container.js-sticky-column`).Find(`.add-to-cart.btn.btn-block.btn-primary-main.shipto_Yes`).AttrOr(`data-pid`, "")
	sid := doc.Find(`.one-size-variation`).First().Text()
	sku := pbItem.Sku{
		SourceId: fmt.Sprintf("%s-%s", cid, sid),
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
			Current:  int32(currentPrice),
			Msrp:     int32(msrp),
			Discount: discount,
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_InStock},
	}

	if strings.Contains(doc.Find(`.availability.col-12.product-availability`).AttrOr(`data-available`, ``), "true") {
		item.Stock.StockStatus = pbItem.Stock_InStock
	}

	// size
	sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
		Type:  pbItem.SkuSpecType_SkuSpecSize,
		Id:    sid,
		Name:  sid,
		Value: sid,
	})

	item.SkuItems = append(item.SkuItems, &sku)

	if sid == "" {

		sel = doc.Find(`.col-12.col-sm-6.col-lg-4.d-none.d-lg-block.product-detail__right-side`).Find(`.custom-select.form-control.select-size`).Find(`option`)

		for i := range sel.Nodes {

			i = i + 1

			node := sel.Eq(i)

			sid = node.AttrOr("data-attr-value", "")

			if sid == "" {
				continue
			}
			sku := pbItem.Sku{
				SourceId: fmt.Sprintf("%s-%s", cid, sid),
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(currentPrice),
					Msrp:     int32(msrp),
					Discount: discount,
				},
				Stock: &pbItem.Stock{StockStatus: pbItem.Stock_InStock},
			}

			if strings.Contains(doc.Find(`.availability.col-12.product-availability`).AttrOr(`data-available`, ``), "true") {
				item.Stock.StockStatus = pbItem.Stock_InStock
			}

			// size
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    sid,
				Name:  sid,
				Value: sid,
			})

			item.SkuItems = append(item.SkuItems, &sku)
		}
	}

	// yield item result
	if err = yield(ctx, &item); err != nil {
		c.logger.Errorf("yield sub request failed, error=%s", err)
		return err
	}

	return nil
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://www.jurlique.com/us/en/homepage",
		//"https://www.jurlique.com/us/face/by-category/shop-all-face-care",
		"https://www.jurlique.com/us/calendula-redness-rescue-soothing-moisturising-cream-CRRC.html",
		//"https://www.jurlique.com/us/moisture-plus-rare-rose-gel-cream-RMPRRG.html",
		//"https://www.jurlique.com/us/rose-love-balm-R09.html",
		//"https://www.jurlique.com/us/jojoba-carrier-oil-J01.html",
		//"https://www.jurlique.com/us/rose-silk-finishing-powder-R03.html",
		//"https://www.jurlique.com/us/herbal-recovery-signature-serum-HRS03.html",
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
	os.Setenv("VOILA_PROXY_URL", "http://52.207.171.114:30216")
	cli.NewApp(&_Crawler{}).Run(os.Args)
}
