// Linktree

package main

import (
	"encoding/json"
	"html"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/voiladev/VoilaCrawler/pkg/cli"
	"github.com/voiladev/VoilaCrawler/pkg/context"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	pbItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	pbProxy "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
)

type _Crawler struct {
	httpClient http.Client

	categoryPathMatcher     *regexp.Regexp
	categoryJsonPathMatcher *regexp.Regexp
	productGroupPathMatcher *regexp.Regexp
	productPathMatcher      *regexp.Regexp
	logger                  glog.Log
}

func (_ *_Crawler) New(_ *cli.Context, client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient:              client,
		categoryPathMatcher:     regexp.MustCompile(`^(/[a-z0-9_-]+)?/(women|men)(/[a-z0-9_-]+){1,6}/cat/?$`),
		categoryJsonPathMatcher: regexp.MustCompile(`^/api/product/search/v2/categories/([a-z0-9]+)`),
		productGroupPathMatcher: regexp.MustCompile(`^(/[a-z0-9_-]+)?(/[a-z0-9_-]+){2}/grp/[0-9]+/?$`),
		productPathMatcher:      regexp.MustCompile(`^(/[a-z0-9_-]+)?(/[a-z0-9_-]+){2}/prd/[0-9]+/?$`),
		logger:                  logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "__linktree__"
}

// Version
func (c *_Crawler) Version() int32 {
	return 1
}

// CrawlOptions
func (c *_Crawler) CrawlOptions(u *url.URL) *crawler.CrawlOptions {
	options := crawler.NewCrawlOptions()
	options.EnableHeadless = false
	options.EnableSessionInit = false
	options.Reliability = pbProxy.ProxyReliability_ReliabilityRealtime
	options.MustCookies = append(options.MustCookies)

	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"linktr.ee"}
}

func (c *_Crawler) CanonicalUrl(rawurl string) (string, error) {
	return rawurl, nil
}

func textClean(s string) string {
	return strings.TrimSpace(html.UnescapeString(s))
}

