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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

var (
	_ xattr.TypeWithValidate   = (*durationType)(nil)
	_ basetypes.StringTypable  = (*durationType)(nil)
	_ basetypes.StringValuable = (*Duration)(nil)
)

type durationType struct {
	basetypes.StringType
}

var (
	DurationType = durationType{}
)

func (t durationType) Equal(o attr.Type) bool {
	other, ok := o.(durationType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (durationType) String() string {
	return "DurationType"
}

func (t durationType) ValueFromString(_ context.Context, in types.String) (basetypes.StringValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return DurationNull(), diags
	}
	if in.IsUnknown() {
		return DurationUnknown(), diags
	}

	valueString := in.ValueString()
	if _, err := time.ParseDuration(valueString); err != nil {
		return DurationUnknown(), diags // Must not return validation errors
	}

	return DurationValueMust(valueString), diags
}

func (t durationType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (durationType) ValueType(context.Context) attr.Value {
	return Duration{}
}

func (t durationType) Validate(ctx context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if !in.IsKnown() || in.IsNull() {
		return diags
	}

	var value string
	err := in.As(&value)
	if err != nil {
		diags.AddAttributeError(
			path,
			"Duration Type Validation Error",
			ProviderErrorDetailPrefix+fmt.Sprintf("Cannot convert value to string: %s", err),
		)
		return diags
	}

	if _, err = time.ParseDuration(value); err != nil {
		diags.AddAttributeError(
			path,
			"Duration Type Validation Error",
			fmt.Sprintf("Value %q cannot be parsed as a Duration.", value),
		)
		return diags
	}

	return diags
}

func DurationNull() Duration {
	return Duration{StringValue: basetypes.NewStringNull()}
}

func DurationUnknown() Duration {
	return Duration{StringValue: basetypes.NewStringUnknown()}
}

func DurationValueMust(value string) Duration {
	return Duration{
		StringValue: basetypes.NewStringValue(value),
		value:       errs.Must(time.ParseDuration(value)),
	}
}

type Duration struct {
	basetypes.StringValue
	value time.Duration
}

func (v Duration) Equal(o attr.Value) bool {
	other, ok := o.(Duration)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (Duration) Type(context.Context) attr.Type {
	return DurationType
}

// ValueDuration returns the known time.Duration value. If Duration is null or unknown, returns 0.
func (v Duration) ValueDuration() time.Duration {
	return v.value
}
