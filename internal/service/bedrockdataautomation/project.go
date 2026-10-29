// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockdataautomation

import (
	"context"
	"reflect"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockdataautomation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockdataautomation/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

// @FrameworkResource("aws_bedrockdataautomation_project", name="Project")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/bedrockdataautomation/types;awstypes;awstypes.DataAutomationProject")
// @Testing(hasNoPreExistingResource=true)
func newProjectResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &projectResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type projectResource struct {
	framework.ResourceWithModel[projectResourceModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *projectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	requiredState := schema.StringAttribute{
		CustomType: fwtypes.StringEnumType[awstypes.State](),
		Required:   true,
	}
	// stateBlock is used for every API structure whose only member is a State.
	stateBlock := func(required bool) schema.ListNestedBlock {
		block := schema.ListNestedBlock{
			CustomType: fwtypes.NewListNestedObjectTypeOf[stateConfigurationModel](ctx),
			Validators: []validator.List{
				listvalidator.SizeAtMost(1),
			},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					names.AttrState: requiredState,
				},
			},
		}
		if required {
			block.Validators = append(block.Validators, listvalidator.IsRequired())
		}
		return block
	}
	sensitiveDataConfigurationBlock := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[sensitiveDataConfigurationModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"detection_mode": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.SensitiveDataDetectionMode](),
					Required:   true,
				},
				"detection_scope": schema.SetAttribute{
					CustomType: fwtypes.SetOfStringEnumType[awstypes.SensitiveDataDetectionScopeType](),
					Optional:   true,
				},
			},
			Blocks: map[string]schema.Block{
				"pii_entities_configuration": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[piiEntitiesConfigurationModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"pii_entity_types": schema.SetAttribute{
								CustomType: fwtypes.SetOfStringEnumType[awstypes.PIIEntityType](),
								Optional:   true,
							},
							"redaction_mask_mode": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.PIIRedactionMaskMode](),
								Optional:   true,
							},
						},
					},
				},
			},
		},
	}
	blueprintItemObject := schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"blueprint_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			"blueprint_stage": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.BlueprintStage](),
				Optional:   true,
			},
			"blueprint_version": schema.StringAttribute{
				Optional: true,
			},
		},
	}

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreationTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 300),
				},
			},
			"last_modified_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9A-Za-z_-]+$`), "must contain only alphanumeric characters, hyphens, and underscores"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project_stage": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DataAutomationProjectStage](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DataAutomationProjectStatus](),
				Computed:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DataAutomationProjectType](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"custom_output_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[customOutputConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"blueprint": schema.ListNestedBlock{
							CustomType:   fwtypes.NewListNestedObjectTypeOf[blueprintItemModel](ctx),
							NestedObject: blueprintItemObject,
						},
						"document": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[documentCustomOutputConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"fallback_blueprint": schema.ListNestedBlock{
										CustomType:   fwtypes.NewListNestedObjectTypeOf[blueprintItemModel](ctx),
										NestedObject: blueprintItemObject,
									},
								},
							},
						},
					},
				},
			},
			"data_automation_library_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dataAutomationLibraryConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"library": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[dataAutomationLibraryItemModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"library_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrEncryptionConfiguration: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[encryptionConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"kms_encryption_context": schema.MapAttribute{
							CustomType: fwtypes.MapOfStringType,
							Optional:   true,
						},
						names.AttrKMSKeyID: schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 2048),
							},
						},
					},
				},
			},
			"override_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[overrideConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"audio": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[audioOverrideConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"language_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[audioLanguageConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"generative_output_language": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.AudioGenerativeOutputLanguage](),
													Optional:   true,
												},
												"identify_multiple_languages": schema.BoolAttribute{
													Optional: true,
												},
												"input_languages": schema.SetAttribute{
													CustomType: fwtypes.SetOfStringEnumType[awstypes.Language](),
													Optional:   true,
												},
											},
										},
									},
									"modality_processing":          stateBlock(false),
									"sensitive_data_configuration": sensitiveDataConfigurationBlock,
								},
							},
						},
						"document": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[documentOverrideConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"modality_processing":          stateBlock(false),
									"sensitive_data_configuration": sensitiveDataConfigurationBlock,
									"splitter":                     stateBlock(false),
								},
							},
						},
						"image": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[modalityOverrideConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"modality_processing":          stateBlock(false),
									"sensitive_data_configuration": sensitiveDataConfigurationBlock,
								},
							},
						},
						"modality_routing": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[modalityRoutingConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"jpeg": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.DesiredModality](),
										Optional:   true,
									},
									"mov": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.DesiredModality](),
										Optional:   true,
									},
									"mp4": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.DesiredModality](),
										Optional:   true,
									},
									"png": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.DesiredModality](),
										Optional:   true,
									},
								},
							},
						},
						"video": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[modalityOverrideConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"modality_processing":          stateBlock(false),
									"sensitive_data_configuration": sensitiveDataConfigurationBlock,
								},
							},
						},
					},
				},
			},
			"standard_output_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[standardOutputConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"audio": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[audioStandardOutputConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"extraction": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[audioStandardExtractionModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"category": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[audioExtractionCategoryModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
														listvalidator.IsRequired(),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrState: requiredState,
															"types": schema.SetAttribute{
																CustomType: fwtypes.SetOfStringEnumType[awstypes.AudioExtractionCategoryType](),
																Optional:   true,
															},
														},
														Blocks: map[string]schema.Block{
															"type_configuration": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[audioExtractionCategoryTypeConfigurationModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Blocks: map[string]schema.Block{
																		"transcript": schema.ListNestedBlock{
																			CustomType: fwtypes.NewListNestedObjectTypeOf[transcriptConfigurationModel](ctx),
																			Validators: []validator.List{
																				listvalidator.SizeAtMost(1),
																			},
																			NestedObject: schema.NestedBlockObject{
																				Blocks: map[string]schema.Block{
																					"channel_labeling": stateBlock(false),
																					"speaker_labeling": stateBlock(false),
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
									"generative_field": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[audioStandardGenerativeFieldModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrState: requiredState,
												"types": schema.SetAttribute{
													CustomType: fwtypes.SetOfStringEnumType[awstypes.AudioStandardGenerativeFieldType](),
													Optional:   true,
												},
											},
										},
									},
								},
							},
						},
						"document": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[documentStandardOutputConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"extraction": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[documentStandardExtractionModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"bounding_box": stateBlock(true),
												"granularity": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[documentExtractionGranularityModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
														listvalidator.IsRequired(),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"types": schema.SetAttribute{
																CustomType: fwtypes.SetOfStringEnumType[awstypes.DocumentExtractionGranularityType](),
																Optional:   true,
															},
														},
													},
												},
											},
										},
									},
									"generative_field": stateBlock(false),
									"output_format": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[documentOutputFormatModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"additional_file_format": stateBlock(true),
												"text_format": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[documentOutputTextFormatModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
														listvalidator.IsRequired(),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"types": schema.SetAttribute{
																CustomType: fwtypes.SetOfStringEnumType[awstypes.DocumentOutputTextFormatType](),
																Optional:   true,
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
						"image": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[imageStandardOutputConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"extraction": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[imageStandardExtractionModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"bounding_box": stateBlock(true),
												"category": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[imageExtractionCategoryModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
														listvalidator.IsRequired(),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrState: requiredState,
															"types": schema.SetAttribute{
																CustomType: fwtypes.SetOfStringEnumType[awstypes.ImageExtractionCategoryType](),
																Optional:   true,
															},
														},
													},
												},
											},
										},
									},
									"generative_field": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[imageStandardGenerativeFieldModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrState: requiredState,
												"types": schema.SetAttribute{
													CustomType: fwtypes.SetOfStringEnumType[awstypes.ImageStandardGenerativeFieldType](),
													Optional:   true,
												},
											},
										},
									},
								},
							},
						},
						"video": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[videoStandardOutputConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"extraction": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[videoStandardExtractionModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"bounding_box": stateBlock(true),
												"category": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[videoExtractionCategoryModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
														listvalidator.IsRequired(),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrState: requiredState,
															"types": schema.SetAttribute{
																CustomType: fwtypes.SetOfStringEnumType[awstypes.VideoExtractionCategoryType](),
																Optional:   true,
															},
														},
													},
												},
											},
										},
									},
									"generative_field": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[videoStandardGenerativeFieldModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrState: requiredState,
												"types": schema.SetAttribute{
													CustomType: fwtypes.SetOfStringEnumType[awstypes.VideoStandardGenerativeFieldType](),
													Optional:   true,
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

func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockDataAutomationClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input bedrockdataautomation.CreateDataAutomationProjectInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Project")))
	if resp.Diagnostics.HasError() {
		return
	}

	// The API requires a standard output configuration, but it can be empty.
	if input.StandardOutputConfiguration == nil {
		input.StandardOutputConfiguration = &awstypes.StandardOutputConfiguration{}
	}
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateDataAutomationProject(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	arn := aws.ToString(out.ProjectArn)
	// Save the ARN and Region so that a failed wait leaves a tainted resource rather than an orphan.
	plan.ARN = fwflex.StringValueToFramework(ctx, arn)
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.SetAttribute(ctx, path.Root(names.AttrARN), plan.ARN))
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.SetAttribute(ctx, path.Root(names.AttrRegion), plan.Region))
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := waitProjectCompleted(ctx, conn, arn, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, project, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockDataAutomationClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)
	out, err := findProjectByARN(ctx, conn, arn)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state projectResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockDataAutomationClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, plan.ARN)

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input bedrockdataautomation.UpdateDataAutomationProjectInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Project")))
		if resp.Diagnostics.HasError() {
			return
		}

		if input.StandardOutputConfiguration == nil {
			input.StandardOutputConfiguration = &awstypes.StandardOutputConfiguration{}
		}

		if _, err := conn.UpdateDataAutomationProject(ctx, &input); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}
	}

	// Waits for an update and also reads back the last modified time, which is unknown in the plan.
	project, err := waitProjectCompleted(ctx, conn, arn, r.UpdateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, project, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockDataAutomationClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)
	input := bedrockdataautomation.DeleteDataAutomationProjectInput{
		ProjectArn: aws.String(arn),
	}
	_, err := conn.DeleteDataAutomationProject(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	if _, err := waitProjectDeleted(ctx, conn, arn, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}
}

