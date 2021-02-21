package proxy

import (
	"context"
	"os"
	"testing"

	"github.com/voiladev/VoilaCrawl/pkg/net/http"
	"github.com/voiladev/VoilaCrawl/pkg/net/http/cookiejar"
	"github.com/voiladev/go-framework/glog"
)

func Test_proxyClient_DoWithOptions(t *testing.T) {
	var (
		apiToken = os.Getenv("PC_API_TOKEN")
		jsToken  = os.Getenv("PC_JS_TOKEN")
	)
	if apiToken == "" || jsToken == "" {
		panic("env PC_API_TOKEN or PC_JS_TOKEN is not set")
	}

	logger := glog.New(glog.LogLevelDebug)
	client, err := NewProxyClient("http://127.0.0.1:6152", cookiejar.New(), logger)
	if err != nil {
		t.Fatal(err)
	}

	u := "https://www.tiktok.com/@kasey.jo.gerst/video/6923743895247506693?sender_device=mobile&sender_web_id=6926525695457117698&is_from_webapp=v2&is_copy_url=0"
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	// req.Header.Set("Accept-Encoding", "identity;q=1, *;q=0")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,pl;q=0.7,zh-TW;q=0.6,ca;q=0.5,mt;q=0.4")
	// req.Header.Set("Cache-Control", "no-cache")
	// req.Header.Set("Connection", "keep-alive")
	// req.Header.Set("Cookie", cookie)
	// req.Header.Set("Pragma", "no-cache")
	// req.Header.Set("Referer", "https://www.tiktok.com/")
	// req.Header.Set("sec-ch-ua", `"Chromium";v="88", "Google Chrome";v="88", ";Not A Brand";v="99"`)
	// req.Header.Set("sec-ch-ua-mobile", "?0")
	// req.Header.Set("Sec-Fetch-Dest", "video")
	// req.Header.Set("Sec-Fetch-Mode", "no-cors")
	// req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.96 Safari/537.36")
	// req.Header.Set("Range", "bytes=0-")
	resp, err := client.DoWithOptions(context.Background(), req, http.Options{EnableProxy: false})
	if err != nil {
		t.Fatalf("proxyClient.DoWithOptions() error = %v", err)
	}
	defer resp.Body.Close()

	// data, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	for _, c := range resp.Cookies() {
		t.Logf("cookie %s", c.Name)
	}
}
