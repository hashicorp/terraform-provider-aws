// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// @FrameworkDataSource("aws_apigatewayv2_authorizer", name="Authorizer")
func newDataSourceAuthorizer(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceAuthorizer{}, nil
}

type dataSourceAuthorizer struct {
	framework.DataSourceWithModel[dataSourceAuthorizerModel]
}

func (d *dataSourceAuthorizer) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_id": schema.StringAttribute{
				Optional: true,
			},
			"authorizer_id": schema.StringAttribute{
				Optional: true,
			},
			"authorizer_credentials_arn": schema.StringAttribute{
				Computed: true,
			},
			"authorizer_payload_format_version": schema.StringAttribute{
				Computed: true,
			},
			"authorizer_result_ttl_in_seconds": schema.Int64Attribute{
				Computed: true,
			},
			"authorizer_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AuthorizerType](),
				Computed:   true,
			},
			"authorizer_uri": schema.StringAttribute{
				Computed: true,
			},
			"enable_simple_responses": schema.BoolAttribute{
				Computed: true,
			},
			"identity_sources": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			"jwt_configuration": framework.DataSourceComputedListOfObjectAttribute[jwtConfigurationModel](ctx),
			"name": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *dataSourceAuthorizer) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data dataSourceAuthorizerModel

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().APIGatewayV2Client(ctx)

	out, err := findAuthorizerByTwoPartKey(ctx, conn, flex.StringValueFromFramework(ctx, data.APIID), flex.StringValueFromFramework(ctx, data.AuthorizerID))
	if err != nil {
		response.Diagnostics.AddError("reading API Gateway Authorizer", err.Error())
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type dataSourceAuthorizerModel struct {
	framework.WithRegionModel
	APIID                          types.String                                           `tfsdk:"api_id"`
	AuthorizerCredentialsArn       types.String                                           `tfsdk:"authorizer_credentials_arn"`
	AuthorizerID                   types.String                                           `tfsdk:"authorizer_id"`
	AuthorizerPayloadFormatVersion types.String                                           `tfsdk:"authorizer_payload_format_version"`
	AuthorizerResultTtlInSeconds   types.Int64                                            `tfsdk:"authorizer_result_ttl_in_seconds"`
	AuthorizerType                 fwtypes.StringEnum[awstypes.AuthorizerType]            `tfsdk:"authorizer_type"`
	AuthorizerUri                  types.String                                           `tfsdk:"authorizer_uri"`
	EnableSimpleResponses          types.Bool                                             `tfsdk:"enable_simple_responses"`
	IdentitySources                fwtypes.ListOfString                                   `tfsdk:"identity_sources"`
	JWTConfiguration               fwtypes.ListNestedObjectValueOf[jwtConfigurationModel] `tfsdk:"jwt_configuration"`
	Name                           types.String                                           `tfsdk:"name"`
}

type jwtConfigurationModel struct {
	Audience fwtypes.ListOfString `tfsdk:"audience"`
	Issuer   types.String         `tfsdk:"issuer"`
}
