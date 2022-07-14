package ses_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfses "github.com/hashicorp/terraform-provider-aws/internal/service/ses"
)

func TestAccSESEventDestination_basic(t *testing.T) {
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix("tf-acc-test-kinesis")
	rName3 := sdkacctest.RandomWithPrefix("tf-acc-test-sns")
	cloudwatchDestinationResourceName := "aws_ses_event_destination.cloudwatch"
	kinesisDestinationResourceName := "aws_ses_event_destination.kinesis"
	snsDestinationResourceName := "aws_ses_event_destination.sns"
	var v1, v2, v3 ses.EventDestination

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventDestinationConfig_basic(rName1, rName2, rName3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventDestinationExists(cloudwatchDestinationResourceName, &v1),
					testAccCheckEventDestinationExists(kinesisDestinationResourceName, &v2),
					testAccCheckEventDestinationExists(snsDestinationResourceName, &v3),
					acctest.CheckResourceAttrRegionalARN(cloudwatchDestinationResourceName, "arn", "ses", fmt.Sprintf("configuration-set/%s:event-destination/%s", rName1, rName1)),
					acctest.CheckResourceAttrRegionalARN(kinesisDestinationResourceName, "arn", "ses", fmt.Sprintf("configuration-set/%s:event-destination/%s", rName1, rName2)),
					acctest.CheckResourceAttrRegionalARN(snsDestinationResourceName, "arn", "ses", fmt.Sprintf("configuration-set/%s:event-destination/%s", rName1, rName3)),
					resource.TestCheckResourceAttr(cloudwatchDestinationResourceName, "name", rName1),
					resource.TestCheckResourceAttr(kinesisDestinationResourceName, "name", rName2),
					resource.TestCheckResourceAttr(snsDestinationResourceName, "name", rName3),
				),
			},
			{
				ResourceName:      cloudwatchDestinationResourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", rName1, rName1),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      kinesisDestinationResourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", rName1, rName2),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      snsDestinationResourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", rName1, rName3),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSESEventDestination_disappears(t *testing.T) {
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix("tf-acc-test-kinesis")
	rName3 := sdkacctest.RandomWithPrefix("tf-acc-test-sns")
	cloudwatchDestinationResourceName := "aws_ses_event_destination.cloudwatch"
	kinesisDestinationResourceName := "aws_ses_event_destination.kinesis"
	snsDestinationResourceName := "aws_ses_event_destination.sns"
	var v1, v2, v3 ses.EventDestination

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventDestinationConfig_basic(rName1, rName2, rName3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventDestinationExists(cloudwatchDestinationResourceName, &v1),
					testAccCheckEventDestinationExists(kinesisDestinationResourceName, &v2),
					testAccCheckEventDestinationExists(snsDestinationResourceName, &v3),
					acctest.CheckResourceDisappears(acctest.Provider, tfses.ResourceEventDestination(), cloudwatchDestinationResourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfses.ResourceEventDestination(), kinesisDestinationResourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfses.ResourceEventDestination(), snsDestinationResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEventDestinationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn

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
			if aws.StringValue(element.Name) == rs.Primary.ID {
				found = true
			}
		}

		if found {
			return fmt.Errorf("The configuration set still exists")
		}

	}

	return nil

}

func testAccCheckEventDestinationExists(n string, v *ses.EventDestination) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES event destination not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES event destination ID not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn

		response, err := conn.DescribeConfigurationSet(&ses.DescribeConfigurationSetInput{
			ConfigurationSetAttributeNames: aws.StringSlice([]string{ses.ConfigurationSetAttributeEventDestinations}),
			ConfigurationSetName:           aws.String(rs.Primary.Attributes["configuration_set_name"]),
		})
		if err != nil {
			return err
		}

		for _, eventDestination := range response.EventDestinations {
			if aws.StringValue(eventDestination.Name) == rs.Primary.ID {
				*v = *eventDestination
				return nil
			}
		}

		return fmt.Errorf("The SES Configuration Set Event Destination was not found")
	}
}

func testAccEventDestinationConfig_basic(rName1, rName2, rName3 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[2]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_iam_role" "test" {
  name = %[2]q

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

resource "aws_kinesis_firehose_delivery_stream" "test" {
  name        = %[2]q
  destination = "s3"

  s3_configuration {
    role_arn   = aws_iam_role.test.arn
    bucket_arn = aws_s3_bucket.test.arn
  }
}

resource "aws_iam_role_policy" "test" {
  name   = %[2]q
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
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

resource "aws_sns_topic" "test" {
  name = %[3]q
}

resource "aws_ses_configuration_set" "test" {
  name = %[1]q
}

resource "aws_ses_event_destination" "kinesis" {
  name                   = %[2]q
  configuration_set_name = aws_ses_configuration_set.test.name
  enabled                = true
  matching_types         = ["bounce", "send"]

  kinesis_destination {
    stream_arn = aws_kinesis_firehose_delivery_stream.test.arn
    role_arn   = aws_iam_role.test.arn
  }
}

resource "aws_ses_event_destination" "cloudwatch" {
  name                   = %[1]q
  configuration_set_name = aws_ses_configuration_set.test.name
  enabled                = true
  matching_types         = ["bounce", "send"]

  cloudwatch_destination {
    default_value  = "default"
    dimension_name = "dimension"
    value_source   = "emailHeader"
  }

  cloudwatch_destination {
    default_value  = "default"
    dimension_name = "ses:source-ip"
    value_source   = "messageTag"
  }
}

resource "aws_ses_event_destination" "sns" {
  name                   = %[3]q
  configuration_set_name = aws_ses_configuration_set.test.name
  enabled                = true
  matching_types         = ["bounce", "send"]

  sns_destination {
    topic_arn = aws_sns_topic.test.arn
  }
}
`, rName1, rName2, rName3)
}
