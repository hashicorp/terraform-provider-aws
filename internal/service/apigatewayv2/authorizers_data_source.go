// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_apigatewayv2_authorizers", name="Authorizers")
func newDataSourceAuthorizers(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceAuthorizers{}, nil
}

type dataSourceAuthorizers struct {
	framework.DataSourceWithModel[dataSourceAuthorizersModel]
}

func (d *dataSourceAuthorizers) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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

func (d *dataSourceAuthorizers) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data dataSourceAuthorizersModel

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().APIGatewayV2Client(ctx)
	input := apigatewayv2.GetAuthorizersInput{
		ApiId: flex.StringFromFramework(ctx, data.APIID),
	}

	authorizers, err := findAuthorizers(ctx, conn, &input)
	if err != nil {
		response.Diagnostics.AddError("reading API Gateway Authorizers", err.Error())
		return
	}

	ids := []string{}

	for _, authorizer := range authorizers {

		ids = append(ids, aws.ToString(authorizer.AuthorizerId))
	}

	data.IDs = flex.FlattenFrameworkStringValueListOfString(ctx, ids)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func findAuthorizers(ctx context.Context, conn *apigatewayv2.Client, input *apigatewayv2.GetAuthorizersInput) ([]awstypes.Authorizer, error) {
	var output []awstypes.Authorizer

	err := getAuthorizersPages(ctx, conn, input, func(page *apigatewayv2.GetAuthorizersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.Items...)

		return !lastPage
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

type dataSourceAuthorizersModel struct {
	framework.WithRegionModel
	APIID types.String         `tfsdk:"api_id"`
	IDs   fwtypes.ListOfString `tfsdk:"ids"`
}
