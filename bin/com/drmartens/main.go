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

// _Crawler defined the crawler struct/class for which is not necessary to be exportable
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
		categoryPathMatcher: regexp.MustCompile(`^/us/en(/[A-Za-z0-9_-]+)*/c/\d+$`),
		productPathMatcher:  regexp.MustCompile(`^(/[A-Za-z0-9_-]+)*/p/[A-Za-z0-9_-]+$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "a9458145ce83ddfe996fb7d6ade9bccf"
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
	}

	return opts
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.drmartens.com"}
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
		u.Host = "www.drmartens.com"
	}
	if !strings.HasPrefix(u.Path, "/us/en") {
		u.Path = "/us/en" + u.Path
	}
	if c.productPathMatcher.MatchString(u.Path) {
		u.RawQuery = ""
	}
	return u.String(), nil
}

// Parse is the entry to run the spider.
// ctx is the context of this run. if may contains the shared values in it.
//   you can also set some value by context.WithValue().
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
	if p == "" || p == "/us/en" {
		return crawler.ErrUnsupportedPath
	}
	if c.productPathMatcher.MatchString(p) {
		return c.parseProduct(ctx, resp, yield) // product deatils page
	} else if c.categoryPathMatcher.MatchString(p) {
		return c.parseCategoryProducts(ctx, resp, yield) // category >> productlist page
	}

	return crawler.ErrUnsupportedPath
}

// nextIndex used to get the index from the shared data.
// item.index is a const key for item index.
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")))
}

func (c *_Crawler) GetCategories(ctx context.Context) ([]*pbItem.Category, error) {
	req, _ := http.NewRequest(http.MethodGet, "https://www.drmartens.com/us/en/", nil)
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
		// duplicated url
		//cateUrlMap = map[string]struct{}{}
	)
	if err := func(yield func(names []string, url string) error) error {
		sel := dom.Find(`.dm-primary-nav>li`)

		for i := range sel.Nodes {
			node := sel.Eq(i)
			cateName := strings.TrimSpace(node.Find(`a`).First().Text())
			if cateName == "" {
				continue
			}

			subSel := node.Find(`.sub-navigation-section`)

			for k := range subSel.Nodes {
				subNode2 := subSel.Eq(k)
				subcat := strings.TrimSpace(subNode2.Find(`a`).First().Text())

				subNode2list := subNode2.Find(`.yCmsComponent`)
				for j := range subNode2list.Nodes {
					subNode := subNode2list.Eq(j)

					subcatname := strings.TrimSpace(subNode.Find(`a`).First().Text())

					if subcatname == "" {
						continue
					}

					href := subNode.Find(`a`).First().AttrOr("href", "")
					if href == "" {
						continue
					}

					href, err := c.CanonicalUrl(href)
					if err != nil {
						c.logger.Errorf("got invalid url %s", href)
						continue
					}

					u, err := url.Parse(href)
					if err != nil {
						c.logger.Error("parse url %s failed", href)
						continue
					}

					if c.categoryPathMatcher.MatchString(u.Path) {
						if err := yield([]string{cateName, subcat, subcatname}, href); err != nil {
							return err
						}
						//cateUrlMap[href] = struct{}{}
					}

				}

				if len(subNode2list.Nodes) == 0 {

					href := subNode2.Find(`a`).First().AttrOr("href", "")
					if href == "" {
						continue
					}

					href, err := c.CanonicalUrl(href)
					if err != nil {
						c.logger.Errorf("got invalid url %s", href)
						continue
					}

					u, err := url.Parse(href)
					if err != nil {
						c.logger.Error("parse url %s failed", href)
						continue
					}

					if c.categoryPathMatcher.MatchString(u.Path) {
						if err := yield([]string{cateName, subcat}, href); err != nil {
							return err
						}
						//cateUrlMap[href] = struct{}{}
					}
				}
			}
		}
		return nil
	}(func(names []string, url string) error {
		if len(names) == 0 {
			return errors.New("no valid category name found")
		}
		//if _, ok := cateUrlMap[url]; ok {
		//	c.logger.Warnf("found duplicated url=%s, path=%v", url, names)
		//	return nil
		//}

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

	sel := dom.Find(`.dm-primary-nav>li`)

	for i := range sel.Nodes {
		node := sel.Eq(i)
		cateName := strings.TrimSpace(node.Find(`a`).First().Text())
		if cateName == "" {
			continue
		}

		//nctx := context.WithValue(ctx, "Category", cateName)

		subSel := node.Find(`.sub-navigation-section`)

		for k := range subSel.Nodes {
			subNode2 := subSel.Eq(k)
			subcat := strings.TrimSpace(subNode2.Find(`a`).First().Text())

			subNode2list := subNode2.Find(`.yCmsComponent`)
			for j := range subNode2list.Nodes {
				subNode := subNode2list.Eq(j)

				subcatname := strings.TrimSpace(subNode.Find(`a`).First().Text())

				if subcatname == "" {
					continue
				}

				href := subNode.Find(`a`).First().AttrOr("href", "")
				if href == "" {
					continue
				}

				finalsubCatName := ""
				if subcat != "" {
					finalsubCatName = subcat + " > " + subcatname
				} else {
					finalsubCatName = subcatname
				}

				fmt.Println(finalsubCatName)

				// u, err := url.Parse(href)
				// if err != nil {
				// 	c.logger.Error("parse url %s failed", href)
				// 	continue
				// }

				// if c.categoryPathMatcher.MatchString(u.Path) {
				// 	nnctx := context.WithValue(nctx, "SubCategory", finalsubCatName)
				// 	req, _ := http.NewRequest(http.MethodGet, href, nil)
				// 	if err := yield(nnctx, req); err != nil {
				// 		return err
				// 	}
				// }

			}

			if len(subNode2list.Nodes) == 0 {
				fmt.Println(subcat)
			}
		}
	}
	return nil
}

type CategoryData struct {
	Results []struct {
		Current struct {
			Code                     string  `json:"code"`
			BaseProductCode          string  `json:"baseProductCode"`
			Summary                  string  `json:"summary"`
			Name                     string  `json:"name"`
			URL                      string  `json:"url"`
			SwatchHexCode            string  `json:"swatchHexCode"`
			ThumbnailImgURL          string  `json:"thumbnailImgUrl"`
			AlternateThumbnailImgURL string  `json:"alternateThumbnailImgUrl"`
			FormattedPrice           string  `json:"formattedPrice"`
			LabelHex                 string  `json:"labelHex"`
			InSale                   bool    `json:"inSale"`
			DisplayPriority          float64 `json:"displayPriority"`
		} `json:"current,omitempty"`
		Siblings []struct {
			Code                     string  `json:"code"`
			BaseProductCode          string  `json:"baseProductCode"`
			Summary                  string  `json:"summary"`
			Name                     string  `json:"name"`
			URL                      string  `json:"url"`
			SwatchHexCode            string  `json:"swatchHexCode"`
			ThumbnailImgURL          string  `json:"thumbnailImgUrl"`
			AlternateThumbnailImgURL string  `json:"alternateThumbnailImgUrl"`
			FormattedPrice           string  `json:"formattedPrice"`
			LabelHex                 string  `json:"labelHex,omitempty"`
			InSale                   bool    `json:"inSale"`
			DisplayPriority          float64 `json:"displayPriority,omitempty"`
			NoOfReviews              int     `json:"noOfReviews,omitempty"`
		} `json:"siblings"`
		Sid     string  `json:"sid"`
		Ratings float64 `json:"ratings"`
		Reviews int     `json:"reviews"`
	} `json:"results"`
	Pagination struct {
		NumberOfPages        int `json:"numberOfPages"`
		TotalNumberOfResults int `json:"totalNumberOfResults"`
		PageSize             int `json:"pageSize"`
		CurrentPage          int `json:"currentPage"`
	} `json:"pagination"`
	Facets []struct {
		Code        string `json:"code"`
		Name        string `json:"name"`
		MultiSelect bool   `json:"multiSelect"`
		Visible     bool   `json:"visible"`
		Values      []struct {
			Code  string `json:"code"`
			Name  string `json:"name"`
			Count int    `json:"count"`
			Query struct {
				Value string `json:"value"`
			} `json:"query"`
			Selected bool `json:"selected"`
		} `json:"values"`
		FacetID string `json:"facetId"`
	} `json:"facets"`
	SortFields []struct {
		Code     string `json:"code"`
		Selected bool   `json:"selected,omitempty"`
		Name     string `json:"name"`
		Desc     bool   `json:"desc,omitempty"`
	} `json:"sortFields"`
	Rid string `json:"rid"`
}

type categoryResultResp struct {
	Results []struct {
		Current struct {
			Code                     string      `json:"code"`
			BaseProductCode          string      `json:"baseProductCode"`
			Summary                  interface{} `json:"summary"`
			Name                     string      `json:"name"`
			URL                      string      `json:"url"`
			ColourSwatchImgURL       interface{} `json:"colourSwatchImgUrl"`
			SwatchHexCode            string      `json:"swatchHexCode"`
			ThumbnailImgURL          string      `json:"thumbnailImgUrl"`
			AlternateThumbnailImgURL string      `json:"alternateThumbnailImgUrl"`
			FormattedPrice           string      `json:"formattedPrice"`
			FormattedWasPrice        interface{} `json:"formattedWasPrice"`
			MiniDescription          interface{} `json:"miniDescription"`
			LabelText                interface{} `json:"labelText"`
			LabelHex                 interface{} `json:"labelHex"`
			InSale                   bool        `json:"inSale"`
			DisplayPriority          interface{} `json:"displayPriority"`
			NoOfReviews              interface{} `json:"noOfReviews"`
		} `json:"current"`
		Siblings []struct {
			Code                     string      `json:"code"`
			BaseProductCode          string      `json:"baseProductCode"`
			Summary                  interface{} `json:"summary"`
			Name                     string      `json:"name"`
			URL                      string      `json:"url"`
			ColourSwatchImgURL       interface{} `json:"colourSwatchImgUrl"`
			SwatchHexCode            string      `json:"swatchHexCode"`
			ThumbnailImgURL          string      `json:"thumbnailImgUrl"`
			AlternateThumbnailImgURL string      `json:"alternateThumbnailImgUrl"`
			FormattedPrice           string      `json:"formattedPrice"`
			FormattedWasPrice        interface{} `json:"formattedWasPrice"`
			MiniDescription          interface{} `json:"miniDescription"`
			LabelText                interface{} `json:"labelText"`
			LabelHex                 interface{} `json:"labelHex"`
			InSale                   bool        `json:"inSale"`
			DisplayPriority          interface{} `json:"displayPriority"`
			NoOfReviews              int         `json:"noOfReviews"`
		} `json:"siblings"`
		Sid     string  `json:"sid"`
		Ratings float64 `json:"ratings"`
		Reviews int     `json:"reviews"`
	} `json:"results"`
	Pagination struct {
		NumberOfPages        int         `json:"numberOfPages"`
		TotalNumberOfResults int         `json:"totalNumberOfResults"`
		PageSize             int         `json:"pageSize"`
		CurrentPage          int         `json:"currentPage"`
		Sort                 interface{} `json:"sort"`
		ViewPage             interface{} `json:"viewPage"`
		SourcePage           interface{} `json:"sourcePage"`
		SelectedProduct      interface{} `json:"selectedProduct"`
	} `json:"pagination"`
	Facets []struct {
		Code        string `json:"code"`
		Name        string `json:"name"`
		MultiSelect bool   `json:"multiSelect"`
		Visible     bool   `json:"visible"`
		Values      []struct {
			Code  string `json:"code"`
			Name  string `json:"name"`
			Count int    `json:"count"`
			Query struct {
				URL   interface{} `json:"url"`
				Value string      `json:"value"`
			} `json:"query"`
			Selected  bool        `json:"selected"`
			RestQuery interface{} `json:"restQuery"`
		} `json:"values"`
		FacetID string `json:"facetId"`
	} `json:"facets"`
	SortFields []struct {
		Code     string      `json:"code"`
		Desc     interface{} `json:"desc"`
		Selected bool        `json:"selected"`
		Name     string      `json:"name"`
	} `json:"sortFields"`
	Rid         string      `json:"rid"`
	RedirectURL interface{} `json:"redirectUrl"`
	Carousels   interface{} `json:"carousels"`
}

// parseCategoryProducts parse api url from web page url
func (c *_Crawler) parseCategoryProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	//respBody, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	c.logger.Debug(err)
	//	return err
	//}
	//
	//var viewData CategoryData
	//matched := productsListExtractReg.FindSubmatch([]byte(respBody))
	//if len(matched) > 1 {
	//
	//	if err := json.Unmarshal(matched[1], &viewData); err != nil {
	//		c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
	//		return err
	//	}
	//}
	//if err := json.Unmarshal(matched[1], &viewData); err != nil {
	//	c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
	//	return err
	//}

	// 发送请求获取所有的商品列表
	// https://www.drmartens.com/us/en/womens/c/01000000/results?page=0
	productReqList := make([]*http.Request, 0, 1000)
	for i := 0; i <= 99999; i++ {
		u := *resp.Request.URL
		if strings.HasSuffix(u.Path, "/") {
			u.Path += "results"
		} else {
			u.Path += "/results"
		}
		vals := url.Values{}
		vals.Set("page", strconv.Format(i))
		u.RawQuery = vals.Encode()
		req, err := http.NewRequest(http.MethodGet, u.String(), nil)
		if err != nil {
			c.logger.Error(err)
			return err
		}
		opts := c.CrawlOptions(resp.Request.URL)
		for k := range opts.MustHeader {
			req.Header.Set(k, opts.MustHeader.Get(k))
		}
		req.Header.Set("accept-encoding", "gzip, deflate, br")
		req.Header.Set("accept", "*/*")
		req.Header.Set("Referer", resp.Request.URL.String())
		req.Header.Set("User-Agent", resp.Request.Header.Get("User-Agent"))

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

		categoryResp, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
			EnableProxy:    true,
			EnableHeadless: c.CrawlOptions(resp.Request.URL).EnableHeadless,
			Reliability:    c.CrawlOptions(resp.Request.URL).Reliability,
		})
		if err != nil {
			c.logger.Error(err)
			return err
		}

		if categoryResp.StatusCode != http.StatusOK {
			c.logger.Errorf("status is %v", categoryResp.StatusCode)
			return fmt.Errorf(categoryResp.Status)
		}

		data, err := io.ReadAll(categoryResp.Body)
		if err != nil {
			c.logger.Error(err)
			return err
		}
		categoryResp.Body.Close()

		var categoryResultRespData categoryResultResp
		if err := json.Unmarshal(data, &categoryResultRespData); err != nil {
			c.logger.Errorf("%s, error=%s", data, err)
			return err
		}
		if len(categoryResultRespData.Results) == 0 {
			break
		}

		// send product request
		for _, result := range categoryResultRespData.Results {
			productUrl, err := c.CanonicalUrl(result.Current.URL)
			if err != nil {
				c.logger.Errorf("got invalid url %s", result.Current.URL)
				continue
			}
			productReq, err := http.NewRequest(http.MethodGet, productUrl, nil)
			if err != nil {
				c.logger.Errorf("load http request of url %s failed, error=%s", productUrl, err)
				return err
			}
			productReqList = append(productReqList, productReq)
		}
	}

	lastIndex := nextIndex(ctx)
	for _, productReq := range productReqList {
		lastIndex += 1
		// set the index of the product crawled in the sub response
		nctx := context.WithValue(ctx, "item.index", lastIndex)
		// yield sub request
		if err := yield(nctx, productReq); err != nil {
			return err
		}
	}

	return nil
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
	return bytes.TrimSpace(resp)
}

func TrimSpaceNewlineInByte(s []byte) []byte {
	re := regexp.MustCompile(`\n`)
	resp := re.ReplaceAll(s, []byte(" "))
	resp = bytes.ReplaceAll(resp, []byte("\\n"), []byte(""))
	resp = bytes.ReplaceAll(resp, []byte("\r"), []byte(""))
	resp = bytes.ReplaceAll(resp, []byte("\t"), []byte(""))
	resp = bytes.ReplaceAll(resp, []byte("&lt;"), []byte("<"))
	resp = bytes.ReplaceAll(resp, []byte("&gt;"), []byte(">"))
	resp = bytes.ReplaceAll(resp, []byte("  "), []byte(" "))

	resp = bytes.ReplaceAll(resp, []byte("} , }"), []byte("} }"))

	return bytes.TrimSpace(resp)
}

var productDetailExtractReg = regexp.MustCompile(`(?Ums)ACC.productTabs.tabsData\s*=\s*({.*});`)
var productsExtractReg = regexp.MustCompile(`(?Ums)<script type="application/ld\+json">\s*({.*})\s*</script>`)
var productsListExtractReg = regexp.MustCompile(`(?Ums)ACC.productList.initPageLoad\(({.*})\);`)

type ProductPageData struct {
	Context     string `json:"@context"`
	Type        string `json:"@type"`
	ID          string `json:"@id"`
	Name        string `json:"name"`
	Image       string `json:"image"`
	Description string `json:"description"`
	Mpn         string `json:"mpn"`
	URL         string `json:"url"`
	Brand       struct {
		Type string `json:"@type"`
		Name string `json:"name"`
	} `json:"brand"`
	Offers struct {
		Type          string `json:"@type"`
		Price         string `json:"price"`
		PriceCurrency string `json:"priceCurrency"`
		ItemCondition string `json:"itemCondition"`
		Seller        struct {
			Type string `json:"@type"`
			Name string `json:"name"`
		} `json:"seller"`
	} `json:"offers"`
}

type ProductDetailData struct {
	ProdDetail struct {
		CloseTitle string `json:"closeTitle"`
		ViewTitle  string `json:"viewTitle"`
		Title      string `json:"title"`
		Content    string `json:"content"`
	} `json:"prodDetail"`
	HowMade struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	} `json:"howMade"`
	Yotpo struct {
		CloseTitle string `json:"closeTitle"`
		ViewTitle  string `json:"viewTitle"`
		Title      string `json:"title"`
		Sku        string `json:"sku"`
		Page       string `json:"page"`
		Reviews    int    `json:"reviews"`
		Language   string `json:"language"`
	} `json:"yotpo"`
}

type orderPlaced struct {
	WebsiteCountry  string `json:"website_country"`
	Dds             bool   `json:"dds"`
	ProductViewType string `json:"product_view_type"`
	PageType        string `json:"page_type"`
	PageName        string `json:"page_name"`
	Ecommerce       struct {
		Detail struct {
			Products []struct {
				ProductColour       string `json:"product_colour"`
				Code                string `json:"code"`
				ProductDiscount     string `json:"product_discount"`
				Price               int    `json:"price"`
				ProductMaterial     string `json:"product_material"`
				ProductAvailability int    `json:"product_availability"`
				Name                string `json:"name"`
				Variants            []struct {
					Code     string `json:"code"`
					Variants []struct {
						Code string `json:"code"`
					} `json:"variants"`
				} `json:"variants"`
				Brand string `json:"brand"`
			} `json:"products"`
		} `json:"detail"`
		CurrencyCode string `json:"currencyCode"`
	} `json:"ecommerce"`
	GaID          string `json:"ga_id"`
	VisitorStatus string `json:"visitor_status"`
	CurrencyCode  string `json:"currencyCode"`
}

// used to trim html labels in description
var htmlTrimRegp = regexp.MustCompile(`</?[^>]+>`)
var orderPlacedDataRegp = regexp.MustCompile(`(?U)orderPlacedData\s*=\s*(.*)\s*</script>`)

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

	var viewData ProductPageData
	var viewDetail ProductDetailData
	matched := productsExtractReg.FindSubmatch(respBody)
	if len(matched) > 1 {
		if err := json.Unmarshal(matched[1], &viewData); err != nil {
			c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
			return err
		}
	}

	matched = productDetailExtractReg.FindSubmatch(respBody)
	if len(matched) > 1 {
		if err := json.Unmarshal(TrimSpaceNewlineInByte(matched[1]), &viewDetail); err != nil {
			c.logger.Errorf("unmarshal Detail data fetched from %s failed, error=%s", resp.Request.URL, err)
			//return err
		}
	}

	canUrl, _ := c.CanonicalUrl(doc.Find(`link[rel="canonical"]`).AttrOr("href", ""))
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}

	brand := viewData.Brand.Name
	if brand == "" {
		brand = "Dr. Martens"
	}

	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           viewData.Mpn,
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
	item.Description = string(TrimSpaceNewlineInString([]byte(htmlTrimRegp.ReplaceAllString(viewDetail.ProdDetail.Content+viewDetail.HowMade.Content, " "))))

	currentPrice, _ := strconv.ParsePrice(doc.Find(`span[class="current-price special-price"]`).Text())
	msrp, _ := strconv.ParsePrice(doc.Find(`.original-price`).Text())

	if msrp == 0 {
		msrp = currentPrice
	}
	discount := 0
	if msrp > currentPrice {
		discount = int(((msrp - currentPrice) / msrp) * 100)
	}

	//images
	sel := doc.Find(`.slider-pdp-nav-thumbnails`).Find(`picture`)
	for j := range sel.Nodes {
		node := sel.Eq(j)
		imgurl := strings.Split(node.Find(`img`).AttrOr(`data-src`, ``), "?")[0]

		item.Medias = append(item.Medias, pbMedia.NewImageMedia(
			strconv.Format(j),
			imgurl,
			imgurl+"?$large$",
			imgurl+"?$medium$",
			imgurl+"?$mediumtablet$",
			"", j == 0))
	}

	// itemListElement
	sel = doc.Find(`.breadcrumb>li`)
	c.logger.Debugf("nodes %d", len(sel.Nodes))
	for i := range sel.Nodes {
		if i == len(sel.Nodes)-1 {
			continue
		}
		node := sel.Eq(i)
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

	details := map[string]json.RawMessage{}

	// Color no cid
	//cid := ""
	colorName := ""
	var colorSelected *pbItem.SkuSpecOption
	sel = doc.Find(`.variant-list.js-variant-list`).Find(`li`)
	for i := range sel.Nodes {
		node := sel.Eq(i)

		if spanClass := node.AttrOr(`class`, ``); strings.Contains(spanClass, `active`) {
			if err := json.Unmarshal([]byte(node.Find(`a`).AttrOr("data-json", "")), &details); err != nil {
				c.logger.Errorf("json Unmarshal err=%s", err)
				continue
			}
			//cid = strings.Trim(string(details["code"]), `"`)
			// 这项数据有的网页没有
			colorName = strings.Trim(string(details["name"]), `"`)
			if colorName == "" {
				// 这项数据有的网页没有
				matched := orderPlacedDataRegp.FindSubmatch(respBody)
				if len(matched) <= 1 {
					return fmt.Errorf("extract orderPlacedData json from product page %s failed", resp.Request.URL.String())
				}

				var orderPlacedData orderPlaced
				if err = json.Unmarshal(bytes.TrimSpace(matched[1]), &orderPlacedData); err != nil {
					c.logger.Debugf("parse orderPlacedData=%s failed, error=%s", matched[1], err)
					return err
				}
				if len(orderPlacedData.Ecommerce.Detail.Products) < 1 {
					colorName = "-"
				} else {
					colorName = strings.ToUpper(orderPlacedData.Ecommerce.Detail.Products[0].ProductColour)
				}
			}
			colorSelected = &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    colorName,
				Name:  colorName,
				Value: colorName,
				Icon:  strings.Trim(string(details["img"]), `"`),
			}
		} else {
			// send other product yield
			otherProductHref := node.Find(`a`).AttrOr("href", "")
			if canonOtherProductHref, err := c.CanonicalUrl(otherProductHref); err != nil {
				c.logger.Errorf("invalid product href=%s, url=%s, err=%s", otherProductHref, resp.Request.URL.String(), err)
				return err
			} else {
				otherProductHref = canonOtherProductHref
			}
			if req, err := http.NewRequest(http.MethodGet, otherProductHref, nil); err != nil {
				return err
			} else if err = yield(ctx, req); err != nil {
				return err
			}
		}
	}

	sizeSel := doc.Find(`#sizeSelector li`)
	// kid,male and female
	for i := range sizeSel.Nodes {
		node := sizeSel.Eq(i).Find(`a`)
		// 因为不同的男女的sku-size的sku-code有可能是一样的，所以这里进行了拼接
		sid := node.AttrOr(`data-sku-code`, ``) + "-" + node.AttrOr(`data-sku-size`, ``)
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

		if spanClass := node.AttrOr("class", ""); strings.Contains(spanClass, "stock-inStock") {
			sku.Stock.StockStatus = pbItem.Stock_InStock
			if item.GetStock().GetStockStatus() == pbItem.Stock_OutOfStock {
				item.Stock.StockStatus = pbItem.Stock_InStock
			}
		}

		if colorSelected != nil {
			sku.Specs = append(sku.Specs, colorSelected)
		}

		// size
		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Id:    node.AttrOr("data-sku-size", ""),
			Name:  node.AttrOr("data-label", ""),
			Value: node.AttrOr("data-sku-size", ""),
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

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://www.drmartens.com/us/en/",
		//"https://www.drmartens.com/us/en/womens/boots/c/01010000",
		"https://www.drmartens.com/us/en/p/26228100",
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
