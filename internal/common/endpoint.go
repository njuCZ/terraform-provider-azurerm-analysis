package common

import (
	"fmt"
	"strconv"
	"strings"
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

func (endpoint Endpoint) GetFullResourceName() string {
	fullResourceName := "Unknown"
	if !endpoint.IsManagementPlane {
		return fullResourceName
	}
	if index := strings.LastIndex(endpoint.Url, "/PROVIDERS/"); index != -1 {
		fullResourceName = endpoint.Url[index+len("/PROVIDERS/"):]
		fullResourceName = strings.TrimRight(fullResourceName, "/")
	}
	return fullResourceName
}

func ConstructEndpoint(str string) (*Endpoint, error) {
	segments := strings.Split(str, " ")
	if len(segments) != 4 {
		return nil, fmt.Errorf("contruct endpoint error: input is %s", str)
	}
	e := Endpoint{
		ApiVersion: segments[0],
		Url:        segments[1],
		HttpMethod: segments[2],
	}
	if isManagementPlane, err := strconv.ParseBool(segments[3]); err != nil {
		return nil, fmt.Errorf("contruct endpoint error: input is %s", str)
	} else {
		e.IsManagementPlane = isManagementPlane
	}
	return &e, nil
}

func ConstructEndpointList(multiLine string) ([]Endpoint, error) {
	multiLine = strings.Trim(multiLine, "\n")
	lines := strings.Split(multiLine, "\n")

	result := make([]Endpoint, len(lines))
	for index, line := range lines {
		endpoint, err := ConstructEndpoint(line)
		if err == nil {
			return nil, err
		}
		if endpoint == nil {
			return nil, fmt.Errorf("should not run here")
		}
		result[index] = *endpoint
	}
	return result, nil
}
