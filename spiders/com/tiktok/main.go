package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/voiladev/VoilaCrawl/pkg/crawler"
	"github.com/voiladev/VoilaCrawl/pkg/net/http"
	"github.com/voiladev/VoilaCrawl/pkg/net/http/cookiejar"
	"github.com/voiladev/VoilaCrawl/pkg/proxy"
	"github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/media"
	pbItem "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
)

type _Crawler struct {
	httpClient http.Client

	detailPageReg          *regexp.Regexp
	detailInternalPageReg  *regexp.Regexp
	detailShortLinkPageReg *regexp.Regexp
	downloadVideoReg       *regexp.Regexp

	logger glog.Log
}

func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient:             client,
		detailPageReg:          regexp.MustCompile(`^/@[0-9a-zA-Z-_.]+/video/[0-9]+/?$`),
		detailInternalPageReg:  regexp.MustCompile(`^/v/[0-9]+.html$`),
		detailShortLinkPageReg: regexp.MustCompile(`^/[a-zA-Z0-9]+/?$`),
		downloadVideoReg:       regexp.MustCompile(`^/video/tos/alisg/tos\-alisg\-pve\-[a-z0-9]+/[a-z0-9]+/?$`),
		logger:                 logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "39b67c9788c4ab57d1b153d9d12141bd"
}

// Version
func (c *_Crawler) Version() int32 {
	return 1
}

// CrawlOptions
func (c *_Crawler) CrawlOptions() *crawler.CrawlOptions {
	options := crawler.NewCrawlOptions()
	options.EnableHeadless = false
	options.LoginRequired = false
	options.MustHeader.Set("Accept-Language", "en-US,en;q=0.8")
	options.MustCookies = append(options.MustCookies)
	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.tiktok.com"}
}

func (c *_Crawler) IsUrlMatch(u *url.URL) bool {
	if c == nil || u == nil {
		return false
	}

	for _, reg := range []*regexp.Regexp{
		c.detailPageReg,
		c.detailShortLinkPageReg,
	} {
		if reg.MatchString(u.Path) {
			return true
		}
	}
	return false
}

func (c *_Crawler) Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	if c.detailShortLinkPageReg.MatchString(resp.Request.URL.Path) ||
		c.detailPageReg.MatchString(resp.Request.URL.Path) ||
		c.detailInternalPageReg.MatchString(resp.Request.URL.Path) {
		return c.parseDetail(ctx, resp, yield)
	} else if c.downloadVideoReg.MatchString(resp.Request.URL.Path) {
		return c.download(ctx, resp, yield)
	}
	return fmt.Errorf("unsupported url %s", resp.Request.URL.String())
}

var (
	videoCoverReg = regexp.MustCompile(`background-image:\s*url\("([^;]+)"\);`)
	initPropReg   = regexp.MustCompile(`<script\s*[^>]*\s*>\s*window.__INIT_PROPS__\s*=\s*([^\r\n]+);?\s*</script>`)
)

