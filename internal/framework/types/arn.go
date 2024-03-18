// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
)

var (
	_ xattr.TypeWithValidate   = (*arnType)(nil)
	_ basetypes.StringTypable  = (*arnType)(nil)
	_ basetypes.StringValuable = (*ARN)(nil)
)

type arnType struct {
	basetypes.StringType
}

var (
	ARNType = arnType{}
)

func (t arnType) Equal(o attr.Type) bool {
	other, ok := o.(arnType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (arnType) String() string {
	return "ARNType"
}

func (t arnType) ValueFromString(_ context.Context, in types.String) (basetypes.StringValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return ARNNull(), diags
	}
	if in.IsUnknown() {
		return ARNUnknown(), diags
	}

	valueString := in.ValueString()
	if _, err := arn.Parse(valueString); err != nil {
		return ARNUnknown(), diags // Must not return validation errors.
	}

	return ARNValueMust(valueString), diags
}

func (t arnType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (arnType) ValueType(context.Context) attr.Value {
	return ARN{}
}

func (t arnType) Validate(ctx context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if !in.IsKnown() || in.IsNull() {
		return diags
	}

	var value string
	err := in.As(&value)
	if err != nil {
		diags.AddAttributeError(
			path,
			"ARN Type Validation Error",
			ProviderErrorDetailPrefix+fmt.Sprintf("Cannot convert value to string: %s", err),
		)
		return diags
	}

	if !arn.IsARN(value) {
		diags.AddAttributeError(
			path,
			"ARN Type Validation Error",
			fmt.Sprintf("Value %q cannot be parsed as an ARN.", value),
		)
		return diags
	}

	return diags
}

func ARNNull() ARN {
	return ARN{StringValue: basetypes.NewStringNull()}
}

func ARNUnknown() ARN {
	return ARN{StringValue: basetypes.NewStringUnknown()}
}

func ARNValue(value string) (ARN, diag.Diagnostics) {
	var diags diag.Diagnostics

	v, err := arn.Parse(value)
	if err != nil {
		diags.Append(diag.NewErrorDiagnostic("Invalid ARN", err.Error()))
		return ARNUnknown(), diags
	}

	return ARN{
		StringValue: basetypes.NewStringValue(value),
		value:       v,
	}, diags
}

func ARNValueMust(value string) ARN {
	return fwdiag.Must(ARNValue(value))
}

type ARN struct {
	basetypes.StringValue
	value arn.ARN
}

func (v ARN) Equal(o attr.Value) bool {
	other, ok := o.(ARN)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (ARN) Type(context.Context) attr.Type {
	return ARNType
}

// ValueARN returns the known arn.ARN value. If ARN is null or unknown, returns {}.
func (v ARN) ValueARN() arn.ARN {
	return v.value
}
