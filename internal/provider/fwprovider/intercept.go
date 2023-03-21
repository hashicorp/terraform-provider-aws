package fwprovider

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

type contextFunc func(context.Context, *conns.AWSClient) context.Context

// wrappedDataSource wraps a data source, adding common functionality.
type wrappedDataSource struct {
	bootstrapContext contextFunc
	inner            datasource.DataSourceWithConfigure
	meta             *conns.AWSClient
	typeName         string
}

func newWrappedDataSource(bootstrapContext contextFunc, inner datasource.DataSourceWithConfigure) datasource.DataSourceWithConfigure {
	return &wrappedDataSource{
		bootstrapContext: bootstrapContext,
		inner:            inner,
		typeName:         strings.TrimPrefix(reflect.TypeOf(inner).String(), "*"),
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

	tflog.Debug(ctx, fmt.Sprintf("%s.Read enter", w.typeName))

	w.inner.Read(ctx, request, response)

	tflog.Debug(ctx, fmt.Sprintf("%s.Read exit", w.typeName))
}

func (w *wrappedDataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	ctx = w.bootstrapContext(ctx, w.meta)
	w.inner.Configure(ctx, request, response)
}

// wrappedResource wraps a resource, adding common functionality.
type wrappedResource struct {
	bootstrapContext contextFunc
	inner            resource.ResourceWithConfigure
	meta             *conns.AWSClient
	typeName         string
}

func newWrappedResource(bootstrapContext contextFunc, inner resource.ResourceWithConfigure) resource.ResourceWithConfigure {
	return &wrappedResource{
		bootstrapContext: bootstrapContext,
		inner:            inner,
		typeName:         strings.TrimPrefix(reflect.TypeOf(inner).String(), "*"),
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
	ctx = w.bootstrapContext(ctx, w.meta)

	tflog.Debug(ctx, fmt.Sprintf("%s.Create enter", w.typeName))

	w.inner.Create(ctx, request, response)

	tflog.Debug(ctx, fmt.Sprintf("%s.Create exit", w.typeName))
}

func (w *wrappedResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	ctx = w.bootstrapContext(ctx, w.meta)

	tflog.Debug(ctx, fmt.Sprintf("%s.Read enter", w.typeName))

	w.inner.Read(ctx, request, response)

	tflog.Debug(ctx, fmt.Sprintf("%s.Read exit", w.typeName))
}

func (w *wrappedResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	ctx = w.bootstrapContext(ctx, w.meta)

	tflog.Debug(ctx, fmt.Sprintf("%s.Update enter", w.typeName))

	w.inner.Update(ctx, request, response)

	tflog.Debug(ctx, fmt.Sprintf("%s.Update exit", w.typeName))
}

func (w *wrappedResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	ctx = w.bootstrapContext(ctx, w.meta)

	tflog.Debug(ctx, fmt.Sprintf("%s.Delete enter", w.typeName))

	w.inner.Delete(ctx, request, response)

	tflog.Debug(ctx, fmt.Sprintf("%s.Delete exit", w.typeName))
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
