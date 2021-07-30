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
	pbProxy "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
)

// _Crawler defined the crawler struct/class for which is not necessory to be exportable
type _Crawler struct {
	// httpClient is the object of an http client
	httpClient          http.Client
	searchPathMatcher   *regexp.Regexp
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
		httpClient:        client,
		searchPathMatcher: regexp.MustCompile(`^/search/?$`),
		// this regular used to match category page url path
		categoryPathMatcher: regexp.MustCompile(`^/(shop|beauty|brand)(/[a-z0-9_\-]+){0,4}$`),
		// this regular used to match product page url path
		productPathMatcher: regexp.MustCompile(`^/product(/[a-zA-Z0-9\pL\pS_\-]+){0,4}$`),
		logger:             logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	// every spider should got an unique id which should not larget than 64 in length
	return "b91cac7f1f123f59da130d53d8f71d50"
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
		EnableSessionInit: true,
		Reliability:       pbProxy.ProxyReliability_ReliabilityMedium,
	}
	opts.MustCookies = append(opts.MustCookies,
		&http.Cookie{Name: "site_language", Value: "en"},
		&http.Cookie{Name: "device_type", Value: "desktop"},
		&http.Cookie{Name: "site_locale", Value: "us"},
		&http.Cookie{Name: "current_country", Value: "US"},
	)
	return opts
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.sephora.com"}
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
		u.Host = "www.sephora.com"
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

	p := strings.TrimSuffix(resp.RawUrl().Path, "/")

	if p == "" {
		return c.parseCategories(ctx, resp, yield)
	}
	if c.searchPathMatcher.MatchString(resp.RawUrl().Path) {
		return c.parseSearch(ctx, resp, yield)
	} else if c.categoryPathMatcher.MatchString(resp.RawUrl().Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	} else if c.productPathMatcher.MatchString(resp.RawUrl().Path) {
		return c.parseProduct(ctx, resp, yield)
	}
	return crawler.ErrUnsupportedPath
}

// nextIndex used to get the index from the shared data.
// item.index is a const key for item index.
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")))
}

type categoryStructure struct {
	Header struct {
		Props struct {
			Headerfootercontent struct {
				Rwdnavigationmenu []struct {
					Componentlist []struct {
						Componentlist []struct {
							Targeturl string `json:"targetUrl,omitempty"`
							Titletext string `json:"titleText,omitempty"`
						} `json:"componentList"`
						Targeturl string `json:"targetUrl,omitempty"`
						Titletext string `json:"titleText,omitempty"`
					} `json:"componentList"`
					Targeturl string `json:"targetUrl,omitempty"`
					Titletext string `json:"titleText"`
				} `json:"rwdNavigationMenu"`
			} `json:"headerFooterContent"`
			Isdownloadappbannerenabled bool `json:"isDownloadAppBannerEnabled"`
		} `json:"props"`
	} `json:"Header"`
}

