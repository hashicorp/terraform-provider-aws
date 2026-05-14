// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/document"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagentcore_harness", name="Harness")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func newHarnessResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &harnessResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type harnessResource struct {
	framework.ResourceWithModel[harnessResourceModel]
	framework.WithTimeouts
}

func (r *harnessResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"allowed_tools": schema.ListAttribute{
				CustomType: fwtypes.ListOfStringType,
				Optional:   true,
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"environment_variables": schema.MapAttribute{
				CustomType: fwtypes.MapOfStringType,
				Optional:   true,
				Sensitive:  true,
			},
			names.AttrExecutionRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			"harness_id": framework.IDAttribute(),
			"harness_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 40),
				},
			},
			"max_iterations": schema.Int32Attribute{
				Optional: true,
			},
			"max_tokens": schema.Int32Attribute{
				Optional: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"timeout_seconds": schema.Int32Attribute{
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			"authorizer_configuration": authorizerConfigurationSchema(ctx),
			names.AttrEnvironment: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[harnessEnvironmentProviderModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"agentcore_runtime_environment": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[harnessAgentCoreRuntimeEnvironmentModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"agent_runtime_arn": schema.StringAttribute{
										Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"agent_runtime_id": schema.StringAttribute{
										Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"agent_runtime_name": schema.StringAttribute{
										Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									// TODO: Share with agent_runtime
									// TODO: https://github.com/hashicorp/terraform-provider-aws/pull/47810.
									"filesystem_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[harnessFilesystemConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"session_storage": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[sessionStorageConfigModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"mount_path": schema.StringAttribute{
																Required: true,
															},
														},
													},
												},
											},
										},
									},
									// TODO: Share with agent_runtime
									// TODO: https://github.com/hashicorp/terraform-provider-aws/pull/47810.
									"lifecycle_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[lifecycleConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"idle_runtime_session_timeout": schema.Int32Attribute{
													Optional: true,
													Validators: []validator.Int32{
														int32validator.Between(60, 28800),
													},
												},
												"max_lifetime": schema.Int32Attribute{
													Optional: true,
													Validators: []validator.Int32{
														int32validator.Between(60, 28800),
													},
												},
											},
										},
									},
									names.AttrNetworkConfiguration: schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[networkConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"network_mode": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.NetworkMode](),
													Required:   true,
												},
											},
											Blocks: map[string]schema.Block{
												"network_mode_config": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[vpcConfigModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrSecurityGroups: schema.SetAttribute{
																CustomType: fwtypes.SetOfStringType,
																Required:   true,
															},
															names.AttrSubnets: schema.SetAttribute{
																CustomType: fwtypes.SetOfStringType,
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
			"environment_artifact": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[harnessEnvironmentArtifactModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"container_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[containerConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"container_uri": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"memory": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[harnessMemoryConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"agentcore_memory_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[harnessAgentCoreMemoryConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrARN: schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
									"actor_id": schema.StringAttribute{
										Optional: true,
									},
									"messages_count": schema.Int32Attribute{
										Optional: true,
									},
								},
								Blocks: map[string]schema.Block{
									"retrieval_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[harnessAgentCoreMemoryRetrievalConfigModel](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"map_block_key": schema.StringAttribute{
													Required: true,
												},
												"relevance_score": schema.Float32Attribute{
													Optional: true,
													Validators: []validator.Float32{
														float32validator.Between(0, 1),
													},
												},
												"strategy_id": schema.StringAttribute{
													Optional: true,
												},
												"top_k": schema.Int32Attribute{
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
			"model": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[harnessModelConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"bedrock_model_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[harnessBedrockModelConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"max_tokens": schema.Int32Attribute{
										Optional: true,
										Validators: []validator.Int32{
											int32validator.AtLeast(1),
										},
									},
									"model_id": schema.StringAttribute{
										Required: true,
									},
									"temperature": schema.Float32Attribute{
										Optional: true,
										Validators: []validator.Float32{
											float32validator.Between(0, 2),
										},
									},
									"top_p": schema.Float32Attribute{
										Optional: true,
										Validators: []validator.Float32{
											float32validator.Between(0, 1),
										},
									},
								},
							},
						},
						"gemini_model_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[harnessGeminiModelConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"api_key_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
									"max_tokens": schema.Int32Attribute{
										Optional: true,
										Validators: []validator.Int32{
											int32validator.AtLeast(1),
										},
									},
									"model_id": schema.StringAttribute{
										Required: true,
									},
									"temperature": schema.Float32Attribute{
										Optional: true,
										Validators: []validator.Float32{
											float32validator.Between(0, 2),
										},
									},
									"top_p": schema.Float32Attribute{
										Optional: true,
										Validators: []validator.Float32{
											float32validator.Between(0, 1),
										},
									},
									"top_k": schema.Int32Attribute{
										Optional: true,
										Validators: []validator.Int32{
											int32validator.Between(0, 500),
										},
									},
								},
							},
						},
						"openai_model_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[harnessOpenAIModelConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"api_key_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
									"max_tokens": schema.Int32Attribute{
										Optional: true,
										Validators: []validator.Int32{
											int32validator.AtLeast(1),
										},
									},
									"model_id": schema.StringAttribute{
										Required: true,
									},
									"temperature": schema.Float32Attribute{
										Optional: true,
										Validators: []validator.Float32{
											float32validator.Between(0, 2),
										},
									},
									"top_p": schema.Float32Attribute{
										Optional: true,
										Validators: []validator.Float32{
											float32validator.Between(0, 1),
										},
									},
								},
							},
						},
					},
				},
			},
			"skill": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[harnessSkillModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrPath: schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			"system_prompt": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[harnessSystemContentBlockModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"text": schema.StringAttribute{
							Required:  true,
							Sensitive: true,
						},
					},
				},
			},
			"tool": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[harnessToolModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName: schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 64),
							},
						},
						names.AttrType: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.HarnessToolType](),
							Required:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[harnessToolConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"agentcore_browser": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[harnessAgentCoreBrowserConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"browser_arn": schema.StringAttribute{
													CustomType: fwtypes.ARNType,
													Optional:   true,
												},
											},
										},
									},
									"agentcore_code_interpreter": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[harnessAgentCoreCodeInterpreterConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"code_interpreter_arn": schema.StringAttribute{
													CustomType: fwtypes.ARNType,
													Optional:   true,
												},
											},
										},
									},
									"agentcore_gateway": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[harnessAgentCoreGatewayConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"gateway_arn": schema.StringAttribute{
													CustomType: fwtypes.ARNType,
													Required:   true,
												},
											},
											Blocks: map[string]schema.Block{
												"outbound_auth": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[harnessGatewayOutboundAuthModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"aws_iam": schema.BoolAttribute{
																Optional: true,
															},
															"none": schema.BoolAttribute{
																Optional: true,
															},
														},
														Blocks: map[string]schema.Block{
															"oauth": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[harnessOAuthCredentialProviderModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"custom_parameters": schema.MapAttribute{
																			CustomType: fwtypes.MapOfStringType,
																			Optional:   true,
																		},
																		"default_return_url": schema.StringAttribute{
																			Optional: true,
																		},
																		"grant_type": schema.StringAttribute{
																			CustomType: fwtypes.StringEnumType[awstypes.OAuthGrantType](),
																			Optional:   true,
																		},
																		"provider_arn": schema.StringAttribute{
																			CustomType: fwtypes.ARNType,
																			Required:   true,
																		},
																		"scopes": schema.ListAttribute{
																			CustomType: fwtypes.ListOfStringType,
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
									"inline_function": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[harnessInlineFunctionConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrDescription: schema.StringAttribute{
													Required: true,
												},
												"input_schema": schema.StringAttribute{
													Required:  true,
													Sensitive: true,
												},
											},
										},
									},
									"remote_mcp": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[harnessRemoteMCPConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"headers": schema.MapAttribute{
													CustomType: fwtypes.MapOfStringType,
													Optional:   true,
													Sensitive:  true,
												},
												names.AttrURL: schema.StringAttribute{
													Required:  true,
													Sensitive: true,
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
			"truncation": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[harnessTruncationConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"strategy": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.HarnessTruncationStrategy](),
							Required:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[harnessTruncationStrategyConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"sliding_window": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[harnessSlidingWindowConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"messages_count": schema.Int32Attribute{
													Optional: true,
												},
											},
										},
									},
									"summarization": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[harnessSummarizationConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"summary_ratio": schema.Float32Attribute{
													Optional: true,
												},
												"preserve_recent_messages": schema.Int32Attribute{
													Optional: true,
												},
												"summarization_system_prompt": schema.StringAttribute{
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *harnessResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data harnessResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	input := expandHarnessCreateInput(ctx, &data, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	var (
		out *bedrockagentcorecontrol.CreateHarnessOutput
		err error
	)
	err = tfresource.Retry(ctx, propagationTimeout, func(ctx context.Context) *tfresource.RetryError {
		out, err = conn.CreateHarness(ctx, input)

		if tfawserr.ErrMessageContains(err, errCodeValidationException, "Role validation failed") {
			return tfresource.RetryableError(err)
		}
		if tfawserr.ErrMessageContains(err, errCodeValidationException, "Access denied") {
			return tfresource.RetryableError(err)
		}

		if err != nil {
			return tfresource.NonRetryableError(err)
		}

		return nil
	})
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.HarnessName.String())
		return
	}

	harnessID := aws.ToString(out.Harness.HarnessId)

	if _, err := waitHarnessCreated(ctx, conn, harnessID, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, harnessID)
		return
	}

	harness, err := findHarnessByID(ctx, conn, harnessID)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, harnessID)
		return
	}

	flattenHarnessToModel(ctx, harness, &data, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *harnessResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data harnessResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	harnessID := fwflex.StringValueFromFramework(ctx, data.HarnessID)
	harness, err := findHarnessByID(ctx, conn, harnessID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, harnessID)
		return
	}

	flattenHarnessToModel(ctx, harness, &data, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *harnessResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old harnessResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	harnessID := fwflex.StringValueFromFramework(ctx, new.HarnessID)

	if diff.HasChanges() {
		input := expandHarnessUpdateInput(ctx, &new, &old, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}

		input.HarnessId = aws.String(harnessID)
		input.ClientToken = aws.String(create.UniqueId(ctx))

		_, err := conn.UpdateHarness(ctx, input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, harnessID)
			return
		}

		if _, err := waitHarnessUpdated(ctx, conn, harnessID, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, harnessID)
			return
		}
	}

	harness, err := findHarnessByID(ctx, conn, harnessID)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, harnessID)
		return
	}

	flattenHarnessToModel(ctx, harness, &new, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *harnessResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data harnessResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	harnessID := fwflex.StringValueFromFramework(ctx, data.HarnessID)
	input := bedrockagentcorecontrol.DeleteHarnessInput{
		HarnessId: aws.String(harnessID),
	}

	_, err := conn.DeleteHarness(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, harnessID)
		return
	}

	if _, err := waitHarnessDeleted(ctx, conn, harnessID, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, harnessID)
		return
	}
}

