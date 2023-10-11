// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwprovider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// A data source interceptor is functionality invoked during the data source's CRUD request lifecycle.
// If a Before interceptor returns Diagnostics indicating an error occurred then
// no further interceptors in the chain are run and neither is the schema's method.
// In other cases all interceptors in the chain are run.
type dataSourceInterceptor interface {
	// read is invoke for a Read call.
	read(context.Context, datasource.ReadRequest, *datasource.ReadResponse, *conns.AWSClient, when, diag.Diagnostics) (context.Context, diag.Diagnostics)
}

type dataSourceInterceptors []dataSourceInterceptor

type resourceCRUDRequest interface {
	resource.CreateRequest | resource.ReadRequest | resource.UpdateRequest | resource.DeleteRequest
}
type resourceCRUDResponse interface {
	resource.CreateResponse | resource.ReadResponse | resource.UpdateResponse | resource.DeleteResponse
}

// A resource interceptor is functionality invoked during the resource's CRUD request lifecycle.
// If a Before interceptor returns Diagnostics indicating an error occurred then
// no further interceptors in the chain are run and neither is the schema's method.
// In other cases all interceptors in the chain are run.
type resourceInterceptor interface {
	// create is invoke for a Create call.
	create(context.Context, resource.CreateRequest, *resource.CreateResponse, *conns.AWSClient, when, diag.Diagnostics) (context.Context, diag.Diagnostics)
	// read is invoke for a Read call.
	read(context.Context, resource.ReadRequest, *resource.ReadResponse, *conns.AWSClient, when, diag.Diagnostics) (context.Context, diag.Diagnostics)
	// update is invoke for an Update call.
	update(context.Context, resource.UpdateRequest, *resource.UpdateResponse, *conns.AWSClient, when, diag.Diagnostics) (context.Context, diag.Diagnostics)
	// delete is invoke for a Delete call.
	delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse, *conns.AWSClient, when, diag.Diagnostics) (context.Context, diag.Diagnostics)
}

type resourceInterceptors []resourceInterceptor

type resourceInterceptorFunc[Request resourceCRUDRequest, Response resourceCRUDResponse] func(context.Context, Request, *Response, *conns.AWSClient, when, diag.Diagnostics) (context.Context, diag.Diagnostics)

// create returns a slice of interceptors that run on resource Create.
func (s resourceInterceptors) create() []resourceInterceptorFunc[resource.CreateRequest, resource.CreateResponse] {
	return slices.ApplyToAll(s, func(e resourceInterceptor) resourceInterceptorFunc[resource.CreateRequest, resource.CreateResponse] {
		return e.create
	})
}

// read returns a slice of interceptors that run on resource Read.
func (s resourceInterceptors) read() []resourceInterceptorFunc[resource.ReadRequest, resource.ReadResponse] {
	return slices.ApplyToAll(s, func(e resourceInterceptor) resourceInterceptorFunc[resource.ReadRequest, resource.ReadResponse] {
		return e.read
	})
}

// update returns a slice of interceptors that run on resource Update.
func (s resourceInterceptors) update() []resourceInterceptorFunc[resource.UpdateRequest, resource.UpdateResponse] {
	return slices.ApplyToAll(s, func(e resourceInterceptor) resourceInterceptorFunc[resource.UpdateRequest, resource.UpdateResponse] {
		return e.update
	})
}

// delete returns a slice of interceptors that run on resource Delete.
func (s resourceInterceptors) delete() []resourceInterceptorFunc[resource.DeleteRequest, resource.DeleteResponse] {
	return slices.ApplyToAll(s, func(e resourceInterceptor) resourceInterceptorFunc[resource.DeleteRequest, resource.DeleteResponse] {
		return e.delete
	})
}

// when represents the point in the CRUD request lifecycle that an interceptor is run.
// Multiple values can be ORed together.
type when uint16

const (
	Before  when = 1 << iota // Interceptor is invoked before call to method in schema
	After                    // Interceptor is invoked after successful call to method in schema
	OnError                  // Interceptor is invoked after unsuccessful call to method in schema
	Finally                  // Interceptor is invoked after After or OnError
)

