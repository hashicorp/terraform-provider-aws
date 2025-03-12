// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_glue_registry", name="Registry")
func newDataSourceRegistry(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceRegistry{}, nil
}

const (
	DSNameRegistry = "Registry Data Source"
)

type dataSourceRegistry struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceRegistry) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (d *dataSourceRegistry) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().GlueClient(ctx)

	var data dataSourceRegistryData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindRegistryByName(ctx, conn, data.Name.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Glue, create.ErrActionReading, DSNameRegistry, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data, flex.WithFieldNamePrefix("Registry"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceRegistryData struct {
	ARN         types.String `tfsdk:"arn"`
	Description types.String `tfsdk:"description"`
	Name        types.String `tfsdk:"name"`
}
