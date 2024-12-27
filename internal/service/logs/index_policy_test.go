// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLogsIndexPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	logGroupName := "/aws/testacc/index-policy-" + sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyDocument := `{"Fields":["eventName"]}`
	resourceName := "aws_cloudwatch_log_index_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIndexPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIndexPolicyConfig_basic(logGroupName, policyDocument),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexPolicyExists(ctx, resourceName),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccIndexPolicyImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrLogGroupName,
			},
		},
	})
}

func TestAccLogsIndexPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	logGroupName := "/aws/testacc/index-policy-" + sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyDocument := `{"Fields":["eventName"]}`
	resourceName := "aws_cloudwatch_log_index_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIndexPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIndexPolicyConfig_basic(logGroupName, policyDocument),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexPolicyExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tflogs.ResourceIndexPolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLogsIndexPolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	logGroupName := "/aws/testacc/index-policy-" + sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyDocument1 := `{"Fields":["eventName"]}`
	policyDocument2 := `{"Fields": ["eventName", "requestId"]}`
	resourceName := "aws_cloudwatch_log_index_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIndexPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIndexPolicyConfig_basic(logGroupName, policyDocument1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_document", policyDocument1),
				),
			},
			{
				Config: testAccIndexPolicyConfig_basic(logGroupName, policyDocument2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_document", policyDocument2),
				),
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

func testAccCheckIndexPolicyExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LogsClient(ctx)

		_, err := tflogs.FindIndexPolicyByLogGroupName(ctx, conn, rs.Primary.Attributes[names.AttrLogGroupName])

		return err
	}
}

func testAccIndexPolicyImportStateIDFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}

		return rs.Primary.Attributes[names.AttrLogGroupName], nil
	}
}

func testAccIndexPolicyConfig_basic(logGroupName string, policyDocument string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_index_policy" "test" {
  log_group_name  = aws_cloudwatch_log_group.test.name
  policy_document = %[2]q
}
`, logGroupName, policyDocument)
}
