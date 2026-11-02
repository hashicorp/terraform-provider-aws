// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// ListEmptied returns true if the old value is a non-empty list and the new value is an empty list.
// Null values are treated as empty lists.
// Unknown values panic.
func ListEmptied(ctx context.Context, old, new basetypes.ListValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	o, d := old.ToListValue(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}
	n, d := new.ToListValue(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	if oLen, nLen := o.Length(basetypes.CollectionLengthOptions{UnhandledNullAsZero: true}), n.Length(basetypes.CollectionLengthOptions{UnhandledNullAsZero: true}); oLen > 0 && nLen == 0 {
		return true, diags
	}
	return false, diags
}
