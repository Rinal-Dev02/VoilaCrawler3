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
func (_ *_Crawler) New(_ *cli.Context, client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient: client,
		// this regular used to match category page url path
		categoryPathMatcher: regexp.MustCompile(`^(/us-en/(.*)/c/[/A-Za-z0-9_-]+)|(/us-en/[/A-Za-z0-9_-]+.html)$`),
		//productPathMatcher:  regexp.MustCompile(`^(/[/A-Za-z0-9_-]+.html)$`),
		productPathMatcher: regexp.MustCompile(`^(/us-en/[/A-Za-z0-9_-]+/p/[/A-Za-z0-9_-]+)|(/us-en/p/[/A-Za-z0-9_-]+)$`),

		logger: logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "ab77778901be4b0fa5ac62f2acd2921c"
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
		MustHeader:        crawler.NewCrawlOptions().MustHeader,
	}
	opts.MustHeader.Add(`cookie`, `JSESSIONID=B72373D29C88B101FC4CACEBBBC88A7F; countryIP=IN; geo-akamai=true; OptanonAlertBoxClosed=2021-09-09T10:11:08.500Z; _gcl_au=1.1.949434437.1631182269; _gid=GA1.2.1216332765.1631182269; _fbp=fb.1.1631182268928.920918955; cbar_uid=5597198485632; cbar_cart_checksum=0; _pin_unauth=dWlkPVl6RmhOV1UyWTJNdE56QmpPUzAwTWpBekxXRmxZbVl0WmpBd01qbGtZemt3WkdWaQ; ftr_ncd=6; counterNewsletterCookie=15; pageUrl=https://www.tods.com/us-en/tods-world/pre-fall21.html; newsletterClosed=true; customerData={"name":"Anonymous","userCode":"anonymous","countryCode":"US","countryName":"United+States","relativePath":"https://www.tods.com/","site":"tods-us","lingua":"en","isGuest":false,"cobrand":"","ctryLang":"us-en","currency":"USD","brochureSite":false,"visualSearch":true,"bookAnAppointment":true,"reserveInStore":true,"omnichannelStore":true,"sapCarEnabled":true,"asmEnabled":true,"newsletter":true,"counterNewsletterCookie":1,"rightToLeft":false,"algoliaSearch":true,"algoliaIndexName":"production_algoliaIndex_tods-us","algoliaApplicationID":"TM06P1CDJX","algoliaSearchAPIKey":"69755e034733f4af2cec64028cacda91","returnInStore":false}; AKA_A2=A; JSESSIONID=071C9B77E2B8EE9BD19838C02A18A3E1; _abck=7C4132A9CFF8987B662EE77CC0AFFBC4~0~YAAQJ1stFxp9lMZ7AQAAaQmzzQYmmnsf2/qU+qv6+aMbd3O+8XKeLP2V5cS2SIELLb6Li5QAGV6tYBBb8SlW6mMICOw3ZgE1+OO0bcKlqvq7pk+swjRfV/AyDY6rkX5kaYQ2SiHOyuos98tSAVmrLeJVo8o0EmqBWWs83ACksDBs5Tu4yxEn4+4ByQSoNB5vGVv69P8Z7loG0gVWfvLfPD7aCI46dKJ8gVyS9TARBw3MyuVYMuW0AFsm6tYm0rAJHa+OjUegJVRE4DWEUKqrhqtdYRjLUQWVIG2KHEz47yq/2+XQIDifjsRE2Uhcvz8zf5iGtY3Zgg8FI5MS3r4/MA98BlvKmuFoFQloaZ4WQ1xuhh4l2ZeL6qGHhoNFcao/l51sN2xHrYVL/5PtM3q11wMtwe6aQw==~-1~-1~-1; bm_sz=B20954987A4205608C0AB42E2B569FE5~YAAQJ1stFxx9lMZ7AQAAaQmzzQ2xNxku7vvD60cdVoxt+lACMKw3v/lCs5Xz1PB7FBo4Vt31TyyumZEdEUh0RWYe5n5/ROoO8r+xhDd/8KugLo4XWX7sGPeox8T4W0C6AUcFQt5RODvR3dFxEXv3LBEX8OgqvBlV8ZW3khUiwbD6l/PofCsUglZbXWqNxpRtc6ET66eskjT01KCdgNkJyusCc+aVdeKf9LQ5OTgjTK0+lDDMTGo8dccQbPsE9+coM1bJNt2gzx9QBuLsu++gvIb+LKPIRjdFXmUnxs/XSNR7~4604210~4470585; forterToken=2e1263fe8a894380bae7c1657dd2ba3d_1631243667640__UDF43_9ck; currentStepNewsletter=8; bm_sv=0F6B4EB741EEBBA4D1626AAFE688CD80~CFcDLL2Iy0L/qbk28bJ87Jaq4AexDiROnSzf+pKeRhN9zhJBRLOFLuiiO5P8fQGN2PJuGDHyweVLWa7ExrkJhWV9PgJ3jq8qtr0bxxgnWtE3yNDV3Du7oEpM0whLiMJZU657zg3c1Z1GQkFwZ1LIeQ==; ak_bmsc=956CB8D9164E3186559BECD9EF1E2E43~000000000000000000000000000000~YAAQJ1stF4V9lMZ7AQAA3R+zzQ0FR2mYrh3DhS10bwUwj9VcCLQDBiKsi9EqNMo6Si8HZcKlnHr3WWDsM7WqEwJ3m60YMojWLGftS8Kekd9zQU6z+G5xdF0CmJOASAjx+uekTIlqkXivI1cfbB+1VJOMmAxyEkD7WSbXLM7E4Ou3PBeWjuB7dfdtE6W+30sw1WMQpLjyqi94U/jUZYkpK9GkrNaaPGmOoWmrsUEwis7NGSbD9maV4+wQoth4Ui1vhe5avzzyZzzQ+LsXsjIL96GzAtnCFPJEf35MY8xIA1vk0lsyLZqC6QDc0VuUGTJ8b+awlfvapTJNl++hX5RkUvxsSZTGv3nP9WXvknJy1OxR+DlJRehE8hm1zmciY+q+s/GHioF3JoTsCk5faRrRiGYdKQEFgMmgz28owryS2Kv2G/GZblOb0bBqjUR++KPQlOQk4OJzLkBI/zRsjEiRNRBFjgZO37ymedNbWsL68xZkNBybMevu0NA=; QueueITAccepted-SDFrts345E-V3_tods=EventId=tods&QueueId=00000000-0000-0000-0000-000000000000&RedirectType=disabled&IssueTime=1631243675&Hash=20534b687e59170d4fc81b8d9efc2bf1f3a4f132a8ff9efcbadf8b83ac8cc9fa; OptanonConsent=isGpcEnabled=0&datestamp=Fri+Sep+10+2021+08:44:34+GMT+0530+(India+Standard+Time)&version=6.21.0&isIABGlobal=false&consentId=93bea2d4-c5a3-4347-9377-7257e41b189e&interactionCount=1&landingPath=NotLandingPage&groups=C0001:1,C0002:1,C0003:1,C0004:1&hosts=H11:1,H26:1,H5:1,H37:1,H6:1,H16:1,H23:1,H30:1,H33:1,H42:1,H25:1,H19:1,H40:1,H2:1,H4:1,H28:1,H20:1,H27:1,H21:1,H22:1&geolocation=;&AwaitingReconsent=false; _uetsid=408c1270115611ec9c408d6fd1a09e59; _uetvid=408c3e90115611ec9bef6f7c9ebf6a46; _ga_YMQ83SWR7E=GS1.1.1631243664.3.1.1631243674.50; stc119143=env:1631243675|20211011031435|20210910034435|1|1086561:20220910031435|uid:1631182269223.1268403493.5008063.119143.87714896.2:20220910031435|srchist:1086561:1631243675:20211011031435:20220910031435|tsa:1631243675089.1041145819.487102.1512017284480207.1:20210910034435; _sp_ses.3fc3=*; inside-eu2=314720467-077440eb89f2c8ce62d1218ba9d01ed6995cf43c55be52ceddca42ad237fbc7c-0-0; _sp_id.3fc3=b47ade9a-abd4-41c5-a042-2610ca11c4d3.1631182304.3.1631243677.1631189636.aa3a60b5-3ada-4a61-8fb3-47f69dad5651; cbar_lvt=1631243675; cbar_sess=3; cbar_sess_pv=2; RT="z=1&dm=tods.com&si=733e87d9-5e54-4d3e-9fba-ae9e83fa77ed&ss=ktdsd1hr&sl=0&tt=0&ld=wjph0&nu=d41d8cd98f00b204e9800998ecf8427e&cl=425j"; _ga=GA1.2.397623229.1631182269; _gat_UA-1567085-14=1`)

	return opts
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.tods.com"}
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
		u.Host = "www.tods.com"
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
	if p == "" || p == "/us-en" {
		return c.parseCategories(ctx, resp, yield)
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
	req, _ := http.NewRequest(http.MethodGet, "https://www.tods.com/us-en/home.html", nil)
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
		sel := dom.Find(`.navigationWrapper>li`)
		for i := range sel.Nodes {
			node := sel.Eq(i)
			cateName := strings.TrimSpace(node.Find(`a`).First().Text())
			if cateName == "" {
				continue
			}

			subSel := node.Find(`.subNavigation__list`)

			for k := range subSel.Nodes {
				subNode2 := subSel.Eq(k)
				subcat2 := strings.TrimSpace(subNode2.Find(`span`).First().Text())

				subNode2list := subNode2.Find(`.thirdNavigation__list>ul>li`)
				for j := range subNode2list.Nodes {
					subNode := subNode2list.Eq(j)
					subcat3 := strings.TrimSpace(subNode.Find(`a`).First().Text())

					if subcat3 == "" {
						continue
					}

					href := subNode.Find(`a`).First().AttrOr("href", "")
					if href == "" {
						continue
					} else if !strings.HasPrefix(href, `http`) {
						href = "https://www.tods.com" + href
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

	sel := dom.Find(`.navigation >.navigationWrapper>li`)
	for i := range sel.Nodes {
		node := sel.Eq(i)
		cateName := strings.TrimSpace(node.Find(`a span`).First().Text())
		if cateName == "" {
			continue
		}

		nctx := context.WithValue(ctx, "Category", cateName)
		//fmt.Println(`Cat Name:`, cateName)

		subSel := node.Find(`.subNavigation__list`)

		for k := range subSel.Nodes {
			subNode2 := subSel.Eq(k)
			subcat := strings.TrimSpace(subNode2.Find(`span`).First().Text())

			subNode2list := subNode2.Find(`.thirdNavigation__list>ul>li`)
			for j := range subNode2list.Nodes {
				subNode := subNode2list.Eq(j)
				subcatname := strings.TrimSpace(subNode.Find(`a`).First().Text())

				if subcatname == "" {
					continue
				}

				href := subNode.Find(`a`).First().AttrOr("href", "")
				fullurl := fmt.Sprintf("%s://%s%s", resp.Request.URL.Scheme, resp.Request.URL.Host, href)
				if href == "" {
					continue
				}

				finalsubCatName := ""
				if subcat != "" {
					finalsubCatName = subcat + " >> " + subcatname
				} else {
					finalsubCatName = subcatname
				}

				// fmt.Println(`SubCategory:`, finalsubCatName)
				// fmt.Println(`href:`, fullurl)

				u, err := url.Parse(href)
				if err != nil {
					c.logger.Error("parse url %s failed", href)
					continue
				}

				if c.categoryPathMatcher.MatchString(u.Path) {
					nnctx := context.WithValue(nctx, "SubCategory", finalsubCatName)
					req, _ := http.NewRequest(http.MethodGet, fullurl, nil)
					if err := yield(nnctx, req); err != nil {
						return err
					}
				}

			}
		}
	}
	return nil
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

	lastIndex := nextIndex(ctx)

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	sel := dom.Find(`.listingItem`)
	for i := range sel.Nodes {
		node := sel.Eq(i)
		href := node.AttrOr("href", "")
		if href == "" {
			html, _ := node.Html()
			c.logger.Debugf("%s", html)
			continue
		}

		rawurl := ""
		if href != "" {
			rawurl = fmt.Sprintf("%s://%s%s", resp.Request.URL.Scheme, resp.Request.URL.Host, href)
		}

		req, err := http.NewRequest(http.MethodGet, rawurl, nil)
		if err != nil {
			c.logger.Errorf("load http request of url %s failed, error=%s", rawurl, err)
			return err
		}

		nctx := context.WithValue(ctx, "item.index", lastIndex)
		lastIndex += 1

		if err := yield(nctx, req); err != nil {
			return err
		}
	}

	// set pagination
	nextUrl := dom.Find(`link[rel="next"]`).AttrOr(`href`, ``)
	if nextUrl == "" {
		return nil
	}

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
	resp = bytes.ReplaceAll(resp, []byte("  "), []byte(" "))
	return resp
}

type parseProductResponse struct {
	CarouselImages []struct {
		AltText string `json:"altText"`
		URL     string `json:"url"`
	} `json:"carouselImages"`
	Categories []struct {
		Code string `json:"code"`
		URL  string `json:"url"`
	} `json:"categories"`
	Code             string `json:"code"`
	Color            string `json:"color"`
	ColorSizeOptions []struct {
		Color     string `json:"color"`
		Image     string `json:"image"`
		SkuOrigin string `json:"skuOrigin"`
	} `json:"colorSizeOptions"`
	Custom              bool          `json:"custom"`
	Description         string        `json:"description"`
	EditorialComponents []interface{} `json:"editorialComponents"`
	FreeTextLabel       string        `json:"freeTextLabel"`
	FullPrice           struct {
		CurrencyIso    string  `json:"currencyIso"`
		FormattedValue string  `json:"formattedValue"`
		PriceType      string  `json:"priceType"`
		Value          float64 `json:"value"`
	} `json:"fullPrice"`
	HasSizeGuide      bool   `json:"hasSizeGuide"`
	IsDiscount        bool   `json:"isDiscount"`
	IsHenderScheme    bool   `json:"isHenderScheme"`
	IsOnlineExclusive bool   `json:"isOnlineExclusive"`
	Name              string `json:"name"`
	Picture           struct {
		URL string `json:"url"`
	} `json:"picture"`
	Price struct {
		CurrencyIso    string  `json:"currencyIso"`
		FormattedValue string  `json:"formattedValue"`
		PriceType      string  `json:"priceType"`
		Value          float64 `json:"value"`
	} `json:"price"`
	SalableStores  bool   `json:"salableStores"`
	SizeDressLabel string `json:"sizeDressLabel"`
	SizeType       string `json:"sizeType"`
	Stock          struct {
		StockLevel          int    `json:"stockLevel"`
		StockLevelAvailable int    `json:"stockLevelAvailable"`
		StockLevelStatus    string `json:"stockLevelStatus"`
	} `json:"stock"`
	Summary string `json:"summary"`
	Thumb   struct {
		AltText string `json:"altText"`
		URL     string `json:"url"`
	} `json:"thumb"`
	URL            string `json:"url"`
	VariantOptions []struct {
		Code          string `json:"code"`
		Color         string `json:"color"`
		FullPriceData struct {
			CurrencyIso    string  `json:"currencyIso"`
			FormattedValue string  `json:"formattedValue"`
			PriceType      string  `json:"priceType"`
			Value          float64 `json:"value"`
		} `json:"fullPriceData"`
		IsDiscount           bool   `json:"isDiscount"`
		MessagePreorder      string `json:"messagePreorder"`
		MessageStock         string `json:"messageStock"`
		MessageWarehouseTods string `json:"messageWarehouseTods"`
		Preorder             int    `json:"preorder"`
		PriceData            struct {
			CurrencyIso    string  `json:"currencyIso"`
			FormattedValue string  `json:"formattedValue"`
			PriceType      string  `json:"priceType"`
			Value          float64 `json:"value"`
		} `json:"priceData"`
		Size     string `json:"size"`
		SizeCode string `json:"sizeCode"`
		Stock    struct {
			StockLevel          int    `json:"stockLevel"`
			StockLevelAvailable int    `json:"stockLevelAvailable"`
			StockLevelStatus    string `json:"stockLevelStatus"`
		} `json:"stock"`
		StockLevel    int    `json:"stockLevel"`
		URL           string `json:"url"`
		WarehouseTods int    `json:"warehouseTods"`
	} `json:"variantOptions"`
}

var productsExtractReg = regexp.MustCompile(`(?Ums)<script type="application/ld\+json">\s*({.*})\s*</script>`)

type ProductPageData struct {
	Context     string   `json:"@context"`
	Type        string   `json:"@type"`
	Name        string   `json:"name"`
	Color       string   `json:"color"`
	URL         string   `json:"url"`
	Image       []string `json:"image"`
	Description string   `json:"description"`
	Mpn         string   `json:"mpn"`
	Sku         string   `json:"sku"`
	Brand       struct {
		Type string `json:"@type"`
		Name string `json:"name"`
	} `json:"brand"`
	MainEntityOfPage struct {
		Type       string `json:"@type"`
		Breadcrumb struct {
			Type            string `json:"@type"`
			ItemListElement []struct {
				Type     string `json:"@type"`
				Position int    `json:"position"`
				Item     struct {
					ID   string `json:"@id"`
					Name string `json:"name"`
				} `json:"item"`
			} `json:"itemListElement"`
		} `json:"breadcrumb"`
	} `json:"mainEntityOfPage"`
}

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

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	var viewDataBreadCrumb ProductPageData
	matched := productsExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}

	if err := json.Unmarshal(matched[1], &viewDataBreadCrumb); err != nil {
		c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		return err
	}

	s := strings.Split(strings.TrimSuffix(resp.Request.URL.Path, `/`), `/`)
	pid := s[len(s)-1]
	rootURL := "https://www.tods.com/rest/v2/tods-us/products/" + pid + "?lang=en&key=undefined"

	respBodyV, _ := c.variationRequest(ctx, rootURL, resp.Request.URL.String())

	var viewData parseProductResponse

	if err := json.Unmarshal(respBodyV, &viewData); err != nil {
		//c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		//return err
	}

	canUrl, _ := c.CanonicalUrl(doc.Find(`link[rel="canonical"]`).AttrOr("href", ""))
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}

	brand := viewDataBreadCrumb.Brand.Name
	if brand == "" {
		brand = "Tods"
	}

	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           viewData.Code,
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
		},
		BrandName: brand,
		Title:     viewData.Name,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
	}

	// desc
	description := viewData.Description + viewData.Summary
	item.Description = htmlTrimRegp.ReplaceAllString(description, ``)

	//images
	var medias []*pbMedia.Media
	for j, mid := range viewData.CarouselImages {
		imgurl := mid.URL
		if strings.HasPrefix(mid.URL, `//`) {
			imgurl = "https:" + mid.URL
		}

		medias = append(medias, pbMedia.NewImageMedia(
			strconv.Format(j),
			imgurl,
			imgurl+"?imwidth=1000",
			imgurl+"?imwidth=800",
			imgurl+"?imwidth=500",
			"", j == 0))
	}

	for i, breadcrumb := range viewDataBreadCrumb.MainEntityOfPage.Breadcrumb.ItemListElement {
		if i == 0 || i == len(viewDataBreadCrumb.MainEntityOfPage.Breadcrumb.ItemListElement)-1 {
			continue
		}

		if i == 1 {
			item.Category = breadcrumb.Item.Name
		} else if i == 2 {
			item.SubCategory = breadcrumb.Item.Name
		} else if i == 3 {
			item.SubCategory2 = breadcrumb.Item.Name
		} else if i == 4 {
			item.SubCategory3 = breadcrumb.Item.Name
		} else if i == 5 {
			item.SubCategory4 = breadcrumb.Item.Name
		}
	}

	for _, rawSku := range viewData.VariantOptions {

		currentPrice, _ := strconv.ParsePrice(rawSku.PriceData.Value)
		msrp, _ := strconv.ParsePrice(rawSku.PriceData.Value)
		// Note: product with discount not found

		if msrp == 0 {
			msrp = currentPrice
		}
		discount := 0
		if msrp > currentPrice {
			discount = (int)(((msrp - currentPrice) / msrp) * 100)
		}

		sku := pbItem.Sku{
			SourceId: fmt.Sprintf(rawSku.Code),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(currentPrice * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},
			Medias: medias,
			Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}

		if rawSku.StockLevel > 0 {
			sku.Stock.StockStatus = pbItem.Stock_InStock
			item.Stock.StockStatus = pbItem.Stock_InStock
		}

		// color
		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecColor,
			Id:    rawSku.Color,
			Name:  rawSku.Color,
			Value: rawSku.Color,
		})

		// size
		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Id:    rawSku.Size,
			Name:  rawSku.Size,
			Value: rawSku.Size,
		})

		if rawSku.Size == "" && rawSku.Color == "" {
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    "-",
				Name:  "-",
				Value: "-",
			})
		}

		item.SkuItems = append(item.SkuItems, &sku)
	}

	// yield item result
	if err = yield(ctx, &item); err != nil {
		c.logger.Errorf("yield sub request failed, error=%s", err)
		return err
	}

	return nil
}

