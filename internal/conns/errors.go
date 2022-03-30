package conns

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ClassicError(service, performingAction, resourceType, identifier string, gotError error) error {
	hf, err := names.FullHumanFriendly(service)

	if err != nil {
		return fmt.Errorf("finding human-friendly name for service (%s) while creating error (%s, %s, %s, %s): %w", service, performingAction, resourceType, identifier, gotError, err)
	}

	return fmt.Errorf("%s %s %s (%s): %w", performingAction, hf, resourceType, identifier, gotError)
}

func Error(service, performingAction, resourceType, identifier string, gotError error) diag.Diagnostics {
	hf, err := names.FullHumanFriendly(service)

	if err != nil {
		return diag.Errorf("finding human-friendly name for service (%s) while creating error (%s, %s, %s, %s): %s", service, performingAction, resourceType, identifier, gotError, err)
	}

	return diag.Errorf("%s %s %s (%s): %s", performingAction, hf, resourceType, identifier, gotError)
}
