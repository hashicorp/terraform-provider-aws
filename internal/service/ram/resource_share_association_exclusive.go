// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ram

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ram"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ram/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @FrameworkResource("aws_ram_resource_share_association_exclusive", name="Resource Share Association Exclusive")
// @ArnIdentity("resource_share_arn")
func newResourceShareAssociationExclusiveResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceShareAssociationExclusiveResource{}, nil
}

type resourceShareAssociationExclusiveResource struct {
	framework.ResourceWithModel[resourceShareAssociationExclusiveResourceModel]
	framework.WithImportByIdentity
}

func (r *resourceShareAssociationExclusiveResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"principals": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						principalValidator(),
					),
				},
			},
			"resource_arns": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						validators.ARN(),
					),
				},
			},
			"resource_share_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"sources": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						validators.AWSAccountID(),
					),
				},
			},
		},
	}
}

// ValidateConfig validates the resource configuration.
// - Service principals cannot be mixed with other principal types (account IDs, ARNs)
// - Sources can only be specified when principals contains only service principals
func (r *resourceShareAssociationExclusiveResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config resourceShareAssociationExclusiveResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Skip validation if principals is null or unknown
	if config.Principals.IsNull() || config.Principals.IsUnknown() {
		// If sources is specified but principals is not, that's an error
		if !config.Sources.IsNull() && !config.Sources.IsUnknown() && !config.Sources.ContainsUnknownElements() {
			var sources []string
			resp.Diagnostics.Append(config.Sources.ElementsAs(ctx, &sources, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			if len(sources) > 0 {
				resp.Diagnostics.AddAttributeError(
					path.Root("sources"),
					"Invalid Configuration",
					"sources can only be specified when principals contains only service principals (e.g., service-id.amazonaws.com)",
				)
			}
		}
		return
	}

	// Skip validation if any element in the set is unknown (e.g., references to other resources)
	if config.Principals.ContainsUnknownElements() {
		return
	}

	var principals []string
	resp.Diagnostics.Append(config.Principals.ElementsAs(ctx, &principals, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Count service principals and non-service principals
	var servicePrincipals, otherPrincipals []string
	for _, principal := range principals {
		if inttypes.IsServicePrincipal(principal) {
			servicePrincipals = append(servicePrincipals, principal)
		} else {
			otherPrincipals = append(otherPrincipals, principal)
		}
	}

	// Validate that service principals are not mixed with other principal types
	if len(servicePrincipals) > 0 && len(otherPrincipals) > 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("principals"),
			"Invalid Configuration",
			"Service principals (e.g., service-id.amazonaws.com) cannot be mixed with other principal types (AWS account IDs, organization ARNs, OU ARNs, IAM role ARNs, or IAM user ARNs)",
		)
		return
	}

	// Validate sources - only allowed when principals contains only service principals
	if !config.Sources.IsNull() && !config.Sources.IsUnknown() && !config.Sources.ContainsUnknownElements() {
		var sources []string
		resp.Diagnostics.Append(config.Sources.ElementsAs(ctx, &sources, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if len(sources) > 0 && len(servicePrincipals) == 0 {
			resp.Diagnostics.AddAttributeError(
				path.Root("sources"),
				"Invalid Configuration",
				"sources can only be specified when principals contains only service principals (e.g., service-id.amazonaws.com)",
			)
		}
	}
}

func (r *resourceShareAssociationExclusiveResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceShareAssociationExclusiveResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RAMClient(ctx)

	// Verify resource share exists
	resourceShareARN := fwflex.StringValueFromFramework(ctx, plan.ResourceShareARN)
	_, err := findResourceShareOwnerSelfByARN(ctx, conn, resourceShareARN)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading RAM Resource Share (%s)", resourceShareARN), err.Error())
		return
	}

	// Get desired state
	wantPrincipals, wantResources, wantSources := fwflex.ExpandFrameworkStringValueSet(ctx, plan.Principals), fwflex.ExpandFrameworkStringValueSet(ctx, plan.ResourceARNs), fwflex.ExpandFrameworkStringValueSet(ctx, plan.Sources)

	// Get current state from AWS
	currentPrincipals, currentResources, err := findAssociationsForResourceShare(ctx, conn, resourceShareARN)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading RAM Resource Share (%s) Associations", resourceShareARN), err.Error())
		return
	}

	// Sync associations
	resp.Diagnostics.Append(r.syncAssociations(ctx, conn, resourceShareARN, currentPrincipals, currentResources, wantPrincipals, wantResources, wantSources)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceShareAssociationExclusiveResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceShareAssociationExclusiveResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RAMClient(ctx)

	// Verify resource share still exists
	resourceShareARN := fwflex.StringValueFromFramework(ctx, state.ResourceShareARN)
	_, err := findResourceShareOwnerSelfByARN(ctx, conn, resourceShareARN)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading RAM Resource Share (%s)", resourceShareARN), err.Error())
		return
	}

	// Get current associations from AWS
	principals, resources, err := findAssociationsForResourceShare(ctx, conn, resourceShareARN)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading RAM Resource Share (%s) Associations", resourceShareARN), err.Error())
		return
	}

	state.Principals = fwflex.FlattenFrameworkStringValueSetOfStringLegacy(ctx, principals)
	state.ResourceARNs = fwflex.FlattenFrameworkStringValueSetOfStringLegacy(ctx, resources)
	// Sources cannot be read from AWS API, preserve from state

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceShareAssociationExclusiveResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceShareAssociationExclusiveResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RAMClient(ctx)
	resourceShareARN := fwflex.StringValueFromFramework(ctx, plan.ResourceShareARN)

	// Get desired state
	wantPrincipals, wantResources, wantSources := fwflex.ExpandFrameworkStringValueSet(ctx, plan.Principals), fwflex.ExpandFrameworkStringValueSet(ctx, plan.ResourceARNs), fwflex.ExpandFrameworkStringValueSet(ctx, plan.Sources)

	// Get current state from AWS
	currentPrincipals, currentResources, err := findAssociationsForResourceShare(ctx, conn, resourceShareARN)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading RAM Resource Share (%s) Associations", resourceShareARN), err.Error())
		return
	}

	// Sync associations
	resp.Diagnostics.Append(r.syncAssociations(ctx, conn, resourceShareARN, currentPrincipals, currentResources, wantPrincipals, wantResources, wantSources)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceShareAssociationExclusiveResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceShareAssociationExclusiveResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RAMClient(ctx)

	// Get current associations
	resourceShareARN := fwflex.StringValueFromFramework(ctx, state.ResourceShareARN)
	principals, resources, err := findAssociationsForResourceShare(ctx, conn, resourceShareARN)
	if retry.NotFound(err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading RAM Resource Share (%s) associations", resourceShareARN), err.Error())
		return
	}

	// Remove all principals
	for _, principal := range principals {
		if err := deleteResourceSharePrincipalAssociation(ctx, conn, resourceShareARN, principal); err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("deleting RAM Resource Share (%s) Associations", resourceShareARN), err.Error())
			return
		}
	}

	// Remove all resources
	for _, resourceARN := range resources {
		if err := deleteResourceShareResourceAssociation(ctx, conn, resourceShareARN, resourceARN); err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("deleting RAM Resource Share (%s) Associations", resourceShareARN), err.Error())
			return
		}
	}
}

