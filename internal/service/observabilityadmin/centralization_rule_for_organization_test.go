// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/observabilityadmin"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfobservabilityadmin "github.com/hashicorp/terraform-provider-aws/internal/service/observabilityadmin"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccObservabilityAdminCentralizationRuleForOrganization_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var rule observabilityadmin.GetCentralizationRuleForOrganizationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_observabilityadmin_centralization_rule_for_organization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCentralizationRuleForOrganizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCentralizationRuleForOrganizationConfig_basic(rName, endpoints.EuWest1RegionID, endpoints.ApSoutheast1RegionID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCentralizationRuleForOrganizationExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "rule_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "rule_arn"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.source.0.regions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.destination.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "rule.0.destination.0.account"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "rule_name"),
				ImportStateVerifyIdentifierAttribute: "rule_name",
			},
		},
	})
}

func TestAccObservabilityAdminCentralizationRuleForOrganization_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var rule observabilityadmin.GetCentralizationRuleForOrganizationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_observabilityadmin_centralization_rule_for_organization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCentralizationRuleForOrganizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCentralizationRuleForOrganizationConfig_basic(rName, endpoints.EuWest1RegionID, endpoints.ApSoutheast1RegionID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCentralizationRuleForOrganizationExists(ctx, resourceName, &rule),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfobservabilityadmin.ResourceCentralizationRuleForOrganization, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccObservabilityAdminCentralizationRuleForOrganization_update(t *testing.T) {
	ctx := acctest.Context(t)
	var rule1, rule2, rule3 observabilityadmin.GetCentralizationRuleForOrganizationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_observabilityadmin_centralization_rule_for_organization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCentralizationRuleForOrganizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCentralizationRuleForOrganizationConfig_basic(rName, endpoints.EuWest1RegionID, endpoints.ApSoutheast1RegionID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCentralizationRuleForOrganizationExists(ctx, resourceName, &rule1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.source.0.regions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.source.0.source_logs_configuration.0.encrypted_log_group_strategy", "SKIP"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.source.0.source_logs_configuration.0.log_group_selection_criteria", "*"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.destination.0.destination_logs_configuration.#", "0"),
					// Test tags
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "test"),
				),
			},
			{
				Config: testAccCentralizationRuleForOrganizationConfig_updated(rName, endpoints.EuWest1RegionID, endpoints.UsWest1RegionID, endpoints.ApSoutheast1RegionID, endpoints.UsEast1RegionID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCentralizationRuleForOrganizationExists(ctx, resourceName, &rule2),
					// Test regions updated from 1 to 2
					resource.TestCheckResourceAttr(resourceName, "rule.0.source.0.regions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.source.0.regions.0", endpoints.ApSoutheast1RegionID),
					resource.TestCheckResourceAttr(resourceName, "rule.0.source.0.regions.1", endpoints.UsEast1RegionID),
					// Test source_logs_configuration updated from SKIP to ALLOW
					resource.TestCheckResourceAttr(resourceName, "rule.0.source.0.source_logs_configuration.0.encrypted_log_group_strategy", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.source.0.source_logs_configuration.0.log_group_selection_criteria", "*"),
					// Test destination_logs_configuration was added (went from non-existent to present)
					resource.TestCheckResourceAttr(resourceName, "rule.0.destination.0.destination_logs_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.destination.0.destination_logs_configuration.0.logs_encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.destination.0.destination_logs_configuration.0.logs_encryption_configuration.0.encryption_strategy", "AWS_OWNED"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.destination.0.destination_logs_configuration.0.backup_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.destination.0.destination_logs_configuration.0.backup_configuration.0.region", endpoints.UsWest1RegionID),
					// Test updated tags
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "production"),
					resource.TestCheckResourceAttr(resourceName, "tags.Team", "observability"),
				),
			},
			{
				Config: testAccCentralizationRuleForOrganizationConfig_updatedLogFilter(rName, endpoints.EuWest1RegionID, endpoints.UsWest1RegionID, endpoints.ApSoutheast1RegionID, endpoints.UsEast1RegionID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCentralizationRuleForOrganizationExists(ctx, resourceName, &rule3),
					// Test log_group_selection_criteria updated from "*" to OAM filter
					resource.TestCheckResourceAttr(resourceName, "rule.0.source.0.source_logs_configuration.0.log_group_selection_criteria", "LogGroupName LIKE '/aws/lambda%'"),
					// Ensure other values remain the same
					resource.TestCheckResourceAttr(resourceName, "rule.0.source.0.regions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.source.0.source_logs_configuration.0.encrypted_log_group_strategy", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.destination.0.destination_logs_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.destination.0.destination_logs_configuration.0.logs_encryption_configuration.0.encryption_strategy", "AWS_OWNED"),
					// Test final tags with additional Filter tag
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "production"),
					resource.TestCheckResourceAttr(resourceName, "tags.Team", "observability"),
					resource.TestCheckResourceAttr(resourceName, "tags.Filter", "lambda-logs"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "rule_name"),
				ImportStateVerifyIdentifierAttribute: "rule_name",
			},
		},
	})
}

func testAccCheckCentralizationRuleForOrganizationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ObservabilityAdminClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_observabilityadmin_centralization_rule_for_organization" {
				continue
			}

			_, err := tfobservabilityadmin.FindCentralizationRuleForOrganizationByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Observability Admin Centralization Rule For Organization %s still exists", rs.Primary.Attributes["rule_name"])
		}

		return nil
	}
}

func testAccCheckCentralizationRuleForOrganizationExists(ctx context.Context, n string, v *observabilityadmin.GetCentralizationRuleForOrganizationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ObservabilityAdminClient(ctx)

		output, err := tfobservabilityadmin.FindCentralizationRuleForOrganizationByID(ctx, conn, rs.Primary.Attributes["rule_name"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ObservabilityAdminClient(ctx)

	input := observabilityadmin.ListCentralizationRulesForOrganizationInput{}
	_, err := conn.ListCentralizationRulesForOrganization(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCentralizationRuleForOrganizationConfig_basic(rName, dstRegion, srcRegion string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_organizations_organization" "current" {}

resource "aws_observabilityadmin_centralization_rule_for_organization" "test" {
  rule_name = %[1]q

  rule {
    destination {
      region  = %[2]q
      account = data.aws_caller_identity.current.account_id
    }

    source {
      regions = [%[3]q]
      scope   = "OrganizationId = '${data.aws_organizations_organization.current.id}'"

      source_logs_configuration {
        encrypted_log_group_strategy = "SKIP"
        log_group_selection_criteria = "*"
      }
    }
  }

  tags = {
    Name        = %[1]q
    Environment = "test"
  }
}
`, rName, dstRegion, srcRegion)
}

func testAccCentralizationRuleForOrganizationConfig_updated(rName, dstRegion, backupRegion, srcRegion1, srcRegion2 string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_organizations_organization" "current" {}

resource "aws_observabilityadmin_centralization_rule_for_organization" "test" {
  rule_name = %[1]q

  rule {
    destination {
      region  = %[2]q
      account = data.aws_caller_identity.current.account_id

      destination_logs_configuration {
        logs_encryption_configuration {
          encryption_strategy = "AWS_OWNED"
        }

        backup_configuration {
          region = %[3]q
        }
      }
    }

    source {
      regions = [%[4]q, %[5]q]
      scope   = "OrganizationId = '${data.aws_organizations_organization.current.id}'"

      source_logs_configuration {
        encrypted_log_group_strategy = "ALLOW"
        log_group_selection_criteria = "*"
      }
    }
  }

  tags = {
    Name        = %[1]q
    Environment = "production"
    Team        = "observability"
  }
}
`, rName, dstRegion, backupRegion, srcRegion1, srcRegion2)
}
func testAccCentralizationRuleForOrganizationConfig_updatedLogFilter(rName, dstRegion, backupRegion, srcRegion1, srcRegion2 string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_organizations_organization" "current" {}

resource "aws_observabilityadmin_centralization_rule_for_organization" "test" {
  rule_name = %[1]q

  rule {
    destination {
      region  = %[2]q
      account = data.aws_caller_identity.current.account_id

      destination_logs_configuration {
        logs_encryption_configuration {
          encryption_strategy = "AWS_OWNED"
        }

        backup_configuration {
          region = %[3]q
        }
      }
    }

    source {
      regions = [%[4]q, %[5]q]
      scope   = "OrganizationId = '${data.aws_organizations_organization.current.id}'"

      source_logs_configuration {
        encrypted_log_group_strategy = "ALLOW"
        log_group_selection_criteria = "LogGroupName LIKE '/aws/lambda%%'"
      }
    }
  }

  tags = {
    Name        = %[1]q
    Environment = "production"
    Team        = "observability"
    Filter      = "lambda-logs"
  }
}
`, rName, dstRegion, backupRegion, srcRegion1, srcRegion2)
}
