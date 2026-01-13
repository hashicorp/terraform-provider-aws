// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"fmt"

	"github.com/YakDriver/regexache"
)

// http://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_CreateGlobalTable.html
func validGlobalTableName(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if (len(value) > 255) || (len(value) < 3) {
		errors = append(errors, fmt.Errorf("%s length must be between 3 and 255 characters: %q", k, value))
	}
	pattern := `^[0-9A-Za-z_.-]+$`
	if !regexache.MustCompile(pattern).MatchString(value) {
		errors = append(errors, fmt.Errorf("%s must only include alphanumeric, underscore, period, or hyphen characters: %q", k, value))
	}
	return
}
