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

func (_ *_Crawler) New(_ *cli.Context, client http.Client, logger glog.Log) (crawler.Crawler, error) {
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
	return u.String(), nil
}

func (c *_Crawler) Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	} else if c.categoryAPIMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryAPIProducts(ctx, resp, yield)
	} else if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yield)
	}
	return crawler.ErrUnsupportedPath
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
			PrebuildID string `json:"prebuildId"`
			// MainPrebuild           string        `json:"mainPrebuild"`
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
		if item == nil {
			item = &pbItem.Product{
				Source: &pbItem.Source{
					Id:           p.ProductGroupID,
					CrawlUrl:     resp.Request.URL.String(),
					CanonicalUrl: canUrl,
				},
				BrandName:   p.Brand,
				Title:       p.FullTitle,
				Description: p.Description,
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
			Value: p.StyleColor,
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

		for _, rawSku := range p.Skus {
			originalPrice, _ := strconv.ParseFloat(p.FullPrice)
			msrp, _ := strconv.ParseFloat(p.FullPrice)
			discount := math.Ceil((msrp - originalPrice) / msrp * 100)

			sku := pbItem.Sku{
				SourceId:    rawSku.ID,
				Title:       p.FullTitle,
				Description: p.Description,
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
				Id:    rawSku.NikeSize,
				Name:  rawSku.LocalizedSize,
				Value: rawSku.LocalizedSize,
			})
			item.SkuItems = append(item.SkuItems, &sku)
		}
	}
	if item != nil {
		return yield(ctx, item)
	}
	return errors.New("no product found")
}

func (c *_Crawler) NewTestRequest(ctx context.Context) []*http.Request {
	var reqs []*http.Request
	for _, u := range []string{
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
	cli.NewApp(&_Crawler{}).Run(os.Args)
}
