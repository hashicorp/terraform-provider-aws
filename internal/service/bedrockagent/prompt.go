// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrockagent

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent/document"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
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

// @FrameworkResource("aws_bedrockagent_prompt", name="Prompt")
// @Tags(identifierAttribute="arn")
func newPromptResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &promptResource{}

	return r, nil
}

type promptResource struct {
	framework.ResourceWithModel[promptResourceModel]
	framework.WithImportByID
}

func (r *promptResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
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
						"additional_model_request_fields": schema.StringAttribute{
							CustomType: jsontypes.NormalizedType{},
							Optional:   true,
						},
						"model_id": schema.StringAttribute{
							Optional: true,
						},
						names.AttrName: schema.StringAttribute{
							Required: true,
						},
						"template_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.PromptTemplateType](),
							Required:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"gen_ai_resource": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[promptGenAiResourceModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("gen_ai_resource"),
									path.MatchRelative().AtParent().AtName("model_id"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"agent": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[promptAgentResourceModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
											listvalidator.ExactlyOneOf(
												path.MatchRelative().AtParent().AtName("agent"),
											),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"agent_identifier": schema.StringAttribute{
													CustomType: fwtypes.ARNType,
													Required:   true,
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
						"metadata": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[promptMetadataEntryModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeBetween(0, 50),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrKey: schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 128),
										},
									},
									names.AttrValue: schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(0, 1024),
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
	data.ARN = fwflex.StringToFramework(ctx, output.Arn)
	data.CreatedAt = timetypes.NewRFC3339TimePointerValue(output.CreatedAt)
	data.ID = fwflex.StringToFramework(ctx, output.Id)
	data.UpdatedAt = timetypes.NewRFC3339TimePointerValue(output.UpdatedAt)
	data.Version = fwflex.StringToFramework(ctx, output.Version)

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

	if retry.NotFound(err) {
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

	if !new.CustomerEncryptionKeyARN.Equal(old.CustomerEncryptionKeyARN) ||
		!new.DefaultVariant.Equal(old.DefaultVariant) ||
		!new.Description.Equal(old.Description) ||
		!new.Name.Equal(old.Name) ||
		!new.Variants.Equal(old.Variants) {
		var input bedrockagent.UpdatePromptInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.PromptIdentifier = fwflex.StringFromFramework(ctx, new.ID)

		output, err := conn.UpdatePrompt(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Bedrock Agent Prompt (%s)", new.ID.ValueString()), err.Error())

			return
		}

		// Set values for unknowns.
		new.UpdatedAt = timetypes.NewRFC3339TimePointerValue(output.UpdatedAt)
		new.Version = fwflex.StringToFramework(ctx, output.Version)
	} else {
		new.UpdatedAt = old.UpdatedAt
		new.Version = old.Version
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

type promptResourceModel struct {
	framework.WithRegionModel
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
	AdditionalModelRequestFields jsontypes.Normalized                                               `tfsdk:"additional_model_request_fields"  autoflex:"-"`
	GenAIResource                fwtypes.ListNestedObjectValueOf[promptGenAiResourceModel]          `tfsdk:"gen_ai_resource"`
	InferenceConfiguration       fwtypes.ListNestedObjectValueOf[promptInferenceConfigurationModel] `tfsdk:"inference_configuration"`
	Metadata                     fwtypes.ListNestedObjectValueOf[promptMetadataEntryModel]          `tfsdk:"metadata"`
	ModelID                      types.String                                                       `tfsdk:"model_id"`
	Name                         types.String                                                       `tfsdk:"name"`
	TemplateConfiguration        fwtypes.ListNestedObjectValueOf[promptTemplateConfigurationModel]  `tfsdk:"template_configuration"`
	TemplateType                 fwtypes.StringEnum[awstypes.PromptTemplateType]                    `tfsdk:"template_type"`
}

var (
	_ fwflex.Expander  = promptVariantModel{}
	_ fwflex.Flattener = &promptVariantModel{}
)

func (m promptVariantModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var result any
	var diags diag.Diagnostics

	var r awstypes.PromptVariant
	diags.Append(fwflex.Expand(ctx, m.GenAIResource, &r.GenAiResource)...)
	diags.Append(fwflex.Expand(ctx, m.InferenceConfiguration, &r.InferenceConfiguration)...)
	diags.Append(fwflex.Expand(ctx, m.Metadata, &r.Metadata)...)
	diags.Append(fwflex.Expand(ctx, m.ModelID, &r.ModelId)...)
	diags.Append(fwflex.Expand(ctx, m.Name, &r.Name)...)
	diags.Append(fwflex.Expand(ctx, m.TemplateConfiguration, &r.TemplateConfiguration)...)
	diags.Append(fwflex.Expand(ctx, m.TemplateType, &r.TemplateType)...)
	if diags.HasError() {
		return nil, diags
	}

	if !m.AdditionalModelRequestFields.IsNull() {
		json, err := tfsmithy.DocumentFromJSONString(fwflex.StringValueFromFramework(ctx, m.AdditionalModelRequestFields), document.NewLazyDocument)
		if err != nil {
			diags.Append(diag.NewErrorDiagnostic(
				"Decoding JSON",
				err.Error(),
			))

			return nil, diags
		}

		r.AdditionalModelRequestFields = json
	}

	result = &r

	return result, diags
}

func (m *promptVariantModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	if v, ok := v.(awstypes.PromptVariant); ok {
		if v.GenAiResource != nil {
			diags.Append(fwflex.Flatten(ctx, v.GenAiResource, &m.GenAIResource)...)
		}
		if v.InferenceConfiguration != nil {
			diags.Append(fwflex.Flatten(ctx, v.InferenceConfiguration, &m.InferenceConfiguration)...)
		}
		diags.Append(fwflex.Flatten(ctx, v.Metadata, &m.Metadata)...)
		diags.Append(fwflex.Flatten(ctx, v.ModelId, &m.ModelID)...)
		diags.Append(fwflex.Flatten(ctx, v.Name, &m.Name)...)
		diags.Append(fwflex.Flatten(ctx, v.TemplateConfiguration, &m.TemplateConfiguration)...)
		diags.Append(fwflex.Flatten(ctx, v.TemplateType, &m.TemplateType)...)
		if diags.HasError() {
			return diags
		}

		if v.AdditionalModelRequestFields != nil {
			json, err := tfsmithy.DocumentToJSONString(v.AdditionalModelRequestFields)
			if err != nil {
				diags.Append(diag.NewErrorDiagnostic(
					"Encoding JSON",
					err.Error(),
				))

				return diags
			}

			m.AdditionalModelRequestFields = jsontypes.NewNormalizedValue(json)
		}
	}

	return diags
}

type promptGenAiResourceModel struct {
	Agent fwtypes.ListNestedObjectValueOf[promptAgentResourceModel] `tfsdk:"agent"`
}

var (
	_ fwflex.Expander  = promptGenAiResourceModel{}
	_ fwflex.Flattener = &promptGenAiResourceModel{}
)

func (m promptGenAiResourceModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var result any
	var diags diag.Diagnostics

	switch {
	case !m.Agent.IsNull():
		agent, d := m.Agent.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.PromptGenAiResourceMemberAgent
		diags.Append(fwflex.Expand(ctx, agent, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		result = &r
	}

	return result, diags
}

func (m *promptGenAiResourceModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch v := v.(type) {
	case awstypes.PromptGenAiResourceMemberAgent:
		var agent promptAgentResourceModel
		diags.Append(fwflex.Flatten(ctx, v.Value, &agent)...)
		if diags.HasError() {
			return diags
		}

		m.Agent = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &agent)
	}

	return diags
}

type promptAgentResourceModel struct {
	AgentIdentifier fwtypes.ARN `tfsdk:"agent_identifier"`
}

type promptInferenceConfigurationModel struct {
	Text fwtypes.ListNestedObjectValueOf[promptModelInferenceConfigurationModel] `tfsdk:"text"`
}

var (
	_ fwflex.Expander  = promptInferenceConfigurationModel{}
	_ fwflex.Flattener = &promptInferenceConfigurationModel{}
)

func (m promptInferenceConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var result any
	var diags diag.Diagnostics

	switch {
	case !m.Text.IsNull():
		text, d := m.Text.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.PromptInferenceConfigurationMemberText
		diags.Append(fwflex.Expand(ctx, text, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		result = &r
	}

	return result, diags
}

func (m *promptInferenceConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch v := v.(type) {
	case awstypes.PromptInferenceConfigurationMemberText:
		var text promptModelInferenceConfigurationModel
		diags.Append(fwflex.Flatten(ctx, v.Value, &text)...)
		if diags.HasError() {
			return diags
		}

		m.Text = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &text)
	}

	return diags
}

type promptModelInferenceConfigurationModel struct {
	MaxTokens     types.Int32          `tfsdk:"max_tokens"`
	StopSequences fwtypes.ListOfString `tfsdk:"stop_sequences"`
	Temperature   types.Float32        `tfsdk:"temperature"`
	TopP          types.Float32        `tfsdk:"top_p"`
}

type promptMetadataEntryModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type promptTemplateConfigurationModel struct {
	Chat fwtypes.ListNestedObjectValueOf[chatPromptTemplateConfigurationModel] `tfsdk:"chat"`
	Text fwtypes.ListNestedObjectValueOf[textPromptTemplateConfigurationModel] `tfsdk:"text"`
}

var (
	_ fwflex.Expander  = promptTemplateConfigurationModel{}
	_ fwflex.Flattener = &promptTemplateConfigurationModel{}
)

func (m promptTemplateConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var result any
	var diags diag.Diagnostics

	switch {
	case !m.Chat.IsNull():
		chat, d := m.Chat.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.PromptTemplateConfigurationMemberChat
		diags.Append(fwflex.Expand(ctx, chat, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		result = &r
	case !m.Text.IsNull():
		text, d := m.Text.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.PromptTemplateConfigurationMemberText
		diags.Append(fwflex.Expand(ctx, text, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		result = &r
	}

	return result, diags
}

func (m *promptTemplateConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch v := v.(type) {
	case awstypes.PromptTemplateConfigurationMemberChat:
		var chat chatPromptTemplateConfigurationModel
		diags.Append(fwflex.Flatten(ctx, v.Value, &chat)...)
		if diags.HasError() {
			return diags
		}

		m.Chat = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &chat)
	case awstypes.PromptTemplateConfigurationMemberText:
		var text textPromptTemplateConfigurationModel
		diags.Append(fwflex.Flatten(ctx, v.Value, &text)...)
		if diags.HasError() {
			return diags
		}

		m.Text = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &text)
	}

	return diags
}

type chatPromptTemplateConfigurationModel struct {
	InputVariables    fwtypes.ListNestedObjectValueOf[promptInputVariableModel] `tfsdk:"input_variable"`
	Messages          fwtypes.ListNestedObjectValueOf[messageModel]             `tfsdk:"message"`
	System            fwtypes.ListNestedObjectValueOf[systemContentBlockModel]  `tfsdk:"system"`
	ToolConfiguration fwtypes.ListNestedObjectValueOf[toolConfigurationModel]   `tfsdk:"tool_configuration"`
}

type promptInputVariableModel struct {
	Name types.String `tfsdk:"name"`
}

type messageModel struct {
	Content fwtypes.ListNestedObjectValueOf[contentBlockModel] `tfsdk:"content"`
	Role    fwtypes.StringEnum[awstypes.ConversationRole]      `tfsdk:"role"`
}

type contentBlockModel struct {
	CachePoint fwtypes.ListNestedObjectValueOf[cachePointBlockModel] `tfsdk:"cache_point"`
	Text       types.String                                          `tfsdk:"text"`
}

var (
	_ fwflex.Expander  = contentBlockModel{}
	_ fwflex.Flattener = &contentBlockModel{}
)

func (m contentBlockModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var result any
	var diags diag.Diagnostics

	switch {
	case !m.CachePoint.IsNull():
		cachePoint, d := m.CachePoint.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.ContentBlockMemberCachePoint
		diags.Append(fwflex.Expand(ctx, cachePoint, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		result = &r
	case !m.Text.IsNull():
		text, d := m.Text.ToStringValue(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.ContentBlockMemberText
		diags.Append(fwflex.Expand(ctx, text, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		result = &r
	}

	return result, diags
}

func (m *contentBlockModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch v := v.(type) {
	case awstypes.ContentBlockMemberCachePoint:
		var cachePoint cachePointBlockModel
		diags.Append(fwflex.Flatten(ctx, v.Value, &cachePoint)...)
		if diags.HasError() {
			return diags
		}

		m.CachePoint = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &cachePoint)
	case awstypes.ContentBlockMemberText:
		m.Text = types.StringValue(v.Value)
	}

	return diags
}

type cachePointBlockModel struct {
	Type fwtypes.StringEnum[awstypes.CachePointType] `tfsdk:"type"`
}

type systemContentBlockModel struct {
	CachePoint fwtypes.ListNestedObjectValueOf[cachePointBlockModel] `tfsdk:"cache_point"`
	Text       types.String                                          `tfsdk:"text"`
}

var (
	_ fwflex.Expander  = systemContentBlockModel{}
	_ fwflex.Flattener = &systemContentBlockModel{}
)

func (m systemContentBlockModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var result any
	var diags diag.Diagnostics

	switch {
	case !m.CachePoint.IsNull():
		cachePoint, d := m.CachePoint.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.SystemContentBlockMemberCachePoint
		diags.Append(fwflex.Expand(ctx, cachePoint, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		result = &r
	case !m.Text.IsNull():
		text, d := m.Text.ToStringValue(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.SystemContentBlockMemberText
		diags.Append(fwflex.Expand(ctx, text, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		result = &r
	}

	return result, diags
}

func (m *systemContentBlockModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch v := v.(type) {
	case awstypes.SystemContentBlockMemberCachePoint:
		var cachePoint cachePointBlockModel
		diags.Append(fwflex.Flatten(ctx, v.Value, &cachePoint)...)
		if diags.HasError() {
			return diags
		}

		m.CachePoint = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &cachePoint)
	case awstypes.SystemContentBlockMemberText:
		m.Text = types.StringValue(v.Value)
	}

	return diags
}

type toolConfigurationModel struct {
	ToolChoice fwtypes.ListNestedObjectValueOf[toolChoiceModel] `tfsdk:"tool_choice"`
	Tools      fwtypes.ListNestedObjectValueOf[toolModel]       `tfsdk:"tool"`
}

type toolModel struct {
	CachePoint fwtypes.ListNestedObjectValueOf[cachePointBlockModel]   `tfsdk:"cache_point"`
	ToolSpec   fwtypes.ListNestedObjectValueOf[toolSpecificationModel] `tfsdk:"tool_spec"`
}

var (
	_ fwflex.Expander  = toolModel{}
	_ fwflex.Flattener = &toolModel{}
)

func (m toolModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var result any
	var diags diag.Diagnostics

	switch {
	case !m.CachePoint.IsNull():
		cachePoint, d := m.CachePoint.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.ToolMemberCachePoint
		diags.Append(fwflex.Expand(ctx, cachePoint, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		result = &r
	case !m.ToolSpec.IsNull():
		toolSpec, d := m.ToolSpec.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.ToolMemberToolSpec
		diags.Append(fwflex.Expand(ctx, toolSpec, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		result = &r
	}

	return result, diags
}

func (m *toolModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch v := v.(type) {
	case awstypes.ToolMemberCachePoint:
		var cachePoint cachePointBlockModel
		diags.Append(fwflex.Flatten(ctx, v.Value, &cachePoint)...)
		if diags.HasError() {
			return diags
		}

		m.CachePoint = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &cachePoint)
	case awstypes.ToolMemberToolSpec:
		var toolSpec toolSpecificationModel
		diags.Append(fwflex.Flatten(ctx, v.Value, &toolSpec)...)
		if diags.HasError() {
			return diags
		}

		m.ToolSpec = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &toolSpec)
	}

	return diags
}

type toolSpecificationModel struct {
	Description types.String                                          `tfsdk:"description"`
	InputSchema fwtypes.ListNestedObjectValueOf[toolInputSchemaModel] `tfsdk:"input_schema"`
	Name        types.String                                          `tfsdk:"name"`
}

type toolInputSchemaModel struct {
	JSON jsontypes.Normalized `tfsdk:"json" autoflex:"-"`
}

var (
	_ fwflex.Expander  = toolInputSchemaModel{}
	_ fwflex.Flattener = &toolInputSchemaModel{}
)

func (m toolInputSchemaModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var result any
	var diags diag.Diagnostics

	switch {
	case !m.JSON.IsNull():
		json, err := tfsmithy.DocumentFromJSONString(fwflex.StringValueFromFramework(ctx, m.JSON), document.NewLazyDocument)
		if err != nil {
			diags.Append(diag.NewErrorDiagnostic(
				"Decoding JSON",
				err.Error(),
			))

			return nil, diags
		}

		var r awstypes.ToolInputSchemaMemberJson
		r.Value = json

		result = &r
	}

	return result, diags
}

func (m *toolInputSchemaModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch v := v.(type) {
	case awstypes.ToolInputSchemaMemberJson:
		if v.Value != nil {
			json, err := tfsmithy.DocumentToJSONString(v.Value)
			if err != nil {
				diags.Append(diag.NewErrorDiagnostic(
					"Encoding JSON",
					err.Error(),
				))

				return diags
			}

			m.JSON = jsontypes.NewNormalizedValue(json)
		}
	}

	return diags
}

type toolChoiceModel struct {
	Any  fwtypes.ListNestedObjectValueOf[anyToolChoiceModel]      `tfsdk:"any"`
	Auto fwtypes.ListNestedObjectValueOf[autoToolChoiceModel]     `tfsdk:"auto"`
	Tool fwtypes.ListNestedObjectValueOf[specificToolChoiceModel] `tfsdk:"tool"`
}

var (
	_ fwflex.Expander  = toolChoiceModel{}
	_ fwflex.Flattener = &toolChoiceModel{}
)

func (m toolChoiceModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var result any
	var diags diag.Diagnostics

	switch {
	case !m.Any.IsNull():
		any_, d := m.Any.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.ToolChoiceMemberAny
		diags.Append(fwflex.Expand(ctx, any_, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		result = &r
	case !m.Auto.IsNull():
		auto, d := m.Auto.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.ToolChoiceMemberAuto
		diags.Append(fwflex.Expand(ctx, auto, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		result = &r
	case !m.Tool.IsNull():
		tool, d := m.Tool.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.ToolChoiceMemberTool
		diags.Append(fwflex.Expand(ctx, tool, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		result = &r
	}

	return result, diags
}

func (m *toolChoiceModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch v := v.(type) {
	case awstypes.ToolChoiceMemberAny:
		var any_ anyToolChoiceModel
		diags.Append(fwflex.Flatten(ctx, v.Value, &any_)...)
		if diags.HasError() {
			return diags
		}

		m.Any = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &any_)
	case awstypes.ToolChoiceMemberAuto:
		var auto autoToolChoiceModel
		diags.Append(fwflex.Flatten(ctx, v.Value, &auto)...)
		if diags.HasError() {
			return diags
		}

		m.Auto = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &auto)
	case awstypes.ToolChoiceMemberTool:
		var tool specificToolChoiceModel
		diags.Append(fwflex.Flatten(ctx, v.Value, &tool)...)
		if diags.HasError() {
			return diags
		}

		m.Tool = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tool)
	}

	return diags
}

type anyToolChoiceModel struct{}

type autoToolChoiceModel struct{}

type specificToolChoiceModel struct {
	Name types.String `tfsdk:"name"`
}

type textPromptTemplateConfigurationModel struct {
	Text           types.String                                              `tfsdk:"text"`
	CachePoint     fwtypes.ListNestedObjectValueOf[cachePointModel]          `tfsdk:"cache_point"`
	InputVariables fwtypes.ListNestedObjectValueOf[promptInputVariableModel] `tfsdk:"input_variable"`
}

type cachePointModel struct {
	Type fwtypes.StringEnum[awstypes.CachePointType] `tfsdk:"type"`
}
