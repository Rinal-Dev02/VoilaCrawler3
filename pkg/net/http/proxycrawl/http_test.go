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

	req, err := http.NewRequest(http.MethodGet, "https://v16-web.tiktok.com/video/tos/alisg/tos-alisg-pve-0037c001/893ed9debcad47a0972b9764ce8b958f/?a=1988&br=3108&bt=1554&cd=0%7C0%7C1&ch=0&cr=0&cs=0&cv=1&dr=0&ds=3&er=&expire=1612298013&l=202102021433240101891950151210FA2B&lr=tiktok_m&mime_type=video_mp4&pl=0&policy=2&qs=0&rc=anRqZG5veGdpeTMzPDczM0ApZWhkNmVmNjtmNzw0PDo6O2dxMDUyZC1ybTFfLS00MTRzcy1iMl8yYjVjLzAvLzRfYDM6Yw%3D%3D&signature=7eeb8b1507570475f1d7dda2ff06afd7&tk=tt_webid_v2&vl=&vr=", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.96 Safari/537.36")
	req.Header.Set("Cookie", "tt_csrf_token=yyeiJdZcDQInuQjPALW4ZFYq; bm_sz=1CE0645BF55474CB79CBC2435EA709BB~YAAQjQcVAql8FQF3AQAApKIpYwrGO+ZTE6QNPxLyAJPEFCbz+ehu7I65dv7Z3v9qysYBtTEOBf3JChUTypumyCfG15pdMIRIpbuWkwTwh69h2CsdI8dJ5/sgEkmM8qv8Ce1PJYEKDUkZHEEuPgHkJXMu4H8UORSFpj2fvYGcS9tUuumpdwFCtegdnsT34Xgh; _abck=3D67976C89560DE5EEBE1F9FE5F05899~-1~YAAQjQcVAqp8FQF3AQAApKIpYwWWKzguwZXKV4AkgL+CDd7XBHQFRWthgDO2RVVUwvmBS0AbR+CIBRoHatZYbnXhlnwVeMsikv5qBvCX6og7ju5YcBCuC54vHjavRkEZxSa1wB6KRVTw4aGentdjiDWZUUJJA7BS8Xds2o78/bv7qB1RDXUykGegu2m+F7E151w4720wjfHNPnsKxK7iCkdx2+DVCHDVN+k4RSruei1qsUpIY/jE5tEgG0FTRrTISgRqjHFKjnP/BmmMEE4x0WVKnDLpcwpQxnoJb7P6QUBMU1jPC7GOXxl6~-1~-1~-1")
	resp, err := client.DoWithOptions(context.Background(), req, http.Options{EnableProxy: false, DisableBackconnect: false})
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
