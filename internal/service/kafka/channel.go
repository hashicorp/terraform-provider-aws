// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kafka

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_msk_channel", name="Channel")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("arn")
// @IdentityAttribute("cluster_arn")
// @ImportIDHandler(channelImportID)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/kafka;kafka.DescribeChannelOutput")
// @Testing(preCheck="testAccPreCheck")
// @Testing(hasNoPreExistingResource=true)
// @Testing(tagsTest=false)
// @Testing(identityRegionOverrideTest=false)
func newChannelResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &channelResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type channelResource struct {
	framework.ResourceWithModel[channelResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *channelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"channel_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cluster_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cluster_operation_arn": schema.StringAttribute{
				Computed: true,
			},
			"creation_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"destination_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ChannelDestinationType](),
				Computed:   true,
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ChannelStatus](),
				Computed:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
			"topic_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[topicConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"topic_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"record_converter": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[recordConverterModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeBetween(1, 1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"value_converter": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.ValueConverter](),
										Required:   true,
									},
								},
							},
						},
						"record_schema": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[recordSchemaModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"gsr_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
								},
							},
						},
					},
				},
			},
			"iceberg_destination": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[icebergDestinationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"append_only": schema.BoolAttribute{
							Required: true,
						},
						"compression_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.IcebergCompressionType](),
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"data_freshness_in_seconds": schema.Int32Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.UseStateForUnknown(),
							},
						},
						"service_execution_role_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"catalog": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[catalogModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"catalog_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Optional:   true,
									},
									"warehouse_location": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Optional:   true,
									},
								},
							},
						},
						"dead_letter_queue_s3": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[deadLetterQueueS3Model](ctx),
							Validators: []validator.List{
								listvalidator.SizeBetween(1, 1),
							},
							NestedObject: deadLetterQueueS3NestedObject(),
						},
						"destination_table": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[destinationTableModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeBetween(1, 1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"destination_database_name": schema.StringAttribute{
										Optional: true,
									},
									"destination_table_name": schema.StringAttribute{
										Optional: true,
									},
								},
								Blocks: map[string]schema.Block{
									"partition_spec": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[partitionSpecModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"partition_strategy": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.PartitionStrategy](),
													Required:   true,
												},
											},
											Blocks: map[string]schema.Block{
												"source": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[partitionSourceModel](ctx),
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"source_name": schema.StringAttribute{
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
						"schema_evolution": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[schemaEvolutionModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeBetween(1, 1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"enable_schema_evolution": schema.BoolAttribute{
										Optional: true,
									},
								},
							},
						},
						"table_creation": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[tableCreationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeBetween(1, 1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"enable_table_creation": schema.BoolAttribute{
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"s3_destination": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[s3DestinationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"data_freshness_in_seconds": schema.Int32Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.UseStateForUnknown(),
							},
						},
						"service_execution_role_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"dead_letter_queue_s3": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[deadLetterQueueS3Model](ctx),
							Validators: []validator.List{
								listvalidator.SizeBetween(1, 1),
							},
							NestedObject: deadLetterQueueS3NestedObject(),
						},
						"storage": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[s3StorageModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeBetween(1, 1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"bucket_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
									"compression_type": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.S3CompressionType](),
										Required:   true,
									},
									"expected_bucket_owner": schema.StringAttribute{
										Optional: true,
									},
									"output_key_template": schema.StringAttribute{
										Optional: true,
									},
									"output_prefix": schema.StringAttribute{
										Optional: true,
									},
									"storage_class": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.S3StorageClass](),
										Required:   true,
									},
								},
							},
						},
					},
				},
			},
			"encryption_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[encryptionConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrKMSKeyARN: schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
						},
					},
				},
			},
			"logging_info": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[channelLoggingInfoModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"cloudwatch_logs": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[cloudWatchLogsModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrEnabled: schema.BoolAttribute{
										Required: true,
									},
									"log_group": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
						"firehose": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[firehoseModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"delivery_stream": schema.StringAttribute{
										Optional: true,
									},
									names.AttrEnabled: schema.BoolAttribute{
										Required: true,
									},
								},
							},
						},
						"s3": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[s3LogModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrBucket: schema.StringAttribute{
										Optional: true,
									},
									names.AttrEnabled: schema.BoolAttribute{
										Required: true,
									},
									names.AttrPrefix: schema.StringAttribute{
										Optional: true,
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

func deadLetterQueueS3NestedObject() schema.NestedBlockObject {
	return schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"bucket_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			"error_output_prefix": schema.StringAttribute{
				Optional: true,
			},
			"expected_bucket_owner": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (r *channelResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("iceberg_destination"),
			path.MatchRoot("s3_destination"),
		),
	}
}

