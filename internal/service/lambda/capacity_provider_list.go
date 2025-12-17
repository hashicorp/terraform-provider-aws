// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"fmt"
	"iter"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/listresource"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @FrameworkListResource("aws_lambda_capacity_provider")
func newCapacityProviderResourceAsListResource() list.ListResourceWithConfigure {
	return &listResourceCapacityProvider{}
}

var _ list.ListResource = &listResourceCapacityProvider{}

type listResourceCapacityProvider struct {
	resourceCapacityProvider
	framework.WithList
}

func (r *listResourceCapacityProvider) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{},
	}
}

func (r *listResourceCapacityProvider) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query capacityProviderListModel

	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	awsClient := r.Meta()
	conn := awsClient.LambdaClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		result := request.NewListResult(ctx)
		var input lambda.ListCapacityProvidersInput
		for capacityProvider, err := range listCapacityProviders(ctx, conn, &input) {
			if err != nil {
				result = list.ListResult{
					Diagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic(
							"Error Listing Remote Resources",
							fmt.Sprintf("Error: %s", err),
						),
					},
				}
				yield(result)
				return
			}

			ctx = tftags.NewContext(ctx, awsClient.DefaultTagsConfig(ctx), awsClient.IgnoreTagsConfig(ctx), awsClient.TagPolicyConfig(ctx))
			var data resourceCapacityProviderModel
			timeoutObject, d := r.ListResourceTimeoutInit(ctx, result)
			result.Diagnostics.Append(d...)
			if result.Diagnostics.HasError() {
				result = list.ListResult{Diagnostics: result.Diagnostics}
				yield(result)
				return
			}

			data.Timeouts.Object = timeoutObject
			data.Tags.MapValue = r.ListResourceTagsInit(ctx, result)
			data.TagsAll.MapValue = r.ListResourceTagsInit(ctx, result)

			params := listresource.InterceptorParams{
				C:      awsClient,
				Result: &result,
			}

			if diags := r.RunResultInterceptors(ctx, listresource.Before, params); diags.HasError() {
				result.Diagnostics.Append(diags...)
				yield(result)
				return
			}

			if diags := flex.Flatten(ctx, capacityProvider, &data, flex.WithFieldNamePrefix(capacityProviderNamePrefix)); diags.HasError() {
				result.Diagnostics.Append(diags...)
				yield(result)
				return
			}

			cpARN, err := arn.Parse(data.ARN.ValueString())
			if err != nil {
				result = list.ListResult{
					Diagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic(
							"Error Listing Remote Resources",
							fmt.Sprintf("Error: %s", err),
						),
					},
				}
				yield(result)
				return
			}

			name := strings.TrimPrefix(cpARN.Resource, "capacity-provider:")
			data.Name = flex.StringValueToFramework(ctx, name)

			if diags := result.Resource.Set(ctx, &data); diags.HasError() {
				result.Diagnostics.Append(diags...)
				yield(result)
				return
			}

			result.DisplayName = name

			if diags := r.RunResultInterceptors(ctx, listresource.After, params); diags.HasError() {
				result.Diagnostics.Append(diags...)
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

type capacityProviderListModel struct {
	framework.WithRegionModel
}

func listCapacityProviders(ctx context.Context, conn *lambda.Client, input *lambda.ListCapacityProvidersInput) iter.Seq2[awstypes.CapacityProvider, error] {
	return func(yield func(awstypes.CapacityProvider, error) bool) {
		pages := lambda.NewListCapacityProvidersPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.CapacityProvider{}, fmt.Errorf("listing Lambda Capacity Providers: %w", err))
				return
			}

			for _, cp := range page.CapacityProviders {
				if !yield(cp, nil) {
					return
				}
			}
		}
	}
}
