package main

// this website exists api robot check. should controller frequence

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
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/voiladev/VoilaCrawl/pkg/crawler"
	"github.com/voiladev/VoilaCrawl/pkg/net/http"
	"github.com/voiladev/VoilaCrawl/pkg/net/http/cookiejar"
	"github.com/voiladev/VoilaCrawl/pkg/proxy"
	"github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/media"
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
		categoryPathMatcher: regexp.MustCompile(`^(/en-[a-z]+)?/c/([a-z0-9]+\-){1,}cat[0-9]+$`),
		productPathMatcher:  regexp.MustCompile(`^(/en-[a-z]+)?/p/([a-z0-9]+\-){1,}prod[0-9]+$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "e7922ae604424feb1e9ad285547b148a"
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
	options.Reliability = 2
	// NOTE: no need to set useragent here for user agent is dynamic
	// options.MustHeader.Set("Accept-Language", "en-US,en;q=0.8")
	// options.MustHeader.Set("X-Requested-With", "XMLHttpRequest")
	options.MustCookies = append(options.MustCookies, &http.Cookie{
		Name:  "profile_data",
		Value: "%7B%22firstName%22%3A%22%22%2C%22currencyPreference%22%3A%22USD%22%2C%22countryPreference%22%3A%22US%22%2C%22securityStatus%22%3A%22Anonymous%22%2C%22cartItemCount%22%3A0%7D",
		Path:  "/",
	})
	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.neimanmarcus.com"}
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

const defaultCategoryProductsPageSize = 120

// parseCategoryProducts parse api url from web page url
func (c *_Crawler) parseCategoryProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if resp.StatusCode == http.StatusForbidden {
		return errors.New("access denied")
	}
	c.logger.Debugf("parse %s", resp.Request.URL)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if isRobotCheckPage(respBody) {
		return errors.New("robot check page")
	}
	// c.logger.Debugf("%s", respBody)

	if !bytes.Contains(respBody, []byte("product-list ")) {
		return errors.New("products not found")
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	lastIndex := nextIndex(ctx)
	sel := doc.Find(`.product-list>.product-thumbnail>a`)
	c.logger.Debugf("nodes %d", len(sel.Nodes))
	for i := range sel.Nodes {
		node := sel.Eq(i)
		if href, _ := node.Attr("href"); href != "" {
			c.logger.Debugf("yield %s", href)
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
	if len(sel.Nodes) < defaultCategoryProductsPageSize {
		return nil
	}

	var page int64 = 1
	if p := resp.Request.URL.Query().Get("page"); p != "" {
		page, _ = strconv.ParseInt(p)
	}

	u := resp.Request.URL
	vals := u.Query()
	vals.Set("page", strconv.Format(page))
	u.RawQuery = vals.Encode()

	nctx := context.WithValue(ctx, "item.index", lastIndex)
	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	return yield(nctx, req)
}

// nextIndex used to get sharingData from context
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")) + 1)
}

type productMedia struct {
	Main struct {
		Thumbnail struct {
			URL string `json:"url"`
			Tag string `json:"tag"`
		} `json:"thumbnail"`
		Medium struct {
			URL string `json:"url"`
			Tag string `json:"tag"`
		} `json:"medium"`
		Large struct {
			URL string `json:"url"`
			Tag string `json:"tag"`
		} `json:"large"`
		MediumShort struct {
			URL string `json:"url"`
			Tag string `json:"tag"`
		} `json:"mediumShort"`
		Small struct {
			URL string `json:"url"`
			Tag string `json:"tag"`
		} `json:"small"`
	} `json:"main"`
	Alternate map[string]struct {
		Thumbnail struct {
			URL string `json:"url"`
			Tag string `json:"tag"`
		} `json:"thumbnail"`
		Medium struct {
			URL string `json:"url"`
			Tag string `json:"tag"`
		} `json:"medium"`
		Large struct {
			URL string `json:"url"`
			Tag string `json:"tag"`
		} `json:"large"`
		MediumShort struct {
			URL string `json:"url"`
			Tag string `json:"tag"`
		} `json:"mediumShort"`
		Small struct {
			URL string `json:"url"`
			Tag string `json:"tag"`
		} `json:"small"`
	} `json:"alternate"`
}

type linkedData struct {
	Context     string `json:"@context"`
	Type        string `json:"@type"`
	Name        string `json:"name"`
	Brand       string `json:"brand"`
	Image       string `json:"image"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Offers      struct {
		Type          string `json:"@type"`
		PriceCurrency string `json:"priceCurrency"`
		Offers        []struct {
			Type          string `json:"@type"`
			PriceCurrency string `json:"priceCurrency"`
			Availability  string `json:"availability"`
			Price         string `json:"price"`
			Sku           string `json:"sku"`
			ItemOffered   struct {
				Type  string `json:"@type"`
				Color string `json:"color"`
			} `json:"itemOffered"`
		} `json:"offers"`
		LowPrice  string `json:"lowPrice"`
		HighPrice string `json:"highPrice"`
	} `json:"offers"`
}

