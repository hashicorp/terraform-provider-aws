// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"testing"

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
