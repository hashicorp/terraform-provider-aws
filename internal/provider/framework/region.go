// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-framework/action"
	aschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	erschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/datasourceattribute"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/listresourceattribute"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/resourceattribute"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func validateInContextRegionInPartition(ctx context.Context, c awsClient) diag.Diagnostics {
	var diags diag.Diagnostics

	if err := c.ValidateInContextRegionInPartition(ctx); err != nil {
		diags.AddAttributeError(path.Root(names.AttrRegion), "Invalid Region Value", err.Error())
	}

	return diags
}

type dataSourceInjectRegionAttributeInterceptor struct {
	isDeprecated bool
}

func (r dataSourceInjectRegionAttributeInterceptor) schema(ctx context.Context, opts interceptorOptions[datasource.SchemaRequest, datasource.SchemaResponse]) {
	switch response, when := opts.response, opts.when; when {
	case After:
		if _, ok := response.Schema.Attributes[names.AttrRegion]; !ok {
			// Inject a top-level "region" attribute.
			if r.isDeprecated {
				response.Schema.Attributes[names.AttrRegion] = datasourceattribute.RegionDeprecated()
			} else {
				response.Schema.Attributes[names.AttrRegion] = datasourceattribute.Region()
			}
		}
	}
}

// dataSourceInjectRegionAttribute injects a top-level "region" attribute into a data source's schema.
func dataSourceInjectRegionAttribute(isDeprecated bool) dataSourceSchemaInterceptor {
	return &dataSourceInjectRegionAttributeInterceptor{
		isDeprecated: isDeprecated,
	}
}

type dataSourceValidateRegionInterceptor struct{}

func (r dataSourceValidateRegionInterceptor) read(ctx context.Context, opts interceptorOptions[datasource.ReadRequest, datasource.ReadResponse]) {
	c := opts.c

	switch when := opts.when; when {
	case Before:
		// As data sources have no ModifyPlan functionality we validate the per-resource Region override value before R.
		opts.response.Diagnostics.Append(validateInContextRegionInPartition(ctx, c)...)
		if opts.response.Diagnostics.HasError() {
			return
		}
	}
}

// dataSourceValidateRegion validates that the value of the top-level `region` attribute is in the configured AWS partition.
func dataSourceValidateRegion() dataSourceCRUDInterceptor {
	return &dataSourceValidateRegionInterceptor{}
}

type dataSourceSetRegionInStateInterceptor struct{}

func (r dataSourceSetRegionInStateInterceptor) read(ctx context.Context, opts interceptorOptions[datasource.ReadRequest, datasource.ReadResponse]) {
	c := opts.c

	switch response, when := opts.response, opts.when; when {
	case After:
		// Set region in state after R, but only if the data source didn't explicitly set it (e.g. aws_region).
		var target types.String
		opts.response.Diagnostics.Append(response.State.GetAttribute(ctx, path.Root(names.AttrRegion), &target)...)
		if opts.response.Diagnostics.HasError() {
			return
		}

		if target.IsNull() {
			opts.response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrRegion), c.Region(ctx))...)
			if opts.response.Diagnostics.HasError() {
				return
			}
		}
	}
}

// dataSourceSetRegionInState set the value of the top-level `region` attribute in state after Read.
func dataSourceSetRegionInState() dataSourceCRUDInterceptor {
	return &dataSourceSetRegionInStateInterceptor{}
}

type ephemeralResourceInjectRegionAttributeInterceptor struct{}

func (r ephemeralResourceInjectRegionAttributeInterceptor) schema(ctx context.Context, opts interceptorOptions[ephemeral.SchemaRequest, ephemeral.SchemaResponse]) {
	switch response, when := opts.response, opts.when; when {
	case After:
		if _, ok := response.Schema.Attributes[names.AttrRegion]; !ok {
			// Inject a top-level "region" attribute.
			response.Schema.Attributes[names.AttrRegion] = erschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: names.ResourceTopLevelRegionAttributeDescription,
			}
		}
	}
}

