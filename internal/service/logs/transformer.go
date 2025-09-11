// Copyright (c) HashiCorp, Inc.
// Copyright 2025 Twilio Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"errors"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_cloudwatch_log_transformer", name="Transformer")
func newResourceTransformer(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceTransformer{}

	return r, nil
}

const (
	ResNameTransformer = "Transformer"
)

type resourceTransformer struct {
	framework.ResourceWithModel[resourceTransformerModel]
}

func (r *resourceTransformer) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"log_group_identifier": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 2048),
					stringvalidator.RegexMatches(regexache.MustCompile(`[\w#+=/:,.@-]*`), "must be a valid log group name or ARN"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"transformer_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[transformerConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeBetween(1, 20),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"add_keys": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[addKeysModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"entries": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[addKeysEntryModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeBetween(1, 5),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrKey: schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 128),
													},
												},
												"overwrite_if_exists": schema.BoolAttribute{
													Optional: true,
													Computed: true,
												},
												names.AttrValue: schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 256),
													},
												},
											},
										},
									},
								},
							},
						},
						"copy_value": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[copyValueModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"entries": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[copyValueEntryModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeBetween(1, 5),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"overwrite_if_exists": schema.BoolAttribute{
													Optional: true,
													Computed: true,
												},
												names.AttrSource: schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 128),
													},
												},
												names.AttrTarget: schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 128),
													},
												},
											},
										},
									},
								},
							},
						},
						"csv": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[csvModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(5),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"columns": schema.ListAttribute{
										Optional:    true,
										Computed:    true,
										CustomType:  fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Validators: []validator.List{
											listvalidator.SizeAtMost(100),
											listvalidator.ValueStringsAre(stringvalidator.LengthBetween(1, 128)),
										},
									},
									"delimiter": schema.StringAttribute{
										Optional: true,
										Computed: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 2),
										},
									},
									"quote_character": schema.StringAttribute{
										Optional: true,
										Computed: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 1),
										},
									},
									names.AttrSource: schema.StringAttribute{
										Optional: true,
										Computed: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 128),
										},
									},
								},
							},
						},
						"date_time_converter": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[dateTimeConverterModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(20),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"locale": schema.StringAttribute{
										Optional: true,
										Computed: true,
										Validators: []validator.String{
											stringvalidator.LengthAtLeast(1),
										},
									},
									"match_patterns": schema.ListAttribute{
										Required:    true,
										CustomType:  fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Validators: []validator.List{
											listvalidator.SizeBetween(1, 5),
											listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
										},
									},
									names.AttrSource: schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 128),
										},
									},
									"source_timezone": schema.StringAttribute{
										Optional: true,
										Computed: true,
										Validators: []validator.String{
											stringvalidator.LengthAtLeast(1),
										},
									},
									names.AttrTarget: schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 128),
										},
									},
									"target_format": schema.StringAttribute{
										Optional: true,
										Computed: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 64),
										},
									},
									"target_timezone": schema.StringAttribute{
										Optional: true,
										Computed: true,
										Validators: []validator.String{
											stringvalidator.LengthAtLeast(1),
										},
									},
								},
							},
						},
						"delete_keys": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[deleteKeysModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(20),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"with_keys": schema.ListAttribute{
										Required:    true,
										CustomType:  fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Validators: []validator.List{
											listvalidator.SizeBetween(1, 5),
											listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
										},
									},
								},
							},
						},
						"grok": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[grokModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"match": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 512),
										},
									},
									names.AttrSource: schema.StringAttribute{
										Optional: true,
										Computed: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 128),
										},
									},
								},
							},
						},
						"list_to_map": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[listToMapModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(20),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"flatten": schema.BoolAttribute{
										Optional: true,
										Computed: true,
									},
									"flattened_element": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.FlattenedElement](),
										Optional:   true,
										Computed:   true,
									},
									names.AttrKey: schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 128),
										},
									},
									names.AttrSource: schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 128),
										},
									},
									names.AttrTarget: schema.StringAttribute{
										Optional: true,
										Computed: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 128),
										},
									},
									"value_key": schema.StringAttribute{
										Optional: true,
										Computed: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 128),
										},
									},
								},
							},
						},
						"lower_case_string": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[lowerCaseStringModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(20),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"with_keys": schema.ListAttribute{
										Required:    true,
										CustomType:  fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Validators: []validator.List{
											listvalidator.SizeBetween(1, 10),
											listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
										},
									},
								},
							},
						},
						"move_keys": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[moveKeysModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(20),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"entries": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[moveKeysEntryModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeBetween(1, 5),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"overwrite_if_exists": schema.BoolAttribute{
													Optional: true,
													Computed: true,
												},
												names.AttrSource: schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 128),
													},
												},
												names.AttrTarget: schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 128),
													},
												},
											},
										},
									},
								},
							},
						},
						"parse_cloudfront": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[parseCloudfrontModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrSource: schema.StringAttribute{
										Optional: true,
										Computed: true,
										Validators: []validator.String{
											stringvalidator.RegexMatches(regexache.MustCompile(`^@message$`), "must be '@message'"),
										},
									},
								},
							},
						},
						"parse_json": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[parseJSONModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(5),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrDestination: schema.StringAttribute{
										Optional: true,
										Computed: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 128),
										},
									},
									names.AttrSource: schema.StringAttribute{
										Optional: true,
										Computed: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 128),
										},
									},
								},
							},
						},
						"parse_key_value": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[parseKeyValueModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(5),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrDestination: schema.StringAttribute{
										Optional: true,
										Computed: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 128),
										},
									},
									"field_delimiter": schema.StringAttribute{
										Optional: true,
										Computed: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 128),
										},
									},
									"key_prefix": schema.StringAttribute{
										Optional: true,
										Computed: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 128),
										},
									},
									"key_value_delimiter": schema.StringAttribute{
										Optional: true,
										Computed: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 128),
										},
									},
									"non_match_value": schema.StringAttribute{
										Optional: true,
										Computed: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 128),
										},
									},
									"overwrite_if_exists": schema.BoolAttribute{
										Optional: true,
										Computed: true,
									},
									names.AttrSource: schema.StringAttribute{
										Optional: true,
										Computed: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 128),
										},
									},
								},
							},
						},
						"parse_postgres": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[parsePostgresModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrSource: schema.StringAttribute{
										Optional: true,
										Computed: true,
										Validators: []validator.String{
											stringvalidator.RegexMatches(regexache.MustCompile(`^@message$`), "must be '@message'"),
										},
									},
								},
							},
						},
						"parse_route53": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[parseRoute53Model](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrSource: schema.StringAttribute{
										Optional: true,
										Computed: true,
										Validators: []validator.String{
											stringvalidator.RegexMatches(regexache.MustCompile(`^@message$`), "must be '@message'"),
										},
									},
								},
							},
						},
						"parse_to_ocsf": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[parseToOCSFModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"event_source": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.EventSource](),
										Required:   true,
									},
									"ocsf_version": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.OCSFVersion](),
										Required:   true,
									},
									names.AttrSource: schema.StringAttribute{
										Optional: true,
										Computed: true,
										Validators: []validator.String{
											stringvalidator.RegexMatches(regexache.MustCompile(`^@message$`), "must be '@message'"),
										},
									},
								},
							},
						},
						"parse_vpc": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[parseVPCModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrSource: schema.StringAttribute{
										Optional: true,
										Computed: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 128),
										},
									},
								},
							},
						},
						"parse_waf": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[parseWAFModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrSource: schema.StringAttribute{
										Optional: true,
										Computed: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 128),
										},
									},
								},
							},
						},
						"rename_keys": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[renameKeysModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(20),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"entries": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[renameKeysEntryModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeBetween(1, 5),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrKey: schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 128),
													},
												},
												"overwrite_if_exists": schema.BoolAttribute{
													Optional: true,
													Computed: true,
												},
												"rename_to": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 128),
													},
												},
											},
										},
									},
								},
							},
						},
						"split_string": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[splitStringModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(20),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"entries": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[splitStringEntryModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeBetween(1, 10),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"delimiter": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 128),
													},
												},
												names.AttrSource: schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 128),
													},
												},
											},
										},
									},
								},
							},
						},
						"substitute_string": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[substituteStringModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(20),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"entries": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[substituteStringEntryModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeBetween(1, 10),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"from": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 128),
													},
												},
												names.AttrSource: schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 128),
													},
												},
												"to": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 128),
													},
												},
											},
										},
									},
								},
							},
						},
						"trim_string": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[trimStringModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(20),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"with_keys": schema.ListAttribute{
										Required:    true,
										CustomType:  fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Validators: []validator.List{
											listvalidator.SizeBetween(1, 10),
											listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
										},
									},
								},
							},
						},
						"type_converter": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[typeConverterModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(20),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"entries": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[typeConverterEntryModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeBetween(1, 5),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrKey: schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 128),
													},
												},
												names.AttrType: schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.Type](),
													Required:   true,
												},
											},
										},
									},
								},
							},
						},
						"upper_case_string": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[upperCaseStringModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(20),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"with_keys": schema.ListAttribute{
										Required:    true,
										CustomType:  fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Validators: []validator.List{
											listvalidator.SizeBetween(1, 10),
											listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
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

func (r *resourceTransformer) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().LogsClient(ctx)

	var plan resourceTransformerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input cloudwatchlogs.PutTransformerInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.PutTransformer(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Logs, create.ErrActionCreating, ResNameTransformer, plan.LogGroupIdentifier.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Logs, create.ErrActionCreating, ResNameTransformer, plan.LogGroupIdentifier.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	transformer, err := findTransformerByLogGroupIdentifier(ctx, conn, plan.LogGroupIdentifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Logs, create.ErrActionReading, ResNameTransformer, plan.LogGroupIdentifier.String(), err),
			err.Error(),
		)
		return
	}

	// Set values for unknowns
	resp.Diagnostics.Append(flex.Flatten(ctx, transformer, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceTransformer) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LogsClient(ctx)

	var state resourceTransformerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findTransformerByLogGroupIdentifier(ctx, conn, state.LogGroupIdentifier.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Logs, create.ErrActionReading, ResNameTransformer, state.LogGroupIdentifier.String(), err),
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

func (r *resourceTransformer) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().LogsClient(ctx)

	var plan, state resourceTransformerModel
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
		var input cloudwatchlogs.PutTransformerInput
		resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.PutTransformer(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Logs, create.ErrActionUpdating, ResNameTransformer, plan.LogGroupIdentifier.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Logs, create.ErrActionUpdating, ResNameTransformer, plan.LogGroupIdentifier.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
	}

	transformer, err := findTransformerByLogGroupIdentifier(ctx, conn, plan.LogGroupIdentifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Logs, create.ErrActionReading, ResNameTransformer, plan.LogGroupIdentifier.String(), err),
			err.Error(),
		)
		return
	}

	// Set values for unknowns
	resp.Diagnostics.Append(flex.Flatten(ctx, transformer, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceTransformer) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().LogsClient(ctx)

	var state resourceTransformerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := cloudwatchlogs.DeleteTransformerInput{
		LogGroupIdentifier: state.LogGroupIdentifier.ValueStringPointer(),
	}

	_, err := conn.DeleteTransformer(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Logs, create.ErrActionDeleting, ResNameTransformer, state.LogGroupIdentifier.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceTransformer) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("log_group_identifier"), req, resp)
}

