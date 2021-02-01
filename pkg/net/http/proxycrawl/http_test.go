package proxycrawl

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/voiladev/VoilaCrawl/pkg/net/http"
	"github.com/voiladev/go-framework/glog"
)

func Test_proxyCrawlClient_DoWithOptions(t *testing.T) {
	var (
		apiToken = os.Getenv("PC_API_TOKEN")
		jsToken  = os.Getenv("PC_JS_TOKEN")
	)
	if apiToken == "" || jsToken == "" {
		panic("env PC_API_TOKEN or PC_JS_TOKEN is not set")
	}

	logger := glog.New(glog.LogLevelDebug)
	client, err := NewProxyCrawlClient(logger, Options{APIToken: apiToken, JSToken: jsToken})
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, "https://www.asos.com/api/product/catalogue/v3/stockprice?productIds=23385813&store=US&currency=USD&keyStoreDataversion=3pmn72e-27", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.96 Safari/537.36")
	req.Header.Set("asos-c-name", "asos-web-productpage")
	req.Header.Set("asos-c-version", "1.0.0-52db3f927a77-2353")
	resp, err := client.DoWithOptions(context.Background(), req, http.Options{EnableProxy: true})
	if err != nil {
		t.Fatalf("proxyCrawlClient.DoWithOptions() error = %v", err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("data: %s", data)
}