// ephemeralResourceInjectRegionAttribute injects a top-level "region" attribute into an ephemeral resource's schema.
func ephemeralResourceInjectRegionAttribute() ephemeralResourceSchemaInterceptor {
	return &ephemeralResourceInjectRegionAttributeInterceptor{}
}

type ephemeralResourceSetRegionInStateInterceptor struct {
	ephemeralResourceNoOpORCInterceptor
}

func (r ephemeralResourceSetRegionInStateInterceptor) open(ctx context.Context, opts interceptorOptions[ephemeral.OpenRequest, ephemeral.OpenResponse]) {
	c := opts.c

	switch response, when := opts.response, opts.when; when {
	case After:
		// Set region in state after R.
		opts.response.Diagnostics.Append(response.Result.SetAttribute(ctx, path.Root(names.AttrRegion), c.Region(ctx))...)
		if opts.response.Diagnostics.HasError() {
			return
		}
	}
}

// ephemeralResourceSetRegionInResult set the value of the top-level `region` attribute in the result after Open.
func ephemeralResourceSetRegionInResult() ephemeralResourceORCInterceptor {
	return &ephemeralResourceSetRegionInStateInterceptor{}
}

type ephemeralResourceValidateRegionInterceptor struct {
	ephemeralResourceNoOpORCInterceptor
}

func (r ephemeralResourceValidateRegionInterceptor) open(ctx context.Context, opts interceptorOptions[ephemeral.OpenRequest, ephemeral.OpenResponse]) {
	c := opts.c

	switch when := opts.when; when {
	case Before:
		// As ephemeral resources have no ModifyPlan functionality we validate the per-resource Region override value here.
		opts.response.Diagnostics.Append(validateInContextRegionInPartition(ctx, c)...)
		if opts.response.Diagnostics.HasError() {
			return
		}
	}
}

// ephemeralResourceValidateRegion validates that the value of the top-level `region` attribute is in the configured AWS partition.
func ephemeralResourceValidateRegion() ephemeralResourceORCInterceptor {
	return &ephemeralResourceValidateRegionInterceptor{}
}

type resourceInjectRegionAttributeInterceptor struct {
	isDeprecated bool
}

func (r resourceInjectRegionAttributeInterceptor) schema(ctx context.Context, opts interceptorOptions[resource.SchemaRequest, resource.SchemaResponse]) {
	switch response, when := opts.response, opts.when; when {
	case After:
		if _, ok := response.Schema.Attributes[names.AttrRegion]; !ok {
			// Inject a top-level "region" attribute.
			if r.isDeprecated {
				response.Schema.Attributes[names.AttrRegion] = resourceattribute.RegionDeprecated()
			} else {
				response.Schema.Attributes[names.AttrRegion] = resourceattribute.Region()
			}
		}
	}
}

// resourceInjectRegionAttribute injects a top-level "region" attribute into a resource's schema.
func resourceInjectRegionAttribute(isDeprecated bool) resourceSchemaInterceptor {
	return &resourceInjectRegionAttributeInterceptor{
		isDeprecated: isDeprecated,
	}
}

type resourceValidateRegionInterceptor struct{}

func (r resourceValidateRegionInterceptor) modifyPlan(ctx context.Context, opts interceptorOptions[resource.ModifyPlanRequest, resource.ModifyPlanResponse]) {
	c := opts.c

	switch when := opts.when; when {
	case Before:
		opts.response.Diagnostics.Append(validateInContextRegionInPartition(ctx, c)...)
		if opts.response.Diagnostics.HasError() {
			return
		}
	}
}

// resourceValidateRegion validates that the value of the top-level `region` attribute is in the configured AWS partition.
func resourceValidateRegion() resourceModifyPlanInterceptor {
	return &resourceValidateRegionInterceptor{}
}

type resourceDefaultRegionInterceptor struct{}

