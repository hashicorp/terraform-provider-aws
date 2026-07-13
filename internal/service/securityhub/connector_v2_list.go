// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_securityhub_connector_v2")
func newConnectorV2ResourceAsListResource() list.ListResourceWithConfigure {
	return &connectorV2ListResource{}
}

var _ list.ListResource = &connectorV2ListResource{}

type connectorV2ListResource struct {
	connectorV2Resource
	framework.WithList
}

func (l *connectorV2ListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().SecurityHubClient(ctx)

	var query listConnectorV2Model
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		var input securityhub.ListConnectorsV2Input
		for item, err := range listConnectorV2s(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			arn := aws.ToString(item.ConnectorArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			connectorID := aws.ToString(item.ConnectorId)
			output, err := findConnectorV2ByID(ctx, conn, connectorID)
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			result := request.NewListResult(ctx)

			var data connectorV2ResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				result.Diagnostics.Append(l.flatten(ctx, output, &data)...)
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = aws.ToString(item.Name)
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listConnectorV2Model struct {
	framework.WithRegionModel
}

func listConnectorV2s(ctx context.Context, conn *securityhub.Client, input *securityhub.ListConnectorsV2Input) iter.Seq2[awstypes.ConnectorSummary, error] {
	return func(yield func(awstypes.ConnectorSummary, error) bool) {
		var stopped bool
		err := listConnectorsV2Pages(ctx, conn, input, func(page *securityhub.ListConnectorsV2Output, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, item := range page.Connectors {
				if !yield(item, nil) {
					stopped = true
					return false
				}
			}

			return !lastPage
		})

		if !stopped && err != nil {
			yield(inttypes.Zero[awstypes.ConnectorSummary](), fmt.Errorf("listing Security Hub V2 Connectors: %w", err))
			return
		}
	}
}
