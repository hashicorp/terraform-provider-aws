// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package devopsagent_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfdevopsagent "github.com/hashicorp/terraform-provider-aws/internal/service/devopsagent"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDevOpsAgentAgentSpace_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_devopsagent_agent_space.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsAgent),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentSpaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentSpaceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentSpaceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "agent_space_id"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "aidevops", regexache.MustCompile(`agentspace/.+`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "agent_space_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "agent_space_id",
			},
		},
	})
}

func TestAccDevOpsAgentAgentSpace_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_devopsagent_agent_space.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsAgent),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentSpaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentSpaceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentSpaceExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfdevopsagent.ResourceAgentSpace, resourceName),
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

func TestAccDevOpsAgentAgentSpace_description(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_devopsagent_agent_space.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsAgent),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentSpaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentSpaceConfig_description(rName, "initial description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentSpaceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "initial description"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "agent_space_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "agent_space_id",
			},
			{
				Config: testAccAgentSpaceConfig_description(rName, "updated description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentSpaceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "updated description"),
				),
			},
		},
	})
}

func TestAccDevOpsAgentAgentSpace_kmsKeyARN(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_devopsagent_agent_space.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsAgent),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentSpaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentSpaceConfig_kmsKeyARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentSpaceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, "aws_kms_key.test", names.AttrARN),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "agent_space_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "agent_space_id",
			},
		},
	})
}

func TestAccDevOpsAgentAgentSpace_locale(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_devopsagent_agent_space.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsAgent),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentSpaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentSpaceConfig_locale(rName, "en"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentSpaceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "locale", "en"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "agent_space_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "agent_space_id",
			},
			{
				Config: testAccAgentSpaceConfig_locale(rName, "fr"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentSpaceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "locale", "fr"),
				),
			},
		},
	})
}

func testAccCheckAgentSpaceDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DevOpsAgentClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_devopsagent_agent_space" {
				continue
			}

			_, err := tfdevopsagent.FindAgentSpaceByID(ctx, conn, rs.Primary.Attributes["agent_space_id"])
			if err == nil {
				return fmt.Errorf("DevOps Agent Space %s still exists", rs.Primary.Attributes["agent_space_id"])
			}
		}

		return nil
	}
}

func testAccCheckAgentSpaceExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DevOpsAgentClient(ctx)

		_, err := tfdevopsagent.FindAgentSpaceByID(ctx, conn, rs.Primary.Attributes["agent_space_id"])
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccAgentSpaceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_devopsagent_agent_space" "test" {
  name = %[1]q
}
`, rName)
}

func testAccAgentSpaceConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_devopsagent_agent_space" "test" {
  name        = %[1]q
  description = %[2]q
}
`, rName, description)
}

func testAccAgentSpaceConfig_kmsKeyARN(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    sid       = "EnableIAMPermissions"
    actions   = ["kms:*"]
    resources = ["*"]

    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }

  statement {
    sid       = "AllowDevOpsAgentDescribeKey"
    actions   = ["kms:DescribeKey"]
    resources = ["*"]

    principals {
      type        = "Service"
      identifiers = ["aidevops.${data.aws_partition.current.dns_suffix}"]
    }
  }

  statement {
    sid = "AllowDevOpsAgentEncryption"
    actions = [
      "kms:Decrypt",
      "kms:Encrypt",
      "kms:GenerateDataKey*",
      "kms:ReEncrypt*",
    ]
    resources = ["*"]

    principals {
      type        = "Service"
      identifiers = ["aidevops.${data.aws_partition.current.dns_suffix}"]
    }

    condition {
      test     = "ArnLike"
      variable = "aws:SourceArn"
      values   = ["arn:${data.aws_partition.current.partition}:aidevops:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:agentspace/*"]
    }

    condition {
      test     = "StringLike"
      variable = "kms:EncryptionContext:aws-crypto-ec:aws:aidevops:arn"
      values   = ["arn:${data.aws_partition.current.partition}:aidevops:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:agentspace/*"]
    }
  }
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  policy                  = data.aws_iam_policy_document.test.json
}

resource "aws_devopsagent_agent_space" "test" {
  name        = %[1]q
  kms_key_arn = aws_kms_key.test.arn
}
`, rName)
}

func testAccAgentSpaceConfig_locale(rName, locale string) string {
	return fmt.Sprintf(`
resource "aws_devopsagent_agent_space" "test" {
  name   = %[1]q
  locale = %[2]q
}
`, rName, locale)
}