func (c *_Crawler) variationRequest(ctx context.Context, url string, referer string) ([]byte, error) {

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	opts := c.CrawlOptions(req.URL)

	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("referer", referer)

	for _, c := range opts.MustCookies {
		req.AddCookie(c)
	}
	for k := range opts.MustHeader {
		req.Header.Set(k, opts.MustHeader.Get(k))
	}

	req.Header.Add(`cookie`, `JSESSIONID=B72373D29C88B101FC4CACEBBBC88A7F; countryIP=IN; geo-akamai=true; OptanonAlertBoxClosed=2021-09-09T10:11:08.500Z; _gcl_au=1.1.949434437.1631182269; _gid=GA1.2.1216332765.1631182269; _fbp=fb.1.1631182268928.920918955; cbar_uid=5597198485632; cbar_cart_checksum=0; _pin_unauth=dWlkPVl6RmhOV1UyWTJNdE56QmpPUzAwTWpBekxXRmxZbVl0WmpBd01qbGtZemt3WkdWaQ; ftr_ncd=6; counterNewsletterCookie=15; pageUrl=https://www.tods.com/us-en/tods-world/pre-fall21.html; newsletterClosed=true; customerData={"name":"Anonymous","userCode":"anonymous","countryCode":"US","countryName":"United+States","relativePath":"https://www.tods.com/","site":"tods-us","lingua":"en","isGuest":false,"cobrand":"","ctryLang":"us-en","currency":"USD","brochureSite":false,"visualSearch":true,"bookAnAppointment":true,"reserveInStore":true,"omnichannelStore":true,"sapCarEnabled":true,"asmEnabled":true,"newsletter":true,"counterNewsletterCookie":1,"rightToLeft":false,"algoliaSearch":true,"algoliaIndexName":"production_algoliaIndex_tods-us","algoliaApplicationID":"TM06P1CDJX","algoliaSearchAPIKey":"69755e034733f4af2cec64028cacda91","returnInStore":false}; AKA_A2=A; JSESSIONID=071C9B77E2B8EE9BD19838C02A18A3E1; _abck=7C4132A9CFF8987B662EE77CC0AFFBC4~0~YAAQJ1stFxp9lMZ7AQAAaQmzzQYmmnsf2/qU+qv6+aMbd3O+8XKeLP2V5cS2SIELLb6Li5QAGV6tYBBb8SlW6mMICOw3ZgE1+OO0bcKlqvq7pk+swjRfV/AyDY6rkX5kaYQ2SiHOyuos98tSAVmrLeJVo8o0EmqBWWs83ACksDBs5Tu4yxEn4+4ByQSoNB5vGVv69P8Z7loG0gVWfvLfPD7aCI46dKJ8gVyS9TARBw3MyuVYMuW0AFsm6tYm0rAJHa+OjUegJVRE4DWEUKqrhqtdYRjLUQWVIG2KHEz47yq/2+XQIDifjsRE2Uhcvz8zf5iGtY3Zgg8FI5MS3r4/MA98BlvKmuFoFQloaZ4WQ1xuhh4l2ZeL6qGHhoNFcao/l51sN2xHrYVL/5PtM3q11wMtwe6aQw==~-1~-1~-1; bm_sz=B20954987A4205608C0AB42E2B569FE5~YAAQJ1stFxx9lMZ7AQAAaQmzzQ2xNxku7vvD60cdVoxt+lACMKw3v/lCs5Xz1PB7FBo4Vt31TyyumZEdEUh0RWYe5n5/ROoO8r+xhDd/8KugLo4XWX7sGPeox8T4W0C6AUcFQt5RODvR3dFxEXv3LBEX8OgqvBlV8ZW3khUiwbD6l/PofCsUglZbXWqNxpRtc6ET66eskjT01KCdgNkJyusCc+aVdeKf9LQ5OTgjTK0+lDDMTGo8dccQbPsE9+coM1bJNt2gzx9QBuLsu++gvIb+LKPIRjdFXmUnxs/XSNR7~4604210~4470585; forterToken=2e1263fe8a894380bae7c1657dd2ba3d_1631243667640__UDF43_9ck; currentStepNewsletter=8; bm_sv=0F6B4EB741EEBBA4D1626AAFE688CD80~CFcDLL2Iy0L/qbk28bJ87Jaq4AexDiROnSzf+pKeRhN9zhJBRLOFLuiiO5P8fQGN2PJuGDHyweVLWa7ExrkJhWV9PgJ3jq8qtr0bxxgnWtE3yNDV3Du7oEpM0whLiMJZU657zg3c1Z1GQkFwZ1LIeQ==; ak_bmsc=956CB8D9164E3186559BECD9EF1E2E43~000000000000000000000000000000~YAAQJ1stF4V9lMZ7AQAA3R+zzQ0FR2mYrh3DhS10bwUwj9VcCLQDBiKsi9EqNMo6Si8HZcKlnHr3WWDsM7WqEwJ3m60YMojWLGftS8Kekd9zQU6z+G5xdF0CmJOASAjx+uekTIlqkXivI1cfbB+1VJOMmAxyEkD7WSbXLM7E4Ou3PBeWjuB7dfdtE6W+30sw1WMQpLjyqi94U/jUZYkpK9GkrNaaPGmOoWmrsUEwis7NGSbD9maV4+wQoth4Ui1vhe5avzzyZzzQ+LsXsjIL96GzAtnCFPJEf35MY8xIA1vk0lsyLZqC6QDc0VuUGTJ8b+awlfvapTJNl++hX5RkUvxsSZTGv3nP9WXvknJy1OxR+DlJRehE8hm1zmciY+q+s/GHioF3JoTsCk5faRrRiGYdKQEFgMmgz28owryS2Kv2G/GZblOb0bBqjUR++KPQlOQk4OJzLkBI/zRsjEiRNRBFjgZO37ymedNbWsL68xZkNBybMevu0NA=; QueueITAccepted-SDFrts345E-V3_tods=EventId=tods&QueueId=00000000-0000-0000-0000-000000000000&RedirectType=disabled&IssueTime=1631243675&Hash=20534b687e59170d4fc81b8d9efc2bf1f3a4f132a8ff9efcbadf8b83ac8cc9fa; OptanonConsent=isGpcEnabled=0&datestamp=Fri+Sep+10+2021+08:44:34+GMT+0530+(India+Standard+Time)&version=6.21.0&isIABGlobal=false&consentId=93bea2d4-c5a3-4347-9377-7257e41b189e&interactionCount=1&landingPath=NotLandingPage&groups=C0001:1,C0002:1,C0003:1,C0004:1&hosts=H11:1,H26:1,H5:1,H37:1,H6:1,H16:1,H23:1,H30:1,H33:1,H42:1,H25:1,H19:1,H40:1,H2:1,H4:1,H28:1,H20:1,H27:1,H21:1,H22:1&geolocation=;&AwaitingReconsent=false; _uetsid=408c1270115611ec9c408d6fd1a09e59; _uetvid=408c3e90115611ec9bef6f7c9ebf6a46; _ga_YMQ83SWR7E=GS1.1.1631243664.3.1.1631243674.50; stc119143=env:1631243675|20211011031435|20210910034435|1|1086561:20220910031435|uid:1631182269223.1268403493.5008063.119143.87714896.2:20220910031435|srchist:1086561:1631243675:20211011031435:20220910031435|tsa:1631243675089.1041145819.487102.1512017284480207.1:20210910034435; _sp_ses.3fc3=*; inside-eu2=314720467-077440eb89f2c8ce62d1218ba9d01ed6995cf43c55be52ceddca42ad237fbc7c-0-0; _sp_id.3fc3=b47ade9a-abd4-41c5-a042-2610ca11c4d3.1631182304.3.1631243677.1631189636.aa3a60b5-3ada-4a61-8fb3-47f69dad5651; cbar_lvt=1631243675; cbar_sess=3; cbar_sess_pv=2; RT="z=1&dm=tods.com&si=733e87d9-5e54-4d3e-9fba-ae9e83fa77ed&ss=ktdsd1hr&sl=0&tt=0&ld=wjph0&nu=d41d8cd98f00b204e9800998ecf8427e&cl=425j"; _ga=GA1.2.397623229.1631182269; _gat_UA-1567085-14=1`)

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

	return respBody, err
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://www.tods.com/us-en/",
		//"https://www.tods.com/us-en/Kate-Loafers-in-Leather/p/XXW79A0DD00NF5S607",
		//"https://www.tods.com/us-en/Men/Shoes/Loafers/c/213-Tods/",
		//"https://www.tods.com/us-en/Men/Shoes/View-all/c/219-Tods/",
		//"https://www.tods.com/us-en/Women/Shoes/Gommini/c/111-Tods/",
		//"https://www.tods.com/us-en/Gommino-Driving-Shoes-in-Suede/p/XXW00G00010RE0R411",
		//"https://www.tods.com/us-en/p/XXW00G00010RE0L012/",
		"https://www.tods.com/us-en/Gommino-Driving-Shoes-in-Leather/p/XXW00G000105J1M025",
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
