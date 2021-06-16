package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	uuid "github.com/satori/go.uuid"
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

type _Crawler struct {
	httpClient http.Client

	categoryPathMatcher *regexp.Regexp
	categoryAPIMatcher  *regexp.Regexp
	productPathMatcher  *regexp.Regexp
	logger              glog.Log
}

func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient:          client,
		categoryPathMatcher: regexp.MustCompile(`^/w(/[a-z0-9_\-]+){1,6}$`),
		categoryAPIMatcher:  regexp.MustCompile(`^/cic/browse/v\d+$`),
		productPathMatcher:  regexp.MustCompile(`^/(u|t)(/[a-zA-Z0-9_\-]+){2,4}$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "67eb6ce2793510f21559225e8648c1c7"
}

// Version
func (c *_Crawler) Version() int32 {
	return 1
}

// CrawlOptions
func (c *_Crawler) CrawlOptions(u *url.URL) *crawler.CrawlOptions {
	options := crawler.NewCrawlOptions()
	options.EnableHeadless = false
	options.Reliability = pbProxy.ProxyReliability_ReliabilityMedium
	options.MustCookies = append(options.MustCookies,
		&http.Cookie{Name: "NIKE_COMMERCE_COUNTRY", Value: "US"},
		&http.Cookie{Name: "NIKE_COMMERCE_LANG_LOCALE", Value: "en_US"},
		&http.Cookie{Name: "MR", Value: "0"},
		&http.Cookie{Name: "IR_gbd", Value: "nike.com"},
		&http.Cookie{Name: "geoloc", Value: "cc=US,rc=CA,tp=vhigh,tz=PST,la=33.9733,lo=-118.2487"},
	)
	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.nike.com"}
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
		u.Host = "www.nike.com"
	}
	if c.productPathMatcher.MatchString(u.Path) {
		u.RawQuery = ""
		return u.String(), nil
	}
	return rawurl, nil
}

func (c *_Crawler) Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}
	p := strings.TrimSuffix(resp.RawUrl().Path, "/")

	if p == "" {
		return c.parseCategories(ctx, resp, yield)
	}
	if c.categoryPathMatcher.MatchString(resp.RawUrl().Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	} else if c.categoryAPIMatcher.MatchString(resp.RawUrl().Path) {
		return c.parseCategoryAPIProducts(ctx, resp, yield)
	} else if c.productPathMatcher.MatchString(resp.RawUrl().Path) {
		return c.parseProduct(ctx, resp, yield)
	}
	return crawler.ErrUnsupportedPath
}

func (c *_Crawler) parseCategories(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	dom, err := resp.Selector()
	if err != nil {
		c.logger.Error(err)
		return err
	}

	sel := dom.Find(`.pre-desktop-menu>li`)

	for i := range sel.Nodes {
		node := sel.Eq(i)
		cateName := strings.TrimSpace(node.Find(`a`).First().Text())
		if cateName == "" {
			continue
		}
		nnctx := context.WithValue(ctx, "Category", cateName)

		subSel := node.Find(`.pre-menu-column`)
		for k := range subSel.Nodes {
			subNode2 := subSel.Eq(k)
			subcat2 := subNode2.Find(`button`).Text()
			if subcat2 == "" {
				subcat2 = subNode2.Find(`a`).First().Text()
			}
			subNode2list := subNode2.Find(`a`)

			for j := range subNode2list.Nodes {
				subNode := subNode2list.Eq(j)
				href := subNode.AttrOr("href", "")
				if href == "" {
					continue
				}

				_, err := url.Parse(href)
				if err != nil {
					c.logger.Error("parse url %s failed", href)
					continue
				}

				subCateName := subcat2 + " > " + strings.TrimSpace(subNode.Text())

				nnnctx := context.WithValue(nnctx, "SubCategory", subCateName)
				req, _ := http.NewRequest(http.MethodGet, href, nil)
				if err := yield(nnnctx, req); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func isRobotCheckPage(respBody []byte) bool {
	return bytes.Contains(respBody, []byte("we believe you are using automation tools to browse the website")) ||
		bytes.Contains(respBody, []byte("Javascript is disabled or blocked by an extension")) ||
		bytes.Contains(respBody, []byte("Your browser does not support cookies"))
}

type CategoryView struct {
	Wall struct {
		PageData struct {
			Next string `json:"next"`
		} `json:"pageData"`
		Products []struct {
			URL string `json:"url"`
		} `json:"products"`
	} `json:"Wall"`
}

var productsExtractReg = regexp.MustCompile(`window\.INITIAL_REDUX_STATE=({.*});?\s*</script>`)

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

	matched := productsExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}
	// c.logger.Debugf("data: %s", matched[1])

	var viewData CategoryView
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal data fetched from %s failed, error=%s", resp.Request.URL, err)
		return err
	}

	lastIndex := nextIndex(ctx)
	for _, idv := range viewData.Wall.Products {
		if idv.URL == "" {
			continue
		}
		rawurl := strings.ReplaceAll(idv.URL, "{countryLang}", "")
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

	if viewData.Wall.PageData.Next != "" {
		var anonymousId string
		cookies, _ := c.httpClient.Jar().Cookies(ctx, resp.Request.URL)
		for _, cookie := range cookies {
			if cookie.Name == "anonymousId" {
				anonymousId = cookie.Value
				break
			}
		}

		if anonymousId == "" {
			anonymousId = fmt.Sprintf("%X", uuid.NewV4().Bytes())
		}

		u, _ := url.Parse("https://api.nike.com/cic/browse/v1")
		vals := u.Query()
		vals.Set("queryid", "products")
		vals.Set("anonymousId", anonymousId)
		vals.Set("country", "us")
		vals.Set("endpoint", viewData.Wall.PageData.Next)
		vals.Set("language", "en")
		vals.Set("localizedRangeStr", "{lowestPrice} — {highestPrice}")
		u.RawQuery = vals.Encode()

		req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
		// update the index of last page
		nctx := context.WithValue(ctx, "item.index", lastIndex)
		return yield(nctx, req)
	}
	return nil
}

func (c *_Crawler) parseCategoryAPIProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil {
		return nil
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var viewData struct {
		Data struct {
			Products struct {
				Products []struct {
					URL string `json:"url"`
				} `json:"products"`
				Pages struct {
					Prev           string      `json:"prev"`
					Next           string      `json:"next"`
					TotalPages     int         `json:"totalPages"`
					TotalResources int         `json:"totalResources"`
					SearchSummary  interface{} `json:"searchSummary"`
				} `json:"pages"`
				ExternalResponses interface{} `json:"externalResponses"`
				TraceID           string      `json:"traceId"`
			} `json:"products"`
		} `json:"data"`
	}

	if err := json.Unmarshal([]byte(respBody), &viewData); err != nil {
		c.logger.Error(err)
		return err
	}

	lastIndex := nextIndex(ctx)
	for _, prod := range viewData.Data.Products.Products {
		u := "https://www.nike.com" + strings.ReplaceAll(prod.URL, "{countryLang}", "")

		req, err := http.NewRequest(http.MethodGet, u, nil)
		if err != nil {
			c.logger.Error(err)
			return err
		}
		nctx := context.WithValue(ctx, "item.index", lastIndex)
		lastIndex += 1

		if err := yield(nctx, req); err != nil {
			return err
		}
	}

	if viewData.Data.Products.Pages.Next != "" {
		u := *resp.Request.URL
		vals := u.Query()
		vals.Set("endpoint", viewData.Data.Products.Pages.Next)

		u.RawQuery = vals.Encode()

		req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
		// update the index of last page
		nctx := context.WithValue(ctx, "item.index", lastIndex)
		return yield(nctx, req)
	}

	return nil
}

// nextIndex used to get sharingData from context
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")) + 1)
}

type parseProductResponse struct {
	Reviews struct {
		Country          string  `json:"country"`
		ProductStyleCode string  `json:"productStyleCode"`
		Total            int     `json:"total"`
		AverageRating    float64 `json:"averageRating"`
		RecommendedCount int     `json:"recommendedCount"`
	} `json:"reviews"`
	Threads struct {
		Products map[string]struct {
			//CV4791100 struct {
			ID                       string        `json:"id"`
			ThreadID                 string        `json:"threadId"`
			ProductID                string        `json:"productId"`
			MainColor                bool          `json:"mainColor"`
			Width                    interface{}   `json:"width"`
			Athletes                 []interface{} `json:"athletes"`
			StyleColor               string        `json:"styleColor"`
			StyleCode                string        `json:"styleCode"`
			Category                 string        `json:"category"`
			Pid                      string        `json:"pid"`
			PreOrder                 bool          `json:"preOrder"`
			PreorderAvailabilityDate string        `json:"preorderAvailabilityDate"`
			CommerceStartDate        time.Time     `json:"commerceStartDate"`
			IsNYBCountDown           bool          `json:"isNYBCountDown"`
			Countdown                int           `json:"countdown"`
			HardLaunch               bool          `json:"hardLaunch"`
			SportTags                []interface{} `json:"sportTags"`
			Genders                  []string      `json:"genders"`
			Brand                    string        `json:"brand"`
			ProductType              string        `json:"productType"`
			NikeIDStyleCode          string        `json:"nikeIdStyleCode"`
			ProductGroupID           string        `json:"productGroupId"`
			StyleType                string        `json:"styleType"`
			Status                   string        `json:"status"`
			DisplayOrder             string        `json:"displayOrder"`
			SubTitle                 string        `json:"subTitle"`
			FullTitle                string        `json:"fullTitle"`
			LangLocale               string        `json:"langLocale"`
			Marketing                string        `json:"marketing"`
			Origin                   []string      `json:"origin"`
			ColorDescription         string        `json:"colorDescription"`
			Description              string        `json:"description"`
			DescriptionPreview       string        `json:"descriptionPreview"`
			DescriptionHeading       string        `json:"descriptionHeading"`
			PrebuildID               string        `json:"prebuildId"`
			// MainPrebuild           string        `json:"mainPrebuild"`
			IsNikeID            bool          `json:"isNikeID"`
			IsNBYDesign         bool          `json:"isNBYDesign"`
			UpdatedNBYDesignKey string        `json:"updatedNBYDesignKey"`
			Piid                string        `json:"piid"`
			PathName            string        `json:"pathName"`
			Vas                 []interface{} `json:"vas"`
			Discounted          bool          `json:"discounted"`
			FullPrice           float64       `json:"fullPrice"`
			CurrentPrice        float64       `json:"currentPrice"`
			EmployeePrice       float64       `json:"employeePrice"`
			Currency            string        `json:"currency"`
			Skus                []struct {
				ID                  string `json:"id"`
				NikeSize            string `json:"nikeSize"`
				SkuID               string `json:"skuId"`
				LocalizedSize       string `json:"localizedSize"`
				LocalizedSizePrefix string `json:"localizedSizePrefix"`
			} `json:"skus"`
			Title string `json:"title"`
			Nodes []struct {
				Nodes []struct {
					Analytics struct {
						HashKey string `json:"hashKey"`
					} `json:"analytics"`
					SubType    string `json:"subType"`
					ID         string `json:"id"`
					Type       string `json:"type"`
					Version    string `json:"version"`
					Properties struct {
						PortraitID   string `json:"portraitId"`
						SquarishURL  string `json:"squarishURL"`
						LandscapeID  string `json:"landscapeId"`
						AltText      string `json:"altText"`
						PortraitURL  string `json:"portraitURL"`
						LandscapeURL string `json:"landscapeURL"`
						Title        string `json:"title"`
						Portrait     struct {
							ID   string `json:"id"`
							Type string `json:"type"`
							URL  string `json:"url"`
						} `json:"portrait"`
						Squarish struct {
							ID   string `json:"id"`
							Type string `json:"type"`
							URL  string `json:"url"`
						} `json:"squarish"`
						SeoName    string `json:"seoName"`
						SquarishID string `json:"squarishId"`
						Subtitle   string `json:"subtitle"`
						ColorTheme string `json:"colorTheme"`
					} `json:"properties"`
				} `json:"nodes"`
			} `json:"nodes"`
			State string `json:"state"`
			//	} `json:"CV4791-100"`
		} `json:"products"`
	} `json:"Threads"`
}

// used to trim html labels in description
var htmlTrimRegp = regexp.MustCompile(`</?[^>]+>`)

func (c *_Crawler) parseProduct(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}
	if resp.StatusCode == http.StatusForbidden {
		return errors.New("access denied")
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if isRobotCheckPage(respBody) {
		return errors.New("robot check page")
	}

	matched := productsExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}
	// c.logger.Debugf("data: %s", matched[1])

	var viewData parseProductResponse
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}
	canUrl := dom.Find(`link[rel="canonical"]`).AttrOr("href", "")
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}

	var item *pbItem.Product
	for _, p := range viewData.Threads.Products {
		desc := p.Description

		enc := json.NewEncoder(os.Stdout)
		enc.SetEscapeHTML(false)
		enc.Encode(desc)

		if item == nil {
			item = &pbItem.Product{
				Source: &pbItem.Source{
					Id:           p.ProductID,
					CrawlUrl:     resp.Request.URL.String(),
					CanonicalUrl: canUrl,
				},
				BrandName:   p.Brand,
				Title:       p.FullTitle,
				Description: htmlTrimRegp.ReplaceAllString(desc, ""),
				CrowdType:   strings.ToLower(strings.Join(p.Genders, ",")),
				Category:    p.Category,
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
				},
				Stats: &pbItem.Stats{
					ReviewCount: int32(viewData.Reviews.Total),
					Rating:      float32(viewData.Reviews.AverageRating),
				},
			}
		}

		colorSpec := pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecColor,
			Id:    p.StyleColor,
			Name:  p.ColorDescription,
			Value: p.ColorDescription,
		}

		var medias []*pbMedia.Media
		for ki, m := range p.Nodes[0].Nodes {
			template := strings.ReplaceAll(m.Properties.Squarish.URL, "t_default", "t_PDP_864_v1")
			medias = append(medias, pbMedia.NewImageMedia(
				strconv.Format(m.ID),
				strings.ReplaceAll(m.Properties.Squarish.URL, "t_default", "t_PDP_1280_v1"),
				strings.ReplaceAll(m.Properties.Squarish.URL, "t_default", "t_PDP_1280_v1"),
				template,
				template,
				"",
				ki == 0,
			))
		}
		item.Medias = medias

		for k, rawSku := range p.Skus {
			originalPrice, _ := strconv.ParseFloat(p.FullPrice)
			msrp, _ := strconv.ParseFloat(p.FullPrice)
			discount := math.Ceil((msrp - originalPrice) / msrp * 100)

			sku := pbItem.Sku{
				SourceId: fmt.Sprintf("%s_%v", rawSku.ID, k),
				Title:    p.FullTitle,
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(originalPrice * 100),
					Msrp:     int32(msrp * 100),
					Discount: int32(discount),
				},
				Medias: medias,
				Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}
			if p.State == "IN_STOCK" {
				sku.Stock.StockStatus = pbItem.Stock_InStock
			}
			sku.Specs = append(sku.Specs, &colorSpec)

			// size
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    strconv.Format(k),
				Name:  rawSku.LocalizedSize,
				Value: rawSku.LocalizedSize,
			})
			item.SkuItems = append(item.SkuItems, &sku)
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

	if item != nil {
		return yield(ctx, item)
	}
	return errors.New("no product found")
}

func (c *_Crawler) NewTestRequest(ctx context.Context) []*http.Request {
	var reqs []*http.Request
	for _, u := range []string{
		//"https://www.nike.com",
		// "https://www.nike.com/w/womens-running-shoes-37v7jz5e1x6zy7ok",
		// "https://www.nike.com/t/react-escape-run-running-shoe-94nDwX/CV3817-003",
		"https://www.nike.com/u/custom-nike-air-zoom-tempo-next-by-you-10000953/8909442240",
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
	cli.NewApp(New).Run(os.Args)
}
