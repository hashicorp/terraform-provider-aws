// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_ssoadmin_instance")
func newInstanceResourceAsListResource() list.ListResourceWithConfigure {
	return &instanceListResource{}
}

var _ list.ListResource = &instanceListResource{}

type instanceListResource struct {
	instanceResource
	framework.WithList
}

func (l *instanceListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().SSOAdminClient(ctx)
	tflog.Info(ctx, "Listing SSO Admin Instances")
	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listInstances(ctx, conn, &ssoadmin.ListInstancesInput{}) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}
			arn := aws.ToString(item.InstanceArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)
			var output *ssoadmin.DescribeInstanceOutput
			if request.IncludeResource {
				output, err = findInstanceByARN(ctx, conn, arn)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}
			result := request.NewListResult(ctx)
			var data instanceResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ARN = fwtypes.ARNValue(arn)
				if request.IncludeResource {
					result.Diagnostics.Append(l.flatten(ctx, output, &data)...)
				}
				result.DisplayName = aws.ToString(item.Name)
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

func listInstances(ctx context.Context, conn *ssoadmin.Client, input *ssoadmin.ListInstancesInput) iter.Seq2[awstypes.InstanceMetadata, error] {
	return func(yield func(awstypes.InstanceMetadata, error) bool) {
		pages := ssoadmin.NewListInstancesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.InstanceMetadata{}, fmt.Errorf("listing SSO Admin Instance resources: %w", err))
				return
			}
			for _, item := range page.Instances {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
