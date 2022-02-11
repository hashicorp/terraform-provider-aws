package rds

import (
	"testing"
)

func TestCompareActualEngineVersion(t *testing.T) {
	t.Parallel()

	type testCase struct {
		oldVersion                  string
		newVersion                  string
		expectedEngineVersion       string
		expectedEngineVersionActual string
	}
	tests := map[string]testCase{
		"point version upgrade": {
			oldVersion:                  "8.0",
			newVersion:                  "8.0.27",
			expectedEngineVersion:       "",
			expectedEngineVersionActual: "8.0.27",
		},
		"minor version upgrade": {
			oldVersion:                  "8.0",
			newVersion:                  "8.1.1",
			expectedEngineVersion:       "8.1.1",
			expectedEngineVersionActual: "8.1.1",
		},
		"major version upgrade": {
			oldVersion:                  "8.1",
			newVersion:                  "9.0.0",
			expectedEngineVersion:       "9.0.0",
			expectedEngineVersionActual: "9.0.0",
		},
		"aurora upgrade": {
			oldVersion:                  "5.7.mysql_aurora.2.07",
			newVersion:                  "5.7.serverless_mysql_aurora.2.08.3",
			expectedEngineVersion:       "5.7.serverless_mysql_aurora.2.08.3",
			expectedEngineVersionActual: "5.7.serverless_mysql_aurora.2.08.3",
		},
		"aurora upgrade - used to crash": {
			oldVersion:                  "5.7.serverless_mysql_aurora.2.07",
			newVersion:                  "5.7.mysql_aurora.2.08.3",
			expectedEngineVersion:       "5.7.mysql_aurora.2.08.3",
			expectedEngineVersionActual: "5.7.mysql_aurora.2.08.3",
		},
		"oracle": {
			oldVersion:                  "12.1.0.2.v20",
			newVersion:                  "19.0.0.0.ru-2021-04.rur-2021-04.r1",
			expectedEngineVersion:       "19.0.0.0.ru-2021-04.rur-2021-04.r1",
			expectedEngineVersionActual: "19.0.0.0.ru-2021-04.rur-2021-04.r1",
		},
		"oracle - used to crash": {
			oldVersion:                  "19.0.0.0.ru-2021-04.rur-2021-04.r1",
			newVersion:                  "12.1.0.2.v20",
			expectedEngineVersion:       "12.1.0.2.v20",
			expectedEngineVersionActual: "12.1.0.2.v20",
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			r := ResourceCluster()
			d := r.Data(nil)
			compareActualEngineVersion(d, test.oldVersion, test.newVersion)

			if want, got := test.expectedEngineVersion, d.Get("engine_version"); got != want {
				t.Errorf("unexpected engine_version; want: %q, got: %q", want, got)
			}
			if want, got := test.expectedEngineVersionActual, d.Get("engine_version_actual"); got != want {
				t.Errorf("unexpected engine_version_actual; want: %q, got: %q", want, got)
			}
		})
	}
}
