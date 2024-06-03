// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAutomationRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var automationRule types.AutomationRulesConfig
	resourceName := "aws_securityhub_automation_rule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutomationRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAutomationRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutomationRuleExists(ctx, resourceName, &automationRule),
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

func testAccAutomationRule_full(t *testing.T) {
	ctx := acctest.Context(t)
	var automationRule types.AutomationRulesConfig
	resourceName := "aws_securityhub_automation_rule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutomationRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAutomationRuleConfig_full(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutomationRuleExists(ctx, resourceName, &automationRule),
					resource.TestCheckResourceAttr(resourceName, "actions.0.finding_fields_update.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "actions.0.finding_fields_update.0.confidence", "20"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.finding_fields_update.0.criticality", "25"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.finding_fields_update.0.types.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "actions.0.finding_fields_update.0.note.0.text", "This is a critical resource. Please review ASAP."),
					resource.TestCheckResourceAttr(resourceName, "actions.0.finding_fields_update.0.note.0.updated_by", "sechub-automation"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.finding_fields_update.0.related_findings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "actions.0.finding_fields_update.0.related_findings.0.id", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "actions.0.finding_fields_update.0.related_findings.0.product_arn", "securityhub", regexache.MustCompile("product/aws/inspector")),
					resource.TestCheckResourceAttr(resourceName, "actions.0.finding_fields_update.0.severity.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "actions.0.finding_fields_update.0.severity.0.label", string(types.SeverityLabelCritical)),
					resource.TestCheckResourceAttr(resourceName, "actions.0.finding_fields_update.0.severity.0.product", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "actions.0.finding_fields_update.0.user_defined_fields.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "actions.0.finding_fields_update.0.workflow.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "actions.0.finding_fields_update.0.workflow.0.status", string(types.WorkflowStatusSuppressed)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAutomationRuleConfig_fullUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutomationRuleExists(ctx, resourceName, &automationRule),
					resource.TestCheckResourceAttr(resourceName, "actions.0.finding_fields_update.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "actions.0.finding_fields_update.0.confidence", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "actions.0.finding_fields_update.0.criticality", "15"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.finding_fields_update.0.types.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "actions.0.finding_fields_update.0.note.0.text", "This is a non-critical resource. Please review in due time."),
					resource.TestCheckResourceAttr(resourceName, "actions.0.finding_fields_update.0.note.0.updated_by", "sechub-automation"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.finding_fields_update.0.related_findings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "actions.0.finding_fields_update.0.related_findings.0.id", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "actions.0.finding_fields_update.0.related_findings.0.product_arn", "securityhub", regexache.MustCompile("product/aws/inspector")),
					resource.TestCheckResourceAttr(resourceName, "actions.0.finding_fields_update.0.severity.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "actions.0.finding_fields_update.0.severity.0.label", string(types.SeverityLabelLow)),
					resource.TestCheckResourceAttr(resourceName, "actions.0.finding_fields_update.0.severity.0.product", "15.5"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.finding_fields_update.0.user_defined_fields.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "actions.0.finding_fields_update.0.workflow.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "actions.0.finding_fields_update.0.workflow.0.status", string(types.WorkflowStatusNew)),
				),
			},
		},
	})
}

func testAccAutomationRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var automationRule types.AutomationRulesConfig
	resourceName := "aws_securityhub_automation_rule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutomationRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAutomationRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutomationRuleExists(ctx, resourceName, &automationRule),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfsecurityhub.ResourceAutomationRule, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAutomationRule_stringFilters(t *testing.T) {
	ctx := acctest.Context(t)
	var automationRule types.AutomationRulesConfig
	resourceName := "aws_securityhub_automation_rule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutomationRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAutomationRuleConfig_stringFilters(rName, string(types.StringFilterComparisonEquals), "1234567890"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutomationRuleExists(ctx, resourceName, &automationRule),
					resource.TestCheckResourceAttr(resourceName, "criteria.0.aws_account_id.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "criteria.0.aws_account_id.0.comparison", string(types.StringFilterComparisonEquals)),
					resource.TestCheckResourceAttr(resourceName, "criteria.0.aws_account_id.0.value", "1234567890"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAutomationRuleConfig_stringFilters(rName, string(types.StringFilterComparisonContains), "0987654321"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutomationRuleExists(ctx, resourceName, &automationRule),
					resource.TestCheckResourceAttr(resourceName, "criteria.0.aws_account_id.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "criteria.0.aws_account_id.0.comparison", string(types.StringFilterComparisonContains)),
					resource.TestCheckResourceAttr(resourceName, "criteria.0.aws_account_id.0.value", "0987654321"),
				),
			},
		},
	})
}

func testAccAutomationRule_numberFilters(t *testing.T) {
	ctx := acctest.Context(t)
	var automationRule types.AutomationRulesConfig
	resourceName := "aws_securityhub_automation_rule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutomationRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAutomationRuleConfig_numberFilters(rName, "eq = 5"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutomationRuleExists(ctx, resourceName, &automationRule),
					resource.TestCheckResourceAttr(resourceName, "criteria.0.confidence.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "criteria.0.confidence.0.eq", "5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAutomationRuleConfig_numberFilters(rName, "lte = 50"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutomationRuleExists(ctx, resourceName, &automationRule),
					resource.TestCheckResourceAttr(resourceName, "criteria.0.confidence.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "criteria.0.confidence.0.lte", "50"),
				),
			},
		},
	})
}

func testAccAutomationRule_dateFilters(t *testing.T) {
	ctx := acctest.Context(t)
	var automationRule types.AutomationRulesConfig
	resourceName := "aws_securityhub_automation_rule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	endDate := time.Now().Add(5 * time.Minute).Format(time.RFC3339)
	startDate := time.Now().Format(time.RFC3339)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutomationRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAutomationRuleConfig_dateFiltersAbsoluteRange(rName, startDate, endDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutomationRuleExists(ctx, resourceName, &automationRule),
					resource.TestCheckResourceAttr(resourceName, "criteria.0.created_at.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "criteria.0.created_at.0.end", endDate),
					resource.TestCheckResourceAttr(resourceName, "criteria.0.created_at.0.start", startDate),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAutomationRuleConfig_dateFiltersRelativeRange(rName, string(types.DateRangeUnitDays), 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutomationRuleExists(ctx, resourceName, &automationRule),
					resource.TestCheckResourceAttr(resourceName, "criteria.0.created_at.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "criteria.0.created_at.0.date_range.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "criteria.0.created_at.0.date_range.0.unit", string(types.DateRangeUnitDays)),
					resource.TestCheckResourceAttr(resourceName, "criteria.0.created_at.0.date_range.0.value", acctest.Ct10),
				),
			},
		},
	})
}

func testAccAutomationRule_mapFilters(t *testing.T) {
	ctx := acctest.Context(t)
	var automationRule types.AutomationRulesConfig
	resourceName := "aws_securityhub_automation_rule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutomationRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAutomationRuleConfig_mapFilters(rName, string(types.MapFilterComparisonEquals), acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutomationRuleExists(ctx, resourceName, &automationRule),
					resource.TestCheckResourceAttr(resourceName, "criteria.0.resource_details_other.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "criteria.0.resource_details_other.0.comparison", string(types.MapFilterComparisonEquals)),
					resource.TestCheckResourceAttr(resourceName, "criteria.0.resource_details_other.0.key", acctest.CtKey1),
					resource.TestCheckResourceAttr(resourceName, "criteria.0.resource_details_other.0.value", acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAutomationRuleConfig_mapFilters(rName, string(types.MapFilterComparisonContains), acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutomationRuleExists(ctx, resourceName, &automationRule),
					resource.TestCheckResourceAttr(resourceName, "criteria.0.resource_details_other.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "criteria.0.resource_details_other.0.comparison", string(types.MapFilterComparisonContains)),
					resource.TestCheckResourceAttr(resourceName, "criteria.0.resource_details_other.0.key", acctest.CtKey2),
					resource.TestCheckResourceAttr(resourceName, "criteria.0.resource_details_other.0.value", acctest.CtValue2),
				),
			},
		},
	})
}

