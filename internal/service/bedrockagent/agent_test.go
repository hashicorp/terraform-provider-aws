// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgent_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent.test"
	var v bedrockagent.GetAgentOutput

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
					resource.TestCheckResourceAttr(resourceName, "description", "basic claude"),
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
	var v bedrockagent.GetAgentOutput

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
					resource.TestCheckResourceAttr(resourceName, "description", "basic claude"),
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
	var v bedrockagent.GetAgentOutput

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
					resource.TestCheckResourceAttr(resourceName, "description", "basic claude"),
				),
			},
			{
				Config: testAccAgentConfig_basic(rName+"-2", "anthropic.claude-v2", "basic claude"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_name", rName+"-2"),
					resource.TestCheckResourceAttr(resourceName, "prompt_override_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "description", "basic claude"),
				),
			},
			{
				Config: testAccAgentConfig_basic(rName+"-3", "anthropic.claude-v2", "basic claude again"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_name", rName+"-3"),
					resource.TestCheckResourceAttr(resourceName, "prompt_override_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "description", "basic claude again"),
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
	var agent bedrockagent.GetAgentOutput

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

			_, err := findAgentByID(ctx, conn, rs.Primary.ID)

			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.BedrockAgent, create.ErrActionCheckingDestroyed, "Bedrock Agent", rs.Primary.ID, err)
			}

			return create.Error(names.BedrockAgent, create.ErrActionCheckingDestroyed, "Bedrock Agent", rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckAgentExists(ctx context.Context, n string, v *bedrockagent.GetAgentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		output, err := findAgentByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func findAgentByID(ctx context.Context, conn *bedrockagent.Client, id string) (*bedrockagent.GetAgentOutput, error) {
	input := &bedrockagent.GetAgentInput{
		AgentId: aws.String(id),
	}

	output, err := conn.GetAgent(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func testAccAgentConfig_basic(rName, model, description string) string {
	return acctest.ConfigCompose(testAccBedrockRole(rName, model), fmt.Sprintf(`
resource "aws_bedrockagent_agent" "test" {
  agent_name              = %[1]q
  agent_resource_role_arn = aws_iam_role.test.arn
  description             = %[3]q
  idle_ttl                = 500
  foundation_model        = %[2]q
}
`, rName, model, description))
}

func testAccAgentConfig_tags1(rName, model, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccBedrockRole(rName, model), fmt.Sprintf(`
resource "aws_bedrockagent_agent" "test" {
  agent_name              = %[1]q
  agent_resource_role_arn = aws_iam_role.test.arn
  idle_ttl                = 500
  foundation_model        = %[2]q

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, model, tagKey1, tagValue1))
}

func testAccAgentConfig_tags2(rName, model, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccBedrockRole(rName, model), fmt.Sprintf(`
resource "aws_bedrockagent_agent" "test" {
  agent_name              = %[1]q
  agent_resource_role_arn = aws_iam_role.test.arn
  idle_ttl                = 500
  foundation_model        = %[2]q

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, model, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccAgentConfig_full(rName, model, desc string) string {
	return acctest.ConfigCompose(testAccBedrockRole(rName, model), fmt.Sprintf(`
resource "aws_bedrockagent_agent" "test" {
  agent_name              = %[1]q
  agent_resource_role_arn = aws_iam_role.test.arn
  description             = %[3]q
  idle_ttl                = 500
  foundation_model        = %[2]q
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
            topk           = 250
            topp           = 1
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
            topk           = 250
            topp           = 1
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
            topk        = 250
            topp        = 1
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
            topk           = 250
            topp           = 1
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

func testAccBedrockRole(rName, model string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
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
      values   = [data.aws_caller_identity.current.account_id]
      variable = "aws:SourceAccount"
    }

    condition {
      test     = "ArnLike"
      values   = ["arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:agent/*"]
      variable = "AWS:SourceArn"
    }
  }
}

data "aws_iam_policy_document" "test_agent_permissions" {
  statement {
    actions = ["bedrock:InvokeModel"]
    resources = [
      "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}::foundation-model/%[2]s",
    ]
  }
}

resource "aws_iam_role_policy" "test" {
  policy = data.aws_iam_policy_document.test_agent_permissions.json
  role   = aws_iam_role.test.id
}

data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

data "aws_partition" "current" {}
`, rName, model)
}
