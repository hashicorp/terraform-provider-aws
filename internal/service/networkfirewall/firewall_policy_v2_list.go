// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkfirewall

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_networkfirewall_firewall_policy_v2")
func newFirewallPolicyV2ResourceAsListResource() list.ListResourceWithConfigure {
	return &firewallPolicyV2ListResource{}
}

var _ list.ListResource = &firewallPolicyV2ListResource{}

type firewallPolicyV2ListResource struct {
	firewallPolicyV2Resource
	framework.WithList
}

func (l *firewallPolicyV2ListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().NetworkFirewallClient(ctx)

	tflog.Info(ctx, "Listing resources")

	stream.Results = func(yield func(list.ListResult) bool) {
		var input networkfirewall.ListFirewallPoliciesV2Input
		for item, err := range listFirewallPoliciesV2(ctx, conn, &input) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			arn := aws.ToString(item.Arn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			var out *networkfirewall.DescribeFirewallPolicyV2Output
			if request.IncludeResource {
				out, err = findFirewallPolicyV2ByARN(ctx, conn, arn)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data firewallPolicyV2ResourceModel

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ARN = fwflex.StringValueToFramework(ctx, arn)

				if request.IncludeResource {
					result.Diagnostics.Append(l.flatten(ctx, out, &data)...)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = aws.ToString(item.Name)
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

func listFirewallPoliciesV2(ctx context.Context, conn *networkfirewall.Client, input *networkfirewall.ListFirewallPoliciesV2Input) iter.Seq2[awstypes.FirewallPolicyV2Metadata, error] {
	return tfiter.ConcatValuesWithError(listFirewallPolicyV2Pages(ctx, conn, input))
}

func listFirewallPolicyV2Pages(ctx context.Context, conn *networkfirewall.Client, input *networkfirewall.ListFirewallPoliciesV2Input) iter.Seq2[[]awstypes.FirewallPolicyV2Metadata, error] {
	return func(yield func([]awstypes.FirewallPolicyV2Metadata, error) bool) {
		pages := networkfirewall.NewListFirewallPoliciesV2Paginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(nil, fmt.Errorf("listing Network Firewall Firewall Policies V2: %w", err))
				return
			}

			if !yield(page.FirewallPoliciesV2, nil) {
				return
			}
		}
	}
}
