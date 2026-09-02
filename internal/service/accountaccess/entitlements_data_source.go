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
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfobjectvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/objectvalidator"
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
			path.MatchRoot(names.AttrAccountID),
			path.MatchRoot(names.AttrPrincipal),
			path.MatchRoot(names.AttrRoleARN),
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
			names.AttrRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrPrincipal: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[principalModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
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
							NestedObject: schema.NestedBlockObject{
								Validators: []validator.Object{
									tfobjectvalidator.ExactlyOneOfChildren(
										path.MatchRelative().AtName("group_id"),
										path.MatchRelative().AtName("user_id"),
									),
								},
								Attributes: map[string]schema.Attribute{
									"group_id": schema.StringAttribute{Optional: true},
									"user_id":  schema.StringAttribute{Optional: true},
								},
							},
						},
					},
				},
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
	filter, diags := entitlementsDataSourcePrincipalRoleFilter(ctx, &data)
	smerr.AddEnrich(ctx, &response.Diagnostics, diags)
	if response.Diagnostics.HasError() {
		return
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

func entitlementsDataSourcePrincipalRoleFilter(ctx context.Context, data *entitlementsDataSourceModel) (*awstypes.PrincipalRoleEntitlementFilter, diag.Diagnostics) {
	var diags diag.Diagnostics

	filter := &awstypes.PrincipalRoleEntitlementFilter{
		Account: data.AccountID.ValueStringPointer(),
	}
	if !data.RoleARN.IsNull() {
		filter.RoleArn = aws.String(fwflex.StringValueFromFramework(ctx, data.RoleARN))
	}
	if data.Principal.IsNull() {
		return filter, diags
	}

	principal, d := data.Principal.ToPtr(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return nil, diags
	}
	identityCenter, d := principal.IdentityCenter.ToPtr(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return nil, diags
	}

	var value awstypes.IdentityCenterPrincipalFilter
	switch {
	case !identityCenter.GroupID.IsNull():
		value = &awstypes.IdentityCenterPrincipalFilterMemberGroupId{
			Value: identityCenter.GroupID.ValueString(),
		}
	case !identityCenter.UserID.IsNull():
		value = &awstypes.IdentityCenterPrincipalFilterMemberUserId{
			Value: identityCenter.UserID.ValueString(),
		}
	}
	filter.Principal = &awstypes.PrincipalFilterMemberIdentityCenter{Value: value}

	return filter, diags
}

func flattenEntitlementSummary(ctx context.Context, summary awstypes.EntitlementsListMember, model *entitlementsDataSourceItemModel) diag.Diagnostics { //nolint:unparam // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	var diags diag.Diagnostics

	model.EntitlementID = fwflex.StringToFramework(ctx, summary.EntitlementId)
	model.CreatedAt = timetypes.NewRFC3339TimeValue(aws.ToTime(summary.CreatedAt))

	principalRole, ok := summary.Entitlement.(*awstypes.EntitlementSummaryMemberPrincipalRole)
	if !ok {
		return diags
	}

	principalRoleModel := principalRoleEntitlementModel{
		Account:     fwflex.StringToFramework(ctx, principalRole.Value.Account),
		AccountName: fwflex.StringToFramework(ctx, principalRole.Value.AccountName),
		RoleARN:     fwflex.StringToFrameworkARN(ctx, principalRole.Value.RoleArn),
	}
	var d diag.Diagnostics
	if principal, ok := principalRole.Value.Principal.(*awstypes.PrincipalMemberIdentityCenter); ok {
		identityCenterModel := identityCenterPrincipalModel{}
		switch value := principal.Value.(type) {
		case *awstypes.IdentityCenterPrincipalMemberGroupId:
			identityCenterModel.GroupID = types.StringValue(value.Value)
		case *awstypes.IdentityCenterPrincipalMemberUserId:
			identityCenterModel.UserID = types.StringValue(value.Value)
		}

		principalModel := principalModel{}
		principalModel.IdentityCenter, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &identityCenterModel)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return diags
		}
		principalRoleModel.Principal, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &principalModel)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return diags
		}
	}

	entitlement := entitlementModel{}
	entitlement.PrincipalRole, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &principalRoleModel)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return diags
	}
	model.Entitlement, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &entitlement)
	smerr.AddEnrich(ctx, &diags, d)

	return diags
}

type entitlementsDataSourceModel struct {
	framework.WithRegionModel
	AccountID      types.String                                                     `tfsdk:"account_id"`
	ApplicationARN fwtypes.ARN                                                      `tfsdk:"application_arn"`
	Entitlements   fwtypes.ListNestedObjectValueOf[entitlementsDataSourceItemModel] `tfsdk:"entitlements"`
	Principal      fwtypes.ListNestedObjectValueOf[principalModel]                  `tfsdk:"principal"`
	RoleARN        fwtypes.ARN                                                      `tfsdk:"role_arn"`
}

type entitlementsDataSourceItemModel struct {
	CreatedAt     timetypes.RFC3339                                 `tfsdk:"created_at"`
	Entitlement   fwtypes.ListNestedObjectValueOf[entitlementModel] `tfsdk:"entitlement"`
	EntitlementID types.String                                      `tfsdk:"entitlement_id"`
}