// interceptedHandler returns a handler that invokes the specified CRUD handler, running any interceptors.
func interceptedHandler[Request resourceCRUDRequest, Response resourceCRUDResponse](interceptors []resourceInterceptorFunc[Request, Response], f func(context.Context, Request, *Response) diag.Diagnostics, meta *conns.AWSClient) func(context.Context, Request, *Response) diag.Diagnostics {
	return func(ctx context.Context, request Request, response *Response) diag.Diagnostics {
		var diags diag.Diagnostics
		// Before interceptors are run first to last.
		forward := interceptors

		when := Before
		for _, v := range forward {
			ctx, diags = v(ctx, request, response, meta, when, diags)

			// Short circuit if any Before interceptor errors.
			if diags.HasError() {
				return diags
			}
		}

		// All other interceptors are run last to first.
		reverse := slices.Reverse(forward)
		diags = f(ctx, request, response)

		if diags.HasError() {
			when = OnError
		} else {
			when = After
		}
		for _, v := range reverse {
			ctx, diags = v(ctx, request, response, meta, when, diags)
		}

		when = Finally
		for _, v := range reverse {
			ctx, diags = v(ctx, request, response, meta, when, diags)
		}

		return diags
	}
}

// contextFunc augments Context.
type contextFunc func(context.Context, *conns.AWSClient) context.Context

// wrappedDataSource represents an interceptor dispatcher for a Plugin Framework data source.
type wrappedDataSource struct {
	// bootstrapContext is run on all wrapped methods before any interceptors.
	bootstrapContext contextFunc
	inner            datasource.DataSourceWithConfigure
	interceptors     dataSourceInterceptors
	meta             *conns.AWSClient
}

func newWrappedDataSource(bootstrapContext contextFunc, inner datasource.DataSourceWithConfigure, interceptors dataSourceInterceptors) datasource.DataSourceWithConfigure {
	return &wrappedDataSource{
		bootstrapContext: bootstrapContext,
		inner:            inner,
		interceptors:     interceptors,
	}
}

func (w *wrappedDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	ctx = w.bootstrapContext(ctx, w.meta)
	w.inner.Metadata(ctx, request, response)
}

func (w *wrappedDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	ctx = w.bootstrapContext(ctx, w.meta)
	w.inner.Schema(ctx, request, response)
}

func (w *wrappedDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	ctx = w.bootstrapContext(ctx, w.meta)
	// TODO Run interceptors.
	w.inner.Read(ctx, request, response)
}

func (w *wrappedDataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	ctx = w.bootstrapContext(ctx, w.meta)
	w.inner.Configure(ctx, request, response)
}

// tagsDataSourceInterceptor implements transparent tagging for data sources.
type tagsDataSourceInterceptor struct {
	tags *types.ServicePackageResourceTags
}

func (r tagsDataSourceInterceptor) read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse, meta *conns.AWSClient, when when, diags diag.Diagnostics) (context.Context, diag.Diagnostics) {
	// TODO
	return ctx, diags
}

// wrappedResource represents an interceptor dispatcher for a Plugin Framework resource.
type wrappedResource struct {
	// bootstrapContext is run on all wrapped methods before any interceptors.
	bootstrapContext contextFunc
	inner            resource.ResourceWithConfigure
	interceptors     resourceInterceptors
	meta             *conns.AWSClient
}

func newWrappedResource(bootstrapContext contextFunc, inner resource.ResourceWithConfigure, interceptors resourceInterceptors) resource.ResourceWithConfigure {
	return &wrappedResource{
		bootstrapContext: bootstrapContext,
		inner:            inner,
		interceptors:     interceptors,
	}
}

func (w *wrappedResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	ctx = w.bootstrapContext(ctx, w.meta)
	w.inner.Metadata(ctx, request, response)
}

func (w *wrappedResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	ctx = w.bootstrapContext(ctx, w.meta)
	w.inner.Schema(ctx, request, response)
}