func (r *channelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().KafkaClient(ctx)

	var plan channelResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	channelName := fwflex.StringValueFromFramework(ctx, plan.ChannelName)
	var input kafka.CreateChannelInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateChannel(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, channelName)
		return
	}

	channelARN := aws.ToString(output.ChannelArn)
	clusterARN := fwflex.StringValueFromFramework(ctx, plan.ClusterARN)

	out, err := waitChannelCreated(ctx, conn, channelARN, clusterARN, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, channelARN)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ChannelARN = fwflex.StringToFramework(ctx, output.ChannelArn)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *channelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().KafkaClient(ctx)

	var state channelResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	channelARN, clusterARN := fwflex.StringValueFromFramework(ctx, state.ChannelARN), fwflex.StringValueFromFramework(ctx, state.ClusterARN)
	out, err := findChannelByTwoPartKey(ctx, conn, channelARN, clusterARN)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, channelARN)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *channelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All non-tag attributes force replacement; tag changes are handled by the
	// transparent tagging interceptor. Persist the plan to state.
	var plan channelResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *channelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().KafkaClient(ctx)

	var state channelResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	channelARN, clusterARN := fwflex.StringValueFromFramework(ctx, state.ChannelARN), fwflex.StringValueFromFramework(ctx, state.ClusterARN)
	input := kafka.DeleteChannelInput{
		ChannelArn: aws.String(channelARN),
		ClusterArn: aws.String(clusterARN),
	}
	_, err := conn.DeleteChannel(ctx, &input)
	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, channelARN)
		return
	}

	if _, err := waitChannelDeleted(ctx, conn, channelARN, clusterARN, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, channelARN)
		return
	}
}

func waitChannelCreated(ctx context.Context, conn *kafka.Client, channelARN, clusterARN string, timeout time.Duration) (*kafka.DescribeChannelOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ChannelStatusCreating),
		Target:                    enum.Slice(awstypes.ChannelStatusActive),
		Refresh:                   statusChannel(conn, channelARN, clusterARN),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*kafka.DescribeChannelOutput); ok {
		return out, err
	}

	return nil, err
}

func waitChannelDeleted(ctx context.Context, conn *kafka.Client, channelARN, clusterARN string, timeout time.Duration) (*kafka.DescribeChannelOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ChannelStatusDeleting, awstypes.ChannelStatusActive),
		Target:  []string{},
		Refresh: statusChannel(conn, channelARN, clusterARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*kafka.DescribeChannelOutput); ok {
		return out, err
	}

	return nil, err
}

