package cookie

import (
	"crypto/md5"
	"errors"
	"fmt"
	"time"

	"github.com/voiladev/VoilaCrawl/pkg/types"
	pbHttp "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/http"
)

type Cookie struct {
	types.Cookie
}

func New(tracingId string, i *pbHttp.Cookie) (*Cookie, error) {
	if i == nil {
		return nil, errors.New("invalid cookie")
	}

	c := Cookie{}
	c.TracingId = tracingId
	c.Name = i.Name
	c.Value = i.Value
	c.Domain = i.Domain
	c.Path = i.Path
	if c.Path == "" {
		c.Path = "/"
	}
	c.Expires = i.Expires
	c.HttpOnly = i.HttpOnly
	c.Session = i.Session
	c.SameSite = i.SameSite
	c.Priority = i.Priority

	if c.GetName() == "" {
		return nil, errors.New("invalid cookie name")
	}
	if c.GetDomain() == "" {
		return nil, errors.New("invalid cookie domain")
	}
	if c.GetExpires() > 0 && c.GetExpires() < time.Now().Unix() {
		return nil, errors.New("invalid cookie, expired")
	}
	c.Id = fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s-%s-%s-%s", c.TracingId, c.Domain, c.Name, c.Path))))

	return &c, nil
}

func (c *Cookie) Validate() error {
	if c == nil {
		return errors.New("empty")
	}

	if c.GetId() == "" {
		return errors.New("invalid id")
	}

	if c.GetName() == "" {
		return errors.New("invalid cookie name")
	}
	if c.GetDomain() == "" {
		return errors.New("invalid cookie domain")
	}
	if c.GetExpires() > 0 && c.GetExpires() < time.Now().Unix() {
		return errors.New("invalid cookie, expired")
	}
	return nil
}

func (c *Cookie) Unmarshal(ret interface{}) error {
	if c == nil || ret == nil {
		return nil
	}

	switch val := ret.(type) {
	case *pbHttp.Cookie:
		val.Name = c.Name
		val.Value = c.Value
		val.Domain = c.Domain
		val.Path = c.Path
		val.Expires = c.Expires
		val.Size = int32(len(val.Value))
		val.HttpOnly = c.HttpOnly
		val.Session = c.Session
		val.SameSite = c.SameSite
		val.Priority = c.Priority
	default:
		return errors.New("unsupported unmarshal type")
	}
	return nil
}
