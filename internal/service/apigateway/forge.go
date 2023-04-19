package apigateway

import (
	"strings"
)

// escapeJSONPointer escapes string per RFC 6901
// so it can be used as path in JSON patch operations
func escapeJSONPointer(path string) string {
	path = strings.Replace(path, "~", "~0", -1)
	path = strings.Replace(path, "/", "~1", -1)
	return path
}
