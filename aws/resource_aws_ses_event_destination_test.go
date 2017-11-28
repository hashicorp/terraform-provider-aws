package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSSESEventDestination_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSESEventDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSESEventDestinationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESEventConfigurationSetExists("aws_ses_configuration_set.test"),
					resource.TestCheckResourceAttr(
						"aws_ses_event_destination.kinesis", "name", "event-destination-kinesis"),
					resource.TestCheckResourceAttr(
						"aws_ses_event_destination.cloudwatch", "name", "event-destination-cloudwatch"),
					resource.TestCheckResourceAttr(
						"aws_ses_event_destination.sns", "name", fmt.Sprintf("event-destination-sns-%d", edRandomInteger)),
					testAccCheckAwsSESEventDestinationSnsArn("aws_ses_event_destination.sns"),
				),
			},
		},
	})
}

func testAccCheckSESEventDestinationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sesConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ses_configuration_set" {
			continue
		}

		response, err := conn.ListConfigurationSets(&ses.ListConfigurationSetsInput{})
		if err != nil {
			return err
		}

		found := false
		for _, element := range response.ConfigurationSets {
			if *element.Name == fmt.Sprintf("some-configuration-set-%d", edRandomInteger) {
				found = true
			}
		}

		if found {
			return fmt.Errorf("The configuration set still exists")
		}

	}

	return nil

}

func testAccCheckAwsSESEventConfigurationSetExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES configuration set not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES configuration set ID not set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sesConn

		response, err := conn.ListConfigurationSets(&ses.ListConfigurationSetsInput{})
		if err != nil {
			return err
		}

		found := false
		for _, element := range response.ConfigurationSets {
			if *element.Name == fmt.Sprintf("some-configuration-set-%d", edRandomInteger) {
				found = true
			}
		}

		if !found {
			return fmt.Errorf("The configuration set was not created")
		}

		return nil
	}
}

func testAccCheckAwsSESEventDestinationSnsArn(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).sesConn
		rs, _ := s.RootModule().Resources[n]

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SES Event Destination ID is set")
		}

		req := &ses.DescribeConfigurationSetInput{
			ConfigurationSetName:           aws.String(fmt.Sprintf("some-configuration-set-%d", edRandomInteger)),
			ConfigurationSetAttributeNames: []*string{aws.String("eventDestinations")},
		}
		resp, err := conn.DescribeConfigurationSet(req)
		if err != nil {
			return err
		}

		found := false
		dest := ""
		for _, element := range resp.EventDestinations {
			if *element.Name == fmt.Sprintf("event-destination-sns-%d", edRandomInteger) {
				dest = *element.SNSDestination.TopicARN
				found = true
			}
		}

		if !found {
			return fmt.Errorf("The SNS event destination was not created")
		}

		expected := fmt.Sprintf(
			"arn:%s:sns:%s:%s:%s",
			testAccProvider.Meta().(*AWSClient).partition,
			testAccProvider.Meta().(*AWSClient).region,
			testAccProvider.Meta().(*AWSClient).accountid,
			fmt.Sprintf("ses-destination-test-%d", edRandomInteger))

		if dest != expected {
			return fmt.Errorf("Incorrect ARN: expected %q, got %q", expected, dest)
		}

		return nil
	}
}

var edRandomInteger = acctest.RandInt()
var testAccAWSSESEventDestinationConfig = fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-format"
  acl = "private"
}

resource "aws_iam_role" "firehose_role" {
   name = "firehose_test_role_test"
   assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    },
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ses.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_kinesis_firehose_delivery_stream" "test_stream" {
  name = "terraform-kinesis-firehose-test-stream-test"
  destination = "s3"
  s3_configuration {
    role_arn = "${aws_iam_role.firehose_role.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"
  }
}

resource "aws_iam_role_policy" "firehose_delivery_policy" {
  name = "tf-delivery-policy-test"
  role = "${aws_iam_role.firehose_role.id}"
  policy = "${data.aws_iam_policy_document.fh_felivery_document.json}"
}

data "aws_iam_policy_document" "fh_felivery_document" {
    statement {
        sid = "GiveSESPermissionToPutFirehose"
        actions = [
            "firehose:PutRecord",
            "firehose:PutRecordBatch",
        ]
        resources = [
            "*",
        ]
    }
}

resource "aws_sns_topic" "ses_destination" {
  name = "ses-destination-test-%d"
}

resource "aws_ses_configuration_set" "test" {
    name = "some-configuration-set-%d"
}

resource "aws_ses_event_destination" "kinesis" {
  name = "event-destination-kinesis",
  configuration_set_name = "${aws_ses_configuration_set.test.name}",
  enabled = true,
  matching_types = ["bounce", "send"],

  kinesis_destination {
    stream_arn = "${aws_kinesis_firehose_delivery_stream.test_stream.arn}",
    role_arn = "${aws_iam_role.firehose_role.arn}"
  }
}

resource "aws_ses_event_destination" "cloudwatch" {
  name = "event-destination-cloudwatch",
  configuration_set_name = "${aws_ses_configuration_set.test.name}",
  enabled = true,
  matching_types = ["bounce", "send"],

  cloudwatch_destination {
    default_value = "default"
	dimension_name = "dimension"
	value_source = "emailHeader"
  }
}

resource "aws_ses_event_destination" "sns" {
  name = "event-destination-sns-%d",
  configuration_set_name = "${aws_ses_configuration_set.test.name}",
  enabled = true,
  matching_types = ["bounce", "send"],

  sns_destination {
    topic_arn = "${aws_sns_topic.ses_destination.arn}"
  }
}

`, edRandomInteger, edRandomInteger, edRandomInteger)
