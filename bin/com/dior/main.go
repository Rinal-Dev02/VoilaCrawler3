package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/voiladev/VoilaCrawler/pkg/cli"
	"github.com/voiladev/VoilaCrawler/pkg/context"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	media "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/media"
	pbMedia "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
)

// _Crawler defined the crawler struct/class for which is not necessory to be exportable
type _Crawler struct {
	// httpClient is the object of an http client
	httpClient                      http.Client
	categoryPathMatcher             *regexp.Regexp
	categorySearchPathMatcher       *regexp.Regexp
	categorySearchPathPagingMatcher *regexp.Regexp
	productPathMatcher              *regexp.Regexp
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
		categoryPathMatcher:             regexp.MustCompile(`^/en_us([/a-z0-9A-Z-]+)$`),
		categorySearchPathMatcher:       regexp.MustCompile(`^/en_us/products/search(.*)$`),
		categorySearchPathPagingMatcher: regexp.MustCompile(`^/1/indexes/\*/queries(.*)$`),
		// this regular used to match product page url path
		productPathMatcher: regexp.MustCompile(`^/en_us/products/(.*)$`),
		logger:             logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "6d47065c786e937ffc4b03e5c7f3ecc1"
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
		EnableHeadless: true,
		// use js api to init session for the first request of the crawl
		EnableSessionInit: false,
	}
	opts.MustCookies = append(opts.MustCookies,
		&http.Cookie{Name: "x-ak-country-code", Value: "US", Path: "/"},
		&http.Cookie{Name: "lang", Value: "v=2&lang=en-us", Path: "/"},
	)
	return opts
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.dior.com"}
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
		u.Host = "www.dior.com"
	}

	if c.productPathMatcher.MatchString(u.Path) {
		u.RawQuery = ""
		return u.String(), nil
	}
	return u.String(), nil
}