func (r *harnessResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("harness_id"), request, response)
}

// Waiters.

func waitHarnessCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*awstypes.Harness, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.HarnessStatusCreating),
		Target:                    enum.Slice(awstypes.HarnessStatusReady),
		Refresh:                   statusHarness(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Harness); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitHarnessUpdated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*awstypes.Harness, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.HarnessStatusUpdating),
		Target:                    enum.Slice(awstypes.HarnessStatusReady),
		Refresh:                   statusHarness(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Harness); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitHarnessDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*awstypes.Harness, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.HarnessStatusDeleting, awstypes.HarnessStatusReady),
		Target:  []string{},
		Refresh: statusHarness(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Harness); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusHarness(conn *bedrockagentcorecontrol.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findHarnessByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

// Finders.

func findHarnessByID(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string) (*awstypes.Harness, error) {
	input := bedrockagentcorecontrol.GetHarnessInput{
		HarnessId: aws.String(id),
	}

	return findHarness(ctx, conn, &input)
}

func findHarness(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.GetHarnessInput) (*awstypes.Harness, error) {
	out, err := conn.GetHarness(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil || out.Harness == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.Harness, nil
}

// Expand helpers.

func expandHarnessCreateInput(ctx context.Context, data *harnessResourceModel, diags *diag.Diagnostics) *bedrockagentcorecontrol.CreateHarnessInput {
	var input bedrockagentcorecontrol.CreateHarnessInput

	input.HarnessName = aws.String(data.HarnessName.ValueString())
	input.ExecutionRoleArn = aws.String(data.ExecutionRoleARN.ValueString())

	if !data.AllowedTools.IsNull() {
		smerr.AddEnrich(ctx, diags, data.AllowedTools.ElementsAs(ctx, &input.AllowedTools, false))
	}

	if !data.EnvironmentVariables.IsNull() {
		input.EnvironmentVariables = make(map[string]string)
		smerr.AddEnrich(ctx, diags, data.EnvironmentVariables.ElementsAs(ctx, &input.EnvironmentVariables, false))
	}

	if !data.MaxIterations.IsNull() {
		input.MaxIterations = aws.Int32(data.MaxIterations.ValueInt32())
	}

	if !data.MaxTokens.IsNull() {
		input.MaxTokens = aws.Int32(data.MaxTokens.ValueInt32())
	}

	if !data.TimeoutSeconds.IsNull() {
		input.TimeoutSeconds = aws.Int32(data.TimeoutSeconds.ValueInt32())
	}

	if !data.Model.IsNull() {
		models, d := data.Model.ToSlice(ctx)
		smerr.AddEnrich(ctx, diags, d)
		if len(models) > 0 {
			expanded, d := models[0].Expand(ctx)
			smerr.AddEnrich(ctx, diags, d)
			if expanded != nil {
				input.Model = expanded.(awstypes.HarnessModelConfiguration)
			}
		}
	}

	if !data.SystemPrompt.IsNull() {
		blocks, d := data.SystemPrompt.ToSlice(ctx)
		smerr.AddEnrich(ctx, diags, d)
		for _, b := range blocks {
			expanded, d := b.Expand(ctx)
			smerr.AddEnrich(ctx, diags, d)
			if expanded != nil {
				input.SystemPrompt = append(input.SystemPrompt, expanded.(awstypes.HarnessSystemContentBlock))
			}
		}
	}

	if !data.Tools.IsNull() {
		tools, d := data.Tools.ToSlice(ctx)
		smerr.AddEnrich(ctx, diags, d)
		for _, t := range tools {
			expanded, d := t.Expand(ctx)
			smerr.AddEnrich(ctx, diags, d)
			input.Tools = append(input.Tools, expanded)
		}
	}

	if !data.Skills.IsNull() {
		skills, d := data.Skills.ToSlice(ctx)
		smerr.AddEnrich(ctx, diags, d)
		for _, s := range skills {
			expanded, d := s.Expand(ctx)
			smerr.AddEnrich(ctx, diags, d)
			if expanded != nil {
				input.Skills = append(input.Skills, expanded.(awstypes.HarnessSkill))
			}
		}
	}

	if !data.Truncation.IsNull() {
		trunc, d := data.Truncation.ToSlice(ctx)
		smerr.AddEnrich(ctx, diags, d)
		if len(trunc) > 0 {
			expanded, d := trunc[0].Expand(ctx)
			smerr.AddEnrich(ctx, diags, d)
			input.Truncation = expanded
		}
	}

	if !data.Environment.IsNull() {
		envs, d := data.Environment.ToSlice(ctx)
		smerr.AddEnrich(ctx, diags, d)
		if len(envs) > 0 {
			expanded, d := envs[0].ExpandRequest(ctx)
			smerr.AddEnrich(ctx, diags, d)
			input.Environment = expanded
		}
	}

	if !data.EnvironmentArtifact.IsNull() {
		artifacts, d := data.EnvironmentArtifact.ToSlice(ctx)
		smerr.AddEnrich(ctx, diags, d)
		if len(artifacts) > 0 {
			expanded, d := artifacts[0].Expand(ctx)
			smerr.AddEnrich(ctx, diags, d)
			if expanded != nil {
				input.EnvironmentArtifact = expanded.(awstypes.HarnessEnvironmentArtifact)
			}
		}
	}

	if !data.AuthorizerConfiguration.IsNull() {
		authConfigs, d := data.AuthorizerConfiguration.ToSlice(ctx)
		smerr.AddEnrich(ctx, diags, d)
		if len(authConfigs) > 0 {
			expanded, d := authConfigs[0].Expand(ctx)
			smerr.AddEnrich(ctx, diags, d)
			if expanded != nil {
				input.AuthorizerConfiguration = expanded.(awstypes.AuthorizerConfiguration)
			}
		}
	}

	if !data.Memory.IsNull() {
		mems, d := data.Memory.ToSlice(ctx)
		smerr.AddEnrich(ctx, diags, d)
		if len(mems) > 0 {
			expanded, d := mems[0].Expand(ctx)
			smerr.AddEnrich(ctx, diags, d)
			if expanded != nil {
				input.Memory = expanded.(awstypes.HarnessMemoryConfiguration)
			}
		}
	}

	return &input
}

func expandHarnessUpdateInput(ctx context.Context, new, old *harnessResourceModel, diags *diag.Diagnostics) *bedrockagentcorecontrol.UpdateHarnessInput {
	var input bedrockagentcorecontrol.UpdateHarnessInput

	input.ExecutionRoleArn = aws.String(new.ExecutionRoleARN.ValueString())

	if !new.AllowedTools.IsNull() {
		smerr.AddEnrich(ctx, diags, new.AllowedTools.ElementsAs(ctx, &input.AllowedTools, false))
	}

	if !new.EnvironmentVariables.IsNull() {
		input.EnvironmentVariables = make(map[string]string)
		smerr.AddEnrich(ctx, diags, new.EnvironmentVariables.ElementsAs(ctx, &input.EnvironmentVariables, false))
	}

	if !new.MaxIterations.IsNull() {
		input.MaxIterations = aws.Int32(new.MaxIterations.ValueInt32())
	}

	if !new.MaxTokens.IsNull() {
		input.MaxTokens = aws.Int32(new.MaxTokens.ValueInt32())
	}

	if !new.TimeoutSeconds.IsNull() {
		input.TimeoutSeconds = aws.Int32(new.TimeoutSeconds.ValueInt32())
	}

	if !new.Model.IsNull() {
		models, d := new.Model.ToSlice(ctx)
		smerr.AddEnrich(ctx, diags, d)
		if len(models) > 0 {
			expanded, d := models[0].Expand(ctx)
			smerr.AddEnrich(ctx, diags, d)
			if expanded != nil {
				input.Model = expanded.(awstypes.HarnessModelConfiguration)
			}
		}
	}

	if !new.SystemPrompt.IsNull() {
		blocks, d := new.SystemPrompt.ToSlice(ctx)
		smerr.AddEnrich(ctx, diags, d)
		for _, b := range blocks {
			expanded, d := b.Expand(ctx)
			smerr.AddEnrich(ctx, diags, d)
			if expanded != nil {
				input.SystemPrompt = append(input.SystemPrompt, expanded.(awstypes.HarnessSystemContentBlock))
			}
		}
	}

	if !new.Tools.IsNull() {
		tools, d := new.Tools.ToSlice(ctx)
		smerr.AddEnrich(ctx, diags, d)
		for _, t := range tools {
			expanded, d := t.Expand(ctx)
			smerr.AddEnrich(ctx, diags, d)
			input.Tools = append(input.Tools, expanded)
		}
	}

	if !new.Skills.IsNull() {
		skills, d := new.Skills.ToSlice(ctx)
		smerr.AddEnrich(ctx, diags, d)
		for _, s := range skills {
			expanded, d := s.Expand(ctx)
			smerr.AddEnrich(ctx, diags, d)
			if expanded != nil {
				input.Skills = append(input.Skills, expanded.(awstypes.HarnessSkill))
			}
		}
	}

	if !new.Truncation.IsNull() {
		trunc, d := new.Truncation.ToSlice(ctx)
		smerr.AddEnrich(ctx, diags, d)
		if len(trunc) > 0 {
			expanded, d := trunc[0].Expand(ctx)
			smerr.AddEnrich(ctx, diags, d)
			input.Truncation = expanded
		}
	}

	if !new.Environment.IsNull() {
		envs, d := new.Environment.ToSlice(ctx)
		smerr.AddEnrich(ctx, diags, d)
		if len(envs) > 0 {
			expanded, d := envs[0].ExpandRequest(ctx)
			smerr.AddEnrich(ctx, diags, d)
			input.Environment = expanded
		}
	}

	// Wrapper types for clearable optional fields.
	if !new.EnvironmentArtifact.Equal(old.EnvironmentArtifact) {
		if new.EnvironmentArtifact.IsNull() {
			input.EnvironmentArtifact = &awstypes.UpdatedHarnessEnvironmentArtifact{}
		} else {
			artifacts, d := new.EnvironmentArtifact.ToSlice(ctx)
			smerr.AddEnrich(ctx, diags, d)
			if len(artifacts) > 0 {
				expanded, d := artifacts[0].Expand(ctx)
				smerr.AddEnrich(ctx, diags, d)
				if expanded != nil {
					input.EnvironmentArtifact = &awstypes.UpdatedHarnessEnvironmentArtifact{
						OptionalValue: expanded.(awstypes.HarnessEnvironmentArtifact),
					}
				}
			}
		}
	}

	if !new.AuthorizerConfiguration.Equal(old.AuthorizerConfiguration) {
		if new.AuthorizerConfiguration.IsNull() {
			input.AuthorizerConfiguration = &awstypes.UpdatedAuthorizerConfiguration{}
		} else {
			authConfigs, d := new.AuthorizerConfiguration.ToSlice(ctx)
			smerr.AddEnrich(ctx, diags, d)
			if len(authConfigs) > 0 {
				expanded, d := authConfigs[0].Expand(ctx)
				smerr.AddEnrich(ctx, diags, d)
				if expanded != nil {
					input.AuthorizerConfiguration = &awstypes.UpdatedAuthorizerConfiguration{
						OptionalValue: expanded.(awstypes.AuthorizerConfiguration),
					}
				}
			}
		}
	}

	if !new.Memory.Equal(old.Memory) {
		if new.Memory.IsNull() {
			input.Memory = &awstypes.UpdatedHarnessMemoryConfiguration{}
		} else {
			mems, d := new.Memory.ToSlice(ctx)
			smerr.AddEnrich(ctx, diags, d)
			if len(mems) > 0 {
				expanded, d := mems[0].Expand(ctx)
				smerr.AddEnrich(ctx, diags, d)
				if expanded != nil {
					input.Memory = &awstypes.UpdatedHarnessMemoryConfiguration{
						OptionalValue: expanded.(awstypes.HarnessMemoryConfiguration),
					}
				}
			}
		}
	}

	return &input
}

// Flatten helpers.

func flattenHarnessToModel(ctx context.Context, harness *awstypes.Harness, data *harnessResourceModel, diags *diag.Diagnostics) {
	data.ARN = fwflex.StringToFramework(ctx, harness.Arn)
	data.HarnessID = fwflex.StringToFramework(ctx, harness.HarnessId)
	data.HarnessName = fwflex.StringToFramework(ctx, harness.HarnessName)
	data.ExecutionRoleARN = fwtypes.ARNValue(aws.ToString(harness.ExecutionRoleArn))
	data.Status = fwflex.StringValueToFramework(ctx, string(harness.Status))
	data.FailureReason = fwflex.StringToFramework(ctx, harness.FailureReason)

	if harness.CreatedAt != nil {
		data.CreatedAt = timetypes.NewRFC3339ValueMust(harness.CreatedAt.Format(time.RFC3339))
	}
	if harness.UpdatedAt != nil {
		data.UpdatedAt = timetypes.NewRFC3339ValueMust(harness.UpdatedAt.Format(time.RFC3339))
	}

	// Only flatten server-defaulted optional fields when the user configured them.
	// The server returns defaults for these even when unset by the user; overwriting
	// the model would cause perpetual drift on subsequent plans.
	if !data.MaxIterations.IsNull() {
		if harness.MaxIterations != nil {
			data.MaxIterations = types.Int32Value(*harness.MaxIterations)
		} else {
			data.MaxIterations = types.Int32Null()
		}
	}

	if !data.MaxTokens.IsNull() {
		if harness.MaxTokens != nil {
			data.MaxTokens = types.Int32Value(*harness.MaxTokens)
		} else {
			data.MaxTokens = types.Int32Null()
		}
	}

	if !data.TimeoutSeconds.IsNull() {
		if harness.TimeoutSeconds != nil {
			data.TimeoutSeconds = types.Int32Value(*harness.TimeoutSeconds)
		} else {
			data.TimeoutSeconds = types.Int32Null()
		}
	}

	if !data.AllowedTools.IsNull() && len(harness.AllowedTools) > 0 {
		data.AllowedTools = fwflex.FlattenFrameworkStringValueListOfString(ctx, harness.AllowedTools)
	}

	if !data.EnvironmentVariables.IsNull() {
		if len(harness.EnvironmentVariables) > 0 {
			data.EnvironmentVariables = fwflex.FlattenFrameworkStringValueMapOfString(ctx, harness.EnvironmentVariables)
		} else {
			data.EnvironmentVariables = fwtypes.NewMapValueOfNull[basetypes.StringValue](ctx)
		}
	}

	// Model.
	var modelConfig harnessModelConfigurationModel
	smerr.AddEnrich(ctx, diags, modelConfig.Flatten(ctx, harness.Model))
	data.Model = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &modelConfig)

	// System prompt.
	if len(harness.SystemPrompt) > 0 {
		var blocks []*harnessSystemContentBlockModel
		for _, b := range harness.SystemPrompt {
			var block harnessSystemContentBlockModel
			smerr.AddEnrich(ctx, diags, block.Flatten(ctx, b))
			blocks = append(blocks, &block)
		}
		data.SystemPrompt = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, blocks)
	} else {
		data.SystemPrompt = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*harnessSystemContentBlockModel{})
	}

	// Tools.
	if len(harness.Tools) > 0 {
		var tools []*harnessToolModel
		for _, t := range harness.Tools {
			var tool harnessToolModel
			smerr.AddEnrich(ctx, diags, tool.Flatten(ctx, t))
			tools = append(tools, &tool)
		}
		data.Tools = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, tools)
	} else {
		data.Tools = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*harnessToolModel{})
	}

	// Skills.
	if len(harness.Skills) > 0 {
		var skills []*harnessSkillModel
		for _, s := range harness.Skills {
			var skill harnessSkillModel
			smerr.AddEnrich(ctx, diags, skill.Flatten(ctx, s))
			skills = append(skills, &skill)
		}
		data.Skills = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, skills)
	} else {
		data.Skills = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*harnessSkillModel{})
	}

	// Truncation — server returns defaults even when unset; preserve null.
	if !data.Truncation.IsNull() {
		if harness.Truncation != nil {
			var trunc harnessTruncationConfigurationModel
			smerr.AddEnrich(ctx, diags, trunc.Flatten(ctx, harness.Truncation))
			data.Truncation = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &trunc)
		} else {
			data.Truncation = fwtypes.NewListNestedObjectValueOfNull[harnessTruncationConfigurationModel](ctx)
		}
	}

	// Environment — server returns defaults even when unset; preserve null.
	if !data.Environment.IsNull() {
		if harness.Environment != nil {
			var env harnessEnvironmentProviderModel
			smerr.AddEnrich(ctx, diags, env.Flatten(ctx, harness.Environment))
			data.Environment = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &env)
		} else {
			data.Environment = fwtypes.NewListNestedObjectValueOfNull[harnessEnvironmentProviderModel](ctx)
		}
	}

	// Environment artifact.
	if harness.EnvironmentArtifact != nil {
		var artifact harnessEnvironmentArtifactModel
		smerr.AddEnrich(ctx, diags, artifact.Flatten(ctx, harness.EnvironmentArtifact))
		data.EnvironmentArtifact = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &artifact)
	} else {
		data.EnvironmentArtifact = fwtypes.NewListNestedObjectValueOfNull[harnessEnvironmentArtifactModel](ctx)
	}

	// Authorizer configuration.
	if harness.AuthorizerConfiguration != nil {
		var authConfig authorizerConfigurationModel
		smerr.AddEnrich(ctx, diags, authConfig.Flatten(ctx, harness.AuthorizerConfiguration))
		data.AuthorizerConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &authConfig)
	} else {
		data.AuthorizerConfiguration = fwtypes.NewListNestedObjectValueOfNull[authorizerConfigurationModel](ctx)
	}

	// Memory.
	if harness.Memory != nil {
		var memConfig harnessMemoryConfigurationModel
		smerr.AddEnrich(ctx, diags, memConfig.Flatten(ctx, harness.Memory))
		data.Memory = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &memConfig)
	} else {
		data.Memory = fwtypes.NewListNestedObjectValueOfNull[harnessMemoryConfigurationModel](ctx)
	}
}

