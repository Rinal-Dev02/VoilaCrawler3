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
	"github.com/voiladev/VoilaCrawler/pkg/cli"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	pbMedia "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
)

type _Crawler struct {
	httpClient http.Client

	categoryPathMatcher     *regexp.Regexp
	categoryJsonPathMatcher *regexp.Regexp
	productGroupPathMatcher *regexp.Regexp
	productPathMatcher      *regexp.Regexp
	logger                  glog.Log
}

func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient: client,
		// /men-bags/COjWAcABAuICAgEY.zso
		// https://www.zappos.com/alex-and-ani-cross-ii-32-expandable-necklace?oosRedirected=true
		categoryPathMatcher: regexp.MustCompile(`^(/[a-z0-9\-]+){1,5}/[a-zA-Z0-9]+\.zso$`),
		productPathMatcher:  regexp.MustCompile(`^(/a/[a-z0-0-]+)?/p(/[a-z0-9_-]+)/product/\d+(/[a-z0-9]+/\d+)?$`),
		logger:              logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "0638b7ff6993f941d9d2ebf80e7166c8"
}

// Version
func (c *_Crawler) Version() int32 {
	return 1
}

// CrawlOptions
func (c *_Crawler) CrawlOptions(u *url.URL) *crawler.CrawlOptions {
	options := crawler.NewCrawlOptions()
	options.EnableHeadless = false
	options.MustCookies = append(options.MustCookies,
		&http.Cookie{Name: "geo", Value: "US/CA/803/LOSANGELES"},
		&http.Cookie{Name: "clouddc", Value: "west1"},
	)
	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.zappos.com"}
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
		u.Host = "www.zappos.com"
	}
	if c.productPathMatcher.MatchString(u.Path) {
		index := strings.Index(u.Path, "/color/")
		if index > 0 {
			u.Path = u.Path[0:index]
		}
		u.RawQuery = ""
		return u.String(), nil
	}
	return rawurl, nil
}

func (c *_Crawler) Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	// product url redirect to a out of stock page
	if resp.Request.URL.Query().Get("oosRedirected") == "true" {
		return crawler.ErrAbort
	}
	p := strings.TrimSuffix(resp.Request.URL.Path, "/")

	if p == "" {
		return c.parseCategories(ctx, resp, yield)
	}
	if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yield)
	} else if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yield)
	}
	return crawler.ErrUnsupportedPath
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

	sel := dom.Find(`.hfHeaderNav`).Find(`li`)

	for i := range sel.Nodes {
		node := sel.Eq(i)
		cateName := strings.TrimSpace(node.Find(`a[data-shyguy]`).First().Text())
		if cateName == "" {
			continue
		}
		nnctx := context.WithValue(ctx, "Category", cateName)
		//fmt.Println(`cateName `, cateName)

		subSel1 := node.Find(`div[data-headercategory]`).Find(`section`)
		for k := range subSel1.Nodes {
			subNodeN := subSel1.Eq(k)
			subCat1 := strings.TrimSpace(subNodeN.Find(`a[data-hfsubnav]`).Text())
			fmt.Println(subCat1)
			subSel := subNodeN.Find(`li`)
			for j := range subSel.Nodes {

				subNode := subSel.Eq(j)
				href := subNode.Find(`a`).AttrOr("href", "")
				if href == "" {
					continue
				}

				_, err := url.Parse(href)
				if err != nil {
					c.logger.Error("parse url %s failed", href)
					continue
				}

				subCateName := subCat1 + " > " + strings.TrimSpace(subNode.Text())
				//fmt.Println(subCateName)
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

// nextIndex used to get sharingData from context
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")) + 1)
}

var (
	productsExtractReg  = regexp.MustCompile(`(?U)<script\s*>window.__INITIAL_STATE__\s*=\s*({.*});?\s*</script>`)
	productQuickViewReg = regexp.MustCompile(`(?Ums)\(new window\.tsr\.quickview\(\s*({.*})\s*\)\s*\)\.initQuickview\(null,\s*true\);?`)
)

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

	matched := productsExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		return fmt.Errorf("extract json from product list page %s failed", resp.Request.URL)
	}

	var r struct {
		Filters struct {
			Page      int64 `json:"page"`
			PageCount int64 `json:"pageCount"`
		} `json:"filters"`
		Meta struct {
			DocumentMeta struct {
				Link struct {
					Rel struct {
						Next string `json:"next"`
					} `json:"rel"`
				} `json:"link"`
			} `json:"documentMeta"`
		} `json:"meta"`
		Products struct {
			TotalProductCount int           `json:"totalProductCount"`
			IsLoading         bool          `json:"isLoading"`
			RequestedURL      string        `json:"requestedUrl"`
			ExecutedSearchURL string        `json:"executedSearchUrl"`
			NoResultsRecos    []interface{} `json:"noResultsRecos"`
			InlineRecos       interface{}   `json:"inlineRecos"`
			OosMessaging      interface{}   `json:"oosMessaging"`
			HeartsList        struct {
			} `json:"heartsList"`
			ProductLimit int `json:"productLimit"`
			List         []struct {
				Sizing struct {
				} `json:"sizing"`
				ProductID   string `json:"productId"`
				ProductName string `json:"productName"`
				ProductURL  string `json:"productUrl"`
			} `json:"list"`
			AllProductsCount int `json:"allProductsCount"`
		} `json:"products"`
	}

	if err = json.Unmarshal(matched[1], &r); err != nil {
		c.logger.Debugf("parse %s failed, error=%s", matched[1], err)
		return err
	}

	lastIndex := nextIndex(ctx)
	for _, prod := range r.Products.List {
		if prod.ProductURL == "" {
			continue
		}
		req, err := http.NewRequest(http.MethodGet, prod.ProductURL, nil)
		if err != nil {
			c.logger.Error(err)
			continue
		}

		nctx := context.WithValue(ctx, "item.index", lastIndex)
		lastIndex += 1
		if err = yield(nctx, req); err != nil {
			return err
		}
	}

	// get current page number
	page, _ := strconv.ParseInt(resp.Request.URL.Query().Get("p"))

	// check if this is the last page
	totalpages := r.Filters.PageCount
	if page >= totalpages || r.Meta.DocumentMeta.Link.Rel.Next == "" {
		return nil
	}

	req, _ := http.NewRequest(http.MethodGet, r.Meta.DocumentMeta.Link.Rel.Next, nil)
	nctx := context.WithValue(ctx, "item.index", lastIndex)

	return yield(nctx, req)

}

