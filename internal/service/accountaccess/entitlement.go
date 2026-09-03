// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess

import (
	"context"
	"fmt"
	"strings"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/accountaccess"
	awstypes "github.com/aws/aws-sdk-go-v2/service/accountaccess/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfobjectvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/objectvalidator"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_accountaccess_entitlement", name="Entitlement")
// @IdentityAttribute("application_arn")
// @IdentityAttribute("entitlement_id")
// @ImportIDHandler("entitlementImportID")
// @Testing(preCheck="testAccPreCheck")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdFunc=testAccEntitlementImportStateIDFunc)
// @Testing(importStateIdAttribute="entitlement_id")
// @Testing(importIgnore="entitlement.0.principal_role.0.account_name", plannableImportAction="NoOp")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/accountaccess;accountaccess.GetEntitlementOutput")
// @Testing(identityRegionOverrideTest=false)
// @Testing(serialize=true)
func newEntitlementResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &entitlementResource{}, nil
}

type entitlementResource struct {
	framework.ResourceWithModel[entitlementResourceModel]
	framework.WithNoUpdate
	framework.WithImportByIdentity
}

func (r *entitlementResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"application_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"entitlement_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"entitlement": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[entitlementModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Validators: []validator.Object{
						tfobjectvalidator.ExactlyOneOfChildren(
							path.MatchRelative().AtName("principal_role"),
						),
					},
					Blocks: map[string]schema.Block{
						"principal_role": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[principalRoleEntitlementModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrAccountID: schema.StringAttribute{
										Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"account_name": schema.StringAttribute{
										Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									names.AttrRoleARN: schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									names.AttrPrincipal: schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[principalModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtLeast(1),
											listvalidator.SizeAtMost(1),
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										NestedObject: schema.NestedBlockObject{
											Validators: []validator.Object{
												tfobjectvalidator.ExactlyOneOfChildren(
													path.MatchRelative().AtName("identity_center"),
												),
											},
											Blocks: map[string]schema.Block{
												"identity_center": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[identityCenterPrincipalModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													PlanModifiers: []planmodifier.List{
														listplanmodifier.RequiresReplace(),
													},
													NestedObject: schema.NestedBlockObject{
														Validators: []validator.Object{
															tfobjectvalidator.ExactlyOneOfChildren(
																path.MatchRelative().AtName("group_id"),
																path.MatchRelative().AtName("user_id"),
															),
														},
														Attributes: map[string]schema.Attribute{
															"group_id": schema.StringAttribute{
																Optional: true,
																PlanModifiers: []planmodifier.String{
																	stringplanmodifier.RequiresReplace(),
																},
															},
															"user_id": schema.StringAttribute{
																Optional: true,
																PlanModifiers: []planmodifier.String{
																	stringplanmodifier.RequiresReplace(),
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *entitlementResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan entitlementResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AccountAccessClient(ctx)

	applicationARN := fwflex.StringValueFromFramework(ctx, plan.ApplicationARN)
	var input accountaccess.CreateEntitlementInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if response.Diagnostics.HasError() {
		return
	}

	output, err := tfresource.RetryWhenIsAErrorMessageContains[*accountaccess.CreateEntitlementOutput, *awstypes.ValidationException](ctx, propagationTimeout,
		func(ctx context.Context) (*accountaccess.CreateEntitlementOutput, error) {
			return conn.CreateEntitlement(ctx, &input)
		},
		"verifying role",
	)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, applicationARN)
		return
	}

	entitlementID := aws.ToString(output.EntitlementId)
	out, err := findEntitlementByTwoPartKey(ctx, conn, applicationARN, entitlementID)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, entitlementID)
		return
	}

	// Set values for unknowns.
	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, out, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *entitlementResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state entitlementResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AccountAccessClient(ctx)

	applicationARN, entitlementID := fwflex.StringValueFromFramework(ctx, state.ApplicationARN), fwflex.StringValueFromFramework(ctx, state.EntitlementID)
	out, err := findEntitlementByTwoPartKey(ctx, conn, applicationARN, entitlementID)
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, entitlementID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, out, &state))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &state))
}

func (r *entitlementResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state entitlementResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AccountAccessClient(ctx)

	applicationARN, entitlementID := fwflex.StringValueFromFramework(ctx, state.ApplicationARN), fwflex.StringValueFromFramework(ctx, state.EntitlementID)
	input := accountaccess.DeleteEntitlementInput{
		ApplicationArn: aws.String(applicationARN),
		EntitlementId:  aws.String(entitlementID),
	}
	_, err := conn.DeleteEntitlement(ctx, &input)
	if isEntitlementNotFoundError(err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, entitlementID)
	}
}

func (r *entitlementResource) flatten(ctx context.Context, out *accountaccess.GetEntitlementOutput, model *entitlementResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.Append(fwflex.Flatten(ctx, out, model)...)
	return diags
}

func findEntitlementByTwoPartKey(ctx context.Context, conn *accountaccess.Client, applicationARN, entitlementID string) (*accountaccess.GetEntitlementOutput, error) {
	input := accountaccess.GetEntitlementInput{
		ApplicationArn: aws.String(applicationARN),
		EntitlementId:  aws.String(entitlementID),
	}
	return findEntitlement(ctx, conn, &input)
}

func findEntitlement(ctx context.Context, conn *accountaccess.Client, input *accountaccess.GetEntitlementInput) (*accountaccess.GetEntitlementOutput, error) {
	output, err := conn.GetEntitlement(ctx, input)
	if isEntitlementNotFoundError(err) {
		return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
	}
	if err != nil {
		return nil, err
	}
	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}
	return output, nil
}

const entitlementImportIDSeparator = intflex.ResourceIdSeparator

func entitlementParseImportID(id string) (string, string, error) {
	parts := strings.Split(id, entitlementImportIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected application_arn%[2]sentitlement_id", id, entitlementImportIDSeparator)
}

var (
	_ inttypes.ImportIDParser = entitlementImportID{}
)

type entitlementImportID struct{}

func (entitlementImportID) Parse(id string) (string, map[string]any, error) {
	applicationARN, entitlementID, err := entitlementParseImportID(id)
	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		"application_arn": applicationARN,
		"entitlement_id":  entitlementID,
	}

	return id, result, nil
}

type entitlementResourceModel struct {
	framework.WithRegionModel
	ApplicationARN fwtypes.ARN                                       `tfsdk:"application_arn"`
	Entitlement    fwtypes.ListNestedObjectValueOf[entitlementModel] `tfsdk:"entitlement"`
	EntitlementID  types.String                                      `tfsdk:"entitlement_id"`
}

type entitlementModel struct {
	PrincipalRole fwtypes.ListNestedObjectValueOf[principalRoleEntitlementModel] `tfsdk:"principal_role"`
}

var (
	_ fwflex.Expander  = entitlementModel{}
	_ fwflex.Flattener = &entitlementModel{}
)

func (m *entitlementModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.EntitlementDetailsMemberPrincipalRole:
		var model principalRoleEntitlementModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		var d diag.Diagnostics
		m.PrincipalRole, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &model)
		smerr.AddEnrich(ctx, &diags, d)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("entitlementModel.Flatten: %T", v),
		)
	}

	return diags
}

