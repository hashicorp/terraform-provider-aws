// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/aws-sdk-go-base/v2/useragent"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/identity"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/importer"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/listresource"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	tfunique "github.com/hashicorp/terraform-provider-aws/internal/unique"
	"github.com/hashicorp/terraform-provider-aws/internal/vcr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Implemented by (Config|Plan|State).GetAttribute().
type getAttributeFunc func(context.Context, path.Path, any) diag.Diagnostics

// wrappedDataSource represents an interceptor dispatcher for a Plugin Framework data source.
type wrappedDataSource struct {
	inner              datasource.DataSourceWithConfigure
	meta               *conns.AWSClient
	servicePackageName string
	spec               *inttypes.ServicePackageFrameworkDataSource
	interceptors       interceptorInvocations
}

func newWrappedDataSource(spec *inttypes.ServicePackageFrameworkDataSource, servicePackageName string) datasource.DataSourceWithConfigure {
	var isRegionOverrideEnabled bool
	if regionSpec := spec.Region; !tfunique.IsHandleNil(regionSpec) && regionSpec.Value().IsOverrideEnabled {
		isRegionOverrideEnabled = true
	}

	var interceptors interceptorInvocations

	if isRegionOverrideEnabled {
		v := spec.Region.Value()

		interceptors = append(interceptors, dataSourceInjectRegionAttribute(v.IsOverrideDeprecated))
		if v.IsValidateOverrideInPartition {
			interceptors = append(interceptors, dataSourceValidateRegion())
		}
		interceptors = append(interceptors, dataSourceSetRegionInState())
	}

	if !tfunique.IsHandleNil(spec.Tags) {
		interceptors = append(interceptors, dataSourceTransparentTagging(spec.Tags))
	}

	inner, _ := spec.Factory(context.TODO())

	return &wrappedDataSource{
		inner:              inner,
		servicePackageName: servicePackageName,
		spec:               spec,
		interceptors:       interceptors,
	}
}

// context is run on all wrapped methods before any interceptors.
func (w *wrappedDataSource) context(ctx context.Context, getAttribute getAttributeFunc, providerMeta *tfsdk.Config, c *conns.AWSClient) (context.Context, diag.Diagnostics) {
	var diags diag.Diagnostics
	var overrideRegion string

	var isRegionOverrideEnabled bool
	if regionSpec := w.spec.Region; !tfunique.IsHandleNil(regionSpec) && regionSpec.Value().IsOverrideEnabled {
		isRegionOverrideEnabled = true
	}

	if isRegionOverrideEnabled && getAttribute != nil {
		var target types.String
		diags.Append(getAttribute(ctx, path.Root(names.AttrRegion), &target)...)
		if diags.HasError() {
			return ctx, diags
		}

		overrideRegion = target.ValueString()
	}

	ctx = conns.NewResourceContext(ctx, w.servicePackageName, w.spec.Name, w.spec.TypeName, overrideRegion)
	if c != nil {
		ctx = tftags.NewContext(ctx, c.DefaultTagsConfig(ctx), c.IgnoreTagsConfig(ctx), c.TagPolicyConfig(ctx))
		ctx = c.RegisterLogger(ctx)
		if s := c.RandomnessSource(); s != nil {
			ctx = vcr.NewContext(ctx, s)
		}
		ctx = fwflex.RegisterLogger(ctx)
	}

	if providerMeta != nil {
		var metadata []string
		diags.Append(providerMeta.GetAttribute(ctx, path.Root("user_agent"), &metadata)...)
		if diags.HasError() {
			return ctx, diags
		}

		if metadata != nil {
			ctx = useragent.Context(ctx, useragent.FromSlice(metadata))
		}
	}

	return ctx, diags
}

func (w *wrappedDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	// This method does not call down to the inner data source.
	response.TypeName = w.spec.TypeName
}