type productStructure struct {
	PixelServer struct {
		Data struct {
			CustomerCountryCode string `json:"customerCountryCode"`
			PageID              string `json:"pageId"`
			PageLang            string `json:"pageLang"`
			PageTitle           bool   `json:"pageTitle"`
			Product             struct {
				Sku         string `json:"sku"`
				StyleID     string `json:"styleId"`
				Price       string `json:"price"`
				Name        string `json:"name"`
				Brand       string `json:"brand"`
				Category    string `json:"category"`
				SubCategory string `json:"subCategory"`
				Gender      string `json:"gender"`
			} `json:"product"`
		} `json:"data"`
		PageType    string `json:"pageType"`
		QueryString string `json:"queryString"`
	} `json:"pixelServer"`
	Product struct {
		ReviewData struct {
			SubmittedReviews []interface{} `json:"submittedReviews"`
			LoadingReviews   []interface{} `json:"loadingReviews"`
		} `json:"reviewData"`
		Detail struct {
			ReviewSummary struct {
				ReviewWithMostVotes      interface{} `json:"reviewWithMostVotes"`
				ReviewWithLeastVotes     interface{} `json:"reviewWithLeastVotes"`
				TotalCriticalReviews     string      `json:"totalCriticalReviews"`
				TotalFavorableReviews    string      `json:"totalFavorableReviews"`
				TotalReviews             string      `json:"totalReviews"`
				TotalReviewScore         interface{} `json:"totalReviewScore"`
				AverageOverallRating     string      `json:"averageOverallRating"`
				ArchRatingPercentages    interface{} `json:"archRatingPercentages"`
				OverallRatingPercentages struct {
					Num5 string `json:"5"`
				} `json:"overallRatingPercentages"`
				SizeRatingPercentages      interface{} `json:"sizeRatingPercentages"`
				WidthRatingPercentages     interface{} `json:"widthRatingPercentages"`
				MaxArchRatingPercentage    interface{} `json:"maxArchRatingPercentage"`
				MaxOverallRatingPercentage struct {
					Percentage string `json:"percentage"`
					Text       string `json:"text"`
				} `json:"maxOverallRatingPercentage"`
				MaxSizeRatingPercentage  interface{} `json:"maxSizeRatingPercentage"`
				MaxWidthRatingPercentage interface{} `json:"maxWidthRatingPercentage"`
				ReviewingAShoe           string      `json:"reviewingAShoe"`
				HasFitRatings            string      `json:"hasFitRatings"`
				AggregateRating          float64     `json:"aggregateRating"`
			} `json:"reviewSummary"`

			Videos             []interface{} `json:"videos"`
			Genders            []string      `json:"genders"`
			DefaultProductURL  string        `json:"defaultProductUrl"`
			DefaultImageURL    string        `json:"defaultImageUrl"`
			DefaultSubCategory interface{}   `json:"defaultSubCategory"`
			PreferredSubsite   interface{}   `json:"preferredSubsite"`
			OverallRating      []string      `json:"overallRating"`
			ProductName        string        `json:"productName"`
			ProductRating      string        `json:"productRating"`
			SizeFit            struct {
				Text       string `json:"text"`
				Percentage string `json:"percentage"`
			} `json:"sizeFit"`
			Description struct {
				BulletPoints []string      `json:"bulletPoints"`
				SizeCharts   []interface{} `json:"sizeCharts"`
			} `json:"description"`
			Zombie             bool   `json:"zombie"`
			BrandID            string `json:"brandId"`
			DefaultProductType string `json:"defaultProductType"`
			DefaultCategory    string `json:"defaultCategory"`
			ReviewCount        string `json:"reviewCount"`
			Styles             []struct {
				StyleDescription interface{} `json:"styleDescription"`
				Color            string      `json:"color"`
				OriginalPrice    string      `json:"originalPrice"`
				PercentOff       string      `json:"percentOff"`
				Price            string      `json:"price"`
				ProductURL       string      `json:"productUrl"`
				Stocks           []struct {
					OnHand  string `json:"onHand"`
					StockID string `json:"stockId"`
					SizeID  string `json:"sizeId"`
					Upc     string `json:"upc"`
					Size    string `json:"size"`
					Width   string `json:"width"`
					Asin    string `json:"asin"`
				} `json:"stocks"`
				HardLaunchDate interface{} `json:"hardLaunchDate"`
				OnSale         string      `json:"onSale"`
				StyleID        string      `json:"styleId"`
				ImageURL       string      `json:"imageUrl"`
				ColorID        string      `json:"colorId"`
				ProductID      string      `json:"productId"`
				IsNew          string      `json:"isNew"`
				Images         []struct {
					ImageID string `json:"imageId"`
					Type    string `json:"type"`
				} `json:"images"`
				SwatchImageID string `json:"swatchImageId"`
				Drop          bool   `json:"drop"`
				FinalSale     bool   `json:"finalSale"`
				TsdImages     struct {
				} `json:"tsdImages"`
				ImageID string `json:"imageId"`
			} `json:"styles"`
			VideoURL  interface{} `json:"videoUrl"`
			ProductID string      `json:"productId"`
			ArchFit   struct {
				Text       string `json:"text"`
				Percentage string `json:"percentage"`
			} `json:"archFit"`
			BrandName string `json:"brandName"`
			WidthFit  struct {
				Percentage string `json:"percentage"`
				Text       string `json:"text"`
			} `json:"widthFit"`
			Oos   bool `json:"oos"`
			Brand struct {
				ID             string `json:"id"`
				Name           string `json:"name"`
				CleanName      string `json:"cleanName"`
				BrandURL       string `json:"brandUrl"`
				ImageURL       string `json:"imageUrl"`
				HeaderImageURL string `json:"headerImageUrl"`
				RealBrandID    string `json:"realBrandId"`
			} `json:"brand"`
			BrandProductName      string      `json:"brandProductName"`
			IsReviewableWithMedia bool        `json:"isReviewableWithMedia"`
			HasHalfSizes          interface{} `json:"hasHalfSizes"`
			ReceivedDescription   string      `json:"receivedDescription"`
			YoutubeData           struct {
			} `json:"youtubeData"`
		} `json:"detail"`
		ReviewsTotalPages int    `json:"reviewsTotalPages"`
		SeoProductURL     string `json:"seoProductUrl"`
		CalledClientSide  bool   `json:"calledClientSide"`
	} `json:"product"`
}

