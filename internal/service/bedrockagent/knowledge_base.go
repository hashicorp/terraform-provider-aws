// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagent_knowledge_base", name="Knowledge Base")
// @Tags(identifierAttribute="arn")
func newKnowledgeBaseResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &knowledgeBaseResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type knowledgeBaseResource struct {
	framework.ResourceWithModel[knowledgeBaseResourceModel]
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *knowledgeBaseResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			"failure_reasons": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			names.AttrRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"knowledge_base_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[knowledgeBaseConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrType: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.KnowledgeBaseType](),
							Required:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"kendra_knowledge_base_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[kendraKnowledgeBaseConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("kendra_knowledge_base_configuration"),
									path.MatchRelative().AtParent().AtName("sql_knowledge_base_configuration"),
									path.MatchRelative().AtParent().AtName("vector_knowledge_base_configuration"),
								),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"kendra_index_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
							},
						},
						"sql_knowledge_base_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[sqlKnowledgeBaseConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrType: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.QueryEngineType](),
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"redshift_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[redshiftConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
											listvalidator.ExactlyOneOf(
												path.MatchRelative().AtParent().AtName("redshift_configuration"),
											),
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"query_engine_configuration": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[redshiftQueryEngineConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.IsRequired(),
														listvalidator.SizeAtLeast(1),
														listvalidator.SizeAtMost(1),
													},
													PlanModifiers: []planmodifier.List{
														listplanmodifier.RequiresReplace(),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrType: schema.StringAttribute{
																CustomType: fwtypes.StringEnumType[awstypes.RedshiftQueryEngineType](),
																Required:   true,
																PlanModifiers: []planmodifier.String{
																	stringplanmodifier.RequiresReplace(),
																},
															},
														},
														Blocks: map[string]schema.Block{
															"provisioned_configuration": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[redshiftProvisionedConfigurationModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																	listvalidator.ExactlyOneOf(
																		path.MatchRelative().AtParent().AtName("provisioned_configuration"),
																		path.MatchRelative().AtParent().AtName("serverless_configuration"),
																	),
																},
																PlanModifiers: []planmodifier.List{
																	listplanmodifier.RequiresReplace(),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		names.AttrClusterIdentifier: schema.StringAttribute{
																			Required: true,
																			PlanModifiers: []planmodifier.String{
																				stringplanmodifier.RequiresReplace(),
																			},
																		},
																	},
																	Blocks: map[string]schema.Block{
																		"auth_configuration": schema.ListNestedBlock{
																			CustomType: fwtypes.NewListNestedObjectTypeOf[redshiftProvisionedAuthConfigurationModel](ctx),
																			Validators: []validator.List{
																				listvalidator.IsRequired(),
																				listvalidator.SizeAtLeast(1),
																				listvalidator.SizeAtMost(1),
																			},
																			PlanModifiers: []planmodifier.List{
																				listplanmodifier.RequiresReplace(),
																			},
																			NestedObject: schema.NestedBlockObject{
																				Attributes: map[string]schema.Attribute{
																					"database_user": schema.StringAttribute{
																						Optional: true,
																						PlanModifiers: []planmodifier.String{
																							stringplanmodifier.RequiresReplace(),
																						},
																					},
																					names.AttrType: schema.StringAttribute{
																						CustomType: fwtypes.StringEnumType[awstypes.RedshiftProvisionedAuthType](),
																						Required:   true,
																						PlanModifiers: []planmodifier.String{
																							stringplanmodifier.RequiresReplace(),
																						},
																					},
																					"username_password_secret_arn": schema.StringAttribute{
																						CustomType: fwtypes.ARNType,
																						Optional:   true,
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
															"serverless_configuration": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[redshiftServerlessConfigurationModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
																PlanModifiers: []planmodifier.List{
																	listplanmodifier.RequiresReplace(),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"workgroup_arn": schema.StringAttribute{
																			CustomType: fwtypes.ARNType,
																			Required:   true,
																			PlanModifiers: []planmodifier.String{
																				stringplanmodifier.RequiresReplace(),
																			},
																		},
																	},
																	Blocks: map[string]schema.Block{
																		"auth_configuration": schema.ListNestedBlock{
																			CustomType: fwtypes.NewListNestedObjectTypeOf[redshiftServerlessAuthConfigurationModel](ctx),
																			Validators: []validator.List{
																				listvalidator.IsRequired(),
																				listvalidator.SizeAtLeast(1),
																				listvalidator.SizeAtMost(1),
																			},
																			PlanModifiers: []planmodifier.List{
																				listplanmodifier.RequiresReplace(),
																			},
																			NestedObject: schema.NestedBlockObject{
																				Attributes: map[string]schema.Attribute{
																					names.AttrType: schema.StringAttribute{
																						CustomType: fwtypes.StringEnumType[awstypes.RedshiftServerlessAuthType](),
																						Required:   true,
																						PlanModifiers: []planmodifier.String{
																							stringplanmodifier.RequiresReplace(),
																						},
																					},
																					"username_password_secret_arn": schema.StringAttribute{
																						CustomType: fwtypes.ARNType,
																						Optional:   true,
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
												"query_generation_configuration": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[queryGenerationConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													PlanModifiers: []planmodifier.List{
														listplanmodifier.RequiresReplace(),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"execution_timeout_seconds": schema.Int64Attribute{
																Optional: true,
																Validators: []validator.Int64{
																	int64validator.Between(1, 200),
																},
																PlanModifiers: []planmodifier.Int64{
																	int64planmodifier.RequiresReplace(),
																},
															},
														},
														Blocks: map[string]schema.Block{
															"generation_context": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[queryGenerationContextModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
																PlanModifiers: []planmodifier.List{
																	listplanmodifier.RequiresReplace(),
																},
																NestedObject: schema.NestedBlockObject{
																	Blocks: map[string]schema.Block{
																		"curated_query": schema.ListNestedBlock{
																			CustomType: fwtypes.NewListNestedObjectTypeOf[curatedQueryModel](ctx),
																			Validators: []validator.List{
																				listvalidator.SizeAtMost(10),
																			},
																			PlanModifiers: []planmodifier.List{
																				listplanmodifier.RequiresReplace(),
																			},
																			NestedObject: schema.NestedBlockObject{
																				Attributes: map[string]schema.Attribute{
																					"natural_language": schema.StringAttribute{
																						Required: true,
																						PlanModifiers: []planmodifier.String{
																							stringplanmodifier.RequiresReplace(),
																						},
																					},
																					"sql": schema.StringAttribute{
																						Required: true,
																						PlanModifiers: []planmodifier.String{
																							stringplanmodifier.RequiresReplace(),
																						},
																					},
																				},
																			},
																		},
																		"table": schema.ListNestedBlock{
																			CustomType: fwtypes.NewListNestedObjectTypeOf[queryGenerationTableModel](ctx),
																			Validators: []validator.List{
																				listvalidator.SizeAtMost(50),
																			},
																			PlanModifiers: []planmodifier.List{
																				listplanmodifier.RequiresReplace(),
																			},
																			NestedObject: schema.NestedBlockObject{
																				Attributes: map[string]schema.Attribute{
																					names.AttrDescription: schema.StringAttribute{
																						Optional: true,
																						Validators: []validator.String{
																							stringvalidator.LengthBetween(1, 200),
																						},
																						PlanModifiers: []planmodifier.String{
																							stringplanmodifier.RequiresReplace(),
																						},
																					},
																					"inclusion": schema.StringAttribute{
																						CustomType: fwtypes.StringEnumType[awstypes.IncludeExclude](),
																						Optional:   true,
																						PlanModifiers: []planmodifier.String{
																							stringplanmodifier.RequiresReplace(),
																						},
																					},
																					names.AttrName: schema.StringAttribute{
																						Required: true,
																						PlanModifiers: []planmodifier.String{
																							stringplanmodifier.RequiresReplace(),
																						},
																					},
																				},
																				Blocks: map[string]schema.Block{
																					"column": schema.ListNestedBlock{
																						CustomType: fwtypes.NewListNestedObjectTypeOf[queryGenerationColumnModel](ctx),
																						PlanModifiers: []planmodifier.List{
																							listplanmodifier.RequiresReplace(),
																						},
																						NestedObject: schema.NestedBlockObject{
																							Attributes: map[string]schema.Attribute{
																								names.AttrDescription: schema.StringAttribute{
																									Optional: true,
																									Validators: []validator.String{
																										stringvalidator.LengthBetween(1, 200),
																									},
																									PlanModifiers: []planmodifier.String{
																										stringplanmodifier.RequiresReplace(),
																									},
																								},
																								"inclusion": schema.StringAttribute{
																									CustomType: fwtypes.StringEnumType[awstypes.IncludeExclude](),
																									Optional:   true,
																									PlanModifiers: []planmodifier.String{
																										stringplanmodifier.RequiresReplace(),
																									},
																								},
																								names.AttrName: schema.StringAttribute{
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
												"storage_configuration": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[redshiftQueryEngineStorageConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.IsRequired(),
														listvalidator.SizeAtLeast(1),
														listvalidator.SizeAtMost(1),
													},
													PlanModifiers: []planmodifier.List{
														listplanmodifier.RequiresReplace(),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrType: schema.StringAttribute{
																CustomType: fwtypes.StringEnumType[awstypes.RedshiftQueryEngineStorageType](),
																Required:   true,
																PlanModifiers: []planmodifier.String{
																	stringplanmodifier.RequiresReplace(),
																},
															},
														},
														Blocks: map[string]schema.Block{
															"aws_data_catalog_configuration": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[redshiftQueryEngineAWSDataCatalogStorageConfigurationModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																	listvalidator.ExactlyOneOf(
																		path.MatchRelative().AtParent().AtName("aws_data_catalog_configuration"),
																		path.MatchRelative().AtParent().AtName("redshift_configuration"),
																	),
																},
																PlanModifiers: []planmodifier.List{
																	listplanmodifier.RequiresReplace(),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"table_names": schema.SetAttribute{
																			CustomType: fwtypes.SetOfStringType,
																			Required:   true,
																			PlanModifiers: []planmodifier.Set{
																				setplanmodifier.RequiresReplace(),
																			},
																		},
																	},
																},
															},
															"redshift_configuration": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[redshiftQueryEngineRedshiftStorageConfigurationModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
																PlanModifiers: []planmodifier.List{
																	listplanmodifier.RequiresReplace(),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		names.AttrDatabaseName: schema.StringAttribute{
																			Required: true,
																			Validators: []validator.String{
																				stringvalidator.LengthBetween(1, 200),
																			},
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
						"vector_knowledge_base_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[vectorKnowledgeBaseConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"embedding_model_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"embedding_model_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[embeddingModelConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"bedrock_embedding_model_configuration": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[bedrockEmbeddingModelConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													PlanModifiers: []planmodifier.List{
														listplanmodifier.RequiresReplace(),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"dimensions": schema.Int64Attribute{
																Optional: true,
																PlanModifiers: []planmodifier.Int64{
																	int64planmodifier.RequiresReplace(),
																},
															},
															"embedding_data_type": schema.StringAttribute{
																CustomType: fwtypes.StringEnumType[awstypes.EmbeddingDataType](),
																Optional:   true,
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
									"supplemental_data_storage_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[supplementalDataStorageConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"storage_location": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[storageLocationModel](ctx),
													Validators: []validator.List{
														listvalidator.IsRequired(),
														listvalidator.SizeAtLeast(1),
														listvalidator.SizeAtMost(1),
													},
													PlanModifiers: []planmodifier.List{
														listplanmodifier.RequiresReplace(),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrType: schema.StringAttribute{
																CustomType: fwtypes.StringEnumType[awstypes.SupplementalDataStorageLocationType](),
																Required:   true,
																PlanModifiers: []planmodifier.String{
																	stringplanmodifier.RequiresReplace(),
																},
															},
														},
														Blocks: map[string]schema.Block{
															"s3_location": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[s3LocationModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
																PlanModifiers: []planmodifier.List{
																	listplanmodifier.RequiresReplace(),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		names.AttrURI: schema.StringAttribute{
																			Required: true,
																			Validators: []validator.String{
																				stringvalidator.RegexMatches(
																					regexache.MustCompile(`^s3://[a-z0-9.-]+(/.*)?$`),
																					"must be a valid S3 URI",
																				),
																			},
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
				},
			},
			"storage_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[storageConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrType: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.KnowledgeBaseStorageType](),
							Required:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"mongo_db_atlas_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[mongoDBAtlasConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("mongo_db_atlas_configuration"),
									path.MatchRelative().AtParent().AtName("neptune_analytics_configuration"),
									path.MatchRelative().AtParent().AtName("opensearch_managed_cluster_configuration"),
									path.MatchRelative().AtParent().AtName("opensearch_serverless_configuration"),
									path.MatchRelative().AtParent().AtName("pinecone_configuration"),
									path.MatchRelative().AtParent().AtName("rds_configuration"),
									path.MatchRelative().AtParent().AtName("redis_enterprise_cloud_configuration"),
									path.MatchRelative().AtParent().AtName("s3_vectors_configuration"),
								),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"collection_name": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"credentials_secret_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									names.AttrDatabaseName: schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									names.AttrEndpoint: schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"endpoint_service_name": schema.StringAttribute{
										Optional: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"text_index_name": schema.StringAttribute{
										Optional: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"vector_index_name": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"field_mapping": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[mongoDBAtlasFieldMappingModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtLeast(1),
											listvalidator.SizeAtMost(1),
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"metadata_field": schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"text_field": schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"vector_field": schema.StringAttribute{
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
						"neptune_analytics_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[neptuneAnalyticsConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"graph_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"field_mapping": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[neptuneAnalyticsFieldMappingModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtLeast(1),
											listvalidator.SizeAtMost(1),
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"metadata_field": schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"text_field": schema.StringAttribute{
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
						"opensearch_managed_cluster_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[openSearchManagedClusterConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"domain_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"domain_endpoint": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.RegexMatches(
												regexache.MustCompile(`^https://.*$`),
												"must be a valid HTTPS URL",
											),
										},
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"vector_index_name": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"field_mapping": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[opensearchManagedClusterFieldMappingModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtLeast(1),
											listvalidator.SizeAtMost(1),
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"metadata_field": schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"text_field": schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"vector_field": schema.StringAttribute{
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
						"opensearch_serverless_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[openSearchServerlessConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"collection_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"vector_index_name": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"field_mapping": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[openSearchServerlessFieldMappingModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtLeast(1),
											listvalidator.SizeAtMost(1),
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"metadata_field": schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"text_field": schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"vector_field": schema.StringAttribute{
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
						"pinecone_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[pineconeConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"connection_string": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"credentials_secret_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									names.AttrNamespace: schema.StringAttribute{
										Optional: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"field_mapping": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[pineconeFieldMappingModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtLeast(1),
											listvalidator.SizeAtMost(1),
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"metadata_field": schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"text_field": schema.StringAttribute{
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
						"rds_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[rdsConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"credentials_secret_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									names.AttrDatabaseName: schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									names.AttrResourceARN: schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									names.AttrTableName: schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"field_mapping": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[rdsFieldMappingModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtLeast(1),
											listvalidator.SizeAtMost(1),
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"custom_metadata_field": schema.StringAttribute{
													Optional: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"metadata_field": schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"primary_key_field": schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"text_field": schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"vector_field": schema.StringAttribute{
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
						"redis_enterprise_cloud_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[redisEnterpriseCloudConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"credentials_secret_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									names.AttrEndpoint: schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"vector_index_name": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"field_mapping": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[redisEnterpriseCloudFieldMappingModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtLeast(1),
											listvalidator.SizeAtMost(1),
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"metadata_field": schema.StringAttribute{
													Optional: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"text_field": schema.StringAttribute{
													Optional: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"vector_field": schema.StringAttribute{
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
						"s3_vectors_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[s3VectorsConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"index_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Optional:   true,
										Validators: []validator.String{
											stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("index_name"), path.MatchRelative().AtParent().AtName("vector_bucket_arn")),
										},
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"index_name": schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("vector_bucket_arn")),
											stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("index_arn")),
										},
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"vector_bucket_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Optional:   true,
										Validators: []validator.String{
											stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("index_name")),
											stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("index_arn")),
										},
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
								Validators: []validator.Object{
									objectvalidator.AtLeastOneOf(path.MatchRelative().AtName("index_arn"), path.MatchRelative().AtName("index_name")),
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

func (r *knowledgeBaseResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data knowledgeBaseResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	var input bedrockagent.CreateKnowledgeBaseInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(id.UniqueId())
	input.Tags = getTagsIn(ctx)

	var output *bedrockagent.CreateKnowledgeBaseOutput
	var err error
	err = tfresource.Retry(ctx, propagationTimeout, func(ctx context.Context) *tfresource.RetryError {
		output, err = conn.CreateKnowledgeBase(ctx, &input)

		// IAM propagation
		if tfawserr.ErrMessageContains(err, errCodeValidationException, "cannot assume role") {
			return tfresource.RetryableError(err)
		}
		if tfawserr.ErrMessageContains(err, errCodeValidationException, "unable to assume the given role") {
			return tfresource.RetryableError(err)
		}

		// Kendra access propagation
		if tfawserr.ErrMessageContains(err, errCodeValidationException, "Encountered AccessDeniedException from Kendra") {
			return tfresource.RetryableError(err)
		}

		// OpenSearch data access propagation
		if tfawserr.ErrMessageContains(err, errCodeValidationException, "storage configuration provided is invalid") {
			return tfresource.RetryableError(err)
		}

		if err != nil {
			return tfresource.NonRetryableError(err)
		}

		return nil
	})

	if err != nil {
		response.Diagnostics.AddError("creating Bedrock Agent Knowledge Base", err.Error())
		return
	}

	kb := output.KnowledgeBase
	knowledgeBaseID := aws.ToString(kb.KnowledgeBaseId)
	data.KnowledgeBaseARN = fwflex.StringToFramework(ctx, kb.KnowledgeBaseArn)
	data.KnowledgeBaseID = fwflex.StringValueToFramework(ctx, knowledgeBaseID)

	kb, err = waitKnowledgeBaseCreated(ctx, conn, knowledgeBaseID, r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Bedrock Agent Knowledge Base (%s) create", knowledgeBaseID), err.Error())
		return
	}

	// Set values for unknowns after creation is complete.
	data.CreatedAt = fwflex.TimeToFramework(ctx, kb.CreatedAt)
	data.FailureReasons = fwflex.FlattenFrameworkStringValueListOfString(ctx, kb.FailureReasons)
	data.UpdatedAt = fwflex.TimeToFramework(ctx, kb.UpdatedAt)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *knowledgeBaseResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data knowledgeBaseResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	knowledgeBaseID := fwflex.StringValueFromFramework(ctx, data.KnowledgeBaseID)
	kb, err := findKnowledgeBaseByID(ctx, conn, knowledgeBaseID)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Agent Knowledge Base (%s)", knowledgeBaseID), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, kb, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *knowledgeBaseResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new knowledgeBaseResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	if !new.Description.Equal(old.Description) ||
		!new.Name.Equal(old.Name) ||
		!new.RoleARN.Equal(old.RoleARN) {
		knowledgeBaseID := fwflex.StringValueFromFramework(ctx, new.KnowledgeBaseID)
		var input bedrockagent.UpdateKnowledgeBaseInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func(ctx context.Context) (any, error) {
			return conn.UpdateKnowledgeBase(ctx, &input)
		}, errCodeValidationException, "cannot assume role")

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Bedrock Agent Knowledge Base (%s)", knowledgeBaseID), err.Error())

			return
		}

		kb, err := waitKnowledgeBaseUpdated(ctx, conn, knowledgeBaseID, r.UpdateTimeout(ctx, new.Timeouts))

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Bedrock Agent Knowledge Base (%s) update", knowledgeBaseID), err.Error())

			return
		}

		new.FailureReasons = fwflex.FlattenFrameworkStringValueListOfString(ctx, kb.FailureReasons)
		new.UpdatedAt = fwflex.TimeToFramework(ctx, kb.UpdatedAt)
	} else {
		new.FailureReasons = old.FailureReasons
		new.UpdatedAt = old.UpdatedAt
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *knowledgeBaseResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data knowledgeBaseResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	knowledgeBaseID := fwflex.StringValueFromFramework(ctx, data.KnowledgeBaseID)
	input := bedrockagent.DeleteKnowledgeBaseInput{
		KnowledgeBaseId: aws.String(knowledgeBaseID),
	}
	_, err := conn.DeleteKnowledgeBase(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Bedrock Agent Knowledge Base (%s)", knowledgeBaseID), err.Error())

		return
	}

	if _, err := waitKnowledgeBaseDeleted(ctx, conn, knowledgeBaseID, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Bedrock Agent Knowledge Base (%s) delete", knowledgeBaseID), err.Error())

		return
	}
}

func waitKnowledgeBaseCreated(ctx context.Context, conn *bedrockagent.Client, id string, timeout time.Duration) (*awstypes.KnowledgeBase, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.KnowledgeBaseStatusCreating),
		Target:  enum.Slice(awstypes.KnowledgeBaseStatusActive),
		Refresh: statusKnowledgeBase(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.KnowledgeBase); ok {
		retry.SetLastError(err, errors.Join(tfslices.ApplyToAll(output.FailureReasons, errors.New)...))

		return output, err
	}

	return nil, err
}

func waitKnowledgeBaseUpdated(ctx context.Context, conn *bedrockagent.Client, id string, timeout time.Duration) (*awstypes.KnowledgeBase, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.KnowledgeBaseStatusUpdating),
		Target:  enum.Slice(awstypes.KnowledgeBaseStatusActive),
		Refresh: statusKnowledgeBase(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.KnowledgeBase); ok {
		retry.SetLastError(err, errors.Join(tfslices.ApplyToAll(output.FailureReasons, errors.New)...))

		return output, err
	}

	return nil, err
}

func waitKnowledgeBaseDeleted(ctx context.Context, conn *bedrockagent.Client, id string, timeout time.Duration) (*awstypes.KnowledgeBase, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.KnowledgeBaseStatusActive, awstypes.KnowledgeBaseStatusDeleting),
		Target:  []string{},
		Refresh: statusKnowledgeBase(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.KnowledgeBase); ok {
		retry.SetLastError(err, errors.Join(tfslices.ApplyToAll(output.FailureReasons, errors.New)...))

		return output, err
	}

	return nil, err
}

func statusKnowledgeBase(conn *bedrockagent.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findKnowledgeBaseByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func findKnowledgeBaseByID(ctx context.Context, conn *bedrockagent.Client, id string) (*awstypes.KnowledgeBase, error) {
	input := bedrockagent.GetKnowledgeBaseInput{
		KnowledgeBaseId: aws.String(id),
	}

	return findKnowledgeBase(ctx, conn, &input)
}

func findKnowledgeBase(ctx context.Context, conn *bedrockagent.Client, input *bedrockagent.GetKnowledgeBaseInput) (*awstypes.KnowledgeBase, error) {
	output, err := conn.GetKnowledgeBase(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.KnowledgeBase == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.KnowledgeBase, nil
}

type knowledgeBaseResourceModel struct {
	framework.WithRegionModel
	CreatedAt                  timetypes.RFC3339                                                `tfsdk:"created_at"`
	Description                types.String                                                     `tfsdk:"description"`
	FailureReasons             fwtypes.ListValueOf[types.String]                                `tfsdk:"failure_reasons"`
	KnowledgeBaseARN           types.String                                                     `tfsdk:"arn"`
	KnowledgeBaseConfiguration fwtypes.ListNestedObjectValueOf[knowledgeBaseConfigurationModel] `tfsdk:"knowledge_base_configuration"`
	KnowledgeBaseID            types.String                                                     `tfsdk:"id"`
	Name                       types.String                                                     `tfsdk:"name"`
	RoleARN                    fwtypes.ARN                                                      `tfsdk:"role_arn"`
	StorageConfiguration       fwtypes.ListNestedObjectValueOf[storageConfigurationModel]       `tfsdk:"storage_configuration"`
	Tags                       tftags.Map                                                       `tfsdk:"tags"`
	TagsAll                    tftags.Map                                                       `tfsdk:"tags_all"`
	Timeouts                   timeouts.Value                                                   `tfsdk:"timeouts"`
	UpdatedAt                  timetypes.RFC3339                                                `tfsdk:"updated_at"`
}

type knowledgeBaseConfigurationModel struct {
	KendraKnowledgeBaseConfiguration fwtypes.ListNestedObjectValueOf[kendraKnowledgeBaseConfigurationModel] `tfsdk:"kendra_knowledge_base_configuration"`
	SQLKnowledgeBaseConfiguration    fwtypes.ListNestedObjectValueOf[sqlKnowledgeBaseConfigurationModel]    `tfsdk:"sql_knowledge_base_configuration"`
	Type                             fwtypes.StringEnum[awstypes.KnowledgeBaseType]                         `tfsdk:"type"`
	VectorKnowledgeBaseConfiguration fwtypes.ListNestedObjectValueOf[vectorKnowledgeBaseConfigurationModel] `tfsdk:"vector_knowledge_base_configuration"`
}

type kendraKnowledgeBaseConfigurationModel struct {
	KendraIndexARN fwtypes.ARN `tfsdk:"kendra_index_arn"`
}

type sqlKnowledgeBaseConfigurationModel struct {
	RedshiftConfiguration fwtypes.ListNestedObjectValueOf[redshiftConfigurationModel] `tfsdk:"redshift_configuration"`
	Type                  fwtypes.StringEnum[awstypes.QueryEngineType]                `tfsdk:"type"`
}

type redshiftConfigurationModel struct {
	QueryEngineConfiguration     fwtypes.ListNestedObjectValueOf[redshiftQueryEngineConfigurationModel]        `tfsdk:"query_engine_configuration"`
	QueryGenerationConfiguration fwtypes.ListNestedObjectValueOf[queryGenerationConfigurationModel]            `tfsdk:"query_generation_configuration"`
	StorageConfigurations        fwtypes.ListNestedObjectValueOf[redshiftQueryEngineStorageConfigurationModel] `tfsdk:"storage_configuration"`
}

type redshiftQueryEngineConfigurationModel struct {
	ProvisionedConfiguration fwtypes.ListNestedObjectValueOf[redshiftProvisionedConfigurationModel] `tfsdk:"provisioned_configuration"`
	ServerlessConfiguration  fwtypes.ListNestedObjectValueOf[redshiftServerlessConfigurationModel]  `tfsdk:"serverless_configuration"`
	Type                     fwtypes.StringEnum[awstypes.RedshiftQueryEngineType]                   `tfsdk:"type"`
}

type redshiftProvisionedConfigurationModel struct {
	AuthConfiguration fwtypes.ListNestedObjectValueOf[redshiftProvisionedAuthConfigurationModel] `tfsdk:"auth_configuration"`
	ClusterIdentifier types.String                                                               `tfsdk:"cluster_identifier"`
}

type redshiftProvisionedAuthConfigurationModel struct {
	DatabaseUser              types.String                                             `tfsdk:"database_user"`
	Type                      fwtypes.StringEnum[awstypes.RedshiftProvisionedAuthType] `tfsdk:"type"`
	UsernamePasswordSecretARN fwtypes.ARN                                              `tfsdk:"username_password_secret_arn"`
}

type redshiftServerlessConfigurationModel struct {
	AuthConfiguration fwtypes.ListNestedObjectValueOf[redshiftServerlessAuthConfigurationModel] `tfsdk:"auth_configuration"`
	WorkgroupARN      fwtypes.ARN                                                               `tfsdk:"workgroup_arn"`
}

type redshiftServerlessAuthConfigurationModel struct {
	Type                      fwtypes.StringEnum[awstypes.RedshiftServerlessAuthType] `tfsdk:"type"`
	UsernamePasswordSecretARN fwtypes.ARN                                             `tfsdk:"username_password_secret_arn"`
}

type queryGenerationConfigurationModel struct {
	ExecutionTimeoutSeconds types.Int64                                                  `tfsdk:"execution_timeout_seconds"`
	GenerationContext       fwtypes.ListNestedObjectValueOf[queryGenerationContextModel] `tfsdk:"generation_context"`
}

type queryGenerationContextModel struct {
	CuratedQueries fwtypes.ListNestedObjectValueOf[curatedQueryModel]         `tfsdk:"curated_query"`
	Tables         fwtypes.ListNestedObjectValueOf[queryGenerationTableModel] `tfsdk:"table"`
}

type curatedQueryModel struct {
	NaturalLanguage types.String `tfsdk:"natural_language"`
	SQL             types.String `tfsdk:"sql"`
}

type queryGenerationTableModel struct {
	Columns     fwtypes.ListNestedObjectValueOf[queryGenerationColumnModel] `tfsdk:"column"`
	Description types.String                                                `tfsdk:"description"`
	Inclusion   fwtypes.StringEnum[awstypes.IncludeExclude]                 `tfsdk:"inclusion"`
	Name        types.String                                                `tfsdk:"name"`
}

type queryGenerationColumnModel struct {
	Description types.String                                `tfsdk:"description"`
	Inclusion   fwtypes.StringEnum[awstypes.IncludeExclude] `tfsdk:"inclusion"`
	Name        types.String                                `tfsdk:"name"`
}

type redshiftQueryEngineStorageConfigurationModel struct {
	AWSDataCatalogConfiguration fwtypes.ListNestedObjectValueOf[redshiftQueryEngineAWSDataCatalogStorageConfigurationModel] `tfsdk:"aws_data_catalog_configuration"`
	RedshiftConfiguration       fwtypes.ListNestedObjectValueOf[redshiftQueryEngineRedshiftStorageConfigurationModel]       `tfsdk:"redshift_configuration"`
	Type                        fwtypes.StringEnum[awstypes.RedshiftQueryEngineStorageType]                                 `tfsdk:"type"`
}

type redshiftQueryEngineAWSDataCatalogStorageConfigurationModel struct {
	TableNames fwtypes.SetOfString `tfsdk:"table_names"`
}

type redshiftQueryEngineRedshiftStorageConfigurationModel struct {
	DatabaseName types.String `tfsdk:"database_name"`
}

type vectorKnowledgeBaseConfigurationModel struct {
	EmbeddingModelARN                    fwtypes.ARN                                                                `tfsdk:"embedding_model_arn"`
	EmbeddingModelConfiguration          fwtypes.ListNestedObjectValueOf[embeddingModelConfigurationModel]          `tfsdk:"embedding_model_configuration"`
	SupplementalDataStorageConfiguration fwtypes.ListNestedObjectValueOf[supplementalDataStorageConfigurationModel] `tfsdk:"supplemental_data_storage_configuration"`
}

type embeddingModelConfigurationModel struct {
	BedrockEmbeddingModelConfiguration fwtypes.ListNestedObjectValueOf[bedrockEmbeddingModelConfigurationModel] `tfsdk:"bedrock_embedding_model_configuration"`
}

type bedrockEmbeddingModelConfigurationModel struct {
	Dimensions        types.Int64                                    `tfsdk:"dimensions"`
	EmbeddingDataType fwtypes.StringEnum[awstypes.EmbeddingDataType] `tfsdk:"embedding_data_type"`
}

type supplementalDataStorageConfigurationModel struct {
	StorageLocations fwtypes.ListNestedObjectValueOf[storageLocationModel] `tfsdk:"storage_location"`
}

type storageLocationModel struct {
	S3Location fwtypes.ListNestedObjectValueOf[s3LocationModel]                 `tfsdk:"s3_location"`
	Type       fwtypes.StringEnum[awstypes.SupplementalDataStorageLocationType] `tfsdk:"type"`
}

type storageConfigurationModel struct {
	MongoDBAtlasConfiguration             fwtypes.ListNestedObjectValueOf[mongoDBAtlasConfigurationModel]             `tfsdk:"mongo_db_atlas_configuration"`
	NeptuneAnalyticsConfiguration         fwtypes.ListNestedObjectValueOf[neptuneAnalyticsConfigurationModel]         `tfsdk:"neptune_analytics_configuration"`
	OpenSearchManagedClusterConfiguration fwtypes.ListNestedObjectValueOf[openSearchManagedClusterConfigurationModel] `tfsdk:"opensearch_managed_cluster_configuration"`
	OpenSearchServerlessConfiguration     fwtypes.ListNestedObjectValueOf[openSearchServerlessConfigurationModel]     `tfsdk:"opensearch_serverless_configuration"`
	PineconeConfiguration                 fwtypes.ListNestedObjectValueOf[pineconeConfigurationModel]                 `tfsdk:"pinecone_configuration"`
	RDSConfiguration                      fwtypes.ListNestedObjectValueOf[rdsConfigurationModel]                      `tfsdk:"rds_configuration"`
	RedisEnterpriseCloudConfiguration     fwtypes.ListNestedObjectValueOf[redisEnterpriseCloudConfigurationModel]     `tfsdk:"redis_enterprise_cloud_configuration"`
	S3VectorsConfiguration                fwtypes.ListNestedObjectValueOf[s3VectorsConfigurationModel]                `tfsdk:"s3_vectors_configuration"`
	Type                                  fwtypes.StringEnum[awstypes.KnowledgeBaseStorageType]                       `tfsdk:"type"`
}

type mongoDBAtlasConfigurationModel struct {
	CollectionName       types.String                                                   `tfsdk:"collection_name"`
	CredentialsSecretARN fwtypes.ARN                                                    `tfsdk:"credentials_secret_arn"`
	DatabaseName         types.String                                                   `tfsdk:"database_name"`
	Endpoint             types.String                                                   `tfsdk:"endpoint"`
	EndpointServiceName  types.String                                                   `tfsdk:"endpoint_service_name"`
	FieldMapping         fwtypes.ListNestedObjectValueOf[mongoDBAtlasFieldMappingModel] `tfsdk:"field_mapping"`
	TextIndexName        types.String                                                   `tfsdk:"text_index_name"`
	VectorIndexName      types.String                                                   `tfsdk:"vector_index_name"`
}

type mongoDBAtlasFieldMappingModel struct {
	MetadataField types.String `tfsdk:"metadata_field"`
	TextField     types.String `tfsdk:"text_field"`
	VectorField   types.String `tfsdk:"vector_field"`
}

type neptuneAnalyticsConfigurationModel struct {
	FieldMapping fwtypes.ListNestedObjectValueOf[neptuneAnalyticsFieldMappingModel] `tfsdk:"field_mapping"`
	GraphARN     fwtypes.ARN                                                        `tfsdk:"graph_arn"`
}

type neptuneAnalyticsFieldMappingModel struct {
	MetadataField types.String `tfsdk:"metadata_field"`
	TextField     types.String `tfsdk:"text_field"`
}

type openSearchManagedClusterConfigurationModel struct {
	DomainARN       fwtypes.ARN                                                                `tfsdk:"domain_arn"`
	DomainEndpoint  types.String                                                               `tfsdk:"domain_endpoint"`
	FieldMapping    fwtypes.ListNestedObjectValueOf[opensearchManagedClusterFieldMappingModel] `tfsdk:"field_mapping"`
	VectorIndexName types.String                                                               `tfsdk:"vector_index_name"`
}

type opensearchManagedClusterFieldMappingModel struct {
	MetadataField types.String `tfsdk:"metadata_field"`
	TextField     types.String `tfsdk:"text_field"`
	VectorField   types.String `tfsdk:"vector_field"`
}

type openSearchServerlessConfigurationModel struct {
	CollectionARN   fwtypes.ARN                                                            `tfsdk:"collection_arn"`
	FieldMapping    fwtypes.ListNestedObjectValueOf[openSearchServerlessFieldMappingModel] `tfsdk:"field_mapping"`
	VectorIndexName types.String                                                           `tfsdk:"vector_index_name"`
}

type openSearchServerlessFieldMappingModel struct {
	MetadataField types.String `tfsdk:"metadata_field"`
	TextField     types.String `tfsdk:"text_field"`
	VectorField   types.String `tfsdk:"vector_field"`
}

type pineconeConfigurationModel struct {
	ConnectionString     types.String                                               `tfsdk:"connection_string"`
	CredentialsSecretARN fwtypes.ARN                                                `tfsdk:"credentials_secret_arn"`
	FieldMapping         fwtypes.ListNestedObjectValueOf[pineconeFieldMappingModel] `tfsdk:"field_mapping"`
	Namespace            types.String                                               `tfsdk:"namespace"`
}

type pineconeFieldMappingModel struct {
	MetadataField types.String `tfsdk:"metadata_field"`
	TextField     types.String `tfsdk:"text_field"`
}

type rdsConfigurationModel struct {
	CredentialsSecretARN fwtypes.ARN                                           `tfsdk:"credentials_secret_arn"`
	DatabaseName         types.String                                          `tfsdk:"database_name"`
	FieldMapping         fwtypes.ListNestedObjectValueOf[rdsFieldMappingModel] `tfsdk:"field_mapping"`
	ResourceARN          fwtypes.ARN                                           `tfsdk:"resource_arn"`
	TableName            types.String                                          `tfsdk:"table_name"`
}

type rdsFieldMappingModel struct {
	CustomMetadataField types.String `tfsdk:"custom_metadata_field"`
	MetadataField       types.String `tfsdk:"metadata_field"`
	PrimaryKeyField     types.String `tfsdk:"primary_key_field"`
	TextField           types.String `tfsdk:"text_field"`
	VectorField         types.String `tfsdk:"vector_field"`
}

type redisEnterpriseCloudConfigurationModel struct {
	CredentialsSecretARN fwtypes.ARN                                                            `tfsdk:"credentials_secret_arn"`
	Endpoint             types.String                                                           `tfsdk:"endpoint"`
	FieldMapping         fwtypes.ListNestedObjectValueOf[redisEnterpriseCloudFieldMappingModel] `tfsdk:"field_mapping"`
	VectorIndexName      types.String                                                           `tfsdk:"vector_index_name"`
}

type redisEnterpriseCloudFieldMappingModel struct {
	MetadataField types.String `tfsdk:"metadata_field"`
	TextField     types.String `tfsdk:"text_field"`
	VectorField   types.String `tfsdk:"vector_field"`
}

type s3VectorsConfigurationModel struct {
	IndexARN        fwtypes.ARN  `tfsdk:"index_arn"`
	IndexName       types.String `tfsdk:"index_name"`
	VectorBucketARN fwtypes.ARN  `tfsdk:"vector_bucket_arn"`
}
