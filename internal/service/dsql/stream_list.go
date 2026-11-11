// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dsql

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dsql"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dsql/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_dsql_stream")
func newStreamResourceAsListResource() list.ListResourceWithConfigure {
	return &streamListResource{}
}

var _ list.ListResource = &streamListResource{}

type streamListResource struct {
	streamResource
	framework.WithList
}

func (l *streamListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			names.AttrClusterIdentifier: listschema.StringAttribute{
				Required:    true,
				Description: "Identifier of the cluster to list streams for.",
			},
		},
	}
}

func (l *streamListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query streamListModel
	if diags := request.Config.Get(ctx, &query); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	conn := l.Meta().DSQLClient(ctx)

	clusterIdentifier := fwflex.StringValueFromFramework(ctx, query.ClusterIdentifier)

	ctx = tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrClusterIdentifier), clusterIdentifier)
	tflog.Info(ctx, "Listing DSQL Streams")

	stream.Results = func(yield func(list.ListResult) bool) {
		input := dsql.ListStreamsInput{
			ClusterIdentifier: aws.String(clusterIdentifier),
		}
		for item, err := range listStreams(ctx, conn, &input) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			streamIdentifier := aws.ToString(item.StreamIdentifier)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("stream_identifier"), streamIdentifier)

			var output *dsql.GetStreamOutput
			if request.IncludeResource {
				var err error
				output, err = findStreamByTwoPartKey(ctx, conn, clusterIdentifier, streamIdentifier)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data streamResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ClusterIdentifier = fwflex.StringValueToFramework(ctx, clusterIdentifier)
				data.StreamIdentifier = fwflex.StringValueToFramework(ctx, streamIdentifier)

				if request.IncludeResource {
					smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, output, &data))
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = streamIdentifier
			})

			if result.Diagnostics.HasError() {
				yield(list.ListResult{Diagnostics: result.Diagnostics})
				return
			}
			if !yield(result) {
				return
			}
		}
	}
}

type streamListModel struct {
	framework.WithRegionModel
	ClusterIdentifier types.String `tfsdk:"cluster_identifier"`
}

func listStreams(ctx context.Context, conn *dsql.Client, input *dsql.ListStreamsInput) iter.Seq2[awstypes.StreamSummary, error] {
	return func(yield func(awstypes.StreamSummary, error) bool) {
		pages := dsql.NewListStreamsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.StreamSummary{}, fmt.Errorf("listing DSQL Streams: %w", err))
				return
			}

			for _, item := range page.Streams {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
