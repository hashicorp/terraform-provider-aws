// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package codepipeline

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_codepipeline")
func newPipelineResourceAsListResource() inttypes.ListResourceForSDK {
	l := pipelineListResource{}
	l.SetResourceSchema(resourcePipeline())
	return &l
}

var _ list.ListResource = &pipelineListResource{}

type pipelineListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type pipelineListResourceModel struct {
	framework.WithRegionModel
}

func (l *pipelineListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.CodePipelineClient(ctx)

	var query pipelineListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input codepipeline.ListPipelinesInput

	tflog.Info(ctx, "Listing CodePipeline pipelines")

	stream.Results = func(yield func(list.ListResult) bool) {
		pages := codepipeline.NewListPipelinesPaginator(conn, &input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("listing CodePipeline Pipelines: %w", err))
				yield(result)
				return
			}

			for _, pipeline := range page.Pipelines {
				name := aws.ToString(pipeline.Name)
				ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), name)

				result := request.NewListResult(ctx)
				rd := l.ResourceData()
				rd.SetId(name)
				rd.Set(names.AttrName, name)

				if request.IncludeResource {
					output, err := findPipelineByName(ctx, conn, name)
					if err != nil {
						result := fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("reading CodePipeline Pipeline (%s): %w", name, err))
						if !yield(result) {
							return
						}
						continue
					}

					if diags := resourcePipelineFlatten(rd, output); diags.HasError() {
						tflog.Error(ctx, "Error reading CodePipeline Pipeline", map[string]any{
							"error": sdkdiag.DiagnosticsString(diags),
						})
						continue
					}
				}

				result.DisplayName = name

				l.SetResult(ctx, awsClient, request.IncludeResource, rd, &result)
				if result.Diagnostics.HasError() {
					tflog.Error(ctx, "Error setting result for CodePipeline Pipeline", map[string]any{
						"error": result.Diagnostics,
					})
					continue
				}

				if !yield(result) {
					return
				}
			}
		}
	}
}
