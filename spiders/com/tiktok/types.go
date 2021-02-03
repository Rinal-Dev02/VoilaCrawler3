package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/media"
	pbItem "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/item"
)

type PropDataV1 struct {
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
			FollowerCount int    `json:"followerCount"`
			HeartCount    string `json:"heartCount"`
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
		Source:   &pbItem.Tiktok_Source{},
		Video:    &media.Media_Video{Cover: &media.Media_Image{}},
		AuthInfo: &pbItem.Tiktok_Item_AuthInfo{},
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
			return prop.PageState.FullURL, item, nil
		}
	}
	return "", nil, errors.New("not found")
}
