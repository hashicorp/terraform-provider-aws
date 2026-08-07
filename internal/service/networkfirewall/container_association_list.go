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
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
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
	awsClient := l.Meta()
	conn := awsClient.NetworkFirewallClient(ctx)

	tflog.Info(ctx, "Listing resources")

	input := networkfirewall.ListContainerAssociationsInput{}

	stream.Results = func(yield func(list.ListResult) bool) {
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

			if request.IncludeResource {
				out, err := findContainerAssociationByARN(ctx, conn, arn)
				if err != nil {
					tflog.Error(ctx, "Reading Network Firewall Container Association", map[string]any{
						"error": err.Error(),
					})
					continue
				}

				l.SetResult(ctx, awsClient, true, &data, &result, func() {
					result.Diagnostics.Append(l.flatten(ctx, out, &data)...)
					if result.Diagnostics.HasError() {
						return
					}
					data.ContainerAssociationARN = fwflex.StringToFramework(ctx, item.Arn)
					result.DisplayName = aws.ToString(item.Name)
				})
			} else {
				l.SetResult(ctx, awsClient, false, &data, &result, func() {
					data.ContainerAssociationARN = fwflex.StringToFramework(ctx, item.Arn)
					result.DisplayName = aws.ToString(item.Name)
				})
			}

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

func listContainerAssociations(ctx context.Context, conn *networkfirewall.Client, input *networkfirewall.ListContainerAssociationsInput) iter.Seq2[awstypes.ContainerAssociationSummary, error] {
	return func(yield func(awstypes.ContainerAssociationSummary, error) bool) {
		pages := networkfirewall.NewListContainerAssociationsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.ContainerAssociationSummary{}, fmt.Errorf("listing Network Firewall Container Associations: %w", err))
				return
			}

			for _, item := range page.ContainerAssociations {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
