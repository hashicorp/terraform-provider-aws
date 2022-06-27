package configservice_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func testAccConfigurationRecorder_basic(t *testing.T) {
	var cr configservice.ConfigurationRecorder
	rInt := sdkacctest.RandInt()
	expectedName := fmt.Sprintf("tf-acc-test-%d", rInt)
	expectedRoleName := fmt.Sprintf("tf-acc-test-awsconfig-%d", rInt)

	resourceName := "aws_config_configuration_recorder.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConfigurationRecorderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationRecorderConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationRecorderExists(resourceName, &cr),
					testAccCheckConfigurationRecorderName(resourceName, expectedName, &cr),
					acctest.CheckResourceAttrGlobalARN(resourceName, "role_arn", "iam", fmt.Sprintf("role/%s", expectedRoleName)),
					resource.TestCheckResourceAttr(resourceName, "name", expectedName),
				),
			},
		},
	})
}

func testAccConfigurationRecorder_allParams(t *testing.T) {
	var cr configservice.ConfigurationRecorder
	rInt := sdkacctest.RandInt()
	expectedName := fmt.Sprintf("tf-acc-test-%d", rInt)
	expectedRoleName := fmt.Sprintf("tf-acc-test-awsconfig-%d", rInt)

	resourceName := "aws_config_configuration_recorder.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConfigurationRecorderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationRecorderConfig_allParams(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationRecorderExists(resourceName, &cr),
					testAccCheckConfigurationRecorderName(resourceName, expectedName, &cr),
					acctest.CheckResourceAttrGlobalARN(resourceName, "role_arn", "iam", fmt.Sprintf("role/%s", expectedRoleName)),
					resource.TestCheckResourceAttr(resourceName, "name", expectedName),
					resource.TestCheckResourceAttr(resourceName, "recording_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "recording_group.0.all_supported", "false"),
					resource.TestCheckResourceAttr(resourceName, "recording_group.0.include_global_resource_types", "false"),
					resource.TestCheckResourceAttr(resourceName, "recording_group.0.resource_types.#", "2"),
				),
			},
		},
	})
}

func testAccConfigurationRecorder_importBasic(t *testing.T) {
	resourceName := "aws_config_configuration_recorder.foo"
	rInt := sdkacctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConfigurationRecorderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationRecorderConfig_basic(rInt),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckConfigurationRecorderName(n string, desired string, obj *configservice.ConfigurationRecorder) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if *obj.Name != desired {
			return fmt.Errorf("Expected configuration recorder %q name to be %q, given: %q",
				n, desired, *obj.Name)
		}

		return nil
	}
}

func testAccCheckConfigurationRecorderExists(n string, obj *configservice.ConfigurationRecorder) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No configuration recorder ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceConn
		out, err := conn.DescribeConfigurationRecorders(&configservice.DescribeConfigurationRecordersInput{
			ConfigurationRecorderNames: []*string{aws.String(rs.Primary.Attributes["name"])},
		})
		if err != nil {
			return fmt.Errorf("Failed to describe configuration recorder: %s", err)
		}
		if len(out.ConfigurationRecorders) < 1 {
			return fmt.Errorf("No configuration recorder found when describing %q", rs.Primary.Attributes["name"])
		}

		cr := out.ConfigurationRecorders[0]
		*obj = *cr

		return nil
	}
}

func testAccCheckConfigurationRecorderDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_config_configuration_recorder_status" {
			continue
		}

		resp, err := conn.DescribeConfigurationRecorders(&configservice.DescribeConfigurationRecordersInput{
			ConfigurationRecorderNames: []*string{aws.String(rs.Primary.Attributes["name"])},
		})

		if err == nil {
			if len(resp.ConfigurationRecorders) != 0 &&
				*resp.ConfigurationRecorders[0].Name == rs.Primary.Attributes["name"] {
				return fmt.Errorf("Configuration recorder still exists: %s", rs.Primary.Attributes["name"])
			}
		}
	}

	return nil
}

func testAccConfigurationRecorderConfig_basic(randInt int) string {
	return fmt.Sprintf(`
resource "aws_config_configuration_recorder" "foo" {
  name     = "tf-acc-test-%d"
  role_arn = aws_iam_role.r.arn
}

resource "aws_iam_role" "r" {
  name = "tf-acc-test-awsconfig-%d"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "config.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "p" {
  name = "tf-acc-test-awsconfig-%d"
  role = aws_iam_role.r.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "s3:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "${aws_s3_bucket.b.arn}",
        "${aws_s3_bucket.b.arn}/*"
      ]
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "b" {
  bucket        = "tf-acc-test-awsconfig-%d"
  force_destroy = true
}

resource "aws_config_delivery_channel" "foo" {
  name           = "tf-acc-test-awsconfig-%d"
  s3_bucket_name = aws_s3_bucket.b.bucket
  depends_on     = [aws_config_configuration_recorder.foo]
}
`, randInt, randInt, randInt, randInt, randInt)
}

func testAccConfigurationRecorderConfig_allParams(randInt int) string {
	return fmt.Sprintf(`
resource "aws_config_configuration_recorder" "foo" {
  name     = "tf-acc-test-%d"
  role_arn = aws_iam_role.r.arn

  recording_group {
    all_supported                 = false
    include_global_resource_types = false
    resource_types                = ["AWS::EC2::Instance", "AWS::CloudTrail::Trail"]
  }
}

resource "aws_iam_role" "r" {
  name = "tf-acc-test-awsconfig-%d"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "config.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "p" {
  name = "tf-acc-test-awsconfig-%d"
  role = aws_iam_role.r.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "s3:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "${aws_s3_bucket.b.arn}",
        "${aws_s3_bucket.b.arn}/*"
      ]
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "b" {
  bucket        = "tf-acc-test-awsconfig-%d"
  force_destroy = true
}

resource "aws_config_delivery_channel" "foo" {
  name           = "tf-acc-test-awsconfig-%d"
  s3_bucket_name = aws_s3_bucket.b.bucket
  depends_on     = [aws_config_configuration_recorder.foo]
}
`, randInt, randInt, randInt, randInt, randInt)
}
