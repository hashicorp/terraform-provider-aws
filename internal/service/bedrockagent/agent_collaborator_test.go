// Copyright IBM Corp. 2014, 2026
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
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagent "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagent"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentAgentCollaborator_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var agentcollaborator awstypes.AgentCollaborator
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent_collaborator.test"
	aliasResourceName := "aws_bedrockagent_agent_alias.test"
	superResourceName := "aws_bedrockagent_agent.test_super"

	instruction := "tell the other agent what to do"
	model := "anthropic.claude-3-5-sonnet-20241022-v2:0"
	description := "basic claude"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentCollaboratorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentCollaboratorConfig_basic(rName, model, description, instruction),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentCollaboratorExists(ctx, t, resourceName, &agentcollaborator),
					resource.TestCheckResourceAttrPair(resourceName, "agent_id", superResourceName, "agent_id"),
					resource.TestCheckResourceAttr(resourceName, "collaborator_name", rName),
					resource.TestCheckResourceAttr(resourceName, "collaboration_instruction", instruction),
					resource.TestCheckResourceAttr(resourceName, "relay_conversation_history", string(awstypes.RelayConversationHistoryToCollaborator)),
					resource.TestCheckResourceAttrPair(resourceName, "agent_descriptor.0.alias_arn", aliasResourceName, "agent_alias_arn"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent_collaborator.test"
	aliasResourceName := "aws_bedrockagent_agent_alias.test"
	aliasUpdateResourceName := "aws_bedrockagent_agent_alias.test_update"

	instruction := "tell the other agent what to do"
	instruction2 := "Other instruction"
	model := "anthropic.claude-3-5-sonnet-20241022-v2:0"
	description := "basic claude"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentCollaboratorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentCollaboratorConfig_basic(rName, model, description, instruction),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentCollaboratorExists(ctx, t, resourceName, &agentcollaborator),
					resource.TestCheckResourceAttr(resourceName, "collaborator_name", rName),
					resource.TestCheckResourceAttr(resourceName, "collaboration_instruction", instruction),
					resource.TestCheckResourceAttr(resourceName, "relay_conversation_history", string(awstypes.RelayConversationHistoryToCollaborator)),
					resource.TestCheckResourceAttrPair(resourceName, "agent_descriptor.0.alias_arn", aliasResourceName, "agent_alias_arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAgentCollaboratorConfig_update(rName, model, description, instruction2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentCollaboratorExists(ctx, t, resourceName, &agentcollaborator),
					resource.TestCheckResourceAttr(resourceName, "collaborator_name", rName),
					resource.TestCheckResourceAttr(resourceName, "collaboration_instruction", instruction2),
					resource.TestCheckResourceAttr(resourceName, "relay_conversation_history", string(awstypes.RelayConversationHistoryDisabled)),
					resource.TestCheckResourceAttrPair(resourceName, "agent_descriptor.0.alias_arn", aliasUpdateResourceName, "agent_alias_arn"),
				),
			},
		},
	})
}

func TestAccBedrockAgentAgentCollaborator_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var agentcollaborator awstypes.AgentCollaborator
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent_collaborator.test"
	instruction := "tell the other agent what to do"
	model := "anthropic.claude-3-5-sonnet-20241022-v2:0"
	description := "basic claude"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentCollaboratorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentCollaboratorConfig_basic(rName, model, description, instruction),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentCollaboratorExists(ctx, t, resourceName, &agentcollaborator),
					acctest.CheckFrameworkResourceDisappearsWithStateFunc(ctx, t, tfbedrockagent.ResourceAgentCollaborator, resourceName, agentCollaboratorDisappearsStateFunc),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Verifies retry behavior when multiple collaborators attempt to associate to the same
