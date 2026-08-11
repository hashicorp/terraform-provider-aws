// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package osis

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/osis"
	awstypes "github.com/aws/aws-sdk-go-v2/service/osis/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_osis_pipeline")
func newPipelineResourceAsListResource() list.ListResourceWithConfigure {
	return &pipelineListResource{}
}

var _ list.ListResource = &pipelineListResource{}

type pipelineListResource struct {
	pipelineResource
	framework.WithList
}

func (l *pipelineListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().OpenSearchIngestionClient(ctx)

	var query listPipelineModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		var input osis.ListPipelinesInput
		for item, err := range listPipelines(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			name := aws.ToString(item.PipelineName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), name)

			result := request.NewListResult(ctx)

			var data pipelineResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				pipeline, err := findPipelineByName(ctx, conn, name)
				if retry.NotFound(err) {
					return
				}
				if err != nil {
					result.Diagnostics.AddError("reading OpenSearch Ingestion Pipeline", err.Error())
					return
				}

				smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, pipeline, &data))
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = name
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listPipelineModel struct {
	framework.WithRegionModel
}

func listPipelines(ctx context.Context, conn *osis.Client, input *osis.ListPipelinesInput) iter.Seq2[awstypes.PipelineSummary, error] {
	return func(yield func(awstypes.PipelineSummary, error) bool) {
		pages := osis.NewListPipelinesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.PipelineSummary{}, fmt.Errorf("listing OpenSearch Ingestion Pipeline resources: %w", err))
				return
			}

			for _, item := range page.Pipelines {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
