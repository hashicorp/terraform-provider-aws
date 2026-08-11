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
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_opensearchserverless_collection_group")
func newCollectionGroupResourceAsListResource() list.ListResourceWithConfigure {
	return &collectionGroupListResource{}
}

var _ list.ListResource = &collectionGroupListResource{}

type collectionGroupListResource struct {
	collectionGroupResource
	framework.WithList
}

func (l *collectionGroupListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.OpenSearchServerlessClient(ctx)

	var query listCollectionGroupModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		var input opensearchserverless.ListCollectionGroupsInput

		for item, err := range listCollectionGroups(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			collectionGroupID := aws.ToString(item.Id)
			ctx = tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), collectionGroupID)

			result := request.NewListResult(ctx)
			var data collectionGroupResourceModel
			data.ID = fwflex.StringToFramework(ctx, item.Id)
			data.Name = fwflex.StringToFramework(ctx, item.Name)

			l.SetResult(ctx, awsClient, request.IncludeResource, &data, &result, func() {
				if request.IncludeResource {
					collectionGroup, err := findCollectionGroup(ctx, conn, &opensearchserverless.BatchGetCollectionGroupInput{
						Ids: []string{collectionGroupID},
					})
					if err != nil {
						result = fwdiag.NewListResultErrorDiagnostic(err)
						return
					}

					result.Diagnostics.Append(fwflex.Flatten(ctx, collectionGroup, &data, fwflex.WithIgnoredFieldNamesAppend("CreatedDate"))...)
					if result.Diagnostics.HasError() {
						return
					}

					data.CreatedDate = timetypes.NewRFC3339ValueMust(flex.Int64ToRFC3339StringValue(collectionGroup.CreatedDate))
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

type listCollectionGroupModel struct {
	framework.WithRegionModel
}

func listCollectionGroups(ctx context.Context, conn *opensearchserverless.Client, input *opensearchserverless.ListCollectionGroupsInput) iter.Seq2[awstypes.CollectionGroupSummary, error] {
	return func(yield func(awstypes.CollectionGroupSummary, error) bool) {
		pages := opensearchserverless.NewListCollectionGroupsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.CollectionGroupSummary{}, fmt.Errorf("listing OpenSearch Serverless Collection Groups: %w", err))
				return
			}

			for _, item := range page.CollectionGroupSummaries {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
