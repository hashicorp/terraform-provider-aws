// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package actions_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwactions "github.com/hashicorp/terraform-provider-aws/internal/framework/actions"
)

func TestTimeoutOr(t *testing.T) {
	t.Parallel()

	defaultValue := 600 * time.Second
	type testCase struct {
		input    types.Int64
		expected time.Duration
	}
	tests := map[string]testCase{
		"null": {
			input:    types.Int64Null(),
			expected: defaultValue,
		},
		"non-null": {
			input:    types.Int64Value(30),
			expected: 30 * time.Second,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := fwactions.TimeoutOr(test.input, defaultValue)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