func (w *wrappedResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	f := func(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) diag.Diagnostics {
		w.inner.Create(ctx, request, response)
		return response.Diagnostics
	}
	ctx = w.bootstrapContext(ctx, w.meta)
	diags := interceptedHandler(w.interceptors.create(), f, w.meta)(ctx, request, response)
	response.Diagnostics = diags
}

func (w *wrappedResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	f := func(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) diag.Diagnostics {
		w.inner.Read(ctx, request, response)
		return response.Diagnostics
	}
	ctx = w.bootstrapContext(ctx, w.meta)
	diags := interceptedHandler(w.interceptors.read(), f, w.meta)(ctx, request, response)
	response.Diagnostics = diags
}

func (w *wrappedResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	f := func(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) diag.Diagnostics {
		w.inner.Update(ctx, request, response)
		return response.Diagnostics
	}
	ctx = w.bootstrapContext(ctx, w.meta)
	diags := interceptedHandler(w.interceptors.update(), f, w.meta)(ctx, request, response)
	response.Diagnostics = diags
}

func (w *wrappedResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	f := func(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) diag.Diagnostics {
		w.inner.Delete(ctx, request, response)
		return response.Diagnostics
	}
	ctx = w.bootstrapContext(ctx, w.meta)
	diags := interceptedHandler(w.interceptors.delete(), f, w.meta)(ctx, request, response)
	response.Diagnostics = diags
}

func (w *wrappedResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		w.meta = v
	}
	ctx = w.bootstrapContext(ctx, w.meta)
	w.inner.Configure(ctx, request, response)
}

func (w *wrappedResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	if v, ok := w.inner.(resource.ResourceWithImportState); ok {
		ctx = w.bootstrapContext(ctx, w.meta)
		v.ImportState(ctx, request, response)

		return
	}

	response.Diagnostics.AddError(
		"Resource Import Not Implemented",
		"This resource does not support import. Please contact the provider developer for additional information.",
	)
}

func (w *wrappedResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	if v, ok := w.inner.(resource.ResourceWithModifyPlan); ok {
		ctx = w.bootstrapContext(ctx, w.meta)
		v.ModifyPlan(ctx, request, response)

		return
	}
}

func (w *wrappedResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	if v, ok := w.inner.(resource.ResourceWithConfigValidators); ok {
		ctx = w.bootstrapContext(ctx, w.meta)
		return v.ConfigValidators(ctx)
	}

	return nil
}

func (w *wrappedResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	if v, ok := w.inner.(resource.ResourceWithValidateConfig); ok {
		ctx = w.bootstrapContext(ctx, w.meta)
		v.ValidateConfig(ctx, request, response)
	}
}

func (w *wrappedResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	if v, ok := w.inner.(resource.ResourceWithUpgradeState); ok {
		ctx = w.bootstrapContext(ctx, w.meta)

		return v.UpgradeState(ctx)
	}

	return nil
}

// tagsResourceInterceptor implements transparent tagging for resources.
type tagsResourceInterceptor struct {
	tags *types.ServicePackageResourceTags
}

func (r tagsResourceInterceptor) create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse, meta *conns.AWSClient, when when, diags diag.Diagnostics) (context.Context, diag.Diagnostics) {
	if r.tags == nil {
		return ctx, diags
	}

	inContext, ok := conns.FromContext(ctx)
	if !ok {
		return ctx, diags
	}

	tagsInContext, ok := tftags.FromContext(ctx)
	if !ok {
		return ctx, diags
	}

	switch when {
	case Before:
		var planTags fwtypes.Map
		diags.Append(request.Plan.GetAttribute(ctx, path.Root(names.AttrTags), &planTags)...)

		if diags.HasError() {
			return ctx, diags
		}

		// Merge the resource's configured tags with any provider configured default_tags.
		tags := tagsInContext.DefaultConfig.MergeTags(tftags.New(ctx, planTags))
		// Remove system tags.
		tags = tags.IgnoreSystem(inContext.ServicePackageName)

		tagsInContext.TagsIn = types.Some(tags)
	case After:
		// Set values for unknowns.
		// Remove any provider configured ignore_tags and system tags from those passed to the service API.
		// Computed tags_all include any provider configured default_tags.
		stateTagsAll := flex.FlattenFrameworkStringValueMapLegacy(ctx, tagsInContext.TagsIn.MustUnwrap().IgnoreSystem(inContext.ServicePackageName).IgnoreConfig(tagsInContext.IgnoreConfig).Map())
		diags.Append(response.State.SetAttribute(ctx, path.Root(names.AttrTagsAll), &stateTagsAll)...)

		if diags.HasError() {
			return ctx, diags
		}
	}

	return ctx, diags
}

