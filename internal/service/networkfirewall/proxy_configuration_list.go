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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_networkfirewall_proxy_configuration")
func newProxyConfigurationResourceAsListResource() list.ListResourceWithConfigure {
	return &proxyConfigurationListResource{}
}

var _ list.ListResource = &proxyConfigurationListResource{}

type proxyConfigurationListResource struct {
	resourceProxyConfiguration
	framework.WithList
}

func (r *proxyConfigurationListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var query listProxyConfigurationModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		result := request.NewListResult(ctx)
		var input networkfirewall.ListProxyConfigurationsInput
		for summary, err := range listProxyConfigurations(ctx, conn, &input) {
			if err != nil {
				result = fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			arn := aws.ToString(summary.Arn)
			out, err := findProxyConfigurationByARN(ctx, conn, arn)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				result = fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			var data resourceProxyConfigurationModel

			r.SetResult(ctx, r.Meta(), request.IncludeResource, &data, &result, func() {
				if diags := fwflex.Flatten(ctx, out.ProxyConfiguration, &data); diags.HasError() {
					result.Diagnostics.Append(diags...)
					yield(result)
					return
				}

				data.UpdateToken = fwflex.StringToFramework(ctx, out.UpdateToken)
				result.DisplayName = aws.ToString(summary.Name)
			})

			if result.Diagnostics.HasError() {
				result = list.ListResult{Diagnostics: result.Diagnostics}
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

type listProxyConfigurationModel struct {
	framework.WithRegionModel
}

func listProxyConfigurations(ctx context.Context, conn *networkfirewall.Client, input *networkfirewall.ListProxyConfigurationsInput) iter.Seq2[awstypes.ProxyConfigurationMetadata, error] {
	return func(yield func(awstypes.ProxyConfigurationMetadata, error) bool) {
		pages := networkfirewall.NewListProxyConfigurationsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.ProxyConfigurationMetadata{}, fmt.Errorf("listing NetworkFirewall Proxy Configuration resources: %w", err))
				return
			}

			for _, item := range page.ProxyConfigurations {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