func findTransformerByLogGroupIdentifier(ctx context.Context, conn *cloudwatchlogs.Client, logGroupIdentifier string) (*cloudwatchlogs.GetTransformerOutput, error) {
	input := cloudwatchlogs.GetTransformerInput{
		LogGroupIdentifier: aws.String(logGroupIdentifier),
	}

	out, err := conn.GetTransformer(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}

	return out, nil
}

type resourceTransformerModel struct {
	framework.WithRegionModel
	LogGroupIdentifier types.String                                            `tfsdk:"log_group_identifier"`
	TransformerConfig  fwtypes.ListNestedObjectValueOf[transformerConfigModel] `tfsdk:"transformer_config"`
}

type transformerConfigModel struct {
	AddKeys           fwtypes.ListNestedObjectValueOf[addKeysModel]           `tfsdk:"add_keys"`
	CopyValue         fwtypes.ListNestedObjectValueOf[copyValueModel]         `tfsdk:"copy_value"`
	CSV               fwtypes.ListNestedObjectValueOf[csvModel]               `tfsdk:"csv"`
	DateTimeConverter fwtypes.ListNestedObjectValueOf[dateTimeConverterModel] `tfsdk:"date_time_converter"`
	DeleteKeys        fwtypes.ListNestedObjectValueOf[deleteKeysModel]        `tfsdk:"delete_keys"`
	Grok              fwtypes.ListNestedObjectValueOf[grokModel]              `tfsdk:"grok"`
	ListToMap         fwtypes.ListNestedObjectValueOf[listToMapModel]         `tfsdk:"list_to_map"`
	LowerCaseString   fwtypes.ListNestedObjectValueOf[lowerCaseStringModel]   `tfsdk:"lower_case_string"`
	MoveKeys          fwtypes.ListNestedObjectValueOf[moveKeysModel]          `tfsdk:"move_keys"`
	ParseCloudfront   fwtypes.ListNestedObjectValueOf[parseCloudfrontModel]   `tfsdk:"parse_cloudfront"`
	ParseJSON         fwtypes.ListNestedObjectValueOf[parseJSONModel]         `tfsdk:"parse_json"`
	ParseKeyValue     fwtypes.ListNestedObjectValueOf[parseKeyValueModel]     `tfsdk:"parse_key_value"`
	ParsePostgres     fwtypes.ListNestedObjectValueOf[parsePostgresModel]     `tfsdk:"parse_postgres"`
	ParseRoute53      fwtypes.ListNestedObjectValueOf[parseRoute53Model]      `tfsdk:"parse_route53"`
	ParseToOCSF       fwtypes.ListNestedObjectValueOf[parseToOCSFModel]       `tfsdk:"parse_to_ocsf"`
	ParseVPC          fwtypes.ListNestedObjectValueOf[parseVPCModel]          `tfsdk:"parse_vpc"`
	ParseWAF          fwtypes.ListNestedObjectValueOf[parseWAFModel]          `tfsdk:"parse_waf"`
	RenameKeys        fwtypes.ListNestedObjectValueOf[renameKeysModel]        `tfsdk:"rename_keys"`
	SplitString       fwtypes.ListNestedObjectValueOf[splitStringModel]       `tfsdk:"split_string"`
	SubstituteString  fwtypes.ListNestedObjectValueOf[substituteStringModel]  `tfsdk:"substitute_string"`
	TrimString        fwtypes.ListNestedObjectValueOf[trimStringModel]        `tfsdk:"trim_string"`
	TypeConverter     fwtypes.ListNestedObjectValueOf[typeConverterModel]     `tfsdk:"type_converter"`
	UpperCaseString   fwtypes.ListNestedObjectValueOf[upperCaseStringModel]   `tfsdk:"upper_case_string"`
}