func (w *wrappedDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	ctx, diags := w.context(ctx, nil, nil, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	interceptedHandler(w.interceptors.dataSourceSchema(), w.inner.Schema, dataSourceSchemaHasError, w.meta)(ctx, request, response)
	if response.Diagnostics.HasError() {
		return
	}

	// Validate the data source's model against the schema.
	if v, ok := w.inner.(framework.DataSourceValidateModel); ok {
		response.Diagnostics.Append(v.ValidateModel(ctx, &response.Schema)...)
		if response.Diagnostics.HasError() {
			response.Diagnostics.AddError("data source model validation error", w.spec.TypeName)
			return
		}
	} else {
		response.Diagnostics.AddError("missing framework.DataSourceValidateModel", w.spec.TypeName)
	}
}

func (w *wrappedDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	ctx, diags := w.context(ctx, request.Config.GetAttribute, &request.ProviderMeta, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	interceptedHandler(w.interceptors.dataSourceRead(), w.inner.Read, dataSourceReadHasError, w.meta)(ctx, request, response)
}

func (w *wrappedDataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		w.meta = v
	}

	ctx, diags := w.context(ctx, nil, nil, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	w.inner.Configure(ctx, request, response)
}

func (w *wrappedDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	if v, ok := w.inner.(datasource.DataSourceWithConfigValidators); ok {
		ctx, diags := w.context(ctx, nil, nil, w.meta)
		if diags.HasError() {
			tflog.Warn(ctx, "wrapping ConfigValidators", map[string]any{
				"data source":            w.spec.TypeName,
				"bootstrapContext error": fwdiag.DiagnosticsString(diags),
			})

			return nil
		}

		return v.ConfigValidators(ctx)
	}

	return nil
}

func (w *wrappedDataSource) ValidateConfig(ctx context.Context, request datasource.ValidateConfigRequest, response *datasource.ValidateConfigResponse) {
	if v, ok := w.inner.(datasource.DataSourceWithValidateConfig); ok {
		ctx, diags := w.context(ctx, request.Config.GetAttribute, nil, w.meta)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		v.ValidateConfig(ctx, request, response)
	}
}

// wrappedEphemeralResource represents an interceptor dispatcher for a Plugin Framework ephemeral resource.
type wrappedEphemeralResource struct {
	inner              ephemeral.EphemeralResourceWithConfigure
	meta               *conns.AWSClient
	servicePackageName string
	spec               *inttypes.ServicePackageEphemeralResource
	interceptors       interceptorInvocations
}

func newWrappedEphemeralResource(spec *inttypes.ServicePackageEphemeralResource, servicePackageName string) ephemeral.EphemeralResourceWithConfigure {
	var isRegionOverrideEnabled bool
	if regionSpec := spec.Region; !tfunique.IsHandleNil(regionSpec) && regionSpec.Value().IsOverrideEnabled {
		isRegionOverrideEnabled = true
	}

	var interceptors interceptorInvocations

	if isRegionOverrideEnabled {
		v := spec.Region.Value()

		interceptors = append(interceptors, ephemeralResourceInjectRegionAttribute())
		if v.IsValidateOverrideInPartition {
			interceptors = append(interceptors, ephemeralResourceValidateRegion())
		}
		interceptors = append(interceptors, ephemeralResourceSetRegionInResult())
	}

	inner, _ := spec.Factory(context.TODO())

	return &wrappedEphemeralResource{
		inner:              inner,
		servicePackageName: servicePackageName,
		spec:               spec,
		interceptors:       interceptors,
	}
}

// context is run on all wrapped methods before any interceptors.
func (w *wrappedEphemeralResource) context(ctx context.Context, getAttribute getAttributeFunc, c *conns.AWSClient) (context.Context, diag.Diagnostics) {
	var diags diag.Diagnostics
	var overrideRegion string

	var isRegionOverrideEnabled bool
	if regionSpec := w.spec.Region; !tfunique.IsHandleNil(regionSpec) && regionSpec.Value().IsOverrideEnabled {
		isRegionOverrideEnabled = true
	}

	if isRegionOverrideEnabled && getAttribute != nil {
		var target types.String
		diags.Append(getAttribute(ctx, path.Root(names.AttrRegion), &target)...)
		if diags.HasError() {
			return ctx, diags
		}

		overrideRegion = target.ValueString()
	}

	ctx = conns.NewResourceContext(ctx, w.servicePackageName, w.spec.Name, w.spec.TypeName, overrideRegion)
	if c != nil {
		ctx = c.RegisterLogger(ctx)
		ctx = fwflex.RegisterLogger(ctx)
		ctx = logging.MaskSensitiveValuesByKey(ctx, logging.HTTPKeyRequestBody, logging.HTTPKeyResponseBody)
	}

	return ctx, diags
}

func (w *wrappedEphemeralResource) Metadata(ctx context.Context, request ephemeral.MetadataRequest, response *ephemeral.MetadataResponse) {
	// This method does not call down to the inner ephemeral resource.
	response.TypeName = w.spec.TypeName
}

func (w *wrappedEphemeralResource) Schema(ctx context.Context, request ephemeral.SchemaRequest, response *ephemeral.SchemaResponse) {
	ctx, diags := w.context(ctx, nil, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	interceptedHandler(w.interceptors.ephemeralResourceSchema(), w.inner.Schema, ephemeralSchemaHasError, w.meta)(ctx, request, response)

	// Validate the ephemeral resource's model against the schema.
	if v, ok := w.inner.(framework.EphemeralResourceValidateModel); ok {
		response.Diagnostics.Append(v.ValidateModel(ctx, &response.Schema)...)
		if response.Diagnostics.HasError() {
			response.Diagnostics.AddError("ephemeral resource model validation error", w.spec.TypeName)
			return
		}
	} else {
		response.Diagnostics.AddError("missing framework.EphemeralResourceValidateModel", w.spec.TypeName)
	}
}

func (w *wrappedEphemeralResource) Open(ctx context.Context, request ephemeral.OpenRequest, response *ephemeral.OpenResponse) {
	ctx, diags := w.context(ctx, request.Config.GetAttribute, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	interceptedHandler(w.interceptors.ephemeralResourceOpen(), w.inner.Open, ephemeralOpenHasError, w.meta)(ctx, request, response)
}

func (w *wrappedEphemeralResource) Configure(ctx context.Context, request ephemeral.ConfigureRequest, response *ephemeral.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		w.meta = v
	}

	ctx, diags := w.context(ctx, nil, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	w.inner.Configure(ctx, request, response)
}

func (w *wrappedEphemeralResource) Renew(ctx context.Context, request ephemeral.RenewRequest, response *ephemeral.RenewResponse) {
	if v, ok := w.inner.(ephemeral.EphemeralResourceWithRenew); ok {
		ctx, diags := w.context(ctx, nil, w.meta)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		interceptedHandler(w.interceptors.ephemeralResourceRenew(), v.Renew, ephemeralRenewHasError, w.meta)(ctx, request, response)
	}
}

func (w *wrappedEphemeralResource) Close(ctx context.Context, request ephemeral.CloseRequest, response *ephemeral.CloseResponse) {
	if v, ok := w.inner.(ephemeral.EphemeralResourceWithClose); ok {
		ctx, diags := w.context(ctx, nil, w.meta)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		interceptedHandler(w.interceptors.ephemeralResourceClose(), v.Close, ephemeralCloseHasError, w.meta)(ctx, request, response)
	}
}

func (w *wrappedEphemeralResource) ConfigValidators(ctx context.Context) []ephemeral.ConfigValidator {
	if v, ok := w.inner.(ephemeral.EphemeralResourceWithConfigValidators); ok {
		ctx, diags := w.context(ctx, nil, w.meta)
		if diags.HasError() {
			tflog.Warn(ctx, "wrapping ConfigValidators", map[string]any{
				"ephemeral resource":     w.spec.TypeName,
				"bootstrapContext error": fwdiag.DiagnosticsString(diags),
			})

			return nil
		}

		return v.ConfigValidators(ctx)
	}

	return nil
}

func (w *wrappedEphemeralResource) ValidateConfig(ctx context.Context, request ephemeral.ValidateConfigRequest, response *ephemeral.ValidateConfigResponse) {
	if v, ok := w.inner.(ephemeral.EphemeralResourceWithValidateConfig); ok {
		ctx, diags := w.context(ctx, request.Config.GetAttribute, w.meta)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		v.ValidateConfig(ctx, request, response)
	}
}

// wrappedAction represents an interceptor dispatcher for a Plugin Framework action.
type wrappedAction struct {
	inner              action.ActionWithConfigure
	meta               *conns.AWSClient
	servicePackageName string
	spec               *inttypes.ServicePackageAction
	interceptors       interceptorInvocations
}

func newWrappedAction(spec *inttypes.ServicePackageAction, servicePackageName string) action.ActionWithConfigure {
	var isRegionOverrideEnabled bool
	if regionSpec := spec.Region; !tfunique.IsHandleNil(regionSpec) && regionSpec.Value().IsOverrideEnabled {
		isRegionOverrideEnabled = true
	}

	var interceptors interceptorInvocations

	if isRegionOverrideEnabled {
		v := spec.Region.Value()

		interceptors = append(interceptors, actionInjectRegionAttribute())
		if v.IsValidateOverrideInPartition {
			interceptors = append(interceptors, actionValidateRegion())
		}
	}

	inner, _ := spec.Factory(context.TODO())

	return &wrappedAction{
		inner:              inner,
		servicePackageName: servicePackageName,
		spec:               spec,
		interceptors:       interceptors,
	}
}

// context is run on all wrapped methods before any interceptors.
func (w *wrappedAction) context(ctx context.Context, getAttribute getAttributeFunc, c *conns.AWSClient) (context.Context, diag.Diagnostics) {
	var diags diag.Diagnostics
	var overrideRegion string

	var isRegionOverrideEnabled bool
	if regionSpec := w.spec.Region; !tfunique.IsHandleNil(regionSpec) && regionSpec.Value().IsOverrideEnabled {
		isRegionOverrideEnabled = true
	}

	if isRegionOverrideEnabled && getAttribute != nil {
		var target types.String
		diags.Append(getAttribute(ctx, path.Root(names.AttrRegion), &target)...)
		if diags.HasError() {
			return ctx, diags
		}

		overrideRegion = target.ValueString()
	}

	ctx = conns.NewResourceContext(ctx, w.servicePackageName, w.spec.Name, w.spec.TypeName, overrideRegion)
	if c != nil {
		ctx = c.RegisterLogger(ctx)
		ctx = fwflex.RegisterLogger(ctx)
		ctx = logging.MaskSensitiveValuesByKey(ctx, logging.HTTPKeyRequestBody, logging.HTTPKeyResponseBody)
	}

	return ctx, diags
}

func (w *wrappedAction) Metadata(ctx context.Context, request action.MetadataRequest, response *action.MetadataResponse) {
	// This method does not call down to the inner action.
	response.TypeName = w.spec.TypeName
}

func (w *wrappedAction) Schema(ctx context.Context, request action.SchemaRequest, response *action.SchemaResponse) {
	ctx, diags := w.context(ctx, nil, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	f := func(ctx context.Context, request action.SchemaRequest, response *action.SchemaResponse) {
		w.inner.Schema(ctx, request, response)
	}
	interceptedHandler(w.interceptors.actionSchema(), f, actionSchemaHasError, w.meta)(ctx, request, response)

	// Validate the action's model against the schema.
	if v, ok := w.inner.(framework.ActionValidateModel); ok {
		response.Diagnostics.Append(v.ValidateModel(ctx, &response.Schema)...)
		if response.Diagnostics.HasError() {
			response.Diagnostics.AddError("action model validation error", w.spec.TypeName)
			return
		}
	} else {
		response.Diagnostics.AddError("missing framework.ActionValidateModel", w.spec.TypeName)
	}
}

func (w *wrappedAction) Invoke(ctx context.Context, request action.InvokeRequest, response *action.InvokeResponse) {
	ctx, diags := w.context(ctx, request.Config.GetAttribute, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	f := func(ctx context.Context, request action.InvokeRequest, response *action.InvokeResponse) {
		w.inner.Invoke(ctx, request, response)
	}
	interceptedHandler(w.interceptors.actionInvoke(), f, actionInvokeHasError, w.meta)(ctx, request, response)
}

func (w *wrappedAction) Configure(ctx context.Context, request action.ConfigureRequest, response *action.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		w.meta = v
	}

	ctx, diags := w.context(ctx, nil, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	w.inner.Configure(ctx, request, response)
}

func (w *wrappedAction) ConfigValidators(ctx context.Context) []action.ConfigValidator {
	if v, ok := w.inner.(action.ActionWithConfigValidators); ok {
		ctx, diags := w.context(ctx, nil, w.meta)
		if diags.HasError() {
			tflog.Warn(ctx, "wrapping ConfigValidators", map[string]any{
				"action":                 w.spec.TypeName,
				"bootstrapContext error": fwdiag.DiagnosticsString(diags),
			})

			return nil
		}

		return v.ConfigValidators(ctx)
	}

	return nil
}

func (w *wrappedAction) ValidateConfig(ctx context.Context, request action.ValidateConfigRequest, response *action.ValidateConfigResponse) {
	if v, ok := w.inner.(action.ActionWithValidateConfig); ok {
		ctx, diags := w.context(ctx, request.Config.GetAttribute, w.meta)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		v.ValidateConfig(ctx, request, response)
	}
}

// wrappedResource represents an interceptor dispatcher for a Plugin Framework resource.
type wrappedResource struct {
	inner              resource.ResourceWithConfigure
	meta               *conns.AWSClient
	servicePackageName string
	spec               *inttypes.ServicePackageFrameworkResource
	interceptors       interceptorInvocations
}

func newWrappedResource(spec *inttypes.ServicePackageFrameworkResource, servicePackageName string) resource.ResourceWithConfigure {
	var isRegionOverrideEnabled bool
	if v := spec.Region; !tfunique.IsHandleNil(v) && v.Value().IsOverrideEnabled {
		isRegionOverrideEnabled = true
	}

	var interceptors interceptorInvocations

	if isRegionOverrideEnabled {
		v := spec.Region.Value()

		interceptors = append(interceptors, resourceInjectRegionAttribute(v.IsOverrideDeprecated))
		if v.IsValidateOverrideInPartition {
			interceptors = append(interceptors, resourceValidateRegion())
		}
		interceptors = append(interceptors, resourceDefaultRegion())
		interceptors = append(interceptors, resourceForceNewIfRegionChanges())
		interceptors = append(interceptors, resourceSetRegionInState())
		if spec.Identity.HasInherentRegion() {
			interceptors = append(interceptors, resourceImportRegionNoDefault())
		} else {
			interceptors = append(interceptors, resourceImportRegion())
		}
	}

	if !tfunique.IsHandleNil(spec.Tags) {
		interceptors = append(interceptors, resourceTransparentTagging(spec.Tags))
		interceptors = append(interceptors, resourceValidateRequiredTags())
	}

	inner, _ := spec.Factory(context.TODO())

	if len(spec.Identity.Attributes) == 0 {
		return &wrappedResource{
			inner:              inner,
			servicePackageName: servicePackageName,
			spec:               spec,
			interceptors:       interceptors,
		}
	}

	interceptors = append(interceptors, newIdentityInterceptor(spec.Identity.Attributes))
	if v, ok := inner.(framework.Identityer); ok {
		v.SetIdentitySpec(spec.Identity)
	}

	if spec.Import.WrappedImport {
		if v, ok := inner.(framework.ImportByIdentityer); ok {
			v.SetImportSpec(spec.Import)
		}
		// If the resource does not implement framework.ImportByIdentityer,
		// it will be caught by `validateResourceSchemas`, so we can ignore it here.
	}

	return &wrappedResourceWithIdentity{
		wrappedResource: wrappedResource{
			inner:              inner,
			servicePackageName: servicePackageName,
			spec:               spec,
			interceptors:       interceptors,
		},
	}
}

// context is run on all wrapped methods before any interceptors.
func (w *wrappedResource) context(ctx context.Context, getAttribute getAttributeFunc, providerMeta *tfsdk.Config, c *conns.AWSClient) (context.Context, diag.Diagnostics) {
	var diags diag.Diagnostics
	var overrideRegion string

	var isRegionOverrideEnabled bool
	if regionSpec := w.spec.Region; !tfunique.IsHandleNil(regionSpec) && regionSpec.Value().IsOverrideEnabled {
		isRegionOverrideEnabled = true
	}

	if isRegionOverrideEnabled && getAttribute != nil {
		var target types.String
		diags.Append(getAttribute(ctx, path.Root(names.AttrRegion), &target)...)
		if diags.HasError() {
			return ctx, diags
		}

		overrideRegion = target.ValueString()
	}

	ctx = conns.NewResourceContext(ctx, w.servicePackageName, w.spec.Name, w.spec.TypeName, overrideRegion)
	if c != nil {
		ctx = tftags.NewContext(ctx, c.DefaultTagsConfig(ctx), c.IgnoreTagsConfig(ctx), c.TagPolicyConfig(ctx))
		ctx = c.RegisterLogger(ctx)
		ctx = fwflex.RegisterLogger(ctx)
	}

	if providerMeta != nil {
		var metadata []string
		diags.Append(providerMeta.GetAttribute(ctx, path.Root("user_agent"), &metadata)...)
		if diags.HasError() {
			return ctx, diags
		}

		if metadata != nil {
			ctx = useragent.Context(ctx, useragent.FromSlice(metadata))
		}
	}

	return ctx, diags
}

func (w *wrappedResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	// This method does not call down to the inner resource.
	response.TypeName = w.spec.TypeName

	if w.spec.Identity.IsMutable {
		response.ResourceBehavior.MutableIdentity = true
	}
}

func (w *wrappedResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	ctx, diags := w.context(ctx, nil, nil, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	interceptedHandler(w.interceptors.resourceSchema(), w.inner.Schema, resourceSchemaHasError, w.meta)(ctx, request, response)

	// Validate the resource's model against the schema.
	if v, ok := w.inner.(framework.ResourceValidateModel); ok {
		response.Diagnostics.Append(v.ValidateModel(ctx, &response.Schema)...)
		if response.Diagnostics.HasError() {
			response.Diagnostics.AddError("resource model validation error", w.spec.TypeName)
			return
		}
	} else if w.spec.TypeName != "aws_lexv2models_bot_version" { // Hacky yukkery caused by attribute of type map[string]Object.
		response.Diagnostics.AddError("missing framework.ResourceValidateModel", w.spec.TypeName)
	}
}

func (w *wrappedResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	ctx, diags := w.context(ctx, request.Plan.GetAttribute, &request.ProviderMeta, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	interceptedHandler(w.interceptors.resourceCreate(), w.inner.Create, resourceCreateHasError, w.meta)(ctx, request, response)
}

func (w *wrappedResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	ctx, diags := w.context(ctx, request.State.GetAttribute, &request.ProviderMeta, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	interceptedHandler(w.interceptors.resourceRead(), w.inner.Read, resourceReadHasError, w.meta)(ctx, request, response)
}

func (w *wrappedResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	ctx, diags := w.context(ctx, request.Plan.GetAttribute, &request.ProviderMeta, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	interceptedHandler(w.interceptors.resourceUpdate(), w.inner.Update, resourceUpdateHasError, w.meta)(ctx, request, response)
}

func (w *wrappedResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	ctx, diags := w.context(ctx, request.State.GetAttribute, &request.ProviderMeta, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	interceptedHandler(w.interceptors.resourceDelete(), w.inner.Delete, resourceDeleteHasError, w.meta)(ctx, request, response)
}

func (w *wrappedResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		w.meta = v
	}

	ctx, diags := w.context(ctx, nil, nil, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	w.inner.Configure(ctx, request, response)
}

func (w *wrappedResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	if v, ok := w.inner.(resource.ResourceWithImportState); ok {
		ctx, diags := w.context(ctx, nil, nil, w.meta)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		ctx = importer.Context(ctx, w.meta)
		interceptedHandler(w.interceptors.resourceImportState(), v.ImportState, resourceImportStateHasError, w.meta)(ctx, request, response)

		return
	}

	response.Diagnostics.AddError(
		"Resource Import Not Implemented",
		"This resource does not support import. Please contact the provider developer for additional information.",
	)
}

func (w *wrappedResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	ctx, diags := w.context(ctx, request.Config.GetAttribute, &request.ProviderMeta, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// We run ModifyPlan interceptors even if the resource has not defined a ModifyPlan method.
	f := func(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	}
	if v, ok := w.inner.(resource.ResourceWithModifyPlan); ok {
		f = v.ModifyPlan
	}
	interceptedHandler(w.interceptors.resourceModifyPlan(), f, resourceModifyPlanHasError, w.meta)(ctx, request, response)
}

func (w *wrappedResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	if v, ok := w.inner.(resource.ResourceWithConfigValidators); ok {
		ctx, diags := w.context(ctx, nil, nil, w.meta)
		if diags.HasError() {
			tflog.Warn(ctx, "wrapping ConfigValidators", map[string]any{
				"resource":               w.spec.TypeName,
				"bootstrapContext error": fwdiag.DiagnosticsString(diags),
			})

			return nil
		}

		return v.ConfigValidators(ctx)
	}

	return nil
}

func (w *wrappedResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	ctx, diags := w.context(ctx, request.Config.GetAttribute, nil, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if v, ok := w.inner.(resource.ResourceWithValidateConfig); ok {
		v.ValidateConfig(ctx, request, response)
	}
}

func (w *wrappedResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	if v, ok := w.inner.(resource.ResourceWithUpgradeState); ok {
		ctx, diags := w.context(ctx, nil, nil, w.meta)
		if diags.HasError() {
			tflog.Warn(ctx, "wrapping UpgradeState", map[string]any{
				"resource":               w.spec.TypeName,
				"bootstrapContext error": fwdiag.DiagnosticsString(diags),
			})

			return nil
		}

		return v.UpgradeState(ctx)
	}

	return nil
}

func (w *wrappedResource) MoveState(ctx context.Context) []resource.StateMover {
	if v, ok := w.inner.(resource.ResourceWithMoveState); ok {
		ctx, diags := w.context(ctx, nil, nil, w.meta)
		if diags.HasError() {
			tflog.Warn(ctx, "wrapping MoveState", map[string]any{
				"resource":               w.spec.TypeName,
				"bootstrapContext error": fwdiag.DiagnosticsString(diags),
			})

			return nil
		}

		return v.MoveState(ctx)
	}

	return nil
}

type wrappedResourceWithIdentity struct {
	wrappedResource
}

func (w *wrappedResourceWithIdentity) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	if len(w.spec.Identity.Attributes) > 0 {
		resp.IdentitySchema = identity.NewIdentitySchema(w.spec.Identity)
	}
}

type wrappedListResourceFramework struct {
	inner              list.ListResourceWithConfigure
	meta               *conns.AWSClient
	servicePackageName string
	spec               *inttypes.ServicePackageFrameworkListResource
	interceptors       interceptorInvocations
}

var _ list.ListResourceWithConfigure = &wrappedListResourceFramework{}

func newWrappedListResourceFramework(spec *inttypes.ServicePackageFrameworkListResource, servicePackageName string) list.ListResourceWithConfigure {
	var interceptors interceptorInvocations

	var isRegionOverrideEnabled bool
	if regionSpec := spec.Region; !tfunique.IsHandleNil(regionSpec) && regionSpec.Value().IsOverrideEnabled {
		isRegionOverrideEnabled = true
	}

	if isRegionOverrideEnabled {
		interceptors = append(interceptors, listResourceInjectRegionAttribute())
		// TODO: validate region in partition, needs tweaked error message
	}

	inner := spec.Factory()

	if v, ok := inner.(framework.Identityer); ok {
		v.SetIdentitySpec(spec.Identity)
	}

	if v, ok := inner.(framework.Lister[listresource.InterceptorParams]); ok {
		if isRegionOverrideEnabled {
			v.AppendResultInterceptor(listresource.SetRegionInterceptor())
		}

		v.AppendResultInterceptor(listresource.IdentityInterceptor(spec.Identity.Attributes))

		// interceptor to set default types for tags, tags_all, and timeouts objects
		v.AppendResultInterceptor(listresource.DefaultObjectInterceptor())

		if !tfunique.IsHandleNil(spec.Tags) {
			v.AppendResultInterceptor(listresource.TagsInterceptor(spec.Tags))
		}
	}

	return &wrappedListResourceFramework{
		inner:              inner,
		servicePackageName: servicePackageName,
		spec:               spec,
		interceptors:       interceptors,
	}
}

// context is run on all wrapped methods before any interceptors.
func (w *wrappedListResourceFramework) context(ctx context.Context, getAttribute getAttributeFunc, c *conns.AWSClient) (context.Context, diag.Diagnostics) {
	var diags diag.Diagnostics
	var overrideRegion string

	var isRegionOverrideEnabled bool
	if regionSpec := w.spec.Region; !tfunique.IsHandleNil(regionSpec) && regionSpec.Value().IsOverrideEnabled {
		isRegionOverrideEnabled = true
	}

	if isRegionOverrideEnabled && getAttribute != nil {
		var target types.String
		diags.Append(getAttribute(ctx, path.Root(names.AttrRegion), &target)...)
		if diags.HasError() {
			return ctx, diags
		}

		if target.IsNull() || target.IsUnknown() {
			overrideRegion = c.AwsConfig(ctx).Region
		} else {
			overrideRegion = target.ValueString()
		}
	}

	ctx = conns.NewResourceContext(ctx, w.servicePackageName, w.spec.Name, w.spec.TypeName, overrideRegion)
	if c != nil {
		ctx = tftags.NewContext(ctx, c.DefaultTagsConfig(ctx), c.IgnoreTagsConfig(ctx), c.TagPolicyConfig(ctx))
		ctx = c.RegisterLogger(ctx)
		ctx = fwflex.RegisterLogger(ctx)
	}

	return ctx, diags
}

func (w *wrappedListResourceFramework) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		w.meta = v
	}

	ctx, diags := w.context(ctx, nil, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	w.inner.Configure(ctx, request, response)
}

func (w *wrappedListResourceFramework) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	stream.Results = tfiter.Null[list.ListResult]()

	ctx, diags := w.context(ctx, request.Config.GetAttribute, w.meta)
	if len(diags) > 0 {
		stream.Results = tfiter.Concat(stream.Results, list.ListResultsStreamDiagnostics(diags))
	}
	if diags.HasError() {
		return
	}

	interceptedListHandler(w.interceptors.resourceList(), w.inner.List, w.meta)(ctx, request, stream)
}

// ListResourceConfigSchema implements list.ListResourceWithConfigure.
func (w *wrappedListResourceFramework) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	ctx, diags := w.context(ctx, nil, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	interceptedHandler(w.interceptors.resourceListResourceConfigSchema(), w.inner.ListResourceConfigSchema, listResourceConfigSchemaHasError, w.meta)(ctx, request, response)
}

// Metadata implements list.ListResourceWithConfigure.
func (w *wrappedListResourceFramework) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	// This method does not call down to the inner resource.
	response.TypeName = w.spec.TypeName
}

type wrappedListResourceSDK struct {
	inner              inttypes.ListResourceForSDK
	meta               *conns.AWSClient
	servicePackageName string
	spec               *inttypes.ServicePackageSDKListResource
	interceptors       interceptorInvocations
}

var _ inttypes.ListResourceForSDK = &wrappedListResourceSDK{}

func newWrappedListResourceSDK(spec *inttypes.ServicePackageSDKListResource, servicePackageName string) inttypes.ListResourceForSDK {
	var interceptors interceptorInvocations

	if v := spec.Region; !tfunique.IsHandleNil(v) && v.Value().IsOverrideEnabled {
		interceptors = append(interceptors, listResourceInjectRegionAttribute())
		// TODO: validate region in partition, needs tweaked error message
	}

	inner := spec.Factory()

	if v, ok := inner.(framework.WithRegionSpec); ok {
		v.SetRegionSpec(spec.Region)
	}

	if v, ok := inner.(framework.Identityer); ok {
		v.SetIdentitySpec(spec.Identity)
	}

	if v, ok := inner.(framework.Lister[listresource.InterceptorParamsSDK]); ok {
		if !tfunique.IsHandleNil(spec.Tags) {
			v.AppendResultInterceptor(listresource.TagsInterceptorSDK(spec.Tags))
		}
	}

	return &wrappedListResourceSDK{
		inner:              inner,
		servicePackageName: servicePackageName,
		spec:               spec,
		interceptors:       interceptors,
	}
}

// Metadata implements list.ListResourceWithConfigure.
func (w *wrappedListResourceSDK) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	// This method does not call down to the inner resource.
	response.TypeName = w.spec.TypeName
}

// context is run on all wrapped methods before any interceptors.
func (w *wrappedListResourceSDK) context(ctx context.Context, getAttribute getAttributeFunc, c *conns.AWSClient) (context.Context, diag.Diagnostics) {
	var diags diag.Diagnostics
	var overrideRegion string

	var isRegionOverrideEnabled bool
	if regionSpec := w.spec.Region; !tfunique.IsHandleNil(regionSpec) && regionSpec.Value().IsOverrideEnabled {
		isRegionOverrideEnabled = true
	}

	if isRegionOverrideEnabled && getAttribute != nil {
		var target types.String
		diags.Append(getAttribute(ctx, path.Root(names.AttrRegion), &target)...)
		if diags.HasError() {
			return ctx, diags
		}

		if target.IsNull() || target.IsUnknown() {
			overrideRegion = c.AwsConfig(ctx).Region
		} else {
			overrideRegion = target.ValueString()
		}
	}

	ctx = conns.NewResourceContext(ctx, w.servicePackageName, w.spec.Name, w.spec.TypeName, overrideRegion)
	if c != nil {
		ctx = tftags.NewContext(ctx, c.DefaultTagsConfig(ctx), c.IgnoreTagsConfig(ctx), c.TagPolicyConfig(ctx))
		ctx = c.RegisterLogger(ctx)
		ctx = fwflex.RegisterLogger(ctx)
	}

	return ctx, diags
}

func (w *wrappedListResourceSDK) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		w.meta = v
	}

	ctx, diags := w.context(ctx, nil, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	w.inner.Configure(ctx, request, response)
}

func (w *wrappedListResourceSDK) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	stream.Results = tfiter.Null[list.ListResult]()

	ctx, diags := w.context(ctx, request.Config.GetAttribute, w.meta)
	if len(diags) > 0 {
		stream.Results = tfiter.Concat(stream.Results, list.ListResultsStreamDiagnostics(diags))
	}
	if diags.HasError() {
		return
	}

	interceptedListHandler(w.interceptors.resourceList(), w.inner.List, w.meta)(ctx, request, stream)
}

// ListResourceConfigSchema implements list.ListResourceWithConfigure.
func (w *wrappedListResourceSDK) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	ctx, diags := w.context(ctx, nil, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	interceptedHandler(w.interceptors.resourceListResourceConfigSchema(), w.inner.ListResourceConfigSchema, listResourceConfigSchemaHasError, w.meta)(ctx, request, response)
}

func (w *wrappedListResourceSDK) RawV5Schemas(ctx context.Context, request list.RawV5SchemaRequest, response *list.RawV5SchemaResponse) {
	if v, ok := w.inner.(list.ListResourceWithRawV5Schemas); ok {
		ctx, diags := w.context(ctx, nil, w.meta)
		if diags.HasError() {
			tflog.Warn(ctx, "wrapping Schemas", map[string]any{
				"resource":               w.spec.TypeName,
				"bootstrapContext error": fwdiag.DiagnosticsString(diags),
			})
		}

		v.RawV5Schemas(ctx, request, response)
	}
}
