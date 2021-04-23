package manager

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	gogoproto "github.com/gogo/protobuf/proto"
	"github.com/voiladev/VoilaCrawl/internal/model/crawler"
	"github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/redis"
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
)

var crawlerOnlineCheckerHander sync.Once

type CrawlerWrap struct {
	storeId string
	crawler *crawl.Crawler
}

type CrawlerManager struct {
	ctx         context.Context
	crawlers    sync.Map
	redisClient *redis.RedisClient
	logger      glog.Log
}

func NewCrawlerManager(ctx context.Context, redisClient *redis.RedisClient, logger glog.Log) (*CrawlerManager, error) {
	if redisClient == nil {
		return nil, errors.New("invalid redis client")
	}
	m := CrawlerManager{
		ctx:         ctx,
		redisClient: redisClient,
		logger:      logger.New("CrawlerManager"),
	}

	crawlerOnlineCheckerHander.Do(func() {
		go func() {
			const checkInterval = time.Second * 30
			timer := time.NewTimer(checkInterval)
			for {
				select {
				case <-ctx.Done():
					return
				case <-timer.C:
					m.crawlers.Range(func(key, value interface{}) bool {
						id, _ := key.(string)
						cw, _ := value.(*crawler.Crawler)
						// check if the crawler is offline
						cacheKey := crawlerDetailCacheKey(id)
						if _, err := m.redisClient.Do("GET", cacheKey); err == redis.ErrNil {
							m.crawlers.Delete(key)
							cw.Close()
						} else if err != nil {
							m.logger.Error(err)
						}
						return true
					})
					timer.Reset(checkInterval)
				}
			}
		}()
	})
	return &m, nil
}

func crawlerDetailCacheKey(id string) string {
	return fmt.Sprintf("cache://stores/-/crawlers/%s", id)
}

func crawlerStoreCacheKey(storeId string) string {
	return fmt.Sprintf("cache://stores/%s/crawlers", storeId)
}

func crawlerStoresCacheKey() string {
	return fmt.Sprintf("cache://stores")
}

// GetByID
func (m *CrawlerManager) GetByID(ctx context.Context, id string) (*crawler.Crawler, error) {
	if m == nil {
		return nil, nil
	}
	logger := m.logger.New("GetByID")

	// check if the crawler is offline
	key := crawlerDetailCacheKey(id)
	data, err := redis.Bytes(m.redisClient.Do("GET", key))
	if err == redis.ErrNil {
		// offline
		return nil, nil
	} else if err != nil {
		// redis error
		logger.Error(err)
		return nil, pbError.ErrInternal.New(err.Error())
	}

	// get from cache if not in cache, create with the info
	if val, ok := m.crawlers.Load(key); ok {
		cw, _ := val.(*crawler.Crawler)
		return cw, nil
	}
	var crawler crawler.Crawler
	if err := gogoproto.Unmarshal(data, &crawler.Crawler); err != nil {
		logger.Error(err)
		return nil, pbError.ErrInternal.New(err)
	}
	m.crawlers.Store(key, &crawler)

	return &crawler, nil
}

// GetByStore
func (m *CrawlerManager) GetByStore(ctx context.Context, storeId string) ([]*crawler.Crawler, error) {
	if m == nil {
		return nil, nil
	}
	logger := m.logger.New("GetByStore")

	key := crawlerStoreCacheKey(storeId)
	vals, err := redis.Values(m.redisClient.Do("ZREVRANGE", key, 0, -1, "WITHSCORES"))
	if err != nil {
		logger.Error(err)
		return nil, pbError.ErrInternal.New(err)
	}

	var ret []*crawler.Crawler
	for i := 0; i < len(vals); i += 2 {
		id, _ := redis.String(vals[i], nil)
		heartbeatUtc, _ := redis.Int64(vals[i+1], nil)

		cw, err := m.GetByID(ctx, id)
		if err != nil {
			logger.Error(err)
			continue
		}
		if cw == nil {
			logger.Warnf("data not consitent, crawler %s registered in store list but instance is expired", id)
			continue
		}
		cw.LastHeartbeatUtc = heartbeatUtc
		ret = append(ret, cw)
	}
	return ret, nil
}

