// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type timestampType struct {
	basetypes.StringType
}

var (
	TimestampType = timestampType{}
)

var (
	_ xattr.TypeWithValidate  = (*timestampType)(nil)
	_ basetypes.StringTypable = (*timestampType)(nil)
)

func (t timestampType) Equal(o attr.Type) bool {
	other, ok := o.(timestampType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (timestampType) String() string {
	return "TimestampType"
}

func (t timestampType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsUnknown() {
		return TimestampUnknown(), diags
	}

	if in.IsNull() {
		return TimestampNull(), diags
	}

	valueString := in.ValueString()
	if _, err := time.Parse(time.RFC3339, valueString); err != nil {
		return TimestampUnknown(), diags // Must not return validation errors
	}

	return TimestampValue(valueString), diags
}

func (t timestampType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)

	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}

func (timestampType) ValueType(context.Context) attr.Value {
	return Timestamp{}
}

func (t timestampType) Validate(ctx context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if !in.IsKnown() || in.IsNull() {
		return diags
	}

	var value string
	err := in.As(&value)
	if err != nil {
		diags.AddAttributeError(
			path,
			"Timestamp Type Validation Error",
			ProviderErrorDetailPrefix+fmt.Sprintf("Cannot convert value to string: %s", err),
		)
		return diags
	}

	if _, err = time.Parse(time.RFC3339, value); err != nil {
		diags.AddAttributeError(
			path,
			"Timestamp Type Validation Error",
			fmt.Sprintf("Value %q cannot be parsed as an RFC 3339 Timestamp.", value),
		)
		return diags
	}

	return diags
}
