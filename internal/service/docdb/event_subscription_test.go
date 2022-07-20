package docdb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/docdb"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfdocdb "github.com/hashicorp/terraform-provider-aws/internal/service/docdb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccDocDBEventSubscription_basic(t *testing.T) {
	var eventSubscription docdb.EventSubscription
	resourceName := "aws_docdb_event_subscription.test"
	snsTopicResourceName := "aws_sns_topic.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig_enabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &eventSubscription),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "event_categories.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "creation"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "failure"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "source_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-cluster"),
					resource.TestCheckResourceAttrPair(resourceName, "sns_topic_arn", snsTopicResourceName, "arn"),
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

func TestAccDocDBEventSubscription_nameGenerated(t *testing.T) {
	var eventSubscription docdb.EventSubscription
	resourceName := "aws_docdb_event_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig_nameGenerated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &eventSubscription),
					create.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", resource.UniqueIdPrefix),
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

func TestAccDocDBEventSubscription_disappears(t *testing.T) {
	var eventSubscription docdb.EventSubscription
	resourceName := "aws_docdb_event_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig_enabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &eventSubscription),
					acctest.CheckResourceDisappears(acctest.Provider, tfdocdb.ResourceEventSubscription(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDocDBEventSubscription_enabled(t *testing.T) {
	var eventSubscription docdb.EventSubscription
	resourceName := "aws_docdb_event_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig_enabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &eventSubscription),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEventSubscriptionConfig_enabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &eventSubscription),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
			{
				Config: testAccEventSubscriptionConfig_enabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &eventSubscription),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
		},
	})
}

func TestAccDocDBEventSubscription_eventCategories(t *testing.T) {
	var eventSubscription docdb.EventSubscription
	resourceName := "aws_docdb_event_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig_categories2(rName, "creation", "failure"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &eventSubscription),
					resource.TestCheckResourceAttr(resourceName, "event_categories.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "creation"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "failure"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEventSubscriptionConfig_categories2(rName, "configuration change", "deletion"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &eventSubscription),
					resource.TestCheckResourceAttr(resourceName, "event_categories.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "configuration change"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "deletion"),
				),
			},
		},
	})
}

func TestAccDocDBEventSubscription_tags(t *testing.T) {
	var eventSubscription docdb.EventSubscription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_docdb_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &eventSubscription),
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
				Config: testAccEventSubscriptionConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &eventSubscription),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccEventSubscriptionConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &eventSubscription),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckEventSubscriptionDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_docdb_event_subscription" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DocDBConn

		_, err := tfdocdb.FindEventSubscriptionByID(context.TODO(), conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("DocDB Event Subscription %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckEventSubscriptionExists(n string, eventSubscription *docdb.EventSubscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DocDB Event Subscription ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DocDBConn

		res, err := tfdocdb.FindEventSubscriptionByID(context.TODO(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*eventSubscription = *res

		return nil
	}
}

func testAccEventSubscriptionBaseConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_docdb_cluster" "test" {
  cluster_identifier  = %[1]q
  availability_zones  = [data.aws_availability_zones.available.names[0], data.aws_availability_zones.available.names[1], data.aws_availability_zones.available.names[2]]
  master_username     = "foo"
  master_password     = "mustbeeightcharaters"
  skip_final_snapshot = true
}

data "aws_docdb_orderable_db_instance" "test" {
  engine                     = "docdb"
  preferred_instance_classes = ["db.t3.medium", "db.r4.large", "db.r5.large", "db.r5.xlarge"]
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}
`, rName))
}

func testAccEventSubscriptionConfig_enabled(rName string, enabled bool) string {
	return acctest.ConfigCompose(
		testAccEventSubscriptionBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_docdb_event_subscription" "test" {
  name             = %[1]q
  enabled          = %[2]t
  event_categories = ["creation", "failure"]
  source_type      = "db-cluster"
  source_ids       = [aws_docdb_cluster.test.id]
  sns_topic_arn    = aws_sns_topic.test.arn
}
`, rName, enabled))
}

func testAccEventSubscriptionConfig_nameGenerated(rName string) string {
	return acctest.ConfigCompose(
		testAccEventSubscriptionBaseConfig(rName),
		`
resource "aws_docdb_event_subscription" "test" {
  enabled          = true
  event_categories = ["creation", "failure"]
  source_type      = "db-cluster"
  source_ids       = [aws_docdb_cluster.test.id]
  sns_topic_arn    = aws_sns_topic.test.arn
}
`)
}

func testAccEventSubscriptionConfig_categories2(rName string, eventCategory1 string, eventCategory2 string) string {
	return acctest.ConfigCompose(
		testAccEventSubscriptionBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_docdb_event_subscription" "test" {
  name             = %[1]q
  enabled          = false
  event_categories = [%[2]q, %[3]q]
  source_type      = "db-cluster"
  source_ids       = [aws_docdb_cluster.test.id]
  sns_topic_arn    = aws_sns_topic.test.arn
}
`, rName, eventCategory1, eventCategory2))
}

func testAccEventSubscriptionConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccEventSubscriptionBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_docdb_event_subscription" "test" {
  name             = %[1]q
  enabled          = true
  event_categories = ["creation", "failure"]
  source_type      = "db-cluster"
  source_ids       = [aws_docdb_cluster.test.id]
  sns_topic_arn    = aws_sns_topic.test.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccEventSubscriptionConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccEventSubscriptionBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_docdb_event_subscription" "test" {
  name             = %[1]q
  enabled          = true
  event_categories = ["creation", "failure"]
  source_type      = "db-cluster"
  source_ids       = [aws_docdb_cluster.test.id]
  sns_topic_arn    = aws_sns_topic.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