func (c *_Crawler) parseCategories(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	matched := productsExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		//c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract category info from %s failed, error=%s", resp.Request.URL, err)
	}

	var viewData categoryStructure
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal cat detail data fialed, error=%s", err)
		return err
	}

	for _, rawCat := range viewData.Header.Props.Headerfootercontent.Rwdnavigationmenu {
		cateName := rawCat.Titletext
		if cateName == "" {
			continue
		}
		baseUrl := ""
		if rawCat.Targeturl != "" {
			baseUrl = "https://www.sephora.com" + strings.Split(strings.Split(rawCat.Targeturl, "/")[0], "-")[0]
		} else {
			baseUrl = "https://www.sephora.com"
		}
		nnctx := context.WithValue(ctx, "Category", cateName)
		//fmt.Println(`cateName `, cateName)
		for _, rawsubCat := range rawCat.Componentlist {

			if rawsubCat.Componentlist != nil {

				for _, rawsub2Cat := range rawsubCat.Componentlist {

					href := baseUrl + rawsub2Cat.Targeturl
					if rawsub2Cat.Targeturl == "" {
						continue
					}

					u, err := url.Parse(href)
					if err != nil {
						c.logger.Error("parse url %s failed", href)
						continue
					}

					subCateName := rawsubCat.Titletext + " > " + rawsub2Cat.Titletext
					//fmt.Println(subCateName, "  ==>  ", href)
					if c.categoryPathMatcher.MatchString(u.Path) {
						nnnctx := context.WithValue(nnctx, "SubCategory", subCateName)
						req, _ := http.NewRequest(http.MethodGet, href, nil)
						if err := yield(nnnctx, req); err != nil {
							return err
						}
					}
				}
			} else {
				href := baseUrl + rawsubCat.Targeturl
				if rawsubCat.Targeturl == "" {
					continue
				}

				u, err := url.Parse(href)
				if err != nil {
					c.logger.Error("parse url %s failed", href)
					continue
				}

				subCateName := rawsubCat.Titletext
				//fmt.Println(subCateName, "  ==>  ", href)
				if c.categoryPathMatcher.MatchString(u.Path) {
					nnnctx := context.WithValue(nnctx, "SubCategory", subCateName)
					req, _ := http.NewRequest(http.MethodGet, href, nil)
					if err := yield(nnnctx, req); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

type rawSearchResponse struct {
	Categories []struct {
		DisplayName string `json:"displayName"`
		Level       int    `json:"level"`
		NodeStr     string `json:"nodeStr"`
		RecordCount string `json:"recordCount"`
	} `json:"categories"`
	Keyword  string `json:"keyword"`
	Products []struct {
		BrandName   string `json:"brandName"`
		DisplayName string `json:"displayName"`
		HeroImage   string `json:"heroImage"`
		Image135    string `json:"image135"`
		Image250    string `json:"image250"`
		Image450    string `json:"image450"`
		ProductID   string `json:"productId"`
		ProductName string `json:"productName"`
		Rating      string `json:"rating"`
		Reviews     string `json:"reviews"`
		TargetURL   string `json:"targetUrl"`
		URL         string `json:"url"`
		MoreColors  int    `json:"moreColors,omitempty"`
	} `json:"products"`
	Refinements []struct {
		DisplayName string `json:"displayName"`
		Type        string `json:"type"`
		Values      []struct {
			Low  int `json:"low,omitempty"`
			High int `json:"high,omitempty"`
		} `json:"values"`
	} `json:"refinements"`
	ResponseSource string `json:"responseSource"`
	TotalProducts  int    `json:"totalProducts"`
	UserSegment    string `json:"userSegment"`
}

func (c *_Crawler) parseSearch(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || resp == nil || yield == nil {
		return nil
	}

	keywords := strings.TrimSpace(resp.Request.URL.Query().Get("keyword"))
	if keywords == "" {
		return nil
	}
	currentPage, _ := strconv.ParseInt(resp.Request.URL.Query().Get("currentPage"))
	if currentPage == 0 {
		currentPage = 1
	}

	const pageSize = 60

	apiUrl, _ := url.Parse("https://www.sephora.com/api/catalog/search?type=keyword&q=&content=true&includeRegionsMap=true&targetSearchEngine=nlp")
	// TODO: &constructorSessionID=8&constructorClientID=3298983f-8d6b-4d57-85d6-329f3f60e6e4&
	vals := apiUrl.Query()
	vals.Set("q", keywords)
	vals.Set("currentPage", strconv.Format(currentPage))
	vals.Set("page", strconv.Format(pageSize))

	if cookies, _ := c.httpClient.Jar().Cookies(ctx, resp.Request.URL); len(cookies) > 0 {
		for _, c := range cookies {
			if c.Name == "ConstructorioID_client_id" {
				vals.Set("constructorClientID", c.Value)
				break
			}
		}
	}
	apiUrl.RawQuery = vals.Encode()

	req, _ := http.NewRequest(http.MethodGet, apiUrl.String(), nil)
	req.Header.Set("Referer", resp.Request.URL.String())

	opts := c.CrawlOptions(req.URL)
	// init custom headers
	for k := range opts.MustHeader {
		req.Header.Set(k, opts.MustHeader.Get(k))
	}
	// init custom cookies
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

	apiResp, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
		EnableProxy: true,
		Reliability: opts.Reliability,
	})
	if err != nil {
		c.logger.Error(err)
		return err
	}
	defer apiResp.Body.Close()

	rawData, err := io.ReadAll(apiResp.Body)
	if err != nil {
		c.logger.Error(err)
		return err
	}
	if apiResp.StatusCode == http.StatusForbidden ||
		apiResp.StatusCode == -1 {
		return fmt.Errorf("api request failed, status=%d, %s", apiResp.StatusCode, rawData)
	}

	var viewData rawSearchResponse
	if err := json.Unmarshal(rawData, &viewData); err != nil {
		c.logger.Error(err)
		return err
	}
	lastIndex := nextIndex(ctx)
	for _, prod := range viewData.Products {
		req, err := http.NewRequest(http.MethodGet, prod.TargetURL, nil)
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
	if pageSize*currentPage >= int64(viewData.TotalProducts) {
		return nil
	}

	vals = resp.Request.URL.Query()
	vals.Set("currentPage", strconv.Format(currentPage+1))
	u := *resp.Request.URL
	u.RawQuery = vals.Encode()

	nextReq, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, nextReq)
}

// below are the golang json data struct of raw website.
// if you get the raw json data of the website,
// then you can use https://mholt.github.io/json-to-go/ to convert it to a golang struct

type ProductCategoryStructure struct {
	Nthcategory struct {
		Class string `json:"class"`
		Props struct {
			Seocategoryname      string `json:"seoCategoryName"`
			Breadcrumbsseojsonld string `json:"breadcrumbsSeoJsonLd"`
			Categories           []struct {
				Categoryid      string `json:"categoryId"`
				Displayname     string `json:"displayName"`
				Hasdropdownmenu bool   `json:"hasDropdownMenu"`
				Level           int    `json:"level"`
				Subcategories   []struct {
					Categoryid      string `json:"categoryId"`
					Displayname     string `json:"displayName"`
					Hasdropdownmenu bool   `json:"hasDropdownMenu"`
					Level           int    `json:"level"`
					Subcategories   []struct {
						Categoryid      string `json:"categoryId"`
						Displayname     string `json:"displayName"`
						Hasdropdownmenu bool   `json:"hasDropdownMenu"`
						Isselected      bool   `json:"isSelected,omitempty"`
						Level           int    `json:"level"`
						Targeturl       string `json:"targetUrl"`
					} `json:"subCategories"`
					Targeturl string `json:"targetUrl"`
				} `json:"subCategories"`
				Targeturl string `json:"targetUrl"`
			} `json:"categories"`
			Categoryid           string `json:"categoryId"`
			Displayname          string `json:"displayName"`
			Enablenoindexmetatag bool   `json:"enableNoindexMetaTag"`
			Navigationseojsonld  string `json:"navigationSeoJsonLd"`
			Pagesize             int64  `json:"pageSize"`
			Parentcategory       struct {
				Categoryid  string `json:"categoryId"`
				Displayname string `json:"displayName"`
				Targeturl   string `json:"targetUrl"`
			} `json:"parentCategory"`
			Products []struct {
				Brandname        string `json:"brandName"`
				Displayname      string `json:"displayName"`
				Heroimage        string `json:"heroImage"`
				Image135         string `json:"image135"`
				Image250         string `json:"image250"`
				Image450         string `json:"image450"`
				Morecolors       int    `json:"moreColors,omitempty"`
				Productid        string `json:"productId"`
				Rating           string `json:"rating"`
				Reviews          string `json:"reviews"`
				Targeturl        string `json:"targetUrl"`
				URL              string `json:"url"`
				Heroimagealttext string `json:"heroImageAltText,omitempty"`
			} `json:"products"`
			Productsandoffersseojsonld string        `json:"productsAndOffersSeoJsonLd"`
			Responsesource             string        `json:"responseSource"`
			Seocanonicalurl            string        `json:"seoCanonicalUrl"`
			Seokeywords                []interface{} `json:"seoKeywords"`
			Seometadescription         string        `json:"seoMetaDescription"`
			Seoname                    string        `json:"seoName"`
			Seotitle                   string        `json:"seoTitle"`
			Targeturl                  string        `json:"targetUrl"`
			Template                   int           `json:"template"`
			Totalproducts              int64         `json:"totalProducts"`
			URL                        string        `json:"url"`
			Thirdpartyimagehost        string        `json:"thirdpartyImageHost"`
		} `json:"props"`
	} `json:"NthCategory"`
}

// used to extract embaded json data in website page.
// more about golang regulation see here https://golang.org/pkg/regexp/syntax/
var productsExtractReg = regexp.MustCompile(`(?U)<script id="linkSPA" type="text/json" data-comp="PageJSON\s*">({.*})</script>`)

// parseCategoryProducts parse api url from web page url
func (c *_Crawler) parseCategoryProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}
	c.logger.Debugf("%s", resp.Request.URL)

	// read the response data from http response
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	matched := productsExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}

	var viewData ProductCategoryStructure
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		return err
	}

	lastIndex := nextIndex(ctx)
	for _, idv := range viewData.Nthcategory.Props.Products {
		rawurl := idv.Targeturl
		req, err := http.NewRequest(http.MethodGet, rawurl, nil)
		if err != nil {
			c.logger.Errorf("load http request of url %s failed, error=%s", rawurl, err)
			return err
		}
		if strings.HasSuffix(req.URL.Path, ".html") {
			req.URL.RawQuery = ""
		}
		// set the index of the product crawled in the sub response
		nctx := context.WithValue(ctx, "item.index", lastIndex)
		lastIndex += 1
		// yield sub request
		if err := yield(nctx, req); err != nil {
			return err
		}
	}

	// get current page number
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("currentPage"))
	if page == 0 {
		page = 1
	}
	totalPageCount := viewData.Nthcategory.Props.Totalproducts
	if viewData.Nthcategory.Props.Pagesize*page > totalPageCount {
		return nil
	}

	// set pagination
	u := *resp.Request.URL
	vals := u.Query()
	vals.Set("currentPage", strconv.Format(page+1))
	u.RawQuery = vals.Encode()

	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	// update the index of last page
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
}