type linkDetail struct {
	Props struct {
		PageProps struct {
			Account struct {
				ID                         int    `json:"id"`
				UUID                       string `json:"uuid"`
				Username                   string `json:"username"`
				Tier                       string `json:"tier"`
				IsActive                   bool   `json:"isActive"`
				ProfilePictureURL          string `json:"profilePictureUrl"`
				PageTitle                  string `json:"pageTitle"`
				DonationsActive            bool   `json:"donationsActive"`
				CauseBanner                string `json:"causeBanner"`
				IsLogoVisible              bool   `json:"isLogoVisible"`
				SocialLinksPosition        string `json:"socialLinksPosition"`
				UseFooterSignup            bool   `json:"useFooterSignup"`
				UseSignupLink              bool   `json:"useSignupLink"`
				CreatedAt                  int64  `json:"createdAt"`
				UpdatedAt                  int64  `json:"updatedAt"`
				ExpandableLinkCaret        bool   `json:"expandableLinkCaret"`
				CustomAvatar               string `json:"customAvatar"`
				PaymentEmail               string `json:"paymentEmail"`
				ShowRepeatVisitorSignupCta bool   `json:"showRepeatVisitorSignupCta"`
				Owner                      struct {
					ID              int  `json:"id"`
					IsEmailVerified bool `json:"isEmailVerified"`
				} `json:"owner"`
				PageMeta     interface{}   `json:"pageMeta"`
				Integrations []interface{} `json:"integrations"`
				Links        []struct {
					ID       int    `json:"id"`
					Type     string `json:"type"`
					Title    string `json:"title"`
					Position int    `json:"position"`
					URL      string `json:"url"`
					Context  struct {
						ServiceIntegration struct {
							ID                  string `json:"id"`
							Type                string `json:"type"`
							Title               string `json:"title"`
							Status              string `json:"status"`
							CurrencyCode        string `json:"currencyCode"`
							LocationID          string `json:"locationId"`
							SquareApplicationID string `json:"squareApplicationId"`
							SquareIntegrationID string `json:"squareIntegrationId"`
						} `json:"serviceIntegration"`
						Options []struct {
							Title  string `json:"title"`
							Amount int    `json:"amount"`
						} `json:"options"`
						DescriptionMessage string `json:"descriptionMessage"`
						SuccessMessage     string `json:"successMessage"`
						RequireDetails     bool   `json:"requireDetails"`
						RequireTax         bool   `json:"requireTax"`
						TaxRate            int    `json:"taxRate"`
						HelpCoverFees      bool   `json:"helpCoverFees"`
					} `json:"context,omitempty"`
					Rules struct {
						Gate struct {
							ActiveOrder interface{} `json:"activeOrder"`
							Age         interface{} `json:"age"`
						} `json:"gate"`
					} `json:"rules"`
				} `json:"links"`
				SocialLinks []struct {
					Type string `json:"type"`
					URL  string `json:"url"`
				} `json:"socialLinks"`
				Theme struct {
					Key string `json:"key"`
				} `json:"theme"`
			} `json:"account"`
			IsProfileVerified  bool        `json:"isProfileVerified"`
			HasConsentedToView bool        `json:"hasConsentedToView"`
			Username           string      `json:"username"`
			PageTitle          string      `json:"pageTitle"`
			Description        interface{} `json:"description"`
			SocialLinks        []struct {
				Type string `json:"type"`
				URL  string `json:"url"`
			} `json:"socialLinks"`
			Integrations []interface{} `json:"integrations"`
			Theme        struct {
				Key    string `json:"key"`
				Colors struct {
					Body           string `json:"body"`
					LinkBackground string `json:"linkBackground"`
					LinkText       string `json:"linkText"`
					LinkShadow     string `json:"linkShadow"`
				} `json:"colors"`
				Fonts struct {
					Primary  string `json:"primary"`
					FontSize string `json:"fontSize"`
				} `json:"fonts"`
				Components struct {
					ProfileBackground struct {
						BackgroundColor string   `json:"backgroundColor"`
						BackgroundStyle string   `json:"backgroundStyle"`
						BackgroundImage []string `json:"backgroundImage"`
					} `json:"ProfileBackground"`
					Header struct {
						FontWeight int    `json:"fontWeight"`
						FontSize   string `json:"fontSize"`
					} `json:"Header"`
					ProfileDescription struct {
						FontSize string `json:"fontSize"`
					} `json:"ProfileDescription"`
					LinkContainer struct {
						BorderType string `json:"borderType"`
						StyleType  string `json:"styleType"`
					} `json:"LinkContainer"`
					LinkText struct {
						FontWeight int `json:"fontWeight"`
					} `json:"LinkText"`
					LinkHeader struct {
						FontWeight int `json:"fontWeight"`
					} `json:"LinkHeader"`
					SocialLink struct {
						Fill string `json:"fill"`
					} `json:"SocialLink"`
					Footer struct {
						Logo   string `json:"logo"`
						URL    string `json:"url"`
						Height string `json:"height"`
					} `json:"Footer"`
				} `json:"components"`
			} `json:"theme"`
			MetaTitle         string `json:"metaTitle"`
			MetaDescription   string `json:"metaDescription"`
			ProfilePictureURL string `json:"profilePictureUrl"`
			Links             []struct {
				ID      string `json:"id"`
				Title   string `json:"title"`
				Context struct {
					ServiceIntegration struct {
						ID                  string `json:"id"`
						Type                string `json:"type"`
						Title               string `json:"title"`
						Status              string `json:"status"`
						CurrencyCode        string `json:"currencyCode"`
						LocationID          string `json:"locationId"`
						SquareApplicationID string `json:"squareApplicationId"`
						SquareIntegrationID string `json:"squareIntegrationId"`
					} `json:"serviceIntegration"`
					Options []struct {
						Title  string `json:"title"`
						Amount int    `json:"amount"`
					} `json:"options"`
					DescriptionMessage string `json:"descriptionMessage"`
					SuccessMessage     string `json:"successMessage"`
					RequireDetails     bool   `json:"requireDetails"`
					RequireTax         bool   `json:"requireTax"`
					TaxRate            int    `json:"taxRate"`
					HelpCoverFees      bool   `json:"helpCoverFees"`
				} `json:"context,omitempty"`
				Animation       interface{} `json:"animation"`
				Thumbnail       interface{} `json:"thumbnail"`
				URL             string      `json:"url"`
				AmazonAffiliate interface{} `json:"amazonAffiliate"`
				Type            string      `json:"type"`
				Rules           struct {
					Gate struct {
						ActiveOrder interface{} `json:"activeOrder"`
						Age         interface{} `json:"age"`
					} `json:"gate"`
				} `json:"rules"`
				Position int         `json:"position"`
				Locked   interface{} `json:"locked"`
			} `json:"links"`
			LeapLink       interface{} `json:"leapLink"`
			IsOwner        bool        `json:"isOwner"`
			IsLogoVisible  bool        `json:"isLogoVisible"`
			MobileDetected bool        `json:"mobileDetected"`
			Stage          string      `json:"stage"`
			Environment    struct {
				STRIPEPAYMENTSAPIENDPOINT string `json:"STRIPE_PAYMENTS_API_ENDPOINT"`
				STRIPEPUBLISHABLEKEY      string `json:"STRIPE_PUBLISHABLE_KEY"`
				PAYPALPAYMENTSAPIENDPOINT string `json:"PAYPAL_PAYMENTS_API_ENDPOINT"`
				PAYPALPAYMENTSCLIENTID    string `json:"PAYPAL_PAYMENTS_CLIENT_ID"`
				METAIMAGEURL              string `json:"META_IMAGE_URL"`
			} `json:"environment"`
			ContentGating       string `json:"contentGating"`
			HasSensitiveContent bool   `json:"hasSensitiveContent"`
		} `json:"pageProps"`
		NSSP bool `json:"__N_SSP"`
	} `json:"props"`
	Page string `json:"page"`
}

