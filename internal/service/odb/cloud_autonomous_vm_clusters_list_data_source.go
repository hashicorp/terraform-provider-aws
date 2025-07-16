// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"

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
// @FrameworkDataSource("aws_odb_cloud_autonomous_vm_clusters_list", name="Cloud Autonomous Vm Clusters List")
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
				Description: "List of Cloud Autonomous VM Clusters (OCID, ID, ARN, OCI URL, Display Name)",
				CustomType:  fwtypes.NewListNestedObjectTypeOf[cloudAutonomousVmClusterSummary](ctx),
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"arn":          types.StringType,
						"id":           types.StringType,
						"oci_url":      types.StringType,
						"ocid":         types.StringType,
						"display_name": types.StringType,
					},
				},
			},
		},
	}
}

// Data sources only have a read method.
func (d *dataSourceCloudAutonomousVmClustersList) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	conn := d.Meta().ODBClient(ctx)

	var data cloudExadataInfrastructuresListDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.ListCloudAutonomousVmClusters(ctx, &odb.ListCloudAutonomousVmClustersInput{})
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

type cloudAutonomousVmClusterListModel struct {
	framework.WithRegionModel
	CloudAutonomousVmClusters fwtypes.ListNestedObjectValueOf[cloudAutonomousVmClusterSummary] `tfsdk:"cloud_autonomous_vm_clusters"`
}

type cloudAutonomousVmClusterSummary struct {
	CloudExadataInfrastructureArn types.String `tfsdk:"arn"`
	CloudAutonomousVmClusterId    types.String `tfsdk:"id"`
	OciUrl                        types.String `tfsdk:"oci_url"`
	Ocid                          types.String `tfsdk:"ocid"`
	DisplayName                   types.String `tfsdk:"display_name"`
}
