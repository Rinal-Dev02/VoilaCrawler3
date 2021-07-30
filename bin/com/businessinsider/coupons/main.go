package main

import (
	"context"
	"encoding/json"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/voiladev/VoilaCrawler/pkg/cli"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	pbItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	pbProxy "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
	"github.com/voiladev/go-framework/timeutil"
)

type _Crawler struct {
	crawler.MustImplementCrawler

	httpClient              http.Client
	categoryPathMatcher     *regexp.Regexp
	categoryJsonPathMatcher *regexp.Regexp
	productGroupPathMatcher *regexp.Regexp
	productPathMatcher      *regexp.Regexp
	logger                  glog.Log
}

func (_ *_Crawler) New(_ *cli.Context, client http.Client, logger glog.Log) (crawler.Crawler, error) {
	c := _Crawler{
		httpClient: client,
		logger:     logger.New("_Crawler"),
	}
	return &c, nil
}

// ID
func (c *_Crawler) ID() string {
	return "1cbf9602a1e69bc2159f537cbd275269"
}

// Version
func (c *_Crawler) Version() int32 {
	return 1
}

// CrawlOptions
func (c *_Crawler) CrawlOptions(u *url.URL) *crawler.CrawlOptions {
	options := crawler.NewCrawlOptions()
	options.EnableHeadless = false
	options.EnableSessionInit = true
	options.Reliability = pbProxy.ProxyReliability_ReliabilityMedium
	options.MustCookies = append(options.MustCookies)

	return options
}

func (c *_Crawler) AllowedDomains() []string {
	return []string{"coupons.businessinsider.com"}
}

type RawVoucher struct {
	ExpireDiff            int         `json:"expireDiff"`
	IsInStore             bool        `json:"isInStore"`
	IsDeal                bool        `json:"isDeal"`
	IsCode                bool        `json:"isCode"`
	IsSales               bool        `json:"isSales"`
	IsBranded             bool        `json:"isBranded"`
	IsEditor              bool        `json:"isEditor"`
	IsExclusive           bool        `json:"isExclusive"`
	IsAffiliate           bool        `json:"isAffiliate"`
	VerifiedDiff          int         `json:"verifiedDiff"`
	ID                    string      `json:"_id"`
	IDVoucher             int         `json:"id_voucher"`
	AffiliateURL          string      `json:"affiliate_url"`
	CreationTime          string      `json:"creation_time"`
	Duplicate             interface{} `json:"duplicate"`
	EncryptedAffiliateURL string      `json:"encrypted_affiliate_url"`
	EndTime               string      `json:"end_time"`
	FeaturedCategoryPos   int         `json:"featured_category_pos"`
	FeaturedMobile        int         `json:"featured_mobile"`
	IDAffnetwork          string      `json:"id_affnetwork"`
	IDCampaign            string      `json:"id_campaign"`
	IDClient              string      `json:"id_client"`
	IDMerchant            int         `json:"id_merchant"`
	IDMerchantPool        int         `json:"id_merchant_pool"`
	Kpi                   float64     `json:"kpi"`
	NumberOfClickouts     int         `json:"number_of_clickouts"`
	Pos                   int         `json:"pos"`
	Published             int         `json:"published"`
	RetailerLogoLarge     string      `json:"retailer_logo_large"`
	RetailerLogoMedium    string      `json:"retailer_logo_medium"`
	RetailerLogoSmall     string      `json:"retailer_logo_small"`
	RetailerThumbLarge    string      `json:"retailer_thumb_large"`
	RetailerThumbMedium   string      `json:"retailer_thumb_medium"`
	RetailerThumbSmall    string      `json:"retailer_thumb_small"`
	StartTime             string      `json:"start_time"`
	Top20Index            int         `json:"top20Index"`
	TopIndex              interface{} `json:"topIndex"`
	Verify                struct {
		Verified bool   `json:"verified"`
		LastDate string `json:"last_date"`
	} `json:"verify"`
	SyncLastModifiedTimestamp   string        `json:"syncLastModifiedTimestamp"`
	SyncCreationTimestamp       string        `json:"syncCreationTimestamp"`
	InStoreHiddenCaptions       []interface{} `json:"inStoreHiddenCaptions"`
	InStoreShownCaptions        []interface{} `json:"inStoreShownCaptions"`
	SubscriberHash              string        `json:"subscriberHash"`
	TermsAndConditionsFirstLoad bool          `json:"terms_and_conditions_first_load"`
	TermsAndConditions          string        `json:"terms_and_conditions"`
	DefaultDynamic              int           `json:"default_dynamic"`
	SubmittedByUser             bool          `json:"submitted_by_user"`
	FakeExpiredVoucher          string        `json:"fake_expired_voucher"`
	CaptionCSSClass             []interface{} `json:"captionCssClass"`
	CSSClass                    []string      `json:"cssClass"`
	Instore                     interface{}   `json:"instore"`
	VoucherImage                string        `json:"voucher_image"`
	Dynamic                     int           `json:"dynamic"`
	ClickoutMonth               int           `json:"clickoutMonth"`
	ClickoutWeek                int           `json:"clickoutWeek"`
	ClickoutDay                 int           `json:"clickoutDay"`
	Tags                        []interface{} `json:"tags"`
	Captions                    []struct {
		Text  string `json:"text"`
		Title string `json:"title"`
	} `json:"captions"`
	OnlyForRegisteredUsers interface{} `json:"only_for_registered_users"`
	Code                   string      `json:"code"`
	Retailer               string      `json:"retailer"`
	IDCategory             int         `json:"id_category"`
	RetailerImageAlt       string      `json:"retailer_image_alt"`
	RetailerSeoURL         string      `json:"retailer_seo_url"`
	RetailerThumb          string      `json:"retailer_thumb"`
	RetailerLogo           string      `json:"retailer_logo"`
	Description            string      `json:"description"`
	Caption2               string      `json:"caption_2"`
	Caption1               string      `json:"caption_1"`
	EditorsPick            int         `json:"editors_pick"`
	HasCodes               int         `json:"has_codes"`
	SalesPeriod            int         `json:"sales_period"`
	BrandedVoucher         int         `json:"branded_voucher"`
	ExclusiveVoucher       int         `json:"exclusive_voucher"`
	AffiliateMode          int         `json:"affiliate_mode"`
	Title                  string      `json:"title"`
	IDVoucherPool          int         `json:"id_voucher_pool"`
}

