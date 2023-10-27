// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package errs

import (
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
)

const (
	errCodeAccessDenied                = "AccessDenied"
	errCodeAuthorizationError          = "AuthorizationError"
	errCodeInternalException           = "InternalException"
	errCodeInternalServiceError        = "InternalServiceError"
	errCodeInvalidAction               = "InvalidAction"
	errCodeInvalidParameterException   = "InvalidParameterException"
	errCodeInvalidParameterValue       = "InvalidParameterValue"
	errCodeInvalidRequest              = "InvalidRequest"
	errCodeOperationDisabledException  = "OperationDisabledException"
	errCodeOperationNotPermitted       = "OperationNotPermitted"
	errCodeUnknownOperationException   = "UnknownOperationException"
	errCodeUnsupportedFeatureException = "UnsupportedFeatureException"
	errCodeUnsupportedOperation        = "UnsupportedOperation"
	errCodeValidationError             = "ValidationError"
	errCodeValidationException         = "ValidationException"
)

// IsUnsupportedOperationInPartitionError checks the partition and specific error
// to make an educated guess about whether the problem stems from a feature not being
// available in a non-standard partitions (e.g. ISO) that is normally available.
// A return value of `true` means that there is an error AND it suggests a feature is not supported in ISO.
// Be careful with a return value of `false“, which means either there is NO error
// or there is an error but not one that suggests an unsupported feature in ISO.
func IsUnsupportedOperationInPartitionError(partition string, err error) bool {
	if partition == endpoints.AwsPartitionID {
		return false
	}

	if err == nil { // not strictly necessary but make logic clearer
		return false
	}

	if tfawserr.ErrCodeContains(err, errCodeAccessDenied) {
		return true
	}

	if tfawserr.ErrCodeContains(err, errCodeAuthorizationError) {
		return true
	}

	if tfawserr.ErrCodeContains(err, errCodeInternalException) {
		return true
	}

	if tfawserr.ErrCodeContains(err, errCodeInternalServiceError) {
		return true
	}

	if tfawserr.ErrCodeContains(err, errCodeInvalidAction) {
		return true
	}

	if tfawserr.ErrCodeContains(err, errCodeInvalidParameterException) {
		return true
	}

	if tfawserr.ErrCodeContains(err, errCodeInvalidParameterValue) {
		return true
	}

	if tfawserr.ErrCodeContains(err, errCodeInvalidRequest) {
		return true
	}

	if tfawserr.ErrCodeContains(err, errCodeOperationDisabledException) {
		return true
	}

	if tfawserr.ErrCodeContains(err, errCodeOperationNotPermitted) {
		return true
	}

	if tfawserr.ErrCodeContains(err, errCodeUnknownOperationException) {
		return true
	}

	if tfawserr.ErrCodeContains(err, errCodeUnsupportedFeatureException) {
		return true
	}

	if tfawserr.ErrCodeContains(err, errCodeUnsupportedOperation) {
		return true
	}

	if tfawserr.ErrMessageContains(err, errCodeValidationError, "not support tagging") {
		return true
	}

	if tfawserr.ErrCodeContains(err, errCodeValidationException) {
		return true
	}

	return false
}
