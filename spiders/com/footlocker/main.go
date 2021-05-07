package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/voiladev/go-crawler/pkg/cli"
	"github.com/voiladev/go-crawler/pkg/crawler"
	"github.com/voiladev/go-crawler/pkg/net/http"
	"github.com/voiladev/go-crawler/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/go-crawler/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/go-crawler/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	pbProxy "github.com/voiladev/go-crawler/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
	"google.golang.org/protobuf/types/known/anypb"
)

type _Crawler struct {
	httpClient http.Client

	categoryPathMatcher *regexp.Regexp
	productPathMatcher  *regexp.Regexp
	imagePathMatcher    *regexp.Regexp
	logger              glog.Log
}

func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient:          client,
		categoryPathMatcher: regexp.MustCompile(`^(/[a-z0-9_-]+)?/category/(women|men)([/a-z0-9_-]+).html$`),
		productPathMatcher:  regexp.MustCompile(`^/product(/[~a-zA-Z0-9\-]+)+.html$`),
		imagePathMatcher:    regexp.MustCompile(`^(/[is/image/EBFL2/]+)(/[a-zA-Z0-9_-]+)([/?req=set,json&id=]+([A-Za-z0-9]+))$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "0c38d862d28ce09a51c3364ff14de43a"
}

// Version
func (c *_Crawler) Version() int32 {
	return 1
}

// CrawlOptions
func (c *_Crawler) CrawlOptions(u *url.URL) *crawler.CrawlOptions {
	options := crawler.NewCrawlOptions()
	options.EnableHeadless = false
	options.EnableSessionInit = false
	options.Reliability = pbProxy.ProxyReliability_ReliabilityMedium
	options.MustCookies = append(options.MustCookies,
		&http.Cookie{Name: "at_check", Value: "false"},
		&http.Cookie{Name: "s_sq", Value: "%5B%5BB%5D%5D"},
	)
	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.footlocker.com"}
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
		u.Host = "www.footlocker.com"
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

	if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	} else if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yield)
		// } else if c.imagePathMatcher.MatchString(resp.Request.URL.Path) {
		// 	return c.parseProduct(ctx, resp, yield)
	}
	return crawler.ErrUnsupportedPath
}

// nextIndex used to get sharingData from context
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")) + 1)
}

var prodDataExtraReg = regexp.MustCompile(`(window.digitalData)\s*=\s*({.*});\s*`)

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

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	lastIndex := nextIndex(ctx)

	sel := dom.Find(`.product-container`)
	for i := range sel.Nodes {
		node := sel.Eq(i)
		u := node.Find(`.ProductCard .ProductCard-link`).AttrOr("href", "")
		if u == "" {
			continue
		}
		req, err := http.NewRequest(http.MethodGet, u, nil)
		if err != nil {
			c.logger.Errorf("invlaud request url %s", u)
			continue
		}
		nctx := context.WithValue(ctx, "item.index", lastIndex)
		lastIndex += 1
		if err := yield(nctx, req); err != nil {
			return err
		}
	}
	if len(sel.Nodes) == 0 {
		c.logger.Errorf("no product found %s", respBody)
		return fmt.Errorf("no product found")
	}

	// get current page number
	href := dom.Find(`.Pagination-option--next>a[aria-label="Go to last page"]`).AttrOr("href", "")
	if href == "" {
		return nil
	}
	finalUrl, err := url.Parse(href)
	if err != nil {
		c.logger.Errorf("got invalud exit url %s", finalUrl)
		return err
	}
	nextPage, _ := strconv.ParseInt(finalUrl.Query().Get("currentPage"))
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("currentPage"))
	if page >= nextPage {
		return nil
	}

	// set pagination
	u := *resp.Request.URL
	vals := u.Query()
	vals.Set("currentPage", strconv.Format(page+1))
	u.RawQuery = vals.Encode()

	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	nctx := context.WithValue(ctx, "item.index", lastIndex)
	return yield(nctx, req)
}

type parseProductResponse struct {
	Details struct {
		Data map[string][]struct {
			Code                           string   `json:"code"`
			DisplayCountDownTimer          bool     `json:"displayCountDownTimer"`
			EligiblePaymentTypesForProduct []string `json:"eligiblePaymentTypesForProduct"`
			FitVariant                     string   `json:"fitVariant"`
			FreeShipping                   bool     `json:"freeShipping"`
			FreeShippingMessage            string   `json:"freeShippingMessage"`
			IsSelected                     bool     `json:"isSelected"`
			LaunchProduct                  bool     `json:"launchProduct"`
			MapEnable                      bool     `json:"mapEnable"`
			Price                          struct {
				CurrencyIso            string  `json:"currencyIso"`
				FormattedOriginalPrice string  `json:"formattedOriginalPrice"`
				FormattedValue         string  `json:"formattedValue"`
				OriginalPrice          float64 `json:"originalPrice"`
				Value                  float64 `json:"value"`
			} `json:"price"`
			RecaptchaOn               bool   `json:"recaptchaOn"`
			Riskified                 bool   `json:"riskified"`
			ShipToAndFromStore        bool   `json:"shipToAndFromStore"`
			ShippingRestrictionExists bool   `json:"shippingRestrictionExists"`
			Sku                       string `json:"sku"`
			SkuExclusions             bool   `json:"skuExclusions"`
			StockLevelStatus          string `json:"stockLevelStatus"`
			WebOnlyLaunch             bool   `json:"webOnlyLaunch"`
			Width                     string `json:"width"`
			Products                  []struct {
				Attributes []struct {
					ID    string `json:"id"`
					Type  string `json:"type"`
					Value string `json:"value"`
				} `json:"attributes"`
				BarCode         string `json:"barCode"`
				Code            string `json:"code"`
				IsBackOrderable bool   `json:"isBackOrderable"`
				IsPreOrder      bool   `json:"isPreOrder"`
				IsRecaptchaOn   bool   `json:"isRecaptchaOn"`
				Price           struct {
					CurrencyIso            string  `json:"currencyIso"`
					FormattedOriginalPrice string  `json:"formattedOriginalPrice"`
					FormattedValue         string  `json:"formattedValue"`
					OriginalPrice          float64 `json:"originalPrice"`
					Value                  float64 `json:"value"`
				} `json:"price"`
				SingleStoreInventory         bool   `json:"singleStoreInventory"`
				SizeAvailableInStores        bool   `json:"sizeAvailableInStores"`
				SizeAvailableInStoresMessage string `json:"sizeAvailableInStoresMessage,omitempty"`
				StockLevelStatus             string `json:"stockLevelStatus"`
				Style                        struct {
					ID    string `json:"id"`
					Type  string `json:"type"`
					Value string `json:"value"`
				} `json:"style"`
				Size struct {
					ID    string `json:"id"`
					Type  string `json:"type"`
					Value string `json:"value"`
				} `json:"size"`
			} `json:"products"`
			Style string `json:"style"`
		} `json:"data"`
		Product map[string]struct {
			Name       string `json:"name"`
			Brand      string `json:"brand"`
			Categories []struct {
				Code string `json:"code"`
				Name string `json:"name"`
			} `json:"categories"`
			IsGiftCard      bool   `json:"isGiftCard"`
			Description     string `json:"description"`
			ModelNumber     string `json:"modelNumber"`
			IsNewProduct    bool   `json:"isNewProduct"`
			IsSaleProduct   bool   `json:"isSaleProduct"`
			IsEmailGiftCard bool   `json:"isEmailGiftCard"`
			SizeChart       []struct {
				Label string   `json:"label"`
				Sizes []string `json:"sizes"`
			} `json:"sizeChart"`
			SizeMessage string `json:"sizeMessage"`
		} `json:"product"`
		Reviews struct {
			P566155CHTML struct {
				Results []struct {
					Rating        int    `json:"rating"`
					ReviewBody    string `json:"reviewBody"`
					Name          string `json:"name"`
					Author        string `json:"author"`
					DatePublished string `json:"datePublished"`
				} `json:"results"`
				Details struct {
					ReviewCount int     `json:"reviewCount"`
					RatingValue float64 `json:"ratingValue"`
					BestRating  float64 `json:"bestRating"`
					WorstRating float64 `json:"worstRating"`
				} `json:"details"`
			} `json:"/product/converse-all-star-lugged-hi-womens/566155C.html"`
		} `json:"reviews"`
		Selected struct {
			P566155CHTML struct {
				Price struct {
					CurrencyIso            string  `json:"currencyIso"`
					FormattedOriginalPrice string  `json:"formattedOriginalPrice"`
					FormattedValue         string  `json:"formattedValue"`
					OriginalPrice          float64 `json:"originalPrice"`
					Value                  float64 `json:"value"`
				} `json:"price"`
				Style            string `json:"style"`
				Width            string `json:"width"`
				StyleSku         string `json:"styleSku"`
				StyleCode        string `json:"styleCode"`
				MapEnable        bool   `json:"mapEnable"`
				FreeShipping     bool   `json:"freeShipping"`
				WebOnlyLaunch    bool   `json:"webOnlyLaunch"`
				SkuExclusions    bool   `json:"skuExclusions"`
				IsLaunchProduct  bool   `json:"isLaunchProduct"`
				IsInStock        bool   `json:"isInStock"`
				IsKlarnaEligible bool   `json:"isKlarnaEligible"`
				Fit              string `json:"fit"`
			} `json:"/product/converse-all-star-lugged-hi-womens/566155C.html"`
		} `json:"selected"`
		AgeBuckets struct {
			P566155CHTML []interface{} `json:"/product/converse-all-star-lugged-hi-womens/566155C.html"`
		} `json:"ageBuckets"`
	} `json:"details"`
	Router struct {
		Location struct {
			Pathname string `json:"pathname"`
			Search   string `json:"search"`
			Hash     string `json:"hash"`
			Key      string `json:"key"`
		} `json:"location"`
		Action string `json:"action"`
	} `json:"router"`
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

var (
	detailReg = regexp.MustCompile(`window.footlocker.STATE_FROM_SERVER\s*=\s*(.*);`)
	//imageRegStart  = regexp.MustCompile(`(altset_)([a-zA-Z0-9(]+)`)
	imageRegStart = regexp.MustCompile(`\(([^;]+)`)
	//imageRegEnd  = regexp.MustCompile(`(,)(?!.*\1)`)

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

	matched := detailReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		return fmt.Errorf("extract produt json from page %s content failed", resp.Request.URL)
	}

	var (
		i parseProductResponse
		q parseImageResponse
	)

	if err = json.Unmarshal(matched[1], &i); err != nil {
		c.logger.Error(err)
		return err
	}
	router := i.Router.Location.Pathname

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	canUrl := dom.Find(`link[rel="canonical"]`).AttrOr("href", "")
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}

	for _, p := range i.Details.Data[strconv.Format(router)] {
		Sku := p.Sku
		item := pbItem.Product{
			Source: &pbItem.Source{
				Id:           strconv.Format(p.Sku),
				CrawlUrl:     resp.Request.URL.String(),
				CanonicalUrl: canUrl,
			},
			Title:       i.Details.Product[router].Name,
			Description: i.Details.Product[router].Description,
			BrandName:   i.Details.Product[router].Brand,
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
			},
		}
		for j, cate := range i.Details.Product[router].Categories {
			switch j {
			case 0:
				item.Category = cate.Name
			case 1:
				item.SubCategory = cate.Name
			case 2:
				item.SubCategory2 = cate.Name
			case 3:
				item.SubCategory3 = cate.Name
			case 4:
				item.SubCategory4 = cate.Name
			}
		}

		for _, rawSize := range p.Products {
			current, _ := strconv.ParseFloat(p.Price.Value)
			msrp, _ := strconv.ParseFloat(p.Price.OriginalPrice)
			discount := (msrp - current) * 100 / msrp

			sku := pbItem.Sku{
				SourceId: strconv.Format(rawSize.Code),
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(current * 100),
					Msrp:     int32(msrp * 100),
					Discount: int32(discount),
				},
				Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}
			if rawSize.StockLevelStatus == "inStock" {
				sku.Stock.StockStatus = pbItem.Stock_InStock
			}

			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    strconv.Format(rawSize.Size.ID),
				Name:  rawSize.Size.Value,
				Value: rawSize.Size.Value,
			})

			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    strconv.Format(rawSize.Style.ID),
				Name:  rawSize.Style.Value,
				Value: rawSize.Style.Value,
			})

			item.SkuItems = append(item.SkuItems, &sku)
		}

		Skud := ",\"" + Sku + "\""

		imgUrl := "https://images.footlocker.com/is/image/EBFL2/" + Sku + "/?req=set,json&id=" + Sku + "&handler=altset_" + Sku
		//imgreq, _ := c.httpClient.Do(ctx, imgUrl)

		req, err := http.NewRequest(http.MethodGet, imgUrl, nil)

		imgreq, err := c.httpClient.Do(ctx, req)
		if err != nil {
			panic(err)
		}
		defer imgreq.Body.Close()

		respBodyImg, err := ioutil.ReadAll(imgreq.Body)
		if err != nil {
			c.logger.Error(err)
			return err
		}

		matched := imageRegStart.FindSubmatch(respBodyImg)
		if len(matched) <= 1 {
			c.logger.Debugf("data %s", respBodyImg)
			return fmt.Errorf("extract produt json from page %s content failed", resp.Request.URL)
		}

		matched = bytes.Split(matched[1], []byte(Skud))
		if len(matched) <= 1 {
			c.logger.Debugf("data %s", respBodyImg)
			return fmt.Errorf("extract produt json from page %s content failed", resp.Request.URL)
		}

		if err = json.Unmarshal(matched[0], &q); err != nil {
			c.logger.Debugf("parse %s failed, error=%s", matched[2], err)
			return err
		}

		isDefault := true
		for key, img := range q.Set.Item {
			if key > 0 {
				isDefault = false
			}
			if strings.Contains(img.I.N, "Image_Not") {
				continue
			}
			itemImg, _ := anypb.New(&media.Media_Image{ // ask?
				OriginalUrl: "https://images.footlocker.com/is/image/" + img.I.N,
				LargeUrl:    "https://images.footlocker.com/is/image/" + img.I.N + "?wid=1000&hei=1333&fmt=png-alpha",
				MediumUrl:   "https://images.footlocker.com/is/image/" + img.I.N + "?wid=600&hei=800&fmt=png-alpha",
				SmallUrl:    "https://images.footlocker.com/is/image/" + img.I.N + "?wid=495&hei=660&fmt=png-alpha",
			})
			item.Medias = append(item.Medias, &media.Media{
				Detail:    itemImg,
				IsDefault: isDefault,
			})
		}

		//fmt.Println(&item)

		// yield item result
		if err = yield(ctx, &item); err != nil {
			return err
		}

	}
	return nil
}

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		"https://www.footlocker.com/category/womens/clothing.html?query=Clothing+Womens%3Arelevance%3Aproducttype%3AClothing%3Agender%3AWomen%27s%3Aclothstyle%3AJackets",
		// "https://www.footlocker.com/product/jordan-true-flight-mens/42964062.html",
		//"https://www.farfetch.com/shopping/women/escada-floral-print-shirt-item-13761571.aspx?rtype=portal_pdp_outofstock_b&rpos=3&rid=027c2611-6135-4842-abdd-59895d30e924",
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
