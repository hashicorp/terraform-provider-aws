// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfbedrockagent "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagent"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentAgent_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent.test"
	var v awstypes.Agent

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentConfig_basic(rName, "anthropic.claude-v2", "basic claude"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_name", rName),
					resource.TestCheckResourceAttr(resourceName, "guardrail_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "prompt_override_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "basic claude"),
					resource.TestCheckResourceAttr(resourceName, "prepare_agent", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_resource_in_use_check"},
			},
		},
	})
}

func TestAccBedrockAgentAgent_full(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent.test"
	var v awstypes.Agent

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentConfig_full(rName, "anthropic.claude-v2", "basic claude"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_name", rName),
					resource.TestCheckResourceAttr(resourceName, "prompt_override_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "basic claude"),
					resource.TestCheckResourceAttr(resourceName, "skip_resource_in_use_check", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_resource_in_use_check"},
			},
		},
	})
}

func TestAccBedrockAgentAgent_singlePrompt(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent.test"
	var v awstypes.Agent

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentConfig_singlePrompt(rName, "anthropic.claude-v2", "basic claude"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_name", rName),
					resource.TestCheckResourceAttr(resourceName, "prompt_override_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "basic claude"),
					resource.TestCheckResourceAttr(resourceName, "skip_resource_in_use_check", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_resource_in_use_check"},
			},
		},
	})
}

func TestAccBedrockAgentAgent_singlePromptUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent.test"
	var v awstypes.Agent

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentConfig_singlePromptUpdate(rName, "anthropic.claude-v2", "basic claude", "500"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_name", rName),
					resource.TestCheckResourceAttr(resourceName, "prompt_override_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "basic claude"),
					resource.TestCheckResourceAttr(resourceName, "skip_resource_in_use_check", acctest.CtTrue),
				),
			},
			{
				Config: testAccAgentConfig_singlePromptUpdate(rName, "anthropic.claude-v2", "basic claude", "501"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_name", rName),
					resource.TestCheckResourceAttr(resourceName, "prompt_override_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "basic claude"),
					resource.TestCheckResourceAttr(resourceName, "skip_resource_in_use_check", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_resource_in_use_check"},
			},
		},
	})
}

func TestAccBedrockAgentAgent_addPrompt(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent.test"
	var v awstypes.Agent

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentConfig_basic(rName, "anthropic.claude-v2", "basic claude"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_name", rName),
					resource.TestCheckResourceAttr(resourceName, "prompt_override_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "prompt_override_configuration.0.prompt_configurations.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "basic claude"),
					resource.TestCheckResourceAttr(resourceName, "skip_resource_in_use_check", acctest.CtFalse),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_resource_in_use_check"},
			},
			{
				Config: testAccAgentConfig_singlePrompt(rName, "anthropic.claude-v2", "basic claude"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_name", rName),
					resource.TestCheckResourceAttr(resourceName, "prompt_override_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "prompt_override_configuration.0.prompt_configurations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "basic claude"),
					resource.TestCheckResourceAttr(resourceName, "skip_resource_in_use_check", acctest.CtTrue),
				),
			},
			{
				Config: testAccAgentConfig_full(rName, "anthropic.claude-v2", "basic claude"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_name", rName),
					resource.TestCheckResourceAttr(resourceName, "prompt_override_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "prompt_override_configuration.0.prompt_configurations.#", "4"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "basic claude"),
					resource.TestCheckResourceAttr(resourceName, "skip_resource_in_use_check", acctest.CtTrue),
				),
			},
			{
				Config: testAccAgentConfig_singlePrompt(rName, "anthropic.claude-v2", "basic claude"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_name", rName),
					resource.TestCheckResourceAttr(resourceName, "prompt_override_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "prompt_override_configuration.0.prompt_configurations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "basic claude"),
					resource.TestCheckResourceAttr(resourceName, "skip_resource_in_use_check", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_resource_in_use_check"},
			},
		},
	})
}

