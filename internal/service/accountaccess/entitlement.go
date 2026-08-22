// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/accountaccess"
	awstypes "github.com/aws/aws-sdk-go-v2/service/accountaccess/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// principalType is the schema-level enum for the principal_type attribute.
// Maps to the IdentityCenterPrincipal union members in the SDK.
type principalType string

const (
	principalTypeUser  principalType = "USER"
	principalTypeGroup principalType = "GROUP"
)

// propagationTimeout bounds retries of CreateEntitlement while the target IAM
// role (often created in the same apply) propagates to the Account Access
// service's role-verification check. Two minutes is the provider-standard IAM
// eventual-consistency timeout.
const propagationTimeout = 2 * time.Minute

func (principalType) Values() []principalType {
	return []principalType{principalTypeUser, principalTypeGroup}
}

// @FrameworkResource("aws_accountaccess_entitlement", name="Entitlement")
// @IdentityAttribute("application_arn")
// @IdentityAttribute("entitlement_id")
// @ImportIDHandler("entitlementImportID")
// @Testing(preCheck="testAccPreCheck")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttributes="application_arn;entitlement_id", importStateIdAttributesSep="flex.ResourceIdSeparator")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/accountaccess;accountaccess.GetEntitlementOutput")
// @Testing(importIgnore="account_name")
// @Testing(identityRegionOverrideTest=false)
// @Testing(plannableImportAction="NoOp")
// @Testing(serialize=true)
func newEntitlementResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &entitlementResource{}, nil
}

type entitlementResource struct {
	framework.ResourceWithModel[entitlementResourceModel]
	framework.WithImportByIdentity
}

func (r *entitlementResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
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
			"application_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"entitlement_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"principal_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"principal_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[principalType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
	}
}

func (r *entitlementResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan entitlementResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AccountAccessClient(ctx)

	// Build the IdentityCenterPrincipal union from principal_id + principal_type.
	var idcPrincipal awstypes.IdentityCenterPrincipal
	switch plan.PrincipalType.ValueEnum() {
	case principalTypeUser:
		idcPrincipal = &awstypes.IdentityCenterPrincipalMemberUserId{Value: plan.PrincipalID.ValueString()}
	case principalTypeGroup:
		idcPrincipal = &awstypes.IdentityCenterPrincipalMemberGroupId{Value: plan.PrincipalID.ValueString()}
	default:
		response.Diagnostics.AddAttributeError(
			path.Root("principal_type"),
			"invalid principal_type",
			fmt.Sprintf("expected USER or GROUP, got %q", plan.PrincipalType.ValueString()),
		)
		return
	}

	input := &accountaccess.CreateEntitlementInput{
		ApplicationArn: plan.ApplicationARN.ValueStringPointer(),
		Entitlement: &awstypes.EntitlementMemberPrincipalRole{
			Value: awstypes.PrincipalRoleEntitlement{
				Principal: &awstypes.PrincipalMemberIdentityCenter{Value: idcPrincipal},
				RoleArn:   fwflex.StringFromFramework(ctx, plan.RoleARN),
			},
		},
	}

	// Retry on the transient role-verification failure. AAM's CreateEntitlement
	// verifies that role_arn exists and trusts the service; when the role was
	// created in the same apply, IAM's eventual consistency means AAM may not
	// see it yet, returning ValidationException "Error while verifying role...".
	output, err := tfresource.RetryWhenIsAErrorMessageContains[*accountaccess.CreateEntitlementOutput, *awstypes.ValidationException](
		ctx, propagationTimeout,
		func(ctx context.Context) (*accountaccess.CreateEntitlementOutput, error) {
			return conn.CreateEntitlement(ctx, input)
		},
		"verifying role",
	)
	if err != nil {
		// NOTE: AccountAccess Create has no clientToken (see questions.md AAM Q3).
		// A 5xx after server-side creation would leave us unable to safely retry.
		// We surface the error and ask the user to import. Revisit when the
		// service team confirms recovery semantics or adds clientToken pre-GA.
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.ApplicationARN.ValueString())
		return
	}
	if output == nil || output.EntitlementId == nil {
		smerr.AddError(ctx, &response.Diagnostics, errors.New("empty output"), smerr.ID, plan.ApplicationARN.ValueString())
		return
	}

	plan.EntitlementID = fwflex.StringToFramework(ctx, output.EntitlementId)
	plan.ID = types.StringValue(buildEntitlementID(plan.ApplicationARN.ValueString(), plan.EntitlementID.ValueString()))

	// Read after create to populate computed account_id / account_name.
	out, err := FindEntitlementByTwoPartKey(ctx, conn, plan.ApplicationARN.ValueString(), plan.EntitlementID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.ID.ValueString())
		return
	}

	flattenEntitlement(ctx, out, &plan)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *entitlementResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state entitlementResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AccountAccessClient(ctx)

	out, err := FindEntitlementByTwoPartKey(ctx, conn, state.ApplicationARN.ValueString(), state.EntitlementID.ValueString())
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.ID.ValueString())
		return
	}

	flattenEntitlement(ctx, out, &state)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &state))
}

