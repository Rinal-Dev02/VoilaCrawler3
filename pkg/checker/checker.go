package checker

import (
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	goHttp "net/http"
	"strings"

	"github.com/voiladev/VoilaCrawler/pkg/context"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/regulation"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl"
	pbItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/protoutil"
	"google.golang.org/protobuf/proto"
)

const (
	imgSizeSmall = iota
	imgSizeMedium
	imgSizeLarge
)

var supportedHttpMethods = map[string]struct{}{}

func init() {
	for _, m := range http.SupportedHttpMethods {
		supportedHttpMethods[m] = struct{}{}
	}
}

func Check(ctx context.Context, i interface{}, logger glog.Log, httpClient http.Client) error {
	if i == nil {
		return errors.New("Checker: got invalid yield item")
	}

	// check context shared values
	if context.GetString(ctx, context.TracingIdKey) == "" {
		logger.Warnf("Checker: shared TracingIdKey not found")
	}
	if context.GetString(ctx, context.ReqIdKey) == "" {
		logger.Warnf("Checker: shared ReqIdKey not found")
	}

	switch v := i.(type) {
	case *http.Request:
		return checkRequest(ctx, v, logger)
	case *crawl.Error:
		return checkError(ctx, v, logger)
	case *pbItem.Product:
		return checkProduct(ctx, v, logger, httpClient)
	case *pbItem.Tiktok_Item:
		return nil
	case *pbItem.Tiktok_Author:
		return nil
	case *pbItem.Youtube_Video:
		return nil
	default:
		return nil
	}
}

func checkRequest(ctx context.Context, req *http.Request, logger glog.Log) error {
	if req == nil {
		return errors.New("Checker.Request: nil")
	}
	if _, ok := supportedHttpMethods[strings.ToUpper(req.Method)]; !ok {
		return errors.New("Checker.Request: invalid http method")
	}
	if req.URL == nil || (req.URL.Host == "" && req.URL.Path == "") {
		return errors.New("Checker.Request: invalid request url")
	}
	return nil
}

