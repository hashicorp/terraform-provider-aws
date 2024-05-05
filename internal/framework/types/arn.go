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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringTypable = (*arnType)(nil)
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

	return ARNValue(valueString), diags
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

var (
	_ basetypes.StringValuable    = (*ARN)(nil)
	_ xattr.ValidateableAttribute = (*ARN)(nil)
)

func ARNNull() ARN {
	return ARN{StringValue: basetypes.NewStringNull()}
}

func ARNUnknown() ARN {
	return ARN{StringValue: basetypes.NewStringUnknown()}
}

// ARNValue initializes a new ARN type with the provided value
//
// This function does not return diagnostics, and therefore invalid ARN values
// are not handled during construction. Invalid values will be detected by the
// ValidateAttribute method, called by the ValidateResourceConfig RPC during
// operations like `terraform validate`, `plan`, or `apply`.
func ARNValue(value string) ARN {
	// swallow any ARN parsing errors here and just pass along the
	// zero value arn.ARN. Invalid values will be handled downstream
	// by the ValidateAttribute method.
	v, _ := arn.Parse(value)

	return ARN{
		StringValue: basetypes.NewStringValue(value),
		value:       v,
	}
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

// ValueARN returns the known arn.ARN value. If ARN is null, unknown, or invalid returns ARN{}.
func (v ARN) ValueARN() arn.ARN {
	return v.value
}

func (v ARN) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if v.IsNull() || v.IsUnknown() {
		return
	}

	if !arn.IsARN(v.ValueString()) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid ARN Value",
			"The provided value cannot be parsed as an ARN.\n\n"+
				"Path: "+req.Path.String()+"\n"+
				"Value: "+v.ValueString(),
		)
	}
}
