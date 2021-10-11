// SEO this spider is used to do seo info fetch

package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"github.com/voiladev/VoilaCrawler/bin/tools/digest/diffbot"
	"github.com/voiladev/VoilaCrawler/bin/tools/digest/util"
	"github.com/voiladev/VoilaCrawler/pkg/brand"
	cmd "github.com/voiladev/VoilaCrawler/pkg/cli"
	"github.com/voiladev/VoilaCrawler/pkg/context"
	ctxutil "github.com/voiladev/VoilaCrawler/pkg/context"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/regulation"
	pbCrawl "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl"
	pbItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	pbProxy "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/protoutil"

	//"github.com/voiladev/go-framework/randutil"
	"html"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/voiladev/go-framework/strconv"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type _Crawler struct {
	httpClient    http.Client
	diffbotClient *diffbot.DiffbotCient

	categoryPathMatcher     *regexp.Regexp
	categoryJsonPathMatcher *regexp.Regexp
	productGroupPathMatcher *regexp.Regexp
	productPathMatcher      *regexp.Regexp
	logger                  glog.Log

	crawlerClient pbCrawl.CrawlerManagerClient
	conversionMap map[string]string
}

func (_ *_Crawler) New(cli *cmd.Context, client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient:              client,
		categoryPathMatcher:     regexp.MustCompile(`^(/[a-z0-9_-]+)?/(women|men)(/[a-z0-9_-]+){1,6}/cat/?$`),
		categoryJsonPathMatcher: regexp.MustCompile(`^/api/product/search/v2/categories/([a-z0-9]+)`),
		productGroupPathMatcher: regexp.MustCompile(`^(/[a-z0-9_-]+)?(/[a-z0-9_-]+){2}/grp/[0-9]+/?$`),
		productPathMatcher:      regexp.MustCompile(`^(/[a-z0-9_-]+)?(/[a-z0-9_-]+){2}/prd/[0-9]+/?$`),
		logger:                  logger.New("_Crawler"),
	}
	if cli.String("diffbot-token") != "" {
		c.diffbotClient, _ = diffbot.New(cli.String("diffbot-token"), logger)
	}
	err := func() error {
		crawletAddr := cli.String("crawlet-addr")
		if crawletAddr != "" {
			conn, err := grpc.DialContext(cli.Context, crawletAddr,
				grpc.WithInsecure(),
				grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(100*1024*1024)),
				grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(100*1024*1024)),
			)
			if err != nil {
				return err
			}
			c.crawlerClient = pbCrawl.NewCrawlerManagerClient(conn)
		}
		conversionSites := cli.String("conver-site")
		if conversionSites != "" {
			c.conversionMap = make(map[string]string)
			for _, conversionSite := range strings.Split(conversionSites, ",") {
				conversionSiteInfo := strings.Split(conversionSite, ":")
				if len(conversionSiteInfo) != 2 {
					return fmt.Errorf("conversionSite %s format error", conversionSite)
				}
				c.conversionMap[strings.TrimSpace(conversionSiteInfo[0])] = strings.TrimSpace(conversionSiteInfo[1])
			}
		}
		return nil
	}()
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "__tools_op_product__"
}

// Version
func (c *_Crawler) Version() int32 {
	return 1
}

// CrawlOptions
func (c *_Crawler) CrawlOptions(u *url.URL) *crawler.CrawlOptions {
	options := crawler.NewCrawlOptions()
	options.EnableHeadless = false
	options.EnableSessionInit = false
	options.Reliability = pbProxy.ProxyReliability_ReliabilityRealtime
	options.MustCookies = append(options.MustCookies)
	options.SkipDoRequest = true

	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*"}
}

func (c *_Crawler) CanonicalUrl(rawurl string) (string, error) {
	return rawurl, nil
}

func textClean(s string) string {
	return strings.TrimSpace(html.UnescapeString(s))
}

