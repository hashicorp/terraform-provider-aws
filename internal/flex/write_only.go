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

// GetWriteOnlyValue returns the value of the write-only attribute from the config.
func GetWriteOnlyValue(d writeOnlyAttrGetter, path cty.Path, attrType cty.Type) (cty.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	valueWO, di := d.GetRawConfigAt(path)
	if di.HasError() {
		diags = append(diags, di...)
		return cty.Value{}, diags
	}

	if !valueWO.Type().Equals(attrType) {
		return cty.Value{}, sdkdiag.AppendErrorf(diags, "SSM Parameter (%s): invalid value_wo type", d.Id())
	}

	return valueWO, diags
}
