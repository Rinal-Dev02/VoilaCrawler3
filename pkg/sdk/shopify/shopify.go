package shopify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type ShopifyClient struct {
	shop        string
	apiVer      string
	apiKey      string
	apiSecret   string
	accessToken string

	httpClient *http.Client
}

// New for api version, see https://shopify.dev/docs/admin-api/rest/reference/products/collect#index-2021-04
func New(shop string, apiVer, apiKey, apiSecret, accessToken string) (*ShopifyClient, error) {
	if shop == "" {
		return nil, errors.New("invalid store name")
	}
	if apiVer == "" {
		apiVer = "2021-04"
	}
	if apiKey == "" {
		return nil, errors.New("invalid app key")
	}
	if accessToken == "" {
		return nil, errors.New("invalid access token")
	}

	client := ShopifyClient{
		shop:        shop,
		apiVer:      apiVer,
		apiKey:      apiKey,
		apiSecret:   apiSecret,
		accessToken: accessToken,
		httpClient:  &http.Client{},
	}
	return &client, nil
}

func (client *ShopifyClient) auth(req *http.Request) *http.Request {
	if client == nil {
		return req
	}
	req.Header.Set("X-Shopify-Access-Token", client.accessToken)

	return req
}

func (client *ShopifyClient) nextLink(resp *http.Response) (string, string) {
	if resp == nil {
		return "", ""
	}

	var (
		preLink  string
		nextLink string
	)
	if vals, _ := resp.Header["Link"]; len(vals) > 0 {
		for _, val := range vals {
			if strings.Contains(val, `rel="next"`) || strings.Contains(val, `rel=next`) {
				nextLink = strings.TrimSpace(strings.SplitN(strings.TrimPrefix(val, "<"), ">", 2)[0])
			} else if strings.Contains(val, `rel="previous"`) || strings.Contains(val, `rel=previous`) {
				preLink = strings.TrimSpace(strings.SplitN(strings.TrimPrefix(val, "<"), ">", 2)[0])
			}
		}
	}
	return preLink, nextLink
}

type ListCollectsRequest struct {
	Limit        int32
	ProductID    string
	CollectionID string

	SinceId string
	Fields  []string
	Link    string
}

type ListCollectsResponse struct {
	Collects []struct {
		ID           uint64      `json:"id"`
		CollectionID uint64      `json:"collection_id"`
		ProductID    uint64      `json:"product_id"`
		CreatedAt    interface{} `json:"created_at"`
		UpdatedAt    interface{} `json:"updated_at"`
		Position     int         `json:"position"`
		SortValue    string      `json:"sort_value"`
	} `json:"collects"`
	NextLink string
}

func (client *ShopifyClient) ListCollects(ctx context.Context, req *ListCollectsRequest) (*ListCollectsResponse, error) {
	if client == nil {
		return nil, nil
	}

	if req == nil {
		return nil, errors.New("invalid request params")
	}
	if req.Limit == 0 {
		req.Limit = 250
	}

	rawurl := req.Link
	if rawurl == "" {
		u, _ := url.Parse(fmt.Sprintf("https://%s.myshopify.com/admin/api/%s/collects.json", client.shop, client.apiVer))
		vals := u.Query()
		if req.Limit != 0 {
			vals.Set("limit", fmt.Sprintf("%d", req.Limit))
		}
		if req.ProductID != "" {
			vals.Set("product_id", req.ProductID)
		}
		if req.CollectionID != "" {
			vals.Set("collection_id", req.CollectionID)
		}
		if req.SinceId != "" {
			vals.Set("since_id", req.SinceId)
		}
		if len(req.Fields) != 0 {
			vals.Set("fields", strings.Join(req.Fields, ","))
		}
		u.RawQuery = vals.Encode()
		rawurl = u.String()
	}
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, rawurl, nil)
	if err != nil {
		return nil, err
	}
	r = client.auth(r)

	resp, err := client.httpClient.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("access %s failed, %s", r.URL, respBody)
	}

	var ret ListCollectsResponse
	if json.Unmarshal(respBody, &ret); err != nil {
		return nil, err
	}
	_, ret.NextLink = client.nextLink(resp)

	return &ret, nil
}

