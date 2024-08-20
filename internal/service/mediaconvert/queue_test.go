// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediaconvert_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/mediaconvert/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfmediaconvert "github.com/hashicorp/terraform-provider-aws/internal/service/mediaconvert"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccMediaConvertQueue_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var queue types.Queue
	resourceName := "aws_media_convert_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConvertServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queue),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "mediaconvert", regexache.MustCompile(`queues/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "pricing_plan", string(types.PricingPlanOnDemand)),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.QueueStatusActive)),
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

func TestAccMediaConvertQueue_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var queue types.Queue
	resourceName := "aws_media_convert_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConvertServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queue),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfmediaconvert.ResourceQueue(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccMediaConvertQueue_withTags(t *testing.T) {
	ctx := acctest.Context(t)
	var queue types.Queue
	resourceName := "aws_media_convert_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConvertServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queue),
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
				Config: testAccQueueConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccQueueConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccMediaConvertQueue_reservationPlanSettings(t *testing.T) {
	acctest.Skip(t, "MediaConvert Reserved Queues are $400/month and cannot be deleted for 1 year.")

	ctx := acctest.Context(t)
	var queue types.Queue
	resourceName := "aws_media_convert_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConvertServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_reserved(rName, string(types.CommitmentOneYear), string(types.RenewalTypeAutoRenew), 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "pricing_plan", string(types.PricingPlanReserved)),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.0.commitment", string(types.CommitmentOneYear)),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.0.renewal_type", string(types.RenewalTypeAutoRenew)),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.0.reserved_slots", acctest.Ct1),
				),
			},
			{
				Config: testAccQueueConfig_reserved(rName, string(types.CommitmentOneYear), string(types.RenewalTypeExpire), 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "pricing_plan", string(types.PricingPlanReserved)),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.0.commitment", string(types.CommitmentOneYear)),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.0.renewal_type", string(types.RenewalTypeExpire)),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.0.reserved_slots", acctest.Ct1),
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

func TestAccMediaConvertQueue_withStatus(t *testing.T) {
	ctx := acctest.Context(t)
	var queue types.Queue
	resourceName := "aws_media_convert_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConvertServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_status(rName, string(types.QueueStatusPaused)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.QueueStatusPaused)),
				),
			},
			{
				Config: testAccQueueConfig_status(rName, string(types.QueueStatusActive)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.QueueStatusActive)),
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

func TestAccMediaConvertQueue_withDescription(t *testing.T) {
	ctx := acctest.Context(t)
	var queue types.Queue
	resourceName := "aws_media_convert_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description1 := sdkacctest.RandomWithPrefix("Description: ")
	description2 := sdkacctest.RandomWithPrefix("Description: ")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConvertServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_description(rName, description1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description1),
				),
			},
			{
				Config: testAccQueueConfig_description(rName, description2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description2),
				),
			},
		},
	})
}

func testAccCheckQueueDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_media_convert_queue" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).MediaConvertClient(ctx)

			_, err := tfmediaconvert.FindQueueByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Media Convert Queue %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckQueueExists(ctx context.Context, n string, v *types.Queue) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaConvertClient(ctx)

		output, err := tfmediaconvert.FindQueueByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccQueueConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_media_convert_queue" "test" {
  name = %[1]q
}
`, rName)
}

func testAccQueueConfig_status(rName, status string) string {
	return fmt.Sprintf(`
resource "aws_media_convert_queue" "test" {
  name   = %[1]q
  status = %[2]q
}
`, rName, status)
}

func testAccQueueConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_media_convert_queue" "test" {
  name        = %[1]q
  description = %[2]q
}
`, rName, description)
}

func testAccQueueConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_media_convert_queue" "test" {
  name = %[1]q

  tags = {
    %[2]s = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccQueueConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_media_convert_queue" "test" {
  name = %[1]q

  tags = {
    %[2]s = %[3]q
    %[4]s = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccQueueConfig_reserved(rName, commitment, renewalType string, reservedSlots int) string {
	return fmt.Sprintf(`
resource "aws_media_convert_queue" "test" {
  name         = %[1]q
  pricing_plan = %[2]q

  reservation_plan_settings {
    commitment     = %[3]q
    renewal_type   = %[4]q
    reserved_slots = %[5]d
  }
}
`, rName, string(types.PricingPlanReserved), commitment, renewalType, reservedSlots)
}
