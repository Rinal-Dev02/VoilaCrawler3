package main

// Refer to https://github.com/soimort/you-get/blob/develop/src/you_get/extractors/tiktok.py

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

	personalVideoList      *regexp.Regexp
	personalVideoJSONList  *regexp.Regexp
	detailPageReg          *regexp.Regexp
	detailInternalPageReg  *regexp.Regexp
	detailShortLinkPageReg *regexp.Regexp
	downloadVideoReg       *regexp.Regexp

	logger glog.Log
}

func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient:             client,
		personalVideoList:      regexp.MustCompile(`^/@[0-9a-zA-Z-_.]+/?$`),
		personalVideoJSONList:  regexp.MustCompile(`^/api/post/item_list/?$`),
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
func (c *_Crawler) CrawlOptions(u *url.URL) *crawler.CrawlOptions {
	options := crawler.NewCrawlOptions()
	options.EnableHeadless = false
	options.LoginRequired = false
	options.MustHeader.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	options.MustHeader.Set("Accept-Charset", "UTF-8,*;q=0.5")
	options.MustHeader.Set("Accept-Language", "en-US,en;q=0.8")
	// options.MustCookies = append(options.MustCookies)
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

	if c.personalVideoList.MatchString(resp.Request.URL.Path) {
		return c.parsePersonalVideoList(ctx, resp, yield)
	} else if c.personalVideoJSONList.MatchString(resp.Request.URL.Path) {
		return c.parsePersonalVideoJSONList(ctx, resp, yield)
	} else if c.detailShortLinkPageReg.MatchString(resp.Request.URL.Path) ||
		c.detailPageReg.MatchString(resp.Request.URL.Path) ||
		c.detailInternalPageReg.MatchString(resp.Request.URL.Path) {
		return c.parseDetail(ctx, resp, yield)
	} else if c.downloadVideoReg.MatchString(resp.Request.URL.Path) {
		return c.download(ctx, resp, yield)
	}
	return fmt.Errorf("unsupported url %s", resp.Request.URL.String())
}

func (c *_Crawler) getCookies(ctx context.Context, rawUrl string) (string, int64, error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return "", 0, err
	}
	var (
		cookies, _ = c.httpClient.Jar().Cookies(ctx, u)
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
	if expiresAt.IsZero() {
		expiresAt = time.Now().Add(time.Hour)
	}
	return cookie, expiresAt.Unix(), nil
}

// nextIndex used to get sharingData from context
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")) + 1)
}

var propsDataReg = regexp.MustCompile(`(?U)<script\s+id="__NEXT_DATA__"\s+type="application/json"[^>]*>\s*(.*)\s*</script>`)