type Collection struct {
	ID                uint64      `json:"id"`
	Handle            string      `json:"handle"`
	UpdatedAt         string      `json:"updated_at"`
	PublishedAt       string      `json:"published_at"`
	SortOrder         string      `json:"sort_order"`
	TemplateSuffix    interface{} `json:"template_suffix"`
	PublishedScope    string      `json:"published_scope"`
	Title             string      `json:"title"`
	BodyHTML          string      `json:"body_html"`
	AdminGraphqlAPIID string      `json:"admin_graphql_api_id"`
	Image             struct {
		CreatedAt string `json:"created_at"`
		Alt       string `json:"alt"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
		Src       string `json:"src"`
	} `json:"image,omitempty"`
	Disjunctive bool `json:"disjunctive"`
	Rules       []struct {
		Column    string `json:"column"`
		Relation  string `json:"relation"`
		Condition string `json:"condition"`
	} `json:"rules"`
}

type GetCollectionRequest struct {
	ID string
}

type GetCollectionResponse struct {
	Collection *Collection
}

func (client *ShopifyClient) GetCollection(ctx context.Context, req *GetCollectionRequest) (*GetCollectionResponse, error) {
	if client == nil {
		return nil, nil
	}

	if req == nil {
		return nil, errors.New("invalid request params")
	}
	if req.ID == "" {
		return nil, errors.New("invalid collection id")
	}

	rawurl := fmt.Sprintf("https://%s.myshopify.com/admin/api/%s/collections/%s.json", client.shop, client.apiVer, req.ID)
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, rawurl, nil)
	if err != nil {
		return nil, err
	}
	r = client.auth(r)

	resp, err := client.httpClient.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("access %s failed, %s", r.URL, respBody)
	}

	var ret GetCollectionResponse
	if json.Unmarshal(respBody, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}

type ListCustomCollectionsRequest struct {
	Limit           int
	IDs             []string
	SinceId         string
	Title           string
	ProductId       string
	Handle          string
	UpdatedAtMin    string
	UpdatedAtMax    string
	PublishedAtMin  string
	PublishedAtMax  string
	PublishedStatus PublishStatus
	Fields          []string
	Link            string
}

type ListCustomCollectionsResponse struct {
	CustomCollections []*Collection `json:"custom_collections"`
	NextLink          string
}

func (client *ShopifyClient) ListCustomCollections(ctx context.Context, req *ListCustomCollectionsRequest) (*ListCustomCollectionsResponse, error) {
	if client == nil {
		return nil, nil
	}

	if req == nil {
		return nil, errors.New("invalid request params")
	}
	if req.Limit == 0 {
		req.Limit = 250
	}

	rawurl := req.Link
	if rawurl == "" {
		u, _ := url.Parse(fmt.Sprintf("https://%s.myshopify.com/admin/api/%s/custom_collections.json", client.shop, client.apiVer))
		vals := u.Query()
		if req.Limit != 0 {
			vals.Set("limit", fmt.Sprintf("%d", req.Limit))
		}
		if len(req.IDs) > 0 {
			vals.Set("ids", strings.Join(req.IDs, ","))
		}
		if req.SinceId != "" {
			vals.Set("since_id", req.SinceId)
		}
		if req.Title != "" {
			vals.Set("title", req.Title)
		}
		if req.ProductId != "" {
			vals.Set("product_id", req.ProductId)
		}
		if req.Handle != "" {
			vals.Set("handle", req.Handle)
		}
		if req.UpdatedAtMin != "" {
			vals.Set("updated_at_min", req.UpdatedAtMin)
		}
		if req.UpdatedAtMax != "" {
			vals.Set("updated_at_max", req.UpdatedAtMax)
		}
		if req.PublishedAtMin != "" {
			vals.Set("published_at_min", req.PublishedAtMin)
		}
		if req.PublishedAtMax != "" {
			vals.Set("published_at_max", req.PublishedAtMax)
		}
		if req.PublishedStatus != "" {
			vals.Set("published_status", string(req.PublishedStatus))
		}
		if len(req.Fields) != 0 {
			vals.Set("fields", strings.Join(req.Fields, ","))
		}
		u.RawQuery = vals.Encode()
		rawurl = u.String()
	}
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, rawurl, nil)
	if err != nil {
		return nil, err
	}
	r = client.auth(r)

	resp, err := client.httpClient.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("access %s failed, %s", r.URL, respBody)
	}

	var ret ListCustomCollectionsResponse
	if json.Unmarshal(respBody, &ret); err != nil {
		return nil, err
	}
	_, ret.NextLink = client.nextLink(resp)

	return &ret, nil
}

type ListSmartCollectionsRequest struct {
	Limit           int
	IDs             []string
	SinceId         string
	Title           string
	ProductId       string
	Handle          string
	UpdatedAtMin    string
	UpdatedAtMax    string
	PublishedAtMin  string
	PublishedAtMax  string
	PublishedStatus PublishStatus
	Fields          []string
	Link            string
}

type ListSmartCollectionsResponse struct {
	SmartCollections []*Collection `json:"smart_collections"`
	NextLink         string
}

func (client *ShopifyClient) ListSmartCollections(ctx context.Context, req *ListSmartCollectionsRequest) (*ListSmartCollectionsResponse, error) {
	if client == nil {
		return nil, nil
	}

	if req == nil {
		return nil, errors.New("invalid request params")
	}
	if req.Limit == 0 {
		req.Limit = 250
	}

	rawurl := req.Link
	if rawurl == "" {
		u, _ := url.Parse(fmt.Sprintf("https://%s.myshopify.com/admin/api/%s/smart_collections.json", client.shop, client.apiVer))
		vals := u.Query()
		if req.Limit != 0 {
			vals.Set("limit", fmt.Sprintf("%d", req.Limit))
		}
		if req.SinceId != "" {
			vals.Set("since_id", req.SinceId)
		}
		if len(req.Fields) != 0 {
			vals.Set("fields", strings.Join(req.Fields, ","))
		}
		u.RawQuery = vals.Encode()
		rawurl = u.String()
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodGet, rawurl, nil)
	if err != nil {
		return nil, err
	}
	r = client.auth(r)

	resp, err := client.httpClient.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("access %s failed, %s", r.URL, respBody)
	}

	var ret ListSmartCollectionsResponse
	if json.Unmarshal(respBody, &ret); err != nil {
		return nil, err
	}
	_, ret.NextLink = client.nextLink(resp)

	return &ret, nil
}

type ProductStatus string

const (
	ProductStatusActive   ProductStatus = "active"
	ProductStatusarchived               = "archived"
	ProductStatusDraft                  = "draft"
)

type PublishStatus string

const (
	Published   PublishStatus = "published"
	Unpublished               = "unpublished"
)

type Product struct {
	ID                uint64 `json:"id"`
	Title             string `json:"title"`
	BodyHTML          string `json:"body_html"`
	Vendor            string `json:"vendor"`
	ProductType       string `json:"product_type"`
	CreatedAt         string `json:"created_at"`
	Handle            string `json:"handle"`
	UpdatedAt         string `json:"updated_at"`
	PublishedAt       string `json:"published_at"`
	TemplateSuffix    string `json:"template_suffix"`
	Status            string `json:"status,omitempty"`
	PublishedScope    string `json:"published_scope"`
	Tags              string `json:"tags"`
	AdminGraphqlAPIID string `json:"admin_graphql_api_id"`
	Variants          []struct {
		ID                   uint64  `json:"id"`
		ProductID            uint64  `json:"product_id"`
		Title                string  `json:"title"`
		Price                string  `json:"price"`
		Sku                  string  `json:"sku"`
		Position             int     `json:"position"`
		InventoryPolicy      string  `json:"inventory_policy"`
		CompareAtPrice       string  `json:"compare_at_price"`
		FulfillmentService   string  `json:"fulfillment_service"`
		InventoryManagement  string  `json:"inventory_management"`
		Option1              string  `json:"option1"`
		Option2              string  `json:"option2"`
		Option3              string  `json:"option3"`
		CreatedAt            string  `json:"created_at"`
		UpdatedAt            string  `json:"updated_at"`
		Taxable              bool    `json:"taxable"`
		Barcode              string  `json:"barcode"`
		Grams                int     `json:"grams"`
		ImageID              uint64  `json:"image_id"`
		Weight               float64 `json:"weight"`
		WeightUnit           string  `json:"weight_unit"`
		InventoryItemID      uint64  `json:"inventory_item_id"`
		InventoryQuantity    int     `json:"inventory_quantity"`
		OldInventoryQuantity int     `json:"old_inventory_quantity"`
		TaxCode              string  `json:"tax_code"`
		RequiresShipping     bool    `json:"requires_shipping"`
		AdminGraphqlAPIID    string  `json:"admin_graphql_api_id"`
	} `json:"variants"`
	Options []struct {
		ID        uint64   `json:"id"`
		ProductID uint64   `json:"product_id"`
		Name      string   `json:"name"`
		Position  int      `json:"position"`
		Values    []string `json:"values"`
	} `json:"options"`
	Images []struct {
		ID                uint64        `json:"id"`
		ProductID         uint64        `json:"product_id"`
		Position          int           `json:"position"`
		CreatedAt         string        `json:"created_at"`
		UpdatedAt         string        `json:"updated_at"`
		Alt               string        `json:"alt"`
		Width             int           `json:"width"`
		Height            int           `json:"height"`
		Src               string        `json:"src"`
		VariantIds        []interface{} `json:"variant_ids"`
		AdminGraphqlAPIID string        `json:"admin_graphql_api_id"`
	} `json:"images"`
	Image struct {
		ID                uint64        `json:"id"`
		ProductID         uint64        `json:"product_id"`
		Position          int           `json:"position"`
		CreatedAt         string        `json:"created_at"`
		UpdatedAt         string        `json:"updated_at"`
		Alt               interface{}   `json:"alt"`
		Width             int           `json:"width"`
		Height            int           `json:"height"`
		Src               string        `json:"src"`
		VariantIds        []interface{} `json:"variant_ids"`
		AdminGraphqlAPIID string        `json:"admin_graphql_api_id"`
	} `json:"image"`
}

type ListCollectionProductsRequest struct {
	Limit        int
	CollectionID string
	Link         string
}

type ListCollectionProductsResponse struct {
	Products []*Product `json:"products"`
	NextLink string
}

func (client *ShopifyClient) ListCollectionProducts(ctx context.Context, req *ListCollectionProductsRequest) (*ListCollectionProductsResponse, error) {
	if client == nil {
		return nil, nil
	}

	if req == nil {
		return nil, errors.New("invalid request params")
	}
	if req.CollectionID == "" || req.CollectionID == "0" {
		return nil, errors.New("invalid collection id")
	}
	if req.Limit == 0 {
		req.Limit = 250
	}

	rawurl := req.Link
	if rawurl == "" {
		u, _ := url.Parse(fmt.Sprintf("https://%s.myshopify.com/admin/api/%s/collections/%s/products.json", client.shop, client.apiVer, req.CollectionID))
		vals := u.Query()
		if req.Limit != 0 {
			vals.Set("limit", fmt.Sprintf("%d", req.Limit))
		}
		u.RawQuery = vals.Encode()
		rawurl = u.String()
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodGet, rawurl, nil)
	if err != nil {
		return nil, err
	}
	r = client.auth(r)

	resp, err := client.httpClient.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("access %s failed, %s", r.URL, respBody)
	}

	var ret ListCollectionProductsResponse
	if json.Unmarshal(respBody, &ret); err != nil {
		return nil, err
	}
	_, ret.NextLink = client.nextLink(resp)
	return &ret, nil
}

// ListProductsRequest referrer to https://shopify.dev/docs/admin-api/rest/reference/products/product#index-2021-04
type ListProductsRequest struct {
	IDs                   []string
	Limit                 int
	SinceId               string        // Return only products after the specified ID
	Title                 string        // return products by product title
	Vendor                string        // return products by product title
	Handle                []string      // Return only products specified by a comma-separated list of product handles
	ProductType           string        // Return products by product type
	Status                ProductStatus // Return only products specified by a comma-separated list of statuses.
	CollectionID          string        //
	CreatedAtMin          string        // Return products created after a specified date. (format: 2014-04-25T16:15:47-04:00)
	CreatedAtMax          string        // Return products created before a specified date. (format: 2014-04-25T16:15:47-04:00)
	UpdatedAtMin          string        // Return products last updated after a specified date. (format: 2014-04-25T16:15:47-04:00)
	UpdatedAtMax          string
	PublishedAtMin        string        // Return products published after a specified date. (format: 2014-04-25T16:15:47-04:00)
	PublishedAtMax        string        // Return products published before a specified date. (format: 2014-04-25T16:15:47-04:00)
	PublishedStatus       PublishStatus // Return products by their published status
	Fields                []string      // Return only certain fields specified by a comma-separated list of field names.
	PresentmentCurrencies []string      // Return presentment prices in only certain currencies, specified by a comma-separated list of ISO 4217 currency codes. https://en.wikipedia.org/wiki/ISO_4217
	Link                  string
}

type ListProductsResponse struct {
	Products []*Product `json:"products"`
	NextLink string
}

func (client *ShopifyClient) ListProducts(ctx context.Context, req *ListProductsRequest) (*ListProductsResponse, error) {
	if client == nil {
		return nil, nil
	}

	if req == nil {
		return nil, errors.New("invalid request params")
	}
	if req.Limit == 0 {
		req.Limit = 250
	}

	rawurl := req.Link
	if rawurl == "" {
		u, _ := url.Parse(fmt.Sprintf("https://%s.myshopify.com/admin/api/%s/products.json", client.shop, client.apiVer))
		vals := u.Query()
		if len(req.IDs) > 0 {
			vals.Set("ids", strings.Join(req.IDs, ","))
		}
		if req.Limit != 0 {
			vals.Set("limit", fmt.Sprintf("%d", req.Limit))
		}
		if req.SinceId != "" {
			vals.Set("since_id", req.SinceId)
		}
		if req.Title != "" {
			vals.Set("title", req.Title)
		}
		if req.Vendor != "" {
			vals.Set("vendor", req.Vendor)
		}
		if len(req.Handle) != 0 {
			vals.Set("handle", strings.Join(req.Handle, ","))
		}
		if req.ProductType != "" {
			vals.Set("product_type", req.ProductType)
		}
		if req.Status != "" {
			vals.Set("status", string(req.Status))
		}
		if req.CollectionID != "" {
			vals.Set("collection_id", req.CollectionID)
		}
		if req.CreatedAtMin != "" {
			vals.Set("created_at_min", req.CreatedAtMin)
		}
		if req.CreatedAtMax != "" {
			vals.Set("created_at_max", req.CreatedAtMax)
		}
		if req.UpdatedAtMin != "" {
			vals.Set("updated_at_min", req.UpdatedAtMin)
		}
		if req.UpdatedAtMax != "" {
			vals.Set("updated_at_max", req.UpdatedAtMax)
		}
		if req.PublishedAtMin != "" {
			vals.Set("published_at_min", req.PublishedAtMin)
		}
		if req.PublishedAtMax != "" {
			vals.Set("published_at_max", req.PublishedAtMax)
		}
		if req.PublishedStatus != "" {
			vals.Set("published_status", string(req.PublishedStatus))
		}
		if len(req.Fields) > 0 {
			vals.Set("fields", strings.Join(req.Fields, ","))
		}
		if len(req.PresentmentCurrencies) != 0 {
			vals.Set("presentment_currencies", strings.Join(req.PresentmentCurrencies, ","))
		}
		u.RawQuery = vals.Encode()
		rawurl = u.String()
	}
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, rawurl, nil)
	if err != nil {
		return nil, err
	}
	r = client.auth(r)

	resp, err := client.httpClient.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("access %s failed, %s", r.URL, respBody)
	}

	var ret ListProductsResponse
	if json.Unmarshal(respBody, &ret); err != nil {
		return nil, err
	}
	_, ret.NextLink = client.nextLink(resp)
	return &ret, nil
}

type GetProductRequest struct {
	ProductID string
}

type GetProductResponse struct {
	Product *Product
}

func (client *ShopifyClient) GetProduct(ctx context.Context, req *GetProductRequest) (*GetProductResponse, error) {
	if client == nil {
		return nil, nil
	}

	if req == nil {
		return nil, errors.New("invalid request params")
	}
	if req.ProductID == "" {
		return nil, errors.New("invalid product id")
	}

	rawurl := fmt.Sprintf("https://%s.myshopify.com/admin/api/%s/products/%s.json", client.shop, client.apiVer, req.ProductID)
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, rawurl, nil)
	if err != nil {
		return nil, err
	}
	r = client.auth(r)

	resp, err := client.httpClient.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("access %s failed, %s", r.URL, respBody)
	}

	var ret GetProductResponse
	if json.Unmarshal(respBody, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}
