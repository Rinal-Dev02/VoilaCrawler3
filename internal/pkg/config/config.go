package config

var (
	CrawlStoreList               = "queue://chameleon.smelter.v1.flow.Store"
	CrawlRequestStoreQueuePrefix = "queue://chameleon.smelter.v1.crawl.Command_Request?storeId="
	CrawlRequestStoreQueue       = CrawlRequestStoreQueuePrefix + "%s"
	CrawlRequestQueueSet         = "queue://chameleon.smelter.v1.crawl.Command_Request.Set"

	// Topic
	CrawlResponseTopic = "chameleon.smelter.v1.crawl.Response"
	// Request
	CrawlRequestTopic            = "chameleon.smelter.v1.crawl.Request"
	CrawlRequestHistoryTopic     = "chameleon.smelter.v1.crawl.RequestHistory"
	CrawlStoreRequestTopicPrefix = "store.Request"
	// Item
	CrawlItemTopic         = "chameleon.smelter.v1.crawl.Item"
	CrawlItemRealtimeTopic = "chameleon.smelter.v1.crawl.Item_Realtime"
	// Error
	CrawlErrorTopic = "chameleon.smelter.v1.crawl.Error"

	// TTL
	DefaultTtlPerRequest int32 = 5 * 60
)