func (c *_Crawler) parseDetail(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	c.logger.Debugf("matched %s", resp.Request.URL.Path)

	if c.detailInternalPageReg.MatchString(resp.Request.URL.Path) {
		if href, exists := doc.Find(`link[rel="canonical"]`).Attr("href"); exists && href != "" {
			href = href + "?" + resp.Request.URL.Query().Encode() + "&source=h5_m"
			if req, _ := http.NewRequest(http.MethodGet, href, nil); req != nil {
				return yield(ctx, req)
			}
		}
	}

	// c.logger.Debugf("%s", respBody)

	var (
		rawurl string
		item   *pbItem.Tiktok_Item
	)

	if rawProp := doc.Find("#__NEXT_DATA__").Text(); strings.TrimSpace(rawProp) != "" {
		if rawurl, item, err = parsePropData([]byte(strings.TrimSpace(rawProp))); err != nil {
			return err
		}
	} else if matched := initPropReg.FindSubmatch(respBody); len(matched) > 1 {
		if rawurl, item, err = parsePropData(matched[1]); err != nil {
			return err
		}
	} else {
		var (
			exists bool
			item   = &pbItem.Tiktok_Item{
				Source: &pbItem.Tiktok_Source{},
				Video:  &media.Media_Video{Cover: &media.Media_Image{}},
			}
		)
		rawurl, exists = doc.Find(`meta[property="og:url"]`).Attr("content")
		if !exists {
			return fmt.Errorf("real url of %s not found", resp.Request.URL)
		}
		rawurl = html.UnescapeString(rawurl)

		if val, exists := doc.Find(`meta[property="og:title"]`).Attr("content"); exists {
			item.Title = val
		} else {
			return fmt.Errorf("title of %s not found", resp.Request.URL)
		}
		// no use
		if val, exists := doc.Find(`meta[property="og:description"]`).Attr("content"); exists {
			item.Description = val
		}

		if val, exists := doc.Find(`meta[property="og:video"]`).Attr("content"); exists {
			item.Video.OriginalUrl = html.UnescapeString(val)
		} else if val, exists = doc.Find(`meta[property="og:video:secure_url"]`).Attr("content"); exists {
			item.Video.OriginalUrl = html.UnescapeString(val)
		} else {
			return fmt.Errorf("video url of %s not found", resp.Request.URL)
		}

		if val, exists := doc.Find(`meta[property="og:video:type"]`).Attr("content"); exists {
			item.Video.Type = html.UnescapeString(val)
		}
		if val, exists := doc.Find(`meta[property="og:video:width"]`).Attr("content"); exists {
			v, _ := strconv.ParseInt(val)
			item.Video.Width = int32(v)
		}
		if val, exists := doc.Find(`meta[property="og:video:height"]`).Attr("content"); exists {
			v, _ := strconv.ParseInt(val)
			item.Video.Height = int32(v)
		}
		if val, exists := doc.Find(`meta[property="og:image"]`).Attr("content"); exists {
			item.Video.Cover.OriginalUrl = html.UnescapeString(val)
		} else if val, exists := doc.Find(`meta[property="og:image:secure_url"]`).Attr("content"); exists {
			item.Video.Cover.OriginalUrl = html.UnescapeString(val)
		}
	}

	if item.Source == nil {
		item.Source = &pbItem.Tiktok_Source{}
	}
	item.Source.CrawlUrl = rawurl

	/*
		// this is not necessory
		mockCookieUrls := []string{}
		reg := regexp.MustCompile(`<script\s+type="text/javascript"\s+src="(https://www.tiktok.com/akam/[a-z0-9]+/[a-z0-9]+)"\s+defer>\s*</script>`)
		if matched := reg.FindSubmatch(respBody); len(matched) > 1 {
			mockCookieUrls = append(mockCookieUrls, string(matched[1]))
		}
		reg = regexp.MustCompile(`src="(https?://www.tiktok.com/akam/[a-z0-9]+/pixel_[a-z0-9]+\?a=[a-zA-Z0-9=+-.]+)"`)
		if matched := reg.FindSubmatch(respBody); len(matched) > 1 {
			mockCookieUrls = append(mockCookieUrls, string(matched[1]))
		}

			for _, u := range mockCookieUrls {
				c.logger.Debugf("mock cookie %s", u)
				if req, err := http.NewRequest(http.MethodGet, u, nil); err != nil {
					return err
				} else if resp, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
					EnableProxy:        false,
					DisableBackconnect: false,
				}); err != nil {
					return err
				} else {
					for _, c := range resp.Cookies() {
						cookies[c.Name] = c
					}
				}
			}
	*/

	var (
		cookies, _ = c.httpClient.Jar().Cookies(ctx, resp.Request.URL)
		cookie     string
		expiresAt  time.Time
	)
	for _, c := range cookies {
		v := fmt.Sprintf("%s=%s", c.Name, c.Value)
		if cookie == "" {
			cookie = v
		} else {
			cookie += "; " + v
		}

		if !c.Expires.IsZero() {
			if expiresAt.IsZero() || expiresAt.After(c.Expires) {
				expiresAt = c.Expires
			}
		} else if c.MaxAge > 0 {
			t := time.Now().Add(time.Second * time.Duration(c.MaxAge))
			if expiresAt.IsZero() || expiresAt.After(t) {
				expiresAt = t
			}
		}
	}
	if item.Headers == nil {
		item.Headers = map[string]string{}
	}
	item.Headers["Referer"] = "https://www.tiktok.com/"
	item.Headers["Cookie"] = cookie

	if expiresAt.IsZero() {
		item.ExpiresUtc = time.Now().Add(time.Hour * 3).Unix()
	} else {
		item.ExpiresUtc = expiresAt.Unix()
	}
	return yield(ctx, item)
}

