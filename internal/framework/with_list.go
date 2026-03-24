// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/listresource"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

// Lister is an interface for resources that support List operations
type Lister[T listresource.InterceptorParams | listresource.InterceptorParamsSDK] interface {
	AppendResultInterceptor(listresource.ListResultInterceptor[T])
}

var _ Lister[listresource.InterceptorParams] = &WithList[any]{}

// WithList provides common functionality for ListResources.
// `T` is type that is flattened into. Usually the resource model.
type WithList[T any] struct {
	withListResourceConfigSchema
	interceptors []listresource.ListResultInterceptor[listresource.InterceptorParams]
}

type FlattenFunc[T any] func(context.Context, *T)

func (w *WithList[T]) AppendResultInterceptor(interceptor listresource.ListResultInterceptor[listresource.InterceptorParams]) {
	w.interceptors = append(w.interceptors, interceptor)
}

func (w *WithList[T]) runResultInterceptors(ctx context.Context, when listresource.When, awsClient *conns.AWSClient, includeResource bool, data any, result *list.ListResult) diag.Diagnostics {
	var diags diag.Diagnostics
	params := listresource.InterceptorParams{
		C:               awsClient,
		IncludeResource: includeResource,
		Data:            data,
		Result:          result,
		When:            when,
	}

	switch when {
	case listresource.Before:
		for interceptor := range slices.Values(w.interceptors) {
			diags.Append(interceptor.Read(ctx, params)...)
		}
	case listresource.After:
		for interceptor := range tfslices.BackwardValues(w.interceptors) {
			diags.Append(interceptor.Read(ctx, params)...)
		}
	}

	return diags
}

func (w *WithList[T]) SetResult(ctx context.Context, awsClient *conns.AWSClient, includeResource bool, result *list.ListResult, f FlattenFunc[T]) T {
	var diags diag.Diagnostics
	var data T

	diags.Append(fwtypes.NullOutObjectPtrFields(ctx, &data)...)
	if diags.HasError() {
		result.Diagnostics.Append(diags...)
		return data
	}

	diags.Append(w.runResultInterceptors(ctx, listresource.Before, awsClient, includeResource, &data, result)...)
	if diags.HasError() {
		result.Diagnostics.Append(diags...)
		return data
	}

	f(ctx, &data)
	if result.Diagnostics.HasError() {
		return data
	}

	diags.Append(result.Resource.Set(ctx, &data)...)
	if diags.HasError() {
		result.Diagnostics.Append(diags...)
		return data
	}

	diags.Append(w.runResultInterceptors(ctx, listresource.After, awsClient, includeResource, &data, result)...)
	if diags.HasError() {
		result.Diagnostics.Append(diags...)
		return data
	}

	return data
}
