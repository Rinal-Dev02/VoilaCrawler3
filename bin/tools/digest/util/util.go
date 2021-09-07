package util

import (
	"net/url"
	"regexp"
)

var (
	nikeImageReg = regexp.MustCompile(`/t_PDP_[a-zA-Z0-9]+_v1/f_[^/]+/`)
)

func FormatImageUrl(rawurl string) (string, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	u.Scheme = "https"

	vals := u.Query()
	switch u.Hostname() {
	case "www.lulus.com":
		vals.Set("w", "560")
	case "static.nike.com":
		// "/t_PDP_864_v1/f_auto,b_rgb:f5f5f5,w_560/"
		u.Path = nikeImageReg.ReplaceAllString(u.Path, "/t_PDP_864_v1/f_auto,w_560/")
	}
	u.RawQuery = vals.Encode()
	return u.String(), nil
}
