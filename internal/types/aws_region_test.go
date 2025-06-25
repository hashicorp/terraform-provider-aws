// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import "testing"

func TestIsAWSRegion(t *testing.T) { // nosemgrep:ci.aws-in-func-name
	t.Parallel()

	for _, tc := range []struct {
		id    string
		valid bool
	}{
		{"us-east-1", true},
		{"jp-711", false},
		{"ap-southeast-7", true},
		{"", false},
		{"eu-isoe-west-1", true},
		{"mars", false},
	} {
		ok := IsAWSRegion(tc.id)
		if got, want := ok, tc.valid; got != want {
			t.Errorf("IsAWSRegion(%q) = %v, want %v", tc.id, got, want)
		}
	}
}
