package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func testAccConfigRemediationConfiguration_basic(t *testing.T) {
	var rc configservice.RemediationConfiguration
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
					testAccCheckConfigRemediationConfigurationName("aws_config_remediation_configuration.foo", expectedName, &rc),
					resource.TestCheckResourceAttr("aws_config_remediation_configuration.foo", "config_rule_name", expectedName),
					resource.TestCheckResourceAttr("aws_config_remediation_configuration.foo", "target_id", "SSM_DOCUMENT"),
					resource.TestCheckResourceAttr("aws_config_remediation_configuration.foo", "target_type", "AWS-PublishSNSNotification"),
					resource.TestCheckResourceAttr("aws_config_remediation_configuration.foo", "parameters.0.resource_value", "Message"),
				),
			},
		},
	})
}

func testAccConfigRemediationConfiguration_importAws(t *testing.T) {
	resourceName := "aws_config_remediation_configuration.foo"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigRemediationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigRemediationConfigurationConfig_ownerAws(rInt),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckConfigRemediationConfigurationName(n, desired string, obj *configservice.RemediationConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.Attributes["cnofig_rule_name"] != *obj.ConfigRuleName {
			return fmt.Errorf("Expected name: %q, given: %q", desired, *obj.ConfigRuleName)
		}
		return nil
	}
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
resource "aws_config_remediation_configuration" "foo" {
	config_rule_name = "${aws_config_config_rule.foo.name}"

	resource_type = ""
	target_id = "SSM_DOCUMENT"
	target_type = "AWS-PublishSNSNotification"
	target_version = "1"

	parameters = [
		{
			resource_value = "Message"
		},
		{
			static_value = {
				key = "TopicArn"
				value = "${aws_sns_topic.foo.arn}"
			}
		},
		{
			static_value = {
				key = "AutomationAssumeRole"
				value = "${aws_iam_role.aar.arn}"
			}
		}
	]

	depends_on = [
		"aws_config_config_rule.foo",
		"aws_sns_topic.foo"
	]
}

resource "aws_sns_topic" "foo" {
  name = "sns_topic_name"
}

resource "aws_config_config_rule" "foo" {
  name = "tf-acc-test-%d"

  source {
    owner             = "AWS"
    source_identifier = "S3_BUCKET_VERSIONING_ENABLED"
  }

  depends_on = ["aws_config_configuration_recorder.foo"]
}

resource "aws_config_configuration_recorder" "foo" {
  name     = "tf-acc-test-%d"
  role_arn = "${aws_iam_role.r.arn}"
}

resource "aws_iam_role" "r" {
  name = "tf-acc-test-awsconfig-%d"

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

resource "aws_iam_role_policy" "p" {
  name = "tf-acc-test-awsconfig-%d"
  role = "${aws_iam_role.r.id}"

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

func testAccConfigRemediationConfigurationConfig_ownerAws(randInt int) string {
	return fmt.Sprintf(`
resource "aws_config_remediation_configuration" "foo" {
	config_rule_name = "${aws_config_config_rule.foo.name}"

	resource_type = ""
	target_id = "SSM_DOCUMENT"
	target_type = "AWS-PublishSNSNotification"
	target_version = "1"

	parameters = [
		{
			resource_value = "Message"
		},
		{
			static_value = {
				key = "TopicArn"
				value = "${aws_sns_topic.foo.arn}"
			}
		},
		{
			static_value = {
				key = "AutomationAssumeRole"
				value = "${aws_iam_role.aar.arn}"
			}
		}
	]

	depends_on = [
		"aws_config_config_rule.foo",
		"aws_sns_topic.foo"
	]
}

resource "aws_sns_topic" "foo" {
	name = "sns_topic_name"
}

resource "aws_config_config_rule" "foo" {
  name        = "tf-acc-test-%d"
  description = "Terraform Acceptance tests"

  source {
    owner             = "AWS"
    source_identifier = "REQUIRED_TAGS"
  }

  scope {
    compliance_resource_id    = "blablah"
    compliance_resource_types = ["AWS::EC2::Instance"]
  }

  input_parameters = <<PARAMS
{"tag1Key":"CostCenter", "tag2Key":"Owner"}
PARAMS

  depends_on = ["aws_config_configuration_recorder.foo"]
}

resource "aws_config_configuration_recorder" "foo" {
  name     = "tf-acc-test-%d"
  role_arn = "${aws_iam_role.r.arn}"
}

resource "aws_iam_role" "r" {
  name = "tf-acc-test-awsconfig-%d"

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

resource "aws_iam_role_policy" "p" {
  name = "tf-acc-test-awsconfig-%d"
  role = "${aws_iam_role.r.id}"

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
