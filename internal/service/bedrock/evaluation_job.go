// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/float32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrock_evaluation_job", name="Evaluation Job")
// @Tags(identifierAttribute="job_arn")
// @ArnIdentity("job_arn")
// @Testing(preCheck="testAccPreCheckEvaluationJob")
// @Testing(hasNoPreExistingResource=true)
// @Testing(checkDestroyNoop=true)
func newEvaluationJobResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &evaluationJobResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type evaluationJobResource struct {
	framework.ResourceWithModel[evaluationJobResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *evaluationJobResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"application_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ApplicationType](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"customer_encryption_key_id": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 2048),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"failure_messages": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"job_arn": framework.ARNAttributeComputedOnly(),
			"job_description": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 200),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"job_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-z0-9](-*[a-z0-9])*$`),
						"must be up to 63 lowercase letters, numbers, and dashes, and must start and end with a letter or number"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"job_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EvaluationJobType](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_modified_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
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
			names.AttrSkipDestroy: schema.BoolAttribute{
				Optional: true,
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EvaluationJobStatus](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"evaluation_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[evaluationConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"automated": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[automatedEvaluationConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("automated"),
									path.MatchRelative().AtParent().AtName("human"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"dataset_metric_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[evaluationDatasetMetricConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeBetween(1, 5),
										},
										NestedObject: evaluationDatasetMetricConfigNestedObject(ctx),
									},
									"evaluator_model_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[evaluatorModelConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"bedrock_evaluator_model": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[bedrockEvaluatorModelModel](ctx),
													Validators: []validator.List{
														listvalidator.IsRequired(),
														listvalidator.SizeBetween(1, 1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"model_identifier": schema.StringAttribute{
																Required: true,
																Validators: []validator.String{
																	stringvalidator.LengthBetween(1, 2048),
																},
															},
														},
													},
												},
											},
										},
									},
									"custom_metric_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[automatedEvaluationCustomMetricConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"custom_metric": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[automatedEvaluationCustomMetricSourceModel](ctx),
													Validators: []validator.List{
														listvalidator.IsRequired(),
														listvalidator.SizeBetween(1, 10),
													},
													NestedObject: schema.NestedBlockObject{
														Blocks: map[string]schema.Block{
															"custom_metric_definition": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[customMetricDefinitionModel](ctx),
																Validators: []validator.List{
																	listvalidator.IsRequired(),
																	listvalidator.SizeBetween(1, 1),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"instructions": schema.StringAttribute{
																			Required: true,
																			Validators: []validator.String{
																				stringvalidator.LengthBetween(1, 5000),
																			},
																		},
																		names.AttrName: schema.StringAttribute{
																			Required: true,
																			Validators: []validator.String{
																				stringvalidator.LengthBetween(1, 63),
																				stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9a-zA-Z\-_.]+$`),
																					"must be alphanumeric, with hyphens, underscores, and periods"),
																			},
																		},
																	},
																	Blocks: map[string]schema.Block{
																		"rating_scale": schema.ListNestedBlock{
																			CustomType: fwtypes.NewListNestedObjectTypeOf[ratingScaleItemModel](ctx),
																			Validators: []validator.List{
																				listvalidator.SizeBetween(1, 10),
																			},
																			NestedObject: schema.NestedBlockObject{
																				Attributes: map[string]schema.Attribute{
																					"definition": schema.StringAttribute{
																						Required: true,
																						Validators: []validator.String{
																							stringvalidator.LengthBetween(1, 100),
																						},
																					},
																				},
																				Blocks: map[string]schema.Block{
																					names.AttrValue: schema.ListNestedBlock{
																						CustomType: fwtypes.NewListNestedObjectTypeOf[ratingScaleItemValueModel](ctx),
																						Validators: []validator.List{
																							listvalidator.IsRequired(),
																							listvalidator.SizeBetween(1, 1),
																						},
																						NestedObject: schema.NestedBlockObject{
																							Attributes: map[string]schema.Attribute{
																								"float_value": schema.Float32Attribute{
																									Optional: true,
																									Validators: []validator.Float32{
																										float32validator.ExactlyOneOf(
																											path.MatchRelative().AtParent().AtName("float_value"),
																											path.MatchRelative().AtParent().AtName("string_value"),
																										),
																									},
																								},
																								"string_value": schema.StringAttribute{
																									Optional: true,
																									Validators: []validator.String{
																										stringvalidator.LengthBetween(1, 100),
																										stringvalidator.ExactlyOneOf(
																											path.MatchRelative().AtParent().AtName("float_value"),
																											path.MatchRelative().AtParent().AtName("string_value"),
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
													},
												},
												"evaluator_model_config": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[customMetricEvaluatorModelConfigModel](ctx),
													Validators: []validator.List{
														listvalidator.IsRequired(),
														listvalidator.SizeBetween(1, 1),
													},
													NestedObject: schema.NestedBlockObject{
														Blocks: map[string]schema.Block{
															"bedrock_evaluator_model": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[customMetricBedrockEvaluatorModelModel](ctx),
																Validators: []validator.List{
																	listvalidator.IsRequired(),
																	listvalidator.SizeBetween(1, 1),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"model_identifier": schema.StringAttribute{
																			Required: true,
																			Validators: []validator.String{
																				stringvalidator.LengthBetween(1, 2048),
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
						"human": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[humanEvaluationConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("automated"),
									path.MatchRelative().AtParent().AtName("human"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"custom_metric": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[humanEvaluationCustomMetricModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(10),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrDescription: schema.StringAttribute{
													Optional: true,
												},
												names.AttrName: schema.StringAttribute{
													Required: true,
												},
												"rating_method": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.OneOf(
															"ThumbsUpDown",
															"IndividualLikertScale",
															"ComparisonLikertScale",
															"ComparisonChoice",
															"ComparisonRank",
														),
													},
												},
											},
										},
									},
									"dataset_metric_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[evaluationDatasetMetricConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeBetween(1, 5),
										},
										NestedObject: evaluationDatasetMetricConfigNestedObject(ctx),
									},
									"human_workflow_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[humanWorkflowConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"flow_definition_arn": schema.StringAttribute{
													CustomType: fwtypes.ARNType,
													Required:   true,
												},
												"instructions": schema.StringAttribute{
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
			"inference_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[inferenceConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"model": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[evaluationModelConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(2),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("model"),
									path.MatchRelative().AtParent().AtName("rag_config"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"bedrock_model": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[evaluationBedrockModelModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
											listvalidator.ExactlyOneOf(
												path.MatchRelative().AtParent().AtName("bedrock_model"),
												path.MatchRelative().AtParent().AtName("precomputed_inference_source"),
											),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"inference_params": schema.StringAttribute{
													Optional: true,
												},
												"model_identifier": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 2048),
													},
												},
											},
											Blocks: map[string]schema.Block{
												"performance_config": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[performanceConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"latency": schema.StringAttribute{
																CustomType: fwtypes.StringEnumType[awstypes.PerformanceConfigLatency](),
																Optional:   true,
															},
														},
													},
												},
											},
										},
									},
									"precomputed_inference_source": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[evaluationPrecomputedInferenceSourceModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
											listvalidator.ExactlyOneOf(
												path.MatchRelative().AtParent().AtName("bedrock_model"),
												path.MatchRelative().AtParent().AtName("precomputed_inference_source"),
											),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"inference_source_identifier": schema.StringAttribute{
													Required: true,
												},
											},
										},
									},
								},
							},
						},
						"rag_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[ragConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("model"),
									path.MatchRelative().AtParent().AtName("rag_config"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"knowledge_base_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[knowledgeBaseConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
											listvalidator.ExactlyOneOf(
												path.MatchRelative().AtParent().AtName("knowledge_base_config"),
												path.MatchRelative().AtParent().AtName("precomputed_rag_source_config"),
											),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"retrieve_and_generate_config": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[knowledgeBaseRetrieveAndGenerateConfigModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
														listvalidator.ExactlyOneOf(
															path.MatchRelative().AtParent().AtName("retrieve_and_generate_config"),
															path.MatchRelative().AtParent().AtName("retrieve_config"),
														),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"knowledge_base_id": schema.StringAttribute{
																Required: true,
															},
															"model_arn": schema.StringAttribute{
																CustomType: fwtypes.ARNType,
																Required:   true,
															},
														},
														Blocks: map[string]schema.Block{
															"retrieval_configuration": knowledgeBaseRetrievalConfigurationNestedBlock(ctx),
														},
													},
												},
												"retrieve_config": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[knowledgeBaseRetrieveConfigModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
														listvalidator.ExactlyOneOf(
															path.MatchRelative().AtParent().AtName("retrieve_and_generate_config"),
															path.MatchRelative().AtParent().AtName("retrieve_config"),
														),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"knowledge_base_id": schema.StringAttribute{
																Required: true,
															},
														},
														Blocks: map[string]schema.Block{
															"knowledge_base_retrieval_configuration": knowledgeBaseRetrievalConfigurationNestedBlock(ctx),
														},
													},
												},
											},
										},
									},
									"precomputed_rag_source_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[precomputedRagSourceConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
											listvalidator.ExactlyOneOf(
												path.MatchRelative().AtParent().AtName("knowledge_base_config"),
												path.MatchRelative().AtParent().AtName("precomputed_rag_source_config"),
											),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"retrieve_and_generate_source_config": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[evaluationPrecomputedRetrieveAndGenerateSourceConfigModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
														listvalidator.ExactlyOneOf(
															path.MatchRelative().AtParent().AtName("retrieve_and_generate_source_config"),
															path.MatchRelative().AtParent().AtName("retrieve_source_config"),
														),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"rag_source_identifier": schema.StringAttribute{
																Required: true,
															},
														},
													},
												},
												"retrieve_source_config": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[evaluationPrecomputedRetrieveSourceConfigModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
														listvalidator.ExactlyOneOf(
															path.MatchRelative().AtParent().AtName("retrieve_and_generate_source_config"),
															path.MatchRelative().AtParent().AtName("retrieve_source_config"),
														),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"rag_source_identifier": schema.StringAttribute{
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
			"output_data_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[evaluationOutputDataConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"s3_uri": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								fwvalidators.S3URI(),
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func evaluationDatasetMetricConfigNestedObject(ctx context.Context) schema.NestedBlockObject {
	return schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"metric_names": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 25),
				},
			},
			"task_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EvaluationTaskType](),
				Required:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"dataset": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[evaluationDatasetModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName: schema.StringAttribute{
							Required: true,
						},
					},
					Blocks: map[string]schema.Block{
						"dataset_location": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[evaluationDatasetLocationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"s3_uri": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											fwvalidators.S3URI(),
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

func knowledgeBaseRetrievalConfigurationNestedBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[knowledgeBaseRetrievalConfigurationModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"vector_search_configuration": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[knowledgeBaseVectorSearchConfigurationModel](ctx),
					Validators: []validator.List{
						listvalidator.IsRequired(),
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"number_of_results": schema.Int32Attribute{
								Optional: true,
								Validators: []validator.Int32{
									int32validator.Between(1, 100),
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *evaluationJobResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BedrockClient(ctx)

	var plan evaluationJobResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input bedrock.CreateEvaluationJobInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientRequestToken = aws.String(create.UniqueId(ctx))
	input.JobTags = getTagsIn(ctx)

	var output *bedrock.CreateEvaluationJobOutput
	err := tfresource.Retry(ctx, propagationTimeout, func(ctx context.Context) *tfresource.RetryError {
		var err error
		output, err = conn.CreateEvaluationJob(ctx, &input)

		// IAM role/policy propagation.
		if tfawserr.ErrMessageContains(err, errCodeValidationException, "access denied exception when attempting to assume roleArn") ||
			tfawserr.ErrMessageContains(err, errCodeValidationException, "does not have permission") ||
			tfawserr.ErrMessageContains(err, errCodeValidationException, "access denied message from S3") {
			return tfresource.RetryableError(err)
		}

		if err != nil {
			return tfresource.NonRetryableError(err)
		}

		return nil
	})
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.JobName.ValueString())
		return
	}

	arn := aws.ToString(output.JobArn)

	findOutput, err := tfresource.RetryWhenNewResourceNotFound(ctx, r.CreateTimeout(ctx, plan.Timeouts), func(ctx context.Context) (*bedrock.GetEvaluationJobOutput, error) {
		return findEvaluationJobByARN(ctx, conn, arn)
	}, true)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, findOutput, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *evaluationJobResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BedrockClient(ctx)

	var state evaluationJobResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	arn := state.JobARN.ValueString()
	job, err := findEvaluationJobByARN(ctx, conn, arn)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, job, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *evaluationJobResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state evaluationJobResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	if state.SkipDestroy.ValueBool() {
		return
	}
	// StopEvaluationJob only supports jobs that are In-Progress.
	// All other status values are terminal states.
	if state.Status.ValueEnum() != awstypes.EvaluationJobStatusInProgress {
		return
	}

	arn := state.JobARN.ValueString()

	conn := r.Meta().BedrockClient(ctx)
	input := bedrock.StopEvaluationJobInput{
		JobIdentifier: aws.String(arn),
	}

	_, err := conn.StopEvaluationJob(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	if _, err := waitEvaluationJobStopped(ctx, conn, arn, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}
}

func findEvaluationJobByARN(ctx context.Context, conn *bedrock.Client, arn string) (*bedrock.GetEvaluationJobOutput, error) {
	input := bedrock.GetEvaluationJobInput{
		JobIdentifier: aws.String(arn),
	}

	output, err := conn.GetEvaluationJob(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func statusEvaluationJob(conn *bedrock.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findEvaluationJobByARN(ctx, conn, arn)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitEvaluationJobStopped(ctx context.Context, conn *bedrock.Client, arn string, timeout time.Duration) (*bedrock.GetEvaluationJobOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.EvaluationJobStatusStopping),
		Target:  enum.Slice(awstypes.EvaluationJobStatusStopped),
		Refresh: statusEvaluationJob(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*bedrock.GetEvaluationJobOutput); ok {
		return output, err
	}

	return nil, err
}

type evaluationJobResourceModel struct {
	framework.WithRegionModel
	ApplicationType         fwtypes.StringEnum[awstypes.ApplicationType]                     `tfsdk:"application_type"`
	CreationTime            timetypes.RFC3339                                                `tfsdk:"created_at"`
	CustomerEncryptionKeyID fwtypes.ARN                                                      `tfsdk:"customer_encryption_key_id"`
	EvaluationConfig        fwtypes.ListNestedObjectValueOf[evaluationConfigModel]           `tfsdk:"evaluation_config"`
	FailureMessages         fwtypes.ListOfString                                             `tfsdk:"failure_messages"`
	InferenceConfig         fwtypes.ListNestedObjectValueOf[inferenceConfigModel]            `tfsdk:"inference_config"`
	JobARN                  types.String                                                     `tfsdk:"job_arn"`
	JobDescription          types.String                                                     `tfsdk:"job_description"`
	JobName                 types.String                                                     `tfsdk:"job_name"`
	JobType                 fwtypes.StringEnum[awstypes.EvaluationJobType]                   `tfsdk:"job_type"`
	LastModifiedTime        timetypes.RFC3339                                                `tfsdk:"last_modified_time"`
	OutputDataConfig        fwtypes.ListNestedObjectValueOf[evaluationOutputDataConfigModel] `tfsdk:"output_data_config"`
	RoleARN                 fwtypes.ARN                                                      `tfsdk:"role_arn"`
	SkipDestroy             types.Bool                                                       `tfsdk:"skip_destroy"`
	Status                  fwtypes.StringEnum[awstypes.EvaluationJobStatus]                 `tfsdk:"status"`
	Tags                    tftags.Map                                                       `tfsdk:"tags"`
	TagsAll                 tftags.Map                                                       `tfsdk:"tags_all"`
	Timeouts                timeouts.Value                                                   `tfsdk:"timeouts"`
}

// evaluationConfigModel maps to the awstypes.EvaluationConfig union (Automated or Human).
type evaluationConfigModel struct {
	Automated fwtypes.ListNestedObjectValueOf[automatedEvaluationConfigModel] `tfsdk:"automated"`
	Human     fwtypes.ListNestedObjectValueOf[humanEvaluationConfigModel]     `tfsdk:"human"`
}

var (
	_ fwflex.Expander  = evaluationConfigModel{}
	_ fwflex.Flattener = &evaluationConfigModel{}
)

func (m evaluationConfigModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.Automated.IsNull():
		data, d := m.Automated.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.EvaluationConfigMemberAutomated
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	case !m.Human.IsNull():
		data, d := m.Human.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.EvaluationConfigMemberHuman
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}

	return nil, diags
}

func (m *evaluationConfigModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.EvaluationConfigMemberAutomated:
		var data automatedEvaluationConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.Automated = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	case awstypes.EvaluationConfigMemberHuman:
		var data humanEvaluationConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.Human = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("evaluation config flatten: %T", v))
	}

	return diags
}

type automatedEvaluationConfigModel struct {
	CustomMetricConfig   fwtypes.ListNestedObjectValueOf[automatedEvaluationCustomMetricConfigModel] `tfsdk:"custom_metric_config"`
	DatasetMetricConfigs fwtypes.ListNestedObjectValueOf[evaluationDatasetMetricConfigModel]         `tfsdk:"dataset_metric_config"`
	EvaluatorModelConfig fwtypes.ListNestedObjectValueOf[evaluatorModelConfigModel]                  `tfsdk:"evaluator_model_config"`
}

type evaluationDatasetMetricConfigModel struct {
	Dataset     fwtypes.ListNestedObjectValueOf[evaluationDatasetModel] `tfsdk:"dataset"`
	MetricNames fwtypes.ListOfString                                    `tfsdk:"metric_names"`
	TaskType    fwtypes.StringEnum[awstypes.EvaluationTaskType]         `tfsdk:"task_type"`
}

type evaluationDatasetModel struct {
	DatasetLocation fwtypes.ListNestedObjectValueOf[evaluationDatasetLocationModel] `tfsdk:"dataset_location"`
	Name            types.String                                                    `tfsdk:"name"`
}

// evaluationDatasetLocationModel maps to the awstypes.EvaluationDatasetLocation union.
type evaluationDatasetLocationModel struct {
	S3URI types.String `tfsdk:"s3_uri"`
}

var (
	_ fwflex.Expander  = evaluationDatasetLocationModel{}
	_ fwflex.Flattener = &evaluationDatasetLocationModel{}
)

func (m evaluationDatasetLocationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	r := awstypes.EvaluationDatasetLocationMemberS3Uri{
		Value: fwflex.StringValueFromFramework(ctx, m.S3URI),
	}

	return &r, diags
}

func (m *evaluationDatasetLocationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.EvaluationDatasetLocationMemberS3Uri:
		m.S3URI = fwflex.StringValueToFramework(ctx, t.Value)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("evaluation dataset location flatten: %T", v))
	}

	return diags
}

// evaluatorModelConfigModel maps to the awstypes.EvaluatorModelConfig union.
type evaluatorModelConfigModel struct {
	BedrockEvaluatorModels fwtypes.ListNestedObjectValueOf[bedrockEvaluatorModelModel] `tfsdk:"bedrock_evaluator_model"`
}

var (
	_ fwflex.Expander  = evaluatorModelConfigModel{}
	_ fwflex.Flattener = &evaluatorModelConfigModel{}
)

func (m evaluatorModelConfigModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	var r awstypes.EvaluatorModelConfigMemberBedrockEvaluatorModels
	smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, m.BedrockEvaluatorModels, &r.Value))
	if diags.HasError() {
		return nil, diags
	}

	return &r, diags
}

func (m *evaluatorModelConfigModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.EvaluatorModelConfigMemberBedrockEvaluatorModels:
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &m.BedrockEvaluatorModels))
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("evaluator model config flatten: %T", v))
	}

	return diags
}

type bedrockEvaluatorModelModel struct {
	ModelIdentifier types.String `tfsdk:"model_identifier"`
}

type automatedEvaluationCustomMetricConfigModel struct {
	CustomMetrics        fwtypes.ListNestedObjectValueOf[automatedEvaluationCustomMetricSourceModel] `tfsdk:"custom_metric"`
	EvaluatorModelConfig fwtypes.ListNestedObjectValueOf[customMetricEvaluatorModelConfigModel]      `tfsdk:"evaluator_model_config"`
}

// automatedEvaluationCustomMetricSourceModel maps to the
// awstypes.AutomatedEvaluationCustomMetricSource union.
type automatedEvaluationCustomMetricSourceModel struct {
	CustomMetricDefinition fwtypes.ListNestedObjectValueOf[customMetricDefinitionModel] `tfsdk:"custom_metric_definition"`
}

var (
	_ fwflex.Expander  = automatedEvaluationCustomMetricSourceModel{}
	_ fwflex.Flattener = &automatedEvaluationCustomMetricSourceModel{}
)

func (m automatedEvaluationCustomMetricSourceModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	var r awstypes.AutomatedEvaluationCustomMetricSourceMemberCustomMetricDefinition
	smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, m.CustomMetricDefinition, &r.Value))
	if diags.HasError() {
		return nil, diags
	}

	return &r, diags
}

func (m *automatedEvaluationCustomMetricSourceModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.AutomatedEvaluationCustomMetricSourceMemberCustomMetricDefinition:
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &m.CustomMetricDefinition))
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("automated evaluation custom metric source flatten: %T", v))
	}

	return diags
}

type customMetricDefinitionModel struct {
	Instructions types.String                                          `tfsdk:"instructions"`
	Name         types.String                                          `tfsdk:"name"`
	RatingScale  fwtypes.ListNestedObjectValueOf[ratingScaleItemModel] `tfsdk:"rating_scale"`
}

type ratingScaleItemModel struct {
	Definition types.String                                               `tfsdk:"definition"`
	Value      fwtypes.ListNestedObjectValueOf[ratingScaleItemValueModel] `tfsdk:"value"`
}

// ratingScaleItemValueModel maps to the awstypes.RatingScaleItemValue union.
type ratingScaleItemValueModel struct {
	FloatValue  types.Float32 `tfsdk:"float_value"`
	StringValue types.String  `tfsdk:"string_value"`
}

var (
	_ fwflex.Expander  = ratingScaleItemValueModel{}
	_ fwflex.Flattener = &ratingScaleItemValueModel{}
)

func (m ratingScaleItemValueModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.FloatValue.IsNull():
		r := awstypes.RatingScaleItemValueMemberFloatValue{
			Value: m.FloatValue.ValueFloat32(),
		}
		return &r, diags
	case !m.StringValue.IsNull():
		r := awstypes.RatingScaleItemValueMemberStringValue{
			Value: fwflex.StringValueFromFramework(ctx, m.StringValue),
		}
		return &r, diags
	}

	return nil, diags
}

func (m *ratingScaleItemValueModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.RatingScaleItemValueMemberFloatValue:
		m.FloatValue = types.Float32Value(t.Value)
	case awstypes.RatingScaleItemValueMemberStringValue:
		m.StringValue = fwflex.StringValueToFramework(ctx, t.Value)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("rating scale item value flatten: %T", v))
	}

	return diags
}

type customMetricEvaluatorModelConfigModel struct {
	BedrockEvaluatorModels fwtypes.ListNestedObjectValueOf[customMetricBedrockEvaluatorModelModel] `tfsdk:"bedrock_evaluator_model"`
}

type customMetricBedrockEvaluatorModelModel struct {
	ModelIdentifier types.String `tfsdk:"model_identifier"`
}

type humanEvaluationConfigModel struct {
	CustomMetrics        fwtypes.ListNestedObjectValueOf[humanEvaluationCustomMetricModel]   `tfsdk:"custom_metric"`
	DatasetMetricConfigs fwtypes.ListNestedObjectValueOf[evaluationDatasetMetricConfigModel] `tfsdk:"dataset_metric_config"`
	HumanWorkflowConfig  fwtypes.ListNestedObjectValueOf[humanWorkflowConfigModel]           `tfsdk:"human_workflow_config"`
}

type humanEvaluationCustomMetricModel struct {
	Description  types.String `tfsdk:"description"`
	Name         types.String `tfsdk:"name"`
	RatingMethod types.String `tfsdk:"rating_method"`
}

type humanWorkflowConfigModel struct {
	FlowDefinitionARN fwtypes.ARN  `tfsdk:"flow_definition_arn"`
	Instructions      types.String `tfsdk:"instructions"`
}

// inferenceConfigModel maps to the awstypes.EvaluationInferenceConfig union (Models or RagConfigs).
type inferenceConfigModel struct {
	Models     fwtypes.ListNestedObjectValueOf[evaluationModelConfigModel] `tfsdk:"model"`
	RAGConfigs fwtypes.ListNestedObjectValueOf[ragConfigModel]             `tfsdk:"rag_config"`
}

var (
	_ fwflex.Expander  = inferenceConfigModel{}
	_ fwflex.Flattener = &inferenceConfigModel{}
)

func (m inferenceConfigModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.Models.IsNull():
		var r awstypes.EvaluationInferenceConfigMemberModels
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, m.Models, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	case !m.RAGConfigs.IsNull():
		var r awstypes.EvaluationInferenceConfigMemberRagConfigs
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, m.RAGConfigs, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}

	return nil, diags
}

func (m *inferenceConfigModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.EvaluationInferenceConfigMemberModels:
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &m.Models))
	case awstypes.EvaluationInferenceConfigMemberRagConfigs:
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &m.RAGConfigs))
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("inference config flatten: %T", v))
	}

	return diags
}

// evaluationModelConfigModel maps to the awstypes.EvaluationModelConfig union (BedrockModel or
// PrecomputedInferenceSource).
type evaluationModelConfigModel struct {
	BedrockModel               fwtypes.ListNestedObjectValueOf[evaluationBedrockModelModel]               `tfsdk:"bedrock_model"`
	PrecomputedInferenceSource fwtypes.ListNestedObjectValueOf[evaluationPrecomputedInferenceSourceModel] `tfsdk:"precomputed_inference_source"`
}

var (
	_ fwflex.Expander  = evaluationModelConfigModel{}
	_ fwflex.Flattener = &evaluationModelConfigModel{}
)

func (m evaluationModelConfigModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.BedrockModel.IsNull():
		data, d := m.BedrockModel.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.EvaluationModelConfigMemberBedrockModel
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	case !m.PrecomputedInferenceSource.IsNull():
		data, d := m.PrecomputedInferenceSource.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.EvaluationModelConfigMemberPrecomputedInferenceSource
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}

	return nil, diags
}

func (m *evaluationModelConfigModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.EvaluationModelConfigMemberBedrockModel:
		var data evaluationBedrockModelModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.BedrockModel = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	case awstypes.EvaluationModelConfigMemberPrecomputedInferenceSource:
		var data evaluationPrecomputedInferenceSourceModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.PrecomputedInferenceSource = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("evaluation model config flatten: %T", v))
	}

	return diags
}

type evaluationBedrockModelModel struct {
	InferenceParams   types.String                                                   `tfsdk:"inference_params"`
	ModelIdentifier   types.String                                                   `tfsdk:"model_identifier"`
	PerformanceConfig fwtypes.ListNestedObjectValueOf[performanceConfigurationModel] `tfsdk:"performance_config"`
}

type performanceConfigurationModel struct {
	Latency fwtypes.StringEnum[awstypes.PerformanceConfigLatency] `tfsdk:"latency"`
}

type evaluationPrecomputedInferenceSourceModel struct {
	InferenceSourceIdentifier types.String `tfsdk:"inference_source_identifier"`
}

// ragConfigModel maps to the awstypes.RAGConfig union (KnowledgeBaseConfig or
// PrecomputedRagSourceConfig).
type ragConfigModel struct {
	KnowledgeBaseConfig        fwtypes.ListNestedObjectValueOf[knowledgeBaseConfigModel]        `tfsdk:"knowledge_base_config"`
	PrecomputedRagSourceConfig fwtypes.ListNestedObjectValueOf[precomputedRagSourceConfigModel] `tfsdk:"precomputed_rag_source_config"`
}

var (
	_ fwflex.Expander  = ragConfigModel{}
	_ fwflex.Flattener = &ragConfigModel{}
)

func (m ragConfigModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.KnowledgeBaseConfig.IsNull():
		data, d := m.KnowledgeBaseConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.RAGConfigMemberKnowledgeBaseConfig
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	case !m.PrecomputedRagSourceConfig.IsNull():
		data, d := m.PrecomputedRagSourceConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.RAGConfigMemberPrecomputedRagSourceConfig
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}

	return nil, diags
}

func (m *ragConfigModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.RAGConfigMemberKnowledgeBaseConfig:
		var data knowledgeBaseConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.KnowledgeBaseConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	case awstypes.RAGConfigMemberPrecomputedRagSourceConfig:
		var data precomputedRagSourceConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.PrecomputedRagSourceConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("rag config flatten: %T", v))
	}

	return diags
}

// knowledgeBaseConfigModel maps to the awstypes.KnowledgeBaseConfig union (RetrieveConfig or
// RetrieveAndGenerateConfig).
type knowledgeBaseConfigModel struct {
	RetrieveAndGenerateConfig fwtypes.ListNestedObjectValueOf[knowledgeBaseRetrieveAndGenerateConfigModel] `tfsdk:"retrieve_and_generate_config"`
	RetrieveConfig            fwtypes.ListNestedObjectValueOf[knowledgeBaseRetrieveConfigModel]            `tfsdk:"retrieve_config"`
}

var (
	_ fwflex.Expander  = knowledgeBaseConfigModel{}
	_ fwflex.Flattener = &knowledgeBaseConfigModel{}
)

func (m knowledgeBaseConfigModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.RetrieveAndGenerateConfig.IsNull():
		data, d := m.RetrieveAndGenerateConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		// awstypes.KnowledgeBaseConfigMemberRetrieveAndGenerateConfig.Value is an
		// awstypes.RetrieveAndGenerateConfiguration, which wraps the knowledge base
		// fields inside a KnowledgeBaseConfiguration member alongside a required Type
		// discriminator. Expand the model into that inner struct, then build the
		// wrapper explicitly, since AutoFlex only matches fields at a single level.
		var knowledgeBaseConfig awstypes.KnowledgeBaseRetrieveAndGenerateConfiguration
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &knowledgeBaseConfig))
		if diags.HasError() {
			return nil, diags
		}
		r := awstypes.KnowledgeBaseConfigMemberRetrieveAndGenerateConfig{
			Value: awstypes.RetrieveAndGenerateConfiguration{
				Type:                       awstypes.RetrieveAndGenerateTypeKnowledgeBase,
				KnowledgeBaseConfiguration: &knowledgeBaseConfig,
			},
		}
		return &r, diags
	case !m.RetrieveConfig.IsNull():
		data, d := m.RetrieveConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.KnowledgeBaseConfigMemberRetrieveConfig
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}

	return nil, diags
}

func (m *knowledgeBaseConfigModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.KnowledgeBaseConfigMemberRetrieveAndGenerateConfig:
		if t.Value.KnowledgeBaseConfiguration == nil {
			diags.AddError("Unsupported Type", "knowledge base config flatten: retrieve and generate config has no knowledge base configuration")
			return diags
		}
		var data knowledgeBaseRetrieveAndGenerateConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value.KnowledgeBaseConfiguration, &data))
		if diags.HasError() {
			return diags
		}
		m.RetrieveAndGenerateConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	case awstypes.KnowledgeBaseConfigMemberRetrieveConfig:
		var data knowledgeBaseRetrieveConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.RetrieveConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("knowledge base config flatten: %T", v))
	}

	return diags
}

type knowledgeBaseRetrieveConfigModel struct {
	KnowledgeBaseID                     types.String                                                              `tfsdk:"knowledge_base_id"`
	KnowledgeBaseRetrievalConfiguration fwtypes.ListNestedObjectValueOf[knowledgeBaseRetrievalConfigurationModel] `tfsdk:"knowledge_base_retrieval_configuration"`
}

// knowledgeBaseRetrieveAndGenerateConfigModel maps to
// awstypes.KnowledgeBaseRetrieveAndGenerateConfiguration.
//
// awstypes.KnowledgeBaseConfigMemberRetrieveAndGenerateConfig.Value is actually
// an awstypes.RetrieveAndGenerateConfiguration, which wraps this struct behind
// a required Type discriminator and a KnowledgeBaseConfiguration field. See
// knowledgeBaseConfigModel.Expand/Flatten for where that wrapping/unwrapping
// happens; only the KNOWLEDGE_BASE variant is supported
// (ExternalSourcesConfiguration is not mapped).
type knowledgeBaseRetrieveAndGenerateConfigModel struct {
	KnowledgeBaseID        types.String                                                              `tfsdk:"knowledge_base_id"`
	ModelARN               fwtypes.ARN                                                               `tfsdk:"model_arn"`
	RetrievalConfiguration fwtypes.ListNestedObjectValueOf[knowledgeBaseRetrievalConfigurationModel] `tfsdk:"retrieval_configuration"`
}

// knowledgeBaseRetrievalConfigurationModel maps to awstypes.KnowledgeBaseRetrievalConfiguration.
type knowledgeBaseRetrievalConfigurationModel struct {
	VectorSearchConfiguration fwtypes.ListNestedObjectValueOf[knowledgeBaseVectorSearchConfigurationModel] `tfsdk:"vector_search_configuration"`
}

// knowledgeBaseVectorSearchConfigurationModel maps to a subset of
// awstypes.KnowledgeBaseVectorSearchConfiguration: only NumberOfResults is mapped.
//
// The remaining fields are not currently mapped:
//   - Filter (RetrievalFilter) is a recursive tagged union with more than a
//     dozen member types (equals, notEquals, greaterThan, in, andAll, orAll, and
//     so on), each of which can itself nest further filters via andAll/orAll.
//     Reproducing that entire tree here would add substantial schema (and memory)
//     overhead.
//   - ImplicitFilterConfiguration (*ImplicitFilterConfiguration)
//     depends on metadata attribute schemas that are defined by the knowledge
//     base being evaluated, and layers additional complexity on top of the same
//     Filter modeling gap above.
//   - OverrideSearchType (SearchType) and
//     RerankingConfiguration (*VectorSearchRerankingConfiguration) are narrow,
//     vector-store-specific tuning knobs which Amazon Bedrock picks defaults for
//     when omitted.
type knowledgeBaseVectorSearchConfigurationModel struct {
	NumberOfResults types.Int32 `tfsdk:"number_of_results"`
}

// precomputedRagSourceConfigModel maps to the awstypes.EvaluationPrecomputedRagSourceConfig union
// (RetrieveAndGenerateSourceConfig or RetrieveSourceConfig).
type precomputedRagSourceConfigModel struct {
	RetrieveAndGenerateSourceConfig fwtypes.ListNestedObjectValueOf[evaluationPrecomputedRetrieveAndGenerateSourceConfigModel] `tfsdk:"retrieve_and_generate_source_config"`
	RetrieveSourceConfig            fwtypes.ListNestedObjectValueOf[evaluationPrecomputedRetrieveSourceConfigModel]            `tfsdk:"retrieve_source_config"`
}

var (
	_ fwflex.Expander  = precomputedRagSourceConfigModel{}
	_ fwflex.Flattener = &precomputedRagSourceConfigModel{}
)

func (m precomputedRagSourceConfigModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.RetrieveAndGenerateSourceConfig.IsNull():
		data, d := m.RetrieveAndGenerateSourceConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.EvaluationPrecomputedRagSourceConfigMemberRetrieveAndGenerateSourceConfig
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	case !m.RetrieveSourceConfig.IsNull():
		data, d := m.RetrieveSourceConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.EvaluationPrecomputedRagSourceConfigMemberRetrieveSourceConfig
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}

	return nil, diags
}

func (m *precomputedRagSourceConfigModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.EvaluationPrecomputedRagSourceConfigMemberRetrieveAndGenerateSourceConfig:
		var data evaluationPrecomputedRetrieveAndGenerateSourceConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.RetrieveAndGenerateSourceConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	case awstypes.EvaluationPrecomputedRagSourceConfigMemberRetrieveSourceConfig:
		var data evaluationPrecomputedRetrieveSourceConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.RetrieveSourceConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("precomputed rag source config flatten: %T", v))
	}

	return diags
}

type evaluationPrecomputedRetrieveAndGenerateSourceConfigModel struct {
	RAGSourceIdentifier types.String `tfsdk:"rag_source_identifier"`
}

type evaluationPrecomputedRetrieveSourceConfigModel struct {
	RAGSourceIdentifier types.String `tfsdk:"rag_source_identifier"`
}

type evaluationOutputDataConfigModel struct {
	S3URI types.String `tfsdk:"s3_uri"`
}
