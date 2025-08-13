// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

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

func TestAccBedrockAgentCoreGateway_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var gateway bedrockagentcorecontrol.GetGatewayOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "authorizer_type", "CUSTOM_JWT"),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", "MCP"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
					resource.TestCheckResourceAttr(resourceName, "authorizer_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.discovery_url", "https://accounts.google.com/.well-known/openid-configuration"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.allowed_audience.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.allowed_audience.*", "test1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.allowed_audience.*", "test2"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "bedrock-agentcore", regexache.MustCompile(`gateway/.+$`)),
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

func TestAccBedrockAgentCoreGateway_full(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var gateway bedrockagentcorecontrol.GetGatewayOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_full(rName, "full configuration test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "full configuration test"),
					resource.TestCheckResourceAttr(resourceName, "exception_level", "DEBUG"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrKMSKeyARN),
					resource.TestCheckResourceAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.discovery_url", "https://accounts.google.com/.well-known/openid-configuration"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.allowed_audience.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.allowed_audience.*", "test1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.allowed_audience.*", "test2"),
					resource.TestCheckResourceAttr(resourceName, "protocol_configuration.0.mcp.0.instructions", "Full test instructions"),
					resource.TestCheckResourceAttr(resourceName, "protocol_configuration.0.mcp.0.search_type", "SEMANTIC"),
					resource.TestCheckResourceAttr(resourceName, "protocol_configuration.0.mcp.0.supported_versions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "protocol_configuration.0.mcp.0.supported_versions.*", "2025-03-26"),
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

func TestAccBedrockAgentCoreGateway_description(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var gateway bedrockagentcorecontrol.GetGatewayOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_description(rName, "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Initial description"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGatewayConfig_description(rName, "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated description"),
				),
			},
		},
	})
}

func TestAccBedrockAgentCoreGateway_authorizerConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var gateway1, gateway2 bedrockagentcorecontrol.GetGatewayOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_authorizerConfiguration(rName, "https://accounts.google.com/.well-known/openid-configuration", "weather", "sports"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, resourceName, &gateway1),
					resource.TestCheckResourceAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.discovery_url", "https://accounts.google.com/.well-known/openid-configuration"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.allowed_audience.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.allowed_audience.*", "weather"),
					resource.TestCheckTypeSetElemAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.allowed_audience.*", "sports"),
				),
			},
			{
				Config: testAccGatewayConfig_authorizerConfiguration(rName, "https://login.microsoftonline.com/common/v2.0/.well-known/openid-configuration", "finance", "technology"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, resourceName, &gateway2),
					testAccCheckGatewayNotRecreated(&gateway1, &gateway2),
					resource.TestCheckResourceAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.discovery_url", "https://login.microsoftonline.com/common/v2.0/.well-known/openid-configuration"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.allowed_audience.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.allowed_audience.*", "finance"),
					resource.TestCheckTypeSetElemAttr(resourceName, "authorizer_configuration.0.custom_jwt_authorizer.0.allowed_audience.*", "technology"),
				),
			},
		},
	})
}

func TestAccBedrockAgentCoreGateway_protocolConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var gateway bedrockagentcorecontrol.GetGatewayOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_protocolConfiguration(rName, "Initial instructions", "SEMANTIC", "2025-03-26", "2025-03-26"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "protocol_configuration.0.mcp.0.instructions", "Initial instructions"),
					resource.TestCheckResourceAttr(resourceName, "protocol_configuration.0.mcp.0.search_type", "SEMANTIC"),
					resource.TestCheckResourceAttr(resourceName, "protocol_configuration.0.mcp.0.supported_versions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "protocol_configuration.0.mcp.0.supported_versions.*", "2025-03-26"),
				),
			},
			{
				Config: testAccGatewayConfig_protocolConfiguration(rName, "Updated instructions", "SEMANTIC", "2025-03-26", "2025-03-26"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "protocol_configuration.0.mcp.0.instructions", "Updated instructions"),
					resource.TestCheckResourceAttr(resourceName, "protocol_configuration.0.mcp.0.search_type", "SEMANTIC"),
					resource.TestCheckResourceAttr(resourceName, "protocol_configuration.0.mcp.0.supported_versions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "protocol_configuration.0.mcp.0.supported_versions.*", "2025-03-26"),
				),
			},
		},
	})
}

func TestAccBedrockAgentCoreGateway_exceptionLevel(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var gateway bedrockagentcorecontrol.GetGatewayOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, resourceName, &gateway),
					resource.TestCheckNoResourceAttr(resourceName, "exception_level"),
				),
			},
			{
				Config: testAccGatewayConfig_exceptionLevel(rName, "DEBUG"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "exception_level", "DEBUG"),
				),
			},
		},
	})
}

