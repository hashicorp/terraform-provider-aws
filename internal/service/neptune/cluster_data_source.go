// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	neptune "github.com/aws/aws-sdk-go-v2/service/neptune"
	awstypes "github.com/aws/aws-sdk-go-v2/service/neptune/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource("aws_neptune_cluster", name="Cluster")
func newDataSourceCluster(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceCluster{}, nil
}

type dataSourceCluster struct {
	framework.DataSourceWithModel[dataSourceClusterModel]
}

func (d *dataSourceCluster) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrIdentifier: schema.StringAttribute{
				Required: true,
			},
			"associated_roles": framework.DataSourceComputedListOfObjectAttribute[associatedRoleModel](ctx),
			"availability_zones": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			"backup_retention_period": schema.Int64Attribute{
				Computed: true,
			},
			"cluster_parameter_group_name": schema.StringAttribute{
				Computed: true,
			},
			"database_name": schema.StringAttribute{
				Computed: true,
			},
			"db_subnet_group": schema.StringAttribute{
				Computed: true,
			},
			"deletion_protection": schema.BoolAttribute{
				Computed: true,
			},
			"enabled_cloudwatch_logs_exports": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			names.AttrEndpoint: schema.StringAttribute{
				Computed: true,
			},
			names.AttrEngine: schema.StringAttribute{
				Computed: true,
			},
			names.AttrEngineVersion: schema.StringAttribute{
				Computed: true,
			},
			"global_cluster_identifier": schema.StringAttribute{
				Computed: true,
			},
			"iam_database_authentication_enabled": schema.BoolAttribute{
				Computed: true,
			},
			"kms_key_id": schema.StringAttribute{
				Computed: true,
			},
			"members": framework.DataSourceComputedListOfObjectAttribute[memberModel](ctx),
			"multi_az": schema.BoolAttribute{
				Computed: true,
			},
			names.AttrPort: schema.Int64Attribute{
				Computed: true,
			},
			"preferred_backup_window": schema.StringAttribute{
				Computed: true,
			},
			"preferred_maintenance_window": schema.StringAttribute{
				Computed: true,
			},
			"reader_endpoint": schema.StringAttribute{
				Computed: true,
			},
			"resource_id": schema.StringAttribute{
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			names.AttrStorageEncrypted: schema.BoolAttribute{
				Computed: true,
			},
			"storage_type": schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"vpc_security_group_ids": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			names.AttrID: framework.IDAttribute(),
		},
	}
}

func findClusterByID(ctx context.Context, conn *neptune.Client, id string) (*awstypes.DBCluster, error) {
	out, err := conn.DescribeDBClusters(ctx, &neptune.DescribeDBClustersInput{
		DBClusterIdentifier: &id,
	})
	if err != nil {
		return nil, err
	}
	if len(out.DBClusters) == 0 {
		return nil, fmt.Errorf("neptune cluster %q not found", id)
	}
	return &out.DBClusters[0], nil
}

func flattenClusterMembers(ctx context.Context, src []awstypes.DBClusterMember) (fwtypes.ListNestedObjectValueOf[memberModel], diag.Diagnostics) {
	memberModels := make([]memberModel, 0, len(src))
	for _, m := range src {
		memberModels = append(memberModels, memberModel{
			DBInstanceIdentifier: types.StringValue(aws.ToString(m.DBInstanceIdentifier)),
			IsClusterWriter:      types.BoolValue(aws.ToBool(m.IsClusterWriter)),
			ParameterGroupStatus: types.StringValue(aws.ToString(m.DBClusterParameterGroupStatus)),
			PromotionTier:        types.Int64Value(int64(aws.ToInt32(m.PromotionTier))),
		})
	}
	return fwtypes.NewListNestedObjectValueOfValueSlice[memberModel](ctx, memberModels)
}

func flattenAssociatedRoles(ctx context.Context, src []awstypes.DBClusterRole) (fwtypes.ListNestedObjectValueOf[associatedRoleModel], diag.Diagnostics) {
	roleModels := make([]associatedRoleModel, 0, len(src))
	for _, r := range src {
		roleModels = append(roleModels, associatedRoleModel{
			RoleArn:     types.StringValue(aws.ToString(r.RoleArn)),
			FeatureName: types.StringValue(aws.ToString(r.FeatureName)),
			Status:      types.StringValue(aws.ToString(r.Status)),
		})
	}
	return fwtypes.NewListNestedObjectValueOfValueSlice[associatedRoleModel](ctx, roleModels)
}

