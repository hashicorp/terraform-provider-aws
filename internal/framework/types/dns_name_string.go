// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-aws/internal/dns"
)

var (
	_ basetypes.StringTypable = (*dnsNameStringType)(nil)
)

type dnsNameStringType struct {
	basetypes.StringType
}

var (
	DNSNameStringType = dnsNameStringType{}
)

func (t dnsNameStringType) Equal(o attr.Type) bool {
	other, ok := o.(dnsNameStringType)
	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (dnsNameStringType) String() string {
	return "DNSNameStringType"
}

func (t dnsNameStringType) ValueFromString(_ context.Context, in types.String) (basetypes.StringValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return DNSNameStringNull(), diags
	}
	if in.IsUnknown() {
		return DNSNameStringUnknown(), diags
	}

	return DNSNameStringValue(in.ValueString()), diags
}

func (t dnsNameStringType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (dnsNameStringType) ValueType(context.Context) attr.Value {
	return DNSNameString{}
}

var (
	_ basetypes.StringValuable                   = (*DNSNameString)(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*DNSNameString)(nil)
)

type DNSNameString struct {
	basetypes.StringValue
}

func DNSNameStringNull() DNSNameString {
	return DNSNameString{StringValue: basetypes.NewStringNull()}
}

func DNSNameStringUnknown() DNSNameString {
	return DNSNameString{StringValue: basetypes.NewStringUnknown()}
}

func DNSNameStringValue(value string) DNSNameString {
	return DNSNameString{StringValue: basetypes.NewStringValue(value)}
}

func (v DNSNameString) Equal(o attr.Value) bool {
	other, ok := o.(DNSNameString)
	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (DNSNameString) Type(context.Context) attr.Type {
	return DNSNameStringType
}

func (v DNSNameString) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(DNSNameString)
	if !ok {
		return false, diags
	}

	old, d := v.ToStringValue(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	new, d := newValue.ToStringValue(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	return dns.Normalize(old.ValueString()) == dns.Normalize(new.ValueString()), diags
}
