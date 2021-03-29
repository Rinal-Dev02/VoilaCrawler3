package main

// this website exists api robot check.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/voiladev/VoilaCrawl/pkg/crawler"
	"github.com/voiladev/VoilaCrawl/pkg/net/http"
	"github.com/voiladev/VoilaCrawl/pkg/net/http/cookiejar"
	"github.com/voiladev/VoilaCrawl/pkg/proxy"
	pbMedia "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
)

type _Crawler struct {
	httpClient http.Client

	categoryPathMatcher     *regexp.Regexp
	categoryJsonPathMatcher *regexp.Regexp
	productPathMatcher      *regexp.Regexp
	productJsonPathMatcher  *regexp.Regexp
	logger                  glog.Log
}

func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient:          client,
		categoryPathMatcher: regexp.MustCompile(`^(/[a-z0-9_-]+)?/w([/a-z0-9_-]+)$`),
		productPathMatcher:  regexp.MustCompile(`^(/[a-z0-9_-]+)?/t([/a-z0-9_-a-zA-Z]+)$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "4b95dd02f3f535e5f2cc6254d64f56fe"
}

// Version
func (c *_Crawler) Version() int32 {
	return 1
}

// CrawlOptions
func (c *_Crawler) CrawlOptions() *crawler.CrawlOptions {
	options := crawler.NewCrawlOptions()
	options.EnableHeadless = false
	options.LoginRequired = false
	options.Reliability = 1
	// NOTE: no need to set useragent here for user agent is dynamic
	// options.MustHeader.Set("Accept-Language", "en-US,en;q=0.8")
	// options.MustHeader.Set("X-Requested-With", "XMLHttpRequest")

	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.nike.com"}
}

func (c *_Crawler) Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	} else if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yield)
	}
	return fmt.Errorf("unsupported url %s", resp.Request.URL.String())
}

func isRobotCheckPage(respBody []byte) bool {
	return bytes.Contains(respBody, []byte("we believe you are using automation tools to browse the website")) ||
		bytes.Contains(respBody, []byte("Javascript is disabled or blocked by an extension")) ||
		bytes.Contains(respBody, []byte("Your browser does not support cookies"))
}

type CategoryView struct {
	Wall struct {
		PageData struct {
			Prev           string      `json:"prev"`
			Next           string      `json:"next"`
			TotalPages     int         `json:"totalPages"`
			TotalResources int         `json:"totalResources"`
			SearchSummary  interface{} `json:"searchSummary"`
		} `json:"pageData"`
		Products []struct {
			AltImages        interface{} `json:"altImages"`
			CardType         string      `json:"cardType"`
			CloudProductID   string      `json:"cloudProductId"`
			ColorDescription string      `json:"colorDescription"`
			Colorways        []struct {
				ColorDescription string `json:"colorDescription"`
				Images           struct {
					PortraitURL string `json:"portraitURL"`
					SquarishURL string `json:"squarishURL"`
				} `json:"images"`
				PdpURL string `json:"pdpUrl"`
				Price  struct {
					Currency               string      `json:"currency"`
					CurrentPrice           float64     `json:"currentPrice"`
					Discounted             bool        `json:"discounted"`
					EmployeePrice          float64     `json:"employeePrice"`
					FullPrice              float64     `json:"fullPrice"`
					MinimumAdvertisedPrice interface{} `json:"minimumAdvertisedPrice"`
				} `json:"price"`
				AltImages         interface{} `json:"altImages"`
				CloudProductID    string      `json:"cloudProductId"`
				InStock           bool        `json:"inStock"`
				IsExcluded        bool        `json:"isExcluded"`
				IsMemberExclusive bool        `json:"isMemberExclusive"`
				IsNew             bool        `json:"isNew"`
				Label             string      `json:"label"`
				Pid               string      `json:"pid"`
				PrebuildID        interface{} `json:"prebuildId"`
				ProductInstanceID interface{} `json:"productInstanceId"`
			} `json:"colorways"`
			Customizable      bool   `json:"customizable"`
			HasExtendedSizing bool   `json:"hasExtendedSizing"`
			ID                string `json:"id"`
			Images            struct {
				PortraitURL string `json:"portraitURL"`
				SquarishURL string `json:"squarishURL"`
			} `json:"images"`
			InStock           bool        `json:"inStock"`
			IsExcluded        bool        `json:"isExcluded"`
			IsJersey          bool        `json:"isJersey"`
			IsMemberExclusive bool        `json:"isMemberExclusive"`
			IsNBA             bool        `json:"isNBA"`
			IsNFL             bool        `json:"isNFL"`
			IsSustainable     bool        `json:"isSustainable"`
			Label             string      `json:"label"`
			NbyColorway       interface{} `json:"nbyColorway"`
			Pid               string      `json:"pid"`
			PrebuildID        interface{} `json:"prebuildId"`
			Price             struct {
				Currency               string      `json:"currency"`
				CurrentPrice           float64     `json:"currentPrice"`
				Discounted             bool        `json:"discounted"`
				EmployeePrice          float64     `json:"employeePrice"`
				FullPrice              float64     `json:"fullPrice"`
				MinimumAdvertisedPrice interface{} `json:"minimumAdvertisedPrice"`
			} `json:"price"`
			PriceRangeCurrent  string      `json:"priceRangeCurrent"`
			PriceRangeEmployee string      `json:"priceRangeEmployee"`
			PriceRangeFull     string      `json:"priceRangeFull"`
			ProductInstanceID  interface{} `json:"productInstanceId"`
			ProductType        string      `json:"productType"`
			Properties         interface{} `json:"properties"`
			SalesChannel       []string    `json:"salesChannel"`
			Subtitle           string      `json:"subtitle"`
			Title              string      `json:"title"`
			URL                string      `json:"url"`
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

		rawurl := fmt.Sprintf("%s://%s%s", resp.Request.URL.Scheme, resp.Request.URL.Host, strings.ReplaceAll(idv.URL, "{countryLang}", ""))

		fmt.Println(rawurl)
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

	// check if this is the last page
	if lastIndex >= viewData.Wall.PageData.TotalResources {
		return nil
	}

	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}

	// set pagination
	u := *resp.Request.URL
	vals := u.Query()
	vals.Set("page", strconv.Format(page+1))
	vals.Set("count", strconv.Format(48))
	u.RawQuery = vals.Encode()

	fmt.Println(u)
	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	// update the index of last page
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
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
			ID            string `json:"id"`
			ThreadID      string `json:"threadId"`
			ProductID     string `json:"productId"`
			MainColor     bool   `json:"mainColor"`
			ProductRollup struct {
				Type string `json:"type"`
				Key  string `json:"key"`
			} `json:"productRollup"`
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
			NbyContentCopy           struct {
			} `json:"nbyContentCopy"`
			PrebuildID             string        `json:"prebuildId"`
			MainPrebuild           string        `json:"mainPrebuild"`
			IsNikeID               bool          `json:"isNikeID"`
			IsNBYDesign            bool          `json:"isNBYDesign"`
			UpdatedNBYDesignKey    string        `json:"updatedNBYDesignKey"`
			Piid                   string        `json:"piid"`
			PathName               string        `json:"pathName"`
			Vas                    []interface{} `json:"vas"`
			SeoProductDescription  string        `json:"seoProductDescription"`
			SeoProductAvailability bool          `json:"seoProductAvailability"`
			SeoProductReleaseDate  time.Time     `json:"seoProductReleaseDate"`
			Discounted             bool          `json:"discounted"`
			FullPrice              float64       `json:"fullPrice"`
			CurrentPrice           float64       `json:"currentPrice"`
			EmployeePrice          float64       `json:"employeePrice"`
			Currency               string        `json:"currency"`
			Skus                   []struct {
				ID                  string `json:"id"`
				NikeSize            string `json:"nikeSize"`
				SkuID               string `json:"skuId"`
				LocalizedSize       string `json:"localizedSize"`
				LocalizedSizePrefix string `json:"localizedSizePrefix"`
			} `json:"skus"`
			Title string `json:"title"`
			Nodes []struct {
				Analytics struct {
					HashKey string `json:"hashKey"`
				} `json:"analytics"`
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
				SubType    string `json:"subType"`
				ID         string `json:"id"`
				Type       string `json:"type"`
				Version    string `json:"version"`
				Properties struct {
					ContainerType string `json:"containerType"`
					Loop          bool   `json:"loop"`
					Subtitle      string `json:"subtitle"`
					ColorTheme    string `json:"colorTheme"`
					AutoPlay      bool   `json:"autoPlay"`
					Title         string `json:"title"`
					Body          string `json:"body"`
					Speed         int    `json:"speed"`
				} `json:"properties"`
			} `json:"nodes"`
			LayoutCards           []interface{} `json:"layoutCards"`
			FirstImageURL         string        `json:"firstImageUrl"`
			FirstImageAltText     string        `json:"firstImageAltText"`
			SustainabilityMessage []interface{} `json:"sustainabilityMessage"`
			Language              string        `json:"language"`
			Marketplace           string        `json:"marketplace"`
			AvailableSkus         []struct {
				ID           string `json:"id"`
				ProductID    string `json:"productId"`
				ResourceType string `json:"resourceType"`
				Links        struct {
					Self struct {
						Ref string `json:"ref"`
					} `json:"self"`
				} `json:"links"`
				Available bool   `json:"available"`
				Level     string `json:"level"`
				SkuID     string `json:"skuId"`
			} `json:"availableSkus"`
			PublishType       string      `json:"publishType"`
			IsLaunchView      bool        `json:"isLaunchView"`
			LaunchView        interface{} `json:"launchView"`
			CollectionTermIds []string    `json:"collectionTermIds"`
			ConceptIds        []string    `json:"conceptIds"`
			SizeChartURL      string      `json:"sizeChartUrl"`
			CatalogID         string      `json:"catalogId"`
			SelectedSkuID     struct {
			} `json:"selectedSkuId"`
			SizeAndFitDescription         string `json:"sizeAndFitDescription"`
			ExclusiveAccess               bool   `json:"exclusiveAccess"`
			NotifyMeIndicator             bool   `json:"notifyMeIndicator"`
			PromoExclusionAccess          bool   `json:"promoExclusionAccess"`
			UseFeedsPromoExclusionMessage bool   `json:"useFeedsPromoExclusionMessage"`
			JerseyIDPathName              string `json:"jerseyIdPathName"`
			State                         string `json:"state"`
			//	} `json:"CV4791-100"`

		} `json:"products"`
	} `json:"Threads"`
}

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

	for _, p := range viewData.Threads.Products {
		// build product data
		item := pbItem.Product{
			Source: &pbItem.Source{
				Id:       strconv.Format(p.ID),
				CrawlUrl: resp.Request.URL.String(),
			},
			BrandName:   p.Brand,
			Title:       p.FullTitle,
			Description: p.Description,
			CrowdType:   p.Genders[0],
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
			},
			Stats: &pbItem.Stats{
				ReviewCount: int32(viewData.Reviews.Total),
				Rating:      float32(viewData.Reviews.AverageRating),
			},
		}
		for ks, rawSku := range p.Skus {
			originalPrice, _ := strconv.ParseFloat(p.FullPrice)
			msrp, _ := strconv.ParseFloat(p.FullPrice)
			discount := (msrp - originalPrice) / msrp * 100

			sku := pbItem.Sku{
				SourceId: strconv.Format(rawSku.ID),
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(originalPrice * 100),
					Msrp:     int32(msrp * 100),
					Discount: int32(discount),
				},
				Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}
			if p.State == "IN_STOCK" {
				sku.Stock.StockStatus = pbItem.Stock_InStock
				//sku.Stock.StockCount = int32(rawSku.TotalQuantityAvailable)
			}

			// color
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    strconv.Format(p.Pid),
				Name:  p.ColorDescription,
				Value: p.ColorDescription,
				//Icon:  color.SwatchMedia.Mobile,
			})

			if ks == 0 {

				isDefault := true
				for ki, m := range p.Nodes[0].Nodes {
					template := strings.ReplaceAll(m.Properties.Squarish.URL, "t_default", "t_PDP_864_v1")
					if ki > 0 {
						isDefault = false
					}

					sku.Medias = append(sku.Medias, pbMedia.NewImageMedia(
						strconv.Format(m.ID),
						template,
						template,
						template,
						template,
						"",
						isDefault,
					))
				}
			}
			// size
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    rawSku.ID,
				Name:  rawSku.LocalizedSize,
				Value: rawSku.LocalizedSize,
			})

			item.SkuItems = append(item.SkuItems, &sku)
		}

		// yield item result
		if err = yield(ctx, &item); err != nil {
			return err
		}
	}

	return nil
}

func (c *_Crawler) NewTestRequest(ctx context.Context) []*http.Request {
	var reqs []*http.Request
	for _, u := range []string{
		"https://www.nike.com/w/womens-running-shoes-37v7jz5e1x6zy7ok",
		"https://www.nike.com/in/t/react-escape-run-running-shoe-94nDwX/CV3817-003",
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

// local test
func main() {
	logger := glog.New(glog.LogLevelDebug)
	// build a http client
	// get proxy's microservice address from env
	os.Setenv("VOILA_PROXY_URL", "http://3.239.93.53:30216")
	client, err := proxy.NewProxyClient(os.Getenv("VOILA_PROXY_URL"), cookiejar.New(), logger)
	if err != nil {
		panic(err)
	}

	// instance the spider locally
	spider, err := New(client, logger)
	if err != nil {
		panic(err)
	}
	opts := spider.CrawlOptions()

	// this callback func is used to do recursion call of sub requests.
	var callback func(ctx context.Context, val interface{}) error
	callback = func(ctx context.Context, val interface{}) error {
		switch i := val.(type) {
		case *http.Request:
			logger.Debugf("Access %s", i.URL)

			// process logic of sub request

			// init custom headers
			for k := range opts.MustHeader {
				i.Header.Set(k, opts.MustHeader.Get(k))
			}

			// init custom cookies
			for _, c := range opts.MustCookies {
				if strings.HasPrefix(i.URL.Path, c.Path) || c.Path == "" {
					val := fmt.Sprintf("%s=%s", c.Name, c.Value)
					if c := i.Header.Get("Cookie"); c != "" {
						i.Header.Set("Cookie", c+"; "+val)
					} else {
						i.Header.Set("Cookie", val)
					}
				}
			}

			// set scheme,host for sub requests. for the product url in category page is just the path without hosts info.
			// here is just the test logic. when run the spider online, the controller will process automatically
			if i.URL.Scheme == "" {
				i.URL.Scheme = "https"
			}
			if i.URL.Host == "" {
				i.URL.Host = "www.nike.com"
			}

			// do http requests here.
			nctx, cancel := context.WithTimeout(ctx, time.Minute*5)
			defer cancel()
			resp, err := client.DoWithOptions(nctx, i, http.Options{
				EnableProxy:       true,
				EnableHeadless:    true,
				EnableSessionInit: spider.CrawlOptions().EnableSessionInit,
				KeepSession:       spider.CrawlOptions().KeepSession,
				Reliability:       spider.CrawlOptions().Reliability,
			})
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			return spider.Parse(ctx, resp, callback)
		default:
			// output the result
			data, err := json.Marshal(i)
			if err != nil {
				return err
			}
			logger.Infof("data: %s", data)
		}
		return nil
	}

	ctx := context.WithValue(context.Background(), "tracing_id", fmt.Sprintf("tracing_%d", time.Now().UnixNano()))
	// start the crawl request
	for _, req := range spider.NewTestRequest(context.Background()) {
		if err := callback(ctx, req); err != nil {
			logger.Fatal(err)
		}
	}
}