func TestAccBedrockAgentCoreGateway_KMSKeyARN(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var gateway bedrockagentcorecontrol.GetGatewayOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_KMSKeyARN(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, resourceName, &gateway),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrKMSKeyARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, "aws_kms_key.test", names.AttrARN),
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

func TestAccBedrockAgentCoreGateway_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var gateway bedrockagentcorecontrol.GetGatewayOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, resourceName, &gateway),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbedrockagentcore.ResourceGateway, resourceName),
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

func testAccCheckGatewayDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_gateway" {
				continue
			}

			_, err := tfbedrockagentcore.FindGatewayByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.BedrockAgentCore, create.ErrActionCheckingDestroyed, tfbedrockagentcore.ResNameGateway, rs.Primary.ID, err)
			}

			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingDestroyed, tfbedrockagentcore.ResNameGateway, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckGatewayExists(ctx context.Context, name string, gateway *bedrockagentcorecontrol.GetGatewayOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameGateway, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameGateway, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindGatewayByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameGateway, rs.Primary.ID, err)
		}

		*gateway = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

	var input bedrockagentcorecontrol.ListGatewaysInput

	_, err := conn.ListGateways(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckGatewayNotRecreated(before, after *bedrockagentcorecontrol.GetGatewayOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.GatewayId), aws.ToString(after.GatewayId); before != after {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingNotRecreated, tfbedrockagentcore.ResNameGateway, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccGatewayConfig_iamRole(rName string) string {
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

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test_assume.json
}
`, rName)
}

func testAccGatewayConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = "https://accounts.google.com/.well-known/openid-configuration"
      allowed_audience = ["test1", "test2"]
    }
  }
}
`, rName))
}

func testAccGatewayConfig_full(rName, description string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "Test key for %[1]s"
  deletion_window_in_days = 7
}

resource "aws_bedrockagentcore_gateway" "test" {
  name            = %[1]q
  description     = %[2]q
  role_arn        = aws_iam_role.test.arn
  exception_level = "DEBUG"
  kms_key_arn     = aws_kms_key.test.arn

  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = "https://accounts.google.com/.well-known/openid-configuration"
      allowed_audience = ["test1", "test2"]
    }
  }

  protocol_configuration {
    mcp {
      instructions       = "Full test instructions"
      search_type        = "SEMANTIC"
      supported_versions = ["2025-03-26"]
    }
  }
}
`, rName, description))
}

func testAccGatewayConfig_description(rName, description string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway" "test" {
  name        = %[1]q
  description = %[2]q
  role_arn    = aws_iam_role.test.arn

  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = "https://accounts.google.com/.well-known/openid-configuration"
      allowed_audience = ["test1", "test2"]
    }
  }
}
`, rName, description))
}

func testAccGatewayConfig_authorizerConfiguration(rName, discoveryUrl, audience1, audience2 string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = %[2]q
      allowed_audience = [%[3]q, %[4]q]
    }
  }
}
`, rName, discoveryUrl, audience1, audience2))
}

func testAccGatewayConfig_protocolConfiguration(rName, instructions, searchType, version1, version2 string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = "https://accounts.google.com/.well-known/openid-configuration"
      allowed_audience = ["test1", "test2"]
    }
  }

  protocol_configuration {
    mcp {
      instructions       = %[2]q
      search_type        = %[3]q
      supported_versions = [%[4]q, %[5]q]
    }
  }
}
`, rName, instructions, searchType, version1, version2))
}

func testAccGatewayConfig_exceptionLevel(rName, exceptionLevel string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway" "test" {
  name            = %[1]q
  role_arn        = aws_iam_role.test.arn
  exception_level = %[2]q

  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = "https://accounts.google.com/.well-known/openid-configuration"
      allowed_audience = ["test1", "test2"]
    }
  }
}
`, rName, exceptionLevel))
}

func testAccGatewayConfig_KMSKeyARN(rName string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "Test key for %[1]s"
  deletion_window_in_days = 7
}

resource "aws_bedrockagentcore_gateway" "test" {
  name        = %[1]q
  role_arn    = aws_iam_role.test.arn
  kms_key_arn = aws_kms_key.test.arn

  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = "https://accounts.google.com/.well-known/openid-configuration"
      allowed_audience = ["test1", "test2"]
    }
  }
}
`, rName))
}
