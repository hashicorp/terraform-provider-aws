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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_networkfirewall_firewall")
func newFirewallResourceAsListResource() inttypes.ListResourceForSDK {
	l := firewallListResource{}
	l.SetResourceSchema(resourceFirewall())
	return &l
}

var _ list.ListResource = &firewallListResource{}

type firewallListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *firewallListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().NetworkFirewallClient(ctx)

	tflog.Info(ctx, "Listing NetworkFirewall Firewall")
	stream.Results = func(yield func(list.ListResult) bool) {
		var input networkfirewall.ListFirewallsInput
		for item, err := range listFirewalls(ctx, conn, &input) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			arn := aws.ToString(item.FirewallArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), arn)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(arn)
			rd.Set(names.AttrARN, arn)

			if request.IncludeResource {
				output, err := findFirewallByARN(ctx, conn, arn)
				if err != nil {
					tflog.Error(ctx, "Reading NetworkFirewall Firewall", map[string]any{
						"error": err,
					})
					continue
				}

				if diags := resourceFirewallFlatten(ctx, rd, l.Meta().EC2Client(ctx), output); diags.HasError() {
					tflog.Error(ctx, "Reading NetworkFirewall Firewall", map[string]any{
						"error": sdkdiag.DiagnosticsString(diags),
					})
					continue
				}
			}

			result.DisplayName = aws.ToString(item.FirewallName)

			l.SetResult(ctx, l.Meta(), request.IncludeResource, rd, &result)
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

func listFirewalls(ctx context.Context, conn *networkfirewall.Client, input *networkfirewall.ListFirewallsInput) iter.Seq2[awstypes.FirewallMetadata, error] {
	return func(yield func(awstypes.FirewallMetadata, error) bool) {
		pages := networkfirewall.NewListFirewallsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.FirewallMetadata{}, fmt.Errorf("listing NetworkFirewall Firewall resources: %w", err))
				return
			}

			for _, item := range page.Firewalls {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
