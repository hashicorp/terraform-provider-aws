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
					resource.TestCheckResourceAttr(resourceName, "action_group_state", "ENABLED"),
					resource.TestCheckResourceAttrPair(resourceName, "agent_id", "aws_bedrockagent_agent.test", "agent_id"),
					resource.TestCheckResourceAttr(resourceName, "agent_version", "DRAFT"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Basic Agent Action"),
					resource.TestCheckNoResourceAttr(resourceName, "parent_action_group_signature"),
					resource.TestCheckResourceAttr(resourceName, "skip_resource_in_use_check", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "action_group_executor.#", acctest.Ct1),
					resource.TestCheckNoResourceAttr(resourceName, "action_group_executor.0.custom_control"),
					resource.TestCheckResourceAttrPair(resourceName, "action_group_executor.0.lambda", "aws_lambda_function.test_lambda", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "api_schema.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "api_schema.0.payload"),
					resource.TestCheckResourceAttr(resourceName, "api_schema.0.s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "function_schema.#", acctest.Ct0),
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

func TestAccBedrockAgentAgentActionGroup_APISchema_s3(t *testing.T) {
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
				Config: testAccAgentActionGroupConfig_APISchema_s3(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentActionGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "action_group_name", rName),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "api_schema.#", acctest.Ct1),
					resource.TestCheckNoResourceAttr(resourceName, "api_schema.0.payload"),
					resource.TestCheckResourceAttr(resourceName, "api_schema.0.s3.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "api_schema.0.s3.0.s3_bucket_name", "aws_s3_bucket.test", names.AttrBucket),
					resource.TestCheckResourceAttrPair(resourceName, "api_schema.0.s3.0.s3_object_key", "aws_s3_object.test", names.AttrKey),
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
				Config: testAccAgentActionGroupConfig_APISchema_s3(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentActionGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "action_group_name", rName),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "api_schema.#", acctest.Ct1),
					resource.TestCheckNoResourceAttr(resourceName, "api_schema.0.payload"),
					resource.TestCheckResourceAttr(resourceName, "api_schema.0.s3.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "api_schema.0.s3.0.s3_bucket_name", "aws_s3_bucket.test", names.AttrBucket),
					resource.TestCheckResourceAttrPair(resourceName, "api_schema.0.s3.0.s3_object_key", "aws_s3_object.test", names.AttrKey),
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
					resource.TestCheckResourceAttr(resourceName, "api_schema.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "api_schema.0.payload"),
					resource.TestCheckResourceAttr(resourceName, "api_schema.0.s3.#", acctest.Ct0),
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

func TestAccBedrockAgentAgentActionGroup_FunctionSchema_memberFunctions(t *testing.T) {
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
				Config: testAccAgentActionGroupConfig_FunctionSchema_memberFunctions(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentActionGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "action_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "function_schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "function_schema.0.member_functions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "function_schema.0.member_functions.0.functions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "function_schema.0.member_functions.0.functions.0.name", "sayHello"),
					resource.TestCheckResourceAttr(resourceName, "function_schema.0.member_functions.0.functions.0.description", "Says Hello"),
					resource.TestCheckResourceAttr(resourceName, "function_schema.0.member_functions.0.functions.0.parameters.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "function_schema.0.member_functions.0.functions.0.parameters.0.map_block_key", names.AttrMessage),
					resource.TestCheckResourceAttr(resourceName, "function_schema.0.member_functions.0.functions.0.parameters.0.type", "string"),
					resource.TestCheckResourceAttr(resourceName, "function_schema.0.member_functions.0.functions.0.parameters.0.description", "The Hello message"),
					resource.TestCheckResourceAttr(resourceName, "function_schema.0.member_functions.0.functions.0.parameters.0.required", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "function_schema.0.member_functions.0.functions.0.parameters.1.map_block_key", "unused"),
					resource.TestCheckResourceAttr(resourceName, "function_schema.0.member_functions.0.functions.0.parameters.1.type", "integer"),
					resource.TestCheckResourceAttr(resourceName, "function_schema.0.member_functions.0.functions.0.parameters.1.required", acctest.CtFalse),
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

func TestAccBedrockAgentAgentActionGroup_ActionGroupExecutor_customControl(t *testing.T) {
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
				Config: testAccAgentActionGroupConfig_ActionGroupExecutor_customControl(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentActionGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "action_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "action_group_executor.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "action_group_executor.0.custom_control", "RETURN_CONTROL"),
					resource.TestCheckNoResourceAttr(resourceName, "action_group_executor.0.lambda"),
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

func testAccAgentActionGroupConfig_APISchema_s3(rName string) string {
	return acctest.ConfigCompose(testAccAgentConfig_basic(rName, "anthropic.claude-v2", "basic claude"),
		testAccAgentActionGroupConfig_lambda(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
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
      s3_bucket_name = aws_s3_bucket.test.bucket
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

func testAccAgentActionGroupConfig_FunctionSchema_memberFunctions(rName string) string {
	return acctest.ConfigCompose(testAccAgentConfig_basic(rName, "anthropic.claude-v2", "basic claude"),
		testAccAgentActionGroupConfig_lambda(rName),
		fmt.Sprintf(`
resource "aws_bedrockagent_agent_action_group" "test" {
  action_group_name          = %[1]q
  agent_id                   = aws_bedrockagent_agent.test.agent_id
  agent_version              = "DRAFT"
  skip_resource_in_use_check = true
  action_group_executor {
    lambda = aws_lambda_function.test_lambda.arn
  }
  function_schema {
    member_functions {
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
}
`, rName))
}

func testAccAgentActionGroupConfig_ActionGroupExecutor_customControl(rName string) string {
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
    custom_control = "RETURN_CONTROL"
  }
  api_schema {
    payload = file("${path.module}/test-fixtures/api_schema.yaml")
  }
}
`, rName))
}
