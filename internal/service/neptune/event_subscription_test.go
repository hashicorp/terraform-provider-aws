// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/neptune"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfneptune "github.com/hashicorp/terraform-provider-aws/internal/service/neptune"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNeptuneEventSubscription_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.EventSubscription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "rds", fmt.Sprintf("es:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, "true"),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-instance"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
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
					testAccCheckEventSubscriptionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, "false"),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-parameter-group"),
				),
			},
		},
	})
}

func TestAccNeptuneEventSubscription_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.EventSubscription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfneptune.ResourceEventSubscription(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccNeptuneEventSubscription_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.EventSubscription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig_nameGenerated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrNameGeneratedWithPrefix(resourceName, names.AttrName, "tf-"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "tf-"),
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

func TestAccNeptuneEventSubscription_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.EventSubscription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig_namePrefix(rName, "tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, "tf-acc-test-prefix-"),
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

func TestAccNeptuneEventSubscription_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.EventSubscription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(ctx, resourceName, &v),
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
					testAccCheckEventSubscriptionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccEventSubscriptionConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccNeptuneEventSubscription_withSourceIDs(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.EventSubscription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig_sourceIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-parameter-group"),
					resource.TestCheckResourceAttr(resourceName, "source_ids.#", "1"),
				),
			},
			{
				Config: testAccEventSubscriptionConfig_updateSourceIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-parameter-group"),
					resource.TestCheckResourceAttr(resourceName, "source_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccNeptuneEventSubscription_withCategories(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.EventSubscription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-instance"),
					resource.TestCheckResourceAttr(resourceName, "event_categories.#", "5"),
				),
			},
			{
				Config: testAccEventSubscriptionConfig_updateCategories(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSubscriptionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-instance"),
					resource.TestCheckResourceAttr(resourceName, "event_categories.#", "1"),
				),
			},
		},
	})
}

func testAccCheckEventSubscriptionExists(ctx context.Context, n string, v *neptune.EventSubscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NeptuneConn(ctx)

		output, err := tfneptune.FindEventSubscriptionByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckEventSubscriptionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NeptuneConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_neptune_event_subscription" {
				continue
			}

			_, err := tfneptune.FindEventSubscriptionByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Neptune Event Subscription %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccEventSubscriptionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_neptune_event_subscription" "test" {
  name          = %[1]q
  sns_topic_arn = aws_sns_topic.test.arn
  source_type   = "db-instance"

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

func testAccEventSubscriptionConfig_update(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_neptune_event_subscription" "test" {
  name          = %[1]q
  sns_topic_arn = aws_sns_topic.test.arn
  enabled       = false
  source_type   = "db-parameter-group"

  event_categories = [
    "configuration change",
  ]
}
`, rName)
}

func testAccEventSubscriptionConfig_nameGenerated(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_neptune_event_subscription" "test" {
  sns_topic_arn = aws_sns_topic.test.arn
  source_type   = "db-instance"

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

func testAccEventSubscriptionConfig_namePrefix(rName, namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_neptune_event_subscription" "test" {
  name_prefix   = %[2]q
  sns_topic_arn = aws_sns_topic.test.arn
  source_type   = "db-instance"

  event_categories = [
    "availability",
    "backup",
    "creation",
    "deletion",
    "maintenance",
  ]
}
`, rName, namePrefix)
}

func testAccEventSubscriptionConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_neptune_event_subscription" "test" {
  name          = %[1]q
  sns_topic_arn = aws_sns_topic.test.arn
  enabled       = false
  source_type   = "db-parameter-group"

  event_categories = [
    "configuration change",
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

resource "aws_neptune_event_subscription" "test" {
  name          = %[1]q
  sns_topic_arn = aws_sns_topic.test.arn
  enabled       = false
  source_type   = "db-parameter-group"

  event_categories = [
    "configuration change",
  ]

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccEventSubscriptionConfig_baseSourceIDs(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_neptune_parameter_group" "test" {
  count = 2

  name   = "%[1]s-${count.index}"
  family = "neptune1"
}
`, rName)
}

func testAccEventSubscriptionConfig_sourceIDs(rName string) string {
	return acctest.ConfigCompose(testAccEventSubscriptionConfig_baseSourceIDs(rName), fmt.Sprintf(`
resource "aws_neptune_event_subscription" "test" {
  name          = %[1]q
  sns_topic_arn = aws_sns_topic.test.arn
  source_type   = "db-parameter-group"
  source_ids    = [aws_neptune_parameter_group.test[0].id]

  event_categories = [
    "configuration change",
  ]
}
`, rName))
}

func testAccEventSubscriptionConfig_updateSourceIDs(rName string) string {
	return acctest.ConfigCompose(testAccEventSubscriptionConfig_baseSourceIDs(rName), fmt.Sprintf(`
resource "aws_neptune_event_subscription" "test" {
  name          = %[1]q
  sns_topic_arn = aws_sns_topic.test.arn
  source_type   = "db-parameter-group"
  source_ids    = aws_neptune_parameter_group.test[*].id

  event_categories = [
    "configuration change",
  ]
}
`, rName))
}

func testAccEventSubscriptionConfig_updateCategories(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_neptune_event_subscription" "test" {
  name          = %[1]q
  sns_topic_arn = aws_sns_topic.test.arn
  source_type   = "db-instance"
  enabled       = false

  event_categories = [
    "availability",
  ]
}
`, rName)
}
