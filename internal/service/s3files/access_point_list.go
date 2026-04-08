// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3files

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/service/s3files"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3files/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
)

// @FrameworkListResource("aws_s3files_access_point")
func newAccessPointResourceAsListResource() list.ListResourceWithConfigure {
	return &listResourceAccessPoint{}
}

var _ list.ListResource = &listResourceAccessPoint{}

type listResourceAccessPoint struct {
	accessPointResource
	framework.WithList
}

func (r *listResourceAccessPoint) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := r.Meta()
	conn := awsClient.S3FilesClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		result := request.NewListResult(ctx)
		for ap, err := range listAccessPoints(ctx, conn) {
			if err != nil {
				result = fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			var data accessPointResourceModel
			r.SetResult(ctx, awsClient, request.IncludeResource, &data, &result, func() {
				data.ID = types.StringPointerValue(ap.AccessPointId)
				data.AccessPointId = types.StringPointerValue(ap.AccessPointId)
				data.AccessPointArn = types.StringPointerValue(ap.AccessPointArn)
				result.DisplayName = data.AccessPointId.ValueString()
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

func listAccessPoints(ctx context.Context, conn *s3files.Client) iter.Seq2[awstypes.ListAccessPointsDescription, error] {
	return func(yield func(awstypes.ListAccessPointsDescription, error) bool) {
		pages := s3files.NewListAccessPointsPaginator(conn, &s3files.ListAccessPointsInput{})
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.ListAccessPointsDescription{}, fmt.Errorf("listing S3 Files Access Points: %w", err))
				return
			}
			for _, ap := range page.AccessPoints {
				if !yield(ap, nil) {
					return
				}
			}
		}
	}
}
