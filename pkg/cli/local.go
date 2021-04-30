package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/voiladev/go-crawler/pkg/crawler"
	"github.com/voiladev/go-crawler/pkg/net/http"
	"github.com/voiladev/go-crawler/pkg/net/http/cookiejar"
	"github.com/voiladev/go-crawler/pkg/proxy"
	pbProxy "github.com/voiladev/go-crawler/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/randutil"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func localCommand(ctx context.Context, newFunc crawler.New) *cli.Command {
	return &cli.Command{
		Name:        "test",
		Usage:       "local test",
		Description: "local test",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "proxy-addr",
				Usage:   "proxy server address",
				EnvVars: []string{"VOILA_PROXY_URL"},
			},
			&cli.StringSliceFlag{
				Name:  "target",
				Usage: "use this target url for test if provided",
			},
			&cli.StringSliceFlag{
				Name:  "level",
				Usage: "proxy level, 1,2,3",
			},
			&cli.BoolFlag{
				Name:  "pretty",
				Usage: "print item detail in pretty",
			},
			&cli.BoolFlag{
				Name:    "debug",
				Usage:   "Enable debug",
				EnvVars: []string{"DEBUG"},
			},
		},
		Action: func(c *cli.Context) error {
			logger := glog.New(glog.LogLevelInfo)
			if c.Bool("debug") {
				logger.SetLevel(glog.LogLevelDebug)
				os.Setenv("DEBUG", "1")
			}

			proxyAddr := c.String("proxy-addr")
			if proxyAddr == "" {
				return errors.New("proxy address not specified")
			}
			level := c.Int("level")

			jar := cookiejar.New()
			client, err := proxy.NewProxyClient(proxyAddr, jar, logger)
			if err != nil {
				logger.Error(err)
				return cli.NewExitError(err, 1)
			}
			node, err := newFunc(client, logger)
			if err != nil {
				logger.Error(err)
				return cli.NewExitError(err, 1)
			}
			var (
				reqFilter = map[string]struct{}{}
				reqChan   = make(chan *http.Request, 100)
				reqCount  = 0
				host      string
			)
			for _, rawurl := range c.StringSlice("target") {
				req, err := http.NewRequest(http.MethodGet, rawurl, nil)
				if err != nil {
					return cli.NewExitError(err, 1)
				}
				reqChan <- req
				reqCount += 1
				reqFilter[req.URL.String()] = struct{}{}
				if host == "" {
					host = req.Host
				}
			}
			if reqCount == 0 {
				for _, req := range node.NewTestRequest(ctx) {
					reqChan <- req
					reqCount += 1
					reqFilter[req.URL.String()] = struct{}{}
					if host == "" {
						host = req.Host
					}
				}
			}

			callback := func(ctx context.Context, val interface{}) error {
				switch i := val.(type) {
				case *http.Request:
					if _, ok := reqFilter[i.URL.String()]; ok {
						return nil
					}
					reqFilter[i.URL.String()] = struct{}{}

					opts := node.CrawlOptions(i.URL)

					// process logic of sub request

					// init custom headers
					for k := range opts.MustHeader {
						i.Header.Set(k, opts.MustHeader.Get(k))
					}

					// init custom cookies
					for _, c := range opts.MustCookies {
						if strings.HasPrefix(i.URL.Path, c.Path) || c.Path == "" {
							val := fmt.Sprintf("%s=%s", c.Name, c.Value)
							if c := i.Header.Get("Cookie"); c != "" {
								i.Header.Set("Cookie", c+"; "+val)
							} else {
								i.Header.Set("Cookie", val)
							}
						}
					}

					// set scheme,host for sub requests. for the product url in category page is just the path without hosts info.
					// here is just the test logic. when run the spider online, the controller will process automatically
					if i.URL.Scheme == "" {
						i.URL.Scheme = "https"
					}
					if i.URL.Host == "" {
						i.URL.Host = host
					}

					reqChan <- i
					return nil
				default:
					marshaler := protojson.MarshalOptions{}
					if c.Bool("pretty") {
						marshaler.Indent = " "
					}

					// output the result
					data, err := marshaler.Marshal(i.(proto.Message))
					if err != nil {
						return err
					}
					logger.Debugf("data: %s", data)
				}
				return nil
			}

			ctx = context.WithValue(ctx, "tracing_id", randutil.MustNewRandomID())
			for {
				var req *http.Request
				select {
				case <-ctx.Done():
					return nil
				case req, _ = <-reqChan:
					if req == nil {
						return nil
					}
				default:
					return nil
				}
				if err := func(i *http.Request) error {
					logger.Infof("Access %s", i.URL)

					canUrl, _ := node.CanonicalUrl(i.URL.String())
					logger.Debugf("Canonical Url %s", canUrl)

					nctx, cancel := context.WithTimeout(ctx, time.Minute*5)
					defer cancel()
					nctx = context.WithValue(nctx, "req_id", randutil.MustNewRandomID())

					opts := node.CrawlOptions(req.URL)
					l := opts.Reliability
					if l == 0 {
						l = pbProxy.ProxyReliability(level)
					}
					resp, err := client.DoWithOptions(nctx, req, http.Options{
						EnableProxy:       true,
						EnableHeadless:    opts.EnableHeadless,
						EnableSessionInit: opts.EnableSessionInit,
						KeepSession:       opts.KeepSession,
						DisableCookieJar:  opts.DisableCookieJar,
						DisableRedirect:   opts.DisableRedirect,
						Reliability:       l,
					})
					if err != nil {
						return err
					}
					defer resp.Body.Close()
					return node.Parse(ctx, resp, callback)
				}(req); err != nil {
					if !errors.Is(err, context.Canceled) {
						logger.Error(err)
					}
					return err
				}
			}
		},
	}
}
