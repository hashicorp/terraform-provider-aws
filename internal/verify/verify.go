package verify

import (
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
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
	ErrCodeAccessDenied                   = "AccessDenied"
	ErrCodeUnknownOperation               = "UnknownOperationException"
	ErrCodeValidationError                = "ValidationError"
	ErrCodeOperationDisabledException     = "OperationDisabledException"
	ErrCodeInternalException              = "InternalException"
	ErrCodeInternalServiceFault           = "InternalServiceError"
	ErrCodeOperationNotPermittedException = "OperationNotPermitted"
)

func CheckISOErrorTagsUnsupported(err error) bool {
	if tfawserr.ErrCodeContains(err, ErrCodeAccessDenied) {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeUnknownOperation) {
		return true
	}

	if tfawserr.ErrMessageContains(err, ErrCodeValidationError, "not support tagging") {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeOperationDisabledException) {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeInternalException) {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeInternalServiceFault) {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeOperationNotPermittedException) {
		return true
	}

	return false
}
