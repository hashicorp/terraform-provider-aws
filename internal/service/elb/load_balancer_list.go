// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elb

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_elb")
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
	conn := l.Meta().ELBClient(ctx)

	var query listLoadBalancerModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing ELB Classic Load Balancer")
	stream.Results = func(yield func(list.ListResult) bool) {
		var input elasticloadbalancing.DescribeLoadBalancersInput
		for item, err := range listLoadBalancers(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			name := aws.ToString(item.LoadBalancerName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), name)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(name)
			rd.Set(names.AttrName, name)

			if request.IncludeResource {
				lbAttrs, err := findLoadBalancerAttributesByName(ctx, conn, name)
				if retry.NotFound(err) {
					tflog.Warn(ctx, "Resource disappeared during listing, skipping")
					continue
				}
				if err != nil {
					tflog.Error(ctx, "Reading ELB Classic Load Balancer attributes", map[string]any{
						"error": err.Error(),
					})
					continue
				}

				if err := resourceLoadBalancerFlatten(ctx, l.Meta(), &item, lbAttrs, rd); err != nil {
					tflog.Error(ctx, "Reading ELB Classic Load Balancer", map[string]any{
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

func listLoadBalancers(ctx context.Context, conn *elasticloadbalancing.Client, input *elasticloadbalancing.DescribeLoadBalancersInput) iter.Seq2[awstypes.LoadBalancerDescription, error] {
	return func(yield func(awstypes.LoadBalancerDescription, error) bool) {
		pages := elasticloadbalancing.NewDescribeLoadBalancersPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.LoadBalancerDescription{}, fmt.Errorf("listing ELB Classic Load Balancer resources: %w", err))
				return
			}

			for _, item := range page.LoadBalancerDescriptions {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
