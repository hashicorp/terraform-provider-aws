// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_opensearchserverless_collection")
func newCollectionResourceAsListResource() list.ListResourceWithConfigure {
	return &collectionListResource{}
}

var _ list.ListResource = &collectionListResource{}

type collectionListResource struct {
	collectionResource
	framework.WithList
}

func (r *collectionListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := r.Meta()
	conn := awsClient.OpenSearchServerlessClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		result := request.NewListResult(ctx)
		var input opensearchserverless.ListCollectionsInput
		for collectionSummary, err := range listCollections(ctx, conn, &input) {
			if err != nil {
				result = fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			collectionID := aws.ToString(collectionSummary.Id)
			collection, err := findCollectionByID(ctx, conn, collectionID)
			if err != nil {
				result = fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			var data collectionResourceModel

			r.SetResult(ctx, awsClient, request.IncludeResource, &data, &result, func() {
				if diags := fwflex.Flatten(ctx, collection, &data); diags.HasError() {
					result.Diagnostics.Append(diags...)
					yield(result)
					return
				}

				result.DisplayName = data.Name.ValueString()
			})

			if result.Diagnostics.HasError() {
				result = list.ListResult{Diagnostics: result.Diagnostics}
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

func listCollections(ctx context.Context, conn *opensearchserverless.Client, input *opensearchserverless.ListCollectionsInput) iter.Seq2[awstypes.CollectionSummary, error] {
	return func(yield func(awstypes.CollectionSummary, error) bool) {
		pages := opensearchserverless.NewListCollectionsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.CollectionSummary{}, fmt.Errorf("listing OpenSearch Serverless Collections: %w", err))
				return
			}

			for _, collectionSummary := range page.CollectionSummaries {
				if !yield(collectionSummary, nil) {
					return
				}
			}
		}
	}
}
