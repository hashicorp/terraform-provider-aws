package logs_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
)

func TestAccLogsResourcePolicy_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_resource_policy.test"
	var resourcePolicy cloudwatchlogs.ResourcePolicy

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourcePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckResourcePolicyResourceBasic1Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicy(resourceName, &resourcePolicy),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttr(resourceName, "policy_document", fmt.Sprintf("{\"Statement\":[{\"Action\":[\"logs:PutLogEvents\",\"logs:CreateLogStream\"],\"Effect\":\"Allow\",\"Principal\":{\"Service\":\"rds.%s\"},\"Resource\":\"arn:%s:logs:*:*:log-group:/aws/rds/*\",\"Sid\":\"\"}],\"Version\":\"2012-10-17\"}", acctest.PartitionDNSSuffix(), acctest.Partition())),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCheckResourcePolicyResourceBasic2Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicy(resourceName, &resourcePolicy),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttr(resourceName, "policy_document", fmt.Sprintf("{\"Statement\":[{\"Action\":[\"logs:PutLogEvents\",\"logs:CreateLogStream\"],\"Effect\":\"Allow\",\"Principal\":{\"Service\":\"rds.%s\"},\"Resource\":\"arn:%s:logs:*:*:log-group:/aws/rds/example.com\",\"Sid\":\"\"}],\"Version\":\"2012-10-17\"}", acctest.PartitionDNSSuffix(), acctest.Partition())),
				),
			},
		},
	})
}

func TestAccLogsResourcePolicy_ignoreEquivalent(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_resource_policy.test"
	var resourcePolicy cloudwatchlogs.ResourcePolicy

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourcePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckResourcePolicyResourceOrderConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicy(resourceName, &resourcePolicy),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttr(resourceName, "policy_document", fmt.Sprintf("{\"Statement\":[{\"Action\":[\"logs:CreateLogStream\",\"logs:PutLogEvents\"],\"Effect\":\"Allow\",\"Principal\":{\"Service\":[\"rds.%s\"]},\"Resource\":[\"arn:%s:logs:*:*:log-group:/aws/rds/example.com\"]}],\"Version\":\"2012-10-17\"}", acctest.PartitionDNSSuffix(), acctest.Partition())),
				),
			},
			{
				Config: testAccCheckResourcePolicyResourceNewOrderConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicy(resourceName, &resourcePolicy),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttr(resourceName, "policy_document", fmt.Sprintf("{\"Statement\":[{\"Action\":[\"logs:CreateLogStream\",\"logs:PutLogEvents\"],\"Effect\":\"Allow\",\"Principal\":{\"Service\":[\"rds.%s\"]},\"Resource\":[\"arn:%s:logs:*:*:log-group:/aws/rds/example.com\"]}],\"Version\":\"2012-10-17\"}", acctest.PartitionDNSSuffix(), acctest.Partition())),
				),
			},
		},
	})
}

func testAccCheckResourcePolicy(pr string, resourcePolicy *cloudwatchlogs.ResourcePolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LogsConn
		rs, ok := s.RootModule().Resources[pr]
		if !ok {
			return fmt.Errorf("Not found: %s", pr)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		policy, exists, err := tflogs.LookupResourcePolicy(conn, rs.Primary.ID, nil)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Resource policy does not exist: %q", rs.Primary.ID)
		}

		*resourcePolicy = *policy

		return nil
	}
}

func testAccCheckResourcePolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LogsConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_log_resource_policy" {
			continue
		}

		_, exists, err := tflogs.LookupResourcePolicy(conn, rs.Primary.ID, nil)

		if err != nil {
			return fmt.Errorf("error reading CloudWatch Log Resource Policy (%s): %w", rs.Primary.ID, err)
		}

		if exists {
			return fmt.Errorf("Resource policy exists: %q", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckResourcePolicyResourceBasic1Config(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = ["arn:${data.aws_partition.current.partition}:logs:*:*:log-group:/aws/rds/*"]

    principals {
      identifiers = ["rds.${data.aws_partition.current.dns_suffix}"]
      type        = "Service"
    }
  }
}

resource "aws_cloudwatch_log_resource_policy" "test" {
  policy_name     = %[1]q
  policy_document = data.aws_iam_policy_document.test.json
}
`, rName)
}

func testAccCheckResourcePolicyResourceBasic2Config(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = ["arn:${data.aws_partition.current.partition}:logs:*:*:log-group:/aws/rds/example.com"]

    principals {
      identifiers = ["rds.${data.aws_partition.current.dns_suffix}"]
      type        = "Service"
    }
  }
}

resource "aws_cloudwatch_log_resource_policy" "test" {
  policy_name     = %[1]q
  policy_document = data.aws_iam_policy_document.test.json
}
`, rName)
}

func testAccCheckResourcePolicyResourceOrderConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_cloudwatch_log_resource_policy" "test" {
  policy_name = %[1]q
  policy_document = jsonencode({
    Statement = [{
      Action = [
        "logs:CreateLogStream",
        "logs:PutLogEvents",
      ]
      Effect = "Allow"
      Resource = [
        "arn:${data.aws_partition.current.partition}:logs:*:*:log-group:/aws/rds/example.com",
      ]
      Principal = {
        Service = [
          "rds.${data.aws_partition.current.dns_suffix}",
        ]
      }
    }]
    Version = "2012-10-17"
  })
}
`, rName)
}

func testAccCheckResourcePolicyResourceNewOrderConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_cloudwatch_log_resource_policy" "test" {
  policy_name = %[1]q
  policy_document = jsonencode({
    Statement = [{
      Action = [
        "logs:PutLogEvents",
        "logs:CreateLogStream",
      ]
      Effect = "Allow"
      Resource = [
        "arn:${data.aws_partition.current.partition}:logs:*:*:log-group:/aws/rds/example.com",
      ]
      Principal = {
        Service = [
          "rds.${data.aws_partition.current.dns_suffix}",
        ]
      }
    }]
    Version = "2012-10-17"
  })
}
`, rName)
}
