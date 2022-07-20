package redshift_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/redshift"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccRedshiftEventSubscription_basic(t *testing.T) {
	var v redshift.EventSubscription
	resourceName := "aws_redshift_event_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "redshift", regexp.MustCompile(`eventsubscription:.+`)),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "severity", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "status", "active"),
					acctest.CheckResourceAttrAccountID(resourceName, "customer_aws_id"),
					resource.TestCheckResourceAttr(resourceName, "event_categories.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_ids.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "sns_topic_arn", "aws_sns_topic.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEventSubscriptionConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "severity", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "source_type", "cluster-snapshot"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "monitoring"),
					resource.TestCheckResourceAttrPair(resourceName, "sns_topic_arn", "aws_sns_topic.test", "arn"),
					acctest.CheckResourceAttrAccountID(resourceName, "customer_aws_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccRedshiftEventSubscription_withSourceIDs(t *testing.T) {
	var v redshift.EventSubscription
	resourceName := "aws_redshift_event_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig_sourceIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "source_type", "cluster-parameter-group"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "source_ids.#", "1"),
				),
			},
			{
				Config: testAccEventSubscriptionConfig_updateSourceIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "source_type", "cluster-parameter-group"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccRedshiftEventSubscription_categoryUpdate(t *testing.T) {
	var v redshift.EventSubscription
	resourceName := "aws_redshift_event_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				Config: testAccEventSubscriptionConfig_updateCategories(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "source_type", "cluster"),
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

func TestAccRedshiftEventSubscription_tags(t *testing.T) {
	var v redshift.EventSubscription
	resourceName := "aws_redshift_event_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &v),
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
					testAccCheckEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccEventSubscriptionConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccRedshiftEventSubscription_disappears(t *testing.T) {
	var v redshift.EventSubscription
	resourceName := "aws_redshift_event_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfredshift.ResourceEventSubscription(), resourceName),
				),
				ExpectNonEmptyPlan: true,
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

		out, err := tfredshift.FindEventSubscriptionByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *out

		return nil
	}
}

func testAccCheckEventSubscriptionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_redshift_event_subscription" {
			continue
		}

		_, err := tfredshift.FindEventSubscriptionByName(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Redshift Event Subscription %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccEventSubscriptionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_redshift_event_subscription" "test" {
  name          = %[1]q
  sns_topic_arn = aws_sns_topic.test.arn
}
`, rName)
}

func testAccEventSubscriptionConfig_update(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_redshift_event_subscription" "test" {
  name          = %[1]q
  sns_topic_arn = aws_sns_topic.test.arn
  enabled       = false
  source_type   = "cluster-snapshot"
  severity      = "INFO"

  event_categories = [
    "monitoring",
  ]
}
`, rName)
}

func testAccEventSubscriptionConfig_sourceIDs(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_redshift_parameter_group" "test" {
  name        = %[1]q
  family      = "redshift-1.0"
  description = "Test parameter group for terraform"
}

resource "aws_redshift_event_subscription" "test" {
  name          = %[1]q
  sns_topic_arn = aws_sns_topic.test.arn
  source_type   = "cluster-parameter-group"
  severity      = "INFO"
  source_ids    = [aws_redshift_parameter_group.test.id]

  event_categories = [
    "configuration",
  ]
}
`, rName)
}

func testAccEventSubscriptionConfig_updateSourceIDs(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_redshift_parameter_group" "test" {
  name        = %[1]q
  family      = "redshift-1.0"
  description = "Test parameter group for terraform"
}

resource "aws_redshift_parameter_group" "foo" {
  name        = "%[1]s-2"
  family      = "redshift-1.0"
  description = "Test parameter group for terraform"
}

resource "aws_redshift_event_subscription" "test" {
  name          = %[1]q
  sns_topic_arn = aws_sns_topic.test.arn
  source_type   = "cluster-parameter-group"
  severity      = "INFO"
  source_ids    = [aws_redshift_parameter_group.test.id, aws_redshift_parameter_group.foo.id]

  event_categories = [
    "configuration",
  ]
}
`, rName)
}

func testAccEventSubscriptionConfig_updateCategories(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_redshift_event_subscription" "test" {
  name          = %[1]q
  sns_topic_arn = aws_sns_topic.test.arn
  source_type   = "cluster"
  severity      = "INFO"

  event_categories = [
    "monitoring",
  ]
}
`, rName)
}

func testAccEventSubscriptionConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_redshift_event_subscription" "test" {
  name          = %[1]q
  sns_topic_arn = aws_sns_topic.test.arn
  source_type   = "cluster"
  severity      = "INFO"

  event_categories = [
    "configuration",
    "management",
    "monitoring",
    "security",
  ]

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccEventSubscriptionConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_redshift_event_subscription" "test" {
  name          = %[1]q
  sns_topic_arn = aws_sns_topic.test.arn
  source_type   = "cluster"
  severity      = "INFO"

  event_categories = [
    "configuration",
    "management",
    "monitoring",
    "security",
  ]

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
