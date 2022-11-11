package comprehend

import (
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tfkms "github.com/hashicorp/terraform-provider-aws/internal/service/kms"
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

var validateKMSKeyOrAlias = validation.Any(
	validateKMSKeyId,
	validateKMSKeyARN,
	validateKMSKeyAliasName,
	validateKMSKeyAliasARN,
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

	if _, err := arn.Parse(value); err != nil {
		errors = append(errors, fmt.Errorf("%q (%s) is an invalid ARN: %s", k, value, err))
		return
	}

	if !isKMSKeyARN(value) {
		errors = append(errors, fmt.Errorf("%q (%s) is not a valid KMS Key ARN", k, value))
		return
	}

	return
}

var validateKMSKeyAliasName = validation.StringMatch(regexp.MustCompile("^"+tfkms.AliasNameRegexPattern+"$"), "must be a KMS Key Alias")

func validateKMSKeyAliasARN(v any, k string) (ws []string, errors []error) {
	value, ok := v.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
		return
	}

	if value == "" {
		return
	}

	if _, err := arn.Parse(value); err != nil {
		errors = append(errors, fmt.Errorf("%q (%s) is an invalid ARN: %s", k, value, err))
		return
	}

	if !isKMSAliasARN(value) {
		errors = append(errors, fmt.Errorf("%q (%s) is not a valid KMS Key Alias ARN", k, value))
		return
	}

	return
}
