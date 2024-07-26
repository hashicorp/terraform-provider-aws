// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/service/redshift"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftEventSubscription_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v redshift.EventSubscription
	resourceName := "aws_redshift_event_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "redshift", regexache.MustCompile(`eventsubscription:.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "severity", "INFO"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "active"),
					acctest.CheckResourceAttrAccountID(resourceName, "customer_aws_id"),
					resource.TestCheckResourceAttr(resourceName, "event_categories.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSNSTopicARN, "aws_sns_topic.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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
					testAccCheckEventSubscriptionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "severity", "INFO"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSourceType, "cluster-snapshot"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "monitoring"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSNSTopicARN, "aws_sns_topic.test", names.AttrARN),
					acctest.CheckResourceAttrAccountID(resourceName, "customer_aws_id"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccRedshiftEventSubscription_withSourceIDs(t *testing.T) {
	ctx := acctest.Context(t)
	var v redshift.EventSubscription
	resourceName := "aws_redshift_event_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig_sourceIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrSourceType, "cluster-parameter-group"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "source_ids.#", acctest.Ct1),
				),
			},
			{
				Config: testAccEventSubscriptionConfig_updateSourceIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrSourceType, "cluster-parameter-group"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "source_ids.#", acctest.Ct2),
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
	ctx := acctest.Context(t)
	var v redshift.EventSubscription
	resourceName := "aws_redshift_event_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				Config: testAccEventSubscriptionConfig_updateCategories(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrSourceType, "cluster"),
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
	ctx := acctest.Context(t)
	var v redshift.EventSubscription
	resourceName := "aws_redshift_event_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEventSubscriptionConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccEventSubscriptionConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccRedshiftEventSubscription_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v redshift.EventSubscription
	resourceName := "aws_redshift_event_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfredshift.ResourceEventSubscription(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEventSubscriptionExists(ctx context.Context, n string, v *redshift.EventSubscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Redshift Event Subscription is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn(ctx)

		out, err := tfredshift.FindEventSubscriptionByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *out

		return nil
	}
}

func testAccCheckEventSubscriptionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshift_event_subscription" {
				continue
			}

			_, err := tfredshift.FindEventSubscriptionByName(ctx, conn, rs.Primary.ID)

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