type addKeysModel struct {
	Entries fwtypes.ListNestedObjectValueOf[addKeysEntryModel] `tfsdk:"entries"`
}

type addKeysEntryModel struct {
	Key               types.String `tfsdk:"key"`
	OverwriteIfExists types.Bool   `tfsdk:"overwrite_if_exists"`
	Value             types.String `tfsdk:"value"`
}

type copyValueModel struct {
	Entries fwtypes.ListNestedObjectValueOf[copyValueEntryModel] `tfsdk:"entries"`
}

type copyValueEntryModel struct {
	OverwriteIfExists types.Bool   `tfsdk:"overwrite_if_exists"`
	Source            types.String `tfsdk:"source"`
	Target            types.String `tfsdk:"target"`
}

type csvModel struct {
	Columns        fwtypes.ListOfString `tfsdk:"columns"`
	Delimiter      types.String         `tfsdk:"delimiter"`
	QuoteCharacter types.String         `tfsdk:"quote_character"`
	Source         types.String         `tfsdk:"source"`
}

type dateTimeConverterModel struct {
	Locale         types.String         `tfsdk:"locale"`
	MatchPatterns  fwtypes.ListOfString `tfsdk:"match_patterns"`
	Source         types.String         `tfsdk:"source"`
	SourceTimezone types.String         `tfsdk:"source_timezone"`
	Target         types.String         `tfsdk:"target"`
	TargetFormat   types.String         `tfsdk:"target_format"`
	TargetTimezone types.String         `tfsdk:"target_timezone"`
}

