// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/listresource"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type Lister interface {
	AppendResultInterceptor(listresource.ListResultInterceptor)
}

var _ Lister = &WithList{}

type WithList struct {
	interceptors []listresource.ListResultInterceptor
}

func (w *WithList) AppendResultInterceptor(interceptor listresource.ListResultInterceptor) {
	w.interceptors = append(w.interceptors, interceptor)
}

func (w WithList) ResultInterceptors() []listresource.ListResultInterceptor {
	return w.interceptors
}

func (w *WithList) RunResultInterceptors(ctx context.Context, when listresource.When, params listresource.InterceptorParams) diag.Diagnostics {
	var diags diag.Diagnostics
	switch when {
	case listresource.Before:
		params.When = listresource.Before
		for interceptor := range slices.Values(w.interceptors) {
			diags.Append(interceptor.Read(ctx, params)...)
		}
		return diags
	case listresource.After:
		params.When = listresource.After
		for interceptor := range tfslices.BackwardValues(w.interceptors) {
			diags.Append(interceptor.Read(ctx, params)...)
		}
		return diags
	}

	return diags
}

func (w *WithList) ListResourceTagsInit(ctx context.Context, result list.ListResult) basetypes.MapValue {
	typ, _ := result.Resource.Schema.TypeAtPath(ctx, path.Root(names.AttrTags))
	tagsType := typ.(attr.TypeWithElementType)

	return basetypes.NewMapNull(tagsType.ElementType())
}

func (w *WithList) ListResourceTimeoutInit(ctx context.Context, result list.ListResult) (basetypes.ObjectValue, diag.Diagnostics) {
	timeoutsType, _ := result.Resource.Schema.TypeAtPath(ctx, path.Root(names.AttrTimeouts))

	return newNullObject(timeoutsType)
}

func newNullObject(typ attr.Type) (obj basetypes.ObjectValue, diags diag.Diagnostics) {
	i, ok := typ.(attr.TypeWithAttributeTypes)
	if !ok {
		diags.AddError(
			"Internal Error",
			"An unexpected error occurred. "+
				"This is always an error in the provider. "+
				"Please report the following to the provider developer:\n\n"+
				fmt.Sprintf("Expected value type to implement attr.TypeWithAttributeTypes, got: %T", typ),
		)
		return
	}

	attrTypes := i.AttributeTypes()

	obj = basetypes.NewObjectNull(attrTypes)

	return obj, diags
}
