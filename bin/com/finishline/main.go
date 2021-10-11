package main

// this website exists api robot check. should controller frequence

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/voiladev/VoilaCrawler/pkg/cli"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/regulation"

	pbMedia "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/media"
	pbItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	pbProxy "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
)

type _Crawler struct {
	httpClient http.Client

	categoryPathMatcher *regexp.Regexp
	productPathMatcher  *regexp.Regexp
	logger              glog.Log
}

func (_ *_Crawler) New(_ *cli.Context, client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient: client,
		// /store/women/shoes/running/_/N-nat3jh
		categoryPathMatcher: regexp.MustCompile(`^/store(/[a-zA-Z0-9\-_]+){1,5}$`),
		productPathMatcher:  regexp.MustCompile(`^((.*)(/product/)(.*))|(/store/product/(.*)(/prod\d+))$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "d6c86bd6d65ee27208f6d63b76964189"
}

// Version
func (c *_Crawler) Version() int32 {
	return 1
}

// CrawlOptions
func (c *_Crawler) CrawlOptions(u *url.URL) *crawler.CrawlOptions {
	opts := crawler.NewCrawlOptions()
	opts.EnableHeadless = false
	opts.EnableSessionInit = false
	opts.Reliability = pbProxy.ProxyReliability_ReliabilityDefault
	return opts
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.finishline.com"}
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
		u.Host = "www.finishline.com"
	}
	if c.productPathMatcher.MatchString(u.Path) {
		u.RawQuery = ""
		return u.String(), nil
	}
	return u.String(), nil
}

// GetCategories
func (c *_Crawler) GetCategories(ctx context.Context) ([]*pbItem.Category, error) {
	req, _ := http.NewRequest(http.MethodGet, "https://www.finishline.com/", nil)
	opts := c.CrawlOptions(req.URL)
	resp, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
		EnableProxy:       true,
		EnableHeadless:    opts.EnableHeadless,
		EnableSessionInit: opts.EnableSessionInit,
		Reliability:       opts.Reliability,
		DisableCookieJar:  opts.DisableCookieJar,
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

	var cates []*pbItem.Category

	sel := dom.Find(`#desktop-mainmenu>li`)
	for i := range sel.Nodes {
		node := sel.Eq(i)
		cateName := strings.TrimSpace(node.AttrOr("data-mainitem", ""))
		if cateName == "" || strings.ToLower(cateName) == "brands" || strings.ToLower(cateName) == "releases" {
			continue
		}
		cate := pbItem.Category{Name: cateName}
		cates = append(cates, &cate)

		subSel := node.Find(`.menu-dropdown-row>div>ul`)
		for k := range subSel.Nodes {
			subNode := subSel.Eq(k)
			subcateName := strings.TrimSpace(subNode.Find(`li`).First().Text())
			href, _ := c.CanonicalUrl(subNode.Find("li>a").First().AttrOr("href", ""))

			subCate := pbItem.Category{Name: subcateName, Url: href}
			cate.Children = append(cate.Children, &subCate)

			subNode2 := subNode.Find(`li`)
			for j := range subNode2.Nodes {
				if j == 0 {
					continue
				}
				subNode := subNode2.Eq(j)
				subCate2Name := subNode.Text()

				href, _ := c.CanonicalUrl(subNode.Find(`a`).AttrOr("href", ""))
				if href == "" {
					continue
				}
				subCate2 := pbItem.Category{Name: subCate2Name, Url: href}
				subCate.Children = append(subCate.Children, &subCate2)
			}
		}

		subSel = node.Find(`.navigation-promo-cta`)
		if len(subSel.Nodes) == 0 {
			subSel = node.Find(`.menu-dropdown-row`).Find(`a`)
		}
		for j := range subSel.Nodes {
			subNode := subSel.Eq(j)

			href := subNode.Find(`a`).AttrOr("href", "")
			if href == "" {
				href = subNode.AttrOr("href", "")
			}
			href, _ = c.CanonicalUrl(href)
			if href == "" {
				continue
			}

			subCateName := strings.TrimSpace(subNode.Find(`a`).Text())
			if subCateName == "" {
				subCateName = subNode.AttrOr("title", "")
			}
			subCate := pbItem.Category{Name: subCateName, Url: href}
			cate.Children = append(cate.Children, &subCate)
		}
	}
	return cates, nil
}

func (c *_Crawler) Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}
	p := strings.TrimSuffix(resp.RawUrl().Path, "/")
	if p == "" {
		return crawler.ErrUnsupportedPath
	}
	if c.productPathMatcher.MatchString(resp.RawUrl().Path) {
		return c.parseProduct(ctx, resp, yield)
	} else if c.categoryPathMatcher.MatchString(resp.RawUrl().Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	}
	return crawler.ErrUnsupportedPath
}

