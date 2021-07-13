package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/voiladev/VoilaCrawler/pkg/cli"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/media"
	"github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/regulation"
	pbItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	pbProxy "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
	"google.golang.org/protobuf/types/known/anypb"
)

type _Crawler struct {
	httpClient http.Client

	categoryPathMatcher   *regexp.Regexp
	productPathMatcher    *regexp.Regexp
	productApiPathMatcher *regexp.Regexp
	logger                glog.Log
}

func New(client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient:            client,
		categoryPathMatcher:   regexp.MustCompile(`^/([/A-Za-z_-]+)$`),
		productPathMatcher:    regexp.MustCompile(`^/shop/products([/A-Za-z0-9_-]+)$`),
		productApiPathMatcher: regexp.MustCompile(`^/api/v2/products([/A-Za-z0-9_-]+)$`),
		logger:                logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "3c608f04da5f4bc6927b473ebcebd17d"
}

// Version
func (c *_Crawler) Version() int32 {
	return 1
}

// CrawlOptions
func (c *_Crawler) CrawlOptions(u *url.URL) *crawler.CrawlOptions {
	options := crawler.NewCrawlOptions()
	options.EnableHeadless = false
	//options.LoginRequired = false
	options.EnableSessionInit = true
	options.Reliability = pbProxy.ProxyReliability_ReliabilityMedium
	options.MustCookies = append(options.MustCookies,
		&http.Cookie{Name: "ckm-ctx-sf", Value: `%2F`, Path: "/"},
	)
	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"*.mejuri.com"}
}

func (c *_Crawler) CanonicalUrl(rawurl string) (string, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	if u.Host == "" {
		u.Host = "www.mejuri.com"
	}
	if c.productPathMatcher.MatchString(u.Path) {
		u.RawQuery = ""
		return u.String(), nil
	}
	return rawurl, nil
}

var countriesPrefix = map[string]struct{}{"/ad": {}, "/ae": {}, "/ar-ae": {}, "/af": {}, "/ag": {}, "/ai": {}, "/al": {}, "/am": {}, "/an": {}, "/ao": {}, "/aq": {}, "/ar": {}, "/at": {}, "/au": {}, "/aw": {}, "/az": {}, "/ba": {}, "/bb": {}, "/bd": {}, "/be": {}, "/bf": {}, "/bg": {}, "/bh": {}, "/ar-bh": {}, "/bi": {}, "/bj": {}, "/bm": {}, "/bn": {}, "/bo": {}, "/br": {}, "/bs": {}, "/bt": {}, "/bv": {}, "/bw": {}, "/by": {}, "/bz": {}, "/ca": {}, "/cc": {}, "/cf": {}, "/cg": {}, "/ch": {}, "/ci": {}, "/ck": {}, "/cl": {}, "/cm": {}, "/cn": {}, "/co": {}, "/cr": {}, "/cv": {}, "/cx": {}, "/cy": {}, "/cz": {}, "/de": {}, "/dj": {}, "/dk": {}, "/dm": {}, "/do": {}, "/dz": {}, "/ec": {}, "/ee": {}, "/eg": {}, "/ar-eg": {}, "/eh": {}, "/es": {}, "/et": {}, "/fi": {}, "/fj": {}, "/fk": {}, "/fm": {}, "/fo": {}, "/fr": {}, "/ga": {}, "/uk": {}, "/gd": {}, "/ge": {}, "/gf": {}, "/gg": {}, "/gh": {}, "/gi": {}, "/gl": {}, "/gm": {}, "/gn": {}, "/gp": {}, "/gq": {}, "/gr": {}, "/gt": {}, "/gu": {}, "/gw": {}, "/gy": {}, "/hk": {}, "/hn": {}, "/hr": {}, "/ht": {}, "/hu": {}, "/ic": {}, "/id": {}, "/ie": {}, "/il": {}, "/in": {}, "/io": {}, "/iq": {}, "/ar-iq": {}, "/is": {}, "/it": {}, "/je": {}, "/jm": {}, "/jo": {}, "/ar-jo": {}, "/jp": {}, "/ke": {}, "/kg": {}, "/kh": {}, "/ki": {}, "/km": {}, "/kn": {}, "/kr": {}, "/kv": {}, "/kw": {}, "/ar-kw": {}, "/ky": {}, "/kz": {}, "/la": {}, "/lb": {}, "/ar-lb": {}, "/lc": {}, "/li": {}, "/lk": {}, "/ls": {}, "/lt": {}, "/lu": {}, "/lv": {}, "/ma": {}, "/mc": {}, "/md": {}, "/me": {}, "/mg": {}, "/mh": {}, "/mk": {}, "/ml": {}, "/mn": {}, "/mo": {}, "/mp": {}, "/mq": {}, "/mr": {}, "/ms": {}, "/mt": {}, "/mu": {}, "/mv": {}, "/mw": {}, "/mx": {}, "/my": {}, "/mz": {}, "/na": {}, "/nc": {}, "/ne": {}, "/nf": {}, "/ng": {}, "/ni": {}, "/nl": {}, "/no": {}, "/np": {}, "/nr": {}, "/nu": {}, "/nz": {}, "/om": {}, "/ar-om": {}, "/pa": {}, "/pe": {}, "/pf": {}, "/pg": {}, "/ph": {}, "/pk": {}, "/pl": {}, "/pm": {}, "/pn": {}, "/pr": {}, "/pt": {}, "/pw": {}, "/py": {}, "/qa": {}, "/ar-qa": {}, "/re": {}, "/ro": {}, "/rs": {}, "/ru": {}, "/rw": {}, "/sa": {}, "/ar-sa": {}, "/sb": {}, "/sc": {}, "/se": {}, "/sg": {}, "/sh": {}, "/si": {}, "/sk": {}, "/sl": {}, "/sm": {}, "/sn": {}, "/sr": {}, "/st": {}, "/sv": {}, "/sz": {}, "/tc": {}, "/td": {}, "/tg": {}, "/th": {}, "/tj": {}, "/tk": {}, "/tl": {}, "/tn": {}, "/to": {}, "/tr": {}, "/tt": {}, "/tv": {}, "/tw": {}, "/tz": {}, "/ua": {}, "/ug": {}, "/uy": {}, "/uz": {}, "/va": {}, "/vc": {}, "/ve": {}, "/vg": {}, "/vi": {}, "/vn": {}, "/vu": {}, "/wf": {}, "/xc": {}, "/ye": {}, "/za": {}, "/zm": {}, "/zw": {}}

func getPathFirstSection(p string) string {
	return "/" + strings.SplitN(strings.TrimPrefix(p, "/"), "/", 2)[0]
}

func (c *_Crawler) Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	prefix := getPathFirstSection(resp.Request.URL.Path)
	if _, ok := countriesPrefix[prefix]; ok {
		req := resp.Request
		req.URL.Path = strings.TrimPrefix(req.URL.Path, prefix)

		opts := c.CrawlOptions(req.URL)
		for k := range opts.MustHeader {
			req.Header.Set(k, opts.MustHeader.Get(k))
		}
		for _, c := range opts.MustCookies {
			if strings.HasPrefix(req.URL.Path, c.Path) || c.Path == "" {
				val := fmt.Sprintf("%s=%s", c.Name, c.Value)
				if c := req.Header.Get("Cookie"); c != "" {
					req.Header.Set("Cookie", c+"; "+val)
				} else {
					req.Header.Set("Cookie", val)
				}
			}
		}
		c.logger.Infof("Access %s", req.URL.String())
		if res, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
			EnableProxy:       true,
			EnableHeadless:    opts.EnableHeadless,
			EnableSessionInit: opts.EnableSessionInit,
			DisableCookieJar:  opts.DisableCookieJar,
			Reliability:       opts.Reliability,
		}); err != nil {
			return err
		} else {
			resp = res
		}
	}

	yieldWrap := func(ctx context.Context, val interface{}) error {
		switch v := val.(type) {
		case *http.Request:
			prefix := getPathFirstSection(v.URL.Path)
			if _, ok := countriesPrefix[prefix]; ok {
				v.URL.Path = strings.TrimPrefix(v.URL.Path, prefix)
			}
			return yield(ctx, v)
		default:
			return yield(ctx, val)
		}
	}

	p := strings.TrimSuffix(resp.RawUrl().Path, "/")
	if p == "/shop/t/type" || p == "/info/gift-guides" || p == "/style-edit" {
		return c.parseCategories(ctx, resp, yield)
	}
	if c.productPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yieldWrap)
	} else if c.productApiPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseProduct(ctx, resp, yieldWrap)
	} else if c.categoryPathMatcher.MatchString(resp.Request.URL.Path) {
		return c.parseCategoryProducts(ctx, resp, yieldWrap)
	}
	return fmt.Errorf("unsupported url %s", resp.Request.URL.String())
}