type productDetail struct {
	Quantity                      int                        `json:"quantity"`
	ActiveMediaIndex              int                        `json:"activeMediaIndex"`
	LinkedData                    *linkedData                `json:"linkedData"`
	LinkedDataWithAllProdsAndSKUs *linkedData                `json:"linkedDataWithAllProdsAndSKUs"`
	VideoActive                   bool                       `json:"videoActive"`
	ActivePDPTab                  int                        `json:"activePDPTab"`
	DeliveryDate                  string                     `json:"deliveryDate"`
	VendorRestrictedDates         []interface{}              `json:"vendorRestrictedDates"`
	BopsErrorForReplenishment     bool                       `json:"bopsErrorForReplenishment"`
	FavAddRemoveStatus            string                     `json:"favAddRemoveStatus"`
	IsPersonalizationSelected     bool                       `json:"isPersonalizationSelected"`
	AddToBagError                 string                     `json:"addToBagError"`
	BopsError                     string                     `json:"bopsError"`
	IsChanel                      bool                       `json:"isChanel"`
	IsZeroDollarProduct           bool                       `json:"isZeroDollarProduct"`
	IsGroup                       bool                       `json:"isGroup"`
	ChildProducts                 map[string]json.RawMessage `json:"childProducts"`
	Options                       struct {
		SelectedColorIndex int `json:"selectedColorIndex"`
		ProductOptions     []struct {
			Label  string `json:"label"`
			Values []struct {
				Name          string        `json:"name"`
				Key           string        `json:"key"`
				DefaultColor  bool          `json:"defaultColor"`
				DisplaySkuImg bool          `json:"displaySkuImg"`
				Url           string        `json:"url"`
				Media         *productMedia `json:"media"`
			} `json:"values"`
		} `json:"productOptions"`
		OptionMatrix [][]string `json:"optionMatrix"`
	} `json:"options"`
	ID                     string `json:"id"`
	Name                   string `json:"name"`
	NameHTML               string `json:"nameHtml"`
	Displayable            bool   `json:"displayable"`
	Perishable             bool   `json:"perishable"`
	Replenishable          bool   `json:"replenishable"`
	IsCustomizable         bool   `json:"isCustomizable"`
	CustomizationSupported bool   `json:"customizationSupported"`
	Details                struct {
		Title                 string `json:"title"`
		LongDesc              string `json:"longDesc"`
		CanonicalURL          string `json:"canonicalUrl"`
		SizeGuide             string `json:"sizeGuide"`
		PickupInStoreEligible bool   `json:"pickupInStoreEligible"`
		ShopRunnerEligible    bool   `json:"shopRunnerEligible"`
		ParentheticalCharge   string `json:"parentheticalCharge"`
		ClearanceItem         bool   `json:"clearanceItem"`
		PreciousJewelryItem   bool   `json:"preciousJewelryItem"`
		SuppressCheckout      bool   `json:"suppressCheckout"`
	} `json:"details"`
	Designer struct {
		Name                string `json:"name"`
		Title               string `json:"title"`
		ShortDesc           string `json:"shortDesc"`
		DesignerBoutiqueURL string `json:"designerBoutiqueUrl"`
	} `json:"designer"`
	Metadata struct {
		CmosCatalogID string `json:"cmosCatalogId"`
		CmosItem      string `json:"cmosItem"`
		PimStyle      string `json:"pimStyle"`
	} `json:"metadata"`
	Media *productMedia `json:"media"`
	Price struct {
		CurrencyCode string `json:"currencyCode"`
		RetailPrice  string `json:"retailPrice"`
	} `json:"price"`
	PreOrder     bool `json:"preOrder"`
	ProductFlags struct {
		HasMoreColors        bool `json:"hasMoreColors"`
		IsOnlyAtNM           bool `json:"isOnlyAtNM"`
		ShowMonogramLabel    bool `json:"showMonogramLabel"`
		IsNewArrival         bool `json:"isNewArrival"`
		InLookBook           bool `json:"inLookBook"`
		PreviewSupported     bool `json:"previewSupported"`
		DynamicImageSkuColor bool `json:"dynamicImageSkuColor"`
		IsEditorial          bool `json:"isEditorial"`
		IsEvening            bool `json:"isEvening"`
	} `json:"productFlags"`
	DepartmentCode          string `json:"departmentCode"`
	IsProactiveChatEligible bool   `json:"isProactiveChatEligible"`
	VendorID                string `json:"vendorId"`
	Skus                    []struct {
		ID               string `json:"id"`
		UseSkuAsset      bool   `json:"useSkuAsset"`
		PreOrder         bool   `json:"preOrder"`
		BackOrder        bool   `json:"backOrder"`
		InStock          bool   `json:"inStock"`
		DropShip         bool   `json:"dropShip"`
		DiscontinuedCode string `json:"discontinuedCode"`
		Metadata         struct {
			CmosSkuID string `json:"cmosSkuId"`
			PimSkuID  string `json:"pimSkuId"`
		} `json:"metadata"`
		Color struct {
			Name         string `json:"name"`
			Key          string `json:"key"`
			DefaultColor bool   `json:"defaultColor"`
		} `json:"color"`
		Size struct {
			Name string `json:"name"`
			Key  string `json:"key"`
		} `json:"size"`
		StockStatusMessage    string `json:"stockStatusMessage"`
		PoQuantity            bool   `json:"poQuantity"`
		ShipFromStore         bool   `json:"shipFromStore"`
		StockLevel            int    `json:"stockLevel"`
		PurchaseOrderQuantity int    `json:"purchaseOrderQuantity"`
		ColorIndex            string `json:"colorIndex"`
		SizeIndex             string `json:"sizeIndex"`
		DisplaySkuImg         bool   `json:"displaySkuImg"`
		Height                int    `json:"height"`
		Width                 int    `json:"width"`
		Depth                 int    `json:"depth"`
		Sellable              bool   `json:"sellable"`
	} `json:"skus"`
	IsFavorite        bool   `json:"isFavorite"`
	DisplayOutfitting bool   `json:"displayOutfitting"`
	SellableDate      string `json:"sellableDate"`
	Hierarchy         []struct {
		Level1 string `json:"level1"`
		Level2 string `json:"level2"`
	} `json:"hierarchy"`
	ServiceLevelCodes []string `json:"serviceLevelCodes"`
	IsOutOfStock      bool     `json:"isOutOfStock"`
}

