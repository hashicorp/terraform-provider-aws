// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package savingsplans

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/shopspring/decimal"
)

var (
	_ basetypes.StringTypable = (*commitmentStringType)(nil)
)

type commitmentStringType struct {
	basetypes.StringType
}

var (
	CommitmentStringType = commitmentStringType{}
)

func (t commitmentStringType) Equal(o attr.Type) bool {
	other, ok := o.(commitmentStringType)
	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (commitmentStringType) String() string {
	return "CommitmentStringType"
}

func (t commitmentStringType) ValueFromString(_ context.Context, in types.String) (basetypes.StringValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return CommitmentStringNull(), diags
	}
	if in.IsUnknown() {
		return CommitmentStringUnknown(), diags
	}

	return CommitmentStringValue(in.ValueString()), diags
}

func (t commitmentStringType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (commitmentStringType) ValueType(context.Context) attr.Value {
	return CommitmentString{}
}

var (
	_ basetypes.StringValuable                   = (*CommitmentString)(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*CommitmentString)(nil)
)

type CommitmentString struct {
	basetypes.StringValue
}

func CommitmentStringNull() CommitmentString {
	return CommitmentString{StringValue: basetypes.NewStringNull()}
}

func CommitmentStringUnknown() CommitmentString {
	return CommitmentString{StringValue: basetypes.NewStringUnknown()}
}

func CommitmentStringValue(value string) CommitmentString {
	return CommitmentString{StringValue: basetypes.NewStringValue(value)}
}

func (v CommitmentString) Equal(o attr.Value) bool {
	other, ok := o.(CommitmentString)
	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (CommitmentString) Type(context.Context) attr.Type {
	return CommitmentStringType
}

func (v CommitmentString) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	oldValue, d := v.ToStringValue(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	newValue, d := newValuable.ToStringValue(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	return commitmentStringSemanticEquals(oldValue.ValueString(), newValue.ValueString()), diags
}

func commitmentStringSemanticEquals(oldValue, newValue string) bool {
	if oldValue == newValue {
		return true
	}

	oldDecimal, oldErr := decimal.NewFromString(oldValue)
	newDecimal, newErr := decimal.NewFromString(newValue)
	if oldErr != nil || newErr != nil {
		return false
	}

	return oldDecimal.Equal(newDecimal)
}