type categoryStructure struct {
	Props struct {
		PageProps struct {
			Headers struct {
				Accept                    string `json:"accept"`
				AcceptEncoding            string `json:"accept-encoding"`
				AcceptLanguage            string `json:"accept-language"`
				CloudfrontForwardedProto  string `json:"cloudfront-forwarded-proto"`
				CloudfrontIsDesktopViewer string `json:"cloudfront-is-desktop-viewer"`
				CloudfrontIsMobileViewer  string `json:"cloudfront-is-mobile-viewer"`
				CloudfrontIsSmarttvViewer string `json:"cloudfront-is-smarttv-viewer"`
				CloudfrontIsTabletViewer  string `json:"cloudfront-is-tablet-viewer"`
				CloudfrontViewerCountry   string `json:"cloudfront-viewer-country"`
				Host                      string `json:"host"`
				SecChUa                   string `json:"sec-ch-ua"`
				SecChUaMobile             string `json:"sec-ch-ua-mobile"`
				SecFetchDest              string `json:"sec-fetch-dest"`
				SecFetchMode              string `json:"sec-fetch-mode"`
				SecFetchSite              string `json:"sec-fetch-site"`
				SecFetchUser              string `json:"sec-fetch-user"`
				UpgradeInsecureRequests   string `json:"upgrade-insecure-requests"`
				UserAgent                 string `json:"user-agent"`
				Via                       string `json:"via"`
				XAmzCfID                  string `json:"x-amz-cf-id"`
				XAmznTraceID              string `json:"x-amzn-trace-id"`
				XForwardedFor             string `json:"x-forwarded-for"`
				XForwardedPort            string `json:"x-forwarded-port"`
				XForwardedProto           string `json:"x-forwarded-proto"`
				XMejuriAPIHost            string `json:"x-mejuri-api-host"`
				XRequestID                string `json:"x-request-id"`
				ContentLength             int    `json:"content-length"`
				XPoweredBy                string `json:"x-powered-by"`
			} `json:"headers"`
			PagePath string `json:"pagePath"`
			PageData struct {
				Locale struct {
					Code     string      `json:"code"`
					Fallback interface{} `json:"fallback"`
				} `json:"locale"`
				Locales []struct {
					Code         string      `json:"code"`
					Name         string      `json:"name"`
					Default      bool        `json:"default"`
					FallbackCode interface{} `json:"fallbackCode"`
					Sys          struct {
						ID      string `json:"id"`
						Type    string `json:"type"`
						Version int    `json:"version"`
					} `json:"sys"`
				} `json:"locales"`
				Footer struct {
					ID          string `json:"_id"`
					ContentType struct {
						ID string `json:"_id"`
					} `json:"_contentType"`
					Locale   string `json:"_locale"`
					Text     string `json:"text"`
					Type     string `json:"type"`
					Children []struct {
						ID          string `json:"_id"`
						ContentType struct {
							ID string `json:"_id"`
						} `json:"_contentType"`
						Locale   string `json:"_locale"`
						Text     string `json:"text"`
						Type     string `json:"type,omitempty"`
						Children []struct {
							ID          string `json:"_id"`
							ContentType struct {
								ID string `json:"_id"`
							} `json:"_contentType"`
							Locale string `json:"_locale"`
							Text   string `json:"text"`
							URL    string `json:"url"`
							Slug   string `json:"slug,omitempty"`
						} `json:"children"`
						ExtraFields struct {
							Legal bool `json:"legal"`
						} `json:"extraFields,omitempty"`
					} `json:"children"`
					Slug string `json:"slug"`
				} `json:"footer"`
				MobileMenu struct {
					ID          string `json:"_id"`
					ContentType struct {
						ID string `json:"_id"`
					} `json:"_contentType"`
					Locale   string `json:"_locale"`
					Text     string `json:"text"`
					Children []struct {
						ID          string `json:"_id"`
						ContentType struct {
							ID string `json:"_id"`
						} `json:"_contentType"`
						Locale   string `json:"_locale"`
						Text     string `json:"text"`
						Children []struct {
							ID          string `json:"_id"`
							ContentType struct {
								ID string `json:"_id"`
							} `json:"_contentType"`
							Locale   string `json:"_locale"`
							Text     string `json:"text"`
							Children []struct {
								ID          string `json:"_id"`
								ContentType struct {
									ID string `json:"_id"`
								} `json:"_contentType"`
								Locale   string `json:"_locale"`
								Text     string `json:"text"`
								Type     string `json:"type,omitempty"`
								URL      string `json:"url,omitempty"`
								Slug     string `json:"slug,omitempty"`
								Children []struct {
									ID          string `json:"_id"`
									ContentType struct {
										ID string `json:"_id"`
									} `json:"_contentType"`
									Locale string `json:"_locale"`
									Text   string `json:"text"`
									URL    string `json:"url"`
									Slug   string `json:"slug,omitempty"`
								} `json:"children,omitempty"`
								ExtraFields struct {
									PosOnly bool `json:"posOnly"`
								} `json:"extraFields,omitempty"`
								Pos bool `json:"pos,omitempty"`
							} `json:"children"`
							ExtraFields struct {
								Current bool `json:"current"`
							} `json:"extraFields,omitempty"`
						} `json:"children"`
					} `json:"children"`
					Slug string `json:"slug"`
				} `json:"mobileMenu"`
				Header struct {
					ID          string `json:"_id"`
					ContentType struct {
						ID string `json:"_id"`
					} `json:"_contentType"`
					Locale   string `json:"_locale"`
					Text     string `json:"text"`
					Type     string `json:"type"`
					Children []struct {
						ID          string `json:"_id"`
						ContentType struct {
							ID string `json:"_id"`
						} `json:"_contentType"`
						Locale   string `json:"_locale"`
						Text     string `json:"text"`
						URL      string `json:"url"`
						Children []struct {
							ID          string `json:"_id"`
							ContentType struct {
								ID string `json:"_id"`
							} `json:"_contentType"`
							Locale   string `json:"_locale"`
							Text     string `json:"text"`
							Type     string `json:"type,omitempty"`
							Children []struct {
								ID          string `json:"_id"`
								ContentType struct {
									ID string `json:"_id"`
								} `json:"_contentType"`
								Locale string `json:"_locale"`
								Text   string `json:"text"`
								URL    string `json:"url"`
								Slug   string `json:"slug,omitempty"`
							} `json:"children"`
							Slug string `json:"slug,omitempty"`
							Pos  bool   `json:"pos,omitempty"`
						} `json:"children,omitempty"`
						Type string `json:"type,omitempty"`
					} `json:"children"`
					Slug string `json:"slug"`
				} `json:"header"`
				Messages struct {
					ID          string `json:"_id"`
					ContentType struct {
						ID string `json:"_id"`
					} `json:"_contentType"`
					Locale string `json:"_locale"`
					Name   string `json:"name"`
					Slug   string `json:"slug"`
					Config struct {
						AppTitle                                              string `json:"app.title"`
						Learnmore                                             string `json:"learnmore"`
						CartEmpty                                             string `json:"cart.empty"`
						AppLoading                                            string `json:"app.loading"`
						CommonClose                                           string `json:"common.close"`
						CommonLogin                                           string `json:"common.login"`
						AppCopyright                                          string `json:"app.copyright"`
						AppTitlePdp                                           string `json:"app.title.pdp"`
						CommonSignin                                          string `json:"common.signin"`
						CommonSignup                                          string `json:"common.signup"`
						ErrorGeneric                                          string `json:"error.generic"`
						HeaderSearch                                          string `json:"header.search"`
						NoMatchTitle                                          string `json:"noMatch.title"`
						FooterCompany                                         string `json:"footer.company"`
						FooterSupport                                         string `json:"footer.support"`
						FormFieldApt                                          string `json:"form.field.apt"`
						HeaderVisitUs                                         string `json:"header.visitUs"`
						CommonContinue                                        string `json:"common.continue"`
						FormFieldCity                                         string `json:"form.field.city"`
						FormFieldName                                         string `json:"form.field.name"`
						NotFoundTitle                                         string `json:"not.found.title"`
						CartBundleEmpty                                       string `json:"cart.bundleEmpty"`
						FooterBarTerms                                        string `json:"footer.bar.terms"`
						FormFieldCcCvv                                        string `json:"form.field.ccCvv"`
						FormFieldEmail                                        string `json:"form.field.email"`
						FormFieldPhone                                        string `json:"form.field.phone"`
						FormFieldState                                        string `json:"form.field.state"`
						CartItemsOnline                                       string `json:"cart.items.online"`
						DisplayFreeAmount                                     string `json:"displayFreeAmount"`
						FooterBarMejuri                                       string `json:"footer.bar.mejuri"`
						FooterJoinTitle                                       string `json:"footer.join.title"`
						FormFieldIsGift                                       string `json:"form.field.isGift"`
						ThankYouPageName                                      string `json:"thankYouPage.name"`
						CartBalanceTotal                                      string `json:"cart.balance.total"`
						CartHeaderAdvice                                      string `json:"cart.header.advice"`
						CartItemsInStore                                      string `json:"cart.items.inStore"`
						CartItemsWalkOut                                      string `json:"cart.items.walkOut"`
						CheckoutStepEdit                                      string `json:"checkout.step.edit"`
						FfCampaignDiscount                                    string `json:"ffCampaignDiscount"`
						FooterBarPrivacy                                      string `json:"footer.bar.privacy"`
						FooterSupportFaq                                      string `json:"footer.support.faq"`
						FormFieldAddress                                      string `json:"form.field.address"`
						HeaderSearchHint                                      string `json:"header.search.hint"`
						ProductModalBtn1                                      string `json:"product.modal.btn1"`
						ProductModalBtn2                                      string `json:"product.modal.btn2"`
						ProductModalTip1                                      string `json:"product.modal.tip1"`
						ProductModalTip2                                      string `json:"product.modal.tip2"`
						ProductModalTip3                                      string `json:"product.modal.tip3"`
						ProductModalTip4                                      string `json:"product.modal.tip4"`
						RelatedEditsTitle                                     string `json:"relatedEdits.title"`
						ShipmentStateJit                                      string `json:"shipment.state.jit"`
						ThankYouPageTitle                                     string `json:"thankYouPage.title"`
						FormErrorDateMax                                      string `json:"form.error.date.max"`
						FormErrorDateMin                                      string `json:"form.error.date.min"`
						FormFieldCcNumber                                     string `json:"form.field.ccNumber"`
						FormFieldLastName                                     string `json:"form.field.lastName"`
						FormFieldPassword                                     string `json:"form.field.password"`
						NoMatchDescription                                    string `json:"noMatch.description"`
						ProductModalTitle                                     string `json:"product.modal.title"`
						CartSuggestionsAdd                                    string `json:"cart.suggestions.add"`
						ChatWithStylistLink                                   string `json:"chatWithStylist.link"`
						CheckoutStepCancel                                    string `json:"checkout.step.cancel"`
						FooterJoinSubtitle                                    string `json:"footer.join.subtitle"`
						FooterLinksCompany                                    string `json:"footer.links.company"`
						FooterLinksSupport                                    string `json:"footer.links.support"`
						FooterLinksVisitUs                                    string `json:"footer.links.visitUs"`
						FormErrorArrayMax                                     string `json:"form.error.array.max"`
						FormErrorArrayMin                                     string `json:"form.error.array.min"`
						FormFieldAgentCode                                    string `json:"form.field.agentCode"`
						FormFieldCcExpYear                                    string `json:"form.field.ccExpYear"`
						FormFieldCountryID                                    string `json:"form.field.countryId"`
						FormFieldFirstName                                    string `json:"form.field.firstName"`
						FormFieldSubscribe                                    string `json:"form.field.subscribe"`
						HeaderBackCheckout                                    string `json:"header.back.checkout"`
						HeaderBackThankYou                                    string `json:"header.back.thankYou"`
						ProductReviewTitle                                    string `json:"product.review.title"`
						ProductStoryTitle1                                    string `json:"product.story.title1"`
						ProductStoryTitle2                                    string `json:"product.story.title2"`
						ProductStoryTitle3                                    string `json:"product.story.title3"`
						ProductWishlistAdd                                    string `json:"product.wishlist.add"`
						ShipmentStateStock                                    string `json:"shipment.state.stock"`
						CartActionsContinue                                   string `json:"cart.actions.continue"`
						CartBalanceShipping                                   string `json:"cart.balance.shipping"`
						CartBalanceSubtotal                                   string `json:"cart.balance.subtotal"`
						CartCouponCodeApply                                   string `json:"cart.couponCode.apply"`
						CartItemsOutOfStock                                   string `json:"cart.items.outOfStock"`
						CartProgressMessage0                                  string `json:"cart.progressMessage0"`
						CartProgressMessage1                                  string `json:"cart.progressMessage1"`
						CartProgressMessage2                                  string `json:"cart.progressMessage2"`
						CartProgressMessage3                                  string `json:"cart.progressMessage3"`
						ChatWithStylistLabel                                  string `json:"chatWithStylist.label"`
						ChatWithStylistTitle                                  string `json:"chatWithStylist.title"`
						CheckoutAddressEdit                                   string `json:"checkout.address.edit"`
						CheckoutAddressHint                                   string `json:"checkout.address.hint"`
						CheckoutPaymentCash                                   string `json:"checkout.payment.cash"`
						FormErrorNumberMax                                    string `json:"form.error.number.max"`
						FormErrorNumberMin                                    string `json:"form.error.number.min"`
						FormErrorStringMax                                    string `json:"form.error.string.max"`
						FormErrorStringMin                                    string `json:"form.error.string.min"`
						FormErrorStringURL                                    string `json:"form.error.string.url"`
						FormFieldAmountPaid                                   string `json:"form.field.amountPaid"`
						FormFieldCcExpMonth                                   string `json:"form.field.ccExpMonth"`
						FormFieldPostalCode                                   string `json:"form.field.postalCode"`
						FormFieldSalesAgent                                   string `json:"form.field.salesAgent"`
						HeaderSearchResults                                   string `json:"header.search.results"`
						NotFoundDescription                                   string `json:"not.found.description"`
						OnboardingIdleTitle                                   string `json:"onboarding.idle.title"`
						ProductSeenInTitle                                    string `json:"product.seen.in.title"`
						ProductSelectorSize                                   string `json:"product.selector.size"`
						TermsAndPrivacyTerms                                  string `json:"termsAndPrivacy.terms"`
						ThankYouPageThankYou                                  string `json:"thankYouPage.thankYou"`
						CartCouponCodeLegend                                  string `json:"cart.couponCode.legend"`
						CartSuggestionsPrice                                  string `json:"cart.suggestions.price"`
						CartSuggestionsTitle                                  string `json:"cart.suggestions.title"`
						CheckoutCompleteOrder                                 string `json:"checkout.completeOrder"`
						CheckoutDeliveryEdit                                  string `json:"checkout.delivery.edit"`
						CheckoutPaymentTitle                                  string `json:"checkout.payment.title"`
						FooterCompanyAboutUs                                  string `json:"footer.company.aboutUs"`
						FooterCompanyCareers                                  string `json:"footer.company.careers"`
						FormErrorMixedOneOf                                   string `json:"form.error.mixed.oneOf"`
						FormErrorStringTrim                                   string `json:"form.error.string.trim"`
						FormFieldGiftMessage                                  string `json:"form.field.giftMessage"`
						OnboardingLoginTitle                                  string `json:"onboarding.login.title"`
						ProductModalQuestion                                  string `json:"product.modal.question"`
						ShipmentDateLabelEta                                  string `json:"shipment.date.labelEta"`
						ShipmentDateLabelEts                                  string `json:"shipment.date.labelEts"`
						ShipmentStateDigital                                  string `json:"shipment.state.digital"`
						ShipmentStatePickUp                                   string `json:"shipment.state.pick_up"`
						ShipmentStateWalkout                                  string `json:"shipment.state.walkout"`
						CartItemsAvailability                                 string `json:"cart.items.availability"`
						CartSuggestionsLegend                                 string `json:"cart.suggestions.legend"`
						ErrorPromoTotalChanged                                string `json:"error.promoTotalChanged"`
						ExcludedFromBlackFriday                               string `json:"excludedFromBlackFriday"`
						FooterJoinBlankEmail                                  string `json:"footer.join.blank.email"`
						FooterJoinPlaceholder                                 string `json:"footer.join.placeholder"`
						FormErrorStringEmail                                  string `json:"form.error.string.email"`
						FormErrorUnableToShip                                 string `json:"form.error.unableToShip"`
						FormFieldGiftCardCode                                 string `json:"form.field.giftCardCode"`
						HeaderSearchNoResults                                 string `json:"header.search.noResults"`
						HeaderUserMenuReturns                                 string `json:"header.userMenu.returns"`
						HeaderUserMenuSignOut                                 string `json:"header.userMenu.signOut"`
						HelpLinksQualityLabel                                 string `json:"helpLinks.quality.label"`
						HelpLinksReturnsLabel                                 string `json:"helpLinks.returns.label"`
						OnboardingIdleCaption                                 string `json:"onboarding.idle.caption"`
						ProductFoursixtyTitle                                 string `json:"product.foursixty.title"`
						ProductSelectorLength                                 string `json:"product.selector.length"`
						ProductSelectorLetter                                 string `json:"product.selector.letter"`
						ProductStorySubtitle1                                 string `json:"product.story.subtitle1"`
						ProductStorySubtitle2                                 string `json:"product.story.subtitle2"`
						ProductWishlistRemove                                 string `json:"product.wishlist.remove"`
						TermsAndPrivacyMessage                                string `json:"termsAndPrivacy.message"`
						TermsAndPrivacyPrivacy                                string `json:"termsAndPrivacy.privacy"`
						FooterLinksSupportFaq                                 string `json:"footer.links.support.faq"`
						FormErrorCcCvvInvalid                                 string `json:"form.error.ccCvv.invalid"`
						FormErrorEmailInvalid                                 string `json:"form.error.email.invalid"`
						FormErrorMixedDefault                                 string `json:"form.error.mixed.default"`
						FormErrorMixedNotType                                 string `json:"form.error.mixed.notType"`
						FormErrorNameRequired                                 string `json:"form.error.name.required"`
						FormErrorStringLength                                 string `json:"form.error.string.length"`
						FormFieldPostalCodeUs                                 string `json:"form.field.postalCode.us"`
						FormFieldStreetAddress                                string `json:"form.field.streetAddress"`
						HeaderUserMenuMyOrders                                string `json:"header.userMenu.myOrders"`
						HelpLinksWarrantyLabel                                string `json:"helpLinks.warranty.label"`
						MobileMenuAccessibility                               string `json:"mobileMenu.accessibility"`
						OnboardingFacebookLogin                               string `json:"onboarding.facebookLogin"`
						OnboardingLoginCaption                                string `json:"onboarding.login.caption"`
						ProductEngravingPrompt                                string `json:"product.engraving.prompt"`
						ProductSelectorClothes                                string `json:"product.selector.clothes"`
						ThankYouPageOrderNumber                               string `json:"thankYouPage.orderNumber"`
						CheckoutAddressContinue                               string `json:"checkout.address.continue"`
						CheckoutAddressPosNotes                               string `json:"checkout.address.posNotes"`
						CheckoutPaymentSubtitle                               string `json:"checkout.payment.subtitle"`
						FooterJoinInvalidEmail                                string `json:"footer.join.invalid.email"`
						FooterSupportContactUs                                string `json:"footer.support.contact.us"`
						FooterSupportRingSizer                                string `json:"footer.support.ring.sizer"`
						FormErrorCcCvvRequired                                string `json:"form.error.ccCvv.required"`
						FormErrorCcDateInvalid                                string `json:"form.error.ccDate.invalid"`
						FormErrorEmailRequired                                string `json:"form.error.email.required"`
						FormErrorMixedNotOneOf                                string `json:"form.error.mixed.notOneOf"`
						FormErrorMixedRequired                                string `json:"form.error.mixed.required"`
						FormErrorNumberInteger                                string `json:"form.error.number.integer"`
						FormErrorStringMatches                                string `json:"form.error.string.matches"`
						HeaderSearchPlaceholder                               string `json:"header.search.placeholder"`
						HeaderUserMenuMyProfile                               string `json:"header.userMenu.myProfile"`
						HelpLinksQualityMessage                               string `json:"helpLinks.quality.message"`
						HelpLinksReturnsMessage                               string `json:"helpLinks.returns.message"`
						KlarnaMethodUnavailable                               string `json:"klarna.method.unavailable"`
						NotificationsPosNotYou                                string `json:"notifications.pos.notYou?"`
						OnboardingLoginEmailOpt                               string `json:"onboarding.login.emailOpt"`
						OnboardingRegisterTitle                               string `json:"onboarding.register.title"`
						ProductAddToCartSoldOut                               string `json:"product.addToCart.soldOut"`
						ProductCollapsableTitle                               string `json:"product.collapsable.title"`
						ThankYouPageOrderSummary                              string `json:"thankYouPage.orderSummary"`
						ThankYouPagePrintReceipt                              string `json:"thankYouPage.printReceipt"`
						ThankYouPageTitlePickup                               string `json:"thankYouPage.title.pickup"`
						CheckoutAddressFieldApt                               string `json:"checkout.address.field.apt"`
						CheckoutAddressLoginText                              string `json:"checkout.address.loginText"`
						CheckoutAddressStepTitle                              string `json:"checkout.address.stepTitle"`
						CheckoutDeliveryContinue                              string `json:"checkout.delivery.continue"`
						CheckoutDeliveryTomorrow                              string `json:"checkout.delivery.tomorrow"`
						CheckoutGiftingListTitle                              string `json:"checkout.giftingList.title"`
						CheckoutGiftingViewTitle                              string `json:"checkout.giftingView.title"`
						CheckoutPaymentStepTitle                              string `json:"checkout.payment.stepTitle"`
						FormErrorNumberLessThan                               string `json:"form.error.number.lessThan"`
						FormErrorNumberMoreThan                               string `json:"form.error.number.moreThan"`
						FormErrorNumberNegative                               string `json:"form.error.number.negative"`
						FormErrorNumberNotEqual                               string `json:"form.error.number.notEqual"`
						FormErrorNumberPositive                               string `json:"form.error.number.positive"`
						FormErrorPasswordLength                               string `json:"form.error.password.length"`
						HeaderUserMenuMyWishList                              string `json:"header.userMenu.myWishList"`
						HeaderUserMenuSpreeAdmin                              string `json:"header.userMenu.spreeAdmin"`
						HelpLinksWarrantyMessage                              string `json:"helpLinks.warranty.message"`
						ProductAddToCartAddToBag                              string `json:"product.addToCart.addToBag"`
						ProductAddToCartNeedHelp                              string `json:"product.addToCart.needHelp"`
						ProductAddToCartPreOrder                              string `json:"product.addToCart.preOrder"`
						ProductAddToCartWaitlist                              string `json:"product.addToCart.waitlist"`
						ProductFairPricingTitle                               string `json:"product.fair.pricing.title"`
						ProductGiftcardHighlight                              string `json:"product.giftcard.highlight"`
						ProductGiftcardLearnmore                              string `json:"product.giftcard.learnmore"`
						ProductModalContactLink                               string `json:"product.modal.contact.link"`
						ProductModalTipSubtitle                               string `json:"product.modal.tip.subtitle"`
						ProductPickupPopupTitle                               string `json:"product.pickup.popup.title"`
						ProductRetailBannerFree                               string `json:"product.retail.banner.free"`
						ProductRetailBannerRate                               string `json:"product.retail.banner.rate"`
						ProductSelectorSizeHelp                               string `json:"product.selector.size.Help"`
						ProductStoryLateralText                               string `json:"product.story.lateral.text"`
						ProductPageReviewMessage                              string `json:"productPage.review.message"`
						ShipmentStateBackordered                              string `json:"shipment.state.backordered"`
						ThankYouPagePaymentMethod                             string `json:"thankYouPage.paymentMethod"`
						ThankYouPagePickupAddress                             string `json:"thankYouPage.pickupAddress"`
						CartCouponCodePlaceHolder                             string `json:"cart.couponCode.placeHolder"`
						CheckoutAddressFieldCity                              string `json:"checkout.address.field.city"`
						CheckoutAddressFieldName                              string `json:"checkout.address.field.name"`
						CheckoutAddressProcessing                             string `json:"checkout.address.processing"`
						CheckoutDeliveryGiftHint                              string `json:"checkout.delivery.gift.hint"`
						CheckoutDeliveryStepTitle                             string `json:"checkout.delivery.stepTitle"`
						CheckoutOrderSummaryEmpty                             string `json:"checkout.orderSummary.empty"`
						CheckoutOverstockContinue                             string `json:"checkout.overstock.continue"`
						CheckoutOverstockNolonger                             string `json:"checkout.overstock.nolonger"`
						CheckoutPaymentCreditCard                             string `json:"checkout.payment.creditCard"`
						FormErrorCcNumberInvalid                              string `json:"form.error.ccNumber.invalid"`
						FormErrorObjectNoUnknown                              string `json:"form.error.object.noUnknown"`
						FormErrorStringLowercase                              string `json:"form.error.string.lowercase"`
						FormErrorStringUppercase                              string `json:"form.error.string.uppercase"`
						FormErrorUnableToRegister                             string `json:"form.error.unableToRegister"`
						FormFieldCreditCardSelect                             string `json:"form.field.creditCardSelect"`
						HeaderUserMenuBillingInfo                             string `json:"header.userMenu.billingInfo"`
						HeaderUserMenuStoreCredit                             string `json:"header.userMenu.storeCredit"`
						OnboardingRegisterCaption                             string `json:"onboarding.register.caption"`
						ProductFairPricingButton                              string `json:"product.fair.pricing.button"`
						ProductFairPricingMejuri                              string `json:"product.fair.pricing.mejuri"`
						ProductFairPricingRetail                              string `json:"product.fair.pricing.retail"`
						ProductNotificationShipOn                             string `json:"product.notification.shipOn"`
						ProductPickupPopupText1                               string `json:"product.pickup.popup.text.1"`
						ProductPickupPopupText2                               string `json:"product.pickup.popup.text.2"`
						ProductPickupPopupText3                               string `json:"product.pickup.popup.text.3"`
						CartCouponCodeDoesNotExist                            string `json:"cart.couponCode.doesNotExist"`
						CheckoutAddressFieldEmail                             string `json:"checkout.address.field.email"`
						CheckoutAddressFieldPhone                             string `json:"checkout.address.field.phone"`
						CheckoutAddressFieldState                             string `json:"checkout.address.field.state"`
						CheckoutDeliveryGiftTitle                             string `json:"checkout.delivery.gift.title"`
						CheckoutDeliveryProcessing                            string `json:"checkout.delivery.processing"`
						CheckoutPaymentBillingInfo                            string `json:"checkout.payment.billingInfo"`
						FooterLinksCompanyAboutUs                             string `json:"footer.links.company.aboutUs"`
						FooterLinksCompanyCareers                             string `json:"footer.links.company.careers"`
						FooterLinksSupportCareers                             string `json:"footer.links.support.careers"`
						FooterLinksVisitUsNewYork                             string `json:"footer.links.visitUs.newYork"`
						FooterLinksVisitUsToronto                             string `json:"footer.links.visitUs.toronto"`
						FooterSupportAccessibility                            string `json:"footer.support.accessibility"`
						FooterSupportVideoStyling                             string `json:"footer.support.video.styling"`
						FormErrorCcNumberRequired                             string `json:"form.error.ccNumber.required"`
						FormErrorGiftcardRequired                             string `json:"form.error.giftcard.required"`
						FormErrorPasswordRequired                             string `json:"form.error.password.required"`
						HeaderUserMenuShippingInfo                            string `json:"header.userMenu.shippingInfo"`
						HelpLinksReturnsLabelFree                             string `json:"helpLinks.returns.label.free"`
						NotificationsPosLogoutText                            string `json:"notifications.pos.logoutText"`
						ProductErrorNoVariantSize                             string `json:"product.error.noVariant.size"`
						ProductFairPricingRetails                             string `json:"product.fair.pricing.retails"`
						ProductGiftcardDetailsTyc                             string `json:"product.giftcard.details.tyc"`
						ProductNotificationInStock                            string `json:"product.notification.inStock"`
						ProductNotificationSoldOut                            string `json:"product.notification.soldOut"`
						ThankYouPageShippingAddress                           string `json:"thankYouPage.shippingAddress"`
						CheckoutAddressFieldIsGift                            string `json:"checkout.address.field.isGift"`
						CheckoutAddressSectionInfo                            string `json:"checkout.address.section.info"`
						CheckoutCashBalanceDueLabel                           string `json:"checkout.cash.balanceDueLabel"`
						CheckoutDeliveryGiftBanner                            string `json:"checkout.delivery.gift.banner"`
						FooterSupportMaterialsCare                            string `json:"footer.support.materials.care"`
						FormErrorCcExpYearRequired                            string `json:"form.error.ccExpYear.required"`
						FormErrorCreditCardDeclined                           string `json:"form.error.creditCardDeclined"`
						HeaderNotificationsDiscount                           string `json:"header.notifications.discount"`
						ProductGiftcardBannerTitle                            string `json:"product.giftcard.banner.title"`
						ProductModalContactMessage                            string `json:"product.modal.contact.message"`
						ProductNotificationPickup1                            string `json:"product.notification.pickup.1"`
						ProductNotificationPickup2                            string `json:"product.notification.pickup.2"`
						ProductNotificationPickup3                            string `json:"product.notification.pickup.3"`
						ProductNotificationPickup4                            string `json:"product.notification.pickup.4"`
						ProductNotificationPickup5                            string `json:"product.notification.pickup.5"`
						ProductPickupPopupTextFaq                             string `json:"product.pickup.popup.text.faq"`
						ThankYouPageContinueShopping                          string `json:"thankYouPage.continueShopping"`
						CheckoutAddressFieldAddress                           string `json:"checkout.address.field.address"`
						CheckoutAddressFieldCountry                           string `json:"checkout.address.field.country"`
						CheckoutOrderSummarySubtotal                          string `json:"checkout.orderSummary.subtotal"`
						FooterLinksSupportContactUs                           string `json:"footer.links.support.contactUs"`
						FooterLinksSupportMaterials                           string `json:"footer.links.support.materials"`
						FooterLinksSupportRingSizer                           string `json:"footer.links.support.ringSizer"`
						FormErrorAmountPaidRequired                           string `json:"form.error.amountPaid.required"`
						FormErrorCcExpMonthRequired                           string `json:"form.error.ccExpMonth.required"`
						FormFieldNewsletterSubscribe                          string `json:"form.field.newsletterSubscribe"`
						HelpLinksReturnsMessageFree                           string `json:"helpLinks.returns.message.free"`
						NotificationsPosBuyerMessage                          string `json:"notifications.pos.buyerMessage"`
						OnboardingResetPasswordTitle                          string `json:"onboarding.resetPassword.title"`
						ProductEngravingLengthError                           string `json:"product.engraving.length.error"`
						ProductErrorNoVariantLetter                           string `json:"product.error.noVariant.letter"`
						ProductGiftcardDetailsTitle                           string `json:"product.giftcard.details.title"`
						ProductModalContactSubtitle                           string `json:"product.modal.contact.subtitle"`
						ProductModelsAreWearingTitle                          string `json:"product.modelsAreWearing.title"`
						ProductRetailBannerWarranty                           string `json:"product.retail.banner.warranty"`
						CartHeaderFreeShippingReached                         string `json:"cart.header.freeShippingReached"`
						CheckoutAddressFieldFullName                          string `json:"checkout.address.field.fullName"`
						CheckoutAddressFieldPassword                          string `json:"checkout.address.field.password"`
						CheckoutAddressSectionPickup                          string `json:"checkout.address.section.pickup"`
						CheckoutDeliveryGiftSubtitle                          string `json:"checkout.delivery.gift.subtitle"`
						CheckoutDeliveryMethodPickUp                          string `json:"checkout.delivery.method.pickUp"`
						CheckoutOrderSummaryTaxLabel                          string `json:"checkout.orderSummary.tax.label"`
						CheckoutOverstockCartUpdated                          string `json:"checkout.overstock.cart-updated"`
						CheckoutSalesAgentBannerGuest                         string `json:"checkout.salesAgentBanner.guest"`
						CheckoutSalesAgentBannerTitle                         string `json:"checkout.salesAgentBanner.title"`
						FooterLinksVisitUsLosAngeles                          string `json:"footer.links.visitUs.losAngeles"`
						FooterSupportShippingReturns                          string `json:"footer.support.shipping.returns"`
						HelpLinksReturnsLabelHoliday                          string `json:"helpLinks.returns.label.holiday"`
						HelpLinksReturnsLabelNoLocal                          string `json:"helpLinks.returns.label.noLocal"`
						OnboardingLoginForgotPassword                         string `json:"onboarding.login.forgotPassword"`
						ProductEngravingDefaultError                          string `json:"product.engraving.default.error"`
						ProductEngravingInitialError                          string `json:"product.engraving.initial.error"`
						ProductErrorNoVariantClothes                          string `json:"product.error.noVariant.clothes"`
						ProductFairPricingPopupTile                           string `json:"product.fair.pricing.popup.tile"`
						ProductSelectorNecklaceLength                         string `json:"product.selector.necklaceLength"`
						ProductPageDetailsTaxIncluded                         string `json:"productPage.details.taxIncluded"`
						CheckoutAddressFieldSubscribe                         string `json:"checkout.address.field.subscribe"`
						CheckoutAddressSectionAddress                         string `json:"checkout.address.section.address"`
						CheckoutDeliveryGiftFromHint                          string `json:"checkout.delivery.gift.from.hint"`
						CheckoutDeliveryMethodWalkout                         string `json:"checkout.delivery.method.walkout"`
						CheckoutPaymentLoadingMessage                         string `json:"checkout.payment.loading.message"`
						FooterLinksSupportStylingHelp                         string `json:"footer.links.support.stylingHelp"`
						FooterNewsletterSubscribeField                        string `json:"footer.newsletterSubscribe.field"`
						FooterNewsletterSubscribeTitle                        string `json:"footer.newsletterSubscribe.title"`
						FormFieldSameAddressForBilling                        string `json:"form.field.sameAddressForBilling"`
						OnboardingResetPasswordCaption                        string `json:"onboarding.resetPassword.caption"`
						ProductEngravingMonogramError                         string `json:"product.engraving.monogram.error"`
						ProductGiftcardBannerSubtitle                         string `json:"product.giftcard.banner.subtitle"`
						ProductNotificationMadeToOrder                        string `json:"product.notification.madeToOrder"`
						ProductNotificationOnlyFewLeft                        string `json:"product.notification.onlyFewLeft"`
						ProductRecommendedGalleryTitle                        string `json:"product.recommendedGallery.title"`
						ProductRetailBannerLearnMore                          string `json:"product.retail.banner.learn.more"`
						ThankYouPageConfirmationMessage                       string `json:"thankYouPage.confirmationMessage"`
						CheckoutDeliveryShippingDateEta                       string `json:"checkout.delivery.shippingDateEta"`
						CheckoutDeliveryShippingDateEts                       string `json:"checkout.delivery.shippingDateEts"`
						CheckoutOrderSummaryTotalLabel                        string `json:"checkout.orderSummary.total.label"`
						CheckoutOverstockPleaseContact                        string `json:"checkout.overstock.please-contact"`
						CheckoutPaymentContinueOnPaypal                       string `json:"checkout.payment.continueOnPaypal"`
						CheckoutPaymentCreditCardError                        string `json:"checkout.payment.creditCard.error"`
						CheckoutPaymentGiftCardApplied                        string `json:"checkout.payment.giftCard.applied"`
						CheckoutSalesAgentSelectorTitle                       string `json:"checkout.salesAgentSelector.title"`
						FormErrorAmountPaidInsuficient                        string `json:"form.error.amountPaid.insuficient"`
						HelpLinksReturnsMessageHoliday                        string `json:"helpLinks.returns.message.holiday"`
						HelpLinksReturnsMessageNoLocal                        string `json:"helpLinks.returns.message.noLocal"`
						ProductEngravingMonogramPrompt                        string `json:"product.engraving.monogram.prompt"`
						ProductGiftcardSubtitleDigital                        string `json:"product.giftcard.subtitle.digital"`
						CheckoutAddressFieldGiftMessage                       string `json:"checkout.address.field.giftMessage"`
						CheckoutDeliveryGiftPlaceHolder                       string `json:"checkout.delivery.gift.placeHolder"`
						CheckoutOrderSummaryDutiesLocal                       string `json:"checkout.orderSummary.duties.local"`
						FooterNewsletterSubscribeWelcome                      string `json:"footer.newsletterSubscribe.welcome"`
						HeaderNotificationsDiscountLink                       string `json:"header.notifications.discount.link"`
						OnboardingResetPasswordDoneTitle                      string `json:"onboarding.resetPasswordDone.title"`
						ProductEngravingOneLengthError                        string `json:"product.engraving.one.length.error"`
						ProductGiftcardSubtitlePhysical                       string `json:"product.giftcard.subtitle.physical"`
						ProductMaterialDescriptionTitle                       string `json:"product.material_description.title"`
						ProductNotificationBackorderable                      string `json:"product.notification.backorderable"`
						ProductPickupPopupTextFaqLink                         string `json:"product.pickup.popup.text.faq.link"`
						CheckoutAddressPreviewGiftLabel                       string `json:"checkout.address.preview.gift.label"`
						CheckoutAddressSubsectionAddress                      string `json:"checkout.address.subsection.address"`
						CheckoutSalesAgentSelectorMessage                     string `json:"checkout.salesAgentSelector.message"`
						FooterLinksVisitUsPiercingStudio                      string `json:"footer.links.visitUs.piercingStudio"`
						FooterNewsletterSubscribeSubtitle                     string `json:"footer.newsletterSubscribe.subtitle"`
						NotificationsPosGuestBuyerMessage                     string `json:"notifications.pos.guestBuyerMessage"`
						CheckoutAddressFieldPostalCodeUs                      string `json:"checkout.address.field.postalCode.us"`
						CheckoutPaymentCreditCardEditCard                     string `json:"checkout.payment.creditCard.editCard"`
						CheckoutPaymentGiftCardLinkLabel                      string `json:"checkout.payment.giftCard.link.label"`
						CheckoutSalesAgentSelectorContinue                    string `json:"checkout.salesAgentSelector.continue"`
						HeaderSearchSuggestedProductsTitle                    string `json:"header.search.suggestedProductsTitle"`
						OnboardingResetPasswordDoneCaption                    string `json:"onboarding.resetPasswordDone.caption"`
						ProductEngravingLetterPlaceholder                     string `json:"product.engraving.letter.placeholder"`
						ProductModalRingMeasurementStep1                      string `json:"product.modal.ring.measurement.step1"`
						ProductModalRingMeasurementStep2                      string `json:"product.modal.ring.measurement.step2"`
						ProductModalRingMeasurementStep3                      string `json:"product.modal.ring.measurement.step3"`
						ProductSelectorNecklaceLengthHelp                     string `json:"product.selector.necklaceLength.help"`
						ProductPageReviewMessageNoReviews                     string `json:"productPage.review.message.noReviews"`
						CheckoutPaymentKlarnaSessionExpired                   string `json:"checkout.payment.klarnaSessionExpired"`
						ProductEngravingDefaultPlaceholder                    string `json:"product.engraving.default.placeholder"`
						ProductEngravingGeneralPlaceholder                    string `json:"product.engraving.general.placeholder"`
						ProductNotificationInStockWithDate                    string `json:"product.notification.inStock.withDate"`
						ProductNotificationPickup1NoStock                     string `json:"product.notification.pickup.1.noStock"`
						ProductRetailBannerHolidayReturns                     string `json:"product.retail.banner.holiday.returns"`
						CheckoutAddressFieldPostalCodeElse                    string `json:"checkout.address.field.postalCode.else"`
						CheckoutOrderSummaryShippingDefault                   string `json:"checkout.orderSummary.shipping.default"`
						CheckoutOrderSummaryShippingExpress                   string `json:"checkout.orderSummary.shipping.express"`
						CheckoutOrderSummaryShippingWalkout                   string `json:"checkout.orderSummary.shipping.walkout"`
						CheckoutPaymentCreditCardAddNewCard                   string `json:"checkout.payment.creditCard.addNewCard"`
						CheckoutPaymentGiftCardButtonLabel                    string `json:"checkout.payment.giftCard.button.label"`
						CheckoutPaymentKlarna4Installments                    string `json:"checkout.payment.klarna-4-installments"`
						HeaderSearchRecommendedProductsTitle                  string `json:"header.search.recommendedProductsTitle"`
						ProductAddToCartWaitlistPopupEmail                    string `json:"product.addToCart.waitlist.popup.email"`
						ProductAddToCartWaitlistPopupPhone                    string `json:"product.addToCart.waitlist.popup.phone"`
						ProductAddToCartWaitlistPopupTitle                    string `json:"product.addToCart.waitlist.popup.title"`
						ProductEngravingMonogramPlaceholder                   string `json:"product.engraving.monogram.placeholder"`
						ProductErrorNoVariantNecklaceLength                   string `json:"product.error.noVariant.necklaceLength"`
						ProductFairPricingPopupDescription                    string `json:"product.fair.pricing.popup.description"`
						ProductModalFingerMeasurementStep1                    string `json:"product.modal.finger.measurement.step1"`
						ProductModalFingerMeasurementStep2                    string `json:"product.modal.finger.measurement.step2"`
						ProductModalFingerMeasurementStep3                    string `json:"product.modal.finger.measurement.step3"`
						ProductModalFingerMeasurementStep4                    string `json:"product.modal.finger.measurement.step4"`
						ProductModalRingMeasurementHeader1                    string `json:"product.modal.ring.measurement.header1"`
						ProductModalRingMeasurementHeader2                    string `json:"product.modal.ring.measurement.header2"`
						ProductModalRingMeasurementHeader3                    string `json:"product.modal.ring.measurement.header3"`
						CheckoutAddressSubsectionPickupNote                   string `json:"checkout.address.subsection.pickup.note"`
						CheckoutDeliveryGiftFromPlaceHolder                   string `json:"checkout.delivery.gift.from.placeHolder"`
						CheckoutDeliveryMethodExpressPlural                   string `json:"checkout.delivery.method.express.plural"`
						CheckoutDeliveryShippingDateSplitted                  string `json:"checkout.delivery.shippingDate.splitted"`
						CheckoutOrderSummaryHeaderViewLabel                   string `json:"checkout.orderSummary.header.view.label"`
						CheckoutOrderSummarySubtotalWithTaxes                 string `json:"checkout.orderSummary.subtotalWithTaxes"`
						CheckoutSalesAgentBannerClickToChange                 string `json:"checkout.salesAgentBanner.clickToChange"`
						FooterLinksSupportShippingAndReturns                  string `json:"footer.links.support.shippingAndReturns"`
						ProductAddToCartWaitlistPopupButton                   string `json:"product.addToCart.waitlist.popup.button"`
						ProductNotificationAvailableForPickup                 string `json:"product.notification.availableForPickup"`
						ProductRetailBannerReturnsDaysFree                    string `json:"product.retail.banner.returns.days.free"`
						ThankYouPageConfirmationMessagePickup                 string `json:"thankYouPage.confirmationMessage.pickup"`
						CheckoutAddressFieldGiftRecipientName                 string `json:"checkout.address.field.giftRecipientName"`
						CheckoutAddressSubsectionPickupHours                  string `json:"checkout.address.subsection.pickup.hours"`
						CheckoutAddressSubsectionTitlePickup                  string `json:"checkout.address.subsection.title.pickup"`
						CheckoutDeliveryGiftEmojiNotSupported                 string `json:"checkout.delivery.gift.emojiNotSupported"`
						CheckoutOrderSummaryHeaderTotalLabel                  string `json:"checkout.orderSummary.header.total.label"`
						CheckoutOrderSummaryShippingEstimated                 string `json:"checkout.orderSummary.shipping.estimated"`
						CheckoutOrderSummaryShippingExpedited                 string `json:"checkout.orderSummary.shipping.expedited"`
						ProductModalFingerMeasurementHeader1                  string `json:"product.modal.finger.measurement.header1"`
						ProductModalFingerMeasurementHeader2                  string `json:"product.modal.finger.measurement.header2"`
						ProductNotificationAvailableForWalkout                string `json:"product.notification.availableForWalkout"`
						ProductRetailBannerReturnsDaysLocal                   string `json:"product.retail.banner.returns.days.local"`
						CheckoutDeliveryMethodExpeditedPlural                 string `json:"checkout.delivery.method.expedited.plural"`
						CheckoutDeliveryMethodExpressSingular                 string `json:"checkout.delivery.method.express.singular"`
						FormFieldSaveUserProfileShippingAddress               string `json:"form.field.saveUserProfileShippingAddress"`
						ProductAddToCartWaitlistPopupSmsTitle                 string `json:"product.addToCart.waitlist.popup.smsTitle"`
						ProductNotificationOnlyFewLeftWithDate                string `json:"product.notification.onlyFewLeft.withDate"`
						CheckoutAddressFieldHasSpecialPackaging               string `json:"checkout.address.field.hasSpecialPackaging"`
						CheckoutAddressFieldSpecialPackageBoxes               string `json:"checkout.address.field.specialPackageBoxes"`
						CheckoutAddressSubsectionTitleShipping                string `json:"checkout.address.subsection.title.shipping"`
						CheckoutDeliveryPreviewGiftLabelPlural                string `json:"checkout.delivery.preview.giftLabel.plural"`
						CheckoutOrderSummaryDutiesInternational               string `json:"checkout.orderSummary.duties.international"`
						CheckoutOrderSummaryEstimatedTotalLabel               string `json:"checkout.orderSummary.estimatedTotal.label"`
						CheckoutOrderSummarySubtotalWithTaxesUk               string `json:"checkout.orderSummary.subtotalWithTaxes.uk"`
						CheckoutPaymentKlarnaWrongCreditCardData              string `json:"checkout.payment.klarnaWrongCreditCardData"`
						ProductAddToCartWaitlistPopupCountryID                string `json:"product.addToCart.waitlist.popup.countryId"`
						ProductAddToCartWaitlistPopupSmsButton                string `json:"product.addToCart.waitlist.popup.smsButton"`
						CheckoutAddressSubsectionSubtitlePickup               string `json:"checkout.address.subsection.subtitle.pickup"`
						CheckoutDeliveryMethodExpeditedSingular               string `json:"checkout.delivery.method.expedited.singular"`
						ProductAddToCartWaitlistPopupDisclaimer               string `json:"product.addToCart.waitlist.popup.disclaimer"`
						CheckoutAddressFieldSpecialPackageMessage             string `json:"checkout.address.field.specialPackageMessage"`
						CheckoutDeliveryPreviewGiftLabelSingular              string `json:"checkout.delivery.preview.giftLabel.singular"`
						ProductAddToCartDigitalGiftCardPopupName              string `json:"product.addToCart.digitalGiftCard.popup.name"`
						ProductAddToCartWaitlistPopupNoSmsButton              string `json:"product.addToCart.waitlist.popup.noSmsButton"`
						ProductAddToCartWaitlistPopupSmsSubTitle              string `json:"product.addToCart.waitlist.popup.smsSubTitle"`
						ProductAddToCartDigitalGiftCardPopupEmail             string `json:"product.addToCart.digitalGiftCard.popup.email"`
						ProductAddToCartDigitalGiftCardPopupTitle             string `json:"product.addToCart.digitalGiftCard.popup.title"`
						ProductAddToCartPhysicalGiftCardPopupTitle            string `json:"product.addToCart.physicalGiftCard.popup.title"`
						CheckoutAddressPreviewSpecialPackagingLabel           string `json:"checkout.address.preview.specialPackaging.label"`
						CheckoutOrderSummaryHeaderItemsLabelPlural            string `json:"checkout.orderSummary.header.items.label.plural"`
						ProductAddToCartDigitalGiftCardPopupMessage           string `json:"product.addToCart.digitalGiftCard.popup.message"`
						CheckoutOrderSummaryDutiesInternationalLabel          string `json:"checkout.orderSummary.duties.international.label"`
						ProductAddToCartDigitalGiftCardPopupSubtitle          string `json:"product.addToCart.digitalGiftCard.popup.subtitle"`
						ProductAddToCartPhysicalGiftCardPopupMessage          string `json:"product.addToCart.physicalGiftCard.popup.message"`
						ProductAddToCartWaitlistPopupSmsDisclaimer1           string `json:"product.addToCart.waitlist.popup.smsDisclaimer.1"`
						ProductAddToCartWaitlistPopupSmsDisclaimer2           string `json:"product.addToCart.waitlist.popup.smsDisclaimer.2"`
						ProductAddToCartWaitlistPopupSmsDisclaimer3           string `json:"product.addToCart.waitlist.popup.smsDisclaimer.3"`
						CheckoutOrderSummaryHeaderItemsLabelSingular          string `json:"checkout.orderSummary.header.items.label.singular"`
						ProductAddToCartPhysicalGiftCardPopupSubtitle         string `json:"product.addToCart.physicalGiftCard.popup.subtitle"`
						ProductAddToCartWaitlistPopupDisclaimerTerms          string `json:"product.addToCart.waitlist.popup.disclaimer.terms"`
						ProductAddToCartWaitlistPopupDisclaimerPrivacy        string `json:"product.addToCart.waitlist.popup.disclaimer.privacy"`
						CheckoutAddressSubsectionTitlePickupNotAvailable      string `json:"checkout.address.subsection.title.pickup.notAvailable"`
						CheckoutAddressSubsectionTitleShippingNotAvailable    string `json:"checkout.address.subsection.title.shipping.notAvailable"`
						CheckoutAddressSubsectionSubtitlePickupNotAvailable   string `json:"checkout.address.subsection.subtitle.pickup.notAvailable"`
						CheckoutAddressSubsectionSubtitleShippingNotAvailable string `json:"checkout.address.subsection.subtitle.shipping.notAvailable"`
						ThankYouPageOnboardingIdleTitle                       string `json:"thankYouPage.onboarding.idle.title"`
						ThankYouPageOnboardingIdleCaption                     string `json:"thankYouPage.onboarding.idle.caption"`
						ThankYouPageOnboardingLoginTitle                      string `json:"thankYouPage.onboarding.login.title"`
						ThankYouPageOnboardingLoginCaption                    string `json:"thankYouPage.onboarding.login.caption"`
						ThankYouPageOnboardingRegisterTitle                   string `json:"thankYouPage.onboarding.register.title"`
						ThankYouPageOnboardingRegisterCaption                 string `json:"thankYouPage.onboarding.register.caption"`
						ThankYouPageOnboardingResetPasswordTitle              string `json:"thankYouPage.onboarding.resetPassword.title"`
						ThankYouPageOnboardingResetPasswordCaption            string `json:"thankYouPage.onboarding.resetPassword.caption"`
						ThankYouPageOnboardingLoginForgotPassword             string `json:"thankYouPage.onboarding.login.forgotPassword"`
						ThankYouPageOnboardingResetPasswordDoneTitle          string `json:"thankYouPage.onboarding.resetPasswordDone.title"`
						ThankYouPageOnboardingResetPasswordDoneCaption        string `json:"thankYouPage.onboarding.resetPasswordDone.caption"`
					} `json:"config"`
				} `json:"messages"`
				LocalizedCommon struct {
					ID          string `json:"_id"`
					ContentType struct {
						ID string `json:"_id"`
					} `json:"_contentType"`
					Locale         string `json:"_locale"`
					Name           string `json:"name"`
					Slug           string `json:"slug"`
					LocalizedItems []struct {
						ID          string `json:"_id"`
						ContentType struct {
							ID string `json:"_id"`
						} `json:"_contentType"`
						Locale  string `json:"_locale"`
						Key     string `json:"key"`
						Content string `json:"content"`
					} `json:"localizedItems"`
				} `json:"localizedCommon"`
				Notifications struct {
					ID          string `json:"_id"`
					ContentType struct {
						ID string `json:"_id"`
					} `json:"_contentType"`
					Locale    string `json:"_locale"`
					Name      string `json:"name"`
					Slug      string `json:"slug"`
					Component struct {
						Data struct {
							Components []struct {
								Data struct {
									Components []struct {
										Data struct {
											Components []struct {
												Data struct {
													Components []struct {
														Data struct {
															ID          string `json:"_id"`
															ContentType struct {
																ID string `json:"_id"`
															} `json:"_contentType"`
															Locale  string `json:"_locale"`
															Name    string `json:"name"`
															Content struct {
																NodeType string `json:"nodeType"`
																Data     struct {
																} `json:"data"`
																Content []struct {
																	NodeType string `json:"nodeType"`
																	Content  []struct {
																		NodeType string        `json:"nodeType"`
																		Value    string        `json:"value,omitempty"`
																		Marks    []interface{} `json:"marks,omitempty"`
																		Content  []struct {
																			NodeType string `json:"nodeType"`
																			Value    string `json:"value"`
																			Marks    []struct {
																				Type string `json:"type"`
																			} `json:"marks"`
																			Data struct {
																			} `json:"data"`
																		} `json:"content,omitempty"`
																		Data struct {
																			URI string `json:"uri"`
																		} `json:"data,omitempty"`
																	} `json:"content"`
																	Data struct {
																	} `json:"data"`
																} `json:"content"`
															} `json:"content"`
														} `json:"data"`
														Name       string        `json:"name"`
														Type       string        `json:"type"`
														HTMLID     string        `json:"htmlId"`
														Labels     []interface{} `json:"labels"`
														Contexts   []interface{} `json:"contexts"`
														ClassName  string        `json:"className"`
														Behaviours []struct {
															Name                string `json:"name"`
															ConfigsByBreakpoint []struct {
																Enabled bool `json:"enabled"`
																Options struct {
																} `json:"options"`
																Breakpoint struct {
																	MinWidth    int    `json:"minWidth"`
																	Orientation string `json:"orientation"`
																} `json:"breakpoint"`
															} `json:"configsByBreakpoint"`
														} `json:"behaviours"`
														ReferenceID     string `json:"referenceId"`
														TrackingName    string `json:"trackingName"`
														CSSByBreakpoint []struct {
															CSS []struct {
																Name  string `json:"name"`
																Value string `json:"value"`
															} `json:"css"`
															CSSHover   []interface{} `json:"cssHover"`
															IsActive   bool          `json:"isActive"`
															CSSActive  []interface{} `json:"cssActive"`
															Breakpoint struct {
																MinWidth    int    `json:"minWidth"`
																Orientation string `json:"orientation"`
															} `json:"breakpoint"`
															IsDisabled  bool          `json:"isDisabled"`
															CSSDisabled []interface{} `json:"cssDisabled"`
														} `json:"cssByBreakpoint"`
													} `json:"components"`
												} `json:"data"`
												Name       string        `json:"name"`
												Type       string        `json:"type"`
												HTMLID     string        `json:"htmlId"`
												Labels     []interface{} `json:"labels"`
												Contexts   []interface{} `json:"contexts"`
												ClassName  string        `json:"className"`
												Behaviours []struct {
													Name                string `json:"name"`
													ConfigsByBreakpoint []struct {
														Enabled bool `json:"enabled"`
														Options struct {
															PageGap                      int    `json:"pageGap"`
															Autoplay                     bool   `json:"autoplay"`
															Infinite                     bool   `json:"infinite"`
															ShowArrows                   bool   `json:"showArrows"`
															ItemsPerPage                 string `json:"itemsPerPage"`
															ShowOverflow                 bool   `json:"showOverflow"`
															AutoplaySpeed                int    `json:"autoplaySpeed"`
															ItemsToScroll                string `json:"itemsToScroll"`
															VariableWidth                bool   `json:"variableWidth"`
															AdaptiveHeight               bool   `json:"adaptiveHeight"`
															ShowPagination               bool   `json:"showPagination"`
															TransitionType               string `json:"transitionType"`
															PaginationStyle              string `json:"paginationStyle"`
															PaginationOrientation        string `json:"paginationOrientation"`
															PaginationPositionVertical   string `json:"paginationPositionVertical"`
															PaginationPositionHorizontal string `json:"paginationPositionHorizontal"`
														} `json:"options"`
														Breakpoint struct {
															MinWidth    int    `json:"minWidth"`
															Orientation string `json:"orientation"`
														} `json:"breakpoint"`
													} `json:"configsByBreakpoint"`
												} `json:"behaviours"`
												ReferenceID     string `json:"referenceId"`
												TrackingName    string `json:"trackingName"`
												CSSByBreakpoint []struct {
													CSS []struct {
														Name  string `json:"name"`
														Value string `json:"value"`
													} `json:"css"`
													CSSHover   []interface{} `json:"cssHover"`
													IsActive   bool          `json:"isActive"`
													CSSActive  []interface{} `json:"cssActive"`
													Breakpoint struct {
														MinWidth    int    `json:"minWidth"`
														Orientation string `json:"orientation"`
													} `json:"breakpoint"`
													IsDisabled  bool          `json:"isDisabled"`
													CSSDisabled []interface{} `json:"cssDisabled"`
												} `json:"cssByBreakpoint"`
											} `json:"components"`
										} `json:"data"`
										Name            string        `json:"name"`
										Type            string        `json:"type"`
										HTMLID          string        `json:"htmlId"`
										Labels          []interface{} `json:"labels"`
										Contexts        []interface{} `json:"contexts"`
										ClassName       string        `json:"className"`
										Behaviours      []interface{} `json:"behaviours"`
										ReferenceID     string        `json:"referenceId"`
										TrackingName    string        `json:"trackingName"`
										CSSByBreakpoint []struct {
											CSS []struct {
												Name  string `json:"name"`
												Value string `json:"value"`
											} `json:"css"`
											CSSHover   []interface{} `json:"cssHover"`
											IsActive   bool          `json:"isActive"`
											CSSActive  []interface{} `json:"cssActive"`
											Breakpoint struct {
												MinWidth    int    `json:"minWidth"`
												Orientation string `json:"orientation"`
											} `json:"breakpoint"`
											IsDisabled  bool          `json:"isDisabled"`
											CSSDisabled []interface{} `json:"cssDisabled"`
										} `json:"cssByBreakpoint"`
									} `json:"components"`
								} `json:"data"`
								Name            string        `json:"name"`
								Type            string        `json:"type"`
								HTMLID          string        `json:"htmlId"`
								Labels          []interface{} `json:"labels"`
								Contexts        []interface{} `json:"contexts"`
								ClassName       string        `json:"className"`
								Behaviours      []interface{} `json:"behaviours"`
								ReferenceID     string        `json:"referenceId"`
								TrackingName    string        `json:"trackingName"`
								CSSByBreakpoint []struct {
									CSS        []interface{} `json:"css"`
									CSSHover   []interface{} `json:"cssHover"`
									IsActive   bool          `json:"isActive"`
									CSSActive  []interface{} `json:"cssActive"`
									Breakpoint struct {
										MinWidth    int    `json:"minWidth"`
										Orientation string `json:"orientation"`
									} `json:"breakpoint"`
									IsDisabled  bool          `json:"isDisabled"`
									CSSDisabled []interface{} `json:"cssDisabled"`
								} `json:"cssByBreakpoint"`
							} `json:"components"`
						} `json:"data"`
						Name            string        `json:"name"`
						Type            string        `json:"type"`
						HTMLID          string        `json:"htmlId"`
						Labels          []interface{} `json:"labels"`
						Contexts        []interface{} `json:"contexts"`
						ClassName       string        `json:"className"`
						Behaviours      []interface{} `json:"behaviours"`
						ReferenceID     string        `json:"referenceId"`
						TrackingName    string        `json:"trackingName"`
						CSSByBreakpoint []struct {
							CSS        []interface{} `json:"css"`
							CSSHover   []interface{} `json:"cssHover"`
							IsActive   bool          `json:"isActive"`
							CSSActive  []interface{} `json:"cssActive"`
							Breakpoint struct {
								MinWidth    int    `json:"minWidth"`
								Orientation string `json:"orientation"`
							} `json:"breakpoint"`
							IsDisabled  bool          `json:"isDisabled"`
							CSSDisabled []interface{} `json:"cssDisabled"`
						} `json:"cssByBreakpoint"`
					} `json:"component"`
				} `json:"notifications"`
				Accessibility struct {
					ID          string `json:"_id"`
					ContentType struct {
						ID string `json:"_id"`
					} `json:"_contentType"`
					Locale    string `json:"_locale"`
					Name      string `json:"name"`
					Slug      string `json:"slug"`
					Component struct {
						Name            string `json:"name"`
						TrackingName    string `json:"trackingName"`
						CSSByBreakpoint []struct {
							Breakpoint struct {
								Orientation string `json:"orientation"`
								MinWidth    int    `json:"minWidth"`
							} `json:"breakpoint"`
							CSS []struct {
								Name  string `json:"name"`
								Value string `json:"value"`
							} `json:"css"`
							CSSHover    []interface{} `json:"cssHover"`
							CSSActive   []interface{} `json:"cssActive"`
							CSSDisabled []interface{} `json:"cssDisabled"`
						} `json:"cssByBreakpoint"`
						Behaviours []interface{} `json:"behaviours"`
						Contexts   []interface{} `json:"contexts"`
						Type       string        `json:"type"`
						Data       struct {
							Components []struct {
								Name            string `json:"name"`
								TrackingName    string `json:"trackingName"`
								CSSByBreakpoint []struct {
									Breakpoint struct {
										Orientation string `json:"orientation"`
										MinWidth    int    `json:"minWidth"`
									} `json:"breakpoint"`
									CSS         []interface{} `json:"css"`
									CSSHover    []interface{} `json:"cssHover"`
									CSSActive   []interface{} `json:"cssActive"`
									CSSDisabled []interface{} `json:"cssDisabled"`
								} `json:"cssByBreakpoint"`
								Behaviours []struct {
									Name                string `json:"name"`
									ConfigsByBreakpoint []struct {
										Breakpoint struct {
											Orientation string `json:"orientation"`
											MinWidth    int    `json:"minWidth"`
										} `json:"breakpoint"`
										Enabled bool `json:"enabled"`
										Options struct {
											URL string `json:"url"`
										} `json:"options"`
									} `json:"configsByBreakpoint"`
								} `json:"behaviours"`
								Contexts []interface{} `json:"contexts"`
								Type     string        `json:"type"`
								Data     struct {
									ID          string `json:"_id"`
									ContentType struct {
										ID string `json:"_id"`
									} `json:"_contentType"`
									Locale  string `json:"_locale"`
									Name    string `json:"name"`
									Content struct {
										NodeType string `json:"nodeType"`
										Data     struct {
										} `json:"data"`
										Content []struct {
											NodeType string `json:"nodeType"`
											Content  []struct {
												NodeType string `json:"nodeType"`
												Value    string `json:"value"`
												Marks    []struct {
													Type string `json:"type"`
												} `json:"marks"`
												Data struct {
												} `json:"data"`
											} `json:"content"`
											Data struct {
											} `json:"data"`
										} `json:"content"`
									} `json:"content"`
								} `json:"data"`
								ReferenceID string `json:"referenceId"`
							} `json:"components"`
						} `json:"data"`
						ReferenceID string `json:"referenceId"`
					} `json:"component"`
				} `json:"accessibility"`
				MightAlsoLike struct {
					ID          string `json:"_id"`
					ContentType struct {
						ID string `json:"_id"`
					} `json:"_contentType"`
					Locale       string   `json:"_locale"`
					Identifier   string   `json:"identifier"`
					Slug         string   `json:"slug"`
					ProductSlugs []string `json:"productSlugs"`
				} `json:"mightAlsoLike"`
				TopSearch struct {
					ID          string `json:"_id"`
					ContentType struct {
						ID string `json:"_id"`
					} `json:"_contentType"`
					Locale       string   `json:"_locale"`
					Identifier   string   `json:"identifier"`
					Slug         string   `json:"slug"`
					ProductSlugs []string `json:"productSlugs"`
				} `json:"topSearch"`
			} `json:"pageData"`
			MejuriAPIHost string `json:"mejuriApiHost"`
			Query         struct {
			} `json:"query"`
			PreviewMode bool   `json:"previewMode"`
			Country     string `json:"country"`
			Page        struct {
				ID          string `json:"_id"`
				ContentType struct {
					ID string `json:"_id"`
				} `json:"_contentType"`
				Locale                string `json:"_locale"`
				Name                  string `json:"name"`
				Title                 string `json:"title"`
				Slug                  string `json:"slug"`
				EnablePageTranslation bool   `json:"enablePageTranslation"`
				Metatags              struct {
					ID          string `json:"_id"`
					ContentType struct {
						ID string `json:"_id"`
					} `json:"_contentType"`
					Locale           string `json:"_locale"`
					Name             string `json:"name"`
					DescriptionTitle string `json:"descriptionTitle"`
					Description      string `json:"description"`
					AplicationName   string `json:"aplicationName"`
				} `json:"metatags"`
				Schema struct {
					ID          string `json:"_id"`
					ContentType struct {
						ID string `json:"_id"`
					} `json:"_contentType"`
					Locale string `json:"_locale"`
					Name   string `json:"name"`
					Schema struct {
						URL          string   `json:"url"`
						Logo         string   `json:"logo"`
						Name         string   `json:"name"`
						Type         string   `json:"@type"`
						SameAs       []string `json:"sameAs"`
						Context      string   `json:"@context"`
						Founders     string   `json:"founders"`
						FoundingDate string   `json:"foundingDate"`
					} `json:"schema"`
				} `json:"schema"`
				Components []struct {
					Data struct {
						Components []struct {
							Data struct {
								Components []interface{} `json:"components"`
							} `json:"data"`
							Name            string        `json:"name"`
							Type            string        `json:"type"`
							Contexts        []interface{} `json:"contexts"`
							Behaviours      []interface{} `json:"behaviours"`
							ReferenceID     string        `json:"referenceId"`
							TrackingName    string        `json:"trackingName"`
							CSSByBreakpoint []struct {
								CSS        []interface{} `json:"css"`
								CSSHover   []interface{} `json:"cssHover"`
								CSSActive  []interface{} `json:"cssActive"`
								Breakpoint struct {
									MinWidth    int    `json:"minWidth"`
									Orientation string `json:"orientation"`
								} `json:"breakpoint"`
								CSSDisabled []interface{} `json:"cssDisabled"`
							} `json:"cssByBreakpoint"`
						} `json:"components"`
					} `json:"data"`
					Name            string        `json:"name"`
					Type            string        `json:"type"`
					Contexts        []interface{} `json:"contexts"`
					Behaviours      []interface{} `json:"behaviours"`
					ReferenceID     string        `json:"referenceId"`
					TrackingName    string        `json:"trackingName"`
					CSSByBreakpoint []struct {
						CSS        []interface{} `json:"css"`
						CSSHover   []interface{} `json:"cssHover"`
						CSSActive  []interface{} `json:"cssActive"`
						Breakpoint struct {
							MinWidth    int    `json:"minWidth"`
							Orientation string `json:"orientation"`
						} `json:"breakpoint"`
						CSSDisabled []interface{} `json:"cssDisabled"`
					} `json:"cssByBreakpoint"`
					HTMLID    string        `json:"htmlId,omitempty"`
					Labels    []interface{} `json:"labels,omitempty"`
					ClassName string        `json:"className,omitempty"`
				} `json:"components"`
			} `json:"page"`
			Products struct {
			} `json:"products"`
		} `json:"pageProps"`
		NSSP bool `json:"__N_SSP"`
	} `json:"props"`
	Page  string `json:"page"`
	Query struct {
	} `json:"query"`
	BuildID      string          `json:"buildId"`
	AssetPrefix  string          `json:"assetPrefix"`
	IsFallback   bool            `json:"isFallback"`
	DynamicIds   []string        `json:"dynamicIds"`
	Gssp         bool            `json:"gssp"`
	CustomServer bool            `json:"customServer"`
	Head         [][]interface{} `json:"head"`
}

