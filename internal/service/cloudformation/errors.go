// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation

import (
	"strings"
)

const (
	errCodeValidationError = "ValidationError"
)

func isRetryableIAMPropagationErr(err error) (bool, error) {
	if err == nil {
		return false, nil
	}

	message := err.Error()

	// IAM eventual consistency
	if strings.Contains(message, "AccountGate check failed") {
		return true, err
	}

	// IAM eventual consistency
	// User: XXX is not authorized to perform: cloudformation:CreateStack on resource: YYY
	if strings.Contains(message, "is not authorized") {
		return true, err
	}

	// IAM eventual consistency
	// XXX role has insufficient YYY permissions
	if strings.Contains(message, "role has insufficient") {
		return true, err
	}

	// IAM eventual consistency
	// Account XXX should have YYY role with trust relationship to Role ZZZ
	if strings.Contains(message, "role with trust relationship") {
		return true, err
	}

	// IAM eventual consistency
	if strings.Contains(message, "The security token included in the request is invalid") {
		return true, err
	}

	return false, err
}