type parseProductBreadCrumbData struct {
	ItemListElement []struct {
		Item struct {
			Name string `json:"name"`
			ID   string `json:"@id"`
		} `json:"item"`
	} `json:"itemListElement"`
}

type childSKU struct {
	AlternateImages []struct {
		AltText  string `json:"altText"`
		ImageURL string `json:"imageUrl"`
	} `json:"alternateImages"`
	BiExclusiveLevel    string `json:"biExclusiveLevel"`
	DisplayName         string `json:"displayName"`
	IngredientDesc      string `json:"ingredientDesc"`
	IsAppExclusive      bool   `json:"isAppExclusive"`
	IsBiReward          bool   `json:"isBiReward"`
	IsFirstAccess       bool   `json:"isFirstAccess"`
	IsFree              bool   `json:"isFree"`
	IsLimitedEdition    bool   `json:"isLimitedEdition"`
	IsLimitedTimeOffer  bool   `json:"isLimitedTimeOffer"`
	IsNew               bool   `json:"isNew"`
	IsOnlineOnly        bool   `json:"isOnlineOnly"`
	IsOnlyFewLeft       bool   `json:"isOnlyFewLeft"`
	IsOutOfStock        bool   `json:"isOutOfStock"`
	IsPickUpEligibleSku bool   `json:"isPickUpEligibleSku"`
	IsRopisEligibleSku  bool   `json:"isRopisEligibleSku"`
	IsSephoraExclusive  bool   `json:"isSephoraExclusive"`
	ListPrice           string `json:"listPrice"`
	MaxPurchaseQuantity int    `json:"maxPurchaseQuantity"`
	PrimarySkinTone     string `json:"primarySkinTone"`
	Refinements         struct {
		FinishRefinements []string `json:"finishRefinements"`
	} `json:"refinements"`
	SalePrice string `json:"salePrice"`
	Size      string `json:"size"`
	SkuID     string `json:"skuId"`
	SkuImages struct {
		ImageURL string `json:"imageUrl"`
	} `json:"skuImages"`
	SmallImage               string `json:"smallImage"`
	TargetURL                string `json:"targetUrl"`
	Type                     string `json:"type"`
	VariationDesc            string `json:"variationDesc"`
	VariationType            string `json:"variationType"`
	VariationTypeDisplayName string `json:"variationTypeDisplayName"`
	VariationValue           string `json:"variationValue"`
}