// Model structs.

type harnessResourceModel struct {
	framework.WithRegionModel
	AllowedTools            fwtypes.ListOfString                                                 `tfsdk:"allowed_tools"`
	ARN                     types.String                                                         `tfsdk:"arn"`
	AuthorizerConfiguration fwtypes.ListNestedObjectValueOf[authorizerConfigurationModel]        `tfsdk:"authorizer_configuration"`
	Environment             fwtypes.ListNestedObjectValueOf[harnessEnvironmentProviderModel]     `tfsdk:"environment"`
	EnvironmentArtifact     fwtypes.ListNestedObjectValueOf[harnessEnvironmentArtifactModel]     `tfsdk:"environment_artifact"`
	EnvironmentVariables    fwtypes.MapOfString                                                  `tfsdk:"environment_variables"`
	ExecutionRoleARN        fwtypes.ARN                                                          `tfsdk:"execution_role_arn"`
	HarnessID               types.String                                                         `tfsdk:"harness_id"`
	HarnessName             types.String                                                         `tfsdk:"harness_name"`
	MaxIterations           types.Int32                                                          `tfsdk:"max_iterations"`
	MaxTokens               types.Int32                                                          `tfsdk:"max_tokens"`
	Memory                  fwtypes.ListNestedObjectValueOf[harnessMemoryConfigurationModel]     `tfsdk:"memory"`
	Model                   fwtypes.ListNestedObjectValueOf[harnessModelConfigurationModel]      `tfsdk:"model"`
	Skills                  fwtypes.ListNestedObjectValueOf[harnessSkillModel]                   `tfsdk:"skill"`
	SystemPrompt            fwtypes.ListNestedObjectValueOf[harnessSystemContentBlockModel]      `tfsdk:"system_prompt"`
	Tags                    tftags.Map                                                           `tfsdk:"tags"`
	TagsAll                 tftags.Map                                                           `tfsdk:"tags_all"`
	Timeouts                timeouts.Value                                                       `tfsdk:"timeouts"`
	TimeoutSeconds          types.Int32                                                          `tfsdk:"timeout_seconds"`
	Tools                   fwtypes.ListNestedObjectValueOf[harnessToolModel]                    `tfsdk:"tool"`
	Truncation              fwtypes.ListNestedObjectValueOf[harnessTruncationConfigurationModel] `tfsdk:"truncation"`
}

