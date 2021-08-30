package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	httpClient             http.Client
	categoryPathMatcher    *regexp.Regexp
	categoryAPIPathMatcher *regexp.Regexp
	productPathMatcher     *regexp.Regexp
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
		categoryPathMatcher:    regexp.MustCompile(`^/([/A-Za-z0-9_-]+)/c([/A-Za-z0-9_-]+)$`),
		categoryAPIPathMatcher: regexp.MustCompile(`^/category-search-ajax$`),
		productPathMatcher:     regexp.MustCompile(`^/c([/A-Za-z0-9_-]+)/p/([/A-Za-z0-9_-]+)`),
		logger:                 logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "1a802ce5da394208b6feeac90dacd332"
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
	opts.MustHeader.Add(`cookie`, `mt.v=2.223379299.1629697502271; ku1-vid=8b51f997-3c01-c058-d172-6f8370ef2a59; _gcl_au=1.1.1755824579.1629697507; CoreID6=00145028208216296975066&ci=52540000|www.clarksusa.com; _ga=GA1.2.151645057.1629697507; _cs_c=0; REVLIFTER={"w":"fe93176c-afed-4170-aba7-cdeb8342c252","u":"c0fb7857-7b63-4bd6-9fa7-3d481082ecae","s":"cd714084-df74-46ac-a404-e75cd202cfdd","se":1632289508}; _fbp=fb.1.1629697508668.1725031793; _lc2_fpi=b5bb2b35bd9c--01fdrrn1dtrbfe3z9yjm9r2pg7; _mibhv=anon-1629697509213-1094887326_5146; _pin_unauth=dWlkPVpXWmhZamxoWkdFdFlUVTRaUzAwWkdGbExXSmhNalV0WldNeVpHUTBabUppWldKbQ; usi_id=vu3hkf_1629697541; akacd_holding_page_us=3807600773~rv=21~id=307a28760ca530a0d965fcb062c40cdf; cmTPSet=Y; JSESSIONID=120D602A5FE74D8878468C7C266BABDD.c2; ROUTEID=.2; b1pi=!5qRyQEAuPKvlGzng6Jezi6NCgogoyHZSn5YuQFy7q+DRUmnjMV8HHH4Y1W2rcjkhI4KTKxpaEQM3hK0=; ku1-sid=PXhVZbHYMtaXh76FyjLpl; CMAVID=none; _li_dcdm_c=.clarksusa.com; clarkscookiepolicy=accept; BVImplmain_site=19244; a1ashgd=zv7o4wo6y7000000zv7o4wo6y7000000; BVBRANDID=f57dd15d-ad65-43aa-a082-409f25452157; _cs_id=27ef90f6-2f4a-aaf6-c8bf-07bcbdb2a6b4.1629697508.4.1630150967.1630147977.1.1663861508382; _uetvid=45de888003d511ec814a7554b126c08c; 52540000|www.clarksusa.com_clogin=l=86231221630147975416&v=1&e=1630152770416; __idcontext=eyJjb29raWVJRCI6IjMyRkxVMlBTTjJMNUtXS0hWTUwzVEZJMkJSTjNRU0pPS0VJN1dVRU5RVzJRPT09PSIsImRldmljZUlEIjoiMzJGTFUyUFNNR01MQTZCR1JJSTZKSDNJRlZaSVlISUVEVVNNS1NVSVNDN0E9PT09IiwiaXYiOiJSNE1DRzZVM1daTEQzS1NWS1Q0RUtWVU41ND09PT09PSIsInYiOjF9; ADRUM=s=1630155133339&r=https://www.clarksusa.com/c/Pure2-Trim/p/26161613?0; ADRUM=s=1630155133525&r=https://www.clarksusa.com/c/Wave2-0-Step-/p/26152359?0; RT="z=1&dm=www.clarksusa.com&si=b8915d30-d65d-41ea-af8a-64d3c5a60ec3&ss=ksvo0hi2&sl=6&tt=2kpt&bcn=//684d0d37.akstat.io/&obo=1&rl=1&ld=49hh2&r=2rq3hgoq&ul=49hhy&hd=49hyw"`)
	return opts
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.clarksusa.com"}
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
		u.Host = "www.clarksusa.com"
	}
	if c.productPathMatcher.MatchString(u.Path) {
		u.RawQuery = ""
		return u.String(), nil
	}
	return rawurl, nil
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
	} else if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) || c.categoryAPIPathMatcher.MatchString(resp.Request.URL.Path) {
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
	req, _ := http.NewRequest(http.MethodGet, "https://www.clarksusa.com/", nil)
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

		sel := dom.Find(`.new-header__main-navigation-list > li`)

		for a := range sel.Nodes {
			node := sel.Eq(a)
			cateName := strings.TrimSpace(node.Find(`button`).First().Text())
			attrName := node.Find(`button`).AttrOr("data-flyout", "")
			if cateName == "" {
				continue
			}

			attrM := "#" + attrName
			subdiv := dom.Find(attrM).Find(`.new-header__flyout-top-links > li`)
			for b := range subdiv.Nodes {
				sublvl2 := subdiv.Eq(b)
				subcat2 := strings.TrimSpace(sublvl2.Find(`a`).First().Text())
				href := sublvl2.Find(`a`).AttrOr("href", "")
				if href == "" {
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

			subdiv = subdiv.Find(`.new-header__flyout-menu-list > li`)
			for b := range subdiv.Nodes {
				sublvl2 := subdiv.Eq(b)
				subcat2 := strings.TrimSpace(sublvl2.Find(`h2`).First().Text())

				selsublvl3 := sublvl2.Find(`ul > li`)
				for k := range selsublvl3.Nodes {
					sublvl3 := selsublvl3.Eq(k)
					subcat3 := strings.TrimSpace(sublvl3.Find(`a`).First().Text())

					href := sublvl2.Find(`a`).AttrOr("href", "")
					if href == "" {
						continue
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
					//lastCate.Children = append(lastCate.Children, cate)
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

	sel := dom.Find(`.new-header__main-navigation-list > li`)
	fmt.Println(len(sel.Nodes))

	for a := range sel.Nodes {
		node := sel.Eq(a)
		cateName := strings.TrimSpace(node.Find(`button`).First().Text())
		attrName := node.Find(`button`).AttrOr("data-flyout", "")
		if cateName == "" {
			continue
		}
		fmt.Println()
		fmt.Println(cateName)
		//nctx := context.WithValue(ctx, "Category", cateName)

		attrM := "#" + attrName
		subdiv := dom.Find(attrM)

		test := subdiv.Find(`.new-header__flyout-top-links > li`)
		for b := range test.Nodes {
			sublvl2 := test.Eq(b)
			sublvl2name := strings.TrimSpace(sublvl2.Find(`a`).First().Text())

			fmt.Println(sublvl2name)

		}

		test = subdiv.Find(`.new-header__flyout-menu-list > li`)
		for b := range test.Nodes {
			sublvl2 := test.Eq(b)
			sublvl2name := strings.TrimSpace(sublvl2.Find(`h2`).First().Text())

			selsublvl3 := sublvl2.Find(`ul > li`)
			for c := range selsublvl3.Nodes {
				sublvl3 := selsublvl3.Eq(c)
				sublvl3name := strings.TrimSpace(sublvl3.Find(`a`).First().Text())
				fmt.Println(sublvl2name + " > " + sublvl3name)

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

	var viewData categoryStructure

	if !c.categoryAPIPathMatcher.MatchString(resp.Request.URL.Path) {

		s := strings.Split(resp.Request.URL.Path, "/c/")

		rootUrl := "https://www.clarksusa.com/category-search-ajax?categoryCode=" + s[len(s)-1]
		req, _ := http.NewRequest(http.MethodGet, rootUrl, nil)
		opts := c.CrawlOptions(req.URL)
		req.Header.Set("accept-language", "en-GB,en-US;q=0.9,en;q=0.8")
		req.Header.Set("accept", "application/json, text/plain, */*")
		req.Header.Set("referer", resp.Request.URL.String())

		for _, c := range opts.MustCookies {
			req.AddCookie(c)
		}
		for k := range opts.MustHeader {
			req.Header.Set(k, opts.MustHeader.Get(k))
		}
		resp, err = c.httpClient.DoWithOptions(ctx, req, http.Options{
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

		respBody, err = ioutil.ReadAll(resp.Body)

		if err := json.NewDecoder(bytes.NewReader(respBody)).Decode(&viewData); err != nil {
			c.logger.Errorf("unmarshal cat detail data fialed, error=%s", err)
			//return nil, err
		}
	} else {

		if err := json.NewDecoder(bytes.NewReader(respBody)).Decode(&viewData); err != nil {
			c.logger.Errorf("unmarshal cat detail data fialed, error=%s", err)
			//return nil, err
		}
	}

	for _, items := range viewData.Products {
		if href := items.URL; href != "" {

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

	// Next page url not found
	if viewData.Pagination.NumberOfPages > 1 {
		fmt.Println(`implement pagination`)
		c.logger.Errorf("implement pagination=%s")
		return nil
	}
	return nil
}

type categoryStructure struct {
	Pagination struct {
		PageSize             int    `json:"pageSize"`
		Sort                 string `json:"sort"`
		CurrentPage          int    `json:"currentPage"`
		TotalNumberOfResults int    `json:"totalNumberOfResults"`
		NumberOfPages        int    `json:"numberOfPages"`
	} `json:"pagination"`
	Products []struct {
		URL string `json:"url"`
	} `json:"products"`
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

var productsReviewExtractReg = regexp.MustCompile(`(?Ums)var attributesProductView=\s*({.*}),attributesProductViewString`)
var imageRegStart = regexp.MustCompile(`\(([^;]+),`)

// used to trim html labels in description
var htmlTrimRegp = regexp.MustCompile(`</?[^>]+>`)

type parseProductResponse struct {
	NosOfReviews  string `json:"nosOfReviews"`
	AvgStarRating string `json:"avgStarRating"`
}

type SkuDetail2 struct {
	//Fit4W struct {
	URL                  string `json:"url"`
	OutOfStock           bool   `json:"outOfStock"`
	AvailableStockAmount int    `json:"availableStockAmount"`
	SkuPrice             string `json:"skuPrice"`
	SkuWasPrice          string `json:"skuWasPrice"`
	SkuPriceDiscount     int    `json:"skuPriceDiscount"`
	SkuPriceTitle        string `json:"skuPriceTitle"`
	SkuWasPriceTitle     string `json:"skuWasPriceTitle"`
	NotifyMeEnabled      bool   `json:"notifyMeEnabled"`
	ProductCode          string `json:"productCode"`
	//} `json:"FIT_4W"`

}

type SkuDetail1 struct {
	SizeDetails map[string]*SkuDetail2
}

type parseVariationProductResponse struct {
	SkuDetails map[string]*SkuDetail1
}

type parseImageResponse struct {
	Set struct {
		Pv   string `json:"pv"`
		Type string `json:"type"`
		N    string `json:"n"`
		Item []struct {
			I struct {
				N string `json:"n"`
			} `json:"i"`
			S struct {
				N string `json:"n"`
			} `json:"s"`
			Dx string `json:"dx"`
			Dy string `json:"dy"`
			Iv string `json:"iv"`
		} `json:"item"`
	} `json:"set"`
}

func DecodeResponseVarWidth(respBody []byte) (*parseVariationProductResponse, error) {
	viewData := parseVariationProductResponse{SkuDetails: map[string]*SkuDetail1{}}

	ret := map[string]json.RawMessage{}
	if err := json.Unmarshal((respBody), &ret); err != nil {
		return nil, err
	}

	for key, msg := range ret {
		//rawData, _ := msg.MarshalJSON()
		if regexp.MustCompile(`SIZE_[0-9]+`).MatchString(key) {
			var (
				rawData, _ = msg.MarshalJSON()
				article    SkuDetail1
			)
			if err := json.Unmarshal(rawData, &article); err != nil {
				continue
			}

			viewData.SkuDetails[key] = &article

			ret2 := map[string]json.RawMessage{}
			if err := json.Unmarshal((rawData), &ret2); err != nil {
				return nil, err
			}

			for key2, msg2 := range ret2 {
				if regexp.MustCompile(`FIT_[0-9]+W`).MatchString(key2) {
					var (
						rawData, _ = msg2.MarshalJSON()
						article2   SkuDetail2
					)
					if err := json.Unmarshal(rawData, &article2); err != nil {
						continue
					}
					if viewData.SkuDetails[key].SizeDetails == nil {
						viewData.SkuDetails[key].SizeDetails = (map[string]*SkuDetail2{})
					}

					viewData.SkuDetails[key].SizeDetails[key2] = &article2
				}
			}
		}
	}
	return &viewData, nil
}

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

	s := strings.Split(resp.Request.URL.Path, `/`)
	pid := s[len(s)-1]
	variantURL := "https://www.clarksusa.com/p/" + pid + "/getProductSizeMatrix"

	respBodyV := c.variationRequest(ctx, variantURL, resp.Request.URL.String())
	viewDataSize, _ := DecodeResponseVarWidth(respBodyV)

	var viewData parseProductResponse
	matched := productsReviewExtractReg.FindSubmatch([]byte(respBody))
	if len(matched) > 1 {
		if err := json.Unmarshal(matched[1], &viewData); err != nil {
			//c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
			//return err
		}
	}

	reviewCount, _ := strconv.ParseInt(viewData.NosOfReviews)
	rating, _ := strconv.ParseFloat(viewData.AvgStarRating)

	canUrl, _ := c.CanonicalUrl(doc.Find(`link[rel="canonical"]`).AttrOr("href", ""))
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}

	brand := doc.Find(`#ao-logo-img`).Text()
	if brand == "" {
		brand = "Clarks"
	}

	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           pid,
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
			//GroupId:      doc.Find(`meta[property="product:age_group"]`).AttrOr(`content`, ``),
		},
		BrandName: brand,
		Title:     doc.Find(`.product-name-panel__h1`).Text(),
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
	description := htmlTrimRegp.ReplaceAllString(doc.Find(`.product-description__text`).Text()+" "+doc.Find(`.product-page-tabs__specifications`).Text(), " ")
	item.Description = string(TrimSpaceNewlineInString([]byte(description)))

	msrp, _ := strconv.ParsePrice(doc.Find(`.js-prev-price`).AttrOr("content", ""))
	currentPrice, _ := strconv.ParsePrice(doc.Find(`.js-current-price`).AttrOr("content", ""))

	if msrp == 0 {
		msrp = currentPrice
	}
	discount := 0
	if msrp > currentPrice {
		discount = (int)(((msrp - currentPrice) / msrp) * 100)
	}

	//-----------------------------------

	imgrequest := "https://clarks.scene7.com/is/image/Pangaea2Build/" + pid + "_SET?req=set,json&s7jsonResponse=axiosJsonpCallback1"

	req, err := http.NewRequest(http.MethodGet, imgrequest, nil)
	req.Header.Set("Referer", resp.Request.URL.String())

	imgreq, err := c.httpClient.Do(ctx, req)
	if err != nil {
		panic(err)
	}
	defer imgreq.Body.Close()

	respBodyImg, err := io.ReadAll(imgreq.Body)
	if err != nil {
		c.logger.Error(err)
		return err
	}

	matched = imageRegStart.FindSubmatch(respBodyImg)
	if len(matched) <= 1 {
		c.logger.Debugf("data %s", respBodyImg)
		return fmt.Errorf("extract produt json from page %s content failed", resp.Request.URL)
	}
	var q parseImageResponse
	if err = json.Unmarshal(matched[1], &q); err != nil {
		c.logger.Debugf("parse %s failed, error=%s", matched[1], err)
		return err
	}

	for key, img := range q.Set.Item {

		if strings.Contains(img.I.N, "Image_Not") || strings.Contains(img.I.N, "_video") {
			continue
		}
		imgURLDefault := "https://clarks.scene7.com/is/image/" + img.I.N

		item.Medias = append(item.Medias, pbMedia.NewImageMedia(
			strconv.Format(img.Iv),
			imgURLDefault,
			imgURLDefault+"?qlt=70&fmt=jpeg&wid=1000&hei=1200",
			imgURLDefault+"?qlt=70&fmt=jpeg&wid=690&hei=810",
			imgURLDefault+"?qlt=70&fmt=jpeg&wid=590&hei=700",
			"",
			key == 0,
		))
	}

	// itemListElement
	sel := doc.Find(`.breadcrumb>li`)
	c.logger.Debugf("nodes %d", len(sel.Nodes))
	for i := range sel.Nodes {
		node := sel.Eq(i)
		breadcrumb := strings.TrimSpace(node.Find(`a`).First().Text())

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

	// Color
	cid := ""
	colorName := ""
	var colorSelected *pbItem.SkuSpecOption
	sel = doc.Find(`.color-swatch-selector>ul>li`)
	for i := range sel.Nodes {
		node := sel.Eq(i)

		if strings.Contains(node.AttrOr(`class`, ``), `selected`) {
			cid = node.AttrOr(`colour-name`, ``)
			icon := strings.ReplaceAll(strings.ReplaceAll(node.Find(`img`).AttrOr(`src`, ""), "background-image: url(", ""), ")", "")
			colorName = node.AttrOr(`colour-name`, "")
			colorSelected = &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    cid,
				Name:  colorName,
				Value: colorName,
				Icon:  icon,
			}
		}
	}

	sel = doc.Find(`.box-selectors__wrapper > label`)
	for i := range sel.Nodes {
		node := sel.Eq(i)

		sizeValue := (node.Find(`input`).First().AttrOr("data-code-value", ""))
		sid := (node.Find(`input`).First().AttrOr("data-prompt-value", ""))
		if sid == "" {
			continue
		}

		for j, rawSku := range viewDataSize.SkuDetails[sizeValue].SizeDetails {

			sku := pbItem.Sku{
				SourceId: fmt.Sprintf("%s-%s-%s", cid, sid, rawSku.ProductCode),
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(currentPrice),
					Msrp:     int32(msrp),
					Discount: int32(discount),
				},
				Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}

			if rawSku.AvailableStockAmount > 0 {
				sku.Stock.StockStatus = pbItem.Stock_InStock
				item.Stock.StockStatus = pbItem.Stock_InStock
			}

			if colorSelected != nil {
				sku.Specs = append(sku.Specs, colorSelected)
			}

			// size
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    rawSku.ProductCode,
				Name:  sid + " " + j,
				Value: sid + " " + j,
			})

			item.SkuItems = append(item.SkuItems, &sku)
		}
	}

	// yield item result
	if err = yield(ctx, &item); err != nil {
		c.logger.Errorf("yield sub request failed, error=%s", err)
		return err
	}

	// // found other color
	sel = doc.Find(`.color-swatch-selector>ul>li`)
	for i := range sel.Nodes {
		node := sel.Eq(i)
		color := strings.TrimSpace(node.AttrOr("colour-name", ""))
		if !strings.Contains(node.AttrOr(`class`, ``), `selected`) {
			c.logger.Debugf("found color %s %t", color, color == colorName)

			u := node.Find(`a`).AttrOr("href", "")
			if u == "" {
				continue
			}
			req, err := http.NewRequest(http.MethodGet, u, nil)
			if err != nil {
				c.logger.Error(err)
				continue
			}
			if err := yield(ctx, req); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *_Crawler) variationRequest(ctx context.Context, url string, referer string) []byte {

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
		//return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)

	return respBody
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://www.clarksusa.com/",
		//"https://www.clarksusa.com/Womens-Best-Sellers/c/us182",
		"https://www.clarksusa.com/c/Wave2-0-Step-/p/26152404",
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
