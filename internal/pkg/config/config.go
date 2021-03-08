package config

var (
	CrawlStoreList         = "queue://chameleon.smelter.v1.flow.Store"
	CrawlRequestStoreQueue = "queue://chameleon.smelter.v1.crawl.Command_Request?storeId=%s"
	CrawlRequestQueueSet   = "queue://chameleon.smelter.v1.crawl.Command_Request.Set"

	// Product item
	CrawlItemProductTopic = "chameleon.smelter.v1.crawl.item.Product"
)
