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
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_bedrockagentcore_resource_policy")
func newResourcePolicyResourceAsListResource() list.ListResourceWithConfigure {
	return &resourcePolicyListResource{}
}

var _ list.ListResource = &resourcePolicyListResource{}

type resourcePolicyListResource struct {
	resourcePolicyResource
	framework.WithList
}

func (l *resourcePolicyListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().BedrockAgentCoreClient(ctx)

	var query listResourcePolicyModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		// Check agent runtimes and their endpoints for resource policies.
		var runtimesInput bedrockagentcorecontrol.ListAgentRuntimesInput
		for runtime, err := range listAgentRuntimesForPolicies(ctx, conn, &runtimesInput) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			if !l.yieldResourcePolicy(ctx, conn, request, aws.ToString(runtime.AgentRuntimeArn), yield) {
				return
			}

			endpointsInput := bedrockagentcorecontrol.ListAgentRuntimeEndpointsInput{
				AgentRuntimeId: runtime.AgentRuntimeId,
			}
			for endpoint, err := range listAgentRuntimeEndpointsForPolicies(ctx, conn, &endpointsInput) {
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}

				if !l.yieldResourcePolicy(ctx, conn, request, aws.ToString(endpoint.AgentRuntimeEndpointArn), yield) {
					return
				}
			}
		}

		// Check gateways for resource policies.
		var gatewaysInput bedrockagentcorecontrol.ListGatewaysInput
		for gateway, err := range listGatewaysForPolicies(ctx, conn, &gatewaysInput) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			gatewayOutput, err := findGatewayByID(ctx, conn, aws.ToString(gateway.GatewayId))
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			if !l.yieldResourcePolicy(ctx, conn, request, aws.ToString(gatewayOutput.GatewayArn), yield) {
				return
			}
		}
	}
}

// yieldResourcePolicy checks if the given ARN has a resource policy attached and,
// if so, yields a list result. Returns false if the caller should stop iteration.
func (l *resourcePolicyListResource) yieldResourcePolicy(ctx context.Context, conn *bedrockagentcorecontrol.Client, request list.ListRequest, resourceARN string, yield func(list.ListResult) bool) bool {
	ctx = tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrResourceARN), resourceARN)

	policy, err := findResourcePolicyByARN(ctx, conn, resourceARN)
	if retry.NotFound(err) {
		return true
	}
	if err != nil {
		yield(fwdiag.NewListResultErrorDiagnostic(err))
		return false
	}

	result := request.NewListResult(ctx)
	result.DisplayName = resourceARN

	var data resourcePolicyResourceModel
	data.ResourceARN = fwtypes.ARNValue(resourceARN)

	l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
		data.Policy = fwtypes.IAMPolicyValue(aws.ToString(policy))
	})

	return yield(result)
}

type listResourcePolicyModel struct {
	framework.WithRegionModel
}

func listAgentRuntimesForPolicies(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.ListAgentRuntimesInput) iter.Seq2[awstypes.AgentRuntime, error] {
	return func(yield func(awstypes.AgentRuntime, error) bool) {
		pages := bedrockagentcorecontrol.NewListAgentRuntimesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[awstypes.AgentRuntime](), fmt.Errorf("listing Bedrock AgentCore Agent Runtimes for resource policies: %w", err))
				return
			}

			for _, item := range page.AgentRuntimes {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}

func listAgentRuntimeEndpointsForPolicies(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.ListAgentRuntimeEndpointsInput) iter.Seq2[awstypes.AgentRuntimeEndpoint, error] {
	return func(yield func(awstypes.AgentRuntimeEndpoint, error) bool) {
		pages := bedrockagentcorecontrol.NewListAgentRuntimeEndpointsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[awstypes.AgentRuntimeEndpoint](), fmt.Errorf("listing Bedrock AgentCore Agent Runtime Endpoints for resource policies: %w", err))
				return
			}

			for _, item := range page.RuntimeEndpoints {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}

func listGatewaysForPolicies(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.ListGatewaysInput) iter.Seq2[awstypes.GatewaySummary, error] {
	return func(yield func(awstypes.GatewaySummary, error) bool) {
		pages := bedrockagentcorecontrol.NewListGatewaysPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[awstypes.GatewaySummary](), fmt.Errorf("listing Bedrock AgentCore Gateways for resource policies: %w", err))
				return
			}

			for _, item := range page.Items {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
