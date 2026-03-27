// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package workmail

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workmail"
	awstypes "github.com/aws/aws-sdk-go-v2/service/workmail/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_workmail_user", name="User")
// @IdentityAttribute("organization_id")
// @IdentityAttribute("user_id")
// @ImportIDHandler("userImportID")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttributes="organization_id;user_id", importStateIdAttributesSep="flex.ResourceIdSeparator")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/workmail;workmail.DescribeUserOutput")
func newUserResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &userResource{}, nil
}

const (
	ResNameUser                = "User"
	userPropagationTimeout     = 2 * time.Minute
	userDeleteTransitionTimout = 2 * time.Minute
	userStateEnabled           = string(awstypes.EntityStateEnabled)
	userStateDisabled          = string(awstypes.EntityStateDisabled)
)

type userResource struct {
	framework.ResourceWithModel[userResourceModel]
	framework.WithImportByIdentity
}

func (r *userResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"city": schema.StringAttribute{
				Description: "City where the user is located.",
				Optional:    true,
			},
			"company": schema.StringAttribute{
				Description: "Company associated with the user.",
				Optional:    true,
			},
			"country": schema.StringAttribute{
				Description: "Country where the user is located.",
				Optional:    true,
			},
			"department": schema.StringAttribute{
				Description: "Department associated with the user.",
				Optional:    true,
			},
			"disabled_date": schema.StringAttribute{
				Description: "Timestamp when the user was disabled from WorkMail use.",
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
			},
			names.AttrDisplayName: schema.StringAttribute{
				Description: "Display name of the user.",
				Required:    true,
			},
			names.AttrEmail: schema.StringAttribute{
				Description: "Primary email address used to register the user with WorkMail.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled_date": schema.StringAttribute{
				Description: "Timestamp when the user was enabled for WorkMail use.",
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
			},
			"first_name": schema.StringAttribute{
				Description: "First name of the user.",
				Optional:    true,
			},
			"hidden_from_global_address_list": schema.BoolAttribute{
				Description: "Whether to hide the user from the global address list.",
				Optional:    true,
				Computed:    true,
			},
			"identity_provider_identity_store_id": schema.StringAttribute{
				Description: "Identity store ID from IAM Identity Center associated with the user.",
				Computed:    true,
			},
			"identity_provider_user_id": schema.StringAttribute{
				Description: "User ID from IAM Identity Center associated with the user.",
				Optional:    true,
			},
			"initials": schema.StringAttribute{
				Description: "Initials of the user.",
				Optional:    true,
			},
			"job_title": schema.StringAttribute{
				Description: "Job title of the user.",
				Optional:    true,
			},
			"last_name": schema.StringAttribute{
				Description: "Last name of the user.",
				Optional:    true,
			},
			"mailbox_deprovisioned_date": schema.StringAttribute{
				Description: "Timestamp when the mailbox was removed for the user.",
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
			},
			"mailbox_provisioned_date": schema.StringAttribute{
				Description: "Timestamp when the mailbox was created for the user.",
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
			},
			names.AttrName: schema.StringAttribute{
				Description: "Username of the user.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"office": schema.StringAttribute{
				Description: "Office where the user is located.",
				Optional:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "Identifier of the WorkMail organization where the user is managed.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrPassword: schema.StringAttribute{
				Description: "Password to set for the user.",
				Optional:    true,
				Sensitive:   true,
			},
			"user_role": schema.StringAttribute{
				Description: "Role assigned to the user.",
				Optional:    true,
				CustomType:  fwtypes.StringEnumType[awstypes.UserRole](),
				Computed:    true,
			},
			names.AttrState: schema.StringAttribute{
				Description: "Current WorkMail state of the user.",
				Computed:    true,
			},
			"street": schema.StringAttribute{
				Description: "Street address of the user.",
				Optional:    true,
			},
			"telephone": schema.StringAttribute{
				Description: "Telephone number of the user.",
				Optional:    true,
			},
			"user_id": schema.StringAttribute{
				Description: "Identifier of the user.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"zip_code": schema.StringAttribute{
				Description: "ZIP or postal code of the user.",
				Optional:    true,
			},
		},
	}
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().WorkMailClient(ctx)

	var plan userResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input workmail.CreateUserInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateUser(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil || out.UserId == nil {
		smerr.AddError(ctx, &resp.Diagnostics, fmt.Errorf("empty output"), smerr.ID, plan.Name.String())
		return
	}

	plan.UserId = flex.StringToFramework(ctx, out.UserId)

	if err := registerUser(ctx, conn, plan); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.UserId.String())
		return
	}

	created, err := waitUserEnabled(ctx, conn, plan.OrganizationId.ValueString(), plan.UserId.ValueString(), userPropagationTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.UserId.String())
		return
	}

	if hasPostCreateUpdate(plan) {
		if err := updateUser(ctx, conn, plan); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.UserId.String())
			return
		}

		created, err = findUserByTwoPartKey(ctx, conn, plan.OrganizationId.ValueString(), plan.UserId.ValueString())
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.UserId.String())
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, created, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().WorkMailClient(ctx)

	var state userResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findUserByTwoPartKey(ctx, conn, state.OrganizationId.ValueString(), state.UserId.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.UserId.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().WorkMailClient(ctx)

	var old, new userResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &new))
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &old))
	if resp.Diagnostics.HasError() {
		return
	}

	// update other optional fields
	if !new.City.Equal(old.City) ||
		!new.Company.Equal(old.Company) ||
		!new.Country.Equal(old.Country) ||
		!new.Department.Equal(old.Department) ||
		!new.DisplayName.Equal(old.DisplayName) ||
		!new.FirstName.Equal(old.FirstName) ||
		!new.HiddenFromGlobalAddressList.Equal(old.HiddenFromGlobalAddressList) ||
		!new.IdentityProviderUserId.Equal(old.IdentityProviderUserId) ||
		!new.Initials.Equal(old.Initials) ||
		!new.JobTitle.Equal(old.JobTitle) ||
		!new.LastName.Equal(old.LastName) ||
		!new.Office.Equal(old.Office) ||
		!new.UserRole.Equal(old.UserRole) ||
		!new.Street.Equal(old.Street) ||
		!new.Telephone.Equal(old.Telephone) ||
		!new.ZipCode.Equal(old.ZipCode) {
		if err := updateUser(ctx, conn, new); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, new.UserId.String())
			return
		}
	}

	// update password needs to call ResetPassword API
	if !new.Password.Equal(old.Password) {
		resetPasswordInput := workmail.ResetPasswordInput{
			OrganizationId: new.OrganizationId.ValueStringPointer(),
			Password:       new.Password.ValueStringPointer(),
			UserId:         new.UserId.ValueStringPointer(),
		}
		_, err := conn.ResetPassword(ctx, &resetPasswordInput)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, old.UserId.String())
			return
		}
	}

	out, err := findUserByTwoPartKey(ctx, conn, old.OrganizationId.ValueString(), old.UserId.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, old.UserId.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &new))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &new))
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().WorkMailClient(ctx)

	var state userResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := findUserByTwoPartKey(ctx, conn, state.OrganizationId.ValueString(), state.UserId.ValueString())
	if retry.NotFound(err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.UserId.String())
		return
	}

	if user.State == awstypes.EntityStateEnabled {
		if err := deregisterUser(ctx, conn, state); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.UserId.String())
			return
		}

		if _, err := waitUserDisabled(ctx, conn, state.OrganizationId.ValueString(), state.UserId.ValueString(), userDeleteTransitionTimout); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.UserId.String())
			return
		}
	}

	_, err = tfresource.RetryWhenIsA[any, *awstypes.EntityStateException](ctx, userDeleteTransitionTimout, func(ctx context.Context) (any, error) {
		deleteUserInput := workmail.DeleteUserInput{
			OrganizationId: state.OrganizationId.ValueStringPointer(),
			UserId:         state.UserId.ValueStringPointer(),
		}
		_, err := conn.DeleteUser(ctx, &deleteUserInput)

		return nil, err
	})
	if err != nil && !errs.IsA[*awstypes.EntityNotFoundException](err) && !errs.IsA[*awstypes.ResourceNotFoundException](err) {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.UserId.String())
		return
	}

	if _, err := waitUserDeleted(ctx, conn, state.OrganizationId.ValueString(), state.UserId.ValueString(), userDeleteTransitionTimout); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.UserId.String())
	}
}

