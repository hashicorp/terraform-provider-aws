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

	wsARN, errorsARN := verify.ValidARN(v, k)

	pattern := `(^arn:[\w-]+:organizations::\d+:(ou|organization)/)?(ou|o)-.+`
	if regexache.MustCompile(pattern).MatchString(value) {
		// Valid organization or OU (ARN or ID format)
		return wsARN, nil
	}

	// If we get here, it's not a valid account ID, ARN, or org/OU
	ws = append(ws, wsARN...)
	errors = append(errors, errorsARN...)
	errors = append(errors, fmt.Errorf("%q does not look like an OU or organization: %q", k, value))
	errors = append(errors, errorsAccount...)

	return ws, errors
}
