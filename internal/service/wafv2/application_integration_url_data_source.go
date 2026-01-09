// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_wafv2_application_integration_url", name="Application Integration URL")
func newDataSourceApplicationIntegrationURL(_ context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceApplicationIntegrationURL{}, nil
}

const (
	ApplicationIntegrationURL = "Captcha API Application Integration URL Data Source"
)

type dataSourceApplicationIntegrationURL struct {
	framework.DataSourceWithModel[dataSourceApplicationIntegrationURLModel]
}

func (d *dataSourceApplicationIntegrationURL) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrURL: schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *dataSourceApplicationIntegrationURL) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().WAFV2Client(ctx)

	var data dataSourceApplicationIntegrationURLModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := wafv2.ListAPIKeysInput{
		Scope: awstypes.ScopeRegional,
		Limit: aws.Int32(1),
	}

	out, err := conn.ListAPIKeys(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionReading, ApplicationIntegrationURL, data.URL.String(), err),
			err.Error(),
		)
		return
	}

	data.URL = types.StringValue(aws.ToString(out.ApplicationIntegrationURL))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

type dataSourceApplicationIntegrationURLModel struct {
	framework.WithRegionModel
	URL types.String `tfsdk:"url"`
}