// Model configuration union.

type harnessModelConfigurationModel struct {
	BedrockModelConfig fwtypes.ListNestedObjectValueOf[harnessBedrockModelConfigModel] `tfsdk:"bedrock_model_config"`
	GeminiModelConfig  fwtypes.ListNestedObjectValueOf[harnessGeminiModelConfigModel]  `tfsdk:"gemini_model_config"`
	OpenAiModelConfig  fwtypes.ListNestedObjectValueOf[harnessOpenAIModelConfigModel]  `tfsdk:"openai_model_config"`
}

var (
	_ fwflex.Expander  = harnessModelConfigurationModel{}
	_ fwflex.Flattener = &harnessModelConfigurationModel{}
)

func (m *harnessModelConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	m.BedrockModelConfig = fwtypes.NewListNestedObjectValueOfNull[harnessBedrockModelConfigModel](ctx)
	m.OpenAiModelConfig = fwtypes.NewListNestedObjectValueOfNull[harnessOpenAIModelConfigModel](ctx)
	m.GeminiModelConfig = fwtypes.NewListNestedObjectValueOfNull[harnessGeminiModelConfigModel](ctx)

	switch t := v.(type) {
	case *awstypes.HarnessModelConfigurationMemberBedrockModelConfig:
		var data harnessBedrockModelConfigModel
		data.ModelID = fwflex.StringToFramework(ctx, t.Value.ModelId)
		if t.Value.MaxTokens != nil {
			data.MaxTokens = types.Int32Value(*t.Value.MaxTokens)
		}
		if t.Value.Temperature != nil {
			data.Temperature = types.Float32Value(*t.Value.Temperature)
		}
		if t.Value.TopP != nil {
			data.TopP = types.Float32Value(*t.Value.TopP)
		}
		m.BedrockModelConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	case *awstypes.HarnessModelConfigurationMemberOpenAiModelConfig:
		var data harnessOpenAIModelConfigModel
		data.ModelID = fwflex.StringToFramework(ctx, t.Value.ModelId)
		data.ApiKeyARN = fwflex.StringToFramework(ctx, t.Value.ApiKeyArn)
		if t.Value.MaxTokens != nil {
			data.MaxTokens = types.Int32Value(*t.Value.MaxTokens)
		}
		if t.Value.Temperature != nil {
			data.Temperature = types.Float32Value(*t.Value.Temperature)
		}
		if t.Value.TopP != nil {
			data.TopP = types.Float32Value(*t.Value.TopP)
		}
		m.OpenAiModelConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	case *awstypes.HarnessModelConfigurationMemberGeminiModelConfig:
		var data harnessGeminiModelConfigModel
		data.ModelID = fwflex.StringToFramework(ctx, t.Value.ModelId)
		data.ApiKeyARN = fwflex.StringToFramework(ctx, t.Value.ApiKeyArn)
		if t.Value.MaxTokens != nil {
			data.MaxTokens = types.Int32Value(*t.Value.MaxTokens)
		}
		if t.Value.Temperature != nil {
			data.Temperature = types.Float32Value(*t.Value.Temperature)
		}
		if t.Value.TopP != nil {
			data.TopP = types.Float32Value(*t.Value.TopP)
		}
		if t.Value.TopK != nil {
			data.TopK = types.Int32Value(*t.Value.TopK)
		}
		m.GeminiModelConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("model configuration flatten: %T", v))
	}
	return diags
}

