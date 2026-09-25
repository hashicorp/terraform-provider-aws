// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networksecuritymanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networksecuritymanager/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_networksecuritymanager_admin_account", name="Admin Account")
// @IdentityAttribute("admin_account_id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(serialize=true)
// @Testing(generator=false)
// @Testing(preCheck="testAccPreCheck")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/networksecuritymanager/types;awstypes;awstypes.AdminAccountDetails")
// @Testing(importStateIdAttribute="admin_account_id")
// @Testing(identityRegionOverrideTest=false)
func newAdminAccountResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &adminAccountResource{}

	// Setting an administrator account again shortly after removing it, or
	// while the service creates its service-linked role, is rejected until
	// the previous removal has settled ("retry the request after a few
	// minutes" per the API reference).
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(15 * time.Minute)
	r.SetDefaultDeleteTimeout(15 * time.Minute)

	return r, nil
}

type adminAccountResource struct {
	framework.ResourceWithModel[adminAccountResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *adminAccountResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"admin_account_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					fwvalidators.AWSAccountID(),
				},
			},
			names.AttrPriority: schema.Int32Attribute{
				Required: true,
				Validators: []validator.Int32{
					int32validator.Between(1, 10),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AdminAccountStatus](),
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"admin_scope": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[adminScopeModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"firewall_type_scope": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[firewallTypeScopeModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"all_firewall_types_enabled": schema.BoolAttribute{
										Optional: true,
									},
									"firewall_types": schema.SetAttribute{
										CustomType:  fwtypes.SetOfStringEnumType[awstypes.PolicyFirewallType](),
										ElementType: types.StringType,
										Optional:    true,
									},
								},
							},
						},
						"scope_filter": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[scopeFilterModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"include_all": schema.BoolAttribute{
										Optional: true,
										Validators: []validator.Bool{
											boolvalidator.Equals(true),
											boolvalidator.ExactlyOneOf(
												path.MatchRelative().AtParent().AtName("exclude_only"),
												path.MatchRelative().AtParent().AtName("include_all"),
												path.MatchRelative().AtParent().AtName("include_only"),
											),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"exclude_only": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[scopeSelectionModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
											listvalidator.ExactlyOneOf(
												path.MatchRelative().AtParent().AtName("exclude_only"),
												path.MatchRelative().AtParent().AtName("include_all"),
												path.MatchRelative().AtParent().AtName("include_only"),
											),
										},
										NestedObject: scopeSelectionNestedBlockObject(ctx),
									},
									"include_only": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[scopeSelectionModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
											listvalidator.ExactlyOneOf(
												path.MatchRelative().AtParent().AtName("exclude_only"),
												path.MatchRelative().AtParent().AtName("include_all"),
												path.MatchRelative().AtParent().AtName("include_only"),
											),
										},
										NestedObject: scopeSelectionNestedBlockObject(ctx),
									},
								},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func scopeSelectionNestedBlockObject(_ context.Context) schema.NestedBlockObject {
	return schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"accounts": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Optional:    true,
			},
			"organizational_units": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Optional:    true,
			},
		},
	}
}