func (c *_Crawler) parsePersonalVideoList(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	vals := resp.Request.URL.Query()
	vals.Del("lang")
	if len(vals) > 0 {
		req := resp.Request
		req.URL.RawQuery = "lang=en"

		return yield(ctx, req)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	matched := propsDataReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		return fmt.Errorf("next data for url %s not found", resp.Request.URL)
	}

	var interData PropDataV1
	if err := json.Unmarshal(matched[1], &interData); err != nil {
		return err
	}

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

	lastIndex := nextIndex(ctx)
	for _, prop := range interData.Props.PageProps.Items {

		item := pbItem.Tiktok_Item{
			Source: &pbItem.Tiktok_Source{},
			Video:  &media.Media_Video{Cover: &media.Media_Image{}},
			Author: &pbItem.Tiktok_Author{},
			Headers: map[string]string{
				"Referer": resp.Request.URL.String(),
			},
			CrawledUtc: time.Now().Unix(),
		}
		item.Source.Id = prop.ID
		item.Source.PublishUtc = prop.CreateTime
		item.Title = prop.Desc
		if prop.Video.DownloadAddr != "" {
			item.Video.OriginalUrl = prop.Video.DownloadAddr
		} else if prop.Video.PlayAddr != "" {
			item.Video.OriginalUrl = prop.Video.PlayAddr
		} else {
			return fmt.Errorf("no download url found for %s", resp.Request.URL)
		}
		item.Video.Width = int32(prop.Video.Width)
		item.Video.Height = int32(prop.Video.Height)
		item.Video.Duration = int32(prop.Video.Duration)
		if prop.Video.OriginCover != "" {
			item.Video.Cover.OriginalUrl = prop.Video.OriginCover
		} else if prop.Video.Cover != "" {
			item.Video.Cover.OriginalUrl = prop.Video.Cover
		}
		item.Author.Id = prop.Author.ID
		item.Author.Name = prop.Author.Nickname
		item.Author.Icon = prop.Author.AvatarLarger

		cookie, expiresAt, err := c.getCookies(ctx, item.Video.OriginalUrl)
		if err != nil {
			return err
		}
		item.Headers["Cookie"] = cookie
		item.ExpiresUtc = expiresAt

		lastIndex += 1
		nctx := context.WithValue(ctx, "item.index", lastIndex)
		if err := yield(nctx, &item); err != nil {
			return err
		}
	}

	if interData.Props.PageProps.VideoListHasMore {
		u, _ := url.Parse("https://m.tiktok.com/api/post/item_list/")
		userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.74 Safari/537.36 Edg/79.0.309.43"
		vals := u.Query()
		vals.Set("aid", "1988")
		vals.Set("app_name", "tiktok_web")
		vals.Set("device_platform", "web")
		refererUrl := resp.Request.URL
		vals.Set("referer", refererUrl.String())
		refererUrl.RawQuery = ""
		vals.Set("root_referer", refererUrl.String()+"?")
		vals.Set("user_agent", userAgent)
		vals.Set("cookie_enabled", "true")
		vals.Set("screen_width", "1920")
		vals.Set("screen_height", "1080")
		vals.Set("browser_language", "en-US")
		vals.Set("browser_platform", "WindowsIntel")
		vals.Set("browser_name", "Mozilla")
		vals.Set("browser_version", "5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.74 Safari/537.36 Edg/79.0.309.43")
		vals.Set("browser_online", "true")
		vals.Set("ac", "4g")
		vals.Set("timezone_name", "America/Chicago")
		vals.Set("page_referer", refererUrl.String()+"?")
		vals.Set("priority_region", "")
		vals.Set("appId", "1233")
		vals.Set("region", "US")
		vals.Set("appType", "m")
		vals.Set("isAndroid", "false")
		vals.Set("isMobile", "false")
		vals.Set("isIOS", "false")
		vals.Set("OS", "windows")
		vals.Set("did", interData.Props.InitialProps.Wid)
		vals.Set("count", "30")
		vals.Set("cursor", strconv.Format(interData.Props.PageProps.VideoListMaxCursor))
		vals.Set("secUid", interData.Props.PageProps.UserInfo.User.SecUID)
		vals.Set("language", "en")

		u.RawQuery = vals.Encode()
		req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
		req.Header.Set("Referer", resp.Request.URL.String())

		nctx := context.WithValue(ctx, "item.index", lastIndex)
		return yield(nctx, req)
	}
	c.logger.Debugf("got %d count", len(interData.Props.PageProps.Items))
	return nil
}

