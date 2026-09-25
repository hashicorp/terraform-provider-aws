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
		// PostgreSQL cases from https://github.com/hashicorp/terraform-provider-aws/issues/39579.
		// A major-only prefix or an older minor must not be considered newer than
		// the running minor (otherwise it would be sent to RDS as a downgrade),
		// while a newer minor or major must be.
		"postgres major prefix not newer than running minor": {
			v1:   "14",
			v2:   "14.22",
			want: false,
		},
		"postgres older minor not newer": {
			v1:   "14.22",
			v2:   "14.23",
			want: false,
		},
		"postgres newer minor": {
			v1:   "14.23",
			v2:   "14.22",
			want: true,
		},
		"postgres newer major": {
			v1:   "15",
			v2:   "14.22",
			want: true,
		},
		// Guards create/import, where engine_version_actual is not yet known: a
		// concrete configured version must be treated as newer than an empty
		// actual so the diff is not suppressed before the resource exists.
		"nonempty newer than empty actual": {
			v1:   "14",
			v2:   "",
			want: true,
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
