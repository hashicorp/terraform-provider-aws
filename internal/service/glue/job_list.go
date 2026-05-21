// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_glue_job")
func newJobResourceAsListResource() inttypes.ListResourceForSDK {
	l := jobListResource{}
	l.SetResourceSchema(resourceJob())
	return &l
}

var _ list.ListResource = &jobListResource{}

type jobListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type listJobModel struct {
	framework.WithRegionModel
}

func (l *jobListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.GlueClient(ctx)

	var query listJobModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing Glue jobs")
	stream.Results = func(yield func(list.ListResult) bool) {
		var input glue.GetJobsInput
		for item, err := range listJobs(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			name := aws.ToString(item.Name)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), name)
			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(name)
			rd.Set(names.AttrName, name)

			if request.IncludeResource {
				if err := resourceJobFlatten(ctx, awsClient, &item, rd); err != nil {
					tflog.Error(ctx, "Flattening Glue job", map[string]any{
						names.AttrName: name,
						"error":        err.Error(),
					})
					continue
				}
			}

			result.DisplayName = name

			l.SetResult(ctx, awsClient, request.IncludeResource, rd, &result)
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

func listJobs(ctx context.Context, conn *glue.Client, input *glue.GetJobsInput) iter.Seq2[awstypes.Job, error] {
	return func(yield func(awstypes.Job, error) bool) {
		pages := glue.NewGetJobsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.Job{}, fmt.Errorf("listing Glue Jobs: %w", err))
				return
			}

			for _, item := range page.Jobs {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
