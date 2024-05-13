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

func TestAccBedrockAgentAgentActionGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent_action_group.test"
	var v awstypes.AgentActionGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentActionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentActionGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentActionGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "action_group_name", rName),
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

func TestAccBedrockAgentAgentActionGroup_s3APISchema(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent_action_group.test"
	var v awstypes.AgentActionGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentActionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentActionGroupConfig_s3APISchema(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentActionGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "action_group_name", rName),
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

func TestAccBedrockAgentAgentActionGroup_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent_action_group.test"
	var v awstypes.AgentActionGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentActionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentActionGroupConfig_s3APISchema(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentActionGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "action_group_name", rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_resource_in_use_check"},
			},
			{
				Config: testAccAgentActionGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentActionGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "action_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Basic Agent Action"),
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

func TestAccBedrockAgentAgentActionGroup_functionSchema(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent_action_group.test"
	var v awstypes.AgentActionGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentActionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentActionGroupConfig_functionSchema(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentActionGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "action_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "function_schema.#", acctest.CtOne),
					resource.TestCheckResourceAttr(resourceName, "function_schema.0.functions.#", acctest.CtOne),
					resource.TestCheckResourceAttr(resourceName, "function_schema.0.functions.0.name", "sayHello"),
					resource.TestCheckResourceAttr(resourceName, "function_schema.0.functions.0.description", "Says Hello"),
					resource.TestCheckResourceAttr(resourceName, "function_schema.0.functions.0.parameters.#", acctest.CtTwo),
					resource.TestCheckResourceAttr(resourceName, "function_schema.0.functions.0.parameters.0.map_block_key", names.AttrMessage),
					resource.TestCheckResourceAttr(resourceName, "function_schema.0.functions.0.parameters.0.type", "string"),
					resource.TestCheckResourceAttr(resourceName, "function_schema.0.functions.0.parameters.0.description", "The Hello message"),
					resource.TestCheckResourceAttr(resourceName, "function_schema.0.functions.0.parameters.0.required", "true"),
					resource.TestCheckResourceAttr(resourceName, "function_schema.0.functions.0.parameters.1.map_block_key", "unused"),
					resource.TestCheckResourceAttr(resourceName, "function_schema.0.functions.0.parameters.1.type", "integer"),
					resource.TestCheckResourceAttr(resourceName, "function_schema.0.functions.0.parameters.1.required", "false"),
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

func testAccCheckAgentActionGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagent_agent_action_group" {
				continue
			}

			_, err := tfbedrockagent.FindAgentActionGroupByThreePartKey(ctx, conn, rs.Primary.Attributes["action_group_id"], rs.Primary.Attributes["agent_id"], rs.Primary.Attributes["agent_version"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent Action Group %s still exists", rs.Primary.ID)
		}

		return nil
	}
}
func testAccCheckAgentActionGroupExists(ctx context.Context, n string, v *awstypes.AgentActionGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		output, err := tfbedrockagent.FindAgentActionGroupByThreePartKey(ctx, conn, rs.Primary.Attributes["action_group_id"], rs.Primary.Attributes["agent_id"], rs.Primary.Attributes["agent_version"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAgentActionGroupConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccAgentConfig_basic(rName, "anthropic.claude-v2", "basic claude"),
		testAccAgentActionGroupConfig_lambda(rName),
		fmt.Sprintf(`
resource "aws_bedrockagent_agent_action_group" "test" {
  action_group_name          = %[1]q
  agent_id                   = aws_bedrockagent_agent.test.agent_id
  agent_version              = "DRAFT"
  description                = "Basic Agent Action"
  skip_resource_in_use_check = true
  action_group_executor {
    lambda = aws_lambda_function.test_lambda.arn
  }
  api_schema {
    payload = file("${path.module}/test-fixtures/api_schema.yaml")
  }
}
`, rName))
}

func testAccAgentActionGroupConfig_s3APISchema(rName string) string {
	return acctest.ConfigCompose(testAccAgentConfig_basic(rName, "anthropic.claude-v2", "basic claude"),
		testAccAgentActionGroupConfig_lambda(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = "api_schema.yaml"
  source = "${path.module}/test-fixtures/api_schema.yaml"
}

resource "aws_bedrockagent_agent_action_group" "test" {
  action_group_name          = %[1]q
  agent_id                   = aws_bedrockagent_agent.test.agent_id
  agent_version              = "DRAFT"
  skip_resource_in_use_check = true
  action_group_executor {
    lambda = aws_lambda_function.test_lambda.arn
  }
  api_schema {
    s3 {
      s3_bucket_name = aws_s3_bucket.test.id
      s3_object_key  = aws_s3_object.test.key
    }
  }
  depends_on = [aws_s3_object.test]
}
`, rName))
}

func testAccAgentActionGroupConfig_lambda(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "lambda_assume" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "lambda" {
  name_prefix        = %[1]q
  assume_role_policy = data.aws_iam_policy_document.lambda_assume.json
}

resource "aws_lambda_function" "test_lambda" {
  filename      = "${path.module}/test-fixtures/lambda_function.zip"
  function_name = %[1]q
  role          = aws_iam_role.lambda.arn
  handler       = "lambda_handler"

  source_code_hash = filebase64sha256("${path.module}/test-fixtures/lambda_function.zip")

  runtime = "python3.9"
}
`, rName)
}

func testAccAgentActionGroupConfig_functionSchema(rName string) string {
	return acctest.ConfigCompose(testAccAgentConfig_basic(rName, "anthropic.claude-v2", "basic claude"),
		testAccAgentActionGroupConfig_lambda(rName),
		fmt.Sprintf(`
resource "aws_bedrockagent_agent_action_group" "test" {
  action_group_name          = %[1]q
  agent_id                   = aws_bedrockagent_agent.test.agent_id
  agent_version              = "DRAFT"
  description                = "Basic Agent Action"
  skip_resource_in_use_check = true
  action_group_executor {
    lambda = aws_lambda_function.test_lambda.arn
  }
  function_schema {
    functions {
      name        = "sayHello"
      description = "Says Hello"
      parameters {
        map_block_key = "message"
        type          = "string"
        description   = "The Hello message"
        required      = true
      }
      parameters {
        map_block_key = "unused"
        type          = "integer"
        required      = false
      }
    }
  }
}
`, rName))
}
