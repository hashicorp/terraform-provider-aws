// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwprovider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	erschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func validateRegionInPartition(ctx context.Context, c *conns.AWSClient, region string) error {
	if got, want := names.PartitionForRegion(region).ID(), c.Partition(ctx); got != want {
		return fmt.Errorf("partition (%s) for per-resource Region (%s) is not the provider's configured partition (%s)", got, region, want)
	}

	return nil
}

func validateInContextRegionInPartition(ctx context.Context, c *conns.AWSClient) diag.Diagnostics {
	var diags diag.Diagnostics

	// Verify that the value of the top-level `region` attribute is in the configured AWS partition.
	if inContext, ok := conns.FromContext(ctx); ok {
		if v := inContext.OverrideRegion(); v != "" {
			if err := validateRegionInPartition(ctx, c, v); err != nil {
				diags.AddAttributeError(path.Root(names.AttrRegion), "Invalid Region Value", err.Error())
			}
		}
	}

	return diags
}

type dataSourceInjectRegionAttributeInterceptor struct{}

func (r dataSourceInjectRegionAttributeInterceptor) schema(ctx context.Context, opts interceptorOptions[datasource.SchemaRequest, datasource.SchemaResponse]) diag.Diagnostics {
	var diags diag.Diagnostics

	switch response, when := opts.response, opts.when; when {
	case After:
		if _, ok := response.Schema.Attributes[names.AttrRegion]; !ok {
			// Inject a top-level "region" attribute.
			response.Schema.Attributes[names.AttrRegion] = dsschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: `The AWS Region to use for API operations. Overrides the Region set in the provider configuration.`,
			}
		}
	}

	return diags
}

// dataSourceInjectRegionAttribute injects a top-level "region" attribute into a data source's schema.
func dataSourceInjectRegionAttribute() dataSourceSchemaInterceptor {
	return &dataSourceInjectRegionAttributeInterceptor{}
}

type dataSourceValidateRegionInterceptor struct{}

func (r dataSourceValidateRegionInterceptor) read(ctx context.Context, opts interceptorOptions[datasource.ReadRequest, datasource.ReadResponse]) diag.Diagnostics {
	c := opts.c
	var diags diag.Diagnostics

	switch when := opts.when; when {
	case Before:
		// As data sources have no ModifyPlan functionality we validate the per-resource Region override value before R.
		diags.Append(validateInContextRegionInPartition(ctx, c)...)
		if diags.HasError() {
			return diags
		}
	}

	return diags
}

// dataSourceValidateRegion validates that the value of the top-level `region` attribute is in the configured AWS partition.
func dataSourceValidateRegion() dataSourceCRUDInterceptor {
	return &dataSourceValidateRegionInterceptor{}
}

type dataSourceSetRegionInStateInterceptor struct{}

func (r dataSourceSetRegionInStateInterceptor) read(ctx context.Context, opts interceptorOptions[datasource.ReadRequest, datasource.ReadResponse]) diag.Diagnostics {
	c := opts.c
	var diags diag.Diagnostics

	switch response, when := opts.response, opts.when; when {
	case After:
		// Set region in state after R, but only if the data source didn't explicitly set it (e.g. aws_region).
		var target types.String
		diags.Append(response.State.GetAttribute(ctx, path.Root(names.AttrRegion), &target)...)
		if diags.HasError() {
			return diags
		}

		if target.IsNull() {
			diags.Append(response.State.SetAttribute(ctx, path.Root(names.AttrRegion), c.Region(ctx))...)
			if diags.HasError() {
				return diags
			}
		}
	}

	return diags
}

// dataSourceSetRegionInState set the value of the top-level `region` attribute in state after Read.
func dataSourceSetRegionInState() dataSourceCRUDInterceptor {
	return &dataSourceSetRegionInStateInterceptor{}
}

type ephemeralResourceInjectRegionAttributeInterceptor struct{}

