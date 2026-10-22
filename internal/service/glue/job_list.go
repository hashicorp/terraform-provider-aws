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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @SDKListResource("aws_glue_job")
func newJobResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceJob(), framework.SDKListSpec[awstypes.Job, framework.RegionQuery]{
		Items: func(ctx context.Context, c *conns.AWSClient, _ framework.RegionQuery) (iter.Seq2[awstypes.Job, error], diag.Diagnostics) {
			var input glue.GetJobsInput

			return listJobs(ctx, c.GlueClient(ctx), &input), nil
		},
		ID: func(j awstypes.Job, _ framework.RegionQuery) string {
			return aws.ToString(j.Name)
		},
		Flatten: resourceJobFlatten,
	})
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
