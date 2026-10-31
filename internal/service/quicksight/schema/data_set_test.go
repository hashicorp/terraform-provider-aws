// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
)

func TestDataSetColumnGroupsSchemaDataSourceSchema(t *testing.T) {
	t.Parallel()

	resourceSchema := DataSetColumnGroupsSchema()
	expectedDataSourceSchema := sdkv2.ComputedOnlyFromSchema(resourceSchema)

	dataSourceSchema := DataSetColumnGroupsSchemaDataSourceSchema()

	if diff := cmp.Diff(dataSourceSchema, expectedDataSourceSchema); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}

func TestDataSetColumnLevelPermissionRulesSchemaDataSourceSchema(t *testing.T) {
	t.Parallel()

	resourceSchema := DataSetColumnLevelPermissionRulesSchema()
	expectedDataSourceSchema := sdkv2.ComputedOnlyFromSchema(resourceSchema)

	dataSourceSchema := DataSetColumnLevelPermissionRulesSchemaDataSourceSchema()

	if diff := cmp.Diff(dataSourceSchema, expectedDataSourceSchema); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}

func TestDataSetUsageConfigurationSchemaDataSourceSchema(t *testing.T) {
	t.Parallel()

	resourceSchema := DataSetUsageConfigurationSchema()
	expectedDataSourceSchema := sdkv2.ComputedOnlyFromSchema(resourceSchema)

	dataSourceSchema := DataSetUsageConfigurationSchemaDataSourceSchema()

	if diff := cmp.Diff(dataSourceSchema, expectedDataSourceSchema); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}

func TestDataSetFieldFoldersSchemaDataSourceSchema(t *testing.T) {
	t.Parallel()

	resourceSchema := DataSetFieldFoldersSchema()
	expectedDataSourceSchema := sdkv2.ComputedOnlyFromSchema(resourceSchema)

	dataSourceSchema := DataSetFieldFoldersSchemaDataSourceSchema()

	if diff := cmp.Diff(dataSourceSchema, expectedDataSourceSchema); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}

func TestDataSetLogicalTableMapSchemaDataSourceSchema(t *testing.T) {
	t.Parallel()

	resourceSchema := DataSetLogicalTableMapSchema()
	expectedDataSourceSchema := sdkv2.ComputedOnlyFromSchema(resourceSchema)

	dataSourceSchema := DataSetLogicalTableMapSchemaDataSourceSchema()

	if diff := cmp.Diff(dataSourceSchema, expectedDataSourceSchema); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}

func TestDataSetPhysicalTableMapSchemaDataSourceSchema(t *testing.T) {
	t.Parallel()

	resourceSchema := DataSetPhysicalTableMapSchema()
	expectedDataSourceSchema := sdkv2.ComputedOnlyFromSchema(resourceSchema)

	dataSourceSchema := DataSetPhysicalTableMapSchemaDataSourceSchema()

	if diff := cmp.Diff(dataSourceSchema, expectedDataSourceSchema); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}

func TestDataSetRowLevelPermissionDataSetSchemaDataSourceSchema(t *testing.T) {
	t.Parallel()

	resourceSchema := DataSetRowLevelPermissionDataSetSchema()
	expectedDataSourceSchema := sdkv2.ComputedOnlyFromSchema(resourceSchema)

	dataSourceSchema := DataSetRowLevelPermissionDataSetSchemaDataSourceSchema()

	if diff := cmp.Diff(dataSourceSchema, expectedDataSourceSchema); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}

func TestDataSetRowLevelPermissionTagConfigurationSchemaDataSourceSchema(t *testing.T) {
	t.Parallel()

	resourceSchema := DataSetRowLevelPermissionTagConfigurationSchema()
	expectedDataSourceSchema := sdkv2.ComputedOnlyFromSchema(resourceSchema)

	dataSourceSchema := DataSetRowLevelPermissionTagConfigurationSchemaDataSourceSchema()

	if diff := cmp.Diff(dataSourceSchema, expectedDataSourceSchema); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}

func TestExpandColumnTag(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		tfMap    map[string]any
		expected awstypes.GeoSpatialDataRole
	}{
		"column_geographic_role omitted": {
			tfMap:    map[string]any{},
			expected: "",
		},
		"column_geographic_role empty string": {
			tfMap:    map[string]any{"column_geographic_role": ""},
			expected: "",
		},
		"column_geographic_role set": {
			tfMap:    map[string]any{"column_geographic_role": string(awstypes.GeoSpatialDataRoleState)},
			expected: awstypes.GeoSpatialDataRoleState,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := expandColumnTag(testCase.tfMap)

			if got == nil {
				t.Fatal("expandColumnTag() returned nil")
			}

			if got.ColumnGeographicRole != testCase.expected {
				t.Errorf("ColumnGeographicRole = %q, want %q", got.ColumnGeographicRole, testCase.expected)
			}
		})
	}
}
