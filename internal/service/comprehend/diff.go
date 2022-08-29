package comprehend

import (
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func diffSuppressKMSKeyId(k, oldValue, newValue string, d *schema.ResourceData) bool {
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
