// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent/document"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"

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
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagent_prompt", name="Prompt")
// @Tags(identifierAttribute="arn")
func newResourcePrompt(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourcePrompt{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNamePrompt = "Prompt"
)

type resourcePrompt struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourcePrompt) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID:  framework.IDAttribute(),
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			"default_variant": schema.StringAttribute{
				Optional: true,
			},
			"customer_encryption_key_arn": schema.StringAttribute{
				Optional: true,
			},
			names.AttrVersion: schema.StringAttribute{
				Computed: true,
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"variant": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[variantModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName: schema.StringAttribute{
							Required: true,
						},
						"model_id": schema.StringAttribute{
							Optional: true,
						},
						"template_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.PromptTemplateType](),
							Required:   true,
						},
						"additional_model_request_fields": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								validators.JSON(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"metadata": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[promptMetadataEntryModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrKey: schema.StringAttribute{
										Required: true,
									},
									names.AttrValue: schema.StringAttribute{
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
						"gen_ai_resource": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[promptGenAiResourceModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("model_id"),
									path.MatchRelative().AtParent().AtName("gen_ai_resource"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"agent": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[promptGenAiResourceMemberAgentModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
											listvalidator.ExactlyOneOf(
												path.MatchRelative().AtParent().AtName("agent"),
											),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"agent_identifier": schema.StringAttribute{
													Required: true,
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
																	Attributes: map[string]schema.Attribute{
																		"text": schema.StringAttribute{
																			Optional: true,
																		},
																	},
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
																CustomType: fwtypes.NewListNestedObjectTypeOf[systemContentBlockMemberCachePointModel](ctx),
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
																					names.AttrType: schema.StringAttribute{
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
																					names.AttrName: schema.StringAttribute{
																						Required: true,
																					},
																					names.AttrDescription: schema.StringAttribute{
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
																								"json": schema.StringAttribute{
																									Optional: true,
																									Validators: []validator.String{
																										stringvalidator.ExactlyOneOf(
																											path.MatchRelative().AtParent().AtName("json"),
																										),
																										validators.JSON(),
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
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourcePrompt) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BedrockAgentClient(ctx)

	var plan resourcePromptModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input bedrockagent.CreatePromptInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input.ClientToken = aws.String(id.UniqueId())
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreatePrompt(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionCreating, ResNamePrompt, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionCreating, ResNamePrompt, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourcePrompt) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BedrockAgentClient(ctx)

	var state resourcePromptModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findPromptByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionSetting, ResNamePrompt, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourcePrompt) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().BedrockAgentClient(ctx)

	var plan, state resourcePromptModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input bedrockagent.UpdatePromptInput
		resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		input.PromptIdentifier = state.ID.ValueStringPointer()

		out, err := conn.UpdatePrompt(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionUpdating, ResNamePrompt, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionUpdating, ResNamePrompt, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourcePrompt) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BedrockAgentClient(ctx)

	var state resourcePromptModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := bedrockagent.DeletePromptInput{
		PromptIdentifier: state.ID.ValueStringPointer(),
	}

	_, err := conn.DeletePrompt(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionDeleting, ResNamePrompt, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourcePrompt) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func findPromptByID(ctx context.Context, conn *bedrockagent.Client, id string) (*bedrockagent.GetPromptOutput, error) {
	in := &bedrockagent.GetPromptInput{
		PromptIdentifier: aws.String(id),
	}

	out, err := conn.GetPrompt(ctx, in)
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

type resourcePromptModel struct {
	ID                       types.String                                  `tfsdk:"id"`
	ARN                      types.String                                  `tfsdk:"arn"`
	Name                     types.String                                  `tfsdk:"name"`
	Description              types.String                                  `tfsdk:"description"`
	DefaultVariant           types.String                                  `tfsdk:"default_variant"`
	CustomerEncryptionKeyARN types.String                                  `tfsdk:"customer_encryption_key_arn"`
	Version                  types.String                                  `tfsdk:"version"`
	CreatedAt                timetypes.RFC3339                             `tfsdk:"created_at"`
	UpdatedAt                timetypes.RFC3339                             `tfsdk:"updated_at"`
	Variants                 fwtypes.ListNestedObjectValueOf[variantModel] `tfsdk:"variant"`
	Timeouts                 timeouts.Value                                `tfsdk:"timeouts"`
	Tags                     tftags.Map                                    `tfsdk:"tags"`
	TagsAll                  tftags.Map                                    `tfsdk:"tags_all"`
}

type variantModel struct {
	Name                         types.String                                                       `tfsdk:"name"`
	ModelID                      types.String                                                       `tfsdk:"model_id"`
	AdditionalModelRequestFields types.String                                                       `tfsdk:"additional_model_request_fields"`
	TemplateType                 fwtypes.StringEnum[awstypes.PromptTemplateType]                    `tfsdk:"template_type"`
	Metadata                     fwtypes.ListNestedObjectValueOf[promptMetadataEntryModel]          `tfsdk:"metadata"`
	InferenceConfiguration       fwtypes.ListNestedObjectValueOf[promptInferenceConfigurationModel] `tfsdk:"inference_configuration"`
	GenAIResource                fwtypes.ListNestedObjectValueOf[promptGenAiResourceModel]          `tfsdk:"gen_ai_resource"`
	TemplateConfiguration        fwtypes.ListNestedObjectValueOf[promptTemplateConfigurationModel]  `tfsdk:"template_configuration"`
}

func (m *variantModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	t := v.(awstypes.PromptVariant)

	diags.Append(flex.Flatten(ctx, t.Name, &m.Name)...)
	diags.Append(flex.Flatten(ctx, t.ModelId, &m.ModelID)...)
	diags.Append(flex.Flatten(ctx, t.TemplateType, &m.TemplateType)...)
	diags.Append(flex.Flatten(ctx, t.Metadata, &m.Metadata)...)
	if t.InferenceConfiguration != nil {
		diags.Append(flex.Flatten(ctx, t.InferenceConfiguration, &m.InferenceConfiguration)...)
	}
	if t.GenAiResource != nil {
		diags.Append(flex.Flatten(ctx, t.GenAiResource, &m.GenAIResource)...)
	}
	diags.Append(flex.Flatten(ctx, t.TemplateConfiguration, &m.TemplateConfiguration)...)
	if diags.HasError() {
		return diags
	}

	if t.AdditionalModelRequestFields != nil {
		additionalFields, err := t.AdditionalModelRequestFields.MarshalSmithyDocument()
		if err != nil {
			diags.AddError("Marshalling additional model request fields", err.Error())
			return diags
		}

		m.AdditionalModelRequestFields = types.StringValue(string(additionalFields))
	}

	return diags
}

func (m variantModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	var r awstypes.PromptVariant
	diags.Append(flex.Expand(ctx, m.Name, &r.Name)...)
	diags.Append(flex.Expand(ctx, m.ModelID, &r.ModelId)...)
	diags.Append(flex.Expand(ctx, m.TemplateType, &r.TemplateType)...)
	diags.Append(flex.Expand(ctx, m.Metadata, &r.Metadata)...)
	diags.Append(flex.Expand(ctx, m.InferenceConfiguration, &r.InferenceConfiguration)...)
	diags.Append(flex.Expand(ctx, m.GenAIResource, &r.GenAiResource)...)
	diags.Append(flex.Expand(ctx, m.TemplateConfiguration, &r.TemplateConfiguration)...)
	if diags.HasError() {
		return nil, diags
	}

	if !m.AdditionalModelRequestFields.IsNull() {
		var doc any
		if err := json.Unmarshal([]byte(m.AdditionalModelRequestFields.ValueString()), &doc); err != nil {
			diags.AddError("Unmarshalling additional model request fields", err.Error())
			return nil, diags
		}
		r.AdditionalModelRequestFields = document.NewLazyDocument(doc)
	}

	return &r, diags
}

type promptMetadataEntryModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
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
type promptGenAiResourceModel struct {
	Agent fwtypes.ListNestedObjectValueOf[promptGenAiResourceMemberAgentModel] `tfsdk:"agent"`
}

func (m *promptGenAiResourceModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.PromptGenAiResourceMemberAgent:
		var model promptGenAiResourceMemberAgentModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Agent = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	default:
		return diags
	}
}

func (m promptGenAiResourceModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.Agent.IsNull():
		promptGenAiResourceAgent, d := m.Agent.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.PromptGenAiResourceMemberAgent
		diags.Append(flex.Expand(ctx, promptGenAiResourceAgent, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

type promptGenAiResourceMemberAgentModel struct {
	AgentIdentifier types.String `tfsdk:"agent_identifier"`
}

// Tagged union
type promptTemplateConfigurationModel struct {
	Chat fwtypes.ListNestedObjectValueOf[promptTemplateConfigurationMemberChatModel] `tfsdk:"chat"`
	Text fwtypes.ListNestedObjectValueOf[promptTemplateConfigurationMemberTextModel] `tfsdk:"text"`
}

func (m *promptTemplateConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
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

func (m promptTemplateConfigurationModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
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
	Text       types.String                                                       `tfsdk:"text"`
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
		m.Text = types.StringValue(t.Value)
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
		contentBlockText, d := m.Text.ToStringValue(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.ContentBlockMemberText
		diags.Append(flex.Expand(ctx, contentBlockText, &r.Value)...)
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

type promptInputVariableModel struct {
	Name types.String `tfsdk:"name"`
}

// Tagged union
type systemContentBlockModel struct {
	CachePoint fwtypes.ListNestedObjectValueOf[systemContentBlockMemberCachePointModel] `tfsdk:"cache_point"`
	Text       types.String                                                             `tfsdk:"text"`
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
		m.Text = types.StringValue(t.Value)
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
		systemContentBlockText, d := m.Text.ToStringValue(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.SystemContentBlockMemberText
		diags.Append(flex.Expand(ctx, systemContentBlockText, &r.Value)...)
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
	Json types.String `tfsdk:"json"`
}

func (m *toolInputSchemaModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.ToolInputSchemaMemberJson:
		if t.Value != nil {
			inputSchema, err := t.Value.MarshalSmithyDocument()
			if err != nil {
				diags.AddError("Marshalling tool input schema", err.Error())
				return diags
			}

			m.Json = types.StringValue(string(inputSchema))
		}

		return diags
	default:
		return diags
	}
}

func (m toolInputSchemaModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.Json.IsNull():
		var r awstypes.ToolInputSchemaMemberJson
		var doc any
		if err := json.Unmarshal([]byte(m.Json.ValueString()), &doc); err != nil {
			diags.AddError("Unmarshalling tool input schema", err.Error())
			return nil, diags
		}
		r.Value = document.NewLazyDocument(doc)

		return &r, diags
	}

	return nil, diags
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
