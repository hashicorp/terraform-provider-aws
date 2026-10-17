// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package acm

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/acm/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_acm_acme_endpoint")
func newACMEEndpointResourceAsListResource() list.ListResourceWithConfigure {
	return &acmeEndpointListResource{}
}

var _ list.ListResource = &acmeEndpointListResource{}

type acmeEndpointListResource struct {
	acmeEndpointResource
	framework.WithList
}

func (l *acmeEndpointListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query acmeEndpointListModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	awsClient := l.Meta()
	conn := awsClient.ACMClient(ctx)

	tflog.Info(ctx, "Listing ACM ACME Endpoints")

	stream.Results = func(yield func(list.ListResult) bool) {
		var input acm.ListAcmeEndpointsInput
		for item, err := range listACMEEndpoints(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			arn := aws.ToString(item.AcmeEndpointArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			result := request.NewListResult(ctx)

			var data acmeEndpointResourceModel
			l.SetResult(ctx, awsClient, request.IncludeResource, &data, &result, func() {
				result.Diagnostics.Append(l.flattenSummary(ctx, &item, &data)...)
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = arn
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

// flattenSummary maps an AcmeEndpointSummary onto the resource model. The summary carries
// every field DescribeAcmeEndpoint returns, so no per-item Describe call is needed.
func (r *acmeEndpointResource) flattenSummary(ctx context.Context, summary *awstypes.AcmeEndpointSummary, data *acmeEndpointResourceModel) (diags diag.Diagnostics) {
	diags.Append(fwflex.Flatten(ctx, summary, data, fwflex.WithFieldNamePrefix("AcmeEndpoint"), fwflex.WithIgnoredFieldNamesAppend(certificateTagsFieldName))...)
	if diags.HasError() {
		return diags
	}

	data.CertificateTags = flattenCertificateTags(ctx, summary.CertificateTags)

	return diags
}

type acmeEndpointListModel struct {
	framework.WithRegionModel
}

func listACMEEndpoints(ctx context.Context, conn *acm.Client, input *acm.ListAcmeEndpointsInput) iter.Seq2[awstypes.AcmeEndpointSummary, error] {
	return func(yield func(awstypes.AcmeEndpointSummary, error) bool) {
		pages := acm.NewListAcmeEndpointsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.AcmeEndpointSummary{}, fmt.Errorf("listing ACM ACME Endpoint resources: %w", err))
				return
			}

			for _, item := range page.AcmeEndpoints {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