func (r ephemeralResourceInjectRegionAttributeInterceptor) schema(ctx context.Context, opts interceptorOptions[ephemeral.SchemaRequest, ephemeral.SchemaResponse]) diag.Diagnostics {
	var diags diag.Diagnostics

	switch response, when := opts.response, opts.when; when {
	case After:
		if _, ok := response.Schema.Attributes[names.AttrRegion]; !ok {
			// Inject a top-level "region" attribute.
			response.Schema.Attributes[names.AttrRegion] = erschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: `The AWS Region to use for API operations. Overrides the Region set in the provider configuration.`,
			}
		}
	}

	return diags
}

// ephemeralResourceInjectRegionAttribute injects a top-level "region" attribute into an ephemeral resource's schema.
func ephemeralResourceInjectRegionAttribute() ephemeralResourceSchemaInterceptor {
	return &ephemeralResourceInjectRegionAttributeInterceptor{}
}

type ephemeralResourceSetRegionInStateInterceptor struct {
	ephemeralResourceNoOpORCInterceptor
}

func (r ephemeralResourceSetRegionInStateInterceptor) open(ctx context.Context, opts interceptorOptions[ephemeral.OpenRequest, ephemeral.OpenResponse]) diag.Diagnostics {
	c := opts.c
	var diags diag.Diagnostics

	switch response, when := opts.response, opts.when; when {
	case After:
		// Set region in state after R.
		diags.Append(response.Result.SetAttribute(ctx, path.Root(names.AttrRegion), c.Region(ctx))...)
		if diags.HasError() {
			return diags
		}
	}

	return diags
}

// ephemeralResourceSetRegionInResult set the value of the top-level `region` attribute in the result after Open.
func ephemeralResourceSetRegionInResult() ephemeralResourceORCInterceptor {
	return &ephemeralResourceSetRegionInStateInterceptor{}
}

type ephemeralResourceValidateRegionInterceptor struct {
	ephemeralResourceNoOpORCInterceptor
}

func (r ephemeralResourceValidateRegionInterceptor) open(ctx context.Context, opts interceptorOptions[ephemeral.OpenRequest, ephemeral.OpenResponse]) diag.Diagnostics {
	c := opts.c
	var diags diag.Diagnostics

	switch when := opts.when; when {
	case Before:
		// As ephemeral resources have no ModifyPlan functionality we validate the per-resource Region override value here.
		diags.Append(validateInContextRegionInPartition(ctx, c)...)
		if diags.HasError() {
			return diags
		}
	}

	return diags
}

// ephemeralResourceValidateRegion validates that the value of the top-level `region` attribute is in the configured AWS partition.
func ephemeralResourceValidateRegion() ephemeralResourceORCInterceptor {
	return &ephemeralResourceValidateRegionInterceptor{}
}

type resourceInjectRegionAttributeInterceptor struct{}

func (r resourceInjectRegionAttributeInterceptor) schema(ctx context.Context, opts interceptorOptions[resource.SchemaRequest, resource.SchemaResponse]) diag.Diagnostics {
	var diags diag.Diagnostics

	switch response, when := opts.response, opts.when; when {
	case After:
		if _, ok := response.Schema.Attributes[names.AttrRegion]; !ok {
			// Inject a top-level "region" attribute.
			response.Schema.Attributes[names.AttrRegion] = rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: `The AWS Region to use for API operations. Overrides the Region set in the provider configuration.`,
			}
		}
	}

	return diags
}

// resourceInjectRegionAttribute injects a top-level "region" attribute into a resource's schema.
func resourceInjectRegionAttribute() resourceSchemaInterceptor {
	return &resourceInjectRegionAttributeInterceptor{}
}

type resourceValidateRegionInterceptor struct{}

func (r resourceValidateRegionInterceptor) modifyPlan(ctx context.Context, opts interceptorOptions[resource.ModifyPlanRequest, resource.ModifyPlanResponse]) diag.Diagnostics {
	c := opts.c
	var diags diag.Diagnostics

	switch when := opts.when; when {
	case Before:
		diags.Append(validateInContextRegionInPartition(ctx, c)...)
		if diags.HasError() {
			return diags
		}
	}

	return diags
}