type productGroupDetail struct {
	Quantity                  int                        `json:"quantity"`
	ActiveMediaIndex          int                        `json:"activeMediaIndex"`
	VideoActive               bool                       `json:"videoActive"`
	ActivePDPTab              int                        `json:"activePDPTab"`
	DeliveryDate              string                     `json:"deliveryDate"`
	VendorRestrictedDates     []interface{}              `json:"vendorRestrictedDates"`
	BopsErrorForReplenishment bool                       `json:"bopsErrorForReplenishment"`
	FavAddRemoveStatus        string                     `json:"favAddRemoveStatus"`
	IsPersonalizationSelected bool                       `json:"isPersonalizationSelected"`
	AddToBagError             string                     `json:"addToBagError"`
	BopsError                 string                     `json:"bopsError"`
	IsChanel                  bool                       `json:"isChanel"`
	IsZeroDollarProduct       bool                       `json:"isZeroDollarProduct"`
	IsGroup                   bool                       `json:"isGroup"`
	ChildProducts             map[string]json.RawMessage `json:"childProducts"`
	Options                   struct {
		SelectedColorIndex int `json:"selectedColorIndex"`
		ProductOptions     []struct {
			Label  string `json:"label"`
			Values []struct {
				Name          string        `json:"name"`
				Key           string        `json:"key"`
				DefaultColor  bool          `json:"defaultColor"`
				DisplaySkuImg bool          `json:"displaySkuImg"`
				Url           string        `json:"url"`
				Media         *productMedia `json:"media"`
			} `json:"values"`
		} `json:"productOptions"`
		OptionMatrix [][]string `json:"optionMatrix"`
	} `json:"options"`
	ID                     string `json:"id"`
	Name                   string `json:"name"`
	NameHTML               string `json:"nameHtml"`
	Displayable            bool   `json:"displayable"`
	Perishable             bool   `json:"perishable"`
	Replenishable          bool   `json:"replenishable"`
	IsCustomizable         bool   `json:"isCustomizable"`
	CustomizationSupported bool   `json:"customizationSupported"`
	Details                struct {
		Title                 string `json:"title"`
		LongDesc              string `json:"longDesc"`
		CanonicalURL          string `json:"canonicalUrl"`
		SizeGuide             string `json:"sizeGuide"`
		PickupInStoreEligible bool   `json:"pickupInStoreEligible"`
		ShopRunnerEligible    bool   `json:"shopRunnerEligible"`
		ParentheticalCharge   string `json:"parentheticalCharge"`
		ClearanceItem         bool   `json:"clearanceItem"`
		PreciousJewelryItem   bool   `json:"preciousJewelryItem"`
		SuppressCheckout      bool   `json:"suppressCheckout"`
	} `json:"details"`
	Designer struct {
		Name                string `json:"name"`
		Title               string `json:"title"`
		ShortDesc           string `json:"shortDesc"`
		DesignerBoutiqueURL string `json:"designerBoutiqueUrl"`
	} `json:"designer"`
	Metadata struct {
		CmosCatalogID string `json:"cmosCatalogId"`
		CmosItem      string `json:"cmosItem"`
		PimStyle      string `json:"pimStyle"`
	} `json:"metadata"`
	Media *productMedia `json:"media"`
	Price struct {
		CurrencyCode string `json:"currencyCode"`
		RetailPrice  string `json:"retailPrice"`
	} `json:"price"`
	PreOrder     bool `json:"preOrder"`
	ProductFlags struct {
		HasMoreColors        bool `json:"hasMoreColors"`
		IsOnlyAtNM           bool `json:"isOnlyAtNM"`
		ShowMonogramLabel    bool `json:"showMonogramLabel"`
		IsNewArrival         bool `json:"isNewArrival"`
		InLookBook           bool `json:"inLookBook"`
		PreviewSupported     bool `json:"previewSupported"`
		DynamicImageSkuColor bool `json:"dynamicImageSkuColor"`
		IsEditorial          bool `json:"isEditorial"`
		IsEvening            bool `json:"isEvening"`
	} `json:"productFlags"`
	DepartmentCode          string `json:"departmentCode"`
	IsProactiveChatEligible bool   `json:"isProactiveChatEligible"`
	VendorID                string `json:"vendorId"`
	Skus                    []struct {
		ID               string `json:"id"`
		UseSkuAsset      bool   `json:"useSkuAsset"`
		PreOrder         bool   `json:"preOrder"`
		BackOrder        bool   `json:"backOrder"`
		InStock          bool   `json:"inStock"`
		DropShip         bool   `json:"dropShip"`
		DiscontinuedCode string `json:"discontinuedCode"`
		Metadata         struct {
			CmosSkuID string `json:"cmosSkuId"`
			PimSkuID  string `json:"pimSkuId"`
		} `json:"metadata"`
		Color struct {
			Name         string `json:"name"`
			Key          string `json:"key"`
			DefaultColor bool   `json:"defaultColor"`
		} `json:"color"`
		Size struct {
			Name string `json:"name"`
			Key  string `json:"key"`
		} `json:"size"`
		StockStatusMessage    string `json:"stockStatusMessage"`
		PoQuantity            bool   `json:"poQuantity"`
		ShipFromStore         bool   `json:"shipFromStore"`
		StockLevel            int    `json:"stockLevel"`
		PurchaseOrderQuantity int    `json:"purchaseOrderQuantity"`
		ColorIndex            string `json:"colorIndex"`
		SizeIndex             string `json:"sizeIndex"`
		DisplaySkuImg         bool   `json:"displaySkuImg"`
		Height                int    `json:"height"`
		Width                 int    `json:"width"`
		Depth                 int    `json:"depth"`
		Sellable              bool   `json:"sellable"`
	} `json:"skus"`
	IsFavorite        bool   `json:"isFavorite"`
	DisplayOutfitting bool   `json:"displayOutfitting"`
	SellableDate      string `json:"sellableDate"`
	Hierarchy         []struct {
		Level1 string `json:"level1"`
		Level2 string `json:"level2"`
	} `json:"hierarchy"`
	ServiceLevelCodes []string `json:"serviceLevelCodes"`
	IsOutOfStock      bool     `json:"isOutOfStock"`
}

