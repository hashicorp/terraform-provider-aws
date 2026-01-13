// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_s3_bucket_replication_configuration", name="Bucket Replication Configuration")
func newDataSourceBucketReplicationConfiguration(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceBucketReplicationConfiguration{}, nil
}

const (
	DSNameBucketReplicationConfiguration = "Bucket Replication Configuration Data Source"
)

type dataSourceBucketReplicationConfiguration struct {
	framework.DataSourceWithModel[dataSourceBucketReplicationConfigurationModel]
}

func (d *dataSourceBucketReplicationConfiguration) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrBucket: schema.StringAttribute{
				Required: true,
			},

			names.AttrRole: schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrRule: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dsBucketRepConfigRuleModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1000),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrID: schema.StringAttribute{
							Computed: true,
						},
						names.AttrPrefix: schema.StringAttribute{
							Computed: true,
						},
						names.AttrPriority: schema.Int64Attribute{
							Computed: true,
						},
						names.AttrStatus: schema.StringAttribute{
							Computed: true,
						},
					},
					Blocks: map[string]schema.Block{
						"delete_marker_replication": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[dsBucketRepConfigDeleteMarkerReplicationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrStatus: schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
						names.AttrDestination: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[dsBucketRepConfigDestinationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"account": schema.StringAttribute{
										Computed: true,
									},
									names.AttrBucket: schema.StringAttribute{
										Computed: true,
									},
									names.AttrStorageClass: schema.StringAttribute{
										Computed: true,
									},
								},
								Blocks: map[string]schema.Block{
									"access_control_translation": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[dsBucketRepConfigAccessControlTranslationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrOwner: schema.StringAttribute{
													Computed: true,
												},
											},
										},
									},
									names.AttrEncryptionConfiguration: schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[dsBucketRepConfigEncryptionConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"replica_kms_key_id": schema.StringAttribute{
													Computed: true,
												},
											},
										},
									},
									"metrics": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[dsBucketRepConfigMetricsModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrStatus: schema.StringAttribute{
													Computed: true,
												},
											},
											Blocks: map[string]schema.Block{
												"event_threshold": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[dsBucketRepConfigEventThresholdModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"minutes": schema.Int64Attribute{
																Computed: true,
															},
														},
													},
												},
											},
										},
									},
									"replication_time": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[dsBucketRepConfigReplicationTimeModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrStatus: schema.StringAttribute{
													Computed: true,
												},
											},
											Blocks: map[string]schema.Block{
												"time": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[dsBucketRepConfigTimeModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"minutes": schema.Int64Attribute{
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
						"existing_object_replication": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[dsBucketRepConfigExistingObjectReplicationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrStatus: schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
						names.AttrFilter: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[dsBucketRepConfigFilterModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrPrefix: schema.StringAttribute{
										Computed: true,
									},
								},
								Blocks: map[string]schema.Block{
									"and": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[dsBucketRepConfigFilterAndModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrPrefix: schema.StringAttribute{
													Computed: true,
												},
											},
											Blocks: map[string]schema.Block{
												"tag": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[dsBucketRepConfigFilterTagModel](ctx),
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrKey: schema.StringAttribute{
																Computed: true,
															},
															names.AttrValue: schema.StringAttribute{
																Computed: true,
															},
														},
													},
												},
											},
										},
									},
									"tag": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[dsBucketRepConfigFilterTagModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrKey: schema.StringAttribute{
													Computed: true,
												},
												names.AttrValue: schema.StringAttribute{
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"source_selection_criteria": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[dsBucketRepConfigSourceSelectionCriteriaModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"replica_modifications": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[dsBucketRepConfigReplicaModificationsModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrStatus: schema.StringAttribute{
													Computed: true,
												},
											},
										},
									},
									"sse_kms_encrypted_objects": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[dsBucketRepConfigSseKmsEncryptedObjectsModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrStatus: schema.StringAttribute{
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
	}
}

func (d *dataSourceBucketReplicationConfiguration) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().S3Client(ctx)

	var data dataSourceBucketReplicationConfigurationModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findReplicationConfiguration(ctx, conn, data.Bucket.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3, create.ErrActionReading, DSNameBucketReplicationConfiguration, data.Bucket.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data, flex.WithNoIgnoredFieldNames())...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceBucketReplicationConfigurationModel struct {
	framework.WithRegionModel
	Bucket types.String                                                `tfsdk:"bucket"`
	Role   types.String                                                `tfsdk:"role"`
	Rule   fwtypes.ListNestedObjectValueOf[dsBucketRepConfigRuleModel] `tfsdk:"rule"`
}

type dsBucketRepConfigRuleModel struct {
	DeleteMarkerReplication   fwtypes.ListNestedObjectValueOf[dsBucketRepConfigDeleteMarkerReplicationModel]   `tfsdk:"delete_marker_replication"`
	Destination               fwtypes.ListNestedObjectValueOf[dsBucketRepConfigDestinationModel]               `tfsdk:"destination"`
	ExistingObjectReplication fwtypes.ListNestedObjectValueOf[dsBucketRepConfigExistingObjectReplicationModel] `tfsdk:"existing_object_replication"`
	Filter                    fwtypes.ListNestedObjectValueOf[dsBucketRepConfigFilterModel]                    `tfsdk:"filter"`
	ID                        types.String                                                                     `tfsdk:"id"`
	Prefix                    types.String                                                                     `tfsdk:"prefix"`
	Priority                  types.Int64                                                                      `tfsdk:"priority"`
	Status                    types.String                                                                     `tfsdk:"status"`
	SourceSelectionCriteria   fwtypes.ListNestedObjectValueOf[dsBucketRepConfigSourceSelectionCriteriaModel]   `tfsdk:"source_selection_criteria"`
}

type dsBucketRepConfigDeleteMarkerReplicationModel struct {
	Status types.String `tfsdk:"status"`
}

type dsBucketRepConfigDestinationModel struct {
	AccessControlTranslation fwtypes.ListNestedObjectValueOf[dsBucketRepConfigAccessControlTranslationModel] `tfsdk:"access_control_translation"`
	Account                  types.String                                                                    `tfsdk:"account"`
	Bucket                   types.String                                                                    `tfsdk:"bucket"`
	EncryptionConfiguration  fwtypes.ListNestedObjectValueOf[dsBucketRepConfigEncryptionConfigurationModel]  `tfsdk:"encryption_configuration"`
	Metrics                  fwtypes.ListNestedObjectValueOf[dsBucketRepConfigMetricsModel]                  `tfsdk:"metrics"`
	ReplicationTime          fwtypes.ListNestedObjectValueOf[dsBucketRepConfigReplicationTimeModel]          `tfsdk:"replication_time"`
	StorageClass             types.String                                                                    `tfsdk:"storage_class"`
}

type dsBucketRepConfigAccessControlTranslationModel struct {
	Owner types.String `tfsdk:"owner"`
}

type dsBucketRepConfigEncryptionConfigurationModel struct {
	ReplicaKmsKeyId types.String `tfsdk:"replica_kms_key_id"`
}

type dsBucketRepConfigMetricsModel struct {
	EventThreshold fwtypes.ListNestedObjectValueOf[dsBucketRepConfigEventThresholdModel] `tfsdk:"event_threshold"`
	Status         types.String                                                          `tfsdk:"status"`
}

type dsBucketRepConfigEventThresholdModel struct {
	Minutes types.Int64 `tfsdk:"minutes"`
}

type dsBucketRepConfigReplicationTimeModel struct {
	Status types.String                                                `tfsdk:"status"`
	Time   fwtypes.ListNestedObjectValueOf[dsBucketRepConfigTimeModel] `tfsdk:"time"`
}

type dsBucketRepConfigTimeModel struct {
	Minutes types.Int64 `tfsdk:"minutes"`
}

type dsBucketRepConfigExistingObjectReplicationModel struct {
	Status types.String `tfsdk:"status"`
}

type dsBucketRepConfigFilterModel struct {
	And    fwtypes.ListNestedObjectValueOf[dsBucketRepConfigFilterAndModel] `tfsdk:"and"`
	Tag    fwtypes.ListNestedObjectValueOf[dsBucketRepConfigFilterTagModel] `tfsdk:"tag"`
	Prefix types.String                                                     `tfsdk:"prefix"`
}

type dsBucketRepConfigFilterAndModel struct {
	Prefix types.String                                                     `tfsdk:"prefix"`
	Tag    fwtypes.ListNestedObjectValueOf[dsBucketRepConfigFilterTagModel] `tfsdk:"tag"`
}

type dsBucketRepConfigFilterTagModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type dsBucketRepConfigSourceSelectionCriteriaModel struct {
	ReplicaModifications   fwtypes.ListNestedObjectValueOf[dsBucketRepConfigReplicaModificationsModel]   `tfsdk:"replica_modifications"`
	SseKmsEncryptedObjects fwtypes.ListNestedObjectValueOf[dsBucketRepConfigSseKmsEncryptedObjectsModel] `tfsdk:"sse_kms_encrypted_objects"`
}

type dsBucketRepConfigReplicaModificationsModel struct {
	Status types.String `tfsdk:"status"`
}

type dsBucketRepConfigSseKmsEncryptedObjectsModel struct {
	Status types.String `tfsdk:"status"`
}