func (r *userResource) flatten(ctx context.Context, out *workmail.DescribeUserOutput, data *userResourceModel) (diags diag.Diagnostics) {
	diags.Append(flex.Flatten(ctx, out, data)...)
	return diags
}

func updateUser(ctx context.Context, conn *workmail.Client, data userResourceModel) error {
	var input workmail.UpdateUserInput
	if diags := flex.Expand(ctx, data, &input); diags.HasError() {
		return fmt.Errorf("expanding workmail user update input: %s", diags.Errors()[0].Detail())
	}
	_, err := conn.UpdateUser(ctx, &input)

	return err
}

func registerUser(ctx context.Context, conn *workmail.Client, data userResourceModel) error {
	err := tfresource.Retry(ctx, userPropagationTimeout, func(ctx context.Context) *tfresource.RetryError {
		input := workmail.RegisterToWorkMailInput{
			Email:          data.Email.ValueStringPointer(),
			EntityId:       data.UserId.ValueStringPointer(),
			OrganizationId: data.OrganizationId.ValueStringPointer(),
		}
		_, err := conn.RegisterToWorkMail(ctx, &input)

		if errs.IsA[*awstypes.MailDomainStateException](err) || errs.IsA[*awstypes.EntityNotFoundException](err) || errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return tfresource.RetryableError(err)
		}

		if err != nil {
			return tfresource.NonRetryableError(err)
		}

		return nil
	})

	return err
}

