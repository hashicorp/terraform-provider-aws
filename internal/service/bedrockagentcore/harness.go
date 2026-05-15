// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/document"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/float32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tfsmithy "github.com/hashicorp/terraform-provider-aws/internal/smithy"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagentcore_harness", name="Harness")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("harness_id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types;awstypes;awstypes.Harness")
// @Testing(generator="testAccRandomHarnessName(t)")
// @Testing(tagsTest=false)
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttribute="harness_id")
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
	framework.WithImportByIdentity
}

func (r *harnessResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"allowed_tools": schema.ListAttribute{
				CustomType: fwtypes.ListOfStringType,
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrARN:         framework.ARNAttributeComputedOnly(),
			names.AttrEnvironment: framework.ResourceOptionalComputedListOfObjectsAttribute[harnessEnvironmentProviderModel](ctx, 1, nil, listplanmodifier.UseStateForUnknown()),
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
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
			},
			"max_tokens": schema.Int32Attribute{
				Optional: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"timeout_seconds": schema.Int32Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
			},
			"truncation": framework.ResourceOptionalComputedListOfObjectsAttribute[harnessTruncationConfigurationModel](ctx, 1, nil, listplanmodifier.UseStateForUnknown()),
		},
		Blocks: map[string]schema.Block{
			"authorizer_configuration": authorizerConfigurationSchema(ctx),
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
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{ // nosemgrep:ci.semgrep.framework.map_block_key-meaningful-names
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
													CustomType: jsontypes.NormalizedType{},
													Required:   true,
													Sensitive:  true,
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

	var input bedrockagentcorecontrol.CreateHarnessInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	var (
		out *bedrockagentcorecontrol.CreateHarnessOutput
		err error
	)
	err = tfresource.Retry(ctx, propagationTimeout, func(ctx context.Context) *tfresource.RetryError {
		out, err = conn.CreateHarness(ctx, &input)

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

	// Set values for unknowns.
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, harness, &data, fwflex.WithFieldNamePrefix("AgentRuntime")))
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

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, harness, &data))
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

	if diff.HasChanges() {
		harnessID := fwflex.StringValueFromFramework(ctx, new.HarnessID)
		var input bedrockagentcorecontrol.UpdateHarnessInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input))
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.ClientToken = aws.String(create.UniqueId(ctx))

		_, err := conn.UpdateHarness(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, harnessID)
			return
		}

		if _, err := waitHarnessUpdated(ctx, conn, harnessID, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, harnessID)
			return
		}
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

func (r *harnessResource) flatten(ctx context.Context, harness *awstypes.Harness, data *harnessResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.Append(fwflex.Flatten(ctx, harness, data)...)
	return diags
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
		retry.SetLastError(err, errors.New(aws.ToString(out.FailureReason)))
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
		retry.SetLastError(err, errors.New(aws.ToString(out.FailureReason)))
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
		retry.SetLastError(err, errors.New(aws.ToString(out.FailureReason)))
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
	switch t := v.(type) {
	case awstypes.HarnessModelConfigurationMemberBedrockModelConfig:
		var data harnessBedrockModelConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.BedrockModelConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	case awstypes.HarnessModelConfigurationMemberGeminiModelConfig:
		var data harnessGeminiModelConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.GeminiModelConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	case awstypes.HarnessModelConfigurationMemberOpenAiModelConfig:
		var data harnessOpenAIModelConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.OpenAiModelConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
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
	case !m.GeminiModelConfig.IsNull():
		data, d := m.GeminiModelConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.HarnessModelConfigurationMemberGeminiModelConfig
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
	case awstypes.HarnessSystemContentBlockMemberText:
		m.Text = fwflex.StringValueToFramework(ctx, t.Value)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("system content block flatten: %T", v))
	}
	return diags
}

func (m harnessSystemContentBlockModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !m.Text.IsNull() {
		return &awstypes.HarnessSystemContentBlockMemberText{Value: fwflex.StringValueFromFramework(ctx, m.Text)}, diags
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
	case awstypes.HarnessSkillMemberPath:
		m.Path = fwflex.StringValueToFramework(ctx, t.Value)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("skill flatten: %T", v))
	}
	return diags
}

func (m harnessSkillModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !m.Path.IsNull() {
		return &awstypes.HarnessSkillMemberPath{Value: fwflex.StringValueFromFramework(ctx, m.Path)}, diags
	}
	return nil, diags
}

// Tool model.

