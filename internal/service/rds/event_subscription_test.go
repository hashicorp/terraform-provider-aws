package rds_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/rds"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccRDSEventSubscription_basic(t *testing.T) {
	var v rds.EventSubscription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "rds", fmt.Sprintf("es:%s", rName)),
					acctest.CheckResourceAttrAccountID(resourceName, "customer_aws_id"),
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

func TestAccRDSEventSubscription_disappears(t *testing.T) {
	var eventSubscription rds.EventSubscription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &eventSubscription),
					acctest.CheckResourceDisappears(acctest.Provider, tfrds.ResourceEventSubscription(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRDSEventSubscription_Name_generated(t *testing.T) {
	var v rds.EventSubscription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionNameGeneratedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &v),
					create.TestCheckResourceAttrNameGenerated(resourceName, "name"),
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

func TestAccRDSEventSubscription_namePrefix(t *testing.T) {
	var v rds.EventSubscription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionNamePrefixConfig(rName, "tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &v),
					create.TestCheckResourceAttrNameFromPrefix(resourceName, "name", "tf-acc-test-prefix-"),
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

func TestAccRDSEventSubscription_tags(t *testing.T) {
	var eventSubscription rds.EventSubscription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionTags1Config(rName, "key1", "value1"),
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
				Config: testAccEventSubscriptionTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &eventSubscription),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccEventSubscriptionTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &eventSubscription),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccRDSEventSubscription_categories(t *testing.T) {
	var v rds.EventSubscription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_event_subscription.test"
	snsTopicResourceName := "aws_sns_topic.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionCategoriesConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "rds", fmt.Sprintf("es:%s", rName)),
					acctest.CheckResourceAttrAccountID(resourceName, "customer_aws_id"),
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
				Config: testAccEventSubscriptionCategoriesUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "rds", fmt.Sprintf("es:%s", rName)),
					acctest.CheckResourceAttrAccountID(resourceName, "customer_aws_id"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "event_categories.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "creation"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
					resource.TestCheckResourceAttrPair(resourceName, "sns_topic", snsTopicResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "source_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-instance"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccEventSubscriptionCategoriesAndSourceTypeUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "rds", fmt.Sprintf("es:%s", rName)),
					acctest.CheckResourceAttrAccountID(resourceName, "customer_aws_id"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "event_categories.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "creation"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "deletion"),
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

func TestAccRDSEventSubscription_sourceIDs(t *testing.T) {
	var v rds.EventSubscription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_event_subscription.test"
	paramGroup1ResourceName := "aws_db_parameter_group.test1"
	paramGroup2ResourceName := "aws_db_parameter_group.test2"
	paramGroup3ResourceName := "aws_db_parameter_group.test3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionSourceIDsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "source_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "source_ids.*", paramGroup1ResourceName, "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "source_ids.*", paramGroup2ResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEventSubscriptionSourceIDsUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "source_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "source_ids.*", paramGroup3ResourceName, "id"),
				),
			},
		},
	})
}

func testAccCheckEventSubscriptionExists(n string, v *rds.EventSubscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No RDS Event Subscription is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

		output, err := tfrds.FindEventSubscriptionByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckEventSubscriptionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_event_subscription" {
			continue
		}

		_, err := tfrds.FindEventSubscriptionByID(conn, rs.Primary.ID)

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

func testAccEventSubscriptionBasicConfig(rName string) string {
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

func testAccEventSubscriptionNameGeneratedConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_db_event_subscription" "test" {
  sns_topic = aws_sns_topic.test.arn
}
`, rName)
}

func testAccEventSubscriptionNamePrefixConfig(rName, namePrefix string) string {
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

func testAccEventSubscriptionTags1Config(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_db_event_subscription" "test" {
  name      = %[1]q
  sns_topic = aws_sns_topic.test.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccEventSubscriptionTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_db_event_subscription" "test" {
  name      = %[1]q
  sns_topic = aws_sns_topic.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccEventSubscriptionCategoriesConfig(rName string) string {
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

func testAccEventSubscriptionCategoriesUpdatedConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_db_event_subscription" "test" {
  name        = %[1]q
  sns_topic   = aws_sns_topic.test.arn
  enabled     = true
  source_type = "db-instance"

  event_categories = [
    "creation",
  ]
}
`, rName)
}

func testAccEventSubscriptionCategoriesAndSourceTypeUpdatedConfig(rName string) string {
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
    "deletion",
  ]
}
`, rName)
}

func testAccEventSubscriptionSourceIDsBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_db_parameter_group" "test1" {
  name   = "%[1]s-1"
  family = "mysql5.6"
}

resource "aws_db_parameter_group" "test2" {
  name   = "%[1]s-2"
  family = "mysql5.6"
}

resource "aws_db_parameter_group" "test3" {
  name   = "%[1]s-3"
  family = "mysql5.6"
}
`, rName)
}

func testAccEventSubscriptionSourceIDsConfig(rName string) string {
	return acctest.ConfigCompose(testAccEventSubscriptionSourceIDsBaseConfig(rName), fmt.Sprintf(`
resource "aws_db_event_subscription" "test" {
  name        = %[1]q
  sns_topic   = aws_sns_topic.test.arn
  source_type = "db-parameter-group"

  event_categories = [
    "configuration change",
  ]

  source_ids = [
    aws_db_parameter_group.test1.id,
    aws_db_parameter_group.test2.id,
  ]
}
`, rName))
}

func testAccEventSubscriptionSourceIDsUpdatedConfig(rName string) string {
	return acctest.ConfigCompose(testAccEventSubscriptionSourceIDsBaseConfig(rName), fmt.Sprintf(`
resource "aws_db_event_subscription" "test" {
  name        = %[1]q
  sns_topic   = aws_sns_topic.test.arn
  source_type = "db-parameter-group"

  event_categories = [
    "configuration change",
  ]

  source_ids = [
    aws_db_parameter_group.test3.id,
  ]
}
`, rName))
}
