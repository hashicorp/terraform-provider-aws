package comprehend

import (
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfkms "github.com/hashicorp/terraform-provider-aws/internal/service/kms"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func diffSuppressKMSKeyId(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	if oldValue == newValue {
		return true
	}

	oldId := oldValue
	if arn.IsARN(oldValue) {
		oldId = kmsKeyIdFromARN(oldValue)
	}

	newId := newValue
	if arn.IsARN(newValue) {
		newId = kmsKeyIdFromARN(newValue)
	}

	if oldId == newId {
		return true
	}

	return false
}

func diffSuppressKMSAlias(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	if oldValue == newValue {
		return true
	}

	oldAlias := oldValue
	if arn.IsARN(oldValue) {
		oldAlias = kmsKeyAliasFromARN(oldValue)
	}

	newAlias := newValue
	if arn.IsARN(newValue) {
		newAlias = kmsKeyAliasFromARN(newValue)
	}

	if oldAlias == newAlias {
		return true
	}

	return false
}

func diffSuppressKMSKeyOrAlias(k, oldValue, newValue string, d *schema.ResourceData) bool {
	if arn.IsARN(newValue) {
		if isKMSKeyARN(newValue) {
			return diffSuppressKMSKeyId(k, oldValue, newValue, d)
		} else {
			return diffSuppressKMSAlias(k, oldValue, newValue, d)
		}
	} else if isKMSAliasName(newValue) {
		return diffSuppressKMSAlias(k, oldValue, newValue, d)
	}
	return diffSuppressKMSKeyId(k, oldValue, newValue, d)
}

func kmsKeyIdFromARN(s string) string {
	arn, err := arn.Parse(s)
	if err != nil {
		return ""
	}

	return kmsKeyIdFromARNResource(arn.Resource)
}

func kmsKeyIdFromARNResource(s string) string {
	re := regexp.MustCompile(`^key/(` + verify.UUIDRegexPattern + ")$")
	matches := re.FindStringSubmatch(s)
	if matches == nil || len(matches) != 2 {
		return ""
	}

	return matches[1]
}

func kmsKeyAliasFromARN(s string) string {
	arn, err := arn.Parse(s)
	if err != nil {
		return ""
	}

	return kmsKeyAliasNameFromARNResource(arn.Resource)
}

func kmsKeyAliasNameFromARNResource(s string) string {
	re := regexp.MustCompile("^" + tfkms.AliasNameRegexPattern + "$")
	if re.MatchString(s) {
		return s
	}

	return ""
}

func isKMSKeyARN(s string) bool {
	parsedARN, err := arn.Parse(s)
	if err != nil {
		return false
	}

	return kmsKeyIdFromARNResource(parsedARN.Resource) != ""
}

func isKMSAliasName(s string) bool {
	return strings.HasPrefix(s, "alias/")
}

func isKMSAliasARN(s string) bool {
	parsedARN, err := arn.Parse(s)
	if err != nil {
		return false
	}

	return isKMSAliasName(parsedARN.Resource)
}
