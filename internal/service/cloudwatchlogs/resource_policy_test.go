package cloudwatchlogs_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudwatchlogs "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatchlogs"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)





func TestAccCloudWatchLogsResourcePolicy_basic(t *testing.T) {
	name := sdkacctest.RandString(5)
	resourceName := "aws_cloudwatch_log_resource_policy.test"
	var resourcePolicy cloudwatchlogs.ResourcePolicy

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCloudWatchLogResourcePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckResourcePolicyResourceBasic1Config(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogResourcePolicy(resourceName, &resourcePolicy),
					resource.TestCheckResourceAttr(resourceName, "policy_name", name),
					resource.TestCheckResourceAttr(resourceName, "policy_document", fmt.Sprintf("{\"Version\":\"2012-10-17\",\"Statement\":[{\"Sid\":\"\",\"Effect\":\"Allow\",\"Principal\":{\"Service\":\"rds.%s\"},\"Action\":[\"logs:PutLogEvents\",\"logs:CreateLogStream\"],\"Resource\":\"arn:%s:logs:*:*:log-group:/aws/rds/*\"}]}", acctest.PartitionDNSSuffix(), acctest.Partition())),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCheckResourcePolicyResourceBasic2Config(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogResourcePolicy(resourceName, &resourcePolicy),
					resource.TestCheckResourceAttr(resourceName, "policy_name", name),
					resource.TestCheckResourceAttr(resourceName, "policy_document", fmt.Sprintf("{\"Version\":\"2012-10-17\",\"Statement\":[{\"Sid\":\"\",\"Effect\":\"Allow\",\"Principal\":{\"Service\":\"rds.%s\"},\"Action\":[\"logs:PutLogEvents\",\"logs:CreateLogStream\"],\"Resource\":\"arn:%s:logs:*:*:log-group:/aws/rds/example.com\"}]}", acctest.PartitionDNSSuffix(), acctest.Partition())),
				),
			},
		},
	})
}

func testAccCheckCloudWatchLogResourcePolicy(pr string, resourcePolicy *cloudwatchlogs.ResourcePolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchLogsConn
		rs, ok := s.RootModule().Resources[pr]
		if !ok {
			return fmt.Errorf("Not found: %s", pr)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		policy, exists, err := tfcloudwatchlogs.LookupResourcePolicy(conn, rs.Primary.ID, nil)
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

func testAccCheckCloudWatchLogResourcePolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchLogsConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_log_resource_policy" {
			continue
		}

		_, exists, err := tfcloudwatchlogs.LookupResourcePolicy(conn, rs.Primary.ID, nil)

		if err != nil {
			return fmt.Errorf("error reading CloudWatch Log Resource Policy (%s): %w", rs.Primary.ID, err)
		}

		if exists {
			return fmt.Errorf("Resource policy exists: %q", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckResourcePolicyResourceBasic1Config(name string) string {
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
  policy_name     = "%s"
  policy_document = data.aws_iam_policy_document.test.json
}
`, name)
}

func testAccCheckResourcePolicyResourceBasic2Config(name string) string {
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
  policy_name     = "%s"
  policy_document = data.aws_iam_policy_document.test.json
}
`, name)
}
