package server

import (
	"bytes"
	"github.com/njucz/terraform-provider-azurerm-analysis/internal/common"
	"time"
)

func getFirstDateOfCurrentMonth() string {
	year, month, _ := time.Now().Date()
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	return firstDay.Format("2006-01-02")
}

func diff(old, new []common.Endpoint) []common.Endpoint {
	oldMap := make(map[common.Endpoint]struct{}, len(old))
	for _, endpoint := range old {
		oldMap[endpoint] = struct{}{}
	}
	result := make([]common.Endpoint, 0)
	for _, endpoint := range new {
		if _, ok := oldMap[endpoint]; !ok {
			result = append(result, endpoint)
		}
	}
	return result
}

func endpointsToString(endpoints []common.Endpoint) string {
	buf := bytes.Buffer{}
	for _, endpoint := range endpoints {
		buf.WriteString(endpoint.String())
		buf.WriteString("\n")
	}
	return buf.String()
}
