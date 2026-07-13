// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package autoscaling

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	awstypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
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
// @SDKListResource("aws_autoscaling_group")
func newGroupResourceAsListResource() inttypes.ListResourceForSDK {
	l := groupListResource{}
	l.SetResourceSchema(resourceGroup())
	return &l
}

var _ list.ListResource = &groupListResource{}

type groupListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type listGroupModel struct {
	framework.WithRegionModel
}

func (l *groupListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.AutoScalingClient(ctx)

	var query listGroupModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing Auto Scaling Groups")

	stream.Results = func(yield func(list.ListResult) bool) {
		input := autoscaling.DescribeAutoScalingGroupsInput{}
		for g, err := range listAutoScalingGroups(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			name := aws.ToString(g.AutoScalingGroupName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), name)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(name)
			rd.Set(names.AttrName, g.AutoScalingGroupName)

			if request.IncludeResource {
				diags := resourceGroupFlatten(ctx, awsClient, &g, rd)
				if diags.HasError() || rd.Id() == "" {
					tflog.Error(ctx, "Reading Auto Scaling Group", map[string]any{
						names.AttrName: name,
						"diags":        sdkdiag.DiagnosticsString(diags),
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

func listAutoScalingGroups(ctx context.Context, conn *autoscaling.Client, input *autoscaling.DescribeAutoScalingGroupsInput) iter.Seq2[awstypes.AutoScalingGroup, error] {
	return func(yield func(awstypes.AutoScalingGroup, error) bool) {
		pages := autoscaling.NewDescribeAutoScalingGroupsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.AutoScalingGroup{}, fmt.Errorf("listing Auto Scaling Groups: %w", err))
				return
			}

			for _, g := range page.AutoScalingGroups {
				if !yield(g, nil) {
					return
				}
			}
		}
	}
}