// resourceValidateRegion validates that the value of the top-level `region` attribute is in the configured AWS partition.
func resourceValidateRegion() resourceModifyPlanInterceptor {
	return &resourceValidateRegionInterceptor{}
}

type resourceDefaultRegionInterceptor struct{}

func (r resourceDefaultRegionInterceptor) modifyPlan(ctx context.Context, opts interceptorOptions[resource.ModifyPlanRequest, resource.ModifyPlanResponse]) diag.Diagnostics {
	c := opts.c
	var diags diag.Diagnostics

	switch request, response, when := opts.request, opts.response, opts.when; when {
	case Before:
		// If the entire plan is null, the resource is planned for destruction.
		if request.Plan.Raw.IsNull() {
			return diags
		}

		var target types.String
		diags.Append(request.Plan.GetAttribute(ctx, path.Root(names.AttrRegion), &target)...)
		if diags.HasError() {
			return diags
		}

		if target.IsNull() || target.IsUnknown() {
			// Set the region to the provider's configured region
			diags.Append(response.Plan.SetAttribute(ctx, path.Root(names.AttrRegion), c.AwsConfig(ctx).Region)...)
			if diags.HasError() {
				return diags
			}
		}
	}

	return diags
}

// resourceDefaultRegion sets the value of the top-level `region` attribute to the provider's configured Region if it is not set.
func resourceDefaultRegion() resourceModifyPlanInterceptor {
	return &resourceDefaultRegionInterceptor{}
}

type resourceForceNewIfRegionChangesInterceptor struct{}

func (r resourceForceNewIfRegionChangesInterceptor) modifyPlan(ctx context.Context, opts interceptorOptions[resource.ModifyPlanRequest, resource.ModifyPlanResponse]) diag.Diagnostics {
	c := opts.c
	var diags diag.Diagnostics

	switch request, response, when := opts.request, opts.response, opts.when; when {
	case Before:
		// If the entire plan is null, the resource is planned for destruction.
		if request.Plan.Raw.IsNull() {
			return diags
		}

		// If the entire state is null, the resource is new.
		if request.State.Raw.IsNull() {
			return diags
		}

		var planRegion types.String
		diags.Append(request.Plan.GetAttribute(ctx, path.Root(names.AttrRegion), &planRegion)...)
		if diags.HasError() {
			return diags
		}

		var stateRegion types.String
		diags.Append(request.State.GetAttribute(ctx, path.Root(names.AttrRegion), &stateRegion)...)
		if diags.HasError() {
			return diags
		}

		providerRegion := c.AwsConfig(ctx).Region
		if stateRegion.IsNull() && planRegion.ValueString() == providerRegion {
			return diags
		}

		if !planRegion.Equal(stateRegion) {
			response.RequiresReplace = path.Paths{path.Root(names.AttrRegion)}
		}
	}

	return diags
}

// resourceForceNewIfRegionChanges forces resource replacement if the value of the top-level `region` attribute changes.
func resourceForceNewIfRegionChanges() resourceModifyPlanInterceptor {
	return &resourceForceNewIfRegionChangesInterceptor{}
}

type resourceSetRegionInStateInterceptor struct {
	resourceNoOpCRUDInterceptor
}

func (r resourceSetRegionInStateInterceptor) read(ctx context.Context, opts interceptorOptions[resource.ReadRequest, resource.ReadResponse]) diag.Diagnostics {
	c := opts.c
	var diags diag.Diagnostics

	switch response, when := opts.response, opts.when; when {
	case After:
		// Set region in state after R.
		diags.Append(response.State.SetAttribute(ctx, path.Root(names.AttrRegion), c.Region(ctx))...)
		if diags.HasError() {
			return diags
		}
	}

	return diags
}

// ephemeralResourceSetRegionInState set the value of the top-level `region` attribute in state after Read.
func resourceSetRegionInState() resourceCRUDInterceptor {
	return &resourceSetRegionInStateInterceptor{}
}
