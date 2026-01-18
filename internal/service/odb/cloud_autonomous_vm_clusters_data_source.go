// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/odb"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource("aws_odb_cloud_autonomous_vm_clusters", name="Cloud Autonomous Vm Clusters")
func newDataSourceCloudAutonomousVmClustersList(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceCloudAutonomousVmClustersList{}, nil
}

const (
	DSNameCloudAutonomousVmClustersList = "Cloud Autonomous Vm Clusters List Data Source"
)

type dataSourceCloudAutonomousVmClustersList struct {
	framework.DataSourceWithModel[cloudAutonomousVmClusterListModel]
}

func (d *dataSourceCloudAutonomousVmClustersList) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cloud_autonomous_vm_clusters": schema.ListAttribute{
				Computed:    true,
				Description: "List of Cloud Autonomous VM Clusters. The list going to contain basic information about the cloud autonomous VM clusters.",
				CustomType:  fwtypes.NewListNestedObjectTypeOf[cloudAutonomousVmClusterSummary](ctx),
			},
		},
	}
}

// Data sources only have a read method.
func (d *dataSourceCloudAutonomousVmClustersList) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ODBClient(ctx)
	var data cloudAutonomousVmClusterListModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := ListCloudAutonomousVmClusters(ctx, conn)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, DSNameCloudAutonomousVmClustersList, "", err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func ListCloudAutonomousVmClusters(ctx context.Context, conn *odb.Client) (*odb.ListCloudAutonomousVmClustersOutput, error) {
	out := odb.ListCloudAutonomousVmClustersOutput{}
	paginator := odb.NewListCloudAutonomousVmClustersPaginator(conn, &odb.ListCloudAutonomousVmClustersInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		out.CloudAutonomousVmClusters = append(out.CloudAutonomousVmClusters, page.CloudAutonomousVmClusters...)
	}
	return &out, nil
}

type cloudAutonomousVmClusterListModel struct {
	framework.WithRegionModel
	CloudAutonomousVmClusters fwtypes.ListNestedObjectValueOf[cloudAutonomousVmClusterSummary] `tfsdk:"cloud_autonomous_vm_clusters"`
}

type cloudAutonomousVmClusterSummary struct {
	CloudAutonomousVmClusterArn  types.String `tfsdk:"arn"`
	CloudAutonomousVmClusterId   types.String `tfsdk:"id"`
	CloudExadataInfrastructureId types.String `tfsdk:"cloud_exadata_infrastructure_id"`
	OdbNetworkId                 types.String `tfsdk:"odb_network_id"`
	OciResourceAnchorName        types.String `tfsdk:"oci_resource_anchor_name"`
	OciUrl                       types.String `tfsdk:"oci_url"`
	Ocid                         types.String `tfsdk:"ocid"`
	DisplayName                  types.String `tfsdk:"display_name"`
}
