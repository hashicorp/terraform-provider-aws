// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
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

// WithTimeouts is intended to be embedded in resources which use the special "timeouts" nested block.
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

func mapHasUnknownElements(m types.Map) bool {
	for _, v := range m.Elements() {
		if v.IsUnknown() {
			return true
		}
	}

	return false
}
