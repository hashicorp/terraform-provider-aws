// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package auditmanager

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/auditmanager/types"
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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_auditmanager_organization_admin_account_registration", name="Organization Admin Account Registration")
func newOrganizationAdminAccountRegistrationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &organizationAdminAccountRegistrationResource{}, nil
}

type organizationAdminAccountRegistrationResource struct {
	framework.ResourceWithModel[organizationAdminAccountRegistrationResourceModel]
	framework.WithImportByID
	framework.WithNoUpdate
}

func (r *organizationAdminAccountRegistrationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"admin_account_id": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					fwvalidators.AWSAccountID(),
				},
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

func (r *organizationAdminAccountRegistrationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data organizationAdminAccountRegistrationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AuditManagerClient(ctx)

	adminAccountID := fwflex.StringValueFromFramework(ctx, data.AdminAccountID)
	input := auditmanager.RegisterOrganizationAdminAccountInput{
		AdminAccountId: aws.String(adminAccountID),
	}
	output, err := conn.RegisterOrganizationAdminAccount(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("registering Audit Manager Organization Admin Account (%s)", adminAccountID), err.Error())

		return
	}

	// Set values for unknowns.
	data.ID = fwflex.StringToFramework(ctx, output.AdminAccountId)
	data.OrganizationID = fwflex.StringToFramework(ctx, output.OrganizationId)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *organizationAdminAccountRegistrationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data organizationAdminAccountRegistrationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AuditManagerClient(ctx)

	output, err := conn.GetOrganizationAdminAccount(ctx, &auditmanager.GetOrganizationAdminAccountInput{})

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Audit Manager Organization Admin Account (%s)", data.ID.ValueString()), err.Error())

		return
	}

	data.AdminAccountID = fwflex.StringToFramework(ctx, output.AdminAccountId)
	data.OrganizationID = fwflex.StringToFramework(ctx, output.OrganizationId)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *organizationAdminAccountRegistrationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data organizationAdminAccountRegistrationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AuditManagerClient(ctx)

	input := auditmanager.DeregisterOrganizationAdminAccountInput{
		AdminAccountId: fwflex.StringFromFramework(ctx, data.ID),
	}
	_, err := conn.DeregisterOrganizationAdminAccount(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "Tenant is not in expected state") {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deregistering Audit Manager Organization Admin Account (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func findOrganizationAdminAccount(ctx context.Context, conn *auditmanager.Client) (*auditmanager.GetOrganizationAdminAccountOutput, error) {
	input := auditmanager.GetOrganizationAdminAccountInput{}
	output, err := conn.GetOrganizationAdminAccount(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AdminAccountId == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type organizationAdminAccountRegistrationResourceModel struct {
	framework.WithRegionModel
	AdminAccountID types.String `tfsdk:"admin_account_id"`
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
}
