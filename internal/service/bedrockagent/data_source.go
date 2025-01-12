// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	hierarchicalLevelConfigurations          = 2
	hierarchicalMaxTokens                    = 8192
	semanticBreakpointPercentileThresholdMin = 50
	semanticBreakpointPercentileThresholdMax = 99
)

// @FrameworkResource(name="Data Source")
func newDataSourceResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &dataSourceResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type dataSourceResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (*dataSourceResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_bedrockagent_data_source"
}

func (r *dataSourceResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"data_deletion_policy": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DataDeletionPolicy](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"data_source_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 200),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"knowledge_base_id": schema.StringAttribute{
				Required: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^([0-9a-zA-Z][_-]?){1,100}$`), "valid characters are a-z, A-Z, 0-9, _ (underscore) and - (hyphen). The name can have up to 100 characters"),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"data_source_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dataSourceConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrType: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.DataSourceType](),
							Required:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"s3_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[s3DataSourceConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"bucket_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
									"bucket_owner_account_id": schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											fwvalidators.AWSAccountID(),
										},
									},
									"inclusion_prefixes": schema.SetAttribute{
										CustomType:  fwtypes.SetOfStringType,
										ElementType: types.StringType,
										Optional:    true,
										Validators: []validator.Set{
											setvalidator.SizeAtMost(1),
											setvalidator.ValueStringsAre(stringvalidator.LengthBetween(1, 300)),
										},
									},
								},
							},
						},
						"salesforce_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[salesforceDataSourceConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"source_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[salesforceSourceConfiguration](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"auth_type": schema.StringAttribute{
													Required:   true,
													CustomType: fwtypes.StringEnumType[awstypes.SalesforceAuthType](),
												},
												"credentials_secret_arn": schema.StringAttribute{
													CustomType: fwtypes.ARNType,
													Required:   true,
												},
												"host_url": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 256),
														stringvalidator.RegexMatches(regexache.MustCompile(`^(https?)://[0-9A-Za-z-+&@#/%?=~_|!:,.;]*[0-9A-Za-z-+&@#/%=~_|]`), "must provide a valid HTTPS url"),
													},
												},
											},
										},
									},
									"crawler_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[salesforceCrawlerConfiguration](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: crawlerFilterConfigurationSchema[crawlFilterConfiguration](ctx),
									},
								},
							},
						},
						"share_point_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[sharepointDataSourceConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"source_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[sharepointSourceConfiguration](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"auth_type": schema.StringAttribute{
													Required:   true,
													CustomType: fwtypes.StringEnumType[awstypes.SharePointAuthType](),
												},
												"credentials_secret_arn": schema.StringAttribute{
													CustomType: fwtypes.ARNType,
													Required:   true,
												},
												"host_type": schema.StringAttribute{
													Required:   true,
													CustomType: fwtypes.StringEnumType[awstypes.SharePointHostType](),
												},
												names.AttrDomain: schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 50),
													},
												},
												"site_urls": schema.SetAttribute{
													CustomType:  fwtypes.SetOfStringType,
													ElementType: types.StringType,
													Required:    true,
													Validators: []validator.Set{
														setvalidator.SizeBetween(1, 100),
														setvalidator.ValueStringsAre(
															stringvalidator.RegexMatches(regexache.MustCompile(`^https://[A-Za-z0-9][^\s]*$`), "must provide a valid HTTPS url"),
														),
													},
												},
												"tenant_id": schema.StringAttribute{
													Optional: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(36, 36),
														stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`), "must provide a valid tenant ID"),
													},
												},
											},
										},
									},
									"crawler_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[sharepointCrawlerConfiguration](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: crawlerFilterConfigurationSchema[crawlFilterConfiguration](ctx),
									},
								},
							},
						},
						"confluence_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[confluenceDataSourceConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"source_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[confluenceSourceConfiguration](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"auth_type": schema.StringAttribute{
													Required:   true,
													CustomType: fwtypes.StringEnumType[awstypes.ConfluenceAuthType](),
												},
												"credentials_secret_arn": schema.StringAttribute{
													CustomType: fwtypes.ARNType,
													Required:   true,
												},
												"host_type": schema.StringAttribute{
													Required:   true,
													CustomType: fwtypes.StringEnumType[awstypes.ConfluenceHostType](),
												},
												"host_url": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.RegexMatches(regexache.MustCompile(`^https://[A-Za-z0-9][^\s]*$`), "must provide a valid HTTPS url"),
													},
												},
											},
										},
									},
									"crawler_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[confluenceCrawlerConfiguration](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: crawlerFilterConfigurationSchema[crawlFilterConfiguration](ctx),
									},
								},
							},
						},
						"web_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[confluenceDataSourceConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"source_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[webSourceConfiguration](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"url_configuration": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[urlConfiguration](ctx),
													Validators: []validator.List{
														listvalidator.IsRequired(),
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Blocks: map[string]schema.Block{
															"seed_urls": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[seedURL](ctx),
																Validators: []validator.List{
																	listvalidator.SizeAtLeast(1),
																	listvalidator.SizeAtMost(100),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		names.AttrURL: schema.StringAttribute{
																			Optional: true,
																			Validators: []validator.String{
																				stringvalidator.RegexMatches(regexache.MustCompile(`^https?://[A-Za-z0-9][^\s]*$`), "must provide a valid HTTPS url"),
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
									"crawler_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[webCrawlerConfiguration](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: crawlerFilterConfigurationSchema[webCrawlerConfiguration](ctx),
									},
								},
							},
						},
					},
				},
			},
			"server_side_encryption_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[serverSideEncryptionConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrKMSKeyARN: schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
			"vector_ingestion_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[vectorIngestionConfigurationModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"chunking_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[chunkingConfigurationModel](ctx),
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"chunking_strategy": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.ChunkingStrategy](),
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"fixed_size_chunking_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[fixedSizeChunkingConfigurationModel](ctx),
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
											listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("hierarchical_chunking_configuration"), path.MatchRelative().AtParent().AtName("semantic_chunking_configuration")),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"max_tokens": schema.Int64Attribute{
													Required: true,
													PlanModifiers: []planmodifier.Int64{
														int64planmodifier.RequiresReplace(),
													},
													Validators: []validator.Int64{
														int64validator.AtLeast(1),
													},
												},
												"overlap_percentage": schema.Int64Attribute{
													Required: true,
													PlanModifiers: []planmodifier.Int64{
														int64planmodifier.RequiresReplace(),
													},
													Validators: []validator.Int64{
														int64validator.Between(1, 99),
													},
												},
											},
										},
									},
									"hierarchical_chunking_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[hierarchicalChunkingConfigurationModel](ctx),
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
											listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("fixed_size_chunking_configuration"), path.MatchRelative().AtParent().AtName("semantic_chunking_configuration")),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"overlap_tokens": schema.Int32Attribute{
													Required: true,
												},
											},
											Blocks: map[string]schema.Block{
												"level_configuration": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[hierarchicalChunkingLevelConfigurationModel](ctx),
													PlanModifiers: []planmodifier.List{
														listplanmodifier.RequiresReplace(),
													},
													Validators: []validator.List{
														listvalidator.SizeBetween(hierarchicalLevelConfigurations, hierarchicalLevelConfigurations),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"max_tokens": schema.Int32Attribute{
																Required: true,
																Validators: []validator.Int32{
																	int32validator.Between(1, hierarchicalMaxTokens),
																},
															},
														},
													},
												},
											},
										},
									},
									"semantic_chunking_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[semanticChunkingConfigurationModel](ctx),
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
											listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("fixed_size_chunking_configuration"), path.MatchRelative().AtParent().AtName("hierarchical_chunking_configuration")),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"breakpoint_percentile_threshold": schema.Int32Attribute{
													Required: true,
													Validators: []validator.Int32{
														int32validator.Between(semanticBreakpointPercentileThresholdMin, semanticBreakpointPercentileThresholdMax),
													},
												},
												"buffer_size": schema.Int32Attribute{
													Required: true,
													Validators: []validator.Int32{
														int32validator.Between(0, 1),
													},
												},
												"max_token": schema.Int32Attribute{
													Required: true,
													Validators: []validator.Int32{
														int32validator.AtLeast(1),
													},
												},
											},
										},
									},
								},
							},
						},
						"custom_transformation_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[customTransformationConfigurationModel](ctx),
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"intermediate_storage": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[intermediaStorageModel](ctx),
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"s3_location": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[s3LocationModel](ctx),
													PlanModifiers: []planmodifier.List{
														listplanmodifier.RequiresReplace(),
													},
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrURI: schema.StringAttribute{
																Required: true,
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
									"transformation": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[transformationModel](ctx),
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"step_to_apply": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.StepType](),
													Required:   true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
											},
											Blocks: map[string]schema.Block{
												"transformation_function": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[transformationFunctionModel](ctx),
													PlanModifiers: []planmodifier.List{
														listplanmodifier.RequiresReplace(),
													},
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Blocks: map[string]schema.Block{
															"transformation_lambda_configuration": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[transformationLambdaConfigurationModel](ctx),
																PlanModifiers: []planmodifier.List{
																	listplanmodifier.RequiresReplace(),
																},
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"lambda_arn": schema.StringAttribute{
																			CustomType: fwtypes.ARNType,
																			Required:   true,
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
						"parsing_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[parsingConfigurationModel](ctx),
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"parsing_strategy": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.ParsingStrategy](),
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"bedrock_foundation_model_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[bedrockFoundationModelConfigurationModel](ctx),
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"model_arn": schema.StringAttribute{
													CustomType: fwtypes.ARNType,
													Required:   true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
											},
											Blocks: map[string]schema.Block{
												"parsing_prompt": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[parsingPromptModel](ctx),
													PlanModifiers: []planmodifier.List{
														listplanmodifier.RequiresReplace(),
													},
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"parsing_prompt_string": schema.StringAttribute{
																Required: true,
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

func (r *dataSourceResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data dataSourceResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	input := &bedrockagent.CreateDataSourceInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	input.ClientToken = aws.String(id.UniqueId())

	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreateDataSource(ctx, input)
	}, errCodeValidationException, "cannot assume role")

	if err != nil {
		response.Diagnostics.AddError("creating Bedrock Agent Data Source", err.Error())

		return
	}

	data.DataSourceID = fwflex.StringToFramework(ctx, outputRaw.(*bedrockagent.CreateDataSourceOutput).DataSource.DataSourceId)
	id, err := data.setID()
	if err != nil {
		response.Diagnostics.AddError("flattening resource ID Bedrock Agent Data Source", err.Error())
		return
	}
	data.ID = types.StringValue(id)

	ds, err := waitDataSourceCreated(ctx, conn, data.DataSourceID.ValueString(), data.KnowledgeBaseID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Bedrock Agent Data Source (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	data.DataDeletionPolicy = fwtypes.StringEnumValue(ds.DataDeletionPolicy)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *dataSourceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data dataSourceResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	ds, err := findDataSourceByTwoPartKey(ctx, conn, data.DataSourceID.ValueString(), data.KnowledgeBaseID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Agent Data Source (%s)", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, ds, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *dataSourceResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new dataSourceResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	input := &bedrockagent.UpdateDataSourceInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.UpdateDataSource(ctx, input)
	}, errCodeValidationException, "cannot assume role")

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating Bedrock Agent Data Source (%s)", new.DataSourceID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *dataSourceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data dataSourceResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	_, err := conn.DeleteDataSource(ctx, &bedrockagent.DeleteDataSourceInput{
		DataSourceId:    data.DataSourceID.ValueStringPointer(),
		KnowledgeBaseId: data.KnowledgeBaseID.ValueStringPointer(),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Bedrock Agent Data Source (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if _, err := waitDataSourceDeleted(ctx, conn, data.DataSourceID.ValueString(), data.KnowledgeBaseID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Bedrock Agent Data Source (%s) delete", data.ID.ValueString()), err.Error())

		return
	}
}

func findDataSourceByTwoPartKey(ctx context.Context, conn *bedrockagent.Client, dataSourceID, knowledgeBaseID string) (*awstypes.DataSource, error) {
	input := &bedrockagent.GetDataSourceInput{
		DataSourceId:    aws.String(dataSourceID),
		KnowledgeBaseId: aws.String(knowledgeBaseID),
	}

	output, err := conn.GetDataSource(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DataSource == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DataSource, nil
}

func statusDataSource(ctx context.Context, conn *bedrockagent.Client, dataSourceID, knowledgeBaseID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDataSourceByTwoPartKey(ctx, conn, dataSourceID, knowledgeBaseID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitDataSourceCreated(ctx context.Context, conn *bedrockagent.Client, dataSourceID, knowledgeBaseID string, timeout time.Duration) (*awstypes.DataSource, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  enum.Slice(awstypes.DataSourceStatusAvailable),
		Refresh: statusDataSource(ctx, conn, dataSourceID, knowledgeBaseID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DataSource); ok {
		tfresource.SetLastError(err, errors.Join(tfslices.ApplyToAll(output.FailureReasons, errors.New)...))

		return output, err
	}

	return nil, err
}

func waitDataSourceDeleted(ctx context.Context, conn *bedrockagent.Client, dataSourceID, knowledgeBaseID string, timeout time.Duration) (*awstypes.DataSource, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DataSourceStatusDeleting),
		Target:  []string{},
		Refresh: statusDataSource(ctx, conn, dataSourceID, knowledgeBaseID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DataSource); ok {
		tfresource.SetLastError(err, errors.Join(tfslices.ApplyToAll(output.FailureReasons, errors.New)...))

		return output, err
	}

	return nil, err
}

func crawlerFilterConfigurationSchema[T crawler](ctx context.Context) schema.NestedBlockObject {
	t := new(T)
	return (*t).GetCrawlerSchema(ctx)
}

type dataSourceResourceModel struct {
	DataDeletionPolicy                fwtypes.StringEnum[awstypes.DataDeletionPolicy]                         `tfsdk:"data_deletion_policy"`
	DataSourceConfiguration           fwtypes.ListNestedObjectValueOf[dataSourceConfigurationModel]           `tfsdk:"data_source_configuration"`
	DataSourceID                      types.String                                                            `tfsdk:"data_source_id"`
	Description                       types.String                                                            `tfsdk:"description"`
	ID                                types.String                                                            `tfsdk:"id"`
	KnowledgeBaseID                   types.String                                                            `tfsdk:"knowledge_base_id"`
	Name                              types.String                                                            `tfsdk:"name"`
	ServerSideEncryptionConfiguration fwtypes.ListNestedObjectValueOf[serverSideEncryptionConfigurationModel] `tfsdk:"server_side_encryption_configuration"`
	Timeouts                          timeouts.Value                                                          `tfsdk:"timeouts"`
	VectorIngestionConfiguration      fwtypes.ListNestedObjectValueOf[vectorIngestionConfigurationModel]      `tfsdk:"vector_ingestion_configuration"`
}

const (
	dataSourceResourceIDPartCount = 2
)

func (m *dataSourceResourceModel) InitFromID() error {
	parts, err := flex.ExpandResourceId(m.ID.ValueString(), dataSourceResourceIDPartCount, false)
	if err != nil {
		return err
	}

	m.DataSourceID = types.StringValue(parts[0])
	m.KnowledgeBaseID = types.StringValue(parts[1])

	return nil
}

func (m *dataSourceResourceModel) setID() (string, error) {
	parts := []string{
		m.DataSourceID.ValueString(),
		m.KnowledgeBaseID.ValueString(),
	}

	return flex.FlattenResourceId(parts, dataSourceResourceIDPartCount, false)
}

type dataSourceConfigurationModel struct {
	Type                    fwtypes.StringEnum[awstypes.DataSourceType]                             `tfsdk:"type"`
	S3Configuration         fwtypes.ListNestedObjectValueOf[s3DataSourceConfigurationModel]         `tfsdk:"s3_configuration"`
	SalesforceConfiguration fwtypes.ListNestedObjectValueOf[salesforceDataSourceConfigurationModel] `tfsdk:"salesforce_configuration"`
	SharePointConfiguration fwtypes.ListNestedObjectValueOf[sharepointDataSourceConfigurationModel] `tfsdk:"share_point_configuration"`
	ConfluenceConfiguration fwtypes.ListNestedObjectValueOf[confluenceDataSourceConfigurationModel] `tfsdk:"confluence_configuration"`
	WebConfiguration        fwtypes.ListNestedObjectValueOf[webDataSourceConfigurationModel]        `tfsdk:"web_configuration"`
}

type s3DataSourceConfigurationModel struct {
	BucketARN            fwtypes.ARN                      `tfsdk:"bucket_arn"`
	BucketOwnerAccountID types.String                     `tfsdk:"bucket_owner_account_id"`
	InclusionPrefixes    fwtypes.SetValueOf[types.String] `tfsdk:"inclusion_prefixes"`
}

type salesforceDataSourceConfigurationModel struct {
	SourceConfiguration  fwtypes.ListNestedObjectValueOf[salesforceSourceConfiguration]  `tfsdk:"source_configuration"`
	CrawlerConfiguration fwtypes.ListNestedObjectValueOf[salesforceCrawlerConfiguration] `tfsdk:"crawler_configuration"`
}

type sharepointDataSourceConfigurationModel struct {
	SourceConfiguration  fwtypes.ListNestedObjectValueOf[sharepointSourceConfiguration]  `tfsdk:"source_configuration"`
	CrawlerConfiguration fwtypes.ListNestedObjectValueOf[sharepointCrawlerConfiguration] `tfsdk:"crawler_configuration"`
}

type confluenceDataSourceConfigurationModel struct {
	SourceConfiguration  fwtypes.ListNestedObjectValueOf[confluenceSourceConfiguration]  `tfsdk:"source_configuration"`
	CrawlerConfiguration fwtypes.ListNestedObjectValueOf[confluenceCrawlerConfiguration] `tfsdk:"crawler_configuration"`
}

type webDataSourceConfigurationModel struct {
	SourceConfiguration  fwtypes.ListNestedObjectValueOf[webSourceConfiguration]  `tfsdk:"source_configuration"`
	CrawlerConfiguration fwtypes.ListNestedObjectValueOf[webCrawlerConfiguration] `tfsdk:"crawler_configuration"`
}

type serverSideEncryptionConfigurationModel struct {
	KMSKeyARN fwtypes.ARN `tfsdk:"kms_key_arn"`
}

type vectorIngestionConfigurationModel struct {
	ChunkingConfiguration             fwtypes.ListNestedObjectValueOf[chunkingConfigurationModel]             `tfsdk:"chunking_configuration"`
	CustomTransformationConfiguration fwtypes.ListNestedObjectValueOf[customTransformationConfigurationModel] `tfsdk:"custom_transformation_configuration"`
	ParsingConfiguration              fwtypes.ListNestedObjectValueOf[parsingConfigurationModel]              `tfsdk:"parsing_configuration"`
}

type parsingConfigurationModel struct {
	ParsingStrategy                     fwtypes.StringEnum[awstypes.ParsingStrategy]                              `tfsdk:"parsing_strategy"`
	BedrockFoundationModelConfiguration fwtypes.ListNestedObjectValueOf[bedrockFoundationModelConfigurationModel] `tfsdk:"bedrock_foundation_model_configuration"`
}

type customTransformationConfigurationModel struct {
	IntermediateStorage fwtypes.ListNestedObjectValueOf[intermediaStorageModel] `tfsdk:"intermediate_storage"`
	Transformation      fwtypes.ListNestedObjectValueOf[transformationModel]    `tfsdk:"transformation"`
}

type intermediaStorageModel struct {
	S3Location fwtypes.ListNestedObjectValueOf[s3LocationModel] `tfsdk:"s3_location"`
}

type s3LocationModel struct {
	Uri types.String `tfsdk:"uri"`
}

type transformationModel struct {
	StepToApply            fwtypes.StringEnum[awstypes.StepType]                        `tfsdk:"step_to_apply"`
	TransformationFunction fwtypes.ListNestedObjectValueOf[transformationFunctionModel] `tfsdk:"transformation_function"`
}

type transformationFunctionModel struct {
	TransformationLambdaConfiguration fwtypes.ListNestedObjectValueOf[transformationLambdaConfigurationModel] `tfsdk:"transformation_lambda_configuration"`
}

type transformationLambdaConfigurationModel struct {
	LambdaArn fwtypes.ARN `tfsdk:"lambda_arn"`
}

type bedrockFoundationModelConfigurationModel struct {
	ModelArn      fwtypes.ARN                                         `tfsdk:"model_arn"`
	ParsingPrompt fwtypes.ListNestedObjectValueOf[parsingPromptModel] `tfsdk:"parsing_prompt"`
}

type parsingPromptModel struct {
	ParsingPromptText types.String `tfsdk:"parsing_prompt_string"`
}

type salesforceSourceConfiguration struct {
	AuthType             fwtypes.StringEnum[awstypes.SalesforceAuthType] `tfsdk:"auth_type"`
	CredentialsSecretARN fwtypes.ARN                                     `tfsdk:"credentials_secret_arn"`
	HostURL              types.String                                    `tfsdk:"host_url"`
}

type sharepointSourceConfiguration struct {
	AuthType             fwtypes.StringEnum[awstypes.SharePointAuthType] `tfsdk:"auth_type"`
	CredentialsSecretARN fwtypes.ARN                                     `tfsdk:"credentials_secret_arn"`
	HostType             fwtypes.StringEnum[awstypes.SharePointHostType] `tfsdk:"host_type"`
	TenantID             types.String                                    `tfsdk:"tenant_id"`
	SiteURLs             fwtypes.SetValueOf[types.String]                `tfsdk:"site_urls"`
	Domain               types.String                                    `tfsdk:"domain"`
}

type confluenceSourceConfiguration struct {
	AuthType             fwtypes.StringEnum[awstypes.ConfluenceAuthType] `tfsdk:"auth_type"`
	CredentialsSecretARN fwtypes.ARN                                     `tfsdk:"credentials_secret_arn"`
	HostType             fwtypes.StringEnum[awstypes.ConfluenceHostType] `tfsdk:"host_type"`
	HostURL              types.String                                    `tfsdk:"host_url"`
}

type webSourceConfiguration struct {
	URLConfiguration fwtypes.ListNestedObjectValueOf[urlConfiguration] `tfsdk:"url_configuration"`
}

type urlConfiguration struct {
	SeedURLs fwtypes.ListNestedObjectValueOf[seedURL] `tfsdk:"seed_urls"`
}

type seedURL struct {
	URL types.String `tfsdk:"url"`
}
type salesforceCrawlerConfiguration struct {
	FilterConfiguration fwtypes.ListNestedObjectValueOf[crawlFilterConfiguration] `tfsdk:"filter_configuration"`
}

type sharepointCrawlerConfiguration struct {
	FilterConfiguration fwtypes.ListNestedObjectValueOf[crawlFilterConfiguration] `tfsdk:"filter_configuration"`
}

type confluenceCrawlerConfiguration struct {
	FilterConfiguration fwtypes.ListNestedObjectValueOf[crawlFilterConfiguration] `tfsdk:"filter_configuration"`
}

type webCrawlerConfiguration struct {
	ExclusionFilters fwtypes.SetValueOf[types.String]                  `tfsdk:"exclusion_filters"`
	InclusionFilters fwtypes.SetValueOf[types.String]                  `tfsdk:"inclusion_filters"`
	Scope            fwtypes.StringEnum[awstypes.WebScopeType]         `tfsdk:"scope"`
	UserAgent        types.String                                      `tfsdk:"user_agent"`
	CrawlerLimits    fwtypes.ListNestedObjectValueOf[webCrawlerLimits] `tfsdk:"crawler_limits"`
}

type webCrawlerLimits struct {
	MaxPages  types.Int32 `tfsdk:"max_pages"`
	RateLimit types.Int32 `tfsdk:"rate_limit"`
}

func (w webCrawlerConfiguration) GetCrawlerSchema(ctx context.Context) schema.NestedBlockObject {
	return schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"exclusion_filters": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 25),
					setvalidator.ValueStringsAre(
						stringvalidator.LengthBetween(1, 1000),
					),
				},
			},
			"inclusion_filters": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 25),
					setvalidator.ValueStringsAre(
						stringvalidator.LengthBetween(1, 1000),
					),
				},
			},
			names.AttrScope: schema.StringAttribute{
				Optional:   true,
				CustomType: fwtypes.StringEnumType[awstypes.WebScopeType](),
			},
			"user_agent": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(15, 40),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"crawler_limits": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[webCrawlerLimits](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"max_pages": schema.Int32Attribute{
							Optional: true,
							Validators: []validator.Int32{
								int32validator.AtLeast(1),
							},
						},
						"rate_limit": schema.Int32Attribute{
							Optional: true,
							Validators: []validator.Int32{
								int32validator.Between(1, 300),
							},
						},
					},
				},
			},
		},
	}
}

