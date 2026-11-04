// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager

import (
	"context"
	"iter"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networksecuritymanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networksecuritymanager/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_networksecuritymanager_policy")
func newPolicyResourceAsListResource() list.ListResourceWithConfigure {
	return &policyListResource{}
}

var _ list.ListResource = &policyListResource{}

type policyListResource struct {
	policyResource
	framework.WithList
}

func (l *policyListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().NetworkSecurityManagerClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		var input networksecuritymanager.ListPoliciesInput
		// A published policy with a pending draft may be listed under both its base and DRAFT-qualified ARNs.
		seen := make(map[string]struct{})
		for item, err := range listPolicies(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}

			arn := policyBaseARN(aws.ToString(item.PolicyArn))
			if _, ok := seen[arn]; ok {
				continue
			}
			seen[arn] = struct{}{}
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			var policy *networksecuritymanager.GetPolicyOutput
			if request.IncludeResource {
				var err error
				policy, err = findPolicyByARN(ctx, conn, arn)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(smerr.NewListResultError(ctx, err, smerr.ID, arn))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data policyResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ARN = fwflex.StringValueToFramework(ctx, arn)

				if request.IncludeResource {
					smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, policy, &data), smerr.ID, arn)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = aws.ToString(item.PolicyName)
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

func listPolicies(ctx context.Context, conn *networksecuritymanager.Client, input *networksecuritymanager.ListPoliciesInput) iter.Seq2[awstypes.PolicySummary, error] {
	return tfiter.ConcatValuesWithError(listPolicyPages(ctx, conn, input))
}

func listPolicyPages(ctx context.Context, conn *networksecuritymanager.Client, input *networksecuritymanager.ListPoliciesInput) iter.Seq2[[]awstypes.PolicySummary, error] {
	return func(yield func([]awstypes.PolicySummary, error) bool) {
		pages := networksecuritymanager.NewListPoliciesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(nil, smarterr.NewError(err))
				return
			}

			if !yield(page.Policies, nil) {
				return
			}
		}
	}
}
