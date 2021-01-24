# VoilaCrawler

VoilaCrawler is a distributed web crawler. It's more like python scrapy except that VoilaCrawler is born to be distributed.

This repo mainly contains three parts: `crawl-api` and `crawlet` and `spiders`.

### crawl-api

`crawl-api` is a request manager. It saves requests to mysql db, meanwhile it send them to the mq to queue up. Requests may come from two ways `API` which is the window we can submit crawl job and `MQ` which is published by crawlet.

#### Features

1. record's all `request` and publish them to `nsq`(a golang based mq).
2. requeue failed request based on `maxRetryCount`
3. requeue timeout request based on `maxTtlPerRequest` (TODO)

### crawlet

`crawlet` is more like `scrapy` which do http request and pass the response to specified spiders. it's a spider manager, and an request/response controller.

#### Features

1. supports golang buildin `http.Request` and `http.Response`
2. supports data sharing bewteen requests
3. buildin http proxy support with [ProxyCrawl](https://proxycrawl.com)
4. golang plugin based spider
5. link dig depth support
6. supports `index` item across list [response](response)
7. cralwer health check

### spiders

`spider` is writing by golang and builded to a golang plugin who MUST matches the `Crawler` interface.

```golang
// HealthChecker used to test if website struct changed
type HealthChecker interface {
	// NewTestRequest generate test requests
	NewTestRequest(ctx context.Context) []*http.Request

	// CheckTestResponse used to check whether website struct changed
	CheckTestResponse(ctx context.Context, resp *http.Response) error
}

// Crawler
type Crawler interface {
	HealthChecker

	// ID returns crawler unique id, which must be the same for all the version.
	ID() string

	// Version returns the version of current this crawler, which should be an active number.
	Version() int32

	// CrawlOptions return crawler action requirement
	CrawlOptions() *CrawlOptions

	// AllowedDomains returns the domains this crawler supportes
	AllowedDomains() []string

	// IsUrlMatch check whether the supplied url matched the crawler's url set.
	// if matched, then can use crawler to extract info from the response of this url.
	IsUrlMatch(*url.URL) bool

	// Parser used to parse http request parse.
	//   param ctx used to share info between parent and child. and it can set the max ttl for parse job.
	//   param resp represents the http response, with act as a real http response.
	//   param yield use to yield data with can be final data or an other http request
	Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error
}
```

The `HealthChecker` is used to check if the target website data struct is changed. `NewTestRequest` returns one or more http.Request. `crawlet` will do http request with this params and pass the response to `CheckTestResponse`. `CheckTestResponse` should check key fields values. You can use Parser to extract data from this response and check if the key field's values is changed or is empty.

It's recommended that the requests returned by `NewTestRequest` is const, so that it's more easy to check.


In `Crawler` interface, `ID` is the id of the spider with `Version` to unique identitify a spider. The `ID` should be constant for all version. 

`CrawlOptions`'s data struct as bellow:

```golang
// CrawlOptions
type CrawlOptions struct {
	// EnableHeadless crawl by http headless brower which supports javascripts running
	EnableHeadless bool `json:"enableHeadless"`

	// LoginRequired indicates that this website needs login before crawl
	// there must be an login subsystem with manages all the robot accounts
	// and cache the cookies after signin. (TODO)
	LoginRequired bool `json:"loginRequired"`

	// MustHeader specify the musted http headers
	MustHeader http.Header `json:"mustHeader"`

	// MustCookies specify the musted cookies
	MustCookies []*http.Cookie `json:"mustCookies"`
}
```

`AllowedDomains` returns the supported domains of this crawler.

`IsUrlMatch` checkes if the crawler can parse the specified url. if `true` then `crawlet` will call Parse with this request.

`Parse` is the **entry** of a cralwer. it accepts three params: ctx, resp, yeild. `ctx` is a golang `context.Context`. `ctx` is a very magic param. it can be used to sharing params between request, it can used to monitor timeout. more docs see [Golang officle doc](https://golang.org/pkg/context/). `resp` is the http.Response.

`Parse.yield` param is a callback function used to pass sub http.request([eg](https://github.com/voiladev/VoilaCrawl/blob/adbe18d7334c5f7f7bf90e92c80ae6868470cdc5/cmd/spiders/com/ruelala/ruelala.go#L268)) or result item([eg](https://github.com/voiladev/VoilaCrawl/blob/adbe18d7334c5f7f7bf90e92c80ae6868470cdc5/cmd/spiders/com/ruelala/ruelala.go#L490)) to `crawlet`. it receives an context and an result which is an interface. The result value currently supports `*http.Request` and items defined in [protobuf](https://github.com/voiladev/protobuf/blob/main/proto/chameleon/smelter/v1/crawl/item/data.proto). it'a a good idea to stop parse if `yield` returns an error.


#### Sharing Data

`context.Context` is a magic package. i mainly used it to sharing data between requests. context is in an inheritance relationship. children can get values from parent ctx, but parents cannot get child shared values.

Because `VoilaCrawl` is a distributed crawl program. so it must serialize and deserialization shared data. **To simplify the deserialization, the values shared must be string type** Other types of shared values will be ignored when serialize.



### How to write a crawler

A crawler plugin must have a entry func `New(client http.Client, logger glog.Log) (crawler.Crawler, error)`. `crawlet` will call this func to get a `Crawler` instance. The `New` func accepts a `http.Client` and a `glog.Log` ([def](https://github.com/voiladev/go-framework/blob/c10e4eb1a4bc0b599116126318e0d1dc1dc48ff9/glog/glog.go#L17)).

Any website must got multi url path. So in `Parse` func, you must support pathes the website support and tell `IsUrlMatch` that it supports this url path. You can use `regexp.Regexp` to match the url path, or you can use string match for simplicity.

Here is a spider demo. [Link](https://github.com/voiladev/VoilaCrawl/blob/main/cmd/spiders/com/ruelala/ruelala.go)

#### Crawler test

You can write a main func and instance an http Client([eg](https://github.com/voiladev/VoilaCrawl/blob/adbe18d7334c5f7f7bf90e92c80ae6868470cdc5/cmd/spiders/com/ruelala/ruelala.go#L510))

Tips:

1. to parse json data, it's more easy to define a go struct and parse the json data. You can use this [tool](https://mholt.github.io/json-to-go/) to convert json to go struct. (you must check again the generated go struct)


### Development

All developers must develop with this workflow:
1. fork repo VoilaCrawl and clone the forked to your local
2. create/update spiders under dir cmd/spiders
3. commit modification and create a merge request to `main` branch.

Others jobs will be done by me currently. 

All spiders must under `cmd/spiders` dir. the final dir must match the reverse domain schema. eg. www.ruelala.com's  spiders is under dir com.ruelala dir


### TODO

1. check `ProxyCrawl` request limit
2. requeue timeout request based on `maxTtlPerRequest`
3. using webassembly to write spiders which can write by any language
4. health check notification support
5. website session support
