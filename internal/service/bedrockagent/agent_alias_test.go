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

func TestAccBedrockAgentAgentAlias_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent_alias.test"
	var v awstypes.AgentAlias

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentAliasConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentAliasExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_alias_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "agent_alias_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "agent_alias_id"),
					resource.TestCheckResourceAttrSet(resourceName, "agent_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "routing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccBedrockAgentAgentAlias_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent_alias.test"
	var v awstypes.AgentAlias

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentAliasConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentAliasExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbedrockagent.ResourceAgentAlias, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBedrockAgentAgentAlias_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	descriptionOld := "Agent Alias Before Update"
	descriptionNew := "Agent Alias After Update"
	resourceName := "aws_bedrockagent_agent_alias.test"
	var v awstypes.AgentAlias

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentAliasConfig_update(rName, descriptionOld),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentAliasExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_alias_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "agent_alias_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "agent_alias_id"),
					resource.TestCheckResourceAttrSet(resourceName, "agent_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionOld),
					resource.TestCheckResourceAttr(resourceName, "routing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "routing_configuration.0.agent_version", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAgentAliasConfig_update(rName, descriptionNew),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentAliasExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_alias_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "agent_alias_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "agent_alias_id"),
					resource.TestCheckResourceAttrSet(resourceName, "agent_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionNew),
					resource.TestCheckResourceAttr(resourceName, "routing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "routing_configuration.0.agent_version", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccBedrockAgentAgentAlias_routingUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent_alias.test"
	var v awstypes.AgentAlias

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccagentAliasConfig_routingUpdateOne(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentAliasExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_alias_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "agent_alias_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "agent_alias_id"),
					resource.TestCheckResourceAttrSet(resourceName, "agent_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Test Alias"),
					resource.TestCheckResourceAttr(resourceName, "routing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "routing_configuration.0.agent_version", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccagentAliasConfig_routingUpdateTwo(rName, acctest.Ct2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentAliasExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_alias_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "agent_alias_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "agent_alias_id"),
					resource.TestCheckResourceAttrSet(resourceName, "agent_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Test Alias"),
					resource.TestCheckResourceAttr(resourceName, "routing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "routing_configuration.0.agent_version", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccBedrockAgentAgentAlias_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent_alias.test"
	var v awstypes.AgentAlias

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentAliasConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentAliasExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAgentAliasConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentAliasExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAgentAliasConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentAliasExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccBedrockAgentAgentAlias_provisionedThroughput(t *testing.T) {
	acctest.Skip(t, "Bedrock Provisioned Model Throughput for Antropic Claude 2 has a minimum 1 month commitment and costs > $45K/month")

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent_alias.test2"
	var v awstypes.AgentAlias

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentAliasConfig_provisionedThroughout(rName, acctest.Ct1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentAliasExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_alias_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "agent_alias_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "agent_alias_id"),
					resource.TestCheckResourceAttrSet(resourceName, "agent_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Test Alias"),
					resource.TestCheckResourceAttr(resourceName, "routing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "routing_configuration.0.agent_version", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "routing_configuration.0.provisioned_throughput"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func testAccCheckAgentAliasDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagent_agent_alias" {
				continue
			}

			_, err := tfbedrockagent.FindAgentAliasByTwoPartKey(ctx, conn, rs.Primary.Attributes["agent_alias_id"], rs.Primary.Attributes["agent_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent Alias %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAgentAliasExists(ctx context.Context, n string, v *awstypes.AgentAlias) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		output, err := tfbedrockagent.FindAgentAliasByTwoPartKey(ctx, conn, rs.Primary.Attributes["agent_alias_id"], rs.Primary.Attributes["agent_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAgentAliasConfig_provisionedModelThroughputBase(rName string) string {
	return fmt.Sprintf(`
data "aws_bedrock_foundation_model" "test" {
  model_id = "anthropic.claude-v2:0:18k"
}

resource "aws_bedrock_provisioned_model_throughput" "test" {
  provisioned_model_name = %[1]q
  model_arn              = data.aws_bedrock_foundation_model.test.model_arn
  commitment_duration    = "OneMonth"
  model_units            = 1
}
`, rName)
}

func testAccAgentAliasConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccAgentConfig_basic(rName, "anthropic.claude-v2", "basic claude"), testAccAgentAliasConfig_alias(rName))
}

func testAccAgentAliasConfig_alias(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagent_agent_alias" "test" {
  agent_alias_name = %[1]q
  agent_id         = aws_bedrockagent_agent.test.agent_id
  description      = "Test Alias"
}
`, rName)
}

func testAccAgentAliasConfig_routing(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagent_agent_alias" "test" {
  agent_alias_name = %[1]q
  agent_id         = aws_bedrockagent_agent.test.agent_id
  description      = "Test Alias"
  routing_configuration {
    agent_version = %[2]q
  }
}
`, rName, version)
}

func testAccagentAliasConfig_routingUpdateOne(rName string) string {
	return acctest.ConfigCompose(
		testAccAgentConfig_basic(rName, "anthropic.claude-v2", "basic claude"),
		testAccAgentAliasConfig_alias(rName),
		fmt.Sprintf(`
resource "aws_bedrockagent_agent_alias" "second" {
  agent_alias_name = %[1]q
  agent_id         = aws_bedrockagent_agent.test.agent_id
  description      = "Test Alias"
  depends_on       = [aws_bedrockagent_agent_alias.test]
}
`, rName+acctest.Ct2),
	)
}

func testAccagentAliasConfig_routingUpdateTwo(rName, version string) string {
	return acctest.ConfigCompose(
		testAccAgentConfig_basic(rName, "anthropic.claude-v2", "basic claude"),
		testAccAgentAliasConfig_routing(rName, version),
	)
}

func testAccAgentAliasConfig_update(rName, desc string) string {
	return acctest.ConfigCompose(testAccAgentConfig_basic(rName, "anthropic.claude-v2", "basic claude"), fmt.Sprintf(`
resource "aws_bedrockagent_agent_alias" "test" {
  agent_alias_name = %[1]q
  agent_id         = aws_bedrockagent_agent.test.agent_id
  description      = %[2]q
}
`, rName, desc))
}

func testAccAgentAliasConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccAgentConfig_basic(rName, "anthropic.claude-v2", "basic claude"), fmt.Sprintf(`
resource "aws_bedrockagent_agent_alias" "test" {
  agent_alias_name = %[1]q
  agent_id         = aws_bedrockagent_agent.test.agent_id
  description      = "Test Alias"
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccAgentAliasConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccAgentConfig_basic(rName, "anthropic.claude-v2", "basic claude"), fmt.Sprintf(`
resource "aws_bedrockagent_agent_alias" "test" {
  agent_alias_name = %[1]q
  agent_id         = aws_bedrockagent_agent.test.agent_id
  description      = "Test Alias"
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccAgentAliasConfig_provisionedThroughout(rName, version string) string {
	return acctest.ConfigCompose(
		testAccAgentAliasConfig_provisionedModelThroughputBase(rName),
		testAccAgentConfig_basic(rName, "anthropic.claude-v2", "basic claude"),
		testAccAgentAliasConfig_alias(rName),
		fmt.Sprintf(`
resource "aws_bedrockagent_agent_alias" "test2" {
  agent_alias_name = %[1]q
  agent_id         = aws_bedrockagent_agent.test.agent_id
  description      = "Test Alias"
  routing_configuration {
    agent_version          = %[2]q
    provisioned_throughput = aws_bedrock_provisioned_model_throughput.test.provisioned_model_arn
  }
}
`, rName+acctest.Ct2, version))
}
