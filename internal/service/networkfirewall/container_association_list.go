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
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_networkfirewall_container_association")
func newContainerAssociationResourceAsListResource() list.ListResourceWithConfigure {
	return &containerAssociationListResource{}
}

var _ list.ListResource = &containerAssociationListResource{}

type containerAssociationListResource struct {
	containerAssociationResource
	framework.WithList
}

func (l *containerAssociationListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().NetworkFirewallClient(ctx)

	var query listContainerAssociationModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing resources")

	stream.Results = func(yield func(list.ListResult) bool) {
		var input networkfirewall.ListContainerAssociationsInput
		for item, err := range listContainerAssociations(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			arn := aws.ToString(item.Arn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("container_association_arn"), arn)

			result := request.NewListResult(ctx)

			var data containerAssociationResourceModel

			l.SetResult(ctx, l.Meta(), true, &data, &result, func() {
				data.ContainerAssociationARN = fwflex.StringValueToFramework(ctx, arn)

				if request.IncludeResource {
					out, err := findContainerAssociationByARN(ctx, conn, arn)
					if retry.NotFound(err) {
						return
					}
					if err != nil {
						result = fwdiag.NewListResultErrorDiagnostic(err)
						return
					}

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

type listContainerAssociationModel struct {
	framework.WithRegionModel
}

func listContainerAssociations(ctx context.Context, conn *networkfirewall.Client, input *networkfirewall.ListContainerAssociationsInput, optFns ...func(*networkfirewall.Options)) iter.Seq2[awstypes.ContainerAssociationSummary, error] {
	return tfiter.ConcatValuesWithError(listContainerAssociationPages(ctx, conn, input, optFns...))
}

func listContainerAssociationPages(ctx context.Context, conn *networkfirewall.Client, input *networkfirewall.ListContainerAssociationsInput, optFns ...func(*networkfirewall.Options)) iter.Seq2[[]awstypes.ContainerAssociationSummary, error] {
	return func(yield func([]awstypes.ContainerAssociationSummary, error) bool) {
		pages := networkfirewall.NewListContainerAssociationsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing Network Firewall Container Associations: %w", err))
				return
			}

			if !yield(page.ContainerAssociations, nil) {
				return
			}
		}
	}
}
