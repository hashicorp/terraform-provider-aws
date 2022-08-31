package comprehend

import (
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	modelIdentifierMaxLen       = 63 // Documentation says 256, Console says 63
	modelIdentifierPrefixMaxLen = modelIdentifierMaxLen - resource.UniqueIDSuffixLength
)

var validModelName = validIdentifier
var validModelVersionName = validation.Any( // nosemgrep:ci.avoid-string-is-empty-validation
	validation.StringIsEmpty,
	validIdentifier,
)
var validModelVersionNamePrefix = validIdentifierPrefix

var validIdentifier = validation.All(
	validation.StringLenBetween(1, modelIdentifierMaxLen),
	validIdentifierPattern,
)

var validIdentifierPrefix = validation.All(
	validation.StringLenBetween(1, modelIdentifierPrefixMaxLen),
	validIdentifierPattern,
)

var validIdentifierPattern = validation.StringMatch(regexp.MustCompile(`^[[:alnum:]-]+$`), "must contain A-Z, a-z, 0-9, and hypen (-)")

var validateKMSKey = validation.Any(
	validateKMSKeyId,
	validateKMSKeyARN,
)

var validateKMSKeyId = validation.StringMatch(regexp.MustCompile("^"+verify.UUIDRegexPattern+"$"), "must be a KMS Key ID")

func validateKMSKeyARN(v any, k string) (ws []string, errors []error) {
	value, ok := v.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
		return
	}

	if value == "" {
		return
	}

	parsedARN, err := arn.Parse(value)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q (%s) is an invalid ARN: %s", k, value, err))
		return
	}

	if parsedARN.Service != "kms" {
		errors = append(errors, fmt.Errorf("%q (%s) is not a valid KMS Key ARN: %s", k, value, err))
		return
	}

	if id := kmsKeyIdFromARNResource(parsedARN.Resource); id == "" {
		errors = append(errors, fmt.Errorf("%q (%s) is not a valid KMS Key ARN: %s", k, value, err))
		return
	}

	return
}
