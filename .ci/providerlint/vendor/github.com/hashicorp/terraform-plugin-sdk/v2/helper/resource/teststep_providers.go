package resource

import (
	"context"
	"fmt"
	"strings"
)

// providerConfig takes the list of providers in a TestStep and returns a
// config with only empty provider blocks. This is useful for Import, where no
// config is provided, but the providers must be defined.
func (s TestStep) providerConfig(_ context.Context) string {
	var providerBlocks, requiredProviderBlocks strings.Builder

	for name, externalProvider := range s.ExternalProviders {
		providerBlocks.WriteString(fmt.Sprintf("provider %q {}\n", name))

		if externalProvider.Source == "" && externalProvider.VersionConstraint == "" {
			continue
		}

		requiredProviderBlocks.WriteString(fmt.Sprintf("    %s = {\n", name))

		if externalProvider.Source != "" {
			requiredProviderBlocks.WriteString(fmt.Sprintf("      source = %q\n", externalProvider.Source))
		}

		if externalProvider.VersionConstraint != "" {
			requiredProviderBlocks.WriteString(fmt.Sprintf("      version = %q\n", externalProvider.VersionConstraint))
		}

		requiredProviderBlocks.WriteString("    }\n")
	}

	if requiredProviderBlocks.Len() > 0 {
		return fmt.Sprintf(`
terraform {
  required_providers {
%[1]s
  }
}

%[2]s
`, strings.TrimSuffix(requiredProviderBlocks.String(), "\n"), providerBlocks.String())
	}

	return providerBlocks.String()
}
