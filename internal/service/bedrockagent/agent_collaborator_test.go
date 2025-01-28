// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	tfbedrockagent "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagent"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentAgentCollaborator_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var agentcollaborator awstypes.AgentCollaborator
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent_collaborator.test"
	resourceAlias := "aws_bedrockagent_agent_alias.test"
	resourceSuper := "aws_bedrockagent_agent.test2"
	instruction := "tell the other agent what to do"
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
				Config: testAccAgentCollaboratorConfig_basic(rName, model, description, instruction),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentCollaboratorExists(ctx, resourceName, &agentcollaborator),
					resource.TestCheckResourceAttrPair(resourceName, "agent_id", resourceSuper, "agent_id"),
					resource.TestCheckResourceAttr(resourceName, "collaborator_name", rName),
					resource.TestCheckResourceAttr(resourceName, "collaboration_instruction", instruction),
					resource.TestCheckResourceAttr(resourceName, "relay_conversation_history", string(awstypes.RelayConversationHistoryToCollaborator)),
					resource.TestCheckResourceAttrPair(resourceName, "agent_descriptor.0.alias_arn", resourceAlias, "agent_alias_arn"),
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

func TestAccBedrockAgentAgentCollaborator_update(t *testing.T) {
	ctx := acctest.Context(t)
	var agentcollaborator awstypes.AgentCollaborator
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent_collaborator.test"
	resourceAlias := "aws_bedrockagent_agent_alias.test"
	resourceAlias2 := "aws_bedrockagent_agent_alias.test2"
	instruction := "tell the other agent what to do"
	instruction2 := "Other instruction"
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
				Config: testAccAgentCollaboratorConfig_basic(rName, model, description, instruction),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentCollaboratorExists(ctx, resourceName, &agentcollaborator),
					resource.TestCheckResourceAttr(resourceName, "collaborator_name", rName),
					resource.TestCheckResourceAttr(resourceName, "collaboration_instruction", instruction),
					resource.TestCheckResourceAttr(resourceName, "relay_conversation_history", string(awstypes.RelayConversationHistoryToCollaborator)),
					resource.TestCheckResourceAttrPair(resourceName, "agent_descriptor.0.alias_arn", resourceAlias, "agent_alias_arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAgentCollaboratorConfig_update(rName, model, description, instruction2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentCollaboratorExists(ctx, resourceName, &agentcollaborator),
					resource.TestCheckResourceAttr(resourceName, "collaborator_name", rName),
					resource.TestCheckResourceAttr(resourceName, "collaboration_instruction", instruction2),
					resource.TestCheckResourceAttr(resourceName, "relay_conversation_history", string(awstypes.RelayConversationHistoryDisabled)),
					resource.TestCheckResourceAttrPair(resourceName, "agent_descriptor.0.alias_arn", resourceAlias2, "agent_alias_arn"),
				),
			},
		},
	})
}

func TestAccBedrockAgentAgentCollaborator_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var agentcollaborator awstypes.AgentCollaborator
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent_collaborator.test"
	instruction := "tell the other agent what to do"
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
				Config: testAccAgentCollaboratorConfig_basic(rName, model, description, instruction),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentCollaboratorExists(ctx, resourceName, &agentcollaborator),
					acctest.CheckFrameworkResourceDisappearsWithStateFunc(ctx, acctest.Provider, tfbedrockagent.ResourceAgentCollaborator, resourceName, agentCollaboratorDisappearsStateFunc),
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
				return err
			}

			return fmt.Errorf("Bedrock Agent Collaborator %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAgentCollaboratorExists(ctx context.Context, n string, v *awstypes.AgentCollaborator) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		output, err := tfbedrockagent.FindAgentCollaboratorByThreePartKey(ctx, conn, rs.Primary.Attributes["agent_id"], rs.Primary.Attributes["agent_version"], rs.Primary.Attributes["collaborator_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func agentCollaboratorDisappearsStateFunc(ctx context.Context, state *tfsdk.State, is *terraform.InstanceState) error {
	v, ok := is.Attributes["agent_id"]
	if !ok {
		return errors.New(`Identifying attribute "agent_id" not defined`)
	}

	if err := fwdiag.DiagnosticsError(state.SetAttribute(ctx, path.Root("agent_id"), v)); err != nil {
		return err
	}

	v, ok = is.Attributes["agent_version"]
	if !ok {
		return errors.New(`Identifying attribute "agent_version" not defined`)
	}

	if err := fwdiag.DiagnosticsError(state.SetAttribute(ctx, path.Root("agent_version"), v)); err != nil {
		return err
	}

	v, ok = is.Attributes["collaborator_id"]
	if !ok {
		return errors.New(`Identifying attribute "collaborator_id" not defined`)
	}

	if err := fwdiag.DiagnosticsError(state.SetAttribute(ctx, path.Root("collaborator_id"), v)); err != nil {
		return err
	}

	if err := fwdiag.DiagnosticsError(state.SetAttribute(ctx, path.Root("prepare_agent"), true)); err != nil {
		return err
	}

	return nil
}

func testAccAgentCollaboratorConfig_basic(rName, model, description, instruction string) string {
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
}

resource "aws_bedrockagent_agent_collaborator" "test" {
  agent_id                   = aws_bedrockagent_agent.test2.agent_id
  collaboration_instruction  = %[4]q
  collaborator_name          = %[1]q
  relay_conversation_history = "TO_COLLABORATOR"

  agent_descriptor {
    alias_arn = aws_bedrockagent_agent_alias.test.agent_alias_arn
  }
}
`, rName, model, description, instruction))
}

func testAccAgentCollaboratorConfig_update(rName, model, description, instruction string) string {
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
}

resource "aws_bedrockagent_agent_collaborator" "test" {
  agent_id                   = aws_bedrockagent_agent.test2.agent_id
  collaboration_instruction  = %[4]q
  collaborator_name          = %[1]q
  relay_conversation_history = "DISABLED"

  agent_descriptor {
    alias_arn = aws_bedrockagent_agent_alias.test2.agent_alias_arn
  }
}

resource "aws_bedrockagent_agent_alias" "test2" {
  agent_alias_name = "%[1]s-update"
  agent_id         = aws_bedrockagent_agent.test.agent_id
  description      = "Test Alias Update"
}

`, rName, model, description, instruction))
}
