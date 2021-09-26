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
	httpClient http.Client

	categoryPathMatcher *regexp.Regexp
	productPathMatcher  *regexp.Regexp
	logger              glog.Log
}

func (*_Crawler) New(_ *cli.Context, client http.Client, logger glog.Log) (crawler.Crawler, error) {
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
	options.EnableSessionInit = true
	options.Reliability = pbProxy.ProxyReliability_ReliabilityMedium

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

	if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yieldWrap)
	} else if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yieldWrap)
	}
	return crawler.ErrUnsupportedPath
}

func (c *_Crawler) GetCategories(ctx context.Context) ([]*pbItem.Category, error) {
	req, _ := http.NewRequest(http.MethodGet, "https://www.thereformation.com/", nil)
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
	type categorylists struct {
		URL   string
		Label string
	}

	var viewData []categorylists

	sel := dom.Find(`.primary-nav`).Find(`li`)

	for j := range sel.Nodes {
		subnode := sel.Eq(j)
		if subnode.AttrOr("data-primary-nav-content", "") == "" {
			continue
		}
		viewData = append(viewData,
			categorylists{URL: "https://www.thereformation.com/menus/" + subnode.AttrOr("data-primary-nav-content", ""),
				Label: subnode.Find(`a`).Text()})
	}

	var (
		cates   []*pbItem.Category
		cateMap = map[string]*pbItem.Category{}
	)
	if err := func(yield func(names []string, url string) error) error {

		for _, catUrl := range viewData {

			req, err := http.NewRequest(http.MethodGet, catUrl.URL, nil)
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

			dom, err := goquery.NewDocumentFromReader(bytes.NewReader(catBody))
			if err != nil {
				c.logger.Error(err)
				return err
			}

			sel := dom.Find(`.taxonomy-content-block`)
			if len(sel.Nodes) > 0 {
				sublnk := sel.Find(`.taxonomy-content-block__container>ul>li`)
				for j := range sublnk.Nodes {
					subnode := sublnk.Eq(j)
					subCateName := subnode.Find(`a`).Text()

					href := subnode.Find(`a`).AttrOr("href", "")
					if href == "" {
						continue
					}

					canonicalhref, err := c.CanonicalUrl(href)
					if err != nil {
						continue
					}

					u, err := url.Parse(canonicalhref)
					if err != nil {
						c.logger.Error("parse url %s failed", href)
						continue
					}

					if c.categoryPathMatcher.MatchString(u.Path) {
						if err := yield([]string{catUrl.Label, subCateName}, canonicalhref); err != nil {
							return err
						}
					}
				}
			}

			selother := dom.Find(`.collection-taxonomy-content-block`)
			if len(selother.Nodes) > 0 {
				sublnk := selother.Find(`.collection-taxonomy-content-block__container>ul>li`)
				for j := range sublnk.Nodes {
					subnode := sublnk.Eq(j)
					subCateName := subnode.Find(`a`).Text()

					href := subnode.Find(`a`).AttrOr("href", "")
					if href == "" {
						continue
					}

					canonicalhref, err := c.CanonicalUrl(href)
					if err != nil {
						continue
					}

					u, err := url.Parse(canonicalhref)
					if err != nil {
						c.logger.Error("parse url %s failed", href)
						continue
					}

					if c.categoryPathMatcher.MatchString(u.Path) {
						if err := yield([]string{catUrl.Label, subCateName}, canonicalhref); err != nil {
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
		if href, _ := node.Find(`a`).First().Attr("href"); href != "" {

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

func TrimSpaceNewlineInString(s string) string {
	re := regexp.MustCompile(`\n`)
	resp := re.ReplaceAllString(s, " ")
	resp = strings.ReplaceAll(resp, "\\n", " ")
	resp = strings.ReplaceAll(resp, "\r", " ")
	resp = strings.ReplaceAll(resp, "\t", " ")

	re = regexp.MustCompile(`\s+`)
	resp = re.ReplaceAllString(resp, " ")
	return strings.TrimSpace(resp)
}

// used to trim html labels in description
var (
	htmlTrimRegp = regexp.MustCompile(`(</?[^>]+>)|(&#[A-Za-z0-9]+;)`)
	imgReg       = regexp.MustCompile(`_\d+(_\d+_)\d+_\d+(:\d+)?/v1`)
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

	canUrl, _ := c.CanonicalUrl(doc.Find(`link[rel="canonical"]`).AttrOr("href", ""))
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}

	desc := ""

	sel := doc.Find(`.data-accordion__content`)
	for i := range sel.Nodes {
		if i == 0 {
			continue
		}
		desc = desc + htmlTrimRegp.ReplaceAllString(sel.Eq(i).Text(), "")
	}

	title := doc.Find(`.pdp__name>h1[itemprop="name"]`).Text()
	if title == "" {
		title = doc.Find(`meta[property='og:title']`).AttrOr("content", "")
	}

	pid := doc.Find(`input[name="product_id"]`).AttrOr("value", "")
	if pid == "" {
		pid = doc.Find(`meta[itemprop="mpn"]`).AttrOr("content", "")
	}

	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           pid,
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
		},
		Title:       TrimSpaceNewlineInString(title),
		Description: TrimSpaceNewlineInString(desc),
		BrandName:   "Reformation",
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
	}

	sel = doc.Find(`.pdp-breadcrumbs__link`)
	for i := range sel.Nodes {
		node := sel.Eq(i)
		breadcrumb := strings.TrimSpace(node.Text())

		if i == len(sel.Nodes)-1 {
			continue
		}

		if i == 1 {
			item.Category = breadcrumb
		} else if i == 2 {
			item.SubCategory = breadcrumb
		} else if i == 3 {
			item.SubCategory2 = breadcrumb
		} else if i == 4 {
			item.SubCategory4 = breadcrumb
		}
	}

	dom := doc
	selColor := doc.Find(`.pdp-color-options__color-group>li`)
	for i := range selColor.Nodes {
		nodeColor := selColor.Eq(i)

		color := TrimSpaceNewlineInString(nodeColor.Text())

		if subClass := nodeColor.AttrOr(`class`, ``); strings.Contains(subClass, `--selected`) {
			dom = doc
		} else {
			// new request
			prodUrl := "https://www.thereformation.com" + nodeColor.Find(`a`).AttrOr(`href`, ``)

			respBodyJs, err := c.variationRequest(ctx, prodUrl, resp.Request.URL.String())
			if err != nil {
				c.logger.Error(err)
				return err
			}

			dom, err = goquery.NewDocumentFromReader(bytes.NewReader(respBodyJs))
			if err != nil {
				c.logger.Error(err)
				return err
			}
		}

		var medias []*media.Media

		sel = dom.Find(`.pdp__mobile-images`).Find(`img`)
		for i := range sel.Nodes {
			node := sel.Eq(i)
			image_url := strings.TrimSpace(node.AttrOr("data-src", ""))

			if image_url == "" || strings.Contains(image_url, `video`) {
				continue
			}
			img := media.Media_Image{
				Id:          strconv.Format(len(medias)),
				OriginalUrl: image_url,
				LargeUrl:    image_url,
				MediumUrl:   image_url,
				SmallUrl:    image_url,
			}
			imgSubMatch := imgReg.FindSubmatch([]byte(image_url))
			if len(imgSubMatch) >= 2 {
				img.MediumUrl = strings.ReplaceAll(image_url, string(imgSubMatch[1]), "_600_")
				img.SmallUrl = strings.ReplaceAll(image_url, string(imgSubMatch[1]), "_500_")
			}

			itemImg, _ := anypb.New(&img)
			medias = append(medias, &media.Media{
				Detail:    itemImg,
				IsDefault: len(medias) == 0,
			})
		}

		sel1 := dom.Find(`.pdp-thumbs__primary-container`).Find(`video`)
		for j := range sel1.Nodes {
			videonode := sel1.Eq(j)
			videcover := videonode.AttrOr(`poster`, ``)

			videos := videonode.Find(`source`)
			for v := range videos.Nodes {
				node := videos.Eq(v)
				videourl := node.AttrOr(`src`, ``)

				medias = append(medias, pbMedia.NewVideoMedia(
					strconv.Format(len(medias)),
					"",
					videourl,
					0, 0, 0, videcover, "",
					len(medias) == 0))
			}
		}
		item.Medias = append(item.Medias, medias...)

		current, _ := strconv.ParseFloat(dom.Find(`.product-prices`).Find(`.product-prices__price`).AttrOr("data-product-sell-price", ""))
		msrp, _ := strconv.ParseFloat(dom.Find(`.product-prices`).Find(`.product-prices__price`).AttrOr("data-product-sell-price", ""))
		discount := 0.0
		if msrp > current {
			discount = ((msrp - current) / msrp) * 100
		}

		sel = dom.Find(`.pdp-size-options__size-button`)

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
				Medias: medias,
				Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}

			if strings.Contains(node.AttrOr("data-pdp-size-button", ""), `"purchasability":"available",`) {
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

			sizeValue := strings.TrimSpace(sel.Eq(i).Text())
			if sizeValue != "" {
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecSize,
					Id:    sizeValue,
					Name:  sizeValue,
					Value: sizeValue,
				})
			}
			item.SkuItems = append(item.SkuItems, &sku)
		}
	}

	if err = yield(ctx, &item); err != nil {
		return err
	}
	return nil
}

func (c *_Crawler) variationRequest(ctx context.Context, url string, referer string) ([]byte, error) {

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	opts := c.CrawlOptions(req.URL)
	req.Header.Set("accept-language", "en-GB,en-US;q=0.9,en;q=0.8")
	req.Header.Set("accept", "*/*")
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
		//"https://www.thereformation.com/",
		//"https://www.thereformation.com/categories/bodysuits",
		//"https://www.thereformation.com/categories/new",
		//"https://www.thereformation.com/products/carina-lace-up-mid-heel-sandal?color=Strawberry&via=Z2lkOi8vcmVmb3JtYXRpb24td2VibGluYy9Xb3JrYXJlYTo6Q2F0YWxvZzo6Q2F0ZWdvcnkvNjA4NzQwODA0M2MyZGMyMGM3NWRkMjVh",
		//"https://www.thereformation.com/products/assunta-strappy-block-heel-mule?color=Almond&via=Z2lkOi8vcmVmb3JtYXRpb24td2VibGluYy9Xb3JrYXJlYTo6Q2F0YWxvZzo6Q2F0ZWdvcnkvNWE2YWRmZDJmOTJlYTExNmNmMDRlOWM0",
		//"https://www.thereformation.com/products/kaleigh-top?color=Black&via=Z2lkOi8vcmVmb3JtYXRpb24td2VibGluYy9Xb3JrYXJlYTo6Q2F0YWxvZzo6Q2F0ZWdvcnkvNjA1Mjg1NjM4OTg2YTIwMTUyYjI1OTEx",
		//"https://www.thereformation.com/products/cynthia-shadow-checked-high-rise-straight-long-jeans?color=Seine+Checkerboard&via=Z2lkOi8vcmVmb3JtYXRpb24td2VibGluYy9Xb3JrYXJlYTo6Q2F0YWxvZzo6Q2F0ZWdvcnkvNWE2YWRmZDNmOTJlYTExNmNmMDRlOWQz",
		//"https://www.thereformation.com/products/cynthia-high-relaxed-jean?via=Z2lkOi8vcmVmb3JtYXRpb24td2VibGluYy9Xb3JrYXJlYTo6Q2F0YWxvZzo6Q2F0ZWdvcnkvNWE2YWRmZDNmOTJlYTExNmNmMDRlOWQz&color=Tahoe",
		"https://www.thereformation.com/products/samuel-dress",
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