func (r *adminAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	var plan adminAccountResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	adminAccountID := plan.AdminAccountID.ValueString()
	input, diags := expandPutAdminAccountInput(ctx, &plan)
	smerr.AddEnrich(ctx, &resp.Diagnostics, diags)
	if resp.Diagnostics.HasError() {
		return
	}

	// PutAdminAccount is rejected with a ConflictException immediately after
	// the same account was removed, or while the service-linked role is being
	// created. Retry until the service accepts the request.
	_, err := tfresource.RetryWhenIsA[any, *awstypes.ConflictException](ctx, r.CreateTimeout(ctx, plan.Timeouts), func(ctx context.Context) (any, error) {
		return conn.PutAdminAccount(ctx, input)
	})
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, adminAccountID)
		return
	}

	output, err := waitAdminAccountOnboarded(ctx, conn, adminAccountID, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, adminAccountID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flattenAdminAccount(ctx, output, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *adminAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	var state adminAccountResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	adminAccountID := state.AdminAccountID.ValueString()
	output, err := findAdminAccountByID(ctx, conn, adminAccountID)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, adminAccountID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flattenAdminAccount(ctx, output, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *adminAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	var plan, state adminAccountResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	adminAccountID := plan.AdminAccountID.ValueString()

	// There is no separate update operation: PutAdminAccount on an existing
	// administrator account replaces its priority and administrative scope.
	if !plan.Priority.Equal(state.Priority) || !plan.AdminScope.Equal(state.AdminScope) {
		input, diags := expandPutAdminAccountInput(ctx, &plan)
		smerr.AddEnrich(ctx, &resp.Diagnostics, diags)
		if resp.Diagnostics.HasError() {
			return
		}

		_, err := tfresource.RetryWhenIsA[any, *awstypes.ConflictException](ctx, r.UpdateTimeout(ctx, plan.Timeouts), func(ctx context.Context) (any, error) {
			return conn.PutAdminAccount(ctx, input)
		})
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, adminAccountID)
			return
		}
	}

	output, err := waitAdminAccountOnboarded(ctx, conn, adminAccountID, r.UpdateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, adminAccountID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flattenAdminAccount(ctx, output, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *adminAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	var state adminAccountResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	adminAccountID := state.AdminAccountID.ValueString()
	input := networksecuritymanager.DeleteAdminAccountInput{
		AccountId: aws.String(adminAccountID),
	}

	// Immediately after the administrator account was set, DeleteAdminAccount
	// can be rejected with an AccessDeniedException ("Unable to determine
	// service/operation name to be authorized") until the service has
	// finished onboarding it.
	_, err := tfresource.RetryWhenIsAErrorMessageContains[any, *awstypes.AccessDeniedException](ctx, r.DeleteTimeout(ctx, state.Timeouts), func(ctx context.Context) (any, error) {
		return conn.DeleteAdminAccount(ctx, &input)
	}, "Unable to determine service/operation name")
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, adminAccountID)
		return
	}

	if _, err := waitAdminAccountDeleted(ctx, conn, adminAccountID, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, adminAccountID)
		return
	}
}

func expandPutAdminAccountInput(ctx context.Context, data *adminAccountResourceModel) (*networksecuritymanager.PutAdminAccountInput, diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	var diags diag.Diagnostics

	var input networksecuritymanager.PutAdminAccountInput
	smerr.AddEnrich(ctx, &diags, flex.Expand(ctx, data, &input))
	if diags.HasError() {
		return nil, diags
	}

	// The account ID argument is named after the API's response field
	// (AdminAccount) rather than its request field (AccountId).
	input.AccountId = data.AdminAccountID.ValueStringPointer()

	return &input, diags
}

func flattenAdminAccount(ctx context.Context, output *awstypes.AdminAccountDetails, data *adminAccountResourceModel) diag.Diagnostics { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	// The response does not always echo AdminAccount; the argument is kept
	// from configuration or import.
	return flex.Flatten(ctx, output, data)
}

func findAdminAccountByID(ctx context.Context, conn *networksecuritymanager.Client, adminAccountID string) (*awstypes.AdminAccountDetails, error) {
	// GetAdminAccount is only permitted for accounts that are currently
	// administrators; for a removed account it fails with an
	// AccessDeniedException rather than a ResourceNotFoundException.
	// ListAdminAccounts answers for the whole organization, so existence is
	// established from the list first.
	if _, err := findAdminAccountSummaryByID(ctx, conn, adminAccountID); err != nil {
		return nil, err
	}

	input := networksecuritymanager.GetAdminAccountInput{
		AccountId: aws.String(adminAccountID),
	}

	output, err := conn.GetAdminAccount(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if output == nil || output.AdminAccountDetails == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	// A removed administrator account may remain visible as OFFBOARDED for a
	// while; it no longer administers anything and is treated as gone.
	if output.AdminAccountDetails.Status == awstypes.AdminAccountStatusOffboarded {
		return nil, smarterr.NewError(&retry.NotFoundError{
			Message: fmt.Sprintf("administrator account %s is %s", adminAccountID, output.AdminAccountDetails.Status),
		})
	}

	return output.AdminAccountDetails, nil
}

func findAdminAccountSummaryByID(ctx context.Context, conn *networksecuritymanager.Client, adminAccountID string) (*awstypes.AdminAccountSummary, error) {
	input := networksecuritymanager.ListAdminAccountsInput{}

	pages := networksecuritymanager.NewListAdminAccountsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.AdminAccounts {
			if aws.ToString(v.AccountId) == adminAccountID {
				return &v, nil
			}
		}
	}

	return nil, smarterr.NewError(&retry.NotFoundError{
		Message: fmt.Sprintf("administrator account %s not found", adminAccountID),
	})
}

