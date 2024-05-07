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

func TestAccBedrockAgent_basic(t *testing.T) {
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
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "basic claude"),
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

func TestAccBedrockAgent_full(t *testing.T) {
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

func TestAccBedrockAgent_update(t *testing.T) {
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
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBedrockAgent_tags(t *testing.T) {
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
				Config: testAccAgentConfig_tags1(rName, "anthropic.claude-v2", "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &agent),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAgentConfig_tags2(rName, "anthropic.claude-v2", "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &agent),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAgentConfig_tags1(rName, "anthropic.claude-v2", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &agent),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
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
}

resource "aws_iam_role_policy" "test_agent" {
  policy = data.aws_iam_policy_document.test_agent_permissions.json
  role   = aws_iam_role.test.id
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
