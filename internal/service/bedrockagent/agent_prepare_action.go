// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/actionwait"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
)

const prepareTimeout = 5 * time.Minute

// @Action(aws_bedrockagent_agent_prepare, name="Agent Prepare")
func newAgentPrepareAction(_ context.Context) (action.ActionWithConfigure, error) {
	return &agentPrepareAction{}, nil
}

var (
	_ action.Action = (*agentPrepareAction)(nil)
)

type agentPrepareAction struct {
	framework.ActionWithModel[agentPrepareActionModel]
}

type agentPrepareActionModel struct {
	framework.WithRegionModel
	AgentID types.String `tfsdk:"agent_id"`
}

func (a *agentPrepareAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Prepares an Amazon Bedrock Agent for use. This action creates a DRAFT version of the agent that contains the latest changes.",
		Attributes: map[string]schema.Attribute{
			"agent_id": schema.StringAttribute{
				Description: "The unique identifier of the agent to prepare.",
				Required:    true,
			},
		},
	}
}

func (a *agentPrepareAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config agentPrepareActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := a.Meta().BedrockAgentClient(ctx)
	agentID := config.AgentID.ValueString()

	tflog.Info(ctx, "Starting Bedrock Agent prepare action", map[string]any{
		"agent_id": agentID,
	})

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Preparing Bedrock Agent %s...", agentID),
	})

	input := bedrockagent.PrepareAgentInput{
		AgentId: aws.String(agentID),
	}

	_, err := retryOpIfPreparing(ctx, prepareTimeout,
		func(ctx context.Context) (*bedrockagent.PrepareAgentOutput, error) {
			return conn.PrepareAgent(ctx, &input)
		},
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Prepare Bedrock Agent",
			fmt.Sprintf("Could not prepare Bedrock Agent %s: %s", agentID, err),
		)
		return
	}

	// Wait for agent to be prepared using actionwait
	result, err := actionwait.WaitForStatus(ctx, func(ctx context.Context) (actionwait.FetchResult[*awstypes.Agent], error) {
		agent, err := findAgentByID(ctx, conn, agentID)
		if err != nil {
			return actionwait.FetchResult[*awstypes.Agent]{}, err
		}
		return actionwait.FetchResult[*awstypes.Agent]{
			Status: actionwait.Status(agent.AgentStatus),
			Value:  agent,
		}, nil
	}, actionwait.Options[*awstypes.Agent]{
		Timeout:            prepareTimeout,
		SuccessStates:      []actionwait.Status{actionwait.Status(awstypes.AgentStatusPrepared)},
		TransitionalStates: []actionwait.Status{actionwait.Status(awstypes.AgentStatusNotPrepared), actionwait.Status(awstypes.AgentStatusPreparing)},
		ProgressInterval:   30 * time.Second,
		ProgressSink: func(fr actionwait.FetchResult[any], meta actionwait.ProgressMeta) {
			resp.SendProgress(action.InvokeProgressEvent{
				Message: fmt.Sprintf("Waiting for Bedrock Agent %s to be prepared (Status: %s, Elapsed: %v)",
					agentID, fr.Status, meta.Elapsed.Round(time.Second)),
			})
		},
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Wait for Bedrock Agent Preparation",
			fmt.Sprintf("Bedrock Agent %s failed to reach prepared state: %s", agentID, err),
		)
		return
	}

	agentStatus := string(result.Value.AgentStatus)
	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Bedrock Agent %s prepared successfully (Status: %s)", agentID, agentStatus),
	})

	tflog.Info(ctx, "Bedrock Agent prepare action completed successfully", map[string]any{
		"agent_id":     agentID,
		"agent_status": agentStatus,
	})
}