func (c *_Crawler) GetCategories(ctx context.Context) ([]*pbItem.Category, error) {
	req, _ := http.NewRequest(http.MethodGet, "https://www.dior.com/en_us", nil)
	opts := c.CrawlOptions(req.URL)
	resp, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
		EnableProxy:       true,
		EnableHeadless:    opts.EnableHeadless,
		EnableSessionInit: opts.EnableSessionInit,
		Reliability:       opts.Reliability,
		DisableCookieJar:  opts.DisableCookieJar,
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

	var cates []*pbItem.Category
	sel := dom.Find(`.navigation-desktop__tab-wrapper`)
	for i := range sel.Nodes {
		node := sel.Eq(i)
		cateName := strings.TrimSpace(node.Find(`a`).First().Text())
		if cateName == "" {
			continue
		}
		cate := pbItem.Category{Name: cateName}
		cates = append(cates, &cate)

		sectionSel := node.Find(`.navigation-tab-content-column .navigation-desktop-section`)
		for sectionIndex := range sectionSel.Nodes {
			sectionNode := sectionSel.Eq(sectionIndex)

			subCateNode := sectionNode.Find(`a.navigation-desktop-section-title`)
			href := subCateNode.AttrOr("href", "")
			href, _ = c.CanonicalUrl(href)

			subCate := pbItem.Category{
				Name: subCateNode.Text(),
				Url:  href,
			}
			cate.Children = append(cate.Children, &subCate)

			subSel := sectionNode.Find(`.navigation-desktop-section-link a`)
			for j := range subSel.Nodes {
				subCate2Node := subSel.Eq(j)
				href := subCate2Node.AttrOr("href", "")
				if href == "" {
					continue
				}

				href, err := c.CanonicalUrl(href)
				if err != nil {
					c.logger.Error("parse url %s failed", href)
					continue
				}
				subCateName := strings.TrimSpace(subCate2Node.Text())
				subCate2 := pbItem.Category{Name: subCateName, Url: href}
				subCate.Children = append(subCate.Children, &subCate2)
			}
		}
	}
	return cates, nil
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

	if p == "/en_us" || p == "" {
		return crawler.ErrUnsupportedPath
	}
	if c.categorySearchPathMatcher.MatchString(p) || c.categorySearchPathPagingMatcher.MatchString(p) {

		return c.parseSearchKeywordProducts(ctx, resp, yield)
	} else if c.productPathMatcher.MatchString(p) {

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

// below are the golang json data struct of raw website.
// if you get the raw json data of the website,
// then you can use https://mholt.github.io/json-to-go/ to convert it to a golang struct

type CategoryStructure struct {
	Props struct {
		InitialReduxState struct {
			ALGOLIA struct {
				Searches map[string]struct {
					//MenRtwSuitsJackets struct {
					Result struct {
						NbHits      int `json:"nbHits"`
						Page        int `json:"page"`
						NbPages     int `json:"nbPages"`
						HitsPerPage int `json:"hitsPerPage"`
						Hits        []struct {
							Type       string `json:"type"`
							Columns    int    `json:"columns,omitempty"`
							ObjectID   string `json:"objectID"`
							Attributes struct {
								ProductLink struct {
									URI string `json:"uri"`
								} `json:"productLink"`
							} `json:"attributes,omitempty"`
						} `json:"hits"`
					} `json:"result"`
					//} `json:"men-rtw-suits-jackets"`
				} `json:"searches"`
			} `json:"ALGOLIA"`
			Content struct {
				Cmscontent struct {
					Type     string `json:"type"`
					Elements []struct {
						Type string `json:"type"`
						// Title       string `json:"title,omitempty"`
						Hidden      bool `json:"hidden,omitempty"`
						Productlist []struct {
							Title string `json:"title"`
							Items []struct {
								Attributes struct {
									Productlink struct {
										URI  string `json:"uri"`
										Type string `json:"type"`
									} `json:"productLink"`
								} `json:"attributes,omitempty"`
							} `json:"items"`
						} `json:"productList,omitempty"`
					} `json:"elements"`
				} `json:"cmsContent"`
			} `json:"CONTENT"`
		} `json:"initialReduxState"`
	} `json:"props"`
}

// used to extract embaded json data in website page.
// more about golang regulation see here https://golang.org/pkg/regexp/syntax/
var productsExtractReg = regexp.MustCompile(`(?U)id="__NEXT_DATA__"\s*type="application/json">\s*({.*})\s*</script>`)
var catproductsExtractReg = regexp.MustCompile(`(?U)algoliaJSONP_3({.*});`)
var pageReg = regexp.MustCompile(`&page=\d+`)

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

	if bytes.Contains(respBody, []byte(`<p class="multiline-text search-results-toolbar-no-results-message">Sorry, there is no results for your search`)) {
		fmt.Println(`Page not found`)
		return nil
	}

	matched := productsExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}

	var viewData CategoryStructure
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		return err
	}

	searchString := ""
	lastIndex := nextIndex(ctx)

	if len(viewData.Props.InitialReduxState.ALGOLIA.Searches) > 0 {
		for key, is := range viewData.Props.InitialReduxState.ALGOLIA.Searches {

			searchString = key
			for _, idv := range is.Result.Hits {

				if idv.Type == "EDITOIMAGE" || idv.Type == "EDITOCONTENT" {
					continue
				}

				rawurl := ""
				if idv.Attributes.ProductLink.URI != "" {
					rawurl = fmt.Sprintf("%s://%s%s", resp.Request.URL.Scheme, resp.Request.URL.Host, idv.Attributes.ProductLink.URI)
				} else {
					rawurl = fmt.Sprintf("%s://%s/en_us/products/couture-%s", resp.Request.URL.Scheme, resp.Request.URL.Host, idv.ObjectID)
				}
				//fmt.Println(rawurl)
				req, err := http.NewRequest(http.MethodGet, rawurl, nil)
				if err != nil {
					c.logger.Errorf("load http request of url %s failed, error=%s", rawurl, err)
					return err
				}

				lastIndex += 1
				// set the index of the product crawled in the sub response
				nctx := context.WithValue(ctx, "item.index", lastIndex)
				// yield sub request
				if err := yield(nctx, req); err != nil {
					return err
				}
			}
		}
	} else if len(viewData.Props.InitialReduxState.Content.Cmscontent.Elements) > 0 {
		for _, is := range viewData.Props.InitialReduxState.Content.Cmscontent.Elements {

			if is.Type != "PRODUCTLISTGROUP" {
				continue
			}
			for _, idp := range is.Productlist {
				for _, idv := range idp.Items {

					rawurl := ""
					if idv.Attributes.Productlink.URI != "" {
						rawurl = fmt.Sprintf("%s://%s/en_us/products%s", resp.Request.URL.Scheme, resp.Request.URL.Host, strings.Split(idv.Attributes.Productlink.URI, "/products")[1])
					} else {
						continue
					}
					req, err := http.NewRequest(http.MethodGet, rawurl, nil)
					if err != nil {
						c.logger.Errorf("load http request of url %s failed, error=%s", rawurl, err)
						return err
					}

					lastIndex += 1
					// set the index of the product crawled in the sub response
					nctx := context.WithValue(ctx, "item.index", lastIndex)
					// yield sub request
					if err := yield(nctx, req); err != nil {
						return err
					}
				}
			}
		}
	}

	// get current page number
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}
	// check if this is the last page
	if searchString != "" {
		if len(viewData.Props.InitialReduxState.ALGOLIA.Searches[searchString].Result.Hits) >= viewData.Props.InitialReduxState.ALGOLIA.Searches[searchString].Result.NbHits ||
			page >= int64(viewData.Props.InitialReduxState.ALGOLIA.Searches[searchString].Result.NbPages) {
			return nil
		}
	} else {
		return nil
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

//search API
func (c *_Crawler) parseSearchKeywordProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	// read the response data from http response
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	url := "https://kpgnq6fji9-1.algolianet.com/1/indexes/*/queries?x-algolia-agent=Algolia%2520for%2520JavaScript%2520(3.35.1)%253B%2520Browser&x-algolia-application-id=KPGNQ6FJI9&x-algolia-api-key=64e489d5d73ec5bbc8ef0d7713096fba"

	query := resp.Request.URL.Query().Get("query")
	payload := ""
	if c.categorySearchPathMatcher.MatchString(resp.Request.URL.Path) {

		if query != "" {
			payload = `{"requests":[{"indexName":"dev_product_en_us","params":"query=` + query + `&hitsPerPage=100&maxValuesPerFacet=10&page=0&highlightPreTag=%3Cais-highlight-0000000000%3E&highlightPostTag=%3C%2Fais-highlight-0000000000%3E&clickAnalytics=true&facets=%5B%22universe%22%5D&tagFilters="}]}`
		}

		req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(payload))

		if err != nil {
			fmt.Println(err)
			return err
		}
		req.Header.Add("Content-Type", "application/json")

		res, err := c.httpClient.Do(ctx, req)
		if err != nil {
			fmt.Println(err)
			return err
		}
		defer res.Body.Close()

		respBody, err = ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	var viewData parseSearchKeywordProducts
	if err := json.Unmarshal(respBody, &viewData); err != nil {
		c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		return err
	}

	lastIndex := nextIndex(ctx)
	if len(viewData.Results) == 0 {
		fmt.Println(`not proper response `)
		return nil
	}
	for _, idv := range viewData.Results[0].Hits {

		rawurl := fmt.Sprintf("%s://%s/en_us/products/couture-%s", resp.Request.URL.Scheme, resp.Request.URL.Host, idv.ObjectID)

		//fmt.Println(rawurl)
		req, err := http.NewRequest(http.MethodGet, rawurl, nil)
		if err != nil {
			c.logger.Errorf("load http request of url %s failed, error=%s", rawurl, err)
			return err
		}

		lastIndex += 1
		// set the index of the product crawled in the sub response
		nctx := context.WithValue(ctx, "item.index", lastIndex)
		// yield sub request
		if err := yield(nctx, req); err != nil {
			return err
		}
	}

	// get current page number
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}
	// check if this is the last page
	if viewData.Results == nil {
		return nil
	}

	if len(viewData.Results[0].Hits) >= viewData.Results[0].NbHits ||
		page >= int64(viewData.Results[0].NbPages) {
		return nil
	}

	if query == "" && viewData.Results[0].Params != "" {
		payload = `{"requests":[{"indexName":"dev_product_en_us","params":"` + viewData.Results[0].Params + `"}]}`
	}
	payload = pageReg.ReplaceAllString(payload, fmt.Sprintf("%s%v", "&page=", viewData.Results[0].Page+1))

	reqn, _ := http.NewRequest(http.MethodPost, url, strings.NewReader(payload))
	// update the index of last page
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, reqn)
}

