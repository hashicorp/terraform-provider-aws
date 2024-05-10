// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringTypable = (*iamPolicyType)(nil)
)

type iamPolicyType struct {
	basetypes.StringType
}

var (
	IAMPolicyType = iamPolicyType{}
)

func (t iamPolicyType) Equal(o attr.Type) bool {
	other, ok := o.(iamPolicyType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t iamPolicyType) String() string {
	return "IAMPolicyType"
}

func (t iamPolicyType) ValueFromString(_ context.Context, in types.String) (basetypes.StringValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return IAMPolicyNull(), diags
	}
	if in.IsUnknown() {
		return IAMPolicyUnknown(), diags
	}

	return IAMPolicy{StringValue: in}, diags
}

func (t iamPolicyType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (t iamPolicyType) ValueType(context.Context) attr.Value {
	return IAMPolicy{}
}

var (
	_ basetypes.StringValuable                   = (*IAMPolicy)(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*IAMPolicy)(nil)
	_ xattr.ValidateableAttribute                = (*IAMPolicy)(nil)
)

func IAMPolicyNull() IAMPolicy {
	return IAMPolicy{StringValue: basetypes.NewStringNull()}
}

func IAMPolicyUnknown() IAMPolicy {
	return IAMPolicy{StringValue: basetypes.NewStringUnknown()}
}

func IAMPolicyValue(value string) IAMPolicy {
	return IAMPolicy{StringValue: basetypes.NewStringValue(value)}
}

type IAMPolicy struct {
	basetypes.StringValue
}

func (v IAMPolicy) Equal(o attr.Value) bool {
	other, ok := o.(IAMPolicy)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (v IAMPolicy) Type(context.Context) attr.Type {
	return IAMPolicyType
}

func (v IAMPolicy) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(IAMPolicy)

	if !ok {
		return false, diags
	}

	return policyStringsEquivalent(v.ValueString(), newValue.ValueString()), diags
}

// See verify.PolicyStringsEquivalent, which can't be called because of import cycles.
func policyStringsEquivalent(s1, s2 string) bool {
	if strings.TrimSpace(s1) == "" && strings.TrimSpace(s2) == "" {
		return true
	}

	if strings.TrimSpace(s1) == "{}" && strings.TrimSpace(s2) == "" {
		return true
	}

	if strings.TrimSpace(s1) == "" && strings.TrimSpace(s2) == "{}" {
		return true
	}

	if strings.TrimSpace(s1) == "{}" && strings.TrimSpace(s2) == "{}" {
		return true
	}

	equivalent, err := awspolicy.PoliciesAreEquivalent(s1, s2)
	if err != nil {
		return false
	}

	return equivalent
}

func (v IAMPolicy) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if v.IsNull() || v.IsUnknown() {
		return
	}

	if !json.Valid([]byte(v.ValueString())) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid IAM Policy Value",
			"The provided value is not valid JSON string format (RFC 7159).\n\n"+
				"Path: "+req.Path.String()+"\n"+
				"Value: "+v.ValueString(),
		)
	}
}
