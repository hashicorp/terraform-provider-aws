package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/rds/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_db_event_subscription", &resource.Sweeper{
		Name: "aws_db_event_subscription",
		F:    testSweepDbEventSubscriptions,
	})
}

func testSweepDbEventSubscriptions(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).rdsconn
	input := &rds.DescribeEventSubscriptionsInput{}
	sweepResources := make([]*testSweepResource, 0)

	err = conn.DescribeEventSubscriptionsPages(input, func(page *rds.DescribeEventSubscriptionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, eventSubscription := range page.EventSubscriptionsList {
			r := resourceAwsDbEventSubscription()
			d := r.Data(nil)
			d.SetId(aws.StringValue(eventSubscription.CustSubscriptionId))

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping RDS Event Subscription sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing RDS Event Subscriptions (%s): %w", region, err)
	}

	err = testSweepResourceOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping RDS Event Subscriptions (%s): %w", region, err)
	}

	return nil
}

func TestAccAWSDBEventSubscription_basic(t *testing.T) {
	var v rds.EventSubscription
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_db_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBEventSubscriptionConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "rds", fmt.Sprintf("es:%s", rName)),
					testAccCheckResourceAttrAccountID(resourceName, "customer_aws_id"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "event_categories.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "source_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_type", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSDBEventSubscription_disappears(t *testing.T) {
	var eventSubscription rds.EventSubscription
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_db_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBEventSubscriptionConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists(resourceName, &eventSubscription),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsDbEventSubscription(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSDBEventSubscription_Name_Generated(t *testing.T) {
	var v rds.EventSubscription
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_db_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBEventSubscriptionConfigNameGenerated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists(resourceName, &v),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
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

func TestAccAWSDBEventSubscription_NamePrefix(t *testing.T) {
	var v rds.EventSubscription
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_db_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBEventSubscriptionConfigNamePrefix(rName, "tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists(resourceName, &v),
					naming.TestCheckResourceAttrNameFromPrefix(resourceName, "name", "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "tf-acc-test-prefix-"),
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

func TestAccAWSDBEventSubscription_Tags(t *testing.T) {
	var eventSubscription rds.EventSubscription
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_db_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBEventSubscriptionConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists(resourceName, &eventSubscription),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDBEventSubscriptionConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists(resourceName, &eventSubscription),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSDBEventSubscriptionConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists(resourceName, &eventSubscription),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSDBEventSubscription_Categories(t *testing.T) {
	var v rds.EventSubscription
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_db_event_subscription.test"
	snsTopicResourceName := "aws_sns_topic.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBEventSubscriptionConfigCategories(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "rds", fmt.Sprintf("es:%s", rName)),
					testAccCheckResourceAttrAccountID(resourceName, "customer_aws_id"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "event_categories.#", "5"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "availability"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "backup"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "creation"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "deletion"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "maintenance"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
					resource.TestCheckResourceAttrPair(resourceName, "sns_topic", snsTopicResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "source_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-instance"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDBEventSubscriptionConfigCategoriesUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "rds", fmt.Sprintf("es:%s", rName)),
					testAccCheckResourceAttrAccountID(resourceName, "customer_aws_id"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "event_categories.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "creation"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
					resource.TestCheckResourceAttrPair(resourceName, "sns_topic", snsTopicResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "source_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-cluster"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSDBEventSubscription_Update(t *testing.T) {
	var v rds.EventSubscription
	rInt := acctest.RandInt()
	resourceName := "aws_db_event_subscription.test"
	subscriptionName := fmt.Sprintf("tf-acc-test-rds-event-subs-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBEventSubscriptionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "rds", regexp.MustCompile(fmt.Sprintf("es:%s$", subscriptionName))),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-instance"),
					resource.TestCheckResourceAttr(resourceName, "name", subscriptionName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     subscriptionName,
			},
			{
				Config: testAccAWSDBEventSubscriptionConfigUpdate(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-parameter-group"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "new-name"),
				),
			},
		},
	})
}

func TestAccAWSDBEventSubscription_withSourceIds(t *testing.T) {
	var v rds.EventSubscription
	rInt := acctest.RandInt()
	resourceName := "aws_db_event_subscription.test"
	subscriptionName := fmt.Sprintf("tf-acc-test-rds-event-subs-with-ids-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBEventSubscriptionConfigWithSourceIds(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-parameter-group"),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tf-acc-test-rds-event-subs-with-ids-%d", rInt)),
					resource.TestCheckResourceAttr(resourceName, "source_ids.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     subscriptionName,
			},
			{
				Config: testAccAWSDBEventSubscriptionConfigUpdateSourceIds(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-parameter-group"),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tf-acc-test-rds-event-subs-with-ids-%d", rInt)),
					resource.TestCheckResourceAttr(resourceName, "source_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSDBEventSubscription_categoryUpdate(t *testing.T) {
	var v rds.EventSubscription
	rInt := acctest.RandInt()
	resourceName := "aws_db_event_subscription.test"
	subscriptionName := fmt.Sprintf("tf-acc-test-rds-event-subs-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBEventSubscriptionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-instance"),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tf-acc-test-rds-event-subs-%d", rInt)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     subscriptionName,
			},
			{
				Config: testAccAWSDBEventSubscriptionConfigUpdateCategories(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-instance"),
				),
			},
		},
	})
}

func testAccCheckAWSDBEventSubscriptionExists(n string, v *rds.EventSubscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No RDS Event Subscription is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).rdsconn

		output, err := finder.EventSubscriptionByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAWSDBEventSubscriptionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).rdsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_event_subscription" {
			continue
		}

		_, err := finder.EventSubscriptionByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("RDS Event Subscription %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccAWSDBEventSubscriptionConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-rds-event-subs-sns-topic-%[1]d"
}

resource "aws_db_event_subscription" "test" {
  name        = "tf-acc-test-rds-event-subs-%[1]d"
  sns_topic   = aws_sns_topic.aws_sns_topic.arn
  source_type = "db-instance"

  event_categories = [
    "availability",
    "backup",
    "creation",
    "deletion",
    "maintenance",
  ]

  tags = {
    Name = "name"
  }
}
`, rInt)
}

func testAccAWSDBEventSubscriptionConfigBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_db_event_subscription" "test" {
  name      = %[1]q
  sns_topic = aws_sns_topic.test.arn
}
`, rName)
}

func testAccAWSDBEventSubscriptionConfigNameGenerated(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_db_event_subscription" "test" {
  sns_topic = aws_sns_topic.test.arn
}
`, rName)
}

func testAccAWSDBEventSubscriptionConfigNamePrefix(rName, namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_db_event_subscription" "test" {
  name_prefix = %[2]q
  sns_topic   = aws_sns_topic.test.arn
}
`, rName, namePrefix)
}

func testAccAWSDBEventSubscriptionConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_db_event_subscription" "test" {
  name        = %[1]q
  sns_topic   = aws_sns_topic.test.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSDBEventSubscriptionConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_db_event_subscription" "test" {
  name        = %[1]q
  sns_topic   = aws_sns_topic.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSDBEventSubscriptionConfigCategories(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_db_event_subscription" "test" {
  name        = %[1]q
  sns_topic   = aws_sns_topic.test.arn
  enabled     = false
  source_type = "db-instance"

  event_categories = [
    "availability",
    "backup",
    "creation",
    "deletion",
    "maintenance",
  ]
}
`, rName)
}

func testAccAWSDBEventSubscriptionConfigCategoriesUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_db_event_subscription" "test" {
  name        = %[1]q
  sns_topic   = aws_sns_topic.test.arn
  enabled     = true
  source_type = "db-cluster"

  event_categories = [
    "creation",
  ]
}
`, rName)
}

func testAccAWSDBEventSubscriptionConfigUpdate(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-rds-event-subs-sns-topic-%[1]d"
}

resource "aws_db_event_subscription" "test" {
  name        = "tf-acc-test-rds-event-subs-%[1]d"
  sns_topic   = aws_sns_topic.aws_sns_topic.arn
  enabled     = false
  source_type = "db-parameter-group"

  event_categories = [
    "configuration change",
  ]

  tags = {
    Name = "new-name"
  }
}
`, rInt)
}

func testAccAWSDBEventSubscriptionConfigWithSourceIds(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-rds-event-subs-sns-topic-%[1]d"
}

resource "aws_db_parameter_group" "test" {
  name        = "db-parameter-group-event-%[1]d"
  family      = "mysql5.6"
  description = "Test parameter group for terraform"
}

resource "aws_db_event_subscription" "test" {
  name        = "tf-acc-test-rds-event-subs-with-ids-%[1]d"
  sns_topic   = aws_sns_topic.aws_sns_topic.arn
  source_type = "db-parameter-group"
  source_ids  = [aws_db_parameter_group.test.id]

  event_categories = [
    "configuration change",
  ]

  tags = {
    Name = "name"
  }
}
`, rInt)
}

func testAccAWSDBEventSubscriptionConfigUpdateSourceIds(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-rds-event-subs-sns-topic-%[1]d"
}

resource "aws_db_parameter_group" "test" {
  name        = "db-parameter-group-event-%[1]d"
  family      = "mysql5.6"
  description = "Test parameter group for terraform"
}

resource "aws_db_parameter_group" "test2" {
  name        = "db-parameter-group-event-2-%[1]d"
  family      = "mysql5.6"
  description = "Test parameter group for terraform"
}

resource "aws_db_event_subscription" "test" {
  name        = "tf-acc-test-rds-event-subs-with-ids-%[1]d"
  sns_topic   = aws_sns_topic.aws_sns_topic.arn
  source_type = "db-parameter-group"
  source_ids  = [aws_db_parameter_group.test.id, aws_db_parameter_group.test2.id]

  event_categories = [
    "configuration change",
  ]

  tags = {
    Name = "name"
  }
}
`, rInt)
}

func testAccAWSDBEventSubscriptionConfigUpdateCategories(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-rds-event-subs-sns-topic-%[1]d"
}

resource "aws_db_event_subscription" "test" {
  name        = "tf-acc-test-rds-event-subs-%[1]d"
  sns_topic   = aws_sns_topic.aws_sns_topic.arn
  source_type = "db-instance"

  event_categories = [
    "availability",
  ]

  tags = {
    Name = "name"
  }
}
`, rInt)
}
