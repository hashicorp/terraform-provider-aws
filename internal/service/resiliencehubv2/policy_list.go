// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehubv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_resiliencehubv2_policy")
func newResourcePolicyAsListResource() list.ListResourceWithConfigure {
	return &policyListResource{}
}

var _ list.ListResource = &policyListResource{}

type policyListResource struct {
	resourcePolicy
	framework.WithList
}

func (l *policyListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().ResilienceHubV2Client(ctx)

	var query listPolicyModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		var input resiliencehubv2.ListPoliciesInput
		for item, err := range listPolicies(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			arn := aws.ToString(item.PolicyArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			output, err := findPolicyByARN(ctx, conn, arn)
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			result := request.NewListResult(ctx)

			var data resourcePolicyModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, output, &data))
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

type listPolicyModel struct {
	framework.WithRegionModel
}

func listPolicies(ctx context.Context, conn *resiliencehubv2.Client, input *resiliencehubv2.ListPoliciesInput) iter.Seq2[awstypes.PolicySummary, error] {
	return func(yield func(awstypes.PolicySummary, error) bool) {
		pages := resiliencehubv2.NewListPoliciesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.PolicySummary{}, fmt.Errorf("listing Resilience Hub V2 Policy resources: %w", err))
				return
			}

			for _, item := range page.PolicySummaries {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
