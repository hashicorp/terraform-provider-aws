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
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCoreAgentRuntimeEndpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var r1, r2 bedrockagentcorecontrol.GetAgentRuntimeEndpointOutput
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_agent_runtime_endpoint.test"
	imageUriV1 := os.Getenv("AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI")
	imageUriV2 := os.Getenv("AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V2_URI")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccAgentRuntimeEndpointPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentRuntimeEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentRuntimeConfig_iamRole(rName),
				Check: resource.ComposeTestCheckFunc(
					// Wait for IAM role and policy to propagate
					acctest.CheckSleep(t, 5*time.Second),
				),
			},
			{
				Config: testAccAgentRuntimeEndpointConfig_basic(rName, "test endpoint", imageUriV1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeEndpointExists(ctx, resourceName, &r1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test endpoint"),
					resource.TestCheckResourceAttr(resourceName, "agent_runtime_version", "1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "bedrock-agentcore", regexache.MustCompile(`runtime/.+/runtime-endpoint/.+$`)),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "agent_runtime_arn", "bedrock-agentcore", regexache.MustCompile(`runtime/.+$`)),
				),
			},
			{
				Config: testAccAgentRuntimeEndpointConfig_basic(rName, "updated endpoint", imageUriV2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeEndpointExists(ctx, resourceName, &r2),
					testAccCheckAgentRuntimeEndpointNotRecreated(&r1, &r2),
					resource.TestCheckResourceAttr(resourceName, "agent_runtime_version", "2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "updated endpoint"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccAgentRuntimeEndpointImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccBedrockAgentCoreAgentRuntimeEndpoint_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var agentruntimeendpoint bedrockagentcorecontrol.GetAgentRuntimeEndpointOutput
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_agent_runtime_endpoint.test"
	imageUri := os.Getenv("AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccAgentRuntimeEndpointPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentRuntimeEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentRuntimeConfig_iamRole(rName),
				Check: resource.ComposeTestCheckFunc(
					// Wait for IAM role and policy to propagate
					acctest.CheckSleep(t, 5*time.Second),
				),
			},
			{
				Config: testAccAgentRuntimeEndpointConfig_basic(rName, "test endpoint", imageUri),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeEndpointExists(ctx, resourceName, &agentruntimeendpoint),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbedrockagentcore.ResourceAgentRuntimeEndpoint, resourceName),
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

func testAccAgentRuntimeEndpointImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}
		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["agent_runtime_id"], rs.Primary.Attributes[names.AttrName]), nil
	}
}
func testAccCheckAgentRuntimeEndpointDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_agent_runtime_endpoint" {
				continue
			}

			endpointName := rs.Primary.Attributes[names.AttrName]

			_, err := tfbedrockagentcore.FindAgentRuntimeEndpointByRuntimeIDAndName(ctx, conn, rs.Primary.Attributes["agent_runtime_id"], endpointName)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.BedrockAgentCore, create.ErrActionCheckingDestroyed, tfbedrockagentcore.ResNameAgentRuntimeEndpoint, endpointName, err)
			}

			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingDestroyed, tfbedrockagentcore.ResNameAgentRuntimeEndpoint, endpointName, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckAgentRuntimeEndpointExists(ctx context.Context, name string, agentruntimeendpoint *bedrockagentcorecontrol.GetAgentRuntimeEndpointOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameAgentRuntimeEndpoint, name, errors.New("not found"))
		}

		endpointARN := rs.Primary.Attributes[names.AttrARN]
		name := rs.Primary.Attributes[names.AttrName]

		if endpointARN == "" {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameAgentRuntimeEndpoint, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindAgentRuntimeEndpointByRuntimeIDAndName(ctx, conn, rs.Primary.Attributes["agent_runtime_id"], name)
		if err != nil {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameAgentRuntimeEndpoint, name, err)
		}

		*agentruntimeendpoint = *resp

		return nil
	}
}

func testAccAgentRuntimeEndpointPreCheck(ctx context.Context, t *testing.T) {
	testAccAgentRuntimeImageVersionsPreCheck(t)

	conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

	input := bedrockagentcorecontrol.ListAgentRuntimeEndpointsInput{
		AgentRuntimeId: aws.String("non_existent_agent-abcde12345"),
	}

	_, err := conn.ListAgentRuntimeEndpoints(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckAgentRuntimeEndpointNotRecreated(before, after *bedrockagentcorecontrol.GetAgentRuntimeEndpointOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.AgentRuntimeEndpointArn), aws.ToString(after.AgentRuntimeEndpointArn); before != after {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingNotRecreated, tfbedrockagentcore.ResNameAgentRuntimeEndpoint, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccAgentRuntimeEndpointConfig_basic(rName, description, imageUri string) string {
	return acctest.ConfigCompose(testAccAgentRuntimeConfig_basic(rName, imageUri), fmt.Sprintf(`
resource "aws_bedrockagentcore_agent_runtime_endpoint" "test" {
  name                  = %[1]q
  description           = %[2]q
  agent_runtime_id      = aws_bedrockagentcore_agent_runtime.test.id
  agent_runtime_version = aws_bedrockagentcore_agent_runtime.test.version
}
`, rName, description))
}
