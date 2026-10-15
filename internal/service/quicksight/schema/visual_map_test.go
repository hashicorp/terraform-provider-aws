// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
)

func TestGeospatialMapStyleOptionsDataSourceSchema(t *testing.T) {
	t.Parallel()
	resourceSchema := geospatialMapStyleOptionsSchema()
	expectedDataSourceSchema := sdkv2.ComputedOnlyFromSchema(resourceSchema)
	dataSourceSchema := geospatialMapStyleOptionsDataSourceSchema()
	if diff := cmp.Diff(dataSourceSchema, expectedDataSourceSchema); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}

func TestGeospatialWindowOptionsDataSourceSchema(t *testing.T) {
	t.Parallel()
	resourceSchema := geospatialWindowOptionsSchema()
	expectedDataSourceSchema := sdkv2.ComputedOnlyFromSchema(resourceSchema)
	dataSourceSchema := geospatialWindowOptionsDataSourceSchema()
	if diff := cmp.Diff(dataSourceSchema, expectedDataSourceSchema); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}
