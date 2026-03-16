// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_cloudfront_key_value_store")
func newKeyValueStoreResourceAsListResource() list.ListResourceWithConfigure {
	return &listResourceKeyValueStore{}
}

var _ list.ListResource = &listResourceKeyValueStore{}

type listResourceKeyValueStore struct {
	keyValueStoreResource
	framework.WithList
}

func (r *listResourceKeyValueStore) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query listKeyValueStoreModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	awsClient := r.Meta()
	conn := awsClient.CloudFrontClient(ctx)

	tflog.Info(ctx, "Listing CloudFront Key Value Stores")
	stream.Results = func(yield func(list.ListResult) bool) {
		var input cloudfront.ListKeyValueStoresInput
		for item, err := range listKeyValueStores(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			id := aws.ToString(item.Id)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)

			output, err := findKeyValueStoreByName(ctx, conn, aws.ToString(item.Name))
			if err != nil {
				tflog.Error(ctx, "Reading CloudFront Key Value Store", map[string]any{
					"error": err.Error(),
				})
				continue
			}

			var data keyValueStoreResourceModel
			r.SetResult(ctx, awsClient, request.IncludeResource, &data, &result, func() {
				diags := fwflex.Flatten(ctx, output.KeyValueStore, &data)
				if diags.HasError() {
					result.Diagnostics.Append(diags...)
					yield(result)
					return
				}

				data.ETag = fwflex.StringToFramework(ctx, output.ETag)
				result.DisplayName = aws.ToString(item.Name)
			})

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

type listKeyValueStoreModel struct{}

func listKeyValueStores(ctx context.Context, conn *cloudfront.Client, input *cloudfront.ListKeyValueStoresInput) iter.Seq2[awstypes.KeyValueStore, error] {
	return func(yield func(awstypes.KeyValueStore, error) bool) {
		pages := cloudfront.NewListKeyValueStoresPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.KeyValueStore{}, fmt.Errorf("listing CloudFront Key Value Store resources: %w", err))
				return
			}

			if page.KeyValueStoreList != nil {
				for _, item := range page.KeyValueStoreList.Items {
					if !yield(item, nil) {
						return
					}
				}
			}
		}
	}
}
