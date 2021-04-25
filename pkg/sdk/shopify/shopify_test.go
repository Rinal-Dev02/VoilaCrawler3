package shopify

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

var client *ShopifyClient

func init() {
	var err error
	if client, err = New("jinglimited", "2021-04", os.Getenv("JING_API_KEY"), os.Getenv("JING_API_SECRET"), os.Getenv("JING_API_ACCESSTOKEN")); err != nil {
		panic(err)
	}
}

func TestShopifyClient_ListCustomCollections(t *testing.T) {
	type args struct {
		ctx context.Context
		req *ListCustomCollectionsRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *ListCustomCollectionsResponse
		wantErr bool
	}{
		{
			name: "custom_collections",
			args: args{
				ctx: context.Background(),
				req: &ListCustomCollectionsRequest{Limit: 50},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.ListCustomCollections(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ShopifyClient.ListCustomCollections() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			d, _ := json.Marshal(got)
			t.Logf("%s", d)
		})
	}
}

func TestShopifyClient_ListSmartCollections(t *testing.T) {
	type args struct {
		ctx context.Context
		req *ListSmartCollectionsRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *ListSmartCollectionsResponse
		wantErr bool
	}{
		{
			name: "smart_collections",
			args: args{
				ctx: context.Background(),
				req: &ListSmartCollectionsRequest{Limit: 50},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.ListSmartCollections(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ShopifyClient.ListSmartCollections() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			data, _ := json.Marshal(got)
			t.Logf("%s", data)
		})
	}
}

func TestShopifyClient_ListProducts(t *testing.T) {
	type args struct {
		ctx context.Context
		req *ListProductsRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *ListProductsResponse
		wantErr bool
	}{
		{
			name: "products",
			args: args{
				ctx: context.Background(),
				req: &ListProductsRequest{Limit: 50, IDs: []string{}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.ListProducts(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ShopifyClient.ListProducts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			data, _ := json.Marshal(got)
			t.Logf("%s", data)

			if got.NextLink != "" {
				tt.args.req.Link = got.NextLink
				got, err := client.ListProducts(tt.args.ctx, tt.args.req)
				if (err != nil) != tt.wantErr {
					t.Errorf("ShopifyClient.ListProducts() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				data, _ := json.Marshal(got)
				t.Logf("%s", data)
			}
		})
	}
}
