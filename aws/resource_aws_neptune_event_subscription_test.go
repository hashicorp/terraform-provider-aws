package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSNeptuneEventSubscription_basic(t *testing.T) {
	var v neptune.EventSubscription
	rInt := acctest.RandInt()
	rName := fmt.Sprintf("tf-acc-test-neptune-event-subs-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneEventSubscriptionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneEventSubscriptionExists("aws_neptune_event_subscription.bar", &v),
					resource.TestMatchResourceAttr("aws_neptune_event_subscription.bar", "arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:rds:[^:]+:[^:]+:es:%s$", rName))),
					resource.TestCheckResourceAttr("aws_neptune_event_subscription.bar", "enabled", "true"),
					resource.TestCheckResourceAttr("aws_neptune_event_subscription.bar", "source_type", "db-instance"),
					resource.TestCheckResourceAttr("aws_neptune_event_subscription.bar", "name", rName),
					resource.TestCheckResourceAttr("aws_neptune_event_subscription.bar", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_neptune_event_subscription.bar", "tags.Name", "tf-acc-test"),
				),
			},
			{
				Config: testAccAWSNeptuneEventSubscriptionConfigUpdate(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneEventSubscriptionExists("aws_neptune_event_subscription.bar", &v),
					resource.TestCheckResourceAttr("aws_neptune_event_subscription.bar", "enabled", "false"),
					resource.TestCheckResourceAttr("aws_neptune_event_subscription.bar", "source_type", "db-parameter-group"),
					resource.TestCheckResourceAttr("aws_neptune_event_subscription.bar", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_neptune_event_subscription.bar", "tags.Name", "tf-acc-test1"),
				),
			},
		},
	})
}

func testAccCheckAWSNeptuneEventSubscriptionExists(n string, v *neptune.EventSubscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Neptune Event Subscription is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).neptuneconn

		opts := neptune.DescribeEventSubscriptionsInput{
			SubscriptionName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeEventSubscriptions(&opts)

		if err != nil {
			return err
		}

		if len(resp.EventSubscriptionsList) != 1 ||
			aws.StringValue(resp.EventSubscriptionsList[0].CustSubscriptionId) != rs.Primary.ID {
			return fmt.Errorf("Neptune Event Subscription not found")
		}

		*v = *resp.EventSubscriptionsList[0]
		return nil
	}
}

func testAccCheckAWSNeptuneEventSubscriptionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).neptuneconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_neptune_event_subscription" {
			continue
		}

		var err error
		resp, err := conn.DescribeEventSubscriptions(
			&neptune.DescribeEventSubscriptionsInput{
				SubscriptionName: aws.String(rs.Primary.ID),
			})

		if ae, ok := err.(awserr.Error); ok && ae.Code() == "SubscriptionNotFound" {
			continue
		}

		if err == nil {
			if len(resp.EventSubscriptionsList) != 0 &&
				aws.StringValue(resp.EventSubscriptionsList[0].CustSubscriptionId) == rs.Primary.ID {
				return fmt.Errorf("Event Subscription still exists")
			}
		}

		newerr, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if newerr.Code() != "SubscriptionNotFound" {
			return err
		}
	}

	return nil
}

func testAccAWSNeptuneEventSubscriptionConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-neptune-event-subs-sns-topic-%d"
}

resource "aws_neptune_event_subscription" "bar" {
  name = "tf-acc-test-neptune-event-subs-%d"
  sns_topic_arn = "${aws_sns_topic.aws_sns_topic.arn}"
  source_type = "db-instance"
  event_categories = [
    "availability",
    "backup",
    "creation",
    "deletion",
    "maintenance"
  ]
  tags {
    Name = "tf-acc-test"
  }
}`, rInt, rInt)
}

func testAccAWSNeptuneEventSubscriptionConfigUpdate(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-neptune-event-subs-sns-topic-%d"
}

resource "aws_neptune_event_subscription" "bar" {
  name = "tf-acc-test-neptune-event-subs-%d"
  sns_topic = "${aws_sns_topic.aws_sns_topic.arn}"
  enabled = false
  source_type = "db-parameter-group"
  event_categories = [
    "configuration change"
  ]
  tags {
    Name = "tf-acc-test1"
  }
}`, rInt, rInt)
}
