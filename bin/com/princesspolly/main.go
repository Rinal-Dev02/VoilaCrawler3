package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"math"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/voiladev/VoilaCrawler/pkg/cli"
	"github.com/voiladev/VoilaCrawler/pkg/context"
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
		httpClient:          client,
		categoryPathMatcher: regexp.MustCompile(`^/collections/[A-Za-z0-9_-]+$`),
		productPathMatcher:  regexp.MustCompile(`^(/collections/[A-Za-z0-9_-]+)?/products/[A-Za-z0-9_-]+$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "f883ca7266f298d2421898f8c18a7cd9"
}

// Version
func (c *_Crawler) Version() int32 {
	return 1
}

// CrawlOptions
func (c *_Crawler) CrawlOptions(u *url.URL) *crawler.CrawlOptions {
	options := crawler.NewCrawlOptions()
	options.EnableHeadless = false
	options.EnableSessionInit = true

	options.Reliability = pbProxy.ProxyReliability_ReliabilityMedium
	options.MustCookies = append(options.MustCookies,
		&http.Cookie{Name: "ckm-ctx-sf", Value: `%2F`, Path: "/"},
	)
	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"us.princesspolly.com"}
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
		u.Host = "us.princesspolly.com"
	}
	if c.productPathMatcher.MatchString(u.Path) {
		u.RawQuery = ""
		if i := strings.Index(u.Path, "/products"); i != 0 {
			u.Path = u.Path[i:]
		}
		return u.String(), nil
	}
	return rawurl, nil
}

func getPathFirstSection(p string) string {
	return "/" + strings.SplitN(strings.TrimPrefix(p, "/"), "/", 2)[0]
}

func (c *_Crawler) Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	if resp.RawUrl().Path == "" || resp.RawUrl().Path == "/" {
		return c.parseCategories(ctx, resp, yield)
	}
	if c.productPathMatcher.MatchString(resp.RawUrl().Path) {
		return c.parseProduct(ctx, resp, yield)
	} else if c.categoryPathMatcher.MatchString(resp.RawUrl().Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	}
	return crawler.ErrUnsupportedPath
}

func (c *_Crawler) parseCategories(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	dom, err := resp.Selector()
	if err != nil {
		c.logger.Error(err)
		return err
	}

	sel := dom.Find(`.header__nav .nav .nav__list>.nav__item`)
	if len(sel.Nodes) == 0 {
		return fmt.Errorf("got no categories")
	}

	urlFilter := map[string]struct{}{}
	for i := range sel.Nodes {
		node := sel.Eq(i)
		cate := strings.TrimSpace(node.Find(`a.nav__link`).First().Text())
		subSel := node.Find(`.nav__list .nav__list-items .nav__item .nav__list>.nav__item>a.nav__link`)

		pctx := context.WithValue(ctx, "item.index", 0)
		pctx = context.WithValue(pctx, "Category", cate)
		for j := range subSel.Nodes {
			subNode := subSel.Eq(j)
			href := subNode.AttrOr("href", "")
			if href == "" {
				continue
			}
			// ignore pages
			u, err := url.Parse(href)
			if err != nil {
				c.logger.Errorf("invalid url %s", href)
				continue
			}
			if strings.HasPrefix(u.Path, "/pages/") {
				continue
			}

			if _, ok := urlFilter[href]; ok {
				continue
			}

			urlFilter[href] = struct{}{}
			subCate := strings.TrimSpace(subNode.Text())

			req, err := http.NewRequest(http.MethodGet, href, nil)
			if err != nil {
				c.logger.Errorf("load request for %s failed, error=%s", href, err)
				continue
			}
			sctx := context.WithValue(pctx, "SubCategory", subCate)
			if err := yield(sctx, req); err != nil {
				return err
			}
		}
	}
	return nil
}

// nextIndex used to get sharingData from context
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")) + 1)
}

type filterProducts struct {
	TotalProduct    int64 `json:"total_product"`
	TotalCollection int   `json:"total_collection"`
	TotalPage       int   `json:"total_page"`
	FromCache       bool  `json:"from_cache"`
	Products        []struct {
		Handle string `json:"handle"`
	} `json:"products"`
}

type productListType struct {
	Results struct {
		Goods []struct {
			Index        int `json:"index"`
			PretreatInfo struct {
				GoodsName             string `json:"goodsName"`
				SeriesOrBrandAnalysis string `json:"seriesOrBrandAnalysis"`
				GoodsDetailURL        string `json:"goodsDetailUrl"`
			} `json:"pretreatInfo"`
		} `json:"goods"`
		CatInfo struct {
			//Page        int    `json:"page"`
			Limit       int    `json:"limit"`
			OriginalURL string `json:"originalUrl"`
		} `json:"cat_info"`
		Sum int `json:"sum"`
	} `json:"results"`
}

var (
	shopifyCollIdReg = regexp.MustCompile(`(?U)window.shopifyCollection\s*=\s*({.*});`)
	collIdReg        = regexp.MustCompile(`collection_id:\s*(\d+),`)
	filterDataReg    = regexp.MustCompile(`BCSfFilterCallback\(({.*})\);\s*$`)
)

// parseCategoryProducts parse api url from web page url
func (c *_Crawler) parseCategoryProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	respBody, err := resp.RawBody()
	if err != nil {
		c.logger.Debug(err)
		return err
	}

	doc, err := resp.Selector()
	if err != nil {
		return err
	}

	sel := doc.Find(`.product-tile__name`)
	lastIndex := nextIndex(ctx)

	if len(sel.Nodes) == 0 {
		query := resp.CurrentUrl().Query()
		page, _ := strconv.ParseInt(query.Get("page"))
		if page == 0 {
			page = 1
		}
		query.Set("page", strconv.Format(page))
		queryStr := ""
		for k, vals := range query {
			for _, v := range vals {
				if k == "pf_opt_size" || k == "pf_opt_color" || k == "pf_v_brand" {
					k = k + "[]"
				}
				kv := fmt.Sprintf("%s=%s", url.QueryEscape(k), url.QueryEscape(v))
				if queryStr == "" {
					queryStr = kv
				} else {
					queryStr = queryStr + "&" + kv
				}
			}
		}

		collectionId := ""
		matched := shopifyCollIdReg.FindSubmatch(respBody)
		if len(matched) > 0 {
			var coll struct {
				Id int64 `json:"id"`
			}
			if err := json.Unmarshal(matched[1], &coll); err != nil {
				c.logger.Errorf("parse %s failed", matched[1])
			}
			collectionId = strconv.Format(coll.Id)
		}
		if collectionId == "" {
			if matched := collIdReg.FindSubmatch(respBody); len(matched) > 0 {
				collectionId = string(matched[1])
			}
		}

		if collectionId == "" {
			return fmt.Errorf("no collection id found")
		}

		const pageLimit = 60
		// check filter
		u := fmt.Sprintf("https://services.mybcapps.com/bc-sf-filter/filter?t=%d&%s&shop=princesspollydev.myshopify.com&currency=usd&limit=%d&display=grid&collection_scope=%s&product_available=true&variant_available=true&build_filter_tree=true&check_cache=false&callback=BCSfFilterCallback&event_type=init",
			time.Now().UnixNano()/1000000,
			queryStr,
			pageLimit,
			collectionId,
		)
		req, err := http.NewRequest(http.MethodGet, u, nil)
		if err != nil {
			c.logger.Error(err)
			return err
		}
		req.Header.Set("referer", "https://us.princesspolly.com/")
		opts := c.CrawlOptions(req.URL)
		apiResp, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
			EnableProxy: true,
			Reliability: opts.Reliability,
		})
		if err != nil {
			c.logger.Error(err)
			return err
		}
		respBody, err := apiResp.RawBody()
		if err != nil {
			c.logger.Error(err)
			return err
		}

		var filterResp filterProducts
		if matched := filterDataReg.FindSubmatch(respBody); len(matched) == 0 {
			return fmt.Errorf("get filter product data fialed")
		} else if err = json.Unmarshal(matched[1], &filterResp); err != nil {
			c.logger.Errorf("parse filtered products failed, error=%s", err)
			return err
		}
		for _, prod := range filterResp.Products {
			u := fmt.Sprintf("/products/%s", prod.Handle)
			req, err := http.NewRequest(http.MethodGet, u, nil)
			if err != nil {
				c.logger.Error(err)
				continue
			}

			nctx := context.WithValue(ctx, "item.index", lastIndex)
			lastIndex += 1

			if err := yield(nctx, req); err != nil {
				return err
			}
		}
		if page*pageLimit < filterResp.TotalProduct {
			// next page
			nextUrl := *resp.RawUrl()
			vals := nextUrl.Query()
			vals.Set("page", strconv.Format(page+1))
			nextUrl.RawQuery = vals.Encode()

			req, _ := http.NewRequest(http.MethodGet, nextUrl.String(), nil)
			nctx := context.WithValue(ctx, "item.index", lastIndex)
			return yield(nctx, req)
		}
	} else {
		for i := range sel.Nodes {
			node := sel.Eq(i)
			href := node.AttrOr("href", "")
			if href == "" {
				continue
			}
			//fmt.Println(href)
			req, err := http.NewRequest(http.MethodGet, href, nil)
			if err != nil {
				c.logger.Error(err)
				continue
			}

			nctx := context.WithValue(ctx, "item.index", lastIndex)
			lastIndex += 1

			if err := yield(nctx, req); err != nil {
				return err
			}
		}

		// get current page number
		page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("page"))
		if page == 0 {
			page = 1
		}

		if bytes.Contains(respBody, []byte(`class="pagination__button pagination__button--next pagination__button--disabled"`)) {
			return nil
		}

		nextUrl := html.UnescapeString(doc.Find(`.pagination__button.pagination__button--next`).AttrOr("href", ""))
		if nextUrl == "" {
			if nextUrl == "" || strings.ToLower(nextUrl) == "null" {
				return nil
			}
		}
		req, _ := http.NewRequest(http.MethodGet, nextUrl, nil)
		nctx := context.WithValue(ctx, "item.index", lastIndex)
		return yield(nctx, req)
	}
	return nil
}

type parseProductResponse struct {
	ID                   int64    `json:"id"`
	Title                string   `json:"title"`
	Handle               string   `json:"handle"`
	Description          string   `json:"description"`
	PublishedAt          string   `json:"published_at"`
	CreatedAt            string   `json:"created_at"`
	Vendor               string   `json:"vendor"`
	Type                 string   `json:"type"`
	Tags                 []string `json:"tags"`
	Price                int      `json:"price"`
	PriceMin             int      `json:"price_min"`
	PriceMax             int      `json:"price_max"`
	Available            bool     `json:"available"`
	PriceVaries          bool     `json:"price_varies"`
	CompareAtPrice       int      `json:"compare_at_price"`
	CompareAtPriceMin    int      `json:"compare_at_price_min"`
	CompareAtPriceMax    int      `json:"compare_at_price_max"`
	CompareAtPriceVaries bool     `json:"compare_at_price_varies"`
	Variants             []struct {
		ID                     int64         `json:"id"`
		Title                  string        `json:"title"`
		Option1                string        `json:"option1"`
		Option2                string        `json:"option2"`
		Option3                interface{}   `json:"option3"`
		Sku                    string        `json:"sku"`
		RequiresShipping       bool          `json:"requires_shipping"`
		Taxable                bool          `json:"taxable"`
		FeaturedImage          interface{}   `json:"featured_image"`
		Available              bool          `json:"available"`
		Name                   string        `json:"name"`
		PublicTitle            string        `json:"public_title"`
		Options                []string      `json:"options"`
		Price                  int           `json:"price"`
		Weight                 int           `json:"weight"`
		CompareAtPrice         int           `json:"compare_at_price"`
		InventoryManagement    string        `json:"inventory_management"`
		Barcode                string        `json:"barcode"`
		RequiresSellingPlan    bool          `json:"requires_selling_plan"`
		SellingPlanAllocations []interface{} `json:"selling_plan_allocations"`
	} `json:"variants"`
	Images        []string `json:"images"`
	FeaturedImage string   `json:"featured_image"`
	Options       []string `json:"options"`
	Media         []struct {
		Alt          interface{} `json:"alt"`
		ID           int64       `json:"id"`
		Position     int         `json:"position"`
		PreviewImage struct {
			AspectRatio float64 `json:"aspect_ratio"`
			Height      int     `json:"height"`
			Width       int     `json:"width"`
			Src         string  `json:"src"`
		} `json:"preview_image"`
		AspectRatio float64 `json:"aspect_ratio"`
		Height      int     `json:"height"`
		MediaType   string  `json:"media_type"`
		Src         string  `json:"src"`
		Width       int     `json:"width"`
	} `json:"media"`
	RequiresSellingPlan bool          `json:"requires_selling_plan"`
	SellingPlanGroups   []interface{} `json:"selling_plan_groups"`
	Content             string        `json:"content"`
}

type parseProductReviewRating struct {
	Context         string `json:"@context"`
	Type            string `json:"@type"`
	AggregateRating struct {
		Type        string `json:"@type"`
		RatingValue string `json:"ratingValue"`
		ReviewCount string `json:"reviewCount"`
	} `json:"aggregateRating"`
	Name string `json:"name"`
}

var (
	detailReg    = regexp.MustCompile(`window.SwymProductInfo.product\s*=\s*({.*});`)
	htmlTrimRegp = regexp.MustCompile(`</?[^>]+>`)
)

func (c *_Crawler) parseProduct(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	respBody, err := resp.RawBody()
	if err != nil {
		c.logger.Error(err)
		return err
	}

	matched := detailReg.FindSubmatch(respBody)
	if len(matched) == 0 {
		c.httpClient.Jar().Clear(ctx, resp.Request.URL)
		return fmt.Errorf("extract produt json from page %s content failed", resp.Request.URL)
	}

	var (
		viewData parseProductResponse
	)
	if err = json.Unmarshal(bytes.TrimRight(matched[1], ","), &viewData); err != nil {
		c.logger.Error(err)
		return err
	}

	doc, err := resp.Selector()
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
			Id:           strconv.Format(viewData.ID),
			CrawlUrl:     resp.Request.URL.String(),
			CanonicalUrl: canUrl,
			GroupId:      "", // TODO: can't found any group id
		},
		Title:       viewData.Title,
		Description: htmlTrimRegp.ReplaceAllString(html.UnescapeString(viewData.Description), " "),
		BrandName:   doc.Find(`.brand`).Text(),
		CrowdType:   "",
		Price: &pbItem.Price{
			Currency: regulation.Currency_USD,
		},
		Stats: &pbItem.Stats{},
	}

	var viewReviewData parseProductReviewRating
	sel := doc.Find(`.y-rich-snippet-script`)
	if len(sel.Nodes) == 0 {
		// TODO
		// fetch https://staticw2.yotpo.com/Qj1vQ6MEldiPQnTrmCa5J9ksqGmTvGR3AIyFz5h2/widget.js
		// to get widget version with Yotpo.version = '2020-11-09_14-45-57';
		const version = "2020-11-09_14-45-57"

		appKeySel := doc.Find(`#yotpo-stars`)
		appKey := appKeySel.AttrOr("data-appkey", "Qj1vQ6MEldiPQnTrmCa5J9ksqGmTvGR3AIyFz5h2")

		u := fmt.Sprintf("https://staticw2.yotpo.com/batch/app_key/%s/domain_key/%s/widget/rich_snippet", appKey, item.Source.Id)
		vals := url.Values{}
		vals.Set("methods",
			fmt.Sprintf(`[{"method":"rich_snippet","params":{"pid":"%s","name":"%s","price":null,"currency":null}}]`,
				item.Source.Id, item.Title))
		vals.Set("app_key", appKey)
		vals.Set("is_mobile", "false")
		vals.Set("widget_version", version)

		body := vals.Encode()
		if req, err := http.NewRequest(http.MethodPost, u, bytes.NewReader([]byte(body))); err != nil {
			c.logger.Errorf("build request for url %s failed", u)
		} else {
			opts := c.CrawlOptions(req.URL)
			nctx, cancel := context.WithTimeout(ctx, time.Minute)
			defer cancel()

			req.Header.Set("referer", "https://us.princesspolly.com/")
			req.Header.Set("content-type", "application/x-www-form-urlencoded")
			req.Header.Set("origin", "https://us.princesspolly.com")
			// NOTE: cookie is set in auto

			if snipResp, err := c.httpClient.DoWithOptions(nctx, req, http.Options{
				EnableProxy: true,
				Reliability: opts.Reliability,
			}); err != nil {
				c.logger.Errorf("get rich snippet failed, error=%s", err)
			} else if body, err := snipResp.RawBody(); err == nil {
				var richResp []struct {
					Method string `json:"method"`
					Result string `json:"result"`
				}
				if err := json.Unmarshal(body, &richResp); err != nil {
					c.logger.Errorf("unmarshal %s failed, error=%s", body, err)
				} else if len(richResp) > 0 && richResp[0].Result != "" {
					if dom, err := goquery.NewDocumentFromReader(strings.NewReader(richResp[0].Result)); err == nil {
						sel = dom.Find(`.y-rich-snippet-script`)
					} else {
						c.logger.Errorf("build html dom failed, error=%s", err)
					}
				}
			} else {
				c.logger.Error("load sub response failed")
			}
		}
	}
	if strings.TrimSpace(sel.Text()) != "" {
		if err = json.Unmarshal([]byte(strings.TrimSpace(sel.Text())), &viewReviewData); err != nil {
			c.logger.Error(err)
			return err
		}
		item.Stats.ReviewCount, _ = strconv.ParseInt32(viewReviewData.AggregateRating.ReviewCount)
		item.Stats.Rating, _ = strconv.ParseFloat32(viewReviewData.AggregateRating.RatingValue)
	}

	item.Category = context.GetString(ctx, "Category")
	item.SubCategory = context.GetString(ctx, "SubCategory")
	item.SubCategory2 = context.GetString(ctx, "SubCategory2")
	item.SubCategory3 = context.GetString(ctx, "SubCategory3")
	item.SubCategory4 = context.GetString(ctx, "SubCategory4")

	var medias []*media.Media
	for i, img := range viewData.Media {
		u, err := url.Parse(img.Src)
		if err != nil {
			return fmt.Errorf("parse img url %s failed", img.Src)
		}
		fields := strings.Split(u.Path, ".")
		tpl := strings.Join(fields[0:len(fields)-1], ".") + "%s." + fields[len(fields)-1]
		itemImg, _ := anypb.New(&media.Media_Image{
			OriginalUrl: img.Src,
			LargeUrl:    img.Src,
			MediumUrl:   strings.ReplaceAll(img.Src, u.Path, fmt.Sprintf(tpl, "_600x")),
			SmallUrl:    strings.ReplaceAll(img.Src, u.Path, fmt.Sprintf(tpl, "_500x")),
		})
		medias = append(medias, &media.Media{
			Detail:    itemImg,
			IsDefault: i == 0,
		})
	}
	item.Medias = medias

	optionSizeIndex := -1
	optionColorIndex := -1
	for i, rawSize := range viewData.Options {
		if strings.ToLower(rawSize) == "size" {
			optionSizeIndex = i
		} else if strings.ToLower(rawSize) == "color" {
			optionColorIndex = i
		}
	}

	for _, rawSize := range viewData.Variants {
		current, _ := strconv.ParseFloat(rawSize.Price)
		msrp, _ := strconv.ParseFloat(rawSize.CompareAtPrice)
		if msrp == 0 {
			msrp = current
		}
		discount := 0.0
		if msrp > current {
			discount = math.Round(((msrp - current) / msrp) * 100)
		}

		sku := pbItem.Sku{
			SourceId: strconv.Format(rawSize.Sku),
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
				Current:  int32(current),
				Msrp:     int32(msrp),
				Discount: int32(discount),
			},
			Medias: medias,
			Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
		}
		if rawSize.Available {
			sku.Stock.StockStatus = pbItem.Stock_InStock
		}

		if optionColorIndex > -1 {
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecColor,
				Id:    strconv.Format(rawSize.ID),
				Name:  rawSize.Options[optionColorIndex],
				Value: rawSize.Options[optionColorIndex],
			})
		}

		if optionSizeIndex > -1 {
			sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
				Type:  pbItem.SkuSpecType_SkuSpecSize,
				Id:    strconv.Format(rawSize.Barcode),
				Name:  rawSize.Options[optionSizeIndex],
				Value: rawSize.Options[optionSizeIndex],
			})
		}
		item.SkuItems = append(item.SkuItems, &sku)
	}

	if err = yield(ctx, &item); err != nil {
		return err
	}
	return nil
}

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://us.princesspolly.com/collections/basics",
		//"https://us.princesspolly.com/collections/graphic-tees",
		"https://us.princesspolly.com/products/innerbloom-jumper-pink",
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
