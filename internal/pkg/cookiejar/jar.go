package cookiejar

import (
	"context"
	"errors"
	rhttp "net/http"
	"net/url"
	"time"

	"github.com/voiladev/VoilaCrawl/pkg/net/http"
	pbHttp "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/http"
	pbSession "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/session"
	"github.com/voiladev/go-framework/glog"
)

type standardJar struct {
	jar *Jar
}

func (j *standardJar) Cookies(u *url.URL) []*http.Cookie {
	if j == nil || j.jar == nil || u == nil {
		return nil
	}

	cs, _ := j.jar.Cookies(context.Background(), u)
	return cs
}

func (j *standardJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	if j == nil || j.jar == nil || u == nil || cookies == nil {
		return
	}
	j.jar.SetCookies(context.Background(), u, cookies)
}

type Jar struct {
	sessionClient pbSession.SessionManagerClient
	logger        glog.Log
}

func New(sessionClient pbSession.SessionManagerClient, logger glog.Log) (http.CookieJar, error) {
	if sessionClient == nil {
		return nil, errors.New("invalid session client")
	}
	if logger == nil {
		return nil, errors.New("invalid logger")
	}

	j := Jar{
		sessionClient: sessionClient,
		logger:        logger,
	}
	return &j, nil
}

func (j *Jar) Jar() rhttp.CookieJar {
	if j == nil {
		return nil
	}

	return nil
}

func (j *Jar) Cookies(ctx context.Context, u *url.URL) (cookies []*http.Cookie, err error) {
	if j == nil || u == nil {
		return
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return
	}
	logger := j.logger.New("Cookies")

	var tracingId string
	if v := ctx.Value("tracing_id"); v != nil {
		tracingId, _ = v.(string)
	}

	resp, err := j.sessionClient.GetCookies(ctx, &pbSession.GetCookiesRequest{
		TracingId: tracingId,
		Url:       u.String(),
	})
	if err != nil {
		logger.Errorf("get cookies failed, error=%s", err)
		return nil, err
	}

	for _, c := range resp.GetData() {
		cookie := http.Cookie{
			Name:     c.GetName(),
			Value:    c.GetValue(),
			Path:     c.GetPath(),
			Domain:   c.GetDomain(),
			HttpOnly: c.GetHttpOnly(),
			SameSite: http.SameSite(c.GetSameSite()),
		}
		if c.GetExpires() > 0 {
			t := time.Unix(c.GetExpires(), 0)
			if t.After(time.Now()) {
				cookie.Expires = t
			}
		}
	}
	return
}

func (j *Jar) SetCookies(ctx context.Context, u *url.URL, cookies []*http.Cookie) (err error) {
	if j == nil || u == nil {
		return
	}
	logger := j.logger.New("SetCookies")

	req := pbSession.SetCookiesRequest{Url: u.String()}
	if v := ctx.Value("tracing_id"); v != nil {
		req.TracingId, _ = v.(string)
	}
	for _, c := range cookies {
		cookie := pbHttp.Cookie{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			HttpOnly: c.HttpOnly,
			SameSite: int32(c.SameSite),
		}
		if c.MaxAge > 0 {
			cookie.Expires = int64(c.MaxAge) + time.Now().Unix()
		} else if !c.Expires.IsZero() {
			cookie.Expires = c.Expires.Unix()
		}
		req.Cookies = append(req.Cookies, &cookie)
	}

	if _, err := j.sessionClient.SetCookies(ctx, &req); err != nil {
		logger.Errorf("set cookies failed, error=%s", err)
		return err
	}
	return
}
