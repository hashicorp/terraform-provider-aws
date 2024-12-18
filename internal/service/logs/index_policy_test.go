// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

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
	var logGroup types.LogGroup
	resourceName := "aws_cloudwatch_log_index_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyDocument := "{\"Fields\":[\"eventName\"]}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIndexPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIndexPolicyConfig_basic(rName, policyDocument),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, "aws_cloudwatch_log_group.test", &logGroup),
					testAccCheckIndexPolicyExists(ctx, t, resourceName, &indexPolicy),
					resource.TestCheckResourceAttr(resourceName, "policy_document", policyDocument),
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
	var ip types.IndexPolicy
	resourceName := "aws_cloudwatch_log_index_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyDocument := "{\"Fields\":[\"eventName\"]}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIndexPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIndexPolicyConfig_basic(rName, policyDocument),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexPolicyExists(ctx, t, resourceName, &ip),
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

			_, err := tflogs.FindIndexPolicyByLogGroupName(ctx, conn, rs.Primary.Attributes[names.AttrLogGroupName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Logs Index Policy still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckIndexPolicyExists(ctx context.Context, t *testing.T, resourceName string, indexPolicy *types.IndexPolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return create.Error(names.Logs, create.ErrActionCheckingExistence, tflogs.ResNameIndexPolicy, resourceName, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Logs, create.ErrActionCheckingExistence, tflogs.ResNameIndexPolicy, resourceName, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		resp, err := tflogs.FindIndexPolicyByLogGroupName(ctx, conn, rs.Primary.Attributes[names.AttrLogGroupName])
		if err != nil {
			return create.Error(names.Logs, create.ErrActionCheckingExistence, tflogs.ResNameIndexPolicy, rs.Primary.ID, err)
		}

		*indexPolicy = *resp
		return nil
	}
}

func testAccIndexPolicyImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes[names.AttrLogGroupName] + ":" + rs.Primary.Attributes[names.AttrName], nil
	}
}

func testAccIndexPolicyConfig_basic(rName, policyDocument string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_index_policy" "test" {
  log_group_name = aws_cloudwatch_log_group.test.name
  policy_document = %[2]q
}
`, rName, policyDocument)
}