func (c *_Crawler) parsePersonalVideoJSONList(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var respData struct {
		Cursor     string       `json:"cursor"`
		HasMore    bool         `json:"hasMore"`
		ItemList   []TiktokItem `json:"itemList"`
		StatusCode int          `json:"statusCode"`
	}
	if err := json.Unmarshal(respBody, &respData); err != nil {
		c.logger.Debugf("decode json response failed, error=%s", err)
		return err
	}
	if respData.StatusCode != 0 {
		return fmt.Errorf("api statusCode is %v", respData.StatusCode)
	}

	lastIndex := nextIndex(ctx)
	for _, prop := range respData.ItemList {
		item := pbItem.Tiktok_Item{
			Source: &pbItem.Tiktok_Source{
				Id:         prop.ID,
				PublishUtc: prop.CreateTime,
			},
			Title:  prop.Desc,
			Video:  &media.Media_Video{Cover: &media.Media_Image{}},
			Author: &pbItem.Tiktok_Author{},
			Headers: map[string]string{
				"Referer": resp.Request.Header.Get("Referer"),
			},
			CrawledUtc: time.Now().Unix(),
		}
		if prop.Video.DownloadAddr != "" {
			item.Video.OriginalUrl = prop.Video.DownloadAddr
		} else if prop.Video.PlayAddr != "" {
			item.Video.OriginalUrl = prop.Video.PlayAddr
		} else {
			return fmt.Errorf("no download url found for %s", resp.Request.URL)
		}
		item.Video.Width = int32(prop.Video.Width)
		item.Video.Height = int32(prop.Video.Height)
		item.Video.Duration = int32(prop.Video.Duration)
		if prop.Video.OriginCover != "" {
			item.Video.Cover.OriginalUrl = prop.Video.OriginCover
		} else if prop.Video.Cover != "" {
			item.Video.Cover.OriginalUrl = prop.Video.Cover
		}
		item.Author.Id = prop.Author.ID
		item.Author.Name = prop.Author.Nickname
		item.Author.Icon = prop.Author.AvatarLarger

		cookie, expiresAt, err := c.getCookies(ctx, item.Video.OriginalUrl)
		if err != nil {
			return err
		}
		item.Headers["Cookie"] = cookie
		item.ExpiresUtc = expiresAt

		lastIndex += 1
		nctx := context.WithValue(ctx, "item.index", lastIndex)
		if err := yield(nctx, &item); err != nil {
			return err
		}
	}

	if respData.HasMore {
		u := *resp.Request.URL
		vals := u.Query()
		vals.Set("cursor", strconv.Format(respData.Cursor))
		u.RawQuery = vals.Encode()

		req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
		if resp.Request.Header.Get("Referer") != "" {
			req.Header.Set("Referer", resp.Request.Header.Get("Referer"))
		} else {
			req.Header.Set("Referer", resp.Request.URL.Query().Get("referer"))
		}

		nctx := context.WithValue(ctx, "item.index", lastIndex)
		return yield(nctx, req)
	}
	return nil
}

var (
	videoCoverReg = regexp.MustCompile(`background-image:\s*url\("([^;]+)"\);`)
	initPropReg   = regexp.MustCompile(`<script\s*[^>]*\s*>\s*window.__INIT_PROPS__\s*=\s*([^\r\n]+);?\s*</script>`)
)

