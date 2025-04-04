// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock/document"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagent_flow", name="Flow")
func newResourceFlow(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceFlow{}

	// TIP: ==== CONFIGURABLE TIMEOUTS ====
	// Users can configure timeout lengths but you need to use the times they
	// provide. Access the timeout they configure (or the defaults) using,
	// e.g., r.CreateTimeout(ctx, plan.Timeouts) (see below). The times here are
	// the defaults if they don't configure timeouts.
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameFlow = "Flow"
)

type resourceFlow struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

// TIP: ==== SCHEMA ====
// In the schema, add each of the attributes in snake case (e.g.,
// delete_automated_backups).
//
// Formatting rules:
// * Alphabetize attributes to make them easier to find.
// * Do not add a blank line between attributes.
//
// Attribute basics:
//   - If a user can provide a value ("configure a value") for an
//     attribute (e.g., instances = 5), we call the attribute an
//     "argument."
//   - You change the way users interact with attributes using:
//   - Required
//   - Optional
//   - Computed
//   - There are only four valid combinations:
//
// 1. Required only - the user must provide a value
// Required: true,
//
//  2. Optional only - the user can configure or omit a value; do not
//     use Default or DefaultFunc
//
// Optional: true,
//
//  3. Computed only - the provider can provide a value but the user
//     cannot, i.e., read-only
//
// Computed: true,
//
//  4. Optional AND Computed - the provider or user can provide a value;
//     use this combination if you are using Default
//
// Optional: true,
// Computed: true,
//
// You will typically find arguments in the input struct
// (e.g., CreateDBInstanceInput) for the create operation. Sometimes
// they are only in the input struct (e.g., ModifyDBInstanceInput) for
// the modify operation.
//
// For more about schema options, visit
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/schemas?page=schemas
func (r *resourceFlow) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"id":  framework.IDAttribute(),
			"name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.All(
						stringvalidator.LengthBetween(1, 50),
						stringvalidator.RegexMatches(regexp.MustCompile(`^[0-9A-Za-z_]+$`), "the name should only contain 0-9, A-Z, a-z, and _"),
					),
				},
			},
			"execution_role_arn": schema.StringAttribute{
				Required: true,
			},
			"customer_encryption_key_arn": schema.StringAttribute{
				Optional: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
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
																						CustomType: fwtypes.NewSmithyJSONType(ctx, document.NewLazyDocument),
																						Optional:   true,
																					},
																				},
																				Blocks: map[string]schema.Block{
																					"template_configuration": schema.ListNestedBlock{
																						CustomType: fwtypes.NewListNestedObjectTypeOf[templateConfigurationModel](ctx),
																						Validators: []validator.List{
																							listvalidator.SizeAtMost(1),
																						},
																						NestedObject: schema.NestedBlockObject{
																							Blocks: map[string]schema.Block{
																								"chat": schema.ListNestedBlock{
																									CustomType: fwtypes.NewListNestedObjectTypeOf[promptTemplateConfigurationMemberChatModel](ctx),
																									Validators: []validator.List{
																										listvalidator.SizeAtMost(1),
																										listvalidator.ExactlyOneOf(
																											path.MatchRelative().AtParent().AtName("chat"),
																											path.MatchRelative().AtParent().AtName("text"),
																										),
																									},
																									NestedObject: schema.NestedBlockObject{
																										Blocks: map[string]schema.Block{
																											"message": schema.ListNestedBlock{
																												CustomType: fwtypes.NewListNestedObjectTypeOf[messageModel](ctx),
																												NestedObject: schema.NestedBlockObject{
																													Attributes: map[string]schema.Attribute{
																														"role": schema.StringAttribute{
																															CustomType: fwtypes.StringEnumType[awstypes.ConversationRole](),
																															Required:   true,
																														},
																													},
																													Blocks: map[string]schema.Block{
																														"content": schema.ListNestedBlock{
																															CustomType: fwtypes.NewListNestedObjectTypeOf[contentBlockModel](ctx),
																															Validators: []validator.List{
																																listvalidator.SizeAtMost(1),
																															},
																															NestedObject: schema.NestedBlockObject{
																																Blocks: map[string]schema.Block{
																																	"cache_point": schema.ListNestedBlock{
																																		CustomType: fwtypes.NewListNestedObjectTypeOf[contentBlockMemberCachePointModel](ctx),
																																		Validators: []validator.List{
																																			listvalidator.SizeAtMost(1),
																																			listvalidator.ExactlyOneOf(
																																				path.MatchRelative().AtParent().AtName("cache_point"),
																																				path.MatchRelative().AtParent().AtName("text"),
																																			),
																																		},
																																		NestedObject: schema.NestedBlockObject{
																																			Attributes: map[string]schema.Attribute{
																																				"type": schema.StringAttribute{
																																					CustomType: fwtypes.StringEnumType[awstypes.CachePointType](),
																																					Required:   true,
																																				},
																																			},
																																		},
																																	},
																																	"text": schema.ListNestedBlock{
																																		CustomType: fwtypes.NewListNestedObjectTypeOf[contentBlockMemberTextModel](ctx),
																																		Validators: []validator.List{
																																			listvalidator.SizeAtMost(1),
																																		},
																																		NestedObject: schema.NestedBlockObject{
																																			Attributes: map[string]schema.Attribute{
																																				"value": schema.StringAttribute{
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
																											"input_variable": schema.ListNestedBlock{
																												CustomType: fwtypes.NewListNestedObjectTypeOf[promptInputVariableModel](ctx),
																												NestedObject: schema.NestedBlockObject{
																													Attributes: map[string]schema.Attribute{
																														"name": schema.StringAttribute{
																															Required: true,
																														},
																													},
																												},
																											},
																											"system": schema.ListNestedBlock{
																												CustomType: fwtypes.NewListNestedObjectTypeOf[systemContentBlockModel](ctx),
																												NestedObject: schema.NestedBlockObject{
																													Blocks: map[string]schema.Block{
																														"cache_point": schema.ListNestedBlock{
																															CustomType: fwtypes.NewListNestedObjectTypeOf[systemContentBlockMemberCachePointModel](ctx),
																															Validators: []validator.List{
																																listvalidator.SizeAtMost(1),
																															},
																															NestedObject: schema.NestedBlockObject{
																																Attributes: map[string]schema.Attribute{
																																	"type": schema.StringAttribute{
																																		CustomType: fwtypes.StringEnumType[awstypes.CachePointType](),
																																		Required:   true,
																																	},
																																},
																															},
																														},
																														"text": schema.ListNestedBlock{
																															CustomType: fwtypes.NewListNestedObjectTypeOf[systemContentBlockMemberTextModel](ctx),
																															Validators: []validator.List{
																																listvalidator.SizeAtMost(1),
																															},
																															NestedObject: schema.NestedBlockObject{
																																Attributes: map[string]schema.Attribute{
																																	"value": schema.StringAttribute{
																																		Required: true,
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
																																		CustomType: fwtypes.NewListNestedObjectTypeOf[toolMemberCachePointModel](ctx),
																																		Validators: []validator.List{
																																			listvalidator.SizeAtMost(1),
																																			listvalidator.ExactlyOneOf(
																																				path.MatchRelative().AtParent().AtName("cache_point"),
																																				path.MatchRelative().AtParent().AtName("tool_spec"),
																																			),
																																		},
																																		NestedObject: schema.NestedBlockObject{
																																			Attributes: map[string]schema.Attribute{
																																				"type": schema.StringAttribute{
																																					CustomType: fwtypes.StringEnumType[awstypes.CachePointType](),
																																					Required:   true,
																																				},
																																			},
																																		},
																																	},
																																	"tool_spec": schema.ListNestedBlock{
																																		CustomType: fwtypes.NewListNestedObjectTypeOf[toolMemberToolSpecModel](ctx),
																																		Validators: []validator.List{
																																			listvalidator.SizeAtMost(1),
																																		},
																																		NestedObject: schema.NestedBlockObject{
																																			Attributes: map[string]schema.Attribute{
																																				"name": schema.StringAttribute{
																																					Required: true,
																																				},
																																				"description": schema.StringAttribute{
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
																																						Blocks: map[string]schema.Block{
																																							"json": schema.ListNestedBlock{
																																								CustomType: fwtypes.NewListNestedObjectTypeOf[toolInputSchemaMemberJsonModel](ctx),
																																								Validators: []validator.List{
																																									listvalidator.SizeAtMost(1),
																																									listvalidator.ExactlyOneOf(
																																										path.MatchRelative().AtParent().AtName("json"),
																																									),
																																								},
																																								NestedObject: schema.NestedBlockObject{
																																									Attributes: map[string]schema.Attribute{
																																										"value": schema.StringAttribute{
																																											CustomType: fwtypes.NewSmithyJSONType(ctx, document.NewLazyDocument),
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
																																		CustomType: fwtypes.NewListNestedObjectTypeOf[toolChoiceMemberAnyModel](ctx),
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
																																		CustomType: fwtypes.NewListNestedObjectTypeOf[toolChoiceMemberAutoModel](ctx),
																																		Validators: []validator.List{
																																			listvalidator.SizeAtMost(1),
																																		},
																																	},
																																	"tool": schema.ListNestedBlock{
																																		CustomType: fwtypes.NewListNestedObjectTypeOf[toolChoiceMemberToolModel](ctx),
																																		Validators: []validator.List{
																																			listvalidator.SizeAtMost(1),
																																		},
																																		NestedObject: schema.NestedBlockObject{
																																			Attributes: map[string]schema.Attribute{
																																				"name": schema.StringAttribute{
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
																									CustomType: fwtypes.NewListNestedObjectTypeOf[promptTemplateConfigurationMemberTextModel](ctx),
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
																														"type": schema.StringAttribute{
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
																														"name": schema.StringAttribute{
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
																									CustomType: fwtypes.NewListNestedObjectTypeOf[promptInferenceConfigurationMemberText](ctx),
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

func (r *resourceFlow) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().BedrockAgentClient(ctx)

	// TIP: -- 2. Fetch the plan
	var plan resourceFlowModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a Create input structure
	var input bedrockagent.CreateFlowInput
	// TIP: Using a field name prefix allows mapping fields such as `ID` to `FlowId`
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 4. Call the AWS Create function
	out, err := conn.CreateFlow(ctx, &input)
	if err != nil {
		// TIP: Since ID has not been set yet, you cannot use plan.ID.String()
		// in error messages at this point.
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionCreating, ResNameFlow, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionCreating, ResNameFlow, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	// TIP: -- 5. Using the output from the create function, set attributes
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 6. Use a waiter to wait for create to complete
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitFlowCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionWaitingForCreation, ResNameFlow, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 7. Save the request plan to response state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceFlow) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().BedrockAgentClient(ctx)

	// TIP: -- 2. Fetch the state
	var state resourceFlowModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Get the resource from AWS using an API Get, List, or Describe-
	// type function, or, better yet, using a finder.
	out, err := findFlowByID(ctx, conn, state.ID.ValueString())
	// TIP: -- 4. Remove resource from state if it is not found
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionSetting, ResNameFlow, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 5. Set the arguments and attributes
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 6. Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceFlow) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().BedrockAgentClient(ctx)

	// TIP: -- 2. Fetch the plan
	var plan, state resourceFlowModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Get the difference between the plan and state, if any
	diff, d := flex.Diff(ctx, plan, state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input bedrockagent.UpdateFlowInput
		resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		input.FlowIdentifier = plan.ID.ValueStringPointer()

		// TIP: -- 4. Call the AWS modify/update function
		out, err := conn.UpdateFlow(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionUpdating, ResNameFlow, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionUpdating, ResNameFlow, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		// TIP: Using the output from the update function, re-set any computed attributes
		resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// TIP: -- 5. Use a waiter to wait for update to complete
	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitFlowUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionWaitingForUpdate, ResNameFlow, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 6. Save the request plan to response state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceFlow) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BedrockAgentClient(ctx)

	// TIP: -- 2. Fetch the state
	var state resourceFlowModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a delete input structure
	input := bedrockagent.DeleteFlowInput{
		FlowIdentifier: state.ID.ValueStringPointer(),
	}

	// TIP: -- 4. Call the AWS delete function
	_, err := conn.DeleteFlow(ctx, &input)
	// TIP: On rare occassions, the API returns a not found error after deleting a
	// resource. If that happens, we don't want it to show up as an error.
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

// TIP: ==== TERRAFORM IMPORTING ====
// If Read can get all the information it needs from the Identifier
// (i.e., path.Root("id")), you can use the PassthroughID importer. Otherwise,
// you'll need a custom import function.
//
// See more:
// https://developer.hashicorp.com/terraform/plugin/framework/resources/import
func (r *resourceFlow) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// TIP: ==== WAITERS ====
// Some resources of some services have waiters provided by the AWS API.
// Unless they do not work properly, use them rather than defining new ones
// here.
//
// Sometimes we define the wait, status, and find functions in separate
// files, wait.go, status.go, and find.go. Follow the pattern set out in the
// service and define these where it makes the most sense.
//
// If these functions are used in the _test.go file, they will need to be
// exported (i.e., capitalized).
//
// You will need to adjust the parameters and names to fit the service.
func waitFlowCreated(ctx context.Context, conn *bedrockagent.Client, id string, timeout time.Duration) (*bedrockagent.GetFlowOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(awstypes.FlowStatusNotPrepared),
		Refresh:                   statusFlow(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagent.GetFlowOutput); ok {
		return out, err
	}

	return nil, err
}

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.
func waitFlowUpdated(ctx context.Context, conn *bedrockagent.Client, id string, timeout time.Duration) (*bedrockagent.GetFlowOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(awstypes.FlowStatusNotPrepared),
		Refresh:                   statusFlow(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagent.GetFlowOutput); ok {
		return out, err
	}

	return nil, err
}

// TIP: ==== STATUS ====
// The status function can return an actual status when that field is
// available from the API (e.g., out.Status). Otherwise, you can use custom
// statuses to communicate the states of the resource.
//
// Waiters consume the values returned by status functions. Design status so
// that it can be reused by a create, update, and delete waiter, if possible.
func statusFlow(ctx context.Context, conn *bedrockagent.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findFlowByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

// TIP: ==== FINDERS ====
// The find function is not strictly necessary. You could do the API
// request from the status function. However, we have found that find often
// comes in handy in other places besides the status function. As a result, it
// is good practice to define it separately.
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

// TIP: ==== DATA STRUCTURES ====
// With Terraform Plugin-Framework configurations are deserialized into
// Go types, providing type safety without the need for type assertions.
// These structs should match the schema definition exactly, and the `tfsdk`
// tag value should match the attribute name.
//
// Nested objects are represented in their own data struct. These will
// also have a corresponding attribute type mapping for use inside flex
// functions.
//
// See more:
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/accessing-values
type resourceFlowModel struct {
	ARN                      types.String                                         `tfsdk:"arn"`
	ID                       types.String                                         `tfsdk:"id"`
	Name                     types.String                                         `tfsdk:"name"`
	ExecutionRoleARN         types.String                                         `tfsdk:"execution_role_arn"`
	CustomerEncryptionKeyARN types.String                                         `tfsdk:"customer_encryption_key_arn"`
	Definition               fwtypes.ListNestedObjectValueOf[flowDefinitionModel] `tfsdk:"definition"`
	Description              types.String                                         `tfsdk:"description"`
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
	TemplateConfiguration        fwtypes.ListNestedObjectValueOf[templateConfigurationModel]        `tfsdk:"template_configuration"`
	TemplateType                 fwtypes.StringEnum[awstypes.PromptTemplateType]                    `tfsdk:"template_type"`
	AdditionalModelRequestFields fwtypes.SmithyJSON[document.Interface]                             `tfsdk:"additional_model_request_fields"`
	InferenceConfiguration       fwtypes.ListNestedObjectValueOf[promptInferenceConfigurationModel] `tfsdk:"inference_configuration"`
}

// Tagged union
type promptInferenceConfigurationModel struct {
	Text fwtypes.ListNestedObjectValueOf[promptInferenceConfigurationMemberText] `tfsdk:"text"`
}

func (m *promptInferenceConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.PromptInferenceConfigurationMemberText:
		var model promptInferenceConfigurationMemberText
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Text = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	default:
		return diags
	}
}

func (m promptInferenceConfigurationModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.Text.IsNull():
		promptInferenceConfigurationText, d := m.Text.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.PromptInferenceConfigurationMemberText
		diags.Append(flex.Expand(ctx, promptInferenceConfigurationText, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

type promptInferenceConfigurationMemberText struct {
	MaxTokens     types.Int32          `tfsdk:"max_tokens"`
	StopSequences fwtypes.ListOfString `tfsdk:"stop_sequences"`
	Temperature   types.Float32        `tfsdk:"temperature"`
	TopP          types.Float32        `tfsdk:"top_p"`
}

// Tagged union
type templateConfigurationModel struct {
	Chat fwtypes.ListNestedObjectValueOf[promptTemplateConfigurationMemberChatModel] `tfsdk:"chat"`
	Text fwtypes.ListNestedObjectValueOf[promptTemplateConfigurationMemberTextModel] `tfsdk:"text"`
}

func (m *templateConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.PromptTemplateConfigurationMemberChat:
		var model promptTemplateConfigurationMemberChatModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Chat = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.PromptTemplateConfigurationMemberText:
		var model promptTemplateConfigurationMemberTextModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Text = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	default:
		return diags
	}
}

func (m templateConfigurationModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.Chat.IsNull():
		promptTemplateConfigurationChat, d := m.Chat.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.PromptTemplateConfigurationMemberChat
		diags.Append(flex.Expand(ctx, promptTemplateConfigurationChat, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	case !m.Text.IsNull():
		promptTemplateConfigurationText, d := m.Text.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.PromptTemplateConfigurationMemberText
		diags.Append(flex.Expand(ctx, promptTemplateConfigurationText, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

type promptTemplateConfigurationMemberChatModel struct {
	Messages          fwtypes.ListNestedObjectValueOf[messageModel]             `tfsdk:"message"`
	InputVariables    fwtypes.ListNestedObjectValueOf[promptInputVariableModel] `tfsdk:"input_variable"`
	System            fwtypes.ListNestedObjectValueOf[systemContentBlockModel]  `tfsdk:"system"`
	ToolConfiguration fwtypes.ListNestedObjectValueOf[toolConfigurationModel]   `tfsdk:"tool_configuration"`
}

type messageModel struct {
	Content fwtypes.ListNestedObjectValueOf[contentBlockModel] `tfsdk:"content"`
	Role    fwtypes.StringEnum[awstypes.ConversationRole]      `tfsdk:"role"`
}

// Tagged union
type contentBlockModel struct {
	CachePoint fwtypes.ListNestedObjectValueOf[contentBlockMemberCachePointModel] `tfsdk:"cache_point"`
	Text       fwtypes.ListNestedObjectValueOf[contentBlockMemberTextModel]       `tfsdk:"text"`
}

func (m *contentBlockModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.ContentBlockMemberCachePoint:
		var model contentBlockMemberCachePointModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.CachePoint = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.ContentBlockMemberText:
		var model contentBlockMemberTextModel
		d := flex.Flatten(ctx, t.Value, &model.Value)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Text = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	default:
		return diags
	}
}

func (m contentBlockModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.CachePoint.IsNull():
		contentBlockCachePoint, d := m.CachePoint.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.ContentBlockMemberCachePoint
		diags.Append(flex.Expand(ctx, contentBlockCachePoint, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	case !m.Text.IsNull():
		contentBlockText, d := m.Text.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.ContentBlockMemberText
		diags.Append(flex.Expand(ctx, contentBlockText.Value, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

type contentBlockMemberCachePointModel struct {
	Type fwtypes.StringEnum[awstypes.CachePointType] `tfsdk:"type"`
}

type contentBlockMemberTextModel struct {
	Value types.String `tfsdk:"value"`
}

type promptInputVariableModel struct {
	Name types.String `tfsdk:"name"`
}

// Tagged union
type systemContentBlockModel struct {
	CachePoint fwtypes.ListNestedObjectValueOf[systemContentBlockMemberCachePointModel] `tfsdk:"cache_point"`
	Text       fwtypes.ListNestedObjectValueOf[systemContentBlockMemberTextModel]       `tfsdk:"text"`
}

func (m *systemContentBlockModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.SystemContentBlockMemberCachePoint:
		var model systemContentBlockMemberCachePointModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.CachePoint = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.SystemContentBlockMemberText:
		var model systemContentBlockMemberTextModel
		d := flex.Flatten(ctx, t.Value, &model.Value)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Text = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	default:
		return diags
	}
}

func (m systemContentBlockModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.CachePoint.IsNull():
		systemContentBlockCachePoint, d := m.CachePoint.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.SystemContentBlockMemberCachePoint
		diags.Append(flex.Expand(ctx, systemContentBlockCachePoint, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	case !m.Text.IsNull():
		systemContentBlockText, d := m.Text.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.SystemContentBlockMemberText
		diags.Append(flex.Expand(ctx, systemContentBlockText.Value, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

type systemContentBlockMemberCachePointModel struct {
	Type fwtypes.StringEnum[awstypes.CachePointType] `tfsdk:"type"`
}

type systemContentBlockMemberTextModel struct {
	Value types.String `tfsdk:"value"`
}

type toolConfigurationModel struct {
	Tools      fwtypes.ListNestedObjectValueOf[toolModel]       `tfsdk:"tool"`
	ToolChoice fwtypes.ListNestedObjectValueOf[toolChoiceModel] `tfsdk:"tool_choice"`
}

// Tagged union
type toolModel struct {
	CachePoint fwtypes.ListNestedObjectValueOf[toolMemberCachePointModel] `tfsdk:"cache_point"`
	ToolSpec   fwtypes.ListNestedObjectValueOf[toolMemberToolSpecModel]   `tfsdk:"tool_spec"`
}

func (m *toolModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.ToolMemberCachePoint:
		var model toolMemberCachePointModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.CachePoint = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.ToolMemberToolSpec:
		var model toolMemberToolSpecModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.ToolSpec = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	default:
		return diags
	}
}

func (m toolModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.CachePoint.IsNull():
		toolCachePoint, d := m.CachePoint.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.ToolMemberCachePoint
		diags.Append(flex.Expand(ctx, toolCachePoint, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	case !m.ToolSpec.IsNull():
		toolToolSpec, d := m.ToolSpec.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.ToolMemberToolSpec
		diags.Append(flex.Expand(ctx, toolToolSpec, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

type toolMemberCachePointModel struct {
	Type fwtypes.StringEnum[awstypes.CachePointType] `tfsdk:"type"`
}

type toolMemberToolSpecModel struct {
	InputSchema fwtypes.ListNestedObjectValueOf[toolInputSchemaModel] `tfsdk:"input_schema"`
	Name        types.String                                          `tfsdk:"name"`
	Description types.String                                          `tfsdk:"description"`
}

// Tagged union
type toolInputSchemaModel struct {
	Json fwtypes.ListNestedObjectValueOf[toolInputSchemaMemberJsonModel] `tfsdk:"json"`
}

func (m *toolInputSchemaModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.ToolInputSchemaMemberJson:
		var model toolInputSchemaMemberJsonModel
		d := flex.Flatten(ctx, t.Value, &model.Value)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Json = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	default:
		return diags
	}
}

func (m toolInputSchemaModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.Json.IsNull():
		toolInputSchemaJson, d := m.Json.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.ToolInputSchemaMemberJson
		diags.Append(flex.Expand(ctx, toolInputSchemaJson.Value, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

type toolInputSchemaMemberJsonModel struct {
	Value fwtypes.SmithyJSON[document.Interface] `tfsdk:"value"`
}

// Tagged union
type toolChoiceModel struct {
	Any  fwtypes.ListNestedObjectValueOf[toolChoiceMemberAnyModel]  `tfsdk:"any"`
	Auto fwtypes.ListNestedObjectValueOf[toolChoiceMemberAutoModel] `tfsdk:"auto"`
	Tool fwtypes.ListNestedObjectValueOf[toolChoiceMemberToolModel] `tfsdk:"tool"`
}

func (m *toolChoiceModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.ToolChoiceMemberAny:
		var model toolChoiceMemberAnyModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Any = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.ToolChoiceMemberAuto:
		var model toolChoiceMemberAutoModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Auto = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.ToolChoiceMemberTool:
		var model toolChoiceMemberToolModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Tool = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	default:
		return diags
	}
}

func (m toolChoiceModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.Any.IsNull():
		toolChoiceAny, d := m.Any.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.ToolChoiceMemberAny
		diags.Append(flex.Expand(ctx, toolChoiceAny, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	case !m.Auto.IsNull():
		toolChoiceAuto, d := m.Auto.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.ToolChoiceMemberAuto
		diags.Append(flex.Expand(ctx, toolChoiceAuto, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	case !m.Tool.IsNull():
		toolChoiceTool, d := m.Tool.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.ToolChoiceMemberTool
		diags.Append(flex.Expand(ctx, toolChoiceTool, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

type toolChoiceMemberAnyModel struct {
}

type toolChoiceMemberAutoModel struct {
}

type toolChoiceMemberToolModel struct {
	Name types.String `tfsdk:"name"`
}

type promptTemplateConfigurationMemberTextModel struct {
	Text           types.String                                              `tfsdk:"text"`
	CachePoint     fwtypes.ListNestedObjectValueOf[cachePointModel]          `tfsdk:"cache_point"`
	InputVariables fwtypes.ListNestedObjectValueOf[promptInputVariableModel] `tfsdk:"input_variable"`
}

type cachePointModel struct {
	Type fwtypes.StringEnum[awstypes.CachePointType] `tfsdk:"type"`
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
