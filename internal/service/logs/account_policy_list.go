// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @SDKListResource("aws_cloudwatch_log_account_policy")
func newAccountPolicyResourceAsListResource() inttypes.ListResourceForSDK {
	l := accountPolicyListResource{}
	l.SetResourceSchema(resourceAccountPolicy())
	return &l
}

var _ list.ListResource = &accountPolicyListResource{}

type accountPolicyListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *accountPolicyListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"policy_type": listschema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.PolicyType](),
				Required:    true,
				Description: "Type of account policy to list.",
			},
		},
	}
}

func (l *accountPolicyListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().LogsClient(ctx)

	var query listAccountPolicyModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	policyType := query.PolicyType.ValueEnum()

	tflog.Info(ctx, "Listing Resources", map[string]any{
		logging.ResourceAttributeKey("policy_type"): policyType,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := cloudwatchlogs.DescribeAccountPoliciesInput{
			PolicyType: policyType,
		}
		for item, err := range listAccountPolicies(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			name := aws.ToString(item.PolicyName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("policy_name"), name)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(name)
			rd.Set("policy_name", name)
			rd.Set("policy_type", item.PolicyType)

			if request.IncludeResource {
				if err := resourceAccountPolicyFlatten(rd, &item); err != nil {
					tflog.Error(ctx, "Reading CloudWatch Logs Account Policy", map[string]any{
						"error": err.Error(),
					})
					continue
				}
			}

			result.DisplayName = name

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

type listAccountPolicyModel struct {
	framework.WithRegionModel
	PolicyType fwtypes.StringEnum[awstypes.PolicyType] `tfsdk:"policy_type"`
}

func listAccountPolicies(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeAccountPoliciesInput) iter.Seq2[awstypes.AccountPolicy, error] {
	return tfiter.ConcatValuesWithError(listAccountPolicyPages(ctx, conn, input))
}

// listAccountPolicyPages pages DescribeAccountPolicies, which has no SDK paginator.
func listAccountPolicyPages(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeAccountPoliciesInput) iter.Seq2[[]awstypes.AccountPolicy, error] {
	return func(yield func([]awstypes.AccountPolicy, error) bool) {
		for {
			page, err := conn.DescribeAccountPolicies(ctx, input)
			if err != nil {
				yield(nil, fmt.Errorf("listing CloudWatch Logs Account Policies: %w", err))
				return
			}

			if !yield(page.AccountPolicies, nil) {
				return
			}

			if aws.ToString(page.NextToken) == "" {
				return
			}
			input.NextToken = page.NextToken
		}
	}
}
