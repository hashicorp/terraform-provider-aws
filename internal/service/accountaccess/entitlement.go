// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/accountaccess"
	awstypes "github.com/aws/aws-sdk-go-v2/service/accountaccess/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

type principalType string

const (
	principalTypeUser  principalType = "USER"
	principalTypeGroup principalType = "GROUP"

	propagationTimeout = 2 * time.Minute
)

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

	applicationARN := fwflex.StringValueFromFramework(ctx, plan.ApplicationARN)
	principalID := fwflex.StringValueFromFramework(ctx, plan.PrincipalID)
	roleARN := fwflex.StringValueFromFramework(ctx, plan.RoleARN)

	var identityCenterPrincipal awstypes.IdentityCenterPrincipal
	switch plan.PrincipalType.ValueEnum() {
	case principalTypeUser:
		identityCenterPrincipal = &awstypes.IdentityCenterPrincipalMemberUserId{Value: principalID}
	case principalTypeGroup:
		identityCenterPrincipal = &awstypes.IdentityCenterPrincipalMemberGroupId{Value: principalID}
	default:
		response.Diagnostics.AddAttributeError(
			path.Root("principal_type"),
			"Invalid principal_type",
			fmt.Sprintf("Expected USER or GROUP, got %q.", plan.PrincipalType.ValueString()),
		)
		return
	}

	conn := r.Meta().AccountAccessClient(ctx)
	input := accountaccess.CreateEntitlementInput{
		ApplicationArn: aws.String(applicationARN),
		Entitlement: &awstypes.EntitlementMemberPrincipalRole{
			Value: awstypes.PrincipalRoleEntitlement{
				Principal: &awstypes.PrincipalMemberIdentityCenter{Value: identityCenterPrincipal},
				RoleArn:   aws.String(roleARN),
			},
		},
	}

	output, err := tfresource.RetryWhenIsAErrorMessageContains[*accountaccess.CreateEntitlementOutput, *awstypes.ValidationException](
		ctx, propagationTimeout,
		func(ctx context.Context) (*accountaccess.CreateEntitlementOutput, error) {
			return conn.CreateEntitlement(ctx, &input)
		},
		"verifying role",
	)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, applicationARN)
		return
	}

	plan.EntitlementID = fwflex.StringToFramework(ctx, output.EntitlementId)
	out, err := findEntitlementByTwoPartKey(ctx, conn, applicationARN, plan.EntitlementID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, buildEntitlementID(applicationARN, plan.EntitlementID.ValueString()))
		return
	}

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

	applicationARN := fwflex.StringValueFromFramework(ctx, state.ApplicationARN)
	entitlementID := fwflex.StringValueFromFramework(ctx, state.EntitlementID)
	conn := r.Meta().AccountAccessClient(ctx)
	out, err := findEntitlementByTwoPartKey(ctx, conn, applicationARN, entitlementID)
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, buildEntitlementID(applicationARN, entitlementID))
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

	applicationARN := fwflex.StringValueFromFramework(ctx, state.ApplicationARN)
	entitlementID := fwflex.StringValueFromFramework(ctx, state.EntitlementID)
	conn := r.Meta().AccountAccessClient(ctx)
	input := accountaccess.DeleteEntitlementInput{
		ApplicationArn: aws.String(applicationARN),
		EntitlementId:  aws.String(entitlementID),
	}
	_, err := conn.DeleteEntitlement(ctx, &input)
	if isEntitlementNotFoundError(err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, buildEntitlementID(applicationARN, entitlementID))
	}
}

func (r *entitlementResource) flatten(ctx context.Context, out *accountaccess.GetEntitlementOutput, model *entitlementResourceModel) diag.Diagnostics { //nolint:unparam
	var diags diag.Diagnostics

	model.ApplicationARN = fwflex.StringToFrameworkARN(ctx, out.ApplicationArn)
	model.EntitlementID = fwflex.StringToFramework(ctx, out.EntitlementId)

	principalRole, ok := out.Entitlement.(*awstypes.EntitlementDetailsMemberPrincipalRole)
	if !ok {
		return diags
	}

	model.AccountID = fwflex.StringToFramework(ctx, principalRole.Value.Account)
	model.RoleARN = fwflex.StringToFrameworkARN(ctx, principalRole.Value.RoleArn)

	if principal, ok := principalRole.Value.Principal.(*awstypes.PrincipalMemberIdentityCenter); ok {
		switch value := principal.Value.(type) {
		case *awstypes.IdentityCenterPrincipalMemberUserId:
			model.PrincipalID = types.StringValue(value.Value)
			model.PrincipalType = fwtypes.StringEnumValue(principalTypeUser)
		case *awstypes.IdentityCenterPrincipalMemberGroupId:
			model.PrincipalID = types.StringValue(value.Value)
			model.PrincipalType = fwtypes.StringEnumValue(principalTypeGroup)
		}
	}

	return diags
}

func findEntitlementByTwoPartKey(ctx context.Context, conn *accountaccess.Client, applicationARN, entitlementID string) (*accountaccess.GetEntitlementOutput, error) {
	input := accountaccess.GetEntitlementInput{
		ApplicationArn: aws.String(applicationARN),
		EntitlementId:  aws.String(entitlementID),
	}
	output, err := conn.GetEntitlement(ctx, &input)
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

func buildEntitlementID(applicationARN, entitlementID string) string {
	return applicationARN + intflex.ResourceIdSeparator + entitlementID
}

type entitlementImportID struct{}

var _ inttypes.ImportIDParser = entitlementImportID{}

func (entitlementImportID) Parse(id string) (string, map[string]any, error) {
	applicationARN, entitlementID, found := strings.Cut(id, intflex.ResourceIdSeparator)
	if !found || applicationARN == "" || entitlementID == "" {
		return "", nil, fmt.Errorf(
			"id %q should be in the format <application-arn>%s<entitlement-id>",
			id, intflex.ResourceIdSeparator,
		)
	}
	return id, map[string]any{
		"application_arn": applicationARN,
		"entitlement_id":  entitlementID,
	}, nil
}

type entitlementResourceModel struct {
	framework.WithRegionModel
	AccountID      types.String                      `tfsdk:"account_id"`
	ApplicationARN fwtypes.ARN                       `tfsdk:"application_arn"`
	EntitlementID  types.String                      `tfsdk:"entitlement_id"`
	PrincipalID    types.String                      `tfsdk:"principal_id"`
	PrincipalType  fwtypes.StringEnum[principalType] `tfsdk:"principal_type"`
	RoleARN        fwtypes.ARN                       `tfsdk:"role_arn"`
}