// checkProduct
func checkProduct(ctx context.Context, item *pbItem.Product, logger glog.Log, httpClient http.Client) error {
	if !context.Exists(ctx, "item.index") {
		logger.Warnf("Checker.Product: no shared value item.index found")
	}

	if item == nil {
		return errors.New("Checker.Product: nil")
	}
	if item.GetSource().GetId() == "" {
		return errors.New("Checker.Product: invalid source id")
	}
	if item.GetSource().GetCrawlUrl() == "" {
		return errors.New("Checker.Product: invalid crawl url")
	}
	if item.GetSource().GetCanonicalUrl() == "" {
		return errors.New("Checker.Product: invalid canonical url")
	}
	if item.BrandName == "" {
		return errors.New("Checker.Product: invalid brand")
	}
	if item.Title == "" {
		return errors.New("Checker.Product: invalid title")
	}
	if item.Description == "" {
		return errors.New("Checker.Product: invalid description")
	}

	if item.Category == "" {
		logger.Warnf("Checker.Product: no category found")
		if item.SubCategory != "" {
			return errors.New("Checker.Product: category is empty but subCategory is not")
		}
	}

	mediaChecker := func(m *media.Media) error {
		if m == nil {
			return errors.New("Checker.Product: invalid media")
		}

		switch m.GetDetail().GetTypeUrl() {
		case protoutil.GetTypeUrl(&media.Media_Image{}):
			var img media.Media_Image
			if err := proto.Unmarshal(m.GetDetail().GetValue(), &img); err != nil {
				return errors.New("Checker.Product: unmarshal image media failed")
			}
			if img.GetOriginalUrl() == "" {
				return errors.New("Checker.Product: invalid image original url")
			}
			if img.GetLargeUrl() == "" {
				return errors.New("Checker.Product: invalid image large url")
			} else if err := checkImage(ctx, img.GetLargeUrl(), imgSizeLarge, httpClient); err != nil {
				return err
			}
			if img.GetMediumUrl() == "" {
				return errors.New("Checker.Product: invalid image medium url")
			} else if err := checkImage(ctx, img.GetMediumUrl(), imgSizeMedium, httpClient); err != nil {
				return err
			}
			if img.GetSmallUrl() == "" {
				return errors.New("Checker.Product: invalid image small url")
			} else if err := checkImage(ctx, img.GetSmallUrl(), imgSizeSmall, httpClient); err != nil {
				return err
			}
			if img.GetSmallUrl() == img.GetLargeUrl() {
				return errors.New("Checker.Product: SmallUrl should in width >=500, MediumUrl should in width >=600, LargeUrl should in width >=800")
			}
		case protoutil.GetTypeUrl(&media.Media_Video{}):
			var video media.Media_Video
			if err := proto.Unmarshal(m.GetDetail().GetValue(), &video); err != nil {
				return errors.New("Checker.Product: unmarshal video media failed")
			}
			if video.GetOriginalUrl() == "" {
				return errors.New("Checker.Product: invalid video OriginalUrl")
			}
			if video.GetCover().GetOriginalUrl() == "" {
				logger.Warnf("Checker.Product: no cover found for video")
			}
		default:
			return fmt.Errorf("Checker.Product: unsupported media type %s", m.GetDetail().GetTypeUrl())
		}
		return nil
	}

	for i, m := range item.Medias {
		if e := mediaChecker(m); e != nil {
			return e
		}
		if i == 0 && !m.IsDefault {
			return fmt.Errorf("Checker.Product: the first media for item should be the default media")
		}
	}

	if item.GetStock().GetStockStatus() == pbItem.Stock_InStock && len(item.SkuItems) == 0 {
		return errors.New("Checker.Product: no valid item found")
	}

	var (
		isSkuMediasFound = false
		isMediasToAllSku = true
		isStatsFound     = (item.GetStats() == nil)
		skuIds           = map[string]struct{}{}
	)
	for _, sku := range item.SkuItems {
		if sku == nil {
			return errors.New("Checker.Product: no sku found")
		}
		if sku.SourceId == "" {
			return errors.New("Checker.Product: invalid sku SourceId")
		}
		if _, ok := skuIds[sku.SourceId]; ok {
			return fmt.Errorf("Checker.Product: sku id %s already exists", sku.SourceId)
		}
		skuIds[sku.SourceId] = struct{}{}

		if len(sku.Specs) == 0 {
			return fmt.Errorf("Checker.Product: no sku spec found for sku %s", sku.SourceId)
		}

		specIds := map[string]struct{}{}
		specTypeFilter := map[pbItem.SkuSpecType]struct{}{}
		for _, spec := range sku.Specs {
			if _, ok := pbItem.SkuSpecType_name[int32(spec.GetType())]; !ok ||
				spec.GetType() == pbItem.SkuSpecType_SkuSpecUnknown {
				return fmt.Errorf("Checker.Product: invalid spec type for sku %v", sku.SourceId)
			}
			if _, ok := specTypeFilter[spec.GetType()]; ok {
				return fmt.Errorf("Checker.Product: sku spec %s has been exists", spec.GetType())
			}
			specTypeFilter[spec.GetType()] = struct{}{}

			if spec.GetId() == "" {
				return fmt.Errorf("Checker.Product: invalid sku spec id, if not spec id found, use spec name or value")
			}
			if spec.GetId() == sku.SourceId {
				return fmt.Errorf("Checker.Product: sku id can not be the id of sku spec")
			}
			if _, ok := specIds[spec.GetId()]; ok {
				return fmt.Errorf("Checker.Product: sku spec id %v has been used, if no sku source id found, you can use sku's spec ids to generate a unique id", spec.GetId())
			}
			specIds[spec.GetId()] = struct{}{}
			if spec.GetName() == "" {
				return fmt.Errorf("Checker.Product: invalid sku spec name, if not name found, use spec value")
			}
			if spec.GetValue() == "" {
				return fmt.Errorf("Checker.Product: invalid sku spec value, if not value found, use spec name")
			}
		}

		if len(sku.Medias) == 0 {
			if isMediasToAllSku {
				isMediasToAllSku = false
			}
		} else {
			isSkuMediasFound = true
			for i, m := range sku.Medias {
				if i == 0 && !m.IsDefault {
					return fmt.Errorf("Checker.Product: the first media for sku %s should be the default media", sku.SourceId)
				}
				if e := mediaChecker(m); e != nil {
					return e
				}
			}
		}

		// Currently only supports USD
		if sku.GetPrice().GetCurrency() != regulation.Currency_USD {
			return fmt.Errorf("Checker.Product: invalid price currency for sku %s", sku.SourceId)
		}
		if sku.GetPrice().GetCurrent() <= 0 {
			return fmt.Errorf("Checker.Product: invalid current price for sku %s", sku.SourceId)
		}
		if sku.GetPrice().GetMsrp() <= 0 {
			return fmt.Errorf("Checker.Product: invalid msrp price for sku %s, if not found, use current price", sku.SourceId)
		}
		if sku.GetPrice().GetDiscount() < 0 {
			return fmt.Errorf("Checker.Product: invalid discount price for sku %s", sku.SourceId)
		}
		if sku.GetPrice().GetDiscount1() < 0 {
			return fmt.Errorf("Checker.Product: invalid discount1 price for sku %s", sku.SourceId)
		}
		if sku.GetStock().GetStockStatus() == pbItem.Stock_OutOfStock && sku.GetStock().GetStockCount() > 0 {
			return fmt.Errorf("Checker.Product: invalid discount1 price for sku %s", sku.SourceId)
		}
		isStatsFound = isStatsFound || sku.GetStats() != nil
	}
	if !isMediasToAllSku {
		if isSkuMediasFound {
			return errors.New("Checker.Pointer: medias must be set for each sku")
		}
		if len(item.Medias) == 0 {
			return errors.New("Checker.Pointer: no medias found")
		}
	}

	if item.GetStock().GetStockStatus() == pbItem.Stock_StockStatusUnknown {
		logger.Errorf("Checker.Product: item Stock status is not set, if can found, you can use sku's StockStatus")
	}
	if !isStatsFound {
		logger.Warnf("Checker.Product: not stats found")
	}
	return nil
}

