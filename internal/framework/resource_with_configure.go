// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

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

type resourceCRUDer[T any] interface {
	// OnCreate is called when the provider must create a new resource.
	// On entry `data` contains Plan values and on return `data` is written to State.
	OnCreate(ctx context.Context, data *T) diag.Diagnostics
	// OnRead is called when the provider must read resource values in order to update state.
	// On entry `data` contains State values and on return `data` is written to State.
	// Set the boolean return value to `true` if the resource was not found.
	OnRead(ctx context.Context, data *T) (bool, diag.Diagnostics)
	// OnUpdate is called to update the state of the resource.
	// On entry `old` contains State values and `new` contains Plan values.
	// On return `new` is written to State.
	OnUpdate(ctx context.Context, old, new *T) diag.Diagnostics
	// OnDelete is called when the provider must delete the resource.
	// On entry `data` contains State values.
	// Nothing is done with `data` on return.
	OnDelete(ctx context.Context, data *T) diag.Diagnostics
}

type ResourceWithConfigureEx[T any] struct {
	ResourceWithConfigure
	impl resourceCRUDer[T]
}

type r[T any, U any] interface {
	resource.ResourceWithConfigure
	resourceCRUDer[U]
	setImpl(resourceCRUDer[U])
	*T
}

func NewResource[T any, U any, V r[T, U]]() V {
	var v V = new(T)
	v.setImpl(v)
	return v
}

// SetImpl sets the CRUDer implementation.
func (r *ResourceWithConfigureEx[T]) setImpl(impl resourceCRUDer[T]) {
	r.impl = impl
}

// Create is called when the provider must create a new resource.
// Config and planned state values should be read from the CreateRequest and new state values set on the CreateResponse.
func (r *ResourceWithConfigureEx[T]) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data T

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(r.impl.OnCreate(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

// Read is called when the provider must read resource values in order to update state.
// Planned state values should be read from the ReadRequest and new state values set on the ReadResponse.
func (r *ResourceWithConfigureEx[T]) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data T

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	notFound, diags := r.impl.OnRead(ctx, &data)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if notFound {
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

// Update is called to update the state of the resource.
// Config, planned state, and prior state values should be read from the UpdateRequest and new state values set on the UpdateResponse.
func (r *ResourceWithConfigureEx[T]) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new T

	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(r.impl.OnUpdate(ctx, &old, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

// Delete is called when the provider must delete the resource.
// Config values may be read from the DeleteRequest.
func (r *ResourceWithConfigureEx[T]) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data T

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(r.impl.OnDelete(ctx, &data)...)
}

func mapHasUnknownElements(m types.Map) bool {
	for _, v := range m.Elements() {
		if v.IsUnknown() {
			return true
		}
	}

	return false
}