func statusChannel(conn *kafka.Client, channelARN, clusterARN string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findChannelByTwoPartKey(ctx, conn, channelARN, clusterARN)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findChannelByTwoPartKey(ctx context.Context, conn *kafka.Client, channelARN, clusterARN string) (*kafka.DescribeChannelOutput, error) {
	input := kafka.DescribeChannelInput{
		ChannelArn: aws.String(channelARN),
		ClusterArn: aws.String(clusterARN),
	}

	output, err := conn.DescribeChannel(ctx, &input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ChannelArn == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type channelResourceModel struct {
	framework.WithRegionModel
	ChannelARN                      types.String                                             `tfsdk:"arn"`
	ChannelName                     types.String                                             `tfsdk:"channel_name"`
	ClusterARN                      fwtypes.ARN                                              `tfsdk:"cluster_arn"`
	ClusterOperationARN             types.String                                             `tfsdk:"cluster_operation_arn"`
	CreationTime                    timetypes.RFC3339                                        `tfsdk:"creation_time"`
	DestinationType                 fwtypes.StringEnum[awstypes.ChannelDestinationType]      `tfsdk:"destination_type"`
	EncryptionConfiguration         fwtypes.ListNestedObjectValueOf[encryptionConfigModel]   `tfsdk:"encryption_configuration"`
	IcebergDestinationConfiguration fwtypes.ListNestedObjectValueOf[icebergDestinationModel] `tfsdk:"iceberg_destination"`
	LoggingInfo                     fwtypes.ListNestedObjectValueOf[channelLoggingInfoModel] `tfsdk:"logging_info"`
	S3DestinationConfiguration      fwtypes.ListNestedObjectValueOf[s3DestinationModel]      `tfsdk:"s3_destination"`
	Status                          fwtypes.StringEnum[awstypes.ChannelStatus]               `tfsdk:"status"`
	Tags                            tftags.Map                                               `tfsdk:"tags"`
	TagsAll                         tftags.Map                                               `tfsdk:"tags_all"`
	Timeouts                        timeouts.Value                                           `tfsdk:"timeouts"`
	TopicConfigurationList          fwtypes.ListNestedObjectValueOf[topicConfigurationModel] `tfsdk:"topic_configuration"`
}

type topicConfigurationModel struct {
	RecordConverter fwtypes.ListNestedObjectValueOf[recordConverterModel] `tfsdk:"record_converter"`
	RecordSchema    fwtypes.ListNestedObjectValueOf[recordSchemaModel]    `tfsdk:"record_schema"`
	TopicARN        fwtypes.ARN                                           `tfsdk:"topic_arn"`
}

type recordConverterModel struct {
	ValueConverter fwtypes.StringEnum[awstypes.ValueConverter] `tfsdk:"value_converter"`
}

type recordSchemaModel struct {
	GSRARN fwtypes.ARN `tfsdk:"gsr_arn"`
}

type icebergDestinationModel struct {
	AppendOnly              types.Bool                                              `tfsdk:"append_only"`
	Catalog                 fwtypes.ListNestedObjectValueOf[catalogModel]           `tfsdk:"catalog"`
	CompressionType         fwtypes.StringEnum[awstypes.IcebergCompressionType]     `tfsdk:"compression_type"`
	DataFreshnessInSeconds  types.Int32                                             `tfsdk:"data_freshness_in_seconds"`
	DeadLetterQueueS3       fwtypes.ListNestedObjectValueOf[deadLetterQueueS3Model] `tfsdk:"dead_letter_queue_s3"`
	DestinationTableList    fwtypes.ListNestedObjectValueOf[destinationTableModel]  `tfsdk:"destination_table"`
	SchemaEvolution         fwtypes.ListNestedObjectValueOf[schemaEvolutionModel]   `tfsdk:"schema_evolution"`
	ServiceExecutionRoleARN fwtypes.ARN                                             `tfsdk:"service_execution_role_arn"`
	TableCreation           fwtypes.ListNestedObjectValueOf[tableCreationModel]     `tfsdk:"table_creation"`
}

type deadLetterQueueS3Model struct {
	BucketARN           fwtypes.ARN  `tfsdk:"bucket_arn"`
	ErrorOutputPrefix   types.String `tfsdk:"error_output_prefix"`
	ExpectedBucketOwner types.String `tfsdk:"expected_bucket_owner"`
}

type destinationTableModel struct {
	DestinationDatabaseName types.String                                        `tfsdk:"destination_database_name"`
	DestinationTableName    types.String                                        `tfsdk:"destination_table_name"`
	PartitionSpec           fwtypes.ListNestedObjectValueOf[partitionSpecModel] `tfsdk:"partition_spec"`
}

type partitionSpecModel struct {
	PartitionStrategy fwtypes.StringEnum[awstypes.PartitionStrategy]        `tfsdk:"partition_strategy"`
	SourceList        fwtypes.ListNestedObjectValueOf[partitionSourceModel] `tfsdk:"source"`
}

type partitionSourceModel struct {
	SourceName types.String `tfsdk:"source_name"`
}

type schemaEvolutionModel struct {
	EnableSchemaEvolution types.Bool `tfsdk:"enable_schema_evolution"`
}

type tableCreationModel struct {
	EnableTableCreation types.Bool `tfsdk:"enable_table_creation"`
}

type catalogModel struct {
	CatalogARN        fwtypes.ARN `tfsdk:"catalog_arn"`
	WarehouseLocation fwtypes.ARN `tfsdk:"warehouse_location"`
}

type s3DestinationModel struct {
	DataFreshnessInSeconds  types.Int32                                             `tfsdk:"data_freshness_in_seconds"`
	DeadLetterQueueS3       fwtypes.ListNestedObjectValueOf[deadLetterQueueS3Model] `tfsdk:"dead_letter_queue_s3"`
	ServiceExecutionRoleARN fwtypes.ARN                                             `tfsdk:"service_execution_role_arn"`
	Storage                 fwtypes.ListNestedObjectValueOf[s3StorageModel]         `tfsdk:"storage"`
}

type s3StorageModel struct {
	BucketARN           fwtypes.ARN                                    `tfsdk:"bucket_arn"`
	CompressionType     fwtypes.StringEnum[awstypes.S3CompressionType] `tfsdk:"compression_type"`
	ExpectedBucketOwner types.String                                   `tfsdk:"expected_bucket_owner"`
	OutputKeyTemplate   types.String                                   `tfsdk:"output_key_template"`
	OutputPrefix        types.String                                   `tfsdk:"output_prefix"`
	StorageClass        fwtypes.StringEnum[awstypes.S3StorageClass]    `tfsdk:"storage_class"`
}

type encryptionConfigModel struct {
	KMSKeyARN fwtypes.ARN `tfsdk:"kms_key_arn"`
}

type channelLoggingInfoModel struct {
	CloudWatchLogs fwtypes.ListNestedObjectValueOf[cloudWatchLogsModel] `tfsdk:"cloudwatch_logs"`
	Firehose       fwtypes.ListNestedObjectValueOf[firehoseModel]       `tfsdk:"firehose"`
	S3             fwtypes.ListNestedObjectValueOf[s3LogModel]          `tfsdk:"s3"`
}

type cloudWatchLogsModel struct {
	Enabled  types.Bool   `tfsdk:"enabled"`
	LogGroup types.String `tfsdk:"log_group"`
}

type firehoseModel struct {
	DeliveryStream types.String `tfsdk:"delivery_stream"`
	Enabled        types.Bool   `tfsdk:"enabled"`
}

type s3LogModel struct {
	Bucket  types.String `tfsdk:"bucket"`
	Enabled types.Bool   `tfsdk:"enabled"`
	Prefix  types.String `tfsdk:"prefix"`
}

var _ inttypes.ImportIDParser = channelImportID{}

type channelImportID struct{}

func (channelImportID) Parse(id string) (string, map[string]any, error) {
	const (
		channelIDParts = 2
	)
	parts, err := intflex.ExpandResourceId(id, channelIDParts, true)
	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		names.AttrARN: parts[0],
		"cluster_arn": parts[1],
	}

	return id, result, nil
}