func (c *_Crawler) Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil || resp == nil || resp.Request == nil {
		return nil
	}

	if c.crawlerClient != nil {
		// convert to our own crawler
		for allowedDomain, siteId := range c.conversionMap {
			if matched, _ := filepath.Match(allowedDomain, resp.Request.URL.Hostname()); matched {
				// build request
				// convert http.Request to pbCrawl.Command_Request and forward
				subreq := pbCrawl.Request{
					TracingId: context.GetString(ctx, context.TracingIdKey),
					JobId:     context.GetString(ctx, context.JobIdKey),
					ReqId:     context.GetString(ctx, context.ReqIdKey),
					SiteId:    siteId,
					Url:       resp.Request.URL.String(),
					Method:    resp.Request.Method,
					Options: &pbCrawl.Request_Options{
						MaxItemCount:        1,
						MaxTtlPerRequest:    60,
						EnableBlockForItems: true,
					},
				}

				if subreq.CustomHeaders == nil {
					subreq.CustomHeaders = make(map[string]string)
				}
				if subreq.SharingData == nil {
					subreq.SharingData = map[string]string{}
				}
				if resp.Request.Body != nil {
					defer resp.Request.Body.Close()
					if data, err := io.ReadAll(resp.Request.Body); err != nil {
						return err
					} else {
						subreq.Body = fmt.Sprintf("%s", data)
					}
				}
				for k := range resp.Request.Header {
					subreq.CustomHeaders[k] = resp.Request.Header.Get(k)
				}

				for k, v := range ctxutil.RetrieveAllValues(ctx) {
					key, ok := k.(string)
					if !ok {
						continue
					}
					val := strconv.Format(v)

					subreq.SharingData[key] = val
				}
				c.logger.Infof("fetch from crawler")

				// do parse
				doParseResp, err := c.crawlerClient.DoParse(ctx, &pbCrawl.DoParseRequest{
					Request:             &subreq,
					EnableBlockForItems: true,
				})
				if err != nil {
					c.logger.Error(err)
					return err
				}
				//c.logger.Debugf("%+v", doParseResp)

				for _, doParseRespDataItem := range doParseResp.GetData() {
					if doParseRespDataItem.GetTypeUrl() != protoutil.GetTypeUrl(&pbCrawl.Item{}) {
						// ignore other type
						continue
					}
					var item pbCrawl.Item
					if err := proto.Unmarshal(doParseRespDataItem.GetValue(), &item); err != nil {
						c.logger.Errorf("unmarshal item failed, error=%s", err)
						continue
					}
					if item.GetData().GetTypeUrl() != protoutil.GetTypeUrl(&pbItem.Product{}) {
						// ignore other type
						continue
					}
					var product pbItem.Product
					if err := proto.Unmarshal(item.GetData().GetValue(), &product); err != nil {
						c.logger.Errorf("unmarshal item failed, error=%s", err)
						continue
					}
					if err := yield(ctx, util.UnmarshalToOpenGraphProduct(&product)); err != nil {
						return err
					}
				}
				return nil
			}
		}
	}

	rawurl := resp.RawUrl().String()
	var (
		diffbotProd  *pbItem.OpenGraph_Product
		diffbotErr   error
		diffDuration time.Duration
		opProd       *pbItem.OpenGraph_Product
		opErr        error
		opDuration   time.Duration
	)

	wg := sync.WaitGroup{}
	if c.diffbotClient != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()

			startTime := time.Now()
			prods, err := c.diffbotClient.Fetch(ctx, rawurl)
			diffDuration = time.Now().Sub(startTime)
			if err != nil {
				c.logger.Error(err)
				diffbotErr = err
				return
			}
			if len(prods) == 0 {
				return
			}
			prod := prods[0]

			diffbotProd = &pbItem.OpenGraph_Product{
				Id:          strings.Trim(prod.ProductID, " ;,"),
				Title:       strings.TrimSpace(prod.Title),
				Description: strings.TrimSpace(prod.Text),
				BrandName:   strings.TrimSpace(prod.Brand),
				Url:         rawurl,
				Price: &pbItem.OpenGraph_Price{
					Currency: regulation.Currency_USD,
				},
			}
			if prod.OfferPriceDetails.Amount != 0 {
				diffbotProd.Price.Value = int32(prod.OfferPriceDetails.Amount * 100)
			} else if prod.RegularPriceDetails.Amount != 0 {
				diffbotProd.Price.Value = int32(prod.RegularPriceDetails.Amount * 100)
			}
			for _, img := range prod.Images {
				if img.URL != "" {
					diffbotProd.Medias = append(diffbotProd.Medias, media.NewImageMedia("", img.URL, "", "", "", "", false))
				}
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()

		startTime := time.Now()
		if opProd, opErr = c.parseOpenGraph(ctx, resp.Request); opErr != nil {
			c.logger.Error(opErr)
		}
		opDuration = time.Now().Sub(startTime)
	}()
	wg.Wait()

	var item *pbItem.OpenGraph_Product
	c.logger.Debugf("duration diffbot:%s, op: %s", diffDuration, opDuration)
	if diffbotProd != nil {
		if opProd != nil {
			c.logger.Debug("merge diffbot and opengraph")
			diffbotProd.Site = opProd.Site
			if diffbotProd.Title == "" {
				diffbotProd.Title = opProd.Title
			}
			if diffbotProd.Description == "" {
				diffbotProd.Description = opProd.Description
			}
			if diffbotProd.Price.Value == 0 {
				diffbotProd.Price = opProd.Price
			}
			if len(diffbotProd.Medias) == 0 || len(diffbotProd.Medias) < len(opProd.Medias) {
				diffbotProd.Medias = opProd.Medias
			}
			if diffbotProd.BrandName == "" {
				diffbotProd.BrandName = opProd.BrandName
			}
		}
		item = diffbotProd
	} else if opProd != nil {
		c.logger.Debug("opengraph")
		item = opProd
	} else {
		err := diffbotErr
		if err == nil {
			err = opErr
		}
		return yield(ctx, &pbCrawl.Error{ErrMsg: err.Error()})
	}

	if item.Site == nil {
		item.Site = &pbItem.OpenGraph_Site{}
	}
	if item.BrandName == "" {
		item.BrandName = brand.GetBrand(resp.Request.URL.Hostname())
		if item.BrandName != "" && item.Site.Name == "" {
			item.Site.Name = item.BrandName
		}
	}

	u := *resp.Request.URL
	u.Path = ""
	u.RawQuery = ""
	u.Fragment = ""
	item.Site.Homepage = u.String()
	item.Site.Domain = u.Hostname()
	if item.Site.Name == "" {
		name := u.Hostname()
		for _, pre := range []string{"www.", "www2.", "shop.", "us.", "fr.", "au.", "eu", "usa.", "uk.", "au.", "ca."} {
			name = strings.TrimPrefix(name, pre)
		}
		fields := strings.Split(name, ".")
		item.Site.Name = strings.Join(fields[0:len(fields)-1], ",")
	}
	item.BrandName = strings.TrimSpace(strings.TrimPrefix(item.BrandName, "brand:"))
	return yield(ctx, item)
}