type deleteKeysModel struct {
	WithKeys fwtypes.ListOfString `tfsdk:"with_keys"`
}

type grokModel struct {
	Match  types.String `tfsdk:"match"`
	Source types.String `tfsdk:"source"`
}

type listToMapModel struct {
	Flatten          types.Bool                                    `tfsdk:"flatten"`
	FlattenedElement fwtypes.StringEnum[awstypes.FlattenedElement] `tfsdk:"flattened_element"`
	Key              types.String                                  `tfsdk:"key"`
	Source           types.String                                  `tfsdk:"source"`
	Target           types.String                                  `tfsdk:"target"`
	ValueKey         types.String                                  `tfsdk:"value_key"`
}

type lowerCaseStringModel struct {
	WithKeys fwtypes.ListOfString `tfsdk:"with_keys"`
}

type moveKeysModel struct {
	Entries fwtypes.ListNestedObjectValueOf[moveKeysEntryModel] `tfsdk:"entries"`
}

type moveKeysEntryModel struct {
	OverwriteIfExists types.Bool   `tfsdk:"overwrite_if_exists"`
	Source            types.String `tfsdk:"source"`
	Target            types.String `tfsdk:"target"`
}

type parseCloudfrontModel struct {
	Source types.String `tfsdk:"source"`
}

