package util

import (
	pbItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	"net/url"
	"strings"
)

func UnmarshalToOpenGraphProduct(src interface{}) *pbItem.OpenGraph_Product {
	var item pbItem.OpenGraph_Product
	switch srcValue := src.(type) {
	case *pbItem.Product:
		u, _ := url.Parse(srcValue.GetSource().GetCrawlUrl())
		u.Path = ""
		u.RawQuery = ""
		u.Fragment = ""
		item.Site = &pbItem.OpenGraph_Site{
			Name:     srcValue.GetSite().GetName(),
			Homepage: u.String(),
			Domain:   u.Hostname(),
		}
		if item.Site.Name == "" {
			name := u.Hostname()
			for _, pre := range []string{"www.", "www2.", "shop.", "us.", "fr.", "au.", "eu", "usa.", "uk.", "au.", "ca."} {
				name = strings.TrimPrefix(name, pre)
			}
			fields := strings.Split(name, ".")
			item.Site.Name = strings.Join(fields[0:len(fields)-1], ",")
		}
		item.Id = srcValue.GetSource().GetId()
		item.Title = srcValue.GetTitle()
		item.Description = srcValue.GetDescription()
		item.BrandName = srcValue.GetBrandName()
		item.Url = srcValue.GetSource().GetCrawlUrl()
		item.Medias = srcValue.GetMedias()
		item.Price = &pbItem.OpenGraph_Price{
			Currency: srcValue.GetPrice().GetCurrency(),
			Value:    srcValue.GetPrice().GetCurrent(),
		}
		if item.GetPrice().GetValue() <= 0 {
			for _, sku := range srcValue.GetSkuItems() {
				item.Price.Currency = sku.GetPrice().GetCurrency()
				item.Price.Value = sku.GetPrice().GetCurrent()
			}
		}
	}
	return &item
}
