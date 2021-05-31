package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/robertkrimen/otto"
)

var productsDataExtractReg = regexp.MustCompile(`(?U)<script type="application/ld\+json">\s*({.*})\s*</script>`)
var categoryPathMatcher = regexp.MustCompile(`^(.*)(\d)$`)

func TrimSpaceNewlineInString(s []byte) []byte {
	re := regexp.MustCompile(`\n`)
	resp := re.ReplaceAll(s, []byte(" "))
	resp = bytes.ReplaceAll(resp, []byte("\\n"), []byte(""))
	resp = bytes.ReplaceAll(resp, []byte("\r"), []byte(""))
	resp = bytes.ReplaceAll(resp, []byte("\t"), []byte(""))
	resp = bytes.ReplaceAll(resp, []byte("&lt;"), []byte("<"))
	resp = bytes.ReplaceAll(resp, []byte("&gt;"), []byte(">"))

	return resp
}

type ProductCategoryPaginationStructure struct {
	Browserlanguage   string `json:"browserLanguage"`
	Catalogparameters struct {
		Isocode        string `json:"isoCode"`
		Idshop         string `json:"idShop"`
		Idsection      string `json:"idSection"`
		Optionalparams struct {
			Idsubsection  string   `json:"idSubSection"`
			Menu          []string `json:"menu"`
			Columnsperrow int      `json:"columnsPerRow"`
		} `json:"optionalParams"`
	} `json:"catalogParameters"`
}

