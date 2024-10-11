// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package timestreamwrite

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/timestreamwrite"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Database")
func newDataSourceDatabase(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceDatabase{}, nil
}

type dataSourceDatabase struct {
	framework.DataSourceWithConfigure
}

const (
	DSNameDatabase = "Database data source"
)

func (d *dataSourceDatabase) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_timestreamwrite_database"
}

func (d *dataSourceDatabase) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 256),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			names.AttrCreatedTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrKMSKeyID: schema.StringAttribute{
				Computed: true,
			},
			"table_count": schema.Int64Attribute{
				Computed: true,
			},
			names.AttrLastUpdatedTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
	}
}

func (d *dataSourceDatabase) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().TimestreamWriteClient(ctx)
	var data dsDescribeDatabase

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in := &timestreamwrite.DescribeDatabaseInput{
		DatabaseName: data.Name.ValueStringPointer(),
	}

	desc, err := conn.DescribeDatabase(ctx, in)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Transfer, create.ErrActionReading, DSNameDatabase, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, desc.Database, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dsDescribeDatabase struct {
	ARN             types.String      `tfsdk:"arn"`
	CreatedTime     timetypes.RFC3339 `tfsdk:"created_time"`
	Name            types.String      `tfsdk:"name"`
	KmsKeyID        types.String      `tfsdk:"kms_key_id"`
	LastUpdatedTime timetypes.RFC3339 `tfsdk:"last_updated_time"`
	TableCount      types.Int64       `tfsdk:"table_count"`
}
