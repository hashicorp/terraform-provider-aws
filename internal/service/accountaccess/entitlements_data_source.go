// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/accountaccess"
	awstypes "github.com/aws/aws-sdk-go-v2/service/accountaccess/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_accountaccess_entitlements", name="Entitlements")
func newEntitlementsDataSource(_ context.Context) (datasource.DataSourceWithConfigure, error) {
	return &entitlementsDataSource{}, nil
}

type entitlementsDataSource struct {
	framework.DataSourceWithModel[entitlementsDataSourceModel]
}

func (d *entitlementsDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot("principal_id"),
			path.MatchRoot(names.AttrRoleARN),
			path.MatchRoot(names.AttrAccountID),
		),
		datasourcevalidator.RequiredTogether(
			path.MatchRoot("principal_id"),
			path.MatchRoot("principal_type"),
		),
	}
}

func (d *entitlementsDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAccountID: schema.StringAttribute{
				Optional: true,
			},
			"application_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			"entitlements": framework.DataSourceComputedListOfObjectAttribute[entitlementsDataSourceItemModel](ctx),
			"principal_id": schema.StringAttribute{
				Optional: true,
			},
			"principal_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[principalType](),
				Optional:   true,
			},
			names.AttrRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
			},
		},
	}
}

func (d *entitlementsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data entitlementsDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	applicationARN := fwflex.StringValueFromFramework(ctx, data.ApplicationARN)
	filter := &awstypes.PrincipalRoleEntitlementFilter{}
	if !data.RoleARN.IsNull() {
		filter.RoleArn = aws.String(fwflex.StringValueFromFramework(ctx, data.RoleARN))
	}
	if !data.AccountID.IsNull() {
		filter.Account = data.AccountID.ValueStringPointer()
	}
	if !data.PrincipalID.IsNull() {
		var principal awstypes.IdentityCenterPrincipalFilter
		switch data.PrincipalType.ValueEnum() {
		case principalTypeUser:
			principal = &awstypes.IdentityCenterPrincipalFilterMemberUserId{Value: data.PrincipalID.ValueString()}
		case principalTypeGroup:
			principal = &awstypes.IdentityCenterPrincipalFilterMemberGroupId{Value: data.PrincipalID.ValueString()}
		}
		filter.Principal = &awstypes.PrincipalFilterMemberIdentityCenter{Value: principal}
	}

	conn := d.Meta().AccountAccessClient(ctx)
	input := accountaccess.ListEntitlementsInput{
		ApplicationArn: aws.String(applicationARN),
		Filter:         &awstypes.EntitlementFilter{PrincipalRole: filter},
	}

	var items []entitlementsDataSourceItemModel
	for entitlement, err := range listEntitlements(ctx, conn, &input) {
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, applicationARN)
			return
		}

		var item entitlementsDataSourceItemModel
		smerr.AddEnrich(ctx, &response.Diagnostics, flattenEntitlementSummary(ctx, entitlement, &item))
		if response.Diagnostics.HasError() {
			return
		}
		items = append(items, item)
	}

	listValue, listDiags := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, items)
	smerr.AddEnrich(ctx, &response.Diagnostics, listDiags)
	if response.Diagnostics.HasError() {
		return
	}
	data.Entitlements = listValue

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func flattenEntitlementSummary(ctx context.Context, summary awstypes.EntitlementsListMember, model *entitlementsDataSourceItemModel) diag.Diagnostics { //nolint:unparam // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	var diags diag.Diagnostics

	model.EntitlementID = fwflex.StringToFramework(ctx, summary.EntitlementId)
	model.CreatedAt = timetypes.NewRFC3339TimeValue(aws.ToTime(summary.CreatedAt))

	principalRole, ok := summary.Entitlement.(*awstypes.EntitlementSummaryMemberPrincipalRole)
	if !ok {
		return diags
	}

	model.AccountID = fwflex.StringToFramework(ctx, principalRole.Value.Account)
	model.AccountName = fwflex.StringToFramework(ctx, principalRole.Value.AccountName)
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

type entitlementsDataSourceModel struct {
	framework.WithRegionModel
	AccountID      types.String                                                     `tfsdk:"account_id"`
	ApplicationARN fwtypes.ARN                                                      `tfsdk:"application_arn"`
	Entitlements   fwtypes.ListNestedObjectValueOf[entitlementsDataSourceItemModel] `tfsdk:"entitlements"`
	PrincipalID    types.String                                                     `tfsdk:"principal_id"`
	PrincipalType  fwtypes.StringEnum[principalType]                                `tfsdk:"principal_type"`
	RoleARN        fwtypes.ARN                                                      `tfsdk:"role_arn"`
}

type entitlementsDataSourceItemModel struct {
	AccountID     types.String                      `tfsdk:"account_id"`
	AccountName   types.String                      `tfsdk:"account_name"`
	CreatedAt     timetypes.RFC3339                 `tfsdk:"created_at"`
	EntitlementID types.String                      `tfsdk:"entitlement_id"`
	PrincipalID   types.String                      `tfsdk:"principal_id"`
	PrincipalType fwtypes.StringEnum[principalType] `tfsdk:"principal_type"`
	RoleARN       fwtypes.ARN                       `tfsdk:"role_arn"`
}
