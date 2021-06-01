package context

import (
	"context"
	"strings"

	pbItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	"github.com/voiladev/go-framework/protoutil"
	"google.golang.org/protobuf/proto"
)

type shareValue struct {
	Name string
}

var (
	TracingIdKey  = &shareValue{}
	JobIdKey      = &shareValue{}
	ReqIdKey      = &shareValue{}
	StoreIdKey    = &shareValue{}
	TargetTypeKey = &shareValue{}
)

var defaultTargetType = protoutil.GetTypeUrl(&pbItem.Product{})

func IsTargetTypeSupported(ctx context.Context, msgs ...proto.Message) bool {
	if ctx == nil {
		return false
	}

	typUrls := map[string]struct{}{}
	for _, msg := range msgs {
		typUrls[protoutil.GetTypeUrl(msg)] = struct{}{}
	}
	val := ctx.Value(TargetTypeKey)
	if val == nil {
		if _, ok := typUrls[defaultTargetType]; ok {
			return true
		}
		return false
	}

	typeStr, _ := val.(string)
	typs := strings.Split(typeStr, ",")
	for _, t := range typs {
		if _, ok := typUrls[t]; ok {
			return true
		}
	}
	return false
}

func IsAllTargetTypeSupported(ctx context.Context, msgs ...proto.Message) bool {
	if ctx == nil || len(msgs) == 0 {
		return false
	}

	typUrls := map[string]struct{}{}
	for _, msg := range msgs {
		typUrls[protoutil.GetTypeUrl(msg)] = struct{}{}
	}

	val := ctx.Value(TargetTypeKey)
	if val == nil {
		return false
	}

	typeStr, _ := val.(string)
	typs := strings.Split(typeStr, ",")
	for _, t := range typs {
		if _, ok := typUrls[t]; !ok {
			return false
		}
	}
	return true
}
