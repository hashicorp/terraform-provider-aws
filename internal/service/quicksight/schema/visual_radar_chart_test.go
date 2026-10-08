// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
)

func TestRadarChartVisualDataSourceSchema(t *testing.T) {
	t.Parallel()
	resourceSchema := radarChartVisualSchema()
	expectedDataSourceSchema := sdkv2.ComputedOnlyFromSchema(resourceSchema)
	dataSourceSchema := radarChartVisualDataSourceSchema()
	if diff := cmp.Diff(dataSourceSchema, expectedDataSourceSchema); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}
