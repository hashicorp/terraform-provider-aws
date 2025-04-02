// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwprovider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// resourceInjectRegionAttribute injects a top-level "region" attribute.
func resourceInjectRegionAttribute(ctx context.Context, c *conns.AWSClient, request resource.SchemaRequest, response *resource.SchemaResponse) {
	if _, ok := response.Schema.Attributes[names.AttrRegion]; !ok {
		// Inject a top-level "region" attribute.
		response.Schema.Attributes[names.AttrRegion] = rschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: `The AWS Region to use for API operations. Overrides the Region set in the provider configuration.`,
		}
	}
}

// validateRegionValueInConfiguredPartition is a config validator that validates that the value of
// the top-level `region` attribute is in the configured AWS partition.
func validateRegionValueInConfiguredPartition(ctx context.Context, c *conns.AWSClient, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var configRegion types.String
	response.Diagnostics.Append(request.Config.GetAttribute(ctx, path.Root(names.AttrRegion), &configRegion)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !configRegion.IsNull() && !configRegion.IsUnknown() {
		if err := validateRegionInPartition(ctx, c, configRegion.ValueString()); err != nil {
			response.Diagnostics.AddAttributeError(path.Root(names.AttrRegion), "Invalid Region Value", err.Error())
		}
	}
}

func validateRegionInPartition(ctx context.Context, c *conns.AWSClient, region string) error {
	if got, want := names.PartitionForRegion(region).ID(), c.Partition(ctx); got != want {
		return fmt.Errorf("partition (%s) for per-resource Region (%s) is not the provider's configured partition (%s)", got, region, want)
	}

	return nil
}

// defaultRegionValue is a plan modifier that sets the value of the top-level `region`
// attribute to the provider's configured Region if it is not set.
func defaultRegionValue(ctx context.Context, c *conns.AWSClient, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	// If the entire plan is null, the resource is planned for destruction.
	if request.Plan.Raw.IsNull() {
		return
	}

	var planRegion types.String
	response.Diagnostics.Append(request.Plan.GetAttribute(ctx, path.Root(names.AttrRegion), &planRegion)...)
	if response.Diagnostics.HasError() {
		return
	}

	if planRegion.IsNull() || planRegion.IsUnknown() {
		// Set the region to the provider's configured region
		response.Diagnostics.Append(response.Plan.SetAttribute(ctx, path.Root(names.AttrRegion), c.AwsConfig(ctx).Region)...)
		if response.Diagnostics.HasError() {
			return
		}
	}
}

// forceNewIfRegionValueChanges is a plan modifier that forces resource replacement
// if the value of the top-level `region` attribute changes.
func forceNewIfRegionValueChanges(ctx context.Context, c *conns.AWSClient, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	// If the entire plan is null, the resource is planned for destruction.
	if request.Plan.Raw.IsNull() {
		return
	}

	// If the entire state is null, the resource is new.
	if request.State.Raw.IsNull() {
		return
	}

	var planRegion types.String
	response.Diagnostics.Append(request.Plan.GetAttribute(ctx, path.Root(names.AttrRegion), &planRegion)...)
	if response.Diagnostics.HasError() {
		return
	}

	var stateRegion types.String
	response.Diagnostics.Append(request.State.GetAttribute(ctx, path.Root(names.AttrRegion), &stateRegion)...)
	if response.Diagnostics.HasError() {
		return
	}

	providerRegion := c.AwsConfig(ctx).Region
	if stateRegion.IsNull() && planRegion.ValueString() == providerRegion {
		return
	}

	if !planRegion.Equal(stateRegion) {
		response.RequiresReplace = path.Paths{path.Root(names.AttrRegion)}
	}
}

// regionDataSourceInterceptor implements per-resource Region override functionality for data sources.
type regionDataSourceInterceptor struct {
	validateRegionInPartition bool
}

func newRegionDataSourceInterceptor(validateRegionInPartition bool) dataSourceInterceptor {
	return &regionDataSourceInterceptor{
		validateRegionInPartition: validateRegionInPartition,
	}
}

func (r regionDataSourceInterceptor) read(ctx context.Context, opts interceptorOptions[datasource.ReadRequest, datasource.ReadResponse]) diag.Diagnostics {
	c := opts.c
	var diags diag.Diagnostics

	switch response, when := opts.response, opts.when; when {
	case Before:
		// As data sources have no ModifyPlan functionality we validate the per-resource Region override value here.
		if r.validateRegionInPartition {
			if inContext, ok := conns.FromContext(ctx); ok {
				if v := inContext.OverrideRegion(); v != "" {
					if err := validateRegionInPartition(ctx, c, v); err != nil {
						diags.AddAttributeError(path.Root(names.AttrRegion), "Invalid Region Value", err.Error())

						return diags
					}
				}
			}
		}
	case After:
		// Set region in state after R.
		diags.Append(response.State.SetAttribute(ctx, path.Root(names.AttrRegion), c.Region(ctx))...)
		if diags.HasError() {
			return diags
		}
	}

	return diags
}

// regionEphemeralResourceInterceptor implements per-resource Region override functionality for ephemeral resources.
type regionEphemeralResourceInterceptor struct {
	validateRegionInPartition bool
}

func newRegionEphemeralResourceInterceptor(validateRegionInPartition bool) ephemeralResourceInterceptor {
	return &regionEphemeralResourceInterceptor{
		validateRegionInPartition: validateRegionInPartition,
	}
}

func (r regionEphemeralResourceInterceptor) open(ctx context.Context, opts interceptorOptions[ephemeral.OpenRequest, ephemeral.OpenResponse]) diag.Diagnostics {
	c := opts.c
	var diags diag.Diagnostics

	switch response, when := opts.response, opts.when; when {
	case Before:
		// As data sources have no ModifyPlan functionality we validate the per-resource Region override value here.
		if r.validateRegionInPartition {
			if inContext, ok := conns.FromContext(ctx); ok {
				if v := inContext.OverrideRegion(); v != "" {
					if err := validateRegionInPartition(ctx, c, v); err != nil {
						diags.AddAttributeError(path.Root(names.AttrRegion), "Invalid Region Value", err.Error())

						return diags
					}
				}
			}
		}
	case After:
		// Set region in state after R.
		diags.Append(response.Result.SetAttribute(ctx, path.Root(names.AttrRegion), c.Region(ctx))...)
		if diags.HasError() {
			return diags
		}
	}

	return diags
}

func (r regionEphemeralResourceInterceptor) renew(ctx context.Context, opts interceptorOptions[ephemeral.RenewRequest, ephemeral.RenewResponse]) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

func (r regionEphemeralResourceInterceptor) close(ctx context.Context, opts interceptorOptions[ephemeral.CloseRequest, ephemeral.CloseResponse]) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}
