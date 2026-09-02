// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package agentregistry

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/agentregistrycontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_agentregistry_registry", name="Registry")
// @Tags(identifierAttribute="registry_arn")
func newRegistryDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &registryDataSource{}, nil
}

type registryDataSource struct {
	framework.DataSourceWithModel[registryDataSourceModel]
}

func (d *registryDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Computed: true,
			},
			"registry_arn": schema.StringAttribute{
				Computed: true,
			},
			"registry_id": schema.StringAttribute{
				Required: true,
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.RegistryStatus](),
				Computed:   true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"approval_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[approvalConfigurationModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"auto_approval_rules": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							Computed:    true,
						},
					},
				},
			},
			"discovery_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[discoveryConfigurationModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"authorizer_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.RegistryAuthorizerType](),
							Computed:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"authorizer_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[authorizerConfigurationModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"allowed_audience": schema.ListAttribute{
										CustomType:  fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Computed:    true,
									},
									"allowed_clients": schema.ListAttribute{
										CustomType:  fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Computed:    true,
									},
									"allowed_scopes": schema.ListAttribute{
										CustomType:  fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Computed:    true,
									},
									"discovery_url": schema.StringAttribute{
										Computed: true,
									},
								},
								Blocks: map[string]schema.Block{
									"custom_claim": schema.SetNestedBlock{
										CustomType: fwtypes.NewSetNestedObjectTypeOf[customJWTAuthorizerCustomClaimModel](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"inbound_token_claim_name": schema.StringAttribute{
													Computed: true,
												},
												"inbound_token_claim_value_type": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.InboundTokenClaimValueType](),
													Computed:   true,
												},
											},
											Blocks: map[string]schema.Block{
												"authorizing_claim_match_value": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[customJWTAuthorizerAuthorizingClaimMatchValueModel](ctx),
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"claim_match_operator": schema.StringAttribute{
																CustomType: fwtypes.StringEnumType[awstypes.ClaimMatchOperatorType](),
																Computed:   true,
															},
														},
														Blocks: map[string]schema.Block{
															"claim_match_value": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[customJWTAuthorizerClaimMatchValueModel](ctx),
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"match_value_string": schema.StringAttribute{
																			Computed: true,
																		},
																		"match_value_string_list": schema.SetAttribute{
																			CustomType:  fwtypes.SetOfStringType,
																			ElementType: types.StringType,
																			Computed:    true,
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
			},
		},
	}
}

func (d *registryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data registryDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().AgentRegistryClient(ctx)

	out, err := findRegistryByID(ctx, conn, data.RegistryID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.RegistryID.ValueString())
		return
	}

	data.RegistryID = fwflex.StringToFramework(ctx, out.RegistryId)
	data.RegistryARN = fwflex.StringToFramework(ctx, out.RegistryArn)
	data.Name = fwflex.StringToFramework(ctx, out.Name)
	data.Description = fwflex.StringToFramework(ctx, out.Description)
	data.Status = fwtypes.StringEnumValue(out.Status)
	data.CreatedAt = timetypes.NewRFC3339TimePointerValue(out.CreatedAt)
	data.UpdatedAt = timetypes.NewRFC3339TimePointerValue(out.UpdatedAt)

	// Reuse the resource's flatten helpers for the nested configuration blocks.
	// They operate on a registryResourceModel, so flatten into a scratch model
	// and copy the shared fields across.
	var rm registryResourceModel
	resp.Diagnostics.Append(flattenApprovalConfiguration(ctx, out.ApprovalConfiguration, &rm)...)
	resp.Diagnostics.Append(flattenDiscoveryConfiguration(ctx, out.DiscoveryConfiguration, &rm)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ApprovalConfiguration = rm.ApprovalConfiguration
	data.DiscoveryConfiguration = rm.DiscoveryConfiguration

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

type registryDataSourceModel struct {
	framework.WithRegionModel
	ApprovalConfiguration  fwtypes.ListNestedObjectValueOf[approvalConfigurationModel]  `tfsdk:"approval_configuration"`
	CreatedAt              timetypes.RFC3339                                            `tfsdk:"created_at"`
	Description            types.String                                                 `tfsdk:"description"`
	DiscoveryConfiguration fwtypes.ListNestedObjectValueOf[discoveryConfigurationModel] `tfsdk:"discovery_configuration"`
	Name                   types.String                                                 `tfsdk:"name"`
	RegistryARN            types.String                                                 `tfsdk:"registry_arn"`
	RegistryID             types.String                                                 `tfsdk:"registry_id"`
	Status                 fwtypes.StringEnum[awstypes.RegistryStatus]                  `tfsdk:"status"`
	Tags                   tftags.Map                                                   `tfsdk:"tags"`
	UpdatedAt              timetypes.RFC3339                                            `tfsdk:"updated_at"`
}
