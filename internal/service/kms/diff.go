package kms

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DiffSuppressKey(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	if oldValue == newValue {
		return true
	}

	oldId := oldValue
	if arn.IsARN(oldValue) {
		oldId = keyIdFromARN(oldValue)
	}

	newId := newValue
	if arn.IsARN(newValue) {
		newId = keyIdFromARN(newValue)
	}

	if oldId == newId {
		return true
	}

	return false
}

func DiffSuppressAlias(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	if oldValue == newValue {
		return true
	}

	oldAlias := oldValue
	if arn.IsARN(oldValue) {
		oldAlias = keyAliasFromARN(oldValue)
	}

	newAlias := newValue
	if arn.IsARN(newValue) {
		newAlias = keyAliasFromARN(newValue)
	}

	if oldAlias == newAlias {
		return true
	}

	return false
}

func DiffSuppressKeyOrAlias(k, oldValue, newValue string, d *schema.ResourceData) bool {
	if arn.IsARN(newValue) {
		if isKeyARN(newValue) {
			return DiffSuppressKey(k, oldValue, newValue, d)
		} else {
			return DiffSuppressAlias(k, oldValue, newValue, d)
		}
	} else if isAliasName(newValue) {
		return DiffSuppressAlias(k, oldValue, newValue, d)
	}
	return DiffSuppressKey(k, oldValue, newValue, d)
}

func keyIdFromARN(s string) string {
	arn, err := arn.Parse(s)
	if err != nil {
		return ""
	}

	return keyIdFromARNResource(arn.Resource)
}

func keyIdFromARNResource(s string) string {
	matches := keyIdResourceRegex.FindStringSubmatch(s)
	if matches == nil || len(matches) != 2 {
		return ""
	}

	return matches[1]
}

func keyAliasFromARN(s string) string {
	arn, err := arn.Parse(s)
	if err != nil {
		return ""
	}

	return keyAliasNameFromARNResource(arn.Resource)
}

func keyAliasNameFromARNResource(s string) string {
	if aliasNameRegex.MatchString(s) {
		return s
	}

	return ""
}

func isKeyARN(s string) bool {
	parsedARN, err := arn.Parse(s)
	if err != nil {
		return false
	}

	return keyIdFromARNResource(parsedARN.Resource) != ""
}

func isAliasName(s string) bool {
	return strings.HasPrefix(s, "alias/")
}

func isAliasARN(s string) bool {
	parsedARN, err := arn.Parse(s)
	if err != nil {
		return false
	}

	return isAliasName(parsedARN.Resource)
}