type harnessToolModel struct {
	Config fwtypes.ListNestedObjectValueOf[harnessToolConfigurationModel] `tfsdk:"config"`
	Name   types.String                                                   `tfsdk:"name"`
	Type   fwtypes.StringEnum[awstypes.HarnessToolType]                   `tfsdk:"type"`
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
	switch t := v.(type) {
	case awstypes.HarnessToolConfigurationMemberAgentCoreBrowser:
		var data harnessAgentCoreBrowserConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.AgentCoreBrowser = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	case awstypes.HarnessToolConfigurationMemberAgentCoreCodeInterpreter:
		var data harnessAgentCoreCodeInterpreterConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.AgentCoreCodeInterpreter = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	case awstypes.HarnessToolConfigurationMemberAgentCoreGateway:
		var data harnessAgentCoreGatewayConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.AgentCoreGateway = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	case awstypes.HarnessToolConfigurationMemberInlineFunction:
		data := harnessInlineFunctionConfigModel{
			Description: fwflex.StringToFramework(ctx, t.Value.Description),
		}
		if t.Value.InputSchema != nil {
			if json, err := tfsmithy.DocumentToJSONString(t.Value.InputSchema); err == nil {
				data.InputSchema = jsontypes.NewNormalizedValue(json)
			}
		}
		m.InlineFunction = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	case awstypes.HarnessToolConfigurationMemberRemoteMcp:
		var data harnessRemoteMCPConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.RemoteMcp = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("tool configuration flatten: %T", v))
	}
	return diags
}

func (m harnessToolConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.AgentCoreBrowser.IsNull():
		data, d := m.AgentCoreBrowser.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.HarnessToolConfigurationMemberAgentCoreBrowser
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
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
			Description: fwflex.StringFromFramework(ctx, data.Description),
		}
		if !data.InputSchema.IsNull() {
			if json, err := tfsmithy.DocumentFromJSONString(fwflex.StringValueFromFramework(ctx, data.InputSchema), document.NewLazyDocument); err == nil {
				r.Value.InputSchema = json
			}
		}
		return &r, diags
	case !m.RemoteMcp.IsNull():
		data, d := m.RemoteMcp.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.HarnessToolConfigurationMemberRemoteMcp
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		return &r, diags
	}
	return nil, diags
}

type harnessAgentCoreBrowserConfigModel struct {
	BrowserARN fwtypes.ARN `tfsdk:"browser_arn"`
}

type harnessAgentCoreCodeInterpreterConfigModel struct {
	CodeInterpreterARN fwtypes.ARN `tfsdk:"code_interpreter_arn"`
}

type harnessAgentCoreGatewayConfigModel struct {
	GatewayARN   fwtypes.ARN                                                      `tfsdk:"gateway_arn"`
	OutboundAuth fwtypes.ListNestedObjectValueOf[harnessGatewayOutboundAuthModel] `tfsdk:"outbound_auth"`
}

type harnessInlineFunctionConfigModel struct {
	Description types.String         `tfsdk:"description"`
	InputSchema jsontypes.Normalized `tfsdk:"input_schema" autoflex:"-"`
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
	m.OAuth = fwtypes.NewListNestedObjectValueOfNull[harnessOAuthCredentialProviderModel](ctx)
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.HarnessGatewayOutboundAuthMemberAwsIam:
		m.AwsIam = types.BoolValue(true)
	case awstypes.HarnessGatewayOutboundAuthMemberNone:
		m.None = types.BoolValue(true)
	case awstypes.HarnessGatewayOutboundAuthMemberOauth:
		var data harnessOAuthCredentialProviderModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
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
	switch t := v.(type) {
	case awstypes.HarnessTruncationStrategyConfigurationMemberSlidingWindow:
		var data harnessSlidingWindowConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.SlidingWindow = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	case awstypes.HarnessTruncationStrategyConfigurationMemberSummarization:
		var data harnessSummarizationConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
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

var (
	_ fwflex.Expander  = harnessEnvironmentProviderModel{}
	_ fwflex.Flattener = &harnessEnvironmentProviderModel{}
)

func (m *harnessEnvironmentProviderModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.HarnessEnvironmentProviderMemberAgentCoreRuntimeEnvironment:
		var data harnessAgentCoreRuntimeEnvironmentModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.AgentCoreRuntimeEnvironment = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("environment provider flatten: %T", v))
	}
	return diags
}

func (m harnessEnvironmentProviderModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !m.AgentCoreRuntimeEnvironment.IsNull() {
		data, d := m.AgentCoreRuntimeEnvironment.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.HarnessEnvironmentProviderRequestMemberAgentCoreRuntimeEnvironment
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		return &r, diags
	}
	return nil, diags
}

