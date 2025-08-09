// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCoreAgentRuntime_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var agentRuntime bedrockagentcorecontrol.GetAgentRuntimeOutput
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_agent_runtime.test"
	rImageUri := os.Getenv("AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckAgentRuntimes(ctx, t)
			testAccAgentRuntimeImageVersionsPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentRuntimeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentRuntimeConfig_iamRole(rName),
				Check: resource.ComposeTestCheckFunc(
					// Wait for IAM role and policy to propagate
					acctest.CheckSleep(t, 5*time.Second),
				),
			},
			{
				Config: testAccAgentRuntimeConfig_basic(rName, rImageUri),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeExists(ctx, resourceName, &agentRuntime),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "artifact.0.container_configuration.0.container_uri", rImageUri),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVersion),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.network_mode", "PUBLIC"),
					resource.TestCheckResourceAttr(resourceName, "workload_identity_details.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "workload_identity_details.0.workload_identity_arn"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "bedrock-agentcore", regexache.MustCompile(`runtime/.+$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
		},
	})
}

func TestAccBedrockAgentCoreAgentRuntime_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var before, after bedrockagentcorecontrol.GetAgentRuntimeOutput
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_agent_runtime.test"
	rImageUriV1 := os.Getenv("AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI")
	rImageUriV2 := os.Getenv("AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V2_URI")
	clientToken1 := sdkacctest.RandStringFromCharSet(33, sdkacctest.CharSetAlphaNum)
	clientToken2 := sdkacctest.RandStringFromCharSet(33, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckAgentRuntimes(ctx, t)
			testAccAgentRuntimeImageVersionsPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentRuntimeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentRuntimeConfig_iamRole(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckSleep(t, 5*time.Second),
				),
			},
			{
				Config: testAccAgentRuntimeConfig_comprehensive(
					rName,
					"Initial comprehensive agent runtime",
					"aws_iam_role.test.arn",
					"ENV_KEY_1",
					"env_value_1",
					clientToken1,
					rImageUriV1,
					"https://accounts.google.com/.well-known/openid-configuration",
					`"weather", "sports", "news"`,
					`"client-999", "client-888", "client-777"`,
					"HTTP",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "artifact.0.container_configuration.0.container_uri", rImageUriV1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Initial comprehensive agent runtime"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.ENV_KEY_1", "env_value_1"),
					resource.TestCheckResourceAttr(resourceName, "client_token", clientToken1),
					resource.TestCheckResourceAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.discovery_url", "https://accounts.google.com/.well-known/openid-configuration"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.allowed_audience.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.allowed_audience.*", "weather"),
					resource.TestCheckTypeSetElemAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.allowed_audience.*", "sports"),
					resource.TestCheckTypeSetElemAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.allowed_audience.*", "news"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.allowed_clients.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.allowed_clients.*", "client-999"),
					resource.TestCheckTypeSetElemAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.allowed_clients.*", "client-888"),
					resource.TestCheckTypeSetElemAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.allowed_clients.*", "client-777"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.network_mode", "PUBLIC"),
					resource.TestCheckResourceAttr(resourceName, "protocol_configuration.0.server_protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "workload_identity_details.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "workload_identity_details.0.workload_identity_arn"),
				),
			},
			{
				Config: testAccAgentRuntimeConfig_comprehensive(
					rName,
					"Updated comprehensive agent runtime",
					"aws_iam_role.test2.arn",
					"ENV_KEY_2",
					"env_value_2_updated",
					clientToken2,
					rImageUriV2,
					"https://login.microsoftonline.com/common/v2.0/.well-known/openid-configuration",
					`"finance", "technology"`,
					`"client-111", "client-222"`,
					"MCP",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeExists(ctx, resourceName, &after),
					testAccCheckAgentRuntimeNotRecreated(&before, &after),
					resource.TestCheckResourceAttr(resourceName, "artifact.0.container_configuration.0.container_uri", rImageUriV2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated comprehensive agent runtime"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test2", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.ENV_KEY_2", "env_value_2_updated"),
					resource.TestCheckResourceAttr(resourceName, "client_token", clientToken2),
					resource.TestCheckResourceAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.discovery_url", "https://login.microsoftonline.com/common/v2.0/.well-known/openid-configuration"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.allowed_audience.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.allowed_audience.*", "finance"),
					resource.TestCheckTypeSetElemAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.allowed_audience.*", "technology"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.allowed_clients.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.allowed_clients.*", "client-111"),
					resource.TestCheckTypeSetElemAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.allowed_clients.*", "client-222"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.network_mode", "PUBLIC"),
					resource.TestCheckResourceAttr(resourceName, "protocol_configuration.0.server_protocol", "MCP"),
					resource.TestCheckResourceAttr(resourceName, "workload_identity_details.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "workload_identity_details.0.workload_identity_arn"),
				),
			},
		},
	})
}

