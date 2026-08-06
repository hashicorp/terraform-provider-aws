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
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_bedrockagentcore_gateway_rule")
func newGatewayRuleResourceAsListResource() list.ListResourceWithConfigure {
	return &gatewayRuleListResource{}
}

var _ list.ListResource = &gatewayRuleListResource{}

type gatewayRuleListResource struct {
	gatewayRuleResource
	framework.WithList
}

func (l *gatewayRuleListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"gateway_identifier": listschema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (l *gatewayRuleListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().BedrockAgentCoreClient(ctx)

	var query listGatewayRuleModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	gatewayIdentifier := fwflex.StringValueFromFramework(ctx, query.GatewayIdentifier)

	tflog.Info(ctx, "Listing Resources", map[string]any{
		logging.ResourceAttributeKey("gateway_identifier"): gatewayIdentifier,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := bedrockagentcorecontrol.ListGatewayRulesInput{
			GatewayIdentifier: aws.String(gatewayIdentifier),
		}
		for item, err := range listGatewayRules(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			id := aws.ToString(item.RuleId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)

			var data gatewayRuleResourceModel
			// awstypes.GatewayRuleDetail has no GatewayIdentifier field.
			data.GatewayIdentifier = fwflex.StringValueToFramework(ctx, gatewayIdentifier)

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				result.Diagnostics.Append(l.flatten(ctx, &item, &data)...)
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = id
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listGatewayRuleModel struct {
	framework.WithRegionModel
	GatewayIdentifier types.String `tfsdk:"gateway_identifier"`
}

func listGatewayRules(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.ListGatewayRulesInput) iter.Seq2[awstypes.GatewayRuleDetail, error] {
	return func(yield func(awstypes.GatewayRuleDetail, error) bool) {
		pages := bedrockagentcorecontrol.NewListGatewayRulesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[awstypes.GatewayRuleDetail](), fmt.Errorf("listing Bedrock AgentCore Gateway Rules: %w", err))
				return
			}

			for _, item := range page.GatewayRules {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
