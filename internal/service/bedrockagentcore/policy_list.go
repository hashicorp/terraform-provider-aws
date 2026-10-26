// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_bedrockagentcore_policy")
func newPolicyResourceAsListResource() list.ListResourceWithConfigure {
	return &policyListResource{}
}

var _ list.ListResource = &policyListResource{}

type policyListResource struct {
	policyResource
	framework.WithList
}

func (l *policyListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"policy_engine_id": listschema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (l *policyListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().BedrockAgentCoreClient(ctx)

	var query listPolicyModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	policyEngineID := fwflex.StringValueFromFramework(ctx, query.PolicyEngineID)

	tflog.Info(ctx, "Listing Resources", map[string]any{
		logging.ResourceAttributeKey("policy_engine_id"): policyEngineID,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := bedrockagentcorecontrol.ListPoliciesInput{
			PolicyEngineId: aws.String(policyEngineID),
		}
		for item, err := range listPolicies(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			arn := aws.ToString(item.PolicyArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			result := request.NewListResult(ctx)

			var data policyResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, &item, &data))
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
	PolicyEngineID types.String `tfsdk:"policy_engine_id"`
}

func listPolicies(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.ListPoliciesInput) iter.Seq2[awstypes.Policy, error] {
	return func(yield func(awstypes.Policy, error) bool) {
		pages := bedrockagentcorecontrol.NewListPoliciesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[awstypes.Policy](), fmt.Errorf("listing Bedrock AgentCore Policies: %w", err))
				return
			}

			for _, item := range page.Policies {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
