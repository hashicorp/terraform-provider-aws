package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func testAccConfigRemediationConfiguration_basic(t *testing.T) {
	var rc configservice.RemediationConfiguration
	resourceName := "aws_config_remediation_configuration.foo"
	rInt := acctest.RandInt()
	expectedName := fmt.Sprintf("tf-acc-test-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigRemediationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigRemediationConfigurationConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRemediationConfigurationExists("aws_config_remediation_configuration.foo", &rc),
					resource.TestCheckResourceAttr("aws_config_remediation_configuration.foo", "config_rule_name", expectedName),
					resource.TestCheckResourceAttr("aws_config_remediation_configuration.foo", "target_id", "SSM_DOCUMENT"),
					resource.TestCheckResourceAttr("aws_config_remediation_configuration.foo", "target_type", "AWS-PublishSNSNotification"),
					resource.TestCheckResourceAttr("aws_config_remediation_configuration.foo", "parameters.#", "2"),
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

func testAccCheckConfigRemediationConfigurationExists(n string, obj *configservice.RemediationConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No config rule ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).configconn
		out, err := conn.DescribeRemediationConfigurations(&configservice.DescribeRemediationConfigurationsInput{
			ConfigRuleNames: []*string{aws.String(rs.Primary.Attributes["name"])},
		})
		if err != nil {
			return fmt.Errorf("Failed to describe config rule: %s", err)
		}
		if len(out.RemediationConfigurations) < 1 {
			return fmt.Errorf("No config rule found when describing %q", rs.Primary.Attributes["name"])
		}

		rc := out.RemediationConfigurations[0]
		*obj = *rc

		return nil
	}
}

func testAccCheckConfigRemediationConfigurationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).configconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_config_remediation_configuration" {
			continue
		}

		resp, err := conn.DescribeRemediationConfigurations(&configservice.DescribeRemediationConfigurationsInput{
			ConfigRuleNames: []*string{aws.String(rs.Primary.Attributes["name"])},
		})

		if err == nil {
			if len(resp.RemediationConfigurations) != 0 &&
				*resp.RemediationConfigurations[0].ConfigRuleName == rs.Primary.Attributes["name"] {
				return fmt.Errorf("remediation configuration(s) still exist for rule: %s", rs.Primary.Attributes["name"])
			}
		}
	}

	return nil
}

func testAccConfigRemediationConfigurationConfig_basic(randInt int) string {
	return fmt.Sprintf(`
resource "aws_config_remediation_configuration" "test" {
	config_rule_name = aws_config_config_rule.test.name

	resource_type = ""
	target_id = "SSM_DOCUMENT"
	target_type = "AWS-PublishSNSNotification"
	target_version = "1"

	parameter {
		resource_value = "Message"
	}

	parameter {
		static_value {
			key   = "TopicArn"
			value = aws_sns_topic.test.arn
		}
	}

	parameter {
		static_value {
			key   = "AutomationAssumeRole"
			value = aws_iam_role.test.arn
		}
	}
}

resource "aws_sns_topic" "test" {
  name = "sns_topic_name"
}

resource "aws_config_config_rule" "test" {
  name = "tf-acc-test"

  source {
    owner             = "AWS"
    source_identifier = "S3_BUCKET_VERSIONING_ENABLED"
  }

  depends_on = [aws_config_configuration_recorder.test]
}

resource "aws_config_configuration_recorder" "test" {
  name     = "tf-acc-test"
  role_arn = aws_iam_role.r.arn
}

resource "aws_iam_role" "test" {
  name = "tf-acc-test-awsconfig"

  assume_role_policy = <<EOF
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
EOF
}

resource "aws_iam_role_policy" "test" {
  name = "tf-acc-test-awsconfig-%d"
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
        "Action": "config:Put*",
        "Effect": "Allow",
        "Resource": "*"

    }
  ]
}
EOF
}
`, randInt, randInt, randInt, randInt)
}
