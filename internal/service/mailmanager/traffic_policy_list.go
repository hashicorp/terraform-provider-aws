// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mailmanager

import (
	"context"
	"iter"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mailmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mailmanager/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_mailmanager_traffic_policy")
func newTrafficPolicyResourceAsListResource() list.ListResourceWithConfigure {
	return &trafficPolicyListResource{}
}

var _ list.ListResource = &trafficPolicyListResource{}

type trafficPolicyListResource struct {
	trafficPolicyResource
	framework.WithList
}

func (l *trafficPolicyListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().MailManagerClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		var input mailmanager.ListTrafficPoliciesInput
		for item, err := range listTrafficPolicies(ctx, conn, &input) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			id := aws.ToString(item.TrafficPolicyId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)
			var out *mailmanager.GetTrafficPolicyOutput
			if request.IncludeResource {
				out, err = findTrafficPolicyByID(ctx, conn, id)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)
			var data trafficPolicyResourceModel

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.DefaultAction = fwtypes.StringEnumValue(item.DefaultAction)
				data.ID = types.StringValue(id)
				data.Name = types.StringPointerValue(item.TrafficPolicyName)

				if request.IncludeResource {
					result.Diagnostics.Append(flex.Flatten(ctx, out, &data, flex.WithFieldNamePrefix("TrafficPolicy"))...)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = aws.ToString(item.TrafficPolicyName)
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

func listTrafficPolicies(ctx context.Context, conn *mailmanager.Client, input *mailmanager.ListTrafficPoliciesInput) iter.Seq2[awstypes.TrafficPolicy, error] {
	return func(yield func(awstypes.TrafficPolicy, error) bool) {
		pages := mailmanager.NewListTrafficPoliciesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.TrafficPolicy{}, smarterr.NewError(err))
				return
			}

			for _, item := range page.TrafficPolicies {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