func testAccAutomationRule_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var automationRule types.AutomationRulesConfig
	resourceName := "aws_securityhub_automation_rule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutomationRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAutomationRuleConfig_tags(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutomationRuleExists(ctx, resourceName, &automationRule),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAutomationRuleConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutomationRuleExists(ctx, resourceName, &automationRule),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAutomationRuleConfig_tags(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutomationRuleExists(ctx, resourceName, &automationRule),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
		},
	})
}

func testAccCheckAutomationRuleExists(ctx context.Context, n string, v *types.AutomationRulesConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubClient(ctx)

		output, err := tfsecurityhub.FindAutomationRuleByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAutomationRuleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securityhub_automation_rule" {
				continue
			}

			_, err := tfsecurityhub.FindAutomationRuleByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Security Hub Automation Rule %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAutomationRuleConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_automation_rule" "test" {
  description = "test description"
  rule_name   = %[1]q
  rule_order  = 1

  actions {
    finding_fields_update {
      severity {
        label   = "LOW"
        product = "0.0"
      }

      types = ["Software and Configuration Checks/Industry and Regulatory Standards"]

      user_defined_fields = {
        key = "value"
      }
    }
    type = "FINDING_FIELDS_UPDATE"
  }

  criteria {
    aws_account_id {
      comparison = "EQUALS"
      value      = "1234567890"
    }
  }

  depends_on = [aws_securityhub_account.test]
}
`, rName)
}

func testAccAutomationRuleConfig_full(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_automation_rule" "test" {
  description = "test description"
  rule_name   = %[1]q
  rule_order  = 1

  actions {
    finding_fields_update {
      confidence  = 20
      criticality = 25
      types       = ["Software and Configuration Checks/Industry and Regulatory Standards"]

      note {
        text       = "This is a critical resource. Please review ASAP."
        updated_by = "sechub-automation"
      }
      related_findings {
        id          = %[1]q
        product_arn = "arn:${data.aws_partition.current.partition}:securityhub:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:product/aws/inspector"
      }
      severity {
        label   = "CRITICAL"
        product = "0"
      }
      user_defined_fields = {
        key = "value"
      }
      workflow {
        status = "SUPPRESSED"
      }
    }
    type = "FINDING_FIELDS_UPDATE"
  }

  criteria {
    aws_account_id {
      comparison = "EQUALS"
      value      = "1234567890"
    }
    created_at {
      date_range {
        unit  = "DAYS"
        value = 10
      }
    }
  }

  depends_on = [aws_securityhub_account.test]
}
`, rName)
}

func testAccAutomationRuleConfig_fullUpdated(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_automation_rule" "test" {
  description = "test description"
  rule_name   = %[1]q
  rule_order  = 1

  actions {
    finding_fields_update {
      confidence  = 10
      criticality = 15
      types       = ["Software and Configuration Checks/Industry and Regulatory Standards"]

      note {
        text       = "This is a non-critical resource. Please review in due time."
        updated_by = "sechub-automation"
      }
      related_findings {
        id          = %[1]q
        product_arn = "arn:${data.aws_partition.current.partition}:securityhub:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:product/aws/inspector"
      }
      severity {
        label   = "LOW"
        product = "15.5"
      }
      user_defined_fields = {
        key = "value"
      }
      workflow {
        status = "NEW"
      }
    }
    type = "FINDING_FIELDS_UPDATE"
  }

  criteria {
    aws_account_id {
      comparison = "EQUALS"
      value      = "1234567890"
    }
    created_at {
      date_range {
        unit  = "DAYS"
        value = 10
      }
    }
  }

  depends_on = [aws_securityhub_account.test]
}
`, rName)
}

func testAccAutomationRuleConfig_stringFilters(rName, comparison, value string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_automation_rule" "test" {
  description = "test description"
  rule_name   = %[1]q
  rule_order  = 1

  actions {
    finding_fields_update {
      user_defined_fields = {
        key = "value"
      }
    }
    type = "FINDING_FIELDS_UPDATE"
  }

  criteria {
    aws_account_id {
      comparison = %[2]q
      value      = %[3]q
    }
  }

  depends_on = [aws_securityhub_account.test]
}
`, rName, comparison, value)
}

