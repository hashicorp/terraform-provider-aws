// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package awsmapping

import "unicode"

// CamelToSnake converts a PascalCase or camelCase name to snake_case.
// Examples:
//
//	"QueueName"         → "queue_name"
//	"KmsMasterKeyId"    → "kms_master_key_id"
//	"FifoQueue"         → "fifo_queue"
//	"SqsManagedSseEnabled" → "sqs_managed_sse_enabled"
func CamelToSnake(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	out := make([]rune, 0, len(runes)+4)

	for i, r := range runes {
		if unicode.IsUpper(r) {
			// Insert underscore before an uppercase letter when:
			// (a) it is not the first character, AND
			// (b) either the previous char was lowercase, OR the next char is
			//     lowercase (handles "KMSKey" → "kms_key", not "k_m_s_key").
			if i > 0 {
				prev := runes[i-1]
				next := rune(0)
				if i+1 < len(runes) {
					next = runes[i+1]
				}
				if unicode.IsLower(prev) || (unicode.IsLower(next) && unicode.IsUpper(prev)) {
					out = append(out, '_')
				}
			}
			out = append(out, unicode.ToLower(r))
		} else {
			out = append(out, r)
		}
	}
	return string(out)
}
