// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kafka

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
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
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
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
// @Testing(identityRegionOverrideTest=false)
// @Testing(importStateIdFunc="testAccChannelImportStateIDFunc")
// @Testing(importStateIdAttribute="arn")
func newChannelResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &channelResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type channelResource struct {
	framework.ResourceWithModel[channelResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func deadLetterQueueS3Block(ctx context.Context, extraValidators ...validator.List) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[deadLetterQueueS3Model](ctx),
		Validators: append([]validator.List{
			listvalidator.IsRequired(),
			listvalidator.SizeAtLeast(1),
			listvalidator.SizeAtMost(1),
		}, extraValidators...),
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"bucket_arn": schema.StringAttribute{
					CustomType: fwtypes.ARNType,
					Required:   true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"error_output_prefix": schema.StringAttribute{
					Optional: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				names.AttrExpectedBucketOwner: schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						fwvalidators.AWSAccountID(),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
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
			"destination_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ChannelDestinationType](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrEncryptionConfiguration: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[encryptionConfigurationModel](ctx),
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
			"iceberg_destination": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[icebergDestinationConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"append_only": schema.BoolAttribute{
							Required: true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.RequiresReplace(),
							},
						},
						"compression_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.IcebergCompressionType](),
							Optional:   true,
							Computed:   true,
							// Optional+Computed: force replacement only when the user changes
							// a configured value, not when the server-computed default is
							// refreshed to unknown while another field is updated in place.
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplaceIfConfigured(),
								stringplanmodifier.UseNonNullStateForUnknown(),
							},
						},
						// data_freshness_in_seconds is the only destination field that can
						// be updated in place (via UpdateChannel); all others force replacement.
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
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"catalog": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[catalogModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"catalog_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Optional:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"warehouse_location": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Optional:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
							},
						},
						"dead_letter_queue_s3": deadLetterQueueS3Block(ctx),
						"destination_table": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[destinationTableModel](ctx),
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
									"destination_database_name": schema.StringAttribute{
										Optional: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"destination_table_name": schema.StringAttribute{
										Optional: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"partition_spec": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[partitionSpecModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"partition_strategy": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.PartitionStrategy](),
													Required:   true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
											},
											Blocks: map[string]schema.Block{
												names.AttrSource: schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[partitionSourceModel](ctx),
													PlanModifiers: []planmodifier.List{
														listplanmodifier.RequiresReplace(),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"source_name": schema.StringAttribute{
																Optional: true,
																PlanModifiers: []planmodifier.String{
																	stringplanmodifier.RequiresReplace(),
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
						"schema_evolution": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[schemaEvolutionModel](ctx),
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
									"enable_schema_evolution": schema.BoolAttribute{
										Optional: true,
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.RequiresReplace(),
										},
									},
								},
							},
						},
						"table_creation": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[tableCreationModel](ctx),
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
									"enable_table_creation": schema.BoolAttribute{
										Optional: true,
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.RequiresReplace(),
										},
									},
								},
							},
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
						names.AttrCloudWatchLogs: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[cloudWatchLogsModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrEnabled: schema.BoolAttribute{
										Required: true,
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.RequiresReplace(),
										},
									},
									"log_group": schema.StringAttribute{
										Optional: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
							},
						},
						"firehose": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[firehoseModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"delivery_stream": schema.StringAttribute{
										Optional: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									names.AttrEnabled: schema.BoolAttribute{
										Required: true,
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.RequiresReplace(),
										},
									},
								},
							},
						},
						"s3": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[s3LogModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrBucket: schema.StringAttribute{
										Optional: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									names.AttrEnabled: schema.BoolAttribute{
										Required: true,
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.RequiresReplace(),
										},
									},
									names.AttrPrefix: schema.StringAttribute{
										Optional: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
							},
						},
					},
				},
			},
			"s3_destination": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[s3DestinationConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						// data_freshness_in_seconds is the only destination field that can
						// be updated in place (via UpdateChannel); all others force replacement.
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
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"dead_letter_queue_s3": deadLetterQueueS3Block(ctx),
						"storage": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[s3StorageModel](ctx),
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
										CustomType: fwtypes.ARNType,
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"compression_type": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.S3CompressionType](),
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									names.AttrExpectedBucketOwner: schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											fwvalidators.AWSAccountID(),
										},
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"output_key_template": schema.StringAttribute{
										Optional: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"output_prefix": schema.StringAttribute{
										Optional: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									names.AttrStorageClass: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.S3StorageClass](),
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
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
			"topic_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[topicConfigurationModel](ctx),
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
						names.AttrTopicARN: schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"record_converter": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[recordConverterModel](ctx),
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
									"value_converter": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.ValueConverter](),
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
							},
						},
						"record_schema": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[recordSchemaModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"gsr_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
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

func (r *channelResource) ConfigValidators(context.Context) []resource.ConfigValidator {
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

	outputCC, err := tfresource.RetryWhenIsAErrorMessageContains[*kafka.CreateChannelOutput, *awstypes.ForbiddenException](ctx, propagationTimeout, func(ctx context.Context) (*kafka.CreateChannelOutput, error) {
		return conn.CreateChannel(ctx, &input)
	}, "Unable to assume the channel's service execution role")
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, channelName)
		return
	}

	channelARN, clusterARN := aws.ToString(outputCC.ChannelArn), fwflex.StringValueFromFramework(ctx, plan.ClusterARN)
	outputGC, err := waitChannelCreated(ctx, conn, channelARN, clusterARN, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		// Taint the resource.
		resp.State.SetAttribute(ctx, path.Root(names.AttrARN), channelARN)
		resp.State.SetAttribute(ctx, path.Root("cluster_arn"), clusterARN)
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, channelARN)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, outputGC, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

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

	setTagsOut(ctx, out.Tags)

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *channelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().KafkaClient(ctx)

	var plan, state channelResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		// data_freshness_in_seconds is the only configurable attribute that can be
		// updated in place (via UpdateChannel). Tag-only changes are handled by the
		// transparent tagging interceptor; every other attribute forces replacement.
		channelARN, clusterARN := fwflex.StringValueFromFramework(ctx, plan.ChannelARN), fwflex.StringValueFromFramework(ctx, plan.ClusterARN)
		input := kafka.UpdateChannelInput{
			ChannelArn: aws.String(channelARN),
			ClusterArn: aws.String(clusterARN),
		}

		switch {
		case slices.Equal(diff.ChangedFieldNames(), []string{"IcebergDestinationConfiguration"}) && !plan.IcebergDestinationConfiguration.IsNull():
			model, d := plan.IcebergDestinationConfiguration.ToPtr(ctx)
			smerr.AddEnrich(ctx, &resp.Diagnostics, d)
			if resp.Diagnostics.HasError() {
				return
			}

			var destination awstypes.IcebergDestinationConfiguration
			smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, model, &destination))
			if resp.Diagnostics.HasError() {
				return
			}
			input.IcebergDestinationUpdate = &awstypes.IcebergDestinationUpdate{
				DataFreshnessInSeconds: destination.DataFreshnessInSeconds,
			}

		case slices.Equal(diff.ChangedFieldNames(), []string{"S3DestinationConfiguration"}) && !plan.S3DestinationConfiguration.IsNull():
			model, d := plan.S3DestinationConfiguration.ToPtr(ctx)
			smerr.AddEnrich(ctx, &resp.Diagnostics, d)
			if resp.Diagnostics.HasError() {
				return
			}

			var destination awstypes.S3DestinationConfiguration
			smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, model, &destination))
			if resp.Diagnostics.HasError() {
				return
			}
			input.S3DestinationUpdate = &awstypes.S3DestinationUpdate{
				DataFreshnessInSeconds: destination.DataFreshnessInSeconds,
			}
		}

		_, err := conn.UpdateChannel(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, channelARN)
			return
		}

		if _, err := waitChannelUpdated(ctx, conn, channelARN, clusterARN, r.UpdateTimeout(ctx, plan.Timeouts)); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, channelARN)
			return
		}
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