// Generate data struct from json https://mholt.github.io/json-to-go/
type productDetailPage struct {
	Navigation struct {
		NavSlider struct {
			SliderMenuVisible bool   `json:"sliderMenuVisible"`
			SearchTerm        string `json:"searchTerm"`
			AccountExpanded   bool   `json:"accountExpanded"`
		} `json:"navSlider"`
		Breadcrumbs []struct {
			ID            string `json:"id"`
			Name          string `json:"name"`
			NameForMobile string `json:"nameForMobile"`
			URL           string `json:"url"`
		} `json:"breadcrumbs"`
		BreadcrumbPath        string `json:"breadcrumbPath"`
		SiloDrawerHoverIntent struct {
			Sensitivity int `json:"sensitivity"`
			Interval    int `json:"interval"`
			Timeout     int `json:"timeout"`
		} `json:"siloDrawerHoverIntent"`
	} `json:"navigation"`
	ProductCatalog struct {
		Product *productDetail      `json:"product"`
		Group   *productGroupDetail `json:"group"`
	} `json:"productCatalog"`
}

func (c *_Crawler) parseProduct(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}
	if resp.StatusCode == http.StatusForbidden {
		return errors.New("access denied")
	}

	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if isRobotCheckPage(respbody) {
		return errors.New("robot check page")
	}

	c.logger.Debugf("%s", respbody)

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respbody))
	if err != nil {
		return err
	}

	stateContext := strings.TrimSpace(doc.Find("#state").Text())
	var pageData productDetailPage
	if stateContext != "" {
		if err := json.Unmarshal([]byte(stateContext), &pageData); err != nil {
			c.logger.Debugf("%s", respbody)
			return err
		}

		var prods []*productDetail
		if !pageData.ProductCatalog.Product.IsGroup {
			prods = append(prods, pageData.ProductCatalog.Product)
		} else {
			for key, val := range pageData.ProductCatalog.Group.ChildProducts {
				if key == "productIds" || !strings.HasPrefix(key, "prod") {
					continue
				}
				var prod productDetail
				data, _ := val.MarshalJSON()
				if err := json.Unmarshal(data, &prod); err != nil {
					c.logger.Debugf("unmarshal child prod failed, error=%s", err)
					continue
				}
				prods = append(prods, &prod)
			}
		}
		c.logger.Debugf("found %d", len(prods))

		for _, i := range prods {
			item := pbItem.Product{
				Source: &pbItem.Source{
					Id:       i.ID,
					CrawlUrl: resp.Request.URL.String(),
				},
				Title:       i.Name,
				Description: i.LinkedData.Description,
				BrandName:   i.LinkedData.Brand,
				CrowdType:   "",
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
				},
			}
			if len(pageData.Navigation.Breadcrumbs) >= 1 {
				item.Category = pageData.Navigation.Breadcrumbs[0].Name
			}
			if len(pageData.Navigation.Breadcrumbs) >= 2 {
				item.SubCategory = pageData.Navigation.Breadcrumbs[1].Name
			}
			if len(pageData.Navigation.Breadcrumbs) >= 3 {
				item.SubCategory2 = pageData.Navigation.Breadcrumbs[2].Name
			}
			for _, v := range []string{"woman", "women", "female"} {
				if strings.Contains(strings.ToLower(item.Category), v) {
					item.CrowdType = "women"
					break
				}
			}
			for _, v := range []string{"man", "men", "male"} {
				if strings.Contains(strings.ToLower(item.Category), v) {
					item.CrowdType = "men"
					break
				}
			}
			for _, v := range []string{"kid", "child", "girl", "boy"} {
				if strings.Contains(strings.ToLower(item.Category), v) {
					item.CrowdType = "kids"
					break
				}
			}

			// Default product image some product only exists the main image
			var mediumUrl = i.Media.Main.Large.URL
			if u, _ := url.Parse(i.Media.Main.Large.URL); u != nil {
				vals := u.Query()
				vals.Set("wid", "600")
				vals.Set("height", "750")
				u.RawQuery = vals.Encode()
				mediumUrl = u.String()
			}
			item.Medias = append(item.Medias,
				media.NewImageMedia("",
					i.Media.Main.Large.URL,
					i.Media.Main.Large.URL,
					mediumUrl,
					i.Media.Main.Medium.URL, "", true))
			for _, m := range i.Media.Alternate {
				mediumUrl = i.Media.Main.Large.URL
				if u, _ := url.Parse(i.Media.Main.Large.URL); u != nil {
					vals := u.Query()
					vals.Set("wid", "600")
					vals.Set("height", "750")
					u.RawQuery = vals.Encode()
					mediumUrl = u.String()
				}
				item.Medias = append(item.Medias,
					media.NewImageMedia("",
						m.Large.URL,
						m.Large.URL,
						mediumUrl,
						m.Medium.URL, "", false))
			}

			var (
				skuSpecOptions = map[string]*pbItem.SkuSpecOption{}
				colorMedias    = map[string][]*media.Media{}
			)
			for _, opt := range i.Options.ProductOptions {
				switch opt.Label {
				case "size":
					for i, val := range opt.Values {
						skuSpecOptions[val.Key] = &pbItem.SkuSpecOption{
							Type:  pbItem.SkuSpecType_SkuSpecSize,
							Id:    val.Key,
							Name:  val.Name,
							Value: val.Key,
							Index: int32(i),
						}
					}
				case "color", "colour":
					for i, val := range opt.Values {
						skuSpecOptions[val.Key] = &pbItem.SkuSpecOption{
							Type:  pbItem.SkuSpecType_SkuSpecColor,
							Id:    val.Key,
							Name:  val.Name,
							Value: val.Key,
							Index: int32(i),
							Icon:  val.Url,
						}

						if val.Media != nil {
							var mediumUrl = val.Media.Main.Large.URL
							if u, _ := url.Parse(val.Media.Main.Large.URL); u != nil {
								vals := u.Query()
								vals.Set("wid", "600")
								vals.Set("height", "750")
								u.RawQuery = vals.Encode()
								mediumUrl = u.String()
							}
							// not all color got images
							colorMedias[val.Key] = append(colorMedias[val.Key],
								media.NewImageMedia("",
									val.Media.Main.Large.URL,
									val.Media.Main.Large.URL,
									mediumUrl,
									val.Media.Main.Medium.URL, "", true))
							for _, m := range val.Media.Alternate {
								mediumUrl = val.Media.Main.Large.URL
								if u, _ := url.Parse(val.Media.Main.Large.URL); u != nil {
									vals := u.Query()
									vals.Set("wid", "600")
									vals.Set("height", "750")
									u.RawQuery = vals.Encode()
									mediumUrl = u.String()
								}
								colorMedias[val.Key] = append(colorMedias[val.Key],
									media.NewImageMedia("",
										m.Large.URL,
										m.Large.URL,
										mediumUrl,
										m.Medium.URL, "", false))
							}
						}
					}
				}
			}
			prices := map[string]*pbItem.Price{}
			for _, offer := range i.LinkedData.Offers.Offers {
				val, _ := strconv.ParseFloat(offer.Price)
				p := pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(val * 100),
				}
				prices[offer.Sku] = &p
			}
			for _, rawSku := range i.Skus {
				sku := pbItem.Sku{
					SourceId: rawSku.ID,
					Price:    prices[rawSku.ID],
					Stock: &pbItem.Stock{
						StockCount: int32(rawSku.StockLevel),
					},
					Stats: &pbItem.Stats{},
				}
				if rawSku.InStock {
					sku.Stock.StockStatus = pbItem.Stock_InStock
				} else {
					sku.Stock.StockStatus = pbItem.Stock_OutOfStock
				}
				if opt := skuSpecOptions[rawSku.Color.Key]; opt != nil {
					sku.Specs = append(sku.Specs, opt)
				}
				if opt := skuSpecOptions[rawSku.Size.Key]; opt != nil {
					sku.Specs = append(sku.Specs, opt)
				}
				if medias := colorMedias[rawSku.Color.Key]; medias != nil {
					sku.Medias = medias
				}
				item.SkuItems = append(item.SkuItems, &sku)
			}
			if err := yield(ctx, &item); err != nil {
				return err
			}
		}
	} else {
		// NOTE: see page https://www.neimanmarcus.com/p/veronica-beard-jacket-dickey-prod194270044
		itemId, exists := doc.Find(`input[name="itemId"]`).Attr("value")
		if !exists {
			return fmt.Errorf("item product id not found for %s", resp.Request.URL)
		}

		prodId, exists := doc.Find(`input#prod0`).Attr("value")
		if !exists {
			return fmt.Errorf("product id not found for %s", resp.Request.URL)
		}

		u := strings.Replace(resp.Request.URL.String(), itemId, prodId, -1)
		req, err := http.NewRequest(http.MethodGet, u, nil)
		if err != nil {
			return err
		}
		return yield(ctx, req)
	}
	return nil
}

