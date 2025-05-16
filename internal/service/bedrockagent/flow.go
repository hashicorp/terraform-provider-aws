// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent

import (
	"context"
	"encoding/json"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent/document"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagent_flow", name="Flow")
// @Tags(identifierAttribute="arn")
func newResourceFlow(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceFlow{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNameFlow = "Flow"
)

type resourceFlow struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceFlow) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.All(
						stringvalidator.LengthBetween(1, 50),
						stringvalidator.RegexMatches(regexp.MustCompile(`^[0-9A-Za-z-_]+$`), "must only contain alphanumeric characters, hyphens and underscores"),
					),
				},
			},
			"execution_role_arn": schema.StringAttribute{
				Required: true,
			},
			"customer_encryption_key_arn": schema.StringAttribute{
				Optional: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrVersion: schema.StringAttribute{
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.FlowStatus](),
				Computed:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
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
									"name": schema.StringAttribute{
										Required: true,
									},
									"source": schema.StringAttribute{
										Required: true,
									},
									"target": schema.StringAttribute{
										Required: true,
									},
									"type": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.FlowConnectionType](),
										Required:   true,
									},
								},
								Blocks: map[string]schema.Block{
									"configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[flowConnectionConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"data": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[flowConnectionConfigurationMemberDataModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
														listvalidator.ExactlyOneOf(
															path.MatchRelative().AtParent().AtName("data"),
															path.MatchRelative().AtParent().AtName("conditional"),
														),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"source_output": schema.StringAttribute{
																Required: true,
															},
															"target_input": schema.StringAttribute{
																Required: true,
															},
														},
													},
												},
												"conditional": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[flowConnectionConfigurationMemberConditionalModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"condition": schema.StringAttribute{
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
						"node": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[flowNodeModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Required: true,
									},
									"type": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.FlowNodeType](),
										Required:   true,
									},
								},
								Blocks: map[string]schema.Block{
									"configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[flowNodeConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"agent": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[flowNodeConfigurationMemberAgentModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
														listvalidator.ExactlyOneOf(
															path.MatchRelative().AtParent().AtName("agent"),
															path.MatchRelative().AtParent().AtName("collector"),
															path.MatchRelative().AtParent().AtName("condition"),
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
																Required: true,
															},
														},
													},
												},
												"collector": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[flowNodeConfigurationMemberCollectorModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
												},
												"condition": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[flowNodeConfigurationMemberConditionModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Blocks: map[string]schema.Block{
															"condition": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[flowConditionModel](ctx),
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"name": schema.StringAttribute{
																			Required: true,
																		},
																		"expression": schema.StringAttribute{
																			Optional: true,
																		},
																	},
																},
															},
														},
													},
												},
												"input": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[flowNodeConfigurationMemberInputModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
												},
												"iterator": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[flowNodeConfigurationMemberIteratorModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
												},
												"knowledge_base": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[flowNodeConfigurationMemberKnowledgeBaseModel](ctx),
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
														},
													},
												},
												"lambda_function": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[flowNodeConfigurationMemberLambdaFunctionModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"lambda_arn": schema.StringAttribute{
																Required: true,
															},
														},
													},
												},
												"lex": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[flowNodeConfigurationMemberLexModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"bot_alias_arn": schema.StringAttribute{
																Required: true,
															},
															"locale_id": schema.StringAttribute{
																Required: true,
															},
														},
													},
												},
												"output": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[flowNodeConfigurationMemberOutputModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
												},
												"prompt": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[flowNodeConfigurationMemberPromptModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Blocks: map[string]schema.Block{
															"source_configuration": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[promptFlowNodeSourceConfigurationModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Blocks: map[string]schema.Block{
																		"inline": schema.ListNestedBlock{
																			CustomType: fwtypes.NewListNestedObjectTypeOf[promptFlowNodeSourceConfigurationMemberInlineModel](ctx),
																			Validators: []validator.List{
																				listvalidator.SizeAtMost(1),
																				listvalidator.ExactlyOneOf(
																					path.MatchRelative().AtParent().AtName("inline"),
																					path.MatchRelative().AtParent().AtName("resource"),
																				),
																			},
																			NestedObject: schema.NestedBlockObject{
																				Attributes: map[string]schema.Attribute{
																					"model_id": schema.StringAttribute{
																						Required: true,
																					},
																					"template_type": schema.StringAttribute{
																						CustomType: fwtypes.StringEnumType[awstypes.PromptTemplateType](),
																						Required:   true,
																					},
																					"additional_model_request_fields": schema.StringAttribute{
																						CustomType: jsontypes.NormalizedType{},
																						Optional:   true,
																					},
																				},
																				Blocks: map[string]schema.Block{
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
																				},
																			},
																		},
																		"resource": schema.ListNestedBlock{
																			CustomType: fwtypes.NewListNestedObjectTypeOf[promptFlowNodeSourceConfigurationMemberResourceModel](ctx),
																			Validators: []validator.List{
																				listvalidator.SizeAtMost(1),
																			},
																			NestedObject: schema.NestedBlockObject{
																				Attributes: map[string]schema.Attribute{
																					"resource_arn": schema.StringAttribute{
																						Required: true,
																					},
																				},
																			},
																		},
																	},
																},
															},
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
														},
													},
												},
												"retrieval": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[flowNodeConfigurationMemberRetrievalModel](ctx),
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
																			CustomType: fwtypes.NewListNestedObjectTypeOf[retrievalFlowNodeServiceConfigurationMemberS3Model](ctx),
																			Validators: []validator.List{
																				listvalidator.SizeAtMost(1),
																				listvalidator.ExactlyOneOf(
																					path.MatchRelative().AtParent().AtName("s3"),
																				),
																			},
																			NestedObject: schema.NestedBlockObject{
																				Attributes: map[string]schema.Attribute{
																					"bucket_name": schema.StringAttribute{
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
													CustomType: fwtypes.NewListNestedObjectTypeOf[flowNodeConfigurationMemberStorageModel](ctx),
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
																			CustomType: fwtypes.NewListNestedObjectTypeOf[storageFlowNodeServiceConfigurationMemberS3Model](ctx),
																			Validators: []validator.List{
																				listvalidator.SizeAtMost(1),
																				listvalidator.ExactlyOneOf(
																					path.MatchRelative().AtParent().AtName("s3"),
																				),
																			},
																			NestedObject: schema.NestedBlockObject{
																				Attributes: map[string]schema.Attribute{
																					"bucket_name": schema.StringAttribute{
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
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"expression": schema.StringAttribute{
													Required: true,
												},
												"name": schema.StringAttribute{
													Required: true,
												},
												"type": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.FlowNodeIODataType](),
													Required:   true,
												},
											},
										},
									},
									"output": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[flowNodeOutputModel](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"name": schema.StringAttribute{
													Required: true,
												},
												"type": schema.StringAttribute{
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
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceFlow) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().BedrockAgentClient(ctx)

	var data resourceFlowModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	var input bedrockagent.CreateFlowInput
	response.Diagnostics.Append(flex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	input.ClientToken = aws.String(id.UniqueId())
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateFlow(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionCreating, ResNameFlow, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *resourceFlow) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BedrockAgentClient(ctx)

	var data resourceFlowModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findFlowByID(ctx, conn, data.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionSetting, ResNameFlow, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceFlow) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().BedrockAgentClient(ctx)

	var new, old resourceFlowModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &new)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, new, old)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input bedrockagent.UpdateFlowInput
		resp.Diagnostics.Append(flex.Expand(ctx, new, &input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		input.FlowIdentifier = new.ID.ValueStringPointer()

		output, err := conn.UpdateFlow(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionUpdating, ResNameFlow, new.ID.String(), err),
				err.Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, output, &new)...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Set values for unknowns.
		new.CreatedAt = timetypes.NewRFC3339TimePointerValue(output.CreatedAt)
		new.UpdatedAt = timetypes.NewRFC3339TimePointerValue(output.UpdatedAt)
		new.Version = flex.StringToFramework(ctx, output.Version)
		new.Status = fwtypes.StringEnumValue(output.Status)
	} else {
		new.CreatedAt = old.CreatedAt
		new.UpdatedAt = old.UpdatedAt
		new.Version = old.Version
		new.Status = old.Status
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &new)...)
}

