// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
)

func TestAxisDisplayOptionsDataSourceSchema(t *testing.T) {
	t.Parallel()

	resourceSchema := axisDisplayOptionsSchema()
	expectedDataSourceSchema := sdkv2.ComputedOnlyFromSchema(resourceSchema)

	dataSourceSchema := axisDisplayOptionsDataSourceSchema()

	if diff := cmp.Diff(dataSourceSchema, expectedDataSourceSchema); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}

func TestChartAxisLabelOptionsDataSourceSchema(t *testing.T) {
	t.Parallel()

	resourceSchema := chartAxisLabelOptionsSchema()
	expectedDataSourceSchema := sdkv2.ComputedOnlyFromSchema(resourceSchema)

	dataSourceSchema := chartAxisLabelOptionsDataSourceSchema()

	if diff := cmp.Diff(dataSourceSchema, expectedDataSourceSchema); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}

func TestItemsLimitConfigurationDataSourceSchema(t *testing.T) {
	t.Parallel()

	resourceSchema := itemsLimitConfigurationSchema()
	expectedDataSourceSchema := sdkv2.ComputedOnlyFromSchema(resourceSchema)

	dataSourceSchema := itemsLimitConfigurationDataSourceSchema()

	if diff := cmp.Diff(dataSourceSchema, expectedDataSourceSchema); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}

func TestContributionAnalysisDefaultsDataSourceSchema(t *testing.T) {
	t.Parallel()

	resourceSchema := contributionAnalysisDefaultsSchema()
	expectedDataSourceSchema := sdkv2.ComputedOnlyFromSchema(resourceSchema)

	dataSourceSchema := contributionAnalysisDefaultsDataSourceSchema()

	if diff := cmp.Diff(dataSourceSchema, expectedDataSourceSchema); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}

func TestReferenceLineDataSourceSchema(t *testing.T) {
	t.Parallel()

	resourceSchema := referenceLineSchema()
	expectedDataSourceSchema := sdkv2.ComputedOnlyFromSchema(resourceSchema)

	dataSourceSchema := referenceLineDataSourceSchema()

	if diff := cmp.Diff(dataSourceSchema, expectedDataSourceSchema); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}

func TestSmallMultiplesOptionsDataSourceSchema(t *testing.T) {
	t.Parallel()

	resourceSchema := smallMultiplesOptionsSchema()
	expectedDataSourceSchema := sdkv2.ComputedOnlyFromSchema(resourceSchema)

	dataSourceSchema := smallMultiplesOptionsDataSourceSchema()

	if diff := cmp.Diff(dataSourceSchema, expectedDataSourceSchema); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}
