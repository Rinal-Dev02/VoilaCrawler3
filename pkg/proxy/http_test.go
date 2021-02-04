package proxy

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/voiladev/VoilaCrawl/pkg/net/http"
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
	client, err := NewProxyClient(nil, logger, Options{APIToken: apiToken, JSToken: jsToken})
	if err != nil {
		t.Fatal(err)
	}

	u := "https://v16-web.tiktok.com/video/tos/alisg/tos-alisg-pve-0037c001/893ed9debcad47a0972b9764ce8b958f/?a=1988&br=3108&bt=1554&cd=0%7C0%7C1&ch=0&cr=0&cs=0&cv=1&dr=0&ds=3&er=&expire=1612366949&l=20210203094221010189056034465A38F4&lr=tiktok_m&mime_type=video_mp4&pl=0&policy=2&qs=0&rc=anRqZG5veGdpeTMzPDczM0ApZWhkNmVmNjtmNzw0PDo6O2dxMDUyZC1ybTFfLS00MTRzcy1iMl8yYjVjLzAvLzRfYDM6Yw%3D%3D&signature=be6bc8aac838fc27673e325159eb922f&tk=tt_webid_v2&vl=&vr="
	cookie := `ak_bmsc=505052C307435CDBCBA6465C0838CFC817302054CB670000FD6F1A60E1932B48~plm2kbufRNrJpfLKSTD7zqFK93Sm6HDEA+GwtLYJKuTIwRDQ7jbhs4dPUuvBR5e20VZnfn92YI4XiXdwlcLRBV1hFd3NYB1ANKa1j/g/bu0h+roqigiQcFtB4mm2R6lR5GJWCGmy3wUgsLBRGXfy4f7oFAWv3/5vCdfBiR29374u0SX0C24KsPQVJi3VuCga0O+nxvasKwH1Fud1Dz22jO8Vkg5Ghfu3wDahy6ft+vIryTj5fJ24Z1uYCizD30ePXRTFlrHXrrCuQswlP6vQrRKuhQXNUmnpGyoj0AQ3vRXjM=; bm_mi=85683063611B0CB2FF5374863035A4ED~4N8Dn5sUVxAsJHVbRLnYNX0SZp7I+BXelrzQjUC8eBjClepWfGgZRMga3Xev+r/IBfOV2d/SRhehkC1bd10opCoE9/a2/rfuzyGvI6hyMRbsOZ7rGABP+5USmEGQOwWqekB1lrrki8Uia51Hl7QUOVodp27/spvijE0ikGVfh4rGQDSRHMbHX8szk7Qf9p8RR3j0yzrwR0UU/j+NVHkElFotkzW5jC996fngk3pe0t8unp6ZjyCNm+qfE4UbSYsJxpACw9VbOnZ0NbIWBhqbqg==; bm_sz=462FD60833344BEFD43F9E23697D4192~YAAQVCAwF4+wh2Z3AQAAxHdFZwrVMEOt2C8htTE380NfcsMt4P8fGwK+P05VgXY/3AvSko4mMpQt4N+j0mwpMs56HJG02/ulkldaFpqN6IUWfW5eVvtdV+Tit7WEROcBX4kRhU/Yyx8n/TukjyDzfL7DFr/656AZRohHm4oChY5mp6+JDiLzTj8JBP8sr9gd; _abck=10F617BF12D330E1276E93B84176B1D9~-1~YAAQVCAwF5Cwh2Z3AQAAxHdFZwW0CWwvWT3qP0vCJypEdTyBhGaYRA3eqxeUN2ntitjwEoEBnLg+kOi5cbHAy2BfwrpnTGzTF+zSQP6mt35QyBIxlMp1L6mDr3ui/si8K7wqo3j/GuonlQIhjMHkwQpQ+9LkWGujxE58jSFqEpHbu872u51BR5q6aQRK9eyHT/fN/l8nubn8XciCtFXzJG7EYZXDgsn6bGlhinPeKFKF0Y5/MNaBWfiBWRSOBRW1CM4clIo8d/Q05t6Evtg4QLrj4jitjuHnfUIyIbklmUm1G3YRC2ucn0+i~-1~-1~-1`
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "identity;q=1, *;q=0")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,pl;q=0.7,zh-TW;q=0.6,ca;q=0.5,mt;q=0.4")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Referer", "https://www.tiktok.com/")
	req.Header.Set("sec-ch-ua", `"Chromium";v="88", "Google Chrome";v="88", ";Not A Brand";v="99"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("Sec-Fetch-Dest", "video")
	req.Header.Set("Sec-Fetch-Mode", "no-cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.96 Safari/537.36")
	req.Header.Set("Range", "bytes=0-")
	resp, err := client.DoWithOptions(context.Background(), req, http.Options{EnableProxy: false})
	if err != nil {
		t.Fatalf("proxyClient.DoWithOptions() error = %v", err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("status: %v, data: %d", resp.StatusCode, len(data))
}
