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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_apigatewayv2_api")
func newAPIResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceAPI(), framework.SDKListSpec[awstypes.Api, framework.RegionQuery]{
		Items: func(ctx context.Context, c *conns.AWSClient, _ framework.RegionQuery) (iter.Seq2[awstypes.Api, error], diag.Diagnostics) {
			var input apigatewayv2.GetApisInput

			return listAPIs(ctx, c.APIGatewayV2Client(ctx), &input), nil
		},
		ID: func(a awstypes.Api, _ framework.RegionQuery) string {
			return aws.ToString(a.ApiId)
		},
		Flatten:     resourceAPIFlattenPage,
		DisplayName: func(a awstypes.Api) string { return aws.ToString(a.Name) },
	})
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

func resourceAPIFlattenPage(ctx context.Context, awsClient *conns.AWSClient, api *awstypes.Api, d *schema.ResourceData) error {
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
