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

func testAccConfigurationRecorderStatus_basic(t *testing.T) {
	var cr configservice.ConfigurationRecorder
	var crs configservice.ConfigurationRecorderStatus
	rInt := sdkacctest.RandInt()
	expectedName := fmt.Sprintf("tf-acc-test-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConfigurationRecorderStatusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationRecorderStatusConfig(rInt, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationRecorderExists("aws_config_configuration_recorder.foo", &cr),
					testAccCheckConfigurationRecorderStatusExists("aws_config_configuration_recorder_status.foo", &crs),
					testAccCheckConfigurationRecorderStatus("aws_config_configuration_recorder_status.foo", false, &crs),
					resource.TestCheckResourceAttr("aws_config_configuration_recorder_status.foo", "is_enabled", "false"),
					resource.TestCheckResourceAttr("aws_config_configuration_recorder_status.foo", "name", expectedName),
				),
			},
		},
	})
}

func testAccConfigurationRecorderStatus_startEnabled(t *testing.T) {
	var cr configservice.ConfigurationRecorder
	var crs configservice.ConfigurationRecorderStatus
	rInt := sdkacctest.RandInt()
	expectedName := fmt.Sprintf("tf-acc-test-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConfigurationRecorderStatusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationRecorderStatusConfig(rInt, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationRecorderExists("aws_config_configuration_recorder.foo", &cr),
					testAccCheckConfigurationRecorderStatusExists("aws_config_configuration_recorder_status.foo", &crs),
					testAccCheckConfigurationRecorderStatus("aws_config_configuration_recorder_status.foo", true, &crs),
					resource.TestCheckResourceAttr("aws_config_configuration_recorder_status.foo", "is_enabled", "true"),
					resource.TestCheckResourceAttr("aws_config_configuration_recorder_status.foo", "name", expectedName),
				),
			},
			{
				Config: testAccConfigurationRecorderStatusConfig(rInt, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationRecorderExists("aws_config_configuration_recorder.foo", &cr),
					testAccCheckConfigurationRecorderStatusExists("aws_config_configuration_recorder_status.foo", &crs),
					testAccCheckConfigurationRecorderStatus("aws_config_configuration_recorder_status.foo", false, &crs),
					resource.TestCheckResourceAttr("aws_config_configuration_recorder_status.foo", "is_enabled", "false"),
					resource.TestCheckResourceAttr("aws_config_configuration_recorder_status.foo", "name", expectedName),
				),
			},
			{
				Config: testAccConfigurationRecorderStatusConfig(rInt, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationRecorderExists("aws_config_configuration_recorder.foo", &cr),
					testAccCheckConfigurationRecorderStatusExists("aws_config_configuration_recorder_status.foo", &crs),
					testAccCheckConfigurationRecorderStatus("aws_config_configuration_recorder_status.foo", true, &crs),
					resource.TestCheckResourceAttr("aws_config_configuration_recorder_status.foo", "is_enabled", "true"),
					resource.TestCheckResourceAttr("aws_config_configuration_recorder_status.foo", "name", expectedName),
				),
			},
		},
	})
}

func testAccConfigurationRecorderStatus_importBasic(t *testing.T) {
	resourceName := "aws_config_configuration_recorder_status.foo"
	rInt := sdkacctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConfigurationRecorderStatusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationRecorderStatusConfig(rInt, true),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckConfigurationRecorderStatusExists(n string, obj *configservice.ConfigurationRecorderStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceConn
		out, err := conn.DescribeConfigurationRecorderStatus(&configservice.DescribeConfigurationRecorderStatusInput{
			ConfigurationRecorderNames: []*string{aws.String(rs.Primary.Attributes["name"])},
		})
		if err != nil {
			return fmt.Errorf("Failed to describe status of configuration recorder: %s", err)
		}
		if len(out.ConfigurationRecordersStatus) < 1 {
			return fmt.Errorf("Configuration Recorder %q not found", rs.Primary.Attributes["name"])
		}

		status := out.ConfigurationRecordersStatus[0]
		*obj = *status

		return nil
	}
}

func testAccCheckConfigurationRecorderStatus(n string, desired bool, obj *configservice.ConfigurationRecorderStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if *obj.Recording != desired {
			return fmt.Errorf("Expected configuration recorder %q recording to be %t, given: %t",
				n, desired, *obj.Recording)
		}

		return nil
	}
}

func testAccCheckConfigurationRecorderStatusDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_config_configuration_recorder_status" {
			continue
		}

		resp, err := conn.DescribeConfigurationRecorderStatus(&configservice.DescribeConfigurationRecorderStatusInput{
			ConfigurationRecorderNames: []*string{aws.String(rs.Primary.Attributes["name"])},
		})

		if err == nil {
			if len(resp.ConfigurationRecordersStatus) != 0 &&
				*resp.ConfigurationRecordersStatus[0].Name == rs.Primary.Attributes["name"] &&
				*resp.ConfigurationRecordersStatus[0].Recording {
				return fmt.Errorf("Configuration recorder is still recording: %s", rs.Primary.Attributes["name"])
			}
		}
	}

	return nil
}

func testAccConfigurationRecorderStatusConfig(randInt int, enabled bool) string {
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
}

resource "aws_config_configuration_recorder_status" "foo" {
  name       = aws_config_configuration_recorder.foo.name
  is_enabled = %t
  depends_on = [aws_config_delivery_channel.foo]
}
`, randInt, randInt, randInt, randInt, randInt, enabled)
}
