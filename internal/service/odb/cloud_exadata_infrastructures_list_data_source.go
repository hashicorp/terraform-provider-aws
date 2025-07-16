//Copyright © 2025, Oracle and/or its affiliates. All rights reserved.

package odb

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_odb_cloud_exadata_infrastructures_list", name="Cloud Exadata Infrastructures List")
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
				Description: "List of Cloud Exadata Infrastructures (OCID, ID, ARN and OCI URL)",
				CustomType:  fwtypes.NewListNestedObjectTypeOf[cloudExadataInfrastructureDataSourceListSummary](ctx),
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"arn":     types.StringType,
						"id":      types.StringType,
						"oci_url": types.StringType,
						"ocid":    types.StringType,
					},
				},
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

	out, err := conn.ListCloudExadataInfrastructures(ctx, &odb.ListCloudExadataInfrastructuresInput{})
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

type cloudExadataInfrastructuresListDataSourceModel struct {
	framework.WithRegionModel
	CloudExadataInfrastructures fwtypes.ListNestedObjectValueOf[cloudExadataInfrastructureDataSourceListSummary] `tfsdk:"cloud_exadata_infrastructures"`
}

type cloudExadataInfrastructureDataSourceListSummary struct {
	CloudExadataInfrastructureArn types.String `tfsdk:"arn"`
	CloudExadataInfrastructureId  types.String `tfsdk:"id"`
	OciUrl                        types.String `tfsdk:"oci_url"`
	Ocid                          types.String `tfsdk:"ocid"`
}
