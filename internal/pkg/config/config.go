package config

var (
	// CrawlRequestTopic = "chameleon.smelter.v1.crawl.Command_Request"
	CrawlRequestQueue    = "queue://chameleon.smelter.v1.crawl.Command_Request"
	CrawlRequestQueueSet = "queue://chameleon.smelter.v1.crawl.Command_Request.Set"
	// Product item
	CrawlItemProductTopic = "chameleon.smelter.v1.crawl.item.Product"
)
