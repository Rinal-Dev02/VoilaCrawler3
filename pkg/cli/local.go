package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/voiladev/VoilaCrawler/pkg/context"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/item"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	"github.com/voiladev/VoilaCrawler/pkg/net/http/cookiejar"
	"github.com/voiladev/VoilaCrawler/pkg/proxy"
	pbProxy "github.com/voiladev/VoilaCrawler/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/randutil"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func localCommand(ctx context.Context, app *App, newFunc crawler.New) *cli.Command {
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
				Name:  "type",
				Usage: "target type to crawl",
				Value: cli.NewStringSlice(item.SupportedTypes()...),
			},
			&cli.StringSliceFlag{
				Name:  "level",
				Usage: "proxy level, 1,2,3",
			},
			&cli.BoolFlag{
				Name:  "enable-headless",
				Usage: "Enable headless",
			},
			&cli.BoolFlag{
				Name:  "enable-session-init",
				Usage: "Enable session init",
			},
			&cli.BoolFlag{
				Name:  "disable-proxy",
				Usage: "Disable proxy",
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
			disableProxy := c.Bool("disable-proxy")

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
			typs := c.StringSlice("type")
			typCtx := context.WithValue(context.Background(), crawler.TargetTypeKey, strings.Join(typs, ","))
			for _, rawurl := range c.StringSlice("target") {
				req, err := http.NewRequestWithContext(typCtx, http.MethodGet, rawurl, nil)
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
				for _, req := range node.NewTestRequest(context.Background()) {
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

					// set scheme,host for sub requests. for the product url in category page is just the path without hosts info.
					// here is just the test logic. when run the spider online, the controller will process automatically
					if i.URL.Scheme == "" {
						i.URL.Scheme = "https"
					}
					if i.URL.Host == "" {
						i.URL.Host = host
					}

					i = i.WithContext(ctx)
					select {
					case reqChan <- i:
						logger.Debugf("appended %s", i.URL.String())
					default:
						logger.Warnf("ignored %s, too many requests", i.URL.String())
					}
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

			ctx = context.WithValue(ctx, context.TracingIdKey, randutil.MustNewRandomID())
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			go func() {
				defer app.cancelFunc()

				var req *http.Request
				for {
					req = nil
					select {
					case <-ctx.Done():
						return
					case req, _ = <-reqChan:
						if req == nil {
							return
						}
					default:
						return
					}

					if err = func(i *http.Request) error {
						logger.Infof("Access %s", i.URL)

						nctx, cancel := context.WithTimeout(ctx, time.Minute*5)
						defer cancel()

						if v := i.Context().Value(context.TargetTypeKey); v != nil {
							nctx = context.WithValue(nctx, context.TargetTypeKey, v)
						}
						nctx = context.WithValue(nctx, context.ReqIdKey, randutil.MustNewRandomID())

						opts := node.CrawlOptions(req.URL)
						httpOpts := http.Options{
							EnableProxy:       !disableProxy,
							EnableHeadless:    opts.EnableHeadless,
							EnableSessionInit: opts.EnableSessionInit,
							KeepSession:       opts.KeepSession,
							DisableCookieJar:  opts.DisableCookieJar,
							DisableRedirect:   opts.DisableRedirect,
							Reliability:       opts.Reliability,
						}
						if c.IsSet("enable-headless") {
							httpOpts.EnableHeadless = c.Bool("enable-headless")
						}
						if c.IsSet("enable-session-init") {
							httpOpts.EnableSessionInit = c.Bool("enable-session-init")
						}
						if c.IsSet("level") {
							httpOpts.Reliability = pbProxy.ProxyReliability(c.Int("level"))
						}

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

						resp, err := client.DoWithOptions(nctx, req, httpOpts)
						if err != nil {
							logger.Error(err)
							return err
						}
						defer resp.Body.Close()

						return node.Parse(nctx, resp, callback)
					}(req); err != nil {
						if !errors.Is(err, context.Canceled) {
							logger.Error(err)
						}
						return
					}
				}
			}()
			<-ctx.Done()
			return err
		},
	}
}