func (r tagsResourceInterceptor) read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse, meta *conns.AWSClient, when when, diags diag.Diagnostics) (context.Context, diag.Diagnostics) {
	if r.tags == nil {
		return ctx, diags
	}

	inContext, ok := conns.FromContext(ctx)
	if !ok {
		return ctx, diags
	}

	sp, ok := meta.ServicePackages[inContext.ServicePackageName]
	if !ok {
		return ctx, diags
	}

	serviceName, err := names.HumanFriendly(inContext.ServicePackageName)
	if err != nil {
		serviceName = "<service>"
	}

	resourceName := inContext.ResourceName
	if resourceName == "" {
		resourceName = "<thing>"
	}

	tagsInContext, ok := tftags.FromContext(ctx)
	if !ok {
		return ctx, diags
	}

	switch when {
	case After:
		// Will occur on a refresh when the resource does not exist in AWS and needs to be recreated, e.g. "_disappears" tests.
		if response.State.Raw.IsNull() {
			return ctx, diags
		}

		// If the R handler didn't set tags, try and read them from the service API.
		if tagsInContext.TagsOut.IsNone() {
			if identifierAttribute := r.tags.IdentifierAttribute; identifierAttribute != "" {
				var identifier string

				diags.Append(response.State.GetAttribute(ctx, path.Root(identifierAttribute), &identifier)...)

				if diags.HasError() {
					return ctx, diags
				}

				// Some old resources may not have the required attribute set after Read:
				// https://github.com/hashicorp/terraform-provider-aws/issues/31180
				if identifier != "" {
					// If the service package has a generic resource list tags methods, call it.
					var err error

					if v, ok := sp.(interface {
						ListTags(context.Context, any, string) error
					}); ok {
						err = v.ListTags(ctx, meta, identifier) // Sets tags in Context
					} else if v, ok := sp.(interface {
						ListTags(context.Context, any, string, string) error
					}); ok && r.tags.ResourceType != "" {
						err = v.ListTags(ctx, meta, identifier, r.tags.ResourceType) // Sets tags in Context
					}

					// ISO partitions may not support tagging, giving error.
					if errs.IsUnsupportedOperationInPartitionError(meta.Partition, err) {
						return ctx, diags
					}

					if err != nil {
						diags.AddError(fmt.Sprintf("listing tags for %s %s (%s)", serviceName, resourceName, identifier), err.Error())

						return ctx, diags
					}
				}
			}
		}

		apiTags := tagsInContext.TagsOut.UnwrapOrDefault()

		// AWS APIs often return empty lists of tags when none have been configured.
		stateTags := tftags.Null
		// Remove any provider configured ignore_tags and system tags from those returned from the service API.
		// The resource's configured tags do not include any provider configured default_tags.
		if v := apiTags.IgnoreSystem(inContext.ServicePackageName).IgnoreConfig(tagsInContext.IgnoreConfig).ResolveDuplicatesFramework(ctx, tagsInContext.DefaultConfig, tagsInContext.IgnoreConfig, response, diags).Map(); len(v) > 0 {
			stateTags = flex.FlattenFrameworkStringValueMapLegacy(ctx, v)
		}
		diags.Append(response.State.SetAttribute(ctx, path.Root(names.AttrTags), &stateTags)...)

		if diags.HasError() {
			return ctx, diags
		}

		// Computed tags_all do.
		stateTagsAll := flex.FlattenFrameworkStringValueMapLegacy(ctx, apiTags.IgnoreSystem(inContext.ServicePackageName).IgnoreConfig(tagsInContext.IgnoreConfig).Map())
		diags.Append(response.State.SetAttribute(ctx, path.Root(names.AttrTagsAll), &stateTagsAll)...)

		if diags.HasError() {
			return ctx, diags
		}
	}

	return ctx, diags
}

