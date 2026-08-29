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
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_apigatewayv2_api")
func newAPIResourceAsListResource() inttypes.ListResourceForSDK {
	l := apiListResource{}
	l.SetResourceSchema(resourceAPI())
	return &l
}

var _ list.ListResource = &apiListResource{}

type apiListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *apiListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().APIGatewayV2Client(ctx)

	var query listAPIModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing Resources")

	stream.Results = func(yield func(list.ListResult) bool) {
		input := apigatewayv2.GetApisInput{}
		for item, err := range listAPIs(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			id := aws.ToString(item.ApiId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(id)

			if request.IncludeResource {
				if err := resourceAPIFlattenPage(ctx, l.Meta(), rd, &item); err != nil {
					tflog.Error(ctx, "Reading API Gateway V2 API", map[string]any{
						"error": err.Error(),
					})
					continue
				}
			}

			result.DisplayName = aws.ToString(item.Name)

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

type listAPIModel struct {
	framework.WithRegionModel
}

func listAPIs(ctx context.Context, conn *apigatewayv2.Client, input *apigatewayv2.GetApisInput) iter.Seq2[awstypes.Api, error] {
	return func(yield func(awstypes.Api, error) bool) {
		var stopped bool
		err := getAPIsPages(ctx, conn, input, func(page *apigatewayv2.GetApisOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}
			for _, item := range page.Items {
				if !yield(item, nil) {
					stopped = true
					return false
				}
			}
			return !lastPage
		})
		if !stopped && err != nil {
			yield(inttypes.Zero[awstypes.Api](), fmt.Errorf("listing API Gateway V2 APIs: %w", err))
		}
	}
}

func resourceAPIFlattenPage(ctx context.Context, awsClient *conns.AWSClient, d *schema.ResourceData, api *awstypes.Api) error {
	d.Set("api_endpoint", api.ApiEndpoint)
	d.Set("api_key_selection_expression", api.ApiKeySelectionExpression)
	d.Set(names.AttrARN, apiARN(ctx, awsClient, d.Id()))
	if err := d.Set("cors_configuration", flattenCORS(api.CorsConfiguration)); err != nil {
		return fmt.Errorf("setting cors_configuration: %w", err)
	}
	d.Set(names.AttrDescription, api.Description)
	d.Set("disable_execute_api_endpoint", api.DisableExecuteApiEndpoint)
	d.Set("execution_arn", apiInvokeARN(ctx, awsClient, d.Id()))
	d.Set(names.AttrIPAddressType, api.IpAddressType)
	d.Set(names.AttrName, api.Name)
	d.Set("protocol_type", api.ProtocolType)
	d.Set("route_selection_expression", api.RouteSelectionExpression)
	d.Set(names.AttrVersion, api.Version)

	setTagsOut(ctx, api.Tags)

	return nil
}
