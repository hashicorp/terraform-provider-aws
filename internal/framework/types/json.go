// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type jsonType struct {
	basetypes.StringType
}

var (
	JSONType = jsonType{}
)

var (
	_ xattr.TypeWithValidate  = (*jsonType)(nil)
	_ basetypes.StringTypable = (*jsonType)(nil)
)

func (j jsonType) Equal(o attr.Type) bool {
	other, ok := o.(jsonType)

	if !ok {
		return false
	}

	return j.StringType.Equal(other.StringType)
}

func (j jsonType) String() string {
	return "JSONType"
}

func (j jsonType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsUnknown() {
		return JSONUnknown(), diags
	}

	if in.IsNull() {
		return JSONNull(), diags
	}

	valueString := in.ValueString()
	output, err := normalizeJSON(valueString)
	if err != nil {
		return JSONUnknown(), diags
	}
	return JSONValue(output), diags
}

func (j jsonType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := j.StringType.ValueFromTerraform(ctx, in)

	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := j.ValueFromString(ctx, stringValue)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}

func (j jsonType) ValueType(context.Context) attr.Value {
	return JSON{}
}

func (j jsonType) Validate(ctx context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if !in.IsKnown() || in.IsNull() {
		return diags
	}

	var value string
	err := in.As(&value)
	if err != nil {
		diags.AddAttributeError(
			path,
			"JSON Type Validation Error",
			ProviderErrorDetailPrefix+fmt.Sprintf("Cannot convert value to string: %s", err),
		)
		return diags
	}

	if !json.Valid([]byte(value)) {
		diags.AddAttributeError(
			path,
			"JSON Type Validation Error",
			fmt.Sprint("Value must be valid JSON"),
		)
		return diags
	}

	return diags
}

var (
	_ basetypes.StringValuable = (*JSON)(nil)
)

type JSON struct {
	basetypes.StringValue
}

func (j JSON) Equal(o attr.Value) bool {
	other, ok := o.(JSON)

	if !ok {
		return false
	}

	return j.StringValue.Equal(other.StringValue)
}

func (j JSON) Type(_ context.Context) attr.Type {
	return JSONType
}

func (j JSON) Normalize() (string, diag.Diagnostics) {
	var diags diag.Diagnostics

	output, err := normalizeJSON(j.ValueString())

	if err != nil {
		diags.AddError(
			"unable to normalize JSON value",
			err.Error(),
		)
		return output, diags
	}

	return output, diags
}

func JSONNull() JSON {
	return JSON{StringValue: types.StringNull()}
}

func JSONUnknown() JSON {
	return JSON{StringValue: types.StringUnknown()}
}

func JSONValue(value string) JSON {
	return JSON{
		StringValue: basetypes.NewStringValue(value),
	}
}

func normalizeJSON(value string) (string, error) {
	var data interface{}

	err := json.Unmarshal([]byte(value), &data)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(output), err
}