func TestAccBedrockAgentAgent_guardrail(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent.test"
	guardrailResourceName := "aws_bedrock_guardrail.test"
	var v awstypes.Agent

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentConfig_guardrail_noConfig(rName, "anthropic.claude-v2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_name", rName),
					resource.TestCheckResourceAttr(resourceName, "guardrail_configuration.#", "0"),
				),
			},
			{
				Config: testAccAgentConfig_guardrail_withConfig(rName, "anthropic.claude-v2", "DRAFT"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_name", rName),
					resource.TestCheckResourceAttr(resourceName, "guardrail_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "guardrail_configuration.0.guardrail_identifier", guardrailResourceName, "guardrail_id"),
					resource.TestCheckResourceAttr(resourceName, "guardrail_configuration.0.guardrail_version", "DRAFT"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_resource_in_use_check"},
			},
			{
				Config: testAccAgentConfig_guardrail_withConfig(rName, "anthropic.claude-v2", "1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_name", rName),
					resource.TestCheckResourceAttr(resourceName, "guardrail_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "guardrail_configuration.0.guardrail_identifier", guardrailResourceName, "guardrail_id"),
					resource.TestCheckResourceAttr(resourceName, "guardrail_configuration.0.guardrail_version", "1"),
				),
			},
			{
				Config: testAccAgentConfig_guardrail_noConfig(rName, "anthropic.claude-v2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_name", rName),
					resource.TestCheckResourceAttr(resourceName, "guardrail_configuration.#", "0"),
				),
			},
		},
	})
}

func TestAccBedrockAgentAgent_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent.test"
	var v awstypes.Agent

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentConfig_basic(rName+"-1", "anthropic.claude-v2", "basic claude"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_name", rName+"-1"),
					resource.TestCheckResourceAttr(resourceName, "prompt_override_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "basic claude"),
				),
			},
			{
				Config: testAccAgentConfig_basic(rName+"-2", "anthropic.claude-v2", "basic claude"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_name", rName+"-2"),
					resource.TestCheckResourceAttr(resourceName, "prompt_override_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "basic claude"),
				),
			},
			{
				Config: testAccAgentConfig_basic(rName+"-3", "anthropic.claude-v2", "basic claude again"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_name", rName+"-3"),
					resource.TestCheckResourceAttr(resourceName, "prompt_override_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "basic claude again"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_resource_in_use_check"},
			},
		},
	})
}

func TestAccBedrockAgentAgent_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent.test"
	var agent awstypes.Agent

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentConfig_tags1(rName, "anthropic.claude-v2", acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &agent),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_resource_in_use_check"},
			},
			{
				Config: testAccAgentConfig_tags2(rName, "anthropic.claude-v2", acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &agent),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAgentConfig_tags1(rName, "anthropic.claude-v2", acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &agent),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccBedrockAgentAgent_kms(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent.test"
	var v awstypes.Agent

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentConfig_kms(rName, "anthropic.claude-v2", "basic claude", "500"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_name", rName),
					resource.TestCheckResourceAttr(resourceName, "guardrail_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "prompt_override_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "basic claude"),
					resource.TestCheckResourceAttr(resourceName, "prepare_agent", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "customer_encryption_key_arn", "aws_kms_key.test_agent", names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_resource_in_use_check"},
			},
			{
				Config: testAccAgentConfig_kms(rName, "anthropic.claude-v2", "basic claude", "501"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_name", rName),
					resource.TestCheckResourceAttr(resourceName, "guardrail_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "prompt_override_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "basic claude"),
					resource.TestCheckResourceAttr(resourceName, "prepare_agent", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "customer_encryption_key_arn", "aws_kms_key.test_agent", names.AttrARN),
				),
			},
		},
	})
}

func TestAccBedrockAgentAgent_agentCollaboration(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent.test"
	var v awstypes.Agent

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentConfig_agentCollaboration(rName, "anthropic.claude-3-5-sonnet-20240620-v1:0", "basic claude", string(awstypes.AgentCollaborationSupervisor)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_name", rName),
					resource.TestCheckResourceAttr(resourceName, "guardrail_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "prompt_override_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "basic claude"),
					resource.TestCheckResourceAttr(resourceName, "prepare_agent", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "agent_collaboration", string(awstypes.AgentCollaborationSupervisor)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_resource_in_use_check", "prepare_agent"},
			},
			{
				Config: testAccAgentConfig_agentCollaboration(rName, "anthropic.claude-3-5-sonnet-20240620-v1:0", "basic claude", string(awstypes.AgentCollaborationSupervisorRouter)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_name", rName),
					resource.TestCheckResourceAttr(resourceName, "guardrail_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "prompt_override_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "basic claude"),
					resource.TestCheckResourceAttr(resourceName, "prepare_agent", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "agent_collaboration", string(awstypes.AgentCollaborationSupervisorRouter)),
				),
			},
		},
	})
}

func TestAccBedrockAgentAgent_memoryConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent.test"
	var v awstypes.Agent

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentConfig_memoryConfiguration(rName, "anthropic.claude-3-sonnet-20240229-v1:0", "basic claude"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "basic claude"),
					resource.TestCheckResourceAttr(resourceName, "memory_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "skip_resource_in_use_check", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_resource_in_use_check"},
			},
		},
	})
}

func testAccCheckAgentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrock_agent" {
				continue
			}

			_, err := tfbedrockagent.FindAgentByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAgentExists(ctx context.Context, n string, v *awstypes.Agent) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		output, err := tfbedrockagent.FindAgentByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAgent_base(rName, model string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test_agent" {
  assume_role_policy = data.aws_iam_policy_document.test_agent_trust.json
  name_prefix        = "AmazonBedrockExecutionRoleForAgents_tf"
}

data "aws_iam_policy_document" "test_agent_trust" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      identifiers = ["bedrock.amazonaws.com"]
      type        = "Service"
    }
    condition {
      test     = "StringEquals"
      values   = [data.aws_caller_identity.current_agent.account_id]
      variable = "aws:SourceAccount"
    }

    condition {
      test     = "ArnLike"
      values   = ["arn:${data.aws_partition.current_agent.partition}:bedrock:${data.aws_region.current_agent.name}:${data.aws_caller_identity.current_agent.account_id}:agent/*"]
      variable = "AWS:SourceArn"
    }
  }
}

data "aws_iam_policy_document" "test_agent_permissions" {
  statement {
    actions = ["bedrock:InvokeModel"]
    resources = [
      "arn:${data.aws_partition.current_agent.partition}:bedrock:${data.aws_region.current_agent.name}::foundation-model/%[2]s",
    ]
  }

  statement {
    actions = ["bedrock:GetAgentAlias", "bedrock:InvokeAgent"]
    resources = [
      "arn:${data.aws_partition.current_agent.partition}:bedrock:${data.aws_region.current_agent.name}:${data.aws_caller_identity.current_agent.account_id}:agent/*",
      "arn:${data.aws_partition.current_agent.partition}:bedrock:${data.aws_region.current_agent.name}:${data.aws_caller_identity.current_agent.account_id}:agent-alias/*"
    ]
  }
}

resource "aws_iam_role_policy" "test_agent" {
  role   = aws_iam_role.test_agent.id
  policy = data.aws_iam_policy_document.test_agent_permissions.json
}

resource "aws_iam_role_policy_attachment" "test_agent_s3" {
  role       = aws_iam_role.test_agent.id
  policy_arn = "arn:${data.aws_partition.current_agent.partition}:iam::aws:policy/AmazonS3FullAccess"
}

resource "aws_iam_role_policy_attachment" "test_agent_lambda" {
  role       = aws_iam_role.test_agent.id
  policy_arn = "arn:${data.aws_partition.current_agent.partition}:iam::aws:policy/AWSLambda_FullAccess"
}


data "aws_caller_identity" "current_agent" {}

data "aws_region" "current_agent" {}

data "aws_partition" "current_agent" {}
`, rName, model)
}

func testAccAgent_guardrail(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "test_agent_guardrail_permissions" {
  statement {
    actions   = ["bedrock:ApplyGuardrail"]
    resources = [aws_bedrock_guardrail.test.guardrail_arn]
  }
}

resource "aws_iam_role_policy" "test_agent_guardrail" {
  role   = aws_iam_role.test_agent.id
  policy = data.aws_iam_policy_document.test_agent_guardrail_permissions.json
}

resource "aws_bedrock_guardrail" "test" {
  name                      = %[1]q
  description               = %[1]q
  blocked_input_messaging   = "Sorry, I cannot answer this question."
  blocked_outputs_messaging = "Sorry, I cannot answer this question."
  content_policy_config {
    filters_config {
      input_strength  = "MEDIUM"
      output_strength = "MEDIUM"
      type            = "HATE"
    }
  }
}
`, rName)
}

func testAccAgentConfig_basic(rName, model, description string) string {
	return acctest.ConfigCompose(testAccAgent_base(rName, model), fmt.Sprintf(`
resource "aws_bedrockagent_agent" "test" {
  agent_name                  = %[1]q
  agent_resource_role_arn     = aws_iam_role.test_agent.arn
  description                 = %[3]q
  idle_session_ttl_in_seconds = 500
  instruction                 = file("${path.module}/test-fixtures/instruction.txt")
  foundation_model            = %[2]q
}
`, rName, model, description))
}

