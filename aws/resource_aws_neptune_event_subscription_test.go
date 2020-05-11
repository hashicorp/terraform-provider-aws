package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSNeptuneEventSubscription_basic(t *testing.T) {
	var v neptune.EventSubscription
	rInt := acctest.RandInt()
	rName := fmt.Sprintf("tf-acc-test-neptune-event-subs-%d", rInt)

	resourceName := "aws_neptune_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneEventSubscriptionConfig(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneEventSubscriptionExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "rds", fmt.Sprintf("es:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-instance"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "tf-acc-test"),
				),
			},
			{
				Config: testAccAWSNeptuneEventSubscriptionConfigUpdate(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-parameter-group"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "tf-acc-test1"),
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

func TestAccAWSNeptuneEventSubscription_withPrefix(t *testing.T) {
	var v neptune.EventSubscription
	rInt := acctest.RandInt()
	startsWithPrefix := regexp.MustCompile("^tf-acc-test-neptune-event-subs-")

	resourceName := "aws_neptune_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneEventSubscriptionConfigWithPrefix(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneEventSubscriptionExists(resourceName, &v),
					resource.TestMatchResourceAttr(resourceName, "name", startsWithPrefix),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccAWSNeptuneEventSubscription_withSourceIds(t *testing.T) {
	var v neptune.EventSubscription
	rInt := acctest.RandInt()

	resourceName := "aws_neptune_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneEventSubscriptionConfigWithSourceIds(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-parameter-group"),
					resource.TestCheckResourceAttr(resourceName, "source_ids.#", "1"),
				),
			},
			{
				Config: testAccAWSNeptuneEventSubscriptionConfigUpdateSourceIds(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-parameter-group"),
					resource.TestCheckResourceAttr(resourceName, "source_ids.#", "2"),
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

func TestAccAWSNeptuneEventSubscription_withCategories(t *testing.T) {
	var v neptune.EventSubscription
	rInt := acctest.RandInt()
	rName := fmt.Sprintf("tf-acc-test-neptune-event-subs-%d", rInt)

	resourceName := "aws_neptune_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneEventSubscriptionConfig(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-instance"),
					resource.TestCheckResourceAttr(resourceName, "event_categories.#", "5"),
				),
			},
			{
				Config: testAccAWSNeptuneEventSubscriptionConfigUpdateCategories(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-instance"),
					resource.TestCheckResourceAttr(resourceName, "event_categories.#", "1"),
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

func testAccAWSNeptuneEventSubscriptionConfig(subscriptionName string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-neptune-event-subs-sns-topic-%[1]d"
}

resource "aws_neptune_event_subscription" "test" {
  name          = %[2]q
  sns_topic_arn = "${aws_sns_topic.aws_sns_topic.arn}"
  source_type   = "db-instance"

  event_categories = [
    "availability",
    "backup",
    "creation",
    "deletion",
    "maintenance",
  ]

  tags = {
    Name = "tf-acc-test"
  }
}
`, rInt, subscriptionName)
}

func testAccAWSNeptuneEventSubscriptionConfigUpdate(subscriptionName string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-neptune-event-subs-sns-topic-%[1]d"
}

resource "aws_neptune_event_subscription" "test" {
  name          = %[2]q
  sns_topic_arn = "${aws_sns_topic.aws_sns_topic.arn}"
  enabled       = false
  source_type   = "db-parameter-group"

  event_categories = [
    "configuration change",
  ]

  tags = {
    Name = "tf-acc-test1"
  }
}
`, rInt, subscriptionName)
}

func testAccAWSNeptuneEventSubscriptionConfigWithPrefix(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-neptune-event-subs-sns-topic-%d"
}

resource "aws_neptune_event_subscription" "test" {
  name_prefix   = "tf-acc-test-neptune-event-subs-"
  sns_topic_arn = "${aws_sns_topic.aws_sns_topic.arn}"
  source_type   = "db-instance"

  event_categories = [
    "availability",
    "backup",
    "creation",
    "deletion",
    "maintenance",
  ]

  tags = {
    Name = "tf-acc-test"
  }
}
`, rInt)
}

func testAccAWSNeptuneEventSubscriptionConfigWithSourceIds(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-neptune-event-subs-sns-topic-%[1]d"
}

resource "aws_neptune_parameter_group" "test" {
  name        = "neptune-parameter-group-event-%[1]d"
  family      = "neptune1"
  description = "Test parameter group for terraform"
}

resource "aws_neptune_event_subscription" "test" {
  name          = "tf-acc-test-neptune-event-subs-with-ids-%[1]d"
  sns_topic_arn = "${aws_sns_topic.aws_sns_topic.arn}"
  source_type   = "db-parameter-group"
  source_ids    = ["${aws_neptune_parameter_group.test.id}"]

  event_categories = [
    "configuration change",
  ]

  tags = {
    Name = "tf-acc-test"
  }
}
`, rInt)
}

func testAccAWSNeptuneEventSubscriptionConfigUpdateSourceIds(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-neptune-event-subs-sns-topic-%[1]d"
}

resource "aws_neptune_parameter_group" "test" {
  name        = "neptune-parameter-group-event-%[1]d"
  family      = "neptune1"
  description = "Test parameter group for terraform"
}

resource "aws_neptune_parameter_group" "test_2" {
  name        = "neptune-parameter-group-event-2-%[1]d"
  family      = "neptune1"
  description = "Test parameter group for terraform"
}

resource "aws_neptune_event_subscription" "test" {
  name          = "tf-acc-test-neptune-event-subs-with-ids-%[1]d"
  sns_topic_arn = "${aws_sns_topic.aws_sns_topic.arn}"
  source_type   = "db-parameter-group"
  source_ids    = ["${aws_neptune_parameter_group.test.id}", "${aws_neptune_parameter_group.test_2.id}"]

  event_categories = [
    "configuration change",
  ]

  tags = {
    Name = "tf-acc-test"
  }
}
`, rInt)
}

func testAccAWSNeptuneEventSubscriptionConfigUpdateCategories(subscriptionName string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-neptune-event-subs-sns-topic-%[1]d"
}

resource "aws_neptune_event_subscription" "test" {
  name          = %[2]q
  sns_topic_arn = "${aws_sns_topic.aws_sns_topic.arn}"
  source_type   = "db-instance"

  event_categories = [
    "availability",
  ]

  tags = {
    Name = "tf-acc-test"
  }
}
`, rInt, subscriptionName)
}