type parseJSONModel struct {
	Destination types.String `tfsdk:"destination"`
	Source      types.String `tfsdk:"source"`
}

type parseKeyValueModel struct {
	Destination       types.String `tfsdk:"destination"`
	FieldDelimiter    types.String `tfsdk:"field_delimiter"`
	KeyPrefix         types.String `tfsdk:"key_prefix"`
	KeyValueDelimiter types.String `tfsdk:"key_value_delimiter"`
	NonMatchValue     types.String `tfsdk:"non_match_value"`
	OverwriteIfExists types.Bool   `tfsdk:"overwrite_if_exists"`
	Source            types.String `tfsdk:"source"`
}

type parsePostgresModel struct {
	Source types.String `tfsdk:"source"`
}

type parseRoute53Model struct {
	Source types.String `tfsdk:"source"`
}

type parseToOCSFModel struct {
	EventSource fwtypes.StringEnum[awstypes.EventSource] `tfsdk:"event_source"`
	OCSFVersion fwtypes.StringEnum[awstypes.OCSFVersion] `tfsdk:"ocsf_version"`
	Source      types.String                             `tfsdk:"source"`
}

type parseVPCModel struct {
	Source types.String `tfsdk:"source"`
}

type parseWAFModel struct {
	Source types.String `tfsdk:"source"`
}

type renameKeysModel struct {
	Entries fwtypes.ListNestedObjectValueOf[renameKeysEntryModel] `tfsdk:"entries"`
}

type renameKeysEntryModel struct {
	Key               types.String `tfsdk:"key"`
	OverwriteIfExists types.Bool   `tfsdk:"overwrite_if_exists"`
	RenameTo          types.String `tfsdk:"rename_to"`
}

type splitStringModel struct {
	Entries fwtypes.ListNestedObjectValueOf[splitStringEntryModel] `tfsdk:"entries"`
}

type splitStringEntryModel struct {
	Delimiter types.String `tfsdk:"delimiter"`
	Source    types.String `tfsdk:"source"`
}

type substituteStringModel struct {
	Entries fwtypes.ListNestedObjectValueOf[substituteStringEntryModel] `tfsdk:"entries"`
}

type substituteStringEntryModel struct {
	From   types.String `tfsdk:"from"`
	Source types.String `tfsdk:"source"`
	To     types.String `tfsdk:"to"`
}

type trimStringModel struct {
	WithKeys fwtypes.ListOfString `tfsdk:"with_keys"`
}

type typeConverterModel struct {
	Entries fwtypes.ListNestedObjectValueOf[typeConverterEntryModel] `tfsdk:"entries"`
}

type typeConverterEntryModel struct {
	Key  types.String                      `tfsdk:"key"`
	Type fwtypes.StringEnum[awstypes.Type] `tfsdk:"type"`
}

type upperCaseStringModel struct {
	WithKeys fwtypes.ListOfString `tfsdk:"with_keys"`
}

func sweepTransformers(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.LogsClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := cloudwatchlogs.NewDescribeLogGroupsPaginator(conn, &cloudwatchlogs.DescribeLogGroupsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.LogGroups {
			input := cloudwatchlogs.GetTransformerInput{
				LogGroupIdentifier: v.LogGroupName,
			}
			transformer, err := conn.GetTransformer(ctx, &input)
			if err != nil {
				return nil, err
			}

			if transformer == nil || len(transformer.TransformerConfig) == 0 {
				continue
			}

			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceTransformer, client,
				sweepfw.NewAttribute("log_group_identifier", aws.ToString(transformer.LogGroupIdentifier))),
			)
		}
	}

	return sweepResources, nil
}
