// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation

import (
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func validPrincipal(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if value == "IAM_ALLOWED_PRINCIPALS" {
		return ws, errors
	}

	// https://docs.aws.amazon.com/lake-formation/latest/dg/lf-permissions-reference.html
	// Principal is an AWS account
	// --principal DataLakePrincipalIdentifier=111122223333
	wsAccount, errorsAccount := verify.ValidAccountID(v, k)
	if len(errorsAccount) == 0 {
		return wsAccount, errorsAccount
	}

	wsARN, errorsARN := verify.ValidARN(v, k)
	ws = append(ws, wsARN...)
	errors = append(errors, errorsARN...)

	pattern := `:(role|user|federated-user|group|ou|organization)/`
	if !regexache.MustCompile(pattern).MatchString(value) {
		errors = append(errors, fmt.Errorf("%q does not look like a user, federated-user, role, group, OU, or organization: %q", k, value))
	}

	if len(errors) > 0 {
		errors = append(errors, errorsAccount...)
	}

	return ws, errors
}
