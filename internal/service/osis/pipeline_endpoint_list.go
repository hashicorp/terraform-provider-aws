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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_osis_pipeline_endpoint")
func newPipelineEndpointResourceAsListResource() list.ListResourceWithConfigure {
	return &pipelineEndpointListResource{}
}

var _ list.ListResource = &pipelineEndpointListResource{}

type pipelineEndpointListResource struct {
	pipelineEndpointResource
	framework.WithList
}

func (l *pipelineEndpointListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().OpenSearchIngestionClient(ctx)

	var query listPipelineEndpointModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		var input osis.ListPipelineEndpointsInput
		for item, err := range listPipelineEndpoints(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			endpointID := aws.ToString(item.EndpointId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), endpointID)

			result := request.NewListResult(ctx)

			var data pipelineEndpointResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				result.Diagnostics.Append(l.flatten(ctx, &item, &data)...)
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = endpointID
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listPipelineEndpointModel struct {
	framework.WithRegionModel
}

func listPipelineEndpoints(ctx context.Context, conn *osis.Client, input *osis.ListPipelineEndpointsInput) iter.Seq2[awstypes.PipelineEndpoint, error] {
	return func(yield func(awstypes.PipelineEndpoint, error) bool) {
		pages := osis.NewListPipelineEndpointsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.PipelineEndpoint{}, fmt.Errorf("listing OpenSearch Ingestion Pipeline Endpoint resources: %w", err))
				return
			}

			for _, item := range page.PipelineEndpoints {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
