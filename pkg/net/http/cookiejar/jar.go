package cookiejar

import (
	"context"
	rhttp "net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/voiladev/go-crawler/pkg/net/http"
	"golang.org/x/net/publicsuffix"
)

type Jar struct {
	jar rhttp.CookieJar
}

func New() http.CookieJar {
	rawJar, _ := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	return &Jar{jar: rawJar}
}

func (j *Jar) Clear(ctx context.Context, u *url.URL) (err error) {
	cs := j.jar.Cookies(u)
	if len(cs) > 0 {
		for _, c := range cs {
			c.MaxAge = -1
		}
		j.jar.SetCookies(u, cs)
	}

	return nil
}

func (j *Jar) Cookies(ctx context.Context, u *url.URL) (cookies []*http.Cookie, err error) {
	if j == nil || u == nil {
		return
	}
	return j.jar.Cookies(u), nil
}

func (j *Jar) SetCookies(ctx context.Context, u *url.URL, cookies []*http.Cookie) (err error) {
	if j == nil || u == nil {
		return
	}

	j.jar.SetCookies(u, cookies)

	return
}