var productsExtractReg = regexp.MustCompile(`(?U)id="__NEXT_DATA__"\s*type="application/json">\s*({.*})\s*</script>`)

func (c *_Crawler) parseCategories(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	matched := productsExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		c.logger.Debugf("%s", respBody)
		return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}

	var viewData categoryStructure
	if err := json.Unmarshal(matched[1], &viewData); err != nil {
		c.logger.Errorf("unmarshal category detail data fialed, error=%s", err)
		return err
	}

	for _, rawCat := range viewData.Props.PageProps.PageData.Header.Children {

		cateName := rawCat.Text
		if cateName == "" {
			continue
		}
		//nnctx := context.WithValue(ctx, "Category", cateName)
		fmt.Println(`cateName `, cateName)

		for _, rawsub1Cat := range rawCat.Children {

			subcat := rawsub1Cat.Text
			if subcat != "" {
				fmt.Println(`SubCat-1`, subcat)
			}

			for _, rawsub2Cat := range rawsub1Cat.Children {
				subcat2 := rawsub2Cat.Text
				if subcat2 == "" {
					continue
				}

				href := rawsub2Cat.URL
				if href == "" {
					continue
				}
				//fmt.Println(`link `, href)

				_, err := url.Parse(href)
				if err != nil {
					//c.logger.Error("parse url %s failed", href)
					continue
				}

				subCateName := subcat + " > " + subcat2
				fmt.Println(`SubCatName `, subCateName)

				// nnnctx := context.WithValue(nnctx, "SubCategory", subCateName)
				// req, _ := http.NewRequest(http.MethodGet, href, nil)
				// if err := yield(nnnctx, req); err != nil {
				// return err
			}
		}
	}
	return nil
}