func isRobotCheckPage(respBody []byte) bool {
	return bytes.Contains(respBody, []byte("we believe you are using automation tools to browse the website")) ||
		bytes.Contains(respBody, []byte("Javascript is disabled or blocked by an extension")) ||
		bytes.Contains(respBody, []byte("Your browser does not support cookies"))
}

const defaultCategoryProductsPageSize = 40

// parseCategoryProducts parse api url from web page url
func (c *_Crawler) parseCategoryProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil {
		return nil
	}

	if resp.StatusCode == http.StatusForbidden {
		return errors.New("access denied")
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	lastIndex := nextIndex(ctx)

	parseItem := func(sel *goquery.Selection) error {
		c.logger.Debugf("nodes %d", len(sel.Nodes))
		for i := range sel.Nodes {
			node := sel.Eq(i)
			if href, _ := node.Attr("href"); href != "" {
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
		return nil
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	if err := parseItem(doc.Find(`.product-card>a`)); err != nil {
		c.logger.Error(err)
		return err
	}
	subDom, err := goquery.NewDocumentFromReader(strings.NewReader(doc.Find("#additionalProducts").Text()))
	if err != nil {
		c.logger.Error(err)
		return err
	}
	if err := parseItem(subDom.Find(`.product-card>a`)); err != nil {
		c.logger.Error(err)
		return err
	}

	nextSel := doc.Find(`.downPagination .pag-button.next`)
	if strings.Contains(nextSel.AttrOr("class", ""), "disabled") {
		return nil
	}
	nextUrl := nextSel.AttrOr("href", "")
	if nextUrl == "" {
		return nil
	}

	nctx := context.WithValue(ctx, "item.index", lastIndex)
	req, _ := http.NewRequest(http.MethodGet, nextUrl, nil)
	return yield(nctx, req)
}

// nextIndex used to get sharingData from context
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")) + 1)
}

var (
	imgWidthTplReg  = regexp.MustCompile(`&+w=\d+`)
	imgHeightTplReg = regexp.MustCompile(`&+h=\d+`)
)

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

	opts := c.CrawlOptions(resp.Request.URL)
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respbody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	var brandregx = regexp.MustCompile(`"(?U)product_brand"\s*:\s*\["(.*)"\],`)
	matched := brandregx.FindSubmatch(respbody)
	if len(matched) < 2 {
		return fmt.Errorf("not brand found")
	}
	brandName := strings.TrimSpace(string(matched[1]))

	canUrl := doc.Find(`link[rel="canonical"]`).AttrOr("href", "")
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}
	item := pbItem.Product{
		Source: &pbItem.Source{
			Id:           doc.Find("#productItemId").AttrOr("value", doc.Find("#tfc_productid").AttrOr("value", "")),
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
		},
		Title:       strings.Trim(doc.Find(`.hmb-2.titleDesk`).Eq(0).Text(), " \n\r"),
		Description: strings.Trim(doc.Find(`#productDescription`).Text(), " \n\r"),
		BrandName:   brandName,
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
	}

	{
		sel := doc.Find(`.breadcrumbs>li`)
		for i := range sel.Nodes {
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
	}

	colorSel := doc.Find(`#alternateColors .colorway`)
	for i := range colorSel.Nodes {
		node := colorSel.Eq(i)
		style := strings.TrimSpace(node.Find(`a`).AttrOr("data-styleid", node.Find(`a`).AttrOr("data-productid", "")))

		var medias []*pbMedia.Media
		if !strings.Contains(node.Find(`a>.color-image`).AttrOr(`class`, ""), "selected") {
			// load images
			rawurl := fmt.Sprintf("https://www.finishline.com/store/browse/gadgets/alternateImage.jsp?colorID=%s&styleID=%s&productName=&productItemId=%s&productIsShoe=true&productIsAccessory=false&productIsGiftCard=false&renderType=desktop&pageName=pdp", style, style, item.Source.Id)
			req, _ := http.NewRequest(http.MethodGet, rawurl, nil)
			req.Header.Set("Referer", resp.Request.URL.String())
			if resp.Request.Header.Get("User-Agent") != "" {
				req.Header.Set("User-Agent", resp.Request.Header.Get("User-Agent"))
			}
			c.logger.Debugf("Access images %s", rawurl)
			resp, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
				EnableProxy:       true,
				EnableHeadless:    opts.EnableHeadless,
				EnableSessionInit: opts.EnableSessionInit,
				Reliability:       opts.Reliability,
			})
			if err != nil {
				c.logger.Error(err)
				return err
			}
			if resp.StatusCode != 200 {
				data, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("%d %s", resp.StatusCode, data)
			}
			dom, err := goquery.NewDocumentFromReader(resp.Body)
			resp.Body.Close()
			if err != nil {
				c.logger.Error(err)
				return err
			}

			seli := dom.Find(`#thumbSlides .thumbSlide .pdp-image`)
			for j := range seli.Nodes {
				node := seli.Eq(j)
				murl := node.AttrOr("data-large", node.AttrOr("data-thumb", ""))
				if murl == "" {
					continue
				}
				medias = append(medias, pbMedia.NewImageMedia(
					strconv.Format(j),
					imgHeightTplReg.ReplaceAllString(imgWidthTplReg.ReplaceAllString(murl, ""), ""),
					imgHeightTplReg.ReplaceAllString(imgWidthTplReg.ReplaceAllString(murl, "&w=1000"), "&h=1000"),
					imgHeightTplReg.ReplaceAllString(imgWidthTplReg.ReplaceAllString(murl, "&w=700"), "&h=700"),
					imgHeightTplReg.ReplaceAllString(imgWidthTplReg.ReplaceAllString(murl, "&w=600"), "&h=600"),
					"",
					j == 0))
			}
		} else {
			seli := doc.Find(`#thumbSlides .thumbSlide .pdp-image`)
			for j := range seli.Nodes {
				node := seli.Eq(j)
				murl := node.AttrOr("data-large", node.AttrOr("data-thumb", ""))
				if murl == "" {
					continue
				}
				medias = append(medias, pbMedia.NewImageMedia(
					strconv.Format(j),
					imgHeightTplReg.ReplaceAllString(imgWidthTplReg.ReplaceAllString(murl, ""), ""),
					imgHeightTplReg.ReplaceAllString(imgWidthTplReg.ReplaceAllString(murl, "&w=1000"), "&h=1000"),
					imgHeightTplReg.ReplaceAllString(imgWidthTplReg.ReplaceAllString(murl, "&w=700"), "&h=700"),
					imgHeightTplReg.ReplaceAllString(imgWidthTplReg.ReplaceAllString(murl, "&w=600"), "&h=600"),
					"",
					j == 0))
			}
		}

		item.Medias = medias
		colorSpec := pbItem.SkuSpecOption{
			Type:  pbItem.SkuSpecType_SkuSpecColor,
			Id:    style,
			Name:  node.Find("a .color-image>img").AttrOr("alt", style),
			Value: style,
			Icon:  node.Find("a .color-image>img").AttrOr("data-src", ""),
		}

		priceSelNode := doc.Find(fmt.Sprintf(`#prices_%s .productPrice`, style))
		orgPrice, _ := strconv.ParsePrice(priceSelNode.Find(`.wasPrice`).Text())
		if orgPrice == 0 || math.IsNaN(orgPrice) {
			orgPrice, _ = strconv.ParsePrice(priceSelNode.Find(`.wasSeePrice`).Text())
		}
		currentPrice, _ := strconv.ParsePrice(priceSelNode.Find(`.nowPrice`).Text())
		if currentPrice == 0 || math.IsNaN(currentPrice) {
			currentPrice, _ = strconv.ParsePrice(priceSelNode.Find(`.wasSeePrice`).Text())
		}
		discount := float64(0)
		if currentPrice == 0 {
			currentPrice, _ = strconv.ParsePrice(priceSelNode.Find(`.fullPrice`).Text())
		}
		if orgPrice == 0 {
			orgPrice = currentPrice
		}
		if orgPrice != currentPrice {
			discount = math.Round((orgPrice - currentPrice) / orgPrice * 100)
		}

		// Note: Color variation is available on product list page therefor not considering multiple color of a product
		sizeSel := doc.Find(fmt.Sprintf(`#sizes_%s .sizeOptions`, style))
		for i := range sizeSel.Nodes {
			snode := sizeSel.Eq(i)

			skuId, _ := base64.RawStdEncoding.DecodeString(snode.AttrOr(`data-sku`, ""))
			if len(skuId) == 0 {
				c.logger.Errorf("invalid sku id %s", style)
				continue
			}
			sizeName := strings.TrimSpace(snode.Text())

			sku := pbItem.Sku{
				SourceId: string(skuId),
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(currentPrice * 100),
					Msrp:     int32(orgPrice * 100),
					Discount: int32(discount),
				},
				Medias: medias,
				Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}
			if !strings.Contains(snode.AttrOr("class", ""), "disabled") {
				sku.Stock.StockStatus = pbItem.Stock_InStock
				item.Stock.StockStatus = pbItem.Stock_InStock
			}

			sku.Specs = append(sku.Specs, &colorSpec)
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    sizeName,
				Name:  sizeName,
				Value: sizeName,
			})
			for _, option := range sku.GetSpecs() {
				sku.SourceId += fmt.Sprintf("-%s", option.GetId())
			}

			item.SkuItems = append(item.SkuItems, &sku)
		}
	}
	if err := yield(ctx, &item); err != nil {
		return err
	}
	return nil
}

func (c *_Crawler) NewTestRequest(ctx context.Context) []*http.Request {
	var reqs []*http.Request
	for _, u := range []string{
		//"https://www.finishline.com",
		// "https://www.finishline.com/store/only-sale-items/shoes/_/N-1g2z5sjZ1nc1fo0?icid=LP_sale_shoes50_PDCT",
		//"https://www.finishline.com/store/women/shoes/running/_/N-nat3jh?mnid=women_shoes_running",
		// "https://www.finishline.com/store/product/womens-nike-air-max-270-casual-shoes/prod2770847?styleId=AH6789&colorId=001",
		// "https://www.finishline.com/store/product/mens-nike-challenger-og-casual-shoes/prod2820864?styleId=CW7645&colorId=003",
		//"https://www.finishline.com/store/product/womens-puma-future-rider-play-on-casual-shoes/prod2795926?styleId=38182501&colorId=100",
		"https://www.finishline.com/store/product/big-kids-nike-air-force-1-low-casual-shoes/prod796065?styleId=314192&colorId=117",
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
