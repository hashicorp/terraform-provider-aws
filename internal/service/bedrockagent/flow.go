// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrockagent

import (
	"context"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent/document"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsmithy "github.com/hashicorp/terraform-provider-aws/internal/smithy"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagent_flow", name="Flow")
// @Tags(identifierAttribute="arn")
func newFlowResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &flowResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNameFlow = "Flow"
)

type flowResource struct {
	framework.ResourceWithModel[flowResourceModel]
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *flowResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
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
			"customer_encryption_key_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 200),
				},
			},
			names.AttrExecutionRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^([0-9a-zA-Z][_-]?){1,100}$`), "must only contain alphanumeric characters, hyphens and underscores"),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.FlowStatus](),
				Computed:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrVersion: schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"definition": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[flowDefinitionModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"connection": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[flowConnectionModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrName: schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z]([_]?[0-9a-zA-Z]){1,100}$`), "must only contain alphanumeric characters"),
										},
									},
									names.AttrSource: schema.StringAttribute{
										Validators: []validator.String{
											stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z]([_]?[0-9a-zA-Z]){1,50}$`), "must only contain alphanumeric characters"),
										},
										Required: true,
									},
									names.AttrTarget: schema.StringAttribute{
										Validators: []validator.String{
											stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z]([_]?[0-9a-zA-Z]){1,50}$`), "must only contain alphanumeric characters"),
										},
										Required: true,
									},
									names.AttrType: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.FlowConnectionType](),
										Required:   true,
									},
								},
								Blocks: map[string]schema.Block{
									names.AttrConfiguration: schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[flowConnectionConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"conditional": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[flowConditionalConnectionConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrCondition: schema.StringAttribute{
																Required: true,
																Validators: []validator.String{
																	stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z]([_]?[0-9a-zA-Z]){1,50}$`), "must only contain alphanumeric characters"),
																},
															},
														},
													},
												},
												"data": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[flowDataConnectionConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
														listvalidator.ExactlyOneOf(
															path.MatchRelative().AtParent().AtName("conditional"),
															path.MatchRelative().AtParent().AtName("data"),
														),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"source_output": schema.StringAttribute{
																Required: true,
																Validators: []validator.String{
																	stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z]([_]?[0-9a-zA-Z]){1,50}$`), "must only contain alphanumeric characters"),
																},
															},
															"target_input": schema.StringAttribute{
																Required: true,
																Validators: []validator.String{
																	stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z]([_]?[0-9a-zA-Z]){1,50}$`), "must only contain alphanumeric characters"),
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
						"node": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[flowNodeModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrName: schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z]([_]?[0-9a-zA-Z]){1,50}$`), "must only contain alphanumeric characters"),
										},
									},
									names.AttrType: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.FlowNodeType](),
										Required:   true,
									},
								},
								Blocks: map[string]schema.Block{
									names.AttrConfiguration: schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[flowNodeConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"agent": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[agentFlowNodeConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
														listvalidator.ExactlyOneOf(
															path.MatchRelative().AtParent().AtName("agent"),
															path.MatchRelative().AtParent().AtName("collector"),
															path.MatchRelative().AtParent().AtName(names.AttrCondition),
															path.MatchRelative().AtParent().AtName("inline_code"),
															path.MatchRelative().AtParent().AtName("input"),
															path.MatchRelative().AtParent().AtName("iterator"),
															path.MatchRelative().AtParent().AtName("knowledge_base"),
															path.MatchRelative().AtParent().AtName("lambda_function"),
															path.MatchRelative().AtParent().AtName("lex"),
															path.MatchRelative().AtParent().AtName("output"),
															path.MatchRelative().AtParent().AtName("prompt"),
															path.MatchRelative().AtParent().AtName("retrieval"),
															path.MatchRelative().AtParent().AtName("storage"),
														),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"agent_alias_arn": schema.StringAttribute{
																CustomType: fwtypes.ARNType,
																Required:   true,
															},
														},
													},
												},
												"collector": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[collectorFlowNodeConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
												},
												names.AttrCondition: schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[conditionFlowNodeConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Blocks: map[string]schema.Block{
															names.AttrCondition: schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[flowConditionModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeBetween(1, 5),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		names.AttrExpression: schema.StringAttribute{
																			Optional: true,
																			Validators: []validator.String{
																				stringvalidator.LengthBetween(1, 64),
																			},
																		},
																		names.AttrName: schema.StringAttribute{
																			Required: true,
																			Validators: []validator.String{
																				stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z]([_]?[0-9a-zA-Z]){1,50}$`), "must only contain alphanumeric characters"),
																			},
																		},
																	},
																},
															},
														},
													},
												},
												"inline_code": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[inlineCodeFlowNodeConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"code": schema.StringAttribute{
																Required: true,
																Validators: []validator.String{
																	stringvalidator.LengthBetween(1, 5000000),
																},
															},
															"language": schema.StringAttribute{
																CustomType: fwtypes.StringEnumType[awstypes.SupportedLanguages](),
																Required:   true,
															},
														},
													},
												},
												"input": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[inputFlowNodeConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
												},
												"iterator": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[iteratorFlowNodeConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
												},
												"knowledge_base": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[knowledgeBaseFlowNodeConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"knowledge_base_id": schema.StringAttribute{
																Required: true,
															},
															"model_id": schema.StringAttribute{
																Required: true,
															},
															"number_of_results": schema.Int64Attribute{
																Optional: true,
																Validators: []validator.Int64{
																	int64validator.Between(1, 100),
																},
															},
														},
														Blocks: map[string]schema.Block{
															"guardrail_configuration": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[guardrailConfigurationModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"guardrail_identifier": schema.StringAttribute{
																			Required: true,
																		},
																		"guardrail_version": schema.StringAttribute{
																			Required: true,
																		},
																	},
																},
															},
															"inference_configuration": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[promptInferenceConfigurationModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Blocks: map[string]schema.Block{
																		"text": schema.ListNestedBlock{
																			CustomType: fwtypes.NewListNestedObjectTypeOf[promptModelInferenceConfigurationModel](ctx),
																			Validators: []validator.List{
																				listvalidator.SizeAtMost(1),
																				listvalidator.ExactlyOneOf(
																					path.MatchRelative().AtParent().AtName("text"),
																				),
																			},
																			NestedObject: schema.NestedBlockObject{
																				Attributes: map[string]schema.Attribute{
																					"max_tokens": schema.Int32Attribute{
																						Optional: true,
																					},
																					"stop_sequences": schema.ListAttribute{
																						CustomType:  fwtypes.ListOfStringType,
																						ElementType: types.StringType,
																						Optional:    true,
																					},
																					"temperature": schema.Float32Attribute{
																						Optional: true,
																					},
																					"top_p": schema.Float32Attribute{
																						Optional: true,
																					},
																				},
																			},
																		},
																	},
																},
															},
															// TODO More fields.
														},
													},
												},
												"lambda_function": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[lambdaFunctionFlowNodeConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"lambda_arn": schema.StringAttribute{
																CustomType: fwtypes.ARNType,
																Required:   true,
															},
														},
													},
												},
												"lex": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[lexFlowNodeConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"bot_alias_arn": schema.StringAttribute{
																CustomType: fwtypes.ARNType,
																Required:   true,
															},
															"locale_id": schema.StringAttribute{
																Required: true,
															},
														},
													},
												},
												// TODO Loop stuff.
												"output": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[outputFlowNodeConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
												},
												"prompt": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[promptFlowNodeConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Blocks: map[string]schema.Block{
															"guardrail_configuration": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[guardrailConfigurationModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"guardrail_identifier": schema.StringAttribute{
																			Required: true,
																		},
																		"guardrail_version": schema.StringAttribute{
																			Required: true,
																		},
																	},
																},
															},
															"source_configuration": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[promptFlowNodeSourceConfigurationModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Blocks: map[string]schema.Block{
																		"inline": schema.ListNestedBlock{
																			CustomType: fwtypes.NewListNestedObjectTypeOf[promptFlowNodeInlineConfigurationModel](ctx),
																			Validators: []validator.List{
																				listvalidator.SizeAtMost(1),
																				listvalidator.ExactlyOneOf(
																					path.MatchRelative().AtParent().AtName("inline"),
																					path.MatchRelative().AtParent().AtName("resource"),
																				),
																			},
																			NestedObject: schema.NestedBlockObject{
																				Attributes: map[string]schema.Attribute{
																					"additional_model_request_fields": schema.StringAttribute{
																						CustomType: jsontypes.NormalizedType{},
																						Optional:   true,
																					},
																					"model_id": schema.StringAttribute{
																						Required: true,
																					},
																					"template_type": schema.StringAttribute{
																						CustomType: fwtypes.StringEnumType[awstypes.PromptTemplateType](),
																						Required:   true,
																					},
																				},
																				Blocks: map[string]schema.Block{
																					"inference_configuration": schema.ListNestedBlock{
																						CustomType: fwtypes.NewListNestedObjectTypeOf[promptInferenceConfigurationModel](ctx),
																						Validators: []validator.List{
																							listvalidator.SizeAtMost(1),
																						},
																						NestedObject: schema.NestedBlockObject{
																							Blocks: map[string]schema.Block{
																								"text": schema.ListNestedBlock{
																									CustomType: fwtypes.NewListNestedObjectTypeOf[promptModelInferenceConfigurationModel](ctx),
																									Validators: []validator.List{
																										listvalidator.SizeAtMost(1),
																										listvalidator.ExactlyOneOf(
																											path.MatchRelative().AtParent().AtName("text"),
																										),
																									},
																									NestedObject: schema.NestedBlockObject{
																										Attributes: map[string]schema.Attribute{
																											"max_tokens": schema.Int32Attribute{
																												Optional: true,
																											},
																											"stop_sequences": schema.ListAttribute{
																												CustomType:  fwtypes.ListOfStringType,
																												ElementType: types.StringType,
																												Optional:    true,
																											},
																											"temperature": schema.Float32Attribute{
																												Optional: true,
																											},
																											"top_p": schema.Float32Attribute{
																												Optional: true,
																											},
																										},
																									},
																								},
																							},
																						},
																					},
																					"template_configuration": schema.ListNestedBlock{
																						CustomType: fwtypes.NewListNestedObjectTypeOf[promptTemplateConfigurationModel](ctx),
																						Validators: []validator.List{
																							listvalidator.SizeAtMost(1),
																						},
																						NestedObject: schema.NestedBlockObject{
																							Blocks: map[string]schema.Block{
																								"chat": schema.ListNestedBlock{
																									CustomType: fwtypes.NewListNestedObjectTypeOf[chatPromptTemplateConfigurationModel](ctx),
																									Validators: []validator.List{
																										listvalidator.SizeAtMost(1),
																										listvalidator.ExactlyOneOf(
																											path.MatchRelative().AtParent().AtName("chat"),
																											path.MatchRelative().AtParent().AtName("text"),
																										),
																									},
																									NestedObject: schema.NestedBlockObject{
																										Blocks: map[string]schema.Block{
																											"input_variable": schema.ListNestedBlock{
																												CustomType: fwtypes.NewListNestedObjectTypeOf[promptInputVariableModel](ctx),
																												Validators: []validator.List{
																													listvalidator.SizeBetween(0, 20),
																												},
																												NestedObject: schema.NestedBlockObject{
																													Attributes: map[string]schema.Attribute{
																														names.AttrName: schema.StringAttribute{
																															Required: true,
																														},
																													},
																												},
																											},
																											names.AttrMessage: schema.ListNestedBlock{
																												CustomType: fwtypes.NewListNestedObjectTypeOf[messageModel](ctx),
																												Validators: []validator.List{
																													listvalidator.IsRequired(),
																													listvalidator.SizeAtLeast(1),
																												},
																												NestedObject: schema.NestedBlockObject{
																													Attributes: map[string]schema.Attribute{
																														names.AttrRole: schema.StringAttribute{
																															CustomType: fwtypes.StringEnumType[awstypes.ConversationRole](),
																															Required:   true,
																														},
																													},
																													Blocks: map[string]schema.Block{
																														names.AttrContent: schema.ListNestedBlock{
																															CustomType: fwtypes.NewListNestedObjectTypeOf[contentBlockModel](ctx),
																															Validators: []validator.List{
																																listvalidator.SizeAtMost(1),
																															},
																															NestedObject: schema.NestedBlockObject{
																																Attributes: map[string]schema.Attribute{
																																	"text": schema.StringAttribute{
																																		Optional: true,
																																	},
																																},
																																Blocks: map[string]schema.Block{
																																	"cache_point": schema.ListNestedBlock{
																																		CustomType: fwtypes.NewListNestedObjectTypeOf[cachePointBlockModel](ctx),
																																		Validators: []validator.List{
																																			listvalidator.SizeAtMost(1),
																																			listvalidator.ExactlyOneOf(
																																				path.MatchRelative().AtParent().AtName("cache_point"),
																																				path.MatchRelative().AtParent().AtName("text"),
																																			),
																																		},
																																		NestedObject: schema.NestedBlockObject{
																																			Attributes: map[string]schema.Attribute{
																																				names.AttrType: schema.StringAttribute{
																																					CustomType: fwtypes.StringEnumType[awstypes.CachePointType](),
																																					Required:   true,
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
																											"system": schema.ListNestedBlock{
																												CustomType: fwtypes.NewListNestedObjectTypeOf[systemContentBlockModel](ctx),
																												NestedObject: schema.NestedBlockObject{
																													Attributes: map[string]schema.Attribute{
																														"text": schema.StringAttribute{
																															Optional: true,
																														},
																													},
																													Blocks: map[string]schema.Block{
																														"cache_point": schema.ListNestedBlock{
																															CustomType: fwtypes.NewListNestedObjectTypeOf[cachePointBlockModel](ctx),
																															Validators: []validator.List{
																																listvalidator.ExactlyOneOf(
																																	path.MatchRelative().AtParent().AtName("cache_point"),
																																	path.MatchRelative().AtParent().AtName("text"),
																																),
																																listvalidator.SizeAtMost(1),
																															},
																															NestedObject: schema.NestedBlockObject{
																																Attributes: map[string]schema.Attribute{
																																	names.AttrType: schema.StringAttribute{
																																		CustomType: fwtypes.StringEnumType[awstypes.CachePointType](),
																																		Required:   true,
																																	},
																																},
																															},
																														},
																													},
																												},
																											},
																											"tool_configuration": schema.ListNestedBlock{
																												CustomType: fwtypes.NewListNestedObjectTypeOf[toolConfigurationModel](ctx),
																												Validators: []validator.List{
																													listvalidator.SizeAtMost(1),
																												},
																												NestedObject: schema.NestedBlockObject{
																													Blocks: map[string]schema.Block{
																														"tool_choice": schema.ListNestedBlock{
																															CustomType: fwtypes.NewListNestedObjectTypeOf[toolChoiceModel](ctx),
																															Validators: []validator.List{
																																listvalidator.SizeAtMost(1),
																															},
																															NestedObject: schema.NestedBlockObject{
																																Blocks: map[string]schema.Block{
																																	"any": schema.ListNestedBlock{
																																		CustomType: fwtypes.NewListNestedObjectTypeOf[anyToolChoiceModel](ctx),
																																		Validators: []validator.List{
																																			listvalidator.SizeAtMost(1),
																																			listvalidator.ExactlyOneOf(
																																				path.MatchRelative().AtParent().AtName("any"),
																																				path.MatchRelative().AtParent().AtName("auto"),
																																				path.MatchRelative().AtParent().AtName("tool"),
																																			),
																																		},
																																	},
																																	"auto": schema.ListNestedBlock{
																																		CustomType: fwtypes.NewListNestedObjectTypeOf[autoToolChoiceModel](ctx),
																																		Validators: []validator.List{
																																			listvalidator.SizeAtMost(1),
																																		},
																																	},
																																	"tool": schema.ListNestedBlock{
																																		CustomType: fwtypes.NewListNestedObjectTypeOf[specificToolChoiceModel](ctx),
																																		Validators: []validator.List{
																																			listvalidator.SizeAtMost(1),
																																		},
																																		NestedObject: schema.NestedBlockObject{
																																			Attributes: map[string]schema.Attribute{
																																				names.AttrName: schema.StringAttribute{
																																					Required: true,
																																				},
																																			},
																																		},
																																	},
																																},
																															},
																														},
																														"tool": schema.ListNestedBlock{
																															CustomType: fwtypes.NewListNestedObjectTypeOf[toolModel](ctx),
																															NestedObject: schema.NestedBlockObject{
																																Blocks: map[string]schema.Block{
																																	"cache_point": schema.ListNestedBlock{
																																		CustomType: fwtypes.NewListNestedObjectTypeOf[cachePointBlockModel](ctx),
																																		Validators: []validator.List{
																																			listvalidator.SizeAtMost(1),
																																			listvalidator.ExactlyOneOf(
																																				path.MatchRelative().AtParent().AtName("cache_point"),
																																				path.MatchRelative().AtParent().AtName("tool_spec"),
																																			),
																																		},
																																		NestedObject: schema.NestedBlockObject{
																																			Attributes: map[string]schema.Attribute{
																																				names.AttrType: schema.StringAttribute{
																																					CustomType: fwtypes.StringEnumType[awstypes.CachePointType](),
																																					Required:   true,
																																				},
																																			},
																																		},
																																	},
																																	"tool_spec": schema.ListNestedBlock{
																																		CustomType: fwtypes.NewListNestedObjectTypeOf[toolSpecificationModel](ctx),
																																		Validators: []validator.List{
																																			listvalidator.SizeAtMost(1),
																																		},
																																		NestedObject: schema.NestedBlockObject{
																																			Attributes: map[string]schema.Attribute{
																																				names.AttrDescription: schema.StringAttribute{
																																					Optional: true,
																																				},
																																				names.AttrName: schema.StringAttribute{
																																					Required: true,
																																				},
																																			},
																																			Blocks: map[string]schema.Block{
																																				"input_schema": schema.ListNestedBlock{
																																					CustomType: fwtypes.NewListNestedObjectTypeOf[toolInputSchemaModel](ctx),
																																					Validators: []validator.List{
																																						listvalidator.SizeAtMost(1),
																																					},
																																					NestedObject: schema.NestedBlockObject{
																																						Attributes: map[string]schema.Attribute{
																																							names.AttrJSON: schema.StringAttribute{
																																								CustomType: jsontypes.NormalizedType{},
																																								Optional:   true,
																																								Validators: []validator.String{
																																									stringvalidator.ExactlyOneOf(
																																										path.MatchRelative().AtParent().AtName(names.AttrJSON),
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
																										},
																									},
																								},
																								"text": schema.ListNestedBlock{
																									CustomType: fwtypes.NewListNestedObjectTypeOf[textPromptTemplateConfigurationModel](ctx),
																									Validators: []validator.List{
																										listvalidator.SizeAtMost(1),
																									},
																									NestedObject: schema.NestedBlockObject{
																										Attributes: map[string]schema.Attribute{
																											"text": schema.StringAttribute{
																												Required: true,
																											},
																										},
																										Blocks: map[string]schema.Block{
																											"cache_point": schema.ListNestedBlock{
																												CustomType: fwtypes.NewListNestedObjectTypeOf[cachePointModel](ctx),
																												Validators: []validator.List{
																													listvalidator.SizeAtMost(1),
																												},
																												NestedObject: schema.NestedBlockObject{
																													Attributes: map[string]schema.Attribute{
																														names.AttrType: schema.StringAttribute{
																															CustomType: fwtypes.StringEnumType[awstypes.CachePointType](),
																															Required:   true,
																														},
																													},
																												},
																											},
																											"input_variable": schema.ListNestedBlock{
																												CustomType: fwtypes.NewListNestedObjectTypeOf[promptInputVariableModel](ctx),
																												NestedObject: schema.NestedBlockObject{
																													Attributes: map[string]schema.Attribute{
																														names.AttrName: schema.StringAttribute{
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
																		"resource": schema.ListNestedBlock{
																			CustomType: fwtypes.NewListNestedObjectTypeOf[promptFlowNodeResourceConfigurationModel](ctx),
																			Validators: []validator.List{
																				listvalidator.SizeAtMost(1),
																			},
																			NestedObject: schema.NestedBlockObject{
																				Attributes: map[string]schema.Attribute{
																					"prompt_arn": schema.StringAttribute{
																						CustomType: fwtypes.ARNType,
																						Required:   true,
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
												"retrieval": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[retrievalFlowNodeConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Blocks: map[string]schema.Block{
															"service_configuration": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[retrievalFlowNodeServiceConfigurationModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Blocks: map[string]schema.Block{
																		"s3": schema.ListNestedBlock{
																			CustomType: fwtypes.NewListNestedObjectTypeOf[retrievalFlowNodeS3ConfigurationModel](ctx),
																			Validators: []validator.List{
																				listvalidator.SizeAtMost(1),
																				listvalidator.ExactlyOneOf(
																					path.MatchRelative().AtParent().AtName("s3"),
																				),
																			},
																			NestedObject: schema.NestedBlockObject{
																				Attributes: map[string]schema.Attribute{
																					names.AttrBucketName: schema.StringAttribute{
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
												"storage": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[storageFlowNodeConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Blocks: map[string]schema.Block{
															"service_configuration": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[storageFlowNodeServiceConfigurationModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Blocks: map[string]schema.Block{
																		"s3": schema.ListNestedBlock{
																			CustomType: fwtypes.NewListNestedObjectTypeOf[storageFlowNodeS3ConfigurationModel](ctx),
																			Validators: []validator.List{
																				listvalidator.SizeAtMost(1),
																				listvalidator.ExactlyOneOf(
																					path.MatchRelative().AtParent().AtName("s3"),
																				),
																			},
																			NestedObject: schema.NestedBlockObject{
																				Attributes: map[string]schema.Attribute{
																					names.AttrBucketName: schema.StringAttribute{
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
									"input": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[flowNodeInputModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeBetween(0, 20),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"category": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.FlowNodeInputCategory](),
													Optional:   true,
												},
												names.AttrExpression: schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 64),
													},
												},
												names.AttrName: schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z]([_]?[0-9a-zA-Z]){1,50}$`), "must only contain alphanumeric characters"),
													},
												},
												names.AttrType: schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.FlowNodeIODataType](),
													Required:   true,
												},
											},
										},
									},
									"output": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[flowNodeOutputModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeBetween(0, 5),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrName: schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z]([_]?[0-9a-zA-Z]){1,50}$`), "must only contain alphanumeric characters"),
													},
												},
												names.AttrType: schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.FlowNodeIODataType](),
													Required:   true,
												},
											},
										}},
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

func (r *flowResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data flowResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	var input bedrockagent.CreateFlowInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	input.ClientToken = aws.String(sdkid.UniqueId())
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateFlow(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionCreating, ResNameFlow, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *flowResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data flowResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	output, err := findFlowByID(ctx, conn, data.ID.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionSetting, ResNameFlow, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *flowResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old flowResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	response.Diagnostics.Append(d...)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input bedrockagent.UpdateFlowInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		input.FlowIdentifier = new.ID.ValueStringPointer()

		output, err := conn.UpdateFlow(ctx, &input)
		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionUpdating, ResNameFlow, new.ID.String(), err),
				err.Error(),
			)
			return
		}

		response.Diagnostics.Append(fwflex.Flatten(ctx, output, &new)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Set values for unknowns.
		new.CreatedAt = old.CreatedAt
		new.UpdatedAt = timetypes.NewRFC3339TimePointerValue(output.UpdatedAt)
		new.Version = fwflex.StringToFramework(ctx, output.Version)
		new.Status = fwtypes.StringEnumValue(output.Status)
	} else {
		new.CreatedAt = old.CreatedAt
		new.UpdatedAt = old.UpdatedAt
		new.Version = old.Version
		new.Status = old.Status
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *flowResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data flowResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	input := bedrockagent.DeleteFlowInput{
		FlowIdentifier: data.ID.ValueStringPointer(),
	}
	_, err := conn.DeleteFlow(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionDeleting, ResNameFlow, data.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func findFlowByID(ctx context.Context, conn *bedrockagent.Client, id string) (*bedrockagent.GetFlowOutput, error) {
	input := bedrockagent.GetFlowInput{
		FlowIdentifier: aws.String(id),
	}
	output, err := conn.GetFlow(ctx, &input)

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

type flowResourceModel struct {
	framework.WithRegionModel
	ARN                      types.String                                         `tfsdk:"arn"`
	CreatedAt                timetypes.RFC3339                                    `tfsdk:"created_at"`
	CustomerEncryptionKeyARN fwtypes.ARN                                          `tfsdk:"customer_encryption_key_arn"`
	Definition               fwtypes.ListNestedObjectValueOf[flowDefinitionModel] `tfsdk:"definition"`
	Description              types.String                                         `tfsdk:"description"`
	ExecutionRoleARN         fwtypes.ARN                                          `tfsdk:"execution_role_arn"`
	ID                       types.String                                         `tfsdk:"id"`
	Name                     types.String                                         `tfsdk:"name"`
	Status                   fwtypes.StringEnum[awstypes.FlowStatus]              `tfsdk:"status"`
	Tags                     tftags.Map                                           `tfsdk:"tags"`
	TagsAll                  tftags.Map                                           `tfsdk:"tags_all"`
	Timeouts                 timeouts.Value                                       `tfsdk:"timeouts"`
	UpdatedAt                timetypes.RFC3339                                    `tfsdk:"updated_at"`
	Version                  types.String                                         `tfsdk:"version"`
}

type flowDefinitionModel struct {
	Connections fwtypes.ListNestedObjectValueOf[flowConnectionModel] `tfsdk:"connection"`
	Nodes       fwtypes.ListNestedObjectValueOf[flowNodeModel]       `tfsdk:"node"`
}

type flowConnectionModel struct {
	Configuration fwtypes.ListNestedObjectValueOf[flowConnectionConfigurationModel] `tfsdk:"configuration"`
	Name          types.String                                                      `tfsdk:"name"`
	Source        types.String                                                      `tfsdk:"source"`
	Target        types.String                                                      `tfsdk:"target"`
	Type          fwtypes.StringEnum[awstypes.FlowConnectionType]                   `tfsdk:"type"`
}

// Tagged union
type flowConnectionConfigurationModel struct {
	Conditional fwtypes.ListNestedObjectValueOf[flowConditionalConnectionConfigurationModel] `tfsdk:"conditional"`
	Data        fwtypes.ListNestedObjectValueOf[flowDataConnectionConfigurationModel]        `tfsdk:"data"`
}

type flowConditionalConnectionConfigurationModel struct {
	Condition types.String `tfsdk:"condition"`
}

type flowDataConnectionConfigurationModel struct {
	SourceOutput types.String `tfsdk:"source_output"`
	TargetInput  types.String `tfsdk:"target_input"`
}

func (m *flowConnectionConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.FlowConnectionConfigurationMemberData:
		var model flowDataConnectionConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Data = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.FlowConnectionConfigurationMemberConditional:
		var model flowConditionalConnectionConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Conditional = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	default:
		return diags
	}
}

func (m flowConnectionConfigurationModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.Data.IsNull():
		flowConnectionConfigurationData, d := m.Data.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.FlowConnectionConfigurationMemberData
		diags.Append(fwflex.Expand(ctx, flowConnectionConfigurationData, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	case !m.Conditional.IsNull():
		flowConnectionConfigurationConditional, d := m.Conditional.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.FlowConnectionConfigurationMemberConditional
		diags.Append(fwflex.Expand(ctx, flowConnectionConfigurationConditional, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

type flowNodeModel struct {
	Configuration fwtypes.ListNestedObjectValueOf[flowNodeConfigurationModel] `tfsdk:"configuration"`
	Inputs        fwtypes.ListNestedObjectValueOf[flowNodeInputModel]         `tfsdk:"input"`
	Name          types.String                                                `tfsdk:"name"`
	Outputs       fwtypes.ListNestedObjectValueOf[flowNodeOutputModel]        `tfsdk:"output"`
	Type          fwtypes.StringEnum[awstypes.FlowNodeType]                   `tfsdk:"type"`
}

// Tagged union
type flowNodeConfigurationModel struct {
	Agent          fwtypes.ListNestedObjectValueOf[agentFlowNodeConfigurationModel]          `tfsdk:"agent"`
	Collector      fwtypes.ListNestedObjectValueOf[collectorFlowNodeConfigurationModel]      `tfsdk:"collector"`
	Condition      fwtypes.ListNestedObjectValueOf[conditionFlowNodeConfigurationModel]      `tfsdk:"condition"`
	InlineCode     fwtypes.ListNestedObjectValueOf[inlineCodeFlowNodeConfigurationModel]     `tfsdk:"inline_code"`
	Input          fwtypes.ListNestedObjectValueOf[inputFlowNodeConfigurationModel]          `tfsdk:"input"`
	Iterator       fwtypes.ListNestedObjectValueOf[iteratorFlowNodeConfigurationModel]       `tfsdk:"iterator"`
	KnowledgeBase  fwtypes.ListNestedObjectValueOf[knowledgeBaseFlowNodeConfigurationModel]  `tfsdk:"knowledge_base"`
	LambdaFunction fwtypes.ListNestedObjectValueOf[lambdaFunctionFlowNodeConfigurationModel] `tfsdk:"lambda_function"`
	Lex            fwtypes.ListNestedObjectValueOf[lexFlowNodeConfigurationModel]            `tfsdk:"lex"`
	// TODO Loop stuff
	Output    fwtypes.ListNestedObjectValueOf[outputFlowNodeConfigurationModel]    `tfsdk:"output"`
	Prompt    fwtypes.ListNestedObjectValueOf[promptFlowNodeConfigurationModel]    `tfsdk:"prompt"`
	Retrieval fwtypes.ListNestedObjectValueOf[retrievalFlowNodeConfigurationModel] `tfsdk:"retrieval"`
	Storage   fwtypes.ListNestedObjectValueOf[storageFlowNodeConfigurationModel]   `tfsdk:"storage"`
}

func (m *flowNodeConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.FlowNodeConfigurationMemberAgent:
		var model agentFlowNodeConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Agent = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.FlowNodeConfigurationMemberCollector:
		var model collectorFlowNodeConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Collector = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.FlowNodeConfigurationMemberCondition:
		var model conditionFlowNodeConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Condition = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.FlowNodeConfigurationMemberInlineCode:
		var model inlineCodeFlowNodeConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.InlineCode = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.FlowNodeConfigurationMemberInput:
		var model inputFlowNodeConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Input = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.FlowNodeConfigurationMemberIterator:
		var model iteratorFlowNodeConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Iterator = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.FlowNodeConfigurationMemberKnowledgeBase:
		var model knowledgeBaseFlowNodeConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.KnowledgeBase = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.FlowNodeConfigurationMemberLambdaFunction:
		var model lambdaFunctionFlowNodeConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.LambdaFunction = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.FlowNodeConfigurationMemberLex:
		var model lexFlowNodeConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Lex = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.FlowNodeConfigurationMemberOutput:
		var model outputFlowNodeConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Output = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.FlowNodeConfigurationMemberPrompt:
		var model promptFlowNodeConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Prompt = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.FlowNodeConfigurationMemberRetrieval:
		var model retrievalFlowNodeConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Retrieval = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.FlowNodeConfigurationMemberStorage:
		var model storageFlowNodeConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Storage = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	default:
		return diags
	}
}

func (m flowNodeConfigurationModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.Agent.IsNull():
		flowNodeConfigurationAgent, d := m.Agent.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.FlowNodeConfigurationMemberAgent
		diags.Append(fwflex.Expand(ctx, flowNodeConfigurationAgent, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	case !m.Collector.IsNull():
		flowNodeConfigurationCollector, d := m.Collector.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.FlowNodeConfigurationMemberCollector
		diags.Append(fwflex.Expand(ctx, flowNodeConfigurationCollector, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	case !m.Condition.IsNull():
		flowNodeConfigurationCondition, d := m.Condition.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.FlowNodeConfigurationMemberCondition
		diags.Append(fwflex.Expand(ctx, flowNodeConfigurationCondition, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	case !m.InlineCode.IsNull():
		flowNodeConfigurationInlineCode, d := m.InlineCode.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.FlowNodeConfigurationMemberInlineCode
		diags.Append(fwflex.Expand(ctx, flowNodeConfigurationInlineCode, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	case !m.Input.IsNull():
		flowNodeConfigurationInput, d := m.Input.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.FlowNodeConfigurationMemberInput
		diags.Append(fwflex.Expand(ctx, flowNodeConfigurationInput, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	case !m.Iterator.IsNull():
		flowNodeConfigurationIterator, d := m.Iterator.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.FlowNodeConfigurationMemberIterator
		diags.Append(fwflex.Expand(ctx, flowNodeConfigurationIterator, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	case !m.KnowledgeBase.IsNull():
		flowNodeConfigurationKnowledgeBase, d := m.KnowledgeBase.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.FlowNodeConfigurationMemberKnowledgeBase
		diags.Append(fwflex.Expand(ctx, flowNodeConfigurationKnowledgeBase, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	case !m.LambdaFunction.IsNull():
		flowNodeConfigurationLambdaFunction, d := m.LambdaFunction.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.FlowNodeConfigurationMemberLambdaFunction
		diags.Append(fwflex.Expand(ctx, flowNodeConfigurationLambdaFunction, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	case !m.Lex.IsNull():
		flowNodeConfigurationLex, d := m.Lex.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.FlowNodeConfigurationMemberLex
		diags.Append(fwflex.Expand(ctx, flowNodeConfigurationLex, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	case !m.Output.IsNull():
		flowNodeConfigurationOutput, d := m.Output.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.FlowNodeConfigurationMemberOutput
		diags.Append(fwflex.Expand(ctx, flowNodeConfigurationOutput, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	case !m.Prompt.IsNull():
		flowNodeConfigurationPrompt, d := m.Prompt.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.FlowNodeConfigurationMemberPrompt
		diags.Append(fwflex.Expand(ctx, flowNodeConfigurationPrompt, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	case !m.Retrieval.IsNull():
		flowNodeConfigurationRetrieval, d := m.Retrieval.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.FlowNodeConfigurationMemberRetrieval
		diags.Append(fwflex.Expand(ctx, flowNodeConfigurationRetrieval, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	case !m.Storage.IsNull():
		flowNodeConfigurationStorage, d := m.Storage.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.FlowNodeConfigurationMemberStorage
		diags.Append(fwflex.Expand(ctx, flowNodeConfigurationStorage, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

type agentFlowNodeConfigurationModel struct {
	AgentAliasARN fwtypes.ARN `tfsdk:"agent_alias_arn"`
}

type collectorFlowNodeConfigurationModel struct {
	// No fields
}

type conditionFlowNodeConfigurationModel struct {
	Conditions fwtypes.ListNestedObjectValueOf[flowConditionModel] `tfsdk:"condition"`
}

type flowConditionModel struct {
	Expression types.String `tfsdk:"expression"`
	Name       types.String `tfsdk:"name"`
}

type inlineCodeFlowNodeConfigurationModel struct {
	Code     types.String                                    `tfsdk:"code"`
	Language fwtypes.StringEnum[awstypes.SupportedLanguages] `tfsdk:"language"`
}

type inputFlowNodeConfigurationModel struct {
	// No fields
}

type iteratorFlowNodeConfigurationModel struct {
	// No fields
}

type knowledgeBaseFlowNodeConfigurationModel struct {
	GuardrailConfiguration fwtypes.ListNestedObjectValueOf[guardrailConfigurationModel]       `tfsdk:"guardrail_configuration"`
	InferenceConfiguration fwtypes.ListNestedObjectValueOf[promptInferenceConfigurationModel] `tfsdk:"inference_configuration"`
	KnowledgeBaseID        types.String                                                       `tfsdk:"knowledge_base_id"`
	ModelID                types.String                                                       `tfsdk:"model_id"`
	NumberOfResults        types.Int64                                                        `tfsdk:"number_of_results"`
	// TODO More fields.
}

type lambdaFunctionFlowNodeConfigurationModel struct {
	LambdaARN fwtypes.ARN `tfsdk:"lambda_arn"`
}

type lexFlowNodeConfigurationModel struct {
	BotAliasARN fwtypes.ARN  `tfsdk:"bot_alias_arn"`
	LocaleID    types.String `tfsdk:"locale_id"`
}

type outputFlowNodeConfigurationModel struct {
	// No fields
}

type promptFlowNodeConfigurationModel struct {
	GuardrailConfiguration fwtypes.ListNestedObjectValueOf[guardrailConfigurationModel]            `tfsdk:"guardrail_configuration"`
	SourceConfiguration    fwtypes.ListNestedObjectValueOf[promptFlowNodeSourceConfigurationModel] `tfsdk:"source_configuration"`
}

// Tagged union
type promptFlowNodeSourceConfigurationModel struct {
	Inline   fwtypes.ListNestedObjectValueOf[promptFlowNodeInlineConfigurationModel]   `tfsdk:"inline"`
	Resource fwtypes.ListNestedObjectValueOf[promptFlowNodeResourceConfigurationModel] `tfsdk:"resource"`
}

func (m *promptFlowNodeSourceConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.PromptFlowNodeSourceConfigurationMemberInline:
		var model promptFlowNodeInlineConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		if t.Value.AdditionalModelRequestFields != nil {
			json, err := tfsmithy.DocumentToJSONString(t.Value.AdditionalModelRequestFields)
			if err != nil {
				diags.Append(diag.NewErrorDiagnostic(
					"Encoding JSON",
					err.Error(),
				))

				return diags
			}

			model.AdditionalModelRequestFields = jsontypes.NewNormalizedValue(json)
		}

		m.Inline = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.PromptFlowNodeSourceConfigurationMemberResource:
		var model promptFlowNodeResourceConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Resource = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	default:
		return diags
	}
}

func (m promptFlowNodeSourceConfigurationModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.Inline.IsNull():
		promptFlowNodeSourceConfigurationInline, d := m.Inline.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.PromptFlowNodeSourceConfigurationMemberInline
		diags.Append(fwflex.Expand(ctx, promptFlowNodeSourceConfigurationInline, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		additionalFields := promptFlowNodeSourceConfigurationInline.AdditionalModelRequestFields
		if !additionalFields.IsNull() {
			json, err := tfsmithy.DocumentFromJSONString(fwflex.StringValueFromFramework(ctx, additionalFields), document.NewLazyDocument)
			if err != nil {
				diags.Append(diag.NewErrorDiagnostic(
					"Decoding JSON",
					err.Error(),
				))

				return nil, diags
			}

			r.Value.AdditionalModelRequestFields = json
		}

		return &r, diags
	case !m.Resource.IsNull():
		promptFlowNodeSourceConfigurationResource, d := m.Resource.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.PromptFlowNodeSourceConfigurationMemberResource
		diags.Append(fwflex.Expand(ctx, promptFlowNodeSourceConfigurationResource, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

type promptFlowNodeInlineConfigurationModel struct {
	AdditionalModelRequestFields jsontypes.Normalized                                               `tfsdk:"additional_model_request_fields" autoflex:"-"`
	InferenceConfiguration       fwtypes.ListNestedObjectValueOf[promptInferenceConfigurationModel] `tfsdk:"inference_configuration"`
	ModelID                      types.String                                                       `tfsdk:"model_id"`
	TemplateConfiguration        fwtypes.ListNestedObjectValueOf[promptTemplateConfigurationModel]  `tfsdk:"template_configuration"`
	TemplateType                 fwtypes.StringEnum[awstypes.PromptTemplateType]                    `tfsdk:"template_type"`
}

type promptFlowNodeResourceConfigurationModel struct {
	PromptARN fwtypes.ARN `tfsdk:"prompt_arn"`
}

type retrievalFlowNodeConfigurationModel struct {
	ServiceConfiguration fwtypes.ListNestedObjectValueOf[retrievalFlowNodeServiceConfigurationModel] `tfsdk:"service_configuration"`
}

// Tagged union
type retrievalFlowNodeServiceConfigurationModel struct {
	S3 fwtypes.ListNestedObjectValueOf[retrievalFlowNodeS3ConfigurationModel] `tfsdk:"s3"`
}

func (m *retrievalFlowNodeServiceConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.RetrievalFlowNodeServiceConfigurationMemberS3:
		var model retrievalFlowNodeS3ConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.S3 = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	default:
		return diags
	}
}

func (m retrievalFlowNodeServiceConfigurationModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.S3.IsNull():
		retrievalFlowNodeServiceConfigurationS3, d := m.S3.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.RetrievalFlowNodeServiceConfigurationMemberS3
		diags.Append(fwflex.Expand(ctx, retrievalFlowNodeServiceConfigurationS3, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

type retrievalFlowNodeS3ConfigurationModel struct {
	BucketName types.String `tfsdk:"bucket_name"`
}

type storageFlowNodeConfigurationModel struct {
	ServiceConfiguration fwtypes.ListNestedObjectValueOf[storageFlowNodeServiceConfigurationModel] `tfsdk:"service_configuration"`
}

// Tagged union
type storageFlowNodeServiceConfigurationModel struct {
	S3 fwtypes.ListNestedObjectValueOf[storageFlowNodeS3ConfigurationModel] `tfsdk:"s3"`
}

func (m *storageFlowNodeServiceConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.StorageFlowNodeServiceConfigurationMemberS3:
		var model storageFlowNodeS3ConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.S3 = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	default:
		return diags
	}
}

func (m storageFlowNodeServiceConfigurationModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.S3.IsNull():
		storageFlowNodeServiceConfigurationS3, d := m.S3.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.StorageFlowNodeServiceConfigurationMemberS3
		diags.Append(fwflex.Expand(ctx, storageFlowNodeServiceConfigurationS3, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

type storageFlowNodeS3ConfigurationModel struct {
	BucketName types.String `tfsdk:"bucket_name"`
}

type flowNodeInputModel struct {
	Category   fwtypes.StringEnum[awstypes.FlowNodeInputCategory] `tfsdk:"category"`
	Expression types.String                                       `tfsdk:"expression"`
	Name       types.String                                       `tfsdk:"name"`
	Type       fwtypes.StringEnum[awstypes.FlowNodeIODataType]    `tfsdk:"type"`
}

type flowNodeOutputModel struct {
	Name types.String                                    `tfsdk:"name"`
	Type fwtypes.StringEnum[awstypes.FlowNodeIODataType] `tfsdk:"type"`
}
