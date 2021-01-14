package url

import (
	"net/url"
	"strings"
)

func Format(rawurl string) string {
	if rawurl == "" {
		return rawurl
	}

	if u, err := url.QueryUnescape(rawurl); err == nil {
		rawurl = u
	}
	if strings.HasPrefix(rawurl, "//") {
		rawurl = "https:" + rawurl
	}

	return rawurl
}
