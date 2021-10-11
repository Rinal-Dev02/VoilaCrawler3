package media

import (
	"strings"

	anypb "google.golang.org/protobuf/types/known/anypb"
)

func NewImageMedia(id, original, large, medium, small, desc string, isDefault bool) *Media {
	if strings.HasPrefix(original, "//") {
		original = "https:" + original
	}
	if strings.HasPrefix(large, "//") {
		large = "https:" + large
	}
	if strings.HasPrefix(medium, "//") {
		medium = "https:" + medium
	}
	if strings.HasPrefix(small, "//") {
		small = "https:" + small
	}
	data, _ := anypb.New(&Media_Image{
		Id:          id,
		OriginalUrl: original,
		LargeUrl:    large,
		MediumUrl:   medium,
		SmallUrl:    small,
	})
	return &Media{
		Detail:    data,
		IsDefault: isDefault,
		Text:      desc,
	}
}

func NewVideoMedia(id, typ, url string, width, height, duration int, coverUrl string, desc string, isDefault bool) *Media {
	if strings.HasPrefix(url, "//") {
		url = "https:" + url
	}
	if strings.HasPrefix(coverUrl, "//") {
		url = "https" + coverUrl
	}
	v := Media_Video{
		Id:          id,
		OriginalUrl: url,
		Width:       int32(width),
		Height:      int32(height),
		Duration:    int32(duration),
	}
	if coverUrl != "" {
		v.Cover = &Media_Image{OriginalUrl: coverUrl}
	}

	data, _ := anypb.New(&v)
	return &Media{
		Detail:    data,
		IsDefault: isDefault,
		Text:      desc,
	}
}
