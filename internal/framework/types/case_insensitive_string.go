// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringTypable = (*caseInsensitiveStringType)(nil)
)

type caseInsensitiveStringType struct {
	basetypes.StringType
}

var (
	CaseInsensitiveStringType = caseInsensitiveStringType{}
)

func (t caseInsensitiveStringType) Equal(o attr.Type) bool {
	other, ok := o.(caseInsensitiveStringType)
	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (caseInsensitiveStringType) String() string {
	return "CaseInsensitiveStringType"
}

func (t caseInsensitiveStringType) ValueFromString(_ context.Context, in types.String) (basetypes.StringValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return CaseInsensitiveStringNull(), diags
	}
	if in.IsUnknown() {
		return CaseInsensitiveStringUnknown(), diags
	}

	return CaseInsensitiveStringValue(in.ValueString()), diags
}

func (t caseInsensitiveStringType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (caseInsensitiveStringType) ValueType(context.Context) attr.Value {
	return CaseInsensitiveString{}
}

var (
	_ basetypes.StringValuable                   = (*CaseInsensitiveString)(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*CaseInsensitiveString)(nil)
)

type CaseInsensitiveString struct {
	basetypes.StringValue
}

func CaseInsensitiveStringNull() CaseInsensitiveString {
	return CaseInsensitiveString{StringValue: basetypes.NewStringNull()}
}

func CaseInsensitiveStringUnknown() CaseInsensitiveString {
	return CaseInsensitiveString{StringValue: basetypes.NewStringUnknown()}
}

func CaseInsensitiveStringValue(value string) CaseInsensitiveString {
	return CaseInsensitiveString{StringValue: basetypes.NewStringValue(value)}
}

func (v CaseInsensitiveString) Equal(o attr.Value) bool {
	other, ok := o.(CaseInsensitiveString)
	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (CaseInsensitiveString) Type(context.Context) attr.Type {
	return CaseInsensitiveStringType
}

func (v CaseInsensitiveString) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	return caseInsensitiveStringSemanticEquals(ctx, v, newValuable)
}

// caseInsensitiveStringSemanticEquals returns whether oldValuable and newValuable are equal under simple Unicode case-folding.
func caseInsensitiveStringSemanticEquals[T basetypes.StringValuable](ctx context.Context, oldValuable T, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(T)
	if !ok {
		return false, diags
	}

	old, d := oldValuable.ToStringValue(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	new, d := newValue.ToStringValue(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	// Case insensitive comparison.
	return strings.EqualFold(old.ValueString(), new.ValueString()), diags
}
