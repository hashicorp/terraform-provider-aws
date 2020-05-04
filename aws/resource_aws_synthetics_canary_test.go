package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/synthetics"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSSyntheticsCanary_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSyntheticsCanaryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSyntheticsCanaryBasicConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSyntheticsCanaryExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "synthetics", regexp.MustCompile(`canary/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"code"},
			},
		},
	})
}

func TestAccAWSSyntheticsCanary_tags(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSyntheticsCanaryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSyntheticsCanaryConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSyntheticsCanaryExists(resourceName),
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
				Config: testAccAWSSyntheticsCanaryConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSyntheticsCanaryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSSyntheticsCanaryConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSyntheticsCanaryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAwsSyntheticsCanaryDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).syntheticsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_synthetics_canary" {
			continue
		}

		name := rs.Primary.ID
		input := &synthetics.GetCanaryInput{
			Name: aws.String(name),
		}

		_, err := conn.GetCanary(input)
		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckAwsSyntheticsCanaryExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("synthetics Canary not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("synthetics Canary name not set")
		}

		name := rs.Primary.ID
		conn := testAccProvider.Meta().(*AWSClient).syntheticsconn

		input := &synthetics.GetCanaryInput{
			Name: aws.String(name),
		}

		_, err := conn.GetCanary(input)
		if err != nil {
			return fmt.Errorf("syntherics Canary %s not found in AWS", name)
		}
		return nil
	}
}

func testAccAWSSyntheticsCanaryConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

data "aws_partition" "current" {}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = "${aws_iam_role.test.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
        "Effect": "Allow",
        "Action": [
            "logs:CreateLogGroup",
            "logs:CreateLogStream",
            "logs:PutLogEvents"
        ],
        "Resource": "arn:${data.aws_partition.current.partition}:logs:*:*:*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:PutObject"
      ],
      "Resource": [
        "${aws_s3_bucket.test.arn}/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetBucketLocation",
        "s3:ListAllMyBuckets"
      ],
      "Resource": [
        "*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "cloudwatch:PutMetricData"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}
`, rName)
}

func testAccAWSSyntheticsCanaryBasicConfig(rName string) string {
	return testAccAWSSyntheticsCanaryConfigBase(rName) + fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn

  code {
    handler  = "exports.handler"
    zip_file = "test-fixtures/lambdatest.zip"
  }

  schedule {
    expression = "rate(0 minute)"
  }
}
`, rName)
}

func testAccAWSSyntheticsCanaryConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSSyntheticsCanaryConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
