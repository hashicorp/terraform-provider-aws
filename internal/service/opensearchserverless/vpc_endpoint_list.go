// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_opensearchserverless_vpc_endpoint")
func newVPCEndpointResourceAsListResource() list.ListResourceWithConfigure {
	return &vpcEndpointListResource{}
}

var _ list.ListResource = &vpcEndpointListResource{}

type vpcEndpointListResource struct {
	vpcEndpointResource
	framework.WithList
}

func (l *vpcEndpointListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.OpenSearchServerlessClient(ctx)

	var query listVPCEndpointModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		var input opensearchserverless.ListVpcEndpointsInput

		for summary, err := range listVPCEndpoints(ctx, conn, &input) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			id := aws.ToString(summary.Id)
			ctx = tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			var output *awstypes.VpcEndpointDetail
			if request.IncludeResource {
				var err error
				output, err = findVPCEndpointByID(ctx, conn, id)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)
			var data vpcEndpointResourceModel
			data.ID = fwflex.StringToFramework(ctx, summary.Id)
			data.Name = fwflex.StringToFramework(ctx, summary.Name)

			l.SetResult(ctx, awsClient, request.IncludeResource, &data, &result, func() {
				if request.IncludeResource {
					result.Diagnostics.Append(l.flatten(ctx, awsClient, output, &data)...)
					if result.Diagnostics.HasError() {
						return
					}
				}

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

type listVPCEndpointModel struct {
	framework.WithRegionModel
}

func listVPCEndpoints(ctx context.Context, conn *opensearchserverless.Client, input *opensearchserverless.ListVpcEndpointsInput) iter.Seq2[awstypes.VpcEndpointSummary, error] {
	return func(yield func(awstypes.VpcEndpointSummary, error) bool) {
		pages := opensearchserverless.NewListVpcEndpointsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.VpcEndpointSummary{}, fmt.Errorf("listing OpenSearch Serverless VPC Endpoints: %w", err))
				return
			}

			for _, summary := range page.VpcEndpointSummaries {
				if !yield(summary, nil) {
					return
				}
			}
		}
	}
}
