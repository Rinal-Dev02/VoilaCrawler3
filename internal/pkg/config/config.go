package config

var (
	CrawlStoreList               = "queue://chameleon.smelter.v1.flow.Store"
	CrawlRequestStoreQueuePrefix = "queue://chameleon.smelter.v1.crawl.Command_Request?storeId="
	CrawlRequestStoreQueue       = CrawlRequestStoreQueuePrefix + "%s"
	CrawlRequestQueueSet         = "queue://chameleon.smelter.v1.crawl.Command_Request.Set"

	// Topic
	CrawlResponseTopic = "chameleon.smelter.v1.crawl.Response"
	// Request
	CrawlRequestTopic        = "chameleon.smelter.v1.crawl.Command_Request"
	CrawlRequestHistoryTopic = "chameleon.smelter.v1.crawl.RequestHistory"
	// Item
	CrawlItemProductTopic = "chameleon.smelter.v1.crawl.item.Product"

	// TTL
	DefaultTtlPerRequest int32 = 5 * 60
)
