// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"

	"github.com/YakDriver/regexache"
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

// @FrameworkDataSource("aws_cloudfront_key_value_store", name="Key Value Store")
func newDataSourceKeyValueStore(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceKeyValueStore{}, nil
}

const (
	DSNameKeyValueStore = "Key Value Store Data Source"
)

type dataSourceKeyValueStore struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceKeyValueStore) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
			},
			names.AttrComment: schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Computed: true,
			},
			"last_modified_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^[a-zA-Z0-9-_]{1,64}$`),
						"must contain only alphanumeric characters, hyphens, and underscores",
					),
				},
			},
		},
	}
}

func (d *dataSourceKeyValueStore) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().CloudFrontClient(ctx)

	var data dataSourceKeyValueStoreModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findKeyValueStoreByName(ctx, conn, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFront, create.ErrActionReading, DSNameKeyValueStore, data.Name.String(), err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out.KeyValueStore, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.setID() // API response has a field named 'Id' which isn't the resource's ID.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceKeyValueStoreModel struct {
	ARN              types.String      `tfsdk:"arn"`
	Comment          types.String      `tfsdk:"comment"`
	ID               types.String      `tfsdk:"id"`
	LastModifiedTime timetypes.RFC3339 `tfsdk:"last_modified_time"`
	Name             types.String      `tfsdk:"name"`
}

func (data *dataSourceKeyValueStoreModel) setID() {
	data.ID = data.Name
}