func (c *_Crawler) NewTestRequest(ctx context.Context) []*http.Request {
	var reqs []*http.Request
	for _, u := range []string{
		// "https://www.neimanmarcus.com/p/veronica-beard-jacket-dickey-prod194270044",
		"https://www.neimanmarcus.com/c/womens-clothing-clothing-coats-jackets-cat77190754?navpath=cat000000_cat000001_cat58290731_cat77190754",
		// "https://www.neimanmarcus.com/p/moncler-hermine-hooded-puffer-jacket-prod197621217?childItemId=NMTS7Q4_41&navpath=cat000000_cat000001_cat58290731_cat77190754&page=0&position=0",
		// "https://www.neimanmarcus.com/p/moncler-moka-shiny-fitted-puffer-coat-with-hood-and-matching-items-prod213210002?childItemId=NMTA8BE_&focusProductId=prod180340224&navpath=cat000000_cat000001_cat58290731_cat77190754&page=0&position=27",
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
				i.URL.Host = "www.neimanmarcus.com"
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

	ctx := context.WithValue(context.Background(), "tracing_id", fmt.Sprintf("tracing_%d", time.Now().UnixNano()))
	// start the crawl request
	for _, req := range spider.NewTestRequest(context.Background()) {
		if err := callback(ctx, req); err != nil {
			logger.Fatal(err)
		}
	}
}