func (r *channelResource) flatten(ctx context.Context, out *kafka.DescribeChannelOutput, data *channelResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.Append(fwflex.Flatten(ctx, out, data)...)
	return diags
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
		if v := out.StateInfo; v != nil {
			retry.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(v.Code), aws.ToString(v.Message)))
		}
		return out, err
	}

	return nil, err
}

func waitChannelUpdated(ctx context.Context, conn *kafka.Client, channelARN, clusterARN string, timeout time.Duration) (*kafka.DescribeChannelOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ChannelStatusUpdating),
		Target:                    enum.Slice(awstypes.ChannelStatusActive),
		Refresh:                   statusChannel(conn, channelARN, clusterARN),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*kafka.DescribeChannelOutput); ok {
		if v := out.StateInfo; v != nil {
			retry.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(v.Code), aws.ToString(v.Message)))
		}
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
		if v := out.StateInfo; v != nil {
			retry.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(v.Code), aws.ToString(v.Message)))
		}
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
	return findChannel(ctx, conn, &input)
}

func findChannel(ctx context.Context, conn *kafka.Client, input *kafka.DescribeChannelInput) (*kafka.DescribeChannelOutput, error) {
	output, err := conn.DescribeChannel(ctx, input)

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
	ChannelARN                      types.String                                                          `tfsdk:"arn"`
	ChannelName                     types.String                                                          `tfsdk:"channel_name"`
	ClusterARN                      fwtypes.ARN                                                           `tfsdk:"cluster_arn"`
	DestinationType                 fwtypes.StringEnum[awstypes.ChannelDestinationType]                   `tfsdk:"destination_type"`
	EncryptionConfiguration         fwtypes.ListNestedObjectValueOf[encryptionConfigurationModel]         `tfsdk:"encryption_configuration"`
	IcebergDestinationConfiguration fwtypes.ListNestedObjectValueOf[icebergDestinationConfigurationModel] `tfsdk:"iceberg_destination"`
	LoggingInfo                     fwtypes.ListNestedObjectValueOf[channelLoggingInfoModel]              `tfsdk:"logging_info"`
	S3DestinationConfiguration      fwtypes.ListNestedObjectValueOf[s3DestinationConfigurationModel]      `tfsdk:"s3_destination"`
	Tags                            tftags.Map                                                            `tfsdk:"tags"`
	TagsAll                         tftags.Map                                                            `tfsdk:"tags_all"`
	Timeouts                        timeouts.Value                                                        `tfsdk:"timeouts"`
	TopicConfigurationList          fwtypes.ListNestedObjectValueOf[topicConfigurationModel]              `tfsdk:"topic_configuration"`
}

type encryptionConfigurationModel struct {
	KMSKeyARN fwtypes.ARN `tfsdk:"kms_key_arn"`
}

type icebergDestinationConfigurationModel struct {
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

type catalogModel struct {
	CatalogARN        fwtypes.ARN `tfsdk:"catalog_arn"`
	WarehouseLocation fwtypes.ARN `tfsdk:"warehouse_location"`
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

type s3DestinationConfigurationModel struct {
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
