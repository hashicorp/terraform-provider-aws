// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentAgentPrepareAction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgent),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: testAccCheckAgentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentPrepareActionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentPrepareActionExecuted(ctx, "aws_bedrockagent_agent.test"),
				),
			},
		},
	})
}

func testAccCheckAgentPrepareActionExecuted(ctx context.Context, agentResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[agentResourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", agentResourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)
		agentID := rs.Primary.ID

		input := bedrockagent.GetAgentInput{
			AgentId: &agentID,
		}

		// Check that the agent exists and has been prepared
		output, err := conn.GetAgent(ctx, &input)
		if err != nil {
			return fmt.Errorf("error getting Bedrock Agent (%s): %w", agentID, err)
		}

		// Verify the agent status indicates it has been prepared
		if output.Agent.AgentStatus != awstypes.AgentStatusPrepared {
			return fmt.Errorf("expected agent status to be %s, got %s", awstypes.AgentStatusPrepared, output.Agent.AgentStatus)
		}

		return nil
	}
}

func testAccAgentPrepareActionConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccAgentConfig_basicUnprepared(rName, "anthropic.claude-v2", "basic claude"),
		`
action "aws_bedrockagent_agent_prepare" "test" {
  config {
    agent_id = aws_bedrockagent_agent.test.agent_id
  }
}

resource "terraform_data" "trigger" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.aws_bedrockagent_agent_prepare.test]
    }
  }

  depends_on = [aws_bedrockagent_agent.test]
}
`)
}
