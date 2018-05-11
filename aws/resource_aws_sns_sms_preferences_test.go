package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSSNSSMSPreferences_empty(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSNSSMSPrefsDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSSNSSMSPreferences_empty,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("aws_sns_sms_preferences.test_pref", "monthly_spend_limit"),
					resource.TestCheckNoResourceAttr("aws_sns_sms_preferences.test_pref", "delivery_status_iam_role_arn"),
					resource.TestCheckNoResourceAttr("aws_sns_sms_preferences.test_pref", "delivery_status_success_sampling_rate"),
					resource.TestCheckNoResourceAttr("aws_sns_sms_preferences.test_pref", "default_sender_id"),
					resource.TestCheckNoResourceAttr("aws_sns_sms_preferences.test_pref", "default_sms_type"),
					resource.TestCheckNoResourceAttr("aws_sns_sms_preferences.test_pref", "usage_report_s3_bucket"),
				),
			},
		},
	})
}

func TestAccAWSSNSSMSPreferences_defaultSMSType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSNSSMSPrefsDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSSNSSMSPreferences_defSMSType,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("aws_sns_sms_preferences.test_pref", "monthly_spend_limit"),
					resource.TestCheckNoResourceAttr("aws_sns_sms_preferences.test_pref", "delivery_status_iam_role_arn"),
					resource.TestCheckNoResourceAttr("aws_sns_sms_preferences.test_pref", "delivery_status_success_sampling_rate"),
					resource.TestCheckNoResourceAttr("aws_sns_sms_preferences.test_pref", "default_sender_id"),
					resource.TestCheckResourceAttr("aws_sns_sms_preferences.test_pref", "default_sms_type", "Transactional"),
					resource.TestCheckNoResourceAttr("aws_sns_sms_preferences.test_pref", "usage_report_s3_bucket"),
				),
			},
		},
	})
}

func TestAccAWSSNSSMSPreferences_almostAll(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSNSSMSPrefsDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSSNSSMSPreferences_almostAll,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_sns_sms_preferences.test_pref", "monthly_spend_limit", "1"),
					resource.TestCheckResourceAttr("aws_sns_sms_preferences.test_pref", "default_sms_type", "Transactional"),
					resource.TestCheckResourceAttr("aws_sns_sms_preferences.test_pref", "usage_report_s3_bucket", "some-bucket"),
				),
			},
		},
	})
}

func TestAccAWSSNSSMSPreferences_deliveryRole(t *testing.T) {
	arnRole := regexp.MustCompile(`^arn:aws:iam::\d+:role/test_smsdelivery_role$`)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSNSSMSPrefsDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSSNSSMSPreferences_deliveryRole,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_sns_sms_preferences.test_pref", "delivery_status_iam_role_arn", arnRole),
					resource.TestCheckResourceAttr("aws_sns_sms_preferences.test_pref", "delivery_status_success_sampling_rate", "75"),
				),
			},
		},
	})
}

func testAccCheckAWSSNSSMSPrefsDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sns_sms_preferences" {
			continue
		}

		return fmt.Errorf("SNS SMS Preferences resource exists when it should be destroyed!")
	}

	return nil
}

const testAccAWSSNSSMSPreferences_empty = `
resource "aws_sns_sms_preferences" "test_pref" {}
`
const testAccAWSSNSSMSPreferences_defSMSType = `
resource "aws_sns_sms_preferences" "test_pref" {
	default_sms_type = "Transactional"
}
`
const testAccAWSSNSSMSPreferences_almostAll = `
resource "aws_sns_sms_preferences" "test_pref" {
	monthly_spend_limit = "1",
	default_sms_type = "Transactional",
	usage_report_s3_bucket = "some-bucket",
}
`
const testAccAWSSNSSMSPreferences_deliveryRole = `
resource "aws_iam_role" "test_smsdelivery_role" {
    name = "test_smsdelivery_role"
    path = "/"
    assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "sns.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "test_smsdelivery_role_policy" {
  name   = "test_smsdelivery_role_policy"
  role   = "${aws_iam_role.test_smsdelivery_role.id}"
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": ["logs:CreateLogGroup","logs:CreateLogStream","logs:PutLogEvents","logs:PutMetricFilter","logs:PutRetentionPolicy"],
      "Resource": "*",
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_sns_sms_preferences" "test_pref" {
	delivery_status_iam_role_arn = "${aws_iam_role.test_smsdelivery_role.arn}",
	delivery_status_success_sampling_rate = "75",
}
`