// CountOfStore
func (m *CrawlerManager) CountOfStore(ctx context.Context, storeId string) (int, error) {
	if m == nil {
		return 0, nil
	}
	logger := m.logger.New("CountOfStore")

	key := crawlerStoreCacheKey(storeId)
	count, err := redis.Int(m.redisClient.Do("ZCARD", key))
	if err != nil {
		logger.Error(err)
		return 0, pbError.ErrInternal.New(err)
	}
	return count, nil
}

// GetStores
func (m *CrawlerManager) GetStores(ctx context.Context) ([]string, error) {
	if m == nil {
		return nil, nil
	}
	logger := m.logger.New("GetStores")

	key := crawlerStoresCacheKey()

	storeIds, err := redis.Strings(m.redisClient.Do("ZREVRANGE", key, 0, -1))
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	return storeIds, nil
}

// List
func (m *CrawlerManager) List(ctx context.Context) (map[string][]*crawler.Crawler, error) {
	if m == nil {
		return nil, nil
	}
	logger := m.logger.New("List")

	key := crawlerStoresCacheKey()

	ret := map[string][]*crawler.Crawler{}
	storeIds, err := redis.Strings(m.redisClient.Do("ZREVRANGE", key, 0, -1))
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	for _, id := range storeIds {
		cws, err := m.GetByStore(ctx, id)
		if err != nil {
			logger.Error(err)
			return nil, err
		}
		ret[id] = cws
	}
	return ret, nil
}

// Cache
func (m *CrawlerManager) Cache(ctx context.Context, crawler *crawler.Crawler, ttl int64) error {
	if m == nil {
		return nil
	}
	logger := m.logger.New("Cache")

	cacheKey := crawlerDetailCacheKey(crawler.GetId())
	storeKey := crawlerStoreCacheKey(crawler.GetStoreId())
	crawlersKey := crawlerStoresCacheKey()

	detailData, _ := gogoproto.Marshal(&crawler.Crawler)
	t := time.Now().Unix()

	// KEYS[1]=crawlersKey, [2]=storeKey, [3]=cacheKey
	// ARGV[1]=storeId, [2]=crawlerId [3]=detailData, [4]=TTL, [5]=timestamp
	script := `redis.call("SET", KEYS[3], ARGV[3], "EX", ARGV[4])
redis.call("ZADD", KEYS[2], ARGV[5], ARGV[2])
redis.call("ZADD", KEYS[1], ARGV[5], ARGV[1])`

	if _, err := m.redisClient.Do("EVAL", script, 3,
		crawlersKey, storeKey, cacheKey,
		crawler.GetStoreId(), crawler.GetId(), detailData, ttl, t); err != nil {
		logger.Error(err)
		return pbError.ErrInternal.New(err)
	}
	return nil
}

// Delete
func (m *CrawlerManager) Delete(ctx context.Context, storeId, id string) error {
	if m == nil {
		return nil
	}
	if storeId == "" || id == "" {
		return pbError.ErrInvalidArgument.New("invalid storeId or id")
	}
	logger := m.logger.New("Delete")

	crawlersKey := crawlerStoresCacheKey()
	cacheKey := crawlerDetailCacheKey(id)
	storeKey := crawlerStoreCacheKey(storeId)

	script := `redis.call("DEL", KEYS[3])
redis.call("ZREM", KEYS[2], ARGV[2])
local count=redis.call("ZCARD", KEYS[2])
if count == 0 then
    redis.call("ZREM", KEYS[1], ARGV[1])
end
return 1
`

	if _, err := m.redisClient.Do("EVAL", script, 3, crawlersKey, storeKey, cacheKey, storeId, id); err != nil {
		logger.Error(err)
		return pbError.ErrInternal.New(err)
	}
	return nil
}

// Clean
func (m *CrawlerManager) Clean(ctx context.Context, storeId string) error {
	if m == nil {
		return nil
	}
	if storeId == "" {
		return pbError.ErrInvalidArgument.New("invalid storeId")
	}
	logger := m.logger.New("Clean")

	storeKey := crawlerStoreCacheKey(storeId)

	script := `local ids = redis.call("ZRANGE", KEYS[1], 0, -1)
for _,id in ipairs(ids) do
	if redis.call("EXISTS", "cache://stores/-/crawlers/"..id) ~= 1 then
	    redis.call("ZREM", KEYS[1], id)
	end
end
return 1`

	if _, err := m.redisClient.Do("EVAL", script, 1, storeKey); err != nil {
		logger.Error(err)
		return pbError.ErrInternal.New(err)
	}
	return nil
}
