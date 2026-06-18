// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"strings"
	"time"

	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/json"
)

// SuppressEquivalentStringCaseInsensitive provides custom difference suppression
// for strings that are equal under case-insensitivity.
func SuppressEquivalentStringCaseInsensitive(k, old, new string, _ *schema.ResourceData) bool {
	return strings.EqualFold(old, new)
}

// SuppressEquivalentJSONDocuments provides custom difference suppression
// for JSON documents in the given strings that are equivalent.
func SuppressEquivalentJSONDocuments(k, old, new string, _ *schema.ResourceData) bool {
	return json.EqualStrings(old, new)
}

// SuppressEquivalentJSONDocumentsWithEmpty provides custom difference suppression
// for JSON documents in the given strings that are equivalent, handling empty
// strings (`""`) and empty JSON strings (`"{}"`) as equivalent.
// This is useful for suppressing diffs for non-IAM JSON policy documents.
func SuppressEquivalentJSONDocumentsWithEmpty(k, old, new string, _ *schema.ResourceData) bool {
	if equalEmptyJSONStrings(old, new) {
		return true
	}

	return json.EqualStrings(old, new)
}

// SuppressEquivalentCloudWatchLogsLogGroupARN provides custom difference suppression
// for strings that represent equal CloudWatch Logs log group ARNs.
func SuppressEquivalentCloudWatchLogsLogGroupARN(_, old, new string, _ *schema.ResourceData) bool {
	return strings.TrimSuffix(old, ":*") == strings.TrimSuffix(new, ":*")
}

// SuppressEquivalentRoundedTime returns a difference suppression function that compares
// two time value with the specified layout rounded to the specified duration.
func SuppressEquivalentRoundedTime(layout string, d time.Duration) schema.SchemaDiffSuppressFunc {
	return func(k, old, new string, _ *schema.ResourceData) bool {
		if old, err := time.Parse(layout, old); err == nil {
			if new, err := time.Parse(layout, new); err == nil {
				return old.Round(d).Equal(new.Round(d))
			}
		}

		return false
	}
}

// SuppressEquivalentTime returns a difference suppression function that suppresses differences
// for time values that represent the same instant in different timezones.
func SuppressEquivalentTime(k, old, new string, d *schema.ResourceData) bool {
	oldTime, err := time.Parse(time.RFC3339, old)
	if err != nil {
		return false
	}

	newTime, err := time.Parse(time.RFC3339, new)
	if err != nil {
		return false
	}

	return oldTime.Equal(newTime)
}

// SuppressEquivalentIAMPolicyDocuments provides custom difference suppression
// for IAM policy documents in the given strings that are equivalent.
func SuppressEquivalentIAMPolicyDocuments(k, old, new string, _ *schema.ResourceData) bool {
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

// SuppressNewStringValueEquivalentToUnset provides custom difference suppression
// for new string values that are equivalent to the unset (unconfigured) value.
func SuppressNewStringValueEquivalentToUnset[T ~string](v T) schema.SchemaDiffSuppressFunc {
	return func(k, old, new string, d *schema.ResourceData) bool {
		return d.Id() != "" && old == "" && T(new) == v
	}
}
