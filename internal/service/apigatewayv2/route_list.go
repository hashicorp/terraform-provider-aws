// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_apigatewayv2_route")
func newRouteResourceAsListResource() inttypes.ListResourceForSDK {
	l := routeListResource{}
	l.SetResourceSchema(resourceRoute())
	return &l
}

var _ list.ListResource = &routeListResource{}

type routeListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *routeListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"api_id": listschema.StringAttribute{
				Required:    true,
				Description: "API identifier.",
			},
		},
	}
}

func (l *routeListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().APIGatewayV2Client(ctx)

	var query listRouteModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	apiID := query.ApiId.ValueString()

	tflog.Info(ctx, "Listing Resources", map[string]any{
		logging.ResourceAttributeKey("api_id"): apiID,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := apigatewayv2.GetRoutesInput{
			ApiId: aws.String(apiID),
		}
		for item, err := range listRoutes(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			routeID := aws.ToString(item.RouteId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("route_id"), routeID)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(routeID)
			rd.Set("api_id", apiID)

			if request.IncludeResource {
				if err := flattenRoutePage(rd, item); err != nil {
					tflog.Error(ctx, "Reading API Gateway V2 Route", map[string]any{
						"error": err.Error(),
					})
					continue
				}
			}

			result.DisplayName = aws.ToString(item.RouteKey)

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

type listRouteModel struct {
	framework.WithRegionModel
	ApiId types.String `tfsdk:"api_id"`
}

func listRoutes(ctx context.Context, conn *apigatewayv2.Client, input *apigatewayv2.GetRoutesInput) iter.Seq2[awstypes.Route, error] {
	return func(yield func(awstypes.Route, error) bool) {
		err := getRoutesPages(ctx, conn, input, func(page *apigatewayv2.GetRoutesOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}
			for _, item := range page.Items {
				if !yield(item, nil) {
					return false
				}
			}
			return !lastPage
		})
		if err != nil {
			yield(awstypes.Route{}, fmt.Errorf("listing API Gateway V2 Route resources: %w", err))
		}
	}
}

func flattenRoutePage(d *schema.ResourceData, item awstypes.Route) error {
	d.Set("api_key_required", item.ApiKeyRequired)
	d.Set("authorization_scopes", item.AuthorizationScopes)
	d.Set("authorization_type", item.AuthorizationType)
	d.Set("authorizer_id", item.AuthorizerId)
	d.Set("model_selection_expression", item.ModelSelectionExpression)
	d.Set("operation_name", item.OperationName)
	d.Set("request_models", item.RequestModels)
	if err := d.Set("request_parameter", flattenRouteRequestParameters(item.RequestParameters)); err != nil {
		return fmt.Errorf("setting request_parameter: %w", err)
	}
	d.Set("route_key", item.RouteKey)
	d.Set("route_response_selection_expression", item.RouteResponseSelectionExpression)
	d.Set(names.AttrTarget, item.Target)

	return nil
}
