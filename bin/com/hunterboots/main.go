package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
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
	categoryPathApiMatcher *regexp.Regexp
	categoryPathMatcher    *regexp.Regexp
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
		categoryPathApiMatcher: regexp.MustCompile(`^(/us/en_us/api/catalog/products/[/A-Za-z0-9_-]+)$`),
		categoryPathMatcher:    regexp.MustCompile(`^/us/en_us/([/A-Za-z0-9_-]+)$`),
		//productPathMatcher:  regexp.MustCompile(`^(/[/A-Za-z0-9_-]+.html)$`),
		productPathMatcher: regexp.MustCompile(`^/us/en_us/(.*)+\d+$`),

		logger: logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "4f9a4aefe3a247089bcb28f37babd7e3"
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
	opts.MustHeader.Add(`cookie`, `1P_JAR=2021-09-08-07; NID=223=XZE69l0bnn9VPRT-b7Nxvci75O9V0cxHwCe9vAHX9bfCA9yBKTLbfM69X_X_Hgb08kbid2I_PFKykeO9hok9wMqXnRWq20MWVX1H77BRNz9-JdIxI0ShNmgvcHMWCLiQkyk7V_ApGM89wBOrcDMyhHJRyv1DrZvTDekOCuABE2Y`)

	return opts
}

