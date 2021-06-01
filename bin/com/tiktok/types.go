package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/media"
	pbItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
)

type TiktokItem struct {
	ID         string `json:"id"`
	Desc       string `json:"desc"`
	CreateTime int64  `json:"createTime"`
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
		CreateTime     int64  `json:"createTime"`
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
		ScheduleSearchTime int64  `json:"scheduleSearchTime"`
	} `json:"music"`
	Challenges []struct {
		ID            string `json:"id"`
		Title         string `json:"title"`
		Desc          string `json:"desc"`
		ProfileLarger string `json:"profileLarger"`
		ProfileMedium string `json:"profileMedium"`
		ProfileThumb  string `json:"profileThumb"`
		CoverLarger   string `json:"coverLarger"`
		CoverMedium   string `json:"coverMedium"`
		CoverThumb    string `json:"coverThumb"`
		IsCommerce    bool   `json:"isCommerce"`
	} `json:"challenges"`
	Stats struct {
		DiggCount    int `json:"diggCount"`
		ShareCount   int `json:"shareCount"`
		CommentCount int `json:"commentCount"`
		PlayCount    int `json:"playCount"`
	} `json:"stats"`
	DuetInfo struct {
		DuetFromID string `json:"duetFromId"`
	} `json:"duetInfo"`
	WarnInfo     []interface{} `json:"warnInfo"`
	OriginalItem bool          `json:"originalItem"`
	OfficalItem  bool          `json:"officalItem"`
	TextExtra    []struct {
		AwemeID      string `json:"awemeId"`
		Start        int    `json:"start"`
		End          int    `json:"end"`
		HashtagID    string `json:"hashtagId"`
		HashtagName  string `json:"hashtagName"`
		Type         int    `json:"type"`
		UserID       string `json:"userId"`
		IsCommerce   bool   `json:"isCommerce"`
		UserUniqueID string `json:"userUniqueId"`
		SecUID       string `json:"secUid"`
	} `json:"textExtra"`
	Secret            bool          `json:"secret"`
	ForFriend         bool          `json:"forFriend"`
	Digged            bool          `json:"digged"`
	ItemCommentStatus int           `json:"itemCommentStatus"`
	ShowNotPass       bool          `json:"showNotPass"`
	Vl1               bool          `json:"vl1"`
	TakeDown          int           `json:"takeDown"`
	ItemMute          bool          `json:"itemMute"`
	EffectStickers    []interface{} `json:"effectStickers"`
	AuthorStats       struct {
		FollowerCount  int `json:"followerCount"`
		FollowingCount int `json:"followingCount"`
		Heart          int `json:"heart"`
		HeartCount     int `json:"heartCount"`
		VideoCount     int `json:"videoCount"`
		DiggCount      int `json:"diggCount"`
	} `json:"authorStats"`
	PrivateItem    bool          `json:"privateItem"`
	DuetEnabled    bool          `json:"duetEnabled"`
	StitchEnabled  bool          `json:"stitchEnabled"`
	StickersOnItem []interface{} `json:"stickersOnItem"`
	IsAd           bool          `json:"isAd"`
	ShareEnabled   bool          `json:"shareEnabled"`
}

type PropDataV1 struct {
	Props struct {
		InitialProps struct {
			StatusCode int    `json:"statusCode"`
			FullURL    string `json:"$fullUrl"`
			CsrfToken  string `json:"$csrfToken"`
			Wid        string `json:"$wid"`
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
					CreateTime int64  `json:"createTime"`
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
						CreateTime     int64  `json:"createTime"`
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
						ScheduleSearchTime int64  `json:"scheduleSearchTime"`
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
			UserInfo struct {
				User struct {
					ID           string `json:"id"`
					ShortID      string `json:"shortId"`
					UniqueID     string `json:"uniqueId"`
					Nickname     string `json:"nickname"`
					AvatarLarger string `json:"avatarLarger"`
					AvatarMedium string `json:"avatarMedium"`
					AvatarThumb  string `json:"avatarThumb"`
					Signature    string `json:"signature"`
					CreateTime   int64  `json:"createTime"`
					Verified     bool   `json:"verified"`
					SecUID       string `json:"secUid"`
					Ftc          bool   `json:"ftc"`
					Relation     int    `json:"relation"`
					OpenFavorite bool   `json:"openFavorite"`
					BioLink      struct {
						Link string `json:"link"`
						Risk int    `json:"risk"`
					} `json:"bioLink"`
					CommentSetting int    `json:"commentSetting"`
					DuetSetting    int    `json:"duetSetting"`
					StitchSetting  int    `json:"stitchSetting"`
					PrivateAccount bool   `json:"privateAccount"`
					Secret         bool   `json:"secret"`
					RoomID         string `json:"roomId"`
				} `json:"user"`
				Stats struct {
					FollowerCount  int `json:"followerCount"`
					FollowingCount int `json:"followingCount"`
					Heart          int `json:"heart"`
					HeartCount     int `json:"heartCount"`
					VideoCount     int `json:"videoCount"`
					DiggCount      int `json:"diggCount"`
				} `json:"stats"`
				ItemList []interface{} `json:"itemList"`
			} `json:"userInfo"`
			FeedConfig struct {
				PageType   int    `json:"pageType"`
				SecUID     string `json:"secUid"`
				ID         string `json:"id"`
				ShowAvatar bool   `json:"showAvatar"`
				EmptyTip   string `json:"emptyTip"`
			} `json:"feedConfig"`
			IsSSR       bool `json:"isSSR"`
			PageOptions struct {
				Footer struct {
					Hidden       bool `json:"hidden"`
					ShowDownload bool `json:"showDownload"`
				} `json:"footer"`
				Header struct {
					ShowUpload bool   `json:"showUpload"`
					Type       string `json:"type"`
				} `json:"header"`
			} `json:"pageOptions"`
			Items               []TiktokItem `json:"items"`
			VideoListHasMore    bool         `json:"videoListHasMore"`
			VideoListMaxCursor  int64        `json:"videoListMaxCursor"`
			VideoListMinCursor  int64        `json:"videoListMinCursor"`
			VideoListStatusCode int          `json:"videoListStatusCode"`
			VideoListMode       string       `json:"videoListMode"`
		} `json:"pageProps"`
	} `json:"props"`
	BuildID      string `json:"buildId"`
	CustomServer bool   `json:"customServer"`
	Gip          bool   `json:"gip"`
	AppGip       bool   `json:"appGip"`
}

