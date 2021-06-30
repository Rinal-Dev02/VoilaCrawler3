package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/voiladev/VoilaCrawler/pkg/cli"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	pbProxy "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
	"google.golang.org/protobuf/types/known/anypb"
)

type _Crawler struct {
	httpClient http.Client

	categoryPathMatcher *regexp.Regexp
	productPathMatcher  *regexp.Regexp
	logger              glog.Log
}

func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient: client,
		// /shopping/women/denim-1/items.aspx
		categoryPathMatcher: regexp.MustCompile(`^/(shopping|sets)/(women|men|kids)(/[a-z0-9_\-]+){1,5}(?:items)?.aspx$`),
		productPathMatcher:  regexp.MustCompile(`^/shopping(/[a-z0-9_\-]+){2,5}\-item\-\d+.aspx$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "4c84c118453034662961d6c74c5c4914"
}

// Version
func (c *_Crawler) Version() int32 {
	return 1
}

// CrawlOptions
func (c *_Crawler) CrawlOptions(u *url.URL) *crawler.CrawlOptions {
	options := crawler.NewCrawlOptions()
	options.EnableHeadless = true
	options.EnableSessionInit = true
	options.Reliability = pbProxy.ProxyReliability_ReliabilityMedium
	options.MustCookies = append(options.MustCookies,
		&http.Cookie{Name: "ckm-ctx-sf", Value: `%2F`, Path: "/"},
	)
	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.farfetch.com"}
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
		u.Host = "www.farfetch.com"
	}
	if c.productPathMatcher.MatchString(u.Path) {
		u.RawQuery = ""
		return u.String(), nil
	}
	return rawurl, nil
}

var countriesPrefix = map[string]struct{}{"/ad": {}, "/ae": {}, "/ar-ae": {}, "/af": {}, "/ag": {}, "/ai": {}, "/al": {}, "/am": {}, "/an": {}, "/ao": {}, "/aq": {}, "/ar": {}, "/at": {}, "/au": {}, "/aw": {}, "/az": {}, "/ba": {}, "/bb": {}, "/bd": {}, "/be": {}, "/bf": {}, "/bg": {}, "/bh": {}, "/ar-bh": {}, "/bi": {}, "/bj": {}, "/bm": {}, "/bn": {}, "/bo": {}, "/br": {}, "/bs": {}, "/bt": {}, "/bv": {}, "/bw": {}, "/by": {}, "/bz": {}, "/ca": {}, "/cc": {}, "/cf": {}, "/cg": {}, "/ch": {}, "/ci": {}, "/ck": {}, "/cl": {}, "/cm": {}, "/cn": {}, "/co": {}, "/cr": {}, "/cv": {}, "/cx": {}, "/cy": {}, "/cz": {}, "/de": {}, "/dj": {}, "/dk": {}, "/dm": {}, "/do": {}, "/dz": {}, "/ec": {}, "/ee": {}, "/eg": {}, "/ar-eg": {}, "/eh": {}, "/es": {}, "/et": {}, "/fi": {}, "/fj": {}, "/fk": {}, "/fm": {}, "/fo": {}, "/fr": {}, "/ga": {}, "/uk": {}, "/gd": {}, "/ge": {}, "/gf": {}, "/gg": {}, "/gh": {}, "/gi": {}, "/gl": {}, "/gm": {}, "/gn": {}, "/gp": {}, "/gq": {}, "/gr": {}, "/gt": {}, "/gu": {}, "/gw": {}, "/gy": {}, "/hk": {}, "/hn": {}, "/hr": {}, "/ht": {}, "/hu": {}, "/ic": {}, "/id": {}, "/ie": {}, "/il": {}, "/in": {}, "/io": {}, "/iq": {}, "/ar-iq": {}, "/is": {}, "/it": {}, "/je": {}, "/jm": {}, "/jo": {}, "/ar-jo": {}, "/jp": {}, "/ke": {}, "/kg": {}, "/kh": {}, "/ki": {}, "/km": {}, "/kn": {}, "/kr": {}, "/kv": {}, "/kw": {}, "/ar-kw": {}, "/ky": {}, "/kz": {}, "/la": {}, "/lb": {}, "/ar-lb": {}, "/lc": {}, "/li": {}, "/lk": {}, "/ls": {}, "/lt": {}, "/lu": {}, "/lv": {}, "/ma": {}, "/mc": {}, "/md": {}, "/me": {}, "/mg": {}, "/mh": {}, "/mk": {}, "/ml": {}, "/mn": {}, "/mo": {}, "/mp": {}, "/mq": {}, "/mr": {}, "/ms": {}, "/mt": {}, "/mu": {}, "/mv": {}, "/mw": {}, "/mx": {}, "/my": {}, "/mz": {}, "/na": {}, "/nc": {}, "/ne": {}, "/nf": {}, "/ng": {}, "/ni": {}, "/nl": {}, "/no": {}, "/np": {}, "/nr": {}, "/nu": {}, "/nz": {}, "/om": {}, "/ar-om": {}, "/pa": {}, "/pe": {}, "/pf": {}, "/pg": {}, "/ph": {}, "/pk": {}, "/pl": {}, "/pm": {}, "/pn": {}, "/pr": {}, "/pt": {}, "/pw": {}, "/py": {}, "/qa": {}, "/ar-qa": {}, "/re": {}, "/ro": {}, "/rs": {}, "/ru": {}, "/rw": {}, "/sa": {}, "/ar-sa": {}, "/sb": {}, "/sc": {}, "/se": {}, "/sg": {}, "/sh": {}, "/si": {}, "/sk": {}, "/sl": {}, "/sm": {}, "/sn": {}, "/sr": {}, "/st": {}, "/sv": {}, "/sz": {}, "/tc": {}, "/td": {}, "/tg": {}, "/th": {}, "/tj": {}, "/tk": {}, "/tl": {}, "/tn": {}, "/to": {}, "/tr": {}, "/tt": {}, "/tv": {}, "/tw": {}, "/tz": {}, "/ua": {}, "/ug": {}, "/uy": {}, "/uz": {}, "/va": {}, "/vc": {}, "/ve": {}, "/vg": {}, "/vi": {}, "/vn": {}, "/vu": {}, "/wf": {}, "/xc": {}, "/ye": {}, "/za": {}, "/zm": {}, "/zw": {}}

func getPathFirstSection(p string) string {
	return "/" + strings.SplitN(strings.TrimPrefix(p, "/"), "/", 2)[0]
}

func (c *_Crawler) Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	prefix := getPathFirstSection(resp.Request.URL.Path)
	if _, ok := countriesPrefix[prefix]; ok {
		req := resp.Request
		req.URL.Path = strings.TrimPrefix(req.URL.Path, prefix)

		opts := c.CrawlOptions(req.URL)
		for k := range opts.MustHeader {
			req.Header.Set(k, opts.MustHeader.Get(k))
		}
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
		c.logger.Infof("Access %s", req.URL.String())
		if res, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
			EnableProxy:       true,
			EnableHeadless:    opts.EnableHeadless,
			EnableSessionInit: opts.EnableSessionInit,
			DisableCookieJar:  opts.DisableCookieJar,
			Reliability:       opts.Reliability,
		}); err != nil {
			return err
		} else {
			resp = res
		}
	}

	yieldWrap := func(ctx context.Context, val interface{}) error {
		switch v := val.(type) {
		case *http.Request:
			prefix := getPathFirstSection(v.URL.Path)
			if _, ok := countriesPrefix[prefix]; ok {
				v.URL.Path = strings.TrimPrefix(v.URL.Path, prefix)
			}
			return yield(ctx, v)
		default:
			return yield(ctx, val)
		}
	}

	p := strings.TrimSuffix(resp.Request.URL.Path, "/")
	if p == "" {
		return c.parseCategories(ctx, resp, yield)
	}
	if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yieldWrap)
	} else if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yieldWrap)
	}
	return crawler.ErrUnsupportedPath
}

// nextIndex used to get sharingData from context
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")) + 1)
}

type productListType struct {
	ListingItems struct {
		Items []struct {
			ID  int    `json:"id"`
			URL string `json:"url"`
		} `json:"items"`
	} `json:"listingItems"`
	ListingPagination struct {
		Index                int    `json:"index"`
		View                 int    `json:"view"`
		TotalItems           int    `json:"totalItems"`
		TotalPages           int    `json:"totalPages"`
		NormalizedTotalItems string `json:"normalizedTotalItems"`
	} `json:"listingPagination"`
}

var prodDataExtraReg = regexp.MustCompile(`(?Ums)window\['__initialState_portal-slices-listing__'\]\s*=\s*({.*});?\s*</script>`)
var prodDataExtraReg1 = regexp.MustCompile(`(?Ums)window\['__initialState__'\]\s*=\s*(".*");</script>`)
var prodDataExtraReg2 = regexp.MustCompile(`(?Ums)window\.__HYDRATION_STATE__\s*=\s*(".*");</script>`)

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

	catUrl := "https://www.farfetch.com/headerslice/meganav/GetGenderChildren?genderId=19018&isPreviewMode=false"
	req, err := http.NewRequest(http.MethodGet, catUrl, nil)
	req.Header.Add("accept", "application/json, text/plain, */*")
	req.Header.Add("referer", "https://www.farfetch.com/")
	req.Header.Add("accept-language", "en-GB,en-US;q=0.9,en;q=0.8")
	req.Header.Add("x-requested-with", "XMLHttpRequest")

	catreq, err := c.httpClient.Do(ctx, req)
	if err != nil {
		panic(err)
	}
	defer catreq.Body.Close()

	catBody, err := ioutil.ReadAll(catreq.Body)
	if err != nil {
		c.logger.Error(err)
		return err
	}

	var viewData categoryChildrenStructure
	if err := json.Unmarshal(catBody, &viewData); err != nil {
		c.logger.Errorf("unmarshal cat detail data fialed, error=%s", err)
		return err
	}

	for _, rawCat := range viewData.Children {
		cateName := "Kids"

		subCatName := rawCat.Title
		if subCatName == "" {
			continue
		}
		nnctx := context.WithValue(ctx, "Category", cateName)

		subCat1Name := ""
		for _, rawsubCat := range rawCat.Children {

			for _, rawsub2Cat := range rawsubCat.Children {

				href := rawsub2Cat.Href
				if rawsub2Cat.Href == "" {
					subCat1Name = rawsub2Cat.Title
					continue
				}

				u, err := url.Parse(href)
				if err != nil {
					c.logger.Error("parse url %s failed", href)
					continue
				}

				subCate2Name := subCatName + " > " + rawsub2Cat.Title
				if subCat1Name != "" {
					subCate2Name = subCatName + " > " + subCat1Name + " > " + rawsub2Cat.Title
				}

				if c.categoryPathMatcher.MatchString(u.Path) {
					nnnctx := context.WithValue(nnctx, "SubCategory", subCate2Name)
					req, _ := http.NewRequest(http.MethodGet, href, nil)
					if err := yield(nnnctx, req); err != nil {
						return err
					}
				}
			}
		}
	}

	sel := dom.Find(`nav[role="navigation"]`)
	fmt.Println(len(sel.Nodes))
	for i := range sel.Nodes {
		node := sel.Eq(i)
		cateName := strings.TrimSpace(node.AttrOr("aria-label", ""))
		if cateName == "" {
			continue
		}
		nnctx := context.WithValue(ctx, "Category", cateName)

		subsel1 := node.Find(`ul[data-test="list"]>li`)
		for k := range subsel1.Nodes {
			subnode1 := subsel1.Eq(k)
			if len(subnode1.Find(`div`).Nodes) == 0 {
				continue
			}
			subcat1 := subnode1.Find(`a`).First().Text()
			if subcat1 == "" {
				continue
			}
			if len(subnode1.Find(`a`).First().Find(`span`).Nodes) > 0 {
				subcat1 = subnode1.Find(`a`).First().Find(`span`).Text()
			}

			subcat2 := subnode1.Find(`div > ul > li > ul > li`)

			subtitle := ""
			for j := range subcat2.Nodes {
				currentsubcat := subcat2.Eq(j)

				if len(currentsubcat.Find(`a`).Nodes) == 0 {
					subtitle = currentsubcat.Text()
					continue
				}

				href := currentsubcat.Find(`a`).AttrOr("href", "")

				u, err := url.Parse(href)
				if err != nil {
					c.logger.Error("parse url %s failed", href)
					continue
				}

				subcat2name := subcat1 + ` > ` + subtitle + ` > ` + currentsubcat.Find(`p`).First().Text()

				if c.categoryPathMatcher.MatchString(u.Path) {
					nnnctx := context.WithValue(nnctx, "SubCategory", subcat2name)
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

type categoryChildrenStructure struct {
	Children []struct {
		Type       string `json:"type"`
		UniqueName string `json:"uniqueName"`
		Children   []struct {
			Type     string `json:"type"`
			Children []struct {
				Title string `json:"title"`
				Href  string `json:"href"`
			} `json:"children"`
		} `json:"children"`
		Title string `json:"title"`
		Href  string `json:"href"`
	} `json:"children"`
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

	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}
	sel := dom.Find(`ul[data-testid="product-card-list"]>li[data-testid="productCard"]>a`)

	c.logger.Debugf("found %d", len(sel.Nodes))

	lastIndex := nextIndex(ctx)
	for i := range sel.Nodes {
		node := sel.Eq(i)
		if node.AttrOr("itemtype", "") != "http://schema.org/Product" {
			c.logger.Debug("item type not match")
			continue
		}
		href := node.AttrOr("href", "")
		if href == "" {
			c.logger.Debug("no href found")
			continue
		}
		if req, err := http.NewRequest(http.MethodGet, href, nil); err != nil {
			c.logger.Error(err)
			return err
		} else {
			nctx := context.WithValue(ctx, "item.index", lastIndex+1)
			lastIndex += 1
			if err = yield(nctx, req); err != nil {
				return err
			}
		}
	}

	nextNode := dom.Find(`div[data-testid="pagination"]>div[data-testid="pagination-section"] a[data-testid="page-next"]`).First()
	if href := nextNode.AttrOr("href", ""); href != "" {
		req, _ := http.NewRequest(http.MethodGet, href, nil)
		nctx := context.WithValue(ctx, "item.index", lastIndex)
		return yield(nctx, req)
	}
	return nil
}

type parseProductResponse struct {
	ProductViewModel struct {
		Breadcrumb []struct {
			Data_ffref string `json:"data-ffref"`
			Data_type  string `json:"data-type"`
			Href       string `json:"href"`
			Text       string `json:"text"`
		} `json:"breadcrumb"`
		DesignerDetails struct {
			Name string `json:"name"`
		} `json:"designerDetails"`
		Details struct {
			AgeGroup    string `json:"ageGroup"`
			Colors      string `json:"colors"`
			Department  string `json:"department"`
			Description string `json:"description"`
			Gender      int64  `json:"gender"`
			GenderName  string `json:"genderName"`

			MerchantID int64 `json:"merchantId"`
			ProductID  int64 `json:"productId"`

			ShortDescription string `json:"shortDescription"`
			StyleID          int64  `json:"styleId"`
		} `json:"details"`

		Images struct {
			Main []struct {
				Six00     string `json:"600"`
				Alt       string `json:"alt"`
				Index     int64  `json:"index"`
				Large     string `json:"large"`
				Medium    string `json:"medium"`
				Size200   string `json:"size200"`
				Size240   string `json:"size240"`
				Size300   string `json:"size300"`
				Small     string `json:"small"`
				Thumbnail string `json:"thumbnail"`
				Zoom      string `json:"zoom"`
			} `json:"main"`
		} `json:"images"`

		PriceInfo struct {
			Default struct {
				CurrencyCode                  string  `json:"currencyCode"`
				FinalPrice                    float32 `json:"finalPrice"`
				FormattedFinalPrice           string  `json:"formattedFinalPrice"`
				FormattedFinalPriceInternal   string  `json:"formattedFinalPriceInternal"`
				FormattedInitialPrice         string  `json:"formattedInitialPrice"`
				FormattedInitialPriceInternal string  `json:"formattedInitialPriceInternal"`
				InitialPrice                  float32 `json:"initialPrice"`
				IsOnSale                      bool    `json:"isOnSale"`
				Labels                        struct {
					Duties   string `json:"duties"`
					Discount string `json:"discount"`
				} `json:"labels"`
				PriceTags []string `json:"priceTags"`
			} `json:"default"`
		} `json:"priceInfo"`
		Sizes struct {
			Available map[string]struct {
				Description string `json:"description"`
				LastInStock bool   `json:"lastInStock"`
				Quantity    int64  `json:"quantity"`
				SizeID      int64  `json:"sizeId"`
				StoreID     int64  `json:"storeId"`
				VariantID   string `json:"variantId"`
			} `json:"available"`
		} `json:"sizes"`
	} `json:"productViewModel"`
	Promotions interface{} `json:"promotions"`
}

var (
	detailReg  = regexp.MustCompile(`(?Ums)window\['__initialState_slice-pdp__'\]\s*=\s*(.*);?\s*</script>`)
	detailReg1 = regexp.MustCompile(`(?Ums)window\['__initialState__'\]\s*=\s*(".*");\s*</script>`)
	detailReg2 = regexp.MustCompile(`(?Ums)window.__HYDRATION_STATE__\s*=\s*(".*");?\s*</script>`)
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

	if bytes.Contains(respBody, []byte(`Sorry, this piece is currently out of stock<`)) {
		fmt.Println(`Product not available`)
		return nil
	}
	matched := detailReg.FindSubmatch(respBody)
	if len(matched) == 0 {
		matched = detailReg1.FindSubmatch(respBody)
	}
	if len(matched) == 0 {
		matched = detailReg2.FindSubmatch(respBody)
	}
	if len(matched) == 0 {
		c.httpClient.Jar().Clear(ctx, resp.Request.URL)
		return fmt.Errorf("extract produt json from page %s content failed", resp.Request.URL)
	}

	var (
		i       *parseProductResponse
		rawData = string(matched[1])
	)
	if strings.HasPrefix(rawData, `"`) {
		if rawData, err = strconv.Unquote(rawData); err != nil {
			c.logger.Errorf("unquote raw data %s failed, error=%s", rawData, err)
			return err
		}
		if strings.Contains(rawData, "initialStates") {
			var resp struct {
				InitialStates struct {
					SliceProduct *parseProductResponse `json:"slice-product"`
				} `json:"initialStates"`
			}
			if err = json.Unmarshal([]byte(rawData), &resp); err != nil {
				c.logger.Error(err)
				return err
			}
			i = resp.InitialStates.SliceProduct
		} else {
			var resp struct {
				SliceProduct *parseProductResponse `json:"slice-product"`
			}
			if err = json.Unmarshal([]byte(rawData), &resp); err != nil {
				c.logger.Error(err)
				return err
			}
			i = resp.SliceProduct
		}
	} else {
		var resp parseProductResponse
		if err = json.Unmarshal(matched[1], &resp); err != nil {
			c.logger.Error(err)
			return err
		}
		i = &resp
	}
	if i == nil {
		return errors.New("no detail found")
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	canUrl, _ := c.CanonicalUrl(doc.Find(`link[rel="canonical"]`).AttrOr("href", ""))
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           strconv.Format(i.ProductViewModel.Details.ProductID),
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
		},
		Title:       i.ProductViewModel.Details.ShortDescription,
		Description: i.ProductViewModel.Details.Description,
		BrandName:   i.ProductViewModel.DesignerDetails.Name,
		CrowdType:   i.ProductViewModel.Details.GenderName,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
	}
	for i, bread := range i.ProductViewModel.Breadcrumb {
		switch i {
		case 1:
			item.Category = bread.Text
		case 2:
			item.SubCategory = bread.Text
		case 3:
			item.SubCategory2 = bread.Text
		case 4:
			item.SubCategory3 = bread.Text
		case 5:
			item.SubCategory4 = bread.Text
		}
	}

	discount, _ := strconv.ParseInt(strings.TrimSuffix(i.ProductViewModel.PriceInfo.Default.Labels.Discount, "% Off"))
	current, _ := strconv.ParseFloat(i.ProductViewModel.PriceInfo.Default.FinalPrice)
	msrp, _ := strconv.ParseFloat(i.ProductViewModel.PriceInfo.Default.InitialPrice)

	var medias []*media.Media
	for _, img := range i.ProductViewModel.Images.Main {
		itemImg, _ := anypb.New(&media.Media_Image{
			OriginalUrl: img.Zoom,
			LargeUrl:    img.Zoom, // $S$, $XXL$
			MediumUrl:   strings.ReplaceAll(img.Zoom, "_1000.jpg", "_600.jpg"),
			SmallUrl:    strings.ReplaceAll(img.Zoom, "_1000.jpg", "_400.jpg"),
		})
		medias = append(medias, &media.Media{
			Detail:    itemImg,
			IsDefault: img.Index == 1,
		})
	}
	item.Medias = medias

	for _, rawSize := range i.ProductViewModel.Sizes.Available {
		color := i.ProductViewModel.Details.Colors

		sku := pbItem.Sku{
			SourceId: strconv.Format(rawSize.SizeID),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(current * 100),
				Msrp:     int32(msrp * 100),
				Discount: int32(discount),
			},
			Medias: medias,
			Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}
		if rawSize.Quantity > 0 {
			sku.Stock.StockStatus = pbItem.Stock_InStock
			sku.Stock.StockCount = int32(rawSize.Quantity)
			item.Stock.StockStatus = pbItem.Stock_InStock
		}

		if color != "" {
			sku.SourceId = fmt.Sprintf("%s-%v", color, rawSize.SizeID)
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    color,
				Name:  color,
				Value: color,
			})
		}
		sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecSize,
			Id:    strconv.Format(rawSize.SizeID),
			Name:  rawSize.Description,
			Value: strconv.Format(rawSize.SizeID),
		})
		item.SkuItems = append(item.SkuItems, &sku)
	}

	if err = yield(ctx, &item); err != nil {
		return err
	}
	return nil
}

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		"https://www.farfetch.com",
		//"https://www.farfetch.com/de/shopping/women/denim-1/items.aspx",
		// "https://www.farfetch.com/shopping/women/denim-1/items.aspx",
		// "https://www.farfetch.com/shopping/women/low-classic-rolled-cuffs-high-waisted-jeans-item-16070965.aspx?storeid=9359",
		// "https://www.farfetch.com/de/shopping/women/aztech-mountain-galena-mantel-item-15896311.aspx?storeid=10254",
		//"https://www.farfetch.com/shopping/women/gucci-x-ken-scott-floral-print-shirt-item-16359693.aspx?storeid=9445",
		//"https://www.farfetch.com/shopping/women/escada-floral-print-shirt-item-13761571.aspx?rtype=portal_pdp_outofstock_b&rpos=3&rid=027c2611-6135-4842-abdd-59895d30e924",
		//"https://www.farfetch.com/sets/women/new-in-this-week-eu-women.aspx?view=90&sort=4&scale=280&category=136310",

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