type crawler interface {
	GetCrawlerSchema(ctx context.Context) schema.NestedBlockObject
}

type crawlFilterConfiguration struct {
	Type                types.String                                                      `tfsdk:"type"`
	PatternObjectFilter fwtypes.ListNestedObjectValueOf[patternObjectFilterConfiguration] `tfsdk:"pattern_object_filter"`
}

type patternObjectFilterConfiguration struct {
	Filters fwtypes.ListNestedObjectValueOf[patternObjectFilter] `tfsdk:"filters"`
}

type patternObjectFilter struct {
	ObjectType       types.String                     `tfsdk:"object_type"`
	ExclusionFilters fwtypes.SetValueOf[types.String] `tfsdk:"exclusion_filters"`
	InclusionFilters fwtypes.SetValueOf[types.String] `tfsdk:"inclusion_filters"`
}

func (c crawlFilterConfiguration) GetCrawlerSchema(ctx context.Context) schema.NestedBlockObject {
	return schema.NestedBlockObject{
		Blocks: map[string]schema.Block{
			"filter_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[crawlFilterConfiguration](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrType: schema.StringAttribute{
							Required: true,
						},
					},
					Blocks: map[string]schema.Block{
						"pattern_object_filter": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[patternObjectFilterConfiguration](ctx),
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"filters": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[patternObjectFilter](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"object_type": schema.StringAttribute{
													Required: true,
												},
												"exclusion_filters": schema.SetAttribute{
													CustomType:  fwtypes.SetOfStringType,
													ElementType: types.StringType,
													Optional:    true,
													Validators: []validator.Set{
														setvalidator.SizeBetween(1, 25),
														setvalidator.ValueStringsAre(
															stringvalidator.LengthBetween(1, 1000),
														),
													},
												},
												"inclusion_filters": schema.SetAttribute{
													CustomType:  fwtypes.SetOfStringType,
													ElementType: types.StringType,
													Optional:    true,
													Validators: []validator.Set{
														setvalidator.SizeBetween(1, 25),
														setvalidator.ValueStringsAre(
															stringvalidator.LengthBetween(1, 1000),
														),
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

type chunkingConfigurationModel struct {
	ChunkingStrategy                  fwtypes.StringEnum[awstypes.ChunkingStrategy]                           `tfsdk:"chunking_strategy"`
	FixedSizeChunkingConfiguration    fwtypes.ListNestedObjectValueOf[fixedSizeChunkingConfigurationModel]    `tfsdk:"fixed_size_chunking_configuration"`
	HierarchicalChunkingConfiguration fwtypes.ListNestedObjectValueOf[hierarchicalChunkingConfigurationModel] `tfsdk:"hierarchical_chunking_configuration"`
	SemanticChunkingConfiguration     fwtypes.ListNestedObjectValueOf[semanticChunkingConfigurationModel]     `tfsdk:"semantic_chunking_configuration"`
}

type fixedSizeChunkingConfigurationModel struct {
	MaxTokens         types.Int64 `tfsdk:"max_tokens"`
	OverlapPercentage types.Int64 `tfsdk:"overlap_percentage"`
}

type semanticChunkingConfigurationModel struct {
	BreakpointPercentileThreshold types.Int32 `tfsdk:"breakpoint_percentile_threshold"`
	BufferSize                    types.Int32 `tfsdk:"buffer_size"`
	MaxTokens                     types.Int32 `tfsdk:"max_token"`
}

type hierarchicalChunkingConfigurationModel struct {
	LevelConfigurations fwtypes.ListNestedObjectValueOf[hierarchicalChunkingLevelConfigurationModel] `tfsdk:"level_configuration"`
	OverlapTokens       types.Int32                                                                  `tfsdk:"overlap_tokens"`
}

type hierarchicalChunkingLevelConfigurationModel struct {
	MaxTokens types.Int32 `tfsdk:"max_tokens"`
}
