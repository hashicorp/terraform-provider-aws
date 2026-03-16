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
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
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
func newUserResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &userResource{}, nil
}

const (
	ResNameUser                = "User"
	userPropagationTimeout     = 2 * time.Minute
	userDeleteTransitionTimout = 2 * time.Minute
)

type userResource struct {
	framework.ResourceWithModel[userResourceModel]
	framework.WithImportByIdentity
}

func (r *userResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"city": schema.StringAttribute{
				Optional: true,
			},
			"company": schema.StringAttribute{
				Optional: true,
			},
			"country": schema.StringAttribute{
				Optional: true,
			},
			"department": schema.StringAttribute{
				Optional: true,
			},
			"disabled_date": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrDisplayName: schema.StringAttribute{
				Required: true,
			},
			names.AttrEmail: schema.StringAttribute{
				Computed: true,
			},
			"enabled_date": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"first_name": schema.StringAttribute{
				Optional: true,
			},
			"hidden_from_global_address_list": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"identity_provider_identity_store_id": schema.StringAttribute{
				Computed: true,
			},
			"identity_provider_user_id": schema.StringAttribute{
				Optional: true,
			},
			"initials": schema.StringAttribute{
				Optional: true,
			},
			"job_title": schema.StringAttribute{
				Optional: true,
			},
			"last_name": schema.StringAttribute{
				Optional: true,
			},
			"mailbox_deprovisioned_date": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"mailbox_provisioned_date": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"office": schema.StringAttribute{
				Optional: true,
			},
			"organization_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(string(awstypes.UserRoleUser)),
				Validators: []validator.String{
					stringvalidator.OneOf(enumUserRoleValues()...),
				},
			},
			names.AttrState: schema.StringAttribute{
				Computed: true,
			},
			"street": schema.StringAttribute{
				Optional: true,
			},
			"telephone": schema.StringAttribute{
				Optional: true,
			},
			"user_id": framework.IDAttribute(),
			"zip_code": schema.StringAttribute{
				Optional: true,
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

	plan.UserId = types.StringPointerValue(out.UserId)

	if hasPostCreateUpdate(plan) {
		if err := updateUser(ctx, conn, plan); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.UserId.String())
			return
		}
	}

	created, err := tfresource.RetryWhenNotFound(ctx, userPropagationTimeout, func(ctx context.Context) (*workmail.DescribeUserOutput, error) {
		return findUserByTwoPartKey(ctx, conn, plan.OrganizationId.ValueString(), plan.UserId.ValueString())
	})
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.UserId.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, plan.OrganizationId.ValueString(), created, &plan))
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

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, state.OrganizationId.ValueString(), out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().WorkMailClient(ctx)

	var plan userResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	if err := updateUser(ctx, conn, plan); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.UserId.String())
		return
	}

	out, err := findUserByTwoPartKey(ctx, conn, plan.OrganizationId.ValueString(), plan.UserId.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.UserId.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, plan.OrganizationId.ValueString(), out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
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
		_, err := conn.DeregisterFromWorkMail(ctx, &workmail.DeregisterFromWorkMailInput{
			EntityId:       state.UserId.ValueStringPointer(),
			OrganizationId: state.OrganizationId.ValueStringPointer(),
		})
		if err != nil && !errs.IsA[*awstypes.EntityNotFoundException](err) && !errs.IsA[*awstypes.ResourceNotFoundException](err) {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.UserId.String())
			return
		}
	}

	_, err = tfresource.RetryWhenIsA[any, *awstypes.EntityStateException](ctx, userDeleteTransitionTimout, func(ctx context.Context) (any, error) {
		_, err := conn.DeleteUser(ctx, &workmail.DeleteUserInput{
			OrganizationId: state.OrganizationId.ValueStringPointer(),
			UserId:         state.UserId.ValueStringPointer(),
		})

		return nil, err
	})
	if err != nil && !errs.IsA[*awstypes.EntityNotFoundException](err) && !errs.IsA[*awstypes.ResourceNotFoundException](err) {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.UserId.String())
	}
}

func (r *userResource) flatten(ctx context.Context, organizationID string, out *workmail.DescribeUserOutput, data *userResourceModel) (diags diag.Diagnostics) {
	diags.Append(flex.Flatten(ctx, out, data)...)
	data.OrganizationId = types.StringValue(organizationID)
	data.Role = stringEnumFrameworkValue(string(out.UserRole))

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

	return out, nil
}

func isStringSet(value types.String) bool {
	return !value.IsNull() && !value.IsUnknown()
}

func stringEnumFrameworkValue(value string) types.String {
	if value == "" {
		return types.StringNull()
	}

	return types.StringValue(value)
}

func enumUserRoleValues() []string {
	values := awstypes.UserRole("").Values()
	result := make([]string, 0, len(values))

	for _, value := range values {
		result = append(result, string(value))
	}

	return result
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
	City                            types.String      `tfsdk:"city"`
	Company                         types.String      `tfsdk:"company"`
	Country                         types.String      `tfsdk:"country"`
	Department                      types.String      `tfsdk:"department"`
	DisabledDate                    timetypes.RFC3339 `tfsdk:"disabled_date"`
	DisplayName                     types.String      `tfsdk:"display_name"`
	Email                           types.String      `tfsdk:"email"`
	EnabledDate                     timetypes.RFC3339 `tfsdk:"enabled_date"`
	FirstName                       types.String      `tfsdk:"first_name"`
	HiddenFromGlobalAddressList     types.Bool        `tfsdk:"hidden_from_global_address_list"`
	IdentityProviderIdentityStoreId types.String      `tfsdk:"identity_provider_identity_store_id"`
	IdentityProviderUserId          types.String      `tfsdk:"identity_provider_user_id"`
	Initials                        types.String      `tfsdk:"initials"`
	JobTitle                        types.String      `tfsdk:"job_title"`
	LastName                        types.String      `tfsdk:"last_name"`
	MailboxDeprovisionedDate        timetypes.RFC3339 `tfsdk:"mailbox_deprovisioned_date"`
	MailboxProvisionedDate          timetypes.RFC3339 `tfsdk:"mailbox_provisioned_date"`
	Name                            types.String      `tfsdk:"name"`
	Office                          types.String      `tfsdk:"office"`
	OrganizationId                  types.String      `tfsdk:"organization_id"`
	Password                        types.String      `tfsdk:"password"`
	Role                            types.String      `tfsdk:"role"`
	State                           types.String      `tfsdk:"state"`
	Street                          types.String      `tfsdk:"street"`
	Telephone                       types.String      `tfsdk:"telephone"`
	UserId                          types.String      `tfsdk:"user_id"`
	ZipCode                         types.String      `tfsdk:"zip_code"`
}
