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
	"github.com/voiladev/VoilaCrawl/pkg/net/http/proxycrawl"
	"github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/media"
	pbItem "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
)

type _Crawler struct {
	httpClient http.Client

	detailPageReg          *regexp.Regexp
	detailShortLinkPageReg *regexp.Regexp

	logger glog.Log
}

func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient:             client,
		detailPageReg:          regexp.MustCompile(`^/@[0-9a-zA-Z-_]+/video/[0-9]+/?$`),
		detailShortLinkPageReg: regexp.MustCompile(`^/[a-zA-Z0-9]+/?$`),
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
	return []string{"vm.tiktok.com", "www.tiktok.com"}
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

	if c.detailShortLinkPageReg.MatchString(resp.Request.URL.Path) || c.detailPageReg.MatchString(resp.Request.URL.Path) {
		return c.parseDetail(ctx, resp, yield)
	}
	return fmt.Errorf("unsupported url %s", resp.Request.URL.String())
}

var videoCoverReg = regexp.MustCompile(`background-image:\s*url\("([^;]+)"\);`)

type PropData struct {
	Props struct {
		InitialProps struct {
			StatusCode int    `json:"statusCode"`
			FullURL    string `json:"$fullUrl"`
			CsrfToken  string `json:"$csrfToken"`
		} `json:"initialProps"`
		PageProps struct {
			Key        string `json:"key"`
			ServerCode int    `json:"serverCode"`
			StatusCode int    `json:"statusCode"`
			StatusMsg  string `json:"statusMsg"`
			ItemInfo   struct {
				ItemStruct struct {
					ID         string `json:"id"`
					Desc       string `json:"desc"`
					CreateTime int    `json:"createTime"`
					Video      struct {
						ID           string   `json:"id"`
						Height       int      `json:"height"`
						Width        int      `json:"width"`
						Duration     int      `json:"duration"`
						Ratio        string   `json:"ratio"`
						Cover        string   `json:"cover"`
						OriginCover  string   `json:"originCover"`
						DynamicCover string   `json:"dynamicCover"`
						PlayAddr     string   `json:"playAddr"`
						DownloadAddr string   `json:"downloadAddr"`
						ShareCover   []string `json:"shareCover"`
						ReflowCover  string   `json:"reflowCover"`
					} `json:"video"`
					Author struct {
						ID             string `json:"id"`
						ShortID        string `json:"shortId"`
						UniqueID       string `json:"uniqueId"`
						Nickname       string `json:"nickname"`
						AvatarLarger   string `json:"avatarLarger"`
						AvatarMedium   string `json:"avatarMedium"`
						AvatarThumb    string `json:"avatarThumb"`
						Signature      string `json:"signature"`
						CreateTime     int    `json:"createTime"`
						Verified       bool   `json:"verified"`
						SecUID         string `json:"secUid"`
						Ftc            bool   `json:"ftc"`
						Relation       int    `json:"relation"`
						OpenFavorite   bool   `json:"openFavorite"`
						CommentSetting int    `json:"commentSetting"`
						DuetSetting    int    `json:"duetSetting"`
						StitchSetting  int    `json:"stitchSetting"`
						PrivateAccount bool   `json:"privateAccount"`
						Secret         bool   `json:"secret"`
						RoomID         string `json:"roomId"`
					} `json:"author"`
					Music struct {
						ID                 string `json:"id"`
						Title              string `json:"title"`
						PlayURL            string `json:"playUrl"`
						CoverLarge         string `json:"coverLarge"`
						CoverMedium        string `json:"coverMedium"`
						CoverThumb         string `json:"coverThumb"`
						AuthorName         string `json:"authorName"`
						Original           bool   `json:"original"`
						Duration           int    `json:"duration"`
						Album              string `json:"album"`
						ScheduleSearchTime int    `json:"scheduleSearchTime"`
					} `json:"music"`
					Stats struct {
						DiggCount    int `json:"diggCount"`
						ShareCount   int `json:"shareCount"`
						CommentCount int `json:"commentCount"`
						PlayCount    int `json:"playCount"`
					} `json:"stats"`
					IsActivityItem bool `json:"isActivityItem"`
					DuetInfo       struct {
						DuetFromID string `json:"duetFromId"`
					} `json:"duetInfo"`
					OriginalItem      bool `json:"originalItem"`
					OfficalItem       bool `json:"officalItem"`
					Secret            bool `json:"secret"`
					ForFriend         bool `json:"forFriend"`
					Digged            bool `json:"digged"`
					ItemCommentStatus int  `json:"itemCommentStatus"`
					ShowNotPass       bool `json:"showNotPass"`
					Vl1               bool `json:"vl1"`
					TakeDown          int  `json:"takeDown"`
					ItemMute          bool `json:"itemMute"`
					AuthorStats       struct {
						FollowerCount  int `json:"followerCount"`
						FollowingCount int `json:"followingCount"`
						Heart          int `json:"heart"`
						HeartCount     int `json:"heartCount"`
						VideoCount     int `json:"videoCount"`
						DiggCount      int `json:"diggCount"`
					} `json:"authorStats"`
					PrivateItem   bool `json:"privateItem"`
					DuetEnabled   bool `json:"duetEnabled"`
					StitchEnabled bool `json:"stitchEnabled"`
					IsAd          bool `json:"isAd"`
					ShareEnabled  bool `json:"shareEnabled"`
				} `json:"itemStruct"`
			} `json:"itemInfo"`
		} `json:"pageProps"`
	} `json:"props"`
	BuildID      string `json:"buildId"`
	CustomServer bool   `json:"customServer"`
	Gip          bool   `json:"gip"`
	AppGip       bool   `json:"appGip"`
}

