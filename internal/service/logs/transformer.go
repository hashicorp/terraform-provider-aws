// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// TIP: ==== FILE STRUCTURE ====
// All resources should follow this basic outline. Improve this resource's
// maintainability by sticking to it.
//
// 1. Package declaration
// 2. Imports
// 3. Main resource struct with schema method
// 4. Create, read, update, delete methods (in that order)
// 5. Other functions (flatteners, expanders, waiters, finders, etc.)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_logs_transformer", name="Transformer")
func newResourceTransformer(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceTransformer{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameTransformer = "Transformer"
)

type resourceTransformer struct {
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
func (r *resourceTransformer) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrCreationTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"last_modified_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"log_group_identifier": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 2048),
					stringvalidator.RegexMatches(regexache.MustCompile(`[\w#+=/:,.@-]*`), "must be a valid log group name or ARN"),
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
						"add_keys": schema.SingleNestedBlock{
							CustomType: fwtypes.NewObjectTypeOf[addKeysModel](ctx),
							Validators: []validator.Object{
								// TBD
							},
							Blocks: map[string]schema.Block{
								"entries": schema.ListNestedBlock{
									CustomType: fwtypes.NewListNestedObjectTypeOf[addKeysEntryModel](ctx),
									Validators: []validator.List{
										listvalidator.IsRequired(),
										listvalidator.SizeBetween(1,5),
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
						"copy_value": schema.SingleNestedBlock{
							CustomType: fwtypes.NewObjectTypeOf[copyValueModel](ctx),
							Validators: []validator.Object{
								// TBD
							},
							Blocks: map[string]schema.Block{
								"entries": schema.ListNestedBlock{
									CustomType: fwtypes.NewListNestedObjectTypeOf[copyValueEntryModel](ctx),
									Validators: []validator.List{
										listvalidator.IsRequired(),
										listvalidator.SizeBetween(1,5),
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
						"csv": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[csvModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(5),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"columns": schema.ListAttribute{
										Optional: true,
										Computed: true,
										CustomType: fwtypes.ListOfStringType,
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
										Required: true,
										CustomType: fwtypes.ListOfStringType,
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
										Required: true,
										CustomType: fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Validators: []validator.List{
											listvalidator.SizeBetween(1, 5),
											listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
										},
									},
								},
							},
						},
						"grok": schema.SingleNestedBlock{
							CustomType: fwtypes.NewObjectTypeOf[grokModel](ctx),
							Validators: []validator.Object{
								// TBD
							},
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
										Optional: true,
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
										Required: true,
										CustomType: fwtypes.ListOfStringType,
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
											listvalidator.SizeBetween(1,5),
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
						"parse_cloudfront": schema.SingleNestedBlock{
							CustomType: fwtypes.NewObjectTypeOf[parseCloudfrontModel](ctx),
							Validators: []validator.Object{
								// TBD
							},
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
						"parse_postgres": schema.SingleNestedBlock{
							CustomType: fwtypes.NewObjectTypeOf[parsePostgresModel](ctx),
							Validators: []validator.Object{
								// TBD
							},
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
						"parse_route53": schema.SingleNestedBlock{
							CustomType: fwtypes.NewObjectTypeOf[parseRoute53Model](ctx),
							Validators: []validator.Object{
								// TBD
							},
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
						"parse_vpc": schema.SingleNestedBlock{
							CustomType: fwtypes.NewObjectTypeOf[parseVPCModel](ctx),
							Validators: []validator.Object{
								// TBD
							},
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
						"parse_waf": schema.SingleNestedBlock{
							CustomType: fwtypes.NewObjectTypeOf[parseWAFModel](ctx),
							Validators: []validator.Object{
								// TBD
							},
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
											listvalidator.SizeBetween(1,5),
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
											listvalidator.SizeBetween(1,10),
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
											listvalidator.SizeBetween(1,10),
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
										Required: true,
										CustomType: fwtypes.ListOfStringType,
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
											listvalidator.SizeBetween(1,5),
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
													Required: true,
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
										Required: true,
										CustomType: fwtypes.ListOfStringType,
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
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

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	transformer, err := waitTransformerCreated(ctx, conn, plan.LogGroupIdentifier.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Logs, create.ErrActionWaitingForCreation, ResNameTransformer, plan.LogGroupIdentifier.String(), err),
			err.Error(),
		)
		return
	}

	// PutTransformer returns an empty body, so we propagate the state from the status call
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
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Logs, create.ErrActionReading, ResNameTransformer, state.ID.String(), err),
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

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	transformer, err := waitTransformerUpdated(ctx, conn, plan.LogGroupIdentifier.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Logs, create.ErrActionWaitingForUpdate, ResNameTransformer, plan.LogGroupIdentifier.String(), err),
			err.Error(),
		)
		return
	}

	// PutTransformer returns an empty body, so we propagate the state from the status call
	resp.Diagnostics.Append(flex.Flatten(ctx, transformer, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceTransformer) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// TIP: ==== RESOURCE DELETE ====
	// Most resources have Delete functions. There are rare situations
	// where you might not need a delete:
	// a. The AWS API does not provide a way to delete the resource
	// b. The point of your resource is to perform an action (e.g., reboot a
	//    server) and deleting serves no purpose.
	//
	// The Delete function should do the following things. Make sure there
	// is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Fetch the state
	// 3. Populate a delete input structure
	// 4. Call the AWS delete function
	// 5. Use a waiter to wait for delete to complete
	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().LogsClient(ctx)

	// TIP: -- 2. Fetch the state
	var state resourceTransformerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a delete input structure
	input := logs.DeleteTransformerInput{
		TransformerId: state.ID.ValueStringPointer(),
	}

	// TIP: -- 4. Call the AWS delete function
	_, err := conn.DeleteTransformer(ctx, &input)
	// TIP: On rare occassions, the API returns a not found error after deleting a
	// resource. If that happens, we don't want it to show up as an error.
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Logs, create.ErrActionDeleting, ResNameTransformer, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 5. Use a waiter to wait for delete to complete
	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitTransformerDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Logs, create.ErrActionWaitingForDeletion, ResNameTransformer, state.ID.String(), err),
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
func (r *resourceTransformer) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

// TIP: ==== STATUS CONSTANTS ====
// Create constants for states and statuses if the service does not
// already have suitable constants. We prefer that you use the constants
// provided in the service if available (e.g., awstypes.StatusInProgress).
const (
	statusNormal        = "Normal"
)

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
func waitTransformerCreated(ctx context.Context, conn *cloudwatchlogs.Client, logGroupIdentifier string, timeout time.Duration) (*cloudwatchlogs.GetTransformerOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusTransformer(ctx, conn, logGroupIdentifier),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*cloudwatchlogs.GetTransformerOutput); ok {
		return out, err
	}

	return nil, err
}

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.
func waitTransformerUpdated(ctx context.Context, conn *cloudwatchlogs.Client, logGroupIdentifier string, timeout time.Duration) (*cloudwatchlogs.GetTransformerOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusTransformer(ctx, conn, logGroupIdentifier),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*cloudwatchlogs.GetTransformerOutput); ok {
		return out, err
	}

	return nil, err
}

// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.
func waitTransformerDeleted(ctx context.Context, conn *cloudwatchlogs.Client, logGroupIdentifier string, timeout time.Duration) (*cloudwatchlogs.GetTransformerOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  []string{},
		Refresh: statusTransformer(ctx, conn, logGroupIdentifier),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*cloudwatchlogs.GetTransformerOutput); ok {
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
func statusTransformer(ctx context.Context, conn *cloudwatchlogs.Client, logGroupIdentifier string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findTransformerByLogGroupIdentifier(ctx, conn, logGroupIdentifier)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, statusNormal, nil
	}
}

func findTransformerByLogGroupIdentifier(ctx context.Context, conn *cloudwatchlogs.Client, logGroupIdentifier string) (*cloudwatchlogs.GetTransformerOutput, error) {
	input := cloudwatchlogs.GetTransformerInput{
		LogGroupIdentifier: aws.String(logGroupIdentifier),
	}

	out, err := conn.GetTransformer(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
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
	CreationTime timetypes.RFC3339Type                                          `tfsdk:"creation_time"`
	LastModifiedTime timetypes.RFC3339Type                                          `tfsdk:"last_modified_time"`
	LogGroupIdentifier     types.String                                          `tfsdk:"log_group_identifier"`
	TransformerConfig              fwtypes.ListNestedObjectValueOf[transformerConfigModel]                                         `tfsdk:"transformer_config"`
	Timeouts        timeouts.Value                                        `tfsdk:"timeouts"`
}

type transformerConfigModel struct {
	AddKeys fwtypes.ObjectValueOf[addKeysModel] `tfsdk:"add_keys"`
	CopyValue fwtypes.ObjectValueOf[copyValueModel] `tfsdk:"copy_value"`
	CSV fwtypes.ListNestedObjectValueOf[csvModel] `tfsdk:"csv"`
	DateTimeConverter fwtypes.ListNestedObjectValueOf[dateTimeConverterModel] `tfsdk:"date_time_converter"`
	DeleteKeys fwtypes.ListNestedObjectValueOf[deleteKeysModel] `tfsdk:"delete_keys"`
	Grok fwtypes.ObjectValueOf[grokModel] `tfsdk:"grok"`
	ListToMap fwtypes.ListNestedObjectValueOf[listToMapModel] `tfsdk:"list_to_map"`
	LowerCaseString fwtypes.ListNestedObjectValueOf[lowerCaseStringModel] `tfsdk:"lower_case_string"`
	MoveKeys fwtypes.ListNestedObjectValueOf[moveKeysModel] `tfsdk:"move_keys"`
	ParseCloudfront fwtypes.ObjectValueOf[parseCloudfrontModel] `tfsdk:"parse_cloudfront"`
	ParseJSON fwtypes.ListNestedObjectValueOf[parseJSONModel] `tfsdk:"parse_json"`
	ParseKeyValue fwtypes.ListNestedObjectValueOf[parseKeyValueModel] `tfsdk:"parse_key_value"`
	ParsePostgres fwtypes.ObjectValueOf[parsePostgresModel] `tfsdk:"parse_postgres"`
	ParseRoute53 fwtypes.ObjectValueOf[parseRoute53Model] `tfsdk:"parse_route53"`
	ParseVPC fwtypes.ObjectValueOf[parseVPCModel] `tfsdk:"parse_vpc"`
	ParseWAF fwtypes.ObjectValueOf[parseWAFModel] `tfsdk:"parse_waf"`
	RenameEntries fwtypes.ListNestedObjectValueOf[renameKeysModel] `tfsdk:"rename_entries"`
	SplitString fwtypes.ListNestedObjectValueOf[splitStringModel] `tfsdk:"split_string"`
	SubstituteString fwtypes.ListNestedObjectValueOf[substituteStringModel] `tfsdk:"substitute_string"`
	TrimString fwtypes.ListNestedObjectValueOf[trimStringModel] `tfsdk:"trim_string"`
	TypeConverter fwtypes.ListNestedObjectValueOf[typeConverterModel] `tfsdk:"type_converter"`
	UpperCaseString fwtypes.ListNestedObjectValueOf[upperCaseStringModel] `tfsdk:"upper_case_string"`
}

type addKeysModel struct {
	Entries fwtypes.ListNestedObjectValueOf[addKeysEntryModel] `tfsdk:"entries"`
}

type addKeysEntryModel struct {
	Key types.String `tfsdk:"key"`
	OverwriteIfExists types.Bool `tfsdk:"overwrite_if_exists"`
	Value types.String `tfsdk:"value"`
}

type copyValueModel struct {
	Entries fwtypes.ListNestedObjectValueOf[copyValueEntryModel] `tfsdk:"entries"`
}

type copyValueEntryModel struct {
	OverwriteIfExists types.Bool `tfsdk:"overwrite_if_exists"`
	Source types.String `tfsdk:"source"`
	Target types.String `tfsdk:"target"`
}

type csvModel struct {
	Columns types.List `tfsdk:"columns"`
	Delimiter types.String `tfsdk:"delimiter"`
	QuoteCharacter types.String `tfsdk:"quote_character"`
	Source types.String `tfsdk:"source"`
}

type dateTimeConverterModel struct {
	Locale types.String `tfsdk:"locale"`
	MatchPatterns types.List `tfsdk:"match_patterns"`
	Source types.String `tfsdk:"source"`
	SourceTimezone types.String `tfsdk:"source_timezone"`
	Target types.String `tfsdk:"target"`
	TargetFormat types.String `tfsdk:"target_format"`
	TargetTimezone types.String `tfsdk:"target_timezone"`
}

type deleteKeysModel struct {
	WithKeys types.List `tfsdk:"with_keys"`
}

type grokModel struct {
	Match types.String `tfsdk:"match"`
	Source types.String `tfsdk:"source"`
}

type listToMapModel struct {
	Flatten types.Bool `tfsdk:"flatten"`
	FlattenedElement fwtypes.StringEnum[awstypes.FlattenedElement] `tfsdk:"flattened_element"`
	Key types.String `tfsdk:"key"`
	Source types.String `tfsdk:"source"`
	Target types.String `tfsdk:"target"`
	ValueKey types.String `tfsdk:"value_key"`
}

type lowerCaseStringModel struct {
	WithKeys types.List `tfsdk:"with_keys"`
}

type moveKeysModel struct {
	Entries fwtypes.ListNestedObjectValueOf[moveKeysEntryModel] `tfsdk:"entries"`
}

type moveKeysEntryModel struct {
	OverwriteIfExists types.Bool `tfsdk:"overwrite_if_exists"`
	Source types.String `tfsdk:"source"`
	Target types.String `tfsdk:"target"`
}

type parseCloudfrontModel struct {
	Source types.String `tfsdk:"source"`
}

type parseJSONModel struct {
	Destination types.String `tfsdk:"destination"`
	Source types.String `tfsdk:"source"`
}

type parseKeyValueModel struct {
	Destination types.String `tfsdk:"destination"`
	FieldDelimiter types.String `tfsdk:"field_delimiter"`
	KeyPrefix types.String `tfsdk:"key_prefix"`
	KeyValueDelimiter types.String `tfsdk:"key_value_delimiter"`
	NonMatchValue types.String `tfsdk:"non_match_value"`
	OverwriteIfExists types.Bool `tfsdk:"overwrite_if_exists"`
	Source types.String `tfsdk:"source"`
}

type parsePostgresModel struct {
	Source types.String `tfsdk:"source"`
}

type parseRoute53Model struct {
	Source types.String `tfsdk:"source"`
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
	Key types.String `tfsdk:"key"`
	OverwriteIfExists types.Bool `tfsdk:"overwrite_if_exists"`
	RenameTo types.String `tfsdk:"rename_to"`
}

type splitStringModel struct {
	Entries fwtypes.ListNestedObjectValueOf[splitStringEntryModel] `tfsdk:"entries"`
}

type splitStringEntryModel struct {
	Delimiter types.String `tfsdk:"delimiter"`
	Source types.String `tfsdk:"source"`
}

type substituteStringModel struct {
	Entries fwtypes.ListNestedObjectValueOf[substituteStringEntryModel] `tfsdk:"entries"`
}

type substituteStringEntryModel struct {
	From types.String `tfsdk:"from"`
	Source types.String `tfsdk:"source"`
	To types.String `tfsdk:"to"`
}

type trimStringModel struct {
	WithKeys types.List `tfsdk:"with_keys"`
}

type typeConverterModel struct {
	Entries fwtypes.ListNestedObjectValueOf[typeConverterEntryModel] `tfsdk:"entries"`
}

type typeConverterEntryModel struct {
	Key types.String `tfsdk:"key"`
	Type fwtypes.StringEnum[awstypes.Type] `tfsdk:"type"`
}

type upperCaseStringModel struct {
	WithKeys types.List `tfsdk:"with_keys"`
}

// TIP: ==== SWEEPERS ====
// When acceptance testing resources, interrupted or failed tests may
// leave behind orphaned resources in an account. To facilitate cleaning
// up lingering resources, each resource implementation should include
// a corresponding "sweeper" function.
//
// The sweeper function lists all resources of a given type and sets the
// appropriate identifers required to delete the resource via the Delete
// method implemented above.
//
// Once the sweeper function is implemented, register it in sweeper.go
// as follows:
//
//	awsv2.Register("aws_logs_transformer", sweepTransformers)
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/running-and-writing-acceptance-tests/#acceptance-test-sweepers
func sweepTransformers(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := logs.ListTransformersInput{}
	conn := client.LogsClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := logs.NewListTransformersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.Transformers {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceTransformer, client,
				sweepfw.NewAttribute(names.AttrID, aws.ToString(v.TransformerId))),
			)
		}
	}

	return sweepResources, nil
}
