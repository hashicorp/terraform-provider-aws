// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCoreResourcePolicy_runtime_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var resourcepolicy string
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	rImageUri := acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI")
	resourceName := "aws_bedrockagentcore_resource_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_runtime(rName, rImageUri),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &resourcepolicy),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, names.AttrPolicy, "user"},
			},
		},
	})
}

func TestAccBedrockAgentCoreResourcePolicy_endpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var resourcepolicy string
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	rImageUri := acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI")
	resourceName := "aws_bedrockagentcore_resource_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_endpoint(rName, rImageUri),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &resourcepolicy),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, names.AttrPolicy, "user"},
			},
		},
	})
}

func TestAccBedrockAgentCoreResourcePolicy_gateway_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var resourcepolicy string
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_resource_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_gateway(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &resourcepolicy),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, names.AttrPolicy, "user"},
			},
		},
	})
}

func testAccCheckResourcePolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_resource_policy" {
				continue
			}

			// Call GetResourcePolicy directly to verify destruction
			input := bedrockagentcorecontrol.GetResourcePolicyInput{
				ResourceArn: aws.String(rs.Primary.ID),
			}

			_, err := conn.GetResourcePolicy(ctx, &input)
			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.BedrockAgentCore, create.ErrActionCheckingDestroyed, tfbedrockagentcore.ResNameResourcePolicy, rs.Primary.ID, err)
			}

			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingDestroyed, tfbedrockagentcore.ResNameResourcePolicy, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckResourcePolicyExists(ctx context.Context, t *testing.T, name string, resourcepolicy *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameResourcePolicy, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameResourcePolicy, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		input := bedrockagentcorecontrol.GetResourcePolicyInput{
			ResourceArn: aws.String(rs.Primary.ID),
		}

		out, err := conn.GetResourcePolicy(ctx, &input)
		if err != nil {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameResourcePolicy, rs.Primary.ID, err)
		}

		if out.Policy != nil {
			*resourcepolicy = aws.ToString(out.Policy)
		} else {
			*resourcepolicy = ""
		}

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

	// Use ListAgentRuntimes as a lightweight service availability check
	input := bedrockagentcorecontrol.ListAgentRuntimesInput{}

	_, err := conn.ListAgentRuntimes(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
func testAccResourcePolicyConfig_runtime(rName string, rImageUri string) string {
	return acctest.ConfigCompose(testAccAgentRuntimeConfig_basic(rName, rImageUri), `
data "aws_iam_policy_document" "resource_policy" {
  statement {
    effect = "Allow"
    actions = [
      "bedrock-agentcore:InvokeAgentRuntime",
    ]
    principals {
      type        = "*"
      identifiers = ["*"]
    }
    resources = [
      aws_bedrockagentcore_agent_runtime.test.agent_runtime_arn
    ]
  }
}

resource "aws_bedrockagentcore_resource_policy" "test" {
  resource_arn = aws_bedrockagentcore_agent_runtime.test.agent_runtime_arn
  policy       = data.aws_iam_policy_document.resource_policy.json
}`)
}

func testAccResourcePolicyConfig_endpoint(rName string, rImageUri string) string {
	return acctest.ConfigCompose(testAccAgentRuntimeEndpointConfig_basic(rName, rImageUri), `
data "aws_iam_policy_document" "resource_policy" {
  statement {
    effect = "Allow"
    actions = [
      "bedrock-agentcore:InvokeAgentRuntime",
    ]
    principals {
      type        = "*"
      identifiers = ["*"]
    }
    resources = [
        aws_bedrockagentcore_agent_runtime_endpoint.test.agent_runtime_arn
    ]
  }
}

resource "aws_bedrockagentcore_resource_policy" "test" {
  resource_arn = aws_bedrockagentcore_agent_runtime_endpoint.test.agent_runtime_arn
  policy       = data.aws_iam_policy_document.resource_policy.json
}`)
}

func testAccResourcePolicyConfig_gateway(rName string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_basic(rName), `
data "aws_iam_policy_document" "resource_policy" {
  statement {
    effect = "Allow"
    actions = [
      "bedrock-agentcore:InvokeAgentRuntime",
    ]
    principals {
      type        = "*"
      identifiers = ["*"]
    }
    resources = [
        aws_bedrockagentcore_gateway.test.agent_runtime_arn
    ]
  }
}

resource "aws_bedrockagentcore_resource_policy" "test" {
  resource_arn = aws_bedrockagentcore_gateway.example.gateway_arn
  policy       = data.aws_iam_policy_document.resource_policy.json
}`)
}