func (m harnessModelConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.BedrockModelConfig.IsNull():
		data, d := m.BedrockModelConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.HarnessModelConfigurationMemberBedrockModelConfig
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		return &r, diags
	case !m.OpenAiModelConfig.IsNull():
		data, d := m.OpenAiModelConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.HarnessModelConfigurationMemberOpenAiModelConfig
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		return &r, diags
	case !m.GeminiModelConfig.IsNull():
		data, d := m.GeminiModelConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.HarnessModelConfigurationMemberGeminiModelConfig
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		return &r, diags
	}
	return nil, diags
}

type harnessBedrockModelConfigModel struct {
	MaxTokens   types.Int32   `tfsdk:"max_tokens"`
	ModelID     types.String  `tfsdk:"model_id"`
	Temperature types.Float32 `tfsdk:"temperature"`
	TopP        types.Float32 `tfsdk:"top_p"`
}

type harnessGeminiModelConfigModel struct {
	ApiKeyARN   fwtypes.ARN   `tfsdk:"api_key_arn"`
	MaxTokens   types.Int32   `tfsdk:"max_tokens"`
	ModelID     types.String  `tfsdk:"model_id"`
	Temperature types.Float32 `tfsdk:"temperature"`
	TopP        types.Float32 `tfsdk:"top_p"`
	TopK        types.Int32   `tfsdk:"top_k"`
}

type harnessOpenAIModelConfigModel struct {
	ApiKeyARN   fwtypes.ARN   `tfsdk:"api_key_arn"`
	MaxTokens   types.Int32   `tfsdk:"max_tokens"`
	ModelID     types.String  `tfsdk:"model_id"`
	Temperature types.Float32 `tfsdk:"temperature"`
	TopP        types.Float32 `tfsdk:"top_p"`
}

// System prompt union.

type harnessSystemContentBlockModel struct {
	Text types.String `tfsdk:"text"`
}

var (
	_ fwflex.Expander  = harnessSystemContentBlockModel{}
	_ fwflex.Flattener = &harnessSystemContentBlockModel{}
)

func (m *harnessSystemContentBlockModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case *awstypes.HarnessSystemContentBlockMemberText:
		m.Text = types.StringValue(t.Value)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("system content block flatten: %T", v))
	}
	return diags
}

func (m harnessSystemContentBlockModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !m.Text.IsNull() {
		return &awstypes.HarnessSystemContentBlockMemberText{Value: m.Text.ValueString()}, diags
	}
	return nil, diags
}

// Skill union.

type harnessSkillModel struct {
	Path types.String `tfsdk:"path"`
}

var (
	_ fwflex.Expander  = harnessSkillModel{}
	_ fwflex.Flattener = &harnessSkillModel{}
)

func (m *harnessSkillModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case *awstypes.HarnessSkillMemberPath:
		m.Path = types.StringValue(t.Value)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("skill flatten: %T", v))
	}
	return diags
}

func (m harnessSkillModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !m.Path.IsNull() {
		return &awstypes.HarnessSkillMemberPath{Value: m.Path.ValueString()}, diags
	}
	return nil, diags
}

// Tool model.

type harnessToolModel struct {
	Config fwtypes.ListNestedObjectValueOf[harnessToolConfigurationModel] `tfsdk:"config"`
	Name   types.String                                                   `tfsdk:"name"`
	Type   fwtypes.StringEnum[awstypes.HarnessToolType]                   `tfsdk:"type"`
}

func (m *harnessToolModel) Flatten(ctx context.Context, v awstypes.HarnessTool) diag.Diagnostics {
	var diags diag.Diagnostics
	m.Type = fwtypes.StringEnumValue(v.Type)
	m.Name = fwflex.StringToFramework(ctx, v.Name)
	if v.Config != nil {
		var cfg harnessToolConfigurationModel
		smerr.AddEnrich(ctx, &diags, cfg.Flatten(ctx, v.Config))
		m.Config = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &cfg)
	} else {
		m.Config = fwtypes.NewListNestedObjectValueOfNull[harnessToolConfigurationModel](ctx)
	}
	return diags
}

