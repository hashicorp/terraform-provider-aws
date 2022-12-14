package logs_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccLogsDestination_basic(t *testing.T) {
	var destination cloudwatchlogs.Destination
	resourceName := "aws_cloudwatch_log_destination.test"
	streamResourceName := "aws_kinesis_stream.test"
	roleResourceName := "aws_iam_role.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDestinationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDestinationExists(resourceName, &destination),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "logs", regexp.MustCompile(`destination:.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", roleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "target_arn", streamResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLogsDestination_disappears(t *testing.T) {
	var destination cloudwatchlogs.Destination
	resourceName := "aws_cloudwatch_log_destination.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDestinationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDestinationExists(resourceName, &destination),
					acctest.CheckResourceDisappears(acctest.Provider, tflogs.ResourceDestination(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLogsDestination_tags(t *testing.T) {
	var destination cloudwatchlogs.Destination
	resourceName := "aws_cloudwatch_log_destination.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDestinationConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDestinationExists(resourceName, &destination),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDestinationConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDestinationExists(resourceName, &destination),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDestinationConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDestinationExists(resourceName, &destination),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckDestinationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LogsConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_log_destination" {
			continue
		}
		_, err := tflogs.FindDestinationByName(context.Background(), conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("CloudWatch Logs Destination still exists: %s", rs.Primary.ID)
	}

	return nil
}

func testAccCheckDestinationExists(n string, v *cloudwatchlogs.Destination) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CloudWatch Logs Destination ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LogsConn

		output, err := tflogs.FindDestinationByName(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccDestinationConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = 1
}

data "aws_region" "current" {}

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
`, rName)
}

func testAccDestinationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDestinationConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_destination" "test" {
  name       = %[1]q
  target_arn = aws_kinesis_stream.test.arn
  role_arn   = aws_iam_role.test.arn

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}

func testAccDestinationConfig_tags1(rName, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(testAccDestinationConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_destination" "test" {
  name       = %[1]q
  target_arn = aws_kinesis_stream.test.arn
  role_arn   = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName, tag1Key, tag1Value))
}

func testAccDestinationConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(testAccDestinationConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_destination" "test" {
  name       = %[1]q
  target_arn = aws_kinesis_stream.test.arn
  role_arn   = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value))
}