// supervisor concurrently
//
// Ref: https://github.com/hashicorp/terraform-provider-aws/issues/42256
func TestAccBedrockAgentAgentCollaborator_multiCollaborator(t *testing.T) {
	ctx := acctest.Context(t)
	var c1, c2, c3 awstypes.AgentCollaborator
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent_collaborator.test"
	resourceName2 := "aws_bedrockagent_agent_collaborator.test2"
	resourceName3 := "aws_bedrockagent_agent_collaborator.test3"
	superResourceName := "aws_bedrockagent_agent.test_super"
	aliasResourceName := "aws_bedrockagent_agent_alias.test"
	aliasResourceName2 := "aws_bedrockagent_agent_alias.test2"
	aliasResourceName3 := "aws_bedrockagent_agent_alias.test3"

	instruction := "Do what the supervisor tells you to do. Think clearly."
	model := "anthropic.claude-3-5-sonnet-20241022-v2:0"
	description := "basic claude"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentCollaboratorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentCollaboratorConfig_multiCollaborator(rName, model, description, instruction),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentCollaboratorExists(ctx, t, resourceName, &c1),
					testAccCheckAgentCollaboratorExists(ctx, t, resourceName2, &c2),
					testAccCheckAgentCollaboratorExists(ctx, t, resourceName3, &c3),
					resource.TestCheckResourceAttrPair(resourceName, "agent_id", superResourceName, "agent_id"),
					resource.TestCheckResourceAttr(resourceName, "collaborator_name", rName+"-1"),
					resource.TestCheckResourceAttr(resourceName2, "collaborator_name", rName+"-2"),
					resource.TestCheckResourceAttr(resourceName3, "collaborator_name", rName+"-3"),
					resource.TestCheckResourceAttr(resourceName, "collaboration_instruction", instruction),
					resource.TestCheckResourceAttr(resourceName, "relay_conversation_history", string(awstypes.RelayConversationHistoryToCollaborator)),
					resource.TestCheckResourceAttrPair(resourceName, "agent_descriptor.0.alias_arn", aliasResourceName, "agent_alias_arn"),
					resource.TestCheckResourceAttrPair(resourceName2, "agent_descriptor.0.alias_arn", aliasResourceName2, "agent_alias_arn"),
					resource.TestCheckResourceAttrPair(resourceName3, "agent_descriptor.0.alias_arn", aliasResourceName3, "agent_alias_arn"),
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

func testAccCheckAgentCollaboratorDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagent_agent_collaborator" {
				continue
			}

			_, err := tfbedrockagent.FindAgentCollaboratorByThreePartKey(ctx, conn, rs.Primary.Attributes["agent_id"], rs.Primary.Attributes["agent_version"], rs.Primary.Attributes["collaborator_id"])

			if retry.NotFound(err) {
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

func testAccCheckAgentCollaboratorExists(ctx context.Context, t *testing.T, n string, v *awstypes.AgentCollaborator) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentClient(ctx)

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
	return acctest.ConfigCompose(
		testAccAgentConfig_basic(rName, model, description),
		testAccAgentAliasConfig_alias(rName),
		fmt.Sprintf(`
resource "aws_bedrockagent_agent" "test_super" {
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
  agent_id                   = aws_bedrockagent_agent.test_super.agent_id
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
	return acctest.ConfigCompose(
		testAccAgentConfig_basic(rName, model, description),
		testAccAgentAliasConfig_alias(rName),
		fmt.Sprintf(`
resource "aws_bedrockagent_agent" "test_super" {
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
  agent_id                   = aws_bedrockagent_agent.test_super.agent_id
  collaboration_instruction  = %[4]q
  collaborator_name          = %[1]q
  relay_conversation_history = "DISABLED"

  agent_descriptor {
    alias_arn = aws_bedrockagent_agent_alias.test_update.agent_alias_arn
  }
}

resource "aws_bedrockagent_agent_alias" "test_update" {
  agent_alias_name = "%[1]s-update"
  agent_id         = aws_bedrockagent_agent.test.agent_id
  description      = "Test Alias Update"
}
`, rName, model, description, instruction))
}

func testAccAgentCollaboratorConfig_multiCollaborator(rName, model, description, instruction string) string {
	return acctest.ConfigCompose(
		testAccAgent_base(rName, model),
		fmt.Sprintf(`
resource "aws_bedrockagent_agent" "test_super" {
  agent_collaboration         = "SUPERVISOR"
  agent_name                  = "%[1]s-super"
  agent_resource_role_arn     = aws_iam_role.test_agent.arn
  description                 = %[3]q
  idle_session_ttl_in_seconds = 500
  instruction                 = file("${path.module}/test-fixtures/instruction.txt")
  foundation_model            = %[2]q
  prepare_agent               = false
}

resource "aws_bedrockagent_agent" "test" {
  agent_name                  = "%[1]s-1"
  agent_resource_role_arn     = aws_iam_role.test_agent.arn
  description                 = %[3]q
  idle_session_ttl_in_seconds = 500
  instruction                 = file("${path.module}/test-fixtures/instruction.txt")
  foundation_model            = %[2]q
}

resource "aws_bedrockagent_agent_alias" "test" {
  agent_alias_name = %[1]q
  agent_id         = aws_bedrockagent_agent.test.agent_id
  description      = "Test Alias"
}

resource "aws_bedrockagent_agent_collaborator" "test" {
  agent_id                   = aws_bedrockagent_agent.test_super.agent_id
  collaboration_instruction  = %[4]q
  collaborator_name          = "%[1]s-1"
  relay_conversation_history = "TO_COLLABORATOR"

  agent_descriptor {
    alias_arn = aws_bedrockagent_agent_alias.test.agent_alias_arn
  }
}

resource "aws_bedrockagent_agent" "test2" {
  agent_name                  = "%[1]s-2"
  agent_resource_role_arn     = aws_iam_role.test_agent.arn
  description                 = %[3]q
  idle_session_ttl_in_seconds = 500
  instruction                 = file("${path.module}/test-fixtures/instruction.txt")
  foundation_model            = %[2]q
}

resource "aws_bedrockagent_agent_alias" "test2" {
  agent_alias_name = "%[1]s-2"
  agent_id         = aws_bedrockagent_agent.test2.agent_id
  description      = "Test Alias 2"
}

resource "aws_bedrockagent_agent_collaborator" "test2" {
  agent_id                   = aws_bedrockagent_agent.test_super.agent_id
  collaboration_instruction  = %[4]q
  collaborator_name          = "%[1]s-2"
  relay_conversation_history = "TO_COLLABORATOR"

  agent_descriptor {
    alias_arn = aws_bedrockagent_agent_alias.test2.agent_alias_arn
  }
}

resource "aws_bedrockagent_agent" "test3" {
  agent_name                  = "%[1]s-3"
  agent_resource_role_arn     = aws_iam_role.test_agent.arn
  description                 = %[3]q
  idle_session_ttl_in_seconds = 500
  instruction                 = file("${path.module}/test-fixtures/instruction.txt")
  foundation_model            = %[2]q
}

resource "aws_bedrockagent_agent_alias" "test3" {
  agent_alias_name = "%[1]s-3"
  agent_id         = aws_bedrockagent_agent.test3.agent_id
  description      = "Test Alias 3"
}

resource "aws_bedrockagent_agent_collaborator" "test3" {
  agent_id                   = aws_bedrockagent_agent.test_super.agent_id
  collaboration_instruction  = %[4]q
  collaborator_name          = "%[1]s-3"
  relay_conversation_history = "TO_COLLABORATOR"

  agent_descriptor {
    alias_arn = aws_bedrockagent_agent_alias.test3.agent_alias_arn
  }
}
`, rName, model, description, instruction))
}