func (r tagsResourceInterceptor) update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse, meta *conns.AWSClient, when when, diags diag.Diagnostics) (context.Context, diag.Diagnostics) {
	if r.tags == nil {
		return ctx, diags
	}

	inContext, ok := conns.FromContext(ctx)
	if !ok {
		return ctx, diags
	}

	sp, ok := meta.ServicePackages[inContext.ServicePackageName]
	if !ok {
		return ctx, diags
	}

	serviceName, err := names.HumanFriendly(inContext.ServicePackageName)
	if err != nil {
		serviceName = "<service>"
	}

	resourceName := inContext.ResourceName
	if resourceName == "" {
		resourceName = "<thing>"
	}

	tagsInContext, ok := tftags.FromContext(ctx)
	if !ok {
		return ctx, diags
	}

	switch when {
	case Before:
		var planTags fwtypes.Map
		diags.Append(request.Plan.GetAttribute(ctx, path.Root(names.AttrTags), &planTags)...)

		if diags.HasError() {
			return ctx, diags
		}

		// Merge the resource's configured tags with any provider configured default_tags.
		tags := tagsInContext.DefaultConfig.MergeTags(tftags.New(ctx, planTags))
		// Remove system tags.
		tags = tags.IgnoreSystem(inContext.ServicePackageName)

		tagsInContext.TagsIn = types.Some(tags)

		var oldTagsAll, newTagsAll fwtypes.Map

		diags.Append(request.State.GetAttribute(ctx, path.Root(names.AttrTagsAll), &oldTagsAll)...)

		if diags.HasError() {
			return ctx, diags
		}

		diags.Append(request.Plan.GetAttribute(ctx, path.Root(names.AttrTagsAll), &newTagsAll)...)

		if diags.HasError() {
			return ctx, diags
		}

		if !newTagsAll.Equal(oldTagsAll) {
			if identifierAttribute := r.tags.IdentifierAttribute; identifierAttribute != "" {
				var identifier string

				diags.Append(request.Plan.GetAttribute(ctx, path.Root(identifierAttribute), &identifier)...)

				if diags.HasError() {
					return ctx, diags
				}

				// Some old resources may not have the required attribute set after Read:
				// https://github.com/hashicorp/terraform-provider-aws/issues/31180
				if identifier != "" {
					// If the service package has a generic resource update tags methods, call it.
					var err error

					if v, ok := sp.(interface {
						UpdateTags(context.Context, any, string, any, any) error
					}); ok {
						err = v.UpdateTags(ctx, meta, identifier, oldTagsAll, newTagsAll)
					} else if v, ok := sp.(interface {
						UpdateTags(context.Context, any, string, string, any, any) error
					}); ok && r.tags.ResourceType != "" {
						err = v.UpdateTags(ctx, meta, identifier, r.tags.ResourceType, oldTagsAll, newTagsAll)
					}

					// ISO partitions may not support tagging, giving error.
					if errs.IsUnsupportedOperationInPartitionError(meta.Partition, err) {
						return ctx, diags
					}

					if err != nil {
						diags.AddError(fmt.Sprintf("updating tags for %s %s (%s)", serviceName, resourceName, identifier), err.Error())

						return ctx, diags
					}
				}
			}
			// TODO If the only change was to tags it would be nice to not call the resource's U handler.
		}
	}

	return ctx, diags
}

func (r tagsResourceInterceptor) delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse, meta *conns.AWSClient, when when, diags diag.Diagnostics) (context.Context, diag.Diagnostics) {
	return ctx, diags
}