// flatten flattens the API object into the model.
//
// The API returns the encryption settings as top-level fields rather than as a
// nested object. It also fills in server-side defaults throughout the standard
// output and override configurations. Blocks can't be computed, so for those two
// configurations only the blocks present in data before flattening are kept.
func (r *projectResource) flatten(ctx context.Context, project *awstypes.DataAutomationProject, data *projectResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	apiObject := *project
	if v := apiObject.CustomOutputConfiguration; v != nil && len(v.Blueprints) == 0 && v.Document == nil {
		apiObject.CustomOutputConfiguration = nil
	}
	if v := apiObject.DataAutomationLibraryConfiguration; v != nil && len(v.Libraries) == 0 {
		apiObject.DataAutomationLibraryConfiguration = nil
	}

	priorOverride, priorStandard := data.OverrideConfiguration, data.StandardOutputConfiguration

	smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, &apiObject, data, fwflex.WithFieldNamePrefix("Project")))
	if diags.HasError() {
		return diags
	}

	var d diag.Diagnostics
	data.OverrideConfiguration, d = keepConfigured(ctx, priorOverride, data.OverrideConfiguration)
	smerr.AddEnrich(ctx, &diags, d)
	data.StandardOutputConfiguration, d = keepConfigured(ctx, priorStandard, data.StandardOutputConfiguration)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return diags
	}

	if project.KmsKeyId == nil {
		data.EncryptionConfiguration = fwtypes.NewListNestedObjectValueOfNull[encryptionConfigurationModel](ctx)
		return diags
	}

	encryptionConfiguration := awstypes.EncryptionConfiguration{
		KmsEncryptionContext: project.KmsEncryptionContext,
		KmsKeyId:             project.KmsKeyId,
	}
	var model encryptionConfigurationModel
	smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, &encryptionConfiguration, &model))
	if diags.HasError() {
		return diags
	}

	data.EncryptionConfiguration, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &model)
	smerr.AddEnrich(ctx, &diags, d)

	return diags
}