func testAccAgentConfig_tags1(rName, model, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccAgent_base(rName, model), fmt.Sprintf(`
resource "aws_bedrockagent_agent" "test" {
  agent_name              = %[1]q
  agent_resource_role_arn = aws_iam_role.test_agent.arn
  instruction             = file("${path.module}/test-fixtures/instruction.txt")
  foundation_model        = %[2]q

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, model, tagKey1, tagValue1))
}

func testAccAgentConfig_tags2(rName, model, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccAgent_base(rName, model), fmt.Sprintf(`
resource "aws_bedrockagent_agent" "test" {
  agent_name              = %[1]q
  agent_resource_role_arn = aws_iam_role.test_agent.arn
  instruction             = file("${path.module}/test-fixtures/instruction.txt")
  foundation_model        = %[2]q

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, model, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccAgentConfig_full(rName, model, desc string) string {
	return acctest.ConfigCompose(testAccAgent_base(rName, model), fmt.Sprintf(`
resource "aws_bedrockagent_agent" "test" {
  agent_name                  = %[1]q
  agent_resource_role_arn     = aws_iam_role.test_agent.arn
  description                 = %[3]q
  idle_session_ttl_in_seconds = 500
  instruction                 = file("${path.module}/test-fixtures/instruction.txt")
  foundation_model            = %[2]q
  skip_resource_in_use_check  = true

  prompt_override_configuration {
    override_lambda = null
    prompt_configurations = [
      {
        base_prompt_template = file("${path.module}/test-fixtures/pre-processing.txt")
        inference_configuration = [
          {
            max_length     = 2048
            stop_sequences = ["Human:"]
            temperature    = 0
            top_k          = 250
            top_p          = 1
          },
        ]
        parser_mode          = "DEFAULT"
        prompt_creation_mode = "OVERRIDDEN"
        prompt_state         = "ENABLED"
        prompt_type          = "PRE_PROCESSING"
      },
      {
        base_prompt_template = file("${path.module}/test-fixtures/knowledge-base-response-generation.txt")
        inference_configuration = [
          {
            max_length     = 2048
            stop_sequences = ["Human:"]
            temperature    = 0
            top_k          = 250
            top_p          = 1
          },
        ]
        parser_mode          = "DEFAULT"
        prompt_creation_mode = "OVERRIDDEN"
        prompt_state         = "ENABLED"
        prompt_type          = "KNOWLEDGE_BASE_RESPONSE_GENERATION"
      },
      {
        base_prompt_template = file("${path.module}/test-fixtures/orchestration.txt")
        inference_configuration = [
          {
            max_length = 2048
            stop_sequences = [
              "</function_call>",
              "</answer>",
              "</error>",
            ]
            temperature = 0
            top_k       = 250
            top_p       = 1
          },
        ]
        parser_mode          = "DEFAULT"
        prompt_creation_mode = "OVERRIDDEN"
        prompt_state         = "ENABLED"
        prompt_type          = "ORCHESTRATION"
      },
      {
        base_prompt_template = file("${path.module}/test-fixtures/post-processing.txt")
        inference_configuration = [
          {
            max_length     = 2048
            stop_sequences = ["Human:"]
            temperature    = 0
            top_k          = 250
            top_p          = 1
          },
        ]
        parser_mode          = "DEFAULT"
        prompt_creation_mode = "OVERRIDDEN"
        prompt_state         = "DISABLED"
        prompt_type          = "POST_PROCESSING"
      },
    ]
  }

}
`, rName, model, desc))
}

func testAccAgentConfig_singlePrompt(rName, model, desc string) string {
	return acctest.ConfigCompose(testAccAgent_base(rName, model), fmt.Sprintf(`
resource "aws_bedrockagent_agent" "test" {
  agent_name                  = %[1]q
  agent_resource_role_arn     = aws_iam_role.test_agent.arn
  description                 = %[3]q
  idle_session_ttl_in_seconds = 500
  instruction                 = file("${path.module}/test-fixtures/instruction.txt")
  foundation_model            = %[2]q
  skip_resource_in_use_check  = true

  prompt_override_configuration {
    override_lambda = null
    prompt_configurations = [
      {
        base_prompt_template = file("${path.module}/test-fixtures/post-processing.txt")
        inference_configuration = [
          {
            max_length     = 2048
            stop_sequences = ["Human:"]
            temperature    = 0
            top_k          = 250
            top_p          = 1
          },
        ]
        parser_mode          = "DEFAULT"
        prompt_creation_mode = "OVERRIDDEN"
        prompt_state         = "DISABLED"
        prompt_type          = "POST_PROCESSING"
      },
    ]
  }

}
`, rName, model, desc))
}

func testAccAgentConfig_singlePromptUpdate(rName, model, desc, timeout string) string {
	return acctest.ConfigCompose(testAccAgent_base(rName, model), fmt.Sprintf(`
resource "aws_bedrockagent_agent" "test" {
  agent_name                  = %[1]q
  agent_resource_role_arn     = aws_iam_role.test_agent.arn
  description                 = %[3]q
  idle_session_ttl_in_seconds = %[4]s
  instruction                 = file("${path.module}/test-fixtures/instruction.txt")
  foundation_model            = %[2]q
  skip_resource_in_use_check  = true

  prompt_override_configuration {
    override_lambda = null
    prompt_configurations = [
      {
        base_prompt_template = file("${path.module}/test-fixtures/post-processing.txt")
        inference_configuration = [
          {
            max_length     = 2048
            stop_sequences = ["Human:"]
            temperature    = 0
            top_k          = 250
            top_p          = 1
          },
        ]
        parser_mode          = "DEFAULT"
        prompt_creation_mode = "OVERRIDDEN"
        prompt_state         = "DISABLED"
        prompt_type          = "POST_PROCESSING"
      },
    ]
  }

}
`, rName, model, desc, timeout))
}

func testAccAgentConfig_guardrail_noConfig(rName, model string) string {
	return acctest.ConfigCompose(testAccAgent_base(rName, model), testAccAgent_guardrail(rName), fmt.Sprintf(`
resource "aws_bedrockagent_agent" "test" {
  agent_name                 = %[1]q
  agent_resource_role_arn    = aws_iam_role.test_agent.arn
  instruction                = file("${path.module}/test-fixtures/instruction.txt")
  foundation_model           = %[2]q
  skip_resource_in_use_check = true
}
`, rName, model))
}

func testAccAgentConfig_guardrail_withConfig(rName, model, guardrailVersion string) string {
	return acctest.ConfigCompose(testAccAgent_base(rName, model), testAccAgent_guardrail(rName), fmt.Sprintf(`
resource "aws_bedrockagent_agent" "test" {
  agent_name                 = %[1]q
  agent_resource_role_arn    = aws_iam_role.test_agent.arn
  instruction                = file("${path.module}/test-fixtures/instruction.txt")
  foundation_model           = %[2]q
  skip_resource_in_use_check = true
  guardrail_configuration {
    guardrail_identifier = aws_bedrock_guardrail.test.guardrail_id
    guardrail_version    = %[3]q
  }
}
`, rName, model, guardrailVersion))
}

func testAccAgentConfig_kms(rName, model, description, timeout string) string {
	return acctest.ConfigCompose(testAccAgent_base(rName, model), fmt.Sprintf(`
resource "aws_bedrockagent_agent" "test" {
  agent_name                  = %[1]q
  agent_resource_role_arn     = aws_iam_role.test_agent.arn
  customer_encryption_key_arn = aws_kms_key.test_agent.arn
  description                 = %[3]q
  idle_session_ttl_in_seconds = %[4]s
  instruction                 = file("${path.module}/test-fixtures/instruction.txt")
  foundation_model            = %[2]q
}

resource "aws_kms_key" "test_agent" {
  description             = "Agent test key"
  enable_key_rotation     = true
  deletion_window_in_days = 7
}
`, rName, model, description, timeout))
}

func testAccAgentConfig_agentCollaboration(rName, model, description, collaboration string) string {
	return acctest.ConfigCompose(testAccAgent_base(rName, model), fmt.Sprintf(`
resource "aws_bedrockagent_agent" "test" {
  agent_collaboration         = %[4]q
  agent_name                  = %[1]q
  agent_resource_role_arn     = aws_iam_role.test_agent.arn
  description                 = %[3]q
  idle_session_ttl_in_seconds = 500
  instruction                 = file("${path.module}/test-fixtures/instruction.txt")
  foundation_model            = %[2]q
  prepare_agent               = false
}
`, rName, model, description, collaboration))
}

func testAccAgentConfig_memoryConfiguration(rName, model, description string) string {
	return acctest.ConfigCompose(testAccAgent_base(rName, model), fmt.Sprintf(`
resource "aws_bedrockagent_agent" "test" {
  agent_name                  = %[1]q
  agent_resource_role_arn     = aws_iam_role.test_agent.arn
  description                 = %[3]q
  idle_session_ttl_in_seconds = 500
  instruction                 = file("${path.module}/test-fixtures/instruction.txt")
  foundation_model            = %[2]q
  skip_resource_in_use_check  = true

  memory_configuration {
    enabled_memory_types = ["SESSION_SUMMARY"]
    storage_days         = 15
  }
}
`, rName, model, description))
}
