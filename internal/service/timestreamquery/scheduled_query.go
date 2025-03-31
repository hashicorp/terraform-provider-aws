// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package timestreamquery

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/timestreamquery"
	awstypes "github.com/aws/aws-sdk-go-v2/service/timestreamquery/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_timestreamquery_scheduled_query", name="Scheduled Query")
// @Tags(identifierAttribute="arn")
func newResourceScheduledQuery(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceScheduledQuery{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameScheduledQuery         = "Scheduled Query"
	ScheduledQueryFieldNamePrefix = "ScheduledQuery"
)

type resourceScheduledQuery struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceScheduledQuery) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreationTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrExecutionRoleARN: schema.StringAttribute{
				Required: true,
			},
			names.AttrKMSKeyID: schema.StringAttribute{
				Optional: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			"next_invocation_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"previous_invocation_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"query_string": schema.StringAttribute{
				Required: true,
			},
			names.AttrState: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ScheduledQueryState](), // ENABLED, DISABLED
				Computed:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"error_report_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[errorReportConfiguration](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeBetween(1, 1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"s3_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[s3Configuration](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeBetween(1, 1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrBucketName: schema.StringAttribute{
										Required: true,
									},
									"encryption_option": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.S3EncryptionOption](),
										Optional:   true,
										Computed:   true,
									},
									"object_key_prefix": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"last_run_summary": schema.ListNestedBlock{ // Entirely Computed
				CustomType: fwtypes.NewListNestedObjectTypeOf[scheduledQueryRunSummary](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"failure_reason": schema.StringAttribute{
							Computed: true,
						},
						"invocation_time": schema.StringAttribute{
							CustomType: timetypes.RFC3339Type{},
							Computed:   true,
						},
						"run_status": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.ScheduledQueryRunStatus](),
							Computed:   true,
						},
						"trigger_time": schema.StringAttribute{
							CustomType: timetypes.RFC3339Type{},
							Computed:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"error_report_location": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[errorReportLocation](ctx),
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"s3_report_location": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[s3ReportLocation](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrBucketName: schema.StringAttribute{
													Computed: true,
												},
												"object_key": schema.StringAttribute{
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"execution_stats": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[executionStats](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"bytes_metered": schema.Int64Attribute{
										Computed: true,
									},
									"cumulative_bytes_scanned": schema.Int64Attribute{
										Computed: true,
									},
									"data_writes": schema.Int64Attribute{
										Computed: true,
									},
									"execution_time_in_millis": schema.Int64Attribute{
										Computed: true,
									},
									"query_result_rows": schema.Int64Attribute{
										Computed: true,
									},
									"records_ingested": schema.Int64Attribute{
										Computed: true,
									},
								},
							},
						},
						"query_insights_response": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[queryInsightsResponse](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"output_bytes": schema.Int64Attribute{
										Computed: true,
									},
									"output_rows": schema.Int64Attribute{
										Computed: true,
									},
									"query_table_count": schema.Int64Attribute{
										Computed: true,
									},
								},
								Blocks: map[string]schema.Block{
									"query_spatial_coverage": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[querySpatialCoverage](ctx),
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												names.AttrMax: schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[querySpatialCoverageMax](ctx),
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"partition_key": schema.ListAttribute{
																CustomType:  fwtypes.ListOfStringType,
																ElementType: types.StringType,
																Computed:    true,
															},
															"table_arn": schema.StringAttribute{
																Computed: true,
															},
															names.AttrValue: schema.Float64Attribute{
																Computed: true,
															},
														},
													},
												},
											},
										},
									},
									"query_temporal_range": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[queryTemporalRange](ctx),
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												names.AttrMax: schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[queryTemporalRangeMax](ctx),
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"table_arn": schema.StringAttribute{
																Computed: true,
															},
															names.AttrValue: schema.Int64Attribute{
																Computed: true,
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
			"notification_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[notificationConfiguration](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeBetween(1, 1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"sns_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[snsConfiguration](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeBetween(1, 1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrTopicARN: schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"recently_failed_runs": schema.ListNestedBlock{ // Entirely Computed
				CustomType: fwtypes.NewListNestedObjectTypeOf[scheduledQueryRunSummary](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"failure_reason": schema.StringAttribute{
							Computed: true,
						},
						"invocation_time": schema.StringAttribute{
							CustomType: timetypes.RFC3339Type{},
							Computed:   true,
						},
						"run_status": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.ScheduledQueryRunStatus](),
							Computed:   true,
						},
						"trigger_time": schema.StringAttribute{
							CustomType: timetypes.RFC3339Type{},
							Computed:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"error_report_location": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[errorReportLocation](ctx),
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"s3_report_location": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[s3ReportLocation](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrBucketName: schema.StringAttribute{
													Computed: true,
												},
												"object_key": schema.StringAttribute{
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"execution_stats": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[executionStats](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"bytes_metered": schema.Int64Attribute{
										Computed: true,
									},
									"cumulative_bytes_scanned": schema.Int64Attribute{
										Computed: true,
									},
									"data_writes": schema.Int64Attribute{
										Computed: true,
									},
									"execution_time_in_millis": schema.Int64Attribute{
										Computed: true,
									},
									"query_result_rows": schema.Int64Attribute{
										Computed: true,
									},
									"records_ingested": schema.Int64Attribute{
										Computed: true,
									},
								},
							},
						},
						"query_insights_response": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[queryInsightsResponse](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"output_bytes": schema.Int64Attribute{
										Computed: true,
									},
									"output_rows": schema.Int64Attribute{
										Computed: true,
									},
									"query_table_count": schema.Int64Attribute{
										Computed: true,
									},
								},
								Blocks: map[string]schema.Block{
									"query_spatial_coverage": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[querySpatialCoverage](ctx),
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												names.AttrMax: schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[querySpatialCoverageMax](ctx),
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"partition_key": schema.ListAttribute{
																CustomType:  fwtypes.ListOfStringType,
																ElementType: types.StringType,
																Computed:    true,
															},
															"table_arn": schema.StringAttribute{
																Computed: true,
															},
															names.AttrValue: schema.Float64Attribute{
																Computed: true,
															},
														},
													},
												},
											},
										},
									},
									"query_temporal_range": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[queryTemporalRange](ctx),
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												names.AttrMax: schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[queryTemporalRangeMax](ctx),
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"table_arn": schema.StringAttribute{
																Computed: true,
															},
															names.AttrValue: schema.Int64Attribute{
																Computed: true,
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
			"schedule_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[scheduleConfiguration](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeBetween(1, 1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrScheduleExpression: schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			"target_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[targetConfiguration](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeBetween(1, 1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"timestream_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[timestreamConfiguration](ctx),
							Validators: []validator.List{
								listvalidator.AtLeastOneOf(
									path.MatchRelative().AtName("mixed_measure_mapping"),
									path.MatchRelative().AtName("multi_measure_mappings"),
								),
								listvalidator.SizeAtMost(1),
								listvalidator.IsRequired(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrDatabaseName: schema.StringAttribute{
										Required: true,
									},
									names.AttrTableName: schema.StringAttribute{
										Required: true,
									},
									"time_column": schema.StringAttribute{
										Required: true,
									},
									"measure_name_column": schema.StringAttribute{
										Optional: true,
									},
								},
								Blocks: map[string]schema.Block{
									"dimension_mapping": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[dimensionMapping](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtLeast(1),
											listvalidator.IsRequired(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"dimension_value_type": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.DimensionValueType](),
													Required:   true,
												},
												names.AttrName: schema.StringAttribute{
													Required: true,
												},
											},
										},
									},
									"mixed_measure_mapping": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[mixedMeasureMapping](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"measure_value_type": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.MeasureValueType](),
													Required:   true,
												},
												"measure_name": schema.StringAttribute{
													Optional: true,
												},
												"source_column": schema.StringAttribute{
													Optional: true,
												},
												"target_measure_name": schema.StringAttribute{
													Optional: true,
												},
											},
											Blocks: map[string]schema.Block{
												"multi_measure_attribute_mapping": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[multiMeasureAttributeMapping](ctx),
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"measure_value_type": schema.StringAttribute{
																CustomType: fwtypes.StringEnumType[awstypes.ScalarMeasureValueType](),
																Required:   true,
															},
															"source_column": schema.StringAttribute{
																Required: true,
															},
															"target_multi_measure_attribute_name": schema.StringAttribute{
																Optional: true,
															},
														},
													},
												},
											},
										},
									},
									"multi_measure_mappings": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[multiMeasureMappings](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"target_multi_measure_name": schema.StringAttribute{
													Optional: true,
												},
											},
											Blocks: map[string]schema.Block{
												"multi_measure_attribute_mapping": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[multiMeasureAttributeMapping](ctx),
													Validators: []validator.List{
														listvalidator.IsRequired(),
														listvalidator.SizeAtLeast(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"measure_value_type": schema.StringAttribute{
																CustomType: fwtypes.StringEnumType[awstypes.ScalarMeasureValueType](),
																Required:   true,
															},
															"source_column": schema.StringAttribute{
																Required: true,
															},
															"target_multi_measure_attribute_name": schema.StringAttribute{
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

func (r *resourceScheduledQuery) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().TimestreamQueryClient(ctx)

	var plan resourceScheduledQueryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input timestreamquery.CreateScheduledQueryInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix(ScheduledQueryFieldNamePrefix))...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientToken := id.UniqueId()
	input.ClientToken = aws.String(clientToken)

	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateScheduledQuery(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamQuery, create.ErrActionCreating, ResNameScheduledQuery, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamQuery, create.ErrActionCreating, ResNameScheduledQuery, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	// CreateScheduledQueryOutput only contains ARN
	plan.ARN = types.StringValue(aws.ToString(out.Arn))

	sqOut, err := waitScheduledQueryCreated(ctx, conn, plan.ARN.ValueString(), r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamQuery, create.ErrActionWaitingForCreation, ResNameScheduledQuery, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, sqOut, &plan, flex.WithFieldNamePrefix(ScheduledQueryFieldNamePrefix))...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceScheduledQuery) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().TimestreamQueryClient(ctx)

	var state resourceScheduledQueryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findScheduledQueryByARN(ctx, conn, state.ARN.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamQuery, create.ErrActionSetting, ResNameScheduledQuery, state.ARN.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state, flex.WithFieldNamePrefix(ScheduledQueryFieldNamePrefix))...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceScheduledQuery) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().TimestreamQueryClient(ctx)

	var state, plan resourceScheduledQueryModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input timestreamquery.UpdateScheduledQueryInput
		resp.Diagnostics.Append(flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix(ScheduledQueryFieldNamePrefix))...)
		if resp.Diagnostics.HasError() {
			return
		}

		output, err := conn.UpdateScheduledQuery(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.MediaPackageV2, create.ErrActionUpdating, ResNameScheduledQuery, state.Name.String(), err),
				err.Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, output, &plan, flex.WithFieldNamePrefix(ScheduledQueryFieldNamePrefix))...)
		if resp.Diagnostics.HasError() {
			return
		}

		_, err = waitScheduledQueryUpdated(ctx, conn, plan.ARN.ValueString(), r.UpdateTimeout(ctx, plan.Timeouts))
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.TimestreamQuery, create.ErrActionWaitingForUpdate, ResNameScheduledQuery, plan.ARN.String(), err),
				err.Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceScheduledQuery) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().TimestreamQueryClient(ctx)

	var state resourceScheduledQueryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := timestreamquery.DeleteScheduledQueryInput{
		ScheduledQueryArn: state.ARN.ValueStringPointer(),
	}

	_, err := conn.DeleteScheduledQuery(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamQuery, create.ErrActionDeleting, ResNameScheduledQuery, state.ARN.String(), err),
			err.Error(),
		)
		return
	}

	_, err = waitScheduledQueryDeleted(ctx, conn, state.ARN.ValueString(), r.DeleteTimeout(ctx, state.Timeouts))
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamQuery, create.ErrActionWaitingForDeletion, ResNameScheduledQuery, state.ARN.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceScheduledQuery) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrARN), req, resp)
}