type productQuickViewStructure struct {
	ReviewSummary struct {
		TotalFavorableReviews   string      `json:"totalFavorableReviews"`
		TotalReviews            string      `json:"totalReviews"`
		SizeRatingPercentages   interface{} `json:"sizeRatingPercentages"`
		WidthRatingPercentages  interface{} `json:"widthRatingPercentages"`
		MaxArchRatingPercentage interface{} `json:"maxArchRatingPercentage"`
		AverageOverallRating    string      `json:"averageOverallRating"`
	} `json:"reviewSummary"`
	Genders            []string `json:"genders"`
	Description        string   `json:"description"`
	DefaultCategory    string   `json:"defaultCategory"`
	ProductRating      string   `json:"productRating"`
	DefaultSubCategory string   `json:"defaultSubCategory"`
	DefaultImageURL    string   `json:"defaultImageUrl"`
	BrandID            string   `json:"brandId"`
	Zombie             bool     `json:"zombie"`
	DefaultProductType string   `json:"defaultProductType"`
	ProductID          string   `json:"productId"`
	BrandName          string   `json:"brandName"`
	ReviewCount        string   `json:"reviewCount"`
	OverallRating      struct {
	} `json:"overallRating"`
	DefaultProductURL string `json:"defaultProductUrl"`
	ProductName       string `json:"productName"`
	Styles            []struct {
		StyleDescription interface{} `json:"styleDescription"`
		Color            string      `json:"color"`
		OriginalPrice    string      `json:"originalPrice"`
		PercentOff       string      `json:"percentOff"`
		Price            string      `json:"price"`
		ProductURL       string      `json:"productUrl"`
		Stocks           []struct {
			OnHand  string `json:"onHand"`
			StockID string `json:"stockId"`
			Upc     string `json:"upc"`
			SizeID  string `json:"sizeId"`
			Width   string `json:"width"`
			Size    string `json:"size"`
			Asin    string `json:"asin"`
		} `json:"stocks"`
		ProductID string `json:"productId"`
		ImageURL  string `json:"imageUrl"`
		OnSale    string `json:"onSale"`
		StyleID   string `json:"styleId"`
		ColorID   string `json:"colorId"`
		IsNew     string `json:"isNew"`
		Images    []struct {
			ImageID string `json:"imageId"`
			Type    string `json:"type"`
		} `json:"images"`
		Oos      bool `json:"oos"`
		Features struct {
		} `json:"features"`
	} `json:"styles"`
	Color    string `json:"color"`
	ColorID  string `json:"colorId"`
	Features struct {
	} `json:"features"`
	OriginalPrice string `json:"originalPrice"`
	PercentOff    string `json:"percentOff"`
	Price         string `json:"price"`
	StyleID       string `json:"styleId"`
}