type parseProductData struct {
	Page struct {
		Product struct {
			Breadcrumbsseojsonld string `json:"breadcrumbsSeoJsonLd"`
			Content              struct {
				Seocanonicalurl    string        `json:"seoCanonicalUrl"`
				Seokeywords        []interface{} `json:"seoKeywords"`
				Seometadescription string        `json:"seoMetaDescription"`
				Seoname            string        `json:"seoName"`
				Seotitle           string        `json:"seoTitle"`
				Targeturl          string        `json:"targetUrl"`
			} `json:"content"`
			Currentsku struct {
				Alternateimages []struct {
					Alttext  string `json:"altText"`
					Imageurl string `json:"imageUrl"`
				} `json:"alternateImages"`
				Biexclusivelevel string `json:"biExclusiveLevel"`
				Displayname      string `json:"displayName"`
				Highlights       []struct {
					Alttext  string `json:"altText"`
					ID       string `json:"id"`
					Imageurl string `json:"imageUrl"`
					Name     string `json:"name"`
				} `json:"highlights"`
				Ingredientdesc      string `json:"ingredientDesc"`
				Isappexclusive      bool   `json:"isAppExclusive"`
				Isbireward          bool   `json:"isBiReward"`
				Isfirstaccess       bool   `json:"isFirstAccess"`
				Isfree              bool   `json:"isFree"`
				Islimitededition    bool   `json:"isLimitedEdition"`
				Islimitedtimeoffer  bool   `json:"isLimitedTimeOffer"`
				Isnew               bool   `json:"isNew"`
				Isonlineonly        bool   `json:"isOnlineOnly"`
				Isonlyfewleft       bool   `json:"isOnlyFewLeft"`
				Isoutofstock        bool   `json:"isOutOfStock"`
				Ispickupeligiblesku bool   `json:"isPickUpEligibleSku"`
				Isropiseligiblesku  bool   `json:"isRopisEligibleSku"`
				Issephoraexclusive  bool   `json:"isSephoraExclusive"`
				Listprice           string `json:"listPrice"`
				Maxpurchasequantity int    `json:"maxPurchaseQuantity"`
				Primaryskintone     string `json:"primarySkinTone"`
				Refinements         struct {
					Finishrefinements []string `json:"finishRefinements"`
				} `json:"refinements"`
				Size      string `json:"size"`
				Skuid     string `json:"skuId"`
				Skuimages struct {
					Imageurl string `json:"imageUrl"`
					Alttext  string `json:"altText"`
				} `json:"skuImages"`
				Smallimage               string `json:"smallImage"`
				Targeturl                string `json:"targetUrl"`
				Type                     string `json:"type"`
				Variationdesc            string `json:"variationDesc"`
				Variationtype            string `json:"variationType"`
				Variationtypedisplayname string `json:"variationTypeDisplayName"`
				Variationvalue           string `json:"variationValue"`
			} `json:"currentSku"`
			Enablenoindexmetatag bool `json:"enableNoindexMetaTag"`
			Flashbanner          struct {
				Ancestorhierarchy []struct {
					Displayname  string `json:"displayName"`
					Nodestatus   int    `json:"nodeStatus"`
					Targetscreen struct {
						Apiurl       string `json:"apiUrl"`
						Targetscreen string `json:"targetScreen"`
						Targeturl    string `json:"targetUrl"`
						Targetvalue  string `json:"targetValue"`
					} `json:"targetScreen"`
				} `json:"ancestorHierarchy"`
				Mediatype int `json:"mediaType"`
				Regions   struct {
					Content []struct {
						Componentname          string `json:"componentName"`
						Componenttype          int    `json:"componentType"`
						Contenttype            string `json:"contentType"`
						Enabletesting          bool   `json:"enableTesting"`
						Modalcomponenttemplate struct {
							Componentname string `json:"componentName"`
							Componenttype int    `json:"componentType"`
							Components    []struct {
								Alttext       string `json:"altText"`
								Componentname string `json:"componentName"`
								Componenttype int    `json:"componentType"`
								Enabletesting bool   `json:"enableTesting"`
								Height        string `json:"height"`
								Imageid       string `json:"imageId"`
								Imagepath     string `json:"imagePath"`
								Name          string `json:"name"`
								Targetscreen  struct {
									Targetscreen string `json:"targetScreen"`
									Targeturl    string `json:"targetUrl"`
									Targetvalue  string `json:"targetValue"`
									Targetwindow int    `json:"targetWindow"`
								} `json:"targetScreen"`
								Width string `json:"width"`
							} `json:"components"`
							Design        string `json:"design"`
							Enabletesting bool   `json:"enableTesting"`
							Name          string `json:"name"`
							Scrollable    bool   `json:"scrollable"`
						} `json:"modalComponentTemplate"`
						Name         string `json:"name"`
						Targetwindow string `json:"targetWindow"`
						Text         string `json:"text"`
					} `json:"content"`
				} `json:"regions"`
				Seocanonicalurl string `json:"seoCanonicalUrl"`
				Seoname         string `json:"seoName"`
				Targeturl       string `json:"targetUrl"`
				Templateurl     string `json:"templateUrl"`
				Title           string `json:"title"`
				Type            string `json:"type"`
			} `json:"flashBanner"`
			Fullsiteproducturl     string `json:"fullSiteProductUrl"`
			Ishidesocial           bool   `json:"isHideSocial"`
			Isreverselookupenabled bool   `json:"isReverseLookupEnabled"`
			Navigationseojsonld    string `json:"navigationSeoJsonLd"`
			Productdetails         struct {
				Brand struct {
					Brandid         string `json:"brandId"`
					Description     string `json:"description"`
					Displayname     string `json:"displayName"`
					Longdescription string `json:"longDescription"`
					Ref             string `json:"ref"`
					Targeturl       string `json:"targetUrl"`
				} `json:"brand"`
				Displayname      string  `json:"displayName"`
				Imagealttext     string  `json:"imageAltText"`
				Longdescription  string  `json:"longDescription"`
				Lovescount       int     `json:"lovesCount"`
				Productid        string  `json:"productId"`
				Rating           float64 `json:"rating"`
				Reviews          int     `json:"reviews"`
				Shortdescription string  `json:"shortDescription"`
				Suggestedusage   string  `json:"suggestedUsage"`
			} `json:"productDetails"`
			Productseojsonld string      `json:"productSeoJsonLd"`
			Regularchildskus []*childSKU `json:"regularChildSkus"`
			OnSaleChildSkus  []*childSKU `json:"onSaleChildSkus"`
		} `json:"product"`
	} `json:"page"`
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

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}
	linkStoreData := strings.Trim(dom.Find(`#linkStore`).Text(), " \r\n\b")
	if linkStoreData == "" {
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}

	var viewData parseProductData
	if err := json.Unmarshal([]byte(linkStoreData), &viewData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}

	var productBreadCrumb parseProductBreadCrumbData
	if err := json.Unmarshal([]byte(viewData.Page.Product.Breadcrumbsseojsonld), &productBreadCrumb); err != nil {
		c.logger.Errorf("unmarshal product breadcrumb data fialed, error=%s", err)
	}

	canUrl := dom.Find(`link[rel="canonical"]`).AttrOr("href", "")
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           viewData.Page.Product.Productdetails.Productid,
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
		},
		BrandName:   viewData.Page.Product.Productdetails.Brand.Displayname,
		Title:       viewData.Page.Product.Productdetails.Displayname,
		Description: viewData.Page.Product.Productdetails.Longdescription,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stats: &pbItem.Stats{
			ReviewCount: int32(viewData.Page.Product.Productdetails.Reviews),
			Rating:      float32(viewData.Page.Product.Productdetails.Rating),
		},
	}

	for i, prodBreadcrumb := range productBreadCrumb.ItemListElement {
		switch i {
		case 0:
			item.Category = prodBreadcrumb.Item.Name
		case 1:
			item.SubCategory = prodBreadcrumb.Item.Name
		case 2:
			item.SubCategory2 = prodBreadcrumb.Item.Name
		case 3:
			item.SubCategory3 = prodBreadcrumb.Item.Name
		case 4:
			item.SubCategory4 = prodBreadcrumb.Item.Name
		}
	}

	childSkuList := viewData.Page.Product.Regularchildskus
	if len(viewData.Page.Product.OnSaleChildSkus) > 0 {
		childSkuList = append(viewData.Page.Product.OnSaleChildSkus)
	}

	for _, rawSku := range childSkuList {
		originalPrice, _ := strconv.ParsePrice(rawSku.SalePrice)
		msrp, _ := strconv.ParsePrice(rawSku.ListPrice)
		if originalPrice == 0 {
			originalPrice = msrp
		}
		discount := 0.0
		if msrp > originalPrice {
			discount = ((msrp - originalPrice) / msrp) * 100
		}
		sku := pbItem.Sku{
			SourceId: rawSku.SkuID,
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(originalPrice * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},
			Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}

		if !rawSku.IsOutOfStock {
			sku.Stock.StockStatus = pbItem.Stock_InStock
		}

		if rawSku.VariationType == "Color" {
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    rawSku.VariationValue,
				Name:  fmt.Sprintf("%s - %s", rawSku.VariationValue, rawSku.VariationDesc),
				Value: rawSku.VariationValue,
				Icon:  "https://www.sephora.com/" + rawSku.SmallImage,
			})
		}

		if rawSku.Size != "" {
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    rawSku.Size,
				Name:  rawSku.Size,
				Value: rawSku.Size,
			})
		}

		// main images
		template := "https://www.sephora.com/" + rawSku.SkuImages.ImageURL
		sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
			strconv.Format(0),
			template, template+"?imwidth=1000", template+"?imwidth=600", template+"?imwidth=500", "", true,
		))
		// secondary images
		for ki, m := range rawSku.AlternateImages {
			template := "https://www.sephora.com/" + m.ImageURL
			sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
				strconv.Format(ki+1),
				template, template+"?imwidth=1000", template+"?imwidth=600", template+"?imwidth=500", "", false,
			))
		}
		item.SkuItems = append(item.SkuItems, &sku)
		item.Medias = sku.Medias
	}

	for _, rawSku := range item.SkuItems {
		if rawSku.Stock.StockStatus == pbItem.Stock_InStock {
			item.Stock = &pbItem.Stock{StockStatus: pbItem.Stock_InStock}
		}
	}

	if item.Stock == nil {
		item.Stock = &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock}
	}

	if len(item.SkuItems) > 0 {
		// yield item result
		if err = yield(ctx, &item); err != nil {
			return err
		}
	} else {
		return errors.New("no sku spec found")
	}
	return nil
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://www.sephora.com",
		//"https://www.sephora.com/search?keyword=skin",
		// "https://www.sephora.com/shop/foundation-makeup",
		"https://www.sephora.com/product/make-no-mistake-foundation-concealer-stick-P420440?skuId=1887520&icid2=products%20grid:p420440",
		// "https://www.sephora.com/product/briogeo-scalp-revival-charcoal-tea-tree-cooling-hydration-mask-dry-itchy-scalp-mask-P469440?icid2=products%20grid:p469440",
		// "https://www.sephora.com/product/marc-jacobs-beauty-extra-shot-caffeine-concealer-foundation-P468838?icid2=products%20grid:p468838",
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
