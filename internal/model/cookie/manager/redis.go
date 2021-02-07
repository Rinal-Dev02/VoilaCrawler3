package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/voiladev/VoilaCrawl/internal/model/cookie"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/redis"
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
)

type CookieManager struct {
	redisClient *redis.RedisClient
	logger      glog.Log
}

// NewCookieManager
func NewCookieManager(redisClient *redis.RedisClient, logger glog.Log) (*CookieManager, error) {
	if redisClient == nil {
		return nil, fmt.Errorf("invalid redis client")
	}
	if logger == nil {
		return nil, fmt.Errorf("invalid logger")
	}

	m := CookieManager{
		redisClient: redisClient,
		logger:      logger,
	}
	return &m, nil
}

func uniqueDomainKey(domain string, tracingId string) string {
	domainFields := strings.FieldsFunc(domain, func(r rune) bool {
		return r == '.'
	})
	for i, j := 0, len(domainFields)-1; i < j; i, j = i+1, j-1 {
		domainFields[i], domainFields[j] = domainFields[j], domainFields[i]
	}

	var rdomain string = domainFields[0]
	if len(domainFields) >= 2 {
		rdomain = strings.Join(domainFields[0:2], ".")
	}
	return fmt.Sprintf("cookies://domain/%s/trace/%s", rdomain, tracingId)
}

func (m *CookieManager) List(ctx context.Context, u *url.URL, tracingId string) ([]*cookie.Cookie, error) {
	if m == nil || u == nil {
		return nil, nil
	}
	logger := m.logger.New("List")

	key := uniqueDomainKey(u.Host, tracingId)
	vals, err := redis.ByteSlices(m.redisClient.Do("HVALS", key))
	if err != nil {
		logger.Errorf("get cookies values of %s failed, error=%s", key, err)
		return nil, pbError.ErrDatabase.New(err)
	}

	var (
		cookies []*cookie.Cookie
		t       = time.Now().Unix()
	)
	for _, val := range vals {
		var c cookie.Cookie
		if json.Unmarshal(val, &c.Cookie); err != nil {
			logger.Errorf("unmarshal cookie failed, error=%s", err)
			return nil, pbError.ErrDataLoss.New(err)
		}
		if c.GetExpires() > 0 && c.GetExpires() <= t {
			continue
		}
		if u.Path != "" && !strings.HasPrefix(u.Path, c.Path) {
			continue
		}
		cookies = append(cookies, &c)
	}
	return cookies, nil
}

func (m *CookieManager) Save(ctx context.Context, c *cookie.Cookie) error {
	if m == nil {
		return nil
	}
	logger := m.logger.New("Save")

	if err := c.Validate(); err != nil {
		return pbError.ErrInvalidArgument.New(err)
	}

	cookieData, err := json.Marshal(&c.Cookie)
	if err != nil {
		logger.Errorf("marshal json failed, error=%s", err)
		return pbError.ErrInvalidArgument.New(err)
	}

	key := uniqueDomainKey(c.Domain, c.TracingId)
	if _, err := m.redisClient.Do("HSET", key, c.TracingId, cookieData); err != nil {
		logger.Errorf("save cookie failed, error=%s", err)
		return pbError.ErrDatabase.New(err)
	}
	if _, err := m.redisClient.Do("EXPIRE", key, 7*24*3600); err != nil {
		logger.Errorf("set ttl of key %s failed, error=%s", key, err)
		return pbError.ErrDatabase.New(err)
	}
	return nil
}

func (m *CookieManager) Delete(ctx context.Context, u *url.URL, tracingId string) error {
	if m == nil {
		return nil
	}
	logger := m.logger.New("Delete")

	key := uniqueDomainKey(u.Host, tracingId)
	if _, err := m.redisClient.Do("DELETE", key); err != nil {
		logger.Errorf("save cookie failed, error=%s", err)
		return pbError.ErrDatabase.New(err)
	}
	return nil
}
