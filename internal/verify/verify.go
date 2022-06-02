package verify

import (
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"gopkg.in/yaml.v2"
)

const UUIDRegexPattern = `[a-f0-9]{8}-[a-f0-9]{4}-[1-5][a-f0-9]{3}-[ab89][a-f0-9]{3}-[a-f0-9]{12}`

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

func CheckISOErrorTagsUnsupported(err error) bool {
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
