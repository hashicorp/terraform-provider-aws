package lakeformation

import (
	"fmt"
	"regexp"
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

	pattern := `:(role|user|group|ou|organization)/`
	if !regexp.MustCompile(pattern).MatchString(value) {
		errors = append(errors, fmt.Errorf("%q does not look like a user, role, group, OU, or organization: %q", k, value))
	}

	if len(errors) > 0 {
		errors = append(errors, errorsAccount...)
	}

	return ws, errors
}
