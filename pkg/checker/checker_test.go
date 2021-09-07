package checker

import (
	"github.com/voiladev/VoilaCrawler/pkg/context"
	"github.com/voiladev/VoilaCrawler/pkg/net/http/cookiejar"
	"github.com/voiladev/VoilaCrawler/pkg/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/randutil"
	"os"
	"testing"
)

func Test_CheckImage(t *testing.T) {
	imgUrl := `https://static.zara.net/photos/2021/I/0/1/p/8218/641/811/2/w/1280/8218641811_1_1_1.jpg?ts=1625671435228`
	client, _ := proxy.NewProxyClient(os.Getenv("VOILA_PROXY_URL"), cookiejar.New(), glog.New(glog.LogLevelDebug))
	ctx := context.WithValue(context.Background(), context.TracingIdKey, randutil.MustNewRandomID())
	ctx = context.WithValue(ctx, context.JobIdKey, randutil.MustNewRandomID())
	err := checkImage(ctx, imgUrl, imgSizeLarge, client)
	if err != nil {
		t.Error(err)
	}
}