// keepConfigured returns current with every block and attribute that is absent
// from prior set to null. Blocks with more than one element are kept as they are.
func keepConfigured[T any](ctx context.Context, prior, current fwtypes.ListNestedObjectValueOf[T]) (fwtypes.ListNestedObjectValueOf[T], diag.Diagnostics) {
	v, diags := keepConfiguredValue(ctx, prior, current)
	if diags.HasError() {
		return current, diags
	}

	return v.(fwtypes.ListNestedObjectValueOf[T]), diags
}

func keepConfiguredValue(ctx context.Context, prior, current fwtypes.NestedObjectCollectionValue) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	typ, ok := current.Type(ctx).(fwtypes.NestedObjectCollectionType)
	if !ok {
		return current, diags
	}

	type collection interface {
		Length(basetypes.CollectionLengthOptions) int
	}
	priorElements, currentElements := 0, 0
	if v, ok := prior.(collection); ok {
		priorElements = v.Length(fwtypes.CollectionLengthUnhandledAsZero)
	}
	if v, ok := current.(collection); ok {
		currentElements = v.Length(fwtypes.CollectionLengthUnhandledAsZero)
	}

	if priorElements == 0 {
		return typ.NullValue(ctx)
	}
	if priorElements > 1 || currentElements != 1 {
		return current, diags
	}

	priorPtr, d := prior.ToObjectPtr(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	currentPtr, d := current.ToObjectPtr(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return current, diags
	}

	priorStruct, currentStruct := reflect.ValueOf(priorPtr).Elem(), reflect.ValueOf(currentPtr).Elem()
	for i := range currentStruct.NumField() {
		c, ok := currentStruct.Field(i).Interface().(fwtypes.NestedObjectCollectionValue)
		if !ok {
			if p, ok := priorStruct.Field(i).Interface().(attr.Value); ok && p.IsNull() {
				currentStruct.Field(i).Set(priorStruct.Field(i))
			}
			continue
		}
		p := priorStruct.Field(i).Interface().(fwtypes.NestedObjectCollectionValue)

		v, d := keepConfiguredValue(ctx, p, c)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return current, diags
		}
		currentStruct.Field(i).Set(reflect.ValueOf(v))
	}

	return typ.ValueFromObjectPtr(ctx, currentPtr)
}

