// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_api_gateway_integration_response")
func newIntegrationResponseResourceAsListResource() inttypes.ListResourceForSDK {
	l := integrationResponseListResource{}
	l.SetResourceSchema(resourceIntegrationResponse())
	return &l
}

var _ list.ListResource = &integrationResponseListResource{}
var _ list.ListResourceWithRawV5Schemas = &integrationResponseListResource{}

type integrationResponseListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *integrationResponseListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"rest_api_id": listschema.StringAttribute{
				Required:    true,
				Description: "ID of the associated REST API.",
			},
			names.AttrResourceID: listschema.StringAttribute{
				Required:    true,
				Description: "ID of the API Gateway Resource.",
			},
			"http_method": listschema.StringAttribute{
				Required:    true,
				Description: "HTTP method of the API Gateway Method.",
			},
		},
	}
}

func (l *integrationResponseListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().APIGatewayClient(ctx)

	var query listIntegrationResponseModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	restAPIID := query.RestAPIID.ValueString()
	resourceID := query.ResourceID.ValueString()
	httpMethod := query.HTTPMethod.ValueString()

	tflog.Info(ctx, "Listing API Gateway Integration Responses", map[string]any{
		logging.ResourceAttributeKey("rest_api_id"):        restAPIID,
		logging.ResourceAttributeKey(names.AttrResourceID): resourceID,
		logging.ResourceAttributeKey("http_method"):        httpMethod,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		integration, err := findIntegrationByThreePartKey(ctx, conn, httpMethod, resourceID, restAPIID)
		if err != nil {
			result := fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("reading API Gateway Integration (%s/%s/%s): %w", restAPIID, resourceID, httpMethod, err))
			yield(result)
			return
		}

		for statusCode := range integration.IntegrationResponses {
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrStatusCode), statusCode)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(resourceIntegrationResponseIDAttr(restAPIID, resourceID, httpMethod, statusCode))
			rd.Set("rest_api_id", restAPIID)
			rd.Set(names.AttrResourceID, resourceID)
			rd.Set("http_method", httpMethod)
			rd.Set(names.AttrStatusCode, statusCode)

			if request.IncludeResource {
				integrationResponse, err := findIntegrationResponseByFourPartKey(ctx, conn, httpMethod, resourceID, restAPIID, statusCode)
				if err != nil {
					tflog.Error(ctx, "Reading API Gateway Integration Response", map[string]any{
						"error": err.Error(),
					})
					continue
				}
				resourceIntegrationResponseFlatten(rd, integrationResponse)
			}

			result.DisplayName = fmt.Sprintf("%s %s", httpMethod, statusCode)

			l.SetResult(ctx, l.Meta(), request.IncludeResource, rd, &result)
			if result.Diagnostics.HasError() {
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

type listIntegrationResponseModel struct {
	framework.WithRegionModel
	RestAPIID  types.String `tfsdk:"rest_api_id"`
	ResourceID types.String `tfsdk:"resource_id"`
	HTTPMethod types.String `tfsdk:"http_method"`
}
