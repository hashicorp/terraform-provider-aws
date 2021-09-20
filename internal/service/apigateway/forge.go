package apigateway

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func buildInvokeURL(client *conns.AWSClient, restApiId, stageName string) string {
	hostname := client.RegionalHostname(fmt.Sprintf("%s.execute-api", restApiId))
	return fmt.Sprintf("https://%s/%s", hostname, stageName)
}

// escapeJSONPointer escapes string per RFC 6901
// so it can be used as path in JSON patch operations
func escapeJSONPointer(path string) string {
	path = strings.Replace(path, "~", "~0", -1)
	path = strings.Replace(path, "/", "~1", -1)
	return path
}