func waitScheduledQueryCreated(ctx context.Context, conn *timestreamquery.Client, id string, timeout time.Duration) (*awstypes.ScheduledQueryDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(awstypes.ScheduledQueryStateEnabled),
		Refresh:                   statusScheduledQuery(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.ScheduledQueryDescription); ok {
		return out, err
	}

	return nil, err
}

func waitScheduledQueryUpdated(ctx context.Context, conn *timestreamquery.Client, arn string, timeout time.Duration) (*awstypes.ScheduledQueryDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ScheduledQueryStateDisabled),
		Target:                    enum.Slice(awstypes.ScheduledQueryStateEnabled),
		Refresh:                   statusScheduledQuery(ctx, conn, arn),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.ScheduledQueryDescription); ok {
		return out, err
	}

	return nil, err
}

func waitScheduledQueryDeleted(ctx context.Context, conn *timestreamquery.Client, arn string, timeout time.Duration) (*awstypes.ScheduledQueryDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ScheduledQueryStateEnabled, awstypes.ScheduledQueryStateDisabled),
		Target:  []string{},
		Refresh: statusScheduledQuery(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.ScheduledQueryDescription); ok {
		return out, err
	}

	return nil, err
}

// statusScheduledQuery is a state refresh function that queries the service
// and returns the state of the scheduled query, not the run status of the most
// recent run.
func statusScheduledQuery(ctx context.Context, conn *timestreamquery.Client, arn string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findScheduledQueryByARN(ctx, conn, arn)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.State), nil
	}
}

