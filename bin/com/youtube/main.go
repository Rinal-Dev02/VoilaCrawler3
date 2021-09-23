package main

// Referer: https://developers.google.com/youtube/v3/quickstart/go
// TODO: added multi key support

import (
	"encoding/json"
	"errors"
	"fmt"
	rhttp "net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	cmdCli "github.com/urfave/cli/v2"
	"github.com/voiladev/VoilaCrawler/pkg/cli"
	"github.com/voiladev/VoilaCrawler/pkg/context"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/media"
	pbItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	pbProxy "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/protoutil"
	"github.com/voiladev/go-framework/timeutil"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type _Crawler struct {
	crawler.MustImplementCrawler

	httpClient   http.Client
	ytHttpClient *rhttp.Client
	ytService    *youtube.Service

	usernamePathReg     *regexp.Regexp
	channelPathReg      *regexp.Regexp
	channelAliasPathReg *regexp.Regexp
	videoPathReg        *regexp.Regexp

	logger glog.Log
}

func (_ *_Crawler) New(cli *cli.Context, client http.Client, logger glog.Log) (crawler.Crawler, error) {
	key := cli.String("youtube-key")
	if key == "" {
		return nil, errors.New("invalid youtube key")
	}
	svc, err := youtube.NewService(cli.Context, option.WithAPIKey(key))
	if err != nil {
		return nil, err
	}

	c := _Crawler{
		httpClient:          client,
		usernamePathReg:     regexp.MustCompile(`^/user/([^/]+)(?:/[^/]*){0,3}$`),
		channelPathReg:      regexp.MustCompile(`^/channel/([^/]+)(?:/[^/]*){0,3}$`),
		channelAliasPathReg: regexp.MustCompile(`^/c/([^/]+)(?:/[^/]*){0,3}$`),
		videoPathReg:        regexp.MustCompile(`^/watch$`),
		ytService:           svc,
		logger:              logger.New("_Crawler"),
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
	opts := crawler.NewCrawlOptions()
	opts.EnableHeadless = false
	opts.Reliability = pbProxy.ProxyReliability_ReliabilityMedium

	if c.usernamePathReg.MatchString(u.Path) ||
		c.channelPathReg.MatchString(u.Path) ||
		c.videoPathReg.MatchString(u.Path) {

		opts.SkipDoRequest = true
	}
	return opts
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

	if c.channelAliasPathReg.MatchString(resp.RawUrl().Path) {
		// extract channel id
		doc, err := resp.Selector()
		if err != nil {
			c.logger.Error(err)
			return err
		}
		href := doc.Find(`link[rel="canonical"]`).AttrOr("href", "")
		u, err := url.Parse(href)
		if err != nil {
			c.logger.Errorf("invalid href %s for %s", href, resp.RawUrl())
			return err
		}
		resp.Request.URL = u
	}

	if c.videoPathReg.MatchString(resp.RawUrl().Path) {
		return c.parseVideo(ctx, resp, yield)
	} else if c.channelPathReg.MatchString(resp.RawUrl().Path) {
		return c.parseChannel(ctx, resp, yield)
	}
	return crawler.ErrUnsupportedTarget
}

type YTCursor struct {
	Playlists []string `json:"p"`
	PageToken string   `json:"t"`
}

func (c *_Crawler) parseChannel(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}
	query := resp.Request.URL.Query()
	rawCursor := query.Get("_yt_cursor")
	var cursor YTCursor
	if rawCursor != "" && rawCursor != "{}" {
		if err := json.Unmarshal([]byte(rawCursor), &cursor); err != nil {
			c.logger.Errorf("parse cursor failed, error=%s", err)
			return fmt.Errorf("cursor is invalid")
		}
	}

	var channel *pbItem.Youtube_Channel
	if len(cursor.Playlists) == 0 {
		var (
			channelId string
			username  string
		)
		if matched := c.usernamePathReg.FindStringSubmatch(resp.RawUrl().Path); len(matched) == 2 {
			username = matched[1]
		} else if matched := c.channelPathReg.FindStringSubmatch(resp.RawUrl().Path); len(matched) == 2 {
			channelId = matched[1]
		}
		if username == "" && channelId == "" {
			return errors.New("no username or channel id found")
		}

		{
			call := c.ytService.Channels.List([]string{"id,snippet,statistics"}).Context(ctx)
			if username != "" {
				call = call.ForUsername(username)
			} else {
				call = call.Id(channelId)
			}
			call = call.MaxResults(50)
			_resp, err := call.Do()
			if err != nil {
				c.logger.Error(err)
				return err
			}
			if len(_resp.Items) > 0 {
				if username != "" {
					c.logger.Infof("got %d channels for user %s", len(_resp.Items), username)
				}
				// 每个用户，只选择第一个Channel
				channelId = _resp.Items[0].Id
				channel = &pbItem.Youtube_Channel{
					Id:           channelId,
					CanonicalUrl: fmt.Sprintf("https://www.youtube.com/channel/%s", channelId),
					Username:     _resp.Items[0].Snippet.CustomUrl,
					Title:        _resp.Items[0].Snippet.Title,
					Description:  _resp.Items[0].Snippet.Description,
					Country:      _resp.Items[0].Snippet.Country,
					PublishedUtc: timeutil.TimeParse(_resp.Items[0].Snippet.PublishedAt).Unix(),
					Stats: &pbItem.Youtube_Channel_Stats{
						SubscribeCount: int32(_resp.Items[0].Statistics.SubscriberCount),
						VideoCount:     int32(_resp.Items[0].Statistics.VideoCount),
						ViewCount:      int32(_resp.Items[0].Statistics.ViewCount),
						CommentCount:   int32(_resp.Items[0].Statistics.CommentCount),
					},
				}
				if _resp.Items[0].Snippet.Thumbnails.High != nil {
					channel.Avatar = _resp.Items[0].Snippet.Thumbnails.High.Url
				} else if _resp.Items[0].Snippet.Thumbnails.Standard != nil {
					channel.Avatar = _resp.Items[0].Snippet.Thumbnails.Standard.Url
				} else if _resp.Items[0].Snippet.Thumbnails.Medium != nil {
					channel.Avatar = _resp.Items[0].Snippet.Thumbnails.Medium.Url
				}
			}
		}
		if channelId == "" {
			return fmt.Errorf("no valid channel found")
		}

		// get playlists for channel
		call := c.ytService.Playlists.List([]string{"id,snippet,contentDetails"}).Context(ctx)
		call.MaxResults(50)
		call.ChannelId(channelId)
		_resp, err := call.Do()
		if err != nil {
			c.logger.Error(err)
			return err
		}
		for _, item := range _resp.Items {
			cursor.Playlists = append(cursor.Playlists, item.Id)
		}
	}

	c.logger.Debugf("%d playlists", len(cursor.Playlists))

	var videos []*pbItem.Youtube_Video
	for _, plid := range cursor.Playlists {
		vs, token, err := c.getVideosByPlaylist(ctx, plid, cursor.PageToken)
		if err != nil {
			c.logger.Error(err)
			return err
		}
		cursor.PageToken = token
		for _, v := range vs {
			v.Channel = channel
			videos = append(videos, v)
		}
		break
	}
	if cursor.PageToken == "" {
		cursor.Playlists = append(cursor.Playlists[0:0], cursor.Playlists[1:]...)
	}

	for _, video := range videos {
		if err := yield(ctx, video); err != nil {
			c.logger.Error(err)
			return err
		}
	}
	data, _ := json.Marshal(cursor)
	u := resp.RawUrl()
	vals := u.Query()
	vals.Set("_yt_cursor", fmt.Sprintf("%s", data))
	u.RawQuery = vals.Encode()

	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	return yield(ctx, req)
}

func (c *_Crawler) parseVideo(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}
	vals := resp.RawUrl().Query()
	listId, videoId := vals.Get("list"), vals.Get("v")

	if listId != "" {
		query := resp.Request.URL.Query()
		rawCursor := query.Get("_yt_cursor")
		var cursor YTCursor
		if rawCursor != "" && rawCursor != "{}" {
			if err := json.Unmarshal([]byte(rawCursor), &cursor); err != nil {
				c.logger.Errorf("parse cursor failed, error=%s", err)
				return fmt.Errorf("cursor is invalid")
			}
		}
		if len(cursor.Playlists) == 0 {
			cursor.Playlists = append(cursor.Playlists, listId)
			cursor.PageToken = ""
		}

		videos, token, err := c.getVideosByPlaylist(ctx, cursor.Playlists[0], cursor.PageToken)
		if err != nil {
			c.logger.Error(err)
			return err
		}
		for _, video := range videos {
			if err := yield(ctx, video); err != nil {
				c.logger.Error(err)
				return err
			}
		}
		cursor.PageToken = token
		if cursor.PageToken != "" {
			data, _ := json.Marshal(cursor)
			u := resp.RawUrl()
			vals := u.Query()
			vals.Set("_yt_cursor", fmt.Sprintf("%s", data))
			u.RawQuery = vals.Encode()

			req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
			if err := yield(ctx, req); err != nil {
				return err
			}
		}
	} else {
		call := c.ytService.Videos.List([]string{"id,snippet,contentDetails,player"}).Context(ctx)
		call = call.Id(videoId)
		_resp, err := call.Do()
		if err != nil {
			c.logger.Error(err)
			return err
		}
		for _, vitem := range _resp.Items {
			item, err := c.videoUnmarshal(vitem)
			if err != nil {
				c.logger.Error(err)
				return err
			}
			if err := yield(ctx, item); err != nil {
				c.logger.Error(err)
				return err
			}
		}
	}
	return nil
}

