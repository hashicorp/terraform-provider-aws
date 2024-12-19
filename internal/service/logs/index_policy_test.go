// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLogsIndexPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var indexPolicy cloudwatchlogs.DescribeIndexPoliciesOutput
	logGroupName := "/aws/testacc/index-policy-" + sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyDocument := `{Fields:[\"eventName\"]}`
	resourceName := "aws_cloudwatch_log_index_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LogsServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIndexPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIndexPolicyConfig_basic(logGroupName, policyDocument),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexPolicyExists(ctx, resourceName, &indexPolicy),
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

func TestAccLogsIndexPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var indexPolicy cloudwatchlogs.DescribeIndexPoliciesOutput
	logGroupName := "/aws/testacc/index-policy-" + sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyDocument := `{Fields:[\"eventName\"]}`
	resourceName := "aws_cloudwatch_log_index_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LogsServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIndexPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIndexPolicyConfig_basic(logGroupName, policyDocument),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexPolicyExists(ctx, resourceName, &indexPolicy),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tflogs.ResourceIndexPolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckIndexPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LogsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_log_index_policy" {
				continue
			}

			_, err := tflogs.FindIndexPolicyByLogGroupName(ctx, conn, rs.Primary.Attributes["log_group_name"])
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Logs, create.ErrActionCheckingDestroyed, tflogs.ResNameIndexPolicy, rs.Primary.ID, err)
			}

			return create.Error(names.Logs, create.ErrActionCheckingDestroyed, tflogs.ResNameIndexPolicy, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckIndexPolicyExists(ctx context.Context, name string, indexPolicy *cloudwatchlogs.DescribeIndexPoliciesOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Logs, create.ErrActionCheckingExistence, tflogs.ResNameIndexPolicy, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Logs, create.ErrActionCheckingExistence, tflogs.ResNameIndexPolicy, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LogsClient(ctx)

		resp, err := tflogs.FindIndexPolicyByLogGroupName(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.Logs, create.ErrActionCheckingExistence, tflogs.ResNameIndexPolicy, rs.Primary.ID, err)
		}

		indexPolicy.IndexPolicies = append(indexPolicy.IndexPolicies, *resp)

		return nil
	}
}

func testAccIndexPolicyConfig_basic(logGroupName string, policyDocument string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_index_policy" "test" {
  log_group_name  = aws_cloudwatch_log_group.test.name
  policy_document = jsonencode(%[2]q)
}
`, logGroupName, policyDocument)
}
