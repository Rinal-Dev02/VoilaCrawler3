package crawler

import (
	ctxutil "github.com/voiladev/go-crawler/pkg/context"
)

var (
	TracingIdKey  = ctxutil.TracingIdKey
	JobIdKey      = ctxutil.JobIdKey
	ReqIdKey      = ctxutil.ReqIdKey
	StoreIdKey    = ctxutil.StoreIdKey
	TargetTypeKey = ctxutil.TargetTypeKey

	IsTargetTypeSupported    = ctxutil.IsTargetTypeSupported
	IsAllTargetTypeSupported = ctxutil.IsAllTargetTypeSupported
)
