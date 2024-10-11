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

// @FrameworkDataSource(aws_prometheus_default_scraper_configuration, name="Default Scraper Configuration")
func newDefaultScraperConfigurationDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &defaultScraperConfigurationDataSource{}, nil
}

type defaultScraperConfigurationDataSource struct {
	framework.DataSourceWithConfigure
}

func (*defaultScraperConfigurationDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	response.TypeName = "aws_prometheus_default_scraper_configuration"
}

func (d *defaultScraperConfigurationDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrConfiguration: schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *defaultScraperConfigurationDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data defaultScraperConfigurationDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().AMPClient(ctx)

	out, err := findDefaultScraperConfiguration(ctx, conn)

	if err != nil {
		response.Diagnostics.AddError("reading Prometheus Default Scraper Configuration", err.Error())

		return
	}

	data.Configuration = fwflex.StringValueToFramework(ctx, string(out))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func findDefaultScraperConfiguration(ctx context.Context, conn *amp.Client) ([]byte, error) {
	input := &amp.GetDefaultScraperConfigurationInput{}
	output, err := conn.GetDefaultScraperConfiguration(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || output.Configuration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Configuration, err
}

type defaultScraperConfigurationDataSourceModel struct {
	Configuration types.String `tfsdk:"configuration"`
}
