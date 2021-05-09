package main

import (
	"encoding/json"
	"fmt"
	rhttp "net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/voiladev/go-crawler/pkg/cli"
	"github.com/voiladev/go-crawler/pkg/context"
	"github.com/voiladev/go-crawler/pkg/crawler"
	"github.com/voiladev/go-crawler/pkg/item"
	"github.com/voiladev/go-crawler/pkg/net/http"
	pbItem "github.com/voiladev/go-crawler/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	pbProxy "github.com/voiladev/go-crawler/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/protoutil"
	"github.com/voiladev/go-framework/timeutil"
	"golang.org/x/oauth2"
	"google.golang.org/api/youtube/v3"
)

type YoutubeClientSecret struct {
	Installed struct {
		ClientID                string   `json:"client_id"`
		ProjectID               string   `json:"project_id"`
		AuthURI                 string   `json:"auth_uri"`
		TokenURI                string   `json:"token_uri"`
		AuthProviderX509CertURL string   `json:"auth_provider_x509_cert_url"`
		ClientSecret            string   `json:"client_secret"`
		RedirectUris            []string `json:"redirect_uris"`
	} `json:"installed"`
}

var (
	ytcs        *YoutubeClientSecret
	accessToken oauth2.Token
	ytConfig    *oauth2.Config
)

func init() {
	ytcs = &YoutubeClientSecret{}
	ytcs.Installed.ProjectID = os.Getenv("YOUTUBE_PROJECT_ID")
	if ytcs.Installed.ProjectID == "" {
		ytcs.Installed.ProjectID = "voilabackconnect"
	}
	ytcs.Installed.ClientID = os.Getenv("YOUTUBE_CLIENT_ID")
	if ytcs.Installed.ClientID == "" {
		ytcs.Installed.ClientID = "1040955902703-flufp9r93u0tdo2lfdourj36v42arhus.apps.googleusercontent.com"
	}
	ytcs.Installed.ClientSecret = os.Getenv("YOUTUBE_CLIENT_SECRET")
	if ytcs.Installed.ClientSecret == "" {
		panic("env YOUTUBE_CLIENT_SECRET not specified")
	}
	ytcs.Installed.AuthURI = "https://accounts.google.com/o/oauth2/auth"
	ytcs.Installed.TokenURI = "https://oauth2.googleapis.com/token"
	ytcs.Installed.AuthProviderX509CertURL = "https://www.googleapis.com/oauth2/v1/certs"
	ytcs.Installed.RedirectUris = []string{"urn:ietf:wg:oauth:2.0:oob", "http://localhost"}

	data, err := os.ReadFile("/youtube/youtube-oauth2.json")
	if os.IsNotExist(err) {
		usr, err := user.Current()
		if err != nil {
			panic(err)
		}
		tokenFile := filepath.Join(usr.HomeDir, ".credentials/youtube-oauth2.json")
		data, err = os.ReadFile(tokenFile)
	}
	if len(data) == 0 {
		panic("AccessToken not found")
	}

	if err := json.Unmarshal(data, &accessToken); err != nil {
		panic("invalid access token")
	}

	ytConfig = &oauth2.Config{
		ClientID:     ytcs.Installed.ClientID,
		ClientSecret: ytcs.Installed.ClientSecret,
		RedirectURL:  ytcs.Installed.RedirectUris[0],
		Scopes:       []string{youtube.YoutubeReadonlyScope},
		Endpoint: oauth2.Endpoint{
			AuthURL:  ytcs.Installed.AuthURI,
			TokenURL: ytcs.Installed.TokenURI,
		},
	}

	item.Register(&pbItem.Youtube_Channel{})
}

type _Crawler struct {
	crawler.MustImplementCrawler

	httpClient   http.Client
	ytHttpClient *rhttp.Client
	ytService    *youtube.Service

	logger glog.Log
}

func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient:   client,
		ytHttpClient: ytConfig.Client(context.Background(), &accessToken),
		logger:       logger.New("_Crawler"),
	}

	var err error
	if c.ytService, err = youtube.New(c.ytHttpClient); err != nil {
		return nil, err
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "7acad725b8f010f612fad38ce0acfe2a"
}

