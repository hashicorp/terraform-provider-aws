// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/batch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/batch/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_batch_job_definition")
func newJobDefinitionResourceAsListResource() inttypes.ListResourceForSDK {
	l := jobDefinitionListResource{}
	l.SetResourceSchema(resourceJobDefinition())
	return &l
}

type jobDefinitionListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type jobDefinitionListResourceModel struct {
	framework.WithRegionModel
}

func (l *jobDefinitionListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.BatchClient(ctx)

	var query jobDefinitionListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input batch.DescribeJobDefinitionsInput

	tflog.Info(ctx, "Listing Batch job definitions")

	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listJobDefinitions(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			if status := aws.ToString(item.Status); status == jobDefinitionStatusInactive {
				continue
			}

			arn := aws.ToString(item.JobDefinitionArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), arn)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(arn)

			tflog.Info(ctx, "Reading Batch job definition")
			diags := resourceJobDefinitionFlatten(ctx, &item, rd)
			if diags.HasError() {
				result = fwdiag.NewListResultSDKDiagnostics(diags)
				yield(result)
				return
			}

			result.DisplayName = aws.ToString(item.JobDefinitionName)

			l.SetResult(ctx, awsClient, request.IncludeResource, &result, rd)
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

func listJobDefinitions(ctx context.Context, conn *batch.Client, input *batch.DescribeJobDefinitionsInput) iter.Seq2[awstypes.JobDefinition, error] {
	return func(yield func(awstypes.JobDefinition, error) bool) {
		pages := batch.NewDescribeJobDefinitionsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.JobDefinition{}, fmt.Errorf("listing Batch Job Definitions: %w", err))
				return
			}

			for _, item := range page.JobDefinitions {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
