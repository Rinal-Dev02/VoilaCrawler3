package cookiejar

import (
	"context"
	"errors"
	// rhttp "net/http"
	"net/url"
	"time"

	ctxutil "github.com/voiladev/go-crawler/pkg/context"
	"github.com/voiladev/go-crawler/pkg/net/http"
	pbHttp "github.com/voiladev/go-crawler/protoc-gen-go/chameleon/api/http"
	pbSession "github.com/voiladev/go-crawler/protoc-gen-go/chameleon/smelter/v1/crawl/session"
	"github.com/voiladev/go-framework/glog"
)

type RemoteJar struct {
	sessionClient pbSession.SessionManagerClient
	logger        glog.Log
}

func NewRemoteJar(sessionClient pbSession.SessionManagerClient, logger glog.Log) (http.CookieJar, error) {
	if sessionClient == nil {
		return nil, errors.New("invalid session client")
	}
	if logger == nil {
		return nil, errors.New("invalid logger")
	}

	j := RemoteJar{
		sessionClient: sessionClient,
		logger:        logger,
	}
	return &j, nil
}

func (j *RemoteJar) Cookies(ctx context.Context, u *url.URL) (cookies []*http.Cookie, err error) {
	if j == nil || u == nil {
		return
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return
	}
	logger := j.logger.New("Cookies")

	tracingId := ctxutil.GetString(ctx, ctxutil.TracingIdKey)
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

func (j *RemoteJar) SetCookies(ctx context.Context, u *url.URL, cookies []*http.Cookie) (err error) {
	if j == nil || u == nil {
		return
	}
	logger := j.logger.New("SetCookies")

	req := pbSession.SetCookiesRequest{Url: u.String()}
	req.TracingId = ctxutil.GetString(ctx, ctxutil.TracingIdKey)
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

func (j *RemoteJar) Clear(ctx context.Context, u *url.URL) (err error) {
	if j == nil || u == nil {
		return
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return
	}
	logger := j.logger.New("Clear")

	tracingId := ctxutil.GetString(ctx, ctxutil.TracingIdKey)
	if _, err := j.sessionClient.ClearCookies(ctx, &pbSession.ClearCookiesRequest{
		TracingId: tracingId,
		Url:       u.String(),
	}); err != nil {
		logger.Errorf("get cookies failed, error=%s", err)
		return err
	}
	return
}
