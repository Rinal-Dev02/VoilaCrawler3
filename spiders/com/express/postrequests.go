package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func productRequest(productCode string, referURL string) ([]byte, error) {

	url := "https://www.express.com/graphql"
	method := "POST"

	payload := strings.NewReader(`{"operationName":"ProductQuery","variables":{"productId":"` + productCode + `"},"query":"query ProductQuery($productId: String!) {\n  product(id: $productId) {\n    bopisEligible\n    clearancePromoMessage\n    collection\n    crossRelDetailMessage\n    crossRelProductURL\n    EFOProduct\n    expressProductType\n    fabricCare\n    fabricDetailImages {\n      caption\n      image\n      __typename\n    }\n    gender\n    internationalShippingAvailable\n    listPrice\n    marketPlaceProduct\n    name\n    newProduct\n    onlineExclusive\n    onlineExclusivePromoMsg\n    productDescription {\n      type\n      content\n      __typename\n    }\n    productFeatures\n    productId\n    productImage\n    productInventory\n    productURL\n    promoMessage\n    recsAlgorithm\n    originRecsAlgorithm\n    salePrice\n    type\n    breadCrumbCategory {\n      categoryName\n      h1CategoryName\n      links {\n        rel\n        href\n        __typename\n      }\n      breadCrumbCategory {\n        categoryName\n        h1CategoryName\n        links {\n          rel\n          href\n          __typename\n        }\n        __typename\n      }\n      __typename\n    }\n    colorSlices {\n      color\n      defaultSlice\n      ipColorCode\n      hasWaistAndInseam\n      swatchURL\n      imageMap {\n        All {\n          LARGE\n          MAIN\n          __typename\n        }\n        Default {\n          LARGE\n          MAIN\n          __typename\n        }\n        Model1 {\n          LARGE\n          MAIN\n          __typename\n        }\n        Model2 {\n          LARGE\n          MAIN\n          __typename\n        }\n        Model3 {\n          LARGE\n          MAIN\n          __typename\n        }\n        __typename\n      }\n      onlineSkus\n      skus {\n        backOrderable\n        backOrderDate\n        displayMSRP\n        displayPrice\n        ext\n        inseam\n        inStoreInventoryCount\n        inventoryMessage\n        isFinalSale\n        isInStockOnline\n        miraklOffer {\n          minimumShippingPrice\n          sellerId\n          sellerName\n          __typename\n        }\n        marketPlaceSku\n        onClearance\n        onSale\n        onlineExclusive\n        onlineInventoryCount\n        size\n        sizeName\n        skuId\n        __typename\n      }\n      __typename\n    }\n    originRecs {\n      listPrice\n      marketPlaceProduct\n      name\n      productId\n      productImage\n      productURL\n      salePrice\n      __typename\n    }\n    relatedProducts {\n      listPrice\n      marketPlaceProduct\n      name\n      productId\n      productImage\n      productURL\n      salePrice\n      colorSlices {\n        color\n        defaultSlice\n        __typename\n      }\n      __typename\n    }\n    icons {\n      icon\n      category\n      __typename\n    }\n    __typename\n  }\n}\n"}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	req.Header.Add("accept", "*/*")
	req.Header.Add("accept-language", "en-GB,en-US;q=0.9,en;q=0.8")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("origin", "https://www.express.com")
	req.Header.Add("referer", referURL)
	req.Header.Add("sec-ch-ua", "\"Google Chrome\";v=\"89\", \"Chromium\";v=\"89\", \";Not A Brand\";v=\"99\"")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("sec-fetch-dest", "empty")
	req.Header.Add("sec-fetch-mode", "cors")
	req.Header.Add("sec-fetch-site", "same-origin")
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.114 Safari/537.36")
	req.Header.Add("x-exp-rvn-cacheable", "true")
	req.Header.Add("x-exp-rvn-query-classification", "product")
	req.Header.Add("x-exp-rvn-source", "app_express.com")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println(string(body))

	//ioutil.WriteFile("C:\\Rinal\\ServiceBasedPRojects\\VoilaWork_new\\VoilaCrawl\\Output_1.html", body, 0644)
	return body, err
}
