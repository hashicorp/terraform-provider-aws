// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"slices"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"

	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_iam_organization_features", name="Organization Features")
func newResourceOrganizationFeatures(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceOrganizationFeatures{}
	return r, nil
}

const (
	ResNameOrganizationFeatures = "IAM Organization Features"
)

type resourceOrganizationFeatures struct {
	framework.ResourceWithConfigure
}

func (r *resourceOrganizationFeatures) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_iam_organization_features"
}

func (r *resourceOrganizationFeatures) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: schema.StringAttribute{
				Computed: true,
			},
			"features": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						enum.FrameworkValidate[awstypes.FeatureType](),
					),
				},
			},
		},
	}
}

func (r *resourceOrganizationFeatures) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().IAMClient(ctx)

	var plan resourceOrganizationFeaturesModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var planFeatures []string
	resp.Diagnostics.Append(plan.EnabledFeatures.ElementsAs(ctx, &planFeatures, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := manageOrganizationFeatures(ctx, conn, planFeatures, []string{})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IAM, create.ErrActionCreating, ResNameOrganizationFeatures, plan.OrganizationId.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceOrganizationFeatures) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().IAMClient(ctx)

	var state resourceOrganizationFeaturesModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input iam.ListOrganizationsFeaturesInput
	out, err := conn.ListOrganizationsFeatures(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IAM, create.ErrActionReading, ResNameOrganizationFeatures, "", err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceOrganizationFeatures) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().IAMClient(ctx)

	var plan, state resourceOrganizationFeaturesModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var stateFeatures, planFeatures []string
	resp.Diagnostics.Append(plan.EnabledFeatures.ElementsAs(ctx, &planFeatures, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(state.EnabledFeatures.ElementsAs(ctx, &stateFeatures, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := manageOrganizationFeatures(ctx, conn, planFeatures, stateFeatures)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IAM, create.ErrActionCreating, ResNameOrganizationFeatures, plan.OrganizationId.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceOrganizationFeatures) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().IAMClient(ctx)

	var state resourceOrganizationFeaturesModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var stateFeatures []string
	resp.Diagnostics.Append(state.EnabledFeatures.ElementsAs(ctx, &stateFeatures, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := manageOrganizationFeatures(ctx, conn, []string{}, stateFeatures)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IAM, create.ErrActionDeleting, ResNameOrganizationFeatures, state.OrganizationId.String(), err),
			err.Error(),
		)
		return
	}

}

func (r *resourceOrganizationFeatures) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

type resourceOrganizationFeaturesModel struct {
	OrganizationId  types.String `tfsdk:"id"`
	EnabledFeatures types.Set    `tfsdk:"features"`
}

func manageOrganizationFeatures(ctx context.Context, conn *iam.Client, planFeatures, stateFeatures []string) (*iam.ListOrganizationsFeaturesOutput, error) {

	var featuresToEnable, featuresToDisable []string
	for _, feature := range planFeatures {
		if !slices.Contains(stateFeatures, feature) {
			featuresToEnable = append(featuresToEnable, feature)
		}
	}
	for _, feature := range stateFeatures {
		if !slices.Contains(planFeatures, feature) {
			featuresToDisable = append(featuresToDisable, feature)
		}
	}

	if slices.Contains(featuresToEnable, string(awstypes.FeatureTypeRootCredentialsManagement)) {
		var input iam.EnableOrganizationsRootCredentialsManagementInput
		_, err := conn.EnableOrganizationsRootCredentialsManagement(ctx, &input)
		if err != nil {
			return nil, err
		}
	}
	if slices.Contains(featuresToEnable, string(awstypes.FeatureTypeRootSessions)) {
		var input iam.EnableOrganizationsRootSessionsInput
		_, err := conn.EnableOrganizationsRootSessions(ctx, &input)
		if err != nil {
			return nil, err
		}
	}
	if slices.Contains(featuresToDisable, string(awstypes.FeatureTypeRootCredentialsManagement)) {
		var input iam.DisableOrganizationsRootCredentialsManagementInput
		_, err := conn.DisableOrganizationsRootCredentialsManagement(ctx, &input)
		if err != nil {
			return nil, err
		}
	}
	if slices.Contains(featuresToDisable, string(awstypes.FeatureTypeRootSessions)) {
		var input iam.DisableOrganizationsRootSessionsInput
		_, err := conn.DisableOrganizationsRootSessions(ctx, &input)
		if err != nil {
			return nil, err
		}
	}
	var input iam.ListOrganizationsFeaturesInput
	out, err := conn.ListOrganizationsFeatures(ctx, &input)
	if err != nil {
		return nil, err
	}
	return out, nil
}
