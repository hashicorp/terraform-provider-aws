// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// Copyright 2025 Twilio Inc.
// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package logs

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cloudwatch_log_transformer", name="Transformer")
// @ArnIdentity("log_group_arn")
// @Testing(hasNoPreExistingResource=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs;cloudwatchlogs.GetTransformerOutput")
func newTransformerResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &transformerResource{}

	return r, nil
}

type transformerResource struct {
	framework.ResourceWithModel[transformerResourceModel]
	framework.WithImportByIdentity
}

func (r *transformerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"log_group_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"transformer_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[processorModel](ctx),
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
									"entry": schema.ListNestedBlock{
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
									"entry": schema.ListNestedBlock{
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
									"entry": schema.ListNestedBlock{
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
											stringvalidator.RegexMatches(regexache.MustCompile(`^@message$`), "must be '@message'"),
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
											stringvalidator.RegexMatches(regexache.MustCompile(`^@message$`), "must be '@message'"),
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
									"entry": schema.ListNestedBlock{
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
									"entry": schema.ListNestedBlock{
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
									"entry": schema.ListNestedBlock{
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
									"entry": schema.ListNestedBlock{
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

func (r *transformerResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data transformerResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	logGroupARN := fwflex.StringValueFromFramework(ctx, data.LogGroupIdentifier)
	var input cloudwatchlogs.PutTransformerInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.PutTransformer(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating CloudWatch Logs Transformer (%s)", logGroupARN), err.Error())

		return
	}

	out, err := findTransformerByLogGroupIdentifier(ctx, conn, logGroupARN)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading CloudWatch Logs Transformer (%s)", logGroupARN), err.Error())

		return
	}

	// Set values for unknowns.
	response.Diagnostics.Append(fwflex.Flatten(ctx, out, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *transformerResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data transformerResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	logGroupARN := fwflex.StringValueFromFramework(ctx, data.LogGroupIdentifier)
	out, err := findTransformerByLogGroupIdentifier(ctx, conn, logGroupARN)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading CloudWatch Logs Transformer (%s)", logGroupARN), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, out, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *transformerResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old transformerResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	response.Diagnostics.Append(d...)
	if response.Diagnostics.HasError() {
		return
	}

	logGroupARN := fwflex.StringValueFromFramework(ctx, new.LogGroupIdentifier)

	if diff.HasChanges() {
		var input cloudwatchlogs.PutTransformerInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.PutTransformer(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating CloudWatch Logs Transformer (%s)", logGroupARN), err.Error())

			return
		}
	}

	out, err := findTransformerByLogGroupIdentifier(ctx, conn, logGroupARN)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading CloudWatch Logs Transformer (%s)", logGroupARN), err.Error())

		return
	}

	// Set values for unknowns.
	response.Diagnostics.Append(fwflex.Flatten(ctx, out, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *transformerResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data transformerResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	logGroupARN := fwflex.StringValueFromFramework(ctx, data.LogGroupIdentifier)
	input := cloudwatchlogs.DeleteTransformerInput{
		LogGroupIdentifier: aws.String(logGroupARN),
	}
	_, err := conn.DeleteTransformer(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting CloudWatch Logs Transformer (%s)", logGroupARN), err.Error())

		return
	}
}

func findTransformerByLogGroupIdentifier(ctx context.Context, conn *cloudwatchlogs.Client, logGroupIdentifier string) (*cloudwatchlogs.GetTransformerOutput, error) {
	input := cloudwatchlogs.GetTransformerInput{
		LogGroupIdentifier: aws.String(logGroupIdentifier),
	}

	return findTransformer(ctx, conn, &input)
}

func findTransformer(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.GetTransformerInput) (*cloudwatchlogs.GetTransformerOutput, error) {
	output, err := conn.GetTransformer(ctx, input)

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

type transformerResourceModel struct {
	framework.WithRegionModel
	LogGroupIdentifier fwtypes.ARN                                     `tfsdk:"log_group_arn" autoflex:",noflatten"`
	TransformerConfig  fwtypes.ListNestedObjectValueOf[processorModel] `tfsdk:"transformer_config"`
}

type processorModel struct {
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
	Entries fwtypes.ListNestedObjectValueOf[addKeysEntryModel] `tfsdk:"entry"`
}

type addKeysEntryModel struct {
	Key               types.String `tfsdk:"key"`
	OverwriteIfExists types.Bool   `tfsdk:"overwrite_if_exists"`
	Value             types.String `tfsdk:"value"`
}

type copyValueModel struct {
	Entries fwtypes.ListNestedObjectValueOf[copyValueEntryModel] `tfsdk:"entry"`
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
	Entries fwtypes.ListNestedObjectValueOf[moveKeysEntryModel] `tfsdk:"entry"`
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
	Entries fwtypes.ListNestedObjectValueOf[renameKeysEntryModel] `tfsdk:"entry"`
}

type renameKeysEntryModel struct {
	Key               types.String `tfsdk:"key"`
	OverwriteIfExists types.Bool   `tfsdk:"overwrite_if_exists"`
	RenameTo          types.String `tfsdk:"rename_to"`
}

type splitStringModel struct {
	Entries fwtypes.ListNestedObjectValueOf[splitStringEntryModel] `tfsdk:"entry"`
}

type splitStringEntryModel struct {
	Delimiter types.String `tfsdk:"delimiter"`
	Source    types.String `tfsdk:"source"`
}

type substituteStringModel struct {
	Entries fwtypes.ListNestedObjectValueOf[substituteStringEntryModel] `tfsdk:"entry"`
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
	Entries fwtypes.ListNestedObjectValueOf[typeConverterEntryModel] `tfsdk:"entry"`
}

type typeConverterEntryModel struct {
	Key  types.String                      `tfsdk:"key"`
	Type fwtypes.StringEnum[awstypes.Type] `tfsdk:"type"`
}

type upperCaseStringModel struct {
	WithKeys fwtypes.ListOfString `tfsdk:"with_keys"`
}
