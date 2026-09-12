// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_ec2_application_status_check")
func newApplicationStatusCheckResourceAsListResource() list.ListResourceWithConfigure {
	return &applicationStatusCheckListResource{}
}

var _ list.ListResource = &applicationStatusCheckListResource{}

type applicationStatusCheckListResource struct {
	applicationStatusCheckResource
	framework.WithList
}

func (l *applicationStatusCheckListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().EC2Client(ctx)

	tflog.Info(ctx, "Listing EC2 Application Status Checks")

	stream.Results = func(yield func(list.ListResult) bool) {
		var input ec2.DescribeApplicationStatusChecksInput
		for item, err := range listApplicationStatusChecks(ctx, conn, &input) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			id := aws.ToString(item.ApplicationStatusCheckId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)
			result := request.NewListResult(ctx)
			var data applicationStatusCheckResourceModel
			tags := keyValueTags(ctx, item.Tags)

			setTagsOut(ctx, item.Tags)
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				result.Diagnostics.Append(l.flatten(ctx, &item, &data)...)
				if v, ok := tags["Name"]; ok {
					result.DisplayName = fmt.Sprintf("%s (%s)", v.ValueString(), id)
				} else {
					result.DisplayName = id
				}
			})
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

func listApplicationStatusChecks(ctx context.Context, conn *ec2.Client, input *ec2.DescribeApplicationStatusChecksInput) iter.Seq2[awstypes.ApplicationStatusCheckResponseObject, error] {
	return func(yield func(awstypes.ApplicationStatusCheckResponseObject, error) bool) {
		err := describeApplicationStatusChecksPages(ctx, conn, input, func(page *ec2.DescribeApplicationStatusChecksOutput, _ bool) bool {
			for _, item := range page.ApplicationStatusChecks {
				if !yield(item, nil) {
					return false
				}
			}
			return true
		})
		if err != nil {
			yield(awstypes.ApplicationStatusCheckResponseObject{}, fmt.Errorf("listing EC2 Application Status Checks: %w", err))
		}
	}
}
