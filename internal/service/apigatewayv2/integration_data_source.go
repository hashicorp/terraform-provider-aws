// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_apigatewayv2_integration", name="Integration")
func newDataSourceIntegration(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceIntegration{}, nil
}

type dataSourceIntegration struct {
	framework.DataSourceWithModel[dataSourceIntegrationModel]
}

func (d *dataSourceIntegration) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_id": schema.StringAttribute{
				Optional: true,
			},
			"api_gateway_managed": schema.BoolAttribute{
				Computed: true,
			},
			"integration_id": schema.StringAttribute{
				Optional: true,
			},
			names.AttrConnectionID: schema.StringAttribute{
				Computed: true,
			},
			"connection_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ConnectionType](),
				Computed:   true,
			},
			"content_handling_strategy": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ContentHandlingStrategy](),
				Computed:   true,
			},
			"credentials_arn": schema.StringAttribute{
				Computed: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			"integration_method": schema.StringAttribute{
				Computed: true,
			},
			"integration_response_selection_expression": schema.StringAttribute{
				Computed: true,
			},
			"integration_subtype": schema.StringAttribute{
				Computed: true,
			},
			"integration_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.IntegrationType](),
				Computed:   true,
			},
			"integration_uri": schema.StringAttribute{
				Computed: true,
			},
			"passthrough_behavior": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.PassthroughBehavior](),
				Computed:   true,
			},
			"payload_format_version": schema.StringAttribute{
				Computed: true,
			},
			"request_parameters": schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			"request_templates": schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			"response_parameters": framework.DataSourceComputedListOfObjectAttribute[responseParametersModel](ctx),
			"template_selection_expression": schema.StringAttribute{
				Computed: true,
			},
			"timeout_milliseconds": schema.Int32Attribute{
				Computed: true,
			},
			"tls_config": framework.DataSourceComputedListOfObjectAttribute[tlsConfigModel](ctx),
		},
	}
}

func (d *dataSourceIntegration) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data dataSourceIntegrationModel

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().APIGatewayV2Client(ctx)
	input := apigatewayv2.GetIntegrationInput{
		ApiId:         flex.StringFromFramework(ctx, data.APIID),
		IntegrationId: flex.StringFromFramework(ctx, data.IntegrationID),
	}

	integration, err := findIntegration(ctx, conn, &input)
	if err != nil {
		response.Diagnostics.AddError("reading API Gateway Integration", err.Error())
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, integration, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	responseParameters, diag := flattenResponseParametersDataSource(ctx, integration.ResponseParameters)
	response.Diagnostics.Append(diag...)
	if response.Diagnostics.HasError() {
		return
	}
	data.ResponseParameters = responseParameters

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type dataSourceIntegrationModel struct {
	framework.WithRegionModel
	APIID                                  types.String                                             `tfsdk:"api_id"`
	IntegrationID                          types.String                                             `tfsdk:"integration_id"`
	APIGatewayManaged                      types.Bool                                               `tfsdk:"api_gateway_managed"`
	ConnectionID                           types.String                                             `tfsdk:"connection_id"`
	ConnectionType                         fwtypes.StringEnum[awstypes.ConnectionType]              `tfsdk:"connection_type"`
	ContentHandlingStrategy                fwtypes.StringEnum[awstypes.ContentHandlingStrategy]     `tfsdk:"content_handling_strategy"`
	CredentialsARN                         types.String                                             `tfsdk:"credentials_arn"`
	Description                            types.String                                             `tfsdk:"description"`
	IntegrationMethod                      types.String                                             `tfsdk:"integration_method"`
	IntegrationResponseSelectionExpression types.String                                             `tfsdk:"integration_response_selection_expression"`
	IntegrationSubtype                     types.String                                             `tfsdk:"integration_subtype"`
	IntegrationType                        fwtypes.StringEnum[awstypes.IntegrationType]             `tfsdk:"integration_type"`
	IntegrationURI                         types.String                                             `tfsdk:"integration_uri"`
	PassthroughBehavior                    fwtypes.StringEnum[awstypes.PassthroughBehavior]         `tfsdk:"passthrough_behavior"`
	PayloadFormatVersion                   types.String                                             `tfsdk:"payload_format_version"`
	RequestParameters                      fwtypes.MapOfString                                      `tfsdk:"request_parameters"`
	RequestTemplates                       fwtypes.MapOfString                                      `tfsdk:"request_templates"`
	ResponseParameters                     fwtypes.ListNestedObjectValueOf[responseParametersModel] `tfsdk:"response_parameters" autoflex:"-"`
	TemplateSelectionExpression            types.String                                             `tfsdk:"template_selection_expression"`
	TimeoutInMillis                        types.Int32                                              `tfsdk:"timeout_milliseconds"`
	TLSConfig                              fwtypes.ListNestedObjectValueOf[tlsConfigModel]          `tfsdk:"tls_config"`
}

type responseParametersModel struct {
	Mappings   fwtypes.MapOfString `tfsdk:"mappings"`
	StatusCode types.String        `tfsdk:"status_code"`
}

type tlsConfigModel struct {
	ServerNameToVerify types.String `tfsdk:"server_name_to_verify"`
}

func flattenResponseParametersDataSource(ctx context.Context, params map[string]map[string]string) (fwtypes.ListNestedObjectValueOf[responseParametersModel], diag.Diagnostics) {
	if len(params) == 0 {
		return fwtypes.NewListNestedObjectValueOfNull[responseParametersModel](ctx), diag.Diagnostics{}
	}

	var output []responseParametersModel
	for statusCode, mappings := range params {
		rp := responseParametersModel{
			StatusCode: types.StringValue(statusCode),
		}

		rawMap := map[string]attr.Value{}
		for key, value := range mappings {
			rawMap[key] = types.StringValue(value)
		}
		fwMap, diag := fwtypes.NewMapValueOf[types.String](ctx, rawMap)
		if diag.HasError() {
			return fwtypes.NewListNestedObjectValueOfNull[responseParametersModel](ctx), diag
		}
		rp.Mappings = fwMap
		output = append(output, rp)
	}
	slices.SortFunc(output, func(a, b responseParametersModel) int {
		return strings.Compare(a.StatusCode.String(), b.StatusCode.String())
	})

	return fwtypes.NewListNestedObjectValueOfValueSlice(ctx, output)
}
