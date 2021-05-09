package item

import (
	"sync"

	"github.com/voiladev/go-framework/protoutil"
	"google.golang.org/protobuf/proto"
)

var (
	registeredItemTypes = map[string]bool{}
	lock                sync.RWMutex
)

func Register(msg proto.Message) {
	if msg == nil {
		return
	}
	typeUrl := protoutil.GetTypeUrl(msg)

	lock.Lock()
	defer lock.Unlock()
	registeredItemTypes[typeUrl] = true
}

func IsRegistered(msg proto.Message) bool {
	if msg == nil {
		return false
	}

	typeUrl := protoutil.GetTypeUrl(msg)
	lock.RLock()
	defer lock.RUnlock()

	if val, _ := registeredItemTypes[typeUrl]; val {
		return true
	}
	return false
}

func SupportedTypes() []string {
	var ret []string
	for key := range registeredItemTypes {
		ret = append(ret, key)
	}
	return ret
}
