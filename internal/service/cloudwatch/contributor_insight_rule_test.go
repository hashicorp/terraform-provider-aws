// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudwatch_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudwatch "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatch"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudWatchContributorInsightRule_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var v types.InsightRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_contributor_insight_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContributorInsightRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContributorInsightRuleConfig_basic(rName, "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContributorInsightRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "rule_name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule_state", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "rule_definition", "{\"Schema\":{\"Name\":\"CloudWatchLogRule\",\"Version\":1},\"AggregateOn\":\"Count\",\"Contribution\":{\"Filters\":[{\"In\":[\"some-keyword\"],\"Match\":\"$.message\"}],\"Keys\":[\"$.country\"]},\"LogFormat\":\"JSON\",\"LogGroupNames\":[\"/aws/lambda/api-prod\"]}"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccContributorInsightRuleImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "rule_name",
				ImportStateVerifyIgnore: []string{
					"rule_definition",
					"rule_state",
				},
			},
		},
	})
}

func TestAccCloudWatchContributorInsightRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var v types.InsightRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_contributor_insight_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContributorInsightRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContributorInsightRuleConfig_basic(rName, "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContributorInsightRuleExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfcloudwatch.ResourceContributorInsightRule, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckContributorInsightRuleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_contributor_insight_rule" {
				continue
			}

			_, err := tfcloudwatch.FindContributorInsightRuleByName(ctx, conn, rs.Primary.Attributes["rule_name"])

			if tfresource.NotFound(err) {
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

func testAccCheckContributorInsightRuleExists(ctx context.Context, name string, contributorinsightrule *types.InsightRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Contributor Insight Rule Rule Name is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchClient(ctx)

		output, err := tfcloudwatch.FindContributorInsightRuleByName(ctx, conn, rs.Primary.Attributes["rule_name"])

		if err != nil {
			return err
		}

		*contributorinsightrule = *output

		return nil
	}
}

func testAccContributorInsightRuleImportStateIDFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}

		return rs.Primary.Attributes["rule_name"], nil
	}
}

func testAccContributorInsightRuleConfig_basic(rName, state string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_contributor_insight_rule" "test" {
  rule_name       = %[1]q
  rule_state      = %[2]q
  rule_definition = "{\"Schema\":{\"Name\":\"CloudWatchLogRule\",\"Version\":1},\"AggregateOn\":\"Count\",\"Contribution\":{\"Filters\":[{\"In\":[\"some-keyword\"],\"Match\":\"$.message\"}],\"Keys\":[\"$.country\"]},\"LogFormat\":\"JSON\",\"LogGroupNames\":[\"/aws/lambda/api-prod\"]}"
}
`, rName, state)
}
