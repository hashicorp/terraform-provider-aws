// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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
			names.AttrRule: framework.DataSourceComputedListOfObjectAttribute[dataBucketRepConfigRuleModel](ctx),
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
	Bucket types.String                                                  `tfsdk:"bucket"`
	Role   types.String                                                  `tfsdk:"role"`
	Rule   fwtypes.ListNestedObjectValueOf[dataBucketRepConfigRuleModel] `tfsdk:"rule"`
}

type dataBucketRepConfigRuleModel struct {
	DeleteMarkerReplication   fwtypes.ListNestedObjectValueOf[dataBucketRepConfigDeleteMarkerReplicationModel]   `tfsdk:"delete_marker_replication"`
	Destination               fwtypes.ListNestedObjectValueOf[dataBucketRepConfigDestinationModel]               `tfsdk:"destination"`
	ExistingObjectReplication fwtypes.ListNestedObjectValueOf[dataBucketRepConfigExistingObjectReplicationModel] `tfsdk:"existing_object_replication"`
	Filter                    fwtypes.ListNestedObjectValueOf[dataBucketRepConfigFilterModel]                    `tfsdk:"filter"`
	ID                        types.String                                                                       `tfsdk:"id"`
	Prefix                    types.String                                                                       `tfsdk:"prefix"`
	Priority                  types.Int64                                                                        `tfsdk:"priority"`
	Status                    types.String                                                                       `tfsdk:"status"`
	SourceSelectionCriteria   fwtypes.ListNestedObjectValueOf[dataBucketRepConfigSourceSelectionCriteriaModel]   `tfsdk:"source_selection_criteria"`
}

type dataBucketRepConfigDeleteMarkerReplicationModel struct {
	Status types.String `tfsdk:"status"`
}

type dataBucketRepConfigDestinationModel struct {
	AccessControlTranslation fwtypes.ListNestedObjectValueOf[dataBucketRepConfigAccessControlTranslationModel] `tfsdk:"access_control_translation"`
	Account                  types.String                                                                      `tfsdk:"account"`
	Bucket                   types.String                                                                      `tfsdk:"bucket"`
	EncryptionConfiguration  fwtypes.ListNestedObjectValueOf[dataBucketRepConfigEncryptionConfigurationModel]  `tfsdk:"encryption_configuration"`
	Metrics                  fwtypes.ListNestedObjectValueOf[dataBucketRepConfigMetricsModel]                  `tfsdk:"metrics"`
	ReplicationTime          fwtypes.ListNestedObjectValueOf[dataBucketRepConfigReplicationTimeModel]          `tfsdk:"replication_time"`
	StorageClass             types.String                                                                      `tfsdk:"storage_class"`
}

type dataBucketRepConfigAccessControlTranslationModel struct {
	Owner types.String `tfsdk:"owner"`
}

type dataBucketRepConfigEncryptionConfigurationModel struct {
	ReplicaKmsKeyId types.String `tfsdk:"replica_kms_key_id"`
}

type dataBucketRepConfigMetricsModel struct {
	EventThreshold fwtypes.ListNestedObjectValueOf[dataBucketRepConfigEventThresholdModel] `tfsdk:"event_threshold"`
	Status         types.String                                                            `tfsdk:"status"`
}

type dataBucketRepConfigEventThresholdModel struct {
	Minutes types.Int64 `tfsdk:"minutes"`
}

type dataBucketRepConfigReplicationTimeModel struct {
	Status types.String                                                  `tfsdk:"status"`
	Time   fwtypes.ListNestedObjectValueOf[dataBucketRepConfigTimeModel] `tfsdk:"time"`
}

type dataBucketRepConfigTimeModel struct {
	Minutes types.Int64 `tfsdk:"minutes"`
}

type dataBucketRepConfigExistingObjectReplicationModel struct {
	Status types.String `tfsdk:"status"`
}

type dataBucketRepConfigFilterModel struct {
	And    fwtypes.ListNestedObjectValueOf[dataBucketRepConfigFilterAndModel] `tfsdk:"and"`
	Tag    fwtypes.ListNestedObjectValueOf[dataBucketRepConfigFilterTagModel] `tfsdk:"tag"`
	Prefix types.String                                                       `tfsdk:"prefix"`
}

type dataBucketRepConfigFilterAndModel struct {
	Prefix types.String                                                       `tfsdk:"prefix"`
	Tag    fwtypes.ListNestedObjectValueOf[dataBucketRepConfigFilterTagModel] `tfsdk:"tag"`
}

type dataBucketRepConfigFilterTagModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type dataBucketRepConfigSourceSelectionCriteriaModel struct {
	ReplicaModifications   fwtypes.ListNestedObjectValueOf[dataBucketRepConfigReplicaModificationsModel]   `tfsdk:"replica_modifications"`
	SseKmsEncryptedObjects fwtypes.ListNestedObjectValueOf[dataBucketRepConfigSseKmsEncryptedObjectsModel] `tfsdk:"sse_kms_encrypted_objects"`
}

type dataBucketRepConfigReplicaModificationsModel struct {
	Status types.String `tfsdk:"status"`
}

type dataBucketRepConfigSseKmsEncryptedObjectsModel struct {
	Status types.String `tfsdk:"status"`
}