// nextIndex used to get sharingData from context
func nextIndex(ctx context.Context) int {
	return int(strconv.MustParseInt(ctx.Value("item.index")) + 1)
}

type productListType struct {
	ItemCount int `json:"itemCount"`
	Products  []struct {
		ID     int    `json:"id"`
		Sku    string `json:"sku"`
		URLKey string `json:"urlKey"`
	} `json:"products"`
	Page int `json:"page"`
}

var prodDataExtraReg = regexp.MustCompile(`window\['plpData'\]\s*=\s*({.*})`)

// parseCategoryProducts parse api url from web page url
func (c *_Crawler) parseCategoryProducts(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.logger.Debug(err)
		return err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return err
	}

	lastIndex := nextIndex(ctx)
	sel := doc.Find(`.product-name`)
	for i := range sel.Nodes {
		node := sel.Eq(i)
		if href, _ := node.Find("a").Attr("href"); href != "" {
			parseurl := resp.Request.URL.Scheme + "://" + resp.Request.URL.Host + strings.ReplaceAll(href, "/shop/", "/api/v2/")
			// c.logger.Debugf("yield %w%s", lastIndex, parseurl)

			req, err := http.NewRequest(http.MethodGet, parseurl, nil)
			if err != nil {
				c.logger.Error(err)
				continue
			}
			lastIndex += 1
			nctx := context.WithValue(ctx, "item.index", lastIndex)
			if err := yield(nctx, req); err != nil {
				return err
			}
		}
	}

	// check if this is the last page
	totalResults, _ := strconv.ParseInt(strings.Split(doc.Find(`.collections-products__products-amount`).Text(), " ")[0])
	if lastIndex >= int(totalResults) {
		return nil
	}

	// no next page

	return nil
}

