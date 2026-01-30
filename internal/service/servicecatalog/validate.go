// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func validSharePrincipal(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	// either account ID, or organization or organization unit

	wsAccount, errorsAccount := verify.ValidAccountID(v, k)

	if len(errorsAccount) == 0 {
		return wsAccount, errorsAccount
	}

	// Match AWS API validation: account ID, org/OU ARN, or org/OU ID
	// Org ID: o-[a-z0-9]{10,32}
	// OU ID: ou-[0-9a-z]{4,32}-[0-9a-z]{4,32}
	pattern := `^(arn:[\w-]+:organizations::\d{12}:(organization/o-[a-z0-9]{10,32}|ou/o-[a-z0-9]{10,32}/ou-[0-9a-z]{4,32}-[0-9a-z]{4,32})|o-[a-z0-9]{10,32}|ou-[0-9a-z]{4,32}-[0-9a-z]{4,32})$`
	if regexache.MustCompile(pattern).MatchString(value) {
		return nil, nil
	}

	errors = append(errors, fmt.Errorf("%q must be a valid account ID, organization ARN/ID, or organizational unit ARN/ID: %q", k, value))

	return ws, errors
}
