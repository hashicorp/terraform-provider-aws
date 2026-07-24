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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_apigatewayv2_stages", name="Stages")
func newDataSourceStages(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceStages{}, nil
}

type dataSourceStages struct {
	framework.DataSourceWithModel[dataSourceStagesModel]
}

func (d *dataSourceStages) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_id": schema.StringAttribute{
				Optional: true,
			},
			names.AttrNames: schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			names.AttrTags: tftags.TagsAttribute(),
		},
	}
}

func (d *dataSourceStages) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data dataSourceStagesModel

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().APIGatewayV2Client(ctx)
	input := apigatewayv2.GetStagesInput{
		ApiId: flex.StringFromFramework(ctx, data.APIID),
	}

	stages, err := findStages(ctx, conn, &input)
	if err != nil {
		response.Diagnostics.AddError("reading API Gateway Stages", err.Error())
		return
	}

	ignoreTagsConfig := d.Meta().IgnoreTagsConfig(ctx)
	tagsToMatch := tftags.New(ctx, data.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	names := []string{}

	for _, stage := range stages {

		if len(tagsToMatch) > 0 && !keyValueTags(ctx, stage.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).ContainsAll(tagsToMatch) {
			continue
		}

		names = append(names, aws.ToString(stage.StageName))
	}

	data.Names = flex.FlattenFrameworkStringValueListOfString(ctx, names)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type dataSourceStagesModel struct {
	framework.WithRegionModel
	APIID types.String         `tfsdk:"api_id"`
	Names fwtypes.ListOfString `tfsdk:"names"`
	Tags  tftags.Map           `tfsdk:"tags"`
}
