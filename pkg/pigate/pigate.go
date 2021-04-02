package pigate

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	pbProxy "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
	"google.golang.org/protobuf/encoding/protojson"
)

type PigateClient struct {
	httpClient *http.Client
	addr       string
	logger     glog.Log
}

func NewPigateClient(addr string, logger glog.Log) (*PigateClient, error) {
	client := PigateClient{
		httpClient: &http.Client{},
		addr:       addr,
		logger:     logger.New("PigateClient"),
	}
	return &client, nil
}

func (client *PigateClient) Do(ctx context.Context, r *pbProxy.Request) (*pbProxy.Response, error) {
	if client == nil {
		return nil, nil
	}
	route := "/"

	if r == nil {
		return nil, pbError.ErrInvalidArgument.New("invalid request data")
	}

	reqBody, _ := protojson.Marshal(r)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s%s", client.addr, route), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var proxyResp pbProxy.Response
	if err := protojson.Unmarshal(respData, &proxyResp); err != nil {
		return nil, err
	}
	return &proxyResp, nil
}
