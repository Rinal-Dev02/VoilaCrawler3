package checker

import (
	"os"
	"testing"

	"github.com/voiladev/VoilaCrawler/pkg/context"
	"github.com/voiladev/VoilaCrawler/pkg/net/http/cookiejar"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	"github.com/voiladev/VoilaCrawler/pkg/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/randutil"
)

func Test_CheckImage(t *testing.T) {
	//imgUrl := `https://static.zara.net/photos/2021/I/0/1/p/8218/641/811/2/w/1280/8218641811_1_1_1.jpg?ts=1625671435228`
	imgUrl := `https://www.tods.com/fashion/tods/XXW00G000105J1M025/XXW00G000105J1M025-06.jpg?imwidth=865`
	pighubClient, err := proxy.NewPbProxyManagerClient(context.Background(), os.Getenv("VOILA_PROXY_URL"))
	if err != nil {
		t.Error(err)
	}
	client, _ := proxy.NewProxyClient(pighubClient, cookiejar.New(), glog.New(glog.LogLevelDebug))
	ctx := context.WithValue(context.Background(), context.TracingIdKey, randutil.MustNewRandomID())
	ctx = context.WithValue(ctx, context.JobIdKey, randutil.MustNewRandomID())
	err = checkImage(ctx, glog.New(glog.LogLevelDebug), imgUrl, imgSizeLarge, client, &item.Product{Source: &item.Source{CrawlUrl: "https://www.tods.com/us-en/Leather-Timeless-Belt-Bag-Micro/p/XBWTSIRC0L0RORB999"}})
	if err != nil {
		t.Error(err)
	}
}
