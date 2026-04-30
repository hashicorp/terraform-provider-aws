// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"fmt"
	"iter"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_api_gateway_rest_api")
func newRestAPIResourceAsListResource() inttypes.ListResourceForSDK {
	l := restAPIListResource{}
	l.SetResourceSchema(resourceRestAPI())
	return &l
}

var _ list.ListResource = &restAPIListResource{}
var _ list.ListResourceWithRawV5Schemas = &restAPIListResource{}

type restAPIListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *restAPIListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.APIGatewayClient(ctx)

	tflog.Info(ctx, "Listing API Gateway REST APIs")

	stream.Results = func(yield func(list.ListResult) bool) {
		input := apigateway.GetRestApisInput{}

		for item, err := range listRestAPIs(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			id := aws.ToString(item.Id)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(id)

			if request.IncludeResource {
				if err := listResourceRestAPIFlatten(ctx, awsClient, rd, &item); err != nil {
					tflog.Error(ctx, "Reading API Gateway REST API", map[string]any{
						"error": err.Error(),
					})
					continue
				}
			}

			result.DisplayName = aws.ToString(item.Name)

			l.SetResult(ctx, awsClient, request.IncludeResource, rd, &result)
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

func listResourceRestAPIFlatten(ctx context.Context, c *conns.AWSClient, d *schema.ResourceData, api *awstypes.RestApi) error {
	d.Set("api_key_source", api.ApiKeySource)
	d.Set(names.AttrARN, apiARN(ctx, c, d.Id()))
	d.Set("binary_media_types", api.BinaryMediaTypes)
	d.Set(names.AttrCreatedDate, api.CreatedDate.Format(time.RFC3339))
	d.Set(names.AttrDescription, api.Description)
	d.Set("disable_execute_api_endpoint", api.DisableExecuteApiEndpoint)
	if err := d.Set("endpoint_configuration", flattenEndpointConfiguration(api.EndpointConfiguration)); err != nil {
		return fmt.Errorf("setting endpoint_configuration: %w", err)
	}
	d.Set("execution_arn", apiInvokeARN(ctx, c, d.Id()))
	if api.MinimumCompressionSize == nil {
		d.Set("minimum_compression_size", nil)
	} else {
		d.Set("minimum_compression_size", flex.Int32ToStringValue(api.MinimumCompressionSize))
	}
	d.Set(names.AttrName, api.Name)

	policy, err := flattenAPIPolicy(api.Policy)
	if err != nil {
		return err
	}

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get(names.AttrPolicy).(string), policy)
	if err != nil {
		return err
	}

	d.Set(names.AttrPolicy, policyToSet)

	setTagsOut(ctx, api.Tags)

	return nil
}

func listRestAPIs(ctx context.Context, conn *apigateway.Client, input *apigateway.GetRestApisInput) iter.Seq2[awstypes.RestApi, error] {
	return func(yield func(awstypes.RestApi, error) bool) {
		pages := apigateway.NewGetRestApisPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.RestApi{}, fmt.Errorf("listing API Gateway REST API resources: %w", err))
				return
			}

			for _, item := range page.Items {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
