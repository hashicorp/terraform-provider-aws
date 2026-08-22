// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/accountaccess"
	awstypes "github.com/aws/aws-sdk-go-v2/service/accountaccess/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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
		// ListEntitlements requires at least one filter dimension. Either pair
		// (principal_id + principal_type), or role_arn, or account_id.
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot("principal_id"),
			path.MatchRoot(names.AttrRoleARN),
			path.MatchRoot(names.AttrAccountID),
		),
		// principal_id and principal_type must be set together.
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
			names.AttrID: schema.StringAttribute{
				Computed: true,
			},
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

	conn := d.Meta().AccountAccessClient(ctx)

	filter := &awstypes.PrincipalRoleEntitlementFilter{}
	if !data.RoleARN.IsNull() {
		filter.RoleArn = data.RoleARN.ValueStringPointer()
	}
	if !data.AccountID.IsNull() {
		filter.Account = data.AccountID.ValueStringPointer()
	}
	if !data.PrincipalID.IsNull() {
		var idcFilter awstypes.IdentityCenterPrincipalFilter
		switch data.PrincipalType.ValueEnum() {
		case principalTypeUser:
			idcFilter = &awstypes.IdentityCenterPrincipalFilterMemberUserId{Value: data.PrincipalID.ValueString()}
		case principalTypeGroup:
			idcFilter = &awstypes.IdentityCenterPrincipalFilterMemberGroupId{Value: data.PrincipalID.ValueString()}
		}
		filter.Principal = &awstypes.PrincipalFilterMemberIdentityCenter{Value: idcFilter}
	}

	input := &accountaccess.ListEntitlementsInput{
		ApplicationArn: data.ApplicationARN.ValueStringPointer(),
		Filter:         &awstypes.EntitlementFilter{PrincipalRole: filter},
	}

	var entitlements []awstypes.EntitlementsListMember
	var nextToken *string
	for {
		input.NextToken = nextToken
		out, err := conn.ListEntitlements(ctx, input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ApplicationARN.ValueString())
			return
		}
		if out == nil {
			smerr.AddError(ctx, &response.Diagnostics, errors.New("empty response"), smerr.ID, data.ApplicationARN.ValueString())
			return
		}
		entitlements = append(entitlements, out.Entitlements...)
		if out.NextToken == nil {
			break
		}
		nextToken = out.NextToken
	}

	items := make([]entitlementsDataSourceItemModel, 0, len(entitlements))
	for _, e := range entitlements {
		item := entitlementsDataSourceItemModel{
			EntitlementID: fwflex.StringToFramework(ctx, e.EntitlementId),
			CreatedAt:     timetypes.NewRFC3339TimeValue(aws.ToTime(e.CreatedAt)),
		}
		flattenEntitlementSummary(ctx, e.Entitlement, &item)
		items = append(items, item)
	}

	listValue, listDiags := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, items)
	smerr.AddEnrich(ctx, &response.Diagnostics, listDiags)
	if response.Diagnostics.HasError() {
		return
	}
	data.Entitlements = listValue

	// Synthetic ID combining the application ARN with each filter axis;
	// good enough to keep Terraform happy without claiming uniqueness.
	data.ID = data.ApplicationARN.StringValue

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func flattenEntitlementSummary(ctx context.Context, summary awstypes.EntitlementSummary, model *entitlementsDataSourceItemModel) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	pre, ok := summary.(*awstypes.EntitlementSummaryMemberPrincipalRole)
	if !ok {
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

type entitlementsDataSourceModel struct {
	framework.WithRegionModel
	AccountID      types.String                                                     `tfsdk:"account_id"`
	ApplicationARN fwtypes.ARN                                                      `tfsdk:"application_arn"`
	Entitlements   fwtypes.ListNestedObjectValueOf[entitlementsDataSourceItemModel] `tfsdk:"entitlements"`
	ID             types.String                                                     `tfsdk:"id"`
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
