package extract

import (
	"regexp"
	"strings"
)

var re = regexp.MustCompile(`/\{.*?\}`)

func formatUrl(url string) string {
	url = re.ReplaceAllString(url, "")
	if !strings.HasSuffix(url, "/") {
		url = url + "/"
	}
	return strings.ToUpper(url)
}