type ProductCategoryStructure struct {
	Version          string `json:"version"`
	Cacheid          string `json:"cacheId"`
	Shop             string `json:"shop"`
	Brand            string `json:"brand"`
	Categoryid       string `json:"categoryId"`
	Lastpage         bool   `json:"lastPage"`
	Columns          int    `json:"columns"`
	Tiposeccion      string `json:"tipoSeccion"`
	Categoriaseccion string `json:"categoriaSeccion"`
	Titleh1          string `json:"titleh1"`
	Description      string `json:"description"`
	Labels           struct {
		Colorlabel                    string `json:"colorLabel"`
		Addtobaglabel                 string `json:"addToBagLabel"`
		Addtobagfulllabel             string `json:"addToBagFullLabel"`
		Nostocklabel                  string `json:"noStockLabel"`
		Nostockandnotifymelabel       string `json:"noStockAndNotifyMeLabel"`
		Selectacolorlabel             string `json:"selectAColorLabel"`
		Selectyoursizelabel           string `json:"selectYourSizeLabel"`
		Addingtobaglabel              string `json:"addingToBagLabel"`
		Seelabel                      string `json:"seeLabel"`
		Addtofavorites                string `json:"addToFavorites"`
		Deletefromfavorites           string `json:"deleteFromFavorites"`
		Firststrikethroughpricelabel  string `json:"firstStrikethroughPriceLabel"`
		Secondstrikethroughpricelabel string `json:"secondStrikethroughPriceLabel"`
		Thirdstrikethroughpricelabel  string `json:"thirdStrikethroughPriceLabel"`
		Currentpricelabel             string `json:"currentPriceLabel"`
		Addedtobaglabel               string `json:"addedToBagLabel"`
		Addtobagerrorlabel            string `json:"addToBagErrorLabel"`
		Colorslabel                   string `json:"colorsLabel"`
		Notifymelabel                 string `json:"notifyMeLabel"`
		Selectacolorandsizelabel      string `json:"selectAColorAndSizeLabel"`
		Addedtofavorites              string `json:"addedToFavorites"`
		Deletedfromfavorites          string `json:"deletedFromFavorites"`
	} `json:"labels"`
	Groups []struct {
		Showseparator   bool        `json:"showSeparator"`
		Separatorimgurl interface{} `json:"separatorImgUrl"`
		Separatortext1  string      `json:"separatorText1"`
		Separatortext2  string      `json:"separatorText2"`
		Garments        map[string]struct {
			//G8706054302 struct {
			ID               string `json:"id"`
			Garmentid        string `json:"garmentId"`
			Type             string `json:"type"`
			Name             string `json:"name"`
			Shortdescription string `json:"shortDescription"`
			Stock            int    `json:"stock"`
			Shownotifyme     bool   `json:"showNotifyMe"`
			Genre            string `json:"genre"`
			Colors           []struct {
				ID           string `json:"id"`
				Label        string `json:"label"`
				Iconurl      string `json:"iconUrl"`
				Linkanchor   string `json:"linkAnchor"`
				Defaultcolor bool   `json:"defaultColor,omitempty"`
				Stock        int    `json:"stock"`
				Price        struct {
					Showcrossedoutprices bool     `json:"showCrossedOutPrices"`
					Crossedoutprices     []string `json:"crossedOutPrices"`
					Saleprice            string   `json:"salePrice"`
					Salepricenocurrency  float64  `json:"salePriceNoCurrency"`
					Discountrate         int      `json:"discountRate"`
					Currency             string   `json:"currency"`
					Locale               struct {
						Language string `json:"language"`
						Country  string `json:"country"`
					} `json:"locale"`
					Accessibilitydiscountlabel string `json:"accessibilityDiscountLabel"`
				} `json:"price"`
				Sizes []struct {
					ID        int         `json:"id"`
					Label     string      `json:"label"`
					Stock     int         `json:"stock"`
					Extrainfo interface{} `json:"extraInfo"`
				} `json:"sizes"`
				Images []struct {
					Img1Src   string `json:"img1Src"`
					Img1Hqsrc string `json:"img1HQSrc"`
					Img2Src   string `json:"img2Src"`
					Img2Hqsrc string `json:"img2HQSrc"`
					Alttext   string `json:"altText"`
				} `json:"images"`
				Analyticseventsdata struct {
					ID             string `json:"id"`
					Name           string `json:"name"`
					Brand          string `json:"brand"`
					Categoryid     string `json:"categoryId"`
					Category       string `json:"category"`
					Colorid        string `json:"colorId"`
					Variant        string `json:"variant"`
					List           string `json:"list"`
					Position       int    `json:"position"`
					Ispersonalized bool   `json:"isPersonalized"`
					Dimension108   string `json:"dimension108"`
				} `json:"analyticsEventsData"`
				Accessibilitytext     string `json:"accessibilityText"`
				Hasanyunavailablesize bool   `json:"hasAnyUnavailableSize,omitempty"`
			} `json:"colors"`
			//} `json:"g8706054302"`
		} `json:"garments"`
	} `json:"groups"`
}

