// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lakeformation

import (
	"fmt"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

var s3TablesCatalogPattern = regexache.MustCompile(`^\d{12}:s3tablescatalog\/[a-zA-Z0-9][a-zA-Z0-9._-]*[a-zA-Z0-9]$`)

func validPrincipal(v any, k string) (ws []string, errors []error) {
	value := v.(string)

	if value == "IAM_ALLOWED_PRINCIPALS" {
		return ws, errors
	}

	// IAMPrincipals special grant has format {account_id}:IAMPrincipals
	if val := strings.Split(value, ":"); len(val) == 2 && val[1] == "IAMPrincipals" {
		wsAccount, errorsAccount := verify.ValidAccountID(val[0], k)
		if len(errorsAccount) == 0 {
			return wsAccount, errorsAccount
		}

		ws = append(ws, wsAccount...)
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

func validCatalogID(v any, k string) (ws []string, errors []error) {
	value := v.(string)

	// First check if it's a standard AWS account ID (12 digits)
	wsAccount, errorsAccount := verify.ValidAccountID(v, k)
	if len(errorsAccount) == 0 {
		return wsAccount, errorsAccount
	}

	// Check if it's an S3 Tables catalog ID format: {account-id}:s3tablescatalog/{table-bucket-name}
	if s3TablesCatalogPattern.MatchString(value) {
		// Additional validation: extract account ID part and validate it
		parts := strings.Split(value, ":")
		if len(parts) >= 2 {
			accountID := parts[0]
			_, accountErrs := verify.ValidAccountID(accountID, k)
			if len(accountErrs) == 0 {
				return ws, errors // Valid S3 Tables catalog ID
			}
		}
	}

	// If neither format matches, return both error messages
	errors = append(errors, fmt.Errorf(
		"%q must be either a valid AWS Account ID (exactly 12 digits) or an S3 Tables catalog ID in the format 'account-id:s3tablescatalog/table-bucket-name': %q",
		k, value))

	return ws, errors
}
