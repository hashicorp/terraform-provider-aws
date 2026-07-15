// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatch_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudwatch "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatch"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func checkContributorInsightRuleARN(ruleName string) knownvalue.Check {
	return tfknownvalue.RegionalARNExact("cloudwatch", "insight-rule/"+ruleName)
}

func TestAccCloudWatchContributorInsightRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.InsightRule
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_contributor_insight_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContributorInsightRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/ContributorInsightRule/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContributorInsightRuleExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrResourceARN), checkContributorInsightRuleARN(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rule_definition"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rule_name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rule_state"), knownvalue.StringExact("ENABLED")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/ContributorInsightRule/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "rule_name"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "rule_name",
				ImportStateVerifyIgnore: []string{
					"rule_definition",
				},
			},
		},
	})
}

func TestAccCloudWatchContributorInsightRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.InsightRule
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_contributor_insight_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContributorInsightRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/ContributorInsightRule/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContributorInsightRuleExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcloudwatch.ResourceContributorInsightRule, resourceName),
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

func TestAccCloudWatchContributorInsightRule_updateRuleState(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.InsightRule
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_contributor_insight_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContributorInsightRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/ContributorInsightRule/rule_state/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"rule_state":    config.StringVariable("DISABLED"),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContributorInsightRuleExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rule_state"), knownvalue.StringExact("DISABLED")),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/ContributorInsightRule/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "rule_name"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "rule_name",
				ImportStateVerifyIgnore: []string{
					"rule_definition",
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/ContributorInsightRule/rule_state/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"rule_state":    config.StringVariable("ENABLED"),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContributorInsightRuleExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rule_state"), knownvalue.StringExact("ENABLED")),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/ContributorInsightRule/rule_state/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"rule_state":    config.StringVariable("DISABLED"),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContributorInsightRuleExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rule_state"), knownvalue.StringExact("DISABLED")),
				},
			},
		},
	})
}

func TestAccCloudWatchContributorInsightRule_updateDefinition(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.InsightRule
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_contributor_insight_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContributorInsightRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/ContributorInsightRule/rule_definition/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"aggregate_on":  config.StringVariable("Count"),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContributorInsightRuleExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/ContributorInsightRule/rule_definition/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"aggregate_on":  config.StringVariable("Sum"),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContributorInsightRuleExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func testAccCheckContributorInsightRuleDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CloudWatchClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_contributor_insight_rule" {
				continue
			}

			_, err := tfcloudwatch.FindInsightRuleByName(ctx, conn, rs.Primary.Attributes["rule_name"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Contributor Insight Rule still exists: %s", rs.Primary.Attributes["rule_name"])
		}

		return nil
	}
}

func testAccCheckContributorInsightRuleExists(ctx context.Context, t *testing.T, n string, v *types.InsightRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CloudWatchClient(ctx)

		output, err := tfcloudwatch.FindInsightRuleByName(ctx, conn, rs.Primary.Attributes["rule_name"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}