func (c *_Crawler) download(ctx context.Context, resp *http.Response, yield interface{}) error {
	if c == nil || yield == nil {
		return nil
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	c.logger.Debugf("download resp: %d, size: %d", resp.StatusCode, len(data))

	return nil
}

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		// "https://vm.tiktok.com/ZScNvr6C/",
		"https://www.tiktok.com/@kasey.jo.gerst/video/6923743895247506693?sender_device=mobile&sender_web_id=6926525695457117698&is_from_webapp=v2&is_copy_url=0",
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

// local test
func main() {
	var (
		apiToken = os.Getenv("PC_API_TOKEN")
		jsToken  = os.Getenv("PC_JS_TOKEN")
	)
	if apiToken == "" || jsToken == "" {
		panic("env PC_API_TOKEN or PC_JS_TOKEN is not set")
	}

	logger := glog.New(glog.LogLevelDebug)
	client, err := proxy.NewProxyClient(
		cookiejar.New(), logger,
		proxy.Options{APIToken: apiToken, JSToken: jsToken},
	)
	if err != nil {
		panic(err)
	}

	spider, err := New(client, logger)
	if err != nil {
		panic(err)
	}
	opts := spider.CrawlOptions()

	var callback func(ctx context.Context, val interface{}) error
	callback = func(ctx context.Context, val interface{}) error {
		switch i := val.(type) {
		case *http.Request:
			logger.Debugf("Access %s", i.URL)

			for k := range opts.MustHeader {
				i.Header.Set(k, opts.MustHeader.Get(k))
			}
			for _, c := range opts.MustCookies {
				if strings.HasPrefix(i.URL.Path, c.Path) || c.Path == "" {
					val := fmt.Sprintf("%s=%s", c.Name, c.Value)
					if c := i.Header.Get("Cookie"); c != "" {
						i.Header.Set("Cookie", c+"; "+val)
					} else {
						i.Header.Set("Cookie", val)
					}
				}
			}

			resp, err := client.DoWithOptions(ctx, i, http.Options{EnableProxy: false})
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			return spider.Parse(ctx, resp, callback)
		case *pbItem.Tiktok_Item:
			logger.Debugf("Access %s", i.Video.OriginalUrl)

			req, err := http.NewRequest(http.MethodGet, i.Video.OriginalUrl, nil)
			if err != nil {
				return err
			}
			for k, v := range i.Headers {
				if k == "Cookie" {
					continue
				}
				logger.Debugf("k=%s,v=%s", k, v)
				req.Header.Set(k, v)
			}
			resp, err := client.DoWithOptions(ctx, req, http.Options{EnableProxy: false})
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			if data, err := ioutil.ReadAll(resp.Body); err != nil {
				logger.Fatal(err)
			} else {
				logger.Infof("status: %v, size: %d", resp.StatusCode, len(data))
			}
		default:
			data, err := json.Marshal(i)
			if err != nil {
				return err
			}
			logger.Infof("data: %s", data)
		}
		return nil
	}

	for _, req := range spider.NewTestRequest(context.Background()) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
		defer cancel()

		if err := callback(ctx, req); err != nil {
			logger.Fatal(err)
		}
	}
}
