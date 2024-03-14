// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package timestreamwrite

import (
	"context"

	// "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	// "github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	// "github.com/hashicorp/terraform-provider-aws/internal/tags"
	// tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource(name="Database")
func newDataSourceDatabase(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceDatabase{}, nil
}

const (
	DSNameDatabase = "Database Data Source"
)

type dataSourceDatabase struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceDatabase) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_timestreamwrite_database"
}

func (d *dataSourceDatabase) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"database_name": schema.StringAttribute{
				Required: true,
			},
			"kms_key_id": schema.StringAttribute{
				Computed: true,
			},
			"table_count": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

func (d *dataSourceDatabase) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().TimestreamWriteClient(ctx)

	var data dataSourceDatabaseData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findDatabaseByName(ctx, conn, data.DatabaseName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamWrite, create.ErrActionReading, DSNameDatabase, data.DatabaseName.String(), err),
			err.Error(),
		)
		return
	}

	data.ARN = flex.StringToFramework(ctx, out.Arn)
	data.DatabaseName = flex.StringToFramework(ctx, out.DatabaseName)
	data.KmsKeyId = flex.StringToFramework(ctx, out.KmsKeyId)
	data.TableCount = flex.Int64ToFramework(ctx, &out.TableCount)
	
	tags, err := listTags(ctx, conn, *out.DatabaseName)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamWrite, create.ErrActionReading, DSNameDatabase, data.DatabaseName.String(), err),
			err.Error(),
		)
		return
	}

	ignoreTagsConfig := d.Meta().IgnoreTagsConfig
	data.Tags = flex.FlattenFrameworkStringValueMapLegacy(ctx, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map())
	
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceDatabaseData struct {
	ARN          types.String `tfsdk:"arn"`
	DatabaseName types.String `tfsdk:"database_name"`
	KmsKeyId     types.String `tfsdk:"kms_key_id"`
	TableCount   types.Int64  `tfsdk:"table_count"`
	Tags		 types.Map 	  `tfsdk:"tags"`
}
