package extract

import (
	"fmt"
	"sync"
)

type Endpoint struct {
	ApiVersion        string
	Url               string
	HttpMethod        string
	IsManagementPlane bool
}

func (endpoint Endpoint) String() string {
	return fmt.Sprintf("%s %s %s %t", endpoint.ApiVersion, endpoint.Url, endpoint.HttpMethod, endpoint.IsManagementPlane)
}

type EndpointList struct {
	endpointsMap map[Endpoint]struct{}
	mutex        sync.RWMutex
	once         sync.Once
}

func (list *EndpointList) init() {
	list.endpointsMap = map[Endpoint]struct{}{}
}

func (list *EndpointList) Add(endpoint Endpoint) bool {
	list.once.Do(list.init)

	list.mutex.RLock()
	_, ok := list.endpointsMap[endpoint]
	list.mutex.RUnlock()
	if !ok {
		list.mutex.Lock()
		list.endpointsMap[endpoint] = struct{}{}
		list.mutex.Unlock()
		return true
	}
	return false
}
