package redshift_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccRedshiftEventSubscription_basicUpdate(t *testing.T) {
	var v redshift.EventSubscription
	rInt := sdkacctest.RandInt()
	rName := fmt.Sprintf("tf-acc-test-redshift-event-subs-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists("aws_redshift_event_subscription.bar", &v),
					resource.TestCheckResourceAttr("aws_redshift_event_subscription.bar", "enabled", "true"),
					resource.TestCheckResourceAttr("aws_redshift_event_subscription.bar", "source_type", "cluster"),
					resource.TestCheckResourceAttr("aws_redshift_event_subscription.bar", "name", rName),
					resource.TestCheckResourceAttr("aws_redshift_event_subscription.bar", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_redshift_event_subscription.bar", "tags.Name", "name"),
				),
			},
			{
				Config: testAccEventSubscriptionUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists("aws_redshift_event_subscription.bar", &v),
					resource.TestCheckResourceAttr("aws_redshift_event_subscription.bar", "enabled", "false"),
					resource.TestCheckResourceAttr("aws_redshift_event_subscription.bar", "source_type", "cluster-snapshot"),
					resource.TestCheckResourceAttr("aws_redshift_event_subscription.bar", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_redshift_event_subscription.bar", "tags.Name", "new-name"),
				),
			},
			{
				ResourceName:      "aws_redshift_event_subscription.bar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRedshiftEventSubscription_withPrefix(t *testing.T) {
	var v redshift.EventSubscription
	rInt := sdkacctest.RandInt()
	rName := fmt.Sprintf("tf-acc-test-redshift-event-subs-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists("aws_redshift_event_subscription.bar", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "source_type", "cluster"),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "name", rName),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "tags.Name", "name"),
				),
			},
			{
				ResourceName:      "aws_redshift_event_subscription.bar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRedshiftEventSubscription_withSourceIDs(t *testing.T) {
	var v redshift.EventSubscription
	rInt := sdkacctest.RandInt()
	rName := fmt.Sprintf("tf-acc-test-redshift-event-subs-with-ids-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionWithSourceIDsConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists("aws_redshift_event_subscription.bar", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "source_type", "cluster-parameter-group"),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "name", rName),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "source_ids.#", "1"),
				),
			},
			{
				Config: testAccEventSubscriptionUpdateSourceIDsConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists("aws_redshift_event_subscription.bar", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "source_type", "cluster-parameter-group"),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "name", rName),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "source_ids.#", "2"),
				),
			},
			{
				ResourceName:      "aws_redshift_event_subscription.bar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRedshiftEventSubscription_categoryUpdate(t *testing.T) {
	var v redshift.EventSubscription
	rInt := sdkacctest.RandInt()
	rName := fmt.Sprintf("tf-acc-test-redshift-event-subs-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists("aws_redshift_event_subscription.bar", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "source_type", "cluster"),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "name", rName),
				),
			},
			{
				Config: testAccEventSubscriptionUpdateCategoriesConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists("aws_redshift_event_subscription.bar", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "source_type", "cluster"),
				),
			},
			{
				ResourceName:      "aws_redshift_event_subscription.bar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRedshiftEventSubscription_tagsUpdate(t *testing.T) {
	var v redshift.EventSubscription
	rInt := sdkacctest.RandInt()
	resourceName := "aws_redshift_event_subscription.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "name"),
				),
			},
			{
				Config: testAccEventSubscriptionUpdateTagsConfig(rInt, "aaaaa"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "name"),
					resource.TestCheckResourceAttr(resourceName, "tags.Test", "aaaaa"),
				),
			},
			{
				Config: testAccEventSubscriptionUpdateTagsConfig(rInt, "bbbbb"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "name"),
					resource.TestCheckResourceAttr(resourceName, "tags.Test", "bbbbb"),
				),
			},
			{
				Config: testAccEventSubscriptionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "name"),
				),
			},
		},
	})
}

func testAccCheckEventSubscriptionExists(n string, v *redshift.EventSubscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Redshift Event Subscription is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn

		opts := redshift.DescribeEventSubscriptionsInput{
			SubscriptionName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeEventSubscriptions(&opts)

		if err != nil {
			return err
		}

		if len(resp.EventSubscriptionsList) != 1 ||
			*resp.EventSubscriptionsList[0].CustSubscriptionId != rs.Primary.ID {
			return fmt.Errorf("Redshift Event Subscription not found")
		}

		*v = *resp.EventSubscriptionsList[0]
		return nil
	}
}

func testAccCheckEventSubscriptionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_redshift_event_subscription" {
			continue
		}

		var err error
		resp, err := conn.DescribeEventSubscriptions(
			&redshift.DescribeEventSubscriptionsInput{
				SubscriptionName: aws.String(rs.Primary.ID),
			})

		if tfawserr.ErrCodeEquals(err, redshift.ErrCodeSubscriptionNotFoundFault) {
			continue
		}

		if err != nil {
			return err
		}

		if len(resp.EventSubscriptionsList) != 0 &&
			*resp.EventSubscriptionsList[0].CustSubscriptionId == rs.Primary.ID {
			return fmt.Errorf("Event Subscription still exists")
		}
	}

	return nil
}

func testAccEventSubscriptionConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-redshift-event-subs-sns-topic-%d"
}

resource "aws_redshift_event_subscription" "bar" {
  name          = "tf-acc-test-redshift-event-subs-%d"
  sns_topic_arn = aws_sns_topic.aws_sns_topic.arn
  source_type   = "cluster"
  severity      = "INFO"

  event_categories = [
    "configuration",
    "management",
    "monitoring",
    "security",
  ]

  tags = {
    Name = "name"
  }
}
`, rInt, rInt)
}

func testAccEventSubscriptionUpdateConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-redshift-event-subs-sns-topic-%d"
}

resource "aws_redshift_event_subscription" "bar" {
  name          = "tf-acc-test-redshift-event-subs-%d"
  sns_topic_arn = aws_sns_topic.aws_sns_topic.arn
  enabled       = false
  source_type   = "cluster-snapshot"
  severity      = "INFO"

  event_categories = [
    "monitoring",
  ]

  tags = {
    Name = "new-name"
  }
}
`, rInt, rInt)
}

func testAccEventSubscriptionWithSourceIDsConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-redshift-event-subs-sns-topic-%d"
}

resource "aws_redshift_parameter_group" "bar" {
  name        = "redshift-parameter-group-event-%d"
  family      = "redshift-1.0"
  description = "Test parameter group for terraform"
}

resource "aws_redshift_event_subscription" "bar" {
  name          = "tf-acc-test-redshift-event-subs-with-ids-%d"
  sns_topic_arn = aws_sns_topic.aws_sns_topic.arn
  source_type   = "cluster-parameter-group"
  severity      = "INFO"
  source_ids    = [aws_redshift_parameter_group.bar.id]

  event_categories = [
    "configuration",
  ]

  tags = {
    Name = "name"
  }
}
`, rInt, rInt, rInt)
}

func testAccEventSubscriptionUpdateSourceIDsConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-redshift-event-subs-sns-topic-%d"
}

resource "aws_redshift_parameter_group" "bar" {
  name        = "tf-acc-redshift-parameter-group-event-%d"
  family      = "redshift-1.0"
  description = "Test parameter group for terraform"
}

resource "aws_redshift_parameter_group" "foo" {
  name        = "tf-acc-redshift-parameter-group-event-2-%d"
  family      = "redshift-1.0"
  description = "Test parameter group for terraform"
}

resource "aws_redshift_event_subscription" "bar" {
  name          = "tf-acc-test-redshift-event-subs-with-ids-%d"
  sns_topic_arn = aws_sns_topic.aws_sns_topic.arn
  source_type   = "cluster-parameter-group"
  severity      = "INFO"
  source_ids    = [aws_redshift_parameter_group.bar.id, aws_redshift_parameter_group.foo.id]

  event_categories = [
    "configuration",
  ]

  tags = {
    Name = "name"
  }
}
`, rInt, rInt, rInt, rInt)
}

func testAccEventSubscriptionUpdateCategoriesConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-redshift-event-subs-sns-topic-%d"
}

resource "aws_redshift_event_subscription" "bar" {
  name          = "tf-acc-test-redshift-event-subs-%d"
  sns_topic_arn = aws_sns_topic.aws_sns_topic.arn
  source_type   = "cluster"
  severity      = "INFO"

  event_categories = [
    "monitoring",
  ]

  tags = {
    Name = "name"
  }
}
`, rInt, rInt)
}

func testAccEventSubscriptionUpdateTagsConfig(rInt int, rString string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-redshift-event-subs-sns-topic-%d"
}

resource "aws_redshift_event_subscription" "bar" {
  name          = "tf-acc-test-redshift-event-subs-%d"
  sns_topic_arn = aws_sns_topic.aws_sns_topic.arn
  source_type   = "cluster"
  severity      = "INFO"

  event_categories = [
    "configuration",
    "management",
    "monitoring",
    "security",
  ]

  tags = {
    Name = "name"
    Test = "%s"
  }
}
`, rInt, rInt, rString)
}