func findProjectByARN(ctx context.Context, conn *bedrockdataautomation.Client, arn string) (*awstypes.DataAutomationProject, error) {
	input := bedrockdataautomation.GetDataAutomationProjectInput{
		ProjectArn: aws.String(arn),
	}

	out, err := conn.GetDataAutomationProject(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil || out.Project == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.Project, nil
}

func statusProject(conn *bedrockdataautomation.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findProjectByARN(ctx, conn, arn)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func waitProjectCompleted(ctx context.Context, conn *bedrockdataautomation.Client, arn string, timeout time.Duration) (*awstypes.DataAutomationProject, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DataAutomationProjectStatusInProgress),
		Target:  enum.Slice(awstypes.DataAutomationProjectStatusCompleted),
		Refresh: statusProject(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.DataAutomationProject); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitProjectDeleted(ctx context.Context, conn *bedrockdataautomation.Client, arn string, timeout time.Duration) (*awstypes.DataAutomationProject, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DataAutomationProjectStatusCompleted, awstypes.DataAutomationProjectStatusInProgress, awstypes.DataAutomationProjectStatusFailed),
		Target:  []string{},
		Refresh: statusProject(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.DataAutomationProject); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

type projectResourceModel struct {
	framework.WithRegionModel
	ARN                                types.String                                                             `tfsdk:"arn"`
	CreationTime                       timetypes.RFC3339                                                        `tfsdk:"creation_time"`
	CustomOutputConfiguration          fwtypes.ListNestedObjectValueOf[customOutputConfigurationModel]          `tfsdk:"custom_output_configuration"`
	DataAutomationLibraryConfiguration fwtypes.ListNestedObjectValueOf[dataAutomationLibraryConfigurationModel] `tfsdk:"data_automation_library_configuration"`
	Description                        types.String                                                             `tfsdk:"description"`
	EncryptionConfiguration            fwtypes.ListNestedObjectValueOf[encryptionConfigurationModel]            `tfsdk:"encryption_configuration"`
	LastModifiedTime                   timetypes.RFC3339                                                        `tfsdk:"last_modified_time"`
	Name                               types.String                                                             `tfsdk:"name"`
	OverrideConfiguration              fwtypes.ListNestedObjectValueOf[overrideConfigurationModel]              `tfsdk:"override_configuration"`
	ProjectStage                       fwtypes.StringEnum[awstypes.DataAutomationProjectStage]                  `tfsdk:"project_stage"`
	StandardOutputConfiguration        fwtypes.ListNestedObjectValueOf[standardOutputConfigurationModel]        `tfsdk:"standard_output_configuration"`
	Status                             fwtypes.StringEnum[awstypes.DataAutomationProjectStatus]                 `tfsdk:"status"`
	Tags                               tftags.Map                                                               `tfsdk:"tags"`
	TagsAll                            tftags.Map                                                               `tfsdk:"tags_all"`
	Timeouts                           timeouts.Value                                                           `tfsdk:"timeouts"`
	Type                               fwtypes.StringEnum[awstypes.DataAutomationProjectType]                   `tfsdk:"type"`
}

// stateConfigurationModel models every API structure whose only member is a State.
type stateConfigurationModel struct {
	State fwtypes.StringEnum[awstypes.State] `tfsdk:"state"`
}

type blueprintItemModel struct {
	BlueprintARN     fwtypes.ARN                                 `tfsdk:"blueprint_arn"`
	BlueprintStage   fwtypes.StringEnum[awstypes.BlueprintStage] `tfsdk:"blueprint_stage"`
	BlueprintVersion types.String                                `tfsdk:"blueprint_version"`
}

type customOutputConfigurationModel struct {
	Blueprints fwtypes.ListNestedObjectValueOf[blueprintItemModel]                     `tfsdk:"blueprint"`
	Document   fwtypes.ListNestedObjectValueOf[documentCustomOutputConfigurationModel] `tfsdk:"document"`
}

type documentCustomOutputConfigurationModel struct {
	FallbackBlueprints fwtypes.ListNestedObjectValueOf[blueprintItemModel] `tfsdk:"fallback_blueprint"`
}

type dataAutomationLibraryConfigurationModel struct {
	Libraries fwtypes.ListNestedObjectValueOf[dataAutomationLibraryItemModel] `tfsdk:"library"`
}

type dataAutomationLibraryItemModel struct {
	LibraryARN fwtypes.ARN `tfsdk:"library_arn"`
}

type overrideConfigurationModel struct {
	Audio           fwtypes.ListNestedObjectValueOf[audioOverrideConfigurationModel]    `tfsdk:"audio"`
	Document        fwtypes.ListNestedObjectValueOf[documentOverrideConfigurationModel] `tfsdk:"document"`
	Image           fwtypes.ListNestedObjectValueOf[modalityOverrideConfigurationModel] `tfsdk:"image"`
	ModalityRouting fwtypes.ListNestedObjectValueOf[modalityRoutingConfigurationModel]  `tfsdk:"modality_routing"`
	Video           fwtypes.ListNestedObjectValueOf[modalityOverrideConfigurationModel] `tfsdk:"video"`
}

type audioOverrideConfigurationModel struct {
	LanguageConfiguration      fwtypes.ListNestedObjectValueOf[audioLanguageConfigurationModel] `tfsdk:"language_configuration"`
	ModalityProcessing         fwtypes.ListNestedObjectValueOf[stateConfigurationModel]         `tfsdk:"modality_processing"`
	SensitiveDataConfiguration fwtypes.ListNestedObjectValueOf[sensitiveDataConfigurationModel] `tfsdk:"sensitive_data_configuration"`
}

type audioLanguageConfigurationModel struct {
	GenerativeOutputLanguage  fwtypes.StringEnum[awstypes.AudioGenerativeOutputLanguage] `tfsdk:"generative_output_language"`
	IdentifyMultipleLanguages types.Bool                                                 `tfsdk:"identify_multiple_languages"`
	InputLanguages            fwtypes.SetOfStringEnum[awstypes.Language]                 `tfsdk:"input_languages"`
}

type documentOverrideConfigurationModel struct {
	ModalityProcessing         fwtypes.ListNestedObjectValueOf[stateConfigurationModel]         `tfsdk:"modality_processing"`
	SensitiveDataConfiguration fwtypes.ListNestedObjectValueOf[sensitiveDataConfigurationModel] `tfsdk:"sensitive_data_configuration"`
	Splitter                   fwtypes.ListNestedObjectValueOf[stateConfigurationModel]         `tfsdk:"splitter"`
}

// modalityOverrideConfigurationModel models the image and video override configurations.
type modalityOverrideConfigurationModel struct {
	ModalityProcessing         fwtypes.ListNestedObjectValueOf[stateConfigurationModel]         `tfsdk:"modality_processing"`
	SensitiveDataConfiguration fwtypes.ListNestedObjectValueOf[sensitiveDataConfigurationModel] `tfsdk:"sensitive_data_configuration"`
}

type modalityRoutingConfigurationModel struct {
	JPEG fwtypes.StringEnum[awstypes.DesiredModality] `tfsdk:"jpeg"`
	MOV  fwtypes.StringEnum[awstypes.DesiredModality] `tfsdk:"mov"`
	MP4  fwtypes.StringEnum[awstypes.DesiredModality] `tfsdk:"mp4"`
	PNG  fwtypes.StringEnum[awstypes.DesiredModality] `tfsdk:"png"`
}

type sensitiveDataConfigurationModel struct {
	DetectionMode            fwtypes.StringEnum[awstypes.SensitiveDataDetectionMode]           `tfsdk:"detection_mode"`
	DetectionScope           fwtypes.SetOfStringEnum[awstypes.SensitiveDataDetectionScopeType] `tfsdk:"detection_scope"`
	PIIEntitiesConfiguration fwtypes.ListNestedObjectValueOf[piiEntitiesConfigurationModel]    `tfsdk:"pii_entities_configuration"`
}

type piiEntitiesConfigurationModel struct {
	PIIEntityTypes    fwtypes.SetOfStringEnum[awstypes.PIIEntityType]   `tfsdk:"pii_entity_types"`
	RedactionMaskMode fwtypes.StringEnum[awstypes.PIIRedactionMaskMode] `tfsdk:"redaction_mask_mode"`
}

type standardOutputConfigurationModel struct {
	Audio    fwtypes.ListNestedObjectValueOf[audioStandardOutputConfigurationModel]    `tfsdk:"audio"`
	Document fwtypes.ListNestedObjectValueOf[documentStandardOutputConfigurationModel] `tfsdk:"document"`
	Image    fwtypes.ListNestedObjectValueOf[imageStandardOutputConfigurationModel]    `tfsdk:"image"`
	Video    fwtypes.ListNestedObjectValueOf[videoStandardOutputConfigurationModel]    `tfsdk:"video"`
}

type audioStandardOutputConfigurationModel struct {
	Extraction      fwtypes.ListNestedObjectValueOf[audioStandardExtractionModel]      `tfsdk:"extraction"`
	GenerativeField fwtypes.ListNestedObjectValueOf[audioStandardGenerativeFieldModel] `tfsdk:"generative_field"`
}

type audioStandardExtractionModel struct {
	Category fwtypes.ListNestedObjectValueOf[audioExtractionCategoryModel] `tfsdk:"category"`
}

type audioExtractionCategoryModel struct {
	State             fwtypes.StringEnum[awstypes.State]                                             `tfsdk:"state"`
	TypeConfiguration fwtypes.ListNestedObjectValueOf[audioExtractionCategoryTypeConfigurationModel] `tfsdk:"type_configuration"`
	Types             fwtypes.SetOfStringEnum[awstypes.AudioExtractionCategoryType]                  `tfsdk:"types"`
}

type audioExtractionCategoryTypeConfigurationModel struct {
	Transcript fwtypes.ListNestedObjectValueOf[transcriptConfigurationModel] `tfsdk:"transcript"`
}

type transcriptConfigurationModel struct {
	ChannelLabeling fwtypes.ListNestedObjectValueOf[stateConfigurationModel] `tfsdk:"channel_labeling"`
	SpeakerLabeling fwtypes.ListNestedObjectValueOf[stateConfigurationModel] `tfsdk:"speaker_labeling"`
}

type audioStandardGenerativeFieldModel struct {
	State fwtypes.StringEnum[awstypes.State]                                 `tfsdk:"state"`
	Types fwtypes.SetOfStringEnum[awstypes.AudioStandardGenerativeFieldType] `tfsdk:"types"`
}

type documentStandardOutputConfigurationModel struct {
	Extraction      fwtypes.ListNestedObjectValueOf[documentStandardExtractionModel] `tfsdk:"extraction"`
	GenerativeField fwtypes.ListNestedObjectValueOf[stateConfigurationModel]         `tfsdk:"generative_field"`
	OutputFormat    fwtypes.ListNestedObjectValueOf[documentOutputFormatModel]       `tfsdk:"output_format"`
}

type documentStandardExtractionModel struct {
	BoundingBox fwtypes.ListNestedObjectValueOf[stateConfigurationModel]            `tfsdk:"bounding_box"`
	Granularity fwtypes.ListNestedObjectValueOf[documentExtractionGranularityModel] `tfsdk:"granularity"`
}

type documentExtractionGranularityModel struct {
	Types fwtypes.SetOfStringEnum[awstypes.DocumentExtractionGranularityType] `tfsdk:"types"`
}

type documentOutputFormatModel struct {
	AdditionalFileFormat fwtypes.ListNestedObjectValueOf[stateConfigurationModel]       `tfsdk:"additional_file_format"`
	TextFormat           fwtypes.ListNestedObjectValueOf[documentOutputTextFormatModel] `tfsdk:"text_format"`
}

type documentOutputTextFormatModel struct {
	Types fwtypes.SetOfStringEnum[awstypes.DocumentOutputTextFormatType] `tfsdk:"types"`
}

type imageStandardOutputConfigurationModel struct {
	Extraction      fwtypes.ListNestedObjectValueOf[imageStandardExtractionModel]      `tfsdk:"extraction"`
	GenerativeField fwtypes.ListNestedObjectValueOf[imageStandardGenerativeFieldModel] `tfsdk:"generative_field"`
}

type imageStandardExtractionModel struct {
	BoundingBox fwtypes.ListNestedObjectValueOf[stateConfigurationModel]      `tfsdk:"bounding_box"`
	Category    fwtypes.ListNestedObjectValueOf[imageExtractionCategoryModel] `tfsdk:"category"`
}

type imageExtractionCategoryModel struct {
	State fwtypes.StringEnum[awstypes.State]                            `tfsdk:"state"`
	Types fwtypes.SetOfStringEnum[awstypes.ImageExtractionCategoryType] `tfsdk:"types"`
}

type imageStandardGenerativeFieldModel struct {
	State fwtypes.StringEnum[awstypes.State]                                 `tfsdk:"state"`
	Types fwtypes.SetOfStringEnum[awstypes.ImageStandardGenerativeFieldType] `tfsdk:"types"`
}

type videoStandardOutputConfigurationModel struct {
	Extraction      fwtypes.ListNestedObjectValueOf[videoStandardExtractionModel]      `tfsdk:"extraction"`
	GenerativeField fwtypes.ListNestedObjectValueOf[videoStandardGenerativeFieldModel] `tfsdk:"generative_field"`
}

type videoStandardExtractionModel struct {
	BoundingBox fwtypes.ListNestedObjectValueOf[stateConfigurationModel]      `tfsdk:"bounding_box"`
	Category    fwtypes.ListNestedObjectValueOf[videoExtractionCategoryModel] `tfsdk:"category"`
}

type videoExtractionCategoryModel struct {
	State fwtypes.StringEnum[awstypes.State]                            `tfsdk:"state"`
	Types fwtypes.SetOfStringEnum[awstypes.VideoExtractionCategoryType] `tfsdk:"types"`
}

type videoStandardGenerativeFieldModel struct {
	State fwtypes.StringEnum[awstypes.State]                                 `tfsdk:"state"`
	Types fwtypes.SetOfStringEnum[awstypes.VideoStandardGenerativeFieldType] `tfsdk:"types"`
}
