// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestSetParameterGroupName(t *testing.T) {
	t.Parallel()

	type testCase struct {
		configuredPG                      string
		currentPG                         string
		hasPendingEngineVersionUpgrade    bool
		expectedParameterGroupName        string
		expectedParameterGroupNameActual  string
	}
	tests := map[string]testCase{
		"no pending upgrade": {
			configuredPG:                     "my-db-pg16",
			currentPG:                        "my-db-pg16",
			hasPendingEngineVersionUpgrade:   false,
			expectedParameterGroupName:       "my-db-pg16",
			expectedParameterGroupNameActual: "my-db-pg16",
		},
		"pending major version upgrade": {
			configuredPG:                     "my-db-pg17",
			currentPG:                        "my-db-pg16",
			hasPendingEngineVersionUpgrade:   true,
			expectedParameterGroupName:       "my-db-pg17",
			expectedParameterGroupNameActual: "my-db-pg16",
		},
		"no pending upgrade with different groups": {
			configuredPG:                     "my-db-pg16-old",
			currentPG:                        "my-db-pg16-new",
			hasPendingEngineVersionUpgrade:   false,
			expectedParameterGroupName:       "my-db-pg16-new",
			expectedParameterGroupNameActual: "my-db-pg16-new",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			r := resourceInstance()
			d := r.Data(nil)
			d.Set(names.AttrParameterGroupName, test.configuredPG)
			setParameterGroupName(d, test.currentPG, test.hasPendingEngineVersionUpgrade)

			if want, got := test.expectedParameterGroupName, d.Get(names.AttrParameterGroupName); got != want {
				t.Errorf("unexpected parameter_group_name; want: %q, got: %q", want, got)
			}
			if want, got := test.expectedParameterGroupNameActual, d.Get("parameter_group_name_actual"); got != want {
				t.Errorf("unexpected parameter_group_name_actual; want: %q, got: %q", want, got)
			}
		})
	}
}
