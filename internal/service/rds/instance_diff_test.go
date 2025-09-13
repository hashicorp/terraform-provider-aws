// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import "testing"

func TestInstanceReplicateSourceDBSuppressDiff(t *testing.T) {
	t.Parallel()

	type testCase struct {
		old, new string
		expected bool
	}
	testCases := map[string]testCase{
		"no values": {
			old:      "",
			new:      "",
			expected: false,
		},

		"old ARN same identifier": {
			old:      "arn:aws:rds:us-west-2:123456789012:db:test", //lintignore:AWSAT003,AWSAT005
			new:      "test",
			expected: true,
		},

		"old ARN different identifier": {
			old:      "arn:aws:rds:us-west-2:123456789012:db:test1", //lintignore:AWSAT003,AWSAT005
			new:      "test2",
			expected: false,
		},

		"new ARN same identifier": {
			old:      "test",
			new:      "arn:aws:rds:us-west-2:123456789012:db:test", //lintignore:AWSAT003,AWSAT005
			expected: true,
		},

		"new ARN different identifier": {
			old:      "test2",
			new:      "arn:aws:rds:us-west-2:123456789012:db:test1", //lintignore:AWSAT003,AWSAT005
			expected: false,
		},

		"both ARN": {
			old:      "arn:aws:rds:us-west-2:123456789012:db:test1", //lintignore:AWSAT003,AWSAT005
			new:      "arn:aws:rds:us-west-2:123456789012:db:test2", //lintignore:AWSAT003,AWSAT005
			expected: false,
		},

		"neither ARN": {
			old:      "test1",
			new:      "test2",
			expected: false,
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			v := instanceReplicateSourceDBSuppressDiff("", test.old, test.new, nil)
			if e, a := test.expected, v; e != a {
				t.Errorf("unexpected result: expected %t, got %t", e, a)
			}
		})
	}
}