func (m harnessToolModel) Expand(ctx context.Context) (awstypes.HarnessTool, diag.Diagnostics) {
	var diags diag.Diagnostics
	tool := awstypes.HarnessTool{
		Type: m.Type.ValueEnum(),
		Name: m.Name.ValueStringPointer(),
	}
	if !m.Config.IsNull() {
		configs, d := m.Config.ToSlice(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if len(configs) > 0 {
			expanded, d := configs[0].Expand(ctx)
			smerr.AddEnrich(ctx, &diags, d)
			if expanded != nil {
				tool.Config = expanded.(awstypes.HarnessToolConfiguration)
			}
		}
	}
	return tool, diags
}

// Tool configuration union.

type harnessToolConfigurationModel struct {
	AgentCoreBrowser         fwtypes.ListNestedObjectValueOf[harnessAgentCoreBrowserConfigModel]         `tfsdk:"agentcore_browser"`
	AgentCoreCodeInterpreter fwtypes.ListNestedObjectValueOf[harnessAgentCoreCodeInterpreterConfigModel] `tfsdk:"agentcore_code_interpreter"`
	AgentCoreGateway         fwtypes.ListNestedObjectValueOf[harnessAgentCoreGatewayConfigModel]         `tfsdk:"agentcore_gateway"`
	InlineFunction           fwtypes.ListNestedObjectValueOf[harnessInlineFunctionConfigModel]           `tfsdk:"inline_function"`
	RemoteMcp                fwtypes.ListNestedObjectValueOf[harnessRemoteMCPConfigModel]                `tfsdk:"remote_mcp"`
}

var (
	_ fwflex.Expander  = harnessToolConfigurationModel{}
	_ fwflex.Flattener = &harnessToolConfigurationModel{}
)

func (m *harnessToolConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	m.AgentCoreBrowser = fwtypes.NewListNestedObjectValueOfNull[harnessAgentCoreBrowserConfigModel](ctx)
	m.AgentCoreCodeInterpreter = fwtypes.NewListNestedObjectValueOfNull[harnessAgentCoreCodeInterpreterConfigModel](ctx)
	m.AgentCoreGateway = fwtypes.NewListNestedObjectValueOfNull[harnessAgentCoreGatewayConfigModel](ctx)
	m.InlineFunction = fwtypes.NewListNestedObjectValueOfNull[harnessInlineFunctionConfigModel](ctx)
	m.RemoteMcp = fwtypes.NewListNestedObjectValueOfNull[harnessRemoteMCPConfigModel](ctx)

	switch t := v.(type) {
	case *awstypes.HarnessToolConfigurationMemberRemoteMcp:
		var data harnessRemoteMCPConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		m.RemoteMcp = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	case *awstypes.HarnessToolConfigurationMemberAgentCoreBrowser:
		var data harnessAgentCoreBrowserConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		m.AgentCoreBrowser = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	case *awstypes.HarnessToolConfigurationMemberAgentCoreGateway:
		var data harnessAgentCoreGatewayConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		m.AgentCoreGateway = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	case *awstypes.HarnessToolConfigurationMemberInlineFunction:
		data := harnessInlineFunctionConfigModel{
			Description: fwflex.StringToFramework(ctx, t.Value.Description),
		}
		if t.Value.InputSchema != nil {
			var schemaMap any
			if err := t.Value.InputSchema.UnmarshalSmithyDocument(&schemaMap); err == nil {
				if b, err := json.Marshal(schemaMap); err == nil {
					data.InputSchema = types.StringValue(string(b))
				}
			}
		}
		m.InlineFunction = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	case *awstypes.HarnessToolConfigurationMemberAgentCoreCodeInterpreter:
		var data harnessAgentCoreCodeInterpreterConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		m.AgentCoreCodeInterpreter = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("tool configuration flatten: %T", v))
	}
	return diags
}

func (m harnessToolConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.RemoteMcp.IsNull():
		data, d := m.RemoteMcp.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.HarnessToolConfigurationMemberRemoteMcp
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		return &r, diags
	case !m.AgentCoreBrowser.IsNull():
		data, d := m.AgentCoreBrowser.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.HarnessToolConfigurationMemberAgentCoreBrowser
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		return &r, diags
	case !m.AgentCoreGateway.IsNull():
		data, d := m.AgentCoreGateway.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.HarnessToolConfigurationMemberAgentCoreGateway
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		return &r, diags
	case !m.InlineFunction.IsNull():
		data, d := m.InlineFunction.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.HarnessToolConfigurationMemberInlineFunction
		r.Value = awstypes.HarnessInlineFunctionConfig{
			Description: aws.String(data.Description.ValueString()),
		}
		if !data.InputSchema.IsNull() {
			var schemaMap map[string]any
			if err := json.Unmarshal([]byte(data.InputSchema.ValueString()), &schemaMap); err == nil {
				r.Value.InputSchema = document.NewLazyDocument(schemaMap)
			}
		}
		return &r, diags
	case !m.AgentCoreCodeInterpreter.IsNull():
		data, d := m.AgentCoreCodeInterpreter.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.HarnessToolConfigurationMemberAgentCoreCodeInterpreter
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		return &r, diags
	}
	return nil, diags
}

type harnessAgentCoreBrowserConfigModel struct {
	BrowserARN fwtypes.ARN `tfsdk:"browser_arn"`
}

type harnessAgentCoreGatewayConfigModel struct {
	GatewayARN   fwtypes.ARN                                                      `tfsdk:"gateway_arn"`
	OutboundAuth fwtypes.ListNestedObjectValueOf[harnessGatewayOutboundAuthModel] `tfsdk:"outbound_auth"`
}

type harnessAgentCoreCodeInterpreterConfigModel struct {
	CodeInterpreterARN fwtypes.ARN `tfsdk:"code_interpreter_arn"`
}

type harnessInlineFunctionConfigModel struct {
	Description types.String `tfsdk:"description"`
	InputSchema types.String `tfsdk:"input_schema"`
}

type harnessRemoteMCPConfigModel struct {
	Headers fwtypes.MapOfString `tfsdk:"headers"`
	URL     types.String        `tfsdk:"url"`
}

// Gateway outbound auth union.

type harnessGatewayOutboundAuthModel struct {
	AwsIam types.Bool                                                           `tfsdk:"aws_iam"`
	None   types.Bool                                                           `tfsdk:"none"`
	OAuth  fwtypes.ListNestedObjectValueOf[harnessOAuthCredentialProviderModel] `tfsdk:"oauth"`
}

var (
	_ fwflex.Expander  = harnessGatewayOutboundAuthModel{}
	_ fwflex.Flattener = &harnessGatewayOutboundAuthModel{}
)

func (m *harnessGatewayOutboundAuthModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	m.OAuth = fwtypes.NewListNestedObjectValueOfNull[harnessOAuthCredentialProviderModel](ctx)

	switch t := v.(type) {
	case *awstypes.HarnessGatewayOutboundAuthMemberAwsIam:
		_ = t
		m.AwsIam = types.BoolValue(true)
	case *awstypes.HarnessGatewayOutboundAuthMemberNone:
		_ = t
		m.None = types.BoolValue(true)
	case *awstypes.HarnessGatewayOutboundAuthMemberOauth:
		var data harnessOAuthCredentialProviderModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		m.OAuth = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("gateway outbound auth flatten: %T", v))
	}
	return diags
}

func (m harnessGatewayOutboundAuthModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.AwsIam.IsNull() && m.AwsIam.ValueBool():
		return &awstypes.HarnessGatewayOutboundAuthMemberAwsIam{}, diags
	case !m.None.IsNull() && m.None.ValueBool():
		return &awstypes.HarnessGatewayOutboundAuthMemberNone{}, diags
	case !m.OAuth.IsNull():
		data, d := m.OAuth.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.HarnessGatewayOutboundAuthMemberOauth
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		return &r, diags
	}
	return nil, diags
}