type harnessAgentCoreRuntimeEnvironmentModel struct {
	AgentRuntimeARN          types.String                                                  `tfsdk:"agent_runtime_arn"`
	AgentRuntimeID           types.String                                                  `tfsdk:"agent_runtime_id"`
	AgentRuntimeName         types.String                                                  `tfsdk:"agent_runtime_name"`
	FilesystemConfigurations fwtypes.ListNestedObjectValueOf[filesystemConfigurationModel] `tfsdk:"filesystem_configuration"`
	LifecycleConfiguration   fwtypes.ListNestedObjectValueOf[lifecycleConfigurationModel]  `tfsdk:"lifecycle_configuration"`
	NetworkConfiguration     fwtypes.ListNestedObjectValueOf[networkConfigurationModel]    `tfsdk:"network_configuration"`
}

// Environment artifact union.

type harnessEnvironmentArtifactModel struct {
	ContainerConfiguration fwtypes.ListNestedObjectValueOf[containerConfigurationModel] `tfsdk:"container_configuration"`
}

var (
	_ fwflex.TypedExpander = harnessEnvironmentArtifactModel{}
	_ fwflex.Flattener     = &harnessEnvironmentArtifactModel{}
)

func (m *harnessEnvironmentArtifactModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.HarnessEnvironmentArtifactMemberContainerConfiguration:
		var data containerConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.ContainerConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("environment artifact flatten: %T", v))
	}
	return diags
}

func (m harnessEnvironmentArtifactModel) ExpandTo(ctx context.Context, targetType reflect.Type) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch targetType {
	case reflect.TypeFor[awstypes.HarnessEnvironmentArtifact]():
		return m.expandToHarnessEnvironmentArtifact(ctx)

	case reflect.TypeFor[awstypes.UpdatedHarnessEnvironmentArtifact]():
		return m.expandToUpdatedHarnessEnvironmentArtifact(ctx)
	}
	return nil, diags
}

func (m harnessEnvironmentArtifactModel) expandToHarnessEnvironmentArtifact(ctx context.Context) (awstypes.HarnessEnvironmentArtifact, diag.Diagnostics) {
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

func (m harnessEnvironmentArtifactModel) expandToUpdatedHarnessEnvironmentArtifact(ctx context.Context) (*awstypes.UpdatedHarnessEnvironmentArtifact, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !m.ContainerConfiguration.IsNull() {
		r, d := m.expandToHarnessEnvironmentArtifact(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		return &awstypes.UpdatedHarnessEnvironmentArtifact{OptionalValue: r}, diags
	}
	return &awstypes.UpdatedHarnessEnvironmentArtifact{}, diags
}

// Memory configuration union.

type harnessMemoryConfigurationModel struct {
	AgentCoreMemoryConfiguration fwtypes.ListNestedObjectValueOf[harnessAgentCoreMemoryConfigurationModel] `tfsdk:"agentcore_memory_configuration"`
}

var (
	_ fwflex.TypedExpander = harnessMemoryConfigurationModel{}
	_ fwflex.Flattener     = &harnessMemoryConfigurationModel{}
)

func (m *harnessMemoryConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.HarnessMemoryConfigurationMemberAgentCoreMemoryConfiguration:
		var data harnessAgentCoreMemoryConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.AgentCoreMemoryConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("memory configuration flatten: %T", v))
	}
	return diags
}

func (m harnessMemoryConfigurationModel) ExpandTo(ctx context.Context, targetType reflect.Type) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch targetType {
	case reflect.TypeFor[awstypes.HarnessMemoryConfiguration]():
		return m.expandToHarnessMemoryConfiguration(ctx)

	case reflect.TypeFor[awstypes.UpdatedHarnessMemoryConfiguration]():
		return m.expandToUpdatedHarnessMemoryConfiguration(ctx)
	}
	return nil, diags
}

func (m harnessMemoryConfigurationModel) expandToHarnessMemoryConfiguration(ctx context.Context) (awstypes.HarnessMemoryConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !m.AgentCoreMemoryConfiguration.IsNull() {
		data, d := m.AgentCoreMemoryConfiguration.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.HarnessMemoryConfigurationMemberAgentCoreMemoryConfiguration
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		return &r, diags
	}
	return nil, diags
}

func (m harnessMemoryConfigurationModel) expandToUpdatedHarnessMemoryConfiguration(ctx context.Context) (*awstypes.UpdatedHarnessMemoryConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !m.AgentCoreMemoryConfiguration.IsNull() {
		r, d := m.expandToHarnessMemoryConfiguration(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		return &awstypes.UpdatedHarnessMemoryConfiguration{OptionalValue: r}, diags
	}
	return &awstypes.UpdatedHarnessMemoryConfiguration{}, diags
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
