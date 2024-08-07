// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSEventSubscription_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.EventSubscription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "rds", fmt.Sprintf("es:%s", rName)),
					acctest.CheckResourceAttrAccountID(resourceName, "customer_aws_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "event_categories.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
					resource.TestCheckResourceAttr(resourceName, "source_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrSourceType, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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
	ctx := acctest.Context(t)
	var eventSubscription types.EventSubscription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(ctx, resourceName, &eventSubscription),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfrds.ResourceEventSubscription(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRDSEventSubscription_Name_generated(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.EventSubscription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig_nameGenerated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "terraform-"),
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
	ctx := acctest.Context(t)
	var v types.EventSubscription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig_namePrefix(rName, "tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tf-acc-test-prefix-"),
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
	ctx := acctest.Context(t)
	var eventSubscription types.EventSubscription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(ctx, resourceName, &eventSubscription),
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
					testAccCheckEventSubscriptionExists(ctx, resourceName, &eventSubscription),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccEventSubscriptionConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(ctx, resourceName, &eventSubscription),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccRDSEventSubscription_categories(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.EventSubscription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_event_subscription.test"
	snsTopicResourceName := "aws_sns_topic.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig_categories(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "rds", fmt.Sprintf("es:%s", rName)),
					acctest.CheckResourceAttrAccountID(resourceName, "customer_aws_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "event_categories.#", "5"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "availability"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "backup"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "creation"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "deletion"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "maintenance"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
					resource.TestCheckResourceAttrPair(resourceName, "sns_topic", snsTopicResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "source_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrSourceType, "db-instance"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEventSubscriptionConfig_categoriesUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "rds", fmt.Sprintf("es:%s", rName)),
					acctest.CheckResourceAttrAccountID(resourceName, "customer_aws_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "event_categories.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "creation"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
					resource.TestCheckResourceAttrPair(resourceName, "sns_topic", snsTopicResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "source_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrSourceType, "db-instance"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccEventSubscriptionConfig_categoriesAndSourceTypeUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "rds", fmt.Sprintf("es:%s", rName)),
					acctest.CheckResourceAttrAccountID(resourceName, "customer_aws_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "event_categories.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "creation"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "deletion"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
					resource.TestCheckResourceAttrPair(resourceName, "sns_topic", snsTopicResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "source_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrSourceType, "db-cluster"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccRDSEventSubscription_sourceIDs(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.EventSubscription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_event_subscription.test"
	paramGroup1ResourceName := "aws_db_parameter_group.test1"
	paramGroup2ResourceName := "aws_db_parameter_group.test2"
	paramGroup3ResourceName := "aws_db_parameter_group.test3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig_sourceIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "source_ids.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "source_ids.*", paramGroup1ResourceName, names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "source_ids.*", paramGroup2ResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEventSubscriptionConfig_sourceIDsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "source_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "source_ids.*", paramGroup3ResourceName, names.AttrID),
				),
			},
		},
	})
}

func testAccCheckEventSubscriptionExists(ctx context.Context, n string, v *types.EventSubscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		output, err := tfrds.FindEventSubscriptionByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckEventSubscriptionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_db_event_subscription" {
				continue
			}

			_, err := tfrds.FindEventSubscriptionByID(ctx, conn, rs.Primary.ID)

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
}

func testAccEventSubscriptionConfig_basic(rName string) string {
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

func testAccEventSubscriptionConfig_nameGenerated(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_db_event_subscription" "test" {
  sns_topic = aws_sns_topic.test.arn
}
`, rName)
}

func testAccEventSubscriptionConfig_namePrefix(rName, namePrefix string) string {
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

func testAccEventSubscriptionConfig_tags1(rName, tagKey1, tagValue1 string) string {
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

func testAccEventSubscriptionConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

func testAccEventSubscriptionConfig_categories(rName string) string {
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

func testAccEventSubscriptionConfig_categoriesUpdated(rName string) string {
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

func testAccEventSubscriptionConfig_categoriesAndSourceTypeUpdated(rName string) string {
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

func testAccEventSubscriptionConfig_sourceIDs(rName string) string {
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

func testAccEventSubscriptionConfig_sourceIDsUpdated(rName string) string {
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
