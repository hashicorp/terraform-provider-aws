// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_bedrock_guardrail_resource_policy")
func newGuardrailResourcePolicyResourceAsListResource() list.ListResourceWithConfigure {
	return &guardrailResourcePolicyListResource{}
}

var _ list.ListResource = &guardrailResourcePolicyListResource{}

type guardrailResourcePolicyListResource struct {
	guardrailResourcePolicyResource
	framework.WithList
}

// List enumerates guardrails and yields a result for each one that has a resource policy attached.
// Bedrock has no API to list resource policies directly, and system-defined guardrail profiles cannot be listed.
func (l *guardrailResourcePolicyListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().BedrockClient(ctx)

	tflog.Info(ctx, "Listing Bedrock Guardrail Resource Policies")

	stream.Results = func(yield func(list.ListResult) bool) {
		var input bedrock.ListGuardrailsInput
		for item, err := range listGuardrails(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}

			resourceARN := aws.ToString(item.Arn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrResourceARN), resourceARN)

			// A policy lookup is needed even without include_resource to know whether this guardrail has a policy.
			policy, err := findGuardrailResourcePolicyByARN(ctx, conn, resourceARN)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				yield(smerr.NewListResultError(ctx, err, smerr.ID, resourceARN))
				return
			}

			result := request.NewListResult(ctx)

			var data guardrailResourcePolicyResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ResourceARN = fwtypes.ARNValue(resourceARN)

				// Normalize the policy the same way Read does when there is no prior state.
				policyToSet, err := verify.PolicyToSet("", aws.ToString(policy))
				if err != nil {
					smerr.AddError(ctx, &result.Diagnostics, err, smerr.ID, resourceARN)
					return
				}
				data.Policy = fwtypes.IAMPolicyValue(policyToSet)

				result.DisplayName = aws.ToString(item.Name)
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

func listGuardrails(ctx context.Context, conn *bedrock.Client, input *bedrock.ListGuardrailsInput) iter.Seq2[awstypes.GuardrailSummary, error] {
	return func(yield func(awstypes.GuardrailSummary, error) bool) {
		pages := bedrock.NewListGuardrailsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.GuardrailSummary{}, fmt.Errorf("listing Bedrock Guardrails: %w", err))
				return
			}

			for _, item := range page.Guardrails {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
