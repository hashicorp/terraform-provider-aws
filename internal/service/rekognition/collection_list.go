// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rekognition

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_rekognition_collection")
func newCollectionResourceAsListResource() list.ListResourceWithConfigure {
	return &collectionListResource{}
}

var _ list.ListResource = &collectionListResource{}

type collectionListResource struct {
	collectionResource
	framework.WithList
}

func (l *collectionListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().RekognitionClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		for collectionID, err := range listCollectionIDs(ctx, conn, &rekognition.ListCollectionsInput{}) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), collectionID)

			result := request.NewListResult(ctx)

			var data collectionResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.CollectionID = flex.StringToFramework(ctx, &collectionID)
				data.ID = flex.StringToFramework(ctx, &collectionID)

				if request.IncludeResource {
					out, err := findCollectionByID(ctx, conn, collectionID)
					if err != nil {
						result.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
						return
					}

					smerr.AddEnrich(ctx, &result.Diagnostics, flex.Flatten(ctx, out, &data, flex.WithFieldNamePrefix("Collection")))
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = collectionID
			})

			if !yield(result) {
				return
			}
		}
	}
}

func listCollectionIDs(ctx context.Context, conn *rekognition.Client, input *rekognition.ListCollectionsInput) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		pages := rekognition.NewListCollectionsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield("", fmt.Errorf("listing Rekognition Collections: %w", err))
				return
			}

			for _, id := range page.CollectionIds {
				if !yield(id, nil) {
					return
				}
			}
		}
	}
}