func (d *dataSourceCluster) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourceClusterModel
	conn := d.Meta().NeptuneClient(ctx)

	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	// Lookup the cluster
	cluster, err := findClusterByID(ctx, conn, data.Identifier.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.Identifier.String())
		return
	}

	// Flatten simple fields (DBCluster* → schema keys)
	smerr.EnrichAppend(ctx, &resp.Diagnostics,
		flex.Flatten(ctx, cluster, &data, flex.WithFieldNamePrefix("DBCluster")),
		smerr.ID, data.Identifier.String(),
	)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set ID = DBClusterIdentifier (Neptune-specific identity)
	dbClusterIdentifier := aws.ToString(cluster.DBClusterIdentifier)
	if dbClusterIdentifier != "" {
		data.ID = types.StringValue(dbClusterIdentifier)
	}

	// Set resource_id (DbClusterResourceId -> resource_id)
	dbClusterResourceID := aws.ToString(cluster.DbClusterResourceId)
	if dbClusterResourceID != "" {
		data.ResourceID = types.StringValue(dbClusterResourceID)
	}

	// Cluster members
	if len(cluster.DBClusterMembers) > 0 {
		memberList, diags := flattenClusterMembers(ctx, cluster.DBClusterMembers)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Members = memberList
	}

	//Associated roles
	if len(cluster.AssociatedRoles) > 0 {
		roleList, diags := flattenAssociatedRoles(ctx, cluster.AssociatedRoles)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.AssociatedRoles = roleList
	}

	// VPC security group IDs
	if len(cluster.VpcSecurityGroups) > 0 {
		elems := make([]attr.Value, 0, len(cluster.VpcSecurityGroups))
		for _, sg := range cluster.VpcSecurityGroups {
			if id := aws.ToString(sg.VpcSecurityGroupId); id != "" {
				elems = append(elems, types.StringValue(id))
			}
		}
		list, diags := fwtypes.NewListValueOf[types.String](ctx, elems)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.VPCSecurityGroupIDs = list
	}

	// Set Tags
	if arn := data.ARN.ValueString(); arn != "" {
		tagset, err := listTags(ctx, conn, arn)
		if err != nil {
			resp.Diagnostics.AddWarning(
				"listing Neptune cluster tags",
				fmt.Sprintf("failed to list tags for %s: %v", arn, err),
			)
		} else {
			ignore := d.Meta().IgnoreTagsConfig(ctx)
			tagset = tagset.IgnoreAWS().IgnoreConfig(ignore)

			data.Tags = tftags.FlattenStringValueMap(ctx, tagset.Map())
		}
	}

	// Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceClusterModel struct {
	framework.WithRegionModel

	ARN        types.String `tfsdk:"arn"`
	ID         types.String `tfsdk:"id"`
	Identifier types.String `tfsdk:"identifier"`

	BackupRetentionPeriod            types.Int64                                          `tfsdk:"backup_retention_period"`
	ClusterParameterGroupName        types.String                                         `tfsdk:"cluster_parameter_group_name"`
	DatabaseName                     types.String                                         `tfsdk:"database_name"`
	DBSubnetGroup                    types.String                                         `tfsdk:"db_subnet_group"`
	DeletionProtection               types.Bool                                           `tfsdk:"deletion_protection"`
	Endpoint                         types.String                                         `tfsdk:"endpoint"`
	Engine                           types.String                                         `tfsdk:"engine"`
	EngineVersion                    types.String                                         `tfsdk:"engine_version"`
	GlobalClusterIdentifier          types.String                                         `tfsdk:"global_cluster_identifier"`
	IAMDatabaseAuthenticationEnabled types.Bool                                           `tfsdk:"iam_database_authentication_enabled"`
	KMSKeyID                         types.String                                         `tfsdk:"kms_key_id"`
	MultiAZ                          types.Bool                                           `tfsdk:"multi_az"`
	Port                             types.Int64                                          `tfsdk:"port"`
	PreferredBackupWindow            types.String                                         `tfsdk:"preferred_backup_window"`
	PreferredMaintenanceWindow       types.String                                         `tfsdk:"preferred_maintenance_window"`
	ReaderEndpoint                   types.String                                         `tfsdk:"reader_endpoint"`
	ResourceID                       types.String                                         `tfsdk:"resource_id"`
	Status                           types.String                                         `tfsdk:"status"`
	StorageEncrypted                 types.Bool                                           `tfsdk:"storage_encrypted"`
	StorageType                      types.String                                         `tfsdk:"storage_type"`
	AvailabilityZones                fwtypes.ListValueOf[types.String]                    `tfsdk:"availability_zones"`
	EnabledCloudwatchLogsExports     fwtypes.ListValueOf[types.String]                    `tfsdk:"enabled_cloudwatch_logs_exports"`
	VPCSecurityGroupIDs              fwtypes.ListValueOf[types.String]                    `tfsdk:"vpc_security_group_ids"`
	AssociatedRoles                  fwtypes.ListNestedObjectValueOf[associatedRoleModel] `tfsdk:"associated_roles"`
	Members                          fwtypes.ListNestedObjectValueOf[memberModel]         `tfsdk:"members"`
	Tags                             tftags.Map                                           `tfsdk:"tags"`
}

type memberModel struct {
	DBInstanceIdentifier types.String `tfsdk:"db_instance_identifier"`
	IsClusterWriter      types.Bool   `tfsdk:"is_cluster_writer"`
	ParameterGroupStatus types.String `tfsdk:"parameter_group_status"`
	PromotionTier        types.Int64  `tfsdk:"promotion_tier"`
}

type associatedRoleModel struct {
	RoleArn     types.String `tfsdk:"role_arn"`
	FeatureName types.String `tfsdk:"feature_name"`
	Status      types.String `tfsdk:"status"`
}
