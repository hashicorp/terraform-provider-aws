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

// @FrameworkDataSource("aws_odb_cloud_exadata_infrastructures", name="Cloud Exadata Infrastructures")
func newDataSourceCloudExadataInfrastructuresList(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceCloudExadataInfrastructuresList{}, nil
}

const (
	DSNameCloudExadataInfrastructuresList = "Cloud Exadata Infrastructures List Data Source"
)

type dataSourceCloudExadataInfrastructuresList struct {
	framework.DataSourceWithModel[cloudExadataInfrastructuresListDataSourceModel]
}

func (d *dataSourceCloudExadataInfrastructuresList) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cloud_exadata_infrastructures": schema.ListAttribute{
				Computed:    true,
				Description: "List of Cloud Exadata Infrastructures. Returns basic information about the Cloud Exadata Infrastructures.",
				CustomType:  fwtypes.NewListNestedObjectTypeOf[cloudExadataInfrastructureDataSourceListSummary](ctx),
			},
		},
	}
}

func (d *dataSourceCloudExadataInfrastructuresList) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ODBClient(ctx)
	var data cloudExadataInfrastructuresListDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := ListCloudExadataInfrastructures(ctx, conn)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, DSNameCloudExadataInfrastructuresList, "", err),
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

func ListCloudExadataInfrastructures(ctx context.Context, conn *odb.Client) (*odb.ListCloudExadataInfrastructuresOutput, error) {
	var out odb.ListCloudExadataInfrastructuresOutput
	paginator := odb.NewListCloudExadataInfrastructuresPaginator(conn, &odb.ListCloudExadataInfrastructuresInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		out.CloudExadataInfrastructures = append(out.CloudExadataInfrastructures, page.CloudExadataInfrastructures...)
	}
	return &out, nil
}

type cloudExadataInfrastructuresListDataSourceModel struct {
	framework.WithRegionModel
	CloudExadataInfrastructures fwtypes.ListNestedObjectValueOf[cloudExadataInfrastructureDataSourceListSummary] `tfsdk:"cloud_exadata_infrastructures"`
}

type cloudExadataInfrastructureDataSourceListSummary struct {
	CloudExadataInfrastructureArn types.String `tfsdk:"arn"`
	CloudExadataInfrastructureId  types.String `tfsdk:"id"`
	OciResourceAnchorName         types.String `tfsdk:"oci_resource_anchor_name"`
	OciUrl                        types.String `tfsdk:"oci_url"`
	Ocid                          types.String `tfsdk:"ocid"`
	DisplayName                   types.String `tfsdk:"display_name"`
}
