// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var (
	_ WithMeta = (*withMeta)(nil)
)

type WithMeta interface {
	Meta() *conns.AWSClient
}

type withMeta struct {
	meta *conns.AWSClient
}

func (w *withMeta) Meta() *conns.AWSClient {
	return w.meta
}

// RegionalARN returns a regional ARN for the specified service namespace and resource.
func (w *withMeta) RegionalARN(service, resource string) string {
	return arn.ARN{
		Partition: w.meta.Partition,
		Service:   service,
		Region:    w.meta.Region,
		AccountID: w.meta.AccountID,
		Resource:  resource,
	}.String()
}

// ResourceWithConfigure is a structure to be embedded within a Resource that implements the ResourceWithConfigure interface.
type ResourceWithConfigure struct {
	withMeta
}

// Configure enables provider-level data or clients to be set in the
// provider-defined Resource type.
func (r *ResourceWithConfigure) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		r.meta = v
	}
}

// SetTagsAll calculates the new value for the `tags_all` attribute.
func (r *ResourceWithConfigure) SetTagsAll(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	// If the entire plan is null, the resource is planned for destruction.
	if request.Plan.Raw.IsNull() {
		return
	}

	defaultTagsConfig := r.Meta().DefaultTagsConfig
	ignoreTagsConfig := r.Meta().IgnoreTagsConfig

	var planTags types.Map

	response.Diagnostics.Append(request.Plan.GetAttribute(ctx, path.Root(names.AttrTags), &planTags)...)

	if response.Diagnostics.HasError() {
		return
	}

	if !planTags.IsUnknown() {
		if !mapHasUnknownElements(planTags) {
			resourceTags := tftags.New(ctx, planTags)
			allTags := defaultTagsConfig.MergeTags(resourceTags).IgnoreConfig(ignoreTagsConfig)

			response.Diagnostics.Append(response.Plan.SetAttribute(ctx, path.Root(names.AttrTagsAll), flex.FlattenFrameworkStringValueMapLegacy(ctx, allTags.Map()))...)
		} else {
			response.Diagnostics.Append(response.Plan.SetAttribute(ctx, path.Root(names.AttrTagsAll), tftags.Unknown)...)
		}
	} else {
		response.Diagnostics.Append(response.Plan.SetAttribute(ctx, path.Root(names.AttrTagsAll), tftags.Unknown)...)
	}
}

// WithImportByID is intended to be embedded in resources which import state via the "id" attribute.
// See https://developer.hashicorp.com/terraform/plugin/framework/resources/import.
type WithImportByID struct{}

func (w *WithImportByID) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), request, response)
}

// WithNoUpdate is intended to be embedded in resources which cannot be updated.
type WithNoUpdate struct{}

func (w *WithNoUpdate) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	response.Diagnostics.Append(diag.NewErrorDiagnostic("not supported", "This resource's Update method should not have been called"))
}

// WithNoOpUpdate is intended to be embedded in resources which have no need of a custom Update method.
// For example, resources where only `tags` can be updated and that is handled via transparent tagging.
type WithNoOpUpdate[T any] struct{}

func (w *WithNoOpUpdate[T]) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var t T

	response.Diagnostics.Append(request.Plan.Get(ctx, &t)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &t)...)
}

// WithNoOpUpdate is intended to be embedded in resources which have no need of a custom Delete method.
type WithNoOpDelete struct{}

func (w *WithNoOpDelete) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
}

// DataSourceWithConfigure is a structure to be embedded within a DataSource that implements the DataSourceWithConfigure interface.
type DataSourceWithConfigure struct {
	withMeta
}

// Configure enables provider-level data or clients to be set in the
// provider-defined DataSource type. It is separately executed for each
// ReadDataSource RPC.
func (d *DataSourceWithConfigure) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		d.meta = v
	}
}

func mapHasUnknownElements(m types.Map) bool {
	for _, v := range m.Elements() {
		if v.IsUnknown() {
			return true
		}
	}

	return false
}
