// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfbedrockagent "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagent"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentAgentCollaborator_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var agentcollaborator types.AgentCollaborator
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent_collaborator.test"
	model := "anthropic.claude-3-5-sonnet-20241022-v2:0"
	description := "basic claude"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentCollaboratorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentCollaboratorConfig_basic(rName, model, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentCollaboratorExists(ctx, resourceName, &agentcollaborator),
					resource.TestCheckResourceAttr(resourceName, "collaborator_name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBedrockAgentAgentCollaborator_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var agentcollaborator types.AgentCollaborator
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent_collaborator.test"
	model := "anthropic.claude-3-5-sonnet-20241022-v2:0"
	description := "basic claude"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentCollaboratorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentCollaboratorConfig_basic(rName, model, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentCollaboratorExists(ctx, resourceName, &agentcollaborator),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbedrockagent.ResourceAgentCollaborator, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAgentCollaboratorDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagent_agent_collaborator" {
				continue
			}

			_, err := tfbedrockagent.FindAgentCollaboratorByThreePartKey(ctx, conn, rs.Primary.Attributes["agent_id"], rs.Primary.Attributes["agent_version"], rs.Primary.Attributes["collaborator_id"])
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.BedrockAgent, create.ErrActionCheckingDestroyed, tfbedrockagent.ResNameAgentCollaborator, rs.Primary.ID, err)
			}

			return create.Error(names.BedrockAgent, create.ErrActionCheckingDestroyed, tfbedrockagent.ResNameAgentCollaborator, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckAgentCollaboratorExists(ctx context.Context, name string, agentcollaborator *types.AgentCollaborator) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.BedrockAgent, create.ErrActionCheckingExistence, tfbedrockagent.ResNameAgentCollaborator, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.BedrockAgent, create.ErrActionCheckingExistence, tfbedrockagent.ResNameAgentCollaborator, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		resp, err := tfbedrockagent.FindAgentCollaboratorByThreePartKey(ctx, conn, rs.Primary.Attributes["agent_id"], rs.Primary.Attributes["agent_version"], rs.Primary.Attributes["collaborator_id"])
		if err != nil {
			return create.Error(names.BedrockAgent, create.ErrActionCheckingExistence, tfbedrockagent.ResNameAgentCollaborator, rs.Primary.ID, err)
		}

		*agentcollaborator = *resp

		return nil
	}
}

func testAccAgentCollaboratorConfig_basic(rName, model, description string) string {
	return acctest.ConfigCompose(testAccAgentConfig_basic(rName, model, description),
		testAccAgentAliasConfig_alias(rName),
		fmt.Sprintf(`
resource "aws_bedrockagent_agent" "test2" {
  agent_collaboration         = "SUPERVISOR"
  agent_name                  = "%[1]s-super"
  agent_resource_role_arn     = aws_iam_role.test_agent.arn
  description                 = %[3]q
  idle_session_ttl_in_seconds = 500
  instruction                 = file("${path.module}/test-fixtures/instruction.txt")
  foundation_model            = %[2]q
  prepare_agent               = false

  depends_on = [
    aws_bedrockagent_agent_alias.test
  ]
}

resource "aws_bedrockagent_agent_collaborator" "test" {
  agent_id                   = aws_bedrockagent_agent.test2.agent_id
  collaboration_instruction  = "tell the other agent what to do"
  collaborator_name          = %[1]q
  relay_conversation_history = "TO_COLLABORATOR"

  agent_descriptor {
    alias_arn = aws_bedrockagent_agent_alias.test.agent_alias_arn
  }
}
`, rName, model, description))
}
