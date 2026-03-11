// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package wafv2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfwafv2 "github.com/hashicorp/terraform-provider-aws/internal/service/wafv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWAFV2RuleGroupPermissionPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_wafv2_rule_group_permission_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAlternateAccount(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckRuleGroupPermissionPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupPermissionPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupPermissionPolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, "aws_wafv2_rule_group.test", names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPolicy},
			},
		},
	})
}

func TestAccWAFV2RuleGroupPermissionPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_wafv2_rule_group_permission_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAlternateAccount(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckRuleGroupPermissionPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupPermissionPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupPermissionPolicyExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfwafv2.ResourceRuleGroupPermissionPolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFV2RuleGroupPermissionPolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_wafv2_rule_group_permission_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAlternateAccount(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckRuleGroupPermissionPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupPermissionPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupPermissionPolicyExists(ctx, t, resourceName),
				),
			},
			{
				Config: testAccRuleGroupPermissionPolicyConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupPermissionPolicyExists(ctx, t, resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPolicy},
			},
		},
	})
}

func testAccCheckRuleGroupPermissionPolicyExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).WAFV2Client(ctx)

		_, err := tfwafv2.FindPermissionPolicyByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckRuleGroupPermissionPolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).WAFV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_wafv2_rule_group_permission_policy" {
				continue
			}

			_, err := tfwafv2.FindPermissionPolicyByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WAFv2 Rule Group Permission Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccRuleGroupPermissionPolicyConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
data "aws_caller_identity" "target" {
  provider = "awsalternate"
}

resource "aws_wafv2_rule_group" "test" {
  name     = %[1]q
  scope    = "REGIONAL"
  capacity = 2

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
}

resource "aws_wafv2_rule_group_permission_policy" "test" {
  resource_arn = aws_wafv2_rule_group.test.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.target.account_id}:root"
      }
      Action = [
        "wafv2:CreateWebACL",
        "wafv2:UpdateWebACL",
        "wafv2:PutFirewallManagerRuleGroups",
        "wafv2:GetRuleGroup",
      ]
    }]
  })
}

data "aws_partition" "current" {}
`, rName))
}

func testAccRuleGroupPermissionPolicyConfig_updated(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
data "aws_caller_identity" "target" {
  provider = "awsalternate"
}

resource "aws_wafv2_rule_group" "test" {
  name     = %[1]q
  scope    = "REGIONAL"
  capacity = 2

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
}

resource "aws_wafv2_rule_group_permission_policy" "test" {
  resource_arn = aws_wafv2_rule_group.test.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.target.account_id}:root"
      }
      Action = [
        "wafv2:CreateWebACL",
        "wafv2:UpdateWebACL",
        "wafv2:PutFirewallManagerRuleGroups",
      ]
    }]
  })
}

data "aws_partition" "current" {}
`, rName))
}
