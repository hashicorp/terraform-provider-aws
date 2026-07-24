// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package neptunegraph

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/neptunegraph"
	awstypes "github.com/aws/aws-sdk-go-v2/service/neptunegraph/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_neptunegraph_graph")
func newGraphResourceAsListResource() list.ListResourceWithConfigure {
	return &listResourceGraph{}
}

var _ list.ListResource = &listResourceGraph{}

type listResourceGraph struct {
	graphResource
	framework.WithList
}

type listGraphModel struct {
	framework.WithRegionModel
}

func (l *listResourceGraph) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query listGraphModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	conn := l.Meta().NeptuneGraphClient(ctx)

	tflog.Info(ctx, "Listing Neptune Analytics Graphs")

	stream.Results = func(yield func(list.ListResult) bool) {
		var input neptunegraph.ListGraphsInput
		for item, err := range listGraphs(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			id := aws.ToString(item.Id)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)

			var data graphResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				if request.IncludeResource {
					graph, err := findGraphByID(ctx, conn, id)
					if err != nil {
						tflog.Error(ctx, "Reading Neptune Analytics Graph", map[string]any{
							"error": err.Error(),
						})
						return
					}
					result.Diagnostics.Append(fwflex.Flatten(ctx, graph, &data)...)
				} else {
					data.ID = fwflex.StringToFramework(ctx, item.Id)
				}
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

func listGraphs(ctx context.Context, conn *neptunegraph.Client, input *neptunegraph.ListGraphsInput) iter.Seq2[awstypes.GraphSummary, error] {
	return func(yield func(awstypes.GraphSummary, error) bool) {
		pages := neptunegraph.NewListGraphsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.GraphSummary{}, fmt.Errorf("listing Neptune Analytics Graph resources: %w", err))
				return
			}

			for _, item := range page.Graphs {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