func (r resourceDefaultRegionInterceptor) modifyPlan(ctx context.Context, opts interceptorOptions[resource.ModifyPlanRequest, resource.ModifyPlanResponse]) {
	c := opts.c

	switch request, response, when := opts.request, opts.response, opts.when; when {
	case Before:
		// If the entire plan is null, the resource is planned for destruction.
		if request.Plan.Raw.IsNull() {
			return
		}

		var target types.String
		opts.response.Diagnostics.Append(request.Plan.GetAttribute(ctx, path.Root(names.AttrRegion), &target)...)
		if opts.response.Diagnostics.HasError() {
			return
		}

		if target.IsNull() || target.IsUnknown() {
			// Set the region to the provider's configured region
			opts.response.Diagnostics.Append(response.Plan.SetAttribute(ctx, path.Root(names.AttrRegion), c.AwsConfig(ctx).Region)...)
			if opts.response.Diagnostics.HasError() {
				return
			}
		}
	}
}

// resourceDefaultRegion sets the value of the top-level `region` attribute to the provider's configured Region if it is not set.
func resourceDefaultRegion() resourceModifyPlanInterceptor {
	return &resourceDefaultRegionInterceptor{}
}

type resourceForceNewIfRegionChangesInterceptor struct{}

func (r resourceForceNewIfRegionChangesInterceptor) modifyPlan(ctx context.Context, opts interceptorOptions[resource.ModifyPlanRequest, resource.ModifyPlanResponse]) {
	c := opts.c

	switch request, response, when := opts.request, opts.response, opts.when; when {
	case Before:
		// If the entire plan is null, the resource is planned for destruction.
		if request.Plan.Raw.IsNull() {
			return
		}

		// If the entire state is null, the resource is new.
		if request.State.Raw.IsNull() {
			return
		}

		var configRegion types.String
		opts.response.Diagnostics.Append(request.Config.GetAttribute(ctx, path.Root(names.AttrRegion), &configRegion)...)
		if opts.response.Diagnostics.HasError() {
			return
		}

		var planRegion types.String
		opts.response.Diagnostics.Append(request.Plan.GetAttribute(ctx, path.Root(names.AttrRegion), &planRegion)...)
		if opts.response.Diagnostics.HasError() {
			return
		}

		var stateRegion types.String
		opts.response.Diagnostics.Append(request.State.GetAttribute(ctx, path.Root(names.AttrRegion), &stateRegion)...)
		if opts.response.Diagnostics.HasError() {
			return
		}

		if stateRegion.IsNull() {
			// Upgrade from pre-v6.0.0 provider and '-refresh=false' in effect.
			if configRegion.IsNull() {
				return
			}

			if providerRegion := c.AwsConfig(ctx).Region; planRegion.ValueString() == providerRegion {
				return
			}
		}

		if !planRegion.Equal(stateRegion) {
			response.RequiresReplace = path.Paths{path.Root(names.AttrRegion)}
		}
	}
}

// resourceForceNewIfRegionChanges forces resource replacement if the value of the top-level `region` attribute changes.
func resourceForceNewIfRegionChanges() resourceModifyPlanInterceptor {
	return &resourceForceNewIfRegionChangesInterceptor{}
}

type resourceSetRegionInStateInterceptor struct {
	resourceNoOpCRUDInterceptor
}

func (r resourceSetRegionInStateInterceptor) read(ctx context.Context, opts interceptorOptions[resource.ReadRequest, resource.ReadResponse]) {
	c := opts.c

	switch response, when := opts.response, opts.when; when {
	case After:
		// Will occur on a refresh when the resource does not exist in AWS and needs to be recreated, e.g. "_disappears" tests.
		if response.State.Raw.IsNull() {
			return
		}

		// Set region in state after R.
		opts.response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrRegion), c.Region(ctx))...)
		if opts.response.Diagnostics.HasError() {
			return
		}
	}
}

// resourceSetRegionInState set the value of the top-level `region` attribute in state after Read.
func resourceSetRegionInState() resourceCRUDInterceptor {
	return &resourceSetRegionInStateInterceptor{}
}

type resourceImportRegionInterceptor struct{}