func (c *_Crawler) getVideosByPlaylist(ctx context.Context, plId string, token string) ([]*pbItem.Youtube_Video, string, error) {
	c.logger.Infof("get videos of playlist %s, nexttoken=%s", plId, token)
	var vids []string
	{
		call := c.ytService.PlaylistItems.List([]string{"snippet"}).Context(ctx)
		call = call.PlaylistId(plId)
		call = call.PageToken(token)
		call = call.MaxResults(50)
		_resp, err := call.Do()
		if err != nil {
			c.logger.Error(err)
			return nil, "", err
		}
		for _, item := range _resp.Items {
			vids = append(vids, item.Snippet.ResourceId.VideoId)
		}
		token = _resp.NextPageToken
	}

	call := c.ytService.Videos.List([]string{"id,snippet,statistics,player"}).Context(ctx)
	call = call.Id(vids...)
	_resp, err := call.Do()
	if err != nil {
		c.logger.Error(err)
		return nil, "", err
	}

	var videos []*pbItem.Youtube_Video
	for _, vitem := range _resp.Items {
		item, err := c.videoUnmarshal(vitem)
		if err != nil {
			c.logger.Error(err)
			return nil, "", err
		}
		videos = append(videos, item)
	}
	c.logger.Infof("next=%s", token)
	return videos, token, nil
}