func (m entitlementModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.PrincipalRole.IsNull():
		model, d := m.PrincipalRole.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.EntitlementMemberPrincipalRole
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, model, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}

	return nil, diags
}

type principalRoleEntitlementModel struct {
	Account     types.String                                    `tfsdk:"account_id"`
	AccountName types.String                                    `tfsdk:"account_name"`
	Principal   fwtypes.ListNestedObjectValueOf[principalModel] `tfsdk:"principal"`
	RoleARN     fwtypes.ARN                                     `tfsdk:"role_arn"`
}

type principalModel struct {
	IdentityCenter fwtypes.ListNestedObjectValueOf[identityCenterPrincipalModel] `tfsdk:"identity_center"`
}

var (
	_ fwflex.Expander  = principalModel{}
	_ fwflex.Flattener = &principalModel{}
)

func (m *principalModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.PrincipalMemberIdentityCenter:
		var model identityCenterPrincipalModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		var d diag.Diagnostics
		m.IdentityCenter, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &model)
		smerr.AddEnrich(ctx, &diags, d)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("principalModel.Flatten: %T", v),
		)
	}

	return diags
}

func (m principalModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.IdentityCenter.IsNull():
		model, d := m.IdentityCenter.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.PrincipalMemberIdentityCenter
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, model, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}

	return nil, diags
}

type identityCenterPrincipalModel struct {
	GroupID types.String `tfsdk:"group_id"`
	UserID  types.String `tfsdk:"user_id"`
}

var (
	_ fwflex.Expander  = identityCenterPrincipalModel{}
	_ fwflex.Flattener = &identityCenterPrincipalModel{}
)

func (m *identityCenterPrincipalModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.IdentityCenterPrincipalMemberGroupId:
		m.GroupID = fwflex.StringValueToFramework(ctx, t.Value)

	case awstypes.IdentityCenterPrincipalMemberUserId:
		m.UserID = fwflex.StringValueToFramework(ctx, t.Value)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("identityCenterPrincipalModel.Flatten: %T", v),
		)
	}

	return diags
}

func (m identityCenterPrincipalModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.GroupID.IsNull():
		r := awstypes.IdentityCenterPrincipalMemberGroupId{
			Value: fwflex.StringValueFromFramework(ctx, m.GroupID),
		}
		return &r, diags

	case !m.UserID.IsNull():
		r := awstypes.IdentityCenterPrincipalMemberUserId{
			Value: fwflex.StringValueFromFramework(ctx, m.UserID),
		}
		return &r, diags
	}

	return nil, diags
}