type Article struct {
	Name                     string `json:"name"`
	Instore                  bool   `json:"inStore"`
	Energyclass              string `json:"energyClass"`
	Energyclassinterval      string `json:"energyClassInterval"`
	Energyclasscode          string `json:"energyClassCode"`
	Energyclassintervalcode  string `json:"energyClassIntervalCode"`
	Colorcode                string `json:"colorCode"`
	Totalstylewitharticles   string `json:"totalStyleWithArticles"`
	Stylewithdefaultarticles string `json:"styleWithDefaultArticles"`
	Productswithstylewith    string `json:"productsWithStyleWith"`
	Selection                bool   `json:"selection"`
	Description              string `json:"description"`
	Images                   []struct {
		Thumbnail  string `json:"thumbnail"`
		Image      string `json:"image"`
		Fullscreen string `json:"fullscreen"`
		Zoom       string `json:"zoom"`
	} `json:"images"`
	Video struct {
	} `json:"video"`
	Sizes []struct {
		Sizecode string `json:"sizeCode"`
		Size     string `json:"size"`
		Name     string `json:"name"`
	} `json:"sizes"`
	Whiteprice       string `json:"whitePrice"`
	Whitepricevalue  string `json:"whitePriceValue"`
	Redpricevalue    string `json:"redPriceValue"`
	Marketingmarkers []struct {
		URL       string `json:"url"`
		Text      string `json:"text"`
		Legaltext string `json:"legalText"`
		Color     string `json:"color"`
		Type      string `json:"type"`
	} `json:"marketingMarkers"`
	Promomarkericon string      `json:"promoMarkerIcon"`
	Priceclub       interface{} `json:"priceClub"`
	Concept         []string    `json:"concept"`
	Scenario1       bool        `json:"scenario_1"`
	Compositions    []string    `json:"compositions"`
	Composition     []struct {
		Compositiontype string `json:"compositionType"`
		Materials       []struct {
			Name   string `json:"name"`
			Amount string `json:"amount"`
		} `json:"materials"`
	} `json:"composition"`
	Detaileddescriptions           []string `json:"detailedDescriptions"`
	URL                            string   `json:"url"`
	Producttransparencyenabled     bool     `json:"productTransparencyEnabled"`
	Comingsoon                     bool     `json:"comingSoon"`
	Suppliersdetailenabled         bool     `json:"suppliersDetailEnabled"`
	Supplierssectiondisabledreason string   `json:"suppliersSectionDisabledReason"`
	Productattributes              struct {
		Details []string `json:"details"`
		Main    []string `json:"main"`
		Values  struct {
			Waistrise                  []string `json:"waistRise"`
			Composition                []string `json:"composition"`
			Material                   []string `json:"material"`
			Detaileddescriptions       []string `json:"detailedDescriptions"`
			Imported                   string   `json:"imported"`
			Concepts                   []string `json:"concepts"`
			Nicetoknow                 []string `json:"niceToKnow"`
			Pricedetails               string   `json:"priceDetails"`
			Countryofproductionmessage bool     `json:"countryOfProductionMessage"`
			Countryofproduction        string   `json:"countryOfProduction"`
			Producttypename            string   `json:"productTypeName"`
			Netquantity                string   `json:"netQuantity"`
			Importedby                 string   `json:"importedBy"`
			Importeddate               string   `json:"importedDate"`
			Manufactureddate           string   `json:"manufacturedDate"`
			Manufacturedby             string   `json:"manufacturedBy"`
			Customercare               string   `json:"customerCare"`
			Disclaimer                 string   `json:"disclaimer"`
			Articlenumber              string   `json:"articleNumber"`
		} `json:"values"`
	} `json:"productAttributes"`
}

type parseProductResponse struct {
	Alternate          string `json:"alternate"`
	Articlecode        string `json:"articleCode"`
	Baseproductcode    string `json:"baseProductCode"`
	Categoryparentkey  string `json:"categoryParentKey"`
	Productkey         string `json:"productKey"`
	Collection         string `json:"collection"`
	Designercollection string `json:"designerCollection"`
	Producttype        string `json:"productType"`
	Agegender          string `json:"ageGender"`
	Presentationtypes  string `json:"presentationTypes"`
	Materialdetails    []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"materialDetails"`
	Articles map[string]*Article
}

func DecodeResponse(respBody string) (*parseProductResponse, error) {
	viewData := parseProductResponse{Articles: map[string]*Article{}}

	ret := map[string]json.RawMessage{}
	if err := json.Unmarshal([]byte(respBody), &ret); err != nil {
		return nil, err
	}

	for key, msg := range ret {
		rawData, _ := msg.MarshalJSON()
		if regexp.MustCompile(`[0-9]+`).MatchString(key) {
			var (
				rawData, _ = msg.MarshalJSON()
				article    Article
			)
			if err := json.Unmarshal(rawData, &article); err != nil {
				fmt.Println(err)
				continue
			}
			viewData.Articles[key] = &article
		} else if key == "productKey" {
			viewData.Productkey = strings.Trim(fmt.Sprintf("%s", rawData), `"`)
		} else if key == "articleCode" {
			viewData.Articlecode = strings.Trim(fmt.Sprintf("%s", rawData), `"`)
		} else if key == "baseProductCode" {
			viewData.Baseproductcode = strings.Trim(fmt.Sprintf("%s", rawData), `"`)
		}
	}
	return &viewData, nil
}

