// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

type writeOnlyAttrGetter interface {
	GetRawConfigAt(path cty.Path) (cty.Value, diag.Diagnostics)
	Id() string
}

// GetWriteOnlyStringValue returns the string value of the write-only attribute from the config.
func GetWriteOnlyStringValue(d writeOnlyAttrGetter, path cty.Path) (string, diag.Diagnostics) {
	valueWO, diags := GetWriteOnlyValue(d, path, cty.String)
	if diags.HasError() {
		return "", diags
	}

	var value string
	if !valueWO.IsNull() {
		value = valueWO.AsString()
	}

	return value, diags
}

// GetWriteOnlyValue returns the value of the write-only attribute from the config.
func GetWriteOnlyValue(d writeOnlyAttrGetter, path cty.Path, attrType cty.Type) (cty.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	valueWO, di := d.GetRawConfigAt(path)
	if di.HasError() {
		diags = append(diags, di...)
		return cty.Value{}, diags
	}

	if !valueWO.Type().Equals(attrType) {
		return cty.Value{}, sdkdiag.AppendErrorf(diags, "invalid type (%s) for resource(%s)", attrType, d.Id())
	}

	return valueWO, diags
}