func testAccAutomationRuleConfig_numberFilters(rName, value string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_automation_rule" "test" {
  description = "test description"
  rule_name   = %[1]q
  rule_order  = 1

  actions {
    finding_fields_update {
      user_defined_fields = {
        key = "value"
      }
    }
    type = "FINDING_FIELDS_UPDATE"
  }

  criteria {
    confidence {
      %[2]s
    }
  }

  depends_on = [aws_securityhub_account.test]
}
`, rName, value)
}

func testAccAutomationRuleConfig_dateFiltersAbsoluteRange(rName, start, end string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_automation_rule" "test" {
  description = "test description"
  rule_name   = %[1]q
  rule_order  = 1

  actions {
    finding_fields_update {
      user_defined_fields = {
        key = "value"
      }
    }
    type = "FINDING_FIELDS_UPDATE"
  }

  criteria {
    created_at {
      end   = %[3]q
      start = %[2]q
    }
  }

  depends_on = [aws_securityhub_account.test]
}
`, rName, start, end)
}

func testAccAutomationRuleConfig_dateFiltersRelativeRange(rName, unit string, value int) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_automation_rule" "test" {
  description = "test description"
  rule_name   = %[1]q
  rule_order  = 1

  actions {
    finding_fields_update {
      user_defined_fields = {
        key = "value"
      }
    }
    type = "FINDING_FIELDS_UPDATE"
  }

  criteria {
    created_at {
      date_range {
        unit  = %[2]q
        value = %[3]d
      }
    }
  }

  depends_on = [aws_securityhub_account.test]
}
`, rName, unit, value)
}

func testAccAutomationRuleConfig_mapFilters(rName, comparison, key, value string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_automation_rule" "test" {
  description = "test description"
  rule_name   = %[1]q
  rule_order  = 1

  actions {
    finding_fields_update {
      user_defined_fields = {
        key = "value"
      }
    }
    type = "FINDING_FIELDS_UPDATE"
  }

  criteria {
    resource_details_other {
      comparison = %[2]q
      key        = %[3]q
      value      = %[4]q
    }
  }

  depends_on = [aws_securityhub_account.test]
}
`, rName, comparison, key, value)
}

func testAccAutomationRuleConfig_tags(rName, key, value string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_automation_rule" "test" {
  description = "test description"
  rule_name   = %[1]q
  rule_order  = 1

  actions {
    finding_fields_update {
      user_defined_fields = {
        key = "value"
      }
    }
    type = "FINDING_FIELDS_UPDATE"
  }

  criteria {
    aws_account_id {
      comparison = "EQUALS"
      value      = "1234567890"
    }
  }
  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_securityhub_account.test]
}
`, rName, key, value)
}

func testAccAutomationRuleConfig_tags2(rName, key, value, key2, value2 string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_automation_rule" "test" {
  description = "test description"
  rule_name   = %[1]q
  rule_order  = 1

  actions {
    finding_fields_update {
      user_defined_fields = {
        key = "value"
      }
    }
    type = "FINDING_FIELDS_UPDATE"
  }

  criteria {
    aws_account_id {
      comparison = "EQUALS"
      value      = "1234567890"
    }
  }
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_securityhub_account.test]
}
`, rName, key, value, key2, value2)
}