type VoucherpopupResponse struct {
	InStorePdfHTML          string        `json:"inStorePdfHtml"`
	MembersOnly             bool          `json:"membersOnly"`
	ShowHidden              bool          `json:"showHidden"`
	ChangeClickoutBehaviour bool          `json:"changeClickoutBehaviour"`
	Voucher                 *RawVoucher   `json:"voucher"`
	OtherCodes              []*RawVoucher `json:"otherCodes"`
	VouchersToReviel        []interface{} `json:"vouchersToReviel"`
	IsUniqueCode            bool          `json:"isUniqueCode"`
	SimilarVouchersCount    int           `json:"similarVouchersCount"`
	Error                   bool          `json:"error"`
}

func (c *_Crawler) Parse(ctx context.Context, resp *http.Response, yield func(context.Context, interface{}) error) error {
	if c == nil || yield == nil {
		return nil
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}

	var (
		items    []*pbItem.PromoCode
		itemDict = map[string]*pbItem.PromoCode{}
	)
	sel := doc.Find(".businessinsiderus-main .businessinsiderus-main-holder .vouchers-listed>.code")
	for i := 0; i < len(sel.Nodes); i++ {
		node := sel.Eq(i)

		item := pbItem.PromoCode{
			Source:    &pbItem.Source{Id: node.AttrOr("data-gtm-voucher-id", node.AttrOr("data-ju-wvxjoly-pk", ""))},
			Type:      pbItem.PromoCode_ProductCode,
			Index:     int32(i + 1),
			ExtraInfo: map[string]string{},
		}
		items = append(items, &item)
		itemDict[item.GetSource().GetId()] = &item
	}

	for _, item := range items {
		if _, ok := itemDict[item.GetSource().Id]; !ok {
			continue
		}

		popupUrl, _ := url.Parse("https://coupons.businessinsider.com/ajax/voucherpopup")
		vals := popupUrl.Query()
		vals.Add("id", item.GetSource().GetId())
		vals.Add("isTablet", "false")
		vals.Add("redirectUrl", resp.Request.URL.Path)
		for _, _item := range items {
			if _, ok := itemDict[_item.GetSource().Id]; !ok || _item == item {
				continue
			}
			vals.Add("otherCodes", _item.GetSource().GetId())
		}
		vals.Add("_", strconv.Format(time.Now().UnixNano()/1000000))
		popupUrl.RawQuery = vals.Encode()

		req, _ := http.NewRequest(http.MethodGet, popupUrl.String(), nil)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		req.Header.Set("Referer", resp.Request.URL.String())
		req.AddCookie(&http.Cookie{Name: "voucher_co", Value: item.GetSource().Id})

		voResp, err := func() (resp *http.Response, err error) {
			for i := 0; i < 3; i++ {
				resp, err = func() (resp *http.Response, err error) {
					nctx, cancel := context.WithTimeout(ctx, time.Second*30)
					defer cancel()

					resp, err = c.httpClient.DoWithOptions(nctx, req, http.Options{
						EnableProxy:       true,
						EnableHeadless:    false,
						EnableSessionInit: false,
						Reliability:       c.CrawlOptions(nil).Reliability,
					})
					if err != nil {
						c.logger.Errorf("access %s failed, error=%s", req.URL, err)
						return
					}
					return
				}()
				if err != nil {
					continue
				}
				return
			}
			return
		}()
		if err != nil {
			continue
		}

		var respData VoucherpopupResponse
		if err := json.NewDecoder(voResp.Body).Decode(&respData); err != nil {
			c.logger.Errorf("decode voucherpopup response failed, error=%s", err)
			return err
		}
		if respData.Error {
			c.logger.Errorf("get data of %s failed, error=%s", item.GetSource().GetId())
			continue
		}

		for _, voucher := range append([]*RawVoucher{respData.Voucher}, respData.OtherCodes...) {
			id := strconv.Format(voucher.IDVoucher)
			item := itemDict[id]
			if item == nil {
				continue
			}

			item.Retailer = voucher.Retailer
			item.Title = voucher.Title
			item.Type = pbItem.PromoCode_ProductCode
			if strings.Contains(strings.ToLower(voucher.Title), "delivery") {
				item.Type = pbItem.PromoCode_DeliveryCode
			}

			item.Description = voucher.Description
			if voucher.TermsAndConditions != "" {
				item.Condition = &pbItem.PromoCode_Condition{
					RawCondition: voucher.TermsAndConditions,
				}
			}
			item.Code = voucher.Code
			item.Discount = &pbItem.PromoCode_Discount{
				RawDiscount: voucher.Caption1,
				Description: strings.Join([]string{voucher.Caption1, voucher.Caption2}, " "),
			}
			if t := timeutil.TimeParse(voucher.StartTime); !t.IsZero() {
				item.StartUtc = t.Unix()
			}
			if t := timeutil.TimeParse(voucher.EndTime); !t.IsZero() {
				item.ExpiresUtc = t.Unix()
			}

			affUrl, err := url.Parse(voucher.AffiliateURL)
			if err != nil {
				c.logger.Errorf("got invalid affiliate url %s", voucher.AffiliateURL)
				continue
			}

			var targetUrl string
			if affUrl.Path == "/deeplink" && affUrl.Query().Get("murl") != "" {
				targetUrl, _ = url.QueryUnescape(affUrl.Query().Get("murl"))
			}

			if targetUrl == "" {
				req, err := http.NewRequest(http.MethodGet, voucher.AffiliateURL, nil)
				if err != nil {
					c.logger.Errorf("got invalid url %s", voucher.AffiliateURL)
					continue
				}

				rawPageResp, err := func() (resp *http.Response, err error) {
					for i := 0; i < 3; i++ {
						resp, err = func() (resp *http.Response, err error) {
							nctx, cancel := context.WithTimeout(ctx, time.Second*30)
							defer cancel()

							resp, err = c.httpClient.DoWithOptions(nctx, req, http.Options{
								EnableProxy: true,
								Reliability: c.CrawlOptions(nil).Reliability,
							})
							if err != nil {
								c.logger.Errorf("got response from url %s failed, error=%s", voucher.AffiliateURL, err)
								return
							}
							return
						}()
						if err != nil {
							continue
						}
						return
					}
					return
				}()
				if err != nil {
					continue
				}

				if rawPageResp.Header.Get("Refresh") != "" {
					targetUrl = strings.TrimPrefix(rawPageResp.Header.Get("Refresh"), "0;url=")
				} else {
					targetUrl = rawPageResp.Request.URL.String()
				}
			}

			target := &pbItem.PromoCode_ApplyTarget{
				Url:    targetUrl,
				RawUrl: voucher.AffiliateURL,
			}
			item.ApplyTargets = append(item.ApplyTargets, target)

			if err := yield(ctx, item); err != nil {
				return err
			}
			delete(itemDict, id)
		}
	}
	return nil
}

func (c *_Crawler) NewTestRequest(ctx context.Context) (reqs []*http.Request) {
	for _, u := range []string{
		"https://coupons.businessinsider.com/shein",
		//	"https://coupons.businessinsider.com/asos",
		// "https://coupons.businessinsider.com/sephora",
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
