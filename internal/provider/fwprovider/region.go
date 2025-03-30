// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwprovider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

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