// AllowedDomains return the domains this spider process will.
// the controller will filter the responses and transfer the matched response to this spider.
// the returned domains is matched in glob regulation.
// more about glob regulation see here https://golang.org/pkg/path/filepath/#Match
func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.hunterboots.com"}
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
		u.Host = "www.hunterboots.com"
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
	fmt.Println(p)

	if p == "" {
		return c.parseCategories(ctx, resp, yield)
	}
	if c.productPathMatcher.MatchString(p) {
		return c.parseProduct(ctx, resp, yield)
	} else if c.categoryPathMatcher.MatchString(p) || c.categoryPathApiMatcher.MatchString(p) {
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

	rootUrl := "https://www.hunterboots.com/us/en_us"
	req, _ := http.NewRequest(http.MethodGet, rootUrl, nil)
	opts := c.CrawlOptions(req.URL)
	req.Header.Set("accept-language", "en-GB,en-US;q=0.9,en")
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")

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

		sel := dom.Find(`.navigation__grid-container>.navigation__sections>li`)

		c.logger.Val("sel.Nodes", len(sel.Nodes))

		for i := range sel.Nodes {

			node := sel.Eq(i)
			catname := strings.TrimSpace(node.Find(`a h2`).First().Text())
			if catname == "" {
				continue
			}

			fmt.Println(`catname `, catname)

			subSel := node.Find(`.navigation-column>.navigation-linkblock`)
			for k := range subSel.Nodes {
				subNode2 := subSel.Eq(k)
				subcat1 := strings.TrimSpace(subNode2.Find(`.navigation-category__link h3`).Text())

				subNode2list := subNode2.Find(`ul>li`)
				for j := range subNode2list.Nodes {
					subNode := subNode2list.Eq(j)
					subcat2 := strings.TrimSpace(subNode.Find(`.image-with-description__title`).First().Text())

					if subcat2 == "" {
						subcat2 = strings.TrimSpace(subNode.Find(`a`).First().Text())
					}

					href := subNode.Find(`a`).First().AttrOr("href", "")

					if href == "" {
						continue
					}

					if !strings.Contains(href, `https://www.hunterboots.com`) {
						href = "https://www.hunterboots.com" + href
					}

					finalsubCatName := ""
					if subcat1 != "" {
						finalsubCatName = subcat1 + " >> " + subcat2
					} else {
						finalsubCatName = subcat2
					}

					fmt.Println(`SubCategory:`, finalsubCatName)
					fmt.Println(`href:`, href)

					canonicalHref, err := c.CanonicalUrl(href)
					if err != nil {
						c.logger.Errorf("got invalid url %s", href)
						continue
					}

					u, _ := url.Parse(canonicalHref)

					if c.categoryPathMatcher.MatchString(u.Path) {
						if err := yield([]string{catname, subcat1, subcat2}, canonicalHref); err != nil {
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

	sel := dom.Find(`.navigation__grid-container>.navigation__sections>li`)
	for i := range sel.Nodes {
		node := sel.Eq(i)
		cateName := strings.TrimSpace(node.Find(`a h2`).First().Text())
		if cateName == "" {
			continue
		}

		//nctx := context.WithValue(ctx, "Category", cateName)
		fmt.Println(`Cat Name:`, cateName)

		subSel := node.Find(`.navigation-column>.navigation-linkblock`)

		for k := range subSel.Nodes {
			subNode2 := subSel.Eq(k)
			subcat := strings.TrimSpace(subNode2.Find(`.navigation-category__link h3`).Text())

			subNode2list := subNode2.Find(`ul>li`)
			for j := range subNode2list.Nodes {
				subNode := subNode2list.Eq(j)
				subcatname := strings.TrimSpace(subNode.Find(`.image-with-description__title`).First().Text())

				if subcatname == "" {
					subcatname = strings.TrimSpace(subNode.Find(`a`).First().Text())
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

				fmt.Println(`SubCategory:`, finalsubCatName)
				fmt.Println(`href:`, fullurl)

				// u, err := url.Parse(href)
				// if err != nil {
				// 	c.logger.Error("parse url %s failed", href)
				// 	continue
				// }

				// if c.categoryPathMatcher.MatchString(u.Path) {
				// 	nnctx := context.WithValue(nctx, "SubCategory", finalsubCatName)
				// 	req, _ := http.NewRequest(http.MethodGet, fullurl, nil)
				// 	if err := yield(nnctx, req); err != nil {
				// 		return err
				// 	}
				// }

			}
		}
	}
	return nil
}

type categoryProductsResponse struct {
	Meta struct {
		Page struct {
			MaxPerPage int `json:"max_per_page"`
			Pages      int `json:"pages"`
			Results    int `json:"results"`
			Page       int `json:"page"`
		} `json:"page"`
		Facets []struct {
			Name        string `json:"name"`
			DisplayName string `json:"displayName"`
			Type        string `json:"type"`
			Values      []struct {
				Value        string `json:"value"`
				DisplayValue string `json:"displayValue"`
				Count        int    `json:"count"`
				Applied      bool   `json:"applied"`
			} `json:"values"`
		} `json:"facets"`
	} `json:"meta"`
	Links struct {
		Self  string `json:"self"`
		First string `json:"first"`
		Last  string `json:"last"`
		Next  string `json:"next"`
	} `json:"links"`
	Data []struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Attributes struct {
			ProductReference string `json:"productReference"`
			Name             string `json:"name"`
			URL              string `json:"url"`
			Pricing          struct {
				Currency     string `json:"currency"`
				TaxInclusive bool   `json:"taxInclusive"`
				Pricing      string `json:"pricing"`
				Min          struct {
					Rendered string `json:"rendered"`
					Currency string `json:"currency"`
					Value    int    `json:"value"`
				} `json:"min"`
				Max struct {
					Rendered string `json:"rendered"`
					Currency string `json:"currency"`
					Value    int    `json:"value"`
				} `json:"max"`
				MaxRetail struct {
					Rendered string `json:"rendered"`
					Currency string `json:"currency"`
					Value    int    `json:"value"`
				} `json:"maxRetail"`
			} `json:"pricing"`
			Images []struct {
				Ratio   string `json:"ratio"`
				RiasURL string `json:"riasUrl"`
			} `json:"images"`
			Attributes struct {
				ColorCode     string      `json:"colorCode"`
				ColorName     string      `json:"colorName"`
				ColorHex      []string    `json:"colorHex"`
				ColorGroupHex []string    `json:"colorGroupHex"`
				CategoryPath  string      `json:"categoryPath"`
				Rating        interface{} `json:"rating"`
			} `json:"attributes"`
			Tags []struct {
				Collection        string `json:"collection,omitempty"`
				Feature           string `json:"feature,omitempty"`
				ScheduledMessage  string `json:"scheduled-message,omitempty"`
				WeatherRatingRain string `json:"weather-rating-rain,omitempty"`
			} `json:"tags"`
		} `json:"attributes"`
	} `json:"data"`
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
	ioutil.WriteFile("D:\\STS5\\New_VoilaCrawl\\VoilaCrawler\\Output_1.html", respBody, 0644)
	lastIndex := nextIndex(ctx)
	var viewData categoryProductsResponse
	fmt.Println(`resp.Request. `, resp.Request.URL.Path)
	if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		s := strings.Split(strings.TrimSuffix(resp.Request.URL.Path, `/`), `/`)
		pid := s[len(s)-1]

		rootURL := "https://www.hunterboots.com/us/en_us/api/catalog/products/" + pid + "/us/EUPG01/en_US/?page[number]=1&page[size]=24"

		respBodyC := c.categoryProductsRequest(ctx, rootURL, resp.Request.URL.String())
		ioutil.WriteFile("D:\\STS5\\New_VoilaCrawl\\VoilaCrawler\\Output"+strconv.Format(lastIndex)+".html", respBodyC, 0644)
		fmt.Println("done...")

		if err := json.Unmarshal(respBodyC, &viewData); err != nil {
			c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
			return err
		}
	}

	if len(viewData.Data) == 0 {
		fmt.Println(`not proper response `)
		return nil
	}

	for _, idv := range viewData.Data {

		rawurl, _ := c.CanonicalUrl(idv.Attributes.URL)

		fmt.Println(lastIndex, " ", rawurl)
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
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("page[number]"))
	fmt.Println(`page `, page)

	// check if this is the last page
	if lastIndex >= viewData.Meta.Page.Results || page >= int64(viewData.Meta.Page.Pages) {
		return nil
	}

	// set pagination
	u := *resp.Request.URL
	vals := u.Query()
	vals.Set("page[number]", strconv.Format(page+1))
	u.RawQuery = vals.Encode()

	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	// update the index of last page
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
}

func (c *_Crawler) categoryProductsRequest(ctx context.Context, url string, referer string) []byte {

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	opts := c.CrawlOptions(req.URL)

	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("referer", "https://www.hunterboots.com"+referer)
	req.Header.Add(`cookie`, `_gcl_au=1.1.1746811912.1630126168; _ga=GA1.2.125166996.1630126168; crl8.fpcuid=124f0070-d28b-4cc3-9a49-1969f3c8aa4a; _hjid=d732b178-1cc3-4e37-ac2c-d43afcc3363e; GlobalE_CT_Data={"CUID":"483184138.449817819.321","CHKCUID":null}; _scid=c1590e95-00a6-4f80-8883-5653adaea8e2; BVBRANDID=7dfb9cbb-d44b-418f-a2e7-0411f849ad6d; _gid=GA1.2.834711025.1631158687; GlobalE_Full_Redirect=false; token=undefined; GlobalE_SupportThirdPartCookies=true; _sctr=1|1631125800000; _aeaid=332b79ea-b641-4761-ba4d-284681aa9cf8; aeatstartmessage=true; GlobalE_Data={"countryISO":"US","currencyCode":"GBP","cultureCode":"en-US"}; stc114663=env:1631165241|20211010052721|20210909055721|1|1041477:20220909052721|uid:1630126171568.1906915484.1347685.114663.1499012269:20220909052721|srchist:1041477:1631165241:20211010052721:20220909052721|tsa:1631165241880.459461039.38041353.763137507379525.1:20210909055721; _hjAbsoluteSessionInProgress=0; _hjIncludedInSessionSample=0; skip_geocode=1; _dc_gtm_UA-11730184-1=1; ometria=2_cid=XaHecdX9PJSd3slP&nses=9&osts=1630126171&sid=2344ee93Gh9E4yXHMGKd&npv=3&tids=&slt=1631253182; stc113516=env:1631251701|20211119052821|20210910062302|3|1028364:20220910055302|uid:1631158756340.184967259.8643341.113516.1967005747.:20220910055302|srchist:1028364:1631251701:20211119052821:20220910055302|tsa:1631251701803.615200736.3742375.4468504780622091.4:20210910062302; ABTasty=uid=9n2sh6yb4npwqm27&fst=1630126168027&pst=1631242644166&cst=1631249083004&ns=14&pvt=70&pvis=4&th=650924.807921.56.4.8.1.1631158754013.1631253238430.1_745561.926508.21.8.5.1.1631158754136.1631170661662.1; ABTastySession=mrasn=&sen=14&lp=https%3A%2F%2Fwww.hunterboots.com%2Fus%2Fen_us%2Fmens-insulated-boots`)

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

	ioutil.WriteFile("D:\\STS5\\New_VoilaCrawl\\VoilaCrawler\\Output_categoryProducts.html", respBody, 0644)
	return respBody
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
	ID                  int    `json:"id"`
	Reference           string `json:"reference"`
	Name                string `json:"name"`
	URL                 string `json:"url"`
	StyleCode           string `json:"styleCode"`
	ColorName           string `json:"colorName"`
	ColorCode           string `json:"colorCode"`
	CategoryPath        string `json:"categoryPath"`
	AvailableToPurchase bool   `json:"availableToPurchase"`
	Tags                []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"tags"`
	Images []struct {
		Ratio   string `json:"ratio"`
		RiasURL string `json:"riasUrl"`
	} `json:"images"`
	Pricing struct {
		AsString   string `json:"asString"`
		Min        string `json:"min"`
		MinAsMoney struct {
			Value    int    `json:"value"`
			Currency string `json:"currency"`
		} `json:"minAsMoney"`
		Max        string `json:"max"`
		MaxAsMoney struct {
			Value    int    `json:"value"`
			Currency string `json:"currency"`
		} `json:"maxAsMoney"`
		MaxRetail        string `json:"maxRetail"`
		MaxRetailAsMoney struct {
			Value    int    `json:"value"`
			Currency string `json:"currency"`
		} `json:"maxRetailAsMoney"`
		Currency       string `json:"currency"`
		TaxIsInclusive bool   `json:"taxIsInclusive"`
	} `json:"pricing"`
	Siblings []struct {
		ID        int      `json:"id"`
		Reference string   `json:"reference"`
		URL       string   `json:"url"`
		ColorName string   `json:"colorName"`
		ColorHex  []string `json:"colorHex"`
		ColorCode string   `json:"colorCode"`
		Images    []struct {
			Ratio   string `json:"ratio"`
			RiasURL string `json:"riasUrl"`
		} `json:"images"`
		Pricing struct {
			AsString   string `json:"asString"`
			Min        string `json:"min"`
			MinAsMoney struct {
				Value    int    `json:"value"`
				Currency string `json:"currency"`
			} `json:"minAsMoney"`
			Max        string `json:"max"`
			MaxAsMoney struct {
				Value    int    `json:"value"`
				Currency string `json:"currency"`
			} `json:"maxAsMoney"`
			MaxRetail        string `json:"maxRetail"`
			MaxRetailAsMoney struct {
				Value    int    `json:"value"`
				Currency string `json:"currency"`
			} `json:"maxRetailAsMoney"`
			Currency       string `json:"currency"`
			TaxIsInclusive bool   `json:"taxIsInclusive"`
		} `json:"pricing"`
	} `json:"siblings"`
	Icons struct {
		Feature []struct {
			Image       string `json:"image"`
			Description string `json:"description"`
		} `json:"feature"`
	} `json:"icons"`
	ProductDetails struct {
		Specification []struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"specification"`
		HasWarranty bool   `json:"hasWarranty"`
		Description string `json:"description"`
		Features    string `json:"features"`
	} `json:"productDetails"`
	Meta struct {
		Title         string `json:"title"`
		Description   string `json:"description"`
		PageReference string `json:"pageReference"`
		CanonicalURL  string `json:"canonicalUrl"`
	} `json:"meta"`
}
type parseProductVariationResponse []struct {
	Ean13       string `json:"ean13"`
	Size        string `json:"size"`
	HasStock    bool   `json:"hasStock"`
	LowStock    bool   `json:"lowStock"`
	CanPreOrder bool   `json:"canPreOrder"`
	CanNotify   bool   `json:"canNotify"`
}

var productsExtractReg = regexp.MustCompile(`(?Ums)<script type="application/ld\+json">\s*({.*})\s*</script>`)

var productsNewExtractReg = regexp.MustCompile(`(?Ums)window\['REACT_QUERY_INITIAL_DATA'\]\.concat\(\[({.*})\]\);`)

type ProductPageData struct {
	Context     string `json:"@context"`
	Type        string `json:"@type"`
	Sku         string `json:"sku"`
	Name        string `json:"name"`
	Image       string `json:"image"`
	Description string `json:"description"`
	Brand       struct {
		Type string `json:"@type"`
		Name string `json:"name"`
	} `json:"brand"`
	Offers []struct {
		Type          string `json:"@type"`
		PriceCurrency string `json:"priceCurrency"`
		Price         int    `json:"price"`
		Sku           string `json:"sku"`
		Gtin          string `json:"gtin"`
		URL           string `json:"url"`
		ItemCondition string `json:"itemCondition"`
		Availability  string `json:"availability"`
		Seller        struct {
			Type string `json:"@type"`
			Name string `json:"name"`
		} `json:"seller"`
	} `json:"offers"`
	AggregateRating struct {
		Type        string `json:"@type"`
		BestRating  string `json:"bestRating"`
		RatingValue string `json:"ratingValue"`
		ReviewCount string `json:"reviewCount"`
	} `json:"aggregateRating"`
	Review []struct {
		Type          string `json:"@type"`
		Author        string `json:"author"`
		DatePublished string `json:"datePublished"`
		ReviewBody    string `json:"reviewBody"`
		Name          string `json:"name"`
		ReviewRating  struct {
			Type        string `json:"@type"`
			BestRating  string `json:"bestRating"`
			RatingValue string `json:"ratingValue"`
			WorstRating string `json:"worstRating"`
		} `json:"reviewRating"`
	} `json:"review"`
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

	ioutil.WriteFile("D:\\STS5\\New_VoilaCrawl\\VoilaCrawler\\Output_Product.html", respBody, 0644)

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	var viewDataProduct ProductPageData
	matched := productsExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}

	if err := json.Unmarshal(matched[1], &viewDataProduct); err != nil {
		c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		return err
	}

	// s := strings.Split(resp.Request.URL.Path, `/`)
	// pid := s[len(s)-1]

	rootURL := "https://www.hunterboots.com/us/en_us/api/product/skus/EUPG01/" + viewDataProduct.Sku

	respBodyV, _ := c.variationRequest(ctx, rootURL, resp.Request.URL.String())

	var viewDataVariation parseProductVariationResponse

	if err := json.Unmarshal(respBodyV, &viewDataVariation); err != nil {
		c.logger.Errorf("unmarshal parseProductVariation data fetched from %s failed, error=%s", resp.Request.URL, err)
		return err
	}

	var viewDataDetail parseProductResponse
	matched = productsNewExtractReg.FindSubmatch(respBody)
	if len(matched) > 1 {

		ret := map[string]json.RawMessage{}
		if err := json.Unmarshal((matched[1]), &ret); err != nil {
			fmt.Println(err)
		}

		//ret["data"] = bytes.TrimSuffix(bytes.TrimPrefix(bytes.ReplaceAll(bytes.ReplaceAll(ret["data"], []byte(`\\`), []byte(``)), []byte(`\"`), []byte(`"`)), []byte(`"`)), []byte(`"`))

		ret["data"] = bytes.TrimSuffix(bytes.TrimPrefix(bytes.ReplaceAll(ret["data"], []byte(`\"`), []byte(`"`)), []byte(`"`)), []byte(`"`))

		if err := json.Unmarshal([]byte(html.UnescapeString(string(ret["data"]))), &viewDataDetail); err != nil {
			c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
			return err
		}
	}

	canUrl, _ := c.CanonicalUrl(doc.Find(`link[rel="canonical"]`).AttrOr("href", ""))
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}

	brand := viewDataProduct.Brand.Name
	if brand == "" {
		brand = "Hunter Boot Ltd"
	}

	reviews, _ := strconv.ParseInt(viewDataProduct.AggregateRating.ReviewCount)
	rating, _ := strconv.ParseFloat(viewDataProduct.AggregateRating.RatingValue)

	// build product data
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           strconv.Format(viewDataProduct.Sku),
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
		},
		BrandName: brand,
		Title:     viewDataDetail.Name,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		Stats: &pbItem.Stats{
			ReviewCount: int32(reviews),
			Rating:      float32(rating),
		},
	}

	// desc
	description := viewDataDetail.ProductDetails.Description + viewDataDetail.ProductDetails.Features
	item.Description = htmlTrimRegp.ReplaceAllString(description, ``)

	//images
	var medias []*pbMedia.Media
	for j, mid := range viewDataDetail.Images {
		imgurl := mid.RiasURL
		if strings.HasPrefix(mid.RiasURL, `//`) {
			imgurl = "https:" + mid.RiasURL
		}
		imgurl = strings.ReplaceAll(strings.ReplaceAll(imgurl, `{quality}`, `85`), `{extension}`, `webp`)

		medias = append(medias, pbMedia.NewImageMedia(
			strconv.Format(j),
			strings.ReplaceAll(imgurl, `{width}`, `600`),
			strings.ReplaceAll(imgurl, `{width}`, `1000`),
			strings.ReplaceAll(imgurl, `{width}`, `800`),
			strings.ReplaceAll(imgurl, `{width}`, `500`),
			"", j == 0))
	}

	item.Medias = medias
	for i, breadcrumb := range strings.Split(viewDataDetail.CategoryPath, `/`) {

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

	currentPrice, _ := strconv.ParsePrice(viewDataDetail.Pricing.MinAsMoney.Value)
	msrp, _ := strconv.ParsePrice(viewDataDetail.Pricing.MaxAsMoney.Value)

	if msrp == 0 {
		msrp = currentPrice
	}
	discount := 0
	if msrp > currentPrice {
		discount = (int)(((msrp - currentPrice) / msrp) * 100)
	}

	for _, rawSku := range viewDataVariation {

		sku := pbItem.Sku{
			SourceId: fmt.Sprintf(rawSku.Ean13),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(currentPrice * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},
			//Medias: medias,
			Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}

		if rawSku.HasStock {
			sku.Stock.StockStatus = pbItem.Stock_InStock
			item.Stock.StockStatus = pbItem.Stock_InStock
		}

		// color
		if viewDataDetail.ColorName != "" {
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    viewDataDetail.ColorName,
				Name:  viewDataDetail.ColorName,
				Value: viewDataDetail.ColorName,
			})
		}

		// size
		if rawSku.Size != "" {
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    rawSku.Size,
				Name:  rawSku.Size,
				Value: rawSku.Size,
			})
		}

		if viewDataDetail.ColorName == "" && rawSku.Size == "" {
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    "-",
				Name:  "-",
				Value: "-",
			})
		}

		item.SkuItems = append(item.SkuItems, &sku)
	}

	if len(viewDataVariation) == 0 {

		sku := pbItem.Sku{
			SourceId: fmt.Sprintf(viewDataProduct.Sku),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(currentPrice * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},
			Medias: medias,
			Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}

		if viewDataDetail.AvailableToPurchase {
			sku.Stock.StockStatus = pbItem.Stock_InStock
			item.Stock.StockStatus = pbItem.Stock_InStock
		}

		// color
		if viewDataDetail.ColorName != "" {
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    viewDataDetail.ColorName,
				Name:  viewDataDetail.ColorName,
				Value: viewDataDetail.ColorName,
			})
		} else {
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    "-",
				Name:  "-",
				Value: "-",
			})
		}

	}

	// yield item result
	if err = yield(ctx, &item); err != nil {
		c.logger.Errorf("yield sub request failed, error=%s", err)
		return err
	}

	// other products
	if ctx.Value("groupId") == nil {
		//nctx := context.WithValue(ctx, "groupId", item.GetSource().GetId())
		for _, colorSizeOption := range viewDataDetail.Siblings {
			if colorSizeOption.Reference == viewDataDetail.Reference {
				continue
			}
			nextProductUrl := fmt.Sprintf("https://www.hunterboots.com%s", colorSizeOption.URL)
			fmt.Println(colorSizeOption.ColorName, " ", nextProductUrl)
			// if req, err := http.NewRequest(http.MethodGet, nextProductUrl, nil); err != nil {
			// 	return err
			// } else if err = yield(nctx, req); err != nil {
			// 	return err
			// }
		}
	}

	return nil
}