func TestAccBedrockAgentCoreAgentRuntime_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var agentRuntime bedrockagentcorecontrol.GetAgentRuntimeOutput
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_agent_runtime.test"
	rImageUri := os.Getenv("AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckAgentRuntimes(ctx, t)
			testAccAgentRuntimeImageVersionsPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentRuntimeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentRuntimeConfig_iamRole(rName),
				Check: resource.ComposeTestCheckFunc(
					// Wait for IAM role and policy to propagate
					acctest.CheckSleep(t, 5*time.Second),
				),
			},
			{
				Config: testAccAgentRuntimeConfig_basic(rName, rImageUri),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeExists(ctx, resourceName, &agentRuntime),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbedrockagentcore.ResourceAgentRuntime, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckAgentRuntimeDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_agent_runtime" {
				continue
			}

			_, err := tfbedrockagentcore.FindAgentRuntimeByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.BedrockAgentCore, create.ErrActionCheckingDestroyed, tfbedrockagentcore.ResNameAgentRuntime, rs.Primary.ID, err)
			}

			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingDestroyed, tfbedrockagentcore.ResNameAgentRuntime, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckAgentRuntimeExists(ctx context.Context, name string, agentruntime *bedrockagentcorecontrol.GetAgentRuntimeOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameAgentRuntime, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameAgentRuntime, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindAgentRuntimeByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameAgentRuntime, rs.Primary.ID, err)
		}

		*agentruntime = *resp

		return nil
	}
}

func testAccPreCheckAgentRuntimes(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

	var input bedrockagentcorecontrol.ListAgentRuntimesInput

	_, err := conn.ListAgentRuntimes(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckAgentRuntimeNotRecreated(before, after *bedrockagentcorecontrol.GetAgentRuntimeOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.AgentRuntimeId), aws.ToString(after.AgentRuntimeId); before != after {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingNotRecreated, tfbedrockagentcore.ResNameAgentRuntime, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccAgentRuntimeImageVersionsPreCheck(t *testing.T) {
	if os.Getenv("AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI") == "" {
		t.Skip("AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI env var must be set for BedrockAgentCore Agent Runtime acceptance tests.")
	}
	if os.Getenv("AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V2_URI") == "" {
		t.Skip("AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V2_URI env var must be set for BedrockAgentCore Agent Runtime acceptance tests.")
	}
}

func testAccAgentRuntimeConfig_iamRole(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "test_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["bedrock-agentcore.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "ecr:GetAuthorizationToken",
      "ecr:BatchGetImage",
      "ecr:GetDownloadUrlForLayer"
    ]
    effect    = "Allow"
    resources = ["*"]
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test_assume.json
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role" "test2" {
  name               = "%[1]s-2"
  assume_role_policy = data.aws_iam_policy_document.test_assume.json
}

resource "aws_iam_role_policy" "test2" {
  role   = aws_iam_role.test2.id
  policy = data.aws_iam_policy_document.test.json
}
`, rName)
}

func testAccAgentRuntimeConfig_basic(rName, rImageUri string) string {
	return acctest.ConfigCompose(testAccAgentRuntimeConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_agent_runtime" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  artifact {
    container_configuration {
      container_uri = %[2]q
    }
  }
}
`, rName, rImageUri))
}

func testAccAgentRuntimeConfig_comprehensive(rName, description, roleArn, envVarKey, envVarValue, clientToken, containerUri, discoveryUrl, allowedAudience, allowedClients, serverProtocol string) string {
	return acctest.ConfigCompose(testAccAgentRuntimeConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_agent_runtime" "test" {
  name        = %[1]q
  description = %[2]q
  role_arn    = %[3]s

  environment_variables = {
    %[4]s = %[5]q
  }

  client_token = %[6]q

  artifact {
    container_configuration {
      container_uri = %[7]q
    }
  }

  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = %[8]q
      allowed_audience = [%[9]s]
      allowed_clients  = [%[10]s]
    }
  }

  network_configuration = {
    network_mode = "PUBLIC"
  }

  protocol_configuration {
    server_protocol = %[11]q
  }
}
`, rName, description, roleArn, envVarKey, envVarValue, clientToken, containerUri, discoveryUrl, allowedAudience, allowedClients, serverProtocol))
}