func (c *_Crawler) parseAuthorInfo(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	var (
		channel   *pbItem.Youtube_Channel
		username  string
		channelId string
	)
	if matched := c.usernamePathReg.FindStringSubmatch(resp.RawUrl().Path); len(matched) == 2 {
		username = matched[1]
	} else if matched := c.channelPathReg.FindStringSubmatch(resp.RawUrl().Path); len(matched) == 2 {
		channelId = matched[1]
	}
	if username == "" && channelId == "" {
		return errors.New("no username or channel id found")
	}

	call := c.ytService.Channels.List([]string{"id,snippet,statistics"}).Context(ctx)
	if username != "" {
		call = call.ForUsername(username)
	} else {
		call = call.Id(channelId)
	}
	call = call.MaxResults(1)
	_resp, err := call.Do()
	if err != nil {
		c.logger.Error(err)
		return err
	}
	if len(_resp.Items) > 0 {
		if username != "" {
			c.logger.Infof("got %d channels for user %s", len(_resp.Items), username)
		}
		// 每个用户，只选择第一个Channel
		channelId = _resp.Items[0].Id
		channel = &pbItem.Youtube_Channel{
			Id:           channelId,
			CanonicalUrl: fmt.Sprintf("https://www.youtube.com/channel/%s", channelId),
			Username:     _resp.Items[0].Snippet.CustomUrl,
			Title:        _resp.Items[0].Snippet.Title,
			Description:  _resp.Items[0].Snippet.Description,
			Country:      _resp.Items[0].Snippet.Country,
			PublishedUtc: timeutil.TimeParse(_resp.Items[0].Snippet.PublishedAt).Unix(),
			Stats: &pbItem.Youtube_Channel_Stats{
				SubscribeCount: int32(_resp.Items[0].Statistics.SubscriberCount),
				VideoCount:     int32(_resp.Items[0].Statistics.VideoCount),
				ViewCount:      int32(_resp.Items[0].Statistics.ViewCount),
				CommentCount:   int32(_resp.Items[0].Statistics.CommentCount),
			},
		}
		if _resp.Items[0].Snippet.Thumbnails.High != nil {
			channel.Avatar = _resp.Items[0].Snippet.Thumbnails.High.Url
		} else if _resp.Items[0].Snippet.Thumbnails.Standard != nil {
			channel.Avatar = _resp.Items[0].Snippet.Thumbnails.Standard.Url
		} else if _resp.Items[0].Snippet.Thumbnails.Medium != nil {
			channel.Avatar = _resp.Items[0].Snippet.Thumbnails.Medium.Url
		}
	}
	if channel != nil {
		return yield(ctx, channel)
	}
	return nil
}

