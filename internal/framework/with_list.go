// Copyright IBM Corp. 2014, 2026
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

var _ Lister[listresource.InterceptorParams] = &WithList{}

// WithList provides common functionality for ListResources
type WithList struct {
	withListResourceConfigSchema
	interceptors []listresource.ListResultInterceptor[listresource.InterceptorParams]
}

type flattenFunc func()

func (w *WithList) AppendResultInterceptor(interceptor listresource.ListResultInterceptor[listresource.InterceptorParams]) {
	w.interceptors = append(w.interceptors, interceptor)
}

func (w WithList) ResultInterceptors() []listresource.ListResultInterceptor[listresource.InterceptorParams] {
	return w.interceptors
}

func (w *WithList) runResultInterceptors(ctx context.Context, when listresource.When, awsClient *conns.AWSClient, includeResource bool, data any, result *list.ListResult) diag.Diagnostics {
	var diags diag.Diagnostics
	params := listresource.InterceptorParams{
		C:               awsClient,
		IncludeResource: includeResource,
		Result:          result,
		Data:            data,
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

func (w *WithList) SetResult(ctx context.Context, awsClient *conns.AWSClient, includeResource bool, data any, result *list.ListResult, f flattenFunc) {
	var diags diag.Diagnostics

	diags.Append(w.runResultInterceptors(ctx, listresource.Before, awsClient, includeResource, data, result)...)
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

	diags.Append(w.runResultInterceptors(ctx, listresource.After, awsClient, includeResource, data, result)...)
	if diags.HasError() {
		return
	}
}
