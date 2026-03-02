// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_lb")
func newLoadBalancerResourceAsListResource() inttypes.ListResourceForSDK {
	l := listResourceLoadBalancer{}
	l.SetResourceSchema(resourceLoadBalancer())
	return &l
}

var _ list.ListResource = &listResourceLoadBalancer{}

type listResourceLoadBalancer struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *listResourceLoadBalancer) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().ELBV2Client(ctx)

	var query listLoadBalancerModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing ELB Load Balancer")
	stream.Results = func(yield func(list.ListResult) bool) {
		var input elasticloadbalancingv2.DescribeLoadBalancersInput
		for item, err := range listLoadBalancers(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			arn := aws.ToString(item.LoadBalancerArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(arn)
			rd.Set(names.AttrARN, arn)

			if request.IncludeResource {
				if err := resourceLoadBalancerFlatten(ctx, l.Meta(), &item, rd); err != nil {
					tflog.Error(ctx, "Reading ELB Load Balancer", map[string]any{
						"error": err.Error(),
					})
					continue
				}
			}

			result.DisplayName = aws.ToString(item.LoadBalancerName)

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &result, rd)
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

type listLoadBalancerModel struct {
	framework.WithRegionModel
}

func listLoadBalancers(ctx context.Context, conn *elasticloadbalancingv2.Client, input *elasticloadbalancingv2.DescribeLoadBalancersInput) iter.Seq2[awstypes.LoadBalancer, error] {
	return func(yield func(awstypes.LoadBalancer, error) bool) {
		pages := elasticloadbalancingv2.NewDescribeLoadBalancersPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.LoadBalancer{}, fmt.Errorf("listing ELB Load Balancer resources: %w", err))
				return
			}

			for _, item := range page.LoadBalancers {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
