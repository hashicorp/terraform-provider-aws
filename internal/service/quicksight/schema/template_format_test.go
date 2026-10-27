// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
)

func TestNumericFormatConfigurationDataSourceSchema(t *testing.T) {
	t.Parallel()

	resourceSchema := numericFormatConfigurationSchema()
	expectedDataSourceSchema := sdkv2.ComputedOnlyFromSchema(resourceSchema)

	dataSourceSchema := numericFormatConfigurationDataSourceSchema()

	if diff := cmp.Diff(dataSourceSchema, expectedDataSourceSchema); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}

func TestDateTimeFormatConfigurationDataSourceSchema(t *testing.T) {
	t.Parallel()

	resourceSchema := dateTimeFormatConfigurationSchema()
	expectedDataSourceSchema := sdkv2.ComputedOnlyFromSchema(resourceSchema)

	dataSourceSchema := dateTimeFormatConfigurationDataSourceSchema()

	if diff := cmp.Diff(dataSourceSchema, expectedDataSourceSchema); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}

func TestNumberDisplayFormatConfigurationDataSourceSchema(t *testing.T) {
	t.Parallel()

	resourceSchema := numberDisplayFormatConfigurationSchema()
	expectedDataSourceSchema := sdkv2.ComputedOnlyFromSchema(resourceSchema)

	dataSourceSchema := numberDisplayFormatConfigurationDataSourceSchema()

	if diff := cmp.Diff(dataSourceSchema, expectedDataSourceSchema); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}

func TestPercentageDisplayFormatConfigurationDataSourceSchema(t *testing.T) {
	t.Parallel()

	resourceSchema := percentageDisplayFormatConfigurationSchema()
	expectedDataSourceSchema := sdkv2.ComputedOnlyFromSchema(resourceSchema)

	dataSourceSchema := percentageDisplayFormatConfigurationDataSourceSchema()

	if diff := cmp.Diff(dataSourceSchema, expectedDataSourceSchema); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}

func TestNumberFormatConfigurationDataSourceSchema(t *testing.T) {
	t.Parallel()

	resourceSchema := numberFormatConfigurationSchema()
	expectedDataSourceSchema := sdkv2.ComputedOnlyFromSchema(resourceSchema)

	dataSourceSchema := numberFormatConfigurationDataSourceSchema()

	if diff := cmp.Diff(dataSourceSchema, expectedDataSourceSchema); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}

func TestStringFormatConfigurationDataSourceSchema(t *testing.T) {
	t.Parallel()

	resourceSchema := stringFormatConfigurationSchema()
	expectedDataSourceSchema := sdkv2.ComputedOnlyFromSchema(resourceSchema)

	dataSourceSchema := stringFormatConfigurationDataSourceSchema()

	if diff := cmp.Diff(dataSourceSchema, expectedDataSourceSchema); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}

func TestDecimalPlacesConfigurationDataSourceSchema(t *testing.T) {
	t.Parallel()

	resourceSchema := decimalPlacesConfigurationSchema()
	expectedDataSourceSchema := sdkv2.ComputedOnlyFromSchema(resourceSchema)

	dataSourceSchema := decimalPlacesConfigurationDataSourceSchema()

	if diff := cmp.Diff(dataSourceSchema, expectedDataSourceSchema); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}

func TestNegativeValueConfigurationDataSourceSchema(t *testing.T) {
	t.Parallel()

	resourceSchema := negativeValueConfigurationSchema()
	expectedDataSourceSchema := sdkv2.ComputedOnlyFromSchema(resourceSchema)

	dataSourceSchema := negativeValueConfigurationDataSourceSchema()

	if diff := cmp.Diff(dataSourceSchema, expectedDataSourceSchema); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}

func TestNullValueConfigurationDataSourceSchema(t *testing.T) {
	t.Parallel()

	resourceSchema := nullValueConfigurationSchema()
	expectedDataSourceSchema := sdkv2.ComputedOnlyFromSchema(resourceSchema)

	dataSourceSchema := nullValueConfigurationDataSourceSchema()

	if diff := cmp.Diff(dataSourceSchema, expectedDataSourceSchema); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}

func TestSeparatorConfigurationDataSourceSchema(t *testing.T) {
	t.Parallel()

	resourceSchema := separatorConfigurationSchema()
	expectedDataSourceSchema := sdkv2.ComputedOnlyFromSchema(resourceSchema)

	dataSourceSchema := separatorConfigurationDataSourceSchema()

	if diff := cmp.Diff(dataSourceSchema, expectedDataSourceSchema); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}

func TestLabelOptionsDataSourceSchema(t *testing.T) {
	t.Parallel()

	resourceSchema := labelOptionsSchema()
	expectedDataSourceSchema := sdkv2.ComputedOnlyFromSchema(resourceSchema)

	dataSourceSchema := labelOptionsDataSourceSchema()

	if diff := cmp.Diff(dataSourceSchema, expectedDataSourceSchema); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}

func TestFontConfigurationDataSourceSchema(t *testing.T) {
	t.Parallel()

	resourceSchema := fontConfigurationSchema()
	expectedDataSourceSchema := sdkv2.ComputedOnlyFromSchema(resourceSchema)

	dataSourceSchema := fontConfigurationDataSourceSchema()

	if diff := cmp.Diff(dataSourceSchema, expectedDataSourceSchema); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}

func TestFormatConfigurationDataSourceSchema(t *testing.T) {
	t.Parallel()

	resourceSchema := formatConfigurationSchema()
	expectedDataSourceSchema := sdkv2.ComputedOnlyFromSchema(resourceSchema)

	dataSourceSchema := formatConfigurationDataSourceSchema()

	if diff := cmp.Diff(dataSourceSchema, expectedDataSourceSchema); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}