type parseProductResponse struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Slug            string `json:"slug"`
	MetaDescription string `json:"meta_description"`
	MetaKeywords    string `json:"meta_keywords"`
	DisplayName     string `json:"display_name"`
	PriceRange      struct {
		Aud struct {
			Min string `json:"min"`
			Max string `json:"max"`
		} `json:"AUD"`
		Cad struct {
			Min string `json:"min"`
			Max string `json:"max"`
		} `json:"CAD"`
		Gbp struct {
			Min string `json:"min"`
			Max string `json:"max"`
		} `json:"GBP"`
		Usd struct {
			Min string `json:"min"`
			Max string `json:"max"`
		} `json:"USD"`
	} `json:"price_range"`
	ProductCare           string      `json:"product_care"`
	Details               string      `json:"details"`
	EngravingType         interface{} `json:"engraving_type"`
	MaterialName          string      `json:"material_name"`
	Jit                   bool        `json:"jit"`
	PreOrder              bool        `json:"pre_order"`
	FakeReceivingPreorder bool        `json:"fake_receiving_preorder"`
	MaterialDescriptions  []struct {
		IconName    string `json:"icon_name"`
		Name        string `json:"name"`
		Description string `json:"description"`
		IconURL     string `json:"icon_url"`
	} `json:"material_descriptions"`
	Sample              bool        `json:"sample"`
	TravelCase          bool        `json:"travel_case"`
	EngagementRing      bool        `json:"engagement_ring"`
	WeddingBand         bool        `json:"wedding_band"`
	EngagementRingType  interface{} `json:"engagement_ring_type"`
	WeddingBandType     interface{} `json:"wedding_band_type"`
	MensWeddingBandType interface{} `json:"mens_wedding_band_type"`
	MarkedSoldOut       bool        `json:"marked_sold_out"`
	Available           bool        `json:"available"`
	Images              []struct {
		Position   int         `json:"position"`
		Alt        interface{} `json:"alt"`
		Attachment struct {
			URLOriginal string `json:"url_original"`
			URLMini     string `json:"url_mini"`
			URLSmall    string `json:"url_small"`
			URLMedium   string `json:"url_medium"`
			URLLarge    string `json:"url_large"`
		} `json:"attachment"`
	} `json:"images"`
	MaterialGroupProducts []struct {
		ID               int    `json:"id"`
		Slug             string `json:"slug"`
		MaterialCategory struct {
			Name        string `json:"name"`
			IconFullURL string `json:"icon_full_url"`
		} `json:"material_category"`
	} `json:"material_group_products"`
	NoRetail           bool   `json:"no_retail"`
	NoWaitlist         bool   `json:"no_waitlist"`
	NoFairPricing      bool   `json:"no_fair_pricing"`
	DigitalGiftcard    bool   `json:"digital_giftcard"`
	PhysicalGiftcard   bool   `json:"physical_giftcard"`
	DisplayRetailPrice int    `json:"display_retail_price"`
	CostPrice          string `json:"cost_price"`
	Material           string `json:"material"`
	Master             struct {
		ID  int    `json:"id"`
		Sku string `json:"sku"`
	} `json:"master"`

	Variants []struct {
		ID           int    `json:"id"`
		Sku          string `json:"sku"`
		OptionValues []struct {
			Name         string `json:"name"`
			Presentation string `json:"presentation"`
			OptionTypeID int    `json:"option_type_id"`
		} `json:"option_values"`
		Prices []struct {
			Currency string `json:"currency"`
			Amount   string `json:"amount"`
		} `json:"prices"`
	} `json:"variants"`
	OptionTypes []struct {
		ID           int    `json:"id"`
		Name         string `json:"name"`
		Presentation string `json:"presentation"`
	} `json:"option_types"`
}

