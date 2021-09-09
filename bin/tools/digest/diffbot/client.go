package diffbot

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/voiladev/VoilaCrawler/bin/tools/digest/util"
	"github.com/voiladev/go-framework/glog"
)

type Product struct {
	Type              string `json:"type"`
	PageURL           string `json:"pageUrl"`
	ResolvedPageUrl   string `json:"resolvedPageUrl"`
	Title             string `json:"title"`
	Text              string `json:"text"`
	Brand             string `json:"brand"`
	OfferPrice        string `json:"offerPrice"`
	RegularPrice      string `json:"regularPrice"`
	ShippingAmount    string `json:"shippingAmount"`
	SaveAmount        string `json:"saveAmount"`
	OfferPriceDetails struct {
		Symbol string  `json:"symbol"`
		Amount float64 `json:"amount"`
		Text   string  `json:"text"`
	} `json:"offerPriceDetails"`
	RegularPriceDetails struct {
		Symbol string  `json:"symbol"`
		Amount float64 `json:"amount"`
		Text   string  `json:"text"`
	} `json:"regularPriceDetails"`
	SaveAmountDetails struct {
		Symbol     string      `json:"symbol"`
		Amount     float64     `json:"amount"`
		Text       string      `json:"text"`
		Percentage interface{} `json:"percentage"`
	} `json:"saveAmountDetails"`
	ProductID string      `json:"productId"`
	UPC       string      `json:"upc"`
	Sku       string      `json:"sku"`
	MPN       string      `json:"mpn"`
	ISBN      string      `json:"isbn"`
	Specs     interface{} `json:"specs"`
	// Specs map[string]string
	Images []struct {
		Xpath         string `json:"xpath"`
		NaturalHeight int    `json:"naturalHeight"`
		Width         int    `json:"width"`
		DiffbotURI    string `json:"diffbotUri"`
		Title         string `json:"title"`
		URL           string `json:"url"`
		NaturalWidth  int    `json:"naturalWidth"`
		Primary       bool   `json:"primary"`
		Height        int    `json:"height"`
	} `json:"images"`
	PrefixCode       string `json:"prefixCode"`
	ProductOrigin    string `json:"productOrigin"`
	HumanLanguage    string `json:"humanLanguage"`
	DiffbotURI       string `json:"diffbotUri"`
	MultipleProducts bool   `json:"multipleProducts"`
	// Availability     bool   `json:"availability"`
	Category string `json:"category"`
}

type _RawResponse struct {
	ErrorCode int32  `json:"errorCode"`
	Error     string `json:"error"`
	Request   struct {
		PageURL string `json:"pageUrl"`
		API     string `json:"api"`
		Version int    `json:"version"`
	} `json:"request"`
	Objects []*Product `json:"objects"`
}

type DiffbotCient struct {
	httpClient *http.Client
	token      string
	apiURI     *url.URL
	logger     glog.Log
}

func New(token string, logger glog.Log) (*DiffbotCient, error) {
	if token == "" {
		return nil, errors.New("invalid token")
	}
	u, _ := url.Parse("https://api.diffbot.com/v3/product")

	client := DiffbotCient{
		httpClient: &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}},
		token:      token,
		apiURI:     u,
		logger:     logger.New("DiffbotClent"),
	}
	return &client, nil
}

var ErrRetry = errors.New("download content failed, retry")

func (c *DiffbotCient) Fetch(ctx context.Context, rawurl string) ([]*Product, error) {
	u := *c.apiURI
	vals := u.Query()
	vals.Set("token", c.token)
	vals.Set("url", rawurl)
	vals.Set("timeout", "30000")
	vals.Set("discussion", "false")
	vals.Set("paging", "false")
	u.RawQuery = vals.Encode()

	var ret _RawResponse
	for i := 0; i < 3; i++ {
		if i > 0 {
			c.logger.Debugf("retry for %s", u.String())
		}
		if err := func() error {
			nctx, cancel := context.WithTimeout(ctx, time.Second*30)
			defer cancel()

			req, _ := http.NewRequestWithContext(nctx, http.MethodGet, u.String(), nil)
			resp, err := c.httpClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode == 200 {
				if err := json.NewDecoder(resp.Body).Decode(&ret); err != nil {
					c.logger.Error(err)
					return err
				}
				if ret.ErrorCode != 0 {
					c.logger.Errorf("fetch faied, code=%v, message=%v", ret.ErrorCode, ret.Error)
					if ret.Error == "Could not download page (403)" {
						return ErrRetry
					}
					return errors.New(ret.Error)
				}
				return nil
			}
			return errors.New(resp.Status)
		}(); err == ErrRetry {
			continue
		} else if err != nil {
			return ret.Objects, err
		}
		break
	}

	filter := map[string]struct{}{}
	for _, prod := range ret.Objects {
		imgs := prod.Images[0:0]
		for _, img := range prod.Images {
			if _, ok := imgBlackList[img.URL]; ok {
				continue
			}

			nurl, err := util.FormatImageUrl(img.URL)
			if err != nil {
				continue
			}
			if _, ok := filter[nurl]; ok {
				continue
			}
			filter[nurl] = struct{}{}

			img.URL = nurl
			imgs = append(imgs, img)
		}
		prod.Images = imgs
	}
	return ret.Objects, nil
}

var imgBlackList = map[string]struct{}{
	"https://is4.revolveassets.com/images/wwu/2017/january/womens/011017_f_popup.jpg": {},
}
