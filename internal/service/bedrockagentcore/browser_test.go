// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCoreBrowser_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var browser bedrockagentcorecontrol.GetBrowserOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_browser.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckBrowser(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrowserDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBrowserConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserExists(ctx, t, resourceName, &browser),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("browser_arn"), tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`browser-custom/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("browser_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrNetworkConfiguration), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"network_mode":      tfknownvalue.StringExact(awstypes.BrowserNetworkModePublic),
						names.AttrVPCConfig: knownvalue.ListSizeExact(0),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
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

func TestAccBedrockAgentCoreBrowser_role_recording(t *testing.T) {
	ctx := acctest.Context(t)
	var browser bedrockagentcorecontrol.GetBrowserOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_browser.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckBrowser(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrowserDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBrowserConfig_role_recording(rName, acctest.RandomWithPrefix(t, "tf-test-bucket")),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserExists(ctx, t, resourceName, &browser),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
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

func TestAccBedrockAgentCoreBrowser_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var browser bedrockagentcorecontrol.GetBrowserOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_browser.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckBrowser(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrowserDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBrowserConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserExists(ctx, t, resourceName, &browser),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagentcore.ResourceBrowser, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
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
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_browser.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckBrowser(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrowserDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBrowserConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserExists(ctx, t, resourceName, &browser),
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
			{
				Config: testAccBrowserConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserExists(ctx, t, resourceName, &browser),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccBrowserConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserExists(ctx, t, resourceName, &browser),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreBrowser_networkConfiguration_vpc(t *testing.T) {
	acctest.Skip(t, "ENIs of type \"agentic_ai\" remain")
	ctx := acctest.Context(t)
	var browser bedrockagentcorecontrol.GetBrowserOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_browser.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckBrowser(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrowserDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBrowserConfig_networkConfiguration_vpc(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserExists(ctx, t, resourceName, &browser),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrNetworkConfiguration), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"network_mode": tfknownvalue.StringExact(awstypes.BrowserNetworkModeVpc),
						names.AttrVPCConfig: knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrSecurityGroups: knownvalue.SetSizeExact(1),
							names.AttrSubnets:        knownvalue.SetSizeExact(2),
						})}),
					})})),
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

func testAccCheckBrowserDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

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

func testAccCheckBrowserExists(ctx context.Context, t *testing.T, n string, v *bedrockagentcorecontrol.GetBrowserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindBrowserByID(ctx, conn, rs.Primary.Attributes["browser_id"])
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccPreCheckBrowser(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

	input := bedrockagentcorecontrol.ListBrowsersInput{}

	_, err := conn.ListBrowsers(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccBrowserConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_browser" "test" {
  name = %[1]q

  network_configuration {
    network_mode = "PUBLIC"
  }
}
`, rName)
}

func testAccBrowserConfig_role_recording(rName, bucketName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["bedrock-agentcore.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "s3:*",
    "Resource": "*"
  }
}
EOF
}

resource "aws_s3_bucket" "test" {
  bucket = %[2]q

  force_destroy = true
}

resource "aws_bedrockagentcore_browser" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  network_configuration {
    network_mode = "PUBLIC"
  }

  recording {
    s3_location {
      bucket = aws_s3_bucket.test.bucket
      prefix = "prefix1"
    }
    enabled = true
  }
}
`, rName, bucketName)
}

func testAccBrowserConfig_tags1(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_browser" "test" {
  name = %[1]q

  network_configuration {
    network_mode = "PUBLIC"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value)
}

func testAccBrowserConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_browser" "test" {
  name = %[1]q

  network_configuration {
    network_mode = "PUBLIC"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccBrowserConfig_networkConfiguration_vpc(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["bedrock-agentcore.amazonaws.com", "network.bedrock-agentcore.amazonaws.com"]
    }
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_bedrockagentcore_browser" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  network_configuration {
    network_mode = "VPC"

    vpc_config {
      security_groups = [aws_security_group.test.id]
      subnets         = aws_subnet.test[*].id
    }
  }
}
`, rName))
}
