// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"strings"

	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/json"
)

// SuppressEquivalentStringCaseInsensitive provides custom difference suppression
// for strings that are equal under case-insensitivity.
func SuppressEquivalentStringCaseInsensitive(k, old, new string, d *schema.ResourceData) bool {
	return strings.EqualFold(old, new)
}

// SuppressEquivalentJSONDocuments provides custom difference suppression
// for JSON documents in the given strings that are equivalent.
func SuppressEquivalentJSONDocuments(k, old, new string, d *schema.ResourceData) bool {
	return json.EqualStrings(old, new)
}

// SuppressEquivalentIAMPolicyDocuments provides custom difference suppression
// for IAM policy documents in the given strings that are equivalent.
func SuppressEquivalentIAMPolicyDocuments(k, old, new string, d *schema.ResourceData) bool {
	if equalEmptyJSONStrings(old, new) {
		return true
	}

	equivalent, err := awspolicy.PoliciesAreEquivalent(old, new)
	if err != nil {
		return false
	}

	return equivalent
}

func equalEmptyJSONStrings(old, new string) bool {
	old, new = strings.TrimSpace(old), strings.TrimSpace(new)
	return (old == "" || old == "{}") && (new == "" || new == "{}")
}
