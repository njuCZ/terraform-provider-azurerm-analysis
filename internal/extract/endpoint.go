package extract

import (
	"fmt"
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