func (c *_Crawler) parseDetail(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	c.logger.Debugf("%s", respBody)

	var (
		rawurl string
		exists bool
		item   = pbItem.Tiktok_Item{
			Video:    &media.Media_Video{Cover: &media.Media_Image{}},
			AuthInfo: &pbItem.Tiktok_Item_AuthInfo{},
		}
	)

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}
	if rawProp := doc.Find("#__NEXT_DATA__").Text(); strings.TrimSpace(rawProp) != "" {
		var prop PropData
		if err := json.Unmarshal([]byte(strings.TrimSpace(rawProp)), &prop); err != nil {
			return err
		}
		video := prop.Props.PageProps.ItemInfo.ItemStruct.Video

		item.Title = prop.Props.PageProps.ItemInfo.ItemStruct.Desc
		if video.PlayAddr == "" && video.DownloadAddr == "" {
			return fmt.Errorf("video url not found for %s", resp.Request.URL)
		}
		item.Video.Id = video.ID
		item.Video.OriginalUrl = video.PlayAddr
		if item.Video.OriginalUrl == "" {
			item.Video.OriginalUrl = video.DownloadAddr
		}
		item.Video.Width = int32(video.Width)
		item.Video.Height = int32(video.Height)
		item.Video.Duration = int32(video.Duration)
		item.Video.Cover.OriginalUrl = video.Cover
		if video.Cover == "" {
			item.Video.Cover.OriginalUrl = video.OriginCover
		}
		rawurl = prop.Props.InitialProps.FullURL
	} else {
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

	var (
		cookies   string
		expiresAt time.Time
	)
	if req, err := http.NewRequest(http.MethodGet, "https://www.tiktok.com/manifest.json", nil); err != nil {
		return err
	} else if resp, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
		EnableProxy:        true,
		DisableBackconnect: true,
	}); err != nil {
		return err
	} else {
		for _, c := range resp.Cookies() {
			v := fmt.Sprintf("%s=%s", c.Name, c.Value)
			if cookies == "" {
				cookies = v
			} else {
				cookies = cookies + "; " + v
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
	}
	item.AuthInfo.Cookies = cookies
	if expiresAt.IsZero() {
		item.AuthInfo.ExpiresAt = time.Now().Add(time.Hour * 24).Unix()
	} else {
		item.AuthInfo.ExpiresAt = expiresAt.Unix()
	}
	return yield(ctx, &item)
}

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		"https://vm.tiktok.com/ZScNvr6C/",
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
	client, err := proxycrawl.NewProxyCrawlClient(logger,
		proxycrawl.Options{APIToken: apiToken, JSToken: jsToken},
	)
	if err != nil {
		panic(err)
	}

	spider, err := New(client, logger)
	if err != nil {
		panic(err)
	}
	opts := spider.CrawlOptions()

	for _, req := range spider.NewTestRequest(context.Background()) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
		defer cancel()

		logger.Debugf("Access %s", req.URL)
		for k := range opts.MustHeader {
			req.Header.Set(k, opts.MustHeader.Get(k))
		}
		for _, c := range opts.MustCookies {
			if strings.HasPrefix(req.URL.Path, c.Path) || c.Path == "" {
				val := fmt.Sprintf("%s=%s", c.Name, c.Value)
				if c := req.Header.Get("Cookie"); c != "" {
					req.Header.Set("Cookie", c+"; "+val)
				} else {
					req.Header.Set("Cookie", val)
				}
			}
		}

		resp, err := client.DoWithOptions(ctx, req, http.Options{EnableProxy: true, DisableBackconnect: true})
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		logger.Debugf("req url: %s", resp.Request.URL)

		if err := spider.Parse(ctx, resp, func(ctx context.Context, val interface{}) error {
			switch i := val.(type) {
			case *http.Request:
				logger.Infof("new request %s", i.URL)
			default:
				data, err := json.Marshal(i)
				if err != nil {
					return err
				}
				logger.Infof("data: %s", data)
			}
			return nil
		}); err != nil {
			panic(err)
		}
	}
}
