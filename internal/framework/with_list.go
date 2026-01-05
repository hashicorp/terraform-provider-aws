// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/listresource"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

// Lister is an interface for resources that support List operations
type Lister[T listresource.InterceptorParams | listresource.InterceptorParamsSDK] interface {
	AppendResultInterceptor(listresource.ListResultInterceptor[T])
}

var _ Lister[listresource.InterceptorParams] = &withList[listresource.InterceptorParams]{}

type WithList = withList[listresource.InterceptorParams]

// WithList provides common functionality for ListResources
type withList[T listresource.InterceptorParams] struct {
	withListResourceConfigSchema
	interceptors []listresource.ListResultInterceptor[T]
}

type flattenFunc func()

func (w *withList[T]) AppendResultInterceptor(interceptor listresource.ListResultInterceptor[T]) {
	w.interceptors = append(w.interceptors, interceptor)
}

func (w withList[T]) ResultInterceptors() []listresource.ListResultInterceptor[T] {
	return w.interceptors
}

func (w *withList[T]) runResultInterceptors(ctx context.Context, when listresource.When, awsClient *conns.AWSClient, data any, result *list.ListResult) diag.Diagnostics {
	var diags diag.Diagnostics
	var params any

	switch when {
	case listresource.Before:
		params = listresource.InterceptorParams{
			C:      awsClient,
			Result: result,
			Data:   data,
			When:   when,
		}
		for interceptor := range slices.Values(w.interceptors) {
			diags.Append(interceptor.Read(ctx, params.(T))...)
		}
		return diags
	case listresource.After:
		params = listresource.InterceptorParams{
			C:      awsClient,
			Result: result,
			Data:   data,
			When:   when,
		}
		for interceptor := range tfslices.BackwardValues(w.interceptors) {
			diags.Append(interceptor.Read(ctx, params.(T))...)
		}
		return diags
	}

	return diags
}

func (w *withList[T]) SetResult(ctx context.Context, awsClient *conns.AWSClient, data any, result *list.ListResult, f flattenFunc) {
	var diags diag.Diagnostics

	diags.Append(w.runResultInterceptors(ctx, listresource.Before, awsClient, data, result)...)
	if diags.HasError() {
		return
	}

	f()
	if result.Diagnostics.HasError() {
		return
	}

	diags.Append(result.Resource.Set(ctx, data)...)
	if diags.HasError() {
		return
	}

	diags.Append(w.runResultInterceptors(ctx, listresource.After, awsClient, data, result)...)
	if diags.HasError() {
		return
	}
}
