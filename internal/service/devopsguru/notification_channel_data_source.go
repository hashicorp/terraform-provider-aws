// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package devopsguru

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

// @FrameworkDataSource(name="Notification Channel")
func newDataSourceNotificationChannel(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceNotificationChannel{}, nil
}

const (
	DSNameNotificationChannel = "Notification Channel Data Source"
)

type dataSourceNotificationChannel struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceNotificationChannel) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_devopsguru_notification_channel"
}

func (d *dataSourceNotificationChannel) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"filters": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[filtersData](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"message_types": schema.ListAttribute{
							Computed:    true,
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
						},
						"severities": schema.ListAttribute{
							Computed:    true,
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
						},
					},
				},
			},
			"sns": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[snsData](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrTopicARN: schema.StringAttribute{
							Computed:   true,
							CustomType: fwtypes.ARNType,
						},
					},
				},
			},
		},
	}
}
func (d *dataSourceNotificationChannel) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().DevOpsGuruClient(ctx)

	var data dataSourceNotificationChannelData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findNotificationChannelByID(ctx, conn, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DevOpsGuru, create.ErrActionReading, DSNameNotificationChannel, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out.Config, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceNotificationChannelData struct {
	Filters fwtypes.ListNestedObjectValueOf[filtersData] `tfsdk:"filters"`
	ID      types.String                                 `tfsdk:"id"`
	Sns     fwtypes.ListNestedObjectValueOf[snsData]     `tfsdk:"sns"`
}
