// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

type unusedConfigListModel struct {
	framework.WithRegionModel
}

type unusedConfigListResource struct{}

func (r *unusedConfigListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	// ruleid: list-resource-unused-config-model
	var query unusedConfigListModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {}
}

type configuredListModel struct {
	framework.WithRegionModel
	Name string
}

type configuredListResource struct{}

func (r *configuredListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	// ok: list-resource-unused-config-model
	var query configuredListModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		_ = query.Name
	}
}

type expandedListResource struct{}

// The model is expanded into the AWS SDK input struct, so the configuration is used.
func (r *expandedListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	// ok: list-resource-unused-config-model
	var query configuredListModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input listInput
	if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(yield func(list.ListResult) bool) {}
}

type expandedByPointerListResource struct{}

// The model is passed by pointer to something other than `request.Config.Get`,
// so the configuration is used.
func (r *expandedByPointerListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	// ok: list-resource-unused-config-model
	var query configuredListModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input listInput
	if diags := fwflex.Expand(ctx, &query, &input); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(yield func(list.ListResult) bool) {}
}

type nestedUseListResource struct{}

// The model field is read inside a composite literal nested in a closure.
func (r *nestedUseListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	// ok: list-resource-unused-config-model
	var query configuredListModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		input := listInput{
			Name: query.Name,
		}
		_ = input
	}
}

type listInput struct {
	Name string
}
