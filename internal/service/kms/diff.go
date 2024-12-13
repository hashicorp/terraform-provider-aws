// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func diffSuppressKey(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	if oldValue == newValue {
		return true
	}

	oldID := oldValue
	if arn.IsARN(oldValue) {
		oldID = keyIDFromARN(oldValue)
	}

	newID := newValue
	if arn.IsARN(newValue) {
		newID = keyIDFromARN(newValue)
	}

	if oldID == newID {
		return true
	}

	return false
}

func diffSuppressAlias(_, oldValue, newValue string, _ *schema.ResourceData) bool {
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

func diffSuppressKeyOrAlias(k, oldValue, newValue string, d *schema.ResourceData) bool {
	if arn.IsARN(newValue) {
		if isKeyARN(newValue) {
			return diffSuppressKey(k, oldValue, newValue, d)
		} else {
			return diffSuppressAlias(k, oldValue, newValue, d)
		}
	} else if isAliasName(newValue) {
		return diffSuppressAlias(k, oldValue, newValue, d)
	}
	return diffSuppressKey(k, oldValue, newValue, d)
}

func keyIDFromARN(s string) string {
	arn, err := arn.Parse(s)
	if err != nil {
		return ""
	}

	return keyIDFromARNResource(arn.Resource)
}

func keyIDFromARNResource(s string) string {
	matches := keyIDResourceRegex.FindStringSubmatch(s)
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

	return keyIDFromARNResource(parsedARN.Resource) != ""
}

func isAliasName(s string) bool {
	return strings.HasPrefix(s, aliasNamePrefix)
}

func isAliasARN(s string) bool {
	parsedARN, err := arn.Parse(s)
	if err != nil {
		return false
	}

	return isAliasName(parsedARN.Resource)
}