type harnessOAuthCredentialProviderModel struct {
	CustomParameters fwtypes.MapOfString                         `tfsdk:"custom_parameters"`
	DefaultReturnURL types.String                                `tfsdk:"default_return_url"`
	GrantType        fwtypes.StringEnum[awstypes.OAuthGrantType] `tfsdk:"grant_type"`
	ProviderARN      fwtypes.ARN                                 `tfsdk:"provider_arn"`
	Scopes           fwtypes.ListOfString                        `tfsdk:"scopes"`
}

// Truncation configuration.

type harnessTruncationConfigurationModel struct {
	Config   fwtypes.ListNestedObjectValueOf[harnessTruncationStrategyConfigurationModel] `tfsdk:"config"`
	Strategy fwtypes.StringEnum[awstypes.HarnessTruncationStrategy]                       `tfsdk:"strategy"`
}

func (m *harnessTruncationConfigurationModel) Flatten(ctx context.Context, v *awstypes.HarnessTruncationConfiguration) diag.Diagnostics {
	var diags diag.Diagnostics
	m.Strategy = fwtypes.StringEnumValue(v.Strategy)
	if v.Config != nil {
		var cfg harnessTruncationStrategyConfigurationModel
		smerr.AddEnrich(ctx, &diags, cfg.Flatten(ctx, v.Config))
		m.Config = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &cfg)
	} else {
		m.Config = fwtypes.NewListNestedObjectValueOfNull[harnessTruncationStrategyConfigurationModel](ctx)
	}
	return diags
}

func (m harnessTruncationConfigurationModel) Expand(ctx context.Context) (*awstypes.HarnessTruncationConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := &awstypes.HarnessTruncationConfiguration{
		Strategy: m.Strategy.ValueEnum(),
	}
	if !m.Config.IsNull() {
		configs, d := m.Config.ToSlice(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if len(configs) > 0 {
			expanded, d := configs[0].Expand(ctx)
			smerr.AddEnrich(ctx, &diags, d)
			if expanded != nil {
				out.Config = expanded.(awstypes.HarnessTruncationStrategyConfiguration)
			}
		}
	}
	return out, diags
}

// Truncation strategy configuration union.

type harnessTruncationStrategyConfigurationModel struct {
	SlidingWindow fwtypes.ListNestedObjectValueOf[harnessSlidingWindowConfigModel] `tfsdk:"sliding_window"`
	Summarization fwtypes.ListNestedObjectValueOf[harnessSummarizationConfigModel] `tfsdk:"summarization"`
}

var (
	_ fwflex.Expander  = harnessTruncationStrategyConfigurationModel{}
	_ fwflex.Flattener = &harnessTruncationStrategyConfigurationModel{}
)

func (m *harnessTruncationStrategyConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	m.SlidingWindow = fwtypes.NewListNestedObjectValueOfNull[harnessSlidingWindowConfigModel](ctx)
	m.Summarization = fwtypes.NewListNestedObjectValueOfNull[harnessSummarizationConfigModel](ctx)

	switch t := v.(type) {
	case *awstypes.HarnessTruncationStrategyConfigurationMemberSlidingWindow:
		var data harnessSlidingWindowConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		m.SlidingWindow = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	case *awstypes.HarnessTruncationStrategyConfigurationMemberSummarization:
		var data harnessSummarizationConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		m.Summarization = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("truncation strategy configuration flatten: %T", v))
	}
	return diags
}

func (m harnessTruncationStrategyConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.SlidingWindow.IsNull():
		data, d := m.SlidingWindow.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.HarnessTruncationStrategyConfigurationMemberSlidingWindow
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		return &r, diags
	case !m.Summarization.IsNull():
		data, d := m.Summarization.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.HarnessTruncationStrategyConfigurationMemberSummarization
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		return &r, diags
	}
	return nil, diags
}

type harnessSlidingWindowConfigModel struct {
	MessagesCount types.Int32 `tfsdk:"messages_count"`
}

type harnessSummarizationConfigModel struct {
	SummaryRatio              types.Float32 `tfsdk:"summary_ratio"`
	PreserveRecentMessages    types.Int32   `tfsdk:"preserve_recent_messages"`
	SummarizationSystemPrompt types.String  `tfsdk:"summarization_system_prompt"`
}

// Environment provider union.

type harnessEnvironmentProviderModel struct {
	AgentCoreRuntimeEnvironment fwtypes.ListNestedObjectValueOf[harnessAgentCoreRuntimeEnvironmentModel] `tfsdk:"agentcore_runtime_environment"`
}

func (m *harnessEnvironmentProviderModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	m.AgentCoreRuntimeEnvironment = fwtypes.NewListNestedObjectValueOfNull[harnessAgentCoreRuntimeEnvironmentModel](ctx)

	switch t := v.(type) {
	case *awstypes.HarnessEnvironmentProviderMemberAgentCoreRuntimeEnvironment:
		data := harnessAgentCoreRuntimeEnvironmentModel{
			AgentRuntimeARN:          fwflex.StringToFramework(ctx, t.Value.AgentRuntimeArn),
			AgentRuntimeID:           fwflex.StringToFramework(ctx, t.Value.AgentRuntimeId),
			AgentRuntimeName:         fwflex.StringToFramework(ctx, t.Value.AgentRuntimeName),
			LifecycleConfiguration:   fwtypes.NewListNestedObjectValueOfNull[lifecycleConfigurationModel](ctx),
			NetworkConfiguration:     fwtypes.NewListNestedObjectValueOfNull[networkConfigurationModel](ctx),
			FilesystemConfigurations: fwtypes.NewListNestedObjectValueOfNull[harnessFilesystemConfigurationModel](ctx),
		}

		if t.Value.LifecycleConfiguration != nil {
			var lc lifecycleConfigurationModel
			smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value.LifecycleConfiguration, &lc))
			data.LifecycleConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &lc)
		}

		if t.Value.NetworkConfiguration != nil {
			var nc networkConfigurationModel
			smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value.NetworkConfiguration, &nc))
			data.NetworkConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &nc)
		}

		if len(t.Value.FilesystemConfigurations) > 0 {
			var fsConfigs []*harnessFilesystemConfigurationModel
			for _, fc := range t.Value.FilesystemConfigurations {
				var fsConfig harnessFilesystemConfigurationModel
				smerr.AddEnrich(ctx, &diags, fsConfig.Flatten(ctx, fc))
				fsConfigs = append(fsConfigs, &fsConfig)
			}
			data.FilesystemConfigurations = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, fsConfigs)
		}

		m.AgentCoreRuntimeEnvironment = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("environment provider flatten: %T", v))
	}
	return diags
}

func (m harnessEnvironmentProviderModel) ExpandRequest(ctx context.Context) (awstypes.HarnessEnvironmentProviderRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !m.AgentCoreRuntimeEnvironment.IsNull() {
		data, d := m.AgentCoreRuntimeEnvironment.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var req awstypes.HarnessAgentCoreRuntimeEnvironmentRequest
		if !data.LifecycleConfiguration.IsNull() {
			lc, d := data.LifecycleConfiguration.ToPtr(ctx)
			smerr.AddEnrich(ctx, &diags, d)
			if lc != nil {
				var lcOut awstypes.LifecycleConfiguration
				smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, lc, &lcOut))
				req.LifecycleConfiguration = &lcOut
			}
		}
		if !data.NetworkConfiguration.IsNull() {
			nc, d := data.NetworkConfiguration.ToPtr(ctx)
			smerr.AddEnrich(ctx, &diags, d)
			if nc != nil {
				var ncOut awstypes.NetworkConfiguration
				smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, nc, &ncOut))
				req.NetworkConfiguration = &ncOut
			}
		}
		if !data.FilesystemConfigurations.IsNull() {
			fcs, d := data.FilesystemConfigurations.ToSlice(ctx)
			smerr.AddEnrich(ctx, &diags, d)
			for _, fc := range fcs {
				expanded, d := fc.Expand(ctx)
				smerr.AddEnrich(ctx, &diags, d)
				if expanded != nil {
					req.FilesystemConfigurations = append(req.FilesystemConfigurations, expanded.(awstypes.FilesystemConfiguration))
				}
			}
		}

		return &awstypes.HarnessEnvironmentProviderRequestMemberAgentCoreRuntimeEnvironment{Value: req}, diags
	}
	return nil, diags
}

