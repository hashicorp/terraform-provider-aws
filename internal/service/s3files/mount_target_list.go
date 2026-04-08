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

// @FrameworkListResource("aws_s3files_mount_target")
func newMountTargetResourceAsListResource() list.ListResourceWithConfigure {
	return &listResourceMountTarget{}
}

var _ list.ListResource = &listResourceMountTarget{}

type listResourceMountTarget struct {
	mountTargetResource
	framework.WithList
}

func (r *listResourceMountTarget) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := r.Meta()
	conn := awsClient.S3FilesClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		result := request.NewListResult(ctx)
		for mt, err := range listMountTargets(ctx, conn) {
			if err != nil {
				result = fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			var data mountTargetResourceModel
			r.SetResult(ctx, awsClient, request.IncludeResource, &data, &result, func() {
				data.ID = types.StringPointerValue(mt.MountTargetId)
				data.MountTargetId = types.StringPointerValue(mt.MountTargetId)
				result.DisplayName = data.MountTargetId.ValueString()
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

func listMountTargets(ctx context.Context, conn *s3files.Client) iter.Seq2[awstypes.ListMountTargetsDescription, error] {
	return func(yield func(awstypes.ListMountTargetsDescription, error) bool) {
		pages := s3files.NewListMountTargetsPaginator(conn, &s3files.ListMountTargetsInput{})
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.ListMountTargetsDescription{}, fmt.Errorf("listing S3 Files Mount Targets: %w", err))
				return
			}
			for _, mt := range page.MountTargets {
				if !yield(mt, nil) {
					return
				}
			}
		}
	}
}
