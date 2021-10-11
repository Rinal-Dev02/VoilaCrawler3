package util

import "net/url"

func UrlCompletion(rawurl string, parenturl ...string) string {
	if rawurl == "" {
		return rawurl
	}

	if u, err := url.QueryUnescape(rawurl); err == nil {
		rawurl = u
	}

	u, err := url.Parse(rawurl)
	if err != nil {
		return ""
	}
	var pu *url.URL
	if len(parenturl) > 0 {
		pu, err = url.Parse(parenturl[0])
	}

	if u.Scheme == "" {
		if pu != nil && pu.Scheme != "" {
			u.Scheme = pu.Scheme
		} else {
			u.Scheme = "https"
		}
	}
	if u.Host == "" {
		if pu != nil && pu.Host != "" {
			u.Host = pu.Host
		} else {
			return ""
		}
	}
	return u.String()
}