func statusAdminAccount(conn *networksecuritymanager.Client, adminAccountID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findAdminAccountByID(ctx, conn, adminAccountID)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return output, string(output.Status), nil
	}
}

func waitAdminAccountOnboarded(ctx context.Context, conn *networksecuritymanager.Client, adminAccountID string, timeout time.Duration) (*awstypes.AdminAccountDetails, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  enum.Slice(awstypes.AdminAccountStatusOnboarded),
		Refresh: statusAdminAccount(conn, adminAccountID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.AdminAccountDetails); ok {
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitAdminAccountDeleted(ctx context.Context, conn *networksecuritymanager.Client, adminAccountID string, timeout time.Duration) (*awstypes.AdminAccountDetails, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AdminAccountStatusOnboarded),
		Target:  []string{},
		Refresh: statusAdminAccount(conn, adminAccountID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.AdminAccountDetails); ok {
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

type adminAccountResourceModel struct {
	framework.WithRegionModel
	AdminAccountID types.String                                     `tfsdk:"admin_account_id"`
	AdminScope     fwtypes.ListNestedObjectValueOf[adminScopeModel] `tfsdk:"admin_scope"`
	Priority       types.Int32                                      `tfsdk:"priority"`
	Status         fwtypes.StringEnum[awstypes.AdminAccountStatus]  `tfsdk:"status"`
	Timeouts       timeouts.Value                                   `tfsdk:"timeouts"`
}

type adminScopeModel struct {
	FirewallTypeScope fwtypes.ListNestedObjectValueOf[firewallTypeScopeModel] `tfsdk:"firewall_type_scope"`
	ScopeFilter       fwtypes.ListNestedObjectValueOf[scopeFilterModel]       `tfsdk:"scope_filter"`
}

type firewallTypeScopeModel struct {
	AllFirewallTypesEnabled types.Bool                                           `tfsdk:"all_firewall_types_enabled"`
	FirewallTypes           fwtypes.SetOfStringEnum[awstypes.PolicyFirewallType] `tfsdk:"firewall_types"`
}

// scopeFilterModel models the AdminScopeFilter tagged union: exactly one of
// include_all, include_only or exclude_only is set.
type scopeFilterModel struct {
	ExcludeOnly fwtypes.ListNestedObjectValueOf[scopeSelectionModel] `tfsdk:"exclude_only"`
	IncludeAll  types.Bool                                           `tfsdk:"include_all"`
	IncludeOnly fwtypes.ListNestedObjectValueOf[scopeSelectionModel] `tfsdk:"include_only"`
}

type scopeSelectionModel struct {
	Accounts            fwtypes.SetOfString `tfsdk:"accounts"`
	OrganizationalUnits fwtypes.SetOfString `tfsdk:"organizational_units"`
}

var (
	_ flex.Flattener = &firewallTypeScopeModel{}
	_ flex.Expander  = scopeFilterModel{}
	_ flex.Flattener = &scopeFilterModel{}
)

// Flatten normalizes the API's empty FirewallTypes list (returned when only
// AllFirewallTypesEnabled was set) to null, matching an omitted argument.
func (m *firewallTypeScopeModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	var apiObject *awstypes.AdminFirewallTypeScope
	switch t := v.(type) {
	case awstypes.AdminFirewallTypeScope:
		apiObject = &t
	case *awstypes.AdminFirewallTypeScope:
		apiObject = t
	default:
		diags.AddError(
			"Unsupported Firewall Type Scope",
			fmt.Sprintf("firewallTypeScopeModel.Flatten: %T", v),
		)
		return diags
	}

	m.AllFirewallTypesEnabled = flex.BoolToFramework(ctx, apiObject.AllFirewallTypesEnabled)
	m.FirewallTypes = fwtypes.NewSetValueOfNull[fwtypes.StringEnum[awstypes.PolicyFirewallType]](ctx)
	if len(apiObject.FirewallTypes) > 0 {
		m.FirewallTypes = flex.FlattenFrameworkStringyValueSetOfStringEnum(ctx, apiObject.FirewallTypes)
	}

	return diags
}

func (m scopeFilterModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case m.IncludeAll.ValueBool():
		return &awstypes.AdminScopeFilterInputMemberIncludeAll{Value: awstypes.Unit{}}, diags

	case !m.IncludeOnly.IsNull():
		model, d := m.IncludeOnly.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.AdminScopeFilterInputMemberIncludeOnly
		smerr.AddEnrich(ctx, &diags, flex.Expand(ctx, model, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.ExcludeOnly.IsNull():
		model, d := m.ExcludeOnly.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.AdminScopeFilterInputMemberExcludeOnly
		smerr.AddEnrich(ctx, &diags, flex.Expand(ctx, model, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}

	return nil, diags
}

func (m *scopeFilterModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	m.IncludeAll = types.BoolNull()
	m.IncludeOnly = fwtypes.NewListNestedObjectValueOfNull[scopeSelectionModel](ctx)
	m.ExcludeOnly = fwtypes.NewListNestedObjectValueOfNull[scopeSelectionModel](ctx)

	switch t := v.(type) {
	case awstypes.AdminScopeFilterMemberIncludeAll:
		m.IncludeAll = types.BoolValue(true)

	case *awstypes.AdminScopeFilterMemberIncludeAll:
		m.IncludeAll = types.BoolValue(true)

	case awstypes.AdminScopeFilterMemberIncludeOnly:
		var d diag.Diagnostics
		m.IncludeOnly, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, flattenScopeSelection(ctx, &t.Value))
		smerr.AddEnrich(ctx, &diags, d)

	case *awstypes.AdminScopeFilterMemberIncludeOnly:
		var d diag.Diagnostics
		m.IncludeOnly, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, flattenScopeSelection(ctx, &t.Value))
		smerr.AddEnrich(ctx, &diags, d)

	case awstypes.AdminScopeFilterMemberExcludeOnly:
		var d diag.Diagnostics
		m.ExcludeOnly, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, flattenScopeSelection(ctx, &t.Value))
		smerr.AddEnrich(ctx, &diags, d)

	case *awstypes.AdminScopeFilterMemberExcludeOnly:
		var d diag.Diagnostics
		m.ExcludeOnly, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, flattenScopeSelection(ctx, &t.Value))
		smerr.AddEnrich(ctx, &diags, d)

	case *awstypes.UnknownUnionMember:
		flex.HandleFlattenUnknownUnionMember(ctx, t.Tag, &diags)

	default:
		diags.AddError(
			"Unsupported Admin Scope Filter",
			fmt.Sprintf("scopeFilterModel.Flatten: %T", v),
		)
	}

	return diags
}

// flattenScopeSelection maps the response form of a selection (account and
// OU references with display metadata) back to the plain IDs used on input.
func flattenScopeSelection(ctx context.Context, apiObject *awstypes.AdminScopeSelection) *scopeSelectionModel { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	m := &scopeSelectionModel{
		Accounts:            fwtypes.NewSetValueOfNull[types.String](ctx),
		OrganizationalUnits: fwtypes.NewSetValueOfNull[types.String](ctx),
	}

	if len(apiObject.Accounts) > 0 {
		m.Accounts = flex.FlattenFrameworkStringValueSetOfString(ctx, tfslices.ApplyToAll(apiObject.Accounts, func(v awstypes.AccountReference) string {
			return aws.ToString(v.AccountId)
		}))
	}
	if len(apiObject.OrganizationalUnits) > 0 {
		m.OrganizationalUnits = flex.FlattenFrameworkStringValueSetOfString(ctx, tfslices.ApplyToAll(apiObject.OrganizationalUnits, func(v awstypes.OrganizationalUnitReference) string {
			return aws.ToString(v.OuId)
		}))
	}

	return m
}
