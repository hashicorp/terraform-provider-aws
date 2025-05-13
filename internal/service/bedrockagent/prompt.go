// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent/document"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagent_prompt", name="Prompt")
// @Tags(identifierAttribute="arn")
func newPromptResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &promptResource{}

	return r, nil
}

type promptResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *promptResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"customer_encryption_key_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
			},
			"default_variant": schema.StringAttribute{
				Optional: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
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
			"variant": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[promptVariantModel](ctx),
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
		},
	}
}

func (r *promptResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data promptResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	var input bedrockagent.CreatePromptInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreatePrompt(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Bedrock Agent Prompt (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *promptResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data promptResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	output, err := findPromptByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Agent Prompt (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *promptResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old promptResourceModel
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
		var input bedrockagent.UpdatePromptInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.PromptIdentifier = old.ID.ValueStringPointer()

		output, err := conn.UpdatePrompt(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Agent Prompt (%s)", new.ID.ValueString()), err.Error())

			return
		}

		response.Diagnostics.Append(fwflex.Flatten(ctx, output, &new)...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *promptResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data promptResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	input := bedrockagent.DeletePromptInput{
		PromptIdentifier: fwflex.StringFromFramework(ctx, data.ID),
	}
	_, err := conn.DeletePrompt(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Bedrock Agent Prompt (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func findPromptByID(ctx context.Context, conn *bedrockagent.Client, id string) (*bedrockagent.GetPromptOutput, error) {
	input := bedrockagent.GetPromptInput{
		PromptIdentifier: aws.String(id),
	}
	output, err := conn.GetPrompt(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

type promptResourceModel struct {
	ARN                      types.String                                        `tfsdk:"arn"`
	CreatedAt                timetypes.RFC3339                                   `tfsdk:"created_at"`
	CustomerEncryptionKeyARN fwtypes.ARN                                         `tfsdk:"customer_encryption_key_arn"`
	DefaultVariant           types.String                                        `tfsdk:"default_variant"`
	Description              types.String                                        `tfsdk:"description"`
	ID                       types.String                                        `tfsdk:"id"`
	Name                     types.String                                        `tfsdk:"name"`
	UpdatedAt                timetypes.RFC3339                                   `tfsdk:"updated_at"`
	Variants                 fwtypes.ListNestedObjectValueOf[promptVariantModel] `tfsdk:"variant"`
	Version                  types.String                                        `tfsdk:"version"`
	Tags                     tftags.Map                                          `tfsdk:"tags"`
	TagsAll                  tftags.Map                                          `tfsdk:"tags_all"`
}

type promptVariantModel struct {
	Name                         types.String                                                       `tfsdk:"name"`
	ModelID                      types.String                                                       `tfsdk:"model_id"`
	AdditionalModelRequestFields types.String                                                       `tfsdk:"additional_model_request_fields"`
	TemplateType                 fwtypes.StringEnum[awstypes.PromptTemplateType]                    `tfsdk:"template_type"`
	Metadata                     fwtypes.ListNestedObjectValueOf[promptMetadataEntryModel]          `tfsdk:"metadata"`
	InferenceConfiguration       fwtypes.ListNestedObjectValueOf[promptInferenceConfigurationModel] `tfsdk:"inference_configuration"`
	GenAIResource                fwtypes.ListNestedObjectValueOf[promptGenAiResourceModel]          `tfsdk:"gen_ai_resource"`
	TemplateConfiguration        fwtypes.ListNestedObjectValueOf[promptTemplateConfigurationModel]  `tfsdk:"template_configuration"`
}

func (m *promptVariantModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	t := v.(awstypes.PromptVariant)

	diags.Append(fwflex.Flatten(ctx, t.Name, &m.Name)...)
	diags.Append(fwflex.Flatten(ctx, t.ModelId, &m.ModelID)...)
	diags.Append(fwflex.Flatten(ctx, t.TemplateType, &m.TemplateType)...)
	diags.Append(fwflex.Flatten(ctx, t.Metadata, &m.Metadata)...)
	if t.InferenceConfiguration != nil {
		diags.Append(fwflex.Flatten(ctx, t.InferenceConfiguration, &m.InferenceConfiguration)...)
	}
	if t.GenAiResource != nil {
		diags.Append(fwflex.Flatten(ctx, t.GenAiResource, &m.GenAIResource)...)
	}
	diags.Append(fwflex.Flatten(ctx, t.TemplateConfiguration, &m.TemplateConfiguration)...)
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

func (m promptVariantModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	var r awstypes.PromptVariant
	diags.Append(fwflex.Expand(ctx, m.Name, &r.Name)...)
	diags.Append(fwflex.Expand(ctx, m.ModelID, &r.ModelId)...)
	diags.Append(fwflex.Expand(ctx, m.TemplateType, &r.TemplateType)...)
	diags.Append(fwflex.Expand(ctx, m.Metadata, &r.Metadata)...)
	diags.Append(fwflex.Expand(ctx, m.InferenceConfiguration, &r.InferenceConfiguration)...)
	diags.Append(fwflex.Expand(ctx, m.GenAIResource, &r.GenAiResource)...)
	diags.Append(fwflex.Expand(ctx, m.TemplateConfiguration, &r.TemplateConfiguration)...)
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
		d := fwflex.Flatten(ctx, t.Value, &model)
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
		diags.Append(fwflex.Expand(ctx, promptInferenceConfigurationText, &r.Value)...)
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
		d := fwflex.Flatten(ctx, t.Value, &model)
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
		diags.Append(fwflex.Expand(ctx, promptGenAiResourceAgent, &r.Value)...)
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
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Chat = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.PromptTemplateConfigurationMemberText:
		var model promptTemplateConfigurationMemberTextModel
		d := fwflex.Flatten(ctx, t.Value, &model)
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
		diags.Append(fwflex.Expand(ctx, promptTemplateConfigurationChat, &r.Value)...)
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
		diags.Append(fwflex.Expand(ctx, promptTemplateConfigurationText, &r.Value)...)
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
		d := fwflex.Flatten(ctx, t.Value, &model)
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
		diags.Append(fwflex.Expand(ctx, contentBlockCachePoint, &r.Value)...)
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
		diags.Append(fwflex.Expand(ctx, contentBlockText, &r.Value)...)
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
		d := fwflex.Flatten(ctx, t.Value, &model)
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
		diags.Append(fwflex.Expand(ctx, systemContentBlockCachePoint, &r.Value)...)
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
		diags.Append(fwflex.Expand(ctx, systemContentBlockText, &r.Value)...)
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
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.CachePoint = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.ToolMemberToolSpec:
		var model toolMemberToolSpecModel
		d := fwflex.Flatten(ctx, t.Value, &model)
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
		diags.Append(fwflex.Expand(ctx, toolCachePoint, &r.Value)...)
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
		diags.Append(fwflex.Expand(ctx, toolToolSpec, &r.Value)...)
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
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Any = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.ToolChoiceMemberAuto:
		var model toolChoiceMemberAutoModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.Auto = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.ToolChoiceMemberTool:
		var model toolChoiceMemberToolModel
		d := fwflex.Flatten(ctx, t.Value, &model)
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
		diags.Append(fwflex.Expand(ctx, toolChoiceAny, &r.Value)...)
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
		diags.Append(fwflex.Expand(ctx, toolChoiceAuto, &r.Value)...)
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
		diags.Append(fwflex.Expand(ctx, toolChoiceTool, &r.Value)...)
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
