// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package workspaces

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workspaces"
	awstypes "github.com/aws/aws-sdk-go-v2/service/workspaces/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_workspaces_pool")
func newPoolResourceAsListResource() list.ListResourceWithConfigure {
	return &poolListResource{}
}

var _ list.ListResource = &poolListResource{}

type poolListResource struct {
	resourcePool
	framework.WithList
}

func (l *poolListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().WorkSpacesClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listPools(ctx, conn, &workspaces.DescribeWorkspacesPoolsInput{}) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			result := request.NewListResult(ctx)

			var data resourcePoolModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				result.Diagnostics.Append(l.flatten(ctx, &item, &data)...)
				if result.Diagnostics.HasError() {
					return
				}
			})

			result.DisplayName = aws.ToString(item.PoolName)

			if !yield(result) {
				return
			}
		}
	}
}

func listPools(ctx context.Context, conn *workspaces.Client, input *workspaces.DescribeWorkspacesPoolsInput) iter.Seq2[awstypes.WorkspacesPool, error] {
	return func(yield func(awstypes.WorkspacesPool, error) bool) {
		err := describeWorkspacesPoolsPages(ctx, conn, input, func(page *workspaces.DescribeWorkspacesPoolsOutput, lastPage bool) bool {
			for _, item := range page.WorkspacesPools {
				if !yield(item, nil) {
					return false
				}
			}
			return true
		})
		if err != nil {
			yield(awstypes.WorkspacesPool{}, fmt.Errorf("listing WorkSpaces Pools: %w", err))
		}
	}
}