func deregisterUser(ctx context.Context, conn *workmail.Client, data userResourceModel) error {
	_, err := tfresource.RetryWhenIsA[any, *awstypes.EntityStateException](ctx, userDeleteTransitionTimout, func(ctx context.Context) (any, error) {
		input := workmail.DeregisterFromWorkMailInput{
			EntityId:       data.UserId.ValueStringPointer(),
			OrganizationId: data.OrganizationId.ValueStringPointer(),
		}
		_, err := conn.DeregisterFromWorkMail(ctx, &input)

		return nil, err
	})

	if errs.IsA[*awstypes.EntityNotFoundException](err) || errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil
	}

	return err
}

func hasPostCreateUpdate(data userResourceModel) bool {
	return isStringSet(data.City) ||
		isStringSet(data.Company) ||
		isStringSet(data.Country) ||
		isStringSet(data.Department) ||
		isStringSet(data.Initials) ||
		isStringSet(data.JobTitle) ||
		isStringSet(data.Office) ||
		isStringSet(data.Street) ||
		isStringSet(data.Telephone) ||
		isStringSet(data.ZipCode)
}

func waitUserEnabled(ctx context.Context, conn *workmail.Client, organizationID, userID string, timeout time.Duration) (*workmail.DescribeUserOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{userStateDisabled},
		Target:                    []string{userStateEnabled},
		Refresh:                   statusUser(conn, organizationID, userID),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*workmail.DescribeUserOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitUserDisabled(ctx context.Context, conn *workmail.Client, organizationID, userID string, timeout time.Duration) (*workmail.DescribeUserOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{userStateEnabled},
		Target:                    []string{userStateDisabled},
		Refresh:                   statusUser(conn, organizationID, userID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*workmail.DescribeUserOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitUserDeleted(ctx context.Context, conn *workmail.Client, organizationID, userID string, timeout time.Duration) (*workmail.DescribeUserOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{userStateDisabled, userStateEnabled},
		Target:  []string{},
		Refresh: statusUser(conn, organizationID, userID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*workmail.DescribeUserOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusUser(conn *workmail.Client, organizationID, userID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findUserByTwoPartKey(ctx, conn, organizationID, userID)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.State), nil
	}
}

func findUserByTwoPartKey(ctx context.Context, conn *workmail.Client, organizationID, userID string) (*workmail.DescribeUserOutput, error) {
	input := workmail.DescribeUserInput{
		OrganizationId: aws.String(organizationID),
		UserId:         aws.String(userID),
	}

	out, err := conn.DescribeUser(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFoundException](err) || errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsA[*awstypes.OrganizationStateException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	if out.State == awstypes.EntityStateDeleted {
		return nil, smarterr.NewError(&retry.NotFoundError{Message: fmt.Sprintf("WorkMail User %s is in Deleted state", userID)})
	}

	return out, nil
}

func isStringSet(value types.String) bool {
	return !value.IsNull() && !value.IsUnknown()
}

var (
	_ inttypes.ImportIDParser = userImportID{}
)

type userImportID struct{}

func (userImportID) Parse(id string) (string, map[string]any, error) {
	organizationID, userID, found := strings.Cut(id, intflex.ResourceIdSeparator)
	if !found {
		return "", nil, fmt.Errorf("id %q should be in the format <organization-id>%s<user-id>", id, intflex.ResourceIdSeparator)
	}

	result := map[string]any{
		"organization_id": organizationID,
		"user_id":         userID,
	}

	return id, result, nil
}

type userResourceModel struct {
	framework.WithRegionModel
	City                            types.String                          `tfsdk:"city"`
	Company                         types.String                          `tfsdk:"company"`
	Country                         types.String                          `tfsdk:"country"`
	Department                      types.String                          `tfsdk:"department"`
	DisabledDate                    timetypes.RFC3339                     `tfsdk:"disabled_date"`
	DisplayName                     types.String                          `tfsdk:"display_name"`
	Email                           types.String                          `tfsdk:"email"`
	EnabledDate                     timetypes.RFC3339                     `tfsdk:"enabled_date"`
	FirstName                       types.String                          `tfsdk:"first_name"`
	HiddenFromGlobalAddressList     types.Bool                            `tfsdk:"hidden_from_global_address_list"`
	IdentityProviderIdentityStoreId types.String                          `tfsdk:"identity_provider_identity_store_id"`
	IdentityProviderUserId          types.String                          `tfsdk:"identity_provider_user_id"`
	Initials                        types.String                          `tfsdk:"initials"`
	JobTitle                        types.String                          `tfsdk:"job_title"`
	LastName                        types.String                          `tfsdk:"last_name"`
	MailboxDeprovisionedDate        timetypes.RFC3339                     `tfsdk:"mailbox_deprovisioned_date"`
	MailboxProvisionedDate          timetypes.RFC3339                     `tfsdk:"mailbox_provisioned_date"`
	Name                            types.String                          `tfsdk:"name"`
	Office                          types.String                          `tfsdk:"office"`
	OrganizationId                  types.String                          `tfsdk:"organization_id"`
	Password                        types.String                          `tfsdk:"password"`
	UserRole                        fwtypes.StringEnum[awstypes.UserRole] `tfsdk:"user_role"`
	State                           types.String                          `tfsdk:"state"`
	Street                          types.String                          `tfsdk:"street"`
	Telephone                       types.String                          `tfsdk:"telephone"`
	UserId                          types.String                          `tfsdk:"user_id"`
	ZipCode                         types.String                          `tfsdk:"zip_code"`
}
