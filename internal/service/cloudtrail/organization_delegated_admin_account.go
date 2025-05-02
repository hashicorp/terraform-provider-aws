// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudtrail

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudtrail/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cloudtrail_organization_delegated_admin_account", name="Organization Delegated Admin Account")
func newOrganizationDelegatedAdminAccountResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &organizationDelegatedAdminAccountResource{}, nil
}

type organizationDelegatedAdminAccountResource struct {
	framework.ResourceWithConfigure
	framework.WithNoUpdate
	framework.WithImportByID
}

func (r *organizationDelegatedAdminAccountResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAccountID: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					fwvalidators.AWSAccountID(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrARN: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrEmail: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service_principal": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *organizationDelegatedAdminAccountResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data organizationDelegatedAdminAccountResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudTrailClient(ctx)

	accountID := data.AccountID.ValueString()
	input := &cloudtrail.RegisterOrganizationDelegatedAdminInput{
		MemberAccountId: aws.String(accountID),
	}

	_, err := conn.RegisterOrganizationDelegatedAdmin(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("registering CloudTrail Organization Delegated Admin Account (%s)", accountID), err.Error())

		return
	}

	// Set values for unknowns.
	data.setID()

	delegatedAdministrator, err := tforganizations.FindDelegatedAdministratorByTwoPartKey(ctx, r.Meta().OrganizationsClient(ctx), accountID, servicePrincipal)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading CloudTrail Organization Delegated Admin Account (%s)", accountID), err.Error())

		return
	}

	data.ARN = fwflex.StringToFramework(ctx, delegatedAdministrator.Arn)
	data.Email = fwflex.StringToFramework(ctx, delegatedAdministrator.Email)
	data.Name = fwflex.StringToFramework(ctx, delegatedAdministrator.Name)
	data.ServicePrincipal = fwflex.StringValueToFramework(ctx, servicePrincipal)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *organizationDelegatedAdminAccountResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data organizationDelegatedAdminAccountResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().OrganizationsClient(ctx)

	delegatedAdministrator, err := tforganizations.FindDelegatedAdministratorByTwoPartKey(ctx, conn, data.ID.ValueString(), servicePrincipal)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading CloudTrail Organization Delegated Admin Account (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	data.ARN = fwflex.StringToFramework(ctx, delegatedAdministrator.Arn)
	data.Email = fwflex.StringToFramework(ctx, delegatedAdministrator.Email)
	data.Name = fwflex.StringToFramework(ctx, delegatedAdministrator.Name)
	data.ServicePrincipal = fwflex.StringValueToFramework(ctx, servicePrincipal)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *organizationDelegatedAdminAccountResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data organizationDelegatedAdminAccountResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudTrailClient(ctx)

	input := cloudtrail.DeregisterOrganizationDelegatedAdminInput{
		DelegatedAdminAccountId: data.ID.ValueStringPointer(),
	}
	_, err := conn.DeregisterOrganizationDelegatedAdmin(ctx, &input)

	if errs.IsA[*awstypes.AccountNotRegisteredException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deregistering CloudTrail Organization Delegated Admin Account (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

type organizationDelegatedAdminAccountResourceModel struct {
	AccountID        types.String `tfsdk:"account_id"`
	ARN              types.String `tfsdk:"arn"`
	Email            types.String `tfsdk:"email"`
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	ServicePrincipal types.String `tfsdk:"service_principal"`
}

func (model *organizationDelegatedAdminAccountResourceModel) InitFromID() error {
	model.AccountID = model.ID

	return nil
}

func (model *organizationDelegatedAdminAccountResourceModel) setID() {
	model.ID = model.AccountID
}
