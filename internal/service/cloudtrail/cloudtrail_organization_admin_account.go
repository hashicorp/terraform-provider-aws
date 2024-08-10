// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudtrail

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudtrail/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameCloudTrailOrganizationAdminAccount = "CloudTrailOrganizationAdminAccount"
)

// @FrameworkResource("aws_cloudtrail_organization_admin_account", name="CloudTrail Organization Admin Account")
func newResourceCloudTrailOrganizationAdminAccount(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceCloudTrailOrganizationAdminAccount{}, nil
}

type resourceCloudTrailOrganizationAdminAccount struct {
	framework.ResourceWithConfigure
	framework.WithNoUpdate
}

func (r *resourceCloudTrailOrganizationAdminAccount) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_cloudtrail_organization_admin_account"
}

func (r *resourceCloudTrailOrganizationAdminAccount) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
			},
			"delegated_admin_account_id": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					fwvalidators.AWSAccountID(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrEmail: schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Computed: true,
			},
			"service_principal": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *resourceCloudTrailOrganizationAdminAccount) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	cloudTrailConn := r.Meta().CloudTrailClient(ctx)
	organizationsConn := r.Meta().OrganizationsClient(ctx)

	var plan resourceCloudTrailOrganizationAdminAccountData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &cloudtrail.RegisterOrganizationDelegatedAdminInput{
		MemberAccountId: aws.String(plan.DelegatedAdminAccountID.ValueString()),
	}

	_, err := cloudTrailConn.RegisterOrganizationDelegatedAdmin(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudTrail, create.ErrActionCreating, ResNameCloudTrailOrganizationAdminAccount, plan.DelegatedAdminAccountID.String(), nil),
			err.Error(),
		)
		return
	}

	// Read after create to get computed attributes
	readOutput, err := FindDelegatedAccountByAccountID(ctx, organizationsConn, plan.DelegatedAdminAccountID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudTrail, create.ErrActionCreating, ResNameCloudTrailOrganizationAdminAccount, plan.DelegatedAdminAccountID.String(), err),
			err.Error(),
		)
		return
	}

	state := plan
	state.ServicePrincipal = flex.StringToFramework(ctx, aws.String(cloudTrailServicePrincipal))

	resp.Diagnostics.Append(flex.Flatten(ctx, readOutput, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceCloudTrailOrganizationAdminAccount) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().OrganizationsClient(ctx)
	var state resourceCloudTrailOrganizationAdminAccountData

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindDelegatedAccountByAccountID(ctx, conn, state.DelegatedAdminAccountID.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudTrail, create.ErrActionSetting, ResNameCloudTrailOrganizationAdminAccount, state.DelegatedAdminAccountID.String(), err),
			err.Error(),
		)
		return
	}

	state.ServicePrincipal = flex.StringToFramework(ctx, aws.String(cloudTrailServicePrincipal))

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceCloudTrailOrganizationAdminAccount) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().CloudTrailClient(ctx)

	var state resourceCloudTrailOrganizationAdminAccountData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeregisterOrganizationDelegatedAdmin(ctx, &cloudtrail.DeregisterOrganizationDelegatedAdminInput{
		DelegatedAdminAccountId: aws.String(state.DelegatedAdminAccountID.ValueString()),
	})
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudTrail, create.ErrActionDeleting, ResNameCloudTrailOrganizationAdminAccount, state.DelegatedAdminAccountID.String(), nil),
			err.Error(),
		)
	}
}

func (r *resourceCloudTrailOrganizationAdminAccount) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("delegated_admin_account_id"), request, response)
}

type resourceCloudTrailOrganizationAdminAccountData struct {
	Arn                     types.String `tfsdk:"arn"`
	DelegatedAdminAccountID types.String `tfsdk:"delegated_admin_account_id"`
	Email                   types.String `tfsdk:"email"`
	Name                    types.String `tfsdk:"name"`
	ServicePrincipal        types.String `tfsdk:"service_principal"`
}