func (c *_Crawler) variationRequest(ctx context.Context, url string, referer string) ([]byte, error) {

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	opts := c.CrawlOptions(req.URL)

	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("referer", referer)
	req.Header.Add(`cookie`, `_gcl_au=1.1.1746811912.1630126168; _ga=GA1.2.125166996.1630126168; crl8.fpcuid=124f0070-d28b-4cc3-9a49-1969f3c8aa4a; _hjid=d732b178-1cc3-4e37-ac2c-d43afcc3363e; GlobalE_CT_Data={"CUID":"483184138.449817819.321","CHKCUID":null}; _scid=c1590e95-00a6-4f80-8883-5653adaea8e2; BVBRANDID=7dfb9cbb-d44b-418f-a2e7-0411f849ad6d; _gid=GA1.2.834711025.1631158687; GlobalE_Full_Redirect=false; token=undefined; GlobalE_SupportThirdPartCookies=true; _sctr=1|1631125800000; _aeaid=332b79ea-b641-4761-ba4d-284681aa9cf8; aeatstartmessage=true; GlobalE_Data={"countryISO":"US","currencyCode":"GBP","cultureCode":"en-US"}; stc114663=env:1631165241|20211010052721|20210909055721|1|1041477:20220909052721|uid:1630126171568.1906915484.1347685.114663.1499012269:20220909052721|srchist:1041477:1631165241:20211010052721:20220909052721|tsa:1631165241880.459461039.38041353.763137507379525.1:20210909055721; _hjAbsoluteSessionInProgress=0; ABTasty=uid=9n2sh6yb4npwqm27&fst=1630126168027&pst=1631249083004&cst=1631251699458&ns=15&pvt=68&pvis=1&th=650924.807921.54.1.9.1.1631158754013.1631251700168.1_745561.926508.21.8.5.1.1631158754136.1631170661662.1; ABTastySession=mrasn=&sen=2&lp=https%3A%2F%2Fwww.hunterboots.com%2Fus%2Fen_us%2Fmens-insulated-boots; _hjIncludedInSessionSample=0; ometria=2_cid=XaHecdX9PJSd3slP&nses=9&osts=1630126171&sid=2344ee93Gh9E4yXHMGKd&npv=1&tids=&slt=1631251701; stc113516=env:1631251701|20211119052821|20210910055821|1|1028364:20220910052821|uid:1631158756340.184967259.8643341.113516.1967005747.:20220910052821|srchist:1028364:1631251701:20211119052821:20220910052821|tsa:1631251701803.615200736.3742375.4468504780622091.4:20210910055821; skip_geocode=1`)

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

	ioutil.WriteFile("D:\\STS5\\New_VoilaCrawl\\VoilaCrawler\\Output_Product_js.html", respBody, 0644)
	return respBody, err
}

// NewTestRequest returns the custom test request which is used to monitor wheather the website struct is changed.
func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://www.hunterboots.com/us/en_us",
		//"https://www.hunterboots.com/us/en_us/womens-footwear-rainboots",
		// "https://www.hunterboots.com/us/en_us/womens-ankle-boots",
		"https://www.hunterboots.com/us/en_us/womens-ankle-boots/womens-original-chelsea-boots/yellow/6503",
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
