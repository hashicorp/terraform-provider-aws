// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package auditmanager

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_auditmanager_organization_admin_account_registration", name="Organization Admin Account Registration")
// @Region(overrideEnabled=false)
func newOrganizationAdminAccountRegistrationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &organizationAdminAccountRegistrationResource{}, nil
}

const (
	ResNameOrganizationAdminAccountRegistration = "OrganizationAdminAccountRegistration"
)

type organizationAdminAccountRegistrationResource struct {
	framework.ResourceWithModel[organizationAdminAccountRegistrationResourceModel]
	framework.WithImportByID
}

func (r *organizationAdminAccountRegistrationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"admin_account_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"organization_id": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *organizationAdminAccountRegistrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().AuditManagerClient(ctx)

	var plan organizationAdminAccountRegistrationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := auditmanager.RegisterOrganizationAdminAccountInput{
		AdminAccountId: plan.AdminAccountID.ValueStringPointer(),
	}
	out, err := conn.RegisterOrganizationAdminAccount(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AuditManager, create.ErrActionCreating, ResNameOrganizationAdminAccountRegistration, plan.AdminAccountID.String(), nil),
			err.Error(),
		)
		return
	}

	state := plan
	state.AdminAccountID = flex.StringToFramework(ctx, out.AdminAccountId)
	state.ID = flex.StringToFramework(ctx, out.AdminAccountId)
	state.OrganizationID = flex.StringToFramework(ctx, out.OrganizationId)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *organizationAdminAccountRegistrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().AuditManagerClient(ctx)

	var state organizationAdminAccountRegistrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.GetOrganizationAdminAccount(ctx, &auditmanager.GetOrganizationAdminAccountInput{})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AuditManager, create.ErrActionReading, ResNameOrganizationAdminAccountRegistration, state.ID.String(), nil),
			err.Error(),
		)
		return
	}
	if out.AdminAccountId == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.AdminAccountID = flex.StringToFramework(ctx, out.AdminAccountId)
	state.ID = flex.StringToFramework(ctx, out.AdminAccountId)
	state.OrganizationID = flex.StringToFramework(ctx, out.OrganizationId)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is a no-op. Changing admin accounts requires the existing admin to
// be deregisterd first (destroy and replace).
func (r *organizationAdminAccountRegistrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *organizationAdminAccountRegistrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().AuditManagerClient(ctx)

	var state organizationAdminAccountRegistrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeregisterOrganizationAdminAccount(ctx, &auditmanager.DeregisterOrganizationAdminAccountInput{
		AdminAccountId: state.AdminAccountID.ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AuditManager, create.ErrActionDeleting, ResNameOrganizationAdminAccountRegistration, state.ID.String(), nil),
			err.Error(),
		)
	}
}

type organizationAdminAccountRegistrationResourceModel struct {
	AdminAccountID types.String `tfsdk:"admin_account_id"`
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
}
