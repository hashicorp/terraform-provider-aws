// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package verify

import (
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"gopkg.in/yaml.v2"
)

const UUIDRegexPattern = `[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[ab89][0-9a-f]{3}-[0-9a-f]{12}`

func SliceContainsString(slice []interface{}, s string) (int, bool) {
	for idx, value := range slice {
		v := value.(string)
		if v == s {
			return idx, true
		}
	}
	return -1, false
}

// Takes a value containing YAML string and passes it through
// the YAML parser. Returns either a parsing
// error or original YAML string.
func checkYAMLString(yamlString interface{}) (string, error) {
	var y interface{}

	if yamlString == nil || yamlString.(string) == "" {
		return "", nil
	}

	s := yamlString.(string)

	err := yaml.Unmarshal([]byte(s), &y)

	return s, err
}

const (
	ErrCodeAccessDenied                = "AccessDenied"
	ErrCodeAuthorizationError          = "AuthorizationError"
	ErrCodeInternalException           = "InternalException"
	ErrCodeInternalServiceError        = "InternalServiceError"
	ErrCodeInvalidAction               = "InvalidAction"
	ErrCodeInvalidParameterException   = "InvalidParameterException"
	ErrCodeInvalidParameterValue       = "InvalidParameterValue"
	ErrCodeInvalidRequest              = "InvalidRequest"
	ErrCodeOperationDisabledException  = "OperationDisabledException"
	ErrCodeOperationNotPermitted       = "OperationNotPermitted"
	ErrCodeUnknownOperationException   = "UnknownOperationException"
	ErrCodeUnsupportedFeatureException = "UnsupportedFeatureException"
	ErrCodeUnsupportedOperation        = "UnsupportedOperation"
	ErrCodeValidationError             = "ValidationError"
	ErrCodeValidationException         = "ValidationException"
)

// ErrorISOUnsupported checks the partition and specific error to make
// an educated guess about whether the problem stems from a feature not being
// available in ISO (or non standard partitions) that is normally available.
// true means that there is an error AND it suggests a feature is not supported
// in ISO. Be careful with false, which means either there is NO error or there
// is an error but not one that suggests an unsupported feature in ISO.
func ErrorISOUnsupported(partition string, err error) bool {
	if partition == endpoints.AwsPartitionID {
		return false
	}

	if err == nil { // not strictly necessary but make logic clearer
		return false
	}

	if tfawserr.ErrCodeContains(err, ErrCodeAccessDenied) {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeAuthorizationError) {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeInternalException) {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeInternalServiceError) {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeInvalidAction) {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeInvalidParameterException) {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeInvalidParameterValue) {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeInvalidRequest) {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeOperationDisabledException) {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeOperationNotPermitted) {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeUnknownOperationException) {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeUnsupportedFeatureException) {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeUnsupportedOperation) {
		return true
	}

	if tfawserr.ErrMessageContains(err, ErrCodeValidationError, "not support tagging") {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeValidationException) {
		return true
	}

	return false
}
