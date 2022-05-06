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

func TestAccLogsDestinationPolicy_basic(t *testing.T) {
	var destination cloudwatchlogs.Destination
	resourceName := "aws_cloudwatch_log_destination_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestinationPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDestinationPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDestinationPolicyExists(resourceName, &destination),
					resource.TestCheckResourceAttrPair(resourceName, "destination_name", "aws_cloudwatch_log_destination.test", "name"),
					resource.TestCheckResourceAttrSet(resourceName, "access_policy"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDestinationPolicyForceUpdateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDestinationPolicyExists(resourceName, &destination),
					resource.TestCheckResourceAttrPair(resourceName, "destination_name", "aws_cloudwatch_log_destination.test", "name"),
					resource.TestCheckResourceAttrSet(resourceName, "access_policy"),
				),
			},
		},
	})
}

func testAccCheckDestinationPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LogsConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_log_destination_policy" {
			continue
		}
		_, exists, err := tflogs.LookupDestination(conn, rs.Primary.ID, nil)

		if err != nil {
			return fmt.Errorf("error reading CloudWatch Log Destination (%s): %w", rs.Primary.ID, err)
		}

		if exists {
			return fmt.Errorf("Bad: Destination Policy still exists: %q", rs.Primary.ID)
		}
	}

	return nil

}

func testAccCheckDestinationPolicyExists(n string, d *cloudwatchlogs.Destination) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LogsConn
		destination, exists, err := tflogs.LookupDestination(conn, rs.Primary.ID, nil)
		if err != nil {
			return err
		}
		if !exists || destination.AccessPolicy == nil {
			return fmt.Errorf("Bad: Destination Policy %q does not exist", rs.Primary.ID)
		}

		*d = *destination

		return nil
	}
}

func testAccDestinationPolicyBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = 1
}

data "aws_region" "current" {
}

data "aws_iam_policy_document" "role" {
  statement {
    effect = "Allow"

    principals {
      type = "Service"

      identifiers = [
        "logs.${data.aws_region.current.name}.amazonaws.com",
      ]
    }

    actions = [
      "sts:AssumeRole",
    ]
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.role.json
}

data "aws_iam_policy_document" "policy" {
  statement {
    effect = "Allow"

    actions = [
      "kinesis:PutRecord",
    ]

    resources = [
      aws_kinesis_stream.test.arn,
    ]
  }

  statement {
    effect = "Allow"

    actions = [
      "iam:PassRole",
    ]

    resources = [
      aws_iam_role.test.arn,
    ]
  }
}

resource "aws_iam_role_policy" "test" {
  name   = %[1]q
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.policy.json
}

resource "aws_cloudwatch_log_destination" "test" {
  name       = %[1]q
  target_arn = aws_kinesis_stream.test.arn
  role_arn   = aws_iam_role.test.arn
  depends_on = [aws_iam_role_policy.test]
}

data "aws_iam_policy_document" "access" {
  statement {
    effect = "Allow"

    principals {
      type = "AWS"

      identifiers = [
        "000000000000",
      ]
    }

    actions = [
      "logs:PutSubscriptionFilter",
    ]

    resources = [
      aws_cloudwatch_log_destination.test.arn,
    ]
  }
}
`, rName)
}

func testAccDestinationPolicyConfig(rName string) string {
	return testAccDestinationPolicyBaseConfig(rName) + `
resource "aws_cloudwatch_log_destination_policy" "test" {
  destination_name = aws_cloudwatch_log_destination.test.name
  access_policy    = data.aws_iam_policy_document.access.json
}
`
}

func testAccDestinationPolicyForceUpdateConfig(rName string) string {
	return testAccDestinationPolicyBaseConfig(rName) + `
data "aws_iam_policy_document" "access2" {
  statement {
    effect = "Allow"

    principals {
      type = "AWS"

      identifiers = [
        "000000000000",
      ]
    }

    actions = [
      "logs:*",
    ]

    resources = [
      aws_cloudwatch_log_destination.test.arn,
    ]
  }
}

resource "aws_cloudwatch_log_destination_policy" "test" {
  destination_name = aws_cloudwatch_log_destination.test.name
  access_policy    = data.aws_iam_policy_document.access2.json
  force_update     = true
}
`
}