type parseSearchKeywordProducts struct {
	Results []struct {
		Hits []struct {
			ObjectID string `json:"objectID"`
		} `json:"hits"`
		NbHits      int    `json:"nbHits"`
		Page        int    `json:"page"`
		NbPages     int    `json:"nbPages"`
		HitsPerPage int    `json:"hitsPerPage"`
		Params      string `json:"params"`
		Index       string `json:"index"`
		QueryID     string `json:"queryID"`
	} `json:"results"`
	Params string `json:"params"`
}

// used to trim html labels in description
var htmlTrimRegp = regexp.MustCompile(`</?[^>]+>`)

type parseProductData struct {
	Props struct {
		Tracking struct {
			Datalayer struct {
				Ecommerce struct {
					Currencycode string `json:"currencyCode"`
					Detail       struct {
						Products []struct {
							ID       string `json:"id"`
							Brand    string `json:"brand"`
							Category string `json:"category"`
						} `json:"products"`
					} `json:"detail"`
				} `json:"ecommerce"`
				Wwwversion      string `json:"wwwVersion"`
				Pagevirtualpath string `json:"pageVirtualPath"`
			} `json:"dataLayer"`
			Usegtm bool `json:"useGTM"`
		} `json:"tracking"`
		InitialReduxState struct {
			CONTENT struct {
				CmsContent struct {
					Type     string `json:"type"`
					Elements []struct {
						Type            string `json:"type"`
						Text            string `json:"text,omitempty"`
						BackgroundColor string `json:"backgroundColor,omitempty"`
						TextColor       string `json:"textColor,omitempty"`
						TimeToLive      int    `json:"timeToLive,omitempty"`
						Capping         int    `json:"capping,omitempty"`
						NodeID          string `json:"nodeId,omitempty"`
						CallToAction    struct {
							Type  string `json:"type"`
							Title string `json:"title"`
							URI   string `json:"uri"`
						} `json:"callToAction,omitempty"`
						Items []struct {
							Type     string `json:"type"`
							Title    string `json:"title"`
							TitleInt string `json:"titleInt,omitempty"`
							URL      string `json:"url"`
							Sku      string `json:"sku"`
							Images   []*struct {
								Target   string `json:"target"`
								URI      string `json:"uri"`
								Width    int    `json:"width"`
								Height   int    `json:"height"`
								ViewCode string `json:"viewCode"`
								Alt      string `json:"alt"`
							} `json:"images"`
							Video *struct {
								Title           string `json:"title"`
								VideoTitle      string `json:"videoTitle"`
								AdditionalTitle string `json:"additionalTitle"`
								ResourceID      string `json:"resourceId"`
								Poster          struct {
									URI    string `json:"uri"`
									Width  int    `json:"width"`
									Height int    `json:"height"`
									Alt    string `json:"alt"`
								} `json:"poster"`
								Transcript string `json:"transcript"`
							} `json:"video"`
						} `json:"items,omitempty"`
						//Title             string        `json:"title,omitempty"`
						Subtitle          string        `json:"subtitle,omitempty"`
						Reference         string        `json:"reference,omitempty"`
						Tags              []interface{} `json:"tags,omitempty"`
						Universe          string        `json:"universe,omitempty"`
						PrimaryCategoryID string        `json:"primaryCategoryId,omitempty"`
						//ProductType          string        `json:"productType,omitempty"`
						PreorderShippingDate interface{} `json:"preorderShippingDate,omitempty"`
						VariationsType       string      `json:"variationsType,omitempty"`
						Code                 string      `json:"code,omitempty"`
						Color                string      `json:"color,omitempty"`
						Variations           []struct {
							Title  string `json:"title"`
							Code   string `json:"code"`
							Sku    string `json:"sku"`
							Detail string `json:"detail"`
							Status string `json:"status"`
							Ean    string `json:"ean"`
							Price  struct {
								Value    int    `json:"value"`
								Currency string `json:"currency"`
							} `json:"price"`
							Tracking []struct {
								Events        []string `json:"events"`
								AddToCartType string   `json:"addToCartType"`
								PageType      string   `json:"pageType"`
								Ecommerce     struct {
									CurrencyCode string `json:"currencyCode"`
									Add          struct {
										Products []struct {
											ID       string `json:"id"`
											Name     string `json:"name"`
											Price    int    `json:"price"`
											Brand    string `json:"brand"`
											Category string `json:"category"`
											Variant  string `json:"variant"`
											Quantity int    `json:"quantity"`
										} `json:"products"`
									} `json:"add"`
								} `json:"ecommerce"`
							} `json:"tracking"`
							Image struct {
								URI    string `json:"uri"`
								Width  int    `json:"width"`
								Height int    `json:"height"`
								Alt    string `json:"alt"`
							} `json:"image"`
						} `json:"variations,omitempty"`
						Sections []struct {
							TitleKey string `json:"titleKey"`
							Content  string `json:"content"`
							Type     string `json:"type"`
						} `json:"sections,omitempty"`
						Declinations []struct {
							Title     string `json:"title"`
							Color     string `json:"color"`
							ColorCode string `json:"colorCode"`
							URI       string `json:"uri"`
							Image     struct {
								Target   string `json:"target"`
								URI      string `json:"uri"`
								Width    int    `json:"width"`
								Height   int    `json:"height"`
								ViewCode string `json:"viewCode"`
								Alt      string `json:"alt"`
							} `json:"image"`
						} `json:"declinations,omitempty"`
						Images []struct {
							Width  int    `json:"width"`
							Height int    `json:"height"`
							Alt    string `json:"alt"`
							URI    string `json:"uri"`
							Target string `json:"target"`
						} `json:"images,omitempty"`
						Price struct {
							Value    int    `json:"value"`
							Currency string `json:"currency"`
						} `json:"price"`
						Sku         string      `json:"sku"`
						Status      string      `json:"status"`
						SizeLabel   string      `json:"sizeLabel"`
						Description interface{} `json:"description,omitempty"`
					} `json:"elements"`
				} `json:"cmsContent"`
			} `json:"CONTENT"`
		} `json:"initialReduxState"`
		Pageprops struct {
			Identifier string `json:"identifier"`
		} `json:"pageprops"`
	} `json:"props"`
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

	if bytes.Contains(respBody, []byte(`<span class="multiline-text">Our apologies, but we weren't able to find the page you are looking for.</span>`)) {
		fmt.Println(`Page not found`)
		return nil
	}

	matched := productsExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}

	var viewData parseProductData
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}
	dom, err := goquery.NewDocumentFromReader(bytes.NewBuffer(respBody))
	if err != nil {
		return err
	}

	contentIndex := getIndex(viewData, "PRODUCTTITLE")
	if contentIndex == -1 {
		fmt.Println(`PRODUCTTITLE not found`)
		return nil
	}
	contentData := viewData.Props.InitialReduxState.CONTENT.CmsContent.Elements[contentIndex]

	contentIndex = getIndex(viewData, "PRODUCTSECTIONDESCRIPTION")
	if contentIndex == -1 {
		fmt.Println(`PRODUCTSECTIONDESCRIPTION not found`)
		return nil
	}
	contentDescriptionData := viewData.Props.InitialReduxState.CONTENT.CmsContent.Elements[contentIndex]

	canUrl := dom.Find(`link[rel="canonical"]`).AttrOr("href", "")
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}
	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           viewData.Props.Pageprops.Identifier,
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
		},
		BrandName:   viewData.Props.Tracking.Datalayer.Ecommerce.Detail.Products[0].Brand,
		Title:       dom.Find(`.product-titles`).Text(),
		Description: htmlTrimRegp.ReplaceAllString(contentDescriptionData.Sections[0].Content, ""),
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
	}
	if item.BrandName == "" {
		item.BrandName = "DIOR"
	}

	if context.GetString(ctx, "Category") != "" {
		item.Category = context.GetString(ctx, "Category")
		item.SubCategory = context.GetString(ctx, "SubCategory")
		item.SubCategory2 = context.GetString(ctx, "SubCategory2")
	} else {
		for i, breadcrumb := range strings.Split(viewData.Props.Tracking.Datalayer.Ecommerce.Detail.Products[0].Category, "/") {
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
	}

	// image
	var itemImg []*media.Media
	contentIndex = getIndex(viewData, "PRODUCTMEDIAS")
	if contentIndex == -1 {
		fmt.Println(`PRODUCTMEDIAS not found`)
		return nil
	}
	contentImgData := viewData.Props.InitialReduxState.CONTENT.CmsContent.Elements[contentIndex]
	for ki, mid := range contentImgData.Items {
		if mid.Type == "IMAGE" && len(mid.Images) > 0 {
			template := strings.Split(mid.Images[0].URI, "?")
			itemImg = append(itemImg, pbMedia.NewImageMedia(
				strconv.Format(ki),
				template[0],
				template[0]+"?sw=800&sh=800&sm=fit",
				template[0]+"?sw=600&sh=600&sm=fit",
				template[0]+"?sw=500&sh=500&sm=fit",
				"",
				len(itemImg) == 0,
			))
		}
	}

	contentIndex = getIndex(viewData, "PRODUCTVARIATIONS")
	if contentIndex > -1 {
		contentData = viewData.Props.InitialReduxState.CONTENT.CmsContent.Elements[contentIndex]
		if contentData.VariationsType != "SIZE" {
			c.logger.Debugf("Variation is not type of SIZE")
		}

		for j, rawSku := range contentData.Variations {
			var itemImg []*media.Media
			contentIndex = getIndex(viewData, "PRODUCTMEDIAS")
			if contentIndex == -1 {
				fmt.Println(`PRODUCTMEDIAS not found`)
				return nil
			}
			contentImgData := viewData.Props.InitialReduxState.CONTENT.CmsContent.Elements[contentIndex]

			for ki, mid := range contentImgData.Items {
				if mid.Sku != "" && rawSku.Sku != mid.Sku {
					continue
				}
				if mid.Type == "IMAGE" && len(mid.Images) > 0 {
					template := strings.Split(mid.Images[0].URI, "?")
					itemImg = append(itemImg, pbMedia.NewImageMedia(
						strconv.Format(ki),
						template[0],
						template[0]+"?sw=800&sh=800&sm=fit",
						template[0]+"?sw=600&sh=600&sm=fit",
						template[0]+"?sw=500&sh=500&sm=fit",
						"",
						len(itemImg) == 0,
					))
				}
			}

			originalPrice := (rawSku.Price.Value)
			//discount, _ := strconv.ParseInt(strings.TrimSuffix(rawSku.DisplayPercentOff, "%"))
			sku := pbItem.Sku{
				SourceId: strconv.Format(rawSku.Sku),
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(originalPrice * 100),
					Msrp:     int32(originalPrice * 100),
				},
				Medias: itemImg,
				Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}
			if rawSku.Status == "AVAILABLE" || rawSku.Status == "" {
				sku.Stock.StockStatus = pbItem.Stock_InStock
			}

			spacId := rawSku.Ean
			if spacId == "" {
				spacId = rawSku.Sku + strconv.Format(j)
			}
			if contentData.VariationsType == "SIZE" || contentData.VariationsType == "CAPACITY" {
				// size
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecSize,
					Id:    spacId,
					Name:  rawSku.Title,
					Value: rawSku.Title,
				})
			} else {
				// Color
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecColor,
					Id:    spacId,
					Name:  rawSku.Title,
					Value: rawSku.Title,
					Icon:  rawSku.Image.URI,
				})
			}

			item.SkuItems = append(item.SkuItems, &sku)
		}
	} else {
		// no size variation

		contentIndex = getIndex(viewData, "PRODUCTUNIQUE")
		if contentIndex > -1 {

			contentData = viewData.Props.InitialReduxState.CONTENT.CmsContent.Elements[contentIndex]
			originalPrice := (contentData.Price.Value)
			//discount, _ := strconv.ParseInt(strings.TrimSuffix(rawSku.DisplayPercentOff, "%"))
			sku := pbItem.Sku{
				SourceId: strconv.Format(contentData.Sku),
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(originalPrice * 100),
					Msrp:     int32(originalPrice * 100),
				},
				Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}
			if contentData.Status == "AVAILABLE" || contentData.Status == "" {
				sku.Stock.StockStatus = pbItem.Stock_InStock
			}
			if contentData.Color != "" {
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecColor,
					Id:    contentData.Color,
					Name:  contentData.Color,
					Value: contentData.Color,
				})
			}
			if contentData.SizeLabel != "" {
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecSize,
					Id:    contentData.Sku + "1",
					Name:  contentData.SizeLabel,
					Value: contentData.SizeLabel,
				})
			}
			sku.Medias = itemImg
			item.SkuItems = append(item.SkuItems, &sku)
			if contentData.Universe != "" && item.CrowdType == "" {
				item.CrowdType = contentData.Universe
			}
		}
	}

	for _, rawSku := range item.SkuItems {
		if rawSku.Stock.StockStatus == pbItem.Stock_InStock {
			item.Stock = &pbItem.Stock{StockStatus: pbItem.Stock_InStock}
		}
	}
	if item.Stock == nil {
		item.Stock = &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock}
	}

	// yield item result
	if err = yield(ctx, &item); err != nil {
		return err
	}

	return nil
}

