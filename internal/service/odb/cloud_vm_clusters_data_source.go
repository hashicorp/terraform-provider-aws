// Copyright IBM Corp. 2014, 2025
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
// @FrameworkDataSource("aws_odb_cloud_vm_clusters", name="Cloud Vm Clusters")
func newDataSourceCloudVmClustersList(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceCloudVmClustersList{}, nil
}

const (
	DSNameCloudVmClustersList = "Cloud Vm Clusters List Data Source"
)

type dataSourceCloudVmClustersList struct {
	framework.DataSourceWithModel[dataSourceCloudVmClustersListModel]
}

func (d *dataSourceCloudVmClustersList) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cloud_vm_clusters": schema.ListAttribute{
				Computed:    true,
				Description: "List of Cloud VM Clusters. It returns only basic information about the cloud VM clusters.",
				CustomType:  fwtypes.NewListNestedObjectTypeOf[cloudVmClusterSummary](ctx),
			},
		},
	}
}

// Data sources only have a read method.
func (d *dataSourceCloudVmClustersList) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ODBClient(ctx)
	var data dataSourceCloudVmClustersListModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := ListCloudVmClusters(ctx, conn)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, DSNameCloudVmClustersList, "", err),
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

func ListCloudVmClusters(ctx context.Context, conn *odb.Client) (*odb.ListCloudVmClustersOutput, error) {
	var out odb.ListCloudVmClustersOutput
	paginator := odb.NewListCloudVmClustersPaginator(conn, &odb.ListCloudVmClustersInput{})
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		out.CloudVmClusters = append(out.CloudVmClusters, output.CloudVmClusters...)
	}
	return &out, nil
}

type dataSourceCloudVmClustersListModel struct {
	framework.WithRegionModel
	CloudVmClusters fwtypes.ListNestedObjectValueOf[cloudVmClusterSummary] `tfsdk:"cloud_vm_clusters"`
}

type cloudVmClusterSummary struct {
	CloudAutonomousVmClusterId   types.String `tfsdk:"id"`
	CloudVmClusterArn            types.String `tfsdk:"arn"`
	CloudExadataInfrastructureId types.String `tfsdk:"cloud_exadata_infrastructure_id"`
	OciResourceAnchorName        types.String `tfsdk:"oci_resource_anchor_name"`
	OdbNetworkId                 types.String `tfsdk:"odb_network_id"`
	OciUrl                       types.String `tfsdk:"oci_url"`
	Ocid                         types.String `tfsdk:"ocid"`
	DisplayName                  types.String `tfsdk:"display_name"`
}
