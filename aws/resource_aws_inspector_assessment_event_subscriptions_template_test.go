package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSInspectorTemplateEventSubscriptions_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSInspectorTemplateDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSInspectorTemplateAssessmentConfigTwoEventSubscritpions(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSInspectorTemplateExists("aws_inspector_assessment_template.test"),
					resource.TestCheckResourceAttr("aws_inspector_assessment_template.test", "subscribe_to_event.#", "2"),
					resource.TestCheckResourceAttr("aws_inspector_assessment_template.test", "subscribe_to_event.1757678435.event", "ASSESSMENT_RUN_STARTED"),
					resource.TestCheckResourceAttrSet("aws_inspector_assessment_template.test", "subscribe_to_event.1757678435.topic_arn"),
					resource.TestCheckResourceAttr("aws_inspector_assessment_template.test", "subscribe_to_event.3071367774.event", "ASSESSMENT_RUN_COMPLETED"),
					resource.TestCheckResourceAttrSet("aws_inspector_assessment_template.test", "subscribe_to_event.3071367774.topic_arn"),
				),
			},
		},
	})
}

func TestAccAWSInspectorTemplateEventSubscriptions_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSInspectorTemplateDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSInspectorTemplateAssessmentConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSInspectorTemplateExists("aws_inspector_assessment_template.test"),
					resource.TestCheckResourceAttr("aws_inspector_assessment_template.test", "subscribe_to_event.#", "0"),
				),
			},
			resource.TestStep{
				Config: testAccAWSInspectorTemplateAssessmentConfigTwoEventSubscritpions(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSInspectorTargetExists("aws_inspector_assessment_template.test"),
					resource.TestCheckResourceAttr("aws_inspector_assessment_template.test", "subscribe_to_event.#", "2"),
					resource.TestCheckResourceAttr("aws_inspector_assessment_template.test", "subscribe_to_event.1757678435.event", "ASSESSMENT_RUN_STARTED"),
					resource.TestCheckResourceAttrSet("aws_inspector_assessment_template.test", "subscribe_to_event.1757678435.topic_arn"),
					resource.TestCheckResourceAttr("aws_inspector_assessment_template.test", "subscribe_to_event.3071367774.event", "ASSESSMENT_RUN_COMPLETED"),
					resource.TestCheckResourceAttrSet("aws_inspector_assessment_template.test", "subscribe_to_event.3071367774.topic_arn"),
				),
			},
			resource.TestStep{
				Config: testAccAWSInspectorTemplateAssessmentConfigReplaceOneEventSubscription(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSInspectorTargetExists("aws_inspector_assessment_template.test"),
					resource.TestCheckResourceAttr("aws_inspector_assessment_template.test", "subscribe_to_event.#", "2"),
					resource.TestCheckResourceAttr("aws_inspector_assessment_template.test", "subscribe_to_event.1757678435.event", "ASSESSMENT_RUN_STARTED"),
					resource.TestCheckResourceAttrSet("aws_inspector_assessment_template.test", "subscribe_to_event.1757678435.topic_arn"),
					resource.TestCheckResourceAttr("aws_inspector_assessment_template.test", "subscribe_to_event.3886651376.event", "FINDING_REPORTED"),
					resource.TestCheckResourceAttrSet("aws_inspector_assessment_template.test", "subscribe_to_event.3886651376.topic_arn"),
				),
			},
			resource.TestStep{
				Config: testAccAWSInspectorTemplateAssessmentConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSInspectorTargetExists("aws_inspector_assessment_template.test"),
					resource.TestCheckResourceAttr("aws_inspector_assessment_template.test", "subscribe_to_event.#", "0"),
				),
			},
		},
	})
}

func testAccAWSInspectorTemplateAssessmentConfigBasic() string {
	return fmt.Sprintf(testAccAWSInspectorTemplateAssessmentConfig, "", "")
}

func testAccAWSInspectorTemplateAssessmentConfigTwoEventSubscritpions() string {
	return fmt.Sprintf(testAccAWSInspectorTemplateAssessmentConfig,
		AWSInspectorEventSubscriptionsSNSTopicAndIAMPolicy,
		AWSInspectorTwoEventSubscriptions)
}

func testAccAWSInspectorTemplateAssessmentConfigReplaceOneEventSubscription() string {
	return fmt.Sprintf(testAccAWSInspectorTemplateAssessmentConfig,
		AWSInspectorEventSubscriptionsSNSTopicAndIAMPolicy,
		AWSInspectorReplacedEventSubscriptions)
}

var testAccAWSInspectorTemplateAssessmentConfig = `
data "aws_inspector_rules_packages" "rules" {}

resource "aws_inspector_resource_group" "test" {
  tags {
    Name = "bar"
  }
}

resource "aws_inspector_assessment_target" "test" {
  name               = "test"
  resource_group_arn = "${aws_inspector_resource_group.test.arn}"
}

%s

resource "aws_inspector_assessment_template" "test" {
  name       = "test template"
  target_arn = "${aws_inspector_assessment_target.test.arn}"
  duration   = 3600

  rules_package_arns = ["${data.aws_inspector_rules_packages.rules.arns}"]

  %s
}`

var AWSInspectorTwoEventSubscriptions = `
subscribe_to_event {
  event     = "ASSESSMENT_RUN_STARTED"
  topic_arn = "${aws_sns_topic.test_sns_topic_for_inspector.arn}"
}

subscribe_to_event {
  event     = "ASSESSMENT_RUN_COMPLETED"
  topic_arn = "${aws_sns_topic.test_sns_topic_for_inspector.arn}"
}
`

var AWSInspectorReplacedEventSubscriptions = `
subscribe_to_event {
  event     = "ASSESSMENT_RUN_STARTED"
  topic_arn = "${aws_sns_topic.test_sns_topic_for_inspector.arn}"
}

subscribe_to_event {
  event     = "FINDING_REPORTED"
  topic_arn = "${aws_sns_topic.test_sns_topic_for_inspector.arn}"
}
`

var AWSInspectorEventSubscriptionsSNSTopicAndIAMPolicy = `
data "aws_caller_identity" "current" { }

resource "aws_sns_topic" "test_sns_topic_for_inspector" {
  name = "test_sns_topic_for_inspector"
}

resource "aws_sns_topic_policy" "test_sns_topic_for_inspector" {
  arn    = "${aws_sns_topic.test_sns_topic_for_inspector.arn}"
  policy = "${data.aws_iam_policy_document.inspector-allow-write-to-test-sns-topic.json}"
}

data "aws_iam_policy_document" "inspector-allow-write-to-test-sns-topic" {
  policy_id = "__default_policy_ID"

  statement {
    sid = "inspector"

    principals {
      type = "AWS"

      identifiers = [
		"arn:aws:iam::162588757376:root",
		"arn:aws:iam::526946625049:root",
		"arn:aws:iam::454640832652:root",
		"arn:aws:iam::406045910587:root",
		"arn:aws:iam::537503971621:root",
		"arn:aws:iam::357557129151:root",
		"arn:aws:iam::316112463485:root",
		"arn:aws:iam::646659390643:root",
		"arn:aws:iam::166987590008:root",
		"arn:aws:iam::758058086616:root"
      ]
    }

    actions = [
      "SNS:Subscribe",
      "SNS:Receive",
      "SNS:Publish",
    ]

    effect    = "Allow"
    resources = ["${aws_sns_topic.test_sns_topic_for_inspector.arn}"]
  }
}
`
