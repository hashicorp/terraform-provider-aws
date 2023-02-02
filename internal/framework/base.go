package framework

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

type withMeta struct {
	meta *conns.AWSClient
}

func (w *withMeta) Meta() *conns.AWSClient {
	return w.meta
}

type withMigratedFromPluginSDK struct {
	migrated bool
}

// SetMigratedFromPluginSDK sets whether or not the resource (or data source) has been migrated from terraform-plugin-sdk.
func (w *withMigratedFromPluginSDK) SetMigratedFromPluginSDK(migrated bool) {
	w.migrated = migrated
}

// MigratedFromPluginSDK returns whether or not the resource (or data source) has been migrated from terraform-plugin-sdk.
func (w *withMigratedFromPluginSDK) MigratedFromPluginSDK() bool {
	return w.migrated
}

// ResourceWithConfigure is a structure to be embedded within a Resource that implements the ResourceWithConfigure interface.
type ResourceWithConfigure struct {
	withMeta
	withMigratedFromPluginSDK
}

// Configure enables provider-level data or clients to be set in the
// provider-defined Resource type.
func (r *ResourceWithConfigure) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		r.meta = v
	}
}

// ExpandTags returns the API tags for the specified "tags" value.
func (r *ResourceWithConfigure) ExpandTags(ctx context.Context, tags types.Map) tftags.KeyValueTags {
	return r.Meta().DefaultTagsConfig.MergeTags(tftags.New(tags))
}

// FlattenTags returns the "tags" value from the specified API tags.
func (r *ResourceWithConfigure) FlattenTags(ctx context.Context, apiTags tftags.KeyValueTags) types.Map {
	// AWS APIs often return empty lists of tags when none have been configured.
	if v := apiTags.IgnoreAWS().IgnoreConfig(r.Meta().IgnoreTagsConfig).RemoveDefaultConfig(r.Meta().DefaultTagsConfig).Map(); len(v) == 0 {
		return tftags.Null
	} else {
		return flex.FlattenFrameworkStringValueMapLegacy(ctx, v)
	}
}

// FlattenTagsAll returns the "tags_all" value from the specified API tags.
func (r *ResourceWithConfigure) FlattenTagsAll(ctx context.Context, apiTags tftags.KeyValueTags) types.Map {
	return flex.FlattenFrameworkStringValueMapLegacy(ctx, apiTags.IgnoreAWS().IgnoreConfig(r.Meta().IgnoreTagsConfig).Map())
}

// SetTagsAll calculates the new value for the `tags_all` attribute.
func (r *ResourceWithConfigure) SetTagsAll(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	defaultTagsConfig := r.Meta().DefaultTagsConfig
	ignoreTagsConfig := r.Meta().IgnoreTagsConfig

	var planTags types.Map

	response.Diagnostics.Append(request.Plan.GetAttribute(ctx, path.Root("tags"), &planTags)...)

	if response.Diagnostics.HasError() {
		return
	}

	if !planTags.IsUnknown() {
		resourceTags := tftags.New(planTags)

		if defaultTagsConfig.TagsEqual(resourceTags) {
			response.Diagnostics.AddError(
				`"tags" are identical to those in the "default_tags" configuration block of the provider`,
				"please de-duplicate and try again")
		}

		allTags := defaultTagsConfig.MergeTags(resourceTags).IgnoreConfig(ignoreTagsConfig)

		response.Diagnostics.Append(response.Plan.SetAttribute(ctx, path.Root("tags_all"), flex.FlattenFrameworkStringValueMapLegacy(ctx, allTags.Map()))...)
	} else {
		response.Diagnostics.Append(response.Plan.SetAttribute(ctx, path.Root("tags_all"), tftags.Unknown)...)
	}
}

// DataSourceWithConfigure is a structure to be embedded within a DataSource that implements the DataSourceWithConfigure interface.
type DataSourceWithConfigure struct {
	withMeta
	withMigratedFromPluginSDK
}

// Configure enables provider-level data or clients to be set in the
// provider-defined DataSource type. It is separately executed for each
// ReadDataSource RPC.
func (d *DataSourceWithConfigure) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		d.meta = v
	}
}

// WithTimeouts is intended to be embedded in resource which use the special "timeouts" nested block.
// See https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts.
type WithTimeouts struct {
	defaultCreateTimeout, defaultReadTimeout, defaultUpdateTimeout, defaultDeleteTimeout time.Duration
}

// SetDefaultCreateTimeout sets the resource's default Create timeout value.
func (w *WithTimeouts) SetDefaultCreateTimeout(timeout time.Duration) {
	w.defaultCreateTimeout = timeout
}

// SetDefaultReadTimeout sets the resource's default Read timeout value.
func (w *WithTimeouts) SetDefaultReadTimeout(timeout time.Duration) {
	w.defaultReadTimeout = timeout
}

// SetDefaultUpdateTimeout sets the resource's default Update timeout value.
func (w *WithTimeouts) SetDefaultUpdateTimeout(timeout time.Duration) {
	w.defaultUpdateTimeout = timeout
}

// SetDefaultDeleteTimeout sets the resource's default Delete timeout value.
func (w *WithTimeouts) SetDefaultDeleteTimeout(timeout time.Duration) {
	w.defaultDeleteTimeout = timeout
}

// CreateTimeout returns any configured Create timeout value or the default value.
func (w *WithTimeouts) CreateTimeout(ctx context.Context, timeouts timeouts.Value) time.Duration {
	timeout, diags := timeouts.Create(ctx, w.defaultCreateTimeout)

	if errors := diags.Errors(); len(errors) > 0 {
		tflog.Warn(ctx, "reading configured Create timeout", map[string]interface{}{
			"summary": errors[0].Summary(),
			"detail":  errors[0].Detail(),
		})

		return w.defaultCreateTimeout
	}

	return timeout
}

// ReadTimeout returns any configured Read timeout value or the default value.
func (w *WithTimeouts) ReadTimeout(ctx context.Context, timeouts timeouts.Value) time.Duration {
	timeout, diags := timeouts.Read(ctx, w.defaultReadTimeout)

	if errors := diags.Errors(); len(errors) > 0 {
		tflog.Warn(ctx, "reading configured Read timeout", map[string]interface{}{
			"summary": errors[0].Summary(),
			"detail":  errors[0].Detail(),
		})

		return w.defaultReadTimeout
	}

	return timeout
}

// UpdateTimeout returns any configured Update timeout value or the default value.
func (w *WithTimeouts) UpdateTimeout(ctx context.Context, timeouts timeouts.Value) time.Duration {
	timeout, diags := timeouts.Update(ctx, w.defaultUpdateTimeout)

	if errors := diags.Errors(); len(errors) > 0 {
		tflog.Warn(ctx, "reading configured Update timeout", map[string]interface{}{
			"summary": errors[0].Summary(),
			"detail":  errors[0].Detail(),
		})

		return w.defaultUpdateTimeout
	}

	return timeout
}

// DeleteTimeout returns any configured Delete timeout value or the default value.
func (w *WithTimeouts) DeleteTimeout(ctx context.Context, timeouts timeouts.Value) time.Duration {
	timeout, diags := timeouts.Delete(ctx, w.defaultDeleteTimeout)

	if errors := diags.Errors(); len(errors) > 0 {
		tflog.Warn(ctx, "reading configured Delete timeout", map[string]interface{}{
			"summary": errors[0].Summary(),
			"detail":  errors[0].Detail(),
		})

		return w.defaultDeleteTimeout
	}

	return timeout
}
