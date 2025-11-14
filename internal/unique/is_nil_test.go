// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package unique_test

import (
	"testing"
	"unique"

	tfunique "github.com/hashicorp/terraform-provider-aws/internal/unique"
)

func TestIsHandleNil(t *testing.T) {
	t.Parallel()

	type test struct {
		value string
	}

	testcases := map[string]struct {
		in       unique.Handle[test]
		expected bool
	}{
		"zero": {
			in:       unique.Handle[test]{},
			expected: true,
		},
		"empty": {
			in:       unique.Make(test{}),
			expected: false,
		},
		"value": {
			in: unique.Make(test{
				"value",
			}),
			expected: false,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			a := tfunique.IsHandleNil(testcase.in)
			e := testcase.expected

			if a != e {
				t.Fatalf("expected %t, got %t", e, a)
			}
		})
	}
}
