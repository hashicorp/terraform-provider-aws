// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amp

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/amp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Default Scraper Configuration")
func newScraperConfigurationDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &scraperConfigurationDataSource{}, nil
}

type scraperConfigurationDataSource struct {
	framework.DataSourceWithConfigure
}

func (*scraperConfigurationDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	response.TypeName = "aws_prometheus_scraper_configuration"
}

func (d *scraperConfigurationDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"default": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttribute(),
		},
	}
}

func (d *scraperConfigurationDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data scraperConfigurationDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().AMPClient(ctx)

	out, err := findScraperConfiguration(ctx, conn)

	if err != nil {
		response.Diagnostics.AddError("reading Prometheus Default Scraper Configuration", err.Error())

		return
	}

	data.ID = fwflex.StringToFramework(ctx, conn.Options().BaseEndpoint)
	data.Default = fwflex.StringValueToFramework(ctx, string(out))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
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

type scraperConfigurationDataSourceModel struct {
	Default types.String `tfsdk:"default"`
	ID      types.String `tfsdk:"id"`
}