var (
	detailReg = regexp.MustCompile(`({.*})`)
)

func (c *_Crawler) parseProduct(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error(err)
		return err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		c.logger.Error(err)
		return err
	}

	opts := c.CrawlOptions(resp.Request.URL)
	if !c.productApiPathMatcher.MatchString(resp.Request.URL.Path) {
		//fmt.Println(resp.Request.URL.String())
		produrl := strings.ReplaceAll(resp.Request.URL.String(), "/shop/", "/api/v2/")
		req, err := http.NewRequest(http.MethodGet, produrl, nil)
		if err != nil {
			c.logger.Error(err)
			return err
		}
		req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
		req.Header.Set("content-type", "application/json;charset=utf-8")
		req.Header.Set("referer", resp.Request.URL.String())

		respNew, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
			EnableProxy: true,
			Reliability: opts.Reliability,
		})

		if err != nil {
			c.logger.Error(err)
			return err
		}

		respBody, err = ioutil.ReadAll(respNew.Body)
		respNew.Body.Close()
		if err != nil {
			return err
		}
	}

	matched := detailReg.FindSubmatch(respBody)
	if len(matched) == 0 {
		c.httpClient.Jar().Clear(ctx, resp.Request.URL)
		return fmt.Errorf("extract produt json from page %s content failed", resp.Request.URL)
	}

	var (
		viewData parseProductResponse
	)

	if err := json.Unmarshal(respBody, &viewData); err != nil {
		c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
		return err
	}

	var materialgrpproducts = make([]string, len(viewData.MaterialGroupProducts)+1)
	for i, proditem := range viewData.MaterialGroupProducts {
		if proditem.ID == viewData.ID {
			materialgrpproducts[0] = proditem.Slug
		} else {
			materialgrpproducts[i+1] = proditem.Slug
		}
	}

	canUrl, _ := c.CanonicalUrl(doc.Find(`link[rel="canonical"]`).AttrOr("href", ""))
	if canUrl == "" {
		canUrl, _ = c.CanonicalUrl(resp.Request.URL.String())
	}

	fmt.Println(doc.Find(`.yotpo-sum-reviews`).Text())
	review, _ := strconv.ParseInt(strings.Split(strings.TrimSpace(doc.Find(`.yotpo-sum-reviews`).Text()), " ")[0])
	rating, _ := strconv.ParseFloat(strconv.Format(strings.ReplaceAll(doc.Find(`.yotpo-stars`).Text(), " star rating", "")))

	for pi, prodslug := range materialgrpproducts {

		if pi > 0 && prodslug != "" {

			produrl := resp.Request.URL.Scheme + "://" + resp.Request.URL.Host + "/api/v2/products/" + prodslug
			fmt.Println(produrl)
			req, err := http.NewRequest(http.MethodGet, produrl, nil)
			if err != nil {
				c.logger.Error(err)
				return err
			}
			req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
			req.Header.Set("content-type", "application/json;charset=utf-8")
			req.Header.Set("referer", canUrl)

			respNew, err := c.httpClient.DoWithOptions(ctx, req, http.Options{
				EnableProxy: true,
				Reliability: opts.Reliability,
			})

			if err != nil {
				c.logger.Error(err)
				return err
			}

			respBody, err = ioutil.ReadAll(respNew.Body)
			respNew.Body.Close()
			if err != nil {
				return err
			}
			if err := json.Unmarshal(respBody, &viewData); err != nil {
				c.logger.Errorf("unmarshal product detail data fialed, error=%s", err)
				return err
			}
		} else if prodslug == "" {
			continue
		}

		color := viewData.MaterialName
		desc := viewData.Description + " " + viewData.Details
		for _, proditem := range viewData.MaterialDescriptions {
			desc = strings.Join(([]string{desc, proditem.Name, ": ", proditem.Description}), " ")
		}

		item := pbItem.Product{
			Source: &pbItem.Source{
				Id:           strconv.Format(viewData.ID),
				CrawlUrl:     resp.Request.URL.String(),
				CanonicalUrl: canUrl,
			},
			Title:       viewData.Name,
			Description: desc,
			BrandName:   "",
			Price: &pbItem.Price{
				Currency: regulation.Currency_USD,
			},
			Stats: &pbItem.Stats{
				ReviewCount: int32(review),
				Rating:      float32(rating),
			},
		}

		var medias []*media.Media

		for _, img := range viewData.Images {
			itemImg, _ := anypb.New(&media.Media_Image{
				OriginalUrl: img.Attachment.URLOriginal,
				LargeUrl:    img.Attachment.URLLarge,
				MediumUrl:   img.Attachment.URLMedium,
				SmallUrl:    img.Attachment.URLSmall,
			})
			medias = append(medias, &media.Media{
				Detail:    itemImg,
				IsDefault: img.Position == 0,
			})
		}
		item.Medias = medias

		current, _ := strconv.ParseFloat(viewData.PriceRange.Usd.Min)
		msrp, _ := strconv.ParseFloat(viewData.PriceRange.Usd.Max)
		discount := 0.0

		sizetypeid := 0
		for _, opttype := range viewData.OptionTypes {
			if strings.Contains(strings.ToLower(opttype.Name), "size") || strings.Contains(strings.ToLower(opttype.Presentation), "size") {
				sizetypeid = opttype.ID
				break
			}
		}

		if len(viewData.Variants) > 0 {

			for _, variation := range viewData.Variants {

				if variation.OptionValues[0].OptionTypeID != sizetypeid {
					continue
				}

				sku := pbItem.Sku{
					SourceId: strconv.Format(variation.ID),
					Price: &pbItem.Price{
						Currency: regulation.Currency_USD,
						Current:  int32(current * 100),
						Msrp:     int32(msrp * 100),
						Discount: int32(discount),
					},

					Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
				}
				if viewData.Available {
					sku.Stock.StockStatus = pbItem.Stock_InStock
				}

				if color != "" {
					//sku.SourceId = fmt.Sprintf("%s-%v", color, viewData.ID)
					sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
						Type:  pbItem.SkuSpecType_SkuSpecColor,
						Id:    fmt.Sprintf("%s-%v", color, viewData.ID),
						Name:  color,
						Value: color,
					})
				}

				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecSize,
					Id:    variation.Sku,
					Name:  variation.OptionValues[0].Name,
					Value: variation.OptionValues[0].Name,
				})
				item.SkuItems = append(item.SkuItems, &sku)
			}
		} else {

			sku := pbItem.Sku{
				SourceId: strconv.Format(viewData.ID),
				Price: &pbItem.Price{
					Currency: regulation.Currency_USD,
					Current:  int32(current * 100),
					Msrp:     int32(msrp * 100),
					Discount: int32(discount),
				},

				Stock: &pbItem.Stock{StockStatus: pbItem.Stock_OutOfStock},
			}
			if viewData.Available {
				sku.Stock.StockStatus = pbItem.Stock_InStock
			}

			if color != "" {
				//sku.SourceId = fmt.Sprintf("%s-%v", color, viewData.ID)
				sku.Specs = append(sku.Specs, &pbItem.SkuSpecOption{
					Type:  pbItem.SkuSpecType_SkuSpecColor,
					Id:    fmt.Sprintf("%s-%v", color, viewData.ID),
					Name:  color,
					Value: color,
				})
			}
			item.SkuItems = append(item.SkuItems, &sku)
		}
		if err = yield(ctx, &item); err != nil {
			return err
		}
	}
	return nil
}

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		"https://mejuri.com/",
		"https://mejuri.com/shop/t/type?fbm=Gold%20Vermeil",
		//"https://mejuri.com/shop/products/large-diamond-necklace",
		//"https://mejuri.com/shop/products/heirloom-ring-garnet",
		//"https://mejuri.com/shop/t/type/pendants",
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
	cli.NewApp(New).Run(os.Args)
}
