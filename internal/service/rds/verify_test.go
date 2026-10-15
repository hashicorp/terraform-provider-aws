// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
)

func TestEngineVersionIsNewer(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		v1   string
		v2   string
		want bool
	}{
		"newer plain version": {
			v1:   "16.3",
			v2:   "16.2",
			want: true,
		},
		"equal plain version": {
			v1:   "16.3",
			v2:   "16.3",
			want: false,
		},
		"older plain version": {
			v1:   "16.2",
			v2:   "16.3",
			want: false,
		},
		"newer aurora patch version": {
			v1:   "8.0.mysql_aurora.3.10.0",
			v2:   "8.0.mysql_aurora.3.9.0",
			want: true,
		},
		"older aurora patch version": {
			v1:   "8.0.mysql_aurora.3.9.0",
			v2:   "8.0.mysql_aurora.3.10.0",
			want: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if diff := cmp.Diff(test.want, tfrds.EngineVersionIsNewer(test.v1, test.v2)); diff != "" {
				t.Errorf("engineVersionIsNewer(%q, %q) mismatch (-want +got):\n%s", test.v1, test.v2, diff)
			}
		})
	}
}
