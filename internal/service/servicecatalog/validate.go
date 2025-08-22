// Copyright (c) HashiCorp, Inc.
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
	ws = append(ws, wsARN...)
	errors = append(errors, errorsARN...)

	pattern := `^arn:[\w-]+:organizations:.*:(ou|organization)/`
	if !regexache.MustCompile(pattern).MatchString(value) {
		errors = append(errors, fmt.Errorf("%q does not look like an OU or organization: %q", k, value))
	}

	if len(errors) > 0 {
		errors = append(errors, errorsAccount...)
	}

	return ws, errors
}
