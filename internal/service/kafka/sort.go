package kafka

import (
	"sort"
	"strings"
)

const (
	endpointSeparator = ","
)

// SortEndpointsString sorts a comma-separated list of endpoints.
func SortEndpointsString(s string) string {
	parts := strings.Split(s, endpointSeparator)
	sort.Strings(parts)
	return strings.Join(parts, endpointSeparator)
}
