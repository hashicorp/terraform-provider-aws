package sagemaker_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccSageMakerDeviceFleet_basic(t *testing.T) {
	var deviceFleet sagemaker.DescribeDeviceFleetOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_device_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDeviceFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceFleetBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFleetExists(resourceName, &deviceFleet),
					resource.TestCheckResourceAttr(resourceName, "device_fleet_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("device-fleet/%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "output_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "output_config.0.s3_output_location", fmt.Sprintf("s3://%s/prefix/", rName)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "enable_iot_role_alias", "false"),
					resource.TestCheckResourceAttr(resourceName, "iot_role_alias", ""),
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

func TestAccSageMakerDeviceFleet_description(t *testing.T) {
	var deviceFleet sagemaker.DescribeDeviceFleetOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_device_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDeviceFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceFleetDescription(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFleetExists(resourceName, &deviceFleet),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeviceFleetDescription(rName, "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFleetExists(resourceName, &deviceFleet),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
				),
			},
		},
	})
}

func TestAccSageMakerDeviceFleet_tags(t *testing.T) {
	var deviceFleet sagemaker.DescribeDeviceFleetOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_device_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDeviceFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceFleetTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFleetExists(resourceName, &deviceFleet),
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
				Config: testAccDeviceFleetTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFleetExists(resourceName, &deviceFleet),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDeviceFleetTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFleetExists(resourceName, &deviceFleet),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccSageMakerDeviceFleet_disappears(t *testing.T) {
	var deviceFleet sagemaker.DescribeDeviceFleetOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_device_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDeviceFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceFleetBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFleetExists(resourceName, &deviceFleet),
					acctest.CheckResourceDisappears(acctest.Provider, tfsagemaker.ResourceDeviceFleet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDeviceFleetDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_device_fleet" {
			continue
		}

		deviceFleet, err := tfsagemaker.FindDeviceFleetByName(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		if aws.StringValue(deviceFleet.DeviceFleetName) == rs.Primary.ID {
			return fmt.Errorf("sagemaker Device Fleet %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckDeviceFleetExists(n string, device_fleet *sagemaker.DescribeDeviceFleetOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker Device Fleet ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn
		resp, err := tfsagemaker.FindDeviceFleetByName(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*device_fleet = *resp

		return nil
	}
}

func testAccDeviceFleetBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
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
        "s3:GetBucketLocation"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonSageMakerEdgeDeviceFleetPolicy"
}
`, rName)
}

func testAccDeviceFleetBasicConfig(rName string) string {
	return testAccDeviceFleetBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_device_fleet" "test" {
  device_fleet_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  output_config {
    s3_output_location = "s3://${aws_s3_bucket.test.bucket}/prefix/"
  }
}
`, rName)
}

func testAccDeviceFleetDescription(rName, desc string) string {
	return testAccDeviceFleetBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_device_fleet" "test" {
  device_fleet_name = %[1]q
  role_arn          = aws_iam_role.test.arn
  description       = %[2]q

  output_config {
    s3_output_location = "s3://${aws_s3_bucket.test.bucket}/prefix/"
  }
}
`, rName, desc)
}

func testAccDeviceFleetTags1Config(rName, tagKey1, tagValue1 string) string {
	return testAccDeviceFleetBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_device_fleet" "test" {
  device_fleet_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  output_config {
    s3_output_location = "s3://${aws_s3_bucket.test.bucket}/prefix/"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccDeviceFleetTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccDeviceFleetBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_device_fleet" "test" {
  device_fleet_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  output_config {
    s3_output_location = "s3://${aws_s3_bucket.test.bucket}/prefix/"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
