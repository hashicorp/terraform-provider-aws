// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package kinesis

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	flex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfstringvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/stringvalidator"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_kinesis_channel", name="Channel")
// @Tags(identifierAttribute="channel_arn", resourceType="Channel")
// @ArnIdentity("channel_arn")
// @Testing(hasNoPreExistingResource=true)
func newChannelResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &channelResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNameChannel = "Channel"

	channelDataFreshnessInSecondsDefault = 300
	channelDataFreshnessInSecondsMin     = 300
	channelDataFreshnessInSecondsMax     = 900
)

type channelResource struct {
	framework.ResourceWithModel[channelResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *channelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"channel_arn": framework.ARNAttributeComputedOnly(),
			"channel_id":  framework.IDAttribute(),
			"channel_creation_timestamp": schema.StringAttribute{
				Description: "The time at which the channel was created.",
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"channel_status": schema.StringAttribute{
				Description: "The current status of the channel.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"channel_status_reason": schema.StringAttribute{
				Description: "A message describing the reason for a FAILED status.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),

			"channel_name": schema.StringAttribute{
				Description: "The name of the channel. The name is unique within your Amazon Web Services account and Amazon Web Services Region.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
					stringvalidator.RegexMatches(
						regexache.MustCompile(`[a-zA-Z0-9_.-]+`),
						"value must contain only letters, numbers, hyphens and underscores.",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"service_execution_role_arn": schema.StringAttribute{
				Description: "The Amazon Resource Name (ARN) of the IAM role that Amazon Kinesis Data Streams assumes to write records to the destination.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 512),
					stringvalidator.RegexMatches(
						regexache.MustCompile(`arn:aws[-a-z0-9]*:iam::\d{12}:role/[a-zA-Z_0-9+=,.@\-_/]+`),
						"value must be an IAM role arn.",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"stream_configuration_list": schema.ListNestedBlock{
				Description: "The source stream configuration for the channel. Currently, one stream is supported per channel.",
				CustomType:  fwtypes.NewListNestedObjectTypeOf[channelSteamConfigurationListModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1000), // Currently, one stream is supported per channel but the max constraint is 1000
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"stream_arn": schema.StringAttribute{
							Description: "The Amazon Resource Name (ARN) of the source Kinesis data stream.",
							Required:    true,
							CustomType:  fwtypes.ARNType,
						},
						"stream_creation_timestamp": schema.StringAttribute{
							Description: "The time at which the source stream was created.",
							CustomType:  timetypes.RFC3339Type{},
							Computed:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"record_configuration": channelRecordConfigurationBlock(ctx),
					},
				},
			},

			"s3_destination_configuration": schema.ListNestedBlock{
				Description: "The configuration for delivery to a general purpose Amazon S3 bucket. Present only when the channel destination is a general purpose Amazon S3 bucket.",
				CustomType:  fwtypes.NewListNestedObjectTypeOf[channelS3DestinationConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.ExactlyOneOf(
						path.MatchRoot("s3_destination_configuration"),
						path.MatchRoot("s3_tables_destination_configuration"),
					),
					listvalidator.ConflictsWith(path.MatchRoot("s3_tables_destination_configuration")),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"data_freshness_in_seconds": schema.Int64Attribute{
							Description: "The maximum age, in seconds, of undelivered data.",
							Optional:    true,
							Default:     int64default.StaticInt64(channelDataFreshnessInSecondsDefault),
							Computed:    true,
							Validators: []validator.Int64{
								int64validator.AtLeast(channelDataFreshnessInSecondsMin),
								int64validator.AtMost(channelDataFreshnessInSecondsMax),
							},
						},
						"dead_letter_queue_s3_configuration": channelDeadLetterQueueS3ConfigurationAttribute(ctx),
					},
					Blocks: map[string]schema.Block{
						"storage_configuration": channelStorageConfigurationBlock(ctx),
					},
				},
			},

			"s3_tables_destination_configuration": schema.ListNestedBlock{
				Description: "The configuration for delivery to streaming tables on Apache Iceberg in Amazon S3 Tables. Present only when the channel destination is a streaming table.",
				CustomType:  fwtypes.NewListNestedObjectTypeOf[channelS3TablesDestinationConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.ExactlyOneOf(
						path.MatchRoot("s3_destination_configuration"),
						path.MatchRoot("s3_tables_destination_configuration"),
					),
					listvalidator.ConflictsWith(path.MatchRoot("s3_destination_configuration")),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"data_freshness_in_seconds": schema.Int64Attribute{
							Description: "The maximum age, in seconds, of undelivered data.",
							Optional:    true,
							Default:     int64default.StaticInt64(channelDataFreshnessInSecondsDefault),
							Computed:    true,
							Validators: []validator.Int64{
								int64validator.AtLeast(channelDataFreshnessInSecondsMin),
								int64validator.AtMost(channelDataFreshnessInSecondsMax),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"dead_letter_queue_s3_configuration": channelDeadLetterQueueS3ConfigurationBlock(ctx),
						"s3_tables_configuration_list":       channelS3TablesConfigurationListBlock(ctx),
					},
				},
			},

			"encryption_configuration": schema.ListNestedBlock{
				Description: "The server-side encryption configuration for the channel.",
				CustomType:  fwtypes.NewListNestedObjectTypeOf[channelEncryptionConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"encryption_type": schema.StringAttribute{
							Description: "The encryption type. The only valid value is KMS.",
							CustomType:  fwtypes.StringEnumType[awstypes.EncryptionType](),
							Required:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"key_id": schema.StringAttribute{
							Description: "The identifier of the customer managed Amazon Web Services KMS key. You cannot use the Amazon Kinesis Data Streams service key (aws/kinesis).",
							CustomType:  types.StringType,
							Required:    true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
								stringvalidator.LengthAtMost(2048),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},

			"logging_configuration": schema.ListNestedBlock{
				Description: "The Amazon CloudWatch Logs configuration for the channel.",
				CustomType:  fwtypes.NewListNestedObjectTypeOf[channelLoggingConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"cloudwatch_logs": schema.ListNestedBlock{
							Description: "The Amazon CloudWatch Logs settings for the channel.",
							CustomType:  fwtypes.NewListNestedObjectTypeOf[cloudWatchLogsModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
								listvalidator.IsRequired(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"enabled": schema.BoolAttribute{
										Description: "Specifies whether logging to Amazon CloudWatch Logs is enabled.",
										CustomType:  types.BoolType,
										Required:    true,
									},
									"log_group_name": schema.StringAttribute{
										Description: "The name of the Amazon CloudWatch Logs log group. Defaults to /aws/kinesis/{channelName}/{channelId}.",
										CustomType:  types.StringType,
										Optional:    true,
										Computed:    true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
										Validators: []validator.String{
											stringvalidator.LengthAtLeast(1),
											stringvalidator.LengthAtMost(512),
											stringvalidator.RegexMatches(
												regexache.MustCompile(`[\.\-_/#A-Za-z0-9]+`),
												"value must contain alphanumerics, hipens, forward and backward slashes, period and underscores only",
											),
										},
									},
									"log_stream_name": schema.StringAttribute{
										Description: "The name of the Amazon CloudWatch Logs log stream. Defaults to DestinationDelivery.",
										CustomType:  types.StringType,
										Optional:    true,
										Computed:    true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
										Validators: []validator.String{
											stringvalidator.LengthAtLeast(1),
											stringvalidator.LengthAtMost(512),
											stringvalidator.RegexMatches(
												regexache.MustCompile(`[^:*]*`),
												"value cannot contain colons or asterisks",
											),
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

func channelS3TablesConfigurationListBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The list of streaming table configurations.",
		CustomType:  fwtypes.NewListNestedObjectTypeOf[s3TablesConfigurationListModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
			listvalidator.SizeAtMost(10000),
			listvalidator.IsRequired(),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"table_bucket_arn": schema.StringAttribute{
					Description: "The Amazon Resource Name (ARN) of the Amazon S3 table bucket.",
					CustomType:  fwtypes.ARNType,
					Required:    true,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
						stringvalidator.LengthAtMost(2048),
						stringvalidator.RegexMatches(
							regexache.MustCompile(`arn:aws[-a-z0-9]*:s3tables:[-a-z0-9]+:\d{12}:bucket/[a-z0-9_-]{3,63}`),
							"value must contain only valid characters for an S3 table arn",
						),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"namespace": schema.StringAttribute{
					Description: "The namespace (database) of the destination table.",
					CustomType:  types.StringType,
					Required:    true,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
						stringvalidator.LengthAtMost(255),
						stringvalidator.RegexMatches(
							regexache.MustCompile(`[0-9a-z_]+`),
							"value must contain only letters, numbers, hipens and underscore",
						),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"table_name": schema.StringAttribute{
					Description: "The name of the destination table. Amazon Kinesis Data Streams creates this table in the specified table bucket.",
					CustomType:  types.StringType,
					Required:    true,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
						stringvalidator.LengthAtMost(255),
						stringvalidator.RegexMatches(
							regexache.MustCompile(`[0-9a-z_]+`),
							"value must contain only letters, numbers, hipens and underscore",
						),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"compression_type": schema.StringAttribute{
					Description: "The compression applied to Parquet data files.",
					CustomType:  fwtypes.StringEnumType[awstypes.S3CompressionType](),
					Required:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"partition_spec": channelPartitionSpecBlock(ctx),
			},
		},
	}
}

func channelPartitionSpecBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The partitioning specification for the destination table.",
		CustomType:  fwtypes.NewListNestedObjectTypeOf[partitionSpecModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"partition_fields": channelPartitionFieldsBlock(ctx),
			},
		},
	}
}

func channelPartitionFieldsBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The list of partition fields.",
		CustomType:  fwtypes.NewListNestedObjectTypeOf[partitionFieldModel](ctx),
		Validators: []validator.List{
			listvalidator.IsRequired(),
			listvalidator.SizeAtLeast(1),
			listvalidator.SizeAtMost(10),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"source_name": schema.StringAttribute{
					Description: "The name of the source column used for partitioning. This column must be of the timestamptz type.",
					CustomType:  types.StringType,
					Required:    true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 255),
						stringvalidator.RegexMatches(
							regexache.MustCompile(`^[a-zA-Z0-9._]+$`),
							"value must contain only alphanumeric characters, dots, and underscores",
						),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"transform": schema.StringAttribute{
					Description: "The partition transform to apply. The only valid value is TIME_HOUR.",
					CustomType:  fwtypes.StringEnumType[awstypes.PartitionTransform](),
					Required:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func channelStorageConfigurationBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The Amazon S3 storage configuration for the channel.",
		CustomType:  fwtypes.NewListNestedObjectTypeOf[storageConfigurationModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
			listvalidator.IsRequired(),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"bucket_arn": schema.StringAttribute{
					Description: "The Amazon Resource Name (ARN) of the destination Amazon S3 bucket.",
					CustomType:  fwtypes.ARNType,
					Required:    true,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
						stringvalidator.LengthAtMost(2048),
						stringvalidator.RegexMatches(
							regexache.MustCompile(`arn:aws[-a-z0-9]*:s3:::[a-z0-9._-]{3,63}`),
							"value must contain only valid characters for an S3 bucket arn",
						),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"expected_bucket_owner": schema.StringAttribute{
					Description: "The Amazon Web Services account ID of the expected owner of the destination bucket. This value helps prevent delivery to an unintended bucket if ownership changes.",
					CustomType:  types.StringType,
					Required:    true,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(12),
						stringvalidator.LengthAtMost(12),
						stringvalidator.RegexMatches(
							regexache.MustCompile(`\d{12}`),
							"value must contain only valid characters for an aws account number",
						),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"output_key_template": schema.StringAttribute{
					Description: "The template used to construct the Amazon S3 object key for delivered objects. If not specified, a default template is used.",
					CustomType:  types.StringType,
					Optional:    true,
					Computed:    true,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
						stringvalidator.LengthAtMost(1024),
						stringvalidator.RegexMatches(
							regexache.MustCompile(`^[0-9A-Za-z!_'.*()/=:{} -]+$`),
							"value must contain only valid characters for an S3 output key template",
						),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"storage_class": schema.StringAttribute{
					Description: "The Amazon S3 storage class for delivered objects. Defaults to STANDARD.",
					CustomType:  fwtypes.StringEnumType[awstypes.S3StorageClass](),
					Optional:    true,
					Computed:    true,
					Default:     fwtypes.StringEnumType[awstypes.S3StorageClass]().AttributeDefault(awstypes.S3StorageClassStandard),
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"compression_type": schema.StringAttribute{
					Description: "The compression applied to delivered objects.",
					CustomType:  fwtypes.StringEnumType[awstypes.S3CompressionType](),
					Required:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func channelDeadLetterQueueS3ConfigurationAttribute(ctx context.Context) schema.ListAttribute {
	attr := framework.ResourceOptionalComputedForceNewSingleNestedObjectAttribute[deadLetterQueueS3ConfigurationModel](ctx)
	attr.Description = "The dead-letter queue configuration for records that cannot be delivered. If not specified, defaults to the destination bucket with an error prefix."
	return attr
}

func channelDeadLetterQueueS3ConfigurationBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The dead-letter queue configuration for records that cannot be delivered.",
		CustomType:  fwtypes.NewListNestedObjectTypeOf[deadLetterQueueS3ConfigurationModel](ctx),
		Validators: []validator.List{
			listvalidator.IsRequired(),
			listvalidator.SizeAtLeast(1),
			listvalidator.SizeAtMost(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"bucket_arn": schema.StringAttribute{
					Description: "The Amazon Resource Name (ARN) of the dead-letter queue Amazon S3 bucket.",
					CustomType:  fwtypes.ARNType,
					Required:    true,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
						stringvalidator.LengthAtMost(2048),
						stringvalidator.RegexMatches(
							regexache.MustCompile(`arn:aws[-a-z0-9]*:s3:::[a-z0-9._-]{3,63}`),
							"value must contain only valid characters for an S3 bucket arn",
						),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"expected_bucket_owner": schema.StringAttribute{
					Description: "The Amazon Web Services account ID of the expected owner of the dead-letter queue bucket.",
					CustomType:  types.StringType,
					Required:    true,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(12),
						stringvalidator.LengthAtMost(12),
						stringvalidator.RegexMatches(
							regexache.MustCompile(`\d{12}`),
							"value must contain only valid characters for an aws account number",
						),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"error_output_prefix": schema.StringAttribute{
					Description: "The Amazon S3 key prefix for error records.",
					CustomType:  types.StringType,
					Optional:    true,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
						stringvalidator.LengthAtMost(512),
						stringvalidator.RegexMatches(
							regexache.MustCompile(`[0-9A-Za-z!\-_'.*()\/]+`),
							"value must contain only valid characters for an S3 bucket prefix",
						),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func channelRecordConfigurationBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The record format configuration for the source stream.",
		CustomType:  fwtypes.NewListNestedObjectTypeOf[recordConfigurationModel](ctx),
		Validators: []validator.List{
			listvalidator.IsRequired(),
			listvalidator.SizeAtLeast(1),
			listvalidator.SizeAtMost(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"record_format_type": schema.StringAttribute{
					Description: "The format of records on the source stream. Valid values are GSR_JSON, JSON, STRING, or BYTE_ARRAY. Note that GSR_JSON is supported only for S3 Tables destinations, while STRING and BYTE_ARRAY are supported only for general-purpose S3 destinations.",
					CustomType:  fwtypes.StringEnumType[awstypes.RecordFormatType](),
					Required:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
					Validators: []validator.String{
						// GSR_JSON cannot be used with regular S3 destinations
						tfstringvalidator.ConflictsWithWhenEquals(
							awstypes.RecordFormatTypeGsrJson,
							path.MatchRoot("s3_destination_configuration"),
						),
						// STRING cannot be used with S3 Tables, and cannot have a Glue Schema Registry ARN
						tfstringvalidator.ConflictsWithWhenEquals(
							awstypes.RecordFormatTypeString,
							path.MatchRoot("s3_tables_destination_configuration"),
							path.MatchRelative().AtParent().AtName("gsr_schema_arn"),
						),
						// BYTE_ARRAY cannot be used with S3 Tables, and cannot have a Glue Schema Registry ARN
						tfstringvalidator.ConflictsWithWhenEquals(
							awstypes.RecordFormatTypeByteArray,
							path.MatchRoot("s3_tables_destination_configuration"),
							path.MatchRelative().AtParent().AtName("gsr_schema_arn"),
						),
						// GSR_JSON always requires gsr_schema_arn
						tfstringvalidator.AlsoRequiresWhenEquals(
							awstypes.RecordFormatTypeGsrJson,
							path.MatchRelative().AtParent().AtName("gsr_schema_arn"),
						),
					},
				},
				"gsr_schema_arn": schema.StringAttribute{
					Description: "The Amazon Resource Name (ARN) of the Amazon Web Services Glue Schema Registry schema used to validate records. Required when the channel destination is a streaming table (Amazon S3 Tables), for both the JSON and GSR_JSON record formats.",
					CustomType:  fwtypes.ARNType,
					Optional:    true,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
						stringvalidator.LengthAtMost(512),
						stringvalidator.RegexMatches(
							regexache.MustCompile(`arn:aws[-a-z0-9]*:glue:[-a-z0-9]+:\d{12}:schema/[-a-zA-Z0-9_$#.]+/[-a-zA-Z0-9_$#.]+`),
							"value must contain only valid characters for a Glue Schema Registry arn",
						),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func (r *channelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().KinesisClient(ctx)

	var plan channelResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input kinesis.CreateChannelInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("Channel")))
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsInMap(ctx)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	out, err := tfresource.RetryWhen(
		ctx, createTimeout,
		func(ctx context.Context) (*kinesis.CreateChannelOutput, error) {
			return conn.CreateChannel(ctx, &input)
		},
		func(err error) (bool, error) {
			if errs.IsA[*awstypes.ValidationException](err) {
				if errs.Contains(err, "Unable to access the S3 bucket") ||
					errs.Contains(err, "service execution role") ||
					errs.Contains(err, "AccessDenied") {
					return true, err
				}
			}
			if errs.IsA[*awstypes.AccessDeniedException](err) {
				return true, err
			}
			return false, err
		},
	)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil || out.ChannelDescription == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	arn := aws.ToString(out.ChannelDescription.ChannelARN)

	channel, err := waitChannelCreated(ctx, conn, arn, createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, channel, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *channelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().KinesisClient(ctx)

	var state channelResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findChannelByArn(ctx, conn, state.ARN.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.Name, state.Name.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *channelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().KinesisClient(ctx)

	var plan, state channelResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		hasUpdate := !plan.LoggingConfiguration.Equal(state.LoggingConfiguration) ||
			!plan.S3DestinationConfiguration.Equal(state.S3DestinationConfiguration) ||
			!plan.S3TablesDestinationConfiguration.Equal(state.S3TablesDestinationConfiguration)

		if hasUpdate {
			var input kinesis.UpdateChannelInput
			smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("Channel")))
			if resp.Diagnostics.HasError() {
				return
			}
			input.ChannelARN = plan.ARN.ValueStringPointer()

			if plan.LoggingConfiguration.IsNull() && !state.LoggingConfiguration.IsNull() {
				input.LoggingConfiguration = &awstypes.ChannelLoggingUpdateInput{
					CloudWatchLogs: &awstypes.CloudWatchLogsUpdateInput{
						Enabled: aws.Bool(false),
					},
				}
			}

			out, err := conn.UpdateChannel(ctx, &input)
			if err != nil {
				smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ARN.ValueString())
				return
			}
			if out == nil || out.ChannelDescription == nil {
				smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.ARN.ValueString())
				return
			}

			updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
			updated, err := waitChannelUpdated(ctx, conn, plan.ARN.ValueString(), updateTimeout)
			if err != nil {
				smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ARN.ValueString())
				return
			}

			smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, updated, &plan))
			if resp.Diagnostics.HasError() {
				return
			}
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *channelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().KinesisClient(ctx)

	var state channelResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := kinesis.DeleteChannelInput{
		ChannelARN: state.ARN.ValueStringPointer(),
	}

	_, err := conn.DeleteChannel(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.ValueString())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitChannelDeleted(ctx, conn, state.ARN.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.ValueString())
		return
	}
}

func (r *channelResource) flatten(ctx context.Context, channel *awstypes.ChannelDescription, data *channelResourceModel) (diags diag.Diagnostics) {
	hadLoggingConfig := data.LoggingConfiguration.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0

	diags.Append(flex.Flatten(ctx, channel, data, flex.WithFieldNamePrefix("Channel"))...)
	data.ARN = types.StringPointerValue(channel.ChannelARN)
	data.ID = types.StringPointerValue(channel.ChannelId)
	data.Name = types.StringPointerValue(channel.ChannelName)

	if !hadLoggingConfig && (channel.LoggingConfiguration == nil || channel.LoggingConfiguration.CloudWatchLogs == nil || !aws.ToBool(channel.LoggingConfiguration.CloudWatchLogs.Enabled)) {
		data.LoggingConfiguration = fwtypes.NewListNestedObjectValueOfNull[channelLoggingConfigurationModel](ctx)
	}

	return diags
}

func getTagsInMap(ctx context.Context) map[string]string {
	if inContext, ok := tftags.FromContext(ctx); ok {
		if tags := inContext.TagsIn.UnwrapOrDefault().Map(); len(tags) > 0 {
			return tags
		}
	}
	return nil
}

func waitChannelCreated(ctx context.Context, conn *kinesis.Client, arn string, timeout time.Duration) (*awstypes.ChannelDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ChannelStatusCreating),
		Target:                    enum.Slice(awstypes.ChannelStatusActive),
		Refresh:                   statusChannel(conn, arn),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.ChannelDescription); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitChannelUpdated(ctx context.Context, conn *kinesis.Client, arn string, timeout time.Duration) (*awstypes.ChannelDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ChannelStatusUpdating),
		Target:                    enum.Slice(awstypes.ChannelStatusActive),
		Refresh:                   statusChannel(conn, arn),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.ChannelDescription); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitChannelDeleted(ctx context.Context, conn *kinesis.Client, arn string, timeout time.Duration) (*awstypes.ChannelDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ChannelStatusDeleting),
		Target:  []string{},
		Refresh: statusChannel(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.ChannelDescription); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusChannel(conn *kinesis.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findChannelByArn(ctx, conn, arn)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.ChannelStatus), nil
	}
}

func findChannelByArn(ctx context.Context, conn *kinesis.Client, arn string) (*awstypes.ChannelDescription, error) {
	input := kinesis.DescribeChannelInput{
		ChannelARN: aws.String(arn),
	}

	out, err := conn.DescribeChannel(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.ChannelDescription == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.ChannelDescription, nil
}

type channelResourceModel struct {
	framework.WithRegionModel
	ARN                              types.String                                                                  `tfsdk:"channel_arn"`
	ID                               types.String                                                                  `tfsdk:"channel_id"`
	Name                             types.String                                                                  `tfsdk:"channel_name"`
	ServiceExecutionRoleARN          types.String                                                                  `tfsdk:"service_execution_role_arn"`
	Status                           types.String                                                                  `tfsdk:"channel_status"`
	StatusReason                     types.String                                                                  `tfsdk:"channel_status_reason"`
	CreationTimestamp                timetypes.RFC3339                                                             `tfsdk:"channel_creation_timestamp"`
	StreamConfigurationList          fwtypes.ListNestedObjectValueOf[channelSteamConfigurationListModel]           `tfsdk:"stream_configuration_list"`
	S3DestinationConfiguration       fwtypes.ListNestedObjectValueOf[channelS3DestinationConfigurationModel]       `tfsdk:"s3_destination_configuration"`
	S3TablesDestinationConfiguration fwtypes.ListNestedObjectValueOf[channelS3TablesDestinationConfigurationModel] `tfsdk:"s3_tables_destination_configuration"`
	EncryptionConfiguration          fwtypes.ListNestedObjectValueOf[channelEncryptionConfigurationModel]          `tfsdk:"encryption_configuration"`
	LoggingConfiguration             fwtypes.ListNestedObjectValueOf[channelLoggingConfigurationModel]             `tfsdk:"logging_configuration"`
	Tags                             tftags.Map                                                                    `tfsdk:"tags"`
	TagsAll                          tftags.Map                                                                    `tfsdk:"tags_all"`
	Timeouts                         timeouts.Value                                                                `tfsdk:"timeouts"`
}

type channelSteamConfigurationListModel struct {
	StreamARN               fwtypes.ARN                                               `tfsdk:"stream_arn"`
	StreamCreationTimestamp timetypes.RFC3339                                         `tfsdk:"stream_creation_timestamp"`
	RecordConfiguration     fwtypes.ListNestedObjectValueOf[recordConfigurationModel] `tfsdk:"record_configuration"`
}

type recordConfigurationModel struct {
	RecordFormatType fwtypes.StringEnum[awstypes.RecordFormatType] `tfsdk:"record_format_type"`
	GSRSchemaARN     fwtypes.ARN                                   `tfsdk:"gsr_schema_arn"`
}

type channelS3DestinationConfigurationModel struct {
	DataFreshnessInSeconds         types.Int64                                                          `tfsdk:"data_freshness_in_seconds"`
	DeadLetterQueueS3Configuration fwtypes.ListNestedObjectValueOf[deadLetterQueueS3ConfigurationModel] `tfsdk:"dead_letter_queue_s3_configuration"`
	StorageConfiguration           fwtypes.ListNestedObjectValueOf[storageConfigurationModel]           `tfsdk:"storage_configuration"`
}

type deadLetterQueueS3ConfigurationModel struct {
	BucketARN           fwtypes.ARN  `tfsdk:"bucket_arn"`
	ExpectedBucketOwner types.String `tfsdk:"expected_bucket_owner"`
	ErrorOutputPrefix   types.String `tfsdk:"error_output_prefix"`
}

type storageConfigurationModel struct {
	BucketARN           fwtypes.ARN                                    `tfsdk:"bucket_arn"`
	ExpectedBucketOwner types.String                                   `tfsdk:"expected_bucket_owner"`
	OutputKeyTemplate   types.String                                   `tfsdk:"output_key_template"`
	StorageClass        fwtypes.StringEnum[awstypes.S3StorageClass]    `tfsdk:"storage_class"`
	CompressionType     fwtypes.StringEnum[awstypes.S3CompressionType] `tfsdk:"compression_type"`
}

type channelS3TablesDestinationConfigurationModel struct {
	DataFreshnessInSeconds         types.Int64                                                          `tfsdk:"data_freshness_in_seconds"`
	DeadLetterQueueS3Configuration fwtypes.ListNestedObjectValueOf[deadLetterQueueS3ConfigurationModel] `tfsdk:"dead_letter_queue_s3_configuration"`
	S3TablesConfigurationList      fwtypes.ListNestedObjectValueOf[s3TablesConfigurationListModel]      `tfsdk:"s3_tables_configuration_list"`
}

type s3TablesConfigurationListModel struct {
	TableBucketARN  fwtypes.ARN                                         `tfsdk:"table_bucket_arn"`
	Namespace       types.String                                        `tfsdk:"namespace"`
	TableName       types.String                                        `tfsdk:"table_name"`
	CompressionType fwtypes.StringEnum[awstypes.S3CompressionType]      `tfsdk:"compression_type"`
	PartitionSpec   fwtypes.ListNestedObjectValueOf[partitionSpecModel] `tfsdk:"partition_spec"`
}
type partitionSpecModel struct {
	PartitionFields fwtypes.ListNestedObjectValueOf[partitionFieldModel] `tfsdk:"partition_fields"`
}

type partitionFieldModel struct {
	SourceName types.String                                    `tfsdk:"source_name"`
	Transform  fwtypes.StringEnum[awstypes.PartitionTransform] `tfsdk:"transform"`
}

type channelEncryptionConfigurationModel struct {
	EncryptionType fwtypes.StringEnum[awstypes.EncryptionType] `tfsdk:"encryption_type"`
	KeyID          types.String                                `tfsdk:"key_id"`
}

type channelLoggingConfigurationModel struct {
	CloudWatchLogs fwtypes.ListNestedObjectValueOf[cloudWatchLogsModel] `tfsdk:"cloudwatch_logs"`
}

type cloudWatchLogsModel struct {
	Enabled       types.Bool   `tfsdk:"enabled"`
	LogGroupName  types.String `tfsdk:"log_group_name"`
	LogStreamName types.String `tfsdk:"log_stream_name"`
}

func sweepChannels(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := kinesis.ListChannelsInput{}
	conn := client.KinesisClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := kinesis.NewListChannelsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.ChannelSummaries {
			sweepResources = append(
				sweepResources, sweepfw.NewSweepResource(newChannelResource, client,
					sweepfw.NewAttribute("channel_arn", aws.ToString(v.ChannelARN))),
			)
		}
	}

	return sweepResources, nil
}