type harnessAgentCoreRuntimeEnvironmentModel struct {
	AgentRuntimeARN          types.String                                                         `tfsdk:"agent_runtime_arn"`
	AgentRuntimeID           types.String                                                         `tfsdk:"agent_runtime_id"`
	AgentRuntimeName         types.String                                                         `tfsdk:"agent_runtime_name"`
	FilesystemConfigurations fwtypes.ListNestedObjectValueOf[harnessFilesystemConfigurationModel] `tfsdk:"filesystem_configuration"`
	LifecycleConfiguration   fwtypes.ListNestedObjectValueOf[lifecycleConfigurationModel]         `tfsdk:"lifecycle_configuration"`
	NetworkConfiguration     fwtypes.ListNestedObjectValueOf[networkConfigurationModel]           `tfsdk:"network_configuration"`
}

// Filesystem configuration union.

type harnessFilesystemConfigurationModel struct {
	SessionStorage fwtypes.ListNestedObjectValueOf[sessionStorageConfigModel] `tfsdk:"session_storage"`
}

var (
	_ fwflex.Expander  = harnessFilesystemConfigurationModel{}
	_ fwflex.Flattener = &harnessFilesystemConfigurationModel{}
)

func (m *harnessFilesystemConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	m.SessionStorage = fwtypes.NewListNestedObjectValueOfNull[sessionStorageConfigModel](ctx)

	switch t := v.(type) {
	case *awstypes.FilesystemConfigurationMemberSessionStorage:
		var data sessionStorageConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		m.SessionStorage = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("filesystem configuration flatten: %T", v))
	}
	return diags
}

func (m harnessFilesystemConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !m.SessionStorage.IsNull() {
		data, d := m.SessionStorage.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.FilesystemConfigurationMemberSessionStorage
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		return &r, diags
	}
	return nil, diags
}

type sessionStorageConfigModel struct {
	MountPath types.String `tfsdk:"mount_path"`
}

// Environment artifact union.

type harnessEnvironmentArtifactModel struct {
	ContainerConfiguration fwtypes.ListNestedObjectValueOf[containerConfigurationModel] `tfsdk:"container_configuration"`
}

var (
	_ fwflex.Expander  = harnessEnvironmentArtifactModel{}
	_ fwflex.Flattener = &harnessEnvironmentArtifactModel{}
)

func (m *harnessEnvironmentArtifactModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ContainerConfiguration = fwtypes.NewListNestedObjectValueOfNull[containerConfigurationModel](ctx)

	switch t := v.(type) {
	case *awstypes.HarnessEnvironmentArtifactMemberContainerConfiguration:
		var data containerConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		m.ContainerConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("environment artifact flatten: %T", v))
	}
	return diags
}

func (m harnessEnvironmentArtifactModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !m.ContainerConfiguration.IsNull() {
		data, d := m.ContainerConfiguration.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.HarnessEnvironmentArtifactMemberContainerConfiguration
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		return &r, diags
	}
	return nil, diags
}

// Memory configuration union.

type harnessMemoryConfigurationModel struct {
	AgentCoreMemoryConfiguration fwtypes.ListNestedObjectValueOf[harnessAgentCoreMemoryConfigurationModel] `tfsdk:"agentcore_memory_configuration"`
}

var (
	_ fwflex.Expander  = harnessMemoryConfigurationModel{}
	_ fwflex.Flattener = &harnessMemoryConfigurationModel{}
)

func (m *harnessMemoryConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	m.AgentCoreMemoryConfiguration = fwtypes.NewListNestedObjectValueOfNull[harnessAgentCoreMemoryConfigurationModel](ctx)

	switch t := v.(type) {
	case *awstypes.HarnessMemoryConfigurationMemberAgentCoreMemoryConfiguration:
		var data harnessAgentCoreMemoryConfigurationModel
		data.ARN = fwtypes.ARNValue(aws.ToString(t.Value.Arn))
		data.ActorID = fwflex.StringToFramework(ctx, t.Value.ActorId)
		if t.Value.MessagesCount != nil {
			data.MessagesCount = types.Int32Value(*t.Value.MessagesCount)
		}
		if len(t.Value.RetrievalConfig) > 0 {
			var entries []*harnessAgentCoreMemoryRetrievalConfigModel
			for k, v := range t.Value.RetrievalConfig {
				entry := &harnessAgentCoreMemoryRetrievalConfigModel{
					Key: types.StringValue(k),
				}
				if v.TopK != nil {
					entry.TopK = types.Int32Value(*v.TopK)
				}
				if v.RelevanceScore != nil {
					entry.RelevanceScore = types.Float32Value(*v.RelevanceScore)
				}
				entry.StrategyID = fwflex.StringToFramework(ctx, v.StrategyId)
				entries = append(entries, entry)
			}
			data.RetrievalConfig = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, entries)
		}
		m.AgentCoreMemoryConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("memory configuration flatten: %T", v))
	}
	return diags
}

func (m harnessMemoryConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !m.AgentCoreMemoryConfiguration.IsNull() {
		data, d := m.AgentCoreMemoryConfiguration.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		cfg := awstypes.HarnessAgentCoreMemoryConfiguration{
			Arn: aws.String(data.ARN.ValueString()),
		}
		if !data.ActorID.IsNull() {
			cfg.ActorId = data.ActorID.ValueStringPointer()
		}
		if !data.MessagesCount.IsNull() {
			cfg.MessagesCount = aws.Int32(data.MessagesCount.ValueInt32())
		}
		if !data.RetrievalConfig.IsNull() {
			entries, d := data.RetrievalConfig.ToSlice(ctx)
			smerr.AddEnrich(ctx, &diags, d)
			if len(entries) > 0 {
				cfg.RetrievalConfig = make(map[string]awstypes.HarnessAgentCoreMemoryRetrievalConfig)
				for _, e := range entries {
					rc := awstypes.HarnessAgentCoreMemoryRetrievalConfig{}
					if !e.TopK.IsNull() {
						rc.TopK = aws.Int32(e.TopK.ValueInt32())
					}
					if !e.RelevanceScore.IsNull() {
						rc.RelevanceScore = aws.Float32(e.RelevanceScore.ValueFloat32())
					}
					if !e.StrategyID.IsNull() {
						rc.StrategyId = e.StrategyID.ValueStringPointer()
					}
					cfg.RetrievalConfig[e.Key.ValueString()] = rc
				}
			}
		}

		return &awstypes.HarnessMemoryConfigurationMemberAgentCoreMemoryConfiguration{Value: cfg}, diags
	}
	return nil, diags
}

type harnessAgentCoreMemoryConfigurationModel struct {
	ARN             fwtypes.ARN                                                                 `tfsdk:"arn"`
	ActorID         types.String                                                                `tfsdk:"actor_id"`
	MessagesCount   types.Int32                                                                 `tfsdk:"messages_count"`
	RetrievalConfig fwtypes.ListNestedObjectValueOf[harnessAgentCoreMemoryRetrievalConfigModel] `tfsdk:"retrieval_config"`
}

type harnessAgentCoreMemoryRetrievalConfigModel struct {
	MapBlockKey    types.String  `tfsdk:"map_block_key"`
	RelevanceScore types.Float32 `tfsdk:"relevance_score"`
	StrategyID     types.String  `tfsdk:"strategy_id"`
	TopK           types.Int32   `tfsdk:"top_k"`
}
