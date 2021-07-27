package crawler

import (
	ctxutil "github.com/voiladev/VoilaCrawler/pkg/context"
)

var (
	TracingIdKey  = ctxutil.TracingIdKey
	JobIdKey      = ctxutil.JobIdKey
	ReqIdKey      = ctxutil.ReqIdKey
	SiteIdKey     = ctxutil.SiteIdKey
	TargetTypeKey = ctxutil.TargetTypeKey

	IsTargetTypeSupported    = ctxutil.IsTargetTypeSupported
	IsAllTargetTypeSupported = ctxutil.IsAllTargetTypeSupported
)

var (
	MainCategoryKey = "MainCategory"
	CategoryKey     = "Category"
	SubCategoryKey  = "SubCategory"
	SubCategory2Key = "SubCategory2"
	SubCategory3Key = "SubCategory3"
	SubCategory4Key = "SubCategory4"
)
