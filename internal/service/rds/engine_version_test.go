// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestCompareActualEngineVersion(t *testing.T) {
	t.Parallel()

	type testCase struct {
		configuredVersion           string
		actualVersion               string
		pendingVersion              string
		expectedEngineVersion       string
		expectedEngineVersionActual string
	}
	tests := map[string]testCase{
		"import": {
			configuredVersion:           "", // no "old" value on import
			actualVersion:               "8.1",
			expectedEngineVersion:       "8.1",
			expectedEngineVersionActual: "8.1",
		},
		"point version upgrade": {
			configuredVersion:           "8.0",
			actualVersion:               "8.0.27",
			expectedEngineVersion:       "8.0",
			expectedEngineVersionActual: "8.0.27",
		},
		"minor version upgrade": {
			configuredVersion:           "8.0",
			actualVersion:               "8.1.1",
			expectedEngineVersion:       "8.1.1",
			expectedEngineVersionActual: "8.1.1",
		},
		"major version upgrade": {
			configuredVersion:           "8.1",
			actualVersion:               "9.0.0",
			expectedEngineVersion:       "9.0.0",
			expectedEngineVersionActual: "9.0.0",
		},
		"pending minor version upgrade": {
			configuredVersion:           "8.1.1",
			actualVersion:               "8.0",
			pendingVersion:              "8.1.1",
			expectedEngineVersion:       "8.1.1",
			expectedEngineVersionActual: "8.0",
		},
		"pending major version upgrade": {
			configuredVersion:           "9.0.0",
			actualVersion:               "8.1",
			pendingVersion:              "9.0.0",
			expectedEngineVersion:       "9.0.0",
			expectedEngineVersionActual: "8.1",
		},
		"aurora upgrade": {
			configuredVersion:           "5.7.mysql_aurora.2.07",
			actualVersion:               "5.7.serverless_mysql_aurora.2.08.3",
			expectedEngineVersion:       "5.7.serverless_mysql_aurora.2.08.3",
			expectedEngineVersionActual: "5.7.serverless_mysql_aurora.2.08.3",
		},
		"aurora upgrade - used to crash": {
			configuredVersion:           "5.7.serverless_mysql_aurora.2.07",
			actualVersion:               "5.7.mysql_aurora.2.08.3",
			expectedEngineVersion:       "5.7.mysql_aurora.2.08.3",
			expectedEngineVersionActual: "5.7.mysql_aurora.2.08.3",
		},
		"oracle": {
			configuredVersion:           "12.1.0.2.v20",
			actualVersion:               "19.0.0.0.ru-2021-04.rur-2021-04.r1",
			expectedEngineVersion:       "19.0.0.0.ru-2021-04.rur-2021-04.r1",
			expectedEngineVersionActual: "19.0.0.0.ru-2021-04.rur-2021-04.r1",
		},
		"oracle - used to crash": {
			configuredVersion:           "19.0.0.0.ru-2021-04.rur-2021-04.r1",
			actualVersion:               "12.1.0.2.v20",
			expectedEngineVersion:       "12.1.0.2.v20",
			expectedEngineVersionActual: "12.1.0.2.v20",
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			r := ResourceCluster()
			d := r.Data(nil)
			d.Set(names.AttrEngineVersion, test.configuredVersion)
			compareActualEngineVersion(d, test.configuredVersion, test.actualVersion, test.pendingVersion)

			if want, got := test.expectedEngineVersion, d.Get(names.AttrEngineVersion); got != want {
				t.Errorf("unexpected engine_version; want: %q, got: %q", want, got)
			}
			if want, got := test.expectedEngineVersionActual, d.Get("engine_version_actual"); got != want {
				t.Errorf("unexpected engine_version_actual; want: %q, got: %q", want, got)
			}
		})
	}
}
