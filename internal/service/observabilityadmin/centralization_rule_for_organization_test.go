// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/observabilityadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/observabilityadmin/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfobservabilityadmin "github.com/hashicorp/terraform-provider-aws/internal/service/observabilityadmin"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccObservabilityAdminCentralizationRuleForOrganization_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

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
				Config: testAccCentralizationRuleForOrganizationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCentralizationRuleForOrganizationExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "rule_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "rule_arn"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

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
				Config: testAccCentralizationRuleForOrganizationConfig_basic(rName),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

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
				Config: testAccCentralizationRuleForOrganizationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCentralizationRuleForOrganizationExists(ctx, resourceName, &rule1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.source.0.regions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.source.0.source_logs_configuration.0.encrypted_log_group_strategy", "SKIP"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.source.0.source_logs_configuration.0.log_group_selection_criteria", "*"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.destination.0.destination_logs_configuration.#", "0"),
				),
			},
			{
				Config: testAccCentralizationRuleForOrganizationConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCentralizationRuleForOrganizationExists(ctx, resourceName, &rule2),
					// Test regions updated from 1 to 2
					resource.TestCheckResourceAttr(resourceName, "rule.0.source.0.regions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.source.0.regions.0", "ap-southeast-1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.source.0.regions.1", "us-east-1"),
					// Test source_logs_configuration updated from SKIP to ALLOW
					resource.TestCheckResourceAttr(resourceName, "rule.0.source.0.source_logs_configuration.0.encrypted_log_group_strategy", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.source.0.source_logs_configuration.0.log_group_selection_criteria", "*"),
					// Test destination_logs_configuration was added (went from non-existent to present)
					resource.TestCheckResourceAttr(resourceName, "rule.0.destination.0.destination_logs_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.destination.0.destination_logs_configuration.0.logs_encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.destination.0.destination_logs_configuration.0.logs_encryption_configuration.0.encryption_strategy", "AWS_OWNED"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.destination.0.destination_logs_configuration.0.backup_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.destination.0.destination_logs_configuration.0.backup_configuration.0.region", "us-west-1"),
				),
			},
			{
				Config: testAccCentralizationRuleForOrganizationConfig_updatedLogFilter(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCentralizationRuleForOrganizationExists(ctx, resourceName, &rule3),
					// Test log_group_selection_criteria updated from "*" to OAM filter
					resource.TestCheckResourceAttr(resourceName, "rule.0.source.0.source_logs_configuration.0.log_group_selection_criteria", "LogGroupName LIKE '/aws/lambda%'"),
					// Ensure other values remain the same
					resource.TestCheckResourceAttr(resourceName, "rule.0.source.0.regions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.source.0.source_logs_configuration.0.encrypted_log_group_strategy", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.destination.0.destination_logs_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.destination.0.destination_logs_configuration.0.logs_encryption_configuration.0.encryption_strategy", "AWS_OWNED"),
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

func TestAccObservabilityAdminCentralizationRuleForOrganization_import(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

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
				Config: testAccCentralizationRuleForOrganizationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCentralizationRuleForOrganizationExists(ctx, resourceName, &rule),
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

			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				continue
			}

			if err != nil {
				return create.Error(names.ObservabilityAdmin, create.ErrActionCheckingDestroyed, tfobservabilityadmin.ResNameCentralizationRuleForOrganization, rs.Primary.ID, err)
			}

			return create.Error(names.ObservabilityAdmin, create.ErrActionCheckingDestroyed, tfobservabilityadmin.ResNameCentralizationRuleForOrganization, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckCentralizationRuleForOrganizationExists(ctx context.Context, name string, rule *observabilityadmin.GetCentralizationRuleForOrganizationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ObservabilityAdmin, create.ErrActionCheckingExistence, tfobservabilityadmin.ResNameCentralizationRuleForOrganization, name, errors.New("not found"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ObservabilityAdminClient(ctx)

		ruleName := rs.Primary.Attributes["rule_name"]

		output, err := tfobservabilityadmin.FindCentralizationRuleForOrganizationByID(ctx, conn, ruleName)
		if err != nil {
			return create.Error(names.ObservabilityAdmin, create.ErrActionCheckingExistence, tfobservabilityadmin.ResNameCentralizationRuleForOrganization, ruleName, err)
		}

		*rule = *output

		return nil
	}
}

func testAccCentralizationRuleForOrganizationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["rule_name"], nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ObservabilityAdminClient(ctx)

	input := &observabilityadmin.ListCentralizationRulesForOrganizationInput{}
	_, err := conn.ListCentralizationRulesForOrganization(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCentralizationRuleForOrganizationConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_organizations_organization" "current" {}

resource "aws_observabilityadmin_centralization_rule_for_organization" "test" {
  rule_name = %[1]q

  rule {
    destination {
      region  = "eu-west-1"
      account = data.aws_caller_identity.current.account_id
    }

    source {
      regions = ["ap-southeast-1"]
      scope   = "OrganizationId = '${data.aws_organizations_organization.current.id}'"

      source_logs_configuration {
        encrypted_log_group_strategy  = "SKIP"
        log_group_selection_criteria = "*"
      }
    }
  }
}
`, rName)
}

func testAccCentralizationRuleForOrganizationConfig_updated(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_organizations_organization" "current" {}

resource "aws_observabilityadmin_centralization_rule_for_organization" "test" {
  rule_name = %[1]q

  rule {
    destination {
      region  = "eu-west-1"
      account = data.aws_caller_identity.current.account_id

      destination_logs_configuration {
        logs_encryption_configuration {
		  encryption_strategy = "AWS_OWNED"
        }

        backup_configuration {
          region = "us-west-1"
        }
      }
    }

    source {
      regions = ["ap-southeast-1", "us-east-1"]
      scope   = "OrganizationId = '${data.aws_organizations_organization.current.id}'"

      source_logs_configuration {
        encrypted_log_group_strategy  = "ALLOW"
        log_group_selection_criteria = "*"
      }
    }
  }
}
`, rName)
}
func testAccCentralizationRuleForOrganizationConfig_updatedLogFilter(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_organizations_organization" "current" {}

resource "aws_observabilityadmin_centralization_rule_for_organization" "test" {
  rule_name = %[1]q

  rule {
    destination {
      region  = "eu-west-1"
      account = data.aws_caller_identity.current.account_id

      destination_logs_configuration {
        logs_encryption_configuration {
		  encryption_strategy = "AWS_OWNED"
        }

        backup_configuration {
          region = "us-west-1"
        }
      }
    }

    source {
      regions = ["ap-southeast-1", "us-east-1"]
      scope   = "OrganizationId = '${data.aws_organizations_organization.current.id}'"

      source_logs_configuration {
        encrypted_log_group_strategy  = "ALLOW"
        log_group_selection_criteria = "LogGroupName LIKE '/aws/lambda%%'"
      }
    }
  }
}
`, rName)
}
