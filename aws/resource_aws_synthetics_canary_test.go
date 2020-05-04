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
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "synthetics", regexp.MustCompile(`canary:.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file"},
			},
		},
	})
}

func TestAccAWSSyntheticsCanary_vpc(t *testing.T) {
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
				Config: testAccAWSSyntheticsCanaryVPCConfig1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSyntheticsCanaryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_config.0.vpc_config.0.vpc_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file"},
			},
			{
				Config: testAccAWSSyntheticsCanaryVPCConfig2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSyntheticsCanaryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "2"),
				),
			},
			{
				Config: testAccAWSSyntheticsCanaryVPCConfig3(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSyntheticsCanaryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
				),
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
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file"},
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
			if isAWSErr(err, synthetics.ErrCodeResourceNotFoundException, "") {
				return nil
			}
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
  handler  = "exports.handler"
  zip_file = "test-fixtures/lambdatest.zip"

  schedule {
    expression = "rate(0 minute)"
  }
}
`, rName)
}

func testAccAWSSyntheticsCanaryVPCConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_subnet" "test1" {
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "${cidrsubnet(aws_vpc.test.cidr_block, 2, 0)}"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "${cidrsubnet(aws_vpc.test.cidr_block, 2, 1)}"
  availability_zone = "${data.aws_availability_zones.available.names[1]}"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test1" {
  vpc_id = "${aws_vpc.test.id}"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test2" {
  vpc_id = "${aws_vpc.test.id}"

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
  role       = "${aws_iam_role.test.name}"
}
`, rName)
}

func testAccAWSSyntheticsCanaryVPCConfig1(rName string) string {
	return testAccAWSSyntheticsCanaryConfigBase(rName) +
		testAccAWSSyntheticsCanaryVPCConfigBase(rName) +
		fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler  = "exports.handler"
  zip_file = "test-fixtures/lambdatest.zip"

  schedule {
    expression = "rate(0 minute)"
  }

  vpc_config {
    subnet_ids         = ["${aws_subnet.test1.id}"]
    security_group_ids = ["${aws_security_group.test1.id}"]
  }

  depends_on = ["aws_iam_role_policy_attachment.test"]
}
`, rName)
}

func testAccAWSSyntheticsCanaryVPCConfig2(rName string) string {
	return testAccAWSSyntheticsCanaryConfigBase(rName) +
		testAccAWSSyntheticsCanaryVPCConfigBase(rName) +
		fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler  = "exports.handler"
  zip_file = "test-fixtures/lambdatest.zip"

  schedule {
    expression = "rate(0 minute)"
  }

  vpc_config {
    subnet_ids         = ["${aws_subnet.test1.id}", "${aws_subnet.test2.id}"]
    security_group_ids = ["${aws_security_group.test1.id}", "${aws_security_group.test2.id}"]
  }

  depends_on = ["aws_iam_role_policy_attachment.test"]

}
`, rName)
}

func testAccAWSSyntheticsCanaryVPCConfig3(rName string) string {
	return testAccAWSSyntheticsCanaryConfigBase(rName) +
		testAccAWSSyntheticsCanaryVPCConfigBase(rName) +
		fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler  = "exports.handler"
  zip_file = "test-fixtures/lambdatest.zip"

  schedule {
    expression = "rate(0 minute)"
  }

  vpc_config {
    subnet_ids         = ["${aws_subnet.test2.id}"]
    security_group_ids = ["${aws_security_group.test2.id}"]
  }

  depends_on = ["aws_iam_role_policy_attachment.test"]
}
`, rName)
}

func testAccAWSSyntheticsCanaryConfigTags1(rName, tagKey1, tagValue1 string) string {
	return testAccAWSSyntheticsCanaryConfigBase(rName) + fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"

  schedule {
    expression = "rate(0 minute)"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSSyntheticsCanaryConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSSyntheticsCanaryConfigBase(rName) + fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"

  schedule {
    expression = "rate(0 minute)"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
