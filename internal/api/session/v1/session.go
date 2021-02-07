package v1

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/voiladev/VoilaCrawl/internal/model/cookie"
	cookieManager "github.com/voiladev/VoilaCrawl/internal/model/cookie/manager"
	pbHttp "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/http"
	pbSession "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/session"
	"github.com/voiladev/go-framework/glog"
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
	"google.golang.org/protobuf/types/known/emptypb"
)

type SessionServer struct {
	pbSession.UnimplementedSessionManagerServer

	ctx           context.Context
	cookieManager *cookieManager.CookieManager
	logger        glog.Log
}

func NewSessionServer(
	ctx context.Context,
	cookieManager *cookieManager.CookieManager,
	logger glog.Log,
) (pbSession.SessionManagerServer, error) {
	if cookieManager == nil {
		return nil, errors.New("invalid cookie manager")
	}
	if logger == nil {
		return nil, errors.New("invalid logger")
	}
	s := SessionServer{
		ctx:           ctx,
		cookieManager: cookieManager,
		logger:        logger.New("SessionServer"),
	}
	return &s, nil
}

func (s *SessionServer) GetCookies(ctx context.Context, req *pbSession.GetCookiesRequest) (*pbSession.GetCookiesResponse, error) {
	if s == nil {
		return nil, nil
	}
	logger := s.logger.New("GetCookies")

	u, err := url.Parse(req.GetUrl())
	if err != nil {
		logger.Errorf("parse url %s failed, error=%s", req.GetUrl(), err)
		return nil, pbError.ErrInvalidArgument.New("invalid url")
	}
	cookies, err := s.cookieManager.List(ctx, u, req.GetTracingId())
	if err != nil {
		logger.Errorf("get cookies failed, error=%s", err)
		return nil, err
	}
	var resp pbSession.GetCookiesResponse
	for _, cookie := range cookies {
		var c pbHttp.Cookie
		if err := cookie.Unmarshal(&c); err != nil {
			logger.Errorf("unmarshal cookie failed, error=%s", err)
			return nil, pbError.ErrInternal.New(err)
		}
		resp.Data = append(resp.Data, &c)
	}
	return &resp, nil
}

func (s *SessionServer) SetCookies(ctx context.Context, req *pbSession.SetCookiesRequest) (*emptypb.Empty, error) {
	if s == nil {
		return nil, nil
	}
	logger := s.logger.New("SetCookies")

	u, err := url.Parse(req.GetUrl())
	if err != nil {
		logger.Errorf("parse url %s failed, error=%s", req.GetUrl(), err)
		return nil, pbError.ErrInvalidArgument.New(err)
	}
	defaultDomain := u.Hostname()

	for _, rawCookie := range req.GetCookies() {
		if rawCookie.GetDomain() != "" {
			// TODO: optimize
			if !strings.HasSuffix(defaultDomain, rawCookie.Domain) {
				logger.Debugf("host %s not match domain %s", defaultDomain, rawCookie.Domain)
				continue
			}
		} else {
			rawCookie.Domain = defaultDomain
		}

		if c, err := cookie.New(req.GetTracingId(), rawCookie); err != nil {
			logger.Errorf("load cookie failed, error=%s", err)
			return nil, pbError.ErrInvalidArgument.New(err)
		} else if err := s.cookieManager.Save(ctx, c); err != nil {
			logger.Errorf("save cookie %s %s failed, error=%s", c.GetTracingId(), c.GetName())
			return nil, pbError.ErrInvalidArgument.New(err)
		}
	}
	return &emptypb.Empty{}, nil
}

func (s *SessionServer) ClearCookies(ctx context.Context, req *pbSession.ClearCookiesRequest) (*emptypb.Empty, error) {
	if s == nil {
		return nil, nil
	}
	logger := s.logger.New("ClearCookies")

	u, err := url.Parse(req.GetUrl())
	if err != nil {
		logger.Errorf("parse url %s failed, error=%s", req.GetUrl(), err)
		return nil, pbError.ErrInvalidArgument.New(err)
	}

	if err := s.cookieManager.Delete(ctx, u, req.GetTracingId()); err != nil {
		logger.Errorf("save cookie %s %s failed, error=%s", req.GetTracingId(), req.GetUrl(), err)
		return nil, pbError.ErrInvalidArgument.New(err)
	}
	return &emptypb.Empty{}, nil
}
