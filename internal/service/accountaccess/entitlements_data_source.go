// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/accountaccess"
	awstypes "github.com/aws/aws-sdk-go-v2/service/accountaccess/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	tfobjectvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/objectvalidator"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
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

func entitlementsFilterBlock(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[entitlementFilterModel](ctx),
		Validators: []validator.List{
			listvalidator.IsRequired(),
			listvalidator.SizeAtLeast(1),
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Validators: []validator.Object{
				tfobjectvalidator.ExactlyOneOfChildren(
					path.MatchRelative().AtName("principal_role"),
				),
			},
			Blocks: map[string]schema.Block{
				"principal_role": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[principalRoleEntitlementFilterModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrAccountID: schema.StringAttribute{
								Optional: true,
								Validators: []validator.String{
									fwvalidators.AWSAccountID(),
								},
							},
							names.AttrRoleARN: schema.StringAttribute{
								CustomType: fwtypes.ARNType,
								Optional:   true,
							},
						},
						Blocks: map[string]schema.Block{
							names.AttrPrincipal: schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[principalFilterModel](ctx),
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
											CustomType: fwtypes.NewListNestedObjectTypeOf[identityCenterPrincipalFilterModel](ctx),
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
													"group_id": schema.StringAttribute{
														Optional: true,
													},
													"user_id": schema.StringAttribute{
														Optional: true,
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

func (d *entitlementsDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"application_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			"entitlements": framework.DataSourceComputedListOfObjectAttribute[entitlementsListMemberModel](ctx),
		},
		Blocks: map[string]schema.Block{
			names.AttrFilter: entitlementsFilterBlock(ctx),
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

	var input accountaccess.ListEntitlementsInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	items, err := findEntitlements(ctx, conn, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, items, &data.Entitlements))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func findEntitlements(ctx context.Context, conn *accountaccess.Client, input *accountaccess.ListEntitlementsInput) ([]awstypes.EntitlementsListMember, error) {
	return tfslices.CollectAndConcatWithError(listEntitlementPages(ctx, conn, input))
}

type entitlementsDataSourceModel struct {
	framework.WithRegionModel
	ApplicationARN fwtypes.ARN                                                  `tfsdk:"application_arn"`
	Entitlements   fwtypes.ListNestedObjectValueOf[entitlementsListMemberModel] `tfsdk:"entitlements"`
	Filter         fwtypes.ListNestedObjectValueOf[entitlementFilterModel]      `tfsdk:"filter"`
}

type entitlementsListMemberModel struct {
	CreatedAt     timetypes.RFC3339                                 `tfsdk:"created_at"`
	Entitlement   fwtypes.ListNestedObjectValueOf[entitlementModel] `tfsdk:"entitlement"`
	EntitlementID types.String                                      `tfsdk:"entitlement_id"`
}