type PropDataV2 struct {
	IsMobile                   bool        `json:"$isMobile"`
	IsIOS                      interface{} `json:"$isIOS"`
	IsAndroid                  bool        `json:"$isAndroid"`
	PageURL                    string      `json:"$pageUrl"`
	Region                     string      `json:"$region"`
	IsIMessage                 bool        `json:"$isIMessage"`
	Language                   string      `json:"$language"`
	OriginalLanguage           string      `json:"$originalLanguage"`
	Os                         string      `json:"$os"`
	ReflowType                 string      `json:"$reflowType"`
	DeviceLimitRegisterExpired bool        `json:"$deviceLimitRegisterExpired"`
	AppID                      int         `json:"$appId"`
	BotType                    string      `json:"$botType"`
	AppType                    string      `json:"$appType"`
	SgOpen                     bool        `json:"$sgOpen"`
	BaseURL                    string      `json:"$baseURL"`
	PageState                  struct {
		RegionAppID int    `json:"regionAppId"`
		Os          string `json:"os"`
		Region      string `json:"region"`
		BaseURL     string `json:"baseURL"`
		AppType     string `json:"appType"`
		FullURL     string `json:"fullUrl"`
	} `json:"pageState"`
	VideoData struct {
		ItemInfos struct {
			ID    string `json:"id"`
			Video struct {
				Urls      []string `json:"urls"`
				VideoMeta struct {
					Width    int `json:"width"`
					Height   int `json:"height"`
					Ratio    int `json:"ratio"`
					Duration int `json:"duration"`
				} `json:"videoMeta"`
			} `json:"video"`
			Covers         []string      `json:"covers"`
			AuthorID       string        `json:"authorId"`
			CoversOrigin   []string      `json:"coversOrigin"`
			ShareCover     []string      `json:"shareCover"`
			Text           string        `json:"text"`
			CommentCount   int           `json:"commentCount"`
			DiggCount      int           `json:"diggCount"`
			PlayCount      int           `json:"playCount"`
			ShareCount     int           `json:"shareCount"`
			CreateTime     string        `json:"createTime"`
			IsActivityItem bool          `json:"isActivityItem"`
			WarnInfo       []interface{} `json:"warnInfo"`
			Liked          bool          `json:"liked"`
			CommentStatus  int           `json:"commentStatus"`
			ShowNotPass    bool          `json:"showNotPass"`
			Secret         bool          `json:"secret"`
			ForFriend      bool          `json:"forFriend"`
			Vl1            bool          `json:"vl1"`
			StitchEnabled  bool          `json:"stitchEnabled"`
			ShareEnabled   bool          `json:"shareEnabled"`
			IsAd           bool          `json:"isAd"`
		} `json:"itemInfos"`
		AuthorInfos struct {
			Verified bool     `json:"verified"`
			SecUID   string   `json:"secUid"`
			UniqueID string   `json:"uniqueId"`
			UserID   string   `json:"userId"`
			NickName string   `json:"nickName"`
			Covers   []string `json:"covers"`
			Relation int      `json:"relation"`
			Secret   bool     `json:"secret"`
		} `json:"authorInfos"`
		MusicInfos struct {
			MusicID    string   `json:"musicId"`
			MusicName  string   `json:"musicName"`
			AuthorName string   `json:"authorName"`
			Covers     []string `json:"covers"`
		} `json:"musicInfos"`
		AuthorStats struct {
			FollowerCount  int `json:"followerCount"`
			FollowingCount int `json:"followingCount"`
			Heart          int `json:"heart"`
			HeartCount     int `json:"heartCount"`
			VideoCount     int `json:"videoCount"`
			DiggCount      int `json:"diggCount"`
		} `json:"authorStats"`
		DuetInfo        string        `json:"duetInfo"`
		StickerTextList []interface{} `json:"stickerTextList"`
	} `json:"videoData"`
	ShareUser struct {
		SecUID       string   `json:"secUid"`
		UserID       string   `json:"userId"`
		UniqueID     string   `json:"uniqueId"`
		NickName     string   `json:"nickName"`
		Signature    string   `json:"signature"`
		Verified     bool     `json:"verified"`
		Covers       []string `json:"covers"`
		CoversMedium []string `json:"coversMedium"`
		CoversLarger []string `json:"coversLarger"`
		IsSecret     bool     `json:"isSecret"`
		Secret       bool     `json:"secret"`
		Relation     int      `json:"relation"`
	} `json:"shareUser"`
	ShareMeta struct {
		Title string `json:"title"`
		Desc  string `json:"desc"`
		Image struct {
			URL    string `json:"url"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
		} `json:"image"`
	} `json:"shareMeta"`
	StatusCode int    `json:"statusCode"`
	WebID      string `json:"webId"`
	RequestID  string `json:"requestId"`
}

func parsePropData(data []byte) (string, *pbItem.Tiktok_Item, error) {
	interData := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &interData); err != nil {
		return "", nil, err
	}

	item := &pbItem.Tiktok_Item{
		Source: &pbItem.Tiktok_Source{},
		Author: &pbItem.Tiktok_Author{
			Stats: &pbItem.Tiktok_Author_Stats{},
		},
		Video: &media.Media_Video{Cover: &media.Media_Image{}},
		Stats: &pbItem.Tiktok_Stats{},
	}
	for key, val := range interData {
		if val == nil {
			continue
		}
		if key == "props" {
			var prop PropDataV1
			propsBytes, _ := val.MarshalJSON()
			if err := json.Unmarshal(propsBytes, &prop.Props); err != nil {
				return "", nil, err
			}
			video := prop.Props.PageProps.ItemInfo.ItemStruct.Video

			item.Source.Id = prop.Props.PageProps.ItemInfo.ItemStruct.ID
			item.Title = prop.Props.PageProps.ItemInfo.ItemStruct.Desc
			if video.PlayAddr == "" && video.DownloadAddr == "" {
				return "", nil, fmt.Errorf("video url not found")
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
			author := prop.Props.PageProps.ItemInfo.ItemStruct.Author
			item.Author.Id = author.ID
			item.Author.Name = author.UniqueID
			item.Author.Nickname = author.Nickname
			item.Author.Avatar = author.AvatarLarger
			item.Author.Description = author.Signature
			authorStats := prop.Props.PageProps.ItemInfo.ItemStruct.AuthorStats
			item.Author.Stats.FollowerCount = int32(authorStats.FollowerCount)
			item.Author.Stats.FollowingCount = int32(authorStats.FollowingCount)
			item.Author.Stats.LikeCount = int32(authorStats.HeartCount)
			item.Author.Stats.VideoCount = int32(authorStats.VideoCount)
			item.Author.Stats.DiggCount = int32(authorStats.DiggCount)

			return prop.Props.InitialProps.FullURL, item, nil
		} else if strings.HasPrefix(key, "/v/") {
			var prop PropDataV2
			propDataBytes, _ := val.MarshalJSON()
			if err := json.Unmarshal(propDataBytes, &prop); err != nil {
				return "", nil, err
			}

			item.Source.Id = prop.VideoData.ItemInfos.ID
			item.Title = prop.VideoData.ItemInfos.Text
			for _, u := range prop.VideoData.ItemInfos.Video.Urls {
				if u != "" {
					item.Video.OriginalUrl = u
					break
				}
			}
			if item.Video.OriginalUrl == "" {
				return "", nil, errors.New("video url not found")
			}
			item.Video.Width = int32(prop.VideoData.ItemInfos.Video.VideoMeta.Width)
			item.Video.Height = int32(prop.VideoData.ItemInfos.Video.VideoMeta.Height)
			item.Video.Duration = int32(prop.VideoData.ItemInfos.Video.VideoMeta.Duration)
			for _, u := range prop.VideoData.ItemInfos.Covers {
				if u != "" {
					item.Video.Cover.OriginalUrl = u
					break
				}
			}

			author := prop.VideoData.AuthorInfos
			authorStats := prop.VideoData.AuthorStats
			item.Author.Id = author.UserID
			item.Author.Name = author.UniqueID
			item.Author.Nickname = author.NickName
			for _, u := range author.Covers {
				if u != "" {
					item.Author.Avatar = u
					break
				}
			}
			item.Author.Stats.FollowerCount = int32(authorStats.FollowerCount)
			item.Author.Stats.FollowingCount = int32(authorStats.FollowingCount)
			item.Author.Stats.LikeCount = int32(authorStats.HeartCount)
			item.Author.Stats.VideoCount = int32(authorStats.VideoCount)
			item.Author.Stats.DiggCount = int32(authorStats.DiggCount)
			return prop.PageState.FullURL, item, nil
		}
	}
	return "", nil, errors.New("not found")
}
