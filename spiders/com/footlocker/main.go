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
	"time"

	//"github.com/gosimple/slug"
	"github.com/voiladev/VoilaCrawl/pkg/crawler"
	"github.com/voiladev/VoilaCrawl/pkg/net/http"
	"github.com/voiladev/VoilaCrawl/pkg/net/http/cookiejar"
	"github.com/voiladev/VoilaCrawl/pkg/proxy"
	"github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/item"
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
		productPathMatcher:  regexp.MustCompile(`^[/product/]+(/[a-zA-Z0-9-]+)+.html$`),
		imagePathMatcher:    regexp.MustCompile(`^(/[is/image/EBFL2/]+)(/[a-zA-Z0-9_-]+)([/?req=set,json&id=]+([A-Za-z0-9]+))$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "8a8f9fe2e6014e87836e164b176ebfa5"
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
	options.EnableSessionInit = true
	options.MustCookies = append(options.MustCookies) //&http.Cookie{Name: "geocountry", Value: `US`, Path: "/"},
	// &http.Cookie{Name: "browseCountry", Value: "US", Path: "/"},
	// &http.Cookie{Name: "browseCurrency", Value: "USD", Path: "/"},
	// &http.Cookie{Name: "browseLanguage", Value: "en-US", Path: "/"},
	// &http.Cookie{Name: "browseSizeSchema", Value: "US", Path: "/"},
	// &http.Cookie{Name: "browseSizeSchema", Value: "US", Path: "/"},
	// &http.Cookie{Name: "storeCode", Value: "US", Path: "/"},

	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.footlocker.com"}
}

func (c *_Crawler) IsUrlMatch(u *url.URL) bool {
	if c == nil || u == nil {
		return false
	}

	for _, reg := range []*regexp.Regexp{
		c.categoryPathMatcher,
		c.productPathMatcher,
		c.imagePathMatcher,
	} {
		if reg.MatchString(u.Path) {
			return true
		}
	}
	return false
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

	return fmt.Errorf("unsupported url %s", resp.Request.URL.String())
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

	// next page
	matched := prodDataExtraReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		return fmt.Errorf("extract json from product list page %s failed", resp.Request.URL)
	}
	var r struct {
		Error struct {
		} `json:"error"`
		Page struct {
			PageInfo struct {
			} `json:"pageInfo"`
		} `json:"page"`
		Order struct {
		} `json:"order"`
		Product struct {
			Attributes struct {
				Interactions struct {
				} `json:"interactions"`
			} `json:"attributes"`
			ProductInfo struct {
				Description string `json:"description"`
				ProductName string `json:"productName"`
				ProductID   string `json:"productId"`
				Price       struct {
					OriginalPrice float64 `json:"originalPrice"`
					Value         float64 `json:"value"`
				} `json:"price"`
				SelectedStyle    string `json:"selectedStyle"`
				ProductThumbnail string `json:"productThumbnail"`
				ProductImage     string `json:"productImage"`
				SizeVariants     []struct {
					Size      string `json:"size"`
					Available bool   `json:"available"`
				} `json:"sizeVariants"`
				StyleVariants []string `json:"styleVariants"`
				Width         string   `json:"width"`
			} `json:"productInfo"`
		} `json:"product"`
		Search struct {
			Sorts []struct {
				Code     string `json:"code"`
				Name     string `json:"name"`
				Selected bool   `json:"selected"`
			} `json:"sorts"`
			Facets []struct {
				Name        string `json:"name"`
				MultiSelect bool   `json:"multiSelect"`
				Values      []struct {
					Name     string `json:"name"`
					Count    int    `json:"count"`
					Selected bool   `json:"selected"`
					Hide     bool   `json:"hide"`
					Query    struct {
						URL   string `json:"url"`
						Query struct {
							Value string `json:"value"`
						} `json:"query"`
					} `json:"query"`
				} `json:"values"`
				Visible       bool   `json:"visible"`
				Category      bool   `json:"category"`
				Priority      int    `json:"priority"`
				Code          string `json:"code"`
				SelectedCount int    `json:"selectedCount"`
			} `json:"facets"`
			QueryID  string `json:"queryID"`
			MetaData struct {
				Signal             string `json:"signal"`
				SearchType         string `json:"searchType"`
				StoreProductSearch bool   `json:"storeProductSearch"`
				QpLatency          int    `json:"qpLatency"`
			} `json:"metaData"`
			Products []struct {
				Badges struct {
					IsNewProduct bool `json:"isNewProduct"`
					IsSale       bool `json:"isSale"`
				} `json:"badges"`
				BaseOptions []struct {
					Selected struct {
						MapEnable bool   `json:"mapEnable"`
						Style     string `json:"style"`
					} `json:"selected"`
				} `json:"baseOptions"`
				BaseProduct string `json:"baseProduct"`
				Images      []struct {
					Format  string `json:"format"`
					URL     string `json:"url"`
					AltText string `json:"altText"`
				} `json:"images"`
				Name          string `json:"name"`
				OriginalPrice struct {
					Value          float64 `json:"value"`
					FormattedValue string  `json:"formattedValue"`
				} `json:"originalPrice"`
				Price struct {
					Value          float64 `json:"value"`
					FormattedValue string  `json:"formattedValue"`
				} `json:"price"`
				Sku            string `json:"sku"`
				URL            string `json:"url"`
				ImageSku       string `json:"imageSku"`
				VariantOptions []struct {
					Images []struct {
						Format  string `json:"format"`
						URL     string `json:"url"`
						AltText string `json:"altText"`
					} `json:"images"`
					Sku      string `json:"sku"`
					ImageSku string `json:"imageSku"`
				} `json:"variantOptions,omitempty"`
				IsSaleProduct bool `json:"isSaleProduct"`
				IsNewProduct  bool `json:"isNewProduct"`
				LaunchProduct bool `json:"launchProduct"`
			} `json:"products"`
			Term    string `json:"term"`
			Results int    `json:"results"`
		} `json:"search"`
		User struct {
			Profile struct {
			} `json:"profile"`
		} `json:"user"`
		Cart struct {
			Items  []interface{} `json:"items"`
			CartID string        `json:"cartID"`
			Price  struct {
			} `json:"price"`
			Fulfillment struct {
			} `json:"fulfillment"`
		} `json:"cart"`
		Transaction struct {
			Items   []interface{} `json:"items"`
			Profile struct {
				ProfileInfo struct {
				} `json:"profileInfo"`
				Address struct {
				} `json:"address"`
				ShippingAddress struct {
				} `json:"shippingAddress"`
			} `json:"profile"`
			Total struct {
			} `json:"total"`
			TransactionID string `json:"transactionID"`
		} `json:"transaction"`
		Events   []interface{} `json:"events"`
		NotFound struct {
		} `json:"notFound"`
	}

	//matched[2] = bytes.ReplaceAll(bytes.ReplaceAll(matched[2], []byte("\\'"), []byte("'")), []byte(`\\"`), []byte(`\"`))
	if err = json.Unmarshal(matched[2], &r); err != nil {
		c.logger.Debugf("parse %s failed, error=%s", matched[1], err)
		return err
	}

	lastIndex := nextIndex(ctx)
	for _, idv := range r.Search.Products {
		if idv.URL == "" {
			continue
		}
		rawurl := fmt.Sprintf("%s://%s/product/a/%s.html", resp.Request.URL.Scheme, resp.Request.URL.Host, idv.URL)
		// prod page
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
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("currentPage"))

	// check if this is the last page
	if len(r.Search.Products) >= r.Search.Results || lastIndex >= r.Search.Results {
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
		Sizes struct {
			P566155CHTML []struct {
				Name       string `json:"name"`
				Code       string `json:"code"`
				IsDisabled bool   `json:"isDisabled"`
			} `json:"/product/converse-all-star-lugged-hi-womens/566155C.html"`
		} `json:"sizes"`
		Styles struct {
			P566155CHTML []struct {
				Sku        string `json:"sku"`
				Code       string `json:"code"`
				Name       string `json:"name"`
				IsDisabled bool   `json:"isDisabled"`
			} `json:"/product/converse-all-star-lugged-hi-womens/566155C.html"`
		} `json:"styles"`
		Failed struct {
			P566155CHTML bool `json:"/product/converse-all-star-lugged-hi-womens/566155C.html"`
		} `json:"failed"`
		Product map[string]struct {
			//P566155CHTML struct {
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
			//} `json:"/product/converse-all-star-lugged-hi-womens/566155C.html"`
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
		c.logger.Debugf("data %s", respBody)
		return fmt.Errorf("extract produt json from page %s content failed", resp.Request.URL)
	}

	var (
		i parseProductResponse
		q parseImageResponse
	)

	//matched[2] = bytes.ReplaceAll(bytes.ReplaceAll(matched[2], []byte("\\'"), []byte("'")), []byte(`\\"`), []byte(`\"`))

	if err = json.Unmarshal(matched[1], &i); err != nil {
		c.logger.Error(err)
		return err
	}
	router := i.Router.Location.Pathname

	for _, p := range i.Details.Data[strconv.Format(router)] {
		Sku := p.Sku
		item := pbItem.Product{
			Source: &pbItem.Source{
				Id:       strconv.Format(p.Sku),
				CrawlUrl: resp.Request.URL.String(),
			},
			Title:       i.Details.Product[router].Name,
			Description: i.Details.Product[router].Description,
			BrandName:   i.Details.Product[router].Brand,
			//CrowdType:    i.Details.GenderName,  // ASK ?
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
			},
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
				//sku.Stock.StockCount = int32(rawSize.Quantity)
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
	logger := glog.New(glog.LogLevelDebug)
	// build a http client
	// get proxy's microservice address from env
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
				i.URL.Host = "www.footlocker.com"
			}

			// do http requests here.
			nctx, cancel := context.WithTimeout(ctx, time.Minute*5)
			defer cancel()
			resp, err := client.DoWithOptions(nctx, i, http.Options{
				EnableProxy:       true,
				EnableHeadless:    false,
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

	ctx := context.WithValue(context.Background(), "tracing_id", fmt.Sprintf("asos_%d", time.Now().UnixNano()))
	// start the crawl request
	for _, req := range spider.NewTestRequest(context.Background()) {
		if err := callback(ctx, req); err != nil {
			logger.Fatal(err)
		}
	}
}