func checkError(ctx context.Context, e *crawl.Error, logger glog.Log) error {
	if e == nil {
		return errors.New("Checker.Error: nil")
	}
	if e.GetErrMsg() == "" {
		return errors.New("Checker.Error: without message")
	}
	return nil
}

// checkImage do http request to check image width
func checkImage(_ context.Context, url string, imgSizeType int, _ http.Client) error {
	var (
		imgSizeName string
		imgSize     int
	)
	switch imgSizeType {
	case imgSizeSmall:
		imgSizeName = "SmallImage"
		imgSize = 500
	case imgSizeMedium:
		imgSizeName = "MediumImage"
		imgSize = 600
	case imgSizeLarge:
		imgSizeName = "LargeImage"
		imgSize = 800
	default:
		return fmt.Errorf("Checker.Product.Media: unsupport image size type %d", imgSizeType)
	}

	//imgReq, _ := http.NewRequest(http.MethodGet, url, nil)
	//imgResp, err := httpClient.Do(ctx, imgReq)
	imgResp, err := goHttp.Get(url)
	if err != nil {
		return fmt.Errorf("Checker.Product.Media: Get %s err=%s", imgSizeName, err)
	}
	defer imgResp.Body.Close()
	img, _, err := image.Decode(imgResp.Body)
	if err != nil {
		return fmt.Errorf("Checker.Product.Media: Read %s Data err=%s", imgSizeName, err)
	}
	if img.Bounds().Dx() < imgSize {
		return fmt.Errorf("Checker.Product.Media: %s width should >=%d, not %d", imgSizeName, imgSize, img.Bounds().Dx())
	}
	return nil
}