func (c *_Crawler) Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	dom, err := resp.Selector()
	if err != nil {
		c.logger.Error(err)
		return err
	}

	err = func() error {
		rawdata := strings.TrimSpace(dom.Find(`#__NEXT_DATA__`).Text())
		var viewData linkDetail
		if err := json.Unmarshal([]byte(rawdata), &viewData); err != nil {
			c.logger.Error(err)
			return err
		}
		item := pbItem.Linktree_Item{
			Profile: &pbItem.Linktree_Item_Profile{
				Id:          strconv.Format(viewData.Props.PageProps.Account.ID),
				Name:        viewData.Props.PageProps.Account.Username,
				Avatar:      viewData.Props.PageProps.Account.ProfilePictureURL,
				LinktreeUrl: resp.Request.URL.String(),
				Email:       viewData.Props.PageProps.Account.PaymentEmail,
			},
		}
		for _, rawlink := range viewData.Props.PageProps.Links {
			if rawlink.URL == "" {
				continue
			}
			link := pbItem.Linktree_Item_Link{
				Id:    rawlink.ID,
				Type:  rawlink.Type,
				Title: rawlink.Title,
				Url:   rawlink.URL,
			}
			item.Links = append(item.Links, &link)
		}
		for _, rawslink := range viewData.Props.PageProps.SocialLinks {
			link := pbItem.Linktree_Item_SocialLink{
				Type: rawslink.Type,
				Url:  rawslink.URL,
			}
			item.SocialLinks = append(item.SocialLinks, &link)
		}
		return yield(ctx, &item)
	}()
	if err == nil {
		return nil
	}
	c.logger.Errorf("extract from embeded json failed, error=%s", err)

	// extract from html
	item := pbItem.Linktree_Item{
		Profile: &pbItem.Linktree_Item_Profile{
			Name:   strings.TrimSpace(dom.Find(`div>h1[id]`).Text()),
			Avatar: dom.Find(`img[data-testid="ProfileImage"]`).AttrOr("src", ""),
		},
	}
	sel := dom.Find(`div>div[data-testid="StyledContainer"]>a`)
	for i := range sel.Nodes {
		node := sel.Eq(i)
		title := strings.TrimSpace(node.Text())
		href := node.AttrOr("href", "")
		u, _ := url.Parse(href)
		if u == nil || href == "" {
			c.logger.Errorf(`loaded invalid link "%s"`, href)
			continue
		}
		link := pbItem.Linktree_Item_Link{
			Title: title,
			Url:   u.String(),
		}
		item.Links = append(item.Links, &link)
	}
	sel2 := dom.Find(`div>div>a[data-testid="SocialIcon"]`)
	for i := range sel2.Nodes {
		node := sel2.Eq(i)
		href := node.AttrOr("href", "")

		link := pbItem.Linktree_Item_SocialLink{
			Type: "",
			Url:  href,
		}
		// TODO: added more supported social medias
		switch {
		case strings.Contains(href, "tik"):
			link.Type = "TIKTOK"
		case strings.Contains(href, "insta"):
			link.Type = "INSTAGRAM"
		case strings.Contains(href, "youtu"):
			link.Type = "YOUTUBE"
		case strings.Contains(href, "linkedin"):
			link.Type = "LINKEDIN"
		case strings.Contains(href, "clubhouse"):
			link.Type = "CLUBHOUSE"
		case strings.Contains(href, "signal"):
			link.Type = "SIGNAL"
		case strings.Contains(href, "pinterest"):
			link.Type = "PINTEREST"
		case strings.Contains(href, "whatsapp"):
			link.Type = "WHATSAPP"
		case strings.Contains(href, "facebook"):
			link.Type = "FACEBOOK"
		case strings.Contains(href, "t.me"):
			link.Type = "TELEGRAM"
		case strings.Contains(href, "@") || strings.Contains(href, "mailto:"):
			link.Type = "EMAIL_ADDRESS"
		default:
			link.Type = "unknown"
		}
		item.SocialLinks = append(item.SocialLinks, &link)
	}
	return yield(ctx, &item)
}

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		"https://linktr.ee/bytrendypep/",
		"https://linktr.ee/kellyydoan",
		"https://linktr.ee/Clairebridges",
	} {
		req, _ := http.NewRequest(http.MethodGet, u, nil)
		reqs = append(reqs, req)
	}
	return reqs
}

func (c *_Crawler) CheckTestResponse(ctx context.Context, resp *http.Response) error {
	if err := c.Parse(ctx, resp, func(c context.Context, i interface{}) error {
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// main func is the entry of golang program. this will not be used by plugin, just for local spider test.
func main() {
	cli.NewApp(&_Crawler{}).Run(os.Args)
}