// syncAssociations synchronizes the configured principals and resources with AWS.
func (r *resourceShareAssociationExclusiveResource) syncAssociations(ctx context.Context, conn *ram.Client, resourceShareARN string, currentPrincipals, currentResources, wantPrincipals, wantResources, wantSources []string) diag.Diagnostics {
	var diags diag.Diagnostics

	// Calculate differences
	addPrincipals, removePrincipals, _ := flex.DiffSlices(currentPrincipals, wantPrincipals, func(a, b string) bool { return a == b })
	addResources, removeResources, _ := flex.DiffSlices(currentResources, wantResources, func(a, b string) bool { return a == b })

	// Remove principals no longer wanted
	for _, principal := range removePrincipals {
		if err := deleteResourceSharePrincipalAssociation(ctx, conn, resourceShareARN, principal); err != nil {
			diags.AddError(fmt.Sprintf("syncing RAM Resource Share (%s) Associations", resourceShareARN), err.Error())
			return diags
		}
	}

	// Remove resources no longer wanted
	for _, resourceARN := range removeResources {
		if err := deleteResourceShareResourceAssociation(ctx, conn, resourceShareARN, resourceARN); err != nil {
			diags.AddError(fmt.Sprintf("syncing RAM Resource Share (%s) Associations", resourceShareARN), err.Error())
			return diags
		}
	}

	// Add new principals
	for _, principal := range addPrincipals {
		if err := createResourceSharePrincipalAssociation(ctx, conn, resourceShareARN, principal, wantSources...); err != nil {
			diags.AddError(fmt.Sprintf("syncing RAM Resource Share (%s) Associations", resourceShareARN), err.Error())
			return diags
		}
	}

	// Add new resources
	for _, resourceARN := range addResources {
		if err := createResourceShareResourceAssociation(ctx, conn, resourceShareARN, resourceARN); err != nil {
			diags.AddError(fmt.Sprintf("syncing RAM Resource Share (%s) Associations", resourceShareARN), err.Error())
			return diags
		}
	}

	return diags
}

// findAssociationsForResourceShare retrieves all principal and resource associations
// for a resource share.
func findAssociationsForResourceShare(ctx context.Context, conn *ram.Client, resourceShareARN string) ([]string, []string, error) {
	var principals, resources []string

	// Fetch principal associations
	principalAssociations, err := findResourceShareAssociations(ctx, conn, &ram.GetResourceShareAssociationsInput{
		AssociationType:   awstypes.ResourceShareAssociationTypePrincipal,
		ResourceShareArns: []string{resourceShareARN},
	})
	if err != nil && !retry.NotFound(err) {
		return nil, nil, err
	}

	for _, v := range principalAssociations {
		if v.Status == awstypes.ResourceShareAssociationStatusAssociated {
			principals = append(principals, aws.ToString(v.AssociatedEntity))
		}
	}

	// Fetch resource associations
	resourceAssociations, err := findResourceShareAssociations(ctx, conn, &ram.GetResourceShareAssociationsInput{
		AssociationType:   awstypes.ResourceShareAssociationTypeResource,
		ResourceShareArns: []string{resourceShareARN},
	})
	if err != nil && !retry.NotFound(err) {
		return nil, nil, err
	}

	for _, v := range resourceAssociations {
		if v.Status == awstypes.ResourceShareAssociationStatusAssociated {
			resources = append(resources, aws.ToString(v.AssociatedEntity))
		}
	}

	return principals, resources, nil
}

type resourceShareAssociationExclusiveResourceModel struct {
	framework.WithRegionModel
	Principals       fwtypes.SetOfString `tfsdk:"principals"`
	ResourceARNs     fwtypes.SetOfString `tfsdk:"resource_arns"`
	ResourceShareARN fwtypes.ARN         `tfsdk:"resource_share_arn"`
	Sources          fwtypes.SetOfString `tfsdk:"sources"`
}
