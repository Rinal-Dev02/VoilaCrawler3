package main

import (
	"bytes"
	"context"
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
	logger              glog.Log
}

func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient:          client,
		categoryPathMatcher: regexp.MustCompile(`^/categories([/A-Za-z0-9_-]+)$`),
		productPathMatcher:  regexp.MustCompile(`^/products/([/A-Za-z0-9_-]+)$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "4ba4a5266f264cf2802fa822903c7664"
}

// Version
func (c *_Crawler) Version() int32 {
	return 1
}

// CrawlOptions
func (c *_Crawler) CrawlOptions(u *url.URL) *crawler.CrawlOptions {
	options := crawler.NewCrawlOptions()
	options.EnableHeadless = false
	//options.LoginRequired = false
	options.EnableSessionInit = true
	options.Reliability = pbProxy.ProxyReliability_ReliabilityMedium
	options.MustCookies = append(options.MustCookies,
		&http.Cookie{Name: "ckm-ctx-sf", Value: `%2F`, Path: "/"},
	)
	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.thereformation.com"}
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
		u.Host = "www.thereformation.com"
	}
	if c.productPathMatcher.MatchString(u.Path) {
		u.RawQuery = ""
		return u.String(), nil
	}
	return rawurl, nil
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

	p := strings.TrimSuffix(resp.Request.URL.Path, "/")
	if p == "" {
		return c.parseCategories(ctx, resp, yield)
	}
	if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yieldWrap)
	} else if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yieldWrap)
	}
	return fmt.Errorf("unsupported url %s", resp.Request.URL.String())
}

func (c *_Crawler) parseCategories(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	ioutil.WriteFile("C:\\NewGIT_SVN\\Project_VoilaCrawler\\VoilaCrawler\\Output.html", respBody, 0644)

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}
	s := []string{}
	sel := dom.Find(`.primary-nav.primary-nav--small-screen>li`)
	fmt.Println(len(sel.Nodes))
	for j := range sel.Nodes {
		subnode := sel.Eq(j)
		s = append(s, "https://www.thereformation.com/menus/"+subnode.AttrOr("data-primary-nav-content", ""))
	}

	for i, catUrl := range s {

		req, err := http.NewRequest(http.MethodGet, catUrl, nil)
		req.Header.Add("accept", "*/*")
		req.Header.Add("referer", "https://www.thereformation.com/")
		req.Header.Add("accept-language", "en-US,en;q=0.9")
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

		ioutil.WriteFile("C:\\NewGIT_SVN\\Project_VoilaCrawler\\VoilaCrawler\\Output"+strconv.Format(i)+".html", catBody, 0644)

		dom, err := goquery.NewDocumentFromReader(bytes.NewReader(catBody))
		if err != nil {
			c.logger.Error(err)
			return err
		}

		cateName := "Category"

		//nnctx := context.WithValue(ctx, "Category", cateName)
		fmt.Println(`cateName `, cateName)

		sel := dom.Find(`.taxonomy-content-block__menu-link`)
		for j := range sel.Nodes {
			subnode := sel.Eq(j)

			href := subnode.AttrOr("href", "")
			if href == "" {
				continue
			}

			_, err := url.Parse(href)
			if err != nil {
				c.logger.Error("parse url %s failed", href)
				continue
			}

			subCateName := subnode.Text()
			fmt.Println(subCateName)
			// nnnctx := context.WithValue(nnctx, "SubCategory", subCateName)
			// req, _ := http.NewRequest(http.MethodGet, href, nil)
			// if err := yield(nnnctx, req); err != nil {
			// 	return err
			// }
		}
	}

	return nil
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

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	lastIndex := nextIndex(ctx)
	sel := doc.Find(`.product-summary__name`)
	for i := range sel.Nodes {
		node := sel.Eq(i)
		if href, _ := node.Find(`a`).Attr("href"); href != "" {

			//c.logger.Debugf(href)

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

	if bytes.Contains(respBody, []byte(`&quot;lastPage&quot;:true`)) || bytes.Contains(respBody, []byte(`"lastPage":true,`)) {
		return nil
	}

	// get current page number
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}

	// set pagination
	u := *resp.Request.URL
	vals := u.Query()
	vals.Set("page", strconv.Format(page+1))
	u.RawQuery = vals.Encode()

	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	// update the index of last page
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

	return resp
}

// used to trim html labels in description
var htmlTrimRegp = regexp.MustCompile(`</?[^>]+>`)

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

	canUrl, _ := c.CanonicalUrl(doc.Find(`link[rel="canonical"]`).AttrOr("href", ""))
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}

	color := strings.ReplaceAll(strings.TrimSpace(doc.Find(`.pdp-color-options__label`).Text()), "Color: ", "")
	desc := ""

	sel := doc.Find(`.data-accordion__content`)
	for i := range sel.Nodes {
		if i == 0 {
			continue
		}
		desc = desc + htmlTrimRegp.ReplaceAllString(sel.Eq(i).Text(), "")
	}

	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           doc.Find(`meta[itemprop="sku"]`).AttrOr("content", ""),
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
		},
		Title:       doc.Find(`.pdp__name`).Text(),
		Description: desc,
		BrandName:   "Reformation",
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
	}

	sel = doc.Find(`.pdp-breadcrumbs__link`)
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

	var medias []*media.Media

	sel = doc.Find(`.pdp__mobile-images`).Find(`img`)

	for i := range sel.Nodes {
		node := sel.Eq(i)
		image_url := strings.TrimSpace(node.AttrOr("data-src", ""))
		if image_url == "" {
			continue
		}
		itemImg, _ := anypb.New(&media.Media_Image{
			OriginalUrl: image_url,
			LargeUrl:    image_url,
			MediumUrl:   image_url,
			SmallUrl:    image_url,
		})
		medias = append(medias, &media.Media{
			Detail:    itemImg,
			IsDefault: i == 0,
		})

	}
	item.Medias = medias

	current, _ := strconv.ParseFloat(doc.Find(`.product-prices`).Find(`.product-prices__price`).AttrOr("data-product-sell-price", ""))
	msrp, _ := strconv.ParseFloat(doc.Find(`.product-prices`).Find(`.product-prices__price`).AttrOr("data-product-sell-price", ""))
	discount := 0.0
	// if msrp > current {
	// 	discount = ((msrp - current) / msrp) * 100
	// }

	sel = doc.Find(`.pdp-size-options__size-button`)

	for i := range sel.Nodes {
		node := sel.Eq(i).Find(`input`)
		sku := pbItem.Sku{
			SourceId: strings.TrimSpace(node.AttrOr("value", "")),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(current * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},

			Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}
		// if i == 0 {
		// 	sku.Medias = medias
		// }

		if strings.Contains(node.AttrOr("data-pdp-size-button", ""), `"purchasability":"available",`) {
			sku.Stock.StockStatus = pbItem.Stock_InStock
		}

		if color != "" {
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    doc.Find(`meta[itemprop="sku"]`).AttrOr("content", ""),
				Name:  color,
				Value: color,
			})
		}

		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Id:    node.AttrOr("value", ""),
			Name:  strings.TrimSpace(sel.Eq(i).Text()),
			Value: strings.TrimSpace(sel.Eq(i).Text()),
		})
		item.SkuItems = append(item.SkuItems, &sku)
	}

	if err = yield(ctx, &item); err != nil {
		return err
	}
	return nil
}

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		"https://www.thereformation.com/",
		//"https://www.thereformation.com/categories/heeled-sandals",
		//"https://www.thereformation.com/products/carina-lace-up-mid-heel-sandal?color=Strawberry&via=Z2lkOi8vcmVmb3JtYXRpb24td2VibGluYy9Xb3JrYXJlYTo6Q2F0YWxvZzo6Q2F0ZWdvcnkvNjA4NzQwODA0M2MyZGMyMGM3NWRkMjVh",
		//"https://www.thereformation.com/products/assunta-strappy-block-heel-mule?color=Almond&via=Z2lkOi8vcmVmb3JtYXRpb24td2VibGluYy9Xb3JrYXJlYTo6Q2F0YWxvZzo6Q2F0ZWdvcnkvNWE2YWRmZDJmOTJlYTExNmNmMDRlOWM0",
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
	os.Setenv("VOILA_PROXY_URL", "http://52.207.171.114:30216")
	cli.NewApp(New).Run(os.Args)
}