func (r *resourceFlow) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BedrockAgentClient(ctx)

	var state resourceFlowModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := bedrockagent.DeleteFlowInput{
		FlowIdentifier: state.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteFlow(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionDeleting, ResNameFlow, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceFlow) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func findFlowByID(ctx context.Context, conn *bedrockagent.Client, id string) (*bedrockagent.GetFlowOutput, error) {
	in := &bedrockagent.GetFlowInput{
		FlowIdentifier: aws.String(id),
	}

	out, err := conn.GetFlow(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceFlowModel struct {
	ARN                      types.String                                         `tfsdk:"arn"`
	ID                       types.String                                         `tfsdk:"id"`
	Name                     types.String                                         `tfsdk:"name"`
	ExecutionRoleARN         types.String                                         `tfsdk:"execution_role_arn"`
	CustomerEncryptionKeyARN types.String                                         `tfsdk:"customer_encryption_key_arn"`
	Definition               fwtypes.ListNestedObjectValueOf[flowDefinitionModel] `tfsdk:"definition"`
	Description              types.String                                         `tfsdk:"description"`
	CreatedAt                timetypes.RFC3339                                    `tfsdk:"created_at"`
	UpdatedAt                timetypes.RFC3339                                    `tfsdk:"updated_at"`
	Version                  types.String                                         `tfsdk:"version"`
	Status                   fwtypes.StringEnum[awstypes.FlowStatus]              `tfsdk:"status"`
	Tags                     tftags.Map                                           `tfsdk:"tags"`
	TagsAll                  tftags.Map                                           `tfsdk:"tags_all"`
	Timeouts                 timeouts.Value                                       `tfsdk:"timeouts"`
}

type flowDefinitionModel struct {
	Connections fwtypes.ListNestedObjectValueOf[flowConnectionModel] `tfsdk:"connection"`
	Nodes       fwtypes.ListNestedObjectValueOf[flowNodeModel]       `tfsdk:"node"`
}

type flowConnectionModel struct {
	Name          types.String                                                      `tfsdk:"name"`
	Source        types.String                                                      `tfsdk:"source"`
	Target        types.String                                                      `tfsdk:"target"`
	Type          fwtypes.StringEnum[awstypes.FlowConnectionType]                   `tfsdk:"type"`
	Configuration fwtypes.ListNestedObjectValueOf[flowConnectionConfigurationModel] `tfsdk:"configuration"`
}

// Tagged union
type flowConnectionConfigurationModel struct {
	Data        fwtypes.ListNestedObjectValueOf[flowConnectionConfigurationMemberDataModel]        `tfsdk:"data"`
	Conditional fwtypes.ListNestedObjectValueOf[flowConnectionConfigurationMemberConditionalModel] `tfsdk:"conditional"`
}

type flowConnectionConfigurationMemberConditionalModel struct {
	Condition types.String `tfsdk:"condition"`
}

type flowConnectionConfigurationMemberDataModel struct {
	SourceOutput types.String `tfsdk:"source_output"`
	TargetInput  types.String `tfsdk:"target_input"`
}

func (m *flowConnectionConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.FlowConnectionConfigurationMemberData:
		var model flowConnectionConfigurationMemberDataModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Data = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.FlowConnectionConfigurationMemberConditional:
		var model flowConnectionConfigurationMemberConditionalModel
		d := flex.Flatten(ctx, t.Value, &model)
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
		diags.Append(flex.Expand(ctx, flowConnectionConfigurationData, &r.Value)...)
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
		diags.Append(flex.Expand(ctx, flowConnectionConfigurationConditional, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

type flowNodeModel struct {
	Name          types.String                                                `tfsdk:"name"`
	Type          fwtypes.StringEnum[awstypes.FlowNodeType]                   `tfsdk:"type"`
	Configuration fwtypes.ListNestedObjectValueOf[flowNodeConfigurationModel] `tfsdk:"configuration"`
	Inputs        fwtypes.ListNestedObjectValueOf[flowNodeInputModel]         `tfsdk:"input"`
	Outputs       fwtypes.ListNestedObjectValueOf[flowNodeOutputModel]        `tfsdk:"output"`
}

// Tagged union
type flowNodeConfigurationModel struct {
	Agent          fwtypes.ListNestedObjectValueOf[flowNodeConfigurationMemberAgentModel]          `tfsdk:"agent"`
	Collector      fwtypes.ListNestedObjectValueOf[flowNodeConfigurationMemberCollectorModel]      `tfsdk:"collector"`
	Condition      fwtypes.ListNestedObjectValueOf[flowNodeConfigurationMemberConditionModel]      `tfsdk:"condition"`
	Input          fwtypes.ListNestedObjectValueOf[flowNodeConfigurationMemberInputModel]          `tfsdk:"input"`
	Iterator       fwtypes.ListNestedObjectValueOf[flowNodeConfigurationMemberIteratorModel]       `tfsdk:"iterator"`
	KnowledgeBase  fwtypes.ListNestedObjectValueOf[flowNodeConfigurationMemberKnowledgeBaseModel]  `tfsdk:"knowledge_base"`
	LambdaFunction fwtypes.ListNestedObjectValueOf[flowNodeConfigurationMemberLambdaFunctionModel] `tfsdk:"lambda_function"`
	Lex            fwtypes.ListNestedObjectValueOf[flowNodeConfigurationMemberLexModel]            `tfsdk:"lex"`
	Output         fwtypes.ListNestedObjectValueOf[flowNodeConfigurationMemberOutputModel]         `tfsdk:"output"`
	Prompt         fwtypes.ListNestedObjectValueOf[flowNodeConfigurationMemberPromptModel]         `tfsdk:"prompt"`
	Retrieval      fwtypes.ListNestedObjectValueOf[flowNodeConfigurationMemberRetrievalModel]      `tfsdk:"retrieval"`
	Storage        fwtypes.ListNestedObjectValueOf[flowNodeConfigurationMemberStorageModel]        `tfsdk:"storage"`
}

func (m *flowNodeConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.FlowNodeConfigurationMemberAgent:
		var model flowNodeConfigurationMemberAgentModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Agent = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.FlowNodeConfigurationMemberCollector:
		var model flowNodeConfigurationMemberCollectorModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Collector = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.FlowNodeConfigurationMemberCondition:
		var model flowNodeConfigurationMemberConditionModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Condition = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.FlowNodeConfigurationMemberInput:
		var model flowNodeConfigurationMemberInputModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Input = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.FlowNodeConfigurationMemberIterator:
		var model flowNodeConfigurationMemberIteratorModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Iterator = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.FlowNodeConfigurationMemberKnowledgeBase:
		var model flowNodeConfigurationMemberKnowledgeBaseModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.KnowledgeBase = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.FlowNodeConfigurationMemberLambdaFunction:
		var model flowNodeConfigurationMemberLambdaFunctionModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.LambdaFunction = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.FlowNodeConfigurationMemberLex:
		var model flowNodeConfigurationMemberLexModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Lex = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.FlowNodeConfigurationMemberOutput:
		var model flowNodeConfigurationMemberOutputModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Output = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.FlowNodeConfigurationMemberPrompt:
		var model flowNodeConfigurationMemberPromptModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Prompt = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.FlowNodeConfigurationMemberRetrieval:
		var model flowNodeConfigurationMemberRetrievalModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Retrieval = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.FlowNodeConfigurationMemberStorage:
		var model flowNodeConfigurationMemberStorageModel
		d := flex.Flatten(ctx, t.Value, &model)
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
		diags.Append(flex.Expand(ctx, flowNodeConfigurationAgent, &r.Value)...)
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
		diags.Append(flex.Expand(ctx, flowNodeConfigurationCollector, &r.Value)...)
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
		diags.Append(flex.Expand(ctx, flowNodeConfigurationCondition, &r.Value)...)
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
		diags.Append(flex.Expand(ctx, flowNodeConfigurationInput, &r.Value)...)
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
		diags.Append(flex.Expand(ctx, flowNodeConfigurationIterator, &r.Value)...)
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
		diags.Append(flex.Expand(ctx, flowNodeConfigurationKnowledgeBase, &r.Value)...)
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
		diags.Append(flex.Expand(ctx, flowNodeConfigurationLambdaFunction, &r.Value)...)
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
		diags.Append(flex.Expand(ctx, flowNodeConfigurationLex, &r.Value)...)
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
		diags.Append(flex.Expand(ctx, flowNodeConfigurationOutput, &r.Value)...)
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
		diags.Append(flex.Expand(ctx, flowNodeConfigurationPrompt, &r.Value)...)
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
		diags.Append(flex.Expand(ctx, flowNodeConfigurationRetrieval, &r.Value)...)
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
		diags.Append(flex.Expand(ctx, flowNodeConfigurationStorage, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

type flowNodeConfigurationMemberAgentModel struct {
	AgentAliasARN types.String `tfsdk:"agent_alias_arn"`
}

type flowNodeConfigurationMemberCollectorModel struct {
	// No fields
}

type flowNodeConfigurationMemberConditionModel struct {
	Conditions fwtypes.ListNestedObjectValueOf[flowConditionModel] `tfsdk:"condition"`
}

type flowConditionModel struct {
	Name       types.String `tfsdk:"name"`
	Expression types.String `tfsdk:"expression"`
}

type flowNodeConfigurationMemberInputModel struct {
	// No fields
}

type flowNodeConfigurationMemberIteratorModel struct {
	// No fields
}

type flowNodeConfigurationMemberKnowledgeBaseModel struct {
	KnowledgeBaseID        types.String                                                 `tfsdk:"knowledge_base_id"`
	GuardrailConfiguration fwtypes.ListNestedObjectValueOf[guardrailConfigurationModel] `tfsdk:"guardrail_configuration"`
	ModelID                types.String                                                 `tfsdk:"model_id"`
}

type flowNodeConfigurationMemberLambdaFunctionModel struct {
	LambdaARN types.String `tfsdk:"lambda_arn"`
}

type flowNodeConfigurationMemberLexModel struct {
	BotAliasARN types.String `tfsdk:"bot_alias_arn"`
	LocaleID    types.String `tfsdk:"locale_id"`
}

type flowNodeConfigurationMemberOutputModel struct {
	// No fields
}

type flowNodeConfigurationMemberPromptModel struct {
	SourceConfiguration    fwtypes.ListNestedObjectValueOf[promptFlowNodeSourceConfigurationModel] `tfsdk:"source_configuration"`
	GuardrailConfiguration fwtypes.ListNestedObjectValueOf[guardrailConfigurationModel]            `tfsdk:"guardrail_configuration"`
}

// Tagged union
type promptFlowNodeSourceConfigurationModel struct {
	Inline   fwtypes.ListNestedObjectValueOf[promptFlowNodeSourceConfigurationMemberInlineModel]   `tfsdk:"inline"`
	Resource fwtypes.ListNestedObjectValueOf[promptFlowNodeSourceConfigurationMemberResourceModel] `tfsdk:"resource"`
}

func (m *promptFlowNodeSourceConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.PromptFlowNodeSourceConfigurationMemberInline:
		var model promptFlowNodeSourceConfigurationMemberInlineModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		if t.Value.AdditionalModelRequestFields != nil {
			additionalFields, err := t.Value.AdditionalModelRequestFields.MarshalSmithyDocument()
			if err != nil {
				diags.AddError("Marshalling additional model request fields", err.Error())
				return diags
			}

			model.AdditionalModelRequestFields = types.StringValue(string(additionalFields))
		}

		m.Inline = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.PromptFlowNodeSourceConfigurationMemberResource:
		var model promptFlowNodeSourceConfigurationMemberResourceModel
		d := flex.Flatten(ctx, t.Value, &model)
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
		diags.Append(flex.Expand(ctx, promptFlowNodeSourceConfigurationInline, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		additionalFields := promptFlowNodeSourceConfigurationInline.AdditionalModelRequestFields
		if !additionalFields.IsNull() {
			var doc any
			if err := json.Unmarshal([]byte(additionalFields.ValueString()), &doc); err != nil {
				diags.AddError("Unmarshalling additional model request fields", err.Error())
				return nil, diags
			}

			r.Value.AdditionalModelRequestFields = document.NewLazyDocument(doc)
		}

		return &r, diags
	case !m.Resource.IsNull():
		promptFlowNodeSourceConfigurationResource, d := m.Resource.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.PromptFlowNodeSourceConfigurationMemberResource
		diags.Append(flex.Expand(ctx, promptFlowNodeSourceConfigurationResource, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

type promptFlowNodeSourceConfigurationMemberInlineModel struct {
	ModelID                      types.String                                                       `tfsdk:"model_id"`
	TemplateConfiguration        fwtypes.ListNestedObjectValueOf[promptTemplateConfigurationModel]  `tfsdk:"template_configuration"`
	TemplateType                 fwtypes.StringEnum[awstypes.PromptTemplateType]                    `tfsdk:"template_type"`
	AdditionalModelRequestFields types.String                                                       `tfsdk:"additional_model_request_fields"`
	InferenceConfiguration       fwtypes.ListNestedObjectValueOf[promptInferenceConfigurationModel] `tfsdk:"inference_configuration"`
}

type promptFlowNodeSourceConfigurationMemberResourceModel struct {
	ResourceARN types.String `tfsdk:"resource_arn"`
}

type flowNodeConfigurationMemberRetrievalModel struct {
	ServiceConfiguration fwtypes.ListNestedObjectValueOf[retrievalFlowNodeServiceConfigurationModel] `tfsdk:"service_configuration"`
}

// Tagged union
type retrievalFlowNodeServiceConfigurationModel struct {
	S3 fwtypes.ListNestedObjectValueOf[retrievalFlowNodeServiceConfigurationMemberS3Model] `tfsdk:"s3"`
}

func (m *retrievalFlowNodeServiceConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.RetrievalFlowNodeServiceConfigurationMemberS3:
		var model retrievalFlowNodeServiceConfigurationMemberS3Model
		d := flex.Flatten(ctx, t.Value, &model)
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
		diags.Append(flex.Expand(ctx, retrievalFlowNodeServiceConfigurationS3, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

type retrievalFlowNodeServiceConfigurationMemberS3Model struct {
	BucketName types.String `tfsdk:"bucket_name"`
}

type flowNodeConfigurationMemberStorageModel struct {
	ServiceConfiguration fwtypes.ListNestedObjectValueOf[storageFlowNodeServiceConfigurationModel] `tfsdk:"service_configuration"`
}

// Tagged union
type storageFlowNodeServiceConfigurationModel struct {
	S3 fwtypes.ListNestedObjectValueOf[storageFlowNodeServiceConfigurationMemberS3Model] `tfsdk:"s3"`
}

func (m *storageFlowNodeServiceConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.StorageFlowNodeServiceConfigurationMemberS3:
		var model storageFlowNodeServiceConfigurationMemberS3Model
		d := flex.Flatten(ctx, t.Value, &model)
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
		diags.Append(flex.Expand(ctx, storageFlowNodeServiceConfigurationS3, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

type storageFlowNodeServiceConfigurationMemberS3Model struct {
	BucketName types.String `tfsdk:"bucket_name"`
}

type flowNodeInputModel struct {
	Expression types.String                                    `tfsdk:"expression"`
	Name       types.String                                    `tfsdk:"name"`
	Type       fwtypes.StringEnum[awstypes.FlowNodeIODataType] `tfsdk:"type"`
}

type flowNodeOutputModel struct {
	Name types.String                                    `tfsdk:"name"`
	Type fwtypes.StringEnum[awstypes.FlowNodeIODataType] `tfsdk:"type"`
}