// Version
func (c *_Crawler) Version() int32 {
	return 1
}

// CrawlOptions
func (c *_Crawler) CrawlOptions(u *url.URL) *crawler.CrawlOptions {
	options := crawler.NewCrawlOptions()
	options.EnableHeadless = false
	options.Reliability = pbProxy.ProxyReliability_ReliabilityMedium
	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.youtube.com"}
}

func (c *_Crawler) Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	if crawler.IsTargetTypeSupported(ctx, &pbItem.Youtube_Channel{}) {
		return c.parseAuthorInfo(ctx, resp, yield)
	}
	return crawler.ErrUnsupportedTarget
}

var (
	usernamePathReg     = regexp.MustCompile(`^/user/([a-zA-Z0-9_\-\pL\pS]+)(?:/[a-zA-Z0-9_\-\pL\pS]+){0,3}$`)
	channelPathReg      = regexp.MustCompile(`^/channel/([a-zA-Z0-9_]+)(?:/[a-zA-Z0-9_\-\pL\pS]+){0,3}$`)
	channelAliasPathReg = regexp.MustCompile(`^/c/([a-zA-Z0-9_\-\pL\pS]+)(?:/[a-zA-Z0-9_\-\pL\pS]+){0,3}$`)
)

func (c *_Crawler) parseAuthorInfo(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	var (
		path      = resp.Request.URL.Path
		channelId string
		username  string
	)
	// extract channelId
	if channelAliasPathReg.MatchString(path) {
		// no matter which page it is, there must exists the channel id
		dom, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			c.logger.Error(err)
			return err
		}
		href := dom.Find(`link[rel="canonical"]`).AttrOr("href", "")
		if href == "" {
			return fmt.Errorf("no canonical url found")
		}
		u, err := url.Parse(href)
		if err != nil {
			return fmt.Errorf("invalid url %s", href)
		}
		path = u.Path
	}
	if channelPathReg.MatchString(path) {
		matched := channelPathReg.FindStringSubmatch(path)
		channelId = matched[1]
	} else if usernamePathReg.MatchString(path) {
		matched := usernamePathReg.FindStringSubmatch(path)
		username = matched[1]
	} else {
		return fmt.Errorf("no channel id or username found")
	}

	call := c.ytService.Channels.List([]string{"snippet,contentDetails,statistics"}).Context(ctx)
	if username != "" {
		call.ForUsername(username)
	}
	if channelId != "" {
		call.Id(channelId)
	}
	call.MaxResults(1)
	infoResp, err := call.Do()
	if err != nil {
		c.logger.Error(err)
		return err
	}
	if len(infoResp.Items) == 0 {
		return fmt.Errorf("no youtube channel found")
	}
	item := infoResp.Items[0]
	snippet := item.Snippet
	stats := item.Statistics
	channel := pbItem.Youtube_Channel{
		Id:           item.Id,
		Username:     snippet.CustomUrl,
		Title:        snippet.Title,
		Description:  snippet.Description,
		Avatar:       snippet.Thumbnails.High.Url,
		Country:      snippet.Country,
		PublishedUtc: timeutil.TimeParse(snippet.PublishedAt).Unix(),
		Stats: &pbItem.Youtube_Channel_Stats{
			SubscribeCount: int32(stats.SubscriberCount),
			VideoCount:     int32(stats.VideoCount),
			ViewCount:      int32(stats.ViewCount),
			CommentCount:   int32(stats.CommentCount),
		},
	}
	return yield(ctx, &channel)
}

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	channelType := protoutil.GetTypeUrl(&pbItem.Youtube_Channel{})
	for _, u := range [][2]string{
		{"https://www.youtube.com/c/AkilaZhang/featured", channelType},
		{"https://www.youtube.com/c/AkilaZhang", channelType},
	} {
		nctx := context.WithValue(ctx, crawler.TargetTypeKey, u[1])
		req, _ := http.NewRequestWithContext(nctx, http.MethodGet, u[0], nil)
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
	cli.NewApp(New).Run(os.Args)
}