func (c *_Crawler) parseDetail(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	vals := resp.Request.URL.Query()
	vals.Del("lang")
	if len(vals) > 0 {
		req := resp.Request
		req.URL.RawQuery = "lang=en"

		return yield(ctx, req)
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

	// {
	// 	cookies := resp.Cookies()
	// 	nc := cookies[0:0]
	// 	for _, cookie := range cookies {
	// 		if cookie.Name != "tt_webid" && cookie.Name != "tt_webid_v2" {
	// 			cookie.MaxAge = -1
	// 			cookie.Expires = time.Time{}
	// 			nc = append(nc, cookie)
	// 		}
	// 	}
	// 	c.httpClient.Jar().SetCookies(ctx, resp.Request.URL, nc)

	// 	nreq := resp.Request.Clone(ctx)
	// 	nreq.Header.Del("Cookie")

	// 	if resp, err = c.httpClient.DoWithOptions(ctx, nreq, http.Options{EnableProxy: false}); err != nil {
	// 		return err
	// 	}
	// 	doc, err = goquery.NewDocumentFromReader(resp.Body)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

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
		mockCookieUrls := []string{
			"https://www.tiktok.com/secsdk_csrf_token",
			resp.Request.URL.String(),
		}
		/*
			reg := regexp.MustCompile(`<script\s+type="text/javascript"\s+src="(https://www.tiktok.com/akam/[a-z0-9]+/[a-z0-9]+)"\s+defer>\s*</script>`)
			if matched := reg.FindSubmatch(respBody); len(matched) > 1 {
				mockCookieUrls = append(mockCookieUrls, string(matched[1]))
			}
			reg = regexp.MustCompile(`src="(https?://www.tiktok.com/akam/[a-z0-9]+/pixel_[a-z0-9]+\?a=[a-zA-Z0-9=+-.]+)"`)
			if matched := reg.FindSubmatch(respBody); len(matched) > 1 {
				mockCookieUrls = append(mockCookieUrls, string(matched[1]))
			}

		opts := c.CrawlOptions()
		for _, u := range mockCookieUrls {
			c.logger.Debugf("mock cookie %s", u)
			req, err := http.NewRequest(http.MethodGet, u, nil)
			if err != nil {
				return err
			}
			if u != resp.Request.URL.String() {
				req.Header.Set("Referer", resp.Request.URL.String())
			}
			for k := range opts.MustHeader {
				req.Header.Set(k, opts.MustHeader.Get(k))
			}
			if _, err := c.httpClient.DoWithOptions(ctx, req, http.Options{EnableProxy: false}); err != nil {
				return err
			}
		}
	*/

	cookie, expiresAt, err := c.getCookies(ctx, item.Video.OriginalUrl)
	if err != nil {
		return err
	}
	item.Headers = map[string]string{
		"Referer": resp.Request.URL.String(),
		"Cookie":  cookie,
	}
	item.ExpiresUtc = expiresAt

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
		// "https://www.tiktok.com/@yessicarodriguez1023?lang=en",
		"https://www.tiktok.com/@willsmith?lang=en",
		// "https://vm.tiktok.com/ZScNvr6C/",
		// "https://www.tiktok.com/@kasey.jo.gerst/video/6923743895247506693?sender_device=mobile&sender_web_id=6926525695457117698&is_from_webapp=v2&is_copy_url=0",
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
	logger := glog.New(glog.LogLevelDebug)
	// build a http client
	// get proxy's microservice address from env
	client, err := proxy.NewProxyClient(os.Getenv("VOILA_PROXY_URL"), cookiejar.New(), logger)
	if err != nil {
		panic(err)
	}

	// instance the spider locally
	spider, err := New(client, logger)
	if err != nil {
		panic(err)
	}

	// this callback func is used to do recursion call of sub requests.
	var callback func(ctx context.Context, val interface{}) error
	callback = func(ctx context.Context, val interface{}) error {
		switch i := val.(type) {
		case *http.Request:
			logger.Debugf("Access %s", i.URL)
			opts := spider.CrawlOptions(i.URL)

			// process logic of sub request

			// init custom headers
			for k := range opts.MustHeader {
				i.Header.Set(k, opts.MustHeader.Get(k))
			}

			// init custom cookies
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

			// do http requests here.
			nctx, cancel := context.WithTimeout(ctx, time.Minute*5)
			defer cancel()
			resp, err := client.DoWithOptions(nctx, i, http.Options{
				EnableProxy:       true,
				EnableHeadless:    false,
				EnableSessionInit: opts.EnableSessionInit,
				KeepSession:       opts.KeepSession,
				Reliability:       opts.Reliability,
			})
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			return spider.Parse(ctx, resp, callback)
		default:
			// output the result
			data, err := json.Marshal(i)
			if err != nil {
				return err
			}
			logger.Infof("data: %s", data)
		}
		return nil
	}

	ctx := context.WithValue(context.Background(), "tracing_id", fmt.Sprintf("asos_%d", time.Now().UnixNano()))
	// start the crawl request
	for _, req := range spider.NewTestRequest(context.Background()) {
		if err := callback(ctx, req); err != nil {
			logger.Fatal(err)
		}
	}
}
