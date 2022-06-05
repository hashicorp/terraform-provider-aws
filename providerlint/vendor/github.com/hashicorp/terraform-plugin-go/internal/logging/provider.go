package logging

import (
	"log"
	"strings"

	tfaddr "github.com/hashicorp/terraform-registry-address"
)

func ProviderLoggerName(providerAddress string) string {
	provider, err := tfaddr.ParseRawProviderSourceString(providerAddress)

	if err != nil {
		log.Printf("[ERROR] Error parsing provider name %q: %s", providerAddress, err)
		return ""
	}

	return strings.ReplaceAll(provider.Type, "-", "_")
}