func getIndex(viewData parseProductData, types string) int {
	cindex := -1

	for i, raw := range viewData.Props.InitialReduxState.CONTENT.CmsContent.Elements {
		if types == raw.Type {
			cindex = i
		}
	}
	return cindex
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://www.dior.com/en_us",
		//"https://www.dior.com/en_us/womens-fashion/ready-to-wear/all-ready-to-wear",
		//"https://www.dior.com/en_us/fragrance/mens-fragrance/all-products",
		//"https://www.dior.com/en_us/products/beauty-Y0998004-sauvage-parfum",
		// "https://www.dior.com/en_us/products/couture-124V03BM211_X5685-ribbed-knit-bar-jacket-navy-blue-double-breasted-virgin-wool",
		// "https://www.dior.com/en_us/products/couture-93C1046A0121_C975-dior-oblique-tie-blue-and-black-silk",
		//"https://www.dior.com/en_us/products/beauty-Y0061201-jules-eau-de-toilette",
		//"https://www.dior.com/en_us/products/beauty-Y0998004-sauvage-parfum",
		//"https://www.dior.com/en_us/products/beauty-Y0028965-dior-lip-tattoo-color-games-collection-limited-edition-colored-lip-tint-bare-lip-sensation-%E2%80%93-extreme-weightless-wear",
		// "https://www.dior.com/en_us/products/couture-141B19A3842_X4813-dioriviera-blouse-raspberry-toile-de-jouy-reverse-cotton-poplin",
		//"https://www.dior.com/en_us/products/beauty-Y0139000-5-couleurs-couture-eyeshadow-palette-high-pigment-long-wear-creamy-powder",
		//"https://www.dior.com/en_us/fragrance/mens-fragrance/all-products",
		//"https://www.dior.com/en_us/products/couture-HYN01TLC0U_C005-bath-mat-cannage",
		"https://www.dior.com/en_us/products/search?query=women",
		//"https://kpgnq6fji9-1.algolianet.com/1/indexes/*/queries?x-algolia-agent=Algolia%2520for%2520JavaScript%2520(3.35.1)%253B%2520Browser&x-algolia-application-id=KPGNQ6FJI9&x-algolia-api-key=64e489d5d73ec5bbc8ef0d7713096fba",
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