package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccConfigRemediationConfiguration_basic(t *testing.T) {
	var rc configservice.RemediationConfiguration
	resourceName := "aws_config_remediation_configuration.foo"
	rInt := acctest.RandInt()
	prefix := "Original"
	version := 1
	expectedName := fmt.Sprintf("%s-tf-acc-test-%d", prefix, rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigRemediationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigRemediationConfigurationConfig(prefix, version, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRemediationConfigurationExists(resourceName, &rc),
					resource.TestCheckResourceAttr(resourceName, "config_rule_name", expectedName),
					resource.TestCheckResourceAttr(resourceName, "target_id", "SSM_DOCUMENT"),
					resource.TestCheckResourceAttr(resourceName, "target_type", "AWS-PublishSNSNotification"),
					resource.TestCheckResourceAttr(resourceName, "parameters.#", "2"),
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

func testAccConfigRemediationConfiguration_disappears(t *testing.T) {
	var rc configservice.RemediationConfiguration
	resourceName := "aws_config_remediation_configuration.test"
	rInt := acctest.RandInt()
	prefix := "original"
	version := 1

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigRemediationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigRemediationConfigurationConfig(prefix, version, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRemediationConfigurationExists(resourceName, &rc),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsConfigRemediationConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccConfigRemediationConfiguration_recreates(t *testing.T) {
	var original configservice.RemediationConfiguration
	var updated configservice.RemediationConfiguration
	resourceName := "aws_config_remediation_configuration.test"
	rInt := acctest.RandInt()

	originalName := "Original"
	updatedName := "Updated"
	version := 1

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigRemediationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigRemediationConfigurationConfig(originalName, version, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRemediationConfigurationExists(resourceName, &original),
					resource.TestCheckResourceAttr(resourceName, "config_rule_name", fmt.Sprintf("%s-tf-acc-test-%d", originalName, rInt)),
				),
			},
			{
				Config: testAccConfigRemediationConfigurationConfig(updatedName, version, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRemediationConfigurationExists(resourceName, &updated),
					testAccCheckConfigRemediationConfigurationRecreated(t, &original, &updated),
					resource.TestCheckResourceAttr(resourceName, "config_rule_name", fmt.Sprintf("%s-tf-acc-test-%d", updatedName, rInt)),
				),
			},
		},
	})
}

func testAccConfigRemediationConfiguration_updates(t *testing.T) {
	var original configservice.RemediationConfiguration
	var updated configservice.RemediationConfiguration
	resourceName := "aws_config_remediation_configuration.test"
	rInt := acctest.RandInt()

	name := "Original"
	originalVersion := 1
	updatedVersion := 2

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigRemediationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigRemediationConfigurationConfig(name, originalVersion, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRemediationConfigurationExists(resourceName, &original),
					resource.TestCheckResourceAttr(resourceName, "target_version", fmt.Sprintf("%d", originalVersion)),
				),
			},
			{
				Config: testAccConfigRemediationConfigurationConfig(name, updatedVersion, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRemediationConfigurationExists(resourceName, &updated),
					testAccCheckConfigRemediationConfigurationNotRecreated(t, &original, &updated),
					resource.TestCheckResourceAttr(resourceName, "target_version", fmt.Sprintf("%d", updatedVersion)),
				),
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

func testAccConfigRemediationConfigurationConfig(namePrefix string, targetVersion, randInt int) string {
	return fmt.Sprintf(`
resource "aws_config_remediation_configuration" "test" {
	config_rule_name = aws_config_config_rule.test.name

	resource_type = ""
	target_id = "SSM_DOCUMENT"
	target_type = "AWS-PublishSNSNotification"
	target_version = "%[2]d"

	parameter {
		resource_value = "Message"
	}

	parameter {
		static_value {
			key   = "TopicArn"
		f	value = aws_sns_topic.test.arn
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
  name = "%[1]s-tf-acc-test-%[3]d"

  source {
    owner             = "AWS"
    source_identifier = "S3_BUCKET_VERSIONING_ENABLED"
  }

  depends_on = [aws_config_configuration_recorder.test]
}

resource "aws_config_configuration_recorder" "test" {
  name     = "%[1]s-tf-acc-test-%[3]d"
  role_arn = aws_iam_role.r.arn
}

resource "aws_iam_role" "test" {
  name = "%[1]s-tf-acc-test-awsconfig-%[3]d"

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
  name = "%[1]s-tf-acc-test-awsconfig-%[3]d"
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
`, namePrefix, targetVersion, randInt)
}

func testAccCheckConfigRemediationConfigurationNotRecreated(t *testing.T,
	before, after *configservice.RemediationConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.Arn != *after.Arn {
			t.Fatalf("AWS Config Remediation arns have changed. Before %s. After %s", *before.Arn, *after.Arn)
		}
		return nil
	}
}

func testAccCheckConfigRemediationConfigurationRecreated(t *testing.T,
	before, after *configservice.RemediationConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.Arn == *after.Arn {
			t.Fatalf("AWS Config Remediation has not been recreated")
		}
		return nil
	}
}