func findScheduledQueryByARN(ctx context.Context, conn *timestreamquery.Client, arn string) (*awstypes.ScheduledQueryDescription, error) {
	in := &timestreamquery.DescribeScheduledQueryInput{
		ScheduledQueryArn: aws.String(arn),
	}

	out, err := conn.DescribeScheduledQuery(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.ScheduledQuery == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.ScheduledQuery, nil
}

type resourceScheduledQueryModel struct {
	// Attributes
	ARN                    types.String                                     `tfsdk:"arn"`
	CreationTime           timetypes.RFC3339                                `tfsdk:"creation_time"`
	ExecutionRoleARN       types.String                                     `tfsdk:"execution_role_arn"`
	KMSKeyID               types.String                                     `tfsdk:"kms_key_id"`
	Name                   types.String                                     `tfsdk:"name"`
	NextInvocationTime     timetypes.RFC3339                                `tfsdk:"next_invocation_time"`
	PreviousInvocationTime timetypes.RFC3339                                `tfsdk:"previous_invocation_time"`
	QueryString            types.String                                     `tfsdk:"query_string"`
	State                  fwtypes.StringEnum[awstypes.ScheduledQueryState] `tfsdk:"state"`
	Tags                   tftags.Map                                       `tfsdk:"tags"`
	TagsAll                tftags.Map                                       `tfsdk:"tags_all"`

	// Blocks
	ErrorReportConfiguration  fwtypes.ListNestedObjectValueOf[errorReportConfiguration]  `tfsdk:"error_report_configuration"`
	LastRunSummary            fwtypes.ListNestedObjectValueOf[scheduledQueryRunSummary]  `tfsdk:"last_run_summary"`
	NotificationConfiguration fwtypes.ListNestedObjectValueOf[notificationConfiguration] `tfsdk:"notification_configuration"`
	RecentlyFailedRuns        fwtypes.ListNestedObjectValueOf[scheduledQueryRunSummary]  `tfsdk:"recently_failed_runs"`
	ScheduleConfiguration     fwtypes.ListNestedObjectValueOf[scheduleConfiguration]     `tfsdk:"schedule_configuration"`
	TargetConfiguration       fwtypes.ListNestedObjectValueOf[targetConfiguration]       `tfsdk:"target_configuration"`
	Timeouts                  timeouts.Value                                             `tfsdk:"timeouts"`
}

type errorReportConfiguration struct {
	S3Configuration fwtypes.ListNestedObjectValueOf[s3Configuration] `tfsdk:"s3_configuration"`
}

type s3Configuration struct {
	BucketName       types.String                                    `tfsdk:"bucket_name"`
	EncryptionOption fwtypes.StringEnum[awstypes.S3EncryptionOption] `tfsdk:"encryption_option"`
	ObjectKeyPrefix  types.String                                    `tfsdk:"object_key_prefix"`
}

type notificationConfiguration struct {
	SNSConfiguration fwtypes.ListNestedObjectValueOf[snsConfiguration] `tfsdk:"sns_configuration"`
}

type snsConfiguration struct {
	TopicARN types.String `tfsdk:"topic_arn"`
}

type scheduleConfiguration struct {
	ScheduleExpression types.String `tfsdk:"schedule_expression"`
}

type targetConfiguration struct {
	TimestreamConfiguration fwtypes.ListNestedObjectValueOf[timestreamConfiguration] `tfsdk:"timestream_configuration"`
}

type timestreamConfiguration struct {
	// Attributes
	DatabaseName      types.String `tfsdk:"database_name"`
	TableName         types.String `tfsdk:"table_name"`
	TimeColumn        types.String `tfsdk:"time_column"`
	MeasureNameColumn types.String `tfsdk:"measure_name_column"`

	// Blocks
	DimensionMapping     fwtypes.ListNestedObjectValueOf[dimensionMapping]     `tfsdk:"dimension_mapping"`
	MixedMeasureMapping  fwtypes.ListNestedObjectValueOf[mixedMeasureMapping]  `tfsdk:"mixed_measure_mapping"`
	MultiMeasureMappings fwtypes.ListNestedObjectValueOf[multiMeasureMappings] `tfsdk:"multi_measure_mappings"`
}

type dimensionMapping struct {
	DimensionValueType fwtypes.StringEnum[awstypes.DimensionValueType] `tfsdk:"dimension_value_type"`
	Name               types.String                                    `tfsdk:"name"`
}

type mixedMeasureMapping struct {
	// Attributes
	MeasureValueType  fwtypes.StringEnum[awstypes.MeasureValueType] `tfsdk:"measure_value_type"`
	MeasureName       types.String                                  `tfsdk:"measure_name"`
	SourceColumn      types.String                                  `tfsdk:"source_column"`
	TargetMeasureName types.String                                  `tfsdk:"target_measure_name"`

	// Blocks
	MultiMeasureAttributeMapping fwtypes.ListNestedObjectValueOf[multiMeasureAttributeMapping] `tfsdk:"multi_measure_attribute_mapping"`
}

type multiMeasureAttributeMapping struct {
	MeasureValueType                fwtypes.StringEnum[awstypes.ScalarMeasureValueType] `tfsdk:"measure_value_type"`
	SourceColumn                    types.String                                        `tfsdk:"source_column"`
	TargetMultiMeasureAttributeName types.String                                        `tfsdk:"target_multi_measure_attribute_name"`
}

type multiMeasureMappings struct {
	// Attributes
	TargetMultiMeasureName types.String `tfsdk:"target_multi_measure_name"`

	// Blocks
	MultiMeasureAttributeMapping fwtypes.ListNestedObjectValueOf[multiMeasureAttributeMapping] `tfsdk:"multi_measure_attribute_mapping"`
}

type scheduledQueryRunSummary struct {
	// Attributes
	FailureReason  types.String                                         `tfsdk:"failure_reason"`
	InvocationTime timetypes.RFC3339                                    `tfsdk:"invocation_time"`
	RunStatus      fwtypes.StringEnum[awstypes.ScheduledQueryRunStatus] `tfsdk:"run_status"`
	TriggerTime    timetypes.RFC3339                                    `tfsdk:"trigger_time"`

	// Blocks
	ErrorReportLocation   fwtypes.ListNestedObjectValueOf[errorReportLocation]   `tfsdk:"error_report_location"`
	ExecutionStats        fwtypes.ListNestedObjectValueOf[executionStats]        `tfsdk:"execution_stats"`
	QueryInsightsResponse fwtypes.ListNestedObjectValueOf[queryInsightsResponse] `tfsdk:"query_insights_response"`
}

type errorReportLocation struct {
	S3ReportLocation fwtypes.ListNestedObjectValueOf[s3ReportLocation] `tfsdk:"s3_report_location"`
}

type s3ReportLocation struct {
	BucketName types.String `tfsdk:"bucket_name"`
	ObjectKey  types.String `tfsdk:"object_key"`
}

type executionStats struct {
	BytesMetered           types.Int64 `tfsdk:"bytes_metered"`
	CumulativeBytesScanned types.Int64 `tfsdk:"cumulative_bytes_scanned"`
	DataWrites             types.Int64 `tfsdk:"data_writes"`
	ExecutionTimeInMillis  types.Int64 `tfsdk:"execution_time_in_millis"`
	QueryResultRows        types.Int64 `tfsdk:"query_result_rows"`
	RecordsIngested        types.Int64 `tfsdk:"records_ingested"`
}

type queryInsightsResponse struct {
	// Attributes
	OutputBytes     types.Int64 `tfsdk:"output_bytes"`
	OutputRows      types.Int64 `tfsdk:"output_rows"`
	QueryTableCount types.Int64 `tfsdk:"query_table_count"`

	// Blocks
	QuerySpatialCoverage fwtypes.ListNestedObjectValueOf[querySpatialCoverage] `tfsdk:"query_spatial_coverage"`
	QueryTemporalRange   fwtypes.ListNestedObjectValueOf[queryTemporalRange]   `tfsdk:"query_temporal_range"`
}

type querySpatialCoverage struct {
	Max fwtypes.ListNestedObjectValueOf[querySpatialCoverageMax] `tfsdk:"max"`
}

type querySpatialCoverageMax struct {
	PartitionKey fwtypes.ListValueOf[types.String] `tfsdk:"partition_key"`
	TableARN     types.String                      `tfsdk:"table_arn"`
	Value        types.Float64                     `tfsdk:"value"`
}

type queryTemporalRange struct {
	Max fwtypes.ListNestedObjectValueOf[queryTemporalRangeMax] `tfsdk:"max"`
}

type queryTemporalRangeMax struct {
	TableARN types.String `tfsdk:"table_arn"`
	Value    types.Int64  `tfsdk:"value"`
}
