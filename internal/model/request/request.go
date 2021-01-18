package request

import (
	"encoding/json"
	"errors"

	"github.com/voiladev/VoilaCrawl/pkg/types"
	pbHttp "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/http"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
)

type Request struct {
	types.Request
}

func NewRequest(i *pbCrawl.Command_Request) (*Request, error) {
	if i == nil {
		return nil, errors.New("invalid request")
	}

	r := Request{}
	r.TracingId = i.GetTracingId()
	r.JobId = i.GetJobId()
	r.ParentId = i.GetParent().GetReqId()
	r.Method = i.GetMethod()
	r.Url = i.GetUrl()
	r.Body = i.GetBody()
	i.GetCustomHeaders()
	if i.GetParent().GetUrl() != "" {
		found := false
		for _, header := range i.GetCustomHeaders() {
			if header.Key == "Referer" {
				header.Values = append(header.Values[0:0], i.GetParent().GetUrl())
				found = true
				break
			}
		}
		if !found {
			i.CustomHeaders = append(i.CustomHeaders, &pbHttp.Header{Key: "Referer", Values: []string{i.GetParent().GetUrl()}})
		}
	}
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
		EnableHeadless:   i.GetOptions().GetEnableHeadless(),
		MaxTtlPerRequest: i.GetOptions().GetMaxTtlPerRequest(),
		MaxRetryCount:    i.GetOptions().GetMaxRetryCount(),
		MaxRequestDepth:  i.GetOptions().GetMaxRequestDepth(),
	}

	return &r, nil
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
			EnableHeadless:   r.GetOptions().GetEnableHeadless(),
			MaxTtlPerRequest: r.GetOptions().GetMaxTtlPerRequest(),
			MaxRetryCount:    r.GetOptions().GetMaxRetryCount(),
			MaxRequestDepth:  r.GetOptions().GetMaxRequestDepth(),
		}
	default:
		return errors.New("unsupported unmarshal type")
	}
	return nil
}
