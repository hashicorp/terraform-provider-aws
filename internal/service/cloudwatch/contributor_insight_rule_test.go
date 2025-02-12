// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudwatch_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"

	// TIP: You will often need to import the package that this test file lives
	// in. Since it is in the "test" context, it must import the package to use
	// any normal context constants, variables, or functions.
	tfcloudwatch "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatch"
)

func TestAccCloudWatchContributorInsightRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var contributorinsightrule cloudwatch.DescribeContributorInsightRuleResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_contributor_insight_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContributorInsightRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContributorInsightRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContributorInsightRuleExists(ctx, resourceName, &contributorinsightrule),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "cloudwatch", regexache.MustCompile(`contributorinsightrule:+.`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccCloudWatchContributorInsightRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var contributorinsightrule cloudwatch.DescribeContributorInsightRuleResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_contributor_insight_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContributorInsightRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContributorInsightRuleConfig_basic(rName, testAccContributorInsightRuleVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContributorInsightRuleExists(ctx, resourceName, &contributorinsightrule),
					// TIP: The Plugin-Framework disappears helper is similar to the Plugin-SDK version,
					// but expects a new resource factory function as the third argument. To expose this
					// private function to the testing package, you may need to add a line like the following
					// to exports_test.go:
					//
					//   var ResourceContributorInsightRule = newResourceContributorInsightRule
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

			input := &cloudwatch.DescribeContributorInsightRuleInput{
				ContributorInsightRuleId: aws.String(rs.Primary.ID),
			}
			_, err := conn.DescribeContributorInsightRule(ctx, &cloudwatch.DescribeContributorInsightRuleInput{
				ContributorInsightRuleId: aws.String(rs.Primary.ID),
			})
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.CloudWatch, create.ErrActionCheckingDestroyed, tfcloudwatch.ResNameContributorInsightRule, rs.Primary.ID, err)
			}

			return create.Error(names.CloudWatch, create.ErrActionCheckingDestroyed, tfcloudwatch.ResNameContributorInsightRule, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckContributorInsightRuleExists(ctx context.Context, name string, contributorinsightrule *cloudwatch.DescribeContributorInsightRuleResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.CloudWatch, create.ErrActionCheckingExistence, tfcloudwatch.ResNameContributorInsightRule, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.CloudWatch, create.ErrActionCheckingExistence, tfcloudwatch.ResNameContributorInsightRule, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchClient(ctx)
		resp, err := conn.DescribeContributorInsightRule(ctx, &cloudwatch.DescribeContributorInsightRuleInput{
			ContributorInsightRuleId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.CloudWatch, create.ErrActionCheckingExistence, tfcloudwatch.ResNameContributorInsightRule, rs.Primary.ID, err)
		}

		*contributorinsightrule = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchClient(ctx)

	input := &cloudwatch.ListContributorInsightRulesInput{}
	_, err := conn.ListContributorInsightRules(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckContributorInsightRuleNotRecreated(before, after *cloudwatch.DescribeContributorInsightRuleResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.ContributorInsightRuleId), aws.ToString(after.ContributorInsightRuleId); before != after {
			return create.Error(names.CloudWatch, create.ErrActionCheckingNotRecreated, tfcloudwatch.ResNameContributorInsightRule, aws.ToString(before.ContributorInsightRuleId), errors.New("recreated"))
		}

		return nil
	}
}

func testAccContributorInsightRuleConfig_basic(rName, state string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_contributor_insight_rule" "test" {
  rule_name             = %[1]q
  rule_state             = %[2]q
  rule_definition  = "{\"Schema\":{\"Name\":\"CloudWatchLogRule\",\"Version\":1},\"AggregateOn\":\"Count\",\"Contribution\":{\"Filters\":[{\"In\":[\"some-keyword\"],\"Match\":\"$.message\"}],\"Keys\":[\"$.country\"]},\"LogFormat\":\"JSON\",\"LogGroupNames\":[\"/aws/lambda/api-prod\"]}"
}
`, rName, state)
}
