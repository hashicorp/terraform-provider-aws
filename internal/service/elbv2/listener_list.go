// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_lb_listener")
func newListenerResourceAsListResource() inttypes.ListResourceForSDK {
	l := listResourceListener{}
	l.SetResourceSchema(resourceListener())
	return &l
}

var _ list.ListResource = &listResourceListener{}

type listResourceListener struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *listResourceListener) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"load_balancer_arn": listschema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Required:    true,
				Description: "ARN of the Load Balancer to list Listeners from.",
			},
		},
	}
}

func (l *listResourceListener) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().ELBV2Client(ctx)

	var query listListenerModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	loadBalancerARN := query.LoadBalancerARN.ValueString()

	tflog.Info(ctx, "Listing Resources", map[string]any{
		"tf_list.request.config.load_balancer_arn": loadBalancerARN,
	})
	stream.Results = func(yield func(list.ListResult) bool) {
		input := elasticloadbalancingv2.DescribeListenersInput{
			LoadBalancerArn: aws.String(loadBalancerARN),
		}

		pages := elasticloadbalancingv2.NewDescribeListenersPaginator(conn, &input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("listing ELB Listener resources: %w", err))
				yield(result)
				return
			}

			var tags map[string][]awstypes.Tag
			if request.IncludeResource {
				arns := make([]string, len(page.Listeners))
				for i, lb := range page.Listeners {
					arns[i] = aws.ToString(lb.ListenerArn)
				}

				tags, err = batchListTags(ctx, conn, arns)
				if err != nil {
					result := fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("listing ELB Listener resource tags: %w", err))
					yield(result)
					return
				}
			}

			for _, item := range page.Listeners {
				arn := aws.ToString(item.ListenerArn)
				ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

				result := request.NewListResult(ctx)
				rd := l.ResourceData()
				rd.SetId(arn)
				rd.Set(names.AttrARN, arn)

				if request.IncludeResource {
					setTagsOut(ctx, tags[arn])
					if err := resourceListenerFlatten(ctx, l.Meta(), &item, rd); err != nil {
						tflog.Error(ctx, "Reading ELB Listener", map[string]any{
							"error": err.Error(),
						})
						continue
					}
				}

				result.DisplayName = arn

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
}

type listListenerModel struct {
	framework.WithRegionModel
	LoadBalancerARN fwtypes.ARN `tfsdk:"load_balancer_arn"`
}
