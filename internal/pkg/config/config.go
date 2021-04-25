package config

var (
	// Request
	CrawlRequestTopic            = "chameleon.smelter.v1.crawl.Request"
	CrawlRequestStatusTopic      = "chameleon.smelter.v1.crawl.RequestStatus"
	CrawlStoreRequestTopicPrefix = "store.Request"
	// Item
	CrawlItemTopic = "chameleon.smelter.v1.crawl.Item"
	// Error
	CrawlErrorTopic = "chameleon.smelter.v1.crawl.Error"

	// TTL
	DefaultTtlPerRequest int32 = 5 * 60
)
