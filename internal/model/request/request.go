package request

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/voiladev/VoilaCrawl/pkg/types"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	"github.com/voiladev/go-framework/randutil"
)

type Request struct {
	types.Request
}

func NewRequest(req interface{}) (*Request, error) {
	if req == nil {
		return nil, errors.New("invalid request")
	}

	r := Request{}
	switch i := req.(type) {
	case *pbCrawl.Command_Request:
		r.TracingId = i.GetTracingId()
		r.JobId = i.GetJobId()
		r.StoreId = i.GetStoreId()
		r.ParentId = i.GetParent().GetReqId()
		r.Method = i.GetMethod()
		r.Url = i.GetUrl()
		r.Body = i.GetBody()
		if len(i.GetCustomHeaders()) > 0 {
			if data, err := json.Marshal(i.GetCustomHeaders()); err != nil {
				return nil, err
			} else {
				r.CustomHeaders = string(data)
			}
		}
		if len(i.GetCustomCookies()) > 0 {
			if data, err := json.Marshal(i.GetCustomCookies()); err != nil {
				return nil, err
			} else {
				r.CustomCookies = string(data)
			}
		}
		if len(i.GetSharingData()) > 0 {
			if data, err := json.Marshal(i.GetSharingData()); err != nil {
				return nil, err
			} else {
				r.SharingData = string(data)
			}
		}
		r.Options = &types.RequestOptions{
			DisableProxy:     i.GetOptions().GetDisableProxy(),
			MaxTtlPerRequest: i.GetOptions().GetMaxTtlPerRequest(),
			MaxRetryCount:    i.GetOptions().GetMaxRetryCount(),
			MaxRequestDepth:  i.GetOptions().GetMaxRequestDepth(),
		}
	case *pbCrawl.FetchRequest:
		r.TracingId = randutil.MustNewRandomID()
		r.JobId = i.GetJobId()
		r.StoreId = i.GetStoreId()
		r.Method = i.GetMethod()
		r.Url = i.GetUrl()
		r.Body = i.GetBody()
		if len(i.GetCustomHeaders()) > 0 {
			if data, err := json.Marshal(i.GetCustomHeaders()); err != nil {
				return nil, err
			} else {
				r.CustomHeaders = string(data)
			}
		}
		if len(i.GetCustomCookies()) > 0 {
			if data, err := json.Marshal(i.GetCustomCookies()); err != nil {
				return nil, err
			} else {
				r.CustomCookies = string(data)
			}
		}
		r.Options = &types.RequestOptions{
			DisableProxy:     i.GetOptions().GetDisableProxy(),
			MaxTtlPerRequest: i.GetOptions().GetMaxTtlPerRequest(),
			MaxRetryCount:    i.GetOptions().GetMaxRetryCount(),
			MaxRequestDepth:  i.GetOptions().GetMaxRequestDepth(),
		}
	default:
		return nil, errors.New("unsupported request load type")
	}

	if r.GetTracingId() == "" {
		r.TracingId = r.GetJobId()
	}
	if r.Options.MaxRequestDepth <= 0 {
		r.Options.MaxRequestDepth = 6
	}
	if r.Options.MaxRetryCount <= 0 {
		r.Options.MaxRetryCount = 3
	}
	if r.Options.MaxTtlPerRequest == 0 {
		// 5mins for one request
		r.Options.MaxTtlPerRequest = 5 * 60
	}
	return &r, nil
}

func (r *Request) Validate() error {
	if r == nil {
		return errors.New("nil request")
	}

	if r.GetTracingId() == "" {
		return errors.New("invalid tracing id")
	}
	if r.GetJobId() == "" {
		return errors.New("invalid request job id")
	}
	if r.GetStoreId() == "" {
		return errors.New("invalid store id")
	}
	if r.GetMethod() != http.MethodGet &&
		r.GetMethod() != http.MethodPost &&
		r.GetMethod() != http.MethodPut {
		return errors.New("unsupported http method")
	}
	if _, err := url.Parse(r.GetUrl()); err != nil {
		return err
	}
	return nil
}

func (r *Request) Unmarshal(ret interface{}) error {
	if r == nil {
		return errors.New("empty")
	}
	if ret == nil {
		return nil
	}

	switch val := ret.(type) {
	case *pbCrawl.Command_Request:
		val.TracingId = r.GetTracingId()
		val.JobId = r.GetJobId()
		val.ReqId = r.GetId()
		val.StoreId = r.GetStoreId()
		val.Method = r.GetMethod()
		val.Url = r.GetUrl()
		val.Body = r.GetBody()
		if r.GetCustomHeaders() != "" {
			if err := json.Unmarshal([]byte(r.GetCustomHeaders()), &val.CustomHeaders); err != nil {
				return err
			}
		}
		if r.GetCustomCookies() != "" {
			if err := json.Unmarshal([]byte(r.GetCustomCookies()), &val.CustomCookies); err != nil {
				return err
			}
		}
		if r.GetSharingData() != "" {
			if err := json.Unmarshal([]byte(r.GetSharingData()), &val.SharingData); err != nil {
				return err
			}
		}
		val.Options = &pbCrawl.Command_Request_Options{
			DisableProxy:     r.GetOptions().GetDisableProxy(),
			MaxTtlPerRequest: r.GetOptions().GetMaxTtlPerRequest(),
			MaxRetryCount:    r.GetOptions().GetMaxRetryCount(),
			MaxRequestDepth:  r.GetOptions().GetMaxRequestDepth(),
		}
	default:
		return errors.New("unsupported unmarshal type")
	}
	return nil
}
