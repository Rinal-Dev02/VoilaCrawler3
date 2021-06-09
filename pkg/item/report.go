package item

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
)

func Report(i interface{}, logger glog.Log) error {
	switch v := i.(type) {
	case *item.Product:
		return productReport(v, logger)
	}
	return nil
}

func productReport(item *item.Product, logger glog.Log) error {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	desc := item.GetDescription()
	if len([]rune(desc)) > 80 {
		desc = string([]rune(desc)[:80]) + "..."
	}
	rows := []table.Row{
		{"Title", item.GetTitle()},
		{"Description", desc},
		{"Brand", item.GetBrandName()},
		{"Category", item.GetCategory()},
		{"Sub Category", item.GetSubCategory()},
		{"Sub Category 2", item.GetSubCategory2()},
		{"Sub Category 3", item.GetSubCategory3()},
		{"Sub Category 4", item.GetSubCategory4()},
		{"Item Medias Count", len(item.GetMedias())},
		{"Item Currency", item.GetPrice().GetCurrency().String()},
		{"Item CurrentPrice", item.GetPrice().GetCurrency()},
		{"Item MSRP", item.GetPrice().GetMsrp()},
		{"Stock", item.GetStock().GetStockStatus().String()},
		{"Rate", item.GetStats().GetRating()},
		{"Review Count", item.GetStats().GetReviewCount()},
	}
	t.AppendRows(rows)
	t.AppendSeparator()
	t.AppendRow(table.Row{"SKU Count", len(item.GetSkuItems())})
	t.AppendSeparator()
	for _, sku := range item.GetSkuItems() {
		skuSpec := "ID: " + sku.GetSourceId()
		for _, spec := range sku.GetSpecs() {
			skuSpec = skuSpec + " | " + fmt.Sprintf("%s: %s", spec.GetType(), spec.GetName())
		}
		skuSpec = skuSpec + " | Price:" + strconv.Format(sku.GetPrice().GetCurrent()) +
			" | MSRP: " + strconv.Format(sku.GetPrice().GetMsrp()) +
			" | MediasCount: " + strconv.Format(len(sku.GetMedias()))
		t.AppendRow(table.Row{"", skuSpec})
	}
	t.Render()
	return nil
}
