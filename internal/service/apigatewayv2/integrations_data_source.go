// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_apigatewayv2_integrations", name="Integrations")
func newDataSourceIntegrations(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceIntegrations{}, nil
}

type dataSourceIntegrations struct {
	framework.DataSourceWithModel[dataSourceIntegrationsModel]
}

func (d *dataSourceIntegrations) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_id": schema.StringAttribute{
				Optional: true,
			},
			names.AttrIDs: schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

func (d *dataSourceIntegrations) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data dataSourceIntegrationsModel

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().APIGatewayV2Client(ctx)
	input := apigatewayv2.GetIntegrationsInput{
		ApiId: flex.StringFromFramework(ctx, data.APIID),
	}

	integrations, err := findIntegrations(ctx, conn, &input)
	if err != nil {
		response.Diagnostics.AddError("reading API Gateway Integrations", err.Error())
		return
	}

	ids := []string{}

	for _, integration := range integrations {
		ids = append(ids, aws.ToString(integration.IntegrationId))
	}

	data.IDs = flex.FlattenFrameworkStringValueListOfString(ctx, ids)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type dataSourceIntegrationsModel struct {
	framework.WithRegionModel
	APIID types.String         `tfsdk:"api_id"`
	IDs   fwtypes.ListOfString `tfsdk:"ids"`
}
