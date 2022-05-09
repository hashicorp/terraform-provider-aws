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

func TestAccSageMakerDevice_basic(t *testing.T) {
	var device sagemaker.DescribeDeviceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_device.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceExists(resourceName, &device),
					resource.TestCheckResourceAttr(resourceName, "device_fleet_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("device-fleet/%[1]s/device/%[1]s", rName)),
					resource.TestCheckResourceAttr(resourceName, "device.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "device.0.device_name", rName),
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

func TestAccSageMakerDevice_description(t *testing.T) {
	var device sagemaker.DescribeDeviceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_device.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceDescription(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceExists(resourceName, &device),
					resource.TestCheckResourceAttr(resourceName, "device.0.description", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeviceDescription(rName, "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceExists(resourceName, &device),
					resource.TestCheckResourceAttr(resourceName, "device.0.description", "test"),
				),
			},
		},
	})
}

func TestAccSageMakerDevice_disappears(t *testing.T) {
	var device sagemaker.DescribeDeviceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_device.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceExists(resourceName, &device),
					acctest.CheckResourceDisappears(acctest.Provider, tfsagemaker.ResourceDevice(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfsagemaker.ResourceDevice(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSageMakerDevice_disappears_fleet(t *testing.T) {
	var device sagemaker.DescribeDeviceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_device.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceExists(resourceName, &device),
					acctest.CheckResourceDisappears(acctest.Provider, tfsagemaker.ResourceDeviceFleet(), "aws_sagemaker_device_fleet.test"),
					acctest.CheckResourceDisappears(acctest.Provider, tfsagemaker.ResourceDevice(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDeviceDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_device" {
			continue
		}

		deviceFleetName, deviceName, err := tfsagemaker.DecodeDeviceId(rs.Primary.ID)
		if err != nil {
			return err
		}

		device, err := tfsagemaker.FindDeviceByName(conn, deviceFleetName, deviceName)
		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		if aws.StringValue(device.DeviceName) == deviceName && aws.StringValue(device.DeviceFleetName) == deviceFleetName {
			return fmt.Errorf("SageMaker Device %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckDeviceExists(n string, device *sagemaker.DescribeDeviceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Sagmaker Device ID is set")
		}

		deviceFleetName, deviceName, err := tfsagemaker.DecodeDeviceId(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn
		resp, err := tfsagemaker.FindDeviceByName(conn, deviceFleetName, deviceName)
		if err != nil {
			return err
		}

		*device = *resp

		return nil
	}
}

func testAccDeviceBaseConfig(rName string) string {
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

resource "aws_sagemaker_device_fleet" "test" {
  device_fleet_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  output_config {
    s3_output_location = "s3://${aws_s3_bucket.test.bucket}/prefix/"
  }
}
`, rName)
}

func testAccDeviceBasicConfig(rName string) string {
	return testAccDeviceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_device" "test" {
  device_fleet_name = aws_sagemaker_device_fleet.test.device_fleet_name

  device {
    device_name = %[1]q
  }
}
`, rName)
}

func testAccDeviceDescription(rName, desc string) string {
	return testAccDeviceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_device" "test" {
  device_fleet_name = aws_sagemaker_device_fleet.test.device_fleet_name

  device {
    device_name = %[1]q
    description = %[2]q
  }
}
`, rName, desc)
}
