// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
)

// @FrameworkListResource("aws_cloudwatch_log_account_policy")
func newAccountPolicyResourceAsListResource() list.ListResourceWithConfigure {
	return &accountPolicyListResource{}
}

var _ list.ListResource = &accountPolicyListResource{}

type accountPolicyListResource struct {
	accountPolicyResource
	framework.WithList
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

			var data accountPolicyResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				result.Diagnostics.Append(l.flatten(ctx, &item, &data)...)
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = name
			})

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
