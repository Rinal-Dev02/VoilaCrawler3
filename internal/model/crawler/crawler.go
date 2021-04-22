package crawler

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/voiladev/VoilaCrawl/pkg/types"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	"google.golang.org/grpc"
)

type Crawler struct {
	types.Crawler
	LastHeartbeatUtc int64

	isConnected bool
	conn        *grpc.ClientConn
	client      pbCrawl.CrawlerNodeClient
	mutex       sync.Mutex
}

func NewCrawler(i *pbCrawl.ConnectRequest_Ping, ip string) (*Crawler, error) {
	if ip == "" {
		return nil, errors.New("invalid crawler server ip")
	}
	if i.ServePort <= 0 {
		return nil, errors.New("invalid crawler server port")
	}
	if len(i.AllowedDomains) == 0 {
		return nil, fmt.Errorf("invalid allowed domains for crawler %s", i.Id)
	}

	crawler := Crawler{}
	crawler.RawId = i.GetId()
	crawler.Id = fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s-%d-%s-%d",
		i.GetId(), crawler.Version, crawler.ServeAddr, crawler.OnlineUtc))))
	crawler.StoreId = i.GetStoreId()
	crawler.Version = i.GetVersion()
	crawler.AllowedDomains = i.GetAllowedDomains()
	crawler.ServeAddr = fmt.Sprintf("%s:%d", ip, i.ServePort)
	crawler.OnlineUtc = time.Now().Unix()
	crawler.LastHeartbeatUtc = crawler.OnlineUtc

	return &crawler, nil
}

func (e *Crawler) IsConnected() bool {
	if e == nil {
		return false
	}
	e.mutex.Lock()
	defer e.mutex.Unlock()

	return e.isConnected
}

func (e *Crawler) Unmarshal(ret interface{}) error {
	if e == nil {
		return errors.New("empty crawler")
	}

	switch v := ret.(type) {
	case *pbCrawl.Crawler:
		v.Id = e.Id
		v.StoreId = e.StoreId
		v.Version = e.Version
		v.AllowedDomains = e.AllowedDomains
		v.Metadata = &pbCrawl.Crawler_Metadata{
			OnlineUtc: e.OnlineUtc,
		}
	default:
		return fmt.Errorf("unsupported unmarshal type")
	}
	return nil
}

func (e *Crawler) Connect(ctx context.Context) (pbCrawl.CrawlerNodeClient, error) {
	if e == nil {
		return nil, errors.New("nil crawler")
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.isConnected {
		return e.client, nil
	}

	var err error
	e.conn, err = grpc.DialContext(ctx, e.GetServeAddr(),
		grpc.WithBlock(), grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(100*1024*1024)), // 100Mi
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(100*1024*1024)), // 100Mi
	)
	if err != nil {
		return nil, err
	}
	e.client = pbCrawl.NewCrawlerNodeClient(e.conn)
	e.isConnected = true

	return e.client, nil
}

func (e *Crawler) Close() error {
	if e == nil || e.conn == nil {
		return nil
	}
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if !e.isConnected {
		return nil
	}

	e.isConnected = false

	err := e.conn.Close()
	e.conn = nil
	e.client = nil

	return err
}