func getThumbnail(nails ...*youtube.Thumbnail) string {
	ns := nails[0:0]
	for _, nail := range nails {
		if nail != nil {
			ns = append(ns, nail)
		}
	}
	if len(ns) == 0 {
		return ""
	}
	return ns[0].Url
}

func (c *_Crawler) videoUnmarshal(_i interface{}) (*pbItem.Youtube_Video, error) {
	switch i := _i.(type) {
	case *youtube.Video:
		item := pbItem.Youtube_Video{
			Source: &pbItem.Youtube_Source{
				Id:         i.Id,
				SourceUrl:  "https://www.youtube.com/watch?v=" + i.Id,
				PublishUtc: timeutil.TimeParse(i.Snippet.PublishedAt).Unix(),
			},
			Title:       strings.TrimSpace(i.Snippet.Title),
			Description: strings.TrimSpace(i.Snippet.Description),
			Channel: &pbItem.Youtube_Channel{
				Id:    i.Snippet.ChannelId,
				Title: i.Snippet.ChannelTitle,
			},
			Player: &pbItem.Youtube_Video_Player{
				EmbedHtml: i.Player.EmbedHtml,
				Video: &media.Media_Video{
					Cover: &media.Media_Image{
						OriginalUrl: getThumbnail(i.Snippet.Thumbnails.Standard, i.Snippet.Thumbnails.High, i.Snippet.Thumbnails.Maxres, i.Snippet.Thumbnails.Default),
						LargeUrl:    getThumbnail(i.Snippet.Thumbnails.High, i.Snippet.Thumbnails.Maxres, i.Snippet.Thumbnails.Standard, i.Snippet.Thumbnails.Medium),
						MediumUrl:   getThumbnail(i.Snippet.Thumbnails.Standard, i.Snippet.Thumbnails.High, i.Snippet.Thumbnails.Maxres, i.Snippet.Thumbnails.Medium),
						SmallUrl:    getThumbnail(i.Snippet.Thumbnails.Standard, i.Snippet.Thumbnails.High, i.Snippet.Thumbnails.Maxres, i.Snippet.Thumbnails.Medium),
					},
				},
			},
			Stats: &pbItem.Youtube_Video_Stats{
				ViewCount:    int32(i.Statistics.ViewCount),
				CommentCount: int32(i.Statistics.CommentCount),
				LikeCount:    int32(i.Statistics.LikeCount),
				DislikeCount: int32(i.Statistics.DislikeCount),
			},
			CrawledUtc: time.Now().Unix(),
		}
		return &item, nil
	}
	return nil, nil
}

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	channelType := protoutil.GetTypeUrl(&pbItem.Youtube_Channel{})
	for _, u := range [][2]string{
		{"https://www.youtube.com/channel/UC1dVfl5-I98WX3yCy8IJQMg", channelType},
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
	cli.NewApp(
		&_Crawler{},
		&cmdCli.StringFlag{Name: "youtube-key", Usage: "youtube api key", Required: true},
	).Run(os.Args)
}