func (r *entitlementResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state entitlementResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AccountAccessClient(ctx)

	_, err := conn.DeleteEntitlement(ctx, &accountaccess.DeleteEntitlementInput{
		ApplicationArn: state.ApplicationARN.ValueStringPointer(),
		EntitlementId:  state.EntitlementID.ValueStringPointer(),
	})
	if isNotFoundError(err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.ID.ValueString())
		return
	}
}

// flattenEntitlement copies fields from the SDK GetEntitlement response onto
// the Terraform model. Handles the EntitlementDetails / Principal Smithy
// unions by pattern-matching on the member types.
func flattenEntitlement(ctx context.Context, out *accountaccess.GetEntitlementOutput, model *entitlementResourceModel) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	model.ApplicationARN = fwtypes.ARNValue(aws.ToString(out.ApplicationArn))
	model.EntitlementID = fwflex.StringToFramework(ctx, out.EntitlementId)
	model.CreatedAt = timetypes.NewRFC3339TimeValue(aws.ToTime(out.CreatedAt))
	model.ID = types.StringValue(buildEntitlementID(aws.ToString(out.ApplicationArn), aws.ToString(out.EntitlementId)))

	pre, ok := out.Entitlement.(*awstypes.EntitlementDetailsMemberPrincipalRole)
	if !ok {
		// Spec only defines principalRole today. If a future variant appears,
		// leave the principal/role/account fields untouched and rely on the
		// drift the user will see on next plan.
		return
	}

	model.RoleARN = fwtypes.ARNValue(aws.ToString(pre.Value.RoleArn))
	model.AccountID = fwflex.StringToFramework(ctx, pre.Value.Account)
	model.AccountName = fwflex.StringToFramework(ctx, pre.Value.AccountName)

	if p, ok := pre.Value.Principal.(*awstypes.PrincipalMemberIdentityCenter); ok {
		switch v := p.Value.(type) {
		case *awstypes.IdentityCenterPrincipalMemberUserId:
			model.PrincipalID = types.StringValue(v.Value)
			model.PrincipalType = fwtypes.StringEnumValue(principalTypeUser)
		case *awstypes.IdentityCenterPrincipalMemberGroupId:
			model.PrincipalID = types.StringValue(v.Value)
			model.PrincipalType = fwtypes.StringEnumValue(principalTypeGroup)
		}
	}
}

// FindEntitlementByTwoPartKey looks up an Entitlement by its (applicationArn, entitlementId).
// Returns retry.NotFoundError when the API returns ResourceNotFoundException.
func FindEntitlementByTwoPartKey(ctx context.Context, conn *accountaccess.Client, applicationArn, entitlementID string) (*accountaccess.GetEntitlementOutput, error) {
	output, err := conn.GetEntitlement(ctx, &accountaccess.GetEntitlementInput{
		ApplicationArn: aws.String(applicationArn),
		EntitlementId:  aws.String(entitlementID),
	})
	if isNotFoundError(err) {
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

// buildEntitlementID composes the resource ID from its two parts.
func buildEntitlementID(applicationArn, entitlementID string) string {
	return applicationArn + intflex.ResourceIdSeparator + entitlementID
}

// entitlementImportID parses the composite import ID
// "<applicationArn>,<entitlementId>" produced by buildEntitlementID, and
// satisfies inttypes.ImportIDParser used by @ImportIDHandler.
type entitlementImportID struct{}

var _ inttypes.ImportIDParser = entitlementImportID{}

func (entitlementImportID) Parse(id string) (string, map[string]any, error) {
	applicationArn, entitlementID, found := strings.Cut(id, intflex.ResourceIdSeparator)
	if !found || applicationArn == "" || entitlementID == "" {
		return "", nil, fmt.Errorf(
			"id %q should be in the format <application-arn>%s<entitlement-id>",
			id, intflex.ResourceIdSeparator,
		)
	}
	return id, map[string]any{
		"application_arn": applicationArn,
		"entitlement_id":  entitlementID,
	}, nil
}

// entitlementResourceModel is the Terraform state representation of an
// Account Access Entitlement.
type entitlementResourceModel struct {
	framework.WithRegionModel
	AccountID      types.String                      `tfsdk:"account_id"`
	AccountName    types.String                      `tfsdk:"account_name"`
	ApplicationARN fwtypes.ARN                       `tfsdk:"application_arn"`
	CreatedAt      timetypes.RFC3339                 `tfsdk:"created_at"`
	EntitlementID  types.String                      `tfsdk:"entitlement_id"`
	ID             types.String                      `tfsdk:"id"`
	PrincipalID    types.String                      `tfsdk:"principal_id"`
	PrincipalType  fwtypes.StringEnum[principalType] `tfsdk:"principal_type"`
	RoleARN        fwtypes.ARN                       `tfsdk:"role_arn"`
}
