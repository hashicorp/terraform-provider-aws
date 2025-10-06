// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCoreBrowser_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var browser bedrockagentcorecontrol.GetBrowserOutput
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_browser.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckBrowser(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrowserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBrowserConfig_basic(rName, "test description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserExists(ctx, resourceName, &browser),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.network_mode", "PUBLIC"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "browser_arn", "bedrock-agentcore", regexache.MustCompile(`browser-custom/.+$`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "browser_id"),
				ImportStateVerifyIdentifierAttribute: "browser_id",
			},
		},
	})
}

func TestAccBedrockAgentCoreBrowser_role_recording(t *testing.T) {
	ctx := acctest.Context(t)
	var browser bedrockagentcorecontrol.GetBrowserOutput
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_browser.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckBrowser(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrowserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBrowserConfig_role_recording(rName, "bucket.test.com", "the_prefix"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserExists(ctx, resourceName, &browser),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrExecutionRoleARN),
					resource.TestCheckResourceAttr(resourceName, "recording.s3_location.bucket", "bucket.test.com"),
					resource.TestCheckResourceAttr(resourceName, "recording.s3_location.prefix", "the_prefix"),
					resource.TestCheckResourceAttr(resourceName, "recording.enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "browser_id"),
				ImportStateVerifyIdentifierAttribute: "browser_id",
			},
		},
	})
}

func TestAccBedrockAgentCoreBrowser_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var browser bedrockagentcorecontrol.GetBrowserOutput
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_browser.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckBrowser(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrowserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBrowserConfig_basic(rName, "test description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserExists(ctx, resourceName, &browser),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbedrockagentcore.ResourceBrowser, resourceName),
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

func TestAccBedrockAgentCoreBrowser_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var browser bedrockagentcorecontrol.GetBrowserOutput
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_browser.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckBrowser(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrowserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBrowserConfig_tags1(rName, "test description", acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserExists(ctx, resourceName, &browser),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "browser_id"),
				ImportStateVerifyIdentifierAttribute: "browser_id",
			},
		},
	})
}

func TestAccBedrockAgentCoreBrowser_networkConfiguration_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	var browser bedrockagentcorecontrol.GetBrowserOutput
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_browser.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckBrowser(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrowserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBrowserConfig_networkConfiguration_vpc(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserExists(ctx, resourceName, &browser),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.network_mode", "VPC"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.network_mode_config.0.security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.network_mode_config.0.subnets.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "network_configuration.0.network_mode_config.0.security_groups.*", "aws_security_group.test", names.AttrID),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "browser_id"),
				ImportStateVerifyIdentifierAttribute: "browser_id",
			},
		},
	})
}

func testAccCheckBrowserDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_browser" {
				continue
			}

			_, err := tfbedrockagentcore.FindBrowserByID(ctx, conn, rs.Primary.Attributes["browser_id"])
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent Core Browser %s still exists", rs.Primary.Attributes["browser_id"])
		}

		return nil
	}
}

func testAccCheckBrowserExists(ctx context.Context, n string, v *bedrockagentcorecontrol.GetBrowserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindBrowserByID(ctx, conn, rs.Primary.Attributes["browser_id"])
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccPreCheckBrowser(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

	input := bedrockagentcorecontrol.ListBrowsersInput{}

	_, err := conn.ListBrowsers(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccBrowserConfig_basic(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_browser" "test" {
  name        = %[1]q
  description = %[2]q

  network_configuration {
    network_mode = "PUBLIC"
  }
}
`, rName, description)
}

func testAccBrowserConfig_IAMRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test_assume.json
}

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
`, rName)
}

func testAccBrowserConfig_role_recording(rName, bucket, prefix string) string {
	return acctest.ConfigCompose(testAccBrowserConfig_IAMRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_browser" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  network_configuration {
    network_mode = "PUBLIC"
  }

  recording = {
    s3_location = {
      bucket = %[2]q
      prefix = %[3]q
    }
    enabled = true
  }
}
`, rName, bucket, prefix))
}

func testAccBrowserConfig_tags1(rName, description, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_browser" "test" {
  name        = %[1]q
  description = %[2]q

  network_configuration {
    network_mode = "PUBLIC"
  }

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, description, tag1Key, tag1Value)
}

func testAccBrowserConfig_networkConfiguration_vpc(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = aws_vpc.test.id
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = %[1]q
}

resource "aws_bedrockagentcore_browser" "test" {
  name        = %[1]q
  description = "test VPC configuration"

  network_configuration {
    network_mode = "VPC"

    network_mode_config {
      security_groups = [aws_security_group.test.id]
      subnets         = aws_subnet.test[*].id
    }
  }
}
`, rName))
}
