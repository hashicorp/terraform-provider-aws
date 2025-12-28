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
type Lister interface {
	AppendResultInterceptor(listresource.ListResultInterceptor)
}

var _ Lister = &WithList{}

// WithList provides common functionality for ListResources
type WithList struct {
	withListResourceConfigSchema
	interceptors []listresource.ListResultInterceptor
}

type flattenFunc func()

func (w *WithList) AppendResultInterceptor(interceptor listresource.ListResultInterceptor) {
	w.interceptors = append(w.interceptors, interceptor)
}

func (w WithList) ResultInterceptors() []listresource.ListResultInterceptor {
	return w.interceptors
}

func (w *WithList) runResultInterceptors(ctx context.Context, when listresource.When, awsClient *conns.AWSClient, data any, result *list.ListResult) diag.Diagnostics {
	var diags diag.Diagnostics
	params := listresource.InterceptorParams{
		C:      awsClient,
		Result: result,
		Data:   data,
	}

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

func (w *WithList) SetResult(ctx context.Context, awsClient *conns.AWSClient, data any, result *list.ListResult, f flattenFunc) {
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