func (r resourceImportRegionInterceptor) importState(ctx context.Context, opts interceptorOptions[resource.ImportStateRequest, resource.ImportStateResponse]) {
	c := opts.c

	switch request, response, when := opts.request, opts.response, opts.when; when {
	case Before:
		// Import ID optionally ends with "@<region>".
		if matches := regexache.MustCompile(`^(.+)@(` + inttypes.CanonicalRegionPatternNoAnchors + `)$`).FindStringSubmatch(request.ID); len(matches) == 3 {
			request.ID = matches[1]
			opts.response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrRegion), matches[2])...)
			if opts.response.Diagnostics.HasError() {
				return
			}
		} else {
			opts.response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrRegion), c.AwsConfig(ctx).Region)...)
			if opts.response.Diagnostics.HasError() {
				return
			}
		}
	}
}

// resourceImportRegion sets the value of the top-level `region` attribute during import.
func resourceImportRegion() resourceImportStateInterceptor {
	return &resourceImportRegionInterceptor{}
}

type resourceImportRegionNoDefaultInterceptor struct{}

func (r resourceImportRegionNoDefaultInterceptor) importState(ctx context.Context, opts interceptorOptions[resource.ImportStateRequest, resource.ImportStateResponse]) {
	switch request, response, when := opts.request, opts.response, opts.when; when {
	case Before:
		// Import ID optionally ends with "@<region>".
		if matches := regexache.MustCompile(`^(.+)@(` + inttypes.CanonicalRegionPatternNoAnchors + `)$`).FindStringSubmatch(request.ID); len(matches) == 3 {
			request.ID = matches[1]
			opts.response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrRegion), matches[2])...)
			if opts.response.Diagnostics.HasError() {
				return
			}
		}
	}
}

// resourceImportRegionNoDefault sets the value of the top-level `region` attribute during import.
func resourceImportRegionNoDefault() resourceImportStateInterceptor {
	return &resourceImportRegionNoDefaultInterceptor{}
}

type actionInjectRegionAttributeInterceptor struct{}

func (a actionInjectRegionAttributeInterceptor) schema(ctx context.Context, opts interceptorOptions[action.SchemaRequest, action.SchemaResponse]) {
	switch response, when := opts.response, opts.when; when {
	case After:
		if _, exists := response.Schema.Attributes[names.AttrRegion]; !exists {
			// Inject a top-level "region" attribute.
			if response.Schema.Attributes == nil {
				response.Schema.Attributes = make(map[string]aschema.Attribute)
			}
			response.Schema.Attributes[names.AttrRegion] = aschema.StringAttribute{
				Optional:    true,
				Description: names.ActionTopLevelRegionAttributeDescription,
			}
		}
	}
}

// actionInjectRegionAttribute injects a top-level "region" attribute into an action's schema.
func actionInjectRegionAttribute() actionSchemaInterceptor {
	return &actionInjectRegionAttributeInterceptor{}
}

type actionValidateRegionInterceptor struct {
}

func (a actionValidateRegionInterceptor) invoke(ctx context.Context, opts interceptorOptions[action.InvokeRequest, action.InvokeResponse]) {
	c := opts.c

	switch when := opts.when; when {
	case Before:
		opts.response.Diagnostics.Append(validateInContextRegionInPartition(ctx, c)...)
	}
}

// actionValidateRegion validates that the value of the top-level `region` attribute is in the configured AWS partition.
func actionValidateRegion() actionInvokeInterceptor {
	return &actionValidateRegionInterceptor{}
}

type listResourceInjectRegionAttributeInterceptor struct{}

func (r listResourceInjectRegionAttributeInterceptor) schema(ctx context.Context, opts interceptorOptions[list.ListResourceSchemaRequest, list.ListResourceSchemaResponse]) {
	switch response, when := opts.response, opts.when; when {
	case After:
		if _, ok := response.Schema.Attributes[names.AttrRegion]; !ok {
			// Inject a top-level "region" attribute.
			response.Schema.Attributes[names.AttrRegion] = listresourceattribute.Region()
		}
	}
}

// listResourceInjectRegionAttribute injects a "region" attribute into a resource's List schema.
func listResourceInjectRegionAttribute() listResourceSchemaInterceptor {
	return &listResourceInjectRegionAttributeInterceptor{}
}
