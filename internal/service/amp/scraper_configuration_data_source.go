// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amp

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/amp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource(name="ScraperConfiguration")
func newDataSourceScraperConfiguration(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceScraperConfiguration{}, nil
}

const (
	DSNameScraperConfiguration = "ScraperConfiguration Data Source"
)

type dataSourceScraperConfiguration struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceScraperConfiguration) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_prometheus_scraper_configuration"
}

func (d *dataSourceScraperConfiguration) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"default": schema.StringAttribute{
				Computed: true,
			},
			"id": framework.IDAttribute(),
		},
	}
}

func (d *dataSourceScraperConfiguration) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().AMPClient(ctx)

	var data dataSourceScraperConfigurationData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findScraperConfiguration(ctx, conn)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AMP, create.ErrActionReading, DSNameScraperConfiguration, "", err),
			err.Error(),
		)
		return
	}

	data.ID = flex.StringToFramework(ctx, conn.Options().BaseEndpoint)
	data.Default = flex.StringValueToFramework(ctx, string(out))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func findScraperConfiguration(ctx context.Context, conn *amp.Client) ([]byte, error) {
	input := &amp.GetDefaultScraperConfigurationInput{}

	out, err := conn.GetDefaultScraperConfiguration(ctx, input)
	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return out.Configuration, err
}

type dataSourceScraperConfigurationData struct {
	Default types.String `tfsdk:"default"`
	ID      types.String `tfsdk:"id"`
}
