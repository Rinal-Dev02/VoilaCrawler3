package cli

import (
	"errors"
	"fmt"
	rhttp "net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/gammazero/deque"
	"github.com/urfave/cli/v2"
	"github.com/voiladev/VoilaCrawler/pkg/checker"
	"github.com/voiladev/VoilaCrawler/pkg/context"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/item"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	"github.com/voiladev/VoilaCrawler/pkg/net/http/cookiejar"
	pbCrawlItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	pbProxy "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/VoilaCrawler/pkg/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/randutil"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func localCommand(ctx context.Context, app *App, newer crawler.NewCrawler, extraFlags []cli.Flag) *cli.Command {
	flags := []cli.Flag{
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
			Name:  "enable-lifo",
			Usage: "Enable queue LIFO for requests",
		},
		&cli.BoolFlag{
			Name:  "disable-checker",
			Usage: "Disable result checker",
		},
		&cli.BoolFlag{
			Name:  "disable-proxy",
			Usage: "Disable proxy",
		},
		&cli.BoolFlag{
			Name:  "pretty",
			Usage: "print item detail in pretty json",
		},
		&cli.BoolFlag{
			Name:    "report",
			Aliases: []string{"r"},
			Usage:   "print item detail in table model",
		},
	}
	flags = append(flags, extraFlags...)
	flags = append(flags, []cli.Flag{
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"v"},
			Usage:   "Verbose model, vv for in detail model",
		},
		&cli.BoolFlag{
			Name:   "vv",
			Usage:  "more verbose model",
			Hidden: true,
		},
		&cli.BoolFlag{
			Name:    "debug",
			Usage:   "Enable debug[Deprecated], use -v instead",
			EnvVars: []string{"DEBUG"},
		},
	}...)

	var (
		subcmds              []*cli.Command
		supportedMethodNames []string
	)
	if _, ok := newer.(crawler.ProductCrawler); ok {
		productCrawlerType := reflect.TypeOf((*crawler.ProductCrawler)(nil)).Elem()
		for i := 0; i < productCrawlerType.NumMethod(); i++ {
			method := productCrawlerType.Method(i)
			supportedMethodNames = append(supportedMethodNames, method.Name)
		}
		callcmd := cli.Command{
			Name:        "call",
			Usage:       "call remote methods",
			Description: "call remote methods",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "name",
					Usage: fmt.Sprintf("method name, supported include %s", strings.Join(supportedMethodNames, ",")),
				},
				&cli.StringFlag{
					Name:  "param",
					Usage: "function params",
				},
			},
			Action: func(c *cli.Context) error {
				name := c.String("name")
				if name == "" {
					return cli.NewExitError("invalid method name", 1)
				}
				if !func() bool {
					for _, n := range supportedMethodNames {
						if name == n {
							return true
						}
					}
					return false
				}() {
					return cli.NewExitError("invalid method name", 1)
				}
				param := c.String("param")

				logger := glog.New(glog.LogLevelInfo)
				verbose := c.Bool("verbose") || c.Bool("debug") || c.Bool("vv")
				if verbose {
					logger.SetLevel(glog.LogLevelDebug)
					os.Setenv("DEBUG", "1")
				}

				proxyAddr := c.String("proxy-addr")
				if proxyAddr == "" {
					return cli.NewExitError("proxy address not specified", 1)
				}

				jar := cookiejar.New()
				client, err := proxy.NewProxyClient(proxyAddr, jar, logger)
				if err != nil {
					logger.Error(err)
					return cli.NewExitError(err, 1)
				}

				val, err := newer.New(c, client, logger)
				if err != nil {
					logger.Error(err)
					return cli.NewExitError(err, 1)
				}

				node := reflect.ValueOf(val)
				caller := node.MethodByName(name)
				errType := caller.Type().Out(1)
				if !errType.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
					return cli.NewExitError(fmt.Sprintf("last output argument of method %s must be implement error interface", name), 1)
				}
				var (
					inArgCount  = caller.Type().NumIn()
					outArgCount = caller.Type().NumOut()
					inputArgs   []reflect.Value
				)
				if inArgCount > 2 || outArgCount != 2 {
					return cli.NewExitError("method define errors, method must define not more than 2 input args, with only two out args", 1)
				}
				switch inArgCount {
				case 0:
				case 1:
					ctx := context.WithValue(app.ctx, context.TracingIdKey, randutil.MustNewRandomID())
					ctx = context.WithValue(ctx, context.JobIdKey, randutil.MustNewRandomID())
					inputArgs = append(inputArgs, reflect.ValueOf(ctx))
				case 2:
					ctx := context.WithValue(app.ctx, context.TracingIdKey, randutil.MustNewRandomID())
					ctx = context.WithValue(ctx, context.JobIdKey, randutil.MustNewRandomID())
					inputArgs = append(inputArgs, reflect.ValueOf(ctx), reflect.ValueOf(param))
				}

				vals := caller.Call(inputArgs)
				if !vals[1].IsNil() {
					return cli.NewExitError(vals[1].Interface(), 1)
				}

				switch val := vals[0].Interface().(type) {
				case []*pbCrawlItem.Category:
					var prettyPrint func(cate *pbCrawlItem.Category, depth int)

					prettyPrint = func(cate *pbCrawlItem.Category, depth int) {
						pending := ""
						if cate.Url != "" {
							pending = " : " + cate.Url
						}
						count := ""
						name := cate.Name
						if len(cate.Children) > 0 {
							count = fmt.Sprintf(" (%d)/", len(cate.Children))
						}
						fmt.Printf("%s%s%s%s\n", strings.Repeat("    ", depth), name, count, pending)
						for _, child := range cate.Children {
							prettyPrint(child, depth+1)
						}
					}
					for _, cate := range val {
						prettyPrint(cate, 0)
					}
				}
				return nil
			},
		}
		subcmds = append(subcmds, &callcmd)
	}

	return &cli.Command{
		Name:        "test",
		Usage:       "local test",
		Description: "local test",
		Subcommands: subcmds,
		Flags:       flags,
		Action: func(c *cli.Context) error {
			logger := glog.New(glog.LogLevelInfo)

			verbose := c.Bool("verbose") || c.Bool("debug") || c.Bool("vv")
			if verbose {
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
			cw, err := newer.New(c, client, logger)
			if err != nil {
				logger.Error(err)
				return cli.NewExitError(err, 1)
			}

			node := cw.(crawler.Crawler)
			var (
				reqFilter = map[string]struct{}{}
				reqQueue  deque.Deque
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

				reqQueue.PushBack(req)
				reqCount += 1
				reqFilter[req.URL.String()] = struct{}{}
				if host == "" {
					host = req.Host
				}
			}
			if reqCount == 0 {
				for _, req := range node.NewTestRequest(context.Background()) {
					reqQueue.PushBack(req)
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

					if !c.Bool("disable-checker") {
						if err := checker.Check(ctx, i, logger); err != nil {
							return err
						}
					}

					i = i.WithContext(ctx)
					reqQueue.PushBack(i)

					sharedVals := context.RetrieveAllValues(ctx)
					vals := map[string]interface{}{}
					for k, v := range sharedVals {
						if ks, ok := k.(string); ok {
							vals[ks] = v
						}
					}
					logger.Debugf("Appended %s, data=%+v", i.URL, vals)

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

					if !c.Bool("disable-checker") {
						if err := checker.Check(ctx, val, logger); err != nil {
							return err
						}
					}
					if c.Bool("report") {
						item.Report(val, logger)
					}
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
					default:
						if reqQueue.Len() == 0 {
							return
						}

						if c.Bool("enable-lifo") {
							if v := reqQueue.PopBack(); v != nil {
								req = v.(*http.Request)
							}
						} else {
							if v := reqQueue.PopFront(); v != nil {
								req = v.(*http.Request)
							}
						}
					}

					if err = func(i *http.Request) error {
						nctx, cancel := context.WithTimeout(ctx, time.Minute*5)
						for k, v := range context.RetrieveAllValues(i.Context()) {
							if k == context.ReqIdKey {
								continue
							}
							nctx = context.WithValue(nctx, k, v)
						}
						defer cancel()
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

						var resp *http.Response
						if opts.SkipDoRequest {
							resp = &http.Response{Response: &rhttp.Response{Request: req}}
						} else {
							resp, err := client.DoWithOptions(nctx, req, httpOpts)
							if err != nil {
								logger.Error(err)
								return err
							}
							if resp.Body == nil {
								return errors.New("not response found")
							}
							defer resp.Body.Close()

							if c.Bool("vv") {
								data, _ := resp.RawBody()
								logger.Debugf("%s", data)
							}
						}
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