type sizeVariation struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// used to extract embaded json data in website page.
// more about golang regulation see here https://golang.org/pkg/regexp/syntax/
var categoryExtractReg = regexp.MustCompile(`(?U)var viewObjectsJson\s*=\s*({.*});`)
var categoryAPIExtractReg = regexp.MustCompile(`({.*})`)

var productsExtractReg = regexp.MustCompile(`(?Ums)var\s*productArticleDetails\s*=\s*({.*});\s*`)
var breadcrumbs = regexp.MustCompile(`"Categories":(\[.*\]),`)
var reviewReg = regexp.MustCompile(`(?U)<script type="application/ld\+json">\s*({.*})\s*</script>`)

var prodpaginationExtraReg = regexp.MustCompile(`(?U)var bcSfFilterMainConfig\s*=\s*({.*});`)
var prodDataExtraReg = regexp.MustCompile(`({.*})`)
var detailReg = regexp.MustCompile(`(?U)Afterpay\.products\.push\(({.*})\)|(?U)let apProduct\s*=\s*({.*});`)

func main() {

	respBody, _ := ioutil.ReadFile("C:\\Rinal\\ServiceBasedPRojects\\VoilaWork_new\\VoilaCrawl\\Output.html")

	matched := productsExtractReg.FindSubmatch(TrimSpaceNewlineInString(respBody))
	if len(matched) <= 1 {

	}

	vm := otto.New()
	jsonStr := "var productArticleDetails = " + string(matched[1])
	vm.Set("isDesktop", true)
	_, err := vm.Run(jsonStr)
	fmt.Println(`err `, err)

	vm.Run(`var obj = JSON.stringify(productArticleDetails);`)
	value, err := vm.Get("obj")
	responseJS, _ := value.ToString()

	fmt.Println(responseJS)

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		fmt.Println(err)
	}

	viewData, err := DecodeResponse(responseJS)

	fmt.Println(strings.Split(viewData.Baseproductcode, "_")[0])

	sel := doc.Find(`#variationDropdown-size`).Find(`option`)
	for j := range sel.Nodes {
		node := sel.Eq(j)

		var sizeData sizeVariation
		if err := json.Unmarshal([]byte(node.AttrOr("data-label", "")), &sizeData); err != nil {
			fmt.Println(err)
		}

		if sizeData.Key != "size" {
			continue
		}
	}

	var p ProductCategoryStructure

	matched = categoryExtractReg.FindSubmatch(respBody)
	if len(matched) <= 1 {
		//c.logger.Debugf("%s", respBody)
		//return fmt.Errorf("extract products info from %s failed, error=%s", resp.Request.URL, err)
	}

	if err := json.Unmarshal(respBody, &p); err != nil {
		fmt.Println(err)
	}

	if len(p.Groups) > 0 {
		for _, g := range p.Groups {
			fmt.Println(len(g.Garments))

			for i, pc := range g.Garments {
				if pc.Type != "P" {
					continue
				}

				rawurl := pc.Colors[0].Linkanchor
				fmt.Println(`i: `, i, ` `, rawurl)

				// req, err := http.NewRequest(http.MethodGet, rawurl, nil)
				// if err != nil {
				// 	c.logger.Errorf("load http request of url %s failed, error=%s", rawurl, err)
				// 	return err
				// }

				// // set the index of the product crawled in the sub response
				// nctx := context.WithValue(ctx, "item.index", lastIndex)
				// // yield sub request
				// if err := yield(nctx, req); err != nil {
				// 	return err
				// }
			}
		}
	}

	//fmt.Println(seli.Eq(0).Html())

	if len(doc.Find(`.final_change_price`).Nodes) == 0 {
		fmt.Println(`nill`)
	}
	fmt.Println(doc.Find(`.products__count`).Text())
}
