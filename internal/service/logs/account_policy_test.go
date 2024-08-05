// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLogsAccountPolicy_basicSubscriptionFilter(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_account_policy.test"
	var accountPolicy types.AccountPolicy

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountPolicyConfig_basicSubscriptionFilter(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPolicyExists(ctx, resourceName, &accountPolicy),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					testAccCheckAccountHasSubscriptionFilterPolicy(resourceName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAccountPolicyImportStateIDFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLogsAccountPolicy_basicDataProtection(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_account_policy.test"
	var accountPolicy types.AccountPolicy

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountPolicyConfig_basicDataProtection(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPolicyExists(ctx, resourceName, &accountPolicy),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttr(resourceName, "policy_type", "DATA_PROTECTION_POLICY"),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "policy_document", `
{
	"Name": "Test",
	"Version": "2021-06-01",
	"Statement": [
		{
			"Sid": "Audit",
			"DataIdentifier": [
				"arn:aws:dataprotection::aws:data-identifier/EmailAddress"
			],
			"Operation": {
				"Audit": {
					"FindingsDestination": {}
				}
			}
		},
		{
			"Sid": "Redact",
			"DataIdentifier": [
				"arn:aws:dataprotection::aws:data-identifier/EmailAddress"
			],
			"Operation": {
				"Deidentify": {
					"MaskConfig": {}
				}
			}
		}
	]
}
`), //lintignore:AWSAT005
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAccountPolicyImportStateIDFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLogsAccountPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_account_policy.test"
	var accountPolicy types.AccountPolicy

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountPolicyConfig_basicDataProtection(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPolicyExists(ctx, resourceName, &accountPolicy),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflogs.ResourceAccountPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLogsAccountPolicy_selectionCriteria(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rSelectionCriteria := fmt.Sprintf("LogGroupName NOT IN [\"%s\"]", rName)
	resourceName := "aws_cloudwatch_log_account_policy.test"
	var accountPolicy types.AccountPolicy

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountPolicyConfig_selectionCriteria(rName, rSelectionCriteria),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPolicyExists(ctx, resourceName, &accountPolicy),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttr(resourceName, "selection_criteria", rSelectionCriteria),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAccountPolicyImportStateIDFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAccountPolicyExists(ctx context.Context, n string, v *types.AccountPolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LogsClient(ctx)

		output, err := tflogs.FindAccountPolicyByTwoPartKey(ctx, conn, types.PolicyType(rs.Primary.Attributes["policy_type"]), rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAccountPolicyImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		policyName := rs.Primary.ID
		policyType := rs.Primary.Attributes["policy_type"]
		stateID := fmt.Sprintf("%s:%s", policyName, policyType)

		return stateID, nil
	}
}

func testAccCheckAccountPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LogsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_log_account_policy" {
				continue
			}

			_, err := tflogs.FindAccountPolicyByTwoPartKey(ctx, conn, types.PolicyType(rs.Primary.Attributes["policy_type"]), rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Logs Resource Policy still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAccountHasSubscriptionFilterPolicy(resourceName string, rName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		expectedJSONTemplate := `{
			"DestinationArn": "arn:%s:lambda:%s:%s:function:%s",
			"FilterPattern" : " ",
			"Distribution" : "Random"
		  }`
		expectedJSON := fmt.Sprintf(expectedJSONTemplate, acctest.Partition(), acctest.Region(), acctest.AccountID(), rName)
		return acctest.CheckResourceAttrEquivalentJSON(resourceName, "policy_document", expectedJSON)(s)
	}
}

func testAccAccountPolicyConfig_lambdaBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_cloudwatch_log_group" "test" {
  name              = %[1]q
  retention_in_days = 1
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "lambda.${data.aws_partition.current.dns_suffix}"
    },
    "Effect": "Allow",
    "Sid": ""
  }]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role       = aws_iam_role.test.name
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  runtime       = "nodejs16.x"
  handler       = "exports.handler"
}

resource "aws_lambda_permission" "test" {
  statement_id  = "AllowExecutionFromCloudWatchLogs"
  action        = "lambda:*"
  function_name = aws_lambda_function.test.arn
  principal     = "logs.${data.aws_partition.current.dns_suffix}"
}
`, rName)
}

func testAccAccountPolicyConfig_basicSubscriptionFilter(rName string) string {
	return acctest.ConfigCompose(testAccAccountPolicyConfig_lambdaBase(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_account_policy" "test" {
  policy_name = %[1]q
  policy_type = "SUBSCRIPTION_FILTER_POLICY"

  policy_document = jsonencode({
    DestinationArn = "${aws_lambda_function.test.arn}"
    FilterPattern  = " "
    Distribution   = "Random"
  })
}
`, rName))
}

func testAccAccountPolicyConfig_selectionCriteria(rName, rSelectionCriteria string) string {
	return acctest.ConfigCompose(testAccAccountPolicyConfig_lambdaBase(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_account_policy" "test" {
  policy_name = %[1]q
  policy_type = "SUBSCRIPTION_FILTER_POLICY"

  policy_document = jsonencode({
    DestinationArn = "${aws_lambda_function.test.arn}"
    FilterPattern  = " "
    Distribution   = "Random"
  })

  selection_criteria = %[2]q
}
`, rName, rSelectionCriteria))
}

func testAccAccountPolicyConfig_basicDataProtection(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_cloudwatch_log_group" "test" {
  name              = %[1]q
  retention_in_days = 1
}

resource "aws_cloudwatch_log_account_policy" "test" {
  policy_name = %[1]q
  policy_type = "DATA_PROTECTION_POLICY"

  policy_document = jsonencode({
    Name    = "Test"
    Version = "2021-06-01"

    Statement = [
      {
        Sid            = "Audit"
        DataIdentifier = ["arn:${data.aws_partition.current.partition}:dataprotection::aws:data-identifier/EmailAddress"]
        Operation = {
          Audit = {
            FindingsDestination = {}
          }
        }
      },
      {
        Sid            = "Redact"
        DataIdentifier = ["arn:${data.aws_partition.current.partition}:dataprotection::aws:data-identifier/EmailAddress"]
        Operation = {
          Deidentify = {
            MaskConfig = {}
          }
        }
      }
    ]
  })
}
`, rName)
}