func (c *_Crawler) parseOpenGraph(ctx context.Context, req *http.Request) (*pbItem.OpenGraph_Product, error) {
	if c == nil || req == nil {
		return nil, nil
	}

	opts := c.CrawlOptions(req.URL)
	resp, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
		EnableProxy:       true,
		EnableHeadless:    opts.EnableHeadless,
		EnableSessionInit: opts.EnableSessionInit,
		DisableCookieJar:  true,
		Reliability:       opts.Reliability,
	})
	if err != nil {
		c.logger.Error(err)
		return nil, err
	}

	dom, err := resp.Selector()
	if err != nil {
		c.logger.Error(err)
		return nil, err
	}

	// ogType := textClean(dom.Find(`meta[property="og:type"]`).AttrOr("content", ""))
	// if ogType != "" && ogType != "product" {
	// 	// unsupported path means this page may not a product detail page
	// 	return crawler.ErrUnsupportedPath
	// }

	item := pbItem.OpenGraph_Product{
		Site:  &pbItem.OpenGraph_Site{},
		Price: &pbItem.OpenGraph_Price{},
	}
	item.Site.Name = dom.Find(`meta[property="og:site_name"]`).AttrOr("content", "")

	for _, key := range []string{
		`branch:deeplink:product`,
	} {
		v := textClean(dom.Find(key).AttrOr("content", ""))
		if v != "" {
			item.Id = v
			break
		}
	}

	for _, key := range []string{
		`meta[property="og:title"]`,
		`meta[name="twitter:title"]`,
		`meta[name="title"]`,
	} {
		v := textClean(dom.Find(key).AttrOr("content", ""))
		if v != "" {
			item.Title = v
			break
		}
	}
	if item.Title == "" {
		item.Title = textClean(dom.Find(`title`).Text())
	}

	for _, key := range []string{
		`meta[property="og:description"]`,
		`meta[property="description"]`,
		`meta[name="twitter:description"]`,
		`meta[name="description"]`,
	} {
		v := textClean(dom.Find(key).AttrOr("content", ""))
		if v != "" {
			item.Description = v
			break
		}
	}

	for _, key := range []string{
		`link[rel="canonical"]`,
		`meta[property="og:url"]`,
		`meta[property="url"]`,
	} {
		v := dom.Find(key).AttrOr("content", dom.Find(key).AttrOr("href", ""))
		if v != "" {
			item.Url = v
			break
		}
	}

	for _, key := range []string{
		`meta[property="og:image:secure_url"]`,
		`meta[property="og:image"]`,
		`meta[property="image:secure_url"]`,
		`meta[property="image"]`,
	} {
		sel := dom.Find(key)
		for i := range sel.Nodes {
			node := sel.Eq(i)
			if v := node.AttrOr("content", ""); v != "" {
				if strings.HasPrefix(v, "//") {
					v = "https:" + v
				} else if strings.HasPrefix(v, "/") {
					v = resp.CurrentUrl().Scheme + "://" + resp.CurrentUrl().Host + v
				}
				v, err := util.FormatImageUrl(v)
				if err != nil {
					continue
				}
				item.Medias = append(item.Medias, media.NewImageMedia("", v, "", "", "", "", false))
			}
		}
		if len(item.Medias) > 0 {
			break
		}
	}

	for _, key := range []string{
		`meta[property="og:price:currency"]`,
		`meta[property="og:product:price:currency"]`,
	} {
		v := strings.ToUpper(dom.Find(key).AttrOr("content", ""))
		if v == "USD" {
			item.Price.Currency = regulation.Currency_USD
		}
	}
	if item.Price.Currency == 0 {
		item.Price.Currency = regulation.Currency_USD
	}

	for _, key := range []string{
		`meta[property="og:price:amount"]`,
		`meta[property="og:product:price:amount"]`,
	} {
		v := dom.Find(key).AttrOr("content", "0")
		if v != "" {
			if vv, _ := strconv.ParsePrice(v); vv > 0 {
				item.Price.Value = int32(vv * 100)
				break
			}
		}
	}

	// feat: added support of brand fetch
	if item.Title != "" && item.BrandName == "" {
		fields := strings.Split(item.Title, " | ")
		lastField := fields[len(fields)-1]
		if len(fields) > 1 && len(lastField) < 20 {
			item.BrandName = lastField
		}
	}
	return &item, nil
}

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		// "https://us.princesspolly.com/collections/basics/products/madelyn-top-green",
		"https://www.revolve.com/house-of-harlow-1960-x-sofia-richie-portofino-dress/dp/HOOF-WD751/?d=Womens&page=1&lc=2&itrownum=1&itcurrpage=1&itview=05",
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
	cmd.NewApp(
		&_Crawler{},
		&cli.StringFlag{Name: "diffbot-token", Usage: "diffbot api token"},
		&cli.StringFlag{Name: "crawlet-addr", Usage: "crawlet server address"},
		&cli.StringFlag{Name: "conver-site", Usage: "conver to self crawl service, example 'domain1:id1,domain2:id2...'"},
	).Run(os.Args)
}