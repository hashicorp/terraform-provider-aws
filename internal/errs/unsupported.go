// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package errs

import (
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	tfawserr_sdkv2 "github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-provider-aws/names"
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
// Be careful with a return value of `false`, which means either there is NO error
// or there is an error but not one that suggests an unsupported feature in ISO.
func IsUnsupportedOperationInPartitionError(partition string, err error) bool {
	if partition == names.StandardPartitionID {
		return false
	}

	if err == nil { // not strictly necessary but make logic clearer
		return false
	}

	if errCodeContains(err, errCodeAccessDenied) {
		return true
	}

	if errCodeContains(err, errCodeAuthorizationError) {
		return true
	}

	if errCodeContains(err, errCodeInternalException) {
		return true
	}

	if errCodeContains(err, errCodeInternalServiceError) {
		return true
	}

	if errCodeContains(err, errCodeInvalidAction) {
		return true
	}

	if errCodeContains(err, errCodeInvalidParameterException) {
		return true
	}

	if errCodeContains(err, errCodeInvalidParameterValue) {
		return true
	}

	if errCodeContains(err, errCodeInvalidRequest) {
		return true
	}

	if errCodeContains(err, errCodeOperationDisabledException) {
		return true
	}

	if errCodeContains(err, errCodeOperationNotPermitted) {
		return true
	}

	if errCodeContains(err, errCodeUnknownOperationException) {
		return true
	}

	if errCodeContains(err, errCodeUnsupportedFeatureException) {
		return true
	}

	if errCodeContains(err, errCodeUnsupportedOperation) {
		return true
	}

	if errMessageContains(err, errCodeValidationError, "not support tagging") {
		return true
	}

	if errCodeContains(err, errCodeValidationException) {
		return true
	}

	return false
}

func errCodeContains(err error, code string) bool {
	return tfawserr.ErrCodeContains(err, code) || tfawserr_sdkv2.ErrCodeContains(err, code)
}

func errMessageContains(err error, code, message string) bool {
	return tfawserr.ErrMessageContains(err, code, message) || tfawserr_sdkv2.ErrMessageContains(err, code, message)
}
