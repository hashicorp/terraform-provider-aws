// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"

	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
)

// TIP: ==== ACCEPTANCE TESTS ====
// This is an example of a basic acceptance test. This should test as much of
// standard functionality of the resource as possible, and test importing, if
// applicable. We prefix its name with "TestAcc", the service, and the
// resource name.
//
// Acceptance test access AWS and cost money to run.
func TestAccLogsIndexPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var indexPolicy types.IndexPolicy
	resourceName := "aws_cloudwatch_log_index_policy.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIndexPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIndexPolicy_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIndexPolicyExists(ctx, resourceName, &indexPolicy),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrLogGroupName, logGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "fields", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccIndexPolicyImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLogsIndexPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var mf types.IndexPolicy
	resourceName := "aws_cloudwatch_log_index_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIndexPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIndexPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexPolicyExists(ctx, resourceName, &mf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflogs.ResourceIndexPolicy(), resourceName),
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

			_, err := tflogs.FindMetricFilterByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrLogGroupName], rs.Primary.ID)
			_, err := tflogs.

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Logs Metric Filter still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckIndexPolicyExists(ctx context.Context, logGroupName string, indexpolicy *cloudwatchlogs.DescribeIndexPolicyResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[logGroupName]
		if !ok {
			return create.Error(names.Logs, create.ErrActionCheckingExistence, tflogs.ResNameIndexPolicy, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Logs, create.ErrActionCheckingExistence, tflogs.ResNameIndexPolicy, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LogsClient(ctx)

		resp, err := tflogs.FindIndexPolicyByLogGroupName(ctx, conn, logGroupName)
		if err != nil {
			return create.Error(names.Logs, create.ErrActionCheckingExistence, tflogs.ResNameIndexPolicy, rs.Primary.ID, err)
		}

		*indexpolicy = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LogsClient(ctx)

	input := &cloudwatchlogs.ListIndexPolicysInput{}

	_, err := conn.ListIndexPolicys(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckIndexPolicyNotRecreated(before, after *cloudwatchlogs.DescribeIndexPolicyResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.IndexPolicyId), aws.ToString(after.IndexPolicyId); before != after {
			return create.Error(names.Logs, create.ErrActionCheckingNotRecreated, tflogs.ResNameIndexPolicy, aws.ToString(before.IndexPolicyId), errors.New("recreated"))
		}

		return nil
	}
}

func testAccIndexPolicyConfig_basic(rName, policyDocument string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_index_policy" "test" {
  log_group_name = %[1]q
  policyDocument = %[2]q
}
`, rName, policyDocument)
}