// parseProduct
func (c *_Crawler) parseProduct(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil {
		return nil
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	matched := productsExtractReg.FindSubmatch(respBody)
	qvMatched := productQuickViewReg.FindSubmatch(respBody)

	if len(matched) > 1 {
		var viewData productStructure
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
		reviews, _ := strconv.ParseInt(viewData.Product.Detail.ReviewSummary.TotalReviews)
		// build product data
		item := pbItem.Product{
			Source: &pbItem.Source{
				Id:           strconv.Format(viewData.Product.Detail.ProductID),
				CrawlUrl:     resp.Request.URL.String(),
				CanonicalUrl: canUrl,
			},
			BrandName:   viewData.Product.Detail.BrandName,
			Title:       viewData.Product.Detail.ProductName,
			Description: viewData.Product.Detail.ReceivedDescription,
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
			},
			CrowdType:   viewData.PixelServer.Data.Product.Gender,
			Category:    viewData.PixelServer.Data.Product.Category,
			SubCategory: viewData.PixelServer.Data.Product.SubCategory,
			Stats: &pbItem.Stats{
				ReviewCount: int32(reviews),
				Rating:      float32(viewData.Product.Detail.ReviewSummary.AggregateRating),
			},
		}

		for _, rawSku := range viewData.Product.Detail.Styles {
			originalPrice, _ := strconv.ParsePrice(rawSku.OriginalPrice)
			price, _ := strconv.ParsePrice(rawSku.Price)
			discount, _ := strconv.ParseInt(strings.TrimSuffix(rawSku.PercentOff, "%"))

			var medias []*pbMedia.Media
			for j, m := range rawSku.Images {
				medias = append(medias, pbMedia.NewImageMedia(
					m.ImageID,
					fmt.Sprintf("https://m.media-amazon.com/images/I/%s.jpg", m.ImageID),
					fmt.Sprintf("https://m.media-amazon.com/images/I/%s._AC_SR1000,750_.jpg", m.ImageID),
					fmt.Sprintf("https://m.media-amazon.com/images/I/%s._AC_SR700,525_.jpg", m.ImageID),
					fmt.Sprintf("https://m.media-amazon.com/images/I/%s._AC_SR500,375_.jpg", m.ImageID),
					"",
					j == 0,
				))
			}

			for _, stock := range rawSku.Stocks {
				sku := pbItem.Sku{
					SourceId: rawSku.StyleID,
					Price: &pbItem.Price{
						Currency: regulation.Currency_USD,
						Current:  int32(price * 100),
						Msrp:     int32(originalPrice * 100),
						Discount: int32(discount),
					},
					Medias: medias,
					Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
				}
				onhandCount, _ := strconv.ParseInt(stock.OnHand)
				if onhandCount > 0 {
					sku.Stock.StockStatus = pbItem.Stock_InStock
					sku.Stock.StockCount = int32(onhandCount)
				}

				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecColor,
					Id:    strconv.Format(rawSku.ColorID),
					Name:  rawSku.Color,
					Value: rawSku.Color,
				})

				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecSize,
					Id:    stock.SizeID,
					Name:  stock.Size,
					Value: stock.Size,
				})

				item.SkuItems = append(item.SkuItems, &sku)
			}
		}

		// yield item result
		if err = yield(ctx, &item); err != nil {
			return err
		}
	} else if len(qvMatched) > 1 {
		var viewData productQuickViewStructure
		if err := json.Unmarshal(qvMatched[1], &viewData); err != nil {
			c.logger.Error(err)
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
		reviews, _ := strconv.ParseInt(viewData.ReviewSummary.TotalReviews)
		rating, _ := strconv.ParseFloat(viewData.ReviewSummary.AverageOverallRating)
		// build product data
		item := pbItem.Product{
			Source: &pbItem.Source{
				Id:           strconv.Format(viewData.ProductID),
				CrawlUrl:     resp.Request.URL.String(),
				CanonicalUrl: canUrl,
			},
			BrandName:   viewData.BrandName,
			Title:       viewData.ProductName,
			Description: viewData.Description,
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
			},
			CrowdType:   strings.Join(viewData.Genders, ","),
			Category:    viewData.DefaultCategory,
			SubCategory: viewData.DefaultSubCategory,
			Stats: &pbItem.Stats{
				ReviewCount: int32(reviews),
				Rating:      float32(rating),
			},
		}

		for _, rawSku := range viewData.Styles {
			originalPrice, _ := strconv.ParsePrice(rawSku.OriginalPrice)
			price, _ := strconv.ParsePrice(rawSku.Price)
			discount, _ := strconv.ParseInt(strings.TrimSuffix(rawSku.PercentOff, "%"))

			var medias []*pbMedia.Media
			for j, m := range rawSku.Images {
				medias = append(medias, pbMedia.NewImageMedia(
					m.ImageID,
					fmt.Sprintf("https://m.media-amazon.com/images/I/%s.jpg", m.ImageID),
					fmt.Sprintf("https://m.media-amazon.com/images/I/%s._AC_SR1000,750_.jpg", m.ImageID),
					fmt.Sprintf("https://m.media-amazon.com/images/I/%s._AC_SR700,525_.jpg", m.ImageID),
					fmt.Sprintf("https://m.media-amazon.com/images/I/%s._AC_SR500,375_.jpg", m.ImageID),
					"",
					j == 0,
				))
			}

			for _, stock := range rawSku.Stocks {
				sku := pbItem.Sku{
					SourceId: stock.StockID,
					Price: &pbItem.Price{
						Currency: regulation.Currency_USD,
						Current:  int32(price * 100),
						Msrp:     int32(originalPrice * 100),
						Discount: int32(discount),
					},
					Medias: medias,
					Stock:  &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
				}
				onhandCount, _ := strconv.ParseInt(stock.OnHand)
				if onhandCount > 0 {
					sku.Stock.StockStatus = pbItem.Stock_InStock
					sku.Stock.StockCount = int32(onhandCount)
				}

				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecColor,
					Id:    strconv.Format(rawSku.ColorID),
					Name:  rawSku.Color,
					Value: rawSku.Color,
				})

				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecSize,
					Id:    stock.SizeID,
					Name:  stock.Size,
					Value: stock.Size,
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
		// yield item result
		if err = yield(ctx, &item); err != nil {
			return err
		}
	} else {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}

	return nil
}

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		//"https://www.zappos.com/",
		// "https://www.zappos.com/men-bags/COjWAcABAuICAgEY.zso",
		// "https://www.zappos.com/p/nike-tanjun-black-white/product/8619473/color/151",
		"https://www.zappos.com/a/the-style-room/p/rag-bone-watch-belt-black/product/9532098/color/3",
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
